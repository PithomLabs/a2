# Prompt 8 — Adversarial Code Review: WorkforceOps Lab Milestone 1

## Objective

Perform an independent, adversarial technical code review of the completed WorkforceOps Lab Milestone 1 implementation.

The milestone report claims that the first integrated workflow works end to end:

`Employee → Employment Terms → BPO Assignment → Published Roster → Clock Events → Attendance → Approved Timesheet → Payroll Readiness`

Both CLI and REST implementations reportedly execute against real PostgreSQL and share the same application-service layer.

Your job is to determine whether those claims are actually supported by the implementation.

Do not assume that green tests prove correctness. Do not accept the implementation report as evidence of correctness without inspecting the source code, database behavior, API contracts and test quality.

**This is a review-only task. Do not fix code, refactor, change schemas, update dependencies or implement additional features.**

Produce an evidence-based adversarial review identifying defects, security weaknesses, incorrect business behavior, missing tests and unsupported claims.

## 1. Review boundaries and repository preservation

The implementation is uncommitted over baseline commit `f435158`, with approximately 79 changed entries.

Before reviewing:

1. Inspect Git status, branch and current commit.
2. Confirm the relationship to `f435158`.
3. Record tracked modifications, deletions, additions and untracked files.
4. Inspect the complete diff against the correct baseline.
5. Inspect untracked files separately; they are not necessarily present in `git diff`.
6. Read `AGENTS.md` and the Milestone 1 implementation report.

Preserve every existing change.

Do not:
- Run `git reset`, `git checkout`, `git clean` or destructive cleanup commands.
- Stage, commit, amend or discard changes.
- Rewrite files or automatically format the code.
- Modify migrations or database fixtures.
- Modify the upstream Phys-Pseudo or Phys-Research repositories.
- Start implementation of Milestone 2.

Use a separate disposable PostgreSQL database for any additional tests. Never drop or modify an existing user database.

The final report must identify the exact baseline and working-tree state reviewed.

## 2. Adversarial review priorities

Investigate the following areas in order of importance.

### A. Application architecture and service boundaries

Verify that:

- There is genuinely one `workforceops` executable.
- The three pre-pivot executables have been removed without leaving broken references.
- CLI and REST handlers call the same application services.
- Business logic is not duplicated in commands or HTTP handlers.
- Services consistently require explicit employer context.
- Repositories cannot accidentally query records belonging to another employer.
- Runtime uses the unprivileged `app` database role.
- No service bypasses the established database constraints or immutability protections.

Inspect actual call paths, not just package names or architectural tests.

Identify any handler or CLI command that performs business calculations, validation or database operations independently of the shared services.

### B. Employer isolation and data integrity

Adversarially test the multi-employer-ready database design.

Inspect every employer-scoped query and relationship.

Look for:

- Missing `employer_id` filters.
- Incorrect composite foreign keys.
- Cross-employer references.
- Employer context derived from untrusted input without validation.
- Queries that retrieve an entity by UUID without checking employer ownership.
- List and export operations that could expose another employer's records.
- Incorrect employer selection by code or UUID.
- Missing employer-scoped uniqueness.
- Inconsistent operator attribution.
- Transactions that allow partially completed business operations.

Use two employers with overlapping business identifiers.

Specifically attempt:
- Employer A employee referenced by Employer B roster.
- Employer A roster assignment referenced by Employer B attendance.
- Employer A timesheet linked to Employer B employee.
- Employer A payroll readiness including Employer B employees.
- Employer A requests retrieving Employer B records.
- Identical employee and programme codes under separate employers.

Do not assume that a test of one cross-employer relationship proves the entire ownership graph safe.

### C. HR and employment-term correctness

Review:

- Employee creation and status transitions.
- Effective-dated employment terms.
- Overlap detection.
- Boundary dates.
- Historical and future-dated terms.
- Compensation-line attribution.
- Employee/team/programme membership.
- Duplicate records and idempotent operations.

Check whether employee status and employment terms can produce contradictory states.

Check whether an employment-term change can retroactively alter data already used by an approved timesheet.

Review validation of salary and other monetary inputs.

No legal interpretation should be inferred from a job title or missing employee classification.

### D. BPO configuration, skills and rostering

Review:
- Client and programme ownership.
- Team membership.
- Skill catalogue and employee skill assignments.
- Programme-required skills.
- Shift templates and activity segments.
- Rotation patterns and assignments.
- Draft roster creation.
- Assignment validation.
- Roster publication.
- Revision handling.

