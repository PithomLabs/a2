The next task is Gate B: execute the actual database-backed POC and collect evidence for Gate C adoption decisions.

Gate A is complete. Step 0 is signed. The agent must not repeat the previous research, reopen approved decisions, or waste time rebuilding what has already been proven.

I would give the next agent this prompt. It is designed to work on a Docker-capable host and to stop honestly if that prerequisite is unavailable.

# GATE B EXECUTION — PostgreSQL 18.6 + DBOS + immudb

## Objective: Execute, Adversarially Verify and Complete the One-Month Synthetic Payroll POC

You are the implementation and integration verification agent for the existing workforce-management project.

You are continuing Phase 6 on branch `stack-poc`.

Your assignment is to execute Gate B of the consolidated three-gate plan.

**Gate A is COMPLETE. Step 0 is FULLY SIGNED. Gate B is the only active assignment.**

Your objective is to establish, through actual execution rather than source inspection or compilation, that the integrated PostgreSQL 18.6 + DBOS + immudb architecture works correctly under realistic payroll complexity, concurrency, crash recovery, adversarial writes and independent evidence verification.

The ultimate deliverable is an executable, fully documented one-month synthetic payroll POC.

---

# 1. AUTHORITATIVE PROJECT STATE

Read these files before doing anything:

- `plans2/plan_prompt2.md`
- `poc/payroll-month/POC-REPORT.md`
- `poc/payroll-month/POC-RUNBOOK.md`
- `docs/stack/MIGRATION.md`
- `docs/stack/TRACEABILITY.md`
- `docs/stack/DECISIONS.md`
- `docs/stack/DECISION-REQUESTS.md`
- `docs/stack/RESEARCH.md`
- `docs/stack/PATTERNS.md`
- `docs/stack/RULES-DYNAMICS.md`
- `docs/stack/EXPERIMENTS.md`
- `docs/stack/ADR-0201.md` through `ADR-0204.md`

Inspect the actual repository instructions and working tree.

The following Gate A results are already established. Treat them as completed work unless new contradictory evidence appears.

### A. Hygiene and build

- Mechanical `a1` → `a2` import-path rename: 95 files, 258/258 imports, zero non-import lines changed.
- Root database test skip fix.
- DBOS Transact Go v1.8.0 verified.
- immudb v1.11.2 verified.
- pgx v5.12.0 compatibility proven.
- Both image digests retrieved from the Hub API and pinned in Compose.
- Both modules build.
- `go vet` and `gofmt` clean.

### B. Architecture and documentation

- Decision traceability matrix completed.
- ADR-0204 rewritten with event-specific provenance keys, append-only back-links, digest-bound markers and calculator digest pinning.
- ADR-0203 extended to T13 and separate-database fallback.
- Ticket disposition documented as recorded-note, not bake-off.
- Dynamic rules catalogue contains 15 classified rule types.
- E1–E6 and DBOS/crash/tamper/forgery/coverage experiments documented.
- F1, F1a and F1b separated.
- Canonical-table hygiene and rung vocabulary documented.
- P1–P5 crosswalk and provisional choices recorded.

### C. Implemented code

- Gate-2 POC DDL: gate, 12 participation triggers, claims/proof-check, close transaction and T13 revokes.
- Tx-threading refactor completed through thin adapters.
- External auditor and genesis procedure implemented.
- PostgreSQL transactional outbox worker implemented.
- Three deterministic failure points.
- Kill matrix.
- Forged-state tests using actual `dbos.*` columns.
- Double-enqueue tests.
- Performance tripwire.
- Data-only rules tests.
- `poc-run` split into synthetic demonstration and full fail-loud mode.
- Capture-once and outbox-once subcommands.
- CI workflow YAML validated.

### D. Verified local evidence

- Synthetic calculation vectors: 10/10.
- Pure-Go reproducibility and boundary checks.
- Demo digest remains stable: `259564d6…`.
- Both modules build.
- `go vet` and `gofmt` pass.
- Database migrations 00001–00012 untouched.
- No migration 00013 exists.

Full evidence and exact implementation status are in `POC-REPORT.md`.

