package datachannel

import (
	"context"
	"errors"
	"fmt"

	"github.com/pion/webrtc/v4"
)

type DataChannel struct {
	label       string
	datachannel *webrtc.DataChannel
	init        *webrtc.DataChannelInit
	ctx         context.Context
}

func CreateDataChannel(ctx context.Context, label string, peerConnection *webrtc.PeerConnection, options ...Option) (*DataChannel, error) {
	dc := &DataChannel{
		label:       label,
		datachannel: nil,
		ctx:         ctx,
	}

	for _, option := range options {
		if err := option(dc); err != nil {
			return nil, err
		}
	}

	datachannel, err := peerConnection.CreateDataChannel(label, dc.init)
	if err != nil {
		return nil, err
	}

	dc.datachannel = datachannel

	return dc.onOpen().onClose(), nil
}

func CreateRawDataChannel(ctx context.Context, channel *webrtc.DataChannel) (*DataChannel, error) {
	dataChannel := &DataChannel{
		label:       channel.Label(),
		datachannel: channel,
		ctx:         ctx,
	}

	return dataChannel.onOpen().onClose(), nil
}

func (dataChannel *DataChannel) GetLabel() string {
	return dataChannel.label
}

func (dataChannel *DataChannel) Close() error {
	if err := dataChannel.datachannel.Close(); err != nil {
		return err
	}

	return nil
}

func (dataChannel *DataChannel) onOpen() *DataChannel {
	dataChannel.datachannel.OnOpen(func() {
		fmt.Printf("dataChannel Open with Label: %s\n", dataChannel.datachannel.Label())
	})
	return dataChannel
}

func (dataChannel *DataChannel) onClose() *DataChannel {
	dataChannel.datachannel.OnClose(func() {
		fmt.Printf("dataChannel Closed with Label: %s\n", dataChannel.datachannel.Label())
	})
	return dataChannel
}

func (dataChannel *DataChannel) DataChannel() *webrtc.DataChannel {
	return dataChannel.datachannel
}

// +++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++

type DataChannels struct {
	datachannel map[string]*DataChannel
	ctx         context.Context
}

func CreateDataChannels(ctx context.Context) *DataChannels {
	return &DataChannels{
		datachannel: map[string]*DataChannel{},
		ctx:         ctx,
	}
}

func (dataChannels *DataChannels) CreateDataChannel(label string, peerConnection *webrtc.PeerConnection, options ...Option) (*DataChannel, error) {
	if _, exits := dataChannels.datachannel[label]; exits {
		return nil, fmt.Errorf("datachannel with id = '%s' already exists", label)
	}

	channel, err := CreateDataChannel(dataChannels.ctx, label, peerConnection, options...)
	if err != nil {
		return nil, err
	}

	dataChannels.datachannel[label] = channel
	return channel, nil
}

func (dataChannels *DataChannels) CreateRawDataChannel(channel *webrtc.DataChannel) (*DataChannel, error) {
	_, exists := dataChannels.datachannel[channel.Label()]
	if exists {
		return nil, fmt.Errorf("data channel already exists with label: %s", channel.Label())
	}

	dataChannel, err := CreateRawDataChannel(dataChannels.ctx, channel)
	if err != nil {
		return nil, err
	}

	dataChannels.datachannel[channel.Label()] = dataChannel

	return dataChannel, nil
}

func (dataChannels *DataChannels) GetDataChannel(label string) (*DataChannel, error) {
	dataChannel, exists := dataChannels.datachannel[label]
	if !exists {
		return nil, errors.New("datachannel does not exists")
	}
	return dataChannel, nil
}

func (dataChannels *DataChannels) Close(label string) (err error) {
	if err = dataChannels.datachannel[label].Close(); err == nil {
		return nil
	}
	delete(dataChannels.datachannel, label)
	return err
}
