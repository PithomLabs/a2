# Platform research: PG18 + DBOS + immudb (Parts I–X per `prompt.md` §10)

Status: RESEARCH GATE — report complete, approval pending. Assembles Phases
0–4; POC build (Phase 6) NOT authorized. Detail lives in
[docs/stack/RESEARCH.md](../docs/stack/RESEARCH.md),
[docs/stack/PATTERNS.md](../docs/stack/PATTERNS.md),
[docs/stack/DECISIONS.md](../docs/stack/DECISIONS.md),
[docs/stack/adr/](../docs/stack/adr/),
[docs/stack/DECISION-REQUESTS.md](../docs/stack/DECISION-REQUESTS.md) — this
file cross-links rather than triplicating.

Binding plan: [plans2/report.md](report.md) (synthetic-only POC; research gate
before POC build; DBOS at-least-once-until-proven; native immudb SDK primary;
00001–00012 frozen; `poc/payroll-month/` on `stack-poc`, uncommitted).
Tags: VERIFIED / DOCUMENTED / INFERRED / ASSUMPTION (see RESEARCH.md).

## Part I — Executive Assessment

Existing architecture: single Go binary (`cmd/workforceops`), CLI/API sharing
`internal/` services, PG16 with 12 frozen migrations, advisory-lock capture,
plain-table audit. Principal weaknesses (DOCUMENTED M2 FAIL): D1 stale-RR
admission (P0), 7 P1 barrier gaps, EF-E3 snapshot fabrication (P0),
EF-F1A/B direct/vacuous validation (P1), ad-hoc two-txn capture→validate with
ambiguous crash state, self-asserted hashes with no external anchor, RR capture
with vacuous revalidation, capture-first-only race tests.

Expected benefits per technology: PG18 — declarative temporal keys + joint
coverage FKs for containment/coverage invariants, modern baseline to 2030;
DBOS — crash-safe capture atomicity (shared pool) + durable month
orchestration (cron/Recv/Sleep/queues); immudb — dual-proof tamper-evident
provenance for hashes/attestations/close records.

Recommended architecture: all three with strict ownership (PG truth, DBOS
execution, immudb evidence) — [ADR-0201](../docs/stack/adr/ADR-0201.md).
Adoption: PG18 ADOPT-FOR-POC (conditional E5); DBOS ADOPT-FOR-POC (conditional
pin+build+crash battery+placement ruling); immudb ADOPT-FOR-POC (conditional
license+custody+genesis+tamper battery). See Part X for go/no-go (NO-GO until
human approval).

## Part II — Existing Codebase Architecture

