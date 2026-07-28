package ui

import (
	"fmt"

	task "github.com/Lev2307/taskman/internal/model"
	tea "github.com/charmbracelet/bubbletea"
)

type screen int

const PATH string = "tasks.json"
const (
	screenMain screen = iota
	screenCreate
)

type App struct {
	screen    screen
	create    CreateModel
	tasks     []task.Task
	path      string
	statusMsg string
}

func New() App {
	return App{}
}

func (a App) Init() tea.Cmd {
	return nil
}

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case taskAddedMsg:
		if msg.err == nil {
			a.tasks = append(a.tasks, msg.task)
			a.screen = screenMain
			a.statusMsg = fmt.Sprintf("Task %q created successfully!", msg.task.Title)
		}
		return a, nil
	case tea.KeyMsg:
		if a.screen == screenMain {
			switch msg.String() {
			case "1", "a":
				a.screen = screenCreate
				a.create = NewCreateModel()
				a.statusMsg = ""
				return a, nil
			case "2", "l":
				return a, tea.Quit // TODO: список тасок
			case "3", "ctrl+c":
				return a, tea.Quit
			}
		}
	}

	// если мы не на главном экране — отдаём msg под-модели
	if a.screen == screenCreate {
		var cmd tea.Cmd
		a.create, cmd = a.create.Update(msg)
		return a, cmd
	}

	return a, nil
}

func (a App) View() string {
	if a.screen == screenCreate {
		return a.create.View()
	}
	view := "It`s a taskman\n\nExisting commands: \n1. Create a new task \n2. Task List \n3. Exit\n\n"
	if a.statusMsg != "" {
		view += "\n" + a.statusMsg + "\n\n"
	}
	return view
}
