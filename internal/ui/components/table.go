package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"stockmap/internal/screener"
	"stockmap/internal/styles"
)

// Column represents a table column
type Column struct {
	Title string
	Width int
}

// Table represents the stock table component
type Table struct {
	columns     []Column
	rows        []*screener.ScreenResult
	cursor      int
	offset      int
	height      int
	width       int
	showDetails bool
}

// NewTable creates a new table component
func NewTable() *Table {
	return &Table{
		columns: []Column{
			{Title: "", Width: 2}, // Pin indicator
			{Title: "TICKER", Width: 8},
			{Title: "PRICE", Width: 10},
			{Title: "CHANGE", Width: 9},
			{Title: "TP", Width: 10},
			{Title: "SL", Width: 10},
			{Title: "RSI", Width: 6},
			{Title: "V%", Width: 6}, // Volatility (was PBV)
			{Title: "SCORE", Width: 12},
		},
		height: 15,
	}
}

// SetSize sets the table dimensions
func (t *Table) SetSize(width, height int) {
	t.width = width
	t.height = height
}

// SetRows updates the table data
func (t *Table) SetRows(rows []*screener.ScreenResult) {
	t.rows = rows
	if t.cursor >= len(rows) {
		t.cursor = len(rows) - 1
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
	if t.cursor < len(t.rows)-1 {
		t.cursor++
		visibleRows := t.height - 3 // Account for header and borders
		if t.cursor >= t.offset+visibleRows {
			t.offset = t.cursor - visibleRows + 1
		}
	}
}

// SelectedRow returns the currently selected row
func (t *Table) SelectedRow() *screener.ScreenResult {
	if t.cursor >= 0 && t.cursor < len(t.rows) {
		return t.rows[t.cursor]
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

	// Render header
	header := t.renderHeader()
	b.WriteString(header)
	b.WriteString("\n")

	// Render separator
	sep := t.renderSeparator()
	b.WriteString(sep)
	b.WriteString("\n")

	// Calculate visible rows
	visibleRows := t.height - 3
	if visibleRows < 1 {
		visibleRows = 10
	}

	endIdx := t.offset + visibleRows
	if endIdx > len(t.rows) {
		endIdx = len(t.rows)
	}

	// Render rows
	for i := t.offset; i < endIdx; i++ {
		row := t.renderRow(i, i == t.cursor)
		b.WriteString(row)
		b.WriteString("\n")
	}

	// Fill remaining space
	for i := endIdx - t.offset; i < visibleRows; i++ {
		b.WriteString(strings.Repeat(" ", t.width))
		b.WriteString("\n")
	}

	return b.String()
}

// renderHeader renders the table header
func (t *Table) renderHeader() string {
	var cells []string

	for _, col := range t.columns {
		cell := styles.TableHeaderStyle.
			Width(col.Width).
			Render(col.Title)
		cells = append(cells, cell)
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, cells...)
}

// renderSeparator renders the header separator
func (t *Table) renderSeparator() string {
	var cells []string

	for _, col := range t.columns {
		sep := strings.Repeat("─", col.Width)
		cell := lipgloss.NewStyle().
			Foreground(styles.ColorMuted).
			Width(col.Width).
			Render(sep)
		cells = append(cells, cell)
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, cells...)
}

// renderRow renders a single table row
func (t *Table) renderRow(idx int, selected bool) string {
	if idx >= len(t.rows) {
		return ""
	}

	row := t.rows[idx]
	var cells []string

	// Base style for the row
	baseStyle := styles.TableRowStyle
	if selected {
		baseStyle = styles.TableRowSelectedStyle
	} else if row.IsPinned {
		baseStyle = styles.TableRowPinnedStyle
	}

	// Pin indicator
	pinText := "  "
	if row.IsPinned {
		pinText = styles.PinStyle.Render("★ ")
	}
	cells = append(cells, lipgloss.NewStyle().Width(2).Render(pinText))

	// Ticker
	tickerStyle := styles.TickerStyle
	if selected {
		tickerStyle = styles.TableRowSelectedStyle.Bold(true)
	}
	cells = append(cells, tickerStyle.Width(8).Render(row.Symbol))

	// Price
	priceText := fmt.Sprintf("$%.2f", row.Price)
	cells = append(cells, baseStyle.Width(10).Render(priceText))

	// Change
	changeText := fmt.Sprintf("%+.2f%%", row.ChangePercent)
	changeStyle := baseStyle
	if !selected {
		if row.ChangePercent >= 0 {
			changeStyle = styles.PriceUpStyle
		} else {
			changeStyle = styles.PriceDownStyle
		}
	}
	cells = append(cells, changeStyle.Width(9).Render(changeText))

	// Take Profit
	tpText := fmt.Sprintf("$%.2f", row.TakeProfit)
	tpStyle := baseStyle
	if !selected {
		tpStyle = styles.PriceUpStyle
	}
	cells = append(cells, tpStyle.Width(10).Render(tpText))

	// Stop Loss
	slText := fmt.Sprintf("$%.2f", row.StopLoss)
	slStyle := baseStyle
	if !selected {
		slStyle = styles.PriceDownStyle
	}
	cells = append(cells, slStyle.Width(10).Render(slText))

	// RSI
	rsiText := fmt.Sprintf("%.1f", row.RSI)
	rsiStyle := baseStyle
	if !selected {
		if row.RSI < 30 {
			rsiStyle = styles.RSIOversoldStyle
		} else if row.RSI > 70 {
			rsiStyle = styles.RSIOverboughtStyle
		}
	}
	cells = append(cells, rsiStyle.Width(6).Render(rsiText))

	// Volatility (V column)
	volText := fmt.Sprintf("%.1f", row.Volatility)
	volStyle := baseStyle
	if !selected {
		if row.Volatility < 20 {
			volStyle = styles.PBVLowStyle // Low volatility = good (green)
		} else if row.Volatility > 40 {
			volStyle = styles.PBVHighStyle // High volatility = bad (red)
		}
	}
	cells = append(cells, volStyle.Width(6).Render(volText))

	// Score with bar
	scoreText := fmt.Sprintf("%.0f ", row.ConfluenceScore)
	barWidth := 6
	if !selected {
		scoreText += styles.ScoreBar(row.ConfluenceScore, barWidth)
		cells = append(cells, lipgloss.NewStyle().Width(12).Render(scoreText))
	} else {
		bar := strings.Repeat("█", int(row.ConfluenceScore/100*float64(barWidth)))
		cells = append(cells, baseStyle.Width(12).Render(scoreText+bar))
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, cells...)
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
