package main

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	focusPlayer focusedModel = iota
	focusLibrary
)

const (
	colorFocus    = lipgloss.Color("13")
	colorFocusDim = lipgloss.Color("5")
)

var (
	objectPlayer   PlayerManager
	terminalWidth  int
	terminalHeight int
	currentFocus   focusedModel
)

func main() {
	// Read Config
	err := readConfig()

	if err != nil {
		panic(err)
	}

	// Create Player
	objectPlayer = createPlayer()

	// getAlbumPage(19)

	// Setup bubbletea logging.
	f, err := tea.LogToFile("debug.log", "debug")

	if err != nil {
		panic(err)
	}

	defer f.Close()

	objectPlayer.loadSong(false)

	p := tea.NewProgram(
		ModelMain{
			modelControls: initializeModelControls(),
			modelLibrary:  initializeModelLibrary(),
		},
		tea.WithAltScreen(),
	)

	currentFocus = focusPlayer

	if _, err = p.Run(); err != nil {
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
