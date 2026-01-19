//go:build cgo_enabled

package transcode

import (
	"context"

	"github.com/asticode/go-astiav"
)

type CanSetDemuxerInputOption interface {
	SetInputOption(key, value string, flags astiav.DictionaryFlags) error
}

type CanSetDemuxerInputFormat interface {
	SetInputFormat(*astiav.InputFormat)
}

type CanDescribeFrameRate interface {
	FrameRate() astiav.Rational
}

type CanDescribeTimeBase interface {
	TimeBase() astiav.Rational
}

type CanSetTimeBase interface {
	SetTimeBase(CanDescribeTimeBase)
}

type CanDescribeMediaPacket interface {
	MediaType() astiav.MediaType
	CodecID() astiav.CodecID
	GetCodecParameters() *astiav.CodecParameters
	CanDescribeFrameRate
	CanDescribeTimeBase
}

type CanProduceMediaPacket interface {
	GetPacket(ctx context.Context) (*astiav.Packet, error)
	PutBack(*astiav.Packet)
}

type CanProduceMediaFrame interface {
	GetFrame(ctx context.Context) (*astiav.Frame, error)
	PutBack(*astiav.Frame)
}

type CanSetFrameRate interface {
	SetFrameRate(CanDescribeFrameRate)
}

type CanDescribeMediaVideoFrame interface {
	CanDescribeFrameRate
	CanDescribeTimeBase
	Height() int
	Width() int
	PixelFormat() astiav.PixelFormat
	SampleAspectRatio() astiav.Rational
	ColorSpace() astiav.ColorSpace
	ColorRange() astiav.ColorRange
}

type CanSetMediaVideoFrame interface {
	CanSetFrameRate
	CanSetTimeBase
	SetHeight(CanDescribeMediaVideoFrame)
	SetWidth(CanDescribeMediaVideoFrame)
	SetPixelFormat(CanDescribeMediaVideoFrame)
	SetSampleAspectRatio(CanDescribeMediaVideoFrame)
	SetColorSpace(CanDescribeMediaVideoFrame)
	SetColorRange(CanDescribeMediaVideoFrame)
}

type CanDescribeMediaFrame interface {
	MediaType() astiav.MediaType
	CanDescribeMediaVideoFrame
	CanDescribeMediaAudioFrame
}

type CanSetMediaAudioFrame interface {
	CanSetTimeBase
	SetSampleRate(CanDescribeMediaAudioFrame)
	SetSampleFormat(CanDescribeMediaAudioFrame)
	SetChannelLayout(CanDescribeMediaAudioFrame)
}

type CanDescribeMediaAudioFrame interface {
	CanDescribeTimeBase
	SampleRate() int
	SampleFormat() astiav.SampleFormat
	ChannelLayout() astiav.ChannelLayout
}

type CanSetMediaPacket interface {
	FillContextContent(CanDescribeMediaPacket) error
	SetCodec(CanDescribeMediaPacket) error
	CanSetFrameRate
	CanSetTimeBase
}

type Demuxer interface {
	Start()
	Close()
	CanProduceMediaPacket
}

type Decoder interface {
	Start()
	Close()
	CanProduceMediaFrame
}

type Filter interface {
	Start()
	Close()
	CanProduceMediaFrame
}

type Encoder interface {
	Start()
	Close()
	CanProduceMediaPacket
}
