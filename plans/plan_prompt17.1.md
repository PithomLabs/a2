## Goal

Amend `plans/plan_prompt17.md` with consolidated review6 amendments A–K and review6.1 final amendments A–G, and execute Milestone 2.1 only through the resolved design gates: (1) a database-enforceable D1 pre-BEGIN admission boundary (or an approved minimal privilege-boundary change), and (2) an architecture decision record selecting one authoritative snapshot-data implementation (canonical PostgreSQL resolver vs canonical Go resolver + restricted execution boundary). All 20 acceptance conditions, single `00013` migration, `00001`–`00012` byte-identical, OT-001 BLOCKED, no statutory calculations, work uncommitted. This execution unit is authorized through the Gate 1 package only (steps 1–6); steps 7–9 proceed solely after Gate 1 approval.

## Success Criteria

- Plan file amended with A–K; the two architectural problems resolved with live proof before the capture path is implemented: the exact stale-RR direct-SQL attack fails, and exactly one compensation resolver + one canonical hash implementation exists (no Go/SQL duplication).
- All prompt17 §15 acceptance conditions hold with adversarial evidence; `plans/milestone2_1_remediation_report.md` (20 sections) written; full battery green; execution stops at the gate, stopping early for approval if D1 needs a privilege change.

## Approach

Retain plan_prompt17's scope, order backbone, and invariant-first principle; apply review6's resequencing (P1a semantics + authority ADR + admission proof all decided before capture/validation implementation) and its restored enumerations. Two approval gates punctuate execution: **Gate 1** (after the ADR + admission-proof steps) presents the D1 mechanism decision and, if needed, the privilege-boundary change request — implementation of the capture path does not proceed past an unproven mechanism; **Gate 2** is the standard acceptance gate.

Key decisions, grounded in research done for this plan:

- **Amendment A (D1): settled empirically, not assumed.** Live probe on PG 16.15 (scratch cluster): granted session-level and xact-level advisory locks produce byte-identical `pg_locks` rows (same locktype/mode/granted, `transactionid` NULL, `waitstart` NULL — `waitstart` populates only for waiting locks). PostgreSQL 16 therefore exposes neither session-vs-xact advisory ownership nor acquisition order relative to the MVCC snapshot. A `pg_locks` ownership check **cannot** prove pre-BEGIN admission — plan_prompt17's proposed check is abandoned as a proof mechanism (kept at most as a held-lock sanity assertion). Execution must reproduce the §A stale-RR attack red, then prove one of: (i) an XID admission ticket verified by visibility, not `xmin` comparison (review6.1 §A: `pg_visible_in_snapshot(ticket_xid, pg_current_snapshot())` plus direct visibility of the committed ticket record inside the caller's RR snapshot; top-level PostgreSQL-generated XID; ticket created by a separate READ COMMITTED operation committed before the RR capture transaction begins; the acquisition helper rejects invocation from an existing RR transaction; all 7 ticket cases tested — before/after/nested-issuance/foreign-backend/expired-or-released/old-xmin-pressure/aborted-txn; `xmin > ticket_xid` never the sole acceptance condition), or (ii) a restricted snapshot-execution boundary ordinary `app` cannot invoke (expected to require the privilege-change approval stop). Timestamps, session variables, and re-acquisition tricks are excluded by prompt17 §2.1 and advisory-lock reentrancy respectively. Planning-time live probe already confirmed the visibility predicate semantics (pre-snapshot ticket visible, post-snapshot ticket invisible, both under real global-`xmin` pressure); execution re-proves inside the ticket lifecycle, never trusts the probe.
- **Lock substitution (review6.1 §B).** Because session/xact locks are indistinguishable in `pg_locks`, a held-lock check alone cannot prove the admission lock was never released and re-acquired as a transaction-level lock. The 6-step lock-swap attack (valid ticket → RR → snapshot → release session lock → acquire xact lock on the same key → invoke capture) is a committed failing test and must fail post-fix. Investigate revoking direct `EXECUTE` on all 21 advisory-lock/unlock overloads from `PUBLIC`/`app` (lock/unlock/try × session/xact × shared/exclusive × bigint/(int,int), plus `unlock_all`; revoke is per-database and belongs in provisioning plus a migration-time guard), with narrowly scoped SECURITY DEFINER wrappers as the only path; every legitimate writer/snapshot lock operation uses the same canonical wrapper and key derivation. `app` must be unable to release an admission lock, substitute a transaction-level lock, acquire an unapproved lock identity, or bypass the ticket lifecycle. If unenforceable under two roles: stop at Gate 1, no silent third role.
- **Amendment B (authority): default Option 2** (Go canonical resolver + genuinely restricted snapshot-writing boundary) per the reviewer's stated preference, but the ADR compares both on all eight review6.1 §E criteria (one authoritative compensation definition, snapshot content completeness, direct-`app` SQL threat, capture-function privileges, canonical hash authority, transaction consistency, operational complexity, frozen-two-role compatibility) with live evidence before any capture code is written, and documents the precise trust boundary. Under Option 2 the ADR must state exactly how the database prevents fabricated snapshot content from ordinary `app` (referential validation alone is not completeness proof against omitted inputs); the database verifies the digest of the exact canonical bytes from the trusted writer and keeps no second serializer. Conditional resequence encoded: if the ADR selects Option 1, compensation work moves ahead of capture-function implementation (review §C note); under Option 2 the D1 privilege question and the snapshot-execution boundary converge on the same boundary and are decided together. Canonical hashing follows the same single-implementation rule. `migration_admin` is never used as a broad runtime writer without an explicit excessive-privilege analysis; any dedicated snapshot-execution identity is a privilege change requiring approval.
- **Ordering and lifecycle (review6.1 §C–D).** Admission follows the 8-step order verbatim: employer population session barrier → fresh READ COMMITTED population enumeration → employee session barriers in UUID order → pack session barriers in identity order → create and commit the admission ticket → BEGIN RR on the same physical backend → revalidate population and lock set → capture. Population is never enumerated from a pre-lock statement snapshot; any mismatch aborts safely under bounded retry. One physical connection carries admission through commit/rollback/release; cleanup is tested after success, rollback, SQL error, cancellation, deadline expiry, panic, unlock failure, and pool return; uncertain session-lock state discards the connection, never returns it healthy.
- **Amendment C (sequence)**: adopt the 12-step order verbatim (Phase 0 → D1 reproduction → P1a semantics decision → authority ADR → admission proof → barrier matrix → capture/validate once → comp → SHG → packs → battery → report). No temporary Go capture path.
- **Amendments D–K**: incorporated as checklists — full DEFINER hardening + per-revocation positive controls (P1d prefers the status-transition function); complete C1/C2 field inventories; explicit `draft → validated`-only with direct-SQL tests for all three unreachable transitions; P1g 7-field drift enumeration with one mutation test each; the 12-category snapshot-value inventory replacing the "13 value classes" claim; the 14-case direct-SQL attack matrix with persisted-state + no-partial-write assertions after every rejection; project boundaries restated (incl. payroll lines as explicit non-goal); `00012` hashed at Phase 0 (no recorded checksum claimed); drop/recreate guarded by verified disposable-identity allowlist, never bare naming convention; three TZ runs; no extra doc files.
- **Documentation home (assumption, reversible)**: no SHG domain-spec file exists in the repo and §K forbids new docs without approval, so the ADR, SHG semantics, and fixture-schema contract live in the 20-section report plus concise enforcement-point code comments — no standalone files.

## Steps

1. **Amend the plan.** Apply review6 A–K and review6.1 A–G to `plans/plan_prompt17.md` (sequence table, D1 pg_locks verdict with the empirical evidence, XID-visibility protocol + 7 ticket cases, lock-substitution attack + 21-overload revoke inventory, 8-step ordering, connection lifecycle, ADR step with conditional resequence, DEFINER/pin/drift/value/attack-matrix checklists, boundary + safeguard restatements, Gate-1 pending decisions). Verify the amended file reads coherently end to end.
2. **Phase 0.** Baseline/inventory per prompt17 §3 as already planned; additionally hash `00012` and record tool versions.
3. **D1 + writer-race reproduction (red).** Writer-first races with `pg_locks` NOT-granted proof; the §A stale-RR attack and the §B 6-step lock-substitution attack as committed failing tests; validation-level T2; the 4 missing writer classes.
4. **P1a semantics decision.** Inspect population derivation; decide effective-dated rules (preserving documented legacy behavior explicitly); record before enumeration work.
5. **Authority ADR.** Live-backed comparison of Option 1 vs 2 (+ hash authority) on all eight §E criteria; recommendation with precise trust boundary and fabrication-prevention mechanism; conditional resequence if Option 1 wins.
6. **Admission proof + Gate 1.** Implement candidate D1 mechanism(s): XID-ticket lifecycle (all 7 §A cases), 21-overload revoke + DEFINER wrappers with dependency audit, 8-step ordering, connection lifecycle with all 8 cleanup cases. Assemble the 7-item §F live-evidence package (stale-RR, ticket visibility, old-xmin false-reject, lock-substitution, legitimate concurrent capture, population/lock-set verification, authority fabrication). Present the mechanism decision; **if a privilege-boundary change is required, stop and request approval** — no weaker guarantee to keep moving. This unit ends here with the Gate 1 package.
7. **Barrier matrix + capture/validate once (post-Gate-1, not authorized in this unit).** P1a–g fixes, both orderings, direct-`app` SQL; single capture/validate implementation per the approved design; DEFINER hardening checklist; attack matrix §I green with positive controls.
8. **Compensation, SHG, packs (post-Gate-1).** Per amended plan (resolver split, liveness alignment, shared mid-period blocker, F-A2 IDs; B1/B2/X-FUND + negatives; source byte verification, real golden execution, pin ceremony, labels, coverage). Fixture packs only; OT-001 BLOCKED at all three levels.
9. **Battery + report (post-Gate-1).** §12 A–G incl. CLI/REST differential, 19-package list + skip table, clean-server/fresh-name acceptance; guarded `00013` Down round trip; 20-section report with per-finding evidence, the four report-claim corrections, and the six §G record items (pg_locks insufficiency, xmin false-reject, ticket mechanism, advisory privilege boundary, canonical authorities, privilege request). Uncommitted; stop at the gate.

