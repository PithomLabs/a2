I recommend one final pre-Milestone 2 adversarial review, focused on whether the integrity guarantees delivered in Milestone 1.3 are sufficient to safely introduce the first financial payroll calculation.

This should not repeat the previous 72-finding review. It should investigate three remaining areas: disposition-versus-approval synchronization, the statutory input lifecycle (particularly SHG), and payroll-run snapshots and approval.

# Prompt 14 — Pre-Milestone 2 Adversarial Code Review

# WorkforceOps: Payroll Calculation Readiness Gate

## 1. Objective

Conduct a narrowly scoped, independent adversarial review of WorkforceOps after the reported successful completion of Milestone 1.3.

The objective is to determine whether the existing operational system and database architecture are sufficiently trustworthy to begin implementing the first independently verified Singapore payroll calculation.

Milestone 1.3 reports:

- Compensation and approval synchronization through a shared employee-level advisory lock and `pay_basis_seq`.
- Group-atomic attendance correction.
- Database-enforced correction lineage.
- Concurrent duplicate-event disposition protection.
- Compensation creation and correction through controlled database functions.
- Exact monetary precision enforcement.
- Fail-closed database privilege enumeration.
- Clean-server test provisioning.
- 429 tests passing under default, UTC and Singapore timezone configurations.
- `SG-MOM-OT-001` remains BLOCKED.
- No statutory calculations were introduced.

The implementation report is:

`plans/milestone1_3_remediation_report.md`

The next intended milestone is the first restricted Singapore monthly payroll calculation.

This review must determine whether any remaining integrity or architectural problem could undermine that calculation.

**Review only. Do not implement payroll calculations or modify application source.**

## 2. Review boundaries and repository preservation

Before starting:

1. Record `git status --short`.
2. Record `git rev-parse HEAD`.
3. Confirm HEAD remains `f43515883668c917edac31be49a7cf8334eca5c2`.
4. Inventory all modified, deleted, renamed and untracked files.
5. Read the complete Milestone 1.3 remediation report.
6. Inspect migration `00011_m13_integrity.sql`.
7. Inspect the current payroll, attendance, timesheet, HR and SHG implementation.

Preserve all existing uncommitted work.

Do not:

- Reset or clean the repository.
- Stage or commit.
- Discard existing work.
- Modify migrations.
- Modify `bchat/`.
- Modify upstream Phys-Research or Phys-Pseudo.
- Deploy to cloud infrastructure.
- Implement statutory calculations.
- Reopen the architecture.
- Expand the review into a general-purpose code audit.

Use disposable PostgreSQL 16 databases for additional tests.

The only permitted new repository file is:

`plans/pre_milestone2_adversarial_review.md`

Temporary SQL probes and logs may be placed outside the repository.

## 3. Review Area A — Disposition and payroll-input approval synchronization

\<badge>Critical focus\</badge>

The Milestone 1.3 report explicitly acknowledges that direct-SQL clock-event dispositions use the disposition pair-lock but do not acquire the employee payroll-input barrier.

The report describes this as an intentional scope limitation.

Determine whether this creates a real business-integrity gap.

Inspect:

- `internal/attendance/disposition.go`
- `internal/attendance/attendance.go`
- `internal/timesheet/timesheet.go`
- `internal/payroll/readiness.go`
- `migrations/core/00011_m13_integrity.sql`
- Clock-event disposition triggers.
- `acquire_payroll_input_lock`.
- `pay_basis_seq`.
- Relevant application-role tests.

### A1. Establish what a disposition changes

Determine exactly how a disposition affects:

- Raw clock-event availability.
- Canonical-event pairing.
- Duplicate-event exclusion.
- Unpaired-event exceptions.
- Attendance interval creation.
- Attendance review.
- Timesheet generation.
- Timesheet approval.
- Payroll readiness.

Distinguish:

- A duplicate disposition before attendance generation.
- A disposition after attendance intervals exist.
- A disposition after timesheet generation.
- A disposition after submission.
- A disposition after approval.

Determine whether a disposition can change the meaning, eligibility or completeness of evidence already incorporated into an approved timesheet.

Do not assume that every disposition requires the payroll lock. Establish the actual effect first.

### A2. Test direct-SQL disposition versus approval

Use two real PostgreSQL connections authenticated as `app`.

Construct a controlled scenario with a canonical event, duplicate event, pending attendance and timesheet.

Test both interleavings.

**Scenario 1: disposition wins**

