// Package e2e holds the POC experiment harnesses. Every test here requires a
// live POC environment (PostgreSQL 18.6 + disposable database, optionally
// immudb and DBOS) and skips loudly without it. A skipped test is NOT
// evidence — see POC-REPORT.md for the honest status of each experiment.
package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/PithomLabs/a2/poc/payroll-month/dbosflow"
)

func pocPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dsn := os.Getenv("POC_DATABASE_URL")
	if dsn == "" {
		t.Skip("POC_DATABASE_URL unset: requires PG18.6 POC stack (see POC-RUNBOOK.md)")
	}
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatalf("connect POC_DATABASE_URL: %v", err)
	}
	// Bounded close everywhere: plain pool.Close() hangs indefinitely
	// when a DBOS runtime leaks a shared-pool conn past Shutdown
	// (see dbosflow.ClosePoolTimeout); normal close is unaffected.
	t.Cleanup(func() { dbosflow.ClosePoolTimeout(pool, 15*time.Second) })
	return pool
}

// pocAdminPool is the harness-setup role (migration_admin): DDL scaffolding
// (CREATE/DROP/ALTER/EXTENSION) runs here because the app role is never
// granted schema CREATE (prod-like posture). Prove-statements (the DML the
// experiment asserts about) always run on the app pool. Default privileges
// (app=arwd on migration_admin tables) make admin-created scaffolding
// writable by app without extra grants.
func pocAdminPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dsn := os.Getenv("POC_DBOS_MIGRATE_URL")
	if dsn == "" {
		t.Skip("POC_DBOS_MIGRATE_URL unset: harness setup needs the migrate role")
	}
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatalf("connect POC_DBOS_MIGRATE_URL: %v", err)
	}
	t.Cleanup(func() { dbosflow.ClosePoolTimeout(pool, 15*time.Second) })
	return pool
}

func execOK(t *testing.T, pool *pgxpool.Pool, sql string, args ...any) {
	t.Helper()
	if _, err := pool.Exec(context.Background(), sql, args...); err != nil {
		t.Fatalf("exec %q: %v", sql, err)
	}
}

// TestE1_WithoutOverlapsOnGenerated answers research experiment E1: does
// WITHOUT OVERLAPS accept a generated range column (VIRTUAL and STORED)?
func TestE1_WithoutOverlapsOnGenerated(t *testing.T) {
	pool := pocPool(t)
	admin := pocAdminPool(t)
	ctx := context.Background()
	execOK(t, admin, `CREATE EXTENSION IF NOT EXISTS btree_gist`)
	execOK(t, admin, `DROP TABLE IF EXISTS poc_e1_stored, poc_e1_virtual`)

	execOK(t, admin, `CREATE TABLE poc_e1_stored (
		id int, effective_from date NOT NULL, effective_to date,
		eff daterange GENERATED ALWAYS AS (
			daterange(effective_from, COALESCE(effective_to, 'infinity'::date), '[]')
		) STORED)`)
	if _, err := admin.Exec(ctx, `ALTER TABLE poc_e1_stored
		ADD CONSTRAINT poc_e1_stored_wot UNIQUE (id, eff WITHOUT OVERLAPS)`); err != nil {
		t.Fatalf("STORED generated range rejected by WITHOUT OVERLAPS: %v", err)
	}
	// Overlap rejected, adjacency accepted.
	execOK(t, pool, `INSERT INTO poc_e1_stored (id, effective_from, effective_to)
		VALUES (1, '2026-10-01', '2026-10-15')`)
	if _, err := pool.Exec(ctx, `INSERT INTO poc_e1_stored (id, effective_from, effective_to)
		VALUES (1, '2026-10-10', '2026-10-20')`); err == nil {
		t.Fatal("overlapping row accepted, want rejection")
	}
	execOK(t, pool, `INSERT INTO poc_e1_stored (id, effective_from, effective_to)
		VALUES (1, '2026-10-16', '2026-10-31')`)

	// VIRTUAL variant: record the outcome either way (the experiment's answer).
	execOK(t, admin, `CREATE TABLE poc_e1_virtual (
		id int, effective_from date NOT NULL, effective_to date,
		eff daterange GENERATED ALWAYS AS (
			daterange(effective_from, COALESCE(effective_to, 'infinity'::date), '[]')
		) VIRTUAL)`)
	_, verr := admin.Exec(ctx, `ALTER TABLE poc_e1_virtual
		ADD CONSTRAINT poc_e1_virtual_wot UNIQUE (id, eff WITHOUT OVERLAPS)`)
	t.Logf("E1 VIRTUAL verdict: err=%v", verr)
}

