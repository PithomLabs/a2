Below is the prompt, ready to paste. I checked three things against current sources, noted after the prompt. One thing I changed from what you asked: the M2.1 constraints (zero statutory calculations, OT-001 BLOCKED) conflict with "1-month payroll run", so the prompt makes the POC use clearly labeled synthetic rules.

````markdown
# Prompt: Platform Architecture Research and 1-Month Payroll POC (PostgreSQL 18.6 + DBOS + immudb)

## 1. Role and mission

You are a senior implementation and research agent working in the WorkforceOps repository (Go, pgx, goose, PostgreSQL, roles `migration_admin` and `app`). Do deep, evidence-based research on the codebase's architecture and determine where **PostgreSQL 18.6**, **DBOS Transact (Go SDK)** and **immudb** can materially make it more robust, flexible and extensible across **dynamic business rules**. Then prove the conclusions with a **proof of concept (POC): one simulated calendar month of payroll for one employer, with non-trivial complexity**.

Four goals, in priority order:

1. **Architecture research.** Map the current architecture and find concrete, justified places where each technology removes complexity, closes a defect class, or enables rule flexibility. Reject places where it doesn't.
2. **POC.** Demonstrate a 1-month payroll run end to end on the three-technology stack (section 6).
3. **Pattern catalogue.** For each technology, define the architectural patterns where it is best used, the anti-patterns, and what it must not be asked to do.
4. **Smooth integration.** Replace the local PostgreSQL 16 Docker setup with PostgreSQL 18.6 and make running all three technologies together as frictionless as possible (dev, tests, CI).

## 2. Context you must read first

- `AGENTS.md` (invariant-first principle), `plans/prompt17.md`, `plans/plan_prompt17.md` and amendments, `plans/milestone2_adversarial_review.md`, the Gate 1 package (`gate1_admission_package.md`) and the D1 review consolidation (`writeup4_review.md`).
- `architecture.md` (attached): a DeepWiki Q&A about immudb. **Treat every claim in it as an unverified hypothesis.** Its sources cover immudb only. Nothing in it was checked against DBOS code or documentation. In particular, verify rather than assume: idempotent immudb appends, "same input hash gives a detectable duplicate", DBOS "exactly-once" for non-database side effects, and temporal-query behavior.
- The migrations `00001`-`00012` and the existing Go packages (`internal/payroll`, `internal/hr`, `internal/compbasis`, `internal/rules`, `core/database`, `tests/integrity`, `tests/e2e`).

## 3. Hard constraints (inherited, non-negotiable)

- Do **not** modify migrations `00001`-`00012`, existing application packages, or Gate 1 artifacts. Do not pre-empt the pending Gate 1 decision on the D1 synchronization design (XID-ticket versus transaction-scoped employer gate). Where your research touches D1, describe how each technology interacts with **both** options and stop there.
- Work in a separate directory or worktree (for example `poc/payroll-month/`) and leave everything uncommitted.
- **No statutory calculations. SG-MOM-OT-001 stays BLOCKED.** The POC must use clearly labeled **synthetic, non-statutory rules** (for example "SYNTH-OT-001: 1.5x above 8 hours/day"). Mark every synthetic rule `synthetic: true` and test that no synthetic rule can be activated as statutory. The POC proves the architecture, not payroll correctness.
- Disposable PostgreSQL databases only, with fresh names. Never touch user databases. Drop-first is allowed only for names matching the test prefix.
- Preserve the two-role model and the threat model of prompt17 (hostile direct SQL as `app`) unless a finding justifies proposing a change. Any such proposal is a **decision request to the human**, never something you adopt silently.
- Any result that depends on a skipped test does not count as evidence.

## 4. Research questions (answer each with evidence)

### 4.1 Cross-cutting
- Where does the current architecture carry accidental complexity that these technologies would remove? Where would they add complexity without benefit?
- What is the smallest integration that delivers the benefit? Rank options by (value, cost, risk).

