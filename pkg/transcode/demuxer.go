//go:build cgo_enabled

package transcode

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/asticode/go-astiav"

	"github.com/harshabose/tools/pkg/buffer"
)

type GeneralDemuxer struct {
	formatContext   *astiav.FormatContext
	inputOptions    *astiav.Dictionary
	inputFormat     *astiav.InputFormat
	stream          *astiav.Stream
	codecParameters *astiav.CodecParameters

	buffer buffer.BufferWithGenerator[*astiav.Packet]

	once   sync.Once
	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
}

func CreateGeneralDemuxer(ctx context.Context, containerAddress string, options ...DemuxerOption) (*GeneralDemuxer, error) {
	astiav.RegisterAllDevices()

	ctx2, cancel2 := context.WithCancel(ctx)

	demuxer := &GeneralDemuxer{
		formatContext: astiav.AllocFormatContext(),
		inputOptions:  astiav.NewDictionary(),
		ctx:           ctx2,
		cancel:        cancel2,
	}

	if demuxer.formatContext == nil {
		return nil, ErrorAllocateFormatContext
	}

	if demuxer.inputOptions == nil {
		return nil, fmt.Errorf("error allocating astiav.Dictionary (%w)", ErrorGeneralAllocate)
	}

	for _, option := range options {
		if err := option(demuxer); err != nil {
			return nil, err
		}
	}

	if err := demuxer.formatContext.OpenInput(containerAddress, demuxer.inputFormat, demuxer.inputOptions); err != nil {
		return nil, err
	}

	if err := demuxer.formatContext.FindStreamInfo(nil); err != nil {
		return nil, ErrorNoStreamFound
	}

	for _, stream := range demuxer.formatContext.Streams() {
		demuxer.stream = stream
		break
	}

	if demuxer.stream == nil {
		return nil, ErrorNoVideoStreamFound
	}
	demuxer.codecParameters = demuxer.stream.CodecParameters()

	if demuxer.buffer == nil {
		demuxer.buffer = buffer.NewChannelBufferWithGenerator(ctx, buffer.CreatePacketPool(), 256, 1)
	}

	return demuxer, nil
}

func (d *GeneralDemuxer) Start() {
	go d.loop()
}

func (d *GeneralDemuxer) Close() {
	d.once.Do(func() {
		if d.cancel != nil {
			d.cancel()

			d.wg.Wait()

			d.close()
		}
	})
}

func (d *GeneralDemuxer) loop() {
	d.wg.Add(1)
	defer d.wg.Done()

loop1:
	for {
		select {
		case <-d.ctx.Done():
			return
		default:
		loop2:
			for {
				packet := d.buffer.Get()

				if err := d.formatContext.ReadFrame(packet); err != nil {
					d.buffer.Put(packet)
					continue loop1
				}

				if packet.StreamIndex() != d.stream.Index() {
					d.buffer.Put(packet)
					continue loop2
				}

				if err := d.pushPacket(packet); err != nil {
					d.buffer.Put(packet)
					continue loop1
				}
				break loop2
			}
		}
	}
}

func (d *GeneralDemuxer) pushPacket(packet *astiav.Packet) error {
	ctx, cancel := context.WithTimeout(d.ctx, 50*time.Millisecond) // TODO: NEEDS TO BE BASED ON FPS ON INPUT_FORMAT
	defer cancel()

	return d.buffer.Push(ctx, packet)
}

func (d *GeneralDemuxer) GetPacket(ctx context.Context) (*astiav.Packet, error) {
	return d.buffer.Pop(ctx)
}

func (d *GeneralDemuxer) PutBack(packet *astiav.Packet) {
	d.buffer.Put(packet)
}

func (d *GeneralDemuxer) close() {
	if d.formatContext != nil {
		d.formatContext.CloseInput()
		d.formatContext.Free()
	}
}

func (d *GeneralDemuxer) SetInputOption(key, value string, flags astiav.DictionaryFlags) error {
	return d.inputOptions.Set(key, value, flags)
}

func (d *GeneralDemuxer) SetInputFormat(format *astiav.InputFormat) {
	d.inputFormat = format
}

func (d *GeneralDemuxer) SetBuffer(buffer buffer.BufferWithGenerator[*astiav.Packet]) {
	d.buffer = buffer
}

func (d *GeneralDemuxer) GetCodecParameters() *astiav.CodecParameters {
	return d.codecParameters
}

func (d *GeneralDemuxer) MediaType() astiav.MediaType {
	return d.codecParameters.MediaType()
}

func (d *GeneralDemuxer) CodecID() astiav.CodecID {
	return d.codecParameters.CodecID()
}

func (d *GeneralDemuxer) FrameRate() astiav.Rational {
	return d.formatContext.GuessFrameRate(d.stream, nil)
}

func (d *GeneralDemuxer) TimeBase() astiav.Rational {
	return d.stream.TimeBase()
}
