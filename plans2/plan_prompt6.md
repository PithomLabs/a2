# Plan to do plans2/prompt6.md (final three-experiment execution)

Target: `plans2/prompt6.md` (final POC execution), read with `plans2/roadmap.md`
as context. Roadmap sets the objective (demonstrate architectural value, not a
payroll product); prompt6 sets the execution contract (three experiments, one
shared month, two final reports, then stop).

## Goal

Execute prompt6's final POC scope: one shared synthetic month (25 employees ×
31 days) plus three comparative experiments (PG integrity, DBOS durability,
immudb evidence), each with a credible without-technology baseline, producing
the two final reports and the POC status update — or, if the bounded
capability check fails, an honest blocked outcome that retains the POC as
incomplete.

## Success Criteria

- Either (A) all three experiments executed with the evidence prompt6 §3–§5
  demands, comparison matrix filled from measured results, both reports
  written; or (B) blocked outcome with the exact prerequisite named, both
  reports written as NOT RUN/BLOCKED, zero fabricated results.
- Shared workload identical across experiments; all outputs labeled
  illustrative/non-statutory; OT-001 untouched; frozen files, PG16, and
  uncommitted work undisturbed; no commit/merge.

## Approach

prompt6 narrows Gate B (no exhaustive E1–E6/T1–T13 matrices — focused cases
only) but adds genuinely new build scope: the three baselines and the
DB-backed shared fixture do not exist yet. Everything else is reuse (synth
10/10 vectors, poc-run modes, kill/forged/idem harnesses, auditor, gate DDL,
pinned compose). Key decisions:

1. **Deliverable paths:** `plans2/POC-FINAL-REPORT.md` + `plans2/POC-RESULTS.md`
   (standing plans2-only rule); `poc/payroll-month/POC-REPORT.md` updated in
   place (prompt6 §7 names it as the authoritative record — no divergent copy).
2. **`plans2/plan_prompt4.md` erratum:** that file was never persisted
   (inline-only delivery) — execute from prompt4's text + POC-REPORT +
   runbook instead; do not reconstruct or invent it.
3. **Baselines stay minimal:** comparison harnesses + disposable baseline
   tables only — never three alternative payroll apps, never weakened real
   schema (Exp1A), ordinary Go worker + PG state for Exp2A, disposable-clone
   tamper for Exp3A. Baselines must be credible, not strawmen.
4. **Focused-case mapping (no matrix sprawl):** Exp1's 4 cases ← Period-FK/
   coverage checks + gate-race + fabrication + close-race harnesses; Exp2's 6
   crashes ← kill-matrix + double-enqueue + approval-wait; Exp3's 10
   scenarios ← tamper battery + auditor + outbox recovery. Nothing else runs
   "for completeness."
5. **Order:** shared fixture + canonical month first → Exp1 (integrity during
   capture) → Exp2 (lifecycle incl. approval/finalize) → Exp3 (evidence on
   the finalized result, per §5). Exp3 depends on Exp2's finalized output.
6. **Capability gate:** docker access re-probed once at execution start (still
   EACCES at plan time). If blocked → path (B): write both reports as blocked
   with the exact fix (docker-group session), append the POC-REPORT attempt
   entry, stop. Unverifiable code written without a host is forbidden by §8's
   spirit — no blind baseline construction.

## Steps

**Phase 0 — Preflight (no docker).** Confirm `stack-poc`, snapshot `git
status`, verify all §1 inputs readable, smoke check (both modules build +
unit green, synth digest still `259564d6…`), record pins/versions. Note the
plan_prompt4 erratum in the working log.

**Phase 1 — Capability assessment (once, bounded).** `docker ps`, compose up
of pinned postgres (digest match), registry reachability, disk/mem actuals;
any failure → path (B) reports + stop.

**Phase 2 — Shared workload (capable host).** Build the DB-backed fixture
(employer, 25 employees, terms, attendance/timesheets, leave, SYNTH overtime,
allowances/deductions, mid-period changes, late approvals/corrections)
reusing the main-app seed patterns where coherent; run the canonical month to
one finalized result + digest; build the 6 outputs (register, payslips,
exceptions, payment preview, snapshot, digest) as read models over
snapshot+calc — no new calculator, no logic duplication.

**Phase 3 — Experiment 1 (PG).** Build the app-only baseline harness
(isolated tables showing races/stale reads/overlaps/gaps/bypass); run the 4
cases against it, then against enforced PG18 (gate, constraints, temporal
coverage, triggers/DEFINER, revokes) as `app`; capture SQL + outputs +
before/after states; conclude the custom infrastructure avoided.

**Phase 4 — Experiment 2 (DBOS).** Build the conventional-job baseline
(crash/double-submit/restart/approval-interrupt pain + recovery-logic
inventory); run the 6 DBOS failure injections on the shared lifecycle; prove
the shared-pool atomic boundary; assert no duplicate/partial financial
artifacts from history; conclude what DBOS eliminates vs what stays in
PG/app idempotency.

**Phase 5 — Experiment 3 (ledger).** PG-only baseline on a disposable clone
(privileged-admin rewrite of snapshot + audit rows, canonical preserved);
anchor canonical evidence via outbox + native SDK (event keys, digest-bound
idempotency); independent auditor from external state + genesis; tamper +
back-link-missing + ambiguous-retry recoveries; no-PII audit; threat-model
limits recorded.

**Phase 6 — Comparison + reports.** Fill the 6-question × 3-technology
assessment and the decision matrix from measured evidence only; write
`plans2/POC-FINAL-REPORT.md` (all 14 sections) + `plans2/POC-RESULTS.md`
(6 sections); update POC-REPORT.md status/tables; per-technology
ADOPT/DEFER/REJECT with cost-vs-value judgment, never auto-adopt.

## Validation Plan

- Preflight: branch, status snapshot, §1 files present, smoke green, digest
  `259564d6…` unchanged (synth-core regression tripwire).
- Each experiment: pre-registered evidence checklist from prompt6
  (SQL/output/state for Exp1; history/checkpoints/counts for Exp2;
  digests/txIDs/auditor checkpoints for Exp3); every output value-inspected
  by independent recomputation; rerun determinism on the finalized digest.
- Final: 7 completion criteria answered one by one; taxonomy complete (every
  experiment PASS/FAIL/BLOCKED/NOT RUN); frozen proofs re-run; synthetic/
  non-statutory labeling present on all outputs; `git status` shows only
  intended files, nothing committed.

## Risks / Open Questions

- Docker access is the dominant risk and still blocked here — path (B) is
  the expected outcome until a docker-group session is provided.
- Fixture + baseline scope creep is the top execution risk: roadmap §4's
  scenario list is the ceiling; anything beyond (banking, filings,
  classifications, rules library) is explicitly out.
- Timed step-window kills stay probabilistic (documented Gate A limitation);
  attempt counts may need host tuning.
- U4 refactor depth is still estimated, not known — Exp2's capture depth
  depends on it; if the estimate breaks, the honest move is narrowing Exp2's
  capture claims, not improvising the refactor mid-experiment.
- Baseline credibility cuts both ways: a weak baseline proves nothing
  (reviewer will probe this), an overbuilt one wastes the POC's remaining
  budget — keep each baseline to the smallest harness that exhibits the
  failure mode.
- Non-goals: exhaustive matrices, Gate A repeats, reopening R1–R7, any
  production decision, merge/commit/00013 without explicit authorization,
  statutory content, new features for demo gloss.
