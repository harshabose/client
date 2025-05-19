package client

import (
	"fmt"
	"time"

	"github.com/pion/interceptor/pkg/cc"
	"github.com/pion/interceptor/pkg/flexfec"
	"github.com/pion/interceptor/pkg/gcc"
	"github.com/pion/interceptor/pkg/jitterbuffer"
	"github.com/pion/interceptor/pkg/nack"
	"github.com/pion/interceptor/pkg/report"
	"github.com/pion/interceptor/pkg/twcc"
	"github.com/pion/sdp/v3"
	"github.com/pion/webrtc/v4"
)

type ClientOption = func(*Client) error

type PacketisationMode uint8

const (
	H264PayloadType    webrtc.PayloadType = 102
	H264RTXPayloadType webrtc.PayloadType = 103
)

const (
	PacketisationMode0 PacketisationMode = 0
	PacketisationMode1 PacketisationMode = 1
	PacketisationMode2 PacketisationMode = 2
)

type ProfileLevel string

const (
	ProfileLevelBaseline21 ProfileLevel = "420015" // Level 2.1 (480p)
	ProfileLevelBaseline31 ProfileLevel = "42001f" // Level 3.1 (720p)
	ProfileLevelBaseline41 ProfileLevel = "420029" // Level 4.1 (1080p)
	ProfileLevelBaseline42 ProfileLevel = "42002a" // Level 4.2 (2K)

	ProfileLevelMain21 ProfileLevel = "4D0015" // Level 2.1
	ProfileLevelMain31 ProfileLevel = "4D001f" // Level 3.1
	ProfileLevelMain41 ProfileLevel = "4D0029" // Level 4.1
	ProfileLevelMain42 ProfileLevel = "4D002a" // Level 4.2

	ProfileLevelHigh21 ProfileLevel = "640015" // Level 2.1
	ProfileLevelHigh31 ProfileLevel = "64001f" // Level 3.1
	ProfileLevelHigh41 ProfileLevel = "640029" // Level 4.1
	ProfileLevelHigh42 ProfileLevel = "64002a" // Level 4.2
)

func WithH264MediaEngine(clockrate uint32, packetisationMode PacketisationMode, profileLevelID ProfileLevel, sps, pps string) ClientOption {
	return func(client *Client) error {
		RTCPFeedback := []webrtc.RTCPFeedback{{Type: webrtc.TypeRTCPFBGoogREMB}, {Type: webrtc.TypeRTCPFBCCM, Parameter: "fir"}, {Type: webrtc.TypeRTCPFBNACK}, {Type: webrtc.TypeRTCPFBNACK, Parameter: "pli"}}
		if err := client.mediaEngine.RegisterCodec(webrtc.RTPCodecParameters{
			RTPCodecCapability: webrtc.RTPCodecCapability{
				MimeType:     webrtc.MimeTypeH264,
				ClockRate:    clockrate,
				Channels:     0,
				SDPFmtpLine:  fmt.Sprintf("level-asymmetry-allowed=1;packetization-mode=%d;profile-level-id=%s;sprop-parameter-sets=%s,%s", packetisationMode, profileLevelID, sps, pps),
				RTCPFeedback: RTCPFeedback,
			},
			PayloadType: H264PayloadType,
		}, webrtc.RTPCodecTypeVideo); err != nil {
			return err
		}

		if err := client.mediaEngine.RegisterCodec(webrtc.RTPCodecParameters{
			RTPCodecCapability: webrtc.RTPCodecCapability{
				MimeType:     webrtc.MimeTypeRTX,
				ClockRate:    clockrate,
				Channels:     0,
				SDPFmtpLine:  fmt.Sprintf("apt=%d", H264PayloadType),
				RTCPFeedback: nil,
			},
			PayloadType: H264RTXPayloadType,
		}, webrtc.RTPCodecTypeVideo); err != nil {
			return err
		}

		return nil
	}
}

func WithDefaultMediaEngine() ClientOption {
	return func(client *Client) error {
		if err := client.mediaEngine.RegisterDefaultCodecs(); err != nil {
			return err
		}
		return nil
	}
}

func WithDefaultInterceptorRegistry() ClientOption {
	return func(client *Client) error {
		if err := webrtc.RegisterDefaultInterceptors(client.mediaEngine, client.interceptorRegistry); err != nil {
			return err
		}
		return nil
	}
}

