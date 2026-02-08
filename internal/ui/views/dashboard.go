package views

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/febritecno/stockmap/internal/screener"
	"github.com/febritecno/stockmap/internal/styles"
	"github.com/febritecno/stockmap/internal/ui/components"
)

// Dashboard is the main view showing the stock table
type Dashboard struct {
	header    *components.Header
	table     *components.Table
	statusBar *components.StatusBar
	width     int
	height    int
}

// NewDashboard creates a new dashboard view
func NewDashboard() *Dashboard {
	return &Dashboard{
		header:    components.NewHeader(),
		table:     components.NewTable(),
		statusBar: components.NewStatusBar(),
	}
}

// SetSize updates the dashboard dimensions
func (d *Dashboard) SetSize(width, height int) {
	d.width = width
	d.height = height

	d.header.SetWidth(width)
	d.statusBar.SetWidth(width)

	// Calculate table height (total - header - status bar - padding)
	tableHeight := height - 6
	if tableHeight < 5 {
		tableHeight = 5
	}
	d.table.SetSize(width, tableHeight)
}

// SetResults updates the table with screen results
func (d *Dashboard) SetResults(results []*screener.ScreenResult) {
	d.table.SetRows(results)
	d.statusBar.SetStats(len(results), len(results))
}

// SetMarketState updates the market status
func (d *Dashboard) SetMarketState(state string) {
	d.header.SetMarketState(state)
}

// SetScanning updates scanning state
func (d *Dashboard) SetScanning(scanning bool, symbol string, scanned int) {
	d.statusBar.SetScanning(scanning, symbol)
	d.statusBar.SetStats(scanned, len(d.table.GetRows()))
}

// SetReloading updates reloading state (simple spinner)
func (d *Dashboard) SetReloading(reloading bool) {
	d.statusBar.SetReloading(reloading)
}

// NextReloadFrame advances the reload spinner
func (d *Dashboard) NextReloadFrame() {
	d.statusBar.NextReloadFrame()
}

// SetAutoReload sets auto-reload state
func (d *Dashboard) SetAutoReload(enabled bool, secondsLeft int) {
	d.statusBar.SetAutoReload(enabled, secondsLeft)
}

// SetMessage sets a status message
func (d *Dashboard) SetMessage(msg string) {
	d.statusBar.SetMessage(msg)
}

// MoveUp moves selection up
func (d *Dashboard) MoveUp() {
	d.table.MoveUp()
}

// MoveDown moves selection down
func (d *Dashboard) MoveDown() {
	d.table.MoveDown()
}

// SelectedResult returns the selected stock
func (d *Dashboard) SelectedResult() *screener.ScreenResult {
	return d.table.SelectedRow()
}

// View renders the dashboard
func (d *Dashboard) View() string {
	headerView := d.header.View()
	divider := components.RenderDivider(d.width)
	tableView := d.table.View()
	statusView := d.statusBar.View()

	content := lipgloss.JoinVertical(lipgloss.Left,
		headerView,
		divider,
		tableView,
		statusView,
	)

	return styles.AppStyle.Render(content)
}

// GetTable returns the table component
func (d *Dashboard) GetTable() *components.Table {
	return d.table
}

// UpdateStats updates the status bar statistics
func (d *Dashboard) UpdateStats(scanned, found int) {
	d.statusBar.SetStats(scanned, found)
}
