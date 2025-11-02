package main

import (
	"fmt"
	"math"
	"slices"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var topSongOfView int = 0

type QueueModel struct {
	queue  *Queue
	cursor int
}

type msgMoveCursor int
type msgClearQueue struct{}
type msgQueueSong *Song
type msgRemoveFromQueue int

func initQueueModel(queue *Queue) QueueModel {
	return QueueModel{
		queue: queue,
	}
}

func (m QueueModel) Init() tea.Cmd {
	return nil
}

func (m QueueModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j":
			cmds = append(cmds, func() tea.Msg { return msgMoveCursor(1) })
		case "k":
			cmds = append(cmds, func() tea.Msg { return msgMoveCursor(-1) })
		case "d":
			cmds = append(cmds, func() tea.Msg { return msgRemoveFromQueue(m.cursor) })
		case "enter":
			cmds = append(cmds, func() tea.Msg {
				m.queue.currentSong = m.cursor
				return msgLoadSong{playNow: true}
			})
		}
	case msgMoveCursor:
		m.cursor += int(msg)
	case msgRemoveFromQueue:
		// If playlist is empty, ignore...
		if len(m.queue.queue) <= 0 {
			break
		}

		// Get deletion postition
		pos := int(msg)

		// If we're deleting the currently playing song...
		if pos == m.queue.currentSong {
			// If last song, stop all playback.
			if len(m.queue.queue) == 1 {
				cmds = append(cmds, func() tea.Msg { return msgStopPlayback{} })
			} else {
				// If we're the last song in the queue, move back to avoid crash
				if m.queue.currentSong == len(m.queue.queue)-1 {
					m.queue.currentSong = len(m.queue.queue) - 2
				}

				cmds = append(cmds, func() tea.Msg { return msgLoadSong{playNow: true} })
			}
		} else { // If we're _not_ deleting current song...
			// If we are deleting a song before it, move current position back.
			if pos < m.queue.currentSong {
				m.queue.currentSong -= 1
			}
		}

		// Delete song
		m.queue.queue = slices.Delete(m.queue.queue, pos, pos+1)
	case msgClearQueue:
		m.queue.queue = []*Song{}
		m.queue.currentSong = 0
	case msgQueueSong:
		m.queue.queue = append(m.queue.queue, msg)
	}

	// Ensure cursor is safe
	if m.cursor < 0 {
		m.cursor = 0
	} else if m.cursor >= len(m.queue.queue) {
		m.cursor = len(m.queue.queue) - 1
	}

	if currentContentFocus != queueFocus {
		m.cursor = m.queue.currentSong
	}

	return m, tea.Batch(cmds...)
}

func (m QueueModel) View() string {
	if len(m.queue.queue) <= 0 {
		return ""
	}
	trackNumberWidth := 5
	durationWidth := 9

	titleWidth := int(float64(contentWidth-
		trackNumberWidth-
		durationWidth) * 0.7)

	artistWidth := contentWidth -
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

	output := lipgloss.NewStyle().
		Background(lipgloss.Color("234")).
		Render(lipgloss.JoinHorizontal(
			lipgloss.Center,
			styleTrackNumber.Render("No."),
			styleTitle.Render("Title"),
			styleArtist.Render("Artist"),
			styleDuration.Render("Duration"),
		))

	for m.cursor < topSongOfView {
		topSongOfView -= 1
	}

	spaceLeft := contentHeight - lipgloss.Height(output)
	index := topSongOfView
	isSpace := true

	for isSpace {
		if index >= len(m.queue.queue) {
			isSpace = false
			continue
		}

		song := m.queue.queue[index]

		styleRow := lipgloss.NewStyle()

		if m.queue.currentSong == index &&
			currentContentFocus == queueFocus {
			styleRow = styleRow.Foreground(colorFocus)
		} else if m.cursor == index {
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

		trackRender := styleRow.Render(
			lipgloss.JoinHorizontal(
				lipgloss.Top,
				styleTrackNumber.Render(trackNumber),
				styleTitle.Render(title),
				styleArtist.Render(artist),
				styleDuration.Render(duration),
			))

		spaceLeft -= lipgloss.Height(trackRender)

		if spaceLeft < 0 {
			if m.cursor >= index {
				newSpace := lipgloss.Height(trackRender)
				for newSpace > 0 {
					topSongOfView += 1
					spaceLeft += lipgloss.Height(tracks[0])
					newSpace -= lipgloss.Height(tracks[0])
					tracks = slices.Delete(tracks, 0, 1)
				}
				tracks = append(tracks, trackRender)
				index += 1
			} else {
				isSpace = false
			}
		} else {
			tracks = append(tracks, trackRender)
			index += 1
		}
	}

	for _, text := range tracks {
		output = lipgloss.JoinVertical(
			lipgloss.Center,
			output,
			text,
		)
	}

	return output
}
