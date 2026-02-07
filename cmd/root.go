package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"stockmap/internal/ui"
)

var (
	version = "1.0.0"
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
		// For now, just launch the TUI
		// In a future version, this could output JSON or table format
		if err := ui.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(scanCmd)
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
