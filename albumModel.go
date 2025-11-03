package main

import (
	"fmt"
	"slices"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

var albumListTop int = 0

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
	cursor         int
}

type msgShowAlbum *Album
type msgAlbumViewSelect struct{}

func initAlbumModel(queue *Queue) AlbumModel {
	albumModel := AlbumModel{
		queue:          queue,
		focusedOnList:  false,
		optionSelected: 0,
		cursor:         0,
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
			if !m.focusedOnList {
				m.optionSelected -= 1
			}
		case "l":
			if !m.focusedOnList {
				m.optionSelected += 1
			}
		case "j":
			if !m.focusedOnList {
				m.focusedOnList = true
			} else {
				m.cursor += 1
			}
		case "k":
			if m.cursor <= 0 {
				m.focusedOnList = false
			} else {
				m.cursor -= 1
			}
		case "enter":
			return m, func() tea.Msg { return msgAlbumViewSelect{} }
		}
	case msgShowAlbum:
		currentContentFocus = albumFocus
		m.album = msg
		return m, func() tea.Msg { return getSonglist(m.album) }
	case msgAlbumViewSelect:
		if m.focusedOnList {
			cmds := []tea.Cmd{func() tea.Msg { return msgQueueSong(m.album.songlist[m.cursor]) }}

			// If nothing is in the queue, just start playing.
			if len(m.queue.queue) <= 0 {
				cmds = append(cmds, func() tea.Msg { return msgLoadSong{playNow: true} })
			}

			return m, tea.Sequence(cmds...)
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

	if m.album != nil {
		if m.cursor >= len(m.album.songlist) {
			m.cursor = len(m.album.songlist) - 1
		}
	}

	if currentContentFocus != albumFocus {
		m.album = nil
		m.optionSelected = 0
		m.focusedOnList = false
		m.cursor = 0
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
			m.renderSonglist(viewWidth-2, viewHeight-lipgloss.Height(albumRender)),
		))

	return output
}

func (m AlbumModel) renderOptions() string {
	output := ""

	for index, element := range m.options {
		opt := lipgloss.NewStyle().
			Padding(0, 1, 0, 1)

		if index == m.optionSelected &&
			!m.focusedOnList &&
			currentMainFocus == contentFocus {
			opt = opt.
				Background(colorFocus).
				Foreground(lipgloss.Color("0"))
		}

		output += opt.Render(element.name)
	}

	return output
}

func (m AlbumModel) renderSonglist(listWidth int, listHeight int) string {
	output := ""

	// Setting static widths
	trackNumberWidth := 5
	durationWidth := 10

	// Creating styles
	posStyle := lipgloss.NewStyle().
		AlignHorizontal(lipgloss.Right).
		Padding(0, 1, 0, 1).
		Width(trackNumberWidth)
	timeStyle := lipgloss.NewStyle().
		AlignHorizontal(lipgloss.Right).
		Padding(0, 1, 0, 1).
		Width(durationWidth)
	songStyle := lipgloss.NewStyle().
		Width(listWidth - trackNumberWidth - durationWidth)

	// Table arrays
	rows := [][]string{}           // Contains every row in table
	heights := []int{}             // Contains heights of every row of table
	styles := [][]lipgloss.Style{} // Contains styling of each row

	// Create table
	t := table.New().
		Border(lipgloss.HiddenBorder()). // Don't display border
		Headers("No.", "Title", "Duration").
		Width(contentWidth - 4).
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
				if col == 0 {
					style = posStyle
				} else if col == 2 {
					style = timeStyle
				} else {
					style = songStyle
				}

				// Darken header
				style = style.
					Background(lipgloss.Color("234"))

				return style
			}

			// Get style for this row/column from styles array
			style = styles[row][col]

			// So you CAN modulo and it just didn't want to before???
			// alright. Alternate color of row.
			if row%2 == 0 {
				style = style.
					Background(lipgloss.Color("235"))
			} else {
				style = style.
					Background(lipgloss.Color("236"))
			}

			return style
		})

	// If top of list is above current cursor position, move back to keep cursor
	// in view
	for m.cursor < albumListTop && albumListTop > 0 {
		albumListTop -= 1
	}

	spaceLeft := listHeight // The available space we have left
	index := albumListTop   // The current position of whole list
	isSpace := true         // If we should continue running the monstrosity
	//						// below (There's still items and space isn't
	//						// filled)

	for isSpace {
		// If we've reached end of songlist, stop
		if index == len(m.album.songlist) {
			isSpace = false
			continue
		}

		// Get current song
		song := m.album.songlist[index]

		// Create copys of styles in case of highlight
		rowPosStyle := posStyle
		rowTimeStyle := timeStyle
		rowSongStyle := songStyle

		// Outputs of table items
		posRender := fmt.Sprintf("%v", index+1)
		timeRender := intToTime(int(song.duration))
		songRender := song.title

		// Highlight selected song in dimmed color
		if index == m.cursor && m.focusedOnList {
			rowPosStyle = rowPosStyle.Foreground(colorFocusDim)
			rowTimeStyle = rowTimeStyle.Foreground(colorFocusDim)
			rowSongStyle = rowSongStyle.Foreground(colorFocusDim)
		}

		// Highlight currently playing song in bright color
		if len(m.queue.queue) > 0 {
			if song == m.queue.queue[m.queue.currentSong] {
				rowPosStyle = rowPosStyle.Foreground(colorFocus)
				rowTimeStyle = rowTimeStyle.Foreground(colorFocus)
				rowSongStyle = rowSongStyle.Foreground(colorFocus)
			}
		}

		// Create a fake output to get the hight of the row (this sucks)
		rowEmulation := lipgloss.JoinHorizontal(
			lipgloss.Left,
			rowPosStyle.Render(posRender),
			rowSongStyle.Render(songRender),
			rowTimeStyle.Render(timeRender),
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
					albumListTop += 1

					spaceLeft += heights[0] // Regain space from first row
					newSpace -= heights[0]  // Inverse for space to lose

					// Remove first element (first row) of all arrays
					heights = slices.Delete(heights, 0, 1)
					rows = slices.Delete(rows, 0, 1)
					styles = slices.Delete(styles, 0, 1)
				}

				// Add current row to list
				rows = append(rows, []string{
					posRender,
					songRender,
					timeRender,
				})

				styles = append(styles, []lipgloss.Style{
					rowPosStyle,
					rowSongStyle,
					rowTimeStyle,
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
				posRender,
				songRender,
				timeRender,
			})

			styles = append(styles, []lipgloss.Style{
				rowPosStyle,
				rowSongStyle,
				rowTimeStyle,
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

func (m AlbumModel) clampOptions() int {
	if m.optionSelected < 0 {
		m.optionSelected = 0
	} else if m.optionSelected >= len(m.options) {
		m.optionSelected = len(m.options) - 1
	}

	return m.optionSelected
}
