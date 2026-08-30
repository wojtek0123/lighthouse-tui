package ui

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/wojtek0123/lighthouse-tui/internal/lighthouse"
)

type ResultView struct {
	audits lighthouse.LighthouseResult
	loaded bool
}

func NewResultView() *ResultView {
	return &ResultView{
		loaded: false,
	}
}

func (r ResultView) Update(msg tea.Msg) (ResultView, tea.Cmd) {
	return r, nil
}

func (r ResultView) View() tea.View {
	if r.loaded {
		return tea.NewView(r.audits.Audits["first-contentful-paint"].DisplayValue)
	} else {
		return tea.NewView(fmt.Sprintf("RESULT %v", r.loaded))
	}
}
