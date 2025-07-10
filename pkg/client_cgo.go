//go:build cgo_enabled

package client

import (
	"errors"
	"time"

	"github.com/pion/webrtc/v4"
)

func (client *Client) CreatePeerConnectionWithBWEstimator(label string, config webrtc.Configuration, options ...PeerConnectionOption) (*PeerConnection, error) {
	var err error

	if _, exists := client.peerConnections[label]; exists {
		return nil, errors.New("peer connection already exists")
	}

	// TODO: CHANGE THE SIGNATURE; SENDING A CANCEL FUNC IS IDIOTIC
	if client.peerConnections[label], err = CreatePeerConnection(client.ctx, client.cancel, label, client.api, config, options...); err != nil {
		return nil, err
	}

	// TODO: THIS WEIRD CHANNEL BASED APPROACH OF SETTING BW CONTROLLER IS REQUIRED BECAUSE OF THE
	// TODO: THE WEIRD DESIGN OF CC INTERCEPTOR IN PION. TRACK THE ISSUE WITH "https://github.com/pion/webrtc/issues/3053"
	if client.peerConnections[label].bwController != nil {
		select {
		case estimator := <-client.estimatorChan:
			client.peerConnections[label].bwController.estimator = estimator
			client.peerConnections[label].bwController.interval = 50 * time.Millisecond
		}
	}

	return client.peerConnections[label], nil
}
