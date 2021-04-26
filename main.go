package main

import (
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime/debug"
	"strconv"

	"github.com/lamhai1401/gologs/logs"
)

func main() {
	// m, err := NewPeerWorker(
	// 	os.Getenv("WSS_URL"),  // "wss://signal-conference-staging.quickom.com",
	// 	os.Getenv("USERNAME"), // "hai",
	// )

	m, err := NewPeerWorker(
		"wss://signal-conference-staging.quickom.com",
		"hai2",
	)

	if err != nil {
		logs.Error(err.Error())
	}
	logs.Warn(m)

	debug.WriteHeapDump(10)
	http.ListenAndServe("localhost:8080", nil)

	select {}
}

func getInterval() int {
	i := 5
	if interval := os.Getenv("REPEER_GAP_TIME"); interval != "" {
		j, err := strconv.Atoi(interval)
		if err == nil {
			i = j
		}
	}
	return i
}

// (rtp.Header) {
// 	Version: (uint8) 2,
// 	Padding: (bool) false,
// 	Extension: (bool) false,
// 	Marker: (bool) false,
// 	PayloadOffset: (int) 12,
// 	PayloadType: (uint8) 98,
// 	SequenceNumber: (uint16) 11165,
// 	Timestamp: (uint32) 4028086595,
// 	SSRC: (uint32) 434924193,
// 	CSRC: ([]uint32) {
// 	},
// 	ExtensionProfile: (uint16) 0,
// 	Extensions: ([]rtp.Extension) <nil>
//    }

//    (rtp.Header) {
// 	Version: (uint8) 2,
// 	Padding: (bool) false,
// 	Extension: (bool) true,
// 	Marker: (bool) true,
// 	PayloadOffset: (int) 20,
// 	PayloadType: (uint8) 98,
// 	SequenceNumber: (uint16) 12225,
// 	Timestamp: (uint32) 3921756224,
// 	SSRC: (uint32) 3522144344,
// 	CSRC: ([]uint32) {
// 	},
// 	ExtensionProfile: (uint16) 48862,
// 	Extensions: ([]rtp.Extension) (len=1 cap=1) {
// 	 (rtp.Extension) {
// 	  id: (uint8) 4,
// 	  payload: ([]uint8) (len=1 cap=1443) {
// 	   00000000  31                                                |1|
// 	  }
// 	 }
// 	}
//    }
