package client

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

type BaseSignal interface {
	Connect(string, string) error
	Close() error
}
