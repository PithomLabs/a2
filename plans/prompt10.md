This review should be independent, adversarial and read-only. The main objective is to establish whether Milestone 1.1 genuinely repaired the defects from the previous review, particularly the new append-only correction model, PostgreSQL protections and attendance-to-payroll evidence chain.

I have included a specific examination of `00009`, its interaction with migrations `00001–00008`, and the `SECURITY DEFINER` trigger design. These deserve more scrutiny than simply rerunning the green test suite.

# Prompt 10 — Independent Adversarial Code Review

# WorkforceOps Lab — Milestone 1.1 Final Integrity Verification

## Objective

Conduct an independent, adversarial code review of the completed WorkforceOps Milestone 1.1 implementation.

Determine whether the remediation actually corrected the defects identified in Prompt 8 and whether the new implementation contains structural, database, transactional or architectural defects that could compromise the reliability of the integrated HR, WFM and payroll backend.

The primary review artifacts are:

- `plans/milestone1_1_structural_audit.md`
- `plans/milestone1_1_remediation_report.md`
- `migrations/core/00009_m11_integrity.sql`
- The complete migration history `00001` through `00009`.
- The current Go implementation and regression tests.

Milestone 1.1 reports PASS against all ten acceptance conditions.

Do not assume that this verdict is correct. Independently establish whether the evidence supports it.

**This is a review-only task. Do not modify source code, migrations, tests, API contracts, configuration or documentation.**

Your job is to verify the implementation, identify remaining defects and assess whether the application is structurally ready to proceed to the first payroll calculation milestone.

Do not manufacture criticism. Preserve findings that are genuinely clean.

---

# 1. Preserve and establish the review baseline

Before starting:

1. Inspect `git status --short`.
2. Record `git rev-parse HEAD`.
3. Confirm the relationship to baseline `f43515883668c917edac31be49a7cf8334eca5c2`.
4. Record all modified, deleted, renamed and untracked entries.
5. Read the complete diff, including untracked files separately.
6. Confirm that migrations `00001`–`00008` remain unchanged and identify all later migrations.
7. Record Go, PostgreSQL and relevant tool versions.

The working tree is expected to contain uncommitted Milestone 1 and Milestone 1.1 work.

Preserve everything.

Do not:

- Reset or clean the repository.
- Stage or commit changes.
- Reformat or rewrite files.
- Modify existing tests.
- Alter migrations.
- Run destructive database operations against existing user databases.
- Modify `bchat/`.
- Modify the upstream Phys-Research or Phys-Pseudo repositories.
- Deploy to Neon or Fly.io.
- Implement new business features.

Any additional database testing must use explicitly verified disposable PostgreSQL databases.

The only permitted new repository file is:

`plans/milestone1_1_final_adversarial_review.md`

Temporary SQL probes, test scripts and logs may be placed outside the repository.

---

# 2. Establish the evidence standard

Read the complete structural audit and remediation report.

Create a finding cross-reference between:

- Original Prompt 8 finding ID.
- Structural audit classification.
- Milestone 1.1 claimed correction.
- Relevant source files and migration objects.
- Regression test.
- Independent verification result.

Use four verification outcomes:

- CONFIRMED FIXED: implementation and independent test demonstrate the invariant.
- CONFIRMED DEFECT: a reproducible violation remains.
- NOT REPRODUCED / DISPROVED: controlled testing contradicts the original claim.
- UNVERIFIED: evidence is insufficient.

A test passing is not automatically proof that its intended invariant is enforced.

Inspect:

- The test's setup.
- The role used.
- The actual SQL executed.
- The persisted database state.
- The negative case.
- Whether the test can pass vacuously.

Do not silently convert an untested hypothesis into a defect or a clean area.

---

# 3. Highest-priority review: migration 00009 and historical integrity

Inspect `migrations/core/00009_m11_integrity.sql` line by line.

Also inspect the complete definitions established by `00001`–`00008`, especially:

- `00002_attendance_payroll.sql`
- `00004_protection.sql`
- `00006_trigger_return_fix.sql`
- `00008_break_events.sql`

