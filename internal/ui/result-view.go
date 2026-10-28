package ui

import (
	"cmp"
	"maps"
	"slices"
	"strings"

	"charm.land/bubbles/v2/progress"
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

// ResultView represents the view model for displaying report results
type ResultView struct {
	activeTab string
	active    bool

	categories []string

	overviewTab *OverviewTab

	err     error
	report  *lighthouse.Report
	metrics []lighthouse.AuditRefs
	loaded  bool

	hostname   string
	reportName string

	tabs         *components.Tab
	table        table.Model
	progressBars []progress.Model
	// viewport     viewport.Model
	cursor int

	width int
	ready bool
}

// NewResultView creates result view.
func NewResultView() *ResultView {
	overviewTab := NewOverviewTab()
	tabs := components.NewTabs()
	table := table.New()
	table.Focus()
	// // s := table.DefaultStyles()
	// s.Header = s.Header.
	// 	BorderStyle(lipgloss.NormalBorder()).
	// 	BorderForeground(lipgloss.Color("240")).
	// 	BorderBottom(true).
	// 	Bold(false)
	// s.Selected = s.Selected.
	// 	Foreground(lipgloss.Color("229")).
	// 	Background(lipgloss.Color("57")).
	// 	Bold(false)
	// t.SetStyles(s)
	// table.SetStyles()

	return &ResultView{
		loaded:      false,
		overviewTab: overviewTab,
		activeTab:   "overview",
		tabs:        tabs,
		cursor:      0,
		table:       table,
	}
}

// Update updates result view state.
func (v ResultView) Update(msg tea.Msg) (ResultView, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "[":
			if !v.active {
				return v, nil
			}

			v.tabs.Previous()
			v.activeTab = v.categories[v.tabs.Index()]
			v.initAudits(80)
			// metrics := v.renderMetrics(80)
			// content := v.renderCategoryAudits()
			// v.viewport.SetContent(lipgloss.JoinVertical(lipgloss.Top, metrics, " ", content))
		case "]":
			if !v.active {
				return v, nil
			}

			v.tabs.Next()
			v.activeTab = v.categories[v.tabs.Index()]
			v.initAudits(80)
			// v.table[v.cursor].Focus()
			// metrics := v.renderMetrics(30)
			// content := v.renderCategoryAudits()
			// v.viewport.SetContent(lipgloss.JoinVertical(lipgloss.Top, metrics, " ", content))
		}
	case ActivePaneMsg:
		active := msg.pane == resultPane
		v.tabs.SetFocus(active)
		v.active = active
		return v, nil
	case HostnameSelectedMsg:
		v.hostname = msg.hostname
		v.reportName = ""
		v.loaded = false
	case ReportSelectedMsg:
		if v.hostname != "" {
			v.reportName = msg.reportName

			report, err := lighthouse.GetReport(v.hostname, v.reportName)

			if err != nil {
				v.err = err
			} else {
				v.report = report
				v.loaded = true
				// r.performanceTab.SetReport(report)
				v.getTabs()
				v.table.Focus()
				v.initAudits(30) // Initialize tables once report is loaded
				// content := v.renderCategoryAudits()
				// v.viewport.SetContent(content)
			}
		}
	}

	v.table, cmd = v.table.Update(msg)
	cmds = append(cmds, cmd)

	return v, tea.Batch(cmds...)
}

// SetActiveTab sets which tab is currently active.
func (v *ResultView) SetActiveTab(tab string) {
	v.activeTab = tab
}

// View returns ui of ResultView as string.
func (v ResultView) View(width, height int) string {
	v.tabs.SetTabs(v.categories)

	if !v.loaded {
		return lipgloss.Place(
			width, height-2,
			lipgloss.Center, lipgloss.Center,
			lipgloss.NewStyle().Foreground(lipgloss.Color("245")).
				Render(lipgloss.JoinVertical(lipgloss.Center, "No results", "", "Press 'n' to enter url and measure metrics")),
		)
	}

	metrics := v.renderMetrics(30)

	return lipgloss.NewStyle().Padding(0, 1).Render(
		lipgloss.JoinVertical(
			lipgloss.Top,
			v.tabs.View(),
			" ",
			metrics,
			v.table.View(),
		),
	)
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

func (v *ResultView) initAudits(width int) {
	// v.table = []table.Model{}
	v.metrics = []lighthouse.AuditRefs{}

	if v.activeTab == "overview" || v.report == nil {
		return
	}

	catID := lighthouse.CategoryID(v.activeTab)
	category, exists := v.report.Categories[catID]
	if !exists {
		return
	}

	grouped := make(map[string][]lighthouse.AuditRefs)
	for _, ref := range category.AuditRefs {
		groupKey := ref.Group
		grouped[groupKey] = append(grouped[groupKey], ref)
	}

	groups := slices.Collect(maps.Keys(grouped))
	slices.Sort(groups)

	columns := []table.Column{
		{Title: "", Width: width / 3},
		{Title: "", Width: width / 3},
		{Title: "", Width: width / 3},
	}
	rows := []table.Row{}

	for _, g := range groups {
		if g == "metrics" || g == "hidden" {
			v.metrics = grouped[g]
		}

		if g == "hidden" {
			continue
		}

		if g == "" {
			continue
		}

		for _, a := range grouped[g] {
			audit, ok := v.report.Audits[a.ID]

			if !ok {
				continue
			}

			if audit.Score == nil {
				continue
			}

			rows = append(rows, table.Row{g, audit.Title, audit.DisplayValue})
		}

	}

	if len(rows) == 0 {
		return
	}

	if len(rows) > 0 {
		// t := table.New()
		v.table.SetHeight(len(rows))
		v.table.SetWidth(width)
		v.table.SetColumns(columns)
		v.table.SetRows(rows)
	}
}

const progressMaxWidth = 75

func (v ResultView) renderMetrics(width int) string {
	views := []string{}

	for _, metric := range v.metrics {
		audit, ok := v.report.Audits[metric.ID]

		if !ok {
			continue
		}

		if audit.Score == nil {
			continue
		}

		p := progress.New()
		p.SetWidth(min(width, progressMaxWidth))

		views = append(views, lipgloss.JoinHorizontal(lipgloss.Left, audit.Title, p.ViewAs(*audit.Score)), " ")
	}

	return lipgloss.JoinVertical(lipgloss.Top, views...)
}

func (v ResultView) renderCategoryAudits() string {
	if v.activeTab == "overview" {
		return ""
	}

	// if len(v.table) == 0 {
	// 	return ""
	// }

	// tableViews := []string{}

	// for i, t := range v.table {
	// 	view := t.View()
	//
	// 	if i == v.cursor {
	// 		view = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("205")).Render(view)
	// 	}
	// 	tableViews = append(tableViews, view, " ")
	// }

	return lipgloss.JoinVertical(lipgloss.Top, v.table.View())
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
func (v *ResultView) SetHostname(hostname string) {
	v.hostname = hostname
}
