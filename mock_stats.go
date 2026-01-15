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

func (stats *MockStatsGetter) Close() error {
	return nil
}
