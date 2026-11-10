package e2e

// DBOS behavior battery (prompt8 §4.2–§4.4): input-version conflict,
// ERROR retry without surgery, and crash-during-approval recovery.
// All assertions target persisted business state; workflow statuses
// are supporting evidence only.

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// runBin runs a poc-run subcommand with extra env, returning combined output.
func runBin(t *testing.T, bin string, env map[string]string, args ...string) (string, error) {
	t.Helper()
	cmd := exec.Command(bin, args...)
	cmd.Env = os.Environ()
	for k, v := range env {
		cmd.Env = append(cmd.Env, k+"="+v)
	}
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// waitFor polls cond until true or the deadline (fail-closed on timeout).
func waitFor(t *testing.T, what string, timeout time.Duration, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(2 * time.Second)
	}
	t.Fatalf("timed out waiting for %s", what)
}

// TestInputConflict_ChangedInputs proves the DB-3 rule on the battery
// employer: a new attempt over CHANGED inputs fails with an explicit
// input-version conflict (never a silent stale replay); restoring the
// inputs restores adoptability.
func TestInputConflict_ChangedInputs(t *testing.T) {
	pool := pocPool(t)
	admin := pocAdminPool(t)
	ctx := context.Background()
	authSetup(t, admin)
	bin := buildPocRun(t)
	benv := map[string]string{"POC_EMPLOYER": authEmployer, "POC_PERIOD": authPeriod}

	if out, err := runBin(t, bin, benv, "capture-once", "--run-id",
		fmt.Sprintf("conflict-base-%d", time.Now().UnixNano())); err != nil {
		t.Fatalf("baseline capture: %v\n%s", err, out)
	}
	// Change the inputs: a new overlapping comp line (authorized op).
	execOK(t, admin, `SELECT add_compensation(
		(SELECT id FROM employer WHERE code = '`+authEmployer+`'),
		'aaaaaaaa-0000-4000-8000-000000000011',
		'aaaaaaaa-0000-4000-8000-000000000021',
		'allowance', 'CONFLICT-PROBE', 1.00, '2026-10-01', '2026-10-31',
		'conflict-probe', 'input-drift probe')`)
	if out, err := runBin(t, bin, benv, "capture-once", "--run-id",
		fmt.Sprintf("conflict-drift-%d", time.Now().UnixNano())); err == nil {
		t.Fatalf("capture over changed inputs accepted:\n%s", out)
	} else if !strings.Contains(string(out), "input-version conflict") {
		t.Fatalf("want input-version conflict, got:\n%s\n%v", out, err)
	}
	// Restore: owner removes the probe line (DELETE has no freeze
	// trigger; the gate trigger only serializes).
	execOK(t, admin, `DELETE FROM compensation_line WHERE code = 'CONFLICT-PROBE'`)
	if out, err := runBin(t, bin, benv, "capture-once", "--run-id",
		fmt.Sprintf("conflict-restored-%d", time.Now().UnixNano())); err != nil {
		t.Fatalf("capture after restore: %v\n%s", err, out)
	}
	var consistency string
	if err := pool.QueryRow(ctx,
		`SELECT poc_run_consistency(r.id) FROM payroll_run r
		  JOIN employer e ON e.id = r.employer_id WHERE e.code = $1`,
		authEmployer).Scan(&consistency); err != nil || consistency != "consistent" {
		t.Fatalf("post-restore consistency = %q err=%v", consistency, err)
	}
	authCleanup(t, admin)
}

// TestErrorRetry_NewAttemptConverges proves DB-4 on a fresh battery-
// employer month: an attempt that ERRORs (approval timeout) is
// retried under a NEW ID with no table surgery; both attempts
// resolve to the single verified outcome.
func TestErrorRetry_NewAttemptConverges(t *testing.T) {
	pool := pocPool(t)
	admin := pocAdminPool(t)
	ctx := context.Background()
	// Battery employer, recycled first: attempt 1 must ERROR on a
	// genuinely unapproved draft; a persisted validated run would
	// change the meaning of both attempts. MERIDIAN stays for Phase E.
	authSetup(t, admin)
	bin := buildPocRun(t)
	benv := map[string]string{"POC_EMPLOYER": authEmployer, "POC_PERIOD": authPeriod}

	// Attempt 1: short approval wait, no approver → ERROR (timeout).
	out, err := runBin(t, bin, benv, "--approval-timeout", "5s")
	if err == nil {
		t.Fatalf("unapproved attempt succeeded:\n%s", out)
	}
	if !strings.Contains(string(out), "approval wait") {
		t.Fatalf("want approval-wait timeout, got:\n%s\n%v", out, err)
	}
	var runID string
	if err := pool.QueryRow(ctx,
		`SELECT r.id FROM payroll_run r JOIN employer e ON e.id = r.employer_id
		  WHERE e.code = $1`, authEmployer).Scan(&runID); err != nil {
		t.Fatal(err)
	}
	// Attempt 2 (long wait) in the background; approve mid-flight.
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
	// Wait until the new attempt exists, then approve in a retry
	// loop: a Send issued before the attempt reaches Recv may miss
	// (proven live), and approve is convergent (re-send safe), so
	// retry until the attempt completes or genuinely fails.
	waitFor(t, "attempt 2 enqueued", 120*time.Second, func() bool {
		var n int
		_ = pool.QueryRow(ctx,
			`SELECT count(*) FROM poc_workflow_attempt a
			  JOIN dbos.workflow_status s ON s.workflow_uuid = a.workflow_id
			 WHERE a.run_id = $1 AND s.status = 'PENDING'`, runID).Scan(&n)
		return n >= 1
	})
	deadline := time.Now().Add(240 * time.Second)
	approved := false
	for time.Now().Before(deadline) {
		approveOut, approveErr := runBin(t, bin, benv, "approve", "--run-id", runID,
			"--decision", "approve", "--approver", "error-retry-test")
		if approveErr == nil {
			approved = true
		} else if !strings.Contains(approveOut, "no rows in result set") {
			t.Fatalf("approve: %v\n%s", approveErr, approveOut)
		}
		// Else: no waiter yet (attempt still starting); retry.
		select {
		case err := <-done:
			if err != nil {
				t.Fatalf("attempt 2: %v", err)
			}
			goto completed
		case <-time.After(20 * time.Second):
		}
	}
	_ = cmd.Process.Kill()
	t.Fatal("attempt 2 did not complete after approval")
completed:
	if !approved {
		t.Fatal("approve never succeeded")
	}
	// Single business effect across both attempts.
	for _, q := range []struct{ label, sql string }{
		{"runs", `SELECT count(*) FROM payroll_run WHERE id = $1`},
		{"outputs", `SELECT count(*) FROM poc_run_output WHERE run_id = $1`},
		{"approvals", `SELECT count(*) FROM poc_approval WHERE run_id = $1`},
		{"outbox", `SELECT count(*) FROM poc_outbox WHERE op_uuid = $1`},
	} {
		var n int
		if err := pool.QueryRow(ctx, q.sql, runID).Scan(&n); err != nil || n != 1 {
			t.Fatalf("%s = %d err=%v, want 1", q.label, n, err)
		}
	}
	authCleanup(t, admin)
}

