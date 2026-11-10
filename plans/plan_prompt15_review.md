I reviewed `plan_prompt15.md`. Verdict: APPROVE WITH SIX TECHNICAL AMENDMENTS. The plan has the right scope and dependency order. It correctly treats payroll governance and reproducibility as prerequisites rather than rushing into formulas.

I would preserve the proposed sequence:

`P3 Compensation → P4 SHG → P5 Rule activation → P2 Snapshot → P1 Run lifecycle → P6 Verification`

However, there are six details I would settle before implementation. They concern consistency, immutability and preventing privileged operations from undermining the new controls.

## 1. Six amendments

### A. P0 — Snapshot capture needs a genuinely consistent database view

The plan proposes an employer population lock followed by employee locks in sorted UUID order, then population revalidation. This is a good foundation.&#x20;

plan_prompt15.md



But acquiring locks and executing queries in one transaction does not automatically mean every query sees the same committed database state under PostgreSQL's default `READ COMMITTED` isolation.

For example, while the snapshot is capturing employees and compensation, another transaction could change a rule-pack activation or another relevant record.

Require:

- A documented transaction isolation strategy for snapshot capture.
- Consistent locking of all mutable inputs that can affect the snapshot.
- An employer-level population barrier.
- Employee locks acquired in deterministic order.
- Rule-pack activation synchronization.
- Population revalidation after acquiring the necessary locks.
- Explicit handling of serialization failures, including bounded retry where appropriate.

The crucial requirement is that the snapshot represents one coherent database view, not a collection of individually valid values observed at different times.

Also, direct SQL paths that can change the expected employee population must participate in the population barrier or be prevented by database privileges and controlled functions. Joining the barrier only in Go services is insufficient.

### B. P0 — Resolve the snapshot immutability and invalidation contradiction

The plan says:

- Every validated run owns an immutable snapshot.
- Snapshot UPDATE and DELETE are forbidden after validation.
- Drift invalidates the snapshot.
- Revalidation creates a new snapshot version.

These are compatible only if snapshot content and snapshot lifecycle are represented separately.

I recommend:

```
Payroll Run
    |
    +-- Snapshot v1 (immutable)
    |       |
    |       +-- Snapshot lifecycle events
    |               +-- VALIDATED
    |               +-- INVALIDATED
    |
    +-- Snapshot v2 (immutable)
            |
            +-- Snapshot lifecycle events
                    +-- VALIDATED
```

Use append-only lifecycle events to record validation, invalidation and supersession.

Do not update the original snapshot's contents or silently rewrite its validity history.

Define the invariant as:

A payroll run may have multiple historical immutable snapshots, but at most one currently valid validated snapshot.

The run's current snapshot reference and lifecycle transitions must themselves be protected by the database.

This avoids a conflict between immutable history and operational revalidation.

### C. P1 — Compensation resolution must preserve effective-date segments

The proposed resolver returns one effective compensation record per employee, classification and pay period.&#x20;

plan_prompt15.md



I would clarify one important boundary.

Suppose monthly basic salary changes from S$3,000 to S$3,200 halfway through a pay period.

Both records could legitimately apply to different parts of that period, despite having the same compensation classification.

The resolver should therefore establish one effective value at each point in time, rather than assume exactly one row must cover the entire payroll period.

For the first restricted profile, you may explicitly block mid-period compensation changes if their treatment is not yet implemented.

Otherwise, return a deterministic sequence of non-overlapping effective compensation segments, each with its own source record, effective dates and amount.

Never:

- Sum original and successor records together.
- Select an arbitrary record using `LIMIT 1`.
- Treat a valid historical transition as a broken chain.

The snapshot should pin the exact record IDs and applicable date intervals.

### D. P0 — Prevent direct SQL from creating fake calculation artifacts

The proposed run lifecycle correctly makes `calculated`, `approved` and `finalized` unreachable until real artifacts exist.&#x20;

plan_prompt15.md



There is one missing database-privilege condition.

The `run_calculation` artifact table has no writer in this milestone. But that does not automatically prevent `app` from inserting rows directly.

Require:

- No INSERT, UPDATE or DELETE permission for `app` on `run_calculation`.
- No direct SQL path that can fabricate a calculation artifact.
- The transition guard must require a genuine calculation artifact with valid lineage.
- An artifact-free run cannot become calculated, approved or finalized.

During Milestone 2, only `draft → validated` should be operationally reachable.

A validated snapshot is not an approved financial payroll.

Do not introduce a placeholder amount or synthetic calculation result to make the later states reachable.

### E. P0 — Rule-pack activation must verify the real evidence package

The activation function is a good governance mechanism. But a PostgreSQL function cannot independently inspect files in `rules/*` on the Go agent's filesystem.

The administrative CLI must therefore do more than pass a user-supplied manifest digest to the database function.

Require the CLI to:

1. Read the actual rule-pack manifest.
2. Validate the required evidence, formal specification, review and fixture artifacts.
3. Recompute the manifest and relevant package digests.
4. Verify the required lifecycle status.
5. Submit the exact pack identity, version and evidence references to the controlled database function.
6. Preserve the activation decision and audit record.

The database function should enforce its registered identity, version, jurisdiction, effective dates and allowed transition. It must not accept arbitrary invented pack versions.

The actual rule packages must remain in their current states unless independently qualified.

In particular, `SG-MOM-OT-001` remains BLOCKED.

Also remember that the privileged CLI operation provides controlled administrative access, not proof of a person's identity. Do not describe operator attribution as authentication.

### F. P1 — Make population and rule-pack pinning fully reproducible

The proposed employee snapshot rows include terms, status, timesheet, compensation, SHG and pack pins.&#x20;

plan_prompt15.md



I would make two additions.

First, preserve the complete evidence references.

For each employee, the snapshot should identify the effective approved timesheet and its underlying line and correction lineage. A single timesheet ID is insufficient unless all referenced records are immutable and their complete source lineage is independently recoverable.

Second, define rule-pack pins at the run level.

Use a dedicated snapshot-rule-pack relationship or equivalent structure containing:

- Pack ID.
- Exact version.
- Jurisdiction.
- Effective interval.
- Activation record ID.
- Manifest digest.
- Evidence/review references.

Employee records can reference the applicable run-level pack decisions.

The content hash must cover all meaningful snapshot inputs, using stable ordering and exact decimal strings.

Exclude volatile operational metadata such as snapshot creation time and random snapshot IDs from the content hash. Identical substantive inputs and versions should produce the same hash, even across different runs.

## 2. Decisions I approve unchanged

| Design decision                               | Verdict                                     |
| --------------------------------------------- | ------------------------------------------- |
| P3 resolver in `internal/hr`                  | Approve                                     |
| SHG modeled after compensation governance     | Approve                                     |
| `verified_none` requires evidence             | Approve                                     |
| `opt_out` blocked until semantics established | Approve                                     |
| Cross-fund SHG ambiguity blocks first profile | Approve                                     |
| Employer population lock                      | Approve                                     |
| Employee locks sorted by UUID                 | Approve                                     |
| Canonical SHA-256 content hash                | Approve                                     |
| Append-only rule-pack activation history      | Approve                                     |
| Controlled administrative activation via CLI  | Approve                                     |
| No normal REST activation endpoint            | Approve                                     |
| No fake calculated/approved/finalized states  | Approve                                     |
| One new migration `00012`                     | Approve, subject to forward-only safeguards |
| No statutory calculations                     | Approve                                     |

The decision to keep SHG cross-fund semantics unresolved rather than invent an aggregation rule is particularly important.&#x20;

plan_prompt15.md



## 3. One additional migration safeguard

Keep migrations `00001–00011` byte-identical.

Given our established forward-only policy, make `00012` fail before destructive rollback if dependent snapshots, SHG successor records or rule-pack activation history exist.

An empty-database rollback may be supported, but do not claim populated rollback safety without testing it.

This should be part of P6 rather than another separate development project.

## 4. Final instruction to amend Prompt 15

# Final Technical Amendments — Prompt15 / Milestone 2

The existing `plan_prompt15.md` is approved in principle.

Preserve its six prerequisites, dependency order, one-migration strategy, frozen architecture and explicit scope restrictions.

Incorporate the following six amendments before execution.

## A. Consistent snapshot isolation