Review the final database behavior, not merely the contents of migration 00009 in isolation.

## A. Append-only timesheet correction chain

The remediation introduces `supersedes_id`, one-successor enforcement, original/correction uniqueness and effective approval resolution.

Verify:

1. An original approved timesheet remains unchanged after reissue.
2. An original rejected timesheet remains unchanged after recovery.
3. A correction creates a new draft without modifying the original.
4. A correction cannot belong to another employer, employee or pay period.
5. Multiple successors cannot be created concurrently.
6. A correction cannot form a cycle.
7. A chain cannot branch.
8. A correction chain cannot cross an invalid period boundary.
9. An approved correction becomes the effective approved version.
10. A later draft or rejected correction does not incorrectly displace the currently effective approved version.
11. Readiness cannot count both the original and correction as effective.
12. Reissuing an approved correction does not corrupt the chain.
13. Historical records remain queryable and immutable.
14. Every state transition has a valid audit record.

Pay particular attention to the distinction between:

- Original record.
- Correction draft.
- Submitted correction.
- Rejected correction.
- Approved correction.
- Effective approved timesheet.

Do not assume that `EffectiveApproved` implements these distinctions correctly merely because its name suggests it does.

## B. Rollback with real correction data

The previous review raised a concern that the 00009 Down migration may fail after correction chains exist because it removes `supersedes_id` and restores the original employee/period uniqueness.

Test this directly.

Use a disposable database:

1. Apply all migrations.
2. Create an employee and pay period.
3. Create an original timesheet.
4. Create at least one correction successor.
5. Approve the appropriate version.
6. Execute the actual Down migration.
7. Record whether it succeeds.
8. Inspect resulting data and schema.
9. Execute the corresponding Up migration again.
10. Verify that no data or integrity behavior has silently changed.

Do not assume a successful empty-database rollback proves rollback safety.

If rollback with real correction data is intentionally unsupported, the migration must fail loudly before destructive changes, and the limitation must be documented.

## C. `is_current` and SECURITY DEFINER trigger

The new explicit current-interval predicate is maintained through a SECURITY DEFINER trigger owned by `migration_admin`.

This is one of the most important parts of the review.

Inspect:

- Function owner.
- `SECURITY DEFINER` behavior.
- Explicit and safe `search_path`.
- Schema-qualified table references.
- Execute privileges.
- Trigger installation and coverage.
- Trigger behavior for INSERT, UPDATE and DELETE.
- App-role privileges on underlying tables.
- Trigger recursion.
- Parent/successor lookup scoping.
- Employer isolation.
- Error handling.
- Concurrency behavior.

Adversarially test:

1. Insert a new current interval as `app`.
2. Supersede an interval.
3. Confirm the predecessor becomes non-current.
4. Confirm the successor becomes current.
5. Attempt to alter `is_current` directly as `app`.
6. Attempt to re-parent a successor to another employer or employee.
7. Attempt a second successor.
8. Attempt a supersession cycle.
9. Attempt concurrent corrections.
10. Attempt direct SQL mutations that could bypass the expected current-interval transition.

The runtime `app` role must not be able to defeat the maintenance trigger or directly manipulate current-state metadata.

Verify that the partial GiST exclusion constraint uses the correct current-record predicate.

A correction successor with a `supersedes_id` must still be recognized as current when appropriate.

## D. Freeze and re-parenting triggers

Inspect all immutability triggers, not just the functions changed in 00009.

For every UPDATE-triggered freeze, verify both OLD and NEW parents.

Test:

- Moving a draft child into an approved parent.
- Moving an approved child to a draft parent.
- Re-parenting across employers.
- Re-parenting across employees.
- Changing the parent during a simultaneous status transition.
- DELETE and INSERT behavior.

The database must not permit indirect mutation of an approved or published record through relationship changes.

Trigger error paths must not expose cross-employer record existence or status.

---

# 4. Attendance: verify the entire evidence lifecycle

Inspect:

- `internal/attendance/attendance.go`
- `internal/attendance/review.go`
- Attendance migrations and triggers.
- Attendance service tests.
- Attendance exception handling.
- Attendance interval listing and correction endpoints.

## A. Roster-anchored event pairing

Verify that the algorithm uses:

- Employer.
- Employee.
- Source/event stream.
- Applicable roster assignment.
- Expected shift window.
- Cross-midnight behavior.
- Multiple shifts per day.
- Event timestamps and ordering.
- Duplicate and conflicting event handling.

The 24-hour ceiling is a defensive maximum, not a substitute for correct roster matching.

Test:

- Normal day shift.
- Evening-to-night shift.
- Cross-midnight shift.
- Multiple shifts on the same date.
- No roster.
- Wrong roster.
- Stale roster assignment.
- Out-of-order events.
- Month-long orphan span.
- Event just within and just outside the pairing boundary.

Unrostered or ambiguously attributable events must be preserved and flagged, not guessed into payable intervals.

## B. The 72-hour rule

Verify that the 72-hour monthly overtime limit is used exclusively for compliance assessment.

Prove that:

- Actual work is never truncated at 72 hours.
- Verified work above 72 hours remains in attendance.
- Verified intervals remain discoverable.
- The compliance flag does not cause verified work to disappear from timesheets.
- An unverified legal consequence does not become an invented payroll deduction.

The 210-hour regression scenario reported by Milestone 1.1 must be inspected and reproduced if practical.

## C. Cross-midnight attribution

Verify that a span crossing local midnight is split into the correct business-date intervals.

For an overnight shift, inspect:

- UTC instants.
- Local timestamps.
- Employee timezone.
- Singapore business date.
- Original shift identity.
- Roster assignment.
- Break intervals.
- Source-event references.
- Daily attendance aggregates.

Test midnight exactly, one second before midnight and one second after midnight.

Test both sides of a pay-period boundary.

Confirm the database session timezone cannot change the result.

## D. Late break attachment

Reproduce the scenario where:

1. Clock-in and clock-out are imported.
2. A work interval is generated.
3. Break events arrive in a later import.
4. The break should attach to the existing interval.
5. The original events and interval remain auditable.
6. The successor intervals are correct.
7. No duplicate work is created.

Test paid and unpaid break evidence, missing break boundaries, duplicate break events and ambiguous attribution.

## E. Import replay and conflicting delivery

Verify:

- Identical source-event replay is idempotent.
- Same ID with different payload becomes a conflict.
- Different IDs representing duplicate work do not double-count.
- Previously stored unpaired events can be processed on replay.
- Process interruption does not permanently strand valid events.
- Import and pairing transaction boundaries are correct.

Inspect all exception persistence paths.

A successful import response must not imply that every event has been successfully paired.

## F. Correction behavior

Verify:

- Correction creates an auditable successor.
- Original evidence remains unchanged.
- Only one successor can be created.
- Concurrent corrections produce exactly one winner.
- Correction cannot mutate an approved evidence chain.
- Draft-timesheet regeneration works.
- A correction crossing dates resolves the correct roster assignment.
- Both affected daily records are recalculated.
- An unrelated correction cannot clear a blocking exception.
- Superseded intervals are excluded from current-work calculations.
- Payroll readiness sees the correct effective evidence.

Inspect whether correction and aggregate updates are committed in one transaction.

---

# 5. Timesheets and payroll readiness

Inspect:

- `internal/timesheet/timesheet.go`
- `internal/payroll/readiness.go`
- `internal/payroll/approval.go`
- `internal/payroll/runs.go`
- Related migrations, triggers and tests.

## A. Timesheet allocation

Prove that:

- An attendance interval is allocated to no more than one effective payroll input.
- One interval cannot be linked to another employee's timesheet.
- One interval cannot be linked across employers.
- Superseded intervals cannot be counted.
- Correction versions cannot double-count the original.
- Overlapping pay periods are prohibited.
- An interval crossing midnight is allocated according to the local business-date contract.
- Generation is transactional and recoverable.

Test concurrent generation for the same employee and period.

Test a crash or failure after interval selection but before line persistence.

