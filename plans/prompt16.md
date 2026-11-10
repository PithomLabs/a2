Milestone 2 deserves an independent adversarial code review before we implement any payroll calculations. The implementation report declares PASS, but that is the implementation agent's result, not yet independent acceptance.

I see four areas that deserve especially rigorous scrutiny:

1. Whether the `REPEATABLE READ` snapshot and advisory-lock ordering genuinely guarantee one consistent population and input view.
2. Whether direct SQL can bypass payroll-run validation, audit or snapshot governance.
3. Whether SHG and rule-pack activation preserve their evidence and effective-version semantics.
4. Whether compensation resolution and snapshot hashing produce reproducible financial inputs.

There is also an explicitly acknowledged limitation: `payroll_line` remains mutable after validation, and direct SQL can perform an outcome-equivalent `draft → validated` transition without the service's pin-stamping and audit. These must be assessed before the first calculation.&#x20;

milestone2_remediation_report.md



## Prompt 16 — Independent adversarial review

# Prompt 16 — WorkforceOps Milestone 2 Final Adversarial Code Review

# Payroll Governance, Snapshot Integrity and Reproducibility

## 1. Objective

Conduct an independent, adversarial technical review of the completed Milestone 2 implementation.

Determine whether the six prerequisites are genuinely implemented and whether WorkforceOps can safely proceed to its first independently verified Singapore payroll calculation.

The implementation report is:

`plans/milestone2_remediation_report.md`

The implementation claims:

- P3 effective compensation resolver complete.
- P4 SHG instruction lifecycle complete.
- P5 controlled rule-pack activation complete.
- P2 immutable payroll snapshots complete.
- P1 draft-to-validated payroll-run lifecycle complete.
- P6 verification complete.

The implementation agent reports PASS, including 19 Go packages under three timezone configurations, real PostgreSQL tests, race tests and clean-server provisioning.

Do not accept that verdict without independently examining the code, migrations, database behavior, concurrency and persisted state.

**This is a read-only adversarial review. Do not modify application source, migrations, tests, API contracts, configuration or documentation.**

The purpose is to identify whether the system has any remaining structural or business-integrity defect that could invalidate the first financial payroll calculation.

Do not redesign the architecture or repeat the previous 72-finding review.

---

# 2. Establish the actual review baseline

The Milestone 2 report records:

- Initial HEAD: `304fec8`, release 0.1.3.
- Clean tree at Phase 0.
- Migrations `00001`–`00011` verified byte-identical.
- One new migration: `00012_m2_governance.sql`.
- All Milestone 2 implementation changes remain uncommitted.

Before reviewing:

1. Inspect `git status --short`.
2. Record `git rev-parse HEAD`.
3. Confirm the initial Milestone 2 baseline.
4. Inventory tracked, untracked, modified, deleted and renamed files.
5. Identify pre-existing user files, particularly `docs/writeup3.md`, and exclude them from review-scope assumptions.
6. Verify `00001`–`00011` against their recorded baseline checksums.
7. Inspect migration `00012` completely.
8. Read the full Milestone 2 remediation report.

Do not reset, clean, stage, commit, reformat or modify anything.

Preserve all existing uncommitted work.

All additional database testing must use explicitly disposable PostgreSQL 16 databases.

The only permitted new repository file is:

`plans/milestone2_adversarial_review.md`

Temporary SQL probes and test scripts may be created outside the repository.

---

# 3. Review Area A — Compensation resolver and historical correctness

Inspect:

- `internal/compbasis/resolve.go`
- Compensation models and repository queries.
- Compensation creation and correction functions.
- `internal/payroll/readiness.go`
- Snapshot capture.
- Relevant migrations and tests.

The resolver was moved into `internal/compbasis` to avoid an HR/payroll import cycle.

Verify that this package has one authoritative implementation and that readiness and snapshots use the same resolver.

## A1. Effective compensation semantics

For each employer, employee and compensation classification, verify that the resolver correctly handles effective-date segments.

