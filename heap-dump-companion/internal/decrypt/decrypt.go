package decrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"errors"
	"fmt"
	"io/fs"
	"os"
)

func DecryptFile(fileSystem fs.FS, key []byte, encryptedFileLocation string, desiredOutputFileLocation string) error {
	// Reading ciphertext file
	cipherText, err := fs.ReadFile(fileSystem, encryptedFileLocation)
	if err != nil {
		return errors.New(fmt.Sprintf("Error reading heap dump %s: %s", encryptedFileLocation, err.Error()))
	}

	// Creating block of algorithm
	block, err := aes.NewCipher(key)
	if err != nil {
		return errors.New(fmt.Sprintf("Error initializing ARE Cipher: %s", err.Error()))
	}

	// Creating GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return errors.New(fmt.Sprintf("Error in GCM Cipher: %s", err.Error()))
	}

	// Deattached nonce and decrypt
	nonce := cipherText[:gcm.NonceSize()]
	cipherText = cipherText[gcm.NonceSize():]
	plainText, err := gcm.Open(nil, nonce, cipherText, nil)
	if err != nil {
		return errors.New(fmt.Sprintf("Decrypting file %s failed: %s", encryptedFileLocation, err.Error()))
	}

	// Writing decryption content
	err = os.WriteFile(desiredOutputFileLocation, plainText, 0666)
	if err != nil {
		return errors.New(fmt.Sprintf("Error writing decrypted heap dump: %s", err.Error()))
	}
	return nil
}
