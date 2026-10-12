package ui

import (
	"context"
	"net/url"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/wojtek0123/lighthouse-tui/internal/lighthouse"
	"github.com/wojtek0123/lighthouse-tui/internal/ui/components"
)

// func renderPane(title, content string, width, height int, isActive bool) string {
// 	borderColor := inactiveBorderColor
// 	titleColor := inactiveTitleColor
// 	if isActive {
// 		borderColor = activeBorderColor
// 		titleColor = activeTitleColor
// 	}
//
// 	border := lipgloss.RoundedBorder()
// 	borderStyle := lipgloss.NewStyle().Foreground(borderColor)
//
// 	// 1. Style and render the title
// 	titleStyle := lipgloss.NewStyle().
// 		Foreground(titleColor).
// 		Bold(isActive)
// 	renderedTitle := titleStyle.Render(" " + title + " ")
//
// 	// Format: ╭─ Title ────────────╮
// 	titleLen := lipgloss.Width(renderedTitle)
// 	topBarLength := width - titleLen - 2 // -2 for the ╭ and ╮ corners
//
// 	topBarLength = max(topBarLength, 0)
//
// 	leftDashCount := 1 // We want one dash before the title
// 	rightDashCount := topBarLength - leftDashCount
//
// 	rightDashCount = max(rightDashCount, 0)
//
// 	topLeft := borderStyle.Render(border.TopLeft)
// 	topRight := borderStyle.Render(border.TopRight)
// 	horizLineLeft := borderStyle.Render(strings.Repeat(border.Top, leftDashCount))
// 	horizLineRight := borderStyle.Render(strings.Repeat(border.Top, rightDashCount))
//
// 	topBorder := topLeft + horizLineLeft + renderedTitle + horizLineRight + topRight
//
// 	contentStyle := lipgloss.NewStyle().
// 		Border(border).
// 		BorderTop(false).
// 		BorderForeground(borderColor).
// 		Width(width).
// 		Height(height - 1) // -1 because the manual top border takes up 1 line
//
// 	renderedContent := contentStyle.Render(content)
//
// 	return lipgloss.JoinVertical(lipgloss.Left, topBorder, renderedContent)
// }

type ActivePaneMsg struct {
	pane Pane
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
	ReportsPane
	ResultPane
)

type Model struct {
	urlsView    *UrlsView
	resultView  *ResultView
	reportsView *ReportsView
	modal       *UrlModal

	showModal bool
	focused   Pane

	width  int
	height int

	error error
}

