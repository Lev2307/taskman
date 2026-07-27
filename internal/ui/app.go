package ui

import tea "github.com/charmbracelet/bubbletea"

type App struct{}

func New() App {
	return App{}
}

func (a App) Init() tea.Cmd {
	return nil
}

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return a, tea.Quit
		}
	}
	return a, nil
}

func (a App) View() string {
	return "taskman\n\npress q to quit\n"
}
