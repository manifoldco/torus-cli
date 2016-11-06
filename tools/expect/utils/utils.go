package utils

import (
	"math/rand"
	"time"
)

// SemverRegex is used for version matching
const SemverRegex = `[0-9]+\.[0-9]+\.[0-9]+`

// GlobalTimeout is used for cases which do not have their own timeout
const GlobalTimeout time.Duration = time.Duration(10) * time.Second
const letterBytes = "abcdefghijklmnopqrstuvwxyz0987654321"

// GlobalNonce is unique string appended to new resource names
var GlobalNonce string

// Nonce generates a unique set of characters for a given int length
func Nonce(n int) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

// Duration will generate a duration value for a given int
func Duration(n int) time.Duration {
	return time.Duration(n) * time.Second
}

// Init establishes the global values
func Init() {
	GlobalNonce = Nonce(10)
}
