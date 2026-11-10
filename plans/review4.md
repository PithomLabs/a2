I have reviewed the complete 512-line Prompt 14 report. The conclusion is clear: Milestone 1.3 passed its integrity objectives, but WorkforceOps is not yet authorized to calculate payroll.

This is an important distinction. The latest review has identified a new category of problems: not primarily whether attendance and compensation records are trustworthy, but whether a payroll run can freeze, govern and reproduce the exact inputs used to calculate money.

READY WITH PREREQUISITES — NOT READY AS-IS

## 1. What the review established

The three principal questions received these answers:

| Question                                                    | Finding                                                                                            |
| ----------------------------------------------------------- | -------------------------------------------------------------------------------------------------- |
| Can direct-SQL dispositions invalidate approved attendance? | No barrier required; live interleaving tests established that approved evidence remains unchanged. |
| Is SHG ready?                                               | No. Its instruction and verification lifecycle is incomplete.                                      |
| Can payroll runs freeze a reproducible input set?           | No. Runs remain draft-oriented and do not capture immutable inputs or rule versions.               |



pre_milestone2_adversarial_review\.md

pre_milestone2_adversarial_review\.md

pre_milestone2_adversarial_review\.md



There is considerable good news: employer population logic, monetary precision, the compensation barrier, database privileges, provisioning and correction-group integrity were reconfirmed as clean in the examined paths.&#x20;

pre_milestone2_adversarial_review\.md



## 2. The six actual blockers

I would treat these as one connected development milestone: Payroll Governance and Reproducibility Foundation.

| Priority | Requirement                      | Problem to solve                                                                      |
| -------- | -------------------------------- | ------------------------------------------------------------------------------------- |
| P1       | Governed payroll-run lifecycle   | Draft runs can be approved directly through SQL without validation.                   |
| P2       | Immutable per-run snapshot       | No frozen record of exactly which employee inputs were used.                          |
| P3       | Effective compensation resolver  | Original and successor compensation records lack a single effective-value resolver.   |
| P4       | SHG instruction lifecycle        | Missing, pending, applicable and verified-none states are not reliably distinguished. |
| P5       | Controlled rule-pack activation  | The application role can promote an unverified or invented rule-pack version.         |
| P6       | Regression and integration tests | The five preceding controls need independent tests and persisted-state verification.  |

These are all explicitly identified as prerequisites in the review.&#x20;

pre_milestone2_adversarial_review\.md



### Why these matter

P1 — Payroll-run lifecycle

Currently, a run can move from draft to approved through direct SQL without completing a validation assessment. Validated runs can also be reopened.

We need a controlled lifecycle:

`Draft → Validated → Approved → Finalized`

The database must enforce the permitted transitions, not just the Go service.

P2 — Immutable payroll snapshot

This is the most consequential architectural gap.

The current input hash is only based on employer, period and run kind. Two different payroll runs for the same period can therefore have the same hash despite different underlying inputs.

Readiness also queries current records each time rather than preserving the precise versions used previously.

We need a real snapshot containing the exact employee population, employment terms, compensation, approved timesheet versions, SHG evidence, statutory rule-pack versions and calculation version.&#x20;

pre_milestone2_adversarial_review\.md



P3 — Compensation resolver

A corrected allowance might have an original value of S$100 and a successor of S$120.

The database preserves both records correctly, but the payroll application still needs a deterministic way to determine which value applies to a particular employee and period.

It must exclude superseded records when resolving the effective compensation basis.

P4 — SHG lifecycle

The current implementation has an especially revealing problem: `verified_none` cannot be recorded through the normal service path, while readiness treats different instruction statuses as equivalent.

We must distinguish missing information from verified non-applicability.

The review also found that direct SQL can alter or delete SHG instructions after approval without triggering an audit or readiness conflict.&#x20;

pre_milestone2_adversarial_review\.md



P5 — Rule-pack activation

This is a critical governance issue.

The application role can directly change a blocked rule pack to `IMPLEMENTABLE` with an invented version.

That means a payroll readiness result could become positive without the statutory research, formalization and independent review having actually been completed.

Activation needs a controlled, audited transition, with a pinned version and review evidence.

P6 — Regression tests

The new lifecycle, snapshots, compensation resolver, SHG behavior and activation safeguards must be tested together.

## 3. One finding I would not misclassify

The review established that direct-SQL dispositions do not need to acquire the employee payroll-input barrier for the tested workflow. Dispositions can remove certain duplicate-event conflicts, but cannot alter already approved evidence or create additional payable hours in the tested cases.

There are two remaining P2 concerns: a consumed-event guard inconsistency and missing audit records on direct-SQL dispositions. These are explicitly deferred, not silently treated as solved.&#x20;

pre_milestone2_adversarial_review\.md



I agree with the review's decision not to widen the transaction barrier unnecessarily.

This is a good example of why we should fix demonstrated integrity failures rather than add locks to every operation.

## 4. My recommended execution sequence

The six prerequisites have dependencies. I would not implement them in the order they happen to appear in the report.

1. P3 — Effective compensation resolver

   Establish one deterministic effective compensation basis.
2. P4 — SHG lifecycle

   Implement verified-none, status-aware readiness, historical protection and governed correction.
3. P5 — Rule-pack activation

   Remove direct application-role activation and establish controlled, audited version governance.
4. P2 — Immutable payroll snapshot

   Capture employee population and all required input versions atomically.
5. P1 — Payroll-run lifecycle

   Build validation, approval and finalization around the frozen snapshot.
6. P6 — Independent regression suite

   Verify the complete lifecycle and prove reproducibility.

P1's state-transition protections can be developed alongside the earlier work, but its validation and approval path must not become usable until P2 snapshots exist. Likewise, P2 cannot safely pin SHG or rule-pack versions until P4 and P5 establish those records.

## 5. What I would deliberately defer

We should not allow this milestone to grow beyond the first payroll scenario.

The review explicitly allows deferring:

- Cross-fund SHG rules requiring additional regulatory research.
- Extended SHG blocker taxonomy.
- Generic disposition auditing and consumed-event hardening.
- Docker/Compose verification where the daemon is unavailable.
- Generic API pagination.
- Rotation materialization.
- Multi-country implementation.
- General-purpose workflow or rules engines.

pre_milestone2_adversarial_review\.md

The first supported employee profile should remain deliberately restricted to a Singapore Citizen, monthly-rated employee with a full salary period, no overtime, no unpaid leave and no variable earnings.

The applicable statutory inputs must still be independently verified.

## 6. My overall assessment

We have reached a meaningful transition.

Milestones 1–1.3 concentrated on building and protecting the operational evidence chain. This review establishes that the next problem is different: how to freeze a coherent financial decision and make it reproducible.

That requires three distinct capabilities:

1. The business inputs must be trustworthy.
2. The payroll run must preserve precisely which versions of those inputs it used.
3. The rule versions used to calculate the result must be governed and independently verified.

Without all three, a payroll result may look correct today but become impossible to reproduce or explain later.

My recommendation is to authorize the six prerequisites as one bounded implementation milestone, without adding statutory calculation code.

Once those six are implemented and independently verified, we should finally begin the first real Singapore payroll calculation. No further general architecture review is necessary unless execution reveals a genuine structural blocker.