type StereoType uint8

const (
	Mono StereoType = 0
	Dual StereoType = 1
)

func WithOpusMediaEngine(samplerate uint32, channelLayout uint16, stereo StereoType) ClientOption {
	return func(client *Client) error {
		if err := client.mediaEngine.RegisterCodec(webrtc.RTPCodecParameters{
			RTPCodecCapability: webrtc.RTPCodecCapability{
				MimeType:    webrtc.MimeTypeOpus,
				ClockRate:   samplerate,
				Channels:    channelLayout,
				SDPFmtpLine: fmt.Sprintf("minptime=10;useinbandfec=1;stereo=%d", stereo),
			},
			PayloadType: 111,
		}, webrtc.RTPCodecTypeAudio); err != nil {
			return err
		}
		return nil
	}
}

type NACKGeneratorOptions []nack.GeneratorOption

var (
	NACKGeneratorLowLatency   NACKGeneratorOptions = []nack.GeneratorOption{nack.GeneratorSize(256), nack.GeneratorSkipLastN(2), nack.GeneratorMaxNacksPerPacket(1), nack.GeneratorInterval(50 * time.Millisecond)}
	NACKGeneratorDefault      NACKGeneratorOptions = []nack.GeneratorOption{nack.GeneratorSize(512), nack.GeneratorSkipLastN(5), nack.GeneratorMaxNacksPerPacket(2), nack.GeneratorInterval(100 * time.Millisecond)}
	NACKGeneratorHighQuality  NACKGeneratorOptions = []nack.GeneratorOption{nack.GeneratorSize(2048), nack.GeneratorSkipLastN(10), nack.GeneratorMaxNacksPerPacket(3), nack.GeneratorInterval(200 * time.Millisecond)}
	NACKGeneratorLowBandwidth NACKGeneratorOptions = []nack.GeneratorOption{nack.GeneratorSize(4096), nack.GeneratorSkipLastN(15), nack.GeneratorMaxNacksPerPacket(4), nack.GeneratorInterval(150 * time.Millisecond)}
)

type NACKResponderOptions []nack.ResponderOption

var (
	NACKResponderLowLatency   NACKResponderOptions = []nack.ResponderOption{nack.ResponderSize(256), nack.DisableCopy()}
	NACKResponderDefault      NACKResponderOptions = []nack.ResponderOption{nack.ResponderSize(1024)}
	NACKResponderHighQuality  NACKResponderOptions = []nack.ResponderOption{nack.ResponderSize(2048)}
	NACKResponderLowBandwidth NACKResponderOptions = []nack.ResponderOption{nack.ResponderSize(4096)}
)

func WithNACKInterceptor(generatorOptions NACKGeneratorOptions, responderOptions NACKResponderOptions) ClientOption {
	return func(client *Client) error {
		var (
			generator *nack.GeneratorInterceptorFactory
			responder *nack.ResponderInterceptorFactory
			err       error
		)
		if generator, err = nack.NewGeneratorInterceptor(); err != nil {
			return err
		}
		if responder, err = nack.NewResponderInterceptor(); err != nil {
			return err
		}

		client.mediaEngine.RegisterFeedback(webrtc.RTCPFeedback{Type: webrtc.TypeRTCPFBNACK}, webrtc.RTPCodecTypeVideo)
		client.mediaEngine.RegisterFeedback(webrtc.RTCPFeedback{Type: webrtc.TypeRTCPFBNACK, Parameter: "pli"}, webrtc.RTPCodecTypeVideo)
		client.interceptorRegistry.Add(responder)
		client.interceptorRegistry.Add(generator)

		return nil
	}
}

type TWCCSenderInterval time.Duration

const (
	TWCCIntervalLowLatency   = TWCCSenderInterval(100 * time.Millisecond)
	TWCCIntervalDefault      = TWCCSenderInterval(100 * time.Millisecond)
	TWCCIntervalHighQuality  = TWCCSenderInterval(200 * time.Millisecond)
	TWCCIntervalLowBandwidth = TWCCSenderInterval(500 * time.Millisecond)
)

