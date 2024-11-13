package utils

import (
	"os"
	"strings"
	"testing"
)

func TestFailedUpload(t *testing.T) {
	fileHandler, _ := os.Open("does_not_exist")
	url := "http://localhost:1337"
	got := UploadToS3(url, fileHandler)
	wantNoNetwork := "Error making request:"
	if !(strings.Contains(got.Error(), wantNoNetwork)) {
		t.Errorf("got wrong error %+v, want %+v", got.Error(), wantNoNetwork)
	}
}
