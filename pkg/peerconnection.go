package client

import (
	"context"
	"errors"
	"fmt"

	"github.com/harshabose/simple_webrtc_comm/mediasink/pkg"
	"github.com/harshabose/simple_webrtc_comm/mediasource/pkg"
	"github.com/pion/webrtc/v4"

	"github.com/harshabose/simple_webrtc_comm/datachannel/pkg"

	"github.com/harshabose/simple_webrtc_comm/client/internal/signal"
)

type PeerConnection struct {
	peerConnection        *webrtc.PeerConnection
	allocatedTracks       map[string]*mediasource.Track
	allocatedSinks        map[string]*mediasink.Sink
	allocatedDataChannels map[string]*data.DataChannel
	config                webrtc.Configuration
	signal                signal.BaseSignal
	ctx                   context.Context
}

func CreatePeerConnection(ctx context.Context, api *webrtc.API, options ...PeerConnectionOption) (*PeerConnection, error) {
	var err error
	pc := &PeerConnection{
		ctx:                   ctx,
		allocatedTracks:       make(map[string]*mediasource.Track),
		allocatedSinks:        make(map[string]*mediasink.Sink),
		allocatedDataChannels: make(map[string]*data.DataChannel),
	}

	if pc.peerConnection, err = api.NewPeerConnection(pc.config); err != nil {
		return nil, err
	}

	for _, option := range options {
		if err := option(pc); err != nil {
			return nil, err
		}
	}

	pc.onTrackEvent()
	pc.onConnectionStateChangeEvent()

	return pc, err
}

func (pc *PeerConnection) onTrackEvent() {
	pc.peerConnection.OnTrack(func(remote *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
		if _, exists := pc.allocatedSinks[remote.ID()]; !exists {
			fmt.Println(errors.New("no allocated sinks for remote track"), remote.ID())
			return
		}

		sink := pc.allocatedSinks[remote.ID()]
		sink.SetTrack(remote)
		sink.Start()
	})
}

func (pc *PeerConnection) onConnectionStateChangeEvent() {
	pc.peerConnection.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		fmt.Printf("peer connection state with label  changed to %s\n", state.String())
	})
}
