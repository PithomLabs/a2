package dbosflow

// Stages A-G payroll workflow (prompt8 §4.1): the DBOS connective
// tissue between PostgreSQL authority (Part 5/6 DEFINERs) and the
// ledger evidence path. Every authoritative write runs inside
// RunAsTransaction (business rows + DBOS checkpoint commit together);
// calculation is pure steps over immutable snapshot rows; approval is
// a durable Recv verified by the DEFINER finalizer.
//
// Identity (DB-3/DB-4 fix): each submission carries a UNIQUE attempt
// workflow ID. Attempts resolve to business runs inside Stage B:
// fresh capture, consistent adopt, or explicit input-version
// conflict — never a silent stale replay. Transient failure retries
// with a NEW attempt ID and converges (no table surgery).
//
// Convergent replay: every txn stage anticipates DBOS at-least-once
// re-execution (commit landed, checkpoint lost) via savepoint +
// re-read: already-captured/persisted/validated outcomes are
// re-verified and adopted, never duplicated or failed.

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/dbos-inc/dbos-transact-golang/dbos"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/PithomLabs/a2/poc/payroll-month/chaos"
	"github.com/PithomLabs/a2/poc/payroll-month/workload"
)

// Pool serves workflow step reads (SELECTs the app role is granted).
// Datasource serves RunAsTransaction bodies (shared-pool atomicity).
// Both wired by main.
var Pool *pgxpool.Pool

// Datasource is the shared-pool DBOS datasource.
var Datasource *dbos.DataSource

// CalcVersionPinned must equal the run's pinned calc_version.
const CalcVersionPinned = CalcVersion

// DefaultApprovalTimeout bounds the approval wait for demo runs.
const DefaultApprovalTimeout = 30 * time.Minute

// PayrollInput identifies one payroll attempt.
type PayrollInput struct {
	EmployerCode string `json:"employer_code"`
	PeriodCode   string `json:"period_code"`
	RequestedBy  string `json:"requested_by"`
	// ApprovalTimeout bounds Stage E (0 → DefaultApprovalTimeout).
	ApprovalTimeout time.Duration `json:"approval_timeout_ns"`
}

// PayrollResult is the checkpointed attempt outcome.
type PayrollResult struct {
	WorkflowID    string `json:"workflow_id"`
	RunID         string `json:"run_id"`
	SnapshotID    string `json:"snapshot_id"`
	InputDigest   string `json:"input_digest"`
	OutputDigest  string `json:"output_digest"`
	PayrollDigest string `json:"payroll_digest"`
	Role          string `json:"role"` // captured | adopted
	Employees     int    `json:"employees"`
	Blocked       int    `json:"blocked"`
	Lines         int    `json:"lines"`
	Validated     bool   `json:"validated"`
	Enqueue       string `json:"enqueue"`
}

// ApprovalMsg is the durable approval decision (Recv payload).
type ApprovalMsg struct {
	Decision string `json:"decision"` // approve | reject
	Approver string `json:"approver"`
	Note     string `json:"note"`
}

// ApprovalTopic is the Recv topic per business run.
func ApprovalTopic(runID string) string { return "approval:" + runID }

// readiness is the resolved Stage A output.
type readiness struct {
	EmployerID string
	PeriodCode string
	PayDate    string
}

