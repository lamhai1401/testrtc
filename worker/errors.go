package worker

import "fmt"

var (
	// ErrNilPeers linter
	ErrNilPeers = fmt.Errorf("Peers is nil")
)

const defaultInterval = 15
