package signal

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	"firebase.google.com/go"
	"google.golang.org/api/option"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/harshabose/simple_webrtc_comm/client/internal/config"

	"github.com/pion/webrtc/v4"
)

type AnswerSignal struct {
	peerConnection *webrtc.PeerConnection
	app            *firebase.App
	client         *firestore.Client
	docRef         *firestore.DocumentRef
	ctx            context.Context
}

func CreateAnswerSignal(ctx context.Context, peerConnection *webrtc.PeerConnection) *AnswerSignal {
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

	return &AnswerSignal{
		app:            app,
		client:         client,
		peerConnection: peerConnection,
		ctx:            ctx,
	}
}

func (signal *AnswerSignal) Connect(category, connectionLabel string) error {
	signal.docRef = signal.client.Collection(category).Doc(connectionLabel)

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	var data map[string]interface{}
loop:
	for {
		select {
		case <-ticker.C:
			snapshot, err := signal.docRef.Get(signal.ctx)
			if err != nil {
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

	return signal.answer(data[FieldOffer].(map[string]interface{}))
}

func (signal *AnswerSignal) answer(offer map[string]interface{}) error {
	sdp, ok := offer[FieldSDP].(string)
	if !ok {
		return fmt.Errorf("invalid SDP format in offer")
	}

	if err := signal.peerConnection.SetRemoteDescription(webrtc.SessionDescription{
		Type: webrtc.SDPTypeOffer,
		SDP:  sdp,
	}); err != nil {
		return err
	}

	gatherComplete := webrtc.GatheringCompletePromise(signal.peerConnection)

	answer, err := signal.peerConnection.CreateAnswer(nil)
	if err != nil {
		return fmt.Errorf("error while creating answer: %w", err)
	}
	if err := signal.peerConnection.SetLocalDescription(answer); err != nil {
		return fmt.Errorf("error while setting local sdp: %w", err)
	}
	<-gatherComplete

	if _, err = signal.docRef.Set(signal.ctx, map[string]interface{}{
		FieldAnswer: map[string]interface{}{
			FieldSDP:       signal.peerConnection.LocalDescription().SDP,
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
