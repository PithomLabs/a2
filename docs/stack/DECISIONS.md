# DECISIONS — Phase 0 decision register (one-month synthetic-payroll POC)

Status: RESEARCH GATE — all entries PROPOSED, owner HUMAN. None approved.
Scope: synthetic data only; no production/statutory/compliance claims
(DOCUMENTED [AGENTS.md](/home/chaschel/Documents/go/a2/AGENTS.md:30-31)).
Source: Phase 0 researcher result (read-only; no commands run).
Companion: [DECISION-REQUESTS.md](DECISION-REQUESTS.md) (human action list),
[RESEARCH.md](RESEARCH.md) (evidence), `adr/` (technical ADRs).

Claim tags: VERIFIED = observed this programme; DOCUMENTED = cited source read;
INFERRED = reasoning only; ASSUMPTION = synthesis beyond researcher evidence.

## 1. Decision register (8 priority decisions)

Vocabulary: "Rung 5" here means DB-enforced protection (the decisions.md rigor
ladder, D4) — NOT reliability-ladder rung 5 (formal verification). The two
ladders are unrelated; conflating them overstates every guarantee below.

| # | Decision | Proposal | Owner | Enforcement (proposed) | Test (proposed) |
|---|---|---|---|---|---|
| D1 | Cutoff definition | Explicit employer-scoped input cutoff timestamp in `Asia/Singapore`, distinct from period-end; eligibility = approved-by-cutoff; late arrivals excluded from run N (DOCUMENTED [decisions.md](/home/chaschel/Documents/go/a2/plans2/checklist/decisions.md:168-172)) | HUMAN | DB rejects ordinary post-cutoff postings into closed population; close txn race-safe (INFERRED) | Approve timesheet T+1 after cutoff → excluded from run N, eligible for adjustment; concurrent write racing close is either committed-before-close-and-included or rejected/redirected, never silently slipped (INFERRED) |
| D2 | Post-cutoff corrections | Lifecycle-staged: pre-cutoff ordinary correction; post-cutoff pre-finalization controlled reopen/replacement run; post-finalization immutable result + linked adjustment in later run preserving effective date + correction date + adjustment period (DOCUMENTED [decisions.md](/home/chaschel/Documents/go/a2/plans2/checklist/decisions.md:215-225)) | HUMAN | Status-gated triggers; no in-place edit of approved artifacts (DOCUMENTED [AGENTS.md](/home/chaschel/Documents/go/a2/AGENTS.md:24-25)) | Post-finalization pay-increasing correction → original run bytes unchanged, adjustment run references it with preserved effective period (INFERRED) |
| D3 | Finality lifecycle | Distinct states draft→captured→validated→approved→finalized→paid; immutable decision artifacts after validation/finalization (DOCUMENTED [decisions.md](/home/chaschel/Documents/go/a2/plans2/checklist/decisions.md:592-598)) | HUMAN | DB state machine + DEFINER-only transitions; closes direct draft→validated bypass, DOCUMENTED P1 ([M2 review](/home/chaschel/Documents/go/a2/plans/milestone2_adversarial_review.md:490-520)) | Direct SQL `UPDATE payroll_run SET status='validated'` as app fails; fabricated/vacuous/placeholder-pin validation rejected (INFERRED) |
| D4 | Adversarial-protection rung | Retain Rung 5 DB-enforced protection for authoritative payroll writes/snapshot/validation; ordinary CRUD stays lower rung (DOCUMENTED [decisions.md](/home/chaschel/Documents/go/a2/plans2/checklist/decisions.md:347-365)) | HUMAN | Revoke app direct INSERT/UPDATE on snapshot*/run pins behind DEFINER functions + triggers (DOCUMENTED [M2 review](/home/chaschel/Documents/go/a2/plans/milestone2_adversarial_review.md:769-777)) | Denial-style battery + fabricated-snapshot/vacuous-validate probes as real app all rejected (INFERRED) |
| D5 | Authoritative input | Only approved business records (approved timesheets, resolved comp basis, covered SHG) feed payroll; raw evidence (clock events, drafts) retained for provenance/correction only (DOCUMENTED [decisions.md](/home/chaschel/Documents/go/a2/plans2/checklist/decisions.md:326-334)) | HUMAN | Capture reads approved set; unapproved/partial coverage → typed blocker, never silent include (INFERRED) | Mid-period comp change and partial-period SHG each produce deterministic blocker; run cannot validate (INFERRED) |
| D6 | Retroactive-change representation | Bitemporal: preserve effective date + recorded date + affected period + adjustment period + successor lineage (DOCUMENTED [decisions.md](/home/chaschel/Documents/go/a2/plans2/checklist/decisions.md:616-623)) | HUMAN | Versioned/successor records; writer and resolver share one effective/coverage definition (DOCUMENTED [prompt_z2](/home/chaschel/Documents/go/a2/plans2/prompt_z2.md:71-77)) | Widening correction + retro comp change resolve deterministically with lineage; no permanently-unresolvable state (INFERRED) |
| D7 | Org boundary | One legal employer for POC; employer = legal boundary, clients/programs = operational; schema keeps employer_id boundary + composite keys (DOCUMENTED [decisions.md](/home/chaschel/Documents/go/a2/plans2/checklist/decisions.md:624-630), [AGENTS.md](/home/chaschel/Documents/go/a2/AGENTS.md:26-29)) | HUMAN | Employer-scoped keys/gates; no cross-employer txn in POC (INFERRED) | Writer cannot move record across employers; client change leaves employment record intact (INFERRED) |
| D8 | Reproducibility evidence | Every validated/finalized run reproducible from immutable inputs + pinned rule versions + approvals; store input set, rule provenance, content hash (DOCUMENTED [decisions.md](/home/chaschel/Documents/go/a2/plans2/checklist/decisions.md:632-638)) | HUMAN | Canon bytes + hash pinned; verification derived from execution/digest-pinned artifacts, never label (DOCUMENTED [prompt_z2](/home/chaschel/Documents/go/a2/plans2/prompt_z2.md:121-125)) | Re-derive run from pinned inputs → byte-identical hash; fabricated digest/failing golden rejected (INFERRED) |

