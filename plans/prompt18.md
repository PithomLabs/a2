This prompt authorizes the agent to run Gate 2: test the existing admission mechanism's remaining attack surface and build an isolated `READ COMMITTED` gate experiment. It does not authorize immediately implementing production migration `00013`.

# Implementation Agent Prompt — Gate 2: Payroll Snapshot Synchronization Redesign

## Role

Act as the senior Go/PostgreSQL architect and implementation engineer responsible for WorkforceOps Lab (`github.com/PithomLabs/a1`).

You are working on a production-oriented HR, workforce management, attendance, timesheet and payroll backend. Correctness, data provenance, transaction semantics and database-enforced invariants take priority over implementation convenience.

You have access to the repository and its PostgreSQL test environment. Inspect the actual implementation before making design assumptions.

## Mission

Execute a focused **Gate 2 architecture experiment** to determine whether the current M2.1 snapshot-admission design should be retained or replaced by:

**Option B — PostgreSQL READ COMMITTED with a transaction-scoped employer gate, enforced across all payroll-affecting write paths.**

Option B is the preferred candidate, not yet an approved production implementation.

Your work has two parts:

1. Test the precise remaining release/reacquisition attack against the existing XID-ticket and session-barrier mechanism. Do not assume that the `DISCARD ALL` observation alone disproves it.
2. Build and adversarially test an isolated Option B experiment using PostgreSQL 16.15. Determine whether it provides the required freshness and authority guarantees with materially less complexity.

Deliver an evidence-backed architecture decision record and a clear GO/NO-GO recommendation.

**Do not immediately implement production migration `00013`. Do not wire either experimental mechanism into the production capture workflow during this gate.**

---

# 1. Non-negotiable engineering principle

Apply **Invariant-First Root-Cause Remediation** to every defect, test, design decision and implementation prompt.

For each defect, answer:

1. What business invariant must always hold?
2. What actual mechanism violated it?
3. What are all possible write paths?
4. At what authoritative boundary must the invariant be enforced?
5. How is the original failure reproduced, the correction tested, and persisted database state inspected?
6. What new failure modes could the correction introduce?
7. What evidence proves the complete failure class is closed?

Mandatory rules:

- No service-only fix for a database invariant that direct SQL can violate.
- No locks without a documented purpose, identity, acquisition point and transaction boundary.
- No timestamps as substitutes for transaction ordering.
- No arbitrary retries, error suppression, constraint weakening or hidden blockers.
- No use of a passing test as proof when the actual failure mechanism has not been reproduced.
- No rewriting historical records where correction or successor records are required.
- No claiming that a failure class is closed when only one code path has been tested.
- Inspect persisted PostgreSQL state after both successful and rejected operations.
- Distinguish mechanism proofs from production-integrity acceptance.

Governing question:

> What invariant was previously unenforced, which operations could violate it, and what now makes every such violation impossible or explicitly blocked within the documented threat model?

---

# 2. Preserve the project boundaries

The following decisions remain binding:

- Go backend and PostgreSQL.
- PostgreSQL 16.15 for the experiment.
- `pgx` and explicit SQL.
- Local PostgreSQL/Docker Compose is the default test environment.
- One employer operationally in the MVP, with employer-scoped database design.
- Existing roles remain `app` and `migration_admin`. Do not introduce a third role without an explicit, evidence-backed recommendation and approval.
- No UI.
- No microservices, event bus, or unnecessary generic framework.
- Do not modify `bchat/`, Phys repositories, deployment configuration, Neon or Fly.io.
- Do not introduce statutory payroll calculations.
- `SG-MOM-OT-001` remains BLOCKED.
- Preserve Go-canonical compensation resolution and canonical hashing under ADR-1 unless a separate, evidence-backed architecture decision changes them.

**Migration protection:**

- Migrations `00001` through `00012` are immutable.
- Do not create or apply the production `00013` migration during Gate 2.
- Experimental DDL must be confined to explicitly disposable scratch databases or an isolated test schema.
- Do not drop, truncate or recreate any user database or non-disposable database.
- Do not overwrite existing uncommitted work.
- Do not discard, reset or clean user changes to make tests pass.

The Gate 1 admission package, existing `admission.go` and its tests are valuable historical and experimental evidence. Preserve them.

---

# 3. Phase 0 — Establish the actual baseline

Before implementing anything:

1. Record the repository root, current branch, HEAD commit and `git status`.
2. Identify existing uncommitted changes and preserve them.
3. Inspect the current payroll snapshot and validation implementation.
4. Inspect the Gate 1 admission implementation and tests.
5. Inspect all migrations, role provisioning scripts, grants, triggers, SECURITY DEFINER functions and privilege tests.
6. Identify the actual PostgreSQL version used by the experiment.
7. Run the existing relevant tests and record their results.
8. Verify the checksums of migrations `00001`–`00012` against the repository's authoritative baseline records.

Do not infer the current code structure from previous plans. Verify it.

Produce a short baseline record identifying:

- Existing implementation behavior.
- Existing red attack tests.
- Existing green mechanism tests.
- Uncommitted files that must not be overwritten.
- Relevant database roles and effective privileges.
- Any discrepancy between the Gate 1 package and the actual repository.

Stop and report any baseline discrepancy that makes the proposed experiment unsafe or invalid.

---

# 4. Phase 1 — Build the complete writer inventory

The synchronization design is only as strong as its least-cooperative writer.

Identify every table and write path that can affect a payroll snapshot or its eligibility, provenance or validation.

At minimum inspect:

- Employer and employer-scoped configuration.
- Employment terms and employment status.
- Compensation records and correction/supersession chains.
- Attendance events, dispositions, corrections and intervals.
- Timesheets, timesheet lines, approval and rejection.
- SHG instructions and their verification/correction lifecycle.
- Payroll population and readiness inputs.
- Rule-pack pins and the relevant activation/verification state.
- Snapshot headers, snapshot children and lifecycle/audit records.

For each relevant table, identify:

- Employer identity source.
- Employee identity source, where applicable.
- Parent and foreign-key relationships.
- Table owner.
- Effective application-role privileges.
- Existing triggers and functions.
- All Go service, CLI, REST, trigger and direct-SQL write paths.
- Whether the path currently participates in admission or another concurrency contract.

Inspect all relevant write operations, not only ordinary Go `INSERT` and `UPDATE` calls:

- `INSERT`
- `UPDATE`
- `DELETE`
- `INSERT ... ON CONFLICT`
- `MERGE`
- `COPY FROM`
- Cascading foreign-key operations
- Partitioned-table writes and direct partition access
- Bulk and multi-row operations
- `TRUNCATE`
- Relevant privileged functions or imports

Produce a writer matrix with exact table names and actual enforcement points.

Do not call an inventory complete merely because every known Go service has been searched.

---

# 5. Phase 2 — Test the remaining Gate 1 attack accurately

The XID-ticket design has demonstrated a valid ordering proof. Do not mischaracterize it.

The outstanding question is whether the barriers proven at ticket issuance are guaranteed to remain effective through ticket verification.

PostgreSQL behavior to preserve in the experiment:

- `DISCARD ALL` releases session-level advisory locks.
- `DISCARD ALL` cannot run inside a transaction block.
- A session-level advisory unlock function can release a session lock while a transaction is open if the caller has permission.
- Transaction-scoped advisory locks cannot be manually released by the caller; they end with the transaction.

The critical attack is therefore a **release or replacement of session barriers after the RR snapshot has been established**, not simply executing `DISCARD ALL` before beginning the transaction.

## 5.1 Test A — Pre-BEGIN DISCARD ALL

On a disposable PG 16.15 database:

1. Acquire the existing session barriers.
2. Issue and commit a legitimate ticket.
3. Execute `DISCARD ALL` before starting the RR transaction.
4. Attempt to reacquire the barriers through permitted paths.
5. Begin RR and establish the snapshot.
6. Verify the ticket and lock-set state.
7. Use a separate backend to exercise a writer interleaving.
8. Inspect the result and persisted state.

Determine whether ticket lifecycle rules reject reacquisition or otherwise ensure that the new snapshot is correctly ordered after reacquisition.

Do not label the design broken merely because `DISCARD ALL` released session locks before the snapshot was taken. That ordering alone does not prove a stale snapshot.

## 5.2 Test B — Post-snapshot release/reacquisition

This is the important attack.

1. Acquire all required session barriers.
2. Issue and commit the XID ticket.
3. Begin a `REPEATABLE READ` transaction.
4. Execute a query to establish its snapshot.
5. As `app`, enumerate every permitted function or operation capable of releasing or altering the session barriers.
6. Attempt each available release path.
7. Where a release path succeeds, use a separate backend to commit a payroll-input writer while the barriers are absent.
8. Attempt to reacquire the barriers on the original backend.
9. Invoke the existing ticket verification and capture admission path.
10. Verify whether the stale RR snapshot can still pass.

