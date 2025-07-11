package client

import (
	"github.com/harshabose/simple_webrtc_comm/client/pkg/datachannel"
	"github.com/harshabose/simple_webrtc_comm/client/pkg/mediasink"
	"github.com/harshabose/simple_webrtc_comm/client/pkg/mediasource"
)

type PeerConnectionOption = func(*PeerConnection) error

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

func WithMediaSources() PeerConnectionOption {
	return func(pc *PeerConnection) error {
		pc.tracks = mediasource.CreateTracks(pc.ctx)

		return nil
	}
}

func WithMediaSinks() PeerConnectionOption {
	return func(pc *PeerConnection) error {
		pc.sinks = mediasink.CreateSinks(pc.ctx, pc.peerConnection)

		return nil
	}
}

func WithDataChannels() PeerConnectionOption {
	return func(pc *PeerConnection) error {
		pc.dataChannels = datachannel.CreateDataChannels(pc.ctx)

		return nil
	}
}
