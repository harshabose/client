//go:build cgo_enabled

package transcode

import (
	"context"

	"github.com/asticode/go-astiav"

	"github.com/harshabose/tools/pkg/buffer"
)

func withVideoSetDecoderContext(demuxer CanDescribeMediaPacket) DecoderOption {
	return func(decoder Decoder) error {
		consumer, ok := decoder.(CanSetMediaPacket)
		if !ok {
			return ErrorInterfaceMismatch
		}

		if err := consumer.SetCodec(demuxer); err != nil {
			return err
		}

		if err := consumer.FillContextContent(demuxer); err != nil {
			return err
		}

		consumer.SetFrameRate(demuxer)
		consumer.SetTimeBase(demuxer)
		return nil
	}
}

func withAudioSetDecoderContext(demuxer CanDescribeMediaPacket) DecoderOption {
	return func(decoder Decoder) error {
		consumer, ok := decoder.(CanSetMediaPacket)
		if !ok {
			return ErrorInterfaceMismatch
		}

		if err := consumer.SetCodec(demuxer); err != nil {
			return err
		}

		if err := consumer.FillContextContent(demuxer); err != nil {
			return err
		}

		consumer.SetTimeBase(demuxer)
		return nil
	}
}

func WithDecoderBuffer(ctx context.Context, size int, pool buffer.Pool[*astiav.Frame]) DecoderOption {
	return func(decoder Decoder) error {
		s, ok := decoder.(CanSetBuffer[*astiav.Frame])
		if !ok {
			return ErrorInterfaceMismatch
		}

		s.SetBuffer(buffer.NewChannelBufferWithGenerator(ctx, pool, uint(size), 1))
		return nil
	}
}
