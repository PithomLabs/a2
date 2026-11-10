# Adversarial Code Review — stack-poc (PostgreSQL 18.6 + DBOS + immudb)

Per plans2/prompt7.md. Read-only review: no code modified, no services touched,
no DB tests run. All verdicts are static code evidence plus the two DB-free
suites executed during recon. Severity scale is prompt7 §9.

Method: every baseline doc was read (both POC reports, runbook, workload
record, traceability, all of `docs/stack`, all four ADRs), then four parallel
read-only inspection tracks opened every file body cited. The five most
load-bearing claims were re-verified against source by the reviewer
(month/capture stubs, snapshot grants, run-lifecycle predicate, close
function, auditor verify path, synth test count).

## 1. Executive verdict

The POC honestly executed what it executed, and its gap disclosures are real.
But its load-bearing impression — a DB-backed month that computed this payroll
exactly once and attested it to a ledger — does not hold. The month workflow
returns five label strings, capture runs `SELECT 1`, the canonical digest
lives only on the DB-free path, and the ledger anchors dummy digests. Worse,
three findings go beyond "stub" into "wrong":

- The `app` role can fabricate a snapshot with valid parents **and drive it
  to a `validated` run by direct SQL** — the Rung 5 boundary the POC was
  meant to demonstrate does not exist in the real schema.
- The "independent auditor" verifies transaction *existence* only; it
  discards artifact binding, cannot detect PG tampering, and its trust
  anchor is uncommitted TOFU under credentials shared with the writer.
- The report's claim that fabrication is "blocked downstream
  (claims/backlink/ledger)" is contradicted by the same report: the claims
  verifier is broken and uncalled, backlinks are squattable by `app`, and
  the ledger anchors whatever it is fed.

Gate C direction: keep the *questions* the POC was built to answer, but do
not treat any of the three per-technology recommendations as earned yet.
Details in §11–12.

## 2. The three key questions, answered

1. **Can the application create a valid-looking but unauthorized payroll
   snapshot?** Yes — fully, as `app`, end to end, and onward to a
   `validated` run. P0, proven from grants + trigger predicates (PG-1).
2. **Does DBOS genuinely protect the entire critical transaction
   boundary?** No. The shared-pool single-transaction mechanism is real in
   library code, but the POC's transaction contains `SELECT 1` and the
   month workflow contains no transaction at all. Every DBOS-side
   "exactly once" assertion counts `workflow_status` rows, never business
   effects (DB-1, DB-2).
3. **Can immudb's auditor detect tampering when the adversary controls the
   ledger host and storage?** No. Genesis is trust-on-first-use with no
   committed value and no enforcement flag; auditor and writer share one
   superuser; the auditor never binds ledger content to PG rows; no
   host-corruption evidence exists. A hostile host admin is also explicitly
   out of the R2 threat model, so even a perfect result here would not
   answer the question as asked (IM-1–IM-3, IM-6, IM-7).

## 3. Findings by severity

P0 (critical): 5 — PG-1 fabrication-to-validated chain; DB-backed month
computes nothing (RC-1); §6 execution evidence missing from the report file
(RC-3); bogus-tx test proves nonexistent-tx errors, not tamper detection
(IM-6); auditor verifies tx existence only, PG tamper invisible (IM-7).

P1 (high): 12 — close path dead for `app` yet directly bypassable (PG-2);
closed-period writer check is dead code (PG-3); claims verifier
broken/uncalled/unwritten, vacuous even if repaired (PG-4); outbox
forgeable, backlink squattable, no real producer (PG-5/IM-12); no capture
path takes the gate (PG-6); empty atomic-transaction membership (DB-1);
exactly-once assertions count status rows (DB-2); same-month-ID
resubmission returns stale results silently (DB-3); ERROR rows permanently
poison deterministic IDs (DB-4); forged-SUCCESS pre-registration untested,
tripwire vacuous (DB-5); recovery reads marker via unverified `Get`
(IM-9); backlink PK silently drops corrections (IM-11); "blocked
downstream" contradicts "verifier BROKEN" in the same report (RC-2).

P2 (medium): 13 — E2 "every spelling" was one spelling (PG-7); E5
"IDENTICAL" with differing counts and missing evidence (PG-8/RC-9); Exp1A
strawman baseline, no committed code, mislabeled (PG-9/RC-4); trigger
coverage gaps vs spec, DELETE bypasses (PG-10); forged-DBOS test
attribution unproven, mod-10 fingerprint (PG-11/DB-5); sticky-ID fix tests
a different regime than production (DB-6); approval wait absent (DB-8);
writer state-dir leaks (IM-4); no server flags/TLS/users (IM-5);
host-corruption harness absent (IM-8); poison path orphans ledger records
(IM-10); no production enqueue path, all dual-write evidence over fiction
rows (IM-13); ledger has no PII but also no real provenance binding
(IM-14); Exp2A honestly absent (RC-5); Exp3A contrast real but implicit
(RC-6); digest tripwire self-referential and unpinned (RC-7); synth
"10/10" is actually 9 (RC-8); F5 provenance overclaims (RC-10).

P3 (low): 6 — self-held gate DoS via granted advisory locks (PG-12);
reproducibility test proves dedup of a placeholder (PG-13); claim/close
table squat surface (PG-14); step-window test replays pure functions only
(DB-9); stdout contamination residuals (DB-10); outbox
convergence-as-error, dead `appended` state, kill-retries poison
accounting (DB-11); non-transactional seeder (DB-12); minor wording nits
(RC-11).

## 4. PostgreSQL review

**PG-1 (P0). Valid-parent fabrication succeeds as `app`, through to
`validated`.** Grants in `migrations/core/00012_m2_governance.sql:1094-1099`
give `app` INSERT on all three snapshot tables, and the file's own verify
block *requires* INSERT to stay (00012:1133-1135). FKs check parent
existence only (00012:1010-1011); child tables are nearly FK-free and pack
pins have no FK beyond `snapshot_id` (00012:1021-1033, 1037-1047). Zero
triggers exist on any snapshot table (full trigger inventory of
`migrations/`). `payroll_run` has no explicit REVOKE anywhere, so `app`
holds UPDATE via default privileges
(`core/database/provision.go:117-118`). The `run_lifecycle` trigger
(00012:1225-1262) checks only attacker-chosen recorded state — exactly one
`valid` tip, decisions in (`included`,`excluded`), packs `IMPLEMENTABLE` —
which the attacker sets at INSERT time. Nothing recomputes any hash or
compares against live tables. There is no DEFINER capture function to
bypass; capture is plain Go SQL as `app` (`captureInTx`,
`internal/payroll/snapshot.go:417-463`; the only snapshot DEFINER is the
one-way `invalidate_payroll_snapshot`, 00012:1053-1089). Four statements
mint a `validated` run on invented content with no audit row (capture audit
is written by Go only, `snapshot.go:464-468`). The battery has no row for
fabricate-then-transition (T10 assumes bare UPDATE), and T10–T12 never
executed. Test pin:
`internal/payroll/snapshot_admission_test.go:534-562`
`TestSnapshotFabricationByApp` (runs as `app`, ends in unconditional
`t.Fatalf("RED…")`).

