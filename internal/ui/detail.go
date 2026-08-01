package ui

import (
	"fmt"
	"strings"
	"time"

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

type BackToListMsg struct{}

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

func (m DetailModel) Update(msg tea.Msg) (DetailModel, tea.Cmd) {
	switch msg := msg.(type) {
	case TaskDetailMsg:
		m.detailedTask = msg.detailTaskMsg
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc":
			return m, backToListCmd()
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

	b.WriteString("[e] edit    [space] toggle done    [x] delete    [esc] back    [ctrl+c] quit\n\n")
	return b.String()
}
