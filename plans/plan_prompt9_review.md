I reviewed `plan_prompt9.md` against the consolidated adversarial report and the original Prompt 9 requirements.

Verdict: APPROVE THE PLAN WITH SIX AMENDMENTS BEFORE EXECUTION.

This is a good remediation plan. It correctly avoids another architecture redesign, keeps migrations append-only, preserves the single-application architecture, and prioritizes the attendance-to-timesheet evidence chain.&#x20;

plan_prompt9.md



However, I found several places where the proposed fixes could introduce new integrity problems if implemented literally.

## 1. Six amendments I consider necessary

### A. Preserve immutability when correcting timesheets

P0 — Important design correction

The plan proposes:

- Rejected timesheet: `rejected → draft`.
- Approved timesheet: `approved → superseded`.
- Create a new correction timesheet.

There is a contradiction: approved and rejected records are supposed to be immutable, but these transitions modify their status.

I recommend a cleaner model:

1. Keep the original approved or rejected timesheet completely unchanged.
2. Create a new draft correction version.
3. Link it to the original through `supersedes_id`.
4. Preserve the original approval or rejection, contents, timestamps and evidence.
5. When the correction is approved, resolve which version is currently effective through the supersession chain.
6. Enforce at most one successor for each timesheet.

This also avoids special trigger exemptions that could gradually weaken the freeze protections.

An approved financial record should never need to be rewritten to explain that a subsequent correction exists.

### B. Do not use the 72-hour rule to block verified attendance

P0 — Important semantic correction

The plan proposes wiring monthly `NeedsReview` into `checkEligible` as a skip condition.&#x20;

plan_prompt9(1).md



We must distinguish:

- Unpaired or ambiguous clock events.
- Implausibly long clock spans.
- Verified actual work exceeding the statutory overtime-hours limit.

The first two can prevent an attendance interval from being approved until investigated.

The third is a separate statutory compliance issue. It must not cause otherwise verified actual working time to disappear from a timesheet.

The 72-hour limit must not:

- Truncate actual hours.
- Automatically reduce payable hours.
- Turn verified attendance into missing attendance.
- Make the underlying work unrecordable.

Keep the 72-hour condition as an independently visible compliance exception until its legal consequences are verified.

### C. A 24-hour pairing window is not enough

The proposed 24-hour limit is a useful safety ceiling, but it does not establish that two clock events belong to the same shift.

plan_prompt9(1).md

Require the pairing algorithm to consider:

- Employee identity.
- Event source.
- Roster assignment.
- Expected shift start/end.
- Cross-midnight allowance.
- Multiple shifts on the same date.
- Duplicate and conflicting events.

A clock-out should not pair with an unrelated unmatched clock-in merely because they occur within 24 hours.

Where no reliable roster association exists, preserve the events and flag the ambiguity.

For our restricted MVP, roster-anchored pairing should be the normal path. An absolute maximum duration should be a second defensive check, not the primary matching rule.

### D. Make imports, corrections and generation recoverable and atomic

The plan addresses concurrency with advisory locks, but it needs an explicit crash-recovery contract.

plan_prompt9(1).md

Consider this failure:

1. Raw clock events are inserted.
2. The process crashes before pairing.
3. The user retries the same import.
4. Every event is identified as a duplicate.
5. No pairing occurs because the retry sees no newly inserted events.

The result is permanently unprocessed attendance evidence.

Require:

- Import and pairing to operate transactionally at the appropriate employee/batch boundary.
- Replays to discover and process previously persisted but unpaired events.
- Exceptions to remain discoverable on replay.
- Corrections and aggregate updates to commit atomically.
- Timesheet eligibility to be rechecked inside the generation transaction.
- Partial generation behavior to be explicitly defined.

Also define the advisory-lock convention. All relevant operations must use the same lock key and transaction-scoped locking mechanism. Use deterministic ordering when acquiring multiple employee locks to prevent deadlocks.

### E. Make payroll readiness population explicit

The plan proposes a union of employees holding approved sheets and a retroactive-change blocker. That is a good starting point.&#x20;

plan_prompt9(1).md



But the expected payroll population cannot be derived solely from employees who already have timesheets.

Otherwise, an employee who should have been paid but has no timesheet could disappear.

Require three distinct sets:

1. Expected population: employees whose effective-dated employment arrangements fall within the pay period.
2. Submitted payroll evidence: employees with approved timesheets for that period.
3. Exceptions: missing timesheets, incomplete inputs, explicit exclusions and historical-term conflicts.

