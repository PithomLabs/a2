I reviewed `plan4.3.md` against the previous review findings.

My verdict: the plan is sound and should proceed, with four amendments before execution. It has the right scope: correct the five P0 issues, land the six smaller improvements, prove database invariants against real PostgreSQL, and avoid reopening the frozen architecture decisions.&#x20;

plan4.3.md



## 1. Decision on Q1: use the real baseline migration

Approved

Choose D1: create the actual baseline migration and test against it, not fixture-only DDL.

I agree with the reasoning in plan4.3. A separate test schema would demonstrate that a copy of the design works, not that the application's real database schema works.&#x20;

plan4.3.md



I approve the proposed boundaries:

- Write the actual initial PostgreSQL migration.
- Include tables, constraints, indexes, privileges and triggers.
- Do not implement business services.
- Do not implement payroll calculations.
- Do not add seed data.
- Run the integrity tests against PostgreSQL 16 using the actual `app` role.
- Permit pre-release migration rewrites.

One additional requirement: the baseline migration must cover the complete revised database design, not just the tables needed by the first three test groups. Otherwise, the migration and the supposedly authoritative database design will diverge before the first implementation slice.

## 2. Four amendments to plan4.3

### Amendment A — Include `payroll_run_exclusion` in the multi-employer model

D4 introduces this entity, but the success criteria do not explicitly include its employer ownership and composite relationships.

It must have:

- `employer_id`
- `payroll_run_id`
- `employee_id`
- `reason`
- `actor`
- `excluded_at`

Its foreign keys must ensure that the run and employee belong to the same employer.

It also needs a database-enforced rule preventing exclusions from being added, changed or removed once the payroll run has been validated.

This is important because exclusions are part of payroll integrity, not merely an administrative note.&#x20;

plan4.3.md



### Amendment B — Make payroll exclusions unambiguous

The current wording says:

> unverified pack → run cannot approve; explicit exclusion → proceeds

That is too easy to misinterpret.

Specify:

- A blocked employee cannot be included in an approved payroll calculation.
- An operator can explicitly exclude that employee before validation, with a recorded reason.
- The employee is not considered paid by that regular payroll run.
- The exclusion remains visible in the payroll register and audit history.
- The operator must handle the outstanding employee payment separately.
- An approved run must disclose its exclusions and must never be described as a complete payroll for all employees without qualification.

The rule is all-or-nothing for the employees included in the run, not permission to silently omit difficult cases.

### Amendment C — Specify the reproducibility contract now

D3 correctly acknowledges that full calculation reproducibility cannot be executed before payroll calculations exist.&#x20;

plan4.3.md



However, simply recording `inputs_hash` and `rule_versions` is not necessarily enough to reproduce a historical payroll result.

Before the design gate, specify that a payroll result must preserve or immutably reference:

1. The exact approved timesheet and employment-term inputs.
2. The exact statutory rule-pack versions.
3. The calculation implementation/version.
4. The rounding and calculation conventions.
5. The resulting payroll lines and payslip.

The eventual payroll slice must test:

`Same immutable inputs + same rule versions + same calculation version = same payroll output.`

Do not claim that this end-to-end test passes in plan4.3. It is a deferred implementation-slice test with an explicit contract, as D3 proposes.

### Amendment D — Make the employer-scoping success criterion exhaustive

Instead of saying "all six named children", enumerate the full employer-owned table inventory from the revised database design, including the new `payroll_run_exclusion`.

For the attendance consistency constraint, explicitly require that an attendance interval's employee and roster assignment employee must match. This must be enforced with a composite foreign key, not an application check.

The test should prove that even two employees belonging to the same employer cannot accidentally have one another's roster assignments linked to their attendance intervals.

## 3. What I approve without further discussion

The rest of the plan is appropriately scoped.

| Decision                                         | Verdict                                |
| ------------------------------------------------ | -------------------------------------- |
| Real PostgreSQL baseline migration               | Approved                               |
| Two database roles                               | Approved                               |
| Trigger-based parent-state immutability          | Approved                               |
| Employer-scoped clock-event deduplication        | Approved                               |
| Pure payroll approval-blocking predicate         | Approved                               |
| Evidenced `verified_none` SHG status             | Approved                               |
| Read-only rule-pack activation endpoint          | Approved                               |
| Structured roster legal-validation evidence      | Approved                               |
| Relational lineage plus defined `source_refs`    | Approved                               |
| CLI and API operator attribution                 | Approved, explicitly non-authenticated |
| No seed data or services in this pass            | Approved                               |
| Full end-to-end payroll reproducibility deferred | Approved, subject to Amendment C       |

The proposed validation commands and PostgreSQL integration tests are also appropriate.&#x20;

plan4.3.md



## 4. Final instruction for the coding agent

I would amend `plan4.3.md` with those four corrections and then authorize execution.

The most important acceptance condition is the one already identified in the plan:

Under the actual PostgreSQL `app` role, draft edits must succeed, while edits to published or approved records must fail.

That test must cover both UPDATE and DELETE wherever immutability is promised. It should also prove that the SQL errors are translated into the documented machine-readable API errors when those service endpoints are implemented.

The design gate should distinguish three outcomes:

- Passed now: PostgreSQL schema integrity, employer isolation, draft/published workflow integrity, payroll approval-blocking logic and lineage constraints.
- Specified but deferred: actual payroll calculation reproducibility.
- Still blocked: any statutory calculation whose rule pack has not reached `IMPLEMENTABLE`.

Final decision: approve plan4.3 with Amendments A–D. Use the real baseline migration. Do not expand the scope beyond those corrections.

This should be the final design-correction pass before the design gate, not another opportunity to redesign WorkforceOps.