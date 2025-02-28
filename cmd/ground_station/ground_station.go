package main

import (
	"context"

	"github.com/harshabose/simple_webrtc_comm/datachannel/pkg"
	"github.com/pion/interceptor"
	"github.com/pion/webrtc/v4"

	"github.com/harshabose/simple_webrtc_comm/client/internal/config"
	"github.com/harshabose/simple_webrtc_comm/client/internal/constants"
	"github.com/harshabose/simple_webrtc_comm/client/pkg"
)

func main() {
	ctx := context.Background()
	mediaEngine := &webrtc.MediaEngine{}
	interceptorRegistry := &interceptor.Registry{}

	if err := webrtc.RegisterDefaultInterceptors(mediaEngine, interceptorRegistry); err != nil {
		panic(err)
	}

	// enable needed client options/capabilities
	groundstation, err := client.CreatePeerConnections(
		ctx, mediaEngine, interceptorRegistry,
		client.WithH264MediaEngine(constants.DefaultVideoClockRate, client.PacketisationMode1, client.ProfileLevelBaseline42),
		client.WithNACKInterceptor(client.NACKGeneratorLowLatency, client.NACKResponderLowLatency),
		client.WithFLEXFECInterceptor(),
		// client.WithJitterBufferInterceptor(),
		client.WithRTCPReportsInterceptor(client.RTCPReportIntervalLowLatency),
		client.WithTWCCSenderInterceptor(client.TWCCIntervalLowLatency),
	)
	if err != nil {
		panic(err)
	}

	// enable needed peer connection options/capabilities
	mainPeerConnection, err := groundstation.CreatePeerConnection(
		"MAIN",
		client.WithRTCConfiguration(config.GetRTCConfiguration()),
		client.WithAnswerSignal,
		client.WithMediaSinks(),
		client.WithDataChannels(),
	)
	if err != nil {
		panic(err)
	}

	if err := mainPeerConnection.CreateDataChannel("MAVLINK", data.WithRandomBindPort, data.WithLoopBackPort(14550)); err != nil {
		panic(err)
	}

	if err := mainPeerConnection.Connect("DELIVERY"); err != nil {
		panic(err)
	}

	groundstation.WaitUntilClosed()
}
