I recommend one final bounded milestone: Milestone 1.3 — Final Payroll-Input Integrity Closure.

This should resolve the remaining nine critical database-integrity gaps, the compensation/approval race, correction-group lifecycle problems and clean-server test reliability. Importantly, we should not spend another cycle redesigning the application.

The most important change to the previous plan is to treat database enforcement, transaction ordering and historical business semantics as one problem. Fixing only the Go services is insufficient.

## Implementation-agent prompt

# Prompt 13 — WorkforceOps Milestone 1.3

# Final Payroll-Input Integrity Closure

## 1. Objective

Implement the smallest justified remediation identified by:

`plans/milestone1_2_final_adversarial_review.md`

The review found no remaining P0 findings, but identified nine material database-integrity gaps, one service-reachable partial-correction defect, compensation lifecycle gaps, and a transaction-ordering race that can silently change approved payroll inputs.

The objective is to close these defects and establish a trustworthy database-enforced evidence chain before the first independently verified Singapore payroll calculation.

The required business chain remains:

`Employer → Employee → Employment Terms → Compensation → Roster → Clock Events → Attendance → Correction Groups → Approved Timesheet → Payroll Readiness`

This is the final targeted integrity-remediation milestone, not another architectural redesign.

**Do not implement statutory payroll calculations in this task.**

## 2. Expected results

The following must be true at completion:

1. A correction group cannot be detached from its predecessor or contain invalid segments.
2. A multi-segment correction cannot be partially superseded.
3. A duplicate-event disposition cannot form a cycle, including through concurrent direct SQL.
4. Cross-midnight business-date attribution cannot be manipulated by direct SQL.
5. Backdated clock or attendance evidence cannot bypass payroll-readiness conflict detection.
6. Compensation and payroll approval cannot race to produce a stale approved basis.
7. Compensation cannot be inserted outside employment terms or overlap ambiguously across terms.
8. PostgreSQL cannot silently round invalid monetary precision.
9. Populated migrations cannot silently destroy correction or disposition history.
10. Clean-server database testing is reliable.
11. The application-role privilege matrix detects new unprotected tables.
12. Compensation has a controlled, auditable correction lifecycle.
13. All previously verified M1.2 invariants remain intact.
14. `SG-MOM-OT-001` remains BLOCKED.
15. No statutory financial calculations are introduced.

---

# 3. Repository preservation and baseline

Before modifying anything:

- Record `git status --short`.
- Record `git rev-parse HEAD`.
- Confirm the baseline remains `f43515883668c917edac31be49a7cf8334eca5c2`.
- Inventory tracked and untracked files.
- Read the complete Prompt 12 adversarial report.
- Read `plans/milestone1_2_remediation_report.md`.
- Inspect migrations `00001` through `00010`.
- Inspect the current attendance correction-group, duplicate disposition and compensation implementations.
- Record active database connection targets.

Preserve all existing uncommitted work.

Do not:

- Reset or clean the repository.
- Stage or commit.
- Discard previous implementation work.
- Modify `bchat/`.
- Modify upstream Phys-Research or Phys-Pseudo.
- Alter migrations `00001`–`00010`.
- Drop or recycle any existing developer database.
- Deploy to Neon or Fly.io.
- Implement payroll calculations.

Use disposable PostgreSQL 16 databases for every destructive or adversarial test.

All changes remain uncommitted.

---

# 4. Stage A — Bind correction groups to their evidence

Address A1-F1, A1-F2 and A1-F4.

Inspect:

- `attendance_correction_group`.
- `attendance_correction`.
- `attendance_interval`.
- `00010_m12_integrity.sql`.
- `internal/attendance/attendance.go`.
- `internal/attendance/review.go`.
- Correction-group and direct-SQL tests.

## A1. Correction-group predecessor ownership

The group must be bound to exactly one valid predecessor interval.

At database level enforce:

- Same employer.
- Same employee.
- Predecessor exists.
- Predecessor is current when the group is created.
- One correction group per predecessor.
- No self-reference.
- No invalid supersession cycle.
- No group with zero segments.
- No group appended to after its lifecycle is closed.

Do not rely exclusively on the `Correct()` service.

Preferred enforcement:

- Composite foreign keys where supported by the schema.
- A narrowly scoped trigger validating predecessor ownership and current state.
- Appropriate uniqueness and CHECK constraints.

The database must reject a group that refers to another employee's or employer's interval.

A malformed group must not consume the legitimate predecessor's unique successor slot and permanently prevent correction.

## A2. Correction-segment binding

Every attendance segment produced by a correction group must satisfy:

`segment.supersedes_id = group.predecessor_interval_id`

The relationship must also prove:

- Same employer.
- Same employee.
- Valid segment business date.
- Correct correction-group identity.
- Original predecessor remains traceable.
- Segment belongs to the same atomic correction operation.

A segment with NULL or mismatched predecessor identity must be rejected.

A closed correction group must not accept new segments.

Add direct `app` SQL tests for:

- NULL `supersedes_id`.
- Mismatched predecessor.
- Cross-employer segment.
- Cross-employee segment.
- Segment added to closed group.
- Orphan correction group.
- Duplicate segment.
- Concurrent attempts to create groups for one predecessor.

After every rejected operation, read back the database state and prove no predecessor or current interval was changed.

## A3. Local business-date validation

Address A1-F4.

The database must prevent an attendance segment from claiming a business date inconsistent with its actual interval.

Use the employee's employment timezone, not:

- PostgreSQL session timezone.
- Host timezone.
- Vendor/event timezone.
- Client/programme timezone.

For each correction segment, validate the employment-local date relationship between `interval_start`, `interval_end` and `local_date`.

The canonical correction logic splits intervals by employment-local midnight. Database validation must enforce the same invariant for direct SQL writes.

Test:

- Correct date.
- Previous date.
- Future date.
- Midnight boundary.
- Cross-midnight source interval represented by incorrectly unsplit segment.
- Different host timezone.
- Different PostgreSQL session timezone.

Do not invent a different attribution policy inside the trigger.

---

# 5. Stage B — Make multi-segment re-correction group-atomic

Address A1-F3.

Current problem:

An operator can correct only one segment of an existing multi-segment correction group, creating a second group for that segment while the sibling segment remains active.

This silently breaks the original correction's business meaning.

Implement group-atomic re-correction.

## Required behavior

1. Every interval returned by `ListIntervals` exposes its `correction_group_id` where applicable.
2. Re-correcting a member of a multi-segment group resolves all current siblings.
3. The operator corrects the group as one logical unit.
4. Every required sibling is included in the re-correction operation.
5. The new correction group becomes the unique successor of the previous group.
6. Old group members remain historical and non-current.
7. New date segments are generated consistently.
8. All segments, day records, roster links and audit events commit atomically.
9. A partial correction request is rejected with a meaningful machine-readable error.
10. Concurrent re-correction attempts produce exactly one valid successor group.

Do not solve this merely by displaying a warning while still allowing partial supersession.

Add tests:

- Re-correct one segment of a two-segment group: rejected.
- Re-correct the complete group: succeeds.
- Re-correct a same-date group.
- Re-correct a cross-midnight group.
- Re-correct across a pay-period boundary.
- Concurrent whole-group re-correction.
- Failure during segment regeneration: zero partial successor groups.
- Subsequent timesheet generation uses only the effective group.
- Original attendance and correction histories remain unchanged.

Include the same-period two-segment lifecycle test:

`Generate → Submit → Approve → Reissue/Correct → Generate → Submit → Approve`

Verify one coherent timesheet and correct total duration.

---

# 6. Stage C — Prevent duplicate-disposition cycles

Address B-CYC-01.

The current trigger checks whether an event is already dispositioned but does not serialize competing decisions involving shared canonical events.

Two concurrent direct-SQL transactions can attempt:

`A → B`

and

`B → A`

and potentially both pass their respective checks.

## Required policy

Define a simple canonical-event policy:

- A canonical event may have multiple duplicates.
- An event already classified as a duplicate cannot become the canonical event of another disposition.
- A canonical event cannot itself be classified as a duplicate.
- Cycles are prohibited.
- Chains are prohibited.
- The original raw event is never modified.
- Dispositions remain explicit and auditable.

If existing code currently permits A→B→C, correct the behavior and update comments and tests to reflect the canonical policy.

