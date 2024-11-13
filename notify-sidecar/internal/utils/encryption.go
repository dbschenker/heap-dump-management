package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

const chunkSize = 64 * 1024 // 64 KB

func EncryptDump(fileSystem fs.FS, fileLocation string, key []byte) (string, error) {
	// Reading plaintext file
	inputFile, err := fileSystem.Open(fileLocation)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Error reading heap dump %s: %s", fileLocation, err.Error()))
	}

	// Creating block of algorithm
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Error initializing ARE Cipher: %s", err.Error()))
	}

	// Creating GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Error in GCM Cipher: %s", err.Error()))
	}

	// Generating random nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", errors.New(fmt.Sprintf("Error generating random nonce: %s", err.Error()))
	}

	buffer := make([]byte, chunkSize)
	outputFile, err := os.Create(fmt.Sprintf("/tmp/%s.%s", filepath.Base(fileLocation), "crypted"))
	if err != nil {
		return "", errors.New(fmt.Sprintf("Error creating encrypted heap dump: %s", err.Error()))
	}

	for {
		n, err := inputFile.Read(buffer)
		if err != nil && err != io.EOF {
			return "", errors.New(fmt.Sprintf("Error reading heap dump: %s", err.Error()))
		}
		if n == 0 {
			break
		}

		encryptedChunk := gcm.Seal(nil, nonce, buffer[:n], nil)
		if _, err := outputFile.Write(encryptedChunk); err != nil {
			return "", errors.New(fmt.Sprintf("Error writing encrypted heap dump: %s", err.Error()))
		}
	}

	return fmt.Sprintf("/tmp/%s.%s", filepath.Base(fileLocation), "crypted"), nil
}
