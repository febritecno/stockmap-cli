package views

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/febritecno/stockmap/internal/screener"
	"github.com/febritecno/stockmap/internal/styles"
	"github.com/febritecno/stockmap/internal/ui/components"
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

	// MACD section
	macdSection := d.renderMACDSection(s)

	// Bollinger section
	bollingerSection := d.renderBollingerSection(s)

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

	// Responsive layout based on width
	if d.width < 80 {
		// Single column layout for narrow screens
		b.WriteString(priceSection)
		b.WriteString(technicalSection)
		b.WriteString(valuationSection)
		b.WriteString(riskSection)
	} else if d.width < 120 {
		// Two column layout for medium screens
		col1 := lipgloss.JoinVertical(lipgloss.Left, priceSection, technicalSection, macdSection)
		col2 := lipgloss.JoinVertical(lipgloss.Left, valuationSection, riskSection, bollingerSection)

		colWidth := (d.width - 4) / 2
		col1Styled := lipgloss.NewStyle().Width(colWidth).Render(col1)
		col2Styled := lipgloss.NewStyle().Width(colWidth).Render(col2)

		columns := lipgloss.JoinHorizontal(lipgloss.Top, col1Styled, "  ", col2Styled)
		b.WriteString(columns)
	} else {
		// Three column layout for wide screens
		col1 := lipgloss.JoinVertical(lipgloss.Left, priceSection, technicalSection)
		col2 := lipgloss.JoinVertical(lipgloss.Left, macdSection, bollingerSection)
		col3 := lipgloss.JoinVertical(lipgloss.Left, valuationSection, riskSection)

		colWidth := (d.width - 6) / 3
		col1Styled := lipgloss.NewStyle().Width(colWidth).Render(col1)
		col2Styled := lipgloss.NewStyle().Width(colWidth).Render(col2)
		col3Styled := lipgloss.NewStyle().Width(colWidth).Render(col3)

		columns := lipgloss.JoinHorizontal(lipgloss.Top, col1Styled, " ", col2Styled, " ", col3Styled)
		b.WriteString(columns)
	}
	b.WriteString("\n\n")

	// Score section (full width)
	b.WriteString(scoreSection)
	b.WriteString("\n\n")

	// Signals
	b.WriteString(d.renderSignals(s))
	b.WriteString("\n\n")

	return b.String()
}

// renderMACDSection renders MACD indicators
func (d *Details) renderMACDSection(s *screener.ScreenResult) string {
	var b strings.Builder

	b.WriteString(styles.TitleStyle.Render("MACD (12,26,9)"))
	b.WriteString("\n")

	// MACD value with color
	macdStr := fmt.Sprintf("%.2f", s.MACD)
	if s.MACD > 0 {
		macdStr = styles.ScoreHighStyle.Render(macdStr)
	} else if s.MACD < 0 {
		macdStr = styles.ScoreLowStyle.Render(macdStr)
	}
	b.WriteString("  " + styles.MutedStyle().Render("MACD: ") + macdStr + "\n")

	// Signal
	b.WriteString("  " + styles.MutedStyle().Render("Signal: ") + styles.InfoStyle.Render(fmt.Sprintf("%.2f", s.MACDSignal)) + "\n")

	// Histogram with color
	histStr := fmt.Sprintf("%.2f", s.MACDHistogram)
	if s.MACDHistogram > 0 {
		histStr = styles.ScoreHighStyle.Render("+" + histStr)
	} else if s.MACDHistogram < 0 {
		histStr = styles.ScoreLowStyle.Render(histStr)
	}
	b.WriteString("  " + styles.MutedStyle().Render("Histogram: ") + histStr + "\n")

	// Crossover signal
	crossStr := s.MACDCrossover
	if crossStr == "" {
		crossStr = "none"
	}
	switch crossStr {
	case "bullish":
		crossStr = styles.ScoreHighStyle.Render("BULLISH")
	case "bearish":
		crossStr = styles.ScoreLowStyle.Render("BEARISH")
	default:
		crossStr = styles.MutedStyle().Render("None")
	}
	b.WriteString("  " + styles.MutedStyle().Render("Crossover: ") + crossStr + "\n")

	return b.String()
}

