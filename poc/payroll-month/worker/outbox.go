// Package worker drains poc_outbox: claim an actionable row, append its
// event to the ledger (idempotent via MustNotExist + digest-bound
// marker), INSERT the PG back-link. Crash-safe at every boundary:
// the row stays actionable until a terminal state commits, and
// duplicates converge through the precondition (see EXPERIMENTS.md
// B2/B3). Recovery reads are VERIFIED (prompt8 §5.4): no back-link
// is ever written from an unverified server response. Poisoning
// policy (POC scope, prompt8 §5.6): attempts are counted, but only
// rows with NO proven append (immu_tx NULL) can go failed/poison —
// a proven append routes to reconciling and stays actionable until
// its back-link is repaired (the guard trigger enforces this even
// against the owner).
package worker

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/PithomLabs/a2/poc/payroll-month/chaos"
	"github.com/PithomLabs/a2/poc/payroll-month/ledger"
)

// ErrNoWork reports an empty pending queue (not a failure).
var ErrNoWork = errors.New("worker: no pending outbox rows")

const maxAttempts = 10

// Payload mirrors outbox.payload JSONB: everything needed to build the
// ledger Record. Written by the authorized enqueue (poc_outbox_enqueue),
// never mutated after (transition guard enforces immutability).
type Payload struct {
	PGTable             string `json:"pg_table"`
	PGID                string `json:"pg_id"`
	EventUUID           string `json:"event_uuid"`
	EmployerID          string `json:"employer_id"`
	InputSnapshotDigest string `json:"input_snapshot_digest"`
	OutputDigest        string `json:"output_digest"`
	RulePins            []struct {
		Pack    string `json:"pack"`
		Version string `json:"version"`
		Digest  string `json:"digest"`
	} `json:"rule_pins"`
	CalcVersion string `json:"calc_version"`
	Approval    struct {
		Approver  string `json:"approver"`
		DecidedAt string `json:"decided_at"`
	} `json:"approval"`
	Supersedes string `json:"supersedes"`
}

// RunOnce processes a single pending row (oldest-first). Safe to call
// concurrently from multiple workers (SKIP LOCKED claims) and to re-run
// after a crash.
func RunOnce(ctx context.Context, pool *pgxpool.Pool, app *ledger.Appender) error {
	return RunOnceOp(ctx, pool, app, "")
}

