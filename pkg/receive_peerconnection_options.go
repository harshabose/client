package client

import "github.com/harshabose/simple_webrtc_comm/client/internal/receiver"

var (
	WithRTCConfigurationReceiver = receiver.WithRTCConfiguration
	WithOfferSignalReceiver      = receiver.WithOfferSignal
	WithAnswerSignalReceiver     = receiver.WithAnswerSignal
	WithMediaSinksReceiver       = receiver.WithMediaSinks
	WithDataChannelsReceiver     = receiver.WithDataChannels
)
