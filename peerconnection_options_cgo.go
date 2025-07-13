//go:build cgo_enabled

package client

func WithBandwidthControl() PeerConnectionOption {
	return func(connection *PeerConnection) error {
		connection.bwController = createBWController(connection.ctx)
		return nil
	}
}