Replaying generation must not duplicate or omit time.

## B. Timesheet lifecycle

Test:

`Generate → Submit → Reject → Reissue → Submit → Approve`

Also test:

- Reissuing an approved sheet.
- Reissuing a correction version.
- Multiple reissue attempts.
- Rejected historical originals.
- Concurrent submission and approval.
- Attempts to change approved contents.
- Attempts to change supersession links.
- Attempts to allocate already-consumed intervals.

Confirm original artifacts remain immutable.

## C. Payroll population

The expected payroll population should be derived from effective-dated employment applicability, not solely from current employee status or existing approved timesheets.

Verify:

- Employees active for some or all of the period.
- Employees terminated after the period.
- Employees terminated during the period.
- Employees with approved historical timesheets.
- Employees expected but missing timesheets.
- Employees with term conflicts.
- Explicit payroll exclusions.
- Exclusions from unrelated periods.
- Employees with empty or zero-hour timesheets.
- Employees whose original timesheet was corrected.

A historical approved employee must never silently disappear from readiness merely because the employee is currently terminated.

## D. Approval semantics

Verify that:

- A missing or blocked statutory rule cannot be bypassed.
- An empty included population cannot pass approval.
- A payroll run belongs to the requested employer and pay period.
- Exclusions belong to that same employer and period.
- Exclusions are frozen at the correct lifecycle state.
- Reissue chains are resolved consistently.
- A zero-hour or empty approved timesheet cannot falsely satisfy payroll completeness.
- Readiness and approval use the same business rules.

No financial calculation is authorized in this review.

---

# 6. HR, WFM and roster lifecycle

Inspect:

- `internal/hr/hr.go`
- `internal/hr/shg.go`
- `internal/wfm/roster.go`
- Relevant migrations and tests.

## A. Effective-dated employee terms

Verify:

- Closing an open terms record.
- Creating a successor in the same transaction.
- No overlapping terms.
- No invalid backdating over approved periods.
- Historical records remain immutable.
- Compensation changes remain traceable.
- Changes to terms are reflected in readiness as explicit conflicts where appropriate.

## B. Employee status transitions

Test the transition matrix:

- Active → suspended.
- Suspended → active.
- Active → resigned.
- Active → terminated.
- Terminal status → attempted resurrection.
- Repeated same-status changes.
- Backdated transitions.
- Concurrent status changes.

The initial employee status must have a coherent historical representation.

Ensure every transition has an actor, reason and effective date.

## C. Roster publication and revision

Verify that publication validates all assignments in one transaction.

Required checks:

- Employee status.
- Team membership.
- Programme skill requirements.
- Skill validity.
- Roster-week boundaries.
- Rest-day assignment.
- Assignment uniqueness.
- Current roster revision.
- Non-empty roster or explicit reason for an empty roster.

Verify:

- Publish is idempotent in its documented form.
- Concurrent publish attempts cannot create duplicate revisions.
- Revise creates a new draft revision.
- Existing published assignments are copied correctly.
- The current revision pointer advances atomically.
- Old published versions remain unchanged.
- Publication metadata is correct.
- Concurrent publish/revise behavior is coherent.

Do not treat operational validation as statutory Singapore shift-worker compliance. The unverified `SG-MOM-SHIFT` package must remain unavailable for legal validation.

Rotation materialization remains explicitly deferred.

---

# 7. Database roles, migrations and fresh installation

Inspect:

- `core/database/migrate.go`
- `core/database/provision.go`
- `cmd/workforceops/migrate.go`
- `scripts/db/init-roles.sh`
- `tests/dbutil/dbutil.go`
- Compose configuration.
- All nine migrations.

## A. Two-role model

Verify:

- `migration_admin` owns protected schema objects and performs DDL.
- `app` runs normal application operations.
- Migration commands use `MIGRATION_DATABASE_URL`.
- Runtime commands use `DATABASE_URL`.
- Fresh installation provisions roles consistently.
- Privileges are verified after migration.
- `app` cannot bypass protected history.
- Service tests actually use `app`.
- Superuser access is confined to controlled provisioning/test setup.

