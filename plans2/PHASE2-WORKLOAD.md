# Phase 2 record: shared workload + canonical month (2026-11-10)

## Classification (decided FIRST, per review clarification A)

| Element | Class | Mechanism |
|---|---|---|
| Regular / OT / zero days (SYNTH-OT-001) | CALCULABLE | synth.ComputePay, unchanged |
| Unpaid-leave day, no minutes (SYNTH-LEAVE-001) | CALCULABLE | synth.ComputePay, unchanged |
| Mid-month pack change v1a (OT 3/2) -> v2a (OT 2/1) | CALCULABLE | fully defined synthetic semantics |
| Allowances / deductions | AS-GIVEN DATA | declared in fixture, reproduced verbatim, `source:"fixture-as-given"` — never derived |
| Intra-month base change (SYN-012) | MUST-BLOCK `B-BASE-CHANGE` | proration unresolved; whole employee month excluded |
| Leave day carrying minutes (SYN-006/Oct-10) | MUST-BLOCK `B-LEAVE-WORKED` | day excluded, needs HR disposition |
| Day outside all packs | MUST-BLOCK `B-PACK-GAP` | demo mini-case, synth.ErrNoCoveringPack |
| Two packs covering one day | MUST-BLOCK `B-PACK-OVERLAP` | demo mini-case, synth.ErrAmbiguousPack |
| Negative minutes | MUST-BLOCK `B-BAD-INPUT` | demo mini-case, synth.ErrBadInput |

`synth/` is UNTOUCHED: no new calculator, no new rule. Monthly base is
month-scoped by workload policy (guard in `Canonical`, tested).

## Workload shape (`poc/payroll-month/workload/`)

- 25 employees (SYN-001..SYN-025, no PII) x 31 days, Oct 2026,
  employer POC-MERIDIAN. Bases $2400..$4320 in $80 steps.
- Cohorts: standard (480/540/0 + 1 leave), short-day (240),
  heavy-OT (600), SYN-023 zero-month, SYN-024 four leaves,
  SYN-025 pack-boundary observer (600 min on Oct-15 AND Oct-16).
- 2 events (EV-001 late approval, EV-002 correction) as fixture
  narrative; timestamps are story data, never calc inputs.

## Canonicalization (`poc-canonical-v1`)

Per employee: ComputePay over calculable days (builder order) ->
EmpResult. Month JSON = struct marshal with Digest unset (field-ordered,
deterministic); digest = sha256 hex of those bytes. Covers calc results,
as-given sums, blocks, exceptions, events.

## Outputs (`poc-run canonical`, DB-free)

register, payslips (24), exceptions (5: 1 day + 1 employee + 3 demos),
payment preview, snapshot (= Month), digest.

## Verified this session (no docker)

- `go test ./workload/ ./synth/`: green (8 new workload tests:
  shape, determinism, classification completeness, hand-computed golden
  OT=937/ded=10000, pack boundary 5400/7200, exceptions, totals
  cross-check incl. day-rows-sum, digest stability tripwire).
- `go build ./...`, `go vet`, full `go test ./...`: green; gofmt clean.
- Canonical digest `a35cea69…` since the EV-002 v2 fixture
  correction (was `3cef505e…`; net 8521857¢, was 8519957¢)
  (stability tripwire, NOT independent
  verification); register base total 8072000¢ hand-checked
  (8400000 - 328000 blocked).
- `seedpoc` compiles (services-mapped seeder: hr/wfm/attendance/
  timesheet + documented raw inserts for leave_request — no leave
  service exists — and pay_period/catalogue per internal/seed pattern).
  UNEXECUTED: needs the docker stack.

## DB-backed seeding (EXECUTED 2026-11-10, was: Deferred)

`poc-run seed-poc` exit 0 (attempt 4): 25 employees, 5 roster weeks,
PP-2026-10, 24 approved sheets (SYN-023 zero-month, no sheet), 29 leave
rows, SYN-009 Oct-12 correction pair (2 rows / 1 current), 1060 clocks,
SYN-004 approved. Fixes en route: Sunday-rest validation (workload
zeroes Oct 4/11/18/25 + `TestSundayRest`, digest re-pinned
`3cef505e…`), EV-002 structural match (BusinessDate+work+IsCurrent,
kind `work`), `poc_require_gate` search_path hardening. Exp1–3 ran on
this identical workload — see `POC-FINAL-REPORT.md` §§3–5.
