# POC Final Report — Three Experiments, One Monthly Payroll

Status: EXECUTED 2026-11-10 (Gate B live). All three experiments ran with
measured evidence; verdicts below are per-experiment PASS/PARTIAL with gaps
named, never rounded up. No result is claimed without a live observation.

Plan: `plan_prompt6.md` + review addendum `plan_prompt6_review.md` (APPROVED
with clarifications A/B/C, all accepted). Branch `stack-poc`, uncommitted
(no commit/push/merge without explicit authorization — preserved).

Suite headline (POC module, disposable PG18.6 + immudb 1.11.2 stack):
e2e 13/13, synth 10/10, workload 9/9, 0 skipped, `go vet` + `gofmt` clean.
Main-module battery (E5): PG18.6 vs PG16.15 behave IDENTICALLY (details §6).

## 1. Purpose and scope

Demonstrate, through one realistic monthly synthetic payroll workload, why
PostgreSQL 18.6, DBOS Transact Go, and immudb are valuable in their
architectural layers — and what engineering complexity each eliminates —
via three comparative experiments with credible without-technology baselines.
Not a payroll product; no new milestones after this. All outputs are
labeled illustrative/non-statutory; SG-MOM-OT-001 stays BLOCKED; no
statutory calculation exists anywhere in the POC.

## 2. Monthly payroll workload description

One legal employer (POC-MERIDIAN), 25 synthetic employees (SYN-001..SYN-025,
no PII) × 31 days, Oct 2026, employer + terms + rosters + paired clocks +
approved timesheets + leave + SYNTH overtime + as-given allowances/
deductions, reusing `synth.ComputePay` (UNTOUCHED) with a shared importable
fixture builder (`workload/`). Two rule packs: v1a (Oct 1–15, OT 3/2) →
v2a (Oct 16–31, OT 2/1); mid-period pack change is calculable (defined
synthetic semantics). Classification gate per clarification A (full table
in `PHASE2-WORKLOAD.md`): intra-month base change (SYN-012), leave day
carrying minutes (SYN-006/Oct-10), pack gap/overlap, and bad input are
explicit TYPED BLOCKERS (`B-BASE-CHANGE`, `B-LEAVE-WORKED`, `B-PACK-GAP`,
`B-PACK-OVERLAP`, `B-BAD-INPUT`) — no proration rule invented.

DB-backed seeding (`poc-run seed-poc`, attempt 4, exit 0): 25 employees,
5 roster weeks, pay period PP-2026-10, 24 approved sheets (SYN-023 has no
sheet — zero-month by design), 29 leave rows, 1060 clock events.
Full month (`poc-run`, exit 0, `overall: pass`): workflow
`month:poc-meridian:2026-10`, phases open→collect→capture→compute→close.

Canonical month (`poc-run canonical`, DB-free): register (25 rows, 24
final + 1 blocked), 24 payslips, 5 exceptions, payment preview, snapshot,
digest `3cef505ea5eb…caa46ea` (`poc-canonical-v1`), register net total
8519957¢. Demo digest stable `259564d6…` (unchanged since preflight).

## 3. Experiment 1 findings — PostgreSQL integrity (PARTIAL PASS)

Focused-case mapping (prompt6 §3 cases 1–4):

| Case | Arm A (app-only baseline) | Arm B (PG-enforced) | Verdict |
|---|---|---|---|
| 1. Overlap/gap compensation records | Plain table (no constraint): 2 overlapping rows INSERTed and silently kept (live probe, this session) | `UNIQUE (k, eff WITHOUT OVERLAPS)`: overlap rejected with exclusion violation, 1 row kept (live probe + `TestE1…`/`TestCoverage…` green) | PROVEN both arms |
| 2. Timesheet change racing capture | No separate harness (gap, see below) | Gate-2 employer-gate mechanics live: same-employer `poc_gate_acquire` BLOCKS until holder commits; different employer never blocks (`TestConcurrency_WriterRacesCapture` green) | MECHANISM proven; full race UNEXECUTED (needs U4 capture bodies — `captureTxn` is a probe that takes no gate) |
| 3. Direct-SQL fabrication | n/a (same table) | Naive fabrication (random UUIDs) rejected by `payroll_run` FK (live probe). Fabrication with VALID parents is NOT prevented — pinned RED by `TestSnapshotFabricationByApp` (EF-E3), failing identically on PG16 and PG18 | NAIVE blocked; REAL blocked only downstream (claims/backlink/ledger), table-level fix still RED |
| 4. Close racing a business write | n/a | `poc_period_close` SI grants verified; close-race e2e not separately harnessed | GRANTS verified; race UNEXECUTED |

