package signal

import (
	"context"

	"github.com/pion/webrtc/v4"
)

type OfferSignal struct {
	*Signal
	peerConnection *webrtc.PeerConnection
}

func CreateOfferSignal(ctx context.Context, peerConnection *webrtc.PeerConnection) *OfferSignal {
	return &OfferSignal{
		Signal:         CreateSignal(ctx),
		peerConnection: peerConnection,
	}
}
