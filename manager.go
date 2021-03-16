package main

import (
	"fmt"
	"os"
	"strconv"
	"sync"

	"github.com/beowulflab/signal/signal-wss"
	"github.com/lamhai1401/gologs/logs"
	log "github.com/lamhai1401/gologs/logs"
	"github.com/lamhai1401/testrtc/peer"
	"github.com/lamhai1401/testrtc/utils"
	"github.com/lamhai1401/testrtc/worker"
	"github.com/mitchellh/mapstructure"
	"github.com/pion/webrtc/v3"
)

// Manager linter
type Manager struct {
	bitrate      int
	codec        string        // video codec vp8 or vp9 only
	worker       worker.Worker // handle peer
	notifySignal *signal.NotifySignal
	iceCache     *utils.AdvanceMap // save (ice candidate - peer connection ID) ice cache when peer connention is not set at this time
	isClosed     bool
	mutex        sync.RWMutex
}

// NewPeerWorker linter
func NewPeerWorker(
	url string,
	id string,
) (*Manager, error) {
	os.Setenv("MULTIPLE_URLL", url)
	m := &Manager{
		codec:   os.Getenv("CODEC_TYPE"),
		bitrate: getBitrate(),
	}

	notifySignal := signal.NewNotifySignal(
		id,
		m.processNotifySignal,
	)

	err := notifySignal.Start()
	if err != nil {
		return nil, err
	}

	m.setNotifySignal(notifySignal)

	//Init worker
	worker := worker.NewPeerWorker(id, m.getBitrate(), notifySignal)
	if err := worker.Start(); err != nil {
		return nil, err
	}
	m.setWorker(worker)

	return m, nil
}

func getBitrate() int {
	df := 500
	if br := os.Getenv("BIT_RATE"); br != "" {
		bit, err := strconv.Atoi(br)
		if err == nil {
			return bit
		}
	}
	return df
}

func (m *Manager) getBitrate() int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.bitrate
}

func (m *Manager) processNotifySignal(values []interface{}) {
	if len(values) == 3 {
		m.processPing(values)
		return
	}

	if len(values) < 5 {
		log.Error(fmt.Sprintf("Len of msg < 5: %v", values))
		return
	}

	signalID, hasSignalID := values[0].(string)
	if !hasSignalID {
		log.Error(fmt.Sprintf("[ProcessSignal] Invalid signal ID: %v", values))
		return
	}

	streamID, hasStreamID := values[1].(string)
	if !hasStreamID {
		log.Error(fmt.Sprintf("[ProcessSignal] Invalid stream ID: %v", values))
		return
	}

	role, isRole := values[2].(string)
	if !isRole {
		log.Error(fmt.Sprintf("[ProcessSignal] Invalid role: %v", values))
		return
	}

	sessionID, hasSessionID := values[3].(string)
	if !hasSessionID {
		log.Error(fmt.Sprintf("[ProcessSignal] Invalid session ID: %v", values))
		return
	}

	event, isEvent := values[4].(string)
	if !isEvent {
		log.Error(fmt.Sprintf("[ProcessSignal] Invalid event: %v", values))
		return
	}

	// format signalID
	var err error
	switch event {
	case "ok":
		log.Debug(fmt.Sprintf("Receive ok from signal peer: %s_%s_%s", signalID, streamID, sessionID))
		err = m.handleOkEvent(signalID, streamID, role, sessionID)
		break
	case "sdp":
		log.Debug(fmt.Sprintf("Receive sdp from signal peer: %s_%s_%s", signalID, streamID, sessionID))
		err = m.handleSDPEvent(signalID, streamID, role, sessionID, values[5])
		break
	case "candidate":
		log.Debug(fmt.Sprintf("Receive candidate from signal peer: %s_%s_%s ==> [%v]", signalID, streamID, sessionID, values[4]))
		err = m.handCandidateEvent(signalID, streamID, role, sessionID, values[5])
		break
	case "close":
		log.Debug(fmt.Sprintf("Receive close from signal peer: %s_%s_%s", signalID, streamID, sessionID))
		if worker := m.getWorker(); worker != nil {
			worker.RemoveConnection(signalID, streamID, sessionID)
		}
		break
	case "reconnect":
		log.Debug(fmt.Sprintf("Receive reconnect from signal peer: %s_%s_%s", signalID, streamID, sessionID))
		err = m.handleReconnect(signalID, streamID, role, sessionID)
		break
	case "error":
		log.Debug(fmt.Sprintf("Receive error from signal peer: %s_%s_%s", signalID, streamID, sessionID))
		logs.Error(values)
		break
	default:
		err = fmt.Errorf("receive not processing event: %s", event)
	}

	if err != nil {
		if err.Error() == peer.ErrAddCandidate.Error() {
			var candidateInit webrtc.ICECandidateInit
			if err1 := mapstructure.Decode(values[5], &candidateInit); err1 == nil {
				m.setPeerICECache(candidateInit, sessionID)
			}
		}

		m.sendError(signalID, streamID, role, sessionID, err.Error())
		log.Error("[processNotifySignal] err: ", err.Error())
	}
}

