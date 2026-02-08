package components

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/febritecno/stockmap-cli/internal/screener"
	"github.com/febritecno/stockmap-cli/internal/styles"
)

// SortColumn represents which column to sort by
type SortColumn int

const (
	SortByScore SortColumn = iota
	SortByTicker
	SortByPrice
	SortByChange
	SortByRSI
	SortByVolatility
)

// Column represents a table column
type Column struct {
	Title    string
	Width    int
	MinWidth int
	MaxWidth int
}

// Table represents the stock table component
type Table struct {
	columns      []Column
	rows         []*screener.ScreenResult
	filteredRows []*screener.ScreenResult
	cursor       int
	offset       int
	height       int
	width        int
	showDetails  bool
	searchQuery  string
	sortColumn   SortColumn
	sortAsc      bool
	compactMode  bool // Use compact layout for narrow screens
}

// NewTable creates a new table component
func NewTable() *Table {
	return &Table{
		columns: []Column{
			{Title: "", MinWidth: 2, MaxWidth: 2},       // Pin
			{Title: "TICKER", MinWidth: 5, MaxWidth: 8}, // Ticker
			{Title: "PRICE", MinWidth: 7, MaxWidth: 10}, // Price
			{Title: "CHG%", MinWidth: 6, MaxWidth: 8},   // Change
			{Title: "TP", MinWidth: 7, MaxWidth: 10},    // Take Profit
			{Title: "SL", MinWidth: 7, MaxWidth: 10},    // Stop Loss
			{Title: "RSI", MinWidth: 4, MaxWidth: 6},    // RSI
			{Title: "VL%", MinWidth: 4, MaxWidth: 6},    // Volatility
			{Title: "SCORE", MinWidth: 6, MaxWidth: 12}, // Score
		},
		height: 15,
	}
}

// SetSize sets the table dimensions and recalculates column widths
func (t *Table) SetSize(width, height int) {
	t.width = width
	t.height = height
	t.recalculateColumns()
}

// recalculateColumns adjusts column widths based on available width
func (t *Table) recalculateColumns() {
	availableWidth := t.width - 2 // Small padding

	// Calculate total min and max widths
	totalMin := 0
	totalMax := 0
	for _, col := range t.columns {
		totalMin += col.MinWidth
		totalMax += col.MaxWidth
	}

	// Determine if we need compact mode
	t.compactMode = availableWidth < totalMin+10

	if t.compactMode {
		// Very narrow - use minimal widths
		for i := range t.columns {
			t.columns[i].Width = t.columns[i].MinWidth
		}
		return
	}

	if availableWidth <= totalMin {
		// Just use min widths
		for i := range t.columns {
			t.columns[i].Width = t.columns[i].MinWidth
		}
		return
	}

	if availableWidth >= totalMax {
		// Use max widths
		for i := range t.columns {
			t.columns[i].Width = t.columns[i].MaxWidth
		}
		return
	}

	// Distribute extra space proportionally
	extra := availableWidth - totalMin
	expandable := totalMax - totalMin

	for i := range t.columns {
		t.columns[i].Width = t.columns[i].MinWidth
		if expandable > 0 {
			colExtra := t.columns[i].MaxWidth - t.columns[i].MinWidth
			allocated := (colExtra * extra) / expandable
			t.columns[i].Width += allocated
		}
	}
}

// getColWidth returns the width for column index
func (t *Table) getColWidth(idx int) int {
	if idx >= 0 && idx < len(t.columns) {
		return t.columns[idx].Width
	}
	return 8
}

// SetRows updates the table data
func (t *Table) SetRows(rows []*screener.ScreenResult) {
	t.rows = rows
	if t.searchQuery != "" {
		t.applyFilter()
	}
	t.applySort()

	displayRows := t.getDisplayRows()
	if t.cursor >= len(displayRows) {
		t.cursor = len(displayRows) - 1
	}
	if t.cursor < 0 {
		t.cursor = 0
	}
}

