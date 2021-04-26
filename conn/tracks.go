package conn

import (
	"fmt"
	"strings"
	"sync"

	"github.com/beowulflab/rtcbase-v2/utils"
	"github.com/davecgh/go-spew/spew"
	"github.com/lamhai1401/gologs/logs"
	"github.com/pion/webrtc/v3"
)

// Tracks linter
type Tracks struct {
	length      int
	payloadType int    // video codecs code VP8 - 98 or VP9 - 99
	codec       string // only video video/VP9 or video/VP8. default audio is opus
	videoTracks map[string]*webrtc.TrackLocalStaticRTP
	audioTracks map[string]*webrtc.TrackLocalStaticRTP
	checkList   *utils.AdvanceMap // id - streamID/0  to check local map was pushing data or not
	listUser    map[string]string // save id to track index
	mutex       sync.RWMutex
}

// NewTracks linter
func NewTracks(length int) *Tracks {
	t := &Tracks{
		length:      length,
		videoTracks: make(map[string]*webrtc.TrackLocalStaticRTP),
		audioTracks: make(map[string]*webrtc.TrackLocalStaticRTP),
		checkList:   utils.NewAdvanceMap(),
		listUser:    make(map[string]string),
	}

	return t
}

func (t *Tracks) initLocalTrack(p *Peer) error {
	conn := p.getConn()
	if conn == nil {
		return fmt.Errorf("Peer connections is nil")
	}

	for i := 1; i <= t.getLength(); i++ {
		id := fmt.Sprintf("%d", i)
		videoTrack, err := t.createVideoTrack(id)
		if err != nil {
			return err
		}

		audioTrack, err := t.createAudioTrack(id)
		if err != nil {
			return err
		}
		// Add this newly created track to the PeerConnection
		_, err = conn.AddTrack(videoTrack)
		if err != nil {
			return err
		}

		// Add this newly created track to the PeerConnection
		_, err = conn.AddTrack(audioTrack)
		if err != nil {
			return err
		}

		// set track
		t.setVideoTracks(id, videoTrack)
		t.setAudioTracks(id, audioTrack)

		// save free track to list
		t.setCheckList(id, defaultTrackState)
	}

	// set process rtcp
	t.processRTCP(conn)
	return nil
}

func (t *Tracks) findFreeTrack() (string, error) {
	var result string
	var err error

	for i := 1; i <= t.getLength(); i++ {
		state := t.getInCheckList(fmt.Sprintf("%d", i))
		if state == defaultTrackState {
			result = state
			break
		}
	}

	if result == "" {
		err = fmt.Errorf("all of local track is full")
		logs.Warn("Local Track map: ")
		spew.Dump(t.getCheckList())
	}

	return result, err
}

func (t *Tracks) registerTrack(sessionID string) (string, error) {
	index, err := t.findFreeTrack()

	if err != nil {
		return "", err
	}

	// save check list
	t.setCheckList(index, sessionID)
	t.setUserInList(sessionID, index)

	logs.Warn(fmt.Sprintf("[registerTrack] %s id register track at index %s", sessionID, index))
	return index, err
}

func (t *Tracks) unRegisterTrack(sessionID string) {
	if index := t.getUserInList(sessionID); index != "" {
		t.releaseCheckList(index)
		t.deleteUserInList(sessionID)
		logs.Warn(fmt.Sprintf("[registerTrack] %s id unRegister track at index %s", sessionID, index))
	}
}

func (t *Tracks) processRTCP(peerConnection *webrtc.PeerConnection) {
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

// CreateAudioTrack linter
func (t *Tracks) createAudioTrack(id string) (*webrtc.TrackLocalStaticRTP, error) {
	return t.createTrack(id, MimeTypeOpus)
}

// CreateAudioTrack linter
func (t *Tracks) createVideoTrack(id string) (*webrtc.TrackLocalStaticRTP, error) {
	return t.createTrack(id, fmt.Sprintf("video/%s", strings.ToUpper(t.getCodecs())))
}

// CreateTrack linter
func (t *Tracks) createTrack(id, mimeType string) (*webrtc.TrackLocalStaticRTP, error) {
	return webrtc.NewTrackLocalStaticRTP(
		webrtc.RTPCodecCapability{MimeType: fmt.Sprintf(mimeType, strings.ToUpper(t.getCodecs()))},
		id,
		id,
	)
}
