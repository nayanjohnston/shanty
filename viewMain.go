package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type focusedModel int

type ModelMain struct {
	modelLibrary  tea.Model
	modelControls tea.Model
}

func (m ModelMain) Init() tea.Cmd {
	var cmds []tea.Cmd

	cmds = append(cmds, awaitMpvEvent(objectPlayer.mpv))
	cmds = append(cmds, m.modelLibrary.Init())

	return tea.Batch(cmds...)
}

func (m ModelMain) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		terminalWidth = msg.Width
		terminalHeight = msg.Height

		m.modelLibrary, cmd = m.modelLibrary.Update(msg)
		cmds = append(cmds, cmd)
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "J":
			currentFocus = focusPlayer
			return m, nil
		case "K":
			currentFocus = focusLibrary
			return m, nil
		}

		switch currentFocus {
		case focusPlayer:
			m.modelControls, cmd = m.modelControls.Update(msg)
			cmds = append(cmds, cmd)
		case focusLibrary:
			m.modelLibrary, cmd = m.modelLibrary.Update(msg)
			cmds = append(cmds, cmd)
		}

	case msgMpvEvent:
		m.modelControls, cmd = m.modelControls.Update(msg)
		cmds = append(cmds, cmd)
	case msgLibraryLoaded:
		m.modelLibrary, cmd = m.modelLibrary.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m ModelMain) View() string {
	playerRender := m.modelControls.View()
	libraryRender := m.modelLibrary.View()

	statusRender := lipgloss.NewStyle().
		Width(terminalWidth).
		Background(lipgloss.Color("15")).
		Foreground(lipgloss.Color("0")).
		AlignHorizontal(lipgloss.Left).
		Padding(0, 1, 0, 1).
		Render("STATUS")

	s := lipgloss.NewStyle().
		Height(terminalHeight).
		MaxHeight(terminalHeight).
		Width(terminalWidth).
		Render(
			lipgloss.JoinVertical(
				lipgloss.Center,
				statusRender,
				lipgloss.NewStyle().
					Height(terminalHeight-4).
					AlignVertical(lipgloss.Bottom).
					Render(libraryRender),
				playerRender,
			),
		)
	return s
}
