package main

import (
	"cmp"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"log"
	"math"
	"net/http"
	"slices"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type LibraryModel struct {
	loaded      bool
	library     []Album
	rows        int
	columns     int
	currentPage int
	selection   int

	sortModel tea.Model
	isSorting bool
}

type msgLibraryLoaded *[]Album
type msgAddAlbumToQueue *Album
type msgUpdatedLibrarySize struct {
	rows    int
	columns int
}
type msgChangePage int
type msgChangeSelection int
type msgSonglistLoaded struct {
	album    *Album
	songlist []*Song
}
type msgArtworkLoaded struct{}

func initLibraryModel() LibraryModel {
	return LibraryModel{
		loaded:      false,
		currentPage: 0,
		selection:   0,
		sortModel:   initSortModel(),
		isSorting:   false,
	}
}

func (m LibraryModel) Init() tea.Cmd {
	return tea.Batch(
		getLibrary(string(sortArtist)),
	)
}

func (m LibraryModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	cmds := []tea.Cmd{}

	if m.isSorting {
		m.sortModel, cmd = m.sortModel.Update(msg)
		cmds = append(cmds, cmd)

		switch msg.(type) {
		case msgSortSelect:
			m.isSorting = false
		}

		return m, tea.Batch(cmds...)
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		cmds = append(cmds, func() tea.Msg {
			albumWidth := albumArtWidth + 4
			albumHeight := albumArtHeight + 5

			columns := int(math.Floor(float64(contentWidth) / float64(albumWidth)))
			rows := int(math.Floor(float64(contentHeight-1) / float64(albumHeight)))

			return msgUpdatedLibrarySize{
				rows:    rows,
				columns: columns,
			}
		})
	case tea.KeyMsg:
		if !m.loaded {
			return m, nil
		}

		switch msg.String() {
		case "s":
			m.isSorting = !m.isSorting

		case "n":
			return m, func() tea.Msg {
				return msgChangePage(1)
			}
		case "p":
			return m, func() tea.Msg {
				return msgChangePage(-1)
			}

		case "h":
			return m, func() tea.Msg {
				return msgChangeSelection(-1)
			}
		case "l":
			return m, func() tea.Msg {
				return msgChangeSelection(1)
			}
		case "j":
			return m, func() tea.Msg {
				return msgChangeSelection(m.columns)
			}
		case "k":
			return m, func() tea.Msg {
				return msgChangeSelection(-m.columns)
			}

		case "enter":
			return m, func() tea.Msg {
				selectedAlbum := m.selection + (m.currentPage * (m.rows * m.columns))
				if selectedAlbum >= 0 && selectedAlbum < len(m.library) {
					return msgShowAlbum(&m.library[selectedAlbum])
				} else {
					return nil
				}
			}
		}
	case msgLibraryLoaded:
		m.library = *msg
		m.loaded = true

		for index, element := range m.library {
			cmds = append(cmds, func() tea.Msg {

				newArt, err := imageArray(element.artworkId)

				if err != nil {
					return msgErrorShouldPanic(err)
				}

				m.library[index].artwork = newArt
				return msgArtworkLoaded{}
			})
		}

		cmds = append(cmds, func() tea.Msg { return msgChangePage(0) })

		return m, tea.Sequence(cmds...)
	case msgSonglistLoaded:
		msg.album.songlist = msg.songlist
	case msgChangePage:
		m.currentPage += int(msg)
		maxPages := int(math.Ceil(float64(len(m.library)) / float64(m.columns*m.rows)))
		if m.currentPage >= maxPages {
			m.currentPage = maxPages - 1
		}
		if m.currentPage < 0 {
			m.currentPage = 0
		}

		cmds = append(cmds, func() tea.Msg {
			return msgChangeSelection(0)
		})
	case msgChangeSelection:
		m.selection += int(msg)
		newSelection := math.Min(float64(m.selection), float64((m.columns*m.rows)-1))
		newSelection = math.Min(newSelection, float64((len(m.library)-1)-
			(m.currentPage*(m.rows*m.columns))))
		newSelection = math.Max(newSelection, 0)

		m.selection = int(newSelection)
	case msgAddAlbumToQueue:
		if !m.loaded {
			return m, nil
		}

		var sequence []tea.Cmd

		for _, song := range msg.songlist {
			sequence = append(sequence, func() tea.Msg {
				return msgQueueAddSong{song: song}
			})
		}

		// If nothing is in the queue, just start playing.
		if len(globalQueue.songlist) == 0 {
			sequence = append(sequence, func() tea.Msg {
				return msgCtrlLoadSong{playNow: true}
			})
		}

		cmd = tea.Sequence(sequence...)
		cmds = append(cmds, cmd)
	case msgUpdatedLibrarySize:
		selectedAlbum := m.selection + (m.currentPage * (m.rows * m.columns))
		m.rows = msg.rows
		m.columns = msg.columns

		if m.rows*m.columns != 0 {
			m.currentPage = int(float64(selectedAlbum / (m.rows * m.columns)))

			m.selection = selectedAlbum - int(float64(m.currentPage)*
				float64(m.rows*m.columns))
		}
	}

	return m, tea.Batch(cmds...)
}

func (m LibraryModel) View() string {
	if !m.loaded {
		return lipgloss.NewStyle().
			Width(contentWidth).
			Height(contentHeight).
			AlignHorizontal(lipgloss.Center).
			AlignVertical(lipgloss.Center).
			Render("Loading library...\n(This may take a while.)")
	}

	firstIndex := (m.currentPage * (m.rows * m.columns))
	lastIndex := int(math.Min(
		float64(firstIndex+(m.rows*m.columns)),
		float64(len(m.library)),
	))

	output := ""
	row := ""

	for i := firstIndex; i < lastIndex; i++ {
		song := &m.library[i]
		index := i - firstIndex

		if index != 0 && math.Mod(float64(index), float64(m.columns)) == 0 {
			output = combineVertical(output, row)

			row = ""
		}

		isSelected := false

		if index == m.selection && currentMainFocus == contentFocus {
			isSelected = true
		}
		row = combineHorizontal(row, renderAlbum(song, isSelected))
	}

	if row != "" {
		output = combineVertical(output, row)
	}

	albumPageRender := lipgloss.NewStyle().
		AlignHorizontal(lipgloss.Center).
		AlignVertical(lipgloss.Center).
		Width(contentWidth).
		Height(contentHeight - 1).
		Render(output)

	result := lipgloss.JoinVertical(
		lipgloss.Center,
		albumPageRender,
		fmt.Sprintf("Page %v/%v", m.currentPage+1, math.Ceil(
			float64(len(m.library))/
				float64((m.rows*m.columns)))),
	)

	if m.isSorting {
		choiceRender := m.sortModel.View()

		row := (contentWidth / 2) - (lipgloss.Width(choiceRender) / 2)
		col := (contentHeight / 2) - (lipgloss.Height(choiceRender) / 2)

		overlayOutput, err := Overlay(result, choiceRender, col, row, true)
		if err == nil {
			result = overlayOutput
		}
	}

	return result
}

func getLibrary(sortMethod string) tea.Cmd {
	return func() tea.Msg {
		maxSize := 500
		currentPageNumber := 0
		shouldContinue := true

		var newLibrary []Album = []Album{}

		for shouldContinue {
			sizeString := fmt.Sprintf("%v", maxSize)
			pageString := fmt.Sprintf("%v", currentPageNumber*maxSize)

			result, err := http.Get(
				config.ServerUrl +
					"/rest/getAlbumList?" +
					"u=" + config.ServerUser +
					"&p=" + config.ServerPassword +
					"&v=1.12.0" +
					"&c=shanty" +
					"&f=json" +
					"&type=" + sortMethod +
					"&size=" + sizeString +
					"&offset=" + pageString,
			)
			if err != nil {
				return msgErrorShouldPanic(err)
			}

			resultBody, _ := io.ReadAll(result.Body)

			var list any
			json.Unmarshal([]byte(resultBody), &list)

			respSubsonic := list.(map[string]any)["subsonic-response"]
			respAlbumList := respSubsonic.(map[string]any)["albumList"]
			respAlbums, ok := respAlbumList.(map[string]any)["album"].([]any)

			// Most likely out of range, so just exit
			if !ok {
				shouldContinue = false
				continue
			}

			for _, element := range respAlbums {
				albumArtUrl := element.(map[string]any)["coverArt"].(string)

				newAlbum := Album{
					title:     element.(map[string]any)["title"].(string),
					artist:    element.(map[string]any)["artist"].(string),
					id:        element.(map[string]any)["id"].(string),
					year:      element.(map[string]any)["year"].(float64),
					artworkId: albumArtUrl,
					artwork:   []string{},
				}

				newLibrary = append(newLibrary, newAlbum)
			}

			currentPageNumber += 1
		}

		switch sortMethod {
		case "alphabeticalByArtist":
			slices.SortFunc(newLibrary, func(a, b Album) int {
				compareNames := cmp.Compare(
					strings.ToLower(a.artist),
					strings.ToLower(b.artist),
				)

				if compareNames == 0 {
					return cmp.Compare(
						a.year,
						b.year,
					)
				} else {
					return compareNames
				}
			})
		}

		return msgLibraryLoaded(&newLibrary)
	}
}

func getSonglist(album *Album) tea.Msg {
	if len(album.songlist) > 0 {
		return nil
	}

	log.Println("Getting songlist")
	newSonglist := []*Song{}

	albumUrl := config.ServerUrl +
		"/rest/getAlbum?" +
		"u=" + config.ServerUser +
		"&p=" + config.ServerPassword +
		"&v=1.12.0" +
		"&c=shanty" +
		"&f=json" +
		"&id=" + album.id

	result, err := http.Get(albumUrl)

	if err != nil {
		return msgErrorShouldPanic(
			errors.New("shanty: Cannot get Songlist for Album."),
		)
	}

	resultBody, _ := io.ReadAll(result.Body)

	var list any
	json.Unmarshal([]byte(resultBody), &list)

	respSubsonic := list.(map[string]any)["subsonic-response"]
	respAlbum := respSubsonic.(map[string]any)["album"]
	respSongs := respAlbum.(map[string]any)["song"].([]any)

	for _, element := range respSongs {
		songId := element.(map[string]any)["id"].(string)
		songTitle := element.(map[string]any)["title"].(string)
		songArtist := element.(map[string]any)["artist"].(string)
		songDuration := element.(map[string]any)["duration"].(float64)

		newSong := Song{
			id:       songId,
			title:    songTitle,
			artist:   songArtist,
			duration: songDuration,
			album:    album,
		}

		newSonglist = append(newSonglist, &newSong)
	}

	log.Println("Got songlist")
	return msgSonglistLoaded{
		album:    album,
		songlist: newSonglist,
	}
}

func renderAlbum(album *Album, isSelected bool) string {
	styleAlbum := lipgloss.NewStyle().
		Width(albumArtWidth+2).
		Padding(0, 1, 0, 1).
		Border(lipgloss.NormalBorder())

	styleArtwork := lipgloss.NewStyle().
		Width(albumArtWidth).
		Height(albumArtHeight).
		AlignHorizontal(lipgloss.Center)

	styleTitle := lipgloss.NewStyle().
		Width(albumArtWidth).
		MaxHeight(2).
		AlignHorizontal(lipgloss.Center)

	if isSelected {
		styleTitle = styleTitle.
			Foreground(colorFocus)
	}

	renderTitle := styleTitle.Render(album.title)

	styleArtist := lipgloss.NewStyle().
		Width(albumArtWidth).
		Height(3 - lipgloss.Height(renderTitle)).
		MaxHeight(3 - lipgloss.Height(renderTitle)).
		Foreground(lipgloss.Color("7")).
		AlignHorizontal(lipgloss.Center)

	if isSelected {
		styleArtist = styleArtist.
			Foreground(colorFocusDim)

		styleAlbum = styleAlbum.
			BorderForeground(colorFocus)
	}

	drawnArtwork := ""

	if len(album.artwork) != 0 {
		drawnArtwork = drawImage(album.artwork)
	}

	return styleAlbum.Render(
		lipgloss.JoinVertical(lipgloss.Center,
			styleArtwork.Render(drawnArtwork),
			renderTitle,
			styleArtist.Render(album.artist),
		))
}

func combineVertical(first string, second string) string {
	if first == "" {
		first = second
	} else {
		first = lipgloss.JoinVertical(
			lipgloss.Left,
			first,
			second,
		)
	}

	return first
}

func combineHorizontal(first string, second string) string {
	if first == "" {
		first = second
	} else {
		first = lipgloss.JoinHorizontal(
			lipgloss.Left,
			first,
			second,
		)
	}

	return first
}
