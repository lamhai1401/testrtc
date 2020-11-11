package peer

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/lamhai1401/gologs/logs"
	"github.com/lamhai1401/testrtc/utils"
	"github.com/pion/rtcp"
	"github.com/pion/rtp"
	"github.com/pion/webrtc/v2"
)

func (p *Peer) getLocalAudioTrack() *webrtc.Track {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.localAudioTrack
}

func (p *Peer) setLocalAudioTrack(t *webrtc.Track) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.localAudioTrack = t
}

func (p *Peer) getLocalVideoTrack() *webrtc.Track {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.localVideoTrack
}

func (p *Peer) setLocalVideoTrack(t *webrtc.Track) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.localVideoTrack = t
}

func (p *Peer) getSessionID() string {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.sessionID
}

func (p *Peer) getBitrate() *int {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.bitrate
}

func (p *Peer) getSignalID() string {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.signalID
}

func (p *Peer) closeConn() {
	if conn := p.getConn(); conn != nil {
		p.setConn(nil)
		conn.Close()
		conn = nil
	}
}

func (p *Peer) getConn() *webrtc.PeerConnection {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.conn
}

func (p *Peer) setConn(c *webrtc.PeerConnection) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.conn = c
}

// PictureLossIndication packet informs the encoder about the loss of an undefined amount of coded video data belonging to one or more pictures
func (p *Peer) pictureLossIndication(remoteTrack *webrtc.Track) {
	ticker := time.NewTicker(time.Millisecond * 500)
	for range ticker.C {
		if p.checkClose() {
			return
		}

		conn := p.getConn()
		if conn == nil {
			return
		}

		errSend := conn.WriteRTCP([]rtcp.Packet{&rtcp.PictureLossIndication{MediaSSRC: remoteTrack.SSRC()}})
		if errSend != nil {
			logs.Error("Picture loss indication write rtcp err: ", errSend.Error())
			// return
		}
	}
}

// RapidResynchronizationRequest packet informs the encoder about the loss of an undefined amount of coded video data belonging to one or more pictures
func (p *Peer) rapidResynchronizationRequest(remoteTrack *webrtc.Track) {
	ticker := time.NewTicker(time.Millisecond * 100)
	for range ticker.C {
		if p.checkClose() {
			return
		}

		conn := p.getConn()
		if conn == nil {
			return
		}
		if routineErr := conn.WriteRTCP([]rtcp.Packet{&rtcp.RapidResynchronizationRequest{SenderSSRC: remoteTrack.SSRC(), MediaSSRC: remoteTrack.SSRC()}}); routineErr != nil {
			logs.Error("rapidResynchronizationRequest write rtcp err: ", routineErr.Error())
			// return
		}
	}
}

// ModifyBitrate so set bitrate when datachannel has signal
// Use this only for video not audio track
func (p *Peer) modifyBitrate(remoteTrack *webrtc.Track) {
	ticker := time.NewTicker(time.Millisecond * 500)
	for range ticker.C {
		bitrate := p.getBitrate()
		if p.checkClose() || bitrate == nil {
			return
		}

		numbers := (*bitrate) * 1024
		if conn := p.getConn(); conn != nil {
			errSend := conn.WriteRTCP([]rtcp.Packet{&rtcp.ReceiverEstimatedMaximumBitrate{
				SenderSSRC: remoteTrack.SSRC(),
				Bitrate:    uint64(numbers),
				// SSRCs:      []uint32{rand.Uint32()},
			}})

			if errSend != nil {
				logs.Error("Modify bitrate write rtcp err: ", errSend.Error())
				// return
			}
		}
	}
}

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

func (p *Peer) writeRTP(packet *rtp.Packet, track *webrtc.Track) error {
	// packet.PayloadType = track.PayloadType()
	packet.SSRC = track.SSRC()
	packet.Header.PayloadType = track.PayloadType()
	return track.WriteRTP(packet)
}

// CreateAudioTrack linter
func (p *Peer) createAudioTrack(trackID string, codesc uint8) error {
	if conn := p.getConn(); conn != nil {
		localTrack, err := conn.NewTrack(codesc, rand.Uint32(), trackID, trackID)
		if err != nil {
			return err
		}
		// Add this newly created track to the PeerConnection
		_, err = conn.AddTrack(localTrack)
		if err != nil {
			return err
		}
		p.setLocalAudioTrack(localTrack)
		return nil
	}
	return fmt.Errorf("cannot create audio track because rtc connection is nil")
}

// CreateVideoTrack linter
func (p *Peer) createVideoTrack(trackID string, codesc uint8) error {
	if conn := p.getConn(); conn != nil {
		localTrack, err := conn.NewTrack(codesc, rand.Uint32(), trackID, trackID)
		if err != nil {
			return err
		}
		// Add this newly created track to the PeerConnection
		_, err = conn.AddTrack(localTrack)
		if err != nil {
			return err
		}
		p.setLocalVideoTrack(localTrack)
		return nil
	}
	return fmt.Errorf("cannot create video track because rtc connection is nil")
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
		ice, ok := value.(*webrtc.ICECandidateInit)
		if ok {
			if err := conn.AddICECandidate(*ice); err != nil {
				return err
			}
		}
	}
	return nil
}

// NewAPI linter
func (p *Peer) addAPI() *webrtc.API {
	return webrtc.NewAPI(webrtc.WithMediaEngine(*p.initMediaEngine()), webrtc.WithSettingEngine(*p.initSettingEngine()))
}

func (p *Peer) initMediaEngine() *webrtc.MediaEngine {
	mediaEngine := &webrtc.MediaEngine{}
	mediaEngine.RegisterCodec(webrtc.NewRTPOpusCodec(defaultAudioCodecs, 48000))
	mediaEngine.RegisterCodec(webrtc.NewRTPVP8Codec(defaultVideoCodecs, 90000))
	return mediaEngine
}

func (p *Peer) initSettingEngine() *webrtc.SettingEngine {
	settingEngine := &webrtc.SettingEngine{}
	// settingEngine.SetEphemeralUDPPortRange(20000, 60000)
	// settingEngine.SetICETimeouts(10*time.Second, 20*time.Second, 1*time.Second)
	return settingEngine
}
