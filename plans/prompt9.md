# Prompt 9 — WorkforceOps Milestone 1.1
# Adversarially Guided Integrity Remediation and Structural Audit

## Objective

Repair the confirmed data-integrity and workflow defects discovered during the Milestone 1 adversarial reviews, and identify any additional structural or architectural blockers that could prevent WorkforceOps Lab from becoming a reliable integrated Singapore HR, WFM and payroll backend.

The objective is to make the existing application trustworthy, not to redesign it or expand its feature set.

The implementation target is a corrected operational workflow:

`Employer → Employee → Employment Terms → BPO Assignment → Published Roster → Clock Events → Attendance → Approved Timesheet → Payroll Readiness`

The system must continue to use:

- One Go executable.
- One PostgreSQL database.
- One shared application-service layer.
- CLI and REST API.
- pgx and explicit SQL.
- Local-first development.
- Employer-scoped database relationships.
- Singapore statutory rule packages with independent verification gates.

The current implementation is uncommitted over baseline `f435158`. Preserve all existing work.

**This task does not authorize statutory payroll calculations.**

---

# 1. Establish the actual working-tree baseline

Before inspecting or modifying code:

1. Run `git status --short`, `git rev-parse HEAD` and `git diff --stat`.
2. Identify all modified, deleted, renamed and untracked entries.
3. Confirm the working tree is the same implementation state reviewed by Prompt 8.
4. Record the current migration versions and database setup.
5. Preserve all existing uncommitted work.

Read:

- `AGENTS.md`
- `plans/plan_prompt7.md`
- `plans/plan_prompt7_imp.md`
- `plans/milestone1_adversarial_review.md`
- `plans/milestone1_adversarial_review.md` findings and remediation sections
- `mvp-scope.md`
- `domain-model.md`
- `database-design.md`
- `api-cli-contract.md`
- `singapore-rule-scope.md`
- `docs/workforceops/rules/adaptation-spec.md`
- Relevant migration files and Go services.

Do not assume review-report line numbers still match the current source. Locate the actual functions and inspect the surrounding implementation.

Do not:
- Reset or clean the working tree.
- Stage or commit files.
- Discard previous work.
- Modify `bchat/`.
- Modify upstream Phys-Pseudo or Phys-Research repositories.
- Run destructive operations against any existing user database.
- Deploy to Neon or Fly.io.
- Implement CPF, SDL, SHG, overtime or other statutory calculations.

Use disposable PostgreSQL 16 databases for verification.

---

# 2. Stage One — Independent structural and architectural blocker audit

Before applying fixes, perform a bounded architectural audit.

The adversarial reviews found that the broad application architecture is sound. Confirm those clean areas, but do not assume the existing report identified every structural blocker.

Investigate the following dimensions.

## A. Application boundaries

Confirm:

- One executable remains.
- CLI and REST invoke shared application services.
- Handlers do not contain duplicated business logic or calculations.
- Repository operations consistently require employer context.
- Core packages do not import business modules incorrectly.
- Runtime application code uses the intended database role.
- The module registry and any retained module descriptors do not introduce obsolete ShiftHub/PayLoop dependencies into the executable.

Inspect whether `modules/sg-wfm-payroll` and its registry metadata are actually reachable from the application. If obsolete references are dead, do not build a new module framework to repair them.

## B. Business data integrity

Inspect the complete lifecycle across:

`employee → employment_terms → roster_assignment → clock_event → attendance_interval → timesheet_line_interval → timesheet → payroll_readiness`

Determine whether each important business invariant is protected at the correct layer:

- Go validation.
- Application-service transaction.
- Database constraint.
- Trigger or privilege.
- Independent integration test.

No critical invariant should rely exclusively on developer discipline when PostgreSQL can enforce it.

## C. Temporal and transaction architecture

Audit:

- UTC instants versus Singapore business dates.
- Employee timezone versus client/programme timezone.
- Cross-midnight shifts.
- Calendar-date and pay-period attribution.
- Effective-dated employment terms and employee status.
- Immutable historical records.
- Superseding attendance intervals.
- Atomicity of imports, pairing, corrections, timesheet generation and approval.

Identify whether the current schema can represent these concepts without contradictory update/immutability rules.

## D. Database and migration architecture

Inspect:

- Fresh-install migration path.
- Existing-baseline upgrade path.
- Migration role and runtime role.
- Default grants and revokes.
- Trigger ownership and behavior.
- Rollback definitions.
- Demo-seed and reset operations.
- Database-test isolation.
- Actual runtime connection configuration.

The earlier review did not execute the upgrade-from-baseline or rollback paths.

Determine whether the eight-version migration chain is safe to use both for a fresh install and for the actual previously established local schema.

If an old schema was never deployed and resetting it is an intentional pre-release decision, document that explicitly. Do not assume existing data can be discarded.

## E. API and CLI operability

Determine whether a real operator can complete the supported HR/WFM/timesheet workflow without SQL.

Inspect:

- Attendance interval discovery and correction.
- Roster revision and correction.
- Rejected timesheet recovery.
- Employment-term updates.
- Employee status transitions.
- Required programme skills.
- Pay-period availability.
- Payroll-readiness inputs.
- API-only versus CLI-only configuration.

Check that OpenAPI descriptions and maturity labels describe actual behavior.

## F. Configuration and deployment consistency

Inspect:

- `.env.example`
- `Taskfile.yml`
- `Dockerfile`
- `fly.toml`
- Migration scripts.
- Database initialization scripts.
- Module manifests.
- README and developer instructions.

The old ShiftHub/PayLoop executable paths must not remain active in the single-binary build or deployment configuration.

No actual cloud deployment is authorized.

## G. Additional blockers

Identify any further structural issue that could cause:

- Loss or corruption of attendance evidence.
- Duplicate time or pay inputs.
- Incorrect historical payroll populations.
- Mutation of approved artifacts.
- Cross-employer relationships.
- Unrecoverable workflow states.
- Unreliable fresh installations.
- Divergence between CLI and API.
- Incorrect financial precision.
- A false payroll-readiness result.

Report only evidence-supported defects.

Do not invent architectural problems merely to extend the review.

### Structural audit deliverable

Before making code changes, create a concise internal findings inventory in:

`plans/milestone1_1_structural_audit.md`

Classify each result:

- Confirmed structural blocker.
- Confirmed implementation defect.
- Risk requiring reproduction.
- Untested hypothesis.
- Clean area.

For each structural blocker, explain whether it can be fixed within the frozen architecture.

If a finding genuinely requires a change to a frozen architectural decision, explain the required decision and stop that particular change for approval. Continue independent fixes that do not depend on it.

Do not begin an additional architecture redesign.

---

# 3. Stage Two — Reconcile contradictory review findings

The earlier review and the newer synthesis disagree about whether certain defects were reproduced.

Re-test the following before fixing them.

## A. Overlapping payroll periods

The earlier report says a live reproduction showed attendance intervals linked into two overlapping pay periods.

The newer report classifies this as untested.

Create two overlapping periods for the same employer and verify whether a single interval can enter both timesheets.

If confirmed:

- Prevent overlapping effective payroll periods.
- Prevent a time interval from being allocated to more than one payroll line.
- Enforce employee consistency between interval, timesheet line and timesheet.
- Add database constraints wherever practical.
- Validate allocation in the service.
- Add fresh-PostgreSQL regression tests.

Do not use roster start date as an arbitrary substitute for correct pay-period attribution.

## B. Cross-midnight attribution

The earlier report claims work after Singapore midnight was incorrectly attributed to the shift's starting date.

The newer synthesis classifies this as untested.

Run a controlled overnight-shift test.

The domain contract says each attendance record represents one business date and cross-midnight work must be split at local midnight.

If confirmed, implement that contract.

