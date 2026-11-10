I would make this the final independent red-team review of the post-Prompt8 implementation, with one additional responsibility: verify that the technical article accurately represents what the code actually accomplishes.

The reviewer must not assume that 14/14 passing conditions mean the architecture is defect-free. It should challenge the implementation, reproduce important attacks, inspect the real transaction boundaries and identify claims that exceed the evidence.

# PROMPT9 — Independent Adversarial Code Review of the Final PostgreSQL + DBOS + immudb POC

## 1. Mission

Act as an independent Principal Software Architect, PostgreSQL Security Engineer, Go Distributed Systems Engineer, Cryptographic Systems Reviewer and Technical Due-Diligence Auditor.

Conduct a final, rigorous adversarial review of the post-Prompt8 implementation.

The objective is to determine whether the completed POC genuinely demonstrates the architectural responsibilities and guarantees attributed to:

- PostgreSQL 18.6 — authoritative business data and integrity.
- DBOS Transact Go v1.8.0 — durable business-process execution.
- Go — business-domain logic and payroll calculation.
- immudb v1.11.2 — cryptographically verifiable historical evidence.
- Independent auditor — reconstruction and verification of actual business evidence.

Do not assume that the reported 14/14 acceptance results establish correctness. Independently inspect the implementation and verify the most important guarantees.

**Your job is to try to disprove the final architecture, not defend it.**

The final goal is to establish whether the system is a credible architectural POC that software engineers and business owners can examine and learn from.

This is not another implementation phase.

---

## 2. Authoritative baseline

Begin by inspecting:

- `AGENTS.md`
- `plans2/POC-REMEDIATION-REPORT.md`
- `plans2/POC-FINAL-REPORT.md`
- `plans2/POC-RESULTS.md`
- `poc/payroll-month/POC-REPORT.md`
- `poc/payroll-month/POC-RUNBOOK.md`
- `plans2/PHASE2-WORKLOAD.md`
- `plans2/P8-PROD-MIGRATION-PROPOSAL.md`
- Relevant ADRs and decision records.
- All Go source under `poc/payroll-month/`.
- All POC SQL, initialization scripts, Compose configuration and tests.

The remediation report is the authoritative post-Prompt8 result. Earlier failures in the historical reports must remain identifiable as superseded results, not current implementation status.

The current claimed execution baseline is:

- PostgreSQL 18.6.
- DBOS Transact Go v1.8.0.
- immudb v1.11.2.
- pgx v5.12.0.

The final October 2026 workload reports:

- 25 synthetic employees.
- 102 payroll lines.
- 24 payslips.
- 24 finalized employees.
- One blocked employee.
- Five exceptions.
- Net payroll of 8,521,857 cents.
- Canonical and DB-backed digest:

`4beee632a05c09b9ec0be0dbedc1937d517461603d8660999d66f51d04b20377`

Reported final verification:

- Fourteen remediation conditions passed.
- `go test ./...` exited zero.
- `go vet` clean.
- `gofmt` clean.
- PostgreSQL 16 data untouched.
- Migrations 00001–00012 unchanged.
- No migration 00013.
- No production-readiness claim.

These are claims to verify independently, not facts to accept without examination.

---

## 3. Review boundaries

The POC is intentionally limited.

It is not a full payroll product and must not be evaluated as one.

Do not demand production functionality that the POC explicitly excludes.

The following constraints remain binding:

1. Synthetic payroll data only.
2. SG-MOM-OT-001 remains BLOCKED.
3. No statutory payroll calculations.
4. Frozen migrations 00001–00012 must remain unchanged.
5. No migration 00013.
6. No production deployment.
7. immudb BSL-1.1 commercial licensing remains subject to counsel review.
8. Production ledger hardening remains a separate promotion concern.
9. The conventional PostgreSQL + Go worker baseline remains NOT RUN.
10. No commits, pushes, merges or destructive operations against existing databases.

