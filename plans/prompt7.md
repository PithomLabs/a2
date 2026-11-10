My recommendation is to stop expanding the research and design work and start building the first working backend vertical slice.

The biggest risk now is spending another week refining specifications without producing a usable application. We already have a tested PostgreSQL foundation, a multi-employer-ready schema, CLI/API contracts and a formal rule-specification process.

The next move should turn that foundation into an actual working HR/WFM workflow, while keeping statutory payroll calculations safely gated.

I would not make the unresolved overtime issue the critical path for all development. We can build employee administration, rostering, attendance and approved timesheets without pretending we have verified every statutory calculation.

## Recommended next milestone



Implementation milestone 1

&#x20;Employee-to-approved-timesheet vertical slice, with payroll readiness gating.



HR

Employer · Employees · Employment terms

WFM

Clients · Programmes · Shifts · Published rosters

Time and attendance

Clock events · Attendance intervals · Exceptions

Timesheets

Review · Corrections · Approval · Source lineage

Payroll readiness

Rule-pack inspection · Blocked calculations · Explicit reasons

The milestone produces real CLI and API functionality. It does not claim completed statutory payroll until the required rule packages are verified and implemented.

## Coding-agent prompt

# Prompt 7 — WorkforceOps First Working Backend Slice

## Objective

Implement the first functional vertical slice of WorkforceOps Lab using the existing, tested PostgreSQL foundation.

The target is a working CLI/API backend that takes an employee through:

Employee setup → Employment terms → BPO programme assignment → Shift roster → Clock events → Attendance → Timesheet approval → Payroll readiness assessment.

This is the first implementation milestone, not a new architecture or research phase.

The project already has:

- A single Go application and executable.
- A single PostgreSQL database.
- A real seven-version baseline migration covering the designed schema.
- Multi-employer-ready database relationships.
- Tested database privileges, constraints and immutability triggers.
- Employer-scoped API/CLI contracts.
- OpenAPI governance.
- A Phys-Pseudo business-rule specification methodology.
- A blocked SG-MOM-OT-001 package.

Use these as foundations. Do not redesign them.

## 1. Inspect before implementing

Read:

- `AGENTS.md`
- `plans/plan4.3_gate.md`
- `plans/prompt5_report.md`
- `mvp-scope.md`
- `domain-model.md`
- `database-design.md`
- `api-cli-contract.md`
- `singapore-rule-scope.md`
- `docs/workforceops/rules/adaptation-spec.md`
- `rules/SG-MOM-OT-001/`
- Existing Go source, migrations, tests, tasks and `openapi.yaml`.

Inspect Git status, branch and current commit.

All existing work is uncommitted. Preserve it. Do not reset, discard, overwrite or stage unrelated changes.

Verify the current test baseline before changing anything.

Identify which application packages, command infrastructure and API handlers already exist. Reuse them instead of creating parallel implementations.

## 2. Implementation boundaries

Architecture remains frozen:

- One Go executable.
- One REST API.
- One CLI.
- One shared application-service layer.
- One PostgreSQL database.
- PostgreSQL accessed through pgx and explicit SQL.
- Local-first Docker Compose.
- Loopback-only API; no authentication in v0.1.
- No web UI.
- No microservices.
- No event bus or asynchronous worker.
- No RAG, embeddings or AI components.

All business operations must be implemented once in application services. CLI commands and HTTP handlers are thin adapters over those services.

All employer-owned operations require explicit employer context.

Use the actual `app` database role for runtime operations.

Do not replace the tested baseline migration with fixture-only DDL.

If a schema correction is genuinely necessary, add a forward migration and update its documentation and tests.

## 3. Implement the first vertical slice

### A. Employer and employee administration

Implement:

- Employer selection and employer lookup by code or UUID.
- Employee creation, listing and retrieval.
- Employee status handling.
- Effective-dated employment-term creation and retrieval.
- Basic salary and fixed compensation records.
- Employment-category and Part IV eligibility fields as explicitly recorded data.
- Effective-date overlap validation.

