package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Model Definition
type MainModel struct {
	contentModel    tea.Model
	controllerModel tea.Model
	queue           *Queue
}

type msgMainChangeFocus tea.Model

// Model Initialisation
func initMainModel() MainModel {
	newQueue := Queue{
		currentSong: 0,
	}

	mainModel := MainModel{
		controllerModel: initControllerModel(&newQueue),
		contentModel:    initContentModel(&newQueue),
		queue:           &newQueue,
	}

	return mainModel
}

// Tea Model Functions
func (m MainModel) Init() tea.Cmd {
	return tea.Batch(
		m.controllerModel.Init(),
		m.contentModel.Init(),
	)
}

func (m MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		terminalWidth = msg.Width
		terminalHeight = msg.Height

		status := m.renderStatus()
		controller := m.controllerModel.View()

		contentHeight = terminalHeight -
			lipgloss.Height(status) -
			lipgloss.Height(controller)

		contentWidth = terminalWidth
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "J":
			currentMainFocus = controllerFocus
		case "K":
			currentMainFocus = contentFocus
		}

		switch currentMainFocus {
		case controllerFocus:
			m.controllerModel, cmd = m.controllerModel.Update(msg)
		case contentFocus:
			m.contentModel, cmd = m.contentModel.Update(msg)
		}

		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...)
	}

	m.controllerModel, cmd = m.controllerModel.Update(msg)
	cmds = append(cmds, cmd)

	m.contentModel, cmd = m.contentModel.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m MainModel) View() string {
	status := m.renderStatus()
	controller := m.controllerModel.View()

	content := lipgloss.NewStyle().
		Height(contentHeight).
		Render(m.contentModel.View())

	return lipgloss.JoinVertical(
		lipgloss.Top,
		status,
		content,
		controller,
	)
}

// Render Functions
func (m MainModel) renderStatus() string {
	statusMessage := "N/A"

	switch currentMainFocus {
	case controllerFocus:
		statusMessage = "Controller"
	case contentFocus:
		switch currentContentFocus {
		case libraryFocus:
			statusMessage = "Library"
		case queueFocus:
			statusMessage = "Queue"
		case albumFocus:
			statusMessage = "Album"
		}
	}

	return lipgloss.NewStyle().
		Width(terminalWidth).
		Height(1).
		MaxHeight(1).
		Background(lipgloss.Color("15")).
		Foreground(lipgloss.Color("0")).
		Padding(0, 1, 0, 1).
		Render(statusMessage)
}
