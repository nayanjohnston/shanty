package main

import (
	"log"

	"github.com/gen2brain/go-mpv"
)

type Player struct {
	mp *mpv.Mpv
	pl *Playlist
}

type Song struct {
	url  string
	name string
	id   string
	art  string
}

type Playlist struct {
	songs []Song
	index int
}

func createPlayer() Player {
	m, err := createMPV()

	if err != nil {
		panic(err)
	}

	return Player{
		mp: m,
		pl: &Playlist{
			index: 0,
		},
	}
}

func createMPV() (*mpv.Mpv, error) {
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

func (p Player) queueSong(songId string) {
	// Create the song's URL.
	songUrl := config.ServerUrl + "/rest/stream.view?u=" + config.ServerUser +
		"&p=" + config.ServerPassword + "&v=1.12.0&c=shanty&id=" + songId

	// Append it to the playlist.
	p.pl.songs = append(p.pl.songs, Song{
		url: songUrl,
		id:  songId,
	})
}

func (p Player) loadSong(play bool) {
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

func (p Player) nextSong() {
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

func (p Player) prevSong() {
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
