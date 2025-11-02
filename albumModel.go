package main

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type AlbumOption struct {
	id   string
	name string
}

type AlbumModel struct {
	album          *Album
	queue          *Queue
	focusedOnList  bool
	options        []AlbumOption
	optionSelected int
}

type msgShowAlbum *Album
type msgAlbumViewSelect struct{}

func initAlbumModel(queue *Queue) AlbumModel {
	albumModel := AlbumModel{
		queue:          queue,
		focusedOnList:  false,
		optionSelected: 0,
	}

	playAlbumOption := AlbumOption{
		id:   "play",
		name: "Play Album",
	}

	queueAlbumOption := AlbumOption{
		id:   "queue",
		name: "Add to queue",
	}

	albumModel.options = []AlbumOption{
		playAlbumOption,
		queueAlbumOption,
	}

	return albumModel
}

func (m AlbumModel) Init() tea.Cmd {
	return nil
}

func (m AlbumModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var _ []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "h":
			m.optionSelected -= 1
		case "l":
			m.optionSelected += 1
		case "j":
			m.focusedOnList = true
		case "k":
			m.focusedOnList = false
		case "enter":
			return m, func() tea.Msg { return msgAlbumViewSelect{} }
		}
	case msgShowAlbum:
		currentContentFocus = albumFocus
		m.album = msg
		return m, func() tea.Msg { return getSonglist(m.album) }
	case msgAlbumViewSelect:
		if m.focusedOnList {

		} else {
			switch m.options[m.optionSelected].id {
			case "play":
				return m, tea.Sequence(
					func() tea.Msg { return msgClearQueue{} },
					func() tea.Msg { return msgAddAlbumToQueue(m.album) },
				)
			case "queue":
				return m, func() tea.Msg { return msgAddAlbumToQueue(m.album) }
			}
		}
	}

	m.optionSelected = m.clampOptions()

	if currentContentFocus != albumFocus {
		m.album = nil
		m.optionSelected = 0
		m.focusedOnList = false
	}

	return m, nil
}

func (m AlbumModel) View() string {
	if m.album == nil {
		return "No album selected (how are you here?)"
	}

	if len(m.album.songlist) == 0 {
		return "Loading songlist (might take a while)"
	}

	viewWidth := contentWidth - 2
	viewHeight := contentHeight - 2

	artwork := drawImage(m.album.artwork)

	infoWidth := viewWidth - lipgloss.Width(artwork) - 1

	options := lipgloss.NewStyle().
		Width(infoWidth).
		Render(m.renderOptions())

	info := lipgloss.NewStyle().
		Padding(0, 1, 0, 1).
		Width(infoWidth).
		Height(lipgloss.Height(artwork) - lipgloss.Height(options)).
		Render(
			lipgloss.JoinVertical(
				lipgloss.Top,
				m.album.title,
				m.album.artist,
				fmt.Sprintf("%v", m.album.year),
			))

	infoRender := lipgloss.JoinVertical(
		lipgloss.Center,
		info,
		options,
	)

	albumRender := lipgloss.JoinHorizontal(
		lipgloss.Left,
		artwork,
		infoRender)

	output := lipgloss.NewStyle().
		Padding(0, 1, 0, 1).
		Width(viewWidth).
		Height(viewHeight).
		Border(lipgloss.NormalBorder()).
		Render(lipgloss.JoinVertical(
			lipgloss.Top,
			albumRender,
			m.renderSonglist(),
		))

	return output
}

func (m AlbumModel) renderOptions() string {
	output := ""

	for index, element := range m.options {
		opt := lipgloss.NewStyle().
			Padding(0, 1, 0, 1)

		if index == m.optionSelected && !m.focusedOnList {
			opt = opt.
				Background(colorFocus).
				Foreground(lipgloss.Color("0"))
		}

		output += opt.Render(element.name)
	}

	return output
}

func (m AlbumModel) renderSonglist() string {
	output := ""

	for _, element := range m.album.songlist {
		songStyle := lipgloss.NewStyle()
		if len(m.queue.queue) > 0 {
			if element == m.queue.queue[m.queue.currentSong] {
				songStyle = songStyle.
					Foreground(colorFocus)
			}
		}

		output = combineVertical(output, songStyle.Render(element.title))
	}

	return output
}

func (m AlbumModel) clampOptions() int {
	if m.optionSelected < 0 {
		m.optionSelected = 0
	} else if m.optionSelected >= len(m.options) {
		m.optionSelected = len(m.options) - 1
	}

	return m.optionSelected
}
