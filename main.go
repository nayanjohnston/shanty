package main

import (
	tea "github.com/charmbracelet/bubbletea"
)

var player Player

func main() {
	// Read Config
	err := readConfig()

	if err != nil {
		panic(err)
	}

	// Create Player
	player = createPlayer()

	// getAlbumPage(19)

	// Setup bubbletea logging.
	f, err := tea.LogToFile("debug.log", "debug")

	if err != nil {
		panic(err)
	}

	defer f.Close()

	// Die Slow - Health
	player.queueSong("CswcJyoHCNG9hsMuG8BMLm")
	// Where Losers Go to Die - Intercourse
	player.queueSong("4GUzBDhXTurVnQmcM2DvOU")
	// Motherfucker, I Am Both_ “Amen” and “Hallelujah”… - Shearling
	player.queueSong("RuQ8j6ArKmWbVSbipoxcO1")

	player.loadSong(false)

	p := tea.NewProgram(
		mainModel{
			pModel: initializePlayerModel(),
		},
		tea.WithAltScreen(),
	)

	if _, err = p.Run(); err != nil {
		panic(err)
	}
}
