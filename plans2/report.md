# Unified implementation plan for `plans2/prompt*.md` (PostgreSQL 18.6 + DBOS + immudb)

## Goal

Execute all five `plans2/prompt*.md` briefs as one coherent programme: deep architectural research on the WorkforceOps codebase, a pattern catalogue with ADRs, and a one-month synthetic-payroll POC on PostgreSQL 18.6 + DBOS Transact (Go) + immudb — without violating frozen migrations, the OT-001 BLOCKED constraint, or the Gate 1 pending decision.

Inputs read in full: `prompt.md`, `prompt_claude.md`, `prompt_deepseek.md`, `prompt_z.md`, `prompt_z2.md`, plus all of `checklist/` and `retrospect/` (17 files). Note: `prompt_z2.md` is byte-identical in substance to `checklist/checklist_Z.md` (the Design-Gate Checklist), so "implementing" it means applying it as the design gate, not building software.

## Success Criteria

1. Research report + pattern catalogue + ADRs exist and every material claim is tagged VERIFIED (experiment run), DOCUMENTED (versioned source), or INFERRED — no memory claims.
2. One-month synthetic payroll POC runs end-to-end from one command on PG 18.6 + DBOS + immudb, with crash-injection, tamper-detection, concurrency, coverage, idempotency, and reproducibility experiments automated and green across three time zones.
3. Migrations `00001`–`00012` untouched; POC lives in `poc/payroll-month/` on branch `stack-poc`, uncommitted until gate review; disposable databases only.
4. Zero statutory calculations: every POC rule is `synthetic: true`, and a test proves no synthetic rule can activate as statutory (OT-001 stays BLOCKED).
5. Each legacy Gate-1 attack has a translated equivalent test on the new stack, or a proof the class is closed by construction; a deletion ledger records every M2.1 mechanism kept/deleted with quantified rationale.
6. `task build && task test && task lint` green; DB-gated tests run with `WORKFORCEOPS_TEST_DATABASE_URL`.

## Approach

Single phased programme, research-first with a hard gate before any POC build (satisfies `prompt.md`'s stop-after-research rule and the other briefs' build mandates in sequence). Business decisions precede technical mechanism, per `checklist/decisions.md` ("Business Closure Before Technical Synchronization") and the AGENTS.md invariant-first principle.

Workspace facts grounding this plan (verified locally, 2026-11-10): a2 is `github.com/PithomLabs/a2`, Go 1.27.3, pgx v5.12, goose; `docker-compose.yml` pins `postgres:16-alpine`; migrations `00001`–`00012` live in `migrations/core/`; DBOS checkout (`dbos-transact-golang`) targets Go 1.25.0 with pgx v5.10 and a `dbos` system schema; immudb checkout CHANGELOG tops at v1.11.1 (2026-06-26) plus Unreleased — so `prompt.md`'s "v1.11.2 latest" must be re-pinned at bring-up, and the pgx v5.12/v5.10 skew must be compatibility-checked.

## Key Decisions (conflicts between the five briefs, resolved)

| # | Conflict | Resolution | Rationale |
|---|---|---|---|
| 1 | `prompt.md` wants real statutory rules; `prompt_claude` forbids them (synthetic only) | Synthetic-only POC (`SYNTH-*`, `synthetic: true`) | AGENTS.md statutory workflow + OT-001 BLOCKED + M2.1 constraints are inherited and non-negotiable; `prompt.md`'s statutory scenarios become synthetic analogues |
| 2 | `prompt.md` says stop after research; others mandate a POC build | Phased gates: research gate first, POC build only after explicit human approval | Satisfies both in sequence; the research report ends with a go/no-go |
| 3 | POC location: `poc/payroll-month/` uncommitted (`prompt_claude`) vs `poc/` on branch `stack-poc` (`prompt_z`) | `poc/payroll-month/` on branch `stack-poc`, uncommitted until gate review | Satisfies both literally |
| 4 | D1 mechanism: don't pre-empt Gate 1 (`prompt_claude`) vs lean transaction-scoped/DBOS (DeepSeek, Z) | Research covers both XID-ticket and transaction-scoped options; ADR decides with a translated attack battery | Gate 1 decision is pending; pre-empting it is forbidden |
| 5 | DBOS "exactly-once steps" (`prompt_z`, DeepSeek) vs at-least-once ordinary steps (`prompt.md`) | `prompt.md` wins until disproven: atomicity only via DBOS datasource transactions; immudb appends are at-least-once with idempotent recovery — proven by crash injection | Affects the entire dual-write design; must be verified, not assumed |
| 6 | immudb via PG-wire protocol (DeepSeek) vs native Go SDK primary (`prompt.md`) | Native Go SDK primary; wire-protocol interface evaluated separately for explicit use cases only | immudb reports as PG14 and skips unsupported DDL — not PG18-compatible |
| 7 | Fate of migration `00013` | Decided by ADR-0202 after research; `00001`–`00012` frozen under all options | Unanimous across briefs |