Test:

- Original compensation without successor.
- One successor.
- Multiple sequential successors.
- Superseded predecessors.
- Non-overlapping historical compensation.
- Future-effective compensation.
- Mid-period compensation changes.
- Broken chains.
- Competing live rows.
- Overlapping compensation across employment-term records.
- Invalid cross-employee or cross-employer successors.

The resolver must not:

- Sum original and successor records.
- Select arbitrary records with `LIMIT 1`.
- Treat legitimate non-overlapping historical changes as chain corruption.
- Ignore an effective compensation segment.
- Return a value outside the applicable employment terms.

For the first supported monthly payroll profile, a mid-period change may remain a blocker, but it must be an explicit and deterministic blocker.

## A2. Snapshot integration

Prove that the snapshot pins exactly the records returned by the resolver.

For each classification, inspect:

- Source compensation ID.
- Effective start/end dates.
- Exact monetary value.
- Employment-term ID.
- Successor lineage.
- Employee basis sequence.

The snapshot must not contain both a superseded predecessor and its effective successor as payable inputs.

## A3. Acceptance

The resolver must produce one deterministic, explainable compensation basis.

Ambiguity must block readiness and snapshot validation.

No statutory calculation is involved in this test.

---

# 4. Review Area B — SHG lifecycle and evidence correctness

Inspect:

- `internal/hr/shg.go`
- `internal/hr/shg_lifecycle_test.go`
- SHG schema and functions in migration `00012`.
- `internal/payroll/readiness.go`
- `checkSHGCoverage`.
- `checkSHGConflict`.
- Relevant direct-SQL integrity tests.

## B1. Status semantics

Verify that these states remain distinct:

- Missing.
- Pending.
- Active.
- Verified-none.
- Opt-out.
- Superseded.
- Corrupt/inconsistent.

Confirm that:

- Missing never means verified-none.
- Pending blocks readiness.
- Active without verification blocks readiness.
- Verified-none requires the evidence triple.
- Opt-out blocks until its meaning is independently established.
- Superseded instructions remain historical and do not become effective again.
- Invalid successor chains block readiness.

Inspect whether `verified_at`, `verified_by` and `verified_basis` are actually enforced together.

## B2. Effective-period coverage

Test:

- One instruction covering the full payroll period.
- Instruction beginning mid-period.
- Instruction ending mid-period.
- Successor effective date.
- Historical instruction.
- Overlapping same-fund instructions.
- Cross-fund combinations.
- Gaps in instruction coverage.
- Conflicting historical instructions.

The system must not treat a partial-period instruction as covering the entire period.

The first supported payroll profile must either have fully evidenced instruction coverage or return a precise blocker.

Do not invent cross-fund SHG semantics.

## B3. Correction and approval interaction

Test a correction after timesheet approval and after snapshot creation.

Prove that:

- Original instructions remain immutable.
- A successor preserves history.
- The employee payroll-input lock and `basis_seq` are used.
- Readiness detects the relevant change.
- A prior snapshot is not rewritten.
- Revalidation produces a new snapshot version where required.

## B4. Direct SQL

As `app`, attempt:

- Invalid verified-none insertion.
- UPDATE/DELETE of existing instructions.
- Cross-employee succession.
- Cross-employer succession.
- Multiple successors.
- Overlapping effective instructions.
- Backdated insertion affecting an approved period.

Verify persisted state and audit atomicity after each operation.

---

# 5. Review Area C — Rule-pack activation governance

Inspect:

- `internal/rules/`
- `internal/rules/*_test.go`
- `cmd/workforceops/admin.go`
- `cmd/workforceops/admin_test.go`
- `rules/*/manifest.json`
- Rule-pack registry schema.
- Activation history.
- `activate_rule_pack()`.
- Parameter-table privileges.
- Migration `00012`.

## C1. On-disk evidence verification

The administrative CLI reads rule-pack files and recomputes their digests.

Verify that it checks the actual files, not just caller-supplied metadata.

