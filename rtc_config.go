package client

import (
	"os"

	"github.com/pion/webrtc/v4"
)

func GetRTCConfiguration() webrtc.Configuration {
	return webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{os.Getenv("STUN_SERVER_URL")},
			},
			// {
			// 	URLs:           []string{os.Getenv("TURN_UDP_SERVER_URL")},
			// 	Username:       os.Getenv("TURN_SERVER_USERNAME"),
			// 	Credential:     os.Getenv("TURN_SERVER_PASSWORD"),
			// 	CredentialType: webrtc.ICECredentialTypePassword,
			// },
			// {
			// 	URLs:           []string{os.Getenv("TURN_TCP_SERVER_URL")},
			// 	Username:       os.Getenv("TURN_SERVER_USERNAME"),
			// 	Credential:     os.Getenv("TURN_SERVER_PASSWORD"),
			// 	CredentialType: webrtc.ICECredentialTypePassword,
			// },
			// {
			// 	URLs:           []string{os.Getenv("TURN_TLS_SERVER_URL")},
			// 	Username:       os.Getenv("TURN_SERVER_USERNAME"),
			// 	Credential:     os.Getenv("TURN_SERVER_PASSWORD"),
			// 	CredentialType: webrtc.ICECredentialTypePassword,
			// },
		},
		ICETransportPolicy: webrtc.ICETransportPolicyAll,
		BundlePolicy:       webrtc.BundlePolicyMaxCompat,
		RTCPMuxPolicy:      webrtc.RTCPMuxPolicyRequire,
		SDPSemantics:       webrtc.SDPSemanticsUnifiedPlan,
	}
}
