package views

import (
	"fmt"
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"stockmap/internal/screener"
	"stockmap/internal/styles"
	"stockmap/internal/ui/components"
)

// Details shows detailed information for a stock
type Details struct {
	width     int
	height    int
	stock     *screener.ScreenResult
	showChart bool
}

// NewDetails creates a new details view
func NewDetails() *Details {
	return &Details{}
}

// SetSize sets the view dimensions
func (d *Details) SetSize(width, height int) {
	d.width = width
	d.height = height
}

// SetStock sets the stock to display
func (d *Details) SetStock(stock *screener.ScreenResult) {
	d.stock = stock
}

// ToggleChart toggles the chart view
func (d *Details) ToggleChart() {
	d.showChart = !d.showChart
}

// IsChartVisible returns whether chart is visible
func (d *Details) IsChartVisible() bool {
	return d.showChart
}

// View renders the details view
func (d *Details) View() string {
	if d.stock == nil {
		return centerText(styles.MutedStyle().Render("No stock selected"), d.width)
	}

	var b strings.Builder

	s := d.stock

	// Header
	pinIndicator := ""
	if s.IsPinned {
		pinIndicator = styles.PinStyle.Render("* ")
	}

	header := pinIndicator + styles.TickerStyle.Render(s.Symbol)
	if s.Name != "" {
		header += " - " + styles.TitleStyle.Render(s.Name)
	}
	b.WriteString(centerText(header, d.width))
	b.WriteString("\n")
	b.WriteString(components.RenderDivider(d.width))
	b.WriteString("\n\n")

	if d.showChart {
		// Chart view
		b.WriteString(d.renderChart(s))
	} else {
		// Normal details view
		b.WriteString(d.renderDetailsView(s))
	}

	// Footer
	chartHelp := "[G] Show Chart"
	if d.showChart {
		chartHelp = "[G] Hide Chart"
	}
	footer := styles.HelpStyle.Render(fmt.Sprintf("Press [A] to add  [R] to remove  %s  [ESC] to go back", chartHelp))
	b.WriteString(centerText(footer, d.width))

	return b.String()
}

// renderDetailsView renders the normal details view
func (d *Details) renderDetailsView(s *screener.ScreenResult) string {
	var b strings.Builder

	// Price section
	priceSection := d.renderSection("PRICE", [][]string{
		{"Current", fmt.Sprintf("$%.2f", s.Price)},
		{"Change", fmt.Sprintf("%+.2f (%.2f%%)", s.Change, s.ChangePercent)},
		{"Volume", components.FormatLargeNumber(s.Volume)},
		{"Market Cap", components.FormatLargeNumber(s.MarketCap)},
	})

	// Technical section
	technicalSection := d.renderSection("TECHNICAL", [][]string{
		{"RSI (14)", fmt.Sprintf("%.1f", s.RSI)},
		{"ATR (14)", fmt.Sprintf("%.2f", s.ATR)},
		{"SMA 20", fmt.Sprintf("$%.2f", s.SMA20)},
		{"SMA 50", fmt.Sprintf("$%.2f", s.SMA50)},
		{"Volatility", fmt.Sprintf("%.1f%%", s.Volatility)},
	})

	// Valuation section
	valuationSection := d.renderSection("VALUATION", [][]string{
		{"P/B Ratio", fmt.Sprintf("%.2f", s.PBV)},
		{"P/E Ratio", fmt.Sprintf("%.2f", s.PERatio)},
		{"EPS", fmt.Sprintf("$%.2f", s.EPS)},
		{"Book Value", fmt.Sprintf("$%.2f", s.BookValue)},
		{"Graham Number", fmt.Sprintf("$%.2f", s.GrahamNumber)},
		{"Graham Upside", fmt.Sprintf("%.1f%%", s.GrahamUpside)},
	})

	// Risk section
	riskSection := d.renderSection("RISK/REWARD", [][]string{
		{"Take Profit", fmt.Sprintf("$%.2f", s.TakeProfit)},
		{"Stop Loss", fmt.Sprintf("$%.2f", s.StopLoss)},
		{"Risk:Reward", fmt.Sprintf("1:%.1f", s.RiskRatio)},
	})

	// Score section
	scoreSection := d.renderScoreSection(s)

	// Layout columns
	col1 := lipgloss.JoinVertical(lipgloss.Left, priceSection, technicalSection)
	col2 := lipgloss.JoinVertical(lipgloss.Left, valuationSection, riskSection)

	colWidth := (d.width - 4) / 2
	col1Styled := lipgloss.NewStyle().Width(colWidth).Render(col1)
	col2Styled := lipgloss.NewStyle().Width(colWidth).Render(col2)

	columns := lipgloss.JoinHorizontal(lipgloss.Top, col1Styled, "  ", col2Styled)
	b.WriteString(columns)
	b.WriteString("\n\n")

	// Score section (full width)
	b.WriteString(scoreSection)
	b.WriteString("\n\n")

	// Signals
	b.WriteString(d.renderSignals(s))
	b.WriteString("\n\n")

	return b.String()
}

