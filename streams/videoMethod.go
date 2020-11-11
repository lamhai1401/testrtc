package streams

import (
	"github.com/beowulflab/mixer-v2/utils"
	"github.com/pion/rtp"
)

func (m *VideoStreamObj) checkClose() bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.isClosed
}

func (m *VideoStreamObj) setClose(state bool) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.isClosed = state
}

func (m *VideoStreamObj) getIM() *utils.IndexManager {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.im
}

func (m *VideoStreamObj) getAudioIn() map[int]chan *rtp.Packet {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.videoInChann
}

func (m *VideoStreamObj) getAudioInChann(index int) chan *rtp.Packet {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.videoInChann[index]
}

func (m *VideoStreamObj) getMixedVideo() chan *rtp.Packet {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.mixedVideo
}

func (m *VideoStreamObj) closeVideoMixing() {
	m.videoMixing.Close()
}
