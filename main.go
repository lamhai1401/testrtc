package main

import (
	"os"

	"github.com/lamhai1401/gologs/logs"
	"github.com/lamhai1401/testrtc/peers"
)

func main() {
	os.Setenv("MULTIPLE_URLL", "wss://signal-conference-staging.quickom.com")
	_, err := peers.NewPeers()
	if err != nil {
		logs.Error(err.Error())
	}

	select {}
}
