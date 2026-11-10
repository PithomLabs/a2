I recommend a root-cause remediation assignment, not a list of patches against individual findings.

The fundamental problem is that the three technologies were demonstrated separately, but the actual payroll business operation was never connected end-to-end. The review also uncovered broken authority boundaries, workflow identity problems and an auditor that doesn't verify payroll content.

The implementation agent must fix the underlying architecture, not make the existing tests green. The review identifies these issues with concrete source references, particularly PG-1, RC-1, DB-1, DB-3, IM-1 and IM-7.&#x20;

adv_review(20261110-093829).md



Here is the prompt I would issue.

# ROOT-CAUSE REMEDIATION — Real End-to-End Payroll Execution Through PostgreSQL, DBOS and immudb

## Mission

Act as Principal Software Architect, Senior Go Engineer, PostgreSQL Security Architect and Distributed Systems Engineer.

You are continuing the existing `stack-poc` implementation.

Your assignment is to correct the underlying architectural and implementation defects identified in:

`plans2/POC-ADVERSARIAL-REVIEW.md`

The objective is to produce a credible, executable demonstration in which an actual monthly synthetic payroll operation travels through PostgreSQL, DBOS and immudb, with each technology performing its intended responsibility.

**Do not patch individual symptoms. Correct the underlying design and execution paths that allowed those symptoms to exist.**

The previous POC demonstrated several useful mechanisms, but the adversarial review established that the actual payroll calculation, database capture, workflow execution and ledger evidence were not genuinely connected.

We must correct that fundamental problem.

The objective is NOT to build a full-blown payroll system.

We need one correctly executed monthly payroll, one coherent end-to-end architecture and sufficient adversarial evidence to establish that the three technologies add genuine architectural value.

Once this is achieved, the POC is complete.

---

# 1. Execution Authority and Project Boundaries

You have full delegated engineering authority to investigate, refactor, correct, test and execute everything necessary to achieve the objective.

Do not repeatedly ask for approval for routine engineering decisions.

Exercise senior engineering judgment, investigate root causes, select maintainable solutions and proceed autonomously.

The following constraints remain binding:

- Continue on `stack-poc`.
- Preserve existing uncommitted work.
- Do not commit, push, merge or rebase.
- Do not modify frozen migrations 00001–00012.
- Do not create migration 00013.
- All additional database schema remains POC-scoped.
- Preserve the existing PostgreSQL 16 database and its volume.
- Use disposable PostgreSQL 18.6 and immudb 1.11.2 environments.
- Use synthetic data only.
- Keep SG-MOM-OT-001 BLOCKED.
- Never invent statutory payroll rules or unresolved business semantics.
- Do not weaken Rung 5 database enforcement.
- Respect all signed R1–R7 and P1–P5 decisions.
- Do not authorize production deployment or commercial immudb usage.

Use the actual pinned software versions:

- PostgreSQL 18.6.
- DBOS Transact Go v1.8.0.
- immudb v1.11.2.
- pgx v5.12.0.

You may add POC-specific tables, functions, roles, triggers and SQL overlays where necessary.

However, the POC implementation must represent a credible design that can subsequently be translated into a properly reviewed production migration. It must not depend on test-only shortcuts or privileged application access.

---

# 2. First Principle: One Genuine Business Operation

The most important correction is to stop treating successful infrastructure workflows as successful payroll execution.

The final architecture must perform this actual business operation:

1. PostgreSQL stores and protects authoritative business inputs.
2. DBOS initiates and coordinates the monthly payroll workflow.
3. PostgreSQL authorizes capture of the applicable inputs.
4. Go executes the actual payroll calculations using the captured database records.
5. PostgreSQL stores the actual employee-level calculations and finalized payroll output.
6. DBOS durably coordinates the calculation, approval and finalization stages.
7. A PostgreSQL transactional outbox records the actual finalized payroll evidence.
8. A durable worker appends the evidence to immudb.
9. PostgreSQL records the verified immudb transaction reference.
10. An independent auditor retrieves the authoritative PostgreSQL evidence, reconstructs the expected content and verifies it against immudb.

