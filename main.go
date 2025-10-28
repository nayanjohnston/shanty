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

func printAlbumList() string {
	s := ""
	keepGoing := true
	offset := 0

	for keepGoing {
		result, err := http.Get(config.ServerUrl + "/rest/getAlbumList?u=" +
			config.ServerUser + "&p=" + config.ServerPassword +
			"&v=1.12.0&c=shanty&f=json&type=alphabeticalByArtist&size=100&offset=" +
			strconv.FormatInt(int64(offset), 10))

		if err != nil {
			panic(err)
		}

		body, _ := io.ReadAll(result.Body)

		var list any
		json.Unmarshal([]byte(body), &list)

		albumList, ok := list.(map[string]any)["subsonic-response"].(map[string]any)["albumList"].(map[string]any)["album"].([]any)

		if ok == false {
			keepGoing = false
			continue
		}

		for _, element := range albumList {
			s += element.(map[string]any)["title"].(string) + " | " +
				element.(map[string]any)["id"].(string) + "\n"
		}

		offset += 100
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

	imar := imageArray("al-0Un3SBpJkfJEx0bk1zxBjQ_68863ba3")

	fmt.Println("----------------------")

	for index, row := range imar {
		if index == len(imar)-1 {
			continue
		}

		fmt.Println("|" + row + "|")
	}

	fmt.Println("----------------------")

	imar = imageArray("al-6bq2WyuPx4OiRggyHSpFZj_6884f0bc")

	fmt.Println("----------------------")

	for index, row := range imar {
		if index == len(imar)-1 {
			continue
		}

		fmt.Println("|" + row + "|")
	}

	fmt.Println("----------------------")

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

	fmt.Println(printAlbumList())

	p := tea.NewProgram(initializePlayerModel(m), tea.WithAltScreen())
	if _, err = p.Run(); err != nil {
		panic(err)
	}
}