Do not repeat these tasks just to produce another report.

---

# 2. BINDING HUMAN RULINGS — DO NOT REOPEN

All R1–R7 rulings were approved and recorded on November 10, 2026.

The following are binding constraints for Gate B.

**R1 — immudb license**

Approved with conditions for non-production synthetic POC use under BSL-1.1.

- immudb v1.11.2.
- Change Date estimated at approximately 2030-09-02.
- Counsel review required before hosted or paid use.
- No vendoring.
- Narrow replaceable interface retained.
- No production authorization.

**R2 — Trusted-state custody**

- Independent external auditor.
- Out-of-band genesis checkpoint.
- Trusted verifier state external to the writing application.
- Follow the approved genesis and key-custody procedure.
- Record the explicit threat boundary.

**R3 — Threat model**

- Retain Rung 5 for authoritative payroll writes and artifacts.
- DBOS system tables are untrusted for business authority.
- No Rewind, Delete or truncate powers for the application runtime.
- No weakening of database enforcement.
- Distinguish application-role protection from privileged-host compromise.

**R4 — PostgreSQL**

- PostgreSQL 18.6 approved for a disposable POC cluster.
- PG16 remains preserved.
- No PG16 volume reuse or destructive migration.
- Production upgrade not authorized.

**R5 — Gate 2**

- Gate-2 transaction-scoped direction approved as POC evidence only.
- Production Gate 1 remains pending.
- Do not treat POC results as production authorization.

**R6 — DBOS placement**

- Shared pool and same-database `dbos` schema are preferred for the atomic capture path.
- Separate-database fallback is documented.
- Verify the actual privilege boundary.
- Do not change the database architecture merely to simplify deployment.

**R7 — Business semantics**

Two tranches:

Signed decisions: D1, D3, D4, D5, D7 and D8.

Deferred business semantics include proration, X-FUND and the remaining unresolved cutoff/coverage questions.

Do not silently invent these rules.

Mid-period changes with unresolved proration semantics must block.

**P1–P5:** Confirmed as documented, using recorded-note ticket disposition, not bake-off.

Consult `DECISION-REQUESTS.md` and `DECISIONS.md` for the authoritative wording.

These decisions are not invitations to restart the decision-making process.

---

# 3. FIRST: VERIFY THE EXECUTION ENVIRONMENT

Before running Gate B:

1. Verify the active branch is `stack-poc`.
2. Inspect `git status`.
3. Preserve every existing uncommitted change.
4. Verify Docker Engine and Docker Compose are actually usable.
5. Verify registry connectivity.
6. Verify sufficient disk and memory for the POC.
7. Verify PostgreSQL 18.6 and immudb v1.11.2 image digests against the recorded pins.
8. Verify that PostgreSQL and immudb services can be started.

Do not repeatedly search for unavailable container runtimes.

If this environment does not have an accessible Docker daemon and registry connectivity, stop immediately after one bounded capability check.

Do not install Docker, attempt sudo workarounds, provision cloud infrastructure or invent a remote execution host without authorization.

Produce an explicit blocker report identifying the external prerequisite.

Do not claim Gate B has started or passed.

If the environment is capable, proceed with the full execution.

---

# 4. GATE B STEP 1 — ISOLATED STACK BRING-UP

Follow `poc/payroll-month/POC-RUNBOOK.md` in the documented order.

Use the existing pinned image digests.

Establish:

- PostgreSQL 18.6.
- DBOS system schema.
- immudb 1.11.2 standalone service.
- Application connection.
- External auditor.
- Trusted genesis checkpoint.

Requirements:

- Explicitly named disposable volumes.
- Isolated Compose project identity.
- Explicit health checks.
- Deterministic initialization.
- Correct roles and grants.
- No superuser application runtime.
- No PII in immudb.
- No dependency on PG16 data.

Verify that restarting every service preserves the expected state.

Test service unavailability and restoration.

Do not run the migration or teardown procedure against the existing PG16 environment.

---

# 5. GATE B STEP 2 — POSTGRESQL 18.6 EXPERIMENTS

