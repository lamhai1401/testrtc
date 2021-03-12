package worker

import "github.com/lamhai1401/testrtc/peer"

// Worker peer connection worker
type Worker interface {
	Start() error
	AddConnection(
		signalID string,
		streamID string,
		sessionID string,
		role string,
		handleAddPeer func(signalID string, streamID string, role string, sessionID string),
		handleFailedPeer func(signalID string, streamID string, role string, sessionID string),
		codec string,
		payloadType int,
	) error
	AddConnections(signalID string)
	GetConnections(signalID string) peer.Connections
	GetConnection(signalID, streamID string) (peer.Connection, error)
	RemoveConnection(signalID, streamID, sessionID string) error
	// RemoveConnections(signalID string)
	// RemoveConnectionsHasStream(streamID string) error
	// add the method add register
	Register(signalID string, streamID string, errHandler func(signalID string, streamID string, subcriberSessionID string, reason string)) error
	// add the method add register
	UnRegister(signalID string, streamID string, subcriberSessionID string)
	// checking peer connection was received data or not, return error if received data or null
	// CheckPeerDataState(signalID, sessionID string) error
	// GetSessionByStream(streamID string) []string
	// GetSessionBySignal(signalID string) []string
	// GetAllConnectionID() map[string][]string
	// GetState(signalID, streamID string) string // get single connection state
}
