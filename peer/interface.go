package peer

import (
	"github.com/pion/rtp"
	"github.com/pion/webrtc/v3"
)

// Connection interface
type Connection interface {
	GetRole() string
	GetSessionID() string  // for handle ice restart
	SetSessionID(s string) // for handle ice restart
	HandleVideoTrack(remoteTrack *webrtc.TrackRemote)
	InitPeer(configs *webrtc.Configuration) (*webrtc.PeerConnection, error)
	IsConnected() bool
	CreateOffer(iceRestart bool) error
	CreateAnswer() error
	AddSDP(values interface{}) error
	AddICECandidate(icecandidate interface{}) error
	AddVideoRTP(packet *rtp.Packet) error
	AddAudioRTP(packet *rtp.Packet) error
	SetIsConnected(states bool)
	// SetRemoteVideoTrack(remoteTrack *webrtc.TrackRemote)
	// SetRemoteAudioTrack(remoteTrack *webrtc.TrackRemote)
	// GetRemoteTrack() (*webrtc.TrackRemote, *webrtc.TrackRemote)
	GetLocalDescription() (*webrtc.SessionDescription, error)
	Close()
}

// Connections linter
type Connections interface {
	Close()
	AddConnection(
		bitrate *int,
		streamID string,
		role string,
		sessionID string,
		configs *webrtc.Configuration,
		handleAddPeer func(signalID string, streamID string, role string, sessionID string),
		handleFailedPeer func(signalID string, streamID string, role string, sessionID string),
		codec string,
		payloadType int,
	) error

	RemoveConnection(
		streamID string,
	)
	GetConnection(streamID string) Connection
	// RemoveConnections()
	GetStates() map[string]string    // get all connection states
	GetState(streamID string) string // get single connection state
	AddCandidate(streamID string, value interface{}) error
	AddSDP(stream string, values interface{}) error
	// SetPeerDataState(sessionID, state string)
	// CheckPeerDataState(sessionID string) error
	GetAllConnection() []Connection
	GetAllStreamID() []string
	// SetVideoCodec(codecs string, payloadType int)
	CountAllPeer() int64
}
