This is a substantially stronger review. My verdict is to accept its FAIL assessment. Milestone 1.1 is not yet ready for the first payroll calculation, but the remaining problems are now sufficiently localized to make a bounded remediation possible.

The most important development is that the review went beyond checking whether the new tests pass. It challenged the actual PostgreSQL security mechanisms, exercised application-role SQL, tested correction chains and reproduced failures in real business workflows.

## 1. What has genuinely improved

Two original P0 defects fixed

The report confirms that:

- Approved timesheet originals remain immutable.
- The original unbounded clock-pairing problem has been addressed.
- The 72-hour limit is compliance-only, not a payable-hours cap.
- Correction chains have single-successor protection.
- Published roster revisions have a working revision workflow.
- Employer isolation, transaction locking and core readiness predicates have been independently exercised.

The correction-chain and database protections are no longer just architectural intentions. Several have passed live PostgreSQL adversarial tests.&#x20;

milestone1_1_final_adversarial_review\.md



This matters because we are no longer dealing with the original broad integrity problems. We are dealing with specific defects in the boundaries between otherwise functioning components.

## 2. The eight defects that must be addressed first

These are the most consequential findings because they are reachable through normal business operations.

| Finding             | What can go wrong                                                                                                 |
| ------------------- | ----------------------------------------------------------------------------------------------------------------- |
| F-OVERLAP-BLOCK     | Duplicate clock feeds can permanently prevent timesheet generation.                                               |
| F-MIDNIGHT-BREAK    | A break spanning midnight can remain unpaired indefinitely.                                                       |
| F-APPROVED-DEADLOCK | Reissuing an approved timesheet still does not permit the associated attendance correction.                       |
| F-COMP-RETRO        | Compensation can change after approval without readiness detecting it.                                            |
| F-TZ-CLASSIFY       | Running the application under UTC can misclassify Singapore break events and overstate paid time.                 |
| F-CORRECT-SPLIT     | Cross-midnight corrections can allocate hours to the wrong business date and pay period.                          |
| F-LATEBREAK-DESYNC  | Late break evidence can change attendance after approval without invalidating or flagging the approved timesheet. |
| F-EMP-STATUS        | Terminated or resigned employees can still be rostered and have their hours flow into timesheets.                 |

These are eight different symptoms, but they reveal three underlying problems.

Evidence lifecycle: F-OVERLAP-BLOCK, F-APPROVED-DEADLOCK and F-LATEBREAK-DESYNC.

Temporal interpretation: F-MIDNIGHT-BREAK, F-TZ-CLASSIFY and F-CORRECT-SPLIT.

Employment history affecting operations: F-COMP-RETRO and F-EMP-STATUS.

I would preserve these groupings in the next remediation plan. They are more useful than treating eight findings as eight completely unrelated patches.&#x20;

milestone1_1_final_adversarial_review\.md



## 3. The most important new discovery: database integrity is not fully protected

The new review exposed a second class of problems.

The `app` database role can perform direct SQL operations that bypass some of the protections we thought were established.

Database hardening still required

The most serious examples are:

- `F-IS-CURRENT-BYPASS`: explicitly inserting `is_current=false` can make verified work disappear.
- `F-PG-TEMP`: a temporary table can redirect a `SECURITY DEFINER` trigger away from the actual attendance table.
- `F-XEMP-SUPERSEDE`: one employee's correction can supersede another employee's attendance interval.
- `F-LINE-MOVE`: a timesheet line can be moved between employees.
- `F-SHEET-KEYS`: timesheet employee or period identity can be changed after evidence is linked.
- `F-ROSTER-REOPEN`: a published roster can be reopened through direct SQL.
- `F-HR-DIRECT-WRITE`: closed employment terms and employee status can be modified outside their governed services.
- `F-SINGLECHAIN-UPDATE` and `F-CHAIN-SCOPE`: correction links can be changed or connected across invalid employees and periods.

These are not all reachable through current CLI or REST operations. However, the project explicitly adopted the principle that the database application role must not be able to circumvent critical historical-integrity guarantees.

I agree with retaining that standard.

The `SECURITY DEFINER` findings deserve particular attention. A function with `search_path=public` is not necessarily safe from `pg_temp` shadowing. The function needs explicit schema-qualified references, and the application's privileges must prevent manipulation of its maintenance context.&#x20;

milestone1_1_final_adversarial_review\.md



## 4. Two fresh-install problems deserve equal urgency

There are two operational findings I would not allow to get lost among the database details.

