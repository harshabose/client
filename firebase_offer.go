package client

import (
	"context"
	"errors"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	"firebase.google.com/go"
	"github.com/pion/webrtc/v4"
	"google.golang.org/api/option"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type OfferSignal struct {
	app            *firebase.App
	firebaseClient *firestore.Client
	docRef         *firestore.DocumentRef
	ctx            context.Context
}

func CreateFirebaseOfferSignal(ctx context.Context) *OfferSignal {
	var (
		configuration  option.ClientOption
		app            *firebase.App
		firebaseClient *firestore.Client
		err            error
	)

	// TODO: I DO NOT LIKE THE 'PANIC' HERE. REMOVE AND RETURN A ERROR
	if configuration, err = GetFirebaseConfiguration(); err != nil {
		panic(err)
	}
	if app, err = firebase.NewApp(ctx, nil, configuration); err != nil {
		panic(err)
	}
	if firebaseClient, err = app.Firestore(ctx); err != nil {
		panic(err)
	}

	return &OfferSignal{
		app:            app,
		firebaseClient: firebaseClient,
		ctx:            ctx,
	}
}

func (signal *OfferSignal) Connect(category string, pc *PeerConnection) error {
	signal.docRef = signal.firebaseClient.Collection(category).Doc(pc.label)
	_, err := signal.docRef.Get(signal.ctx)

	if err != nil && status.Code(err) != codes.NotFound {
		fmt.Println(status.Code(err))
		return err
	}

	if err == nil {
		if _, err := signal.docRef.Delete(signal.ctx); err != nil {
			return err
		}
	}

	offer, err := pc.peerConnection.CreateOffer(nil)
	if err != nil {
		return fmt.Errorf("error while creating offer: %w", err)
	}

	if err := pc.peerConnection.SetLocalDescription(offer); err != nil {
		return fmt.Errorf("error while setting local sdp: %w", err)
	}

	timer := time.NewTicker(30 * time.Second)
	defer timer.Stop()

	select {
	case <-webrtc.GatheringCompletePromise(pc.peerConnection):
		fmt.Println("ICE Gathering complete")
	case <-timer.C:
		return errors.New("failed to gather ICE candidates within 30 seconds")
	}

	if _, err = signal.docRef.Set(signal.ctx, map[string]interface{}{
		FieldOffer: map[string]interface{}{
			FieldCreatedAt: firestore.ServerTimestamp,
			FieldSDP:       pc.peerConnection.LocalDescription().SDP,
			FieldUpdatedAt: firestore.ServerTimestamp,
		},
		FieldStatus: FieldStatusPending,
	}, firestore.MergeAll); err != nil {
		return fmt.Errorf("error while setting data to firestore: %w", err)
	}

	fmt.Println("Offer updated in firestore. Waiting for peer connection...")

	return signal.offer(pc)
}

func (signal *OfferSignal) offer(pc *PeerConnection) error {

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

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

			answer, exists := snapshot.Data()[FieldAnswer].(map[string]interface{})
			if !exists {
				continue loop
			}

			sdp, ok := answer[FieldSDP].(string)
			if !ok {
				continue loop
			}
			if err = pc.peerConnection.SetRemoteDescription(webrtc.SessionDescription{
				Type: webrtc.SDPTypeAnswer,
				SDP:  sdp,
			}); err != nil {
				fmt.Printf("error while setting remote description: %s", err.Error())
				continue loop
			}

			break loop
		}
	}

	return nil
}

func (signal *OfferSignal) Close() error {
	if err := signal.firebaseClient.Close(); err != nil {
		fmt.Printf("failed to close firebase client with error: %v\n", err)
		return err
	}

	return nil
}