Also test:

- Raw advisory unlock calls as `app`.
- All callable SECURITY DEFINER release helpers.
- Ticket single-use and hold lifecycle behavior.
- `DISCARD ALL` inside the open RR transaction, expecting PostgreSQL to reject it.
- Connection reset and backend termination behavior where applicable.

Do not invent a new attack merely to obtain a failure. Record the exact SQL statements, backend identities, interleaving, result codes and persisted database state.

**Acceptance for A:** either prove that all permitted release/reacquisition attempts fail closed after the snapshot is established, or demonstrate the vulnerability. Do not alter the production admission design during this phase.

---

# 6. Phase 3 — Design Option B

Build a focused experiment of:

**READ COMMITTED + transaction-scoped employer gate + database-enforced writer participation.**

The initial candidate is one employer-wide transaction-scoped advisory gate for the current one-employer MVP.

Do not introduce one lock per employee, compensation record, period or rule pack unless the experiment provides concrete evidence that one employer gate is inadequate.

The aim is to replace the session-lock/ticket protocol with one synchronization authority.

## 6.1 Capture protocol

The candidate capture transaction is:

```
BEGIN ISOLATION LEVEL READ COMMITTED;

-- Acquire the employer gate before reading authoritative inputs.
SELECT ... acquire_employer_payroll_gate(...);

-- Resolve and read all payroll-affecting inputs.
-- Go performs the existing canonical resolver and serialization.
-- The database verifies the submitted claims.

-- Persist immutable snapshot content.
-- Validate the snapshot and its provenance.
-- Record the validation audit.
-- Transition the payroll run to validated.

COMMIT;
```

This is a conceptual protocol, not permission to hard-code these calls without inspecting the existing code.

Requirements:

1. Capture must reject transactions not using `READ COMMITTED`.
2. The gate must be acquired before capture reads authoritative input data.
3. Capture must not trust caller-supplied input content derived from reads made before the gate.
4. All canonical input reads must occur after the gate is acquired.
5. The existing Go compensation resolver and canonical byte/hash implementation remain authoritative under ADR-1.
6. The database verifier must independently establish the completeness, identity and validity of the Go-supplied claims.
7. Snapshot creation, proof verification, validation audit and the `draft → validated` transition must have one protected decision boundary.
8. Failure must leave no partially validated run or partially committed snapshot.
9. Preserve immutable snapshot content and append-only lifecycle history.

PostgreSQL `READ COMMITTED` uses a fresh snapshot for each command. That allows capture's subsequent reads to see transactions committed before those commands begin, while the gate prevents participating payroll writers from committing during the protected interval.

The gate is not a substitute for verifying every source.

## 6.2 Gate identity and canonical key derivation

Use one canonical employer-gate identity.

- Define its purpose and scope.
- Define the exact employer identifier used to derive it.
- Centralize key derivation.
- Prevent independent Go, SQL and trigger implementations from silently diverging.
- Document hash collisions and their consequences if a hashed advisory key is used. Unintended collisions may serialize unrelated employers, but must never permit two writers for the same employer to use different keys.
- Use the same key for capture and every relevant writer.
- If a transaction can change multiple employers, either prohibit that operation through an explicit invariant or acquire all affected employer gates in canonical sorted order.

The current MVP has one operational employer. Do not build speculative multi-employer transaction machinery without a demonstrated requirement.

## 6.3 Writer participation

Every relevant writer transaction must hold the same employer gate until commit or rollback.

Implement two complementary enforcement layers:

**Application service layer**

Existing Go writer services should acquire the employer gate before doing business reads that determine the write. This prevents those services from making avoidable decisions using stale pre-wait reads.

**Database enforcement layer**

Direct SQL as `app` must not be able to modify payroll-affecting inputs without participating.

Evaluate and implement the smallest correct trigger or controlled-function design. The likely candidate is a database trigger that acquires the transaction-scoped gate before the input mutation proceeds.

Do not rely on Go helpers as the sole enforcement boundary.

Inspect and test:

- `BEFORE` trigger behavior.
- Row-level versus statement-level trigger behavior.
- Employer identity derivation for inserts, updates and deletes.
- Old/new employer changes.
- Cascades and foreign-key side effects.
- Partitioned tables and direct partition writes.
- Bulk and multi-row changes.
- `COPY FROM`, `MERGE` and `ON CONFLICT`.
- `TRUNCATE` and its separate privilege behavior.

