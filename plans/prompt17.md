Agreed. We need to change how we approach remediation from this point forward.

A remediation must correct the underlying invariant, not merely suppress the observed failure. This is now a mandatory principle I will incorporate into every future implementation prompt.

For example, the snapshot defect is not simply that we need another lock. The underlying problem is that PostgreSQL's MVCC snapshot can be established before a transaction finishes waiting for the synchronization barrier. The correct fix must establish the required synchronization before the snapshot exists, and enforce the same protocol across every operation that can change the captured data.

Likewise, the snapshot hash problem is not simply missing permissions. The underlying problem is that the component responsible for creating the snapshot is not the exclusive authority over its contents.

## 1. The remediation philosophy we are adopting

I would make the following a permanent engineering rule for WorkforceOps:

Invariant-first, root-cause remediation.

Every defect must be approached through six questions:

1. What business invariant should always hold? Describe the truth the system must preserve, independently of implementation.
2. What mechanism allowed it to fail? Identify the actual transaction, query, relationship, privilege, state transition or interpretation responsible.
3. Where are all the write paths? Examine CLI, REST, application services, triggers, database functions, imports, administrative operations and direct `app` SQL.
4. What is the authoritative enforcement boundary? Put the guarantee where it cannot be bypassed, not only in the most convenient Go function.
5. How do we prove the correction? Reproduce the original failure, test both operation orderings, inspect persisted state and verify that failure leaves no partial changes.
6. What new failure could the correction introduce? Check deadlocks, stale views, historical corruption, privilege escalation, lost evidence and unintended business restrictions.

A green test is evidence only when its test actually exercises the failure mechanism.

We should also avoid indiscriminate locking. Every lock must protect a named invariant, use a documented identity and participate in a consistent acquisition order.

# 2. My recommended next milestone

Milestone 2 not yet accepted

Milestone 2.1 — Payroll Snapshot Integrity and Trust Closure

The implementation should address the four principal code-remediation groups, the three SHG/rule-pack decisions, and the missing live verification.

The independent review explicitly identifies these as the minimum prerequisites before the first payroll calculation.&#x20;

milestone2_adversarial_review\.md



# Prompt 17 — WorkforceOps Milestone 2.1

# Payroll Snapshot Integrity and Trust Closure

## Root-Cause Remediation, Not Symptom Patching

## 1. Objective

Implement the minimum corrections required by:

`plans/milestone2_adversarial_review.md`

The independent Prompt 16 review returned:

**FAIL — NOT READY FOR THE FIRST PAYROLL CALCULATION.**

The decisive findings are:

- D1-P0: Snapshot capture and validation can use a stale but internally consistent PostgreSQL REPEATABLE READ view.
- EF-E3-P0: The ordinary PostgreSQL `app` role can fabricate or extend snapshot records without changing the content hash.
- P1a–g: Seven missing synchronization or privilege protections affect snapshot completeness.
- EF-F1A/B: Direct SQL can validate a run using fabricated, placeholder or empty snapshot content without the required audit and offer checks.
- F-A1-WIDEN: Compensation chain resolution fails after certain widening corrections.
- F-A1-FREETAIL: Certain valid compensation writes create permanently ambiguous effective compensation.
- F-A1-MIDPERIOD: Mid-period compensation changes can be included without verified proration semantics.
- F-B1-ACTIVE: Active SHG instructions can count as complete without an established verification standard.
- F-B2-PARTIAL: Partial-period SHG instructions can be treated as full-period coverage.
- X-FUND: Ambiguous multi-fund SHG instructions can silently pass.
- C1-SOURCES: Manifest source digests are not verified against actual source-file bytes.
- C1-GOLDENS: Golden verification is asserted without executing or verifying a genuine verification artifact.
- C2-REPIN: The rule-pack registry trust root can be changed without preserving an auditable pin history.
- Additional P1 prerequisites concerning rule-pack labels, effective-period coverage and independent verification.

The objective is to close these defects, establish a complete synchronization contract and independently verify the results using real PostgreSQL.

**Do not implement CPF, SDL, SHG contribution calculations, overtime, gross pay, net pay or payslips.**

This milestone must establish that a payroll snapshot is complete, authentic, reproducible and governed.

---

# 2. Mandatory engineering principle: invariant-first root-cause remediation

This principle applies to every task in this milestone and every future WorkforceOps remediation milestone.

Do not begin by changing the line of code identified in a finding.

First identify the underlying invariant and the mechanism that violated it.

For each defect, document:

1. The business invariant.
2. The existing failure mechanism.
3. All relevant write and read paths.
4. The actual database and transaction semantics.
5. The authoritative enforcement point.
6. The proposed correction.
7. How the correction prevents the entire class of failures.
8. How the regression test reproduces the original failure.
9. What new failure modes the correction might introduce.
10. How persisted-state verification proves correctness.

### Examples of the required reasoning

**Stale snapshot**

Bad remediation: add another advisory lock somewhere in `CaptureSnapshot`.

Correct remediation: establish the synchronization barriers before PostgreSQL establishes the REPEATABLE READ MVCC view, and ensure every relevant writer participates in the same barrier contract.

**Fabricated snapshot**

Bad remediation: validate JSON more carefully in Go.

Correct remediation: remove direct application-role authority to write snapshot contents. Only a controlled database capture operation may derive and persist snapshot records from authoritative business records.

**Compensation resolver**

Bad remediation: special-case widening corrections.

Correct remediation: separate complete supersession-ancestry loading from period-span carving, and make write-time liveness validation agree with the resolver's effective-value semantics.

**SHG coverage**

Bad remediation: add another `EXISTS` condition.

