This should be the final focused adversarial review of Milestone 1.2, not another full architecture audit. Its purpose is to verify the five remaining questions and determine whether the reported PASS is justified.

# Prompt 12 — Final Focused Adversarial Code Review

# WorkforceOps Lab — Milestone 1.2 Acceptance Verification

## Objective

Independently verify the Milestone 1.2 implementation and determine whether it is safe to close the integrity-remediation phase and begin the first independently verified Singapore payroll calculation.

The implementation agent reports:

- All eight service-reachable P1 defects corrected.
- All ten PostgreSQL application-role bypass areas addressed.
- Fresh-install and employer onboarding repaired.
- Five additional Stage-E corrections implemented.
- New migration `00010_m12_integrity.sql`.
- 24 Go packages passing.
- Full suite passing under normal and UTC environments.
- Race tests passing across five service packages.
- Singapore overtime package still BLOCKED.
- No statutory payroll calculations introduced.

The remediation report is:

`plans/milestone1_2_remediation_report.md`

The implementation is uncommitted over baseline:

`f43515883668c917edac31be49a7cf8334eca5c2`

The reported PASS is provisional until independently examined.

**This is a focused, review-only task. Do not modify application source, migrations, tests, contracts, configuration or documentation.**

Concentrate on the five verification areas below. Do not repeat the entire Prompt 10 review or initiate another architecture redesign.

---

# 1. Preserve the repository and establish the baseline

Before reviewing:

1. Record `git status --short`.
2. Record `git rev-parse HEAD`.
3. Confirm HEAD remains the expected baseline.
4. Inventory staged, unstaged, deleted, renamed and untracked entries.
5. Verify migration files `00001`–`00009` remain untouched.
6. Confirm `00010_m12_integrity.sql` is the new forward migration.
7. Read the complete Milestone 1.2 remediation report.
8. Record Go and PostgreSQL versions.

Preserve all existing work.

Do not:

- Reset or clean the repository.
- Stage or commit changes.
- Rewrite or format files.
- Modify migrations.
- Drop or recycle any existing user database.
- Modify `bchat/`.
- Modify Phys-Research or Phys-Pseudo.
- Deploy to Neon or Fly.io.
- Implement new business functionality.

Use disposable PostgreSQL 16 databases for additional tests.

The only permitted new repository file is:

`plans/milestone1_2_final_adversarial_review.md`

Temporary SQL probes and test scripts may be placed outside the repository.

---

# 2. Review Area A — Cross-midnight correction-group integrity

\<badge>Highest priority\</badge>

The remediation introduces an `attendance_correction_group` to preserve the one-successor-per-original invariant while permitting one original attendance interval to be corrected into multiple local-date segments.

Inspect:

- `migrations/core/00010_m12_integrity.sql`
- `internal/attendance/attendance.go`
- `internal/attendance/review.go`
- Correction-group models and repository functions.
- Related attendance tests.
- Timesheet allocation and readiness logic.

## A1. Data-model correctness

Verify that the database represents these relationships correctly:

```
Original Attendance Interval
             │
             ▼
    Correction Group
             │
        ┌────┴────┐
        ▼         ▼
    Segment 1  Segment 2
    Date A     Date B
```

The actual model may differ, but it must preserve the following invariants:

- One original interval cannot have competing correction groups.
- A correction group belongs to exactly one employer and employee.
- Every successor segment belongs to the correction group.
- Every segment has a valid local business date.
- Segments cannot belong to another employee or employer.
- The group cannot be partially superseded.
- No segment can be duplicated.
- A group cannot cross an invalid pay-period or employee boundary.
- A correction group cannot form a lineage cycle.
- A second correction must supersede the current correction group through a controlled new version, not mutate the original.

Determine whether these invariants are enforced by PostgreSQL constraints/triggers, application services, or both.

Do not accept a service-only check where direct SQL by `app` can invalidate the lineage.

## A2. Atomicity

Use a disposable PostgreSQL database to test:

1. Create a valid original interval.
2. Correct it across local midnight.
3. Confirm the correction group and all segments persist.
4. Force an insertion failure after the first segment.
5. Verify the entire correction transaction rolls back.
6. Replay the correction.
7. Verify exactly one correction group and the correct segments exist.

Test concurrent correction attempts against the same original.

Exactly one valid correction group must win.

## A3. Date and roster attribution

Test:

- Same-date correction.
- Cross-midnight correction.
- Midnight exactly.
- One second before and after midnight.
- Cross-pay-period correction.
- New date with a different valid roster assignment.
- New date with no roster assignment.
- Correction with a break crossing midnight.
- Correction involving an already superseded interval.

Confirm that:

- Each segment uses the correct employee business date.
- Each segment resolves the appropriate roster assignment.
- A missing roster never inherits the previous date's assignment.
- Both affected daily records are recomputed.
- All audit records and source references are preserved.

## A4. Timesheet consumption

Verify that a correction group is treated coherently during timesheet generation.

A two-segment correction must not:

- Double-count the original.
- Lose one segment.
- Place both segments on the same business date.
- Allocate a segment to another pay period.
- Allow one segment to be independently approved while its group remains unresolved.
- Allow readiness to count both the original and correction.

Test generation, submission, rejection, reissue and approval of a timesheet containing correction-group segments.

**Acceptance:** the one-to-many segmentation must coexist with exactly one correction lineage per predecessor and correct downstream allocation.

---

# 3. Review Area B — Duplicate-event disposition and evidence lineage

The remediation introduces `clock_event_disposition` with operator-confirmed duplicate handling.

Inspect the actual migration, model, services, CLI and REST endpoints.

The earlier plan required a duplicate event to identify the canonical event it duplicates. The remediation summary does not fully establish whether this relationship was implemented.

Do not assume it exists.

## B1. Canonical event relationship

Determine whether the disposition contains a canonical reference such as:

`duplicate_of_event_id`

If present, verify its constraints.

If absent, determine whether another explicit, equivalent lineage mechanism exists.

If neither exists, report the missing canonical relationship as a defect or justified design limitation.

Where a canonical relationship exists, verify:

- Same employer.
- Same employee.
- Compatible event kind.
- No self-reference.
- No circular reference.
- Canonical event is not itself a duplicate.
- Duplicate cannot become eligible for pairing after disposition.
- Canonical evidence remains available for pairing and timesheet generation.

## B2. Adversarial duplicate scenarios

Test:

1. Identical event delivered twice with the same source ID.
2. Identical event under different source IDs.
3. Same source ID with different payload.
4. Duplicate event belonging to another employee.
5. Duplicate event belonging to another employer.
6. Duplicate clock-in versus clock-out.
7. Duplicate break event.
8. Attempt to duplicate an already dispositioned canonical event.
9. Conflicting concurrent disposition attempts.
10. Repeated confirmation by the same operator.

The system must preserve both original and duplicate raw events.

A duplicate disposition must not permit two actual work intervals or two payroll allocations.

## B3. Exception and audit lifecycle

Verify that:

- The original exception remains traceable.
- The disposition contains operator, reason and timestamp.
- The raw events cannot be rewritten.
- The decision is immutable or superseded through an explicit correction workflow.
- A genuine overlap cannot be classified as a duplicate merely to unblock generation.
- An unresolved overlap continues to block affected payroll inputs.
- A successfully confirmed duplicate does not permanently block timesheet generation.

**Acceptance:** duplicate handling is an auditable evidence decision, not merely a flag that removes an event from pairing.

---

# 4. Review Area C — Timezone independence and jurisdiction separation

The remediation report states that full tests passed under `TZ=UTC`.

The original plan also required verification under `TZ=Asia/Singapore` and mixed source-zone inputs.

Review:

- `internal/attendance/attendance.go`
- `internal/attendance/review.go`
- `internal/duration/`
- Employee/employment jurisdiction and timezone models.
- Any source timestamp parsing and PostgreSQL decoding code.
- All relevant timezone tests.

## C1. Verify three separate timezone concepts

The system must distinguish:

1. **Source/event timezone:** how an external vendor or clock device represents its timestamp.
2. **Employee employment timezone:** the timezone determining the employee's business-date and Singapore payroll attribution.
3. **Client/service timezone:** the timezone in which BPO service windows and client operating schedules are expressed.

For the current Singapore employment profile:

`Asia/Singapore`

must determine employment business-date attribution.

A client timezone or source-event timezone must not silently replace it.

## C2. Timestamp interpretation

Verify that:

- Offset-aware timestamps are interpreted exactly once.
- An explicitly supplied offset is not overwritten by an assumed event timezone.
- Offset-free timestamps require an explicit timezone context.
- Unknown timezone identifiers fail clearly.
- Daylight-saving transitions are handled when relevant to source or client timestamps.
- UTC instants are preserved.
- Employee-local business dates are derived deliberately.
- PostgreSQL session timezone does not affect the result.
- Go process timezone does not affect the result.