`app` must not own the relevant tables, disable their triggers, change table ownership or assume a role with those capabilities. Explicitly inspect effective privileges and membership.

Do not assume that a trigger on a parent table automatically covers every access path. Prove it against the actual PostgreSQL schema.

If `TRUNCATE` is not required, revoke it from `app` and test the denial. Do not attempt to protect it using a row-level trigger.

A caller that voluntarily acquires the same transaction-scoped gate becomes another participant. It may cause contention or denial of service, but it must not be able to release a different transaction's gate or violate snapshot freshness.

## 6.4 Important same-transaction attack

A transaction-scoped advisory lock is reentrant for the same transaction.

Therefore, capture acquiring the gate is not sufficient to prevent the same application transaction from modifying source inputs after it has created the snapshot but before it commits validation.

Test this explicitly:

1. Begin a READ COMMITTED transaction.
2. Acquire the employer gate through the capture entry point.
3. Create the snapshot and complete the validation decision.
4. Before committing, attempt direct SQL as `app` to change a payroll input.
5. Verify that the change is rejected and no source mutation can follow the protected snapshot decision in that transaction.

The implementation must provide a database-enforced way to distinguish a transaction that has entered the capture/validation phase from an ordinary writer transaction.

Investigate the smallest safe design, such as a database-owned transaction marker associated with the top-level transaction ID, or another mechanism whose state cannot be forged or deleted by `app`.

Do not rely on:

- Go flags.
- Caller-supplied transaction state.
- Arbitrary custom GUC values that `app` can set.
- A marker table writable by `app`.
- The assumption that the caller will not issue further SQL.

The marker must be written by the controlled capture boundary, checked by input writers, rolled back automatically on failure, and scoped so unrelated later transactions are not incorrectly blocked.

Document its lifecycle and retention. Do not add an unnecessarily generic state framework.

If no safe mechanism can be demonstrated, stop and report the unresolved invariant rather than claiming the gate closes it.

## 6.5 Capture-to-validation atomicity

The protected boundary must cover the actual validation decision.

Preferred design: snapshot derivation, completeness verification, immutable persistence, validation audit and `draft → validated` happen in one gated transaction.

If the actual architecture requires capture and validation to be separate transactions:

- Validation must acquire the same employer gate.
- It must verify that the snapshot still reflects the complete authoritative input state.
- It must reject stale or fabricated content.
- It must make the validation decision and audit atomically.
- It must not accept the assumption that the previous capture gate remained held across transactions.

Document what happens to a validated snapshot when new corrections or input changes occur after its validation transaction commits. Preserve the project's historical succession and correction semantics. Do not silently rewrite or mutate the prior snapshot.

---

# 7. Phase 4 — Execute the Option B adversarial experiment

Use a disposable PostgreSQL 16.15 database and a small fixture that resembles the real schema.

At minimum, exercise representative writers for:

- Compensation changes.
- Timesheet approval or correction.
- SHG instruction changes.

Add population/employment status and attendance paths where necessary to prove the design's actual writer contract.

## 7.1 Required concurrency tests

Prove both orderings for each relevant writer class.

**Writer-first:**

1. Writer obtains the gate.
2. Writer modifies the input.
3. Capture attempts to acquire the gate and waits.
4. Writer commits.
5. Capture acquires the gate.
6. Capture reads and includes the committed input.

**Capture-first:**

1. Capture obtains the gate.
2. Capture establishes the authoritative read set.
3. Writer attempts the relevant mutation and waits.
4. Capture completes validation and commits.
5. Writer continues and commits after capture.

Also test:

- Capture during an uncommitted writer.
- Writer during capture.
- Multiple writers contending for the same employer.
- Different employers, if supported by the fixture.
- Concurrent correction and supersession.
- Population changes.
- Attendance/timesheet changes.
- The same-transaction post-capture mutation attack.
- Capture failure and rollback.
- Writer failure and rollback.
- Connection loss.
- Cancellation and deadline.
- Bounded waiting and typed timeout/deadlock errors.

Use independent database backends and explicit coordination. Do not rely on sleeps alone when advisory lock observations, channels or barriers can synchronize the test.

Assert actual waits and lock ownership where useful. Inspect committed and rolled-back data afterward.

## 7.2 Direct SQL as `app`

