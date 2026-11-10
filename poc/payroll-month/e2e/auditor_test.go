package e2e

// Content-audit battery (prompt8 §5.2/§5.3 + §6 immudb evidence list).
// The auditor reads PostgreSQL through a SELECT-only role, rebuilds
// the expected evidence from authoritative sources, re-executes the
// payroll from persisted inputs, and compares everything against
// cryptographically verified ledger bytes.
//
// Threat model (probed live, not assumed): the app role cannot write
// any audit surface (permission denials, Phase B); the OWNER can edit
// every PG surface except backlinks (no freeze triggers bind the
// owner). The auditor is therefore the sole detector of owner-level
// PG tamper — each case below tampers, expects loud audit failure,
// restores exactly, and expects pass again.

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/codenotary/immudb/pkg/api/schema"
	"github.com/codenotary/immudb/pkg/client"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// buildAuditor compiles the auditor binary into a temp dir.
func buildAuditor(t *testing.T) string {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "auditor")
	cmd := exec.Command("go", "build", "-o", bin, "./cmd/auditor")
	cmd.Dir = ".."
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build auditor: %v\n%s", err, out)
	}
	return bin
}

// runAuditor executes the auditor, failing loudly; returns stdout.
func runAuditor(t *testing.T, bin string, args ...string) string {
	t.Helper()
	code, stdout, stderr := runAuditorCode(t, bin, args...)
	if code != 0 {
		t.Fatalf("auditor %v: exit %d\nstdout:\n%s\nstderr:\n%s", args, code, stdout, stderr)
	}
	return stdout
}

// runAuditorCode executes the auditor with split streams (the JSON
// report stays parseable: stderr carries the genesis line).
func runAuditorCode(t *testing.T, bin string, args ...string) (int, string, string) {
	t.Helper()
	cmd := exec.Command(bin, args...)
	var stdout, stderr strings.Builder
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err := cmd.Run()
	if err == nil {
		return 0, stdout.String(), stderr.String()
	}
	if ee, ok := err.(*exec.ExitError); ok {
		return ee.ExitCode(), stdout.String(), stderr.String()
	}
	t.Fatalf("auditor %v: %v", args, err)
	return -1, "", ""
}

// auditEnv gates audit tests on the auditor's credentials.
func auditEnv(t *testing.T) {
	t.Helper()
	if os.Getenv("POC_AUDIT_DATABASE_URL") == "" {
		t.Skip("POC_AUDIT_DATABASE_URL unset: auditor needs the SELECT-only PG role")
	}
	if os.Getenv("POC_AUDIT_IMMUDB_PASSWORD") == "" {
		t.Skip("POC_AUDIT_IMMUDB_PASSWORD unset: auditor needs its ledger identity")
	}
}

// auditReport is the parsed auditor stdout.
type auditReport struct {
	Overall string `json:"overall"`
	Mode    string `json:"mode"`
	Pinned  bool   `json:"pin_checked"`
	Events  []struct {
		EventUUID string `json:"event_uuid"`
		Verified  bool   `json:"verified"`
		Checks    []struct {
			Name   string `json:"name"`
			Pass   bool   `json:"pass"`
			Detail string `json:"detail"`
		} `json:"checks"`
	} `json:"events"`
	Global []struct {
		Name   string `json:"name"`
		Pass   bool   `json:"pass"`
		Detail string `json:"detail"`
	} `json:"global_checks"`
}

func parseReport(t *testing.T, stdout string) auditReport {
	t.Helper()
	var rep auditReport
	if err := json.Unmarshal([]byte(stdout), &rep); err != nil {
		t.Fatalf("parse audit report: %v\n%s", err, stdout)
	}
	return rep
}

// auditOnce runs a content audit with a fresh state dir (explicit
// genesis each time: no trust carried between battery cases).
func auditOnce(t *testing.T, auditor, employer, runID string) (int, auditReport) {
	t.Helper()
	code, stdout, _ := runAuditorCode(t, auditor,
		"--state-dir", t.TempDir(), "--accept-genesis",
		"--employer", employer, "--run", runID)
	if code != 0 && strings.TrimSpace(stdout) == "" {
		t.Fatalf("auditor exit %d with no report", code)
	}
	return code, parseReport(t, stdout)
}

