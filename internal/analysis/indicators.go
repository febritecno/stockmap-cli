package analysis

import (
	"math"
)

// RSI calculates the Relative Strength Index for a given price series
// period is typically 14 days
func RSI(prices []float64, period int) float64 {
	if len(prices) < period+1 {
		return 50.0 // neutral if not enough data
	}

	var gains, losses float64

	// Calculate initial average gain/loss
	for i := 1; i <= period; i++ {
		change := prices[i] - prices[i-1]
		if change > 0 {
			gains += change
		} else {
			losses += math.Abs(change)
		}
	}

	avgGain := gains / float64(period)
	avgLoss := losses / float64(period)

	// Apply Wilder's smoothing for remaining periods
	for i := period + 1; i < len(prices); i++ {
		change := prices[i] - prices[i-1]
		if change > 0 {
			avgGain = (avgGain*float64(period-1) + change) / float64(period)
			avgLoss = (avgLoss * float64(period-1)) / float64(period)
		} else {
			avgGain = (avgGain * float64(period-1)) / float64(period)
			avgLoss = (avgLoss*float64(period-1) + math.Abs(change)) / float64(period)
		}
	}

	if avgLoss == 0 {
		return 100.0
	}

	rs := avgGain / avgLoss
	rsi := 100.0 - (100.0 / (1.0 + rs))

	return rsi
}

// ATR calculates the Average True Range
// high, low, close are price arrays; period is typically 14
func ATR(high, low, close []float64, period int) float64 {
	if len(high) < period+1 || len(low) < period+1 || len(close) < period+1 {
		return 0.0
	}

	n := len(high)
	trueRanges := make([]float64, n)

	// First TR is just high - low
	trueRanges[0] = high[0] - low[0]

	// Calculate True Range for each period
	for i := 1; i < n; i++ {
		hl := high[i] - low[i]
		hc := math.Abs(high[i] - close[i-1])
		lc := math.Abs(low[i] - close[i-1])
		trueRanges[i] = math.Max(hl, math.Max(hc, lc))
	}

	// Calculate initial ATR (simple average)
	var sum float64
	for i := 0; i < period; i++ {
		sum += trueRanges[i]
	}
	atr := sum / float64(period)

	// Apply Wilder's smoothing
	for i := period; i < n; i++ {
		atr = (atr*float64(period-1) + trueRanges[i]) / float64(period)
	}

	return atr
}

// SMA calculates Simple Moving Average
func SMA(prices []float64, period int) float64 {
	if len(prices) < period {
		return 0.0
	}

	var sum float64
	start := len(prices) - period
	for i := start; i < len(prices); i++ {
		sum += prices[i]
	}

	return sum / float64(period)
}

// EMA calculates Exponential Moving Average
func EMA(prices []float64, period int) float64 {
	if len(prices) < period {
		return 0.0
	}

	multiplier := 2.0 / float64(period+1)
	ema := SMA(prices[:period], period) // Start with SMA

	for i := period; i < len(prices); i++ {
		ema = (prices[i]-ema)*multiplier + ema
	}

	return ema
}

// IsOversold returns true if RSI is below threshold (typically 30)
func IsOversold(rsi float64, threshold float64) bool {
	return rsi < threshold
}

// IsOverbought returns true if RSI is above threshold (typically 70)
func IsOverbought(rsi float64, threshold float64) bool {
	return rsi > threshold
}

// MACDResult holds the MACD calculation results
type MACDResult struct {
	MACD      float64 // MACD line (fast EMA - slow EMA)
	Signal    float64 // Signal line (EMA of MACD)
	Histogram float64 // Histogram (MACD - Signal)
	IsBullish bool    // MACD above signal line
	IsBearish bool    // MACD below signal line
	Crossover string  // "bullish", "bearish", or "none"
}

