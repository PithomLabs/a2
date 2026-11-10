// Command poc-run is the single-command POC runner: preflight checks, the
// deterministic synthetic demo month (always runs, no services needed), and
// — when POC_DATABASE_URL is set — the DBOS month workflow on the shared-pool
// path. It prints a JSON report to stdout and exits non-zero unless every
// required check passes. Subcommand "migrate" runs DBOS system-table DDL
// (privileged URL) without running the month.
package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/codenotary/immudb/pkg/auth"
	"github.com/codenotary/immudb/pkg/client"
	"github.com/dbos-inc/dbos-transact-golang/dbos"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/PithomLabs/a2/poc/payroll-month/dbosflow"
	"github.com/PithomLabs/a2/poc/payroll-month/ledger"
	"github.com/PithomLabs/a2/poc/payroll-month/seedpoc"
	"github.com/PithomLabs/a2/poc/payroll-month/synth"
	"github.com/PithomLabs/a2/poc/payroll-month/worker"
	"github.com/PithomLabs/a2/poc/payroll-month/workload"
)

// Check is one report line. Required=false checks are informational.
type Check struct {
	Name     string `json:"name"`
	Required bool   `json:"required"`
	Pass     bool   `json:"pass"`
	Detail   string `json:"detail"`
}

// Report is the machine-readable run outcome.
type Report struct {
	Mode    string         `json:"mode"`
	Checks  []Check        `json:"checks"`
	Data    map[string]any `json:"data"`
	Overall string         `json:"overall"` // "pass" or "fail"
}

func main() {
	os.Exit(run())
}

func run() int {
	ctx := context.Background()
	rep := &Report{Data: map[string]any{}}
	add := func(name string, required, pass bool, detail string) {
		rep.Checks = append(rep.Checks, Check{name, required, pass, detail})
	}

	if len(os.Args) > 1 && os.Args[1] == "migrate" {
		return migrateCmd(ctx, rep, add)
	}
	if len(os.Args) > 1 && os.Args[1] == "outbox-once" {
		return outboxOnceCmd(ctx, rep, add)
	}
	if len(os.Args) > 1 && os.Args[1] == "capture-once" {
		return captureOnceCmd(ctx, rep, add)
	}
	if len(os.Args) > 1 && os.Args[1] == "approve" {
		return approveCmd(ctx, rep, add)
	}
	if len(os.Args) > 1 && os.Args[1] == "recover" {
		return recoverCmd(ctx, rep, add)
	}
	if len(os.Args) > 1 && os.Args[1] == "canonical" {
		return canonicalCmd()
	}
	if len(os.Args) > 1 && os.Args[1] == "seed-poc" {
		return seedPocCmd(ctx)
	}
	if len(os.Args) > 1 && os.Args[1] == "immudb-users" {
		return immudbUsersCmd(ctx, rep, add)
	}
	demoOnly := len(os.Args) > 1 && os.Args[1] == "--demo-synth-only"

	// 1. Synthetic demo month: always runs, proves determinism without services.
	demoDigest, employees, days, ot, ded, err := demoMonth()
	if err != nil {
		add("synth-demo-month", true, false, err.Error())
	} else {
		rep.Data["demo_digest"] = demoDigest
		rep.Data["demo_employees"] = employees
		rep.Data["demo_days"] = days
		rep.Data["demo_ot_cents"] = ot
		rep.Data["demo_ded_cents"] = ded
		add("synth-demo-month", true, true,
			fmt.Sprintf("%d employees x %d days, digest %s", employees, days, demoDigest[:16]))
	}

	if demoOnly {
		rep.Mode = "demo-synth-only"
		add("mode", false, true, "demo only by explicit flag: DB/database checks skipped")
		return finish(rep)
	}

	// 2. Full mode: DBOS month workflow, database URL required.
	dsn := os.Getenv("POC_DATABASE_URL")
	if dsn == "" {
		rep.Mode = "full-month"
		add("env:POC_DATABASE_URL", true, false, "unset: full month needs the PG18.6 POC stack")
		return finish(rep)
	}
	rep.Mode = "full-month"
	if err := runMonthWorkflow(ctx, dsn, rep); err != nil {
		add("dbos-month-workflow", true, false, err.Error())
	} else {
		add("dbos-month-workflow", true, true, "month workflow completed")
	}
	return finish(rep)
}

