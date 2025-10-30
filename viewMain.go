package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gen2brain/go-mpv"
)

type focusedModel int

type ModelMain struct {
	modelLibrary  tea.Model
	modelControls tea.Model
}

func awaitMpvEvent(m *mpv.Mpv) tea.Cmd {
	return func() tea.Msg {
		return msgMpvEvent(m.WaitEvent(10000))
	}
}

func (m ModelMain) Init() tea.Cmd {
	var cmds []tea.Cmd

	cmds = append(cmds, awaitMpvEvent(objectPlayer.mp))

	return tea.Batch(cmds...)
}

func (m ModelMain) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		terminalWidth = msg.Width
		terminalHeight = msg.Height
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
			break
		}

	case msgMpvEvent:
		m.modelControls, cmd = m.modelControls.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m ModelMain) View() string {
	playerRender := m.modelControls.View()

	s := lipgloss.NewStyle().Height(terminalHeight).Render(
		lipgloss.JoinVertical(
			lipgloss.Center,
			lipgloss.NewStyle().
				Height(terminalHeight-lipgloss.Height(playerRender)).
				Render("This is a funny test."),
			playerRender,
		),
	)
	return s
}
