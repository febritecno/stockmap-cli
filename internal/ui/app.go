package ui

import (
	"strconv"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"stockmap/internal/alerts"
	"stockmap/internal/fetcher"
	"stockmap/internal/history"
	"stockmap/internal/screener"
	"stockmap/internal/ui/views"
)

// View represents the current view state
type View int

const (
	ViewSplash View = iota
	ViewDashboard
	ViewScanner
	ViewDetails
	ViewWatchlist
	ViewHistory
	ViewConnection
	ViewScanMode
	ViewAlerts
)

// Model is the main bubbletea model
type Model struct {
	width             int
	height            int
	currentView       View
	splash            *views.Splash
	dashboard         *views.Dashboard
	scanner           *views.Scanner
	details           *views.Details
	watchlist         *views.WatchlistView
	historyView       *views.HistoryView
	connectionView    *views.ConnectionView
	scanModeView      *views.ScanModeView
	alertsView        *views.AlertsView
	engine            *screener.Engine
	historyMgr        *history.Manager
	alertsMgr         *alerts.Manager
	scanning          bool
	results           []*screener.ScreenResult
	totalScanned      int
	err               error
	loadedFromHistory bool
	loadedHistoryID   string
	scanSymbols       []string // Custom symbols to scan
	isReload          bool     // Track if this is a reload (to update history instead of create new)
	lastHistoryID     string   // Last saved history ID for reload
	// Auto-reload
	autoReload        bool // Auto-reload enabled
	autoReloadSeconds int  // Seconds between reloads (default 60)
	autoReloadCounter int  // Current countdown
	// Triggered alerts
	triggeredAlerts []alerts.TriggeredAlert
}

// ScanProgressMsg is sent during scanning
type ScanProgressMsg struct {
	Completed    int
	Total        int
	Current      string
	SuccessCount int
	ErrorCount   int
	LastError    string // Last error message for verbose display
	ErrorSymbol  string // Symbol that caused the last error
}

// ScanCompleteMsg is sent when scanning is complete
type ScanCompleteMsg struct {
	Results []*screener.ScreenResult
}

// MarketStatusMsg contains market status
type MarketStatusMsg struct {
	Status string
}

// ErrorMsg represents an error
type ErrorMsg struct {
	Err error
}

// HistorySavedMsg is sent when history is saved
type HistorySavedMsg struct {
	ID string
}

// ConnectionTestMsg triggers connection testing
type ConnectionTestMsg struct{}

// ConnectionResultMsg contains connection test result
type ConnectionResultMsg struct {
	Result *fetcher.ConnectionResult
}

// AnimationTickMsg triggers animation updates
type AnimationTickMsg struct{}

// ReloadTickMsg triggers reload animation updates
type ReloadTickMsg struct{}

// AutoReloadTickMsg triggers auto-reload countdown
type AutoReloadTickMsg struct{}

// AlertTriggeredMsg is sent when an alert is triggered
type AlertTriggeredMsg struct {
	Alert alerts.TriggeredAlert
}

// NewModel creates a new app model
func NewModel() *Model {
	alertsMgr := alerts.NewManager("")
	return &Model{
		currentView:       ViewSplash,
		splash:            views.NewSplash(),
		dashboard:         views.NewDashboard(),
		scanner:           views.NewScanner(),
		details:           views.NewDetails(),
		watchlist:         views.NewWatchlistView(),
		historyView:       views.NewHistoryView(),
		connectionView:    views.NewConnectionView(),
		scanModeView:      views.NewScanModeView(),
		alertsView:        views.NewAlertsView(alertsMgr),
		engine:            screener.NewEngine(10), // 10 workers with rate limiting
		historyMgr:        history.NewManager(),
		alertsMgr:         alertsMgr,
		autoReloadSeconds: 60, // Default 60 seconds
	}
}

// Init initializes the model
func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		tea.EnterAltScreen,
		m.splashTick(),
	)
}

// splashTick returns a command for splash animation
func (m *Model) splashTick() tea.Cmd {
	return tea.Tick(150*time.Millisecond, func(t time.Time) tea.Msg {
		return AnimationTickMsg{}
	})
}

// animationTick returns a command for general animation
func (m *Model) animationTick() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return AnimationTickMsg{}
	})
}

// reloadTick returns a command for reload animation (faster tick for spinner)
func (m *Model) reloadTick() tea.Cmd {
	return tea.Tick(200*time.Millisecond, func(t time.Time) tea.Msg {
		return ReloadTickMsg{}
	})
}

