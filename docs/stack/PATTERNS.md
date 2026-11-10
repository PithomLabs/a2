# PATTERNS — pattern catalogue for the PG18 + DBOS + immudb stack

Status: RESEARCH GATE — patterns PROPOSED for POC use, not yet proven.
Evidence: [RESEARCH.md](RESEARCH.md). Business context: [DECISIONS.md](DECISIONS.md).
Binding constraints: synthetic-only POC; DBOS at-least-once-until-proven;
native immudb SDK primary; migrations 00001–00012 frozen.

Tags: VERIFIED / DOCUMENTED / INFERRED (researcher evidence); ASSUMPTION
(synthesis beyond evidence).

## 1. Decision matrix (concern × owner × why × anti-pattern × failure-if-misapplied)

| # | Concern | Owner tech | Why this owner | Anti-pattern | Failure if misapplied |
|---|---|---|---|---|---|
| C1 | Operational truth (mutable state, relations, effective dates) | PG18 | Relational constraints, temporal keys, txn isolation; immudb wire reports PG14 + skips DDL (DOCUMENTED Phase 4 §5) | Storing HR entities primarily in immudb; driving payroll reads from ledger | Silent-skip DDL corruption; no effective-dating; erasure impossible |
| C2 | Temporal no-overlap (plain tables) | PG18 WITHOUT OVERLAPS | Declarative key, optimizer-visible, same enforcement as exclusions (DOCUMENTED Phase 2 §1.1) | Hand-rolled overlap checks in Go | Race window; direct SQL bypasses service (AGENTS.md invariant rule) |
| C3 | Containment / joint coverage (comp ⊆ terms, pack covers period) | PG18 PERIOD FK | Combined-coverage semantics no app code can match atomically (DOCUMENTED Phase 2 §1.2) | App-side coverage validation only | TOCTOU: pack revoked between check and capture |
| C4 | Live-basis successor logic (comp/SHG) | PG triggers + DEFINER fns | "An exclusion cannot express liveness" (VERIFIED Phase 2 §1.1) | WITHOUT OVERLAPS for successor chains | Superseded history collides; corrections unrepresentable |
| C5 | Capture atomicity (gate + read + write + checkpoint) | PG txn inside DBOS `RunAsTransaction`, shared pool | Only path where business write + durability record commit atomically (DOCUMENTED Phase 3 §2) | Ordinary `RunAsStep` around capture | Crash between write and checkpoint → ambiguous state, re-execution risk |
| C6 | Month orchestration (schedule, waits, approvals, fan-out) | DBOS workflows | Durable sleep/Recv/queues/cron survive crashes (DOCUMENTED Phase 3 §2–3) | Cron + Go sleep + polling tables | Crash loses waits; duplicate wakeups; no recovery story |
| C7 | External side effects (ledger append, payment-file preview) | DBOS `RunAsStep` + idempotent sink | At-least-once is the proven guarantee (DOCUMENTED Phase 3 §2: verbatim) | Treating a step as exactly-once | Crash after effect, before checkpoint → duplicate ledger/payment records |
| C8 | Tamper-evident provenance (hashes, attestations, close records) | immudb standalone, native SDK | Dual proofs + external auditor; writer cannot rewrite checkpoint (DOCUMENTED Phase 4 §3); event-uuid keys + append-only PG backlink table (ADR-0204 Gate A) | Plain PG audit table as tamper evidence; PII in ledger; `immu_tx` in payload (circular); UPDATE back-links on frozen rows; reusable prov keys | EF-E3-class fabrication undetectable; GDPR erasure impossible; trigger rejection; evidence aliasing |
| C9 | Money/statutory purity | Go `rules/` + PG NUMERIC | Arch-enforced purity; exact intermediates (DOCUMENTED AGENTS.md §§2–5) | DBOS JSON `any` payloads carrying `math/big` without codec | Precision loss through serialization (INFERRED Phase 3 §5) |
| C10 | Writer discipline (who may write what) | PG revokes + triggers + DEFINER | Service-only fixes bypassable via direct SQL (AGENTS.md mandatory principle) | App-role checks without DB enforcement | P1a–g / EF-E3 / F1 recurrence (DOCUMENTED M2 §6–8) |
| C11 | Crash recovery across PG↔ledger | PG outbox + DBOS durable worker + immudb preconditions | No native atomicity spans PG↔immudb (binding plan risk) | Dual-write without outbox/idempotency | Missing or orphaned ledger records after crash (INFERRED Phase 4 §4) |
| C12 | Rule dynamics (new rule types, pack versions) | Effective-dated PG data + versioned packs + DB coverage gate | Rates/thresholds are data, never constants (DOCUMENTED AGENTS.md §4) | Hardcoded SYNTH rules in Go for POC convenience | POC proves nothing about rule dynamics; rewrite at first real rule |

