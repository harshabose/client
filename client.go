package client

import (
	"context"
	"errors"
	"iter"
	"sync"
	"time"

	"github.com/pion/interceptor"
	"github.com/pion/interceptor/pkg/cc"
	"github.com/pion/interceptor/pkg/stats"
	"github.com/pion/webrtc/v4"

	"github.com/harshabose/tools/pkg/multierr"
)

type Client struct {
	pcs                 map[string]*PeerConnection
	mediaEngine         *webrtc.MediaEngine
	settingsEngine      *webrtc.SettingEngine
	interceptorRegistry *interceptor.Registry
	api                 *webrtc.API

	estimatorChan chan cc.BandwidthEstimator
	getterChan    chan stats.Getter

	mux sync.RWMutex
	ctx context.Context
}

func NewClient(
	ctx context.Context,
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

	c := &Client{
		mediaEngine:         mediaEngine,
		interceptorRegistry: interceptorRegistry,
		settingsEngine:      settings,
		pcs:                 make(map[string]*PeerConnection),
		estimatorChan:       make(chan cc.BandwidthEstimator, 10),
		ctx:                 ctx,
	}

	for _, option := range options {
		if err := option(c); err != nil {
			return nil, err
		}
	}

	c.api = webrtc.NewAPI(webrtc.WithMediaEngine(c.mediaEngine), webrtc.WithInterceptorRegistry(c.interceptorRegistry), webrtc.WithSettingEngine(*c.settingsEngine))

	return c, nil
}

func (c *Client) CreatePeerConnection(label string, config webrtc.Configuration) (*PeerConnection, error) {
	c.mux.Lock()
	defer c.mux.Unlock()

	if _, exists := c.pcs[label]; exists {
		return nil, errors.New("peer connection already exists")
	}

	pc, err := CreatePeerConnection(c.ctx, label, c.api, config)
	if err != nil {
		return nil, err
	}

	c.pcs[label] = pc

	return pc, nil
}

func (c *Client) CreatePeerConnectionWithBWEstimator(label string, config webrtc.Configuration) (*PeerConnection, error) {
	pc, err := c.CreatePeerConnection(label, config)
	if err != nil {
		return nil, err
	}

	// TODO: THIS WEIRD CHANNEL BASED APPROACH OF SETTING BW CONTROLLER IS REQUIRED BECAUSE OF THE
	// TODO: THE WEIRD DESIGN OF CC INTERCEPTOR IN PION. TRACK THE ISSUE WITH "https://github.com/pion/webrtc/issues/3053"
	if pc.bwc != nil {
		select {
		case estimator := <-c.estimatorChan:
			pc.bwc.estimator = estimator
			pc.bwc.interval = 50 * time.Millisecond
		}
	}

	return pc, nil
}

func (c *Client) GetPeerConnection(label string) (*PeerConnection, error) {
	c.mux.RLock()
	defer c.mux.RUnlock()

	if _, exists := c.pcs[label]; !exists {
		return nil, errors.New("peer connection not found")
	}
	return c.pcs[label], nil
}

func (c *Client) PeerConnections() iter.Seq2[string, *PeerConnection] {
	return func(yield func(string, *PeerConnection) bool) {
		c.mux.RLock()
		defer c.mux.RUnlock()

		for label, pc := range c.pcs {
			if !yield(label, pc) {
				return
			}
		}
	}
}

func (c *Client) Connect(category string, signal BaseSignal) error {
	var merr error

	for _, pc := range c.PeerConnections() {
		if err := signal.Connect(category, pc); err != nil {
			merr = multierr.Append(merr, err)
		}
	}

	return merr
}

func (c *Client) ClosePeerConnection(label string) error {
	pc, err := c.GetPeerConnection(label)
	if err != nil {
		return err
	}

	if err := pc.Close(); err != nil {
		return err
	}

	c.mux.Lock()
	defer c.mux.Unlock()

	delete(c.pcs, label)
	return nil
}

func (c *Client) Close() error {
	for _, pc := range c.PeerConnections() {
		if err := pc.Close(); err != nil {
			return err
		}
	}

	c.mux.Lock()
	defer c.mux.Unlock()

	c.pcs = make(map[string]*PeerConnection)
	return nil
}
