package main

import (
	"fmt"
	"slices"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

var queueListTop int = 0

type QueueModel struct {
	queue  *Queue
	cursor int
}

type msgMoveCursor int
type msgMoveSong struct {
	from int
	to   int
}
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
		case "ctrl+j":
			cmds = append(cmds, func() tea.Msg {
				return msgMoveSong{
					from: m.cursor,
					to:   m.cursor + 1,
				}
			})
		case "ctrl+k":
			cmds = append(cmds, func() tea.Msg {
				return msgMoveSong{
					from: m.cursor,
					to:   m.cursor - 1,
				}
			})
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
	case msgMoveSong:
		// If moving out of bounds, ignore
		if msg.to < 0 || msg.to >= len(m.queue.queue) {
			return m, nil
		}

		song := m.queue.queue[msg.from]

		movePlaying := false

		if m.queue.currentSong == msg.from {
			m.queue.currentSong += msg.to - msg.from
			movePlaying = true
		}

		m.queue.queue = slices.Delete(m.queue.queue, msg.from, msg.from+1)
		m.queue.queue = slices.Insert(m.queue.queue, msg.to, song)

		if !movePlaying {
			if m.queue.currentSong >= msg.from {
				m.queue.currentSong -= 1
			}
			if m.queue.currentSong >= msg.to {
				m.queue.currentSong += 1
			}
		}

		return m, func() tea.Msg { return msgMoveCursor(msg.to - msg.from) }

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

	m.cursor = m.cursorClamp()

	return m, tea.Batch(cmds...)
}

func (m QueueModel) View() string {
	if len(m.queue.queue) <= 0 {
		return "No queue"
	}

	output := ""

	// Setting static widths
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
		AlignHorizontal(lipgloss.Right).
		Padding(0, 1, 0, 0)

	// Table arrays
	rows := [][]string{}           // Contains every row in table
	heights := []int{}             // Contains heights of every row of table
	styles := [][]lipgloss.Style{} // Contains styling of each row

	// Create table
	t := table.New().
		Border(lipgloss.HiddenBorder()). // Don't display border
		Headers("No.", "Title", "Artist", "Duration").
		Width(contentWidth).
		BorderTop(false).    // We
		BorderLeft(false).   // don't
		BorderBottom(false). // want
		BorderRight(false).  // any
		BorderColumn(false). // borders
		BorderRow(false).    // please,
		BorderHeader(false). // thanks
		StyleFunc(func(row, col int) lipgloss.Style {
			var style lipgloss.Style

			// If we're the header, set the proper styling
			if row < 0 {
				switch col {
				case 0:
					style = styleTrackNumber
				case 1:
					style = styleTitle
				case 2:
					style = styleArtist
				case 3:
					style = styleDuration
				}

				// Darken header
				style = style.
					Background(lipgloss.Color("234"))

				return style
			}

			// Get style for this row/column from styles array
			style = styles[row][col]

			// Alternate color of row.
			if row%2 == 0 {
				style = style.
					Background(lipgloss.Color("235"))
			} else {
				style = style.
					Background(lipgloss.Color("236"))
			}

			return style
		})

	if m.cursor < queueListTop && queueListTop > 0 {
		queueListTop -= 1
	}

	spaceLeft := contentHeight - lipgloss.Height(t.Render())
	index := queueListTop
	isSpace := true

	for isSpace {
		// If we've reached end of queue, stop
		if index == len(m.queue.queue) {
			isSpace = false
			continue
		}

		// Get current song
		song := m.queue.queue[index]

		// Create copys of styles in case of highlight
		rowStyleTrackNumber := styleTrackNumber
		rowStyleTitle := styleTitle
		rowStyleArtist := styleArtist
		rowStyleDuration := styleDuration

		// Outputs of table items
		trackNumber := fmt.Sprintf("%v", index+1)
		title := song.title
		artist := song.artist
		duration := intToTime(int(song.duration))

		if index == m.cursor && currentMainFocus == contentFocus {
			rowStyleTrackNumber = rowStyleTrackNumber.Foreground(colorFocusDim)
			rowStyleTitle = rowStyleTitle.Foreground(colorFocusDim)
			rowStyleArtist = rowStyleArtist.Foreground(colorFocusDim)
			rowStyleDuration = rowStyleDuration.Foreground(colorFocusDim)
		}

		if m.queue.currentSong == index {
			rowStyleTrackNumber = rowStyleTrackNumber.Foreground(colorFocus)
			rowStyleTitle = rowStyleTitle.Foreground(colorFocus)
			rowStyleArtist = rowStyleArtist.Foreground(colorFocus)
			rowStyleDuration = rowStyleDuration.Foreground(colorFocus)
		}

		// Create a fake output to get the hight of the row (this sucks)
		rowEmulation := lipgloss.JoinHorizontal(
			lipgloss.Left,
			rowStyleTrackNumber.Render(trackNumber),
			rowStyleTitle.Render(title),
			rowStyleArtist.Render(artist),
			rowStyleDuration.Render(duration),
		)

		// Remove that hight from the space left
		spaceLeft -= lipgloss.Height(rowEmulation)

		// If we're out of space...
		if spaceLeft <= 0 {
			// If cursor is ahead of current last song...
			if m.cursor >= index {
				// Create copy of space we need to regain
				newSpace := lipgloss.Height(rowEmulation)

				// While there's still space to regain...
				for newSpace > 0 {
					// Move top index down
					queueListTop += 1

					spaceLeft += heights[0] // Regain space from first row
					newSpace -= heights[0]  // Inverse for space to lose

					// Remove first element (first row) of all arrays
					heights = slices.Delete(heights, 0, 1)
					rows = slices.Delete(rows, 0, 1)
					styles = slices.Delete(styles, 0, 1)
				}

				// Add current row to list
				rows = append(rows, []string{
					trackNumber,
					title,
					artist,
					duration,
				})

				styles = append(styles, []lipgloss.Style{
					rowStyleTrackNumber,
					rowStyleTitle,
					rowStyleArtist,
					rowStyleDuration,
				})

				heights = append(heights, lipgloss.Height(rowEmulation))

				// Increase index
				index += 1
			} else {
				// Cursor is in view, so just exit loop
				isSpace = false
			}
		} else {
			// We're not finished, so just add to list
			rows = append(rows, []string{
				trackNumber,
				title,
				artist,
				duration,
			})

			styles = append(styles, []lipgloss.Style{
				rowStyleTrackNumber,
				rowStyleTitle,
				rowStyleArtist,
				rowStyleDuration,
			})

			heights = append(heights, lipgloss.Height(rowEmulation))

			index += 1
		}
	}

	// Add rows to table
	t.Rows(rows...)

	// Render
	output = t.Render()

	return output
}

func (m QueueModel) cursorClamp() int {
	// Ensure cursor is safe
	if m.cursor < 0 {
		m.cursor = 0
	} else if m.cursor >= len(m.queue.queue) {
		m.cursor = len(m.queue.queue) - 1
	}

	if currentContentFocus != queueFocus {
		m.cursor = m.queue.currentSong
	}

	return m.cursor
}
