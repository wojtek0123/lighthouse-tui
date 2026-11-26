package ui

import (
	"context"
	"fmt"
	"net/url"

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

var (
	activeBorderColor   = lipgloss.Color("62")  // Indigo/Purple.
	inactiveBorderColor = lipgloss.Color("240") // Dark Gray.
	// activeTitleColor    = lipgloss.Color("212") // Pink.
	inactiveTitleColor = lipgloss.Color("240") // Dark Gray.
)

type runningCmd struct {
	ctx    context.Context
	cancel context.CancelFunc
}

// Pane indicates active pane.
type Pane int

// ReportViewSizeMsg is a Bubble Tea message dispatched when result view size is changed.
type ReportViewSizeMsg struct {
	width, height int
}

// ActivePaneMsg is a Bubble Tea message dispatched when active pane changed.
type ActivePaneMsg struct {
	pane Pane
}

// Model represents app main type.
type Model struct {
	hostnamesView  *HostnamesView
	resultView     *ReportView
	reportsView    *ReportsView
	modal          *URLModal
	reportListView *ReportListView

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

	hostnamesView := newHostnamesView(true)
	resultView := NewReportView()
	reportsView := NewReportsView(false)
	modal := NewURLModal()

	if len(urls) > 0 {
		reportsView.GetReports(urls[0])
		resultView.SetHostname(urls[0])
	}

	return Model{
		reportListView: NewReportListView(),
		hostnamesView:  hostnamesView,
		resultView:     resultView,
		reportsView:    reportsView,
		focused:        hostnamesPane,
		modal:          modal,
		ready:          false,
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

		return m, tea.Batch(getLighthouseReport(cmd, msg.URL), addUniqueHostname(&m.hostnamesView.list, HostnameItem{hostname: u.Hostname(), loading: true, cmd: cmd}), m.hostnamesView.spinner.Tick)
	case RestartMsg:
		ctx, cancel := context.WithCancel(context.Background())
		cmd := &runningCmd{ctx: ctx, cancel: cancel}

		m.hostnamesView.setLoading(cmd, msg.Index, true)

		return m, tea.Batch(getLighthouseReport(cmd, msg.URL), m.hostnamesView.spinner.Tick)
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
		m.setWindowSize(msg.Width, msg.Height)
	}

	*m.reportListView, cmd = m.reportListView.Update(msg)
	cmds = append(cmds, cmd)

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

	if m.width < minTerminalWidth || m.height < minTerminalHeight {
		v.SetContent(fmt.Sprintf("Terminal is too small %dx%d", m.width, m.height))
		return v
	}

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
		v.SetContent(m.reportListView.View())
	}

	return v
}

func (m *Model) setWindowSize(w, h int) {
	m.ready = true
	m.width = w
	m.height = h
	m.reportListView.SetSize(w, h)
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
