package client

import (
	"context"
	"errors"
	"time"

	"github.com/harshabose/simple_webrtc_comm/mediasource/pkg"
	"github.com/pion/interceptor/pkg/cc"
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
}

func (bwc *bwController) Subscribe(track *mediasource.Track) error {
	channel := make(chan int64)

	if err := mediasource.WithBitrateControl(channel)(track); err != nil {
		return err
	}

	if _, exists := bwc.subs[track]; exists {
		return errors.New("bwc track subscriber already exists")
	}
	bwc.subs[track] = channel

	return nil
}

func (bwc *bwController) loop() {
	var totalPriority mediasource.Priority
	ticker := time.NewTicker(bwc.interval)
	defer ticker.Stop()

	for {
		select {
		case <-bwc.ctx.Done():
			return
		case <-ticker.C:
			if len(bwc.subs) == 0 {
				continue
			}

			for track := range bwc.subs {
				totalPriority += track.GetPriority()
			}

			totalBitrate := bwc.estimator.GetTargetBitrate()

			for track, channel := range bwc.subs {
				if track.GetPriority() == mediasource.Level0 {
					continue
				}
				bitrate := int64(float64(totalBitrate) * float64(track.GetPriority()) / float64(totalPriority))
				bwc.send(channel, bitrate)
			}
		}
	}
}

func (bwc *bwController) send(channel chan int64, bitrate int64) {
	ctx, cancel := context.WithTimeout(bwc.ctx, bwc.interval/time.Duration(len(bwc.subs)))
	defer cancel()

	select {
	case <-ctx.Done():
		return
	case channel <- bitrate:
	}
}
