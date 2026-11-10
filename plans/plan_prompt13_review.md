I've reviewed `plan_prompt13.md`. My verdict: APPROVE WITH FOUR REQUIRED TECHNICAL AMENDMENTS. The plan is well structured, follows the agreed execution priorities and keeps the architecture frozen. I would not request another design cycle.

There is, however, one important omission concerning the compensation/approval race, and two areas where the proposed data model needs more precise invariants.

## 1. What I approve

The plan makes three particularly good architectural decisions.

D-RACE: shared lock and sequence. Using a transaction-scoped advisory lock followed by sequence assignment is a sound approach to ordering compensation changes against approval. It avoids treating transaction-start timestamps as commit-order evidence.&#x20;

plan_prompt13.md



Controlled compensation functions. Revoking direct application-role INSERT and routing compensation creation and correction through narrowly scoped database functions is consistent with the existing employment-history protections. It also creates a natural place to enforce audit and monetary precision.&#x20;

plan_prompt13.md



Correction-group predecessor sets. Moving from one predecessor column to a predecessor relationship table is a reasonable way to represent group-atomic corrections involving multiple dated intervals.&#x20;

plan_prompt13.md



The execution order is also appropriate: clean-server testing, D-RACE, correction groups, compensation lifecycle, duplicate concurrency, evidence timestamps and final regression.&#x20;

plan_prompt13.md



## 2. Four amendments before execution

### A. P0 — Extend the payroll-basis barrier beyond compensation

This is the most important amendment.

The plan correctly identifies that compensation transactions and approval transactions can race. But its own risk section acknowledges that employment terms and attendance have similar transaction-start timestamp comparisons, and leaves them as watch items.&#x20;

plan_prompt13.md



Consider this scenario:

1. An attendance-import transaction starts.
2. A timesheet is approved.
3. The earlier attendance transaction commits afterwards.
4. Its `created_at` predates `approved_at` because PostgreSQL `now()` represents transaction-start time.
5. Readiness fails to recognize the newly committed attendance evidence.

The same type of problem can affect employment-term changes.

Normalizing `created_at` with another trigger that uses `now()` does not solve this.

Required decision: prove that attendance, break, correction and employment-term mutations already serialize with timesheet approval through a common lock. If they do not, extend the shared payroll-input barrier to cover them.

At minimum, consider:

- `AddCompensation` and `CorrectCompensation`.
- Employment-term changes.
- Clock-event imports and late break attachment.
- Attendance corrections and correction groups.
- Timesheet submission and approval.
- Relevant employee-status changes affecting payroll population.

The objective is not necessarily to give every table another sequence column. The objective is that an input mutation and the approval that freezes its interpretation cannot cross each other unnoticed.

Also, the database approval trigger itself must acquire the shared lock before assigning `basis_seq`. Otherwise, a direct SQL UPDATE by `app` could bypass the Go service's locking step.

The sequence is reliable as a lock-serialized basis version for each employee, not as a universal database commit-order sequence. Document that distinction.

### B. P0 — Specify the correction-group relationship precisely

The proposed predecessor-set design is good, but its relationship between predecessors, correction groups and successor segments needs one more explicit definition.

Consider the initial correction:

```
Original interval O
        |
        v
Correction group G1
        |
       / \
      v   v
     A     B
```

A subsequent correction must replace both A and B, not only one.

```
A + B
  |
  v
Correction group G2
  |
  v
C + D
```

The database should enforce:

- Each predecessor belongs to at most one successor correction group.
- All current siblings of an earlier group are included in a re-correction.
- Every successor segment references a valid predecessor in the new group's predecessor set.
- Every predecessor in that set is actually superseded by the completed group.
- The group cannot be left with zero successors.
- The group cannot be appended to after its creating transaction.
- A failed correction leaves no orphan predecessor or segment rows.

There is an additional subtlety: do not accidentally retain a per-interval uniqueness restriction on `supersedes_id` that prevents one original interval from producing multiple date segments. The single-successor restriction must operate at correction-group level, while the segment mapping must still preserve individual provenance.

Finally, `Correct()` requires new bounds to cover the union of the predecessor intervals. Make sure this does not turn an unworked gap or actual unpaid break into payable time. The corrected work segments must preserve the original break and work-run boundaries.

### C. P0 — Compensation successor and overlap semantics

The plan introduces `correct_compensation` with immutable predecessors and successor records. That is appropriate, but the overlap rules need to account for legitimate correction.

Suppose an erroneous allowance is recorded as S$150.00 and its corrected value is S$125.00 for the same effective period.

The original must remain historically visible, while the successor becomes the effective record.

The database must permit the successor to supersede its predecessor without permitting an unrelated third compensation row to overlap either one.

Require explicit verification of:

- One successor per compensation record.
- Same employer, employee and compensation classification.
- No cycles.
- The predecessor is excluded from the effective compensation basis after succession.
- The successor is included.
- Other overlapping same-code records remain prohibited.
- Correction does not silently rewrite an approved period.
- Readiness identifies an unapproved compensation correction.

This is especially important because the purpose of `correct_compensation` is to solve a lifecycle problem, not merely introduce another INSERT operation.

### D. P1 — Tighten rollback and transaction-verification criteria

I approve the forward-only policy for populated M1.2 and later integrity migrations.

The `00011` Down migration should fail before destructive changes whenever populated records would lose correction, disposition or compensation-history relationships.

Also require tests for:

- Empty-database Down.
- Populated-database Down.
- Failure before destructive operations.
- Migration version after failure.
- Reapplication of Up where supported.

