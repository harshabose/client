package pkg

import "github.com/pion/webrtc/v4"

type PeerConnection struct {
	peerConnection *webrtc.PeerConnection
	dataChannels   []*webrtc.DataChannel
}
