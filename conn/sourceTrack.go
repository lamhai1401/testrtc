package conn

import (
	"fmt"
	"strings"
	"sync"

	"github.com/pion/webrtc/v3"
)

type SourceTrack struct {
	videoStrackIDs []string
	audioStrackIDs []string
	videoTracks    map[string]*webrtc.TrackLocalStaticRTP
	audioTracks    map[string]*webrtc.TrackLocalStaticRTP
	payloadType    int    // video codecs code VP8 - 98 or VP9 - 99
	codec          string // only video video/VP9 or video/VP8. default audio is opus
	mutex          sync.RWMutex
}

// NewTracks linter
func NewSouceTracks(
	audioStrackIDs []string,
	videoStrackIDs []string,
	payloadType int, // video codecs code VP8 - 98 or VP9 - 99
	codec string,
) LocalTrack { // only video video/VP9 or video/VP8. default audio is opus) *Tracks {
	t := &SourceTrack{
		videoStrackIDs: videoStrackIDs,
		videoTracks:    make(map[string]*webrtc.TrackLocalStaticRTP),
		audioTracks:    make(map[string]*webrtc.TrackLocalStaticRTP),
		payloadType:    payloadType,
		codec:          codec,
	}

	return t
}

func (t *SourceTrack) InitLocalTrack(
	p *Peer,
) error {
	conn := p.getConn()
	if conn == nil {
		return fmt.Errorf("Peer connections is nil")
	}

	for _, v := range t.videoStrackIDs {
		if err := t.initLocalVideoTrack(v, conn); err != nil {
			return err
		}
	}

	for _, v := range t.audioStrackIDs {
		if err := t.initLocalAudioTrack(v, conn); err != nil {
			return err
		}
	}

	// set process rtcp
	t._processRTCP(conn)
	return nil
}

func (t *SourceTrack) GetVideoTrack(id string) (*webrtc.TrackLocalStaticRTP, error) {
	return t.getVideoTrack(id), nil
}

func (t *SourceTrack) GetAudioTrack(id string) (*webrtc.TrackLocalStaticRTP, error) {
	return t.getAudioTrack(id), nil
}

// CreateAudioTrack linter
func (t *SourceTrack) _createAudioTrack(id string) (*webrtc.TrackLocalStaticRTP, error) {
	return t._createTrack(id, MimeTypeOpus)
}

// CreateAudioTrack linter
func (t *SourceTrack) _createVideoTrack(id string) (*webrtc.TrackLocalStaticRTP, error) {
	return t._createTrack(id, fmt.Sprintf("video/%s", strings.ToUpper(t.getCodecs())))
}

// CreateTrack linter
func (t *SourceTrack) _createTrack(id, mimeType string) (*webrtc.TrackLocalStaticRTP, error) {
	return webrtc.NewTrackLocalStaticRTP(
		webrtc.RTPCodecCapability{MimeType: mimeType},
		id,
		id,
	)
}

func (t *SourceTrack) _processRTCP(peerConnection *webrtc.PeerConnection) {
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

func (t *SourceTrack) getCodecs() string {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.codec
}

func (t *SourceTrack) getPayloadType() int {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.payloadType
}

func (t *SourceTrack) initLocalVideoTrack(id string, conn *webrtc.PeerConnection) error {
	videoTrack, err := t._createVideoTrack(id)
	if err != nil {
		return err
	}

	// Add this newly created track to the PeerConnection
	_, err = conn.AddTrack(videoTrack)
	if err != nil {
		return err
	}

	// set track
	t.setVideoTracks(id, videoTrack)
	return nil
}

func (t *SourceTrack) initLocalAudioTrack(id string, conn *webrtc.PeerConnection) error {
	audioTrack, err := t._createAudioTrack(id)
	if err != nil {
		return err
	}

	// Add this newly created track to the PeerConnection
	_, err = conn.AddTrack(audioTrack)
	if err != nil {
		return err
	}

	// set track
	t.setAudioTracks(id, audioTrack)
	return nil
}

func (t *SourceTrack) setVideoTracks(id string, track *webrtc.TrackLocalStaticRTP) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.videoTracks[id] = track
}

func (t *SourceTrack) setAudioTracks(id string, track *webrtc.TrackLocalStaticRTP) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.audioTracks[id] = track
}

func (t *SourceTrack) getVideoTrack(index string) *webrtc.TrackLocalStaticRTP {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.videoTracks[index]
}

func (t *SourceTrack) getAudioTrack(index string) *webrtc.TrackLocalStaticRTP {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.audioTracks[index]
}