**PG-2 (P1). `poc_close_period` is dead for `app` and bypassable by it.**
The function (`poc/payroll-month/gate/gate2_poc.sql:153-154`) is
invoker-rights (no `SECURITY DEFINER`) but UPDATEs a table where `app`
holds only SELECT+INSERT (gate2_poc.sql:147-148) — every legitimate `app`
close dies 42501. Meanwhile `app` can directly INSERT close rows (no
trigger), squat the `(employer_id, period)` PK to block legitimate closes
permanently (`IF NOT FOUND → 'already closed'`, lines 162-165), or
pre-date `closed_at`, contradicting the "set ONLY by poc_close_period, DB
clock" comment (line 141).

**PG-3 (P1). Closed-period writer rejection never fires.** The trigger
reads `(NEW)->>'period'`/`'pay_period'` (gate2_poc.sql:47-56), but none of
the 12 gated tables has such a column (verified against each table's DDL in
00001/00002/00003/00010: timesheets carry `pay_period_id uuid`, others
carry dates) — the expression is always `''`, which never matches a
`YYYY-MM` close row (`dbosflow/month.go:13`). All 12 triggers degrade to
pure gate-acquire; D1 close semantics are unenforced. (Employer extraction
is fine — all 12 tables have `employer_id`.)

**PG-4 (P1). Coverage verification is absent; fail-open as a system.**
`poc_verify_claims` references nonexistent
`payroll_snapshot_employee.source_table/source_row_id` columns
(gate2_poc.sql:105-134, self-documents "BROKEN ... every call fails with
42703"), has zero callers and zero claim writers anywhere in the repo
(repo-wide grep: only hits are the SQL file and test comments). Pinned
executably by `TestCoverage_BlockedByConstraint` Part 2
(`poc/payroll-month/e2e/poc_test.go:199-206`), which *requires* the 42703
failure. Worse than broken: the function counts claims lacking membership,
so with zero claims ever written, even a repaired version returns
`missing_count=0` — a vacuous pass. The spec it ports
(`core/database/gate2_optionb_test.go:292-349`) checks both directions
including omission inside the capture txn; the POC port checks one
direction post-hoc. The port also grants EXECUTE to `app` (line 134) while
the spec deliberately withholds it (spec lines 350-351).

**PG-5 (P1). Outbox/backlink integrity is grant-only and squattable.**
`app` holds SIU on `poc_outbox` (`poc/payroll-month/ledger/outbox.sql:24-26`)
with only a value CHECK, no transition guard: it can forge `backlinked`
with an invented tx id, mark rows `failed` to suppress anchoring silently,
rewrite pending `payload`/`payload_hash` (worker anchors attacker bytes,
`worker/outbox.go:103-115`), or INSERT arbitrary rows — proven executable
because the *only* producer in the repo is the test helper `seedOutbox`,
which INSERTs as the app pool (`e2e/chaos_test.go:71-74`). Backlink squat:
`app` INSERT on `poc_ledger_backlink` plus the worker's `ON CONFLICT
(pg_table, pg_id) DO NOTHING` (`worker/outbox.go:137-143`) means a
pre-INSERTed forged backlink permanently shadows the genuine anchor —
unfixable, since the append-only trigger binds even the owner. Auditor
fails loudly on the forged tx (fail-closed at audit) but the genuine row
can never be repaired.

**PG-6 (P1). No capture path takes the gate.** Real capture uses advisory
input/population/pack locks and direct INSERTs
(`internal/payroll/snapshot.go:242-275, 417-463`); POC `captureTxn` is
`SELECT 1` (self-documented, `dbosflow/capture.go:81-86`). Repo-wide grep
confirms no Go caller of `poc_gate_acquire`; `admission.go`'s
session-barrier protocol is explicitly "NOT wired into CaptureSnapshot or
ValidateRun". The gate file's coverage claim (lines 68-71: snapshot
tables "written only via capture/close functions, never directly by app
writers") is false in both directions. `TestConcurrency_WriterRacesCapture`
(`e2e/poc_test.go:317-363`) therefore proves the advisory-lock primitive
only — its own comment (lines 318-323) says so.

**PG-7 (P2). E2: "every spelling rejected 42601" — exactly one spelling
was attempted.** `TestE2_PeriodFKJointCover` (`e2e/poc_test.go:106-130`)
tries a single spelling (`FOREIGN KEY (k, need PERIOD) REFERENCES
poc_e2_parent (k, eff PERIOD)`, lines 122-124) and asserts 42601. Report
claim at `plans2/POC-FINAL-REPORT.md:61-63`. Joint coverage was not and
cannot be tested (feature absent); the tripwire only pins absence.

**PG-8 (P2). E5 "IDENTICALLY": counts differ, evidence section is a stub,
signal is mostly intentional RED.** `POC-FINAL-REPORT.md:13` claims
identical behavior, but: (a) cited "details §6" is a literal
`...[truncated 6186 chars]` stub (file lines 143-146, file is 145 lines);
(b) `POC-REPORT.md:192-200` gives 16 ok/15 fail vs 16 ok/14 fail —
different counts, reconciled only by an unexplained "Gate2 load-flake
swap"; (c) 11 fails are intentional TDD RED plus 1 TZ-env artifact — the
comparison largely shows deliberately-failing tests failing on both
versions, a weak regression signal; (d) concurrency evidence covers the
lock primitive, not the capture race, on both versions equally.

**PG-9 (P2). E1 baseline unfairness: probe-vs-probe labeled as
"compensation records"; production already blocks the demonstrated
attack.** Arm A ("Plain table ... 2 overlapping rows INSERTed and silently
kept") is an uncommitted live probe on an undescribed scratch table — no
committed test contains a plain-table arm (`TestE1…`,
`e2e/poc_test.go:67-104`, uses only `poc_e1_stored`/`poc_e1_virtual`
probe tables). The case is labeled "Overlap/gap compensation records" but
no `compensation_line` rows were used. Status quo is stronger than Arm A:
production `compensation_line` already rejects same-code overlaps via
`check_compensation_line` (`00011_m13_integrity.sql`,
`compensation_line_containment` BEFORE INSERT trigger). The experiment
contrasts WITHOUT OVERLAPS against *nothing*, not against the existing
trigger.

**PG-10 (P2). POC trigger coverage gaps vs the spec it ports; T1–T12 never
executed.** Divergences from `gate2_input_gate`
(`core/database/gate2_optionb_test.go:154-223`): (a) POC triggers are
`INSERT OR UPDATE` only (gate2_poc.sql:83) — DELETE bypasses the gate
(spec covers DELETE, lines 212-223); (b) POC reads only `NEW.employer_id`
— cross-employer UPDATE takes only the new key (spec locks both sorted,
lines 192-199); (c) the spec's phase-marker check is dropped entirely;
(d) `payroll_run`/`pay_period`/exclusion/correction-group writers are
ungated, so the validate/transition path races capture outside the gate.
T1–T12 exist only as ADR spec rows.

**PG-11 (P2). Forged-DBOS test is vacuous; "executor ignores forgeries"
attribution unproven; fingerprint truncates counts mod 10.**
`TestForgedState_MovesNothing` (`e2e/forged_test.go:28-75`) forges into
`dbos.*` and asserts business counts unchanged — but no business code path
reads `dbos.*` tables, so the test cannot tell "executor ignored it" from
"nothing reads it". `POC-FINAL-REPORT.md:97-101` then attributes
resilience to the executor — unproven by this test. `businessCounts`
(`forged_test.go:79-95`) fingerprints each count as a single mod-10
digit, so counts differing by multiples of 10 compare equal.

**PG-12 (P3). `app` retains raw xact-advisory + direct gate EXECUTE.**
gate2_poc.sql:247-263 grants all 8 xact advisory overloads to `app`, and
line 27 grants `poc_gate_acquire` directly. A compromised/buggy app worker
can stall a whole employer's payroll path with one statement in a
long-running txn. Session family correctly withheld. DoS only, no
correctness break.

**PG-13 (P3). `TestReproducibility_RerunIdentical` proves DBOS dedup, not
month-digest identity.** `e2e/poc_test.go:365-381` compares stdout of two
`capture-once` runs with the same workflow ID — byte-identical because the
second is a dedup hit on a placeholder (`InputsHash:
"placeholder:"+runID`, `capture.go:96`). Month-level digest identity
across independent executions is not shown.

**PG-14 (P3). `poc_claim`/`poc_period_close` squat surface (no writers,
attacker-writable PKs).** `app` holds INSERT on `poc_claim`
(gate2_poc.sql:102-103, PK `(run_id, table_name, row_id)`) with no
legitimate writer in the repo. Any future proof-check trusting claim rows
inherits forgeable inputs. Latent; becomes P1 if the verifier is ever
repaired without an authorized-writer design.

## 5. DBOS review

Library inspected at
`/home/chaschel/go/pkg/mod/github.com/dbos-inc/dbos-transact-golang@v1.8.0`
(no `replace` directive; pinned in `poc/payroll-month/go.mod:8`).

**DB-1 (P1). The atomic capture mechanism is real, but its current
membership is empty.** `SameEngine` is pointer equality on `*pgxpool.Pool`
(lib `dbos/datasource.go:390-394`); POC passes the identical `pool` to
both `dbos.NewContext` and `dbos.NewDataSource`
(`cmd/poc-run/main.go:161,181`), so `RunAsTransaction` collapses onto
`runAsTxn`: `fn(tx)`, then `RecordOperationResult(Tx: tx)`, then
`tx.Commit` in ONE pgx txn (lib `workflow.go:2994,3065,3072`). Business
writes + durability record genuinely commit atomically **if both inside
`fn`** — but `fn` is `SELECT 1` + failpoint + placeholder string
(`dbosflow/capture.go:87-97`). Gate acquisition, snapshot INSERTs, and
outbox row are all TODO (`capture.go:81-86`). The month workflow is worse:
no `RunAsTransaction` at all — five pure `RunAsStep` string formats
(`month.go:36-44,52-54`). A later slice adding snapshot writes adjacent
to (not inside) `captureTxn`, or via `*pgxpool.Pool` services that cannot
see `dbos.Tx` (the flagged U4 gap), silently exits the atomic boundary.

**DB-2 (P1). "Exactly 1 SUCCESS" asserts workflow-status rows, never
business effects.** Every DBOS kill/dedup assertion queries only
`dbos.workflow_status`, e.g. `e2e/chaos_test.go:245-252`,
`chaos_test.go:278-286`, `e2e/idempotency_test.go:73-80`,
`e2e/forged_test.go:62-68`. "Exactly-once from history (30 SUCCESS / 98
checkpoints)" is status-row convergence — DBOS dedup works — not "each
business effect happened once." The only kill test asserting real effects
is the outbox boundary test (`chaos_test.go:178-217`), and the outbox
worker is deliberately not a DBOS workflow. A future non-idempotent
capture body would pass every current exactly-once test.

**DB-3 (P1). Duplicate submission with the same month ID silently returns
the stale result; differing input is dropped without comparison.**
Production month IDs are deterministic
(`cmd/poc-run/main.go:218-220`). Library default
`WorkflowIDReusePolicyReturnExisting` (lib `workflow.go:899-900`):
conflicting status insert is a silent no-op
(`system_database.go:1493-1495`), new input silently discarded with no
comparison (`INSERT ... workflow_input ... ON CONFLICT DO NOTHING`,
`:1571-1572`), terminal-or-existing row never executes (skip path,
`workflow.go:1513-1531`). Name mismatch errors but input mismatch is
ignored. Oct data corrected and resubmitted as
`month:poc-meridian:2026-10` → exit 0 with OLD phases/digest, nothing
re-executes. `TestDoubleEnqueue_SingleExecution` only resubmits identical
input, so this is untested.

**DB-4 (P1). ERROR is sticky: a failed month can never be retried under
the same ID with the default policy.** `insertStatusResult.Status ==
WorkflowStatusError` is in the `shouldSkip` disjunction
(`workflow.go:1513-1516`); no `WithWorkflowIDReusePolicy` override exists
anywhere in the POC. Transient failure → row `ERROR` → operator retry
replays the recorded error via polling handle without executing. Recovery
requires a new workflow ID (two IDs for one month) or manual
`workflow_status` surgery. No POC test covers ERROR→retry.

**DB-5 (P1). Forged-SUCCESS pre-registration (the real dedup-poisoning
attack) is not tested; the tripwire is vacuous two ways.**
`e2e/forged_test.go:28-75` forges under a random ID no workflow uses,
then runs a real workflow under a different ID. Under DB-3's `shouldSkip`
logic, pre-registering SUCCESS for the real month ID means the month never
executes and `GetResult()` returns attacker-chosen output — untested.
Further: forge outcomes are only logged, never asserted (`:49`), so grant
regressions are invisible; and `businessCounts` is a mod-10 fingerprint
over tables the workflows never write (`:83-93`), so "moved nothing"
would pass even if the executor consumed forged rows.

**DB-6 (P2). The "sticky IDs" fix tests a different regime than
production and weakens the suite.** Per-run nanos IDs
(`e2e/chaos_test.go:227` et al.) are legitimate hermeticity hygiene, but
production month IDs are deterministic — F3/F4 live exactly in the gap
between unique test IDs and deterministic prod IDs. `UnixNano` uniqueness
across processes/restarts is probabilistic (collision → dedup-hit → the
original flake returns in principle). `capture-once`'s default run ID is
still sticky `"capture:default"` (`cmd/poc-run/main.go:242,258`), so any
rerun without `--run-id` silently no-ops via dedup.

**DB-7 (P1 production / P3 POC CLI). Shutdown closes the user-owned
shared pool; the LISTEN-conn hold is the leak mechanism; the workaround
abandons a `Close()` goroutine.** Chained library facts:
`SysDB.Shutdown` calls `s.pool.Close()` on the same pool the app passed
in (`system_database.go:1288-1348`, `dbq.go:126`); the notification
listener holds a shared-pool conn via `LISTEN` for the loop's lifetime
(`:4087-4100`, `:1215-1220`), so `pool.Close()` blocks until it exits —
and permanently shrinks app-pool capacity by one slot at runtime;
`dbosflow/close.go:21-32` abandons the blocked `Close()` (fine for the
CLI; leaks accumulate across tests in the e2e binary). All `Shutdown`
errors are discarded (`_ =`, `main.go:211,250`), so timeout-vs-clean is
invisible. Production implications: DBOS restart without process restart
is impossible; Shutdown-timeout leaves in-flight workflows running while
their pool closes beneath them. (Defer order in `runMonthWorkflow` is
correct LIFO — credit where due.)

**DB-8 (P2). Approval wait does not exist anywhere in the POC.** Grep for
`Recv|Sleep|SetEvent` over `poc/payroll-month/**/*.go` finds only the
month.go future-work comment, `time.Sleep` in test pacing, and the forged
test's fake `'approval'` notification row. Injection #4 and forged-approval
paths are unproven — honestly disclosed in the report.

**DB-9 (P3 today, P2 constraint on future slices). RunAsStep side-effect
audit: no side-effecting steps exist.** `runPhase`/`readinessCheck`/
`hashVerify` are pure; `captureTxn` is `SELECT 1`; ledger append +
backlink (`worker.RunOnce`) is a plain process, not a step — idempotent
via immudb `MustNotExist` + digest-bound marker + backlink conflict guard
+ guarded transition: the one genuinely proven effect-once path (B2/B3
kills). `TestKillMatrix_StepWindow` proves status convergence, not
effect-once. Constraint: placing the ledger append inside `RunAsStep`
re-opens crash-between-effect-and-checkpoint (execute and checkpoint are
separate txns, lib `workflow.go:2852,2893`). Keep the append outside DBOS
with its own idempotency; do not "simplify" it into a step.

**DB-10 (P3). stdout JSON report: contamination mostly mitigated, two
residual paths.** Good: DBOS logger to stderr at WARN
(`main.go:162-169`), helper logs to stderr, `finish()` prints only the
report (`main.go:119-120`). Residuals: (1) lib binary-hash failure path
`Printf`s to stdout (`dbos/dbos.go:computeApplicationVersion`), breaking
`poc-run | jq` and byte-equality tests when it triggers; (2)
test-harness-only: `spawnFailpoint`'s deadline is checked only between
stderr lines (`chaos_test.go:93-107`) — a silently hung `poc-run` blocks
`Scanner.Scan()` until the outer `go test` timeout.

**DB-11 (P3). Outbox worker: converged-duplicate surfaces as failure;
dead `appended` state; poison-vs-appended tension.**
`worker/outbox.go:146-155` returns an error when a duplicate correctly
converged (exit 1 with a scary message; a real multi-worker deploy will
hit this). No code path ever sets `state='appended'`, though the
transition lists it. Poison accounting counts kill-retries as attempts
(`:75-84`): 10 crash-retries → `failed` even if the ledger append
committed — operator must reconcile a failed row with a live ledger
record, with no runbook line.

**DB-12 (P3). Seeder is non-transactional with refuse-to-rerun
semantics.** `seedpoc/seedpoc.go:64-157`: hundreds of pool ops, no
wrapping transaction; mid-way failure leaves a half-seeded DB and rerun
is refused (`:71-73`). Correct posture for a disposable POC DB; do not
carry the pattern into non-disposable seeding. The raw-SQL leave inserts
(`:342-357`) bypass future service-level invariants silently.

## 6. immudb review

**IM-1 (P0). Genesis anchor is TOFU with no committed out-of-band value
and no re-verification flag.** First auditor run bootstraps trust purely
from whatever server answers. The anchor prints to stderr
(`cmd/auditor/main.go:100-102`) and is never compared on later runs — no
`--expect-state-hash`/`--expect-uuid` flag exists, and no committed value
exists anywhere in git (grep for `4003ecfe`/`state_hash` finds only the
print statement and report mentions). Attacker (or misconfiguration)
serving a forked immudb at first-run time owns the anchor permanently.
Reports honestly mark the out-of-band commit "still pending", but the
"Independent auditor verification | PROVEN" verdict coexists with it.

**IM-2 (P1). Auditor "refuses identity mismatch" only via SDK default on
dir reuse — and no test ever reuses a dir.** Mechanism verified in
`immudb@v1.11.2`: identity check defaults enabled, binds UUID to the
address string via a state-dir file. But a server swap preserving
`host:port` with a wiped state dir is accepted, and every e2e run uses a
fresh `t.TempDir()` + `--accept-genesis`
(`e2e/poc_test.go:274-285`), so mismatch-refusal is never exercised —
the header claim in `cmd/auditor/main.go:1-5` ("later runs refuse
identity mismatch") is untested.

**IM-3 (P1). Auditor uses IDENTICAL credentials to the writer; "own
creds" comment is false.** `docker-compose.poc.yml:55` claims "own
container, own volume, own creds", but lines 66-70 use the same
`POC_IMMUDB_USER/PASSWORD/DB` vars the writer uses
(`cmd/poc-run/main.go:288-292`), defaulting to superuser `immudb` on
`defaultdb` — the same var feeds `IMMUDB_ADMIN_PASSWORD` (compose line
43). Volumes are separate (true); credential separation is zero. Expected:
read-only auditor user per R2 model.

**IM-4 (P2). Writer state dir is neither hermetic nor consistent across
harnesses.** `chaos_test.go:161` sets `POC_WRITER_STATE_DIR=t.TempDir()`
for spawned binaries, but `:199` calls `ledger.Dial(...)` which hardcodes
dir `"."` (`ledger/ledger.go:77-79`). `TestTamperDetection` sets no
override (`e2e/poc_test.go:231-255`). `e2e/.poc-writer-state/` on disk
contains live trusted-state files shared across test runs.

**IM-5 (P2). immudb server runs with zero flags: no `--signingKey`, no
TLS, no dedicated users.** `docker-compose.poc.yml:31-52` has no
`command:`, only `IMMUDB_ADMIN_PASSWORD`; clients use
`insecure.NewCredentials()` (`ledger/ledger.go:92`,
`cmd/auditor/client.go:25`). Acceptable for localhost POC, but no
promotion checklist pins what must change (key custody, TLS, per-role
users) — the R2 custody story covers the auditor volume only.

**IM-6 (P0 — claim scope). The bogus-tx test proves "nonexistent tx
errors", NOT tamper detection.** `e2e/poc_test.go:279-288` feeds tx
`999999999`, never written, with a fresh state dir. `VerifiedTxByID`
fails because the server has no such tx — no inclusion proof is checked
against prior trusted state, no server byte was altered. This test cannot
distinguish an honest empty server from a tampered one. F4 rows ("Tamper
fails verification loudly: PASS") and "Forensic readiness: PROVEN LIVE"
overclaim what this establishes.

**IM-7 (P0). The auditor verifies tx *existence*, never content binding
— PG tamper is NOT detected by the auditor as built.**
`cmd/auditor/client.go:37-42` discards the verified tx (`if _, err :=`);
`cmd/auditor/main.go:112-148` parses the backlinks file but uses only
field 4 (tx id) — `pg_table`/`pg_id`/`event_uuid` are parsed and dropped.
Consequences: (a) editing a PG snapshot row changes nothing the auditor
checks — it still passes, directly contradicting runbook §4.2 ("edit a
snapshot row in PG … The auditor … must fail verification loudly");
(b) swapping two valid tx ids between artifacts passes; (c) the auditor
trusts the backlinks *file* it is handed — no code produces that file
from PG, no ledger enumeration, no PG↔ledger reconciliation.

**IM-8 (P2 — acknowledged gap, confirmed). No host-file corruption
harness exists.** Full-tree grep finds only comments and the manual
runbook recipe. Smallest safe disposable test (proposed, not executed):
bring up stack, seed 1 row, stop immudb, copy the data volume, flip bytes
in the copy, swap-mount, start, run auditor with the *pre-corruption*
state dir, expect proof failure; `down -v` after. Never against a needed
volume — corruption is one-way.

**IM-9 (P1). `recoverTx` reads the marker via UNVERIFIED `Get` —
wrong-tx backlinks pass audit silently.** `worker/outbox.go:166-178` →
`ledger/ledger.go:161-167` uses `cli.Get`, not `VerifiedGet` (SDK has
`VerifiedGet`/`VerifiedGetSince`/`VerifiedGetAt`). Precondition-hit (the
normal B3 crash-recovery path) trusts the server's `(value, tx)` with no
proof. A faulty server returning a wrong-but-well-formed tx produces a
backlink to the wrong tx; the IM-7-blind auditor verifies *that* tx
cleanly. Expected: `VerifiedGet` or re-`VerifiedTxByID` on the returned
tx + value match. Exploitability within the stated threat model is
partial (requires faulty/compromised server at the recovery instant).

**IM-10 (P2). Poison path (`failed` after 10 attempts) can orphan a
ledger record with no backlink and no reconciler.**
`worker/outbox.go:85-95` marks `failed` and returns; the row is never
re-driven (drain selects `pending` only, `:63-68`). Append succeeds,
backlink INSERT fails 10× → ledger holds an unattested record, PG holds
`failed`, no job re-attempts or alerts. Silent *after* the loud return.

**IM-11 (P1 design). Backlink PK `(pg_table, pg_id)` + `ON CONFLICT DO
NOTHING` silently drops corrections.** Correction for the same artifact
appends a new ledger tx, then its backlink INSERT conflicts with the
original row and is discarded — while the outbox row is marked
`backlinked` with the *new* tx id (`worker/outbox.go:146-149`). End
state: `outbox.immu_tx` (new) ≠ `backlink.immu_tx` (old), auditor
verifies the old tx, correction unattested, zero error. Latent (no
capture body enqueues corrections — IM-13), but the schema/worker
contract already encodes it. Expected: PK on
`(pg_table, pg_id, event_uuid)` or conflict → loud error.

**IM-12 (P1). `poc_outbox` state machine and payload are grant-only —
`app` can forge completion or mutate payload.** See PG-5. Additionally:
`app` holds INSERT on `poc_ledger_backlink`
(`ledger/outbox.sql:50-51`) with no constraint that `immu_tx` exists —
forged backlinks to arbitrary *valid* txs pass the IM-7-blind auditor.

**IM-13 (P1, scope). There is NO production enqueue path — the PG→ledger
binding is unimplemented, not merely untested.** `INSERT INTO poc_outbox`
appears in zero non-test Go files (only `e2e/chaos_test.go:71-74`,
`e2e/poc_test.go:144-146`); `captureTxn` is an explicit placeholder
with the outbox row listed as future work
(`dbosflow/capture.go:81-97`). Every dual-write property (B1/B2/B3
kill-matrix, idempotency, marker binding) is proven over test-seeded
fiction rows, never real capture output. Valid for the worker mechanism;
says nothing about the capture↔outbox atomicity the POC claims as its
core (ADR-0203 single-txn path).

**IM-14 (P2, mixed). Ledger `Record` carries no PII — but also no real
provenance: all digests in tests are fiction.** Actual struct
(`ledger/ledger.go:24-38`): event/artifact UUIDs, table/id refs,
employer, sha256, rule versions/digests, calc digest, decided_by/at,
supersedes. No names, NRIC/UEN, bank details, or amounts. But every
executed path stubs `payload_hash=ab`, `rule_pack_digest='d'`,
`calc_digest='c'` (`e2e/chaos_test.go:68-74`); the real digest
computations (`synth.go:130-136`, `workload.go:337-343`) never flow into
any outbox row — no code wires `workload → outbox → ledger`. Honest
statement: "ledger bytes contain UUIDs + digests + operator label; the
digest↔source binding is currently unwired."

**IM-15 (P3). Marker value DOES bind event_uuid + content digest;
aliasing fails loudly — with two caveats.** Derivation:
`ledger.go:127` (`rec.EventUUID + ":" + rec.Digest()`, sha256 over
canonical Record JSON, `:42-54`); `IdemKey` at `:65-67`; preconditions
on both keys (`:139-142`); `NoWait:false` (`:143`). Recovery compares
exact strings and returns the entry's origin tx
(`worker/outbox.go:166-178`). Caveats: (a) doc comment at
`ledger.go:114-116` says "TxScan" but the code uses `Get` — stale, and
`TxScan` appears nowhere else in `poc/`; (b) same-key-different-payload
across different rows is impossible by construction (`op_uuid` PK feeds
the key), so the aliasing tripwire only fires on payload mutation or
caller bugs — narrower than comments imply. No aliasing window in the
marker scheme itself.

## 7. Comparative baseline fairness

**PostgreSQL baseline: unfair.** See PG-9/RC-4: Arm A is an uncommitted
ad-hoc probe on an undescribed scratch table, "cleaned up after", scored
"PROVEN both arms". A bare table tests no validation at all; a competent
app checker's real failure mode is TOCTOU races, untested. Production
already rejects same-code compensation overlaps via trigger — the honest
comparison (WITHOUT OVERLAPS vs existing trigger) was never run. Committed
evidence uses pack probe tables, not compensation records; the "gap" half
has no Arm A/B evidence (only the DB-free `demo-gap` mini-case,
`workload.go:194-197`). Overlap detection vs gap/coverage detection are
conflated under one "PROVEN".

**DBOS baseline: fairly disclosed as absent.** Zero conventional-job code
exists (search hits only prose admissions). Reports consistently say NOT
BUILT. The comparison is DBOS-measured vs recovery-inventory-argued, not
measured-vs-measured — correctly labeled. Minimum credible mechanisms an
ordinary Go+PostgreSQL worker would need (per the report's own inventory,
accepted as accurate): dedup table, checkpoint rows, fencing, restart
coordinator. DBOS eliminates checkpoint/dedup/restart machinery; PG
retains capture-transaction exactly-once; the app retains idempotent
outbox design. `RunAsStep` side effects are not claimed exactly-once —
correctly.

**immudb baseline: fair on the role axis, overread in conclusion.** Code:
`e2e/poc_test.go:143,167` (`DELETE FROM poc_outbox` as admin must
succeed or the test fails) vs `:161-166` (app DELETE fails 42501);
backlink DELETE fails P0001 as the same role (`:257-266`). Same role,
same op, different outcome proves the trigger does the work. But
`poc_outbox` was never promised owner-immutability, so "ordinary audit
rows are rewritable" is proven only of a table that was never hardened —
and the deleted row is synthetic cleanup data, not an audit row under
test. The contrast is load-bearing cleanup, not an asserted test step.
A fair PG-only baseline would harden equivalently or state the precise
invariant: "grants stop `app` (42501); only the trigger stops the owner
(P0001)." The relevant comparison — whether PostgreSQL alone can
establish historical integrity against an adversary beyond its admin
boundary — is weakened twice: the PG side is deliberately soft, and R2
places a hostile host admin out of scope anyway. Do not use this
deliberately weak baseline to overstate immudb's incremental value.

## 8. Payroll output and reproducibility verification

**RC-1 (P0). The DB-backed month computes nothing; canonical digest
`3cef505e` is unreachable from any DB-backed path.** Canonical digest is
produced DB-free only: `canonicalCmd()` → `workload.Canonical(...)` →
`synth.ComputePay` per employee (`cmd/poc-run/main.go:395`,
`workload/workload.go:306`; comment at main.go:396: "DB-free: the
canonical month runs through synth.ComputePay only"). The DB-backed month
(`dbosflow/month.go:26`) runs five `RunAsStep` phases calling `runPhase`,
which ignores its context and returns `"period:phase"` labels
(`month.go:52-53`); the TODO above (`:48-51`) confirms bodies don't
exist. `captureTxn` executes only `SELECT 1` and returns
`"placeholder:"+runID` (`capture.go:87-97`, "mechanical placeholder
proving the atomic path shape"). `hashVerify` echoes its input
(`:101-106`, TODO). `runMonthWorkflow` records only workflow ID +
phase-name strings (`main.go:228-229`). No code compares DB-backed output
to `3cef505e`; no DB-backed function calls `ComputePay`/`Canonical`
(callers are `canonicalCmd`, `demoMonth`, tests only). Rows written by a
DB-backed run: seeded *inputs* (separate `seed-poc` command) +
DBOS-internal history. Zero payroll outputs persisted by month/capture.
Sharing between paths is input-only (seeder reads `workload.Build()`).
Reports frame "one shared synthetic month … through three comparative
experiments" adjacent to "Full month … overall: pass" and "Canonical
month … digest 3cef505e", with §4 "No duplicate or partial financial
artifacts" and F1b/F5 PASS — inviting the conclusion that the DB-backed
month computed 24 payslips / 8519957¢ exactly once. The DB-backed compute
phase is a stub returning a label, and "exactly-once financial artifacts"
is vacuous on that path.

**RC-7 (P2). Digest tripwire is self-referential AND unpinned.**
`TestDigestStability` (`workload/workload_test.go:201-216`) asserts
`Canonical(Build()) == Canonical(Build())` — any deterministic wrong
implementation passes. `3cef505e` appears in zero `.go` files
("re-pinned" means docs edited, not a golden updated). Independent
verification exists only for tiny hand values (OT=937/ded=10000,
`workload_test.go:95-108`; 5400/7200 boundary, `:113`; base total
8072000¢, arithmetic confirmed). The full-month digest has no independent
check. The honest hedge "(stability tripwire, NOT independent
verification)" appears only in `PHASE2-WORKLOAD.md:49-50` — neither
Gate C-facing report carries it, while F1b claims "re-derivation incl.
calc digest … PASS" on rerun-identical (same code twice, plus
capture-dedup of a placeholder string, `e2e/poc_test.go:368-381`).
Determinism ≠ re-derivation.

**RC-10 (P2). F5 "provenance attests" overclaims.** All outbox rows come
from test-only `seedOutbox` (`e2e/chaos_test.go:62-78`) with hardcoded
`"rule_pack_digest":"d","calc_digest":"c"` — never from month/capture
workflows (which enqueue nothing). No backlink binds `3cef505e`, v1a/v2a,
or any employee result. Mechanism (marker-bound idempotency,
backlink-tx == marker-origin) genuinely proven
(`chaos_test.go:197-217`); content provenance of the payroll is not.
"Old re-derives" has no swap-and-re-derive execution (only the
single-month 5400/7200 boundary observation).

## 9. Report-versus-code discrepancies

**RC-2 (P1). Internal contradiction on fabrication blocking.**
`POC-FINAL-REPORT.md:58`: "REAL blocked only downstream
(claims/backlink/ledger), table-level fix still RED." Same report lines
61-66: "the procedural verifier poc_verify_claims is BROKEN (42703 …
proven live…). T10/T11 claim proof stays UNPROVEN." A broken verifier
blocks nothing. Backlink/ledger detect rewrites of anchored rows, not
fabrication of unanchored snapshot rows.

**RC-3 (P0). §6 execution evidence is missing from the file.**
`POC-FINAL-REPORT.md` ends at lines 143-146 in a literal
`...[truncated 6186 chars]`. Stack versions, E5 failure enumeration,
runbook outputs, and "harness lessons" referenced by `POC-REPORT.md:201`
are absent from the primary Gate C record. Every §6-dependent claim
(notably E5 composition) is unverifiable from the file itself. Verified
in the file itself (file is 145 lines, 9885 bytes).

**RC-8 (P2). Count errors: "synth 10/10" is wrong (9 exist, all green).**
Executed DB-free suites: synth shows 9 `--- PASS` lines (7 in
`synth_test.go` + 2 in `rules_data_test.go`; no `Example`/`TestMain`),
workload 9/9. All three reports claim "synth 10/10". e2e: 13 top-level
`Test` funcs confirmed (8 poc + 1 forged + 1 idempotency + 3 chaos) —
"13/13" accurate at Test-function level. Minor: `POC-RESULTS.md:15`
parenthetical enumerates only 12 (omits `TestCrashMatrix`).

**RC-9 (P2). E5 "zero version deltas" reconciles only with mixed units;
flake identities unlocatable; "11 RED + 1 TZ" composition verified.**
Verified: 11 RED-bearing Test funcs, all in
`internal/payroll/snapshot_admission_test.go` (12 funcs;
`TestWriterFirstEmployeeCreation` has no RED site; 12 RED call-sites
across the other 11 — matches "11 intentional RED (D1/T2/EF-E3/§A/§B)").
`TestSnapshotFabricationByApp` ends in unconditional `t.Fatalf("RED…")`
— fails identically on any PG version, cannot hide a version delta. TZ
test asserts `occurred_at::text` contains "09:00:00"
(`internal/attendance/m11_test.go:161-168`) — genuinely
session-TZ-sensitive, correctly labeled env artifact. Arithmetic: main
module has 26 packages, 19 with tests — "16 ok / 15 fail" only closes as
16 ok packages + 15/14 failing tests; the unit switch is never stated.
"Gate2 load flakes" appear only in prose; no matching test exists in
`tests/`, `internal/`, or `poc/`. "Zero PG18-specific failures" is
properly a delta claim, but its basis is unverifiable from surviving
text. Medium confidence that flakes (not hidden deltas) explain the
15-vs-14 gap.

**RC-5 (P2). Exp2A confirmed nonexistent; honest gap, correctly
labeled.** No fairness issue beyond the acknowledged measured-vs-argued
comparison.

**RC-6 (P2). Exp3A privileged-DELETE contrast is real but implicit; role
is the migrate/admin role.** `e2e/poc_test.go:143,167` admin DELETE must
succeed; `:161-166` app DELETE fails 42501. Admin pool =
`POC_DBOS_MIGRATE_URL` role (`migration_admin` per runbook). Second
instance: `chaos_test.go:167`. "PROVEN by contrast" is fair, but weaker
than a dedicated assertion.

**RC-11 (P3). Smaller overreaches and nits.** F1a "DB-backed blocking via
gate + constraints live": the gate serializes employers and constraints
guard probe tables, but no B-* blocker is enforced on any DB path —
conflation of generic mechanics with blocker enforcement. §4 inj #2
"PROVEN on the worker side": accurate for harness-seeded rows; fine.
Exp3 PASS with genesis pending and host-corruption unbuilt: generous but
disclosed; acceptable as PASS (gaps noted). Stale comment:
`seedpoc.go:31` says "SYN-022 (zero clocks)" but the edge is `e==22` →
SYN-023 (`workload.go:148-149`); reports correctly say SYN-023.
Spot-checks that passed: 29 leave rows (22+5+2, incl. SYN-024
double-match), base total 8072000¢, Sunday-rest guard +
`TestSundayRest`, demo digest `259564d6` kept distinct from canonical.

**RC-4 (P1). Exp1A baseline is a strawman with no committed code.**
Covered in §7.

## 10. Detailed finding register

| ID | Sev | Area | One-line description | Primary refs |
|----|-----|------|----------------------|--------------|
| PG-1 | P0 | PG | Valid-parent fabrication → direct-SQL validated run as app | 00012:1094-1099,1225-1262; provision.go:117 |
| RC-1 | P0 | recon | DB-backed month computes nothing; digest DB-free only | month.go:52; capture.go:87; main.go:395 |
| RC-3 | P0 | recon | §6 evidence missing (truncated stub in file) | POC-FINAL-REPORT.md:143-146 |
| IM-6 | P0 | immudb | Bogus-tx test proves nonexistent-tx errors, not tamper | poc_test.go:279-288 |
| IM-7 | P0 | immudb | Auditor checks tx existence only; PG tamper invisible | auditor/client.go:37; auditor/main.go:112 |
| PG-2 | P1 | PG | Close dead for app (42501) yet directly bypassable | gate2_poc.sql:137-171 |
| PG-3 | P1 | PG | Closed-period writer check never fires (dead predicate) | gate2_poc.sql:47-56 |
| PG-4 | P1 | PG | Claims verifier broken/uncalled/unwritten; vacuous if fixed | gate2_poc.sql:105-134 |
| PG-5 | P1 | PG | Outbox forgeable, backlink squattable, no producer | outbox.sql; outbox.go:51-160 |
| PG-6 | P1 | PG | No capture path takes the gate | snapshot.go:417; capture.go:87 |
| DB-1 | P1 | DBOS | Atomic txn mechanism real; membership empty | capture.go:87-97; lib workflow.go:2994 |
| DB-2 | P1 | DBOS | Exactly-once assertions count status rows, not effects | chaos_test.go:245; idem_test.go:73 |
| DB-3 | P1 | DBOS | Same month ID resubmission returns stale result silently | main.go:218; lib sysdb.go:1493 |
| DB-4 | P1 | DBOS | ERROR sticky; month ID permanently poisoned | lib workflow.go:1513-1516 |
| DB-5 | P1 | DBOS | Forged-SUCCESS pre-registration untested; vacuous tripwire | forged_test.go:28-75 |
| IM-9 | P1 | immudb | Recovery reads marker via unverified Get | outbox.go:166; ledger.go:161 |
| IM-11 | P1 | immudb | Backlink PK silently drops corrections | outbox.sql:31; outbox.go:137 |
| IM-12 | P1 | immudb | Outbox/backlink grant-only; app forges completion | outbox.sql:24-26,50-51 |
| RC-2 | P1 | recon | "Blocked downstream" contradicts "verifier BROKEN" | FINAL-REPORT.md:58 vs :61-66 |
| PG-7 | P2 | PG | E2 "every spelling" was one spelling | poc_test.go:106-130 |
| PG-8 | P2 | PG | E5 IDENTICAL: counts differ, evidence stub | FINAL-REPORT.md:13; POC-REPORT.md:192 |
| PG-9 | P2 | PG | Exp1A strawman; production already blocks attack | 00011 trigger; FINAL-REPORT.md:56 |
| PG-10 | P2 | PG | Trigger gaps vs spec; DELETE bypasses; T1-T12 unrun | gate2_poc.sql:83; ADR-0203 |
| PG-11 | P2 | PG | Forged-DBOS attribution unproven; mod-10 fingerprint | forged_test.go:79-95 |
| DB-6 | P2 | DBOS | Sticky-ID fix tests different regime than production | chaos_test.go:227; main.go:242 |
| DB-8 | P2 | DBOS | No approval wait exists anywhere | month.go:22-25 |
| IM-4 | P2 | immudb | Writer state-dir leaks; non-hermetic harnesses | chaos_test.go:161,199; ledger.go:77 |
| IM-5 | P2 | immudb | Zero server flags: no signing key, TLS, users | compose.poc.yml:31-52 |
| IM-8 | P2 | immudb | Host-corruption harness absent (acknowledged) | POC-REPORT.md:203-207 |
| IM-10 | P2 | immudb | Poison path orphans ledger records silently | outbox.go:85-95 |
| IM-13 | P2 | immudb | No production enqueue path; fiction-row evidence | capture.go:81-97 |
| IM-14 | P2 | immudb | No PII in ledger, but no real provenance either | ledger.go:24-38; chaos_test.go:68 |
| RC-4 | P2 | recon | Exp1A: see PG-9 (baseline fairness) | FINAL-REPORT.md:56 |
| RC-5 | P2 | recon | Exp2A absent; honestly labeled (no action) | POC-RESULTS.md:23 |
| RC-6 | P2 | recon | Exp3A contrast real but implicit | poc_test.go:143,161-167 |
| RC-7 | P2 | recon | Digest tripwire self-referential and unpinned | workload_test.go:201-216 |
| RC-8 | P2 | recon | "synth 10/10" wrong; 9 exist, all green | synth_test.go (7); rules_data (2) |
| RC-10 | P2 | recon | F5 overclaims; backlinks anchor dummy digests | chaos_test.go:62-78 |
| PG-12 | P3 | PG | Self-held gate DoS via granted advisory locks | gate2_poc.sql:247-263 |
| PG-13 | P3 | PG | Reproducibility test proves dedup of placeholder | poc_test.go:365-381 |
| PG-14 | P3 | PG | Claim/close-table squat surface (latent) | gate2_poc.sql:102-103 |
| DB-7 | P3 | DBOS | Shutdown closes shared pool; abandon-workaround leaks | close.go:21-32 |
| DB-9 | P3 | DBOS | Step-window test replays pure functions only | month.go:52; chaos_test.go:255 |
| DB-10 | P3 | DBOS | Stdout contamination residuals (lib Printf) | dbos.go:computeApplicationVersion |
| DB-11 | P3 | DBOS | Convergence-as-error; dead appended; poison counting | outbox.go:146-155,75-84 |
| DB-12 | P3 | DBOS | Non-transactional seeder, refuse-to-rerun | seedpoc.go:64-157 |
| IM-15 | P3 | immudb | Marker scheme sound; stale TxScan comment | ledger.go:114-116,127 |
| RC-11 | P3 | recon | Minor wording overreaches and nits | seedpoc.go:31; F1a wording |

Reproduction: all findings reproduce by reading the cited file bodies; no
live stack is required. The three highest-value live confirmations (all
disposable-stack only): (1) run the PG-1 four-statement chain as `app`
and observe `validated`; (2) mutate one PG snapshot row post-backlink and
observe the auditor still passing (IM-7); (3) resubmit a corrected month
under the same month ID and observe the stale result (DB-3).

## 11. Technology-by-technology adoption reassessment

**PostgreSQL 18.6: prior ADOPT → DOWNGRADE to CONDITIONAL ADOPT for POC
continuation.** The exclusion constraint, advisory-gate primitive, and
backlink trigger are genuinely demonstrated. But the POC's headline PG
value proposition — temporal coverage and Rung 5 snapshot authority — is
contradicted, not merely unproven: PERIOD FKs don't exist on 18.6, and
the fabrication-to-validated chain runs through the real schema
untouched. The honest 18.6 value shown here is baseline consistency
(migrations apply, suite behaves equivalently modulo flakes), not new
temporal functionality. Condition: close PG-1/PG-2/PG-3 or explicitly
descope Rung 5 from the POC's claims before any production-ward step.

**DBOS: CONDITIONAL ADOPT → HOLD (do not advance the condition set).**
The shared-pool atomic mechanism is verified and worth keeping as the
POC's execution layer. But every production-relevant semantic around it
is either vacuous (empty txn), wrong for the stated ID design (stale
success, ERROR poisoning), or untested where it matters (forged-SUCCESS
pre-registration). No new condition is satisfiable until capture bodies
exist; advancing DBOS now means blessing semantics proven only over
placeholders. The missing Exp2A baseline stays a gap — do not let
"recovery inventory argued" harden into "measured".

**immudb: CONDITIONAL ADOPT → HOLD.** The append-only/verified-append
mechanism and marker construction are sound, and the backlink trigger
genuinely binds the owner. But "independent verification" as built is
tx-existence checking over TOFU trust with shared superuser credentials
and no PG↔ledger binding — the auditor cannot detect the attacks the
runbook says it must. Combined with the still-pending BSL counsel review
and the R2 exclusion of hostile host admins (which concedes the
strongest version of the tamper question), immudb's incremental value
over a hardened PG audit table is asserted, not measured. Condition set
cannot shrink until genesis custody is real, the auditor binds content,
and at least the disposable host-corruption test executes.

**Combined architecture: prior ADOPT-as-direction → DOWNGRADE to RETAIN
AS EXPERIMENTAL DIRECTION ONLY.** The layering (PG truth, DBOS
execution, ledger evidence) remains coherent on paper and the POC
disproved nothing about the shape. But no end-to-end property — compute
once, gate writes, attest content, verify independently — has been
demonstrated over real payroll data flowing through all three layers,
because the layers were never connected: capture writes nothing, nothing
enqueues, the ledger anchors dummies, the auditor checks existence.
Retain the direction; reset every guarantee claim to unproven.

Scope discipline: nothing above authorizes production architecture,
production migration 00013, hosted/paid immudb use, or any statutory
reading of the synthetic outputs. OT-001 stays BLOCKED; that boundary was
respected throughout.

Guarantee classification summary (PROVEN / PARTIALLY PROVEN /
NOT PROVEN / CONTRADICTED):

- No-overlap enforcement: PROVEN (exclusion works live).
- Period coverage declared: CONTRADICTED (PERIOD FK absent; verifier
  broken).
- Snapshot fabrication blocked: CONTRADICTED (app mints valid-looking
  snapshots).
- Validated-run integrity: CONTRADICTED (direct SQL validates invented
  tips).
- Employer-gate serialization: PARTIALLY PROVEN (primitive only; writers
  ungated).
- Close atomicity/exclusion: NOT PROVEN (close dead yet bypassable).
- Capture atomicity with DBOS: PARTIALLY PROVEN (mechanism real; body is
  SELECT 1).
- Exactly-once business effects: NOT PROVEN (status rows counted, not
  effects).
- Dedup under resubmission: CONTRADICTED (silent stale success; ERROR
  sticks).
- Durable approval waits: NOT PROVEN (no wait exists in workflow).
- Forged-state resilience: NOT PROVEN (tripwire tests wrong attack).
- Outbox idempotent recovery: PARTIALLY PROVEN (mechanism sound;
  fiction rows only).
- Independent audit verifies: CONTRADICTED (tx existence only; TOFU
  genesis).
- Ledger tamper detected: NOT PROVEN (bogus-tx test proves no such
  thing).
- Host-corruption detected: NOT PROVEN (harness unbuilt; R2 excludes
  admin).
- Backlink append-only: PROVEN (P0001 binds even owner).
- No PII in ledger: PROVEN (UUIDs plus digests only).
- Provenance binds payroll: NOT PROVEN (dummy digests anchored).
- Canonical reproducibility: PARTIALLY PROVEN (deterministic; digest
  unpinned).
- DB-backed output equals canonical: CONTRADICTED (DB path computes
  nothing).
- PG18 vs PG16 identical: PARTIALLY PROVEN (plausible; evidence stub;
  flakes).

## 12. Final recommendation

Do not archive this POC as a success, and do not discard it either.
Archive it as what it is: a well-instrumented scaffolding run that proved
mechanisms (exclusion constraints, advisory serialization, shared-pool
atomic shape, verified appends, append-only backlinks) while refuting or
leaving unproven nearly every end-to-end guarantee. Before any Gate C
adoption — even POC-continuation adoption — the cheapest decisive work is:
wire one real capture body through gate → snapshot → outbox → ledger →
backlink → content-binding audit, then re-run the three §2 questions
against that connected path. If PG-1, DB-3/DB-5, and IM-1/IM-7 survive
contact with a connected path, the architecture keeps its direction.
Until then, every PROVEN in the current reports that touches snapshots,
exactly-once effects, or independent verification should be read as NOT
PROVEN.

Remediation pointers (concrete, per finding): PG-1 needs the documented
revoke-plus-DEFINER-capture fix in a real migration proposal (not POC
SQL) with the fabricate-then-transition battery row; PG-2 needs
`SECURITY DEFINER` plus revoke of direct INSERT (or a legitimized
insert-then-close protocol, not both shapes at once); PG-3 needs the
writer check rewritten against real columns (`pay_period_id` joins, not
`->>'period'`); DB-3/DB-4 need an explicit reuse policy plus input-drift
rejection or versioned month IDs; DB-5 needs forge-then-run-same-ID;
IM-1 needs committed genesis plus enforcement flags; IM-7 needs
content-binding verification (recompute digest from PG, compare to ledger
value) before any tamper claim; RC-3 needs §6 rewritten from any
surviving session logs or explicitly marked lost.
