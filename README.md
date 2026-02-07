# StockMap: Deep Value Stock Screener for Terminal

<p align="center">
  <img src="docs/screenshot.png" alt="StockMap Screenshot" width="800">
</p>

<p align="center">
  <a href="https://github.com/febritecno/stockmap/blob/main/LICENSE"><img src="https://img.shields.io/badge/License-MIT-blue.svg" alt="License"></a>
  <a href="https://go.dev/"><img src="https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go&logoColor=white" alt="Go Version"></a>
  <a href="https://github.com/charmbracelet/bubbletea"><img src="https://img.shields.io/badge/TUI-Bubble%20Tea-ff69b4" alt="Bubble Tea"></a>
  <a href="#"><img src="https://img.shields.io/badge/Platform-macOS%20|%20Linux%20|%20Windows-lightgrey" alt="Platform"></a>
</p>

<p align="center">
  <b>A beautiful, interactive terminal UI stock screener built with Go and the Charm ecosystem.</b><br>
  Find undervalued stocks using technical analysis, valuation metrics, and confluence scoring.
</p>

---

## Features

| Feature | Description |
|---------|-------------|
| **Real-time Data** | Fetches live stock data from Yahoo Finance API |
| **Technical Analysis** | RSI, ATR, SMA/EMA indicators calculated natively |
| **Valuation Metrics** | P/B Ratio, P/E Ratio, Graham Number, Book Value |
| **Risk Management** | Dynamic Stop-Loss/Take-Profit based on ATR volatility |
| **Confluence Scoring** | Combined weighted score (Technical + Valuation + Risk) |
| **Interactive TUI** | Navigate with arrow keys or vim-style bindings |
| **Watchlist** | Pin favorite stocks with persistent JSON storage |
| **154+ Stocks** | Default scan covers major US equities across all sectors |

---

## Demo

```
┌──────────────────────────────────────────────────────────────────┐
│  ◉ STOCKMAP v1.0         Market: OPEN      Strategy: Deep Value │
├──────────────────────────────────────────────────────────────────┤
│                                                                  │
│  TICKER │ PRICE   │ TP      │ SL      │ RSI  │ PBV  │ SCORE     │
│  ───────┼─────────┼─────────┼─────────┼──────┼──────┼───────────│
│  ★ SLV  │ $22.10  │ $24.50  │ $20.80  │ 31.2 │ 1.1  │ 88 ████▌  │
│  ★ WDC  │ $64.50  │ $72.10  │ $61.20  │ 34.5 │ 1.4  │ 82 ████   │
│    INTC │ $30.15  │ $35.00  │ $28.50  │ 28.9 │ 0.9  │ 95 █████  │
│    VZ   │ $39.80  │ $43.20  │ $38.10  │ 30.1 │ 1.0  │ 78 ███▊   │
│                                                                  │
├──────────────────────────────────────────────────────────────────┤
│  [S]can  [W]atchlist  [F]ilter  [D]etails  [Q]uit               │
│  Scanned: 154 │ Found: 4 │ Last Update: 12:34:56                │
└──────────────────────────────────────────────────────────────────┘
```

---

## Installation

### Quick Install (macOS/Linux)

```bash
curl -sSL https://raw.githubusercontent.com/febritecno/stockmap/main/install.sh | bash
```

### Go Install

```bash
go install github.com/febritecno/stockmap@latest
```

### Build from Source

```bash
# Clone repository
git clone https://github.com/febritecno/stockmap.git
cd stockmap

# Build
go build -o stockmap

# Run
./stockmap
```

### Pre-built Binaries

