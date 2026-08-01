package ui

import (
	"fmt"
	"strconv"
	"strings"

	task "github.com/Lev2307/taskman/internal/model"
	storage "github.com/Lev2307/taskman/internal/storage"
	tea "github.com/charmbracelet/bubbletea"
)

type ListModel struct {
	tasksList   []task.Task
	taskIdInput string
	err         error
}

type TaskListMsg struct {
	listTasks []task.Task
	err       error
}

type BackToMainPageMsg struct{}

func NewListModel() ListModel {
	return ListModel{}
}

func ListTasksCmd(path string) tea.Cmd {
	return func() tea.Msg {
		tasks, err := storage.LoadTasksJson(path)
		return TaskListMsg{listTasks: tasks, err: err}
	}
}

func BackToMainPageCmd() tea.Cmd {
	return func() tea.Msg {
		return BackToMainPageMsg{}
	}
}

func (m ListModel) Init() tea.Cmd {
	return ListTasksCmd(PATH)
}

func (m ListModel) Update(msg tea.Msg) (ListModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			taskID, err := strconv.Atoi(m.taskIdInput)
			if err != nil {
				m.err = fmt.Errorf("Wrong input task id")
				return m, nil
			}
			taskFromDb, errorDb := storage.GetTaskByID(PATH, taskID)
			if errorDb != nil {
				m.err = errorDb
				return m, nil
			}
			m.err = nil
			return m, DetailTaskCmd(taskFromDb, errorDb)
		case "ctrl+c":
			return m, tea.Quit
		case "backspace":
			if len(m.taskIdInput) > 0 {
				m.taskIdInput = m.taskIdInput[:len(m.taskIdInput)-1]
			}
		case "esc":
			return m, BackToMainPageCmd()
		default:
			if msg.Type == tea.KeyRunes {
				m.taskIdInput += msg.String()
				m.err = nil
			}
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
		fmt.Fprintf(&b, "%d. ", m.tasksList[i].ID)
		fmt.Fprintf(&b, "%s ", m.tasksList[i].Title)
		if !m.tasksList[i].DueAt.IsZero() {
			dueAtFormatted := m.tasksList[i].DueAt.Format("2006-01-02")
			fmt.Fprintf(&b, "(due: %s)", dueAtFormatted)
		} else {
			fmt.Fprintf(&b, "(due: —)")
		}
		if !m.tasksList[i].Done {
			fmt.Fprintf(&b, ` ❌`)
		} else {
			fmt.Fprintf(&b, ` ✅`)
		}
		b.WriteString("\n\n")
	}
	fmt.Fprintf(&b, "You can choose task to work with by entering task id in form below: %s", m.taskIdInput)
	if m.err != nil {
		fmt.Fprintf(&b, "\n\n⚠️  %s", m.err.Error())
	}
	return b.String()
}
