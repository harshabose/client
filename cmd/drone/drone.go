package main

import (
	"context"
	"time"

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

	deliveryDrone, err := client.CreatePeerConnections(
		ctx, mediaEngine, interceptorRegistry,
		client.WithDataChannels(),
		client.WithMediaSources(
			mediasource.WithH264MediaEngine(constants.DefaultVideoClockRate, mediasource.PacketisationMode1, mediasource.ProfileLevelBaseline42),
			mediasource.WithNACKInterceptor(mediasource.NACKGeneratorLowLatency, mediasource.NACKResponderLowLatency),
			mediasource.WithFLEXFECInterceptor(),
			mediasource.WithJitterBufferInterceptor(),
			mediasource.WithRTCPReportsInterceptor(mediasource.RTCPReportIntervalLowLatency),
			mediasource.WithTWCCSenderInterceptor(mediasource.TWCCIntervalLowLatency),
			mediasource.WithBandwidthEstimatorInterceptor(2500, 50*time.Millisecond),
		),
	)
	if err != nil {
		panic(err)
	}

	_, err = deliveryDrone.CreatePeerConnection(
		"MAIN",
		client.WithRTCConfiguration(config.GetRTCConfiguration()),
		client.WithOfferSignal,
	)
	if err != nil {
		panic(err)
	}

	if err := deliveryDrone.CreateDataChannel("MAVLINK", "MAIN",
		data.WithRandomBindPort,
		// data.WithMAVP2P(os.Getenv("MAVP2P_EXE_PATH"), os.Getenv("MAVLINK_SERIAL")),
	); err != nil {
		panic(err)
	}

	if err := deliveryDrone.CreateMediaSource("MAIN",
		mediasource.WithH264Track(constants.DefaultVideoClockRate, "A8-MINI"),
		mediasource.WithPriority(mediasource.Level5),
		mediasource.WithStream(
			// mediasource.WithBufferSize(int(constants.DefaultVideoFPS*3)),
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

	if err := deliveryDrone.Connect("DELIVERY", "MAIN"); err != nil {
		panic(err)
	}
	deliveryDrone.WaitUntilClosed()
}
