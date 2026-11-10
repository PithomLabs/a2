// Package chaos holds test-only failure injection. Trip blocks the process
// (sleep) when POC_FAILPOINT matches, giving the kill-matrix harness a
// deterministic window to SIGKILL. Production builds must never set
// POC_FAILPOINT; Trip is a no-op when it is unset.
package chaos

import (
	"fmt"
	"os"
	"time"
)

// Trip blocks for 5 minutes when POC_FAILPOINT == name, logging a marker the
// harness waits for before killing. Otherwise it returns immediately.
func Trip(name string) {
	if os.Getenv("POC_FAILPOINT") != name {
		return
	}
	fmt.Fprintf(os.Stderr, "failpoint armed: %s (pid %d, sleeping)\n", name, os.Getpid())
	time.Sleep(5 * time.Minute)
}
