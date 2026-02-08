package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/febritecno/stockmap-cli/internal/history"
	"github.com/febritecno/stockmap-cli/internal/styles"
	"github.com/febritecno/stockmap-cli/internal/ui/components"
)

// HistoryView shows saved scan history
type HistoryView struct {
	width   int
	height  int
	manager *history.Manager
	records []*history.ScanRecord
	cursor  int
	offset  int
}

// NewHistoryView creates a new history view
func NewHistoryView() *HistoryView {
	return &HistoryView{
		manager: history.NewManager(),
		records: make([]*history.ScanRecord, 0),
	}
}

// SetSize sets the view dimensions
func (h *HistoryView) SetSize(width, height int) {
	h.width = width
	h.height = height
}

// Refresh reloads history from disk
func (h *HistoryView) Refresh() error {
	records, err := h.manager.List()
	if err != nil {
		return err
	}
	h.records = records

	// Reset cursor if out of bounds
	if h.cursor >= len(h.records) {
		h.cursor = len(h.records) - 1
	}
	if h.cursor < 0 {
		h.cursor = 0
	}

	return nil
}

// MoveUp moves selection up
func (h *HistoryView) MoveUp() {
	if h.cursor > 0 {
		h.cursor--
		if h.cursor < h.offset {
			h.offset = h.cursor
		}
	}
}

// MoveDown moves selection down
func (h *HistoryView) MoveDown() {
	if h.cursor < len(h.records)-1 {
		h.cursor++
		visibleRows := h.height - 10
		if visibleRows < 5 {
			visibleRows = 5
		}
		if h.cursor >= h.offset+visibleRows {
			h.offset = h.cursor - visibleRows + 1
		}
	}
}

// SelectedRecord returns the currently selected record
func (h *HistoryView) SelectedRecord() *history.ScanRecord {
	if h.cursor >= 0 && h.cursor < len(h.records) {
		return h.records[h.cursor]
	}
	return nil
}

// LoadSelected loads the full data for selected record
func (h *HistoryView) LoadSelected() (*history.ScanRecord, error) {
	selected := h.SelectedRecord()
	if selected == nil {
		return nil, fmt.Errorf("no record selected")
	}
	return h.manager.Load(selected.ID)
}

// DeleteSelected deletes the selected record
func (h *HistoryView) DeleteSelected() error {
	selected := h.SelectedRecord()
	if selected == nil {
		return fmt.Errorf("no record selected")
	}

	if err := h.manager.Delete(selected.ID); err != nil {
		return err
	}

	return h.Refresh()
}

// GetManager returns the history manager
func (h *HistoryView) GetManager() *history.Manager {
	return h.manager
}

// View renders the history view
func (h *HistoryView) View() string {
	var b strings.Builder

	// Header
	title := styles.TitleStyle.Render("SCAN HISTORY")
	b.WriteString(centerText(title, h.width))
	b.WriteString("\n")
	b.WriteString(components.RenderDivider(h.width))
	b.WriteString("\n\n")

	// Check if empty
	if len(h.records) == 0 {
		emptyHeight := h.height - 12
		for i := 0; i < emptyHeight/2; i++ {
			b.WriteString("\n")
		}

		emptyText := styles.MutedStyle().Render("No scan history yet")
		b.WriteString(centerText(emptyText, h.width))
		b.WriteString("\n\n")

		hint := styles.HelpStyle.Render("Press [S] to start a new scan")
		b.WriteString(centerText(hint, h.width))
		b.WriteString("\n")
	} else {
		// Render table with lipgloss for proper alignment
		b.WriteString(h.renderTable())
	}

	// Status bar
	b.WriteString("\n")
	b.WriteString(h.renderStatusBar())

	return b.String()
}

