package pkg

import (
	"github.com/harshabose/simple_webrtc_comm/datachannel/pkg"
	"github.com/pion/webrtc/v4"
)

type PeerConnection struct {
	peerConnection *webrtc.PeerConnection
	dataChannels   *data.DataChannels
}
