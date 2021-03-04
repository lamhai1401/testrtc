package main

import (
	"fmt"
	"strings"
)

var (
	// ErrNilWorker linter
	ErrNilWorker = fmt.Errorf("Worker is nil")
	// ErrNilSignal linter
	ErrNilSignal = fmt.Errorf("Notify signal is nil")
	// ErrPeerPerMission linter
	ErrPeerPerMission = fmt.Errorf("Don't have permission to peer")
)

func getFirstPayload(str, codec string) (string, error) {
	list := strings.Split(strings.Replace(str, "\r\n", "\n", -1), "\n")
	for _, item := range list {
		result, err := getPayloadFromCodec(item, codec)
		if err == nil && len(result) > 0 {
			return result, err
		}
	}
	return "", fmt.Errorf(fmt.Sprintf("Can't find payload for codec %s in %s", codec, str))
}

func getPayloadFromCodec(str, codec string) (string, error) {
	// fmt.Println("Start find codec for str ", str)
	//Convert to lowercase
	data := strings.ToLower(str)
	newCodec := strings.ToLower(codec)
	lastIndex := strings.LastIndex(data, newCodec)
	// fmt.Println(lastIndex)
	if lastIndex <= 0 {
		// fmt.Println("Can't find codec")
		return "", fmt.Errorf(fmt.Sprintf("Can't find codec %s in %s", newCodec, data))
	}
	//Get substring from beginning to last index
	end := lastIndex - 1
	substring := data[0:end]
	// fmt.Println(substring)
	// fmt.Println(substring[end:end])
	if substring[end:end] != "" {
		// fmt.Println("String invalid")
		return "", fmt.Errorf(fmt.Sprintf("String %s invalid with codec %s", data, newCodec))
	}
	//Trim space at the last of substring
	newSubstring := strings.TrimSpace(substring)
	//Get index of the last " " in substring
	lastSpaceIndex := strings.LastIndex(newSubstring, " ")
	payLoadStr := ""
	if lastSpaceIndex < 0 {
		payLoadStr = newSubstring
	} else {
		start := lastSpaceIndex + 1
		payLoadStr = newSubstring[start:]
	}
	// fmt.Println(payLoadStr)
	//Get index of the last : in payLoadStr
	li := strings.LastIndex(payLoadStr, ":")
	if li < 0 {
		// fmt.Println("Can't find payload")
		return "", fmt.Errorf(fmt.Sprintf("Can't find payload for codec %s in %s", newCodec, data))
	}
	beginning := li + 1
	if beginning > end {
		// fmt.Println("Can't find payload")
		return "", fmt.Errorf(fmt.Sprintf("Can't find payload for codec %s in %s", newCodec, data))
	}
	//Get payload
	payload := payLoadStr[beginning:end]
	// fmt.Println(payload)
	//Trim space
	result := strings.TrimSpace(payload)
	// fmt.Println(result)
	return result, nil
}