## Steps

### Phase 0 — Business decision gate (from `prompt_z2` + `checklist/decisions.md`)

- Run the Design-Gate Checklist as a workshop artifact: answer the Ten Killer Questions and sections A–L for the payroll-month scope, or record explicit deferrals with owner + trigger.
- Settle the 8 priority decisions before any schema work: cutoff definition, post-cutoff corrections, finality lifecycle, adversarial-protection rung, authoritative input, retroactive-change representation, org boundary (one employer MVP), reproducibility evidence.
- Output: decision register (decision, owner, rationale, test, revisit trigger). Blocks Phase 5 schema/topology choices.

### Phase 1 — Codebase architecture map (all briefs §2/`3.1`/A)

- Map Go packages (`internal/payroll`, `internal/hr`, `internal/compbasis`, `internal/rules`, `core/database`, …), services, repos, error handling, config; every migration `00001`–`00012`; the complete writer matrix (service, trigger, DB function, script, direct SQL) × failure class (D1/P1a–g/EF-E3); capture path; lock/sync protocol; test battery (`tests/integrity`, `tests/e2e`).
- Every observation cites `file:line`; capabilities claimed only with executed evidence.
- Output: current-state map section of the research report.

### Phase 2 — PostgreSQL 18.6 assessment

- Verify against 18.6 release notes (not memory): `WITHOUT OVERLAPS` temporal keys, `PERIOD` FKs (including adjacent-row joint coverage), `uuidv7()`, virtual generated columns, `RETURNING OLD/NEW`, Docker image changes (`/var/lib/postgresql/18/docker` data dir, volume layout), initdb defaults, `SECURITY DEFINER`/privilege/advisory-lock/isolation behavior changes.
- Run migrations `00001`–`00012` on 18.6 on a disposable DB; record any test/assertion drift (e.g. pinned advisory counts).
- Assess 16→18.6 path (`pg_upgrade` vs dump/restore for this jump), extension inventory (`btree_gist`, `pgcrypto`), reindex needs; design temporal-constraint SQL for comp terms / SHG coverage / rule-pack validity with gap/overlap detection demos.
- Output: PG18 feature→requirement map, upgrade plan with rollback, incompatibility list (fix proposals only — frozen files untouched).

### Phase 3 — DBOS assessment (Go SDK at the local checkout)

- Record exact SDK version, min Go version, PG compat; map: what is checkpointed, where the system DB lives (same DB `dbos` schema vs separate DB — privileges, ops, and the "one database" constraint assessed), determinism/replay rules, versioning, queues, cron, durable sleep, dedup IDs, approval waits, cancellation.
- Critical experiment: can a business write + step checkpoint commit atomically (datasource transaction), or only via idempotency? Crash-injection at every boundary (before/after PG commit, before/after ledger append, before/after back-link); kill −9 worker; show recovery with no duplicate/missing records.
- Design candidate workflows (capture pipeline, month-run orchestration open→close, PG→immudb reconciliation, retro adjustments) with step/txn boundaries, retry policies, idempotency keys; verify transaction-scoped locks survive DBOS recovery on different connections and DBOS tables never enter the writer matrix.
- Output: workflow catalogue (durable-workflow-worthy vs plain service), failure/recovery matrix.

### Phase 4 — immudb assessment (local checkout + docs)

- First: read the local `LICENSE` text, record exact license + use-grant terms, flag constraints; stop and report if unacceptable (stop condition).
- Decide embedded library vs sidecar container (trust model, deployment, Go-version footprint, backup/upgrade, single-gated write path).
- Design provenance record schema (inputs hash, rule-pack version/digest, output hash, PG tx ref, actor, timestamp), back-link (PG row → immudb tx ID), dual-proof verification (`VerifyDualProof`/`VerifiedTxByID`) — including where the verifier's trusted state lives (held outside the writing app, anchored externally) and the immutability-vs-erasure rule (hashes/digests/decisions in; PII and erasure-subject data out).
- Verify `HISTORY OF` / `SINCE TX` / `UNTIL TX` behavior by experiment; name explicit non-use-cases (operational state, effective-dating, concurrency stay in PG).
- Output: evidence model, verification protocol, dual-write protocol (PG commit → ledger append → back-link) with crash recovery.

### Phase 5 — Integrated architecture, pattern catalogue, ADRs (research gate)