Also require validation that a referenced payroll run belongs to the same employer and pay period before applying its exclusions.

An empty population must not pass the approval predicate through vacuous truth.

### F. Close the audit and financial precision gaps

The plan addresses exact money representation, but it does not make audit persistence sufficiently explicit.

Add mandatory audit-event writes for:

- Employment-term and status changes.
- Roster publication and revision.
- Clock-event import and conflicting replay.
- Attendance corrections.
- Timesheet submission, rejection and approval.
- Payroll exclusions.

Require real operator attribution for audited writes. `api` and `cli` defaults should not silently replace the actual operator identity.

For money, apply the exact representation consistently across the whole path:

`JSON decimal string → Go int64 minor units → PostgreSQL NUMERIC(12,2)`

Remove `float64` from monetary domain models and avoid silently rounding invalid inputs to two decimal places. This includes basic salary, allowances and compensation amounts.

## 2. Additional requirements the plan should explicitly include

These are smaller omissions or details that deserve to be carried into the implementation instructions.

| Area              | Additional requirement                                                                                                                                                                   |
| ----------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Attendance schema | Define the exact current/live interval predicate used by the partial overlap constraint. A correction successor must not accidentally be excluded just because it has a `supersedes_id`. |
| Break evidence    | Unknown or unclassifiable break events remain unpaired and persist their exceptions.                                                                                                     |
| Correction        | Corrections crossing a calendar date must resolve the correct roster assignment and affected daily records.                                                                              |
| Roster            | Revalidate week boundaries, current employee status, membership, skills and rest-day compatibility before publication.                                                                   |
| Seed              | Explicitly test `seed demo --reset` using a controlled administrative connection. Never bypass immutable-record triggers or grant broad runtime DELETE permissions.                      |
| Migration         | Preserve existing databases; create a fresh database for testing. Do not discard any existing local data under the pre-release-reset decision without approval.                          |
| Tests             | Resolve fresh-database parallel setup races, but only tolerate expected duplicate-creation races after verifying the existing object is correct.                                         |
| API               | Make sure the interval listing endpoint exposes enough information for an operator to correct attendance without SQL.                                                                    |
| Readiness         | Reject missing, nonexistent or wrong-period payroll-run references rather than silently ignoring exclusions.                                                                             |

One other structural detail: the partial GiST exclusion constraint must distinguish current intervals from historical superseded intervals. That distinction should be explicit in the schema, not inferred from `supersedes_id IS NULL` unless that condition genuinely identifies all current records.

# 3. Decisions in plan9 that I approve unchanged

These are good decisions and should remain frozen.

| Decision                                                      | Verdict                                         |
| ------------------------------------------------------------- | ----------------------------------------------- |
| New migrations `00009+`; freeze `00001–00008`                 | Approve                                         |
| Check existing databases before deciding upgrade/reset policy | Approve                                         |
| Fresh-install role preflight                                  | Approve                                         |
| `migration_admin` for migrations; `app` for runtime           | Approve                                         |
| Integer minor units for monetary handling                     | Approve                                         |
| No new decimal dependency without necessity                   | Approve                                         |
| Database-backed uniqueness and overlap protection             | Approve                                         |
| Advisory locks for concurrent operations                      | Approve, with the additional requirements above |
| Domain-specific idempotency instead of generic framework      | Approve                                         |
| No rotation materialization in this pass                      | Approve                                         |
| No statutory payroll calculations                             | Approve                                         |
| No architectural redesign                                     | Approve                                         |

The rotation decision is particularly sensible. Generating rosters from rotations is additional behavior, not necessary to correct the current integrity defects.&#x20;

plan_prompt9(1).md



# 4. Consolidated instruction to amend plan9

Send this to the agent before execution.

# Plan9 Review Amendment — Final Clarifications Before Milestone 1.1 Execution

The overall plan is approved. Preserve its task sequence, scope, frozen architecture and migration decisions.

Make the following targeted amendments before execution. Do not initiate another architecture redesign.

## A. Immutable timesheet correction

Do not mutate an approved timesheet from `approved` to `superseded` if the database contract says approved records are immutable.

Likewise, do not mutate the original rejected timesheet back to draft if rejected artifacts are intended to remain immutable.

Prefer append-only correction versions:

- Original approved/rejected timesheet remains unchanged.
- New draft correction timesheet links to the original through `supersedes_id`.
- Approval of the correction establishes the effective successor.
- Payroll readiness resolves the effective version from the supersession chain.
- Enforce one successor per original timesheet.
- Preserve all historical snapshots and decisions.

