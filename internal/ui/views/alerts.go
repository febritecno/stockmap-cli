package views

import (
	"fmt"
	"strings"

	"github.com/febritecno/stockmap-cli/internal/alerts"
	"github.com/febritecno/stockmap-cli/internal/screener"
	"github.com/febritecno/stockmap-cli/internal/styles"
	"github.com/febritecno/stockmap-cli/internal/ui/components"
)

// AlertsView displays and manages price alerts
type AlertsView struct {
	width        int
	height       int
	alertsMgr    *alerts.Manager
	alerts       []*alerts.Alert
	triggered    []alerts.TriggeredAlert
	cursor       int
	inputActive  bool
	inputBuffer  string
	inputField   int // 0=symbol, 1=type, 2=threshold
	newSymbol    string
	newType      alerts.AlertType
	newThreshold string
	currentStock *screener.ScreenResult // Current stock context for quick add
	message      string
}

// NewAlertsView creates a new alerts view
func NewAlertsView(mgr *alerts.Manager) *AlertsView {
	return &AlertsView{
		alertsMgr: mgr,
		newType:   alerts.AlertBelow,
	}
}

// SetSize sets the view dimensions
func (a *AlertsView) SetSize(width, height int) {
	a.width = width
	a.height = height
}

// SetCurrentStock sets the current stock context for quick alert creation
func (a *AlertsView) SetCurrentStock(stock *screener.ScreenResult) {
	a.currentStock = stock
	if stock != nil {
		a.newSymbol = stock.Symbol
	}
}

// Refresh reloads alerts from manager
func (a *AlertsView) Refresh() {
	a.alerts = a.alertsMgr.GetAll()
	a.triggered = a.alertsMgr.GetTriggeredAlerts()
	if a.cursor >= len(a.alerts) {
		a.cursor = len(a.alerts) - 1
	}
	if a.cursor < 0 {
		a.cursor = 0
	}
}

// MoveUp moves cursor up
func (a *AlertsView) MoveUp() {
	if a.cursor > 0 {
		a.cursor--
	}
}

// MoveDown moves cursor down
func (a *AlertsView) MoveDown() {
	if a.cursor < len(a.alerts)-1 {
		a.cursor++
	}
}

// SelectedAlert returns the currently selected alert
func (a *AlertsView) SelectedAlert() *alerts.Alert {
	if a.cursor >= 0 && a.cursor < len(a.alerts) {
		return a.alerts[a.cursor]
	}
	return nil
}

// DeleteSelected removes the selected alert
func (a *AlertsView) DeleteSelected() error {
	if alert := a.SelectedAlert(); alert != nil {
		if err := a.alertsMgr.Remove(alert.ID); err != nil {
			return err
		}
		a.Refresh()
		a.message = "Alert deleted"
	}
	return nil
}

// ToggleSelected toggles the active state of selected alert
func (a *AlertsView) ToggleSelected() error {
	if alert := a.SelectedAlert(); alert != nil {
		active, err := a.alertsMgr.ToggleActive(alert.ID)
		if err != nil {
			return err
		}
		a.Refresh()
		if active {
			a.message = "Alert activated"
		} else {
			a.message = "Alert deactivated"
		}
	}
	return nil
}

// ResetSelected resets a triggered alert
func (a *AlertsView) ResetSelected() error {
	if alert := a.SelectedAlert(); alert != nil && alert.IsTriggered {
		if err := a.alertsMgr.ResetAlert(alert.ID); err != nil {
			return err
		}
		a.Refresh()
		a.message = "Alert reset"
	}
	return nil
}

// IsInputActive returns whether input mode is active
func (a *AlertsView) IsInputActive() bool {
	return a.inputActive
}

// ToggleInput toggles input mode for adding new alert
func (a *AlertsView) ToggleInput() {
	a.inputActive = !a.inputActive
	if a.inputActive {
		a.inputField = 0
		a.inputBuffer = ""
		if a.currentStock != nil {
			a.newSymbol = a.currentStock.Symbol
			a.inputField = 1 // Skip to type selection
		}
	}
}

// AddChar adds a character to input buffer
func (a *AlertsView) AddChar(c rune) {
	if !a.inputActive {
		return
	}

	switch a.inputField {
	case 0: // Symbol
		if c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z' {
			a.newSymbol += strings.ToUpper(string(c))
		}
	case 2: // Threshold
		if (c >= '0' && c <= '9') || c == '.' {
			a.newThreshold += string(c)
		}
	}
}

// Backspace removes last character from input
func (a *AlertsView) Backspace() {
	if !a.inputActive {
		return
	}

	switch a.inputField {
	case 0:
		if len(a.newSymbol) > 0 {
			a.newSymbol = a.newSymbol[:len(a.newSymbol)-1]
		}
	case 2:
		if len(a.newThreshold) > 0 {
			a.newThreshold = a.newThreshold[:len(a.newThreshold)-1]
		}
	}
}

// NextInputField moves to next input field
func (a *AlertsView) NextInputField() {
	if a.inputField < 2 {
		a.inputField++
	}
}

// PrevInputField moves to previous input field
func (a *AlertsView) PrevInputField() {
	if a.inputField > 0 {
		a.inputField--
	}
}

// CycleAlertType cycles through alert types
func (a *AlertsView) CycleAlertType() {
	types := []alerts.AlertType{
		alerts.AlertBelow,
		alerts.AlertAbove,
		alerts.AlertCross,
		alerts.AlertChange,
		alerts.AlertRSILow,
		alerts.AlertRSIHigh,
	}

	for i, t := range types {
		if t == a.newType {
			a.newType = types[(i+1)%len(types)]
			return
		}
	}
	a.newType = alerts.AlertBelow
}