Every stage must operate on the same identifiable payroll run and its actual contents.

No placeholder operations.

No `SELECT 1` capture in place of business capture.

No workflow returning phase-name strings as proof of payroll completion.

No dummy digests.

No test-seeded ledger rows masquerading as payroll provenance.

No DB-free canonical calculation presented as the result of the DB-backed workflow.

The DB-backed monthly payroll result must be independently reproducible from the captured inputs and pinned calculation implementation.

---

# 3. PostgreSQL: Correct the Authoritative Business-Data Boundary

This is the first priority because the database currently permits an application-role writer to fabricate a valid-looking snapshot and promote it to a validated payroll run.

That is a fundamental architectural defect.

## 3.1 Snapshot authority

Inspect the existing snapshot tables, grants, triggers, functions and lifecycle transitions.

Implement a proper database-enforced authority boundary.

The desired architecture:

- The application runtime must not directly INSERT, UPDATE or DELETE authoritative snapshot records.
- Snapshot creation must occur through a narrowly scoped authorized database operation.
- Validation must not accept attacker-supplied evidence merely because parent identifiers exist.
- Finalization must not depend on caller-selected lifecycle labels.
- All authoritative transitions must verify the actual underlying records.

Use an appropriately owned `SECURITY DEFINER` function or equivalent database-controlled operation.

Do not simply add a trigger that rejects the particular SQL statement used in the adversarial test.

The secure operation must derive or verify its data from authoritative business tables.

It must not trust caller-supplied:

- Employee membership.
- Input classifications.
- Rule-pack pins.
- Snapshot digests.
- Validation status.
- Claims or source references.

The database operation must verify that the requested snapshot corresponds to an eligible and complete set of approved business records.

Use a restricted owner role, secure `search_path`, schema-qualified references and explicit execution privileges.

Do not allow the application runtime to invoke unrestricted administrative operations.

## 3.2 Fabrication-to-validation attack

Create a genuine regression test covering the entire exploit chain:

1. Create legitimate parent business records.
2. Use the real `app` role.
3. Attempt to insert fabricated snapshot headers.
4. Attempt to insert fabricated snapshot members.
5. Attempt to insert fabricated rule pins.
6. Attempt to manufacture validation evidence.
7. Attempt direct promotion to `validated`.

Every unauthorized operation must fail.

Also verify that the legitimate capture operation succeeds.

An architecture that blocks legitimate processing is not acceptable.

The test must verify persisted data and actual state transitions, not merely error messages.

## 3.3 Employer gate and writer participation

The existing implementation proves an advisory-lock primitive, but the actual capture path does not take the gate.

Correct that.

Every relevant business writer must participate in the same authoritative synchronization protocol.

Review:

- Timesheet creation and approval.
- Employment-term changes and closure.
- Compensation changes.
- Pack activation.
- Leave and attendance disposition.
- Period and employer changes.
- Payroll snapshot capture.
- Payroll validation and finalization.

Include DELETE operations where the business model permits them.

Cross-employer updates must acquire the necessary employer gates in deterministic sorted order.

The gate must operate inside the same database transaction as the associated business change.

Execute both writer-first and capture-first interleavings.

Do not claim synchronization merely because `poc_gate_acquire` works in isolation.

## 3.4 Period closure

Correct the actual close operation.

The current close path is both unusable through its intended application route and bypassable by direct SQL.

Resolve:

- Function privileges.
- Authorized close creation.
- Database-controlled close timestamp.
- Duplicate close handling.
- Close-record fabrication.
- Closed-period writer rejection.
- Transactions racing closure.

The existing writer trigger must derive the affected pay period using actual schema relationships, such as `pay_period_id`, rather than nonexistent generic `period` fields.

Test all participating tables.

## 3.5 Claims and coverage verification

Do not retain a verifier that always fails with SQLSTATE 42703.

Redesign its data model and implementation around actual source-table relationships.

A valid verifier must:

- Identify all expected source records.
- Detect missing records.
- Detect unexpected claims.
- Detect duplicate membership.
- Detect invalid employee-to-source relationships.
- Detect uncovered periods.
- Reject empty or vacuous evidence.
- Execute within the authoritative capture transaction.

