package ui

import (
	tea "charm.land/bubbletea/v2"
	"github.com/wojtek0123/lighthouse-tui/internal/ui/components"
)

type ReportListView struct {
	table *components.Table
}

var headers = []components.TableHeader{
	{Title: "hostname", Width: 20},
	{Title: "Perf", Width: 10},
	{Title: "A11Y", Width: 10},
	{Title: "BP", Width: 10},
	{Title: "SEO", Width: 10},
}

var rows = []components.TableRow{
	{Values: []string{"wojciechpietraszuk.vercel"}, Children: []components.TableChildRow{
		{"/", "99", "99", "100", "98"},
	}},
	{Values: []string{"github.com 1"}, Children: []components.TableChildRow{
		{"/test", "99", "99", "100", "98"},
	}},
	{Values: []string{"github.com 2"}, Children: []components.TableChildRow{
		{"/test", "99", "99", "100", "98"},
	}},
	{Values: []string{"github.com 3"}, Children: []components.TableChildRow{
		{"/test", "99", "99", "100", "98"},
		{"/test", "99", "99", "100", "98"},
	}},
	{Values: []string{"github.com 4"}, Children: []components.TableChildRow{}},
	{Values: []string{"github.com 5"}, Children: []components.TableChildRow{}},
	{Values: []string{"github.com 6"}, Children: []components.TableChildRow{}},
	{Values: []string{"github.com 7"}, Children: []components.TableChildRow{}},
	{Values: []string{"github.com 8"}, Children: []components.TableChildRow{}},
	{Values: []string{"github.com 9"}, Children: []components.TableChildRow{}},
	{Values: []string{"github.com 10"}, Children: []components.TableChildRow{}},
	{Values: []string{"github.com 11"}, Children: []components.TableChildRow{
		{"/test", "99", "99", "100", "98"},
		{"/test1", "99", "99", "100", "98"},
		{"/test2", "99", "99", "100", "98"},
		{"/test3", "99", "99", "100", "98"},
	}},
	{Values: []string{"github.com 12"}, Children: []components.TableChildRow{}},
	{Values: []string{"github.com 13"}, Children: []components.TableChildRow{}},
	{Values: []string{"github.com 14"}, Children: []components.TableChildRow{}},
	{Values: []string{"github.com 15"}, Children: []components.TableChildRow{}},
	{Values: []string{"github.com 16"}, Children: []components.TableChildRow{}},
	{Values: []string{"github.com 17"}, Children: []components.TableChildRow{}},
	{Values: []string{"github.com 18"}, Children: []components.TableChildRow{
		{"/test1", "99", "99", "100", "98"},
		{"/test2", "99", "99", "100", "98"},
		{"/test3", "99", "99", "100", "98"},
	}},
	{Values: []string{"github.com 20"}, Children: []components.TableChildRow{}},
	{Values: []string{"github.com 21"}, Children: []components.TableChildRow{
		{"/test", "99", "99", "100", "98"},
	}},
	{Values: []string{"github.com 22"}, Children: []components.TableChildRow{
		{"/test", "99", "99", "100", "98"},
		{"/test", "99", "99", "100", "98"},
	}},
	{Values: []string{"github.com 23"}, Children: []components.TableChildRow{
		{"/test", "99", "99", "100", "98"},
		{"/test1", "99", "99", "100", "98"},
		{"/test2", "99", "99", "100", "98"},
		{"/test3", "99", "99", "100", "98"},
		{"/test", "99", "99", "100", "98"},
		{"/test1", "99", "99", "100", "98"},
		{"/test2", "99", "99", "100", "98"},
		{"/test3", "99", "99", "100", "98"},
		{"/test", "99", "99", "100", "98"},
		{"/test1", "99", "99", "100", "98"},
		{"/test2", "99", "99", "100", "98"},
		{"/test3", "99", "99", "100", "98"},
		{"/test", "99", "99", "100", "98"},
		{"/test1", "99", "99", "100", "98"},
		{"/test2", "99", "99", "100", "98"},
		{"/test3", "99", "99", "100", "98"},
		{"/test", "99", "99", "100", "98"},
		{"/test1", "99", "99", "100", "98"},
		{"/test2", "99", "99", "100", "98"},
		{"/test3", "99", "99", "100", "98"},
		{"/test", "99", "99", "100", "98"},
		{"/test1", "99", "99", "100", "98"},
		{"/test2", "99", "99", "100", "98"},
		{"/test3", "99", "99", "100", "98"},
		{"/test", "99", "99", "100", "98"},
		{"/test1", "99", "99", "100", "98"},
		{"/test2", "99", "99", "100", "98"},
		{"/test3", "99", "99", "100", "98"},
		{"/test", "99", "99", "100", "98"},
		{"/test1", "99", "99", "100", "98"},
		{"/test2", "99", "99", "100", "98"},
		{"/test3", "99", "99", "100", "98"},
		{"/test", "99", "99", "100", "98"},
		{"/test1", "99", "99", "100", "98"},
		{"/test2", "99", "99", "100", "98"},
		{"/test3", "99", "99", "100", "98"},
		{"/test", "99", "99", "100", "98"},
		{"/test1", "99", "99", "100", "98"},
		{"/test2", "99", "99", "100", "98"},
		{"/test3", "99", "99", "100", "98"},
		{"/test", "99", "99", "100", "98"},
		{"/test1", "99", "99", "100", "98"},
		{"/test2", "99", "99", "100", "98"},
		{"/test3", "99", "99", "100", "98"},
	}},
	{Values: []string{"github.com 24"}, Children: []components.TableChildRow{
		{"/test1", "99", "99", "100", "98"},
	}},
	{Values: []string{"github.com 25"}, Children: []components.TableChildRow{
		{"/test1", "99", "99", "100", "98"},
		{"/test1", "99", "99", "100", "98"},
	}},
}

func NewReportListView() *ReportListView {
	table := components.NewTable()

	table.SetHeaders(headers)
	table.SetRows(rows)

	return &ReportListView{
		table: table,
	}
}

func (v ReportListView) Update(msg tea.Msg) (ReportListView, tea.Cmd) {
	var cmd tea.Cmd
	v.table, cmd = v.table.Update(msg)

	return v, cmd
}

func (v *ReportListView) UpdateHeaders(width int) {
	var headers = []components.TableHeader{
		{Title: "hostname", Width: width - 40},
		{Title: "Perf", Width: 10},
		{Title: "A11Y", Width: 10},
		{Title: "BP", Width: 10},
		{Title: "SEO", Width: 10},
	}

	v.table.SetHeaders(headers)
}

func (v ReportListView) View() string {
	return v.table.View()
}

func (v *ReportListView) SetSize(w, h int) {
	v.table.SetWidth(w)
	v.table.SetHeight(h)
	v.UpdateHeaders(w)
}
