package client

import (
	"context"
	"errors"

	"github.com/harshabose/simple_webrtc_comm/datachannel/pkg"
	"github.com/harshabose/simple_webrtc_comm/mediasink/pkg"
	"github.com/harshabose/simple_webrtc_comm/mediasource/pkg"
	"github.com/pion/interceptor"
	"github.com/pion/webrtc/v4"
)

type PeerConnections struct {
	peerConnections     map[string]*PeerConnection
	mediaEngine         *webrtc.MediaEngine
	interceptorRegistry *interceptor.Registry
	api                 *webrtc.API
	ctx                 context.Context

	dataChannels *data.DataChannels
	tracks       *mediasource.Tracks
	sinks        *mediasink.Sinks
}

func CreatePeerConnections(ctx context.Context, mediaEngine *webrtc.MediaEngine, interceptorRegistry *interceptor.Registry, options ...PeerConnectionsOption) (*PeerConnections, error) {
	peerConnections := &PeerConnections{
		mediaEngine:         mediaEngine,
		interceptorRegistry: interceptorRegistry,
		peerConnections:     make(map[string]*PeerConnection),
		ctx:                 ctx,
	}

	for _, option := range options {
		if err := option(peerConnections); err != nil {
			return nil, err
		}
	}

	peerConnections.api = webrtc.NewAPI(webrtc.WithMediaEngine(peerConnections.mediaEngine), webrtc.WithInterceptorRegistry(peerConnections.interceptorRegistry))

	return peerConnections, nil
}

func (pc *PeerConnections) CreatePeerConnection(label string, options ...PeerConnectionOption) (*PeerConnection, error) {
	var err error

	if _, exists := pc.peerConnections[label]; exists {
		return nil, errors.New("peer connection already exists")
	}

	if pc.peerConnections[label], err = CreatePeerConnection(pc.ctx, pc.api, options...); err != nil {
		return nil, err
	}

	return pc.peerConnections[label], nil
}

func (pc *PeerConnections) GetPeerConnection(label string) (*PeerConnection, error) {
	if _, exists := pc.peerConnections[label]; !exists {
		return nil, errors.New("peer connection not found")
	}
	return pc.peerConnections[label], nil
}

func (pc *PeerConnections) CreateDataChannel(label string, peerConnectionLabel string, options ...data.LoopBackOption) error {
	peerConnection, err := pc.GetPeerConnection(peerConnectionLabel)
	if err != nil {
		return err
	}
	return pc.dataChannels.CreateDataChannel(label, peerConnection.peerConnection, options...)
}

func (pc *PeerConnections) CreateMediaSource(peerConnectionLabel string, options ...mediasource.TrackOption) error {
	peerConnection, err := pc.GetPeerConnection(peerConnectionLabel)
	if err != nil {
		return err
	}
	track, err := pc.tracks.CreateTrack(peerConnection.peerConnection, options...)
	if err != nil {
		return err
	}
	if _, exists := peerConnection.allocatedTracks[track.GetTrack().ID()]; exists {
		return errors.New("track with same name allocated to the peer connection")
	}
	peerConnection.allocatedTracks[track.GetTrack().ID()] = track

	return nil
}

// CreateMediaSink needs to have the label same as the remote track id (case-sensitive)
func (pc *PeerConnections) CreateMediaSink(label string, peerConnectionLabel string, options ...mediasink.StreamOption) error {
	peerConnection, err := pc.GetPeerConnection(peerConnectionLabel)
	if err != nil {
		return err
	}

	sink, err := pc.sinks.CreateSink(label, options...)
	if err != nil {
		return err
	}

	if _, exists := peerConnection.allocatedSinks[label]; exists {
		return errors.New("sink with same name allocated to the peer connection")
	}
	peerConnection.allocatedSinks[label] = sink

	return nil
}

func (pc *PeerConnections) GetMediaSource(label string) (*mediasource.Track, error) {
	return pc.tracks.GetTrack(label)
}

func (pc *PeerConnections) GetMediaSink(label string) (*mediasink.Sink, error) {
	return pc.sinks.GetSink(label)
}

func (pc *PeerConnections) Connect(category, peerConnectionLabel string) error {
	peerConnection, err := pc.GetPeerConnection(peerConnectionLabel)
	if err != nil {
		return err
	}

	if err := peerConnection.signal.Connect(category, peerConnectionLabel); err != nil {
		return err
	}
	if pc.tracks != nil {
		pc.tracks.StartAll()
	}

	return nil
}

func (pc *PeerConnections) WaitUntilClosed() {
	<-pc.ctx.Done()
}
