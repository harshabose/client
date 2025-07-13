package client

import (
	"time"

	"github.com/pion/webrtc/v4"

	"github.com/harshabose/simple_webrtc_comm/client/pkg/datachannel"
	"github.com/harshabose/simple_webrtc_comm/client/pkg/mediasink"
	"github.com/harshabose/simple_webrtc_comm/client/pkg/mediasource"
)

type ClientConfig struct {
	Name string
	// Media configuration
	H264 *H264Config `json:"h264,omitempty"`
	VP8  *VP8Config  `json:"vp8,omitempty"`
	Opus *OpusConfig `json:"opus,omitempty"`

	// Interceptor configurations
	NACK        *NACKPreset        `json:"nack,omitempty"`
	RTCPReports *RTCPReportsPreset `json:"rtcp_reports,omitempty"`
	TWCC        *TWCCPreset        `json:"twcc,omitempty"`
	Bandwidth   *BandwidthConfig   `json:"bandwidth,omitempty"`

	// Feature flags
	SimulcastExtensions bool `json:"simulcast_extensions,omitempty"`
	TWCCHeaderExtension bool `json:"twcc_header_extension,omitempty"`
}

type H264Config struct {
	ClockRate         uint32                        `json:"clock_rate"`
	PacketisationMode mediasource.PacketisationMode `json:"packetisation_mode"`
	ProfileLevel      mediasource.ProfileLevel      `json:"profile_level"`
	SPSBase64         string                        `json:"sps_base64"`
	PPSBase64         string                        `json:"pps_base64"`
}

type VP8Config struct {
	ClockRate uint32 `json:"clock_rate"`
}

type OpusConfig struct {
	SampleRate    uint32                 `json:"sample_rate"`
	ChannelLayout uint16                 `json:"channel_layout"`
	Stereo        mediasource.StereoType `json:"stereo"`
}

type NACKPreset string
type RTCPReportsPreset string
type TWCCPreset string

const (
	NACKLowLatency   NACKPreset = "low_latency"
	NACKDefault      NACKPreset = "default"
	NACKHighQuality  NACKPreset = "high_quality"
	NACKLowBandwidth NACKPreset = "low_bandwidth"

	RTCPReportsLowLatency   RTCPReportsPreset = "low_latency"
	RTCPReportsDefault      RTCPReportsPreset = "default"
	RTCPReportsHighQuality  RTCPReportsPreset = "high_quality"
	RTCPReportsLowBandwidth RTCPReportsPreset = "low_bandwidth"

	TWCCLowLatency   TWCCPreset = "low_latency"
	TWCCDefault      TWCCPreset = "default"
	TWCCHighQuality  TWCCPreset = "high_quality"
	TWCCLowBandwidth TWCCPreset = "low_bandwidth"
)

type BandwidthConfig struct {
	Initial  uint64        `json:"initial"`
	Minimum  uint64        `json:"minimum"`
	Maximum  uint64        `json:"maximum"`
	Interval time.Duration `json:"interval"`
}

type optionBuilder struct {
	options []ClientOption
}

func (ob *optionBuilder) add(option ClientOption) *optionBuilder {
	if option != nil {
		ob.options = append(ob.options, option)
	}
	return ob
}

var (
	nackGeneratorPresets = map[NACKPreset]NACKGeneratorOptions{
		NACKLowLatency:   NACKGeneratorLowLatency,
		NACKDefault:      NACKGeneratorDefault,
		NACKHighQuality:  NACKGeneratorHighQuality,
		NACKLowBandwidth: NACKGeneratorLowBandwidth,
	}

	nackResponderPresets = map[NACKPreset]NACKResponderOptions{
		NACKLowLatency:   NACKResponderLowLatency,
		NACKDefault:      NACKResponderDefault,
		NACKHighQuality:  NACKResponderHighQuality,
		NACKLowBandwidth: NACKResponderLowBandwidth,
	}

	rtcpReportsPresets = map[RTCPReportsPreset]RTCPReportInterval{
		RTCPReportsLowLatency:   RTCPReportIntervalLowLatency,
		RTCPReportsDefault:      RTCPReportIntervalDefault,
		RTCPReportsHighQuality:  RTCPReportIntervalHighQuality,
		RTCPReportsLowBandwidth: RTCPReportIntervalLowBandwidth,
	}

	twccPresets = map[TWCCPreset]TWCCSenderInterval{
		TWCCLowLatency:   TWCCIntervalLowLatency,
		TWCCDefault:      TWCCIntervalDefault,
		TWCCHighQuality:  TWCCIntervalHighQuality,
		TWCCLowBandwidth: TWCCIntervalLowBandwidth,
	}
)

func (c *ClientConfig) ToOptions() []ClientOption {
	builder := &optionBuilder{}

	return builder.
		add(c.h264Option()).
		add(c.vp8Option()).
		add(c.opusOption()).
		add(c.nackOption()).
		add(c.rtcpReportsOption()).
		add(c.twccOption()).
		add(c.bandwidthOption()).
		add(c.simulcastOption()).
		add(c.twccHeaderOption()).
		options
}

