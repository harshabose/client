package client

import (
	"context"
	"errors"

	"github.com/pion/interceptor"
	"github.com/pion/interceptor/pkg/cc"
	"github.com/pion/webrtc/v4"

	"github.com/harshabose/tools/pkg/multierr"
)

type Client struct {
	peerConnections     map[string]*PeerConnection
	mediaEngine         *webrtc.MediaEngine
	settingsEngine      *webrtc.SettingEngine
	interceptorRegistry *interceptor.Registry
	estimatorChan       chan cc.BandwidthEstimator
	api                 *webrtc.API
	ctx                 context.Context
	cancel              context.CancelFunc
}

func NewClient(
	ctx context.Context, cancel context.CancelFunc,
	mediaEngine *webrtc.MediaEngine, interceptorRegistry *interceptor.Registry,
	settings *webrtc.SettingEngine, options ...ClientOption,
) (*Client, error) {
	if mediaEngine == nil {
		mediaEngine = &webrtc.MediaEngine{}
	}
	if interceptorRegistry == nil {
		interceptorRegistry = &interceptor.Registry{}
	}
	if settings == nil {
		settings = &webrtc.SettingEngine{}
	}

	settings.DetachDataChannels()

	peerConnections := &Client{
		mediaEngine:         mediaEngine,
		interceptorRegistry: interceptorRegistry,
		settingsEngine:      settings,
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

	peerConnections.api = webrtc.NewAPI(webrtc.WithMediaEngine(peerConnections.mediaEngine), webrtc.WithInterceptorRegistry(peerConnections.interceptorRegistry), webrtc.WithSettingEngine(*peerConnections.settingsEngine))

	return peerConnections, nil
}

func NewClientFromConfig(
	ctx context.Context, cancel context.CancelFunc,
	mediaEngine *webrtc.MediaEngine, interceptorRegistry *interceptor.Registry,
	settings *webrtc.SettingEngine, config *ClientConfig,
) (*Client, error) {
	return NewClient(
		ctx, cancel, mediaEngine, interceptorRegistry, settings,
		config.ToOptions()..., // The magic spread operator!
	)
}

func (client *Client) CreatePeerConnection(label string, config webrtc.Configuration, options ...PeerConnectionOption) (*PeerConnection, error) {
	var err error

	if _, exists := client.peerConnections[label]; exists {
		return nil, errors.New("peer connection already exists")
	}

	// TODO: CHANGE THE SIGNATURE; SENDING A CANCEL FUNC IS IDIOTIC
	if client.peerConnections[label], err = CreatePeerConnection(client.ctx, client.cancel, label, client.api, config, options...); err != nil {
		return nil, err
	}

	return client.peerConnections[label], nil
}

func (client *Client) CreatePeerConnectionFromConfig(config PeerConnectionConfig) (*PeerConnection, error) {
	pc, err := client.CreatePeerConnection(config.Name, config.RTCConfig, config.ToOptions()...)
	if err != nil {
		return nil, err
	}

	if err := config.CreateDataChannels(pc); err != nil {
		return nil, err
	}

	if err := config.CreateMediaSources(pc); err != nil {
		return nil, err
	}

	if err := config.CreateMediaSinks(pc); err != nil {
		return nil, err
	}

	return pc, nil
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

func (client *Client) Connect(category string) error {
	var merr error
	for _, pc := range client.peerConnections {
		if err := pc.Connect(category); err != nil {
			merr = multierr.Append(merr, err)
		}
	}

	return merr
}

func (client *Client) ClosePeerConnection(label string) error {
	pc, err := client.GetPeerConnection(label)
	if err != nil {
		return err
	}

	return pc.Close()
}

func (client *Client) Close() error {
	for _, peerConnection := range client.peerConnections {
		if err := peerConnection.Close(); err != nil {
			return err
		}
	}

	return nil
}
