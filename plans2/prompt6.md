I recommend making this the final execution assignment. Three experiments, one shared monthly payroll fixture, one final report. No more architectural expansion, no additional product features and no endless research.

The three experiments should demonstrate the value of each technology against a credible alternative without it.

# FINAL POC EXECUTION — Three Experiments, One Monthly Payroll

## Mission

Complete the workforce-management architecture POC by executing exactly three experiments demonstrating the distinct architectural value of:

1. PostgreSQL 18.6 — business-data integrity.
2. DBOS Transact Go — durable workflow execution.
3. immudb — independently verifiable, tamper-evident evidence.

The objective is NOT to build a complete payroll system.

The objective is to demonstrate, through one realistic monthly synthetic payroll workload, why these three technologies are valuable when applied to their respective architectural responsibilities, and what additional engineering complexity would otherwise be required.

**This is the final POC scope. Once the three experiments have been executed, evaluated and documented, the POC is complete.**

Do not invent additional milestones, expand product functionality or start another research programme.

---

# 1. Existing State and Authority

Continue from the existing `stack-poc` branch.

Read the current:

- `poc/payroll-month/POC-REPORT.md`
- `poc/payroll-month/POC-RUNBOOK.md`
- `plans2/plan_prompt2.md`
- `plans2/plan_prompt4.md`, `plans2/prompt5_closeout.md` where relevant
- `docs/stack/TRACEABILITY.md`
- `docs/stack/DECISIONS.md`
- `docs/stack/DECISION-REQUESTS.md`
- `docs/stack/RESEARCH.md`
- `docs/stack/PATTERNS.md`
- `docs/stack/RULES-DYNAMICS.md`
- ADR-0201 through ADR-0204
- `AGENTS.md`

Use the latest actual implementation and evidence, not older architectural proposals.

Gate A and Step 0 are complete and signed.

All human rulings R1–R7 and P1–P5 remain binding.

You have full delegated engineering authority to execute the work, fix implementation defects, configure the designated POC environment and make routine engineering decisions without repeatedly asking for permission.

Preserve the existing working branch and uncommitted changes.

Do not commit, push or merge.

The existing PostgreSQL 16 environment and data must remain untouched.

Migrations 00001–00012 remain frozen. Do not create migration 00013.

All POC data must be synthetic.

---

# 2. Shared Workload: One Complete Monthly Payroll

Use the existing synthetic payroll calculator and database-backed POC implementation.

The common workload is:

- One legal employer.
- Twenty-five synthetic employees.
- One complete 31-day monthly payroll period.
- Different compensation arrangements.
- Approved attendance and timesheets.
- Leave and absence.
- Synthetic overtime.
- Allowances and deductions.
- Effective-dated compensation and rule changes.
- Late approvals and corrections.
- Payroll calculation.
- Exception handling.
- Approval and finalization.

Use the existing ten passing synthetic calculation vectors and their established rules.

Produce one canonical payroll snapshot and one finalized payroll result.

The identical underlying payroll workload should be used in all three experiments so that their results are comparable.

Do not build a new payroll calculator or duplicate the existing business logic.

The synthetic payroll must be clearly identified as illustrative and non-statutory.

`SG-MOM-OT-001` remains BLOCKED. Do not implement or claim statutory Singapore overtime compliance as part of this POC.

The required outputs are:

- Payroll register.
- Employee-level payslip calculations.
- Exception report.
- Payment-file preview.
- Finalized payroll snapshot.
- Canonical content digest.

---

# 3. EXPERIMENT ONE — PostgreSQL: Business Integrity Under Adversity

## Research question

What complexity do we inherit when business invariants are enforced only in application code rather than at the database boundary?

What does PostgreSQL 18.6 materially provide when we place those invariants in the correct database mechanisms?

## Scenario

Use the monthly payroll workload to demonstrate realistic conflicts between competing business operations.

Select a small number of representative cases:

1. Two competing compensation records that overlap or leave an uncovered period.
2. An approved timesheet being changed while payroll is capturing its authoritative inputs.
3. An unauthorized direct SQL attempt to fabricate or modify a payroll snapshot.
4. A cutoff or finalization transaction racing a competing business write.

## Experiment design

Demonstrate two approaches in isolated disposable test fixtures:

**A. Application-only enforcement**

Use a controlled baseline where the relevant protections are absent or limited to application-side validation.

Demonstrate the failure mode:

- Race conditions.
- Stale decisions.
- Overlapping records.
- Incomplete coverage.
- Direct SQL bypass.

Do not weaken the actual POC schema or remove its security protections to perform these demonstrations. Use isolated disposable baseline tables or a controlled comparison harness.

**B. PostgreSQL-enforced integrity**

Demonstrate the actual PostgreSQL implementation using:

- Transactions.
- Appropriate isolation.
- Transaction-scoped employer gate.
- Constraints.
- Temporal coverage checks.
- Triggers and DEFINER functions.
- Database privileges and revocations.

Run the competing operations concurrently and verify persisted state.

Use the actual `app` role, not superuser.

## Evidence required

Show:

- What the application-only baseline permits or mishandles.
- Which PostgreSQL mechanism prevents each problem.
- Which invalid operations are rejected.
- Which concurrent operations are serialized.
- Which missing or incomplete records produce blockers.
- Whether finalized payroll artifacts remain protected.

Include actual SQL, test output and resulting database states.

## Success criterion

Demonstrate that PostgreSQL protects authoritative business state even when application-level validation is insufficient or bypassed by direct SQL through the application role.

Conclude what custom concurrency, validation and enforcement infrastructure would otherwise be necessary.

---

# 4. EXPERIMENT TWO — DBOS: Durable Execution Under Failure

## Research question

What additional infrastructure must we build ourselves when long-running business processes use ordinary Go workers and PostgreSQL without a durable workflow engine?

What does DBOS remove from that engineering burden?

## Scenario

Process the same monthly payroll workload through the complete lifecycle:

1. Capture approved inputs.
2. Resolve compensation and applicable synthetic rules.
3. Calculate payroll.
4. Identify exceptions.
5. Wait for approval.
6. Finalize the payroll.
7. Produce the evidence/outbox event.

## Experiment design

Compare:

**A. Conventional job execution**

Demonstrate a controlled baseline representing ordinary Go background processing with PostgreSQL state but without durable workflow checkpoints.

Show what happens when:

- The worker crashes after completing a step.
- The same request is submitted twice.
- The process restarts.
- An approval wait is interrupted.
- A previously completed operation is attempted again.

Record the additional state-tracking, recovery, idempotency and coordination logic needed to make this approach reliable.

**B. DBOS durable execution**

Run the equivalent workflow using the actual DBOS Transact Go implementation.

Inject representative failures:

1. Crash before the capture transaction commits.
2. Crash immediately after the capture transaction commits.
3. Crash after a completed payroll calculation step.
4. Crash while awaiting approval.
5. Duplicate workflow submission.
6. Restart and recovery of an interrupted workflow.

Verify the actual shared-pool transaction boundary.

The business capture and DBOS durability record must commit atomically through the appropriate `RunAsTransaction` path.

Do not claim ordinary `RunAsStep` external side effects are exactly-once.

The immudb outbox worker must remain retry-safe through deterministic idempotency.

## Evidence required

Record:

- Workflow execution history.
- Step checkpoints.
- Transaction outcomes.
- Crash locations.
- Recovery behavior.
- Number of executions and financial postings.
- Whether duplicate or partial payroll artifacts occurred.
- Time and effort needed by the baseline recovery mechanism versus DBOS.

## Success criterion

Demonstrate that DBOS substantially reduces the amount of custom workflow infrastructure needed to coordinate payroll processes and recover from interrupted execution.

The experiment must show that the payroll workflow resumes correctly without duplicating or silently modifying authoritative financial artifacts.

Conclude which orchestration mechanisms DBOS eliminates, which still belong in PostgreSQL, and which require application-level idempotency.

---

# 5. EXPERIMENT THREE — immudb: Independent Evidence and Tamper Detection

## Research question

What assurance do we lose when finalized payroll history is kept exclusively in an ordinary operational database?

How does immudb provide cryptographically verifiable evidence that ordinary PostgreSQL audit tables cannot independently establish?

## Scenario

Use the finalized payroll result from the previous experiments.

Create immutable evidence covering:

- Canonical payroll snapshot digest.
- Rule-pack identity and content digest.
- Calculator implementation/build digest.
- Finalized payroll result digest.
- Approval and finalization references.
- Relevant PostgreSQL artifact identifiers.
- Evidence-event identifiers.

Keep PII, bank information, NRICs, secrets and unnecessary employee identifiers out of immudb.

## Experiment design

Demonstrate:

**A. PostgreSQL-only evidence**

Show an ordinary operational audit implementation and its limitations under an appropriately privileged database administrator.

Use a disposable clone or isolated database for tampering experiments.

Demonstrate that modifying the payroll snapshot and its corresponding ordinary audit records can undermine trust in the database's internal history.

Do not destroy or corrupt the canonical original payroll result.

**B. PostgreSQL + immudb evidence**

Anchor the original canonical payroll evidence into immudb through the existing transactional outbox and native SDK integration.

Use the approved event-specific keys and digest-bound idempotency design.

Verify the resulting transaction using the immudb verification APIs.

Run an independent auditor using trusted state and genesis information outside the writing application's mutable control.

Then modify the disposable PostgreSQL copy of the finalized snapshot and attempt to reconcile it against the original anchored evidence.

Demonstrate:

- Digest mismatch detection.
- Cryptographic proof verification.
- Independent auditor verification.
- The inability to silently replace the anchored event.
- Recovery after the ledger append succeeds but the PostgreSQL back-link is missing.
- Safe retry after an ambiguous network outcome.

Do not confuse immudb's ability to verify a record with proof that the original business calculation was correct.

## Evidence required

Record:

- Original canonical digest.
- immudb transaction ID and ALH.
- Verified transaction result.
- Independent auditor checkpoint.
- Tampering operation.
- Reconciliation result.
- Cryptographic verification outcome.
- Recovery behavior.
- Any limitations of the actual threat model.

## Success criterion

Demonstrate that independently anchored evidence makes unauthorized historical modification detectable, whereas ordinary PostgreSQL audit records alone do not provide the same independent cryptographic assurance.

Conclude the additional assurance and the extra operational complexity introduced by immudb.

---

# 6. Comparison and Final Assessment

After executing all three experiments, produce one concise comparative assessment.

For each technology, answer:

1. What problem does it solve?
2. What did we have to implement without it?
3. What did the technology eliminate or simplify?
4. What complexity did it introduce?
5. What did the experiment actually prove?
6. What remains our application responsibility?

Include a final decision matrix:

| Layer               | Without designated technology              | With designated technology | Demonstrated value |
| ------------------- | ------------------------------------------ | -------------------------- | ------------------ |
| Data integrity      | Custom application checks and coordination | PostgreSQL enforcement     | Actual result      |
| Workflow durability | Custom jobs, checkpoints and recovery      | DBOS                       | Actual result      |
| Historical evidence | Ordinary audit and custom verification     | immudb                     | Actual result      |

Make the conclusions evidence-based.

Do not automatically recommend all three technologies merely because the POC uses them.

Assess their incremental architectural value independently.

---

# 7. Final Deliverables

Produce only the following final POC deliverables:

**1. `POC-FINAL-REPORT.md`**

A consolidated report containing:

- Purpose and scope.
- Monthly payroll workload description.
- Experiment 1 findings.
- Experiment 2 findings.
- Experiment 3 findings.
- Actual commands and execution evidence.
- Screenshots or logs where useful.
- Comparison of the three architectural approaches.
- Engineering complexity eliminated.
- Complexity introduced.
- Technology-by-technology recommendations.
- Remaining limitations.

**2. `POC-RESULTS.md`**

A concise demonstration-oriented summary containing:

- What was executed.
- What failed in each baseline.
- What PostgreSQL prevented.
- What DBOS recovered.
- What immudb detected.
- The final payroll output and its verification status.

Update `POC-REPORT.md` with the final execution status, but do not create another parallel collection of architecture plans or ADRs.

---

# 8. Scope Restrictions — This Ends the POC

There are exactly three experiments.

Do not:

- Build a complete payroll system.
- Implement additional HR modules.
- Build a production statutory payroll engine.
- Implement the blocked SG-MOM overtime package.
- Introduce another workflow engine.
- Introduce another ledger technology.
- Add a new rules engine.
- Expand the workforce-management domain unnecessarily.
- Repeat Gate A research.
- Re-run every previously completed experiment without a reason.
- Execute exhaustive E1–E6 and T1–T13 matrices as separate projects. Use only the focused cases necessary to establish these three experiments.
- Start new migrations or modify 00001–00012.
- Create migration 00013.
- Start production adoption or deployment.
- Add features merely to make the demonstration look sophisticated.

Reuse existing implementations and tests.

Fix only defects that prevent one of the three experiments from executing correctly.

Do not optimize or expand functionality outside this scope.

## Execution autonomy

You have full delegated engineering authority within the existing signed POC boundaries.

Resolve routine infrastructure and implementation obstacles autonomously.

Use a Docker-capable authorized host. If the current environment cannot provide Docker access, make one bounded capability assessment and report the exact execution blocker. Do not fabricate results or waste time repeating inaccessible commands.

Do not repeatedly ask for routine approval.

Preserve all existing uncommitted work. Do not commit, push or merge.

## Final definition of completion

The POC is complete when:

1. A complete synthetic monthly payroll has actually been calculated.
2. PostgreSQL's business-integrity value has been demonstrated.
3. DBOS's durable-execution value has been demonstrated.
4. immudb's independent evidence value has been demonstrated.
5. All three have actual execution evidence.
6. Their incremental engineering value and operational costs have been assessed.
7. The final two reports have been produced.

If one experiment cannot be executed because of an external environment limitation, report the exact limitation and retain the POC as incomplete. Do not substitute compilation or theoretical reasoning for execution.

**Final instruction: Three experiments. One monthly payroll. One final assessment. Then stop. We are done with this POC.**

My recommendation is to keep the final demonstration extremely focused: one payroll result, three deliberately induced architectural problems, and three measurable solutions. That will make the value of the architecture much easier to understand than a sprawling payroll application.