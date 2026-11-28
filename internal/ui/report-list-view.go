package ui

import (
	"strconv"
	"uuid"

	tea "charm.land/bubbletea/v2"
	"github.com/wojtek0123/lighthouse-tui/internal/store"
	"github.com/wojtek0123/lighthouse-tui/internal/ui/components"
)

// SelectReportMsg is a Bubble tea message dispatched when report is selected.
type SelectReportMsg struct {
	ID uuid.UUID
}

// ReportListView represents state of the view.
type ReportListView struct {
	summariesByHostname map[string]store.Summaries
	table               *components.Table
}

// NewReportListView creates new instance of ReportListView.
func NewReportListView() *ReportListView {
	table := components.NewTable()

	// table.SetHeaders(headers)
	// table.SetRows(rows)

	return &ReportListView{
		table: table,
	}
}

// Update states of the ReportListView.
func (v ReportListView) Update(msg tea.Msg) (ReportListView, tea.Cmd) {
	//nolint:gocritic
	switch msg := msg.(type) {
	case SummariesLoadedMsg:
		v.summariesByHostname = msg.summaries.GroupByHostname()
		v.updateTableRows(msg.summaries.GroupByHostname())
	}

	var cmd tea.Cmd
	v.table, cmd = v.table.Update(msg)

	return v, cmd
}

func (v *ReportListView) updateTableRows(summariesByHostname map[string]store.Summaries) {
	rows := make([]components.TableRow, 0, len(summariesByHostname))

	for hostname, summaries := range summariesByHostname {
		summaries.Sort()

		row := components.TableRow{
			Values: []string{hostname},
		}

		children := make([]components.TableChildRow, 0, len(summaries))

		for _, summary := range summaries {
			children = append(children, components.TableChildRow{
				summary.Path(),
				strconv.Itoa(summary.Scores.Performance),
				strconv.Itoa(summary.Scores.Accessibility),
				strconv.Itoa(summary.Scores.BestPractices),
				strconv.Itoa(summary.Scores.SEO),
			})
		}

		row.Children = children
		rows = append(rows, row)
	}

	v.table.SetRows(rows)
}

func (v *ReportListView) updateTableHeaders(width int) {
	var headers = []components.TableHeader{
		{Title: "Host / Page", Width: width - 70},
		{Title: "Perf", Width: 10},
		{Title: "A11Y", Width: 10},
		{Title: "BP", Width: 10},
		{Title: "SEO", Width: 10},
		{Title: "Date", Width: 30},
	}

	v.table.SetHeaders(headers)
}

// View renders ReportListView layout.
func (v ReportListView) View() string {
	return v.table.View()
}

// SetSize sets size of report list view instance.
func (v *ReportListView) SetSize(w, h int) {
	v.table.SetWidth(w)
	v.table.SetHeight(h)
	v.updateTableHeaders(w)
}
