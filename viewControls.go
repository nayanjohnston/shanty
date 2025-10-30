package main

import (
	"fmt"
	"log"
	"math"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gen2brain/go-mpv"
)

// Styles
var styleOutputUnfocused = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderTop(true)

var styleOutputFocused = styleOutputUnfocused.
	BorderForeground(colorFocus)

var styleTimeUnfocused = lipgloss.NewStyle().
	Align(lipgloss.Center).
	Width(11)

var styleTimeFocused = styleTimeUnfocused.
	Foreground(colorFocus)

var styleInfoUnfocused = lipgloss.NewStyle().
	Width(22)

var styleInfoFocused = styleInfoUnfocused.
	Foreground(colorFocus)

type ModelControls struct {
	modelProgressBar progress.Model
}

type msgMpvEvent *mpv.Event

func initializeModelControls() ModelControls {
	newProgressBar := progress.New(progress.WithDefaultGradient())
	newProgressBar.ShowPercentage = false

	return ModelControls{
		modelProgressBar: newProgressBar,
	}
}

func (p ModelControls) Init() tea.Cmd {
	return nil
}

func (p ModelControls) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case " ":
			objectPlayer.mpv.Command([]string{"cycle", "pause"})
		case "h":
			objectPlayer.mpv.Command([]string{"seek", "-5", "relative"})
		case "l":
			objectPlayer.mpv.Command([]string{"seek", "5", "relative"})
		case "p":
			objectPlayer.prevSong()
		case "n":
			objectPlayer.nextSong()
		case "k":
			objectPlayer.mpv.Command([]string{"add", "volume", "5"})
		case "j":
			objectPlayer.mpv.Command([]string{"add", "volume", "-5"})
		}

	case msgMpvEvent:
		var e mpv.Event = *msg

		switch e.EventID {
		case mpv.EventEnd:
			if e.EndFile().Reason == mpv.EndFileEOF {
				log.Println("Song Ended.")
				objectPlayer.nextSong()
			}
		}
		cmds = append(cmds, awaitMpvEvent(objectPlayer.mpv))
	}

	return p, tea.Batch(cmds...)
}

func (p ModelControls) View() string {
	styleTime := styleTimeUnfocused
	styleInfo := styleInfoUnfocused
	styleOutput := styleOutputUnfocused

	if currentFocus == focusPlayer {
		styleTime = styleTimeFocused
		styleInfo = styleInfoFocused
		styleOutput = styleOutputFocused
	}

	// Get played percentage.
	property, _ := objectPlayer.mpv.GetProperty("percent-pos", mpv.FormatDouble)
	percentPos, _ := property.(float64)

	// Create timer and controls strings.
	positionString := p.getPositionString()
	remainingString := p.getLengthString()
	pausedString := p.getPausedString()
	volumeString := p.getVolumeString()

	// Render time strings with style.
	renderPosition := styleTime.Render(positionString)
	renderRemaining := styleTime.Render(remainingString)

	// Set progress bar width.
	p.modelProgressBar.Width = terminalWidth -
		lipgloss.Width(renderPosition) -
		lipgloss.Width(renderRemaining)

	// Create progress render.
	renderProgress := lipgloss.JoinHorizontal(
		lipgloss.Center,

		renderPosition,
		p.modelProgressBar.ViewAs(percentPos/100.0),
		renderRemaining,
	)

	// Create renders for information row.
	renderControlsLeft := styleInfo.
		AlignHorizontal(lipgloss.Left).
		Render(volumeString)

	renderControlsRight := styleInfo.
		AlignHorizontal(lipgloss.Right).
		Render("")

	renderControlsCenter := styleInfo.
		AlignHorizontal(lipgloss.Center).
		Width(terminalWidth -
			lipgloss.Width(renderControlsLeft) -
			lipgloss.Width(renderControlsRight) -
			lipgloss.Width(renderPosition) -
			lipgloss.Width(renderRemaining),
		).
		Render(pausedString)

	renderControls := lipgloss.JoinHorizontal(
		lipgloss.Center,

		renderControlsLeft,
		renderControlsCenter,
		renderControlsRight,
	)

	output := styleOutput.Render(
		lipgloss.JoinVertical(
			lipgloss.Center,

			renderControls,
			renderProgress,
		),
	)

	return output
}

func awaitMpvEvent(m *mpv.Mpv) tea.Cmd {
	return func() tea.Msg {
		return msgMpvEvent(m.WaitEvent(10000))
	}
}

func (p ModelControls) getLengthString() string {
	s := ""

	property, _ := objectPlayer.mpv.GetProperty("duration", mpv.FormatInt64)
	length, _ := property.(int64)

	current_progress := time.Duration(length) * time.Second
	seconds := math.Floor(math.Mod(current_progress.Seconds(), 60))
	minutes := math.Floor(math.Mod(current_progress.Minutes(), 60))
	hours := math.Floor(current_progress.Hours())

	if hours > 0 {
		s += fmt.Sprintf("%v:", hours)
	}

	s += fmt.Sprintf("%02v:", minutes)
	s += fmt.Sprintf("%02v", seconds)

	return s
}

func (p ModelControls) getPositionString() string {
	s := ""

	property, _ := objectPlayer.mpv.GetProperty("time-pos", mpv.FormatInt64)
	progress, _ := property.(int64)

	current_progress := time.Duration(progress) * time.Second
	seconds := math.Floor(math.Mod(current_progress.Seconds(), 60))
	minutes := math.Floor(math.Mod(current_progress.Minutes(), 60))
	hours := math.Floor(current_progress.Hours())

	if hours > 0 {
		s += fmt.Sprintf("%v:", hours)
	}

	s += fmt.Sprintf("%02v:", minutes)
	s += fmt.Sprintf("%02v", seconds)

	return s
}

func (p ModelControls) getPausedString() string {
	s := ""

	property, _ := objectPlayer.mpv.GetProperty("pause", mpv.FormatFlag)
	paused, _ := property.(bool)

	if paused {
		s = "||"
	} else {
		s = "|>"
	}

	return s
}

func (p ModelControls) getVolumeString() string {
	s := ""

	property, _ := objectPlayer.mpv.GetProperty("volume", mpv.FormatInt64)
	volume, _ := property.(int64)

	// Hacky way to keep the icon stable when volume changes.
	icon := " "

	if volume <= 33 {
		icon = " "
	} else if volume <= 66 {
		icon = " "
	}

	s = fmt.Sprintf("%v %v%%", icon, volume)

	return s
}