func InitialModel() Model {
	urls, err := lighthouse.GetHostnames()

	if err != nil {
		return Model{
			error: err,
		}
	}

	urlsView := NewUrlsView(true)
	resultView := NewResultView()
	reportsView := NewReportsView(false)
	modal := NewUrlModal()

	if len(urls) > 0 {
		reportsView.GetReports(urls[0])
		resultView.SetHostname(urls[0])
	}

	return Model{
		urlsView:    urlsView,
		resultView:  resultView,
		reportsView: reportsView,
		focused:     UrlsPane,
		modal:       modal,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "n":
			m.showModal = true
			return m, nil
		case "ctrl+c", "q":
			return m, tea.Quit
		case "tab":
			if m.focused == 2 {
				m.focused = UrlsPane
			} else {
				m.focused++
			}

			return m, m.setActivePane(m.focused)
		case "shift+tab":
			if m.focused == 0 {
				m.focused = ResultPane
			} else {
				m.focused--
			}

			return m, m.setActivePane(m.focused)
		case "0":
			m.focused = UrlsPane
			return m, m.setActivePane(UrlsPane)
		case "1":
			m.focused = ReportsPane
			return m, m.setActivePane(ReportsPane)
		case "2":
			m.focused = ResultPane
			return m, m.setActivePane(ResultPane)
		}
	case CloseModalMsg:
		m.showModal = false
		return m, nil
	case EnteredUrlMsg:
		ctx, cancel := context.WithCancel(context.Background())

		cmd := &runningCmd{ctx: ctx, cancel: cancel}

		u, err := url.Parse(msg.Url)
		if err != nil {
			return m, nil
		}

		return m, tea.Batch(getLighthouseResult(cmd, msg.Url), AddUniqueUrl(&m.urlsView.list, UrlItem{hostname: u.Hostname(), loading: true, cmd: cmd}), m.urlsView.spinner.Tick)
	case RestartMsg:
		ctx, cancel := context.WithCancel(context.Background())
		cmd := &runningCmd{ctx: ctx, cancel: cancel}

		m.urlsView.SetLoading(cmd, msg.Index, true)

		return m, tea.Batch(getLighthouseResult(cmd, msg.Url), m.urlsView.spinner.Tick)
	case LighthouseMsg:
		if msg.err != nil {
			return m, nil
		}

		m.resultView.audits = msg.result.Report
		m.resultView.loaded = true

		cmds = append(cmds, StopLoading(&m.urlsView.list, msg.url))

		u, err := url.Parse(msg.url)

		if err != nil {
			return m, nil
		}

		if item, ok := m.urlsView.list.SelectedItem().(UrlItem); ok && item.hostname == u.Hostname() {
			cmds = append(cmds, m.reportsView.InsertReport(msg.result.ReportName))
		}

		return m, tea.Batch(cmds...)
	case components.ActiveTabMsg:
		m.resultView.SetActiveTab(msg.Index)

		return m, nil
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	if m.showModal {
		*m.modal, cmd = m.modal.Update(msg)
		cmds = append(cmds, cmd)
	} else {
		*m.urlsView, cmd = m.urlsView.Update(msg)
		cmds = append(cmds, cmd)

		*m.reportsView, cmd = m.reportsView.Update(msg)
		cmds = append(cmds, cmd)

		*m.resultView, cmd = m.resultView.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() tea.View {
	var v tea.View

	if m.error != nil {
		v = tea.NewView(m.error.Error())
	} else if m.showModal {
		v = m.modal.View(m.width, m.height)
	} else {
		bordersHeight := 2
		sidePanelView := ""
		resultView := ""

		if m.width > 150 {
			asideWidth := min(m.width/5, 50)

			urlsView := RenderPane("[0] Urls", m.urlsView.View(asideWidth, m.height-bordersHeight), asideWidth, m.height, m.focused == UrlsPane)
			reportsView := RenderPane("[1] Reports", m.reportsView.View(m.height-bordersHeight), asideWidth, m.height, m.focused == ReportsPane)

			sidePanelView = lipgloss.NewStyle().Height(m.height).Render(lipgloss.JoinHorizontal(lipgloss.Left, urlsView, reportsView))

			resultView = RenderPane("[2] Result", m.resultView.View(m.width-asideWidth*2, m.height), m.width-asideWidth*2, m.height, m.focused == ResultPane)
		} else {
			asideWidth := min(m.width/3, 50)
			hPanesCount := 2

			spareSpace := m.height - ((m.height / hPanesCount) * hPanesCount)

			urlsView := RenderPane("[0] Urls", m.urlsView.View(asideWidth, (m.height/hPanesCount)-bordersHeight), asideWidth, m.height/2, m.focused == UrlsPane)
			reportsView := RenderPane("[1] Reports", m.reportsView.View((m.height/hPanesCount)+spareSpace-bordersHeight), asideWidth, (m.height/2)+spareSpace, m.focused == ReportsPane)

			sidePanelView = lipgloss.NewStyle().Height(m.height).Render(lipgloss.JoinVertical(lipgloss.Top, urlsView, reportsView))
			resultView = RenderPane("[2] Result", m.resultView.View(m.width-asideWidth, m.height), m.width-asideWidth, m.height, m.focused == ResultPane)
		}

		v = tea.NewView(lipgloss.JoinHorizontal(lipgloss.Top, sidePanelView, resultView))
	}

	v.AltScreen = true

	return v
}

func (m *Model) setActivePane(pane Pane) tea.Cmd {
	m.reportsView.SetActive(m.focused == ReportsPane)
	m.urlsView.SetActive(m.focused == UrlsPane)

	return func() tea.Msg {
		return ActivePaneMsg{
			pane: pane,
		}
	}
}
