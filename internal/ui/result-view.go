package ui

import (
	"strings"

	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/wojtek0123/lighthouse-tui/internal/lighthouse"
	"github.com/wojtek0123/lighthouse-tui/internal/ui/components"
)

// Tab represents available tabs in result view.
type Tab int

const (
	overview Tab = iota
	performance
)

// ResultView represents the view model for displaying report results,
// managing sub-tabs, and tracking interactive UI state in the TUI.
type ResultView struct {
	activeTab Tab
	active    bool

	overviewTab    *OverviewTab
	performanceTab *PerformanceTab

	err    error
	table  table.Model
	audits *lighthouse.Report
	loaded bool

	hostname   string
	reportName string

	tabs *components.Tabs
}

// NewResultView creates result view.
func NewResultView() *ResultView {
	t := table.New()
	overviewTab := NewOverviewTab()
	performanceTab := NewPerformanceTab()
	tabs := components.NewTabs()
	tabs.SetTabs([]string{"Overview", "Performance"})

	return &ResultView{
		loaded:         false,
		table:          t,
		overviewTab:    overviewTab,
		activeTab:      overview,
		performanceTab: performanceTab,
		tabs:           tabs,
	}
}

// Update updates result view state.
func (r ResultView) Update(msg tea.Msg) (ResultView, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case ActivePaneMsg:
		active := msg.pane == resultPane

		r.tabs.SetFocus(active)
		r.active = active
		return r, nil
	case HostnameSelectedMsg:
		r.hostname = msg.hostname
		r.reportName = ""
		r.loaded = false
	case ReportSelectedMsg:
		if r.hostname != "" {
			r.reportName = msg.reportName

			report, err := lighthouse.GetReport(r.hostname, r.reportName)

			if err != nil {
				r.err = err
			} else {
				r.audits = report
				r.loaded = true
			}
		}
	}

	if r.active {
		r.table, cmd = r.table.Update(msg)
		cmds = append(cmds, cmd)

		*r.tabs, cmd = r.tabs.Update(msg)
		cmds = append(cmds, cmd)
	}

	return r, tea.Batch(cmds...)
}

// SetActiveTab sets which tab is currently active.
func (r *ResultView) SetActiveTab(index int) {
	r.activeTab = Tab(index)
}

// View returns ui of ResultView as string.
func (r ResultView) View(width, height int) string {
	if r.loaded {
		tab := ""

		switch r.activeTab {
		case overview:
			categories := r.audits.Categories

			perf := categories[lighthouse.Performance].Score
			accessibility := categories[lighthouse.Accessibility].Score
			seo := categories[lighthouse.SEO].Score
			agentic := categories[lighthouse.AgenticBrowsing].Score
			practises := categories[lighthouse.BestPractises].Score

			tab = r.overviewTab.View(perf, accessibility, practises, seo, agentic, width)
		case performance:
			tab = r.performanceTab.View()
		}

		return lipgloss.NewStyle().Padding(0, 1).Render(lipgloss.JoinVertical(lipgloss.Top, r.tabs.View(), " ", tab))
	}

	return lipgloss.Place(width, height-2, lipgloss.Center, lipgloss.Center, lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render(lipgloss.JoinVertical(lipgloss.Center, "No results", "", "Press 'n' to enter url and measure metrics")))
}

// RenderPane renders pane with borders. Border color change based on focus state.
func RenderPane(header, content string, width, height int, focus bool) string {
	borderColor := inactiveBorderColor
	titleColor := inactiveTitleColor

	if focus {
		borderColor = activeBorderColor
		titleColor = lipgloss.Color("170")
	}

	borderStyle := lipgloss.NewStyle().Foreground(borderColor)
	border := lipgloss.RoundedBorder()
	renderedTitle := lipgloss.NewStyle().Padding(0, 1).Foreground(titleColor).Render(header)

	leftDashCount := 3
	corners := 2

	spareSpace := width - lipgloss.Width(renderedTitle) - leftDashCount - corners

	topLeftBorder := borderStyle.Render(border.TopLeft)
	leftDash := strings.Repeat(borderStyle.Render(border.Top), leftDashCount)
	rightDash := strings.Repeat(borderStyle.Render(border.Top), max(spareSpace, 5))
	topRightBorder := borderStyle.Render(border.TopRight)

	topBorder := topLeftBorder + leftDash + renderedTitle + rightDash + topRightBorder

	body := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Width(width).Height(height - 1).BorderTop(false).BorderForeground(borderColor).Render(content)

	return lipgloss.NewStyle().Width(width).Height(height - 1).Render(lipgloss.JoinVertical(lipgloss.Left, topBorder, body))
}

// SetHostname sets which hostname is focus from HostnamesView.
func (r *ResultView) SetHostname(hostname string) {
	r.hostname = hostname
}
