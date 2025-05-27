package client

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/pion/interceptor/pkg/cc"

	"github.com/harshabose/simple_webrtc_comm/mediasource/pkg"
	"github.com/harshabose/simple_webrtc_comm/transcode/pkg"
)

type bwController struct {
	estimator cc.BandwidthEstimator
	interval  time.Duration
	subs      map[*mediasource.Track]transcode.UpdateBitrateCallBack
	ctx       context.Context
}

func createBWController(ctx context.Context) *bwController {
	return &bwController{
		subs:      make(map[*mediasource.Track]transcode.UpdateBitrateCallBack),
		estimator: nil,
		ctx:       ctx,
	}
}

func (bwc *bwController) Start() {
	if bwc.estimator == nil {
		return
	}
	go bwc.loop()
	fmt.Println("bw estimator started")
}

func (bwc *bwController) Subscribe(track *mediasource.Track) error {
	if _, exists := bwc.subs[track]; exists {
		return errors.New("bwc track subscriber already exists")
	}

	canGetUpdateBitrateCallBack, err := mediasource.WithBitrateControl(track)
	if err != nil {
		return err
	}

	bwc.subs[track] = canGetUpdateBitrateCallBack.OnUpdateBitrate()
	fmt.Println("new subscriber added with label:", track.GetTrack().ID())

	return nil
}

func (bwc *bwController) loop() {
	ticker := time.NewTicker(bwc.interval)
	defer ticker.Stop()

	for {
		select {
		case <-bwc.ctx.Done():
			return
		case <-ticker.C:
			var totalPriority = mediasource.Level0
			if len(bwc.subs) == 0 {
				continue
			}

			for track := range bwc.subs {
				totalPriority += track.GetPriority()
			}

			totalBitrate, err := bwc.getBitrate()
			if err != nil {
				continue
			}

			for track, callBack := range bwc.subs {
				if track.GetPriority() == mediasource.Level0 {
					continue
				}
				bitrate := int64(float64(totalBitrate) * float64(track.GetPriority()) / float64(totalPriority))
				go bwc.send(callBack, bitrate)
			}
		}
	}
}

func (bwc *bwController) send(callBack transcode.UpdateBitrateCallBack, bitrate int64) {
	ctx, cancel := context.WithTimeout(bwc.ctx, bwc.interval/time.Duration(len(bwc.subs)))
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- callBack(bitrate)
	}()

	select {
	case err := <-done:
		if err != nil {
			fmt.Printf("bitrate update callback failed: %v\n", err)
		}
	case <-ctx.Done():
		fmt.Println("context expired for bitrate update callback...")
	}
}

func (bwc *bwController) getBitrate() (int, error) {
	if bwc.estimator == nil {
		return 0, errors.New("estimator is nil")
	}
	return bwc.estimator.GetTargetBitrate(), nil
}