Preserve the existing `stack-poc` branch and all uncommitted work.

Do not repair defects during the review.

Do not add features, introduce new infrastructure or expand the payroll domain.

You may execute focused tests and non-destructive experiments against disposable POC environments.

---

# 4. PostgreSQL: independently challenge the authority model

The central claim is that PostgreSQL is the authoritative business-data layer.

Determine whether this is genuinely true.

## A. Application-role authority

Inspect `pg_authority.sql` and every relevant SECURITY DEFINER function.

Review:

- Function ownership.
- `EXECUTE` privileges.
- Secure `search_path`.
- Schema qualification.
- Role memberships.
- Direct INSERT/UPDATE/DELETE privileges.
- Trigger ownership and behavior.
- Transition guards.
- Function input validation.

Try to establish whether the `app` role can still:

- Fabricate a payroll snapshot.
- Fabricate snapshot members.
- Insert arbitrary rule pins.
- Manufacture validation evidence.
- Approve an unauthorized result.
- Promote a fabricated payroll to validated or finalized.
- Bypass the authorized capture function.
- Manufacture or suppress outbox events.
- Forge ledger back-links.

Test with actual application credentials.

Also inspect what happens if a SECURITY DEFINER function receives malicious or inconsistent identifiers.

A correctly restricted function must not become a privileged confused deputy.

## B. Capture and employer gate

Trace the real production POC call path.

Verify that:

- Actual capture invokes the authorized database operation.
- The employer gate is acquired inside the authoritative transaction.
- Approved inputs are resolved under the same consistency boundary.
- Claims and coverage verification operate on actual records.
- Snapshot creation is atomic.
- DBOS durability records commit with the relevant application writes.

Test concurrent capture and writer operations for the same employer.

Also test two different employers to establish that isolation is appropriately scoped.

Review lock ordering, transaction isolation and potential deadlocks.

## C. Period closure

Inspect the real closing path.

Verify:

- Only the authorized operation can create a close record.
- Application-role direct close fabrication fails.
- Closed-period writes are rejected using actual schema relationships.
- A competing write and close cannot both silently succeed in an invalid order.
- The close timestamp is database-controlled.
- The original payroll result remains unchanged afterward.

## D. Claims and coverage

Inspect the actual verifier and its callers.

Challenge whether it can detect:

- Missing expected membership.
- Unexpected extra membership.
- Duplicate source records.
- Invalid employee membership.
- Uncovered effective periods.
- Invalid rule-pack references.
- An empty or vacuous claim set.
- A stale or inconsistent source snapshot.

Verify that it fails closed.

Check that all claim evidence is derived through authorized operations rather than arbitrary application-provided assertions.

## E. Temporal correctness

Inspect the actual PostgreSQL range and temporal mechanisms.

Do not assume unsupported PostgreSQL PERIOD foreign keys exist.

Determine what is enforced through:

- Range constraints.
- Exclusion constraints.
- Triggers.
- Procedural coverage verification.
- Transactional locking.

Distinguish proven no-overlap enforcement from complete temporal coverage.

Do not overstate the capabilities demonstrated by PostgreSQL 18.6.

---

# 5. DBOS: independently challenge durable execution

The central claim is that DBOS durably coordinates the actual payroll business process.

Trace Stages A–G in `dbosflow/payroll.go` and all supporting implementation.

## A. Real business execution

Verify that the workflow:

1. Establishes payroll readiness.
2. Captures actual authoritative records.
3. Reloads the persisted snapshot.
4. Executes the actual Go calculator.
5. Persists employee-level results.
6. Performs the required approval wait.
7. Validates and finalizes.
8. Enqueues real evidence.

Confirm that no phase is merely a status-label operation or placeholder.

The calculation must be based on persisted snapshot data, not regenerated fixture data.

The persisted payroll output must match the independently reconstructed result.

## B. Transaction boundaries

Inspect every `RunAsTransaction` call.

