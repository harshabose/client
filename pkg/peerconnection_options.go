package client

import (
	"github.com/pion/webrtc/v4"

	"github.com/harshabose/simple_webrtc_comm/datachannel/pkg"
	"github.com/harshabose/simple_webrtc_comm/mediasink/pkg"
	"github.com/harshabose/simple_webrtc_comm/mediasource/pkg"
)

type PeerConnectionOption = func(*PeerConnection) error

func WithRTCConfiguration(config *webrtc.Configuration) PeerConnectionOption {
	return func(connection *PeerConnection) error {
		connection.config = config
		return nil
	}
}

func WithFirebaseOfferSignal(connection *PeerConnection) error {
	connection.signal = CreateFirebaseOfferSignal(connection.ctx, connection)
	return nil
}

func WithFirebaseAnswerSignal(connection *PeerConnection) error {
	connection.signal = CreateFirebaseAnswerSignal(connection.ctx, connection)
	return nil
}

func WithFileOfferSignal(offerPath, answerPath string) PeerConnectionOption {
	return func(connection *PeerConnection) error {
		connection.signal = CreateFileOfferSignal(connection.ctx, connection, offerPath, answerPath)
		return nil
	}
}

func WithFileAnswerSignal(offerPath, answerPath string) PeerConnectionOption {
	return func(connection *PeerConnection) error {
		connection.signal = CreateFileAnswerSignal(connection.ctx, connection, offerPath, answerPath)
		return nil
	}
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

func WithBandwidthControl() PeerConnectionOption {
	return func(connection *PeerConnection) error {
		connection.bwController = createBWController(connection.ctx)
		return nil
	}
}
