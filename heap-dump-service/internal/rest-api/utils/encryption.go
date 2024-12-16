package utils

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
)

func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Could not generate random bytes %s", err.Error()))
	}
	return b, nil
}

func EncodeKey(key []byte) string {
	return base64.StdEncoding.EncodeToString(key)
}
