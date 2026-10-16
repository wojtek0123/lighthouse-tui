// Package components provides custom components.
package components

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// Tabs represents tabs state, manages tabs, and active and focus state.
type Tabs struct {
	items  []string
	focus  bool
	active int
}

// NewTabs creates tab component.
func NewTabs() *Tabs {
	return &Tabs{}
}

// Update is an update loop.
func (t Tabs) Update(msg tea.Msg) (Tabs, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "[":
			if t.active > 0 {
				t.active--
			}
			return t, t.ChangeActiveTab()
		case "]":
			if t.active < len(t.items)-1 {
				t.active++
			}
			return t, t.ChangeActiveTab()
		}
	case tea.WindowSizeMsg:
		return t, nil
	}

	return t, cmd
}

// View returns tabs view.
func (t Tabs) View() string {
	inactiveStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	activeTabStyle := lipgloss.NewStyle().Bold(false).Foreground(lipgloss.Color("212"))

	tabs := []string{}

	for i, tab := range t.items {
		if i == t.active && t.focus {
			tabs = append(tabs, activeTabStyle.Render(tab))
		} else {
			tabs = append(tabs, inactiveStyle.Render(tab))
		}
	}

	return lipgloss.JoinHorizontal(lipgloss.Left, strings.Join(tabs, " "))
}

// SetTabs sets available tabs.
func (t *Tabs) SetTabs(tabs []string) {
	t.items = tabs
}

// SetActive sets active tab by index.
func (t *Tabs) SetActive(i int) {
	t.active = i
}

// Active return active tab index.
func (t Tabs) Active() int {
	return t.active
}

// SetFocus sets which tab is focus.
func (t *Tabs) SetFocus(focus bool) {
	t.focus = focus
}

// ActiveTabMsg is a Bubble Tea message dispatched when active tab changed.
type ActiveTabMsg struct {
	Index int
	Label string
}

// ChangeActiveTab returns tea.Cmd to notify others views.
func (t Tabs) ChangeActiveTab() tea.Cmd {
	if len(t.items) == 0 {
		return nil
	}

	return func() tea.Msg {
		return ActiveTabMsg{
			Index: t.active,
			Label: t.items[t.active],
		}
	}
}
