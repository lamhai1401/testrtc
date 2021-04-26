package conn

import (
	"github.com/beowulflab/rtcbase-v2/utils"
	"github.com/pion/webrtc/v3"
)

func (t *Tracks) getLength() int {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.length
}

func (t *Tracks) setVideoTracks(id string, track *webrtc.TrackLocalStaticRTP) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.videoTracks[id] = track
}

func (t *Tracks) setAudioTracks(id string, track *webrtc.TrackLocalStaticRTP) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.audioTracks[id] = track
}

func (t *Tracks) getVideoTrack(index string) *webrtc.TrackLocalStaticRTP {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.videoTracks[index]
}

func (t *Tracks) getAudioTrack(index string) *webrtc.TrackLocalStaticRTP {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.audioTracks[index]
}

func (t *Tracks) getVideoTracks() map[string]*webrtc.TrackLocalStaticRTP {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.videoTracks
}

func (t *Tracks) getAudioTracks() map[string]*webrtc.TrackLocalStaticRTP {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.audioTracks
}

func (t *Tracks) getListUser() map[string]string {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.listUser
}

func (t *Tracks) getUserInList(sessionID string) string {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.listUser[sessionID]
}

func (t *Tracks) setUserInList(sessionID, index string) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.listUser[sessionID] = index
}

func (t *Tracks) deleteUserInList(sessionID string) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	delete(t.listUser, sessionID)
}

func (t *Tracks) getCheckList() *utils.AdvanceMap {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.checkList
}

func (t *Tracks) setCheckList(index, streamID string) {
	if lst := t.getCheckList(); lst != nil {
		lst.Set(index, streamID)
	}
}

func (t *Tracks) releaseCheckList(index string) {
	if lst := t.getCheckList(); lst != nil {
		lst.Set(index, defaultTrackState)
	}
}

// check input index has stream id or not, return defaultTrackState if not
func (t *Tracks) getInCheckList(index string) string {
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

func (t *Tracks) getCodecs() string {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.codec
}

func (t *Tracks) setCodecs(c string) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.codec = c
}

func (t *Tracks) getPayloadType() int {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.payloadType
}

func (t *Tracks) setPayloadType(c int) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.payloadType = c
}