Negative results (also findings): temporal PERIOD FKs do NOT exist on
PG18.6 — every spelling rejected 42601 (`TestE2_PeriodFKJointCover`
tripwire green). Joint need⊆coverage therefore cannot be declared; the
procedural verifier `poc_verify_claims` is BROKEN (42703 `source_table`,
proven live, pinned by Coverage Part 2). T10/T11 claim proof stays UNPROVEN
pending an in-txn JSONB re-port per `gate2_verify_claims`.

What PG prevented (measured): overlapping pack versions (exclusion),
cross-talk between employers (gate serialization), TEMP-based shadow
attacks (TEMP revoked), direct snapshot fabrication without parents (FK),
outbox history deletion by app (42501, no DELETE grant), backlink rewrite
by ANY role incl. owner (P0001 trigger — also Exp3 evidence).

Custom infrastructure avoided: overlap/gap polling validators, app-side
distributed locks for the gate, delete-protection middleware — replaced by
one exclusion constraint, one advisory xact lock, and two triggers.

## 4. Experiment 2 findings — DBOS durability (PARTIAL PASS)

Lifecycle executed: capture workflow (`capture:<run-id>`) + month workflow
(`month:poc-meridian:2026-10`, 5 phases). Shared-pool atomic boundary
VERIFIED: pre-commit SIGKILL leaves ZERO `SUCCESS` rows — business capture
and the DBOS durability record commit together via `RunAsTransaction` or
roll back together (`TestKillMatrix_CaptureBoundary` green).

Injection results (prompt6 §4 list):

| # | Injection | Evidence | Verdict |
|---|---|---|---|
| 1 | Crash before capture commit | Kill at `pre-commit-capture` failpoint → no checkpoint; clean restart → exactly 1 SUCCESS | PROVEN |
| 2 | Crash right after capture commit | Outbox-boundary kills (`after-pg-commit-before-append`, `after-append-before-backlink`): state stays `pending`, no backlink; restart converges to `backlinked`, marker origin == backlink tx | PROVEN on the worker side |
| 3 | Crash after a completed calc step | Timed step-window sweep (5 kills, 50–250 ms) + final run → exactly 1 SUCCESS | APPROXIMATE (probabilistic by design; window not hookable from user code) |
| 4 | Crash while awaiting approval | NOT EXECUTED — no approval wait exists in the workflow shape | GAP |
| 5 | Duplicate submission | Same workflow ID twice → identical results, exactly 1 SUCCESS row (`TestDoubleEnqueue…`); same capture ID twice → byte-identical reports (`TestReproducibility…`) | PROVEN |
| 6 | Restart/recovery | Every kill test restarts clean and converges to a single outcome | PROVEN |

No duplicate or partial financial artifacts: 30 SUCCESS workflows /
98 step checkpoints in history; forged rows in `dbos.*` move NO business
state and a real workflow alongside is unaffected
(`TestForgedState_MovesNothing` green — Exp2 resilience comes from the
executor ignoring forgeries, since same-DB grants are unavoidable per P3).

Baseline gap (honest): the Exp2A conventional-job baseline was NOT BUILT.
The comparison is DBOS-measured vs recovery-inventory-argued (dedup table,
checkpoint rows, fencing, restart coordinator — the mechanisms the kill
tests would need without DBOS), not measured-vs-measured. DBOS costs
measured: `Shutdown` leaks a shared-pool conn (pool `Close` hangs; bounded
`ClosePoolTimeout` workaround in place), INFO logs on stdout (routed to
stderr WARN), workflow IDs sticky across processes (unique IDs per run).

Eliminated vs retained: DBOS eliminates the checkpoint/dedup/restart
machinery; PG retains exactly-once of the capture transaction; the app
retains idempotent outbox design (deterministic `IdemKey` markers) —
`RunAsStep` side effects are NOT claimed exactly-once.

