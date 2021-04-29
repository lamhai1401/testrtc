package conn

import "fmt"

var (
	// ErrNilVideoTrack linter
	ErrNilVideoTrack = "Video track is nil"
	// ErrNilAudioTrack linter
	ErrNilAudioTrack = "Audio track is nil"
	// ErrNilPeerconnection linter
	ErrNilPeerconnection = "Peer connection is nil"
	// ErrNillPeerDataState linter
	ErrNillPeerDataState = fmt.Errorf("Invalid data state of this peer connection")
	// ErrReceivedData linter
	ErrReceivedData = fmt.Errorf("This peer connecition was received data")
	// ErrNilDataState linter
	ErrNilDataState = fmt.Errorf("Data state is nil")
	// ErrAddCandidate linter
	ErrAddCandidate = fmt.Errorf("Cannot add candiate peer connnection with input streamID is nil")
)