Preserve:
- Original clock events.
- UTC instants.
- IANA timezone.
- Local business-date context.
- Shift identity.
- Correct roster assignment relationships.
- Break evidence.
- Source lineage.

Do not let the PostgreSQL session timezone determine a payroll business date.

## C. Late break events

Import clock-in and clock-out first.

Then import the corresponding break start/end events separately.

Determine whether the second import can attach the actual break to the existing work interval.

If it cannot, repair the pairing/correction process without modifying raw clock events.

Unresolved or ambiguous breaks must remain exceptions.

## D. Concurrent attendance corrections

Run two concurrent corrections against the same original interval.

Determine whether multiple successors can be created.

If confirmed:
- Serialize the correction operation.
- Enforce one successor per original interval at the database level.
- Preserve the complete supersession chain.
- Ensure attendance aggregates cannot count both the original and its successor.

## E. Migration upgrade and rollback

Test:
- Fresh install.
- Upgrade from the actual previously established schema.
- Migration status.
- Down/up behavior on a disposable database.

Do not drop existing user data to make a migration test pass.

Report any historical migration limitation explicitly.

---

# 4. Stage Three — Mandatory data-integrity remediation

Fix the following confirmed blockers.

## P0-1 — Approved timesheets must not silently diverge from corrected attendance

`Correct()` currently permits an attendance interval linked to a submitted or approved timesheet to be superseded, while the old approved timesheet remains unchanged.

Requirements:

1. Find whether the interval is allocated to any timesheet.
2. Reject correction when it belongs to a submitted or approved timesheet unless a controlled correction workflow is invoked.
3. Preserve the original approved timesheet.
4. Provide an auditable correction/reissue process.
5. Ensure payroll readiness cannot treat a stale approved timesheet as complete.
6. Preserve the link between old and new records.

No silent update or deletion of approved financial source data.

Add a cross-service regression test covering:

`Generate → Submit → Approve → Correct linked attendance`

The approved result must remain immutable and the correction must have an explicit downstream disposition.

## P0-2 — Unbounded clock-event spans must never become clean payable time

A lone clock-in on one date must not pair with an unrelated clock-out several weeks later and become a reviewed work interval.

Requirements:

- Define a defensible pairing boundary from the supported shift model.
- Prefer matching against roster assignment and expected shift identity.
- Reject implausible or ambiguous pairings.
- Preserve unmatched raw events.
- Persist a machine-readable exception.
- Prevent affected attendance from generating an approved timesheet until reviewed.
- Test late clock-outs and very long spans.

Important:

This is not permission to truncate actual payable hours.

Do not discard work or use the 72-hour monthly statutory limit as a pairing threshold.

A 72-hour breach is a separate working-time compliance exception. Actual recorded work remains preserved.

## P1-1 — Freeze-trigger re-parenting bypass

Inspect every parent-state immutability trigger.

For UPDATE operations, validate both:

- The OLD parent.
- The NEW parent.

A row must not be moved from a draft parent into a published or approved parent.

Scope parent lookups by employer as well.

Add direct PostgreSQL tests under the `app` role covering re-parenting into frozen:
- Timesheets.
- Payroll runs.
- Rosters/revisions.
- Other approval-gated parents.

## P1-2 and P1-21 — Duplicate and overlapping attendance intervals

Different source event IDs may represent duplicate underlying work events.

Do not rely exclusively on `(employer_id, source, source_event_id)` for interval uniqueness.

Requirements:

- Detect overlapping current work intervals for the same employee.
- Reject or flag duplicate overlapping intervals before timesheet generation.
- Prevent correction intervals from overlapping live sibling intervals.
- Preserve historical superseded intervals without treating them as current work.
- Serialize operations where concurrent imports/corrections could duplicate intervals.
- Add database protection where the current/historical model permits it.
- Use service validation and locking where necessary.

Do not use a constraint that incorrectly prohibits legitimate historical supersession.

