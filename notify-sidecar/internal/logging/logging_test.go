package logging

import (
	"os"
	"reflect"
	"testing"

	log "github.com/sirupsen/logrus"
)

func TestSetupLogging(t *testing.T) {
	SetupLogging()
	if log.GetLevel() != log.WarnLevel {
		t.Errorf("got %+v, want %+v", log.GetLevel(), log.WarnLevel)
	}
	if !reflect.DeepEqual(log.StandardLogger().Formatter, &log.JSONFormatter{}) {
		t.Errorf("got %+v, want %+v", log.StandardLogger().Formatter, &log.JSONFormatter{})
	}
	if !reflect.DeepEqual(log.StandardLogger().Out, os.Stdout) {
		t.Errorf("got %+v, want %+v", log.StandardLogger().Out, os.Stdout)
	}
}

func TestSetupLoggingWithEnv(t *testing.T) {
	os.Setenv("NOTIFY_SIDECAR_LOG_LEVEL", "INFO")
	defer os.Unsetenv("NOTIFY_SIDECAR_LOG_LEVEL")
	SetupLogging()
	if log.GetLevel() != log.InfoLevel {
		t.Errorf("got %+v, want %+v", log.GetLevel(), log.InfoLevel)
	}
}

func TestSetupLoggingWithBadEnv(t *testing.T) {
	os.Setenv("NOTIFY_SIDECAR_LOG_LEVEL", "Biggus Dickus")
	defer os.Unsetenv("NOTIFY_SIDECAR_LOG_LEVEL")
	defer func() { _ = recover() }()
	SetupLogging()
	t.Errorf("Setting and invalid log level should result in a panic")
}