Do not rely on a check-then-insert trigger without concurrency protection.

## Concurrency protection

Use a deterministic lock strategy.

Acceptable approaches include:

- Transaction-scoped advisory locking using a stable, deterministically ordered pair key.
- Locking the involved clock-event rows in deterministic ID order.

Do not introduce nondeterministic lock ordering that can deadlock under reversed pairs.

After acquiring locks:

1. Recheck the disposition state.
2. Confirm neither endpoint violates the canonical-event policy.
3. Insert the disposition.
4. Persist the audit record.
5. Commit atomically.

Add a real concurrent direct-SQL `app` test for A→B versus B→A.

Prove that:

- At most one valid disposition is committed.
- No cycle exists.
- One canonical clock pair remains usable.
- No event becomes permanently hidden.
- The rejected transaction leaves no disposition or audit residue.

Also test multiple legitimate duplicate events against the same canonical event.

---

# 7. Stage D — Prevent backdated attendance evidence

Address A1-F5.

Current problem:

Payroll readiness uses `created_at > approved_at` to identify new attendance evidence.

However, clock events and attendance intervals allow caller-controlled `created_at`.

A direct SQL insert can supply a timestamp earlier than the approval time and evade the conflict check.

## Required correction

Add a forward-migration database trigger that normalizes or protects `created_at` for:

- `clock_event`.
- `attendance_interval`.
- Other attendance evidence records used by readiness.

Application callers must not be able to create apparently historical records by supplying a backdated creation timestamp.

Preserve:

- Actual event timestamp.
- Source timestamp and offset.
- Employment business date.
- Server-side record-creation metadata.

Do not confuse event time with record-creation time.

Test:

1. Create and approve a timesheet.
2. Insert a new clock event directly as `app`.
3. Supply a `created_at` earlier than the approval.
4. Confirm the stored creation metadata cannot be backdated.
5. Confirm readiness identifies the new evidence as an unresolved conflict.
6. Repeat with an attendance interval.
7. Test a correction-group segment.
8. Verify that normal event-time backdating for a legitimate late-arriving source event remains representable without falsifying record-creation time.

**Important:** this fix does not resolve D-RACE. Transaction serialization is a separate required fix.

---

# 8. Stage E — Compensation and payroll-approval consistency

This is the most important part of the milestone.

Address:

- D-RACE.
- D-CONTAIN.
- D-ROUND.
- D-AUDIT.
- D-NOCORRECT.

These findings affect the correctness of the financial basis before any statutory calculation occurs.

## E1. D-RACE — Solve the compensation/approval race

Current problem:

Compensation insertion and payroll approval use PostgreSQL `now()`, which represents transaction-start time.

A compensation transaction can begin before approval but commit afterwards. Its `created_at` may therefore appear earlier than `approved_at`, allowing readiness to miss the new compensation.

Do not attempt to solve this using `clock_timestamp()` alone.

That only reduces the timing window; it does not establish commit ordering.

### Required design

Implement a shared transaction-ordering barrier.

Every operation that can change the compensation basis and every operation that finalizes the corresponding approved payroll input must serialize against the same employee-level or otherwise correctly scoped lock.

At minimum include:

- Compensation insertion.
- Compensation correction/supersession.
- Timesheet approval.
- Any payroll approval/finalization path that freezes the employee's input basis.
- Any other supported operation that can change the basis after approval.

Preferred approach:

- A transaction-scoped employee or employment lock.
- Deterministic lock acquisition order for multiple employees.
- Acquire the lock before validating the relevant historical period.
- Re-read the compensation and approval state after acquiring the lock.
- Perform the business mutation or approval within the same transaction.
- Release the lock only at transaction completion.

The important ordering property is:

**If compensation wins the lock, approval must observe it. If approval wins the lock, compensation must observe the approval and create an explicit conflict or be rejected.**

Do not base this guarantee solely on wall-clock timestamps.

### Application-role protection

The same protection must apply to direct SQL INSERT paths available to `app`.

Either:

- Restrict direct insertion and route writes through a controlled database function; or
- Use a trigger that acquires and validates the same lock and rules.

The approval path must use the equivalent lock, so direct SQL cannot evade the barrier.

Do not create different locking conventions in Go and PostgreSQL.

