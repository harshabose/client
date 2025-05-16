package internal

import (
	"io"

	"github.com/pion/webrtc/v4"

	"github.com/harshabose/simple_webrtc_comm/datachannel/pkg"
	"github.com/harshabose/simple_webrtc_comm/mediasink/pkg"
)

type PeerConnection interface {
	GetPeerConnection() *webrtc.PeerConnection
	GetLabel() string
	CreateDataChannel(label string, options ...data.LoopBackOption) (*data.DataChannel, error)
	CreateMediaSink(label string, options ...mediasink.StreamOption) error
	Connect(category string) error
	io.Closer
}