Execute E1–E6 as specified in `docs/stack/EXPERIMENTS.md`.

The most important questions:

- Do the existing migrations 00001–00012 apply cleanly?
- Do existing assertions and tests retain their intended meaning?
- Do the relevant temporal keys work with actual range types?
- Do `WITHOUT OVERLAPS` constraints enforce the expected no-overlap conditions?
- Do `PERIOD` foreign keys reject incomplete coverage?
- Can adjacent effective-dated records provide valid joint coverage?
- Do the existing compensation and SHG live-basis triggers remain correct?
- Are there unexpected changes in transaction locking or advisory-lock observations?

Run the existing database suite against PostgreSQL 18.6.

Perform the PG16/18.6 canonical snapshot parity exercise using the MERIDIAN fixture.

Do not modify migrations 00001–00012.

Do not create migration 00013.

Do not adopt new temporal DDL merely because it compiles.

Every proposed constraint must have executable gap, overlap, containment and boundary tests.

---

# 6. GATE B STEP 3 — DATABASE ENFORCEMENT AND T1–T13

Execute the translated adversarial battery specified in ADR-0203 and its latest T13 extension.

Test using the real `app` role.

Never perform the adversarial probes exclusively as a superuser.

Verify every relevant writer and transition boundary.

At a minimum:

- Stale-RR admission protection.
- Approval versus capture.
- Employment-terms closure versus capture.
- Direct disposition insertion.
- Direct status-history insertion.
- Rule-pack activation races.
- Draft-line DML.
- Period/employer unauthorized updates.
- Snapshot fabrication.
- Direct run-state promotion.
- Vacuous validation.
- Cutoff races.
- Additional T13 revocation and privilege checks.

Exercise writer-first and capture-first interleavings wherever specified.

Use deterministic synchronization rather than timing-dependent sleeps.

For each case:

1. Establish initial state.
2. Start the competing transactions.
3. Force the specified interleaving.
4. Execute both operations.
5. Verify persisted data.
6. Verify expected rejection, inclusion or blocking.
7. Verify no partially authorized artifact exists.

Record database state before and after.

The goal is to prove that an application role cannot bypass the authoritative database boundaries through direct SQL.

Any red test involving an actual invariant is a blocker.

Do not weaken assertions to obtain green results.

---

# 7. GATE B STEP 4 — DBOS INTEGRATION AND CRASH RECOVERY

Use the exact DBOS Transact Go version pinned by Gate A.

Verify the actual shared-pool transaction boundary.

The critical capture path must use the DBOS-supplied transaction consistently.

Prove that the following commit together:

- Gate acquisition.
- Claim and proof checks.
- Approved input resolution.
- Population resolution.
- Snapshot construction.
- Snapshot inserts.
- Rule pins.
- Outbox insertion.
- DBOS durability checkpoint.

Execute DBOS experiments DBOS-1 through DBOS-8.

Pay particular attention to:

- Crash before transaction commit.
- Crash immediately after transaction commit.
- Worker termination during a durable step.
- Duplicate enqueue.
- Worker ownership and fencing.
- Workflow restart and recovery.
- Workflow code-version/replay behavior.
- Waiting for approval and resuming the workflow.

Verify that approved artifacts are never silently recalculated or modified through workflow replay.

No Rewind or Delete operations may be used to manipulate authoritative payroll history.

Do not confuse external at-least-once delivery with database transactional guarantees.

All reported exactly-once behavior must name the precise atomic boundary and the test that proves it.

---

# 8. GATE B STEP 5 — immudb PROVENANCE AND ADVERSARIAL VERIFICATION

Execute the standalone immudb integration using its native Go SDK.

Test the approved event-specific, append-only provenance design.

Required evidence:

- Unique event identities.
- Digest-bound idempotency.
- Duplicate append handling.
- Reuse of an idempotency key with a different payload rejected.
- Outbox-to-ledger recovery.
- Back-link consistency.
- VerifiedTxByID.
- Dual-proof verification.
- Independent trusted-state verification.
- Genesis pinning.
- Auditor behavior after server and client restarts.

Run every deterministic crash/failure injection point.

