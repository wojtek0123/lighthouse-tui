// Package main is heart of the app.
package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/wojtek0123/lighthouse-tui/internal/store"
	"github.com/wojtek0123/lighthouse-tui/internal/ui"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "fatal:", err)
		os.Exit(1)
	}
}

func run() (err error) {
	f, err := tea.LogToFile("debug.log", "debug")
	if err != nil {
		return fmt.Errorf("open log file: %w", err)
	}
	defer func() {
		if cerr := f.Close(); cerr != nil && err == nil {
			err = fmt.Errorf("close log file: %w", cerr)
		}
	}()

	dir, err := store.DefaultDir()

	if err != nil {
		return fmt.Errorf("failed to get default dir %w", err)
	}

	reportsStore := store.NewReportStore(dir)

	p := tea.NewProgram(ui.InitialModel(reportsStore))

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		return fmt.Errorf("error running program %w: ", err)
	}

	return nil
}
