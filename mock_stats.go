package client

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"math"
	"time"

	"github.com/pion/webrtc/v4"
)

type MockStatsGetter struct{}

func NewMockStatsGetter() *MockStatsGetter {
	fmt.Println("Using MOCK Stats Getter...")
	return &MockStatsGetter{}
}

func (stats *MockStatsGetter) Generate(_ *PeerConnection) Stat {
	now := time.Now()
	timestamp := webrtc.StatsTimestamp(now.UnixNano() / int64(time.Microsecond))

	return Stat{
		PeerConnectionStat: webrtc.PeerConnectionStats{
			Timestamp:          timestamp,
			Type:               webrtc.StatsTypePeerConnection,
			ID:                 "peer_connection_1",
			DataChannelsOpened: 2,
			DataChannelsClosed: 0,
		},

		ICECandidatePairStat: webrtc.ICECandidatePairStats{
			Timestamp:                   timestamp,
			Type:                        webrtc.StatsTypeCandidatePair,
			ID:                          "candidate_pair_1",
			TransportID:                 "ice_transport_1",
			LocalCandidateID:            "local_candidate_1",
			RemoteCandidateID:           "remote_candidate_1",
			State:                       webrtc.StatsICECandidatePairStateSucceeded,
			Nominated:                   true,
			PacketsSent:                 15420,
			PacketsReceived:             14890,
			BytesSent:                   2048576, // ~2MB
			BytesReceived:               1953792, // ~1.9MB
			LastPacketSentTimestamp:     timestamp - 100,
			LastPacketReceivedTimestamp: timestamp - 150,
			TotalRoundTripTime:          0.045,   // 45ms
			CurrentRoundTripTime:        0.048,   // 48ms
			AvailableOutgoingBitrate:    1500000, // 1.5Mbps
			AvailableIncomingBitrate:    1200000, // 1.2Mbps
			RequestsReceived:            1,
			RequestsSent:                1,
			ResponsesReceived:           1,
			ResponsesSent:               1,
		},

		// ICECandidateLocalStat: webrtc.ICECandidateStats{
		// 	Timestamp:     timestamp,
		// 	Type:          webrtc.StatsTypeLocalCandidate,
		// 	ID:            "local_candidate_1",
		// 	IP:            "192.168.1.100",
		// 	Port:          54321,
		// 	Protocol:      "udp",
		// 	CandidateType: webrtc.ICECandidateTypeHost,
		// 	Priority:      2113667326,
		// 	URL:           "",
		// 	Deleted:       false,
		// },
		//
		// ICECandidateRemoteStat: webrtc.ICECandidateStats{
		// 	Timestamp:     timestamp,
		// 	Type:          webrtc.StatsTypeRemoteCandidate,
		// 	ID:            "remote_candidate_1",
		// 	IP:            "203.0.113.45",
		// 	Port:          12345,
		// 	Protocol:      "udp",
		// 	CandidateType: webrtc.ICECandidateTypePrflx,
		// 	Priority:      1685987326,
		// 	URL:           "stun:stun.example.com:19302",
		// },

		CertificateStats: map[string]webrtc.CertificateStats{
			"cert_1": {
				Timestamp:            timestamp,
				Type:                 webrtc.StatsTypeCertificate,
				ID:                   "cert_1",
				Fingerprint:          "A1:B2:C3:D4:E5:F6:G7:H8:I9:J0:K1:L2:M3:N4:O5:P6:Q7:R8:S9:T0",
				FingerprintAlgorithm: "sha-256",
				Base64Certificate:    generateFakeBase64Cert(),
				IssuerCertificateID:  "self-signed",
			},
		},

		CodecStats: map[string]webrtc.CodecStats{
			"codec_h264": {
				Timestamp:   timestamp,
				Type:        webrtc.StatsTypeCodec,
				ID:          "codec_h264",
				PayloadType: 96,
				MimeType:    "video/H264",
				ClockRate:   90000,
				Channels:    0,
				SDPFmtpLine: "profile-level-id=42e01f;packetization-mode=1",
			},
			"codec_opus": {
				Timestamp:   timestamp,
				Type:        webrtc.StatsTypeCodec,
				ID:          "codec_opus",
				PayloadType: 111,
				MimeType:    "audio/opus",
				ClockRate:   48000,
				Channels:    2,
				SDPFmtpLine: "minptime=10;useinbandfec=1",
			},
			"codec_vp9": {
				Timestamp:   timestamp,
				Type:        webrtc.StatsTypeCodec,
				ID:          "codec_vp9",
				PayloadType: 98,
				MimeType:    "video/VP9",
				ClockRate:   90000,
				Channels:    0,
				SDPFmtpLine: "profile-id=0",
			},
		},

		ICETransportStat: webrtc.TransportStats{
			Timestamp:               timestamp,
			Type:                    webrtc.StatsTypeTransport,
			ID:                      "ice_transport_1",
			PacketsSent:             15420,
			PacketsReceived:         14890,
			BytesSent:               2048576, // ~2MB
			BytesReceived:           1953792, // ~1.9MB
			ICERole:                 webrtc.ICERoleControlling,
			DTLSState:               webrtc.DTLSTransportStateConnected,
			ICEState:                webrtc.ICETransportStateConnected,
			SelectedCandidatePairID: "candidate_pair_1",
			LocalCertificateID:      "cert_1",
			RemoteCertificateID:     "remote_cert_1",
			DTLSCipher:              "TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256",
			SRTPCipher:              "AES_CM_128_HMAC_SHA1_80",
		},

		SCTPTransportStat: webrtc.SCTPTransportStats{
			Timestamp:             timestamp,
			Type:                  webrtc.StatsTypeSCTPTransport,
			ID:                    "sctp_transport_1",
			TransportID:           "ice_transport_1",
			SmoothedRoundTripTime: 0.048,  // 48ms
			CongestionWindow:      65536,  // 64KB
			ReceiverWindow:        131072, // 128KB
			MTU:                   1200,
			UNACKData:             0,
			BytesSent:             524288, // 512KB
			BytesReceived:         487424, // ~475KB
		},

		DataChannelStats: map[string]webrtc.DataChannelStats{
			"dc_chat": {
				Timestamp:             timestamp,
				Type:                  webrtc.StatsTypeDataChannel,
				ID:                    "dc_chat",
				Label:                 "chat",
				Protocol:              "",
				DataChannelIdentifier: 0,
				TransportID:           "sctp_transport_1",
				State:                 webrtc.DataChannelStateOpen,
				MessagesSent:          156,
				BytesSent:             45120, // ~44KB
				MessagesReceived:      142,
				BytesReceived:         38976, // ~38KB
			},
			"dc_file_transfer": {
				Timestamp:             timestamp,
				Type:                  webrtc.StatsTypeDataChannel,
				ID:                    "dc_file_transfer",
				Label:                 "file-transfer",
				Protocol:              "",
				DataChannelIdentifier: 1,
				TransportID:           "sctp_transport_1",
				State:                 webrtc.DataChannelStateOpen,
				MessagesSent:          89,
				BytesSent:             479168, // ~467KB
				MessagesReceived:      76,
				BytesReceived:         448448, // ~437KB
			},
		},
	}
}

