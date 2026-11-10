This is the strongest milestone result so far. I accept Milestone 1.3 as PASS based on the submitted execution report. We have reached the point where the operational integrity work can close and development can move to actual payroll calculations.

The important achievement is not simply that 429 tests passed. It is that the project has addressed the specific database, concurrency and historical-data problems that repeatedly prevented us from trusting payroll inputs.

## 1. What has been achieved

Milestone 1.3 — PASS (reported)

The most significant results are:

| Area                       | Achievement                                                        |
| -------------------------- | ------------------------------------------------------------------ |
| Compensation/approval race | Shared advisory lock and per-employee `pay_basis_seq`              |
| Attendance corrections     | Multi-predecessor correction groups and group-atomic re-correction |
| Compensation               | Controlled creation and correction functions                       |
| Monetary precision         | Exact-scale PostgreSQL constraints and exact Go representation     |
| Duplicate events           | Pair locking and bidirectional canonical-event checks              |
| Historical evidence        | Protected creation timestamps and correction lineage               |
| PostgreSQL privileges      | Fail-closed matrix covering 55 tables and 30 functions             |
| Migration policy           | Forward-only when dependent business history exists                |
| Test reliability           | Clean-server provisioning verified twice                           |
| Statutory calculations     | None introduced; OT-001 remains BLOCKED                            |



milestone1_3_remediation_report.md

milestone1_3_remediation_report.md



Three decisions deserve particular attention.

### A. Compensation approval is now transaction-aware

The earlier compensation race exposed a fundamental problem with using timestamps as evidence of transaction ordering.

Milestone 1.3 introduces a shared employee-level advisory lock and a sequence assigned after acquiring that lock. Both compensation changes and approval transitions participate in the same protocol.

The report records six live two-connection race tests covering both possible lock-winner orders.

This is a major improvement: the system now has an explicit mechanism for coordinating changes to the financial basis with approval.&#x20;

milestone1_3_remediation_report.md



### B. Correction groups now represent the actual business operation

The move to a predecessor-link table allows one correction operation to replace multiple previous intervals with multiple date-segmented successors.

Crucially, the database enforces group completeness, ownership, binding and sibling consistency, while Go provides the operator-facing workflow.

This is considerably more expressive than a simple `supersedes_id` relationship.&#x20;

milestone1_3_remediation_report.md



### C. Compensation now has a proper historical lifecycle

Compensation creation and correction go through controlled database functions. Original rows remain intact, successors are linked, audit entries are atomic, and approved-period corrections remain visible as conflicts until the correction is approved through the downstream workflow.

The database also rejects excess monetary precision instead of silently converting `199.999` into `200.00`.&#x20;

milestone1_3_remediation_report.md



These capabilities are directly relevant to a trustworthy payroll engine.

## 2. The verification evidence is materially stronger

# 429

Top-level test cases passed

# 0

Test failures

# 55

Database tables classified

# 30

Database functions inspected

Additional verification

- Full suite under default, UTC and Singapore host timezones.
- Race detector across five service packages.
- 61 integrity tests passed, with one intentional deferred test.
- Clean PostgreSQL provisioning passed twice.
- CLI/REST parity verified.
- Lint, build and diff checks passed.

milestone1_3_remediation_report.md

I particularly appreciate that the agent caught another fresh-server provisioning race during execution, reproduced it as a failure and then verified the corrected implementation. That is much better evidence than merely testing repeatedly against a warmed-up database.

The original migrations `00001–00010` were also verified byte-identical, with only one new forward migration, `00011`. No committed history was changed.&#x20;

milestone1_3_remediation_report.md



## 3. Three remaining considerations

These do not overturn the milestone PASS, but they should be carried into the next development stage.

First: disposition and approval synchronization.

The report acknowledges that direct SQL disposition operations do not acquire the employee payroll-input barrier. This was outside the specified disposition concurrency work.