Inspect all uses of:

- `time.Local`
- `time.Parse`
- `time.ParseInLocation`
- `time.LoadLocation`
- pgx timestamp decoding.
- PostgreSQL `TIMESTAMPTZ` date conversion.

## C3. Required environmental matrix

Run the same attendance fixtures under:

- `TZ=UTC`
- `TZ=Asia/Singapore`
- A mixed source-event timezone fixture.

Compare persisted results, including:

- UTC instants.
- Local timestamps.
- Employment business dates.
- Durations.
- Break classifications.
- Roster assignments.
- Correction-group segments.
- Exception outcomes.

The business results must be identical wherever the represented instants and employment jurisdiction are identical.

Include:

- Singapore midnight boundaries.
- Source events carrying explicit offsets.
- Offset-free input with declared timezone.
- Mixed client timezone.
- Unknown timezone.

Do not merely rely on passing a full test suite under UTC. Provide a dedicated comparison of the same fixtures.

**Acceptance:** host, source and client timezones cannot silently change Singapore employment-business-date or paid-duration results.

---

# 5. Review Area D — Compensation immutability and historical payroll inputs

This is especially important because compensation is one of the inputs to the first real payroll calculation.

Inspect:

- `internal/hr/hr.go`
- Compensation models and persistence.
- `internal/payroll/readiness.go`
- Migration `00010`.
- Database grants and triggers.
- Compensation and readiness tests.
- Audit services.

The remediation report says compensation is insert-only and readiness checks for retroactive compensation conflicts.

Verify this directly.

## D1. Application-service behavior

Test:

1. Add compensation before the pay period.
2. Add compensation during a future pay period.
3. Add compensation affecting an already approved period.
4. Change compensation after timesheet approval.
5. Reissue an approved timesheet and change compensation.
6. Attempt a retroactive correction through the supported service.
7. Check readiness before and after each operation.

A compensation change that affects approved historical inputs must be rejected or represented as an explicit conflict.

It must never silently change a historical payroll basis.

## D2. Direct SQL as `app`

Attempt:

- UPDATE compensation amount.
- DELETE compensation.
- Change employee or employer identity.
- Change effective dates.
- Re-parent compensation to another employment-term record.
- Insert overlapping compensation for the same classification.
- Backdate a new compensation row into an approved period.

Verify the persisted state after each attempt.

Direct application-role SQL must not silently modify approved historical compensation.

## D3. Effective-date correctness

Verify that compensation dates remain within the parent employment-term interval.

Test:

- Compensation starting before employment.
- Compensation ending after employment.
- Compensation overlapping a successor employment-term interval.
- Retroactive compensation after approval.
- Future-effective compensation.

Check whether readiness distinguishes:

- Expected compensation.
- Approved inputs.
- Missing input.
- Retroactive conflict.
- Explicit historical correction.

The conflict must not be determined solely by mutable `created_at` unless that timestamp is correctly maintained and independently verified.

## D4. Monetary representation

Verify:

- No float64 money.
- API and CLI monetary input remains exact decimal text.
- Go validation rejects negative and excessive precision as specified.
- PostgreSQL `NUMERIC` preserves the supplied value.
- No implicit database rounding changes monetary input.
- Compensation values are represented consistently in audit and readiness output.

Do not implement statutory rounding.

**Acceptance:** historical compensation used by an approved period cannot silently change through service operations or direct `app` SQL.

---

# 6. Review Area E — Migration 00010 and populated-database rollback

Inspect `migrations/core/00010_m12_integrity.sql` completely, especially its Down section.

The previous migration review found that empty-database rollback and populated-database rollback can behave very differently.

This review must test the actual M1.2 schema and data.

## E1. Fresh installation

On a disposable PostgreSQL 16 database:

1. Provision roles using the actual `scripts/db/init-roles.sh`.
2. Verify role ownership and privileges.
3. Apply migrations `00001`–`00010`.
4. Seed the employer.
5. Create employee and employment terms.
6. Create attendance events and breaks.
7. Create a cross-midnight attendance interval.
8. Create a correction group with multiple dated segments.
9. Create and approve a timesheet containing the corrected evidence.
10. Verify all relationships and constraints.

## E2. Populated Down migration

On a separate disposable database, create:

- Original and corrected attendance intervals.
- Multi-segment correction groups.
- Duplicate-event dispositions.
- Published roster revisions.
- Approved and corrected timesheets.
- Attendance exceptions.
- Audit records.
- Effective-dated employment terms.
- Compensation rows.

Then execute the actual `00010` Down migration.

Determine whether it:

- Fails loudly and atomically when rollback would destroy meaningful data.
- Silently drops new records.
- Removes constraints while leaving inconsistent relationships.
- Damages previously approved artifacts.
- Recreates a schema state that is not compatible with reapplying Up.

If populated Down is unsupported, the migration must fail before destructive changes and document the restriction.

Do not claim rollback safety from an empty database.

## E3. Down/Up verification

Test:

- Clean-database Down.
- Populated-database Down.
- Down followed by Up.
- Failure partway through Down.
- Migration version tracking after failure.
- Persisted data after each attempt.

Verify whether:

- Correction-group records survive.
- Original evidence remains unchanged.
- Timesheet lineage remains coherent.
- Exception records remain available.
- Existing migration versions remain valid.

Do not make any claim about migration rollback that was not directly tested.

## E4. Migration privilege and trigger consistency

Verify:

- `migration_admin` owns migration objects.
- `app` cannot execute destructive DDL.
- SECURITY DEFINER functions have the intended owner.
- PUBLIC EXECUTE privileges are restricted as documented.
- TEMPORARY is revoked from the intended roles.
- Search paths and schema-qualified references remain safe.
- Function and trigger installation is complete after fresh migration.

**Acceptance:** migration 00010 has a verified fresh-install path and an explicit, safe populated-database rollback policy.

---

# 7. Recheck the complete application-role protections

Do not repeat the entire Prompt 10 matrix unless a newly discovered interaction requires it.

Verify the specific final invariants introduced in M1.2:

- Correction-group ownership.
- Duplicate-event disposition ownership.
- Compensation immutability.
- Attendance currency maintenance.
- Supersession identity.
- Timesheet line/sheet identity.
- Published-roster history.
- Employment-term and status history.
- Controlled employment close/transition functions.
- SECURITY DEFINER owner and search path.

Execute direct SQL as the actual `app` role.

For every rejected operation:

1. Capture the actual SQL result.
2. Re-read the affected rows.
3. Confirm no partial mutation occurred.
4. Confirm any audit record is consistent with the outcome.

Do not use superuser tests to claim application-role protection.

---

# 8. Verify the reported test evidence

The remediation report states:

- 24 Go packages.
- Seven packages with no test files.
- Full suite green.
- Full suite green under UTC.
- Race tests green across attendance, timesheet, payroll, HR and WFM.
- 38 Stage-C integrity tests.
- 16 audit tests.
- Nine concurrency tests.
- Fresh provisioning tests.
- Employer onboarding on CLI and REST.
- CLI/REST persisted-state parity.

Independently execute or inspect the evidence for these claims.

At minimum, run:

- `task lint`
- `task build`
- `go test -count=1 ./...`
- `TZ=UTC go test -count=1 ./...`
- `TZ=Asia/Singapore go test -count=1 ./...`
- Relevant `go test -race` suites.
- Database application-role integrity suites.
- Fresh-install provisioning test.
- Cross-midnight correction-group tests.
- Duplicate disposition tests.
- Compensation immutability tests.
- Populated migration Down/Up probes.
- `git diff --check`.

Use `GOWORK=off` if the unrelated parent Go workspace otherwise captures the module. Record that environment adjustment.

Verify that the test suite does not silently skip PostgreSQL tests.

Do not claim all tests passed if one environment or database-gated suite was skipped.

Reconcile the test count using an explicit method rather than repeating an unreconciled historical count.

---

# 9. Additional structural questions

Answer these questions explicitly:

1. Can a correction group contain segments for more than one employee or employer?
2. Can two correction groups supersede the same predecessor?
3. Can one segment be allocated twice to timesheets?
4. Can a correction group's segments be partially approved or rejected?
5. Can duplicate-event dispositions form chains or cycles?
6. Can a duplicate disposition hide the canonical event?
7. Can the same source instant be assigned different Singapore business dates under different host timezones?
8. Can a client timezone alter employee payroll-date attribution?
9. Can compensation be changed through SQL without an audit or readiness conflict?
10. Can migration 00010 Down silently remove meaningful correction or exception data?
11. Can an approved timesheet become inconsistent after a late break or compensation change?
12. Can the application create an employer concurrently without violating the single-employer MVP constraint?
13. Can the application role execute any unintended protected function?
14. Can a failed multi-segment correction leave orphan correction-group rows?
15. Can readiness count original and correction timesheets simultaneously?

