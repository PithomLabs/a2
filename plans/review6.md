I consolidated the additional review against `plan_prompt17.md`. I agree with its central conclusion: the plan is strong, but I would not authorize execution until the D1 enforcement boundary and the authoritative snapshot-data design are resolved.

The second review adds an important issue that was not sufficiently addressed in my previous amendments: we must avoid implementing two competing definitions of effective compensation and snapshot contents.

I would retain the original plan's scope and execution order, with the following consolidated amendments.

# 1. Consolidated verdict

APPROVE WITH REQUIRED AMENDMENTS

| Area                        | Disposition                                     |
| --------------------------- | ----------------------------------------------- |
| Invariant-first remediation | Keep unchanged                                  |
| D1 pre-BEGIN enforcement    | Critical unresolved design decision             |
| Snapshot capture authority  | Requires an explicit architectural decision     |
| Compensation resolver       | Avoid duplicating Go and SQL implementations    |
| Writer/lock sequence        | Reorder population semantics before enumeration |
| SECURITY DEFINER            | Add complete hardening requirements             |
| Rule-pack trust and pinning | Expand audit and provenance requirements        |
| Direct-SQL tests            | Enumerate explicitly                            |
| Migration safety            | Add disposable-database restrictions            |
| Historical migrations       | Preserve `00001–00012`                          |

## 2. The two substantive architectural problems

### A. D1: `pg_locks` proves lock ownership, not pre-BEGIN ordering

The proposed plan assumes a SECURITY DEFINER function can inspect `pg_locks` and establish that the necessary session-level barriers were acquired before the REPEATABLE READ snapshot.

That conclusion is not justified.

PostgreSQL 16 documents both session-level and transaction-level advisory locks as visible through `pg_locks`. Its documented columns expose the lock target, process, mode, grant status and transaction identifiers, but do not provide an explicit lock-acquisition timestamp or a field certifying that an advisory lock was acquired before the current transaction's MVCC snapshot.&#x20;

