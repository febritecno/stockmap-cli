package analysis

import (
	"math"
)

// RiskReward represents stop-loss and take-profit levels
type RiskReward struct {
	StopLoss      float64
	TakeProfit    float64
	RiskPercent   float64
	RewardPercent float64
	RiskRatio     float64 // Reward:Risk ratio
}

// CalculateSLTP calculates dynamic Stop-Loss and Take-Profit based on ATR
// Uses ATR multipliers for volatility-adjusted levels
func CalculateSLTP(currentPrice, atr float64, slMultiplier, tpMultiplier float64) RiskReward {
	if atr <= 0 {
		// Default to percentage-based if ATR not available
		atr = currentPrice * 0.02 // 2% as default
	}

	sl := currentPrice - (atr * slMultiplier)
	tp := currentPrice + (atr * tpMultiplier)

	riskPercent := ((currentPrice - sl) / currentPrice) * 100
	rewardPercent := ((tp - currentPrice) / currentPrice) * 100

	var riskRatio float64
	if riskPercent > 0 {
		riskRatio = rewardPercent / riskPercent
	}

	return RiskReward{
		StopLoss:      math.Round(sl*100) / 100,
		TakeProfit:    math.Round(tp*100) / 100,
		RiskPercent:   math.Round(riskPercent*100) / 100,
		RewardPercent: math.Round(rewardPercent*100) / 100,
		RiskRatio:     math.Round(riskRatio*100) / 100,
	}
}

// CalculateSupportSL calculates stop-loss based on support levels
func CalculateSupportSL(currentPrice float64, supportLevels []float64, buffer float64) float64 {
	// Find nearest support below current price
	var nearestSupport float64
	for _, support := range supportLevels {
		if support < currentPrice && support > nearestSupport {
			nearestSupport = support
		}
	}

	if nearestSupport > 0 {
		// Add buffer below support
		return nearestSupport * (1 - buffer)
	}

	// Default: 5% below current price
	return currentPrice * 0.95
}

// PositionSize calculates recommended position size based on risk
func PositionSize(accountSize, riskPercent, entryPrice, stopLoss float64) int {
	riskAmount := accountSize * (riskPercent / 100)
	riskPerShare := math.Abs(entryPrice - stopLoss)

	if riskPerShare <= 0 {
		return 0
	}

	shares := riskAmount / riskPerShare
	return int(math.Floor(shares))
}

// MaxDrawdown calculates maximum drawdown from price series
func MaxDrawdown(prices []float64) float64 {
	if len(prices) < 2 {
		return 0.0
	}

	peak := prices[0]
	maxDD := 0.0

	for _, price := range prices {
		if price > peak {
			peak = price
		}
		drawdown := (peak - price) / peak
		if drawdown > maxDD {
			maxDD = drawdown
		}
	}

	return maxDD * 100 // Return as percentage
}

// Volatility calculates annualized volatility from daily returns
func Volatility(prices []float64) float64 {
	if len(prices) < 2 {
		return 0.0
	}

	returns := make([]float64, len(prices)-1)
	for i := 1; i < len(prices); i++ {
		returns[i-1] = (prices[i] - prices[i-1]) / prices[i-1]
	}

	// Calculate mean return
	var sum float64
	for _, r := range returns {
		sum += r
	}
	mean := sum / float64(len(returns))

	// Calculate variance
	var variance float64
	for _, r := range returns {
		variance += (r - mean) * (r - mean)
	}
	variance /= float64(len(returns))

	// Annualize (252 trading days)
	dailyStdDev := math.Sqrt(variance)
	annualizedVol := dailyStdDev * math.Sqrt(252)

	return annualizedVol * 100 // Return as percentage
}

// RiskScore returns a risk score from 0-100 (higher = riskier)
func RiskScore(volatility, maxDrawdown, beta float64) float64 {
	var score float64

	// Volatility component (max 40 points)
	if volatility > 60 {
		score += 40
	} else if volatility > 40 {
		score += 30
	} else if volatility > 25 {
		score += 20
	} else if volatility > 15 {
		score += 10
	}

	// Max Drawdown component (max 40 points)
	if maxDrawdown > 50 {
		score += 40
	} else if maxDrawdown > 30 {
		score += 30
	} else if maxDrawdown > 20 {
		score += 20
	} else if maxDrawdown > 10 {
		score += 10
	}

	// Beta component (max 20 points)
	if beta > 2.0 {
		score += 20
	} else if beta > 1.5 {
		score += 15
	} else if beta > 1.0 {
		score += 10
	} else if beta > 0.5 {
		score += 5
	}

	return score
}
