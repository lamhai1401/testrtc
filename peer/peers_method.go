package peer

import (
	"fmt"
	"time"

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

		// remove from fwdm
		if fwdm := p.getVideoFwdm(); fwdm != nil {
			fwdm.Unregister(streamID, client.getSessionID())
		}
		if fwdm := p.getAudioFwdm(); fwdm != nil {
			fwdm.Unregister(streamID, client.getSessionID())
		}
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

func (p *Peers) sendCandidate(singalID, streamID, role, sessionID string, candidate *webrtc.ICECandidate) {
	p.sendSignal(singalID, streamID, role, sessionID, "candidate", candidate.ToJSON())
}

// func (p *Peers) sendSDP(singalID, streamID, role, sessionID string, sdp interface{}) {
// 	p.sendSignal(singalID, streamID, role, sessionID, "SDP", sdp)
// }

// func (p *Peers) sendOK(singalID, streamID, role, sessionID string) {
// 	p.sendSignal(singalID, streamID, role, sessionID, "ok")
// }

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
	p.setState(streamID, state)

	switch state {
	case "connected":
		if !peer.IsConnected() { // notif if this is new peer
			peer.SetIsConnected(true)
			// set data state
			// p.setPeerDataState(peer.getSessionID(), NOTYET)
			// call peer success
			// p.handleTransceiver(peer)
			if handleAddPeer != nil {
				handleAddPeer(signalID, streamID, peer.getRole(), peer.getSessionID())
			}
		}
		break
	case "failed":
		go func() {
			if err := p.checkFailedState(streamID, peer.getSessionID()); err != nil {
				logs.Warn(fmt.Sprintf("Remove old peer connection (%s_%s_%s) has state %s", signalID, streamID, peer.getSessionID(), state))
				p.RemoveConnection(peer.getStreamID())
				if handleFailedPeer != nil {
					handleFailedPeer(p.getSignalID(), streamID, peer.getRole(), peer.getSessionID())
				}
			}
		}()
		break
	// case "closed":
	// 	sessionID := peer.getSessionID()
	// 	logs.Info(fmt.Sprintf("%s_%s_%s ice state is %s", signalID, streamID, peer.getSessionID(), state))
	// 	if conn := p.GetConnection(streamID); conn != nil {
	// 		if sessionID == peer.getSessionID() {
	// 			logs.Warn(fmt.Sprintf("Remove old peer connection (%s_%s_%s) has state %s", signalID, streamID, peer.getSessionID(), state))
	// 			p.RemoveConnection(peer.getStreamID())
	// 			// p.RemoveConnections(p.getSignalID())
	// 		}
	// 	}
	// 	break
	default:
		return
	}
}

// func (p *Peers) handleTransceiver(peer *Peer) {
// 	switch peer.getRole() {
// 	case "self":
// 		go p.pushBack(peer.getVideoTransceiver().Receiver().Track(), peer.getLocalVideoTrack(), peer)
// 		// go p.pushBack(peer.getAudioTransceiver().Receiver().Track(), peer.getLocalAudioTrack(), peer)
// 		break
// 	case "source":
// 		// go p.pushToFwd(p.getAudioFwdm(), peer.getAudioTransceiver().Receiver().Track(), peer.getStreamID(), "audio", peer)
// 		go p.pushToFwd(p.getVideoFwdm(), peer.getVideoTransceiver().Receiver().Track(), peer.getStreamID(), "video", peer)
// 		break
// 	default:
// 		// logs.Info(fmt.Sprintf("Current %s track has role %s. Not is source/self role. No need to read RTP", kind, peer.getRole()))
// 		return
// 	}
// }

// handle peer remotetrack with streamID
func (p *Peers) handleOnTrack(remoteTrack *webrtc.TrackRemote, peer *Peer) {
	kind := remoteTrack.Kind().String()
	var fwdm utils.Fwdm
	// var localTrack *webrtc.TrackLocalStaticRTP

	logs.Warn("Remote track STREAMID: ====[ ", remoteTrack.StreamID(), " ]====")
	logs.Warn("Remote track ID: ====[ ", remoteTrack.ID(), " ]====")
	logs.Warn("Remote track RID: ====[ ", remoteTrack.RID(), " ]====")
	logs.Debug(fmt.Sprintf("Has %s remote track of id %s_%s", kind, p.getSignalID(), peer.getStreamID()))

	switch kind {
	case "video":
		peer.HandleVideoTrack(remoteTrack)
		// peer.setRemoteVideoTrack(remoteTrack)
		fwdm = p.getVideoFwdm()
		// localTrack = peer.getLocalVideoTrack()
	case "audio":
		// peer.setRemoteAudioTrack(remoteTrack)
		fwdm = p.getAudioFwdm()
		// localTrack = peer.getLocalAudioTrack()
	default:
		return
	}

	switch peer.getRole() {
	// case "self":
	// 	go p.pushBack(remoteTrack, localTrack, peer)
	// 	break
	case "source":
		go p.pushToFwd(fwdm, remoteTrack, peer.getStreamID(), kind, peer)
	default:
		logs.Info(fmt.Sprintf("Current %s track has role %s. Not is source/self role. No need to read RTP", kind, peer.getRole()))
		return
	}
}

var lastVideoHeader *rtp.Header
var lastVideoSequenceNumber uint16

var lastAudioHeader *rtp.Header
var lastAudioSequenceNumber uint16

var lastTimestamp uint32

