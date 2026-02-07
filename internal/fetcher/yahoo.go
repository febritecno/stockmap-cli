package fetcher

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/piquette/finance-go/chart"
	"github.com/piquette/finance-go/datetime"
	"github.com/piquette/finance-go/equity"
	"github.com/piquette/finance-go/quote"
)

// StockData contains all fetched data for a stock
type StockData struct {
	Symbol           string
	Price            float64
	Change           float64
	ChangePercent    float64
	Volume           int64
	MarketCap        int64
	PERatio          float64
	EPS              float64
	BookValue        float64
	DividendYield    float64
	FiftyTwoWeekHigh float64
	FiftyTwoWeekLow  float64
	HistoricalPrices []float64
	HistoricalHighs  []float64
	HistoricalLows   []float64
	HistoricalCloses []float64
	ShortName        string
	Exchange         string
	MarketState      string
	Error            error
	FetchDuration    time.Duration
}

// YahooClient wraps the finance-go library
type YahooClient struct {
	timeout time.Duration
}

// NewYahooClient creates a new Yahoo Finance client
func NewYahooClient() *YahooClient {
	return &YahooClient{
		timeout: 5 * time.Second, // Fast timeout
	}
}

// FetchQuote fetches real-time quote data for a symbol
func (c *YahooClient) FetchQuote(symbol string) (*StockData, error) {
	q, err := quote.Get(symbol)
	if err != nil {
		return nil, err
	}

	data := &StockData{
		Symbol:           q.Symbol,
		Price:            q.RegularMarketPrice,
		Change:           q.RegularMarketChange,
		ChangePercent:    q.RegularMarketChangePercent,
		Volume:           int64(q.RegularMarketVolume),
		FiftyTwoWeekHigh: q.FiftyTwoWeekHigh,
		FiftyTwoWeekLow:  q.FiftyTwoWeekLow,
		ShortName:        q.ShortName,
		Exchange:         q.FullExchangeName,
		MarketState:      string(q.MarketState),
	}

	return data, nil
}

// FetchEquity fetches fundamental data for a symbol
func (c *YahooClient) FetchEquity(symbol string) (*StockData, error) {
	eq, err := equity.Get(symbol)
	if err != nil {
		return nil, err
	}

	data := &StockData{
		Symbol:    symbol,
		PERatio:   eq.TrailingPE,
		EPS:       eq.EpsTrailingTwelveMonths,
		BookValue: eq.BookValue,
	}

	return data, nil
}

// FetchHistorical fetches historical price data
func (c *YahooClient) FetchHistorical(symbol string, period int) (*StockData, error) {
	// Calculate start and end dates
	end := time.Now()
	start := end.AddDate(0, 0, -period)

	params := &chart.Params{
		Symbol:   symbol,
		Start:    datetime.New(&start),
		End:      datetime.New(&end),
		Interval: datetime.OneDay,
	}

	iter := chart.Get(params)

	var prices, highs, lows, closes []float64
	for iter.Next() {
		bar := iter.Bar()
		closePrice, _ := bar.Close.Float64()
		highPrice, _ := bar.High.Float64()
		lowPrice, _ := bar.Low.Float64()

		prices = append(prices, closePrice)
		highs = append(highs, highPrice)
		lows = append(lows, lowPrice)
		closes = append(closes, closePrice)
	}

	if err := iter.Err(); err != nil {
		return nil, err
	}

	return &StockData{
		Symbol:           symbol,
		HistoricalPrices: prices,
		HistoricalHighs:  highs,
		HistoricalLows:   lows,
		HistoricalCloses: closes,
	}, nil
}

// FetchComplete fetches all available data for a symbol using parallel requests
func (c *YahooClient) FetchComplete(ctx context.Context, symbol string) (*StockData, error) {
	start := time.Now()

	var (
		quoteData  *StockData
		equityData *StockData
		histData   *StockData
		quoteErr   error
		equityErr  error
		histErr    error
		wg         sync.WaitGroup
	)

	// Fetch all data in parallel for speed
	wg.Add(3)

	// Fetch quote data
	go func() {
		defer wg.Done()
		quoteData, quoteErr = c.FetchQuote(symbol)
	}()

	// Fetch equity fundamentals
	go func() {
		defer wg.Done()
		equityData, equityErr = c.FetchEquity(symbol)
	}()

	// Fetch historical data (60 days for RSI/ATR calculation)
	go func() {
		defer wg.Done()
		histData, histErr = c.FetchHistorical(symbol, 60)
	}()

	wg.Wait()

	// If quote fails, return error (quote is essential)
	if quoteErr != nil {
		return &StockData{Symbol: symbol, Error: quoteErr, FetchDuration: time.Since(start)}, nil
	}

	// Merge equity data if available
	if equityErr == nil && equityData != nil {
		quoteData.PERatio = equityData.PERatio
		quoteData.EPS = equityData.EPS
		quoteData.BookValue = equityData.BookValue
	}

	// Merge historical data if available
	if histErr == nil && histData != nil {
		quoteData.HistoricalPrices = histData.HistoricalPrices
		quoteData.HistoricalHighs = histData.HistoricalHighs
		quoteData.HistoricalLows = histData.HistoricalLows
		quoteData.HistoricalCloses = histData.HistoricalCloses
	}

	quoteData.FetchDuration = time.Since(start)
	return quoteData, nil
}

