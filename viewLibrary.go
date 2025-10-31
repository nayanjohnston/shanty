package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"strconv"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	pageRows    int = 2
	pageColumns int = 3
)

type ModelLibrary struct {
	currentPageNumber  int
	currentPageContent []Album
	library            []Album
	selectedAlbum      int
}

type msgLibraryLoaded []Album

var styleTitleUnfocused = lipgloss.NewStyle().
	MaxWidth(22).
	Width(22).
	MaxHeight(2).
	Padding(0, 1, 0, 1).
	AlignHorizontal(lipgloss.Center)

var styleTitleFocused = styleTitleUnfocused.
	Foreground(colorFocus)

var styleArtistUnfocused = lipgloss.NewStyle().
	MaxWidth(22).
	Width(22).
	MaxHeight(1).
	Padding(0, 1, 0, 1).
	Foreground(lipgloss.Color("7")).
	AlignHorizontal(lipgloss.Center)

var styleArtistFocused = styleArtistUnfocused.
	Foreground(colorFocusDim)

var styleArt = lipgloss.NewStyle().
	Width(albumArtWidth).
	Height(albumArtHeight)

var styleAlbumUnfocused = lipgloss.NewStyle().
	MaxWidth(24).
	Width(22).
	MaxHeight(14).
	Height(12).
	Border(lipgloss.NormalBorder())

var styleAlbumFocused = styleAlbumUnfocused.
	BorderForeground(colorFocus)

func initializeModelLibrary() ModelLibrary {
	modelLibrary := ModelLibrary{
		currentPageNumber: 0,
		selectedAlbum:     0,
	}

	return modelLibrary
}

func (l ModelLibrary) Init() tea.Cmd {
	return getLibrary()
}

func (l ModelLibrary) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "l":
			l.selectedAlbum = l.changeSelection(1)
			return l, nil
		case "h":
			l.selectedAlbum = l.changeSelection(-1)
			return l, nil
		case "j":
			l.selectedAlbum = l.changeSelection(pageColumns)
			return l, nil
		case "k":
			l.selectedAlbum = l.changeSelection(-pageColumns)
			return l, nil
		case "n":
			l.currentPageNumber = l.changePage(1)
			l.selectedAlbum = 0
			l.currentPageContent = l.getAlbumPage(l.currentPageNumber)
			return l, nil
		case "p":
			l.currentPageNumber = l.changePage(-1)
			l.selectedAlbum = 0
			l.currentPageContent = l.getAlbumPage(l.currentPageNumber)
			return l, nil
		case "enter":
			objectPlayer.queue.songs = []Song{}
			objectPlayer.queue.index = 0
			for _, song := range l.currentPageContent[l.selectedAlbum].songs {
				objectPlayer.queueSong(song)
			}
			objectPlayer.loadSong(true)
		}
	case msgLibraryLoaded:
		l.library = msg
		l.currentPageContent = l.getAlbumPage(l.currentPageNumber)
		return l, nil
	}

	return l, nil
}

func (l ModelLibrary) View() string {
	// If library isn't loaded, then don't display anything.
	if len(l.library) == 0 {
		return "LOADING LIBRARY"
	}

	outputString := ""
	rowString := ""

	// Create album grid with current page
	for index, element := range l.currentPageContent {
		// Create new row if full
		if index != 0 && math.Mod(float64(index), 3) == 0 {
			outputString = lipgloss.JoinVertical(
				lipgloss.Left,
				outputString,
				rowString,
			)

			rowString = ""
		}

		styleAlbum := styleAlbumUnfocused
		styleArtist := styleArtistUnfocused
		styleTitle := styleTitleUnfocused

		if index == l.selectedAlbum && currentFocus == focusLibrary {
			styleAlbum = styleAlbumFocused
			styleArtist = styleArtistFocused
			styleTitle = styleTitleFocused
		}

		albumDisplay := styleAlbum.Render(
			lipgloss.JoinVertical(
				lipgloss.Center,
				styleArt.Render(drawImage(element.art)),
				styleTitle.Render(truncateText(element.title, 40)),
				styleArtist.Render(truncateText(element.artist, 20)),
			),
		)

		rowString = lipgloss.JoinHorizontal(
			lipgloss.Top,
			rowString,
			albumDisplay,
		)
	}

	outputString = lipgloss.JoinVertical(lipgloss.Left,
		outputString,
		rowString,
	)

	rowString = ""

	outputString = lipgloss.JoinVertical(lipgloss.Center,
		outputString,
		fmt.Sprintf("Page %v/%v", l.currentPageNumber+1, l.getPageAmount()),
	)

	return outputString
}

