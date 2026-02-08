package fetcher

import (
	"context"
	"testing"
	"time"
)

func TestDirectYahooClient_FetchQuote(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping network test in short mode")
	}

	client := NewDirectYahooClient()
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	data, err := client.FetchQuote(ctx, "AAPL")
	if err != nil {
		t.Fatalf("FetchQuote failed: %v", err)
	}

	if data.Symbol != "AAPL" {
		t.Errorf("Expected symbol AAPL, got %s", data.Symbol)
	}

	if data.Price <= 0 {
		t.Errorf("Expected positive price, got %f", data.Price)
	}

	t.Logf("AAPL: $%.2f (%.2f%%)", data.Price, data.ChangePercent)
}

func TestDirectYahooClient_FetchHistorical(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping network test in short mode")
	}

	client := NewDirectYahooClient()
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	data, err := client.FetchHistorical(ctx, "AAPL", 30)
	if err != nil {
		t.Fatalf("FetchHistorical failed: %v", err)
	}

	if len(data.HistoricalCloses) == 0 {
		t.Error("Expected historical data, got none")
	}

	t.Logf("AAPL: %d historical data points", len(data.HistoricalCloses))
}

func TestDirectYahooClient_FetchComplete(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping network test in short mode")
	}

	client := NewDirectYahooClient()
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	data, err := client.FetchComplete(ctx, "MSFT")
	if err != nil {
		t.Fatalf("FetchComplete failed: %v", err)
	}

	if data.Error != nil {
		t.Fatalf("FetchComplete returned error: %v", data.Error)
	}

	if data.Price <= 0 {
		t.Error("Expected positive price")
	}

	if len(data.HistoricalCloses) == 0 {
		t.Error("Expected historical data")
	}

	t.Logf("MSFT: $%.2f, %d historical points, took %v",
		data.Price, len(data.HistoricalCloses), data.FetchDuration)
}

func TestDirectYahooClient_CheckConnection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping network test in short mode")
	}

	client := NewDirectYahooClient()
	defer client.Close()

	result := client.CheckConnection()

	if !result.Connected {
		t.Errorf("Connection failed: %s", result.Error)
		for _, detail := range result.Details {
			t.Log(detail)
		}
	}

	if !result.QuoteWorks {
		t.Error("Quote API not working")
	}

	if !result.ChartWorks {
		t.Error("Chart API not working")
	}

	t.Logf("Connection OK, latency: %v", result.Latency)
}

func TestWorkerPool_Start(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping network test in short mode")
	}

	pool := NewWorkerPool(3)

	symbols := []string{"AAPL", "MSFT", "GOOGL"}

	resultChan := pool.Start(symbols)

	count := 0
	for data := range resultChan {
		count++
		if data.Error != nil {
			t.Logf("%s: error - %v", data.Symbol, data.Error)
		} else {
			t.Logf("%s: $%.2f", data.Symbol, data.Price)
		}
	}

	if count != len(symbols) {
		t.Errorf("Expected %d results, got %d", len(symbols), count)
	}
}
