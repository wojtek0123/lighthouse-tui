package ui

import (
	"strconv"
	"strings"

	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/wojtek0123/lighthouse-tui/internal/lighthouse"
)

type Tab int

const (
	Overview Tab = iota
)

type ResultView struct {
	Tabs       []string
	TabContent []string
	activeTab  Tab

	overviewTab *OverviewTab

	err    error
	table  table.Model
	audits lighthouse.LighthouseReport
	loaded bool
}

func NewResultView() *ResultView {
	t := table.New()
	overviewTab := NewOverviewTab()

	return &ResultView{
		loaded:      false,
		table:       t,
		overviewTab: overviewTab,
		activeTab:   Overview,
	}
}

func (r ResultView) Update(msg tea.Msg) (ResultView, tea.Cmd) {
	var cmd tea.Cmd

	r.table, cmd = r.table.Update(msg)
	return r, cmd
}

func (r ResultView) View(width, height int) string {
	if r.loaded {

		fcp := r.audits.Audits["first-contentful-paint"]
		lcp := r.audits.Audits["largest-contentful-paint"]
		cls := r.audits.Audits["cumulative-layout-shift"]
		tbt := r.audits.Audits["total-blocking-time"]
		si := r.audits.Audits["speed-index"]

		tableWidth := width - 10
		columnWidth := tableWidth / 3

		columns := []table.Column{
			{Title: "Name", Width: columnWidth},
			{Title: "Score", Width: columnWidth},
			{Title: "Unit", Width: columnWidth},
		}

		rows := []table.Row{
			{fcp.Title, strconv.FormatFloat(fcp.NumericValue, 'f', -1, 64), fcp.Unit},
			{lcp.Title, strconv.FormatFloat(lcp.NumericValue, 'f', -1, 64), lcp.Unit},
			{cls.Title, strconv.FormatFloat(cls.NumericValue, 'f', -1, 64), cls.Unit},
			{tbt.Title, strconv.FormatFloat(tbt.NumericValue, 'f', -1, 64), tbt.Unit},
			{si.Title, strconv.FormatFloat(si.NumericValue, 'f', -1, 64), si.Unit},
		}

		// s := table.DefaultStyles()
		r.table.SetColumns(columns)
		r.table.SetRows(rows)
		r.table.SetWidth(tableWidth)
		r.table.SetHeight(height / 2)
		// r.table.SetStyles(s)

		tabs := []string{"Overview", "Performance"}
		tabsView := lipgloss.JoinHorizontal(lipgloss.Left, tabs...)

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
		}

		return lipgloss.JoinVertical(lipgloss.Top, tabsView, "", tab)
	} else {
		return lipgloss.Place(width, height-2, lipgloss.Center, lipgloss.Center, lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render(lipgloss.JoinVertical(lipgloss.Center, "No results", "", "Press 'n' to enter url and measure metrics")))
	}
}

func Render(title, content string, width, height int, focus bool) string {
	borderColor := inactiveBorderColor
	// titleColor := inactiveTitleColor

	if focus {
		borderColor = activeBorderColor
		// titleColor := lipgloss.Color(lipgloss.White)
	}

	borderStyle := lipgloss.NewStyle().Foreground(borderColor)
	border := lipgloss.RoundedBorder()
	renderedTitle := lipgloss.NewStyle().Padding(0, 1).Render(title)

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