Verify that:

- The correct shared PostgreSQL pool is used.
- Database transactions are propagated through all adapters.
- No operation silently obtains a separate pool connection.
- The business writes and DBOS durability record commit together.
- Rollback is correct on failure.
- No external immudb operation is mistakenly treated as part of the PostgreSQL atomic transaction.

Identify the exact business operations covered by each atomic boundary.

## C. Business identity and workflow identity

Inspect:

- Business payroll-run identity.
- Input digest identity.
- Workflow attempt identity.
- DBOS reuse policies.
- Reject-on-reuse.
- Adopt-or-conflict behavior.
- Concurrent submissions.
- Error retry.
- Recovery.

Challenge the following cases:

- Identical input submitted twice.
- Changed input submitted under the same business run.
- A transient failure followed by retry.
- Two concurrent attempts for the same payroll run.
- A finalized run submitted again.
- A forged SUCCESS record under the intended workflow ID.

Verify that no stale payroll result is silently returned and no corrected input is silently discarded.

## D. Durable approval

Verify the actual approval mechanism using a real DBOS wait.

Execute or inspect the interruption and recovery path.

Confirm that approval:

- Is recorded in PostgreSQL.
- Is associated with the correct payroll run.
- Survives process restart.
- Cannot be manufactured through a forged DBOS status record.
- Cannot authorize a different payroll input digest.

## E. Exactly-once semantics

Be precise.

Distinguish:

- DBOS workflow deduplication.
- Atomic PostgreSQL capture.
- Atomic PostgreSQL result persistence.
- External at-least-once delivery.
- Idempotent immudb append.
- Unique financial artifact creation.

Verify actual persisted payroll records.

Do not use workflow-status counts as the sole evidence of exactly-once business effects.

## F. Operational limitations

Inspect:

- DBOS Shutdown behavior.
- Connection-pool lifecycle.
- The ClosePoolTimeout workaround.
- Workflow logging.
- Standard-output contamination.
- Cancellation handling.
- Worker restart.
- Recovery latency.

Determine whether the documented workaround is safe for repeated workflow executions and clean process termination.

---

# 6. immudb: independently challenge evidence integrity

The central claim is that immudb provides independent evidence of actual payroll artifacts.

Inspect the native SDK integration, the outbox worker, the auditor and the genesis procedure.

## A. Actual content binding

Verify that the actual evidence payload contains correct references and digests for:

- Payroll input snapshot.
- Finalized payroll output.
- Rule-pack identity and digest.
- Calculator version.
- Approval.
- Finalization.
- Event identity.
- Correction lineage.

Independently reconstruct at least one complete payroll evidence event from PostgreSQL.

Check that the reconstructed content exactly matches the canonical bytes and the immudb record.

Do not accept a version label as an implementation digest.

The current calculator digest intentionally hashes a version label rather than actual implementation bytes. Confirm that all documentation accurately describes this limitation.

## B. Independent auditor

Verify that the auditor:

- Uses an independent identity.
- Has read-only PostgreSQL access.
- Has read-only immudb access.
- Uses its own trusted state.
- Enforces the pinned genesis identity and content hash.
- Refuses unexpected server identity changes.
- Reconstructs evidence directly from PostgreSQL.
- Re-executes the relevant payroll calculation.
- Checks event identity and payload binding.
- Checks input and output digests.
- Checks rule-pack provenance.
- Checks approvals and corrections.
- Fails loudly on content mismatch.

Challenge whether the auditor trusts a supplied file or reconstructs the actual authoritative records independently.

## C. Adversarial tamper tests

Verify the reported 9-case tamper battery.

Independently select representative cases:

- Modify an authoritative payroll amount.
- Modify a stored employee result.
- Change an input snapshot.
- Change a rule digest.
- Swap transaction references.
- Forge a PostgreSQL backlink.
- Remove an expected evidence event.
- Modify correction lineage.
- Attempt to replace or substitute genesis.

