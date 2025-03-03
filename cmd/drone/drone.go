package main

import (
	"context"

	"github.com/asticode/go-astiav"
	"github.com/harshabose/simple_webrtc_comm/mediasource/pkg"
	"github.com/harshabose/simple_webrtc_comm/transcode/pkg"
	"github.com/pion/interceptor"
	"github.com/pion/webrtc/v4"

	"github.com/harshabose/simple_webrtc_comm/datachannel/pkg"

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

	deliveryDrone, err := client.CreateClient(
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

	mainPeerConnection, err := deliveryDrone.CreatePeerConnection(
		"MAIN",
		client.WithRTCConfiguration(config.GetRTCConfiguration()),
		client.WithOfferSignal,
		client.WithMediaSources(),
		client.WithDataChannels(),
	)
	if err != nil {
		panic(err)
	}

	if err := mainPeerConnection.CreateDataChannel("MAVLINK",
		data.WithRandomBindPort,
		// data.WithMAVP2P(os.Getenv("MAVP2P_EXE_PATH"), os.Getenv("MAVLINK_SERIAL")),
	); err != nil {
		panic(err)
	}

	if err := mainPeerConnection.CreateMediaSource("A8-MINI",
		mediasource.WithH264Track(constants.DefaultVideoClockRate, mediasource.PacketisationMode1, mediasource.ProfileLevelBaseline42),
		mediasource.WithPriority(mediasource.Level5),
		mediasource.WithStream(
			mediasource.WithBufferSize(int(constants.DefaultVideoFPS*3)),
			mediasource.WithDemuxer(
				"/dev/video0",
				// "rtsp://192.168.144.25:8554/main.264",
				// transcode.WithRTSPInputOption,
				transcode.WithDemuxerBufferSize(int(constants.DefaultVideoFPS)*3),
			),
			mediasource.WithDecoder(transcode.WithDecoderBufferSize(int(constants.DefaultVideoFPS)*3)),
			mediasource.WithFilter(
				transcode.VideoFilters,
				transcode.WithFilterBufferSize(int(constants.DefaultVideoFPS)*3),
				transcode.WithVideoScaleFilterContent(constants.DefaultVideoWidth, constants.DefaultVideoHeight),
				transcode.WithVideoPixelFormatFilterContent(constants.DefaultPixelFormat),
				transcode.WithVideoFPSFilterContent(constants.DefaultVideoFPS),
			),
			mediasource.WithEncoder(
				astiav.CodecIDH264,
				transcode.WithEncoderBufferSize(int(constants.DefaultVideoFPS)*3),
				transcode.WithX264LowLatencyOptions,
			),
		),
	); err != nil {
		panic(err)
	}

	if err := mainPeerConnection.Connect("DELIVERY"); err != nil {
		panic(err)
	}

	deliveryDrone.WaitUntilClosed()
}
