package main

import (
	"log"
	"os"

	"github.com/gen2brain/go-mpv"
	"github.com/godbus/dbus/v5"
	"github.com/quarckster/go-mpris-server/pkg/events"
	"github.com/quarckster/go-mpris-server/pkg/server"
	"github.com/quarckster/go-mpris-server/pkg/types"
)

var globalMpris *server.Server = initMpris()
var globalMprisEventHandler events.EventHandler = *events.NewEventHandler(globalMpris)

type Root struct{}
type Player struct{}
type LoopStatus struct{}

var (
	_ types.OrgMprisMediaPlayer2Adapter                 = Root{}
	_ types.OrgMprisMediaPlayer2PlayerAdapter           = Player{}
	_ types.OrgMprisMediaPlayer2PlayerAdapterLoopStatus = Player{}
)

// Implement other methods of `pkg.types.OrgMprisMediaPlayer2Adapter`
func (r Root) Raise() error {
	log.Println("Raised")
	return nil
}

func (r Root) Quit() error {
	return nil
}
func (r Root) CanQuit() (bool, error) {
	return true, nil
}
func (r Root) CanRaise() (bool, error) {
	return true, nil
}
func (r Root) HasTrackList() (bool, error) {
	return false, nil
}
func (r Root) Identity() (string, error) {
	return "shanty", nil
}
func (r Root) SupportedUriSchemes() ([]string, error) {
	return []string{}, nil
}
func (r Root) SupportedMimeTypes() ([]string, error) {
	return []string{}, nil
}

func (p Player) Next() error {
	globalProgram.Send(msgCtrlChangeSong{amount: 1})
	return nil
}
func (p Player) Previous() error {
	globalProgram.Send(msgCtrlChangeSong{amount: -1})
	return nil
}
func (p Player) Pause() error {
	globalProgram.Send(msgCtrlSetPaused{paused: true})
	return nil
}
func (p Player) PlayPause() error {
	var paused bool

	property, err := globalMpv.GetProperty("pause", mpv.FormatFlag)
	if err == nil {
		paused, _ = property.(bool)
	}

	globalProgram.Send(msgCtrlSetPaused{paused: !paused})
	return nil
}
func (p Player) Stop() error {
	globalProgram.Send(msgCtrlStop{})
	return nil
}
func (p Player) Play() error {
	globalProgram.Send(msgCtrlSetPaused{paused: false})
	return nil
}
func (p Player) Seek(offset types.Microseconds) error {
	return nil
}
func (p Player) SetPosition(trackId string, position types.Microseconds) error {
	seconds := microsecondsToSeconds(position)
	globalProgram.Send(msgCtrlSeek{
		amount:   int64(seconds),
		seekType: "absolute",
	})
	return nil
}
func (p Player) OpenUri(uri string) error {
	return nil
}
func (p Player) PlaybackStatus() (types.PlaybackStatus, error) {
	if globalMpv == nil {
		return types.PlaybackStatusStopped, nil
	}

	var paused bool
	property, err := globalMpv.GetProperty("pause", mpv.FormatFlag)
	if err == nil {
		paused, _ = property.(bool)
	}

	if paused {
		return types.PlaybackStatusPaused, nil
	} else {
		return types.PlaybackStatusPlaying, nil
	}
}
func (p Player) Rate() (float64, error) {
	return 1.2, nil
}
func (p Player) SetRate(float64) error {
	return nil
}
func (p Player) Metadata() (types.Metadata, error) {
	if len(globalQueue.songlist) == 0 {
		return types.Metadata{
			TrackId: "/org/mpris/MediaPlayer2/TrackList/NoTrack",
		}, nil
	}

	userCache, err := os.UserCacheDir()

	if err != nil {
		return types.Metadata{
			TrackId: "/org/mpris/MediaPlayer2/TrackList/NoTrack",
		}, err
	}

	track_id := "/shanty/album/" + globalQueue.getCurrentSong().album.id +
		"/track/" + globalQueue.getCurrentSong().id
	return types.Metadata{
		TrackId: dbus.ObjectPath(track_id),
		Length:  secondsToMicroseconds(globalQueue.getCurrentSong().duration),
		Title:   globalQueue.getCurrentSong().title,
		Artist:  []string{globalQueue.getCurrentSong().artist},
		Album:   globalQueue.getCurrentSong().album.title,
		ArtUrl: userCache + "/shanty/art/" +
			globalQueue.getCurrentSong().album.artworkId + ".jpg",
		//AlbumArtist:    []string{},
		//AsText:         "",
		//AudioBPM:       0,
		//AutoRating:     0.0,
		//Comment:        []string{},
		//Composer:       []string{},
		//ContentCreated: "",
		//DiscNumber:     0,
		//FirstUsed:      "",
		//Genre:          []string{},
		//LastUsed:       "",
		//Lyricist:       []string{},
		//TrackNumber:    0,
		//Url:            "",
		//UseCount:       0,
		//UserRating:     0.0,
	}, nil
}
func (p Player) Volume() (float64, error) {
	if globalMpv == nil {
		return 1.0, nil
	}

	var volume int64

	property, err := globalMpv.GetProperty("volume", mpv.FormatInt64)
	if err == nil {
		volume, _ = property.(int64)
	}

	return float64(volume) / 100.0, nil
}
func (p Player) SetVolume(in float64) error {
	log.Printf("%v", in)
	globalProgram.Send(msgCtrlChangeVolume{
		amount:     int64(in * 100),
		volumeType: "set",
	})
	return nil
}
func (p Player) Position() (int64, error) {
	if globalMpv == nil {
		return 0, nil
	}

	var position int64

	property, err := globalMpv.GetProperty("time-pos", mpv.FormatInt64)
	if err == nil {
		position, _ = property.(int64)
	}

	return int64(secondsToMicroseconds(float64(position))), nil
}
func (p Player) MinimumRate() (float64, error) {
	return 0, nil
}
func (p Player) MaximumRate() (float64, error) {
	return 0, nil
}

func (p Player) LoopStatus() (types.LoopStatus, error) {
	switch loopMode {
	case loopQueue:
		return types.LoopStatusPlaylist, nil
	case loopSong:
		return types.LoopStatusTrack, nil
	case loopNone:
		return types.LoopStatusNone, nil
	}
	return types.LoopStatusNone, nil
}

func (p Player) SetLoopStatus(status types.LoopStatus) error {
	globalProgram.Send(msgCtrlToggleLoopMode{})
	return nil
}

func (p Player) CanGoNext() (bool, error) {
	if len(globalQueue.songlist) == 0 {
		return false, nil
	}

	return true, nil
}
func (p Player) CanGoPrevious() (bool, error) {
	if len(globalQueue.songlist) == 0 {
		return false, nil
	}
	return true, nil
}
func (p Player) CanPlay() (bool, error) {
	if len(globalQueue.songlist) > 0 {
		return true, nil
	}
	return false, nil
}
func (p Player) CanPause() (bool, error) {
	if len(globalQueue.songlist) > 0 {
		return true, nil
	}
	return false, nil
}
func (p Player) CanSeek() (bool, error) {
	return true, nil
}
func (p Player) CanControl() (bool, error) {
	return true, nil
}

func initMpris() *server.Server {
	r := Root{}
	p := Player{}
	s := server.NewServer("shanty", r, p)

	go s.Listen()
	return s
}

func microsecondsToSeconds(m types.Microseconds) float64 {
	return float64(m) / 1_000_000
}

func secondsToMicroseconds(s float64) types.Microseconds {
	return types.Microseconds(s * 1_000_000)
}
