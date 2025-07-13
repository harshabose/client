package mediasource

import (
	"errors"

	"github.com/pion/webrtc/v4"
)

type TrackOption = func(*Track) error

func WithH264Track(clockrate uint32, packetisationMode PacketisationMode, profileLevel ProfileLevel) TrackOption {
	return func(track *Track) error {
		if track.codecCapability != nil {
			return errors.New("multiple tracks are not supported on single media source")
		}
		track.codecCapability = &webrtc.RTPCodecCapability{}
		track.codecCapability.MimeType = webrtc.MimeTypeH264
		track.codecCapability.ClockRate = clockrate
		track.codecCapability.Channels = 0

		return nil
	}
}

func WithVP8Track(clockrate uint32) TrackOption {
	return func(track *Track) error {
		if track.codecCapability != nil {
			return errors.New("multiple tracks are not supported on single media source")
		}
		track.codecCapability = &webrtc.RTPCodecCapability{}
		track.codecCapability.MimeType = webrtc.MimeTypeVP8
		track.codecCapability.ClockRate = clockrate
		track.codecCapability.Channels = 0

		return nil
	}
}

func WithOpusTrack(samplerate uint32, channelLayout uint16, stereo StereoType) TrackOption {
	return func(track *Track) error {
		if track.codecCapability != nil {
			return errors.New("multiple tracks are not supported on single media source")
		}
		track.codecCapability = &webrtc.RTPCodecCapability{}
		track.codecCapability.MimeType = webrtc.MimeTypeOpus
		track.codecCapability.ClockRate = samplerate
		track.codecCapability.Channels = channelLayout

		return nil
	}
}

func WithPriority(level Priority) TrackOption {
	return func(track *Track) error {
		track.priority = level
		return nil
	}
}