// RunOnceOp is RunOnce pinned to one op: when wantOp is non-empty only
// that row is claimable. The kill matrix pins its seeded row so a stale
// pending row from another run can neither steal the claim nor let a
// failpoint kill land on the wrong row (silent vacuity). Production
// drains keep oldest-first via RunOnce.
func RunOnceOp(ctx context.Context, pool *pgxpool.Pool, app *ledger.Appender, wantOp string) error {
	// 1. Claim: one actionable row, attempts bumped, lock released on
	//    commit. A crash after commit but before append leaves the row
	//    actionable: re-driven, converging via the ledger precondition
	//    (pending) or verified re-check (appended/reconciling).
	var opID, employer, action, state string
	var payloadJSON []byte
	var payloadHash []byte
	var anchored *int64
	claimTx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("worker: begin claim: %w", err)
	}
	defer claimTx.Rollback(ctx) // no-op after Commit; guards early returns
	claimSQL := `
		SELECT op_uuid, employer_id, action, payload, payload_hash, state, immu_tx
		  FROM poc_outbox
		 WHERE state IN ('pending', 'appended', 'reconciling')
		 ORDER BY created_at LIMIT 1
		 FOR UPDATE SKIP LOCKED`
	claimArgs := []any{}
	if wantOp != "" {
		claimSQL = `
			SELECT op_uuid, employer_id, action, payload, payload_hash, state, immu_tx
			  FROM poc_outbox
			 WHERE state IN ('pending', 'appended', 'reconciling')
			   AND op_uuid = $1::uuid
			 FOR UPDATE SKIP LOCKED`
		claimArgs = append(claimArgs, wantOp)
	}
	row := claimTx.QueryRow(ctx, claimSQL, claimArgs...)
	if err := row.Scan(&opID, &employer, &action, &payloadJSON, &payloadHash, &state, &anchored); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNoWork
		}
		return fmt.Errorf("worker: claim: %w", err)
	}
	if _, err := claimTx.Exec(ctx,
		`UPDATE poc_outbox SET attempts = attempts + 1, updated_at = now()
		  WHERE op_uuid = $1`, opID); err != nil {
		return fmt.Errorf("worker: bump attempts: %w", err)
	}
	var attempts int
	if err := claimTx.QueryRow(ctx,
		`SELECT attempts FROM poc_outbox WHERE op_uuid = $1`, opID).Scan(&attempts); err != nil {
		return fmt.Errorf("worker: read attempts: %w", err)
	}
	// Poison gate: ONLY rows with no proven append may go failed.
	// A proven append (immu_tx set) routes to the reconcile path
	// below no matter how high the counter stands (prompt8 §5.6;
	// the guard trigger fails closed if this ever regresses).
	if attempts > maxAttempts && anchored == nil {
		if _, err := claimTx.Exec(ctx,
			`UPDATE poc_outbox SET state = 'failed', updated_at = now()
			  WHERE op_uuid = $1`, opID); err != nil {
			return fmt.Errorf("worker: mark failed: %w", err)
		}
		if err := claimTx.Commit(ctx); err != nil {
			return fmt.Errorf("worker: commit poisoned: %w", err)
		}
		return fmt.Errorf("worker: op %s failed after %d attempts with no proven append (marked failed)", opID, attempts)
	}
	if err := claimTx.Commit(ctx); err != nil {
		return fmt.Errorf("worker: commit claim: %w", err)
	}

	// 2. Build the expected record from the immutable payload.
	var p Payload
	if err := json.Unmarshal(payloadJSON, &p); err != nil {
		return fmt.Errorf("worker: bad payload for op %s: %w", opID, err)
	}
	rec := buildRecord(employer, action, payloadHash, p)
	idemKey := ledger.IdemKey(employer, action, opID)
	provKey := ledger.ProvKey(rec.EmployerID, rec.Artifact, rec.ArtifactID, rec.EventUUID)

	var txID uint64
	curState := state
	if anchored == nil {
		// 3a. Append path (pending): B2 window, append (or verified
		// recover on precondition), B3 window, mark appended, B4
		// window, then the shared back-link tail.
		chaos.Trip("after-pg-commit-before-append") // B2 (EXPERIMENTS.md)
		txID, err = app.Append(ctx, rec, idemKey)
		if err != nil {
			if status.Code(err) == codes.FailedPrecondition ||
				status.Code(errors.Unwrap(err)) == codes.FailedPrecondition {
				txID, err = recoverTxVerified(ctx, app, idemKey, provKey, rec)
				if err != nil {
					return fmt.Errorf("worker: recover op %s: %w", opID, err)
				}
			} else {
				return fmt.Errorf("worker: append op %s: %w", opID, err)
			}
		}

		chaos.Trip("after-append-before-backlink") // B3 (EXPERIMENTS.md)

		appendTx, err := pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("worker: begin appended: %w", err)
		}
		defer func() { _ = appendTx.Rollback(ctx) }()
		tag, err := appendTx.Exec(ctx, `
			UPDATE poc_outbox
			   SET state = 'appended', immu_tx = $2, updated_at = now()
			 WHERE op_uuid = $1 AND state = 'pending'`, opID, txID)
		if err != nil {
			return fmt.Errorf("worker: mark appended: %w", err)
		}
		if tag.RowsAffected() == 1 {
			if err := appendTx.Commit(ctx); err != nil {
				return fmt.Errorf("worker: commit appended: %w", err)
			}
			curState = "appended"
		} else {
			// Restart path: another drive already marked it; verify
			// the anchored tx matches before proceeding.
			var st string
			var have uint64
			if err := appendTx.QueryRow(ctx,
				`SELECT state, immu_tx FROM poc_outbox WHERE op_uuid = $1`,
				opID).Scan(&st, &have); err != nil {
				return fmt.Errorf("worker: reread appended: %w", err)
			}
			if (st != "appended" && st != "reconciling") || have != txID {
				return fmt.Errorf("worker: op %s state %s tx %d diverges from append tx %d",
					opID, st, have, txID)
			}
			curState = st
		}

		chaos.Trip("after-appended-before-backlink") // B4 (reconcile battery)
	} else {
		// 3b. Reconcile path (appended/reconciling): the append is
		// proven; re-verify the anchor (tx + marker + record bytes
		// at that tx) before touching the back-link. Never
		// re-appends, never poisons.
		txID = uint64(*anchored)
		if err := reconcileTxVerified(ctx, app, idemKey, provKey, rec, txID); err != nil {
			return fmt.Errorf("worker: reconcile op %s: %w", opID, err)
		}
	}

	// 4. Back-link txn via the authorized recorder (squat-proof: the
	// outbox row must pre-exist with a matching tx; exact replays
	// converge). Duplicate convergence is ordinary success, not error.
	// A link FAILURE with a proven append moves appended ->
	// reconciling (stays actionable); it never discards.
	backTx, err := pool.Begin(ctx)
	if err != nil {
		return markReconciling(ctx, pool, opID, curState,
			fmt.Errorf("worker: begin backlink: %w", err))
	}
	defer func() { _ = backTx.Rollback(ctx) }()
	var linked string
	if err := backTx.QueryRow(ctx,
		`SELECT poc_backlink_record($1, $2, $3, $4, $5)`,
		p.PGTable, p.PGID, p.EventUUID, txID, []byte(rec.Digest())).Scan(&linked); err != nil {
		_ = backTx.Rollback(ctx)
		return markReconciling(ctx, pool, opID, curState,
			fmt.Errorf("worker: backlink record: %w", err))
	}
	if linked != "linked" && linked != "replayed" {
		_ = backTx.Rollback(ctx)
		return markReconciling(ctx, pool, opID, curState,
			fmt.Errorf("worker: backlink record returned %q", linked))
	}
	// NOTE: alh stores the content digest (re-verification anchor is the tx
	// id via VerifiedTxByID, not this column). See ADR-0204.
	tag, err := backTx.Exec(ctx, `
		UPDATE poc_outbox
		   SET state = CASE WHEN state = 'reconciling' THEN 'reconciled' ELSE 'backlinked' END,
		       updated_at = now()
		 WHERE op_uuid = $1 AND state IN ('appended', 'reconciling')`, opID)
	if err != nil {
		_ = backTx.Rollback(ctx)
		return markReconciling(ctx, pool, opID, curState,
			fmt.Errorf("worker: outbox transition: %w", err))
	}
	if tag.RowsAffected() != 1 {
		var st string
		if err := backTx.QueryRow(ctx,
			`SELECT state FROM poc_outbox WHERE op_uuid = $1`, opID).Scan(&st); err != nil {
			_ = backTx.Rollback(ctx)
			return markReconciling(ctx, pool, opID, curState,
				fmt.Errorf("worker: reread terminal: %w", err))
		}
		if st != "backlinked" && st != "reconciled" {
			_ = backTx.Rollback(ctx)
			return markReconciling(ctx, pool, opID, curState,
				fmt.Errorf("worker: op %s in unexpected state %s", opID, st))
		}
		// Concurrent duplicate already converged: ordinary success.
	}
	if err := backTx.Commit(ctx); err != nil {
		return markReconciling(ctx, pool, opID, curState,
			fmt.Errorf("worker: commit backlink: %w", err))
	}
	return nil
}

