package screener

import (
	"math"

	"stockmap/internal/analysis"
	"stockmap/internal/fetcher"
)

// ScreenResult represents a screened stock with all calculated metrics
type ScreenResult struct {
	Symbol        string
	Name          string
	Price         float64
	Change        float64
	ChangePercent float64
	Volume        int64
	MarketCap     int64
	Exchange      string

	// Technical Indicators
	RSI   float64
	ATR   float64
	SMA20 float64
	SMA50 float64

	// Valuation Metrics
	PBV           float64
	PERatio       float64
	EPS           float64
	BookValue     float64
	GrahamNumber  float64
	GrahamUpside  float64
	DividendYield float64

	// Risk Metrics
	StopLoss   float64
	TakeProfit float64
	RiskRatio  float64
	Volatility float64

	// Historical Data (for charts)
	HistoricalPrices []float64

	// Scores
	TechnicalScore  float64
	ValuationScore  float64
	RiskScore       float64
	ConfluenceScore float64

	// Flags
	IsOversold    bool
	IsUndervalued bool
	IsPinned      bool
	HasError      bool
	ErrorMessage  string
}

// CalculateMetrics computes all metrics for a stock
func CalculateMetrics(data *fetcher.StockData) *ScreenResult {
	result := &ScreenResult{
		Symbol:        data.Symbol,
		Name:          data.ShortName,
		Price:         data.Price,
		Change:        data.Change,
		ChangePercent: data.ChangePercent,
		Volume:        data.Volume,
		MarketCap:     data.MarketCap,
		Exchange:      data.Exchange,
		PERatio:       data.PERatio,
		EPS:           data.EPS,
		BookValue:     data.BookValue,
	}

	if data.Error != nil {
		result.HasError = true
		result.ErrorMessage = data.Error.Error()
		return result
	}

	// Calculate RSI
	if len(data.HistoricalPrices) > 14 {
		result.RSI = analysis.RSI(data.HistoricalPrices, 14)
		result.IsOversold = analysis.IsOversold(result.RSI, 35)
	}

	// Calculate ATR
	if len(data.HistoricalHighs) > 14 {
		result.ATR = analysis.ATR(data.HistoricalHighs, data.HistoricalLows, data.HistoricalCloses, 14)
	}

	// Calculate SMAs
	if len(data.HistoricalPrices) >= 20 {
		result.SMA20 = analysis.SMA(data.HistoricalPrices, 20)
	}
	if len(data.HistoricalPrices) >= 50 {
		result.SMA50 = analysis.SMA(data.HistoricalPrices, 50)
	}

	// Calculate PBV
	if data.BookValue > 0 {
		result.PBV = analysis.PBV(data.Price, data.BookValue)
	}

	// Calculate Graham Number
	if data.EPS > 0 && data.BookValue > 0 {
		result.GrahamNumber = analysis.GrahamNumber(data.EPS, data.BookValue)
		result.GrahamUpside = analysis.GrahamUpside(data.Price, result.GrahamNumber)
	}

	// Check if undervalued
	result.IsUndervalued = analysis.IsUndervalued(result.PBV, result.GrahamUpside)

	// Calculate SL/TP
	risk := analysis.CalculateSLTP(data.Price, result.ATR, 2.0, 3.0)
	result.StopLoss = risk.StopLoss
	result.TakeProfit = risk.TakeProfit
	result.RiskRatio = risk.RiskRatio

	// Calculate Volatility
	if len(data.HistoricalPrices) > 10 {
		result.Volatility = analysis.Volatility(data.HistoricalPrices)
	}

	// Store historical prices for chart display
	result.HistoricalPrices = data.HistoricalPrices

	// Calculate Scores
	result.TechnicalScore = calculateTechnicalScore(result)
	result.ValuationScore = analysis.ValuationScore(result.PBV, result.GrahamUpside, result.PERatio)
	result.RiskScore = calculateRiskAdjustedScore(result)
	result.ConfluenceScore = calculateConfluenceScore(result)

	return result
}

// calculateTechnicalScore returns a score based on technical indicators
func calculateTechnicalScore(r *ScreenResult) float64 {
	var score float64

	// RSI scoring (max 50 points) - favor oversold conditions for deep value
	if r.RSI > 0 {
		if r.RSI < 25 {
			score += 50
		} else if r.RSI < 30 {
			score += 45
		} else if r.RSI < 35 {
			score += 40
		} else if r.RSI < 40 {
			score += 30
		} else if r.RSI < 50 {
			score += 20
		} else if r.RSI < 60 {
			score += 10
		}
		// RSI > 60 gets 0 points (overbought territory)
	}

	// Price vs SMA scoring (max 30 points)
	if r.SMA20 > 0 && r.Price < r.SMA20 {
		discount := ((r.SMA20 - r.Price) / r.SMA20) * 100
		if discount > 10 {
			score += 30
		} else if discount > 5 {
			score += 20
		} else {
			score += 10
		}
	}

	// Risk/Reward ratio scoring (max 20 points)
	if r.RiskRatio >= 3.0 {
		score += 20
	} else if r.RiskRatio >= 2.0 {
		score += 15
	} else if r.RiskRatio >= 1.5 {
		score += 10
	} else if r.RiskRatio >= 1.0 {
		score += 5
	}

	return score
}

// calculateRiskAdjustedScore adjusts score based on risk
func calculateRiskAdjustedScore(r *ScreenResult) float64 {
	// Lower volatility = lower risk = better score
	if r.Volatility < 20 {
		return 100
	} else if r.Volatility < 30 {
		return 80
	} else if r.Volatility < 40 {
		return 60
	} else if r.Volatility < 50 {
		return 40
	} else if r.Volatility < 60 {
		return 20
	}
	return 10
}

// calculateConfluenceScore combines all factors into final score
func calculateConfluenceScore(r *ScreenResult) float64 {
	// Weighted average of all scores
	// Technical: 30%, Valuation: 40%, Risk: 30%
	score := (r.TechnicalScore * 0.30) +
		(r.ValuationScore * 0.40) +
		(r.RiskScore * 0.30)

	// Bonus points for confluence signals
	bonusPoints := 0.0

	// Oversold + Undervalued confluence
	if r.IsOversold && r.IsUndervalued {
		bonusPoints += 10
	}

	// Low PBV bonus
	if r.PBV > 0 && r.PBV < 1.0 {
		bonusPoints += 5
	}

	// Strong Graham Upside bonus
	if r.GrahamUpside > 50 {
		bonusPoints += 5
	}

	// Excellent Risk/Reward bonus
	if r.RiskRatio >= 2.5 {
		bonusPoints += 5
	}

	score += bonusPoints

	// Cap at 100
	return math.Min(score, 100)
}

// ScoreToGrade converts numeric score to letter grade
func ScoreToGrade(score float64) string {
	switch {
	case score >= 90:
		return "A+"
	case score >= 85:
		return "A"
	case score >= 80:
		return "A-"
	case score >= 75:
		return "B+"
	case score >= 70:
		return "B"
	case score >= 65:
		return "B-"
	case score >= 60:
		return "C+"
	case score >= 55:
		return "C"
	case score >= 50:
		return "C-"
	case score >= 45:
		return "D"
	default:
		return "F"
	}
}
