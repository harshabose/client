package signal

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"cloud.google.com/go/firestore"
	"firebase.google.com/go"
	"google.golang.org/api/option"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/harshabose/simple_webrtc_comm/client/internal/config"
)

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

func validateParams(category, name string) error {
	if category == "" || name == "" {
		return errors.New("invalid category or name")
	}
	return nil
}

type BaseSignal interface {
	Setup(ctx context.Context, category, name string, receiverIndex int, offer bool) error
	Close() error
}

type Signal struct {
	app    *firebase.App
	client *firestore.Client
	docRef *firestore.DocumentRef
}

func CreateSignal(ctx context.Context) *Signal {
	var (
		configuration option.ClientOption
		app           *firebase.App
		client        *firestore.Client
		err           error
	)

	if configuration, err = config.GetFirebaseConfiguration(); err != nil {
		panic(err)
	}
	if app, err = firebase.NewApp(ctx, nil, configuration); err != nil {
		panic(err)
	}
	if client, err = app.Firestore(ctx); err != nil {
		panic(err)
	}

	return &Signal{
		app:    app,
		client: client,
		docRef: nil,
	}
}

func (signal *Signal) Setup(ctx context.Context, category, name string, receiverIndex int, offer bool) error {
	if err := validateParams(category, name); err != nil {
		return fmt.Errorf("validation failed for category and name: %w", err)
	}
	signal.docRef = signal.client.Collection(category).Doc(name).Collection("sdp").Doc(strconv.Itoa(receiverIndex))

	if _, err := signal.docRef.Get(ctx); err != nil {
		if status.Code(err) == codes.NotFound {
			// Doc doesn't exist and this is an offer - create new
			if offer {
				if _, err := signal.docRef.Set(ctx, map[string]interface{}{
					FieldStatus:    FieldStatusPending,
					FieldCreatedAt: firestore.ServerTimestamp,
				}); err != nil {
					return fmt.Errorf("failed to create document: %w", err)
				}
			}
			return nil
		}
		return fmt.Errorf("failed to check document existence: %w", err)
	}

	// Document exists
	if offer {
		// Delete and create new for offer
		if _, err := signal.docRef.Delete(ctx); err != nil {
			return fmt.Errorf("failed to delete existing document: %w", err)
		}
		if _, err := signal.docRef.Set(ctx, map[string]interface{}{
			FieldStatus:    FieldStatusPending,
			FieldCreatedAt: firestore.ServerTimestamp,
		}); err != nil {
			return fmt.Errorf("failed to create document: %w", err)
		}
	}

	return nil
}

func (signal *Signal) Close() error {
	return signal.client.Close()
}
