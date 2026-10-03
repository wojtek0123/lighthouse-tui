package ui

import (
	tea "charm.land/bubbletea/v2"
	"github.com/wojtek0123/lighthouse-tui/internal/lighthouse"
)

type LighthouseMsg struct {
	url    string
	result lighthouse.LighthouseResult
	report string
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
