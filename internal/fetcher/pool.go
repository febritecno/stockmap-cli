package fetcher

import (
	"context"
	"sync"
)

// WorkerPool manages concurrent fetching of stock data
type WorkerPool struct {
	workers    int
	client     *DirectYahooClient
	symbols    chan string
	results    chan *StockData
	wg         sync.WaitGroup
	ctx        context.Context
	cancel     context.CancelFunc
	onProgress func(completed, total int)
}

// NewWorkerPool creates a new worker pool
func NewWorkerPool(workers int) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())
	return &WorkerPool{
		workers: workers,
		client:  NewDirectYahooClient(),
		ctx:     ctx,
		cancel:  cancel,
	}
}

// SetProgressCallback sets a callback for progress updates
func (p *WorkerPool) SetProgressCallback(cb func(completed, total int)) {
	p.onProgress = cb
}

// Start begins processing symbols
func (p *WorkerPool) Start(symbols []string) <-chan *StockData {
	// Cancel any previous context and create new one
	if p.cancel != nil {
		p.cancel()
	}
	p.ctx, p.cancel = context.WithCancel(context.Background())

	// Reset WaitGroup
	p.wg = sync.WaitGroup{}

	p.symbols = make(chan string, len(symbols))
	p.results = make(chan *StockData, len(symbols))

	// Start workers
	for i := 0; i < p.workers; i++ {
		p.wg.Add(1)
		go p.worker()
	}

	// Feed symbols to workers
	go func() {
		for _, sym := range symbols {
			select {
			case <-p.ctx.Done():
				break
			case p.symbols <- sym:
			}
		}
		close(p.symbols)
	}()

	// Close results when all workers are done
	go func() {
		p.wg.Wait()
		close(p.results)
	}()

	return p.results
}

// worker processes symbols from the channel
func (p *WorkerPool) worker() {
	defer p.wg.Done()

	for {
		select {
		case <-p.ctx.Done():
			return
		case symbol, ok := <-p.symbols:
			if !ok {
				return
			}

			data, err := p.client.FetchComplete(p.ctx, symbol)
			if err != nil {
				data = &StockData{Symbol: symbol, Error: err}
			}

			select {
			case <-p.ctx.Done():
				return
			case p.results <- data:
			}
		}
	}
}

// Stop cancels all pending work
func (p *WorkerPool) Stop() {
	if p.cancel != nil {
		p.cancel()
	}
}

// FetchAll fetches data for all symbols and returns when complete
func (p *WorkerPool) FetchAll(symbols []string) []*StockData {
	results := make([]*StockData, 0, len(symbols))
	completed := 0
	total := len(symbols)

	resultChan := p.Start(symbols)
	for data := range resultChan {
		results = append(results, data)
		completed++
		if p.onProgress != nil {
			p.onProgress(completed, total)
		}
	}

	return results
}

// DefaultSymbols returns a list of popular stock symbols to scan
func DefaultSymbols() []string {
	return []string{
		// Tech
		"AAPL", "MSFT", "GOOGL", "AMZN", "META", "NVDA", "AMD", "INTC", "CRM", "ORCL",
		// Finance
		"JPM", "BAC", "WFC", "C", "GS", "MS", "BRK-B", "V", "MA", "AXP",
		// Healthcare
		"JNJ", "UNH", "PFE", "MRK", "ABBV", "LLY", "BMY", "AMGN", "GILD", "CVS",
		// Consumer
		"WMT", "PG", "KO", "PEP", "COST", "HD", "MCD", "NKE", "SBUX", "TGT",
		// Energy
		"XOM", "CVX", "COP", "SLB", "EOG", "MPC", "VLO", "PSX", "OXY", "HAL",
		// Industrial
		"BA", "CAT", "GE", "MMM", "HON", "UPS", "RTX", "LMT", "DE", "UNP",
		// Materials
		"LIN", "APD", "SHW", "ECL", "FCX", "NEM", "NUE", "DOW", "DD", "PPG",
		// Telecom
		"VZ", "T", "TMUS", "CMCSA", "DIS", "NFLX", "CHTR",
		// Utilities
		"NEE", "DUK", "SO", "D", "AEP", "EXC", "SRE", "XEL", "WEC", "ES",
		// REITs
		"PLD", "AMT", "CCI", "EQIX", "PSA", "SPG", "O", "WELL", "DLR", "AVB",
		// ETFs for metals
		"GLD", "SLV", "GDX", "GDXJ", "IAU",
		// Other value candidates
		"WDC", "PARA", "WBA", "VFC", "LUMN", "AAL", "UAL", "DAL", "F", "GM",
		// Additional Tech
		"CSCO", "IBM", "QCOM", "TXN", "AVGO", "MU", "AMAT", "LRCX", "KLAC", "SNPS",
		// Additional Consumer
		"LOW", "TJX", "ROST", "DG", "DLTR", "YUM", "CMG", "DPZ", "DKNG",
		// Additional Finance
		"SCHW", "BLK", "SPGI", "MCO", "ICE", "CME", "AON", "MMC", "TRV", "MET",
		// Additional Industrial
		"FDX", "NSC", "CSX", "WM", "RSG", "GD", "NOC", "TDG", "ITW", "EMR",
		// Biotech
		"REGN", "VRTX", "MRNA", "BIIB", "ILMN", "SGEN", "ALXN", "INCY",
	}
}
