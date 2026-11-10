I reviewed `plan_prompt11.md`. The execution plan is well structured and substantially follows the agreed remediation priorities. I approve its overall sequence, but recommend six technical amendments before authorizing execution.

I would not send the agent back for another broad design exercise. The plan already has a sensible dependency order: timezone handling before temporal fixes, chain-aware correction before late-break recovery, database protections after service behavior is settled, and installation/onboarding after the role model is corrected.&#x20;

plan_prompt11.md



There are, however, a few important interactions that need resolving before code changes begin.

## 1. Six amendments

### A. P0 — Cross-midnight correction conflicts with single-successor integrity

This is the most important design issue I found.

The plan proposes:

- Reusing `splitRun` in `Correct`.
- Returning multiple successor interval IDs.
- Creating one successor per local business-date segment.

But the existing attendance model enforces one successor per original interval.

Suppose an original attendance interval is corrected to span two business dates. `Correct` now needs to create two dated successor intervals from one predecessor.

That conflicts with the existing single-successor invariant.

The agent must explicitly solve this before implementing B6.

My preferred design is a correction group:

- One correction operation identifies the original interval.
- The operation creates one correction group.
- That group owns multiple dated successor segments.
- The predecessor is superseded by the group, not independently by multiple competing corrections.
- Each successor segment retains its own local date, roster attribution, duration and source lineage.
- All segments are approved or rejected as one correction operation.
- A second correction cannot create another competing group.

An equivalent relational design is acceptable, but simply weakening the single-successor constraint is not.

The design must preserve both one-successor-per-correction and one-to-many date segmentation.

### B. P0 — Separate source timezone, employment timezone and client timezone

B5 proposes threading the declared event timezone into `classifyBreak`. That is an improvement, but we need three separate concepts:

| Timezone                     | Responsibility                                                                        |
| ---------------------------- | ------------------------------------------------------------------------------------- |
| Source/event timezone        | Interpret timestamps originating from an external clock source.                       |
| Employee employment timezone | Determine local business dates used for attendance and Singapore payroll attribution. |
| Client/service timezone      | Determine when the BPO client expects service coverage.                               |

For the current Singapore profile, the employment timezone is `Asia/Singapore`.

An external clock device or vendor might report a timestamp with another timezone. That timezone should help resolve the original instant, but it must not automatically change the employee's Singapore payroll date.

The tests must prove that the same instant produces the same employment-business-date result regardless of the host operating system timezone or client timezone.

Also, explicitly handle timestamps that already carry a UTC offset. Do not apply an event timezone twice.

### C. P1 — Duplicate disposition needs a canonical event relationship

The proposed `clock_event_disposition` table records:

`event_id, verdict, reason, operator, decided_at`

That is a good start, but `duplicate_confirmed` needs to identify what the event duplicates.

Otherwise, the system knows that an event was discarded from pairing, but cannot conclusively establish which canonical event it corresponds to.

Add a relationship such as:

`duplicate_of_event_id`

Validate that the original and duplicate:

- Belong to the same employer and employee.
- Have compatible event types.
- Represent the same underlying occurrence under an explicitly defined matching rule.
- Do not create circular duplicate references.
- Reference a canonical event that is itself not dispositioned as a duplicate.

Retain both raw records and the operator's decision.

This makes the duplicate disposition useful as audit evidence, not merely an exception-clearing flag.

### D. P0 — SECURITY DEFINER protection must include PUBLIC privileges

C2 correctly identifies the `pg_temp` shadowing vulnerability.

However, the privilege fix needs precise PostgreSQL handling.

Temporary-table privileges are commonly granted to `PUBLIC` by default. Revoking TEMPORARY only from `app` does not necessarily remove its inherited effective permission.

Require:

- Schema-qualified protected table references, including `public.attendance_interval`.
- A restricted and explicitly defined function `search_path`.
- Appropriate TEMPORARY revocation from both `app` and `PUBLIC`.
- No unnecessary CREATE privileges on protected schemas.
- Verification of function owner and execution privileges.
- Direct `pg_temp` exploit tests as `app`.

