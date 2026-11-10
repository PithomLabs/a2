I recommend naming this Milestone 2 — Payroll Governance and Reproducibility Foundation.

This milestone implements the six prerequisites established by the latest adversarial review. It does not yet calculate payroll. Its purpose is to make payroll runs governable, freeze their inputs, resolve the correct compensation and SHG records, and ensure only properly verified rule versions can be activated.

The six requirements are grounded in the review's findings C1–C6, SHG-B1/B2 and S5-F1.&#x20;

pre_milestone2_adversarial_review\.md



# Prompt 15 — WorkforceOps Milestone 2

# Payroll Governance and Reproducibility Foundation

## 1. Objective

Implement the six prerequisites identified in:

`plans/pre_milestone2_adversarial_review.md`

The objective is to establish the infrastructure required for the first independently verified Singapore payroll calculation.

The six deliverables are:

1. **P1 — Governed payroll-run lifecycle.**
2. **P2 — Immutable per-run input snapshots.**
3. **P3 — Effective compensation resolver.**
4. **P4 — Complete first-path SHG instruction lifecycle.**
5. **P5 — Controlled, evidence-backed statutory rule-pack activation.**
6. **P6 — Independent regression and integration tests.**

The existing operational foundation is substantially complete:

`Employer → Employee → Employment Terms → Compensation → Roster → Attendance → Timesheet → Payroll Readiness`

Milestone 2 extends this workflow into:

`Payroll Run → Validate → Snapshot → Review/Approval Gates → Reproducibility`

**No statutory payroll calculations are authorized in this milestone.**

Do not implement CPF, SDL, SHG contribution amounts, overtime, gross earnings, deductions, net pay or financial payslips.

The objective is to make the system capable of preparing and governing a reliable payroll calculation without pretending that the calculation itself has been completed.

---

# 2. Repository baseline and preservation

Before implementation:

1. Inspect `git status --short`.
2. Record `git rev-parse HEAD`.
3. Confirm the expected baseline:\
   `f43515883668c917edac31be49a7cf8334eca5c2`.
4. Record all staged, unstaged, deleted, renamed and untracked entries.
5. Read `AGENTS.md`.
6. Read `plans/milestone1_3_remediation_report.md`.
7. Read `plans/pre_milestone2_adversarial_review.md` in full.
8. Inspect all six canonical prerequisite findings.
9. Inspect migrations `00001` through `00011`.
10. Inspect current payroll, HR, SHG, rule-pack, attendance and timesheet services.

Preserve all existing uncommitted work.

Do not:

- Reset or clean the repository.
- Stage or commit changes.
- Discard previous work.
- Modify `bchat/`.
- Modify upstream Phys-Research or Phys-Pseudo.
- Alter migrations `00001`–`00011`.
- Drop or recycle an existing user database.
- Deploy to Neon or Fly.io.
- Implement statutory calculations.

Use disposable PostgreSQL 16 databases for testing.

All changes remain uncommitted.

Use one new forward migration beginning at `00012`, provided the changes can be implemented coherently in one migration. If a second migration is genuinely required, explain why before introducing it.

---

# 3. Frozen architectural decisions

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
- No distributed processing.
- No RLS in v0.1.
- No generic plugin framework.
- No generic payroll rules engine.
- Singapore-first jurisdiction configuration.

Retain Phys-Research and Phys-Pseudo as the methodology for high-consequence business rules.

Do not introduce another country package.

The purpose is to complete the existing Singapore backend, not generalize it prematurely.

---

# 4. Execution dependency order

Implement the work in this order:

1. P3 — Effective compensation resolver.
2. P4 — SHG instruction lifecycle.
3. P5 — Controlled rule-pack activation.
4. P2 — Immutable payroll-input snapshot.
5. P1 — Governed payroll-run lifecycle.
6. P6 — Full integration and regression testing.

This order matters.

A snapshot cannot reliably pin compensation until the effective compensation resolver exists. It cannot pin SHG until the SHG lifecycle is complete. It cannot pin trustworthy statutory versions until rule-pack governance is established.

P1's state-transition guards may be developed in parallel, but approval and finalization must not become reachable before a valid snapshot and, where required, a real calculation artifact exist.

Do not calculate payroll merely to demonstrate the new lifecycle.

---

# 5. P3 — Implement the effective compensation resolver

The adversarial review established that compensation correction preserves both original and successor records, but existing retrieval does not provide one authoritative effective value for an employee, compensation classification and pay period.

Inspect:

- `internal/hr/hr.go`
- Compensation models and repository queries.
- `add_compensation`.
- `correct_compensation`.
- Compensation successor constraints.
- Migration `00011`.
- Payroll readiness.

## A. Implement the resolver

Create one shared application-service/repository operation conceptually equivalent to:

`ResolveEffectiveCompensation(employer, employee, period)`

It must resolve compensation by:

- Employer.
- Employee.
- Compensation classification/code.
- Applicable effective period.
- Supersession chain.
- Historical validity.

Return only the effective record for each supported compensation classification.

The resolver must:

- Exclude superseded predecessors.
- Include the valid current successor.
- Respect effective start/end dates.
- Validate parent employment-term containment.
- Avoid combining overlapping historical and successor values.
- Detect ambiguous or invalid chains.
- Reject multiple effective values where the model requires one.
- Never select an arbitrary row using `LIMIT 1`.

Use exact monetary representations.

## B. Correction-chain semantics

Test:

1. Original compensation without a successor.
2. Original followed by one successor.
3. Successor followed by another successor.
4. Attempted competing successors.
5. Attempted correction across employees.
6. Attempted correction across employers.
7. Attempted cycle.
8. Superseded compensation outside the current period.
9. Correction effective only in a future period.
10. Historical compensation relevant to an earlier period.

A valid chain should resolve one effective value without rewriting the original.

If a chain is invalid or multiple effective records exist, return a machine-readable conflict.

## C. Use the resolver consistently

Replace direct compensation-list assumptions in payroll readiness and future snapshot capture.

Do not maintain independent "effective compensation" implementations in different packages.

Provide an explicit result containing:

- Effective compensation record ID.
- Compensation classification.
- Effective date interval.
- Exact amount.
- Parent employment-term ID.
- Supersession lineage reference.

## D. Acceptance

A single effective compensation basis must be deterministically resolved for every supported employee and pay period.

An ambiguous basis must block readiness and snapshot validation.

No payroll amount may be calculated.

---

# 6. P4 — Complete the first-path SHG instruction lifecycle

The review found that the existing SHG implementation cannot reliably distinguish missing information, verified non-applicability and applicable instructions.

It also allows post-approval changes that can silently alter the historical payroll basis.

Inspect:

- `internal/hr/shg.go`
- SHG migrations and constraints.
- SHG instruction privileges.
- `internal/payroll/readiness.go`
- Existing payroll-input locks and basis sequences.
- Employee and employment-term relationships.
- SHG tests.

## A. Represent instruction states explicitly

Support a deliberate first-path status model:

- `pending`
- `active`
- `verified_none`
- Existing `opt_out` only where its meaning and evidence are explicitly defined.

Do not treat these statuses as interchangeable.

The system must distinguish:

| State                 | Readiness meaning                                       |
| --------------------- | ------------------------------------------------------- |
| Missing               | Blocked                                                 |
| Pending               | Blocked                                                 |
| Active but unverified | Blocked                                                 |
| Active and verified   | Eligible for downstream applicable-rule evaluation      |
| Verified none         | Eligible only with complete supporting evidence         |
| Opt-out               | Requires explicitly established and supported semantics |
| Superseded            | Historical, not effective                               |

A missing row must never automatically mean verified non-applicability.

Do not infer employee eligibility from missing records.

## B. Verified-none evidence

Make `verified_none` recordable through the normal supported service.

Require:

- Employer and employee.
- Effective period.
- Verification date.
- Actor attribution.
- Evidence reference.
- Provenance/basis.
- Reason.

The existing CHECK constraint must be satisfied through the actual supported workflow.

Use machine-readable validation and readiness outcomes.

Do not create a verified-none record merely to make a demo employee payroll-ready without actual supporting evidence.

## C. Governed instruction correction and succession

Implement a controlled SHG instruction lifecycle.

The original instruction must remain historically visible.

A correction must:

1. Identify the original instruction.
2. Create an immutable successor.
3. Preserve the original.
4. Enforce one successor per predecessor.
5. Preserve employer and employee identity.
6. Enforce effective-date consistency.
7. Prevent overlapping effective instructions of the same fund/classification where prohibited.
8. Record actor, reason and provenance.
9. Participate in the payroll-input barrier.
10. Create an explicit readiness conflict when it affects an approved period.

Do not use arbitrary UPDATE or DELETE to correct historical instructions.

## D. Database protections

For `shg_instruction`:

- Revoke unrestricted application-role UPDATE and DELETE.
- Provide controlled insertion/correction operations.
- Ensure every permitted INSERT acquires the shared payroll-input lock.
- Assign `basis_seq` through the database-controlled mechanism.
- Normalize record-creation timestamps.
- Persist one canonical audit event atomically.
- Prevent direct SQL from bypassing post-approval conflict detection.

Verify direct SQL under `app`.

The service path and direct database path must preserve the same historical guarantees.

## E. SHG readiness

Replace presence-only checking with status-aware evaluation.

Return machine-readable blockers, including appropriate equivalents of:

- `SHG_INSTRUCTION_MISSING`
- `SHG_INSTRUCTION_PENDING`
- `SHG_INSTRUCTION_UNVERIFIED`
- `SHG_INSTRUCTION_CONFLICT`
- `SHG_CORRECTION_REQUIRES_REISSUE`
- `SHG_INSTRUCTION_INVALID`

A verified-none instruction is complete only when its evidence, actor and effective date are valid.

An `opt_out` row must not automatically become verified non-applicability.

If legal semantics for a particular status or fund combination are unresolved, mark that case unsupported and block the corresponding payroll profile.

## F. Scope of SHG behavior

Do not invent cross-fund aggregation rules.

The current review explicitly leaves cross-fund semantics for Phys-Research in Milestone 2's statutory research work.

For the first supported payroll profile:

- Require an explicitly evidenced applicable SHG instruction or verified-none result.
- Reject ambiguous multiple-fund combinations until the relevant Singapore rule package establishes their treatment.
- Do not calculate SHG deductions in this milestone.

## G. Acceptance

Demonstrate:

- Missing and pending states block readiness.
- Applicable verified instructions are distinguishable.
- Verified-none is recordable and evidence-backed.
- Supersession preserves historical records.
- Post-approval changes produce a visible conflict.
- Direct SQL cannot silently rewrite approved SHG inputs.
- The payroll-input sequence detects relevant changes.
- All tests use real PostgreSQL and `app`.

---

# 7. P5 — Controlled, evidence-backed rule-pack activation

This is a critical governance requirement.

The adversarial review found that `app` could directly change a rule pack from `BLOCKED` to `IMPLEMENTABLE` using an invented version.

That must become impossible.

Inspect:

- `rules/*/manifest.json`
- Rule-pack activation schema.
- Rule-pack parameter tables.
- Existing read-only activation/status API.
- `internal/payroll/readiness.go`
- `internal/payroll/approval.go`
- Phys-Research and Phys-Pseudo adaptation documents.
- Existing pack lifecycle tests.

## A. Database enforcement

The application role must not directly INSERT or UPDATE:

- `rule_pack_activation`
- Statutory parameter tables.
- Rule-pack effective-version records.
- Any equivalent authoritative activation record.

Provide a narrowly scoped controlled activation mechanism.

It must validate:

- Known rule-pack identity.
- Exact version.
- Effective dates.
- Jurisdiction.
- Rule-pack scope.
- Manifest digest.
- Source/evidence references.
- Specification version.
- Independent review record.
- Verification status.
- Actor attribution.
- Reason for activation.

A pack cannot become IMPLEMENTABLE merely because a caller supplies the string `IMPLEMENTABLE`.

## B. Activation state transitions

Implement a controlled transition policy.

At minimum:

- `BLOCKED` remains the default.
- Only a qualified reviewed version may transition to `IMPLEMENTABLE`.
- Invalid or missing evidence is rejected.
- A different version requires a separate reviewed activation record.
- An approved activation cannot be silently rewritten.
- Superseding a version preserves the original activation history.
- Concurrent activation attempts are deterministic.
- The previous applicable version remains historically identifiable.

The exact implementation should follow existing rule-package lifecycle conventions rather than create a new generic governance framework.

## C. Activation privilege boundary

The v0.1 application has no authentication system. `X-Operator` and CLI operator labels are attribution, not identity verification.

Do not imply otherwise.

Normal `app` operations and the HTTP API must not be sufficient to activate a statutory pack.

Use the existing controlled administrative database privilege boundary for activation:

- Keep ordinary application reads available as appropriate.
- Revoke direct activation and parameter-table mutation from `app`.
- Expose activation only through an explicitly controlled administrative CLI operation using the designated privileged connection.
- Require explicit actor, reason, version and evidence references.
- Audit the activation atomically.
- Do not expose an activation endpoint through the normal REST server.
- Do not repurpose an operator string as authorization.

Do not introduce a third runtime service or dynamic plugin mechanism.

If using the existing `migration_admin` credential for a controlled offline activation operation, keep that operation separate from schema migration commands and document its elevated privilege implications.

