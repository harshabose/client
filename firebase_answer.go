package client

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	"firebase.google.com/go"
	"google.golang.org/api/option"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/pion/webrtc/v4"
)

type AnswerSignal struct {
	app    *firebase.App
	client *firestore.Client
	docRef *firestore.DocumentRef
	ctx    context.Context
}

func CreateFirebaseAnswerSignal(ctx context.Context) (*AnswerSignal, error) {
	var (
		configuration option.ClientOption
		app           *firebase.App
		client        *firestore.Client
		err           error
	)

	if configuration, err = GetFirebaseConfiguration(); err != nil {
		return nil, err
	}
	if app, err = firebase.NewApp(ctx, nil, configuration); err != nil {
		return nil, err
	}
	if client, err = app.Firestore(ctx); err != nil {
		return nil, err
	}

	return &AnswerSignal{
		app:    app,
		client: client,
		ctx:    ctx,
	}, nil
}

func (signal *AnswerSignal) Connect(category string, pc *PeerConnection) error {
	signal.docRef = signal.client.Collection(category).Doc(pc.label)

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	var data map[string]interface{}
loop:
	for {
		select {
		case <-signal.ctx.Done():
			return fmt.Errorf("failed to get offer within context deadline; err: %w", signal.ctx.Err())
		case <-ticker.C:
			snapshot, err := signal.docRef.Get(signal.ctx)
			if err != nil || snapshot == nil {
				if status.Code(err) == codes.NotFound {
					continue loop
				}
			}
			data = snapshot.Data()

			if currentStatus, exists := data[FieldStatus]; !exists || currentStatus != FieldStatusPending {
				continue loop
			}

			break loop
		}
	}
	fmt.Println("Found Offer. Creating answer...")

	return signal.answer(data[FieldOffer].(map[string]interface{}), pc)
}

func (signal *AnswerSignal) answer(offer map[string]interface{}, pc *PeerConnection) error {
	sdp, ok := offer[FieldSDP].(string)
	if !ok {
		return fmt.Errorf("invalid SDP format in offer")
	}

	if err := pc.peerConnection.SetRemoteDescription(webrtc.SessionDescription{
		Type: webrtc.SDPTypeOffer,
		SDP:  sdp,
	}); err != nil {
		return err
	}

	answer, err := pc.peerConnection.CreateAnswer(nil)
	if err != nil {
		return fmt.Errorf("error while creating answer: %w", err)
	}

	if err := pc.peerConnection.SetLocalDescription(answer); err != nil {
		return fmt.Errorf("error while setting local sdp: %w", err)
	}

	select {
	case <-webrtc.GatheringCompletePromise(pc.peerConnection):
	case <-signal.ctx.Done():
		return fmt.Errorf("failed to gather ICE candidates within context deadline; err: %w", signal.ctx.Err())
	}

	if _, err = signal.docRef.Set(signal.ctx, map[string]interface{}{
		FieldAnswer: map[string]interface{}{
			FieldSDP:       pc.peerConnection.LocalDescription().SDP,
			FieldUpdatedAt: firestore.ServerTimestamp,
		},
		FieldStatus: FieldStatusConnected,
	}, firestore.MergeAll); err != nil {
		return err
	}

	return nil
}

func (signal *AnswerSignal) Close() error {
	return signal.client.Close()
}
