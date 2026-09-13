package ui

import (
	"strconv"

	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/wojtek0123/lighthouse-tui/internal/lighthouse"
)

var sectionHeader = lipgloss.NewStyle().Bold(true)

type ResultView struct {
	Tabs       []string
	TabContent []string
	activeTab  int

	table  table.Model
	audits lighthouse.LighthouseResult
	loaded bool
}

func NewResultView() *ResultView {
	t := table.New()

	return &ResultView{
		loaded: false,
		table:  t,
	}
}

func (r ResultView) Update(msg tea.Msg) (ResultView, tea.Cmd) {
	var cmd tea.Cmd

	r.table, cmd = r.table.Update(msg)
	return r, cmd
}

func (r ResultView) View(width, height int) tea.View {
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
		r.table.SetHeight(100)
		// r.table.SetStyles(s)

		tabs := []string{"Overview", "Performance"}
		tabsView := lipgloss.JoinHorizontal(lipgloss.Left, tabs...)

		return tea.NewView(lipgloss.JoinVertical(lipgloss.Top, tabsView, ""))
	} else {
		return tea.NewView(lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render(lipgloss.JoinVertical(lipgloss.Center, "No results", "", "Press 'n' to enter url and measure metrics"))))
	}
}
