package e2e

// Kill-matrix harness (EXPERIMENTS.md B1–B5): spawn poc-run subcommands with
// POC_FAILPOINT set, SIGKILL inside the deterministic window, restart clean,
// and assert persisted state in PG + ledger. Skips loudly without the stack.
// B1–B3 are deterministic (failpoint sleep); the DBOS step-checkpoint window
// (inside DBOS runtime, not hookable from user code) is covered by timed
// kills with bounded attempts — see TestKillMatrix_StepWindow.

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/PithomLabs/a2/poc/payroll-month/ledger"
)

func buildPocRun(t *testing.T) string {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "poc-run")
	cmd := exec.Command("go", "build", "-o", bin, "./cmd/poc-run")
	cmd.Dir = ".."
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build poc-run: %v\n%s", err, out)
	}
	return bin
}

func immudbEnv(t *testing.T) (host string, port int, user, pass, db string) {
	t.Helper()
	addr := os.Getenv("POC_IMMUDB_ADDR")
	if addr == "" {
		t.Skip("POC_IMMUDB_ADDR unset: kill matrix needs the POC stack")
	}
	host, port = "localhost", 3322
	if h, p, ok := strings.Cut(addr, ":"); ok {
		host = h
		fmt.Sscanf(p, "%d", &port)
	} else {
		host = addr
	}
	return host, port, getenvOr("POC_IMMUDB_USER", "poc_writer"),
		os.Getenv("POC_IMMUDB_PASSWORD"), getenvOr("POC_IMMUDB_DB", "defaultdb")
}

func getenvOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

// seedOutbox inserts a harness outbox row as the OWNER (admin pool).
// Direct app INSERT is revoked (authority overlay Part 2); the product
// producer is poc_outbox_enqueue (Phase C wires it). Payload uses the
// canonical enqueue shape with fiction digests (mechanism tests only;
// content provenance arrives with the real producer in Phase C/D).
func seedOutbox(t *testing.T, admin *pgxpool.Pool, opID, pgID string) {
	t.Helper()
	seedOutboxAt(t, admin, opID, pgID, "now()")
}

// seedOutboxFull seeds a harness row with explicit action/table (the
// forgery battery needs payroll.finalized-shaped rows the ledger
// never saw).
func seedOutboxFull(t *testing.T, admin *pgxpool.Pool, opID, pgID, action, pgTable string) {
	t.Helper()
	ctx := context.Background()
	payload := `{"pg_table":"` + pgTable + `","pg_id":"` + pgID + `","event_uuid":"` + opID + `",` +
		`"employer_id":"aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",` +
		`"input_snapshot_digest":"fiction-input","output_digest":"fiction-output",` +
		`"rule_pins":[{"pack":"SYNTH-OT-001","version":"v1","digest":"d"}],` +
		`"calc_version":"poc-synth/v1",` +
		`"approval":{"approver":"poc-harness","decided_at":"2026-10-31T00:00:00Z"},"supersedes":""}`
	_, err := admin.Exec(ctx, `INSERT INTO poc_outbox (op_uuid, employer_id, action, payload_hash, payload)
		VALUES ($1, 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', $3, decode('ab','hex'), $2)
		ON CONFLICT (op_uuid) DO UPDATE SET state='pending', payload_hash=EXCLUDED.payload_hash,
			payload=EXCLUDED.payload, attempts=0, updated_at=now()`, opID, payload, action)
	if err != nil {
		t.Fatalf("seed outbox: %v", err)
	}
}

