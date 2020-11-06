package utils

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/davecgh/go-spew/spew"
)

// SDPTemp save sdp temp to convert Pion SDP
type SDPTemp struct {
	SDP  string `json:"sdp"`
	Type string `json:"type"`
}

// TurnConfigList get resp body
type TurnConfigList struct {
	Success bool         `json:"success"`
	Message string       `json:"message"`
	Data    []TurnConfig `json:"data"`
}

// TurnConfig handle foreach turn config in list
type TurnConfig struct {
	URLs     string `json:"urls"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// TurnRequestBody linter
type TurnRequestBody struct {
	CallType  string `json:"callType"`
	RequestID string `json:"requestID"`
}

// ToByte convert data to []byte
func ToByte(data interface{}) ([]byte, error) {
	return json.Marshal(data)
}

// ToJSON convert []byte to json
func ToJSON(data []byte, des interface{}) error {
	return json.Unmarshal(data, des)
}

// ToString convert struct to string
func ToString(data interface{}) (string, error) {
	out, err := ToByte(data)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// CutID cutting id to clientID, peerID
func CutID(id string) (string, string) {
	s := strings.SplitN(id, "_", 2)
	if len(s) < 2 {
		return "", ""
	}
	clientID, teacherID := s[0], s[1]
	return clientID, teacherID
}

// MergeID merge clientID, peerID to id
func MergeID(clientID string, peerID string) string {
	return fmt.Sprintf("%s_%s", clientID, peerID)
}

// Handlepanic prevent panic
func handlepanic(data ...interface{}) {
	if a := recover(); a != nil {
		spew.Println("===========This data make signal panic==============")
		spew.Dump(data...)
		fmt.Println("RECOVER", a)
	}
}
