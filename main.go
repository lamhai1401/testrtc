package main

import (
	_ "net/http/pprof"

	"github.com/lamhai1401/gologs/logs"
)

func main() {
	m, err := NewPeerWorker(
		"wss://signal-conference-staging.quickom.com",
		"hai",
	)

	if err != nil {
		logs.Error(err.Error())
	}
	logs.Warn(m)
	select {}
}
