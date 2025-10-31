package main

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

type ModelQueue struct {
}

func initializeModelQueue() ModelQueue {
	return ModelQueue{}
}

func (q ModelQueue) Init() tea.Cmd {
	return nil
}

func (q ModelQueue) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var _ tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "H":
			focusedModel = focusLibrary
		case "J":
			focusedView = focusPlayer
		}
	}

	return q, tea.Batch(cmds...)
}

func (q ModelQueue) View() string {
	text := ""
	for index, song := range objectPlayer.queue.songs {
		playText := "  "
		if index == objectPlayer.queue.index {
			playText = "|>"
		}

		text += fmt.Sprintf(" %v | %v | %v | %v | %v | %v\n",
			playText,
			index+1,
			song.title,
			song.artist,
			song.album.title,
			song.duration)
	}

	return text
}