## P1-3 — Conflicting duplicate source-event delivery

When an existing `(employer_id, source, source_event_id)` is imported again:

- Compare employee, event kind and timestamp with the stored event.
- An identical replay is an idempotent duplicate.
- A different payload under the same source event ID is a conflict.
- Do not silently discard the conflicting payload.
- Preserve the original event.
- Surface a machine-readable conflict requiring review.

## P1-4 and P1-6 — Ambiguous break evidence and exception lifecycle

An unclassifiable break must not be silently treated as a correctly paired work/break event.

Requirements:

- Do not cite ambiguous break IDs as successfully consumed by unrelated work intervals.
- Keep unclassifiable break evidence discoverable.
- Persist the associated exception.
- Never let an unrelated attendance correction reset an unresolved exception to `reviewed`.
- Clear an exception only after a successful validation pass establishes that its cause has been resolved.
- Recalculate day aggregates atomically with the correction.

## P1-5 — Singapore business-date filtering

Find all queries that cast `TIMESTAMPTZ` directly to `DATE`.

Replace session-timezone-dependent behavior with the documented employee/business timezone policy.

At minimum test:
- Midnight Singapore time.
- Midnight UTC.
- Cross-midnight work.
- Unpaired events immediately before and after a pay-period boundary.
- Different imported event timezone values.

The server's default timezone must not change attendance or timesheet results.

## P1-7 and P1-24 — Runtime and migration role separation

Restore and prove the two-role arrangement:

- `migration_admin` performs DDL and owns protected schema objects.
- `app` runs normal application queries and mutations.

Requirements:

1. Honour `MIGRATION_DATABASE_URL` in migration commands.
2. Use `DATABASE_URL` for runtime.
3. Ensure the actual local fresh-install workflow provisions the required roles.
4. Make migrations fail loudly when required roles or privileges are missing.
5. Verify post-migration privileges and object ownership.
6. Do not rely on `IF EXISTS app` to silently skip essential protections.
7. Run service and parity tests through the actual `app` role wherever they represent application behavior.
8. Retain elevated database access exclusively for migration and isolated test setup.

Do not solve the issue by granting `app` broad DELETE privileges or bypassing triggers.

Run an actual application workflow using `app`:
- Employee operations.
- Attendance import.
- Correction.
- Timesheet generation.
- Approval.
- Readiness.

## P1-8 — Employee status transitions

Support repeated valid effective-dated transitions.

The initial active status must have a coherent historical representation.

Close the previous open status interval and create the new interval atomically.

Define a transition matrix, including:
- Active to suspended.
- Suspended to active.
- Active to resigned.
- Active to terminated.
- Terminal-status behavior.

Do not allow arbitrary resurrection of terminated employees.

Historical records must remain intact.

## P1-9 — Historical employment state and payroll population

Readiness must not silently omit an employee with an approved timesheet merely because the employee is currently terminated or resigned.

Determine the intended payroll population using the employee's effective-dated employment state and relevant pay period.

Preserve the status and employment-term information applicable to the approved timesheet.

Prevent retrospective employment-term changes from silently changing a historical approved payroll result.

If historical terms are corrected, require an explicit auditable correction mechanism.

Readiness must identify:
- Employees expected in the period.
- Employees with approved timesheets.
- Employees missing timesheets.
- Explicitly excluded employees.
- Employees with incomplete or blocked statutory inputs.

## P1-10 — Roster publication validation

Publishing must revalidate all assignments in one transaction.

Validate:
- Non-empty roster, unless an explicit documented reason permits an empty roster.
- Employee status.
- Effective team membership.
- Programme skill requirements.
- Skill validity.
- Roster-week date boundaries.
- Rest-day compatibility.
- Assignment uniqueness.
- Current revision identity.

Do not assume an assignment remains valid merely because it passed checks when first inserted.

Persist the result of operational validation.

Do not describe operational checks as Singapore legal validation. `SG-MOM-SHIFT` remains unverified.

