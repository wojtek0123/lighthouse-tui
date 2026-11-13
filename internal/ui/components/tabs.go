// Package components provides custom components.
package components

import (
	"strings"

	"charm.land/lipgloss/v2"
)

// Tab represents tabs state, manages tabs, and active and focus state.
type Tab struct {
	items []string
	focus bool
	index int
}

// NewTabs creates tab component.
func NewTabs() *Tab {
	return &Tab{}
}

// View returns tabs view.
func (t *Tab) View() string {
	inactiveStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	activeTabStyle := lipgloss.NewStyle().Bold(false).Foreground(lipgloss.Color("212"))

	tabs := []string{}

	for i, tab := range t.items {
		if i == t.index && t.focus {
			tabs = append(tabs, activeTabStyle.Render(tab))
		} else {
			tabs = append(tabs, inactiveStyle.Render(tab))
		}
	}

	return lipgloss.JoinHorizontal(lipgloss.Left, strings.Join(tabs, " "))
}

// SetTabs sets available tabs.
func (t *Tab) SetTabs(tabs []string) {
	t.items = tabs
}

// SetActive sets active tab by index.
func (t *Tab) SetActive(i int) {
	t.index = i
}

// Active return active tab index.
func (t Tab) Active() int {
	return t.index
}

// SetFocus sets which tab is focus.
func (t *Tab) SetFocus(focus bool) {
	t.focus = focus
}

// Next changes tab index to next
func (t *Tab) Next() {
	if t.index == len(t.items)-1 {
		t.index = 0
	} else {
		t.index++
	}
}

// Previous changes tab index to previous
func (t *Tab) Previous() {
	if t.index == 0 {
		t.index = len(t.items) - 1
	} else {
		t.index--
	}
}

func (t Tab) Index() int {
	return t.index
}
