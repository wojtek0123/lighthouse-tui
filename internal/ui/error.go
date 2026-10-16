package ui

import tea "charm.land/bubbletea/v2"

type errMsg struct{ err error }

// ThrowError returns a Bubble Tea command that wraps the provided error
// into an internal errMsg message, allowing asynchronous errors to be
// safely handled and displayed within the application's update loop.
func ThrowError(err error) tea.Cmd {
	return func() tea.Msg {
		return errMsg{
			err: err,
		}
	}
}

// Error returns error message as string.
func (e errMsg) Error() string { return e.err.Error() }
