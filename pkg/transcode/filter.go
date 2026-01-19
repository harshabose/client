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

type GeneralFilter struct {
	content          string
	decoder          CanProduceMediaFrame
	buffer           buffer.BufferWithGenerator[*astiav.Frame]
	graph            *astiav.FilterGraph
	input            *astiav.FilterInOut
	output           *astiav.FilterInOut
	srcContext       *astiav.BuffersrcFilterContext
	sinkContext      *astiav.BuffersinkFilterContext
	srcContextParams *astiav.BuffersrcFilterContextParameters // NOTE: THIS BECOMES NIL AFTER INITIALISATION

	once   sync.Once
	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
}

func CreateGeneralFilter(ctx context.Context, canProduceMediaFrame CanProduceMediaFrame, filterConfig FilterConfig, options ...FilterOption) (*GeneralFilter, error) {
	ctx2, cancel2 := context.WithCancel(ctx)
	filter := &GeneralFilter{
		graph:            astiav.AllocFilterGraph(),
		decoder:          canProduceMediaFrame,
		input:            astiav.AllocFilterInOut(),
		output:           astiav.AllocFilterInOut(),
		srcContextParams: astiav.AllocBuffersrcFilterContextParameters(),
		ctx:              ctx2,
		cancel:           cancel2,
	}

	filterSrc := astiav.FindFilterByName(filterConfig.Source.String())
	if filterSrc == nil {
		return nil, ErrorNoFilterName
	}

	filterSink := astiav.FindFilterByName(filterConfig.Sink.String())
	if filterSink == nil {
		return nil, ErrorNoFilterName
	}

	srcContext, err := filter.graph.NewBuffersrcFilterContext(filterSrc, "in")
	if err != nil {
		return nil, ErrorAllocSrcContext
	}
	filter.srcContext = srcContext

	sinkContext, err := filter.graph.NewBuffersinkFilterContext(filterSink, "out")
	if err != nil {
		return nil, ErrorAllocSinkContext
	}
	filter.sinkContext = sinkContext

	canDescribeMediaFrame, ok := canProduceMediaFrame.(CanDescribeMediaFrame)
	if !ok {
		return nil, ErrorInterfaceMismatch
	}

	if canDescribeMediaFrame.MediaType() != astiav.MediaTypeVideo && canDescribeMediaFrame.MediaType() != astiav.MediaTypeAudio {
		return nil, ErrorUnsupportedMedia
	}

	o := withVideoSetFilterContextParameters(canDescribeMediaFrame)
	if canDescribeMediaFrame.MediaType() == astiav.MediaTypeAudio {
		o = withAudioSetFilterContextParameters(canDescribeMediaFrame)
	}

	options = append([]FilterOption{o}, options...)

	for _, option := range options {
		if err = option(filter); err != nil {
			return nil, err
		}
	}

	if filter.buffer == nil {
		filter.buffer = buffer.NewChannelBufferWithGenerator(ctx, buffer.CreateFramePool(), 256, 1)
	}

	if err = filter.srcContext.SetParameters(filter.srcContextParams); err != nil {
		return nil, ErrorSrcContextSetParameter
	}

	if err = filter.srcContext.Initialize(astiav.NewDictionary()); err != nil {
		return nil, ErrorSrcContextInitialise
	}

	filter.output.SetName("in")
	filter.output.SetFilterContext(filter.srcContext.FilterContext())
	filter.output.SetPadIdx(0)
	filter.output.SetNext(nil)

	filter.input.SetName("out")
	filter.input.SetFilterContext(filter.sinkContext.FilterContext())
	filter.input.SetPadIdx(0)
	filter.input.SetNext(nil)

	if filter.content == "" {
		fmt.Println(WarnNoFilterContent)
	}

	if err = filter.graph.Parse(filter.content, filter.input, filter.output); err != nil {
		return nil, ErrorGraphParse
	}

	if err = filter.graph.Configure(); err != nil {
		return nil, ErrorGraphConfigure
	}

	if filter.srcContextParams != nil {
		filter.srcContextParams.Free()
	}

	return filter, nil
}

func (f *GeneralFilter) Start() {
	go f.loop()
}

func (f *GeneralFilter) Close() {
	f.once.Do(func() {
		if f.cancel != nil {
			f.cancel()
		}

		f.wg.Wait()

		f.close()
	})
}

func (f *GeneralFilter) loop() {
	f.wg.Add(1)
	defer f.wg.Done()

loop1:
	for {
		select {
		case <-f.ctx.Done():
			return
		default:
			srcFrame, err := f.getFrame()
			if err != nil {
				continue
			}
			if err := f.srcContext.AddFrame(srcFrame, astiav.NewBuffersrcFlags(astiav.BuffersrcFlagKeepRef)); err != nil {
				f.buffer.Put(srcFrame)
				continue loop1
			}
		loop2:
			for {
				sinkFrame := f.buffer.Get()
				if err = f.sinkContext.GetFrame(sinkFrame, astiav.NewBuffersinkFlags()); err != nil {
					f.buffer.Put(sinkFrame)
					break loop2
				}

				if err := f.pushFrame(sinkFrame); err != nil {
					f.buffer.Put(sinkFrame)
					continue loop2
				}
			}
			f.decoder.PutBack(srcFrame)
		}
	}
}