Ensure claims are actually written through an authorized mechanism.

A zero-row claims table must never result in a successful validation merely because zero missing claims were counted.

Do not assume PostgreSQL 18.6 provides PERIOD foreign-key functionality that the actual test environment rejected.

Use supported range constraints and correctly synchronized procedural verification where necessary.

## 3.6 Outbox and backlink authority

Correct the current grant-only design.

The `app` role must not be able to:

- Forge completed outbox records.
- Mark pending records failed to suppress evidence.
- Rewrite pending evidence payloads.
- Substitute content digests.
- Insert forged back-links.
- Squat on legitimate event identities.
- Delete outstanding evidence events.

Use narrowly scoped authorized enqueue and transition functions.

The enqueue operation must derive its evidence payload from actual persisted business artifacts, not accept arbitrary caller-supplied evidence.

Separate application and worker responsibilities where appropriate.

Use an explicit state machine with valid transitions, recovery handling and auditability.

The PostgreSQL back-link must identify the actual immutable event, including its event identity. Multiple legitimate corrections must not be silently dropped by a primary key that permits only one event per business artifact.

A conflicting back-link must fail loudly unless it is an exact verified replay of the same event.

---

# 4. DBOS: Replace Workflow Stubs With Actual Payroll Processing

This is the second priority.

The current shared-pool atomic mechanism is a legitimate architectural capability, but its actual transaction body contains only a placeholder.

The real payroll operation must execute through the intended durable workflow architecture.

## 4.1 Correct workflow structure

Implement a coherent monthly payroll workflow with meaningful durable boundaries.

A suitable design is:

**Stage A — Request and readiness**

- Establish employer and pay-period identity.
- Resolve the authoritative input version.
- Verify the period is eligible for processing.
- Reject ambiguous or incomplete business inputs.
- Establish a stable input digest.

**Stage B — Authorized capture**

Execute the actual capture inside `RunAsTransaction`.

Within this transaction:

- Acquire the employer gate.
- Verify all participating records.
- Resolve the approved employee population.
- Resolve effective-dated compensation and rule packs.
- Validate coverage.
- Create the authoritative payroll snapshot.
- Persist input and rule provenance.
- Create the corresponding application-owned operation record.
- Commit the DBOS durability record atomically.

The snapshot and DBOS checkpoint must belong to the same actual transaction.

Do not obtain a separate pool connection for business writes.

**Stage C — Actual calculation**

Retrieve the captured snapshot from PostgreSQL.

Do not call the fixture builder or regenerate the original workload to calculate payroll.

Use the actual persisted snapshot as the source of calculation.

Invoke the existing `synth.ComputePay` or its appropriate existing calculation interface.

Do not duplicate or rewrite the calculation logic.

Use exact decimal/integer arithmetic and established rounding behavior.

Calculate employee-level results and aggregate the monthly payroll.

Record typed blockers instead of silently processing invalid cases.

**Stage D — Persist calculation results**

Persist actual employee-level outputs in PostgreSQL.

At minimum:

- Employee payroll lines.
- Gross earnings.
- Overtime component.
- Allowances.
- Deductions.
- Net pay.
- Applied rule-pack identities and digests.
- Calculation version.
- Input snapshot digest.
- Output digest.

Use `RunAsTransaction` for the authoritative result persistence and corresponding DBOS durability record.

Use appropriate unique constraints and business-level idempotency.

A successful workflow status must never be accepted as evidence that payroll output exists.

**Stage E — Approval**

Implement an actual durable approval wait using the supported DBOS v1.8.0 mechanisms.

Use a simple authorized POC command or equivalent interface for submitting an approval decision.

The approval event must be verified against PostgreSQL business authority before it can authorize finalization.

Test application restart while the workflow is waiting.

Do not build an approval UI or a complete authorization subsystem.

**Stage F — Finalization**

Finalize through a protected database operation.

Verify:

- Actual calculation rows exist.
- The input snapshot digest matches.
- The calculator version is recorded.
- All required approvals exist.
- No blockers remain.
- The output digest has been recomputed.
- The payroll period is eligible for closure.

