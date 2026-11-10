I recommend a red-team code review, not another general architecture review. The reviewer should assume the implementation is wrong until the actual code, SQL enforcement and execution evidence establish otherwise.

In particular, three findings deserve scrutiny: valid-parent snapshot fabrication remains possible, the DBOS conventional-job comparison was never built, and immudb's host-file corruption test and out-of-band genesis commitment remain outstanding.&#x20;

POC-FINAL-REPORT.md



Here is the prompt.

# FINAL ADVERSARIAL CODE REVIEW — PostgreSQL 18.6 + DBOS + immudb Payroll POC

## 1. Assignment

Act as an independent Principal Engineer, Database Security Architect, Go Runtime Specialist, Distributed Systems Engineer and Red-Team Reviewer.

Conduct a rigorous adversarial review of the completed `stack-poc` implementation.

Your job is to challenge the correctness of the code, SQL enforcement, workflow execution, evidence integrity and the conclusions in the final POC reports.

You are NOT being asked to implement more functionality, extend the payroll product or repeat the original experiments.

The objective is to answer one question:

**Does the actual implementation justify the architectural claims made about PostgreSQL, DBOS and immudb, or are there hidden defects, misleading test results, incomplete guarantees or unsupported conclusions?**

Be skeptical. A successful build, passing test or successful workflow is not proof that the underlying architectural guarantee is correct.

Inspect the implementation and independently verify the claims.

## 2. Authoritative baseline

Read:

- `AGENTS.md`
- `plans2/POC-FINAL-REPORT.md`
- `plans2/POC-RESULTS.md`
- `poc/payroll-month/POC-REPORT.md`
- `poc/payroll-month/POC-RUNBOOK.md`
- `plans2/PHASE2-WORKLOAD.md`
- `plans2/TRACEABILITY.md` or its current documented equivalent
- `docs/stack/RESEARCH.md`
- `docs/stack/PATTERNS.md`
- `docs/stack/DECISIONS.md`
- `docs/stack/DECISION-REQUESTS.md`
- `docs/stack/RULES-DYNAMICS.md`
- ADR-0201 through ADR-0204.

Inspect the actual repository and all relevant code, SQL, migrations, Go tests, fixtures, Compose configuration and workflow implementations.

The authoritative execution versions are:

- PostgreSQL 18.6.
- DBOS Transact Go v1.8.0.
- immudb v1.11.2.
- pgx v5.12.0.

The POC is on branch `stack-poc`, with uncommitted work.

Preserve that state.

The POC is a demonstration of three architectural responsibilities, not a complete payroll product.

The shared workload is one synthetic employer, 25 employees and October 2026's 31-day payroll period.

All payroll outputs are illustrative and non-statutory. `SG-MOM-OT-001` remains BLOCKED.

## 3. Review rules

1. Independently inspect the implementation rather than trusting report descriptions.
2. Follow actual data and transaction paths, not just function names.
3. Review database enforcement under real runtime roles.
4. Challenge all claims of atomicity, exactly-once behavior, idempotency and immutability.
5. Distinguish verified guarantees from guarantees asserted by code structure.
6. Reproduce meaningful suspected defects in isolated disposable fixtures.
7. Do not damage the canonical payroll data or the existing immudb ledger.
8. Do not modify migrations 00001–00012.
9. Do not create migration 00013.
10. Do not change business semantics or weaken security controls.
11. Do not fix defects during the review.
12. Do not commit, push or merge.

You may run non-destructive tests and create disposable test fixtures to prove a finding.

This is an independent review. Do not turn the review into a new implementation phase.

## 4. PostgreSQL adversarial review

Investigate whether PostgreSQL actually enforces the business invariants attributed to it.

### A. Effective-dated integrity

Inspect:

- Range representations.
- `WITHOUT OVERLAPS` constraints.
- Exclusion constraints.
- Compensation and employment-term relationships.
- Rule-pack validity.
- Coverage verification.
- Successor and correction semantics.

Challenge these claims:

- Can overlapping effective-dated records be created?
- Can uncovered periods silently pass?
- Are all required relationships checked?
- Can two concurrent writers bypass the intended protection?
- Does the implementation confuse no-overlap with complete temporal coverage?

Particular attention:

The report found that PostgreSQL 18.6 rejected the proposed PERIOD foreign-key syntax with SQLSTATE 42601, and that `poc_verify_claims` is broken with SQLSTATE 42703 (`source_table`).

Trace the actual fallback mechanism. Determine whether it fails closed.