1. Transaction A begins a direct-SQL duplicate disposition.
2. It acquires the disposition pair-lock.
3. It pauses before commit.
4. Transaction B attempts timesheet submission/approval.
5. Complete transaction A.
6. Observe what transaction B sees.
7. Inspect the approved timesheet and readiness result.

**Scenario 2: approval wins**

1. Transaction B starts approval and acquires the payroll-input barrier.
2. It pauses before commit.
3. Transaction A attempts a direct-SQL disposition.
4. Complete transaction B.
5. Observe whether transaction A is blocked, rejected, or commits.
6. Inspect the original and approved evidence.

A disposition must not silently change an already approved payroll-input interpretation.

If the existing consumed-event guards make this impossible, prove that with direct SQL and persisted-state assertions.

If it is possible, report the exact vulnerability.

### A3. Test service and trigger consistency

Verify whether CLI, REST and direct SQL enforce the same applicable synchronization rules.

If a shared lock is needed, establish a single deterministic lock ordering across:

- Employee payroll-input lock.
- Disposition pair-lock.

Avoid lock-order inversion and deadlocks.

The employee must be identified from the actual canonical/duplicate event relationships, not caller-supplied arbitrary identity.

The database must enforce the required lock behavior for direct application-role operations.

### A4. Acceptance

Return one of:

- No additional barrier required: direct disposition cannot alter approved evidence, proven by tests.
- Barrier required: direct disposition can affect payroll inputs and must participate in the employee payroll-input synchronization protocol.
- Correction workflow required: post-approval disposition needs explicit reissue or evidence-correction handling.

Do not change the implementation during this review.

---

## 4. Review Area B — SHG instruction and statutory-input lifecycle

The first intended payroll scenario requires complete and verified employee inputs.

The M1.3 report explicitly defers:

- Full SHG lifecycle.
- Verified-none evidence.
- Complete payroll-run state machine.

Inspect:

- `internal/hr/shg.go`
- SHG models and database schema.
- Relevant migrations.
- `internal/payroll/readiness.go`
- Payroll exclusion logic.
- SHG audit tests.
- Rule-pack activation and applicability checks.
- Current employee and employment-term models.

The question is not whether SHG calculations should be implemented now.

The question is whether the existing data model can distinguish the different states needed for a defensible payroll calculation.

### B1. Distinguish absence from verified non-applicability

Determine whether the current system can distinguish:

1. No SHG instruction supplied.
2. Instruction awaiting verification.
3. Applicable SHG contribution instruction.
4. Verified non-applicability.
5. Corrected or superseded instruction.
6. Historical instruction applicable to an earlier period.

An absent row must not automatically mean verified non-applicability.

An explicitly verified-none state must carry its supporting evidence, actor, effective date and provenance.

Do not infer SHG eligibility or non-eligibility from missing information.

### B2. Effective dating and historical correctness

Verify:

- Instruction belongs to the correct employer and employee.
- Effective dates are coherent with employment.
- Historical payroll periods resolve the correct instruction.
- Backdated changes cannot silently alter approved periods.
- Changes after approval create an explicit conflict or correction requirement.
- Superseded instructions remain historically visible.
- Concurrent instruction changes and approval cannot bypass the established payroll-input barrier.

### B3. Privacy and data minimization

Inspect whether the current model stores only information necessary to determine the applicable payroll instruction.

Do not introduce unnecessary sensitive personal attributes merely to simplify future calculations.

The first payroll scenario must use an explicit supported employee-input profile and verified instructions rather than infer statutory contribution applicability from incomplete employee data.

### B4. Readiness behavior

Test:

- Missing SHG instruction.
- Pending instruction.
- Applicable verified instruction.
- Verified non-applicability.
- Conflicting historical instructions.
- Retroactive instruction change.
- Instruction supersession.
- Instruction change after timesheet approval.

Determine whether readiness produces a precise machine-readable blocker for each incomplete state.

### B5. Acceptance

State whether SHG currently has sufficient lifecycle and evidence support for the first restricted payroll scenario.

If not, provide the minimum necessary SHG prerequisite.

Do not implement SHG calculations or expand the SHG data model during this review.

---

## 5. Review Area C — Payroll-run lifecycle and immutable input snapshots

\<badge>Critical focus\</badge>

The Milestone 1.3 report confirms that the current `internal/payroll/runs.go` primarily implements run creation and exclusions.