The original finalized artifact must never be overwritten.

Corrections must create explicitly linked successor or adjustment artifacts.

**Stage G — Evidence publication**

Create the transactional outbox event as part of the same authoritative PostgreSQL transaction that records finalization and its DBOS checkpoint, using the appropriate DBOS transaction boundary.

The resulting event must represent the actual finalized payroll content.

## 4.2 Workflow identity and changed-input handling

Correct the silent stale-result and sticky-ERROR defects.

Distinguish three identities:

1. Business payroll-run identity.
2. Immutable payroll-input version or digest.
3. Individual workflow execution attempt.

Define explicit behavior:

- Identical request and identical input digest: safe idempotent replay.
- Same employer and period but changed input digest: reject with an explicit input-version conflict, or require an explicitly authorized new payroll revision.
- Transient workflow failure: a new execution attempt must be possible without corrupting or duplicating the business operation.
- Finalized payroll: never silently replace it with a rerun.
- Concurrent submissions: only one authoritative final result may be produced for the same approved business version.

Inspect DBOS v1.8.0's actual workflow-ID reuse policies.

Do not manually modify DBOS system tables to recover failed workflows.

Prefer separate attempt identities linked to one authoritative business run, with application-owned state and database-enforced uniqueness.

## 4.3 DBOS system-state trust

Respect the signed decision that DBOS system tables are not authoritative business state.

Implement and test the actual forged-SUCCESS pre-registration scenario:

1. Pre-register a forged successful workflow result under the exact intended workflow identity.
2. Submit the legitimate payroll operation.
3. Verify that no attacker-controlled result can be returned or finalized as legitimate payroll.
4. Verify that PostgreSQL business records remain authoritative.
5. Verify that the application detects a missing or inconsistent business result and fails closed.

A DBOS `workflow_status` row must never independently authorize a financial artifact.

## 4.4 Exactly-once discipline

Use precise terminology.

Demonstrate separately:

- Atomic capture transaction.
- Atomic result persistence.
- Workflow completion deduplication.
- External side-effect idempotency.
- Outbox delivery.
- Financial artifact uniqueness.

Prove business effects using actual persisted payroll artifacts and unique keys.

Counting `workflow_status` SUCCESS rows is not sufficient.

Do not place a non-idempotent external immudb operation inside an ordinary `RunAsStep` and then claim exactly-once external delivery.

Preserve the external outbox/idempotency boundary.

---

# 5. immudb: Establish Real Content-Bound Independent Evidence

This is the third priority.

The current implementation proves some verified-append and idempotency mechanisms, but the auditor does not verify the actual relationship between PostgreSQL payroll records and immudb evidence.

Correct the evidence architecture itself.

## 5.1 Actual evidence payload

The ledger must anchor real records derived from PostgreSQL, not hardcoded dummy digests.

Use a canonical evidence record containing:

- Employer and payroll artifact references.
- Payroll run identity.
- Event UUID.
- Input snapshot digest.
- Rule-pack identity and content digest.
- Calculator implementation/build digest.
- Employee population digest where appropriate.
- Finalized payroll output digest.
- Approval and finalization references.
- Previous event/supersedes reference.
- Relevant PostgreSQL record references.

Do not store PII, NRICs, bank details, individual employee names, unnecessary identifiers or secrets.

Verify the actual bytes written to immudb.

Do not rely exclusively on checking the Go struct definition.

## 5.2 Independent content verification

This is a critical requirement.

Redesign the auditor so that it does more than verify that a transaction ID exists.

The auditor must:

1. Connect to PostgreSQL using read-only credentials.
2. Retrieve the actual finalized payroll artifact and its authoritative references.
3. Reconstruct the canonical evidence record.
4. Recompute the expected content digest.
5. Retrieve the corresponding immudb evidence through cryptographically verified operations.
6. Verify the event identity and artifact identity.
7. Verify the correct PostgreSQL record references.
8. Verify the rule-pack digest.
9. Verify the calculator digest.
10. Verify the payroll result digest.
11. Compare all reconstructed values.
12. Fail loudly on any mismatch.

