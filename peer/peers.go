package peer

import (
	"fmt"
	"sync"

	"github.com/beowulflab/signal/signal-wss"
	"github.com/lamhai1401/gologs/logs"
	"github.com/lamhai1401/testrtc/utils"
	"github.com/pion/webrtc/v3"
)

// Peers handle mutilpe peer connection
type Peers struct {
	signalID string
	peers    *utils.AdvanceMap    // save streamID - peer
	signal   *signal.NotifySignal // send data signal (sdp, candidate, etc...)
	isClosed bool
	mutex    sync.RWMutex
}

// NewPeers mutilpe peer controller
func NewPeers(
	signalID string, // client signal ID
	signal *signal.NotifySignal,
	audioFwdm utils.Fwdm,
	videoFwdm utils.Fwdm,
) Connections {
	ps := &Peers{
		signalID: signalID,
		peers:    utils.NewAdvanceMap(),
		signal:   signal,
		isClosed: false,
	}

	// ps.serve()
	return ps
}

// Close linter
func (p *Peers) Close() {
	if !p.wasClosed() {
		p.setClosed(true)
		// p.closeICEChann()
		p.RemoveConnections()
	}
}

// RemoveConnection remove existing connection
func (p *Peers) RemoveConnection(streamID string) {
	p.closePeer(streamID)
	// p.deleteState(streamID)
	// p.deleteDatachannel(streamID)
	// p.deletePeerICECache(streamID)
}

// RemoveConnections remove all connection
func (p *Peers) RemoveConnections() {
	if peers := p.getPeers(); peers != nil {
		keys := peers.GetKeys()
		for _, key := range keys {
			p.RemoveConnection(key)
		}
	}
}

// AddConnection add new peer connection
func (p *Peers) AddConnection(
	bitrate *int,
	streamID string,
	role string,
	sessionID string,
	configs *webrtc.Configuration,
	handleAddPeer func(signalID string, streamID string, role string, sessionID string),
	handleFailedPeer func(signalID string, streamID string, role string, sessionID string),
	codec string,
	payloadType int,
) error {
	// remove if exist
	if peer := p.GetConnection(streamID); peer != nil {
		logs.Info(fmt.Sprintf("Remove existing peer connection of id %s_%s", streamID, sessionID))
		p.RemoveConnection(streamID)
	}

	//add new one
	peer := newPeerConnection(
		bitrate,
		streamID,
		role,
		sessionID,
		codec,
		payloadType,
	)

	conn, err := peer.InitPeer(configs)
	if err != nil {
		return err
	}

	conn.OnICECandidate(func(candidate *webrtc.ICECandidate) {
		if candidate != nil {
			p.sendCandidate(p.getSignalID(), streamID, peer.GetSessionID(), candidate)
		}
	})

	conn.OnICEConnectionStateChange(func(is webrtc.ICEConnectionState) {
		p.handleICEConnectionState(
			p.getSignalID(),
			streamID,
			is.String(),
			handleAddPeer,
			handleFailedPeer,
		)
	})

	conn.OnTrack(func(t *webrtc.TrackRemote, r *webrtc.RTPReceiver) {
		p.handleOnTrack(t, peer)
	})

	p.setPeer(streamID, peer)
	return nil
}

// GetConnection get peer connection
func (p *Peers) GetConnection(streamID string) Connection {
	peer := p.getPeer(streamID)
	if peer != nil {
		return peer
	}
	return nil
}

// AddCandidate linter
func (p *Peers) AddCandidate(streamID string, value interface{}) error {
	conn := p.getPeer(streamID)
	if conn == nil {
		logs.Warn(fmt.Sprintf("Cannot add candidate peer connection with input streamID %s is nil", streamID))
		return ErrAddCandidate
	}

	// add candidate
	err := conn.AddICECandidate(value)
	if err != nil {
		return err
	}

	return nil
}

// AddSDP linter
func (p *Peers) AddSDP(streamID string, value interface{}) error {
	conn := p.getPeer(streamID)
	if conn == nil {
		return fmt.Errorf("Peer connection is il")
	}
	return conn.AddSDP(value)
}

// CountAllPeer count all existing peer
func (p *Peers) CountAllPeer() int64 {
	if peers := p.getPeers(); peers != nil {
		return peers.Len()
	}
	return 0
}

// GetAllConnection linter
func (p *Peers) GetAllConnection() []Connection {
	tmp, _ := p.getAllPeer()
	return tmp
}

// GetAllStreamID linter
func (p *Peers) GetAllStreamID() []string {
	_, tmp := p.getAllPeer()
	return tmp
}
