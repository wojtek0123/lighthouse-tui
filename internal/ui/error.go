package ui

import tea "charm.land/bubbletea/v2"

type errMsg struct{ err error }

func ThrowError(err error) tea.Cmd {
	return func() tea.Msg {
		return errMsg{
			err: err,
		}
	}
}

func (e errMsg) Error() string { return e.err.Error() }