## P1-11 — Roster revision workflow

Published rosters must be correctable without mutating published history.

Implement a minimum `Revise` workflow:

1. Preserve the existing published revision.
2. Create the next revision.
3. Copy existing assignments into the new draft revision.
4. Advance the current revision pointer atomically.
5. Permit edits only on the new draft.
6. Revalidate before publishing.
7. Populate the revision's publication metadata.
8. Preserve the old version for audit and replay.

Expose the workflow through CLI and REST.

Do not implement automatic schedule optimization.

## P1-12 — Attendance interval discovery

Add an employer-scoped read endpoint to list attendance intervals by employee and date range.

It must expose:
- Interval ID.
- Employee.
- Start and end.
- Local business date.
- Interval kind.
- Current/superseded status.
- Source provenance.
- Relevant exception status.
- Roster assignment reference.

Provide a corresponding CLI command.

The operator must be able to review and correct attendance without accessing PostgreSQL directly.

## P1-13 — Fresh-database test setup race

Reproduce the parallel test failure against a clean disposable PostgreSQL server.

Correct the database-creation, extension and migration concurrency issue.

The supported test workflow must pass reliably.

A temporary serial workaround is acceptable only if:
- It is explicitly the canonical test command.
- The reason is documented.
- The normal developer/CI workflow invokes it.
- The underlying nondeterminism is recorded for subsequent remediation.

Prefer reliable database provisioning and migration coordination over silently rerunning failed tests.

---

# 5. Additional workflow and integrity corrections

The following findings are not all classified as Milestone 1 blockers by the consolidated report. Nevertheless, address those that are required to make the next payroll implementation safe and the existing advertised workflow usable.

## A. Effective-dated employment terms

Support closing the current terms period and creating a successor atomically.

Do not allow overlapping terms or untracked backdated changes over approved periods.

Keep historical terms immutable once closed, except through an explicit correction process.

## B. Rejected timesheet recovery

A rejected timesheet must not permanently prevent the employee from receiving a corrected timesheet for that period.

Implement an explicit reopen or supersession workflow.

Preserve the rejected artifact and its reason.

Add:

`Generate → Submit → Reject → Correct → Resubmit → Approve`

## C. Payroll-readiness exclusions

Validate that any referenced payroll run:
- Exists.
- Belongs to the same employer.
- Belongs to the assessed pay period.
- Has a compatible status.

A run from another period must not affect readiness.

An empty included population must not be approved through a vacuously true predicate. Align readiness and approval behavior.

An empty or all-superseded timesheet must not count as complete.

## D. Money representation

Before implementing any financial calculation, eliminate `float64` monetary inputs.

Use an exact representation:
- Integer minor units where suitable.
- Decimal strings at API boundaries.
- `NUMERIC(12,2)` in PostgreSQL.
- `math/big.Rat` or another deliberately chosen exact intermediate representation where required.

Do not use binary floating-point arithmetic for money.

Validate non-negative values and supported decimal precision in Go, returning machine-readable validation errors rather than exposing database errors.

Do not implement statutory rounding in this milestone.

## E. Transaction boundaries and audit

Review the complete operations:
- Clock import and pairing.
- Attendance correction and daily aggregate updates.
- Timesheet generation.
- Timesheet submit/approve/reject.
- Roster publication and revision.
- Employment status and terms changes.
- Payroll exclusions.

Related mutations and their resulting aggregates must commit atomically.

If the application crashes after saving input evidence but before pairing, replay must be able to repair the pending work.

Persist audit events for consequential mutations, including actor, reason, entity and before/after references where applicable.

The actor string remains attribution, not authentication.

Do not silently default missing audited operators to generic `api` or `cli` identities where the contract requires an operator.

## F. Idempotency

Review the existing `idempotency_key` table and documented `Idempotency-Key` promises.

For every critical mutation, choose one explicit behavior:

- Enforce the idempotency key and return the original result on replay; or
- Use a documented domain-specific idempotency mechanism and remove any unsupported generic contract claim.

At minimum, ensure deterministic replay for:
- Clock-event import.
- Roster publication.
- Timesheet generation.
- Timesheet approval.
- Attendance corrections.

Do not introduce a large generalized workflow framework merely to support a few commands.

## G. API errors and contracts

Align runtime and OpenAPI error envelopes.

Return a machine-readable code and request ID.

Do not expose raw PostgreSQL error messages, table names, column names or DSNs in client-facing errors.

Malformed dates and times must return validation errors, not internal errors.

Ensure served routes, request/response schemas and maturity labels agree.

## H. API and CLI workflow completeness

Provide the minimum operations required to configure and use the supported business workflow without raw SQL.

At minimum verify:
- Employer selection/setup for the single-employer profile.
- Pay-period availability.
- Employee terms and compensation inputs.
- Required skills and memberships.
- Attendance interval listing and correction.
- Roster revision.
- SHG instruction capture where readiness requires it.

Where an optional operation is intentionally deferred, document that clearly rather than exposing an incomplete implementation.

Do not expand into a generic HR administration platform.

---

# 6. Additional structural verification required before closing the milestone

Resolve the following findings or hypotheses:

1. Whether overlapping pay periods actually permit double allocation.
2. Whether cross-midnight work is incorrectly assigned to one business date.
3. Whether late break events can attach after initial in/out pairing.
4. Whether duplicate intervals can be generated concurrently.
5. Whether one attendance interval can be allocated to multiple timesheet lines.
6. Whether timesheet-line employee identity is enforced at the database level.
7. Whether corrections crossing dates resolve the correct roster assignment.
8. Whether rotation assignments have actual materialization into rosters. If not, either implement the promised simple deterministic behavior or clearly mark rotation generation as deferred.
9. Whether roster publication metadata is written to the current revision.
10. Whether the fresh migration chain upgrades the existing pre-pivot schema safely.
11. Whether seed demo and seed demo --reset actually work under their intended roles.
12. Whether Dockerfile and Fly configuration reference deleted executable names.
13. Whether module registry references to ShiftHub/PayLoop are reachable or dormant.
14. Whether all service and parity tests operate with production-like application permissions.
15. Whether all indeterminate statutory paths map to explicit blocked responses.

For each item, mark:
- Confirmed and fixed.
- Confirmed but deferred, with justification.
- Tested and disproved.
- Still unverified.

Do not silently convert an untested hypothesis into a defect.

---

# 7. Architectural constraints — do not reopen

The following remain frozen:

- One Go executable.
- One PostgreSQL database.
- One API and one CLI.
- One shared application-service layer.
- No frontend or UI.
- No microservices.
- No event bus or distributed synchronization.
- No generic plugin framework.
- No rules engine.
- No Phys-Pseudo interpreter or compiler.
- No PostgreSQL RLS in v0.1.
- One employer operationally in v0.1; schema remains employer-scoped and multi-employer-capable.
- Local-first operation.
- No actual cloud deployment without separate approval.

Preserve all confirmed clean areas identified in the adversarial report.

If a proposed correction appears to require abandoning one of these decisions, explain the structural problem and request a decision rather than making the change silently.

---

# 8. Testing and acceptance gates

Use real PostgreSQL 16.

All destructive test operations must target explicitly disposable databases whose names and connection targets are verified before use.

## A. Mandatory regression cases

Test:

- Correct attendance after timesheet generation.
- Attempt to correct attendance after timesheet submission.
- Attempt to correct attendance after approval.
- Month-long or otherwise implausible unmatched clock spans.
- Duplicate interval imports using different event IDs.
- Conflicting replay under the same event ID.
- Ambiguous break evidence.
- Correction on a day with an unresolved exception.
- Singapore midnight and cross-midnight shifts.
- Overlapping pay periods.
- Concurrent imports and corrections.
- Concurrent roster publication and revision.
- Repeated employee status transitions.
- Effective-dated term changes.
- Terminated employee with an approved historical timesheet.
- Payroll exclusion belonging to another period.
- Rejected timesheet recovery.
- Empty payroll-readiness population.
- Trigger re-parenting into an approved/published parent.
- Fresh-install migration with correct roles.
- Runtime service writes as `app`.

