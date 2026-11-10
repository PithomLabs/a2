package e2e

// Phase B authority battery (prompt8 §3.2 + §6 PG-authority list).
// Proves the fabrication-to-validation chain is closed AND the
// authorized path composes, against persisted state (never error
// strings alone). Runs on a minimal second employer (POC-AUTH-01) so
// the main POC-MERIDIAN run stays untouched for Phase E; every test
// cleans its run-scoped artifacts first (rerunnable, order-free).

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const authEmployer = "POC-AUTH-01"
const authPeriod = "PP-AUTH-01"

// authCleanup deletes run-scoped artifacts for the battery employer.
// Input rows (employees/terms/period/sheets/intervals/schedule) are
// stable across reruns (authSetup is idempotent); only the per-run
// artifacts are recycled. Backlinks are append-only but the battery
// never enqueues, so none exist for this employer.
func authCleanup(t *testing.T, admin *pgxpool.Pool) {
	t.Helper()
	emp := authEmpID(t, admin)
	if emp == "" {
		return
	}
	recycleEmployer(t, admin, emp, "battery")
}

// recycleEmployer deletes run-scoped artifacts for one employer and
// reopens its period (harness scaffolding on the disposable POC DB;
// backlinks persist by design, scoped by run). Phase E uses it for
// POC-MERIDIAN; the battery tests for POC-AUTH-01.
func recycleEmployer(t *testing.T, admin *pgxpool.Pool, emp, why string) {
	t.Helper()
	ctx := context.Background()
	for _, q := range []string{
		`DELETE FROM poc_workflow_attempt WHERE employer_id = $1`,
		`DELETE FROM poc_outbox WHERE employer_id = $1`,
		`DELETE FROM poc_approval WHERE employer_id = $1`,
		`DELETE FROM poc_run_output WHERE employer_id = $1`,
		`DELETE FROM poc_claim WHERE employer_id = $1`,
		`DELETE FROM payroll_snapshot_pack WHERE snapshot_id IN
			(SELECT id FROM payroll_snapshot WHERE employer_id = $1)`,
		`DELETE FROM payroll_snapshot_employee WHERE snapshot_id IN
			(SELECT id FROM payroll_snapshot WHERE employer_id = $1)`,
		`DELETE FROM payroll_snapshot WHERE employer_id = $1`,
		`DELETE FROM payroll_line WHERE employer_id = $1`,
		`DELETE FROM payslip WHERE employer_id = $1`,
		`DELETE FROM poc_period_close WHERE employer_id = $1`,
		`DELETE FROM payroll_run WHERE employer_id = $1`,
		// Reopen AFTER the close rows are gone (the writer trigger
		// would reject the status flip while a close stands): this
		// recycles the battery period for the next run. Production
		// close stays irreversible — no reopen path exists there.
		`UPDATE pay_period SET status = 'open'
		  WHERE employer_id = $1 AND status = 'locked'`,
	} {
		if _, err := admin.Exec(ctx, q, emp); err != nil {
			t.Fatalf("%s recycle %q: %v", why, q, err)
		}
	}
}

func authEmpID(t *testing.T, admin *pgxpool.Pool) string {
	t.Helper()
	var id string
	err := admin.QueryRow(context.Background(),
		`SELECT id FROM employer WHERE code = $1`, authEmployer).Scan(&id)
	if err != nil {
		return ""
	}
	return id
}