// checkReadiness resolves employer/period eligibility (Stage A).
// Read-only; typed rejections for unknown/blocked inputs.
func checkReadiness(ctx context.Context, pool *pgxpool.Pool, employerCode, periodCode string) (readiness, error) {
	var r readiness
	if employerCode == "" || periodCode == "" {
		return r, fmt.Errorf("readiness: empty employer or period")
	}
	err := pool.QueryRow(ctx,
		`SELECT id FROM employer WHERE code = $1`, employerCode).Scan(&r.EmployerID)
	if err != nil {
		return r, fmt.Errorf("readiness: unknown employer %s", employerCode)
	}
	var status string
	err = pool.QueryRow(ctx,
		`SELECT status, pay_date::text FROM pay_period WHERE employer_id = $1 AND code = $2`,
		r.EmployerID, periodCode).Scan(&status, &r.PayDate)
	if err != nil {
		return r, fmt.Errorf("readiness: unknown period %s", periodCode)
	}
	if status != "open" && status != "processing" {
		return r, fmt.Errorf("readiness: period %s is %s", periodCode, status)
	}
	var packs, pop int
	if err := pool.QueryRow(ctx,
		`SELECT count(*) FROM poc_rule_pack_schedule WHERE employer_id = $1`,
		r.EmployerID).Scan(&packs); err != nil || packs == 0 {
		return r, fmt.Errorf("readiness: no rule-pack schedule for employer")
	}
	if err := pool.QueryRow(ctx,
		`SELECT count(*) FROM employee WHERE employer_id = $1`,
		r.EmployerID).Scan(&pop); err != nil || pop == 0 {
		return r, fmt.Errorf("readiness: empty population")
	}
	r.PeriodCode = periodCode
	return r, nil
}

// captureOut is the Stage B outcome (shared by both workflows).
type captureOut struct {
	RunID       string
	SnapshotID  string
	InputDigest string
	Role        string
}

func isAlreadyCaptured(err error) bool {
	return err != nil && isP0001(err) && strings.Contains(err.Error(), "already captured")
}

func isP0001(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "P0001"
	}
	return strings.Contains(err.Error(), "P0001")
}

// stageBCapture resolves the attempt to a business run (Stage B):
// fresh capture, consistent adopt, or explicit conflict. Runs inside
// the caller's transaction (business rows + attempt claim + DBOS
// checkpoint commit together).
func stageBCapture(ctx context.Context, tx dbos.Tx, workflowID string, in PayrollInput, r readiness) (captureOut, error) {
	var out captureOut
	actor := "dbos:" + workflowID + ":" + in.RequestedBy
	// Existing run? (FOR UPDATE serializes attempts on the same run.)
	var runID, status string
	// No FOR UPDATE: app holds SELECT-only on runs (authority
	// overlay). Serialization lives in the DEFINER (gate +
	// one-draft partial unique index + owner-side retry); duplicate
	// racers converge through the already-captured catch below.
	err := tx.QueryRow(ctx,
		`SELECT id, status FROM payroll_run
		  WHERE employer_id = $1 AND pay_period_id =
			(SELECT id FROM pay_period WHERE employer_id = $1 AND code = $2)
		    AND kind = 'regular' ORDER BY created_at LIMIT 1`,
		r.EmployerID, r.PeriodCode).Scan(&runID, &status)
	switch {
	case err == nil && status == "validated":
		return adoptOrConflict(ctx, tx, workflowID, in, r, runID)
	case err == nil:
		// Draft: capture if uncaptured, else adopt-or-conflict.
		var tip string
		if err := tx.QueryRow(ctx,
			`SELECT id FROM payroll_snapshot WHERE payroll_run_id = $1 AND status = 'valid'`,
			runID).Scan(&tip); err == nil {
			return adoptOrConflict(ctx, tx, workflowID, in, r, runID)
		}
	}
	// Fresh capture under a savepoint (a lost race rolls back to the
	// savepoint and adopts instead of poisoning this transaction).
	if _, err := tx.Exec(ctx, `SAVEPOINT p8_capture`); err != nil {
		return out, err
	}
	chaos.Trip("pre-commit-capture") // B1 kill window (inside the txn)
	var snapID, digest string
	err = tx.QueryRow(ctx,
		`SELECT * FROM poc_snapshot_capture($1, $2, $3)`,
		r.EmployerID, r.PeriodCode, actor).Scan(&runID, &snapID, &digest)
	if err != nil {
		if isAlreadyCaptured(err) {
			if _, rb := tx.Exec(ctx, `ROLLBACK TO SAVEPOINT p8_capture`); rb != nil {
				return out, rb
			}
			return adoptAfterRace(ctx, tx, workflowID, in, r)
		}
		return out, fmt.Errorf("capture: %w", err)
	}
	if _, err := tx.Exec(ctx, `RELEASE SAVEPOINT p8_capture`); err != nil {
		return out, err
	}
	if err := claimAttempt(ctx, tx, workflowID, in, r, runID, digest, "captured"); err != nil {
		return out, err
	}
	return captureOut{RunID: runID, SnapshotID: snapID, InputDigest: digest, Role: "captured"}, nil
}

