package conn

import (
	"os"
	"strconv"
)

const (
	hostType      = "host"
	presenterType = "presenter"
	nodeType      = "node"
	clientType    = "member"
	mixerType     = "mixer"
)

const splitStr = "-"

const (
	destRole   = "dest"
	sourceRole = "source"
	mixerRole  = "mixer"
)

const (
	// MimeTypeH264 H264 MIME type.
	// Note: Matching should be case insensitive.
	MimeTypeH264 = "video/H264"
	// MimeTypeOpus Opus MIME type
	// Note: Matching should be case insensitive.
	MimeTypeOpus = "audio/opus"
	// MimeTypeVP8 VP8 MIME type
	// Note: Matching should be case insensitive.
	MimeTypeVP8 = "video/VP8"
	// MimeTypeVP9 VP9 MIME type
	// Note: Matching should be case insensitive.
	MimeTypeVP9 = "video/VP9"
	// MimeTypeG722 G722 MIME type
	// Note: Matching should be case insensitive.
	MimeTypeG722 = "audio/G722"
	// MimeTypePCMU PCMU MIME type
	// Note: Matching should be case insensitive.
	MimeTypePCMU = "audio/PCMU"
	// MimeTypePCMA PCMA MIME type
	// Note: Matching should be case insensitive.
	MimeTypePCMA = "audio/PCMA"
)

func getNumberTracks() int {
	if l := os.Getenv("NUMBER_TRACK"); l != "" {
		length, err := strconv.Atoi(l)
		if err == nil && length > 1 {
			return length
		}
	}
	return 1
}

const defaultTrackState = "0"
