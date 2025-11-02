package main

import (
	"github.com/gen2brain/go-mpv"
)

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
	title    string
	artist   string
	year     float64
	songlist []*Song
	artwork  []string
	id       string
}

type Queue struct {
	queue       []*Song
	currentSong int
}