// authSetup provisions the battery inputs idempotently (admin role,
// harness scaffolding): employer + period + 2 employees
// (AUTH-001: single terms span + approved sheet + one interval +
// comp lines → included; AUTH-002: two terms spans → excluded) +
// pack schedule pins.
func authSetup(t *testing.T, admin *pgxpool.Pool) string {
	t.Helper()
	// Recycle first: a previous run may have left the battery period
	// locked, which would reject even the idempotent setup INSERTs
	// (BEFORE triggers fire before ON CONFLICT). No-op on first run.
	authCleanup(t, admin)
	execOK(t, admin, `INSERT INTO employer (id, code, name, uen, address)
		VALUES ('aaaaaaaa-0000-4000-8000-000000000001', '`+authEmployer+`',
			'Authority battery (synthetic)', 'SYNTH-AUTH-001', '1 Synthetic Row')
		ON CONFLICT (id) DO NOTHING`)
	emp := authEmpID(t, admin)
	execOK(t, admin, `INSERT INTO pay_period
		(id, employer_id, code, starts_on, ends_on, pay_date, salary_deadline, ot_deadline, status)
		VALUES ('aaaaaaaa-0000-4000-8000-000000000002', '`+emp+`', '`+authPeriod+`',
			'2026-10-01', '2026-10-31', '2026-11-07', '2026-11-07', '2026-11-14', 'open')
		ON CONFLICT (id) DO NOTHING`)
	for _, e := range []struct{ id, no string }{
		{"aaaaaaaa-0000-4000-8000-000000000011", "AUTH-001"},
		{"aaaaaaaa-0000-4000-8000-000000000012", "AUTH-002"},
	} {
		execOK(t, admin, `INSERT INTO employee
			(id, employer_id, employee_no, legal_name, address, gender,
			 date_of_birth, citizenship, status, status_effective_date)
			VALUES ('`+e.id+`', '`+emp+`', '`+e.no+`', 'Battery `+e.no+`',
				'1 Synthetic Row', 'X', '1990-01-01', 'SG', 'active', '2026-01-01')
			ON CONFLICT (id) DO NOTHING`)
		execOK(t, admin, `INSERT INTO employee_status_history
			(employer_id, employee_id, status, effective_from, reason)
			SELECT '`+emp+`', '`+e.id+`', 'active', '2026-01-01', 'battery'
			WHERE NOT EXISTS (SELECT 1 FROM employee_status_history
				WHERE employee_id = '`+e.id+`')`)
	}
	// AUTH-001: single span; AUTH-002: two spans (base change).
	execOK(t, admin, `INSERT INTO employment_terms
		(id, employer_id, employee_id, effective_from, job_title, job_duties,
		 employment_category, part_iv_eligible, eligibility_basis, monthly_basic,
		 working_arrangement, normal_hours_weekly, working_days, daily_hours,
		 rest_day, start_date, contract_type, place_of_work, salary_period)
		VALUES ('aaaaaaaa-0000-4000-8000-000000000021',
			'`+emp+`', 'aaaaaaaa-0000-4000-8000-000000000011', '2026-01-01',
			'Agent', 'Support', 'non_workman', true, 'battery', 2400,
			'shift', 44, 'Mon-Fri', 8, 'Sunday', '2026-01-01', 'permanent',
			'Singapore', 'monthly')
		ON CONFLICT (id) DO NOTHING`)
	for _, s := range []struct{ id, from, to, basic string }{
		{"aaaaaaaa-0000-4000-8000-000000000022", "2026-10-01", "2026-10-15", "3280"},
		{"aaaaaaaa-0000-4000-8000-000000000023", "2026-10-16", "'infinity'", "3360"},
	} {
		to := s.to
		if to == "'infinity'" {
			to = "NULL"
		} else {
			to = "'" + to + "'"
		}
		execOK(t, admin, `INSERT INTO employment_terms
			(id, employer_id, employee_id, effective_from, effective_to, job_title, job_duties,
			 employment_category, part_iv_eligible, eligibility_basis, monthly_basic,
			 working_arrangement, normal_hours_weekly, working_days, daily_hours,
			 rest_day, start_date, contract_type, place_of_work, salary_period)
			VALUES ('`+s.id+`', '`+emp+`', 'aaaaaaaa-0000-4000-8000-000000000012', '`+s.from+`',
				`+to+`, 'Agent', 'Support', 'non_workman', true, 'battery', `+s.basic+`,
				'shift', 44, 'Mon-Fri', 8, 'Sunday', '2026-01-01', 'permanent',
				'Singapore', 'monthly')
			ON CONFLICT (id) DO NOTHING`)
	}
	// AUTH-001: approved sheet + one current work interval + comp lines.
	execOK(t, admin, `INSERT INTO timesheet
		(id, employer_id, employee_id, pay_period_id, status, approved_by, approved_at)
		VALUES ('aaaaaaaa-0000-4000-8000-000000000031', '`+emp+`',
			'aaaaaaaa-0000-4000-8000-000000000011',
			'aaaaaaaa-0000-4000-8000-000000000002', 'approved', 'battery', now())
		ON CONFLICT (id) DO NOTHING`)
	execOK(t, admin, `INSERT INTO attendance_interval
		(id, employer_id, employee_id, interval_start, interval_end, timezone,
		 local_date, kind, source, source_refs, is_current)
		VALUES ('aaaaaaaa-0000-4000-8000-000000000041', '`+emp+`',
			'aaaaaaaa-0000-4000-8000-000000000011',
			'2026-10-05T09:00:00+08:00', '2026-10-05T17:00:00+08:00', 'Asia/Singapore',
			'2026-10-05', 'work', 'assumption', '{"assumption":"battery"}', true)
		ON CONFLICT (id) DO NOTHING`)
	execOK(t, admin, `SELECT add_compensation('`+emp+`',
		'aaaaaaaa-0000-4000-8000-000000000011',
		'aaaaaaaa-0000-4000-8000-000000000021',
		'allowance', 'AUTH-ALLOW', 100.00, '2026-10-01', '2026-10-31',
		'battery', 'battery as-given')
		WHERE NOT EXISTS (SELECT 1 FROM compensation_line
			WHERE employee_id = 'aaaaaaaa-0000-4000-8000-000000000011'
			  AND code = 'AUTH-ALLOW')`)
	// Pack schedule pins (same pinned versions AND params as the POC
	// employer: params must parse as the pinned RulePack versions).
	for _, p := range []struct{ v, from, to, digest, numer, denom string }{
		{"SYNTH-2026-10-v1a", "2026-10-01", "2026-10-15",
			"ca584074d3bbd3eb1dd874e5a6b0766e98c7f1adaa9210a664535b1556d115f6", "3", "2"},
		{"SYNTH-2026-10-v2a", "2026-10-16", "2026-10-31",
			"323fd2694a5f015a2f61cbe8cbbad61c802d3ee35efeb02d8abeba0eb1ec273a", "2", "1"},
	} {
		params := `{"version":"` + p.v + `","effective_from":"` + p.from +
			`","effective_to":"` + p.to + `","synthetic":true,` +
			`"overtime_threshold_min":480,"overtime_numer":` + p.numer +
			`,"overtime_denom":` + p.denom + `,"leave_deduction_denom":30}`
		execOK(t, admin, `INSERT INTO poc_rule_pack_schedule
			(employer_id, pack, version, effective_from, effective_to, status, digest, params)
			VALUES ('`+emp+`', 'SYNTH-OT-001', '`+p.v+`', '`+p.from+`', '`+p.to+`',
				'IMPLEMENTABLE', '`+p.digest+`', '`+params+`')
			ON CONFLICT (employer_id, pack, version) DO UPDATE SET params = EXCLUDED.params`)
	}
	return emp
}

