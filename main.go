package main

import (
	"fmt"
	"math"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
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

// Die Slow - Health
var songId = "CswcJyoHCNG9hsMuG8BMLm"
var songUrl = ""

type mpvEventMsg int64

func pollMpv(m *mpv.Mpv) tea.Cmd {
	return func() tea.Msg {
		_ = m.WaitEvent(10000)
		return mpvEventMsg(0)
	}
}

type playerModel struct {
	mpvPlayer *mpv.Mpv
}

func initializePlayerModel(m *mpv.Mpv) playerModel {
	return playerModel{
		mpvPlayer: m,
	}
}

func (p playerModel) Init() tea.Cmd {
	return pollMpv(p.mpvPlayer)
}

func (p playerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return p, tea.Quit
		}
	case mpvEventMsg:
		return p, pollMpv(p.mpvPlayer)
	}
	return p, nil
}

func (p playerModel) View() string {
	s := ""

	property, _ := p.mpvPlayer.GetProperty("time-pos", mpv.FormatInt64)
	progress, _ := property.(int64)

	current_progress := time.Duration(progress) * time.Second
	seconds := math.Floor(math.Mod(current_progress.Seconds(), 60))
	minutes := math.Floor(math.Mod(current_progress.Minutes(), 60))
	hours := math.Floor(current_progress.Hours())

	if hours > 0 {
		s += fmt.Sprintf("%vh ", hours)
	}

	if minutes > 0 {
		s += fmt.Sprintf("%vm ", minutes)
	}

	s += fmt.Sprintf("%02vs", seconds)

	return s
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

func createMPV() (*mpv.Mpv, error) {
	// Create MPV player
	m := mpv.New()

	// Observe time changes
	_ = m.ObserveProperty(0, "time-pos", mpv.FormatDouble)

	// Disable video (make sure by doing all 3 lmao)
	_ = m.SetOption("no-video", mpv.FormatFlag, true)
	_ = m.SetOptionString("vo", "null")
	_ = m.SetOptionString("vid", "")

	// Init player and return
	err := m.Initialize()

	if err != nil {
		return nil, err
	}

	return m, nil
}

func main() {
	// Setup bubbletea logging.
	f, err := tea.LogToFile("debug.log", "debug")

	if err != nil {
		panic(err)
	}

	defer f.Close()

	// Read Config
	err = readConfig()

	if err != nil {
		panic(err)
	}

	// Create song url
	songUrl = config.ServerUrl + "/rest/stream.view?u=" + config.ServerUser + "&p=" + config.ServerPassword + "&v=1.12.0&c=shanty&id=" + songId

	// Create MPV player
	m, err := createMPV()

	if err != nil {
		panic(err)
	}

	m.Command([]string{"loadfile", songUrl})

	p := tea.NewProgram(initializePlayerModel(m), tea.WithAltScreen())
	if _, err = p.Run(); err != nil {
		panic(err)
	}
}