### Required live race test

Use two separate PostgreSQL connections.

Transaction A:

1. Begin compensation insertion.
2. Acquire the employee lock.
3. Pause before commit.

Transaction B:

1. Attempt the relevant approval.
2. Verify it blocks on the same lock.

Then test the reverse order:

1. Approval acquires the lock and commits.
2. Compensation attempts the same operation.
3. It must detect the approved basis and be rejected or create an explicit correction conflict.

Verify actual persisted values and readiness results.

Do not accept a test using only mocked transactions or manually assigned timestamps.

## E2. D-CONTAIN — Enforce compensation within employment terms

Current problem:

Compensation rows can be inserted outside the parent employment-term period, and overlapping same-code compensation across different terms can create an ambiguous basis.

Service validation is insufficient.

Add database-level enforcement.

Requirements:

- Compensation belongs to the same employer and employee as its employment terms.
- Effective dates fall within the parent employment-term interval.
- Compensation cannot outlive a closed parent term.
- Overlapping compensation with the same employee and compensation code is prohibited across terms.
- Future-effective compensation is represented accurately.
- Superseded compensation remains historical.
- Changes are protected against concurrent conflicting insertions.

Use the appropriate combination of:

- Composite foreign keys.
- CHECK constraints.
- PostgreSQL range/exclusion constraints where feasible.
- Trigger validation with employee-level transaction locking.

Do not enforce overlap only within one `terms_id`. The invariant applies to the employee and compensation code across applicable employment terms.

Test direct SQL as `app`:

- Compensation before employment starts.
- Compensation after employment ends.
- Compensation outliving a closed parent term.
- Overlapping same-code rows under different terms.
- Concurrent overlapping inserts.
- Valid non-overlapping historical and future compensation.

## E3. D-ROUND — Prevent silent monetary rounding

Current problem:

A `NUMERIC(12,2)` destination can round `199.999` to `200.00` before a CHECK constraint examines the stored value.

A service-side validation does not protect direct SQL.

### Required correction

For monetary fields requiring exactly two decimal places:

- Use unconstrained `NUMERIC` or an equivalent exact representation.
- Add an exact-scale database constraint, such as equality with `trunc(value, 2)`.
- Retain suitable magnitude and range restrictions.
- Reject rather than silently round excess precision.

Apply consistently to at least:

- `compensation_line.amount_monthly`.
- `employment_terms.monthly_basic`.
- Any equivalent monetary source fields that feed payroll readiness.

Do not use `float64`.

Test:

- `199.99` accepted.
- `199.90` accepted.
- `199.999` rejected.
- Negative values rejected where prohibited.
- Excessive magnitude rejected.
- Go and direct SQL produce the same acceptance result.
- Persisted database values match the exact input.

Do not introduce statutory monetary rounding.

## E4. D-AUDIT — Audit all permitted compensation writes

Current problem:

Direct SQL compensation insertion can bypass the service audit.

Make audit persistence mandatory for every permitted compensation mutation.

Acceptable approaches:

- A narrowly scoped database trigger that writes an audit event.
- A controlled SECURITY DEFINER function that writes the compensation and audit event atomically.

Avoid duplicate audit rows when the Go service also records the event.

Requirements:

- One canonical audit event per committed business operation.
- Actor and reason.
- Employee and employer.
- Compensation identity and effective dates.
- Exact persisted monetary value.
- Link to any superseded compensation.
- Audit and business record commit atomically.

A failed transaction must not leave an audit event for a nonexistent compensation row.

Canonicalize monetary audit values consistently, for example `100.00`, not `100`.

## E5. D-NOCORRECT — Add controlled compensation correction

The existing insert-only model protects history but provides no legitimate correction path for erroneous compensation.

Implement a minimal controlled successor lifecycle.

The original compensation record remains immutable.

A correction:

1. Identifies the original compensation record.
2. Creates an auditable successor or controlled replacement.
3. Preserves employer, employee and compensation classification.
4. Enforces one successor per original.
5. Validates effective periods.
6. Preserves the original value and history.
7. Updates readiness with an explicit correction state.
8. Does not rewrite an already approved timesheet or payroll input silently.

A correction affecting an approved period must be visibly blocked or require a controlled downstream correction.