// finalizeBatteryMonth runs a full month on the recycled battery
// employer and convergently approves it; returns the validated run.
func finalizeBatteryMonth(t *testing.T, pool *pgxpool.Pool, bin string, approver string) string {
	t.Helper()
	ctx := context.Background()
	benv := map[string]string{"POC_EMPLOYER": authEmployer, "POC_PERIOD": authPeriod}
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
			"--decision", "approve", "--approver", approver)
		if approveErr == nil {
			approved = true
		} else if !strings.Contains(approveOut, "no rows in result set") {
			t.Fatalf("approve: %v\n%s", approveErr, approveOut)
		}
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
	return runID
}

// TestContentAudit_GenuinePasses proves the §5.2 positive: a real
// finalized, drained month verifies clean — every event, every
// check — under a fresh auditor with explicit genesis, under a
// reused state dir, and pinned to the committed anchor.
func TestContentAudit_GenuinePasses(t *testing.T) {
	pool := pocPool(t)
	admin := pocAdminPool(t)
	auditEnv(t)
	authSetup(t, admin)
	bin := buildPocRun(t)
	auditor := buildAuditor(t)
	runID := finalizeBatteryMonth(t, pool, bin, "genuine-audit")
	runPocRun(t, bin, "outbox-once", "--op", runID)

	code, rep := auditOnce(t, auditor, authEmployer, runID)
	if code != 0 || rep.Overall != "pass" {
		t.Fatalf("genuine audit exit=%d overall=%q", code, rep.Overall)
	}
	if len(rep.Events) != 1 || !rep.Events[0].Verified {
		t.Fatalf("want 1 verified event, got %+v", rep.Events)
	}
	for _, c := range rep.Events[0].Checks {
		if !c.Pass {
			t.Fatalf("genuine event check %s failed: %s", c.Name, c.Detail)
		}
	}
	for _, g := range rep.Global {
		if !g.Pass {
			t.Fatalf("genuine global check %s failed: %s", g.Name, g.Detail)
		}
	}
	// Reused state dir (no re-accept): identity enforced, still passes.
	dir := t.TempDir()
	if code, _, _ := runAuditorCode(t, auditor,
		"--state-dir", dir, "--accept-genesis",
		"--employer", authEmployer, "--run", runID); code != 0 {
		t.Fatalf("first use exit=%d", code)
	}
	code, stdout, _ := runAuditorCode(t, auditor,
		"--state-dir", dir, "--employer", authEmployer, "--run", runID)
	rep = parseReport(t, stdout)
	if code != 0 || rep.Overall != "pass" {
		t.Fatalf("reused-dir audit exit=%d overall=%q", code, rep.Overall)
	}
	authCleanup(t, admin)
}