// autoReloadTick returns a command for auto-reload countdown (1 second)
func (m *Model) autoReloadTick() tea.Cmd {
	return tea.Tick(1*time.Second, func(t time.Time) tea.Msg {
		return AutoReloadTickMsg{}
	})
}

// checkMarketStatus fetches the market status
func (m *Model) checkMarketStatus() tea.Msg {
	client := fetcher.NewDirectYahooClient()
	defer client.Close()
	status := client.GetMarketStatus()
	return MarketStatusMsg{Status: status}
}

// runConnectionTest runs the connection test
func (m *Model) runConnectionTest() tea.Cmd {
	return func() tea.Msg {
		client := fetcher.NewDirectYahooClient()
		defer client.Close()
		result := client.CheckConnection()
		return ConnectionResultMsg{Result: &result}
	}
}

// startScan starts the stock scanning process
func (m *Model) startScan() tea.Cmd {
	return tea.Batch(
		// Start animation tick
		m.animationTick(),
		// Start actual scan
		func() tea.Msg {
			// Use custom symbols if set, otherwise default
			symbols := m.scanSymbols
			if len(symbols) == 0 {
				symbols = fetcher.DefaultSymbols()
			}
			// Only take first 50 symbols for "Scan All" to prevent rate limiting issues for now
			// until we implement better batching/backoff
			if m.scanSymbols == nil && len(symbols) > 50 {
				symbols = symbols[:50]
			}
			total := len(symbols)

			// Set verbose progress callback
			m.engine.SetVerboseProgressCallback(func(progress screener.ScanProgress) {
				// Progress will be checked via tickScan
			})

			// Start scanning in goroutine
			go func() {
				m.engine.Scan(symbols)
			}()

			// Return initial progress
			return ScanProgressMsg{Completed: 0, Total: total, Current: "Starting..."}
		},
	)
}

// saveHistory saves current results to history
func (m *Model) saveHistory() tea.Cmd {
	return func() tea.Msg {
		if len(m.results) == 0 {
			return nil
		}

		var record *history.ScanRecord
		var err error

		// If this is a reload and we have a previous history ID, update instead of create new
		if m.isReload && m.lastHistoryID != "" {
			record, err = m.historyMgr.Update(m.lastHistoryID, m.results, m.totalScanned)
		} else {
			record, err = m.historyMgr.Save(m.results, m.totalScanned)
		}

		if err != nil {
			return ErrorMsg{Err: err}
		}

		return HistorySavedMsg{ID: record.ID}
	}
}

