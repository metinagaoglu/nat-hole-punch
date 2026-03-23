package handler

import (
	"fmt"
	"regexp"
)

const (
	maxKeyLength     = 64
	maxPayloadLength = 512
)

var validKeyPattern = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

func validatePayload(payload string) error {
	if len(payload) == 0 {
		return fmt.Errorf("payload cannot be empty")
	}
	if len(payload) > maxPayloadLength {
		return fmt.Errorf("payload too large: %d bytes (max %d)", len(payload), maxPayloadLength)
	}
	return nil
}

func validateKey(key string) error {
	if len(key) == 0 {
		return fmt.Errorf("room key cannot be empty")
	}
	if len(key) > maxKeyLength {
		return fmt.Errorf("room key too long: %d chars (max %d)", len(key), maxKeyLength)
	}
	if !validKeyPattern.MatchString(key) {
		return fmt.Errorf("room key contains invalid characters (allowed: a-z, A-Z, 0-9, -, _)")
	}
	return nil
}