// MoveUp moves cursor up
func (t *Table) MoveUp() {
	if t.cursor > 0 {
		t.cursor--
		if t.cursor < t.offset {
			t.offset = t.cursor
		}
	}
}

// MoveDown moves cursor down
func (t *Table) MoveDown() {
	displayRows := t.getDisplayRows()
	if t.cursor < len(displayRows)-1 {
		t.cursor++
		visibleRows := t.height - 3
		if t.cursor >= t.offset+visibleRows {
			t.offset = t.cursor - visibleRows + 1
		}
	}
}

// SelectedRow returns the currently selected row
func (t *Table) SelectedRow() *screener.ScreenResult {
	displayRows := t.getDisplayRows()
	if t.cursor >= 0 && t.cursor < len(displayRows) {
		return displayRows[t.cursor]
	}
	return nil
}

// SelectedIndex returns the current cursor position
func (t *Table) SelectedIndex() int {
	return t.cursor
}

// View renders the table
func (t *Table) View() string {
	var b strings.Builder

	displayRows := t.getDisplayRows()

	// Render header
	b.WriteString(t.renderHeader())
	b.WriteString("\n")

	// Render separator
	b.WriteString(t.renderSeparator())
	b.WriteString("\n")

	// Calculate visible rows
	visibleRows := t.height - 3
	if visibleRows < 1 {
		visibleRows = 10
	}

	endIdx := t.offset + visibleRows
	if endIdx > len(displayRows) {
		endIdx = len(displayRows)
	}

	// Render rows
	for i := t.offset; i < endIdx; i++ {
		b.WriteString(t.renderRowData(displayRows[i], i == t.cursor))
		b.WriteString("\n")
	}

	// Fill remaining space
	for i := endIdx - t.offset; i < visibleRows; i++ {
		b.WriteString("\n")
	}

	return b.String()
}

// renderHeader renders the table header
func (t *Table) renderHeader() string {
	cells := make([]string, len(t.columns))

	for i, col := range t.columns {
		cells[i] = lipgloss.NewStyle().
			Width(col.Width).
			Bold(true).
			Foreground(styles.ColorPrimary).
			Render(col.Title)
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, cells...)
}

// renderSeparator renders the header separator
func (t *Table) renderSeparator() string {
	cells := make([]string, len(t.columns))

	for i, col := range t.columns {
		cells[i] = lipgloss.NewStyle().
			Width(col.Width).
			Foreground(styles.ColorMuted).
			Render(strings.Repeat("─", col.Width))
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, cells...)
}

