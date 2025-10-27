package main

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gen2brain/go-mpv"
	"github.com/pelletier/go-toml/v2"
)

// Config object
type ShantyConfig struct {
	ServerUrl      string
	ServerUser     string
	ServerPassword string
}

var config ShantyConfig

var (
	// Motherfucker, I Am Both_ “Amen” and “Hallelujah”… - Shearling
	songId = "RuQ8j6ArKmWbVSbipoxcO1"
	// Where Losers Go to Die - Intercourse
	//songId = "4GUzBDhXTurVnQmcM2DvOU"
	// Die Slow - Health
	//songId  = "CswcJyoHCNG9hsMuG8BMLm"
	songUrl = ""
)

type playerModel struct {
	width       int
	height      int
	mpvPlayer   *mpv.Mpv
	progressBar progress.Model
}

type mpvEventMsg *mpv.Event

func pollMpv(m *mpv.Mpv) tea.Cmd {
	return func() tea.Msg {
		return mpvEventMsg(m.WaitEvent(10000))
	}
}

func initializePlayerModel(m *mpv.Mpv) playerModel {
	prgs := progress.New(progress.WithDefaultGradient())
	prgs.ShowPercentage = false

	return playerModel{
		mpvPlayer:   m,
		progressBar: prgs,
	}
}

func (p playerModel) Init() tea.Cmd {
	return pollMpv(p.mpvPlayer)
}

func (p playerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		p.width = msg.Width
		p.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return p, tea.Quit
		case " ":
			p.mpvPlayer.Command([]string{"cycle", "pause"})
		case "left", "h":
			p.mpvPlayer.Command([]string{"seek", "-5", "relative"})
		case "right", "l":
			p.mpvPlayer.Command([]string{"seek", "5", "relative"})
		case "up", "k":
			p.mpvPlayer.Command([]string{"add", "volume", "5"})
		case "down", "j":
			p.mpvPlayer.Command([]string{"add", "volume", "-5"})
		}
	case mpvEventMsg:
		var e mpv.Event = *msg

		switch e.EventID {
		case mpv.EventEnd:
			if e.EndFile().Reason == mpv.EndFileEOF {
				p.mpvPlayer.Command([]string{"loadfile", songUrl})
			}
		}

		return p, pollMpv(p.mpvPlayer)
	}
	return p, nil
}

func (p playerModel) getLengthString() string {
	s := ""

	property, _ := p.mpvPlayer.GetProperty("duration", mpv.FormatInt64)
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

	property, _ := p.mpvPlayer.GetProperty("time-pos", mpv.FormatInt64)
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

	property, _ := p.mpvPlayer.GetProperty("pause", mpv.FormatFlag)
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

	property, _ := p.mpvPlayer.GetProperty("volume", mpv.FormatInt64)
	volume, _ := property.(int64)

	s = fmt.Sprintf("%v%%", volume)

	return s
}

func (p playerModel) View() string {
	property, _ := p.mpvPlayer.GetProperty("percent-pos", mpv.FormatDouble)
	percentPos, _ := property.(float64)

	leftText := p.getPositionString()
	centerText := p.getPausedString() + " - " + p.getVolumeString()
	rightText := p.getLengthString()

	leftRender := lipgloss.NewStyle().Align(lipgloss.Center).Width(11).Render(leftText)
	rightRender := lipgloss.NewStyle().Align(lipgloss.Center).Width(11).Render(rightText)

	p.progressBar.Width = p.width - lipgloss.Width(leftRender) - lipgloss.Width(rightRender)

	return lipgloss.Place(
		p.width,
		p.height,
		lipgloss.Center,
		lipgloss.Bottom,
		lipgloss.JoinVertical(
			lipgloss.Center,
			centerText,
			lipgloss.JoinHorizontal(lipgloss.Center,
				leftRender,
				p.progressBar.ViewAs(percentPos/100.0),
				rightRender,
			),
		),
	)
}

func readConfig() error {
	// Read config data
	configData, err := os.ReadFile("config.toml")

	if err != nil {
		return err
	}

	// Convert toml to config object
	err = toml.Unmarshal([]byte(configData), &config)

	return nil
}

func printAlbumList(offset int64) string {
	s := ""

	offset = offset * 16

	result, err := http.Get(config.ServerUrl + "/rest/getAlbumList?u=" +
		config.ServerUser + "&p=" + config.ServerPassword +
		"&v=1.12.0&c=shanty&f=json&type=alphabeticalByArtist&size=16&offset=" +
		strconv.FormatInt(offset, 10))

	body, _ := io.ReadAll(result.Body)

	if err != nil {
		panic(err)
	}

	var list any

	json.Unmarshal([]byte(body), &list)

	subsonicResponse := list.(map[string]any)["subsonic-response"]
	albumListContainer := subsonicResponse.(map[string]any)["albumList"]
	albumList := albumListContainer.(map[string]any)["album"].([]any)

	for index, element := range albumList {
		s += element.(map[string]any)["title"].(string)
		if index != 15 {
			s += "\n"
		}
	}

	return s
}

func main() {
	fmt.Println("Reading config...")

	// Read Config
	err := readConfig()

	if err != nil {
		panic(err)
	}

	// Create song url
	songUrl = config.ServerUrl + "/rest/stream.view?u=" + config.ServerUser + "&p=" + config.ServerPassword + "&v=1.12.0&c=shanty&id=" + songId

	fmt.Println("Initializing MPV...")

	// Create MPV player
	m, err := createMPV()

	if err != nil {
		panic(err)
	}

	fmt.Println("Setting up TUI...")

	// Setup bubbletea logging.
	f, err := tea.LogToFile("debug.log", "debug")

	if err != nil {
		panic(err)
	}

	defer f.Close()

	m.Command([]string{"loadfile", songUrl})

	fmt.Println(printAlbumList(0))

	p := tea.NewProgram(initializePlayerModel(m), tea.WithAltScreen())
	if _, err = p.Run(); err != nil {
		panic(err)
	}
}
