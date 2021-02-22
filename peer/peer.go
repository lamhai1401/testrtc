package peer

import (
	"fmt"
	"sync"

	"github.com/lamhai1401/testrtc/utils"
	"github.com/mitchellh/mapstructure"
	"github.com/pion/rtp"
	"github.com/pion/webrtc/v3"
)

// var (
// 	defaultAudioCodecs = uint8(webrtc.DefaultPayloadTypeOpus)
// 	defaultVideoCodecs = uint8(webrtc.DefaultPayloadTypeVP8)
// )

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
	sessionID        string
	signalID         string
	bitrate          *int
	iceCache         *utils.AdvanceMap
	conn             *webrtc.PeerConnection
	localVideoTrack  *webrtc.TrackLocalStaticRTP
	localAudioTrack  *webrtc.TrackLocalStaticRTP
	remoteAudioTrack *webrtc.TrackRemote
	remoteVideoTrack *webrtc.TrackRemote
	state            string
	isConnected      bool
	isClosed         bool
	mutex            sync.RWMutex
}

// NewPeer linter
func NewPeer(
	bitrate *int,
	sessionID string,
	signalID string,
) *Peer {
	p := &Peer{
		bitrate:     bitrate,
		iceCache:    utils.NewAdvanceMap(),
		sessionID:   sessionID,
		signalID:    signalID,
		isClosed:    false,
		isConnected: false,
	}

	return p
}

// NewConnection linte
func (p *Peer) NewConnection(config *webrtc.Configuration) (*webrtc.PeerConnection, error) {
	api := p.addAPI()
	conn, err := api.NewPeerConnection(*config)
	if err != nil {
		return nil, err
	}
	p.setConn(conn)

	if err := p.createAudioTrack(p.getSignalID()); err != nil {
		return nil, err
	}

	if err := p.createVideoTrack(p.getSignalID()); err != nil {
		return nil, err
	}

	return conn, nil
}

// Close linter
func (p *Peer) Close() {
	if !p.checkClose() {
		p.setClose(true)
		p.closeConn()
	}
}

// AddVideoRTP write rtp to local video track
func (p *Peer) AddVideoRTP(packet *rtp.Packet) error {
	track := p.getLocalVideoTrack()
	if track == nil {
		return fmt.Errorf("ErrNilVideoTrack")
	}
	return p.writeRTP(packet, track)
}

// AddAudioRTP write rtp to local audio track
func (p *Peer) AddAudioRTP(packet *rtp.Packet) error {
	track := p.getLocalAudioTrack()
	if track == nil {
		return fmt.Errorf("ErrNilAudioTrack")
	}
	return p.writeRTP(packet, track)
}

// AddICECandidate to add candidate
func (p *Peer) AddICECandidate(icecandidate interface{}) error {
	var candidateInit webrtc.ICECandidateInit
	err := mapstructure.Decode(icecandidate, &candidateInit)
	if err != nil {
		return err
	}

	conn := p.getConn()
	if conn == nil {
		p.addIceCache(&candidateInit)
		return fmt.Errorf("ErrNilPeerconnection")
	}

	if conn.RemoteDescription() == nil {
		p.addIceCache(&candidateInit)
	}

	return conn.AddICECandidate(candidateInit)
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
		return fmt.Errorf("ErrNilPeerconnection")
	}

	var data utils.SDPTemp
	err := mapstructure.Decode(values, &data)
	if err != nil {
		return err
	}

	sdp := &webrtc.SessionDescription{
		Type: NewSDPType(data.Type),
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

// SetConnected linter
func (p *Peer) SetConnected() {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.isConnected = true
}

// CheckConnected linter
func (p *Peer) CheckConnected() bool {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.isConnected
}

// GetLocalDescription get current peer local description
func (p *Peer) GetLocalDescription() (*webrtc.SessionDescription, error) {
	conn := p.getConn()
	if conn == nil {
		return nil, fmt.Errorf("rtc connection is nil")
	}
	return conn.LocalDescription(), nil
}

// GetConn linter
func (p *Peer) GetConn() *webrtc.PeerConnection {
	return p.getConn()
}

// GetSignalID linter
func (p *Peer) GetSignalID() string {
	return p.getSignalID()
}

// GetSessionID linter
func (p *Peer) GetSessionID() string {
	return p.getSessionID()
}
