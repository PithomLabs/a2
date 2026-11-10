# RESEARCH — Phases 1–4 synthesis (research gate)

Status: RESEARCH GATE — evidence synthesis, not approval. All POC build awaits
human sign-off per the binding plan ([report.md](/home/chaschel/Documents/go/a2/plans2/report.md:29-31):
synthetic-only POC, research gate before POC build).
Companions: [DECISIONS.md](DECISIONS.md) (business decisions),
[PATTERNS.md](PATTERNS.md) (pattern catalogue), `adr/` (technical ADRs),
[DECISION-REQUESTS.md](DECISION-REQUESTS.md) (human action list).

Claim tags (preserved from researchers): VERIFIED = directly observed this
programme; DOCUMENTED = cited source opened; INFERRED = reasoning only, needs
experiment; ASSUMPTION = synthesis beyond researcher evidence (mine).

Researcher evidence base: all five phases were READ-ONLY. No phase ran the app,
migrated a database, or executed the test suite. Everything below that depends
on execution is therefore DOCUMENTED or INFERRED at best — see Limitations §8.

## 1. Current-state map (Phase 1)

Module `github.com/PithomLabs/a2` (DOCUMENTED `go.mod:1`), but Go imports use
`github.com/PithomLabs/a1/...` (DOCUMENTED `internal/payroll/snapshot.go:18-24`)
— UNRESOLVED U1, build layout unverified. Single binary `cmd/workforceops`
(DOCUMENTED `Taskfile.yml:15-19`). CLI/API share `internal/` services
(DOCUMENTED `internal/httpapi/httpapi.go:1-27`); API is loopback
`127.0.0.1:8081`, no auth v0.1 (DOCUMENTED `openapi.yaml:1-15`).

Key packages (DOCUMENTED imports): `core/database` (pgx pool, goose migrate,
advisory-lock helpers, snapshot-tx); `internal/compbasis` (single
effective-comp resolver, `BasisConflict` on ambiguity); `internal/payroll`
(runs, readiness, snapshot capture, validate, drift); `internal/rules`
(manifest load, evidence verify); `rules/SG-MOM-OT-001` L1 spec package,
`status:BLOCKED`, no Go impl.

Migrations 00001–00012 (VERIFIED 5484 lines total): foundation → attendance/
payroll → rules/audit → protection (revokes, freeze triggers, GiST no-overlap)
→ uuid defaults → trigger fixes → renames → break kinds → M1.1 insert-only
enforcement → M1.2 dispositions/corrections/HR functions → M1.3 barrier locks,
comp lifecycle → M2 SHG lifecycle, pack governance, snapshot, run lifecycle.
Roles: `migration_admin` owns schema/DDL; `app` runtime;
`REVOKE TEMPORARY`, `REVOKE CREATE ON SCHEMA public FROM PUBLIC`
(DOCUMENTED `scripts/db/init-roles.sh`). Boot check `RefusePrivilegedRuntime`
(DOCUMENTED `core/database/db.go:18-32`).

Capture path as implemented (DOCUMENTED `snapshot.go`, `validate.go`,
`core/database/lock.go`): `BeginSnapshotTx` REPEATABLE READ → run `FOR UPDATE`
→ population lock → expected population → per-employee input locks (UUID-sorted)
→ pack locks → vacuous same-tx revalidation (D5-P2b) → resolve → canonical
SHA-256 → INSERT header+children+pins → audit. Validate: run lock →
tipDrifted (zero locks) → invalidate+recapture or version-match →
`snapshotBlockers` → stamp `inputs_hash`. Lock order population → employee →
pack. `core/database/admission.go` XID-ticket mechanism is experimental, NOT
called by capture/validate; Gate 2 ADR supersedes it with Option B (RC +
tx-scoped employer gate + claim proof-check), also NOT implemented.

Known defect classes (DOCUMENTED M2 review §6–8): D1 stale-RR admission (P0);
P1a approval lacks population lock; P1b terms-close lacks input lock; P1c
disposition direct-SQL lacks input lock; P1d `status_history` direct INSERT no
trigger; P1e `ensure_pack_blocked` no lock; P1f draft-line direct DML no lock;
P1g period/employer UPDATE unlocked; EF-E3 snapshot-table direct INSERT
fabrication (P0); EF-F1A/B direct run→validated / vacuous validate (P1).

## 2. PostgreSQL 18.6 findings (Phase 2)