func mustFailClosed(t *testing.T, pool *pgxpool.Pool, label, sql string, codes ...string) {
	t.Helper()
	if _, err := pool.Exec(context.Background(), sql); err == nil {
		t.Fatalf("%s accepted, want rejection", label)
	} else {
		ok := false
		for _, c := range codes {
			if strings.Contains(err.Error(), c) {
				ok = true
			}
		}
		if !ok {
			t.Fatalf("%s: want one of %v, got: %v", label, codes, err)
		}
	}
}

// TestAuthority_ExploitChain is prompt8 §3.2: the full PG-1 chain as
// the real app role, every step rejected, persisted state unchanged.
func TestAuthority_ExploitChain(t *testing.T) {
	pool := pocPool(t)
	admin := pocAdminPool(t)
	ctx := context.Background()
	emp := authSetup(t, admin)
	authCleanup(t, admin)

	// Legitimate parents exist (the chain's precondition).
	var nEmp int
	if err := pool.QueryRow(ctx,
		`SELECT count(*) FROM employee WHERE employer_id = $1`, emp).Scan(&nEmp); err != nil || nEmp != 2 {
		t.Fatalf("battery parents: emp=%d err=%v", nEmp, err)
	}
	// Steps 3-5: fabricated header/members/pins rejected (42501).
	mustFailClosed(t, pool, "fabricated snapshot header",
		`INSERT INTO payroll_snapshot (employer_id, payroll_run_id, pay_period_id,
		  version, status, content_hash, jurisdiction, currency, captured_by)
		 VALUES ('`+emp+`', gen_random_uuid(),
		  'aaaaaaaa-0000-4000-8000-000000000002', 1, 'valid', 'FORGED', 'SG', 'SGD', 'chain')`,
		"42501")
	mustFailClosed(t, pool, "fabricated snapshot member",
		`INSERT INTO payroll_snapshot_employee (snapshot_id, employee_id,
		  employee_basis_seq, decision, inputs, row_hash)
		 VALUES (gen_random_uuid(), 'aaaaaaaa-0000-4000-8000-000000000011',
		  0, 'included', '{}', 'FORGED')`,
		"42501")
	mustFailClosed(t, pool, "fabricated pack pin",
		`INSERT INTO payroll_snapshot_pack (snapshot_id, pack, version, status)
		 VALUES (gen_random_uuid(), 'SYNTH-OT-001', 'vFAKE', 'IMPLEMENTABLE')`,
		"42501")
	// Step 6: manufactured validation evidence (claims) rejected.
	mustFailClosed(t, pool, "manufactured claim",
		`INSERT INTO poc_claim (run_id, employer_id, table_name, row_id, input_hash)
		 VALUES (gen_random_uuid(), '`+emp+`', 'employee', 'x', decode('ab','hex'))`,
		"42501")
	// Step 7: direct promotion rejected (no UPDATE on runs).
	mustFailClosed(t, pool, "direct run promotion",
		`UPDATE payroll_run SET status = 'validated'
		  WHERE employer_id = '`+emp+`'`,
		"42501")
	// Persisted state: nothing fabricated exists for this employer.
	var nSnap, nRun int
	if err := pool.QueryRow(ctx,
		`SELECT count(*) FROM payroll_snapshot WHERE employer_id = $1`, emp).Scan(&nSnap); err != nil || nSnap != 0 {
		t.Fatalf("snapshots persisted: n=%d err=%v", nSnap, err)
	}
	if err := pool.QueryRow(ctx,
		`SELECT count(*) FROM payroll_run WHERE employer_id = $1`, emp).Scan(&nRun); err != nil || nRun != 0 {
		t.Fatalf("runs persisted: n=%d err=%v", nRun, err)
	}
}

