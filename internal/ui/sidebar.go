package ui

import tea "charm.land/bubbletea/v2"

type Sidebar struct {
	Urls   []string
	Cursor int
	Width  int
}

func (s Sidebar) NewSidebar() *Sidebar {
	return &Sidebar{Urls: []string{}, Cursor: 0, Width: 100}
}

func (s Sidebar) View() tea.View {
	return s.View()
}

func (s Sidebar) Update(msg tea.Msg) (Sidebar, tea.Cmd) {
	return s, nil
}
