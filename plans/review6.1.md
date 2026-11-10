I reviewed `plan_prompt17.1.md`. This is a substantial improvement, and I approve its overall direction and Gate 1 approach. However, I would make two final technical amendments before authorizing execution beyond that gate.

The most important improvement is that the plan no longer treats `pg_locks` as proof of pre-BEGIN ordering. It records the PostgreSQL 16 experiment, introduces an XID-based admission ticket and makes the snapshot-authority decision an explicit architecture gate.&#x20;

plan_prompt17.1.md



There are two remaining problems worth resolving carefully.

# 1. D1: improve the XID admission-ticket proof

The plan currently proposes:

> Capture asserts that the caller's snapshot `xmin` is newer than the ticket XID.

This is a conservative check, but it has a weakness.

Consider an unrelated database transaction that began a long time ago and remains open. It can hold the global snapshot `xmin` behind the ticket's transaction ID, even though the ticket was successfully committed before the snapshot began.

The result would be a false rejection.

PostgreSQL 16 offers a more precise mechanism: `pg_visible_in_snapshot(xid8, pg_snapshot)`. It tests whether a particular transaction ID is visible according to a particular snapshot. That is more appropriate than comparing the global `xmin` with a ticket XID.&#x20;

[image](https://www.google.com/s2/favicons?domain=https://www.postgresql.org\&sz=32)

postgresql.org+1



I recommend testing:

```
SELECT pg_visible_in_snapshot(
    :ticket_xid,
    pg_current_snapshot()
);
```

Also verify that the ticket record itself is visible to the snapshot.

The intended protocol becomes:

1. Acquire the session-level barriers in a separate READ COMMITTED operation.
2. Record a server-generated ticket and its top-level transaction XID.
3. Commit that ticket operation.
4. Begin REPEATABLE READ on the same physical connection.
5. Verify that the ticket is visible in the transaction's snapshot.
6. Verify that the required barriers remain held.
7. Perform capture.

If a ticket is created after the REPEATABLE READ snapshot, it must not be visible to that snapshot.

This should be tested on PostgreSQL 16 with both a stale-snapshot attack and an unrelated long-running transaction. The documented semantics of REPEATABLE READ make the distinction important: its view is taken at the first non-transaction-control statement, not refreshed when a later lock becomes available.&#x20;

[image](https://www.google.com/s2/favicons?domain=https://www.postgresql.org\&sz=32)

postgresql.org+1



### An additional lock-ownership attack

There is a second subtlety.

Your plan reports that session-level and transaction-level advisory locks produce identical `pg_locks` rows in the tested PostgreSQL 16.15 environment.

That means a caller could potentially:

1. Acquire a valid session-level ticket.
2. Begin REPEATABLE READ.
3. Release the session-level lock.
4. Acquire a transaction-level lock using the same key.
5. Invoke capture while `pg_locks` still shows the key as granted.

The original ticket would still exist, but the lock would no longer provide the intended session-level protection.

The plan must prevent an ordinary `app` caller from manipulating the underlying advisory locks directly.

I recommend testing whether direct execution privileges on PostgreSQL's advisory-lock and unlock functions can be revoked from `PUBLIC`/`app`, while granting the required operations only to the narrowly controlled SECURITY DEFINER wrappers.

All application and trigger lock operations would then use those wrappers and the existing canonical lock-key helpers.

If this cannot be safely enforced under the two-role model, the agent must stop at Gate 1 and request approval for the smallest privilege-boundary change.

The exact stale-RR attack and the lock-swap attack must both fail.

# 2. Snapshot authority: the architecture decision must be real

The plan defaults to Option 2: Go remains the canonical resolver and a restricted snapshot-writing boundary persists its results.

I agree that this is a promising direction, particularly because we have already developed `internal/compbasis/resolve.go`.

But we must avoid recreating the original problem with two independently maintained definitions of compensation.

The agent needs to answer one question:

How does the database know that the snapshot created by the Go resolver is complete and authentic without independently reimplementing the same compensation-resolution algorithm?

There are two credible choices.

Option 1 — PostgreSQL is the canonical resolver

Move effective compensation resolution into one authoritative SQL function or view. Go consumes the same implementation.

The database derives the snapshot and its canonical hash.

One database authority

Option 2 — Go is the canonical resolver

Keep the compensation resolver in Go. Introduce a genuinely restricted snapshot-writing capability that ordinary `app` SQL cannot invoke to fabricate data.

The database verifies source identity, ownership, effective dates, lineage, population completeness and persisted integrity, without maintaining a second compensation-selection algorithm.

One Go business resolver

I prefer Option 2, consistent with your existing Go architecture.

However, it must define a real trust boundary. If ordinary `app` credentials can simply call the snapshot writer with arbitrary JSON, Option 2 has not solved the problem.

A narrowly privileged snapshot executor might be necessary. That would be a privilege-boundary change and must come back for approval rather than being introduced silently.

The canonical hash also needs exactly one implementation. The database can verify the digest of the exact canonical bytes supplied by the trusted snapshot writer; it must not maintain a separate, drifting serializer.

Do not implement both resolvers.

# 3. Other parts of plan17.1 that I approve

The following revisions are appropriate:

- P1a population semantics are decided before population enumeration.
- The authority ADR precedes capture implementation.
- Gate 1 prevents progress past an unproven snapshot-enforcement mechanism.
- The report explicitly distinguishes the recorded checksums for `00001–00011` from a newly calculated baseline hash for `00012`.
- The plan restores the complete writer-race classes.
- The 20-section remediation report is retained.
- One forward migration, `00013`, is preserved.
- The drop-first test strategy is restricted to verified disposable databases.
- `OT-001` remains BLOCKED.
- No financial calculations are introduced.&#x20;

  plan_prompt17.1.md



I would add only three small clarifications:

1. Employee enumeration must occur in a fresh READ COMMITTED statement after the employer population barrier has been acquired, before acquiring the individual employee locks.
2. Lock cleanup must cover success, rollback, cancellation and panic. A connection with uncertain session-lock state must be discarded, never returned as healthy to the pool.
3. Change `Open questions: None` to explicitly state that the D1 mechanism and snapshot-authority decision remain unresolved until Gate 1.&#x20;

   plan_prompt17.1.md



# 4. Consolidated final amendment for the agent

# Final Technical Amendment — Plan Prompt17.1

# D1 Admission Proof and Snapshot Authority

The execution plan `plan_prompt17.1.md` is approved in principle.

Preserve its scope, invariant-first principle, 12-step execution order, single migration `00013`, historical migrations `00001–00012` untouched, two-role model unless an approved privilege change is necessary, and Gate 1 stop condition.

Apply these final amendments.

## A. Improve XID-ticket visibility proof

The proposed `snapshot xmin > ticket XID` comparison can falsely reject a legitimate ticket when an unrelated long-running transaction holds the global snapshot xmin behind the ticket.

Investigate using:

`pg_visible_in_snapshot(ticket_xid, pg_current_snapshot())`

and direct visibility of the committed ticket record inside the caller's REPEATABLE READ snapshot.

Use a top-level XID generated by PostgreSQL for the admission-ticket transaction.

The ticket must be created by a separate READ COMMITTED operation and committed before the REPEATABLE READ capture transaction begins.

The acquisition helper must reject invocation from an existing REPEATABLE READ transaction.

Test:

- Ticket committed before snapshot: accepted.
- Ticket committed after snapshot: rejected.
- Ticket created inside the existing RR transaction: rejected.
- Ticket from another backend: rejected.
- Expired or released ticket: rejected.
- Unrelated old transaction holding global xmin: valid committed ticket remains acceptable.
- Ticket transaction aborted: rejected.

Do not use `xmin > ticket_xid` as the sole acceptance condition.

Document and test the exact PostgreSQL 16 behavior.

## B. Prevent advisory-lock substitution

The supplied empirical probe reports that session-level and transaction-level advisory locks have indistinguishable `pg_locks` rows in the tested environment.

Therefore, a current-lock-ownership check alone cannot establish that the session-level admission lock remains held.

Test the following attack:

1. Acquire a valid admission ticket and session-level locks.
2. Begin REPEATABLE READ.
3. Establish a snapshot.
4. Release the session-level lock.
5. Attempt to acquire a transaction-level lock with the same key.
6. Invoke privileged capture.

The attack must fail.

Investigate revoking direct `EXECUTE` privileges on the required PostgreSQL advisory-lock/unlock function overloads from `PUBLIC` and `app`, while granting narrowly required execution to the controlled SECURITY DEFINER wrappers.

All legitimate writer and snapshot lock operations must use the same canonical wrapper and lock-key derivation.

The app role must not be able to:

- Release an admission lock directly.
- Replace it with a transaction-level lock.
- Acquire an unapproved lock identity.
- Bypass the ticket lifecycle.

If the function privilege model cannot safely enforce this, request approval for the smallest snapshot-execution privilege change.

Do not silently add a third role.

## C. Population enumeration ordering

Use the following ordering:

1. Acquire employer population session barrier.
2. Execute a separate READ COMMITTED statement to enumerate the expected employee population.
3. Acquire employee session barriers in deterministic UUID order.
4. Acquire required rule-pack session barriers in deterministic identity order.
5. Create and commit the admission ticket.
6. Begin REPEATABLE READ on the same physical backend.
7. Revalidate population and required lock set.
8. Capture and persist the snapshot.

Do not enumerate population using a statement snapshot established before a population lock wait.

Any population mismatch must abort safely and follow bounded retry.

## D. Session-lock lifecycle

The shared Go implementation must use the same physical PostgreSQL connection for:

- Admission.
- Lock acquisition.
- Ticket verification.
- REPEATABLE READ capture.
- Commit or rollback.
- Lock release.

Test cleanup after:

- Success.
- Rollback.
- SQL error.
- Context cancellation.
- Deadline expiry.
- Panic.
- Unlock failure.
- Pool return.

Never return a connection with uncertain session-lock state to the healthy pool.

## E. Snapshot-authority architecture decision

Before implementing capture functions, compare the two alternatives:

1. Canonical PostgreSQL compensation resolver and snapshot derivation.
2. Canonical Go compensation resolver with a genuinely restricted snapshot-writing authority.

The ADR must compare:

- One authoritative compensation definition.
- Snapshot content completeness.
- Direct app SQL threat.
- Capture-function privileges.
- Canonical hash authority.
- Transaction consistency.
- Operational complexity.
- Compatibility with the frozen two-role model.

Do not duplicate the compensation resolver.

Do not maintain two canonical serializers or hashes.

If Option 2 is chosen, define precisely how the database prevents an ordinary app caller from submitting fabricated snapshot content.

Referential validation alone is not proof of completeness if the caller can omit effective inputs.

If a dedicated snapshot-execution identity or other privilege-boundary change is required, stop and request approval.

Do not use migration_admin as a broad runtime writer without explicitly analyzing the excessive privilege this creates.

## F. Final decision gate

Gate 1 must include live evidence for:

- The original stale-RR attack.
- Ticket visibility under the actual RR snapshot.
- False rejection under an unrelated long-running transaction.
- Lock substitution/release attack.
- Legitimate concurrent capture.
- Complete population and lock-set verification.
- Snapshot-authority direct-SQL fabrication attack.

If the candidate XID design or two-role privilege model cannot prevent the original failure class, stop and request approval.

Do not downgrade the acceptance guarantee.

## G. Report corrections

Record explicitly:

- Why `pg_locks` alone is insufficient.
- Why global xmin comparison may false-reject.
- The exact ticket visibility mechanism selected.
- The raw advisory-function privilege boundary.
- The selected canonical resolver and hash authority.
- Any required privilege-change request.

Replace `Open questions: None` with the explicit Gate 1 decision status until these choices have passed live verification.

## Execution authorization

The plan is authorized for:

- Baseline inspection.
- Red-first reproduction.
- Population-semantics inspection.
- Snapshot-authority ADR.
- D1 candidate experiments.

Full capture/validation implementation may proceed only after Gate 1 proves the selected mechanism and no unapproved privilege-boundary change remains.

All existing scope restrictions remain.

Keep migrations `00001–00012` unchanged, use one migration `00013`, preserve `OT-001 BLOCKED`, leave work uncommitted and stop at the acceptance gate.

# 5. Final verdict

Plan17.1 structure

Approved

Initial Gate 1 work

Authorized

Full capture implementation

Gate 1 dependent

D1 proof

Must demonstrate

Payroll calculations

Deferred

My recommendation is to proceed with the revised plan through Gate 1.

One final lesson is worth preserving: a synchronization token is only trustworthy when the system can prove both when it became valid and that the underlying synchronization resource has not subsequently been substituted or released.

Likewise, a snapshot writer is only trustworthy when it has one authoritative definition of the business inputs and a privilege boundary that prevents fabrication.

That is the standard our invariant-first principle requires. We should not lower it to preserve the schedule.