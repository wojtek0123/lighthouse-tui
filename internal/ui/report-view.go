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

type status string

const (
	nonApplicableStatus status = "N/A"
	manualStatus        status = "MANUAL"
	errorStatus         status = "Error"
	informativeStatus   status = "INFO"
	goodStatus          status = "GOOD"
	poorStatus          status = "POOR"
	averageStatus       status = "AVERAGE"
	nilStatus           status = "-"
)

type progressMetric struct {
	label string
	score float64
}

// LighthouseMsg is a Bubble Tea message dispatched when lighthouse report is finished.
type LighthouseMsg struct {
	url    string
	result *lighthouse.Result
	err    error
}

func getLighthouseReport(cmd *runningCmd, url string) tea.Cmd {
	return func() tea.Msg {
		result, err := lighthouse.RunLighthouse(cmd.ctx, url)

		return LighthouseMsg{
			url:    url,
			result: result,
			err:    err,
		}
	}
}

// ReportView represents the view model for displaying report results.
type ReportView struct {
	activeTab string
	active    bool

	categories []lighthouse.Category

	err            error
	report         *lighthouse.Report
	metrics        []progressMetric
	overviewScores []progressMetric

	hostname   string
	reportName string

	showHeader bool
	header     *ReportHeader
	tabs       *components.Tab
	table      table.Model

	width  int
	height int
}

// NewReportView creates result view.
func NewReportView() *ReportView {
	tabs := components.NewTabs()
	table := table.New(
		table.WithFocused(true),
	)
	header := NewReportHeader()

	return &ReportView{
		activeTab:  "performance",
		tabs:       tabs,
		table:      table,
		active:     false,
		report:     nil,
		header:     header,
		showHeader: true,
	}
}

// Update updates result view state.
func (v ReportView) Update(msg tea.Msg) (ReportView, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "c":
			if !v.active {
				return v, nil
			}

			v.showHeader = !v.showHeader
			v.refreshTableHeight()
			return v, nil
		case "[":
			if !v.active {
				return v, nil
			}
			v.tabs.Previous()
			v.activeTab = v.categories[v.tabs.Index()].Title
			v.buildAudit()
		case "]":
			if !v.active {
				return v, nil
			}
			v.tabs.Next()
			v.activeTab = v.categories[v.tabs.Index()].Title
			v.buildAudit()
		}
	case ReportViewSizeMsg:
		v.width = msg.width
		v.height = msg.height
		v.refreshTableHeight()
		v.table.SetWidth(msg.width)
	case ActivePaneMsg:
		active := msg.pane == resultPane
		v.tabs.SetFocus(active)
		v.active = active
		v.table.Focus()
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
			}

			v.initView(*report)
		}
	}

	v.table, cmd = v.table.Update(msg)

	return v, cmd
}

func (v *ReportView) initView(report lighthouse.Report) {
	categories := slices.Collect(maps.Values(report.Categories))
	tabs := make([]string, 0, len(categories))

	sortCategories(categories)

	for _, c := range categories {
		tabs = append(tabs, c.Title)
	}

	v.categories = categories
	v.report = &report
	v.tabs.SetTabs(tabs)
	v.buildAudit()
	v.header.Init(report, categories)
	v.showHeader = true
}

// SetActiveTab sets which tab is currently active.
func (v *ReportView) SetActiveTab(tab string) {
	v.activeTab = tab
}

// View returns ui of ReportView as string.
func (v ReportView) View() string {
	if v.report == nil {
		return lipgloss.Place(
			v.width, v.height,
			lipgloss.Center, lipgloss.Center,
			lipgloss.NewStyle().Foreground(lipgloss.Color("245")).
				Render(lipgloss.JoinVertical(lipgloss.Center, "No results", "", "Press 'n' to enter url and measure metrics")),
		)
	}

	parts := []string{v.tabs.View()}

	if v.showHeader {
		metadata := v.header.View()
		parts = append([]string{metadata}, parts...)
	}
	if metrics := renderProgressMetrics(v.metrics, min(v.width, 30)); metrics != "" {
		parts = append(parts, metrics)
	}
	if overviewScore := renderProgressMetrics(v.overviewScores, min(v.width, 31)); overviewScore != "" {
		parts = append(parts, overviewScore)
	}
	parts = append(parts, v.table.View())

	return lipgloss.JoinVertical(lipgloss.Top, parts...)
}

func (v *ReportView) refreshTableHeight() {
	headerHeight := 0
	metricsHeight := lipgloss.Height(renderProgressMetrics(v.metrics, min(v.width, 30)))

	if v.showHeader {
		headerHeight = lipgloss.Height(v.header.View())
	}

	v.table.SetHeight(max(v.height-lipgloss.Height(v.tabs.View())-headerHeight-metricsHeight, 1))
}

