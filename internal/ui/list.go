package ui

import (
	"fmt"
	"strings"

	task "github.com/Lev2307/taskman/internal/model"
	storage "github.com/Lev2307/taskman/internal/storage"
	tea "github.com/charmbracelet/bubbletea"
)

type ListModel struct {
	tasksList    []task.Task
	selectedTask int
	err          error
}

type TaskListMsg struct {
	listTasks []task.Task
	err       error
}

func NewListModel() ListModel {
	return ListModel{}
}

func ListTasksCmd(path string) tea.Cmd {
	return func() tea.Msg {
		tasks, err := storage.LoadTasksJson(path)
		return TaskListMsg{listTasks: tasks, err: err}
	}
}

func (m ListModel) Init() tea.Cmd {
	return ListTasksCmd(PATH)
}

func (m ListModel) Update(msg tea.Msg) (ListModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		}
	case TaskListMsg:
		m.tasksList = msg.listTasks
		m.err = msg.err
	}
	return m, nil
}

func (m ListModel) View() string {
	var b strings.Builder
	b.WriteString("You are on a list page!\n\n")
	b.WriteString("Here are your tasks 👇\n")
	for i := range m.tasksList {
		fmt.Fprintf(&b, "%d. ", i+1)
		fmt.Fprintf(&b, "%s ", m.tasksList[i].Title)
		if !m.tasksList[i].DueAt.IsZero() {
			dueAtFormatted := m.tasksList[i].DueAt.Format("2006-01-02")
			fmt.Fprintf(&b, "(due: %s)", dueAtFormatted)
		} else {
			fmt.Fprintf(&b, "(due: —)")
		}
		if !m.tasksList[i].Done {
			fmt.Fprintf(&b, `❌`)
		} else {
			fmt.Fprintf(&b, `✅`)
		}
		b.WriteString("\n")
	}
	return b.String()
}
