//go:build cgo_enabled

package transcode

import (
	"github.com/asticode/go-astiav"

	"github.com/harshabose/tools/pkg/buffer"
)

func WithRTSPInputOption(demuxer Demuxer) error {
	s, ok := demuxer.(CanSetDemuxerInputOption)
	if !ok {
		return ErrorInterfaceMismatch
	}
	if err := s.SetInputOption("rtsp_transport", "tcp", 0); err != nil {
		return err
	}
	if err := s.SetInputOption("stimeout", "5000000", 0); err != nil {
		return err
	}
	if err := s.SetInputOption("fflags", "nobuffer", 0); err != nil {
		return err
	}
	if err := s.SetInputOption("flags", "low_delay", 0); err != nil {
		return err
	}
	if err := s.SetInputOption("reorder_queue_size", "0", 0); err != nil {
		return err
	}

	return nil
}

func WithFileInputOption(demuxer Demuxer) error {
	s, ok := demuxer.(CanSetDemuxerInputOption)
	if !ok {
		return ErrorInterfaceMismatch
	}
	if err := s.SetInputOption("re", "", 0); err != nil {
		return err
	}
	// // Additional options for smooth playback
	// if err := demuxer.inputOptions.SetInputOption("fflags", "+genpts", 0); err != nil {
	// 	return err
	// }

	return nil
}

func WithAlsaInputFormatOption(demuxer Demuxer) error {
	s, ok := demuxer.(CanSetDemuxerInputFormat)
	if !ok {
		return ErrorInterfaceMismatch
	}
	f := astiav.FindInputFormat("alsa")
	if f == nil {
		return ErrorInputFormatDoesNotExists
	}

	s.SetInputFormat(f)
	return nil
}

func WithAvFoundationInputFormatOption(demuxer Demuxer) error {
	setInputFormat, ok := demuxer.(CanSetDemuxerInputFormat)
	if !ok {
		return ErrorInterfaceMismatch
	}
	f := astiav.FindInputFormat("avfoundation")
	if f == nil {
		return ErrorInputFormatDoesNotExists
	}

	setInputFormat.SetInputFormat(f)

	setInputOption, ok := demuxer.(CanSetDemuxerInputOption)
	if !ok {
		return ErrorInterfaceMismatch
	}

	if err := setInputOption.SetInputOption("video_size", "1280x720", 0); err != nil {
		return err
	}

	if err := setInputOption.SetInputOption("framerate", "30", 0); err != nil {
		return err
	}

	if err := setInputOption.SetInputOption("pixel_format", "uyvy422", 0); err != nil {
		return err
	}

	return nil
}

func WithDemuxerBuffer(size int, pool buffer.Pool[*astiav.Packet]) DemuxerOption {
	return func(demuxer Demuxer) error {
		s, ok := demuxer.(CanSetBuffer[*astiav.Packet])
		if !ok {
			return ErrorInterfaceMismatch
		}
		s.SetBuffer(buffer.NewChannelBufferWithGenerator(demuxer.Ctx(), pool, uint(size), 1))
		return nil
	}
}
