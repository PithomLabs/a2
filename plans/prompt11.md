I recommend naming this Milestone 1.2 — Payroll Input Integrity Closure.

The objective is to fix the eight service-reachable defects, close the critical PostgreSQL bypasses, repair first-install and onboarding, and verify the complete evidence chain. The agent must not turn this into another comprehensive rewrite or attempt to implement all 72 findings.

The following prompt incorporates the latest adversarial review, including the newly discovered `SECURITY DEFINER`, correction-chain and direct-SQL vulnerabilities.

# Prompt 11 — WorkforceOps Milestone 1.2

# Payroll Input Integrity Closure

## Objective

Implement the smallest justified remediation that makes WorkforceOps trustworthy enough to begin its first independently verified payroll calculation.

Milestone 1.2 follows the final adversarial review documented in:

`plans/milestone1_1_final_adversarial_review.md`

The review identified eight service-reachable P1 defects affecting attendance, employment history and payroll inputs, additional PostgreSQL integrity bypasses, and two first-install/onboarding blockers.

The objective is to close these defects, strengthen the relevant database invariants, and independently verify the complete operational evidence chain.

The target workflow remains:

`Employer → Employee → Employment Terms → BPO Assignment → Published Roster → Clock Events → Attendance → Approved Timesheet → Payroll Readiness`

This is an integrity-remediation milestone, not a new feature-development programme.

**The first statutory payroll calculation is still deferred.**

Do not implement CPF, SDL, SHG, overtime, gross salary, deductions, net pay or financial payslips.

---

# 1. Repository baseline and preservation

Before making changes:

1. Inspect `git status --short`.
2. Record `git rev-parse HEAD`.
3. Confirm the baseline is `f43515883668c917edac31be49a7cf8334eca5c2`.
4. Record all staged, unstaged, renamed, deleted and untracked entries.
5. Read the complete Milestone 1.1 final adversarial review.
6. Read the Milestone 1.1 structural audit and remediation report.
7. Inspect the existing migrations, especially `00009_m11_integrity.sql`.
8. Record all active PostgreSQL database targets before running any database operation.

Preserve all existing uncommitted work.

Do not:

- Reset or clean the working tree.
- Stage or commit changes.
- Discard previous work.
- Modify `bchat/`.
- Modify upstream Phys-Research or Phys-Pseudo.
- Deploy to Neon or Fly.io.
- Modify previously applied migrations.
- Drop, recycle or overwrite an existing user database.
- Implement statutory payroll calculations.

Use fresh disposable PostgreSQL 16 databases for verification.

All work remains uncommitted.

---

# 2. Frozen architectural decisions

Preserve the following:

- One Go application and executable.
- One PostgreSQL database.
- One REST API.
- One CLI.
- One shared application-service layer.
- pgx and explicit SQL.
- Employer-scoped database relationships.
- One employer operationally in v0.1.
- Multi-employer-ready schema.
- `migration_admin` for migrations and protected schema operations.
- `app` for ordinary runtime operations.
- Loopback-only API.
- No UI.
- No microservices.
- No event bus.
- No distributed synchronization.
- No RLS in v0.1.
- No generic plugin framework.
- No generic payroll rules engine.

Singapore remains the initial jurisdiction.

Phys-Research and Phys-Pseudo remain the methodology for statutory and financial rules.

Do not reorganize the repository to accommodate hypothetical future countries.

If a finding genuinely requires changing a frozen architectural decision, document the exact structural problem and stop that change for approval. Continue independent work that does not depend on it.

---

# 3. Stage A — Reconfirm findings against actual code

Do not blindly implement the adversarial report's proposed fixes.

For every finding, inspect the actual function, schema constraint, trigger and regression tests.

Distinguish:

- Confirmed defect.
- Previously fixed behavior.
- New regression.
- Untested hypothesis.
- Accepted limitation.

The eight service-reachable findings are:

