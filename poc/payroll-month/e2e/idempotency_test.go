package e2e

// Duplicate-submission behavior (prompt8 §4.2/§4.4): attempt IDs are
// Reject-on-reuse — the same attempt submitted twice fails the second
// submission LOUDLY, and business state shows a single effect.
// (Cross-attempt idempotency is adoption, proven in dbos_test.go.)

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func TestDoubleEnqueue_SingleExecution(t *testing.T) {
	pool := pocPool(t)
	ctx := context.Background()
	bin := buildPocRun(t)
	runID := fmt.Sprintf("idem-dup-%d", time.Now().UnixNano())

	cmd := exec.Command(bin, "capture-once", "--run-id", runID)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("first submit: %v\n%s", err, out)
	}
	cmd = exec.Command(bin, "capture-once", "--run-id", runID)
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("duplicate attempt accepted, want loud rejection:\n%s", out)
	}
	if !strings.Contains(string(out), "already in use") && !strings.Contains(err.Error(), "already in use") {
		// The CLI wraps the DBOS error; accept either surface but
		// require the Reject signal, not a generic failure.
		combined := string(out) + err.Error()
		if !strings.Contains(combined, "WorkflowIDInUse") && !strings.Contains(combined, "already") {
			t.Fatalf("want Reject signal, got:\n%s\n%v", out, err)
		}
	}
	// Business effect: exactly one SUCCESS for the attempt, one run.
	var n int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM dbos.workflow_status
		WHERE workflow_uuid = $1 AND status = 'SUCCESS'`, "capture:"+runID).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("SUCCESS rows = %d, want exactly 1", n)
	}
	var runs int
	if err := pool.QueryRow(ctx,
		`SELECT count(*) FROM payroll_run r JOIN employer e ON e.id = r.employer_id
		  WHERE e.code = 'POC-MERIDIAN'`).Scan(&runs); err != nil || runs != 1 {
		t.Fatalf("business runs = %d err=%v, want 1", runs, err)
	}
}