This is mandatory.

Connect as the actual ordinary application role, not a superuser or migration administrator.

Attempt every relevant direct write operation, including:

- Insert.
- Update.
- Delete.
- Multi-row operations.
- Upsert.
- Merge.
- Copy.
- Cascading operations.
- Direct partition writes.
- Truncate.

Prove that every permitted payroll-affecting input mutation participates in the gate.

Also test:

- `app` cannot disable relevant triggers.
- `app` cannot change table ownership.
- `app` cannot assume `migration_admin`.
- `app` cannot bypass the trigger using a privileged session setting.
- `app` cannot directly manipulate capture markers, snapshot children, validation state or audit records.

Do not claim that direct SQL is covered merely because a Go service test passes.

## 7.3 Failure and persisted-state assertions

After every failed capture or writer operation, inspect the database.

Assert:

- No partial snapshot content.
- No invalid snapshot lifecycle event.
- No validation audit for a rejected operation.
- No transition to `validated` after an unsuccessful proof.
- No lost, duplicated or rewritten historical source evidence.
- No permanent gate or marker state that blocks unrelated transactions.
- No partial source correction.
- No stale snapshot accepted as current.

A test passing without inspecting persisted state is insufficient.

---

# 8. Phase 5 — Evaluate the alternatives without expanding scope

Option B is the leading candidate. Do not build several competing production architectures.

You may evaluate a parent-row gate as a narrow comparison if it is genuinely useful. Compare it with the single employer-wide transaction-scoped advisory gate on:

- Database enforcement.
- Direct SQL coverage.
- Lock conflict semantics.
- FK/cascade behavior.
- Privileges required.
- Deadlock risk.
- Writer implementation complexity.
- Testability.
- Operational behavior.

Choose one synchronization authority. Do not maintain both mechanisms in production.

Do not implement append-only event sourcing in this gate. Record it as a possible longer-term data-model improvement. Do not use `MAX(event_id)` or sequence allocation as a proxy for commit order.

Do not pursue SERIALIZABLE unless the business explicitly confirms that a serialization-point snapshot is acceptable instead of the stronger existing capture-boundary requirement. If you believe SSI is materially superior, document that as a separate decision proposal; do not silently weaken the requirement.

Do not introduce a third role during this experiment.

---

# 9. Snapshot authority remains P0 under either design

The synchronization gate does not solve direct-SQL snapshot fabrication.

Preserve ADR-1:

- Go compensation resolver remains canonical.
- Go canonical byte representation and SHA-256 remain the single canonical implementation.
- SQL must not implement a second compensation resolver.
- SQL must not independently recreate the Go canonical serializer.
- Database verification must establish the completeness and validity of the submitted claims.

The experiment must include negative tests for:

- Omitted expected employee.
- Fabricated employee.
- Omitted compensation segment.
- Fabricated compensation amount or source.
- Incorrect effective-date coverage.
- Incorrect supersession lineage.
- Wrong employer or pay period.
- Stale input version.
- Invalid SHG source or verification.
- Fabricated rule-pack reference.
- Incomplete or altered snapshot child set.
- Forged hash or validation status.
- Direct mutation of validated content.

Verify that a matching hash does not allow fabricated source content to pass.

Do not weaken snapshot immutability or validation privileges to make the spike easier.

---

# 10. Phase 6 — Gate 2 decision criteria

Compare Option A and Option B using evidence.

The following must be answered:

1. Can `app` release or replace the current session barriers after an RR snapshot is established?
2. Does the XID-ticket prove only ticket ordering, or can the implementation also prove uninterrupted barrier ownership?
3. Does Option B serialize all relevant writers and capture under the actual `app` threat model?
4. Can any relevant direct SQL write bypass the Option B trigger/gate?
5. Can the application transaction modify payroll inputs after capture has entered its protected phase?
6. Does the gate remain held through the validation decision?
7. Are the authoritative Go claims completely verified by the database?
8. Can all failure paths roll back without leaving partial state?
9. Is the selected design measurably simpler to implement, test and operate?
10. What residual limitations remain, and are they explicit and acceptable?

Do not choose a design based on the number of green tests or the amount of previous implementation work.

Do not claim Option B is correct just because it uses fewer objects. Its writer boundary and same-transaction protection must pass the same invariant-level scrutiny.

## GO criteria for Option B

Option B may be recommended for production implementation only when:

- Capture rejects inappropriate isolation levels.
- Capture obtains the employer gate before authoritative reads.
- All relevant writers participate through database-enforced paths.
- Direct SQL as `app` cannot bypass writer participation.
- Both concurrency orderings pass for every relevant writer class.
- Snapshot derivation, verification and validation are one protected decision.
- Same-transaction post-capture source mutation is rejected.
- Authority and provenance tests pass.
- Rollback, cancellation and connection failure leave correct persisted state.
- Existing historical and statutory boundaries remain intact.

If any of these conditions remains unresolved, report NO-GO for production implementation.

## GO criteria for retaining Option A

Option A should remain a viable fallback only if the post-snapshot release/reacquisition attack is demonstrably closed under the actual application privileges and the same writer/authority acceptance standards.

A pre-BEGIN `DISCARD ALL` test alone is not sufficient to reject it or approve it.

---

# 11. Deliverables

Produce the following:

### A. Gate 2 experiment implementation

Only isolated scratch/test code, fixtures, SQL and test harnesses needed to prove or reject the candidate designs.

Do not create or apply production `00013`.

Keep experimental code clearly separated from production code. Preserve the existing Gate 1 experiment.

### B. Gate 2 report

Create `plans/gate2_synchronization_experiment.md` with:

1. Executive decision.
2. Exact invariant and capture-boundary semantics.
3. Repository and database baseline.
4. Complete payroll-input writer inventory.
5. Option A attack sequence, SQL, result and persisted-state evidence.
6. Option B architecture and exact transaction protocol.
7. Canonical key derivation and trigger coverage.
8. Same-transaction post-capture protection.
9. Direct-SQL privilege tests.
10. Writer-first and capture-first results.
11. Snapshot authority and negative-test results.
12. Rollback, timeout, cancellation and connection-failure results.
13. Known limitations and residual risks.
14. Option A versus Option B comparison.
15. Final GO/NO-GO decision.
16. Exact proposed production migration scope if GO.
17. Explicitly deferred work.

Distinguish clearly between:

- Observed database results.
- Tests actually executed.
- Code inspected.
- Agent inference.
- Untested assumptions.
- Recommended next actions.

Do not call a test green unless it was executed successfully against PostgreSQL.

### C. Decision record

Create a concise ADR that records:

- Chosen synchronization architecture.
- Why the alternative was rejected or retained.
- The capture-boundary definition.
- Writer participation contract.
- Snapshot authority contract.
- Known limitations.
- Required future changes to Prompt17 and the M2.1 plan.

If Option B passes, explicitly supersede Prompt17's pre-BEGIN RR/session-lock protocol. Do not leave conflicting specifications in circulation.

### D. Final agent response

Report:

- What you actually implemented.
- Exact files changed or added.
- Exact tests executed and results.
- Database version and role identities used.
- Whether the A attack passed or failed.
- Whether Option B passed or failed each acceptance condition.
- Remaining risks.
- GO/NO-GO recommendation.
- Whether production migration `00013` is authorized by the evidence (it is not authorized merely by completing the experiment).

Do not claim to have independently verified anything that was not observed.

---

# 12. Explicit prohibitions

Do not:

- Implement production migration `00013` before Gate 2 approval.
- Modify migrations `00001`–`00012`.
- Modify `bchat/` or unrelated repositories.
- Add a third database role without approval.
- Introduce a new ORM or replace `pgx`.
- Duplicate the compensation resolver or canonical hash implementation.
- Use application-only checks to enforce database invariants.
- Accept a `SECURITY DEFINER` function without correct ownership, restricted `search_path`, schema-qualified references and narrow execution grants.
- Grant `app` broad table ownership or DDL capability.
- Permit `app` to fabricate snapshots, validation records, gate markers or audit records.
- Replace a failing test with a weaker assertion.
- Suppress a blocker or reinterpret a failed attack as an expected result without evidence.
- Modify historical data to make acceptance pass.
- Implement payroll statutory calculations.
- Unblock `SG-MOM-OT-001`.
- Declare the milestone complete simply because the experimental tests pass.

**The purpose of Gate 2 is to find the simplest database-enforced synchronization boundary that satisfies the real payroll snapshot invariant—not to justify the design that has already consumed the most engineering effort.**

The key safeguard added to this prompt is the same-transaction post-capture mutation test. A transaction-scoped lock alone does not prevent the same transaction from reacquiring that lock and attempting another input write; the agent must prove that this is blocked before recommending Option B for production.