Correct remediation: define the exact evidence and effective-period coverage model, then ensure readiness evaluates complete coverage rather than the existence of one qualifying row.

**Rule-pack verification**

Bad remediation: reject malformed hash strings.

Correct remediation: bind source identities to actual source files, recompute the file digests, execute or verify the independent golden gate and preserve an auditable reviewed trust root.

Do not accept a patch that makes the original example pass while leaving the underlying invariant unenforced.

## 2.1 Permanent implementation requirements

For every critical invariant:

- Identify every supported mutation path.
- Enforce it at the appropriate database boundary.
- Keep Go validation for usable error reporting.
- Test direct SQL under the real `app` role.
- Test concurrency using separate real PostgreSQL connections.
- Test both possible lock-winner orders.
- Inspect persisted state after success and failure.
- Verify transaction atomicity.
- Verify historical evidence remains unchanged.
- Verify failed operations leave no partial state.
- Add a regression test reproducing the original failure.
- Add a positive control proving legitimate operations still work.

Do not solve defects by:

- Adding timestamps to imitate commit ordering.
- Adding locks without defining their scope and order.
- Increasing timeout values.
- Catching and suppressing errors.
- Adding generic retries that hide consistency failures.
- Excluding difficult records from readiness.
- Weakening database constraints.
- Treating missing evidence as verified non-applicability.
- Using application-only checks to protect database invariants.

A root-cause correction must explain why the entire identified class of failure is no longer possible within the supported threat model.

---

# 3. Repository preservation and baseline

Before modifying anything:

1. Record `git status --short`.
2. Record `git rev-parse HEAD`.
3. Confirm the M2 baseline is `304fec88daaadae493309bfbad65f7f472d89c62`.
4. Inventory tracked modifications and untracked files.
5. Confirm `docs/writeup3.md` is a pre-existing user file and preserve it.
6. Verify migrations `00001`–`00012` against their recorded baseline.
7. Read `plans/milestone2_adversarial_review.md` in full.
8. Inspect all referenced source files and database objects.
9. Record PostgreSQL, Go, Goose and relevant tooling versions.

Preserve all existing uncommitted work.

Do not:

- Reset or clean the repository.
- Stage or commit.
- Discard previous work.
- Modify `bchat/`.
- Modify upstream Phys-Research or Phys-Pseudo.
- Modify historical migrations `00001`–`00012`.
- Drop or recycle any existing user database.
- Deploy to Neon or Fly.io.
- Implement statutory calculations.

Use explicitly disposable PostgreSQL 16 databases.

The only permitted new repository files are the required forward migration(s), implementation changes, regression tests and:

`plans/milestone2_1_remediation_report.md`

All work remains uncommitted.

Use one forward migration beginning at `00013`, unless a demonstrated dependency requires another isolated migration. Explain the reason before adding additional migrations.

---

# 4. Frozen architecture

Preserve:

- One Go executable.
- One PostgreSQL database.
- One REST API.
- One CLI.
- One shared application-service layer.
- pgx and explicit SQL.
- Employer-scoped database relationships.
- One employer operationally in v0.1.
- Multi-employer-ready schema.
- `migration_admin` for migrations and controlled administrative operations.
- `app` for ordinary application operations.
- Loopback-only API.
- No UI.
- No microservices.
- No event bus.
- No distributed workflow engine.
- No RLS in v0.1.
- No generic payroll rules engine.
- No additional country packages.

Retain Singapore as the initial jurisdiction.

Retain Phys-Research and Phys-Pseudo as the methodology for statutory rules.

Do not reopen the architecture unless implementation discovers a genuine structural impossibility. Explain the impossibility and request approval rather than changing a frozen decision silently.

---

# 5. Priority 1 — Fix D1: correct snapshot isolation at its root

This is the central P0.

## 5.1 Reproduce the failure first

Reproduce the exact writer-first interleaving documented in Prompt 16.

Use separate PostgreSQL connections authenticated as `app`.

Writer A:

1. Begin a transaction.
2. Acquire the applicable population, employee or pack barrier.
3. Insert or modify an input.
4. Pause before commit.

Snapshot B:

1. Begin the current snapshot path.
2. Attempt to acquire the same barrier.
3. Wait for Writer A.
4. Allow Writer A to commit.
5. Continue snapshot capture.
6. Inspect whether the writer's committed change is included.

Reproduce at least:

- Employee creation.
- Compensation insertion.
- Rule-pack pin or activation change.
- Timesheet approval.

Also reproduce the validation-level variant.

Capture `pg_locks` to prove the reader is actually waiting rather than merely executing later by chance.

Preserve the red reproduction as a regression test design.

## 5.2 Correct the MVCC/lock ordering

The current defect exists because the REPEATABLE READ view may be established before a transaction finishes waiting for an advisory lock.

The required consistency protocol is:

**Acquire all synchronization barriers before BEGIN REPEATABLE READ on the same physical PostgreSQL connection.**

Conceptually:

```
Acquire dedicated physical PostgreSQL connection
                 |
                 v
Acquire employer population session lock
                 |
                 v
Read stable expected employee IDs
                 |
                 v
Acquire employee session locks in sorted UUID order
                 |
                 v
Acquire required rule-pack session locks in sorted order
                 |
                 v
Acquire required run-operation lock
                 |
                 v
BEGIN ISOLATION LEVEL REPEATABLE READ
                 |
                 v
Read authoritative inputs
                 |
                 v
Build and persist snapshot
                 |
                 v
Calculate canonical hash
                 |
                 v
COMMIT / ROLLBACK
                 |
                 v
Release session locks in reverse order
                 |
                 v
Return connection safely to pool
```