Download from the [Releases](https://github.com/febritecno/stockmap/releases) page.

---

## Usage

```bash
# Launch interactive TUI
stockmap

# Show version
stockmap version

# Quick scan
stockmap scan
```

---

## Keyboard Shortcuts

### Main Dashboard

| Key | Action |
|-----|--------|
| `S` | Start new scan |
| `W` | Toggle watchlist view |
| `H` | View scan history |
| `D` / `Enter` | View stock details |
| `A` | Add to watchlist |
| `R` | Remove from watchlist |
| `↑` / `k` | Move up |
| `↓` / `j` | Move down |
| `Esc` | Go back |
| `Q` / `Ctrl+C` | Quit |

### History View

| Key | Action |
|-----|--------|
| `Enter` | Load selected scan |
| `X` | Delete selected scan |
| `↑` / `k` | Move up |
| `↓` / `j` | Move down |
| `Esc` | Back to dashboard |

---

## Screening Strategy

StockMap uses a **Deep Value** strategy combining multiple factors:

### Technical Indicators
- **RSI < 40** — Oversold/neutral territory preferred
- **Price < SMA20** — Trading below short-term average
- **Risk:Reward ≥ 1:1.5** — Favorable entry points

### Valuation Metrics
- **P/B Ratio < 2.0** — Trading near or below book value
- **Graham Upside > 0%** — Potential upside to intrinsic value
- **Low P/E Ratio** — Earnings relative to price

### Confluence Score (0-100)

| Component | Weight | Criteria |
|-----------|--------|----------|
| Technical | 30% | RSI, price vs SMA, risk/reward |
| Valuation | 40% | PBV, Graham upside, P/E |
| Risk | 30% | Volatility-adjusted |

**Bonus Points:**
- ✅ Both oversold AND undervalued
- ✅ PBV < 1.0 (below book value)
- ✅ Graham Upside > 50%
- ✅ Risk:Reward ≥ 1:2.5

---

## Configuration

### Watchlist

Default watchlist stored in `config/watchlist.json`:

```json
{
  "symbols": ["SLV", "WDC", "GDX"]
}
```

Edit this file to customize your pinned stocks.

### Scan History

Scan results are automatically saved to `config/history/` as JSON files. Each scan creates a timestamped file:

```
config/history/
├── scan_20240207_143052.json
├── scan_20240207_120000.json
└── scan_20240206_093015.json
```

**Features:**
- Auto-save after each scan completes
- Browse history with `[H]` key
- Load any previous scan with `[Enter]`
- Delete old scans with `[X]`
- Shows timestamp, stocks scanned, and stocks found

### Color Scheme (Tokyo Night)

| Element | Color |
|---------|-------|
| Background | `#1a1b26` |
| Primary | `#7aa2f7` |
| Success | `#9ece6a` |
| Warning | `#e0af68` |
| Danger | `#f7768e` |

---

## Architecture

```
stockmap/
├── cmd/
│   └── root.go                 # Cobra CLI entry
├── internal/
│   ├── analysis/
│   │   ├── indicators.go       # RSI, ATR, SMA, EMA
│   │   ├── valuation.go        # PBV, Graham Number
│   │   └── risk.go             # SL/TP calculations
│   ├── fetcher/
│   │   ├── yahoo.go            # Yahoo Finance client
│   │   └── pool.go             # Worker pool (10 concurrent)
│   ├── history/
│   │   └── history.go          # Scan history management
│   ├── screener/
│   │   ├── engine.go           # Core screening logic
│   │   └── scoring.go          # Confluence score
│   ├── styles/
│   │   └── styles.go           # Lipgloss styling
│   ├── ui/
│   │   ├── app.go              # Main Bubble Tea model
│   │   ├── views/              # Dashboard, Scanner, Details, History
│   │   └── components/         # Table, Header, StatusBar
│   └── watchlist/
│       └── watchlist.go        # JSON CRUD
├── config/
│   ├── history/                # Saved scan results
│   └── watchlist.json          # User watchlist
├── main.go
├── go.mod
└── README.md
```

---

## Requirements

- **Go** v1.21 or higher
- **Terminal** with UTF-8 and True Color support
- **Internet** connection for Yahoo Finance API

---

## Dependencies

| Package | Purpose |
|---------|---------|
| [bubbletea](https://github.com/charmbracelet/bubbletea) | TUI framework |
| [lipgloss](https://github.com/charmbracelet/lipgloss) | Terminal styling |
| [finance-go](https://github.com/piquette/finance-go) | Yahoo Finance API |
| [cobra](https://github.com/spf13/cobra) | CLI framework |

---

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

---

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## Disclaimer

> **⚠️ This tool is for educational and informational purposes only.**
>
> It is not financial advice. The screening criteria and confluence scores are based on quantitative metrics and do not guarantee future performance. Always do your own research before making investment decisions.

---

<p align="center">
  Made with ❤️ using <a href="https://github.com/charmbracelet/bubbletea">Bubble Tea</a>
</p>
