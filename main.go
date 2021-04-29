package main

import (
	"net/http"
	_ "net/http/pprof"
	"runtime/debug"

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
}
