package client

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/pion/webrtc/v4"
)

// FileOfferSignal implements BaseSignal interface for file-based signaling (offer side)
type FileOfferSignal struct {
	peerConnection *PeerConnection
	ctx            context.Context
	offerPath      string
	answerPath     string
}

// CreateFileOfferSignal creates a new FileOfferSignal
func CreateFileOfferSignal(ctx context.Context, peerConnection *PeerConnection, offerPath string, answerPath string) *FileOfferSignal {
	return &FileOfferSignal{
		peerConnection: peerConnection,
		ctx:            ctx,
		offerPath:      offerPath,
		answerPath:     answerPath,
	}
}

// Connect implements the BaseSignal interface
func (signal *FileOfferSignal) Connect(category, connectionLabel string) error {
	// Use category and connectionLabel to create unique filenames if needed
	if category != "" && connectionLabel != "" {
		signal.offerPath = filepath.Join(signal.offerPath, category, connectionLabel, "offer.txt")
		signal.answerPath = filepath.Join(signal.answerPath, category, connectionLabel, "answer.txt")

		// Ensure directory exists
		dir := filepath.Dir(signal.offerPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("error creating directory %s: %w", dir, err)
		}
	}

	// Remove existing answer file if it exists
	if _, err := os.Stat(signal.offerPath); err == nil {
		if err := os.Remove(signal.offerPath); err != nil {
			return fmt.Errorf("error removing existing offer file: %w", err)
		}
	}

	// Remove existing answer file if it exists
	if _, err := os.Stat(signal.answerPath); err == nil {
		if err := os.Remove(signal.answerPath); err != nil {
			return fmt.Errorf("error removing existing answer file: %w", err)
		}
	}

	// Create offer
	offer, err := signal.peerConnection.peerConnection.CreateOffer(nil)
	if err != nil {
		return fmt.Errorf("error creating offer: %w", err)
	}

	if err := signal.peerConnection.peerConnection.SetLocalDescription(offer); err != nil {
		return fmt.Errorf("error setting local description: %w", err)
	}

	// Wait for ICE gathering to complete
	timer := time.NewTicker(30 * time.Second)
	defer timer.Stop()

	select {
	case <-webrtc.GatheringCompletePromise(signal.peerConnection.peerConnection):
		fmt.Println("ICE Gathering complete")
	case <-timer.C:
		return errors.New("failed to gather ICE candidates within 30 seconds")
	}

	// Save offer to file
	if err := signal.saveSDPToFile(signal.peerConnection.peerConnection.LocalDescription(), signal.offerPath); err != nil {
		return fmt.Errorf("error saving offer to file: %w", err)
	}

	fmt.Printf("Offer saved to %s. Waiting for answer...\n", signal.offerPath)

	// Wait for answer file
	return signal.waitForAnswer()
}

// waitForAnswer waits for the answer file to appear and processes it
func (signal *FileOfferSignal) waitForAnswer() error {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	timeout := time.NewTimer(5 * time.Minute)
	defer timeout.Stop()

	for {
		select {
		case <-ticker.C:
			// Check if answer file exists
			if _, err := os.Stat(signal.answerPath); err == nil {
				// Load answer from file
				answer, err := signal.loadSDPFromFile(signal.answerPath)
				if err != nil {
					fmt.Printf("Error loading answer: %v, retrying...\n", err)
					continue
				}

				// SetInputOption remote description
				if err = signal.peerConnection.peerConnection.SetRemoteDescription(answer); err != nil {
					return fmt.Errorf("error setting remote description: %w", err)
				}

				fmt.Println("Answer processed successfully")
				return nil
			}
		case <-timeout.C:
			return errors.New("timeout waiting for answer file")
		case <-signal.ctx.Done():
			return errors.New("context canceled")
		}
	}
}

// Close implements the BaseSignal interface
func (signal *FileOfferSignal) Close() error {
	// Nothing to close for file-based signaling
	return nil
}

// saveSDPToFile saves a SessionDescription to a file
func (signal *FileOfferSignal) saveSDPToFile(sdp *webrtc.SessionDescription, filename string) error {
	// Ensure directory exists
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("error creating directory %s: %w", dir, err)
	}

	// Encode SDP
	b, err := json.Marshal(sdp)
	if err != nil {
		return err
	}
	encoded := base64.StdEncoding.EncodeToString(b)

	// Write to file
	if err := os.WriteFile(filename, []byte(encoded), 0644); err != nil {
		return err
	}

	fmt.Printf("SDP saved to %s (%d bytes)\n", filename, len(encoded))
	return nil
}

// loadSDPFromFile loads a SessionDescription from a file
func (signal *FileOfferSignal) loadSDPFromFile(filename string) (webrtc.SessionDescription, error) {
	var sdp webrtc.SessionDescription

	data, err := os.ReadFile(filename)
	if err != nil {
		return sdp, fmt.Errorf("error reading %s: %w", filename, err)
	}

	encoded := string(data)
	b, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return sdp, fmt.Errorf("base64 decode error: %w", err)
	}

	if err = json.Unmarshal(b, &sdp); err != nil {
		return sdp, fmt.Errorf("JSON unmarshal error: %w", err)
	}

	fmt.Printf("SDP loaded from %s (%d bytes)\n", filename, len(data))
	return sdp, nil
}