Do not provide an administrator bypass that silently edits the original record.

Add tests:

- Correct an erroneous draft-period compensation record.
- Correct a future-effective record.
- Attempt a correction affecting an approved period.
- Attempt two concurrent successors.
- Attempt a cross-employee successor.
- Verify audit chain and exact monetary values.
- Verify readiness before and after correction approval/disposition.

---

# 9. Stage F — Protect populated migration history

The Prompt 12 review found that `00010` Down can lose duplicate-disposition decisions and may make correction-history restoration impossible.

Do not alter migrations `00001`–`00010`.

The product currently has no supported migration Down command, so do not build a generic rollback subsystem.

Adopt an explicit migration policy:

**M1.2 and later integrity migrations are forward-only once dependent business data exists.**

Requirements:

1. Document this policy in the migration/database documentation.
2. Ensure application and Taskfile workflows expose no casual Down operation.
3. Do not claim populated Down+Up safety.
4. Preserve original events, disposition decisions, correction groups and audit history.
5. Any attempted manual rollback that would destroy these records must be treated as unsupported.
6. Require a verified database backup and controlled restore/rebuild procedure before destructive recovery.
7. New migration `00011` must have its own safe rollback behavior.

For `00011` Down:

- Inspect the dependent data.
- Fail before destructive changes when records would be lost or relationships invalidated.
- Do not silently drop correction, disposition or compensation-successor history.

A documented forward-only policy is preferable to pretending populated rollback is supported.

---

# 10. Stage G — Fix clean-server test reliability

Address `S7-P2-FRESHFLAKE`.

The previous review found that the test suite can fail on a clean PostgreSQL server because a test connects to a database before it has been created and because role creation races.

Correct test setup.

Requirements:

- Create the required base test database before connecting.
- Synchronize global role provisioning.
- Tolerate only the expected duplicate-role creation race, after verifying the role's actual attributes and privileges.
- Use unique disposable databases per package/test where necessary.
- Never reuse a stale migration state unknowingly.
- Ensure the first execution against a clean server passes without retry.
- Retain application-role testing.
- Never substitute superuser success for app-role tests.

Run the suite on a genuinely clean PostgreSQL instance, not merely a warm shared test server.

## Fail-closed privilege matrix

Address `S7-P2-MATRIX-OPEN`.

The existing allowlist-only test can remain green even when a future business table inherits full DML grants.

Change it to enumerate all relevant application schemas/tables using PostgreSQL catalog queries.

Every table must have an explicit expected privilege classification.

A newly introduced table without an entry must fail the test.

Verify:

- SELECT.
- INSERT.
- UPDATE.
- DELETE.
- Trigger-owned fields.
- Protected historical tables.
- Controlled function execution.
- Default privileges.
- PUBLIC privileges.

Distinguish intentionally mutable operational tables from immutable history tables.

The objective is a fail-closed privilege test, not blanket denial of every write.

---

# 11. Stage H — Small correctness and contract cleanup

Address only the remaining findings that directly affect the above work.

## H1. B-CHAIN-01

Update the disposition trigger comments and tests to reflect the actual canonical-event policy.

The implementation must enforce the documented policy, not merely describe an intended one.

## H2. C-TZ-01 and C-TZ-02

Normalize diagnostic and audit timestamp rendering so the same instant produces the same stable representation across host timezones.

Prefer UTC for machine-stable audit values, while retaining employee-local business dates as explicit fields.

Do not change correct business-date calculations.

## H3. C-TZ-03 and C-TZ-04

Correct the API/CLI contract and comments so they state clearly:

- Employment timezone determines attendance business-date attribution.
- Vendor timezone is source provenance.
- Client/programme timezone governs service operations.
- Host timezone never determines employment results.

## H4. C-TZ-05

Validate IANA programme timezone values at configuration boundaries if the application currently stores them as accepted operational configuration.

Do not introduce a complex multi-jurisdiction engine.

## H5. B-DAY-01

Make exception status and timesheet-generation blocking behavior consistent.

An unresolved overlap must not be displayed as fully reviewed if it still blocks the employee.

A confirmed, properly dispositioned duplicate must no longer block generation.

## H6. S7-P3-LEGACY-EXEC