Use synthetic reference data only.

Do not introduce real NRIC handling or claim complete statutory employment-record compliance.

### B. BPO configuration

Implement:

- Client creation and listing.
- Programme creation and listing.
- Team creation and employee membership.
- Global skill catalogue lookup.
- Employee skill assignment.
- Programme required-skill assignment.

Implement the minimum validations needed to prevent an employee without the required programme skills from being rostered.

Do not introduce advanced multi-skill optimization.

### C. Shift templates and roster management

Implement:

- Shift-template creation and listing.
- Shift segments for work, administration, paid breaks and unpaid meals.
- Simple deterministic rotation assignments where supported by the existing model.
- Draft roster creation.
- Employee shift assignments.
- Rest-day assignments.
- Roster publication.
- Immutable roster revisions.

Draft assignments must remain editable.

Published revisions and their assignments must remain immutable according to the tested database constraints.

Validate:

- Employee belongs to the selected employer.
- Employee has the required skills.
- Employee belongs to the applicable programme/team.
- No duplicate employee assignment for the same roster date/revision.
- No assignment on a designated rest day.

Roster scheduling is operational scheduling. Until SG-MOM-SHIFT reaches IMPLEMENTABLE, report legal validation as unavailable. Never label a roster legally compliant merely because operational validations pass.

### D. Clock events and attendance

Implement a deterministic attendance-import workflow.

Support a documented CSV input format containing:

- Source.
- Source event ID.
- Employee number.
- Event timestamp.
- Event timezone.
- Event type (`in` or `out`).

Requirements:

1. Preserve raw clock events as immutable records.
2. Deduplicate by `(employer_id, source, source_event_id)`.
3. Replayed imports return the original result without duplicating events.
4. Convert valid clock pairs into attendance intervals.
5. Preserve original UTC and local-time context.
6. Support cross-midnight shifts.
7. Preserve shift identity when an interval crosses calendar midnight.
8. Separate work, paid breaks and unpaid breaks.
9. Retain full actual recorded time.

Missing punches, unmatched events, out-of-order events, overlapping intervals and ambiguous attribution must become explicit exceptions.

Do not silently repair or invent missing clock events.

Attendance corrections must be auditable and create superseding records. Never modify original clock events.

The system must distinguish:

- Scheduled time.
- Recorded attendance.
- Actual work intervals.
- Paid time.

Do not collapse them into one hours-worked value.

Important: the 72-hour overtime working limit is not a rule for deleting or capping recorded payable work. Preserve all actual recorded intervals. Any potential statutory-hours-limit breach must be a separate exception, not a truncation of actual hours or an invented payroll deduction.

### E. Timesheet generation and approval

Implement:

- Timesheet generation from reviewed attendance intervals.
- Explicit source references from timesheet lines to the underlying intervals.
- Draft and submitted states.
- Review and rejection.
- Supervisor approval.
- Auditable corrections.
- Immutable approved timesheets and their lines.

Only eligible, attributable and reviewed attendance records should enter an approvable timesheet.

Do not silently convert missing or ambiguous time data into zero hours.

### F. Payroll readiness — no statutory calculations yet

Implement a read-only payroll readiness assessment using the existing rule-pack status and blocking matrix.

It should report:

- Employer.
- Pay period.
- Included employees.
- Applicable required rule packs.
- Current activation status and version.
- Blocked calculation components.
- Missing or ambiguous employee information.
- Whether the proposed payroll population is ready for a future calculation run.

Reuse the existing payroll approval predicate and rule activation contracts where applicable.

Do not implement:

- Overtime calculation.
- CPF calculation.
- SDL calculation.
- SHG contribution calculations.
- Leave deductions.
- Public-holiday pay.
- Net salary calculations.
- Payslip generation based on unverified calculations.

Do not create invented or placeholder financial outputs.

