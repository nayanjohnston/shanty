package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gen2brain/go-mpv"
)

type mainModel struct {
	pModel tea.Model
	width  int
	height int
}

func pollMpv(m *mpv.Mpv) tea.Cmd {
	return func() tea.Msg {
		return mpvEventMsg(m.WaitEvent(10000))
	}
}

func (m mainModel) Init() tea.Cmd {
	var cmds []tea.Cmd

	cmds = append(cmds, pollMpv(player.mp))

	return tea.Batch(cmds...)
}

func (m mainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		if pM, ok := m.pModel.(playerModel); ok {
			pM.width = m.width
			pM.height = m.height
			m.pModel = pM
		}
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		}

		m.pModel, cmd = m.pModel.Update(msg)
		cmds = append(cmds, cmd)

	case mpvEventMsg:
		m.pModel, cmd = m.pModel.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m mainModel) View() string {
	s := lipgloss.JoinVertical(
		lipgloss.Center,
		"Some test text, just to see how this works!",
		m.pModel.View(),
	)
	return s
}
