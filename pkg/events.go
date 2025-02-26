package client

type Event = func(*PeerConnections) error