Revoke unnecessary direct EXECUTE privileges from the runtime app role on legacy trigger-only functions.

Extend the catalog privilege test accordingly.

Do not revoke execution required for legitimate trigger or controlled-function operation.

## H7. Counts and documentation

Correct the remediation and test count discrepancies.

Keep test claims tied to actual commands and outputs.

Clearly distinguish:

- Passed tests.
- Tests with no test files.
- Intentionally deferred tests.
- Environment-dependent unverified checks.

---

# 12. Scope boundaries and explicit deferrals

Do not expand Milestone 1.3 into all remaining Prompt 12 findings.

The following remain deferred unless directly required by a correction:

- Full SHG lifecycle and verified-none evidence.
- Complete payroll-run validate/approve state machine.
- Rotation materialization.
- Generic idempotency framework.
- General list pagination.
- Full CLI/REST read parity for optional features.
- Enterprise SSO.
- Multi-employer authentication.
- PostgreSQL RLS.
- Microservices.
- Distributed event processing.
- UI.
- Advanced roster optimization.
- Production cloud deployment.
- General performance/load benchmarking.

Record deferred items and their dependencies honestly.

---

# 13. Architecture remains frozen

Preserve:

- One Go executable.
- One PostgreSQL database.
- One shared application-service layer.
- CLI and REST.
- pgx and explicit SQL.
- Singapore-first jurisdiction strategy.
- Employer-scoped schema.
- Separate migration and runtime roles.
- Phys-Research and Phys-Pseudo methodology.

No generic legal rules engine.

No Phys-Pseudo parser or compiler.

No new country package.

No architecture redesign.

If a proposed fix requires changing a frozen decision, explain the exact issue and request approval rather than implementing the change silently.

---

# 14. Verification and acceptance tests

Use real PostgreSQL 16.

All direct SQL adversarial tests must use the actual `app` role.

## A. Correction-group tests

- Cross-employee group insertion rejected.
- Cross-employer group insertion rejected.
- Zero-segment group rejected.
- Mismatched segment predecessor rejected.
- NULL segment predecessor rejected.
- Closed-group append rejected.
- Incorrect local-date segment rejected.
- Partial group re-correction rejected.
- Whole-group re-correction succeeds.
- Concurrent correction-group creation has exactly one winner.
- Failed multi-segment correction leaves zero orphan segments/groups.
- Timesheet generation uses the current group only.

## B. Duplicate-disposition tests

- Multiple duplicates against one canonical event succeed.
- A duplicate cannot become a canonical event.
- Sequential cycles rejected.
- Concurrent A→B and B→A direct SQL attempts cannot both commit.
- No canonical event becomes unavailable.
- Original and duplicate raw evidence remain preserved.
- Exception lifecycle remains correct.
- Disposition and audit commit atomically.

## C. Compensation tests

- Direct-SQL excess decimal precision rejected.
- Direct-SQL negative values rejected where prohibited.
- Cross-terms same-code overlap rejected.
- Compensation outside parent terms rejected.
- Children cannot outlive closed parent terms.
- Corrected compensation has an immutable successor.
- Direct compensation INSERT is audited.
- Failed insertion produces no audit residue.
- Compensation and approval race serializes correctly.
- Both possible lock-winner orders produce correct outcomes.
- Readiness detects every relevant compensation conflict.

## D. Timestamp integrity tests

- Caller-supplied backdated `created_at` is rejected or normalized.
- Actual event timestamp remains preserved.
- Late-arriving source events remain representable.
- Readiness detects post-approval evidence.
- D-RACE is not incorrectly considered fixed by timestamp normalization.

## E. Privilege tests

Run all tests as `app`:

- Group ownership.
- Segment binding.
- Local-date mismatch.
- Backdated evidence.
- Disposition cycle.
- Compensation range containment.
- Compensation overlap.
- Monetary precision.
- Compensation audit.
- Compensation successor.
- Concurrent approval and compensation changes.

Assert persisted state after every rejected operation.

## F. Test harness

On a clean PostgreSQL server:

1. Provision roles.
2. Create the base test database.
3. Run all test packages in the supported parallel configuration.
4. Verify role creation and grants.
5. Verify every table has a privilege classification.
6. Repeat to verify idempotent test setup.
7. Run race detector tests.