// seedOutboxAt is seedOutbox with an explicit created_at expression (the
// kill matrix backdates a decoy row to prove --op pinning ignores older
// pending rows that oldest-first would serve first).
func seedOutboxAt(t *testing.T, admin *pgxpool.Pool, opID, pgID, createdAtExpr string) {
	t.Helper()
	ctx := context.Background()
	// NOTE: backlink rows can NEVER be cleaned (append-only trigger binds
	// every role including admins — proven live: P0001 on owner DELETE),
	// so each run seeds globally-unique op/pg ids (see caller).
	payload := `{"pg_table":"payroll_snapshot","pg_id":"` + pgID + `","event_uuid":"` + opID + `",` +
		`"employer_id":"aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",` +
		`"input_snapshot_digest":"fiction-input","output_digest":"fiction-output",` +
		`"rule_pins":[{"pack":"SYNTH-OT-001","version":"v1","digest":"d"}],` +
		`"calc_version":"poc-synth/v1",` +
		`"approval":{"approver":"poc-harness","decided_at":"2026-10-31T00:00:00Z"},"supersedes":""}`
	_, err := admin.Exec(ctx, `INSERT INTO poc_outbox (op_uuid, employer_id, action, payload_hash, payload, created_at)
		VALUES ($1, 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'snapshot', decode('ab','hex'), $2, `+createdAtExpr+`)
		ON CONFLICT (op_uuid) DO UPDATE SET state='pending', payload_hash=EXCLUDED.payload_hash,
			payload=EXCLUDED.payload, attempts=0, updated_at=now()`, opID, payload)
	if err != nil {
		t.Fatalf("seed outbox: %v", err)
	}
}

// spawnFailpoint starts bin with POC_FAILPOINT=name and waits for the armed
// marker (fail) or natural exit (ok=false). Caller must Kill on marker.
func spawnFailpoint(t *testing.T, bin string, name string, args ...string) (cmd *exec.Cmd, killed func()) {
	t.Helper()
	cmd = exec.Command(bin, args...)
	cmd.Env = append(os.Environ(), "POC_FAILPOINT="+name)
	stderr, err := cmd.StderrPipe()
	if err != nil {
		t.Fatalf("stderr pipe: %v", err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatalf("start: %v", err)
	}
	deadline := time.Now().Add(3 * time.Minute)
	sc := bufio.NewScanner(stderr)
	for sc.Scan() {
		if strings.Contains(sc.Text(), "failpoint armed: "+name) {
			return cmd, func() {
				_ = cmd.Process.Kill()
				_, _ = io.Copy(io.Discard, stderr)
				_ = cmd.Wait()
			}
		}
		if time.Now().After(deadline) {
			_ = cmd.Process.Kill()
			t.Fatalf("failpoint %s: marker not seen in 3m", name)
		}
	}
	t.Fatalf("failpoint %s: process exited before arming", name)
	return nil, nil
}

func runCaptureOnce(t *testing.T, bin, runID string) {
	t.Helper()
	runPocRun(t, bin, "capture-once", "--run-id", runID)
}

// runPocRun executes the built poc-run with args, failing loudly; returns
// combined output for content assertions.
func runPocRun(t *testing.T, bin string, args ...string) string {
	t.Helper()
	cmd := exec.Command(bin, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("poc-run %v: %v\n%s", args, err, out)
	}
	return string(out)
}

// runPocRunStdout is runPocRun for byte comparisons: DBOS INFO logs on
// stderr carry timestamps, so only stdout (the JSON report) is compared.
func runPocRunStdout(t *testing.T, bin string, args ...string) string {
	t.Helper()
	cmd := exec.Command(bin, args...)
	out, err := cmd.Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			t.Fatalf("poc-run %v: %v\n%s", args, err, ee.Stderr)
		}
		t.Fatalf("poc-run %v: %v", args, err)
	}
	return string(out)
}

func outboxState(t *testing.T, pool *pgxpool.Pool, opID string) string {
	t.Helper()
	var s string
	err := pool.QueryRow(context.Background(),
		`SELECT state FROM poc_outbox WHERE op_uuid = $1`, opID).Scan(&s)
	if err != nil {
		t.Fatalf("outbox state: %v", err)
	}
	return s
}

// outboxAttempts proves a claim landed on the intended row: with --op
// pinning, the failpoint run must bump exactly this row to 1 (a kill on
// any other row — or no claim at all — fails here, not silently).
func outboxAttempts(t *testing.T, pool *pgxpool.Pool, opID string) int {
	t.Helper()
	var n int
	err := pool.QueryRow(context.Background(),
		`SELECT attempts FROM poc_outbox WHERE op_uuid = $1`, opID).Scan(&n)
	if err != nil {
		t.Fatalf("outbox attempts: %v", err)
	}
	return n
}

