package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	tui "github.com/titusdmoore/cli-newsfeed/tui"
)

func main() {
	program := tea.NewProgram(tui.GenerateInitialStateModel(), tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := program.Run(); err != nil {
		fmt.Println("Unable to start tea program")
		os.Exit(1)
	}
}
