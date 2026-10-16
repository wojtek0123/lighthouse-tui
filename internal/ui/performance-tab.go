package ui

import (
	tea "charm.land/bubbletea/v2"
)

// PerformanceTab represents the performance results of lighthouse report.
type PerformanceTab struct {
}

// NewPerformanceTab returns performance tab object.
func NewPerformanceTab() *PerformanceTab {
	return &PerformanceTab{}
}

// Update runs in loop to keep up to date performance tab.
func (t PerformanceTab) Update() (PerformanceTab, tea.Cmd) {
	var cmd tea.Cmd

	return t, cmd
}

// View renders performance tab for lighthouse report in result view.
func (t PerformanceTab) View() string {
	// perfLine := t.viewResult(width, "Performance", perf)
	// accessibilityLine := t.viewResult(width, "Accessibility", accessibility)
	// practisesLine := t.viewResult(width, "Best Practises", practises)
	// seoLine := t.viewResult(width, "SEO", seo)
	// agenticLine := t.viewResult(width, "Agentic Browsing", agentic)

	return "PerformanceTab XD"
}