Summarized from Phase 1; full map: [RESEARCH.md §1](../docs/stack/RESEARCH.md#1-current-state-map-phase-1).

- Repository/package map: `go.mod` module `a2` vs `a1` import paths
  (UNRESOLVED U1); packages `core/*`, `internal/*`, `migrations/core`,
  `rules/SG-MOM-OT-001` (BLOCKED, no Go impl).
- Business-domain map: HR (employees/terms/comp/SHG) → attendance → timesheet
  → payroll (runs/readiness/snapshot/validate/drift); WFM operational only.
- Database architecture: 00001–00012 (5484 lines VERIFIED); roles
  `migration_admin`/`app`; DEFINER functions; GiST exclusions; live-basis
  triggers; snapshot + run lifecycle in 00012.
- Existing payroll lifecycle: draft→(capture)→validated… with direct-SQL
  bypasses live (EF-F1A/B).
- Existing invariants: employer-scoped keys, NUMERIC money, insert-only
  history, supersede-not-edit, effective-dated comp/SHG/packs.
- Architectural debt/gaps: writer-matrix P1a–g; D1; EF-E3/F1; payroll-line
  boundary unbuilt (G-PREREQ); temporal fragility (created_at + basis_seq
  dual legs); no durable execution; no tamper ledger; advisory discipline is
  convention.

## Part III — PostgreSQL 18.6 Assessment

Full: [RESEARCH.md §2](../docs/stack/RESEARCH.md#2-postgresql-186-findings-phase-2).

- Feature→requirement: WITHOUT OVERLAPS HELPS narrowly (plain no-overlap
  tables; NOT comp/SHG liveness); PERIOD FK HELPS containment + single-pack
  coverage (multi-row joint cover DOCUMENTED); uuidv7/virtual/OLD-NEW NEUTRAL;
  AIO/skip-scan/checksums/MAINTAIN N/A.
- Temporal constraint design: STORED range column + `UNIQUE (…, eff WITHOUT
  OVERLAPS)`; degenerate `(employer_id, id, eff)` key for containment;
  per-pack coverage table for PERIOD FK (all INFERRED sketches; E1–E3 unexecuted).
- Effective-dating examples: comp containment, pack-validity coverage with
  adjacent-version stitching (researcher sketches).
- Migration 16→18.6: fresh 18.6 cluster + migrate + reload (INFERRED
  recommendation; no production data); `pg_upgrade` 16→18 supported but
  reserved; volume moves to `/var/lib/postgresql` (DOCUMENTED); old 16 volume
  preserved; rollback = restart 16 stack.
- Compatibility risks: E5 unexecuted (migrations + suite on 18.6); pinned
  advisory/`pg_locks` assertions may drift; AFTER-trigger-role change neutral
  here; no DEFINER/advisory/isolation changes (VERIFIED release-page search).

## Part IV — DBOS Assessment

Full: [RESEARCH.md §3](../docs/stack/RESEARCH.md#3-dbos-transact-go-findings-phase-3).

- Go SDK: checkout floor go 1.25.0, pgx 5.10.0 vs a2 5.12.0 (skew INFERRED low
  risk); NO consumable version pin (VERIFIED gap); pool-sharing is the key
  compat point (`sharesSystemDB` DOCUMENTED).
- Workflow candidates: capture pipeline `capture:{runID}`; month-run
  open-to-close (cron + children + approval Recv + Sleep); PG→immudb
  reconciliation (sink idempotent on `(snapshot_id, hash)`); retro adjustments
  as children (no Rewind on approved artifacts).
- Transaction semantics: shared-pool `RunAsTransaction` = atomic business
  write + checkpoint (DOCUMENTED); separate-DB = two txns + replay
  (deterministic `fn` required); `RunAsStep` = at-least-once VERBATIM
  (DOCUMENTED); externals at-least-once only (INFERRED).
- Recovery/concurrency: step-ID + name-match replay guards; Patch-gated
  evolution; Enterprise-only AutoUpgrade; owner_xid fencing (no business locks
  survive); queue SKIP LOCKED dequeue; cancel-mid-step re-executes.
- Recommended integration: capture gate INSIDE one shared-pool
  `RunAsTransaction`; services need `*pgxpool.Pool` → `Tx` refactor (cost
  unestimated); DBOS tables never enter writer matrix; `math/big` JSON codec
  care; Rewind/Delete governed off approved artifacts.

## Part V — immudb Assessment

Full: [RESEARCH.md §4](../docs/stack/RESEARCH.md#4-immudb-findings-phase-4).

- Evidence model: KV `prov/…` canonical JSON (hashes, rule versions,
  attestations, close records, receipts); idempotency keys + preconditions;
  PG back-links set only after `VerifiedTxByID`.
- Cryptographic verification: inclusion + dual proofs (`ErrCorruptedData` on
  tamper DOCUMENTED); ECDSA server signatures; trusted state OUTSIDE writer;
  independent auditor + webhook; genesis pinning UNRESOLVED.
- Native Go integration: `pkg/client` over gRPC to standalone server;
  `ExecAll` atomic; `noWait=false` on append path; wire protocol REJECTED.
- Privacy: hashes/digests/metadata ONLY — Delete/expirable logical-only,
  retention truncates value log only (DOCUMENTED); PII/secrets/entities/churn
  stay out (non-use-cases enumerated).
- Retry/reconciliation: outbox + deterministic IDs + precondition dedup +
  `TxScan` recovery; `ledger_tx` canonical sequence (INFERRED sketch).
- Operational deployment: standalone server + auditor (recommended over
  embedded: trust model, no-flock corruption risk, BSL optics); hot backup,
  replicas, `synced=true`; license BSL-1.1 v1.11.1 (2026-06-26), 4-year
  per-version Apache conversion; re-pin at bring-up (plan cites v1.11.2).

## Part VI — Integrated Architecture

- Component responsibilities: [PATTERNS.md ownership table](../docs/stack/PATTERNS.md#3-ownership-table).
- Architecture diagram (textual): CLI/API → `internal/` services → PG 18.6
  (truth) in DBOS `RunAsTransaction` (execution) → outbox → DBOS worker →
  immudb (evidence) → back-link → auditor verifies (ASSUMPTION sketch;
  §6A-annotated data flow: [PATTERNS.md §4](../docs/stack/PATTERNS.md#4-data-flow-description-candidate-6a-sequence-annotated)).
- Sequence: request → workflow resume → PG validate → PG commit + outbox →
  durable evidence event → ledger append → back-link → reconcile (§4 steps 1–9).
- Transaction boundaries: gate+read+write+checkpoint = ONE PG txn (shared
  pool); outbox→ledger→back-link = three coordinated steps with idempotent
  recovery (no spanning atomicity — binding risk).
- Failure/recovery matrix: [PATTERNS.md §5](../docs/stack/PATTERNS.md#5-failure-matrix-component--step).
- Dynamic rule-management: versioned SYNTH packs as effective-dated PG data +
  DB coverage gate (PERIOD FK) + immudb pack-digest attestation; no dedicated
  rules engine for POC (simpler pack architecture sufficient — ASSUMPTION per
  prompt.md §6C evaluation rule).

Defect → pattern → technology → enforcement → test mapping (ASSUMPTION table
on DOCUMENTED defects; patterns C/K per PATTERNS.md):

| Defect (DOCUMENTED) | Pattern | Technology | Enforcement | Test |
|---|---|---|---|---|
| D1 stale-RR admission | C5, K6 | PG txn in DBOS RunAsTransaction | RC + txn-scoped employer gate + proof-check | Battery T1 both orderings |
| P1a–g barrier gaps | C10 | PG18 | Revokes + triggers + DEFINER; PERIOD/FK where declarative | Battery T2–T8 as app |
| EF-E3 fabrication | C8, K3 | PG + immudb | Revoke + DEFINER-only; external dual-proof anchor | T9 + ledger re-verify |
| EF-F1A/B direct/vacuous | K3, K4 | PG18 | State-machine + ≥1-row + coverage triggers | T10–T11 |
| Crash-ambiguous capture | C5–C7 | DBOS + PG outbox | Atomic txn + checkpointed workflow | Kill -9 experiments 1–8 |
| Missing/misordered ledger | C11, K7 | immudb + outbox | Preconditions + idempotent append + TxScan | Crash matrix 2–4 |
| Silent tamper | C8, K8 | immudb | Dual proofs + auditor + webhook | Tamper experiments |
| Coverage gaps/overhangs | C3–C4, K5/K9 | PG18 temporal + triggers | PERIOD FK + live-basis triggers | E2–E3 demos; Blocker fixtures |
| Non-reproducible runs | K10, D8 | PG + immudb | Canon bytes + pinned versions + attestations | Re-derivation equality |
| Mid-period ambiguity | K9, D5 | PG18 | Typed blockers (proration absent) | Mid-period fixtures |

## Part VII — One-Month Payroll POC

Scope/assumptions: synthetic ONLY (`SYNTH-*`, `synthetic: true`; OT-001 stays
BLOCKED with a test proving synthetic rules can't activate as statutory);
25–50 employees, 2 pay groups; one legal employer (D7 PROPOSED); disposable
DBs; `poc/payroll-month/` on `stack-poc`, uncommitted; three time zones.
Business scenarios: mid-month hire/leaver; mid-period comp change (→blocker);
backdated change → supersession chain; SHG/coverage union incl. gap +
multi-fund; SYNTH-OT + SYNTH-leave; pack vN→vN+1 mid-month with per-output
citation + re-derivability; late correction pre-capture + post-close
adjustment; period close → immudb + post-close rejection; adversarial writers
during capture; kill -9 at (a)/(b); auditor re-derivation; on-disk tamper;
rerun byte-identical (binding Phase 6 list). Data model: POC-scoped DDL only
([ADR-0202](../docs/stack/adr/ADR-0202.md)). Lifecycle: period create →
snapshot → ingest → reconcile → validate → SYNTH calc → exceptions → review →
approve → finalize → close → evidence → register/payslips/exceptions/summaries/
payment-preview/close-report/provenance-report. Workflow design: capture +
month-close + reconciliation + retro candidates (Part IV). Rule-pack design:
effective-dated SYNTH packs, DB coverage gate, digest attestation.
Acceptance tests: F1–F7 tables, kill/tamper logs, attack mapping, deletion
ledger, limitations → `poc/POC-REPORT.md` (Phase 6 output, not this gate).

## Part VIII — Infrastructure and Migration

- Compose design (Phase 6): digest-pinned `postgres:18.6` + pinned immudb +
  app/DBOS; one-command start, health checks, readiness ordering, teardown;
  "clone to green" runbook (ASSUMPTION target state).
- PG 16→18.6: fresh cluster + `init-roles.sh` + `migrate up` + `seed`; volume
  `/var/lib/postgresql` (DOCUMENTED); 16 volume preserved; dev-data path +
  parity check in Phase 7 ([report.md](report.md) Phase 7).
- DBOS/immudb startup: privileged role runs DBOS DDL once, app `SkipMigrations`
  (DOCUMENTED grants); immudb mTLS/auth + auditor scheduling (INFERRED).
- Env config: env-first, secrets via env only (AGENTS.md convention).
- Local dev: single start command; no manual three-system wrangling (target).
- CI: matrix on PG18 + kill-test job + tamper-test job + translated battery
  required (binding Phase 7).

## Part IX — Implementation Roadmap

Ordered, independently verifiable; each milestone: objective, files, deps,
tasks, DB changes, tests, verification, risks. (ASSUMPTION plan; gates binding.)

- M1 Bring-up: compose (18.6 pinned + immudb pinned + app), health/ordering,
  teardown. Files: `poc/payroll-month/*`, compose. Tests: E5 (migrations +
  suite on 18.6). Risk: assertion drift.
- M2 Temporal POC DDL: range columns, WITHOUT OVERLAPS, PERIOD FKs, close
  state, outbox, back-links (POC-scoped). Tests: E1–E4, coverage demos.
  Risk: E1 negative (generated-column rejection) → STORED fallback.
- M3 DBOS wrap: version pin, `go mod` build, capture workflow + gate txn,
  `Tx` refactor. Tests: crash experiments 1–8. Risk: SDK drift, refactor size.
- M4 Provenance: KV schema, appender, back-links, auditor, genesis procedure.
  Tests: round-trip, crash 2–4, tamper, PII scan, re-derivation. Risk: license
  R1 negative → void, fallback PG+DBOS.
- M5 Translated battery: T1–T12 green both orderings + writer-first suite.
  Risk: mechanism red reopens ADR-0203.
- M6 Scenario matrix + POC-REPORT: full §VII scenarios, three TZs, F1–F7.
  Gate: `task build && task test && task lint` green + DB-gated tests run.

## Part X — Adversarial Findings and Decisions

- Weaknesses discovered: D1/P1/EF-F classes (DOCUMENTED); dual-write has no
  native atomicity (highest risk); DBOS replay/Rewind governance gaps;
  immudb genesis + custody gaps; PG18 E1–E6 unexecuted; DBOS pin + MVS
  unproven; `math/big`-over-JSON hazard; new distributed failure modes
  (ledger-down, executor fencing races, outbox backlog).
- Unresolved questions: carried gaps — Phase 0 (8 sign-offs, B1/B2/X-FUND,
  proration, G-PREREQ, sync scope), Phase 1 U1–U7, Phase 2 E1–E6 + 5,
  Phase 3 (1–7), Phase 4 (1–7); see RESEARCH.md §8 + DECISION-REQUESTS.md R1–R7.
- Rejected alternatives: status-quo PG16; PG-alone/immudb-wire/embedded-ledger/
  ordinary-step-capture/XID-resurrect/early-00013 (ADR-0201/0203/0204 alts).
- 4-option trade-off (ASSUMPTION synthesis on researcher evidence):

| Option | Integrity gain | Durability gain | Evidence gain | Cost/complexity | Verdict |
|---|---|---|---|---|---|
| PG alone | ◐ temporal keys only | ○ | ○ | Lowest | REJECT for POC (gaps 1–2 survive) |
| PG+DBOS | ◐ | ● crash-safe + orchestration | ○ | +SDK + placement ruling | Fallback if R1 negative |
| PG+immudb | ◐ | ○ | ● tamper-evidence | +ops + BSL + custody | Fallback if DBOS unfit |
| All three | ● | ● | ● | Highest, bounded by ownership table | PROPOSED for POC (conditional) |

- Technical risks: RESEARCH.md §7 (8 risks with owners).
- Go/no-go: NO-GO on Phase 6 until HUMAN answers R1–R7 and approves the
  research gate. The gate is NOT passed by this report. On approval with
  conditions, conditions become acceptance criteria and M1–M6 proceed in order.

