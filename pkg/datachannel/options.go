package datachannel

import "github.com/pion/webrtc/v4"

type Option = func(*DataChannel) error

func WithDataChannelInit(init *webrtc.DataChannelInit) Option {
	return func(channel *DataChannel) error {
		channel.init = init
		return nil
	}
}

var (
	OrderedTrue              = true
	MaxRetransmits    uint16 = 2  // either MaxRetransmits or MaxPacketLifeTime can be specified at once
	MaxPacketLifeTime uint16 = 50 // milliseconds
	Protocol                 = "binary"
	NegotiatedTrue           = true
	IDOne             uint16 = 1
)
