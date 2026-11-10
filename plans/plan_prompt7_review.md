I reviewed \`plan_prompt7.md. My verdict: this is the right implementation milestone. Approve it with two important corrections and one smaller clarification. I would not send the agent back for another architecture review.

The plan has a sensible dependency order: HR → BPO → rosters → attendance → timesheets → payroll readiness. It also reuses the existing PostgreSQL foundation and shares application services between CLI and REST. \<FileCite ref_id="turn87file0" line_range_start={12} line_range_end={36}/>

## 1. P0 — Actual breaks cannot be derived from the proposed CSV

This is the most important omission.

The attendance import currently describes only two clock-event types:

```
in
out
```

Yet the plan promises to distinguish:

- Actual work intervals.
- Actual paid breaks.
- Actual unpaid meals.

Two clock events marking the beginning and end of a shift cannot establish when the employee actually took a meal break.

The shift template tells us when a break was scheduled, not when it was actually taken.

Required amendment: support explicit break evidence in the attendance-import contract. Either extend the event model to distinguish break start/end events, with the necessary schema change, or accept explicit interval records with their own source identifiers and timestamps.

When break evidence is missing, create an exception. Never silently turn a planned break into an actual break.

This is important for accurate paid-time attribution and for eventual payroll calculations.

## 2. P0 — Deleting old executables while leaving their Taskfile commands

The plan proposes deleting:

```
cmd/shifthub-api
cmd/payloop-api
cmd/sync-worker
```

That is consistent with our single-executable architecture.

However, it also proposes leaving old Taskfile targets such as `run:shifthub` and other pre-pivot commands untouched because they are considered harmless. \<FileCite ref_id="turn87file0" line_range_start={16} line_range_end={20}/>

They are not necessarily harmless. They may still invoke the executables being deleted.

Require the agent to:

1. Search all references to the three old entry points.
2. Identify Taskfile, scripts, documentation and CI dependencies.
3. Remove obsolete targets or replace them with the new `workforceops` commands.
4. Verify the documented developer workflow after consolidation.

Do not leave broken commands behind in the name of minimizing scope.

## 3. P1 — Make the 1.5× employment assumption explicit

The agent chose to restrict the MVP to employment arrangements paying exactly the statutory minimum 1.5×. That is a reasonable way to limit the first implementation, but MOM states a minimum rate, not an exact contractual rate. \<FileCite ref_id="turn87file0" line_range_start={18} line_range_end={22}/>

The selected reference employer should therefore have an explicit, documented contractual policy establishing 1.5× for the supported employee profile. Other contractual rates should be unsupported rather than silently treated as 1.5×.

No overtime calculation needs to be implemented in this milestone.

## 4. Exact amendments to give the agent

\<WritingBlock id="92841" variant="document"># Prompt7 Review Amendments — Required Before Execution

The overall Prompt7 plan is approved. Do not redesign the architecture or expand the milestone.

Make these three targeted corrections:

## A. Attendance break evidence (P0)

The proposed CSV contains only `in` and `out` events, but the implementation promises separate work, paid-break and unpaid-break intervals.

Clock-in and clock-out alone cannot establish actual break times. Scheduled shift segments must never be treated as proof that a break occurred.

Before implementing attendance import, define one minimal supported input contract that can represent actual break evidence.

Either:

- Extend clock-event kinds to represent break start/end, with the necessary migration and tests; or
- Support explicit, source-identified attendance intervals for work and breaks alongside raw clock events.

Choose the smallest design consistent with the existing schema and lineage guarantees.

Requirements:

- Preserve original event evidence.
- Deduplicate imported records.
- Distinguish actual breaks from scheduled breaks.
- Retain paid/unpaid classifications.
- Missing break evidence must produce an explicit exception.
- Never silently infer an actual break from a roster template.
- A demo-only assumption must be explicitly identified as an assumption and must not be represented as an observed fact.

Update the import format, validation rules, CLI documentation, API contract and tests accordingly.

## B. Single-executable consolidation (P0)

The proposed removal of the three pre-pivot command entry points is accepted.

Before deleting them, search all repository references, including:

- Taskfile targets.
- Shell scripts.
- CI workflows.
- README and AGENTS instructions.
- Tests and build tasks.

Remove or replace obsolete references. Do not leave commands that invoke deleted executables.

Verify the new single-binary developer workflow from a clean build.

## C. Contractual overtime-rate boundary (P1)

The restricted reference employer profile supports only contractual overtime arrangements at exactly 1.5× for this MVP.

Document this as an explicit employer/contract assumption, not as the legal meaning of the statutory phrase "at least 1.5 times".

Higher or otherwise different contractual rates must be flagged unsupported until the later payroll calculation slice models them.

No overtime calculation is authorized in Prompt7.

## Execution instruction

After applying these corrections, execute Prompt7.

Preserve all existing uncommitted work. Keep the existing architecture, database integrity decisions and rule-pack lifecycle unchanged.

Do not introduce further design-review cycles unless an actual implementation blocker is discovered.

Complete the vertical slice, run the specified tests against real PostgreSQL, and stop at the Prompt7 acceptance gate.

Report any unavoidable schema migration and its reason.

## 5. My final recommendation

Proceed with Prompt7 after these amendments.

I would explicitly protect the following boundary:

| Deliverable                       | This milestone |
| --------------------------------- | -------------- |
| Employee administration           | Implement      |
| BPO programme and team management | Implement      |
| Shift templates and rosters       | Implement      |
| Attendance and break evidence     | Implement      |
| Timesheet generation and approval | Implement      |
| Payroll readiness                 | Implement      |
| Overtime calculation              | Deferred       |
| CPF/SDL/SHG calculations          | Deferred       |
| Payslip financial generation      | Deferred       |

One working business workflow is worth more now than another round of theoretical completeness.

After Prompt7, we should have an actual HR/WFM backend that can demonstrate the operational journey and expose exactly what prevents payroll from being completed. That is meaningful progress toward the MVP.