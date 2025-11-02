package main

import tea "github.com/charmbracelet/bubbletea"

type AlbumModel struct {
	album *Album
	queue *Queue
}

type msgShowAlbum *Album

func initAlbumModel(queue *Queue) AlbumModel {
	return AlbumModel{
		queue: queue,
	}
}

func (m AlbumModel) Init() tea.Cmd {
	return nil
}

func (m AlbumModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var _ []tea.Cmd

	switch msg := msg.(type) {
	case msgShowAlbum:
		currentContentFocus = albumFocus
		m.album = msg
	}

	if currentContentFocus != albumFocus {
		m.album = nil
	}

	return m, nil
}

func (m AlbumModel) View() string {
	if m.album == nil {
		return "No album selected (how are you here?)"
	}

	output := ""
	return output
}
