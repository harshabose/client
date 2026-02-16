package client

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"sync"

	"github.com/pion/webrtc/v4"

	"github.com/harshabose/simple_webrtc_comm/client/pkg/datachannel"
	"github.com/harshabose/simple_webrtc_comm/client/pkg/mediasink"
	"github.com/harshabose/simple_webrtc_comm/client/pkg/mediasource"
	"github.com/harshabose/tools/pkg/cond"
	"github.com/harshabose/tools/pkg/multierr"
)

/* the following code blocks are from pion webrtc.
	// note: commented out values are not available
	stat := CandidatePairStats{
				Timestamp:         time.Now(),
				LocalCandidateID:  cp.Local.ID(),
				RemoteCandidateID: cp.Remote.ID(),
				State:             cp.state,
				Nominated:         cp.nominated,
				// PacketsSent uint32
				// PacketsReceived uint32
				// BytesSent uint64
				// BytesReceived uint64
				// LastPacketSentTimestamp time.Time
				// LastPacketReceivedTimestamp time.Time
				FirstRequestTimestamp:         cp.FirstRequestSentAt(),
				LastRequestTimestamp:          cp.LastRequestSentAt(),
				FirstResponseTimestamp:        cp.FirstReponseReceivedAt(),
				LastResponseTimestamp:         cp.LastResponseReceivedAt(),
				FirstRequestReceivedTimestamp: cp.FirstRequestReceivedAt(),
				LastRequestReceivedTimestamp:  cp.LastRequestReceivedAt(),

				TotalRoundTripTime:   cp.TotalRoundTripTime(),
				CurrentRoundTripTime: cp.CurrentRoundTripTime(),
				// AvailableOutgoingBitrate float64
				// AvailableIncomingBitrate float64
				// CircuitBreakerTriggerCount uint32
				RequestsReceived:  cp.RequestsReceived(),
				RequestsSent:      cp.RequestsSent(),
				ResponsesReceived: cp.ResponsesReceived(),
				ResponsesSent:     cp.ResponsesSent(),
				// RetransmissionsReceived uint64
				// RetransmissionsSent uint64
				// ConsentRequestsSent uint64
				// ConsentExpiredTimestamp time.Time
			}
	stats := TransportStats{
				Timestamp: statsTimestampFrom(time.Now()),
				Type:      StatsTypeTransport,
				ID:        "iceTransport",
				BytesSent = conn.BytesSent()
				BytesReceived = conn.BytesReceived(),
				// PacketsSent
    			// PacketsReceived
    			// RTCPTransportStatsID
    			// ICERole
    			// DTLSState
    			// ICEState
    			// SelectedCandidatePairID
    			// LocalCertificateID
    			// RemoteCertificateID
    			// DTLSCipher
    			// SRTPCipher
			}
	stats := SCTPTransportStats{
				Timestamp: statsTimestampFrom(time.Now()),
				Type:      StatsTypeSCTPTransport,
				ID:        "sctpTransport",
				// TransportID string
				// UNACKData uint32
				BytesSent: association.BytesSent()
				BytesReceived: association.BytesReceived()
				SmoothedRoundTripTime: association.SRTT() * 0.001 // convert milliseconds to seconds
				CongestionWindow: association.CWND()
				ReceiverWindow: association.RWND()
				MTU: association.MTU()
			}

	stats := CertificateStats{
				Timestamp:            statsTimestampFrom(time.Now()),
				Type:                 StatsTypeCertificate,
				ID:                   c.statsID,
				Fingerprint:          fingerPrintAlgo[0].Value,
				FingerprintAlgorithm: fingerPrintAlgo[0].Algorithm,
				Base64Certificate:    base64Certificate,
				IssuerCertificateID:  c.x509Cert.Issuer.String(),
			}

	stats := CodecStats{
				Timestamp:   statsTimestampFrom(time.Now()),
				Type:        StatsTypeCodec,
				ID:          codec.statsID,
				PayloadType: codec.PayloadType,
				// CodecType CodecType
				// TransportID string
				MimeType:    codec.MimeType,
				ClockRate:   codec.ClockRate,
				Channels:    uint8(codec.Channels), //nolint:gosec // G115
				SDPFmtpLine: codec.SDPFmtpLine,
				// Implementation string
			}
*/

