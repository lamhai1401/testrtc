package conn

import (
	"fmt"
	"time"

	log "github.com/lamhai1401/gologs/logs"
	"github.com/lamhai1401/testrtc/utils"
	"github.com/pion/rtcp"
	"github.com/pion/rtp"
	"github.com/pion/webrtc/v3"
)

func (p *Peer) checkClose() bool {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.isClosed
}

func (p *Peer) setClose(state bool) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.isClosed = state
}

func (p *Peer) writeRTP(packet *rtp.Packet, track *webrtc.TrackLocalStaticRTP) error {
	return track.WriteRTP(packet)
}

// SetBitrate linter
func (p *Peer) setBitrate(bitrate *int) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.bitrate = bitrate
}

func (p *Peer) getOffeerID() string {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.offerID
}

func (p *Peer) getCookieID() string {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.cookieID
}

func (p *Peer) getHasVideoTrack() bool {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.hasVideo
}

func (p *Peer) setHasVideoTrack(state bool) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.hasVideo = state
}

func (p *Peer) getHasAudioTrack() bool {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.hasAudio
}

func (p *Peer) setHasAudioTrack(state bool) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.hasAudio = state
}

func (p *Peer) setConn(conn *webrtc.PeerConnection) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.conn = conn
}

func (p *Peer) getConn() *webrtc.PeerConnection {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.conn
}

func (p *Peer) closeConn() error {
	if conn := p.getConn(); conn != nil {
		p.setConn(nil)
		err := conn.Close()
		if err != nil {
			return err
		}
		conn = nil
	}
	return nil
}

func (p *Peer) getCodecs() string {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.codec
}

func (p *Peer) setCodecs(c string) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.codec = c
}

func (p *Peer) getPayloadType() int {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.payloadType
}

func (p *Peer) setPayloadType(c int) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.payloadType = c
}

func (p *Peer) getBitrate() *int {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.bitrate
}

// ModifyBitrate so set bitrate when datachannel has signal
// Use this only for video not audio track
func (p *Peer) modifyBitrate(remoteTrack *webrtc.TrackRemote) {
	ticker := time.NewTicker(time.Millisecond * 500)
	for range ticker.C {
		bitrate := p.getBitrate()
		if p.checkClose() || bitrate == nil {
			return
		}

		numbers := (*bitrate) * 1024
		if conn := p.getConn(); conn != nil {
			errSend := conn.WriteRTCP([]rtcp.Packet{&rtcp.ReceiverEstimatedMaximumBitrate{
				SenderSSRC: uint32(remoteTrack.SSRC()),
				Bitrate:    uint64(numbers),
				// SSRCs:      []uint32{rand.Uint32()},
			}})

			if errSend != nil {
				log.Error("Modify bitrate write rtcp err: ", errSend.Error())
				// return
			}
		}
	}
}

// PictureLossIndication packet informs the encoder about the loss of an undefined amount of coded video data belonging to one or more pictures
func (p *Peer) pictureLossIndication(remoteTrack *webrtc.TrackRemote) {
	ticker := time.NewTicker(time.Millisecond * 500)
	for range ticker.C {
		if p.checkClose() {
			return
		}

		conn := p.getConn()
		if conn == nil {
			return
		}
		errSend := conn.WriteRTCP([]rtcp.Packet{
			&rtcp.PictureLossIndication{MediaSSRC: uint32(remoteTrack.SSRC())},
			&rtcp.SliceLossIndication{MediaSSRC: uint32(remoteTrack.SSRC())},
			&rtcp.RapidResynchronizationRequest{SenderSSRC: uint32(remoteTrack.SSRC()), MediaSSRC: uint32(remoteTrack.SSRC())},
		})

		if errSend != nil {
			log.Error("Picture loss indication write rtcp err: ", errSend.Error())
			// return
		}
	}
}

func (p *Peer) getStreamID() string {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.streamID
}

func (p *Peer) setStreamID(s string) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.streamID = s
}

func (p *Peer) getSessionID() string {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.sessionID
}

func (p *Peer) setSessionID(s string) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.sessionID = s
}

func (p *Peer) getTracks() *Tracks {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	return p.tracks
}

func (p *Peer) getRole() string {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.role
}

// GetLocalDescription get current peer local description
func (p *Peer) getLocalDescription() (*webrtc.SessionDescription, error) {
	conn := p.getConn()
	if conn == nil {
		return nil, fmt.Errorf("rtc connection is nil")
	}
	return conn.LocalDescription(), nil
}

func (p *Peer) getIceCache() *utils.AdvanceMap {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.iceCache
}

func (p *Peer) addIceCache(ice *webrtc.ICECandidateInit) {
	if cache := p.getIceCache(); cache != nil {
		cache.Set(ice.Candidate, ice)
	}
}

// setCacheIce add ice save in cache
func (p *Peer) setCacheIce() error {
	cache := p.getIceCache()
	if cache == nil {
		return fmt.Errorf("ICE cache map is nil")
	}
	conn := p.getConn()
	if conn == nil {
		return fmt.Errorf("Peer connection is nil")
	}

	captureCache := cache.Capture()
	for _, value := range captureCache {
		// ice, ok := value.(*webrtc.ICECandidateInit)
		// if ok {
		// 	if err := p.AddICECandidate(ice); err != nil {
		// 		return err
		// 	}
		// }
		if err := p.AddICECandidate(value); err != nil {
			return err
		}
	}
	return nil
}

// AddOffer add client offer and return answer
func (p *Peer) addOffer(offer *webrtc.SessionDescription) error {
	conn := p.getConn()
	if conn == nil {
		return fmt.Errorf("rtc connection is nil")
	}

	//set remote desc
	err := conn.SetRemoteDescription(*offer)
	if err != nil {
		return err
	}

	err = p.setCacheIce()
	if err != nil {
		return err
	}

	err = p.CreateAnswer()
	if err != nil {
		return err
	}

	return nil
}

// AddAnswer add client answer and set remote desc
func (p *Peer) addAnswer(answer *webrtc.SessionDescription) error {
	conn := p.getConn()
	if conn == nil {
		return fmt.Errorf("Peer connection is nil")
	}

	//set remote desc
	err := conn.SetRemoteDescription(*answer)
	if err != nil {
		return err
	}
	return p.setCacheIce()
}