This is a logical protocol, not permission to introduce redundant locks.

Determine the smallest set of barriers necessary to protect the exact inputs being captured.

### Important PostgreSQL implementation details

1. Acquire session-level advisory locks while the connection is outside the snapshot transaction.
2. Use the same physical connection for lock acquisition, the REPEATABLE READ transaction and lock release.
3. Do not acquire a lock through `pgxpool` and then begin the transaction on another connection.
4. Do not release the physical connection to the pool while session-level locks remain held.
5. Release locks in reverse acquisition order.
6. On unlock failure, do not return the connection to the pool as healthy. Discard or safely reset it.
7. Handle cancellation, error, panic and rollback paths.
8. Verify every required lock was released.
9. Ensure retries start with a clean connection and a newly acquired barrier set.

Use one shared Go implementation for this protocol.

Do not independently reimplement lock-key derivation in individual services.

### Prevent the original failure, not just the test example

The critical ordering is:

`Synchronization first → MVCC snapshot second`

It is not:

`Begin REPEATABLE READ → acquire locks → assume snapshot is fresh`

A transaction-level advisory lock acquired after the transaction snapshot has been established does not repair the stale view.

## 5.3 Stable population enumeration

After acquiring the employer population session lock, establish the expected employee set through a committed READ COMMITTED statement on the same physical connection before beginning REPEATABLE READ.

The population lock must prevent population-changing writers from committing while the set is being enumerated and employee barriers are being acquired.

Then:

1. Acquire employee locks in deterministic UUID order.
2. Acquire the required rule-pack locks.
3. Begin REPEATABLE READ.
4. Recompute/revalidate the expected population.
5. Confirm it matches the pre-BEGIN lock set.
6. Only then capture source records.

If the population differs, abort and retry within a bounded policy.

Never omit an employee merely because an employee lock was acquired too late.

## 5.4 CaptureSnapshot and ValidateRun

Correct both operations.

`CaptureSnapshot` must acquire the appropriate barriers before beginning the snapshot transaction.

`ValidateRun` must do the same for:

- Snapshot tip resolution.
- Drift detection.
- Validation-offer verification.
- Validation transition.
- Rule-version pinning.

The drift-check path must not use a stale REPEATABLE READ view.

Validation must never approve a snapshot that predates a source transaction which committed while validation was waiting for its barrier.

## 5.5 Handle serialization errors honestly

Map SQLSTATE `40001` from relevant run/snapshot locking paths to a typed conflict result.

Do not misclassify a serialization failure as an internal server error.

Retry only when the full operation is safely repeatable and starts with a fresh transaction and correct lock ordering.

Do not hide repeated serialization failures with unlimited retries.

---

# 6. Priority 2 — Close all seven synchronization gaps

The D1 fix is incomplete until every relevant writer participates in the synchronization contract.

Build a writer-to-barrier matrix.

## P1a — Timesheet approval and employee population

Inspect why timesheet approval can add an employee to the expected population.

Correct the population semantics first.

Prefer deriving expected employee membership from effective-dated employment applicable to the pay period, not merely the existence of an approved timesheet.

If historical approved timesheets legitimately extend the population for a documented legacy reason, preserve that behavior explicitly.

Where approval can affect population membership, it must acquire:

- Employer population barrier.
- Employee input barrier.

The order must be consistent with snapshot capture.

Test direct SQL and service approval.

## P1b — Employment-term closure

`close_employment_terms` currently takes the population lock without consistently taking the employee payroll-input lock.

Correct the database function and any service paths.

Require:

`Employer population lock → Employee input lock → mutation`

Use the same employee lock helper and key as compensation and snapshot capture.

Test:

- Terms closure before capture.
- Terms closure waiting behind capture.
- Direct function execution.
- Concurrent successor creation.
- Historical-period conflict.

## P1c — Direct-SQL event disposition

The disposition trigger currently uses the pair-lock but does not fully participate in the employee payroll-input barrier.

Determine how disposition changes affect snapshotted clock-event evidence.

If the disposition can change the evidence interpretation included in the snapshot, make it acquire the employee input barrier.

Use one lock ordering for all paths:

`Employee input lock → Disposition pair-lock`

or another rigorously documented order that is used consistently everywhere.

Do not acquire these in opposite orders across service and trigger paths.

Verify the employee identity from the canonical source-event records, not caller-supplied identity.

Test direct SQL under `app`.

## P1d — Employee status-history INSERT

`app` can currently insert directly into `employee_status_history`.

Revoke direct INSERT and route it through the protected status-transition function, or implement a database trigger that participates in the required population and employee barriers.

The preferred design is the existing controlled status-transition service/function.

A direct status-history INSERT must not create a status transition that bypasses employee population consistency or audit.

## P1e — `ensure_pack_blocked`

The current `ensure_pack_blocked` function must participate in the same rule-pack synchronization contract as snapshot capture and activation.

Use the canonical pack lock helper.

A snapshot must not miss a concurrent change from `BLOCKED` to another state or vice versa.

Test both writer-first and snapshot-first interleavings.

## P1f — Draft timesheet lines

Direct draft-line INSERT/UPDATE/DELETE can currently change information consumed by snapshotting without acquiring the employee input barrier.

Choose one coherent policy:

- Revoke unrestricted line mutation and route it through locked application operations; or
- Add narrowly scoped database triggers that acquire the owning employee's payroll-input barrier.

The employee must be derived through the parent timesheet.

Do not trust caller-supplied employee IDs.

Preserve legitimate draft editing.

Ensure submitted and approved timesheets remain protected by the existing immutability rules.

## P1g — Employer and pay-period identity