// renderRowData renders a row from data
func (t *Table) renderRowData(row *screener.ScreenResult, selected bool) string {
	cells := make([]string, len(t.columns))

	// Determine base style
	baseStyle := lipgloss.NewStyle().Foreground(styles.ColorText)
	if selected {
		baseStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(styles.ColorBackground).
			Background(styles.ColorPrimary)
	} else if row.IsPinned {
		baseStyle = lipgloss.NewStyle().Foreground(styles.ColorWarning)
	}

	// Column 0: Pin indicator
	pinText := " "
	if row.IsPinned {
		if selected {
			pinText = "*"
		} else {
			pinText = styles.PinStyle.Render("*")
		}
	}
	cells[0] = lipgloss.NewStyle().Width(t.getColWidth(0)).Render(pinText)

	// Column 1: Ticker
	tickerStyle := baseStyle.Bold(true)
	if !selected {
		tickerStyle = lipgloss.NewStyle().Bold(true).Foreground(styles.ColorCyan)
	}
	cells[1] = tickerStyle.Width(t.getColWidth(1)).Render(truncate(row.Symbol, t.getColWidth(1)))

	// Column 2: Price
	priceText := fmt.Sprintf("$%.2f", row.Price)
	if t.getColWidth(2) < 8 {
		priceText = fmt.Sprintf("%.1f", row.Price)
	}
	cells[2] = baseStyle.Width(t.getColWidth(2)).Render(truncate(priceText, t.getColWidth(2)))

	// Column 3: Change %
	changeText := fmt.Sprintf("%+.1f%%", row.ChangePercent)
	changeStyle := baseStyle
	if !selected {
		if row.ChangePercent >= 0 {
			changeStyle = lipgloss.NewStyle().Foreground(styles.ColorSuccess)
		} else {
			changeStyle = lipgloss.NewStyle().Foreground(styles.ColorDanger)
		}
	}
	cells[3] = changeStyle.Width(t.getColWidth(3)).Render(truncate(changeText, t.getColWidth(3)))

	// Column 4: Take Profit
	tpText := fmt.Sprintf("$%.2f", row.TakeProfit)
	if t.getColWidth(4) < 8 {
		tpText = fmt.Sprintf("%.1f", row.TakeProfit)
	}
	tpStyle := baseStyle
	if !selected {
		tpStyle = lipgloss.NewStyle().Foreground(styles.ColorSuccess)
	}
	cells[4] = tpStyle.Width(t.getColWidth(4)).Render(truncate(tpText, t.getColWidth(4)))

	// Column 5: Stop Loss
	slText := fmt.Sprintf("$%.2f", row.StopLoss)
	if t.getColWidth(5) < 8 {
		slText = fmt.Sprintf("%.1f", row.StopLoss)
	}
	slStyle := baseStyle
	if !selected {
		slStyle = lipgloss.NewStyle().Foreground(styles.ColorDanger)
	}
	cells[5] = slStyle.Width(t.getColWidth(5)).Render(truncate(slText, t.getColWidth(5)))

	// Column 6: RSI
	rsiText := fmt.Sprintf("%.0f", row.RSI)
	rsiStyle := baseStyle
	if !selected {
		if row.RSI < 30 {
			rsiStyle = lipgloss.NewStyle().Foreground(styles.ColorSuccess)
		} else if row.RSI > 70 {
			rsiStyle = lipgloss.NewStyle().Foreground(styles.ColorDanger)
		} else {
			rsiStyle = lipgloss.NewStyle().Foreground(styles.ColorText)
		}
	}
	cells[6] = rsiStyle.Width(t.getColWidth(6)).Render(truncate(rsiText, t.getColWidth(6)))

	// Column 7: Volatility
	volText := fmt.Sprintf("%.0f", row.Volatility)
	volStyle := baseStyle
	if !selected {
		if row.Volatility < 20 {
			volStyle = lipgloss.NewStyle().Foreground(styles.ColorSuccess)
		} else if row.Volatility > 40 {
			volStyle = lipgloss.NewStyle().Foreground(styles.ColorDanger)
		} else {
			volStyle = lipgloss.NewStyle().Foreground(styles.ColorText)
		}
	}
	cells[7] = volStyle.Width(t.getColWidth(7)).Render(truncate(volText, t.getColWidth(7)))

	// Column 8: Score with bar
	scoreWidth := t.getColWidth(8)
	scoreText := fmt.Sprintf("%.0f", row.ConfluenceScore)

	if scoreWidth >= 10 && !selected {
		// Show bar
		barWidth := scoreWidth - 4
		if barWidth > 6 {
			barWidth = 6
		}
		scoreText += " " + renderMiniBar(row.ConfluenceScore, barWidth)
		cells[8] = lipgloss.NewStyle().Width(scoreWidth).Render(scoreText)
	} else {
		// Just number
		scoreStyle := baseStyle
		if !selected {
			if row.ConfluenceScore >= 75 {
				scoreStyle = lipgloss.NewStyle().Bold(true).Foreground(styles.ColorSuccess)
			} else if row.ConfluenceScore >= 50 {
				scoreStyle = lipgloss.NewStyle().Foreground(styles.ColorWarning)
			} else {
				scoreStyle = lipgloss.NewStyle().Foreground(styles.ColorDanger)
			}
		}
		cells[8] = scoreStyle.Width(scoreWidth).Render(truncate(scoreText, scoreWidth))
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, cells...)
}