// Helper function to generate a fake base64 certificate
func generateFakeBase64Cert() string {
	// Generate some random bytes for a fake certificate
	certBytes := make([]byte, 256)
	_, _ = rand.Read(certBytes)
	return base64.StdEncoding.EncodeToString(certBytes)
}

// Helper function to simulate realistic network variations
func addNetworkJitter(baseValue float64, jitterPercent float64) float64 {
	jitter := (math.Sin(float64(time.Now().UnixNano())/1e9) * jitterPercent * baseValue) / 100
	return math.Max(0, baseValue+jitter)
}

// func (stats *MockStatsGetter) GenerateWithConditions(pc *PeerConnection, condition string) (Stat, error) {
// 	baseStat, _ := stats.Generate(pc)
//
// 	switch condition {
// 	case "poor_network":
// 		// Simulate poor network conditions
// 		baseStat.ICECandidatePairStat.CurrentRoundTripTime = addNetworkJitter(0.250, 20)                                  // 250ms ± 20%
// 		baseStat.ICECandidatePairStat.PacketsReceived = uint32(float64(baseStat.ICECandidatePairStat.PacketsSent) * 0.85) // 15% loss
// 		baseStat.ICETransportStat.BytesReceived = uint64(float64(baseStat.ICETransportStat.BytesSent) * 0.85)
// 		baseStat.SCTPTransportStat.CongestionWindow = 16384 // 16KB (small window)
// 		baseStat.SCTPTransportStat.UNACKData = 8192         // Some unacknowledged data
//
// 	case "excellent_network":
// 		// Simulate excellent network conditions
// 		baseStat.ICECandidatePairStat.CurrentRoundTripTime = addNetworkJitter(0.015, 5)           // 15ms ± 5%
// 		baseStat.ICECandidatePairStat.PacketsReceived = baseStat.ICECandidatePairStat.PacketsSent // No loss
// 		baseStat.ICETransportStat.BytesReceived = baseStat.ICETransportStat.BytesSent
// 		baseStat.SCTPTransportStat.CongestionWindow = 262144 // 256KB (large window)
// 		baseStat.SCTPTransportStat.UNACKData = 0             // All data acknowledged
//
// 	case "connecting":
// 		// Simulate connection in progress
// 		baseStat.ICECandidatePairStat.State = webrtc.StatsICECandidatePairStateInProgress
// 		baseStat.ICETransportStat.ICEState = webrtc.ICETransportStateChecking
// 		baseStat.ICETransportStat.DTLSState = webrtc.DTLSTransportStateConnecting
// 		baseStat.DataChannelStats["dc_chat"].State = webrtc.DataChannelStateConnecting
// 		baseStat.DataChannelStats["dc_file_transfer"].State = webrtc.DataChannelStateConnecting
// 	}
//
// 	return baseStat, nil
// }

func (stats *MockStatsGetter) Close() error {
	return nil
}
