package peer

import (
	"fmt"
	"sync"

	"github.com/lamhai1401/gologs/logs"
	"github.com/lamhai1401/testrtc/utils"
	"github.com/mitchellh/mapstructure"
	"github.com/pion/rtp"
	"github.com/pion/webrtc/v3"
)

// NewSDPType linter
func NewSDPType(raw string) webrtc.SDPType {
	switch raw {
	case "offer":
		return webrtc.SDPTypeOffer
	case "answer":
		return webrtc.SDPTypeAnswer
	default:
		return webrtc.SDPType(webrtc.Unknown)
	}
}

// Peer linter
type Peer struct {
	streamID         string
	payloadType      int    // video codecs code VP8 - 98 or VP9 - 99
	codec            string // only video video/VP9 or video/VP8. default audio is opus
	sessionID        string
	signalID         string
	state            string
	conn             *webrtc.PeerConnection
	localVideoTrack  *webrtc.TrackLocalStaticRTP
	localAudioTrack  *webrtc.TrackLocalStaticRTP
	remoteAudioTrack *webrtc.TrackRemote
	remoteVideoTrack *webrtc.TrackRemote
	iceCache         *utils.AdvanceMap
	isConnected      bool
	isClosed         bool
	bitrate          *int
	mutex            sync.RWMutex
}

// NewPeer linter
func NewPeer(
	bitrate *int,
	streamID string,
	sessionID string,
	signalID string,
	codec string,
	payloadType int,
) Connection {
	p := &Peer{
		streamID:    streamID,
		bitrate:     bitrate,
		iceCache:    utils.NewAdvanceMap(),
		sessionID:   sessionID,
		signalID:    signalID,
		isClosed:    false,
		isConnected: false,
		codec:       codec,
		payloadType: payloadType,
	}

	if bitrate == nil {
		br := 200
		p.bitrate = &br
	}

	return p
}

func newPeerConnection(
	bitrate *int,
	streamID string,
	role string,
	sessionID string,
	codec string,
	payloadType int,
) *Peer {
	p := &Peer{
		sessionID: sessionID,
		streamID:  streamID,
		// role:        role,
		bitrate:     bitrate,
		isConnected: false,
		isClosed:    false,
		codec:       codec,
		payloadType: payloadType,
		iceCache:    utils.NewAdvanceMap(),
	}

	if bitrate == nil {
		br := 200
		p.bitrate = &br
	}
	logs.Warn(fmt.Sprintf("%s_%s peer connection was created with video codec, code, bitrate (%s/%d/%d)", streamID, sessionID, codec, payloadType, *p.bitrate))
	return p
}

// InitPeer linter
func (p *Peer) InitPeer(config *webrtc.Configuration) (*webrtc.PeerConnection, error) {
	api := p.initAPI()
	if api == nil {
		return nil, fmt.Errorf("webrtc api is nil")
	}

	conn, err := api.NewPeerConnection(*config)
	if err != nil {
		return nil, err
	}
	p.setConn(conn)

	err = p.createAudioTrack(p.getStreamID())
	if err != nil {
		return nil, err
	}

	err = p.createVideoTrack(p.getStreamID())
	if err != nil {
		return nil, err
	}

	// warning rtcp report
	go p.processRTCP(conn)

	return conn, err
}

// Close close peer connection
func (p *Peer) Close() {
	if !p.checkClose() {
		p.setClose(true)
		p.setBitrate(nil)
		if err := p.closeConn(); err != nil {
			logs.Error(err.Error())
		}
		logs.Warn(fmt.Sprintf("%s_%s webrtc peer connection was closed", p.getStreamID(), p.getSessionID()))
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
func (p *Peer) AddVideoRTP(packet *rtp.Packet) error {
	track := p.getLocalVideoTrack()
	if track == nil {
		return fmt.Errorf(ErrNilVideoTrack)
	}
	return p.writeRTP(packet, track)
}

// AddAudioRTP write rtp to local audio track
func (p *Peer) AddAudioRTP(packet *rtp.Packet) error {
	track := p.getLocalAudioTrack()
	if track == nil {
		return fmt.Errorf(ErrNilAudioTrack)
	}
	return p.writeRTP(packet, track)
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
	logs.Info(fmt.Sprintf("Add ice candidate from %s successfully", p.GetSessionID()))
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
		break
	case "answer":
		if err := p.addAnswer(sdp); err != nil {
			return err
		}
		break
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

// GetLocalDescription get current peer local description
func (p *Peer) getLocalDescription() (*webrtc.SessionDescription, error) {
	conn := p.getConn()
	if conn == nil {
		return nil, fmt.Errorf("rtc connection is nil")
	}
	return conn.LocalDescription(), nil
}

// HandleVideoTrack handle all video track
func (p *Peer) HandleVideoTrack(remoteTrack *webrtc.TrackRemote) {
	go p.modifyBitrate(remoteTrack)
	go p.pictureLossIndication(remoteTrack)
	go p.rapidResynchronizationRequest(remoteTrack)
}

func (p *Peer) processRTCP(peerConnection *webrtc.PeerConnection) {
	// Read incoming RTCP packets
	// Before these packets are retuned they are processed by interceptors. For things
	// like NACK this needs to be called.
	processRTCP := func(rtpSender *webrtc.RTPSender) {
		rtcpBuf := make([]byte, 1500)
		for {
			if _, _, rtcpErr := rtpSender.Read(rtcpBuf); rtcpErr != nil {
				return
			}
		}
	}
	for _, rtpSender := range peerConnection.GetSenders() {
		go processRTCP(rtpSender)
	}
}
