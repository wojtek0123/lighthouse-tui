package ui

import (
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"errors"
	"net/url"
)

type EnteredUrlMsg struct {
	Url string
}

type CloseModalMsg struct{}

func closeModal() tea.Cmd {
	return func() tea.Msg {
		return CloseModalMsg{}
	}
}

func enterUrl(url string) tea.Cmd {
	return func() tea.Msg {
		return EnteredUrlMsg{
			Url: url,
		}
	}
}

func validateUrl(rawUrl string) error {
	u, err := url.ParseRequestURI(rawUrl)

	if err != nil {
		return err
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return errors.New("URL must have start with 'http' or 'https'")
	}

	if u.Host == "" {
		return errors.New("URL doesn't have host part")
	}

	return nil
}

type Modal struct {
	input textinput.Model
	err   error
}

func NewModal() *Modal {
	ti := textinput.New()
	ti.Placeholder = "Enter url"
	ti.SetVirtualCursor(true)
	ti.Focus()
	ti.SetValue("https://wojciechpietraszuk.vercel.app")

	return &Modal{input: ti}
}

func (m Modal) Init() tea.Cmd {
	return nil
}

func (m Modal) Update(msg tea.Msg) (Modal, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return m, closeModal()
		case "enter":
			url := m.input.Value()

			m.err = validateUrl(url)

			return m, tea.Batch(enterUrl(url), closeModal())
		}
	}

	m.input, cmd = m.input.Update(msg)

	if m.input.Err != nil {
		m.err = m.input.Err
	}

	return m, cmd
}

func (m Modal) View(width, height int) tea.View {
	m.input.SetWidth(width / 4)

	background := lipgloss.NewStyle().Background(lipgloss.Black).Padding(1, 2)

	return tea.NewView(lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, background.Render(m.input.View())))
}

func RenderModal(backdrop, modal string) string {
	return ""
}