1. `F-OVERLAP-BLOCK`
2. `F-MIDNIGHT-BREAK`
3. `F-APPROVED-DEADLOCK`
4. `F-COMP-RETRO`
5. `F-TZ-CLASSIFY`
6. `F-CORRECT-SPLIT`
7. `F-LATEBREAK-DESYNC`
8. `F-EMP-STATUS`

The database protection findings include:

- `F-IS-CURRENT-BYPASS`
- `F-XEMP-SUPERSEDE`
- `F-PG-TEMP`
- `F-INTERVAL-SELFLOOP`
- `F-LINE-MOVE`
- `F-SHEET-KEYS`
- `F-ROSTER-REOPEN`
- `F-HR-DIRECT-WRITE`
- `F-SINGLECHAIN-UPDATE`
- `F-CHAIN-SCOPE`

Fresh-install and onboarding findings:

- `F-INITROLES`
- `F-SUPERUSER-DEFAULT`
- `P1-23` — employer creation is not available through the supported interfaces.

Do not create duplicate finding IDs for equivalent problems.

If a proposed correction would break a previously verified invariant, stop and resolve the interaction before implementing it.

---

# 4. Stage B — Fix the eight service-reachable P1 defects

These are the highest priority because they affect the normal business workflow and can corrupt or permanently block payroll inputs.

## B1. F-OVERLAP-BLOCK — Duplicate vendor deliveries must be recoverable

Current defect:

A duplicate work span delivered under fresh source event IDs can create an `overlapping_span` exception. The import succeeds, but the events remain unpaired and block timesheet generation indefinitely.

Do not solve this by blindly treating every overlap as a duplicate.

Implement an explicit operator disposition.

The operator must be able to:

- Inspect the original and overlapping events.
- Identify the exact event IDs involved.
- Mark a confirmed duplicate delivery as duplicate.
- Provide a reason and operator attribution.
- Preserve all raw events.
- Retain the exception and its disposition history.
- Permit timesheet generation using only the validated original work evidence.

A legitimate second shift or genuinely overlapping work record must not be silently discarded.

Requirements:

1. Distinguish duplicate delivery from genuine overlapping work.
2. Persist the disposition.
3. Ensure an event cannot be independently allocated after being confirmed duplicate.
4. Make duplicate disposition idempotent.
5. Preserve an auditable record of the decision.
6. Recalculate affected day eligibility.
7. Ensure a handled duplicate does not permanently prevent generation.
8. Ensure an unresolved legitimate overlap continues to block approval.

Add an E2E test:

`Duplicate feed with fresh IDs → Inspect exception → Confirm duplicate → Generate timesheet → Verify one valid allocation`

## B2. F-MIDNIGHT-BREAK — Break intervals spanning midnight

Current defect:

A break event spanning local midnight does not fit inside either half of a midnight-split work interval and remains unpaired.

Implement break splitting using the same employee-local calendar boundaries as work intervals.

For example:

- Work: 22:00–06:00.
- Break: 23:45–00:15.

The resulting break evidence must be represented on the appropriate two local-date segments.

Preserve:

- Original break events.
- Original timestamps.
- IANA timezone.
- Paid/unpaid classification.
- Source-event references.
- Correspondence to both work segments.
- Supersession history.

Do not lose or double-count work around the break.

Support:

- Break exactly at midnight.
- Break starting before midnight and ending after midnight.
- Break entirely within one date.
- Missing break start/end.
- Duplicate break events.
- Late break delivery.
- Paid and unpaid breaks.

The break classification must not depend on the host process timezone.

Add regression fixtures for all relevant boundaries.

## B3. F-APPROVED-DEADLOCK — Make approved-timesheet correction workflow functional

Current defect:

`Correct()` rejects any interval linked to a submitted or approved timesheet.

Reissuing the approved timesheet creates a new draft, but the original approved link remains. Consequently, the reissue does not unblock the correction.

The original approved artifact must remain immutable.

Implement chain-aware correction.

Required behavior:

1. A direct correction against an effective approved timesheet is rejected.
2. The operator explicitly reissues the approved timesheet.
3. A correction draft is created and linked to the original.
4. The approved original remains unchanged.
5. Attendance correction is permitted only when an appropriate successor draft exists to receive the corrected evidence.
6. The correction supersedes the original attendance interval.
7. The correction draft is regenerated atomically with the attendance change.
8. The old approved timesheet retains its historical inputs.
9. Readiness must report an attendance/timesheet conflict until the correction version is approved.
10. Approval of the correction establishes the new effective timesheet.
11. A rejected correction remains immutable and can be replaced by another correction version.

Do not solve this by simply ignoring all approved timesheet links.

The correction must prove that a valid downstream correction path exists.

Add a regression test covering:

`Generate → Submit → Approve → Attempt correction → Reissue → Correct → Regenerate → Submit → Approve`

Also test:

- Correction after reissue of an already corrected approved sheet.
- Concurrent correction and reissue.
- Rejected correction recovery.
- Readiness before and after correction approval.
- Original approved evidence remains unchanged.

## B4. F-COMP-RETRO — Protect approved payroll compensation inputs

Current defect:

A fixed allowance can be added retroactively after a timesheet has been approved without readiness detecting the change.

Implement the same historical-integrity discipline used for employment terms.

Requirements:

1. Compensation entries must be effective-dated.
2. Compensation updates affecting an approved period must be rejected or recorded as an explicit conflict requiring a controlled correction.
3. Existing approved-period compensation must not silently change.
4. Readiness must examine compensation as well as employment terms and SHG instructions.
5. Backdated compensation must not bypass conflict detection by preserving `created_at`.
6. Any permitted correction must be auditable.
7. Compensation effective dates must be consistent with their parent employment terms.
8. Monetary inputs must retain exact decimal representation.

Add tests:

- Add compensation before approval.
- Add compensation after approval for a past period.
- Add compensation for a future period.
- Attempt to update an existing historical compensation record directly as `app`.
- Readiness after a retroactive change attempt.
- Corrected compensation under an explicitly controlled workflow, if supported.

Do not create an unverified payroll calculation.

## B5. F-TZ-CLASSIFY — Eliminate host-timezone-dependent break classification

Current defect:

Break classification can use `time.Local` through pgx timestamp decoding. Running the application on a UTC host can shift Singapore shift windows and misclassify valid breaks.

The application's result must not depend on the host machine timezone.

Requirements:

- Use the explicit event timezone.
- Resolve timezone IDs through `time.LoadLocation`.
- Convert timestamps deliberately before classification.
- Never assume `start.Location()` represents the employee's business timezone.
- Preserve original UTC instants.
- Preserve employee/business IANA timezone.

Run the actual service test process with:

`TZ=UTC`

Also test a mixed-zone input fixture.

Prove that the same input evidence produces identical business results regardless of host process timezone.

A test that only runs under Asia/Singapore is insufficient.

## B6. F-CORRECT-SPLIT — Corrections crossing local midnight

Current defect:

The import pathway splits intervals by local midnight, but `Correct()` can create a single replacement interval spanning two business dates.

Reuse the same canonical interval-splitting logic for both import and correction.

Requirements:

1. A cross-date correction creates the correct per-date successor intervals.
2. Every segment retains the original shift identity where appropriate.
3. Each date resolves its own roster assignment.
4. If the new date has no valid roster, do not reuse the old day's roster assignment.
5. Create an explicit `no_roster_link` or equivalent exception where required.
6. Preserve all source references.
7. Recalculate both affected daily records.
8. Write audit evidence for both affected dates.
9. Commit the correction, intervals, aggregates and audit records atomically.

Test:

- Same-date correction.
- Singapore midnight crossing.
- Correction at midnight exactly.
- Correction one second before/after midnight.
- Correction crossing a pay-period boundary.
- New date with a valid different assignment.
- New date without an assignment.
- Correction overlapping an existing interval.

## B7. F-LATEBREAK-DESYNC — Late break evidence after timesheet approval

Current defect:

A break arriving after timesheet approval can cause attendance successors to be created while the approved timesheet continues to represent the previous work interval.

This must never happen silently.

Implement a linkage-aware late-break workflow.

Requirements:

- Raw break evidence is always preserved.
- If linked timesheets are draft, regenerate them atomically.
- If linked timesheets are submitted or approved, do not mutate their source evidence silently.
- Persist a blocking evidence-conflict or correction-required exception.
- Allow the operator to reissue the affected timesheet through the controlled correction workflow.
- Once a valid correction draft exists, attach the break evidence to its corresponding attendance successors.
- Regenerate that correction draft transactionally.
- Preserve the original approved timesheet and its original intervals.
- Readiness remains blocked until the evidence conflict has been resolved by the approved correction workflow.

Test:

`Clock events → Attendance → Generate → Approve → Late break import → Conflict → Reissue → Attach break → Regenerate → Approve correction`

Also test:

- Late break after a submitted but unapproved sheet.
- Late break during an existing correction draft.
- Duplicate late break.
- Break spanning midnight.
- Paid break.
- Unclassifiable break.

Do not lose incoming evidence merely because an earlier timesheet has been approved.

## B8. F-EMP-STATUS — Prevent non-active employees from receiving publishable shifts

Current defect:

Terminated employees can retain team and skill records, receive assignments and have those hours flow into timesheets.

Add employee-status validation to the shared placement validation logic.

At both assignment and publication:

- Employee must be active on the relevant scheduled dates.
- Suspended employees must not be scheduled.
- Resigned or terminated employees must not be scheduled after their effective end date.
- Historical shifts before termination remain valid.
- Effective-dated status must be checked for the roster's dates, not only current wall-clock status.

Do not invalidate legitimately worked historical time merely because the employee is now terminated.

Add tests for:

- Active employee.
- Suspended employee.
- Resigned employee.
- Terminated employee.
- Employee terminated after the roster period.
- Employee terminated during a multi-day roster.
- Assignment created while active but employee becomes inactive before publication.

---

# 5. Stage C — Close PostgreSQL application-role bypasses

The project deliberately tests the `app` role as a legitimate adversary.

These protections must therefore hold even when SQL is issued directly as `app`, not only when a well-behaved Go service is used.

Use a new forward migration beginning at `00010`.

Never alter `00001`–`00009`.

## C1. F-IS-CURRENT-BYPASS

Current defect:

`app` can insert an attendance interval with explicit `is_current=false`, bypassing the intended current-state maintenance.

Requirements:

- The maintenance trigger unconditionally establishes `NEW.is_current = true` for a new valid interval.
- `app` cannot control `is_current` on INSERT or UPDATE.
- The application role cannot directly change current-state metadata.
- The partial exclusion constraint operates on the actual current predicate.
- Valid supersession automatically makes the predecessor non-current and successor current.
- Failed supersession leaves the original state unchanged.

Add direct SQL tests under `app`.

Cover explicit false insertion, overlapping non-current inserts and attempted metadata updates.

## C2. F-PG-TEMP — Harden the SECURITY DEFINER function

Current defect:

The trigger's SECURITY DEFINER function can resolve `attendance_interval` through `pg_temp` even though the function specifies `search_path=public`.

Requirements:

1. Schema-qualify all protected table references, for example `public.attendance_interval`.
2. Do not rely exclusively on `SET search_path=public`.
3. Use a deliberately restricted and documented function search path.
4. Revoke unnecessary CREATE/TEMP privileges from `app` and PUBLIC.
5. Retain required migration ownership.
6. Verify actual function ownership.
7. Verify execute privileges.
8. Ensure callers cannot create a temporary decoy table that redirects maintenance operations.
9. Preserve the two-role architecture.

Test the original temporary-table exploit as `app`.

The protected public-table state must remain correct even when a caller attempts the exploit.

Do not introduce RLS or another runtime database role.

## C3. F-XEMP-SUPERSEDE and F-INTERVAL-SELFLOOP

