package client

import (
	"github.com/harshabose/simple_webrtc_comm/datachannel/pkg"
	"github.com/harshabose/simple_webrtc_comm/mediasink/pkg"
	"github.com/harshabose/simple_webrtc_comm/mediasource/pkg"
)

type PeerConnectionsOption = func(*PeerConnections) error

func WithMediaSources(options ...mediasource.TracksOption) PeerConnectionsOption {
	return func(pc *PeerConnections) error {
		var err error

		if pc.tracks, err = mediasource.CreateTracks(pc.ctx, pc.mediaEngine, pc.interceptorRegistry, options...); err != nil {
			return err
		}

		return nil
	}
}

func WithMediaSinks(options ...mediasink.SinksOptions) PeerConnectionsOption {
	return func(pc *PeerConnections) error {
		var err error

		if pc.sinks, err = mediasink.CreateSinks(pc.ctx, pc.mediaEngine, pc.interceptorRegistry, options...); err != nil {
			return err
		}

		return nil
	}
}

func WithDataChannels() PeerConnectionsOption {
	return func(pc *PeerConnections) error {
		var err error
		if pc.dataChannels, err = data.CreateDataChannels(pc.ctx); err != nil {
			return err
		}

		return nil
	}
}