func finish(rep *Report) int {
	rep.Overall = "pass"
	for _, c := range rep.Checks {
		if c.Required && !c.Pass {
			rep.Overall = "fail"
		}
	}
	raw, _ := json.MarshalIndent(rep, "", "  ")
	fmt.Println(string(raw))
	if rep.Overall == "pass" {
		return 0
	}
	return 1
}

// migrateCmd runs DBOS system-table DDL against POC_DBOS_MIGRATE_URL
// (privileged role). The app role later runs with SkipMigrations.
func migrateCmd(ctx context.Context, rep *Report, add func(string, bool, bool, string)) int {
	rep.Mode = "migrate"
	url := os.Getenv("POC_DBOS_MIGRATE_URL")
	if url == "" {
		add("env:POC_DBOS_MIGRATE_URL", true, false, "unset: privileged DBOS migrate URL required")
		return finish(rep)
	}
	if err := dbos.Migrate(ctx, url); err != nil {
		add("dbos-migrate", true, false, err.Error())
		return finish(rep)
	}
	add("dbos-migrate", true, true, "system tables migrated")
	return finish(rep)
}

// wireDBOS connects the shared pool, builds the executor context + client,
// creates the datasource, and wires it for capture txns. Shared pool (same
// *pgxpool.Pool for system and app) is what makes the capture txn atomic
// (ADR-0203). Caller closes the pool.
func wireDBOS(ctx context.Context, dsn string) (*pgxpool.Pool, dbos.Context, dbos.Client, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, nil, nil, fmt.Errorf("ping: %w", err)
	}
	// App paths (capture-once, full-month) never own the schema: system
	// tables are created once by `poc-run migrate` (privileged role).
	// SkipMigrations verifies instead of migrating (Gate B: app-role
	// migrate attempts fail with permission denied for database).
	cfg := dbos.Config{AppName: "poc-payroll-month", SystemDBPool: pool, SkipMigrations: true}
	// DBOS's default logger writes INFO to STDOUT, interleaving with the
	// JSON report (breaks `poc-run | jq` and byte comparisons — proven
	// at Gate B). Route to stderr at WARN; POC_DBOS_DEBUG restores all.
	dbosLevel := slog.LevelWarn
	if os.Getenv("POC_DBOS_DEBUG") != "" {
		dbosLevel = slog.LevelDebug
	}
	cfg.Logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: dbosLevel}))
	execCtx, err := dbos.NewContext(ctx, cfg)
	if err != nil {
		pool.Close()
		return nil, nil, nil, fmt.Errorf("dbos context: %w", err)
	}
	// Same logger for the client: its default also writes INFO to stdout.
	cli, err := dbos.NewClient(ctx, dbos.ClientConfig{AppName: "poc-payroll-month", SystemDBPool: pool, Logger: cfg.Logger})
	if err != nil {
		pool.Close()
		return nil, nil, nil, fmt.Errorf("dbos client: %w", err)
	}
	ds, err := dbos.NewDataSource(cli, pool)
	if err != nil {
		pool.Close()
		return nil, nil, nil, fmt.Errorf("dbos datasource: %w", err)
	}
	dbosflow.Datasource = ds
	dbosflow.Pool = pool
	return pool, execCtx, cli, nil
}

// payrollWorkflow/captureWorkflow are typed top-level workflow funcs (the DBOS
// documented recommended form: stable derived names). Invoke ONLY via the
// dbos.RunWorkflow package function, which derives the registry name — the
// Context.RunWorkflow method carries no name and fails NonExistentWorkflow
// (proven at Gate B; the earlier "execCtx.RunWorkflow fix" compiled but
// could never enqueue).
func payrollWorkflow(c dbos.Context, in dbosflow.PayrollInput) (dbosflow.PayrollResult, error) {
	return dbosflow.PayrollWorkflow(c, in)
}

func captureWorkflow(c dbos.Context, in dbosflow.CaptureInput) (dbosflow.CaptureResult, error) {
	return dbosflow.CaptureWorkflow(c, in)
}

