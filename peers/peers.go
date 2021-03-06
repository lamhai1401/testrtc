package peers

import (
	"fmt"
	"sync"

	v2 "github.com/beowulflab/mixer-v2/v2"
	"github.com/beowulflab/rtcbase-v2/utils"
	"github.com/beowulflab/signal/signal-wss"
	"github.com/lamhai1401/gologs/logs"
	"github.com/pion/webrtc/v2"
)

const (
	mixerID = "mixedStreamID"
)

// Peers linter
type Peers struct {
	id      string
	bitrate int
	signal  *signal.NotifySignal // send socket
	conns   *utils.AdvanceMap
	configs *webrtc.Configuration
	mixer   v2.Mixer
	mutex   sync.RWMutex
}

// NewPeers litner
func NewPeers() (*Peers, error) {
	p := &Peers{
		id:      mixerID,
		conns:   utils.NewAdvanceMap(),
		bitrate: 1000,
		configs: utils.GetTurns(),
		mixer: v2.NewMixer(
			10,
			"mixedStreamID",
			1500,
			true),
	}

	err := p.mixer.Start()
	if err != nil {
		return nil, err
	}

	sig := signal.NewNotifySignal("123", p.processNotifySignal)
	go sig.Start()
	p.signal = sig
	return p, nil
}

func (ps *Peers) processNotifySignal(values []interface{}) {
	if len(values) < 3 {
		logs.Error("Len of msg < 4")
		return
	}

	signalID, hasSignalID := values[0].(string)
	if !hasSignalID {
		logs.Error(fmt.Sprintf("[ProcessSignal] Invalid signal ID: %v", signalID))
		return
	}

	sessionID, hasSessionID := values[1].(string)
	if !hasSessionID {
		logs.Error(fmt.Sprintf("[ProcessSignal] Invalid session ID: %v", sessionID))
		return
	}

	event, isEvent := values[2].(string)
	if !isEvent {
		logs.Error(fmt.Sprintf("[ProcessSignal] Invalid event: %v", event))
		return
	}

	var err error
	switch event {
	case "ok":
		logs.Debug(fmt.Sprintf("Receive ok from id: %s_%s", signalID, sessionID))
		err = ps.handleOkEvent(signalID, sessionID)
		break
	case "candidate":
		err = ps.handCandidateEvent(signalID, sessionID, values[3])
		break
	case "sdp":
		logs.Debug(fmt.Sprintf("Receive sdp from id: %s_%s", signalID, sessionID))
		err = ps.handleSDPEvent(signalID, sessionID, values[3])
		break
	}

	if err != nil {
		logs.Error(err.Error())
		ps.sendError(signalID, sessionID, err.Error())
	}
}

func (ps *Peers) handleOkEvent(signalID string, sessionID string) error {
	ps.sendOk(signalID, sessionID)
	return nil
}

func (ps *Peers) handCandidateEvent(signalID string, sessionID string, value interface{}) error {
	return ps.addCandidate(signalID, sessionID, value)
}

func (ps *Peers) addCandidate(id, session string, values interface{}) error {
	if conn := ps.getConn(id); conn != nil {
		return conn.AddICECandidate(values)
	}
	return fmt.Errorf("Connection with id %s is nil", id)
}

func (ps *Peers) handleSDPEvent(signalID, sessionID string, value interface{}) error {
	return ps.addSDP(signalID, sessionID, value)
}

func (ps *Peers) addSDP(id, session string, values interface{}) error {

	var err error
	peer := ps.getConn(id)

	if peer != nil {
		ps.closeConn(id)
	}

	peer, err = ps.addConn(id, session)
	if err != nil {
		return err
	}

	_, err = peer.NewConnection(values, ps.getConfig())
	if err != nil {
		return err
	}
	ps.handleConnEvent(peer)

	err = peer.AddSDP(values)
	if err != nil {
		return err
	}

	answer, err := peer.GetLocalDescription()
	if err != nil {
		return err
	}

	ps.sendSDP(id, session, answer)
	return err
}
