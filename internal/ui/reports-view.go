package ui

import (
	"fmt"
	"io"
	"strings"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"github.com/wojtek0123/lighthouse-tui/internal/lighthouse"
)

type ReportMsg struct {
	reports []string
}

type ReportItem struct {
	filename string
}

type ReportsView struct {
	list     list.Model
	delegate *reportItemDelegate

	url    string
	active bool
}

func (i ReportItem) FilterValue() string { return i.filename }

type reportItemDelegate struct {
	styles styles
	focus  bool
}

func (d reportItemDelegate) Height() int                                  { return 1 }
func (d reportItemDelegate) Spacing() int                                 { return 0 }
func (d reportItemDelegate) Update(msg tea.Msg, list *list.Model) tea.Cmd { return nil }
func (d reportItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(ReportItem)

	if !ok {
		return
	}

	// if i.loading == true {
	// 	spinner = d.spinner.View()
	// }

	str := fmt.Sprintf("%d. %s", index+1, i.filename)

	fn := d.styles.item.Render
	if index == m.Index() && d.focus {
		fn = func(s ...string) string {
			return d.styles.selectedItem.Render(strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}

func NewReportsView(active bool) *ReportsView {
	urls, err := lighthouse.GetUrls()

	if err != nil {
	}

	reports, err := lighthouse.GetReports(urls[0])

	if err != nil {
	}

	items := []list.Item{}

	for _, report := range reports {
		items = append(items, ReportItem{
			filename: report,
		})
	}

	list := list.New(items, reportItemDelegate{}, 20, listHeight)

	delegate := &reportItemDelegate{styles: newStyles(false), focus: active}

	list.SetFilteringEnabled(false)
	list.SetShowStatusBar(false)
	list.SetShowHelp(false)
	list.SetDelegate(delegate)
	list.SetShowTitle(false)
	list.SetShowPagination(false)

	return &ReportsView{
		list:     list,
		url:      urls[0],
		delegate: delegate,
		active:   active,
	}
}

func (v ReportsView) Update(msg tea.Msg) (ReportsView, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case SelectedUrlMsg:
		v.GetReports(msg.url)

		return v, nil
	}

	if v.active {
		v.list, cmd = v.list.Update(msg)
		cmds = append(cmds, cmd)
	}

	return v, tea.Batch(cmds...)
}

func (v ReportsView) View(height int) string {
	v.list.SetHeight(height)

	return v.list.View()
}

func (v *ReportsView) GetReports(url string) {
	reports, err := lighthouse.GetReports(url)

	if err != nil {
	}

	items := []list.Item{}

	for _, r := range reports {
		items = append(items, ReportItem{filename: r})
	}

	v.list.SetItems(items)
	v.list.ResetSelected()
	v.url = url
}

func (v *ReportsView) InsertReport(reportName string) tea.Cmd {
	items := v.list.Items()
	items = append(items, ReportItem{filename: reportName})
	cmd := v.list.SetItems(items)
	v.list.Select(len(v.list.Items()))

	return cmd
}

func (v *ReportsView) SetActive(active bool) {
	v.active = active
	v.delegate.focus = active
}
