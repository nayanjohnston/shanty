package main

import (
	tea "github.com/charmbracelet/bubbletea"
)

const (
	focusPlayer focusedModel = iota
	focusLibrary
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

	// Die Slow - Health
	objectPlayer.queueSong("CswcJyoHCNG9hsMuG8BMLm")
	// Where Losers Go to Die - Intercourse
	objectPlayer.queueSong("4GUzBDhXTurVnQmcM2DvOU")
	// Motherfucker, I Am Both_ “Amen” and “Hallelujah”… - Shearling
	objectPlayer.queueSong("RuQ8j6ArKmWbVSbipoxcO1")
	// Shame - Chat Pile
	objectPlayer.queueSong("lBAiNj6W8XhQos1hwWWPv1")

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
