package client

import (
	"context"
	"errors"
	"fmt"

	"github.com/pion/interceptor/pkg/cc"
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
	cancel         context.CancelFunc
}

func CreatePeerConnection(ctx context.Context, cancel context.CancelFunc, label string, api *webrtc.API, options ...PeerConnectionOption) (*PeerConnection, error) {
	var err error
	pc := &PeerConnection{
		label:  label,
		config: &webrtc.Configuration{},
		ctx:    ctx,
		cancel: cancel,
	}

	for _, option := range options {
		if err := option(pc); err != nil {
			return nil, err
		}
	}

	if pc.peerConnection, err = api.NewPeerConnection(*pc.config); err != nil {
		return nil, err
	}

	return pc.onConnectionStateChangeEvent().onTrack().onICEConnectionStateChange().onICEGatheringStateChange().onICECandidate(), err
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

func (pc *PeerConnection) GetBWEstimator() (cc.BandwidthEstimator, error) {
	if pc.bwController == nil || pc.bwController.estimator == nil {
		return nil, errors.New("estimator is nil")
	}

	return pc.bwController.estimator, nil
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

		// RTSP HOST ASSUMES A RTSP SERVER IS RUNNING IN THE GIVEN CONFIG
		config := rtsp.LocalHostConfig()
		config.StreamPath = remote.ID()

		sink, err := pc.sinks.CreateSink(remote.ID(), mediasink.WithRTSPHost(config, nil, rtsp.WithOptionsFromRemote(remote)))
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

func (pc *PeerConnection) CreateDataChannel(label string, sink *mediasink.Sink) (*data.DataChannel, error) {
	if pc.dataChannels == nil {
		return nil, errors.New("data channels are not enabled")
	}
	return pc.dataChannels.CreateDataChannel(label, pc.peerConnection, sink)
}

func (pc *PeerConnection) CreateMediaSource(label string, withBWController bool, options ...mediasource.TrackOption) (*mediasource.Track, error) {
	if pc.tracks == nil {
		return nil, errors.New("media source are not enabled")
	}

	track, err := pc.tracks.CreateTrack(label, pc.peerConnection, options...)
	if err != nil {
		return nil, err
	}

	if pc.bwController != nil && pc.bwController.estimator != nil && withBWController {
		fmt.Printf("subscribing media source with label '%s' to bw estimator\n", label)
		return nil, pc.bwController.Subscribe(track)
	}

	return track, nil
}

func (pc *PeerConnection) CreateMediaSink(label string, options ...mediasink.StreamOption) (*mediasink.Sink, error) {
	if pc.sinks == nil {
		return nil, errors.New("media sink are not enabled")
	}
	sink, err := pc.sinks.CreateSink(label, options...)
	if err != nil {
		return nil, err
	}
	return sink, nil
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

func (pc *PeerConnection) Close() error {
	// clear data channels if any
	// clear tracks if any
	// clear sinks if any
	// clear bwController ??
	return pc.peerConnection.Close()
}
