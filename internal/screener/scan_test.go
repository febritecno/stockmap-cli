package screener

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"

	"github.com/febritecno/stockmap/internal/fetcher"
)

// cleanup removes temporary config files created during tests
func cleanup() {
	cwd, _ := os.Getwd()
	configDir := filepath.Join(cwd, "config")
	os.RemoveAll(configDir)
}

func TestScan_All_Subset(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping network test in short mode")
	}

	defer cleanup()

	// 1. Setup Engine
	// Use small worker count for testing
	engine := NewEngine(3)

	// Clear default watchlist to avoid extra results
	if err := engine.GetWatchlistManager().Clear(); err != nil {
		t.Fatalf("Failed to clear watchlist: %v", err)
	}

	// 2. Prepare "All" scan (subset)
	allSymbols := fetcher.DefaultSymbols()
	if len(allSymbols) == 0 {
		t.Fatal("DefaultSymbols returned empty list")
	}

	// Take a small subset to avoid rate limits and long execution
	subsetSize := 3
	if len(allSymbols) < subsetSize {
		subsetSize = len(allSymbols)
	}
	scanSymbols := allSymbols[:subsetSize]

	// Relax criteria to ensure all scanned symbols are returned
	engine.SetCriteria(FilterCriteria{
		MinRSI:          0,
		MaxRSI:          100,
		MaxPBV:          100,
		MinGrahamUpside: -1000,
		MinConfluence:   0,
	})

	// 3. Run Scan
	// We use a channel to track progress callbacks to verify it runs
	progressCount := 0
	engine.SetProgressCallback(func(completed, total int, current string) {
		progressCount++
	})

	start := time.Now()
	results := engine.Scan(scanSymbols)
	elapsed := time.Since(start)

	// 4. Verify
	t.Logf("Scanned %d symbols in %v", len(scanSymbols), elapsed)
	t.Logf("Results count: %d", len(results))

	if progressCount != len(scanSymbols) {
		t.Errorf("Expected %d progress callbacks, got %d", len(scanSymbols), progressCount)
	}

	// Results might be less than scanSymbols if filtering is active and items are not pinned
	// But we expect the scan to complete without panic
	if len(results) > len(scanSymbols) {
		t.Errorf("Got more results (%d) than scanned symbols (%d)", len(results), len(scanSymbols))
	}
}

func TestScan_Custom(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping network test in short mode")
	}

	defer cleanup()

	// 1. Setup Engine
	engine := NewEngine(3)

	// Clear default watchlist to avoid extra results
	if err := engine.GetWatchlistManager().Clear(); err != nil {
		t.Fatalf("Failed to clear watchlist: %v", err)
	}

	// 2. Define Custom Symbols
	customSymbols := []string{"AAPL", "MSFT"}

	// Set relaxed criteria to ensure results appear if fetched successfully
	engine.SetCriteria(FilterCriteria{
		MinRSI:          0,
		MaxRSI:          100, // Accept all RSI
		MaxPBV:          100, // Accept all PBV
		MinGrahamUpside: -1000,
		MinConfluence:   0,
	})

	// 3. Run Scan
	results := engine.Scan(customSymbols)

	// 4. Verify
	t.Logf("Custom scan results: %d", len(results))

	// Check if results contain requested symbols
	foundCount := 0
	for _, r := range results {
		for _, s := range customSymbols {
			if r.Symbol == s {
				foundCount++
				break
			}
		}
	}

	if len(results) > 0 && foundCount == 0 {
		t.Log("Warning: No matching symbols found in results (network error or symbol not found?)")
	}
}

func TestScan_Watchlist(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping network test in short mode")
	}

	defer cleanup()

	// 1. Setup Engine
	engine := NewEngine(3)
	wm := engine.GetWatchlistManager()

	// 2. Setup Watchlist
	// Ensure clean state
	if err := wm.Clear(); err != nil {
		t.Fatalf("Failed to clear watchlist: %v", err)
	}

	watchlistSymbols := []string{"NVDA", "AMD"}
	// Sort for consistent comparison
	sort.Strings(watchlistSymbols)

	for _, sym := range watchlistSymbols {
		if err := wm.Add(sym); err != nil {
			t.Fatalf("Failed to add to watchlist: %v", err)
		}
	}

	// 3. Get symbols from watchlist (simulating App logic)
	scanSymbols := wm.GetAll()
	sort.Strings(scanSymbols)

	if len(scanSymbols) != len(watchlistSymbols) {
		t.Fatalf("Watchlist count mismatch. Expected %d, got %d", len(watchlistSymbols), len(scanSymbols))
	}

	// 4. Run Scan
	results := engine.Scan(scanSymbols)

	// 5. Verify
	// Pinned items (watchlist items) should ALWAYS be in results because they are pinned.
	if len(results) != len(watchlistSymbols) {
		t.Errorf("Expected %d results (pinned items), got %d", len(watchlistSymbols), len(results))
	}

	for _, r := range results {
		if !r.IsPinned {
			t.Errorf("Result %s should be pinned", r.Symbol)
		}
	}
}
