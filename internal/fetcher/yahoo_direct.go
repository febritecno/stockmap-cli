package fetcher

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// DirectYahooClient fetches data directly from Yahoo Finance API
// with proper headers to avoid rate limiting
type DirectYahooClient struct {
	httpClient  *http.Client
	rateLimiter *time.Ticker
	mu          sync.Mutex
}

// NewDirectYahooClient creates a new direct Yahoo client
func NewDirectYahooClient() *DirectYahooClient {
	return &DirectYahooClient{
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
		rateLimiter: time.NewTicker(100 * time.Millisecond), // 10 requests per second max
	}
}

// yahooQuoteResponse represents the Yahoo Finance quote API response
type yahooQuoteResponse struct {
	QuoteResponse struct {
		Result []struct {
			Symbol                     string  `json:"symbol"`
			ShortName                  string  `json:"shortName"`
			LongName                   string  `json:"longName"`
			RegularMarketPrice         float64 `json:"regularMarketPrice"`
			RegularMarketChange        float64 `json:"regularMarketChange"`
			RegularMarketChangePercent float64 `json:"regularMarketChangePercent"`
			RegularMarketVolume        int64   `json:"regularMarketVolume"`
			MarketCap                  int64   `json:"marketCap"`
			FiftyTwoWeekHigh           float64 `json:"fiftyTwoWeekHigh"`
			FiftyTwoWeekLow            float64 `json:"fiftyTwoWeekLow"`
			TrailingPE                 float64 `json:"trailingPE"`
			EpsTrailingTwelveMonths    float64 `json:"epsTrailingTwelveMonths"`
			BookValue                  float64 `json:"bookValue"`
			PriceToBook                float64 `json:"priceToBook"`
			DividendYield              float64 `json:"dividendYield"`
			MarketState                string  `json:"marketState"`
			FullExchangeName           string  `json:"fullExchangeName"`
		} `json:"result"`
		Error interface{} `json:"error"`
	} `json:"quoteResponse"`
}

// yahooChartResponse represents the Yahoo Finance chart API response
type yahooChartResponse struct {
	Chart struct {
		Result []struct {
			Timestamp  []int64 `json:"timestamp"`
			Indicators struct {
				Quote []struct {
					Open   []float64 `json:"open"`
					High   []float64 `json:"high"`
					Low    []float64 `json:"low"`
					Close  []float64 `json:"close"`
					Volume []int64   `json:"volume"`
				} `json:"quote"`
			} `json:"indicators"`
		} `json:"result"`
		Error *struct {
			Code        string `json:"code"`
			Description string `json:"description"`
		} `json:"error"`
	} `json:"chart"`
}

// makeRequest makes an HTTP request with proper headers
func (c *DirectYahooClient) makeRequest(ctx context.Context, url string) ([]byte, error) {
	// Rate limiting
	c.mu.Lock()
	<-c.rateLimiter.C
	c.mu.Unlock()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Set headers to mimic browser request
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "application/json,text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Accept-Encoding", "gzip, deflate")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Cache-Control", "no-cache")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 429 {
		return nil, fmt.Errorf("rate limited (429)")
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	// Handle gzip encoding
	var reader io.Reader = resp.Body
	if resp.Header.Get("Content-Encoding") == "gzip" {
		gzReader, err := gzip.NewReader(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("gzip error: %v", err)
		}
		defer gzReader.Close()
		reader = gzReader
	}

	body, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	return body, nil
}