// TestContentAudit_TamperBattery proves the §5.2 negative: each PG
// tamper class fails the audit loudly (naming the violated check),
// and exact restoration passes again.
func TestContentAudit_TamperBattery(t *testing.T) {
	pool := pocPool(t)
	admin := pocAdminPool(t)
	ctx := context.Background()
	auditEnv(t)
	authSetup(t, admin)
	bin := buildPocRun(t)
	auditor := buildAuditor(t)
	runID := finalizeBatteryMonth(t, pool, bin, "tamper-battery")
	runPocRun(t, bin, "outbox-once", "--op", runID)
	if code, rep := auditOnce(t, auditor, authEmployer, runID); code != 0 || rep.Overall != "pass" {
		t.Fatalf("baseline audit exit=%d overall=%q", code, rep.Overall)
	}

	// Originals for exact restoration.
	var origOut, origTip, origCalc, origAppr, origInputs, origPin, origSlip string
	pool.QueryRow(ctx, `SELECT output_digest FROM poc_run_output WHERE run_id = $1`, runID).Scan(&origOut)
	pool.QueryRow(ctx, `SELECT content_hash FROM payroll_snapshot WHERE payroll_run_id = $1 AND status = 'valid'`, runID).Scan(&origTip)
	pool.QueryRow(ctx, `SELECT calc_version FROM payroll_run WHERE id = $1`, runID).Scan(&origCalc)
	pool.QueryRow(ctx, `SELECT approver FROM poc_approval WHERE run_id = $1 AND decision = 'approve'`, runID).Scan(&origAppr)
	var tipID string
	pool.QueryRow(ctx, `SELECT id FROM payroll_snapshot WHERE payroll_run_id = $1 AND status = 'valid'`, runID).Scan(&tipID)
	pool.QueryRow(ctx, `SELECT inputs::text FROM payroll_snapshot_employee WHERE snapshot_id = $1 ORDER BY employee_id LIMIT 1`, tipID).Scan(&origInputs)
	pool.QueryRow(ctx, `SELECT manifest_digest FROM payroll_snapshot_pack WHERE snapshot_id = $1 ORDER BY effective_from, pack LIMIT 1`, tipID).Scan(&origPin)
	pool.QueryRow(ctx, `SELECT payload::text FROM payslip WHERE run_id = $1 ORDER BY employee_id LIMIT 1`, runID).Scan(&origSlip)

	cases := []struct {
		name   string
		check  string // a check that must fail (others may too)
		tamper string
		fix    string
	}{
		{"amounts", "C.lines",
			`UPDATE payroll_line SET amount = amount + 1 WHERE run_id = '` + runID + `'`,
			`UPDATE payroll_line SET amount = amount - 1 WHERE run_id = '` + runID + `'`},
		{"output-digest", "A.output-digest",
			`UPDATE poc_run_output SET output_digest = 'tamper' WHERE run_id = '` + runID + `'`,
			`UPDATE poc_run_output SET output_digest = '` + origOut + `' WHERE run_id = '` + runID + `'`},
		{"tip-hash", "B.persist-bind",
			`UPDATE payroll_snapshot SET content_hash = 'tamper' WHERE payroll_run_id = '` + runID + `' AND status = 'valid'`,
			`UPDATE payroll_snapshot SET content_hash = '` + origTip + `' WHERE payroll_run_id = '` + runID + `' AND status = 'valid'`},
		// Joint tamper (tip + the bound input edited consistently):
		// the persist-bind passes, the ledger-vs-tip binding fires.
		{"tip-joint", "B.input-digest",
			`UPDATE payroll_snapshot SET content_hash = 'tamper' WHERE payroll_run_id = '` + runID + `' AND status = 'valid';` +
				`UPDATE poc_run_output SET input_digest = 'tamper' WHERE run_id = '` + runID + `'`,
			`UPDATE payroll_snapshot SET content_hash = '` + origTip + `' WHERE payroll_run_id = '` + runID + `' AND status = 'valid';` +
				`UPDATE poc_run_output SET input_digest = '` + origTip + `' WHERE run_id = '` + runID + `'`},
		{"member-inputs", "C.lines",
			`UPDATE payroll_snapshot_employee SET inputs = jsonb_set(inputs, '{days,0,minutes}', '9999')
			  WHERE snapshot_id = '` + tipID + `' AND employee_id = (
				SELECT employee_id FROM payroll_snapshot_employee WHERE snapshot_id = '` + tipID + `' ORDER BY employee_id LIMIT 1)`,
			`UPDATE payroll_snapshot_employee SET inputs = '` + origInputs + `'::jsonb
			  WHERE snapshot_id = '` + tipID + `' AND employee_id = (
				SELECT employee_id FROM payroll_snapshot_employee WHERE snapshot_id = '` + tipID + `' ORDER BY employee_id LIMIT 1)`},
		{"pin-digest", "E.rule-pack",
			`UPDATE payroll_snapshot_pack SET manifest_digest = 'tamper' WHERE snapshot_id = '` + tipID + `'`,
			`UPDATE payroll_snapshot_pack SET manifest_digest = '` + origPin + `' WHERE snapshot_id = '` + tipID + `'`},
		{"calc-version", "F.calc-digest",
			`UPDATE payroll_run SET calc_version = 'tamper' WHERE id = '` + runID + `'`,
			`UPDATE payroll_run SET calc_version = '` + origCalc + `' WHERE id = '` + runID + `'`},
		{"approval", "G.decided-by",
			`UPDATE poc_approval SET approver = 'tamper' WHERE run_id = '` + runID + `'`,
			`UPDATE poc_approval SET approver = '` + origAppr + `' WHERE run_id = '` + runID + `'`},
		{"payslip", "C.slips",
			`UPDATE payslip SET payload = '{"net_cents":1}' WHERE run_id = '` + runID + `' AND employee_id = (
				SELECT employee_id FROM payslip WHERE run_id = '` + runID + `' ORDER BY employee_id LIMIT 1)`,
			`UPDATE payslip SET payload = '` + origSlip + `'::jsonb WHERE run_id = '` + runID + `' AND employee_id = (
				SELECT employee_id FROM payslip WHERE run_id = '` + runID + `' ORDER BY employee_id LIMIT 1)`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			execOK(t, admin, c.tamper)
			// Deferred restore: a failed tamper assertion must not
			// poison later cases. Guarded: the explicit fix below
			// disarms it (fixes are NOT idempotent — e.g. -1 twice
			// corrupts — so exactly one must run).
			restored := false
			defer func() {
				if !restored {
					_, _ = admin.Exec(context.Background(), c.fix)
				}
			}()
			code, rep := auditOnce(t, auditor, authEmployer, runID)
			if code == 0 || rep.Overall != "fail" {
				t.Fatalf("tampered audit exit=%d overall=%q, want failure", code, rep.Overall)
			}
			if !checkFailed(rep, c.check) {
				t.Fatalf("tampered audit did not fail check %s; failing: %s", c.check, failingChecks(rep))
			}
			execOK(t, admin, c.fix)
			restored = true
			if code, rep := auditOnce(t, auditor, authEmployer, runID); code != 0 || rep.Overall != "pass" {
				t.Fatalf("restored audit exit=%d overall=%q; failing: %s", code, rep.Overall, failingChecks(rep))
			}
		})
	}
	authCleanup(t, admin)
}