`app` can update employer and pay-period fields that affect snapshot interpretation.

Correct the authority boundary:

- Revoke unrestricted UPDATE where mutation is not required; or
- Provide controlled mutations acquiring the required population and other relevant barriers.

Snapshot drift detection must compare:

- Period start and end.
- Employer jurisdiction.
- Currency.
- Other header fields that materially affect interpretation.

A payroll snapshot must not silently change meaning because a parent row was edited in place.

Prefer immutable versioning or controlled parent succession where history matters.

## Lock-order acceptance

Document one canonical lock hierarchy.

For example:

```
Employer population
        |
        v
Employee locks (sorted UUID)
        |
        v
Rule-pack locks (sorted identity)
        |
        v
Operation-specific locks
```

Adapt this to the actual operations, but ensure there is no reverse acquisition path.

Every new or existing helper must use the same canonical lock-key derivation.

---

# 7. Priority 3 — Close fabricated snapshot and validation paths

This addresses EF-E3-P0 and EF-F1A/B.

## 7.1 Remove direct snapshot INSERT authority

The `app` role must not directly INSERT snapshot headers or child rows.

Revoke INSERT on all protected snapshot tables from `app`.

The existing snapshot service must use one controlled capture operation.

Use a narrowly scoped `SECURITY DEFINER` function or equivalent database-owned mechanism.

### Critical requirement

Do not replace direct INSERT with a SECURITY DEFINER function that simply accepts arbitrary caller-supplied JSON and writes it.

That would merely move the fabrication capability behind a function call.

The controlled capture operation must derive or independently verify every snapshot value against authoritative source records.

It must validate:

- Run identity.
- Employer identity.
- Expected employee population.
- Employment terms.
- Effective compensation.
- Approved timesheet version.
- Attendance lineage.
- SHG instruction.
- Rule-pack activation.
- Inclusion/exclusion decisions.
- Snapshot version.
- Canonical content.

The caller may supply the run identity and authorized operation context.

It must not be authoritative for arbitrary child records, employee identities, rule pins or content hashes.

### Pre-BEGIN lock proof

Do not assume a SECURITY DEFINER function automatically solves D1.

The capture path must have a verifiable synchronization protocol proving that the required barriers were acquired before the RR snapshot was established.

A caller must not be able to begin a stale RR transaction, acquire the session lock afterwards, invoke the privileged function and create a stale snapshot.

Design the privilege/function boundary accordingly.

If this cannot be enforced under the frozen two-role model, explain the exact PostgreSQL limitation and propose the smallest necessary privilege-boundary change for approval. Do not silently claim the problem is fixed.

All SECURITY DEFINER functions must have:

- Restricted `search_path`.
- Schema-qualified protected references.
- Correct ownership.
- Narrow EXECUTE privileges.
- No mutable caller-controlled authorization setting.
- Atomic audit.

## 7.2 Snapshot child immutability

Verify that:

- App cannot INSERT snapshot children directly.
- App cannot append employee or rule-pack rows after capture.
- App cannot fabricate invalidation events.
- App cannot change the current snapshot pointer.
- App cannot alter the content hash.
- App cannot re-parent child records.

Snapshot content remains immutable.

Revalidation creates a new version and append-only lifecycle records.

## 7.3 Controlled validation

Revoke direct app UPDATE permissions on:

- `payroll_run.status`.
- `payroll_run.inputs_hash`.
- `payroll_run.rule_versions`.
- Any other fields representing validation authority.

Introduce a controlled validation function that:

1. Acquires/verifies the required pre-BEGIN barrier context.
2. Resolves the actual current snapshot tip.
3. Verifies the tip is valid and not invalidated.
4. Recomputes or verifies the content hash.
5. Confirms the run references that exact tip.
6. Confirms rule pins equal the snapshot pins.
7. Confirms at least one employee row exists.
8. Confirms every expected employee has a terminal inclusion or exclusion decision.
9. Refuses unresolved blockers.
10. Verifies the validation offer identity and freshness.
11. Checks source drift using the correctly synchronized view.
12. Stamps validation metadata.
13. Writes exactly one atomic audit record.
14. Performs the permitted `draft → validated` transition.

A snapshot with zero employees must never validate.

An empty rule-version object must never pass merely because an empty blocker query returns no rows.

The audit event is part of the business decision, not optional decoration.

## 7.4 Direct-SQL tests

As `app`, test:

- Fabricated snapshot header.
- Fabricated employee row.
- Fabricated rule-pack row.
- Extra child inserted after validation.
- False content hash.
- Snapshotless validation.
- Placeholder pins.
- Empty population.
- Missing inclusion/exclusion decisions.
- Stale offer.
- Post-validation pin tampering.
- Direct invalidation.
- Direct current-tip manipulation.

Inspect persisted state after every operation.

---

# 8. Priority 4 — Correct the compensation resolver at its root

Inspect:

- `internal/compbasis/resolve.go`
- Compensation write/correction services and functions.
- Compensation triggers.
- Readiness.
- Snapshot resolver integration.
- Existing compensation tests.

## 8.1 F-A1-WIDEN — Separate ancestry loading from span carving

Current defect:

The resolver loads predecessors using a period-overlap query. A widening successor can replace the predecessor over a wider interval, leaving the ancestor outside the requested payroll period.

The resolver then sees a child whose parent is missing and reports a dangling supersession chain.

This is a query-design problem.

### Required correction

Separate two distinct operations:

**Lineage loading:** load the entire relevant supersession ancestry needed to prove chain integrity, independently of whether each ancestor overlaps the requested period.

**Effective-span carving:** after the lineage is established, calculate which date segments are effective within the requested period.

