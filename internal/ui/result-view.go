package ui

import (
	"cmp"
	"image/color"
	"log"
	"maps"
	"slices"
	"sort"
	"strings"

	"charm.land/bubbles/v2/progress"
	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/wojtek0123/lighthouse-tui/internal/lighthouse"
	"github.com/wojtek0123/lighthouse-tui/internal/ui/components"
)

const progressMaxWidth = 75
const badgeWidth = 11

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

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

	hostname   string
	reportName string

	tabs       *components.Tab
	table      table.Model
	sortColumn int
	ascSort    bool

	width  int
	height int
	ready  bool
}

// NewResultView creates result view.
func NewResultView() *ResultView {
	overviewTab := NewOverviewTab()
	tabs := components.NewTabs()
	table := table.New(
		table.WithFocused(true),
	)

	return &ResultView{
		overviewTab: overviewTab,
		activeTab:   "overview",
		tabs:        tabs,
		table:       table,
		active:      false,
		report:      nil,
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
			v.refresh()
		case "]":
			if !v.active {
				return v, nil
			}
			v.tabs.Next()
			v.activeTab = v.categories[v.tabs.Index()]
			v.refresh()
		}
	case ResultSizeMsg:
		v.width = msg.width
		v.height = msg.height
		// v.refresh()
	case ActivePaneMsg:
		active := msg.pane == resultPane
		v.tabs.SetFocus(active)
		v.active = active
		v.table.Focus()
		v.refresh()
		return v, nil
	case HostnameSelectedMsg:
		v.hostname = msg.hostname
		v.reportName = ""
	case ReportSelectedMsg:
		if v.hostname != "" {
			v.reportName = msg.reportName

			report, err := lighthouse.GetReport(v.hostname, v.reportName)

			if err != nil {
				v.err = err
			} else {
				v.report = report
				v.getTabs()
				v.refresh()
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
func (v ResultView) View() string {
	if v.report == nil {
		return lipgloss.Place(
			v.width, v.height,
			lipgloss.Center, lipgloss.Center,
			lipgloss.NewStyle().Foreground(lipgloss.Color("245")).
				Render(lipgloss.JoinVertical(lipgloss.Center, "No results", "", "Press 'n' to enter url and measure metrics")),
		)
	}

	parts := []string{v.tabs.View()}
	if metrics := v.renderMetrics(min(v.width, 30)); metrics != "" {
		parts = append(parts, "", metrics)
	}
	parts = append(parts, "", v.table.View())

	return lipgloss.JoinVertical(lipgloss.Top, parts...)
}

// refresh rebuilds rows/metrics and sizes the table. Call it from Update only.
func (v *ResultView) refresh() {
	v.tabs.SetTabs(v.categories)
	v.buildAudit()

	metrics := v.renderMetrics(min(v.width, 30))

	tableHeight := v.height - lipgloss.Height(v.tabs.View())
	if metrics != "" {
		tableHeight -= lipgloss.Height(metrics)
	}

	v.table.SetWidth(v.width)
	v.table.SetHeight(max(tableHeight, 1))
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
func (v *ResultView) buildAudit() {
	v.metrics = []lighthouse.AuditRefs{}
	v.table.SetRows(nil)

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
	rows := []table.Row{}

	for _, g := range groups {
		if g == "metrics" {
			v.metrics = grouped[g]
		}

		rows = append(rows, v.buildRows(g, grouped)...)
	}

	sort.Slice(rows, func(i, j int) bool {
		return rows[i][0] < rows[j][0]
	})

	v.table.SetColumns(buildColumns(v.activeTab, v.width))
	v.table.SetRows(rows)
	v.table.SetCursor(0)
}

func buildColumns(activeTab string, width int) []table.Column {
	columns := []table.Column{
		{Title: "Status", Width: badgeWidth},
		{Title: "Category", Width: 20},
		{Title: "Name", Width: width - 41},
	}

	if activeTab == string(lighthouse.Performance) || activeTab == string(lighthouse.AgenticBrowsing) {
		columns = slices.Insert(columns, 1, table.Column{Title: "Savings", Width: 20})
		columns[3].Width = width - 61
	}

	return columns
}

func (v ResultView) buildRows(group string, grouped map[string][]lighthouse.AuditRefs) []table.Row {
	rows := []table.Row{}

	if group == "hidden" || group == "" {
		return []table.Row{}
	}

	for _, a := range grouped[group] {
		audit, ok := v.report.Audits[a.ID]

		if !ok {
			log.Printf("error while getting audit")
			continue
		}

		row := table.Row{
			scoreToStatus(audit.Score, audit.ScoreDisplayMode),
			strings.TrimSpace(group),
			strings.TrimSpace(audit.Title),
		}

		if v.activeTab == string(lighthouse.Performance) || v.activeTab == string(lighthouse.AgenticBrowsing) {
			row = slices.Insert(row, 1, audit.DisplayValue)

			rows = append(rows, row)
		} else {
			rows = append(rows, row)
		}

	}

	return rows
}

func scoreToStatus(score *float64, mode lighthouse.ScoreDisplayMode) string {
	switch mode {
	case lighthouse.NonApplicable:
		return badge("N/A", lipgloss.BrightBlack)
	case lighthouse.Manual:
		return badge("MANUAL", lipgloss.Yellow)
	case lighthouse.Error:
		return badge("ERROR", lipgloss.Red)
	case lighthouse.Informative:
		return badge("INFO", lipgloss.Blue)
	}

	if score == nil {
		return badge("-", lipgloss.BrightBlack)
	}

	switch {
	case *score > 0.89:
		return badge("GOOD", lipgloss.Green)
	case *score > 0.50:
		return badge("AVERAGE", lipgloss.Yellow)
	default:
		return badge("POOR", lipgloss.Red)
	}
}

func (v ResultView) renderMetrics(width int) string {
	views := []string{}

	if len(v.metrics) == 0 {
		return ""
	}

	for _, metric := range v.metrics {
		audit, ok := v.report.Audits[metric.ID]

		if !ok {
			continue
		}

		if audit.Score == nil {
			continue
		}

		color := lipgloss.Red

		if *audit.Score > 0.9 {
			color = lipgloss.Green
		} else if *audit.Score > 0.5 {
			color = lipgloss.Yellow
		}

		p := progress.New(progress.WithColors(color))
		p.SetWidth(min(width, progressMaxWidth))

		freeSpace := 60 - lipgloss.Width(audit.Title) - lipgloss.Width(p.ViewAs(*audit.Score))

		separator := strings.Repeat(" ", freeSpace)

		views = append(views, lipgloss.JoinHorizontal(lipgloss.Left, audit.Title, separator, p.ViewAs(*audit.Score)))
	}

	return lipgloss.JoinVertical(lipgloss.Top, views...)
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

	body := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Width(width).Height(height).BorderTop(false).BorderForeground(borderColor).Render(content)

	return lipgloss.NewStyle().Width(width).Height(height).Render(lipgloss.JoinVertical(lipgloss.Left, topBorder, body))
}

// SetHostname sets which hostname is focus from HostnamesView.
func (v *ResultView) SetHostname(hostname string) {
	v.hostname = hostname
}

func badge(label string, fg color.Color) string {
	return lipgloss.NewStyle().
		Foreground(fg).
		Bold(true).
		Width(badgeWidth).
		Render(label)
}

// SetReport sets lighthouse report.
func (v *ResultView) SetReport(r *lighthouse.Report) {
	v.report = r
}