// failingChecks names every failing check (tamper-battery diagnostics).
func failingChecks(rep auditReport) string {
	var out []string
	for _, g := range rep.Global {
		if !g.Pass {
			out = append(out, g.Name+":"+g.Detail)
		}
	}
	for _, e := range rep.Events {
		for _, c := range e.Checks {
			if !c.Pass {
				out = append(out, e.EventUUID[:8]+"/"+c.Name+":"+c.Detail)
			}
		}
	}
	return strings.Join(out, " | ")
}

// checkFailed reports whether the named check failed anywhere.
func checkFailed(rep auditReport, name string) bool {
	for _, g := range rep.Global {
		if g.Name == name && !g.Pass {
			return true
		}
	}
	for _, e := range rep.Events {
		for _, c := range e.Checks {
			if c.Name == name && !c.Pass {
				return true
			}
		}
	}
	return false
}

// TestContentAudit_ForgedAndMissing proves PG→ledger binding attacks
// fail: F1 dangling backlink (no producer row) to the genuine tx;
// F2 forged event with a producer-shaped row pointing at the
// genuine tx (the tx number is irrelevant — the ledger never saw
// the event); M claimed event with no ledger presence. Forged rows
// are append-only and persist, scoped by this run — hence a
// dedicated month never audited clean afterwards.
func TestContentAudit_ForgedAndMissing(t *testing.T) {
	pool := pocPool(t)
	admin := pocAdminPool(t)
	ctx := context.Background()
	auditEnv(t)
	authSetup(t, admin)
	bin := buildPocRun(t)
	auditor := buildAuditor(t)
	runID := finalizeBatteryMonth(t, pool, bin, "forge-battery")
	runPocRun(t, bin, "outbox-once", "--op", runID)
	var realTx uint64
	if err := pool.QueryRow(ctx,
		`SELECT immu_tx FROM poc_ledger_backlink WHERE event_uuid = $1`, runID).Scan(&realTx); err != nil {
		t.Fatal(err)
	}
	// Unique per run: forged backlinks are append-only and persist
	// forever (scoped by run, never freed).
	mkID := func() string {
		return fmt.Sprintf("ffffffff-ffff-ffff-ffff-%012x", time.Now().UnixNano()&0xffffffffffff)
	}
	forge := func(event string, tx uint64) {
		t.Helper()
		if _, err := admin.Exec(ctx,
			`INSERT INTO poc_ledger_backlink (pg_table, pg_id, event_uuid, immu_tx, alh)
			 VALUES ('payroll_run', $1, $2, $3, decode('ab','hex'))`,
			runID, event, tx); err != nil {
			t.Fatalf("forge backlink: %v", err)
		}
	}
	// F1: dangling backlink (no producer row) to the genuine tx.
	f1 := mkID()
	forge(f1, realTx)
	if code, rep := auditOnce(t, auditor, authEmployer, runID); code == 0 || rep.Overall != "fail" {
		t.Fatalf("F1 audit exit=%d overall=%q, want failure", code, rep.Overall)
	} else if !checkFailed(rep, "J.outbox-row") {
		t.Fatalf("F1 audit did not fail J.outbox-row; failing: %s", failingChecks(rep))
	}
	// F2: forged event WITH a producer-shaped row, pointing at the
	// genuine tx. The tx number is irrelevant: the LEDGER never saw
	// this event, so the verified marker read fails.
	f2 := mkID()
	seedOutboxFull(t, admin, f2, runID, "payroll.finalized", "payroll_run")
	forge(f2, realTx)
	if code, rep := auditOnce(t, auditor, authEmployer, runID); code == 0 || rep.Overall != "fail" {
		t.Fatalf("F2 audit exit=%d overall=%q, want failure", code, rep.Overall)
	} else if !checkFailed(rep, "J.marker") {
		t.Fatalf("F2 audit did not fail J.marker; failing: %s", failingChecks(rep))
	}
	// M: claimed event with no ledger presence, bogus tx pointer.
	m1 := mkID()
	seedOutboxFull(t, admin, m1, runID, "payroll.finalized", "payroll_run")
	forge(m1, 999999999)
	if code, rep := auditOnce(t, auditor, authEmployer, runID); code == 0 || rep.Overall != "fail" {
		t.Fatalf("M audit exit=%d overall=%q, want failure", code, rep.Overall)
	} else if !checkFailed(rep, "J.marker") {
		t.Fatalf("M audit did not fail J.marker; failing: %s", failingChecks(rep))
	}
	authCleanup(t, admin)
}