func (f *GeneralFilter) pushFrame(frame *astiav.Frame) error {
	ctx, cancel := context.WithTimeout(f.ctx, 50*time.Millisecond)
	defer cancel()

	return f.buffer.Push(ctx, frame)
}

func (f *GeneralFilter) getFrame() (*astiav.Frame, error) {
	ctx, cancel := context.WithTimeout(f.ctx, 50*time.Millisecond)
	defer cancel()

	return f.decoder.GetFrame(ctx)
}

func (f *GeneralFilter) PutBack(frame *astiav.Frame) {
	f.buffer.Put(frame)
}

func (f *GeneralFilter) GetFrame(ctx context.Context) (*astiav.Frame, error) {
	return f.buffer.Pop(ctx)
}

func (f *GeneralFilter) close() {
	if f.graph != nil {
		f.graph.Free()
	}
	if f.input != nil {
		f.input.Free()
	}
	if f.output != nil {
		f.output.Free()
	}
}

func (f *GeneralFilter) SetBuffer(buffer buffer.BufferWithGenerator[*astiav.Frame]) {
	f.buffer = buffer
}

func (f *GeneralFilter) AddToFilterContent(content string) {
	f.content += content
}

func (f *GeneralFilter) SetFrameRate(describe CanDescribeFrameRate) {
	f.srcContextParams.SetFramerate(describe.FrameRate())
}

func (f *GeneralFilter) SetTimeBase(describe CanDescribeTimeBase) {
	f.srcContextParams.SetTimeBase(describe.TimeBase())
}

func (f *GeneralFilter) SetHeight(describe CanDescribeMediaVideoFrame) {
	f.srcContextParams.SetHeight(describe.Height())
}

func (f *GeneralFilter) SetWidth(describe CanDescribeMediaVideoFrame) {
	f.srcContextParams.SetWidth(describe.Width())
}

func (f *GeneralFilter) SetPixelFormat(describe CanDescribeMediaVideoFrame) {
	f.srcContextParams.SetPixelFormat(describe.PixelFormat())
}

func (f *GeneralFilter) SetSampleAspectRatio(describe CanDescribeMediaVideoFrame) {
	f.srcContextParams.SetSampleAspectRatio(describe.SampleAspectRatio())
}

func (f *GeneralFilter) SetColorSpace(describe CanDescribeMediaVideoFrame) {
	f.srcContextParams.SetColorSpace(describe.ColorSpace())
}

func (f *GeneralFilter) SetColorRange(describe CanDescribeMediaVideoFrame) {
	f.srcContextParams.SetColorRange(describe.ColorRange())
}

func (f *GeneralFilter) SetSampleRate(describe CanDescribeMediaAudioFrame) {
	f.srcContextParams.SetSampleRate(describe.SampleRate())
}

func (f *GeneralFilter) SetSampleFormat(describe CanDescribeMediaAudioFrame) {
	f.srcContextParams.SetSampleFormat(describe.SampleFormat())
}

func (f *GeneralFilter) SetChannelLayout(describe CanDescribeMediaAudioFrame) {
	f.srcContextParams.SetChannelLayout(describe.ChannelLayout())
}

func (f *GeneralFilter) MediaType() astiav.MediaType {
	return f.sinkContext.MediaType()
}

func (f *GeneralFilter) FrameRate() astiav.Rational {
	return f.sinkContext.FrameRate()
}

func (f *GeneralFilter) TimeBase() astiav.Rational {
	return f.sinkContext.TimeBase()
}

func (f *GeneralFilter) Height() int {
	return f.sinkContext.Height()
}

func (f *GeneralFilter) Width() int {
	return f.sinkContext.Width()
}

func (f *GeneralFilter) PixelFormat() astiav.PixelFormat {
	return f.sinkContext.PixelFormat()
}

func (f *GeneralFilter) SampleAspectRatio() astiav.Rational {
	return f.sinkContext.SampleAspectRatio()
}

func (f *GeneralFilter) ColorSpace() astiav.ColorSpace {
	return f.sinkContext.ColorSpace()
}

func (f *GeneralFilter) ColorRange() astiav.ColorRange {
	return f.sinkContext.ColorRange()
}

func (f *GeneralFilter) SampleRate() int {
	return f.sinkContext.SampleRate()
}

func (f *GeneralFilter) SampleFormat() astiav.SampleFormat {
	return f.sinkContext.SampleFormat()
}

func (f *GeneralFilter) ChannelLayout() astiav.ChannelLayout {
	return f.sinkContext.ChannelLayout()
}