// runMonthWorkflow invokes one payroll attempt with a UNIQUE workflow ID
// (Reject on reuse: duplicate submissions fail closed, loudly).
// --approval-timeout (Go duration, default 30m) bounds the Stage E wait.
func runMonthWorkflow(ctx context.Context, dsn string, rep *Report) error {
	pool, execCtx, cli, err := wireDBOS(ctx, dsn)
	if err != nil {
		return err
	}
	defer dbosflow.ClosePoolTimeout(pool, 15*time.Second)
	defer func() { _ = execCtx.Shutdown(cli, 10*time.Second) }()
	dbos.RegisterWorkflow(execCtx, payrollWorkflow)
	if err := execCtx.Launch(); err != nil {
		return fmt.Errorf("dbos launch: %w", err)
	}
	employer := getenv("POC_EMPLOYER", "POC-MERIDIAN")
	period := getenv("POC_PERIOD", "PP-2026-10")
	timeout := dbosflow.DefaultApprovalTimeout
	if raw := flagValue("--approval-timeout", ""); raw != "" {
		if d, err := time.ParseDuration(raw); err == nil {
			timeout = d
		} else {
			return fmt.Errorf("bad --approval-timeout %q: %w", raw, err)
		}
	}
	wid := fmt.Sprintf("payrun:%s:%s:%d", employer, period, time.Now().UnixNano())
	handle, err := dbos.RunWorkflow(execCtx, payrollWorkflow,
		dbosflow.PayrollInput{EmployerCode: employer, PeriodCode: period,
			RequestedBy: "poc-run", ApprovalTimeout: timeout},
		dbos.WithWorkflowID(wid),
		dbos.WithWorkflowIDReusePolicy(dbos.WorkflowIDReusePolicyReject))
	if err != nil {
		return fmt.Errorf("run workflow: %w", err)
	}
	res, err := handle.GetResult()
	if err != nil {
		return fmt.Errorf("workflow result: %w", err)
	}
	// Submitter-side verification (DB-5 fail-closed): a returned result
	// is trusted only when the persisted business state matches it.
	if err := verifyPayrollResult(ctx, pool, res); err != nil {
		return fmt.Errorf("result verification: %w", err)
	}
	rep.Data["month_workflow_id"] = res.WorkflowID
	rep.Data["payroll_run_id"] = res.RunID
	rep.Data["input_digest"] = res.InputDigest
	rep.Data["output_digest"] = res.OutputDigest
	rep.Data["payroll_digest"] = res.PayrollDigest
	rep.Data["role"] = res.Role
	return nil
}

// verifyPayrollResult re-reads the persisted business state and requires
// it to match the returned workflow result (run validated, tip and
// output digests equal). A forged SUCCESS row returns output the
// database cannot corroborate — and fails here, loudly.
func verifyPayrollResult(ctx context.Context, pool *pgxpool.Pool, res dbosflow.PayrollResult) error {
	var status, tipHash, outDigest string
	if err := pool.QueryRow(ctx,
		`SELECT r.status, s.content_hash, o.output_digest
		   FROM payroll_run r
		   JOIN payroll_snapshot s ON s.payroll_run_id = r.id AND s.status = 'valid'
		   JOIN poc_run_output o ON o.run_id = r.id
		  WHERE r.id = $1`, res.RunID).Scan(&status, &tipHash, &outDigest); err != nil {
		return fmt.Errorf("run %s not corroborated: %w", res.RunID, err)
	}
	if status != "validated" {
		return fmt.Errorf("run %s status %s, want validated", res.RunID, status)
	}
	if tipHash != res.InputDigest || outDigest != res.OutputDigest {
		return fmt.Errorf("run %s digest mismatch (db %s/%s vs result %s/%s)",
			res.RunID, tipHash[:16], outDigest[:16], res.InputDigest[:16], res.OutputDigest[:16])
	}
	return nil
}

