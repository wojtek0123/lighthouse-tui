package components

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type Modal struct {
}

func NewModal() *Modal {
	return &Modal{}
}

func (m Modal) Init() tea.Cmd {
	return nil
}

func (m Modal) Update(msg tea.Msg) (Modal, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		}
	}

	return m, cmd
}

func (m Modal) View(width, height int) tea.View {

	background := lipgloss.NewStyle().Background(lipgloss.Black).Padding(1, 2)

	return tea.NewView(lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, background.Render("MODAL")))
}

func RenderModal(backdrop, modal string) string {
	return ""
}
