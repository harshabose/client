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
	"github.com/harshabose/tools/pkg/multierr"
)

type Stat struct {
	PeerConnectionStat     webrtc.PeerConnectionStats          `json:"peer_connection_stat"`
	ICECandidatePairStat   webrtc.ICECandidatePairStats        `json:"ice_candidate_pair_stat"`
	ICECandidateLocalStat  map[string]webrtc.ICECandidateStats `json:"ice_candidate_local_stat"`
	ICECandidateRemoteStat map[string]webrtc.ICECandidateStats `json:"ice_candidate_remote_stat"`
	CertificateStats       map[string]webrtc.CertificateStats  `json:"certificate_stats"`
	CodecStats             map[string]webrtc.CodecStats        `json:"codec_stats"`
	ICETransportStat       webrtc.TransportStats               `json:"ice_transport_stat"`
	SCTPTransportStat      webrtc.SCTPTransportStats           `json:"sctp_transport_stat"`
	DataChannelStats       map[string]webrtc.DataChannelStats  `json:"data_channel_stats"`
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
			ICECandidateLocalStat:  make(map[string]webrtc.ICECandidateStats),
			ICECandidateRemoteStat: make(map[string]webrtc.ICECandidateStats),
			CertificateStats:       make(map[string]webrtc.CertificateStats),
			CodecStats:             make(map[string]webrtc.CodecStats),
		},
	}
}

func (s *stat) Consume(stats webrtc.Stats) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	switch stat := stats.(type) {
	case webrtc.DataChannelStats:
		_, err := s.pc.GetDataChannel(stat.Label)
		if err != nil {
			return err
		}

		s.DataChannelStats[stat.Label] = stat
		return nil

	case webrtc.PeerConnectionStats:
		s.Stat.PeerConnectionStat = stat
		return nil

	case webrtc.ICECandidateStats:
		if stat.Type == webrtc.StatsTypeLocalCandidate {
			s.ICECandidateLocalStat[stat.ID] = stat
			return nil
		}

		if stat.Type == webrtc.StatsTypeRemoteCandidate {
			s.ICECandidateRemoteStat[stat.ID] = stat
			return nil
		}

		return errors.New("ICE candidate stat is neither local or remote")

	case webrtc.ICECandidatePairStats:
		s.ICECandidatePairStat = stat
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
		return errors.New("stat type is not managed")
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
		PeerConnectionStat:     s.Stat.PeerConnectionStat,
		ICECandidatePairStat:   s.ICECandidatePairStat,
		ICECandidateLocalStat:  s.ICECandidateLocalStat,
		ICECandidateRemoteStat: s.ICECandidateRemoteStat,
		CertificateStats:       certificatesCopy,
		CodecStats:             codecCopy,
		ICETransportStat:       s.ICETransportStat,
		SCTPTransportStat:      s.SCTPTransportStat,
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
	}

	pc.stat = newStat(pc)

	return pc.onConnectionStateChangeEvent().onICEConnectionStateChange().onICEGatheringStateChange().onICECandidate(), err
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

		if state == webrtc.PeerConnectionStateDisconnected || state == webrtc.PeerConnectionStateFailed {
			fmt.Printf("closing peer connection (id=%s)\n", pc.label)
			if err := pc.Close(); err != nil {
				fmt.Printf("error while closing peer connection (id=%s); err=%v\n", pc.label, err)
			}
		}
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

	if pc.bwc.estimator == nil {
		return nil, errors.New("bitrate estimator not yet assigned")
	}

	return pc.bwc, nil
}

func (pc *PeerConnection) Close() error {
	fmt.Printf("[PeerConnection] Starting close for peer connection: %s\n", pc.label)

	var merr error
	pc.once.Do(func() {
		fmt.Printf("[PeerConnection] Canceling context for peer: %s\n", pc.label)
		if pc.cancel != nil {
			pc.cancel()
		}

		fmt.Printf("[PeerConnection] Closing underlying WebRTC peer connection for: %s\n", pc.label)
		if err := pc.peerConnection.Close(); err != nil {
			fmt.Printf("[PeerConnection] ERROR: Failed to close WebRTC peer connection for %s: %v\n", pc.label, err)
			merr = multierr.Append(merr, err)
		} else {
			fmt.Printf("[PeerConnection] ✓ WebRTC peer connection closed for: %s\n", pc.label)
		}

		if pc.bwc != nil {
			fmt.Printf("[PeerConnection] Closing bandwidth controller for: %s\n", pc.label)
			if err := pc.bwc.Close(); err != nil {
				fmt.Printf("[PeerConnection] ERROR: Failed to close bandwidth controller for %s: %v\n", pc.label, err)
				merr = multierr.Append(merr, err)
			} else {
				fmt.Printf("[PeerConnection] ✓ Bandwidth controller closed for: %s\n", pc.label)
			}
		}

		if merr == nil {
			fmt.Printf("[PeerConnection] ✓ Successfully closed peer connection: %s\n", pc.label)
		} else {
			fmt.Printf("[PeerConnection] ⚠ Peer connection closed with errors for %s: %v\n", pc.label, merr)
		}
	})

	return merr
}