func TestKillMatrix_OutboxBoundaries(t *testing.T) {
	pool := pocPool(t)
	admin := pocAdminPool(t)
	host, port, user, pass, db := immudbEnv(t)
	// Hermetic writer state: spawned binaries inherit test env, so pin
	// the file-cache dir to a temp dir (never the repo, never shared).
	t.Setenv("POC_WRITER_STATE_DIR", t.TempDir())
	bin := buildPocRun(t)
	// Drain stale kill-test rows (bbbb-*/dddd-* namespaces): leftovers
	// from killed/failed runs must not pollute later assertions.
	// Backlinks/markers from old runs are keyed by unique op ids and
	// never interfere.
	if _, err := admin.Exec(context.Background(), `DELETE FROM poc_outbox
		WHERE op_uuid::text LIKE 'bbbbbbbb-%' OR op_uuid::text LIKE 'dddddddd-%'`); err != nil {
		t.Fatalf("drain stale outbox rows: %v", err)
	}
	// Decoy: an OLDER pending row that oldest-first would serve ahead
	// of every seeded row. The pinned runs below must leave it
	// untouched (pending, attempts 0); any claim of it is a pinning
	// regression. Removed at the end so the tamper battery's
	// stale-guard (and later runs) see a clean queue.
	decoyID := fmt.Sprintf("dddddddd-dddd-dddd-dddd-%012x", time.Now().UnixNano()&0xffffffffffff)
	seedOutboxAt(t, admin, decoyID, "decoy", "now() - interval '2 hours'")
	defer func() {
		if _, err := admin.Exec(context.Background(),
			`DELETE FROM poc_outbox WHERE op_uuid = $1`, decoyID); err != nil {
			t.Errorf("decoy cleanup: %v", err)
		}
	}()
	for i, fp := range []string{"after-pg-commit-before-append", "after-append-before-backlink"} {
		// Unique per RUN (nanos + index, 48 bits): backlink rows are
		// append-only and can never be cleaned for reruns.
		uniq := fmt.Sprintf("%012x", (time.Now().UnixNano()+int64(i))&0xffffffffffff)
		opID := "bbbbbbbb-bbbb-bbbb-bbbb-" + uniq
		seedOutbox(t, admin, opID, "snap-"+uniq[:8])
		// Pinned drain: stale pending rows from other runs must not
		// steal the claim (oldest-first would serve them first).
		_, kill := spawnFailpoint(t, bin, fp, "outbox-once", "--op", opID)
		kill()
		if got := outboxState(t, pool, opID); got != "pending" {
			t.Fatalf("%s: after kill, state = %q, want pending", fp, got)
		}
		if got := outboxAttempts(t, pool, opID); got != 1 {
			t.Fatalf("%s: after kill, attempts = %d, want 1 (kill missed the seeded row)", fp, got)
		}
		var n int
		if err := pool.QueryRow(context.Background(),
			`SELECT count(*) FROM poc_ledger_backlink WHERE event_uuid = $1`, opID).Scan(&n); err != nil {
			t.Fatal(err)
		}
		if n != 0 {
			t.Fatalf("%s: backlink present after kill", fp)
		}
		// Restart clean: must converge exactly once.
		cmd := exec.Command(bin, "outbox-once", "--op", opID)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("%s: restart: %v\n%s", fp, err, out)
		}
		if got := outboxState(t, pool, opID); got != "backlinked" {
			t.Fatalf("%s: after restart, state = %q, want backlinked", fp, got)
		}
		if got := outboxAttempts(t, pool, opID); got != 2 {
			t.Fatalf("%s: after restart, attempts = %d, want 2 (kill claim + restart claim)", fp, got)
		}
		// Ledger side: marker readable, origin tx matches the backlink row.
		ctx := context.Background()
		app, err := ledger.Dial(ctx, host, port, user, pass, db)
		if err != nil {
			t.Fatalf("ledger dial: %v", err)
		}
		marker, tx, err := app.ReadMarker(ctx, ledger.IdemKey("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", "snapshot", opID))
		_ = app.Close(ctx)
		if err != nil {
			t.Fatalf("%s: read marker: %v", fp, err)
		}
		if !strings.HasPrefix(marker, opID+":") {
			t.Fatalf("%s: marker %q does not bind op", fp, marker)
		}
		var linked uint64
		if err := pool.QueryRow(ctx, `SELECT immu_tx FROM poc_ledger_backlink WHERE event_uuid = $1`, opID).Scan(&linked); err != nil {
			t.Fatal(err)
		}
		if linked != tx {
			t.Fatalf("%s: backlink tx %d != marker origin %d", fp, linked, tx)
		}
	}
	// Pinning proof: the older decoy row was never claimed.
	if got := outboxState(t, pool, decoyID); got != "pending" {
		t.Fatalf("decoy state = %q, want pending (pinning regression)", got)
	}
	if got := outboxAttempts(t, pool, decoyID); got != 0 {
		t.Fatalf("decoy attempts = %d, want 0 (pinning regression)", got)
	}
}

