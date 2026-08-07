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
	screenList
	screenDetail
)

type App struct {
	screen    screen
	create    CreateModel
	list      ListModel
	detail    DetailModel
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
	case TaskAddedMsg:
		if msg.err == nil {
			a.tasks = append(a.tasks, msg.task)
			a.screen = screenMain
			a.statusMsg = fmt.Sprintf("Task %q created successfully!", msg.task.Title)
		}
		return a, nil
	case TaskListMsg:
		if msg.err == nil {
			a.screen = screenList
		}
	case TaskDetailMsg:
		if msg.err == nil {
			a.screen = screenDetail
		}
	case RedirectToEditTask:
		a.screen = screenCreate
		a.create = NewCreateModel(modeEdit, msg.task)
	case BackToListMsg:
		a.screen = screenList
		a.list = NewListModel()
		return a, a.list.Init()
	case BackToMainPageMsg:
		a.screen = screenMain
	case TaskToggledMsg:
		a.screen = screenDetail
		a.detail = NewDetailModel(msg.taskID, "")
	case DeleteTaskMsg:
		a.screen = screenList
		a.list = NewListModel()
		var cmd tea.Cmd
		a.list, cmd = a.list.Update(msg)
		return a, tea.Batch(cmd, a.list.Init())
	case TaskEditedMsg:
		if msg.err == nil {
			a.screen = screenDetail
			a.detail = NewDetailModel(msg.task.ID, "🟢 Your task was edited successfully!!")
			return a, nil
		}
		// ошибка: остаёмся на форме, msg дойдёт до CreateModel ниже
	case tea.KeyMsg:
		if a.screen == screenMain {
			switch msg.String() {
			case "1", "a":
				a.screen = screenCreate
				a.create = NewCreateModel(modeCreate, task.Task{})
				a.statusMsg = ""
				return a, nil

			case "2", "l":
				a.screen = screenList
				a.list = NewListModel()
				a.statusMsg = ""
				return a, a.list.Init()

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
	if a.screen == screenList {
		var cmd tea.Cmd
		a.list, cmd = a.list.Update(msg)
		return a, cmd
	}
	if a.screen == screenDetail {
		var cmd tea.Cmd
		a.detail, cmd = a.detail.Update(msg)
		return a, cmd
	}

	return a, nil
}

func (a App) View() string {
	if a.screen == screenCreate {
		return a.create.View()
	}
	if a.screen == screenList {
		return a.list.View()
	}
	if a.screen == screenDetail {
		return a.detail.View()
	}
	view := "It`s a taskman\n\nExisting commands: \n1. Create a new task \n2. Task List \n3. Exit\n\n"
	if a.statusMsg != "" {
		view += "\n" + a.statusMsg + "\n\n"
	}
	return view
}
