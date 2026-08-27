package ui

import (
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
)

type Header struct {
	urlInput textinput.Model
}

func NewTopBar() *Header {
	ti := textinput.New()
	ti.Placeholder = "Enter url"
	ti.SetVirtualCursor(true)
	ti.Focus()
	ti.CharLimit = 156
	ti.SetWidth(20)

	return &Header{urlInput: ti}
}

func (h Header) Init() tea.Cmd {
	return h.urlInput.Focus()
}

func (h Header) Update(msg tea.Msg) (Header, tea.Cmd) {
	var cmd tea.Cmd

	h.urlInput, cmd = h.urlInput.Update(msg)
	return h, cmd
}

func (h Header) View() tea.View {
	return tea.NewView(h.urlInput.View())
}
