package conn

import (
	"fmt"
	"sync"

	log "github.com/lamhai1401/gologs/logs"
	"github.com/lamhai1401/testrtc/utils"
	"github.com/mitchellh/mapstructure"
	"github.com/pion/rtp"
	"github.com/pion/webrtc/v3"
)

// Peer webrtc peer connection
type Peer struct {
	payloadType int               // video codecs code VP8 - 98 or VP9 - 99
	codec       string            // only video video/VP9 or video/VP8. default audio is opus
	sessionID   string            // save session old or new
	streamID    string            // save to stream id
	offerID     string            // save offerID for unique peer connection
	cookieID    string            // internal cookieID to check remove
	role        string            // source or des
	iceCache    *utils.AdvanceMap // save all ice before set remote description
	isConnected bool              // check to this is connection init or ice restart
	isClosed    bool              // check peer close or not
	hasVideo    bool              // has video track or not
	hasAudio    bool              // has audio track or not
	bitrate     *int              // bitrate
	tracks      *Tracks           // for handle multi video and audio
	conn        *webrtc.PeerConnection
	mutex       sync.RWMutex
}

// NewPeerConnection linter
func NewPeerConnection(
	bitrate *int,
	streamID string,
	role string,
	sessionID string,
	codec string,
	payloadType int,
	offerID string,
) Connection {
	return newPeerConnection(bitrate, streamID, role, sessionID, codec, payloadType, offerID)
}

func newPeerConnection(
	bitrate *int,
	streamID string,
	role string,
	sessionID string,
	codec string,
	payloadType int,
	offerID string,
) *Peer {
	p := &Peer{
		sessionID:   sessionID,
		streamID:    streamID,
		role:        role,
		bitrate:     bitrate,
		isConnected: false,
		isClosed:    false,
		codec:       codec,
		offerID:     offerID,
		payloadType: payloadType,
		iceCache:    utils.NewAdvanceMap(),
		cookieID:    utils.GenerateID(),
		hasVideo:    false,
		hasAudio:    false,
	}

	if bitrate == nil {
		br := 200
		p.bitrate = &br
	}
	log.Warn(fmt.Sprintf("%s_%s_%s_%s peer connection was created with video codec, code, bitrate (%s/%d/%d)", streamID, sessionID, offerID, p.getCookieID(), codec, payloadType, *p.bitrate))
	return p
}

// InitPeer linter
func (p *Peer) InitPeer(
	configs *webrtc.Configuration,
) (*webrtc.PeerConnection, error) {
	api := p.initAPI()
	if api == nil {
		return nil, fmt.Errorf("webrtc api is nil")
	}

	conn, err := api.NewPeerConnection(*configs)
	if err != nil {
		return nil, err
	}
	p.setConn(conn)

	tracks := NewTracks(1, p.payloadType, p.codec)
	if err := tracks.initLocalTrack(p); err != nil {
		return nil, err
	}
	p.tracks = tracks

	return conn, nil
}

// Close close peer connection
func (p *Peer) Close() {
	if !p.checkClose() {
		p.setClose(true)
		p.setBitrate(nil)
		if err := p.closeConn(); err != nil {
			log.Error(err.Error())
		}
		log.Warn(fmt.Sprintf("%s_%s webrtc peer connection was closed", p.getStreamID(), p.getSessionID()))
	}
}

// SetIsConnected linter
func (p *Peer) SetIsConnected(states bool) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.isConnected = states
}

// IsConnected check this is init connection or retrieve
func (p *Peer) IsConnected() bool {
	return p.isConnected
}

// AddVideoRTP write rtp to local video track
func (p *Peer) AddVideoRTP(sessionID string, packet *rtp.Packet) error {
	tracks := p.getTracks()
	if tracks == nil {
		return fmt.Errorf("Tracks of peer (%s) is nil", sessionID)
	}

	index, err := tracks.getIndexOf(sessionID)
	if err != nil {
		return err
	}

	track := tracks.getVideoTrack(index)
	if track == nil {
		return fmt.Errorf(ErrNilVideoTrack)
	}
	if err = p.writeRTP(packet, track); err != nil {
		return err
	}

	log.Stack(fmt.Sprintf("Write %s video rtp to %s at index %s", p.getStreamID(), p.GetSessionID(), index))
	return nil
}

