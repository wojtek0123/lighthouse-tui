package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/wojtek0123/lighthouse-tui/internal/ui"
)

func main() {
	p := tea.NewProgram(ui.InitialModel())

	// Run the program
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
