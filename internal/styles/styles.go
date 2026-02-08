package styles

import (
	"github.com/charmbracelet/lipgloss"
)

// Color scheme based on Tokyo Night theme
var (
	ColorBackground = lipgloss.Color("#1a1b26")
	ColorPrimary    = lipgloss.Color("#7aa2f7")
	ColorSuccess    = lipgloss.Color("#9ece6a")
	ColorWarning    = lipgloss.Color("#e0af68")
	ColorDanger     = lipgloss.Color("#f7768e")
	ColorText       = lipgloss.Color("#c0caf5")
	ColorMuted      = lipgloss.Color("#565f89")
	ColorHighlight  = lipgloss.Color("#bb9af7")
	ColorCyan       = lipgloss.Color("#7dcfff")
	ColorOrange     = lipgloss.Color("#ff9e64")
)

// Styles for various UI elements
var (
	// App frame
	AppStyle = lipgloss.NewStyle().
			Background(ColorBackground)

	// Header styles
	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary).
			Background(ColorBackground).
			Padding(0, 1)

	LogoStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorHighlight)

	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorText)

	TextStyle = lipgloss.NewStyle().
			Foreground(ColorText)

	// Status indicators
	MarketOpenStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorSuccess)

	MarketClosedStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorDanger)

	// Table styles
	TableHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorPrimary).
				BorderBottom(true).
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(ColorMuted)

	TableRowStyle = lipgloss.NewStyle().
			Foreground(ColorText)

	TableRowSelectedStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorBackground).
				Background(ColorPrimary)

	TableRowPinnedStyle = lipgloss.NewStyle().
				Foreground(ColorWarning)

	// Cell styles
	TickerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorCyan)

	PriceStyle = lipgloss.NewStyle().
			Foreground(ColorText)

	PriceUpStyle = lipgloss.NewStyle().
			Foreground(ColorSuccess)

	PriceDownStyle = lipgloss.NewStyle().
			Foreground(ColorDanger)

	// Score styles
	ScoreHighStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorSuccess)

	ScoreMediumStyle = lipgloss.NewStyle().
				Foreground(ColorWarning)

	ScoreLowStyle = lipgloss.NewStyle().
			Foreground(ColorDanger)

	// RSI styles
	RSIOversoldStyle = lipgloss.NewStyle().
				Foreground(ColorSuccess)

	RSINeutralStyle = lipgloss.NewStyle().
			Foreground(ColorText)

	RSIOverboughtStyle = lipgloss.NewStyle().
				Foreground(ColorDanger)

	// PBV styles
	PBVLowStyle = lipgloss.NewStyle().
			Foreground(ColorSuccess)

	PBVNeutralStyle = lipgloss.NewStyle().
			Foreground(ColorText)

	PBVHighStyle = lipgloss.NewStyle().
			Foreground(ColorDanger)

	// Status bar styles
	StatusBarStyle = lipgloss.NewStyle().
			Foreground(ColorMuted).
			Background(ColorBackground).
			Padding(0, 1)

	StatusItemStyle = lipgloss.NewStyle().
			Foreground(ColorText)

	KeyStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary)

	HelpStyle = lipgloss.NewStyle().
			Foreground(ColorMuted)

	// Progress styles
	SpinnerStyle = lipgloss.NewStyle().
			Foreground(ColorPrimary)

	ProgressStyle = lipgloss.NewStyle().
			Foreground(ColorSuccess)

	// Border styles
	BorderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorMuted)

	// Dialog styles
	DialogStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorPrimary).
			Padding(1, 2)

	// Error styles
	ErrorStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorDanger)

	// Info styles
	InfoStyle = lipgloss.NewStyle().
			Foreground(ColorCyan)

	// Pin/Star style
	PinStyle = lipgloss.NewStyle().
			Foreground(ColorWarning)
)

// MutedStyle returns a muted text style
func MutedStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(ColorMuted)
}

// ScoreBar returns a visual bar representation of a score
func ScoreBar(score float64, width int) string {
	filled := int(score / 100.0 * float64(width))
	if filled > width {
		filled = width
	}

	bar := ""
	for i := 0; i < filled; i++ {
		bar += "█"
	}
	for i := filled; i < width; i++ {
		bar += "░"
	}

	var style lipgloss.Style
	switch {
	case score >= 75:
		style = ScoreHighStyle
	case score >= 50:
		style = ScoreMediumStyle
	default:
		style = ScoreLowStyle
	}

	return style.Render(bar)
}

// FormatChange returns styled change percentage
func FormatChange(change float64) string {
	if change >= 0 {
		return PriceUpStyle.Render("+" + formatPercent(change))
	}
	return PriceDownStyle.Render(formatPercent(change))
}

// FormatRSI returns styled RSI value
func FormatRSI(rsi float64) string {
	text := formatFloat(rsi, 1)
	switch {
	case rsi < 30:
		return RSIOversoldStyle.Render(text)
	case rsi > 70:
		return RSIOverboughtStyle.Render(text)
	default:
		return RSINeutralStyle.Render(text)
	}
}

// FormatPBV returns styled PBV value
func FormatPBV(pbv float64) string {
	text := formatFloat(pbv, 2)
	switch {
	case pbv < 1.0:
		return PBVLowStyle.Render(text)
	case pbv > 2.0:
		return PBVHighStyle.Render(text)
	default:
		return PBVNeutralStyle.Render(text)
	}
}

// FormatScore returns styled score
func FormatScore(score float64) string {
	text := formatFloat(score, 0)
	switch {
	case score >= 75:
		return ScoreHighStyle.Render(text)
	case score >= 50:
		return ScoreMediumStyle.Render(text)
	default:
		return ScoreLowStyle.Render(text)
	}
}

// Helper functions
func formatFloat(f float64, decimals int) string {
	format := "%." + string(rune('0'+decimals)) + "f"
	return lipgloss.NewStyle().Render(sprintf(format, f))
}

func formatPercent(f float64) string {
	return sprintf("%.2f%%", f)
}

func sprintf(format string, a ...interface{}) string {
	return lipgloss.NewStyle().Render(sprintfReal(format, a...))
}

func sprintfReal(format string, args ...interface{}) string {
	switch format {
	case "%.0f":
		return sprintfFloat(args[0].(float64), 0)
	case "%.1f":
		return sprintfFloat(args[0].(float64), 1)
	case "%.2f":
		return sprintfFloat(args[0].(float64), 2)
	case "%.2f%%":
		return sprintfFloat(args[0].(float64), 2) + "%"
	default:
		return ""
	}
}

func sprintfFloat(f float64, decimals int) string {
	if decimals == 0 {
		return lipgloss.NewStyle().Render(intToString(int(f)))
	}

	intPart := int(f)
	fracPart := f - float64(intPart)
	if fracPart < 0 {
		fracPart = -fracPart
	}

	multiplier := 1
	for i := 0; i < decimals; i++ {
		multiplier *= 10
	}

	frac := int(fracPart*float64(multiplier) + 0.5)

	fracStr := intToString(frac)
	for len(fracStr) < decimals {
		fracStr = "0" + fracStr
	}

	return intToString(intPart) + "." + fracStr
}

func intToString(n int) string {
	if n == 0 {
		return "0"
	}

	negative := n < 0
	if negative {
		n = -n
	}

	digits := ""
	for n > 0 {
		digits = string(rune('0'+n%10)) + digits
		n /= 10
	}

	if negative {
		digits = "-" + digits
	}

	return digits
}
