package client

import (
	"errors"

	"github.com/harshabose/simple_webrtc_comm/client/internal/receiver"
)

func (client *Client) CreatePeerConnection(label string, options ...receiver.PeerConnectionOption) (*receiver.PeerConnection, error) {
	var err error

	if _, exists := client.peerConnections[label]; exists {
		return nil, errors.New("peer connection already exists")
	}

	if client.peerConnections[label], err = receiver.CreatePeerConnection(client.ctx, label, client.api, options...); err != nil {
		return nil, err
	}

	return client.peerConnections[label].(*receiver.PeerConnection), nil
}

func (client *Client) GetReceiverPeerConnection(label string) (*receiver.PeerConnection, error) {
	if _, exists := client.peerConnections[label]; !exists {
		return nil, errors.New("peer connection not found")
	}
	return client.peerConnections[label].(*receiver.PeerConnection), nil
}
