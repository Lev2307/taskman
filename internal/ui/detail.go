package ui

import (
	"fmt"
	"strings"
	"time"

	task "github.com/Lev2307/taskman/internal/model"
	storage "github.com/Lev2307/taskman/internal/storage"
	tea "github.com/charmbracelet/bubbletea"
)

type DetailModel struct {
	store        *storage.Store
	detailedTask task.Task
	statusID     int
	statusMsg    string
	err          error
	statusDelete bool
}

type TaskDetailMsg struct {
	detailTaskMsg task.Task
	err           error
}

type TaskToggledMsg struct {
	taskID int
	err    error
}

type BackToListMsg struct{}

type DeleteTaskMsg struct {
	title string
	err   error
}

type RedirectToEditTask struct {
	task task.Task
}

type ClearStatusMsg struct {
	id int
}

func DetailTaskCmd(task task.Task, err error) tea.Cmd {
	return func() tea.Msg {
		return TaskDetailMsg{detailTaskMsg: task, err: err}
	}
}

func backToListCmd() tea.Cmd {
	return func() tea.Msg {
		return BackToListMsg{}
	}
}

func toggleDoneCmd(s *storage.Store, taskID int) tea.Cmd {
	return func() tea.Msg {
		err := s.ToggleDone(taskID)
		return TaskToggledMsg{taskID: taskID, err: err}
	}
}

func ClearStatusCmd(d time.Duration, statusID int) tea.Cmd {
	return tea.Tick(d, func(time.Time) tea.Msg {
		return ClearStatusMsg{id: statusID}
	})
}

func DeleteTaskCmd(s *storage.Store, title string, taskID int) tea.Cmd {
	return func() tea.Msg {
		err := s.Delete(taskID)
		return DeleteTaskMsg{title: title, err: err}
	}
}

func redirectToEditTaskCmd(task task.Task) tea.Cmd {
	return func() tea.Msg {
		return RedirectToEditTask{task: task}
	}
}

func NewDetailModel(s *storage.Store, taskID int, statusMessage string) DetailModel {
	task, err := s.GetByID(taskID)
	if err != nil {
		return DetailModel{err: err, statusMsg: "error", store: s}
	}
	return DetailModel{detailedTask: task, statusMsg: statusMessage, store: s}
}

func (m DetailModel) Update(msg tea.Msg) (DetailModel, tea.Cmd) {
	switch msg := msg.(type) {
	case TaskDetailMsg:
		m.detailedTask = msg.detailTaskMsg
		return m, nil
	case tea.KeyMsg:
		if m.statusDelete {
			switch msg.String() {
			case "y", "Y":
				m.statusDelete = false
				return m, DeleteTaskCmd(m.store, m.detailedTask.Title, m.detailedTask.ID)
			case "n", "N", "esc":
				m.statusDelete = false
			case "ctrl+c":
				return m, tea.Quit
			}
			return m, nil
		}
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc":
			return m, backToListCmd()
		case " ":
			if time.Now().After(m.detailedTask.DueAt) {
				m.statusMsg = "⚠️ You can`t mark your task because due time was expired."
				m.statusID += 1
				return m, ClearStatusCmd(2*time.Second, m.statusID)
			} else {
				return m, toggleDoneCmd(m.store, m.detailedTask.ID)
			}
		case "x":
			m.statusDelete = true
			return m, nil
		case "e":
			return m, redirectToEditTaskCmd(m.detailedTask)
		}
	case ClearStatusMsg:
		if msg.id == m.statusID {
			m.statusMsg = ""
		}
	}
	return m, nil
}

func (m DetailModel) View() string {
	var b strings.Builder
	b.WriteString("\n")
	status := "❌ not done"
	if m.detailedTask.Done {
		status = "💚 done"
	}
	fmt.Fprintf(&b, "Task #%d - %s \n\n", m.detailedTask.ID, status)
	fmt.Fprintf(&b, "  Title:    %s\n", m.detailedTask.Title)
	fmt.Fprintf(&b, "  Notes:    %s\n", m.detailedTask.Notes)

	b.WriteString("  Tags:    ")
	if len(m.detailedTask.Tags) == 0 {
		b.WriteString(" —")
	} else {
		for _, tag := range m.detailedTask.Tags {
			fmt.Fprintf(&b, " %s", tag)
		}
	}
	b.WriteString("\n")
	fmt.Fprintf(&b, "  Created:  %s\n", m.detailedTask.CreatedAt.Format("2006-01-02 15:04"))

	if m.detailedTask.DueAt.IsZero() {
		b.WriteString("  Due:      —\n\n")
	} else {
		daysLeft := int(time.Until(m.detailedTask.DueAt).Minutes() / 60 / 24)
		if daysLeft < 0 {
			fmt.Fprintf(&b, "  Due:      %s (overdue by %d d)\n\n", m.detailedTask.DueAt.Format("2006-01-02"), -daysLeft)
		} else {
			fmt.Fprintf(&b, "  Due:      %s (in %d d)\n\n", m.detailedTask.DueAt.Format("2006-01-02"), daysLeft)
		}
	}
	b.WriteString("___________________________________________________________\n\n")

	if m.statusDelete {
		fmt.Fprintf(&b, "Delete \"%s\"? [y/n]\n\n", m.detailedTask.Title)
	} else {
		if m.detailedTask.Done {
			b.WriteString("[e] edit    [space] toggle UNdone    [x] delete    [esc] back    [ctrl+c] quit\n\n")
		} else {
			b.WriteString("[e] edit    [space] toggle done    [x] delete    [esc] back    [ctrl+c] quit\n\n")
		}
	}

	if m.statusMsg != "" {
		b.WriteString("\n" + m.statusMsg + "\n")
	}
	return b.String()
}
