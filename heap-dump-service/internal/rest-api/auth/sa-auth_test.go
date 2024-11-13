package auth

import (
	"testing"

	_ "github.com/shaj13/libcache/fifo"
)

func TestGoGuardianSetup(t *testing.T) {
	got := setupGoGuardian("test_token")
	if got == nil {
		t.Errorf("Could not build authentication strategy")
	}
}