// adoptAfterRace re-resolves after losing a capture race.
func adoptAfterRace(ctx context.Context, tx dbos.Tx, workflowID string, in PayrollInput, r readiness) (captureOut, error) {
	var out captureOut
	var runID string
	err := tx.QueryRow(ctx,
		`SELECT id FROM payroll_run
		  WHERE employer_id = $1 AND pay_period_id =
			(SELECT id FROM pay_period WHERE employer_id = $1 AND code = $2)
		    AND kind = 'regular' ORDER BY created_at LIMIT 1`,
		r.EmployerID, r.PeriodCode).Scan(&runID)
	if err != nil {
		return out, fmt.Errorf("adopt after race: run vanished: %w", err)
	}
	return adoptOrConflict(ctx, tx, workflowID, in, r, runID)
}

// adoptOrConflict implements the DB-3 rule: consistent live inputs
// adopt the existing verified run; changed inputs raise an explicit
// input-version conflict (never a silent stale replay).
func adoptOrConflict(ctx context.Context, tx dbos.Tx, workflowID string, in PayrollInput, r readiness, runID string) (captureOut, error) {
	var out captureOut
	var consistency string
	if err := tx.QueryRow(ctx,
		`SELECT poc_run_consistency($1)`, runID).Scan(&consistency); err != nil {
		return out, fmt.Errorf("adopt: consistency probe: %w", err)
	}
	if consistency != "consistent" {
		return out, fmt.Errorf("input-version conflict on run %s: %s (open a new revision; finalized results are never silently replaced)", runID, consistency)
	}
	var snapID, digest string
	if err := tx.QueryRow(ctx,
		`SELECT id, content_hash FROM payroll_snapshot
		  WHERE payroll_run_id = $1 AND status = 'valid'`, runID).Scan(&snapID, &digest); err != nil {
		return out, fmt.Errorf("adopt: tip lookup: %w", err)
	}
	if err := claimAttempt(ctx, tx, workflowID, in, r, runID, digest, "adopted"); err != nil {
		return out, err
	}
	return captureOut{RunID: runID, SnapshotID: snapID, InputDigest: digest, Role: "adopted"}, nil
}

