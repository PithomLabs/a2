# Implementation-Agent Brief — PostgreSQL 18 + DBOS + immudb

*Usage: paste this entire document into your implementation agent. If available, also attach: `prompt17.md`, `plan_prompt17.md`, `gate1_admission_package.md`, and `architecture.md` (the immudb/PostgreSQL-18 research notes). The brief is self-contained otherwise.*

---

## 0. Mission

You are a senior systems architect and implementation agent working in the WorkforceOps repository (Go + PostgreSQL payroll system). You have four objectives, in this order:

1. **DEEP RESEARCH** the codebase architecture and determine **where and whether** PostgreSQL 18 (target: latest 18.x point release, nominally 18.6), **DBOS** (durable execution), and **immudb** (immutable ledger DB) can **materially** make this codebase more **robust** (integrity, recovery), **flexible** (dynamic business rules), and **extensible** (new writers, new rule packs, new pay components without re-proving everything). "Materially" has a strict definition in §2 — vague modernization claims will be rejected.
2. **PATTERN MAP**: identify the architectural patterns where each technology's paradigm is applied **within its best use case** — and name the anti-patterns where it must NOT be used.
3. **POC**: build a proof-of-concept that runs **one full month of payroll** with deliberately non-trivial complexity (§5.2 scenario matrix), on the superseding stack, with Gate-1-grade evidence.
4. **ADOPTION PLAN**: make the integration of the three technologies as smooth as possible, **superseding the local PostgreSQL 16 Docker setup** — one-command bring-up, data migration path, CI update, rollback at every step.

Everything you produce must obey the method rules in §2. The ideal outcome of this engagement is that significant hand-rolled mechanism **gets deleted** — simplification is a first-class deliverable, not a side effect.

---

## 1. Context you must load

### 1.1 The system and its history

- **WorkforceOps**: Go service + PostgreSQL. Payroll snapshots must be complete, authentic, reproducible, and governed. Business logic lives in Go; the trust boundary lives in the database (ADR-1).
- **Current stack**: PostgreSQL **16** running in local Docker — this is what you will supersede.
- **Role model**: `app` (untrusted application role, limited DML) and `migration_admin` (owns DDL, triggers, and security-critical functions). Triggers must be owned by a non-`app` role so `app` cannot disable them.
- **Milestone 2.1 (M2.1) remediation context** — the reason this brief exists:
  - **D1 defect**: PostgreSQL establishes a REPEATABLE READ snapshot at the *first statement*, not at `BEGIN`; advisory-lock waits never refresh snapshots, so "lock, then read" reads stale data.
  - **Gate-1 XID-ticket admission mechanism** (already built, deliberately unwired): converts lock ordering into visibility via a ticket row committed in READ COMMITTED, checked with `pg_visible_in_snapshot`. Evidence: 31 green mechanism proofs + 12 pinned-red attack tests.
  - **DISCARD ALL finding**: session-level advisory locks are unprotectable (`DISCARD ALL` releases them, cannot run in a transaction, cannot be revoked) — the structural objection to session-lock designs; transaction-scoped locks are immune.
  - **Planned migration 00013**: revoke all 21 `pg_advisory_*` overloads from PUBLIC; redefine helpers as SECURITY DEFINER; trigger-enforced writer participation across the writer matrix.
  - **Authority invariants (EF-E3/EF-F1)**: exclusive creation authority over snapshot content; "derive, don't accept" (capture derives values from source records, never accepts caller JSON). These are **P0 under every option** — do not propose weakening them.
  - **Writer matrix (P1a–g)**: the snapshot invariant is only as strong as the *least-cooperative writer*; every mutation path on snapshotted inputs must join one protocol.
  - **Spec discipline**: if `prompt17.md` exists in the repo, read it first. Its §3 permanent requirements and §15 adversarial acceptance standard **win over this brief** wherever they conflict; its §4/§5.2 frozen *mechanisms* may be superseded **only via an explicit ADR with a translated attack battery** (§2 rule 5).