[image](https://www.google.com/s2/favicons?domain=https://www.postgresql.org\&sz=32)

postgresql.org+3



There is another subtlety: PostgreSQL permits the owning session to acquire an advisory lock it already holds without waiting, even when other sessions are waiting for that lock. Therefore, lock presence is not proof of a valid admission sequence.&#x20;

[image](https://www.google.com/s2/favicons?domain=https://www.postgresql.org\&sz=32)

postgresql.org+1



The attack must remain in scope:

```
BEGIN ISOLATION LEVEL REPEATABLE READ;

-- Establish a potentially stale MVCC view.
SELECT count(*) FROM public.employee;

-- Acquire barriers afterwards.
SELECT pg_advisory_lock(...);

-- Attempt privileged capture.
SELECT public.capture_payroll_snapshot(...);
```

A function that merely checks `pg_locks` could accept this operation.

Required decision: establish a database-enforceable pre-BEGIN admission boundary, or request approval for the smallest necessary privilege-boundary change.

A ticket or timestamp is only a proposed design, not a proven solution. Any candidate must withstand this exact direct-SQL attack.

### B. Snapshot capture must have one authoritative definition of business data

This is the second important issue.

The compensation resolver currently lives in Go:

`internal/compbasis/resolve.go`

The snapshot capture function is intended to operate inside PostgreSQL using SECURITY DEFINER.

There are three possible approaches, but we must select one deliberately.

Option 1 — PostgreSQL as canonical calculation authority.

Move the effective compensation resolution into one SQL function or view. Both the Go resolver and snapshot capture consume that same implementation.

Advantages: one authoritative definition and strong database derivation.

Cost: compensation resolution changes from a Go-owned implementation to a database-owned implementation. The compensation work must precede snapshot capture.

Option 2 — Go as canonical resolver, with a separate restricted snapshot-writer boundary.

Keep `internal/compbasis` authoritative. Introduce a narrowly privileged snapshot-execution capability that ordinary `app` credentials cannot use. The trusted Go service derives the snapshot using the canonical resolver, while the database controls persistence, lineage, audit, validation and immutability.

Advantages: avoids duplicating business logic.

Cost: introduces a privileged execution boundary and makes the trusted Go service part of the snapshot-authority boundary.

Option 3 — Duplicate and independently maintain both resolvers.

Reject this.

Two compensation resolvers will eventually disagree about supersession, effective dates or surviving spans. This would recreate the exact class of defect we are trying to eliminate.

The same question applies to canonical SHA-256 serialization. It needs one authoritative implementation, not independently maintained Go and PL/pgSQL serializers.

My preference: retain Go as the canonical business resolver unless there is a compelling, demonstrated reason to move it to SQL. But do not let ordinary `app` SQL fabricate snapshots. Require the agent to produce a short architecture decision record comparing the two viable approaches before implementing the capture path.

This decision affects the phase sequence. If SQL becomes authoritative, compensation resolution must move ahead of capture-function implementation.

# 3. Other valid amendments to consolidate

The attached review identifies several omissions that should be added directly to the plan.

| Subject               | Consolidated requirement                                                                                             |
| --------------------- | -------------------------------------------------------------------------------------------------------------------- |
| Population semantics  | Decide P1a effective-dated population rules before implementing enumeration                                          |
| Capture/validation    | Design the final authority boundary once; do not implement Go capture and later replace it with duplicated SQL logic |
| SECURITY DEFINER      | Restricted `search_path`, schema-qualified references, correct ownership, narrow EXECUTE and atomic audit            |
| Compensation resolver | One authoritative resolver and canonical hash implementation                                                         |
| C2-REPIN              | Full immutable pin history and concurrent single-winner behavior                                                     |
| Run lifecycle         | Only `draft → validated` reachable                                                                                   |
| P1g                   | Explicit employer, jurisdiction, currency, period start/end and other material header comparisons                    |
| C1-SOURCES            | Explicit source schema, byte verification and bundle digest                                                          |
| Direct SQL            | Enumerated attack matrix, not merely a general reference to §12B                                                     |
| Tests                 | Positive controls for every authority revoked                                                                        |
| Migration tests       | Guard destructive operations to verified disposable databases                                                        |
| Architecture          | Preserve `bchat/`, upstream repositories and user databases                                                          |
| Payroll lines         | Explicit non-goal                                                                                                    |
| Stop condition        | Stop when the acceptance gate passes                                                                                 |

There are also several specific requirements to restore:

Rule-pack pin history must include previous and new versions, digests, jurisdiction, scope, effective period, actor, reason, timestamp, bundle digest, specification version, independent review reference, gate-artifact digest and evidence references.

SHG verification must include verifier identity, verification date and verification basis. Full-period coverage must be tested.

The direct SQL matrix must include stale-RR invocation, fabricated header, employee and pack rows, invalidation, current-tip manipulation, post-validation pin tampering and empty-population validation.

Baseline verification must be precise: the recorded checksum file covers `00001–00011`; `00012` must be independently hashed at Phase 0 and rechecked at the gate. Do not falsely claim an existing recorded checksum for it.

The new golden fixture schema should be kept within the permitted implementation/test artifacts unless additional documentation is explicitly authorized.

# 4. Consolidated amendment prompt for the implementation agent

# Final Consolidated Plan Amendments — Prompt17 / Milestone 2.1

The existing `plan_prompt17.md` and its execution plan are approved in principle, but execution is NOT yet authorized.

Preserve:

- The mandatory Invariant-First Root-Cause Remediation principle.
- The original scope and 20 acceptance conditions.
- The single forward migration `00013`.
- Historical migrations `00001–00012` byte-identical.
- The existing Go modular monolith and PostgreSQL architecture.
- The two-role model unless an explicit privilege-boundary change is approved.
- OT-001 BLOCKED.
- No statutory calculations.

Incorporate the following consolidated amendments.

## A. Resolve D1 pre-BEGIN enforcement before implementation

The proposed `pg_locks` check establishes current lock ownership, not historical acquisition order or proof that the session lock preceded the REPEATABLE READ MVCC snapshot.

Do not claim otherwise.

Verify whether PostgreSQL 16 exposes sufficient information to distinguish:

- Session-level advisory-lock ownership.
- Transaction-level advisory-lock ownership.
- Acquisition order relative to the current transaction's snapshot.

Do not assume that `pg_locks` provides these guarantees.

Reproduce the direct-SQL attack:

1. BEGIN REPEATABLE READ.
2. Execute a query establishing the transaction's MVCC view.
3. Acquire the required session advisory locks.
4. Invoke the privileged capture or validation function.
5. Determine whether the stale view is accepted.

The attack must fail.

Evaluate:

1. A database-verifiable pre-BEGIN admission protocol.
2. A narrowly restricted snapshot-execution privilege boundary.
3. Any other mechanism that genuinely prevents the stale-RR invocation.

A ticket, timestamp, session variable or `pg_locks` ownership check is not accepted without proving that a caller cannot create it after an already-established stale snapshot.

If the existing two-role model cannot enforce this invariant, prepare a short privilege-boundary decision request explaining the smallest necessary change.

Do not silently add a third role or claim D1 fixed.

## B. Establish one authoritative snapshot-data implementation

The Go compensation resolver lives in `internal/compbasis/resolve.go`.

The proposed SECURITY DEFINER capture function must not duplicate its effective-compensation logic or accept arbitrary caller-supplied JSON as authoritative.

Before implementing the capture path, produce a short architecture decision record comparing:

**Option 1: Canonical PostgreSQL resolver**

- Effective compensation resolution moves into a shared SQL function or view.
- Go and database snapshot capture consume the same authoritative implementation.
- Canonical hashing has one authoritative implementation.
- Compensation work is moved ahead of capture-function implementation.

**Option 2: Canonical Go resolver with a restricted snapshot-execution boundary**

- `internal/compbasis` remains the authoritative resolver.
- Ordinary `app` credentials cannot INSERT or mutate snapshot contents.
- A separate narrowly scoped trusted execution capability performs snapshot persistence.
- Database constraints, source identity checks, immutable history, audit and validation remain enforced at the database boundary.
- Arbitrary `app` SQL cannot fabricate snapshot content.

Compare both against:

- Database integrity.
- Direct-`app` threat model.
- Code duplication.
- Transaction consistency.
- Testability.
- Operational complexity.
- Existing two-role architecture.

Recommend one approach and document the precise trust boundary.

Do not implement two independent compensation resolvers.

Do not independently maintain multiple canonical hash algorithms.

## C. Correct sequencing

Reorder the plan as follows:

1. Phase 0: baseline and inventory.
2. Reproduce D1 and all missing writer races.
3. Inspect and decide P1a expected-population semantics.
4. Decide the canonical compensation/snapshot authority.
5. Design and prove the pre-BEGIN admission and lock protocol.
6. Implement the complete writer-to-barrier matrix.
7. Implement capture and validation authority once, according to the approved design decision.
8. Implement compensation resolver corrections.
9. Implement SHG semantics and full-period coverage.
10. Implement rule-pack source, golden and pin verification.
11. Execute the complete adversarial battery.
12. Produce the 20-section report.

Do not implement a temporary Go capture/validation path and subsequently replace it with duplicated SECURITY DEFINER logic.

## D. Complete SECURITY DEFINER hardening

Every SECURITY DEFINER function introduced or changed must have:

- Correct owner.
- Restricted `search_path`.
- Schema-qualified protected table references.
- Narrow EXECUTE grants.
- No mutable caller-controlled authorization settings.
- Correct application/admin privilege separation.
- Atomic business and audit writes.
- Positive and negative direct-SQL tests.

For P1d, prefer the existing controlled employee-status transition function rather than a broad trigger that permits arbitrary status-history insertion.

## E. Restore complete rule-pack trust requirements

C1-SOURCES must define:

- Stable source ID.
- Relative file path.
- SHA-256 digest.
- Authority/source identity.
- Retrieval metadata.
- Regular-file validation.
- Path-traversal rejection.
- Symlink-escape rejection.
- Actual byte rehashing.
- Source inclusion in the bundle digest.
- Explicit prohibition on treating `SOURCES.md` presence as source verification.

C1-GOLDENS must genuinely execute or independently verify golden, anti-golden and boundary fixtures.

C2-REPIN must preserve:

- Previous and new versions.
- Previous and new digests.
- Rule-pack identity.
- Jurisdiction and scope.
- Effective period.
- Actor and reason.
- Timestamp.
- Bundle digest.
- Specification version.
- Independent review reference.
- Gate-artifact digest.
- Evidence references.

The pin operation must reject invented versions, serialize concurrent changes to one winner and persist an atomic audit event.

Ensure the registry's supported repinning path cannot bypass the history. Document the privileged database-owner trust boundary honestly.

SHG records must preserve verifier identity, verification date and verification basis.

Rule-pack effective coverage must be checked against the entire payroll period.

OT-001 remains BLOCKED.

## F. Make the run lifecycle restriction explicit

Only `draft → validated` may be reachable in this milestone.

`calculated`, `approved` and `finalized` must remain structurally unavailable without genuine later-stage artifacts.

Add direct-SQL regression tests for all three transitions.

A snapshot is not a calculation artifact.

## G. Enumerate P1g drift fields

Drift comparison must explicitly include:

- Employer identity.
- Jurisdiction.
- Currency.
- Payroll period start.
- Payroll period end.
- Run kind.
- Other fields that affect snapshot interpretation.

Add one mutation test per material field.

## H. Correct the snapshot-value inventory

Replace the unexplained claim of 13 value classes with an explicit inventory matching Prompt17.

Cover:

- Run identity.
- Employer identity and jurisdiction.
- Expected employee population.
- Employment terms and status.
- Effective compensation.
- Approved timesheet version.
- Attendance and correction lineage.
- SHG instruction/evidence.
- Rule-pack pins and activation evidence.
- Inclusion/exclusion decisions.
- Snapshot schema/version.
- Canonical content hash.

The final schema must identify where each category is represented.

## I. Complete direct-SQL and positive-control tests

Enumerate, rather than merely reference, the attack cases:

- Stale-RR privileged invocation.
- Fabricated snapshot header.
- Fabricated employee child.
- Fabricated pack child.
- Direct invalidation.
- Current-tip manipulation.
- Post-validation pin tampering.
- Placeholder hash.
- Empty employee population.
- Missing inclusion/exclusion decision.
- Unfresh validation offer.
- Unauthorized rule-pack repin.
- Unverified source digest.
- Failing golden assertion.

For every revoked privilege, provide a positive control demonstrating that the legitimate controlled operation still works.

Every rejected operation must be followed by persisted-state inspection and a no-partial-write assertion.

## J. Preserve the original project boundaries

Explicitly retain:

- `bchat/` untouched.
- Upstream Phys-Research and Phys-Pseudo untouched.
- Existing user databases untouched.
- No Neon/Fly deployment.
- No historical migration modifications.
- No payroll-line behavior.
- No statutory calculations.
- No architecture redesign.
- Stop immediately once the acceptance criteria pass.

## K. Baseline and clean database safeguards

Recorded checksums cover `00001–00011`.

Hash `00012` independently during Phase 0 and verify it again at acceptance.

Do not claim an existing baseline checksum for `00012`.

Every automated database drop/recreate must require a verified disposable database identity or allowlist. A naming convention alone is insufficient unless the test harness proves the target is disposable and isolated.

Use the default environment, `TZ=UTC` and `TZ=Asia/Singapore` for the required three-run suite.

Do not create extra documentation files outside the allowed deliverables without approval.

## Execution authorization

After documenting the D1 and snapshot-authority decisions, proceed only where the design is demonstrably enforceable.

If D1 requires a privilege-boundary change, stop and request approval.

Do not proceed under a weaker guarantee merely to keep the implementation moving.

Preserve all valid work and leave changes uncommitted.

Stop at the original Milestone 2.1 acceptance gate.

# 5. My final recommendation

There is one principle I want to emphasize beyond the amendments:

Do not let the snapshot capture design create a second implementation of business truth.

If Go determines the effective compensation, then either Go remains the canonical resolver behind a genuinely restricted snapshot-writing boundary, or we intentionally make PostgreSQL the canonical resolver and have Go call it. Maintaining both independently is unacceptable.

And on D1, do not settle for a function that can see a lock. We need proof that the snapshot's visibility boundary was established after the required synchronization.

The attached review and your written critique independently converge on these same problems.&#x20;

Pasted text(20261109-073640).txt

Pasted text(20261109-073640).txt



My recommendation: amend the execution plan with this consolidated instruction, obtain an explicit decision on the D1 privilege boundary and snapshot authority, then authorize implementation. This is exactly where our locked invariant-first principle matters most: a plausible mechanism is not a proven guarantee.