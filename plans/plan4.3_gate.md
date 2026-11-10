# plan4.3 design-gate report

Executed plan4.3 as amended by plan4.3_review (Amendments A–D). All proof ran
against real PostgreSQL 16.15 (scratch cluster, TCP 127.0.0.1:5433) using the
actual baseline migration — no fixture DDL — with the integrity suite
connected as the unprivileged `app` role in its own database
(`a1_test_integrity`; identity asserted at suite setup).

## Verdict: PASS

Per review §4's three outcomes:

- **Passed now:** PostgreSQL schema integrity, employer isolation,
  draft/published workflow integrity, payroll approval-blocking logic,
  lineage constraints.
- **Specified but deferred:** end-to-end payroll reproducibility
  (`TestReproducibilityDeferred` skips with the contract pointer;
  contract in database-design §8).
- **Still blocked:** every statutory calculation whose pack has not reached
  `IMPLEMENTABLE` (no calculation code exists yet; predicate tested pure).

## Amendment verdicts

### A — `payroll_run_exclusion` in the multi-employer model: PASS

Migration `00002` defines all six columns (`employer_id`,
`payroll_run_id`, `employee_id`, `reason`, `actor`, `excluded_at`) with
composite FKs pinning run and employee to the same employer
(`UNIQUE(employer_id, payroll_run_id, employee_id)` additionally blocks
double exclusion). Draft-only rule enforced by
`trg_payroll_exclusion_locked` (migration `00004`, body fixed in `00006`).
`TestRunFreezeMatrix` proves insert/update/delete succeed in draft and all
three fail from `validated` on.

### B — unambiguous exclusion semantics: PASS

singapore-rule-scope §2 now states all six clauses verbatim:
all-or-nothing per regular run, explicit pre-validation exclusion with
recorded reason, excluded employee not considered paid, exclusion visible
via its persistent row (deletion trigger-blocked once validated), separate
handling required, approved runs disclose exclusions and are never
unqualified complete payrolls. `internal/payroll/approval.go` enforces the
decidable subset (reason+actor mandatory; every included non-excluded
employee's packs `IMPLEMENTABLE`); 7/7 unit cases pass and
`TestActivationFeedsApprovalPredicate` wires a real activation row through it.

### C — reproducibility contract specified, end-to-end deferred: PASS

database-design §8 specifies all five preserved elements and the exact
equation (`same immutable inputs + same rule versions + same calculation
version = same payroll output`). The suite proves the foundation (frozen
runs, `inputs_hash`/`rule_versions`/`calc_version`/`rounding_conventions`
all `NOT NULL`, mandatory lineage keys) and skips the end-to-end test with
an explicit pointer. No passing claim is made for the deferred test.

### D — exhaustive employer scoping + same-employer attendance rule: PASS

database-design §1 coverage list enumerates the full employer-owned
inventory including `payroll_run_exclusion` (verified against the
migrations: all 50 tables accounted for). The interval/record triple-FK
rule is enforced in DDL, and `TestAttendanceAssignmentEmployeeMatch`
proves two employees of the SAME employer cannot link to each other's
assignments (interval and record paths), while the matching link succeeds.

## §3 approval-table verdicts

All twelve approved as specified: real baseline migration (7 versions,
complete design coverage, no services/calcs/seeds); two roles
(`migration_admin` owns schema/DDL, `app` runtime; both proven
non-superuser); trigger parent-state immutability; employer-scoped
clock-event dedup (`UNIQUE(employer_id, source, source_event_id)`,
per-employer namespaces proven); pure approval predicate; evidenced
`verified_none` (CHECK-requires basis+timestamp); read-only activation
endpoint (contract §3 + openapi path, no write surface); structured roster
legal-validation evidence; relational lineage + defined `source_refs`
schemas; operator attribution explicitly non-authenticated on CLI flag and
`X-Operator` header; full reproducibility deferred per Amendment C.

## §4 acceptance condition: PASS

Under the `app` role, draft edits succeed and published/approved edits
fail, covering UPDATE and DELETE wherever immutability is promised
(timesheet chain, roster assignments, runs+lines, exclusions, leave, all 7
insert-only tables). SQL→422 translation is conditional on service
endpoints that do not exist yet; trigger messages use distinct `frozen:` /
`draft` prefixes ready for that mapping — deferred, not missing.

## Consistency audit

- Trigger names diverged from the spec (`timesheet_freeze` vs
  `trg_timesheet_frozen`): fixed forward in `00007`, pinned by
  `TestSpecTriggerNamesExist`. All five §3 names now exist on all eight
  bindings.
- All `§N` cross-references resolve (checked mechanically).
- openapi.yaml holds 28 paths including the new activation endpoint; still
  parses (validated by existing contract tooling).
- Parked-term sweep over live docs + openapi + README/AGENTS: one hit,
  mvp-scope's deliberate out-of-scope contrast ("separate
  ShiftHub/PayLoop systems") — legitimate, no action.

## Bugs the proof caught (all fixed, all re-proven)

1. Trigger functions returned `OLD` on UPDATE, silently discarding allowed
   updates (fixed in `00006`; tests now assert post-update state via
   `requireStatus`, so this class cannot hide again).
2. Suite DSN derivation used `pgx.ConnConfig.ConnString()`, which replays
   the original string and ignored the database/user override — the suite
   silently ran as superuser in the wrong database (fixed with string-space
   derivation + a `current_user`/`current_database` setup assertion).
3. Goose splits pl/pgsql on semicolons without `StatementBegin/End`
   annotations (fixed in `00004`).
4. Foundation tables lacked `id DEFAULT`s (fixed forward in `00005`).
5. Parallel `CREATE DATABASE` can hit template1 contention (retry added to
   the shared helper).

## Evidence

- `task lint`, `task build`: green.
- `go test -count=1 ./...`: 7/7 packages green, repeated (3× consecutive
  clean runs plus a from-dropped-databases fresh-install run).
- Integrity suite: 20 PASS + 1 documented SKIP (`TestReproducibilityDeferred`).
- Predicate units: 7/7. Migration idempotence + status: green.
- Scratch PostgreSQL 16.15 (Ubuntu 16.15-0ubuntu0.24.04.1), user-space
  install; no Docker, no cloud, no credentials involved.

## Recommendation

Close the design-correction pass. The schema, isolation, freezing,
approval-blocking and lineage contracts are proven and may be built upon;
no further redesign before the first implementation slice.
