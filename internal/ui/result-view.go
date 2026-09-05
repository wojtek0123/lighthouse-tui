package ui

import (
	"fmt"
	"strconv"

	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/wojtek0123/lighthouse-tui/internal/lighthouse"
)

var sectionHeader = lipgloss.NewStyle().Bold(true)

type ResultView struct {
	table  table.Model
	audits lighthouse.LighthouseResult
	loaded bool

	width int
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

func (r ResultView) View() tea.View {
	if r.loaded {

		fcp := r.audits.Audits["first-contentful-paint"]
		lcp := r.audits.Audits["largest-contentful-paint"]
		cls := r.audits.Audits["cumulative-layout-shift"]
		tbt := r.audits.Audits["total-blocking-time"]
		si := r.audits.Audits["speed-index"]

		tableWidth := r.width - 10
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

		return tea.NewView(r.table.View())

		// return tea.NewView(lipgloss.JoinVertical(lipgloss.Center, sectionHeader.Render("Audits (Performance)"), fmt.Sprintf("%s %f %s", fcp.Title, fcp.NumericValue, fcp.Unit)))
	} else {
		return tea.NewView(fmt.Sprintf("RESULT %v", r.loaded))
	}
}