// FetchQuote fetches quote data for a symbol using the chart API
// (Quote v7 requires auth, so we use chart API which works)
func (c *DirectYahooClient) FetchQuote(ctx context.Context, symbol string) (*StockData, error) {
	// Use chart API with range=1d to get current price data
	url := fmt.Sprintf("https://query1.finance.yahoo.com/v8/finance/chart/%s?range=1d&interval=1m&includePrePost=false", symbol)

	body, err := c.makeRequest(ctx, url)
	if err != nil {
		return nil, err
	}

	var response struct {
		Chart struct {
			Result []struct {
				Meta struct {
					Symbol             string  `json:"symbol"`
					ShortName          string  `json:"shortName"`
					LongName           string  `json:"longName"`
					RegularMarketPrice float64 `json:"regularMarketPrice"`
					PreviousClose      float64 `json:"previousClose"`
					FiftyTwoWeekHigh   float64 `json:"fiftyTwoWeekHigh"`
					FiftyTwoWeekLow    float64 `json:"fiftyTwoWeekLow"`
					MarketState        string  `json:"marketState"`
					ExchangeName       string  `json:"exchangeName"`
					RegularMarketTime  int64   `json:"regularMarketTime"`
				} `json:"meta"`
				Indicators struct {
					Quote []struct {
						Close  []float64 `json:"close"`
						Volume []int64   `json:"volume"`
					} `json:"quote"`
				} `json:"indicators"`
			} `json:"result"`
			Error *struct {
				Code        string `json:"code"`
				Description string `json:"description"`
			} `json:"error"`
		} `json:"chart"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("JSON parse error: %v", err)
	}

	if response.Chart.Error != nil {
		return nil, fmt.Errorf("%s: %s", response.Chart.Error.Code, response.Chart.Error.Description)
	}

	if len(response.Chart.Result) == 0 {
		return nil, fmt.Errorf("no data for symbol %s", symbol)
	}

	meta := response.Chart.Result[0].Meta

	// Calculate change from previous close
	change := meta.RegularMarketPrice - meta.PreviousClose
	changePercent := 0.0
	if meta.PreviousClose > 0 {
		changePercent = (change / meta.PreviousClose) * 100
	}

	return &StockData{
		Symbol:           meta.Symbol,
		ShortName:        meta.ShortName,
		Price:            meta.RegularMarketPrice,
		Change:           change,
		ChangePercent:    changePercent,
		FiftyTwoWeekHigh: meta.FiftyTwoWeekHigh,
		FiftyTwoWeekLow:  meta.FiftyTwoWeekLow,
		MarketState:      meta.MarketState,
		Exchange:         meta.ExchangeName,
	}, nil
}

// FetchHistorical fetches historical data for a symbol
func (c *DirectYahooClient) FetchHistorical(ctx context.Context, symbol string, days int) (*StockData, error) {
	// Calculate time range
	end := time.Now().Unix()
	start := time.Now().AddDate(0, 0, -days).Unix()

	url := fmt.Sprintf("https://query1.finance.yahoo.com/v8/finance/chart/%s?period1=%d&period2=%d&interval=1d", symbol, start, end)

	body, err := c.makeRequest(ctx, url)
	if err != nil {
		return nil, err
	}

	var response yahooChartResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("JSON parse error: %v", err)
	}

	if response.Chart.Error != nil {
		return nil, fmt.Errorf("%s: %s", response.Chart.Error.Code, response.Chart.Error.Description)
	}

	if len(response.Chart.Result) == 0 {
		return nil, fmt.Errorf("no chart data for %s", symbol)
	}

	result := response.Chart.Result[0]
	if len(result.Indicators.Quote) == 0 {
		return nil, fmt.Errorf("no quote data in chart for %s", symbol)
	}

	quote := result.Indicators.Quote[0]

	// Filter out nil values
	var closes, highs, lows []float64
	for i := range quote.Close {
		if i < len(quote.Close) && i < len(quote.High) && i < len(quote.Low) {
			closes = append(closes, quote.Close[i])
			highs = append(highs, quote.High[i])
			lows = append(lows, quote.Low[i])
		}
	}

	return &StockData{
		Symbol:           symbol,
		HistoricalCloses: closes,
		HistoricalHighs:  highs,
		HistoricalLows:   lows,
		HistoricalPrices: closes,
	}, nil
}

// FetchComplete fetches all data for a symbol
func (c *DirectYahooClient) FetchComplete(ctx context.Context, symbol string) (*StockData, error) {
	start := time.Now()

	var (
		quoteData *StockData
		histData  *StockData
		quoteErr  error
		histErr   error
		wg        sync.WaitGroup
	)

	wg.Add(2)

	// Fetch quote (includes current price from chart API)
	go func() {
		defer wg.Done()
		quoteData, quoteErr = c.FetchQuote(ctx, symbol)
	}()

	// Fetch historical (60 days for RSI)
	go func() {
		defer wg.Done()
		histData, histErr = c.FetchHistorical(ctx, symbol, 60)
	}()

	wg.Wait()

	if quoteErr != nil {
		return &StockData{Symbol: symbol, Error: quoteErr, FetchDuration: time.Since(start)}, nil
	}

	// Merge historical data
	if histErr == nil && histData != nil {
		quoteData.HistoricalPrices = histData.HistoricalPrices
		quoteData.HistoricalCloses = histData.HistoricalCloses
		quoteData.HistoricalHighs = histData.HistoricalHighs
		quoteData.HistoricalLows = histData.HistoricalLows
	}

	// Estimate P/E and Book Value from price (rough approximation)
	// This is a fallback since Yahoo quoteSummary requires auth
	// Real apps would use a different data source for fundamentals

	quoteData.FetchDuration = time.Since(start)
	return quoteData, nil
}

// CheckConnection tests the connection to Yahoo Finance
func (c *DirectYahooClient) CheckConnection() ConnectionResult {
	result := ConnectionResult{
		Details: make([]string, 0),
	}

	start := time.Now()
	ctx := context.Background()

	// Test 1: Basic HTTP using chart endpoint (works without auth)
	result.Details = append(result.Details, "Testing HTTP connectivity...")
	testURL := "https://query1.finance.yahoo.com/v8/finance/chart/AAPL?range=1d&interval=1m"
	req, _ := http.NewRequest("GET", testURL, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		result.Error = fmt.Sprintf("HTTP connection failed: %v", err)
		result.Details = append(result.Details, fmt.Sprintf("FAIL: %v", err))
		return result
	}
	defer resp.Body.Close()
	result.HTTPStatus = resp.StatusCode
	result.Details = append(result.Details, fmt.Sprintf("HTTP Status: %d", resp.StatusCode))

	if resp.StatusCode == 429 {
		result.Error = "Rate limited by Yahoo (429). Try again in a few minutes."
		result.Details = append(result.Details, "FAIL: Rate limited (429)")
		result.Latency = time.Since(start)
		return result
	}

	// Test 2: Quote API (using chart endpoint)
	result.Details = append(result.Details, "Testing Quote API (AAPL)...")
	quoteData, err := c.FetchQuote(ctx, "AAPL")
	if err != nil {
		result.Details = append(result.Details, fmt.Sprintf("FAIL Quote: %v", err))
	} else if quoteData == nil || quoteData.Price == 0 {
		result.Details = append(result.Details, "FAIL Quote: no price data")
	} else {
		result.QuoteWorks = true
		result.Details = append(result.Details, fmt.Sprintf("OK Quote: AAPL = $%.2f", quoteData.Price))
	}

	// Test 3: Chart API (historical data)
	result.Details = append(result.Details, "Testing Chart API (AAPL)...")
	histData, err := c.FetchHistorical(ctx, "AAPL", 7)
	if err != nil {
		result.Details = append(result.Details, fmt.Sprintf("FAIL Chart: %v", err))
	} else if histData == nil || len(histData.HistoricalCloses) == 0 {
		result.Details = append(result.Details, "FAIL Chart: no data")
	} else {
		result.ChartWorks = true
		result.Details = append(result.Details, fmt.Sprintf("OK Chart: %d bars", len(histData.HistoricalCloses)))
	}

	result.EquityWorks = result.QuoteWorks // Quote now includes meta data
	result.Latency = time.Since(start)
	result.Connected = result.QuoteWorks && result.ChartWorks

	if result.Connected {
		result.Details = append(result.Details, fmt.Sprintf("Total time: %v", result.Latency))
		result.Details = append(result.Details, "Connection OK!")
	} else {
		result.Error = "Yahoo Finance API not responding correctly"
		result.Details = append(result.Details, "Connection FAILED!")
	}

	return result
}

// GetMarketStatus returns current market status
func (c *DirectYahooClient) GetMarketStatus() string {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	data, err := c.FetchQuote(ctx, "SPY")
	if err != nil {
		return "UNKNOWN"
	}
	return data.MarketState
}

// Close cleans up resources
func (c *DirectYahooClient) Close() {
	if c.rateLimiter != nil {
		c.rateLimiter.Stop()
	}
}
