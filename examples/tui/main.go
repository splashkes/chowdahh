package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/splashkes/chowdahh_recipes/examples/tui/api"
	"github.com/splashkes/chowdahh_recipes/examples/tui/ui"
)

func main() {
	token := api.LoadToken()
	client := api.NewClient("https://chowdahh.com", token)

	initial := ui.ScreenStreams
	if token == "" {
		initial = ui.ScreenAuth
	}

	app := ui.NewApp(client, initial)
	p := tea.NewProgram(app, tea.WithAltScreen(), tea.WithMouseCellMotion())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
