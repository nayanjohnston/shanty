package main

import (
	"slices"

	"github.com/gen2brain/go-mpv"
)

var globalMpv *mpv.Mpv = initMpv()
var globalQueue Queue = Queue{currentSong: 0}

func initMpv() *mpv.Mpv {
	// Create mpv player
	m := mpv.New()

	// Observe time changes
	_ = m.ObserveProperty(0, "time-pos", mpv.FormatDouble)

	// Disable video
	_ = m.SetOption("no-video", mpv.FormatFlag, true)
	_ = m.SetOptionString("vo", "null")
	_ = m.SetOptionString("vid", "")

	// Start player and return
	err := m.Initialize()
	if err != nil {
		panic(err)
	}

	return m
}

// Song definition
type Song struct {
	title    string
	artist   string
	album    *Album
	duration float64
	id       string
}

func (s Song) getUrl() string {
	return config.ServerUrl +
		"/rest/download.view?" +
		"u=" + config.ServerUser +
		"&p=" + config.ServerPassword +
		"&v=1.12.0" +
		"&c=shanty" +
		"&f=json" +
		"&id=" + s.id
}

// Album Definition
type Album struct {
	title     string
	artist    string
	year      float64
	songlist  []*Song
	artwork   []string
	artworkId string
	id        string
}

type Queue struct {
	songlist    []*Song
	currentSong int
}

func (q *Queue) getCurrentSong() *Song {
	return q.songlist[q.currentSong]
}

func (q *Queue) addSong(song *Song) {
	q.songlist = append(q.songlist, song)
}

func (q *Queue) removeSong(position int) {
	// If songlist is empty, ignore.
	if len(q.songlist) == 0 {
		return
	}

	// If current song is either last in list, or after a song we're deleting,
	// move back.
	if q.currentSong == len(q.songlist)-1 || q.currentSong > position {
		q.updatePosition(-1)
	}

	q.songlist = slices.Delete(q.songlist, position, position+1)
}

func (q *Queue) moveSong(from int, to int) {
	// If moving out of bounds, ignore
	if to < 0 || to >= len(q.songlist) {
		return
	}

	song := q.songlist[from]

	// Update current position.
	if q.currentSong == from {
		q.updatePosition(to - from)
	} else {
		if from < q.currentSong {
			q.updatePosition(-1)
		}
		if to < q.currentSong {
			q.updatePosition(1)
		}
	}

	q.songlist = slices.Delete(q.songlist, from, from+1)
	q.songlist = slices.Insert(q.songlist, to, song)
}

func (q *Queue) updatePosition(amount int) {
	q.currentSong += amount

	if q.currentSong >= len(q.songlist) {
		q.currentSong = len(q.songlist) - 1
	}

	// Do this after, because if songlist is empty, the above will make it -1,
	// which can cause headaches later.
	if q.currentSong < 0 {
		q.currentSong = 0
	}
}

func (q *Queue) clearQueue() {
	q.songlist = []*Song{}
	q.currentSong = 0
}