Enforce supersession identity:

- Same employer.
- Same employee.
- Valid predecessor identity.
- No self-supersession.
- No invalid lineage cycle.
- A successor cannot invalidate another employee's interval.
- One successor per original interval.

Enforce this at database level.

Do not rely exclusively on Go service validation.

Test direct SQL attempts as `app`.

## C4. F-SINGLECHAIN-UPDATE and F-CHAIN-SCOPE

The single-chain and scope guards must protect UPDATE as well as INSERT.

An `app` UPDATE must not:

- Repoint a timesheet interval to another chain.
- Create another successor.
- Cross employer boundaries.
- Cross employee boundaries.
- Cross pay-period boundaries.
- Create self-links or invalid cycles.

Add BEFORE INSERT OR UPDATE protection where appropriate.

Use composite foreign keys and explicit constraints wherever possible.

Ensure `EffectiveApproved` refuses ambiguous effective versions rather than choosing an arbitrary result using `LIMIT 1`.

## C5. F-LINE-MOVE — Timesheet line identity

A linked timesheet line must not be moved to another employee or pay period while retaining its original attendance links.

Prefer prohibiting changes to the parent identity once interval links exist.

If a move is necessary for a legitimate workflow, require complete revalidation of:

- Employer.
- Employee.
- Pay period.
- Linked interval ownership.
- Interval allocation uniqueness.

Add direct `app` SQL regression tests.

## C6. F-SHEET-KEYS — Timesheet identity immutability

Once a timesheet has linked lines or interval evidence, its employer, employee and pay-period identity must not be changed.

A status update to approved must not be combined with an identity transplant.

The identity protection must apply whether the row is draft, submitted, approved or part of a correction chain.

Test direct UPDATE operations as `app`.

## C7. F-ROSTER-REOPEN — Published-roster immutability

Current defect:

A published roster can be reverted to draft via direct SQL, permitting edits to historically published assignments.

Requirements:

- Publication writes immutable publication metadata.
- A published revision cannot be reopened or silently demoted.
- Published assignments cannot be edited.
- The current revision pointer cannot be moved backwards to re-enable mutation.
- Revision uses a new draft revision.
- Publication of that revision applies the normal validation gate.
- The original version remains historically queryable.

Add appropriate parent-state guards and database triggers.

Test unpublish, assignment edits, pointer rollback and cross-revision reparenting as `app`.

## C8. F-HR-DIRECT-WRITE — Protect employment history

Direct application-role SQL must not silently modify:

- Closed employment terms.
- Closed compensation.
- Historical employee status.
- Published rosters.
- Historical team or skill relationships where those relationships affect approved work.

Do not simply revoke all UPDATE privileges if valid workflows still need to close an open effective-dated record.

Design narrowly scoped controlled mutations.

Acceptable approaches include:

- Restricted owner-executed database functions with specific parameters and validations.
- Strong database triggers that permit only narrowly defined successor-closing operations.
- Explicit revokes plus controlled application-service operations.

Do not use mutable caller-controlled session variables as the sole authorization mechanism.

The final design must ensure:

- Historical salary and status cannot be rewritten.
- Valid new effective-dated records can still be created.
- Prior intervals close consistently.
- No overlapping history is introduced.
- Approved-period changes produce an explicit conflict.
- Status transition rules cannot be bypassed through direct SQL.

Test each protected entity using the actual `app` role.

## C9. Database privilege regression suite

For every new trigger, constraint or function:

- Verify function owner.
- Verify SECURITY DEFINER status where applicable.
- Verify `search_path`.
- Verify schema-qualified object access.
- Verify privileges.
- Execute positive and negative SQL tests.
- Confirm the result by reading persisted state.
- Confirm failed changes leave the database unchanged.

A test that merely receives an error is insufficient if the transaction partially modified related rows.

---

# 6. Stage D — Repair first installation and employer onboarding

## D1. F-INITROLES — Repair the actual role-provisioning script

The current script uses psql variable substitution inside a dollar-quoted DO block, which fails.