## D. Rule evidence verification

The activation workflow must inspect the actual rule package and its declared evidence metadata.

The package must not claim IMPLEMENTABLE if:

- Its authoritative source requirements are incomplete.
- Its formal specification has unresolved load-bearing questions.
- Independent review is incomplete.
- Required fixtures are missing.
- Required verification status has not been reached.

Preserve the Phys-Research and Phys-Pseudo lifecycle.

The database should pin the evidence/version identity. Do not pretend that a hash alone proves a legal interpretation is correct.

## E. Parameter integrity

Statutory parameter records must be bound to:

- Exact rule-pack ID.
- Exact version.
- Jurisdiction.
- Effective interval.
- Currency/units where applicable.
- Evidence references.

The `app` role must not directly introduce an invented parameter version.

Avoid generic arbitrary JSON as a replacement for meaningful parameter validation.

## F. OT-001 protection

`SG-MOM-OT-001` remains BLOCKED.

Do not:

- Activate it.
- Invent its missing statutory source.
- Populate golden fixtures using unresolved interpretations.
- Implement overtime calculations.
- Change its manifest to make tests pass.

Use a separate isolated test fixture pack to test the activation mechanism.

Other statutory packs must remain in their actual current states unless their evidence independently qualifies them. Do not automatically activate CPF, SDL or SHG rules merely because the activation mechanism works.

---

# 8. P2 — Immutable per-run input snapshots

This is the central deliverable of Milestone 2.

The current implementation creates draft runs but does not capture a reproducible input set.

The existing `inputs_hash` is not content-bound and does not distinguish different employee or rule versions.

Implement a real immutable snapshot model.

Inspect:

- `internal/payroll/runs.go`
- `internal/payroll/readiness.go`
- `internal/payroll/approval.go`
- Effective compensation resolver.
- Employment-term and status history.
- Timesheet correction chains.
- SHG instruction lifecycle.
- Rule-pack activation.
- `pay_basis_seq`.
- Existing employer population logic.

## A. Expected employee population

Derive the expected employee population from effective-dated employment applicability for the requested period.

Do not derive it solely from employees with approved timesheets.

Capture separately:

- Expected employees.
- Employees with eligible approved timesheets.
- Employees missing timesheets.
- Employees with incomplete inputs.
- Explicitly excluded employees.
- Employees with historical term conflicts.
- Employees with status changes affecting the period.

An employee who has been terminated after the pay period must not silently disappear from that historical population.

## B. Define a coherent snapshot

Each payroll run must have an immutable snapshot containing:

### Employer and run identity

- Employer ID.
- Jurisdiction.
- Currency.
- Payroll period.
- Payroll run ID.
- Snapshot schema version.
- Snapshot creation metadata.

### Employee population

For each expected employee:

- Employee ID.
- Effective employment-term version.
- Relevant employee-status record/version.
- Employee pay-basis sequence.
- Inclusion/exclusion decision.
- Explicit blocker references.

### Compensation

- Effective compensation line IDs.
- Exact compensation classifications.
- Effective dates.
- Exact amounts.
- Supersession lineage.
- Resolved compensation result.

Never snapshot all original and successor rows as though they were all payable inputs.

### Timesheets and attendance

- Effective approved timesheet ID/version.
- Correction lineage.
- Relevant attendance evidence references.
- Effective business-date intervals.
- Relevant conflict and exception state.

### SHG

- Exact instruction record ID/version.
- Effective status.
- Verification evidence reference.
- Effective dates.
- SHG applicability result or explicit unresolved blocker.

### Rule packs

For every required package:

- Rule-pack ID.
- Exact version.
- Activation status.
- Manifest digest.
- Effective interval.
- Review/evidence reference.

### Calculation contract

Record:

- Snapshot schema version.
- Intended calculation scope.
- Calculation contract/specification version where established.

Do not invent an executable calculator version.

Since no calculation engine exists in this milestone, explicitly represent the calculation engine as unimplemented/unavailable. A snapshot cannot be marked as having been calculated or finalized.

When the real calculation engine is introduced, its exact implementation/version must be pinned separately.

## C. Snapshot capture must be atomic

A snapshot must not combine inconsistent employee versions.

Capture it within one controlled database transaction.

Requirements:

1. Acquire the employer payroll-population barrier.
2. Derive the expected employee set.
3. Acquire employee-level pay-basis locks in deterministic UUID order.
4. Revalidate the population after locks.
5. Resolve effective employment, compensation, timesheet and SHG inputs.
6. Capture required rule-pack versions.
7. Assess all blockers.
8. Persist snapshot headers and per-employee records.
9. Calculate the canonical content hash.
10. Commit atomically.

