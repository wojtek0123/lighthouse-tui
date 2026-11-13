package ui

import (
	"context"
	"net/url"
	"strconv"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/wojtek0123/lighthouse-tui/internal/lighthouse"
)

const (
	minTerminalWidth  int = 50
	minTerminalHeight int = 20
)

const (
	hostnamesPane Pane = iota
	reportsPane
	resultPane
)

// Pane indicates active pane.
type Pane int

// ActivePaneMsg is a Bubble Tea message dispatched when active pane changed.
type ActivePaneMsg struct {
	pane Pane
}

type Layout string

const (
	// DualColumn: 2 columns total (Sidebar with stacked lists + Main result view).
	DualColumn Layout = "dual-column"

	// TripleColumn: 3 columns total (Hostname column + Report column + Result view column).
	TripleColumn Layout = "triple-column"

	// SingleColumn: 1 column total (All elements stacked vertically for mobile/narrow viewports).
	SingleColumn Layout = "single-column"
)

type PaneSize struct {
	reportListWidth,
	reportListHeight,
	hostnameListWidth,
	hostnameListHeight,
	resultWidth,
	resultHeight int
}

// Model represents app main type.
type Model struct {
	hostnamesView *HostnamesView
	resultView    *ResultView
	reportsView   *ReportsView
	modal         *URLModal

	showModal bool
	focused   Pane

	ready  bool
	width  int
	height int

	layout   Layout
	paneSize *PaneSize

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

	hostnamesView := newHostnamesView(true)
	resultView := NewResultView()
	reportsView := NewReportsView(false)
	modal := NewURLModal()

	if len(urls) > 0 {
		reportsView.GetReports(urls[0])
		resultView.SetHostname(urls[0])
	}

	return Model{
		hostnamesView: hostnamesView,
		resultView:    resultView,
		reportsView:   reportsView,
		focused:       hostnamesPane,
		modal:         modal,
		ready:         false,
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

		m.resultView.SetReport(msg.result.Report)
		cmds = append(cmds, stopLoading(&m.hostnamesView.list, msg.url))

		u, err := url.Parse(msg.url)

		if err != nil {
			return m, nil
		}

		if item, ok := m.hostnamesView.list.SelectedItem().(HostnameItem); ok && item.hostname == u.Hostname() {
			cmds = append(cmds, m.reportsView.InsertReport(msg.result.ReportName))
		}

		return m, tea.Batch(cmds...)
	case tea.WindowSizeMsg:
		m.ready = true
		m.width = msg.Width
		m.height = msg.Height
		m.calculatePanesSize()

		return m, m.updateResultSize()
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

	if !m.ready {
		v.SetContent("Initializing lighthouse tui...")

		return v
	}

	v.AltScreen = true
	v.MouseMode = tea.MouseModeCellMotion

	if m.error != nil {
		v.SetContent(m.error.Error())

		return v
	}

	if m.showModal {
		v.SetContent(m.modal.View(m.width, m.height))
	} else {
		v.SetContent(m.renderLayout())
	}

	return v
}

func (m Model) renderLayout() string {
	if m.width < minTerminalWidth || m.height < minTerminalHeight {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, lipgloss.JoinVertical(lipgloss.Center, "Window is too small", lipgloss.JoinHorizontal(lipgloss.Center, strconv.Itoa(m.width), "x", strconv.Itoa(m.height))))
	}

	hostnameListView := RenderPane("[0] Hostnames", m.hostnamesView.view(m.paneSize.hostnameListWidth, m.paneSize.hostnameListHeight), m.paneSize.hostnameListWidth, m.paneSize.hostnameListHeight, m.focused == hostnamesPane)
	reportsView := RenderPane("[1] Reports", m.reportsView.View(m.paneSize.reportListHeight), m.paneSize.reportListWidth, m.paneSize.reportListHeight, m.focused == reportsPane)
	resultView := RenderPane("[2] Result", m.resultView.View(), m.paneSize.resultWidth, m.paneSize.resultHeight, m.focused == resultPane)

	switch m.layout {
	case SingleColumn:
		return lipgloss.JoinVertical(lipgloss.Top, hostnameListView, reportsView, resultView)
	case DualColumn:
		sidePanelView := lipgloss.NewStyle().Height(m.height).Render(lipgloss.JoinVertical(lipgloss.Top, hostnameListView, reportsView))

		return lipgloss.JoinHorizontal(lipgloss.Top, sidePanelView, resultView)
	case TripleColumn:
		return lipgloss.JoinHorizontal(lipgloss.Top, hostnameListView, reportsView, resultView)
	default:
		return ""
	}
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

func (m *Model) calculatePanesSize() {
	if m.width > 150 {
		m.layout = TripleColumn

		firstColumnWidth := m.width / 5
		secondColumnWidth := m.width / 5
		thirdColumnWidth := m.width - firstColumnWidth - secondColumnWidth

		m.paneSize = &PaneSize{
			hostnameListWidth:  firstColumnWidth,
			hostnameListHeight: m.height - 2,
			reportListWidth:    secondColumnWidth,
			reportListHeight:   m.height - 2,
			resultWidth:        thirdColumnWidth,
			resultHeight:       m.height - 2,
		}
	} else if m.width > 100 {
		m.layout = DualColumn

		leftPaneWidth := m.width / 3
		leftTopPaneHeight := (m.height / 2) - 2
		leftBottomPaneHeight := m.height - leftTopPaneHeight - 4
		rightPaneWidth := m.width - leftPaneWidth

		m.paneSize = &PaneSize{
			hostnameListWidth:  leftPaneWidth,
			hostnameListHeight: leftTopPaneHeight,
			reportListWidth:    leftPaneWidth,
			reportListHeight:   leftBottomPaneHeight,
			resultWidth:        rightPaneWidth,
			resultHeight:       m.height - 2,
		}
	} else {
		m.layout = SingleColumn

		firstRowHeight := m.height / 5
		secondRowHeight := m.height / 5
		thirstRowHeight := m.height - firstRowHeight - secondRowHeight

		m.paneSize = &PaneSize{
			hostnameListWidth:  m.width,
			hostnameListHeight: firstRowHeight,
			reportListWidth:    m.width,
			reportListHeight:   secondRowHeight,
			resultWidth:        m.width,
			resultHeight:       thirstRowHeight,
		}
	}
}

type ResultSizeMsg struct {
	width, height int
}

func (m *Model) updateResultSize() tea.Cmd {
	return func() tea.Msg {
		return ResultSizeMsg{
			width:  m.paneSize.resultWidth,
			height: m.paneSize.resultHeight,
		}
	}
}
