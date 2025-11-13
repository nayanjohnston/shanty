package main

import (
	"fmt"
	"math"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Color definitions
const (
	colorFocus    = lipgloss.Color("13")
	colorFocusDim = lipgloss.Color("5")
)

// Main focus (for MainModel)
type MainFocus int
type ContentFocus int

const (
	contentFocus MainFocus = iota
	controllerFocus

	libraryFocus ContentFocus = iota
	queueFocus
	albumFocus
)

var (
	config *ShantyConfig

	terminalWidth  int
	terminalHeight int

	contentWidth  int
	contentHeight int

	currentMainFocus    MainFocus    = controllerFocus
	currentContentFocus ContentFocus = libraryFocus

	globalProgram *tea.Program
)

func main() {
	// Read Config
	config = &ShantyConfig{}
	err := readConfig(config)
	if err != nil {
		fmt.Println("CONFIG FILE NOT FOUND!\nPlease create a config file in \"~/.config/shanty/config.toml\".")
		return
	}

	// Setup bubbletea logging
	tempFolder := os.TempDir()

	f, err := tea.LogToFile(tempFolder+"/shanty.log", "debug")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	// Create bubbletea program
	globalProgram = tea.NewProgram(
		initMainModel(),
		tea.WithAltScreen(),
	)

	if _, err := globalProgram.Run(); err != nil {
		panic(err)
	}
}

// Converts seconds as an interger into a readable string
func intToTime(sec int) string {
	s := ""
	current_progress := time.Duration(sec) * time.Second
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