### 1.2 The three technologies (verify everything yourself; pin versions at bring-up)

- **PostgreSQL 18.6** — pin the exact 18.x image tag/digest that actually exists at pull time and record it. Features known to be relevant (verify against official release notes before relying on any):
  - **Temporal keys**: `WITHOUT OVERLAPS` on PRIMARY KEY/UNIQUE, and `PERIOD` foreign keys (adjacent referenced rows may jointly cover a referencing period — directly relevant to coverage unions and pack full-period coverage).
  - PG19 is expected to add `UPDATE/DELETE ... FOR PORTION OF` — design tables now so that DML becomes a cheap upgrade later.
  - `uuidv7()` (idempotency keys), async I/O (`io_method`), OAuth auth, B-tree skip scan, virtual generated columns (17+).
- **DBOS** — durable execution built on Postgres: durable workflows, steps with exactly-once semantics, queues/cron, automatic crash recovery, workflow state stored in Postgres itself. For Go, verify the current SDK artifact (e.g. `dbos-transact-golang` or successor), its pinned version, maturity, and license.
- **immudb** — open-source immutable database: append-only (new versions, never update/delete), linear hash chain + Merkle-tree **"dual proofs"** verified **client-side without trusting the server** (`VerifyDualProof`, `VerifiedTxByID`), SQL + KV + document models, temporal queries (`SINCE TX`, `UNTIL TX`), `HISTORY OF` / `DIFF OF`, embeddable Go library **or** standalone sidecar service.

### 1.3 Working hypothesis (verify and adjust with evidence — do not assume)

- **PostgreSQL 18** = system of record; temporal/effective-dating constraints; concurrency control; snapshot/capture compute.
- **DBOS** = durable-execution glue: multi-step payroll runs, exactly-once cross-system steps, crash recovery, scheduling, and the **reconciliation engine** for the PG↔immudb dual-write.
- **immudb** = tamper-evident provenance and attestation ledger: output hashes, rule-pack versions, period-close records, auditor verification path. **Never** operational state, effective-dating, or concurrency — those stay in Postgres.
- **The known integration crux**: the dual-write problem between a Postgres commit and its immudb provenance append. Solve it inside DBOS durable steps (checkpoint → idempotent append → back-link of the immudb tx ID onto the Postgres row). A bare "write both" call is a defect, not a design.

---

## 2. Method rules (non-negotiable)

1. **Materiality bar.** A technology "materially helps" only if it does at least one of:
   (a) **deletes** a hand-rolled mechanism or closes a defect class we currently defend with code and tests;
   (b) removes a substantial amount of mechanism or test burden (≥200 LOC or its battery) while adding fewer new moving parts;
   (c) unlocks a capability we cannot otherwise ship (e.g., dynamic business rules) at bounded complexity.
   Quantify every verdict. "It's modern" is not evidence.
2. **Evidence-first.** Every claim cites `file:line` in the codebase, or an upstream doc with version and link. Anything unverifiable is marked `ASSUMPTION`.
3. **Rung-4 evidence.** No robustness claim without a test that exercises its failure mechanism (red-first where feasible). Translate the existing 12-attack battery to the new stack: each legacy attack → new equivalent test, or a demonstrated proof that the class is closed by construction.
4. **Pin criteria before experiments.** Acceptance criteria go into ADRs *before* the POC is built (no post-hoc goalpost moves).
5. **Invariants over mechanisms.** The four snapshot invariants hold under any stack: writer-first inclusion (complete), exclusive creation authority (authentic), derivable content hash (reproducible), universal writer participation (governed). Mechanisms may change; invariants may not.
6. **Least-cooperative-writer rule.** Every proposed pattern must state how out-of-band writers are handled: direct SQL as `app`, migrations, other services, future scripts.
7. **No silent dual-writes** across Postgres/immudb, ever.
8. **Keep a deletion ledger.** Every piece of M2.1 mechanism that the stack replaces (barrier protocol? ticket mechanism? trigger battery? audit scaffolding?) is listed as deleted/kept with quantified rationale.
9. **Honest no-go.** If a technology fails the materiality bar for this codebase, say so, scope it out of the POC, and record why. Do not force adoption to please the brief.
10. **Stop condition** (in the spirit of prompt17 §17): if evidence contradicts a pinned assumption, stop, report the contradiction, and propose options. Do not improvise around it.