## Validation Plan

- Amendment audit: re-read the amended `plan_prompt17.md` fully; every A–K item traceable to a step, checklist, or gate; no contradiction with prompt17 §15/§16 or `AGENTS.md`.
- Red-first proof: stale-RR attack and writer-first races fail before their fixes (observed, not assumed); each fix re-verified by its race plus both orderings, persisted-state inspection, and atomicity/no-partial-write assertions.
- Highest-risk validation: Gate 1 live evidence — the stale-RR attack failing against the chosen mechanism while legitimate concurrent captures still succeed (no false-reject livelock). If neither candidate survives, the correct outcome is the privilege-request stop, not a workaround.
- Final acceptance exactly per prompt17 §12G; exact commands, DB identities, and outcomes recorded.

## Risks / Gate-1 Pending Decisions

- Risks: the XID-ticket design must survive all 7 §A cases including foreign-backend and aborted-txn tickets (execution proves each; the restricted-boundary alternative avoids ticket complexity at the cost of a privilege change); the 21-overload revoke is per-database behavioral change (mitigated by overload inventory + dependency audit + migration guard); Option 1 would move substantial resolver logic into SQL (ADR must weigh duplication risk vs trust-root strength honestly); the two gates can stall progress by design — that is the review's intent, not a defect.
- **Unresolved until Gate 1 passes live verification:** (1) the D1 admission mechanism (XID-ticket vs restricted boundary, and whether a privilege change is required); (2) the snapshot-authority selection (Option 1 vs 2, hash authority, exact trust boundary, migration_admin analysis). Full capture/validation implementation is explicitly not authorized until these are proven. All other choices are pinned by review6/review6.1, prompt17, or workspace precedent.

## Gate 1 Decision Records

### P1a — Expected-population semantics (decided, pre-enumeration)

Inspected `expectedPopulation` (`internal/payroll/snapshot.go:693`), its
five call sites (capture enumerate + revalidate, drift detection,
validate × 2), and the readiness copy (`internal/payroll/readiness.go:145`).

- **Semantics (pinned, no change in M2.1):** employer-scoped members
  are employees with ≥1 `employment_terms` row overlapping
  `[period.start, period.end]` (NULL `effective_to` = open) **OR** ≥1
  `status = 'approved'` timesheet for the period. Status-independent:
  termination surfaces via pinned `StatusIDs`/conflicts, never via
  membership. Deterministic order (UUID for lock acquisition;
  `employee_no` acceptable for display-only readiness).
- **Approved-sheet disjunct preserved explicitly** as documented
  legacy behavior (relied upon by capture/readiness agreement and
  pinned in test comments); removing it would change validation
  outcomes and is out of scope.
- **Divergence risk recorded:** readiness carries its own textual
  copy of the predicate (same shape today, verified by inspection).
  Post-Gate-1 hardening item: route readiness through the shared
  definition with identical behavior, covered by the barrier-matrix
  tests. Not touched in this unit.

### ADR-1 — Snapshot authority: Go-canonical resolver + verifying writer (decided)

