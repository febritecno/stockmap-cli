package fetcher

// Category represents a group of stock symbols
type Category struct {
	Name    string
	Symbols []string
}

// DefaultCategories returns the list of default categories and their symbols
func DefaultCategories() []Category {
	return []Category{
		{
			Name: "Technology",
			Symbols: []string{
				"AAPL", "MSFT", "GOOGL", "AMZN", "META", "NVDA", "AMD", "INTC", "CRM", "ORCL",
				"CSCO", "IBM", "QCOM", "TXN", "AVGO", "MU", "AMAT", "LRCX", "KLAC", "SNPS",
			},
		},
		{
			Name: "Finance",
			Symbols: []string{
				"JPM", "BAC", "WFC", "C", "GS", "MS", "BRK-B", "V", "MA", "AXP",
				"SCHW", "BLK", "SPGI", "MCO", "ICE", "CME", "AON", "MMC", "TRV", "MET",
			},
		},
		{
			Name: "Healthcare",
			Symbols: []string{
				"JNJ", "UNH", "PFE", "MRK", "ABBV", "LLY", "BMY", "AMGN", "GILD", "CVS",
			},
		},
		{
			Name: "Biotech",
			Symbols: []string{
				"REGN", "VRTX", "MRNA", "BIIB", "ILMN", "INCY",
			},
		},
		{
			Name: "Consumer",
			Symbols: []string{
				"WMT", "PG", "KO", "PEP", "COST", "HD", "MCD", "NKE", "SBUX", "TGT",
				"LOW", "TJX", "ROST", "DG", "DLTR", "YUM", "CMG", "DPZ", "DKNG",
			},
		},
		{
			Name: "Energy",
			Symbols: []string{
				"XOM", "CVX", "COP", "SLB", "EOG", "MPC", "VLO", "PSX", "OXY", "HAL",
			},
		},
		{
			Name: "Industrial",
			Symbols: []string{
				"BA", "CAT", "GE", "MMM", "HON", "UPS", "RTX", "LMT", "DE", "UNP",
				"FDX", "NSC", "CSX", "WM", "RSG", "GD", "NOC", "TDG", "ITW", "EMR",
			},
		},
		{
			Name: "Materials",
			Symbols: []string{
				"LIN", "APD", "SHW", "ECL", "FCX", "NEM", "NUE", "DOW", "DD", "PPG",
			},
		},
		{
			Name: "Telecom",
			Symbols: []string{
				"VZ", "T", "TMUS", "CMCSA", "DIS", "NFLX", "CHTR",
			},
		},
		{
			Name: "Utilities",
			Symbols: []string{
				"NEE", "DUK", "SO", "D", "AEP", "EXC", "SRE", "XEL", "WEC", "ES",
			},
		},
		{
			Name: "Real Estate",
			Symbols: []string{
				"PLD", "AMT", "CCI", "EQIX", "PSA", "SPG", "O", "WELL", "DLR", "AVB",
			},
		},
		{
			Name: "ETFs (Metals)",
			Symbols: []string{
				"GLD", "SLV", "GDX", "GDXJ", "IAU",
			},
		},
		{
			Name: "Value Picks",
			Symbols: []string{
				"WDC", "PARA", "WBA", "VFC", "LUMN", "AAL", "UAL", "DAL", "F", "GM",
			},
		},
	}
}

// DefaultSymbols returns a flattened list of all default stock symbols
func DefaultSymbols() []string {
	var symbols []string
	for _, cat := range DefaultCategories() {
		symbols = append(symbols, cat.Symbols...)
	}
	return symbols
}