// Update handles messages
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.splash.SetSize(msg.Width, msg.Height)
		m.dashboard.SetSize(msg.Width, msg.Height)
		m.scanner.SetSize(msg.Width, msg.Height)
		m.details.SetSize(msg.Width, msg.Height)
		m.watchlist.SetSize(msg.Width, msg.Height)
		m.historyView.SetSize(msg.Width, msg.Height)
		m.connectionView.SetSize(msg.Width, msg.Height)
		m.scanModeView.SetSize(msg.Width, msg.Height)
		m.alertsView.SetSize(msg.Width, msg.Height)
		return m, nil

	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case AnimationTickMsg:
		if m.currentView == ViewSplash {
			m.splash.NextFrame()
			if m.splash.IsDone() {
				m.currentView = ViewDashboard
				return m, m.checkMarketStatus
			}
			return m, m.splashTick()
		}
		// Animate scanner view during scanning (only for non-reload scans)
		if m.currentView == ViewScanner && m.scanning {
			m.scanner.NextFrame()
			return m, m.animationTick()
		}
		return m, nil

	case ReloadTickMsg:
		// Update reload spinner animation
		if m.scanning && m.isReload && m.currentView == ViewDashboard {
			m.dashboard.NextReloadFrame()
			return m, m.reloadTick()
		}
		return m, nil

	case AutoReloadTickMsg:
		// Auto-reload countdown
		if m.autoReload && !m.scanning {
			m.autoReloadCounter--
			m.dashboard.SetAutoReload(true, m.autoReloadCounter)

			if m.autoReloadCounter <= 0 {
				// Trigger reload - reload ALL symbols from current table results
				if len(m.results) > 0 {
					symbols := make([]string, 0, len(m.results))
					for _, r := range m.results {
						symbols = append(symbols, r.Symbol)
					}
					m.scanSymbols = symbols
					m.autoReloadCounter = m.autoReloadSeconds
					m.isReload = true
					m.scanning = true
					m.dashboard.SetReloading(true)
					return m, tea.Batch(m.startScan(), m.reloadTick(), m.autoReloadTick())
				}
				// No results, disable auto-reload
				m.autoReload = false
				m.dashboard.SetAutoReload(false, 0)
				m.dashboard.SetMessage("Auto-reload stopped: no data")
				return m, nil
			}
			return m, m.autoReloadTick()
		}
		return m, nil

	case MarketStatusMsg:
		m.dashboard.SetMarketState(msg.Status)
		return m, nil

	case ScanProgressMsg:
		m.scanner.SetVerboseProgress(msg.Completed, msg.Total, msg.Current, msg.SuccessCount, msg.ErrorCount, msg.LastError, msg.ErrorSymbol)

		// Update dashboard based on reload or regular scan
		if m.isReload && m.currentView == ViewDashboard {
			// Reload mode - just update stats, spinner is already showing
			m.dashboard.SetScanning(false, "", msg.Completed) // Don't show scanning text, just stats
		} else {
			m.dashboard.SetScanning(true, msg.Current, msg.Completed)
		}

		// Set totalScanned if not set yet
		if msg.Total > 0 {
			m.totalScanned = msg.Total
		}

		// Check completion - ensure Total > 0 to avoid false completion
		if msg.Total > 0 && msg.Completed >= msg.Total {
			// Scanning complete
			m.scanning = false
			m.results = m.engine.GetResults()
			m.scanner.SetFoundCount(len(m.results))
			m.dashboard.SetResults(m.results)
			m.dashboard.SetScanning(false, "", msg.Completed)
			m.dashboard.SetReloading(false) // Stop reload spinner
			m.watchlist.SetResults(m.results)
			m.loadedFromHistory = false
			m.loadedHistoryID = ""

			// Check alerts for each result
			m.checkAlerts()

			// Continue auto-reload timer if enabled
			var cmds []tea.Cmd

			// Only save history for new scans, NOT for reload
			if !m.isReload {
				cmds = append(cmds, m.saveHistory())
			} else {
				m.dashboard.SetMessage("Reloaded")
				m.isReload = false
			}

			if m.autoReload {
				m.autoReloadCounter = m.autoReloadSeconds
				m.dashboard.SetAutoReload(true, m.autoReloadCounter)
				cmds = append(cmds, m.autoReloadTick())
			}

			if len(cmds) > 0 {
				return m, tea.Batch(cmds...)
			}
			return m, nil
		}

		// Continue checking for progress
		return m, m.tickScan()

	case ScanCompleteMsg:
		m.scanning = false
		m.results = msg.Results
		m.scanner.SetFoundCount(len(m.results))
		m.dashboard.SetResults(m.results)
		m.dashboard.SetScanning(false, "", len(m.results))
		m.dashboard.SetReloading(false)
		m.watchlist.SetResults(m.results)
		return m, m.saveHistory()

	case HistorySavedMsg:
		// Save the history ID for future reloads
		if m.isReload {
			m.dashboard.SetMessage("Reloaded")
		} else {
			m.lastHistoryID = msg.ID
			m.dashboard.SetMessage("Saved")
		}
		m.isReload = false // Reset reload flag
		return m, nil

	case ErrorMsg:
		m.err = msg.Err
		m.dashboard.SetMessage("Error: " + msg.Err.Error())
		return m, nil

	case ConnectionTestMsg:
		m.connectionView.SetTesting(true)
		return m, m.runConnectionTest()

	case ConnectionResultMsg:
		m.connectionView.SetResult(msg.Result)
		return m, nil
	}

	return m, nil
}

// tickScan returns a command to check scan progress and animate
func (m *Model) tickScan() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		progress := m.engine.GetProgress()

		return ScanProgressMsg{
			Completed:    progress.Completed,
			Total:        progress.Total,
			Current:      progress.Current,
			SuccessCount: progress.SuccessCount,
			ErrorCount:   progress.ErrorCount,
			LastError:    progress.LastError,
			ErrorSymbol:  progress.ErrorSymbol,
		}
	})
}

