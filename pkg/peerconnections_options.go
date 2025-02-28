package client

import (
	"fmt"
	"time"

	"github.com/pion/interceptor/pkg/flexfec"
	"github.com/pion/interceptor/pkg/jitterbuffer"
	"github.com/pion/interceptor/pkg/nack"
	"github.com/pion/interceptor/pkg/report"
	"github.com/pion/interceptor/pkg/twcc"
	"github.com/pion/sdp/v3"
	"github.com/pion/webrtc/v4"
)

type PeerConnectionsOption = func(*PeerConnections) error

type PacketisationMode uint8

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

func WithH264MediaEngine(clockrate uint32, packetisationMode PacketisationMode, profileLevelID ProfileLevel) PeerConnectionsOption {
	return func(pc *PeerConnections) error {
		if err := pc.mediaEngine.RegisterCodec(webrtc.RTPCodecParameters{
			RTPCodecCapability: webrtc.RTPCodecCapability{
				MimeType:    webrtc.MimeTypeH264,
				ClockRate:   clockrate,
				Channels:    0,
				SDPFmtpLine: fmt.Sprintf("level-asymmetry-allowed=1;packetization-mode=%d;profile-level-id=%s", packetisationMode, profileLevelID),
			},
			PayloadType: 96,
		}, webrtc.RTPCodecTypeVideo); err != nil {
			return err
		}
		return nil
	}
}

type StereoType uint8

const (
	Mono   StereoType = 0
	Stereo StereoType = 1
)

func WithOpusMediaEngine(samplerate uint32, channelLayout uint16, stereo StereoType) PeerConnectionsOption {
	return func(pc *PeerConnections) error {
		if err := pc.mediaEngine.RegisterCodec(webrtc.RTPCodecParameters{
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

func WithNACKInterceptor(generatorOptions NACKGeneratorOptions, responderOptions NACKResponderOptions) PeerConnectionsOption {
	return func(pc *PeerConnections) error {
		var (
			generator *nack.GeneratorInterceptorFactory
			responder *nack.ResponderInterceptorFactory
			err       error
		)
		if generator, err = nack.NewGeneratorInterceptor(generatorOptions...); err != nil {
			return err
		}
		if responder, err = nack.NewResponderInterceptor(responderOptions...); err != nil {
			return err
		}

		pc.mediaEngine.RegisterFeedback(webrtc.RTCPFeedback{Type: "nack"}, webrtc.RTPCodecTypeVideo)
		pc.interceptorRegistry.Add(responder)
		pc.interceptorRegistry.Add(generator)

		return nil
	}
}

type TWCCSenderInterval time.Duration

const (
	TWCCIntervalLowLatency   = TWCCSenderInterval(50 * time.Millisecond)
	TWCCIntervalDefault      = TWCCSenderInterval(100 * time.Millisecond)
	TWCCIntervalHighQuality  = TWCCSenderInterval(200 * time.Millisecond)
	TWCCIntervalLowBandwidth = TWCCSenderInterval(500 * time.Millisecond)
)

func WithTWCCSenderInterceptor(interval TWCCSenderInterval) PeerConnectionsOption {
	return func(pc *PeerConnections) error {
		var (
			generator *twcc.SenderInterceptorFactory
			err       error
		)

		pc.mediaEngine.RegisterFeedback(webrtc.RTCPFeedback{Type: webrtc.TypeRTCPFBTransportCC}, webrtc.RTPCodecTypeVideo)
		if err := pc.mediaEngine.RegisterHeaderExtension(webrtc.RTPHeaderExtensionCapability{URI: sdp.TransportCCURI}, webrtc.RTPCodecTypeVideo); err != nil {
			return err
		}

		pc.mediaEngine.RegisterFeedback(webrtc.RTCPFeedback{Type: webrtc.TypeRTCPFBTransportCC}, webrtc.RTPCodecTypeAudio)
		if err := pc.mediaEngine.RegisterHeaderExtension(webrtc.RTPHeaderExtensionCapability{URI: sdp.TransportCCURI}, webrtc.RTPCodecTypeAudio); err != nil {
			return err
		}

		if generator, err = twcc.NewSenderInterceptor(twcc.SendInterval(time.Duration(interval))); err != nil {
			return err
		}

		pc.interceptorRegistry.Add(generator)
		return nil
	}
}

// WARN: DO NOT USE THIS, PION HAS SOME ISSUE WITH THIS WHICH MAKES THE ONTRACK CALLBACK NOT FIRE
func WithJitterBufferInterceptor() PeerConnectionsOption {
	return func(pc *PeerConnections) error {
		var (
			jitterBuffer *jitterbuffer.InterceptorFactory
			err          error
		)

		if jitterBuffer, err = jitterbuffer.NewInterceptor(); err != nil {
			return err
		}
		pc.interceptorRegistry.Add(jitterBuffer)
		return nil
	}
}

type RTCPReportInterval time.Duration

const (
	RTCPReportIntervalLowLatency   = RTCPReportInterval(50 * time.Millisecond)
	RTCPReportIntervalDefault      = RTCPReportInterval(1 * time.Second)
	RTCPReportIntervalHighQuality  = RTCPReportInterval(200 * time.Millisecond)
	RTCPReportIntervalLowBandwidth = RTCPReportInterval(2 * time.Second)
)

func WithRTCPReportsInterceptor(interval RTCPReportInterval) PeerConnectionsOption {
	return func(pc *PeerConnections) error {
		var (
			sender   *report.SenderInterceptorFactory
			receiver *report.ReceiverInterceptorFactory
			err      error
		)

		if sender, err = report.NewSenderInterceptor(report.SenderInterval(time.Duration(interval))); err != nil {
			return err
		}
		if receiver, err = report.NewReceiverInterceptor(report.ReceiverInterval(time.Duration(interval))); err != nil {
			return err
		}

		pc.interceptorRegistry.Add(receiver)
		pc.interceptorRegistry.Add(sender)

		return nil
	}
}

// WARN: DO NOT USE FLEXFEC YET, AS THE FECOPTION ARE NOT YET IMPLEMENTED
func WithFLEXFECInterceptor() PeerConnectionsOption {
	return func(pc *PeerConnections) error {
		var (
			fecInterceptor *flexfec.FecInterceptorFactory
			err            error
		)

		// NOTE: Pion's FLEXFEC does not implement FecOption yet, if needed, someone needs to contribute to the repo
		if fecInterceptor, err = flexfec.NewFecInterceptor(); err != nil {
			return err
		}

		pc.interceptorRegistry.Add(fecInterceptor)
		return nil
	}
}