// TestAuthority_CaptureSucceeds proves the authorized path composes:
// capture derives 2 members (included + excluded), pins, claims,
// content hash and run fields — all verified from persisted rows.
func TestAuthority_CaptureSucceeds(t *testing.T) {
	pool := pocPool(t)
	admin := pocAdminPool(t)
	ctx := context.Background()
	emp := authSetup(t, admin)
	authCleanup(t, admin)

	var runID, snapID, hash string
	if err := pool.QueryRow(ctx,
		`SELECT * FROM poc_snapshot_capture($1, $2, 'authority-battery')`,
		emp, authPeriod).Scan(&runID, &snapID, &hash); err != nil {
		t.Fatalf("authorized capture: %v", err)
	}
	if hash == "" || hash == "pending" {
		t.Fatalf("content hash not derived: %q", hash)
	}
	var inc, exc, miss int
	if err := pool.QueryRow(ctx,
		`SELECT count(*) FILTER (WHERE decision = 'included'),
		        count(*) FILTER (WHERE decision = 'excluded'),
		        count(*) FILTER (WHERE decision = 'missing-timesheet')
		   FROM payroll_snapshot_employee WHERE snapshot_id = $1`,
		snapID).Scan(&inc, &exc, &miss); err != nil {
		t.Fatal(err)
	}
	if inc != 1 || exc != 1 || miss != 0 {
		t.Fatalf("decisions inc/exc/miss = %d/%d/%d, want 1/1/0", inc, exc, miss)
	}
	var blocker string
	if err := pool.QueryRow(ctx,
		`SELECT blockers->0->>'code' FROM payroll_snapshot_employee
		  WHERE snapshot_id = $1 AND decision = 'excluded'`, snapID).Scan(&blocker); err != nil || blocker != "B-BASE-CHANGE" {
		t.Fatalf("excluded blocker = %q err=%v, want B-BASE-CHANGE", blocker, err)
	}
	var pins int
	if err := pool.QueryRow(ctx,
		`SELECT count(*) FROM payroll_snapshot_pack WHERE snapshot_id = $1`,
		snapID).Scan(&pins); err != nil || pins != 2 {
		t.Fatalf("pins = %d err=%v, want 2", pins, err)
	}
	// Claims: 2 emp + 3 terms + 1 sheet + 1 interval + 0 leave + 1 comp + 2 packs = 10.
	var claims int
	if err := pool.QueryRow(ctx,
		`SELECT count(*) FROM poc_claim WHERE run_id = $1`, runID).Scan(&claims); err != nil || claims != 10 {
		t.Fatalf("claims = %d err=%v, want 10", claims, err)
	}
	var inputsHash, calc string
	if err := pool.QueryRow(ctx,
		`SELECT inputs_hash, calc_version FROM payroll_run WHERE id = $1`, runID).Scan(&inputsHash, &calc); err != nil {
		t.Fatal(err)
	}
	if inputsHash != hash || calc != "poc-synth/v1" {
		t.Fatalf("run derived fields: hash match=%v calc=%q", inputsHash == hash, calc)
	}
	// Second capture on the same run is rejected (capture-once rule).
	if _, err := pool.Exec(ctx,
		`SELECT * FROM poc_snapshot_capture($1, $2, 'authority-battery')`,
		emp, authPeriod); err == nil {
		t.Fatal("second capture accepted, want correction-path rejection")
	} else if !strings.Contains(err.Error(), "P0001") {
		t.Fatalf("want P0001, got: %v", err)
	}
}

