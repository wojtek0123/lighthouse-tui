// Package ui provides visual aspect of an app
package ui

import (
	"errors"
	"fmt"
	"net/url"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// EnteredURLMsg is a Bubble Tea message dispatched when a user submits a URL.
type EnteredURLMsg struct {
	URL string
}

// CloseModalMsg is a Bubble Tea message signaling that any active modal should be closed.
type CloseModalMsg struct{}

// closeModal returns a Bubble Tea command that triggers a CloseModalMsg,
// instructing the update loop to dismiss the current modal.
func closeModal() tea.Cmd {
	return func() tea.Msg {
		return CloseModalMsg{}
	}
}

// enterURL returns a Bubble Tea command that wraps the given URL into an
// EnteredURLMsg message for processing in the update loop.
func enterURL(url string) tea.Cmd {
	return func() tea.Msg {
		return EnteredURLMsg{
			URL: url,
		}
	}
}

// validateURL checks whether provided url is valid.
func validateURL(rawURL string) error {
	u, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return errors.New("URL must start with http or https")
	}

	if u.Host == "" {
		return errors.New("URL is missing a host")
	}

	return nil
}

// URLModal represents URL modal state with input and error.
type URLModal struct {
	input    textinput.Model
	hasError bool
	err      error
}

// NewURLModal returns URLModal object.
func NewURLModal() *URLModal {
	ti := textinput.New()
	ti.Placeholder = "Enter URL"
	ti.SetVirtualCursor(true)
	ti.Focus()
	ti.SetValue("https://wojciechpietraszuk.vercel.app")

	return &URLModal{input: ti}
}

// Update updates URL modal state.
func (m URLModal) Update(msg tea.Msg) (URLModal, tea.Cmd) {
	var cmd tea.Cmd

	//nolint:gocritic // keeping switch structure for future key cases
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return m, closeModal()
		case "enter":
			input := m.input.Value()

			m.err = validateURL(input)

			if m.err != nil {
				m.hasError = true
				return m, nil
			}

			m.hasError = false

			return m, tea.Batch(enterURL(input), closeModal())
		}
	}

	m.input, cmd = m.input.Update(msg)

	return m, cmd
}

// View returns url modal ui.
func (m URLModal) View(width, height int) string {
	// TODO:
	// - handle smaller resolutions

	m.input.SetWidth(width / 4)

	background := lipgloss.NewStyle().Background(lipgloss.Black).Padding(1, 2)
	// errMsgStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("1"))

	errorMsg := ""

	// if m.hasError {
	// 	errorMsg = errMsgStyle.Render(m.err.Error())
	// }

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, background.Render(lipgloss.JoinVertical(lipgloss.Left, "", m.input.View(), errorMsg)))
}
