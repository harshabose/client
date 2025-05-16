package client

import (
	"context"
	"io"

	"github.com/pion/interceptor"
	"github.com/pion/interceptor/pkg/cc"
	"github.com/pion/webrtc/v4"
)

type Client struct {
	peerConnections     map[string]interface{}
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

	c := &Client{
		mediaEngine:         mediaEngine,
		interceptorRegistry: interceptorRegistry,
		peerConnections:     make(map[string]interface{}),
		estimatorChan:       make(chan cc.BandwidthEstimator, 10),
		ctx:                 ctx,
		cancel:              cancel,
	}

	for _, option := range options {
		if err := option(c); err != nil {
			return nil, err
		}
	}

	c.api = webrtc.NewAPI(webrtc.WithMediaEngine(c.mediaEngine), webrtc.WithInterceptorRegistry(c.interceptorRegistry))

	return c, nil
}

func (client *Client) WaitUntilClosed() {
	<-client.ctx.Done()
}

func (client *Client) Close() error {
	for _, peerConnection := range client.peerConnections {
		if err := peerConnection.(io.Closer).Close(); err != nil {
			return err
		}
	}

	return nil
}