A historical ancestor may be outside the requested period and still be required to establish that the current successor is legitimate.

Do not discard ancestry merely because its original effective range does not overlap the current query.

Add widening and disjoint-successor tests.

## 8.2 F-A1-FREETAIL — Make write-time liveness agree with resolver semantics

Current defect:

A predecessor is narrowed by a successor, but an old effective tail remains outside the successor range.

A fresh compensation row can then be inserted for that tail in a way that creates an ambiguous effective basis.

This means the writer and resolver enforce different interpretations of effective compensation.

### Required correction

Define effective ownership after supersession.

Write-time validation must evaluate the predecessor's surviving effective spans, not simply ignore the predecessor because it has a successor.

Reject fresh rows that overlap surviving predecessor spans.

Alternatively, require the operator to correct the chain tip through the supported correction workflow.

A successful write must be guaranteed resolvable by the authoritative resolver.

Do not allow records to be stored successfully if every subsequent period query involving them will remain permanently blocked.

Tests must demonstrate:

- Original open-ended row.
- Narrowing successor.
- Attempted new tail row.
- Correct resolution of all permitted periods.
- A later correction of the effective tip.
- No permanent ambiguity.

Use real PostgreSQL and service execution as `app`.

## 8.3 F-A1-MIDPERIOD — Explicitly block unsupported monthly proration

For the first restricted monthly payroll profile, a compensation classification that changes during the pay period must not silently be treated as one ordinary full-month salary input or as two automatically payable segments.

Until proration is independently researched and implemented:

- Detect more than one effective compensation segment for the relevant monthly classification.
- Return `COMP_MID_PERIOD_CHANGE`.
- Include the affected classification and dates.
- Block readiness.
- Block snapshot validation.
- Preserve both historical compensation records.
- Do not calculate or infer prorated amounts.

The same blocker must be produced consistently by readiness, snapshot capture and validation.

Do not add a special-case only to one Go endpoint.

## 8.4 F-A2 provenance observation

The report identifies a non-blocking omission of `terms_id` and explicit record identity in snapshot compensation segments.

Evaluate whether the existing lineage reference is sufficient to reconstruct the parent employment-term relationship.

If not, add the exact parent and source record identifiers to the snapshot segment.

Do not defer necessary provenance merely because the finding was initially classified P2.

Do not build a generic provenance framework.

---

# 9. Priority 5 — Resolve SHG instruction semantics and coverage

The SHG implementation has good underlying lifecycle controls, but its readiness interpretation remains incomplete.

Inspect:

- `internal/hr/shg.go`
- `internal/payroll/readiness.go`
- Snapshot SHG capture.
- SHG CHECK constraints and controlled functions.
- SHG tests.

## 9.1 B1 — Define the meaning of active

The current model permits an active instruction without a clear verification path and readiness treats it as complete.

This is a specification problem, not just a missing condition.

Adopt the conservative default:

**An active SHG instruction may count as complete only when its evidential standard has been explicitly established and satisfied.**

Unless authoritative, reviewed evidence supports declaration-only treatment:

- A declaration without verification remains pending or unverified.
- Verification must identify the supporting evidence.
- The verifier, verification date and basis must be preserved.
- Readiness blocks until the instruction satisfies its verification standard.

Do not silently redefine `active` to mean verified without recording the semantics.

If the data model needs a separate verified-active status, make that distinction explicit.

The chosen model must be documented in the SHG domain specification and tested.

## 9.2 B2 — Require complete effective-period coverage

A qualifying SHG instruction that begins halfway through a pay period does not automatically cover the entire period.

Define the effective-coverage algorithm.

For the first profile:

- Resolve the effective instruction history.
- Determine the valid live applicable or verified-none instructions.
- Calculate their date coverage.
- Require the union of valid coverage intervals to cover the entire payroll period without gaps.
- Identify uncovered intervals explicitly.
- Reject conflicting overlapping instructions.

A simpler one-record-must-span-the-whole-period rule is acceptable for the first restricted profile if the implementation and rationale are explicit.

Do not infer coverage from one existential `active` row.

Add the nine missing coverage scenarios:

- Starts mid-period.
- Ends mid-period.
- Complete-period coverage.
- Gap between instructions.
- Adjacent successor instructions.
- Pending predecessor superseded mid-period.
- Verified-none successor.
- Conflicting overlap.
- Historical instruction outside the period.

## 9.3 X-FUND — Conservatively block unresolved cross-fund ambiguity

The system cannot currently establish all necessary eligibility and interaction semantics from the stored information.

Until authoritative research establishes the permitted combination:

- Multiple relevant SHG funds must produce an explicit blocking conflict.
- Readiness must not accept the first qualifying fund and ignore the rest.
- Snapshot capture must preserve the conflict.
- Validation must remain blocked.

Do not invent cross-fund aggregation rules.

If additional employee eligibility information is required, identify it as a separate evidence prerequisite.

Do not infer ethnicity or fund eligibility from missing or unrelated fields.

## 9.4 SHG negative evidence

Complete direct-SQL tests for:

- Cross-employee correction.
- Cross-employer correction.
- Invalid successor.
- Multiple successors.
- Partial-period instructions.
- Backdated instruction affecting an approved period.
- Post-snapshot SHG drift.
- Rejected function calls leaving no partial writes.
- Rejected operations leaving no audit residue.

SHG and attendance evidence must continue to use the correct payroll-input synchronization barrier.

---

# 10. Priority 6 — Repair rule-pack source and verification trust

This is the trust foundation for statutory implementation.

Inspect:

- `internal/rules/rules.go`
- `cmd/workforceops/admin.go`
- `rules/*/manifest.json`
- `rules/*/sources/`
- Review and specification artifacts.
- Fixture gate artifacts.
- `rule_pack` registry and activation history.
- Rule-pack database functions.

## 10.1 C1-SOURCES — Bind manifest source identities to actual bytes

Current problem:

The manifest declares source digests, but the verifier does not read and hash the actual source files.

A fabricated digest can therefore be accepted when its format is valid and `SOURCES.md` merely exists.

### Required correction

Extend the manifest schema with explicit source-file references.

For each source:

- Stable source ID.
- Relative source artifact path.
- SHA-256 digest.
- Source authority/identity.
- Retrieval metadata where required.

The verifier must:

1. Resolve the path under the package root.
2. Reject path traversal.
3. Reject missing files.
4. Reject symlinks or verify their resolved target remains inside the package root.
5. Require a regular file.
6. Read the actual bytes.
7. Recompute SHA-256.
8. Compare exact bytes to the declared digest.
9. Include source artifacts in the bundle digest.
10. Reject mismatches before activation.

Do not rely on regex-valid digest strings or file-presence checks.

Do not trust `SOURCES.md` as proof of the underlying files.

The bundle digest must cover all substantive verification artifacts, including source bytes, manifest, specification, review and fixture/gate artifacts as defined by the package contract.

## 10.2 C1-GOLDENS — Never hardcode GOLDENS_VERIFIED

Current problem:

The verifier sets `GOLDENS_VERIFIED` when fixture files are merely present and readable.

That is not verification.

### Required correction

Use one of these two approaches:

**Preferred:** execute the package's actual golden, anti-golden and boundary fixture verification gate and derive the verification status from its result.

**Alternative:** consume a separately produced, immutable, digest-pinned independent gate artifact proving that the fixtures were executed or independently verified.

In either case:

- Parse and validate fixture structure.
- Reject malformed or unrecognized fixtures.
- Reject failing fixture verification.
- Verify reviewer and evidence references.
- Pin the exact gate artifact digest.
- Derive `GOLDENS_VERIFIED` from successful evidence, never a hardcoded default.

Do not pretend a fixture has been executed when no execution or accepted external verification has occurred.

The absence of a payroll calculator does not justify hardcoding the verification label.

Keep the rule-pack lifecycle honest.

## 10.3 C2-REPIN — Protect the reviewed registry trust root

Current problem:

The reviewed version and digest can be changed directly through an unaudited administrative SQL operation.

That allows an existing activation function to trust a different version without preserving the previous pin.

### Required correction

Establish an audited pin ceremony.

Every initial pin or repin must preserve:

- Previous version and digest.
- New version and digest.
- Rule-pack identity.
- Jurisdiction and scope.
- Effective period.
- Actor attribution.
- Reason.
- Exact source bundle digest.
- Formal specification version.
- Independent review reference.
- Golden/gate verification reference.
- Timestamp.

The pin operation must:

- Validate exact registry identity.
- Require a verified on-disk package digest.
- Reject invented versions.
- Reject silent replacement.
- Preserve append-only pin history.
- Serialize concurrent changes.
- Write audit atomically.

Direct registry modification must never silently bypass this history.

Treat `migration_admin` as a privileged database trust boundary. Do not claim protection against a malicious schema owner. Ensure ordinary administrative operations and supported pin changes are auditable. If the required guarantees cannot be established with the current ownership model, explain the exact limitation and request the smallest necessary privilege change.

## 10.4 C2-EVIDENCE-LABELS

The activation function must compare submitted:

- Specification version.
- Verification status.
- Review reference.
- Manifest digest.

with the exact reviewed registry pin.

Reject arbitrary caller-supplied labels that differ from the pinned evidence.

Do not let the activation function record a self-attested verification status inconsistent with the actual package.

## 10.5 C2-EFFECTIVE-COVERAGE

Define rule-pack effective-period coverage.

For the first payroll profile, a required activated pack must cover the full payroll period.

If it only covers part of the period:

- Return a precise blocker.
- Do not silently apply it to the full period.
- Do not infer carry-forward or retrospective validity.

If future periods require multiple versions inside one pay period, that must be a separately specified and verified capability.

The snapshot must pin the exact applicable activation record and period coverage decision.

## 10.6 OT-001 protection

`SG-MOM-OT-001` remains BLOCKED.

Do not:

- Pin it.
- Activate it.
- Change its version to satisfy a test.
- Create fabricated golden fixtures.
- Implement overtime.
- Treat its existence in the registry as proof that it is implementable.

Use isolated synthetic fixture packs to test the activation mechanism.

Do not activate real statutory packages merely because the activation mechanism works.

---

# 11. Priority 7 — Close validation and snapshot governance

Address EF-F1A and EF-F1B.

The `app` role must not directly validate a payroll run.

Revoke direct UPDATE authority for:

- `payroll_run.status`.
- `payroll_run.inputs_hash`.
- `payroll_run.rule_versions`.
- Any equivalent validation-authority fields.

Use a narrowly scoped controlled database function.

The function must verify:

- Actual current snapshot tip.
- Exact content hash.
- Rule-version pins.
- Snapshot schema version.
- At least one expected employee.
- No unresolved population blockers.
- Terminal inclusion/exclusion decision for every expected employee.
- No invalidated or superseded snapshot.
- Fresh validation offer.
- No source drift.
- Required rule-pack eligibility.
- Atomic validation audit.

A zero-employee snapshot must not pass through vacuous truth.

A placeholder hash must never be accepted.

The validation function must derive or verify the actual snapshot content rather than trusting arbitrary caller-supplied fields.

Only `draft → validated` is reachable.

