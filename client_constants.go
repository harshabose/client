package client

import (
	"time"

	"github.com/pion/interceptor/pkg/nack"
	"github.com/pion/webrtc/v4"
)

const (
	H264PayloadType    webrtc.PayloadType = 102
	H264RTXPayloadType webrtc.PayloadType = 103
	VP8PayloadType     webrtc.PayloadType = 96
	VP8RTXPayloadType  webrtc.PayloadType = 97
	OpusPayloadType    webrtc.PayloadType = 111
)

type NACKGeneratorOptions []nack.GeneratorOption

var (
	NACKGeneratorLowLatency   NACKGeneratorOptions = []nack.GeneratorOption{nack.GeneratorSize(256), nack.GeneratorSkipLastN(2), nack.GeneratorMaxNacksPerPacket(1), nack.GeneratorInterval(10 * time.Millisecond)}
	NACKGeneratorDefault      NACKGeneratorOptions = []nack.GeneratorOption{nack.GeneratorSize(512), nack.GeneratorSkipLastN(5), nack.GeneratorMaxNacksPerPacket(2), nack.GeneratorInterval(50 * time.Millisecond)}
	NACKGeneratorHighQuality  NACKGeneratorOptions = []nack.GeneratorOption{nack.GeneratorSize(4096), nack.GeneratorSkipLastN(10), nack.GeneratorMaxNacksPerPacket(5), nack.GeneratorInterval(100 * time.Millisecond)}
	NACKGeneratorLowBandwidth NACKGeneratorOptions = []nack.GeneratorOption{nack.GeneratorSize(256), nack.GeneratorSkipLastN(15), nack.GeneratorMaxNacksPerPacket(1), nack.GeneratorInterval(200 * time.Millisecond)}
)

type NACKResponderOptions []nack.ResponderOption

var (
	NACKResponderLowLatency   NACKResponderOptions = []nack.ResponderOption{nack.ResponderSize(256)}
	NACKResponderDefault      NACKResponderOptions = []nack.ResponderOption{nack.ResponderSize(1024)}
	NACKResponderHighQuality  NACKResponderOptions = []nack.ResponderOption{nack.ResponderSize(4096)}
	NACKResponderLowBandwidth NACKResponderOptions = []nack.ResponderOption{nack.ResponderSize(256)}
)

type TWCCSenderInterval time.Duration

const (
	TWCCIntervalLowLatency   = TWCCSenderInterval(100 * time.Millisecond)
	TWCCIntervalDefault      = TWCCSenderInterval(200 * time.Millisecond)
	TWCCIntervalHighQuality  = TWCCSenderInterval(300 * time.Millisecond)
	TWCCIntervalLowBandwidth = TWCCSenderInterval(500 * time.Millisecond)
)

type RTCPReportInterval time.Duration

const (
	RTCPReportIntervalLowLatency   = RTCPReportInterval(1 * time.Second)
	RTCPReportIntervalDefault      = RTCPReportInterval(3 * time.Second)
	RTCPReportIntervalHighQuality  = RTCPReportInterval(2 * time.Second)
	RTCPReportIntervalLowBandwidth = RTCPReportInterval(10 * time.Second)
)
