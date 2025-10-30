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
	pageIndex      int
	pageAmount     int
	libraryContent []Album
	albumSelection int
}

type msgLibraryLoaded []Album

var styleTitle = lipgloss.NewStyle().
	MaxWidth(22).
	Width(22).
	MaxHeight(3).
	Height(3).
	Padding(0, 1, 0, 1).
	AlignHorizontal(lipgloss.Center)

var styleArt = lipgloss.NewStyle().
	Width(albumArtWidth).
	Height(albumArtHeight)

var styleAlbumUnfocused = lipgloss.NewStyle().
	Border(lipgloss.NormalBorder())

var styleAlbumFocused = styleAlbumUnfocused.
	BorderForeground(lipgloss.Color("99"))

func initializeModelLibrary() ModelLibrary {
	modelLibrary := ModelLibrary{
		pageIndex:      0,
		albumSelection: 0,
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
			l.albumSelection = l.changeSelection(1)
			return l, nil
		case "h":
			l.albumSelection = l.changeSelection(-1)
			return l, nil
		case "j":
			l.albumSelection = l.changeSelection(pageColumns)
			return l, nil
		case "k":
			l.albumSelection = l.changeSelection(-pageColumns)
			return l, nil
		case "n":
			l.pageIndex = l.changePage(1)
			return l, nil
		case "p":
			l.pageIndex = l.changePage(-1)
			return l, nil
		}
	case msgLibraryLoaded:
		l.libraryContent = msg

		l.pageAmount = int(math.Ceil(
			float64(len(l.libraryContent)) /
				float64(pageRows*pageColumns),
		))

		return l, nil
	}

	return l, nil
}

func (l ModelLibrary) View() string {
	outputString := ""
	rowString := ""

	if len(l.libraryContent) == 0 {
		outputString += "LOADING LIBRARY"
	} else {

		for index, element := range l.getAlbumPage(l.pageIndex) {
			if index != 0 {
				if math.Mod(float64(index), 3) == 0 {
					outputString = lipgloss.JoinVertical(lipgloss.Left,
						outputString,
						rowString,
					)

					rowString = ""
				}
			}

			styleAlbum := styleAlbumUnfocused

			if index == l.albumSelection {
				styleAlbum = styleAlbumFocused
			}

			albumDisplay := styleAlbum.
				Render(
					lipgloss.JoinVertical(lipgloss.Center,
						styleArt.Render(drawImage(element.art)),
						styleTitle.Render(element.title),
					),
				)

			rowString = lipgloss.JoinHorizontal(lipgloss.Top,
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
			fmt.Sprintf("Page %v/%v", l.pageIndex+1, l.pageAmount),
		)
	}

	return outputString
}

func getLibrary() tea.Cmd {
	return func() tea.Msg {
		maxSize := 500
		currentPage := 0
		shouldContinue := true

		var newLibrary []Album = []Album{}

		for shouldContinue {
			sizeString := strconv.FormatInt(int64(maxSize), 10)
			pageString := strconv.FormatInt(int64(currentPage*maxSize), 10)

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
					title: element.(map[string]any)["title"].(string),
					id:    element.(map[string]any)["id"].(string),
					art:   albumArt,
				})
			}

			currentPage += 1
		}

		return msgLibraryLoaded(newLibrary)
	}
}

func (l ModelLibrary) getAlbumPage(page int) []Album {
	albumPage := []Album{}
	albumIndexStart := page * (pageRows * pageColumns)
	albumIndexEnd := (page + 1) * (pageRows * pageColumns)

	albumIndex := albumIndexStart

	for albumIndex < albumIndexEnd && albumIndex < len(l.libraryContent) {
		albumPage = append(albumPage, l.libraryContent[albumIndex])
		albumIndex += 1
	}

	return albumPage
}

func (l ModelLibrary) changeSelection(amount int) int {
	newSelection := l.albumSelection + amount

	if newSelection < 0 {
		return l.albumSelection
	} else if newSelection > (pageRows*pageColumns)-1 {
		return l.albumSelection
	}

	return newSelection
}

func (l ModelLibrary) changePage(amount int) int {
	newPage := l.pageIndex + amount

	if newPage > l.pageAmount {
		return l.pageIndex
	} else if newPage < 0 {
		return l.pageIndex
	}

	return newPage
}