Do not accept a warm-server-only PASS.

## G. Full regression

Run:

- `task lint`
- `task build`
- `GOWORK=off go test -count=1 ./...`
- Full suite under `TZ=UTC`.
- Full suite under `TZ=Asia/Singapore`.
- `go test -race` for attendance, timesheet, payroll, HR and WFM.
- PostgreSQL integrity suite.
- Clean-server database provisioning.
- Application-role SQL suite.
- CLI E2E.
- REST E2E.
- Persisted-state parity.
- OpenAPI validation.
- `git diff --check`.

Record sandbox or environment limitations precisely.

Do not silently use serial mode to hide a parallel test race.

Do not claim Docker build or Compose verification if Docker is unavailable.

---

# 15. Completion conditions

Milestone 1.3 PASS requires:

1. A1-F1, A1-F2, A1-F3, A1-F4 and A1-F5 closed.
2. B-CYC-01 closed with actual concurrent direct-SQL proof.
3. D-RACE closed with actual concurrent compensation/approval tests.
4. D-CONTAIN closed at database level.
5. D-ROUND closed for both compensation amounts and monthly basic salary.
6. Compensation insertion and correction are auditable and controlled.
7. Populated Down limitations are explicit and cannot silently destroy data through supported tooling.
8. Privilege-matrix testing fails closed for unclassified tables.
9. Clean-server test provisioning succeeds without retry.
10. All original M1.2 invariants remain intact.
11. The full operational workflow passes under both host timezone configurations.
12. No P0 or material P1 remains that can silently corrupt or stale approved payroll inputs.
13. `SG-MOM-OT-001` remains BLOCKED.
14. No statutory calculation has been added.

If any required condition fails, return FAIL and list the exact remaining blockers.

Do not declare PASS based only on green Go unit tests.

---

# 16. Deliverables

Create:

`plans/milestone1_3_remediation_report.md`

Include:

1. Initial Git baseline and uncommitted state.
2. Findings reconfirmed.
3. New migration and schema changes.
4. Correction-group enforcement.
5. Duplicate disposition concurrency protection.
6. Compensation and approval transaction barrier.
7. Compensation containment and precision enforcement.
8. Compensation correction and audit lifecycle.
9. Migration rollback policy.
10. Fail-closed privilege matrix.
11. Clean-server test evidence.
12. CLI/REST persisted-state parity.
13. Test commands and exact outcomes.
14. Deferred findings and reasons.
15. Environment-dependent limitations.
16. Final PASS/FAIL verdict.

Keep all changes uncommitted.

Do not modify historical migrations `00001`–`00010`.

Use one new forward migration, `00011`, unless a demonstrated technical necessity requires a second isolated migration. Explain such necessity.

---

# 17. Stop condition

Stop immediately when Milestone 1.3's acceptance criteria are satisfied.

Do not proceed into payroll calculations in this task.

The next milestone is the first independently verified Singapore monthly payroll calculation, with:

- Monthly-rated Singapore Citizen employee.
- Full salary period.
- No overtime.
- No unpaid leave.
- No variable earnings.
- Independently verified CPF, SDL and applicable SHG inputs.
- Phys-Research source provenance.
- Phys-Pseudo formal specification.
- Independent derivation.
- Golden, anti-golden and boundary fixtures.
- Reproducible payroll output from immutable inputs and pinned rule versions.

**The purpose of Milestone 1.3 is to close the remaining loopholes in the financial evidence chain, not to create another framework. Fix the identified invariants, prove them under PostgreSQL, preserve the working application and then move forward.**

## Execution priority

I would insist on this order:

1. Compensation/approval serialization first. This is the most consequential race because it can silently change the basis of an approved payroll result.
2. Correction-group database enforcement. Prevent malformed lineage and partial re-correction.
3. Compensation containment, precision and correction lifecycle.
4. Duplicate-disposition concurrency.
5. Clean-server test and privilege verification.

There is one important distinction in this prompt: it does not treat `created_at` normalization as a solution to the compensation race. Record-creation timestamps and transaction serialization solve different problems.

My target is a single definitive Milestone 1.3 PASS. After that, stop integrity remediation and build the first real, independently verified payroll calculation.