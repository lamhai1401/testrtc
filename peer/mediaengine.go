package peer

import (
	"fmt"
	"strings"

	"github.com/lamhai1401/gologs/logs"
	"github.com/pion/interceptor"
	"github.com/pion/webrtc/v3"
)

// InitAPI linter
// InitAPI linter
func (p *Peer) initAPI() *webrtc.API {
	// init media engine
	m := p.initMediaEngine()
	// Create a InterceptorRegistry. This is the user configurable RTP/RTCP Pipeline.
	// This provides NACKs, RTCP Reports and other features. If you use `webrtc.NewPeerConnection`
	// this is enabled by default. If you are manually managing You MUST create a InterceptorRegistry
	// for each PeerConnection.
	i := &interceptor.Registry{}

	// Use the default set of Interceptors
	if err := webrtc.RegisterDefaultInterceptors(m, i); err != nil {
		logs.Error("initAPI RegisterDefaultInterceptors error: ", err.Error())
		return nil
	}

	return webrtc.NewAPI(
		webrtc.WithMediaEngine(m),
		webrtc.WithInterceptorRegistry(i),
		webrtc.WithSettingEngine(*p.initSettingEngine()))
}

func (p *Peer) initSettingEngine() *webrtc.SettingEngine {
	settingEngine := &webrtc.SettingEngine{}
	// settingEngine.SetTrickle(true)
	// settingEngine.SetEphemeralUDPPortRange(20000, 60000)
	// settingEngine.SetICETimeouts(10*time.Second, 20*time.Second, 1*time.Second)
	return settingEngine
}

func (p *Peer) initMediaEngine() *webrtc.MediaEngine {
	mediaEngine := &webrtc.MediaEngine{}
	// init audio
	err := p.initMediaEngineAudio(mediaEngine)
	if err != nil {
		logs.Error(err)
	}

	err = p.initMediaEngineVideo(p.getPayloadType(), mediaEngine)
	if err != nil {
		logs.Error(err)
	}

	// init video
	return mediaEngine
}

func (p *Peer) initMediaEngineAudio(mediaEngine *webrtc.MediaEngine) error {
	// if err := mediaEngine.RegisterCodec(webrtc.RTPCodecParameters{
	// 	RTPCodecCapability: webrtc.RTPCodecCapability{MimeType: "audio/opus", ClockRate: 48000, Channels: 2, SDPFmtpLine: "minptime=10;useinbandfec=1", RTCPFeedback: nil},
	// 	PayloadType:        111,
	// }, webrtc.RTPCodecTypeAudio); err != nil {
	// 	logs.Error("initMediaEngine: ", err)
	// }
	for _, codec := range []webrtc.RTPCodecParameters{
		{
			RTPCodecCapability: webrtc.RTPCodecCapability{MimeType: MimeTypeOpus, ClockRate: 48000, Channels: 2, SDPFmtpLine: "minptime=10;useinbandfec=1", RTCPFeedback: nil},
			PayloadType:        111,
		},
		{
			RTPCodecCapability: webrtc.RTPCodecCapability{MimeType: MimeTypeG722, ClockRate: 8000, Channels: 0, SDPFmtpLine: "", RTCPFeedback: nil},
			PayloadType:        9,
		},
		{
			RTPCodecCapability: webrtc.RTPCodecCapability{MimeType: MimeTypePCMU, ClockRate: 8000, Channels: 0, SDPFmtpLine: "", RTCPFeedback: nil},
			PayloadType:        0,
		},
		{
			RTPCodecCapability: webrtc.RTPCodecCapability{MimeType: MimeTypePCMA, ClockRate: 8000, Channels: 0, SDPFmtpLine: "", RTCPFeedback: nil},
			PayloadType:        8,
		},
	} {
		if err := mediaEngine.RegisterCodec(codec, webrtc.RTPCodecTypeAudio); err != nil {
			return err
		}
	}

	// Default Pion Audio Header Extensions
	for _, extension := range []string{
		"urn:ietf:params:rtp-hdrext:sdes:mid",
		"urn:ietf:params:rtp-hdrext:sdes:rtp-stream-id",
		"urn:ietf:params:rtp-hdrext:sdes:repaired-rtp-stream-id",
	} {
		if err := mediaEngine.RegisterHeaderExtension(webrtc.RTPHeaderExtensionCapability{URI: extension}, webrtc.RTPCodecTypeAudio); err != nil {
			logs.Error("initMediaEngineAudio: ", err)
		}
	}
	return nil
}