Run the complete normal application workflow as `app`.

Inspect default privileges, object ownership and extension ownership.

## B. Migration preflight

Verify that migration preflight fails loudly for:

- Wrong database role.
- Missing `app`.
- Missing required extensions.
- Incorrect schema ownership.
- Missing required grants.
- Existing incompatible migration state.

The fresh-install-only pre-release decision must be documented accurately.

No existing user database may be dropped, recycled or overwritten.

The reported upgrade failure from the pre-pivot partial history must remain a documented migration limitation unless a real deployed database is discovered.

## C. Seed reset

Inspect the database-recycle implementation.

Verify:

- Explicit administrator DSN is required.
- Protected database names are refused.
- Same-server checks are enforced.
- No database outside the intended disposable target can be dropped.
- `app` does not receive broad DELETE privileges.
- Triggers are not disabled.
- No superuser bypass is exposed through normal application commands.
- Recycle provisions, migrates and seeds correctly.
- Failed recycle operations do not leave an ambiguous target.

Test `seed demo --reset` on a dedicated disposable database.

Do not point it at existing developer data.

---

# 8. Security review of the new database mechanisms

This review is not a general security certification, but inspect the high-risk boundaries.

In particular:

- SECURITY DEFINER functions.
- Function ownership and `search_path`.
- Trigger execution privileges.
- Runtime database grants.
- Employer-scoped trigger lookups.
- SQL parameterization.
- Error sanitization.
- DSN and secret exposure.
- Localhost binding.
- Operator attribution.
- Employer context resolution.
- Destructive database recycle checks.

`X-Operator` and `--operator` are attribution mechanisms, not authentication. Confirm that the application never represents them as verified identities.

Do not introduce RLS, SSO or multi-employer authentication as part of this review.

---

# 9. Application-service and API/CLI consistency

Verify that business behavior remains in shared Go application services.

Inspect for:

- Business logic duplicated in HTTP handlers.
- Business logic duplicated in Cobra commands.
- Direct handler-side database queries that should use a service.
- CLI-only business operations required for the supported workflow.
- REST-only mutations that do not have a CLI equivalent where one is required.
- Error mapping differences.
- Inconsistent transaction boundaries.
- Different default values between adapters.

Verify OpenAPI against actual served routes.

The remediation report states:

- 43 paths.
- All served operations match implemented operations.
- Draft operations remain unserved.
- A shared `{error, message, request_id}` error envelope.
- Employer-scoped interval retrieval and correction.

Check these claims directly.

The CLI and REST E2E comparison should compare persisted business results and readiness fields, not only response codes.

Rotation materialization and generic Idempotency-Key infrastructure are deferred. Verify that the documentation does not promise them as implemented.

---

# 10. Financial and statutory-rule boundaries

Inspect the code and schema for accidental statutory calculations or financial approximations.

Verify:

- Monetary inputs use decimal strings at API/CLI boundaries.
- Go uses integer minor units.
- PostgreSQL stores monetary values as `NUMERIC`.
- `float64` is not used for money.
- No invented statutory rounding has been introduced.
- No overtime, CPF, SDL, SHG or net-pay calculation exists.
- Readiness reports blocked conditions honestly.
- `SG-MOM-OT-001` remains BLOCKED.
- No fabricated golden fixtures exist.
- Provisional MOM transcripts are not described as byte-grade pinned legislation.
- The 72-hour provision remains separate from payment calculations.
- The 1.5× rate remains a Meridian contractual assumption, not a universal legal rule.

Inspect all paths that could bypass the rule-pack activation gate.

There must be no hidden calculation path inside a service, seed helper, API handler or CLI command.

---

# 11. Review the reported new defects and fixes

Independently verify the eight new defects described in the remediation report:

1. Original timesheet unique constraint prevented append-only correction.
2. No explicit current-interval predicate.
3. Duplicate spans with new source event IDs caused batch failure.
4. Seed reset failed against immutable records.
5. Error envelope leaked database internals and lacked request IDs.
6. Required workflow operations were absent from both interfaces.
7. Dockerfile/Fly configuration referenced deleted executables.
8. Cross-date corrections incorrectly derived day records and roster assignments.