// outboxTx returns the anchored tx (0 when none). Reconcile-battery
// rows must carry a real anchor before the back-link is repaired.
func outboxTx(t *testing.T, pool *pgxpool.Pool, opID string) uint64 {
	t.Helper()
	var tx *uint64
	err := pool.QueryRow(context.Background(),
		`SELECT immu_tx FROM poc_outbox WHERE op_uuid = $1`, opID).Scan(&tx)
	if err != nil {
		t.Fatalf("outbox tx: %v", err)
	}
	if tx == nil {
		return 0
	}
	return *tx
}

// TestKillMatrix_AppendedBoundary extends the matrix past B3: a kill
// AFTER mark-appended commits but BEFORE the back-link txn leaves an
// appended row with a proven anchor and no back-link. Re-driving must
// re-verify the anchor (verified reads only) and complete to
// backlinked — never re-append, never poison.
func TestKillMatrix_AppendedBoundary(t *testing.T) {
	pool := pocPool(t)
	admin := pocAdminPool(t)
	host, port, user, pass, db := immudbEnv(t)
	t.Setenv("POC_WRITER_STATE_DIR", t.TempDir())
	bin := buildPocRun(t)
	if _, err := admin.Exec(context.Background(), `DELETE FROM poc_outbox WHERE op_uuid::text LIKE 'eeeeeeee-%'`); err != nil {
		t.Fatalf("drain stale outbox rows: %v", err)
	}
	uniq := fmt.Sprintf("%012x", time.Now().UnixNano()&0xffffffffffff)
	opID := "eeeeeeee-eeee-eeee-eeee-" + uniq
	pgID := "snap-" + uniq[:8]
	seedOutbox(t, admin, opID, pgID)
	_, kill := spawnFailpoint(t, bin, "after-appended-before-backlink", "outbox-once", "--op", opID)
	kill()
	if got := outboxState(t, pool, opID); got != "appended" {
		t.Fatalf("after kill, state = %q, want appended", got)
	}
	anchor := outboxTx(t, pool, opID)
	if anchor == 0 {
		t.Fatal("after kill, no anchored tx (B4 window missed)")
	}
	if got := outboxAttempts(t, pool, opID); got != 1 {
		t.Fatalf("after kill, attempts = %d, want 1", got)
	}
	var n int
	if err := pool.QueryRow(context.Background(),
		`SELECT count(*) FROM poc_ledger_backlink WHERE event_uuid = $1`, opID).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatal("backlink present after kill")
	}
	// Restart clean: reconciles the proven anchor to backlinked.
	cmd := exec.Command(bin, "outbox-once", "--op", opID)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("restart: %v\n%s", err, out)
	}
	if got := outboxState(t, pool, opID); got != "backlinked" {
		t.Fatalf("after restart, state = %q, want backlinked", got)
	}
	if got := outboxAttempts(t, pool, opID); got != 2 {
		t.Fatalf("after restart, attempts = %d, want 2", got)
	}
	ctx := context.Background()
	app, err := ledger.Dial(ctx, host, port, user, pass, db)
	if err != nil {
		t.Fatalf("ledger dial: %v", err)
	}
	marker, tx, err := app.ReadMarker(ctx, ledger.IdemKey("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", "snapshot", opID))
	_ = app.Close(ctx)
	if err != nil {
		t.Fatalf("read marker: %v", err)
	}
	if !strings.HasPrefix(marker, opID+":") {
		t.Fatalf("marker %q does not bind op", marker)
	}
	var linked uint64
	if err := pool.QueryRow(ctx, `SELECT immu_tx FROM poc_ledger_backlink WHERE event_uuid = $1`, opID).Scan(&linked); err != nil {
		t.Fatal(err)
	}
	if linked != tx || linked != anchor {
		t.Fatalf("backlink tx %d != marker origin %d / anchor %d", linked, tx, anchor)
	}
}

