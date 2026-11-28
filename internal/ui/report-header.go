package ui

import (
	"slices"
	"strconv"

	"charm.land/lipgloss/v2"
	"github.com/wojtek0123/lighthouse-tui/internal/report"
)

var labelStyle = lipgloss.NewStyle().Foreground(lipgloss.BrightBlack).MarginRight(2)
var metricValueStyle = lipgloss.NewStyle()
var viewStyle = lipgloss.NewStyle().PaddingBottom(1)

// ReportHeader represents metadata and summary of a report.
type ReportHeader struct {
	labels []string
	values []string

	maxLabelWidth int
}

// NewReportHeader creates new class of ReportHeader.
func NewReportHeader() *ReportHeader {
	return &ReportHeader{
		labels: []string{},
		values: []string{},
	}
}

// Init initializes report header class.
func (r *ReportHeader) Init(report report.Report, categories []report.Category) {
	n := len(categories) + 2
	labels := make([]string, 0, n)
	values := make([]string, 0, n)

	labels = append(labels, "URL", "Fetch time")
	values = append(values, report.FinalURL, report.FetchTime)

	for _, c := range categories {
		color := lipgloss.Green

		if c.Score <= 0.5 {
			color = lipgloss.Red
		} else if c.Score <= 0.9 {
			color = lipgloss.Yellow
		}

		labels = append(labels, c.Title)
		values = append(values, metricValueStyle.
			Foreground(color).
			Render(strconv.FormatFloat(c.Score, 'f', -1, 64)))
	}

	r.labels = labels
	r.values = values
	r.maxLabelWidth = maxLabelWidth(labels)
}

func maxLabelWidth(labels []string) int {
	if len(labels) == 0 {
		return 0
	}

	widths := []int{}

	for _, l := range labels {
		widths = append(widths, lipgloss.Width(l))
	}

	return slices.Max(widths)
}

// View return ReportHeader view.
func (r ReportHeader) View() string {
	rows := make([]string, 0, len(r.labels))

	for i, l := range r.labels {
		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Left, labelStyle.Width(r.maxLabelWidth).Render(l), r.values[i]))
	}

	return viewStyle.Render(lipgloss.JoinVertical(lipgloss.Top, rows...))
}
