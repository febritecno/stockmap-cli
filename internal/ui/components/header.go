package components

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"

	"stockmap/internal/styles"
)

// Header represents the app header component
type Header struct {
	width       int
	title       string
	version     string
	marketState string
	strategy    string
}

// NewHeader creates a new header component
func NewHeader() *Header {
	return &Header{
		title:       "STOCKMAP",
		version:     "v1.0",
		marketState: "CLOSED",
		strategy:    "Deep Value",
	}
}

// SetWidth sets the header width
func (h *Header) SetWidth(w int) {
	h.width = w
}

// SetMarketState updates the market state
func (h *Header) SetMarketState(state string) {
	h.marketState = state
}

// View renders the header
func (h *Header) View() string {
	// Logo
	logo := styles.LogoStyle.Render("◉ " + h.title)

	// Version
	version := styles.MutedStyle().Render(h.version)

	// Market status
	var marketStatus string
	switch h.marketState {
	case "REGULAR", "PRE", "POST":
		marketStatus = styles.MarketOpenStyle.Render("Market: OPEN")
	default:
		marketStatus = styles.MarketClosedStyle.Render("Market: CLOSED")
	}

	// Strategy
	strategy := styles.InfoStyle.Render("Strategy: " + h.strategy)

	// For narrow screens, use compact layout
	if h.width < 70 {
		// Compact single-line layout
		compact := lipgloss.JoinHorizontal(lipgloss.Center, logo, "  ", version, "  ", marketStatus)
		return styles.HeaderStyle.Width(h.width).Render(compact)
	}

	// Left side
	left := lipgloss.JoinHorizontal(lipgloss.Center, logo, "  ", version)

	// Right side
	right := lipgloss.JoinHorizontal(lipgloss.Center, marketStatus, "    ", strategy)

	// Calculate spacing
	leftWidth := lipgloss.Width(left)
	rightWidth := lipgloss.Width(right)
	spacing := h.width - leftWidth - rightWidth - 4

	if spacing < 0 {
		spacing = 2
	}

	spacer := lipgloss.NewStyle().Width(spacing).Render("")

	header := lipgloss.JoinHorizontal(lipgloss.Center, left, spacer, right)

	return styles.HeaderStyle.Width(h.width).Render(header)
}

// RenderDivider returns a horizontal divider
func RenderDivider(width int) string {
	divider := ""
	for i := 0; i < width; i++ {
		divider += "─"
	}
	return lipgloss.NewStyle().Foreground(styles.ColorMuted).Render(divider)
}

// RenderTitle renders a section title
func RenderTitle(title string, width int) string {
	left := "── "
	right := " ──"
	titleStyled := styles.TitleStyle.Render(title)

	remaining := width - lipgloss.Width(left) - lipgloss.Width(titleStyled) - lipgloss.Width(right)
	if remaining > 0 {
		for i := 0; i < remaining; i++ {
			right += "─"
		}
	}

	return lipgloss.NewStyle().Foreground(styles.ColorMuted).Render(left) +
		titleStyled +
		lipgloss.NewStyle().Foreground(styles.ColorMuted).Render(right)
}

// FormatPrice formats a price with dollar sign
func FormatPrice(price float64) string {
	return fmt.Sprintf("$%.2f", price)
}

// FormatLargeNumber formats large numbers with K/M/B suffix
func FormatLargeNumber(n int64) string {
	if n >= 1_000_000_000 {
		return fmt.Sprintf("%.1fB", float64(n)/1_000_000_000)
	}
	if n >= 1_000_000 {
		return fmt.Sprintf("%.1fM", float64(n)/1_000_000)
	}
	if n >= 1_000 {
		return fmt.Sprintf("%.1fK", float64(n)/1_000)
	}
	return fmt.Sprintf("%d", n)
}
