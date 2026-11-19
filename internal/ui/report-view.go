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
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/wojtek0123/lighthouse-tui/internal/lighthouse"
	"github.com/wojtek0123/lighthouse-tui/internal/ui/components"
)

var (
	missingIndicatorStyle = lipgloss.NewStyle().Foreground(lipgloss.BrightBlack)
	goodIndicatorStyle    = lipgloss.NewStyle().Foreground(lipgloss.Green)
	averageIndicatorStyle = lipgloss.NewStyle().Foreground(lipgloss.Yellow)
	poorIndicatorStyle    = lipgloss.NewStyle().Foreground(lipgloss.Red)
)

const progressMaxWidth = 75

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
	activeTab lighthouse.CategoryID
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
			v.updateTableHeight()
			return v, nil
		case "[":
			if !v.active {
				return v, nil
			}
			v.tabs.Previous()
			v.activeTab = v.categories[v.tabs.Index()].ID
			err := v.buildAudits()

			if err != nil {
				v.err = err
			}

			return v, nil
		case "]":
			if !v.active {
				return v, nil
			}
			v.tabs.Next()
			v.activeTab = v.categories[v.tabs.Index()].ID
			err := v.buildAudits()

			if err != nil {
				v.err = err
			}

			return v, nil
		}
	case ReportViewSizeMsg:
		v.width = msg.width
		v.height = msg.height
		v.updateTableHeight()
		v.table.SetWidth(msg.width)
	case ActivePaneMsg:
		// TODO: potencial improvement
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

			err = v.initView(*report)

			if err != nil {
				v.err = err
			}

			return v, nil
		}
	}

	v.table, cmd = v.table.Update(msg)

	return v, cmd
}

func (v *ReportView) initView(report lighthouse.Report) error {
	categories := slices.Collect(maps.Values(report.Categories))
	tabs := make([]string, 0, len(categories))

	sortCategories(categories)

	for _, c := range categories {
		tabs = append(tabs, c.Title)
	}

	v.categories = categories
	v.report = &report
	v.tabs.SetTabs(tabs)
	v.header.Init(report, categories)
	v.showHeader = true

	err := v.buildAudits()

	if err != nil {
		return fmt.Errorf("failed to build audits %w", err)
	}

	return nil
}

// SetActiveTab sets which tab is currently active.
func (v *ReportView) SetActiveTab(tab lighthouse.CategoryID) {
	v.activeTab = tab
}

// View returns ui of ReportView as string.
func (v ReportView) View() string {
	if v.err != nil {
		return lipgloss.Place(
			v.width, v.height,
			lipgloss.Center, lipgloss.Center,
			lipgloss.NewStyle().Foreground(lipgloss.Red).
				Render(v.err.Error()),
		)
	}

	if v.report == nil {
		return lipgloss.Place(
			v.width, v.height,
			lipgloss.Center, lipgloss.Center,
			lipgloss.NewStyle().Foreground(lipgloss.BrightBlack).
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

func (v *ReportView) updateTableHeight() {
	headerHeight := 0
	metricsHeight := lipgloss.Height(renderProgressMetrics(v.metrics, min(v.width, 30)))

	if v.showHeader {
		headerHeight = lipgloss.Height(v.header.View())
	}

	v.table.SetHeight(max(v.height-lipgloss.Height(v.tabs.View())-headerHeight-metricsHeight, 1))
}

func (v *ReportView) buildAudits() error {
	v.table.SetRows([]table.Row{})
	v.metrics = make([]progressMetric, 0)
	v.overviewScores = make([]progressMetric, 0)

	if v.report == nil {
		return nil
	}

	category := v.categories[v.tabs.Index()]
	rows := []table.Row{}

	for _, ref := range category.AuditRefs {
		if ref.Group == "metrics" {
			metric, ok := v.report.Audits[ref.ID]

			if !ok {
				return fmt.Errorf("expected Audit %T", metric)
			}

			v.metrics = append(v.metrics, progressMetric{label: metric.Title, score: *metric.Score})
		} else if ref.Group != "hidden" && ref.Group != "" {
			audit, ok := v.report.Audits[ref.ID]

			if !ok {
				return fmt.Errorf("expected Audit but got %T", audit)
			}

			// score := 0.0
			//
			// if audit.Score != nil {
			// 	score = *audit.Score
			// } else {
			// 	score = 1.1
			// }

			indicator := scoreIndicator(audit)
			priority := strconv.Itoa(audit.ScorePriority())

			row := table.Row{
				priority,
				indicator,
				ref.Group,
				audit.Title,
			}

			if v.activeTab == lighthouse.PerformanceCategoryID || v.activeTab == lighthouse.AgenticBrowsingCategoryID {
				row = table.Row{priority, indicator, audit.DisplayValue, ref.Group, audit.Title}
			}

			rows = append(rows, row)
		}
	}

	slices.SortFunc(rows, func(i, j table.Row) int {
		return cmp.Compare(i[0], j[0])
	})

	for i := range rows {
		if len(rows[i]) > 0 {
			rows[i] = rows[i][1:]
		}
	}

	v.table.SetColumns(buildColumns(v.activeTab, v.width))
	v.table.SetRows(rows)
	v.table.SetCursor(0)

	return nil
}

func buildColumns(activeTab lighthouse.CategoryID, width int) []table.Column {
	columns := []table.Column{
		{Title: "", Width: 1},
		{Title: "Category", Width: 30},
		{Title: "Name", Width: width - 39},
	}

	if activeTab == lighthouse.PerformanceCategoryID || activeTab == lighthouse.AgenticBrowsingCategoryID {
		columns = slices.Insert(columns, 1, table.Column{Title: "Savings", Width: 26})
		columns[3].Width = width - 67
	}

	return columns
}

func scoreIndicator(a lighthouse.Audit) string {
	if a.Score == nil {
		return missingIndicatorStyle.Render("○")
	}

	switch {
	case *a.Score >= lighthouse.NeedImprovementThreshold:
		return goodIndicatorStyle.Render("●")
	case *a.Score >= lighthouse.PoorThreshold:
		return averageIndicatorStyle.Render("■")
	default:
		return poorIndicatorStyle.Render("▲")
	}
}

func renderProgressMetrics(metrics []progressMetric, w int) string {
	if len(metrics) == 0 {
		return ""
	}

	views := []string{}

	for _, metric := range metrics {
		color := lipgloss.Red

		if metric.score >= lighthouse.NeedImprovementThreshold {
			color = lipgloss.Green
		} else if metric.score >= lighthouse.PoorThreshold {
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
