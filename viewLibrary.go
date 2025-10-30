package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"

	tea "github.com/charmbracelet/bubbletea"
)

var (
	pageRows    int = 2
	pageColumns int = 3
)

type ModelLibrary struct {
	pageIndex   int
	pageContent []Album
}

type msgAlbumPageLoaded []Album

func initializeModelLibrary() ModelLibrary {
	modelLibrary := ModelLibrary{
		pageIndex: 0,
	}

	return modelLibrary
}

func (l ModelLibrary) Init() tea.Cmd {
	return getAlbumPage(l.pageIndex)
}

func (l ModelLibrary) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "l":
			l.pageIndex += 1
			return l, getAlbumPage(l.pageIndex)
		case "h":
			l.pageIndex -= 1
			return l, getAlbumPage(l.pageIndex)
		}
	case msgAlbumPageLoaded:
		l.pageContent = msg
		return l, nil
	}

	return l, nil
}

func (l ModelLibrary) View() string {
	if len(l.pageContent) > 0 {
		return drawImage(l.pageContent[0].art)
	} else {
		return ""
	}
}

func getAlbumPage(page int) tea.Cmd {
	return func() tea.Msg {
		offset := 0 + (page * (pageRows * pageColumns))

		offsetString := strconv.FormatInt(int64(offset), 10)
		sizeString := strconv.FormatInt(int64(pageRows*pageColumns), 10)

		result, err := http.Get(
			config.ServerUrl + "/rest/getAlbumList?u=" +
				config.ServerUser + "&p=" + config.ServerPassword +
				"&v=1.12.0&c=shanty&f=json&type=alphabeticalByArtist&size=" +
				sizeString + "&offset=" + offsetString,
		)

		if err != nil {
			panic(err)
		}

		body, _ := io.ReadAll(result.Body)

		var list any
		json.Unmarshal([]byte(body), &list)

		al, _ := list.(map[string]any)["subsonic-response"].(map[string]any)["albumList"].(map[string]any)["album"].([]any)

		newAlbumList := []Album{}

		for _, element := range al {
			albumArt, err := imageArray(element.(map[string]any)["coverArt"].(string))

			if err != nil {
				panic(err)
			}

			newAlbumList = append(newAlbumList, Album{
				title: element.(map[string]any)["title"].(string),
				id:    element.(map[string]any)["id"].(string),
				art:   albumArt,
			})

			log.Println(element.(map[string]any)["title"].(string))
		}

		return msgAlbumPageLoaded(newAlbumList)
	}
}
