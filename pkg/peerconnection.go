package client

import (
	"context"
	"errors"
	"fmt"

	"github.com/harshabose/simple_webrtc_comm/mediasink/pkg"
	"github.com/harshabose/simple_webrtc_comm/mediasink/pkg/rtsp"
	"github.com/harshabose/simple_webrtc_comm/mediasource/pkg"
	"github.com/pion/webrtc/v4"

	"github.com/harshabose/simple_webrtc_comm/datachannel/pkg"
)

type PeerConnection struct {
	peerConnection *webrtc.PeerConnection
	dataChannels   *data.DataChannels
	tracks         *mediasource.Tracks
	sinks          *mediasink.Sinks
	config         webrtc.Configuration
	signal         BaseSignal
	bwController   *bwController
	ctx            context.Context
}

func CreatePeerConnection(ctx context.Context, api *webrtc.API, options ...PeerConnectionOption) (*PeerConnection, error) {
	var err error
	pc := &PeerConnection{
		config: webrtc.Configuration{},
		ctx:    ctx,
	}

	if pc.peerConnection, err = api.NewPeerConnection(pc.config); err != nil {
		return nil, err
	}

	pc.onConnectionStateChangeEvent()

	for _, option := range options {
		if err := option(pc); err != nil {
			return nil, err
		}
	}

	return pc, err
}

func (pc *PeerConnection) onTrackEvent() {
	fmt.Println("setting up on track event")
	pc.peerConnection.OnTrack(func(remote *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
		fmt.Println("got a track with ID: ", remote.ID())

		if pc.sinks == nil {
			fmt.Println("got a remote track but sinks are not enabled...")
		}
		if sink, err := pc.sinks.GetSink(remote.ID()); err == nil {
			fmt.Println("found existing sink for track ID:", remote.ID())
			sink.SetTrack(remote)
			sink.Start()
			return
		}

		fmt.Println("no sink pre-set for ID:", remote.ID())
		fmt.Println("creating a temporary sink...")

		sink, err := pc.sinks.CreateSink(remote.ID(), mediasink.WithRTSPHost(8554, remote.ID(), rtsp.WithH264OptionsFromRemote(remote)))
		if err != nil {
			fmt.Println("failed to create sink:", err)
		} else {
			sink.SetTrack(remote)
			sink.Start()
			fmt.Println("temporary sink created and started for track ID:", remote.ID())
		}
	})
}

func (pc *PeerConnection) onConnectionStateChangeEvent() {
	pc.peerConnection.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		fmt.Printf("peer connection state with label changed to %s\n", state.String())
	})
}

func (pc *PeerConnection) CreateDataChannel(label string, options ...data.LoopBackOption) error {
	if pc.dataChannels == nil {
		return errors.New("data channels are not enabled")
	}
	return pc.dataChannels.CreateDataChannel(label, pc.peerConnection, options...)
}

func (pc *PeerConnection) CreateMediaSource(label string, withBWController bool, options ...mediasource.TrackOption) error {
	if pc.tracks == nil {
		return errors.New("media source are not enabled")
	}

	track, err := pc.tracks.CreateTrack(label, pc.peerConnection, options...)
	if err != nil {
		return err
	}

	if pc.bwController.estimator != nil && withBWController {
		return pc.bwController.Subscribe(track)
	}

	return nil
}

func (pc *PeerConnection) CreateMediaSink(label string, options ...mediasink.StreamOption) error {
	if pc.sinks == nil {
		return errors.New("media sink are not enabled")
	}
	if _, err := pc.sinks.CreateSink(label, options...); err != nil {
		return err
	}
	return nil
}

func (pc *PeerConnection) Connect(category string) error {
	if err := pc.signal.Connect(category, "MAIN"); err != nil {
		return err
	}

	if pc.tracks != nil {
		fmt.Println("Starting all media source tracks...")
		pc.tracks.StartAll()
	}

	pc.bwController.Start()

	return nil
}
