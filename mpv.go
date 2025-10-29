package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/gen2brain/go-mpv"
)

type PlayerManager struct {
	mp *mpv.Mpv
	pl *Playlist
}

type Song struct {
	url   string
	title string
	id    string
	art   []string
}

type Playlist struct {
	songs []Song
	index int
}

func createPlayer() PlayerManager {
	m, err := createMpv()

	if err != nil {
		panic(err)
	}

	return PlayerManager{
		mp: m,
		pl: &Playlist{
			index: 0,
		},
	}
}

func createMpv() (*mpv.Mpv, error) {
	// Create MPV player
	m := mpv.New()

	// Observe time changes
	_ = m.ObserveProperty(0, "time-pos", mpv.FormatDouble)

	// Disable video (make sure by doing all 3 lmao)
	_ = m.SetOption("no-video", mpv.FormatFlag, true)
	_ = m.SetOptionString("vo", "null")
	_ = m.SetOptionString("vid", "")

	// Init player and return
	err := m.Initialize()

	if err != nil {
		return nil, err
	}

	return m, nil
}

func (p PlayerManager) queueSong(songId string) {
	// Get URL for song information.
	infoUrl := config.ServerUrl + "/rest/getSong?u=" + config.ServerUser +
		"&p=" + config.ServerPassword + "&v=1.12.0&c=shanty&f=json&id=" + songId

	// Get http response.
	result, err := http.Get(infoUrl)
	if err != nil {
		panic(err)
	}

	// Read http body.
	body, err := io.ReadAll(result.Body)
	if err != nil {
		panic(err)
	}

	// Parse http resonse as json.
	var jsonResponse any
	json.Unmarshal([]byte(body), &jsonResponse)

	// Get the song information from json
	songInfo, ok := jsonResponse.(map[string]any)["subsonic-response"].(map[string]any)["song"].(map[string]any)
	if ok == false {
		return
	}

	// Get title
	var songTitle string = songInfo["title"].(string)

	songArt, err := imageArray(songInfo["coverArt"].(string))

	// Create the song's URL.
	songUrl := config.ServerUrl + "/rest/stream.view?u=" + config.ServerUser +
		"&p=" + config.ServerPassword + "&v=1.12.0&c=shanty&id=" + songId

	// Append it to the playlist.
	p.pl.songs = append(p.pl.songs, Song{
		url:   songUrl,
		id:    songId,
		title: songTitle,
		art:   songArt,
	})
}

func (p PlayerManager) loadSong(play bool) {
	// Check if we're in the range of the current playlist...
	if p.pl.index < 0 {
		return
	}

	if p.pl.index > len(p.pl.songs)-1 {
		return
	}

	// If so, load the songs URL in MPV.
	p.mp.Command([]string{
		"loadfile",
		p.pl.songs[p.pl.index].url,
	})

	// Set the "pause" property to what we've defined in function.
	p.mp.SetProperty("pause", mpv.FormatFlag, !play)
}

func (p PlayerManager) nextSong() {
	// If we're the last song in the playlist, go to the first song and pause.
	if p.pl.index >= len(p.pl.songs)-1 {
		p.pl.index = 0
		p.loadSong(false)

		log.Printf("Next (Wrapped): Song %v", p.pl.index)
		return
	}

	// Otherwise, go forward a song and reload.
	p.pl.index += 1
	p.loadSong(true)

	log.Printf("Next: Song %v", p.pl.index)
}

func (p PlayerManager) prevSong() {
	// Check if we are near the start of the current song...
	property, _ := p.mp.GetProperty("time-pos", mpv.FormatInt64)
	progress, _ := property.(int64)

	// If we are, and we're not the first song in the playlist, go back a song.
	if progress < 2 {
		if p.pl.index > 0 {
			p.pl.index -= 1
		}
	}

	// Then, reload the song.
	p.loadSong(true)

	log.Printf("Previous: Song %v", p.pl.index)
}
