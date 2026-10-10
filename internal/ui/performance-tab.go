package ui

import (
	tea "charm.land/bubbletea/v2"
)

type PerformanceTab struct {
}

func NewPerformanceTab() *PerformanceTab {

	return &PerformanceTab{}
}

func (t PerformanceTab) Update(msg tea.Msg) (PerformanceTab, tea.Cmd) {
	var cmd tea.Cmd

	return t, cmd
}

func (t PerformanceTab) View() string {
	// perfLine := t.viewResult(width, "Performance", perf)
	// accessibilityLine := t.viewResult(width, "Accessibility", accessibility)
	// practisesLine := t.viewResult(width, "Best Practises", practises)
	// seoLine := t.viewResult(width, "SEO", seo)
	// agenticLine := t.viewResult(width, "Agentic Browsing", agentic)

	return "PerformanceTab XD"
}

// func (t PerformanceTab) viewResult(windowWidth int, label string, value float64) string {
// 	// maxWidth := 75
// 	// width := min(windowWidth, maxWidth)
// 	// bar := t.progress.ViewAs(value)
// 	// gap := width - lipgloss.Width(label) - lipgloss.Width(bar) - 2
//
// 	return label + " PerformanceTab XD"
// }
