package main

type ModelAlbum struct {
	currentAlbum Album
}

// TODO: A view to show after an album is pressed, should have information about
// the album with a tracklist for users to add songs to the queue. The user
// should also be able to add a whole album to the queue (clearing the queue
// before obviously).
//
// +-----+ | TITLE - ARTIST
// |ALBUM| | play|queue
// | ART | | 1. title
// +-----+ | etc...
