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
	ErrUpdateEncoderNotReady = errors.New("update encoder not in ready state")
)

type UpdateEncoderConfig struct {
	MaxBitrate, MinBitrate     int64
	MinBitrateChangePercentage float64
}

func (c UpdateEncoderConfig) validate() error {
	if c.MinBitrate > c.MaxBitrate {
		return fmt.Errorf("update encoder config: minimum bitrate is higher than maximum bitrate ")
	}

	return nil
}

type UpdateEncoder struct {
	encoder Encoder
	config  UpdateEncoderConfig
	builder *GeneralEncoderBuilder

	cond *cond.ContextCond
	ctx  context.Context
}

func NewUpdateEncoder(ctx context.Context, config UpdateEncoderConfig, builder *GeneralEncoderBuilder) (*UpdateEncoder, error) {
	updater := &UpdateEncoder{
		config:  config,
		builder: builder,
		cond:    cond.NewContextCond(&sync.Mutex{}),
		ctx:     ctx,
	}

	if err := config.validate(); err != nil {
		return nil, err
	}

	encoder, err := builder.Build(ctx)
	if err != nil {
		return nil, err
	}

	updater.encoder = encoder

	return updater, nil
}

func (u *UpdateEncoder) Start() {
	u.cond.L.Lock()
	defer u.cond.L.Unlock()

	u.encoder.Start()
}

func (u *UpdateEncoder) GetPacket(ctx context.Context) (*astiav.Packet, error) {
	u.cond.L.Lock()
	defer u.cond.L.Unlock()

	for {
		if u.encoder == nil {
			if err := u.cond.Wait(ctx); err != nil {
				return nil, ErrUpdateEncoderNotReady
			}

			continue
		}
		p, err := u.encoder.GetPacket(ctx)
		if err != nil {
			return nil, err
		}

		return p, nil
	}
}

func (u *UpdateEncoder) PutBack(packet *astiav.Packet) {
	u.cond.L.Lock()
	defer u.cond.L.Unlock()

	u.encoder.PutBack(packet)
}

func (u *UpdateEncoder) Close() {
	u.cond.L.Lock()
	defer u.cond.L.Unlock()

	u.encoder.Close()
}

func (u *UpdateEncoder) AdaptBitrate(bps int64) error {
	bps = u.cutoff(bps)

	g, ok := u.encoder.(CanGetCurrentBitrate)
	if !ok {
		return ErrorInterfaceMismatch
	}

	current, err := g.GetCurrentBitrate()
	if err != nil {
		return err
	}

	_, change := calculateBitrateChange(current, bps)
	if change < u.config.MinBitrateChangePercentage {
		return nil
	}

	if err := u.builder.AdaptBitrate(bps); err != nil {
		return err
	}

	newEncoder, err := u.builder.Build(u.ctx)
	if err != nil {
		return err
	}

	newEncoder.Start()

	u.cond.L.Lock()
	oldEncoder := u.encoder
	u.encoder = newEncoder
	u.cond.L.Unlock()

	u.cond.Broadcast()

	if oldEncoder != nil {
		oldEncoder.Close()
	}

	return nil
}

func (u *UpdateEncoder) cutoff(bps int64) int64 {
	if bps > u.config.MaxBitrate {
		bps = u.config.MaxBitrate
	}

	if bps < u.config.MinBitrate {
		bps = u.config.MinBitrate
	}

	return bps
}

func (u *UpdateEncoder) GetParameterSets() (sps []byte, pps []byte, err error) {
	p, ok := u.encoder.(CanGetParameterSets)
	if !ok {
		return nil, nil, ErrorInterfaceMismatch
	}

	return p.GetParameterSets()
}

func calculateBitrateChange(currentBps, newBps int64) (absoluteChange int64, percentageChange float64) {
	absoluteChange = newBps - currentBps
	if absoluteChange < 0 {
		absoluteChange = -absoluteChange
	}

	if currentBps > 0 {
		percentageChange = (float64(absoluteChange) / float64(currentBps)) * 100
	}

	return absoluteChange, percentageChange
}
