package mediasource

import (
	"context"
	"errors"
	"fmt"
	"iter"

	"github.com/pion/webrtc/v4"
)

type Tracks struct {
	tracks  map[string]*Track
	tracks2 map[string]*RTPTrack
	ctx     context.Context
}

func CreateTracks(ctx context.Context) *Tracks {
	return &Tracks{
		tracks:  make(map[string]*Track),
		tracks2: make(map[string]*RTPTrack),
		ctx:     ctx,
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

func (tracks *Tracks) CreateRTPTrack(label string, peerConnection *webrtc.PeerConnection, options ...TrackOption) (*RTPTrack, error) {
	if _, exists := tracks.tracks2[label]; exists {
		return nil, fmt.Errorf("track with id = '%s' already exists", label)
	}

	track, err := CreateRTPTrack(tracks.ctx, label, peerConnection, options...)
	if err != nil {
		return nil, err
	}

	tracks.tracks2[label] = track
	return track, nil
}

func (tracks *Tracks) GetTrack(id string) (*Track, error) {
	track, exists := tracks.tracks[id]
	if !exists {
		return nil, errors.New("track does not exits")
	}

	return track, nil
}

func (tracks *Tracks) GetRTPTrack(id string) (*RTPTrack, error) {
	track, exists := tracks.tracks2[id]
	if !exists {
		return nil, errors.New("track does not exits")
	}

	return track, nil
}

func (tracks *Tracks) Tracks() iter.Seq2[string, *Track] {
	return func(yield func(string, *Track) bool) {
		for id, track := range tracks.tracks {
			if !yield(id, track) {
				return
			}
		}
	}
}

func (tracks *Tracks) RTPTracks() iter.Seq2[string, *RTPTrack] {
	return func(yield func(string, *RTPTrack) bool) {
		for id, track := range tracks.tracks2 {
			if !yield(id, track) {
				return
			}
		}
	}
}
