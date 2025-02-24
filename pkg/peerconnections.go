package client

import (
	"context"
	"errors"
	"github.com/harshabose/simple_webrtc_comm/datachannel/pkg"
	"github.com/harshabose/simple_webrtc_comm/mediasource/pkg"
	"github.com/pion/interceptor"
	"github.com/pion/webrtc/v4"
)

type PeerConnections struct {
	peerConnections     map[string]*webrtc.PeerConnection
	mediaEngine         *webrtc.MediaEngine
	interceptorRegistry *interceptor.Registry
	api                 *webrtc.API
	dataChannels        *data.DataChannels
	tracks              *mediasource.Tracks
	ctx                 context.Context
}

func CreatePeerConnections(ctx context.Context, mediaEngine *webrtc.MediaEngine, interceptorRegistry *interceptor.Registry, options ...PeerConnectionOption) (*PeerConnections, error) {
	peerConnection := &PeerConnections{
		mediaEngine:         mediaEngine,
		interceptorRegistry: interceptorRegistry,
		peerConnections:     make(map[string]*webrtc.PeerConnection),
		ctx:                 ctx,
	}

	for _, option := range options {
		if err := option(peerConnection); err != nil {
			return nil, err
		}
	}

	peerConnection.api = webrtc.NewAPI(webrtc.WithMediaEngine(peerConnection.mediaEngine), webrtc.WithInterceptorRegistry(peerConnection.interceptorRegistry))

	return peerConnection, nil
}

func (pc *PeerConnections) CreatePeerConnection(label string, config webrtc.Configuration) error {
	var err error

	if _, exists := pc.peerConnections[label]; exists {
		return errors.New("peer connection already exists")
	}

	if pc.peerConnections[label], err = pc.api.NewPeerConnection(config); err != nil {
		return err
	}
	return nil
}

func (pc *PeerConnections) GetPeerConnection(label string) (*webrtc.PeerConnection, error) {
	if _, exists := pc.peerConnections[label]; !exists {
		return nil, errors.New("peer connection not found")
	}
	return pc.peerConnections[label], nil
}