// handleKeyPress handles keyboard input
func (m *Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle splash screen
	if m.currentView == ViewSplash {
		// Any key skips splash
		m.splash.Skip()
		m.currentView = ViewDashboard
		return m, m.checkMarketStatus
	}

	// Global keys
	switch msg.String() {
	case "ctrl+c", "q":
		if m.currentView == ViewDashboard {
			return m, tea.Quit
		}
		m.currentView = ViewDashboard
		return m, nil

	case "esc":
		if m.currentView != ViewDashboard {
			m.currentView = ViewDashboard
			return m, nil
		}
		return m, tea.Quit
	}

	// View-specific keys
	switch m.currentView {
	case ViewDashboard:
		return m.handleDashboardKeys(msg)
	case ViewScanner:
		return m.handleScannerKeys(msg)
	case ViewDetails:
		return m.handleDetailsKeys(msg)
	case ViewWatchlist:
		return m.handleWatchlistKeys(msg)
	case ViewHistory:
		return m.handleHistoryKeys(msg)
	case ViewConnection:
		return m.handleConnectionKeys(msg)
	case ViewScanMode:
		return m.handleScanModeKeys(msg)
	case ViewAlerts:
		return m.handleAlertsKeys(msg)
	}

	return m, nil
}

// handleDashboardKeys handles dashboard-specific keys
func (m *Model) handleDashboardKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "s", "S":
		// Show scan mode selection
		m.scanModeView.Reset()
		m.scanModeView.SetWatchlistCount(m.engine.GetWatchlistManager().Count())
		m.currentView = ViewScanMode
		return m, nil

	case "c", "C":
		// Check connection
		m.currentView = ViewConnection
		m.connectionView.SetTesting(true)
		return m, m.runConnectionTest()

	case "w", "W":
		// Switch to watchlist view
		m.currentView = ViewWatchlist
		m.watchlist.Refresh()
		return m, nil

	case "h", "H":
		// Switch to history view
		m.currentView = ViewHistory
		m.historyView.Refresh()
		return m, nil

	case "d", "D", "enter":
		// Show details
		if selected := m.dashboard.SelectedResult(); selected != nil {
			m.details.SetStock(selected)
			m.currentView = ViewDetails
		}
		return m, nil

	case "a", "A":
		// Add to watchlist
		if selected := m.dashboard.SelectedResult(); selected != nil {
			m.engine.AddToWatchlist(selected.Symbol)
			m.results = m.engine.GetResults()
			m.dashboard.SetResults(m.results)
			m.dashboard.SetMessage("Added " + selected.Symbol + " to watchlist")
		}
		return m, nil

	case "r", "R":
		// Reload/Refresh - toggle: if scanning, cancel; if not, start reload
		if m.scanning {
			// Cancel current reload/scan
			m.engine.Stop()
			m.scanning = false
			m.isReload = false
			m.dashboard.SetReloading(false)
			m.dashboard.SetScanning(false, "", 0)
			m.dashboard.SetMessage("Reload cancelled")
			return m, nil
		}

		// Determine if we can reload
		canReload := false

		// 1. If we have results, refresh ONLY them
		if len(m.results) > 0 {
			// Extract symbols from current results in table
			symbols := make([]string, 0, len(m.results))
			for _, r := range m.results {
				symbols = append(symbols, r.Symbol)
			}
			m.scanSymbols = symbols
			canReload = true
		} else if m.totalScanned > 0 || len(m.scanSymbols) > 0 {
			// 2. If no results but we scanned before (Retry last scan)
			// Keep existing m.scanSymbols
			canReload = true
		}

		if canReload {
			m.isReload = true
			m.scanning = true
			m.dashboard.SetReloading(true)
			m.dashboard.SetMessage("Reloading...")
			return m, tea.Batch(m.startScan(), m.reloadTick())
		}

		// No results to reload, show message
		m.dashboard.SetMessage("No data to reload. Press [S] to scan first.")
		return m, nil

	case "t", "T":
		// Toggle auto-reload
		if len(m.results) > 0 || len(m.scanSymbols) > 0 {
			m.autoReload = !m.autoReload
			if m.autoReload {
				m.autoReloadCounter = m.autoReloadSeconds
				m.dashboard.SetAutoReload(true, m.autoReloadCounter)
				m.dashboard.SetMessage("Auto-reload ON (60s)")
				return m, m.autoReloadTick()
			} else {
				m.dashboard.SetAutoReload(false, 0)
				m.dashboard.SetMessage("Auto-reload OFF")
			}
		} else {
			m.dashboard.SetMessage("Run a scan first before enabling auto-reload")
		}
		return m, nil

	case "p", "P":
		// Switch to alerts view
		m.alertsView.Refresh()
		m.currentView = ViewAlerts
		return m, nil

	case "up", "k":
		m.dashboard.MoveUp()
		return m, nil

	case "down", "j":
		m.dashboard.MoveDown()
		return m, nil

	case "x", "X", "backspace":
		// Remove selected stock from results
		if selected := m.dashboard.SelectedResult(); selected != nil {
			// Filter out the selected symbol
			newResults := make([]*screener.ScreenResult, 0, len(m.results)-1)
			for _, r := range m.results {
				if r.Symbol != selected.Symbol {
					newResults = append(newResults, r)
				}
			}
			m.results = newResults
			m.dashboard.SetResults(m.results)
			m.dashboard.SetMessage("Removed " + selected.Symbol)
		}
		return m, nil
	}

	return m, nil
}

