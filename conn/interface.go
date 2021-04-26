package conn

import (
	"github.com/pion/rtp"
	"github.com/pion/webrtc/v3"
)

// Connection interface
type Connection interface {
	HasVideoTrack() bool
	HasAudioTrack() bool
	GetCookieID() string
	GetOfferID() string
	GetRole() string
	GetSessionID() string  // for handle ice restart
	SetSessionID(s string) // for handle ice restart
	HandleVideoTrack(remoteTrack *webrtc.TrackRemote)
	InitPeer(configs *webrtc.Configuration) (*webrtc.PeerConnection, error)
	IsConnected() bool
	CreateOffer(iceRestart bool) error
	CreateAnswer() error
	AddSDP(values interface{}) error
	AddICECandidate(icecandidate interface{}) error
	AddVideoRTP(packet *rtp.Packet) error
	AddAudioRTP(packet *rtp.Packet) error
	SetIsConnected(states bool)
	SetRemoteVideoTrack(remoteTrack *webrtc.TrackRemote)
	SetRemoteAudioTrack(remoteTrack *webrtc.TrackRemote)
	GetRemoteTrack() (*webrtc.TrackRemote, *webrtc.TrackRemote)
	GetLocalDescription() (*webrtc.SessionDescription, error)
	Close()
}
