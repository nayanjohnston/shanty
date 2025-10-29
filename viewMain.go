package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gen2brain/go-mpv"
)

type focusedModel int

const (
	focusPlayer focusedModel = iota
	focusLibrary
)

type modelMain struct {
	modelLibrary  tea.Model
	modelControls tea.Model
	focus         focusedModel
}

func awaitMpvEvent(m *mpv.Mpv) tea.Cmd {
	return func() tea.Msg {
		return msgMpvEvent(m.WaitEvent(10000))
	}
}

func (m modelMain) Init() tea.Cmd {
	var cmds []tea.Cmd

	m.ChangeFocus(focusPlayer)

	cmds = append(cmds, awaitMpvEvent(objectPlayer.mp))

	return tea.Batch(cmds...)
}

func (m modelMain) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			m.ChangeFocus(focusPlayer)
		case "K":
			m.ChangeFocus(focusLibrary)
		}

		switch m.focus {
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

func (m modelMain) ChangeFocus(newFocus focusedModel) {
	m.focus = newFocus

	if m.focus == focusPlayer {
		if pM, ok := m.modelControls.(modelControls); ok {
			pM.isFocused = true
			m.modelControls = pM
		}
	}
}

func (m modelMain) View() string {
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