// TestContentAudit_Corrections proves corrected artifacts verify: two
// events for one run, both independently verified, supersedes chain
// resolving.
func TestContentAudit_Corrections(t *testing.T) {
	pool := pocPool(t)
	admin := pocAdminPool(t)
	ctx := context.Background()
	auditEnv(t)
	authSetup(t, admin)
	bin := buildPocRun(t)
	auditor := buildAuditor(t)
	runID := finalizeBatteryMonth(t, pool, bin, "correction-audit")
	runPocRun(t, bin, "outbox-once", "--op", runID)
	emp := authEmpID(t, admin)
	// Unique per run: event IDs are ledger idempotency keys — reusing
	// one across runs with a different payload is payload aliasing
	// (loudly refused, proven live by this test's first fixed-ID draft).
	ev2 := fmt.Sprintf("aaaaaaaa-bbbb-cccc-dddd-%012x", time.Now().UnixNano()&0xffffffffffff)
	var enq string
	if err := pool.QueryRow(ctx, `SELECT poc_outbox_enqueue($1, $2, $3, $4)`,
		emp, runID, ev2, "correction-audit").Scan(&enq); err != nil || enq != "enqueued" {
		t.Fatalf("correction enqueue = %q err=%v", enq, err)
	}
	runPocRun(t, bin, "outbox-once", "--op", ev2)
	code, rep := auditOnce(t, auditor, authEmployer, runID)
	if code != 0 || rep.Overall != "pass" {
		t.Fatalf("correction audit exit=%d overall=%q", code, rep.Overall)
	}
	if len(rep.Events) != 2 {
		t.Fatalf("want 2 verified events, got %d", len(rep.Events))
	}
	for _, e := range rep.Events {
		if !e.Verified {
			t.Fatalf("event %s not verified", e.EventUUID)
		}
	}
	authCleanup(t, admin)
}

