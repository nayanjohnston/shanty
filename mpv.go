package main

import (
	"log"

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
	songs  []Song
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

func (p PlayerManager) queueSong(song Song) {
	// Append it to the playlist.
	p.playlist.songs = append(p.playlist.songs, song)
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
