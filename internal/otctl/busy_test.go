package otctl

import (
	"errors"
	"fmt"
	"testing"
)

func TestIsDaemonBusy(t *testing.T) {
	if IsDaemonBusy(nil) {
		t.Fatal("nil is not busy")
	}
	if !IsDaemonBusy(ErrDaemonBusy) {
		t.Fatal("ErrDaemonBusy")
	}
	if !IsDaemonBusy(fmt.Errorf("%w: another command is in progress", ErrDaemonBusy)) {
		t.Fatal("wrapped ErrDaemonBusy")
	}
	if !IsDaemonBusy(errors.New(`"history netinfo 30": Error 5: Busy`)) {
		t.Fatal("daemon Error 5")
	}
	if IsDaemonBusy(errors.New("Error 35: InvalidCommand")) {
		t.Fatal("other errors are not busy")
	}
}
