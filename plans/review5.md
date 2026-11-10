This review is a decisive setback for Milestone 2 acceptance, but it is also one of the most valuable reviews we've conducted. It found defects that ordinary green Go tests would not have exposed.

My verdict: I agree with the independent review. Milestone 2 is FAIL — NOT READY for payroll calculation. Do not proceed to CPF, SDL or SHG calculation yet.

The fundamental problem is that the application can construct a plausible payroll snapshot without guaranteeing that it contains all the business information committed before the snapshot's synchronization boundary.

There are two P0 defects, several blocking P1 defects, and a relatively well-defined set of corrections. We do not need another architectural redesign.
1. The two critical failures
P0 — D1
Snapshot isolation can produce a stale payroll population.

PostgreSQL establishes the REPEATABLE READ MVCC view when the transaction executes its first relevant statement. In the current implementation, the transaction can establish that view before it finishes waiting for an advisory lock.

A writer commits while the snapshot transaction is waiting. The reader then acquires the lock but continues using its earlier view.

Live reproduction: a newly created employee and other committed inputs were missing, yet validation returned valid=true and the run became validated.


P0 — EF-E3
The app role can fabricate or extend payroll snapshots.

Direct SQL under app can insert a fabricated snapshot header and child records, including employee and rule-pack records, without changing the original content hash.

Live reproduction: new employee and pack rows were inserted into a standing valid snapshot; its SHA-256 remained unchanged.


These are different failures.

The first means the snapshot may omit legitimate employees or inputs.

The second means the snapshot can contain invented or unauthorized inputs.

Together, they undermine the entire premise of payroll reproducibility. 
milestone2_adversarial_review.md
 milestone2_adversarial_review.md

2. The most important technical discovery: the population barrier is ordered incorrectly
The report establishes that merely using advisory locks inside a REPEATABLE READ transaction is insufficient.

The required correction is:
1. Acquire dedicated PostgreSQL connection
             |
             v
2. Acquire session-level population barrier
             |
             v
3. Acquire required session-level pack barriers
             |
             v
4. BEGIN ISOLATION LEVEL REPEATABLE READ
             |
             v
5. Acquire employee locks in sorted order
             |
             v
6. Read and revalidate employee population
             |
             v
7. Capture compensation, attendance, SHG and rule pins
             |
             v
8. Persist snapshot and calculate content hash
             |
             v
9. COMMIT
             |
             v
10. Release session-level locks


The critical difference is that the snapshot's database view is established only after the necessary synchronization barriers have been acquired.

The review actually tested this: lock acquisition before beginning the repeatable-read transaction saw the writer's committed changes, whereas acquisition inside the transaction did not. 
milestone2_adversarial_review.md

One additional implementation warning: session-level PostgreSQL advisory locks survive transaction completion. The Go implementation must explicitly release them on every success and failure path, using the same physical connection, before returning that connection to the pool. Otherwise, connection pooling can introduce a new locking defect.
3. Seven additional barrier gaps
The review found that several input-changing operations do not participate in the appropriate synchronization protocol.
Operation	Required correction
Timesheet approval	Acquire employer population lock as well as employee lock
Employment-term closure	Acquire employee input lock after population lock
Clock-event disposition	Acquire employee input lock
Direct status-history insertion	Revoke direct INSERT or use a protected trigger/function
ensure_pack_blocked	Participate in rule-pack synchronization
Draft timesheet-line mutation	Acquire employee input lock or restrict to controlled service operations
Employer/pay-period changes	Revoke unrestricted updates or provide synchronized mutations and drift checks

These are not seven independent architecture problems. They are incomplete implementations of one shared consistency contract.

The fix must ensure that every operation affecting a payroll snapshot's meaning participates in the appropriate barrier. 
milestone2_adversarial_review.md

4. Compensation resolver: three blocking defects
The single-resolver architecture is sound, but the review found three problems in its actual behavior.
Finding	Problem	Required correction
F-A1-WIDEN	Widening a compensation correction can permanently break chain resolution	Load the complete supersession ancestry independently of the period-overlap query
F-A1-FREETAIL	Narrowing a predecessor and adding a new compensation can create permanently overlapping effective values	Align write-time overlap rules with the resolver's surviving effective spans
F-A1-MIDPERIOD	A monthly salary change halfway through a month is silently accepted as two segments	Return COMP_MID_PERIOD_CHANGE until proration semantics are implemented

The third problem is particularly important.

The first payroll scenario assumes a full salary period with a stable monthly salary. We should not silently accept two different basic-salary amounts during that period until the treatment is independently researched and implemented.

The snapshot, readiness and validation layers must agree on that restriction. 
milestone2_adversarial_review.md

5. SHG: the missing distinctions must be resolved
The SHG implementation has good underlying mechanics, but the review found three important semantic gaps.

Active versus verified. An active instruction currently counts as complete even though the system has no clear verification path for that state. The project must decide whether an active declaration itself is sufficient evidence or whether it requires explicit verification.

Period coverage. An instruction beginning halfway through a month currently can count as coverage for the entire month. The system must verify full-period coverage or explicitly identify the uncovered dates.

Cross-fund combinations. Multiple funds can currently coexist, and readiness accepts the first qualifying instruction while ignoring ambiguity in the others.

For our restricted first payroll profile, I recommend the conservative policy: block cross-fund ambiguity until the relevant eligibility and interaction rules have been established.

The report also identifies incomplete negative-test evidence for SHG audit behavior. 
milestone2_adversarial_review.md

6. Rule packs: we must protect the trust root
This is another important discovery.

