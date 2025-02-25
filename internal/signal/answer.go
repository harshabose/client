package signal

import (
	"context"

	"github.com/pion/webrtc/v4"
)

type AnswerSignal struct {
	*Signal
	peerConnection *webrtc.PeerConnection
}

func CreateAnswerSignal(ctx context.Context, peerConnection *webrtc.PeerConnection) *AnswerSignal {
	return &AnswerSignal{
		Signal:         CreateSignal(ctx),
		peerConnection: peerConnection,
	}
}
