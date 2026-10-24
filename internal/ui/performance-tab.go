package ui

import (
	"strings"

	"charm.land/bubbles/v2/progress"
	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/wojtek0123/lighthouse-tui/internal/lighthouse"
)

// PerformanceTab represents the performance results of lighthouse report.
type PerformanceTab struct {
	report   *lighthouse.Report
	progress progress.Model
	table    table.Model
}

// NewPerformanceTab returns performance tab object.
func NewPerformanceTab() *PerformanceTab {
	progress := progress.New()
	table := table.New()

	return &PerformanceTab{
		progress: progress,
		table:    table,
	}
}

// Update runs in loop to keep up to date performance tab.
func (t PerformanceTab) Update(msg tea.Msg) (PerformanceTab, tea.Cmd) {
	var cmd tea.Cmd

	// press i to open information with metrics thresholds

	t.table, cmd = t.table.Update(msg)

	return t, cmd
}

// View renders performance tab for lighthouse report in result view.
func (t PerformanceTab) View(width int) string {
	if t.report == nil {
		return ""
	}

	// metrics := []string{"first-contentful-paint", "largest-contentful-paint", "cumulative-layout-shift", "total-blocking-time", "speed-index"}

	labels := []string{}

	// for _, m := range metrics {
	// labels = append(labels, t.viewResult(width, t.report.Audits[m].Title, t.report.Audits[m].Score), "")
	// }

	auditRefs := t.report.Categories[lighthouse.Performance].AuditRefs

	columns := []table.Column{
		{Title: "Diagnostic", Width: 10},
	}
	rows := []table.Row{}

	for _, audit := range auditRefs {
		rows = append(rows, table.Row{t.report.Audits[audit.ID].Title})
	}

	t.table.SetColumns(columns)
	t.table.SetRows(rows)
	t.table.SetHeight(20)
	t.table.SetWidth(width)

	return lipgloss.JoinVertical(lipgloss.Top, lipgloss.JoinVertical(lipgloss.Top, labels...), t.table.View())
}

// SetReport sets current selected lighthouse report.
func (t *PerformanceTab) SetReport(report *lighthouse.Report) {
	t.report = report
}

func (t PerformanceTab) viewResult(windowWidth int, label string, value float64) string {
	maxWidth := 75
	width := min(windowWidth, maxWidth)
	bar := t.progress.ViewAs(value)
	gap := width - lipgloss.Width(label) - lipgloss.Width(bar) - 2

	return label + strings.Repeat(" ", gap) + bar
}
