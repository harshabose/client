package transcode

import (
	"github.com/harshabose/tools/pkg/buffer"
)

type CanSetBuffer[T any] interface {
	SetBuffer(buffer buffer.BufferWithGenerator[T])
}

type CanAddToFilterContent interface {
	AddToFilterContent(string)
}

type CanPauseUnPauseEncoder interface {
	PauseEncoding() error
	UnPauseEncoding() error
}

type CanGetParameterSets interface {
	GetParameterSets() (sps, pps []byte, err error)
}

type codecSettings interface {
	ForEach(func(string, string) error) error
}

type CanSetEncoderCodecSettings interface {
	SetEncoderCodecSettings(codecSettings) error
}

type CanUpdateBitrate interface {
	UpdateBitrate(int64) error
}

type CanGetCurrentBitrate interface {
	GetCurrentBitrate() (int64, error)
}

type UpdateBitrateCallBack func(bps int64) error

type CanGetUpdateBitrateCallBack interface {
	OnUpdateBitrate() UpdateBitrateCallBack
}
