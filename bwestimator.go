package client

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"sync"
	"time"

	"github.com/pion/interceptor/pkg/cc"

	"github.com/harshabose/simple_webrtc_comm/client/pkg/mediasource"
)

type UpdateBitrateCallBack = func(bps int64) error

type subscriber struct {
	id       string // unique identifier
	priority mediasource.Priority
	callback UpdateBitrateCallBack
}

type BWEController struct {
	estimator cc.BandwidthEstimator
	interval  time.Duration
	subs      map[string]*subscriber
	once      sync.Once
	mux       sync.RWMutex
	wg        sync.WaitGroup
	ctx       context.Context
	cancel    context.CancelFunc
}

func createBWController(ctx context.Context) *BWEController {
	ctx2, cancel2 := context.WithCancel(ctx)

	return &BWEController{
		subs:      make(map[string]*subscriber),
		estimator: nil,
		ctx:       ctx2,
		cancel:    cancel2,
	}
}

func (bwc *BWEController) set(estimator cc.BandwidthEstimator, interval time.Duration) {
	bwc.mux.Lock()
	defer bwc.mux.Unlock()

	bwc.estimator = estimator
	bwc.interval = interval
}

func (bwc *BWEController) get() cc.BandwidthEstimator {
	bwc.mux.RLock()
	defer bwc.mux.RUnlock()

	return bwc.estimator
}

func (bwc *BWEController) Start() {
	go bwc.loop()
}

func (bwc *BWEController) Subscribe(id string, priority mediasource.Priority, callback UpdateBitrateCallBack) error {
	bwc.mux.Lock()
	defer bwc.mux.Unlock()

	if _, exists := bwc.subs[id]; exists {
		return errors.New("subscriber already exists")
	}

	bwc.subs[id] = &subscriber{
		id:       id,
		priority: priority,
		callback: callback,
	}

	return nil
}

func (bwc *BWEController) subscribers() iter.Seq2[string, *subscriber] {
	return func(yield func(string, *subscriber) bool) {
		bwc.mux.RLock()
		defer bwc.mux.RUnlock()

		for id, sub := range bwc.subs {
			if !yield(id, sub) {
				return
			}
		}
	}
}

func (bwc *BWEController) calculateTotalPriority() mediasource.Priority {
	var totalPriority = mediasource.Level0

	for _, sub := range bwc.subscribers() {
		totalPriority += sub.priority
	}

	return totalPriority
}

func (bwc *BWEController) loop() {
	bwc.wg.Add(1)
	defer bwc.wg.Done()

	ticker := time.NewTicker(bwc.interval)
	defer ticker.Stop()

	for {
		select {
		case <-bwc.ctx.Done():
			return
		case <-ticker.C:
			if bwc.get() == nil {
				continue
			}

			totalPriority := bwc.calculateTotalPriority()
			if totalPriority == mediasource.Level0 {
				continue // No active priorities
			}

			totalBitrate, err := bwc.getBitrate()
			if err != nil {
				continue
			}

			for _, sub := range bwc.subscribers() {
				if sub.priority == mediasource.Level0 {
					continue
				}
				bitrate := int64(float64(totalBitrate) * float64(sub.priority) / float64(totalPriority))
				go bwc.sendBitrateUpdate(sub.id, sub.callback, bitrate)
			}
		}
	}
}

func (bwc *BWEController) sendBitrateUpdate(id string, callback UpdateBitrateCallBack, bitrate int64) {
	done := make(chan error, 1)

	go func() {
		done <- callback(bitrate)
	}()

	select {
	case err := <-done:
		if err != nil {
			fmt.Printf("bitrate update callback (id=%s) failed: %v. Unsubscribing...\n", id, err)
			bwc.Unsubscribe(id)
		}
	case <-bwc.ctx.Done():
		return
	}
}

func (bwc *BWEController) getBitrate() (int, error) {
	if bwc.get() == nil {
		return 0, errors.New("estimator is nil")
	}
	return bwc.estimator.GetTargetBitrate(), nil
}

func (bwc *BWEController) Unsubscribe(id string) {
	bwc.mux.Lock()
	defer bwc.mux.Unlock()

	if _, exists := bwc.subs[id]; !exists {
		return
	}

	delete(bwc.subs, id)
}

func (bwc *BWEController) Close() {
	bwc.once.Do(func() {
		if bwc.cancel != nil {
			bwc.cancel()
		}

		bwc.wg.Wait()

		bwc.mux.Lock()
		defer bwc.mux.Unlock()

		// if bwc.estimator != nil {
		// 	if err := bwc.estimator.Close(); err != nil {
		// 		return
		// 	}
		// }

		// NOTE: CLOSED BY PC
		// if err := bwc.estimator.Close(); err != nil {
		// 	merr = multierr.Append(merr, err)
		// }

		bwc.subs = nil
	})
}
