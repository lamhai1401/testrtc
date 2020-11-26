package utils

import (
	"os"

	"github.com/pion/webrtc/v3"
)

var (
	// DEBUG to logging debug
	DEBUG = os.Getenv("DEBUG")
	// TIMEOUT linter
	TIMEOUT = os.Getenv("TIMEOUT")
)

// NewSDPType format sdp type between pion and javascript
func NewSDPType(raw string) webrtc.SDPType {
	switch raw {
	case "offer":
		return webrtc.SDPTypeOffer
	case "answer":
		return webrtc.SDPTypeAnswer
	default:
		return webrtc.SDPType(webrtc.Unknown)
	}
}