F-INITROLES: the actual `init-roles.sh` fails because psql variable substitution does not occur inside the dollar-quoted `DO` block. That means the documented provisioning path cannot create its required roles.

F-SUPERUSER-DEFAULT: the normal local configuration still defaults the runtime connection to a superuser. That effectively bypasses the privilege protections we worked to establish.

Fixing triggers is not enough if the documented developer workflow does not use the role against which those triggers and privileges were designed.

Also, P1-23 remains: a fresh database cannot create an employer through the supported CLI/API workflow without manual SQL or seed data.

These three problems should be solved together as part of the installation and onboarding path.&#x20;

milestone1_1_final_adversarial_review\.md



## 5. What the final review tells us about the quality of the architecture

There is good news.

The review does not recommend changing the architecture.

It found that:

- One application and shared services work as intended.
- Employer isolation is holding in the examined paths.
- The core correction-chain design is viable.
- The main database locking mechanisms work in the tested cases.
- The 72-hour compliance separation is correct.
- No hidden statutory calculation paths were identified.
- The Singapore overtime rule package remains BLOCKED.
- The CLI/REST persisted-state comparison passed.

The current problems are therefore better characterized as incomplete enforcement and business-lifecycle defects inside a sound architectural direction, not evidence that we chose the wrong architecture.&#x20;

milestone1_1_final_adversarial_review\.md



There are still meaningful unverified areas, including TZ=UTC execution, crash/replay behavior under process termination, some concurrency scenarios, rollback with published rosters, and the formal OpenAPI validator output.

We should close the most consequential ones through targeted tests rather than create another expansive review campaign.

## 6. My recommended next move

I would authorize one final, tightly bounded remediation milestone.

# Milestone 1.2 — Evidence and Database Integrity Closure

Workstream A — Attendance and payroll-input correctness

Fix the eight service-reachable P1 defects, with regression tests for each.

Workstream B — Database protection

Close the `is_current`, `SECURITY DEFINER`, cross-employee lineage, timesheet identity, roster immutability and HR direct-write bypasses.

Workstream C — Fresh installation

Repair role provisioning, remove superuser runtime defaults and support SQL-free employer onboarding.

Workstream D — Verification

Execute the critical timezone, break, correction, approval, concurrency and application-role tests against fresh PostgreSQL 16 databases.

I would not take on all 33 P2 findings indiscriminately. The next milestone should include only P2 findings that directly affect these corrections or the first payroll scenario.

For example, the SHG instruction lifecycle and `verified_none` problem will need attention when SHG is included in the first payroll profile. Rotation materialization can remain deferred.

## 7. One important clarification about the next payroll milestone

I would retain our earlier strategy:

Do not make the unresolved Singapore overtime calculation the prerequisite for the entire payroll engine.

Once the operational evidence chain is trustworthy, we should implement a restricted monthly payroll calculation for an employee with:

- Singapore Citizen status.
- Monthly-rated employment.
- A full salary period.
- No overtime.
- No unpaid leave.
- No variable earnings.
- Fully verified CPF, SDL and applicable SHG inputs.

The statutory rule packages still need their own Phys-Research and Phys-Pseudo approval gates.

This would allow the first real payroll amount to be calculated without waiting for the separate overtime research problem to be resolved.

## 8. Final decision

| Area                                              | My disposition                  |
| ------------------------------------------------- | ------------------------------- |
| Original P0 defects                               | Fixed                           |
| Service-reachable attendance/timesheet P1 defects | Must fix                        |
| Database-role hardening                           | Must fix                        |
| Fresh-install and employer onboarding             | Must fix                        |
| Architecture                                      | Retain                          |
| Remaining non-critical P2/P3                      | Prioritize by payroll relevance |
| Overtime calculation                              | Remain blocked                  |
| First payroll calculation                         | Defer until integrity closure   |

One reporting detail also deserves correction: the claimed 98-test total is not independently reconciled, and the successful Docker build was not directly executed because the daemon was unavailable. These should remain explicit evidence limitations, not be silently described as verified passes.&#x20;

milestone1_1_final_adversarial_review\.md



My conclusion: We have reached a much better position than the earlier FAIL reviews. The remaining defects are identifiable, reproducible and grouped around a manageable set of invariants.

I would not redesign WorkforceOps, add more features or revisit the jurisdiction architecture.

The next objective is simple: make the evidence chain trustworthy under normal operations and under the actual PostgreSQL `app` role. Then proceed to the first independently verified payroll calculation.

That should be the final integrity gate—not another open-ended review cycle.