// TestAuthority_VacuousRejected proves empty evidence never validates:
// an owner-planted header-only snapshot fails the verifier loudly.
func TestAuthority_VacuousRejected(t *testing.T) {
	admin := pocAdminPool(t)
	ctx := context.Background()
	emp := authSetup(t, admin)
	authCleanup(t, admin)

	// Owner plants a header-only snapshot + draft run (the DEFINER
	// would never produce this shape; only the owner can).
	var runID, snapID string
	if err := admin.QueryRow(ctx,
		`INSERT INTO payroll_run (id, employer_id, pay_period_id, kind, status,
		  inputs_hash, rule_versions, calc_version, rounding_conventions)
		 VALUES (gen_random_uuid(), $1,
		  'aaaaaaaa-0000-4000-8000-000000000002',
		  'regular', 'draft', 'x', '{}', 'poc-synth/v1', '{}') RETURNING id`,
		emp).Scan(&runID); err != nil {
		t.Fatalf("plant run: %v", err)
	}
	if err := admin.QueryRow(ctx,
		`INSERT INTO payroll_snapshot (employer_id, payroll_run_id, pay_period_id,
		  version, status, content_hash, jurisdiction, currency, captured_by)
		 VALUES ($1, $2, 'aaaaaaaa-0000-4000-8000-000000000002',
		  1, 'valid', 'x', 'SG', 'SGD', 'battery') RETURNING id`,
		emp, runID).Scan(&snapID); err != nil {
		t.Fatalf("plant snapshot: %v", err)
	}
	if _, err := admin.Exec(ctx,
		`SELECT poc_verify_claims($1, $2)`, runID, snapID); err == nil {
		t.Fatal("vacuous evidence verified, want loud rejection")
	} else if !strings.Contains(err.Error(), "vacuous") {
		t.Fatalf("want vacuous-evidence rejection, got: %v", err)
	}
	authCleanup(t, admin)
}

