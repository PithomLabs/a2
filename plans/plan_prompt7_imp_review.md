Milestone 1 is a genuine success. I would accept it as PASS for its stated scope and move on to the first real payroll implementation. We have crossed an important threshold: WorkforceOps is no longer just an architecture and database design. It is a functioning business application.

There are, however, two operational gaps I would close before treating the CLI/API workflow as fully usable by an actual operator.

## 1. What has actually been delivered

Milestone 1 — PASS

The report demonstrates a working flow:

1. Employee created with employment terms.
2. Employee assigned to BPO team and skills.
3. Five-shift roster created and published.
4. Twenty clock events imported, including break evidence.
5. Attendance reviewed and corrected.
6. Timesheet generated, submitted and approved.
7. Payroll readiness assessed, with statutory blockers reported.

The demonstrated timesheet contains 39.5 approved regular hours. An employee with an unmatched punch was correctly excluded. The CLI and REST results matched in all six normalized comparisons.&#x20;

plan_prompt7_imp.md



The engineering evidence is encouraging:

- 13/13 Go packages passed, twice.
- Real PostgreSQL 16 integration tests.
- Single executable.
- OpenAPI validation passed.
- Break evidence introduced through migration `00008`.
- Previous architecture and database integrity guarantees retained.&#x20;

  plan_prompt7_imp.md



This is a meaningful vertical slice.

## 2. What is not yet complete

We should be precise about the product status.

| Capability                    | Status          |
| ----------------------------- | --------------- |
| HR administration             | Implemented     |
| BPO configuration             | Implemented     |
| Shift and roster management   | Implemented     |
| Clock events and attendance   | Implemented     |
| Timesheets                    | Implemented     |
| Payroll readiness             | Implemented     |
| Gross payroll calculation     | Not implemented |
| CPF, SDL and SHG calculations | Not implemented |
| Overtime calculation          | Blocked         |
| Financial payslips            | Not implemented |

The readiness result is currently false, with all nine rule packs blocked and an SHG instruction missing for the test employee. This is an honest result, not a failed calculation. But it also means we have not demonstrated a successful payroll-readiness case yet.&#x20;

plan_prompt7_imp.md



## 3. Two small issues worth fixing

First: attendance interval discoverability.

The report says interval IDs can only be discovered through SQL, while corrections require those IDs.&#x20;

plan_prompt7_imp.md



That is an operational API gap. A REST client should be able to retrieve an employee's attendance intervals by employee and date range, inspect their provenance and submit a correction without accessing PostgreSQL directly.

Second: Idempotency-Key enforcement.

Clock imports deduplicate source event IDs, but the general Idempotency-Key contract is not enforced. Roster publication and timesheet approval deserve reliable replay behavior.&#x20;

plan_prompt7_imp.md



I would combine these into one small maintenance patch, not another milestone of architectural work.

## 4. My most important recommendation: do not make overtime the first payroll calculation

This is where I would change our previous sequencing.

The overtime package has already consumed substantial research effort. It is important, but it does not have to be the first payroll component we implement.

Our own blocking matrix distinguishes overtime lines from other payroll components. `SG-MOM-OT-001` blocks overtime calculations; it does not logically require us to block every ordinary salary calculation.

I recommend targeting this restricted payroll scenario first:

First supported payroll case

- One Singapore Citizen employee.
- Monthly-rated, full-time employment.
- Full salary period.
- Fixed monthly basic salary.
- No overtime.
- No unpaid leave.
- No public-holiday or rest-day work.
- No variable or additional wages.
- Explicit SHG contribution instruction or evidenced `verified_none`.
- Applicable CPF and SDL rules independently verified.

That gives us a much more direct route to a genuine payroll run.

It also provides a sensible sequence of research packages:

1. CPF monthly contributions for the restricted 2026 Singapore Citizen population.
2. SDL calculation.
3. SHG instructions and contribution calculation.
4. Basic monthly salary and payroll aggregation.
5. Itemised payslip and statutory record requirements.
6. Overtime as a separately verified additional earning component.

We must still qualify each statutory package independently. We should not assume that simpler arithmetic makes a rule automatically verified.

The key insight is that we can eventually produce a real monthly payroll for employees who have no overtime, while the overtime package remains blocked. This is a much better way to avoid letting one unresolved rule hold the entire product hostage.

## 5. What I would do next



Next milestone recommendation

&#x20;Milestone 2 — First real regular payroll calculation.



The work should have two tightly bounded parts:

A. Small operational patch: expose attendance interval retrieval through REST and CLI, enforce idempotency for the critical workflow transitions, and test replay behavior.

B. Payroll implementation: research, independently verify and implement the restricted regular-monthly-salary case, beginning with CPF. Reuse the existing payroll architecture, approval gate, source lineage and immutable-run design.

Do not reopen the domain model, database architecture, employer isolation or WFM implementation. Do not promote `SG-MOM-OT-001` just to make payroll appear complete.

## 6. One thing I want before approving any commit

The report says there are 79 changed entries and nothing has been committed. I would preserve that working tree for now.&#x20;

plan_prompt7_imp.md



For the next adversarial review, I would like the actual changes in these areas:

- `internal/attendance/`
- `internal/timesheet/`
- `internal/payroll/`
- `cmd/` and API handlers
- Migration `00008_break_events`
- The updated `openapi.yaml`
- `git diff --check` and `git diff --stat`

The implementation report establishes a credible test result, but it is not a substitute for examining the implementation itself.

My final assessment: accept Milestone 1. Do a small workflow-accessibility correction, then move directly into the first real payroll case, prioritizing ordinary monthly salary and CPF rather than waiting indefinitely for overtime research.

We now have the operational foundation. The next major objective is to make WorkforceOps calculate its first independently verified dollar amount.