func claimAttempt(ctx context.Context, tx dbos.Tx, workflowID string, in PayrollInput, r readiness, runID, digest, role string) error {
	_, err := tx.Exec(ctx,
		`INSERT INTO poc_workflow_attempt (workflow_id, run_id, employer_id, period_code, role, input_digest)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		workflowID, runID, r.EmployerID, r.PeriodCode, role, digest)
	return err
}

// calcOut is the Stage C outcome (JSON-serializable for checkpoints).
type calcOut struct {
	Month        workload.Month `json:"month"`
	Digest       string         `json:"digest"`
	LinesJSON    string         `json:"lines_json"`
	PayslipsJSON string         `json:"payslips_json"`
	LineCount    int            `json:"line_count"`
}

// calcStep runs Stage C: snapshot → ComputePay → persist JSON.
func calcStep(ctx context.Context, pool *pgxpool.Pool, cap captureOut, r readiness, employerCode string) (calcOut, error) {
	var out calcOut
	members, pins, err := LoadSnapshot(ctx, pool, cap.RunID)
	if err != nil {
		return out, err
	}
	packs, err := PacksToRulePacks(pins)
	if err != nil {
		return out, err
	}
	m, digest, err := CalcMonth(members, packs, "2026-10", employerCode)
	if err != nil {
		return out, err
	}
	lines, slips := BuildPersistLines(members, pins, m, r.PayDate)
	rawLines, err := json.Marshal(lines)
	if err != nil {
		return out, err
	}
	rawSlips, err := json.Marshal(slips)
	if err != nil {
		return out, err
	}
	out = calcOut{Month: m, Digest: digest, LinesJSON: string(rawLines),
		PayslipsJSON: string(rawSlips), LineCount: len(lines)}
	return out, nil
}

// stageDPersist runs Stage D with convergent replay: an attempt that
// arrives after results exist (at-least-once re-execution, or a new
// attempt over a finalized run) adopts the recorded output digest
// instead of duplicating or failing.
func stageDPersist(ctx context.Context, tx dbos.Tx, cap captureOut, calc calcOut, actor string) (string, error) {
	if _, err := tx.Exec(ctx, `SAVEPOINT p8_persist`); err != nil {
		return "", err
	}
	var digest string
	err := tx.QueryRow(ctx,
		`SELECT poc_result_persist($1, $2, $3, $4::jsonb, $5::jsonb, $6)`,
		cap.RunID, CalcVersion, cap.InputDigest, calc.LinesJSON, calc.PayslipsJSON, actor).Scan(&digest)
	if err != nil {
		if isP0001(err) && (strings.Contains(err.Error(), "already persisted") ||
			strings.Contains(err.Error(), "want draft")) {
			if _, rb := tx.Exec(ctx, `ROLLBACK TO SAVEPOINT p8_persist`); rb != nil {
				return "", rb
			}
			// Convergent adopt: the output must exist AND bind to
			// this attempt's adopted input digest.
			var status, boundInput string
			if err := tx.QueryRow(ctx,
				`SELECT r.status FROM payroll_run r WHERE r.id = $1`,
				cap.RunID).Scan(&status); err != nil {
				return "", fmt.Errorf("persist replay: run vanished: %w", err)
			}
			if err := tx.QueryRow(ctx,
				`SELECT output_digest, input_digest FROM poc_run_output WHERE run_id = $1`,
				cap.RunID).Scan(&digest, &boundInput); err != nil {
				return "", fmt.Errorf("persist replay: run %s (%s) without output: %w",
					cap.RunID, status, err)
			}
			if boundInput != cap.InputDigest {
				return "", fmt.Errorf("persist replay: output binds %s, attempt adopted %s",
					boundInput, cap.InputDigest)
			}
			return digest, nil
		}
		return "", fmt.Errorf("persist: %w", err)
	}
	if _, err := tx.Exec(ctx, `RELEASE SAVEPOINT p8_persist`); err != nil {
		return "", err
	}
	return digest, nil
}

// finalizeOut is the Stage F+G outcome.
type finalizeOut struct {
	Enqueue string
}

// stageFFinalize runs Stages F+G in ONE transaction: validate, then
// enqueue the evidence event (event identity = the run UUID: stable
// across replays, unique per business run).
func stageFFinalize(ctx context.Context, tx dbos.Tx, employerID, runID, actor string) (finalizeOut, error) {
	var out finalizeOut
	if _, err := tx.Exec(ctx, `SAVEPOINT p8_finalize`); err != nil {
		return out, err
	}
	var st string
	err := tx.QueryRow(ctx,
		`SELECT poc_run_validate($1, $2)`, runID, actor).Scan(&st)
	if err != nil {
		if isP0001(err) && strings.Contains(err.Error(), "want draft") {
			if _, rb := tx.Exec(ctx, `ROLLBACK TO SAVEPOINT p8_finalize`); rb != nil {
				return out, rb
			}
			// Convergent replay: re-verify every finalization
			// precondition instead of trusting the label.
			var status string
			if err := tx.QueryRow(ctx,
				`SELECT status FROM payroll_run WHERE id = $1`, runID).Scan(&status); err != nil || status != "validated" {
				return out, fmt.Errorf("finalize replay: run status %q", status)
			}
			var n int
			if err := tx.QueryRow(ctx,
				`SELECT count(*) FROM poc_run_output WHERE run_id = $1`, runID).Scan(&n); err != nil || n != 1 {
				return out, fmt.Errorf("finalize replay: output missing")
			}
			if err := tx.QueryRow(ctx,
				`SELECT count(*) FROM poc_approval WHERE run_id = $1 AND decision = 'approve'`,
				runID).Scan(&n); err != nil || n != 1 {
				return out, fmt.Errorf("finalize replay: approval missing")
			}
		} else {
			return out, fmt.Errorf("finalize validate: %w", err)
		}
	} else if _, err := tx.Exec(ctx, `RELEASE SAVEPOINT p8_finalize`); err != nil {
		return out, err
	}
	if err := tx.QueryRow(ctx,
		`SELECT poc_outbox_enqueue($1, $2, $3, $4)`,
		employerID, runID, runID, actor).Scan(&out.Enqueue); err != nil {
		return out, fmt.Errorf("finalize enqueue: %w", err)
	}
	return out, nil
}

// PayrollWorkflow runs stages A through G for one payroll attempt.
func PayrollWorkflow(ctx dbos.Context, in PayrollInput) (PayrollResult, error) {
	var res PayrollResult
	wid, err := dbos.GetWorkflowID(ctx)
	if err != nil {
		return res, err
	}
	if in.EmployerCode == "" || in.PeriodCode == "" || in.RequestedBy == "" {
		return res, fmt.Errorf("payroll: employer, period and requester are required")
	}
	if Pool == nil || Datasource == nil {
		return res, fmt.Errorf("payroll: dbosflow pool/datasource not wired")
	}
	ready, err := dbos.RunAsStep(ctx, func(ctx context.Context) (readiness, error) {
		return checkReadiness(ctx, Pool, in.EmployerCode, in.PeriodCode)
	}, dbos.WithStepName("readiness"))
	if err != nil {
		return res, err
	}
	cap, err := dbos.RunAsTransaction(ctx, Datasource,
		func(ctx context.Context, tx dbos.Tx) (captureOut, error) {
			return stageBCapture(ctx, tx, wid, in, ready)
		}, dbos.WithStepName("capture"))
	if err != nil {
		return res, err
	}
	calc, err := dbos.RunAsStep(ctx, func(ctx context.Context) (calcOut, error) {
		return calcStep(ctx, Pool, cap, ready, in.EmployerCode)
	}, dbos.WithStepName("calculate"))
	if err != nil {
		return res, err
	}
	actor := "dbos:" + wid + ":" + in.RequestedBy
	outDigest, err := dbos.RunAsTransaction(ctx, Datasource,
		func(ctx context.Context, tx dbos.Tx) (string, error) {
			return stageDPersist(ctx, tx, cap, calc, actor)
		}, dbos.WithStepName("persist"))
	if err != nil {
		return res, err
	}
	timeout := in.ApprovalTimeout
	if timeout <= 0 {
		timeout = DefaultApprovalTimeout
	}
	msg, err := dbos.Recv[ApprovalMsg](ctx, ApprovalTopic(cap.RunID), timeout)
	if err != nil {
		return res, fmt.Errorf("approval wait for run %s: %w", cap.RunID, err)
	}
	if msg.Decision != "approve" {
		return res, fmt.Errorf("run %s rejected by %s: %s", cap.RunID, msg.Approver, msg.Note)
	}
	fin, err := dbos.RunAsTransaction(ctx, Datasource,
		func(ctx context.Context, tx dbos.Tx) (finalizeOut, error) {
			return stageFFinalize(ctx, tx, ready.EmployerID, cap.RunID, actor)
		}, dbos.WithStepName("finalize"))
	if err != nil {
		return res, err
	}
	return PayrollResult{
		WorkflowID: wid, RunID: cap.RunID, SnapshotID: cap.SnapshotID,
		InputDigest: cap.InputDigest, OutputDigest: outDigest,
		PayrollDigest: calc.Digest, Role: cap.Role,
		Employees: len(calc.Month.Employees), Blocked: len(calc.Month.Blocked),
		Lines: calc.LineCount, Validated: true, Enqueue: fin.Enqueue,
	}, nil
}