### 4.2 PostgreSQL 18.6
Verify against the 18.6 release notes and the PostgreSQL 18 documentation, not memory. Establish:
- Temporal constraints: `PRIMARY KEY/UNIQUE ... WITHOUT OVERLAPS`, `FOREIGN KEY ... PERIOD`. Can they replace hand-written overlap logic for compensation, SHG instructions and rule-pack pins? Can `PERIOD` foreign keys enforce full-period coverage by one or several adjacent referenced rows? What are the limits (range types, `btree_gist`, supported column types, `ON DELETE/UPDATE` actions)?
- Other 18 features that matter here: `uuidv7()`, virtual generated columns, `RETURNING OLD/NEW`, async I/O, skip scan, default data checksums, changes to `pg_upgrade`. Evaluate relevance only; do not adopt for novelty.
- Docker image changes for 18 (data directory and volume layout), initdb defaults, and any behavior changes affecting `SECURITY DEFINER` functions, privileges, advisory locks, or isolation semantics.
- Whether migrations `00001`-`00012` apply cleanly on 18.6 and whether any test or privilege assertion (for example the pinned advisory-overload count) changes. **Run them.**
- Do not depend on PostgreSQL 19 beta features. List any you want, flagged as future.

### 4.3 DBOS Transact (Go)
Verify against the current Go SDK source and docs, with the exact version, minimum Go version and PostgreSQL compatibility recorded.
- What exactly is checkpointed, where (the system database), and how does the system database relate to the application database? One database or two? What privileges does it need?
- Can a **business write and its step checkpoint commit atomically** in the Go SDK, or only be made safe through idempotency? This decides the dual-write story. Prove it with a crash-injection experiment.
- Determinism and replay rules: what may workflow code do? What happens on code changes while workflows are in flight (versioning)?
- Queues, concurrency limits, per-employee fan-out with child workflows, deduplication IDs, durable sleep and scheduling, notifications and events for human approval steps, cancellation, and workflow management.
- Failure semantics: step timeouts, retries, and what is replayed after a kill at each point.
- Operational footprint: the optional hosted control plane must not be required. Prefer the open-source library only.

### 4.4 immudb
Verify against the current release and license text.
- **License:** record the exact license of the version you would use and the use-grant terms, and assess fit for this project. Flag it prominently if there is any constraint.
- Embedded library versus server: consequences for the trust model, deployment, dependency footprint and Go version.
- Verified write and read paths, and the proof model. **Where does the verifier's trusted state live?** A tamper-evidence claim is only meaningful if someone other than the writing application holds a previously trusted state. Design and test where that state is kept and how it is anchored externally.
- Idempotency and preconditions for appends, behavior on partial failure, and ordering.
- Which data belongs in the ledger (hashes, digests, decisions, rule-pack attestations, period-close records) and which must stay out (personal data, anything subject to erasure obligations). Raise the immutability-versus-erasure conflict explicitly.
- Query features (`HISTORY OF`, `SINCE/UNTIL TX`) as actually behaving, not as described.

## 5. Pattern catalogue (required output)

For each technology, give: best-fit patterns, anti-patterns, a one-paragraph "use when / do not use when", and one worked example from this codebase. Include at minimum:
- PostgreSQL 18: the state-of-truth role, declarative invariants versus code, effective-dating, how pack and instruction coverage is enforced.
- DBOS: the durable payroll-period workflow (open, capture, validate, calculate, approve, close), per-employee fan-out, saga-style compensation for failures, and approval waits. State clearly what DBOS does **not** replace (database-level barriers and isolation for snapshot freshness).
- immudb: the evidence ledger for attestations, snapshot hashes, period-close records and rule-pack digests, with the dual-write protocol (Postgres commit, ledger append, back-link) and its crash recovery.
- Cross-technology: ownership table (what each owns), data-flow diagram, failure matrix (what happens if each component fails at each step), and a decision table for new features ("where does this go?").

## 6. The POC: a 1-month payroll run

Scope: one employer, 25-50 synthetic employees, one calendar month with daily attendance, simulated time (no real-time waiting). It must include, at minimum:
- a mid-period compensation change and a mid-period termination
- SHG or fund variations including a gap and a multi-fund case
- a synthetic overtime-like rule and a synthetic leave rule
- a rule-pack version change effective mid-month
- a late attendance correction before capture and another after close (handled as an adjustment in the next period)
- a rerun producing an identical content hash (reproducibility)
- DBOS workflow orchestration with per-employee fan-out and a human-approval step
- immudb attestations for pack digests, snapshot hash, decisions and period close
- calculation using only synthetic rules, as a pure Go function with test vectors

