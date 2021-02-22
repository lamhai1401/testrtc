package peer

import (
	"fmt"
	"strings"
	"time"

	"github.com/lamhai1401/gologs/logs"
	"github.com/lamhai1401/testrtc/utils"
	"github.com/pion/rtcp"
	"github.com/pion/rtp"
	"github.com/pion/webrtc/v3"
)

func (p *Peer) getLocalAudioTrack() *webrtc.TrackLocalStaticRTP {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.localAudioTrack
}

func (p *Peer) setLocalAudioTrack(t *webrtc.TrackLocalStaticRTP) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.localAudioTrack = t
}

func (p *Peer) getLocalVideoTrack() *webrtc.TrackLocalStaticRTP {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.localVideoTrack
}

func (p *Peer) setLocalVideoTrack(t *webrtc.TrackLocalStaticRTP) {
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
func (p *Peer) PictureLossIndication(remoteTrack *webrtc.TrackRemote) {
	ticker := time.NewTicker(time.Millisecond * 500)
	for range ticker.C {
		if p.checkClose() {
			return
		}

		conn := p.getConn()
		if conn == nil {
			return
		}

		errSend := conn.WriteRTCP([]rtcp.Packet{&rtcp.PictureLossIndication{MediaSSRC: uint32(remoteTrack.SSRC())}})
		if errSend != nil {
			logs.Error("Picture loss indication write rtcp err: ", errSend.Error())
			// return
		}
	}
}

// RapidResynchronizationRequest packet informs the encoder about the loss of an undefined amount of coded video data belonging to one or more pictures
func (p *Peer) RapidResynchronizationRequest(remoteTrack *webrtc.TrackRemote) {
	ticker := time.NewTicker(time.Millisecond * 100)
	for range ticker.C {
		if p.checkClose() {
			return
		}

		conn := p.getConn()
		if conn == nil {
			return
		}
		if routineErr := conn.WriteRTCP([]rtcp.Packet{&rtcp.RapidResynchronizationRequest{SenderSSRC: uint32(remoteTrack.SSRC()), MediaSSRC: uint32(remoteTrack.SSRC())}}); routineErr != nil {
			logs.Error("rapidResynchronizationRequest write rtcp err: ", routineErr.Error())
			// return
		}
	}
}

// ModifyBitrate so set bitrate when datachannel has signal
// Use this only for video not audio track
func (p *Peer) ModifyBitrate(remoteTrack *webrtc.TrackRemote) {
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

func (p *Peer) writeRTP(packet *rtp.Packet, track *webrtc.TrackLocalStaticRTP) error {
	// packet.PayloadType = track.PayloadType()
	// packet.SSRC = track.SSRC()
	// packet.Header.PayloadType = track.PayloadType()

	// writeErr := track.wr(packet)
	// if writeErr != nil && !errors.Is(writeErr, io.ErrClosedPipe) {
	// 	panic(writeErr)
	// }
	return track.WriteRTP(packet)
}

// CreateAudioTrack linter
func (p *Peer) createAudioTrack(trackID string) error {
	if conn := p.getConn(); conn != nil {
		// localTrack, err := conn.NewTrack(codesc, rand.Uint32(), trackID, trackID)
		// if err != nil {
		// 	return err
		// }
		localTrack, err := webrtc.NewTrackLocalStaticRTP(
			webrtc.RTPCodecCapability{MimeType: "audio/opus"},
			fmt.Sprintf(trackID),
			fmt.Sprintf(trackID),
		)
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
func (p *Peer) createVideoTrack(trackID string) error {
	if conn := p.getConn(); conn != nil {
		localTrack, err := webrtc.NewTrackLocalStaticRTP(
			webrtc.RTPCodecCapability{MimeType: "video/VP9"},
			trackID,
			trackID,
		)
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
	return webrtc.NewAPI(webrtc.WithMediaEngine(p.initMediaEngine()), webrtc.WithSettingEngine(*p.initSettingEngine()))
}

func (p *Peer) initMediaEngineAudio(mediaEngine *webrtc.MediaEngine) error {
	// if err := mediaEngine.RegisterCodec(webrtc.RTPCodecParameters{
	// 	RTPCodecCapability: webrtc.RTPCodecCapability{MimeType: "audio/opus", ClockRate: 48000, Channels: 2, SDPFmtpLine: "minptime=10;useinbandfec=1", RTCPFeedback: nil},
	// 	PayloadType:        111,
	// }, webrtc.RTPCodecTypeAudio); err != nil {
	// 	logs.Error("initMediaEngine: ", err)
	// }

	// Default Pion Audio Codecs
	for _, codec := range []webrtc.RTPCodecParameters{
		{
			RTPCodecCapability: webrtc.RTPCodecCapability{MimeType: MimeTypeOpus, ClockRate: 48000, Channels: 2, SDPFmtpLine: "minptime=10;useinbandfec=1", RTCPFeedback: nil},
			PayloadType:        111,
		},
		{
			RTPCodecCapability: webrtc.RTPCodecCapability{MimeType: MimeTypeG722, ClockRate: 8000, Channels: 0, SDPFmtpLine: "", RTCPFeedback: nil},
			PayloadType:        9,
		},
		{
			RTPCodecCapability: webrtc.RTPCodecCapability{MimeType: MimeTypePCMU, ClockRate: 8000, Channels: 0, SDPFmtpLine: "", RTCPFeedback: nil},
			PayloadType:        0,
		},
		{
			RTPCodecCapability: webrtc.RTPCodecCapability{MimeType: MimeTypePCMA, ClockRate: 8000, Channels: 0, SDPFmtpLine: "", RTCPFeedback: nil},
			PayloadType:        8,
		},
	} {
		if err := mediaEngine.RegisterCodec(codec, webrtc.RTPCodecTypeAudio); err != nil {
			return err
		}
	}

	// Default Pion Audio Header Extensions
	for _, extension := range []string{
		"urn:ietf:params:rtp-hdrext:sdes:mid",
		"urn:ietf:params:rtp-hdrext:sdes:rtp-stream-id",
		"urn:ietf:params:rtp-hdrext:sdes:repaired-rtp-stream-id",
	} {
		if err := mediaEngine.RegisterHeaderExtension(webrtc.RTPHeaderExtensionCapability{URI: extension}, webrtc.RTPCodecTypeAudio); err != nil {
			logs.Error("initMediaEngineAudio: ", err)
		}
	}
	return nil
}

func (p *Peer) initMediaEngineVideo(codecCode int, mediaEngine *webrtc.MediaEngine) error {
	var tmp []webrtc.RTPCodecParameters
	videoRTCPFeedback := []webrtc.RTCPFeedback{{Type: "goog-remb", Parameter: ""}, {Type: "ccm", Parameter: "fir"}, {Type: "nack", Parameter: ""}, {Type: "nack", Parameter: "pli"}}

	switch strings.ToLower("vp9") {
	case "vp8":
		tmp = []webrtc.RTPCodecParameters{
			{
				RTPCodecCapability: webrtc.RTPCodecCapability{MimeType: MimeTypeVP8, ClockRate: 90000, Channels: 0, SDPFmtpLine: "", RTCPFeedback: videoRTCPFeedback},
				PayloadType:        webrtc.PayloadType(codecCode),
			},
			{
				RTPCodecCapability: webrtc.RTPCodecCapability{MimeType: "video/rtx", ClockRate: 90000, Channels: 0, SDPFmtpLine: fmt.Sprintf("apt=%d", codecCode), RTCPFeedback: nil},
				PayloadType:        webrtc.PayloadType(codecCode + 1),
			},
		}
		break
	default: // defaul vp9
		tmp = []webrtc.RTPCodecParameters{
			{
				RTPCodecCapability: webrtc.RTPCodecCapability{MimeType: "video/VP9", ClockRate: 90000, Channels: 0, SDPFmtpLine: "profile-id=0", RTCPFeedback: videoRTCPFeedback},
				PayloadType:        webrtc.PayloadType(codecCode),
			},
			{
				RTPCodecCapability: webrtc.RTPCodecCapability{MimeType: "video/rtx", ClockRate: 90000, Channels: 0, SDPFmtpLine: fmt.Sprintf("apt=%d", codecCode), RTCPFeedback: nil},
				PayloadType:        webrtc.PayloadType(codecCode + 1),
			},

			{
				RTPCodecCapability: webrtc.RTPCodecCapability{MimeType: "video/VP9", ClockRate: 90000, Channels: 0, SDPFmtpLine: "profile-id=1", RTCPFeedback: videoRTCPFeedback},
				PayloadType:        webrtc.PayloadType(codecCode + 2),
			},
			{
				RTPCodecCapability: webrtc.RTPCodecCapability{MimeType: "video/rtx", ClockRate: 90000, Channels: 0, SDPFmtpLine: fmt.Sprintf("apt=%d", codecCode+2), RTCPFeedback: nil},
				PayloadType:        webrtc.PayloadType(codecCode + 3),
			},

			{
				RTPCodecCapability: webrtc.RTPCodecCapability{MimeType: "video/ulpfec", ClockRate: 90000, Channels: 0, SDPFmtpLine: "", RTCPFeedback: nil},
				PayloadType:        116,
			},
		}
		break
	}

	// Default Pion Audio Codecs
	for _, codec := range tmp {
		if err := mediaEngine.RegisterCodec(codec, webrtc.RTPCodecTypeVideo); err != nil {
			logs.Error("initMediaEngine: ", err)
		}
	}

	// Default Pion Video Header Extensions
	for _, extension := range []string{
		"urn:ietf:params:rtp-hdrext:sdes:mid",
		"urn:ietf:params:rtp-hdrext:sdes:rtp-stream-id",
		"urn:ietf:params:rtp-hdrext:sdes:repaired-rtp-stream-id",
	} {
		if err := mediaEngine.RegisterHeaderExtension(webrtc.RTPHeaderExtensionCapability{URI: extension}, webrtc.RTPCodecTypeVideo); err != nil {
			logs.Error("initMediaEngine: ", err)
		}
	}
	return nil
}

func (p *Peer) initMediaEngine() *webrtc.MediaEngine {
	mediaEngine := &webrtc.MediaEngine{}
	// mediaEngine.RegisterCodec(webrtc.NewRTPOpusCodec(webrtc.DefaultPayloadTypeOpus, 48000))
	// mediaEngine.RegisterCodec(webrtc.NewRTPVP8Codec(webrtc.DefaultPayloadTypeVP8, 90000))

	err := p.initMediaEngineAudio(mediaEngine)
	if err != nil {
		logs.Error(err)
	}

	err = p.initMediaEngineVideo(98, mediaEngine)
	if err != nil {
		logs.Error(err)
	}

	return mediaEngine
}

func (p *Peer) initSettingEngine() *webrtc.SettingEngine {
	settingEngine := &webrtc.SettingEngine{}
	// settingEngine.SetEphemeralUDPPortRange(20000, 60000)
	// settingEngine.SetICETimeouts(10*time.Second, 20*time.Second, 1*time.Second)
	return settingEngine
}

func (p *Peer) setState(s string) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.state = s
}

func (p *Peer) getState() string {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.state
}