Calculated, approved and finalized states remain structurally unavailable without genuine financial artifacts.

Do not fabricate calculation artifacts.

---

# 12. Verification and regression tests

All critical verification must use real PostgreSQL 16.

Every direct SQL adversarial test must authenticate as `app`, except explicitly controlled administrative pin-ceremony tests.

## A. Snapshot and MVCC tests

Add writer-first races for:

1. Employee creation.
2. Compensation insertion.
3. Rule-pack activation/pinning.
4. Timesheet approval.
5. Employment-term closure.
6. Clock-event disposition.
7. Employee status-history mutation.
8. Draft timesheet-line mutation.
9. Employer/pay-period header mutation where still permitted.

For each:

- Writer acquires its barrier and begins the mutation.
- Capture/validation attempts its pre-BEGIN session barrier.
- Writer commits.
- Capture/validation begins the REPEATABLE READ transaction.
- Newly committed data must be visible or the operation must safely fail.
- No missing employee or input may validate.

Add the T2 validation-level proof.

Add:

- Capture-first/writer-second interleaving.
- Correction-during-capture.
- Dual-capture race.
- Dual-invalidation race.
- Session-lock release after success.
- Session-lock release after rollback.
- Session-lock release after error.
- Connection cancellation and discard behavior.
- SQLSTATE 40001 mapping.

Use `pg_locks` to prove actual waiting and acquisition order.

## B. Snapshot fabrication tests

As `app`:

- Direct header INSERT rejected.
- Direct employee-child INSERT rejected.
- Direct pack-child INSERT rejected.
- Direct invalidation INSERT rejected.
- Direct validated-run transition rejected.
- Placeholder hash rejected.
- Empty population rejected.
- Missing decision rejected.
- Post-validation pin update rejected.

The controlled capture and validation paths must still work.

## C. Compensation tests

Use real service and database tests for:

- Widening successor.
- Narrowing successor.
- Surviving predecessor tail.
- New compensation over freed and non-freed ranges.
- Sequential correction.
- Disjoint successor.
- Mid-period compensation change.
- Readiness blocker.
- Snapshot blocker.
- Validation blocker.
- Historical period resolution.

Perform the three live service-as-app probes required by Prompt 16.

A successful compensation write must be resolvable for every period it claims to cover.

## D. SHG tests

Test:

- Active with verification evidence.
- Active without verification evidence.
- Pending.
- Verified-none.
- Opt-out.
- Missing.
- Partial-period start.
- Partial-period end.
- Complete coverage.
- Multiple adjacent instructions.
- Gap.
- Supersession during period.
- Cross-fund ambiguity.
- Post-approval update.
- Post-snapshot drift.
- Cross-employee and cross-employer successor.
- Audit atomicity.

## E. Rule-pack tests

Test:

- Fabricated source digest rejected.
- Missing source file rejected.
- Source-byte mismatch rejected.
- Path traversal rejected.
- Symlink escape rejected.
- Failing golden rejected.
- Unparseable fixture rejected.
- Missing independent gate artifact rejected.
- Arbitrary version rejected.
- Registry repin preserved in audit history.
- Forged evidence label rejected.
- Period coverage mismatch rejected.
- Concurrent pin/supersession has one valid winner.
- Real OT-001 remains unpinned and BLOCKED.

## F. CLI/API differential

Run the same supported scenario through CLI and REST against disposable databases.

Compare:

- Persisted business records.
- Snapshot IDs and substantive content.
- Content hashes.
- Validation result.
- Blocker details.
- Error envelope.
- Request IDs.
- Audit results.

Include at least one invalid operation and one concurrent pair.

Publish the complete 19-package test-bearing list and per-run skip table. Explain the seven packages without test files.

## G. Clean database acceptance

Goose does not automatically prove the contents of already-applied migration files remain identical.

Prevent tests from silently running against stale database schemas.

Use either:

- Unique fresh database names for every acceptance run; or
- A reliable automated drop/recreate procedure restricted to explicitly disposable test databases; or
- A migration-content checksum guard.

Do not rely only on prose instructing developers to drop `a1test*` databases.

The acceptance run must use a provably clean PostgreSQL server and fresh database names.

Run:

- `task lint`
- `task build`
- `GOWORK=off go test -count=1 ./...`
- Full suite under `TZ=UTC`.
- Full suite under `TZ=Asia/Singapore`.
- Relevant race tests.
- PostgreSQL integrity suite.
- All writer-first tests.
- All direct `app` SQL tests.
- Clean-server provisioning.
- CLI/API differential.
- OpenAPI contract tests.
- `git diff --check`.

Do not count skipped DB tests as passes.

---

# 13. Deferred work

The following are not part of this milestone unless a demonstrated dependency requires otherwise:

- A2 compensation-term provenance expansion, if existing lineage remains sufficient.
- C2-ADMIN-DIRECT-WRITE.
- C2-HISTORY-STATUS cosmetic lifecycle status.
- Generic list pagination.
- Generalized idempotency infrastructure.
- Symlink hardening beyond source-package verification directly required here.
- Full `payroll_line` governance.
- CPF calculation.
- SDL calculation.
- SHG contribution calculation.
- Overtime.
- Gross/net salary.
- Payslip financial generation.
- Multi-country packages.
- RLS/SSO.
- UI.
- Microservices.
- Event bus.
- Advanced WFM.
- Production deployment.

`payroll_line` is harmless today because no production calculation path uses it.

However, before the first calculator writes financial lines, it becomes a mandatory prerequisite to implement:

- Snapshot FK and version.
- Verified line-level provenance.
- Exact monetary representation.
- Post-calculation immutability.
- Versioned recalculation/supersession.
- Revocation of direct `app` writes.

