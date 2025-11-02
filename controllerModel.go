package main

import (
	"fmt"
	"slices"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gen2brain/go-mpv"
)

// Model Definition
type ControllerModel struct {
	mpv         *mpv.Mpv
	queue       *Queue
	progressBar progress.Model
}

type msgMpvEvent *mpv.Event
type msgClearQueue struct{}
type msgQueueSong *Song
type msgLoadSong struct{ playNow bool }
type msgNextSong struct{}
type msgPrevSong struct{}
type msgRemoveFromQueue int
type msgStopPlayback struct{}

// Model Initialisation
func initControllerModel(queue *Queue) ControllerModel {
	newProgressBar := progress.New()
	newProgressBar.ShowPercentage = false

	controllerModel := ControllerModel{
		mpv:         initMpv(),
		progressBar: newProgressBar,
		queue:       queue,
	}

	return controllerModel
}

// Tea Model Functions
func (m ControllerModel) Init() tea.Cmd {
	return tea.Batch(
		func() tea.Msg { return msgMpvEvent(m.mpv.WaitEvent(10000)) },
	)
}

func (m ControllerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case " ":
			m.mpv.Command([]string{"cycle", "pause"})
		case "h":
			m.mpv.Command([]string{"seek", "-5", "relative"})
		case "l":
			m.mpv.Command([]string{"seek", "5", "relative"})
		case "k":
			m.mpv.Command([]string{"add", "volume", "5"})
		case "j":
			m.mpv.Command([]string{"add", "volume", "-5"})
		case "n":
			cmds = append(cmds, func() tea.Msg { return msgNextSong{} })
		case "p":
			cmds = append(cmds, func() tea.Msg { return msgPrevSong{} })
		}
	case msgMpvEvent:
		var e mpv.Event = *msg

		switch e.EventID {
		case mpv.EventEnd:
			if e.EndFile().Reason == mpv.EndFileEOF {
				cmd = func() tea.Msg { return msgNextSong{} }
				cmds = append(cmds, cmd)
			}
		}

		cmd = func() tea.Msg { return msgMpvEvent(m.mpv.WaitEvent(10000)) }
		cmds = append(cmds, cmd)
	case msgClearQueue:
		m.queue.queue = []*Song{}
		m.queue.currentSong = 0
	case msgQueueSong:
		m.queue.queue = append(m.queue.queue, msg)
	case msgLoadSong:
		if len(m.queue.queue) != 0 {
			// Clamp current song value to queue length
			if m.queue.currentSong < 0 {
				m.queue.currentSong = 0
			} else if m.queue.currentSong > len(m.queue.queue)-1 {
				m.queue.currentSong = len(m.queue.queue) - 1
			}

			// Load via mpv
			m.mpv.Command([]string{
				"loadfile",
				m.queue.queue[m.queue.currentSong].getUrl(),
			})

			m.mpv.SetProperty("pause", mpv.FormatFlag, !msg.playNow)
		}
	case msgNextSong:
		shouldPlay := true
		// If we're the last song in the queue, go to the first song and pause.
		if m.queue.currentSong >= len(m.queue.queue)-1 {
			m.queue.currentSong = 0
			shouldPlay = false
		} else {
			// Otherwise, go forward a song and reload.
			m.queue.currentSong += 1
		}

		cmds = append(cmds, tea.Sequence(
			func() tea.Msg { return msgLoadSong{playNow: shouldPlay} },
		))
	case msgPrevSong:
		property, _ := m.mpv.GetProperty("time-pos", mpv.FormatInt64)
		position, _ := property.(int64)

		if position <= 5 {
			if m.queue.currentSong > 0 {
				m.queue.currentSong -= 1
			}
		}

		cmds = append(cmds, tea.Sequence(
			func() tea.Msg { return msgLoadSong{playNow: true} },
		))
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
			} else { // Otherwise, play next song...
				cmds = append(cmds, func() tea.Msg { return msgLoadSong{playNow: true} })
			}
		} else { // If we're _not_ deleting current song...
			// If we are deleting a song before it, move current position back.
			if pos <= m.queue.currentSong {
				m.queue.currentSong -= 1
			}
		}

		// Delete song
		m.queue.queue = slices.Delete(m.queue.queue, pos, pos+1)

		// If we deleted a song at the end, move current song backwards (stops
		// out of bounds error).
		if m.queue.currentSong >= len(m.queue.queue) {
			m.queue.currentSong = len(m.queue.queue) - 1
		}
	case msgStopPlayback:
		m.mpv.Command([]string{"stop"})
	}

	return m, tea.Batch(cmds...)
}

