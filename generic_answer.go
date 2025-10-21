package client

import (
	"context"
	"errors"
	"fmt"

	"github.com/pion/webrtc/v4"
)

type GenericAnswerSignal struct {
	ctx context.Context

	forOffer ForOffer
	onAnswer OnAnswer
}

func NewGenericAnswerSignal(ctx context.Context, onAnswer OnAnswer, forOffer ForOffer) *GenericAnswerSignal {
	return &GenericAnswerSignal{
		ctx:      ctx,
		forOffer: forOffer,
		onAnswer: onAnswer,
	}
}

func (s *GenericAnswerSignal) Connect(_ string, pc *PeerConnection) error {
	if s.onAnswer == nil || s.forOffer == nil {
		return errors.New("connect method cannot be used. use offer and answer methods instead")
	}

	offerSDP, err := s.forOffer(s.ctx)
	if err != nil {
		return err
	}

	if err := s.Offer(pc, offerSDP); err != nil {
		return err
	}

	answerSDP, err := s.Answer(pc)
	if err != nil {
		return err
	}

	return s.onAnswer(s.ctx, answerSDP)
}

// Offer sets the remote offer SDP on the PeerConnection.
func (s *GenericAnswerSignal) Offer(pc *PeerConnection, sdp string) error {
	if err := pc.peerConnection.SetRemoteDescription(webrtc.SessionDescription{
		Type: webrtc.SDPTypeOffer,
		SDP:  sdp,
	}); err != nil {
		return fmt.Errorf("failed to set remote description (pc=%s); err: %w", pc.GetLabel(), err)
	}
	return nil
}

// Answer creates and sets the local answer, waits for ICE gathering to complete, and returns the local SDP.
func (s *GenericAnswerSignal) Answer(pc *PeerConnection) (string, error) {
	answer, err := pc.peerConnection.CreateAnswer(nil)
	if err != nil {
		return "", fmt.Errorf("error while creating answer: %w", err)
	}

	if err := pc.peerConnection.SetLocalDescription(answer); err != nil {
		return "", fmt.Errorf("error while setting local sdp: %w", err)
	}

	select {
	case <-s.ctx.Done():
		return "", fmt.Errorf("failed to gather ICE candidates within context deadline; err: %w", s.ctx.Err())
	case <-webrtc.GatheringCompletePromise(pc.peerConnection):
	}

	return pc.peerConnection.LocalDescription().SDP, nil
}

func (s *GenericAnswerSignal) Close() error {
	// NOTE: INTENTIONALLY EMPTY
	return nil
}