// handleScannerKeys handles scanner-specific keys
func (m *Model) handleScannerKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.scanner.IsComplete() {
		// Any key returns to dashboard
		m.currentView = ViewDashboard
		return m, nil
	}

	switch msg.String() {
	case "esc":
		m.engine.Stop()
		m.scanning = false
		// Save whatever results we have so far
		m.results = m.engine.GetResults()
		m.dashboard.SetResults(m.results)
		m.currentView = ViewDashboard
		return m, nil
	}

	return m, nil
}

// handleDetailsKeys handles details-specific keys
func (m *Model) handleDetailsKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "g", "G":
		// Toggle chart view
		m.details.ToggleChart()
		return m, nil

	case "a", "A":
		if m.details != nil {
			stock := m.dashboard.SelectedResult()
			if stock != nil {
				m.engine.AddToWatchlist(stock.Symbol)
				m.results = m.engine.GetResults()
				m.dashboard.SetResults(m.results)
				m.details.SetStock(m.findStock(stock.Symbol))
			}
		}
		return m, nil

	case "r", "R":
		if m.details != nil {
			stock := m.dashboard.SelectedResult()
			if stock != nil {
				m.engine.RemoveFromWatchlist(stock.Symbol)
				m.results = m.engine.GetResults()
				m.dashboard.SetResults(m.results)
				m.details.SetStock(m.findStock(stock.Symbol))
			}
		}
		return m, nil
	}

	return m, nil
}

// handleWatchlistKeys handles watchlist-specific keys
func (m *Model) handleWatchlistKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle input mode
	if m.watchlist.IsInputActive() {
		switch msg.String() {
		case "esc":
			m.watchlist.ClearInput()
			return m, nil
		case "enter":
			symbol := m.watchlist.GetInputSymbol()
			if symbol != "" {
				m.engine.AddToWatchlist(symbol)
				m.results = m.engine.GetResults()
				m.dashboard.SetResults(m.results)
				m.watchlist.SetResults(m.results)
				m.dashboard.SetMessage("Added " + symbol + " to watchlist")
			}
			m.watchlist.ClearInput()
			return m, nil
		case "backspace":
			m.watchlist.Backspace()
			return m, nil
		default:
			if len(msg.String()) == 1 {
				m.watchlist.AddChar(rune(msg.String()[0]))
			}
			return m, nil
		}
	}

	// Handle Category Mode
	if m.watchlist.IsCategoryMode() {
		switch msg.String() {
		case "esc":
			m.watchlist.ToggleCategoryMode()
			return m, nil
		case "enter":
			// Add all symbols from selected category
			symbols := m.watchlist.GetSelectedCategorySymbols()
			catName := m.watchlist.GetSelectedCategoryName()
			if len(symbols) > 0 {
				count := 0
				for _, s := range symbols {
					if err := m.engine.AddToWatchlist(s); err == nil {
						count++
					}
				}
				m.results = m.engine.GetResults()
				m.dashboard.SetResults(m.results)
				m.watchlist.SetResults(m.results)
				m.dashboard.SetMessage("Added " + strconv.Itoa(count) + " stocks from " + catName)
			}
			m.watchlist.ToggleCategoryMode()
			return m, nil
		case "up", "k":
			m.watchlist.MoveUp()
			return m, nil
		case "down", "j":
			m.watchlist.MoveDown()
			return m, nil
		}
		return m, nil
	}

	switch msg.String() {
	case "h", "H":
		m.watchlist.ToggleCategoryMode()
		return m, nil

	case "a", "A":
		// Add new symbol to watchlist
		m.watchlist.ToggleInput()
		return m, nil

	case "up", "k":
		m.watchlist.MoveUp()
		return m, nil

	case "down", "j":
		m.watchlist.MoveDown()
		return m, nil

	case "d", "D", "enter":
		if selected := m.watchlist.SelectedResult(); selected != nil {
			m.details.SetStock(selected)
			m.currentView = ViewDetails
		}
		return m, nil

	case "r", "R":
		if selected := m.watchlist.SelectedResult(); selected != nil {
			m.engine.RemoveFromWatchlist(selected.Symbol)
			m.results = m.engine.GetResults()
			m.dashboard.SetResults(m.results)
			m.watchlist.SetResults(m.results)
		}
		return m, nil
	}

	return m, nil
}