func (m ControllerModel) View() string {
	isFocused := false

	if currentMainFocus == controllerFocus {
		isFocused = true
	}

	infoString := ""

	if len(m.queue.queue) > 0 {
		infoString = m.queue.queue[m.queue.currentSong].title +
			" - " + m.queue.queue[m.queue.currentSong].artist
	} else {
		infoString = "No Song Playing!"
	}

	info := m.renderInfo(infoString, isFocused)
	status := m.renderStatus(isFocused)
	progress := m.renderProgress(isFocused)

	return lipgloss.JoinVertical(
		lipgloss.Center,
		info,
		status,
		progress,
	)
}

// Render Functions
func (m ControllerModel) renderInfo(info string, isFocused bool) string {
	styleInfo := lipgloss.NewStyle().
		Width(terminalWidth).
		MaxHeight(1).
		Background(lipgloss.Color("15")).
		Foreground(lipgloss.Color("0")).
		AlignHorizontal(lipgloss.Center)

	if isFocused {
		styleInfo = styleInfo.
			Background(colorFocus)
	}

	return styleInfo.Render(info)
}

func (m ControllerModel) renderStatus(isFocused bool) string {
	width := terminalWidth - 22

	property, _ := m.mpv.GetProperty("pause", mpv.FormatFlag)
	paused, _ := property.(bool)

	property, _ = m.mpv.GetProperty("volume", mpv.FormatInt64)
	volume, _ := property.(int64)

	icon := " "

	if volume <= 33 {
		icon = " "
	} else if volume <= 66 {
		icon = " "
	}

	leftText := ""
	rightText := fmt.Sprintf("%v %v%%", icon, volume)
	centerText := "|>"

	if paused {
		centerText = "||"
	}

	leftRender := lipgloss.NewStyle().
		Width(10).
		Render(leftText)

	rightRender := lipgloss.NewStyle().
		Width(10).
		AlignHorizontal(lipgloss.Right).
		Render(rightText)

	centerRender := lipgloss.NewStyle().
		Width(width -
			lipgloss.Width(leftRender) -
			lipgloss.Width(rightRender)).
		AlignHorizontal(lipgloss.Center).
		Render(centerText)

	styleStatus := lipgloss.NewStyle().
		Width(width)

	if isFocused {
		styleStatus = styleStatus.
			Foreground(colorFocus)
	}

	return styleStatus.
		Render(
			lipgloss.JoinHorizontal(
				lipgloss.Center,
				leftRender,
				centerRender,
				rightRender,
			),
		)
}

func (m ControllerModel) renderProgress(isFocused bool) string {
	property, _ := m.mpv.GetProperty("duration", mpv.FormatInt64)
	duration, _ := property.(int64)

	property, _ = m.mpv.GetProperty("time-pos", mpv.FormatInt64)
	position, _ := property.(int64)

	property, _ = m.mpv.GetProperty("percent-pos", mpv.FormatDouble)
	percentPos, _ := property.(float64)

	styleProgress := lipgloss.NewStyle().
		Width(terminalWidth).
		MaxHeight(1).
		AlignHorizontal(lipgloss.Center)

	styleTime := lipgloss.NewStyle().
		Width(11).
		AlignHorizontal(lipgloss.Center)

	m.progressBar.FullColor = "15"

	if isFocused {
		styleTime = styleTime.
			Foreground(colorFocus)
		m.progressBar.FullColor = "13"
	}

	positionRender := styleTime.Render(intToTime(int(position)))
	durationRender := styleTime.Render(intToTime(int(duration)))

	m.progressBar.Width = terminalWidth -
		lipgloss.Width(positionRender) -
		lipgloss.Width(durationRender)

	return styleProgress.Render(
		lipgloss.JoinHorizontal(
			lipgloss.Center,
			positionRender,
			m.progressBar.ViewAs(percentPos/100.0),
			durationRender,
		),
	)
}
