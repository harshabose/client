//go:build cgo_enabled

package transcode

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/asticode/go-astiav"

	"github.com/harshabose/tools/pkg/buffer"
)

type UpdateFilterConfig struct {
	InitialFPS     uint8
	MaxFPS, MinFPS uint8
	SwitchLimitBPS int64
}

func (c UpdateFilterConfig) validate() error {
	if c.MinFPS > c.MaxFPS {
		return fmt.Errorf("update filter config: minimum fps is higher than maximum fps")
	}

	return nil
}

type UpdateFilter struct {
	filter  Filter
	config  UpdateFilterConfig
	builder *GeneralFilterBuilder

	buffer buffer.BufferWithGenerator[*astiav.Frame]
	mux    sync.RWMutex

	// wg     sync.WaitGroup // TODO: ADD CLOSE METHODS ON TRANSCODER ELEMENTS
	ctx context.Context
	// cancel context.CancelFunc
}

func (f *UpdateFilter) MediaType() astiav.MediaType {
	f.mux.RLock()
	defer f.mux.RUnlock()

	return f.filter.(CanDescribeMediaFrame).MediaType()
}

func (f *UpdateFilter) FrameRate() astiav.Rational {
	f.mux.RLock()
	defer f.mux.RUnlock()

	return f.filter.(CanDescribeMediaFrame).FrameRate()
}

func (f *UpdateFilter) TimeBase() astiav.Rational {
	f.mux.RLock()
	defer f.mux.RUnlock()

	return f.filter.(CanDescribeMediaFrame).TimeBase()
}

func (f *UpdateFilter) Height() int {
	f.mux.RLock()
	defer f.mux.RUnlock()

	return f.filter.(CanDescribeMediaFrame).Height()
}

func (f *UpdateFilter) Width() int {
	f.mux.RLock()
	defer f.mux.RUnlock()

	return f.filter.(CanDescribeMediaFrame).Width()
}

func (f *UpdateFilter) PixelFormat() astiav.PixelFormat {
	f.mux.RLock()
	defer f.mux.RUnlock()

	return f.filter.(CanDescribeMediaFrame).PixelFormat()
}

func (f *UpdateFilter) SampleAspectRatio() astiav.Rational {
	f.mux.RLock()
	defer f.mux.RUnlock()

	return f.filter.(CanDescribeMediaFrame).SampleAspectRatio()
}

func (f *UpdateFilter) ColorSpace() astiav.ColorSpace {
	f.mux.RLock()
	defer f.mux.RUnlock()

	return f.filter.(CanDescribeMediaFrame).ColorSpace()
}

func (f *UpdateFilter) ColorRange() astiav.ColorRange {
	f.mux.RLock()
	defer f.mux.RUnlock()

	return f.filter.(CanDescribeMediaFrame).ColorRange()
}

func (f *UpdateFilter) SampleRate() int {
	f.mux.RLock()
	defer f.mux.RUnlock()

	return f.filter.(CanDescribeMediaFrame).SampleRate()
}

func (f *UpdateFilter) SampleFormat() astiav.SampleFormat {
	f.mux.RLock()
	defer f.mux.RUnlock()

	return f.filter.(CanDescribeMediaFrame).SampleFormat()
}

func (f *UpdateFilter) ChannelLayout() astiav.ChannelLayout {
	f.mux.RLock()
	defer f.mux.RUnlock()

	return f.filter.(CanDescribeMediaFrame).ChannelLayout()
}

func NewUpdateFilter(ctx context.Context, config UpdateFilterConfig, builder *GeneralFilterBuilder, fps uint8) (*UpdateFilter, error) {
	if err := config.validate(); err != nil {
		return nil, err
	}

	if err := builder.AdaptFPS(fps); err != nil {
		return nil, err
	}

	f, err := builder.Build(ctx)
	if err != nil {
		return nil, err
	}

	filter := &UpdateFilter{
		filter:  f,
		config:  config,
		builder: builder,
		buffer:  buffer.NewChannelBufferWithGenerator(ctx, buffer.CreateFramePool(), 30, 1), // TODO: CHANGE THIS ASAP
		mux:     sync.RWMutex{},
		ctx:     ctx,
	}

	go filter.loop()

	return filter, nil
}

func (f *UpdateFilter) Ctx() context.Context {
	return f.ctx
}

func (f *UpdateFilter) Start() {
	f.mux.RLock()
	defer f.mux.RUnlock()

	f.filter.Start()
}

func (f *UpdateFilter) GetFrame(ctx context.Context) (*astiav.Frame, error) {
	return f.buffer.Pop(ctx)
}

func (f *UpdateFilter) PutBack(frame *astiav.Frame) {
	f.mux.RLock()
	defer f.mux.RUnlock()

	f.filter.PutBack(frame)
}

func (f *UpdateFilter) Stop() {
	f.mux.Lock()
	defer f.mux.Unlock()

	f.filter.Stop()
}

func (f *UpdateFilter) AdaptBitrate(bps int64) error {
	fps := f.config.MaxFPS
	if bps < f.config.SwitchLimitBPS {
		fps = f.config.MinFPS
	}

	if err := f.builder.AdaptFPS(fps); err != nil {
		return err
	}

	nf, err := f.builder.Build(f.ctx)
	if err != nil {
		return err
	}

	nf.Start()

	f.mux.Lock()
	old := f.filter
	f.filter = nf
	f.mux.Unlock()

	if old != nil {
		old.Stop()
	}

	return nil
}

func (f *UpdateFilter) GetCurrentFPS() (uint8, error) {
	return f.builder.GetCurrentFPS()
}

func (f *UpdateFilter) getFrame() (*astiav.Frame, error) {
	f.mux.RLock()
	defer f.mux.RUnlock()

	if f.filter != nil {
		ctx2, cancel2 := context.WithTimeout(f.ctx, 100*time.Millisecond)
		defer cancel2()

		frame, err := f.filter.GetFrame(ctx2)
		if err != nil {
			return nil, err
		}

		return frame, nil
	}

	return nil, errors.New("filter is nil")
}

func (f *UpdateFilter) pushFrame(frame *astiav.Frame) error {
	if frame == nil {
		return nil
	}

	ctx2, cancel2 := context.WithTimeout(f.ctx, 100*time.Millisecond)
	defer cancel2()

	return f.buffer.Push(ctx2, frame)
}

func (f *UpdateFilter) loop() {
	for {
		select {
		case <-f.ctx.Done():
			return
		default:
			frame, err := f.getFrame()
			if err != nil {
				fmt.Printf("error getting frame from update filter; err=%v\n", err)
				continue
			}

			if err := f.pushFrame(frame); err != nil {
				fmt.Printf("error pushing frame in update filter; err=%v\n", err)
				continue
			}
		}
	}
}
