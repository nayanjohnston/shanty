package main

import (
	"encoding/json"
	"fmt"
	"io"
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

// Styles
var styleTitleUnfocused = lipgloss.NewStyle().
	MaxWidth(albumArtWidth).
	Width(albumArtWidth).
	MaxHeight(2).
	AlignHorizontal(lipgloss.Center)

var styleTitleFocused = styleTitleUnfocused.
	Foreground(colorFocus)

var styleArtistUnfocused = lipgloss.NewStyle().
	MaxWidth(albumArtWidth).
	Width(albumArtWidth).
	MaxHeight(1).
	Foreground(lipgloss.Color("7")).
	AlignHorizontal(lipgloss.Center)

var styleArtistFocused = styleArtistUnfocused.
	Foreground(colorFocusDim)

var styleArt = lipgloss.NewStyle().
	Width(albumArtWidth).
	Height(albumArtHeight)

var styleAlbumUnfocused = lipgloss.NewStyle().
	MaxWidth(albumArtWidth+4).
	Width(albumArtWidth+2).
	Padding(0, 1, 0, 1).
	MaxHeight(albumArtHeight + 5).
	Height(albumArtHeight + 3).
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
	case tea.WindowSizeMsg:
		l = l.changePage(0)
		return l, nil
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
			l = l.changePage(1)
			return l, nil
		case "p":
			l = l.changePage(-1)
			return l, nil
		case "enter":
			objectPlayer.queue.songs = []Song{}
			objectPlayer.queue.index = 0

			var cmds []tea.Cmd

			for _, song := range l.currentPageContent[l.selectedAlbum].tracklist {
				objectPlayer.queueSong(song)
			}
			objectPlayer.loadSong(true)

			return l, tea.Sequence(cmds...)
		}
	case msgLibraryLoaded:
		l.library = msg
		l = l.changePage(0)
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
		if index != 0 && math.Mod(float64(index), float64(pageColumns)) == 0 {
			if index == pageColumns {
				outputString = rowString
			} else {
				outputString = lipgloss.JoinVertical(
					lipgloss.Left,
					outputString,
					rowString,
				)
			}

			rowString = ""
		}

		// Create album styles.
		styleAlbum := styleAlbumUnfocused
		styleArtist := styleArtistUnfocused
		styleTitle := styleTitleUnfocused

		// Change them to focused if so.
		if index == l.selectedAlbum && focusedView == focusMain {
			styleAlbum = styleAlbumFocused
			styleArtist = styleArtistFocused
			styleTitle = styleTitleFocused
		}

		// Create render of information
		albumDisplay := styleAlbum.Render(
			lipgloss.JoinVertical(
				lipgloss.Center,
				styleArt.Render(drawImage(element.art)),
				styleTitle.Render(truncateText(element.title, albumArtWidth*2)),
				styleArtist.Render(truncateText(element.artist, albumArtWidth)),
			),
		)

		// Add it to horizontal string
		rowString = lipgloss.JoinHorizontal(
			lipgloss.Top,
			rowString,
			albumDisplay,
		)
	}

	// If there's some leftovers...
	if outputString != "" && rowString != "" {
		// If there's already a row, just add it on.
		outputString = lipgloss.JoinVertical(lipgloss.Left,
			outputString,
			rowString,
		)
	} else if outputString == "" {
		// Otherwise, just set output to row
		outputString = rowString
	}

	// Render One: The albums from left to right, positioned on the top left
	// corner.
	outputString = lipgloss.NewStyle().
		Height(pageRows * (albumArtHeight + 5)).
		Width(pageColumns * (albumArtWidth + 4)).
		AlignVertical(lipgloss.Top).
		AlignHorizontal(lipgloss.Left).
		Render(outputString)

	// Render Two: The output of the previous, but centered in the space
	// available
	outputString = lipgloss.NewStyle().
		Height(sizeMainHeight - 1).
		AlignVertical(lipgloss.Center).
		Render(outputString)

	// Add page count to bottom of view.
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

			respSubsonic, _ := list.(map[string]any)["subsonic-response"]
			respAlbumList, _ := respSubsonic.(map[string]any)["albumList"]
			respAlbums, ok := respAlbumList.(map[string]any)["album"].([]any)

			// Most likely out of range, so just exit
			if ok == false {
				shouldContinue = false
				continue
			}

			for _, element := range respAlbums {
				albumArtUrl := element.(map[string]any)["coverArt"].(string)
				albumArt, err := imageArray(albumArtUrl)

				if err != nil {
					panic(err)
				}

				newAlbum := Album{
					title:  element.(map[string]any)["title"].(string),
					artist: element.(map[string]any)["artist"].(string),
					id:     element.(map[string]any)["id"].(string),
					art:    albumArt,
				}

				newAlbum.tracklist = getTracklist(&newAlbum)

				newLibrary = append(newLibrary, newAlbum)
			}

			currentPageNumber += 1
		}

		return msgLibraryLoaded(newLibrary)
	}
}

