package peer

import (
	"fmt"

	"github.com/beowulflab/signal/signal-wss"
	"github.com/lamhai1401/gologs/logs"
	"github.com/lamhai1401/testrtc/utils"
	"github.com/pion/rtp"
	"github.com/pion/webrtc/v3"
)

func (p *Peers) setClosed(state bool) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.isClosed = state
}

func (p *Peers) wasClosed() bool {
	return p.isClosed
}

func (p *Peers) getSignal() *signal.NotifySignal {
	return p.signal
}

func (p *Peers) setSignal(s *signal.NotifySignal) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.signal = s
}

func (p *Peers) getPeers() *utils.AdvanceMap {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.peers
}

func (p *Peers) getPeer(streamID string) *Peer {
	if clients := p.getPeers(); clients != nil {
		c, ok := clients.Get(streamID)
		if !ok {
			return nil
		}
		client, ok := c.(*Peer)
		if ok {
			return client
		}
	}
	return nil
}

func (p *Peers) deletePeer(streamID string) {
	if clients := p.getPeers(); clients != nil {
		clients.Delete(streamID)
	}
}

func (p *Peers) setPeer(streamID string, c *Peer) {
	if p := p.getPeers(); p != nil {
		p.Set(streamID, c)
	}
}

func (p *Peers) closePeer(streamID string) {
	if client := p.getPeer(streamID); client != nil {
		p.deletePeer(streamID)
		// close peer
		client.Close()
		logs.Info(fmt.Sprintf("%s_%s_%s peer connection was removed", p.getSignalID(), streamID, client.getSessionID()))
		client = nil
	}
}

func (p *Peers) getSignalID() string {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.signalID
}

func (p *Peers) sendCandidate(singalID string, streamID string, sessionID string, candidate *webrtc.ICECandidate) {
	p.sendSignal(singalID, streamID, sessionID, "candidate", candidate.ToJSON())
}

func (p *Peers) sendSDP(singalID string, streamID string, sessionID string, sdp interface{}) {
	p.sendSignal(singalID, streamID, sessionID, "SDP", sdp)
}

func (p *Peers) sendOK(singalID string, streamID string, sessionID string) {
	p.sendSignal(singalID, streamID, sessionID, "ok")
}

func (p *Peers) sendSignal(input ...interface{}) {
	if signal := p.getSignal(); signal != nil {
		signal.Send(input...)
	}
}

func (p *Peers) getAllPeer() ([]Connection, []string) {
	var conns []Connection
	var ids []string

	if ps := p.getPeers(); ps != nil {
		ps.Iter(func(key, value interface{}) bool {
			connection, ok1 := value.(Connection)
			if ok1 {
				conns = append(conns, connection)
			}

			streamID, ok2 := key.(string)
			if ok2 {
				ids = append(ids, streamID)
			}
			return true
		})
	}

	return conns, ids
}

func (p *Peers) handleICEConnectionState(
	signalID string, streamID string, state string,
	handleAddPeer func(signalID string, streamID string, role string, sessionID string),
	handleFailedPeer func(signalID string, streamID string, role string, sessionID string),
) {
	peer := p.getPeer(streamID)

	if peer == nil || state == "" {
		return
	}

	logs.Warn(fmt.Sprintf("%s_%s_%s current ICE states is: %s", signalID, streamID, peer.getSessionID(), state))
	// p.setState(streamID, state)

	switch state {
	case "connected":
		if !peer.IsConnected() { // notif if this is new peer
			peer.SetIsConnected(true)
			// set data state
			// p.setPeerDataState(peer.getSessionID(), NOTYET)
			// call peer success
			// handleAddPeer(signalID, streamID, peer.getRole(), peer.getSessionID())
		}
		break
	case "closed":
		sessionID := peer.getSessionID()
		logs.Info(fmt.Sprintf("%s_%s_%s ice state is %s", signalID, streamID, peer.getSessionID(), state))
		if conn := p.GetConnection(streamID); conn != nil {
			if sessionID == peer.getSessionID() {
				logs.Warn("Remove old peer connection (%s_%s_%s) has state %s", signalID, streamID, peer.getSessionID(), state)
				p.RemoveConnection(peer.getStreamID())
			}
		}
		break
	default:
		return
	}
}

// handle peer remotetrack with streamID
func (p *Peers) handleOnTrack(remoteTrack *webrtc.TrackRemote, peer *Peer) {
	kind := remoteTrack.Kind().String()
	logs.Debug(fmt.Sprintf("Has %s remote track of id %s_%s", kind, p.getSignalID(), peer.getStreamID()))
	switch kind {
	case "video":
		peer.HandleVideoTrack(remoteTrack)
		// peer.setRemoteVideoTrack(remoteTrack)
		// fwdm = p.getVideoFwdm()
		go p.pushToFwd(remoteTrack, peer.getLocalVideoTrack(), peer)
		break
	case "audio":
		// peer.setRemoteAudioTrack(remoteTrack)
		// fwdm = p.getAudioFwdm()
		go p.pushToFwd(remoteTrack, peer.getLocalAudioTrack(), peer)
		break
	default:
		return
	}
}

func (p *Peers) pushToFwd(remoteTrack *webrtc.TrackRemote, localTrack *webrtc.TrackLocalStaticRTP, peer *Peer) {
	var rtp *rtp.Packet
	var err error

	for {
		rtp, _, err = remoteTrack.ReadRTP()
		if err != nil {
			if peer.checkClose() {
				return
			}
			continue
		}

		if localTrack == nil {
			logs.Warn("Local track is nil")
			return
		}
		if err := localTrack.WriteRTP(rtp); err != nil {
			logs.Error(err.Error())
		}

		logs.Stack(fmt.Sprintf("Push %s rtp pkg to localTrack %s", remoteTrack.Kind().String(), peer.getSessionID()))
		rtp = nil
		err = nil
		// fwd = nil
	}
}