// TestReconcile_RecoversBacklinkMissing proves prompt8 §5.6: a row in
// reconciling (proven append, back-link outstanding) re-verifies its
// anchor and completes to reconciled; a duplicate drive afterwards
// converges (exit 0, single back-link, no state churn).
func TestReconcile_RecoversBacklinkMissing(t *testing.T) {
	pool := pocPool(t)
	admin := pocAdminPool(t)
	t.Setenv("POC_WRITER_STATE_DIR", t.TempDir())
	bin := buildPocRun(t)
	ctx := context.Background()
	if _, err := admin.Exec(ctx, `DELETE FROM poc_outbox WHERE op_uuid::text LIKE 'eeeeeeee-%'`); err != nil {
		t.Fatalf("drain stale outbox rows: %v", err)
	}
	uniq := fmt.Sprintf("%012x", time.Now().UnixNano()&0xffffffffffff)
	opID := "eeeeeeee-eeee-eeee-eeee-" + uniq
	seedOutbox(t, admin, opID, "snap-"+uniq[:8])
	// Reach appended honestly (B4 kill), then record the back-link
	// failure the way the worker does (appended -> reconciling).
	_, kill := spawnFailpoint(t, bin, "after-appended-before-backlink", "outbox-once", "--op", opID)
	kill()
	if got := outboxState(t, pool, opID); got != "appended" {
		t.Fatalf("after kill, state = %q, want appended", got)
	}
	execOK(t, admin, `UPDATE poc_outbox SET state = 'reconciling' WHERE op_uuid = '`+opID+`'`)
	// Re-drive: verified reconcile repairs the link (-> reconciled).
	cmd := exec.Command(bin, "outbox-once", "--op", opID)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("reconcile drive: %v\n%s", err, out)
	}
	if got := outboxState(t, pool, opID); got != "reconciled" {
		t.Fatalf("after reconcile, state = %q, want reconciled", got)
	}
	var linked uint64
	if err := pool.QueryRow(ctx, `SELECT immu_tx FROM poc_ledger_backlink WHERE event_uuid = $1`, opID).Scan(&linked); err != nil {
		t.Fatalf("backlink missing after reconcile: %v", err)
	}
	if linked != outboxTx(t, pool, opID) {
		t.Fatalf("backlink tx %d != anchored %d", linked, outboxTx(t, pool, opID))
	}
	// Duplicate drive: ordinary convergence (exit 0, no churn).
	cmd = exec.Command(bin, "outbox-once", "--op", opID)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("duplicate drive: %v\n%s", err, out)
	}
	if got := outboxState(t, pool, opID); got != "reconciled" {
		t.Fatalf("after duplicate, state = %q, want reconciled", got)
	}
	var n int
	if err := pool.QueryRow(ctx,
		`SELECT count(*) FROM poc_ledger_backlink WHERE event_uuid = $1`, opID).Scan(&n); err != nil || n != 1 {
		t.Fatalf("backlinks = %d err=%v, want 1", n, err)
	}
}

