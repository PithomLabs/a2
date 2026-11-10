I have reviewed the complete 747-line Prompt 12 report. I agree with its overall FAIL verdict. The important distinction is that the normal application workflows have improved substantially, but the PostgreSQL application role can still corrupt or silently change payroll inputs through direct SQL.

This is now a much narrower problem than the earlier Milestone 1 failures. I would not redesign the architecture.

## 1. What the latest review establishes



Milestone 1.2 — FAIL

No remaining P0 findings



The review examined eight evidence scopes and identified several new defects.

| Area                       | Finding      | Significance                                                                   |
| -------------------------- | ------------ | ------------------------------------------------------------------------------ |
| Correction groups          | A1-F1, A1-F2 | Direct SQL can create invalid or cross-employee correction lineage.            |
| Correction lifecycle       | A1-F3        | A normal correction can supersede only one segment of a multi-segment group.   |
| Date attribution           | A1-F4        | Direct SQL can misdate a corrected interval.                                   |
| Evidence timestamps        | A1-F5        | Backdated attendance evidence can bypass readiness conflict detection.         |
| Duplicate dispositions     | B-CYC-01     | Concurrent SQL operations may create a cycle and permanently block processing. |
| Compensation               | D-RACE       | Concurrent compensation changes can evade approval conflict detection.         |
| Compensation relationships | D-CONTAIN    | Direct SQL can create overlapping or out-of-employment compensation.           |
| Monetary precision         | D-ROUND      | PostgreSQL silently rounds three-decimal monetary input into `NUMERIC(12,2)`.  |

The last three compensation findings are particularly important. They demonstrate that the database can accept a pay-basis change that the application subsequently treats as valid.&#x20;

milestone1_2_final_adversarial_review\.md



## 2. The three most consequential discoveries

### A. Correction groups are not yet fully protected

The Go correction service creates coherent correction groups, but the database does not enforce all their relationships.

For example, an `app` SQL operation can create a group referring to another employee's attendance interval. That consumes the unique predecessor slot and can permanently prevent the legitimate employee correction from succeeding.

Another defect allows an operator to re-correct one segment of a two-segment group, leaving the other segment active. That creates split lineage without an obvious indication to the operator.

This is a particularly important lesson: one-to-many segmentation requires group-level lifecycle rules, not just individual interval rules.

The report identifies these separately as A1-F1, A1-F2 and A1-F3.&#x20;

milestone1_2_final_adversarial_review\.md



### B. The compensation approval race is real

This is arguably the most important finding.

The application uses PostgreSQL's `now()` to timestamp compensation creation and approval. PostgreSQL `now()` represents transaction-start time, not commit time.

Consider:

1. A compensation transaction starts.
2. A payroll approval transaction starts later.
3. Payroll is approved.
4. The original compensation transaction commits afterwards.

The compensation row can have an earlier `created_at` than the approval timestamp, even though it became visible after approval.

The readiness check therefore misses the retroactive change.

The review reproduced this behavior against PostgreSQL. The two timestamps and the missed conflict are recorded in the report.&#x20;

milestone1_2_final_adversarial_review\.md



My preferred fix is a shared transaction-locking barrier, not reliance on wall-clock timestamps.

Both compensation creation and payroll approval should serialize on the same employee-level or equivalent payroll-basis lock.

Then:

- If compensation obtains the lock first, approval waits and subsequently sees the change.
- If approval obtains the lock first, compensation waits and must detect that the period is already approved.

Use a consistent lock order and transaction isolation policy. `clock_timestamp()` alone does not resolve the race.

### C. Monetary precision must be enforced by PostgreSQL

The review directly inserted `199.999` into a `NUMERIC(12,2)` compensation column and PostgreSQL stored `200.00`.

The Go service rejects excessive precision, but a direct SQL insert bypasses that validation.

A CHECK constraint on the existing `NUMERIC(12,2)` column cannot detect the original third decimal place because PostgreSQL has already rounded the value.

The proposed solution is unconstrained `NUMERIC` with an exact-scale constraint, or an equivalent reusable domain. The same issue applies to monthly basic salary.&#x20;

