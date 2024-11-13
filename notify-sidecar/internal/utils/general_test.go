package utils

import (
	"reflect"
	"strings"
	"testing"
	"testing/fstest"
)

func TestCheckError(*testing.T) {
	CheckError(nil)
}

func TestNamespace(t *testing.T) {
	fs := fstest.MapFS{
		"var/run/secrets/kubernetes.io/serviceaccount/namespace": {Data: []byte("testNamespace")},
	}
	want := "testNamespace"
	wantError := "Could not get Namespace"
	got, err := GetCurrentNamespace(fs)

	if err != nil {
		t.Errorf("Failed to encrypt test file: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}
	fs = fstest.MapFS{
		"NotInK8s": {Data: []byte("testNamespace")},
	}
	got, err = GetCurrentNamespace(fs)
	if err == nil {
		t.Errorf("This should Fail")
	}
	if !(strings.Contains(err.Error(), wantError)) {
		t.Errorf("got %+v, want %+v", err.Error(), wantError)
	}
}