// renderTable renders the history table with proper responsive columns
func (h *HistoryView) renderTable() string {
	var b strings.Builder

	// Calculate available width
	availWidth := h.width - 2
	if availWidth < 60 {
		availWidth = 60
	}

	// Define columns with min and preferred widths
	type col struct {
		title     string
		minWidth  int
		prefWidth int
	}

	// Different layouts based on width
	var cols []col
	if availWidth < 80 {
		// Compact: #, DateTime, Scanned, Found
		cols = []col{
			{"#", 3, 4},
			{"DATE/TIME", 14, 18},
			{"SCAN", 6, 8},
			{"FOUND", 6, 8},
		}
	} else {
		// Full: #, Date, Time, Scanned, Found, Age
		cols = []col{
			{"#", 3, 5},
			{"DATE", 10, 12},
			{"TIME", 8, 10},
			{"SCANNED", 8, 10},
			{"FOUND", 6, 8},
			{"AGE", 10, 16},
		}
	}

	// Calculate total min width
	totalMin := 0
	for _, c := range cols {
		totalMin += c.minWidth + 1 // +1 for spacing
	}

	// Determine actual widths
	widths := make([]int, len(cols))
	remaining := availWidth - totalMin
	for i, c := range cols {
		widths[i] = c.minWidth
		if remaining > 0 {
			extra := (c.prefWidth - c.minWidth)
			if extra > remaining/(len(cols)-i) {
				extra = remaining / (len(cols) - i)
			}
			if extra > 0 {
				widths[i] += extra
				remaining -= extra
			}
		}
	}

	// Render header
	headerCells := make([]string, len(cols))
	for i, c := range cols {
		headerCells[i] = lipgloss.NewStyle().
			Width(widths[i]).
			Bold(true).
			Foreground(styles.ColorPrimary).
			Render(c.title)
	}
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, headerCells...))
	b.WriteString("\n")

	// Render separator
	sepCells := make([]string, len(cols))
	for i := range cols {
		sepCells[i] = lipgloss.NewStyle().
			Width(widths[i]).
			Foreground(styles.ColorMuted).
			Render(strings.Repeat("─", widths[i]))
	}
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, sepCells...))
	b.WriteString("\n")

	// Calculate visible rows
	visibleRows := h.height - 12
	if visibleRows < 5 {
		visibleRows = 5
	}

	endIdx := h.offset + visibleRows
	if endIdx > len(h.records) {
		endIdx = len(h.records)
	}

	// Render rows
	for i := h.offset; i < endIdx; i++ {
		record := h.records[i]
		selected := i == h.cursor

		var rowCells []string

		if availWidth < 80 {
			// Compact format
			rowCells = []string{
				fmt.Sprintf("%d", i+1),
				record.Timestamp.Format("01-02 15:04"),
				fmt.Sprintf("%d", record.TotalScanned),
				fmt.Sprintf("%d", record.TotalFound),
			}
		} else {
			// Full format
			rowCells = []string{
				fmt.Sprintf("%d", i+1),
				record.Timestamp.Format("2006-01-02"),
				record.Timestamp.Format("15:04:05"),
				fmt.Sprintf("%d", record.TotalScanned),
				fmt.Sprintf("%d", record.TotalFound),
				history.FormatTimestamp(record.Timestamp),
			}
		}

		// Style each cell
		styledCells := make([]string, len(rowCells))
		for j, cell := range rowCells {
			style := lipgloss.NewStyle().Width(widths[j])

			if selected {
				style = style.Bold(true).
					Foreground(styles.ColorBackground).
					Background(styles.ColorPrimary)
			} else if record.TotalFound == 0 {
				style = style.Foreground(styles.ColorMuted)
			} else {
				style = style.Foreground(styles.ColorText)
			}

			styledCells[j] = style.Render(cell)
		}

		b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, styledCells...))
		b.WriteString("\n")
	}

	// Fill remaining space
	for i := endIdx - h.offset; i < visibleRows; i++ {
		b.WriteString("\n")
	}

	return b.String()
}

// renderStatusBar renders the status bar
func (h *HistoryView) renderStatusBar() string {
	divider := components.RenderDivider(h.width)

	var keys string
	if h.width < 60 {
		// Compact
		keys = styles.KeyStyle.Render("[Enter]") + " " +
			styles.KeyStyle.Render("[X]") + " " +
			styles.KeyStyle.Render("[ESC]")
	} else {
		// Full
		keys = styles.KeyStyle.Render("[Enter]") + styles.HelpStyle.Render(" Load  ") +
			styles.KeyStyle.Render("[X]") + styles.HelpStyle.Render(" Delete  ") +
			styles.KeyStyle.Render("[ESC]") + styles.HelpStyle.Render(" Back")
	}

	count := fmt.Sprintf("%d scans", len(h.records))
	stats := styles.MutedStyle().Render(" | ") + styles.StatusItemStyle.Render(count)

	return lipgloss.JoinVertical(lipgloss.Left, divider, keys+stats)
}

// Count returns the number of history records
func (h *HistoryView) Count() int {
	return len(h.records)
}