func (c *ClientConfig) h264Option() ClientOption {
	if c.H264 == nil {
		return nil
	}
	return WithH264MediaEngine(
		c.H264.ClockRate,
		c.H264.PacketisationMode,
		c.H264.ProfileLevel,
		c.H264.SPSBase64,
		c.H264.PPSBase64,
	)
}

func (c *ClientConfig) vp8Option() ClientOption {
	if c.VP8 == nil {
		return nil
	}

	return WithVP8MediaEngine(c.VP8.ClockRate)
}

func (c *ClientConfig) opusOption() ClientOption {
	if c.Opus == nil {
		return nil
	}

	return WithOpusMediaEngine(
		c.Opus.SampleRate,
		c.Opus.ChannelLayout,
		c.Opus.Stereo,
	)
}

func (c *ClientConfig) nackOption() ClientOption {
	if c.NACK == nil {
		return nil
	}

	generator, generatorExists := nackGeneratorPresets[*c.NACK]
	responder, responderExists := nackResponderPresets[*c.NACK]

	if !generatorExists || !responderExists {
		return nil
	}

	return WithNACKInterceptor(generator, responder)
}

func (c *ClientConfig) rtcpReportsOption() ClientOption {
	if c.RTCPReports == nil {
		return nil
	}

	interval, exists := rtcpReportsPresets[*c.RTCPReports]
	if !exists {
		return nil
	}

	return WithRTCPReportsInterceptor(interval)
}

func (c *ClientConfig) twccOption() ClientOption {
	if c.TWCC == nil {
		return nil
	}

	interval, exists := twccPresets[*c.TWCC]
	if !exists {
		return nil
	}

	return WithTWCCSenderInterceptor(interval)
}

func (c *ClientConfig) bandwidthOption() ClientOption {
	if c.Bandwidth == nil {
		return nil
	}
	return WithBandwidthControlInterceptor(
		c.Bandwidth.Initial,
		c.Bandwidth.Minimum,
		c.Bandwidth.Maximum,
		c.Bandwidth.Interval,
	)
}

func (c *ClientConfig) simulcastOption() ClientOption {
	if !c.SimulcastExtensions {
		return nil
	}
	return WithSimulcastExtensionHeaders()
}

func (c *ClientConfig) twccHeaderOption() ClientOption {
	if !c.TWCCHeaderExtension {
		return nil
	}
	return WithTWCCHeaderExtensionSender()
}

type PeerConnectionConfig struct {
	// Basic settings
	Name string `json:"name"`

	// RTC and Singnaling Control
	FirebaseOfferSignal *bool                `json:"firebase_offer_signal,omitempty"`
	FirebaseOfferAnswer *bool                `json:"firebase_offer_answer"`
	RTCConfig           webrtc.Configuration `json:"rtc_config"`

	// Declarative resource definitions
	DataChannels []DataChannelSpec `json:"data_channels_specs,omitempty"`
	MediaSources []MediaSourceSpec `json:"media_sources_specs,omitempty"`
	MediaSinks   []MediaSinkSpec   `json:"media_sinks_specs,omitempty"`
}

type DataChannelSpec struct {
	Label             string  `json:"label"`
	ID                *uint16 `json:"id,omitempty"`
	Ordered           *bool   `json:"ordered,omitempty"`
	Protocol          *string `json:"protocol,omitempty"`
	Negotiated        *bool   `json:"negotiated,omitempty"`
	MaxPacketLifeTime *uint16 `json:"max_packet_life_time,omitempty"`
	MaxRetransmits    *uint16 `json:"max_retransmits,omitempty"`
}

type MediaSourceSpec struct {
	Name     string                `json:"name"`
	H264     *H264TrackConfig      `json:"h264,omitempty"`
	VP8      *VP8TrackConfig       `json:"vp8,omitempty"`
	Opus     *OpusTrackConfig      `json:"opus,omitempty"`
	Priority *mediasource.Priority `json:"priority,omitempty"`
}

type trackOptionBuilder struct {
	options []mediasource.TrackOption
}

func (ob *trackOptionBuilder) add(option mediasource.TrackOption) *trackOptionBuilder {
	if option != nil {
		ob.options = append(ob.options, option)
	}
	return ob
}

func (c *MediaSourceSpec) withTrackOption() mediasource.TrackOption {
	if c.H264 != nil {
		return mediasource.WithH264Track(c.H264.ClockRate, c.H264.PacketisationMode, c.H264.ProfileLevel)
	}

	if c.VP8 != nil {
		return mediasource.WithVP8Track(c.VP8.ClockRate)
	}

	if c.Opus != nil {
		return mediasource.WithOpusTrack(c.Opus.Samplerate, c.Opus.ChannelLayout, c.Opus.Stereo)
	}

	return nil
}

func (c *MediaSourceSpec) ToOptions() []mediasource.TrackOption {
	builder := trackOptionBuilder{}

	return builder.add(c.withTrackOption()).add(c.withPriority()).options
}