Inspect:

- Manifest identity.
- Exact version.
- Jurisdiction.
- Scope.
- Effective dates.
- Source references.
- Formal specification.
- Independent review record.
- Required fixtures.
- Verification status.
- Artifact digests.

Check path traversal, symlink handling, missing files, invalid manifest references and digest mismatch.

Determine whether a file can be changed between validation and the administrative database operation in a way that activates an unverified package.

## C2. Database activation function

Verify:

- Exact registry identity.
- Exact reviewed version.
- Permitted transition.
- Activation effective period.
- Evidence-reference consistency.
- Supersession history.
- Single effective activation where required.
- Concurrency serialization.
- Atomic audit.

The database must not accept an arbitrary version simply because an administrator passes a version string and digest.

The administrative trust boundary must remain separate from normal application operations.

## C3. Application-role privileges

As `app`, test direct INSERT/UPDATE/DELETE on:

- Registry.
- Activation history.
- Parameter tables.
- Any related effective-version records.

All unauthorized mutations must fail without partial changes.

## C4. Actual rule-pack status

Verify that:

- All ten registry identities remain correctly represented.
- No unsupported pack has become IMPLEMENTABLE.
- No pack is activated without verified evidence.
- `SG-MOM-OT-001` remains BLOCKED in the manifest, registry and activation history.

The current tests use temporary fixture packs. Do not mistake a successful fixture activation for real statutory-package approval.

## C5. Acceptance

The mechanism must be secure and auditable, but must not falsely imply that an actual statutory calculation package is ready.

---

# 6. Review Area D — Payroll snapshot isolation and population consistency

This is the highest-priority review area.

Inspect:

- Snapshot capture implementation.
- `BeginSnapshotTx`.
- `REPEATABLE READ` usage.
- Employer population lock helpers.
- Employee pay-basis lock helpers.
- Rule-pack lock helpers.
- All population-affecting writers.
- All payroll-input writers.
- Snapshot tables and triggers in `00012`.
- `snapshot_race_test.go`
- `snapshot_governance_test.go`

The report states that snapshot capture uses REPEATABLE READ because advisory locks alone do not establish a consistent database view.

That is correct as a design objective, but it must be verified against actual PostgreSQL snapshot timing.

## D1. Critical snapshot-acquisition ordering test

Investigate this subtle possibility:

A transaction at REPEATABLE READ can acquire its MVCC snapshot when it executes its first query. If that first query waits for an advisory lock, the transaction might retain a snapshot that predates the writer's commit even though the lock is acquired afterwards.

Do not assume the lock acquisition and the snapshot acquisition occur in the correct order.

Test explicitly:

1. Transaction A begins a population-affecting write.
2. It acquires the employer population lock.
3. It inserts a new employee or changes an employment record.
4. It pauses before commit.
5. Transaction B starts snapshot capture at REPEATABLE READ.
6. Transaction B attempts to acquire the same employer population lock and waits.
7. Transaction A commits.
8. Transaction B acquires the lock.
9. Transaction B captures the employee population.
10. Verify that Transaction B sees the committed employee/change.

Repeat the same experiment for:

- Employee compensation lock.
- Rule-pack activation lock.
- Timesheet approval lock.

If the snapshot misses a writer that committed while the lock request was waiting, the isolation protocol is defective.

Establish the actual PostgreSQL behavior and the exact point at which the MVCC snapshot is fixed.

If necessary, consider acquiring a session-level advisory lock before starting the REPEATABLE READ transaction, or an equivalent carefully justified protocol.

Do not modify code during this review.

## D2. Population barrier coverage

Create a source-to-lock matrix for every operation affecting the expected population:

- Employee creation.
- Employment-term creation and changes.
- Employee status transitions.
- Termination.
- Any other supported employee inclusion change.

Verify both Go-service and direct-SQL paths.

A direct SQL mutation must either:

- Acquire the employer population barrier through a trigger, or
- Be prevented by privileges and exposed only through a controlled function that participates in the barrier.

