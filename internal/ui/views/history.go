package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"stockmap/internal/history"
	"stockmap/internal/styles"
	"stockmap/internal/ui/components"
)

// HistoryView shows saved scan history
type HistoryView struct {
	width     int
	height    int
	manager   *history.Manager
	records   []*history.ScanRecord
	cursor    int
	offset    int
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
	title := styles.TitleStyle.Render("📁 SCAN HISTORY")
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

		emptyIcon := styles.MutedStyle().Render("📭")
		b.WriteString(centerText(emptyIcon, h.width))
		b.WriteString("\n\n")

		emptyText := styles.MutedStyle().Render("No scan history yet")
		b.WriteString(centerText(emptyText, h.width))
		b.WriteString("\n\n")

		hint := styles.HelpStyle.Render("Press [S] to start a new scan")
		b.WriteString(centerText(hint, h.width))
		b.WriteString("\n")
	} else {
		// Table header
		header := fmt.Sprintf("  %-4s  %-20s  %-10s  %-10s  %s",
			"#",
			"DATE & TIME",
			"SCANNED",
			"FOUND",
			"AGE",
		)
		b.WriteString(styles.TableHeaderStyle.Render(header))
		b.WriteString("\n")
		b.WriteString(components.RenderDivider(h.width))
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

			// Format row
			num := fmt.Sprintf("%d", i+1)
			dateTime := record.Timestamp.Format("2006-01-02 15:04:05")
			scanned := fmt.Sprintf("%d", record.TotalScanned)
			found := fmt.Sprintf("%d", record.TotalFound)
			age := history.FormatTimestamp(record.Timestamp)

			row := fmt.Sprintf("  %-4s  %-20s  %-10s  %-10s  %s",
				num,
				dateTime,
				scanned,
				found,
				age,
			)

			if selected {
				row = styles.TableRowSelectedStyle.Width(h.width).Render(row)
			} else {
				// Color code found count
				if record.TotalFound > 10 {
					row = styles.TableRowStyle.Render(row)
				} else if record.TotalFound > 0 {
					row = styles.TableRowStyle.Render(row)
				} else {
					row = styles.MutedStyle().Render(row)
				}
			}

			b.WriteString(row)
			b.WriteString("\n")
		}

		// Fill remaining space
		for i := endIdx - h.offset; i < visibleRows; i++ {
			b.WriteString("\n")
		}
	}

	// Status bar
	b.WriteString("\n")
	b.WriteString(h.renderStatusBar())

	return b.String()
}

// renderStatusBar renders the status bar
func (h *HistoryView) renderStatusBar() string {
	divider := components.RenderDivider(h.width)

	keys := styles.KeyStyle.Render("[Enter]") + styles.HelpStyle.Render(" Load  ") +
		styles.KeyStyle.Render("[X]") + styles.HelpStyle.Render(" Delete  ") +
		styles.KeyStyle.Render("[ESC]") + styles.HelpStyle.Render(" Back  ") +
		styles.KeyStyle.Render("[Q]") + styles.HelpStyle.Render("uit")

	count := fmt.Sprintf("%d saved scans", len(h.records))
	stats := styles.MutedStyle().Render(" │ ") + styles.StatusItemStyle.Render(count)

	return lipgloss.JoinVertical(lipgloss.Left, divider, keys+stats)
}

// Count returns the number of history records
func (h *HistoryView) Count() int {
	return len(h.records)
}