type Stat struct {
	PeerConnectionStat   webrtc.PeerConnectionStats         `json:"peer_connection_stat"` // note: peer connection stats are fully fulfilled
	ICECandidatePairStat webrtc.ICECandidatePairStats       `json:"ice_candidate_pair_stat"`
	CertificateStats     map[string]webrtc.CertificateStats `json:"certificate_stats"`
	CodecStats           map[string]webrtc.CodecStats       `json:"codec_stats"`
	ICETransportStat     webrtc.TransportStats              `json:"ice_transport_stat"`
	SCTPTransportStat    webrtc.SCTPTransportStats          `json:"sctp_transport_stat"`
}

type stat struct {
	*Stat
	pc  *PeerConnection
	mux sync.RWMutex
}

func newStat(pc *PeerConnection) *stat {
	return &stat{
		pc: pc,
		Stat: &Stat{
			CertificateStats: make(map[string]webrtc.CertificateStats),
			CodecStats:       make(map[string]webrtc.CodecStats),
		},
	}
}

func (s *stat) Consume(stats webrtc.Stats) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	switch stat := stats.(type) {
	case webrtc.PeerConnectionStats:
		s.Stat.PeerConnectionStat = stat
		return nil
	case webrtc.ICECandidatePairStats:
		if stat.State == webrtc.StatsICECandidatePairStateSucceeded && stat.Nominated {
			s.ICECandidatePairStat = stat
		}

		return nil

	case webrtc.CertificateStats:
		s.CertificateStats[stat.ID] = stat
		return nil

	case webrtc.CodecStats:
		s.CodecStats[stat.ID] = stat
		return nil

	case webrtc.TransportStats:
		s.ICETransportStat = stat
		return nil

	case webrtc.SCTPTransportStats:
		s.SCTPTransportStat = stat
		return nil

	default:
		return nil
	}
}

func (s *stat) Generate() Stat {
	s.mux.RLock()
	defer s.mux.RUnlock()

	certificatesCopy := make(map[string]webrtc.CertificateStats, len(s.CertificateStats))
	for k, v := range s.CertificateStats {
		certificatesCopy[k] = v
	}

	codecCopy := make(map[string]webrtc.CodecStats, len(s.CodecStats))
	for k, v := range s.CodecStats {
		codecCopy[k] = v
	}

	return Stat{
		PeerConnectionStat:   s.Stat.PeerConnectionStat,
		ICECandidatePairStat: s.ICECandidatePairStat,
		CertificateStats:     certificatesCopy,
		CodecStats:           codecCopy,
		ICETransportStat:     s.ICETransportStat,
		SCTPTransportStat:    s.SCTPTransportStat,
	}
}

type PeerConnection struct {
	label          string
	peerConnection *webrtc.PeerConnection

	dataChannels *datachannel.DataChannels
	tracks       *mediasource.Tracks
	sinks        *mediasink.Sinks
	bwc          *BWEController
	stat         *stat

	cond   *cond.ContextCond
	state  webrtc.PeerConnectionState
	once   sync.Once
	ctx    context.Context
	cancel context.CancelFunc
}

func CreatePeerConnection(ctx context.Context, label string, api *webrtc.API, config webrtc.Configuration) (*PeerConnection, error) {
	peerConnection, err := api.NewPeerConnection(config)
	if err != nil {
		return nil, err
	}

	ctx2, cancel2 := context.WithCancel(ctx)

	pc := &PeerConnection{
		label:          label,
		peerConnection: peerConnection,
		ctx:            ctx2,
		cancel:         cancel2,
		dataChannels:   datachannel.CreateDataChannels(ctx2),
		bwc:            createBWController(ctx2),
		tracks:         mediasource.CreateTracks(ctx2),
		sinks:          mediasink.CreateSinks(ctx2, peerConnection),
		state:          webrtc.PeerConnectionStateUnknown,
		cond:           cond.NewContextCond(&sync.Mutex{}),
	}

	pc.stat = newStat(pc)

	return pc.onConnectionStateChangeEvent().onICEConnectionStateChange().onICEGatheringStateChange().onICECandidate(), err
}

func (pc *PeerConnection) WaitTillOpen(ctx context.Context) error {
	pc.cond.L.Lock()
	defer pc.cond.L.Unlock()

	for pc.state != webrtc.PeerConnectionStateConnected {
		if err := pc.cond.Wait(ctx); err != nil {
			return err
		}
	}

	return nil
}

func (pc *PeerConnection) Done() <-chan struct{} {
	return pc.ctx.Done()
}

func (pc *PeerConnection) GetLabel() string {
	return pc.label
}

func (pc *PeerConnection) GetPeerConnection() *webrtc.PeerConnection {
	return pc.peerConnection
}