A complete payroll-run validation/approval lifecycle was explicitly deferred.

Before implementing the first financial calculation, establish whether the current data model and services can safely support a reproducible payroll run.

Inspect:

- `internal/payroll/runs.go`
- `internal/payroll/readiness.go`
- `internal/payroll/approval.go`
- `internal/payroll` models and repositories.
- Timesheet effective-version resolution.
- Compensation successor logic.
- `pay_basis_seq`.
- Rule-pack activation and manifests.
- Payroll-run and exclusion tables.
- Relevant database triggers and permissions.

Do not implement the lifecycle.

### C1. Existing payroll-run state

Determine the actual states supported by the current application.

Identify whether the system currently distinguishes:

- Draft.
- Validated.
- Calculated.
- Under review.
- Approved.
- Finalized.
- Superseded or corrected.

Do not assume these states exist just because they appear in the future roadmap.

Inspect actual schema constraints and service transitions.

Identify what is currently enforced and what is still only a design intention.

### C2. Payroll input snapshot

Determine whether the existing model can create an immutable input snapshot containing, at minimum:

**Employer and jurisdiction**

- Employer identity.
- Jurisdiction.
- Currency.
- Payroll period.

**Employee population**

- Expected employees.
- Effective-dated employment terms.
- Applicable employee status.
- Explicit exclusions and reasons.

**Compensation**

- Effective compensation records.
- Salary basis.
- Compensation successor versions.
- Relevant pay-basis sequence.

**Time and attendance**

- Effective approved timesheet version.
- Attendance correction lineage.
- Relevant attendance-conflict state.

**Statutory inputs**

- Applicable rule-pack identities and versions.
- Rule-pack lifecycle status.
- SHG instruction or verified-none evidence where applicable.
- Required employee classification inputs.

**Calculation**

- Calculation version.
- Input snapshot identity.
- Monetary precision policy.
- Result provenance.

The snapshot must preserve the exact versions used by the calculation, not merely query whichever row happens to be current later.

Determine whether source record IDs and immutable version references are sufficient, or whether materialized snapshot records or content digests are necessary.

Do not implement a generic snapshot framework if a small immutable payroll-input record is sufficient.

### C3. Snapshot concurrency

Inspect how payroll population and employee basis sequence are captured.

A snapshot must not mix:

- A newly approved timesheet.
- An older compensation record.
- A later employee status.
- A different SHG instruction.

Test whether the existing employee-level barrier can provide a coherent per-employee input version.

Determine whether a payroll run spanning several employees needs to:

- Lock employees in deterministic order.
- Capture their basis sequences.
- Revalidate the full population.
- Commit the snapshot atomically.

A sequence number must not be interpreted as a universal commit timestamp.

The review should explain the actual consistency guarantee supported by the current design.

### C4. Post-snapshot changes

Determine what happens if, after a payroll snapshot is created:

- Compensation is corrected.
- Employment terms change.
- Attendance is corrected.
- A late break arrives.
- A disposition is added.
- The effective timesheet version changes.
- An SHG instruction changes.
- A rule-pack version is superseded.

The system must either preserve the existing snapshot as a historical calculation basis or explicitly invalidate/block the pending calculation.

An approved or finalized payroll result must never silently change because current source records have changed.

### C5. Payroll population and exclusions

Review whether the existing expected-population logic is suitable for the first real payroll run.

The expected population must not be derived solely from employees who already have approved timesheets.

Verify:

- Employees active during some or all of the period.
- Employees terminated after the period.
- Employees missing timesheets.
- Employees with incomplete statutory inputs.
- Explicit exclusions.
- Exclusion identity and effective period.
- Empty population handling.
- All-or-nothing approval rules.

An empty or incomplete population must not pass through a vacuously true approval predicate.

### C6. Reproducibility

Determine whether the current model can satisfy:

`Same immutable inputs + same pinned rule versions + same calculation version = same payroll output`

Identify any missing version references or mutable dependencies.

Do not build the calculation engine.

### C7. Acceptance

Provide a clear verdict:

- Existing payroll-run architecture is ready for the first calculation, with a bounded lifecycle addition.
- Existing architecture needs specific input-snapshot prerequisites before calculation.
- Existing architecture has a structural defect requiring design work before implementation.

Give the minimum necessary changes for Milestone 2.

---

## 6. Review Area D — Reconfirm critical financial boundaries

This is a limited review of financial correctness prerequisites, not a second full database audit.

Verify:

- No `float64` monetary calculations exist.
- Exact decimal input is preserved through Go and PostgreSQL.
- Excessive monetary precision is rejected, not rounded silently.
- Compensation successors resolve to one effective basis.
- Compensation approval conflicts remain correct under concurrency.
- The payroll readiness gate cannot be bypassed.
- No hidden overtime, CPF, SDL, SHG or net-pay calculation exists.
- `SG-MOM-OT-001` remains BLOCKED.
- No invented golden fixtures have been introduced.

Inspect the first restricted payroll profile:

- Singapore Citizen.
- Monthly-rated employment.
- Full salary period.
- No overtime.
- No unpaid leave.
- No variable earnings.

Determine whether the existing employee, employment and compensation fields can represent this profile without inference or synthetic defaults that conceal missing data.

Do not make legal conclusions about CPF, SDL or SHG during this code review.

The applicable statutory rules must be researched and independently verified in Milestone 2.

---

## 7. Review Area E — Verify the transaction barrier as a system contract

The M1.3 `pay_basis_seq` design is now a central integrity mechanism.

Inspect all current mutation paths capable of changing payroll inputs.

Build a source-to-lock matrix:

| Mutation                | Shared employee lock      | Basis sequence     | Audit  | Readiness consequence |
| ----------------------- | ------------------------- | ------------------ | ------ | --------------------- |
| Compensation add        | Verify                    | Verify             | Verify | Verify                |
| Compensation correction | Verify                    | Verify             | Verify | Verify                |
| Employment-term change  | Verify                    | Verify             | Verify | Verify                |
| Employee-status change  | Verify                    | Verify             | Verify | Verify                |
| Clock import            | Verify                    | Verify             | Verify | Verify                |
| Break attachment        | Verify                    | Verify             | Verify | Verify                |
| Attendance correction   | Verify                    | Verify             | Verify | Verify                |
| Duplicate disposition   | Verify                    | Verify             | Verify | Verify                |
| Timesheet approval      | Verify                    | Assign             | Verify | Freeze basis          |
| Payroll exclusion       | Verify if basis-affecting | Verify if required | Verify | Verify                |

Do not assume every mutation necessarily increments the sequence.

For each operation, explain whether it changes:

- Raw source evidence.
- Interpreted attendance.
- Approved time.
- Compensation basis.
- Employee population.
- Applicable statutory instructions.

Verify that all mutations that can change an approved payroll interpretation either participate in the barrier or are protected by an equivalent invariant that prevents the change.

If direct SQL has a different synchronization contract from service calls, identify it explicitly.

Do not widen the barrier unnecessarily if an immutable or consumed-evidence rule already proves the operation harmless.

---

## 8. Review Area F — Environment and developer-workflow closure

The M1.3 report honestly notes that Docker Compose was not executed and the serve-loopback smoke test was not performed.

Inspect:

- `Dockerfile`
- `docker-compose.yml` or compose configuration.
- `.env.example`
- `Taskfile.yml`
- `scripts/db/init-roles.sh`
- README setup instructions.
- Runtime host configuration.

Determine whether the first Milestone 2 payroll development workflow can be reproduced with the supported local PostgreSQL setup.

Verify:

- Runtime uses `app`.
- Migration operations use `migration_admin`.
- Superuser defaults are refused.
- Rule packages and manifests load correctly.
- The same environment variables are used in CLI and REST.
- No configuration silently bypasses payroll-readiness validation.

If Docker is unavailable, do not claim Compose verification.

The missing Docker/serve smoke is an environment limitation unless code inspection identifies a defect.

Do not make deployment a blocker for starting local Milestone 2 implementation.

---

## 9. Testing requirements

Use disposable PostgreSQL 16 databases.

Every direct SQL adversarial test must use the actual `app` role.

At minimum verify:

### Disposition/approval

- Direct-SQL disposition before approval.
- Direct-SQL disposition during approval.
- Direct-SQL disposition after approval.
- Service disposition and approval.
- Consumed canonical and duplicate-event behavior.
- Correct readiness and evidence lineage.

### SHG readiness

- Missing instruction.
- Pending instruction.
- Verified applicable instruction.
- Verified non-applicability, if supported.
- Historical instruction.
- Correction after approval.
- Exact blocker behavior.

### Payroll snapshot

