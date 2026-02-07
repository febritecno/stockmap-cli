package screener

import (
	"testing"
	"time"
)

func TestEngine_Scan(t *testing.T) {
	engine := NewEngine(3)

	symbols := []string{"AAPL", "MSFT"}

	progressCalls := 0
	engine.SetVerboseProgressCallback(func(progress ScanProgress) {
		progressCalls++
		t.Logf("Progress: %d/%d - %s", progress.Completed, progress.Total, progress.Current)
	})

	start := time.Now()
	results := engine.Scan(symbols)
	elapsed := time.Since(start)

	t.Logf("Scan completed in %v", elapsed)
	t.Logf("Results: %d, Progress calls: %d", len(results), progressCalls)

	if progressCalls != len(symbols) {
		t.Errorf("Expected %d progress calls, got %d", len(symbols), progressCalls)
	}

	if elapsed > 30*time.Second {
		t.Errorf("Scan took too long: %v", elapsed)
	}
}

func TestEngine_ScanWithFilter(t *testing.T) {
	engine := NewEngine(3)

	// Set strict criteria
	engine.SetCriteria(FilterCriteria{
		MinRSI:          0,
		MaxRSI:          30, // Very strict
		MaxPBV:          1.0,
		MinGrahamUpside: 50,
		MinConfluence:   80,
	})

	symbols := []string{"AAPL", "MSFT", "GOOGL"}
	results := engine.Scan(symbols)

	t.Logf("With strict filter: %d results", len(results))

	// Most stocks won't pass strict filter
	for _, r := range results {
		t.Logf("  %s: RSI=%.1f, PBV=%.2f, Score=%.0f",
			r.Symbol, r.RSI, r.PBV, r.ConfluenceScore)
	}
}

func TestCalculateMetrics(t *testing.T) {
	// Test with mock data
	data := &mockStockData{
		symbol:    "TEST",
		price:     100.0,
		change:    2.0,
		changePct: 2.0,
		bookValue: 50.0,
		eps:       5.0,
		historicalCloses: []float64{
			90, 92, 91, 93, 94, 95, 96, 97, 98, 99,
			100, 101, 100, 99, 98, 97, 96, 95, 94, 93,
		},
		historicalHighs: []float64{
			91, 93, 92, 94, 95, 96, 97, 98, 99, 100,
			101, 102, 101, 100, 99, 98, 97, 96, 95, 94,
		},
		historicalLows: []float64{
			89, 91, 90, 92, 93, 94, 95, 96, 97, 98,
			99, 100, 99, 98, 97, 96, 95, 94, 93, 92,
		},
	}

	// Can't directly test CalculateMetrics since it expects *fetcher.StockData
	// This is a placeholder for when we add more testable interfaces
	t.Logf("Mock data created for %s", data.symbol)
}

// Mock stock data for testing
type mockStockData struct {
	symbol           string
	price            float64
	change           float64
	changePct        float64
	bookValue        float64
	eps              float64
	historicalCloses []float64
	historicalHighs  []float64
	historicalLows   []float64
}
