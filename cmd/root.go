package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/febritecno/stockmap-cli/internal/fetcher"
	"github.com/febritecno/stockmap-cli/internal/screener"
	"github.com/febritecno/stockmap-cli/internal/ui"
)

var (
	version = "1.0.1"
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "stockmap",
	Short: "A beautiful terminal UI stock screener",
	Long: `StockMap is an interactive terminal-based stock screener
that helps you find undervalued stocks using deep value criteria.

Features:
  • Real-time stock data from Yahoo Finance
  • Technical indicators (RSI, ATR)
  • Valuation metrics (PBV, Graham Number)
  • Dynamic SL/TP based on volatility
  • Confluence scoring system
  • Interactive TUI with keyboard navigation
  • Watchlist management`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := ui.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

// versionCmd prints the version
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("stockmap v%s\n", version)
	},
}

// scanCmd runs a scan without the TUI
var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Run a quick scan and print results",
	Long:  "Run a stock scan and print results to stdout without launching the TUI.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Starting stock scan...")

		symbols := fetcher.DefaultSymbols()
		engine := screener.NewEngine(10)

		// Set progress callback
		engine.SetProgressCallback(func(completed, total int, current string) {
			fmt.Fprintf(os.Stderr, "\rScanning %d/%d: %s        ", completed, total, current)
		})

		results := engine.Scan(symbols)
		fmt.Fprintf(os.Stderr, "\nScan complete. Found %d results.\n\n", len(results))

		// Print results in a table
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "SYMBOL\tPRICE\tRSI\tPBV\tGRAHAM%\tSCORE\tGRADE")

		for _, r := range results {
			// Skip placeholders or errors if any
			if r.HasError || r.Price == 0 {
				continue
			}

			grade := screener.ScoreToGrade(r.ConfluenceScore)

			fmt.Fprintf(w, "%s\t%.2f\t%.2f\t%.2f\t%.1f%%\t%.1f\t%s\n",
				r.Symbol, r.Price, r.RSI, r.PBV, r.GrahamUpside, r.ConfluenceScore, grade)
		}
		w.Flush()
	},
}

// debugCmd runs connection diagnostics
var debugCmd = &cobra.Command{
	Use:   "debug",
	Short: "Run connection diagnostics",
	Long:  "Test connectivity to Yahoo Finance API and report any errors.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Running connection diagnostics...")
		fmt.Println("--------------------------------")

		client := fetcher.NewDirectYahooClient()
		defer client.Close()

		result := client.CheckConnection()

		for _, detail := range result.Details {
			fmt.Println(detail)
		}

		fmt.Println("--------------------------------")
		if result.Connected {
			fmt.Println("Result: SUCCESS - Connection verified")
		} else {
			fmt.Println("Result: FAILED - " + result.Error)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(scanCmd)
	rootCmd.AddCommand(debugCmd)
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
