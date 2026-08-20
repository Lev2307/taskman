package ui

import (
	"fmt"
	"strconv"
	"strings"

	task "github.com/Lev2307/taskman/internal/model"
	tea "github.com/charmbracelet/bubbletea"
)

type ListModel struct {
	store       TaskStore
	tasksList   []task.Task
	taskIdInput string
	statusMsg   string
	err         error
}

type TaskListMsg struct {
	listTasks []task.Task
	err       error
}

type BackToMainPageMsg struct{}

func NewListModel(s TaskStore) ListModel {
	return ListModel{store: s}
}

func ListTasksCmd(s TaskStore) tea.Cmd {
	return func() tea.Msg {
		tasks, err := s.List()
		return TaskListMsg{listTasks: tasks, err: err}
	}
}

func BackToMainPageCmd() tea.Cmd {
	return func() tea.Msg {
		return BackToMainPageMsg{}
	}
}

func (m ListModel) Init() tea.Cmd {
	return ListTasksCmd(m.store)
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
			taskFromDb, errorDb := m.store.GetByID(taskID)
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
	case DeleteTaskMsg:
		if msg.err == nil {
			m.statusMsg = "🚮 Your task `" + msg.title + "` was deleted successfully!"
		} else {
			m.err = fmt.Errorf("error while deleting: %w", msg.err)
		}
	}
	return m, nil
}

func (m ListModel) View() string {
	var b strings.Builder
	b.WriteString("You are on a list page!\n\n")
	if len(m.tasksList) == 0 {
		b.WriteString(`✏️` + "  You haven`t created any tasks yet... Press [esc] and create a new one!\n\n")
	} else {
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
	}

	if m.statusMsg != "" {
		b.WriteString("\n" + m.statusMsg + "\n")
	}

	return b.String()
}
