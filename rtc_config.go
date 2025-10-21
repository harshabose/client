package client

import (
	"os"

	"github.com/pion/webrtc/v4"
)

func GetFullRTCConfiguration() webrtc.Configuration {
	return webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{os.Getenv("STUN_SERVER_URL")},
			},
			{
				URLs:           []string{os.Getenv("TURN_UDP_SERVER_URL")},
				Username:       os.Getenv("TURN_SERVER_USERNAME"),
				Credential:     os.Getenv("TURN_SERVER_PASSWORD"),
				CredentialType: webrtc.ICECredentialTypePassword,
			},
			{
				URLs:           []string{os.Getenv("TURN_TCP_SERVER_URL")},
				Username:       os.Getenv("TURN_SERVER_USERNAME"),
				Credential:     os.Getenv("TURN_SERVER_PASSWORD"),
				CredentialType: webrtc.ICECredentialTypePassword,
			},
			{
				URLs:           []string{os.Getenv("TURN_TLS_SERVER_URL")},
				Username:       os.Getenv("TURN_SERVER_USERNAME"),
				Credential:     os.Getenv("TURN_SERVER_PASSWORD"),
				CredentialType: webrtc.ICECredentialTypePassword,
			},
		},
	}
}

func GetSTUNOnlyRTCConfiguration() webrtc.Configuration {
	return webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{os.Getenv("STUN_SERVER_URL")},
			},
		},
	}
}