// GetMarketStatus returns current market status
func (c *YahooClient) GetMarketStatus() string {
	// Check SPY as proxy for US market
	q, err := quote.Get("SPY")
	if err != nil {
		return "UNKNOWN"
	}
	return string(q.MarketState)
}

// ConnectionResult contains connection test results
type ConnectionResult struct {
	Connected   bool
	Latency     time.Duration
	HTTPStatus  int
	QuoteWorks  bool
	EquityWorks bool
	ChartWorks  bool
	Error       string
	Details     []string
}

// CheckConnection tests the connection to Yahoo Finance API
func (c *YahooClient) CheckConnection() ConnectionResult {
	result := ConnectionResult{
		Details: make([]string, 0),
	}

	start := time.Now()

	// Step 1: Basic HTTP connectivity check
	result.Details = append(result.Details, "Testing HTTP connectivity...")
	httpClient := &http.Client{Timeout: 10 * time.Second}
	resp, err := httpClient.Get("https://query1.finance.yahoo.com/v1/test/getcrumb")
	if err != nil {
		result.Error = fmt.Sprintf("HTTP connection failed: %v", err)
		result.Details = append(result.Details, fmt.Sprintf("FAIL: %v", err))
		return result
	}
	defer resp.Body.Close()
	result.HTTPStatus = resp.StatusCode
	result.Details = append(result.Details, fmt.Sprintf("HTTP Status: %d", resp.StatusCode))

	// Step 2: Test Quote API
	result.Details = append(result.Details, "Testing Quote API (AAPL)...")
	q, err := quote.Get("AAPL")
	if err != nil {
		result.Details = append(result.Details, fmt.Sprintf("FAIL Quote: %v", err))
	} else if q == nil {
		result.Details = append(result.Details, "FAIL Quote: nil response")
	} else if q.RegularMarketPrice == 0 {
		result.Details = append(result.Details, "FAIL Quote: price is 0")
	} else {
		result.QuoteWorks = true
		result.Details = append(result.Details, fmt.Sprintf("OK Quote: AAPL = $%.2f", q.RegularMarketPrice))
	}

	// Step 3: Test Equity API
	result.Details = append(result.Details, "Testing Equity API (AAPL)...")
	eq, err := equity.Get("AAPL")
	if err != nil {
		result.Details = append(result.Details, fmt.Sprintf("FAIL Equity: %v", err))
	} else if eq == nil {
		result.Details = append(result.Details, "FAIL Equity: nil response")
	} else {
		result.EquityWorks = true
		result.Details = append(result.Details, fmt.Sprintf("OK Equity: P/E = %.2f", eq.TrailingPE))
	}

	// Step 4: Test Chart API
	result.Details = append(result.Details, "Testing Chart API (AAPL)...")
	end := time.Now()
	chartStart := end.AddDate(0, 0, -7)
	params := &chart.Params{
		Symbol:   "AAPL",
		Start:    datetime.New(&chartStart),
		End:      datetime.New(&end),
		Interval: datetime.OneDay,
	}
	iter := chart.Get(params)
	count := 0
	for iter.Next() {
		count++
	}
	if err := iter.Err(); err != nil {
		result.Details = append(result.Details, fmt.Sprintf("FAIL Chart: %v", err))
	} else if count == 0 {
		result.Details = append(result.Details, "FAIL Chart: no data returned")
	} else {
		result.ChartWorks = true
		result.Details = append(result.Details, fmt.Sprintf("OK Chart: %d bars returned", count))
	}

	result.Latency = time.Since(start)
	result.Connected = result.QuoteWorks // At minimum quote must work
	result.Details = append(result.Details, fmt.Sprintf("Total time: %v", result.Latency))

	if result.Connected {
		result.Details = append(result.Details, "Connection OK!")
	} else {
		result.Error = "Yahoo Finance API not responding correctly"
		result.Details = append(result.Details, "Connection FAILED!")
	}

	return result
}
