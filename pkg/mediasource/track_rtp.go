package mediasource

import (
	"context"
	"errors"
	"fmt"

	"github.com/pion/rtp"
	"github.com/pion/webrtc/v4"

	"github.com/harshabose/mediapipe/pkg/consumers"
)

type track struct {
	codecCapability *webrtc.RTPCodecCapability
	rtpSender       *webrtc.RTPSender
	priority        Priority
}

type RTPTrack struct {
	*track
	consumer consumers.CanConsumePionRTPPackets
	ctx      context.Context
}

func CreateRTPTrack(ctx context.Context, label string, pc *webrtc.PeerConnection, options ...TrackOption) (*RTPTrack, error) {
	track := &RTPTrack{
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

	consumer, err := webrtc.NewTrackLocalStaticRTP(*track.codecCapability, label, "webrtc")
	if err != nil {
		return nil, err
	}
	track.consumer = consumer

	if track.rtpSender, err = pc.AddTrack(consumer); err != nil {
		return nil, err
	}

	go track.rtpSenderLoop()

	return track, nil
}

func (track *RTPTrack) GetPriority() Priority {
	return track.priority
}

func (track *RTPTrack) rtpSenderLoop() {
	// THIS IS NEEDED AS interceptors (pion) doesnt work
	for {
		select {
		case <-track.ctx.Done():
			return
		default:
			rtcpBuf := make([]byte, 1500)
			if _, _, err := track.rtpSender.Read(rtcpBuf); err != nil {
				// fmt.Println("error while reading rtcp packets")
				continue
			}
		}
	}
}

func (track *RTPTrack) WriteRTP(packet *rtp.Packet) error {
	if packet == nil {
		return nil
	}
	if err := track.consumer.WriteRTP(packet); err != nil {
		fmt.Printf("error while writing samples to track (id: ); err; %v. Continuing...", err)
	}

	return nil
}
