package ui

import (
	tea "charm.land/bubbletea/v2"
	"github.com/wojtek0123/lighthouse-tui/internal/ui/components"
)

// ReportListView represents state of the view.
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
		{"/test-asdf", "99", "99", "100", "98"},
	}},
	{Values: []string{"github.com 2"}, Children: []components.TableChildRow{
		{"/test112j", "99", "99", "100", "98"},
	}},
	{Values: []string{"github.com 3"}, Children: []components.TableChildRow{
		{"/test1", "99", "99", "100", "98"},
		{"/test2", "99", "99", "100", "98"},
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
		{"/test112", "99", "99", "100", "98"},
		{"/test29023", "99", "99", "100", "98"},
		{"/test-test", "99", "99", "100", "98"},
	}},
	{Values: []string{"github.com 20"}, Children: []components.TableChildRow{}},
	{Values: []string{"github.com 21"}, Children: []components.TableChildRow{
		{"/testasdfldasffkldjaslafjksdjklafdsjkl", "99", "99", "100", "98"},
	}},
	{Values: []string{"github.com 22"}, Children: []components.TableChildRow{
		{"/testdasfafdssafdsdfa", "99", "99", "100", "98"},
		{"/testasdflkj", "99", "99", "100", "98"},
	}},
	{Values: []string{"github.com 23"}, Children: []components.TableChildRow{
		{"/test12333", "99", "99", "100", "98"},
		{"/test1111", "99", "99", "100", "98"},
		{"/test22", "99", "99", "100", "98"},
		{"/test33j", "99", "99", "100", "98"},
		{"/testasdf", "99", "99", "100", "98"},
		{"/test11", "99", "99", "100", "98"},
		{"/test22", "99", "99", "100", "98"},
		{"/test3443", "99", "99", "100", "98"},
		{"/test1243", "99", "99", "100", "98"},
		{"/test111221", "99", "99", "100", "98"},
		{"/test21221", "99", "99", "100", "98"},
		{"/test301", "99", "99", "100", "98"},
		{"/test90ldask", "99", "99", "100", "98"},
		{"/test190ldas", "99", "99", "100", "98"},
		{"/test2asdf", "99", "99", "100", "98"},
		{"/test3mnlm", "99", "99", "100", "98"},
		{"/testmasdmsda", "99", "99", "100", "98"},
		{"/test1lkl", "99", "99", "100", "98"},
		{"/test2dj", "99", "99", "100", "98"},
		{"/test31", "99", "99", "100", "98"},
		{"/testlmlml", "99", "99", "100", "98"},
		{"/test11", "99", "99", "100", "98"},
		{"/test21", "99", "99", "100", "98"},
		{"/test31", "99", "99", "100", "98"},
		{"/testlmlmml", "99", "99", "100", "98"},
		{"/test1lmlmlm", "99", "99", "100", "98"},
		{"/test2mlasdldksaffd", "99", "99", "100", "98"},
		{"/test3asdlkfjlasdfkj", "99", "99", "100", "98"},
		{"/testdasfldfslj", "99", "99", "100", "98"},
		{"/test1asdfljkdfjkls", "99", "99", "100", "98"},
		{"/test2adsfljdflskaj", "99", "99", "100", "98"},
		{"/test3alj", "99", "99", "100", "98"},
		{"/test", "99", "99", "100", "98"},
		{"/test1132", "99", "99", "100", "98"},
		{"/test212lk", "99", "99", "100", "98"},
		{"/test3asd", "99", "99", "100", "98"},
		{"/testasdffdsa", "99", "99", "100", "98"},
		{"/test1asddf", "99", "99", "100", "98"},
		{"/test2asdf1", "99", "99", "100", "98"},
		{"/test311124324032", "99", "99", "100", "98"},
		{"/testdlaskjdfjl", "99", "99", "100", "98"},
		{"/test1lasdkfjlasdfkj", "99", "99", "100", "98"},
		{"/test2aldsfkjfdajkl", "99", "99", "100", "98"},
		{"/test3asdfljkdfslajk", "99", "99", "100", "98"},
	}},
	{Values: []string{"github.com 24"}, Children: []components.TableChildRow{
		{"/test11110", "99", "99", "100", "98"},
	}},
	{Values: []string{"github.com 25"}, Children: []components.TableChildRow{
		{"/adlcvkjtest1", "99", "99", "100", "98"},
		{"/sdfljasdfjtest1", "99", "99", "100", "98"},
	}},
}

// NewReportListView creates new instance of ReportListView.
func NewReportListView() *ReportListView {
	table := components.NewTable()

	table.SetHeaders(headers)
	table.SetRows(rows)

	return &ReportListView{
		table: table,
	}
}

// Update states of the ReportListView.
func (v ReportListView) Update(msg tea.Msg) (ReportListView, tea.Cmd) {
	var cmd tea.Cmd
	v.table, cmd = v.table.Update(msg)

	return v, cmd
}

func (v *ReportListView) updateTableHeaders(width int) {
	var headers = []components.TableHeader{
		{Title: "hostname", Width: width - 40},
		{Title: "Perf", Width: 10},
		{Title: "A11Y", Width: 10},
		{Title: "BP", Width: 10},
		{Title: "SEO", Width: 10},
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
