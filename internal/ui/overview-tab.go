package ui

import (
	"strings"

	"charm.land/bubbles/v2/progress"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// OverviewTab represents main tab of lighthouse report.
type OverviewTab struct {
	progress progress.Model
}

// NewOverviewTab returns OverviewTab.
func NewOverviewTab() *OverviewTab {
	p := progress.New()

	return &OverviewTab{
		progress: p,
	}
}

// Update keeps OverviewTab updated.
func (t OverviewTab) Update(msg tea.Msg) (OverviewTab, tea.Cmd) {
	var cmd tea.Cmd

	progress, cmd := t.progress.Update(msg)
	t.progress = progress

	return t, cmd
}

// View return overview tab ui.
func (t OverviewTab) View(perf, accessibility, practises, seo, agentic float64, width int) string {
	perfLine := t.viewResult(width, "Performance", perf)
	accessibilityLine := t.viewResult(width, "Accessibility", accessibility)
	practisesLine := t.viewResult(width, "Best Practises", practises)
	seoLine := t.viewResult(width, "SEO", seo)
	agenticLine := t.viewResult(width, "Agentic Browsing", agentic)

	return lipgloss.JoinVertical(lipgloss.Top, perfLine, "", accessibilityLine, "", practisesLine, "", seoLine, "", agenticLine)
}

func (t OverviewTab) viewResult(windowWidth int, label string, value float64) string {
	maxWidth := 75
	width := min(windowWidth, maxWidth)
	bar := t.progress.ViewAs(value)
	gap := width - lipgloss.Width(label) - lipgloss.Width(bar) - 2

	return label + strings.Repeat(" ", gap) + bar
}