---

## 3. Phase 1 — Deep architecture research → `docs/stack/RESEARCH.md`

Answer the following with evidence. Timebox: ~2–3 agent workdays equivalent.

**3.1 Codebase inventory.** Data model; migrations 00001–00012 (+00013 if written); the complete **writer matrix** — every path that mutates snapshotted inputs (services, triggers, DB functions, scripts), each mapped to its failure class (D1 / P1a–g / EF-E3); the capture path; the audit log; current lock/sync protocol; the test battery structure.

**3.2 PostgreSQL 18 materiality.** For each candidate feature, verdict = materially-helps / neutral / not-applicable, with the concrete tables and defect classes affected:
- Temporal keys: `WITHOUT OVERLAPS` for compensation terms / SHG coverage / rule-pack validity periods; `PERIOD` FKs for "this pack covers the full period".
- Whether any MVCC, advisory-lock, or snapshot-behavior change between 16 and 18 affects the D1 analysis or the Gate-1 mechanism — verify explicitly, do not assume.
- `uuidv7()`, virtual generated columns (e.g., for canonical byte forms), async I/O, OAuth, skip scan — evaluate only where they touch our invariants or performance envelope.
- **Upgrade mechanics 16 → 18.6**: `pg_upgrade` vs dump/restore (verify which is supported for this jump), extension inventory (`btree_gist`, `pgcrypto`, …) compatibility, reindex requirements, downtime profile for a dev environment.

**3.3 DBOS materiality.** Verify the Go SDK: artifact name, pinned version, maturity signals (releases, issues), license. Map durable workflows onto concrete flows:
- capture pipeline (gate → read inputs → canonicalize → insert snapshot → audit) as a durable workflow;
- the **full month-run orchestration** (open → collect → capture → compute → close) with per-step retries;
- the PG→immudb dual-write as exactly-once steps;
- retro-run / adjustment flows after period close;
- queues/cron for scheduled runs; idempotency keys (`uuidv7()`).
- **Interplay with our gate**: DBOS recovery may resume steps on different connections. Verify transaction-scoped locks survive this; confirm nothing session-scoped is required. Confirm DBOS system tables never touch business tables (they must not appear in the writer matrix).

**3.4 immudb materiality.** Provenance record schema (inputs hash, rule-pack version, output hash, PG tx ref, actor, timestamp); back-link (PG row → immudb tx ID); dual-proof verification flow for auditors; `HISTORY OF` / `SINCE TX` for reconstructing ledger state; rule-pack attestation; period-close records. Deployment decision: **sidecar container vs embedded Go library** — evaluate on backup/upgrade story, verification ergonomics, and keeping the write path single-gated. Also name where immudb is **not** applicable in this codebase.

**3.5 Dynamic business rules (the flexibility axis — be thorough).**
Inventory every rule currently hardcoded in Go (eligibility, proration, rounding, supersession, approval thresholds, pay components, pack coverage). Classify each into exactly one of:
(i) **becomes effective-dated data in PG18** protected by temporal constraints;
(ii) **becomes a versioned rule pack** with immudb attestation ("output X was produced under pack vN");
(iii) **remains code**, but wrapped in a DBOS durable step so executions are versioned and replayable.
Deliver this as a classification table. This table is the heart of the "flexible and extensible" objective.

**3.6 Gap analysis / deletion ledger.** What M2.1 hand-rolls that this stack would replace, and what it must keep (e.g., EF-E3 revokes and privilege separation stay under every option). Quantify: LOC deleted, attack surface removed, test burden changed.

