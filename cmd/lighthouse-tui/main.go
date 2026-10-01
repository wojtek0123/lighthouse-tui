package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/wojtek0123/lighthouse-tui/internal/ui"
)

func main() {
	f, err := tea.LogToFile("debug.log", "debug")

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	defer f.Close()

	p := tea.NewProgram(ui.InitialModel())

	// Run the program
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
