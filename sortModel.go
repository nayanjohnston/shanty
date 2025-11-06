package main

import (
	"math"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type SortModel struct{}

type msgSortSelect struct{}

type sortEnum string

const (
	sortRecent   sortEnum = "recent"
	sortArtist   sortEnum = "alphabeticalByArtist"
	sortAlbum    sortEnum = "alphabeticalByName"
	sortFrequent sortEnum = "frequent"
)

func initSortModel() SortModel {
	return SortModel{}
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
			cmds = append(cmds, getLibrary(string(sortArtist)))
		case "w":
			cmds = append(cmds, getLibrary(string(sortAlbum)))
		case "e":
			cmds = append(cmds, getLibrary(string(sortRecent)))
		case "r":
			cmds = append(cmds, getLibrary(string(sortFrequent)))
		}
		cmds = append(cmds, func() tea.Msg { return msgSortSelect{} })
	}

	return m, tea.Batch(cmds...)
}

func (m SortModel) View() string {
	choiceStyle := lipgloss.NewStyle().
		Width(int(math.Min(30, float64(contentWidth-2)))).
		Border(lipgloss.NormalBorder())

	choiceRender := choiceStyle.Render(
		"Press a letter to sort\n\nq: Artist/Year\nw: Album\ne: Last Played\nr: Most Played",
	)

	return choiceRender
}
