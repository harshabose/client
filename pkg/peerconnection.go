package client

import (
	"context"

	"github.com/pion/webrtc/v4"

	"github.com/harshabose/simple_webrtc_comm/client/internal/signal"
)

type PeerConnection struct {
	peerConnection *webrtc.PeerConnection
	config         webrtc.Configuration
	signal         signal.BaseSignal
	ctx            context.Context
}

func CreatePeerConnection(ctx context.Context, api *webrtc.API, options ...PeerConnectionOption) (*PeerConnection, error) {
	var err error
	pc := &PeerConnection{ctx: ctx}

	if pc.peerConnection, err = api.NewPeerConnection(pc.config); err != nil {
		return nil, err
	}

	for _, option := range options {
		if err := option(pc); err != nil {
			return nil, err
		}
	}

	return pc, err
}