Rationale per decision is recorded in the Phase 0 result; headline: period-end ≠
operational deadline (D1); history stable, recalculation bounded (D2);
preview ≠ authorization ≠ legal consequence (D3); M2 proved app-role
fabrication live (D4); raw ≠ interpreted ≠ approved ≠ frozen (D5); November
correction of October entitlement uses October rules (D6); client ≠ employer
confusion is expensive to unwind (D7); uncaptured evidence cannot be
reconstructed (D8).

## 2. Ten Killer Questions → payroll-month scope

Questions DOCUMENTED [prompt_z2](/home/chaschel/Documents/go/a2/plans2/prompt_z2.md:21-37).
All answers PROPOSED, owner HUMAN.

| # | Question | Proposal | Maps to |
|---|---|---|---|
| 1 | As-of-when / freeze | Cutoff-consistent cut (Design B), not continuous-freshness; close txn race-safe; snapshot from closed population | D1+D3 (INFERRED) |
| 2 | Who may create | Creation authority per artifact (snapshot/validation via DEFINER only; app cannot fabricate) | D3+D4; test E3/F1 battery (INFERRED) |
| 3 | Vacuous pass | Zero rows fails: empty employee tip, empty evidence, missing SHG/comp coverage → typed blocker | D5; trigger + service parity incl. ≥1-employee-row rule (INFERRED) |
| 4 | Status semantics | Each lifecycle word = verified fact with evidence per transition; publish state machine | D3; BLOCKER: B1 active-evidence standard undecided (DOCUMENTED [M2 review](/home/chaschel/Documents/go/a2/plans/milestone2_adversarial_review.md:786-790)) |
| 5 | Whole-period coverage | Union-must-cover-period else uncovered-span blocker; mid-period change blocks until proration specified | D5; needs B2 + X-FUND decisions (DOCUMENTED ibid) |
| 6 | Shared "effective" | Single writer/resolver definition; resolve widening/freetail alignment before POC calc | D6 (INFERRED) |
| 7 | Verification from execution | Derive from digest re-hash + executed goldens; hardcoded GOLDENS_VERIFIED forbidden | D8; blocks first statutory pin (DOCUMENTED [M2 review](/home/chaschel/Documents/go/a2/plans/milestone2_adversarial_review.md:225-250)) |
| 8 | Every write path | Writer matrix artifact required; close P1a–g barrier gaps; new-writer onboarding rule + race test | D4 (DOCUMENTED ibid :379-417) |
| 9 | Missing evidence blocks loudly | Missing ≠ verified-non-applicable; identical typed blockers on readiness/capture/validation | D5 (DOCUMENTED [prompt_z2](/home/chaschel/Documents/go/a2/plans2/prompt_z2.md:109-117)) |
| 10 | Reproducible years later | POC: reproduce from stored snapshot + pins within POC retention only; long-term retention/legal archive DEFERRED, trigger = pre-production decision | D8 (INFERRED) |