// TestAuthority_GateInterleavings proves the capture path and the
// business writers share one authoritative protocol (prompt8 §3.3):
// a writer blocks while the gate is held elsewhere, and two
// concurrent captures serialize to exactly one winner.
func TestAuthority_GateInterleavings(t *testing.T) {
	pool := pocPool(t)
	admin := pocAdminPool(t)
	ctx := context.Background()
	emp := authSetup(t, admin)
	authCleanup(t, admin)

	// Writer-first ordering: hold the employer gate on the admin
	// connection; an app writer must BLOCK (statement timeout proves
	// the wait — instant success would mean no participation).
	hold, err := admin.Acquire(ctx)
	if err != nil {
		t.Fatalf("acquire holder: %v", err)
	}
	defer hold.Release()
	holdTx, err := hold.Begin(ctx)
	if err != nil {
		t.Fatalf("begin holder: %v", err)
	}
	if _, err := holdTx.Exec(ctx, `SELECT poc_gate_acquire($1)`, emp); err != nil {
		t.Fatalf("holder acquire: %v", err)
	}
	wconn, err := pool.Acquire(ctx)
	if err != nil {
		t.Fatalf("acquire writer: %v", err)
	}
	defer wconn.Release()
	if _, err := wconn.Exec(ctx, `SET statement_timeout = '3s'`); err != nil {
		t.Fatalf("writer timeout: %v", err)
	}
	wstart := time.Now()
	_, werr := wconn.Exec(ctx, `INSERT INTO clock_event
		(id, employer_id, employee_id, source, source_event_id, occurred_at, timezone, kind)
		VALUES (gen_random_uuid(), $1, 'aaaaaaaa-0000-4000-8000-000000000011',
			'battery', 'gate-wait-`+fmt.Sprintf("%d", time.Now().UnixNano())+`',
			'2026-10-07T09:00:00+08:00', 'Asia/Singapore', 'in')`, emp)
	if werr == nil {
		t.Fatal("writer INSERT succeeded while the gate was held elsewhere")
	}
	if !strings.Contains(werr.Error(), "57014") {
		t.Fatalf("want statement-timeout cancel (57014), got: %v", werr)
	}
	if time.Since(wstart) < 2*time.Second {
		t.Fatalf("writer failed fast (%v): no gate wait observed", time.Since(wstart))
	}
	// Release: the holder rolls back, the writer proceeds immediately.
	if err := holdTx.Rollback(ctx); err != nil {
		t.Fatalf("holder rollback: %v", err)
	}
	if _, err := wconn.Exec(ctx, `SET statement_timeout = '30s'`); err != nil {
		t.Fatalf("writer timeout reset: %v", err)
	}
	execOK(t, pool, `INSERT INTO clock_event
		(id, employer_id, employee_id, source, source_event_id, occurred_at, timezone, kind)
		VALUES (gen_random_uuid(), '`+emp+`', 'aaaaaaaa-0000-4000-8000-000000000011',
			'battery', 'gate-free-`+fmt.Sprintf("%d", time.Now().UnixNano())+`',
			'2026-10-07T09:00:00+08:00', 'Asia/Singapore', 'in')`)

	// Capture race: two concurrent captures, exactly one winner.
	type capRes struct {
		run, snap, hash string
		err             error
	}
	res := make(chan capRes, 2)
	for i := 0; i < 2; i++ {
		go func() {
			var r capRes
			r.err = pool.QueryRow(ctx,
				`SELECT * FROM poc_snapshot_capture($1, $2, 'authority-race')`,
				emp, authPeriod).Scan(&r.run, &r.snap, &r.hash)
			res <- r
		}()
	}
	wins := 0
	for i := 0; i < 2; i++ {
		if r := <-res; r.err == nil {
			wins++
		} else if !strings.Contains(r.err.Error(), "P0001") {
			t.Fatalf("loser error want P0001, got: %v", r.err)
		}
	}
	if wins != 1 {
		t.Fatalf("concurrent captures: %d winners, want exactly 1", wins)
	}
	var tips int
	if err := pool.QueryRow(ctx,
		`SELECT count(*) FROM payroll_snapshot s JOIN payroll_run r ON r.id = s.payroll_run_id
		  WHERE r.employer_id = $1 AND s.status = 'valid'`, emp).Scan(&tips); err != nil || tips != 1 {
		t.Fatalf("valid tips = %d err=%v, want 1", tips, err)
	}
	authCleanup(t, admin)
}

