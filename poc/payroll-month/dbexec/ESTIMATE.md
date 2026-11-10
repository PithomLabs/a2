# U4 estimate: Tx-threading the capture path (spike result + cost)

## Spike result (VERIFIED by compilation)

Zero-adapter unification was attempted and REFUTED: `dbos.Tx` returns
sysdb-internal interfaces (`Result`/`Rows`/`Row`), not pgx types, so neither
`dbos.Tx` nor any pgx type satisfies a shared raw interface. The spike
therefore ships thin adapters (`PoolExec`, `PgTxExec`, `DbosExec`) over our
own minimal `Executor` (`Exec` → own `Result`; `QueryRow` → own `Row`);
internal types are never named (values captured eagerly; Scan pinned by
assertion). `SameTxProbe` is the shared runtime-proof body; its execution
through `RunAsTransaction` is Gate B acceptance (B2), still open.

Good news: `internal/payroll/snapshot.go:35-42` already has a `qtx` seam
unifying pool + pgx.Tx — capture core is pre-abstracted, so the mechanical
swap below is smaller than the raw census suggests (but `qtx` is still
pgx-typed, so DBOS-facing call sites change regardless).

## Census (VERIFIED 2026-11-10, `*.go` excluding tests)

- 105 `*pgxpool.Pool` references across `internal/` + `core/`; 4 `*pgx.Conn`.
- 434 pool method calls: `QueryRow` 298, `Exec` 71, `Query` 41, `Begin` 24.
- Capture-path core files: `snapshot.go` (6 pool refs), `runs.go` (3),
  `readiness.go` (3), `validate.go` (2), `core/database/{admission,db,lock}.go`
  (5). `compbasis.ResolveEffectiveCompensation` already takes an interface
  (`resolveQuerier`) — no change needed there.

## Cost estimate (INFERRED, needs live-PG confirmation)

- Mechanical: swap ~25 signatures in the capture-called chain from
  `*pgxpool.Pool` to `dbexec.Executor` (call sites unchanged — same methods).
- Design (the real work): 24 `pool.Begin` sites need boundary review —
  callees must stop opening their own transactions inside a DBOS txn, or the
  atomic path silently splits. Each Begin gets an owner: caller-txn or
  explicit new-txn with a written reason.
- Re-verify: full suite + T1–T13 + crash battery on the converted path.
- Total: 2–4 days mechanical + 1 day boundary review + battery re-run.
  Full-tree conversion (105 sites incl. timesheet×13) is NOT required for the
  POC and is explicitly deferred.

## What would overturn this estimate

- `dbos.Tx` growing mutating methods the interface must exclude (re-check on
  every DBOS version bump; pinned v1.8.0 re-verified this gate).
- Capture calling a service that needs `CopyFrom`/LISTEN (none found in the
  census: zero `CopyFrom`, zero `Begin` inside `compbasis`).

## U4 design (EVIDENCED 2026-11-10, Gate B census)

Principle: concrete pgx types exist ONLY at txn origins (pool begins);
everything downstream takes `Executor`. No logic duplicated (AGENTS.md #1):
pool/txn wrappers keep their signatures and wrap at the boundary.

Census (exact, `*.go` excluding tests):

- `qtx` (`snapshot.go:35-41`): QueryRow->pgx.Row, Query->pgx.Rows,
  Exec->(pgconn.CommandTag,error). Takers: 9 snapshot.go helpers
  (expectedPopulation, runExclusions, requiredPackSet, resolveEmployee,
  resolveSheet, sheetChain, resolveComp, resolveSHG, resolvePacks) + 7
  readiness.go fns (6 check* + requiredPacks).
- `pgx.Tx` params IN-path (change to Executor): captureInTx
  (`snapshot.go:282`); 3 locks (`core/database/lock.go:29,48,59` —
  Exec-only bodies calling acquire_* SQL helpers); tipDrifted
  (`validate.go:266` — passes tx to expectedPopulation at :267, called
  once from Validate at :149).
- `pgx.Tx` params OUT-of-path (keep as-is): snapshotBlockers,
  packBlockers, tipBlockers, transitionValidated (no qtx-passing; direct
  tx use + audit.Write via DBTX, which pgx.Tx still satisfies).
- Zero tag reads: no RowsAffected/CommandTag use in
  snapshot/readiness/runs/validate (only the qtx def line) — Exec
  adaptation points: none.
- Rows iteration is Next/Scan/Close/Err only — a mirrored Rows iface
  compiles all loops unchanged. QueryRow(...).Scan unchanged.
- `pgx.ErrNoRows`: ~20 `==` sites — DbosRow.Scan MUST return the inner
  error UNWRAPPED (identity preserved) or the design changes to
  errors.Is with 20-site churn. Live-PG gate, not assumed.
- audit.Write: DBTX Exec-only, ignores tag. 4 in-path call sites
  (runs.go:100,158; snapshot.go:464; validate.go:487) move to an
  additive audit.WriteExec; the other 8 caller files untouched.
- Pool->qtx wrap sites (1-line PoolExec): snapshot.go
  :553,:567,:571,:584,:606,:1170; readiness.go :137,:141,:188,:196,
  :202,:208,:214 + requiredPacks sites.
- Txn origins (signatures unchanged, wrap inside): CaptureSnapshot :246
  (BeginSnapshotTx, REPEATABLE READ), CreateRun :36 + AddExclusion :114
  (extract InTx cores), Validate (pool-begun, multi-commit — OUT of the
  DBOS path; wraps its tx at :149,:162,:223 only).
- No cross-package ripple: timesheet.EffectiveApproved takes its own
  `dbtx` (pool still satisfies); refactor touches payroll +
  core/database + audit only.

Layering: Executor + PoolExec + PgTxExec live in `core/database` (main
module); DbosExec stays in the POC (main module must never import
dbos). POC imports main via require+replace — one direction, no cycle.

Steps: (1) core/database/exec.go (~100 lines); (2) qtx redefinition to
the Executor method set (same method names); (3) signature changes +
~25 call-site wraps; (4) audit.WriteExec; (5) DbosExec + require;
(6) battery: ErrNoRows-identity gate, rows parity, SameTxProbe through
RunAsTransaction, full suite + T1-T13 + crash battery. Estimate
stands: 2-4d + 1d boundary review + battery.
