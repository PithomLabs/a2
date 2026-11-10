## Goal

Implement prompt7's Milestone 1: a working employee→approved-timesheet→payroll-readiness backend slice (CLI + REST over shared `internal/` services, real Postgres), with statutory calculations gated, OT-001 staying BLOCKED, and all 60 uncommitted entries preserved.

## Success Criteria

- The 12-step §10 workflow runs end to end on a fresh local Postgres via CLI and via REST, ending in an explicit blocked-readiness result.
- `task lint`, `task build`, `go test -count=1 ./...` (with DB suites) green; OpenAPI parses and every served route is a specified path.
- No Go payroll math, no golden fixtures from unresolved rules, no `min(H,72)` payable-hours logic anywhere, no status promotion of OT-001.
- Final 9-item §11 report delivered.

## Approach

Build services first, thin adapters second, in dependency order (HR → BPO → rosters → attendance → timesheets → readiness). All business logic lives once in new `internal/` packages using hand-written pgx against the existing 7-migration schema; Cobra commands and Echo handlers only adapt. Reuse `core/database` (pools/migrate), `core/api` (server), `core/config`, `internal/payroll.AuthorizeApproval`, and `tests/dbutil` for DB-gated tests. No redesign of frozen pieces.

**Key decisions**

- **Serve consolidation:** add `workforceops serve` (loopback-only API) as the single executable and delete the three stale pre-pivot mains (`cmd/shifthub-api`, `cmd/payloop-api`, `cmd/sync-worker`). Rationale: prompt7 §2 and the pivot docs mandate one executable; parallel servers would violate the frozen architecture. Stale Taskfile tasks (`run:shifthub`, `oracle:*`, `scenario:*`, etc.) left untouched — out of scope, harmless.
- **Readiness endpoint:** add `GET /api/v1/employers/{employer_id}/payroll-readiness` to `openapi.yaml` (no such path exists among the 28) and ship it `implemented`; reuse of `payroll-runs` GET as semantically wrong (no run is created) is rejected.
- **Maturity flips:** only slice-covered paths flip `draft`→`implemented` (employers, employees, terms, clients, programmes, teams, skills, shift-templates, rosters+publish, clock-events, attendance+corrections, timesheets+approve, pay-periods read, rule-packs/activations read, readiness, health). Leave/payroll-run/payslip/register paths stay `draft`.
- **1.5× minimum (imp-review P0):** take the narrow option — MVP supports only arrangements paid at exactly the statutory minimum 1.5×; anything else is unsupported/`indeterminate`, documented in the package. Rationale: no pay computation exists in this slice, so contractual-rate modeling has no consumer yet; option 2 returns with the calculation engine.
- **Spec-semantics cleanup (§8) is in scope** as bounded doc edits per prompt5_imp_review (5 P0 + 4 P1 meaning fixes); the 4 cross-document one-liners (clock uniqueness, employer_id wording, AGENTS multi-tenancy, rule-scope §2/§3) ride along. Assumption: both stay paragraphs/sentences, no research campaign — say so if you want them split out.

## Steps

1. **Baseline:** `git status`, confirm test baseline green with scratch Postgres, confirm no schema change needed. No edits.
2. **Doc cleanup first:** OT-001 semantic corrections (compliance-vs-remuneration split with `min(H,72)` deleted; cap branches → single R₁/R₂ gap with rounding `?`; 1.5× minimum wording; eligibility chain givens; C22 split into SPEC givens + `?`s; employer-request demotion; basic-salary exclusions; SHIFT dependency link) + 4 cross-doc one-liners. Re-sweep SPEC for status words.
3. **Service skeleton:** `internal/hr`, `internal/wfm`, `internal/attendance`, `internal/timesheet`; shared employer-context resolution (code-or-UUID), operator attribution, JSON error envelope; readiness in `internal/payroll` reusing the predicate.
4. **Slices A–B (HR/BPO):** employer lookup, employees, status, effective-dated terms + overlap rejection, compensation lines, clients, programmes, teams/memberships, skills, required-skill gating for rostering.
5. **Slice C (rosters):** templates + segments, draft rosters, assignments (employer/skill/team/duplicate/rest-day validations), publish, revision immutability; legal validation always `unavailable`.
6. **Slice D (attendance):** documented CSV import → immutable clock events, `(employer_id, source, source_event_id)` dedup, idempotent replay, deterministic in/out pairing with UTC+local preservation, cross-midnight + shift identity, work/paid/unpaid split, scheduled/recorded/actual/paid kept distinct, exceptions for missing/unmatched/out-of-order/overlapping, auditable superseding corrections. 72h-breach is an exception, never a truncation.
7. **Slice E (timesheets):** generation from reviewed intervals with `timesheet_line_interval` lineage, draft/submit/review/reject/approve, idempotent approval, trigger-backed immutability.
8. **Slice F (readiness):** read-only assessment (employer, period, population, required packs, activation status/version, blocked components, missing info, explicit not-ready verdict). Zero financial output.
9. **CLI:** replace stubs with the §4 command set (file inputs, `--format json`, `--operator`, exit codes) + `seed demo` (deterministic Meridian fixture incl. overnight shifts, duplicate delivery, missing punch, approved/rejected timesheets) + `serve`.
10. **REST:** handlers for the §5 surface, employer scoping, `X-Operator` on writes, OpenAPI request/response schemas + maturity flips.
11. **Tests:** per-§6 suites as DB-gated integration tests (no mocked repos); CLI↔API parity test (same workflow, separate fixtures, normalized compare); served-routes ⊆ openapi-paths contract test.
12. **Gates + report:** full §10 command set, fresh-DB E2E demo run, 9-item §11 report. Stop there.

## Validation Plan

- `task lint`, `task build`, `go test -count=1 ./...` with `WORKFORCEOPS_TEST_DATABASE_URL` on scratch PG16 — expect 7/7 packages incl. new suites, repeated 2× for flake check.
- Fresh-DB acceptance: drop DBs → migrate → `seed demo` → run the 12-step workflow via CLI, then via REST; diff normalized business results.
- OpenAPI: YAML parses; every served route exists in the spec; every flipped path has schema + handler + contract test.
- Doc checks: SPEC status-word grep empty; `min(H,72)` absent from code and payable-hours prose; OT-001 manifest still `BLOCKED`.
- Highest-risk validation: the CLI↔API parity test — it is the only check proving one shared service layer rather than two implementations.

## Risks / Assumptions / Non-goals

- Risk: slice D pairing edge cases (out-of-order, ambiguous attribution) — contained by making every non-clean case an exception, never a guess.
- Assumptions: consolidation deletes 3 stale mains; §8 + cross-doc fixes ride in this milestone; no forward migration needed (schema already covers the slice; 00008 only if genuinely required).
- Non-goals: everything in §9 (UI, optimization, RLS, integrations, any payroll math, fixtures from unresolved rules), plus SSO retrieval and OT-001 promotion.

## Open Questions

None — all material choices are grounded above; the two assumptions are flagged for veto.
