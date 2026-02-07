package screener

import (
	"sort"
	"sync"

	"stockmap/internal/fetcher"
	"stockmap/internal/watchlist"
)

// FilterCriteria defines the screening criteria
type FilterCriteria struct {
	MinRSI          float64
	MaxRSI          float64
	MaxPBV          float64
	MinGrahamUpside float64
	MinConfluence   float64
	OnlyOversold    bool
	OnlyUndervalued bool
}

// DefaultCriteria returns the default deep value criteria
func DefaultCriteria() FilterCriteria {
	return FilterCriteria{
		MinRSI:          0,
		MaxRSI:          40, // Only oversold/neutral
		MaxPBV:          2.0,
		MinGrahamUpside: 0,
		MinConfluence:   50,
		OnlyOversold:    false,
		OnlyUndervalued: false,
	}
}

// ScanProgress contains verbose progress information
type ScanProgress struct {
	Completed    int
	Total        int
	Current      string
	SuccessCount int
	ErrorCount   int
	LastError    string
	ErrorSymbol  string
}

// Engine handles the screening process
type Engine struct {
	pool         *fetcher.WorkerPool
	watchlist    *watchlist.Manager
	criteria     FilterCriteria
	results      []*ScreenResult
	mu           sync.RWMutex
	onProgress   func(completed, total int, current string)
	onProgressV2 func(progress ScanProgress)
	successCount int
	errorCount   int
	lastError    string
	errorSymbol  string
}

// NewEngine creates a new screening engine
func NewEngine(workers int) *Engine {
	return &Engine{
		pool:      fetcher.NewWorkerPool(workers),
		watchlist: watchlist.NewManager(""),
		criteria:  DefaultCriteria(),
	}
}

// SetCriteria updates the filter criteria
func (e *Engine) SetCriteria(c FilterCriteria) {
	e.criteria = c
}

// SetProgressCallback sets the progress callback
func (e *Engine) SetProgressCallback(cb func(completed, total int, current string)) {
	e.onProgress = cb
}

// SetVerboseProgressCallback sets the verbose progress callback
func (e *Engine) SetVerboseProgressCallback(cb func(progress ScanProgress)) {
	e.onProgressV2 = cb
}

// Scan performs the stock screening
func (e *Engine) Scan(symbols []string) []*ScreenResult {
	e.mu.Lock()
	e.results = make([]*ScreenResult, 0)
	e.successCount = 0
	e.errorCount = 0
	e.lastError = ""
	e.errorSymbol = ""
	e.mu.Unlock()

	total := len(symbols)
	completed := 0

	resultChan := e.pool.Start(symbols)

	for data := range resultChan {
		result := CalculateMetrics(data)

		// Track errors for verbose output
		e.mu.Lock()
		if data.Error != nil {
			e.errorCount++
			e.lastError = data.Error.Error()
			e.errorSymbol = data.Symbol
		} else {
			e.successCount++
		}
		localSuccessCount := e.successCount
		localErrorCount := e.errorCount
		localLastError := e.lastError
		localErrorSymbol := e.errorSymbol
		e.mu.Unlock()

		// Mark if pinned in watchlist
		result.IsPinned = e.watchlist.IsPinned(result.Symbol)

		// Only add if passes filter (or is pinned)
		if result.IsPinned || e.passesFilter(result) {
			e.mu.Lock()
			e.results = append(e.results, result)
			e.mu.Unlock()
		}

		completed++
		if e.onProgress != nil {
			e.onProgress(completed, total, data.Symbol)
		}
		if e.onProgressV2 != nil {
			e.onProgressV2(ScanProgress{
				Completed:    completed,
				Total:        total,
				Current:      data.Symbol,
				SuccessCount: localSuccessCount,
				ErrorCount:   localErrorCount,
				LastError:    localLastError,
				ErrorSymbol:  localErrorSymbol,
			})
		}
	}

	// Add placeholders for watchlist symbols that weren't in scan results
	e.addWatchlistPlaceholders()

	// Sort results
	e.sortResults()

	return e.GetResults()
}