- Ownership table (PG = operational truth, DBOS = durable execution, immudb = verifiable evidence), data-flow + sequence diagrams, transaction boundaries, failure matrix (each component failing at each step), decision table for new features, per-technology use-when/do-not-use + worked example + anti-patterns.
- Rule-dynamics classification table: every hardcoded rule → (i) effective-dated PG18 data, (ii) versioned attested pack, or (iii) code in a versioned durable step; pack coverage enforced by DB constraint, not app code.
- ADRs with acceptance criteria pinned before POC: ADR-0201 stack supersede, ADR-0202 fate of `00013`, ADR-0203 capture mechanism (with translated 12-attack battery spec), ADR-0204 provenance schema; plus DeepSeek's defect→pattern→technology→enforcement→test mapping table and the 4-option trade-off (PG / PG+DBOS / PG+immudb / all three) with adopt/adopt-for-POC/defer/reject per technology.
- Outputs: `plans/platform_research_pg18_dbos_immudb.md` (Parts I–X per `prompt.md` §10), `docs/stack/RESEARCH.md`, `docs/stack/PATTERNS.md`, `docs/stack/adr/`, decision-request list for the human (license, trusted-state location, threat-model changes, upgrade timing). **Gate: human approval before Phase 6.**

### Phase 6 — POC build: one synthetic payroll month

- Bring-up: digest-pinned `postgres:18.6`, pinned immudb image, app + DBOS SDK; one-command start (compose or testcontainers), health checks, readiness ordering, goose flow, fresh-DB provisioning, teardown; "clone to green" runbook.
- Scenario matrix (merged, all synthetic): 25–50 employees, 2 pay groups; mid-month hire/leaver; mid-period comp change; backdated change → supersession chain (closed records never mutated); SHG/coverage union over adjacent parents incl. a gap + multi-fund case; SYNTH-OT + SYNTH-leave rules; pack vN→vN+1 mid-month with per-output pack citation and re-derivability; late correction before capture + post-close correction as next-period adjustment; period close to immudb with post-close posting rejection; adversarial writers during capture (direct SQL as `app` fails; forged ledger record detected); kill −9 at (a)/(b); auditor re-derivation script (snapshot hash from sources + pack, dual-proof verify); on-disk tamper → loud failure; rerun byte-identical.
- Pure-Go synthetic calculation function with test vectors; DBOS orchestration with per-employee fan-out + human-approval step; immudb attestations (pack digests, snapshot hash, decisions, close).
- Output: `poc/payroll-month/` + `poc/POC-REPORT.md` (F1–F7 tables, kill/tamper logs, attack mapping, deletion ledger, limitations).

### Phase 7 — Adoption & migration plan

- `docs/stack/MIGRATION.md`: PG16 supersede (dev-data path with parity check, old volume preserved, PG16 service removed), sequenced adoption with rollback per step (criteria → bring-up → PG18 constraints → DBOS wrap → immudb provenance → cutover + deletion), CI matrix on PG18 (kill-test job, tamper-test job, translated battery required), risk register (SDK maturity, immudb ops/licensing, dual-write under load, PG18 ext compat, DBOS privilege boundary), who-owns-what table + onboarding path.

## Validation Plan

- Per-phase: research claims tagged VERIFIED/DOCUMENTED/INFERRED with exact versions (PG, Go, DBOS SDK, immudb, drivers); negative results recorded; no claim from memory; no skipped-test evidence.
- POC experiments (each automated, deterministic, single-command, run across three time zones): crash injection at all 6 boundaries; tamper (PG row edit; ledger server edit; verifier-with-prior-state); concurrency writer-vs-capture both orderings; rule dynamics (new SYNTH rule type + pack version without schema/Go changes — report what still needed code); coverage block by DB constraint; idempotency (same period twice, same employee enqueued twice); rerun identical content hash; performance sanity timings (no optimization).
- Attack parity: 12 legacy Gate-1 attacks → translated tests green on the new stack, mapping table committed.
- Gates: `task build && task test && task lint` green at every code phase; DB-gated tests via `WORKFORCEOPS_TEST_DATABASE_URL`; Phase 5 research gate requires human approval; stop conditions (unviable tech, unacceptable license, threat-model breach, frozen-file edit required) halt with a report, never a workaround.

## Risks / Open Questions

- immudb license (BSL 1.1 + use grant per `prompt_claude`, unverified) may be unacceptable — first check in Phase 4, stop condition armed.
- DBOS Go SDK maturity/API drift vs briefs' assumptions (incl. exactly-once wording); pgx v5.12 (a2) vs v5.10 (DBOS) skew; Go 1.27.3 vs 1.25.0 toolchain.
- PG 16→18.6 jump mechanics unverified (`pg_upgrade` support for this jump, ext compat, advisory/isolation drift); PG19 `FOR PORTION OF` flagged future-only.
- Pending Gate 1 decision (XID-ticket vs transaction-scoped gate) constrains Phase 5; DBOS system-DB placement vs "one database" constraint needs an explicit ruling.
- Dual-write PG↔immudb has no native atomicity — the outbox + deterministic IDs + reconciliation design is the highest-risk validation step (crash matrix in Phase 3/6).
- Open: verifier trusted-state custody + external anchoring; threat-model change proposals (decision requests, never silent); `prompt.md`'s statutory-scenario items that have no synthetic analogue (listed as limitations, not guessed).
