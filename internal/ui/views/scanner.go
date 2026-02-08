package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/febritecno/stockmap-cli/internal/styles"
)

// Spinner frames for loading animation
var spinnerFrames = []string{
	"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏",
}

// Progress bar styles
var progressChars = struct {
	filled string
	empty  string
	head   string
}{
	filled: "█",
	empty:  "░",
	head:   "▓",
}

// Scanner shows the scanning progress with animations
type Scanner struct {
	width        int
	height       int
	total        int
	completed    int
	current      string
	isComplete   bool
	foundCount   int
	frame        int
	successCount int
	errorCount   int
	lastError    string
	errorSymbol  string
	recentErrors []string // Keep last few errors for display
}

// NewScanner creates a new scanner view
func NewScanner() *Scanner {
	return &Scanner{}
}

// SetSize sets the view dimensions
func (s *Scanner) SetSize(width, height int) {
	s.width = width
	s.height = height
}

// SetProgress updates the scan progress
func (s *Scanner) SetProgress(completed, total int, current string) {
	s.completed = completed
	s.total = total
	s.current = current
	s.isComplete = completed >= total && total > 0
	s.frame++
}

// NextFrame advances the animation frame
func (s *Scanner) NextFrame() {
	s.frame++
}

// GetCompleted returns completed count
func (s *Scanner) GetCompleted() int {
	return s.completed
}

// GetTotal returns total count
func (s *Scanner) GetTotal() int {
	return s.total
}

// GetCurrent returns current symbol
func (s *Scanner) GetCurrent() string {
	return s.current
}

// SetVerboseProgress updates progress with verbose error information
func (s *Scanner) SetVerboseProgress(completed, total int, current string, successCount, errorCount int, lastError, errorSymbol string) {
	s.completed = completed
	s.total = total
	s.current = current
	s.isComplete = completed >= total && total > 0
	s.successCount = successCount
	s.errorCount = errorCount
	s.frame++

	// Track recent errors (keep last 5)
	if lastError != "" && errorSymbol != "" {
		errMsg := errorSymbol + ": " + lastError
		// Only add if it's a new error (not already the last one)
		if len(s.recentErrors) == 0 || s.recentErrors[len(s.recentErrors)-1] != errMsg {
			s.recentErrors = append(s.recentErrors, errMsg)
			if len(s.recentErrors) > 5 {
				s.recentErrors = s.recentErrors[1:]
			}
		}
	}
	s.lastError = lastError
	s.errorSymbol = errorSymbol
}

// SetFoundCount sets the number of stocks found
func (s *Scanner) SetFoundCount(count int) {
	s.foundCount = count
}

// View renders the scanner view
func (s *Scanner) View() string {
	var b strings.Builder

	// Center content vertically
	contentHeight := 18
	topPadding := (s.height - contentHeight) / 2
	if topPadding < 0 {
		topPadding = 1
	}

	for i := 0; i < topPadding; i++ {
		b.WriteString("\n")
	}

	if s.isComplete {
		b.WriteString(s.renderComplete())
	} else {
		b.WriteString(s.renderScanning())
	}

	return b.String()
}