func getLibrary() tea.Cmd {
	return func() tea.Msg {
		maxSize := 500
		currentPageNumber := 0
		shouldContinue := true

		var newLibrary []Album = []Album{}

		for shouldContinue {
			sizeString := strconv.FormatInt(int64(maxSize), 10)
			pageString := strconv.FormatInt(int64(currentPageNumber*maxSize), 10)

			result, err := http.Get(
				config.ServerUrl + "/rest/getAlbumList?u=" +
					config.ServerUser + "&p=" + config.ServerPassword +
					"&v=1.12.0&c=shanty&f=json&type=alphabeticalByArtist&size=" +
					sizeString + "&offset=" + pageString,
			)

			if err != nil {
				panic(err)
			}

			resultBody, _ := io.ReadAll(result.Body)

			var list any
			json.Unmarshal([]byte(resultBody), &list)

			albumList, ok := list.(map[string]any)["subsonic-response"].(map[string]any)["albumList"].(map[string]any)["album"].([]any)

			// Most likely out of range, so just exit
			if ok == false {
				shouldContinue = false
				continue
			}

			for _, element := range albumList {
				albumArt, err := imageArray(element.(map[string]any)["coverArt"].(string))

				if err != nil {
					panic(err)
				}

				newLibrary = append(newLibrary, Album{
					title:  element.(map[string]any)["title"].(string),
					artist: element.(map[string]any)["artist"].(string),
					id:     element.(map[string]any)["id"].(string),
					art:    albumArt,
					songs:  getTracklist(element.(map[string]any)["id"].(string)),
				})
			}

			currentPageNumber += 1
		}

		return msgLibraryLoaded(newLibrary)
	}
}

func getTracklist(albumId string) []Song {
	newTracklist := []Song{}

	albumUrl := config.ServerUrl + "/rest/getAlbum?u=" + config.ServerUser +
		"&p=" + config.ServerPassword + "&v=1.12.0&c=shanty&f=json&id=" + albumId

	result, err := http.Get(albumUrl)

	if err != nil {
		panic(err)
	}

	resultBody, _ := io.ReadAll(result.Body)

	var list any
	json.Unmarshal([]byte(resultBody), &list)

	songsList, _ := list.(map[string]any)["subsonic-response"].(map[string]any)["album"].(map[string]any)["song"].([]any)

	for _, element := range songsList {
		songId := element.(map[string]any)["id"].(string)
		songTitle := element.(map[string]any)["title"].(string)
		songArtist := element.(map[string]any)["artist"].(string)
		songDuration := element.(map[string]any)["duration"].(float64)

		log.Printf("%v", songDuration)

		songUrl := config.ServerUrl + "/rest/download.view?u=" + config.ServerUser +
			"&p=" + config.ServerPassword + "&v=1.12.0&c=shanty&f=json&id=" + songId
		newSong := Song{
			id:       songId,
			url:      songUrl,
			title:    songTitle,
			artist:   songArtist,
			duration: songDuration,
		}

		newTracklist = append(newTracklist, newSong)
	}
	return newTracklist
}

func (l ModelLibrary) getAlbumPage(page int) []Album {
	albumPage := []Album{}
	albumIndexStart := page * (pageRows * pageColumns)
	albumIndexEnd := (page + 1) * (pageRows * pageColumns)

	albumIndex := albumIndexStart

	for albumIndex < albumIndexEnd && albumIndex < len(l.library) {
		albumPage = append(albumPage, l.library[albumIndex])
		albumIndex += 1
	}

	return albumPage
}

func (l ModelLibrary) changeSelection(amount int) int {
	newSelection := l.selectedAlbum + amount

	if newSelection < 0 {
		return l.selectedAlbum
	} else if newSelection > len(l.currentPageContent)-1 {
		return l.selectedAlbum
	}

	return newSelection
}

func (l ModelLibrary) changePage(amount int) int {
	newPage := l.currentPageNumber + amount

	if newPage > l.getPageAmount()-1 {
		return l.currentPageNumber
	} else if newPage < 0 {
		return l.currentPageNumber
	}

	return newPage
}

func (l ModelLibrary) getPageAmount() int {
	return int(math.Ceil(
		float64(len(l.library)) /
			float64(pageRows*pageColumns),
	))
}