## 2. Claim inventory with rungs

Rung ladder (DOCUMENTED [decisions.md](/home/chaschel/Documents/go/a2/plans2/checklist/decisions.md:347-365)
per Phase 0 D4): Rung 5 = DB-enforced; lower rungs = service/convention.
Every claim below names its rung; Rung 5 claims need a DB mechanism, not prose.

| # | Claim | Rung | Mechanism (proposed) | Proven by (proposed test) |
|---|---|---|---|---|
| K1 | Post-cutoff postings into closed population rejected | 5 | Close-state trigger + DEFINER close txn (INFERRED) | Translated battery T12 |
| K2 | Finalized run bytes never change; corrections are linked adjustments | 5 | Status-gated freeze triggers (DOCUMENTED 00004-pattern) | Byte-compare original after adjustment |
| K3 | Snapshot/validation creatable only via DEFINER path | 5 | Revoke app INSERT/UPDATE on snapshot*/pins (DOCUMENTED M2 fix) | Battery E3/F1 translations (T9–T11) |
| K4 | Empty/vacuous validation impossible | 5 | ≥1-employee-row + coverage triggers (INFERRED) | Vacuous-pass probes |
| K5 | Comp line ⊆ parent terms; pack covers full period | 5 | PERIOD FKs (DOCUMENTED PG18 semantics) | Gap/overhang rejection demos (E2–E3) |
| K6 | Capture is atomic incl. durability record | 5 (PG txn) + workflow | Shared-pool `RunAsTransaction` (DOCUMENTED Phase 3) | Kill -9 pre-commit → zero partial rows |
| K7 | Ledger appends idempotent; no dup/miss across crashes | 4–5 (app+ledger) | Outbox + event-uuid keys + `PreconditionKeyMustNotExist` on both keys + digest-bound marker (ADR-0204 Gate A) | Crash matrix 2–4 + re-anchor dedup + same-key-different-payload loud fail |
| K8 | Ledger tampering detected loudly | 5 (crypto) | Dual proofs + independent auditor (DOCUMENTED) | On-disk tamper → `ErrCorruptedData` + webhook |
| K9 | Mid-period change blocks until proration specified | 5 | Coverage triggers → typed blocker (INFERRED) | Mid-period comp/SHG fixtures |
| K10 | Rerun from pinned inputs is byte-identical | 4 (determinism) | Canon bytes + pinned versions (DOCUMENTED pattern) | Re-derivation script equality |

## 3. Ownership table

Canonical technology-allocation table (MIGRATION.md §5 summarizes it).

| Layer | Technology | Owns | Must never own |
|---|---|---|---|
| Operational truth | PG 18.6 | Mutable state, relations, effective dates, constraints, concurrency, inputs, results, period state | Tamper evidence; durable waits; cross-crash workflow state |
| Durable execution | DBOS Transact Go | Orchestration, checkpoints, retries, recovery, queues, schedules, approval waits | Business invariants (barriers stay in PG); money math; exactly-once externals |
| Verifiable evidence | immudb (standalone, native SDK) | Hashes/digests, rule stamps, attestations, close records, receipts | PII; primary entities; operational/churn state; effective-dating |

Hypothesis source: `prompt.md` §6; validated against Phases 1–4 evidence in
[RESEARCH.md](RESEARCH.md). ASSUMPTION: allocation holds for POC scope.

