package receiver

import (
	data "github.com/harshabose/simple_webrtc_comm/datachannel/pkg"
	mediasink "github.com/harshabose/simple_webrtc_comm/mediasink/pkg"
	"github.com/pion/webrtc/v4"

	"github.com/harshabose/simple_webrtc_comm/client/internal"
)

type PeerConnectionOption = func(*PeerConnection) error

func WithRTCConfiguration(config *webrtc.Configuration) PeerConnectionOption {
	return func(connection *PeerConnection) error {
		connection.config = config
		return nil
	}
}

func WithOfferSignal(connection *PeerConnection) error {
	connection.signal = internal.CreateOfferSignal(connection.ctx, connection)
	return nil
}

func WithAnswerSignal(connection *PeerConnection) error {
	connection.signal = internal.CreateAnswerSignal(connection.ctx, connection)
	return nil
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
