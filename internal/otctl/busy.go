package otctl

import (
	"errors"
	"strings"
)

// IsDaemonBusy reports whether err means the OpenThread daemon rejected the
// command because another session is already using it (local lock or Error 5).
func IsDaemonBusy(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, ErrDaemonBusy) {
		return true
	}
	return strings.Contains(err.Error(), "Error 5:") && strings.Contains(err.Error(), "Busy")
}
