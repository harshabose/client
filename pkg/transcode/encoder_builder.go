//go:build cgo_enabled

package transcode

import (
	"context"

	"github.com/asticode/go-astiav"
)

type GeneralEncoderBuilder struct {
	codecID  astiav.CodecID
	producer CanProduceMediaFrame
	options  []EncoderOption

	settings codecSettings
}

func NewEncoderBuilder(codecID astiav.CodecID, settings codecSettings, producer CanProduceMediaFrame, options ...EncoderOption) *GeneralEncoderBuilder {
	return &GeneralEncoderBuilder{
		codecID:  codecID,
		producer: producer,
		options:  options,
		settings: settings,
	}
}

func (b *GeneralEncoderBuilder) AdaptBitrate(bps int64) error {
	s, ok := b.settings.(CanAdaptBitrate)
	if !ok {
		return ErrorInterfaceMismatch
	}

	return s.AdaptBitrate(bps)
}

func (b *GeneralEncoderBuilder) BuildWithProducer(ctx context.Context, producer CanProduceMediaFrame) (Encoder, error) {
	b.producer = producer
	return b.Build(ctx)
}

func (b *GeneralEncoderBuilder) Build(ctx context.Context) (Encoder, error) {
	return CreateGeneralEncoder(ctx, b.codecID, b.producer, append(b.options, WithCodecSettings(b.settings))...)
}

func (b *GeneralEncoderBuilder) GetCurrentBitrate() (int64, error) {
	g, ok := b.settings.(CanGetCurrentBitrate)
	if !ok {
		return 0, ErrorInterfaceMismatch
	}

	return g.GetCurrentBitrate()
}
