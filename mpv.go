package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/gen2brain/go-mpv"
)

type PlayerManager struct {
	mpv      *mpv.Mpv
	playlist *Playlist
}

type Song struct {
	url    string
	title  string
	artist string
	id     string
}

type Album struct {
	id     string
	art    []string
	title  string
	artist string
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
		mpv: m,
		playlist: &Playlist{
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
	var songArtist string = songInfo["artist"].(string)

	// songArt, err := imageArray(songInfo["coverArt"].(string))

	// Create the song's URL.
	songUrl := config.ServerUrl + "/rest/stream.view?u=" + config.ServerUser +
		"&p=" + config.ServerPassword + "&v=1.12.0&c=shanty&id=" + songId

	// Append it to the playlist.
	p.playlist.songs = append(p.playlist.songs, Song{
		url:    songUrl,
		id:     songId,
		title:  songTitle,
		artist: songArtist,
		// art:   songArt,
	})
}

func (p PlayerManager) loadSong(play bool) {
	// Check if we're in the range of the current playlist...
	if p.playlist.index < 0 {
		return
	}

	if p.playlist.index > len(p.playlist.songs)-1 {
		return
	}

	// If so, load the songs URL in MPV.
	p.mpv.Command([]string{
		"loadfile",
		p.playlist.songs[p.playlist.index].url,
	})

	// Set the "pause" property to what we've defined in function.
	p.mpv.SetProperty("pause", mpv.FormatFlag, !play)
}

func (p PlayerManager) nextSong() {
	// If we're the last song in the playlist, go to the first song and pause.
	if p.playlist.index >= len(p.playlist.songs)-1 {
		p.playlist.index = 0
		p.loadSong(false)

		log.Printf("Next (Wrapped): Song %v", p.playlist.index)
		return
	}

	// Otherwise, go forward a song and reload.
	p.playlist.index += 1
	p.loadSong(true)

	log.Printf("Next: Song %v", p.playlist.index)
}

func (p PlayerManager) prevSong() {
	// Check if we are near the start of the current song...
	property, _ := p.mpv.GetProperty("time-pos", mpv.FormatInt64)
	progress, _ := property.(int64)

	// If we are, and we're not the first song in the playlist, go back a song.
	if progress < 2 {
		if p.playlist.index > 0 {
			p.playlist.index -= 1
		}
	}

	// Then, reload the song.
	p.loadSong(true)

	log.Printf("Previous: Song %v", p.playlist.index)
}
