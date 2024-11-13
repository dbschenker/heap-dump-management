package utils

import (
	"crypto/rand"
	"errors"
	"fmt"
	"os"
	"reflect"
	"testing"
	"testing/fstest"
)

func cleanup(target string) {
	os.Remove(target)
}

func TestEncryption(t *testing.T) {
	fs := fstest.MapFS{
		"test_heap_dump": {Data: []byte("asdfasdfasdf")},
	}
	test_key := []byte{52, 74, 93, 7, 97, 74, 50, 186, 172, 14, 125, 208, 130, 218, 177, 215, 219, 219, 247, 163, 81, 86, 105, 60, 22, 162, 54, 81, 19, 37, 212, 49}
	want := "/tmp/test_heap_dump.crypted"
	got, err := EncryptDump(fs, "test_heap_dump", test_key)

	if err != nil {
		t.Errorf("Failed to encrypt test file: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}
	_, err = os.Stat(want)
	if err != nil {
		t.Errorf("Encrypted test file does not exist: %v", err)
	}
	cleanup(want)
}

func TestBadEncryption(t *testing.T) {
	fs := fstest.MapFS{
		"test_heap_dump": {Data: []byte("asdfasdfasdf")},
	}
	test_key := make([]byte, 8)
	rand.Read(test_key)

	_, got := EncryptDump(fs, "test_heap_dump", test_key)
	want := errors.New(fmt.Sprintf("Error initializing ARE Cipher: %s", "crypto/aes: invalid key size 8"))

	if got == nil {
		t.Errorf("AES should not be used with a key size of 8")
	}
	if !reflect.DeepEqual(got.Error(), want.Error()) {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestFileNotFound(t *testing.T) {
	fs := fstest.MapFS{
		"test_heap_dump": {Data: []byte("asdfasdfasdf")},
	}
	test_key := make([]byte, 8)
	rand.Read(test_key)

	_, got := EncryptDump(fs, "invalid", test_key)
	want := errors.New(fmt.Sprintf("Error reading heap dump %s: %s", "invalid", "open invalid: file does not exist"))

	if got == nil {
		t.Errorf("File should not exist")
	}
	if !reflect.DeepEqual(got.Error(), want.Error()) {
		t.Errorf("got %+v, want %+v", got, want)
	}
}
