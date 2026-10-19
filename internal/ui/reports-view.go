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

// ReportItem represents report list item, has filename which indicates where report has been finished.
type ReportItem struct {
	filename string
}

// ReportsView represents the view model for displaying report results,
// managing reports list, and tracking interactive UI state in the TUI.
type ReportsView struct {
	list     list.Model
	delegate *reportItemDelegate

	hostname string
	active   bool
}

// FilterValue filters list items.
func (i ReportItem) FilterValue() string { return i.filename }

type reportItemDelegate struct {
	styles styles
	focus  bool

	resultActive bool
}

func (d reportItemDelegate) Height() int                             { return 1 }
func (d reportItemDelegate) Spacing() int                            { return 0 }
func (d reportItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
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
			return lipgloss.JoinHorizontal(lipgloss.Left, lipgloss.NewStyle().Foreground(lipgloss.Color("170")).Render(">"), strings.Join(s, " "))
		}
	}

	if _, err := fmt.Fprint(w, fn(str)); err != nil {
		log.Printf("error rendering url list item: %v", err)
	}
}

// it should return (*ReportsView, error)?

// NewReportsView creates a ReportView.
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

// Update updates reports view state.
func (v ReportsView) Update(msg tea.Msg) (ReportsView, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case ActivePaneMsg:
		v.setResultPaneActive(msg.pane == resultPane)

		if msg.pane != reportsPane {
			return v, nil
		}

		if selectedItem, ok := v.list.SelectedItem().(ReportItem); ok {
			return v, SelectReport(selectedItem.filename)
		}

		return v, nil

	case DeleteHostnameMsg:
		v.list.SetItems([]list.Item{})
		return v, nil
	case HostnameSelectedMsg:
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
			cmds = append(cmds, SelectReport(selectedItem.filename))
		}
	}

	return v, tea.Batch(cmds...)
}

// View returns result view.
func (v ReportsView) View(height int) string {
	v.list.SetHeight(height)

	return v.list.View()
}

// GetReports gets reports based on provided hostname.
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

// InsertReport adds reportName to reports view list.
func (v *ReportsView) InsertReport(reportName string) tea.Cmd {
	items := v.list.Items()
	items = append(items, ReportItem{filename: reportName})
	cmd := v.list.SetItems(items)
	v.list.Select(len(v.list.Items()))

	return cmd
}

// SetActive sets list item which is currently active.
func (v *ReportsView) SetActive(active bool) {
	v.active = active
	v.delegate.focus = active
}

// ReportSelectedMsg represents a name of the selected report.
type ReportSelectedMsg struct {
	reportName string
}

// SelectReport returns tea.Cmd with selected report.
func SelectReport(reportName string) tea.Cmd {
	return func() tea.Msg {
		return ReportSelectedMsg{
			reportName: reportName,
		}
	}
}

// setResultPaneActive updates the active state of the result pane
// within the view's delegate.
func (v *ReportsView) setResultPaneActive(active bool) {
	v.delegate.resultActive = active
}

// SelectedReport returns the current selected report item.
func (v *ReportsView) SelectedReport() string {
	i, ok := v.list.SelectedItem().(ReportItem)

	if !ok {
		return ""
	}

	return i.filename
}
