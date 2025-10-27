package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	tea "github.com/charmbracelet/bubbletea"
)

var (
	// Motherfucker, I Am Both_ “Amen” and “Hallelujah”… - Shearling
	songId = "RuQ8j6ArKmWbVSbipoxcO1"
	// Where Losers Go to Die - Intercourse
	//songId = "4GUzBDhXTurVnQmcM2DvOU"
	// Die Slow - Health
	//songId  = "CswcJyoHCNG9hsMuG8BMLm"
	songUrl = ""
)

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
