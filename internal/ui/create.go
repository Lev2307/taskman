package ui

import (
	"fmt"
	"strings"
	"time"

	task "github.com/Lev2307/taskman/internal/model"
	storage "github.com/Lev2307/taskman/internal/storage"
	tea "github.com/charmbracelet/bubbletea"
)

const lastField int = 3
const dueAtLayout = "2006-01-02"

var availableTags = []string{"work", "personal", "sport", "urgent"}

type CreateModel struct {
	titleInput string
	notesInput string

	tagChoices  []string         // варианты тегов на выбор
	tagCursor   int              // на каком теге сейчас курсор
	tagSelected map[int]struct{} // какие индексы отмечены

	dueAtInput string
	focusIndex int // какое поле сейчас активно
	err        error
}

type taskAddedMsg struct {
	task task.Task
	err  error
}

func NewCreateModel() CreateModel {
	return CreateModel{
		tagSelected: make(map[int]struct{}),
		tagChoices:  availableTags,
	}
}

func AddTaskCmd(path string, t task.Task) tea.Cmd {
	return func() tea.Msg {
		err := storage.AddTask(path, t)
		return taskAddedMsg{task: t, err: err}
	}
}

func (m CreateModel) Update(msg tea.Msg) (CreateModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if m.focusIndex == lastField {
				due, err := time.Parse(dueAtLayout, m.dueAtInput)
				if err != nil {
					m.err = fmt.Errorf("Wrong data layout: %w", err)
					return m, nil
				}
				if due.Before(time.Now()) {
					m.err = fmt.Errorf("Due time is in past")
					return m, nil
				}
				tags := make([]string, 0, len(m.tagChoices))
				for i := range m.tagSelected {
					tags = append(tags, m.tagChoices[i])
				}
				lastId, err := storage.GetLastID(PATH)
				if err != nil {
					return m, nil
				}
				t := task.Task{
					ID:        lastId + 1,
					Title:     m.titleInput,
					Notes:     m.notesInput,
					Tags:      tags,
					DueAt:     due,
					CreatedAt: time.Now(),
				}
				return m, AddTaskCmd(PATH, t)
			}
			m.focusIndex++
			return m, nil
		case "up", "k":
			if m.focusIndex == 2 && m.tagCursor > 0 {
				m.tagCursor--
			}
			return m, nil
		case "down", "j":
			if m.focusIndex == 2 && m.tagCursor < len(m.tagChoices)-1 {
				m.tagCursor++
			}
			return m, nil
		case "backspace":
			switch m.focusIndex {
			case 0:
				if len(m.titleInput) > 0 {
					m.titleInput = m.titleInput[:len(m.titleInput)-1]
				}
			case 1:
				if len(m.notesInput) > 0 {
					m.notesInput = m.notesInput[:len(m.notesInput)-1]
				}
			case 3:
				if len(m.dueAtInput) > 0 {
					m.dueAtInput = m.dueAtInput[:len(m.dueAtInput)-1]
				}
			}
			return m, nil
		case " ":
			switch m.focusIndex {
			case 0:
				m.titleInput += " "
			case 1:
				m.notesInput += " "
			case 2:
				if _, ok := m.tagSelected[m.tagCursor]; ok {
					delete(m.tagSelected, m.tagCursor)
				} else {
					m.tagSelected[m.tagCursor] = struct{}{}
				}
			}
			return m, nil
		case "ctrl+c":
			return m, tea.Quit
		default:
			if msg.Type == tea.KeyRunes {
				switch m.focusIndex {
				case 0:
					m.titleInput += msg.String()
				case 1:
					m.notesInput += msg.String()
				case 3:
					m.dueAtInput += msg.String()
				}
			}
			return m, nil
		}
	}
	return m, nil
}

func (m CreateModel) View() string {
	var b strings.Builder

	fmt.Fprintf(&b, "Title: %s\n", m.titleInput)
	fmt.Fprintf(&b, "Notes: %s\n", m.notesInput)
	b.WriteString("Tags:\n")
	for i, tag := range m.tagChoices {

		// Is the cursor pointing at this choice?
		cursor := " " // no cursor
		if m.tagCursor == i {
			cursor = ">" // cursor!
		}

		// Is this choice selected?
		checked := " " // not selected
		if _, ok := m.tagSelected[i]; ok {
			checked = "x" // selected!
		}
		// Render the row
		fmt.Fprintf(&b, "%s [%s] %s\n", cursor, checked, tag)
	}
	fmt.Fprintf(&b, "Due at (%s): %s\n", dueAtLayout, m.dueAtInput)

	if m.err != nil {
		fmt.Fprintf(&b, "\nError: %s\n", m.err)
	}
	return b.String()
}
