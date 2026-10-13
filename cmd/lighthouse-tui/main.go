package main

import (
	"fmt"
	"log"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/wojtek0123/lighthouse-tui/internal/ui"
)

func main() {
	f, err := tea.LogToFile("debug.log", "debug")

	if err != nil {
		log.Fatalf("fatal error opening log file: %v", err)
	}

	defer func() {
		if closeErr := f.Close(); closeErr != nil {
			log.Printf("error closing log file: %v", closeErr)
		}
	}()

	p := tea.NewProgram(ui.InitialModel())

	// Run the program
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
