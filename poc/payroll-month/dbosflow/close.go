package dbosflow

import (
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ClosePoolTimeout closes a pool shared with a DBOS runtime, waiting at most
// d. DBOS-Go v1.8.0 Shutdown is timeout-bounded but does not reliably release
// every shared-pool connection (proven at Gate B: "system database
// connection pool did not close in time", then pgxpool.Close blocks
// indefinitely in puddle WaitGroup.Wait while a leaked background goroutine
// holds a conn). A plain pool.Close() after Shutdown therefore hangs the
// process nondeterministically. This helper waits briefly, then logs and
// continues: leaked connections are reaped at process exit (CLI) or test
// binary exit (e2e), harmless on the disposable POC database. Revisit on
// DBOS version bumps; file upstream with the POC_DBOS_DEBUG shutdown trace.
func ClosePoolTimeout(pool *pgxpool.Pool, d time.Duration) {
	done := make(chan struct{})
	go func() {
		pool.Close()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(d):
		fmt.Fprintf(os.Stderr, "dbosflow: pool.Close timed out after %v; continuing (conns reaped at exit)\n", d)
	}
}
