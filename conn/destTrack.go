package conn

import (
	"fmt"
	"strings"
	"sync"

	"github.com/davecgh/go-spew/spew"
	"github.com/lamhai1401/gologs/logs"
	"github.com/lamhai1401/testrtc/utils"
	"github.com/pion/webrtc/v3"
)

// Tracks linter
type DestTrack struct {
	length      int    // length of list track
	payloadType int    // video codecs code VP8 - 98 or VP9 - 99
	codec       string // only video video/VP9 or video/VP8. default audio is opus
	videoTracks map[string]*webrtc.TrackLocalStaticRTP
	audioTracks map[string]*webrtc.TrackLocalStaticRTP
	checkList   *utils.AdvanceMap // id - streamID/0  to check local map was pushing data or not
	listUser    map[string]string // save id to track index
	mutex       sync.RWMutex
}

// NewTracks linter
func NewDestTracks(
	length int,
	payloadType int, // video codecs code VP8 - 98 or VP9 - 99
	codec string,
) LocalTrack { // only video video/VP9 or video/VP8. default audio is opus) *Tracks {
	t := &DestTrack{
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

func (t *DestTrack) InitLocalTrack(p *Peer) error {
	conn := p.getConn()
	if conn == nil {
		return fmt.Errorf("Peer connections is nil")
	}

	for i := 1; i <= t.getLength(); i++ {
		id := fmt.Sprintf("%d", i)
		if err := t._initTrack(conn, id); err != nil {
			return err
		}
	}

	// set process rtcp
	t._processRTCP(conn)
	return nil
}

// TODO : Check input seesion id when call wrtie rtp to make sure the source session not own peer session
func (t *DestTrack) GetVideoTrack(sessionID string) (*webrtc.TrackLocalStaticRTP, error) {
	index, err := t.getIndexOf(sessionID)
	if err != nil {
		return nil, err
	}
	return t.getVideoTrack(index), nil
}

func (t *DestTrack) GetAudioTrack(sessionID string) (*webrtc.TrackLocalStaticRTP, error) {
	index, err := t.getIndexOf(sessionID)
	if err != nil {
		return nil, err
	}
	return t.getAudioTrack(index), nil
}

func (t *DestTrack) getLength() int {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.length
}

func (t *DestTrack) setVideoTracks(id string, track *webrtc.TrackLocalStaticRTP) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.videoTracks[id] = track
}

func (t *DestTrack) setAudioTracks(id string, track *webrtc.TrackLocalStaticRTP) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.audioTracks[id] = track
}

func (t *DestTrack) getVideoTrack(index string) *webrtc.TrackLocalStaticRTP {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.videoTracks[index]
}

func (t *DestTrack) getAudioTrack(index string) *webrtc.TrackLocalStaticRTP {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.audioTracks[index]
}

func (t *DestTrack) getVideoTracks() map[string]*webrtc.TrackLocalStaticRTP {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.videoTracks
}

func (t *DestTrack) getAudioTracks() map[string]*webrtc.TrackLocalStaticRTP {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.audioTracks
}

func (t *DestTrack) getListUser() map[string]string {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.listUser
}

func (t *DestTrack) getUserInList(sessionID string) string {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.listUser[sessionID]
}

func (t *DestTrack) setUserInList(sessionID, index string) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.listUser[sessionID] = index
}

func (t *DestTrack) deleteUserInList(sessionID string) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	delete(t.listUser, sessionID)
}

func (t *DestTrack) getCheckList() *utils.AdvanceMap {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.checkList
}

func (t *DestTrack) setCheckList(index, streamID string) {
	if lst := t.getCheckList(); lst != nil {
		lst.Set(index, streamID)
	}
}

func (t *DestTrack) releaseCheckList(index string) {
	if lst := t.getCheckList(); lst != nil {
		lst.Set(index, defaultTrackState)
	}
}

// check input index has stream id or not, return defaultTrackState if not
func (t *DestTrack) getInCheckList(index string) string {
	if cl := t.getCheckList(); cl != nil {
		s, ok := cl.Get(index)
		if ok {
			state, ok := s.(string)
			if ok {
				return state
			}
		}
	}
	return ""
}

func (t *DestTrack) getCodecs() string {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.codec
}

func (t *DestTrack) setCodecs(c string) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.codec = c
}

func (t *DestTrack) getPayloadType() int {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.payloadType
}

func (t *DestTrack) setPayloadType(c int) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.payloadType = c
}

func (t *DestTrack) _initTrack(conn *webrtc.PeerConnection, id string) error {
	videoTrack, err := t._createVideoTrack(id)
	if err != nil {
		return err
	}

	audioTrack, err := t._createAudioTrack(id)
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
	return nil
}

// CreateAudioTrack linter
func (t *DestTrack) _createAudioTrack(id string) (*webrtc.TrackLocalStaticRTP, error) {
	return t._createTrack(id, MimeTypeOpus)
}

// CreateAudioTrack linter
func (t *DestTrack) _createVideoTrack(id string) (*webrtc.TrackLocalStaticRTP, error) {
	return t._createTrack(id, fmt.Sprintf("video/%s", strings.ToUpper(t.getCodecs())))
}

// CreateTrack linter
func (t *DestTrack) _createTrack(id, mimeType string) (*webrtc.TrackLocalStaticRTP, error) {
	return webrtc.NewTrackLocalStaticRTP(
		webrtc.RTPCodecCapability{MimeType: mimeType},
		id,
		id,
	)
}

func (t *DestTrack) _processRTCP(peerConnection *webrtc.PeerConnection) {
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

func (t *DestTrack) _registerTrack(sessionID string) (string, error) {
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

func (t *DestTrack) _findFreeTrack() (string, error) {
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

// releaseTrack free a index of a track, but this method is not you at that time
func (t *DestTrack) releaseTrack(sessionID string) {
	if index := t.getUserInList(sessionID); index != "" {
		t.releaseCheckList(index)
		t.deleteUserInList(sessionID)
		logs.Warn(fmt.Sprintf("[registerTrack] %s id unRegister track at index %s", sessionID, index))
	}
}

// getIndexOf find index with input string and return, if not return ""
func (t *DestTrack) getIndexOf(sessionID string) (string, error) {
	var index string
	var err error
	index = t.getUserInList(sessionID)
	if index == "" {
		return t._findFreeTrack()
	}
	return index, err
}