// captureOnceCmd invokes one capture attempt (kill-matrix + smoke use).
// --run-id names the attempt (workflow ID "capture:"+runID, Reject on
// reuse); employer/period from POC_EMPLOYER/POC_PERIOD (real codes).
func captureOnceCmd(ctx context.Context, rep *Report, add func(string, bool, bool, string)) int {
	rep.Mode = "capture-once"
	dsn := os.Getenv("POC_DATABASE_URL")
	if dsn == "" {
		add("env:POC_DATABASE_URL", true, false, "unset")
		return finish(rep)
	}
	runID := flagValue("--run-id", "default")
	employer := getenv("POC_EMPLOYER", "POC-MERIDIAN")
	period := getenv("POC_PERIOD", "PP-2026-10")
	pool, execCtx, cli, err := wireDBOS(ctx, dsn)
	if err != nil {
		add("dbos-wire", true, false, err.Error())
		return finish(rep)
	}
	defer dbosflow.ClosePoolTimeout(pool, 15*time.Second)
	defer func() { _ = execCtx.Shutdown(cli, 10*time.Second) }()
	dbos.RegisterWorkflow(execCtx, captureWorkflow)
	if err := execCtx.Launch(); err != nil {
		add("dbos-launch", true, false, err.Error())
		return finish(rep)
	}
	handle, err := dbos.RunWorkflow(execCtx, captureWorkflow,
		dbosflow.CaptureInput{EmployerCode: employer, PeriodCode: period, RequestedBy: "poc-run"},
		dbos.WithWorkflowID("capture:"+runID),
		dbos.WithWorkflowIDReusePolicy(dbos.WorkflowIDReusePolicyReject))
	if err != nil {
		add("capture-run", true, false, err.Error())
		return finish(rep)
	}
	res, err := handle.GetResult()
	if err != nil {
		add("capture-result", true, false, err.Error())
		return finish(rep)
	}
	if err := verifyCaptureResult(ctx, pool, res); err != nil {
		add("capture-verify", true, false, err.Error())
		return finish(rep)
	}
	rep.Data["capture_workflow_id"] = res.WorkflowID
	rep.Data["capture_run_id"] = res.RunID
	rep.Data["capture_inputs_hash"] = res.InputDigest
	rep.Data["capture_role"] = res.Role
	add("capture-run", true, true, "workflow "+res.WorkflowID)
	return finish(rep)
}

// verifyCaptureResult corroborates a capture result against persisted rows.
func verifyCaptureResult(ctx context.Context, pool *pgxpool.Pool, res dbosflow.CaptureResult) error {
	var tipHash string
	if err := pool.QueryRow(ctx,
		`SELECT content_hash FROM payroll_snapshot
		  WHERE id = $1 AND payroll_run_id = $2 AND status = 'valid'`,
		res.SnapshotID, res.RunID).Scan(&tipHash); err != nil {
		return fmt.Errorf("snapshot %s not corroborated: %w", res.SnapshotID, err)
	}
	if tipHash != res.InputDigest {
		return fmt.Errorf("snapshot %s digest mismatch", res.SnapshotID)
	}
	return nil
}

// recoverCmd starts a recovery-only executor: Launch() recovers PENDING
// workflows of this app, then it idles until --for (default 60s) so
// recovery completes. Operations use for post-crash recovery; the
// approval-crash test uses it to resume a killed waiter.
func recoverCmd(ctx context.Context, rep *Report, add func(string, bool, bool, string)) int {
	rep.Mode = "recover"
	dsn := os.Getenv("POC_DATABASE_URL")
	if dsn == "" {
		add("env:POC_DATABASE_URL", true, false, "unset")
		return finish(rep)
	}
	dur := 60 * time.Second
	if raw := flagValue("--for", ""); raw != "" {
		if d, err := time.ParseDuration(raw); err == nil {
			dur = d
		} else {
			add("args", true, false, "bad --for "+raw)
			return finish(rep)
		}
	}
	pool, execCtx, cli, err := wireDBOS(ctx, dsn)
	if err != nil {
		add("dbos-wire", true, false, err.Error())
		return finish(rep)
	}
	defer dbosflow.ClosePoolTimeout(pool, 15*time.Second)
	defer func() { _ = execCtx.Shutdown(cli, 10*time.Second) }()
	dbos.RegisterWorkflow(execCtx, payrollWorkflow)
	dbos.RegisterWorkflow(execCtx, captureWorkflow)
	if err := execCtx.Launch(); err != nil {
		add("dbos-launch", true, false, err.Error())
		return finish(rep)
	}
	var pending int
	if err := pool.QueryRow(ctx,
		`SELECT count(*) FROM dbos.workflow_status WHERE status = 'PENDING'`).Scan(&pending); err != nil {
		add("pending-probe", true, false, err.Error())
		return finish(rep)
	}
	rep.Data["pending_at_start"] = pending
	// Watch the queue until it drains or the window elapses; exit
	// early once PENDING reaches zero twice in a row (settle).
	settled := 0
	deadline := time.Now().Add(dur)
	for time.Now().Before(deadline) {
		time.Sleep(5 * time.Second)
		var cur int
		if err := pool.QueryRow(ctx,
			`SELECT count(*) FROM dbos.workflow_status WHERE status = 'PENDING'`).Scan(&cur); err != nil {
			add("pending-probe", true, false, err.Error())
			return finish(rep)
		}
		if cur == 0 {
			settled++
			if settled >= 2 {
				add("recover", true, true, "queue drained")
				return finish(rep)
			}
		} else {
			settled = 0
		}
	}
	add("recover", true, true, "recovery window elapsed")
	return finish(rep)
}

