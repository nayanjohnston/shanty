package main

import (
	tea "github.com/charmbracelet/bubbletea"
)

// type Album {
// 	id    string
// 	title string
// 	image []string
// }
//
// var albumList []Album

// func getAlbumPage(page int64) {
// 	offset := 0 + (page * 6)
//
// 	result, err := http.Get(config.ServerUrl + "/rest/getAlbumList?u=" +
// 		config.ServerUser + "&p=" + config.ServerPassword +
// 		"&v=1.12.0&c=shanty&f=json&type=alphabeticalByArtist&size=6&offset=" +
// 		strconv.FormatInt(int64(offset), 10))
//
// 	if err != nil {
// 		panic(err)
// 	}
//
// 	body, _ := io.ReadAll(result.Body)
//
// 	var list any
// 	json.Unmarshal([]byte(body), &list)
//
// 	al, _ := list.(map[string]any)["subsonic-response"].(map[string]any)["albumList"].(map[string]any)["album"].([]any)
//
// 	for _, element := range al {
// 		albumImage, err := imageArray(element.(map[string]any)["coverArt"].(string))
//
// 		if err != nil {
// 			panic(err)
// 		}
//
// 		albumList = append(albumList, Album{
// 			title: element.(map[string]any)["title"].(string),
// 			id:    element.(map[string]any)["id"].(string),
// 			image: albumImage,
// 		})
// 	}
// }

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

	p := tea.NewProgram(initializePlayerModel(), tea.WithAltScreen())
	if _, err = p.Run(); err != nil {
		panic(err)
	}
}