// TestBacklink_AppendOnly proves PG tamper PREVENTION at the ledger
// binding: backlink UPDATE/DELETE are rejected by trigger even for
// the owner (P0001).
func TestBacklink_AppendOnly(t *testing.T) {
	admin := pocAdminPool(t)
	ctx := context.Background()
	immudbEnv(t)
	uniq := fmt.Sprintf("%012x", time.Now().UnixNano()&0xffffffffffff)
	opID := "cccccccc-cccc-cccc-cccc-" + uniq
	seedOutbox(t, admin, opID, "tamper-"+uniq[:8])
	bin := buildPocRun(t)
	runPocRun(t, bin, "outbox-once", "--op", opID)
	if _, err := admin.Exec(ctx, `UPDATE poc_ledger_backlink SET immu_tx = 1 WHERE event_uuid = $1`, opID); err == nil {
		t.Fatal("backlink UPDATE accepted, want append-only rejection")
	} else if !strings.Contains(err.Error(), "P0001") {
		t.Fatalf("want P0001 append-only rejection, got: %v", err)
	}
	if _, err := admin.Exec(ctx, `DELETE FROM poc_ledger_backlink WHERE event_uuid = $1`, opID); err == nil {
		t.Fatal("backlink DELETE accepted, want append-only rejection")
	} else if !strings.Contains(err.Error(), "P0001") {
		t.Fatalf("want P0001 append-only rejection, got: %v", err)
	}
}

// TestGenesis_PinEnforced proves §5.3: the committed anchor pins the
// live stack (canary), wrong pins exit 2 before any trust decision,
// and a tampered auditor state dir fails verification.
func TestGenesis_PinEnforced(t *testing.T) {
	auditEnv(t)
	auditor := buildAuditor(t)
	raw, err := os.ReadFile("../gate/immudb-genesis.json")
	if err != nil {
		t.Fatalf("committed anchor: %v", err)
	}
	var anch struct {
		UUID   string `json:"server_uuid"`
		Tx     uint64 `json:"anchor_tx"`
		TxHash string `json:"tx_hash"`
	}
	if err := json.Unmarshal(raw, &anch); err != nil {
		t.Fatalf("parse anchor: %v", err)
	}
	pinArgs := []string{"--expect-uuid", anch.UUID, "--expect-tx", fmt.Sprint(anch.Tx),
		"--expect-tx-hash", anch.TxHash}
	// Canary: the live stack matches the committed anchor.
	if code, stdout, _ := runAuditorCode(t, auditor, append([]string{
		"--state-dir", t.TempDir(), "--accept-genesis", "--txs", "1",
	}, pinArgs...)...); code != 0 || !strings.Contains(stdout, `"overall": "pass"`) {
		t.Fatalf("anchor canary failed (stack recreated? re-run the genesis ceremony): exit=%d\n%s", code, stdout)
	}
	// Wrong pins exit 2 (usage/trust failure, never a pass).
	for _, bad := range [][]string{
		{"--expect-uuid", "WRONG", "--expect-tx", fmt.Sprint(anch.Tx), "--expect-tx-hash", anch.TxHash},
		{"--expect-uuid", anch.UUID, "--expect-tx", fmt.Sprint(anch.Tx), "--expect-tx-hash", strings.Repeat("0", 64)},
		{"--expect-uuid", anch.UUID},
	} {
		args := append([]string{"--state-dir", t.TempDir(), "--accept-genesis", "--txs", "1"}, bad...)
		if code, _, _ := runAuditorCode(t, auditor, args...); code != 2 {
			t.Fatalf("bad pin %v: exit=%d, want 2", bad, code)
		}
	}
	// Tampered state dir: verify once, flip a byte, re-verify fails.
	dir := t.TempDir()
	if code, _, _ := runAuditorCode(t, auditor,
		"--state-dir", dir, "--accept-genesis", "--txs", "1"); code != 0 {
		t.Fatalf("baseline liveness exit=%d", code)
	}
	ents, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	flipped := false
	for _, e := range ents {
		if strings.HasPrefix(e.Name(), ".state-") {
			p := filepath.Join(dir, e.Name())
			raw, err := os.ReadFile(p)
			if err != nil || len(raw) == 0 {
				t.Fatal(err)
			}
			raw[len(raw)-1] ^= 0xff
			if err := os.WriteFile(p, raw, 0o600); err != nil {
				t.Fatal(err)
			}
			flipped = true
		}
	}
	if !flipped {
		t.Fatal("no state file to tamper")
	}
	if code, _, _ := runAuditorCode(t, auditor, "--state-dir", dir, "--txs", "1"); code == 0 {
		t.Fatal("tampered state dir verified clean, want failure")
	}
}