Do not rely on `search_path=public` alone.

### E. P0 — C8 must prevent direct SQL from bypassing HR transitions

The plan permits tightly controlled updates to open employment-term and status-history records, particularly closing their effective periods.

But there is an important distinction between permitting an application service to close an open record and permitting the `app` role to execute an arbitrary UPDATE that happens to pass a trigger.

The application role is explicitly within our database threat model.

For these records, prefer:

1. Revoke unrestricted UPDATE/DELETE.
2. Provide narrowly scoped database functions for legitimate predecessor closure, where required.
3. Validate the new effective period and transition within those functions.
4. Record audit evidence in the same transaction.
5. Expose only the required execution privileges to `app`.

The function must not accept arbitrary caller-controlled table names, employee identities or mutable session variables as authorization.

Apply the same reasoning to compensation and published roster history.

### F. P1 — Do not round actual attendance to whole minutes prematurely

The plan proposes integer-minute quantization using round-half-up.

I would change this.

Actual attendance timestamps can preserve precision to seconds or database-supported microseconds. Rounding all measured durations to whole minutes during ingestion introduces an additional business assumption that we have not established as universally applicable.

My recommendation:

- Preserve actual elapsed duration as integer seconds or microseconds.
- Do not discard sub-minute information during attendance processing.
- Apply payroll-specific rounding only at an explicitly verified calculation boundary.
- Keep any legally prescribed rounding convention inside the relevant jurisdiction rule package.

This distinction is important: exact attendance evidence should not be altered just because a future payroll calculation might use minute-based reporting.

## 2. Three additional safeguards

These are smaller, but worth including now.

Employer creation concurrency: the single-employer MVP restriction should handle concurrent `employer create` requests. Two simultaneous requests must not create two employers simply because both observed an empty database. Use a transaction lock or equivalent enforcement and test the race.

Runtime role identity: D2 should verify `current_user = 'app'`, not merely reject `is_superuser = true`. A different non-superuser role may still have more privileges than intended.

Database recycling: `--confirm-db` matching the DSN target is not by itself a sufficient disposable-database safeguard. Add an explicit disposable database naming policy or allowlist. Verify both administrator and migration connections target the intended database and server before any destructive operation.

## 3. What I approve unchanged

| Plan decision                                            | Verdict                         |
| -------------------------------------------------------- | ------------------------------- |
| B5 timezone work before temporal fixtures                | Approve                         |
| B3 chain-aware correction before B7                      | Approve                         |
| B6 correcting both affected business dates               | Approve, subject to Amendment A |
| New `00010` migration                                    | Approve                         |
| `00001`–`00009` remain untouched                         | Approve                         |
| Domain-specific duplicate disposition                    | Approve                         |
| Employee status validation at assignment and publication | Approve                         |
| Two-role database provisioning                           | Approve                         |
| SQL-free employer creation                               | Approve                         |
| Audit events in business transactions                    | Approve                         |
| No rotation materialization                              | Approve                         |
| No generic idempotency framework                         | Approve                         |
| No statutory calculations                                | Approve                         |

The proposed E2E harness under `tests/e2e/` is particularly good. It makes the CLI/REST persisted-state comparison repeatable rather than relying on an uncommitted temporary script.&#x20;

plan_prompt11.md



## 4. Instruction to amend the execution plan

# Final Plan Review Amendments — Prompt11 / Milestone 1.2

The existing `plan_prompt11.md` is approved in principle.

Preserve its scope, task sequence, frozen architecture, migration approach and acceptance criteria.

Make these targeted corrections before execution:

## A. Cross-midnight attendance correction model

B6 proposes multiple dated successor intervals but the existing attendance correction model enforces one successor per original interval.

Do not weaken that invariant.

Define a correction-group model before implementing B6:

- One original interval is superseded by one correction group.
- The correction group contains one or more dated successor segments.
- Each segment has its own local business date, roster assignment, duration and source lineage.
- One correction group is the unique successor of the original.
- A second correction group cannot supersede the same original.
- Approval and rejection apply coherently to the correction operation.

