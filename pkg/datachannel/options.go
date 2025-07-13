package datachannel

import "github.com/pion/webrtc/v4"

type Option = func(*DataChannel) error

func WithDataChannelInit(init *webrtc.DataChannelInit) Option {
	return func(channel *DataChannel) error {
		channel.init = init
		return nil
	}
}
