package main

import (
	"github.com/gen2brain/go-mpv"
)

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