milestone1_2_final_adversarial_review\.md



This should be fixed before implementing any financial calculations.

## 3. What is actually clean

There is meaningful progress worth preserving.

The review live-verified:

- Single-successor correction protection in the tested paths.
- Original timesheet and attendance evidence immutability in the tested paths.
- Core employment-timezone attribution, including adversarial session and host timezones.
- Canonical duplicate-event protection in ordinary service workflows.
- Compensation UPDATE/DELETE restrictions.
- Cross-employer checks in the examined correction paths.
- Published-timesheet correction-chain behavior.
- Payroll readiness detection of ordinary sequential compensation changes.
- No hidden overtime, CPF, SDL, SHG or net-pay calculations.
- `SG-MOM-OT-001` remains BLOCKED.

The full Go suite passed under three timezone configurations in the warm disposable environment, as did the five race-tested packages. The report is also honest about the tests it could not perform.&#x20;

milestone1_2_final_adversarial_review\.md



I particularly appreciate that the review separated actual failures from unexecuted hypotheses, instead of inflating all outstanding questions into defects.

## 4. One important testing problem remains

The report found that a genuinely clean PostgreSQL server can still cause the test suite to fail due to database-creation and role-provisioning races.

The warm-server runs pass, but that is not enough to establish reliable fresh-environment testing.

It also found that the privilege test uses a fixed table allowlist. A future table that accidentally inherits full application-role permissions could go unprotected without making the test fail.

These are two different improvements:

- Make test initialization reliable on a clean server.
- Make privilege verification fail closed when an unexpected table appears.

Both deserve inclusion in the next pass.&#x20;

milestone1_2_final_adversarial_review\.md



## 5. My recommended next milestone

I would authorize one final bounded pass:

# Milestone 1.3 — Final Payroll Input Integrity Closure

1\. Database correction lineage

One forward migration to bind correction groups, predecessors, segments, employee ownership and business dates. Close the direct-SQL bypasses.

2\. Payroll basis consistency

Fix the compensation/approval race using shared transaction serialization. Enforce parent-term containment, employee-level compensation overlap protection and exact monetary precision.

3\. Correction lifecycle

Make multi-segment re-correction group-atomic. Add the missing compensation correction lifecycle with auditable succession.

4\. Safe rollback and duplicate handling

Guard destructive populated Down migrations. Test concurrent duplicate dispositions rather than relying on static lock reasoning.

5\. Verification reliability

Fix clean-server database initialization and make the privilege matrix enumerate all relevant tables, rejecting unknown permissions.

The report's proposed six-item remediation sequence is a good starting point. I would keep the scope close to it.&#x20;

milestone1_2_final_adversarial_review\.md



One nuance: populated migration rollback does not have to be fully supported if we adopt an explicit forward-only migration policy. But the Down migration must fail safely before destroying existing correction or duplicate-disposition history. It must not silently delete decisions.

## 6. What I would not do

Do not revisit:

- The modular-monolith architecture.
- Multi-employer-ready database structure.
- The Go/pgx technology choice.
- The shared CLI/API service design.
- The Singapore-first jurisdiction strategy.
- The Phys-Research/Phys-Pseudo methodology.
- The already verified timezone and ordinary correction paths.

Do not implement payroll calculations yet.

There is no reason to add CPF or overtime code while the database can still accept ambiguous compensation records or miss a concurrent pay-basis change.

## 7. The next acceptance criterion

I would simplify the goal to one fundamental guarantee:

No operation performed with the ordinary `app` database credentials, whether through Go services or direct SQL, may silently corrupt, misattribute, duplicate or retroactively alter the evidence used to produce an approved payroll result.

Once that statement is backed by PostgreSQL tests, concurrent transaction tests and a reliable clean-install test, we should stop the integrity reviews.

Then move to the first restricted monthly payroll calculation.

My recommendation is to use the compensation/approval race as the central adversarial acceptance scenario: it tests whether the system understands the difference between transaction time, commit visibility and historical business approval.

The architecture is holding. What remains is to make the database and payroll workflow enforce the same historical truth, even under concurrency and imperfect application behavior.