Verify that each operation is detected by the appropriate check.

Distinguish PostgreSQL write prevention from immudb evidence verification.

## D. Genesis and host corruption

Inspect:

- `gate/immudb-genesis.json`.
- The ceremony implementation.
- The auditor pin enforcement.
- Canary checks.
- The host-corruption harness.

Verify that trust is established against an out-of-band committed anchor rather than silently bootstrapped from whichever server answers first.

Inspect the actual IM-8 execution evidence.

Determine precisely what the four-cell corruption experiment proved.

Do not overstate host-admin protection beyond the threat model.

## E. Recovery and correction lineage

Inspect verified reads during retries.

Verify that a recovered transaction is bound to the exact expected event and content.

Test duplicate delivery and missing-backlink recovery.

Verify that correction events receive new identities and that original evidence remains verifiable.

Check that poison accounting cannot silently orphan a proven append.

---

# 7. Cross-layer architecture review

This is the most important architectural section.

Follow one payroll operation from its original inputs through final verification.

Trace the complete sequence:

PostgreSQL authoritative records

→ authorized database capture

→ DBOS transactional checkpoint

→ persisted input snapshot

→ Go payroll calculation

→ PostgreSQL result persistence

→ actual approval

→ finalization

→ transactional outbox

→ DBOS-controlled durable delivery

→ immudb append

→ verified PostgreSQL backlink

→ independent auditor reconstruction

→ verified payroll evidence.

For every boundary, establish:

- What is authoritative?
- Which component is responsible?
- What transaction protects it?
- What happens if the process crashes?
- What happens if the same operation is retried?
- How is duplication prevented?
- How is an invalid transition rejected?
- What evidence remains afterward?

Look for mismatches between what a layer claims and what it actually controls.

The architecture should not depend on undocumented sequencing assumptions or on one component trusting another component's success label.

---

# 8. Review the technical writeup

The associated article is titled:

**When Payroll Fails, Who Remembers What Happened?**

Its purpose is to explain the lessons from this experiment to software engineers and business owners.

Review the technical claims made in the article against the current implementation.

In particular, verify:

1. PostgreSQL governs authoritative business state.
2. DBOS supplies durable memory for long-running business processes.
3. PostgreSQL and DBOS share the appropriate atomic transaction boundary.
4. The actual payroll calculation runs from captured database records.
5. immudb anchors actual evidence rather than arbitrary audit messages.
6. The independent auditor reconstructs and verifies actual content.
7. The October payroll figures and digest are correct.
8. The crash-during-approval example reflects the actual executed test.
9. The original finalized result remains unchanged after corrections.
10. The article accurately distinguishes ledger immutability from calculation correctness.

Flag:

- Statements that overstate exactly-once guarantees.
- Statements suggesting immudb proves the correctness of business calculations.
- Statements implying complete protection against privileged host compromise.
- Statements suggesting production readiness.
- Statements that omit known material limitations.
- Comparisons that claim measured engineering savings when no conventional baseline was executed.

Do not rewrite the article.

Provide an article-claim verification table with PASS, QUALIFY, UNSUPPORTED or CONTRADICTED.

The intention is to preserve the article's accessible narrative while ensuring that every technical claim is defensible.

---

# 9. Known limitations: verify, do not silently erase

The remediation report identifies several deliberate limitations.

Verify that they still accurately describe the code:

1. Calculator digest hashes a version label, not implementation bytes.
2. Multi-pack joint digest is future work.
3. Payroll period is hardcoded to October 2026 in the POC workflow and auditor.
4. Cross-version JSONB text canonicalization is not established.
5. Supersedes ordering lacks a sub-microsecond tie-breaker.
6. Production immudb TLS, signing keys, retention, monitoring and stable-volume hardening remain unimplemented.
7. BSL-1.1 counsel review remains pending.
8. Conventional-job baseline remains unmeasured.
9. DBOS Shutdown pool issue remains subject to the documented workaround.
10. Some operational writers remain gate-only without full period mapping.
11. Host-corruption verification is bounded by the tested server-side threat model.
12. Back-link and audit-history retention require future operational policy.

