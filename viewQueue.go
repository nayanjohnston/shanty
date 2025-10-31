package main

import (
	"fmt"
	"log"
	"math"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type msgAddSongToQueue *Song
type msgLoadedCurrentSong int

type ModelQueue struct {
	songSelection int
}

func initializeModelQueue() ModelQueue {
	return ModelQueue{
		songSelection: 0,
	}
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
		case "j":
			q.songSelection += 1
			q.songSelection = int(math.Min(
				float64(q.songSelection),
				float64(len(objectPlayer.queue.songs)-1),
			))
		case "k":
			q.songSelection -= 1
			q.songSelection = int(math.Max(
				float64(q.songSelection),
				0,
			))
		case "enter":
			objectPlayer.queue.index = q.songSelection
			objectPlayer.loadSong(true)
		}
	case msgAddSongToQueue:
		log.Println(msg.artist)
	case msgLoadedCurrentSong:
		log.Println("Song loaded")
	}

	return q, tea.Batch(cmds...)
}

func (q ModelQueue) View() string {
	trackNumberWidth := 5
	durationWidth := 9

	titleWidth := int(float64(sizeMainWidth-
		trackNumberWidth-
		durationWidth) * 0.7)

	artistWidth := sizeMainWidth -
		titleWidth -
		trackNumberWidth -
		durationWidth

	styleTrackNumber := lipgloss.NewStyle().
		Width(trackNumberWidth).
		AlignHorizontal(lipgloss.Right).
		Padding(0, 1, 0, 1)

	styleTitle := lipgloss.NewStyle().
		Width(titleWidth).
		Padding(0, 1, 0, 0)

	styleArtist := lipgloss.NewStyle().
		Width(artistWidth).
		Padding(0, 1, 0, 0)

	styleDuration := lipgloss.NewStyle().
		Width(durationWidth).
		AlignHorizontal(lipgloss.Center).
		Padding(0, 1, 0, 0)

	tracks := []string{}

	for index, song := range objectPlayer.queue.songs {
		styleRow := lipgloss.NewStyle()

		if q.songSelection == index &&
			focusedView == focusMain {
			styleRow = styleRow.Foreground(colorFocus)
		} else if objectPlayer.queue.index == index {
			styleRow = styleRow.Foreground(colorFocusDim)
		}

		if math.Mod(float64(index), 2) == 0 {
			styleRow = styleRow.
				Background(lipgloss.Color("235"))
		} else {
			styleRow = styleRow.
				Background(lipgloss.Color("236"))
		}

		trackNumber := fmt.Sprintf("%v", index+1)
		title := song.title
		artist := song.artist
		duration := intToTime(int(song.duration))

		tracks = append(tracks, styleRow.Render(
			lipgloss.JoinHorizontal(
				lipgloss.Top,
				styleTrackNumber.Render(trackNumber),
				styleTitle.Render(title),
				styleArtist.Render(artist),
				styleDuration.Render(duration),
			)))
	}

	output := lipgloss.NewStyle().
		Background(lipgloss.Color("234")).
		Render(lipgloss.JoinHorizontal(
			lipgloss.Center,
			styleTrackNumber.Render("No."),
			styleTitle.Render("Title"),
			styleArtist.Render("Artist"),
			styleDuration.Render("Duration"),
		))

	for _, text := range tracks {
		output = lipgloss.JoinVertical(
			lipgloss.Center,
			output,
			text,
		)
	}

	return output
}
