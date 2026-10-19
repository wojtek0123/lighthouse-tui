package ui

import (
	"cmp"
	"maps"
	"slices"
	"strings"

	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/wojtek0123/lighthouse-tui/internal/lighthouse"
	"github.com/wojtek0123/lighthouse-tui/internal/ui/components"
)

// LighthouseMsg is a Bubble Tea message dispatched when lighthouse report is finished.
type LighthouseMsg struct {
	url    string
	result *lighthouse.Result
	err    error
}

func getLighthouseResult(cmd *runningCmd, url string) tea.Cmd {
	return func() tea.Msg {
		result, err := lighthouse.RunLighthouse(cmd.ctx, url)

		return LighthouseMsg{
			url:    url,
			result: result,
			err:    err,
		}
	}
}

// Tab represents available tabs in result view.
// type Tab int
//
// const (
// 	overview Tab = iota
// 	performance
// )

// ResultView represents the view model for displaying report results,
// managing sub-tabs, and tracking interactive UI state in the TUI.
type ResultView struct {
	activeTab string
	active    bool

	categories []string

	overviewTab    *OverviewTab
	performanceTab *PerformanceTab

	err    error
	report *lighthouse.Report
	loaded bool

	hostname   string
	reportName string

	tabs         *components.Tabs
	tables       []table.Model
	focusedTable int
}

// NewResultView creates result view.
func NewResultView() *ResultView {
	overviewTab := NewOverviewTab()
	performanceTab := NewPerformanceTab()
	tabs := components.NewTabs()
	// tabs.SetTabs([]string{"Overview", "Performance"})

	return &ResultView{
		loaded:         false,
		overviewTab:    overviewTab,
		activeTab:      "overview",
		performanceTab: performanceTab,
		tabs:           tabs,
		focusedTable:   0,
		tables:         []table.Model{},
	}
}

// Update updates result view state.
func (r ResultView) Update(msg tea.Msg) (ResultView, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case ActivePaneMsg:
		active := msg.pane == resultPane

		r.tabs.SetFocus(active)
		r.active = active
		return r, nil
	case HostnameSelectedMsg:
		r.hostname = msg.hostname
		r.reportName = ""
		r.loaded = false
	case ReportSelectedMsg:
		if r.hostname != "" {
			r.reportName = msg.reportName

			report, err := lighthouse.GetReport(r.hostname, r.reportName)

			if err != nil {
				r.err = err
			} else {
				r.report = report
				r.loaded = true
				r.performanceTab.SetReport(report)
				r.getTabs()
			}
		}
	}

	if r.active {
		// r.table, cmd = r.table.Update(msg)
		// cmds = append(cmds, cmd)

		// if len(r.tables) > 0 && r.focusedTable < len(r.tables) {
		// 	r.tables[r.focusedTable], cmd = r.tables[r.focusedTable].Update(msg)
		// }

		for i := range r.tables {
			r.tables[i], cmd = r.tables[i].Update(msg)
			cmds = append(cmds, cmd)
		}

		*r.tabs, cmd = r.tabs.Update(msg)
		cmds = append(cmds, cmd)
	}

	return r, tea.Batch(cmds...)
}

// SetActiveTab sets which tab is currently active.
func (r *ResultView) SetActiveTab(tab string) {
	r.activeTab = tab
}

// View returns ui of ResultView as string.
func (r ResultView) View(width, height int) string {
	r.tabs.SetTabs(r.categories)

	if r.loaded {
		tabContent := r.renderCategoryAudits(width - 4)

		return lipgloss.NewStyle().Padding(0, 1).Render(lipgloss.JoinVertical(lipgloss.Top, r.tabs.View(), " ", tabContent))
	}

	return lipgloss.Place(width, height-2, lipgloss.Center, lipgloss.Center, lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render(lipgloss.JoinVertical(lipgloss.Center, "No results", "", "Press 'n' to enter url and measure metrics")))
}

func (v *ResultView) getTabs() {
	v.categories = v.categories[:0]

	for _, c := range slices.Collect(maps.Values(v.report.Categories)) {
		v.categories = append(v.categories, c.ID)
	}

	v.categories = append(v.categories, "overview")

	tabsOrder := map[string]int{
		"overview":                         0,
		string(lighthouse.Performance):     1,
		string(lighthouse.Accessibility):   2,
		string(lighthouse.BestPractises):   3,
		string(lighthouse.SEO):             4,
		string(lighthouse.AgenticBrowsing): 5,
	}

	slices.SortFunc(v.categories, func(a, b string) int {
		pA, existsA := tabsOrder[a]
		pB, existsB := tabsOrder[b]

		if !existsA {
			pA = 999
		}
		if !existsB {
			pB = 999
		}

		return cmp.Compare(pA, pB)
	})
}

func (v *ResultView) renderCategoryAudits(width int) string {
	if v.activeTab == "overview" {
		return ""
	}

	auditRefs := v.report.Categories[lighthouse.CategoryID(v.activeTab)].AuditRefs

	grouped := make(map[string][]lighthouse.AuditRefs)

	// component := map[string]string{
	// 	"metrics": "progress",
	// }

	for _, ref := range auditRefs {
		groupKey := ref.Group
		grouped[groupKey] = append(grouped[groupKey], ref)
	}

	groups := slices.Collect(maps.Keys(grouped))

	// tables := []table.Model{}

	for _, g := range groups {
		if g == "metrics" {
			continue
		}

		t := table.New()

		columns := []table.Column{
			{Title: "Name", Width: width},
		}
		rows := []table.Row{}

		for _, a := range grouped[g] {
			audit := v.report.Audits[a.ID]

			rows = append(rows, table.Row{audit.Title})
		}

		t.SetHeight(len(rows))
		t.SetWidth(width)
		t.SetColumns(columns)
		t.SetRows(rows)
		v.tables = append(v.tables, t)
	}

	// table := table.New()

	// catergoryGroups

	// group == metric  to progress bar, reszta tabelka

	// auditIDs map[lighthouse.CategoryID][]string
	//
	// group audits per group field
	// metrics -> progress bar
	// insights -> table with audits -> audit details (separate view?)
	// diagnostics -> table with audits

	// view := ""

	tableViews := []string{}

	for _, t := range v.tables {
		tableViews = append(tableViews, t.View(), " ")
	}

	return lipgloss.JoinVertical(lipgloss.Top, tableViews...)
}

// RenderPane renders pane with borders. Border color change based on focus state.
func RenderPane(header, content string, width, height int, focus bool) string {
	borderColor := inactiveBorderColor
	titleColor := inactiveTitleColor

	if focus {
		borderColor = activeBorderColor
		titleColor = lipgloss.Color("170")
	}

	borderStyle := lipgloss.NewStyle().Foreground(borderColor)
	border := lipgloss.RoundedBorder()
	renderedTitle := lipgloss.NewStyle().Padding(0, 1).Foreground(titleColor).Render(header)

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

// SetHostname sets which hostname is focus from HostnamesView.
func (r *ResultView) SetHostname(hostname string) {
	r.hostname = hostname
}
