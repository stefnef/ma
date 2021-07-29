package crypt

import (
	"encoding/base64"
	"testing"
)

func TestGenerateToken(t *testing.T) {
	token := GenerateToken()
	if token == "" {
		t.Error("could not generate new Token")
		t.Fail()
	}
	tokenByte, err := base64.URLEncoding.DecodeString(token)
	if len(tokenByte) != tokenLength || err != nil {
		t.Error(err)
		t.Errorf("wrong Token length: %d != %d", tokenLength, len(tokenByte))
		t.Fail()
	}

}
