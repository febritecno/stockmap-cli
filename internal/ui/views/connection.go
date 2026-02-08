package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/febritecno/stockmap-cli/internal/fetcher"
	"github.com/febritecno/stockmap-cli/internal/styles"
)

// ConnectionView shows connection test results
type ConnectionView struct {
	width   int
	height  int
	result  *fetcher.ConnectionResult
	testing bool
	frame   int
}

// NewConnectionView creates a new connection view
func NewConnectionView() *ConnectionView {
	return &ConnectionView{}
}

// SetSize sets the view dimensions
func (c *ConnectionView) SetSize(width, height int) {
	c.width = width
	c.height = height
}

// SetTesting marks the view as testing
func (c *ConnectionView) SetTesting(testing bool) {
	c.testing = testing
	c.frame++
}

// SetResult sets the connection test result
func (c *ConnectionView) SetResult(result *fetcher.ConnectionResult) {
	c.result = result
	c.testing = false
}

// View renders the connection view
func (c *ConnectionView) View() string {
	var b strings.Builder

	// Center content vertically
	contentHeight := 20
	topPadding := (c.height - contentHeight) / 2
	if topPadding < 0 {
		topPadding = 1
	}

	for i := 0; i < topPadding; i++ {
		b.WriteString("\n")
	}

	// Header
	header := "CONNECTION TEST"
	headerStyled := styles.TitleStyle.Render(header)
	b.WriteString(centerText(headerStyled, c.width))
	b.WriteString("\n\n")

	// Border
	border := strings.Repeat("═", 50)
	b.WriteString(centerText(styles.MutedStyle().Render(border), c.width))
	b.WriteString("\n\n")

	if c.testing {
		// Show testing animation
		spinner := spinnerFrames[c.frame%len(spinnerFrames)]
		testingLine := fmt.Sprintf("%s Testing connection to Yahoo Finance...", spinner)
		b.WriteString(centerText(styles.InfoStyle.Render(testingLine), c.width))
		b.WriteString("\n\n")

		// Loading animation
		wave := []string{"▁", "▂", "▃", "▄", "▅", "▆", "▇", "█", "▇", "▆", "▅", "▄", "▃", "▂"}
		waveLen := len(wave)
		result := ""
		for i := 0; i < 12; i++ {
			idx := (c.frame + i) % waveLen
			char := wave[idx]
			style := lipgloss.NewStyle().Foreground(styles.ColorPrimary)
			result += style.Render(char)
		}
		b.WriteString(centerText(result, c.width))
		b.WriteString("\n\n")
	} else if c.result != nil {
		// Show results
		successStyle := lipgloss.NewStyle().Foreground(styles.ColorSuccess).Bold(true)
		errorStyle := lipgloss.NewStyle().Foreground(styles.ColorDanger).Bold(true)
		infoStyle := lipgloss.NewStyle().Foreground(styles.ColorCyan)

		// Connection status
		if c.result.Connected {
			status := successStyle.Render("CONNECTED")
			b.WriteString(centerText(status, c.width))
		} else {
			status := errorStyle.Render("FAILED")
			b.WriteString(centerText(status, c.width))
		}
		b.WriteString("\n\n")

		// Latency
		latencyLine := fmt.Sprintf("Latency: %v", c.result.Latency)
		b.WriteString(centerText(infoStyle.Render(latencyLine), c.width))
		b.WriteString("\n\n")

		// API Status
		quoteStatus := renderStatus("Quote API", c.result.QuoteWorks)
		b.WriteString(centerText(quoteStatus, c.width))
		b.WriteString("\n")

		equityStatus := renderStatus("Equity API", c.result.EquityWorks)
		b.WriteString(centerText(equityStatus, c.width))
		b.WriteString("\n")

		chartStatus := renderStatus("Chart API", c.result.ChartWorks)
		b.WriteString(centerText(chartStatus, c.width))
		b.WriteString("\n\n")

		// Details
		if len(c.result.Details) > 0 {
			detailHeader := styles.MutedStyle().Render("Details:")
			b.WriteString(centerText(detailHeader, c.width))
			b.WriteString("\n")

			for _, detail := range c.result.Details {
				// Truncate long lines
				displayDetail := detail
				if len(displayDetail) > 60 {
					displayDetail = displayDetail[:57] + "..."
				}

				// Color based on content
				var detailStyle lipgloss.Style
				if strings.Contains(detail, "FAIL") {
					detailStyle = errorStyle
				} else if strings.Contains(detail, "OK") {
					detailStyle = successStyle
				} else {
					detailStyle = styles.MutedStyle()
				}

				detailLine := detailStyle.Render("  " + displayDetail)
				b.WriteString(centerText(detailLine, c.width))
				b.WriteString("\n")
			}
			b.WriteString("\n")
		}

		// Error message
		if c.result.Error != "" {
			errLine := errorStyle.Render("Error: " + c.result.Error)
			b.WriteString(centerText(errLine, c.width))
			b.WriteString("\n\n")
		}
	} else {
		// No result yet
		noResult := styles.MutedStyle().Render("Press [C] to test connection")
		b.WriteString(centerText(noResult, c.width))
		b.WriteString("\n\n")
	}

	// Bottom border
	b.WriteString(centerText(styles.MutedStyle().Render(border), c.width))
	b.WriteString("\n\n")

	// Hint
	hint := styles.HelpStyle.Render("Press [ESC] to go back, [C] to test again")
	b.WriteString(centerText(hint, c.width))

	return b.String()
}

func renderStatus(name string, ok bool) string {
	successStyle := lipgloss.NewStyle().Foreground(styles.ColorSuccess)
	errorStyle := lipgloss.NewStyle().Foreground(styles.ColorDanger)

	if ok {
		return fmt.Sprintf("%s: %s", name, successStyle.Render("OK"))
	}
	return fmt.Sprintf("%s: %s", name, errorStyle.Render("FAIL"))
}