External sources accessed 2026-11-10. Local baseline VERIFIED:
`postgres:16-alpine`, volume `a1-pgdata:/var/lib/postgresql/data`; extensions
only `pgcrypto` + `btree_gist`; all effective ranges are two DATE columns, no
range-typed column exists; all DEFINER functions `SET search_path = pg_catalog`,
owner `migration_admin`; pgx v5.12.0, goose v3.27.3.

### 2.1 WITHOUT OVERLAPS temporal keys — HELPS (narrowly)

DOCUMENTED ([PG18 CREATE TABLE](https://www.postgresql.org/docs/18/sql-createtable.html)):
`UNIQUE/PK (cols..., range_col WITHOUT OVERLAPS)` behaves like
`EXCLUDE USING GIST (cols WITH =, range WITH &&)`; last column must be
range/multirange; scalar columns need `btree_gist`; GiST-enforced.
Replaces plain no-overlap exclusions (employment_terms, pay_period,
leave_request, team_membership, employee_skill, rotation_assignment, status
history) with declarative keys — same enforcement, better metadata
(INFERRED). CANNOT replace live-basis successor logic: comp
(`check_compensation_line`, `00011:622`) and SHG (`check_shg_instruction`,
`00012:81`) deliberately dropped exclusions because superseded history must not
collide — "an exclusion cannot express liveness" (VERIFIED `00011:606-613`).
Adoption needs a real range column; WITHOUT OVERLAPS on a generated column is
UNCONFIRMED — experiment E1 required.

### 2.2 PERIOD foreign keys — HELPS for containment/coverage

DOCUMENTED (same page): a PERIOD FK "is considered satisfied if the referenced
table has matching records ... whose combined PERIOD values completely cover
the referencing record's" — several adjacent referenced rows jointly cover.
Limits: NO `RESTRICT`/`CASCADE`/`SET NULL`/`SET DEFAULT` (only `NO ACTION`);
referenced side must carry a WITHOUT OVERLAPS PK/UNIQUE.
Comp containment (line ⊆ parent terms) expressible via degenerate key
(INFERRED, sketch in researcher result). SHG coverage: NEUTRAL — successor
chains + pending/opt_out/verified_none states don't fit a gap-free referenced
key; `check_shg_instruction` stays (INFERRED). Rule-pack full-period coverage:
helps for single-pack cases; needs param-table restructure + one coverage row
per pack; multi-pack join stays app-side (INFERRED).

### 2.3 Neutral / N/A

- `uuidv7()` core, ms timestamp (DOCUMENTED
  [PG18 §9.14](https://www.postgresql.org/docs/18/functions-uuid.html)) — NEUTRAL:
  low-volume HR workload; do not migrate existing PKs; optional for new
  append-only tables (INFERRED).
- Virtual generated columns (DOCUMENTED
  [PG18 §5.4](https://www.postgresql.org/docs/18/ddl-generated-columns.html))
  — NEUTRAL, no adoption driver (INFERRED).
- `RETURNING OLD/NEW` (DOCUMENTED
  [PG18 §6.4](https://www.postgresql.org/docs/18/dml-returning.html)) — NEUTRAL
  nice-to-have; no current defect closed (INFERRED).
- Async I/O, skip scan, checksums, MAINTAIN — N/A, transparent (DOCUMENTED,
  INFERRED).

### 2.4 16 → 18.6 upgrade

Exact versions DOCUMENTED: PG 18.6 released 2026-08-13 (18.5 skipped);
16.15 co-released; 16 EOL 2028-11-09; 18 EOL 2030-11-14
([versioning](https://www.postgresql.org/support/versioning/)).
`pg_upgrade` 16→18 direct supported (DOCUMENTED
[pg_upgrade](https://www.postgresql.org/docs/18/pgupgrade.html)), but
INFERRED recommendation: fresh 18.6 cluster + migrate + reload (no production
data; recycle-by-drop is the norm). Volume/layout change DOCUMENTED (official
image docs): PGDATA version-specific `/var/lib/postgresql/18/docker`, VOLUME
now `/var/lib/postgresql`; old 16 volume NOT reusable in place (INFERRED).
Rollback: keep 16 compose file + `down` (not `-v`) (INFERRED).
Extensions `btree_gist`/`pgcrypto` ship on 18, no removal; no reindex required
for this schema (uuid/date/daterange/tstzrange/text), 18.6 float-NaN/bit-only
reindex advisory (DOCUMENTED release notes). Behavior changes: checksums on by
default (adopt free, INFERRED); AFTER triggers run as queue-time role (neutral
here, INFERRED — one AFTER constraint trigger, no mid-txn role change); no
DEFINER/advisory/isolation changes in 17 or 18 migration lists (VERIFIED by
full-text search of both release pages). Migrations 00001–00012 expected to
apply cleanly (INFERRED — must be proven by E5, NOT executed).

Experiments E1–E6 specified, NONE executed (read-only phase, no server).

### 2.5 PG18 verdict: ADOPT-FOR-POC (conditional)

Adopt 18.6 as the POC baseline (fresh cluster + migrate + reload) conditional
on E5 green; adopt WITHOUT OVERLAPS/PERIOD FK selectively for containment and
single-pack coverage conditional on E1–E3; defer uuidv7/virtual/OLD-NEW
(ASSUMPTION: synthesis judgment on researcher evidence).

## 3. DBOS Transact Go findings (Phase 3)

Checkout: `go 1.25.0` floor, `toolchain go1.26.10` (DOCUMENTED
`dbos-transact-golang/go.mod:3-5`); pgx v5.10.0 vs a2 v5.12.0 — skew INFERRED
low risk (stable `pgxpool`/`pgx.Tx` surface, MVS unifies); Go skew INFERRED low
risk (a2 toolchain 1.27.3 VERIFIED via `go version`). NO consumable version pin
(no git repo, no VERSION file — VERIFIED). Supports Postgres/CockroachDB/SQLite
(DOCUMENTED `dbos/dbos.go:44-50`).

System DB (`dbos` schema default, configurable): `workflow_status`,
`operation_outputs(workflow_uuid, function_id)` — every step checkpointed,
step identity `(workflow_id, integer step_id)` in call order — plus
notifications, events, streams, payload, queue, schedule, version tables
(DOCUMENTED migrations + `system_database.go`). Placement same-DB vs
separate-DB is purely which pool/URL you hand it (INFERRED); same-DB enables
the atomic path. Privileges: privileged role runs DDL once; app role
`SkipMigrations` with `USAGE` + `ALL` on system tables (DOCUMENTED
`dbos/migrations.go:93-108`).

Atomic business write + checkpoint: YES, but ONLY on shared-pool path —
`runAsTxn` commits app writes + `RecordOperationResult` in one txn (DOCUMENTED
`workflow.go:2908-2909,2941-2978`). Separate user DB = two txns; crash window
closed by REPLAY not atomicity; `fn` must be deterministic — a losing owner's
committed write can be adopted by the winner (DOCUMENTED
`datasource.go:538-540`). Ordinary `RunAsStep` runs OUTSIDE any DB txn;
checkpoint is a separate write after success (DOCUMENTED `workflow.go:2853-2897`).

Retry/replay: `RunAsStep` = at-least-once, DOCUMENTED verbatim
(`workflow.go:2741`); default maxRetries 0. `RunAsTransaction` =
effectively-once via `(workflow_id, step_id)` PK (DOCUMENTED). External side
effects: at-least-once only — use idempotency keys at callee (INFERRED).
Cancel mid-step: step NOT checkpointed, resume re-executes (DOCUMENTED).
Workflow fn must be deterministic; enforcement structural (step-ID sequence +
function-name matching → `UnexpectedStep`/`StepNondeterminism`), not sandbox
(DOCUMENTED). In-flight code changes break replay unless `Patch` markers or
Enterprise AutoUpgrade (DOCUMENTED `dbos.go:739-746`).
`RewindWorkflow` deletes history, re-runs same ID, replays `Send` as second
copy (DOCUMENTED) — NOT suitable for approved artifacts (INFERRED, AGENTS.md §7).

Locks across recovery: NO business locks survive — only fencing is the
`owner_xid` token with `SELECT ... FOR NO KEY UPDATE` inside checkpoint txns
(DOCUMENTED); advisory locks appear only for migrations. Gate 2's RC +
txn-scoped employer gate must live INSIDE one `RunAsTransaction` txn; DBOS
cannot hold it across steps or crashes (DOCUMENTED + INFERRED).

Candidates: capture pipeline (`capture:{runID}` workflow, readiness step →
atomic capture txn → hash-verify → `SetEvent`); month-run open-to-close (cron
per employer, child workflows, `Recv` approval gate, durable `Sleep`);
PG→immudb reconciliation (at-least-once anchor ⇒ sink idempotent on
`(snapshot_id, hash)`); retro adjustments as child workflows (supersede-only;
no `Rewind`). 8 crash-injection experiments specified, NONE executed.
DBOS does NOT replace: Gate-2 DB barriers, isolation choice, money/statutory
purity (JSON `any` serialization needs `math/big` codec care — INFERRED),
immutability governance, exactly-once externals.

### 3.1 DBOS verdict: ADOPT-FOR-POC (conditional, at-least-once-until-proven)

Adopt for capture pipeline + month-close orchestration on the shared-pool
(same-DB `dbos` schema) path, conditional on: consumable version pin, scratch
`go mod` + build proving pgx MVS interop, crash-injection battery green, and
human ruling on system-DB placement vs "one database" (binding plan conflict #5:
`prompt.md` at-least-once wins until disproven). Defer AutoUpgrade/Enterprise
(ASSUMPTION: synthesis judgment).

## 4. immudb findings (Phase 4)

### 4.1 License — SOURCE-AVAILABLE, NOT OPEN-SOURCE

Exact license DOCUMENTED (`immudb/LICENSE:1-48`): Business Source License 1.1,
licensor Codenotary, work "immudb Version 1.9DOM.2 or later". Use grant:
copy/modify/redistribute + NON-PRODUCTION use; production allowed UNLESS
offering the work to third parties hosted/embedded to compete with Codenotary
paid versions. Competitive = paid-basis offering significantly overlapping
Codenotary paid capabilities; internal org use NOT competitive; each version
converts to Apache 2.0 FOUR YEARS after first public distribution
(per-version). Version under review DOCUMENTED: v1.11.1 (2026-06-26),
`CHANGELOG.md:7-8`; Go floor 1.25.0, compatible direction with a2 1.27.3
(INFERRED). a2 is MIT (DOCUMENTED `a2/LICENSE:1`).

Constraints (INFERRED from grant text): license-class mismatch with
open-source backend — non-redistributable-as-pure-open-source until each pinned
version's 4-year conversion; self-hosted/internal use permitted incl.
production; paid-hosted multi-tenant future could trigger the
hosted-or-embedded-to-compete clause — needs new decision + likely commercial
license before any paid-hosted launch. Mitigation: narrow replaceable interface
(standalone server), pin exact version + Change Date. NOTE: binding plan cites
"v1.11.2 latest" (report.md:22) vs researcher-pinned v1.11.1 — re-pin at
bring-up (DOCUMENTED plan instruction).

### 4.2 Deployment: standalone server, NOT embedded (recommended)

Trade-off table DOCUMENTED: standalone keeps verifier state + auditor outside
the writing app, server serializes writes, cleaner BSL boundary, hot backup,
replica APIs — at the cost of an extra container + mTLS/auth + auditor
scheduling. Embedded: same process holds data+writer+verifier (defeats
tamper-evidence if compromised — INFERRED); NO lock file, two processes on one
dir = silent corruption (DOCUMENTED `docs/EMBEDDING.md:102-122`); "Embedded"
is a defined license term (worse optics — INFERRED). Single-gated write path:
all appends through one service function → one immudb user/database; KV txns
atomic all-or-nothing serialized (DOCUMENTED `transactions.md:6-9`); keep
`synced=true` default (DOCUMENTED).

### 4.3 Provenance schema + verification (KV preferred)

Key `prov/{employer_id}/{artifact}/{artifact_uuid}` → canonical JSON
(artifact, pg_table/pg_id, sha256 of canonical bytes, rule_versions,
decided_by/at, supersedes, immu_tx); idempotency key `idem/...` in same
`ExecAll` with `PreconditionKeyMustNotExist` (constructors DOCUMENTED
`preconditions.go:19-48`). Back-links: PG rows gain nullable `ledger_tx`,
`ledger_alh`, `ledger_verified_at`, set only after verified append (INFERRED
design; supersede = new immu tx + new PG row, never edit). Trusted state
(latest `(TxId, TxHash[, Signature])` per serverUUID/db) MUST live OUTSIDE the
writing app's mutable reach — separate auditor host/volume — else a compromised
writer rewrites history + checkpoint together (INFERRED from verified-op
semantics DOCUMENTED `client.go:1119-1138`). Verified ops: inclusion proof +
dual proof via `store.VerifyInclusion`/`VerifyDualProof`, fail
`store.ErrCorruptedData`; optional ECDSA server signature; independent auditor
(`immuclient audit-mode`, `--audit-verify-values`, tamper webhook). First-ever
state bootstraps FROM THE SERVER (`state_service.go:113-122`) — pin genesis
out-of-band (INFERRED gap, UNRESOLVED).

### 4.4 Dual-write protocol: PG → ledger → back-link

Outbox sketch: (1) PG txn INSERT business rows (ledger_* NULL) + outbox row,
COMMIT; (2) single-gated appender `ExecAll` with idempotency precondition,
`noWait=false`; (3) `VerifiedTxByID` → alh; (4) PG txn fills back-link + marks
outbox done. Crash 2–4: re-`ExecAll`, precondition failure = already appended,
recover `immu_tx` via `Get`/`TxScan`, then step 4 (INFERRED). NEVER `noWait`
on append path (DOCUMENTED freshness semantics). `ledger_tx` is canonical
sequence, not PG commit order (INFERRED).

### 4.5 Non-use-cases

PII/erasure-subject data OUT: `Delete`/`ExpirableSet` logical only, values
remain; retention truncates value log only (DOCUMENTED `deleting.md:6`,
`retention.md:14-20`). NOT primary system of record; NOT secrets/NRICs/bank
details; NOT high-churn operational state. NOT via PG-wire protocol: reports
PG 14.0, SILENTLY SKIPS unsupported DDL (`CREATE FUNCTION/TRIGGER/GRANT…`,
DOCUMENTED `pg.md:36`) — correctness hazard (INFERRED). What belongs: content
hashes/digests, rule-version stamps, decision/approval attestations, run-close
records, KET receipts, outbox checkpoints, auditor root hashes (INFERRED).

### 4.6 immudb verdict: ADOPT-FOR-POC (conditional, native SDK, hashes-only)

Adopt as provenance sidecar via native Go SDK, standalone server, hashes/digests
only, conditional on: human license decision, trusted-state custody decision,
genesis pinning procedure, and crash/tamper battery green. Wire-protocol
interface REJECTED for POC use (ASSUMPTION: synthesis judgment; binding plan
conflict #6 already mandates native-SDK-primary).

## 5. Materiality matrix (features × robust / flexible / extensible)

Objective A criteria: robust = integrity/correctness under adversity;
flexible = configurable without code change; extensible = new capabilities
without rewrite. Ratings are synthesis judgments (ASSUMPTION) grounded in the
cited researcher evidence.

| Feature / capability | Robust | Flexible | Extensible | Evidence anchor |
|---|---|---|---|---|
| WITHOUT OVERLAPS plain no-overlap keys | ● helps | ○ | ○ | Phase 2 §1.1 |
| PERIOD FK containment (comp ⊆ terms) | ● helps | ○ | ○ | Phase 2 §1.2 |
| PERIOD FK multi-row joint coverage (pack validity) | ● helps | ◐ single-pack | ○ | Phase 2 §1.2 |
| Live-basis successor logic (comp/SHG) in PG triggers | ● stays | ○ | ○ | Phase 2 "exclusion cannot express liveness" |
| Fresh 18.6 cluster + checksums default | ● ops | ○ | ○ | Phase 2 §2 |
| DBOS atomic capture (shared-pool RunAsTransaction) | ● crash-safe | ○ | ○ | Phase 3 §2 |
| DBOS month-close orchestration (cron/Recv/Sleep/queues) | ◐ | ● durable waits | ● new phases as children | Phase 3 §3 |
| DBOS replay protection (step IDs + name match) | ● divergence loud | ○ | ◐ Patch-gated evolution | Phase 3 §2 |
| immudb dual-proof provenance (hashes-only) | ● tamper-evident | ○ | ◐ new artifact types = new keys | Phase 4 §3–4 |
| immudb outbox + idempotent append | ● no dup/miss | ○ | ○ | Phase 4 §4 |
| Versioned rule packs + coverage gate | ● | ● effective-dated data | ● new SYNTH types | Phase 1 pack gov + Phase 2 coverage |
| Existing advisory-lock discipline as-is | ○ 7 P1 gaps | ○ | ○ | Phase 1 writer matrix |

Legend: ● material gain · ◐ partial · ○ no gain claimed.

## 6. Deletion ledger (M2.1 mechanisms kept/deleted — PROPOSED for POC)

Per binding success criterion 5, every legacy mechanism needs a keep/delete
ruling with rationale. These are PROPOSALS (ASSUMPTION); the POC build proves
or overturns each via the translated battery ([ADR-0203](adr/ADR-0203.md)).

| # | Mechanism | Disposition | Rationale |
|---|---|---|---|
| L1 | `BeginSnapshotTx` REPEATABLE READ capture | DELETE (replace with Gate-2 RC + txn-scoped employer gate inside DBOS `RunAsTransaction`) | D1 stale-RR admission live-proven (DOCUMENTED M2 §6); DBOS cannot hold cross-step locks so gate must be single-txn (Phase 3 §2) |
| L2 | `admission.go` XID-ticket mechanism | DELETE (already unwired; Gate 2 ADR retired Option A) | DOCUMENTED `admission.go:1-10` not called; Gate 2 ADR supersedes |
| L3 | Same-tx revalidation in `captureInTx` | DELETE (provably vacuous per D5-P2b) | DOCUMENTED review D5-P2b |
| L4 | Population/employee/pack advisory-lock helpers | KEEP (narrowed: inside capture txn; audit P1a–g closures) | Lock order + key derivation sound; discipline gaps are coverage, not concept (Phase 1 §4) |
| L5 | `tipDrifted` lock-free drift check | DELETE (replace with claim proof-check inside gate txn) | Takes zero locks; Gate 2 replaces with proof-check (Phase 1 §4, Gate 2 ADR) |
| L6 | Manual stale-reoffer protocol (capture→validate) | DELETE (replace with DBOS checkpointed workflow) | Two ad-hoc txs + ambiguous crash state; no idempotent step log (Phase 1 gap 1) |
| L7 | Plain-table `audit_event` as tamper evidence | KEEP as operational audit, DELETE as tamper evidence (replaced by immudb dual-proof) | Self-asserted content; nothing external anchors hash (Phase 1 gap 2, Phase 4) |
| L8 | Live-basis triggers (comp/SHG) | KEEP | No PG18 declarative equivalent (Phase 2 §1.1–1.2) |
| L9 | GiST exclusions on plain no-overlap tables | KEEP for POC; MIGRATE to WITHOUT OVERLAPS post-E1 selectively | Same enforcement; migration is metadata/optimizer hygiene, not a defect fix (Phase 2 §1.1) |
| L10 | Capture-first-only race tests | DELETE as sufficient suite (add writer-first interleavings) | DOCUMENTED hole D6-P2c |
| L11 | DEFINER HR/comp/SHG/pack functions | KEEP | Search-path hardened, ownership model sound (Phase 2 §0) |

## 7. Risks (top, with owner)

1. Dual-write PG↔immudb has no native atomicity — outbox + deterministic IDs +
   reconciliation is the highest-risk validation step (binding plan risk;
   Phase 4 sketch INFERRED, unexecuted). Owner: POC build (Phase 6).
2. DBOS exactly-once wording vs at-least-once reality for externals — binding
   resolution: at-least-once until proven by crash injection. Owner: POC build.
3. immudb BSL-1.1 license-class mismatch with open-source backend; paid-hosted
   future needs counsel + possibly commercial license. Owner: HUMAN (decision
   request).
4. Trusted-state custody: verifier state inside the writer defeats
   tamper-evidence; custody + genesis pinning undecided. Owner: HUMAN.
5. Pending Gate 1 decision (XID-ticket vs transaction-scoped) constrains ADR-0203;
   pre-empting it is forbidden — POC uses Gate 2 direction on disposable DBs
   only, production migration NOT authorized. Owner: HUMAN.
6. PG18 behavior drift: E5 (migrations + suite green on 18.6) not run; pinned
   advisory counts / `pg_locks` assertions may drift. Owner: POC bring-up.
7. DBOS Go SDK maturity/API drift; no consumable version pin; pgx MVS interop
   unbuilt; `math/big` JSON codec care; Rewind/Delete governance gap. Owner:
   POC bring-up.
8. Threat-model change proposals (rung changes, `migration_admin` powers,
   truncate rights) must be decision requests, never silent. Owner: HUMAN.

## 8. Limitations (honest)

- All five research phases were READ-ONLY: no DB was migrated, no test
  executed, no container started. Every "expected green / applies cleanly" is
  INFERRED until the POC bring-up runs it.
- PG18 experiments E1–E6, DBOS crash experiments 1–8, and the translated
  12-attack battery are SPECIFICATIONS, not results.
- Phase 1 gaps U1–U7 carried (go.mod a2-vs-a1 layout, trigger-body line reads,
  quarantine fixture body, module runtime role, OpenAPI parity, PG18 deltas,
  red-attack current status).
- `prompt.md` statutory scenarios have synthetic analogues only; items with no
  synthetic analogue are listed as limitations in the POC plan, not guessed
  (binding Key Decision #1).
- This synthesis adds no new experiments; anything beyond researcher evidence
  is marked ASSUMPTION.