// approveCmd records an approval decision (authorized DEFINER op) and
// wakes the waiting workflow attempt via DBOS Send. Convergent: if the
// decision already stands, it re-sends the wake-up instead of failing.
func approveCmd(ctx context.Context, rep *Report, add func(string, bool, bool, string)) int {
	rep.Mode = "approve"
	dsn := os.Getenv("POC_DATABASE_URL")
	if dsn == "" {
		add("env:POC_DATABASE_URL", true, false, "unset")
		return finish(rep)
	}
	runID := flagValue("--run-id", "")
	decision := flagValue("--decision", "")
	approver := flagValue("--approver", "poc-operator")
	note := flagValue("--note", "")
	if runID == "" || (decision != "approve" && decision != "reject") {
		add("args", true, false, "need --run-id R --decision approve|reject [--approver A --note N]")
		return finish(rep)
	}
	pool, _, cli, err := wireDBOS(ctx, dsn)
	if err != nil {
		add("dbos-wire", true, false, err.Error())
		return finish(rep)
	}
	defer dbosflow.ClosePoolTimeout(pool, 15*time.Second)
	// No Launch/Shutdown: approve is a client-only command (record +
	// Send). Starting an executor here could claim the very workflow
	// it tries to wake (recovery interference, observed live).
	// Resolve the waiting attempt (the Recv destination).
	var wid string
	if err := pool.QueryRow(ctx,
		`SELECT a.workflow_id FROM poc_workflow_attempt a
		   JOIN dbos.workflow_status s ON s.workflow_uuid = a.workflow_id
		  WHERE a.run_id = $1 AND s.status = 'PENDING'
		  ORDER BY a.created_at DESC LIMIT 1`, runID).Scan(&wid); err != nil {
		add("resolve-waiter", true, false,
			fmt.Sprintf("run %s: %v", runID, err))
		return finish(rep)
	}
	var dec string
	err = pool.QueryRow(ctx,
		`SELECT poc_approve($1, $2, $3, $4)`, runID, approver, decision, note).Scan(&dec)
	if err != nil {
		// Convergent re-send: the decision may already stand (draft
		// runs raise 'already decided'; validated runs raise 'want
		// draft'). Either way, wake the waiter only if the STORED
		// decision matches the requested one — never upgrade a
		// recorded reject to an approve wake-up.
		if strings.Contains(err.Error(), "already decided") ||
			strings.Contains(err.Error(), "want draft") {
			var stored string
			if qerr := pool.QueryRow(ctx,
				`SELECT decision FROM poc_approval WHERE run_id = $1`, runID).Scan(&stored); qerr != nil {
				add("approve-record", true, false, "no recorded decision: "+err.Error())
				return finish(rep)
			} else if stored != decision {
				add("approve-record", true, false,
					fmt.Sprintf("recorded decision is %q, requested %q", stored, decision))
				return finish(rep)
			}
			rep.Data["decision"] = stored
			add("approve-record", false, true, "decision "+stored+" already stands; re-sending wake-up")
		} else {
			add("approve-record", true, false, err.Error())
			return finish(rep)
		}
	} else {
		rep.Data["decision"] = dec
		add("approve-record", true, true, "decision "+dec+" by "+approver)
	}
	if err := dbos.Send(cli, wid,
		dbosflow.ApprovalMsg{Decision: decision, Approver: approver, Note: note},
		dbosflow.ApprovalTopic(runID)); err != nil {
		add("approve-send", true, false, err.Error())
		return finish(rep)
	}
	rep.Data["waiting_workflow"] = wid
	add("approve-send", true, true, "woke "+wid)
	return finish(rep)
}

