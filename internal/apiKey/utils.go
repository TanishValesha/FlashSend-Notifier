package apikey

import (
	"crypto/rand"
	"encoding/hex"
)

func GenerateAPIKey() (string, error) {
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}

	return "fs_apiKey_" + hex.EncodeToString(bytes), nil
}
