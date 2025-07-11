package mediasource

import (
	"github.com/pion/webrtc/v4"
)

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

type TrackOption = func(*Track) error

func WithH264Track(clockrate uint32, packetisationMode PacketisationMode, profileLevel ProfileLevel) TrackOption {
	return func(track *Track) error {

		track.rtcCapability.MimeType = webrtc.MimeTypeH264
		track.rtcCapability.ClockRate = clockrate
		track.rtcCapability.Channels = 0
		// NOTE: NOT NEEDED
		// track.rtcCapability.SDPFmtpLine = track.rtcCapability.SDPFmtpLine + fmt.Sprintf("level-asymmetry-allowed=1;packetization-mode=%d;profile-level-id=%s", packetisationMode, profileLevel)

		return nil
	}
}

func WithVP8Track(clockrate uint32) TrackOption {
	return func(track *Track) error {
		track.rtcCapability.MimeType = webrtc.MimeTypeVP8
		track.rtcCapability.ClockRate = clockrate
		track.rtcCapability.Channels = 0

		return nil
	}
}

type StereoType uint8

const (
	StereoMono = StereoType(0)
	StereoDual = StereoType(1)
)

func WithOpusTrack(samplerate uint32, channelLayout uint16, stereo StereoType) TrackOption {
	return func(track *Track) error {

		track.rtcCapability.MimeType = webrtc.MimeTypeOpus
		track.rtcCapability.ClockRate = samplerate
		track.rtcCapability.Channels = channelLayout
		// NOTE: NOT NEEDED
		// track.rtcCapability.SDPFmtpLine = track.rtcCapability.SDPFmtpLine + fmt.Sprintf("minptime=10;useinbandfec=1;stereo=%d", stereo)

		return nil
	}
}

func WithPriority(level Priority) TrackOption {
	return func(track *Track) error {
		track.priority = level
		return nil
	}
}

type Priority uint8

const (
	Level0 Priority = 0
	Level1 Priority = 1
	Level2 Priority = 2
	Level3 Priority = 3
	Level4 Priority = 4
	Level5 Priority = 5
)