// renderChart renders an ASCII price chart
func (d *Details) renderChart(s *screener.ScreenResult) string {
	var b strings.Builder

	b.WriteString(styles.TitleStyle.Render("PRICE CHART (60 Days)"))
	b.WriteString("\n\n")

	prices := s.HistoricalPrices
	if len(prices) == 0 {
		b.WriteString(centerText(styles.MutedStyle().Render("No historical data available"), d.width))
		b.WriteString("\n\n")
		return b.String()
	}

	// Chart dimensions
	chartWidth := d.width - 10
	if chartWidth < 40 {
		chartWidth = 40
	}
	chartHeight := 12

	// Sample prices to fit width
	sampledPrices := samplePrices(prices, chartWidth)

	// Find min and max
	minPrice, maxPrice := sampledPrices[0], sampledPrices[0]
	for _, p := range sampledPrices {
		if p < minPrice {
			minPrice = p
		}
		if p > maxPrice {
			maxPrice = p
		}
	}

	// Add some padding to range
	priceRange := maxPrice - minPrice
	if priceRange == 0 {
		priceRange = 1
	}

	// Render chart
	for row := 0; row < chartHeight; row++ {
		// Y-axis label
		price := maxPrice - (float64(row)/float64(chartHeight-1))*priceRange
		label := fmt.Sprintf("%7.2f ", price)
		b.WriteString(styles.MutedStyle().Render(label))

		// Chart row
		for _, p := range sampledPrices {
			normalized := (p - minPrice) / priceRange
			chartRow := chartHeight - 1 - int(normalized*float64(chartHeight-1))

			if chartRow == row {
				// Price is at this level
				if len(prices) > 1 && p >= prices[len(prices)-1] {
					b.WriteString(styles.ScoreHighStyle.Render("*"))
				} else {
					b.WriteString(styles.ScoreLowStyle.Render("*"))
				}
			} else if chartRow < row {
				// Below the price line
				b.WriteString(" ")
			} else {
				// Above the price line
				b.WriteString(" ")
			}
		}
		b.WriteString("\n")
	}

	// X-axis
	b.WriteString(strings.Repeat(" ", 8))
	b.WriteString(styles.MutedStyle().Render(strings.Repeat("-", len(sampledPrices))))
	b.WriteString("\n")

	// X-axis labels
	b.WriteString(strings.Repeat(" ", 8))
	b.WriteString(styles.MutedStyle().Render("60d ago"))
	b.WriteString(strings.Repeat(" ", len(sampledPrices)-14))
	b.WriteString(styles.MutedStyle().Render("Today"))
	b.WriteString("\n\n")

	// Strategy visualization
	b.WriteString(d.renderStrategyLevels(s))
	b.WriteString("\n")

	return b.String()
}

// renderStrategyLevels shows entry, SL, TP levels visually
func (d *Details) renderStrategyLevels(s *screener.ScreenResult) string {
	var b strings.Builder

	b.WriteString(styles.TitleStyle.Render("STRATEGY LEVELS"))
	b.WriteString("\n\n")

	// Calculate visual positions
	levelWidth := d.width - 20
	if levelWidth < 30 {
		levelWidth = 30
	}

	// Find range
	minLevel := math.Min(s.StopLoss, math.Min(s.Price, s.TakeProfit)) * 0.95
	maxLevel := math.Max(s.TakeProfit, math.Max(s.Price, s.StopLoss)) * 1.05
	priceRange := maxLevel - minLevel
	if priceRange == 0 {
		priceRange = 1
	}

	// Calculate positions
	slPos := int((s.StopLoss - minLevel) / priceRange * float64(levelWidth))
	pricePos := int((s.Price - minLevel) / priceRange * float64(levelWidth))
	tpPos := int((s.TakeProfit - minLevel) / priceRange * float64(levelWidth))

	// Clamp positions
	slPos = clamp(slPos, 0, levelWidth-1)
	pricePos = clamp(pricePos, 0, levelWidth-1)
	tpPos = clamp(tpPos, 0, levelWidth-1)

	// Render level bar
	line := make([]rune, levelWidth)
	for i := range line {
		line[i] = '-'
	}

	// Mark levels
	if slPos >= 0 && slPos < levelWidth {
		line[slPos] = 'S'
	}
	if pricePos >= 0 && pricePos < levelWidth {
		line[pricePos] = 'P'
	}
	if tpPos >= 0 && tpPos < levelWidth {
		line[tpPos] = 'T'
	}

	// Build colored line
	lineStr := string(line)
	coloredLine := ""
	for i := range lineStr {
		if i == slPos {
			coloredLine += styles.ScoreLowStyle.Render("S")
		} else if i == pricePos {
			coloredLine += styles.InfoStyle.Render("P")
		} else if i == tpPos {
			coloredLine += styles.ScoreHighStyle.Render("T")
		} else {
			coloredLine += styles.MutedStyle().Render("-")
		}
	}

	b.WriteString("  " + coloredLine + "\n")

	// Legend
	legend := fmt.Sprintf("  %s Stop Loss: $%.2f   %s Price: $%.2f   %s Take Profit: $%.2f",
		styles.ScoreLowStyle.Render("S"),
		s.StopLoss,
		styles.InfoStyle.Render("P"),
		s.Price,
		styles.ScoreHighStyle.Render("T"),
		s.TakeProfit,
	)
	b.WriteString(legend + "\n")

	// Risk/Reward info
	riskInfo := fmt.Sprintf("  Risk: $%.2f (%.1f%%)  |  Reward: $%.2f (%.1f%%)  |  R:R = 1:%.1f",
		s.Price-s.StopLoss,
		(s.Price-s.StopLoss)/s.Price*100,
		s.TakeProfit-s.Price,
		(s.TakeProfit-s.Price)/s.Price*100,
		s.RiskRatio,
	)
	b.WriteString(styles.MutedStyle().Render(riskInfo) + "\n")

	return b.String()
}

