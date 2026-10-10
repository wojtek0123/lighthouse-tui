package ui

import (
	"strings"

	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/wojtek0123/lighthouse-tui/internal/lighthouse"
)

type Tab int

const (
	Overview Tab = iota
	Performance
)

type ResultView struct {
	activeTab Tab

	overviewTab    *OverviewTab
	performanceTab *PerformanceTab

	err    error
	table  table.Model
	audits lighthouse.LighthouseReport
	loaded bool

	hostname   string
	reportName string
}

func NewResultView() *ResultView {
	t := table.New()
	overviewTab := NewOverviewTab()
	performanceTab := NewPerformanceTab()

	return &ResultView{
		loaded:         false,
		table:          t,
		overviewTab:    overviewTab,
		activeTab:      Overview,
		performanceTab: performanceTab,
	}
}

func (r ResultView) Update(msg tea.Msg) (ResultView, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case UrlSelectedMsg:
		r.hostname = msg.hostname
		r.reportName = ""
		r.loaded = false
	case ReportSelectedMsg:
		if r.hostname != "" {
			r.reportName = msg.reportName

			report, err := lighthouse.GetReport(r.hostname, r.reportName)

			if err != nil {
				// return r, nil
			} else {
				r.audits = report
				r.loaded = true
			}

		}
	}

	r.table, cmd = r.table.Update(msg)
	cmds = append(cmds, cmd)

	return r, tea.Batch(cmds...)
}

func (r *ResultView) SetActiveTab(index int) {
	r.activeTab = Tab(index)
}

func (r ResultView) View(width, height int) string {
	if r.loaded {
		tab := ""

		switch r.activeTab {
		case Overview:
			categories := r.audits.Categories

			perf := categories[lighthouse.Performance].Score
			accessibility := categories[lighthouse.Accessibility].Score
			seo := categories[lighthouse.SEO].Score
			agentic := categories[lighthouse.AgenticBrowsing].Score
			practises := categories[lighthouse.BestPractises].Score

			tab = r.overviewTab.View(perf, accessibility, practises, seo, agentic, width)
		case Performance:
			tab = r.performanceTab.View()
		}

		return tab
	} else {
		return lipgloss.Place(width, height-2, lipgloss.Center, lipgloss.Center, lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render(lipgloss.JoinVertical(lipgloss.Center, "No results", "", "Press 'n' to enter url and measure metrics")))
	}
}

func Render(header, content string, width, height int, focus bool) string {
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

func (v *ResultView) SetHostname(hostname string) {
	v.hostname = hostname
}