// TestApprovalCrash_RecoverAndComplete proves injection #4 on a FRESH
// battery-employer month: SIGKILL the executor mid-approval-wait,
// recover with a fresh executor, approve, and converge to exactly
// one validated, enqueued outcome.
func TestApprovalCrash_RecoverAndComplete(t *testing.T) {
	pool := pocPool(t)
	admin := pocAdminPool(t)
	ctx := context.Background()
	authSetup(t, admin)
	bin := buildPocRun(t)
	benv := map[string]string{"POC_EMPLOYER": authEmployer, "POC_PERIOD": authPeriod}

	// Start the month; wait until results are persisted and the
	// attempt sits in the approval wait.
	cmd := exec.Command(bin, "--approval-timeout", "300s")
	cmd.Env = os.Environ()
	for k, v := range benv {
		cmd.Env = append(cmd.Env, k+"="+v)
	}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	var runID string
	waitFor(t, "battery month at approval wait", 150*time.Second, func() bool {
		return pool.QueryRow(ctx,
			`SELECT r.id FROM payroll_run r JOIN employer e ON e.id = r.employer_id
			  JOIN poc_run_output o ON o.run_id = r.id
			  JOIN poc_workflow_attempt a ON a.run_id = r.id
			  JOIN dbos.workflow_status s ON s.workflow_uuid = a.workflow_id
			 WHERE e.code = $1 AND s.status = 'PENDING' LIMIT 1`,
			authEmployer).Scan(&runID) == nil
	})
	// Crash the executor mid-wait. The workflow stays PENDING (no
	// executor holds it); nothing completes without recovery.
	_ = cmd.Process.Kill()
	_ = cmd.Wait()
	time.Sleep(3 * time.Second)
	var st string
	if err := pool.QueryRow(ctx,
		`SELECT r.status FROM payroll_run r WHERE r.id = $1`, runID).Scan(&st); err != nil || st != "draft" {
		t.Fatalf("run status after crash = %q err=%v, want draft", st, err)
	}
	// Recover with a fresh executor, approve mid-recovery, converge.
	rec := exec.Command(bin, "recover", "--for", "180s")
	rec.Env = os.Environ()
	if err := rec.Start(); err != nil {
		t.Fatal(err)
	}
	recdone := make(chan error, 1)
	go func() { recdone <- rec.Wait() }()
	// Approve in a retry loop (early Sends may miss a not-yet-resumed
	// Recv; approve is convergent), until the run validates.
	validated := false
	deadline := time.Now().Add(150 * time.Second)
	for time.Now().Before(deadline) && !validated {
		if out, err := runBin(t, bin, benv, "approve", "--run-id", runID,
			"--decision", "approve", "--approver", "crash-test"); err != nil {
			if !strings.Contains(string(out), "no rows in result set") {
				t.Fatalf("approve after recovery: %v\n%s", err, out)
			}
		}
		for i := 0; i < 6 && !validated; i++ {
			var s string
			_ = pool.QueryRow(ctx,
				`SELECT status FROM payroll_run WHERE id = $1`, runID).Scan(&s)
			validated = s == "validated"
			if !validated {
				time.Sleep(5 * time.Second)
			}
		}
	}
	if !validated {
		t.Fatal("run not validated after recovery + approval")
	}
	select {
	case <-recdone:
	case <-time.After(60 * time.Second):
		_ = rec.Process.Kill()
		t.Fatal("recover did not exit")
	}
	// Single effect: 1 run, 1 output, 1 approval, 1 outbox row.
	for _, q := range []struct{ label, sql string }{
		{"outputs", `SELECT count(*) FROM poc_run_output WHERE run_id = $1`},
		{"approvals", `SELECT count(*) FROM poc_approval WHERE run_id = $1`},
		{"outbox", `SELECT count(*) FROM poc_outbox WHERE op_uuid = $1`},
	} {
		var n int
		if err := pool.QueryRow(ctx, q.sql, runID).Scan(&n); err != nil || n != 1 {
			t.Fatalf("%s = %d err=%v, want 1", q.label, n, err)
		}
	}
	authCleanup(t, admin)
}