The auditor must not trust a supplied backlinks file as the sole source of truth.

Read the relevant PostgreSQL records and establish their relationship to the immutable evidence.

Test:

- Modified payroll amounts.
- Modified employee result rows.
- Changed snapshot content.
- Changed rule-pack digest.
- Changed calculator digest.
- Swapped transaction IDs between artifacts.
- Forged PostgreSQL back-links.
- Missing events.
- Superseded and corrected artifacts.

The auditor must detect mismatches, not just nonexistent transaction IDs.

## 5.3 Trusted genesis and auditor identity

Complete the approved R2 design.

Implement:

- An externally retained genesis checkpoint.
- Pinned server UUID.
- Pinned state hash.
- Mandatory verification against the trusted checkpoint.
- Refusal to silently accept a new server identity.
- One-time controlled initialization.
- Subsequent verification that cannot silently bootstrap a new trust anchor.

`--accept-genesis` must not be the default behavior for ordinary audit runs.

The genesis commitment must be retained outside the writer's mutable control and documented as an executed operator-controlled step.

Use distinct immudb identities for:

- Writer.
- Independent auditor.

The auditor must have read-only permissions.

Do not use the same immudb superuser credentials for both.

The POC environment may use localhost-only transport where expressly permitted, but clearly distinguish that experimental configuration from a production security design.

## 5.4 Verified recovery

Replace unverified marker `Get` operations during recovery with cryptographically verified retrieval, such as the SDK's appropriate `VerifiedGet` mechanism.

After a precondition collision:

- Recover the actual original event.
- Verify the returned content.
- Verify the event UUID.
- Verify the content digest.
- Verify the originating transaction.
- Verify that the returned event belongs to the intended payroll artifact.

Never create a PostgreSQL back-link based merely on an unverified server response.

## 5.5 Back-link and correction semantics

Correct the backlink identity model.

A payroll artifact may legitimately have successive evidence events.

The database key must represent event identity, not silently impose one immutable ledger transaction per PostgreSQL row forever.

Use an appropriate composite identity such as:

`(pg_table, pg_id, event_uuid)`

or a similarly sound design.

A conflicting event must fail loudly rather than being discarded through `ON CONFLICT DO NOTHING`.

Verify that:

- Original finalization remains anchored.
- Corrections append new evidence events.
- Supersedes lineage remains intact.
- The old event remains independently verifiable.
- The new event binds to the corrected artifact.

## 5.6 Outbox recovery and poison handling

Correct the underlying state machine.

A ledger append that succeeded but whose PostgreSQL back-link failed must not become permanently orphaned merely because retry accounting reached a threshold.

Design a recovery/reconciliation path that can determine whether an event exists, verify it and repair the legitimate back-link safely.

Distinguish:

- Transient delivery attempts.
- Confirmed ledger append.
- Verification failure.
- Back-link failure.
- Permanent poison events.
- Successfully reconciled events.

Make duplicate convergence an ordinary successful recovery outcome, not a misleading worker failure.

---

# 6. End-to-End Acceptance Tests

Do not add dozens of unrelated tests.

Build a focused acceptance battery covering the actual connected business operation.

## PostgreSQL authority

- Valid-parent fabricated snapshot rejected.
- Fabricated snapshot cannot become validated.
- Authorized snapshot creation succeeds.
- Actual capture participates in the gate.
- Writer-first and capture-first concurrency are tested.
- Closed-period writes are rejected through actual schema relationships.
- Empty claims/evidence cannot validate.
- Direct outbox and backlink fabrication is rejected.

## DBOS execution

- Actual capture rows and DBOS checkpoint commit atomically.
- Actual calculation results persist.
- Duplicate identical request returns the same verified result.
- Changed inputs cannot silently return stale output.
- ERROR workflows can be safely retried through a documented attempt model.
- Forged SUCCESS under the intended workflow ID cannot authorize payroll.
- Crash recovery does not duplicate financial artifacts.
- Approval waiting survives process restart.
- Actual finalization is protected and irreversible.

## immudb evidence