func (m *Manager) getNotifySignal() *signal.NotifySignal {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.notifySignal
}

func (m *Manager) setNotifySignal(s *signal.NotifySignal) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.notifySignal = s
}

func (m *Manager) processPing(values []interface{}) {
	// [<fromId>,”process-mgr”,”ping”]
	fromID, hasFromID := values[0].(string)
	if !hasFromID {
		log.Error(fmt.Sprintf("[processPing] Invalid fromID: %v", values))
		return
	}

	event, hasEvent := values[1].(string)
	if !hasEvent {
		log.Error(fmt.Sprintf("[processPing] Invalid event: %v", values))
		return
	}

	ping, hasPing := values[2].(string)
	if !hasPing {
		log.Error(fmt.Sprintf("[processPing] Invalid ping: %v", values))
		return
	}

	if event == "process-mgr" && ping == "ping" {
		m.sendPong(fromID)
	} else {
		logs.Warn("Wrong format ping msg: %v", values)
	}
}

func (m *Manager) sendPong(fromID string) error {
	if signal := m.getNotifySignal(); signal != nil {
		signal.Send(fromID, "process-mgr", "pong")
		log.Debug(fmt.Sprintf("==== Send pong to: %s", fromID))
		return nil
	}
	return ErrNilSignal
}

func (m *Manager) setPeerICECache(iceCandidate interface{}, peerConnectionID string) {
	if caches := m.getICECache(); caches != nil {
		caches.Seti(iceCandidate, peerConnectionID)
	}
}

func (m *Manager) getWorker() worker.Worker {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.worker
}

func (m *Manager) setWorker(w worker.Worker) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.worker = w
}

func (m *Manager) getICECache() *utils.AdvanceMap {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.iceCache
}

func (m *Manager) sendError(signalID, streamID, role, sessionID, reason string) error {
	if signal := m.getNotifySignal(); signal != nil {
		signal.Send(signalID, streamID, role, sessionID, "error", reason)
		return nil
	}

	return ErrNilSignal
}

func (m *Manager) sendOk(signalID, streamID, role, sessionID string) error {
	if signal := m.getNotifySignal(); signal != nil {
		signal.Send(signalID, streamID, role, sessionID, "ok")
		logs.Info(fmt.Sprintf("Send ok to %s_%s_%s_%s", signalID, streamID, role, sessionID))
		return nil
	}
	return ErrNilSignal
}

func (m *Manager) sendSDP(signalID, streamID, role, sessionID string, sdp interface{}) error {
	if signal := m.getNotifySignal(); signal != nil {
		signal.Send(signalID, streamID, role, sessionID, "sdp", sdp)
		log.Debug(fmt.Sprintf("==== Send sdp to: %s_%s_%s", signalID, streamID, sessionID))
		return nil
	}
	return ErrNilSignal
}

func (m *Manager) getConnections(signalID string) (peer.Connections, error) {
	w := m.getWorker()
	if w == nil {
		return nil, ErrNilWorker
	}
	return w.GetConnections(signalID), nil
}

