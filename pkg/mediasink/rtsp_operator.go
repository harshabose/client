package mediasink

import (
	"context"
	"time"

	"github.com/harshabose/mediapipe"
	"github.com/harshabose/mediapipe/pkg/rtpgenerator"
	"github.com/harshabose/mediapipe/pkg/rtsp"
	"github.com/pion/rtp"
	"github.com/pion/webrtc/v4"
)

func RTSPSink(config *rtsp.ClientConfig) func(context.Context, *webrtc.TrackRemote) error {
	return func(ctx context.Context, remote *webrtc.TrackRemote) error {
		client, err := rtsp.NewClient(ctx, config, nil, rtsp.WithOptionsFromRemote(remote))
		if err != nil {
			return err
		}

		client.Start()
		time.Sleep(5 * time.Second)

		r := mediapipe.NewIdentityAnyReader(rtpgenerator.NewPionRTPGenerator(remote))
		w := mediapipe.NewIdentityAnyWriter[*rtp.Packet](client)

		mediapipe.NewAnyPipe(ctx, r, w).Start()
		return nil
	}
}