The plan already acknowledges that migration rollback must not silently remove meaningful history. Make the fail-before-destruction behavior a mandatory acceptance test.&#x20;

plan_prompt13.md



## 3. Smaller implementation clarifications

These do not require another plan section, but should be added to the relevant tests.

| Area                      | Clarification                                                                                                                            |
| ------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------- |
| Advisory locking          | Use exactly one shared SQL helper for the lock key. Go services, triggers and functions must not independently calculate different keys. |
| Multi-employee operations | Acquire employee locks in deterministic order.                                                                                           |
| Sequence                  | Assign after acquiring the lock; gaps from rolled-back transactions are acceptable.                                                      |
| Approval trigger          | Direct SQL approval must follow the same lock and sequence protocol as CLI/API approval.                                                 |
| Legacy records            | Keep pre-migration records at a clearly defined legacy sequence value, such as zero; do not fabricate historical commit ordering.        |
| Compensation audit        | Persist exact canonical monetary values and audit records atomically.                                                                    |
| Privilege matrix          | Enumerate relevant tables and fail when an unexpected table has no declared privilege classification.                                    |
| Clean-server testing      | Create the test database before connecting and verify concurrent role-provisioning outcomes, rather than relying on warm-server retries. |

The plan already accounts for most of these principles; the purpose here is to make their acceptance tests unambiguous.&#x20;

plan_prompt13.md



## 4. Instruction to amend the plan

# Final Review Amendments — Prompt13 / Milestone 1.3

The existing `plan_prompt13.md` is approved in principle.

Preserve its scope, five-phase engineering priority, one-migration strategy, frozen architecture and explicit deferrals.

Make the following four clarifications before execution.

## A. Shared payroll-input barrier

The planned D-RACE barrier covers compensation writes and timesheet approval.

The existing readiness design also uses transaction-start timestamps to detect some employment-term and attendance changes.

Do not leave these paths as unverified watch items.

Establish whether attendance, break, correction and employment-term operations already serialize with timesheet approval through the same employee-level lock.

If they do, prove both transaction orderings using separate PostgreSQL connections.

If they do not, extend the payroll-input barrier to cover these operations.

The barrier must cover all supported mutations capable of changing the inputs or employee population represented by an approved timesheet.

The approval trigger must independently acquire the shared lock and assign its basis sequence, including for direct SQL approval as `app`.

All Go services, triggers and database functions must invoke the same lock helper.

The sequence is a per-employee lock-serialized basis version, not a general claim about global transaction commit order.

Do not treat `created_at = now()` as a substitute for transaction serialization.

## B. Correction-group relationship invariants

Clarify the distinction between:

- Predecessor interval.
- Correction group.
- Successor segment.

A correction group may replace multiple predecessor segments and produce multiple date-segmented successor intervals.

Enforce:

- One successor correction group per predecessor.
- Every re-correction includes all current predecessor siblings.
- Every successor segment belongs to the correction group.
- Every segment's predecessor reference belongs to that group's predecessor set.
- Every predecessor is accounted for by the completed correction group.
- No orphan group or zero-successor group can commit.
- No later transaction can append segments to a completed group.
- No correction lineage can be duplicated, cycled or cross-employer/employee boundaries.

Do not accidentally retain per-interval uniqueness on `supersedes_id` if it would prevent one predecessor from producing multiple date-segmented successors. Enforce uniqueness at the appropriate group level.

Preserve actual work and break boundaries. A bounding interval must never convert an unworked gap or unpaid break into payable time.

Add direct application-role SQL tests and group-atomicity tests.

## C. Compensation successor and overlap policy

Clarify that a compensation correction successor may legitimately overlap its direct predecessor for the purpose of replacing an erroneous historical value.

The predecessor remains immutable and visible but is not part of the effective compensation basis after succession.

Enforce:

- One successor per predecessor.
- Same employer and employee.
- Same compensation classification unless an explicitly supported correction permits otherwise.
- No invalid cycles.
- No unrelated overlapping effective compensation.
- Correct effective compensation resolution.
- Audit preservation.
- Readiness conflict until an appropriate downstream correction is approved.

Test both legitimate predecessor succession and unrelated overlapping inserts as `app`.

## D. Populated migration rollback safety

The `00011` Down migration must fail before destructive changes when populated correction, disposition or compensation-successor history would be lost or made invalid.

Test:

- Empty Down.
- Populated Down.
- Pre-destruction guard.
- Failure atomicity.
- Migration-version state.
- Re-Up where the documented policy supports it.

Do not claim populated rollback safety unless it has been directly demonstrated.

## Execution

After incorporating these clarifications, execute Prompt13.

Keep migrations `00001`–`00010` byte-untouched.

Use one new forward migration `00011`.

Do not reopen the architecture or expand the jurisdiction scope.

Preserve all work and keep it uncommitted.

Stop at the Milestone 1.3 acceptance gate.

## 5. Final verdict

Plan13 direction

Approved

Execution unchanged

Amend first

Architecture redesign

Not required

Statutory payroll calculations

Deferred

My strongest recommendation is Amendment A. We should not close the compensation race while leaving attendance or employment-term changes exposed to the same transaction-start timestamp problem. Those are also payroll inputs.

After that, make correction-group and compensation-successor semantics explicit, execute the plan, and stop when the tests establish the invariants.

We should then move to the first independently verified payroll calculation. No further general architecture review is warranted unless execution uncovers a genuine structural blocker.