## 3. Blockers on Phase 5 schema/topology choices

- D1, D2, D3, D6, D7 block final payroll snapshot schema/migration scope
  (DOCUMENTED [decisions.md](/home/chaschel/Documents/go/a2/plans2/checklist/decisions.md:640)):
  eight decisions before finalizing schema; schema reflects settled semantics,
  migration 00013 only after (DOCUMENTED ibid :690-714).
- D4 sets enforcement architecture cost tier (DOCUMENTED
  [prompt_z2](/home/chaschel/Documents/go/a2/plans2/prompt_z2.md:140-148));
  D5+D8 fix capture population and provenance columns (INFERRED).
- Gate 2.1 hash/snapshot-identity/run-lifecycle/XID/writer-coverage fixes remain
  blockers regardless (DOCUMENTED [decisions.md](/home/chaschel/Documents/go/a2/plans2/checklist/decisions.md:709-711)).
- Production migration NOT authorized: Gate 2 technical direction (RC +
  txn-scoped employer gate, Option A retired) is DOCUMENTED in the
  [gate2 ADR](/home/chaschel/Documents/go/a2/plans/gate2_synchronization_adr.md:5-11)
  but ibid :3 withholds production authorization; M2 adversarial review verdict
  FAIL DOCUMENTED in the [M2 review](/home/chaschel/Documents/go/a2/plans/milestone2_adversarial_review.md:3-10).

## 4. UNRESOLVED (carried — none VERIFIED, Phase 0 ran no commands)

- All 8 defaults await HUMAN sign-off.
- Cutoff sub-questions open: grace period, operator extension authority,
  per-pay-group vs employer-wide, weekend/holiday rule, reopening authority
  (DOCUMENTED [decisions.md](/home/chaschel/Documents/go/a2/plans2/checklist/decisions.md:155-166)).
- B1 active-evidence, B2 coverage rule, X-FUND semantics undecided (DOCUMENTED
  [M2 review](/home/chaschel/Documents/go/a2/plans/milestone2_adversarial_review.md:786-790)).
- Proration semantics unspecified → mid-period must block (DOCUMENTED ibid :107-124).
- Rule-pack effective-coverage semantics undefined; pre-pin prerequisite
  (DOCUMENTED ibid :292-304).
- payroll_line governance (6-item G-PREREQ) required at calculator introduction
  (DOCUMENTED ibid :800-803).
- Cutoff-vs-validation sync scope: does business close narrow Gate 2 Option B,
  or does full boundary remain for POC (DOCUMENTED
  [decisions.md](/home/chaschel/Documents/go/a2/plans2/checklist/decisions.md:396-419)).