Do not claim complete-period coverage unless the implementation proves it.

### B. Employer gate and concurrency

Inspect the full transaction path for:

- Employer gate acquisition.
- Transaction isolation.
- Claim creation.
- Participation triggers.
- Lock ordering.
- Proof checks.
- Close transaction.
- Writer participation.

A gate-acquisition test does not automatically prove that the real payroll capture path participates in that gate.

Verify that every authoritative writer participates in the required boundary.

Investigate:

- Same-employer serialization.
- Different-employer concurrency.
- Writer-first versus capture-first interleavings.
- Deadlock possibilities.
- Stale-read opportunities.
- Cutoff races.
- Close-versus-write races.

The report explicitly states that the full writer-versus-capture race was not executed.

Determine whether the implementation's actual architecture closes this gap or merely demonstrates the lock primitive.

### C. Snapshot fabrication — critical

The report records that fabrication using random parent identifiers is rejected by a foreign key, but fabrication using valid parent records remains RED in `TestSnapshotFabricationByApp`.

Inspect this finding in depth.

Can the actual `app` role fabricate a snapshot using legitimate existing parents?

Can it bypass the snapshot creation function?

Can it fabricate child rows or rule pins?

Can it create apparently valid but unearned validation state?

Can it cause the application or auditor to accept a fabricated artifact?

Inspect all direct INSERT, UPDATE and DELETE paths.

Examine ownership, grants, triggers, DEFINER functions and validation transitions.

Identify the exact boundary at which the implementation does or does not enforce authority.

**Do not downgrade this finding simply because a later ledger or verification step might detect it.**

Database authority must be established before a fabricated artifact becomes authoritative.

### D. State-machine and close protections

Inspect:

- Direct status transitions.
- Validation creation.
- Empty/vacuous validation.
- Period finalization.
- Closed-period write protection.
- Correction and adjustment lineage.
- Back-link immutability.

Determine whether direct SQL under the runtime `app` role can bypass any transition.

Distinguish a properly protected finalization operation from a workflow that merely happens not to expose an unsafe API.

### E. PostgreSQL 18.6 versus 16.15

Independently examine the recorded E5 results.

Verify whether the conclusions accurately distinguish:

- Genuine PostgreSQL-version differences.
- Existing intentional RED tests.
- Test-environment timezone artifacts.
- Gate concurrency flakes.
- Test-harness defects.

Do not accept "zero PG18-specific failures" as equivalent to "the application is fully correct."

Also challenge whether the actual value of PostgreSQL 18.6 for this POC is predominantly new temporal functionality or operational consistency with the desired long-term baseline.

## 5. DBOS adversarial review

### A. Atomic transaction boundary

Trace the actual implementation of `RunAsTransaction`.

Verify:

- The correct shared pool.
- Transaction ownership.
- Transaction propagation through adapters and services.
- No accidental acquisition of independent pool connections.
- Business writes and DBOS durability records committed together.
- Rollback behavior.
- Error propagation.

Identify any operations that occur outside the atomic boundary but are described as part of it.

### B. Workflow determinism and replay

Inspect:

- Workflow identity generation.
- Step ordering.
- Step identifiers.
- Retry behavior.
- Restart and recovery.
- Workflow code changes.
- Patch usage.
- Duplicate requests.
- Concurrent workers.
- Ownership fencing.

Verify that unique workflow IDs are generated correctly across processes and restarts.

Review the fix for sticky DBOS workflow IDs and ensure it does not merely hide a test-order dependency.

### C. Exactly-once claims

Challenge every occurrence of "exactly once."

Clearly distinguish:

1. Atomic database transaction.
2. DBOS workflow completion deduplication.
3. At-least-once workflow step execution.
4. External side-effect idempotency.
5. Outbox delivery.
6. Financial artifact uniqueness.

Determine precisely which operations have been proven exactly-once and which rely on deterministic identifiers and retry-safe handlers.

Look for duplicate payments, duplicate snapshots, partially completed workflows and duplicate provenance events.

### D. Approval waiting and interruptions

The report confirms that crash injection while awaiting approval was not executed because the workflow contains no actual approval wait.

Verify whether the current workflow supports durable human approval at all.

Do not assume DBOS's ability to wait makes the application workflow's approval implementation correct.

Record this as a coverage gap, not an automatically successful DBOS capability.

### E. Conventional-job baseline

Exp2A was not built.

Do not pretend this comparison was measured.

Evaluate whether the recovery mechanisms identified in the report accurately describe what a competent PostgreSQL + Go implementation would need.