The highest-risk scenarios are:

1. PostgreSQL commits before the ledger append.
2. Ledger append commits but the client loses the response.
3. Ledger append succeeds but PostgreSQL back-link fails.
4. Duplicate delivery after a process restart.
5. immudb unavailable during append.
6. PostgreSQL unavailable during reconciliation.
7. Idempotency collision with different evidence.
8. Ledger data tampering.
9. Trusted-state tampering.
10. Auditor recovery from a previously pinned checkpoint.

Verify that tampering is detected and reported.

Do not merely inspect the returned transaction ID.

The auditor must independently validate cryptographic evidence using its separately controlled trust state.

Record which tampering operations are detected and which threats are outside the POC's stated threat model.

Verify that no sensitive employee or financial data enters the immutable ledger.

---

# 9. GATE B STEP 6 — COMPLETE THE ONE-MONTH PAYROLL RUN

This is the main demonstration.

Execute the actual full database-backed month using the synthetic payroll fixture.

Do not substitute `--demo-synth-only` for the full run.

The POC must demonstrate:

- Employee population.
- Employment terms.
- Compensation.
- Attendance.
- Timesheets.
- Leave and absence.
- Overtime.
- Allowances.
- Deductions.
- Versioned business rules.
- Effective-dated policy selection.
- Payroll calculation.
- Exception handling.
- Approval.
- Finalization.
- Immutable evidence.
- Independent verification.
- Period closure.
- Historical reproducibility.

Use the synthetic workforce and approved POC rules.

Do not make statutory or compliance claims.

### Required adversarial scenarios

Execute:

- Concurrent payroll requests for the same run.
- Duplicate submission.
- Late attendance arrival.
- Late approval.
- Mid-period salary change.
- Rule-pack change.
- Missing evidence.
- Incomplete temporal coverage.
- Worker crash during calculation.
- Worker crash during approval wait.
- immudb unavailability.
- Post-finalization correction.
- Recalculation from historical pinned inputs.

Unresolved proration cases must produce explicit blockers.

### Required outputs

Generate and inspect:

1. Payroll register.
2. Employee-level payslip calculations.
3. Payroll exception report.
4. Payment-file preview.
5. Period-close report.
6. Immutable provenance report.
7. Independent auditor verification report.
8. Reproducibility report.

Verify actual values, not merely file existence.

A repeated run with identical authoritative inputs, pinned rules and calculator implementation must reproduce identical canonical outputs.

Post-finalization corrections must produce new linked artifacts while leaving the original finalized artifact unchanged.

---

# 10. ADVERSARIAL PERFORMANCE AND FAILURE REVIEW

Execute the existing performance tripwire and record actual measurements.

Measure:

- Capture latency.
- Payroll calculation duration.
- DBOS workflow overhead.
- Queue and worker throughput.
- PostgreSQL lock waits.
- Outbox backlog.
- Ledger append and verification latency.
- Reconciliation throughput.
- Restart/recovery behavior.

Identify whether any significant bottleneck comes from:

- PostgreSQL transaction scope.
- Gate contention.
- Excessive advisory locking.
- DBOS checkpointing.
- Serial ledger appends.
- Canonical serialization.
- Repeated rule resolution.

Do not optimize prematurely.

The purpose is to identify the actual limiting factor at POC scale.

Do not claim production-scale performance from a small synthetic workload.

---

# 11. UPDATE THE EVIDENCE AND DECISION RECORDS

Update:

`poc/payroll-month/POC-REPORT.md`

This is the authoritative record of Gate B.

For every experiment, include:

- Experiment identifier.
- Required environment.
- Exact command.
- Date and time.
- Software versions.
- Image digests.
- Input fixture.
- Expected behavior.
- Actual behavior.
- Exit code.
- Assertions.
- Evidence/log path.
- Result: PASS / FAIL / BLOCKED / NOT RUN.
- Explanation of any deviation.

Distinguish:

**PASS:** Executed and assertions satisfied.

**FAIL:** Executed and an expected invariant was violated or a required assertion failed.

**BLOCKED:** External dependency or prerequisite unavailable.