// AddVideoRTP write rtp to local video track
func (p *Peer) AddAudioRTP(sessionID string, packet *rtp.Packet) error {
	tracks := p.getTracks()
	if tracks == nil {
		return fmt.Errorf("Tracks of peer (%s) is nil", sessionID)
	}

	index, err := tracks.getIndexOf(sessionID)
	if err != nil {
		return err
	}

	track := tracks.getAudioTrack(index)
	if track == nil {
		return fmt.Errorf(ErrNilAudioTrack)
	}

	if err = p.writeRTP(packet, track); err != nil {
		return err
	}

	log.Stack(fmt.Sprintf("Write %s audio rtp to %s at index %s", p.getStreamID(), p.GetSessionID(), index))
	return nil
}

// HandleVideoTrack handle all video track
func (p *Peer) HandleVideoTrack(remoteTrack *webrtc.TrackRemote) {
	go p.modifyBitrate(remoteTrack)
	go p.pictureLossIndication(remoteTrack)
	// go p.rapidResynchronizationRequest(remoteTrack)
}

// AddICECandidate to add candidate
func (p *Peer) AddICECandidate(icecandidate interface{}) error {
	// var candidateInit webrtc.ICECandidateInit
	candidateInit, ok := icecandidate.(*webrtc.ICECandidateInit)
	if !ok {
		err := mapstructure.Decode(icecandidate, &candidateInit)
		if err != nil {
			return err
		}
	}

	conn := p.getConn()
	if conn == nil {
		p.addIceCache(candidateInit)
		return fmt.Errorf(ErrNilPeerconnection)
	}

	if conn.RemoteDescription() == nil {
		p.addIceCache(candidateInit)
	}

	err := conn.AddICECandidate(*candidateInit)
	if err != nil {
		return err
	}
	log.Info(fmt.Sprintf("Add ice candidate for %s_%s_%s_%s successfully", p.getStreamID(), p.GetSessionID(), p.getOffeerID(), p.GetCookieID()))
	return nil
}

// CreateOffer add offer
func (p *Peer) CreateOffer(iceRestart bool) error {
	conn := p.getConn()
	if conn == nil {
		return fmt.Errorf("webrtc connection is nil")
	}

	// opt := &webrtc.OfferOptions{}

	// if iceRestart {
	// 	opt.ICERestart = iceRestart
	// }

	// set local desc
	offer, err := conn.CreateOffer(nil)
	if err != nil {
		return err
	}

	err = conn.SetLocalDescription(offer)
	if err != nil {
		return err
	}
	return nil
}

// CreateAnswer add answer
func (p *Peer) CreateAnswer() error {
	conn := p.getConn()
	if conn == nil {
		return fmt.Errorf("webrtc connection is nil")
	}
	// set local desc
	answer, err := conn.CreateAnswer(nil)
	if err != nil {
		return err
	}

	err = conn.SetLocalDescription(answer)
	if err != nil {
		return err
	}
	return nil
}

// AddSDP add sdp
func (p *Peer) AddSDP(values interface{}) error {
	conns := p.getConn()
	if conns == nil {
		return fmt.Errorf(ErrNilPeerconnection)
	}

	var data utils.SDPTemp
	err := mapstructure.Decode(values, &data)
	if err != nil {
		return err
	}

	sdp := &webrtc.SessionDescription{
		Type: utils.NewSDPType(data.Type),
		SDP:  data.SDP,
	}

	switch data.Type {
	case "offer":
		if err := p.addOffer(sdp); err != nil {
			return err
		}
	case "answer":
		if err := p.addAnswer(sdp); err != nil {
			return err
		}
	default:
		return fmt.Errorf("Invalid sdp type: %s", data.Type)
	}
	return nil
}

// GetLocalDescription get current peer local description to send to client
func (p *Peer) GetLocalDescription() (*webrtc.SessionDescription, error) {
	return p.getLocalDescription()
}

// GetSessionID linter
func (p *Peer) GetSessionID() string {
	return p.getSessionID()
}

// SetSessionID linter
func (p *Peer) SetSessionID(s string) {
	p.setSessionID(s)
}

// GetRole linter
func (p *Peer) GetRole() string {
	return p.getRole()
}

// GetOfferID linter
func (p *Peer) GetOfferID() string {
	return p.getOffeerID()
}

// GetCookieID linter
func (p *Peer) GetCookieID() string {
	return p.getCookieID()
}

// HasVideoTrack check this peer has video or not
func (p *Peer) HasVideoTrack() bool {
	return p.getHasVideoTrack()
}

// HasAudioTrack check this peer has audio or not
func (p *Peer) HasAudioTrack() bool {
	return p.getHasAudioTrack()
}