// buildRecord reconstructs the expected ledger record from the immutable
// outbox payload. The auditor rebuilds the same bytes independently
// from PostgreSQL (prompt8 §5.2); any divergence fails loudly.
// CalcDigest binds the pinned calc_version LABEL (sha256), not
// implementation bytes (see Record provenance notes).
func buildRecord(employer, action string, payloadHash []byte, p Payload) ledger.Record {
	ruleVersions := map[string]string{}
	packDigest := ""
	for _, pin := range p.RulePins {
		ruleVersions[pin.Pack] = pin.Version
		if packDigest == "" {
			packDigest = pin.Digest // first pin; multi-pin joint digest is future work
		}
	}
	calcSum := sha256.Sum256([]byte(p.CalcVersion))
	return ledger.Record{
		EventUUID: p.EventUUID, Artifact: action, ArtifactID: p.PGID,
		PGTable: p.PGTable, PGID: p.PGID, EmployerID: employer,
		SHA256:      hex.EncodeToString(payloadHash),
		InputDigest: p.InputSnapshotDigest, OutputDigest: p.OutputDigest,
		RuleVersions: ruleVersions,
		PackDigest:   packDigest, CalcDigest: hex.EncodeToString(calcSum[:]),
		DecidedBy: p.Approval.Approver, DecidedAt: p.Approval.DecidedAt,
		Supersedes: p.Supersedes,
	}
}

