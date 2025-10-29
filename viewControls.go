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

var styleControls = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderTop(true)

var styleTime = lipgloss.NewStyle().
	Align(lipgloss.Center).
	Width(11)

var styleTopRow = lipgloss.NewStyle().
	Width(22)

type modelControls struct {
	modelProgressBar progress.Model
	isFocused        bool
}

type msgMpvEvent *mpv.Event

func initializeModelControls() modelControls {
	newProgressBar := progress.New(progress.WithDefaultGradient())
	newProgressBar.ShowPercentage = false

	return modelControls{
		modelProgressBar: newProgressBar,
	}
}

func (p modelControls) Init() tea.Cmd {
	return nil
}

func (p modelControls) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case " ":
			objectPlayer.mp.Command([]string{"cycle", "pause"})
		case "ctrl+h":
			objectPlayer.mp.Command([]string{"seek", "-5", "relative"})
		case "ctrl+l":
			objectPlayer.mp.Command([]string{"seek", "5", "relative"})
		case "ctrl+p":
			objectPlayer.prevSong()
		case "ctrl+n":
			objectPlayer.nextSong()
		case "ctrl+k":
			objectPlayer.mp.Command([]string{"add", "volume", "5"})
		case "ctrl+j":
			objectPlayer.mp.Command([]string{"add", "volume", "-5"})
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
		cmds = append(cmds, awaitMpvEvent(objectPlayer.mp))
	}

	return p, tea.Batch(cmds...)
}

func (p modelControls) getLengthString() string {
	s := ""

	property, _ := objectPlayer.mp.GetProperty("duration", mpv.FormatInt64)
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

func (p modelControls) getPositionString() string {
	s := ""

	property, _ := objectPlayer.mp.GetProperty("time-pos", mpv.FormatInt64)
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

func (p modelControls) getPausedString() string {
	s := ""

	property, _ := objectPlayer.mp.GetProperty("pause", mpv.FormatFlag)
	paused, _ := property.(bool)

	if paused {
		s = "||"
	} else {
		s = "|>"
	}

	return s
}

func (p modelControls) getVolumeString() string {
	s := ""

	property, _ := objectPlayer.mp.GetProperty("volume", mpv.FormatInt64)
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

func (p modelControls) View() string {
	// Get played percentage.
	property, _ := objectPlayer.mp.GetProperty("percent-pos", mpv.FormatDouble)
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
	renderControlsLeft := styleTopRow.
		AlignHorizontal(lipgloss.Left).
		Render(volumeString)

	renderControlsRight := styleTopRow.
		AlignHorizontal(lipgloss.Right).
		Render("")

	renderControlsCenter := styleTopRow.
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

	output := styleControls.
		Render(
			lipgloss.JoinVertical(
				lipgloss.Center,

				renderControls,
				renderProgress,
			),
		)

	return output
}
