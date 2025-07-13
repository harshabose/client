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

// FileAnswerSignal implements BaseSignal interface for file-based signaling (answer side)
type FileAnswerSignal struct {
	peerConnection *PeerConnection
	ctx            context.Context
	offerPath      string
	answerPath     string
}

// CreateFileAnswerSignal creates a new FileAnswerSignal
func CreateFileAnswerSignal(ctx context.Context, peerConnection *PeerConnection, offerPath string, answerPath string) *FileAnswerSignal {
	return &FileAnswerSignal{
		peerConnection: peerConnection,
		ctx:            ctx,
		offerPath:      offerPath,
		answerPath:     answerPath,
	}
}

// Connect implements the BaseSignal interface
func (signal *FileAnswerSignal) Connect(category, connectionLabel string) error {
	// Use category and connectionLabel to create unique filenames if needed
	if category != "" && connectionLabel != "" {
		signal.offerPath = filepath.Join(signal.offerPath, category, connectionLabel, "offer.txt")
		signal.answerPath = filepath.Join(signal.answerPath, category, connectionLabel, "answer.txt")

		// Ensure directory exists
		dir := filepath.Dir(signal.answerPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("error creating directory %s: %w", dir, err)
		}
	}

	// Wait for offer file to exist
	offer, err := signal.waitForOffer()
	if err != nil {
		return err
	}

	// Process the offer
	return signal.processOffer(offer)
}

// waitForOffer waits for the offer file to appear and loads it
func (signal *FileAnswerSignal) waitForOffer() (webrtc.SessionDescription, error) {
	var offer webrtc.SessionDescription

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	timeout := time.NewTimer(5 * time.Minute)
	defer timeout.Stop()

	fmt.Printf("Waiting for offer file at %s...\n", signal.offerPath)

	for {
		select {
		case <-ticker.C:
			// Check if offer file exists
			if _, err := os.Stat(signal.offerPath); err == nil {
				// Load offer from file
				offer, err := signal.loadSDPFromFile(signal.offerPath)
				if err != nil {
					fmt.Printf("Error loading offer: %v, retrying...\n", err)
					continue
				}
				return offer, nil
			}
		case <-timeout.C:
			return offer, errors.New("timeout waiting for offer file")
		case <-signal.ctx.Done():
			return offer, errors.New("context canceled")
		}
	}
}

// processOffer processes the offer and creates an answer
func (signal *FileAnswerSignal) processOffer(offer webrtc.SessionDescription) error {
	// SetInputOption remote description
	if err := signal.peerConnection.peerConnection.SetRemoteDescription(offer); err != nil {
		return fmt.Errorf("error setting remote description: %w", err)
	}

	// Create answer
	answer, err := signal.peerConnection.peerConnection.CreateAnswer(nil)
	if err != nil {
		return fmt.Errorf("error creating answer: %w", err)
	}

	// SetInputOption local description
	if err := signal.peerConnection.peerConnection.SetLocalDescription(answer); err != nil {
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

	// Save answer to file
	if err := signal.saveSDPToFile(signal.peerConnection.peerConnection.LocalDescription(), signal.answerPath); err != nil {
		return fmt.Errorf("error saving answer to file: %w", err)
	}

	fmt.Printf("Answer saved to %s\n", signal.answerPath)
	return nil
}

// Close implements the BaseSignal interface
func (signal *FileAnswerSignal) Close() error {
	// Nothing to close for file-based signaling
	return nil
}

// saveSDPToFile saves a SessionDescription to a file
func (signal *FileAnswerSignal) saveSDPToFile(sdp *webrtc.SessionDescription, filename string) error {
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
func (signal *FileAnswerSignal) loadSDPFromFile(filename string) (webrtc.SessionDescription, error) {
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
