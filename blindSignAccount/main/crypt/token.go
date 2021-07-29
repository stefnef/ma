package crypt

import (
	"crypto/rand"
	"encoding/base64"
)

const tokenLength = 16 // bytes

func GenerateToken() string {

	b := make([]byte, tokenLength)
	if _, err := rand.Read(b); err != nil {
		return ""
	}

	return base64.URLEncoding.EncodeToString(b)
}
