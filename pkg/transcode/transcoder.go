//go:build cgo_enabled

package transcode

import (
	"context"

	"github.com/asticode/go-astiav"
)

type DecoderOption = func(Decoder) error
type DemuxerOption = func(Demuxer) error
type FilterOption = func(Filter) error
type EncoderOption = func(Encoder) error
type TranscoderOption = func(*Transcoder) error

type Transcoder struct {
	demuxer Demuxer
	decoder Decoder
	filter  Filter
	encoder Encoder
}

func CreateTranscoder(options ...TranscoderOption) (*Transcoder, error) {
	t := &Transcoder{}
	for _, option := range options {
		if err := option(t); err != nil {
			return nil, err
		}
	}

	return t, nil
}

func NewTranscoder(demuxer Demuxer, decoder Decoder, filter Filter, encoder Encoder) *Transcoder {
	return &Transcoder{
		demuxer: demuxer,
		decoder: decoder,
		filter:  filter,
		encoder: encoder,
	}
}

func (t *Transcoder) Start() {
	t.demuxer.Start()
	t.decoder.Start()
	t.filter.Start()
	t.encoder.Start()
}

func (t *Transcoder) Close() {
	t.encoder.Close()
	t.filter.Close()
	t.decoder.Close()
	t.demuxer.Close()
}

func (t *Transcoder) GetPacket(ctx context.Context) (*astiav.Packet, error) {
	return t.encoder.GetPacket(ctx)
}

func (t *Transcoder) PutBack(packet *astiav.Packet) {
	t.encoder.PutBack(packet)
}

// Generate method is to satisfy mediapipe.CanGenerate interface.
func (t *Transcoder) Generate(ctx context.Context) (*astiav.Packet, error) {
	packet, err := t.encoder.GetPacket(ctx)
	if err != nil {
		return nil, err
	}
	return packet, nil
}

func (t *Transcoder) GetParameterSets() (sps, pps []byte, err error) {
	p, ok := t.encoder.(CanGetParameterSets)
	if !ok {
		return nil, nil, ErrorInterfaceMismatch
	}

	return p.GetParameterSets()
}

func (t *Transcoder) AdaptBitrate(bps int64) error {
	f, ok := t.filter.(CanAdaptBitrate)
	if ok {
		if err := f.AdaptBitrate(bps); err != nil {
			return err
		}
	}

	u, ok := t.encoder.(CanAdaptBitrate)
	if ok {
		if err := u.AdaptBitrate(bps); err != nil {
			return err
		}

		return nil
	}

	return ErrorInterfaceMismatch
}

func (t *Transcoder) GetCurrentBitrate() (int64, error) {
	e, ok := t.encoder.(CanGetCurrentBitrate)
	if ok {
		return e.GetCurrentBitrate()
	}

	return 0, ErrorInterfaceMismatch
}

func (t *Transcoder) GetCurrentFPS() (uint8, error) {
	f, ok := t.filter.(CanGetCurrentFPS)
	if ok {
		return f.GetCurrentFPS()
	}

	return 0, ErrorInterfaceMismatch
}

func (t *Transcoder) OnUpdateBitrate() UpdateBitrateCallBack {
	return t.AdaptBitrate
}
