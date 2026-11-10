# Milestone 1.1 — Remediation Report (`milestone1_1_remediation_report.md`)

## 1. Initial Git and working-tree state

- HEAD `f43515883668c917edac31be49a7cf8334eca5c2` ("Phase 1 scaffold
  baseline (pre-pivot)"). No commits were made during M1.1; all work is
  uncommitted, per prompt9 §10.
- Baseline tree: 79 entries (12 modified, 4 deleted, 35 renamed, 28
  untracked) — identical to the tree Prompt 8 reviewed.
- Final tree: 87 entries (16 modified, 4 deleted, 35 renamed, 32
  untracked). The +8 delta is new M1.1 files (services, tests, audit,
  this report).
- Migrations on disk at baseline: `00001` … `00008`. At completion:
  `00001` … `00009` (`00001`–`00008` untouched; see §7).
- Environment: Go 1.25, real PostgreSQL 16 (scratch server, disposable
  databases only), `GOWORK=off` (a parent `go.work` outside the repo
  would otherwise capture the module).

## 2. Structural audit findings

Full text: `plans/milestone1_1_structural_audit.md` (prompt9 §2, read-only).

- 17 structural blockers, all re-verified by reading the cited bodies,
  all fixable inside the frozen architecture. Every one is remediated
  below except where explicitly deferred with justification.
- 17 clean review areas confirmed standing; module registry proven
  dormant (`go list -deps ./cmd/workforceops` references neither
  `modules/sg-wfm-payroll` nor `core/module` — re-verified at close:
  0 references).
- Frozen-decision check: no finding required abandoning any §7 frozen
  decision. Work never stopped for approval.

## 3. Confirmed defects and exact remediation

### P0-1 — Approved timesheets silently diverged from corrected attendance

`Correct` never consulted `timesheet_line_interval`. Fixed per review
amendment A with append-only correction versions, originals immutable:

- `internal/attendance/review.go` `Correct`: submitted/approved-linked
  intervals reject loudly (reissue the sheet instead); linked drafts
  regenerate in the same transaction; rejected links are inert history.
- `internal/timesheet/timesheet.go` `Reissue`: new draft links via
  `supersedes_id`; one open successor per chain enforced in-tx plus
  `timesheet_single_correction_idx`; `EffectiveApproved` resolves the
  approval that counts; `RegenerateDraft` rebuilds draft lines.
- Schema (`migrations/core/00009_m11_integrity.sql`): `supersedes_id`,
  single-successor partial index, `enforce_interval_single_chain`
  trigger (correction versions share evidence; cross-period
  double-count rejected).
- Rejected recovery (§5B): `Generate → Submit → Reject → Correct →
  Resubmit → Approve` proven by `TestRejectedRecoveryFlow`.
- Exposed on CLI (`timesheet reissue`) and REST
  (`POST …/timesheets/{id}/reissue`), audit `timesheet.reissue`.

### P0-2 — Unbounded clock-event spans became clean payable time

`pairSpans` had no span guard. Fixed per amendments B + C:

- Roster-anchored pairing is the primary rule (employee, source stream,
  assignment, expected window, cross-midnight, multi-shift, ordering,
  duplicate/conflict handling); 24h ceiling stays a second defensive
  check (`ExcSpanTooLong`); unrostered/ambiguous events persist
  exceptions, never guesses (`internal/attendance/attendance.go`).
- 72h monthly excess is compliance-only: it flags `NeedsReview`, never
  truncates, caps, or hides verified hours. Proven by
  `TestMonthlyExcessIsComplianceOnly` (75h excess flags AND all 210h
  generate). No 72h payable cap exists anywhere.
- Cross-midnight splits at local midnight per the domain contract
  (`TestCrossMidnightSplit`); session-TZ casts replaced with
  `(occurred_at AT TIME ZONE timezone)::date` in both queries.

### P1-1 — Freeze-trigger re-parenting bypass

`00006` non-INSERT branches resolved the OLD parent only. `00009`
redefines the functions with NEW-parent checks and employer-scoped
lookups. Proven by `TestFreezeReparenting` (re-parenting into an
approved/published parent fails; legitimate moves succeed).

### P1-2 / P1-21 — Duplicate and overlapping intervals

- Partial GiST exclusion on live intervals keyed on the explicit
  `is_current` predicate (maintained by a SECURITY DEFINER trigger
  owned by `migration_admin`; app cannot UPDATE the table).
- Service pre-checks (pairing overlap → `overlapping_span` exception,
  §5 below; `Correct` overlap → loud conflict) plus per-employee
  transaction-scoped advisory locks.
- Concurrent corrections: exactly one successor
  (`TestConcurrentCorrectionsSingleSuccessor`); concurrent imports:
  exactly-once storage (`TestConcurrentImportsExactlyOnce`).

### P1-3 — Conflicting duplicate delivery

Fetch-compare-conflict replaces `ON CONFLICT DO NOTHING`: stored event
never overwritten, `conflicting_redelivery` persisted, audit
`clock_event.conflict` written. `TestConflictingRedelivery`.

### P1-4 / P1-6 — Break evidence and exception lifecycle

Only consumed break events are cited; unclassifiable breaks stay
unpaired with persisted exceptions; flags re-validate before clearing
(`deriveDayStatuses` recomputes from open blocking exceptions).
`TestMissingBreakEvidence`, `TestLateBreakAttach`,
`TestCorrectOnExceptionDay` (correction neither blocked by nor
clearing an unrelated open exception).

### P1-5 — Singapore business-date filtering

Covered under P0-2 (TZ-safe casts in generate + review).

### P1-7 / P1-24 — Runtime and migration role separation

- `MIGRATION_DATABASE_URL` honored by `migrate up` and `migrate
  status`; `Preflight` asserts `migration_admin` (+ extensions).
- `scripts/db/init-roles.sh` (compose) and `core/database/provision.go`
  (recycle) and `tests/dbutil` (suites) run the same grants.
- Every service suite opens its pool on the app DSN and hard-fails
  unless `current_user = 'app'` (7 suites + parity + integrity).
- Fresh-install role test: `TestRolesAreNotSuperuser`, migrate privilege
  checks, `TestDemoResetRecycles` (seed runs as app).

### P1-8 — Employee status transitions

Close-prior in-tx + transition matrix (terminated final, no
resurrection, no same-status edits, reason + operator mandatory,
effective date ordered). `TestSetStatus` (repeated transitions).

### P1-9 — Historical employment state and payroll population

Effective-dated population: terms overlapping the period UNION approved
sheet holders (terminated staff with approved history stay visible);
retroactive terms changes surface as term conflicts instead of silent
drift. `TestAssessTerminatedWithApprovedSheet`, `TestAssessTermConflict`.

### P1-10 / P1-11 — Roster publication validation and revision

`Publish` is an in-tx gate (non-empty, week window, status, membership,
skills, rest-day revalidated; shared `validatePlacement`). `Revise`
copy-forwards into a new draft revision and advances the pointer.
Revision rows stay insert-only; publication metadata lives in the
`roster.publish` audit event (`TestPublishMetadataInAudit`).
Concurrency: exactly-once publish, coherent publish-vs-revise race
(`TestConcurrentPublishExactlyOnce`,
`TestConcurrentPublishReviseCoherent`). Both exposed on CLI + REST.

### P1-12 — Attendance interval discovery

`GET …/attendance/intervals` + `attendance intervals`: bounds, lineage,
source refs, linked sheets (which decide correctability), date-level
open exceptions. `TestListIntervals`. The E2E retrieves every corrected
ID through this interface; no step uses SQL.

### P1-13 — Fresh-database test setup race

`WithSetupLock` serializes `MigrateUp`; 23505 tolerated only after
verifying the winner's object. Default `go test ./...` is reliable;
no serial fallback was taken.

### P1-14 — Rejected sheets dead-end

Covered under P0-1 (amendment-A versions; rejected originals immutable).

### §5A — Effective-dated employment terms

Atomic close-prior + successor, overlap exclusion, backdate rejection,
closed-history immutability, loud refusal when a new range intersects
an approved period. `TestTermsOverlap`, succession cases in
`hr_test.go`.

### §5C — Payroll-readiness exclusions

Run must exist for the same employer + period (else NotFound);
empty Included never approves (`AuthorizeApproval` fails closed);
empty/zero-hour approvals do not count; run status reported;
additive `expected`/`with_approved`/`included`/`excluded`/
`term_conflicts` sets. `TestAssessWithExclusion`,
`TestAssessUnknownRun`, `TestAssessRunWrongPeriod`,
`TestAssessEmptyPopulation`, approval predicate cases.

### §5D — Money representation

`JSON decimal string → Go int64 minor units → NUMERIC(12,2)`;
`float64` holds only measured hours/minutes, never money (verified by
grep — the only money-adjacent floats are comments). Invalid precision
and negatives return machine-readable validation, never PG errors. No
statutory rounding implemented.

### §5E — Transaction boundaries and audit

Import, pairing, correction + draft regen + aggregates, generation,
submit/approve/reject/reissue, assign, publish, revise, terms/status,
SHG, exclusions, period create: each commits atomically with its
aggregates and audit event. Replay-after-crash repairs pending work
(import re-pairs stored unpaired events; exceptions stay discoverable).
Audit covers every amendment-F action; audited mutations require real
operator attribution (`X-Operator` / `--operator` / local `$USER` —
never silent `api`/`cli`).

### §5F — Idempotency

Domain mechanisms, documented; no generic framework; the misleading
generic `Idempotency-Key` claim stays corrected to a variance note in
`docs/workforceops/api-cli-contract.md`. Deterministic replay verified:
import (dedup), publish (second publish → `invalid_state`, no second
revision), generate (existing sheets skip), approve (idempotent
re-approve), corrections (second successor rejected).

### §5G — API errors and contracts

One envelope `{error, message, request_id}` shared by CLI and REST
(spec `Error` schema aligned); request IDs flow from the echo
middleware; `FromDB` no longer leaks constraint/table/column names;
raw errors log server-side and render as safe `internal`. Malformed
dates return `validation`. Served routes ≡ implemented spec ops
(contract tests, 43 paths, kin-openapi SPEC-VALID).

### §5H — API and CLI workflow completeness

All minimum operations work without SQL on both surfaces (single-
employer selection via `--employer`/path; periods list+create; terms
list/set/compensate; skills require/grant; memberships; intervals
list + correct; roster revise; SHG set; payroll runs + exclusions).
Rotation materialization stays explicitly deferred (no generator
exists; assignment CRUD is catalogue only).

## 4. Contradictory review findings and their disposition

Stage 2 repro probes ran on disposable PG16 databases (scripts in
`/tmp/m11_stage2*.sh`, preserved for inspection, not committed):

| ID | Question | Disposition |
|----|----------|-------------|
| R1 | Overlapping periods double-allocate | CONFIRMED pre-fix (one interval linked into 2 sheets). Fixed: `pay_period_no_overlap` + single-chain trigger. Battery: overlap insert now 23P01; `TestCreatePeriod`. |
| R2 | Cross-midnight stamped one date | CONFIRMED pre-fix (`local_date` from in-event). Contract requires split. Fixed: midnight split citing shared boundary events. `TestCrossMidnightSplit`. |
| R3 | Late breaks attach | CONFIRMED pre-fix failure (zero intervals + orphans). Fixed: late-break attach supersedes into split successors. `TestLateBreakAttach`. |
| R4 | Concurrent corrections dual successors | CONFIRMED pre-fix race. Fixed: advisory lock + exclusion + single-successor index. `TestConcurrentCorrectionsSingleSuccessor`. |
| R5a/b/c | Fresh/upgrade/rollback | Fresh install clean (9 migrations + seed as app, re-proven in §8B bootstrap). Upgrade from pre-pivot partial history FAILS LOUDLY at v2 (`relation "employee" does not exist`) — no half-apply. Down/up SQL round-trip OK. Decision: pre-release fresh-install-only (no deployed user DB; nothing ever dropped). Documented in `docs/workforceops/database-design.md` §10. |
| R6 | Concurrent generate/import/publish | Safe: 1-wins-1-skips exactly-once (`TestConcurrentGenerateExactlyOnce`, `TestConcurrentImportsExactlyOnce`, `TestConcurrentPublishExactlyOnce`). |
| R7 | `seed demo --reset` wedge | CONFIRMED: published/decided rows are undeletable by every role, so row-level reset can never work. Fixed by redesign (§5): database-level recycle via explicit admin DSN. `TestDemoResetRecycles`. |
| R8 | `docker build` failure | CONFIRMED by inspection (3 of 4 `cmd/` paths deleted). Fixed: single-binary Dockerfile + single-process `fly.toml`. Docker daemon unavailable in this sandbox, so the exact Dockerfile build steps were executed directly (`CGO_ENABLED=0 go build … ./cmd/workforceops` → binary runs `version`). |

## 5. New defects discovered (this milestone)

1. **`timesheet` UNIQUE blocked append-only corrections.** The original
   `(employer, employee, period)` UNIQUE forbade a second correction
   sheet. Replaced with a partial unique on original sheets +
   `supersedes_id` + one-successor enforcement (amendment-A shape).
2. **No explicit current-interval predicate.** "Current" was inferred
   via `NOT EXISTS successor`. Now `is_current`, schema-maintained,
   keying the partial exclusion (amendment G).
3. **Duplicate-span resend failed the whole import.** Same clock span
   under fresh vendor IDs hit the live-overlap exclusion and aborted
   the batch. Now an explicit persisted `overlapping_span` exception:
   the import succeeds, nothing double-pairs, the operator sees the
   resend (`TestDuplicateSpanFreshIDs`). The code is intentionally
   non-blocking — the day's reviewed originals still generate.
4. **Seed reset wedged on frozen rows (R7).** Published rosters and
   decided timesheets cannot be row-deleted by any role, so the
   40-statement `resetEmployer` could never work. Replaced with
   database-level recycle: `seed demo --reset --admin-dsn …` drops,
   recreates, provisions and migrates the named disposable database
   (protected names refused, same-server enforced), then reloads.
   `TestDemoResetRecycles` seeds as app, decides records, recycles,
   reseeds. No trigger bypass, no new DELETE grants.
5. **Error envelope leaked internals and lacked request IDs.**
   `FromDB` exposed constraint/table/column names; `fail()` emitted no
   `request_id`; audited mutations silently defaulted to `api`/`cli`.
   Fixed (§3 §5G); spec `Error` schema aligned; operator required on
   mutations (`TestFailEnvelope`,
   `TestResolveRequiresOperatorOnMutations`).
6. **Workflow operations missing from both surfaces.** No interval
   listing, no reissue/revise exposure, no SHG capture, no pay-period
   create, no compensation exposure, no payroll-run/exclusion writes,
   no REST require-skills. All implemented with services + CLI + REST +
   spec + tests (§3 §5H).
7. **Dockerfile/Fly referenced deleted binaries (R8).** Fixed (§4).
8. **`Correct` misderived day records on cross-date moves (§6.7
   battery).** Records re-derived only for dates that already had a
   record row (the new date's record was never created, leaving
   monthly-excess attribution stale), and the old date was re-derived
   under the NEW assignment. Now both affected dates upsert with
   their own assignment. `TestCorrectCrossDateResolvesAssignment`
   pins the successor assignment, per-day minutes (0/90), and
   per-day record assignments.

## 6. Files and migrations changed

- Migrations: `migrations/core/00009_m11_integrity.sql` (new, 400+
  lines). `00001`–`00008` frozen and untouched.
- Schema/service core: `internal/attendance/attendance.go`,
  `internal/attendance/review.go` (pairing rewrite, corrections,
  intervals listing); `internal/timesheet/timesheet.go` (lifecycle,
  periods); `internal/payroll/readiness.go`, `approval.go`, `runs.go`
  (new); `internal/hr/hr.go`, `shg.go` (new); `internal/wfm/roster.go`
  (assign/publish/revise); `internal/seed/seed.go` (reset redesign);
  `internal/apperr/apperr.go` (sanitized mapping);
  `internal/httpapi/httpapi.go` (+11 routes, envelope, operator rule);
  `core/database/migrate.go` (preflight), `core/database/provision.go`
  (new); `cmd/workforceops/*` (+12 commands, operator rule);
  `tests/dbutil/dbutil.go` (roles, setup lock).
- Contract/docs: `openapi.yaml` (43 paths, aligned envelopes and list
  shapes, draft ops unserved), `docs/workforceops/api-cli-contract.md`
  (endpoint map, operator rule, variance notes),
  `docs/workforceops/database-design.md` (reset + pre-release policy),
  `Dockerfile`, `fly.toml` (single binary).
- Tests (98 `Test*` functions across 14 packages): new suites
  `shg_test.go`, `runs_test.go`, `period_test.go`,
  `intervals_test.go`, `envelope_test.go`, `seed_test.go`, plus
  M11/concurrency/regression additions in attendance, timesheet, wfm,
  payroll, hr, integrity and parity suites.

## 7. Database constraints and transaction changes

`00009_m11_integrity.sql` (append-only; nothing earlier altered):

- `pay_period_no_overlap` exclusion; `enforce_interval_single_chain`
  + `enforce_line_interval_employee` triggers;
  `timesheet_single_original_idx` + `timesheet_single_correction_idx`;
  `is_current` + SECURITY DEFINER maintenance + live-overlap partial
  GiST exclusion; `employment_terms_no_overlap`; same-chain
  `supersedes` guards; freeze re-parenting (NEW-parent checks);
  `attendance_exception` table (append-only for app); exclusion
  draft-only guards; status backfill.

Transaction boundaries: every mutating service commits evidence +
aggregates + audit atomically (import, pair, correct + draft regen,
generate-one, transitions, assign, publish, revise, terms/status,
SHG, exclusions, period create). Advisory-lock convention: one
transaction-scoped key per `(employer, employee)` —
`pg_advisory_xact_lock(hashtext('attendance:' || employer || ':' ||
employee))` — acquired in sorted employee order wherever multiple
employees lock; roster publish/revise serialize on `FOR UPDATE OF
roster`.

## 8. Test commands actually executed and results

Real PostgreSQL 16, disposable databases, app-role suites
(`WORKFORCEOPS_TEST_DATABASE_URL` naming `a1test`; seed-reset and
probes on dedicated names). Sandbox note: each shell call runs in a
private network namespace, so every DB command ran as
start-server → work → stop-server in a single call over
`127.0.0.1:5433`.

| Command | Result |
|---------|--------|
| `task lint` (gofmt gate + `go vet ./...` + lib/pq ban) | PASS |
| `task build` (single `build/workforceops`) | PASS |
| `go test -count=1 ./...` | PASS, exit 0, 14 packages ok, 0 failures |
| `go test -race` on attendance/timesheet/wfm/payroll/hr | PASS, no data races |
| Fresh-install migration (roles → `migrate up` → seed as app) | PASS, 9 applied |
| Upgrade-from-partial-history probe | Loud v2 failure, nothing half-applied (decision recorded, §4 R5) |
| `seed demo --reset` (`TestDemoResetRecycles` + CLI shape) | PASS |
| CLI + REST E2E (`/tmp/e2e_m11.sh`, §10) | E2E-PASS, exit 0, twice |
| OpenAPI validation (kin-openapi + contract tests) | SPEC-VALID; served ≡ implemented; drafts unserved |
| `git diff --check` | clean |

No suite skipped (DB-gated tests ran; no `Skip` without the DSN set),
no serial fallback, no narrowed runs.

## 9. Fresh-install and upgrade findings

- Fresh install works end to end: superuser bootstrap (extensions,
  roles, grants — identical SQL in `init-roles.sh`, `provision.go`
  and `dbutil`), `migrate up` as `migration_admin` (preflight
  enforced), `seed demo` as `app`. 9 migrations apply; seed loads.
- Upgrade from a pre-pivot partial history is unsupported and fails
  loudly at version 2 instead of half-applying (§4 R5). No deployed
  user database exists; the compose volume could not be inspected in
  this sandbox, so the report records that caveat and the standing
  rule: project tooling never drops an existing database (recycle
  refuses protected names and needs an explicit admin DSN).
- Pre-release decision (documented in `database-design.md` §10):
  fresh-install-only for M1.1; the upgrade question reopens only with
  a released database to migrate.

## 10. CLI/REST E2E evidence

`/tmp/e2e_m11.sh` (preserved, not committed): two fresh databases,
seeded identically, then driven through all 12 §8B steps — employee +
terms + compensation + SHG, programme/team/skills/membership,
template, roster create + publish, clock + break import, review,
interval-ID retrieval through the interface, correction, generate,
submit → reject → reissue → submit → approve, run + exclusion,
readiness with and without the run. Zero SQL intervention; every ID
the flow needs comes from a prior CLI/REST response.

- `CLI-LEG-OK`, `REST-LEG-OK` (REST served on loopback `:18081`).
- Persisted-state diff: 15 normalized dumps (employees, terms,
  compensation, SHG, memberships, skills, rosters, assignments,
  intervals, exceptions, sheets, lines, lineage, exclusions, audit
  counts — 137 lines, UUIDs/timestamps/operators normalized out):
  `PERSISTED-DIFF-EMPTY`.
- Readiness diff on business fields: `READINESS-DIFF-EMPTY`. Observed
  verdict: `not ready: 4 blocked, 2 missing input, 0 term conflicts` —
  E2E-001 included and blocked on all four packs (honest BLOCKED
  state), DEMO-002 visibly excluded, DEMO-001/003 missing approvals.

## 11. Remaining limitations and untested hypotheses

- Docker daemon unavailable here: `docker build` itself unrun; the
  exact Dockerfile steps were executed directly and the binary runs.
  Re-run `docker build` where a daemon exists.
- Compose volume contents unverified (daemon unavailable); the
  no-deployed-database finding carries that caveat. Nothing was
  dropped regardless.
- Rotation materialization explicitly deferred (assignment CRUD only;
  no roster generator; documented in contract + report).
- Generic `Idempotency-Key` infrastructure deferred; domain replay
  mechanisms cover the five critical mutations.
- Leave, KET, payslips, payroll register, rule-packs: draft spec ops,
  unserved, contract-enforced.
- `overlapping_span` is advisory (non-blocking) by design: a vendor
  resend never stalls generation on reviewed originals; the operator
  dispositions it via correction.
- Concurrent-publish loser shape is `invalid_state` (deterministic,
  no duplicate revision) — not silent idempotence; callers retry
  reads, not writes.
- OT-001 stays BLOCKED on: SSO legislation text (403), threshold
  effective date, 13.636→13.60 rounding rule/stage, Fourth Schedule
  pinning, 72h-breach consequences. No calculations were built.
- Seed `DEMO-002` carries an intentionally unpaired break event
  (missing-evidence fixture); E2E/readiness expectations account
  for it.

## 12. Final M1.1 PASS/FAIL verdict — PASS

Against prompt9 §8D:

1. Both P0 defects corrected (amendment-A versions; roster-anchored
   pairing + compliance-only 72h). ✔
2. All 13 M1-blocking P1 findings corrected (§3; none merely
   disproved — each has a schema/service fix and a test). ✔
3. Prior-report discrepancies adjudicated (R1–R8 table, §4). ✔
4. No remaining P0/P1 makes attendance, timesheet or readiness
   untrustworthy (frozen records immutable; evidence uncapped and
   attributable; readiness explicit). ✔
5. Runtime and migration roles work on a fresh install (§9; app-role
   suites; preflight). ✔
6. Correction and recovery workflows work without SQL (§10 E2E). ✔
7. Default test execution is reliable (`go test ./...` exit 0;
   setup race fixed; no serial fallback). ✔
8. CLI and REST produce equivalent persisted business results
   (byte-identical 137-line dump diff). ✔
9. `SG-MOM-OT-001` remains BLOCKED (manifest untouched,
   `status: BLOCKED`). ✔
10. No statutory payroll calculations or fabricated golden fixtures
    introduced (grep-clean; readiness computes no money). ✔

## 13. Smallest next implementation milestone

Payroll-run lifecycle execution (no calculations yet): run
`draft → validated → approved` transitions with exclusion freeze at
validation, approval gated on the readiness verdict, and run-status
history — the immediate next workflow step after M1.1's exclusion
container, and the last gate before any calculation slice may begin.
After that: one IMPLEMENTABLE pack's deterministic calculation slice
behind pinned rule versions.
