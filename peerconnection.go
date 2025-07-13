package client

import (
	"context"
	"errors"
	"fmt"

	"github.com/pion/webrtc/v4"

	"github.com/harshabose/simple_webrtc_comm/client/pkg/datachannel"
	"github.com/harshabose/simple_webrtc_comm/client/pkg/mediasink"
	"github.com/harshabose/simple_webrtc_comm/client/pkg/mediasource"
)

type PeerConnection struct {
	label          string
	peerConnection *webrtc.PeerConnection
	dataChannels   *datachannel.DataChannels
	tracks         *mediasource.Tracks
	sinks          *mediasink.Sinks
	signal         BaseSignal
	bwController   *BWEController
	ctx            context.Context
	cancel         context.CancelFunc
}

func CreatePeerConnection(ctx context.Context, cancel context.CancelFunc, label string, api *webrtc.API, config webrtc.Configuration, options ...PeerConnectionOption) (*PeerConnection, error) {
	pc := &PeerConnection{
		label:  label,
		ctx:    ctx,
		cancel: cancel,
	}

	peerConnection, err := api.NewPeerConnection(config)
	if err != nil {
		return nil, err
	}
	pc.peerConnection = peerConnection

	for _, option := range options {
		if err := option(pc); err != nil {
			return nil, err
		}
	}

	if pc.signal == nil {
		return nil, errors.New("signaling protocol not provided")
	}

	return pc.onConnectionStateChangeEvent().onICEConnectionStateChange().onICEGatheringStateChange().onICECandidate(), err
}

func (pc *PeerConnection) GetLabel() string {
	return pc.label
}

func (pc *PeerConnection) GetPeerConnection() (*webrtc.PeerConnection, error) {
	if pc.peerConnection == nil {
		return nil, errors.New("raw peer connection is nil")
	}

	return pc.peerConnection, nil
}

func (pc *PeerConnection) onConnectionStateChangeEvent() *PeerConnection {
	pc.peerConnection.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		fmt.Printf("peer connection state with label changed to %s\n", state.String())
		if state == webrtc.PeerConnectionStateDisconnected || state == webrtc.PeerConnectionStateClosed || state == webrtc.PeerConnectionStateFailed {
			fmt.Println("tying to cancel context for restart")
			pc.cancel()
		}
	})
	return pc
}

func (pc *PeerConnection) onICEConnectionStateChange() *PeerConnection {
	pc.peerConnection.OnICEConnectionStateChange(func(state webrtc.ICEConnectionState) {
		fmt.Printf("ICE Connection State changed: %s\n", state.String())
	})
	return pc
}

func (pc *PeerConnection) onICEGatheringStateChange() *PeerConnection {
	pc.peerConnection.OnICEGatheringStateChange(func(state webrtc.ICEGatheringState) {
		fmt.Printf("ICE Gathering State changed: %s\n", state.String())
	})
	return pc
}

func (pc *PeerConnection) onICECandidate() *PeerConnection {
	pc.peerConnection.OnICECandidate(func(candidate *webrtc.ICECandidate) {
		if candidate == nil {
			fmt.Println("ICE gathering complete")
			return
		}

		fmt.Printf("Found candidate: %s (type: %s)\n", candidate.String(), candidate.Typ)
	})
	return pc
}

func (pc *PeerConnection) CreateDataChannel(label string, options ...datachannel.Option) (*datachannel.DataChannel, error) {
	if pc.dataChannels == nil {
		return nil, errors.New("data channels are not enabled")
	}
	channel, err := pc.dataChannels.CreateDataChannel(label, pc.peerConnection, options...)
	if err != nil {
		return nil, err
	}

	return channel, nil
}

func (pc *PeerConnection) CreateMediaSource(label string, options ...mediasource.TrackOption) (*mediasource.Track, error) {
	if pc.tracks == nil {
		return nil, errors.New("media source are not enabled")
	}

	track, err := pc.tracks.CreateTrack(label, pc.peerConnection, options...)
	if err != nil {
		return nil, err
	}

	return track, nil
}

func (pc *PeerConnection) CreateMediaSink(label string, options ...mediasink.SinkOption) (*mediasink.Sink, error) {
	if pc.sinks == nil {
		return nil, errors.New("media sinks are not enabled")
	}

	sink, err := pc.sinks.CreateSink(label, options...)
	if err != nil {
		return nil, err
	}

	return sink, nil
}

func (pc *PeerConnection) Connect(category string) error {
	if pc.signal == nil {
		return errors.New("no signaling protocol provided")
	}
	if err := pc.signal.Connect(category, pc.label); err != nil {
		return err
	}

	return nil
}

func (pc *PeerConnection) Close() error {
	// TODO:
	// clear data channels if any
	// clear tracks if any
	// clear sinks if any
	// clear bwController ??
	return pc.peerConnection.Close()
}
