package worker

import (
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/beowulflab/signal/signal-wss"
	"github.com/davecgh/go-spew/spew"
	"github.com/lamhai1401/gologs/logs"
	"github.com/lamhai1401/testrtc/peer"
	"github.com/lamhai1401/testrtc/utils"
	"github.com/pion/webrtc/v3"
)

// PeerWorker Set
type PeerWorker struct {
	bitrate int
	id      string
	// audioFwdm utils.Fwdm            // forward audio pkg
	// videoFwdm utils.Fwdm            // forward video pkg
	configs  *webrtc.Configuration // peer connection config
	peers    *utils.AdvanceMap     // save all peers with signalID
	signal   *signal.NotifySignal
	isClosed bool
	mutex    sync.RWMutex
}

// NewPeerWorker linter
func NewPeerWorker(
	id string,
	bitrate int,
	signal *signal.NotifySignal,
) Worker {
	w := &PeerWorker{
		id:      id,
		bitrate: bitrate,
		signal:  signal,
		// audioFwdm: utils.NewForwarderMannager(id),
		// videoFwdm: utils.NewForwarderMannager(id),
		peers: utils.NewAdvanceMap(),
	}

	return w
}

// Start linter
func (w *PeerWorker) Start() error {
	// get turn configs
	config := utils.GetTurnsByAPI()
	w.setTurnConfig(config)
	spew.Dump("Turn config list: ", config)
	go w.countInterVal()
	return nil
}

// AddConnections add new connections
func (w *PeerWorker) AddConnections(signalID string) {
	connections := peer.NewPeers(
		signalID,
		w.getSignal(),
		nil, // w.getAudioFwdm(),
		nil, // w.getVideoFwdm(),
	)
	if peers := w.getPeers(); peers != nil {
		peers.Set(signalID, connections)
	}
}

// AddConnection add new peer connection
func (w *PeerWorker) AddConnection(
	signalID string,
	streamID string,
	sessionID string,
	role string,
	handleAddPeer func(signalID string, streamID string, role string, sessionID string),
	handleFailedPeer func(signalID string, streamID string, role string, sessionID string),
	codec string,
	payloadType int,
) error {
	// get connections
	connections := w.getConnections(signalID)
	if connections == nil {
		return fmt.Errorf("Connections of signalID %s is nil", signalID)
	}

	err := connections.AddConnection(
		&w.bitrate,
		streamID,
		role,
		sessionID,
		w.getTurnConfig(),
		handleAddPeer,
		handleFailedPeer,
		codec,
		payloadType,
	)

	if err != nil {
		return fmt.Errorf("Add new connection %s err: %s", streamID, err.Error())
	}

	return nil
}

func (w *PeerWorker) setBitrate(i int) {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	w.bitrate = i
}

func (w *PeerWorker) getPeer(signalID, streamID string) peer.Connection {
	if peers := w.getPeers(); peers != nil {
		ps, has := peers.Get(signalID)
		if has {
			connections, ok := ps.(peer.Connections)
			if ok {
				return connections.GetConnection(streamID)
			}
		}
	}
	return nil
}

func (w *PeerWorker) countInterVal() {
	interval := getInterval()
	logs.Warn(fmt.Sprintf("Count interval start with %d every second", interval))
	ticker := time.NewTicker(time.Second * time.Duration(interval))
	for range ticker.C {
		logs.Warn("====== Count peer interval ======")
		w.countAllPeer()
	}
}

func getInterval() int {
	i := defaultInterval
	if interval := os.Getenv("SYNC_INTERVAL"); interval != "" {
		j, err := strconv.Atoi(interval)
		if err == nil {
			i = j
		}
	}
	return i
}

func (w *PeerWorker) countAllPeer() {
	var all int64
	if peers := w.getPeers(); peers != nil {
		peers.Iter(func(key, value interface{}) bool {
			signalID, ok1 := key.(string)
			connections, ok2 := value.(peer.Connections)
			if ok1 && ok2 {
				count := connections.CountAllPeer()
				all = all + count
				logs.Warn(fmt.Sprintf("==== %s has %d connections", signalID, count))
			}
			return true
		})
	}
	logs.Warn("==== Total connections is: ", all)
}

// GetConnections could be nil if not exist
func (w *PeerWorker) GetConnections(signalID string) peer.Connections {
	return w.getConnections(signalID)
}

// GetConnection couble be nil if not exist
func (w *PeerWorker) GetConnection(signalID, streamID string) (peer.Connection, error) {
	conns := w.getConnections(signalID)
	if conns == nil {
		return nil, fmt.Errorf("Connections with signal id %s is nil", signalID)
	}

	// conn := conns.GetConnection(streamID)
	// if conn == nil {
	// 	return nil, fmt.Errorf("Connection with stream id %s is nil", streamID)
	// }

	return conns.GetConnection(streamID), nil
}

// RemoveConnection remove existing peer connection
func (w *PeerWorker) RemoveConnection(signalID, streamID, sessionID string) error {
	// get connections
	connections := w.getConnections(signalID)
	if connections == nil {
		return fmt.Errorf("Connections of signalID %s is nil", signalID)
	}

	conn := connections.GetConnection(streamID)
	if conn != nil {
		if conn.GetSessionID() == sessionID {
			connections.RemoveConnection(streamID)
		} else {
			logs.Error(fmt.Errorf("%s_%s input session != peer session (%s != %s). Dont remove", signalID, streamID, sessionID, conn.GetSessionID()))
		}
	}
	w.countAllPeer()
	return nil
}

// RemoveConnections remove all connections
func (w *PeerWorker) RemoveConnections(signalID string) {
	w.closeConnections(signalID)
}
