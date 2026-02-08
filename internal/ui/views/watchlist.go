package views

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"stockmap/internal/fetcher"
	"stockmap/internal/screener"
	"stockmap/internal/styles"
	"stockmap/internal/ui/components"
)

// WatchlistView shows only watchlist stocks
type WatchlistView struct {
	width           int
	height          int
	header          *components.Header
	table           *components.Table
	statusBar       *components.StatusBar
	allStocks       []*screener.ScreenResult
	inputActive     bool
	inputText       string
	categoryMode    bool
	categories      []fetcher.Category
	currentCategory int
}

// NewWatchlistView creates a new watchlist view
func NewWatchlistView() *WatchlistView {
	return &WatchlistView{
		header:     components.NewHeader(),
		table:      components.NewTable(),
		statusBar:  components.NewStatusBar(),
		categories: fetcher.DefaultCategories(),
	}
}

// SetSize sets the view dimensions
func (w *WatchlistView) SetSize(width, height int) {
	w.width = width
	w.height = height

	w.header.SetWidth(width)
	w.statusBar.SetWidth(width)

	tableHeight := height - 8
	if tableHeight < 5 {
		tableHeight = 5
	}
	w.table.SetSize(width, tableHeight)
}

// SetResults sets all stock results
func (w *WatchlistView) SetResults(results []*screener.ScreenResult) {
	w.allStocks = results
	w.filterWatchlist()
}

// filterWatchlist filters to only show pinned stocks
func (w *WatchlistView) filterWatchlist() {
	pinned := make([]*screener.ScreenResult, 0)
	for _, stock := range w.allStocks {
		if stock.IsPinned {
			pinned = append(pinned, stock)
		}
	}
	w.table.SetRows(pinned)
}

// Refresh updates the view
func (w *WatchlistView) Refresh() {
	w.filterWatchlist()
}

// MoveUp moves selection up
func (w *WatchlistView) MoveUp() {
	if w.categoryMode {
		if w.currentCategory > 0 {
			w.currentCategory--
		}
	} else {
		w.table.MoveUp()
	}
}

// MoveDown moves selection down
func (w *WatchlistView) MoveDown() {
	if w.categoryMode {
		if w.currentCategory < len(w.categories)-1 {
			w.currentCategory++
		}
	} else {
		w.table.MoveDown()
	}
}

// SelectedResult returns the selected stock
func (w *WatchlistView) SelectedResult() *screener.ScreenResult {
	if w.categoryMode {
		return nil
	}
	return w.table.SelectedRow()
}

// ToggleInput toggles input mode
func (w *WatchlistView) ToggleInput() {
	w.inputActive = !w.inputActive
	if w.inputActive {
		w.categoryMode = false
		w.inputText = ""
	}
}

// ToggleCategoryMode toggles category selection mode
func (w *WatchlistView) ToggleCategoryMode() {
	w.categoryMode = !w.categoryMode
	if w.categoryMode {
		w.inputActive = false
		w.currentCategory = 0
	}
}

// IsCategoryMode returns whether category mode is active
func (w *WatchlistView) IsCategoryMode() bool {
	return w.categoryMode
}

// GetSelectedCategorySymbols returns symbols for the selected category
func (w *WatchlistView) GetSelectedCategorySymbols() []string {
	if w.categoryMode && w.currentCategory >= 0 && w.currentCategory < len(w.categories) {
		return w.categories[w.currentCategory].Symbols
	}
	return nil
}

// GetSelectedCategoryName returns the name of the selected category
func (w *WatchlistView) GetSelectedCategoryName() string {
	if w.categoryMode && w.currentCategory >= 0 && w.currentCategory < len(w.categories) {
		return w.categories[w.currentCategory].Name
	}
	return ""
}

// IsInputActive returns whether input is active
func (w *WatchlistView) IsInputActive() bool {
	return w.inputActive
}

// AddChar adds a character to input
func (w *WatchlistView) AddChar(c rune) {
	w.inputText += strings.ToUpper(string(c))
}

// Backspace removes last character
func (w *WatchlistView) Backspace() {
	if len(w.inputText) > 0 {
		w.inputText = w.inputText[:len(w.inputText)-1]
	}
}

// GetInputSymbol returns the input symbol (trimmed and uppercased)
func (w *WatchlistView) GetInputSymbol() string {
	return strings.TrimSpace(strings.ToUpper(w.inputText))
}

// ClearInput clears the input
func (w *WatchlistView) ClearInput() {
	w.inputText = ""
	w.inputActive = false
}