The employer-level population barrier is necessary because locking only the employees already discovered does not prevent a concurrent new employment record from entering the expected population during capture.

All operations capable of changing the expected employee population must participate in the same population-barrier convention.

At minimum inspect:

- Employee creation.
- Employment-term creation and changes.
- Employee-status transitions.
- Any other supported population-affecting operation.

Define one lock-order convention across:

- Employer population lock.
- Employee pay-basis locks.
- Relevant secondary locks.

Avoid lock-order inversion and deadlocks.

If an existing isolation or locking mechanism already proves the population stable, use it and demonstrate the proof. Do not add unnecessary locks merely for appearance.

## D. Content-bound hash

Replace the placeholder hash with a canonical SHA-256 digest over the complete snapshot content.

The canonical form must:

- Use stable field ordering.
- Sort employee records deterministically.
- Sort compensation and reference collections deterministically.
- Preserve exact monetary values.
- Preserve dates, timestamps and timezone identities.
- Include exact rule-pack identities and versions.
- Include correction lineage.
- Include inclusion/exclusion decisions.
- Include the snapshot schema version.

Do not hash only employer ID, period and run kind.

Two runs with different actual inputs must not have identical hashes merely because their employer and pay period match.

Two snapshots containing exactly identical canonical inputs and versions should produce the same content hash.

Store both:

- Snapshot identity.
- Content hash.

A hash is an integrity fingerprint, not a substitute for retaining the actual immutable snapshot.

## E. Snapshot immutability

After validation:

- Snapshot contents are immutable.
- Employee membership is immutable.
- Input references are immutable.
- Rule-pack pins are immutable.
- Content hash is immutable.
- No child row can be silently re-parented into or out of the snapshot.

The database must enforce this under `app`.

Any new source change must not silently rewrite an existing snapshot.

## F. Snapshot drift detection

At validation and before a later approval/finalization transition, compare the snapshot with the current source state.

Detect:

- Employee population drift.
- Employment-term changes.
- Status changes affecting the period.
- Compensation successor changes.
- Attendance correction.
- Late-break evidence.
- Effective timesheet version changes.
- SHG instruction changes.
- Rule-pack activation/version changes.

If the input set no longer matches:

- Mark the snapshot as invalidated or create a new snapshot version through a controlled operation.
- Preserve the original snapshot.
- Record the exact drift reason.
- Prevent approval/finalization from proceeding using stale inputs.

Do not silently update a snapshot in place.

## G. Multi-employee consistency

The pay-basis sequence is per employee, not a global transaction timestamp.

Never compare two employees' `basis_seq` values as though they share a global ordering.

Capture each employee's sequence alongside that employee's exact input IDs.

Use employer-level population synchronization plus deterministically sorted employee locks to establish one coherent population boundary.

Prove that a concurrent:

- Employee creation.
- Employment-term change.
- Termination.
- Compensation change.
- Timesheet approval.

cannot produce a mixed-version snapshot.

## H. Acceptance

Demonstrate that two payroll runs using different source inputs produce different hashes and immutable snapshots.

Demonstrate that identical inputs produce identical canonical hashes.

Demonstrate that source changes after snapshot creation do not mutate the snapshot and instead produce an explicit drift/invalidation result.

---

# 9. P1 — Implement governed payroll-run lifecycle

The current run implementation supports creation and exclusions but does not enforce a complete validation/approval lifecycle.

Inspect:

- `internal/payroll/runs.go`
- `internal/payroll/approval.go`
- Existing status CHECK constraints and triggers.
- Snapshot implementation from P2.

## A. Define the state model explicitly

Use a controlled lifecycle with unambiguous meanings.

The minimum states should distinguish:

- `draft`
- `validated`
- `calculated`
- `approved`
- `finalized`

An invalidated or superseded state may be included if the snapshot lifecycle requires it.

Define the exact permitted transition graph.

Do not allow:

- Draft directly to approved.
- Validated back to draft.
- Approved back to draft.
- Reopening a finalized run.
- Multiple effective approved runs for the same employer/period/run kind without a controlled correction relationship.

## B. Respect the no-calculation boundary

There is no calculator in this milestone.

Therefore:

- Draft creation is available.
- Validation captures an immutable snapshot and verifies readiness.
- A validated run must have exactly one valid snapshot version.
- `calculated` cannot be entered without a real calculation artifact.
- `approved` cannot be entered without a real calculated result and appropriate review evidence.
- `finalized` cannot be entered without an approved calculation result and finalization metadata.

