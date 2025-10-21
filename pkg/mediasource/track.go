package mediasource

import (
	"context"
	"errors"

	"github.com/pion/webrtc/v4"
	"github.com/pion/webrtc/v4/pkg/media"

	"github.com/harshabose/mediapipe/pkg/consumers"
)

// NO BUFFER IMPLEMENTATION

type Track struct {
	*track
	consumer consumers.CanConsumePionSamplePacket
	ctx      context.Context
}

func CreateTrack(ctx context.Context, label string, peerConnection *webrtc.PeerConnection, options ...TrackOption) (*Track, error) {
	track := &Track{
		track: &track{},
		ctx:   ctx,
	}

	for _, option := range options {
		if err := option(track.track); err != nil {
			return nil, err
		}
	}

	if track.codecCapability == nil {
		return nil, errors.New("no track capabilities given")
	}

	consumer, err := webrtc.NewTrackLocalStaticSample(*track.codecCapability, label, "webrtc")
	if err != nil {
		return nil, err
	}
	track.consumer = consumer

	if track.rtpSender, err = peerConnection.AddTrack(consumer); err != nil {
		return nil, err
	}

	go track.rtpSenderLoop()

	return track, nil
}

func (track *Track) GetPriority() Priority {
	return track.priority
}

func (track *Track) rtpSenderLoop() {
	// THIS IS NEEDED AS interceptors (pion) doesnt work
	for {
		select {
		case <-track.ctx.Done():
			return
		default:
			rtcpBuf := make([]byte, 1500)
			if _, _, err := track.rtpSender.Read(rtcpBuf); err != nil {
				// fmt.Println("error while reading rtcp packets")
			}
		}
	}
}

func (track *Track) WriteSample(sample media.Sample) error {
	if err := track.consumer.WriteSample(sample); err != nil {
		return err
	}

	return nil
}
