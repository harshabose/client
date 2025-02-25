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

func (pc *PeerConnections) CreatePeerConnection(label string, options ...PeerConnectionOption) error {
	var err error

	if _, exists := pc.peerConnections[label]; exists {
		return errors.New("peer connection already exists")
	}

	if pc.peerConnections[label], err = CreatePeerConnection(pc.ctx, pc.api, options...); err != nil {
		return err
	}

	return nil
}

func (pc *PeerConnections) GetPeerConnection(label string) (*PeerConnection, error) {
	if _, exists := pc.peerConnections[label]; !exists {
		return nil, errors.New("peer connection not found")
	}
	return pc.peerConnections[label], nil
}

func (pc *PeerConnections) CreateDataChannel(label string, peerConnection *PeerConnection, options ...data.LoopBackOption) error {
	return pc.dataChannels.CreateDataChannel(label, peerConnection.peerConnection, options...)
}

func (pc *PeerConnections) CreateMediaSource(peerConnection *PeerConnection, options ...mediasource.TrackOption) error {
	return pc.tracks.CreateTrack(peerConnection.peerConnection, options...)
}

// CreateMediaSink needs to have the label same as the remote track id (case-sensitive)
func (pc *PeerConnections) CreateMediaSink(label string, options ...mediasink.StreamOption) error {
	return pc.sinks.CreateSink(label, options...)
}

func (pc *PeerConnections) GetMediaSink(label string) (*mediasink.Sink, error) {
	return pc.sinks.GetSink(label)
}

func (pc *PeerConnections) Connect(peerConnection *PeerConnection, events ...Event) error {
	return nil
}

func (pc *PeerConnections) WaitUntilClosed() {
	<-pc.ctx.Done()
}
