package ui

import (
	"fmt"
	"io"
	"log"
	"strings"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/wojtek0123/lighthouse-tui/internal/lighthouse"
)

// type ReportMsg struct{}

type ReportItem struct {
	filename string
}

type ReportsView struct {
	list     list.Model
	delegate *reportItemDelegate

	hostname string
	active   bool
}

func (i ReportItem) FilterValue() string { return i.filename }

type reportItemDelegate struct {
	styles styles
	focus  bool

	resultActive bool
}

func (d reportItemDelegate) Height() int                                  { return 1 }
func (d reportItemDelegate) Spacing() int                                 { return 0 }
func (d reportItemDelegate) Update(msg tea.Msg, list *list.Model) tea.Cmd { return nil }
func (d reportItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(ReportItem)

	if !ok {
		log.Printf("error while asserting report item type")
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i.filename)

	fn := d.styles.item.Render
	if index == m.Index() && d.focus {
		fn = func(s ...string) string {
			return d.styles.selectedItem.Render(strings.Join(s, " "))
		}
	} else if index == m.Index() && d.resultActive {
		fn = func(s ...string) string {
			return lipgloss.JoinHorizontal(lipgloss.Left, ">", strings.Join(s, " "))
		}
	}

	if _, err := fmt.Fprint(w, fn(str)); err != nil {
		log.Printf("error rendering url list item: %v", err)
	}

}

// it should return (*ReportsView, error)
func NewReportsView(active bool) *ReportsView {
	urls, err := lighthouse.GetHostnames()

	if err != nil {
		return &ReportsView{}
	}

	reports, err := lighthouse.GetReports(urls[0])

	if err != nil {
		return &ReportsView{}
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
		hostname: urls[0],
		delegate: delegate,
		active:   active,
	}
}

func (v ReportsView) Update(msg tea.Msg) (ReportsView, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case ActivePaneMsg:
		v.setResultPaneActive(msg.pane == ResultPane)

		if msg.pane != ReportsPane {
			return v, nil
		}

		if selectedItem, ok := v.list.SelectedItem().(ReportItem); ok {
			return v, selectReport(selectedItem.filename)
		}

		return v, nil

	case DeleteHostnameMsg:
		v.list.SetItems([]list.Item{})
		return v, nil
	case UrlSelectedMsg:
		v.GetReports(msg.hostname)

		return v, nil
	}

	previousIndex := v.list.Index()

	if v.active {
		v.list, cmd = v.list.Update(msg)
		cmds = append(cmds, cmd)
	}

	if v.list.Index() != previousIndex {
		if selectedItem, ok := v.list.SelectedItem().(ReportItem); ok {
			cmds = append(cmds, selectReport(selectedItem.filename))
		}
	}

	return v, tea.Batch(cmds...)
}

func (v ReportsView) View(height int) string {
	v.list.SetHeight(height)

	return v.list.View()
}

func (v *ReportsView) GetReports(hostname string) {
	reports, err := lighthouse.GetReports(hostname)

	if err != nil {
		log.Printf("error getting reports from hostname: %v", err)
		return
	}

	items := []list.Item{}

	for _, r := range reports {
		items = append(items, ReportItem{filename: r})
	}

	v.list.SetItems(items)
	v.list.ResetSelected()
	v.hostname = hostname
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

type ReportSelectedMsg struct {
	reportName string
}

func selectReport(reportName string) tea.Cmd {
	return func() tea.Msg {
		return ReportSelectedMsg{
			reportName: reportName,
		}
	}
}

func (v *ReportsView) setResultPaneActive(active bool) {
	v.delegate.resultActive = active
}
