package utils

import (
	"fmt"
	"os"
	"strconv"

	"github.com/lamhai1401/gologs/logs"
)

func init() {
	NodeLevel = GetNodeLevel()
}

var (
	// MixerStreamID linter
	MixerStreamID = os.Getenv("MIXERSTREAMID")
	mixerLength   = os.Getenv("MIXERLENGTH")
	nodeLevel     = os.Getenv("NODELEVEL")
	// NodeLevel linter
	NodeLevel = -1
)

// GetNodeLevel get level of current node, default is 0
func GetNodeLevel() int {
	if nodeLevel == "" {
		return -1
	}

	level, err := strconv.Atoi(nodeLevel)
	if err != nil {
		logs.Error("Get node level err: ", err.Error())
		return -1
	}

	return level
}

// GetMixerLength get numbers of mixer inputs, default is 1
func GetMixerLength() int {
	if mixerLength == "" {
		return 1
	}

	level, err := strconv.Atoi(mixerLength)
	if err != nil {
		logs.Error("Get node level err: ", err.Error())
		return 1
	}

	return level
}

// FormatSignalID return format signalID
func FormatSignalID(signalID, role, sessionID string) string {
	return fmt.Sprintf("%s_%s_%s", signalID, role, sessionID)
}
