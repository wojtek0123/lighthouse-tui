package ui

import (
	"context"

	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/wojtek0123/lighthouse-tui/internal/lighthouse"
)

// TODO:
// 1. Handle cancellation
// 2. Add help
// 3. Add url validation
// 4. Remove '>' from input

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

type Header struct {
	urlInput textinput.Model
	loading  bool
	spinner  spinner.Model

	cmd    *runningCmd
	result lighthouse.LighthouseResult
	err    error

	width int
}

func NewHeader() *Header {
	ti := textinput.New()
	ti.Placeholder = "Enter url"
	ti.SetVirtualCursor(true)
	ti.Focus()
	ti.SetValue("https://wojciechpietraszuk.vercel.app")

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return &Header{urlInput: ti, spinner: s}
}

type LighthouseMsg struct {
	result lighthouse.LighthouseResult
	err    error
}

func getLighthouseResult(ctx context.Context, url string) tea.Cmd {
	return func() tea.Msg {
		result, err := lighthouse.RunLighthouse(ctx, url)

		return LighthouseMsg{
			result: result,
			err:    err,
		}
	}
}

func cancelLighthouseExecution(h Header) (Header, tea.Cmd) {
	if h.cmd == nil {
		return h, nil
	}

	h.loading = false
	h.cmd.cancel()
	h.cmd = nil

	return h, nil
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

			ctx, cancel := context.WithCancel(context.Background())

			h.cmd = &runningCmd{ctx: ctx, cancel: cancel}

			return h, tea.Batch(getLighthouseResult(ctx, url), h.spinner.Tick)

		case "esc", "ctrl+c":
			return cancelLighthouseExecution(h)
		}
	case LighthouseMsg:
		if msg.err != nil {
			h.err = msg.err

			return h, nil
		}

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
	if h.loading {
		return tea.NewView(lipgloss.JoinHorizontal(lipgloss.Left, h.urlInput.Value(), lipgloss.NewStyle().Foreground(lipgloss.Color("245")).MarginLeft(1).MarginRight(1).Render("Loading..."), h.spinner.View()))
	} else {
		return tea.NewView(lipgloss.JoinHorizontal(lipgloss.Left, h.urlInput.View()))
	}
}

func (h Header) SetWindowWidth(width int) {
	h.urlInput.SetWidth(width)
}
