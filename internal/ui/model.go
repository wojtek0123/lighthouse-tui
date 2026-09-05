package ui

import (
	"context"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/wojtek0123/lighthouse-tui/internal/ui/components"
)

func renderPane(title, content string, width, height int, isActive bool) string {
	borderColor := inactiveBorderColor
	titleColor := inactiveTitleColor
	if isActive {
		borderColor = activeBorderColor
		titleColor = activeTitleColor
	}

	border := lipgloss.RoundedBorder()
	borderStyle := lipgloss.NewStyle().Foreground(borderColor)

	// 1. Style and render the title
	titleStyle := lipgloss.NewStyle().
		Foreground(titleColor).
		Bold(isActive)
	renderedTitle := titleStyle.Render(" " + title + " ")

	// 2. Construct the custom top border line
	// Format: ╭─ Title ────────────╮
	titleLen := lipgloss.Width(renderedTitle)
	topBarLength := width - titleLen - 2 // -2 for the ╭ and ╮ corners

	// Ensure we don't panic on extremely small terminals
	if topBarLength < 0 {
		topBarLength = 0
	}

	leftDashCount := 1 // We want one dash before the title
	rightDashCount := topBarLength - leftDashCount
	if rightDashCount < 0 {
		rightDashCount = 0
	}

	topLeft := borderStyle.Render(border.TopLeft)
	topRight := borderStyle.Render(border.TopRight)
	horizLineLeft := borderStyle.Render(strings.Repeat(border.Top, leftDashCount))
	horizLineRight := borderStyle.Render(strings.Repeat(border.Top, rightDashCount))

	topBorder := topLeft + horizLineLeft + renderedTitle + horizLineRight + topRight

	// 3. Render the content box
	// We disable the top border because we just built it manually above
	contentStyle := lipgloss.NewStyle().
		Border(border).
		BorderTop(false).
		BorderForeground(borderColor).
		Width(width).
		Height(height - 1) // -1 because the manual top border takes up 1 line

	renderedContent := contentStyle.Render(content)

	// 4. Stack the top border and the content box
	return lipgloss.JoinVertical(lipgloss.Left, topBorder, renderedContent)
}

type Pane int

type Status int

const (
	Idle Status = iota
	Loading
	Error
	Loaded
)

const (
	HeaderPane Pane = iota
	ResultPane
)

type App struct {
	// sidebar *Sidebar
	header *HeaderView
	result *ResultView
	modal  *components.Modal

	cmd       *runningCmd
	showModal bool
	focused   Pane

	width  int
	height int
}

func InitialModel() App {
	header := NewHeader()
	resultView := NewResultView()
	modal := components.NewModal()

	return App{
		header:  header,
		result:  resultView,
		focused: HeaderPane,
		modal:   modal,
	}
}

func (a App) Init() tea.Cmd {
	return a.header.urlInput.Focus()
}

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "esc":
			return cancelLighthouseExecution(a)
		case "n":
			a.showModal = true
			return a, nil
		case "ctrl+c", "q":
			return a, tea.Quit
		case "tab":
			if a.focused == 1 {
				a.focused = HeaderPane
			} else {
				a.focused++
			}

			return a, nil
		case "shift+tab":
			if a.focused == 0 {
				a.focused = ResultPane
			} else {
				a.focused--
			}

			return a, nil
		case "0":
			a.focused = HeaderPane
			return a, nil
		case "1":
			a.focused = ResultPane
			return a, nil
		case "r":
			if a.header.enteredUrl == "" {
				return a, nil
			}

			return a, func() tea.Msg { return EnteredUrlMsg{Url: a.header.enteredUrl} }
		}
	case CloseModalMsg:
		a.showModal = false
		return a, nil
	case EnteredUrlMsg:
		a.header.enteredUrl = msg.Url
		a.header.loading = true

		ctx, cancel := context.WithCancel(context.Background())
		a.cmd = &runningCmd{ctx: ctx, cancel: cancel}

		return a, tea.Batch(getLighthouseResult(ctx, msg.Url), a.header.spinner.Tick)
	case LighthouseMsg:
		a.result.audits = msg.result
		a.result.loaded = true
		a.header.loading = false

		return a, nil
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.header.urlInput.SetWidth(msg.Width)
	}

	header, cmd := a.header.Update(msg)
	*a.header = header
	cmds = append(cmds, cmd)

	if a.showModal == true {
		modal, cmd := a.modal.Update(msg)
		*a.modal = modal
		cmds = append(cmds, cmd)
	}

	return a, tea.Batch(cmds...)
}

func (a App) View() tea.View {
	var v tea.View

	if a.showModal {
		v = a.modal.View(a.width, a.height)
	} else {
		v = tea.NewView(lipgloss.JoinVertical(lipgloss.Center, renderPane("[0] URL", a.header.View().Content, a.width, 2, a.focused == HeaderPane), renderPane("[1] Results", a.result.View(a.width, a.height).Content, a.width, a.height, a.focused == ResultPane)))
	}

	v.AltScreen = true

	return v
}