- Expected population with complete evidence.
- Expected employee missing timesheet.
- Employee with terminated current status but historical employment.
- Correction chain resolution.
- Compensation version resolution.
- Missing statutory pack.
- Missing SHG instruction.
- Post-snapshot compensation change.
- Post-snapshot attendance correction.
- Rule-pack version change.
- Repeat snapshot with identical inputs.

### Financial representation

- Exact input and persisted monetary value.
- Excess precision rejected.
- No floating-point conversion.
- Approved compensation version preserved.

Run existing relevant tests and report exact commands.

Do not re-run the entire historical review suite unnecessarily.

Do not modify application code or tests.

---

## 10. Findings classification

Classify each result:

**P0 — Critical**

- Approved payroll basis can silently change.
- Financial result can be produced using inconsistent employee inputs.
- Database or service bypass can change approved financial evidence.
- Rule-pack gating can be bypassed.

**P1 — High**

- Payroll calculation cannot reliably establish its expected population.
- Missing or ambiguous SHG evidence can be treated as verified.
- Concurrent operations can create inconsistent snapshots.
- Critical mutations bypass the shared payroll-input consistency contract.

**P2 — Medium**

- Operational diagnostics or correction visibility gaps.
- Bounded run-lifecycle omissions with safe workarounds.
- Non-critical snapshot metadata gaps.

**P3 — Low**

- Documentation inconsistencies.
- Developer-experience improvements.
- Environment-dependent smoke tests not executable in the sandbox.

For every finding, include:

1. Finding ID.
2. Severity.
3. Source file and function or database object.
4. Concrete failure scenario.
5. Test or inspection evidence.
6. Whether it blocks Milestone 2.
7. Minimum correction.

Keep confirmed defects, risks and untested hypotheses separate.

---

## 11. Required report

Create:

`plans/pre_milestone2_adversarial_review.md`

Include:

1. Git baseline and review scope.
2. Disposition/approval synchronization findings.
3. SHG instruction and evidence-lifecycle findings.
4. Payroll-run lifecycle assessment.
5. Input-snapshot and reproducibility assessment.
6. Pay-basis sequence and lock-coverage matrix.
7. Financial precision and statutory-rule boundary findings.
8. Environment and developer-workflow findings.
9. Tests actually executed.
10. Confirmed defects.
11. Risks and untested hypotheses.
12. Clean areas.
13. Milestone 2 prerequisites.
14. Explicitly deferred concerns.
15. Overall readiness verdict.

Return one of:

### READY FOR MILESTONE 2

No material integrity defect remains in the first restricted payroll path.

### READY WITH PREREQUISITES

The operational data is trustworthy, but a small number of defined payroll-input, SHG or run-lifecycle prerequisites must be addressed before financial calculations.

### NOT READY

A remaining defect can silently change, misattribute or invalidate the inputs on which the first payroll calculation would depend.

Do not declare READY merely because all existing tests pass.

---

## 12. Scope restrictions

This review does not authorize:

- CPF implementation.
- SDL implementation.
- SHG calculations.
- Overtime calculations.
- Gross or net salary calculations.
- Payslip generation.
- A generic payroll rules engine.
- A Phys-Pseudo interpreter/compiler.
- Multi-country packages.
- SSO or RLS.
- UI.
- Microservices.
- New simulation infrastructure.
- General architecture redesign.

Keep Singapore as the initial jurisdiction.

Keep the modular monolith, employer-scoped database and shared CLI/API services.

Do not change the status of `SG-MOM-OT-001`.

---

## 13. Stop condition

The review must produce a concrete decision on whether the current backend can safely support its first restricted payroll calculation.

Do not create another open-ended remediation roadmap.

If issues are found, list only the prerequisites directly affecting the first payroll calculation.

If no material issue is found, recommend closing the integrity-review phase.

**The objective is to verify readiness, not delay Milestone 2 indefinitely. Establish whether a payroll input snapshot can be trusted, whether the statutory instructions are explicit, and whether approval cannot race with changes to its underlying evidence. Then stop.**

## My acceptance standard

I would focus most heavily on three questions:

1. Can a clock-event disposition alter an already approved attendance interpretation without triggering the payroll-input barrier?
2. Can the system distinguish a missing SHG instruction from verified non-applicability?
3. Can a payroll run freeze one coherent, reproducible set of employee, compensation, timesheet and rule-pack versions?

These are the remaining questions that matter directly to the first financial calculation.

After this review, I want a decisive transition: either a short list of Milestone 2 prerequisites or authorization to begin the first Singapore payroll calculation. No more general architecture reviews.