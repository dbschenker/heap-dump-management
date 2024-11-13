package utils

import (
	"encoding/base64"
	"testing"
)

func TestGenerateRandomString(t *testing.T) {
	randomBytes, err := GenerateRandomBytes(32)
	rnd := EncodeKey(randomBytes)
	if err != nil {
		t.Errorf(err.Error())
	}
	clearKey, err := base64.StdEncoding.DecodeString(rnd)
	if err != nil {
		t.Errorf(err.Error())
	}
	if len(clearKey) != 32 {
		t.Errorf("String is too short! Should be %d, is %d", 32, len(rnd))
	}
}

func TestGenerateRandomBytes(t *testing.T) {
	rnd, err := GenerateRandomBytes(32)
	if err != nil {
		t.Errorf(err.Error())
	}
	if len(rnd) != 32 {
		t.Errorf("bytes is too short! Should be %d, is %d", 32, len(rnd))
	}
}
