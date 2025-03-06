package client

import (
	"context"
	"errors"
	"fmt"
	
	"github.com/pion/webrtc/v4"
	
	"github.com/harshabose/simple_webrtc_comm/mediasink/pkg"
	"github.com/harshabose/simple_webrtc_comm/mediasink/pkg/rtsp"
	"github.com/harshabose/simple_webrtc_comm/mediasource/pkg"
	
	"github.com/harshabose/simple_webrtc_comm/datachannel/pkg"
)

type PeerConnection struct {
	label          string
	peerConnection *webrtc.PeerConnection
	dataChannels   *data.DataChannels
	tracks         *mediasource.Tracks
	sinks          *mediasink.Sinks
	config         *webrtc.Configuration
	signal         BaseSignal
	bwController   *bwController
	ctx            context.Context
}

func CreatePeerConnection(ctx context.Context, label string, api *webrtc.API, options ...PeerConnectionOption) (*PeerConnection, error) {
	var err error
	pc := &PeerConnection{
		label:  label,
		config: &webrtc.Configuration{},
		ctx:    ctx,
	}
	
	if pc.peerConnection, err = api.NewPeerConnection(*pc.config); err != nil {
		return nil, err
	}
	
	for _, option := range options {
		if err := option(pc); err != nil {
			return nil, err
		}
	}
	
	return pc.onConnectionStateChangeEvent().onDataChannel(), err
}

func (pc *PeerConnection) GetLabel() string {
	return pc.label
}

func (pc *PeerConnection) onTrackEvent() *PeerConnection {
	pc.peerConnection.OnTrack(func(remote *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
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
	
	return pc
}

func (pc *PeerConnection) onConnectionStateChangeEvent() *PeerConnection {
	pc.peerConnection.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		fmt.Printf("peer connection state with label changed to %s\n", state.String())
	})
	return pc
}

func (pc *PeerConnection) onDataChannel() *PeerConnection {
	pc.peerConnection.OnDataChannel(func(channel *webrtc.DataChannel) {
		fmt.Println("got a non pre-negotiated datachannel with label:", channel.Label())
		fmt.Println("creating a ad-hoc datachannel interface...")
		dataChannel, err := pc.dataChannels.CreateDataChannel(channel.Label(), pc.peerConnection, data.WithRandomBindPort)
		if err != nil {
			fmt.Println(errors.New("failed to create datachannel"))
		}
		fmt.Println("successfully created a raw datachannel with bind port:", dataChannel.GetBindPort())
		fmt.Println("send and receive to and from the above bind port")
	})
	return pc
}

func (pc *PeerConnection) CreateDataChannel(label string, options ...data.LoopBackOption) (*data.DataChannel, error) {
	if pc.dataChannels == nil {
		return nil, errors.New("data channels are not enabled")
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
		fmt.Printf("subscribing media source with label '%s' to bw estimator\n", label)
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
	if err := pc.signal.Connect(category, pc.label); err != nil {
		return err
	}
	
	if pc.tracks != nil {
		fmt.Println("Starting all media source tracks...")
		pc.tracks.StartAll()
	}
	if pc.bwController != nil {
		pc.bwController.Start()
	}
	
	return nil
}
