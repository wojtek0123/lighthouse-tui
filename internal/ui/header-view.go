package ui

import (
	"context"

	"charm.land/lipgloss/v2"
)

// TODO:
// 1. Handle cancellation
// 2. Add help
// 3. Add url validation
// 4. Remove '>' from input
// pisanie umożliwić dopiero po wciśnięciu 'i'
// pisanie kończy się po kliknięciu 'enter' lub 'esc' (anuluj)
// dodać datę kiedy uruchomiony został test

var (
	activeBorderColor   = lipgloss.Color("62")  // Indigo/Purple.
	inactiveBorderColor = lipgloss.Color("240") // Dark Gray.
	// activeTitleColor    = lipgloss.Color("212") // Pink.
	inactiveTitleColor = lipgloss.Color("240") // Dark Gray.
)

type runningCmd struct {
	ctx    context.Context
	cancel context.CancelFunc
}
