package ui

import (
	"cmp"
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
	loaded  bool

	hostname   string
	reportName string

	tabs       *components.Tab
	table      table.Model
	sortColumn int
	ascSort    bool

	progressBars []progress.Model
	// viewport     viewport.Model
	cursor int

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
		loaded:      false,
		overviewTab: overviewTab,
		activeTab:   "overview",
		tabs:        tabs,
		cursor:      0,
		table:       table,
		active:      false,
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
		return v, nil
	case ActivePaneMsg:
		active := msg.pane == resultPane
		v.tabs.SetFocus(active)
		v.active = active
		v.table.Focus()
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
	if !v.loaded {
		return lipgloss.Place(
			v.width, v.height,
			lipgloss.Center, lipgloss.Center,
			lipgloss.NewStyle().Foreground(lipgloss.Color("245")).
				Render(lipgloss.JoinVertical(lipgloss.Center, "No results", "", "Press 'n' to enter url and measure metrics")),
		)
	}

	parts := []string{v.tabs.View()}
	if metrics := v.renderMetrics(min(v.width, 30)); metrics != "" {
		parts = append(parts, metrics)
	}
	parts = append(parts, v.table.View())

	return lipgloss.JoinVertical(lipgloss.Top, parts...)
}

// refresh rebuilds rows/metrics and sizes the table. Call it from Update only.
func (v *ResultView) refresh() {
	if !v.loaded {
		return
	}

	v.tabs.SetTabs(v.categories)
	v.buildRows()

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
func (v *ResultView) buildRows() {
	// v.table = []table.Model{}
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
	slices.Sort(groups)

	// scoreWidth := 25
	// categoryWidth := min(v.width-10, v.width/3-scoreWidth-5)

	// different columns for different tab
	columns := []table.Column{
		{Title: "Score", Width: 20},
		{Title: "Savings", Width: 20},
		{Title: "Category", Width: 20},
		{Title: "Name", Width: v.width - 70},
	}

	rows := []table.Row{}

	for _, g := range groups {
		if g == "metrics" {
			v.metrics = grouped[g]
		}

		if g == "hidden" || g == "" {
			continue
		}

		for _, a := range grouped[g] {
			audit, ok := v.report.Audits[a.ID]

			// displayValue only in performance tab

			if !ok {
				rows = append(rows, table.Row{
					scoreToStatus(audit.Score, audit.ScoreDisplayMode),
					audit.DisplayValue,
					strings.TrimSpace(g),
					strings.TrimSpace(audit.Title),
				})
			} else {
				rows = append(rows, table.Row{
					scoreToStatus(audit.Score, audit.ScoreDisplayMode),
					audit.DisplayValue,
					strings.TrimSpace(g),
					strings.TrimSpace(audit.Title),
				})
			}

		}

	}

	sort.Slice(rows, func(i, j int) bool {
		return rows[i][0] < rows[j][0]
	})

	if len(rows) > 0 {
		v.table.SetColumns(columns)
		v.table.SetRows(rows)
		v.table.SetCursor(0)
	}
}

func scoreToStatus(score *float64, scoreDisplayMode lighthouse.ScoreDisplayMode) string {
	if scoreDisplayMode == lighthouse.NonApplicable {
		return "N/A"
	}

	if scoreDisplayMode == lighthouse.Manual {
		return lipgloss.NewStyle().Foreground(lipgloss.Yellow).Render("NEED VERIFICATION")
	}

	if scoreDisplayMode == lighthouse.Error {
		return lipgloss.NewStyle().Foreground(lipgloss.Red).Render("AUDIT FAILED TO RUN")
	}

	if scoreDisplayMode == lighthouse.Informative {
		return lipgloss.NewStyle().Foreground(lipgloss.Blue).Render("DIAGNOSTIC")
	}

	if score == nil {
		return "-"
	}

	if *score > 0.89 {
		return lipgloss.NewStyle().Foreground(lipgloss.Green).Render("GOOD")
	}

	if *score > 0.50 {
		return lipgloss.NewStyle().Foreground(lipgloss.Yellow).Render("NEEDS IMPROVEMENT")
	}

	return lipgloss.NewStyle().Foreground(lipgloss.Red).Render("POOR")
}

const progressMaxWidth = 75

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
