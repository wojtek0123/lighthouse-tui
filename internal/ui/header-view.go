package ui

import (
	"context"

	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// TODO:
// 1. Handle cancellation
// 2. Add help
// 3. Add url validation
// 4. Remove '>' from input
// pisanie umożliwić dopiero po wciśnięciu 'i'
// pisanie kończy się po kliknięciu 'enter' lub 'esc' (anuluj)
// dodać datę kiedy uruchomiony został test

var (
	activeBorderColor   = lipgloss.Color("62")  // Indigo/Purple
	inactiveBorderColor = lipgloss.Color("240") // Dark Gray
	activeTitleColor    = lipgloss.Color("212") // Pink
	inactiveTitleColor  = lipgloss.Color("240") // Dark Gray
)

type runningCmd struct {
	ctx    context.Context
	cancel context.CancelFunc
}

type HeaderView struct {
	// urlInput textinput.Model
	spinner spinner.Model
	loading bool

	enteredUrl string
}

func NewHeaderView() *HeaderView {
	ti := textinput.New()
	ti.Placeholder = "Enter url"
	ti.SetVirtualCursor(true)
	ti.Focus()
	ti.SetValue("https://wojciechpietraszuk.vercel.app")

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return &HeaderView{spinner: s, loading: false}
}

func (h HeaderView) Init() tea.Cmd {
	return nil
}

func (h HeaderView) Update(msg tea.Msg) (HeaderView, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		}
	case spinner.TickMsg:
		h.spinner, cmd = h.spinner.Update(msg)

		if h.loading == false {
			return h, nil
		}

		return h, cmd
	}

	return h, cmd
}

func (h HeaderView) View() tea.View {
	if h.loading == true {
		return tea.NewView(lipgloss.JoinHorizontal(lipgloss.Left, h.enteredUrl, lipgloss.NewStyle().Foreground(lipgloss.Color("245")).MarginLeft(1).MarginRight(1).Render("Loading..."), h.spinner.View()))
	} else {
		if h.enteredUrl == "" {

			return tea.NewView(lipgloss.JoinHorizontal(lipgloss.Left, lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render("Press 'n' to enter url")))
		} else {
			return tea.NewView(lipgloss.JoinHorizontal(lipgloss.Left, lipgloss.NewStyle().Foreground(lipgloss.Color("15")).Render(h.enteredUrl)))
		}
	}
}
