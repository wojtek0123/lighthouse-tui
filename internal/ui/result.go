package ui

import (
	"context"

	tea "charm.land/bubbletea/v2"
	"github.com/wojtek0123/lighthouse-tui/internal/lighthouse"
)

type LighthouseMsg struct {
	url    string
	result lighthouse.LighthouseResult
	err    error
}

func getLighthouseResult(ctx context.Context, url string) tea.Cmd {
	return func() tea.Msg {
		result, err := lighthouse.RunLighthouse(ctx, url)

		return LighthouseMsg{
			url:    url,
			result: result,
			err:    err,
		}
	}
}

func cancelLighthouseExecution(a App) (App, tea.Cmd) {
	if a.cmd == nil {
		return a, nil
	}

	a.CancelCommand()

	return a, nil
}