// markReconciling records a back-link failure distinctly: appended ->
// reconciling (stays actionable for the next drive). Rows already in
// reconciling are left there. The original error is preserved; the
// move itself failing is attached, never swallowed.
func markReconciling(ctx context.Context, pool *pgxpool.Pool, opID, cur string, cause error) error {
	if cur != "appended" {
		return cause
	}
	if _, err := pool.Exec(ctx,
		`UPDATE poc_outbox SET state = 'reconciling', updated_at = now()
		  WHERE op_uuid = $1 AND state = 'appended'`, opID); err != nil {
		return fmt.Errorf("%v (reconciling move failed: %v)", cause, err)
	}
	return fmt.Errorf("%v (moved to reconciling)", cause)
}

// verifyAnchor binds a claimed tx to the expected record through
// verified reads only: the tx is part of attested history, and the
// record bytes that tx committed digest-match the expectation and
// belong to the intended artifact.
func verifyAnchor(ctx context.Context, app *ledger.Appender, provKey []byte, rec ledger.Record, tx uint64) error {
	if tx == 0 {
		return fmt.Errorf("anchor tx unknown")
	}
	if err := app.VerifiedTx(ctx, tx); err != nil {
		return fmt.Errorf("anchor tx: %w", err)
	}
	got, err := app.VerifiedReadRecordAt(ctx, provKey, tx)
	if err != nil {
		return fmt.Errorf("anchor record: %w", err)
	}
	if got.Digest() != rec.Digest() {
		return fmt.Errorf("anchor record digest %s != expected %s",
			got.Digest(), rec.Digest())
	}
	if got.EventUUID != rec.EventUUID || got.Artifact != rec.Artifact ||
		got.PGTable != rec.PGTable || got.PGID != rec.PGID {
		return fmt.Errorf("anchor record binds %s/%s event %s, want %s/%s event %s",
			got.PGTable, got.PGID, got.EventUUID, rec.PGTable, rec.PGID, rec.EventUUID)
	}
	return nil
}

// recoverTxVerified implements the precondition-hit path (§5.4): the
// marker is re-read VERIFIED and digest-matched (aliasing fails
// loudly); the origin tx is re-verified and bound to the exact
// record bytes it committed. Requires the writer never
// deletes/expires idem keys (ADR-0204).
func recoverTxVerified(ctx context.Context, app *ledger.Appender, idemKey, provKey []byte, rec ledger.Record) (uint64, error) {
	value, tx, err := app.VerifiedReadMarker(ctx, idemKey)
	if err != nil {
		return 0, fmt.Errorf("verified marker read: %w", err)
	}
	if want := rec.EventUUID + ":" + rec.Digest(); value != want {
		return 0, fmt.Errorf("payload aliasing: marker %q != %q", value, want)
	}
	if err := verifyAnchor(ctx, app, provKey, rec, tx); err != nil {
		return 0, err
	}
	return tx, nil
}

// reconcileTxVerified re-verifies a claimed anchor before repairing
// its back-link (§5.4/§5.6): tx attested, marker verified and bound
// to this op with the origin tx matching the claim, record bytes at
// that tx digest-matched and artifact-bound.
func reconcileTxVerified(ctx context.Context, app *ledger.Appender, idemKey, provKey []byte, rec ledger.Record, tx uint64) error {
	value, origin, err := app.VerifiedReadMarker(ctx, idemKey)
	if err != nil {
		return fmt.Errorf("verified marker read: %w", err)
	}
	if want := rec.EventUUID + ":" + rec.Digest(); value != want {
		return fmt.Errorf("payload aliasing: marker %q != %q", value, want)
	}
	if origin != tx {
		return fmt.Errorf("marker origin tx %d != claimed anchor %d", origin, tx)
	}
	return verifyAnchor(ctx, app, provKey, rec, tx)
}