// handleHistoryKeys handles history-specific keys
func (m *Model) handleHistoryKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		m.historyView.MoveUp()
		return m, nil

	case "down", "j":
		m.historyView.MoveDown()
		return m, nil

	case "enter":
		// Load selected history
		record, err := m.historyView.LoadSelected()
		if err != nil {
			m.dashboard.SetMessage("Error loading history: " + err.Error())
			return m, nil
		}

		// Set results from history
		m.results = record.Results
		m.totalScanned = record.TotalScanned
		m.dashboard.SetResults(m.results)
		m.watchlist.SetResults(m.results)
		m.loadedFromHistory = true
		m.loadedHistoryID = record.ID
		m.dashboard.SetMessage("Loaded scan from " + history.FormatTimestamp(record.Timestamp))
		m.currentView = ViewDashboard
		return m, nil

	case "x", "X":
		// Delete selected history
		if err := m.historyView.DeleteSelected(); err != nil {
			m.dashboard.SetMessage("Error deleting: " + err.Error())
		}
		return m, nil

	case "h", "H":
		// Back to dashboard
		m.currentView = ViewDashboard
		return m, nil
	}

	return m, nil
}

// handleConnectionKeys handles connection view keys
func (m *Model) handleConnectionKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "c", "C":
		// Test again
		m.connectionView.SetTesting(true)
		return m, m.runConnectionTest()

	case "esc":
		m.currentView = ViewDashboard
		return m, nil
	}

	return m, nil
}

// handleScanModeKeys handles scan mode selection keys
func (m *Model) handleScanModeKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// If input is active, handle typing
	if m.scanModeView.IsInputActive() {
		switch msg.String() {
		case "esc":
			m.scanModeView.ToggleInput()
			return m, nil
		case "enter":
			m.scanModeView.ToggleInput()
			return m, nil
		case "backspace":
			m.scanModeView.Backspace()
			return m, nil
		default:
			// Add character to input
			if len(msg.String()) == 1 {
				m.scanModeView.AddChar(rune(msg.String()[0]))
			}
			return m, nil
		}
	}

	switch msg.String() {
	case "1":
		m.scanModeView.Reset()
		// Scan all - start immediately
		m.scanSymbols = nil
		m.isReload = false // New scan, not reload
		m.currentView = ViewScanner
		m.scanning = true
		return m, m.startScan()

	case "2":
		// Scan watchlist only
		watchlistSymbols := m.engine.GetWatchlistManager().GetAll()
		if len(watchlistSymbols) == 0 {
			m.dashboard.SetMessage("Watchlist is empty! Add stocks first.")
			m.currentView = ViewDashboard
			return m, nil
		}
		m.scanSymbols = watchlistSymbols
		m.isReload = false // New scan, not reload
		m.currentView = ViewScanner
		m.scanning = true
		return m, m.startScan()

	case "3":
		// Select custom mode
		m.scanModeView.ToggleInput()
		return m, nil

	case "up", "k":
		m.scanModeView.MoveUp()
		return m, nil

	case "down", "j":
		m.scanModeView.MoveDown()
		return m, nil

	case "tab":
		// Toggle input for custom mode
		if m.scanModeView.GetSelectedMode() == views.ScanModeCustom {
			m.scanModeView.ToggleInput()
		}
		return m, nil

	case "enter":
		// Start scan with selected mode
		mode := m.scanModeView.GetSelectedMode()
		switch mode {
		case views.ScanModeAll:
			m.scanSymbols = nil
		case views.ScanModeWatchlist:
			watchlistSymbols := m.engine.GetWatchlistManager().GetAll()
			if len(watchlistSymbols) == 0 {
				m.dashboard.SetMessage("Watchlist is empty!")
				m.currentView = ViewDashboard
				return m, nil
			}
			m.scanSymbols = watchlistSymbols
		case views.ScanModeCustom:
			customSymbols := m.scanModeView.GetCustomSymbols()
			if len(customSymbols) == 0 {
				m.dashboard.SetMessage("No symbols entered!")
				m.currentView = ViewDashboard
				return m, nil
			}
			m.scanSymbols = customSymbols
		}
		m.isReload = false // New scan, not reload
		m.currentView = ViewScanner
		m.scanning = true
		return m, m.startScan()

	case "esc":
		m.currentView = ViewDashboard
		return m, nil
	}

	return m, nil
}

