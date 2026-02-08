package views

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"stockmap/internal/screener"
	"stockmap/internal/styles"
)

// FilterField represents which field is currently being edited
type FilterField int

const (
	FilterFieldMinRSI FilterField = iota
	FilterFieldMaxRSI
	FilterFieldMaxPBV
	FilterFieldMinScore
	FilterFieldCount // Sentinel for counting fields
)

// FilterView allows editing filter criteria
type FilterView struct {
	width        int
	height       int
	criteria     screener.FilterCriteria
	selectedRow  int
	inputActive  bool
	inputBuffer  string
	currentField FilterField
}

// NewFilterView creates a new filter view
func NewFilterView() *FilterView {
	return &FilterView{
		criteria: screener.DefaultCriteria(),
	}
}

// SetSize sets the view dimensions
func (f *FilterView) SetSize(width, height int) {
	f.width = width
	f.height = height
}

// SetCriteria sets the current filter criteria
func (f *FilterView) SetCriteria(c screener.FilterCriteria) {
	f.criteria = c
}

// GetCriteria returns the current filter criteria
func (f *FilterView) GetCriteria() screener.FilterCriteria {
	return f.criteria
}

// MoveUp moves selection up
func (f *FilterView) MoveUp() {
	if f.selectedRow > 0 {
		f.selectedRow--
	}
}

// MoveDown moves selection down
func (f *FilterView) MoveDown() {
	if f.selectedRow < int(FilterFieldCount)-1 {
		f.selectedRow++
	}
}

// IsInputActive returns whether input mode is active
func (f *FilterView) IsInputActive() bool {
	return f.inputActive
}

// ToggleInput toggles input mode for current field
func (f *FilterView) ToggleInput() {
	f.inputActive = !f.inputActive
	if f.inputActive {
		f.currentField = FilterField(f.selectedRow)
		// Pre-fill with current value
		switch f.currentField {
		case FilterFieldMinRSI:
			f.inputBuffer = fmt.Sprintf("%.0f", f.criteria.MinRSI)
		case FilterFieldMaxRSI:
			f.inputBuffer = fmt.Sprintf("%.0f", f.criteria.MaxRSI)
		case FilterFieldMaxPBV:
			f.inputBuffer = fmt.Sprintf("%.1f", f.criteria.MaxPBV)
		case FilterFieldMinScore:
			f.inputBuffer = fmt.Sprintf("%.0f", f.criteria.MinConfluence)
		}
	}
}

// ClearInput clears input and exits input mode
func (f *FilterView) ClearInput() {
	f.inputBuffer = ""
	f.inputActive = false
}

// AddChar adds a character to input
func (f *FilterView) AddChar(c rune) {
	// Only allow digits and decimal point
	if (c >= '0' && c <= '9') || c == '.' {
		f.inputBuffer += string(c)
	}
}

// Backspace removes last character
func (f *FilterView) Backspace() {
	if len(f.inputBuffer) > 0 {
		f.inputBuffer = f.inputBuffer[:len(f.inputBuffer)-1]
	}
}

// SubmitInput applies the current input value
func (f *FilterView) SubmitInput() bool {
	if f.inputBuffer == "" {
		f.inputActive = false
		return false
	}

	val, err := strconv.ParseFloat(f.inputBuffer, 64)
	if err != nil {
		f.inputActive = false
		return false
	}

	switch f.currentField {
	case FilterFieldMinRSI:
		if val >= 0 && val <= 100 && val <= f.criteria.MaxRSI {
			f.criteria.MinRSI = val
		}
	case FilterFieldMaxRSI:
		if val >= 0 && val <= 100 && val >= f.criteria.MinRSI {
			f.criteria.MaxRSI = val
		}
	case FilterFieldMaxPBV:
		if val >= 0 {
			f.criteria.MaxPBV = val
		}
	case FilterFieldMinScore:
		if val >= 0 && val <= 100 {
			f.criteria.MinConfluence = val
		}
	}

	f.inputBuffer = ""
	f.inputActive = false
	return true
}

// Increment increases current field value
func (f *FilterView) Increment() {
	switch FilterField(f.selectedRow) {
	case FilterFieldMinRSI:
		if f.criteria.MinRSI < f.criteria.MaxRSI {
			f.criteria.MinRSI += 5
		}
	case FilterFieldMaxRSI:
		if f.criteria.MaxRSI < 100 {
			f.criteria.MaxRSI += 5
		}
	case FilterFieldMaxPBV:
		f.criteria.MaxPBV += 0.5
	case FilterFieldMinScore:
		if f.criteria.MinConfluence < 100 {
			f.criteria.MinConfluence += 5
		}
	}
}

// Decrement decreases current field value
func (f *FilterView) Decrement() {
	switch FilterField(f.selectedRow) {
	case FilterFieldMinRSI:
		if f.criteria.MinRSI > 0 {
			f.criteria.MinRSI -= 5
			if f.criteria.MinRSI < 0 {
				f.criteria.MinRSI = 0
			}
		}
	case FilterFieldMaxRSI:
		if f.criteria.MaxRSI > f.criteria.MinRSI {
			f.criteria.MaxRSI -= 5
		}
	case FilterFieldMaxPBV:
		if f.criteria.MaxPBV > 0.5 {
			f.criteria.MaxPBV -= 0.5
		}
	case FilterFieldMinScore:
		if f.criteria.MinConfluence > 0 {
			f.criteria.MinConfluence -= 5
			if f.criteria.MinConfluence < 0 {
				f.criteria.MinConfluence = 0
			}
		}
	}
}

