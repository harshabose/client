package mediasource

import (
	"context"
	"fmt"

	"github.com/pion/webrtc/v4"
	"github.com/pion/webrtc/v4/pkg/media"

	_ "github.com/harshabose/mediapipe"
)

// NO BUFFER IMPLEMENTATION

type Track struct {
	track         *webrtc.TrackLocalStaticSample
	rtcCapability *webrtc.RTPCodecCapability
	rtpSender     *webrtc.RTPSender
	priority      Priority
	ctx           context.Context
}

func CreateTrack(ctx context.Context, label string, peerConnection *webrtc.PeerConnection, options ...TrackOption) (*Track, error) {
	var err error
	track := &Track{ctx: ctx, rtcCapability: &webrtc.RTPCodecCapability{}}

	for _, option := range options {
		if err := option(track); err != nil {
			return nil, err
		}
	}

	if track.track, err = webrtc.NewTrackLocalStaticSample(*track.rtcCapability, label, "webrtc"); err != nil {
		return nil, err
	}

	if track.rtpSender, err = peerConnection.AddTrack(track.track); err != nil {
		return nil, err
	}

	return track, nil
}

func (track *Track) GetTrack() *webrtc.TrackLocalStaticSample {
	return track.track
}

func (track *Track) GetPriority() Priority {
	return track.priority
}

func (track *Track) rtpSenderLoop() {
	// THIS IS NEEDED AS interceptors (pion) doesnt work
	for {
		rtcpBuf := make([]byte, 1500)
		if _, _, err := track.rtpSender.Read(rtcpBuf); err != nil {
			fmt.Printf("error while reading rtcp packets")
		}
	}
}

func (track *Track) Consume(sample media.Sample) error {
	if err := track.track.WriteSample(sample); err != nil {
		fmt.Printf("error while writing samples to track (id: ); err; %v. Continuing...", err)
	}

	return nil
}
