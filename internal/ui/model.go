package ui

import tea "charm.land/bubbletea/v2"

type App struct {
	// sidebar *Sidebar
	header *Header
	// result  *Result
}

func InitialModel() App {
	top := NewTopBar()

	return App{
		header: top,
	}
}

func (a App) Init() tea.Cmd {
	return a.header.urlInput.Focus()
}

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return a, tea.Quit
		}
	}

	topBar, cmd := a.header.Update(msg)
	cmds = append(cmds, cmd)

	a.header = &topBar

	return a, tea.Batch(cmds...)
}

func (a App) View() tea.View {
	v := a.header.View()

	v.AltScreen = true

	return v
}
