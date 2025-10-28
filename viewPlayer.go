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

var playerStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderTop(true)

var timeStyle = lipgloss.NewStyle().
	Align(lipgloss.Center).
	Width(11)

var infoStyle = lipgloss.NewStyle().
	Width(22)

type playerModel struct {
	progressBar progress.Model
	width       int
	height      int
}

type mpvEventMsg *mpv.Event

func initializePlayerModel() playerModel {
	prgs := progress.New(progress.WithDefaultGradient())
	prgs.ShowPercentage = false

	return playerModel{
		progressBar: prgs,
	}
}

func (p playerModel) Init() tea.Cmd {
	return nil
}

func (p playerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case " ":
			player.mp.Command([]string{"cycle", "pause"})
		case "ctrl+h":
			player.mp.Command([]string{"seek", "-5", "relative"})
		case "ctrl+l":
			player.mp.Command([]string{"seek", "5", "relative"})
		case "ctrl+p":
			player.prevSong()
		case "ctrl+n":
			player.nextSong()
		case "ctrl+k":
			player.mp.Command([]string{"add", "volume", "5"})
		case "ctrl+j":
			player.mp.Command([]string{"add", "volume", "-5"})
		}

	case mpvEventMsg:
		var e mpv.Event = *msg

		switch e.EventID {
		case mpv.EventEnd:
			if e.EndFile().Reason == mpv.EndFileEOF {
				log.Println("Song Ended.")
				player.nextSong()
			}
		}
		cmds = append(cmds, pollMpv(player.mp))
	}

	return p, tea.Batch(cmds...)
}

func (p playerModel) getLengthString() string {
	s := ""

	property, _ := player.mp.GetProperty("duration", mpv.FormatInt64)
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

func (p playerModel) getPositionString() string {
	s := ""

	property, _ := player.mp.GetProperty("time-pos", mpv.FormatInt64)
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

func (p playerModel) getPausedString() string {
	s := ""

	property, _ := player.mp.GetProperty("pause", mpv.FormatFlag)
	paused, _ := property.(bool)

	if paused {
		s = "||"
	} else {
		s = "|>"
	}

	return s
}

func (p playerModel) getVolumeString() string {
	s := ""

	property, _ := player.mp.GetProperty("volume", mpv.FormatInt64)
	volume, _ := property.(int64)

	s = fmt.Sprintf("%v%%", volume)

	return s
}

func (p playerModel) View() string {
	// Get played percentage.
	property, _ := player.mp.GetProperty("percent-pos", mpv.FormatDouble)
	percentPos, _ := property.(float64)

	// Create timer and information strings.
	positionString := p.getPositionString()
	remainingString := p.getLengthString()
	pausedString := p.getPausedString()
	volumeString := p.getVolumeString()

	// Render time strings with style.
	positionRender := timeStyle.Render(positionString)
	remainingRender := timeStyle.Render(remainingString)

	// Set progress bar width.
	p.progressBar.Width = p.width -
		lipgloss.Width(positionRender) -
		lipgloss.Width(remainingRender)

	// Create progress render.
	progressRender := lipgloss.JoinHorizontal(
		lipgloss.Center,

		positionRender,
		p.progressBar.ViewAs(percentPos/100.0),
		remainingRender,
	)

	// Create renders for information row.
	infoLeftRender := infoStyle.
		AlignHorizontal(lipgloss.Left).
		Render("")

	infoRightRender := infoStyle.
		AlignHorizontal(lipgloss.Right).
		Render(volumeString)

	infoCenterRender := infoStyle.
		AlignHorizontal(lipgloss.Center).
		Width(p.width -
			lipgloss.Width(infoLeftRender) -
			lipgloss.Width(infoRightRender) -
			lipgloss.Width(positionRender) -
			lipgloss.Width(remainingRender),
		).
		Render(pausedString)

	informationRender := lipgloss.JoinHorizontal(
		lipgloss.Center,

		infoLeftRender,
		infoCenterRender,
		infoRightRender,
	)

	output := playerStyle.
		Render(
			lipgloss.JoinVertical(
				lipgloss.Center,

				informationRender,
				progressRender,
			),
		)

	// Place player at the bottom of the terminal.
	return lipgloss.Place(
		p.width,
		lipgloss.Height(output),
		lipgloss.Center,
		lipgloss.Bottom,
		output,
	)
}