// TestE2_PeriodFKJointCover answers E2: do adjacent referenced rows jointly
// cover a referencing period, and are gaps rejected?
func TestE2_PeriodFKJointCover(t *testing.T) {
	// E2 NEGATIVE verdict on PG18.6 (proven live): temporal FOREIGN KEYs
	// (SQL:2011 PERIOD syntax) are not implemented — every spelling is
	// rejected with 42601. Joint coverage therefore cannot be declared;
	// it must be enforced procedurally (gate + claim/proof pattern). This
	// test pins the absence as a loud tripwire: if a future PG accepts
	// the syntax, rewrite this test and Coverage to use real temporal FKs.
	admin := pocAdminPool(t)
	ctx := context.Background()
	execOK(t, admin, `CREATE EXTENSION IF NOT EXISTS btree_gist`)
	execOK(t, admin, `DROP TABLE IF EXISTS poc_e2_child, poc_e2_parent`)
	execOK(t, admin, `CREATE TABLE poc_e2_parent (
		k text, eff daterange NOT NULL,
		UNIQUE (k, eff WITHOUT OVERLAPS))`)
	if _, err := admin.Exec(ctx, `CREATE TABLE poc_e2_child (
		k text, need daterange NOT NULL,
		FOREIGN KEY (k, need PERIOD) REFERENCES poc_e2_parent (k, eff PERIOD))`); err == nil {
		t.Fatal("PERIOD FK accepted: engine grew temporal FKs — rewrite E2/Coverage to use them")
	} else if !strings.Contains(err.Error(), "42601") {
		t.Fatalf("want syntax error 42601, got: %v", err)
	}
	t.Logf("E2 verdict: no temporal FK on PG18.6 (42601); coverage enforced procedurally")
}

