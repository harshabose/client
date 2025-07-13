package mediasource

import (
	"context"
	"errors"
	"fmt"

	"github.com/pion/webrtc/v4"
)

type Tracks struct {
	tracks map[string]*Track
	ctx    context.Context
}

func CreateTracks(ctx context.Context) *Tracks {
	return &Tracks{
		tracks: make(map[string]*Track),
		ctx:    ctx,
	}
}

func (tracks *Tracks) CreateTrack(label string, peerConnection *webrtc.PeerConnection, options ...TrackOption) (*Track, error) {
	if _, exists := tracks.tracks[label]; exists {
		return nil, fmt.Errorf("track with id = '%s' already exists", label)
	}

	track, err := CreateTrack(tracks.ctx, label, peerConnection, options...)
	if err != nil {
		return nil, err
	}

	tracks.tracks[label] = track
	return track, nil
}

func (tracks *Tracks) GetTrack(id string) (*Track, error) {
	track, exists := tracks.tracks[id]
	if !exists {
		return nil, errors.New("track does not exits")
	}

	return track, nil
}