## B. End-to-end acceptance

A fresh database must execute the supported workflow without SQL intervention:

1. Migrate.
2. Bootstrap/seed the employer.
3. Create or retrieve an employee and terms.
4. Configure programme, team and skills.
5. Create and publish a roster.
6. Import clock events and actual break evidence.
7. Review attendance and exceptions.
8. Retrieve interval IDs through the supported interface.
9. Correct a permitted attendance interval.
10. Generate a timesheet.
11. Submit, reject/recover where applicable, and approve.
12. Assess payroll readiness with correct historical population and blockers.

Run the scenario through both CLI and REST using shared services.

Compare actual persisted business results, not merely exit codes or HTTP status codes.

The demo workflow must not require manual SQL to discover interval IDs or complete a promised operation.

## C. Test commands

Run:

- `task lint`
- `task build`
- `go test -count=1 ./...` with the required disposable PostgreSQL database.
- DB-gated integrity suites.
- Relevant concurrency tests.
- Fresh database migration.
- Upgrade-from-baseline test or an explicit documented pre-release migration decision.
- CLI E2E.
- REST E2E.
- OpenAPI validation.
- `git diff --check`.

Do not report a passing result if database suites skipped.

Do not silently switch to serial tests after a parallel failure without documenting the decision.

## D. Completion conditions

Milestone 1.1 can PASS only when:

1. Both P0 defects are corrected.
2. All 13 identified M1-blocking P1 findings are corrected or demonstrably disproved by controlled tests.
3. Prior-report discrepancies have been adjudicated.
4. No remaining P0 or P1 finding makes attendance, timesheet or payroll-readiness data untrustworthy.
5. Runtime and migration roles work on a fresh install.
6. The correction and recovery workflows work without SQL.
7. Default test execution is reliable.
8. CLI and REST produce equivalent persisted business results.
9. `SG-MOM-OT-001` remains BLOCKED.
10. No statutory payroll calculations or fabricated golden fixtures have been introduced.

---

# 9. Explicitly deferred items

Do not attempt to fix all 72 report entries.

The following can remain deferred unless the structural audit proves they are directly blocking the supported workflow:

- Pagination for every list endpoint.
- Full REST parity for optional rotation and catalogue management.
- General-purpose API idempotency infrastructure beyond the critical operations.
- Generic simulation and scenario engines.
- Advanced concurrency/performance benchmarking.
- Multi-employer authentication and RLS.
- SSO.
- External payroll or telephony integrations.
- Production deployment.
- Cosmetic CLI improvements.
- Extensive module registry redesign.
- Advanced roster optimization.

Where an API contract is already misleading, correct the documentation or make the smallest appropriate implementation change.

---

# 10. Final deliverables

Create:

`plans/milestone1_1_remediation_report.md`

Include:

1. Initial Git and working-tree state.
2. Structural audit findings.
3. Confirmed defects and exact remediation.
4. Contradictory review findings and their disposition.
5. New defects discovered.
6. Files and migrations changed.
7. Database constraints and transaction changes.
8. Test commands actually executed and results.
9. Fresh-install and upgrade findings.
10. CLI/REST E2E evidence.
11. Remaining limitations and untested hypotheses.
12. Final M1.1 PASS/FAIL verdict.
13. Smallest next implementation milestone.

Keep all work uncommitted and preserve the working tree.

Do not start payroll calculations in this task.

The goal is to repair the existing product, establish its structural integrity and prepare it for the first independently verified payroll calculation slice.

**Stop once the M1.1 acceptance gate passes. Do not expand the architecture or continue into unrelated features.**