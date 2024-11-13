package config

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"testing"
)

var Want = AppConfig{
	Metrics: struct {
		Port int
		Path string
	}{
		Port: 8081,
		Path: "/metrics",
	},
	WatchPath: struct{ Path string }{
		Path: "/test",
	},
	Middleware: struct{ Endpoint string }{
		Endpoint: "https://test.svc.cluster.local",
	},
	ServiceOwner: struct {
		Tenant string
	}{
		Tenant: "testTenant",
	},
}

func TestLoadConfigFromFile(t *testing.T) {
	got, err := LoadConfigFromFile("../../config/test/test-config.json")
	if err != nil {
		t.Errorf("Failed to construct config: %v", err)
	}
	if !reflect.DeepEqual(got, Want) {
		t.Errorf("got %+v, want %+v", got, Want)
	}
}

func TestFailLoadConfigFromFile(t *testing.T) {
	want := errors.New(fmt.Sprintf("Failed to load config file '%v': %v", "does-not-exist.json", "open does-not-exist.json: no such file or directory"))
	_, got := LoadConfigFromFile("does-not-exist.json")
	if got == nil {
		t.Errorf("This should produce an error!")
	}
	if !reflect.DeepEqual(got.Error(), want.Error()) {
		t.Errorf("got %+v, want %+v", got, want)
	}
	want = errors.New(fmt.Sprintf("Failed to parse json data of file '%v': %v", "../../config/test/bad-config.json", "json: unknown field \"invalid\""))
	_, got = LoadConfigFromFile("../../config/test/bad-config.json")
	if got == nil {
		t.Errorf("This should produce an error!")
	}
	if !reflect.DeepEqual(got.Error(), want.Error()) {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestLoadConfigFromEnv(t *testing.T) {
	os.Setenv("TEST_APP_CONFIG_FILE", "../../config/test/test-config.json")
	defer os.Unsetenv("TEST_APP_CONFIG_FILE")
	got, err := LoadConfigFromEnvironment("TEST_APP_CONFIG_FILE")
	if err != nil {
		t.Errorf("Failed to construct config: %v", err)
	}
	if !reflect.DeepEqual(got, Want) {
		t.Errorf("got %+v, want %+v", got, Want)
	}
}

func TestFailLoadConfigFromEnvironment(t *testing.T) {
	want := errors.New(fmt.Sprintf("Failed to load config file '%v': %v", "does-not-exist.json", "open does-not-exist.json: no such file or directory"))
	os.Setenv("TEST_APP_CONFIG_FILE", "does-not-exist.json")
	defer os.Unsetenv("TEST_APP_CONFIG_FILE")
	_, got := LoadConfigFromEnvironment("TEST_APP_CONFIG_FILE")
	if got == nil {
		t.Errorf("This should produce an error!")
	}
	if !reflect.DeepEqual(got.Error(), want.Error()) {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestFailNoConfigEnv(t *testing.T) {
	want := errors.New(fmt.Sprintf("Environment variable for config file not set: %s", "TEST_NO_ENV"))
	_, got := LoadConfigFromEnvironment("TEST_NO_ENV")
	if got == nil {
		t.Errorf("This should produce an error!")
	}
	if !reflect.DeepEqual(got.Error(), want.Error()) {
		t.Errorf("got %+v, want %+v", got, want)
	}
}