Do not fake a calculation result, placeholder amount or synthetic successful approval.

If the business requires an input-snapshot approval before calculation, represent that as a separately named snapshot approval state. Do not confuse it with approval of financial payroll.

Expose only operations that are genuinely implemented.

## C. Validation service

Implement a single shared validation service.

It must:

1. Acquire the required population and employee locks.
2. Recompute expected population.
3. Resolve effective compensation.
4. Resolve effective timesheets.
5. Validate SHG inputs.
6. Verify required rule-pack activation.
7. Verify no unresolved attendance or employment conflicts.
8. Capture the immutable snapshot.
9. Compute content hash.
10. Record a detailed validation result.
11. Transition the run from draft to validated only when all mandatory conditions pass.

If blockers exist:

- Preserve the draft.
- Return machine-readable blockers.
- Do not create a falsely validated snapshot.
- Record the attempted validation outcome if the audit contract requires it.

## D. Database transition enforcement

Use a database trigger or equivalent protection.

The `app` role must not directly:

- Approve a draft.
- Reopen a validated run.
- Approve a run without a snapshot.
- Finalize without a calculated and approved result.
- Change run identity after snapshot capture.
- Alter a validated snapshot.

Every permitted transition must be atomic and auditable.

## E. Run-scoped exclusions

Preserve the existing draft-only exclusion model.

Validate:

- Employer.
- Employee.
- Pay period.
- Run identity.
- Reason.
- Actor.
- Effective status.

An exclusion must not accidentally apply to another payroll run or period.

Empty population must not pass approval through vacuous truth.

## F. Correction runs

A later corrected payroll must not rewrite the original run.

Use a controlled successor relationship between runs or an equivalent versioned correction model.

Ensure a correction run does not become a second unrelated approved regular run for the same period.

Do not implement actual correction amounts or financial adjustments in this milestone.

---

# 10. P6 — Independent regression and integration tests

Testing must prove the six prerequisites together, not merely pass isolated unit tests.

Use real PostgreSQL 16.

All direct SQL adversarial tests must use the actual `app` role.

Use disposable databases and verify their targets.

## A. Compensation resolver

Test:

- Original compensation.
- Single successor.
- Multiple sequential successors.
- Superseded predecessor excluded.
- Correct effective-date selection.
- Cross-period historical resolution.
- Ambiguous effective basis rejected.
- Multiple classifications resolved independently.
- Invalid chain blocked.

## B. SHG lifecycle

Test:

- Missing instruction.
- Pending instruction.
- Active but unverified instruction.
- Applicable verified instruction.
- Verified-none with evidence.
- Opt-out with defined supported semantics.
- Invalid verified-none without evidence.
- Historical instruction.
- Superseding instruction.
- Cross-fund ambiguity.
- Post-approval change.
- Direct SQL UPDATE/DELETE rejection.
- Direct SQL INSERT synchronization.
- Audit atomicity.
- Readiness status and machine-readable blockers.

## C. Rule-pack governance

Test:

- Blocked pack cannot activate.
- Invented version rejected.
- Unknown pack rejected.
- Manifest digest mismatch rejected.
- Missing review evidence rejected.
- Missing golden/verification status rejected.
- Invalid lifecycle transition rejected.
- Correct reviewed version accepted only through controlled administrative path.
- Direct `app` UPDATE/INSERT rejected.
- Parameter mutation by `app` rejected.
- Activation audit preserved.
- Version supersession preserves history.

Use isolated test packs.

Do not activate real statutory packs unless they have independently qualified evidence.

## D. Snapshot integrity

Test:

- Complete employee population.
- Missing timesheet.
- Historical terminated employee.
- Missing compensation.
- Superseded compensation.
- Missing SHG instruction.
- Verified-none SHG.
- Missing required rule pack.
- Blocked rule pack.
- Explicit exclusion.
- Empty population.
- Multiple employees with different basis sequences.
- Concurrent employee creation.
- Concurrent employment-term change.
- Concurrent termination.
- Concurrent compensation insertion.
- Concurrent timesheet approval.
- Concurrent attendance correction.
- Concurrent SHG instruction change.
- Concurrent rule-pack supersession.
- Snapshot hash determinism.
- Different inputs produce different hashes.
- Direct SQL attempts to mutate snapshot rows.
- Post-snapshot source drift.
- Snapshot invalidation.
- Revalidation producing a new snapshot version.