The employer population lock and employee pay-basis locks are necessary but must be combined with a documented consistent database-view strategy.

Define the transaction isolation level and locking strategy for snapshot capture.

Ensure all operations capable of changing:

- Employee population.
- Employment terms or status.
- Compensation.
- Timesheet approval.
- SHG instructions.
- Rule-pack activation.

participate in the appropriate synchronization mechanism.

Direct SQL must not bypass population stability or snapshot input protection.

Use deterministic lock ordering.

Handle serialization failures explicitly, with bounded retries where necessary.

Do not assume that one transaction under READ COMMITTED automatically produces a consistent multi-statement snapshot.

## B. Append-only snapshot lifecycle

Separate immutable snapshot content from snapshot lifecycle state.

Allow a run to retain multiple immutable historical snapshot versions, but enforce at most one currently valid validated snapshot.

Represent validation, invalidation and supersession through append-only lifecycle events.

Never rewrite the previous snapshot's contents to accommodate changed source records.

Revalidation creates a new snapshot version.

Protect the run's active-snapshot relationship and lifecycle transitions at database level.

## C. Effective compensation over time

Define compensation resolution as one effective value per employee/classification/time interval.

A valid compensation change during a pay period may produce multiple non-overlapping effective segments.

Return all applicable segments with their exact source references and effective dates.

For the first restricted payroll profile, explicitly block mid-period changes if their calculation treatment is not implemented.

Do not:

- Add predecessors and successors together.
- Use arbitrary `LIMIT 1`.
- Treat valid non-overlapping historical transitions as broken chains.

Snapshot capture must pin the exact effective record IDs and intervals.

## D. Calculation-artifact protection

Revoke `app` INSERT/UPDATE/DELETE on `run_calculation` while no calculation writer exists.

The database must reject direct SQL attempts to fabricate calculation artifacts.

Only `draft → validated` is reachable during Milestone 2.

`calculated`, `approved` and `finalized` remain unreachable without a genuine calculation artifact and appropriate review evidence.

A validated snapshot is not an approved financial payroll.

## E. Actual rule-package verification

The administrative CLI must inspect the actual on-disk rule package before activation.

Recompute relevant digests and verify:

- Manifest identity.
- Evidence references.
- Formal specification status.
- Independent review status.
- Required fixture/verification status.
- Exact version and effective interval.

The database activation function must validate the registered package identity and enforce the permitted transition.

Do not trust a caller-supplied version string or digest alone.

Do not expose activation through the normal REST application.

No real statutory pack may be activated in this milestone unless its research and verification lifecycle independently qualifies it.

`SG-MOM-OT-001` remains BLOCKED.

## F. Complete snapshot provenance and deterministic hashing

Ensure the snapshot retains:

- Effective approved timesheet identity.
- Timesheet line references.
- Attendance/correction lineage necessary for reproducibility.
- Effective employment and status references.
- Exact compensation records and intervals.
- SHG instruction/evidence references.
- Run-level exact rule-pack version pins.
- Inclusion/exclusion decisions and blockers.

Define canonical hashing over substantive immutable content with stable ordering.

Include snapshot schema version, jurisdiction, period, exact monetary strings and all relevant source/version references.

Exclude random IDs and volatile timestamps from the content hash so identical substantive inputs and versions produce identical hashes.

## Migration and execution

Keep migrations `00001–00011` byte-identical.

Migration `00012` must have a guarded Down policy that refuses before destructive changes whenever dependent snapshot, SHG or activation history exists.

After incorporating these clarifications, execute Prompt15.

Do not initiate another architecture review.

Do not introduce statutory payroll calculations.

Preserve the working application and keep all changes uncommitted.

Stop at the Milestone 2 acceptance gate.

## 5. Final verdict

Plan15 architecture and scope

Approved

Execution as submitted

Amend first

Architecture redesign

Not required

Statutory calculations

Deferred

One final point: do not let the snapshot implementation become a generic data-warehouse project. We need enough immutable records to reconstruct one payroll decision, not a universal historical-data platform.

Once these amendments are incorporated, I would authorize execution without another planning cycle. The objective is to make one payroll run governable and reproducible, then finally move on to the first independently verified Singapore payroll calculation.