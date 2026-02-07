package analysis

import (
	"math"
)

// PBV calculates Price-to-Book Value ratio
func PBV(price, bookValuePerShare float64) float64 {
	if bookValuePerShare <= 0 {
		return math.Inf(1)
	}
	return price / bookValuePerShare
}

// GrahamNumber calculates Benjamin Graham's intrinsic value formula
// GrahamNumber = sqrt(22.5 * EPS * BVPS)
func GrahamNumber(eps, bookValuePerShare float64) float64 {
	if eps <= 0 || bookValuePerShare <= 0 {
		return 0.0
	}
	return math.Sqrt(22.5 * eps * bookValuePerShare)
}

// GrahamUpside calculates the percentage upside to Graham Number
func GrahamUpside(price, grahamNumber float64) float64 {
	if grahamNumber <= 0 || price <= 0 {
		return 0.0
	}
	return ((grahamNumber - price) / price) * 100.0
}

// PERatio calculates Price-to-Earnings ratio
func PERatio(price, eps float64) float64 {
	if eps <= 0 {
		return math.Inf(1)
	}
	return price / eps
}

// DividendYield calculates dividend yield as percentage
func DividendYield(annualDividend, price float64) float64 {
	if price <= 0 {
		return 0.0
	}
	return (annualDividend / price) * 100.0
}

// IsUndervalued checks if stock is potentially undervalued
// Uses multiple criteria: PBV < 1.5, Graham Upside > 20%, etc.
func IsUndervalued(pbv, grahamUpside float64) bool {
	return pbv < 1.5 && grahamUpside > 20.0
}

// ValuationScore returns a score from 0-100 based on valuation metrics
func ValuationScore(pbv, grahamUpside, peRatio float64) float64 {
	var score float64

	// PBV scoring (max 35 points)
	if pbv < 0.5 {
		score += 35
	} else if pbv < 1.0 {
		score += 30
	} else if pbv < 1.5 {
		score += 25
	} else if pbv < 2.0 {
		score += 15
	} else if pbv < 3.0 {
		score += 5
	}

	// Graham Upside scoring (max 35 points)
	if grahamUpside > 50 {
		score += 35
	} else if grahamUpside > 30 {
		score += 30
	} else if grahamUpside > 20 {
		score += 25
	} else if grahamUpside > 10 {
		score += 15
	} else if grahamUpside > 0 {
		score += 5
	}

	// PE Ratio scoring (max 30 points)
	if !math.IsInf(peRatio, 1) {
		if peRatio < 10 {
			score += 30
		} else if peRatio < 15 {
			score += 25
		} else if peRatio < 20 {
			score += 20
		} else if peRatio < 25 {
			score += 10
		}
	}

	return score
}