Correct `scripts/db/init-roles.sh`.

Requirements:

- Execute the actual script against a disposable PostgreSQL 16 server.
- Provision `migration_admin` and `app`.
- Apply the intended grants and revokes.
- Verify object ownership and privileges.
- Do not silently skip existing-role checks.
- Do not hardcode real passwords in the repository.
- Use safe psql variable handling.
- Fail loudly on incomplete provisioning.

The test must invoke the script itself, not a separately reproduced SQL bootstrap.

## D2. F-SUPERUSER-DEFAULT — Correct local runtime configuration

The default documented workflow must not run the application as a superuser.

Update:

- `.env.example`
- `Taskfile.yml`
- README and fresh-install instructions.
- Migration command documentation.
- Runtime startup checks.

Use:

- `DATABASE_URL` for runtime role `app`.
- `MIGRATION_DATABASE_URL` for `migration_admin`.

Requirements:

- `serve`, normal CLI commands and seed operations use `app`.
- Migrations use `migration_admin`.
- Startup refuses or clearly rejects a superuser runtime connection.
- Migration commands fail loudly when connected through the wrong role.
- The environment file documents both connections separately.
- Tests verify actual `current_user`.

Do not retain `DATABASE_URL` pointing to a superuser while claiming runtime privilege separation.

## D3. P1-23 — SQL-free employer onboarding

Provide the minimum employer creation operation.

The v0.1 system remains single-employer operationally.

The bootstrap workflow should support:

```
workforceops employer create --file employer.json
workforceops employer show
```

And a corresponding REST endpoint:

`POST /api/v1/employers`

Requirements:

- Employer creation is available on an empty fresh database.
- Enforce the v0.1 single-employer operational restriction.
- Preserve multi-employer-ready schema constraints.
- Validate the employer code, name, UEN and jurisdiction.
- Require operator attribution.
- Write an audit record.
- Make repeat attempts safe and explicit.
- Do not require raw SQL or `seed demo` merely to create an employer.

Employer creation is a bootstrap operation and must not require an existing employer context.

---

# 7. Stage E — Additional directly relevant corrections

Do not fix all remaining P2/P3 findings.

Only address the following where they directly affect the eight primary defects or the first payroll scenario.

## E1. F-REST-DAY

Replace free-text rest-day values with a canonical, validated representation.

Apply Go validation and a database CHECK or equivalent enforcement.

Unsupported values must not silently bypass the relevant roster checks.

## E2. F-MOVE-UNROSTERED

When a correction moves work to a date without an assignment:

- Do not retain the old date's roster assignment.
- Resolve the assignment for the new date.
- If no valid assignment exists, use an explicit no-roster reference and create an exception.
- Do not invent a new programme or roster assignment.

## E3. F-TIME-ROUND

The current time-duration representation converts float hours into NUMERIC fields with implicit database rounding.

Define an explicit duration-precision policy.

Prefer integer minutes or seconds for measured durations.

At minimum:

- No implicit database rounding may silently alter a business duration.
- Go must validate and quantize duration values explicitly.
- The same policy must be used in attendance, timesheets and payroll inputs.
- Add fractional-duration regression tests.

Do not confuse time-duration precision with monetary precision.

## E4. F-AUDIT-GAPS

Verify that operator attribution is actually persisted, not merely accepted and discarded.

At minimum, audit consequential mutations involving:

- Employee creation.
- Employment-term changes.
- Employee status changes.
- Team membership changes.
- Skill assignment changes.
- Roster publication and revision.
- Attendance event conflicts and corrections.
- Timesheet transitions.
- Payroll exclusions.

Each audit entry should identify the actor, action, entity, timestamp and relevant before/after references or reason.

Audit insertion must be atomic with the business operation.

## E5. Reset safety

Inspect `seed demo --reset`.

Keep it restricted to explicitly disposable databases.

Strengthen the target check so that a typo cannot casually recycle an unrelated developer database.

Test:

- Protected database name.
- Unknown database name.
- Wrong server.
- Missing or invalid migration DSN.
- Password-authenticated server.
- Normal successful disposable reset.

No broad DELETE permissions or trigger-bypass mechanisms.

---

# 8. Explicitly deferred findings

Do not expand this milestone to all 72 findings.

The following remain deferred unless a demonstrated dependency makes one a necessary part of the accepted fixes:

- `F-SHG-NOCLOSE` — complete SHG supersession workflow, addressed with the first SHG calculation package.
- `F-SHG-VERIFY` — verified-none evidence schema, addressed with the first applicable SHG rule.
- `F-RUN-LIFECYCLE` — payroll-run validate/approve state machine, addressed in the payroll calculation milestone.
- Rotation materialization.
- Advanced roster optimization.
- Generic `Idempotency-Key` framework.
- General REST/CLI read parity for optional surfaces.
- Full list pagination.
- SSO.
- Multi-employer authentication.
- PostgreSQL RLS.
- Advanced integration and simulation frameworks.
- Production cloud deployment.
- Docker-specific refinements that cannot be validated without a daemon.

Record each deferred item explicitly.

A deferred item must not be represented as an implemented capability.

---

# 9. Testing and acceptance requirements

Use real PostgreSQL 16.

Every destructive test must use an explicitly verified disposable database.

## A. Eight primary service-path regressions

Implement and run tests for:

1. Duplicate vendor feed → explicit disposition → timesheet generation succeeds without double-counting.
2. Midnight-spanning break → correct dated break segments → successful timesheet generation.
3. Approved timesheet → reissue → attendance correction → regenerated correction approval.
4. Post-approval retroactive compensation → rejected or explicitly blocked conflict.
5. Identical attendance evidence under `TZ=UTC` and `TZ=Asia/Singapore` → identical classification.
6. Cross-midnight correction → correctly split dated intervals and roster attribution.
7. Late break after approval → no silent mutation → controlled correction workflow.
8. Terminated/suspended employee → cannot receive a publishable new shift.

## B. Database application-role matrix

Run as `app`.

Test:

- Explicit `is_current=false` insert.
- Direct `is_current` update.
- Temporary-table shadowing.
- Cross-employee supersession.
- Cross-employer supersession.
- Self-supersession.
- Two-node cycle.
- Supersession across pay periods.
- Single-chain UPDATE re-pointing.
- Timesheet line movement.
- Timesheet identity mutation.
- Published-roster reopening.
- Published-assignment mutation.
- Closed employment-term mutation.
- Employee status resurrection.
- Compensation mutation affecting an approved period.

For every test, inspect persisted state after success or failure.

## C. Fresh installation and onboarding

Execute:

1. Start disposable PostgreSQL.
2. Run `scripts/db/init-roles.sh`.
3. Verify roles and privileges.
4. Run migrations using `migration_admin`.
5. Verify database ownership and function privileges.
6. Start application using `app`.
7. Create the employer using CLI.
8. Create the employer through REST on a second fresh database.
9. Complete the HR/WFM/attendance/timesheet workflow.
10. Verify the application cannot operate as a superuser by default.

## D. Concurrency and transaction tests

Test actual concurrent database transactions for:

- Attendance correction.
- Late break attachment.
- Duplicate-event import.
- Timesheet reissue.
- Correction approval.
- Employment-term successor.
- Employee status transition.
- Roster publication.
- Roster revision.

Verify one valid winner where required.

Simulate relevant failures between evidence insertion, pairing, interval creation and aggregate updates.

Replaying the same import or operation must recover incomplete work without losing or duplicating evidence.

## E. Full regression suite

Run:

- `task lint`
- `task build`
- `go test -count=1 ./...`
- Race tests for attendance, timesheet, payroll, HR and WFM.
- Database integrity suite.
- Fresh-install test.
- Role privilege test.
- CLI E2E.
- REST E2E.
- Persisted-state CLI/REST comparison.
- OpenAPI validation.
- `git diff --check`.

