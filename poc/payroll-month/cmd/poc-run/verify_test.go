package main

// Unit collateral for the submitter-side verification (DB-5 second
// layer): results the database cannot corroborate fail closed.
// Requires the POC stack (reads real business rows); skips loudly.

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/PithomLabs/a2/poc/payroll-month/dbosflow"
)

func TestVerifyCaptureResult(t *testing.T) {
	dsn := os.Getenv("POC_DATABASE_URL")
	if dsn == "" {
		t.Skip("POC_DATABASE_URL unset: needs the POC stack")
	}
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer pool.Close()
	ctx := context.Background()

	// Forged result fails closed.
	fake := dbosflow.CaptureResult{
		WorkflowID: "capture:verify-fake", RunID: "ffffffff-ffff-4fff-8fff-ffffffffffff",
		SnapshotID:  "ffffffff-ffff-4fff-8fff-fffffffffffe",
		InputDigest: "FORGED", Role: "captured",
	}
	if err := verifyCaptureResult(ctx, pool, fake); err == nil {
		t.Fatal("forged capture result verified, want failure")
	}
	// Genuine result verifies (if a capture exists on this database).
	var runID, snapID, digest string
	err = pool.QueryRow(ctx,
		`SELECT r.id, s.id, s.content_hash FROM payroll_run r
		  JOIN payroll_snapshot s ON s.payroll_run_id = r.id AND s.status = 'valid'
		 LIMIT 1`).Scan(&runID, &snapID, &digest)
	if err != nil {
		t.Skip("no captured run on this database")
	}
	genuine := dbosflow.CaptureResult{
		WorkflowID: "capture:verify-genuine", RunID: runID,
		SnapshotID: snapID, InputDigest: digest, Role: "adopted",
	}
	if err := verifyCaptureResult(ctx, pool, genuine); err != nil {
		t.Fatalf("genuine capture result rejected: %v", err)
	}
	// Right IDs, wrong digest fails closed.
	tampered := genuine
	tampered.InputDigest = "TAMPERED"
	if err := verifyCaptureResult(ctx, pool, tampered); err == nil {
		t.Fatal("tampered digest verified, want failure")
	}
}
