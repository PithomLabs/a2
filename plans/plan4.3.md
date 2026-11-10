## Goal

Implement `review2.md`: correct the five P0 issues and six smaller improvements (5 P1 + 1 P2) in the five design docs, then prove the three integrity test groups against actual PostgreSQL as the reviewer demands before the design gate.

## Success Criteria

- `employer_id` is explicit on every employer-owned table including all six named children, with composite FKs throughout; an attendance interval cannot reference another employee's roster assignment even within one employer (composite-FK enforcement).
- Draft rosters, timesheets, and interval allocations are mutable; published/approved records are immutable via parent-state-aware triggers; insert-only tables stay insert-only.
- Clock-event dedup is `UNIQUE(employer_id, source, source_event_id)` with the per-employer namespace principle recorded for future integration IDs.
- Payroll approval is all-or-nothing per run: any blocked employee halts approval unless explicitly excluded pre-validation with an audit reason; exclusions are represented, never silent.
- `employer` gains `code` with `UNIQUE(code)` and `UNIQUE(jurisdiction, uen)`; CLI accepts code or UUID.
- The six improvements land: read-only activation endpoint, structured `legal_validation`, specified `source_refs` schemas with relational links, operator attribution for CLI and API writes, and an explicit evidenced SHG no-deduction outcome.
- All three test groups (two-employer, workflow, payroll integrity) pass against real PostgreSQL; frozen §7 decisions untouched.

## Approach

**Authority.** review2 governs; its frozen table (§7) is not reopened. This pass supersedes my earlier "no migrations writes" fence — deliberately and for one stated reason below.

**Key decisions:**

- D1 — The reviewer requires tests against actual PostgreSQL *before* gate approval. Fixture-copy DDL would prove a copy, not the design, and rot on contact. So this pass writes the real baseline migration (DDL only: tables, keys, triggers, no seed, no services) and runs all three groups against it. Pre-release migrations may be rewritten; immutability starts at the first tagged release.
- D2 — One pre-gate production file: the approval-blocking predicate (`unverified pack → run cannot approve; explicit exclusion → proceeds`) with table-driven unit tests. It encodes singapore-rule-scope §2–§3, is reused verbatim by the payroll slice, and contains no calculations.
- D3 — Partial pre-gate honesty split: FK/trigger/privilege/lineage properties execute now; true end-to-end reproducibility (same inputs → same outputs) needs calculation code, so its exact contract is specified now and executes at the payroll slice — never claimed early.
- D4 — Exclusions become a first-class entity (`payroll_run_exclusion` with reason, actor, timestamp, pre-validation-only rule); SHG gains a `verified_none` status requiring basis plus verification source, while a missing instruction row still blocks net pay.

## Steps

1. **P0 doc corrections.** `employer_id` plus composite FKs on all six children and the employee-assignment consistency key; the draft-mutable/published-frozen matrix with trigger specs; employer-scoped dedup key; all-or-nothing approval with the exclusion entity and updated blocking matrix; CLI code-or-UUID acceptance.
2. **P1/P2 corrections.** Read-only activation endpoint (plus OpenAPI path); structured `legal_validation` columns; `source_refs` JSON schemas with mandatory references and the relational-link rule; `--operator` on CLI mutations and `X-Operator` on API writes as documented non-authenticated attribution; evidenced SHG `verified_none`.
3. **Baseline migration.** Real initial migration per revised `database-design.md` (replacing the pre-pivot 00001); `init-roles.sh` reduced to the two-role model; seed untouched (belongs to the seed slice).
4. **Integrity tests.** G1: the eight two-employer contracts; G2: draft-mutable/published-frozen plus insert-only proofs, privilege assertions via `SET ROLE app`; G3: approval-predicate units, lineage-constraint proofs, version-pinning checks; reproducibility contract specified for the payroll slice.
5. **Audit and gate report.** Cross-document consistency re-verified, OpenAPI parsed, full suite green, parked-term sweep; report with per-finding verdicts and the gate recommendation.

## Validation Plan

- Each review2 item traced to its revised section; each test-group item traced to a named passing test or an explicitly deferred contract with its owning slice.
- `task lint`, `task build`, `go test ./...` with DB-gated suites against scratch PostgreSQL 16 (Compose-compatible); `openapi.yaml` parses; all 27+ paths carry correct maturity flags.
- Highest-risk check: G2 trigger behavior under the `app` role — the reviewer's most concrete defect (revoked-draft-editing) must invert exactly: draft edits succeed, post-approval/publish edits fail, and the failure surfaces as the documented machine-readable code.

## Risks / Open Questions

- **Q1 (decision needed before step 3):** Confirm the real-baseline-migration vehicle for pre-gate proof. If you prefer fixture-only DDL instead, step 3 swaps to fixture files plus a drift manifest — say so and I'll re-plan that step.
- Risks: migration churn if the gate requests changes (fence: pre-release rewrite rule, resets documented); trigger-logic creep (fence: triggers stay short status predicates, business rules stay out); scope creep into services or calculations (fence: D2 names the single exception; everything else waits for its slice).
