package peer

import (
	"fmt"

	"github.com/pion/webrtc/v3"
)

// initTransceiver init peer Transceiver
func initTransceiver(conn *webrtc.PeerConnection, kind string, s *webrtc.RTPSender, track webrtc.TrackLocal) (*webrtc.RTPTransceiver, error) {
	if conn == nil {
		return nil, fmt.Errorf("Peer connection is nil")
	}

	// var k webrtc.RTPCodecType

	// if kind == "video" {
	// 	k = webrtc.RTPCodecTypeVideo
	// } else {
	// 	k = webrtc.RTPCodecTypeAudio
	// }

	t, err := conn.AddTransceiverFromTrack(track, webrtc.RTPTransceiverInit{Direction: webrtc.RTPTransceiverDirectionSendrecv})
	// t, err := conn.AddTransceiverFromKind(k, webrtc.RTPTransceiverInit{Direction: webrtc.RTPTransceiverDirectionSendrecv})
	if err != nil {
		return nil, err
	}

	err = t.SetSender(s, track)
	if err != nil {
		return nil, err
	}
	return t, nil
}
