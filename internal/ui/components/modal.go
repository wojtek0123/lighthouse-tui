package components

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// Modal represents modal state.
type Modal struct{}

// NewModal returns modal.
func NewModal() *Modal {
	return &Modal{}
}

// Init initializes modal state.
func (m Modal) Init() tea.Cmd {
	return nil
}

// Update runs in a loop to keep modal state updated.
func (m Modal) Update() (Modal, tea.Cmd) {
	return m, nil
}

// View renders modal content.
func (m Modal) View(width, height int) tea.View {
	background := lipgloss.NewStyle().Background(lipgloss.Black).Padding(1, 2)

	return tea.NewView(lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, background.Render("MODAL")))
}

// RenderModal returns modal as string.
func RenderModal() string {
	return ""
}
