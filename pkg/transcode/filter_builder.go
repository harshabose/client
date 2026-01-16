//go:build cgo_enabled

package transcode

import (
	"context"

	"github.com/asticode/go-astiav"

	"github.com/harshabose/tools/pkg/buffer"
)

type GeneralFilterBuilder struct {
	producer CanProduceMediaFrame
	config   FilterConfig
	bufsize  int
	pool     buffer.Pool[*astiav.Frame]
	options  []FilterOption

	fps       uint8
	fpsOption FilterOption
}

func NewGeneralFilterBuilder(config FilterConfig, producer CanProduceMediaFrame, bufsize int, pool buffer.Pool[*astiav.Frame], options ...FilterOption) *GeneralFilterBuilder {
	return &GeneralFilterBuilder{
		producer: producer,
		config:   config,
		bufsize:  bufsize,
		pool:     pool,
		options:  options,
	}
}

func (b *GeneralFilterBuilder) AdaptFPS(fps uint8) error {
	b.fps = fps
	b.fpsOption = WithVideoFPSFilterContent(fps)
	return nil
}

func (b *GeneralFilterBuilder) GetCurrentFPS() (uint8, error) {
	return b.fps, nil
}

func (b *GeneralFilterBuilder) Build(ctx context.Context) (Filter, error) {
	return CreateGeneralFilter(ctx, b.producer, b.config, append(b.options, b.fpsOption)...)
}