Do not introduce payroll-line behavior during this milestone.

---

# 14. Required report

Create:

`plans/milestone2_1_remediation_report.md`

Include:

1. Git baseline and preservation evidence.
2. Root-cause analysis for each original finding.
3. Invariants identified.
4. All write paths examined.
5. MVCC and session-lock protocol.
6. Writer-to-barrier matrix.
7. Snapshot and validation privilege changes.
8. Compensation resolver and liveness correction.
9. SHG state and full-period coverage decisions.
10. Rule-pack source and golden verification.
11. Audited pin ceremony.
12. Migration changes.
13. Direct `app` SQL results.
14. Live concurrency interleavings.
15. CLI/API differential results.
16. Test-bearing package list and skip table.
17. Confirmed remaining defects.
18. Explicitly deferred findings.
19. Environment limitations.
20. Final PASS/FAIL decision.

For every corrected finding, identify:

- Original finding ID.
- Root cause.
- Invariant.
- Affected write paths.
- Database enforcement.
- Application behavior.
- Regression test.
- Persisted-state evidence.

Correct any inaccurate claims in the Prompt 16 report.

Do not mark a finding fixed merely because a corresponding test file exists.

---

# 15. Acceptance conditions

Milestone 2.1 PASS requires all of the following:

1. Snapshot capture acquires all required session barriers before the REPEATABLE READ view is established.
2. Validation uses the same safe protocol.
3. Every relevant population and employee input writer participates in the synchronization contract or is demonstrably unable to alter the captured inputs.
4. Writer-first races prove newly committed changes cannot be omitted from a validated snapshot.
5. App cannot fabricate, append or alter snapshot contents.
6. App cannot directly validate a run with placeholder, stale or empty pins.
7. Empty employee population cannot validate.
8. Compensation widening and free-tail corrections remain resolvable.
9. Unsupported mid-period compensation changes block consistently.
10. SHG active verification semantics are explicitly decided and enforced.
11. SHG evidence fully covers the pay period or produces an exact blocker.
12. Ambiguous cross-fund inputs block unless an independently supported rule permits them.
13. Actual source bytes are verified against manifest digests.
14. Golden verification is real and not hardcoded.
15. Registry pin history and repinning are auditable.
16. Rule-pack effective dates are checked against the payroll period.
17. OT-001 remains BLOCKED.
18. CLI/API differential passes.
19. Clean-server tests cannot silently reuse stale schemas.
20. No P0 or material P1 defect remains that can produce an incomplete, inconsistent, fabricated or ungoverned payroll input set.

If any condition fails, report FAIL with exact remaining blockers.

Do not declare PASS solely because Go packages are green.

The database must enforce the invariants under the ordinary application role.

---

# 16. Execution priority

Execute in this order:

1. D1 and P1a–g: synchronization protocol and all writer paths.
2. EF-E3 and EF-F1A/B: controlled snapshot capture and validation.
3. Compensation resolver: WIDEN, FREETAIL and MIDPERIOD.
4. SHG decisions and period coverage.
5. Rule-pack source/golden verification and pin governance.
6. Full live concurrency, direct SQL, CLI/API and clean-server verification.

Do not start a separate redesign phase.

Do not spend time on P2/P3 findings that do not affect the restricted first payroll scenario.

---

# 17. Final principle and stop condition

The purpose of this milestone is not to make the current test failures disappear.

It is to make the underlying failures impossible within the application's documented threat model.

For every change, ask:

**What invariant was previously unenforced, which operations could violate it, and what now makes every such violation impossible or explicitly blocked?**

A correction is incomplete if:

- The service path is fixed but direct SQL remains unsafe.
- The capture path is fixed but validation remains stale.
- The data hash is correct but child rows remain writable.
- The source digest is syntactically valid but the source bytes are not checked.
- An instruction exists but does not cover the period.
- A compensation correction is writable but cannot be resolved.
- A test passes without reproducing the original failure.

Do not add generic frameworks or overengineer unrelated concerns.

Preserve all existing valid work.

Keep changes uncommitted.

Stop when the acceptance conditions pass.

The next milestone is the first independently verified Singapore monthly payroll calculation. It must use immutable payroll inputs and genuinely verified CPF, SDL and applicable SHG rule versions.

**Fix the causes, enforce the invariants at their authoritative boundaries, prove the fixes adversarially, and then move forward.**

# 3. The underlying lessons we should carry forward

There are five deeper lessons in this review that I want to become part of our engineering culture.

First: correctness is about ordering, not just locking. A lock acquired at the wrong point in a transaction can provide a false sense of safety. We must understand the database's actual execution and visibility rules.

Second: immutability requires exclusive authority over creation as well as modification. Preventing UPDATE and DELETE is insufficient when an ordinary application role can insert fabricated child records.

Third: a resolver and a writer must share the same definition of effective data. If one permits a record that the other cannot interpret, the system can store data that is permanently unusable.

Fourth: evidence must be verified, not described as verified. A manifest digest, a `GOLDENS_VERIFIED` label or an audit row is useful only when it is tied to the actual source files, test outcomes and reviewed version.

Fifth: a valid record does not automatically provide complete coverage. This applies to employment terms, compensation, SHG instructions, rule-pack effective dates and payroll populations. The system must evaluate the complete applicable period, not simply find one matching row.

These lessons are more important than any particular migration or function.

My recommendation is to authorize Prompt 17. This should be the final focused closure of the Milestone 2 acceptance defects. After a clean independent PASS, we can proceed to the first restricted Singapore payroll calculation without reopening the broader architecture.