func TestKillMatrix_CaptureBoundary(t *testing.T) {
	pool := pocPool(t)
	admin := pocAdminPool(t)
	// Battery employer, recycled first: the B1 window only exists on
	// the claim path, and a persisted validated run (adopt path) would
	// skip the failpoint entirely. MERIDIAN stays reserved for Phase E.
	authSetup(t, admin)
	t.Setenv("POC_EMPLOYER", authEmployer)
	t.Setenv("POC_PERIOD", authPeriod)
	bin := buildPocRun(t)
	// Unique per ATTEMPT: attempt IDs are Reject-on-reuse (DB-4), so a
	// killed attempt is retried under a NEW ID with no table surgery.
	// Both IDs converge on the single business run (below).
	runID := fmt.Sprintf("kill-capture-%d", time.Now().UnixNano())
	_, kill := spawnFailpoint(t, bin, "pre-commit-capture", "capture-once", "--run-id", runID)
	kill()
	// Pre-commit kill: the capture txn rolled back; no checkpoint row may exist.
	ctx := context.Background()
	var n int
	if err := pool.QueryRow(ctx,
		`SELECT count(*) FROM dbos.workflow_status WHERE workflow_uuid = $1 AND status = 'SUCCESS'`,
		"capture:"+runID).Scan(&n); err != nil {
		t.Fatalf("workflow status probe: %v", err)
	}
	if n != 0 {
		t.Fatalf("pre-commit kill left SUCCESS checkpoint")
	}
	// Restart clean under a NEW attempt ID: succeeds (claims the run,
	// or adopts it if Launch recovery already resumed the killed
	// attempt — that race is designed DBOS behavior, not a test
	// outcome; the invariant is single-run convergence).
	retryID := fmt.Sprintf("kill-capture-retry-%d", time.Now().UnixNano())
	cmd := exec.Command(bin, "capture-once", "--run-id", retryID)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("restart: %v\n%s", err, out)
	}
	if err := pool.QueryRow(ctx,
		`SELECT count(*) FROM dbos.workflow_status WHERE workflow_uuid = $1 AND status = 'SUCCESS'`,
		"capture:"+retryID).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("after restart, SUCCESS rows = %d, want 1", n)
	}
	// Business effect: exactly one run, and every attempt row (the
	// retry's, plus the killed attempt's if Launch recovery resumed
	// it) references that same run — no partial, no duplicate.
	var runs int
	if err := pool.QueryRow(ctx,
		`SELECT count(*) FROM payroll_run r JOIN employer e ON e.id = r.employer_id
		  WHERE e.code = $1`, authEmployer).Scan(&runs); err != nil || runs != 1 {
		t.Fatalf("business runs = %d err=%v, want 1", runs, err)
	}
	var distinct int
	if err := pool.QueryRow(ctx,
		`SELECT count(DISTINCT run_id) FROM poc_workflow_attempt
		  WHERE workflow_id IN ($1, $2)`,
		"capture:"+runID, "capture:"+retryID).Scan(&distinct); err != nil || distinct != 1 {
		t.Fatalf("distinct runs across killed+retry = %d err=%v, want 1", distinct, err)
	}
	if err := pool.QueryRow(ctx,
		`SELECT count(*) FROM poc_workflow_attempt WHERE workflow_id = $1`,
		"capture:"+retryID).Scan(&n); err != nil || n != 1 {
		t.Fatalf("retry attempt rows = %d err=%v, want 1", n, err)
	}
	authCleanup(t, admin)
}

// TestKillMatrix_StepWindow covers the effect→checkpoint window inside DBOS
// runtime (not hookable from user code) with bounded timed kills: whether or
// not any single kill lands in the window, completion always converges to a
// single outcome row.
func TestKillMatrix_StepWindow(t *testing.T) {
	pool := pocPool(t)
	bin := buildPocRun(t)
	// Unique per ATTEMPT (Reject-on-reuse): 5 killed attempts + 1 clean
	// final. Every survivor adopts the single business run; business
	// state converges regardless of where kills landed.
	ctx := context.Background()
	tag := fmt.Sprintf("kill-window-%d", time.Now().UnixNano())
	ids := make([]string, 0, 6)
	for i := 0; i < 5; i++ {
		id := fmt.Sprintf("%s-%d", tag, i)
		ids = append(ids, "capture:"+id)
		cmd := exec.Command(bin, "capture-once", "--run-id", id)
		if err := cmd.Start(); err != nil {
			t.Fatal(err)
		}
		time.Sleep(time.Duration(50+i*50) * time.Millisecond)
		_ = cmd.Process.Kill() // may already have exited; that is fine
		_ = cmd.Wait()
	}
	finalID := tag + "-final"
	ids = append(ids, "capture:"+finalID)
	cmd := exec.Command(bin, "capture-once", "--run-id", finalID)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("final run: %v\n%s", err, out)
	}
	var n int
	if err := pool.QueryRow(ctx,
		`SELECT count(*) FROM dbos.workflow_status WHERE workflow_uuid = $1 AND status = 'SUCCESS'`,
		"capture:"+finalID).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("final SUCCESS rows = %d, want exactly 1", n)
	}
	// Business convergence: every attempt that claimed anything
	// claimed the SAME single run (no duplicates, no partial runs).
	var runs int
	if err := pool.QueryRow(ctx,
		`SELECT count(DISTINCT run_id) FROM poc_workflow_attempt
		  WHERE workflow_id = ANY($1)`, ids).Scan(&runs); err != nil || runs != 1 {
		t.Fatalf("distinct runs across attempts = %d err=%v, want 1", runs, err)
	}
}