Do not rely solely on Go services.

## D3. Employee lock coverage

Verify the shared per-employee payroll-input lock is consistently used by all operations that can change a snapshotted employee's financial or operational basis.

Inspect:

- Compensation add/correct.
- Employment-term changes.
- Employee status changes.
- Timesheet approval.
- Attendance correction.
- Late break attachment.
- Relevant SHG changes.
- Any other supported source mutation.

The lock identity must be identical across Go, triggers and functions.

Acquire multiple employee locks in deterministic UUID order.

## D4. Rule-pack lock coverage

Verify:

- Snapshot captures exact activation records.
- Administrative activation uses the same pack synchronization convention.
- A concurrent pack supersession cannot produce a mixed snapshot.
- Snapshot records cannot combine an older manifest with a newer activation.
- A blocked pack cannot become silently accepted.

## D5. Repeatable-read guarantee

Inspect all queries within snapshot capture.

Verify that:

- The snapshot uses one transaction.
- It does not accidentally escape to another database connection.
- No helper function runs outside the transaction.
- No query reads from a different isolation level or connection.
- Errors roll back the entire capture.
- Serialization errors are handled as specified.

The report says reader transactions cannot hit 40001 and no retry loop is required. Verify this claim from the actual access pattern and tests.

## D6. Mixed-version snapshot test

Use separate real PostgreSQL connections to race:

- Employee creation.
- Employment-term change.
- Employee termination.
- Compensation insertion.
- Timesheet approval.
- Attendance correction.
- SHG instruction correction.
- Rule-pack supersession.

Each mutation must either:

- Commit before snapshot capture and be included consistently, or
- Commit after snapshot capture and be detectable as drift.

A snapshot must never combine values from incompatible transaction points.

---

# 7. Review Area E — Snapshot content, hash and lifecycle

Inspect:

- `payroll_snapshot`.
- `payroll_snapshot_employee`.
- `payroll_snapshot_pack`.
- Snapshot lifecycle events.
- Freeze triggers.
- Invalidation functions.
- Canonical JSON and SHA-256 implementation.
- `InspectSnapshot`.
- Revalidation logic.

## E1. Snapshot completeness

Verify the snapshot includes:

Employer and run:

- Employer.
- Jurisdiction.
- Payroll period.
- Currency.
- Run type.
- Snapshot schema version.

Employee:

- Expected population membership.
- Effective employment-term identity.
- Relevant employee-status identity.
- Employee basis sequence.
- Inclusion/exclusion decision.
- Blockers.

Compensation:

- Exact effective compensation IDs.
- Classification.
- Effective intervals.
- Exact amount strings.
- Employment-term references.
- Supersession lineage.

Time:

- Effective approved timesheet version.
- Timesheet lines.
- Attendance interval identities.
- Correction-group lineage.
- Relevant break evidence and conflict status.

SHG:

- Instruction identity.
- Effective status.
- Verification evidence.
- Effective interval.

Rule packs:

- Exact pack identity.
- Version.
- Jurisdiction.
- Effective interval.
- Activation record.
- Manifest digest.
- Evidence references.

Check for omitted inputs that could change payroll without changing the snapshot hash.

## E2. Canonical hash

Verify the hash is based on canonical substantive content.

Inspect:

- JSON field ordering.
- Collection sorting.
- Exact monetary strings.
- Timezone and business-date representation.
- Stable rule-pack ordering.
- Effective intervals.
- Correction lineage.
- Inclusion/exclusion decisions.
- Schema version.

Verify:

- Same substantive content in two different runs produces the same hash.
- Different employee identities produce different content.
- Different compensation values produce different content.
- Different timesheet versions produce different content.
- Different rule-pack versions produce different content.
- Different employment intervals produce different content.
- Changed inclusion/exclusion decisions produce different content.

Random snapshot UUIDs and volatile timestamps should not make substantively identical snapshots hash differently.

A hash is an integrity fingerprint, not a substitute for stored snapshot contents.

