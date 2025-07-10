//go:build cgo_enabled

package client

import "errors"

func (pc *PeerConnection) GetBWEstimator() (*BWEController, error) {
	if pc.bwController == nil || pc.bwController.estimator == nil {
		return nil, errors.New("bitrate control is not enabled")
	}

	return pc.bwController, nil
}