## 5. Experiment 3 findings — immudb evidence (PASS, gaps noted)

Anchored evidence (live): 19 backlinks, max immu_tx 20, via transactional
outbox + native SDK, event-UUID keys (ADR-0204), digest-bound idempotency.
Independent auditor (fresh state dir, explicit genesis, outside writer's
control): txs 18/19/20 all `verified: true`, `overall: pass`, exit 0,
genesis `state_hash 4003ecfe…` (must be committed out-of-band — operator
step, still pending).

| prompt6 §5 demonstration | Evidence | Verdict |
|---|---|---|
| PG-only baseline limitation | Admin `DELETE` on trigger-less `poc_outbox` SUCCEEDS (live, in `TestOutbox…` cleanup) — ordinary audit rows are rewritable by privilege | PROVEN by contrast |
| Digest mismatch detection | Auditor over backlinks file with bogus tx 999999999 fails LOUDLY (exit ≠ 0) | PROVEN |
| Cryptographic proof verification | `VerifiedTxByID` per backlink row in the auditor | PROVEN |
| Independent auditor verification | Fresh state dir + `--accept-genesis`, `overall: pass` | PROVEN |
| Cannot silently replace anchored event | No API path (append-only by construction) | BY CONSTRUCTION; host-file corruption harness UNBUILT (gap) |
| Backlink-missing recovery | `after-append-before-backlink` kill → restart converges; backlink tx == marker origin tx; marker binds op (`opID:` prefix) | PROVEN |
| Ambiguous-retry safety | Deterministic `IdemKey(employer, artifact, op)` marker read before append | PROVEN |
| No-PII audit | Payload carries event UUIDs + digests only (schema inspection) | BY INSPECTION |

PG tamper PREVENTION (distinct from ledger detection): backlink
UPDATE/DELETE rejected with P0001 even for the table owner
(`TestTamperDetection` Part 0). Threat-model limits recorded: ledger
proves a record is unchanged, NOT that the business calculation was
correct; verifier custody + genesis commit remain operator-owned;
immudb BSL-1.1 Change Date (~2030-09-02) needs counsel review.

## 6. Actual commands and execution evidence

Stack: `a2-poc-postgres` (postgres:18.6-alpine, digest-pinned) + 
...[truncated 6186 chars]
## 7. Remediation supersession (prompt8, 2026-11-10, branch `stack-poc`)

This section supersedes §§3–5 where they disagree; the historical
verdicts above stand as the pre-remediation record. Full evidence:
[POC-REMEDIATION-REPORT.md](POC-REMEDIATION-REPORT.md).

- §3 (PostgreSQL integrity): PARTIAL PASS → PASS. Root cause PG-1
  (app-side fabrication) corrected at the database layer: grants
  revoked, DEFINER-only capture/validate/approve/close/enqueue,
  trigger guards; the exploit battery (5/5 green ×2, live) proves
  every fabrication path fails closed.
- §4 (DBOS durability): PARTIAL PASS → PASS. The real month
  workflow (Stages A–G) executes on DBOS with transactional
  capture; crash batteries (B1/error-retry/approval-crash) prove
  recovery converges without duplicates or stale acceptance;
  changed inputs conflict explicitly; E5 PG18≡PG16 zero deltas.
- §5 (immudb evidence): PASS-with-gaps → PASS. Content auditor
  rebuilt and green on the real month; tamper/forgery/missing
  batteries fail loudly; verified recovery; reconciling machine
  with owner-proof guard; corrections with lineage; genesis
  pinned + canary-tested; least-privilege identities; IM-8 executed.
- Phase E (new): October-2026 MERIDIAN demonstration PASSED —
  digest `4beee632…` == canonical, net 8521857¢, close+lockout,
  correction lineage (§8 of the remediation report).
- §11 completion: 14/14 PASSED (remediation report §0). Gates:
  `go test ./...` exit 0 (e2e 638 s), `go vet`/`gofmt` clean,
  2026-11-10.
- Unchanged: OT-001 BLOCKED; baseline NOT RUN; frozen migrations
  untouched; no production claims.