func (p *Peers) pushToFwd(fwdm utils.Fwdm, remoteTrack *webrtc.TrackRemote, streamID, kind string, peer *Peer) {
	var pkg *rtp.Packet
	var err error

	for {
		pkg, _, err = remoteTrack.ReadRTP()
		if err != nil {
			if peer.checkClose() {
				return
			}
			continue
		}

		if kind == "video" {
			// Change the timestamp to only be the delta
			// oldTimestamp := pkg.Timestamp
			// if lastTimestamp == 0 {
			// 	pkg.Timestamp = 0
			// } else {
			// 	pkg.Timestamp -= lastTimestamp
			// }
			// lastTimestamp = oldTimestamp

			if lastVideoHeader == nil {
				lastVideoHeader = &pkg.Header
				// lastTimestamp = pkg.Timestamp
				// lastVideoHeader.Timestamp = pkg.Timestamp
			}
		}

		if kind == "audio" && lastAudioHeader == nil {
			lastAudioHeader = &pkg.Header
			// lastAudioHeader.Timestamp = pkg.Timestamp
		}

		// push video to fwd
		fwd := fwdm.GetForwarder(streamID)
		if fwd == nil {
			fwd = fwdm.AddNewForwarder(streamID)
		}

		if fwd != nil {
			if kind == "video" {
				fwd.Push(&utils.Wrapper{
					Pkg: *p.formatVideoData(pkg),
				})
			} else {
				fwd.Push(&utils.Wrapper{
					Pkg: *p.formatAudioData(pkg),
				})
			}
			logs.Stack(fmt.Sprintf("Push %s rtp pkg to fwd %s", kind, streamID))
		}

		pkg = nil
		err = nil
		fwd = nil
	}
}

var currTimestamp uint32

func (p *Peers) formatVideoData(data *rtp.Packet) *rtp.Packet {
	// currTimestamp += data.Timestamp
	// data.Timestamp = currTimestamp
	if lastVideoSequenceNumber == 0 {
		lastVideoSequenceNumber = data.SequenceNumber
		return data
	}

	data.SequenceNumber = lastVideoSequenceNumber + 1
	// if math.Abs(float64(data.SequenceNumber)-float64(lastVideoSequenceNumber)) > 1 {
	// 	data.SequenceNumber = lastVideoSequenceNumber + 1
	// }
	lastVideoSequenceNumber = data.SequenceNumber
	data.PayloadOffset = lastVideoHeader.PayloadOffset
	// data.Timestamp = lastTimestamp

	data.Extension = lastVideoHeader.Extension
	data.ExtensionProfile = lastVideoHeader.ExtensionProfile
	data.Extensions = lastVideoHeader.Extensions
	return data
}

func (p *Peers) formatAudioData(data *rtp.Packet) *rtp.Packet {
	if lastAudioSequenceNumber == 0 {
		lastAudioSequenceNumber = data.SequenceNumber
		return data
	}

	data.SequenceNumber = lastAudioSequenceNumber + 1

	// if math.Abs(float64(data.SequenceNumber)-float64(lastAudioSequenceNumber)) > 1 {
	// 	data.SequenceNumber = lastAudioSequenceNumber + 1
	// }
	lastAudioSequenceNumber = data.SequenceNumber
	data.PayloadOffset = lastAudioHeader.PayloadOffset
	data.SSRC = lastAudioHeader.SSRC
	// data.Timestamp = lastVideoHeader.Timestamp
	// data.Extension = lastAudioHeader.Extension
	// data.ExtensionProfile = lastAudioHeader.ExtensionProfile
	// data.Extensions = lastAudioHeader.Extensions
	return data
}

func (p *Peers) pushBack(remoteTrack *webrtc.TrackRemote, localTrack *webrtc.TrackLocalStaticRTP, peer *Peer) {
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

		if localTrack != nil {
			if err := localTrack.WriteRTP(rtp); err != nil {
				logs.Error(err.Error())
				return
			}
		}

		rtp = nil
		err = nil
	}
}

func (p *Peers) checkFailedState(streamID, sessionID string) error {
	time.Sleep(10 * time.Second)
	state := p.getState(streamID)
	peer := p.GetConnection(streamID)
	if peer == nil {
		return nil
	}
	if (state == "failed" || state == "disconnected" || state == "closed") && sessionID == peer.GetSessionID() {
		str := fmt.Sprintf("%s state still %s after 10s", streamID, state)
		logs.Error(str)
		return fmt.Errorf(str)
	}
	return nil
}

func (p *Peers) getState(streamID string) string {
	if states := p.getStates(); states != nil {
		state, has := states.Get(streamID)
		if !has {
			return ""
		}
		s, ok := state.(string)
		if ok {
			return s
		}
	}
	return ""
}

func (p *Peers) getStates() *utils.AdvanceMap {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.states
}

func (p *Peers) deleteState(streamID string) {
	if states := p.getStates(); states != nil {
		states.Delete(streamID)
	}
}

// exportStates iter state and export into a map string
func (p *Peers) exportStates() map[string]string {
	temp := make(map[string]string, 0)
	if states := p.getStates(); states != nil {
		states.Iter(func(key, value interface{}) bool {
			k, ok1 := key.(string)
			v, ok2 := value.(string)
			if ok1 && ok2 {
				temp[k] = v
			}
			return true
		})
	}
	return temp
}

func (p *Peers) setState(streamID string, state string) {
	if states := p.getStates(); states != nil {
		states.Set(streamID, state)
	}
}

func (p *Peers) getAudioFwdm() utils.Fwdm {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.audioFwdm
}

func (p *Peers) getVideoFwdm() utils.Fwdm {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.videoFwdm
}
