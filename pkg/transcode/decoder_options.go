//go:build cgo_enabled

package transcode

import (
	"github.com/asticode/go-astiav"

	"github.com/harshabose/tools/pkg/buffer"
)

type DecoderOption = func(decoder Decoder) error

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

func WithDecoderBuffer(size int, pool buffer.Pool[*astiav.Frame]) DecoderOption {
	return func(decoder Decoder) error {
		s, ok := decoder.(CanSetBuffer[*astiav.Frame])
		if !ok {
			return ErrorInterfaceMismatch
		}
		s.SetBuffer(buffer.CreateChannelBuffer(decoder.Ctx(), size, pool))
		return nil
	}
}
