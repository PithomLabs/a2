# EXPERIMENTS — pinned specifications (Gate A, unexecuted)

Every experiment below is specified but NOT RUN (no database in this
environment). Gate B executes them in Claude's order (E5 → E1–E3+Tx/kill →
forged-state → outbox/tamper → month). Expected outcomes are part of the pin:
a different outcome is a finding, not a pass.

## PG18 experiments (from Phase 2 §4, quoted)

- **E1. WITHOUT OVERLAPS on generated range.** On 18.6 scratch DB:
  `CREATE EXTENSION btree_gist`; table with `effective_from/to DATE` +
  `eff daterange GENERATED ALWAYS AS (daterange(...)) VIRTUAL`; attempt
  `UNIQUE (id, eff WITHOUT OVERLAPS)`; repeat with `STORED`; test empty-range
  rejection, GiST index presence, and `ADD CONSTRAINT` on a pre-populated
  table. Expect: STORED accepted; VIRTUAL verdict recorded either way.
  Harness: `e2e/poc_test.go TestE1_WithoutOverlapsOnGenerated`.
- **E2. PERIOD FK multi-row cover + boundaries.** Parent with
  `UNIQUE (k, eff WITHOUT OVERLAPS)` holding adjacent `[Jan1,Jan31]` +
  `[Feb1,Feb28]`; child period `[Jan15,Feb15]` must pass; child with a 1-day
  gap must fail; `ON DELETE CASCADE` must error at DDL; cost at 10k rows
  recorded. Harness: `TestE2_PeriodFKJointCover`.
- **E3. Degenerate containment key.** Parent
  `UNIQUE (employer_id, id, eff WITHOUT OVERLAPS)` + child FK
  `(employer_id, terms_id, eff PERIOD)`; subset passes, 1-day overhang fails
  both directions, concurrent parent-close vs child-insert serializes or fails
  loudly under the employee advisory lock.
- **E4. Auth/privilege parity.** `SHOW password_encryption`; SCRAM roles;
  `has_database_privilege('app',...,'TEMP') = false`; each trigger fn invoked
  directly as app (must fail) and via DML (must work); AFTER constraint
  trigger sees queue-time role.
- **E5. Full battery on 18.6.** Digest-pinned `postgres:18.6-alpine`, volume
  at `/var/lib/postgresql`, `init-roles.sh`, `migrate up` 00001–00012,
  `go test ./...` with `WORKFORCEOPS_TEST_DATABASE_URL` set; diff any pinned
  advisory-count / `pg_locks` / timing assertions. On failure: stop, record
  verbatim, never edit frozen migrations.
- **E6. `pg_upgrade --check` 16→18 dry run** (both binaries present) to record
  the exact checklist output for the future real-data path.

## DBOS crash experiments (from Phase 3 §4, quoted)

1. Inside `capture` txn, pre-commit (`kill -9`): zero partial rows (single
   txn rollback); recovery replays from last checkpoint. Assert: no
   `transaction_completion`/`operation_outputs` row for step; re-run succeeds
   once.
2. Between txn1-commit and txn2 (separate-DB only): kill after user commit,
   before system checkpoint. Layer-2 replay: `fn` not re-run, output adopted.
   Assert via row timestamps + `fn` invocation counter.
3. Between `RunAsStep` side effect and checkpoint: kill after external write.
   Expect re-execution (at-least-once). Sink must show dedup via idempotency.
4. During `Sleep`/approval `Recv`: kill mid-wait; resume with remainder only.
   Assert wake time ≈ original deadline ± skew.
5. Ownership fencing: two executors, kill old mid-step, recover on new. Old's
   late checkpoint refused (`ErrConflictingWorkflowID`), single outcome row.
6. Cancel mid-step: no checkpoint, `CANCELLED`, resume re-executes.
7. Dequeue race: two executors, partitioned employer queue; kill claimer.
   Redelivery, `recovery_attempts` increment, dead-letter after 100.
