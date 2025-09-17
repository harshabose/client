package mediasink

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"reflect"
	"sync"
	"time"

	"github.com/pion/interceptor"
	"github.com/pion/rtp"
	"github.com/pion/webrtc/v4"

	"github.com/harshabose/mediapipe/pkg/generators"
)

type Sink struct {
	generator       generators.CanGeneratePionRTPPacket
	codecCapability *webrtc.RTPCodecParameters
	rtpReceiver     *webrtc.RTPReceiver
	mux             sync.RWMutex
	ctx             context.Context
}

func CreateSink(ctx context.Context, options ...SinkOption) (*Sink, error) {
	sink := &Sink{ctx: ctx}

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

func (s *Sink) setGenerator(generator generators.CanGeneratePionRTPPacket) {
	s.mux.Lock()
	defer s.mux.Unlock()

	s.generator = generator
}

func (s *Sink) setRTPReceiver(receiver *webrtc.RTPReceiver) {
	s.mux.Lock()
	defer s.mux.Unlock()

	s.rtpReceiver = receiver
}

func (s *Sink) readRTPReceiver(rtcpBuf []byte) {
	s.mux.RLock()
	defer s.mux.RUnlock()

	if s.rtpReceiver == nil {
		time.Sleep(10 * time.Millisecond)
		return
	}

	if _, _, err := s.rtpReceiver.Read(rtcpBuf); err != nil {
		fmt.Printf("error while reading rtcp packets")
	}
}

func (s *Sink) rtpReceiverLoop() {
	// THIS IS NEEDED AS interceptors (pion) do not work
	for {
		select {
		case <-s.ctx.Done():
			return
		default:
			rtcpBuf := make([]byte, 1500)
			s.readRTPReceiver(rtcpBuf)
		}
	}
}

func (s *Sink) ReadRTP() (*rtp.Packet, interceptor.Attributes, error) {
	s.mux.RLock()
	defer s.mux.RUnlock()

	if s.generator == nil {
		return nil, interceptor.Attributes{}, nil
	}

	return s.generator.ReadRTP()
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
			// TODO: MAYBE SET A DEFAULT SINK?
			return
		}

		if !CompareRTPCodecParameters(remote.Codec(), *(sink.codecCapability)) {
			fmt.Println("sink registered codec did not match. skipping...")
			return
		}

		sink.setRTPReceiver(receiver)
		sink.setGenerator(remote)

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
