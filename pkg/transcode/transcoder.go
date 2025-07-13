//go:build cgo_enabled

package transcode

import (
	"context"
	"fmt"

	"github.com/asticode/go-astiav"
)

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
	fmt.Println("started encoder")
	t.demuxer.Start()
	t.decoder.Start()
	t.filter.Start()
	t.encoder.Start()
}

func (t *Transcoder) Stop() {
	t.encoder.Stop()
	t.filter.Stop()
	t.decoder.Stop()
	t.demuxer.Stop()
}

func (t *Transcoder) GetPacket(ctx context.Context) (*astiav.Packet, error) {
	return t.encoder.GetPacket(ctx)
}

func (t *Transcoder) PutBack(packet *astiav.Packet) {
	t.encoder.PutBack(packet)
}

// Generate method is to satisfy mediapipe.CanGenerate interface. TODO: but I would prefer to integrate with PutBack
func (t *Transcoder) Generate() (*astiav.Packet, error) {
	packet, err := t.encoder.GetPacket(t.encoder.Ctx())
	if err != nil {
		return nil, err
	}
	return packet, nil
}

func (t *Transcoder) PauseEncoding() error {
	p, ok := t.encoder.(CanPauseUnPauseEncoder)
	if !ok {
		return ErrorInterfaceMismatch
	}

	return p.PauseEncoding()
}

func (t *Transcoder) UnPauseEncoding() error {
	p, ok := t.encoder.(CanPauseUnPauseEncoder)
	if !ok {
		return ErrorInterfaceMismatch
	}

	return p.UnPauseEncoding()
}

func (t *Transcoder) GetParameterSets() (sps, pps []byte, err error) {
	p, ok := t.encoder.(CanGetParameterSets)
	if !ok {
		return nil, nil, ErrorInterfaceMismatch
	}

	return p.GetParameterSets()
}

func (t *Transcoder) UpdateBitrate(bps int64) error {
	u, ok := t.encoder.(CanUpdateBitrate)
	if !ok {
		return ErrorInterfaceMismatch
	}

	return u.UpdateBitrate(bps)
}

func (t *Transcoder) OnUpdateBitrate() UpdateBitrateCallBack {
	return t.UpdateBitrate
}
