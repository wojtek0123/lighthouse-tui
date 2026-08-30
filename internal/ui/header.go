package ui

import (
	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/wojtek0123/lighthouse-tui/internal/lighthouse"
)

var (
	activeBorderColor   = lipgloss.Color("62")  // Indigo/Purple
	inactiveBorderColor = lipgloss.Color("240") // Dark Gray
	activeTitleColor    = lipgloss.Color("212") // Pink
	inactiveTitleColor  = lipgloss.Color("240") // Dark Gray
)

type Header struct {
	urlInput textinput.Model
	loading  bool
	result   lighthouse.LighthouseResult
	spinner  spinner.Model

	width int
}

func NewHeader() *Header {
	ti := textinput.New()
	ti.Placeholder = "Enter url"
	ti.SetVirtualCursor(true)
	ti.Focus()

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return &Header{urlInput: ti, spinner: s}
}

type LighthouseMsg struct {
	result lighthouse.LighthouseResult
	err    error
}

func getLighthouseResult(url string) tea.Cmd {
	return func() tea.Msg {
		result, err := lighthouse.RunLighthouse(url)

		return LighthouseMsg{
			result: result,
			err:    err,
		}
	}
}

func (h Header) Init() tea.Cmd {
	return h.urlInput.Focus()
}

func (h Header) Update(msg tea.Msg) (Header, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "enter":
			url := h.urlInput.Value()

			h.loading = true

			return h, tea.Batch(getLighthouseResult(url), h.spinner.Tick)
		}
	case LighthouseMsg:
		h.result = msg.result
		h.loading = false

		return h, nil
	case spinner.TickMsg:
		h.spinner, cmd = h.spinner.Update(msg)

		if h.loading == false {
			return h, nil
		}

		return h, cmd
	}

	h.urlInput, cmd = h.urlInput.Update(msg)
	return h, cmd
}

func (h Header) GetLighthouseResult() lighthouse.LighthouseResult {
	return h.result
}

func (h Header) View() tea.View {
	v := tea.NewView(lipgloss.JoinHorizontal(lipgloss.Left, h.urlInput.View(), h.spinner.View()))

	return v
}

func (h Header) SetWindowWidth(width int) {
	h.urlInput.SetWidth(width)
}