func getTracklist(album *Album) []Song {
	newTracklist := []Song{}

	albumUrl := config.ServerUrl + "/rest/getAlbum?u=" + config.ServerUser +
		"&p=" + config.ServerPassword + "&v=1.12.0&c=shanty&f=json&id=" + album.id

	result, err := http.Get(albumUrl)

	if err != nil {
		panic(err)
	}

	resultBody, _ := io.ReadAll(result.Body)

	var list any
	json.Unmarshal([]byte(resultBody), &list)

	respSubsonic, _ := list.(map[string]any)["subsonic-response"]
	respAlbum, _ := respSubsonic.(map[string]any)["album"]
	respSongs, _ := respAlbum.(map[string]any)["song"].([]any)

	for _, element := range respSongs {
		songId := element.(map[string]any)["id"].(string)
		songTitle := element.(map[string]any)["title"].(string)
		songArtist := element.(map[string]any)["artist"].(string)
		songDuration := element.(map[string]any)["duration"].(float64)

		songUrl := config.ServerUrl + "/rest/download.view?u=" + config.ServerUser +
			"&p=" + config.ServerPassword + "&v=1.12.0&c=shanty&f=json&id=" + songId

		newSong := Song{
			id:       songId,
			url:      songUrl,
			title:    songTitle,
			artist:   songArtist,
			duration: songDuration,
			album:    album,
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

// This is also used for dynamically changing the amount of rows/columns on
// screen (It just changes the page by an amount of 0).
func (l ModelLibrary) changePage(amount int) ModelLibrary {
	// Get current album focused.
	albumIndex := (l.currentPageNumber * (pageColumns * pageRows)) + l.selectedAlbum

	// Calculate columns and rows.
	pageColumns = int(math.Floor(float64(terminalWidth / (albumArtWidth + 4))))
	pageRows = int(math.Floor(float64((sizeMainHeight - 1) / (albumArtHeight + 5))))

	// Switch to page with selected album.
	l.currentPageNumber = int(math.Floor(
		float64(albumIndex / (pageColumns * pageRows)),
	))

	l.selectedAlbum = albumIndex -
		int(float64(l.currentPageNumber)*
			float64(pageColumns*pageRows))

	// Change page
	l.currentPageNumber += amount

	// Clamp pages
	if l.currentPageNumber > l.getPageAmount()-1 {
		l.currentPageNumber = int(math.Max(0, float64(l.getPageAmount()-1)))
	} else if l.currentPageNumber < 0 {
		l.currentPageNumber = 0
	}

	// Get content
	l.currentPageContent = l.getAlbumPage(l.currentPageNumber)

	// Update selection if needed
	l.selectedAlbum = int(math.Min(
		math.Max(
			float64(l.selectedAlbum),
			0,
		),
		float64(len(l.currentPageContent)-1),
	))

	return l
}

func (l ModelLibrary) getPageAmount() int {
	return int(math.Ceil(
		float64(len(l.library)) /
			float64(pageRows*pageColumns),
	))
}
