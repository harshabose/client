package client

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/pion/interceptor"
	"github.com/pion/interceptor/pkg/cc"
	"github.com/pion/webrtc/v4"
)

type Client struct {
	peerConnections     map[string]*PeerConnection
	mediaEngine         *webrtc.MediaEngine
	interceptorRegistry *interceptor.Registry
	estimatorChan       chan cc.BandwidthEstimator
	api                 *webrtc.API
	ctx                 context.Context
	cancel              context.CancelFunc
}

func CreateClient(ctx context.Context, cancel context.CancelFunc, mediaEngine *webrtc.MediaEngine, interceptorRegistry *interceptor.Registry, options ...ClientOption) (*Client, error) {
	if mediaEngine == nil {
		mediaEngine = &webrtc.MediaEngine{}
	}
	if interceptorRegistry == nil {
		interceptorRegistry = &interceptor.Registry{}
	}

	peerConnections := &Client{
		mediaEngine:         mediaEngine,
		interceptorRegistry: interceptorRegistry,
		peerConnections:     make(map[string]*PeerConnection),
		estimatorChan:       make(chan cc.BandwidthEstimator, 10),
		ctx:                 ctx,
		cancel:              cancel,
	}

	for _, option := range options {
		if err := option(peerConnections); err != nil {
			return nil, err
		}
	}

	peerConnections.api = webrtc.NewAPI(webrtc.WithMediaEngine(peerConnections.mediaEngine), webrtc.WithInterceptorRegistry(peerConnections.interceptorRegistry))

	return peerConnections, nil
}

func (client *Client) CreatePeerConnection(label string, options ...PeerConnectionOption) (*PeerConnection, error) {
	var err error

	if _, exists := client.peerConnections[label]; exists {
		return nil, errors.New("peer connection already exists")
	}

	if client.peerConnections[label], err = CreatePeerConnection(client.ctx, client.cancel, label, client.api, options...); err != nil {
		return nil, err
	}

	// TODO: THIS WEIRD CHANNEL BASED APPROACH OF SETTING BW CONTROLLER IS REQUIRED BECAUSE OF THE
	// TODO: THE WEIRD DESIGN OF CC INTERCEPTOR IN PION. TRACK THE ISSUE WITH "https://github.com/pion/webrtc/issues/3053"
	if client.peerConnections[label].bwController != nil {
		select {
		case estimator := <-client.estimatorChan:
			fmt.Printf("successfully set bwe estimator for %s peer connection\n", label)
			client.peerConnections[label].bwController.estimator = estimator
			client.peerConnections[label].bwController.interval = 500 * time.Millisecond
		}
	}

	return client.peerConnections[label], nil
}

func (client *Client) GetPeerConnection(label string) (*PeerConnection, error) {
	if _, exists := client.peerConnections[label]; !exists {
		return nil, errors.New("peer connection not found")
	}
	return client.peerConnections[label], nil
}

func (client *Client) WaitUntilClosed() {
	<-client.ctx.Done()
}

func (client *Client) Close() error {
	for _, peerConnection := range client.peerConnections {
		if err := peerConnection.Close(); err != nil {
			return err
		}
	}

	return nil
}
