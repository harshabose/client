package main

import (
	"context"

	"github.com/harshabose/simple_webrtc_comm/datachannel/pkg"
	"github.com/harshabose/simple_webrtc_comm/mediasink/pkg"
	"github.com/harshabose/simple_webrtc_comm/mediasink/pkg/rtsp"
	"github.com/pion/interceptor"
	"github.com/pion/webrtc/v4"

	"github.com/harshabose/simple_webrtc_comm/client/internal/config"
	"github.com/harshabose/simple_webrtc_comm/client/pkg"
)

func main() {
	ctx := context.Background()
	mediaEngine := &webrtc.MediaEngine{}
	interceptorRegistry := &interceptor.Registry{}

	groundstation, err := client.CreatePeerConnections(
		ctx, mediaEngine, interceptorRegistry,
		client.WithDataChannels(),
		client.WithMediaSinks(),
	)
	if err != nil {
		panic(err)
	}

	if err := groundstation.CreatePeerConnection(
		"MAIN",
		client.WithRTCConfiguration(config.GetRTCConfiguration()),
		client.WithOfferSignal,
	); err != nil {
		panic(err)
	}

	peerConnection, err := groundstation.GetPeerConnection("MAIN")
	if err != nil {
		panic(err)
	}

	if err := groundstation.CreateDataChannel("MAVLINK", peerConnection, data.WithBindPort, data.WithLoopBackPort); err != nil {
		panic(err)
	}

	host, err := rtsp.CreateHost(8554, "A8-MINI", rtsp.WithH264Options(rtsp.PacketisationMode1, nil, nil))
	if err != nil {
		panic(err)
	}

	if err := groundstation.CreateMediaSink("A8-MINI", mediasink.WithHost(host)); err != nil {
		panic(err)
	}

	if err := groundstation.Connect(peerConnection); err != nil {
		panic(err)
	}
	groundstation.WaitUntilClosed()
}