// TestOutbox_IdempotentEnqueue proves the outbox half of the dual-write
// TestOutbox_Authority pins the outbox/backlink authority posture
// (prompt8 remediation of PG-5/IM-12): the app role can neither forge
// rows nor skip states nor mutate evidence. Producer idempotency
// (enqueue-twice converges) is proven over the real validated run in
// the Phase E integration, not here.
func TestOutbox_Authority(t *testing.T) {
	pool := pocPool(t)
	admin := pocAdminPool(t)
	ctx := context.Background()
	// The authority overlay must be live: re-applying outbox.sql here
	// (as this test once did) would RE-GRANT app INSERT and silently
	// undo the boundary under test — hence this loud precondition.
	var states string
	if err := pool.QueryRow(ctx,
		`SELECT pg_get_constraintdef(oid) FROM pg_constraint
		  WHERE conname = 'poc_outbox_state_check'`).Scan(&states); err != nil ||
		!strings.Contains(states, "reconciling") {
		t.Fatalf("authority overlay not applied (runbook step 3.6): %v %q", err, states)
	}
	const op = "11111111-1111-1111-1111-111111111111"
	execOK(t, admin, `DELETE FROM poc_outbox WHERE op_uuid = '`+op+`'`)

	// 1. Direct app INSERT rejected (enqueue is DEFINER-only).
	if _, err := pool.Exec(ctx, `INSERT INTO poc_outbox (op_uuid, employer_id, action, payload_hash)
		VALUES ('`+op+`', '22222222-2222-2222-2222-222222222222', 'capture', decode('ab','hex'))`); err == nil {
		t.Fatal("app INSERT on poc_outbox accepted, want permission denial")
	} else if !isDenied(err) {
		t.Fatalf("want permission denial (42501), got: %v", err)
	}
	// 2. Direct backlink INSERT rejected (recorder is DEFINER-only).
	if _, err := pool.Exec(ctx, `INSERT INTO poc_ledger_backlink (pg_table, pg_id, event_uuid, immu_tx, alh)
		VALUES ('payroll_run', 'x', '`+op+`', 1, decode('ab','hex'))`); err == nil {
		t.Fatal("app INSERT on poc_ledger_backlink accepted, want permission denial")
	} else if !isDenied(err) {
		t.Fatalf("want permission denial (42501), got: %v", err)
	}
	// 3-4. State skip + payload mutation rejected by the guard.
	seedOutbox(t, admin, op, "auth-probe")
	if _, err := pool.Exec(ctx, `UPDATE poc_outbox SET state = 'backlinked'
		WHERE op_uuid = '`+op+`'`); err == nil {
		t.Fatal("pending->backlinked skip accepted, want P0001 guard rejection")
	} else if !strings.Contains(err.Error(), "P0001") {
		t.Fatalf("want P0001 guard rejection, got: %v", err)
	}
	if _, err := pool.Exec(ctx, `UPDATE poc_outbox SET payload = '{}'
		WHERE op_uuid = '`+op+`'`); err == nil {
		t.Fatal("payload mutation accepted, want P0001 guard rejection")
	} else if !strings.Contains(err.Error(), "P0001") {
		t.Fatalf("want P0001 guard rejection, got: %v", err)
	}
	// 5. App DELETE rejected; owner DELETE succeeds (Exp3A contrast:
	// grants stop app, only the trigger stops the owner — and the
	// backlink trigger stops even the owner, see TestBacklink_AppendOnly).
	if _, err := pool.Exec(ctx, `DELETE FROM poc_outbox WHERE op_uuid = '`+op+`'`); err == nil {
		t.Fatal("app DELETE on poc_outbox accepted, want permission denial")
	} else if !isDenied(err) {
		t.Fatalf("want permission denial (42501), got: %v", err)
	}
	// 6. IM-10 closure (§5.6): proven appends can never be discarded.
	// pending->appended WITHOUT a tx is rejected (nothing proven).
	if _, err := pool.Exec(ctx, `UPDATE poc_outbox SET state = 'appended'
		WHERE op_uuid = '`+op+`'`); err == nil {
		t.Fatal("appended-without-tx accepted, want P0001")
	} else if !strings.Contains(err.Error(), "P0001") {
		t.Fatalf("want P0001, got: %v", err)
	}
	// pending->failed is allowed (append never proven: retryable);
	// failed->pending resets it (operator re-drive).
	execOK(t, pool, `UPDATE poc_outbox SET state = 'failed' WHERE op_uuid = '`+op+`'`)
	execOK(t, pool, `UPDATE poc_outbox SET state = 'pending' WHERE op_uuid = '`+op+`'`)
	// Owner anchors a (test-only) tx: pending->appended allowed.
	execOK(t, admin, `UPDATE poc_outbox SET state = 'appended', immu_tx = 999999
		WHERE op_uuid = '`+op+`'`)
	// Proven append -> failed/poison rejected, even for the owner:
	// retry accounting cannot orphan attested evidence.
	for _, st := range []string{"failed", "poison"} {
		if _, err := pool.Exec(ctx, `UPDATE poc_outbox SET state = '`+st+`'
			WHERE op_uuid = '`+op+`'`); err == nil {
			t.Fatalf("app appended->%s accepted, want P0001", st)
		} else if !strings.Contains(err.Error(), "P0001") {
			t.Fatalf("want P0001, got: %v", err)
		}
		if _, err := admin.Exec(ctx, `UPDATE poc_outbox SET state = '`+st+`'
			WHERE op_uuid = '`+op+`'`); err == nil {
			t.Fatalf("owner appended->%s accepted, want P0001", st)
		} else if !strings.Contains(err.Error(), "P0001") {
			t.Fatalf("want P0001, got: %v", err)
		}
	}
	// appended->reconciling allowed; reconciling->failed rejected;
	// reconciling->reconciled allowed (stays actionable, then rests).
	execOK(t, admin, `UPDATE poc_outbox SET state = 'reconciling' WHERE op_uuid = '`+op+`'`)
	if _, err := pool.Exec(ctx, `UPDATE poc_outbox SET state = 'failed'
		WHERE op_uuid = '`+op+`'`); err == nil {
		t.Fatal("reconciling->failed accepted, want P0001")
	} else if !strings.Contains(err.Error(), "P0001") {
		t.Fatalf("want P0001, got: %v", err)
	}
	execOK(t, admin, `UPDATE poc_outbox SET state = 'reconciled' WHERE op_uuid = '`+op+`'`)
	// 7. Construction invariant: INSERTs with a mismatched state/tx
	// pair are rejected (probed as owner; app INSERT is revoked).
	const bad1 = "11111111-1111-1111-1111-111111111112"
	const bad2 = "11111111-1111-1111-1111-111111111113"
	for _, ins := range []string{
		`INSERT INTO poc_outbox (op_uuid, employer_id, action, payload_hash, state, immu_tx)
			VALUES ('` + bad1 + `', '22222222-2222-2222-2222-222222222222',
				'capture', decode('ab','hex'), 'pending', 7)`,
		`INSERT INTO poc_outbox (op_uuid, employer_id, action, payload_hash, state)
			VALUES ('` + bad2 + `', '22222222-2222-2222-2222-222222222222',
				'capture', decode('ab','hex'), 'appended')`,
	} {
		if _, err := admin.Exec(ctx, ins); err == nil {
			t.Fatalf("mismatched INSERT accepted, want P0001: %s", ins)
		} else if !strings.Contains(err.Error(), "P0001") {
			t.Fatalf("want P0001, got: %v", err)
		}
	}
	execOK(t, admin, `DELETE FROM poc_outbox WHERE op_uuid = '`+op+`'`)
}

