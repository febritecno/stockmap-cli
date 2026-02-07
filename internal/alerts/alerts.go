package alerts

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// AlertType represents the type of price alert
type AlertType string

const (
	AlertAbove   AlertType = "above"    // Alert when price goes above threshold
	AlertBelow   AlertType = "below"    // Alert when price goes below threshold
	AlertCross   AlertType = "cross"    // Alert when price crosses threshold (either direction)
	AlertChange  AlertType = "change"   // Alert on % change
	AlertRSILow  AlertType = "rsi_low"  // Alert when RSI goes below threshold
	AlertRSIHigh AlertType = "rsi_high" // Alert when RSI goes above threshold
)

// Alert represents a single price alert
type Alert struct {
	ID          string    `json:"id"`
	Symbol      string    `json:"symbol"`
	Type        AlertType `json:"type"`
	Threshold   float64   `json:"threshold"` // Price or RSI threshold
	CreatedAt   time.Time `json:"created_at"`
	TriggeredAt time.Time `json:"triggered_at,omitempty"`
	IsTriggered bool      `json:"is_triggered"`
	IsActive    bool      `json:"is_active"`
	Message     string    `json:"message,omitempty"`
	LastPrice   float64   `json:"last_price,omitempty"` // Last known price when created
}

// TriggeredAlert represents an alert that has been triggered
type TriggeredAlert struct {
	Alert        Alert
	CurrentPrice float64
	CurrentRSI   float64
	Timestamp    time.Time
}

// AlertsFile represents the JSON structure for storing alerts
type AlertsFile struct {
	Alerts []Alert `json:"alerts"`
}

// Manager handles alert CRUD operations and checking
type Manager struct {
	filePath        string
	alerts          map[string]*Alert   // key is alert ID
	symbolAlerts    map[string][]*Alert // alerts grouped by symbol
	triggeredAlerts []TriggeredAlert
	mu              sync.RWMutex
	onTrigger       func(TriggeredAlert) // callback when alert triggers
}

// NewManager creates a new alerts manager
func NewManager(customPath string) *Manager {
	path := customPath
	if path == "" {
		execPath, _ := os.Executable()
		execDir := filepath.Dir(execPath)
		path = filepath.Join(execDir, "config", "alerts.json")

		if _, err := os.Stat(path); os.IsNotExist(err) {
			cwd, _ := os.Getwd()
			path = filepath.Join(cwd, "config", "alerts.json")
		}
	}

	m := &Manager{
		filePath:     path,
		alerts:       make(map[string]*Alert),
		symbolAlerts: make(map[string][]*Alert),
	}

	m.Load()
	return m
}

// SetOnTrigger sets the callback for when an alert is triggered
func (m *Manager) SetOnTrigger(fn func(TriggeredAlert)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.onTrigger = fn
}

// Load reads alerts from disk
func (m *Manager) Load() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, err := os.ReadFile(m.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No alerts file yet
		}
		return err
	}

	var af AlertsFile
	if err := json.Unmarshal(data, &af); err != nil {
		return err
	}

	m.alerts = make(map[string]*Alert)
	m.symbolAlerts = make(map[string][]*Alert)

	for i := range af.Alerts {
		alert := &af.Alerts[i]
		m.alerts[alert.ID] = alert
		symbol := strings.ToUpper(alert.Symbol)
		m.symbolAlerts[symbol] = append(m.symbolAlerts[symbol], alert)
	}

	return nil
}

// Save writes alerts to disk
func (m *Manager) Save() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.saveUnsafe()
}

