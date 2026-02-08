package components

import (
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss"

	"stockmap/internal/styles"
)

// Spinner frames for reload animation
var reloadSpinnerFrames = []string{".", "..", "...", "....", "....."}

// StatusBar represents the bottom status bar
type StatusBar struct {
	width         int
	scanned       int
	found         int
	lastUpdate    time.Time
	message       string
	isScanning    bool
	currentSym    string
	isReloading   bool
	reloadFrame   int
	autoReload    bool
	autoReloadSec int // seconds until next reload
}

// NewStatusBar creates a new status bar
func NewStatusBar() *StatusBar {
	return &StatusBar{
		lastUpdate: time.Now(),
	}
}

// SetWidth sets the status bar width
func (s *StatusBar) SetWidth(w int) {
	s.width = w
}

// SetStats updates the statistics
func (s *StatusBar) SetStats(scanned, found int) {
	s.scanned = scanned
	s.found = found
	s.lastUpdate = time.Now()
}

// SetScanning sets the scanning state
func (s *StatusBar) SetScanning(scanning bool, currentSymbol string) {
	s.isScanning = scanning
	s.currentSym = currentSymbol
}

// SetReloading sets the reloading state (simple spinner)
func (s *StatusBar) SetReloading(reloading bool) {
	s.isReloading = reloading
	if !reloading {
		s.reloadFrame = 0
	}
}

// NextReloadFrame advances the reload spinner
func (s *StatusBar) NextReloadFrame() {
	s.reloadFrame++
}

// SetAutoReload sets auto-reload state
func (s *StatusBar) SetAutoReload(enabled bool, secondsLeft int) {
	s.autoReload = enabled
	s.autoReloadSec = secondsLeft
}

// SetMessage sets a custom message
func (s *StatusBar) SetMessage(msg string) {
	s.message = msg
}

// View renders the status bar
func (s *StatusBar) View() string {
	// Key bindings - responsive based on width
	var keysLine string

	if s.width < 80 {
		// Compact keys for narrow screens
		keys := []struct {
			key  string
			desc string
		}{
			{"S", ""},
			{"R", ""},
			{"F", ""},
			{"D", ""},
			{"W", ""},
			{"I", ""},
			{"Q", ""},
		}

		for i, k := range keys {
			if i > 0 {
				keysLine += " "
			}
			keysLine += styles.KeyStyle.Render("[" + k.key + "]")
		}
	} else {
		// Full keys for wide screens
		keys := []struct {
			key  string
			desc string
		}{
			{"S", "can"},
			{"R", "eload"},
			{"T", "imer"},
			{"F", "ilter"},
			{"D", "etails"},
			{"W", "atch"},
			{"H", "istory"},
			{"I", "nfo"},
			{"Q", "uit"},
		}

		for i, k := range keys {
			if i > 0 {
				keysLine += "  "
			}
			keysLine += styles.KeyStyle.Render("["+k.key+"]") + styles.HelpStyle.Render(k.desc)
		}
	}

	// Divider
	divider := RenderDivider(s.width)

	// Stats line
	var statsLine string

	if s.isReloading {
		// Simple reload spinner at bottom
		spinner := reloadSpinnerFrames[s.reloadFrame%len(reloadSpinnerFrames)]
		statsLine = styles.SpinnerStyle.Render("Reloading"+spinner) +
			styles.MutedStyle().Render(fmt.Sprintf(" (%d/%d)", s.scanned, s.found))
	} else if s.isScanning {
		statsLine = styles.SpinnerStyle.Render("Scanning: ") +
			styles.InfoStyle.Render(s.currentSym) +
			styles.MutedStyle().Render(fmt.Sprintf(" (%d scanned)", s.scanned))
	} else {
		statsLine = styles.StatusItemStyle.Render(fmt.Sprintf("Scanned: %d", s.scanned)) +
			styles.MutedStyle().Render(" | ") +
			styles.StatusItemStyle.Render(fmt.Sprintf("Found: %d", s.found)) +
			styles.MutedStyle().Render(" | ") +
			styles.StatusItemStyle.Render("Updated: "+s.lastUpdate.Format("15:04:05"))
	}

	// Auto-reload indicator
	if s.autoReload {
		autoStatus := ""
		if s.isReloading {
			autoStatus = styles.ScoreHighStyle.Render(" [AUTO ON]")
		} else {
			autoStatus = styles.ScoreHighStyle.Render(fmt.Sprintf(" [AUTO: %ds]", s.autoReloadSec))
		}
		statsLine += autoStatus
	}

	if s.message != "" && !s.isReloading {
		statsLine += styles.MutedStyle().Render(" | ") + styles.InfoStyle.Render(s.message)
	}

	// Combine
	content := lipgloss.JoinVertical(lipgloss.Left,
		divider,
		keysLine,
		statsLine,
	)

	return styles.StatusBarStyle.Width(s.width).Render(content)
}

// ViewCompact renders a compact version for smaller screens
func (s *StatusBar) ViewCompact() string {
	keys := styles.KeyStyle.Render("[S]") + "can " +
		styles.KeyStyle.Render("[W]") + "atch " +
		styles.KeyStyle.Render("[D]") + "etails " +
		styles.KeyStyle.Render("[Q]") + "uit"

	stats := fmt.Sprintf("Found: %d", s.found)

	return styles.StatusBarStyle.Render(keys + " | " + stats)
}
