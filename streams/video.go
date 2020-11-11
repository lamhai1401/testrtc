package streams

import (
	"fmt"
	"sync"

	"github.com/beowulflab/mediamixer/videoLayoutMixer"
	videoStaticMixer "github.com/beowulflab/mediamixer/videoLayoutMixer/staticMixer"
	"github.com/beowulflab/mixer-v2/utils"
	"github.com/pion/rtp"
)

var (
	cacheLength = 1000
	empty       = "empty"
	// notEmpty = "notempty"
	errNilIm = fmt.Errorf("IM is nil")
)

// VideoStreamObj linter
type VideoStreamObj struct {
	im           *utils.IndexManager
	mixedVideo   chan *rtp.Packet
	videoInChann map[int]chan *rtp.Packet
	videoMixing  videoLayoutMixer.Mixer
	isClosed     bool
	mutex        sync.RWMutex
}

// NewVideoStreamObj linter
func NewVideoStreamObj(
	length int,
	streamID string,
) *VideoStreamObj {
	m := &VideoStreamObj{
		im:           utils.NewIndexManager(length),
		isClosed:     false,
		mixedVideo:   make(chan *rtp.Packet, cacheLength),
		videoInChann: make(map[int]chan *rtp.Packet),
		videoMixing:  videoStaticMixer.GetGstreamerMixer(length),
	}

	return m
}

// IsExist check peer connection id has index or not
func (m *VideoStreamObj) IsExist(peerConnectionID string) (int, bool) {
	if im := m.getIM(); im != nil {
		index, _ := im.CheckIndex(peerConnectionID)
		if index != 0 {
			return index, true
		}
	}
	return 0, false
}

// PushVideo push data of input peer connection id except its self
func (m *VideoStreamObj) PushVideo(index int, data *rtp.Packet) error {
	im := m.getIM()
	if im == nil {
		return errNilIm
	}
	if chann := m.getAudioInChann(index); chann != nil {
		chann <- data
	}
	return nil
}

// AddVideo add new video source, return current index and error
func (m *VideoStreamObj) AddVideo(peerConnectionID string) (int, error) {
	// find index
	var index int
	var err error
	im := m.getIM()
	if im == nil {
		return 0, errNilIm
	}
	index, err = im.FindIndex(peerConnectionID)
	if err != nil {
		return 0, err
	}
	return index, err
}

// RemoveVideo remove existing audio source
func (m *VideoStreamObj) RemoveVideo(peerConnectionID string) {
	if im := m.getIM(); im != nil {
		im.ReleaseIndex(peerConnectionID)
	}
}

// Start linter
func (m *VideoStreamObj) Start() error {
	err := m.videoMixing.Start(m.getAudioIn(), m.getMixedVideo())
	if err != nil {
		return err
	}
	return nil
}

// Close linter
func (m *VideoStreamObj) Close() {
	if m.checkClose() {
		m.setClose(true)
		m.closeVideoMixing()
	}
}

// GetMixedVideo linter
func (m *VideoStreamObj) GetMixedVideo() chan *rtp.Packet {
	return m.getMixedVideo()
}
