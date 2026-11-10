package e2e

// Forged-result tripwire (prompt8 §4.3/DB-5): a SUCCESS row forged
// directly in DBOS system tables for an attempt ID must FAIL CLOSED
// at the submitter — the forged output cannot be corroborated by
// persisted business rows, so capture-once exits nonzero and moves
// no business state. A real attempt alongside is unaffected.

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func dbosSchema() string {
	if s := os.Getenv("POC_DBOS_SCHEMA"); s != "" {
		return s
	}
	return "dbos"
}

func TestForgedState_MovesNothing(t *testing.T) {
	pool := pocPool(t)
	ctx := context.Background()
	schema := dbosSchema()
	uniq := fmt.Sprintf("%d", time.Now().UnixNano())
	forgedID := "capture:forged-" + uniq
	fakeRun := "ffffffff-ffff-4fff-8fff-ffffffffffff"

	before := businessCounts(t, pool)

	// 1. Forge a SUCCESS row for the attempt ID with invented output
	//    in the real CaptureResult shape (least privilege does NOT
	//    hold on dbos.* — same-DB grants are unavoidable — so the
	//    forge lands; the executor must ignore it).
	if _, err := pool.Exec(ctx, `INSERT INTO `+schema+`.workflow_status
		(workflow_uuid, status, name, output) VALUES ($1, 'SUCCESS', 'captureWorkflow',
		 '{"workflow_id":"`+forgedID+`","run_id":"`+fakeRun+`",`+
		`"snapshot_id":"ffffffff-ffff-4fff-8fff-fffffffffffe",`+
		`"input_digest":"FORGED","role":"captured"}')
		ON CONFLICT (workflow_uuid) DO NOTHING`, forgedID); err != nil {
		t.Fatalf("forge status: %v", err)
	}

	// 2. Submitting that attempt ID is REJECTED (Reject-on-reuse):
	//    the forged output is never returned, never trusted.
	bin := buildPocRun(t)
	cmd := exec.Command(bin, "capture-once", "--run-id", "forged-"+uniq)
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("forged SUCCESS accepted:\n%s", out)
	}
	if !strings.Contains(string(out)+err.Error(), "already in use") {
		t.Fatalf("want Reject signal, got:\n%s\n%v", out, err)
	}

	// 3. No business state moved: counts identical, no run for the
	//    forged ID, no attempt claim.
	after := businessCounts(t, pool)
	if before != after {
		t.Fatalf("forged rows moved business state: %q -> %q", before, after)
	}
	var n int
	if err := pool.QueryRow(ctx,
		`SELECT count(*) FROM payroll_run WHERE id = $1`, fakeRun).Scan(&n); err != nil || n != 0 {
		t.Fatalf("forged run exists: n=%d err=%v", n, err)
	}
	if err := pool.QueryRow(ctx,
		`SELECT count(*) FROM poc_workflow_attempt WHERE workflow_id = $1`,
		forgedID).Scan(&n); err != nil || n != 0 {
		t.Fatalf("forged attempt claimed: n=%d err=%v", n, err)
	}

	// 4. A REAL attempt alongside is unaffected.
	realID := "forged-real-" + uniq
	cmd = exec.Command(bin, "capture-once", "--run-id", realID)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("real attempt: %v\n%s", err, out)
	}

	// 5. Cleanup forged rows (best effort; disposable DB either way).
	_, _ = pool.Exec(ctx, `DELETE FROM `+schema+`.workflow_status WHERE workflow_uuid = $1`, forgedID)
}

// businessCounts fingerprints the POC-written business tables.
func businessCounts(t *testing.T, pool *pgxpool.Pool) string {
	t.Helper()
	ctx := context.Background()
	var sb strings.Builder
	for _, q := range []string{
		`SELECT count(*) FROM poc_ledger_backlink`,
		`SELECT count(*) FROM poc_outbox WHERE state = 'backlinked'`,
		`SELECT count(*) FROM payroll_snapshot`,
		`SELECT count(*) FROM payroll_run`,
		`SELECT count(*) FROM payroll_line`,
		`SELECT count(*) FROM poc_workflow_attempt`,
	} {
		var n int
		if err := pool.QueryRow(ctx, q).Scan(&n); err != nil {
			t.Fatalf("business count %q: %v", q, err)
		}
		sb.WriteString(q + "=" + string(rune('0'+n%10)) + ";")
	}
	return sb.String()
}