Required experiments (each with recorded evidence):
1. **Crash injection:** kill the process at every step boundary (before and after the Postgres commit, before and after the ledger append, before and after the back-link); show recovery completes with no duplicate or missing records.
2. **Tamper detection:** alter a snapshot row directly; alter ledger data on the server; show detection by the verifier, including the case where the verifier holds prior state.
3. **Concurrency:** a writer racing capture, in both orderings, under the isolation approach you evaluate.
4. **Rule dynamics:** add a new synthetic rule type and a new rule-pack version without schema or Go changes beyond the rule definition; report what, if anything, still required code.
5. **Coverage:** a rule pack that does not cover the full period is blocked by a database constraint, not by application code.
6. **Idempotency:** run the same period twice, and enqueue the same employee twice; show exactly-once outcomes.
7. **Performance sanity:** report timings and workflow counts. Do not optimize.

Acceptance: every experiment is automated, deterministic and repeatable from one command, and is run across the three time zones used by the existing battery.

## 7. Integration and environment (replace local PostgreSQL 16)

- Pin PostgreSQL **18.6** by image digest. Provide one compose file or a test harness (for example testcontainers) that starts PostgreSQL 18.6, immudb and any DBOS requirements with one command.
- Document and test: roles and privileges, the DBOS system database placement, connection strings, migration flow with goose, health checks, readiness ordering, volume layout changes, and teardown.
- Update provisioning so acceptance runs use fresh database names and cannot reuse stale schemas.
- Report any incompatibility between existing code and PostgreSQL 18.6, with a fix proposal but no changes to frozen files.
- Provide a short "from clone to green" runbook.

## 8. Evidence and honesty rules

- Tag every claim **VERIFIED** (reproduced by an experiment you ran), **DOCUMENTED** (with URL, version and access date) or **INFERRED** (reasoning only). No version or behavior claim from memory.
- Include negative results. If a technology does not fit somewhere, say so.
- Record exact versions: PostgreSQL, Go, DBOS SDK, immudb, drivers.
- Do not copy long passages from documentation; summarize and link.

## 9. Deliverables

1. `plans/platform_research_pg18_dbos_immudb.md`: the research report (current-state map, findings per technology, pattern catalogue, risks, recommendation per technology: adopt / adopt for POC only / defer / reject).
2. Draft ADRs for each recommended adoption, each with context, decision, alternatives, consequences, and a revisit trigger.
3. The POC under `poc/payroll-month/`, with tests and a single-command runner.
4. The integration runbook and environment files.
5. A **decision-request list** for the human: license acceptability, trusted-state location, threat-model changes, upgrade timing, anything needing approval.
6. A limitations section: what the POC does not prove.

## 10. Stop conditions

Stop and report instead of proceeding if: a required technology is not viable for the stated use; the immudb license is unacceptable; DBOS cannot be made safe without a threat-model change; or any step would require editing frozen files.
````

Checks I ran against current sources:
- **PostgreSQL 18.6 exists.** It was released on 2026-08-13, alongside 17.11 and 16.15. One side effect: the release notes say the 18.4 to 18.6 range was skipped. The prompt tells the agent to record exact versions rather than assume.
- **DBOS has a Go SDK.** It is a lightweight, MIT-licensed library that checkpoints workflow state in Postgres. The versions I saw were pre-1.0, so the prompt asks the agent to confirm the current version and its maturity.
- **immudb's license changed.** Releases from 1.9DOM.2 on are under the Business Source License 1.1 with an additional use grant, converting to Apache-2.0 four years after each version's publication. That's a source I haven't read directly, so I made license review the first item in the immudb section and a stop condition.

I also added one point that `architecture.md` doesn't cover: immudb's tamper evidence only holds if someone other than the writing application keeps a previously trusted state. The prompt makes the agent design and test where that state lives.