func WithTWCCSenderInterceptor(interval TWCCSenderInterval) ClientOption {
	return func(client *Client) error {
		var (
			generator *twcc.SenderInterceptorFactory
			err       error
		)

		client.mediaEngine.RegisterFeedback(webrtc.RTCPFeedback{Type: webrtc.TypeRTCPFBTransportCC}, webrtc.RTPCodecTypeVideo)
		if err := client.mediaEngine.RegisterHeaderExtension(webrtc.RTPHeaderExtensionCapability{URI: sdp.TransportCCURI}, webrtc.RTPCodecTypeVideo); err != nil {
			return err
		}

		client.mediaEngine.RegisterFeedback(webrtc.RTCPFeedback{Type: webrtc.TypeRTCPFBTransportCC}, webrtc.RTPCodecTypeAudio)
		if err := client.mediaEngine.RegisterHeaderExtension(webrtc.RTPHeaderExtensionCapability{URI: sdp.TransportCCURI}, webrtc.RTPCodecTypeAudio); err != nil {
			return err
		}

		if generator, err = twcc.NewSenderInterceptor(); err != nil {
			return err
		}

		client.interceptorRegistry.Add(generator)
		return nil
	}
}

// WARN: DO NOT USE THIS, PION HAS SOME ISSUE WITH THIS WHICH MAKES THE ONTRACK CALLBACK NOT FIRE
func WithJitterBufferInterceptor() ClientOption {
	return func(client *Client) error {
		var (
			jitterBuffer *jitterbuffer.InterceptorFactory
			err          error
		)

		if jitterBuffer, err = jitterbuffer.NewInterceptor(); err != nil {
			return err
		}
		client.interceptorRegistry.Add(jitterBuffer)
		return nil
	}
}

type RTCPReportInterval time.Duration

const (
	RTCPReportIntervalLowLatency   = RTCPReportInterval(1000 * time.Millisecond)
	RTCPReportIntervalDefault      = RTCPReportInterval(1 * time.Second)
	RTCPReportIntervalHighQuality  = RTCPReportInterval(1500 * time.Millisecond)
	RTCPReportIntervalLowBandwidth = RTCPReportInterval(2 * time.Second)
)

func WithRTCPReportsInterceptor(interval RTCPReportInterval) ClientOption {
	return func(client *Client) error {
		sender, err := report.NewSenderInterceptor()
		if err != nil {
			return err
		}
		receiver, err := report.NewReceiverInterceptor()
		if err != nil {
			return err
		}

		client.interceptorRegistry.Add(receiver)
		client.interceptorRegistry.Add(sender)

		return nil
	}
}

// WARN: DO NOT USE FLEXFEC YET, AS THE FECOPTION ARE NOT YET IMPLEMENTED
func WithFLEXFECInterceptor() ClientOption {
	return func(client *Client) error {
		var (
			fecInterceptor *flexfec.FecInterceptorFactory
			err            error
		)

		// NOTE: Pion's FLEXFEC does not implement FecOption yet, if needed, someone needs to contribute to the repo
		if fecInterceptor, err = flexfec.NewFecInterceptor(); err != nil {
			return err
		}

		client.interceptorRegistry.Add(fecInterceptor)
		return nil
	}
}

func WithSimulcastExtensionHeaders() ClientOption {
	return func(client *Client) error {
		return webrtc.ConfigureSimulcastExtensionHeaders(client.mediaEngine)
	}
}

func WithBandwidthControlInterceptor(initialBitrate int, interval time.Duration) ClientOption {
	return func(client *Client) error {
		congestionController, err := cc.NewInterceptor(func() (cc.BandwidthEstimator, error) {
			return gcc.NewSendSideBWE(gcc.SendSideBWEInitialBitrate(initialBitrate), gcc.SendSideBWEMaxBitrate(initialBitrate*2))
		})
		if err != nil {
			return err
		}

		congestionController.OnNewPeerConnection(func(id string, estimator cc.BandwidthEstimator) {
			fmt.Println("sending estimator")
			client.estimatorChan <- estimator
			fmt.Println("send estimator")
		})

		client.interceptorRegistry.Add(congestionController)

		// TODO: NOT SURE IF I NEED THE FOLLOWING
		// client.mediaEngine.RegisterFeedback(webrtc.RTCPFeedback{Type: webrtc.TypeRTCPFBGoogREMB}, webrtc.RTPCodecTypeVideo)

		return nil
	}
}

func WithTWCCHeaderExtensionSender() ClientOption {
	return func(client *Client) error {
		return webrtc.ConfigureTWCCHeaderExtensionSender(client.mediaEngine, client.interceptorRegistry)
	}
}