// immudbUsersCmd provisions the least-privilege ledger identities
// (prompt8 §5.3): poc_writer (read-write) for the outbox drain and
// poc_auditor (read-only) for the independent auditor. Runs with the
// admin bootstrap credential; idempotent (existing users are kept,
// never reset). Run once per immudb volume (runbook step).
func immudbUsersCmd(ctx context.Context, rep *Report, add func(string, bool, bool, string)) int {
	rep.Mode = "immudb-users"
	adminPW := os.Getenv("POC_IMMUDB_ADMIN_PASSWORD")
	writerUser := getenv("POC_IMMUDB_USER", "poc_writer")
	writerPW := os.Getenv("POC_IMMUDB_PASSWORD")
	auditUser := getenv("POC_AUDIT_IMMUDB_USER", "poc_auditor")
	auditPW := os.Getenv("POC_AUDIT_IMMUDB_PASSWORD")
	db := getenv("POC_IMMUDB_DB", "defaultdb")
	for _, kv := range [][2]string{
		{"POC_IMMUDB_ADMIN_PASSWORD", adminPW},
		{"POC_IMMUDB_PASSWORD", writerPW},
		{"POC_AUDIT_IMMUDB_PASSWORD", auditPW},
	} {
		if kv[1] == "" {
			add("env:"+kv[0], true, false, "unset")
			return finish(rep)
		}
	}
	if writerUser == "immudb" || auditUser == "immudb" {
		add("users", true, false, "writer/auditor must not be the superuser")
		return finish(rep)
	}
	if writerUser == auditUser {
		add("users", true, false, "writer and auditor must be distinct users")
		return finish(rep)
	}
	host, port := splitHostPort(getenv("POC_IMMUDB_ADDR", "localhost:3322"), 3322)
	opts := client.DefaultOptions()
	opts.Address, opts.Port = host, port
	opts.DialOptions = []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	cli := client.NewClient().WithOptions(opts)
	if err := cli.OpenSession(ctx, []byte("immudb"), []byte(adminPW), db); err != nil {
		add("admin-session", true, false, err.Error())
		return finish(rep)
	}
	defer cli.CloseSession(ctx)
	for _, u := range []struct {
		name string
		pass string
		perm uint32
	}{
		{writerUser, writerPW, auth.PermissionRW},
		{auditUser, auditPW, auth.PermissionR},
	} {
		if err := cli.CreateUser(ctx, []byte(u.name), []byte(u.pass), u.perm, db); err != nil {
			if strings.Contains(strings.ToLower(err.Error()), "already exists") ||
				strings.Contains(strings.ToLower(err.Error()), "alreadyexists") {
				add("user:"+u.name, true, true, "exists, kept")
				continue
			}
			add("user:"+u.name, true, false, err.Error())
			return finish(rep)
		}
		add("user:"+u.name, true, true, "created")
	}
	return finish(rep)
}

// outboxOnceCmd drains a single outbox row (kill-matrix + worker smoke
// use). --op UUID pins the drain to one row (oldest-first otherwise).
func outboxOnceCmd(ctx context.Context, rep *Report, add func(string, bool, bool, string)) int {
	rep.Mode = "outbox-once"
	dsn := os.Getenv("POC_DATABASE_URL")
	if dsn == "" {
		add("env:POC_DATABASE_URL", true, false, "unset")
		return finish(rep)
	}
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		add("pool", true, false, err.Error())
		return finish(rep)
	}
	defer pool.Close()
	host, port := splitHostPort(getenv("POC_IMMUDB_ADDR", "localhost:3322"), 3322)
	app, err := ledger.DialDir(ctx, host, port,
		getenv("POC_IMMUDB_USER", "poc_writer"), os.Getenv("POC_IMMUDB_PASSWORD"),
		getenv("POC_IMMUDB_DB", "defaultdb"),
		getenv("POC_WRITER_STATE_DIR", ".poc-writer-state"))
	if err != nil {
		add("ledger-dial", true, false, err.Error())
		return finish(rep)
	}
	defer app.Close(ctx)
	wantOp := flagValue("--op", "")
	if wantOp != "" {
		rep.Data["op"] = wantOp
	}
	if err := worker.RunOnceOp(ctx, pool, app, wantOp); err != nil {
		if err == worker.ErrNoWork {
			add("outbox-drain", true, true, "no pending rows")
			return finish(rep)
		}
		add("outbox-drain", true, false, err.Error())
		return finish(rep)
	}
	add("outbox-drain", true, true, "processed one row")
	return finish(rep)
}