For each item:

- Inspect the original defect.
- Inspect its correction.
- Identify the regression test.
- Reproduce the failure on the old version only if safely possible on a disposable database or from preserved probe evidence.
- Verify the corrected behavior.
- Record any remaining limitation.

Do not assume the remediation report's description is sufficient proof.

---

# 12. Execute independent verification

Use disposable PostgreSQL 16 databases.

Do not reuse databases with unknown contents.

Run the actual supported developer commands.

At minimum:

- `task lint`
- `task build`
- `go test -count=1 ./...`
- Race tests for attendance, timesheet, WFM, payroll and HR.
- Fresh-install migrations and role verification.
- Complete CLI E2E.
- Complete REST E2E.
- Persisted-state parity comparison.
- OpenAPI validation.
- `git diff --check`.

For migration Down/Up, use a populated disposable database containing actual attendance corrections, timesheet correction chains, published rosters and approved artifacts.

For the SECURITY DEFINER mechanism, perform direct SQL attempts under the real `app` role.

For concurrency, execute actual concurrent transactions rather than relying solely on static reasoning or `go test -race`.

Do not:

- Count skipped database tests as passes.
- Substitute superuser tests for application-role tests.
- Silently use serial execution after a parallel failure.
- Claim Docker build success without a Docker daemon.
- Claim Compose volume inspection when the volume was inaccessible.

Record commands actually executed and their outputs.

---

# 13. Additional structural questions

Determine whether the following issues remain or have been fully resolved:

1. Can a timesheet correction chain have multiple current approved versions?
2. Can a correction chain become cyclic or cross employer/period boundaries?
3. Can `EffectiveApproved` return the wrong version under concurrent approval?
4. Can a current attendance interval be removed from exclusion enforcement by changing its metadata?
5. Can a SECURITY DEFINER function be manipulated through a writable schema or unsafe search path?
6. Can a raw clock event remain stored but never become pairable or discoverable?
7. Can late break evidence create an incomplete supersession chain?
8. Can a corrected interval point to the wrong roster assignment?
9. Can an approved timesheet lose its original inputs through a later HR or attendance update?
10. Can two payroll runs reference the same period and create ambiguous approval state?
11. Can an exclusion reference the wrong employee or a correction version?
12. Can a historical employee be omitted from readiness?
13. Can a run pass approval with zero eligible employees?
14. Can a fresh install produce different privileges from the documented Compose or local workflow?
15. Can migration rollback destroy or invalidate correction history?
16. Can CLI and REST produce different behavior under invalid, missing or concurrent input?
17. Are audit records atomic with the mutations they describe?
18. Can any error path expose protected database or cross-employer information?

For each question, provide:

- Confirmed answer.
- Evidence.
- Test or reproduction.
- Remaining uncertainty.

These questions are intended to expose structural defects beyond the previously reported findings, not to require additional features without justification.

---

# 14. Finding classification

Use the following severity definitions:

**P0 — Critical**

- Incorrect attendance or payroll input that can silently become approved.
- Cross-employer data exposure.
- Unrecoverable loss of original evidence.
- Mutation of approved financial records.
- Bypass of mandatory statutory-rule gating.
- Fundamental database integrity failure.

**P1 — High**

- Broken business transitions.
- Incorrect temporal attribution.
- Duplicate or missing attendance.
- Inconsistent historical population.
- Unsafe migration/runtime role configuration.
- Concurrency defects that can corrupt business records.
- Operator cannot complete a promised workflow.

**P2 — Medium**

- Material usability limitations.
- Incomplete API contract.
- Missing recovery or diagnostics.
- Important but bounded test-coverage gaps.
- Operational configuration inconsistencies.

**P3 — Low**

- Cosmetic issues.
- Documentation quality.
- Non-critical maintainability.
- Future enhancement opportunities.

Do not inflate severity.

Distinguish:

- Confirmed defects.
- Risks requiring additional evidence.
- Untested hypotheses.
- Accepted MVP limitations.
- Clean areas.