## E3. Snapshot immutability

As `app`, attempt:

- UPDATE snapshot header.
- UPDATE child employee rows.
- UPDATE pack pins.
- DELETE child rows.
- Re-parent children.
- Replace the content hash.
- Insert new children after validation.
- Alter the current snapshot pointer.

Every protected change must fail or follow a deliberate invalidation/revalidation lifecycle.

Verify failed operations leave persisted state unchanged.

## E4. Invalidation and revalidation

Create a validated snapshot, then change each source category independently:

- Compensation.
- Employment terms.
- Status.
- Attendance.
- Timesheet successor.
- SHG instruction.
- Rule-pack activation.
- Employee population.

Verify:

- Original snapshot content remains byte-identical.
- An append-only invalidation event is recorded.
- The old snapshot is no longer the valid current tip.
- Revalidation creates a new snapshot version.
- The new snapshot has the correct new hash.
- There cannot be two valid tips.
- A stale validation offer cannot silently adopt a different snapshot.

---

# 8. Review Area F — Payroll-run lifecycle and direct-SQL governance

Inspect:

- `internal/payroll/runs.go`
- `internal/payroll/validate.go`
- Run lifecycle trigger.
- Run CHECK constraints.
- Partial unique indexes.
- Snapshot relationships.
- Run exclusions.
- Relevant tests.

The report acknowledges a specific limitation:

The `app` role can perform an outcome-equivalent `draft → validated` transition when a valid snapshot tip exists, but bypass service pin-stamping and audit.

Determine whether this is acceptable or a governance defect.

## F1. Direct validation bypass

Test direct SQL under `app`:

1. Create a valid draft run.
2. Create a valid snapshot tip.
3. Attempt `draft → validated` without using `ValidateRun`.
4. Inspect run state, snapshot references, basis pins and audit.

Determine whether the bypass can cause:

- Missing validation actor.
- Missing validation audit.
- Missing validation-offer identity.
- Stale or incorrect snapshot association.
- Inconsistent current tip.
- Different service and database validation outcomes.

If the direct transition creates a validated run without complete governance evidence, classify it appropriately.

Do not assume the trigger's terminality check makes the service audit dispensable.

## F2. State-machine correctness

Verify:

- Draft creation.
- Draft-only exclusions.
- Validation.
- Blocked validation.
- Stale offer.
- Re-offer.
- Invalidation.
- Revalidation.

`calculated`, `approved` and `finalized` must remain structurally unreachable without genuine financial artifacts.

Test direct SQL bypass attempts.

The current milestone must not fabricate a calculation artifact.

## F3. Correction-run chains

Verify:

- Same employer.
- Same payroll period.
- Correct run kind.
- One validated regular run.
- One validated successor where permitted.
- Linear correction chains.
- No forks.
- No cycles.
- No arbitrary identity changes.
- Historical runs remain immutable.

## F4. Exclusions and population accounting

Verify:

- Exclusions are run-scoped.
- Employer and employee match.
- Period matches.
- Reasons and actor are preserved.
- Expected population remains visible.
- Empty population does not pass through vacuous truth.
- A missing timesheet cannot silently remove an expected employee.

## F5. Deletion and run history

The report says bare draft runs can be deleted, while history-backed runs are protected by foreign keys.

Determine whether:

- Draft deletion can remove useful audit evidence.
- A draft with exclusions can be deleted.
- A run with snapshots or lifecycle events can be deleted.
- Deletion is exposed through supported application interfaces.
- Deletion can make historical payroll-readiness decisions untraceable.

Classify allowed draft deletion as accepted behavior only if it does not compromise a documented audit or replay guarantee.

---

# 9. Review Area G — Payroll-line and future calculation boundary

The Milestone 2 report explicitly states that `payroll_line` remains mutable after validation until the calculator pins it.

This is a known limitation.

Inspect:

- `payroll_line` schema.
- Grants and triggers.
- Snapshot references.
- Current reads/writes.
- Any future calculation interface.

