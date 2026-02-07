package watchlist

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Watchlist represents the JSON structure
type Watchlist struct {
	Symbols []string `json:"symbols"`
}

// Manager handles watchlist CRUD operations
type Manager struct {
	filePath string
	symbols  map[string]bool
	mu       sync.RWMutex
}

// NewManager creates a new watchlist manager
func NewManager(customPath string) *Manager {
	path := customPath
	if path == "" {
		// Default path in config directory
		execPath, _ := os.Executable()
		execDir := filepath.Dir(execPath)
		path = filepath.Join(execDir, "config", "watchlist.json")

		// Also check current working directory
		if _, err := os.Stat(path); os.IsNotExist(err) {
			cwd, _ := os.Getwd()
			path = filepath.Join(cwd, "config", "watchlist.json")
		}
	}

	m := &Manager{
		filePath: path,
		symbols:  make(map[string]bool),
	}

	m.Load()
	return m
}

// Load reads the watchlist from disk
func (m *Manager) Load() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, err := os.ReadFile(m.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			// Initialize with defaults
			m.symbols = map[string]bool{
				"SLV": true,
				"WDC": true,
				"GDX": true,
			}
			return m.saveUnsafe()
		}
		return err
	}

	var wl Watchlist
	if err := json.Unmarshal(data, &wl); err != nil {
		return err
	}

	m.symbols = make(map[string]bool)
	for _, sym := range wl.Symbols {
		m.symbols[strings.ToUpper(sym)] = true
	}

	return nil
}

// Save writes the watchlist to disk
func (m *Manager) Save() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.saveUnsafe()
}

// saveUnsafe saves without locking (must hold lock)
func (m *Manager) saveUnsafe() error {
	// Ensure directory exists
	dir := filepath.Dir(m.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	symbols := make([]string, 0, len(m.symbols))
	for sym := range m.symbols {
		symbols = append(symbols, sym)
	}

	wl := Watchlist{Symbols: symbols}
	data, err := json.MarshalIndent(wl, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(m.filePath, data, 0644)
}

// Add adds a symbol to the watchlist
func (m *Manager) Add(symbol string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	symbol = strings.ToUpper(symbol)
	m.symbols[symbol] = true

	return m.saveUnsafe()
}

// Remove removes a symbol from the watchlist
func (m *Manager) Remove(symbol string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	symbol = strings.ToUpper(symbol)
	delete(m.symbols, symbol)

	return m.saveUnsafe()
}

// IsPinned checks if a symbol is in the watchlist
func (m *Manager) IsPinned(symbol string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.symbols[strings.ToUpper(symbol)]
}

// GetAll returns all watchlist symbols
func (m *Manager) GetAll() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	symbols := make([]string, 0, len(m.symbols))
	for sym := range m.symbols {
		symbols = append(symbols, sym)
	}
	return symbols
}

// Count returns the number of symbols in watchlist
func (m *Manager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.symbols)
}

// Toggle adds or removes a symbol from watchlist
func (m *Manager) Toggle(symbol string) (added bool, err error) {
	symbol = strings.ToUpper(symbol)

	if m.IsPinned(symbol) {
		err = m.Remove(symbol)
		return false, err
	}

	err = m.Add(symbol)
	return true, err
}

// Clear removes all symbols from watchlist
func (m *Manager) Clear() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.symbols = make(map[string]bool)
	return m.saveUnsafe()
}