// View renders the watchlist view
func (w *WatchlistView) View() string {
	var b strings.Builder

	// Header
	title := styles.TitleStyle.Render("* WATCHLIST")
	b.WriteString(centerText(title, w.width))
	b.WriteString("\n")
	b.WriteString(components.RenderDivider(w.width))
	b.WriteString("\n")

	// Category Selection Mode
	if w.categoryMode {
		b.WriteString("\n")
		b.WriteString(styles.TitleStyle.Render("  Add Category to Watchlist"))
		b.WriteString("\n\n")

		// Render category list
		for i, cat := range w.categories {
			prefix := "  "
			nameStyle := styles.TextStyle
			countStyle := styles.MutedStyle()

			if i == w.currentCategory {
				prefix = "> "
				nameStyle = styles.KeyStyle
				countStyle = styles.InfoStyle
			}

			b.WriteString(prefix)
			b.WriteString(nameStyle.Render(cat.Name))
			b.WriteString(" ")
			b.WriteString(countStyle.Render("(" + intToStr(len(cat.Symbols)) + ")"))
			b.WriteString("\n")
		}
		b.WriteString("\n")
		b.WriteString(styles.HelpStyle.Render("  Press [Enter] to add all stocks in category, [ESC] to cancel"))
		return b.String()
	}

	// Input form if active
	if w.inputActive {
		b.WriteString("\n")
		b.WriteString(styles.TitleStyle.Render("  Add Symbol to Watchlist"))
		b.WriteString("\n\n")

		inputBox := styles.KeyStyle.Render("  Symbol: ") +
			styles.InfoStyle.Render(w.inputText) +
			styles.MutedStyle().Render("_")
		b.WriteString(inputBox)
		b.WriteString("\n\n")

		hint := styles.HelpStyle.Render("  Type symbol and press [Enter] to add, [ESC] to cancel")
		b.WriteString(hint)
		b.WriteString("\n")
	} else {
		// Check if empty
		rows := w.table.GetRows()
		if len(rows) == 0 {
			// Empty state
			emptyHeight := w.height - 10
			for i := 0; i < emptyHeight/2; i++ {
				b.WriteString("\n")
			}

			emptyIcon := styles.MutedStyle().Render("*")
			b.WriteString(centerText(emptyIcon, w.width))
			b.WriteString("\n\n")

			emptyText := styles.MutedStyle().Render("No stocks in watchlist")
			b.WriteString(centerText(emptyText, w.width))
			b.WriteString("\n\n")

			hint := styles.HelpStyle.Render("Press [A] to add a stock symbol, or [H] to add by category")
			b.WriteString(centerText(hint, w.width))
			b.WriteString("\n")
		} else {
			// Table
			b.WriteString(w.table.View())
		}
	}

	// Status bar
	rows := w.table.GetRows()
	w.statusBar.SetMessage("Watchlist Mode")
	w.statusBar.SetStats(len(rows), len(rows))
	b.WriteString("\n")
	b.WriteString(w.renderStatusBar())

	return b.String()
}

// renderStatusBar renders a custom status bar for watchlist
func (w *WatchlistView) renderStatusBar() string {
	divider := components.RenderDivider(w.width)

	keys := styles.KeyStyle.Render("[A]") + styles.HelpStyle.Render("dd  ") +
		styles.KeyStyle.Render("[H]") + styles.HelpStyle.Render("Category  ") +
		styles.KeyStyle.Render("[R]") + styles.HelpStyle.Render("emove  ") +
		styles.KeyStyle.Render("[D]") + styles.HelpStyle.Render("etails  ") +
		styles.KeyStyle.Render("[ESC]") + styles.HelpStyle.Render(" Back  ") +
		styles.KeyStyle.Render("[Q]") + styles.HelpStyle.Render("uit")

	count := len(w.table.GetRows())
	stats := styles.MutedStyle().Render("│ ") +
		styles.StatusItemStyle.Render(intToStr(count)+" stocks pinned")

	return lipgloss.JoinVertical(lipgloss.Left, divider, keys+stats)
}

// intToStr converts int to string without fmt
func intToStr(n int) string {
	if n == 0 {
		return "0"
	}

	negative := n < 0
	if negative {
		n = -n
	}

	digits := ""
	for n > 0 {
		digits = string(rune('0'+n%10)) + digits
		n /= 10
	}

	if negative {
		digits = "-" + digits
	}

	return digits
}

// GetTable returns the table component
func (w *WatchlistView) GetTable() *components.Table {
	return w.table
}