Determine whether `payroll_line` currently influences:

- Readiness.
- Snapshot hash.
- Payroll run validation.
- Run exclusions.
- Any amount or financial summary.

A mutable table that has no current role in the snapshot may be an accepted pre-calculation limitation.

However, before the first calculation is introduced, the system must establish:

- Exact relationship between payroll lines and the frozen snapshot.
- Line-level provenance.
- Exact monetary representation.
- Immutability after calculation.
- Recalculation/supersession behavior.
- Protection against direct SQL alteration of calculated results.

Do not implement those features in this review.

State clearly whether the current mutability is harmless today and a mandatory prerequisite for the next calculation milestone.

---

# 10. Review Area H — Test harness and migration reliability

The report identifies a known test hazard:

Goose applies each migration version once, so appending new sections to an already-applied migration can leave reusable databases behind the current migration contents.

The project's standing procedure drops all `a1test*` databases before full runs.

Verify that the test harness cannot silently use stale database schemas.

Use new uniquely named disposable databases for this review where possible.

Inspect:

- Goose migration checks.
- Database recycling.
- `tests/dbutil`.
- Test provisioning.
- Test package parallelism.
- Migration version table.
- Test environment variables.

Do not reuse an existing database merely because it has the expected name.

Verify that a migration checksum or equivalent schema/version guard can detect an unapplied migration-content change where relevant.

Do not modify the harness.

The review must distinguish:

- Verified clean-server execution.
- Warm-server execution.
- Tests with no test files.
- Intentionally deferred calculation tests.
- Skipped database tests.

The reported 19-package result must be independently reconciled, not merely copied.

---

# 11. Focused independent test execution

Use real PostgreSQL 16 and the actual `app` role.

Run only the relevant suites in addition to the reported full battery.

At minimum:

- `task lint`
- `task build`
- `GOWORK=off go test -count=1 ./...`
- Full suite under `TZ=UTC`
- Full suite under `TZ=Asia/Singapore`
- Race tests for HR, payroll, attendance, timesheet and WFM
- Compensation resolver tests
- SHG lifecycle tests
- Rule-pack activation tests
- Snapshot capture and hash tests
- Snapshot concurrency tests
- Direct-SQL application-role privilege tests
- Payroll-run transition tests
- CLI/REST parity tests
- OpenAPI contract test
- `git diff --check`

The most important additional live tests are:

1. Snapshot transaction waits on population lock while a writer commits.
2. Snapshot transaction waits on employee or pack lock while a writer commits.
3. Direct SQL draft-to-validated attempt and audit inspection.
4. Snapshot drift after approval of a source record.
5. Rule-pack activation racing with snapshot capture.
6. Employee creation racing with snapshot capture.
7. Two identical payroll runs produce identical hashes.
8. Two runs with different compensation or timesheet versions produce different hashes.
9. Direct SQL cannot mutate validated snapshot contents.

Do not:

- Count skipped database tests as passes.
- Use a superuser to claim application-role integrity.
- Use mocked timestamps instead of real transaction interleavings.
- Rely on a warm database when a clean-server test is required.
- Claim Docker or Compose verification without a Docker daemon.

Record exact commands, database identities and results.

---

# 12. Findings classification

Use:

**P0 — Critical**

- Payroll snapshot can mix incompatible source versions.
- Approved inputs can silently change.
- Expected employees can be omitted.
- Statutory rule-pack gate can be bypassed.
- Snapshot content can be altered without detection.

**P1 — High**

- Validation or approval can bypass required governance.
- Missing, pending or ambiguous SHG input can be treated as complete.
- Compensation resolver can return a wrong effective basis.
- Rule-pack version or evidence identity can change silently.
- Snapshot drift is not detected.
- Concurrent operations can produce inconsistent runs.

**P2 — Medium**

- Bounded audit or diagnostics gaps.
- Draft deletion or lifecycle limitations without financial-integrity consequences.
- Test harness limitations with documented safe workarounds.
- Future calculation protections not yet needed by the current validated-only lifecycle.