Database suites must not silently skip.

Do not use a superuser to claim application-role test success.

Do not silently switch to serial execution after a parallel test failure.

If the environment prevents Docker or Compose execution, report the exact limitation and verify the affected underlying build/provision steps independently.

---

# 10. Overall acceptance gate

Milestone 1.2 can PASS only when:

1. All eight service-reachable P1 defects are fixed.
2. No service operation can silently change approved timesheet inputs.
3. Late break evidence cannot silently invalidate an approved timesheet.
4. Cross-midnight attendance and corrections produce correct business-date attribution.
5. Host timezone cannot alter attendance classification.
6. Verified working time is never truncated by the 72-hour compliance limit.
7. Terminated or suspended employees cannot receive publishable new shifts.
8. All listed `app`-role lineage and immutability bypasses are closed.
9. SECURITY DEFINER maintenance cannot be redirected through `pg_temp`.
10. Runtime and migration role separation works in the documented fresh-install path.
11. Employer onboarding works through CLI and REST without SQL.
12. Consequential audit events are atomic with their business mutations.
13. Existing employer isolation and correction-chain concurrency guarantees remain intact.
14. The complete supported workflow works through CLI and REST.
15. No unverified statutory calculation or fabricated golden fixture has been introduced.
16. `SG-MOM-OT-001` remains BLOCKED.

If any required condition fails, report FAIL with the precise remaining blocker.

Do not declare PASS based solely on green unit tests.

---

# 11. Deliverables

Produce:

1. `plans/milestone1_2_remediation_report.md`
2. Any updated structural audit reference necessary to explain newly discovered architecture issues.
3. New forward migration(s), starting at `00010`.
4. Regression tests for every corrected finding.
5. Updated CLI/API/OpenAPI contracts where operations or errors change.
6. Updated setup documentation for the two database roles and employer onboarding.

The remediation report must contain:

- Initial Git state and baseline.
- Confirmed findings and their disposition.
- Structural/architectural observations.
- Files and migrations modified.
- Database constraints and triggers changed.
- Transaction and audit changes.
- Tests actually executed.
- Application-role SQL evidence.
- Fresh-install and onboarding evidence.
- CLI/REST parity evidence.
- Remaining deferred issues.
- Unverified environmental limitations.
- Final PASS/FAIL decision.

Keep all changes uncommitted.

---

# 12. Stop condition and next milestone

Stop once the Milestone 1.2 acceptance gate passes.

Do not continue into payroll calculations.

The next milestone, after a verified PASS, will be the first restricted Singapore payroll calculation:

- Monthly-rated Singapore Citizen.
- Full salary period.
- No overtime.
- No unpaid leave.
- No variable earnings.
- Verified CPF, SDL and applicable SHG inputs.
- Independent golden and boundary fixtures.
- Reproducible payroll output from immutable inputs and pinned rule versions.

That work must follow the existing Phys-Research and Phys-Pseudo methodology.

**The objective of Milestone 1.2 is to make the operational evidence trustworthy enough that the first payroll calculation can be trusted. Preserve the architecture, fix only what the evidence justifies, and do not extend the scope beyond this integrity gate.**

## My execution priority

One additional point: I would have the agent tackle the work in this order:

1. Attendance correctness first: overlap disposition, midnight breaks, timezone classification, cross-midnight corrections and late-break handling.
2. Historical integrity second: approved-timesheet correction chains, compensation changes and employee status.
3. Database protection third: `is_current`, `SECURITY DEFINER`, lineage, parent identity and immutability.
4. Installation and onboarding fourth: role provisioning, runtime DSN and employer creation.
5. Full regression and parity last.

This order addresses the actual business evidence before the surrounding operational plumbing.

The key completion test is simple: no clock evidence should be lost, double-counted or silently reinterpreted; no approved record should silently change; and the application must enforce those guarantees even when operating as its ordinary PostgreSQL `app` role.

Once that holds, we should stop reviewing infrastructure and move to the first verified payroll calculation.