Verify exact persisted state after every operation.

## E. Payroll-run lifecycle

Test:

- Draft creation.
- Draft exclusion.
- Validation with blockers.
- Successful validation and snapshot capture.
- Direct SQL draft-to-approved rejection.
- Direct SQL reopening of validated run rejection.
- Approved transition without a calculation artifact rejection.
- Finalized transition without an approved calculation artifact rejection.
- Duplicate validation replay.
- Concurrent validation.
- Snapshot drift before later approval.
- Corrected-run relationship.
- Multiple approved regular-run prevention.

The absence of calculation code is an intentional invariant, not a test failure.

## F. End-to-end workflow

Using the current synthetic Singapore employer and employee profile:

1. Bootstrap employer.
2. Create or retrieve employee.
3. Establish monthly-rated employment terms.
4. Record compensation.
5. Assign BPO programme/team.
6. Publish roster.
7. Import attendance and break evidence.
8. Generate and approve timesheet.
9. Record a supported SHG instruction or evidence-backed verified-none state.
10. Ensure required statutory rule packs have their actual verified states.
11. Create a payroll run.
12. Validate readiness.
13. Capture immutable input snapshot where all requirements permit validation.
14. Verify snapshot identity and content hash.
15. Attempt source changes and prove the original snapshot is preserved and invalidated when appropriate.

If required statutory rule packs remain blocked, the run must remain blocked. Do not manufacture a successful validation by enabling unverified packs.

## G. Full regression

Run:

- `task lint`
- `task build`
- `GOWORK=off go test -count=1 ./...`
- Full suite under `TZ=UTC`.
- Full suite under `TZ=Asia/Singapore`.
- Race tests for attendance, timesheet, HR and payroll.
- PostgreSQL application-role integrity suite.
- Clean-server provisioning.
- Snapshot concurrency tests.
- Rule-pack activation tests.
- SHG lifecycle tests.
- CLI E2E.
- REST E2E.
- Persisted-state parity.
- OpenAPI validation.
- `git diff --check`.

Do not silently skip database-gated tests.

Do not use superuser execution as proof of application-role correctness.

Record exact commands and outcomes.

---

# 11. Required database and API/CLI changes

Add only the schema required for:

- Payroll input snapshots.
- Per-employee snapshot membership and version references.
- Snapshot hashes and lifecycle.
- Controlled payroll-run transitions.
- SHG instruction lifecycle and succession.
- Controlled rule-pack activation and evidence references.
- Any narrowly required population synchronization mechanism.

Protect all immutable snapshot and activation records under `app`.

Expose minimum CLI and REST operations for:

- Payroll-run creation.
- Payroll-run validation.
- Payroll-run inspection.
- Snapshot inspection.
- SHG instruction creation, verification and controlled correction.
- Rule-pack read-only status.
- Controlled administrative activation via CLI only.

Do not expose statutory rule activation through the normal REST server.

Complex operations should use JSON file input where appropriate.

CLI and REST must share application services.

Every mutation must retain actor attribution and reason where applicable. Do not represent `X-Operator` as authentication.

Ensure OpenAPI accurately reflects implemented operations and states.

---

# 12. Phys-Research and Phys-Pseudo boundaries

Preserve the established methodology.

Do not weaken:

- Source provenance.
- Effective-date requirements.
- Claims and assumptions.
- Formal specification requirements.
- Independent derivation.
- Golden/anti-golden/boundary testing.
- Independent review.
- Rule-pack lifecycle.

No rule becomes IMPLEMENTABLE merely because:

- A version string exists.
- A manifest was edited.
- A calculation looks mathematically plausible.
- A coding agent says it is complete.
- A test passes against its own implementation output.

Rule-pack activation is a governance operation, not a substitute for research.

`SG-MOM-OT-001` remains BLOCKED.

No CPF, SDL or SHG calculation may be implemented during this milestone.

If the existing research evidence is insufficient to activate a package, keep it blocked and report the missing evidence.

---

# 13. Explicitly deferred items

Do not expand Milestone 2 into unrelated product work.

Defer:

- Actual CPF calculation.
- Actual SDL calculation.
- SHG contribution calculation.
- Overtime.
- Gross/net payroll.
- Payslip financial generation.
- Salary payment or banking integrations.
- Statutory submission.
- Payroll taxation.
- Cross-fund SHG semantics until researched.
- Full multi-employer authentication.
- RLS.
- SSO.
- Additional jurisdictions.
- Generic workflow engine.
- Generic rules engine.
- Phys-Pseudo interpreter/compiler.
- Microservices.
- Event bus.
- UI.
- Advanced WFM forecasting.
- Rotation materialization.
- Production deployment.

