package components

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type Tabs struct {
	items  []string
	focus  bool
	active int
}

func NewTabs() *Tabs {
	return &Tabs{}
}

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
	}

	return t, cmd
}

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

func (t *Tabs) SetTabs(tabs []string) {
	t.items = tabs
}

func (t *Tabs) SetActive(i int) {
	t.active = i
}

func (t Tabs) Active() int {
	return t.active
}

func (t *Tabs) SetFocus(focus bool) {
	t.focus = focus
}

type ActiveTabMsg struct {
	Index int
	Label string
}

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