Identify which mechanisms DBOS eliminates and which application responsibilities remain.

Do not build the missing baseline during this review.

### F. Resource and shutdown behavior

Investigate the reported DBOS shutdown connection leak and `ClosePoolTimeout` workaround.

Determine:

- Whether connections are leaked.
- Whether application shutdown can hang.
- Whether the workaround creates a resource leak or hides one.
- Whether workflow completion and pool closure can race.
- Whether stdout logging can interfere with machine-readable output.

Assess the production implications without expanding the POC.

## 6. immudb adversarial review

### A. Provenance construction

Inspect the actual v1.11.2 native SDK integration.

Verify:

- Event-specific keys.
- Digest-bound idempotency markers.
- Append-only event behavior.
- `MustNotExist` semantics.
- Event identity uniqueness.
- `supersedes` lineage.
- Canonical JSON serialization.
- Transaction ID and ALH recording.
- PostgreSQL back-link integrity.

Challenge whether an idempotency marker can ever be reused, expired, deleted or rebound to a different payload.

Check whether the marker binds the intended operation and canonical content digest.

### B. Dual-write recovery

Follow the actual sequence:

PostgreSQL business commit → outbox → DBOS worker → immudb append → verification → PostgreSQL back-link.

Inspect every failure window.

Confirm whether retries are safe after:

- Lost responses.
- Ambiguous timeouts.
- Worker restart.
- Already-committed append.
- Missing back-link.
- Duplicate delivery.

Determine whether any state can become permanently inconsistent without alerting the operator.

### C. Independent auditor and trust anchors

The report records that transactions 18, 19 and 20 passed auditor verification using a fresh state directory and explicit genesis.

However, the out-of-band genesis commitment remains pending.

Investigate whether the auditor's initial trust state actually originates independently of the writer and ledger server.

Determine what happens when the writer, ledger server, host or storage administrator is compromised.

Explicitly distinguish:

- Independent cryptographic verification of transaction records.
- Detection of an invalid transaction reference.
- Detection of actual ledger storage corruption.
- Protection against a privileged host administrator.
- Independent retention of the genesis checkpoint.

Do not treat a fresh state directory alone as proof of independent trust.

### D. Host-file corruption

The host-level corruption harness was not built.

Assess whether the implementation's current claims exceed the evidence.

Do not modify the canonical ledger.

Identify the smallest safe disposable test that would establish whether actual persisted ledger-file corruption is detected by the independent auditor.

### E. Privacy and payload integrity

Verify the actual stored payloads.

Do not rely exclusively on schema inspection.

Check whether identifiers or metadata can expose personal or financial information.

Confirm that the immutable evidence identifies the correct business artifact without storing sensitive payroll data.

### F. Commercial and operational dependencies

Inspect the replaceable interface, dependency pinning, configuration and operational boundaries.

Record the BSL-1.1 license as a commercial adoption condition, not a code defect.

Determine whether application code is improperly coupled to immudb-specific implementation details.

## 7. Review whether the three experiments are fair

This is an especially important part of the review.

### PostgreSQL baseline

The report demonstrates that an unconstrained plain table accepted overlapping records.

Determine whether this baseline fairly represents application-only enforcement or simply an unprotected table.

Clearly state what the test proves and what it does not.

Do not equate an unconstrained table with a competent Go application that performs validation.

Also verify that the described compensation overlap experiment actually exercised compensation records or whether the real fixture was a rule-pack table.

Distinguish overlap detection from gap and coverage detection.

### DBOS baseline

The conventional-job baseline was not implemented.

Confirm that no measured-vs-measured conclusion is presented.

Identify the minimum credible engineering mechanisms an ordinary Go + PostgreSQL worker would need to reproduce the observed guarantees.

### immudb baseline

The PostgreSQL-only baseline demonstrates that privileged DELETE can rewrite a trigger-less outbox.

Challenge whether this sufficiently represents a deliberately append-only PostgreSQL audit design.

Distinguish the security of ordinary runtime users from the capabilities of a privileged database administrator.

The relevant comparison is whether PostgreSQL alone can independently establish historical integrity against an adversary who controls its administrative boundary.

Do not use a deliberately weak baseline to overstate immudb's incremental value.

## 8. Reconcile reports with implementation

Check every material statement in:

- `plans2/POC-FINAL-REPORT.md`
- `plans2/POC-RESULTS.md`
- `poc/payroll-month/POC-REPORT.md`

Find contradictions between:

- Test output and written claims.
- Experiment status and evidence.
- PASS versus PARTIAL classifications.
- Architectural guarantees and actual implementation.
- Canonical payroll outputs and DB-backed workflow results.
- PostgreSQL version comparison and regression conclusions.

In particular, establish whether the final DB-backed month actually uses and persists the same calculation results represented by the canonical digest, or whether the canonical result is generated separately from the database-backed workflow.

The existence of a successful workflow status must not be mistaken for proof that every expected payroll output was computed, persisted and verified.

Preserve the distinction between:

- IMPLEMENTED
- BUILT
- EXECUTED
- PASSED
- FAILED
- BLOCKED
- UNEXECUTED

## 9. Severity classification

Use the following severity levels:

**P0 — Critical:** Incorrect financial results, unauthorized authoritative artifact creation, broken transaction guarantees or undetected alteration of finalized payroll evidence.

**P1 — High:** Security boundary bypass, incorrect concurrency behavior, duplicate business effects, false verification or significant recovery defects.

**P2 — Medium:** Incomplete coverage, unreliable operational behavior, non-deterministic recovery or misleading architectural claims.

**P3 — Low:** Maintainability, clarity, observability, configuration or minor test-quality issues.

For each finding include:

- Finding ID.
- Severity.
- Affected experiment.
- Exact file and line references.
- Actual code path and relevant SQL.
- Attack or failure scenario.
- Expected invariant.
- Actual behavior.
- Reproduction instructions.
- Evidence.
- Business impact.
- Whether it affects the POC verdict, Gate C recommendation or future production readiness.
- Recommended remediation.

Avoid generic advice.

A finding must be concrete and supported by the actual repository.

## 10. Required report

Create:

`plans2/POC-ADVERSARIAL-REVIEW.md`

Use this structure:

1. Executive verdict.
2. Scope and methodology.
3. Findings summary by severity.
4. PostgreSQL review.
5. DBOS review.
6. immudb review.
7. Comparative baseline fairness.
8. Payroll output and reproducibility verification.
9. Report-versus-code discrepancies.
10. Detailed finding register.
11. Technology-by-technology adoption reassessment.
12. Final recommendation.

Include a compact summary table:

| Technology | Claimed guarantee | Independently verified | Finding | Confidence |
| ---------- | ----------------- | ---------------------- | ------- | ---------- |

Classify each guarantee as:

- PROVEN.
- PARTIALLY PROVEN.
- NOT PROVEN.
- CONTRADICTED.

## 11. Final recommendation

Reassess the previous Gate C recommendations:

- PostgreSQL 18.6: ADOPT.
- DBOS: CONDITIONAL ADOPT.
- immudb: CONDITIONAL ADOPT.
- Combined architecture: ADOPT as the preferred architectural direction, subject to individual conditions.

Keep or change these recommendations only when the evidence justifies it.

Clearly distinguish:

- POC adoption.
- Development-stack adoption.
- Production architecture authorization.
- Commercial licensing authorization.

Do not infer production readiness from a successful synthetic payroll run.

## 12. Execution boundaries

This is an adversarial review, not another development phase.

Do not add payroll functionality.

Do not implement the missing baseline.

Do not fix the existing RED tests during this assignment.

Do not modify the POC implementation, existing reports or frozen migrations.

Do not commit or merge.

Use isolated disposable fixtures for any attack that requires modifying data.

The one new deliverable is `plans2/POC-ADVERSARIAL-REVIEW.md`.

## Final instruction

Assume that a sophisticated reviewer will challenge every architectural claim.

Your job is to discover what they would find before they do.

Be especially skeptical of claims involving:

- Rung 5 database enforcement.
- Valid-parent snapshot fabrication.
- Transaction atomicity.
- Exactly-once behavior.
- Historical immutability.
- Independent verification.
- Baseline fairness.
- Historical payroll reproducibility.

Do not defend the architecture.

Do not try to make the POC look successful.

**Establish what is actually correct, what is only partially demonstrated, and what is wrong.**

The final report must be candid enough that we can confidently decide whether the architecture deserves to be retained.

## My recommendation

I would make this the last technical review before archiving the POC.

The three most important questions for the reviewer are:

1. Can the application create a valid-looking but unauthorized payroll snapshot?
2. Does DBOS genuinely protect the entire critical transaction boundary, rather than just its own workflow history?
3. Can immudb's independent auditor still detect tampering when the adversary controls the ledger host and its storage?

Those questions will tell us much more about architectural credibility than another round of green unit tests.