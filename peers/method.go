package peers

import (
	"fmt"
	"io"

	"github.com/beowulflab/signal/signal-wss"
	"github.com/lamhai1401/gologs/logs"
	"github.com/lamhai1401/testrtc/peer"
	"github.com/lamhai1401/testrtc/streams"
	"github.com/lamhai1401/testrtc/utils"
	"github.com/pion/webrtc/v3"
)

func (ps *Peers) getID() string {
	ps.mutex.RLock()
	defer ps.mutex.RUnlock()
	return ps.id
}

func (ps *Peers) getSignal() *signal.NotifySignal {
	ps.mutex.RLock()
	defer ps.mutex.RUnlock()
	return ps.signal
}

func (ps *Peers) sendOk(id, session string) {
	if signal := ps.getSignal(); signal != nil {
		signal.Send(id, session, "ok")
	}
}

func (ps *Peers) sendSDP(id, session string, sdp interface{}) {
	if signal := ps.getSignal(); signal != nil {
		signal.Send(id, session, "sdp", sdp)
	}
}

func (ps *Peers) sendCandidate(id, session string, candidate interface{}) {
	if signal := ps.getSignal(); signal != nil {
		signal.Send(id, session, "candidate", candidate)
	}
}

func (ps *Peers) sendError(id, session string, reason interface{}) {
	if signal := ps.getSignal(); signal != nil {
		signal.Send(id, session, "error", reason)
	}
}

func (ps *Peers) getConns() *utils.AdvanceMap {
	ps.mutex.RLock()
	defer ps.mutex.RUnlock()
	return ps.conns
}

func (ps *Peers) getConn(id string) *peer.Peer {
	if conns := ps.getConns(); conns != nil {
		conn, has := conns.Get(id)
		if has {
			peer, ok := conn.(*peer.Peer)
			if ok {
				return peer
			}
		}
	}
	return nil
}

func (ps *Peers) setConn(id string, peer *peer.Peer) {
	if conns := ps.getConns(); conns != nil {
		conns.Set(id, peer)
	}
}

func (ps *Peers) getAudioFwd() utils.Fwdm {
	ps.mutex.RLock()
	defer ps.mutex.RUnlock()
	return ps.audioFwdm
}

func (ps *Peers) getVideoFwd() utils.Fwdm {
	ps.mutex.RLock()
	defer ps.mutex.RUnlock()
	return ps.videoFwdm
}

func (ps *Peers) getConfig() *webrtc.Configuration {
	ps.mutex.RLock()
	defer ps.mutex.RUnlock()
	return ps.configs
}

func (ps *Peers) register(id string, clientID string, handler func(wrapper *utils.Wrapper) error) {
	if fwdm := ps.getVideoFwd(); fwdm != nil {
		fwdm.Register(id, clientID, handler)
	}
}

func (ps *Peers) unRegister(id, clientID string) {
	if fwdm := ps.getVideoFwd(); fwdm != nil {
		fwdm.Unregister(id, clientID)
	}
}

func (ps *Peers) addConn(id, session string) (*peer.Peer, error) {
	peer := peer.NewPeer(&ps.bitrate, session, id)
	ps.setConn(id, peer)
	return peer, nil
}

func (ps *Peers) deleteConn(id string) {
	if conns := ps.getConns(); conns != nil {
		conns.Delete(id)
	}
}

func (ps *Peers) closeConn(id string) {
	if conn := ps.getConn(id); conn != nil {
		ps.deleteConn(id)
		conn.Close()
		ps.unRegister(ps.getID(), conn.GetSignalID())

		if mixer := ps.getVideoMixer(); mixer != nil {
			mixer.RemoveVideo(conn.GetSignalID())
		}
		conn = nil
	}
}

func (ps *Peers) handleConnEvent(peer *peer.Peer) {
	conn := peer.GetConn()

	conn.OnICECandidate(func(i *webrtc.ICECandidate) {
		if i == nil {
			return
		}
		ps.sendCandidate(peer.GetSignalID(), peer.GetSessionID(), i.ToJSON())
	})

	conn.OnICEConnectionStateChange(func(is webrtc.ICEConnectionState) {
		state := is.String()
		logs.Warn(fmt.Sprintf("Connection %s has states %s", peer.GetSignalID(), state))
		switch state {
		case "connected":
			peer.SetConnected()
			// register video source
			ps.register(mixerID, peer.GetSignalID(), func(wrapper *utils.Wrapper) error {
				err := peer.AddVideoRTP(&wrapper.Pkg)
				if err != nil {
					return err
				}
				logs.Stack(fmt.Sprintf("Write mixer video rtp to %s", peer.GetSignalID()))
				return nil
			})
			break
		case "closed":
			ps.closeConn(peer.GetSignalID())
			break
		case "failed":
			ps.closeConn(peer.GetSignalID())
			break
		default:
			break
		}
	})

	conn.OnTrack(func(remoteTrack *webrtc.TrackRemote, r *webrtc.RTPReceiver) {
		kind := remoteTrack.Kind().String()
		logs.Info(fmt.Sprintf("Has remote %s track of ID %s", kind, peer.GetSignalID()))

		go peer.RapidResynchronizationRequest(remoteTrack)
		go peer.ModifyBitrate(remoteTrack)
		go peer.PictureLossIndication(remoteTrack)

		fmt.Printf("Track has started, of type %d: %s \n", remoteTrack.PayloadType(), remoteTrack.Codec().MimeType)

		// register to video mixer
		if kind == "video" {
			// start to register index
			// mixer := ps.getVideoMixer()
			// index, err := mixer.AddVideo(peer.GetSignalID())
			// if err != nil {
			// 	logs.Error(err.Error())
			// 	return
			// }
			for {
				// Read RTP packets being sent to Pion
				rtp, _, readErr := remoteTrack.ReadRTP()
				if readErr != nil {
					if readErr == io.EOF {
						return
					}
					panic(readErr)
				}

				// logs.Stack(fmt.Sprintf("Push %s video data to %d chann", peer.GetSignalID(), index))
				// mixer.PushVideo(index, rtp)

				// write video data
				err := peer.AddVideoRTP(rtp)
				if err != nil {
					logs.Error("AddVideoRTP: ", err)
				}

				rtp = nil
			}
		}
	})
}

func (ps *Peers) getVideoMixer() *streams.VideoStreamObj {
	ps.mutex.RLock()
	defer ps.mutex.RUnlock()
	return ps.videoMixer
}
