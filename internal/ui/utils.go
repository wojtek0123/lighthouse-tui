package ui

import (
	"strings"

	"charm.land/lipgloss/v2"
)

// SpaceBetween renders label with value and spaces between them.
func SpaceBetween(w int, label, value string) string {
	space := w - lipgloss.Width(label) - lipgloss.Width(value) - 2

	if space < 0 {
		return lipgloss.JoinVertical(lipgloss.Left, label, value)
	}

	separator := strings.Repeat(" ", space)

	return lipgloss.JoinHorizontal(lipgloss.Left, label, separator, value)
}
