# Prompt7 Milestone 1 — Implementation Report (`plan_prompt7_imp.md`)

Milestone 1 is complete: the first vertical slice (employee → approved
timesheet → payroll-readiness) works end to end via CLI and REST against
real PostgreSQL. All work is uncommitted over `f435158`; nothing was
committed. Calculations remain gated; OT-001 remains BLOCKED.

Executed per `plans/plan_prompt7.md` as amended by
`plans/plan_prompt7_review.md` (Amendments A–C).

## 1. Features implemented

- Employee administration: create, list, status transitions, effective-dated terms.
- BPO configuration: clients, programmes, teams, team memberships, skill catalogue and grants.
- Shift templates with work/break segments; rotation patterns and pattern assignments.
- Draft → published rosters with team-membership and required-skill validation.
- CSV clock import with explicit break evidence and idempotent replay.
- Attendance pairing, day-record review, and audited superseding corrections.
- Timesheet generate → submit → approve / reject with REG lines over reviewed intervals.
- Payroll readiness: per-period report with explicit blockers; no money computed.
- Single `workforceops` binary; CLI and REST share one service layer (no duplicated logic).

## 2. CLI commands delivered

```text
version, migrate up|status, seed demo [--reset], serve
employer show
employee create|list|status|terms set
client|programme|team create|list, team add-member
skill list|assign
shift-template create|list
rotation pattern create|list, rotation assign|list
roster create|assign|show|publish|list
attendance import|review|correct
timesheet generate|show|submit|approve|reject
payroll readiness
```

## 3. API endpoints delivered

36 spec paths / 46 operations in `openapi.yaml`: 38 marked implemented,
8 frozen draft (KET, leave ×3, payroll-runs, payslips, payroll-register,
rule-packs). New paths added during alignment: `GET payroll-readiness`
(planned, D2), `POST teams/{id}/members` (required so §10 step 5 works via
REST), plus roster detail/assignments, timesheet generate/detail/submit/
reject, and rotation listing. Serves loopback `127.0.0.1:8081` only.

## 4. Database migrations changed

One new migration: `00008_break_events` (Amendment A). Clock-event kinds
now include `break_start`/`break_end`; only break events establish actual
breaks — scheduled shift segments can never pose as observed breaks.
Missing break evidence produces an explicit exception.

## 5. Test results

- `task lint` green (gofmt over module dirs, `go vet`, lib/pq guard).
- `task build` green (single `build/workforceops` binary).
- `go test -count=1 ./...` green twice on scratch PG16: 13/13 packages,
  including DB-gated service suites, integrity G1/G2/G3, and arch boundary.
- New tests: rotation service tests, served⊆spec contract test (both
  directions), CLI↔API parity test (4 cross-adapter legs on one database).
- OpenAPI structural validation: 36 paths, 46 unique operationIds, all
  `$ref`s resolve, every op has responses and a valid maturity — zero errors.

## 6. Demonstrated end-to-end workflow

§10 12-step workflow on two fresh databases (CLI run, then REST run):

1. migrate up; 2. seed MERIDIAN; 3. create E2E-001 + terms; 4. team CARE-A
   membership + EN-SUPPORT grant; 5. roster week 2026-08-10, 5 assignments,
   published; 6. 20-event clock import; 7. review 2026-08-10..14;
8. correction (Fri work interval 18:00 → 17:30, "approved early release");
9. generate PP-2026-08; 10. submit; 11. approve; 12. readiness.

Observed business outcomes (identical both legs):

- Timesheet approved with one REG line of 39.5 hours.
- DEMO-002 correctly skipped at generation (1 unpaired clock event);
  DEMO-001/DEMO-003 generated.
- Readiness: `ready: false`, verdict "not ready: no included employees
  with complete input", all 9 rule packs BLOCKED, per-employee
  missing-info reasons (E2E-001: no SHG instruction).
- Normalized CLI↔REST comparison: 6/6 MATCH (employee, publish, import,
  review, sheet, readiness).

## 7. Known limitations

- List operations return bare JSON arrays (no `{data[], page, …}`
  envelope, no pagination) — documented as Milestone 1 variance in the
  API/CLI contract, rev. 3.
- `Idempotency-Key` is not enforced; clock-import idempotency rests on
  source event IDs.
- Attendance interval IDs are discoverable only via SQL (no
  list-intervals endpoint); corrections need an ID the API cannot list.
- Rotation patterns are CLI-only; skill catalogue listing is CLI-only.
- `X-Operator` / `--operator` is attribution, not authentication
  (acceptable loopback-only, synthetic data).
- Stale `oracle:*` / `scenario:*` / `investigate` / `demo` Taskfile tasks
  untouched (future phases, out of milestone scope).

## 8. Remaining statutory blockers

- OT-001 stays BLOCKED: SSO legislation text uncaptured (403), Fourth
  Schedule text, amendment history, rounding rule/stage, and 72h-breach
  consequences all unresolved.
- No overtime, CPF, SDL, SHG, or payslip calculations implemented — by
  design, not omission. Above-minimum contractual overtime rates are
  unsupported (`indeterminate`) until the calculation slice models them.
- Professional payroll review remains a validation milestone before any
  production/compliance claim.

## 9. Next highest-impact implementation slice

The first independently verified payroll rule package and its calculation
engine — not another architecture redesign. Concretely: capture the SSO
legislation text, pin the OT-001 minimum-rate derivation with golden
tests, and implement the calculation slice behind the readiness gate so
`payroll readiness` can turn ready for verified cases.

## Amendment compliance

- A (break evidence, P0): `break_start`/`break_end` kinds, migration
  `00008`, import contract §7, CLI/API/tests updated. Done.
- B (single-executable consolidation, P0): `cmd/shifthub-api`,
  `cmd/payloop-api`, `cmd/sync-worker` deleted; Taskfile builds only
  `workforceops` and exposes a single `serve` task; full repository
  straggler search clean (remaining mentions are historical plans and
  legacy-schema absence assertions); clean-build developer workflow
  verified. Done.
- C (contractual overtime-rate boundary, P1): exactly-1.5× documented as
  an explicit MERIDIAN employer assumption in `demo-data.md`, not as the
  legal meaning of "at least 1.5 times". Done.

Do not claim WorkforceOps MVP payroll is complete. This milestone
establishes the working HR/WFM/timesheet foundation and an honest
payroll-readiness gate.