**NOT RUN:** Not executed.

Never mark an experiment passed because its harness compiles.

Update:

- T1–T13 results.
- E1–E6 results.
- DBOS-1–8 results.
- Crash and tamper matrices.
- Coverage matrix.
- F1–F7 and F1a/F1b outcomes.
- Deletion ledger outcomes.
- Adoption readiness.
- Remaining architectural risks.

Record any discovered implementation defects and their precise remediation.

---

# 12. GATE C PREPARATION — TECHNOLOGY-BY-TECHNOLOGY RECOMMENDATION

Do NOT automatically approve adoption merely because the integrated POC completes.

After Gate B, provide independent recommendations for:

### PostgreSQL 18.6

- ADOPT / DEFER / REJECT.
- Migration and parity evidence.
- Temporal-key benefits.
- Remaining compatibility risks.
- PG16 retirement recommendation.
- Explicit migration rollback.

### DBOS Transact Go

- ADOPT / DEFER / REJECT.
- Atomic capture proof.
- Crash-recovery evidence.
- Workflow replay behavior.
- Operational overhead.
- Whether its benefits justify retaining it.

### immudb

- ADOPT / DEFER / REJECT.
- Cryptographic tamper-verification evidence.
- Dual-write recovery proof.
- Independent auditor evidence.
- Licensing implications.
- Operational overhead.
- Commercial dependency risk.

Include a fourth recommendation for the combined architecture.

Explain whether the three technologies materially improve the actual application over PostgreSQL alone and PostgreSQL + DBOS.

Do not recommend immudb merely because it worked technically. Determine whether the incremental assurance justifies its operational and licensing cost.

All production decisions remain separate and require explicit authorization.

---

# 13. STOP CONDITIONS

Stop and report immediately when any of the following happens:

- Existing uncommitted work is at risk.
- A required human ruling is missing or contradicted.
- Existing PG16 data could be affected.
- The actual DBOS transactional boundary cannot be demonstrated.
- An application role can fabricate authoritative payroll artifacts.
- A finalized payroll artifact can be modified.
- An unresolved business rule is being silently invented.
- A ledger idempotency failure can produce duplicate or missing evidence.
- Independent verification cannot establish the trusted checkpoint.
- A real invariant fails.

Do not bypass a stop condition to get a green report.

---

# 14. FINAL DEFINITION OF SUCCESS

Gate B succeeds only after all mandatory database-backed experiments pass and the full one-month synthetic payroll run has actually completed.

The final evidence must demonstrate that:

1. PostgreSQL 18.6 correctly protects business state.
2. DBOS durably coordinates the payroll lifecycle.
3. Crash recovery does not create duplicate or partial financial artifacts.
4. immudb preserves independently verifiable evidence.
5. Historical calculations can be reproduced.
6. Period closure and retroactive corrections behave as specified.
7. The architecture remains extensible for dynamic business rules.
8. The application does not depend on privileged runtime access.
9. The complete environment can be reproduced by another developer.
10. The project remains within all approved POC boundaries.

**Final instruction:**

Do not repeat Gate A. Do not reopen signed rulings. Do not redesign settled architecture without evidence.

Execute the existing runbook, prove every critical guarantee, document every failure honestly, and deliver the evidence needed for Gate C.

No production migration, no migration 00013, no merge and no commit without explicit authorization.

## What happens after this prompt

There are only three legitimate outcomes:

| Environment/result           | Next action                                                                       |
| ---------------------------- | --------------------------------------------------------------------------------- |
| No Docker-capable host       | Stop and preserve Gate A. Gate B remains blocked.                                 |
| Docker available, tests fail | Fix genuine defects, repeat affected experiments, retain failure history.         |
| Full Gate B passes           | Produce Gate C recommendations for each technology and the combined architecture. |

My strongest recommendation: Do not spend more engineering effort expanding the POC feature set now. You have already invested in the architecture and implementation. The highest-value remaining work is obtaining real PostgreSQL, DBOS and immudb execution evidence.

That is what will tell us whether this architecture is genuinely better than a well-engineered PostgreSQL-only solution.