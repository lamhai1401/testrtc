package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/pion/webrtc/v3"
)

// GetTurnURL linter
func GetTurnURL() string {
	if url := os.Getenv("TURN_URL"); url != "" {
		return url
	}
	return "https://call-config.beowulfchain.com/api/config"
}

// GetTurnConfigList get turn list via API
func GetTurnConfigList() (*webrtc.Configuration, error) {
	url := GetTurnURL()

	requestBody := TurnRequestBody{
		CallType:  "string",
		RequestID: "string",
	}

	stringBody, err := ToString(requestBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer([]byte(stringBody)))
	req.Header.Set("Content-Type", "application/json")

	if err != nil {
		return nil, err
	}

	// For control over HTTP client headers,
	// redirect policy, and other settings,
	// create a Client
	// A Client is an HTTP client
	client := &http.Client{}

	// Send the request via a client
	// Do sends an HTTP request and
	// returns an HTTP response
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	// Callers should close resp.Body
	// when done reading from it
	// Defer the closing of the body
	defer resp.Body.Close()

	// Fill the record with the data from the JSON
	var record TurnConfigList

	// Use json.Decode for reading streams of JSON data
	if err := json.NewDecoder(resp.Body).Decode(&record); err != nil {
		return nil, err
	}

	// convert to webrtc configure
	turnList := &webrtc.Configuration{
		SDPSemantics: webrtc.SDPSemanticsUnifiedPlan,
	}

	for _, turn := range record.Data {
		iceSever := webrtc.ICEServer{
			URLs:           []string{turn.URLs},
			Username:       turn.Username,
			Credential:     turn.Password,
			CredentialType: webrtc.ICECredentialTypePassword,
		}
		turnList.ICEServers = append(turnList.ICEServers, iceSever)
	}

	return turnList, nil
}

// GetTurnsByAPI get turn config from server or default
func GetTurnsByAPI() *webrtc.Configuration {
	tempChan := make(chan *webrtc.Configuration, 1)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	go func() {
		turnList, err := GetTurnConfigList()
		if err != nil {
			log.Println(fmt.Errorf("Get turn list from server error: %v", err.Error()))
			return
		}
		tempChan <- turnList
	}()

	go func() {
		select {
		case <-ctx.Done():
			tempChan <- GetTurns()
			return
		}
	}()

	temp := <-tempChan
	return temp
}

// GetTurns use it if get via API failed
func GetTurns() *webrtc.Configuration {
	return &webrtc.Configuration{
		SDPSemantics: webrtc.SDPSemanticsUnifiedPlan,
		ICEServers: []webrtc.ICEServer{
			webrtc.ICEServer{
				URLs: []string{"stun:34.92.44.253:3478"},
			},
			webrtc.ICEServer{
				URLs: []string{"stun:35.247.153.249:3478"},
			},
			// Asia
			webrtc.ICEServer{
				URLs:           []string{"turn:34.92.44.253:5349"},
				Username:       "username",
				Credential:     "password",
				CredentialType: webrtc.ICECredentialTypePassword,
			},
			webrtc.ICEServer{
				URLs:           []string{"turn:35.247.153.249:5349"},
				Username:       "username",
				Credential:     "password",
				CredentialType: webrtc.ICECredentialTypePassword,
			},
		},
	}
}