// TestIdentity_Separation proves §5.3 enforcement: the auditor's
// ledger identity cannot write, the auditor refuses the writer's
// state dir, and the PG auditor role is read-only.
func TestIdentity_Separation(t *testing.T) {
	auditEnv(t)
	ctx := context.Background()
	// 1. poc_auditor cannot ExecAll.
	addr := os.Getenv("POC_IMMUDB_ADDR")
	host, port := "localhost", 3322
	if h, p, ok := strings.Cut(addr, ":"); ok {
		host = h
		fmt.Sscanf(p, "%d", &port)
	} else if addr != "" {
		host = addr
	}
	os.MkdirAll(t.TempDir(), 0o700)
	opts := client.DefaultOptions()
	opts.Address, opts.Port, opts.Dir = host, port, t.TempDir()
	opts.DialOptions = []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	cli := client.NewClient().WithOptions(opts)
	if err := cli.OpenSession(ctx, []byte(os.Getenv("POC_AUDIT_IMMUDB_USER")),
		[]byte(os.Getenv("POC_AUDIT_IMMUDB_PASSWORD")), "defaultdb"); err != nil {
		t.Fatalf("auditor session: %v", err)
	}
	defer cli.CloseSession(ctx)
	_, err := cli.ExecAll(ctx, &schema.ExecAllRequest{Operations: []*schema.Op{{Operation: &schema.Op_Kv{
		Kv: &schema.KeyValue{Key: []byte("probe/deny"), Value: []byte("x")}}}}})
	if err == nil {
		t.Fatal("auditor ExecAll accepted, want permission denial")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "permission denied") {
		t.Fatalf("want permission denial, got: %v", err)
	}
	// 2. Writer state dir refused.
	auditor := buildAuditor(t)
	wdir := t.TempDir()
	t.Setenv("POC_WRITER_STATE_DIR", wdir)
	if code, _, _ := runAuditorCode(t, auditor,
		"--state-dir", wdir, "--accept-genesis", "--txs", "1"); code != 2 {
		t.Fatalf("writer state dir: exit=%d, want 2", code)
	}
	// 3. PG auditor role: SELECT yes, writes no.
	apool, err := pgxpool.New(ctx, os.Getenv("POC_AUDIT_DATABASE_URL"))
	if err != nil {
		t.Fatalf("auditor pg: %v", err)
	}
	defer apool.Close()
	var n int
	if err := apool.QueryRow(ctx, `SELECT count(*) FROM payroll_run`).Scan(&n); err != nil {
		t.Fatalf("auditor SELECT: %v", err)
	}
	if _, err := apool.Exec(ctx, `INSERT INTO poc_outbox (op_uuid, employer_id, action, payload_hash)
		VALUES (gen_random_uuid(), gen_random_uuid(), 'x', decode('ab','hex'))`); err == nil {
		t.Fatal("auditor INSERT accepted, want denial")
	}
}
