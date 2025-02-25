package client

type Event = func(*PeerConnection) error