We have established the rule-pack registry and activation mechanism, but the independent review found that some claimed verification steps are weaker than their names suggest.

Three issues matter:
- Source verification: manifest digests are checked for format and source-list presence, but actual official-source bytes are not rehashed.
- Golden verification: the CLI can label fixtures GOLDENS_VERIFIED without actually executing or independently verifying them.
- Registry repinning: migration_admin can change the reviewed version and digest through raw SQL without an audit record.
There are also evidence-label mismatches and an unresolved rule-pack effective-period coverage check.

I would treat a fabricated source digest or a failing golden being recorded as verified as a serious governance failure, even though the current real packs remain blocked. 
milestone2_adversarial_review.md

The positive finding is that all ten registered identities remain unpinned and BLOCKED, with no real statutory pack activated. That has prevented these weaknesses from producing a false claim of payroll compliance. 
milestone2_adversarial_review.md

7. Direct SQL can also bypass payroll-run governance
The implementation report previously described the direct draft → validated transition as outcome-equivalent to the service workflow.

The latest review disproved that characterization.

As app, direct SQL can:
- Validate a run using placeholder hashes.
- Supply empty rule-version pins.
- Bypass the validation offer.
- Bypass validation audit.
- Revalidate with fabricated snapshot content.
- Change validation pins after the transition.
- Validate a fabricated snapshot containing zero employees.
The database lifecycle trigger checks some conditions, but not the full governance contract.

The correction should be:
1. Revoke direct application-role updates to run status, input hash and rule-version pins.
2. Move validation into a narrowly scoped SECURITY DEFINER function.
3. Have that function verify the actual current snapshot tip.
4. Verify the hash and pins against that tip.
5. Require a non-empty employee population.
6. Verify terminal inclusion/exclusion decisions and rule-pack eligibility.
7. Stamp the validation actor and audit record atomically.
This is a P1, but it is a prerequisite to any trustworthy validated payroll run. 
milestone2_adversarial_review.md

8. What is actually good and should be preserved
I do not want the latest FAIL to obscure the substantial amount of work that is correct.

The review independently reconfirmed:
- The single effective compensation resolver is architecturally sound.
- Compensation does not double-count ordinary predecessor/successor records in the tested cases.
- SHG's missing, pending, verified-none and corrupt states have useful underlying protections.
- Rule-pack activation is protected against ordinary app role writes.
- Snapshot canonical hashing is deterministic.
- Snapshot UPDATE/DELETE protections work in tested paths.
- Snapshot invalidation and revalidation preserve historical contents.
- The payroll-run correction chain and singleton constraints work in tested cases.
- The full 19-package test battery passes across three timezone settings.
- The race detector is clean for the five tested packages.
The problem is that the live PostgreSQL attacks found weaknesses in the boundaries not adequately exercised by the existing test suite. 
milestone2_adversarial_review.md


9. My recommended next milestone
Do not proceed directly to financial calculation.

I recommend a tightly bounded follow-on milestone:
Milestone 2.1 — Payroll Snapshot and Trust Integrity Closure
A. Fix snapshot isolation and all seven barrier gaps.

Session-level pre-BEGIN synchronization, complete writer participation, and live writer-first tests.

B. Protect snapshot and validation entry points.

Controlled snapshot capture, controlled validation, non-empty population, pinned hashes and atomic audit.

C. Fix compensation resolution.

Widening, freed-tail behavior and the explicit mid-period salary blocker.

D. Fix SHG coverage semantics.

Active verification decision, full-period coverage and cross-fund ambiguity.

E. Strengthen rule-pack evidence governance.

Rehash actual sources, execute or verify golden evidence, audit registry pinning and validate effective-period coverage.

F. Complete independent verification.

Live PostgreSQL tests, direct SQL as app, writer-first races, fresh-server provisioning and CLI/REST differential testing.

The report provides this exact minimum prerequisite breakdown in §16. 
milestone2_adversarial_review.md

10. What we should defer
The report explicitly identifies payroll_line as harmless today because no production calculation path uses it, but a mandatory prerequisite on the day a calculator starts writing financial lines.

It needs:
- A direct relationship to the frozen snapshot.
- Verified line-level provenance.
- Exact monetary precision.
- Post-calculation immutability.
- Versioned correction/supersession.
- Revocation of direct application-role writes.
Do not introduce a payroll-line framework now; schedule these controls as part of the first calculation implementation. 
milestone2_adversarial_review.md

Other reporting cleanup, symlink hardening, audit wording and minor migration-test improvements can remain prioritized below the actual financial-input blockers.

11. Final assessment
Area	Verdict
Compensation and employee-input integrity	FAIL
SHG and rule-pack governance	FAIL
Snapshot and payroll-run integrity	FAIL
CLI/API operability	PASS WITH CONDITIONS
Overall Milestone 2	FAIL

milestone2_adversarial_review.md
My strongest recommendation is to treat the snapshot-isolation defect as the central issue of the next implementation pass.

It exposes an important engineering lesson: even if every individual record is valid, a database transaction can still assemble an incomplete business picture if synchronization and MVCC snapshot creation are ordered incorrectly.

The second lesson is that a hash does not make a snapshot immutable. The database must protect the content that the hash represents.

Once those two problems, compensation resolution, SHG coverage and rule-pack trust are fixed and independently reverified, we should finally close the infrastructure phase.

The first financial calculation should then be a restricted, full-period Singapore monthly payroll scenario with pinned, independently verified CPF, SDL and applicable SHG rules.

We are close to the calculation stage, but this review has correctly stopped us from turning a technically reproducible calculation into a financially unreliable one.