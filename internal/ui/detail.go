package ui

import (
	"fmt"

	task "github.com/Lev2307/taskman/internal/model"
	tea "github.com/charmbracelet/bubbletea"
)

type DetailModel struct {
	detailedTask task.Task
	err          error
}

type TaskDetailMsg struct {
	detailTaskMsg task.Task
	err           error
}

func DetailTaskCmd(task task.Task, err error) tea.Cmd {
	return func() tea.Msg {
		return TaskDetailMsg{detailTaskMsg: task, err: err}
	}
}

func (m DetailModel) Update(msg tea.Msg) (DetailModel, tea.Cmd) {
	switch msg := msg.(type) {
	case TaskDetailMsg:
		m.detailedTask = msg.detailTaskMsg
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m DetailModel) View() string {
	return fmt.Sprintf("detail page for task with id - %s", m.detailedTask.Title)
}
