package ui

import (
	"context"
	"fmt"
	"time"
	"uuid"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/wojtek0123/lighthouse-tui/internal/report"
	"github.com/wojtek0123/lighthouse-tui/internal/store"
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

// SummariesLoadedMsg is a Bubble Tea message dispatched when summary is loaded.
type SummariesLoadedMsg struct {
	summaries store.Summaries
}

// ActivePaneMsg is a Bubble Tea message dispatched when active pane changed.
type ActivePaneMsg struct {
	pane Pane
}

// Model represents app main type.
type Model struct {
	reportView     *ReportView
	modal          *URLModal
	reportListView *ReportListView

	reportsStore   *store.ReportStore
	summariesStore *store.SummaryStore

	showModal bool
	focused   Pane

	ready  bool
	width  int
	height int

	error error
}

// InitialModel creates initial state of the app.
func InitialModel(dir string) Model {
	reportStore := store.NewReportStore(dir)

	return Model{
		reportListView: NewReportListView(),
		reportView:     NewReportView(reportStore),
		focused:        hostnamesPane,
		modal:          NewURLModal(),
		ready:          false,
		reportsStore:   reportStore,
		summariesStore: store.NewSummaryStore(dir),
	}
}

// Init returns tea.Cmd to keep up to date others views.
func (m Model) Init() tea.Cmd {
	summaries, err := m.summariesStore.List()

	if err != nil {
		return ThrowError(err)
	}

	return func() tea.Msg {
		return SummariesLoadedMsg{
			summaries: summaries,
		}
	}
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
		}
	case CloseModalMsg:
		m.showModal = false
		return m, nil
	case EnteredURLMsg:
		//nolint:gosec // TODO: handle cancellation
		ctx, cancel := context.WithCancel(context.Background())
		cmd := &runningCmd{ctx: ctx, cancel: cancel}

		return m, getLighthouseReport(cmd, msg.URL)
	// case RestartMsg:
	// 	ctx, cancel := context.WithCancel(context.Background())
	// 	cmd := &runningCmd{ctx: ctx, cancel: cancel}
	//
	// 	m.hostnamesView.setLoading(cmd, msg.Index, true)
	//
	// 	return m, tea.Batch(getLighthouseReport(cmd, msg.URL), m.hostnamesView.spinner.Tick)
	case LighthouseMsg:
		if msg.err != nil {
			m.error = msg.err
			return m, nil
		}

		r, err := report.Parse(msg.output)

		if err != nil {
			return m, ThrowError(err)
		}

		createdAt, err := time.Parse("", r.FetchTime)

		if err != nil {
			m.error = msg.err
			return m, nil
		}

		id := uuid.NewV7()
		summary := store.Summary{
			ID:        id,
			URL:       r.FinalURL,
			CreatedAt: createdAt,
			Scores: store.Score{
				Performance:     r.Metrics()[report.PerformanceCategoryID],
				Accessibility:   r.Metrics()[report.AccessibilityCategoryID],
				BestPractices:   r.Metrics()[report.BestPractisesCategoryID],
				SEO:             r.Metrics()[report.SeoCategoryID],
				AgenticBrowsing: r.Metrics()[report.AgenticBrowsingCategoryID],
			},
		}

		err = m.reportsStore.Save(id, msg.output)

		if err != nil {
			return m, ThrowError(err)
		}

		err = m.summariesStore.Save(&summary)

		if err != nil {
			return m, ThrowError(err)
		}

		m.reportView.SetReport(r)

		return m, tea.Batch(cmds...)
	case tea.WindowSizeMsg:
		m.setWindowSize(msg.Width, msg.Height)
	}

	if m.showModal {
		*m.modal, cmd = m.modal.Update(msg)
		cmds = append(cmds, cmd)
	} else {
		*m.reportListView, cmd = m.reportListView.Update(msg)
		cmds = append(cmds, cmd)

		*m.reportView, cmd = m.reportView.Update(msg)
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
		// TODO: user cannot leave this screen. It should be a modal.
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
	// m.reportsView.SetActive(m.focused == reportsPane)
	// m.hostnamesView.setActive(m.focused == hostnamesPane)

	return func() tea.Msg {
		return ActivePaneMsg{
			pane: pane,
		}
	}
}