Adversarial cases:

1. Employee without a required skill.
2. Employee whose skill validity has expired.
3. Employee whose team membership has expired.
4. Employee assigned to a programme outside the relevant team.
5. Duplicate assignment for the same business date.
6. Assignment on a rest day.
7. Rotation assignment with invalid dates or conflicting ownership.
8. Cross-midnight shift.
9. Shift template with invalid or overlapping segments.
10. Publishing a roster twice.
11. Concurrent publication attempts.
12. Modification of an old published revision.
13. Creation of a new revision without preserving the previous published state.

Check whether the system validates all constraints before publishing, or partially mutates the roster before discovering an error.

Confirm that operationally published rosters are never falsely described as legally validated.

`SG-MOM-SHIFT` remains unverified. Legal validation must remain unavailable where the required rule pack is not IMPLEMENTABLE.

### E. Attendance and clock-event correctness

This is a particularly high-risk area. Review the complete implementation of migration `00008_break_events`, CSV parsing, event deduplication, pairing, interval generation, corrections and review.

The system must preserve raw clock evidence.

Test or reason through:

- Valid clock-in/clock-out pairs.
- Duplicate source event delivery.
- Same source event ID under different employers.
- Same source event ID under different sources.
- Out-of-order events.
- Missing clock-in.
- Missing clock-out.
- Duplicate clock-in.
- Duplicate clock-out.
- Overlapping intervals.
- Multiple shifts on the same day.
- Cross-midnight shifts.
- Cross-calendar-date attribution.
- Break start without break end.
- Break end without break start.
- Duplicate break events.
- Paid versus unpaid breaks.
- Break evidence inconsistent with the roster.
- Break events occurring outside a work interval.
- Correction of an existing attendance interval.
- Repeated correction requests.
- Concurrent event imports.

Verify that the system does not:

- Treat scheduled breaks as observed breaks.
- Invent missing clock events.
- Silently discard unmatched events.
- Double-count attendance intervals.
- Lose original source-event provenance.
- Alter raw clock events during correction.
- Convert missing information to zero hours without an explicit rule.
- Delete actual work merely because a statutory working-time limit was exceeded.

Inspect timezone handling and conversion to Singapore business dates.

The 72-hour overtime limit is a working-time compliance issue, not a payable-hours truncation rule. No `min(H,72)` payable-hours behavior may exist.

Check whether event pairing is deterministic: identical source records and identical input ordering must not produce different attendance results.

### F. Attendance review and correction

Inspect whether corrections:

- Preserve the original evidence.
- Create auditable superseding records.
- Require a reason and actor.
- Preserve the old and new values.
- Maintain correct employee and employer relationships.
- Update downstream attendance aggregates consistently.
- Are rejected when the underlying record is already frozen or otherwise unavailable for correction.

Look specifically for stale attendance aggregates or timesheets that continue referencing superseded records without showing that their source has changed.

A correction must not silently rewrite an approved timesheet.

### G. Timesheet generation and approval

Verify:

- Timesheets are generated only from eligible reviewed attendance.
- Every timesheet line has correct source-interval lineage.
- Duplicate generation does not double-count hours.
- Unmatched or ambiguous attendance cannot silently become approved time.
- Draft timesheets can be edited.
- Submitted timesheets follow the permitted workflow.
- Rejected timesheets can be corrected and resubmitted according to the contract.
- Approved timesheets and their lines are immutable.
- Approval is transactional and idempotent.
- Cross-employer and cross-employee links are rejected.

Adversarially test interval overlap and duplicate interval allocation.

One actual attendance interval must not be counted twice in the same payable timesheet unless a specifically documented and justified split is supported.

Check whether a correction after timesheet approval requires a new auditable process rather than modifying the approved artifact.

### H. Payroll readiness and statutory-rule isolation

Inspect `internal/payroll` and the payroll-readiness implementation carefully.

Verify that:

- Readiness is an assessment, not a payroll calculation.
- No gross salary, overtime, CPF, SDL, SHG or net salary is invented.
- Required rule-pack statuses are read correctly.
- Only IMPLEMENTABLE versions can be used.
- Missing or ambiguous employee inputs produce explicit blockers.
- Payroll exclusions are visible.
- The all-or-nothing approval predicate remains intact.
- A blocked employee is never silently omitted.
- No readiness result is described as statutory payroll compliance.

Check the apparent milestone behavior:

The E2E test produced an approved 39.5-hour regular timesheet, yet payroll readiness was false because there were no included employees with complete input and all nine packs were BLOCKED.

