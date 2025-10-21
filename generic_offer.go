package client

import (
	"context"
	"errors"
	"fmt"

	"github.com/pion/webrtc/v4"
)

type GenericOfferSignal struct {
	ctx context.Context

	onOffer   OnOffer
	forAnswer ForAnswer
}

func NewGenericOfferSignal(ctx context.Context, onOffer OnOffer, forAnswer ForAnswer) *GenericOfferSignal {
	return &GenericOfferSignal{
		ctx:       ctx,
		onOffer:   onOffer,
		forAnswer: forAnswer,
	}
}

func (s *GenericOfferSignal) Connect(_ string, pc *PeerConnection) error {
	if s.onOffer == nil || s.forAnswer == nil {
		return errors.New("connect method cannot be used. use offer and answer methods instead")
	}

	offer, err := s.Offer(pc)
	if err != nil {
		return err
	}

	if err := s.onOffer(s.ctx, offer); err != nil {
		return err
	}

	answer, err := s.forAnswer(s.ctx)
	if err != nil {
		return err
	}

	if err := s.Answer(pc, answer); err != nil {
		return err
	}

	return nil
}

func (s *GenericOfferSignal) Offer(pc *PeerConnection) (string, error) {
	offer, err := pc.peerConnection.CreateOffer(nil)
	if err != nil {
		return "", fmt.Errorf("error while creating offer: %w", err)
	}

	if err := pc.peerConnection.SetLocalDescription(offer); err != nil {
		return "", fmt.Errorf("error while setting local sdp: %w", err)

	}

	select {
	case <-s.ctx.Done():
		return "", fmt.Errorf("failed to gather ICE candidates within context deadline; err: %w", s.ctx.Err())
	case <-webrtc.GatheringCompletePromise(pc.peerConnection):
	}

	return pc.peerConnection.LocalDescription().SDP, nil
}

func (s *GenericOfferSignal) Answer(pc *PeerConnection, sdp string) error {
	if err := pc.peerConnection.SetRemoteDescription(webrtc.SessionDescription{
		Type: webrtc.SDPTypeAnswer,
		SDP:  sdp,
	}); err != nil {
		return fmt.Errorf("failed to set remote description (pc=%s); err: %w", pc.GetLabel(), err)
	}

	return nil
}

func (s *GenericOfferSignal) Close() error {
	// NOTE: INTENTIONALLY EMPTY
	return nil
}