Record each deferred capability and its dependency.

Do not introduce placeholder financial values or fabricated successful payroll examples.

---

# 14. Completion conditions

Milestone 2 can PASS only when all six prerequisites are complete.

### P1 — Run governance

- Draft-to-validated lifecycle enforced.
- Direct SQL cannot bypass validation.
- Future calculated/approved/finalized states are guarded.
- No fake calculation transition exists.
- Run identity and exclusions are protected.

### P2 — Snapshot

- Expected population is correct.
- Capture is atomic and concurrency-safe.
- Compensation, timesheet, employment, SHG and rule-pack versions are pinned.
- Content hash is genuinely content-bound.
- Snapshot is immutable.
- Source drift is detectable.
- Revalidation produces a new version rather than rewriting history.

### P3 — Compensation resolver

- One effective compensation basis per employee, classification and period.
- Superseded records excluded.
- Ambiguity blocks validation.
- Historical and future periods resolve correctly.

### P4 — SHG

- Missing, pending, applicable and verified-none states are distinct.
- Verified-none is recordable with evidence.
- Readiness is status-aware.
- Post-approval changes are detectable.
- Historical instructions are protected.
- Controlled correction/supersession exists.
- No unsupported statutory inference is introduced.

### P5 — Rule-pack activation

- Direct app-role activation is impossible.
- Administrative activation is controlled and auditable.
- Pack identity, version and evidence are pinned.
- Unsupported versions and invalid transitions are rejected.
- OT-001 remains BLOCKED.

### P6 — Verification

- All relevant tests use real PostgreSQL.
- Direct SQL tests use `app`.
- Concurrency tests use separate connections.
- Fresh-server provisioning passes.
- CLI/REST workflows produce equivalent persisted state.
- No critical regression remains.

If any prerequisite fails, return FAIL or PASS WITH CONDITIONS and list the exact blocker.

Do not declare PASS because the isolated Go packages are green.

---

# 15. Deliverables

Create:

`plans/milestone2_remediation_report.md`

Include:

1. Initial Git baseline and uncommitted state.
2. Six prerequisite findings and disposition.
3. Compensation resolver design and evidence.
4. SHG lifecycle and evidence model.
5. Rule-pack activation governance.
6. Payroll-run state machine.
7. Snapshot schema and canonical hashing.
8. Snapshot concurrency and population consistency.
9. Database constraints, triggers and privileges.
10. CLI/API changes.
11. Migrations.
12. Independent regression tests.
13. Exact commands and results.
14. Remaining limitations and risks.
15. Explicitly deferred capabilities.
16. Final PASS/FAIL decision.

Create no additional architecture roadmap unless a genuine structural defect is discovered.

Keep all work uncommitted.

Do not modify historical migrations `00001`–`00011`.

---

# 16. Stop condition and next milestone

Stop when all six prerequisites pass.

Do not implement any statutory financial calculation.

The following milestone will be the first actual Singapore monthly payroll calculation using:

- Singapore Citizen employee.
- Monthly-rated employment.
- Full salary period.
- No overtime.
- No unpaid leave.
- No variable earnings.
- Independently verified CPF, SDL and applicable SHG requirements.
- Pinned statutory rule-pack versions.
- Immutable payroll input snapshot.
- Independent golden, anti-golden and boundary fixtures.
- Reproducible financial output.

The first calculation must not proceed until this milestone passes.

**The goal of Milestone 2 is to make every payroll run governable, every input identifiable, every rule version accountable and every snapshot reproducible. Build this foundation without inventing a single financial result. Preserve the existing application and stop at the acceptance gate.**

## My strongest implementation recommendation

There are two details I would watch particularly closely during execution.

1\. Do not confuse input approval with payroll approval. A validated snapshot is not an approved financial payroll. Until a real calculation exists, the application must not allow a run to claim it has been calculated, financially approved or finalized.

2\. Protect the expected employee population, not just the employees already discovered. Locking the current employee IDs is insufficient if a new employee or employment record can enter the pay period while the snapshot is being captured. The snapshot transaction needs an explicit population-stability guarantee.

These two details are particularly important because they determine whether the first payroll result will be a coherent financial decision rather than a collection of individually valid records.

My recommendation is to authorize Prompt 15 as written. Once it passes, we should stop building payroll infrastructure and finally implement the first verified Singapore payroll calculation.