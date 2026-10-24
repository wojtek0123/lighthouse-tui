package ui

import (
	"context"
	"net/url"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/wojtek0123/lighthouse-tui/internal/lighthouse"
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

// ActivePaneMsg is a Bubble Tea message dispatched when active pane changed.
type ActivePaneMsg struct {
	pane Pane
}

// Pane indicates active pane.
type Pane int

const (
	hostnamesPane Pane = iota
	reportsPane
	resultPane
)

// Model represents app main type.
type Model struct {
	hostnamesView *HostnamesView
	resultView    *ResultView
	reportsView   *ReportsView
	modal         *URLModal

	showModal bool
	focused   Pane

	width  int
	height int

	error error
}

// InitialModel creates initial state of the app.
func InitialModel() Model {
	urls, err := lighthouse.GetHostnames()

	if err != nil {
		return Model{
			error: err,
		}
	}

	urlsView := newHostnamesView(true)
	resultView := NewResultView()
	reportsView := NewReportsView(false)
	modal := NewURLModal()

	if len(urls) > 0 {
		reportsView.GetReports(urls[0])
		resultView.SetHostname(urls[0])
	}

	return Model{
		hostnamesView: urlsView,
		resultView:    resultView,
		reportsView:   reportsView,
		focused:       hostnamesPane,
		modal:         modal,
	}
}

// Init returns tea.Cmd to keep up to date others views.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update is running in loop to keep state up to date.
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
				m.focused = hostnamesPane
			} else {
				m.focused++
			}

			return m, m.setActivePane(m.focused)
		case "shift+tab":
			if m.focused == 0 {
				m.focused = resultPane
			} else {
				m.focused--
			}

			return m, m.setActivePane(m.focused)
		case "0":
			m.focused = hostnamesPane
			return m, m.setActivePane(hostnamesPane)
		case "1":
			m.focused = reportsPane
			return m, m.setActivePane(reportsPane)
		case "2":
			if m.focused != reportsPane {
				return m, nil
			}

			m.focused = resultPane
			return m, m.setActivePane(resultPane)
		}
	case CloseModalMsg:
		m.showModal = false
		return m, nil
	case EnteredURLMsg:
		ctx, cancel := context.WithCancel(context.Background())

		cmd := &runningCmd{ctx: ctx, cancel: cancel}

		u, err := url.Parse(msg.URL)
		if err != nil {
			return m, nil
		}

		return m, tea.Batch(getLighthouseResult(cmd, msg.URL), addUniqueHostname(&m.hostnamesView.list, HostnameItem{hostname: u.Hostname(), loading: true, cmd: cmd}), m.hostnamesView.spinner.Tick)
	case RestartMsg:
		ctx, cancel := context.WithCancel(context.Background())
		cmd := &runningCmd{ctx: ctx, cancel: cancel}

		m.hostnamesView.setLoading(cmd, msg.Index, true)

		return m, tea.Batch(getLighthouseResult(cmd, msg.URL), m.hostnamesView.spinner.Tick)
	case LighthouseMsg:
		if msg.err != nil {
			return m, nil
		}

		m.resultView.report = msg.result.Report
		m.resultView.loaded = true

		cmds = append(cmds, stopLoading(&m.hostnamesView.list, msg.url))

		u, err := url.Parse(msg.url)

		if err != nil {
			return m, nil
		}

		if item, ok := m.hostnamesView.list.SelectedItem().(HostnameItem); ok && item.hostname == u.Hostname() {
			cmds = append(cmds, m.reportsView.InsertReport(msg.result.ReportName))
		}

		return m, tea.Batch(cmds...)
	// case components.ActiveTabMsg:
	// 	m.resultView.SetActiveTab(msg.Label)
	//
	// 	return m, nil
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	if m.showModal {
		*m.modal, cmd = m.modal.Update(msg)
		cmds = append(cmds, cmd)
	} else {
		*m.hostnamesView, cmd = m.hostnamesView.Update(msg)
		cmds = append(cmds, cmd)

		*m.reportsView, cmd = m.reportsView.Update(msg)
		cmds = append(cmds, cmd)

		*m.resultView, cmd = m.resultView.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// View returns main app ui.
func (m Model) View() tea.View {
	var v tea.View

	v.AltScreen = true
	v.MouseMode = tea.MouseModeCellMotion

	if m.error != nil {
		v.SetContent(m.error.Error())

		return v
	}

	if m.showModal {
		v.SetContent(m.modal.View(m.width, m.height))
	} else {
		bordersHeight := 2
		sidePanelView := ""
		resultView := ""

		resultTitle := "Result"

		if m.focused == reportsPane || m.focused == resultPane {
			resultTitle = "[2] Result"
		}

		if m.width > 150 {
			asideWidth := min(m.width/5, 50)

			urlsView := RenderPane("[0] Hostnames", m.hostnamesView.view(asideWidth, m.height-bordersHeight), asideWidth, m.height, m.focused == hostnamesPane)
			reportsView := RenderPane("[1] Reports", m.reportsView.View(m.height-bordersHeight), asideWidth, m.height, m.focused == reportsPane)

			sidePanelView = lipgloss.NewStyle().Height(m.height).Render(lipgloss.JoinHorizontal(lipgloss.Left, urlsView, reportsView))

			resultView = RenderPane(resultTitle, m.resultView.View(m.width-asideWidth*2, m.height), m.width-asideWidth*2, m.height, m.focused == resultPane)
		} else {
			asideWidth := min(m.width/3, 50)
			hPanesCount := 2

			spareSpace := m.height - ((m.height / hPanesCount) * hPanesCount)

			urlsView := RenderPane("[0] Hostnames", m.hostnamesView.view(asideWidth, (m.height/hPanesCount)-bordersHeight), asideWidth, m.height/2, m.focused == hostnamesPane)
			reportsView := RenderPane("[1] Reports", m.reportsView.View((m.height/hPanesCount)+spareSpace-bordersHeight), asideWidth, (m.height/2)+spareSpace, m.focused == reportsPane)

			sidePanelView = lipgloss.NewStyle().Height(m.height).Render(lipgloss.JoinVertical(lipgloss.Top, urlsView, reportsView))
			resultView = RenderPane(resultTitle, m.resultView.View(m.width-asideWidth, m.height), m.width-asideWidth, m.height, m.focused == resultPane)
		}

		v.SetContent(lipgloss.JoinHorizontal(lipgloss.Top, sidePanelView, resultView))
	}

	return v
}

func (m *Model) setActivePane(pane Pane) tea.Cmd {
	m.reportsView.SetActive(m.focused == reportsPane)
	m.hostnamesView.setActive(m.focused == hostnamesPane)

	return func() tea.Msg {
		return ActivePaneMsg{
			pane: pane,
		}
	}
}
