//go:build cgo_enabled

package transcode

import (
	"context"
)

type GeneralFilterBuilder struct {
	producer CanProduceMediaFrame
	config   FilterConfig
	options  []FilterOption

	fps       uint8
	fpsOption FilterOption
}

func NewGeneralFilterBuilder(config FilterConfig, producer CanProduceMediaFrame, options ...FilterOption) *GeneralFilterBuilder {
	return &GeneralFilterBuilder{
		producer: producer,
		config:   config,
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