// Reset restores default criteria
func (f *FilterView) Reset() {
	f.criteria = screener.DefaultCriteria()
}

// View renders the filter view
func (f *FilterView) View() string {
	var b strings.Builder

	// Calculate border width based on screen width
	borderWidth := f.width - 10
	if borderWidth > 60 {
		borderWidth = 60
	}
	if borderWidth < 30 {
		borderWidth = 30
	}

	// Center content
	contentHeight := 20
	topPadding := (f.height - contentHeight) / 2
	if topPadding < 1 {
		topPadding = 1
	}

	for i := 0; i < topPadding; i++ {
		b.WriteString("\n")
	}

	// Title
	title := "FILTER CRITERIA"
	titleStyled := styles.TitleStyle.Render(title)
	b.WriteString(centerText(titleStyled, f.width))
	b.WriteString("\n\n")

	// Border
	border := strings.Repeat("─", borderWidth)
	b.WriteString(centerText(styles.MutedStyle().Render(border), f.width))
	b.WriteString("\n\n")

	// Description
	desc := "Adjust filter criteria for stock screening"
	b.WriteString(centerText(styles.MutedStyle().Render(desc), f.width))
	b.WriteString("\n\n")

	// Filter fields
	fields := []struct {
		name        string
		value       string
		description string
	}{
		{
			name:        "Min RSI",
			value:       fmt.Sprintf("%.0f", f.criteria.MinRSI),
			description: "Minimum RSI threshold (0-100)",
		},
		{
			name:        "Max RSI",
			value:       fmt.Sprintf("%.0f", f.criteria.MaxRSI),
			description: "Maximum RSI threshold (oversold < 30)",
		},
		{
			name:        "Max PBV",
			value:       fmt.Sprintf("%.1f", f.criteria.MaxPBV),
			description: "Maximum Price/Book ratio (< 1 = undervalued)",
		},
		{
			name:        "Min Score",
			value:       fmt.Sprintf("%.0f", f.criteria.MinConfluence),
			description: "Minimum confluence score (0-100)",
		},
	}

	for i, field := range fields {
		selected := i == f.selectedRow

		// Build row
		var row string
		indicator := "  "
		if selected {
			indicator = "> "
		}

		nameStyle := styles.TextStyle
		valueStyle := styles.InfoStyle
		if selected {
			nameStyle = nameStyle.Bold(true)
			valueStyle = lipgloss.NewStyle().Foreground(styles.ColorHighlight).Bold(true)
		}

		// If editing this field
		displayValue := field.value
		if f.inputActive && i == f.selectedRow {
			displayValue = f.inputBuffer + "_"
			valueStyle = lipgloss.NewStyle().Foreground(styles.ColorSuccess).Bold(true)
		}

		row = fmt.Sprintf("%s%-12s %s",
			indicator,
			nameStyle.Render(field.name+":"),
			valueStyle.Render(fmt.Sprintf("[%6s]", displayValue)),
		)

		if selected && !f.inputActive {
			row += styles.MutedStyle().Render("  [-/+]")
		}

		b.WriteString(centerText(row, f.width))
		b.WriteString("\n")

		// Show description for selected
		if selected {
			descText := styles.MutedStyle().Render("    " + field.description)
			b.WriteString(centerText(descText, f.width))
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")

	// RSI Bar visualization
	rsiBar := f.renderRSIBar()
	b.WriteString(centerText(rsiBar, f.width))
	b.WriteString("\n\n")

	// Border
	b.WriteString(centerText(styles.MutedStyle().Render(border), f.width))
	b.WriteString("\n\n")

	// Help
	var help string
	if f.inputActive {
		help = "[Enter] Apply  [Esc] Cancel"
	} else {
		help = "[Up/Down] Select  [-/+] Adjust  [Enter] Edit  [R] Reset  [Esc] Back & Apply"
	}
	b.WriteString(centerText(styles.HelpStyle.Render(help), f.width))

	return b.String()
}

// renderRSIBar renders a visual RSI range indicator
func (f *FilterView) renderRSIBar() string {
	barWidth := 40
	var bar strings.Builder

	bar.WriteString("RSI Range: ")

	// Calculate positions
	minPos := int(f.criteria.MinRSI / 100 * float64(barWidth))
	maxPos := int(f.criteria.MaxRSI / 100 * float64(barWidth))

	for i := 0; i < barWidth; i++ {
		var char string
		var style lipgloss.Style

		if i < minPos {
			char = "░"
			style = styles.MutedStyle()
		} else if i <= maxPos {
			char = "█"
			if i < barWidth/3 {
				style = lipgloss.NewStyle().Foreground(styles.ColorDanger) // Oversold zone
			} else if i < barWidth*2/3 {
				style = lipgloss.NewStyle().Foreground(styles.ColorWarning) // Neutral zone
			} else {
				style = lipgloss.NewStyle().Foreground(styles.ColorSuccess) // Overbought zone
			}
		} else {
			char = "░"
			style = styles.MutedStyle()
		}

		bar.WriteString(style.Render(char))
	}

	bar.WriteString(fmt.Sprintf(" [%.0f-%.0f]", f.criteria.MinRSI, f.criteria.MaxRSI))

	return bar.String()
}
