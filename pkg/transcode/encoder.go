//go:build cgo_enabled

package transcode

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/asticode/go-astiav"

	"github.com/harshabose/tools/pkg/buffer"
)

type GeneralEncoder struct {
	buffer   buffer.BufferWithGenerator[*astiav.Packet]
	producer CanProduceMediaFrame

	codec           *astiav.Codec
	encoderContext  *astiav.CodecContext
	codecFlags      *astiav.Dictionary
	encoderSettings codecSettings
	sps             []byte
	pps             []byte

	once   sync.Once
	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
}

func CreateGeneralEncoder(ctx context.Context, codecID astiav.CodecID, canProduceMediaFrame CanProduceMediaFrame, options ...EncoderOption) (*GeneralEncoder, error) {
	ctx2, cancel2 := context.WithCancel(ctx)
	encoder := &GeneralEncoder{
		producer:   canProduceMediaFrame,
		codecFlags: astiav.NewDictionary(),
		ctx:        ctx2,
		cancel:     cancel2,
	}

	encoder.codec = astiav.FindEncoder(codecID)
	if encoder.encoderContext = astiav.AllocCodecContext(encoder.codec); encoder.encoderContext == nil {
		return nil, ErrorAllocateCodecContext
	}

	canDescribeMediaFrame, ok := canProduceMediaFrame.(CanDescribeMediaFrame)
	if !ok {
		return nil, ErrorInterfaceMismatch
	}
	if canDescribeMediaFrame.MediaType() == astiav.MediaTypeAudio {
		withAudioSetEncoderContextParameters(canDescribeMediaFrame, encoder.encoderContext)
	}
	if canDescribeMediaFrame.MediaType() == astiav.MediaTypeVideo {
		withVideoSetEncoderContextParameter(canDescribeMediaFrame, encoder.encoderContext)
	}

	for _, option := range options {
		if err := option(encoder); err != nil {
			return nil, err
		}
	}

	if encoder.encoderSettings == nil {
		fmt.Println("warn: no encoder settings are provided")
	}

	encoder.encoderContext.SetFlags(astiav.NewCodecContextFlags(astiav.CodecContextFlagGlobalHeader))

	if err := encoder.encoderContext.Open(encoder.codec, encoder.codecFlags); err != nil {
		return nil, err
	}

	if encoder.buffer == nil {
		encoder.buffer = buffer.NewChannelBufferWithGenerator(ctx2, buffer.CreatePacketPool(), 256, 1)
	}

	encoder.findParameterSets(encoder.encoderContext.ExtraData())

	return encoder, nil
}

func (e *GeneralEncoder) Start() {
	go e.loop()
}

func (e *GeneralEncoder) GetParameterSets() ([]byte, []byte, error) {
	e.findParameterSets(e.encoderContext.ExtraData())
	return e.sps, e.pps, nil
}

func (e *GeneralEncoder) TimeBase() astiav.Rational {
	return e.encoderContext.TimeBase()
}

func (e *GeneralEncoder) loop() {
	e.wg.Add(1)
	defer e.wg.Done()

loop1:
	for {
		select {
		case <-e.ctx.Done():
			return
		default:
			frame, err := e.getFrame()
			if err != nil {
				continue
			}
			if err := e.encoderContext.SendFrame(frame); err != nil {
				e.producer.PutBack(frame)
				if !errors.Is(err, astiav.ErrEagain) {
					continue loop1
				}
			}
		loop2:
			for {
				packet := e.buffer.Get()
				if err = e.encoderContext.ReceivePacket(packet); err != nil {
					e.buffer.Put(packet)
					break loop2
				}

				if err := e.pushPacket(packet); err != nil {
					e.buffer.Put(packet)
					continue loop2
				}
			}
			e.producer.PutBack(frame)
		}
	}
}

func (e *GeneralEncoder) getFrame() (*astiav.Frame, error) {
	ctx, cancel := context.WithTimeout(e.ctx, 100*time.Millisecond)
	defer cancel()

	return e.producer.GetFrame(ctx)
}

func (e *GeneralEncoder) GetPacket(ctx context.Context) (*astiav.Packet, error) {
	return e.buffer.Pop(ctx)
}

func (e *GeneralEncoder) pushPacket(packet *astiav.Packet) error {
	ctx, cancel := context.WithTimeout(e.ctx, 100*time.Millisecond)
	defer cancel()

	return e.buffer.Push(ctx, packet)
}

func (e *GeneralEncoder) PutBack(packet *astiav.Packet) {
	e.buffer.Put(packet)
}

func (e *GeneralEncoder) Close() {
	e.once.Do(func() {
		if e.cancel != nil {
			e.cancel()
		}

		e.wg.Wait()

		e.close()
	})
}

func (e *GeneralEncoder) close() {
	if e.encoderContext != nil {
		e.encoderContext.Free()
	}

	if e.codecFlags != nil {
		e.codecFlags.Free()
	}
}

func (e *GeneralEncoder) findParameterSets(extraData []byte) {
	if len(extraData) > 0 {
		// Find the first start code (0x00000001)
		for i := 0; i < len(extraData)-4; i++ {
			if extraData[i] == 0 && extraData[i+1] == 0 && extraData[i+2] == 0 && extraData[i+3] == 1 {
				// Skip start code to get the NAL type
				nalType := extraData[i+4] & 0x1F

				// Find the next start code or end
				nextStart := len(extraData)
				for j := i + 4; j < len(extraData)-4; j++ {
					if extraData[j] == 0 && extraData[j+1] == 0 && extraData[j+2] == 0 && extraData[j+3] == 1 {
						nextStart = j
						break
					}
				}

				if nalType == 7 { // SPS
					e.sps = make([]byte, nextStart-i)
					copy(e.sps, extraData[i:nextStart])
				} else if nalType == 8 { // PPS
					e.pps = make([]byte, len(extraData)-i)
					copy(e.pps, extraData[i:])
				}

				i = nextStart - 1
			}
		}
		// fmt.Println("SPS for current encoder: ", e.sps)
		// fmt.Println("\tSPS for current encoder in Base64:", base64.StdEncoding.EncodeToString(e.sps))
		// fmt.Println("PPS for current encoder: ", e.pps)
		// fmt.Println("\tPPS for current encoder in Base64:", base64.StdEncoding.EncodeToString(e.pps))
	}
}

func (e *GeneralEncoder) SetBuffer(buffer buffer.BufferWithGenerator[*astiav.Packet]) {
	e.buffer = buffer
}

func (e *GeneralEncoder) SetEncoderCodecSettings(settings codecSettings) error {
	e.encoderSettings = settings
	return e.encoderSettings.ForEach(func(key string, value string) error {
		if value == "" {
			return nil
		}
		return e.codecFlags.Set(key, value, 0)
	})
}

func (e *GeneralEncoder) GetCurrentBitrate() (int64, error) {
	g, ok := e.encoderSettings.(CanGetCurrentBitrate)
	if !ok {
		return 0, ErrorInterfaceMismatch
	}

	return g.GetCurrentBitrate()
}
