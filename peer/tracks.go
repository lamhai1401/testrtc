package peer

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
func NewTracks(
	length int,
	payloadType int, // video codecs code VP8 - 98 or VP9 - 99
	codec string,
) *Tracks { // only video video/VP9 or video/VP8. default audio is opus) *Tracks {
	t := &Tracks{
		length:      length,
		videoTracks: make(map[string]*webrtc.TrackLocalStaticRTP),
		audioTracks: make(map[string]*webrtc.TrackLocalStaticRTP),
		checkList:   utils.NewAdvanceMap(),
		listUser:    make(map[string]string),
		payloadType: payloadType,
		codec:       codec,
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
		if err := t.initLocalVideoTrack(id, conn); err != nil {
			return err
		}

		if err := t.initLocalAudioTrack(id, conn); err != nil {
			return err
		}

		// save free track to list
		t.setCheckList(id, defaultTrackState)
	}

	// set process rtcp
	t._processRTCP(conn)
	return nil
}

func (t *Tracks) initLocalVideoTrack(id string, conn *webrtc.PeerConnection) error {
	videoTrack, err := t._createVideoTrack(id)
	if err != nil {
		return err
	}

	// // Add this newly created track to the PeerConnection
	// sender, err := conn.AddTrack(videoTrack)
	// if err != nil {
	// 	return err
	// }

	// // set track
	t.setVideoTracks(id, videoTrack)

	_, err = initTransceiver(conn, "video", nil, videoTrack)
	return err
}

func (t *Tracks) initLocalAudioTrack(id string, conn *webrtc.PeerConnection) error {
	audioTrack, err := t._createAudioTrack(id)
	if err != nil {
		return err
	}

	// // Add this newly created track to the PeerConnection
	// sender, err := conn.AddTrack(audioTrack)
	// if err != nil {
	// 	return err
	// }

	// set track
	t.setAudioTracks(id, audioTrack)
	_, err = initTransceiver(conn, "audio", nil, audioTrack)
	return err
}

// releaseTrack free a index of a track
func (t *Tracks) releaseTrack(sessionID string) {
	if index := t.getUserInList(sessionID); index != "" {
		t.releaseCheckList(index)
		t.deleteUserInList(sessionID)
		logs.Warn(fmt.Sprintf("[registerTrack] %s id unRegister track at index %s", sessionID, index))
	}
}

// getIndexOf find index with input string and return, if not return ""
func (t *Tracks) getIndexOf(sessionID string) (string, error) {
	var index string
	var err error
	index = t.getUserInList(sessionID)
	if index == "" {
		return t._findFreeTrack()
	}
	return index, err
}

func (t *Tracks) _registerTrack(sessionID string) (string, error) {
	index, err := t._findFreeTrack()

	if err != nil {
		return "", err
	}

	// save check list
	t.setCheckList(index, sessionID)
	t.setUserInList(sessionID, index)

	logs.Warn(fmt.Sprintf("[registerTrack] %s id register track at index %s", sessionID, index))
	return index, err
}

func (t *Tracks) _findFreeTrack() (string, error) {
	var result string
	var err error

	for i := 1; i <= t.getLength(); i++ {
		state := t.getInCheckList(fmt.Sprintf("%d", i))
		if state == defaultTrackState {
			result = fmt.Sprintf("%d", i)
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

// CreateAudioTrack linter
func (t *Tracks) _createAudioTrack(id string) (*webrtc.TrackLocalStaticRTP, error) {
	return t._createTrack(id, MimeTypeOpus)
}

// CreateAudioTrack linter
func (t *Tracks) _createVideoTrack(id string) (*webrtc.TrackLocalStaticRTP, error) {
	return t._createTrack(id, fmt.Sprintf("video/%s", strings.ToUpper(t.getCodecs())))
}

// CreateTrack linter
func (t *Tracks) _createTrack(id, mimeType string) (*webrtc.TrackLocalStaticRTP, error) {
	return webrtc.NewTrackLocalStaticRTP(
		webrtc.RTPCodecCapability{MimeType: mimeType},
		id,
		id,
	)
}

func (t *Tracks) _processRTCP(peerConnection *webrtc.PeerConnection) {
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
