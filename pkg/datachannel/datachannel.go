package datachannel

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"sync"

	"github.com/pion/webrtc/v4"

	"github.com/harshabose/tools/pkg/cond"
	"github.com/harshabose/tools/pkg/multierr"
)

type DataChannel struct {
	label       string
	datachannel *webrtc.DataChannel
	init        *webrtc.DataChannelInit
	cond        *cond.ContextCond
	ctx         context.Context
}

func CreateDataChannel(ctx context.Context, label string, peerConnection *webrtc.PeerConnection, options ...Option) (*DataChannel, error) {
	dc := &DataChannel{
		label:       label,
		datachannel: nil,
		cond:        cond.NewContextCond(&sync.Mutex{}),
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

func (dc *DataChannel) GetLabel() string {
	return dc.label
}

func (dc *DataChannel) Close() error {
	if err := dc.datachannel.Close(); err != nil {
		return err
	}

	return nil
}

func (dc *DataChannel) onOpen() *DataChannel {
	dc.datachannel.OnOpen(func() {
		dc.cond.Broadcast()
		fmt.Printf("data channel (id=%s) opened\n", dc.datachannel.Label())
	})

	return dc
}

func (dc *DataChannel) onClose() *DataChannel {
	dc.datachannel.OnClose(func() {
		fmt.Printf("data channel (id=%s) closed\n", dc.datachannel.Label())
	})
	return dc
}

func (dc *DataChannel) WaitTillOpen(ctx context.Context) error {
	dc.cond.L.Lock()
	defer dc.cond.L.Unlock()

	for dc.datachannel.ReadyState() != webrtc.DataChannelStateOpen {
		if err := dc.cond.Wait(ctx); err != nil {
			return err
		}
	}

	return nil
}

func (dc *DataChannel) DataChannel() *webrtc.DataChannel {
	return dc.datachannel
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

func (dataChannels *DataChannels) DataChannels() iter.Seq2[string, *DataChannel] {
	return func(yield func(string, *DataChannel) bool) {
		for id, channel := range dataChannels.datachannel {
			if !yield(id, channel) {
				return
			}
		}
	}
}

func (dataChannels *DataChannels) Close() error {
	var merr error
	for label, datachannel := range dataChannels.datachannel {
		if err := datachannel.Close(); err != nil {
			merr = multierr.Append(merr, err)
		}
		delete(dataChannels.datachannel, label)
	}
	return merr
}
