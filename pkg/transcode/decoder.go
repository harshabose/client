//go:build cgo_enabled

package transcode

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/asticode/go-astiav"

	"github.com/harshabose/tools/pkg/buffer"
)

type GeneralDecoder struct {
	demuxer        CanProduceMediaPacket
	decoderContext *astiav.CodecContext
	codec          *astiav.Codec
	buffer         buffer.BufferWithGenerator[*astiav.Frame]

	once   sync.Once
	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
}

func CreateGeneralDecoder(ctx context.Context, canProduceMediaType CanProduceMediaPacket, options ...DecoderOption) (*GeneralDecoder, error) {
	ctx2, cancel := context.WithCancel(ctx)

	decoder := &GeneralDecoder{
		demuxer: canProduceMediaType,
		ctx:     ctx2,
		cancel:  cancel,
	}

	canDescribeMediaPacket, ok := canProduceMediaType.(CanDescribeMediaPacket)
	if !ok {
		return nil, ErrorInterfaceMismatch
	}

	if canDescribeMediaPacket.MediaType() != astiav.MediaTypeVideo && canDescribeMediaPacket.MediaType() != astiav.MediaTypeAudio {
		return nil, ErrorUnsupportedMedia
	}

	o := withVideoSetDecoderContext(canDescribeMediaPacket)
	if canDescribeMediaPacket.MediaType() == astiav.MediaTypeAudio {
		o = withAudioSetDecoderContext(canDescribeMediaPacket)
	}

	options = append([]DecoderOption{o}, options...)

	for _, option := range options {
		if err := option(decoder); err != nil {
			return nil, err
		}
	}

	if decoder.buffer == nil {
		decoder.buffer = buffer.NewChannelBufferWithGenerator(ctx, buffer.CreateFramePool(), 256, 1)
	}

	if err := decoder.decoderContext.Open(decoder.codec, nil); err != nil {
		return nil, err
	}

	return decoder, nil
}

func (d *GeneralDecoder) Start() {
	go d.loop()
}

func (d *GeneralDecoder) Close() {
	d.once.Do(func() {
		if d.cancel != nil {
			d.cancel()
		}

		d.wg.Wait()

		d.close()
	})
}

func (d *GeneralDecoder) loop() {
	d.wg.Add(1)
	defer d.wg.Done()

loop1:
	for {
		select {
		case <-d.ctx.Done():
			return
		default:
			packet, err := d.getPacket()
			if err != nil {
				continue
			}
			if err := d.decoderContext.SendPacket(packet); err != nil {
				d.demuxer.PutBack(packet)
				if !errors.Is(err, astiav.ErrEagain) {
					continue loop1
				}
			}
		loop2:
			for {
				frame := d.buffer.Get()
				if err := d.decoderContext.ReceiveFrame(frame); err != nil {
					d.buffer.Put(frame)
					break loop2
				}

				frame.SetPictureType(astiav.PictureTypeNone) // this is needed as the ffmpeg decoder picture type is different

				if err := d.pushFrame(frame); err != nil {
					d.buffer.Put(frame)
					continue loop2
				}
			}
			d.demuxer.PutBack(packet)
		}
	}
}

func (d *GeneralDecoder) pushFrame(frame *astiav.Frame) error {
	ctx, cancel := context.WithTimeout(d.ctx, 50*time.Millisecond)
	defer cancel()

	return d.buffer.Push(ctx, frame)
}

func (d *GeneralDecoder) getPacket() (*astiav.Packet, error) {
	ctx, cancel := context.WithTimeout(d.ctx, 50*time.Millisecond)
	defer cancel()

	return d.demuxer.GetPacket(ctx)
}

func (d *GeneralDecoder) GetFrame(ctx context.Context) (*astiav.Frame, error) {
	return d.buffer.Pop(ctx)
}

func (d *GeneralDecoder) PutBack(frame *astiav.Frame) {
	d.buffer.Put(frame)
}

func (d *GeneralDecoder) close() {
	if d.decoderContext != nil {
		d.decoderContext.Free()
	}
}

func (d *GeneralDecoder) SetBuffer(buffer buffer.BufferWithGenerator[*astiav.Frame]) {
	d.buffer = buffer
}

func (d *GeneralDecoder) SetCodec(producer CanDescribeMediaPacket) error {
	if d.codec = astiav.FindDecoder(producer.CodecID()); d.codec == nil {
		return ErrorNoCodecFound
	}
	d.decoderContext = astiav.AllocCodecContext(d.codec)
	if d.decoderContext == nil {
		return ErrorAllocateCodecContext
	}

	return nil
}

func (d *GeneralDecoder) FillContextContent(producer CanDescribeMediaPacket) error {
	return producer.GetCodecParameters().ToCodecContext(d.decoderContext)
}

func (d *GeneralDecoder) SetFrameRate(producer CanDescribeFrameRate) {
	d.decoderContext.SetFramerate(producer.FrameRate())
}

func (d *GeneralDecoder) SetTimeBase(producer CanDescribeTimeBase) {
	d.decoderContext.SetTimeBase(producer.TimeBase())
}

// ### IMPLEMENTS CanDescribeMediaVideoFrame

func (d *GeneralDecoder) FrameRate() astiav.Rational {
	return d.decoderContext.Framerate()
}

func (d *GeneralDecoder) TimeBase() astiav.Rational {
	return d.decoderContext.TimeBase()
}

func (d *GeneralDecoder) Height() int {
	return d.decoderContext.Height()
}

func (d *GeneralDecoder) Width() int {
	return d.decoderContext.Width()
}

func (d *GeneralDecoder) PixelFormat() astiav.PixelFormat {
	return d.decoderContext.PixelFormat()
}

func (d *GeneralDecoder) SampleAspectRatio() astiav.Rational {
	return d.decoderContext.SampleAspectRatio()
}

func (d *GeneralDecoder) ColorSpace() astiav.ColorSpace {
	return d.decoderContext.ColorSpace()
}

func (d *GeneralDecoder) ColorRange() astiav.ColorRange {
	return d.decoderContext.ColorRange()
}

// ## CanDescribeMediaAudioFrame

func (d *GeneralDecoder) SampleRate() int {
	return d.decoderContext.SampleRate()
}

func (d *GeneralDecoder) SampleFormat() astiav.SampleFormat {
	return d.decoderContext.SampleFormat()
}

func (d *GeneralDecoder) ChannelLayout() astiav.ChannelLayout {
	return d.decoderContext.ChannelLayout()
}

// ## CanDescribeMediaFrame

func (d *GeneralDecoder) MediaType() astiav.MediaType {
	return d.decoderContext.MediaType()
}