8. Non-determinism trip: step-order change without Patch over PENDING
   workflow; expect `UnexpectedStep`/`StepNondeterminism`, never silent
   divergence.

## Dual-write crash matrix (PG↔immudb, 5 boundaries + worker)

Failpoints in `poc-run` (`POC_FAILPOINT`): `pre-commit-capture` (B1),
`after-pg-commit-before-append` (B2), `after-append-before-backlink` (B3).
Each reached boundary blocks (sleep) so the harness can SIGKILL
deterministically. The DBOS step effect→checkpoint window (B4/B5) lives
inside DBOS runtime and is NOT hookable from user code: it is covered by
bounded timed kills (`TestKillMatrix_StepWindow`, 5 attempts) plus Claude
DBOS experiments 3–5 semantics, never claimed as deterministic.

- B1 pre-commit capture: kill → zero partial rows; restart replays once.
- B2 after PG commit, before append: kill → outbox row open; restart appends
  once (precondition free), back-link INSERTs.
- B3 after append, before back-link: kill → restart hits BOTH preconditions,
  marker digest compared (equal → TxScan recover + backlink INSERT;
  different → loud aliasing failure).
- B4/B5 DBOS step boundaries: kill → re-execution absorbed by sink
  idempotency; single outcome row.
- PG-down: steps fail, DBOS retries, no ledger write without PG commit.
- immudb-down: append fails loudly, outbox stays open, PG truth intact.
Harness: kill-matrix test (spawns `poc-run` per failpoint, SIGKILLs,
restarts, asserts PG + ledger persisted state). Skips loudly without stack.

## Tamper battery

- T-amper-PG: edit a snapshot row directly → auditor re-derivation mismatch,
  loud failure, original evidence intact.
- T-amper-ledger: rewrite ledger bytes on the immudb server volume →
  `VerifiedTxByID`/dual-proof fails `ErrCorruptedData` + auditor webhook;
  PG truth intact.
- Auditor holds prior trusted state from its OWN volume (never mounted into
  the writer); genesis hash recorded out-of-band. Threat model tested: ledger
  operator / disk tampering — NOT hostile host admin holding both volumes
  (explicitly beyond the POC model).

## Forged-state test (R3 tripwire)

As the `app` role: insert/alter DBOS `workflow_status`, `operation_outputs`,
and `notifications` rows (forged approval `Recv`, forged step output).
Assert: no business state transition occurs without a PG-recorded,
database-enforced authorization; forged checkpoints never move payroll
state. If this fails, R6 reopens (finding #3).

## Idempotency / coverage / performance / reproducibility

- Double-enqueue: same workflow ID enqueued twice → single execution, single
  outcome row (DBOS dedup IDs); same op_uuid twice → one outbox row (already
  harnessed: `TestOutbox_IdempotentEnqueue`).
- Coverage: `TestCoverage_BlockedByConstraint` (PERIOD FK violation, not app
  code); SHG/comp mid-period fixtures → typed blockers (F1a).
- Perf sanity: synth-calc throughput + per-phase timings logged (no brittle
  thresholds); DB-phase timings recorded on first green run, never optimized.
- Reproducibility: full month rerun from pinned inputs + pack + calc digest →
  byte-identical content digest (F1b when computed paths exist).

## Coverage map (spec → harness → gate)

| Spec | Harness | Runs in |
|---|---|---|
| E1, E2, coverage, outbox | `e2e/poc_test.go` | B1–B2 |
| E3, E4, E6 | runbook SQL + CHECK scripts (A6/A12) | B1–B2 |
| E5 | runbook step 6 (a2 suite on :5544) | B1 |
| DBOS 1–8 | kill-matrix + executor tests | B2–B3 |
| B1–B5, tamper, forged-state | kill-matrix + tamper + forged tests | B3–B4 |
| Month + 6 artifacts | `poc-run` full mode + month test | B5 |