If there is a compelling reason to retain lifecycle status transitions on an existing row, document exactly which fields remain immutable and prove that the trigger cannot permit content mutation during a status-only transition.

Do not weaken the general freeze trigger.

## B. Separate the 72-hour compliance rule from attendance eligibility

A monthly overtime-limit breach must not truncate or discard verified actual working time.

The `checkEligible` skip condition should concern unpaired, ambiguous or otherwise unreliable attendance evidence.

Keep the 72-hour condition as a separate compliance exception. Do not treat a verified period of work as missing merely because its monthly overtime hours exceed the statutory limit.

No 72-hour payable-hours cap is permitted.

## C. Pair clock events using the roster, not only a 24-hour window

The 24-hour ceiling may remain as a defensive bound, but it is insufficient as the primary pairing rule.

Require matching based on:

- Employee.
- Source/event stream.
- Applicable roster assignment.
- Expected shift window.
- Cross-midnight behavior.
- Multiple shifts/day.
- Event ordering and duplicate/conflict handling.

Unrostered or ambiguously attributable events must remain explicit exceptions unless an independently documented manual-assignment process supports them.

Preserve all original events.

## D. Import/correction/generation atomicity and replay recovery

Add explicit acceptance tests for:

1. Clock events stored but pairing interrupted by process failure.
2. Identical import replay after partial processing.
3. Reprocessing previously stored unpaired events.
4. Persisted exception retrieval after replay.
5. Correction transaction failure after interval insertion but before aggregate update.
6. Timesheet generation racing with a new attendance import.
7. Repeated and concurrent generation for the same employee and pay period.

Related evidence and aggregate changes must be atomic at the chosen transaction boundary.

Use consistent transaction-scoped advisory locks where necessary. Define a stable lock-key convention and deterministic acquisition order.

## E. Define payroll population and exclusion semantics

Readiness must derive its expected employee population from effective-dated employment for the assessed period, not just employees who already have approved timesheets.

Report separately:

- Expected employees.
- Employees with approved timesheets.
- Employees missing timesheets.
- Employees blocked by incomplete inputs.
- Explicitly excluded employees.
- Employees with historical employment-term conflicts.

Validate any referenced payroll run by employer, pay period and compatible status before applying its exclusions.

An empty population must not pass payroll approval.

## F. Audit and exact monetary representation

Persist audit events for all consequential changes, including employment terms/status, roster publication/revision, clock imports, attendance corrections, timesheet transitions and payroll exclusions.

Do not silently substitute generic `api` or `cli` labels where actual operator attribution is required.

Apply exact monetary representation end to end:

- JSON decimal strings.
- Go int64 minor units.
- PostgreSQL NUMERIC(12,2).
- Explicit validation of negative and excess-precision input.
- No float64 monetary calculations.

## G. Additional safeguards

Ensure:

- Current versus superseded attendance intervals have an explicit, consistent database predicate.
- A successor interval can remain current while its predecessor is historical.
- Cross-date corrections resolve the correct roster assignment and affected attendance records.
- Roster publication revalidates date range, employee status, membership, skills and rest-day compatibility.
- `seed demo --reset` is tested against a disposable database using a controlled administrative operation.
- No broad DELETE permissions or trigger-bypass mechanisms are granted to the runtime app role.
- Existing databases are never dropped or overwritten without approval.
- Parallel database setup only tolerates verified expected creation races.
- API interval retrieval exposes enough information to complete correction without SQL.

## Execution

After incorporating these clarifications, execute plan9.

Preserve the frozen architecture and current working tree.

Do not add rotation materialization, statutory payroll calculations, a rules engine, RLS, microservices or a UI.

Produce the structural audit and remediation report specified by plan9.

Stop at the Milestone 1.1 acceptance gate.

## 5. Final assessment

Plan9 overall direction

Approved

Execution unchanged

Amend first

Architecture redesign

Not required

Statutory payroll calculations

Remain deferred

The newer review identified 17 clean areas, which is encouraging: the problem is concentrated in temporal integrity, workflow transitions, attendance lineage and runtime database-role enforcement—not a fundamentally broken architecture.&#x20;

milestone1_adversarial_review\.md



My recommendation: amend plan9 with these clarifications and execute it. The next success criterion is not a larger feature set. It is proving that clock evidence cannot be lost, duplicated or misattributed, and that an approved timesheet remains historically trustworthy even when the underlying attendance later needs correction.