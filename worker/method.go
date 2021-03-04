package worker

import (
	"github.com/beowulflab/signal/signal-wss"
	"github.com/lamhai1401/testrtc/peer"
	"github.com/lamhai1401/testrtc/utils"
	"github.com/pion/webrtc/v3"
)

func (w *PeerWorker) getID() string {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	return w.id
}

func (w *PeerWorker) getSignal() *signal.NotifySignal {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	return w.signal
}

func (w *PeerWorker) setSignal(s *signal.NotifySignal) {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	w.signal = s
}

func (w *PeerWorker) getTurnConfig() *webrtc.Configuration {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	return w.configs
}

func (w *PeerWorker) setTurnConfig(c *webrtc.Configuration) {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	w.configs = c
}

func (w *PeerWorker) getBitrate() int {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	return w.bitrate
}

func (w *PeerWorker) getPeers() *utils.AdvanceMap {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	return w.peers
}

func (w *PeerWorker) getConnections(signalID string) peer.Connections {
	if ps := w.getPeers(); ps != nil {
		connections, has := ps.Get(signalID)
		if has {
			conns, ok := connections.(peer.Connections)
			if ok {
				return conns
			}
		}
	}
	return nil
}

func (w *PeerWorker) closeConnections(signalID string) {
	if conns := w.getConnections(signalID); conns != nil {
		w.deleteConnections(signalID)
		conns.Close()
		conns = nil
	}
}

func (w *PeerWorker) deleteConnections(signalID string) {
	if ps := w.getPeers(); ps != nil {
		ps.Delete(signalID)
	}
}