The current statutory rule packages are not all IMPLEMENTABLE. A payroll readiness response must say so explicitly.

Do not mark a payroll run approved merely because timesheets are approved.

Do not bypass the all-or-nothing payroll approval or explicit-exclusion rules.

## 4. CLI contract

Implement only commands required by the vertical slice.

Representative command workflow:

```
workforceops migrate status
workforceops seed demo

workforceops employer show

workforceops employee create --file employee.json
workforceops employee list
workforceops employee terms set EMP-001 --file terms.json

workforceops client create --file client.json
workforceops programme create --file programme.json
workforceops team create --file team.json
workforceops skill assign EMP-001 --file skills.json

workforceops shift-template create --file shift.json

workforceops roster create --file roster.json
workforceops roster show ROSTER-001
workforceops roster publish ROSTER-001

workforceops attendance import --file attendance.csv
workforceops attendance review
workforceops attendance correct --file correction.json

workforceops timesheet generate PERIOD
workforceops timesheet review
workforceops timesheet approve

workforceops payroll readiness PERIOD
```

Use JSON file inputs for complex commands.

Support machine-readable JSON output and meaningful error codes.

Every mutating CLI operation must carry an operator attribution label. Clearly document that it is attribution, not authentication.

## 5. REST API

Implement the corresponding versioned employer-scoped API endpoints.

Use the existing OpenAPI contract as the source of truth.

Do not serve endpoints that are not specified.

Only change a path from `draft` to `implemented` when:

- Its request/response schema is documented.
- Its handler uses the shared service layer.
- Its contract tests pass.
- Its authorization/employer-context boundaries are tested.

Expose:

- Employer and employee operations.
- Employment terms.
- Client/programme/team/skill operations.
- Shift templates and rosters.
- Roster publication and revisions.
- Clock-event import.
- Attendance review and correction.
- Timesheet generation, submission and approval.
- Payroll readiness.
- Health probes.

The read-only employer rule-pack activation/status endpoint should be implemented if it is already part of the approved contract.

Require `X-Operator` for audited API writes as documented. It is an attribution label, not identity verification.

## 6. Testing requirements

Use real PostgreSQL integration tests for database behavior.

Add tests for:

### HR

- Employee creation.
- Duplicate employee number within an employer.
- Same employee number in another employer.
- Effective-dated term overlap rejection.
- Employer context enforcement.

### WFM

- Required skill enforcement.
- Valid and invalid roster assignments.
- Draft assignment edits succeed.
- Published assignment edits fail.
- New roster revision preserves the previous published version.
- Cross-midnight shifts.

### Attendance

- Valid in/out pairing.
- Duplicate source-event import.
- Idempotent replay.
- Missing punch.
- Out-of-order events.
- Cross-midnight interval.
- Cross-calendar-date attribution.
- Employee/roster-assignment consistency.
- Correction creates a superseding record.
- Original clock event remains unchanged.

### Timesheets

- Generation from attendance intervals.
- Complete source lineage.
- Submission, approval and rejection.
- Draft editing succeeds.
- Approved changes fail.
- Duplicate approval request is idempotent.
- Cross-employer linkage is rejected.

### Payroll readiness

- Unverified required pack produces a blocked result.
- IMPLEMENTABLE pack version is reflected accurately.
- Missing or ambiguous required employee instruction blocks readiness.
- Explicit exclusion is visible.
- No financial calculation is performed.
- No blocked employee is silently omitted.

### Interface consistency

Run the same key workflow through CLI and REST API using separate identical PostgreSQL fixtures. Compare normalized business results.

Do not use mocked repositories as substitutes for database integration tests.

## 7. Deterministic demo environment

Implement or complete:

```
workforceops seed demo
```

Use Meridian CX Services Pte. Ltd. and a small synthetic workforce.

Provide:

- One Singapore legal employer.
- Two fictional BPO clients.
- Several programmes.
- Two or more teams.
- Employees with different skills.
- Day, evening and overnight shifts.
- Rotating roster examples.
- Cross-midnight attendance examples.
- Duplicate clock-event delivery.
- Missing-punch exception.
- Approved and rejected timesheet examples.

Seed data must be deterministic and re-runnable.

All synthetic records must be clearly identifiable as synthetic.

Do not claim anonymization, legal compliance or production behavior from synthetic fixtures.

## 8. Rule-specification safety

Keep `SG-MOM-OT-001` BLOCKED.

Do not treat provisional MOM transcripts as byte-grade pinned legislation.

Do not implement a calculation from the current Phys-Pseudo specification.

Maintain all unresolved `?` findings.

Apply the agreed semantic corrections to the existing documentation where needed:

- Separate the 72-hour working-time limit from remuneration.
- Do not cap payable hours at 72.
- Remove the artificial two-branch cap interpretation.
- Preserve the S$13.63636 versus S$13.60 rounding question.
- Distinguish the statutory minimum overtime rate from the contractual rate.
- Formalize eligibility inputs and applicability.
- Keep emergency-work treatment unresolved unless supported by the correct source.

Do not let this documentation cleanup become another research campaign. The rule package remains blocked and isolated from executable payroll calculations.

## 9. Scope restrictions

Do not implement:

- Any UI.
- Advanced forecasting.
- Erlang C.
- Automatic roster optimization.
- Intraday monitoring.
- Telephony or ACD integrations.
- Multi-employer SaaS authentication.
- PostgreSQL RLS.
- External payroll or HR integrations.
- Banking or statutory submission.
- Advanced leave calculations.
- Overtime or CPF calculations.
- Generic simulation engines.
- Phys-Pseudo interpreter/compiler.
- A generic plugin or rules engine.

Do not expand the project beyond the first working vertical slice.

## 10. Completion criteria

The slice is complete when a fresh local PostgreSQL environment can execute this workflow:

1. Start the database.
2. Run migrations.
3. Load the deterministic demo employer.
4. Create or retrieve an employee and employment terms.
5. Assign the employee to a BPO programme/team.
6. Create and publish a shift roster.
7. Import clock events.
8. Generate attendance intervals.
9. Review and correct an attendance exception.
10. Generate a timesheet.
11. Submit and approve the timesheet.
12. Request payroll readiness and receive an accurate, explicit blocked result while statutory packs remain unverified.

The same workflow must work through the CLI and the REST API.

Run:

- `task lint`
- `task build`
- `go test -count=1 ./...`
- PostgreSQL integration suites.
- OpenAPI validation.

Report exactly which tests ran and their results.

## 11. Stop condition and final report

Stop when the first vertical slice works end to end and its acceptance tests pass.

Do not continue into the next module merely because time remains.

Report:

1. Features implemented.
2. CLI commands delivered.
3. API endpoints delivered.
4. Database migrations changed, if any.
5. Test results.
6. Demonstrated end-to-end workflow.
7. Known limitations.
8. Remaining statutory blockers.
9. Next highest-impact implementation slice.

Do not claim WorkforceOps MVP payroll is complete. This milestone establishes the working HR/WFM/timesheet foundation and an honest payroll-readiness gate.

Preserve all existing uncommitted work and provide a clean, reviewable change summary.

## Why this is the right move

It creates tangible functionality while protecting us against three expensive mistakes:

1. Research paralysis: we do not wait for every statutory interpretation before building the operational system.
2. False compliance: we do not manufacture payroll amounts from unresolved legal rules.
3. Disconnected modules: we prove that HR, WFM, attendance and timesheets share one actual business workflow.

After this slice, the next major implementation target is the first independently verified payroll rule package and its calculation engine—not another architecture redesign.

One important distinction: this milestone is the first working backend foundation, not yet the finished Singapore payroll MVP. Actual payroll approval and payslip generation must wait until all required statutory calculations for the selected workforce profile pass their evidence and implementation gates.