func (m *Manager) addConnections(signalID string) (peer.Connections, error) {
	// Tiem peer connection
	w := m.getWorker()
	if w == nil {
		return nil, ErrNilWorker
	}
	w.AddConnections(signalID)
	return w.GetConnections(signalID), nil
}

func (m *Manager) getConnection(signalID, streamID string) (peer.Connection, error) {
	w := m.getWorker()
	if w == nil {
		return nil, ErrNilWorker
	}

	conn, err := w.GetConnection(signalID, streamID)
	if err != nil {
		return nil, err
	}

	if conn == nil {
		return nil, fmt.Errorf("Connection with id: %s is nil", signalID)
	}

	return conn, nil
}

func (m *Manager) getCodec() string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.codec
}

func (m *Manager) setCodec(c string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.codec = c
}

func (m *Manager) addConnection(
	signalID string,
	streamID string,
	sessionID string,
	role string,
	handleAddPeer func(signalID string, streamID string, role string, sessionID string),
	handleFailedPeer func(signalID string, streamID string, role string, sessionID string),
	codec string,
	payloadType int,
) (peer.Connection, error) {
	// Tiem peer connection
	w := m.getWorker()
	if w == nil {
		return nil, ErrNilWorker
	}

	err := w.AddConnection(
		signalID,
		streamID,
		sessionID,
		role,
		handleAddPeer,
		handleFailedPeer,
		codec,
		payloadType,
	)
	if err != nil {
		return nil, err
	}

	return m.getConnection(signalID, streamID)
}

// getPeerICECache get all ice cache with input peer connection
func (m *Manager) getPeerICECache(peerConnectionID string) []interface{} {
	var tmp []interface{}
	if caches := m.getICECache(); caches != nil {
		caches.Iter(func(key, value interface{}) bool {
			values, ok := value.(string)
			if ok {
				if values == peerConnectionID {
					tmp = append(tmp, key)
				}
			}
			return true
		})
	}
	return tmp
}

func (m *Manager) removePeerICECache(iceCandidate interface{}) {
	if caches := m.getICECache(); caches != nil {
		caches.Delete(iceCandidate)
	}
}

func (m *Manager) addIceCache(conn peer.Connection) {
	tmp := m.getPeerICECache(conn.GetSessionID())
	if len(tmp) == 0 || conn == nil {
		return
	}

	for _, value := range tmp {
		err := conn.AddICECandidate(value)
		if err != nil {
			logs.Error(fmt.Sprintf("Add %s ice candidate in cache err: %s ", conn.GetSessionID(), err.Error()))
		} else {
			logs.Warn(fmt.Sprintf("Add %s ice candiate from cache successfully", conn.GetSessionID()))
			m.removePeerICECache(value)
		}
	}
}

func (m *Manager) handleOkEvent(signalID, streamID, role, sessionID string) error {
	// remove candidate in cache
	m.removePeerICECacheWithID(sessionID)
	m.sendOk(signalID, streamID, role, sessionID)
	return nil
}

func (m *Manager) handleSDPEvent(signalID, streamID, role, sessionID string, sdp interface{}) error {
	return m.addSDP(signalID, streamID, role, sessionID, sdp)
}

