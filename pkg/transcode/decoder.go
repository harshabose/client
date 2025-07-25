//go:build cgo_enabled

package transcode

import (
	"context"
	"errors"
	"time"

	"github.com/asticode/go-astiav"

	"github.com/harshabose/tools/pkg/buffer"
)

type GeneralDecoder struct {
	demuxer        CanProduceMediaPacket
	decoderContext *astiav.CodecContext
	codec          *astiav.Codec
	buffer         buffer.BufferWithGenerator[*astiav.Frame]
	ctx            context.Context
	cancel         context.CancelFunc
}

func CreateGeneralDecoder(ctx context.Context, canProduceMediaType CanProduceMediaPacket, options ...DecoderOption) (*GeneralDecoder, error) {
	var (
		err           error
		contextOption DecoderOption
		decoder       *GeneralDecoder
	)

	ctx2, cancel := context.WithCancel(ctx)
	decoder = &GeneralDecoder{
		demuxer: canProduceMediaType,
		ctx:     ctx2,
		cancel:  cancel,
	}

	canDescribeMediaPacket, ok := canProduceMediaType.(CanDescribeMediaPacket)
	if !ok {
		return nil, ErrorInterfaceMismatch
	}

	if canDescribeMediaPacket.MediaType() == astiav.MediaTypeVideo {
		contextOption = withVideoSetDecoderContext(canDescribeMediaPacket)
	}
	if canDescribeMediaPacket.MediaType() == astiav.MediaTypeAudio {
		contextOption = withAudioSetDecoderContext(canDescribeMediaPacket)
	}

	options = append([]DecoderOption{contextOption}, options...)

	for _, option := range options {
		if err = option(decoder); err != nil {
			return nil, err
		}
	}

	if decoder.buffer == nil {
		decoder.buffer = buffer.CreateChannelBuffer(ctx, 256, buffer.CreateFramePool())
	}

	if err = decoder.decoderContext.Open(decoder.codec, nil); err != nil {
		return nil, err
	}

	return decoder, nil
}

func (decoder *GeneralDecoder) Ctx() context.Context {
	return decoder.ctx
}

func (decoder *GeneralDecoder) Start() {
	go decoder.loop()
}

func (decoder *GeneralDecoder) Stop() {
	decoder.cancel()
}

func (decoder *GeneralDecoder) loop() {
	defer decoder.close()

loop1:
	for {
		select {
		case <-decoder.ctx.Done():
			return
		default:
			packet, err := decoder.getPacket()
			if err != nil {
				// fmt.Println("unable to get packet from demuxer; err:", err.Error())
				continue
			}
			if err := decoder.decoderContext.SendPacket(packet); err != nil {
				decoder.demuxer.PutBack(packet)
				if !errors.Is(err, astiav.ErrEagain) {
					continue loop1
				}
			}
		loop2:
			for {
				frame := decoder.buffer.Generate()
				if err := decoder.decoderContext.ReceiveFrame(frame); err != nil {
					decoder.buffer.PutBack(frame)
					break loop2
				}

				frame.SetPictureType(astiav.PictureTypeNone)

				if err := decoder.pushFrame(frame); err != nil {
					decoder.buffer.PutBack(frame)
					continue loop2
				}
			}
			decoder.demuxer.PutBack(packet)
		}
	}
}

func (decoder *GeneralDecoder) pushFrame(frame *astiav.Frame) error {
	ctx, cancel := context.WithTimeout(decoder.ctx, 50*time.Millisecond)
	defer cancel()

	return decoder.buffer.Push(ctx, frame)
}

func (decoder *GeneralDecoder) getPacket() (*astiav.Packet, error) {
	ctx, cancel := context.WithTimeout(decoder.ctx, 50*time.Millisecond)
	defer cancel()

	return decoder.demuxer.GetPacket(ctx)
}

func (decoder *GeneralDecoder) GetFrame(ctx context.Context) (*astiav.Frame, error) {
	return decoder.buffer.Pop(ctx)
}

func (decoder *GeneralDecoder) PutBack(frame *astiav.Frame) {
	decoder.buffer.PutBack(frame)
}

func (decoder *GeneralDecoder) close() {
	if decoder.decoderContext != nil {
		decoder.decoderContext.Free()
	}
}

func (decoder *GeneralDecoder) SetBuffer(buffer buffer.BufferWithGenerator[*astiav.Frame]) {
	decoder.buffer = buffer
}

func (decoder *GeneralDecoder) SetCodec(producer CanDescribeMediaPacket) error {
	if decoder.codec = astiav.FindDecoder(producer.CodecID()); decoder.codec == nil {
		return ErrorNoCodecFound
	}
	decoder.decoderContext = astiav.AllocCodecContext(decoder.codec)
	if decoder.decoderContext == nil {
		return ErrorAllocateCodecContext
	}

	return nil
}

func (decoder *GeneralDecoder) FillContextContent(producer CanDescribeMediaPacket) error {
	return producer.GetCodecParameters().ToCodecContext(decoder.decoderContext)
}

func (decoder *GeneralDecoder) SetFrameRate(producer CanDescribeFrameRate) {
	decoder.decoderContext.SetFramerate(producer.FrameRate())
}

func (decoder *GeneralDecoder) SetTimeBase(producer CanDescribeTimeBase) {
	decoder.decoderContext.SetTimeBase(producer.TimeBase())
}

// ### IMPLEMENTS CanDescribeMediaVideoFrame

func (decoder *GeneralDecoder) FrameRate() astiav.Rational {
	return decoder.decoderContext.Framerate()
}

func (decoder *GeneralDecoder) TimeBase() astiav.Rational {
	return decoder.decoderContext.TimeBase()
}

func (decoder *GeneralDecoder) Height() int {
	return decoder.decoderContext.Height()
}

func (decoder *GeneralDecoder) Width() int {
	return decoder.decoderContext.Width()
}

func (decoder *GeneralDecoder) PixelFormat() astiav.PixelFormat {
	return decoder.decoderContext.PixelFormat()
}

func (decoder *GeneralDecoder) SampleAspectRatio() astiav.Rational {
	return decoder.decoderContext.SampleAspectRatio()
}

func (decoder *GeneralDecoder) ColorSpace() astiav.ColorSpace {
	return decoder.decoderContext.ColorSpace()
}

func (decoder *GeneralDecoder) ColorRange() astiav.ColorRange {
	return decoder.decoderContext.ColorRange()
}

// ## CanDescribeMediaAudioFrame

func (decoder *GeneralDecoder) SampleRate() int {
	return decoder.decoderContext.SampleRate()
}

func (decoder *GeneralDecoder) SampleFormat() astiav.SampleFormat {
	return decoder.decoderContext.SampleFormat()
}

func (decoder *GeneralDecoder) ChannelLayout() astiav.ChannelLayout {
	return decoder.decoderContext.ChannelLayout()
}

// ## CanDescribeMediaFrame

func (decoder *GeneralDecoder) MediaType() astiav.MediaType {
	return decoder.decoderContext.MediaType()
}
