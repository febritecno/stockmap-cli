package history

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/febritecno/stockmap/internal/screener"
)

// ScanRecord represents a saved scan result
type ScanRecord struct {
	ID           string                   `json:"id"`
	Timestamp    time.Time                `json:"timestamp"`
	TotalScanned int                      `json:"total_scanned"`
	TotalFound   int                      `json:"total_found"`
	Results      []*screener.ScreenResult `json:"results"`
}

// Manager handles history CRUD operations
type Manager struct {
	historyDir string
}

// NewManager creates a new history manager
func NewManager() *Manager {
	// Default path in config/history directory
	execPath, _ := os.Executable()
	execDir := filepath.Dir(execPath)
	historyDir := filepath.Join(execDir, "config", "history")

	// Also check current working directory
	if _, err := os.Stat(historyDir); os.IsNotExist(err) {
		cwd, _ := os.Getwd()
		historyDir = filepath.Join(cwd, "config", "history")
	}

	// Create directory if not exists
	os.MkdirAll(historyDir, 0755)

	return &Manager{
		historyDir: historyDir,
	}
}

// Save saves scan results to history
func (m *Manager) Save(results []*screener.ScreenResult, totalScanned int) (*ScanRecord, error) {
	now := time.Now()
	id := now.Format("20060102_150405")

	record := &ScanRecord{
		ID:           id,
		Timestamp:    now,
		TotalScanned: totalScanned,
		TotalFound:   len(results),
		Results:      results,
	}

	filename := filepath.Join(m.historyDir, fmt.Sprintf("scan_%s.json", id))

	data, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return nil, err
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return nil, err
	}

	return record, nil
}

// Update updates an existing scan record (for reload functionality)
func (m *Manager) Update(id string, results []*screener.ScreenResult, totalScanned int) (*ScanRecord, error) {
	now := time.Now()

	record := &ScanRecord{
		ID:           id,
		Timestamp:    now,
		TotalScanned: totalScanned,
		TotalFound:   len(results),
		Results:      results,
	}

	filename := filepath.Join(m.historyDir, fmt.Sprintf("scan_%s.json", id))

	data, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return nil, err
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return nil, err
	}

	return record, nil
}

// Load loads a specific scan record by ID
func (m *Manager) Load(id string) (*ScanRecord, error) {
	filename := filepath.Join(m.historyDir, fmt.Sprintf("scan_%s.json", id))

	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var record ScanRecord
	if err := json.Unmarshal(data, &record); err != nil {
		return nil, err
	}

	return &record, nil
}

// List returns all saved scan records (metadata only, no results)
func (m *Manager) List() ([]*ScanRecord, error) {
	files, err := os.ReadDir(m.historyDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []*ScanRecord{}, nil
		}
		return nil, err
	}

	var records []*ScanRecord

	for _, file := range files {
		if file.IsDir() || !strings.HasPrefix(file.Name(), "scan_") || !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		// Extract ID from filename
		id := strings.TrimPrefix(file.Name(), "scan_")
		id = strings.TrimSuffix(id, ".json")

		// Load just metadata
		record, err := m.Load(id)
		if err != nil {
			continue
		}

		// Clear results for list view (save memory)
		listRecord := &ScanRecord{
			ID:           record.ID,
			Timestamp:    record.Timestamp,
			TotalScanned: record.TotalScanned,
			TotalFound:   record.TotalFound,
			Results:      nil, // Don't load results for list
		}

		records = append(records, listRecord)
	}

	// Sort by timestamp descending (newest first)
	sort.Slice(records, func(i, j int) bool {
		return records[i].Timestamp.After(records[j].Timestamp)
	})

	return records, nil
}

// Delete removes a scan record
func (m *Manager) Delete(id string) error {
	filename := filepath.Join(m.historyDir, fmt.Sprintf("scan_%s.json", id))
	return os.Remove(filename)
}

// DeleteAll removes all scan records
func (m *Manager) DeleteAll() error {
	files, err := os.ReadDir(m.historyDir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasPrefix(file.Name(), "scan_") && strings.HasSuffix(file.Name(), ".json") {
			os.Remove(filepath.Join(m.historyDir, file.Name()))
		}
	}

	return nil
}

// Count returns the number of saved scans
func (m *Manager) Count() int {
	records, err := m.List()
	if err != nil {
		return 0
	}
	return len(records)
}

// GetLatest returns the most recent scan record
func (m *Manager) GetLatest() (*ScanRecord, error) {
	records, err := m.List()
	if err != nil {
		return nil, err
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("no history found")
	}

	// Load full record
	return m.Load(records[0].ID)
}

// FormatTimestamp formats timestamp for display
func FormatTimestamp(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	if diff < time.Minute {
		return "just now"
	} else if diff < time.Hour {
		mins := int(diff.Minutes())
		if mins == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", mins)
	} else if diff < 24*time.Hour {
		hours := int(diff.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	} else if diff < 7*24*time.Hour {
		days := int(diff.Hours() / 24)
		if days == 1 {
			return "yesterday"
		}
		return fmt.Sprintf("%d days ago", days)
	}

	return t.Format("Jan 02, 2006 15:04")
}
