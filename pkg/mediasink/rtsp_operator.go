package mediasink

import (
	"context"
	"time"

	"github.com/pion/rtp"
	"github.com/pion/webrtc/v4"

	"github.com/harshabose/mediapipe"
	"github.com/harshabose/mediapipe/pkg/duplexers"
	"github.com/harshabose/mediapipe/pkg/generators"
)

func RTSPSink(config *duplexers.RTSPClientConfig) func(context.Context, *webrtc.TrackRemote) error {
	return func(ctx context.Context, remote *webrtc.TrackRemote) error {
		client, err := duplexers.NewRTSPClient(ctx, config, nil, duplexers.WithOptionsFromRemote(remote))
		if err != nil {
			return err
		}

		client.Start()
		time.Sleep(5 * time.Second)

		r := mediapipe.NewIdentityAnyReader(generators.NewPionRTPGenerator(remote))
		w := mediapipe.NewIdentityAnyWriter[*rtp.Packet](client)

		mediapipe.NewAnyPipe(ctx, r, w).Start()
		return nil
	}
}