// samplePrices samples prices to fit a given width
func samplePrices(prices []float64, width int) []float64 {
	if len(prices) <= width {
		return prices
	}

	sampled := make([]float64, width)
	step := float64(len(prices)) / float64(width)

	for i := 0; i < width; i++ {
		idx := int(float64(i) * step)
		if idx >= len(prices) {
			idx = len(prices) - 1
		}
		sampled[i] = prices[idx]
	}

	return sampled
}

// clamp clamps a value between min and max
func clamp(val, min, max int) int {
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}

// renderSection renders a labeled section with key-value pairs
func (d *Details) renderSection(title string, items [][]string) string {
	var b strings.Builder

	b.WriteString(styles.TitleStyle.Render(title))
	b.WriteString("\n")

	for _, item := range items {
		key := styles.MutedStyle().Render(item[0] + ": ")
		value := styles.InfoStyle.Render(item[1])
		b.WriteString("  " + key + value + "\n")
	}

	return b.String()
}

// renderScoreSection renders the score section with bars
func (d *Details) renderScoreSection(s *screener.ScreenResult) string {
	var b strings.Builder

	b.WriteString(components.RenderDivider(d.width))
	b.WriteString("\n")
	b.WriteString(styles.TitleStyle.Render("CONFLUENCE SCORE"))
	b.WriteString("\n\n")

	// Main score
	mainScore := fmt.Sprintf("Overall: %.0f/100 ", s.ConfluenceScore)
	mainBar := styles.ScoreBar(s.ConfluenceScore, 20)
	grade := screener.ScoreToGrade(s.ConfluenceScore)
	gradeStyled := ""
	if s.ConfluenceScore >= 75 {
		gradeStyled = styles.ScoreHighStyle.Render(" (" + grade + ")")
	} else if s.ConfluenceScore >= 50 {
		gradeStyled = styles.ScoreMediumStyle.Render(" (" + grade + ")")
	} else {
		gradeStyled = styles.ScoreLowStyle.Render(" (" + grade + ")")
	}
	b.WriteString("  " + mainScore + mainBar + gradeStyled + "\n\n")

	// Component scores
	scores := []struct {
		name  string
		value float64
	}{
		{"Technical", s.TechnicalScore},
		{"Valuation", s.ValuationScore},
		{"Risk-Adj", s.RiskScore},
	}

	for _, score := range scores {
		label := styles.MutedStyle().Render(fmt.Sprintf("  %-12s", score.name))
		value := fmt.Sprintf("%.0f ", score.value)
		bar := styles.ScoreBar(score.value, 15)
		b.WriteString(label + value + bar + "\n")
	}

	return b.String()
}

// renderSignals renders buy/sell signals
func (d *Details) renderSignals(s *screener.ScreenResult) string {
	var b strings.Builder

	b.WriteString(styles.TitleStyle.Render("SIGNALS"))
	b.WriteString("\n")

	signals := []struct {
		condition bool
		bullish   bool
		text      string
	}{
		{s.RSI < 30, true, "RSI Oversold (<30)"},
		{s.RSI < 35, true, "RSI Near Oversold (<35)"},
		{s.PBV < 1.0, true, "Trading Below Book Value"},
		{s.PBV < 1.5, true, "Low P/B Ratio (<1.5)"},
		{s.GrahamUpside > 30, true, "Strong Graham Upside (>30%)"},
		{s.Price < s.SMA20, true, "Below SMA20"},
		{s.RiskRatio >= 2.0, true, "Favorable Risk:Reward (>=1:2)"},
		{s.RSI > 70, false, "RSI Overbought (>70)"},
		{s.PBV > 3.0, false, "High P/B Ratio (>3)"},
	}

	bullishCount := 0
	bearishCount := 0

	for _, sig := range signals {
		if sig.condition {
			var icon, text string
			if sig.bullish {
				icon = styles.ScoreHighStyle.Render("+")
				text = styles.ScoreHighStyle.Render(sig.text)
				bullishCount++
			} else {
				icon = styles.ScoreLowStyle.Render("-")
				text = styles.ScoreLowStyle.Render(sig.text)
				bearishCount++
			}
			b.WriteString("  " + icon + " " + text + "\n")
		}
	}

	if bullishCount == 0 && bearishCount == 0 {
		b.WriteString("  " + styles.MutedStyle().Render("No significant signals") + "\n")
	}

	return b.String()
}