// TestAuthority_ClosedPeriod proves close semantics end to end on the
// battery employer: capture → persist → approve → validate → close,
// then period-mapped writers are rejected (INSERT/UPDATE/DELETE)
// while unmapped writers stay gate-only.
func TestAuthority_ClosedPeriod(t *testing.T) {
	pool := pocPool(t)
	admin := pocAdminPool(t)
	ctx := context.Background()
	emp := authSetup(t, admin)
	authCleanup(t, admin)

	var runID, snapID, hash string
	if err := pool.QueryRow(ctx,
		`SELECT * FROM poc_snapshot_capture($1, $2, 'authority-battery')`,
		emp, authPeriod).Scan(&runID, &snapID, &hash); err != nil {
		t.Fatalf("capture: %v", err)
	}
	var termsID string
	if err := pool.QueryRow(ctx,
		`SELECT id FROM employment_terms
		  WHERE employee_id = 'aaaaaaaa-0000-4000-8000-000000000011'`).Scan(&termsID); err != nil {
		t.Fatal(err)
	}
	lines := `[{"employee_no":"AUTH-001","component":"basic","pay_code":"REG",` +
		`"hours":null,"amount_cents":240000,"cpf_treatment":"OW",` +
		`"work_period":"2026-10","payable_date":"2026-11-07",` +
		`"contribution_period":"2026-10",` +
		`"source_refs":{"timesheet_line_ids":[],"terms_id":"` + termsID + `",` +
		`"rule_pack_versions":{},"calculation":{"by":"battery"}}}]`
	var outDigest string
	if err := pool.QueryRow(ctx,
		`SELECT poc_result_persist($1, 'poc-synth/v1', $2, $3::jsonb, '[]'::jsonb, 'authority-battery')`,
		runID, hash, lines).Scan(&outDigest); err != nil {
		t.Fatalf("persist: %v", err)
	}
	if outDigest == "" {
		t.Fatal("empty output digest")
	}
	var dec string
	if err := pool.QueryRow(ctx,
		`SELECT poc_approve($1, 'battery-approver', 'approve', 'battery')`,
		runID).Scan(&dec); err != nil || dec != "approve" {
		t.Fatalf("approve: %q %v", dec, err)
	}
	var st string
	if err := pool.QueryRow(ctx,
		`SELECT poc_run_validate($1, 'authority-battery')`, runID).Scan(&st); err != nil || st != "validated" {
		t.Fatalf("validate: %q %v", st, err)
	}
	var closed time.Time
	if err := pool.QueryRow(ctx,
		`SELECT poc_close_period($1, $2, $3)`, emp, authPeriod, runID).Scan(&closed); err != nil {
		t.Fatalf("close: %v", err)
	}
	if closed.IsZero() {
		t.Fatal("empty close timestamp")
	}
	// Period-mapped writers rejected on all three operations.
	mustFailClosed(t, pool, "closed-period timesheet INSERT",
		`INSERT INTO timesheet (id, employer_id, employee_id, pay_period_id, status)
		 VALUES (gen_random_uuid(), '`+emp+`',
		  'aaaaaaaa-0000-4000-8000-000000000011',
		  'aaaaaaaa-0000-4000-8000-000000000002', 'draft')`,
		"P0001")
	mustFailClosed(t, pool, "closed-period timesheet UPDATE",
		`UPDATE timesheet SET status = 'submitted'
		  WHERE id = 'aaaaaaaa-0000-4000-8000-000000000031'`,
		"P0001")
	mustFailClosed(t, pool, "closed-period timesheet DELETE",
		`DELETE FROM timesheet WHERE id = 'aaaaaaaa-0000-4000-8000-000000000031'`,
		"P0001")
	var sheetAlive int
	if err := pool.QueryRow(ctx,
		`SELECT count(*) FROM timesheet
		  WHERE id = 'aaaaaaaa-0000-4000-8000-000000000031'`).Scan(&sheetAlive); err != nil || sheetAlive != 1 {
		t.Fatalf("rejected DELETE removed the row: n=%d err=%v", sheetAlive, err)
	}
	// Unmapped writers stay gate-only (documented limitation): a clock
	// row for the closed period still writes (no period relationship).
	execOK(t, pool, `INSERT INTO clock_event
		(id, employer_id, employee_id, source, source_event_id, occurred_at, timezone, kind)
		VALUES (gen_random_uuid(), '`+emp+`', 'aaaaaaaa-0000-4000-8000-000000000011',
			'battery', 'closed-unmapped-`+fmt.Sprintf("%d", time.Now().UnixNano())+`',
			'2026-10-06T09:00:00+08:00', 'Asia/Singapore', 'in')`)
	// Duplicate close rejected loudly.
	if _, err := pool.Exec(ctx, `SELECT poc_close_period($1, $2, $3)`,
		emp, authPeriod, runID); err == nil {
		t.Fatal("duplicate close accepted, want loud rejection")
	} else if !strings.Contains(err.Error(), "P0001") {
		t.Fatalf("want P0001, got: %v", err)
	}
	authCleanup(t, admin)
}
