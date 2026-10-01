package ui

import (
	"context"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/wojtek0123/lighthouse-tui/internal/lighthouse"
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

	contentStyle := lipgloss.NewStyle().
		Border(border).
		BorderTop(false).
		BorderForeground(borderColor).
		Width(width).
		Height(height - 1) // -1 because the manual top border takes up 1 line

	renderedContent := contentStyle.Render(content)

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
	UrlsPane Pane = iota
	ResultPane
	ReportsPane
)

type App struct {
	urlsView    *UrlsView
	resultView  *ResultView
	reportsView *ReportsView
	modal       *UrlModal

	showModal bool
	focused   Pane

	width  int
	height int
}

func InitialModel() App {
	urlsView := NewUrlsView()
	resultView := NewResultView()
	reportsView := NewReportsView()
	modal := NewUrlModal()

	urls, err := lighthouse.GetUrls()

	if err != nil {
	}

	reportsView.GetReports(urls[0])

	return App{
		urlsView:    urlsView,
		resultView:  resultView,
		reportsView: reportsView,
		focused:     UrlsPane,
		modal:       modal,
	}
}

func (a App) Init() tea.Cmd {
	return nil
}

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "n":
			a.showModal = true
			return a, nil
		case "ctrl+c", "q":
			return a, tea.Quit
		case "tab":
			if a.focused == 1 {
				a.focused = UrlsPane
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
			a.focused = UrlsPane
			a.reportsView.active = false
			return a, nil
		case "1":
			a.focused = ReportsPane
			a.reportsView.active = true
			return a, nil
		case "2":
			a.focused = ResultPane
			a.reportsView.active = false
			return a, nil
		}
	case CloseModalMsg:
		a.showModal = false
		return a, nil
	case EnteredUrlMsg:
		ctx, cancel := context.WithCancel(context.Background())

		cmd := &runningCmd{ctx: ctx, cancel: cancel}

		index := 0

		return a, tea.Batch(getLighthouseResult(cmd, msg.Url), a.urlsView.list.InsertItem(index, Item{url: msg.Url, loading: true, cmd: cmd}), a.urlsView.spinner.Tick)
	case RestartMsg:
		ctx, cancel := context.WithCancel(context.Background())
		cmd := &runningCmd{ctx: ctx, cancel: cancel}

		a.urlsView.SetLoading(cmd, msg.Index, true)

		return a, tea.Batch(getLighthouseResult(cmd, msg.Url), a.urlsView.spinner.Tick)

	case LighthouseMsg:
		// if msg.err == nil {
		a.resultView.audits = msg.result
		a.resultView.loaded = true

		for index, item := range a.urlsView.list.Items() {
			if i, ok := item.(Item); ok == true {
				if msg.url == i.url {
					a.urlsView.StopLoading(index)
					break
				}
			}
		}
		// a.resultView.err = msg.err
		// } else {
		// 	a.resultView.err = msg.err
		// }

		return a, nil
	case SelectedUrlMsg:
		a.reportsView.GetReports(msg.url)

		return a, nil

	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
	}

	urlsView, cmd := a.urlsView.Update(msg)
	*a.urlsView = urlsView
	cmds = append(cmds, cmd)

	reportsView, cmd := a.reportsView.Update(msg)
	*a.reportsView = reportsView
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
		const leftPanelWidth = 60

		urlsView := renderPane("[0] Urls", a.urlsView.View(), leftPanelWidth, a.height/2, a.focused == UrlsPane)
		reportsView := renderPane("[1] Reports", a.reportsView.View(), leftPanelWidth, a.height/2, a.focused == ReportsPane)
		resultView := renderPane("[2] Results", a.resultView.View(a.width-leftPanelWidth, a.height).Content, a.width-leftPanelWidth, a.height, a.focused == ResultPane)

		sidePanelView := lipgloss.JoinVertical(lipgloss.Top, urlsView, reportsView)

		v = tea.NewView(lipgloss.JoinHorizontal(lipgloss.Top, sidePanelView, resultView))
	}

	v.AltScreen = true

	return v
}
