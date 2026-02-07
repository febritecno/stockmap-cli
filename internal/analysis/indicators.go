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