Question (review6.1 §2): how does the database know a Go-resolved
snapshot is complete and authentic without a second
compensation-selection algorithm? Compared live-backed on all eight
§E criteria. Evidence: `internal/compbasis/resolve.go` (366 lines;
`carveSegments` alone carries chain-integrity, date-carving, and
conflict-taxonomy logic), canonical hash = Go `json.Marshal` struct
order + sorted map keys + SHA-256, live privilege probe (app holds
INSERT on all snapshot tables and UPDATE on `payroll_run`),
`TestSnapshotFabricationByApp` (red: invented `valid` snapshot
readable), and the existing revoke-plus-DEFINER precedents
(`activate_rule_pack` not app-executable;
`add_shg_instruction`/`transition_employee_status` app-executable
narrow writers, all `migration_admin`-owned).

1. One authoritative compensation definition: **O2.** Go's tested
   resolver (16 unit tests + M2 matrices) stays the single
   definition. O1 would port date arithmetic and the conflict
   taxonomy to plpgsql with unquantifiable port-fidelity risk.
2. Snapshot content completeness: **O2, via proof-checking, not
   re-solving.** The writer verifies claimed segments against the
   tables: set-equality of claimed line IDs with candidate lines
   (no omission, no invention), per-row ownership/date/amount
   equality, lineage-link integrity, liveness as local
   successor-existence checks, conflict claims as COUNT/overlap
   checks over the claimed set. Verification is O(rows) local
   checks over caller-supplied claims; selection stays in Go.
   Decisions are derived by the writer from verified pins (single
   decision authority); Go's assessment moves out, not duplicated.
3. Direct-`app` SQL threat: **O2 closes it.** `app` keeps EXECUTE
   on the writer (same credentials as the trusted path) but every
   persisted byte is derived or verified; invocation without
   content control cannot fabricate. Residual: invocation itself
   (spurious captures on arbitrary drafts) is governed by run
   state + ticket + versioning, never an integrity violation.
4. Capture-function privileges: **O2, no third role required.**
   `migration_admin`-owned DEFINER writer + ticket + wrappers,
   each writing only its narrow table set — consistent with
   established precedent, not broad runtime writing.
   `migration_admin` remains human/migration-held; the functions
   are the capability boundary. (Contingent on post-Gate-1 live
   proof of every verification in the 14-case matrix.)
5. Canonical hash authority: **O2, single serializer.** Go keeps
   the one canonical form (unchanged bytes ⇒ M2 history stays
   comparable); SQL asserts parsed-content field coverage over
   verified pins and recomputes digests over the supplied bytes.
   SQL never generates canonical bytes. O1 would fork the byte
   form across the M2/M2.1 boundary.
6. Transaction consistency: tie — both run inside the caller's
   transaction; both need the ticket (O2) or equivalent freshness
   binding.
7. Operational complexity: **O2.** Go service shape preserved;
   new SQL is verification procedures, not business logic. O1
   moves future canonical-format changes into plpgsql.
8. Frozen-two-role compatibility: **O2 without modification.**
   No new role, no new credential. O1 likewise compatible but
   loses on 1/5/7.

**Decision: Option 2.** Conditional resequence not triggered
(compensation work stays after capture/validation implementation).
The D1 question and the snapshot-execution boundary converge:
ticket + revoke + verifying writer, all provable under two roles
or escalated at Gate 1 with evidence.

### D1 — Admission boundary: XID-ticket + session barriers + full revoke (decided, privilege approval pending)

Evidence: `core/database/admission.go` (experimental, unwired),
`core/database/admission_test.go` (31 green), 12 red attack tests
still failing, full package in `plans/gate1_admission_package.md`
(§F × 7, §G × 6). All §A 7 ticket cases, 21-overload revoke with
dependency audit, 8-step ordering, all 8 cleanup cases green.

- **Selected: boundary A (full 21-overload revoke).** Session
  barriers + XID ticket + DEFINER wrappers; readers and writers
  share one namespace and one key derivation. Self-enforcing and
  deadlock-free by construction. Boundary B (session-only revoke)
  rejected: integrity-sound but leaves bounded xact-DoS and
  unenforced discipline.
- **Ordering proof is XID visibility** (row visibility +
  `pg_visible_in_snapshot`), never xmin comparison, never `pg_locks`
  scope inference (session/xact rows byte-identical on PG16.15).
- **Privilege request (Gate 1 approval required):** revoke 21 from
  PUBLIC (migration_admin only); redefine the 3 INVOKER helpers as
  DEFINER with byte-identical bodies (proven transparent to
  direct/writer/trigger callers); bootstrap wrapper; flip 3
  `definer` pins in `privilege_test.go`. No third role.
- **Stated limits:** subtransactions, prepared xacts, wraparound
  out of scope for short-lived tickets; canonical key function and
  timeout-bounded waits are 00013/post-Gate-1 items.
- Capture/validation implementation stays unauthorized until the
  privilege request is approved.
