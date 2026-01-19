//go:build cgo_enabled

package transcode

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/asticode/go-astiav"

	"github.com/harshabose/tools/pkg/cond"
)

var (
	ErrUpdateFilterNotReady = errors.New("update filter not in ready state")
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

	cond *cond.ContextCond
	ctx  context.Context
}

func (f *UpdateFilter) MediaType() astiav.MediaType {
	f.cond.L.Lock()
	defer f.cond.L.Unlock()

	return f.filter.(CanDescribeMediaFrame).MediaType()
}

func (f *UpdateFilter) FrameRate() astiav.Rational {
	f.cond.L.Lock()
	defer f.cond.L.Unlock()

	return f.filter.(CanDescribeMediaFrame).FrameRate()
}

func (f *UpdateFilter) TimeBase() astiav.Rational {
	f.cond.L.Lock()
	defer f.cond.L.Unlock()

	return f.filter.(CanDescribeMediaFrame).TimeBase()
}

func (f *UpdateFilter) Height() int {
	f.cond.L.Lock()
	defer f.cond.L.Unlock()

	return f.filter.(CanDescribeMediaFrame).Height()
}

func (f *UpdateFilter) Width() int {
	f.cond.L.Lock()
	defer f.cond.L.Unlock()

	return f.filter.(CanDescribeMediaFrame).Width()
}

func (f *UpdateFilter) PixelFormat() astiav.PixelFormat {
	f.cond.L.Lock()
	defer f.cond.L.Unlock()

	return f.filter.(CanDescribeMediaFrame).PixelFormat()
}

func (f *UpdateFilter) SampleAspectRatio() astiav.Rational {
	f.cond.L.Lock()
	defer f.cond.L.Unlock()

	return f.filter.(CanDescribeMediaFrame).SampleAspectRatio()
}

func (f *UpdateFilter) ColorSpace() astiav.ColorSpace {
	f.cond.L.Lock()
	defer f.cond.L.Unlock()

	return f.filter.(CanDescribeMediaFrame).ColorSpace()
}

func (f *UpdateFilter) ColorRange() astiav.ColorRange {
	f.cond.L.Lock()
	defer f.cond.L.Unlock()

	return f.filter.(CanDescribeMediaFrame).ColorRange()
}

func (f *UpdateFilter) SampleRate() int {
	f.cond.L.Lock()
	defer f.cond.L.Unlock()

	return f.filter.(CanDescribeMediaFrame).SampleRate()
}

func (f *UpdateFilter) SampleFormat() astiav.SampleFormat {
	f.cond.L.Lock()
	defer f.cond.L.Unlock()

	return f.filter.(CanDescribeMediaFrame).SampleFormat()
}

func (f *UpdateFilter) ChannelLayout() astiav.ChannelLayout {
	f.cond.L.Lock()
	defer f.cond.L.Unlock()

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
		cond:    cond.NewContextCond(&sync.Mutex{}),
		ctx:     ctx,
	}

	return filter, nil
}

func (f *UpdateFilter) Start() {
	f.cond.L.Lock()
	defer f.cond.L.Unlock()

	f.filter.Start()
}

func (f *UpdateFilter) GetFrame(ctx context.Context) (*astiav.Frame, error) {
	f.cond.L.Lock()
	defer f.cond.L.Unlock()

	for {
		if f.filter == nil {
			if err := f.cond.Wait(ctx); err != nil {
				return nil, ErrUpdateFilterNotReady
			}

			continue
		}

		frame, err := f.filter.GetFrame(ctx)
		if err != nil {
			return nil, err
		}

		return frame, nil
	}
}

func (f *UpdateFilter) PutBack(frame *astiav.Frame) {
	f.cond.L.Lock()
	defer f.cond.L.Unlock()

	f.filter.PutBack(frame)
}

func (f *UpdateFilter) Close() {
	f.cond.L.Lock()
	defer f.cond.L.Unlock()

	f.filter.Close()
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

	f.cond.L.Lock()
	old := f.filter
	f.filter = nf
	f.cond.L.Unlock()

	f.cond.Broadcast()

	if old != nil {
		old.Close()
	}

	return nil
}

func (f *UpdateFilter) GetCurrentFPS() (uint8, error) {
	return f.builder.GetCurrentFPS()
}
