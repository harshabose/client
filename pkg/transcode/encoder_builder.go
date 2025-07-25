//go:build cgo_enabled

package transcode

import (
	"context"

	"github.com/asticode/go-astiav"

	"github.com/harshabose/tools/pkg/buffer"
)

type GeneralEncoderBuilder struct {
	codecID    astiav.CodecID
	bufferSize int
	pool       buffer.Pool[*astiav.Packet]
	settings   codecSettings
	producer   CanProduceMediaFrame
}

func NewEncoderBuilder(codecID astiav.CodecID, settings codecSettings, producer CanProduceMediaFrame, bufferSize int, pool buffer.Pool[*astiav.Packet]) *GeneralEncoderBuilder {
	return &GeneralEncoderBuilder{
		bufferSize: bufferSize,
		pool:       pool,
		codecID:    codecID,
		settings:   settings,
		producer:   producer,
	}
}

func (b *GeneralEncoderBuilder) UpdateBitrate(bps int64) error {
	s, ok := b.settings.(CanUpdateBitrate)
	if !ok {
		return ErrorInterfaceMismatch
	}

	return s.UpdateBitrate(bps)
}

func (b *GeneralEncoderBuilder) BuildWithProducer(ctx context.Context, producer CanProduceMediaFrame) (Encoder, error) {
	b.producer = producer
	return b.Build(ctx)
}

func (b *GeneralEncoderBuilder) Build(ctx context.Context) (Encoder, error) {
	codec := astiav.FindEncoder(b.codecID)
	if codec == nil {
		return nil, ErrorNoCodecFound
	}

	ctx2, cancel := context.WithCancel(ctx)
	encoder := &GeneralEncoder{
		producer:   b.producer,
		codec:      codec,
		codecFlags: astiav.NewDictionary(),
		ctx:        ctx2,
		cancel:     cancel,
	}

	encoder.encoderContext = astiav.AllocCodecContext(codec)
	if encoder.encoderContext == nil {
		return nil, ErrorAllocateCodecContext
	}

	canDescribeMediaFrame, ok := encoder.producer.(CanDescribeMediaFrame)
	if !ok {
		return nil, ErrorInterfaceMismatch
	}

	if canDescribeMediaFrame.MediaType() == astiav.MediaTypeAudio {
		withAudioSetEncoderContextParameters(canDescribeMediaFrame, encoder.encoderContext)
	}
	if canDescribeMediaFrame.MediaType() == astiav.MediaTypeVideo {
		withVideoSetEncoderContextParameter(canDescribeMediaFrame, encoder.encoderContext)
	}

	if err := encoder.SetEncoderCodecSettings(b.settings); err != nil {
		return nil, err
	}

	if err := WithEncoderBufferSize(b.bufferSize, b.pool)(encoder); err != nil {
		return nil, err
	}

	encoder.encoderContext.SetFlags(astiav.NewCodecContextFlags(astiav.CodecContextFlagGlobalHeader))

	if err := encoder.encoderContext.Open(encoder.codec, encoder.codecFlags); err != nil {
		return nil, err
	}

	encoder.findParameterSets(encoder.encoderContext.ExtraData())

	return encoder, nil
}

func (b *GeneralEncoderBuilder) GetCurrentBitrate() (int64, error) {
	g, ok := b.settings.(CanGetCurrentBitrate)
	if !ok {
		return 0, ErrorInterfaceMismatch
	}

	return g.GetCurrentBitrate()
}
