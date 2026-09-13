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

type UrlModal struct {
	input    textinput.Model
	hasError bool
	err      error
}

func NewUrlModal() *UrlModal {
	ti := textinput.New()
	ti.Placeholder = "Enter url"
	ti.SetVirtualCursor(true)
	ti.Focus()
	ti.SetValue("https://wojciechpietraszuk.vercel.app")

	return &UrlModal{input: ti}
}

func (m UrlModal) Init() tea.Cmd {
	return nil
}

func (m UrlModal) Update(msg tea.Msg) (UrlModal, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return m, closeModal()
		case "enter":
			url := m.input.Value()

			m.err = validateUrl(url)

			if m.err != nil {
				m.hasError = true
				return m, nil
			} else {
				m.hasError = false
			}

			return m, tea.Batch(enterUrl(url), closeModal())
		}
	}

	m.input, cmd = m.input.Update(msg)

	return m, cmd
}

func (m UrlModal) View(width, height int) tea.View {
	// TODO:
	// - handle smaller resolutions

	m.input.SetWidth(width / 4)

	background := lipgloss.NewStyle().Background(lipgloss.Black).Padding(1, 2)
	errMsgStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("1"))

	errorMsg := ""

	if m.hasError == true {
		errorMsg = errMsgStyle.Render(m.err.Error())
	}

	return tea.NewView(lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, background.Render(lipgloss.JoinVertical(lipgloss.Left, "", m.input.View(), errorMsg))))
}
