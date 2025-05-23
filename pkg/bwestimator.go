package client

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/pion/interceptor/pkg/cc"

	"github.com/harshabose/simple_webrtc_comm/mediasource/pkg"
)

type bwController struct {
	estimator cc.BandwidthEstimator
	interval  time.Duration
	subs      map[*mediasource.Track]chan int64
	ctx       context.Context
}

func createBWController(ctx context.Context) *bwController {
	return &bwController{
		subs:      make(map[*mediasource.Track]chan int64),
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
	channel := make(chan int64)

	if _, exists := bwc.subs[track]; exists {
		return errors.New("bwc track subscriber already exists")
	}

	if err := mediasource.WithBitrateControl(channel)(track); err != nil {
		return err
	}

	bwc.subs[track] = channel
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

			for track, channel := range bwc.subs {
				if track.GetPriority() == mediasource.Level0 {
					continue
				}
				bitrate := int64(float64(totalBitrate) * float64(track.GetPriority()) / float64(totalPriority))
				bwc.send(channel, bitrate/1000)
			}
		}
	}
}

func (bwc *bwController) send(channel chan int64, bitrate int64) {
	_, cancel := context.WithTimeout(bwc.ctx, bwc.interval/time.Duration(len(bwc.subs)))
	defer cancel()

	select {
	// case <-ctx.Done():
	// 	return
	case channel <- bitrate:
	}
}

func (bwc *bwController) getBitrate() (int, error) {
	// ctx, cancel := context.WithTimeout(bwc.ctx, bwc.interval)
	// defer cancel()
	//
	// resultCh := make(chan int, 1)
	//
	// // Run GetTargetBitrate in a separate goroutine
	// go func() {
	// 	resultCh <- bwc.estimator.GetTargetBitrate()
	// }()
	//
	// // Wait for either the result or timeout
	// select {
	// case bitrate := <-resultCh:
	// 	bitrate = bitrate / 1000
	// 	return bitrate, nil
	// case <-ctx.Done():
	// 	return 0, ctx.Err()
	// }
	return bwc.estimator.GetTargetBitrate(), nil
}