// TestCoverage_BlockedByConstraint encodes the rule-pack coverage experiment:
// a pack that does not cover the full period is blocked by a DB constraint,
// not application code.
func TestCoverage_BlockedByConstraint(t *testing.T) {
	// Coverage enforcement, remediated scope (prompt8 Phase B): pack-side
	// overlap is blocked by a DB constraint (WITHOUT OVERLAPS, E1-proven
	// pattern); the need-side (need ⊆ coverage) is verified procedurally
	// by the redesigned poc_verify_claims (Part 4 of the authority
	// overlay), which runs inside authorized transactions only — the app
	// role cannot invoke it directly (Part 2 below). Temporal FKs remain
	// absent (E2 tripwire); joint coverage is a synchronized procedural
	// check, never a declaration.
	pool := pocPool(t)
	admin := pocAdminPool(t)
	ctx := context.Background()
	execOK(t, admin, `CREATE EXTENSION IF NOT EXISTS btree_gist`)
	execOK(t, admin, `DROP TABLE IF EXISTS poc_cov_pack`)
	execOK(t, admin, `CREATE TABLE poc_cov_pack (
		pack_key text, eff daterange NOT NULL,
		UNIQUE (pack_key, eff WITHOUT OVERLAPS))`)
	execOK(t, pool, `INSERT INTO poc_cov_pack VALUES ('SYNTH-PACK', '[2026-10-01,2026-10-15]')`)
	if _, err := pool.Exec(ctx, `INSERT INTO poc_cov_pack VALUES ('SYNTH-PACK', '[2026-10-10,2026-10-31]')`); err == nil {
		t.Fatal("overlapping pack versions accepted, want exclusion violation")
	}
	execOK(t, pool, `INSERT INTO poc_cov_pack VALUES ('SYNTH-PACK', '[2026-10-16,2026-10-31]')`)
	var n int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM poc_cov_pack`).Scan(&n); err != nil || n != 2 {
		t.Fatalf("pack rows = %d, err = %v, want 2", n, err)
	}
	// Part 2: the verifier is fixed AND withheld — direct app calls
	// fail 42501 (authorized transactions call it as the DEFINER
	// owner). Its correctness is proven through the capture paths in
	// the authority battery (authority_test.go), not through this pin.
	if _, err := pool.Exec(ctx, `SELECT * FROM poc_verify_claims(
		'33333333-3333-3333-3333-333333333333', '44444444-4444-4444-4444-444444444444')`); err == nil {
		t.Fatal("poc_verify_claims callable by app, want permission denial")
	} else if !isDenied(err) {
		t.Fatalf("want permission denial (42501), got: %v", err)
	}
}

func isDenied(err error) bool {
	// 42501 = insufficient_privilege.
	return err != nil && strings.Contains(err.Error(), "42501")
}

// TestCrashMatrix executes the specified crash procedure via the kill-matrix
// harnesses (no duplication): (a) inside capture txn pre-commit =
// TestKillMatrix_CaptureBoundary; (b)+(c) PG-commit/append/backlink
// boundaries = TestKillMatrix_OutboxBoundaries; plus the step-window sweep.
func TestCrashMatrix(t *testing.T) {
	t.Run("capture", TestKillMatrix_CaptureBoundary)
	t.Run("outbox", TestKillMatrix_OutboxBoundaries)
	t.Run("stepwindow", TestKillMatrix_StepWindow)
}

// Tamper detection moved to the content-audit battery (auditor_test.go):
// TestBacklink_AppendOnly (prevention), TestContentAudit_GenuinePasses,
// TestContentAudit_TamperBattery, TestContentAudit_ForgedAndMissing,
// TestContentAudit_Corrections (detection), TestGenesis_PinEnforced and
// TestIdentity_Separation (§5.3). The old tx-existence checks they
// replace could not see PG tamper (IM-7).

// TestConcurrency_WriterRacesCapture pins the gate-access posture;
// the live race proof moved to TestAuthority_GateInterleavings
// (authority_test.go), which races REAL writers against the REAL
// capture path. Direct app calls to poc_gate_acquire are revoked:
// writers reach the gate through the DEFINER triggers only.
func TestConcurrency_WriterRacesCapture(t *testing.T) {
	pool := pocPool(t)
	ctx := context.Background()
	if _, err := pool.Exec(ctx, `SELECT public.poc_gate_acquire(gen_random_uuid())`); err == nil {
		t.Fatal("direct app gate call accepted, want permission denial")
	} else if !isDenied(err) {
		t.Fatalf("want permission denial (42501), got: %v", err)
	}
}

// TestReproducibility_RerunIdentical proves cross-attempt identity:
// two DIFFERENT attempt IDs over unchanged inputs adopt the same
// business run with byte-identical input digests (same attempt ID
// twice is Reject-on-reuse — see TestDoubleEnqueue_SingleExecution).
func TestReproducibility_RerunIdentical(t *testing.T) {
	pool := pocPool(t)
	bin := buildPocRun(t)
	tag := fmt.Sprintf("repro-%d", time.Now().UnixNano())
	type captureData struct {
		Data struct {
			RunID      string `json:"capture_run_id"`
			InputsHash string `json:"capture_inputs_hash"`
			Role       string `json:"capture_role"`
		} `json:"data"`
	}
	ids := make([]captureData, 0, 2)
	for _, suffix := range []string{"a", "b"} {
		out := runPocRunStdout(t, bin, "capture-once", "--run-id", tag+"-"+suffix)
		var parsed captureData
		if err := json.Unmarshal([]byte(out), &parsed); err != nil {
			t.Fatalf("parse report: %v\n%s", err, out)
		}
		if parsed.Data.RunID == "" || parsed.Data.InputsHash == "" {
			t.Fatalf("empty capture data:\n%s", out)
		}
		ids = append(ids, parsed)
	}
	if ids[0].Data.RunID != ids[1].Data.RunID {
		t.Fatalf("attempts resolved to different runs: %s vs %s",
			ids[0].Data.RunID, ids[1].Data.RunID)
	}
	if ids[0].Data.InputsHash != ids[1].Data.InputsHash {
		t.Fatalf("input digests differ:\n%s\n%s",
			ids[0].Data.InputsHash, ids[1].Data.InputsHash)
	}
	// Both attempts claimed the same run in the registry.
	var n int
	if err := pool.QueryRow(context.Background(),
		`SELECT count(DISTINCT run_id) FROM poc_workflow_attempt
		  WHERE workflow_id IN ($1, $2)`,
		"capture:"+tag+"-a", "capture:"+tag+"-b").Scan(&n); err != nil || n != 1 {
		t.Fatalf("distinct runs = %d err=%v, want 1", n, err)
	}
}

// TestCorrections_AppendNewEvents proves prompt8 §5.5: a second evidence
// event for the same finalized artifact appends (new event uuid, new
// ledger tx, supersedes lineage) without touching the original
// finalization row. Both events stay independently verifiable.
func TestCorrections_AppendNewEvents(t *testing.T) {
	pool := pocPool(t)
	admin := pocAdminPool(t)
	ctx := context.Background()
	authSetup(t, admin)
	bin := buildPocRun(t)
	benv := map[string]string{"POC_EMPLOYER": authEmployer, "POC_PERIOD": authPeriod}

	// Finalize a battery month (background wait + mid-flight approve).
	cmd := exec.Command(bin, "--approval-timeout", "150s")
	cmd.Env = os.Environ()
	for k, v := range benv {
		cmd.Env = append(cmd.Env, k+"="+v)
	}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	var runID string
	waitFor(t, "battery month persisted", 120*time.Second, func() bool {
		return pool.QueryRow(ctx,
			`SELECT r.id FROM payroll_run r JOIN employer e ON e.id = r.employer_id
			  JOIN poc_run_output o ON o.run_id = r.id
			  WHERE e.code = $1`, authEmployer).Scan(&runID) == nil
	})
	deadline := time.Now().Add(240 * time.Second)
	approved := false
	for time.Now().Before(deadline) {
		approveOut, approveErr := runBin(t, bin, benv, "approve", "--run-id", runID,
			"--decision", "approve", "--approver", "correction-test")
		if approveErr == nil {
			approved = true
		} else if !strings.Contains(approveOut, "no rows in result set") {
			t.Fatalf("approve: %v\n%s", approveErr, approveOut)
		}
		// Else: no waiter yet; approve is convergent, keep re-sending.
		select {
		case err := <-done:
			if err != nil {
				t.Fatalf("battery month: %v", err)
			}
			goto completed
		case <-time.After(20 * time.Second):
		}
	}
	_ = cmd.Process.Kill()
	t.Fatal("battery month did not complete after approval")
completed:
	if !approved {
		t.Fatal("approve never succeeded")
	}
	// Event 1 (the month's own finalize enqueue, op = run id) drains.
	runPocRun(t, bin, "outbox-once", "--op", runID)
	if got := outboxState(t, pool, runID); got != "backlinked" {
		t.Fatalf("event 1 state = %q, want backlinked", got)
	}
	var tx1 uint64
	if err := pool.QueryRow(ctx,
		`SELECT immu_tx FROM poc_ledger_backlink
		  WHERE pg_table = 'payroll_run' AND pg_id = $1 AND event_uuid = $1`,
		runID).Scan(&tx1); err != nil {
		t.Fatalf("event 1 backlink: %v", err)
	}
	// Event 2 (correction): a new event for the SAME artifact, via the
	// authorized enqueue as the app role. Must chain supersedes.
	emp := authEmpID(t, admin)
	ev2 := fmt.Sprintf("aaaaaaaa-bbbb-cccc-dddd-%012x", time.Now().UnixNano()&0xffffffffffff)
	var enq string
	if err := pool.QueryRow(ctx, `SELECT poc_outbox_enqueue($1, $2, $3, $4)`,
		emp, runID, ev2, "correction-test").Scan(&enq); err != nil || enq != "enqueued" {
		t.Fatalf("correction enqueue = %q err=%v, want enqueued", enq, err)
	}
	var sup string
	if err := pool.QueryRow(ctx,
		`SELECT payload->>'supersedes' FROM poc_outbox WHERE op_uuid = $1`, ev2).Scan(&sup); err != nil || sup != runID {
		t.Fatalf("event 2 supersedes = %q err=%v, want event 1 %s", sup, err, runID)
	}
	runPocRun(t, bin, "outbox-once", "--op", ev2)
	if got := outboxState(t, pool, ev2); got != "backlinked" {
		t.Fatalf("event 2 state = %q, want backlinked", got)
	}
	// Lineage: two events, two txs, original untouched.
	var n, distinct int
	if err := pool.QueryRow(ctx,
		`SELECT count(*), count(DISTINCT immu_tx) FROM poc_ledger_backlink
		  WHERE pg_table = 'payroll_run' AND pg_id = $1`, runID).Scan(&n, &distinct); err != nil || n != 2 || distinct != 2 {
		t.Fatalf("backlinks = %d distinct txs = %d err=%v, want 2/2", n, distinct, err)
	}
	var tx1again uint64
	if err := pool.QueryRow(ctx,
		`SELECT immu_tx FROM poc_ledger_backlink
		  WHERE pg_table = 'payroll_run' AND pg_id = $1 AND event_uuid = $1`,
		runID).Scan(&tx1again); err != nil || tx1again != tx1 {
		t.Fatalf("event 1 tx moved: %d -> %d err=%v", tx1, tx1again, err)
	}
	authCleanup(t, admin)
}

var _ = pgx.ErrNoRows