Classify these accurately.

A documented limitation is not automatically a defect in the POC.

However, if the implementation has additional risks not included in the report, document them.

---

# 10. Severity and reporting discipline

Use:

**P0 — Critical:** Incorrect financial calculation, unauthorized payroll authority, broken atomic capture, undetected evidence alteration.

**P1 — High:** Important security bypass, concurrency defect, duplicate financial effects, stale-result acceptance or broken recovery.

**P2 — Medium:** Incomplete guarantee, operational weakness, reproducibility limitation or misleading architectural claim.

**P3 — Low:** Maintainability, observability, clarity or minor test issues.

Every finding must include:

- Unique identifier.
- Severity.
- Exact file and line reference.
- Relevant function or SQL operation.
- Attack or failure scenario.
- Expected invariant.
- Actual behavior.
- Reproduction steps.
- Evidence.
- Business impact.
- Whether the issue affects the completed POC, technical writeup, or future production readiness.

Do not generate speculative findings merely to increase the finding count.

Report only concrete, supportable findings.

---

# 11. Deliverable

Create one new file:

`plans2/POC-FINAL-ADVERSARIAL-REVIEW.md`

Use this structure:

1. Executive verdict.
2. Scope and review methodology.
3. Post-Prompt8 architecture summary.
4. PostgreSQL findings.
5. DBOS findings.
6. immudb findings.
7. Cross-layer transaction and recovery analysis.
8. October payroll reconstruction verification.
9. Article technical-claim verification.
10. Known limitations and additional risks.
11. Findings register by severity.
12. Final architecture assessment.

Include a summary:

| Component             | Architectural responsibility | Independently verified | Residual risk | Verdict |
| --------------------- | ---------------------------- | ---------------------- | ------------- | ------- |
| PostgreSQL            | Business authority           | ...                    | ...           | ...     |
| DBOS                  | Durable execution            | ...                    | ...           | ...     |
| Go                    | Calculation and domain logic | ...                    | ...           | ...     |
| immudb                | Evidence integrity           | ...                    | ...           | ...     |
| Independent auditor   | Content verification         | ...                    | ...           | ...     |
| Combined architecture | End-to-end guarantees        | ...                    | ...           | ...     |

For every important guarantee, classify it as:

- PROVEN.
- PARTIALLY PROVEN.
- NOT PROVEN.
- CONTRADICTED.

---

# 12. Execution boundaries

This is a read-only adversarial review.

Do not modify implementation source code.

Do not modify SQL or migrations.

Do not modify existing reports or the article.

Do not introduce new features.

Do not change the signed POC scope.

Do not create additional experiments beyond those needed for safe independent verification.

Use disposable databases and isolated test fixtures.

Do not damage the canonical payroll result or the live immutable ledger.

Do not commit, push or merge.

The only new deliverable is `plans2/POC-FINAL-ADVERSARIAL-REVIEW.md`.

---

# Final directive

The remediation report says the architecture now works.

Your job is to establish whether that statement survives independent technical scrutiny.

Do not reward passing tests when they fail to exercise the intended business behavior.

Do not confuse a secure component with a secure system.

Do not confuse an immutable record with a correct calculation.

Do not confuse workflow completion with business correctness.

Follow the actual data.

Follow the actual transactions.

Follow the actual authority boundaries.

Follow the actual evidence.

**If the post-Prompt8 implementation withstands this review, we can confidently preserve the POC as a credible technical demonstration. If it does not, identify precisely why, without turning this review into another open-ended development programme.**

My recommendation: Keep this review read-only. Prompt8 has already completed the root-cause remediation. The purpose now is independent verification and determining whether the technical article is defensible—not starting another round of implementation.