An equivalent relational implementation is acceptable if it preserves these properties.

Add direct PostgreSQL tests for:

- One predecessor producing two dated segments.
- Concurrent correction attempts.
- Duplicate correction groups.
- Partial segment insertion failure.
- Employee and pay-period consistency across the group.

## B. Explicit timezone responsibilities

Distinguish:

1. Source/event timezone for interpreting external timestamps.
2. Employee employment timezone for local business-date and payroll attribution.
3. Client/service timezone for BPO operating schedules.

For the current Singapore employment profile, use `Asia/Singapore` for employment-business-date attribution.

Do not allow a vendor timezone or client timezone to change Singapore payroll dates.

Preserve offset-aware source timestamps without applying a timezone conversion twice.

Test identical event instants under:

- `TZ=UTC`
- `TZ=Asia/Singapore`
- A mixed source-event timezone fixture.

The resulting Singapore business dates and durations must be identical.

## C. Duplicate event disposition

Extend B1's `clock_event_disposition` to identify the canonical event being duplicated.

Use an explicit relation such as `duplicate_of_event_id`.

Enforce:

- Same employer.
- Same employee.
- Compatible event kind.
- No self-reference.
- No circular duplicate references.
- Canonical events cannot themselves be dispositioned as duplicates.

Preserve both raw events, operator, reason, original exception and disposition history.

Only the duplicate event is excluded from pairing and timesheet eligibility.

The canonical evidence remains usable.

## D. SECURITY DEFINER privilege hardening

C2 must revoke TEMPORARY privileges from both the runtime app role and PUBLIC, subject to the dedicated database's requirements.

Schema-qualify every protected table reference.

Use a restricted function search path.

Verify:

- Owner.
- SECURITY DEFINER.
- Search path.
- Function references.
- Execute privileges.
- PUBLIC and app database privileges.
- Direct temporary-table exploit behavior.

Do not claim protection based solely on `SET search_path=public`.

## E. Controlled HR history mutation

For C8, do not rely solely on triggers that permit arbitrary app-role UPDATE operations on open employment records.

Prefer restricted database functions for:

- Closing an open employment-term interval.
- Closing an open employee-status interval.

Revoke unrestricted direct UPDATE/DELETE privileges from `app` where the existing application model permits it.

Controlled operations must validate:

- Employer and employee identity.
- Existing row state.
- Effective-date ordering.
- Non-overlap.
- Approved-period conflicts.
- Audit event atomicity.

Compensation, membership, skills and published roster history remain protected from arbitrary UPDATE/DELETE.

## F. Preserve actual duration precision

Replace E3's whole-minute round-half-up quantization for actual attendance.

Preserve measured elapsed duration as integer seconds or microseconds.

Do not round actual work or break evidence during import, pairing or correction.

Any payroll rounding must be performed later at an explicit, independently verified jurisdiction-rule boundary.

Add sub-minute duration and aggregation regression tests.

## G. Additional safeguards

- Enforce the single-employer bootstrap restriction under concurrent employer creation.
- Require exact runtime `current_user = 'app'`; rejecting superuser alone is insufficient.
- Require both administrator and migration connections to target the same intended database/server before recycling.
- Add a disposable database naming policy or allowlist. `--confirm-db` matching the target name is necessary but not sufficient by itself.

## Execution

After making these amendments, execute the existing plan.

Do not initiate another architecture review.

Keep the single Go application and PostgreSQL architecture unchanged.

Preserve all existing work and leave changes uncommitted.

Stop at the Milestone 1.2 acceptance gate.

## 5. Final verdict

Plan direction

Approved

Execution as currently written

Amend first

Architecture redesign

Not required

Payroll calculations

Remain deferred

The most important amendment is the relationship between cross-midnight correction and single-successor integrity. Resolve that before coding because it affects the underlying data model, not just a function signature.

The second is preserving exact attendance duration rather than introducing unverified rounding.

Once these amendments are incorporated, I would authorize Prompt11 execution without another planning cycle. The objective should now be to complete the integrity work, verify it against real PostgreSQL, and move on to the first independently verified payroll calculation.