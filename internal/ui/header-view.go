package ui

import (
	"context"
	"errors"
	"net/url"

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
// pisanie umożliwić dopiero po wciśnięciu 'i'
// pisanie kończy się po kliknięciu 'enter' lub 'esc' (anuluj)
// dodać datę kiedy uruchomiony został test

var (
	activeBorderColor   = lipgloss.Color("62")  // Indigo/Purple
	inactiveBorderColor = lipgloss.Color("240") // Dark Gray
	activeTitleColor    = lipgloss.Color("212") // Pink
	inactiveTitleColor  = lipgloss.Color("240") // Dark Gray
)

type Status int

const (
	Idle Status = iota
	Loading
	Error
	Loaded
	Inserting
)

type runningCmd struct {
	ctx    context.Context
	cancel context.CancelFunc
}

type HeaderView struct {
	urlInput textinput.Model
	spinner  spinner.Model
	status   Status

	cmd    *runningCmd
	result lighthouse.LighthouseResult
	err    error

	width int
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

func NewHeader() *HeaderView {
	ti := textinput.New()
	ti.Placeholder = "Enter url"
	ti.SetVirtualCursor(true)
	ti.Focus()
	ti.SetValue("https://wojciechpietraszuk.vercel.app")

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return &HeaderView{urlInput: ti, spinner: s, status: Idle}
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

func cancelLighthouseExecution(h HeaderView) (HeaderView, tea.Cmd) {
	if h.cmd == nil {
		return h, nil
	}

	h.status = Idle
	h.cmd.cancel()
	h.cmd = nil

	return h, nil
}

func (h HeaderView) Init() tea.Cmd {
	return h.urlInput.Focus()
}

func (h HeaderView) Update(msg tea.Msg) (HeaderView, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "enter":
			url := h.urlInput.Value()
			h.err = validateUrl(url)

			if h.err != nil {
				h.status = Error
				return h, nil
			}

			h.status = Loading

			ctx, cancel := context.WithCancel(context.Background())

			h.cmd = &runningCmd{ctx: ctx, cancel: cancel}

			return h, tea.Batch(getLighthouseResult(ctx, url), h.spinner.Tick)
		case "esc":
			if h.status == Error {
				h.err = nil
				h.status = Idle

				return h, nil
			}

			if h.status == Inserting {
				h.status = Idle
				return h, nil
			}

			if h.status != Loading {
				return h, nil
			}

			return cancelLighthouseExecution(h)
		case "r":
			if h.status == Loading {
				return h, nil
			}

			h.urlInput.Reset()
			h.status = Idle
			h.err = nil

			return h, nil
		case "i":
			h.status = Inserting
			return h, nil
		}
	case LighthouseMsg:
		// if msg.err != nil {
		// 	h.err = msg.err
		//
		// 	return h, nil
		// }

		h.result = msg.result
		h.status = Loaded

		return h, nil
	case spinner.TickMsg:
		h.spinner, cmd = h.spinner.Update(msg)

		if h.status != Loading {
			return h, nil
		}

		return h, cmd
	}

	h.urlInput, cmd = h.urlInput.Update(msg)

	if h.urlInput.Err != nil {
		h.err = h.urlInput.Err
	}

	return h, cmd
}

func (h HeaderView) GetLighthouseResult() lighthouse.LighthouseResult {
	return h.result
}

func (h HeaderView) View() tea.View {
	switch h.status {
	case Error:
		errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("1"))

		return tea.NewView(errorStyle.Render("Error:", h.err.Error()))
	case Loading:
		return tea.NewView(lipgloss.JoinHorizontal(lipgloss.Left, h.urlInput.Value(), lipgloss.NewStyle().Foreground(lipgloss.Color("245")).MarginLeft(1).MarginRight(1).Render("Loading..."), h.spinner.View()))
	case Inserting:
		return tea.NewView(lipgloss.JoinHorizontal(lipgloss.Left, h.urlInput.View()))
	default:
		return tea.NewView(lipgloss.JoinHorizontal(lipgloss.Left, lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render("Press 'i' to insert url")))
	}

}

func (h HeaderView) SetWindowWidth(width int) {
	h.urlInput.SetWidth(width)
}
