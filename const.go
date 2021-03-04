package main

import (
	"os"
	"strings"
)

const defaultInterval = 15
const defaultStompWsUrl = "wss://signal-controller-testing.quickom.com/signaling/classroom/"
const defaultContentType = "application/json"
const defaultSignalUrl = "wss://signal-test.dechen.app"
const (
	offlineStatus = "offline"
	onlineStatus  = "online"
)

const (
	destRole   = "dest"
	sourceRole = "source"
)
const errCountToUnregister = 500

const (
	nodeType   = "node"
	clientType = "member"
	mixerType  = "mixer"
)

const splitStr = "-"
const defaultIp = "127.0.0.1"
const defaultTimeout = 30

var (
	nodeLevel = os.Getenv("NODELEVEL")
)

const (
	mainUpType    = "mainUp"
	mainMixedType = "mainMixed"
)

const (
	peeringState = "peering"
	successState = "success"
	failedState  = "failed"
)

const stompSendMsgDelay = 200   // in ms
const resetPeerStateTimeout = 3 // in s

const logTime = 100 //in ms

const (
	stompResponse = "stomp"
	httpResponse  = "http"
)

const (
	defaultCodecVP8   = "VP8"
	defaultCodecVP9   = "VP9"
	defaultPayloadVP8 = 96
	defaultPayloadVP9 = 98
)

// checkPayload return default codec payload type of input codec (vp8, vp9)
func checkPayload(codec string) int {
	var payloadType int

	switch strings.ToLower(codec) {
	case "vp9":
		payloadType = defaultPayloadVP9
		break
	// case "vp8":
	// 	payloadType = defaultPayloadVP8
	// 	break
	default:
		payloadType = defaultPayloadVP8
	}
	return payloadType
}