// handleAlertsKeys handles alerts-specific keys
func (m *Model) handleAlertsKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle input mode
	if m.alertsView.IsInputActive() {
		switch msg.String() {
		case "esc":
			m.alertsView.ClearInput()
			return m, nil
		case "enter":
			// Get current price for the symbol
			lastPrice := 0.0
			if stock := m.findStock(m.alertsView.SelectedAlert().Symbol); stock != nil {
				lastPrice = stock.Price
			}
			m.alertsView.SubmitAlert(lastPrice)
			return m, nil
		case "tab":
			m.alertsView.NextInputField()
			return m, nil
		case "shift+tab":
			m.alertsView.PrevInputField()
			return m, nil
		case " ":
			m.alertsView.CycleAlertType()
			return m, nil
		case "backspace":
			m.alertsView.Backspace()
			return m, nil
		default:
			if len(msg.String()) == 1 {
				m.alertsView.AddChar(rune(msg.String()[0]))
			}
			return m, nil
		}
	}

	switch msg.String() {
	case "n", "N":
		// New alert
		if selected := m.dashboard.SelectedResult(); selected != nil {
			m.alertsView.SetCurrentStock(selected)
		}
		m.alertsView.ToggleInput()
		return m, nil

	case "d", "D":
		// Delete selected alert
		m.alertsView.DeleteSelected()
		return m, nil

	case "t", "T":
		// Toggle active state
		m.alertsView.ToggleSelected()
		return m, nil

	case "r", "R":
		// Reset triggered alert
		m.alertsView.ResetSelected()
		return m, nil

	case "c", "C":
		// Clear triggered alerts
		m.alertsView.ClearTriggered()
		return m, nil

	case "up", "k":
		m.alertsView.MoveUp()
		return m, nil

	case "down", "j":
		m.alertsView.MoveDown()
		return m, nil

	case "esc":
		m.currentView = ViewDashboard
		return m, nil
	}

	return m, nil
}

// findStock finds a stock by symbol in results
func (m *Model) findStock(symbol string) *screener.ScreenResult {
	for _, r := range m.results {
		if r.Symbol == symbol {
			return r
		}
	}
	return nil
}

// View renders the current view
func (m *Model) View() string {
	switch m.currentView {
	case ViewSplash:
		return m.splash.View()
	case ViewScanner:
		return m.scanner.View()
	case ViewDetails:
		return m.details.View()
	case ViewWatchlist:
		return m.watchlist.View()
	case ViewHistory:
		return m.historyView.View()
	case ViewConnection:
		return m.connectionView.View()
	case ViewScanMode:
		return m.scanModeView.View()
	case ViewAlerts:
		return m.alertsView.View()
	default:
		return m.dashboard.View()
	}
}

// checkAlerts checks all results against active alerts
func (m *Model) checkAlerts() {
	for _, result := range m.results {
		triggered := m.alertsMgr.CheckPrice(result.Symbol, result.Price, result.RSI)
		m.triggeredAlerts = append(m.triggeredAlerts, triggered...)
	}

	// Show notification if alerts were triggered
	if len(m.triggeredAlerts) > 0 {
		count := len(m.triggeredAlerts)
		msg := "ALERT: 1 price alert triggered! Press [P] to view."
		if count > 1 {
			msg = "ALERT: Multiple price alerts triggered! Press [P] to view."
		}
		m.dashboard.SetMessage(msg)
	}
}

// Run starts the application
func Run() error {
	p := tea.NewProgram(NewModel(), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