// MACD calculates Moving Average Convergence Divergence
// Default periods: fast=12, slow=26, signal=9
func MACD(prices []float64, fastPeriod, slowPeriod, signalPeriod int) MACDResult {
	result := MACDResult{}

	if len(prices) < slowPeriod+signalPeriod {
		return result
	}

	// Calculate EMAs for MACD line
	fastEMA := EMAFull(prices, fastPeriod)
	slowEMA := EMAFull(prices, slowPeriod)

	if len(fastEMA) == 0 || len(slowEMA) == 0 {
		return result
	}

	// Calculate MACD line (difference between fast and slow EMA)
	macdLine := make([]float64, len(slowEMA))
	offset := len(fastEMA) - len(slowEMA)
	for i := 0; i < len(slowEMA); i++ {
		macdLine[i] = fastEMA[i+offset] - slowEMA[i]
	}

	// Calculate Signal line (EMA of MACD line)
	signalLine := EMAFull(macdLine, signalPeriod)

	if len(signalLine) == 0 {
		return result
	}

	// Get current values
	result.MACD = macdLine[len(macdLine)-1]
	result.Signal = signalLine[len(signalLine)-1]
	result.Histogram = result.MACD - result.Signal

	// Determine bullish/bearish
	result.IsBullish = result.MACD > result.Signal
	result.IsBearish = result.MACD < result.Signal

	// Detect crossovers (check last 2 values)
	if len(macdLine) >= 2 && len(signalLine) >= 2 {
		prevMacd := macdLine[len(macdLine)-2]
		prevSignal := signalLine[len(signalLine)-2]

		// Bullish crossover: MACD crosses above Signal
		if prevMacd <= prevSignal && result.MACD > result.Signal {
			result.Crossover = "bullish"
		} else if prevMacd >= prevSignal && result.MACD < result.Signal {
			// Bearish crossover: MACD crosses below Signal
			result.Crossover = "bearish"
		} else {
			result.Crossover = "none"
		}
	}

	return result
}

// EMAFull calculates EMA for entire price series and returns all values
func EMAFull(prices []float64, period int) []float64 {
	if len(prices) < period {
		return nil
	}

	ema := make([]float64, len(prices)-period+1)
	multiplier := 2.0 / float64(period+1)

	// First EMA is SMA
	var sum float64
	for i := 0; i < period; i++ {
		sum += prices[i]
	}
	ema[0] = sum / float64(period)

	// Calculate remaining EMAs
	for i := period; i < len(prices); i++ {
		ema[i-period+1] = (prices[i]-ema[i-period])*multiplier + ema[i-period]
	}

	return ema
}

// BollingerResult holds Bollinger Bands calculation results
type BollingerResult struct {
	Upper      float64 // Upper band (SMA + k*StdDev)
	Middle     float64 // Middle band (SMA)
	Lower      float64 // Lower band (SMA - k*StdDev)
	Width      float64 // Band width percentage
	PercentB   float64 // %B indicator (where price is relative to bands)
	IsSqueeze  bool    // Bands are contracting (low volatility)
	IsBreakout bool    // Price near/outside bands
}

// BollingerBands calculates Bollinger Bands
// Default: period=20, multiplier=2.0
func BollingerBands(prices []float64, period int, multiplier float64) BollingerResult {
	result := BollingerResult{}

	if len(prices) < period {
		return result
	}

	// Calculate SMA (middle band)
	result.Middle = SMA(prices, period)

	// Calculate Standard Deviation
	start := len(prices) - period
	var sumSq float64
	for i := start; i < len(prices); i++ {
		diff := prices[i] - result.Middle
		sumSq += diff * diff
	}
	stdDev := math.Sqrt(sumSq / float64(period))

	// Calculate bands
	result.Upper = result.Middle + (multiplier * stdDev)
	result.Lower = result.Middle - (multiplier * stdDev)

	// Calculate band width (percentage)
	if result.Middle > 0 {
		result.Width = ((result.Upper - result.Lower) / result.Middle) * 100
	}

	// Calculate %B (where current price is relative to bands)
	currentPrice := prices[len(prices)-1]
	bandRange := result.Upper - result.Lower
	if bandRange > 0 {
		result.PercentB = (currentPrice - result.Lower) / bandRange
	}

	// Detect squeeze (bands narrowing - low volatility)
	// Compare current width to average width over last 50 periods
	if len(prices) >= period+50 {
		var widthSum float64
		for i := 0; i < 50; i++ {
			startIdx := len(prices) - period - 50 + i
			endIdx := startIdx + period
			if endIdx <= len(prices) {
				subPrices := prices[startIdx:endIdx]
				subMid := SMA(subPrices, period)
				var subSumSq float64
				for _, p := range subPrices {
					diff := p - subMid
					subSumSq += diff * diff
				}
				subStd := math.Sqrt(subSumSq / float64(period))
				subWidth := ((subMid + multiplier*subStd) - (subMid - multiplier*subStd)) / subMid * 100
				widthSum += subWidth
			}
		}
		avgWidth := widthSum / 50
		result.IsSqueeze = result.Width < avgWidth*0.8 // 20% below average
	}

	// Detect breakout (price near or outside bands)
	if currentPrice >= result.Upper*0.98 || currentPrice <= result.Lower*1.02 {
		result.IsBreakout = true
	}

	return result
}