func (p *Peer) initMediaEngineVideo(codecCode int, mediaEngine *webrtc.MediaEngine) error {
	var tmp []webrtc.RTPCodecParameters
	videoRTCPFeedback := []webrtc.RTCPFeedback{{Type: "goog-remb", Parameter: ""}, {Type: "ccm", Parameter: "fir"}, {Type: "nack", Parameter: ""}, {Type: "nack", Parameter: "pli"}}

	switch strings.ToLower(p.getCodecs()) {
	case "vp8":
		tmp = []webrtc.RTPCodecParameters{
			{
				RTPCodecCapability: webrtc.RTPCodecCapability{MimeType: MimeTypeVP8, ClockRate: 90000, Channels: 0, SDPFmtpLine: "", RTCPFeedback: videoRTCPFeedback},
				PayloadType:        webrtc.PayloadType(codecCode),
			},
			{
				RTPCodecCapability: webrtc.RTPCodecCapability{MimeType: "video/rtx", ClockRate: 90000, Channels: 0, SDPFmtpLine: fmt.Sprintf("apt=%d", codecCode), RTCPFeedback: nil},
				PayloadType:        webrtc.PayloadType(codecCode + 1),
			},
		}
	default: // defaul vp9
		tmp = []webrtc.RTPCodecParameters{
			{
				RTPCodecCapability: webrtc.RTPCodecCapability{MimeType: "video/VP9", ClockRate: 90000, Channels: 0, SDPFmtpLine: "profile-id=0", RTCPFeedback: videoRTCPFeedback},
				PayloadType:        webrtc.PayloadType(codecCode),
			},
			{
				RTPCodecCapability: webrtc.RTPCodecCapability{MimeType: "video/rtx", ClockRate: 90000, Channels: 0, SDPFmtpLine: fmt.Sprintf("apt=%d", codecCode), RTCPFeedback: nil},
				PayloadType:        webrtc.PayloadType(codecCode + 1),
			},

			{
				RTPCodecCapability: webrtc.RTPCodecCapability{MimeType: "video/VP9", ClockRate: 90000, Channels: 0, SDPFmtpLine: "profile-id=1", RTCPFeedback: videoRTCPFeedback},
				PayloadType:        webrtc.PayloadType(codecCode + 2),
			},
			{
				RTPCodecCapability: webrtc.RTPCodecCapability{MimeType: "video/rtx", ClockRate: 90000, Channels: 0, SDPFmtpLine: fmt.Sprintf("apt=%d", codecCode+2), RTCPFeedback: nil},
				PayloadType:        webrtc.PayloadType(codecCode + 3),
			},

			{
				RTPCodecCapability: webrtc.RTPCodecCapability{MimeType: "video/ulpfec", ClockRate: 90000, Channels: 0, SDPFmtpLine: "", RTCPFeedback: nil},
				PayloadType:        116,
			},
		}
	}

	// Default Pion Audio Codecs
	for _, codec := range tmp {
		if err := mediaEngine.RegisterCodec(codec, webrtc.RTPCodecTypeVideo); err != nil {
			logs.Error("initMediaEngine: ", err)
		}
	}

	// Default Pion Video Header Extensions
	for _, extension := range []string{
		"urn:ietf:params:rtp-hdrext:sdes:mid",
		"urn:ietf:params:rtp-hdrext:sdes:rtp-stream-id",
		"urn:ietf:params:rtp-hdrext:sdes:repaired-rtp-stream-id",
	} {
		if err := mediaEngine.RegisterHeaderExtension(webrtc.RTPHeaderExtensionCapability{URI: extension}, webrtc.RTPCodecTypeVideo); err != nil {
			logs.Error("initMediaEngine: ", err)
		}
	}
	return nil
}