func (pc *PeerConnection) GetDataChannel(label string) (*datachannel.DataChannel, error) {
	return pc.dataChannels.GetDataChannel(label)
}

func (pc *PeerConnection) GetMediaSource(label string) (*mediasource.Track, error) {
	return pc.tracks.GetTrack(label)
}

func (pc *PeerConnection) GetRTPMediaSource(label string) (*mediasource.RTPTrack, error) {
	return pc.tracks.GetRTPTrack(label)
}

func (pc *PeerConnection) GetMediaSink(label string) (*mediasink.Sink, error) {
	return pc.sinks.GetSink(label)
}

func (pc *PeerConnection) DataChannels() iter.Seq2[string, *datachannel.DataChannel] {
	return pc.dataChannels.DataChannels()
}

func (pc *PeerConnection) MediaSources() iter.Seq2[string, *mediasource.Track] {
	return pc.tracks.Tracks()
}

func (pc *PeerConnection) MediaSinks() iter.Seq2[string, *mediasink.Sink] {
	return pc.sinks.Sinks()
}

func (pc *PeerConnection) onConnectionStateChangeEvent() *PeerConnection {
	pc.peerConnection.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		fmt.Printf("peer connection state with label changed to %s\n", state.String())
		pc.cond.L.Lock()
		defer pc.cond.L.Unlock()

		pc.state = state
		pc.cond.Broadcast()
	})
	return pc
}

func (pc *PeerConnection) onICEConnectionStateChange() *PeerConnection {
	pc.peerConnection.OnICEConnectionStateChange(func(state webrtc.ICEConnectionState) {
		fmt.Printf("ICE Connection State changed: %s\n", state.String())
	})
	return pc
}

func (pc *PeerConnection) onICEGatheringStateChange() *PeerConnection {
	pc.peerConnection.OnICEGatheringStateChange(func(state webrtc.ICEGatheringState) {
		fmt.Printf("ICE Gathering State changed: %s\n", state.String())
	})
	return pc
}

func (pc *PeerConnection) onICECandidate() *PeerConnection {
	pc.peerConnection.OnICECandidate(func(candidate *webrtc.ICECandidate) {
		if candidate == nil {
			fmt.Println("ICE gathering complete")
			return
		}

		fmt.Printf("Found candidate: %s (type=%s)\n", candidate.String(), candidate.Typ)
	})
	return pc
}

func (pc *PeerConnection) CreateDataChannel(label string, options ...datachannel.Option) (*datachannel.DataChannel, error) {
	if pc.dataChannels == nil {
		return nil, errors.New("data channels are not enabled")
	}
	channel, err := pc.dataChannels.CreateDataChannel(label, pc.peerConnection, options...)
	if err != nil {
		return nil, err
	}

	return channel, nil
}

func (pc *PeerConnection) CreateMediaSource(label string, options ...mediasource.TrackOption) (*mediasource.Track, error) {
	if pc.tracks == nil {
		return nil, errors.New("media source are not enabled")
	}

	track, err := pc.tracks.CreateTrack(label, pc.peerConnection, options...)
	if err != nil {
		return nil, err
	}

	return track, nil
}

func (pc *PeerConnection) CreateRTPMediaSource(label string, options ...mediasource.TrackOption) (*mediasource.RTPTrack, error) {
	if pc.tracks == nil {
		return nil, errors.New("media source are not enabled")
	}

	track, err := pc.tracks.CreateRTPTrack(label, pc.peerConnection, options...)
	if err != nil {
		return nil, err
	}

	return track, nil
}

func (pc *PeerConnection) CreateMediaSink(label string, options ...mediasink.SinkOption) (*mediasink.Sink, error) {
	if pc.sinks == nil {
		return nil, errors.New("media sinks are not enabled")
	}

	sink, err := pc.sinks.CreateSink(label, options...)
	if err != nil {
		return nil, err
	}

	return sink, nil
}

func (pc *PeerConnection) GetBWEstimator() (*BWEController, error) {
	if pc.bwc == nil {
		return nil, errors.New("bitrate control is not enabled")
	}

	if pc.bwc.get() == nil {
		return nil, errors.New("bitrate estimator not yet assigned")
	}

	return pc.bwc, nil
}

func (pc *PeerConnection) Close() error {
	var merr error
	pc.once.Do(func() {
		if pc.cancel != nil {
			pc.cancel()
		}

		if err := pc.peerConnection.Close(); err != nil {
			merr = multierr.Append(merr, err)
			return
		}

		if pc.bwc != nil {
			pc.bwc.Close()
		}
	})

	return merr
}
