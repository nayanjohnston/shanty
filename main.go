package main

import (
	"fmt"
	"math"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	focusPlayer focusView = iota
	focusMain
)

const (
	focusLibrary focusModel = iota
	focusQueue
)

const (
	colorFocus    = lipgloss.Color("13")
	colorFocusDim = lipgloss.Color("5")
)

var (
	objectPlayer   PlayerManager
	objectProgram  *tea.Program
	terminalWidth  int
	terminalHeight int
	focusedView    focusView
	focusedModel   focusModel
)

func main() {
	// Read Config
	err := readConfig()

	if err != nil {
		panic(err)
	}

	// Create Player
	objectPlayer = createPlayer()

	// Setup bubbletea logging.
	f, err := tea.LogToFile("debug.log", "debug")

	if err != nil {
		panic(err)
	}

	defer f.Close()

	objectPlayer.loadSong(false)

	objectProgram = tea.NewProgram(
		initializeModelMain(),
		tea.WithAltScreen(),
	)

	focusedView = focusPlayer
	focusedModel = focusLibrary

	if _, err := objectProgram.Run(); err != nil {
		panic(err)
	}
}

// Truncate a string, adding ellipsis (...), to fit within maxLength
func truncateText(s string, max int) string {
	if max > len(s) {
		return s
	}

	lastIndex := strings.LastIndexAny(s[:max-1], " .,,:;-")

	if lastIndex < 0 {
		return s[:max] + "…"
	}

	return s[:lastIndex] + "…"
}

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
