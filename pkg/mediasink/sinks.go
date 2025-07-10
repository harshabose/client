package mediasink

import (
	"context"
	"fmt"
	"sync"

	"github.com/pion/webrtc/v4"
)

type Sink struct {
	operator func(context.Context, *webrtc.TrackRemote) error
}

func (s *Sink) CreateSink(operator func(context.Context, *webrtc.TrackRemote) error) {
	s.operator = operator
}

type Sinks struct {
	sinks map[string]*Sink
	mux   sync.RWMutex
	ctx   context.Context
}

func CreateSinks(ctx context.Context, pc *webrtc.PeerConnection) *Sinks {
	s := &Sinks{
		sinks: make(map[string]*Sink),
		ctx:   ctx,
	}

	s.onTrack(pc)
	return s
}

func (s *Sinks) onTrack(pc *webrtc.PeerConnection) {
	pc.OnTrack(func(remote *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
		fmt.Println("triggered on track")
		s.mux.RLock()
		sink, exists := s.sinks[remote.ID()]
		if !exists {
			s.mux.RUnlock()
			fmt.Printf("ERROR: no sink set for track with id %s; ignoring track...\n", remote.ID())
			return
		}
		s.mux.RUnlock()

		if err := sink.operator(s.ctx, remote); err != nil {
			fmt.Printf("ERROR: failed to operate on track (id=%s); err: %v\n", remote.ID(), err)
		}
	})
}

func (s *Sinks) CreateSink(label string, operator func(context.Context, *webrtc.TrackRemote) error) (*Sink, error) {
	s.mux.Lock()
	defer s.mux.Unlock()

	if _, exists := s.sinks[label]; exists {
		return nil, fmt.Errorf("sink with id = '%s' already exists", label)
	}

	s.sinks[label] = &Sink{operator: operator}
	return s.sinks[label], nil
}