For each question, state:

- Confirmed result.
- Source or schema evidence.
- Test performed.
- Persisted-state result.
- Remaining uncertainty.

Do not invent further features to address hypothetical future needs.

---

# 10. Review findings and severity

Use these classifications:

**P0 — Critical**

- Loss or duplication of verified attendance.
- Mutation of approved financial evidence.
- Cross-employer data exposure.
- Bypass of mandatory payroll-rule gating.
- Unrecoverable migration or correction-history corruption.

**P1 — High**

- Incorrect business-date attribution.
- Incorrect compensation history.
- Invalid correction-group ownership.
- Database application-role bypass of critical invariants.
- Fresh installation that cannot establish the intended privilege model.

**P2 — Medium**

- Bounded workflow or test gaps that do not silently corrupt payroll inputs.
- Missing operational diagnostics.
- Non-critical migration limitations.

**P3 — Low**

- Documentation inaccuracies.
- Non-critical maintainability issues.
- Future enhancements outside the current MVP scope.

Every finding must include:

- Finding ID.
- Severity.
- Exact source or migration location.
- Concrete failure scenario.
- Evidence.
- Why prior tests missed it.
- Recommended correction.
- Whether it blocks the first payroll calculation.

Keep risks and untested hypotheses separate from confirmed defects.

---

# 11. Required final report

Create:

`plans/milestone1_2_final_adversarial_review.md`

Include:

1. Git baseline and review scope.
2. Actual commands executed.
3. Correction-group integrity findings.
4. Duplicate-event provenance findings.
5. Timezone and jurisdiction-attribution findings.
6. Compensation immutability findings.
7. Migration 00010 and populated rollback findings.
8. Application-role privilege findings.
9. Regression-test verification.
10. Structural questions and their dispositions.
11. Confirmed defects, risks and untested hypotheses.
12. Clean areas.
13. Four independent verdicts:
    - Functional correctness.
    - Database integrity.
    - CLI/API operability.
    - Payroll-readiness integrity.
14. Overall acceptance verdict.
15. The smallest necessary remediation list.
16. Explicitly unverified environment-dependent items.

Do not modify any existing application files.

Keep the entire working tree uncommitted.

---

# 12. Final acceptance decision

Return one of three verdicts.

### PASS

All five review areas are independently verified, no material P0/P1 issue remains, and M1.2 provides a trustworthy operational foundation for the first restricted payroll calculation.

### PASS WITH CONDITIONS

The core evidence chain is trustworthy, but clearly bounded non-critical conditions remain. State precisely which conditions must be fulfilled before broader use.

### FAIL

Any P0 or material P1 finding remains that can lose, duplicate, misattribute or silently alter attendance, timesheet, compensation or payroll-readiness inputs.

Do not declare PASS solely because:

- The full test suite is green.
- The new migration applies.
- The Go service validates an operation.
- The application report says all findings are fixed.

The database must independently enforce the relevant integrity guarantees under `app`.

---

# 13. Scope restrictions and stop condition

Do not implement:

- CPF.
- SDL.
- SHG calculations.
- Overtime.
- Gross or net salary.
- Payslip financial generation.
- A generic rules engine.
- A Phys-Pseudo interpreter or compiler.
- RLS or SSO.
- Multi-employer authentication.
- Microservices.
- UI.
- Rotation materialization.
- Generic idempotency infrastructure.

Do not reopen the frozen architecture.

The next intended engineering milestone, after acceptance, is the first independently verified restricted Singapore monthly payroll calculation.

That milestone will have its own Phys-Research and Phys-Pseudo research gate and will use immutable inputs, independently verified fixtures and pinned rule versions.

**This review is the final integrity verification before that milestone. Be skeptical, precise and fair. Establish whether the system is ready, identify only real remaining blockers, and stop when the evidence supports a decisive verdict.**

My acceptance standard: the correction groups must preserve original evidence, timezone differences must never alter Singapore payroll dates, compensation cannot silently change after approval, and migration rollback must not unexpectedly destroy historical records.

Once those are verified, I would close the integrity phase and move directly to the first restricted payroll calculation—not initiate another general architecture review.