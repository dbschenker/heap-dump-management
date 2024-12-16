package decrypt

import (
	"os"
	"strings"
	"testing"
	"testing/fstest"
)

func cleanup(target string) {
	os.Remove(target)
}

func TestDecrypt(t *testing.T) {
	fs := fstest.MapFS{
		"test_heap_dump.crypted": {Data: []byte{249, 184, 229, 140, 162, 106, 204, 138, 217, 134, 0, 193, 0, 94, 138, 198, 87, 151, 61, 2, 150, 92, 171, 128, 156, 23, 5, 153, 140, 69, 83, 173, 163, 164, 4, 58, 155, 75, 53, 198}},
	}
	test_key := []byte{52, 74, 93, 7, 97, 74, 50, 186, 172, 14, 125, 208, 130, 218, 177, 215, 219, 219, 247, 163, 81, 86, 105, 60, 22, 162, 54, 81, 19, 37, 212, 49}
	testTargetFile := "/tmp/test_heap_dump"
	want := "asdfasdfasdf"
	err := DecryptFile(fs, test_key, "test_heap_dump.crypted", testTargetFile)

	if err != nil {
		t.Errorf("Failed to encrypt test file: %v", err)
	}
	_, err = os.Stat(testTargetFile)
	if err != nil {
		t.Errorf("Encrypted test file does not exist: %v", err)
	}

	clearText, err := os.ReadFile(testTargetFile)
	if err != nil {
		t.Errorf("Encrypted test file does not exist: %v", err)
	}
	if string(clearText) != want {
		t.Errorf("Unexpected decrypted data: want: %s, got %s", want, string(clearText))
	}
	cleanup(want)
}

func TestFileNotFound(t *testing.T) {
	fs := fstest.MapFS{
		"test_heap_dump_with_typo": {Data: []byte{249, 184, 229, 140, 162, 106, 204, 138, 217, 134, 0, 193, 0, 94, 138, 198, 87, 151, 61, 2, 150, 92, 171, 128, 156, 23, 5, 153, 140, 69, 83, 173, 163, 164, 4, 58, 155, 75, 53, 198}},
	}
	test_key := []byte{52, 74, 93, 7, 97, 74, 50, 186, 172, 14, 125, 208, 130, 218, 177, 215, 219, 219, 247, 163, 81, 86, 105, 60, 22, 162, 54, 81, 19, 37, 212, 49}
	testTargetFile := "/tmp/test_heap_dump"
	want := "Error reading heap dump"
	err := DecryptFile(fs, test_key, "test_heap_dump.crypted", testTargetFile)

	if err == nil {
		t.Errorf("Failed to encrypt test file: %v", err)
	}

	if !strings.Contains(err.Error(), want) {
		t.Errorf("Got %s want: %s", err.Error(), want)
	}
}

func TestBadAESKey(t *testing.T) {
	fs := fstest.MapFS{
		"test_heap_dump.crypted": {Data: []byte{249, 184, 229, 140, 162, 106, 204, 138, 217, 134, 0, 193, 0, 94, 138, 198, 87, 151, 61, 2, 150, 92, 171, 128, 156, 23, 5, 153, 140, 69, 83, 173, 163, 164, 4, 58, 155, 75, 53, 198}},
	}
	test_key := []byte{52, 74, 93, 7, 97, 74, 50, 186, 172, 14, 125, 208, 130, 218, 177, 215, 219, 219, 247, 163, 81, 86, 105, 60}
	testTargetFile := "/tmp/test_heap_dump"
	want := "message authentication failed"
	err := DecryptFile(fs, test_key, "test_heap_dump.crypted", testTargetFile)

	if err == nil {
		t.Errorf("Failed to encrypt test file: %v", err)
	}

	if !strings.Contains(err.Error(), want) {
		t.Errorf("Got %s want: %s", err.Error(), want)
	}
}

func TestBadOutputFile(t *testing.T) {
	fs := fstest.MapFS{
		"test_heap_dump.crypted": {Data: []byte{249, 184, 229, 140, 162, 106, 204, 138, 217, 134, 0, 193, 0, 94, 138, 198, 87, 151, 61, 2, 150, 92, 171, 128, 156, 23, 5, 153, 140, 69, 83, 173, 163, 164, 4, 58, 155, 75, 53, 198}},
	}
	test_key := []byte{52, 74, 93, 7, 97, 74, 50, 186, 172, 14, 125, 208, 130, 218, 177, 215, 219, 219, 247, 163, 81, 86, 105, 60, 22, 162, 54, 81, 19, 37, 212, 49}
	testTargetFile := "/asdf/test_heap_dump"
	want := "Error writing decrypted heap dump"
	err := DecryptFile(fs, test_key, "test_heap_dump.crypted", testTargetFile)

	if err == nil {
		t.Errorf("Failed to encrypt test file: %v", err)
	}

	if !strings.Contains(err.Error(), want) {
		t.Errorf("Got %s want: %s", err.Error(), want)
	}
}