// renderBollingerSection renders Bollinger Bands
func (d *Details) renderBollingerSection(s *screener.ScreenResult) string {
	var b strings.Builder

	b.WriteString(styles.TitleStyle.Render("BOLLINGER (20,2)"))
	b.WriteString("\n")

	b.WriteString("  " + styles.MutedStyle().Render("Upper: ") + styles.InfoStyle.Render(fmt.Sprintf("$%.2f", s.BBUpper)) + "\n")
	b.WriteString("  " + styles.MutedStyle().Render("Middle: ") + styles.InfoStyle.Render(fmt.Sprintf("$%.2f", s.BBMiddle)) + "\n")
	b.WriteString("  " + styles.MutedStyle().Render("Lower: ") + styles.InfoStyle.Render(fmt.Sprintf("$%.2f", s.BBLower)) + "\n")

	// Band width
	b.WriteString("  " + styles.MutedStyle().Render("Width: ") + styles.InfoStyle.Render(fmt.Sprintf("%.1f%%", s.BBWidth)) + "\n")

	// %B indicator (position within bands)
	var percentBStr string
	if s.BBPercentB >= 1.0 {
		percentBStr = styles.ScoreLowStyle.Render(fmt.Sprintf("%.2f (Above Upper)", s.BBPercentB))
	} else if s.BBPercentB <= 0 {
		percentBStr = styles.ScoreHighStyle.Render(fmt.Sprintf("%.2f (Below Lower)", s.BBPercentB))
	} else if s.BBPercentB > 0.8 {
		percentBStr = styles.ScoreMediumStyle.Render(fmt.Sprintf("%.2f (Near Upper)", s.BBPercentB))
	} else if s.BBPercentB < 0.2 {
		percentBStr = styles.ScoreMediumStyle.Render(fmt.Sprintf("%.2f (Near Lower)", s.BBPercentB))
	} else {
		percentBStr = styles.InfoStyle.Render(fmt.Sprintf("%.2f", s.BBPercentB))
	}
	b.WriteString("  " + styles.MutedStyle().Render("%B: ") + percentBStr + "\n")

	// Squeeze indicator
	if s.BBSqueeze {
		b.WriteString("  " + styles.ScoreHighStyle.Render("SQUEEZE") + styles.MutedStyle().Render(" - Low volatility") + "\n")
	}

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

// renderStrategyLevels shows entry, SL, TP levels visually with support/resistance
func (d *Details) renderStrategyLevels(s *screener.ScreenResult) string {
	var b strings.Builder

	b.WriteString(styles.TitleStyle.Render("STRATEGY LEVELS"))
	b.WriteString("\n\n")

	// Calculate support and resistance levels from historical prices
	support, resistance := calculateSupportResistance(s.HistoricalPrices, s.Price)

	// Determine BULL/BEAR position
	isBull := s.Price > s.SMA20 && s.SMA20 > 0
	isBear := s.Price < s.SMA20 && s.SMA20 > 0

	// Determine position relative to support/resistance
	nearSupport := len(support) > 0 && s.Price <= support[0]*1.02
	nearResistance := len(resistance) > 0 && s.Price >= resistance[0]*0.98

	// Market Position Indicator
	var positionText string
	if isBull {
		positionText = styles.ScoreHighStyle.Render("BULL") + styles.MutedStyle().Render(" (Above SMA20)")
	} else if isBear {
		positionText = styles.ScoreLowStyle.Render("BEAR") + styles.MutedStyle().Render(" (Below SMA20)")
	} else {
		positionText = styles.ScoreMediumStyle.Render("NEUTRAL")
	}
	b.WriteString("  Position: " + positionText + "\n")

	// Support/Resistance Position
	var srText string
	if nearSupport {
		srText = styles.ScoreHighStyle.Render("Near SUPPORT") + " - Good entry zone"
	} else if nearResistance {
		srText = styles.ScoreLowStyle.Render("Near RESISTANCE") + " - Caution zone"
	} else {
		srText = styles.MutedStyle().Render("Between S/R levels")
	}
	b.WriteString("  S/R Zone: " + srText + "\n\n")

	// Calculate visual positions
	levelWidth := d.width - 20
	if levelWidth < 40 {
		levelWidth = 40
	}

	// Find range including support/resistance
	minLevel := s.StopLoss
	maxLevel := s.TakeProfit

	if len(support) > 0 && support[len(support)-1] < minLevel {
		minLevel = support[len(support)-1]
	}
	if len(resistance) > 0 && resistance[len(resistance)-1] > maxLevel {
		maxLevel = resistance[len(resistance)-1]
	}

	minLevel *= 0.98
	maxLevel *= 1.02
	priceRange := maxLevel - minLevel
	if priceRange == 0 {
		priceRange = 1
	}

	// Build the level line
	line := make([]string, levelWidth)
	for i := range line {
		line[i] = styles.MutedStyle().Render("-")
	}

	// Helper to calculate position
	calcPos := func(price float64) int {
		pos := int((price - minLevel) / priceRange * float64(levelWidth-1))
		return clamp(pos, 0, levelWidth-1)
	}

	// Mark support levels (s1, s2)
	for i, sup := range support {
		if i >= 2 {
			break
		}
		pos := calcPos(sup)
		label := fmt.Sprintf("%d", i+1)
		line[pos] = styles.ScoreHighStyle.Render(label)
	}

	// Mark resistance levels (r1, r2)
	for i, res := range resistance {
		if i >= 2 {
			break
		}
		pos := calcPos(res)
		label := fmt.Sprintf("%d", i+1)
		line[pos] = styles.ScoreLowStyle.Render(label)
	}

	// Mark SL, Price, TP
	slPos := calcPos(s.StopLoss)
	pricePos := calcPos(s.Price)
	tpPos := calcPos(s.TakeProfit)

	line[slPos] = styles.ScoreLowStyle.Render("S")
	line[tpPos] = styles.ScoreHighStyle.Render("T")
	line[pricePos] = styles.InfoStyle.Render("P")

	// Render line
	b.WriteString("  ")
	for _, c := range line {
		b.WriteString(c)
	}
	b.WriteString("\n")

	// Labels
	b.WriteString(fmt.Sprintf("  %s=StopLoss  %s=Price  %s=TakeProfit  %s=Support  %s=Resistance\n",
		styles.ScoreLowStyle.Render("S"),
		styles.InfoStyle.Render("P"),
		styles.ScoreHighStyle.Render("T"),
		styles.ScoreHighStyle.Render("1,2"),
		styles.ScoreLowStyle.Render("1,2"),
	))
	b.WriteString("\n")

	// Price levels table
	b.WriteString(styles.TitleStyle.Render("  KEY LEVELS"))
	b.WriteString("\n")

	// Stop Loss and Take Profit
	b.WriteString(fmt.Sprintf("  %s $%.2f  |  %s $%.2f  |  R:R 1:%.1f\n",
		styles.ScoreLowStyle.Render("Stop Loss:"),
		s.StopLoss,
		styles.ScoreHighStyle.Render("Take Profit:"),
		s.TakeProfit,
		s.RiskRatio,
	))

	// Support levels
	b.WriteString(fmt.Sprintf("  %s ", styles.ScoreHighStyle.Render("Support:")))
	if len(support) > 0 {
		for i, sup := range support {
			if i >= 3 {
				break
			}
			if i > 0 {
				b.WriteString(" | ")
			}
			pct := (s.Price - sup) / s.Price * 100
			b.WriteString(fmt.Sprintf("S%d: $%.2f (%.1f%%)", i+1, sup, pct))
		}
	} else {
		b.WriteString("N/A")
	}
	b.WriteString("\n")

	// Resistance levels
	b.WriteString(fmt.Sprintf("  %s ", styles.ScoreLowStyle.Render("Resistance:")))
	if len(resistance) > 0 {
		for i, res := range resistance {
			if i >= 3 {
				break
			}
			if i > 0 {
				b.WriteString(" | ")
			}
			pct := (res - s.Price) / s.Price * 100
			b.WriteString(fmt.Sprintf("R%d: $%.2f (+%.1f%%)", i+1, res, pct))
		}
	} else {
		b.WriteString("N/A")
	}
	b.WriteString("\n")

	return b.String()
}

// calculateSupportResistance calculates support and resistance levels from historical prices
func calculateSupportResistance(prices []float64, currentPrice float64) (support []float64, resistance []float64) {
	if len(prices) < 10 {
		return nil, nil
	}

	// Find local minima (support) and maxima (resistance)
	var allSupports, allResistances []float64

	for i := 2; i < len(prices)-2; i++ {
		// Local minimum - support
		if prices[i] < prices[i-1] && prices[i] < prices[i-2] &&
			prices[i] < prices[i+1] && prices[i] < prices[i+2] {
			allSupports = append(allSupports, prices[i])
		}
		// Local maximum - resistance
		if prices[i] > prices[i-1] && prices[i] > prices[i-2] &&
			prices[i] > prices[i+1] && prices[i] > prices[i+2] {
			allResistances = append(allResistances, prices[i])
		}
	}

	// Filter: support must be below current price, resistance above
	for _, s := range allSupports {
		if s < currentPrice {
			support = append(support, s)
		}
	}
	for _, r := range allResistances {
		if r > currentPrice {
			resistance = append(resistance, r)
		}
	}

	// Sort support descending (nearest first), resistance ascending (nearest first)
	sort.Slice(support, func(i, j int) bool {
		return support[i] > support[j]
	})
	sort.Slice(resistance, func(i, j int) bool {
		return resistance[i] < resistance[j]
	})

	// Keep only top 3
	if len(support) > 3 {
		support = support[:3]
	}
	if len(resistance) > 3 {
		resistance = resistance[:3]
	}

	return support, resistance
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
		// MACD signals
		{s.MACDCrossover == "bullish", true, "MACD Bullish Crossover"},
		{s.MACD > 0 && s.MACDHistogram > 0, true, "MACD Positive Momentum"},
		{s.MACDCrossover == "bearish", false, "MACD Bearish Crossover"},
		// Bollinger Band signals
		{s.BBPercentB <= 0, true, "Price Below Lower BB (Oversold)"},
		{s.BBPercentB < 0.2 && s.BBPercentB > 0, true, "Price Near Lower BB"},
		{s.BBSqueeze, true, "BB Squeeze (Breakout Potential)"},
		{s.BBPercentB >= 1.0, false, "Price Above Upper BB (Overbought)"},
		{s.BBPercentB > 0.8 && s.BBPercentB < 1.0, false, "Price Near Upper BB"},
		// Bearish signals
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
