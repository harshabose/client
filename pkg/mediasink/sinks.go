package mediasink

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"reflect"
	"sync"

	"github.com/pion/interceptor"
	"github.com/pion/rtp"
	"github.com/pion/webrtc/v4"

	"github.com/harshabose/tools/pkg/cond"
)

type Sink struct {
	generator       *webrtc.TrackRemote
	codecCapability *webrtc.RTPCodecParameters
	rtpReceiver     *webrtc.RTPReceiver
	mux             sync.RWMutex
	cond            *cond.ContextCond
	ctx             context.Context
}

func CreateSink(ctx context.Context, options ...SinkOption) (*Sink, error) {
	sink := &Sink{ctx: ctx}
	sink.cond = cond.NewContextCond(&(sink.mux))

	for _, option := range options {
		if err := option(sink); err != nil {
			return nil, err
		}
	}

	if sink.codecCapability == nil {
		return nil, errors.New("no sink capabilities given")
	}

	return sink, nil
}

func (s *Sink) setGenerator(generator *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
	s.mux.Lock()
	defer s.mux.Unlock()

	s.generator = generator
	s.rtpReceiver = receiver

	s.cond.Broadcast()
}

func (s *Sink) readRTPReceiver(ctx context.Context, rtcpBuf []byte) error {
	s.mux.Lock()

	for s.rtpReceiver == nil {
		if err := s.cond.Wait(ctx); err != nil {
			s.mux.Unlock()
			return err
		}
	}
	reader := s.rtpReceiver

	s.mux.Unlock()

	if _, _, err := reader.Read(rtcpBuf); err != nil {
		fmt.Printf("error while reading rtcp packets (err=%v)\n", err)
		return err
	}
	return nil
}

func (s *Sink) rtpReceiverLoop() {
	// THIS IS NEEDED AS interceptors (pion) do not work
	for {
		select {
		case <-s.ctx.Done():
			return
		default:
			rtcpBuf := make([]byte, 1500)
			if err := s.readRTPReceiver(s.ctx, rtcpBuf); err != nil {
				return
			}
		}
	}
}

func (s *Sink) ReadRTP(ctx context.Context) (*rtp.Packet, interceptor.Attributes, error) {
	s.cond.L.Lock()

	for s.generator == nil {
		if err := s.cond.Wait(ctx); err != nil {
			s.cond.L.Unlock()
			return nil, nil, err
		}
	}
	reader := s.generator

	s.cond.L.Unlock()

	return reader.ReadRTP()
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
		sink, err := s.GetSink(remote.ID())
		if err != nil {
			fmt.Printf("failed to trigger on track callback with err: %v\n", err)
			return
		}

		if !CompareRTPCodecParameters(remote.Codec(), *(sink.codecCapability)) {
			fmt.Println("sink registered codec did not match. skipping...")
			return
		}

		sink.setGenerator(remote, receiver)

		go sink.rtpReceiverLoop()
	})
}

func (s *Sinks) CreateSink(label string, options ...SinkOption) (*Sink, error) {
	s.mux.Lock()
	defer s.mux.Unlock()

	if _, exists := s.sinks[label]; exists {
		return nil, fmt.Errorf("sink with id='%s' already exists", label)
	}

	sink, err := CreateSink(s.ctx, options...)
	if err != nil {
		return nil, err
	}

	s.sinks[label] = sink
	return sink, nil
}

func (s *Sinks) GetSink(label string) (*Sink, error) {
	s.mux.RLock()
	defer s.mux.RUnlock()

	sink, exists := s.sinks[label]
	if !exists {
		return nil, fmt.Errorf("ERROR: no sink set for track with id %s; ignoring track...\n", label)
	}

	return sink, nil
}

func (s *Sinks) Sinks() iter.Seq2[string, *Sink] {
	return func(yield func(string, *Sink) bool) {
		for id, sink := range s.sinks {
			if !yield(id, sink) {
				return
			}
		}
	}
}

func CompareRTPCodecParameters(a, b webrtc.RTPCodecParameters) bool {
	identical := true

	if a.PayloadType != b.PayloadType {
		fmt.Printf("PayloadType differs: %v != %v\n", a.PayloadType, b.PayloadType)
		identical = false
	}

	if a.MimeType != b.MimeType {
		fmt.Printf("MimeType differs: %s != %s\n", a.MimeType, b.MimeType)
		identical = false
	}

	if a.ClockRate != b.ClockRate {
		fmt.Printf("ClockRate differs: %d != %d\n", a.ClockRate, b.ClockRate)
		identical = false
	}

	if a.Channels != b.Channels {
		fmt.Printf("Channels differs: %d != %d\n", a.Channels, b.Channels)
		identical = false
	}

	if a.SDPFmtpLine != b.SDPFmtpLine {
		fmt.Printf("SDPFmtpLine differs (ignored): %s != %s\n", a.SDPFmtpLine, b.SDPFmtpLine)
	}

	if !reflect.DeepEqual(a.RTCPFeedback, b.RTCPFeedback) {
		fmt.Printf("RTCPFeedback differs (ignored): %v != %v\n", a.RTCPFeedback, b.RTCPFeedback)
	}

	return identical
}
