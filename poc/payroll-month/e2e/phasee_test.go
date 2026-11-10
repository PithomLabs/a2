package e2e

// Phase E: the shared October-2026 demonstration (prompt8 §7). One
// employer (POC-MERIDIAN), 25 employees, 31 days, existing fixtures —
// the actual database-backed workflow runs all thirteen steps with a
// crash injected mid-approval (kill -9 + DBOS recovery), and the
// database-backed result must be demonstrably identical to the
// canonical DB-free computation (same snapshot inputs, same
// calculator version).
//
// The test recycles MERIDIAN's run-scoped artifacts first (harness
// scaffolding on the disposable POC DB; backlinks persist by design,
// scoped by run) and reopens the period at the end (production close
// stays irreversible — the reopen exists only so later suite runs
// keep a writable fixture).

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/PithomLabs/a2/poc/payroll-month/dbosflow"
	"github.com/PithomLabs/a2/poc/payroll-month/workload"
)

const (
	phaseEEmployer = "POC-MERIDIAN"
	phaseEPeriod   = "PP-2026-10"
)

// TestPhaseE_OctoberDemonstration is the §7 demonstration with failure
// injection: crash during the approval wait, DBOS recovery, approval,
// validation, close (then harness reopen), evidence drain, content
// audit, and a correction — ending with digest equality between the
// database-backed result and the canonical computation.
func TestPhaseE_OctoberDemonstration(t *testing.T) {
	pool := pocPool(t)
	admin := pocAdminPool(t)
	ctx := context.Background()
	auditEnv(t)
	menv := map[string]string{"POC_EMPLOYER": phaseEEmployer, "POC_PERIOD": phaseEPeriod}
	bin := buildPocRun(t)
	auditor := buildAuditor(t)

	// Canonical reference (DB-free): the same calculator over the
	// shared fixture inputs.
	canon, err := workload.Canonical(workload.Build())
	if err != nil {
		t.Fatalf("canonical: %v", err)
	}
	canonDigest, err := workload.MonthPayrollDigest(canon)
	if err != nil {
		t.Fatalf("canonical digest: %v", err)
	}
	couts := workload.BuildOutputs(canon)

	// Recycle MERIDIAN run-scoped artifacts (rerunnable demonstration).
	var emp string
	if err := admin.QueryRow(ctx,
		`SELECT id FROM employer WHERE code = $1`, phaseEEmployer).Scan(&emp); err != nil {
		t.Fatalf("meridian employer: %v", err)
	}
	recycleEmployer(t, admin, emp, "phasee")

	// Steps 1-9 with crash injection: start the month, kill -9 once
	// outputs persist (mid approval wait), recover, approve, complete.
	cmd := exec.Command(bin, "--approval-timeout", "150s")
	cmd.Env = os.Environ()
	for k, v := range menv {
		cmd.Env = append(cmd.Env, k+"="+v)
	}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	var runID string
	waitFor(t, "meridian outputs persisted", 180*time.Second, func() bool {
		return pool.QueryRow(ctx,
			`SELECT r.id FROM payroll_run r JOIN employer e ON e.id = r.employer_id
			  JOIN poc_run_output o ON o.run_id = r.id
			  WHERE e.code = $1`, phaseEEmployer).Scan(&runID) == nil
	})
	if err := cmd.Process.Kill(); err != nil {
		t.Fatalf("kill month: %v", err)
	}
	_ = cmd.Wait()
	rec := exec.Command(bin, "recover", "--for", "150s")
	rec.Env = os.Environ()
	for k, v := range menv {
		rec.Env = append(rec.Env, k+"="+v)
	}
	if err := rec.Start(); err != nil {
		t.Fatal(err)
	}
	recDone := make(chan error, 1)
	go func() { recDone <- rec.Wait() }()
	deadline := time.Now().Add(240 * time.Second)
	approved := false
	for time.Now().Before(deadline) {
		approveOut, approveErr := runBin(t, bin, menv, "approve", "--run-id", runID,
			"--decision", "approve", "--approver", "phasee-demo")
		if approveErr == nil {
			approved = true
		} else if !strings.Contains(approveOut, "no rows in result set") {
			t.Fatalf("approve: %v\n%s", approveErr, approveOut)
		}
		select {
		case err := <-recDone:
			if err != nil {
				t.Fatalf("recover: %v", err)
			}
			goto recovered
		case <-time.After(20 * time.Second):
		}
	}
	_ = rec.Process.Kill()
	t.Fatal("recovery did not complete after approval")
recovered:
	if !approved {
		t.Fatal("approve never succeeded")
	}
	var status string
	if err := pool.QueryRow(ctx,
		`SELECT status FROM payroll_run WHERE id = $1`, runID).Scan(&status); err != nil || status != "validated" {
		t.Fatalf("run status = %q err=%v, want validated", status, err)
	}

	// Steps 1-8 artifacts (deterministic fixture pins the counts).
	var members, included, excluded, claims, pins int
	pool.QueryRow(ctx, `SELECT count(*) FROM payroll_snapshot_employee WHERE snapshot_id =
		(SELECT id FROM payroll_snapshot WHERE payroll_run_id = $1 AND status = 'valid')`, runID).Scan(&members)
	pool.QueryRow(ctx, `SELECT count(*) FROM payroll_snapshot_employee WHERE snapshot_id =
		(SELECT id FROM payroll_snapshot WHERE payroll_run_id = $1 AND status = 'valid') AND decision = 'included'`, runID).Scan(&included)
	pool.QueryRow(ctx, `SELECT count(*) FROM payroll_snapshot_employee WHERE snapshot_id =
		(SELECT id FROM payroll_snapshot WHERE payroll_run_id = $1 AND status = 'valid') AND decision <> 'included'`, runID).Scan(&excluded)
	pool.QueryRow(ctx, `SELECT count(*) FROM poc_claim WHERE employer_id = $1`, emp).Scan(&claims)
	pool.QueryRow(ctx, `SELECT count(*) FROM payroll_snapshot_pack WHERE snapshot_id =
		(SELECT id FROM payroll_snapshot WHERE payroll_run_id = $1 AND status = 'valid')`, runID).Scan(&pins)
	if members != 25 || included != 24 || excluded != 1 {
		t.Fatalf("members=%d included=%d excluded=%d, want 25/24/1", members, included, excluded)
	}
	if claims != 673 || pins != 2 {
		t.Fatalf("claims=%d pins=%d, want 673/2", claims, pins)
	}
	var lines, slips int
	var outDigest string
	if err := pool.QueryRow(ctx,
		`SELECT lines, payslips, output_digest FROM poc_run_output WHERE run_id = $1`,
		runID).Scan(&lines, &slips, &outDigest); err != nil {
		t.Fatal(err)
	}
	if slips != 24 || lines <= 0 {
		t.Fatalf("lines=%d payslips=%d, want >0/24", lines, slips)
	}

	// Step 13 (digest level): independent recompute from persisted
	// snapshot inputs must equal the canonical digest.
	members2, snapPins, err := dbosflow.LoadSnapshot(ctx, pool, runID)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	packs, err := dbosflow.PacksToRulePacks(snapPins)
	if err != nil {
		t.Fatalf("packs: %v", err)
	}
	dbm, dbDigest, err := dbosflow.CalcMonth(members2, packs, "2026-10", phaseEEmployer)
	if err != nil {
		t.Fatalf("recompute: %v", err)
	}
	if dbDigest != canonDigest {
		t.Fatalf("DB-backed %s != canonical %s", dbDigest, canonDigest)
	}
	if len(dbm.Blocked) != len(canon.Blocked) {
		t.Fatalf("blocked %d != canonical %d", len(dbm.Blocked), len(canon.Blocked))
	}
	// Net pay equality (exact cents, no floating point anywhere).
	var netDB int64
	if err := pool.QueryRow(ctx,
		`SELECT SUM((amount*100)::bigint) FROM payroll_line WHERE run_id = $1`, runID).Scan(&netDB); err != nil {
		t.Fatal(err)
	}
	if netDB != couts.Register.Total.NetCents {
		t.Fatalf("net pay %d != canonical %d", netDB, couts.Register.Total.NetCents)
	}

	// Step 10: close the period as the app role; writers lock out;
	// duplicate close rejected; then harness reopen (documented).
	var closed time.Time
	if err := pool.QueryRow(ctx,
		`SELECT poc_close_period($1, $2, $3)`, emp, phaseEPeriod, runID).Scan(&closed); err != nil {
		t.Fatalf("close: %v", err)
	}
	if closed.IsZero() {
		t.Fatal("empty close timestamp")
	}
	var mEmp, mPer string
	if err := pool.QueryRow(ctx,
		`SELECT id FROM employee WHERE employer_id = $1 LIMIT 1`, emp).Scan(&mEmp); err != nil {
		t.Fatal(err)
	}
	if err := pool.QueryRow(ctx,
		`SELECT id FROM pay_period WHERE employer_id = $1 AND code = $2`, emp, phaseEPeriod).Scan(&mPer); err != nil {
		t.Fatal(err)
	}
	mustFailClosed(t, pool, "closed-period timesheet INSERT",
		`INSERT INTO timesheet (id, employer_id, employee_id, pay_period_id, status)
		 VALUES (gen_random_uuid(), '`+emp+`', '`+mEmp+`', '`+mPer+`', 'draft')`,
		"P0001")
	if _, err := pool.Exec(ctx, `SELECT poc_close_period($1, $2, $3)`,
		emp, phaseEPeriod, runID); err == nil {
		t.Fatal("duplicate close accepted, want loud rejection")
	} else if !strings.Contains(err.Error(), "P0001") {
		t.Fatalf("want P0001, got: %v", err)
	}
	execOK(t, admin, `DELETE FROM poc_period_close WHERE employer_id = '`+emp+`'`)
	execOK(t, admin, `UPDATE pay_period SET status = 'open' WHERE employer_id = '`+emp+`' AND code = '`+phaseEPeriod+`'`)

	// Steps 11-12: enqueue (automatic at finalization) + drain.
	runPocRun(t, bin, "outbox-once", "--op", runID)
	if got := outboxState(t, pool, runID); got != "backlinked" {
		t.Fatalf("event state = %q, want backlinked", got)
	}
	// Step 13 (content level): the independent auditor verifies.
	if code, rep := auditOnce(t, auditor, phaseEEmployer, runID); code != 0 || rep.Overall != "pass" {
		t.Fatalf("meridian audit exit=%d overall=%q", code, rep.Overall)
	}
	// Correction: second event appends; both verify with lineage.
	ev2 := fmt.Sprintf("5e5e5e5e-5e5e-5e5e-5e5e-%012x", time.Now().UnixNano()&0xffffffffffff)
	var enq string
	if err := pool.QueryRow(ctx, `SELECT poc_outbox_enqueue($1, $2, $3, $4)`,
		emp, runID, ev2, "phasee-demo").Scan(&enq); err != nil || enq != "enqueued" {
		t.Fatalf("correction enqueue = %q err=%v", enq, err)
	}
	runPocRun(t, bin, "outbox-once", "--op", ev2)
	code, rep := auditOnce(t, auditor, phaseEEmployer, runID)
	if code != 0 || rep.Overall != "pass" || len(rep.Events) != 2 {
		t.Fatalf("correction audit exit=%d overall=%q events=%d", code, rep.Overall, len(rep.Events))
	}

	t.Logf("PHASE-E DEMONSTRATION run=%s digest=%s net=%d¢ finalized=%d blocked=%d exceptions=%d lines=%d payslips=%d events=%d",
		runID, dbDigest, netDB, included, len(dbm.Blocked), len(couts.Exceptions), lines, slips, len(rep.Events))
}