**3.7 Verdict.** A one-page materiality matrix: rows = technology features, columns = robust / flexible / extensible, cells = material gain with citation, or "—".

---

## 4. Phase 2 — Pattern map & ADRs → `docs/stack/PATTERNS.md` + `docs/stack/adr/`

**4.1 Decision matrix.** One row per architectural concern; columns: owner technology | why (best use case) | anti-pattern warning | failure mode if misapplied. Cover at minimum: system of record, effective-dating & overlap prevention, concurrency gate, capture compute, run orchestration, retries/recovery, scheduling, provenance & attestation, tamper evidence, auditor access, rule-pack distribution, period close, retro adjustments.

**4.2 Claim inventory with rung assignment.** Restate the correctness claims as falsifiable sentences and assign each a reliability-ladder rung: CRUD/report rendering = rung 2; proration/supersession/conservation arithmetic = rung 3 (property tests); capture freshness, exclusive authority, tamper evidence = rung 4 (attack battery). The POC's test depth follows these rungs.

**4.3 ADRs** (each with pinned acceptance criteria and a red-test battery spec):
- ADR-0201 — Stack supersede: PostgreSQL 18.6 + DBOS + immudb replaces PG16-only local stack.
- ADR-0202 — Fate of migration 00013 under the new stack (which parts become obsolete, which stay; prove by test, don't assume).
- ADR-0203 — Capture mechanism under DBOS (if superseding the Gate-1 ticket mechanism: show the translated 12-attack battery passes and the D1 class is closed by construction; otherwise keep the gate and wrap it).
- ADR-0204 — Provenance & attestation schema in immudb (record format, back-links, verification protocol).

**Checkpoint:** Phases 1–2 produce review artifacts **before** Phase 3 begins. If running autonomously with no reviewer available, apply §2 rules to self-approve, and record the self-approval decision and its evidence in the ADR.

---

## 5. Phase 3 — POC: one month of payroll → `poc/` on branch `stack-poc`

### 5.1 Stack bring-up

`docker-compose.yml` (or equivalent) with pinned images: `postgres:18.x` (record exact tag), `immudb` official image (pin), and the app service with the DBOS Go SDK. DBOS system tables in a dedicated schema (e.g. `dbos`) on the same Postgres instance; immudb as sidecar with its own volume. Healthchecks, one-command `up`, seeded fixture, `make poc-run`. A clean machine must go from clone → month-run report with one command plus documented prerequisites.

### 5.2 The month — scenario matrix (non-trivial complexity is mandatory)

A seeded month (e.g., 25 employees, 2 pay groups) that exercises, at minimum:

1. Baseline month opens; population captured under the new stack.
2. **Mid-month hire** — proration from start date.
3. **Mid-month leaver** — proration to end date.
4. **Backdated compensation change** effective mid-previous-period → supersession chain; closed records are never mutated (append + version).
5. **SHG/coverage union** spanning two adjacent parent records (exercises PG18 `PERIOD` FK joint-coverage semantics).
6. **Rule-pack vN → vN+1 effective mid-month**; every output cites the pack version that produced it; both versions remain re-derivable.
7. **Period close** — close record written to immudb; post-close posting attempt must fail in Postgres; late changes land as next-period adjustments.
8. **Adversarial writers during capture**: direct SQL as `app` against snapshotted inputs must fail (privilege/trigger); a forged immudb record must be detectable by proof verification.
9. **Crash tests**: `kill -9` the worker (a) between PG commit and immudb append, (b) between append and back-link. DBOS recovery must complete the chain exactly once — no provenance gap, no orphan proof, no double execution.
10. **Auditor path**: an independent script re-derives the month snapshot hash from source records + rule pack and verifies immudb dual proofs.
11. **Tamper test**: modify immudb data on disk → verification must fail loudly.
12. **Replay**: re-run the month from workflow history; outputs byte-identical (reproducibility).

### 5.3 Acceptance criteria — pin these BEFORE building (extend, don't weaken)

- **F1 Oracle-correct**: outputs match a hand-computed fixture on the seeded dataset.
- **F2 Attack parity**: each of the 12 legacy Gate-1 attacks has a translated test, green on the new stack; the old→new mapping table is committed.
- **F3 Durability**: kill-tests pass; exactly-once demonstrated from workflow history evidence.
- **F4 Tamper-evidence**: tamper test fails verification loudly.
- **F5 Dynamic rules**: pack swap changes outputs; provenance attests pack version; old outputs re-derive.
- **F6 One-command bring-up** on a clean machine; README with prerequisites.
- **F7 Deletion ledger**: every M2.1 mechanism deleted/kept, with rationale.
- Optional (rung 4.5): a 1–2 page PlusCal/TLA+ model of capture ordering under the new stack.

### 5.4 Evidence

`poc/POC-REPORT.md` in the Gate-1 style: green/red tables per criterion, kill-test and tamper-test logs, the old→new attack mapping, the deletion ledger, and an honest limitations section.

---

## 6. Phase 4 — Adoption & migration plan → `docs/stack/MIGRATION.md`

1. **Supersede PG16 Docker**: upgrade mechanics for dev data (verify `pg_upgrade` 16→18 vs dump/restore; whichever is used, include a data-parity check script), extension compatibility, image pinning, compose service removal for PG16, README/CI updates. "Superseded" means: the PG16 service no longer exists in the dev stack.
2. **Sequenced adoption with rollback per step**: (0) criteria pinned → (1) stack bring-up → (2) PG18 features first (temporal constraints on comp/SHG/packs — independent of the gate) → (3) DBOS wrapping capture & run orchestration → (4) immudb provenance & attestation → (5) cutover + deletion of superseded hand-rolled code. Each step: what proves it worked, how to roll back.
3. **CI**: matrix on PG18; a kill-test job (chaos step); a tamper-test job; the translated attack battery in the required pipeline.
4. **Risk register**: DBOS Go SDK maturity; immudb operational burden (backup/export/upgrade story — verify); dual-write semantics under load; PG18 extension compat; licensing of all three (verify current terms); team learning curve; interplay of DBOS internals with the security boundary (verify DBOS never bypasses the gate or gains privileges `app` lacks).
5. **Team-facing docs**: a "who owns what" table (Postgres/DBOS/immudb/application) and a new-developer onboarding path.

---

## 7. Out of scope

Kubernetes/HA/clustering, multi-region, SaaS hardening, UI work, production statutory content maintenance, analytics/data-lake concerns, rung-5 formal verification of application code (optional PlusCal model excepted).

---

## 8. Deliverables & definition of done

- [ ] `docs/stack/RESEARCH.md` — §3 complete, materiality matrix included, every cell cited
- [ ] `docs/stack/PATTERNS.md` — decision matrix + anti-patterns + claim inventory with rungs
- [ ] `docs/stack/adr/ADR-0201…0204` — pinned criteria, translated battery specs
- [ ] `poc/` on branch `stack-poc` — one-command bring-up; scenario matrix §5.2 demonstrated end-to-end
- [ ] `poc/POC-REPORT.md` — F1–F7 green/red tables, kill & tamper logs, attack-mapping table, deletion ledger
- [ ] `docs/stack/MIGRATION.md` — PG16 supersede path, sequenced adoption, rollback per step, risk register
- [ ] Final chat summary ≤ 500 words with artifact links

## 9. Final report format

Summary (≤500 words): verdicts per technology (material / not material, where, why — one line each with citation), what got simpler (deletion ledger highlights), POC result (criteria table), the single riskiest remaining item, and the recommended first adoption step. Link every artifact. Close with the honest list: what this stack does **not** solve, and what remains hand-rolled by design.