func (c *MediaSourceSpec) withPriority() mediasource.TrackOption {
	if c.Priority == nil {
		return nil
	}

	return mediasource.WithPriority(*c.Priority)
}

// MediaSinkSpec defines a media sink to create
type MediaSinkSpec struct {
	Name string           `json:"name"`
	H264 *H264TrackConfig `json:"h264,omitempty"`
	VP8  *VP8TrackConfig  `json:"vp8,omitempty"`
	Opus *OpusTrackConfig `json:"opus,omitempty"`
}

type sinkOptionBuilder struct {
	options []mediasink.SinkOption
}

func (ob *sinkOptionBuilder) add(option mediasink.SinkOption) *sinkOptionBuilder {
	if option != nil {
		ob.options = append(ob.options, option)
	}
	return ob
}

func (c *MediaSinkSpec) withTrackOption() mediasink.SinkOption {
	if c.H264 != nil {
		return mediasink.WithH264Track(c.H264.ClockRate)
	}

	if c.VP8 != nil {
		return mediasink.WithVP8Track(c.VP8.ClockRate)
	}

	if c.Opus != nil {
		return mediasink.WithOpusTrack(c.Opus.Samplerate, c.Opus.ChannelLayout)
	}

	return nil
}

func (c *MediaSinkSpec) ToOptions() []mediasink.SinkOption {
	builder := sinkOptionBuilder{}

	return builder.add(c.withTrackOption()).options
}

type H264TrackConfig struct {
	ClockRate         uint32                        `json:"clock_rate"`
	PacketisationMode mediasource.PacketisationMode `json:"packetisation_mode"`
	ProfileLevel      mediasource.ProfileLevel      `json:"profile_level"`
}

type VP8TrackConfig struct {
	ClockRate uint32 `json:"clock_rate"`
}

type OpusTrackConfig struct {
	Samplerate    uint32                 `json:"samplerate"`
	ChannelLayout uint16                 `json:"channel_layout"`
	Stereo        mediasource.StereoType `json:"stereo"`
}

type pcOptionBuilder struct {
	options []PeerConnectionOption
}

func (ob *pcOptionBuilder) add(option PeerConnectionOption) *pcOptionBuilder {
	if option != nil {
		ob.options = append(ob.options, option)
	}
	return ob
}

func (c *PeerConnectionConfig) ToOptions() []PeerConnectionOption {
	builder := pcOptionBuilder{}

	return builder.
		add(c.withFirebaseOfferSignal()).
		add(c.withFirebaseOfferAnswer()).
		add(c.withDataChannels()).
		add(c.withMediaSource()).
		add(c.withMediaSinks()).
		options
}

func (c *PeerConnectionConfig) withFirebaseOfferSignal() PeerConnectionOption {
	if c.FirebaseOfferSignal == nil {
		return nil
	}
	return WithFirebaseOfferSignal
}

func (c *PeerConnectionConfig) withFirebaseOfferAnswer() PeerConnectionOption {
	if c.FirebaseOfferAnswer == nil {
		return nil
	}
	return WithFirebaseAnswerSignal
}

func (c *PeerConnectionConfig) withMediaSource() PeerConnectionOption {
	if len(c.MediaSources) == 0 {
		return nil
	}
	return WithMediaSources()
}

func (c *PeerConnectionConfig) withMediaSinks() PeerConnectionOption {
	if len(c.MediaSinks) == 0 {
		return nil
	}
	return WithMediaSinks()
}

func (c *PeerConnectionConfig) withDataChannels() PeerConnectionOption {
	if len(c.DataChannels) == 0 {
		return nil
	}
	return WithDataChannels()
}

func (c *PeerConnectionConfig) CreateDataChannels(pc *PeerConnection) error {
	if len(c.DataChannels) == 0 {
		return nil
	}

	for _, config := range c.DataChannels {
		if _, err := pc.CreateDataChannel(config.Label, datachannel.WithDataChannelInit(&webrtc.DataChannelInit{
			Ordered:           config.Ordered,
			MaxPacketLifeTime: config.MaxPacketLifeTime,
			MaxRetransmits:    config.MaxRetransmits,
			Protocol:          config.Protocol,
			Negotiated:        config.Negotiated,
			ID:                config.ID,
		})); err != nil {
			return err
		}
	}
	return nil
}

func (c *PeerConnectionConfig) CreateMediaSources(pc *PeerConnection) error {
	if len(c.MediaSources) == 0 {
		return nil
	}

	for _, config := range c.MediaSources {
		if _, err := pc.CreateMediaSource(c.Name, config.ToOptions()...); err != nil {
			return err
		}
	}

	return nil
}

func (c *PeerConnectionConfig) CreateMediaSinks(pc *PeerConnection) error {
	if len(c.MediaSinks) == 0 {
		return nil
	}

	for _, config := range c.MediaSinks {
		if _, err := pc.CreateMediaSink(c.Name, config.ToOptions()...); err != nil {
			return err
		}
	}

	return nil
}