## 4. Data-flow description (candidate §6A sequence, annotated)

`prompt.md` §6A candidate flow, with evidence annotations (ASSUMPTION: flow
adopted for POC; each step's guarantee tagged):

1. Business event / payroll request enters (DBOS queue or CLI/API → service).
2. DBOS initiates/resumes workflow (durable; dedup ID = `run:{runID}` —
   DOCUMENTED dedup semantics Phase 3 §2).
3. PG validates business state + applicable rules (Rung 5 constraints fire
   inside txn — DOCUMENTED pattern).
4. PG commits authoritative operation (single txn with Gate-2 employer gate +
   proof-check — Gate 2 ADR DOCUMENTED, NOT implemented).
5. App writes evidence/outbox event in the SAME PG txn (outbox pattern —
   INFERRED Phase 4 §4; no native PG↔immudb atomicity).
6. DBOS durable worker processes outbox event (at-least-once — DOCUMENTED).
7. immudb records immutable evidence via idempotent `ExecAll` (atomic
   server-side; precondition dedups — DOCUMENTED).
8. App INSERTs immu tx ref into append-only `poc_ledger_backlink` in PG
   (only after `VerifiedTxByID` — INFERRED Phase 4 §4, reshaped Gate A).
9. Verification/reconciliation: auditor re-derives snapshot hash from sources +
   pack, dual-proof verify (DOCUMENTED primitives; script INFERRED).

Dual-write crash matrix: after-PG-commit/before-ledger → outbox replays, no
loss; after-ledger/before-back-link → precondition says already-appended,
recover `immu_tx`, fill back-link; ambiguous network failure → retry with same
idempotency key, sink dedups (all INFERRED Phase 4 §4 — POC must prove).

## 5. Failure matrix (component × step)

| Failing component | Fails at step (§4) | Expected behavior (PROPOSED) | Proved by |
|---|---|---|---|
| App worker | 4 (PG txn pre-commit) | Single-txn rollback; zero partial rows; DBOS replays from checkpoint | Kill -9 experiment 1 |
| App worker | 6–7 (after effect, before checkpoint) | Re-execution; sink idempotency absorbs duplicate | Kill -9 experiment 3 |
| App worker | 8 (after ledger, before back-link) | Outbox open; recovery re-appends (precondition hit → marker digest compared: equal recovers tx via TxScan and INSERTs backlink; different fails loudly) | Crash matrix 2–4 |
| PG | any | Workflow step fails; DBOS retry/backoff; no ledger write without PG commit | PG-down experiment |
| immudb | 7 | Append fails loudly; outbox stays open; PG truth unaffected; retry on recovery | immudb-down experiment |
| immudb disk | post-commit | Auditor dual-proof fails `ErrCorruptedData` + webhook; PG truth intact | Tamper experiment |
| DBOS executor | mid-step | Ownership fencing: late checkpoint refused, winner stands (DOCUMENTED) | Experiment 5 |
| Clock/skew | waits | Sleep/Recv resume remainder only (DOCUMENTED) | Experiment 4 |

## 6. New-feature decision table

For any proposed feature, walk this table in order (ASSUMPTION: synthesis rule):

| # | Question | Yes → | No → next |
|---|---|---|---|
| 1 | Is it a business invariant direct SQL could violate? | PG Rung 5 (constraint/trigger/DEFINER) | 2 |
| 2 | Is it effective-dated business data (rate, threshold, coverage)? | PG18 temporal data + coverage gate | 3 |
| 3 | Must it survive crashes mid-process (waits, retries, fan-out)? | DBOS workflow/step (idempotent externals) | 4 |
| 4 | Must third parties verify it after the fact (hash, attestation)? | immudb hashes-only record | 5 |
| 5 | Is it pure calculation? | Go `rules/` (stdlib + math/big, golden-tested) | 6 |
| 6 | Otherwise | Plain Go service + PG, no new mechanism | — |

Rules: no feature may add a second enforcement of an invariant owned elsewhere
(C-table); no new advisory lock without purpose/identity/order/boundary
(AGENTS.md); every new writer joins the writer matrix + battery or the feature
is incomplete.