---

# 15. Required report

Create:

`plans/milestone1_1_final_adversarial_review.md`

Include:

1. Git and working-tree baseline.
2. Review scope and methodology.
3. Actual commands executed.
4. Migration `00009` and historical-schema review.
5. Timesheet correction-chain review.
6. SECURITY DEFINER and `is_current` analysis.
7. Attendance and temporal correctness.
8. HR and WFM lifecycle correctness.
9. Payroll readiness integrity.
10. Database-role and migration review.
11. API/CLI consistency.
12. Statutory-rule isolation.
13. Verification of all Milestone 1.1 claims.
14. Confirmed defects, risks and untested hypotheses.
15. Clean areas.
16. Four independent verdicts:

- Functional correctness.
- Database integrity.
- CLI/API operability.
- Payroll-readiness integrity.

17. Overall recommendation.
18. Smallest justified remediation list.
19. Explicit unverified areas.

Use exact filenames, functions, migration objects and line numbers wherever possible.

For each finding give:

- Finding ID.
- Severity.
- Affected source.
- Invariant.
- Concrete failure scenario.
- Test evidence.
- Why the previous tests did not catch it.
- Recommended correction.
- Whether it blocks the first payroll calculation milestone.

Do not duplicate equivalent findings under different IDs.

---

# 16. Final decision criteria

The review should return one of:

### PASS

Milestone 1.1 is independently verified. No P0 or P1 defect threatens the supported employee-to-timesheet evidence chain or payroll readiness.

### PASS WITH CONDITIONS

The core evidence chain and database integrity are trustworthy, but a bounded set of non-critical or explicitly scoped limitations must be resolved before broader use.

### FAIL

A critical or high-priority defect remains that can corrupt, misattribute, duplicate, lose or silently change attendance, timesheet, employment or payroll-readiness data.

The first independently verified payroll calculation must not begin while a P0 or material P1 defect threatens its inputs or rule-gating behavior.

---

# 17. Scope restrictions

This is not authorization to build the next milestone.

Do not implement:

- CPF calculations.
- SDL calculations.
- SHG calculations.
- Overtime calculations.
- Net-pay calculations.
- Financial payslips.
- A generic payroll rules engine.
- Phys-Pseudo interpreter or compiler.
- Multi-employer authentication.
- RLS.
- Microservices.
- Event bus.
- UI.
- Rotation materialization.
- Generic Idempotency-Key framework.

Preserve the existing modular monolith and the Singapore-first, jurisdiction-extensible product direction.

Do not reopen frozen architectural decisions.

If a genuine architectural defect requires changing a frozen decision, explain it and request approval rather than changing the architecture silently.

## Final instruction

Be rigorous and skeptical, but fair.

Milestone 1.1 reports that the original 2 P0 and 13 blocking P1 findings have been corrected and that eight additional defects were discovered and fixed.

Your task is to independently verify that claim, with particular attention to immutable historical records, evidence lineage, SECURITY DEFINER behavior, temporal correctness and populated-database rollback.

**The goal is to establish that WorkforceOps is ready to calculate its first independently verified payroll amount on top of trustworthy operational data—not to keep the project in an endless review cycle.**

Stop when the evidence supports a clear verdict and a bounded remediation list. Keep all work uncommitted and make no code changes.

## What makes this review different

I have deliberately concentrated the review on three areas that became more consequential after Milestone 1.1:

1. Correction-chain integrity: the original approved record, its correction versions and the currently effective version must remain distinguishable throughout the lifecycle.
2. Database-enforced state: the new `is_current` mechanism and `SECURITY DEFINER` trigger must actually protect the invariants under the real application role.
3. Historical reproducibility: rollback, later employment changes and attendance corrections must not silently alter the interpretation of an already approved period.

The agent should also verify the reported 98 tests and the CLI/REST comparison, but the most important test is whether the system remains correct when events arrive late, corrections occur after approval, and transactions race.

My target for this review is a definitive PASS or a short, actionable list of defects—not another open-ended architecture exercise.