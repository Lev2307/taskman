package main

import (
	"fmt"
	"os"

	//task "github.com/Lev2307/taskman/internal/model"
	storage "github.com/Lev2307/taskman/internal/storage"
	ui "github.com/Lev2307/taskman/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
)

const PATH string = "tasks.json"

func main() {
	store := storage.NewStore(PATH)
	app := ui.New(store)
	if _, err := tea.NewProgram(app).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