Before the first payroll run, prove that a disposition cannot change the interpretation of an attendance population while approval is occurring. If the event can change eligibility or clear a blocking exception, it needs the same barrier.&#x20;

milestone1_3_remediation_report.md



Second: SHG and payroll-run lifecycle.

The SHG lifecycle and complete payroll-run validation/approval state machine remain deferred. They are not reasons to reopen Milestone 1.3, but they are dependencies of the first usable payroll scenario. Missing SHG information must never be interpreted automatically as verified non-applicability.&#x20;

milestone1_3_remediation_report.md



Third: deployment verification.

Docker Compose and the actual server loopback smoke test were not executed because the Docker daemon was unavailable. Local PostgreSQL testing was successful. These environment limitations should be closed before a release or deployment claim, rather than holding up the next local development milestone.&#x20;

milestone1_3_remediation_report.md



# 4. My recommendation: move to Milestone 2

Next development objective

Milestone 2 — First independently verified Singapore payroll calculation.

I recommend that we stop expanding the operational platform and concentrate on the first real payroll result.

The initial supported scenario should be deliberately narrow:

- One Singapore legal employer.
- One Singapore Citizen employee.
- Monthly-rated, full-time employment.
- A complete salary period.
- Fixed monthly basic salary.
- No overtime.
- No unpaid leave.
- No variable earnings.
- Complete and verified statutory inputs.

I would structure the next milestone into five deliverables.

### 1. Freeze the payroll input snapshot

Before calculating anything, establish a reproducible input snapshot containing:

- Employee and employment-term versions.
- Effective compensation records.
- Approved effective timesheet version.
- Payroll period.
- Jurisdiction.
- Applicable rule-pack versions.
- Relevant SHG instruction and evidence.
- Explicit exclusions and unresolved conflicts.

The snapshot should be immutable once the payroll calculation begins.

### 2. Complete the required statutory research

Use the established Phys-Research and Phys-Pseudo process independently for:

- CPF.
- SDL.
- Applicable SHG treatment.

Each rule needs verified source provenance, applicability, derivation and independently established test outcomes.

Do not revive the blocked overtime package merely to complete this scenario.

### 3. Implement the payroll-run lifecycle

The existing payroll readiness endpoint is not a complete payroll workflow.

Implement the smallest necessary lifecycle:

`Create → Validate → Calculate → Review → Approve → Finalize`

Approval must freeze the relevant input snapshot and rule versions. The run must fail closed when required statutory inputs are unresolved.

### 4. Implement the first financial calculations

Use exact monetary arithmetic and the verified rule packages.

Produce separately identifiable:

- Gross basic earnings.
- Employee statutory deductions.
- Employer statutory contributions.
- Employer levies.
- Net pay, only when all required components are verified.

Generate an itemized financial record from the calculation result. Payslip output should represent the approved calculation, not recompute it independently.

### 5. Independently verify the result

Create golden, anti-golden and boundary fixtures from the research and derivation packages.

Test reproducibility:

`Same immutable inputs + same rule versions + same calculation version = same result`

Also test an amended payroll input, an incorrect compensation record and a missing SHG instruction.

Do not claim payroll compliance for unsupported employee populations or rule scenarios.

## 5. What I would preserve as a permanent engineering lesson

We have spent several iterations discovering that a payroll formula is only the last stage of a much longer correctness chain.

The system must first establish:

1. Who the employee is and which employment terms apply.
2. Which shifts were scheduled.
3. What work actually occurred.
4. Which clock events support that interpretation.
5. Which corrections superseded earlier evidence.
6. Which time was approved.
7. Which compensation basis applied.
8. Which statutory rules governed the period.

Only then should it calculate money.

The architecture is now mature enough to attempt that calculation.

My recommendation is to declare the operational integrity phase closed, retain all current work for review before committing, and make Milestone 2 the first genuinely financial milestone.

The objective is no longer to prove that WorkforceOps can process attendance and timesheets. It is to prove that it can calculate one real Singapore payroll result, explain every component, reproduce it independently and defend the evidence behind it.