package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type SortModel struct {
	cursor int
}

type sortEnum int

const (
	sortRecent sortEnum = iota
	sortArtist
	sortAlbum
)

func initSortModel() SortModel {
	return SortModel{
		cursor: 0,
	}
}

func (m SortModel) Init() tea.Cmd {
	return nil
}

func (m SortModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			cmds = append(cmds, getLibrary("alphabeticalByArtist"))
		case "w":
			cmds = append(cmds, getLibrary("recent"))
		}

		currentContentFocus = libraryFocus
	}

	return m, tea.Batch(cmds...)
}

func (m SortModel) View() string {
	choiceStyle := lipgloss.NewStyle()

	choiceRender := choiceStyle.Render(
		"Press a letter to sort\nq: Artist/Year\nw: Recent",
	)

	output := lipgloss.Place(
		20,
		20,
		lipgloss.Center,
		lipgloss.Center,
		choiceRender,
	)

	return output
}