// renderMiniBar renders a small score bar
func renderMiniBar(score float64, width int) string {
	filled := int(score / 100 * float64(width))
	if filled > width {
		filled = width
	}

	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)

	var style lipgloss.Style
	if score >= 75 {
		style = lipgloss.NewStyle().Foreground(styles.ColorSuccess)
	} else if score >= 50 {
		style = lipgloss.NewStyle().Foreground(styles.ColorWarning)
	} else {
		style = lipgloss.NewStyle().Foreground(styles.ColorDanger)
	}

	return style.Render(bar)
}

// truncate truncates string to fit width
func truncate(s string, width int) string {
	if len(s) <= width {
		return s
	}
	if width <= 1 {
		return s[:width]
	}
	return s[:width-1] + "."
}

// GetRows returns all rows
func (t *Table) GetRows() []*screener.ScreenResult {
	return t.rows
}

// SetCursor sets the cursor position
func (t *Table) SetCursor(pos int) {
	if pos >= 0 && pos < len(t.rows) {
		t.cursor = pos
	}
}

// SetSearch sets the search query and filters rows
func (t *Table) SetSearch(query string) {
	t.searchQuery = strings.ToUpper(query)
	t.applyFilter()
}

// ClearSearch clears the search query
func (t *Table) ClearSearch() {
	t.searchQuery = ""
	t.filteredRows = nil
}

// GetSearchQuery returns the current search query
func (t *Table) GetSearchQuery() string {
	return t.searchQuery
}

// applyFilter applies the search filter to rows
func (t *Table) applyFilter() {
	if t.searchQuery == "" {
		t.filteredRows = nil
		return
	}

	t.filteredRows = make([]*screener.ScreenResult, 0)
	for _, row := range t.rows {
		if strings.Contains(strings.ToUpper(row.Symbol), t.searchQuery) ||
			strings.Contains(strings.ToUpper(row.Name), t.searchQuery) {
			t.filteredRows = append(t.filteredRows, row)
		}
	}

	t.cursor = 0
	t.offset = 0
}

// CycleSort cycles through sort columns
func (t *Table) CycleSort() {
	t.sortColumn = (t.sortColumn + 1) % 6
	t.sortAsc = false
	t.applySort()
}

// ToggleSortDirection toggles sort direction
func (t *Table) ToggleSortDirection() {
	t.sortAsc = !t.sortAsc
	t.applySort()
}

// GetSortInfo returns current sort column name and direction
func (t *Table) GetSortInfo() (string, bool) {
	names := []string{"Score", "Ticker", "Price", "Change", "RSI", "Volatility"}
	return names[t.sortColumn], t.sortAsc
}

// applySort sorts the rows based on current sort column
func (t *Table) applySort() {
	rowsToSort := t.rows
	if t.filteredRows != nil {
		rowsToSort = t.filteredRows
	}

	sort.Slice(rowsToSort, func(i, j int) bool {
		// Pinned items always first
		if rowsToSort[i].IsPinned && !rowsToSort[j].IsPinned {
			return true
		}
		if !rowsToSort[i].IsPinned && rowsToSort[j].IsPinned {
			return false
		}

		var less bool
		switch t.sortColumn {
		case SortByScore:
			less = rowsToSort[i].ConfluenceScore < rowsToSort[j].ConfluenceScore
		case SortByTicker:
			less = rowsToSort[i].Symbol < rowsToSort[j].Symbol
		case SortByPrice:
			less = rowsToSort[i].Price < rowsToSort[j].Price
		case SortByChange:
			less = rowsToSort[i].ChangePercent < rowsToSort[j].ChangePercent
		case SortByRSI:
			less = rowsToSort[i].RSI < rowsToSort[j].RSI
		case SortByVolatility:
			less = rowsToSort[i].Volatility < rowsToSort[j].Volatility
		}

		if t.sortAsc {
			return less
		}
		return !less
	})
}

// getDisplayRows returns rows to display (filtered or all)
func (t *Table) getDisplayRows() []*screener.ScreenResult {
	if t.filteredRows != nil {
		return t.filteredRows
	}
	return t.rows
}
