package mediasink

import (
	"errors"

	"github.com/pion/webrtc/v4"
)

type SinkOption = func(*Sink) error

// TODO: CLOCKRATE, STEREO, PROFILE etc ARE IN MEDIA SOURCE. MAYBE BE INCLUDE THEM HERE?

func WithH264Track(clockrate uint32) SinkOption {
	return func(track *Sink) error {
		if track.codecCapability != nil {
			return errors.New("multiple tracks are not supported on single media source")
		}
		track.codecCapability = &webrtc.RTPCodecParameters{}
		track.codecCapability.PayloadType = webrtc.PayloadType(102)
		track.codecCapability.MimeType = webrtc.MimeTypeH264
		track.codecCapability.ClockRate = clockrate
		track.codecCapability.Channels = 0

		return nil
	}
}

func WithVP8Track(clockrate uint32) SinkOption {
	return func(track *Sink) error {
		if track.codecCapability != nil {
			return errors.New("multiple tracks are not supported on single media source")
		}
		track.codecCapability = &webrtc.RTPCodecParameters{}
		track.codecCapability.PayloadType = webrtc.PayloadType(96)
		track.codecCapability.MimeType = webrtc.MimeTypeVP8
		track.codecCapability.ClockRate = clockrate
		track.codecCapability.Channels = 0

		return nil
	}
}

func WithOpusTrack(samplerate uint32, channelLayout uint16) SinkOption {
	return func(track *Sink) error {
		if track.codecCapability != nil {
			return errors.New("multiple tracks are not supported on single media source")
		}
		track.codecCapability = &webrtc.RTPCodecParameters{}
		track.codecCapability.PayloadType = webrtc.PayloadType(111)
		track.codecCapability.MimeType = webrtc.MimeTypeOpus
		track.codecCapability.ClockRate = samplerate
		track.codecCapability.Channels = channelLayout

		return nil
	}
}