func (m *Manager) saveUnsafe() error {
	dir := filepath.Dir(m.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	alerts := make([]Alert, 0, len(m.alerts))
	for _, alert := range m.alerts {
		alerts = append(alerts, *alert)
	}

	af := AlertsFile{Alerts: alerts}
	data, err := json.MarshalIndent(af, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(m.filePath, data, 0644)
}

// Add creates a new alert
func (m *Manager) Add(symbol string, alertType AlertType, threshold float64, lastPrice float64) (*Alert, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	id := generateID()
	symbol = strings.ToUpper(symbol)

	alert := &Alert{
		ID:        id,
		Symbol:    symbol,
		Type:      alertType,
		Threshold: threshold,
		CreatedAt: time.Now(),
		IsActive:  true,
		LastPrice: lastPrice,
	}

	m.alerts[id] = alert
	m.symbolAlerts[symbol] = append(m.symbolAlerts[symbol], alert)

	if err := m.saveUnsafe(); err != nil {
		return nil, err
	}

	return alert, nil
}

// Remove deletes an alert
func (m *Manager) Remove(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	alert, exists := m.alerts[id]
	if !exists {
		return nil
	}

	delete(m.alerts, id)

	// Remove from symbol alerts
	symbol := alert.Symbol
	filtered := make([]*Alert, 0)
	for _, a := range m.symbolAlerts[symbol] {
		if a.ID != id {
			filtered = append(filtered, a)
		}
	}
	m.symbolAlerts[symbol] = filtered

	return m.saveUnsafe()
}

// RemoveBySymbol removes all alerts for a symbol
func (m *Manager) RemoveBySymbol(symbol string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	symbol = strings.ToUpper(symbol)
	alerts := m.symbolAlerts[symbol]

	for _, alert := range alerts {
		delete(m.alerts, alert.ID)
	}
	delete(m.symbolAlerts, symbol)

	return m.saveUnsafe()
}

// GetAll returns all alerts
func (m *Manager) GetAll() []*Alert {
	m.mu.RLock()
	defer m.mu.RUnlock()

	alerts := make([]*Alert, 0, len(m.alerts))
	for _, alert := range m.alerts {
		alerts = append(alerts, alert)
	}
	return alerts
}

// GetBySymbol returns all alerts for a symbol
func (m *Manager) GetBySymbol(symbol string) []*Alert {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.symbolAlerts[strings.ToUpper(symbol)]
}

// GetActiveCount returns the number of active alerts
func (m *Manager) GetActiveCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	count := 0
	for _, alert := range m.alerts {
		if alert.IsActive && !alert.IsTriggered {
			count++
		}
	}
	return count
}

// CheckPrice checks if any alerts should be triggered for a symbol
func (m *Manager) CheckPrice(symbol string, currentPrice, currentRSI float64) []TriggeredAlert {
	m.mu.Lock()
	defer m.mu.Unlock()

	symbol = strings.ToUpper(symbol)
	alerts := m.symbolAlerts[symbol]
	triggered := []TriggeredAlert{}

	for _, alert := range alerts {
		if !alert.IsActive || alert.IsTriggered {
			continue
		}

		shouldTrigger := false

		switch alert.Type {
		case AlertAbove:
			if currentPrice >= alert.Threshold {
				shouldTrigger = true
				alert.Message = "Price crossed above threshold"
			}
		case AlertBelow:
			if currentPrice <= alert.Threshold {
				shouldTrigger = true
				alert.Message = "Price crossed below threshold"
			}
		case AlertCross:
			// Triggered if price crosses threshold in either direction
			if (alert.LastPrice < alert.Threshold && currentPrice >= alert.Threshold) ||
				(alert.LastPrice > alert.Threshold && currentPrice <= alert.Threshold) {
				shouldTrigger = true
				alert.Message = "Price crossed threshold"
			}
		case AlertChange:
			// Threshold is percentage change
			if alert.LastPrice > 0 {
				pctChange := ((currentPrice - alert.LastPrice) / alert.LastPrice) * 100
				if pctChange >= alert.Threshold || pctChange <= -alert.Threshold {
					shouldTrigger = true
					alert.Message = "Price changed significantly"
				}
			}
		case AlertRSILow:
			if currentRSI > 0 && currentRSI <= alert.Threshold {
				shouldTrigger = true
				alert.Message = "RSI dropped below threshold"
			}
		case AlertRSIHigh:
			if currentRSI > 0 && currentRSI >= alert.Threshold {
				shouldTrigger = true
				alert.Message = "RSI rose above threshold"
			}
		}

		if shouldTrigger {
			alert.IsTriggered = true
			alert.TriggeredAt = time.Now()

			ta := TriggeredAlert{
				Alert:        *alert,
				CurrentPrice: currentPrice,
				CurrentRSI:   currentRSI,
				Timestamp:    time.Now(),
			}
			triggered = append(triggered, ta)
			m.triggeredAlerts = append(m.triggeredAlerts, ta)

			if m.onTrigger != nil {
				go m.onTrigger(ta)
			}
		}
	}

	if len(triggered) > 0 {
		m.saveUnsafe()
	}

	return triggered
}

// GetTriggeredAlerts returns all recently triggered alerts
func (m *Manager) GetTriggeredAlerts() []TriggeredAlert {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.triggeredAlerts
}

// ClearTriggeredAlerts clears the triggered alerts buffer
func (m *Manager) ClearTriggeredAlerts() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.triggeredAlerts = nil
}

// ResetAlert resets a triggered alert so it can trigger again
func (m *Manager) ResetAlert(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	alert, exists := m.alerts[id]
	if !exists {
		return nil
	}

	alert.IsTriggered = false
	alert.TriggeredAt = time.Time{}
	alert.Message = ""

	return m.saveUnsafe()
}

// ToggleActive toggles the active state of an alert
func (m *Manager) ToggleActive(id string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	alert, exists := m.alerts[id]
	if !exists {
		return false, nil
	}

	alert.IsActive = !alert.IsActive
	return alert.IsActive, m.saveUnsafe()
}

// generateID creates a simple unique ID
func generateID() string {
	return time.Now().Format("20060102150405.000")
}

// FormatAlertType returns a human-readable alert type
func FormatAlertType(t AlertType) string {
	switch t {
	case AlertAbove:
		return "Price Above"
	case AlertBelow:
		return "Price Below"
	case AlertCross:
		return "Price Cross"
	case AlertChange:
		return "% Change"
	case AlertRSILow:
		return "RSI Low"
	case AlertRSIHigh:
		return "RSI High"
	default:
		return string(t)
	}
}