func (v *ReportView) buildAudit() {
	v.table.SetRows([]table.Row{})
	v.metrics = make([]progressMetric, 0)
	v.overviewScores = make([]progressMetric, 0)

	if v.report == nil {
		return
	}

	category := v.categories[v.tabs.Index()]

	groupedAuditRefs := make(map[string][]lighthouse.AuditRefs)

	for _, ref := range category.AuditRefs {
		groupedAuditRefs[ref.Group] = append(groupedAuditRefs[ref.Group], ref)
	}

	rows := []table.Row{}

	for group, auditRefs := range groupedAuditRefs {
		if group == "metrics" {
			for _, metric := range auditRefs {
				audit, ok := v.report.Audits[metric.ID]

				if !ok {
					log.Print("failed to get audit metric")
					continue
				}

				v.metrics = append(v.metrics, progressMetric{label: audit.Title, score: *audit.Score})
			}
		}

		rows = append(rows, v.buildRows(group, groupedAuditRefs)...)
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

	if activeTab == string(lighthouse.PerformanceCategoryID) || activeTab == string(lighthouse.AgenticBrowsingCategoryID) {
		columns = slices.Insert(columns, 1, table.Column{Title: "Savings", Width: 20})
		columns[3].Width = width - 61
	}

	return columns
}

func (v ReportView) buildRows(group string, grouped map[string][]lighthouse.AuditRefs) []table.Row {
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

		score := scoreToStatus(audit.Score, audit.ScoreDisplayMode)

		row := table.Row{
			score,
			group,
			audit.Title,
		}

		if v.activeTab == string(lighthouse.PerformanceCategoryID) || v.activeTab == string(lighthouse.AgenticBrowsingCategoryID) {
			row = slices.Insert(row, 1, audit.DisplayValue)
		}
		rows = append(rows, row)
	}

	return rows
}

func scoreToStatus(score *float64, mode lighthouse.ScoreDisplayMode) string {
	switch mode {
	case lighthouse.NonApplicable:
		return badge(nonApplicableStatus, lipgloss.BrightBlack)
	case lighthouse.Manual:
		return badge(manualStatus, lipgloss.Yellow)
	case lighthouse.Error:
		return badge(errorStatus, lipgloss.Red)
	case lighthouse.Informative:
		return badge(informativeStatus, lipgloss.Blue)
	case lighthouse.Binary, lighthouse.MetricSavings, lighthouse.Numeric:
		if score == nil {
			return badge(nilStatus, lipgloss.BrightBlack)
		}

		switch {
		case *score >= 0.9:
			return badge(goodStatus, lipgloss.Green)
		case *score >= 0.5:
			return badge(averageStatus, lipgloss.Yellow)
		default:
			return badge(poorStatus, lipgloss.Red)
		}
	default:
		return badge(nilStatus, lipgloss.BrightBlack)
	}
}

func renderProgressMetrics(metrics []progressMetric, w int) string {
	if len(metrics) == 0 {
		return ""
	}

	views := []string{}

	for _, metric := range metrics {
		color := lipgloss.Red

		if metric.score > 0.9 {
			color = lipgloss.Green
		} else if metric.score > 0.5 {
			color = lipgloss.Yellow
		}

		p := progress.New(progress.WithColors(color))
		p.SetWidth(min(w, progressMaxWidth))

		freeSpace := 60 - lipgloss.Width(metric.label) - lipgloss.Width(p.ViewAs(metric.score))

		separator := strings.Repeat(" ", freeSpace)

		views = append(views, lipgloss.JoinHorizontal(lipgloss.Left, metric.label, separator, p.ViewAs(metric.score)))
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
func (v *ReportView) SetHostname(hostname string) {
	v.hostname = hostname
}

func badge(s status, fg color.Color) string {
	return lipgloss.NewStyle().
		Foreground(fg).
		Bold(true).
		Width(badgeWidth).
		Render(string(s))
}

// SetReport sets lighthouse report.
func (v *ReportView) SetReport(r *lighthouse.Report) {
	v.report = r
}

func sortCategories(categories []lighthouse.Category) {
	tabsOrder := map[lighthouse.CategoryID]int{
		lighthouse.PerformanceCategoryID:     0,
		lighthouse.AccessibilityCategoryID:   1,
		lighthouse.BestPractisesCategoryID:   2,
		lighthouse.SeoCategoryID:             3,
		lighthouse.AgenticBrowsingCategoryID: 4,
	}

	slices.SortFunc(categories, func(a, b lighthouse.Category) int {
		pA, existsA := tabsOrder[a.ID]
		pB, existsB := tabsOrder[b.ID]

		if !existsA {
			pA = -1
		}
		if !existsB {
			pB = -1
		}

		return cmp.Compare(pA, pB)
	})
}
