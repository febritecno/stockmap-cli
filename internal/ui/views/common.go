package views

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// centerText centers text within a given width
func centerText(text string, width int) string {
	textWidth := lipgloss.Width(text)
	if textWidth >= width {
		return text
	}

	padding := (width - textWidth) / 2
	return strings.Repeat(" ", padding) + text
}
