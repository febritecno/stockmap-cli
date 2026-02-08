package views

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/febritecno/stockmap-cli/internal/styles"
)

// HelpView displays tutorial, legends, and keybindings information
type HelpView struct {
	width  int
	height int
}

// NewHelpView creates a new help view
func NewHelpView() *HelpView {
	return &HelpView{}
}

// SetSize sets the view dimensions
func (h *HelpView) SetSize(width, height int) {
	h.width = width
	h.height = height
}

// View renders the help view
func (h *HelpView) View() string {
	var b strings.Builder

	// Styles
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(styles.ColorPrimary).
		MarginBottom(1)

	sectionStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(styles.ColorCyan).
		MarginTop(1)

	keyStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(styles.ColorWarning)

	descStyle := lipgloss.NewStyle().
		Foreground(styles.ColorText)

	legendLabelStyle := lipgloss.NewStyle().
		Width(12).
		Foreground(styles.ColorMuted)

	// Title
	b.WriteString(titleStyle.Render("STOCKMAP - Help & Legends"))
	b.WriteString("\n\n")

	// Keyboard Shortcuts Section
	b.WriteString(sectionStyle.Render("KEYBOARD SHORTCUTS"))
	b.WriteString("\n\n")

	shortcuts := []struct {
		key  string
		desc string
	}{
		{"S", "Open scan mode selection"},
		{"R", "Reload/Refresh data (toggle)"},
		{"T", "Toggle auto-reload (60s interval)"},
		{"F", "Open filter criteria editor"},
		{"W", "View watchlist"},
		{"H", "View scan history"},
		{"P", "View price alerts"},
		{"D", "View stock details"},
		{"A", "Add selected to watchlist"},
		{"X", "Clear all results"},
		{"C", "Check connection status"},
		{"/", "Quick search by symbol"},
		{"Tab", "Cycle sort column"},
		{"I", "Show this help"},
		{"Q", "Quit / Go back"},
	}

	for _, s := range shortcuts {
		b.WriteString(keyStyle.Render("["+s.key+"]") + " " + descStyle.Render(s.desc) + "\n")
	}

	b.WriteString("\n")

	// Score Legend Section
	b.WriteString(sectionStyle.Render("CONFLUENCE SCORE"))
	b.WriteString("\n\n")

	scoreColors := []struct {
		range_ string
		color  lipgloss.Color
		desc   string
	}{
		{"75-100", styles.ColorSuccess, "Strong Buy Signal"},
		{"50-74", styles.ColorWarning, "Moderate Signal"},
		{"0-49", styles.ColorDanger, "Weak Signal"},
	}

	for _, sc := range scoreColors {
		colorBox := lipgloss.NewStyle().
			Foreground(sc.color).
			Render("███")
		b.WriteString(legendLabelStyle.Render(sc.range_) + colorBox + " " + descStyle.Render(sc.desc) + "\n")
	}

	b.WriteString("\n")

	// RSI Legend
	b.WriteString(sectionStyle.Render("RSI INDICATOR"))
	b.WriteString("\n\n")

	rsiLevels := []struct {
		range_ string
		color  lipgloss.Color
		desc   string
	}{
		{"< 30", styles.ColorSuccess, "Oversold (Buy opportunity)"},
		{"30-70", styles.ColorText, "Neutral"},
		{"> 70", styles.ColorDanger, "Overbought (Sell signal)"},
	}

	for _, r := range rsiLevels {
		colorBox := lipgloss.NewStyle().
			Foreground(r.color).
			Render("███")
		b.WriteString(legendLabelStyle.Render(r.range_) + colorBox + " " + descStyle.Render(r.desc) + "\n")
	}

	b.WriteString("\n")

	// PBV Legend
	b.WriteString(sectionStyle.Render("P/B VALUE (PBV)"))
	b.WriteString("\n\n")

	pbvLevels := []struct {
		range_ string
		color  lipgloss.Color
		desc   string
	}{
		{"< 1.0", styles.ColorSuccess, "Undervalued (Below book value)"},
		{"1.0-2.0", styles.ColorText, "Fair value"},
		{"> 2.0", styles.ColorDanger, "Overvalued"},
	}

	for _, p := range pbvLevels {
		colorBox := lipgloss.NewStyle().
			Foreground(p.color).
			Render("███")
		b.WriteString(legendLabelStyle.Render(p.range_) + colorBox + " " + descStyle.Render(p.desc) + "\n")
	}

	b.WriteString("\n")

	// Volatility Legend
	b.WriteString(sectionStyle.Render("VOLATILITY"))
	b.WriteString("\n\n")

	volLevels := []struct {
		range_ string
		color  lipgloss.Color
		desc   string
	}{
		{"< 20%", styles.ColorSuccess, "Low risk"},
		{"20-40%", styles.ColorText, "Moderate risk"},
		{"> 40%", styles.ColorDanger, "High risk"},
	}

	for _, v := range volLevels {
		colorBox := lipgloss.NewStyle().
			Foreground(v.color).
			Render("███")
		b.WriteString(legendLabelStyle.Render(v.range_) + colorBox + " " + descStyle.Render(v.desc) + "\n")
	}

	b.WriteString("\n")

	// Symbols Legend
	b.WriteString(sectionStyle.Render("SYMBOLS"))
	b.WriteString("\n\n")

	symbols := []struct {
		symbol string
		desc   string
	}{
		{"*", "Pinned/Watchlist stock"},
		{"TP", "Take Profit target price"},
		{"SL", "Stop Loss price"},
	}

	for _, s := range symbols {
		symbolStyled := lipgloss.NewStyle().
			Bold(true).
			Foreground(styles.ColorWarning).
			Width(4).
			Render(s.symbol)
		b.WriteString(symbolStyled + descStyle.Render(s.desc) + "\n")
	}

	b.WriteString("\n")

	// Footer
	footerStyle := lipgloss.NewStyle().
		Foreground(styles.ColorMuted).
		MarginTop(1)

	b.WriteString(footerStyle.Render("Press [Esc] or [I] to close"))

	// Wrap in box
	content := b.String()

	boxWidth := h.width - 4
	if boxWidth < 50 {
		boxWidth = 50
	}
	if boxWidth > 70 {
		boxWidth = 70
	}

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.ColorPrimary).
		Padding(1, 2).
		Width(boxWidth)

	return lipgloss.Place(
		h.width,
		h.height,
		lipgloss.Center,
		lipgloss.Center,
		boxStyle.Render(content),
	)
}