// SubmitAlert creates a new alert from current input
func (a *AlertsView) SubmitAlert(lastPrice float64) error {
	if a.newSymbol == "" {
		a.message = "Symbol required"
		return nil
	}

	var threshold float64
	if a.newThreshold != "" {
		fmt.Sscanf(a.newThreshold, "%f", &threshold)
	}

	if threshold <= 0 {
		a.message = "Valid threshold required"
		return nil
	}

	_, err := a.alertsMgr.Add(a.newSymbol, a.newType, threshold, lastPrice)
	if err != nil {
		return err
	}

	a.message = fmt.Sprintf("Alert created for %s", a.newSymbol)
	a.inputActive = false
	a.newSymbol = ""
	a.newThreshold = ""
	a.Refresh()
	return nil
}

// ClearInput clears input mode
func (a *AlertsView) ClearInput() {
	a.inputActive = false
	a.newSymbol = ""
	a.newThreshold = ""
	a.inputField = 0
}

// ClearTriggered clears triggered alerts
func (a *AlertsView) ClearTriggered() {
	a.alertsMgr.ClearTriggeredAlerts()
	a.triggered = nil
	a.message = "Triggered alerts cleared"
}

// GetTriggeredCount returns the number of triggered alerts
func (a *AlertsView) GetTriggeredCount() int {
	return len(a.triggered)
}

// View renders the alerts view
func (a *AlertsView) View() string {
	var b strings.Builder

	// Header
	b.WriteString(styles.TitleStyle.Render("PRICE ALERTS"))
	b.WriteString("\n")
	b.WriteString(components.RenderDivider(a.width))
	b.WriteString("\n\n")

	// Show triggered alerts if any
	if len(a.triggered) > 0 {
		b.WriteString(styles.ScoreHighStyle.Render("TRIGGERED ALERTS"))
		b.WriteString("\n")
		for _, ta := range a.triggered {
			icon := styles.ScoreHighStyle.Render("!")
			msg := fmt.Sprintf("%s %s @ $%.2f - %s",
				ta.Alert.Symbol,
				alerts.FormatAlertType(ta.Alert.Type),
				ta.CurrentPrice,
				ta.Alert.Message,
			)
			b.WriteString(fmt.Sprintf("  %s %s\n", icon, msg))
		}
		b.WriteString("\n")
	}

	// Input form if active
	if a.inputActive {
		b.WriteString(styles.TitleStyle.Render("NEW ALERT"))
		b.WriteString("\n")

		// Symbol field
		symbolLabel := "Symbol: "
		if a.inputField == 0 {
			symbolLabel = styles.ScoreHighStyle.Render("> ") + symbolLabel
		} else {
			symbolLabel = "  " + symbolLabel
		}
		b.WriteString(symbolLabel + styles.InfoStyle.Render(a.newSymbol) + "_\n")

		// Type field
		typeLabel := "Type: "
		if a.inputField == 1 {
			typeLabel = styles.ScoreHighStyle.Render("> ") + typeLabel
		} else {
			typeLabel = "  " + typeLabel
		}
		b.WriteString(typeLabel + styles.InfoStyle.Render(alerts.FormatAlertType(a.newType)) + " (press SPACE to cycle)\n")

		// Threshold field
		thresholdLabel := "Threshold: "
		if a.inputField == 2 {
			thresholdLabel = styles.ScoreHighStyle.Render("> ") + thresholdLabel
		} else {
			thresholdLabel = "  " + thresholdLabel
		}
		thresholdValue := a.newThreshold
		if thresholdValue == "" {
			thresholdValue = "0"
		}
		b.WriteString(thresholdLabel + styles.InfoStyle.Render(thresholdValue) + "_\n")

		b.WriteString("\n")
		b.WriteString(styles.HelpStyle.Render("Press TAB to switch fields, SPACE to cycle type, ENTER to create, ESC to cancel"))
		b.WriteString("\n\n")
	}

	// Alerts list
	b.WriteString(styles.TitleStyle.Render("ALL ALERTS"))
	activeCount := a.alertsMgr.GetActiveCount()
	b.WriteString(fmt.Sprintf(" (%d active)\n", activeCount))

	if len(a.alerts) == 0 {
		b.WriteString(styles.MutedStyle().Render("  No alerts set. Press [N] to add one."))
		b.WriteString("\n")
	} else {
		for i, alert := range a.alerts {
			cursor := "  "
			if i == a.cursor {
				cursor = styles.ScoreHighStyle.Render("> ")
			}

			// Status indicator
			var status string
			if alert.IsTriggered {
				status = styles.ScoreHighStyle.Render("[TRIGGERED]")
			} else if !alert.IsActive {
				status = styles.MutedStyle().Render("[INACTIVE]")
			} else {
				status = styles.ScoreMediumStyle.Render("[ACTIVE]")
			}

			// Alert info
			info := fmt.Sprintf("%s %s $%.2f",
				alert.Symbol,
				alerts.FormatAlertType(alert.Type),
				alert.Threshold,
			)

			b.WriteString(fmt.Sprintf("%s%s %s\n", cursor, status, info))
		}
	}

	b.WriteString("\n")

	// Message
	if a.message != "" {
		b.WriteString(styles.InfoStyle.Render(a.message))
		b.WriteString("\n")
	}

	// Help
	if a.inputActive {
		b.WriteString(styles.HelpStyle.Render("[ESC] Cancel"))
	} else {
		b.WriteString(styles.HelpStyle.Render("[N] New Alert  [D] Delete  [T] Toggle  [R] Reset  [C] Clear Triggered  [ESC] Back"))
	}

	return b.String()
}
