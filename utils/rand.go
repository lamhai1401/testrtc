package utils

import (
	"math/rand"
	"time"

	"github.com/segmentio/ksuid"
)

// GenerateID to create new string id
func GenerateID() string {
	id := ksuid.New().String()
	return id
}

// RandomInt to return value in rang (min, max)
func RandomInt(min, max int) int {
	rand.Seed(time.Now().Unix())
	return rand.Intn(max-min) + min
}
