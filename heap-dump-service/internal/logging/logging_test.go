package logging

import (
	"os"
	"reflect"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
)

func TestSetupLogging(t *testing.T) {
	SetupLogging()
	if log.GetLevel() != log.InfoLevel {
		t.Errorf("got %+v, want %+v", log.GetLevel(), log.InfoLevel)
	}
	if !reflect.DeepEqual(log.StandardLogger().Formatter, &log.JSONFormatter{}) {
		t.Errorf("got %+v, want %+v", log.StandardLogger().Formatter, &log.JSONFormatter{})
	}
	if !reflect.DeepEqual(log.StandardLogger().Out, os.Stdout) {
		t.Errorf("got %+v, want %+v", log.StandardLogger().Out, os.Stdout)
	}
}

func TestDuration(t *testing.T) {
	start := time.Now()
	duration := GetDurationInMillseconds(start)
	if duration != 0 {
		t.Errorf("Duration Calculation took too long! %v", duration)
	}
}
