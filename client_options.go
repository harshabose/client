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
	"github.com/pion/interceptor/pkg/stats"
	"github.com/pion/interceptor/pkg/twcc"
	"github.com/pion/sdp/v3"
	"github.com/pion/webrtc/v4"
)

type ClientOption = func(*Client) error

func WithH264MediaEngine(clockrate uint32) ClientOption {
	return func(client *Client) error {
		RTCPFeedback := []webrtc.RTCPFeedback{{Type: webrtc.TypeRTCPFBGoogREMB}, {Type: webrtc.TypeRTCPFBCCM, Parameter: "fir"}, {Type: webrtc.TypeRTCPFBNACK}, {Type: webrtc.TypeRTCPFBNACK, Parameter: "pli"}}
		if err := client.mediaEngine.RegisterCodec(webrtc.RTPCodecParameters{
			RTPCodecCapability: webrtc.RTPCodecCapability{
				MimeType:     webrtc.MimeTypeH264,
				ClockRate:    clockrate,
				Channels:     0,
				SDPFmtpLine:  fmt.Sprintf("level-asymmetry-allowed=1"),
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

func WithVP8MediaEngine(clockrate uint32) ClientOption {
	return func(client *Client) error {
		RTCPFeedback := []webrtc.RTCPFeedback{{Type: webrtc.TypeRTCPFBGoogREMB}, {Type: webrtc.TypeRTCPFBCCM, Parameter: "fir"}, {Type: webrtc.TypeRTCPFBNACK}, {Type: webrtc.TypeRTCPFBNACK, Parameter: "pli"}}
		if err := client.mediaEngine.RegisterCodec(webrtc.RTPCodecParameters{
			RTPCodecCapability: webrtc.RTPCodecCapability{
				MimeType:     webrtc.MimeTypeVP8,
				ClockRate:    clockrate,
				RTCPFeedback: RTCPFeedback,
				SDPFmtpLine:  fmt.Sprintf(""),
			},
			PayloadType: VP8PayloadType,
		}, webrtc.RTPCodecTypeVideo); err != nil {
			return err
		}

		if err := client.mediaEngine.RegisterCodec(webrtc.RTPCodecParameters{
			RTPCodecCapability: webrtc.RTPCodecCapability{
				MimeType:     webrtc.MimeTypeRTX,
				ClockRate:    clockrate,
				RTCPFeedback: nil,
				SDPFmtpLine:  fmt.Sprintf("apt=%d", VP8PayloadType),
			},
			PayloadType: VP8RTXPayloadType,
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

func WithOpusMediaEngine(samplerate uint32, channelLayout uint16) ClientOption {
	return func(client *Client) error {
		if err := client.mediaEngine.RegisterCodec(webrtc.RTPCodecParameters{
			RTPCodecCapability: webrtc.RTPCodecCapability{
				MimeType:     webrtc.MimeTypeOpus,
				ClockRate:    samplerate,
				Channels:     channelLayout,
				RTCPFeedback: nil,
				SDPFmtpLine:  fmt.Sprintf("minptime=10;useinbandfec=1"),
			},
			PayloadType: OpusPayloadType,
		}, webrtc.RTPCodecTypeAudio); err != nil {
			return err
		}

		return nil
	}
}

func WithNACKInterceptor(generatorOptions NACKGeneratorOptions, responderOptions NACKResponderOptions) ClientOption {
	return func(client *Client) error {
		generator, err := nack.NewGeneratorInterceptor(generatorOptions...)
		if err != nil {
			return err
		}
		responder, err := nack.NewResponderInterceptor(responderOptions...)
		if err != nil {
			return err
		}

		client.mediaEngine.RegisterFeedback(webrtc.RTCPFeedback{Type: webrtc.TypeRTCPFBNACK}, webrtc.RTPCodecTypeVideo)
		client.mediaEngine.RegisterFeedback(webrtc.RTCPFeedback{Type: webrtc.TypeRTCPFBNACK, Parameter: "pli"}, webrtc.RTPCodecTypeVideo)
		client.interceptorRegistry.Add(responder)
		client.interceptorRegistry.Add(generator)

		return nil
	}
}

func WithTWCCSenderInterceptor(interval TWCCSenderInterval) ClientOption {
	return func(client *Client) error {
		client.mediaEngine.RegisterFeedback(webrtc.RTCPFeedback{Type: webrtc.TypeRTCPFBTransportCC}, webrtc.RTPCodecTypeVideo)
		if err := client.mediaEngine.RegisterHeaderExtension(webrtc.RTPHeaderExtensionCapability{URI: sdp.TransportCCURI}, webrtc.RTPCodecTypeVideo); err != nil {
			return err
		}

		client.mediaEngine.RegisterFeedback(webrtc.RTCPFeedback{Type: webrtc.TypeRTCPFBTransportCC}, webrtc.RTPCodecTypeAudio)
		if err := client.mediaEngine.RegisterHeaderExtension(webrtc.RTPHeaderExtensionCapability{URI: sdp.TransportCCURI}, webrtc.RTPCodecTypeAudio); err != nil {
			return err
		}

		generator, err := twcc.NewSenderInterceptor(twcc.SendInterval(time.Duration(interval)))
		if err != nil {
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

func WithRTCPReportsInterceptor(interval RTCPReportInterval) ClientOption {
	return func(client *Client) error {
		receiver, err := report.NewReceiverInterceptor(report.ReceiverInterval(time.Duration(interval)))
		if err != nil {
			return err
		}
		sender, err := report.NewSenderInterceptor(report.SenderInterval(time.Duration(interval)))
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

func WithBandwidthControlInterceptor(initialBitrate, minimumBitrate, maximumBitrate int64, interval time.Duration) ClientOption {
	return func(client *Client) error {
		congestionController, err := cc.NewInterceptor(func() (cc.BandwidthEstimator, error) {
			return gcc.NewSendSideBWE(gcc.SendSideBWEInitialBitrate(int(initialBitrate)), gcc.SendSideBWEMinBitrate(int(minimumBitrate)), gcc.SendSideBWEMaxBitrate(int(maximumBitrate)))
		})
		if err != nil {
			return err
		}

		congestionController.OnNewPeerConnection(func(id string, estimator cc.BandwidthEstimator) {
			client.estimatorChan <- estimator
		})

		client.interceptorRegistry.Add(congestionController)

		return nil
	}
}

func WithTWCCHeaderExtensionSender() ClientOption {
	return func(client *Client) error {
		return webrtc.ConfigureTWCCHeaderExtensionSender(client.mediaEngine, client.interceptorRegistry)
	}
}

func WithStatsCollector() ClientOption {
	return func(c *Client) error {
		g, err := stats.NewInterceptor()
		if err != nil {
			return err
		}

		g.OnNewPeerConnection(func(id string, getter stats.Getter) {
			c.getterChan <- getter
		})

		c.interceptorRegistry.Add(g)

		return nil
	}
}
