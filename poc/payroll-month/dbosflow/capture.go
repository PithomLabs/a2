package dbosflow

// CaptureWorkflow runs stages A+B only: readiness + capture-or-adopt.
// Used by `capture-once` (kill-matrix B1, dedup, reproducibility) and
// shared by the full PayrollWorkflow (identical Stage B code path).

import (
	"context"
	"fmt"

	"github.com/dbos-inc/dbos-transact-golang/dbos"
)

// CaptureInput identifies one capture attempt.
type CaptureInput struct {
	EmployerCode string `json:"employer_code"`
	PeriodCode   string `json:"period_code"`
	RequestedBy  string `json:"requested_by"`
}

// CaptureResult is the checkpointed capture outcome.
type CaptureResult struct {
	WorkflowID  string `json:"workflow_id"`
	RunID       string `json:"run_id"`
	SnapshotID  string `json:"snapshot_id"`
	InputDigest string `json:"input_digest"`
	Role        string `json:"role"` // captured | adopted
}

// CaptureWorkflow executes readiness then the capture transaction.
func CaptureWorkflow(ctx dbos.Context, in CaptureInput) (CaptureResult, error) {
	var res CaptureResult
	wid, err := dbos.GetWorkflowID(ctx)
	if err != nil {
		return res, err
	}
	if in.EmployerCode == "" || in.PeriodCode == "" || in.RequestedBy == "" {
		return res, fmt.Errorf("capture: employer, period and requester are required")
	}
	if Pool == nil || Datasource == nil {
		return res, fmt.Errorf("capture: dbosflow pool/datasource not wired")
	}
	ready, err := dbos.RunAsStep(ctx, func(ctx context.Context) (readiness, error) {
		return checkReadiness(ctx, Pool, in.EmployerCode, in.PeriodCode)
	}, dbos.WithStepName("readiness"))
	if err != nil {
		return res, err
	}
	cap, err := dbos.RunAsTransaction(ctx, Datasource,
		func(ctx context.Context, tx dbos.Tx) (captureOut, error) {
			return stageBCapture(ctx, tx, wid, PayrollInput{
				EmployerCode: in.EmployerCode, PeriodCode: in.PeriodCode,
				RequestedBy: in.RequestedBy,
			}, ready)
		}, dbos.WithStepName("capture"))
	if err != nil {
		return res, err
	}
	return CaptureResult{
		WorkflowID: wid, RunID: cap.RunID, SnapshotID: cap.SnapshotID,
		InputDigest: cap.InputDigest, Role: cap.Role,
	}, nil
}
