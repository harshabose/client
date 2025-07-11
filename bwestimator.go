//go:build cgo_enabled

package client

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/pion/interceptor/pkg/cc"

	"github.com/harshabose/simple_webrtc_comm/client/pkg/mediasource"
	"github.com/harshabose/simple_webrtc_comm/client/pkg/transcode"
)

type subscriber struct {
	id       string // unique identifier
	priority mediasource.Priority
	callback transcode.UpdateBitrateCallBack
}

type BWEController struct {
	estimator cc.BandwidthEstimator
	interval  time.Duration
	subs      []subscriber
	mux       sync.RWMutex
	ctx       context.Context
}

func createBWController(ctx context.Context) *BWEController {
	return &BWEController{
		subs:      make([]subscriber, 0),
		estimator: nil,
		ctx:       ctx,
	}
}

func (bwc *BWEController) Start() {
	go bwc.loop()
}

func (bwc *BWEController) Subscribe(id string, priority mediasource.Priority, callback transcode.UpdateBitrateCallBack) error {
	bwc.mux.Lock()
	defer bwc.mux.Unlock()

	for _, sub := range bwc.subs {
		if sub.id == id {
			return errors.New("subscriber already exists")
		}
	}

	bwc.subs = append(bwc.subs, subscriber{
		id:       id,
		priority: priority,
		callback: callback,
	})

	return nil
}

// getSubscribers returns a copy of subscribers for safe iteration
func (bwc *BWEController) getSubscribers() []subscriber {
	bwc.mux.RLock()
	defer bwc.mux.RUnlock()

	// Return a copy to avoid holding the lock during iteration
	subs := make([]subscriber, len(bwc.subs))
	copy(subs, bwc.subs)
	return subs
}

// calculateTotalPriority calculates sum of all subscriber priorities
func (bwc *BWEController) calculateTotalPriority(subs []subscriber) mediasource.Priority {
	var totalPriority = mediasource.Level0

	for _, sub := range subs {
		totalPriority += sub.priority
	}

	return totalPriority
}

func (bwc *BWEController) loop() {
	ticker := time.NewTicker(bwc.interval)
	defer ticker.Stop()

	for {
		select {
		case <-bwc.ctx.Done():
			return
		case <-ticker.C:
			if bwc.estimator == nil {
				continue
			}

			subs := bwc.getSubscribers()
			if len(subs) == 0 {
				continue
			}

			totalPriority := bwc.calculateTotalPriority(subs)
			if totalPriority == mediasource.Level0 {
				continue // No active priorities
			}

			totalBitrate, err := bwc.getBitrate()
			if err != nil {
				continue
			}

			// Process each subscriber
			for _, sub := range subs {
				if sub.priority == mediasource.Level0 {
					continue
				}

				bitrate := int64(float64(totalBitrate) * float64(sub.priority) / float64(totalPriority))
				go bwc.sendBitrateUpdate(len(subs), sub.callback, bitrate)
			}
		}
	}
}

func (bwc *BWEController) sendBitrateUpdate(n int, callback transcode.UpdateBitrateCallBack, bitrate int64) {
	timeout := bwc.interval
	if n > 0 {
		timeout = bwc.interval / time.Duration(n)
	}

	ctx, cancel := context.WithTimeout(bwc.ctx, timeout)
	defer cancel()

	done := make(chan error, 1)

	go func() {
		done <- callback(bitrate)
	}()

	select {
	case err := <-done:
		if err != nil {
			fmt.Printf("bitrate update callback failed: %v\n", err)
			return
		}
	case <-ctx.Done():
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			fmt.Println("bitrate update callback timed out")
			return
		}
		fmt.Println("bitrate update callback cancelled")
	}
}

func (bwc *BWEController) getBitrate() (int, error) {
	if bwc.estimator == nil {
		return 0, errors.New("estimator is nil")
	}
	return bwc.estimator.GetTargetBitrate(), nil
}

func (bwc *BWEController) Unsubscribe(id string) error {
	bwc.mux.Lock()
	defer bwc.mux.Unlock()

	for i, sub := range bwc.subs {
		if sub.id == id {
			// Remove the subscriber by swapping with the last element
			bwc.subs[i] = bwc.subs[len(bwc.subs)-1]
			bwc.subs = bwc.subs[:len(bwc.subs)-1]
			return nil
		}
	}

	return errors.New("subscriber not found")
}
