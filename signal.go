package client

import "context"

const (
	FieldOffer           = "offer"
	FieldAnswer          = "answer"
	FieldSDP             = "sdp"
	FieldUpdatedAt       = "updated-at"
	FieldStatus          = "status"
	FieldStatusPending   = "pending"
	FieldStatusConnected = "connected"
	FieldCreatedAt       = "created-at"
)

type (
	BaseSignal interface {
		Connect(string, *PeerConnection) error
		Close() error
	}
	ForOffer  func(ctx context.Context) (string, error)
	OnAnswer  func(ctx context.Context, sdp string) error
	OnOffer   func(ctx context.Context, sdp string) error
	ForAnswer func(ctx context.Context) (string, error)
)