func (m *Manager) addSDP(signalID, streamID, role, sessionID string, sdp interface{}) error {
	// dest peer
	conns, err := m.getConnections(signalID)
	if err != nil {
		return err
	}
	if conns == nil {
		conns, err = m.addConnections(signalID)
		if err != nil {
			return err
		}
	}

	conn, _ := m.getConnection(signalID, streamID)
	if conn != nil {
		conns.RemoveConnection(streamID)
	}

	// get codec and payload type
	var data utils.SDPTemp
	codec := defaultCodecVP9 // to check user sdp has defautl codec in env
	payloadType := defaultPayloadVP9
	err = mapstructure.Decode(sdp, &data)

	if err != nil {
		return err
	}

	if m.getCodec() != "" {
		codec = m.getCodec()
		// call thi chu func
		pl, err := getFirstPayload(data.SDP, codec)
		logs.Warn(fmt.Sprintf("%s_%s_%s getFirstPayload is %s", signalID, streamID, sessionID, pl))
		if err != nil {
			if num, err := strconv.Atoi(pl); err != nil {
				payloadType = num
			}
			logs.Error("[getFirstPayload] has err: %s", err.Error())
		}
	}

	logs.Warn(fmt.Sprintf("[processDestPeer] %s_%s_%s peer connection was created with codec (%s/%d)", signalID, streamID, sessionID, codec, payloadType))
	conn, err = m.addConnection(
		signalID,
		streamID,
		sessionID,
		role,
		m.handleSuccessPeer,
		nil, // m.handleFailPeer,
		codec,
		payloadType,
	)

	if err != nil {
		return err
	}

	err = conn.AddSDP(sdp)
	if err != nil {
		return nil
	}
	m.addIceCache(conn)

	SDP, err := conn.GetLocalDescription()
	if err != nil {
		return nil
	}
	// offerId := m.getOfferId(sessionID)
	//m.setPeerConnectionState(sessionID, peeringState)
	return m.sendSDP(signalID, streamID, role, sessionID, SDP)
}

func (m *Manager) handCandidateEvent(signalID, streamID, role, sessionID string, value interface{}) error {
	w := m.getWorker()
	if w == nil {
		return ErrNilWorker
	}
	connections := w.GetConnections(signalID)
	if connections == nil {
		return fmt.Errorf("Connections with id %s is nil", signalID)
	}

	// destrucre candidate
	candidateInit, ok := value.(*webrtc.ICECandidateInit)
	if !ok {
		err := mapstructure.Decode(value, &candidateInit)
		if err != nil {
			return err
		}
	}

	err := connections.AddCandidate(streamID, candidateInit)
	if err != nil {
		logs.Warn(fmt.Sprintf("RECEIVE CANDIDATE TOO SOON: %v", candidateInit))
		m.setPeerICECache(candidateInit, sessionID)
	}

	if conn := connections.GetConnection(streamID); conn != nil {
		m.addIceCache(conn)
	}
	return err
}

func (m *Manager) handleReconnect(signalID, streamID, role, sessionID string) error {
	signal := m.getNotifySignal()
	if signal == nil {
		return ErrNilSignal
	}
	signal.Send(signalID, streamID, role, sessionID, "reconnect-ok")
	// signal.Send(signalID, streamID, sessionID, "ok")
	return nil
}

func (m *Manager) handleSuccessPeer(signalId, streamId, role, subcriberSessionID string) {
	// var err error
	logs.Error(fmt.Sprintf("%s (signalID) %s (streamID) %s (sessionID) %s (role) peer success", signalId, streamId, subcriberSessionID, role))

	// if conn := m.getConnection(signalId, streamId)

	if role == "self" || role == "source" {
		return
	}

	logs.Error(fmt.Sprintf("Regiter id %s (signalID) %s (streamID) %s (sessionID) %s (role)", signalId, streamId, subcriberSessionID, role))
	if err := m.register(signalId, streamId); err != nil {
		m.sendError(signalId, streamId, role, subcriberSessionID, err.Error())
	}

	m.removePeerICECacheWithID(subcriberSessionID)
}

// register regis source peer to des peer
func (m *Manager) register(signalID string, streamID string) error {
	w := m.getWorker()
	if w == nil {
		return ErrNilWorker
	}

	// streamID = utils.GetStreamID(streamID)

	return w.Register(
		signalID,
		streamID,
		func(signalID, streamID, subcriberSessionID, reason string) {
			m.unRegister(signalID, streamID, subcriberSessionID)
		},
	)
}

func (m *Manager) unRegister(signalID, streamID, subcriberSessionID string) error {
	w := m.getWorker()
	if w == nil {
		return ErrNilWorker
	}
	w.UnRegister(signalID, streamID, subcriberSessionID)
	return nil
}

func (m *Manager) removePeerICECacheWithID(peerConnectionID string) {
	if caches := m.getICECache(); caches != nil {
		caches.Iter(func(key, value interface{}) bool {
			values, ok := value.(string)
			if ok {
				if values == peerConnectionID {
					m.removePeerICECache(key)
				}
			}
			return true
		})
	}
}