**P3 — Low**

- Documentation.
- Maintainability.
- Non-critical developer experience.

Every finding must state:

- ID.
- Severity.
- Source or database object.
- Concrete failure scenario.
- Actual test or inspection evidence.
- Whether it blocks the first payroll calculation.
- Minimum correction.

Separate confirmed defects, risks, untested hypotheses and accepted limitations.

---

# 13. Final report

Create:

`plans/milestone2_adversarial_review.md`

Include:

1. Git baseline and review scope.
2. Actual commands executed.
3. Compensation resolver findings.
4. SHG lifecycle findings.
5. Rule-pack activation findings.
6. Snapshot isolation and population-barrier findings.
7. Snapshot completeness, hashing and immutability findings.
8. Payroll-run lifecycle and direct-SQL findings.
9. Payroll-line future-calculation boundary.
10. Migration and test-harness verification.
11. Confirmed defects.
12. Risks and untested hypotheses.
13. Clean areas.
14. Four verdicts:

- Compensation and employee-input integrity.
- SHG and rule-pack governance.
- Snapshot and payroll-run integrity.
- CLI/API operability.

15. Overall Milestone 2 acceptance decision.
16. Minimum prerequisites before the first calculation.
17. Explicitly deferred issues.

Do not modify any application source, migration, test, API contract or configuration.

Keep all work uncommitted.

---

# 14. Final decision

Return one of:

### PASS — READY FOR FIRST PAYROLL CALCULATION

All required input, rule-version, snapshot and validation guarantees are independently verified. No material P0/P1 issue threatens the restricted first payroll scenario.

### PASS WITH CONDITIONS

The foundation is trustworthy, but a small number of clearly defined prerequisites must be completed before calculation implementation.

### FAIL — NOT READY

A remaining defect can produce an incomplete, inconsistent, untraceable or ungoverned payroll-input set.

Do not declare PASS just because the 19 packages are green.

The database, direct application-role operations and concurrent transaction behavior must support the same business invariants as the Go services.

---

# 15. Scope restrictions

Do not implement:

- CPF.
- SDL.
- SHG contribution calculations.
- Overtime.
- Gross or net salary.
- Financial payslips.
- A generic rules engine.
- Phys-Pseudo interpreter/compiler.
- More countries.
- RLS or SSO.
- UI.
- Microservices.
- Event bus.
- Advanced WFM features.
- New architecture frameworks.

`SG-MOM-OT-001` remains BLOCKED.

The real statutory packages remain in their actual evidence-backed states.

---

# 16. Stop condition

The purpose of this review is to decide whether the existing governance and snapshot architecture can support one restricted, independently verified payroll calculation.

If the review passes, recommend closing the infrastructure and governance phase.

If it identifies defects, list only the minimum corrections that directly affect the first payroll scenario.

Do not produce another broad architecture roadmap.

**Establish whether one payroll run can capture a complete and coherent employee population, resolve the correct historical inputs, pin verified rule versions, preserve an immutable snapshot and prevent unauthorized lifecycle transitions. Then stop and provide a decisive verdict.**

## My particular concern

I would give the agent one additional instruction verbally: pay special attention to the REPEATABLE READ snapshot acquisition test.

The report says the snapshot uses a consistent transaction view and advisory locks. That is the correct general approach, but it is important to establish exactly when PostgreSQL fixes the MVCC snapshot relative to a lock wait. This should be tested with two live transactions, not assumed from the design description.

Also, the implementation report explicitly records the direct-SQL validation/audit gap and mutable `payroll_line` limitation. These must be evaluated as pre-calculation requirements, not silently forgotten.&#x20;

milestone2_remediation_report.md

milestone2_remediation_report.md



My recommendation: execute Prompt 16 before writing any financial calculation code. If it returns PASS or a short list of bounded conditions, we can finally move to the first Singapore payroll amount with a defensible foundation.