- Real payroll digest is anchored.
- Actual rule and calculator digests are bound.
- Auditor reads PostgreSQL and immudb independently.
- Modified PostgreSQL payroll evidence fails verification.
- Transaction substitution fails.
- Wrong genesis identity or state hash fails.
- Writer cannot forge auditor state.
- Backlink-missing recovery succeeds.
- Duplicate delivery converges safely.
- Corrections append new events without overwriting historical evidence.

Every critical test must verify persisted business results, not just workflow status or successful API calls.

Run attack tests using the intended application and worker roles.

Do not use superuser for ordinary adversarial probes.

---

# 7. Shared Monthly Payroll Demonstration

Reuse the original October 2026 synthetic workload:

- One employer.
- Twenty-five employees.
- Thirty-one days.
- Existing employment, attendance and timesheet fixtures.
- Existing compensation and synthetic rule packs.
- Existing leave, overtime, allowances and deductions.
- Existing pure calculation implementation.

Do not duplicate the calculator.

The actual database-backed workflow must:

1. Read authoritative records from PostgreSQL.
2. Capture them into an immutable payroll snapshot.
3. Reconstruct the calculation inputs from that persisted snapshot.
4. Execute the existing calculator.
5. Persist actual employee-level payroll outputs.
6. Produce the monthly register.
7. Produce the payslip calculations.
8. Record exceptions.
9. Obtain and validate the POC approval.
10. Finalize and close the payroll period.
11. Enqueue actual provenance evidence.
12. Anchor the evidence in immudb.
13. Independently recompute and verify the result.

The canonical payroll digest and database-backed result must be demonstrably identical when derived from the same immutable input snapshot and calculator version.

The final demonstration must distinguish:

- Finalized employees.
- Blocked employees.
- Exceptions.
- Total net pay.
- Original payroll result.
- Any subsequent correction.

Do not calculate unresolved proration or statutory rules.

---

# 8. Architecture and Implementation Quality

Do not solve these issues through test-specific SQL or hardcoded exceptions.

The corrected architecture must remain coherent.

Requirements:

- Business rules remain in the existing Go domain/calculation layer.
- PostgreSQL enforces authoritative data invariants.
- DBOS orchestrates durable processes.
- immudb records verifiable evidence.
- The outbox is the explicit boundary between PostgreSQL and immudb.
- The auditor independently binds actual PostgreSQL content to ledger evidence.
- Runtime roles have the minimum necessary permissions.
- Administrative functions are isolated.
- Transaction ownership is explicit.
- Exact financial arithmetic is preserved.
- Workflow identity and business-run identity are distinct.
- Corrections preserve historical lineage.
- All evidence is reproducible.

Avoid unnecessary new packages or infrastructure abstractions.

Use existing code where sound.

Where a fundamental design is incorrect, replace it properly rather than adding another layer around it.

---

# 9. Execution and verification strategy

Proceed in dependency order:

**Phase A — Establish the truth**

Reproduce the critical findings on disposable environments.

Record the initial exploit behavior and actual placeholder execution paths.

**Phase B — Correct PostgreSQL authority**

Fix snapshot fabrication, transition authority, gate participation, close semantics, coverage verification and outbox/back-link permissions.

Run the relevant tests before proceeding.

**Phase C — Connect DBOS to actual payroll**

Replace placeholder capture and phase functions with the real persisted-input and calculation paths.

Correct run identity, input-drift behavior, retry semantics and business-effect assertions.

**Phase D — Connect real payroll to immudb**

Wire finalized payroll evidence into the outbox.

Implement content-bound append, verified recovery, independent auditor identity and genesis verification.

**Phase E — Execute the complete integration**

Execute one complete monthly payroll using the shared fixture.

Inject the critical failures.

Verify every persisted result and digest independently.

**Phase F — Reassess the architecture**

Determine whether PostgreSQL, DBOS and immudb now provide the intended distinct capabilities.

Do not automatically retain any technology whose value cannot be demonstrated.

---

# 10. Reporting and Evidence

Produce:

`plans2/POC-REMEDIATION-REPORT.md`

Update the existing:

