package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	sizeStatusWidth  int
	sizeStatusHeight int

	sizeMainWidth  int
	sizeMainHeight int

	sizeControlsWidth  int
	sizeControlsHeight int
)

var styleStatus = lipgloss.NewStyle().
	Background(lipgloss.Color("15")).
	Foreground(lipgloss.Color("0")).
	AlignHorizontal(lipgloss.Left).
	Padding(0, 1, 0, 1)

type focusView int
type focusModel int

type ModelMain struct {
	modelControls tea.Model
	modelLibrary  tea.Model
	modelQueue    tea.Model
}

func initializeModelMain() ModelMain {
	modelMain := ModelMain{
		modelControls: initializeModelControls(),
		modelLibrary:  initializeModelLibrary(),
		modelQueue:    initializeModelQueue(),
	}

	return modelMain
}

func (m ModelMain) Init() tea.Cmd {
	var cmds []tea.Cmd

	cmds = append(cmds, awaitMpvEvent(objectPlayer.mpv))
	cmds = append(cmds, m.modelLibrary.Init())
	cmds = append(cmds, m.modelQueue.Init())

	return tea.Batch(cmds...)
}

func (m ModelMain) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		terminalWidth = msg.Width
		terminalHeight = msg.Height

		sizeControlsWidth = terminalWidth
		sizeControlsHeight = 3

		sizeStatusWidth = terminalWidth
		sizeStatusHeight = 1

		sizeMainWidth = terminalWidth
		sizeMainHeight = terminalHeight - sizeStatusHeight - sizeControlsHeight
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		}

		switch focusedView {
		case focusPlayer:
			m.modelControls, cmd = m.modelControls.Update(msg)
			cmds = append(cmds, cmd)
		case focusMain:
			switch focusedModel {
			case focusLibrary:
				m.modelLibrary, cmd = m.modelLibrary.Update(msg)
			case focusQueue:
				m.modelQueue, cmd = m.modelQueue.Update(msg)
			}
			cmds = append(cmds, cmd)
		}

		return m, tea.Batch(cmds...)
	}

	m.modelControls, cmd = m.modelControls.Update(msg)
	cmds = append(cmds, cmd)

	m.modelLibrary, cmd = m.modelLibrary.Update(msg)
	cmds = append(cmds, cmd)

	m.modelQueue, cmd = m.modelQueue.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m ModelMain) View() string {
	styleMain := lipgloss.NewStyle().
		MaxHeight(sizeMainHeight).
		Height(sizeMainHeight).
		MaxWidth(sizeMainWidth).
		AlignVertical(lipgloss.Bottom)

	playerRender := m.modelControls.View()

	statusString := ""

	switch focusedView {
	case focusMain:
		switch focusedModel {
		case focusLibrary:
			statusString = "LIBRARY"
		case focusQueue:
			statusString = "QUEUE"
		}
	case focusPlayer:
		statusString = "PLAYER"
	}

	statusRender := styleStatus.Width(sizeStatusWidth).Render(statusString)

	mainRender := ""

	switch focusedModel {
	case focusLibrary:
		mainRender = styleMain.Render(m.modelLibrary.View())
	case focusQueue:
		mainRender = styleMain.Render(m.modelQueue.View())
	default:
		mainRender = styleMain.Render("")
	}

	s := lipgloss.NewStyle().
		Height(terminalHeight).
		MaxHeight(terminalHeight).
		Width(terminalWidth).
		Render(
			lipgloss.JoinVertical(
				lipgloss.Center,
				statusRender,
				mainRender,
				playerRender,
			),
		)
	return s
}