// addWatchlistPlaceholders adds placeholder results for watchlist symbols not in results
func (e *Engine) addWatchlistPlaceholders() {
	watchlistSymbols := e.watchlist.GetAll()

	e.mu.Lock()
	defer e.mu.Unlock()

	// Build set of existing symbols
	existing := make(map[string]bool)
	for _, r := range e.results {
		existing[r.Symbol] = true
	}

	// Add placeholders for missing watchlist symbols
	for _, symbol := range watchlistSymbols {
		if !existing[symbol] {
			placeholder := &ScreenResult{
				Symbol:   symbol,
				Name:     symbol,
				IsPinned: true,
			}
			e.results = append(e.results, placeholder)
		}
	}
}

// passesFilter checks if a result meets the filter criteria
func (e *Engine) passesFilter(r *ScreenResult) bool {
	if r.HasError {
		return false
	}

	// RSI filter
	if r.RSI < e.criteria.MinRSI || r.RSI > e.criteria.MaxRSI {
		if r.RSI > 0 { // Only filter if RSI was calculated
			return false
		}
	}

	// PBV filter
	if r.PBV > e.criteria.MaxPBV && r.PBV > 0 {
		return false
	}

	// Graham Upside filter
	if r.GrahamUpside < e.criteria.MinGrahamUpside {
		return false
	}

	// Confluence score filter
	if r.ConfluenceScore < e.criteria.MinConfluence {
		return false
	}

	// Boolean filters
	if e.criteria.OnlyOversold && !r.IsOversold {
		return false
	}

	if e.criteria.OnlyUndervalued && !r.IsUndervalued {
		return false
	}

	return true
}

// sortResults sorts by pinned status first, then by confluence score
func (e *Engine) sortResults() {
	e.mu.Lock()
	defer e.mu.Unlock()

	sort.Slice(e.results, func(i, j int) bool {
		// Pinned items first
		if e.results[i].IsPinned && !e.results[j].IsPinned {
			return true
		}
		if !e.results[i].IsPinned && e.results[j].IsPinned {
			return false
		}
		// Then by confluence score (descending)
		return e.results[i].ConfluenceScore > e.results[j].ConfluenceScore
	})
}

// GetResults returns current results
func (e *Engine) GetResults() []*ScreenResult {
	e.mu.RLock()
	defer e.mu.RUnlock()

	results := make([]*ScreenResult, len(e.results))
	copy(results, e.results)
	return results
}

// GetWatchlistManager returns the watchlist manager
func (e *Engine) GetWatchlistManager() *watchlist.Manager {
	return e.watchlist
}

// Stop cancels any ongoing scan
func (e *Engine) Stop() {
	e.pool.Stop()
}

// RefreshWatchlist reloads the watchlist and updates pin status
func (e *Engine) RefreshWatchlist() {
	e.watchlist.Load()

	e.mu.Lock()
	for _, r := range e.results {
		r.IsPinned = e.watchlist.IsPinned(r.Symbol)
	}
	e.mu.Unlock()

	e.sortResults()
}

// AddToWatchlist adds a symbol to watchlist
func (e *Engine) AddToWatchlist(symbol string) error {
	if err := e.watchlist.Add(symbol); err != nil {
		return err
	}

	// Check if symbol already exists in results
	e.mu.Lock()
	found := false
	for _, r := range e.results {
		if r.Symbol == symbol {
			r.IsPinned = true
			found = true
			break
		}
	}

	// If not found, create placeholder result so it shows in watchlist view
	if !found {
		placeholder := &ScreenResult{
			Symbol:   symbol,
			Name:     symbol, // Will be updated on scan
			IsPinned: true,
		}
		e.results = append(e.results, placeholder)
	}
	e.mu.Unlock()

	e.sortResults()
	return nil
}

// RemoveFromWatchlist removes a symbol from watchlist
func (e *Engine) RemoveFromWatchlist(symbol string) error {
	if err := e.watchlist.Remove(symbol); err != nil {
		return err
	}

	e.mu.Lock()
	// Find the result and update IsPinned or remove if placeholder
	newResults := make([]*ScreenResult, 0, len(e.results))
	for _, r := range e.results {
		if r.Symbol == symbol {
			// Check if it's a placeholder (no price data)
			if r.Price == 0 && r.RSI == 0 {
				// Skip - remove placeholder
				continue
			}
			// Has real data, just unpin
			r.IsPinned = false
		}
		newResults = append(newResults, r)
	}
	e.results = newResults
	e.mu.Unlock()

	e.sortResults()
	return nil
}