- `plans2/POC-FINAL-REPORT.md`
- `plans2/POC-RESULTS.md`
- `poc/payroll-month/POC-REPORT.md`

Do not erase historical failed results.

Explicitly supersede earlier claims with new evidence.

The remediation report must contain:

1. Original root causes.
2. Corrected architecture.
3. PostgreSQL authority changes.
4. DBOS workflow and transaction changes.
5. immudb evidence and auditor changes.
6. Exact file and SQL changes.
7. Before-and-after attack results.
8. Actual monthly payroll output.
9. Digest and reproducibility evidence.
10. Remaining limitations.
11. Revised technology recommendations.

Every result must be classified:

- IMPLEMENTED.
- BUILT.
- EXECUTED.
- PASSED.
- FAILED.
- BLOCKED.
- NOT RUN.

Do not claim success based on compilation.

---

# 11. Definition of Completion

The remediation is complete only when all of the following are demonstrated:

1. The application role cannot fabricate and validate an unauthorized snapshot.
2. The real capture path participates in PostgreSQL's authoritative gate.
3. Actual payroll inputs are captured inside the DBOS transactional boundary.
4. Actual payroll calculations execute against persisted captured inputs.
5. Actual employee payroll outputs persist in PostgreSQL.
6. DBOS recovers the real business operation without duplicates or stale-result acceptance.
7. Changed inputs require explicit version or revision handling.
8. Finalization and approval are protected.
9. The actual finalized payroll evidence enters the transactional outbox.
10. immudb anchors the actual snapshot and calculation provenance.
11. The independent auditor recomputes and verifies actual payroll content.
12. Historical alterations are detected within the stated threat model.
13. Original finalized artifacts remain immutable through corrections.
14. The final one-month payroll can be reproduced from its pinned evidence.

If any of these conditions fails, document the exact result and do not declare the remediation complete.

---

# 12. Scope Discipline

This is a final correction of the existing POC, not a new product-development phase.

Do not:

- Build a complete payroll system.
- Add new HR modules.
- Implement statutory Singapore overtime.
- Reopen settled business decisions.
- Add a new workflow engine.
- Add a separate rules engine.
- Add an alternative ledger.
- Build three alternative payroll applications.
- Run exhaustive T1–T13 matrices unless directly necessary to verify a corrected root cause.
- Create production migration 00013.
- Change the frozen core migrations.
- Add features merely for presentation.

The conventional-job baseline and additional comparative work should remain explicitly classified according to the evidence actually available. Do not pretend they were measured.

The immediate objective is to fix the end-to-end architecture and demonstrate real business execution.

## Final directive

We do not need more architectural diagrams showing PostgreSQL, DBOS and immudb connected by arrows.

We need the actual code paths connected.

PostgreSQL must authorize the real capture.

DBOS must durably execute the real payroll process.

Go must calculate the real payroll from the captured database records.

PostgreSQL must persist the actual financial results.

immudb must anchor those actual results.

The independent auditor must independently reconstruct and verify those results.

**Fix the underlying authority, execution, transaction and evidence boundaries. Do not patch individual tests. Do not make placeholders appear successful.**

The final deliverable is one credible monthly payroll POC proving that the three technologies work together in their respective architectural layers.

When that proof exists, stop. The POC is finished.

## My emphasis: three things the agent must not compromise

1\. Fix the source of authority, not merely the symptoms.

For example, adding another trigger to reject the particular SQL statement used in PG-1 is insufficient. The application must not possess direct authority to manufacture an authoritative payroll snapshot in the first place.

2\. Make DBOS transact on real business work.

The most important DBOS correction is replacing the empty transaction body with genuine capture and persistence operations. Otherwise, we are proving a library feature, not an application architecture.

3\. Make immudb verify actual content.

The auditor must reconstruct the payroll evidence from PostgreSQL and compare it against cryptographically verified ledger content. Verifying that transaction IDs exist is insufficient.

My recommendation: Give the agent this assignment with the existing full autonomy authorization. Do not let it spend time fixing all 46 findings indiscriminately. Prioritize the root causes that prevent the actual payroll from becoming authoritative, durable and independently verifiable. That is the credible POC we set out to build.