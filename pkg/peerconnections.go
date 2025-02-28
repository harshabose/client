package client

import (
	"context"
	"errors"

	"github.com/pion/interceptor"
	"github.com/pion/webrtc/v4"
)

type PeerConnections struct {
	peerConnections     map[string]*PeerConnection
	mediaEngine         *webrtc.MediaEngine
	interceptorRegistry *interceptor.Registry
	api                 *webrtc.API
	ctx                 context.Context
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

func (pc *PeerConnections) WaitUntilClosed() {
	<-pc.ctx.Done()
}
