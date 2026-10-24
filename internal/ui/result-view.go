package ui

import (
	"cmp"
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"

	"charm.land/bubbles/v2/progress"
	"charm.land/bubbles/v2/table"
	"charm.land/bubbles/v2/viewport"
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

	err    error
	report *lighthouse.Report
	loaded bool

	hostname   string
	reportName string

	tabs         *components.Tab
	tables       []table.Model
	progressBars []progress.Model
	focusedTable int

	ready    bool
	viewport viewport.Model
}

// NewResultView creates result view.
func NewResultView() *ResultView {
	overviewTab := NewOverviewTab()
	tabs := components.NewTabs()

	return &ResultView{
		loaded:       false,
		overviewTab:  overviewTab,
		activeTab:    "overview",
		tabs:         tabs,
		focusedTable: 0,
		tables:       []table.Model{},
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
			v.initTables(80)
			metrics := v.renderMetrics()
			content := v.renderCategoryAudits()
			v.viewport.SetContent(lipgloss.JoinVertical(lipgloss.Top, metrics, " ", content))
		case "]":
			if !v.active {
				return v, nil
			}

			v.tabs.Next()
			v.activeTab = v.categories[v.tabs.Index()]
			v.initTables(80)
			metrics := v.renderMetrics()
			content := v.renderCategoryAudits()
			v.viewport.SetContent(lipgloss.JoinVertical(lipgloss.Top, metrics, " ", content))
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
				v.initTables(30) // Initialize tables once report is loaded
				content := v.renderCategoryAudits()
				v.viewport.SetContent(content)
			}
		}
	case tea.WindowSizeMsg:
		headerHeight := 1

		if !v.ready {
			v.viewport = viewport.New(viewport.WithWidth(90), viewport.WithHeight(msg.Height-10))
			v.viewport.YPosition = headerHeight
			v.viewport.HighlightNext()
			v.ready = true
		} else {
			v.viewport.SetWidth(msg.Width - 100)
			v.viewport.SetHeight(msg.Height - 10)
			v.initTables(msg.Width - 40) // Re-initialize/resize tables on window resize if needed
		}
	}

	if v.active {
		if len(v.tables) > 0 && v.focusedTable >= 0 && v.focusedTable < len(v.tables) {
			v.tables[v.focusedTable], cmd = v.tables[v.focusedTable].Update(msg)
			cmds = append(cmds, cmd)
		}

		v.viewport, cmd = v.viewport.Update(msg)
		cmds = append(cmds, cmd)
	}

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

	return lipgloss.NewStyle().Padding(0, 1).Render(
		lipgloss.JoinVertical(
			lipgloss.Top,
			v.activeTab,
			strconv.Itoa(len(v.tables)),
			v.tabs.View(),
			" ",
			v.viewport.View(),
			v.footerView(),
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

func (v *ResultView) initTables(width int) {
	v.tables = []table.Model{}

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

	for _, g := range groups {
		if g == "metrics" || g == "hidden" {
			v.buildMetrics(grouped[g], width)
		}

		if g == "hidden" {
			continue
		}

		if g == "" {
			continue
		}

		columns := []table.Column{
			{Title: g, Width: width / 2},
			{Title: "", Width: width / 2},
		}
		rows := []table.Row{}

		for _, a := range grouped[g] {
			audit, ok := v.report.Audits[a.ID]

			if !ok {
				continue
			}

			if audit.Score == nil {
				continue
			}

			rows = append(rows, table.Row{audit.Title, audit.DisplayValue})
		}

		if len(rows) == 0 {
			continue
		}

		if len(rows) > 0 {
			t := table.New()
			t.SetHeight(len(rows))
			t.SetWidth(width)
			t.SetColumns(columns)
			t.SetRows(rows)

			v.tables = append(v.tables, t)
		}
	}

	// Reset focused table index if it goes out of bounds
	if v.focusedTable >= len(v.tables) {
		v.focusedTable = 0
	}
}

const progressMaxWidth = 75

func (v *ResultView) buildMetrics(auditRefs []lighthouse.AuditRefs, width int) {
	v.progressBars = []progress.Model{}

	for _, a := range auditRefs {
		audit, ok := v.report.Audits[a.ID]

		if !ok {
			continue
		}

		if audit.Score == nil {
			continue
		}

		p := progress.New()
		p.SetWidth(min(width, progressMaxWidth))

		p.SetPercent(*audit.Score)

		v.progressBars = append(v.progressBars, p)
	}
}

func (v ResultView) renderMetrics() string {

	views := []string{}

	for _, p := range v.progressBars {
		views = append(views, p.View(), " ")
	}

	return lipgloss.JoinVertical(lipgloss.Top, views...)
}

func (v ResultView) renderCategoryAudits() string {
	if v.activeTab == "overview" {
		return ""
	}

	if len(v.tables) == 0 {
		return ""
	}

	tableViews := []string{}

	for i, t := range v.tables {
		view := t.View()
		// Optional: Highlight the focused table if you want visual feedback
		if i == v.focusedTable {
			view = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("205")).Render(view)
		}
		tableViews = append(tableViews, view, " ")
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

var (
	titleStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Right = "├"
		return lipgloss.NewStyle().BorderStyle(b).Padding(0, 1)
	}()

	infoStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Left = "┤"
		return titleStyle.BorderStyle(b)
	}()
)

func (v ResultView) footerView() string {
	info := infoStyle.Render(fmt.Sprintf("%3.f%%:%3.f%%", v.viewport.ScrollPercent()*100, v.viewport.HorizontalScrollPercent()*100))
	line := strings.Repeat("─", max(0, v.viewport.Width()-lipgloss.Width(info)))
	return lipgloss.JoinHorizontal(lipgloss.Center, line, info)
}

// SetHostname sets which hostname is focus from HostnamesView.
func (v *ResultView) SetHostname(hostname string) {
	v.hostname = hostname
}
