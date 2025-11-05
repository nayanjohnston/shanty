package main

import (
	"fmt"
	"log"
	"math"
	"net/http"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gen2brain/go-mpv"
)

type LoopMode int

const (
	loopNone LoopMode = iota
	loopQueue
	loopSong
)

// Model Definition
type ControllerModel struct {
	progressBar progress.Model
	scrobble    bool
	loopMode    LoopMode
}

type msgMpvEvent *mpv.Event
type msgLoadSong struct{ playNow bool }
type msgNextSong struct{}
type msgPrevSong struct{}
type msgStopPlayback struct{}
type msgSetScrobbled bool
type msgSwitchLoopMode struct{}

// Model Initialisation
func initControllerModel() ControllerModel {
	newProgressBar := progress.New()
	newProgressBar.ShowPercentage = false

	controllerModel := ControllerModel{
		progressBar: newProgressBar,
		loopMode:    loopNone,
	}

	return controllerModel
}

// Tea Model Functions
func (m ControllerModel) Init() tea.Cmd {
	return tea.Batch(
		func() tea.Msg { return msgMpvEvent(globalMpv.WaitEvent(10000)) },
	)
}

func (m ControllerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case " ":
			globalMpv.Command([]string{"cycle", "pause"})
			cmds = append(cmds, checkScrobble(&m, 0, false))
		case "h":
			globalMpv.Command([]string{"seek", "-5", "relative"})
		case "l":
			globalMpv.Command([]string{"seek", "5", "relative"})
		case "k":
			globalMpv.Command([]string{"add", "volume", "5"})
		case "j":
			globalMpv.Command([]string{"add", "volume", "-5"})
		case "n":
			cmds = append(cmds, func() tea.Msg { return msgNextSong{} })
		case "p":
			cmds = append(cmds, func() tea.Msg { return msgPrevSong{} })
		case "r":
			cmds = append(cmds, func() tea.Msg { return msgSwitchLoopMode{} })
		}
	case msgMpvEvent:
		var e mpv.Event = *msg

		switch e.EventID {
		case mpv.EventPropertyChange:
			prop := e.Property()
			data, ok := prop.Data.(float64)

			// Check property name/ok status
			if prop.Name == "time-pos" && ok {
				// Check if we've reached scrobbling theshold
				if !m.scrobble && config.ShouldScrobble {
					cmd = checkScrobble(&m, data, true)
					cmds = append(cmds, cmd)
				}
			}
		case mpv.EventEnd:
			if e.EndFile().Reason == mpv.EndFileEOF {
				cmd = func() tea.Msg { return msgNextSong{} }
				cmds = append(cmds, cmd)
			}
		}

		cmd = func() tea.Msg { return msgMpvEvent(globalMpv.WaitEvent(10000)) }
		cmds = append(cmds, cmd)
	case msgLoadSong:
		// Ignore if songlist is empty.
		if len(globalQueue.songlist) == 0 {
			break
		}

		// Load via mpv
		globalMpv.Command([]string{
			"loadfile",
			globalQueue.getCurrentSong().getUrl(),
		})

		globalMpv.SetProperty("pause", mpv.FormatFlag, !msg.playNow)

		m.scrobble = false
		return m, checkScrobble(&m, 0, false)

	case msgNextSong:
		shouldPlay := true

		switch m.loopMode {
		case loopNone:
			// If we're the last song in the queue, go to the first song and pause.
			if globalQueue.currentSong >= len(globalQueue.songlist)-1 {
				globalQueue.currentSong = 0
				shouldPlay = false
			} else {
				globalQueue.updatePosition(1)
			}
		case loopQueue:
			// If we're the last song in the queue, go to the first song.
			if globalQueue.currentSong >= len(globalQueue.songlist)-1 {
				globalQueue.currentSong = 0
			} else {
				globalQueue.updatePosition(1)
			}
		case loopSong:
			// Do nothing.
		}

		cmds = append(cmds, func() tea.Msg {
			return msgLoadSong{playNow: shouldPlay}
		})
	case msgPrevSong:
		property, _ := globalMpv.GetProperty("time-pos", mpv.FormatInt64)
		position, _ := property.(int64)

		if position <= 5 && globalQueue.currentSong > 0 {
			globalQueue.updatePosition(-1)
		}

		cmds = append(cmds, func() tea.Msg {
			return msgLoadSong{playNow: true}
		})
	case msgStopPlayback:
		globalMpv.Command([]string{"stop"})
	case msgSetScrobbled:
		m.scrobble = bool(msg)
	case msgSwitchLoopMode:
		switch m.loopMode {
		case loopNone:
			m.loopMode = loopQueue
		case loopQueue:
			m.loopMode = loopSong
		case loopSong:
			m.loopMode = loopNone
		}
	}

	return m, tea.Batch(cmds...)
}

func (m ControllerModel) View() string {
	isFocused := false

	if currentMainFocus == controllerFocus {
		isFocused = true
	}

	infoString := ""

	if len(globalQueue.songlist) > 0 {
		infoString = globalQueue.getCurrentSong().title +
			" - " + globalQueue.getCurrentSong().artist
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

	property, _ := globalMpv.GetProperty("pause", mpv.FormatFlag)
	paused, _ := property.(bool)

	property, _ = globalMpv.GetProperty("volume", mpv.FormatInt64)
	volume, _ := property.(int64)

	volumeIcon := " "

	if volume <= 33 {
		volumeIcon = " "
	} else if volume <= 66 {
		volumeIcon = " "
	}

	loopIcon := "󰑗 "

	switch m.loopMode {
	case loopQueue:
		loopIcon = "󰑖 "
	case loopSong:
		loopIcon = "󰑘 "
	}

	leftText := loopIcon
	rightText := fmt.Sprintf("%v %v%%", volumeIcon, volume)
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
	property, _ := globalMpv.GetProperty("duration", mpv.FormatInt64)
	duration, _ := property.(int64)

	property, _ = globalMpv.GetProperty("time-pos", mpv.FormatInt64)
	position, _ := property.(int64)

	property, _ = globalMpv.GetProperty("percent-pos", mpv.FormatDouble)
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

func checkScrobble(m *ControllerModel, currentPos float64, submission bool) tea.Cmd {
	return func() tea.Msg {
		if !config.ShouldScrobble {
			return nil
		}

		song := globalQueue.getCurrentSong()

		if song.duration < 30 {
			m.scrobble = true
			return nil
		}

		if submission {
			scrobbleThresh := float64(song.duration) / 2
			scrobbleThresh = math.Min(scrobbleThresh, 240)

			if currentPos > scrobbleThresh {
				log.Printf("Scrobbling!")

				http.Get(
					config.ServerUrl +
						"/rest/scrobble?" +
						"u=" + config.ServerUser +
						"&p=" + config.ServerPassword +
						"&v=1.12.0" +
						"&c=shanty" +
						"&f=json" +
						"&id=" + song.id +
						"&submission=True")

				return msgSetScrobbled(true)
			}
		} else {
			property, _ := globalMpv.GetProperty("pause", mpv.FormatFlag)
			paused, _ := property.(bool)

			if paused {
				return nil
			}

			log.Printf("Sending nowPlaying")
			http.Get(
				config.ServerUrl +
					"/rest/scrobble?" +
					"u=" + config.ServerUser +
					"&p=" + config.ServerPassword +
					"&v=1.12.0" +
					"&c=shanty" +
					"&f=json" +
					"&id=" + song.id +
					"&submission=False")
		}

		return nil
	}
}
