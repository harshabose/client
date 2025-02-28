package client

import (
	"github.com/harshabose/simple_webrtc_comm/datachannel/pkg"
	"github.com/harshabose/simple_webrtc_comm/mediasink/pkg"
	"github.com/harshabose/simple_webrtc_comm/mediasource/pkg"
	"github.com/pion/webrtc/v4"

	"github.com/harshabose/simple_webrtc_comm/client/internal/signal"
)

type PeerConnectionOption = func(*PeerConnection) error

func WithRTCConfiguration(config webrtc.Configuration) PeerConnectionOption {
	return func(connection *PeerConnection) error {
		connection.config = config
		return nil
	}
}

func WithOfferSignal(connection *PeerConnection) error {
	connection.signal = signal.CreateOfferSignal(connection.ctx, connection.peerConnection)
	return nil
}

func WithAnswerSignal(connection *PeerConnection) error {
	connection.signal = signal.CreateAnswerSignal(connection.ctx, connection.peerConnection)
	return nil
}

func WithMediaSources(options ...mediasource.TracksOption) PeerConnectionOption {
	return func(pc *PeerConnection) error {
		var err error

		if pc.tracks, err = mediasource.CreateTracks(pc.ctx, options...); err != nil {
			return err
		}

		return nil
	}
}

func WithMediaSinks(options ...mediasink.SinksOptions) PeerConnectionOption {
	return func(pc *PeerConnection) error {
		var err error

		if pc.sinks, err = mediasink.CreateSinks(pc.ctx, options...); err != nil {
			return err
		}
		// pc.onTrackEvent()

		return nil
	}
}

func WithDataChannels() PeerConnectionOption {
	return func(pc *PeerConnection) error {
		var err error
		if pc.dataChannels, err = data.CreateDataChannels(pc.ctx); err != nil {
			return err
		}

		return nil
	}
}