Determine whether this is the correct consequence of the current blocking matrix, or whether the implementation is incorrectly conflating:
- An unverified payroll component.
- A missing employee instruction.
- An incomplete employee population.
- A payroll period with no eligible employees.
- An employee with an approved timesheet.

Do not change the blocking policy during this review. Report any ambiguity or incorrect implementation.

### I. API correctness and OpenAPI contract

The report states 36 specification paths, 46 operations and 38 implemented operations.

Verify those counts independently.

Review:
- Every served route and HTTP method.
- Request validation.
- Response schemas.
- Error envelopes and machine-readable codes.
- Employer-scoped path handling.
- `X-Operator` enforcement on audited writes.
- Loopback-only binding.
- CORS behavior, if present.
- Unexpected method handling.
- Pagination and list response shape.
- Idempotency behavior.
- State-transition enforcement.
- OpenAPI maturity flags.
- Undocumented handlers.
- Specified but unimplemented handlers.
- Incorrect success responses when a service operation fails.

The current report explicitly acknowledges:
- Bare arrays instead of the documented collection envelope.
- No enforced general `Idempotency-Key`.
- No attendance-interval listing endpoint.
- CLI-only rotation-pattern operations.
- CLI-only skill catalogue listing.

Assess the practical impact of these limitations.

In particular, determine whether an API consumer can discover an attendance interval before submitting a correction. If not, identify the missing workflow and its severity.

Verify that every route advertised as implemented is genuinely usable and has a contract test.

Do not count a handler that returns a placeholder or incomplete response as a functioning API operation.

### J. CLI quality and operational usability

Verify:
- Commands documented in `--help`.
- File-based input validation.
- JSON output.
- Exit codes.
- Operator attribution.
- Employer selection.
- Error handling.
- Idempotent commands.
- Useful identifiers in success responses.
- Correct handling of invalid or unsupported business cases.

Check whether the documented fresh-install workflow works without requiring undocumented SQL.

Identify any important operation that cannot be completed through either CLI or REST without manually modifying PostgreSQL.

### K. PostgreSQL migrations, roles and transactions

Inspect the complete migration chain, especially `00008_break_events`, and the existing trigger/privilege foundations.

Verify:
- Migration order.
- Upgrade from the baseline.
- Fresh database installation.
- Constraints and indexes.
- Runtime role grants.
- Trigger behavior under the actual `app` role.
- Transaction boundaries.
- Concurrent writes.
- Rollback behavior.
- Database error translation.
- Referential integrity.

Review whether any service relies solely on Go validation where a database constraint is needed for a critical invariant.

Check for SQL injection, unsafe dynamic SQL and resource leaks.

Do not modify migrations during review.

### L. Security and privacy

The system is loopback-only and synthetic-data-only in v0.1.

Verify that:
- API binding cannot silently become externally accessible.
- Employer identifiers cannot be used to bypass employer context.
- Operator attribution is never presented as authentication.
- Sensitive data is not leaked through logs, errors or responses.
- SQL queries are parameterized.
- CSV import cannot inject unintended data or execute unexpected operations.
- File inputs are bounded and validated.
- Untrusted JSON cannot create invalid state.
- Database errors do not expose unnecessary internals.

Clearly distinguish v0.1 accepted limitations from defects that prevent even local safe operation.

## 3. Review the tests adversarially

Do not just run the tests.

Inspect whether the tests actually detect the failure conditions they claim to cover.

Look for:
- Assertions that merely check a non-nil result.
- Tests that never assert persisted state.
- Tests that pass due to empty fixtures.
- Mocks that hide database behavior.
- Incorrect database identities.
- Tests using superuser instead of `app`.
- Weak or missing negative tests.
- Test-order dependencies.
- Concurrency gaps.
- Tests that duplicate implementation logic to calculate expected values.

Run the relevant existing test suite against a disposable PostgreSQL 16 database.

At minimum:
- `task lint`
- `task build`
- `go test -count=1 ./...`
- DB-gated integrity and service tests.
- OpenAPI structural validation.
- Fresh-database migration and demo workflow.

Do not claim any test passed unless it was actually executed.

The report's claim of CLI↔REST parity must be checked for meaningful business-result comparison, not just matching HTTP status codes or command exit codes.

## 4. Examine the known limitations explicitly

Evaluate these reported limitations and decide whether each is:

- Acceptable for Milestone 1.
- A defect requiring correction before the next milestone.
- A future enhancement that can safely wait.