func flagValue(name, def string) string {
	for i, a := range os.Args {
		if a == name && i+1 < len(os.Args) {
			return os.Args[i+1]
		}
	}
	return def
}

func splitHostPort(addr string, defPort int) (string, int) {
	h, p, ok := strings.Cut(addr, ":")
	if !ok {
		return addr, defPort
	}
	port, err := strconv.Atoi(p)
	if err != nil {
		return h, defPort
	}
	return h, port
}

// demoMonth runs 25 synthetic employees over October 2026 with mixed
// patterns (standard days, overtime days, an unpaid-leave day, zero days)
// under one pack, returning the digest-of-digests plus totals.
func demoMonth() (digest string, employees, days int, ot, ded int64, err error) {
	pack := synth.RulePack{
		Version: "SYNTH-2026-10-v1", EffectiveFrom: "2026-10-01", EffectiveTo: "2026-10-31",
		Synthetic: true, OvertimeThresholdMin: 480,
		OvertimeNumer: 3, OvertimeDenom: 2, LeaveDeductionDenom: 30,
	}
	h := sha256.New()
	for e := 0; e < 25; e++ {
		base := int64(240000 + e*8000) // $2400..$4320 monthly
		var inputs []synth.DayInput
		for d := 1; d <= 31; d++ {
			day := fmt.Sprintf("2026-10-%02d", d)
			in := synth.DayInput{Day: day, MonthlyBaseCents: base}
			switch {
			case d == 15:
				in.OnUnpaidLeave = true // one leave day each
			case d%7 == 0:
				in.WorkedMin = 0 // zero day
			case d%5 == 0:
				in.WorkedMin = 540 // 9h: 60 OT min
			default:
				in.WorkedMin = 480
			}
			inputs = append(inputs, in)
		}
		res, err := synth.ComputePay(inputs, []synth.RulePack{pack})
		if err != nil {
			return "", 0, 0, 0, 0, fmt.Errorf("employee %d: %w", e, err)
		}
		h.Write([]byte(res.Digest))
		ot += res.TotalOTCents
		ded += res.TotalDedCents
		days = len(res.Days)
	}
	sum := h.Sum(nil)
	return hex.EncodeToString(sum), 25, days, ot, ded, nil
}

// seedPocCmd loads the Phase 2 shared workload via seedpoc.Seed using
// POC_DATABASE_URL (app role: seeding doubles as a grant-path proof).
func seedPocCmd(ctx context.Context) int {
	dsn := os.Getenv("POC_DATABASE_URL")
	if dsn == "" {
		fmt.Fprintln(os.Stderr, "seed-poc: POC_DATABASE_URL unset")
		return 1
	}
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		fmt.Fprintln(os.Stderr, "seed-poc: pool:", err)
		return 1
	}
	defer pool.Close()
	summary, err := seedpoc.Seed(ctx, pool)
	if err != nil {
		fmt.Fprintln(os.Stderr, "seed-poc:", err)
		return 1
	}
	fmt.Println(summary)
	return 0
}

// canonicalCmd prints the six Phase 2 outputs for the shared workload.
// DB-free: the canonical month runs through synth.ComputePay only.
func canonicalCmd() int {
	m, err := workload.Canonical(workload.Build())
	if err != nil {
		fmt.Fprintln(os.Stderr, "canonical:", err)
		return 1
	}
	raw, err := workload.OutputsJSON(workload.BuildOutputs(m))
	if err != nil {
		fmt.Fprintln(os.Stderr, "canonical:", err)
		return 1
	}
	fmt.Println(string(raw))
	return 0
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
