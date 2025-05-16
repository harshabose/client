package receiver

import (
	"context"
	"errors"
	"fmt"

	data "github.com/harshabose/simple_webrtc_comm/datachannel/pkg"
	mediasink "github.com/harshabose/simple_webrtc_comm/mediasink/pkg"
	"github.com/harshabose/simple_webrtc_comm/mediasink/pkg/rtsp"
	"github.com/pion/webrtc/v4"

	"github.com/harshabose/simple_webrtc_comm/client/internal"
)

type PeerConnection struct {
	label          string
	peerConnection *webrtc.PeerConnection
	dataChannels   *data.DataChannels
	sinks          *mediasink.Sinks
	config         *webrtc.Configuration
	signal         internal.BaseSignal
	ctx            context.Context
	cancel         context.CancelFunc
}

func CreatePeerConnection(ctx context.Context, label string, api *webrtc.API, options ...PeerConnectionOption) (*PeerConnection, error) {
	ctx2, cancel := context.WithCancel(ctx)
	pc := &PeerConnection{
		label:  label,
		config: &webrtc.Configuration{},
		cancel: cancel,
		ctx:    ctx2,
	}

	for _, option := range options {
		if err := option(pc); err != nil {
			return nil, err
		}
	}

	return pc.Setup(api)
}

func (pc *PeerConnection) GetPeerConnection() *webrtc.PeerConnection {
	return pc.peerConnection
}

func (pc *PeerConnection) Setup(api *webrtc.API) (*PeerConnection, error) {
	peerConnection, err := api.NewPeerConnection(*pc.config)
	if err != nil {
		return nil, err
	}

	pc.peerConnection = peerConnection
	return pc.onConnectionStateChangeEvent().onDataChannel().onTrack().onICEConnectionStateChange().onICEGatheringStateChange().onICECandidate(), nil
}

func (pc *PeerConnection) GetLabel() string {
	return pc.label
}

func (pc *PeerConnection) onTrack() *PeerConnection {
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

		sink, err := pc.sinks.CreateSink(remote.ID(), mediasink.WithRTSPHost(8554, remote.ID(), rtsp.WithOptionsFromRemote(remote)))
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

	return nil
}

func (pc *PeerConnection) Close() error {
	// clear data channels if any
	// clear tracks if any
	// clear sinks if any
	// TODO: ADD THIS ASAP
	// clear bwController ??
	return pc.peerConnection.Close()
}