Known limitations:

1. Bare JSON arrays instead of the collection envelope and pagination.
2. General `Idempotency-Key` not enforced.
3. Attendance interval IDs only discoverable through SQL.
4. Rotation patterns are CLI-only.
5. Skill catalogue listing is CLI-only.
6. Operator attribution is not authentication.
7. Legacy oracle/scenario Taskfile commands remain for future phases.

Do not automatically classify every documented limitation as harmless. Determine its effect on the intended business workflow.

## 5. Review Phys-Research and Phys-Pseudo boundaries

Inspect the changes for unintended weakening of the statutory-rule methodology.

Verify that:
- `SG-MOM-OT-001` remains BLOCKED.
- No overtime calculation has been implemented.
- No CPF, SDL, SHG or net-pay calculation has been invented.
- No golden fixture comes from an unresolved legal interpretation.
- The 72-hour limit is never treated as a payable-hours cap.
- Provisional MOM transcripts are not described as byte-grade pinned legislation.
- Unsupported cases map to explicit machine-readable blocked/indeterminate paths.
- The demo's exactly-1.5× contractual rate is clearly identified as a Meridian employer assumption, not a universal legal requirement.
- The Go code does not reintroduce statutory interpretations removed from the formal specification.

Phys-Pseudo remains a specification method, not an executable rules engine.

## 6. Severity classification

Classify findings:

**P0 — Critical**
- Incorrect business results.
- Cross-employer data leakage.
- Payroll or attendance data corruption.
- Unrecoverable loss of source evidence.
- A path allowing unverified statutory calculations.
- Approved records silently changing.
- Major security boundary failure.

**P1 — High**
- Incorrect workflow transitions.
- Missing critical validation.
- Incorrect attendance or timesheet attribution.
- Material API defects that make a promised operation unusable.
- Concurrency or idempotency defects likely to create duplicate business records.
- A contract that cannot be fulfilled through the exposed interfaces.

**P2 — Medium**
- Operational limitations.
- Missing useful endpoints.
- Incomplete error responses.
- Weak test coverage for non-critical behavior.
- Documentation/API inconsistencies.

**P3 — Low**
- Maintainability, naming, code organization, minor documentation and non-critical developer experience issues.

Do not inflate severity. Explain concrete impact.

## 7. Required finding format

Every finding must include:

1. Finding ID.
2. Severity.
3. Affected file and exact line/function.
4. The violated invariant or requirement.
5. Concrete failure scenario.
6. Why the existing tests did not catch it.
7. Reproduction procedure or test evidence, where possible.
8. Recommended correction.
9. Whether it blocks Milestone 1 acceptance or only the next milestone.

Separate confirmed defects from potential risks and untested hypotheses.

Do not report stylistic preferences as defects.

## 8. Final assessment

Give four independent verdicts:

### A. Milestone 1 functional correctness
Does the employee-to-approved-timesheet workflow genuinely work?

### B. Database integrity
Are employer isolation, attendance lineage and record immutability correctly enforced?

### C. CLI/API usability
Can a real operator complete the promised workflow without SQL intervention or undocumented workarounds?

### D. Payroll readiness integrity
Does the implementation correctly refuse to invent financial results while required rule packages remain blocked?

Then give an overall recommendation:

- PASS — acceptable to proceed.
- PASS WITH CONDITIONS — proceed after explicitly listed bounded corrections.
- FAIL — one or more material defects must be corrected before the next milestone.

## 9. Remediation prioritization

Do not implement fixes.

Instead, conclude with the smallest justified remediation list, grouped into:

1. Must fix before Milestone 2.
2. Should fix during Milestone 2.
3. Can safely defer.

Avoid proposing architecture redesign.

The goal is to preserve the working application and make the smallest changes necessary to improve correctness and delivery confidence.

## 10. Deliverable

Create:

`plans/milestone1_adversarial_review.md`

Include:
- Review scope and Git baseline.
- Review methodology and actual commands executed.
- Independent test results.
- Ranked findings.
- Known-limitation assessment.
- Four verdicts.
- Overall recommendation.
- Minimal remediation sequence.
- Explicit items not independently verified.

Do not modify application source, tests, migrations, OpenAPI contracts or existing design documents.

Keep all changes uncommitted. The only new file permitted is the adversarial review report.

If no defect is found in an area, say so explicitly and identify the evidence examined.

Be skeptical of the implementation report, but fair to the engineering work. The objective is to discover defects before they become expensive, not to manufacture criticism.