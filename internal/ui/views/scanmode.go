package views

import (
	"fmt"
	"strings"

	"github.com/febritecno/stockmap/internal/styles"
)

// ScanMode represents the scan mode
type ScanMode int

const (
	ScanModeAll ScanMode = iota
	ScanModeWatchlist
	ScanModeCustom
)

// ScanModeView shows scan mode selection and custom input
type ScanModeView struct {
	width          int
	height         int
	selectedMode   ScanMode
	customInput    string
	inputActive    bool
	watchlistCount int
}

// NewScanModeView creates a new scan mode view
func NewScanModeView() *ScanModeView {
	return &ScanModeView{
		selectedMode: ScanModeAll,
	}
}

// SetSize sets the view dimensions
func (s *ScanModeView) SetSize(width, height int) {
	s.width = width
	s.height = height
}

// SetWatchlistCount sets the watchlist count for display
func (s *ScanModeView) SetWatchlistCount(count int) {
	s.watchlistCount = count
}

// GetSelectedMode returns the selected scan mode
func (s *ScanModeView) GetSelectedMode() ScanMode {
	return s.selectedMode
}

// GetCustomSymbols returns parsed custom symbols
func (s *ScanModeView) GetCustomSymbols() []string {
	if s.customInput == "" {
		return nil
	}

	// Parse comma or space separated symbols
	input := strings.ToUpper(s.customInput)
	input = strings.ReplaceAll(input, ",", " ")
	parts := strings.Fields(input)

	symbols := make([]string, 0)
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			symbols = append(symbols, p)
		}
	}
	return symbols
}

// IsInputActive returns if custom input is active
func (s *ScanModeView) IsInputActive() bool {
	return s.inputActive
}

// MoveUp moves selection up
func (s *ScanModeView) MoveUp() {
	if !s.inputActive {
		if s.selectedMode > ScanModeAll {
			s.selectedMode--
		}
	}
}

// MoveDown moves selection down
func (s *ScanModeView) MoveDown() {
	if !s.inputActive {
		if s.selectedMode < ScanModeCustom {
			s.selectedMode++
		}
	}
}

// ToggleInput toggles custom input mode
func (s *ScanModeView) ToggleInput() {
	if s.selectedMode == ScanModeCustom {
		s.inputActive = !s.inputActive
	}
}

// AddChar adds a character to custom input
func (s *ScanModeView) AddChar(ch rune) {
	if s.inputActive {
		s.customInput += string(ch)
	}
}

// Backspace removes last character
func (s *ScanModeView) Backspace() {
	if s.inputActive && len(s.customInput) > 0 {
		s.customInput = s.customInput[:len(s.customInput)-1]
	}
}

// ClearInput clears the custom input
func (s *ScanModeView) ClearInput() {
	s.customInput = ""
}

// Reset resets the view state
func (s *ScanModeView) Reset() {
	s.selectedMode = ScanModeAll
	s.customInput = ""
	s.inputActive = false
}

// View renders the scan mode view - simplified version
func (s *ScanModeView) View() string {
	var b strings.Builder

	// Simple top padding
	b.WriteString("\n\n")

	// Header
	b.WriteString("  SCAN MODE\n")
	b.WriteString("  =========\n\n")

	// Options - simple text menu
	options := []struct {
		key  string
		name string
		desc string
	}{
		{"1", "Scan All", fmt.Sprintf("(%d stocks)", 148)},
		{"2", "Scan Watchlist", fmt.Sprintf("(%d pinned)", s.watchlistCount)},
		{"3", "Custom Symbols", "(type symbols)"},
	}

	for i, opt := range options {
		mode := ScanMode(i)
		prefix := "  "
		if s.selectedMode == mode {
			prefix = "> "
		}
		b.WriteString(fmt.Sprintf("%s[%s] %s %s\n", prefix, opt.key, opt.name, styles.MutedStyle().Render(opt.desc)))
	}

	b.WriteString("\n")

	// Show input for custom mode
	if s.selectedMode == ScanModeCustom {
		b.WriteString("  Symbols: ")
		if s.inputActive {
			b.WriteString(styles.InfoStyle.Render(s.customInput))
			b.WriteString("_") // cursor
		} else {
			if s.customInput == "" {
				b.WriteString(styles.MutedStyle().Render("(press Tab to type)"))
			} else {
				b.WriteString(s.customInput)
			}
		}
		b.WriteString("\n")

		// Show parsed
		if symbols := s.GetCustomSymbols(); len(symbols) > 0 {
			b.WriteString(fmt.Sprintf("  Will scan: %s\n", strings.Join(symbols, ", ")))
		}
		b.WriteString("\n")
	}

	// Simple help
	b.WriteString("  ---\n")
	if s.inputActive {
		b.WriteString("  Type symbols, Enter=done, Esc=cancel, Backspace=delete\n")
	} else {
		b.WriteString("  1/2/3=select, Enter=start, Tab=edit, Esc=back\n")
	}

	return b.String()
}
