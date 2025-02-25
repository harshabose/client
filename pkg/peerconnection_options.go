package client

import (
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
