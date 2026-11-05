package ui

import (
	"context"
	"net/url"
	"strconv"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/wojtek0123/lighthouse-tui/internal/lighthouse"
)

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
	urlsView    *HostnamesView
	resultView  *ResultView
	reportsView *ReportsView
	modal       *URLModal

	showModal bool
	focused   Pane

	ready  bool
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
		urlsView:    urlsView,
		resultView:  resultView,
		reportsView: reportsView,
		focused:     hostnamesPane,
		modal:       modal,
		ready:       false,
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

		return m, tea.Batch(getLighthouseResult(cmd, msg.URL), addUniqueHostname(&m.urlsView.list, HostnameItem{hostname: u.Hostname(), loading: true, cmd: cmd}), m.urlsView.spinner.Tick)
	case RestartMsg:
		ctx, cancel := context.WithCancel(context.Background())
		cmd := &runningCmd{ctx: ctx, cancel: cancel}

		m.urlsView.setLoading(cmd, msg.Index, true)

		return m, tea.Batch(getLighthouseResult(cmd, msg.URL), m.urlsView.spinner.Tick)
	case LighthouseMsg:
		if msg.err != nil {
			return m, nil
		}

		m.resultView.loaded = true

		cmds = append(cmds, stopLoading(&m.urlsView.list, msg.url))

		u, err := url.Parse(msg.url)

		if err != nil {
			return m, nil
		}

		if item, ok := m.urlsView.list.SelectedItem().(HostnameItem); ok && item.hostname == u.Hostname() {
			cmds = append(cmds, m.reportsView.InsertReport(msg.result.ReportName))
		}

		return m, tea.Batch(cmds...)
	case tea.WindowSizeMsg:
		m.ready = true
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

// View returns main app ui.
func (m Model) View() tea.View {
	var v tea.View

	v.AltScreen = true

	if m.error != nil {
		v = tea.NewView(m.error.Error())

		return v
	}

	if m.showModal {
		v = m.modal.View(m.width, m.height)
	} else {
		v = tea.NewView(m.renderView())
	}

	return v
}

func (m Model) renderView() string {
	if !m.ready {
		return "Initializing lighthouse tui..."
	}

	if m.width < minTerminalWidth || m.height < minTerminalHeight {
		return lipgloss.JoinHorizontal(lipgloss.Center, "Window is to small: ", strconv.Itoa(m.width), "x", strconv.Itoa(m.height))
	}

	// if m.width > 100 {
	// 	return m.renderVerticalLayout()
	// }

	return m.renderHorizontalLayout()
}

func (m Model) renderHorizontalLayout() string {
	// bordersHeight := 2

	// height / 5
	// urls 1/5
	// reports 1/5
	// result 3/5

	urlsView := RenderPane("[0] Hostnames", m.urlsView.view(m.width, m.height/5), m.width, m.height/5, m.focused == hostnamesPane)
	reportsView := RenderPane("[1] Reports", m.reportsView.View(m.height/5), m.width, m.height/5, m.focused == reportsPane)
	resultView := RenderPane("[2] Result", m.resultView.View(m.width, (m.height/5)*3), m.width, (m.height/5)*3, m.focused == resultPane)

	return lipgloss.JoinVertical(lipgloss.Top, urlsView, reportsView, resultView)
}

func (m Model) renderVerticalLayout() string {
	bordersHeight := 2

	if m.width > 150 {
		// three column layout
		asideWidth := min(m.width/5, 50)

		urlsView := RenderPane("[0] Hostnames", m.urlsView.view(asideWidth, m.height-bordersHeight), asideWidth, m.height, m.focused == hostnamesPane)
		reportsView := RenderPane("[1] Reports", m.reportsView.View(m.height-bordersHeight), asideWidth, m.height, m.focused == reportsPane)

		sidePanelView := lipgloss.NewStyle().Height(m.height).Render(lipgloss.JoinHorizontal(lipgloss.Left, urlsView, reportsView))

		resultView := RenderPane("[2] Result", m.resultView.View(m.width-asideWidth*2, m.height), m.width-asideWidth*2, m.height, m.focused == resultPane)
		return lipgloss.JoinHorizontal(lipgloss.Top, sidePanelView, resultView)
	} else {
		// two column layout
		asideWidth := min(m.width/3, 50)
		hPanesCount := 2
		spareSpace := m.height - ((m.height / hPanesCount) * hPanesCount)

		urlsView := RenderPane("[0] Hostnames", m.urlsView.view(asideWidth, (m.height/hPanesCount)-bordersHeight), asideWidth, m.height/2, m.focused == hostnamesPane)
		reportsView := RenderPane("[1] Reports", m.reportsView.View((m.height/hPanesCount)+spareSpace-bordersHeight), asideWidth, (m.height/2)+spareSpace, m.focused == reportsPane)

		sidePanelView := lipgloss.NewStyle().Height(m.height).Render(lipgloss.JoinVertical(lipgloss.Top, urlsView, reportsView))
		resultView := RenderPane("[2] Result", m.resultView.View(m.width-asideWidth, m.height), m.width-asideWidth, m.height, m.focused == resultPane)

		return lipgloss.JoinHorizontal(lipgloss.Top, sidePanelView, resultView)
	}
}

const (
	minTerminalWidth  int = 50
	minTerminalHeight int = 50
)

func (m *Model) setActivePane(pane Pane) tea.Cmd {
	m.reportsView.SetActive(m.focused == reportsPane)
	m.urlsView.setActive(m.focused == hostnamesPane)

	return func() tea.Msg {
		return ActivePaneMsg{
			pane: pane,
		}
	}
}