// renderScanning renders the scanning animation
func (s *Scanner) renderScanning() string {
	var b strings.Builder

	// Animated header
	spinner := spinnerFrames[s.frame%len(spinnerFrames)]

	header := "SCANNING STOCKS"
	headerStyled := styles.TitleStyle.Render(header)
	b.WriteString(centerText(headerStyled, s.width))
	b.WriteString("\n\n")

	// Animated border
	borderChar := "═"
	if s.frame%2 == 0 {
		borderChar = "─"
	}
	border := strings.Repeat(borderChar, 50)
	b.WriteString(centerText(styles.MutedStyle().Render(border), s.width))
	b.WriteString("\n\n")

	// Progress bar with animation
	progressWidth := 40
	var progress float64
	if s.total > 0 {
		progress = float64(s.completed) / float64(s.total)
	}

	filled := min(int(progress*float64(progressWidth)), progressWidth)

	// Animated progress bar
	bar := ""
	for i := 0; i < progressWidth; i++ {
		if i < filled {
			bar += progressChars.filled
		} else if i == filled {
			// Animated head
			if s.frame%2 == 0 {
				bar += progressChars.head
			} else {
				bar += progressChars.filled
			}
		} else {
			bar += progressChars.empty
		}
	}

	progressBar := styles.ProgressStyle.Render(bar[:filled])
	if filled < progressWidth {
		remaining := lipgloss.NewStyle().Foreground(styles.ColorMuted).Render(bar[filled:])
		progressBar += remaining
	}

	// Add percentage
	percent := fmt.Sprintf(" %.0f%%", progress*100)
	progressBar += styles.InfoStyle.Render(percent)

	b.WriteString(centerText(progressBar, s.width))
	b.WriteString("\n\n")

	// Stats with animation - now with success/error breakdown
	statsLine := fmt.Sprintf("%s  %d / %d stocks scanned", spinner, s.completed, s.total)
	b.WriteString(centerText(styles.InfoStyle.Render(statsLine), s.width))
	b.WriteString("\n")

	// Verbose success/error counts
	successStyle := lipgloss.NewStyle().Foreground(styles.ColorSuccess)
	errorStyle := lipgloss.NewStyle().Foreground(styles.ColorDanger)
	verboseStats := fmt.Sprintf("   %s %d success   %s %d errors",
		successStyle.Render("✓"),
		s.successCount,
		errorStyle.Render("✗"),
		s.errorCount)
	b.WriteString(centerText(verboseStats, s.width))
	b.WriteString("\n\n")

	// Current symbol with animation
	if s.current != "" {
		// Animated dots
		dots := strings.Repeat(".", (s.frame%3)+1)
		for len(dots) < 3 {
			dots += " "
		}

		currentLine := fmt.Sprintf("Processing: %s%s", styles.TickerStyle.Render(s.current), dots)
		b.WriteString(centerText(currentLine, s.width))
		b.WriteString("\n\n")
	}

	// Show recent errors if any (verbose output)
	if len(s.recentErrors) > 0 {
		errorHeader := errorStyle.Render("Recent Errors:")
		b.WriteString(centerText(errorHeader, s.width))
		b.WriteString("\n")

		// Show last few errors with truncation
		for _, errMsg := range s.recentErrors {
			// Truncate long error messages
			displayErr := errMsg
			if len(displayErr) > 60 {
				displayErr = displayErr[:57] + "..."
			}
			errLine := styles.MutedStyle().Render("  " + displayErr)
			b.WriteString(centerText(errLine, s.width))
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	// Animated loading indicator
	loadingAnim := s.renderLoadingAnimation()
	b.WriteString(centerText(loadingAnim, s.width))
	b.WriteString("\n\n")

	// Bottom border
	b.WriteString(centerText(styles.MutedStyle().Render(border), s.width))
	b.WriteString("\n\n")

	// Cancel hint
	hint := styles.HelpStyle.Render("Press [ESC] to cancel")
	b.WriteString(centerText(hint, s.width))

	return b.String()
}

// renderLoadingAnimation renders an animated loading indicator
func (s *Scanner) renderLoadingAnimation() string {
	// Wave animation
	wave := []string{"▁", "▂", "▃", "▄", "▅", "▆", "▇", "█", "▇", "▆", "▅", "▄", "▃", "▂"}
	waveLen := len(wave)

	result := ""
	for i := 0; i < 12; i++ {
		idx := (s.frame + i) % waveLen
		char := wave[idx]

		// Color gradient
		var style lipgloss.Style
		switch i % 4 {
		case 0:
			style = lipgloss.NewStyle().Foreground(styles.ColorPrimary)
		case 1:
			style = lipgloss.NewStyle().Foreground(styles.ColorCyan)
		case 2:
			style = lipgloss.NewStyle().Foreground(styles.ColorHighlight)
		case 3:
			style = lipgloss.NewStyle().Foreground(styles.ColorSuccess)
		}
		result += style.Render(char)
	}

	return result
}

// renderComplete renders the completion screen
func (s *Scanner) renderComplete() string {
	var b strings.Builder

	// Success icon
	icon := "✓"
	iconStyled := styles.ScoreHighStyle.Bold(true).Render(icon)
	b.WriteString(centerText(iconStyled, s.width))
	b.WriteString("\n\n")

	// Title
	title := "SCAN COMPLETE!"
	titleStyled := styles.ScoreHighStyle.Bold(true).Render(title)
	b.WriteString(centerText(titleStyled, s.width))
	b.WriteString("\n\n")

	// Border
	border := strings.Repeat("━", 40)
	b.WriteString(centerText(styles.MutedStyle().Render(border), s.width))
	b.WriteString("\n\n")

	// Stats
	scannedLine := fmt.Sprintf("Stocks Scanned:  %s", styles.InfoStyle.Render(fmt.Sprintf("%d", s.total)))
	b.WriteString(centerText(scannedLine, s.width))
	b.WriteString("\n")

	foundLine := fmt.Sprintf("Matches Found:   %s", styles.ScoreHighStyle.Render(fmt.Sprintf("%d", s.foundCount)))
	b.WriteString(centerText(foundLine, s.width))
	b.WriteString("\n\n")

	// Result indicator
	var resultMsg string
	if s.foundCount > 10 {
		resultMsg = "🎯 Excellent! Many opportunities found!"
	} else if s.foundCount > 5 {
		resultMsg = "👍 Good! Several stocks match your criteria"
	} else if s.foundCount > 0 {
		resultMsg = "📊 A few stocks match your criteria"
	} else {
		resultMsg = "📉 No stocks match current criteria"
	}
	b.WriteString(centerText(styles.InfoStyle.Render(resultMsg), s.width))
	b.WriteString("\n\n")

	// Border
	b.WriteString(centerText(styles.MutedStyle().Render(border), s.width))
	b.WriteString("\n\n")

	// Continue hint
	hint := styles.HelpStyle.Render("Press any key to view results...")
	b.WriteString(centerText(hint, s.width))

	return b.String()
}

// IsComplete returns whether scanning is complete
func (s *Scanner) IsComplete() bool {
	return s.isComplete
}
