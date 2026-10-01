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
	list   list.Model
	url    string
	active bool
}

func (i ReportItem) FilterValue() string { return "" }

type reportItemDelegate struct {
	styles styles
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

	str := fmt.Sprintf("%s", i.filename)

	fn := d.styles.item.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return d.styles.selectedItem.Render(strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}

func NewReportsView() *ReportsView {
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

	list.SetFilteringEnabled(false)
	list.SetShowStatusBar(false)
	list.SetShowHelp(false)
	list.SetDelegate(reportItemDelegate{styles: newStyles(false)})
	list.SetShowTitle(false)

	return &ReportsView{
		list: list,
		url:  urls[0],
	}
}

func (v ReportsView) Update(msg tea.Msg) (ReportsView, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	if v.active {
		v.list, cmd = v.list.Update(msg)
		cmds = append(cmds, cmd)
	}

	return v, tea.Batch(cmds...)
}

func (v ReportsView) View() string {
	return v.list.View() + "\n" + v.url
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
	v.url = url
}
