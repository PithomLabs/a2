# Milestone 1 Adversarial Review — WorkforceOps Lab

Synthesized from 8 scoped child reviews (prompt8 sections A–L, §3 tests, §4
limitations, §5 Phys boundaries, plus an independent critic gap analysis).
This synthesis performed no new code edits and ran no new test suites; all
executed results below are attributed to the child that ran them. Severity
was normalized against prompt8 §6 (see dedup map); disagreements between
children are adjudicated inline rather than averaged.

## 1. Scope and Git baseline

- **Scope:** prompt8 A (architecture) through L (security), §3 adversarial
  test review, §4 known limitations, §5 Phys-Research/Phys-Pseudo boundaries.
  Out of scope: `bchat/` (read-only reference, untouched), statutory rule
  content beyond boundary checks, performance benchmarking.
- **Git baseline:** HEAD `f435158` ("Phase 1 scaffold baseline (pre-pivot)").
  The M1 implementation is **uncommitted working-tree state**: 79
  changed/untracked entries (12 modified, 4 deleted, 35 renamed, 28
  untracked; `git diff --stat`: 51 files, +1231/−636). All findings were
  raised against this dirty tree, not against a tagged commit. Any fix
  verification must re-confirm the tree state first.
- **Prior-report inputs:** M1 implementation report incl. Amendments B/C,
  E2E §10 12-step workflow claim, and 7 acknowledged known limitations
  (assessed in §6).

## 2. Methodology and actual commands executed

Read-only inspection by 7 scoped reviewers (all cited file/line bodies were
opened and read), one live-execution pass on disposable infrastructure, and
one critic pass. Commands actually executed (attributed):

- Live execution (result 7, disposable TCP-only PG16 scratch clusters, repo
  unmodified, clusters stopped after use):
  - `task lint` → FAIL environmentally (parent `/home/chaschel/Documents/go/go.work`
    breaks `go vet ./...`); with `GOWORK=off`: gofmt clean, `go vet` clean,
    `lib/pq` absent → PASS.
  - `task build` (`GOWORK=off`) → PASS (`build/workforceops`, 19 MB).
  - `GOWORK=off go test -count=1 ./...` without DB → PASS (DB-gated skip).
  - Fresh DB + default parallel `go test -count=1 ./...` → FAIL exit 1, 18
    setup-race failures (23505), reproduced twice; same-DB rerun → PASS;
    `-p 1` serial on fresh DB → PASS, all 20 packages incl. integrity+parity.
  - OpenAPI structural check (`/tmp/scope7_openapi.py`): 36 paths, 38
    implemented + 8 draft ops, 46 unique operationIds, all local `$ref`s
    resolve → STRUCT_OK.
  - Fresh DB: `migrate up` → "applied 8 migration(s)", status 1–8 applied;
    `seed demo` → "seeded MERIDIAN: 3 employees, 1 programme, 1 roster week,
    1 pay period"; CLI employee list showed business rows; top-level `demo`
    → "unknown command".
- API/CLI probing (result 4): `GOWORK=off go test ./internal/httpapi/
  ./core/api/` → ok; `./build/workforceops --help` + subcommand helps;
  rotation-pattern shadowing probes; spec counts via grep (36/46/38/8/46).
- Build/contract probes: `GOWORK=off go build ./cmd/workforceops` → ok
  (results 1, 5); `GOWORK=off go build ./internal/hr/ ./internal/wfm/` → ok
  and `go test -run TestServedSubsetOfSpec|TestSpecImplementedAllServed
  ./internal/httpapi/` → pass (result 6); `GOWORK=off go build ./...` +
  httpapi contract tests → pass (result 3); `go vet` on
  timesheet/payroll/attendance + `TestAuthorizeApproval` → pass (result 2);
  `go build ./cmd/shifthub-api|payloop-api|sync-worker` → "directory not
  found" (result 1); `git diff HEAD --check` clean (result 1).
- Explicitly **not** executed by any child: live concurrent-race runs,
  live 404/405/prod-body probes, browser CORS, `docker build`, `seed demo
  --reset`, rollback (`Down`), upgrade from baseline `f435158`, full E2E
  12-step re-run (see §10).

## 3. Independent test results (executed, disposable PG16 unless noted)

| Check | Result |
|---|---|
| `task build` (`GOWORK=off`) | PASS, 19 MB binary |
| Lint (`GOWORK=off`) | PASS (gofmt + vet clean, no `lib/pq`); bare `task lint` FAILS on parent `go.work` env issue |
| `go test ./...`, no DB | PASS, DB-gated suites skip |
| `go test ./...`, fresh DB, parallel | **FAIL exit 1, 18 setup races** (SYN-P1-13); rerun PASS |
| `go test ./... -p 1`, fresh DB | PASS, 20 pkgs incl. integrity + parity |
| `go test ./internal/httpapi/ ./core/api/` | PASS (presence-only, see SYN-P2-26) |
| Fresh `migrate up` + `seed demo` + CLI list | PASS (8 migrations, MERIDIAN rows visible) |
| OpenAPI structural validation | PASS as YAML (shape-vs-implementation mismatches are SYN-P2-1/SYN-P2-2, not YAML defects) |
| Integrity suite (serial, disposable DB) | PASS incl. freeze/isolation/lineage matrices |
| `rules/` Go purity/quarantine arch tests | N/A — `rules/` holds only markdown/fixtures (no Go yet); future work per `plans/plan4.2`, not a current gap |

## 4. Ranked findings (prompt8 §7 format)

Format per finding: ID | severity | file:line | violated invariant |
failure scenario | why tests missed it | repro | fix | blocks-M1.
"Blocks-M1" means M1 acceptance; all others are must/should-fix-before-M2
(see §9) or deferred.

### Dedup and severity-normalization map

Raw input held ~90 entries with 9 duplicates, all merged below (provenance
kept in brackets): correction-after-approval [F-01+G2] → SYN-P0-1 (P0 kept:
approved records silently diverge); interval-listing [I-01+F-04] →
SYN-P1-12; role separation [r5-F2+r1-F2] → SYN-P1-7; import race
[E-05+I-06] → SYN-P1-20; approve race [G5+F13] → SYN-P2-7 (P2: spurious
error only, no duplicate records); multi-grant [D6+F12] → SYN-P2-10 (P2:
setup friction + SQL recovery, not a broken core transition);
spec-shape [S7-02+I-04] → SYN-P2-1 (**P1→P2**: operations work consistently
across adapters; spec-conformant clients break — a documentation/API
inconsistency per §6, not an unusable operation); envelope+leak
[I-02+F9+L-01] → SYN-P2-2 (P2); attribution [r5-F15+I-07+E-10] → SYN-P2-16
(P2); run-id [H1+F7] → SYN-P2-8; atomicity [F-05+r5-F4+r5-F5] → SYN-P2-4;
malformed-input-500 [D10+C4+E-13a] → SYN-P2-14; contract-test thinness
[I-08+S7-07] → SYN-P2-26; stale tasks [S7-09+F4] → SYN-P3-12 (**P2→P3**:
fails loudly, cosmetic); transient exceptions [E-06] → SYN-P2-17
(**P1→P2**: discoverability gap, correction itself is the P1 in SYN-P1-12);
unknown-employee drop [E-09] stays P2 (vendor can redeliver; system-side
loss only, not irrecoverable evidence loss).

### 4a. Confirmed defects — P0 (critical)

**SYN-P0-1 | P0 | `internal/attendance/review.go:147-251` Correct ×
`internal/timesheet/timesheet.go:289-295` + `migrations/core/00004_protection.sql:32-80`**
| Invariant: approved timesheets never silently diverge from source
evidence; post-approval change needs a new auditable process. | Scenario:
Correct() never queries `timesheet_line_interval`; correcting an interval
already linked to an approved (or submitted) sheet succeeds, inserts a
superseding interval, recomputes day aggregates, while the approved line
keeps old hours citing a dead interval id. Readiness only checks
status=approved, so payroll proceeds on stale hours with no flag anywhere.
| Missed: correction tests run before any timesheet exists; freeze matrix
covers timesheet-row edits, not interval supersede under a decided sheet;
no cross-service test. | Repro (disposable DB): Generate→Submit→Approve,
then Correct a linked interval 18:00→12:00; sheet still shows old hours.
| Fix: reject Correct when the interval links to a non-draft timesheet
(join via `timesheet_line_interval`); add a reissue/regeneration flow.
| blocks-M1: **yes**.

**SYN-P0-2 | P0 | `internal/attendance/attendance.go:256-333` pairSpans,
`:340-405` storeSpan** | Invariant: unbounded spans raise an exception,
never become clean payable time. | Scenario: orphan `in` Aug-1 pairs with
month-late `out` Aug-31 into one ~696h `work` interval, local_date=Aug-1,
NO exception (no span-length guard; `ExcLimit72hBreach`/`Exc72hNeedsReview`
defined at :37-38 but never emitted). Day is `reviewed`, so Generate
includes ~696 REG hours approvable as-is; the monthly-excess flag is never
consulted by checkEligible. | Missed: fixtures use same-day spans only; no
late-out/span-length/72h test. | Repro: import `in` Aug-01 09:00+08 alone,
then `out` Aug-31 18:00+08; inspect interval duration and generated hours.
| Fix: cap the pairing window (leave orphans unpaired) and emit a
span-breach exception; wire monthly NeedsReview into checkEligible as skip.
| blocks-M1: **yes**.

### 4b. Confirmed defects — P1 (high), blocking M1

**SYN-P1-1 | P1 | `migrations/core/00006_trigger_return_fix.sql:23-103`
freeze triggers** | Invariant: approved/published artifacts immutable;
corrections only via new records. | Scenario: UPDATE re-parenting bypasses
freeze — non-INSERT branches check the OLD parent only, so `UPDATE
timesheet_line SET timesheet_id=<approved>` on a draft line, or
`payroll_line.run_id`/ `roster_assignment.revision_id` moves onto frozen
parents, succeed (same-employer FK still holds). Exploitable only via raw
SQL with DB credentials (no service path); `forbid_unless_run_draft`
(:105-131) already checks both parents, proving the pattern was known.
| Missed: freeze matrices test in-place UPDATE/DELETE/INSERT only, never an
UPDATE changing the parent FK. | Repro: disposable DB, approve sheet A,
`UPDATE timesheet_line SET timesheet_id=A` on a draft line → expect P0001,
succeeds. | Fix: on UPDATE also resolve the NEW parent and require
mutable, mirroring `:122-127`; add re-parenting freeze cases.
| blocks-M1: **yes**.

**SYN-P1-2 | P1 | `internal/timesheet/timesheet.go:251-276` generateOne ×
`migrations/core/00002_attendance_payroll.sql:9-38` (no overlap guard)**
| Invariant: one interval counted once; no double-count. | Scenario:
duplicate vendor feed under different `source_event_id`s (dedup key is
only `(employer,source,source_event_id)`) creates two identical 9h
intervals; Generate sums both into an 18h REG line that submits/approves
cleanly. Same via a correction overlapping siblings (see SYN-P1-21).
| Missed: dedup tests cover same-ID replay only; no non-overlap or
new-ID re-import assertion. | Repro: import same in/out twice with
different ids, generate, observe doubled hours. | Fix: gist exclusion on
`tstzrange(start,end)` per (employer,employee) for current rows +
service-level overlap checks in Generate/Correct + regression test.
| blocks-M1: **yes**.

**SYN-P1-3 | P1 | `internal/attendance/attendance.go:102-114` ON CONFLICT
DO NOTHING** | Invariant: raw vendor evidence preserved; conflicting
redelivery surfaces, never silently dropped. | Scenario: re-sending an
event id with DIFFERENT kind/time/employee (vendor correction or id
collision) counts Duplicates++ and is discarded with no comparison or
conflict signal; a vendor time-correction is ignored forever.
| Missed: replay test re-sends byte-identical CSV only. | Repro: import
evt-1 as `in` 09:00, re-import as `out` 18:00 → Duplicates=1, original
retained, no error. | Fix: on conflict fetch the row, compare
(employee,kind,occurred_at), emit a conflict exception/validation error on
mismatch. | blocks-M1: **yes**.

**SYN-P1-4 | P1 | `internal/attendance/attendance.go:361-396` storeSpan**
| Invariant: missing information never converts to zero hours; paired
status means evidenced. | Scenario: an unclassifiable break (no roster /
no segment / zero-dual overlap → ExcAmbiguousBreak) gets NO interval, yet
workRuns splits around it so break minutes vanish from actual+paid; worse,
ALL break ids are cited in the WORK intervals' refs (:372-375), marking
the unclassifiable events paired — only trace is the ephemeral import
Report plus a codeless flag (see SYN-P1-6). | Missed: fixture always
provides matching roster + 1h unpaid break; no no-roster/no-segment test.
| Repro: import in/out + break_start/end with no roster for the date; no
break interval, events absent from unpaired list, actual_min short.
| Fix: do not cite unclassifiable break ids in work refs; persist the
exception against the day record. | blocks-M1: **yes**.

**SYN-P1-5 | P1 | `internal/timesheet/timesheet.go:196-201` checkEligible +
`internal/attendance/review.go:85-92`** | Invariant: employees with
unpaired events in the period are excluded from generation. | Scenario:
both queries filter `occurred_at::date`, casting timestamptz in SESSION
timezone; no `SET TIME ZONE` exists (`core/database/db.go` has none), so
server default (UTC) governs. Unpaired `in` Aug-01 00:30+08 (= Jul-31
16:30Z) casts to Jul-31, falls outside August, and is NOT counted →
employee generated as clean. Reverse mis-attribution at the other edge.
| Missed: fixtures use mid-day +08 events where UTC/SG dates coincide; no
near-midnight case. | Repro: lone `in` 2026-08-01T00:30:00+08:00, Generate
August → employee generated instead of skipped. | Fix: compare business
dates in the event's own zone (`(occurred_at AT TIME ZONE
timezone)::date`) in both queries. | blocks-M1: **yes**.

**SYN-P1-6 | P1 | `internal/attendance/attendance.go:545-554` upsertRecord
× `review.go:242-249`** | Invariant: exception flags persist until the
cause is resolved. | Scenario: upsertRecord unconditionally upserts
status=reviewed; Import re-flags only the CURRENT import's dates, and
Correct() calls upsertRecord after every correction — so any correction on
an exception-flagged day (e.g. the SYN-P1-4 gap, which has no interval to
correct) silently clears it to reviewed, making the day generatable with
dropped minutes and no remaining flag. | Missed: no test asserts
exception-status survival or post-correction re-validation. | Repro:
flag an ambiguous-break day, Correct() an unrelated bound that day,
observe status=reviewed and Generate including the employee. | Fix:
re-run day validation on correction; only clear resolved causes; never
reset exception→reviewed without a validation pass. | blocks-M1: **yes**.

**SYN-P1-7 | P1 | `cmd/workforceops/migrate.go:57-63` mustDSN ×
`migrations/core/00004_protection.sql:15-30` × `scripts/db/init-roles.sh:35-36`
× `.env.example:17-19`** | Invariant: least-privilege insert-only history
+ migration_admin/app role separation hold on fresh install; runtime uses
the unprivileged app role. | Scenario (threefold silent degradation):
(1) `MIGRATION_DATABASE_URL` is documented but read NOWHERE in code, so
`task db:migrate` with default `.env` runs as superuser `a1`; tables end
up owner `a1`, `ALTER DEFAULT PRIVILEGES FOR ROLE migration_admin` never
fires, and app gets USAGE-only — while dev serve/seed also run as
superuser, under which the 00004 REVOKEs never apply (owner bypass), so
insert-only is unenforced in the documented dev path. (2) 00004 REVOKEs
are conditional (`IF EXISTS app`): installs without pre-created roles
(e.g. Neon, explicitly excluded in init-roles.sh:3) skip revokes silently
and history stays writable forever. (3) `tests/integrity/setup_test.go:105-123`
hand-grants/hand-revokes, so the suite passes while production diverges.
Net: fresh installs either break as app (42501 surfaced as internal) or
run without the protection layer, with no migrate-time warning.
| Missed: no test exercises the real fresh-install path or asserts
role/owner/grants afterward; service suites all use the superuser DSN
(see SYN-P1-24). | Repro: fresh PG16 without roles, MigrateUp, inspect
`role_table_grants` for `clock_event` → missing revokes/grants depending
on owner. | Fix: assert `current_user='migration_admin'` + app existence
at 00004 start (fail loudly); honor `MIGRATION_DATABASE_URL` in migrate;
add post-migrate grant/revoke verification; document per-command roles.
| blocks-M1: **yes**.

**SYN-P1-8 | P1 | `internal/hr/hr.go:118-158` SetStatus ×
`migrations/core/00004_protection.sql:212-217`** | Invariant: repeated
status transitions recordable; exclusion needs non-overlapping ranges.
| Scenario: SetStatus always inserts `effective_to=NULL` (=infinity) and
never closes the prior row (no trigger does either). First SetStatus
succeeds; ANY second one inserts an overlapping [from,infinity] range →
23P01 conflict. Suspend→reinstate and suspend→terminate are impossible.
| Missed: TestSetStatus changes status exactly once. | Repro: create
employee, suspend 2026-08-01 ok, re-activate 2026-08-02 → overlap
conflict. | Fix: close the prior open row in the same tx before insert;
add a double-transition test. | blocks-M1: **yes**.

**SYN-P1-9 | P1 | `internal/hr/hr.go:118-226` (no timesheet consultation) ×
`internal/timesheet/timesheet.go:82` × `internal/payroll/readiness.go:87,224-228`**
| Invariant: approved artifacts stable; status/terms changes must not
silently rewrite approved-timesheet consequences. | Scenario: (a) approve
a sheet, then SetStatus terminated (possibly backdated) → Assess lists
status=active only, so the employee silently VANISHES from population with
no excluded/blocked/missing entry; (b) SetTerms inserts a historic row
covering an approved period → readiness re-evaluates terms-at-period-start
at assessment time (not snapshotted at approval), flipping verdicts
post-approval. SetTerms also accepts terms for terminated staff.
| Missed: no HR test touches timesheets; readiness fixtures fixed; no
retroactive-change test. | Repro: seed→generate→approve→SetStatus
terminated→Assess shows a smaller population, no trace of the approved
employee. | Fix: snapshot terms/status at approval, or union Assess
population with approved-sheet holders and flag mismatches as blockers;
reject backdated changes overlapping approved periods without a
correction record. | blocks-M1: **yes**.

**SYN-P1-10 | P1 | `internal/wfm/roster.go:376-396` Publish** | Invariant:
all constraints validated before publishing; publication must not bless
invalid/empty schedules. | Scenario: Publish is a single UPDATE with zero
assignment checks — empty rosters publish (proven by
parity_test.go:191-204), as do rosters whose skills/membership/status went
stale after Assign, and out-of-week dates. No coverage/rest/skill gate
exists (`programme_staffing_requirement` has no readers, `terms.rest_day`
never read, rotation off-days never consulted). | Missed: publish test
uses one valid 2-assignment roster; no negative publish tests. | Repro:
create roster, publish with zero assignments → 200 published; or Assign,
raise RequireSkills, publish → 200 despite unqualified assignee. | Fix:
in one tx re-run membership+skill+status checks per assignment/date,
reject empty rosters (or allow with a reason code), record the check
result; add negative tests. | blocks-M1: **yes**.

**SYN-P1-11 | P1 | `internal/wfm/roster.go:188` (v1 only), `:248` (error
promises nonexistent reopen)** | Invariant: corrections supersede via new
auditable records; published rosters remain correctable. | Scenario: after
Publish the frozen trigger rejects every assignment change, and NO
service/CLI/route creates v2 or reopens to draft → roster permanently
frozen; the only repair is raw SQL (as isolation_test.go:160-198 does),
which doesn't even carry v1 assignments forward (v2 starts empty; Show
renders current revision only). | Missed: revision behavior tested only
via raw SQL with app role; no service-level revise test exists. | Repro:
publish, then attempt any correction via CLI/REST → frozen error, no
revise command. | Fix: Revise service (reopen + vN+1 carrying copied
assignments + pointer advance, atomically) on CLI+REST; forbid
assignment-orphaning pointer moves. | blocks-M1: **yes**.

**SYN-P1-12 | P1 | `internal/httpapi/httpapi.go:380-410` ×
`internal/attendance/review.go:52-56,147`** | Invariant: every implemented
route genuinely usable; the review→correct loop completable without SQL.
| Scenario: GET /attendance returns records/unpaired/monthly-excess with NO
interval ids, but POST corrections requires `interval_id` — an API-only
operator can never correct. Timesheet Get exposes ids only post-generation,
and generation skips exactly the exception employees needing correction.
| Missed: contract tests assert route presence only; parity skips
attendance; E2E step 8 itself used SQL for the id. | Repro: seed demo,
GET attendance → no interval ids → POST correction unobtainable. | Fix:
GET intervals endpoint (employer+employee+date, current-vs-superseded)
before M2. | blocks-M1: **yes** (see LIM-3).

**SYN-P1-13 | P1 | `tests/dbutil/dbutil.go:21-72` CreateDatabase +
EnsureExtensions** | Invariant: documented `go test -count=1 ./...` passes
on a fresh disposable DB. | Scenario: parallel packages race on CREATE
DATABASE/EXTENSION/MigrateUp; only 42P04/55006 tolerated, but concurrent
CREATEs raise 23505 on `pg_database`/`pg_extension`/`pg_type` → **18 setup
failures, exit 1, reproduced twice live**; rerun and `-p 1` serial pass.
| Missed: race manifests only on first run against an empty server.
| Repro (executed): fresh PG16 + `GOWORK=off go test -count=1 ./...` →
exit 1 with 18 setup FAILs. | Fix: short-term `-p 1` in `task test:db`
or a file lock; long-term tolerate 23505 collisions + advisory-lock
MigrateUp. | blocks-M1: **yes** (with trivial workaround).

### 4c. Confirmed defects — P1 (high), not blocking M1

**SYN-P1-14 [G3] | P1 | `internal/timesheet/timesheet.go:375-403`
Reject/transition × `00004_protection.sql:34-44`** | Invariant: rejected
sheets correctable/resubmittable. | Scenario: Reject moves
submitted→rejected with no way back; the trigger freezes rejected exactly
like approved, and Generate maps the UNIQUE conflict to skip "already
exists". Operator rejects a bad sheet → employee can never get a timesheet
for that period (regenerate skips, resubmit impossible, SQL DELETE
forbidden). Terminal dead end. | Missed: zero reject/resubmit coverage.
| Repro: Generate→Submit→Reject→Generate → Skipped; Submit/Approve on the
rejected id → invalid_state. | Fix: rejected→draft resubmit transition
with a trigger exemption for that path (or void-and-regenerate) + test.
| blocks-M1: no.

**SYN-P1-15 [D3] | P1 | `internal/wfm/roster.go:222-322` Assign**
| Invariant: assignment validation covers status, week scope, rest, duplicates.
| Scenario: (a) no employee-status check → resigned/terminated/suspended
staff assignable, then silently skipped by Generate; (b) no
business_date-within-week check → far-future dates assignable to any
roster; (c) `terms.rest_day` and rotation off-days never consulted →
rest-day staff rostered on rest days; (d) duplicate same-date relies on DB
unique only, untested, returns conflict instead of idempotent success.
Checks run outside any tx (see SYN-P2-6). | Missed: skill-present/level
tests only; no negative assign tests. | Repro: assign a terminated
employee / far-future date / Sunday shift for Sunday-rest terms → all 201.
| Fix: status-active check (or reason override), week-window check,
rest-day cross-checks (or explicit descope), duplicate-date test, tx-wrap.
| blocks-M1: no.

**SYN-P1-16 [D4] | P1 | `internal/wfm/roster.go:53-100` CreateTemplate**
| Invariant: shift templates internally consistent. | Scenario:
overlapping segments, segments past shift length, kind/paid incoherence
(no DB check), and contradictory cross-midnight flags (09:00–18:00 +
cross_midnight=true) all accepted; downstream break/attendance logic
consumes the garbage. | Missed: only valid templates ever created in
tests. | Repro: POST overlapping work segments → 201. | Fix: validate
non-overlap, containment within span, kind/paid coherence,
start/end-vs-flag consistency + negative tests. | blocks-M1: no.

**SYN-P1-17 [D5] | P1 | `internal/wfm/wfm.go:188` AddMembership, `:261`
AssignSkills, `rotation.go:126` CreateRotationAssign** | Invariant: every
validity end the validator reads must be settable; succession possible.
| Scenario: no `effective_to`/`valid_to` params — expired paths reachable
only via SQL; stale memberships grant permanent access after transfers;
second-ever grant for the same key always 23P01-conflicts, so level
upgrades, re-grants, pattern succession are impossible via API.
| Missed: no expiry/succession tests; rotation test assigns once. | Repro:
AssignSkills L3 then L4 same skill → conflict. | Fix: end-date params (or
close/renew endpoints) + close-prior-on-insert in a tx + tests.
| blocks-M1: no.

**SYN-P1-18 [C2] | P1 | `internal/hr/hr.go:119-123` SetStatus**
| Invariant: terminal statuses terminal. | Scenario: any→any transitions
accepted; terminated staff flipped back to active with only a reason
string, resurrecting them into Generate and readiness population.
| Missed: only active→suspended tested. | Repro: terminate then re-activate
via CLI → succeeds. | Fix: transition matrix (active↔suspended,
→resigned/terminated, terminals final) + edge tests. | blocks-M1: no.

**SYN-P1-19 [C3] | P1 | `internal/hr/hr.go:169,182,255` float64 money**
| Invariant: no binary floating point for money (AGENTS.md #3). | Scenario:
monthly_basic/amount_monthly travel Go↔JSON↔pgx as float64 (2600.10
unrepresentable); >2dp inputs silently rounded by NUMERIC(12,2); negatives
caught only by DB CHECK (23514), not 422 validation; future payroll calc
inherits corrupt cents. | Missed: tests use whole 2600 only; no
fractional/negative/precision tests. | Repro: SetTerms 2600.105 / −5 →
rounding / DB-check error instead of 422. | Fix: string/decimal (or int64
cents) input, Go validation (≥0, ≤2dp), keep NUMERIC(12,2). | blocks-M1: no.

**SYN-P1-20 [E-05+I-06] | P1 | `internal/attendance/attendance.go:75-134`
Import (no txn/lock) + `review.go:186-202` (no UNIQUE on supersedes_id)**
| Invariant: identical source never double-counts; one live successor per
supersede. | Scenario: (a) two concurrent Imports both read the same unpaired
set and both insert intervals citing the same clock events → duplicate
work intervals, double-paid, no DB constraint prevents it; (b) two
concurrent Correct() both pass the EXISTS(superseded) check and insert two
live successors (no unique index); upsertRecord sums BOTH. Sequential paths
correct; only the races open. | Missed: no concurrency test anywhere.
| Repro (not executed live): parallel imports / parallel Corrects, count
live intervals. | Fix: per-(employer,employee) advisory lock or
serializable pairing txn; UNIQUE index on `supersedes_id`. | blocks-M1: no.

**SYN-P1-21 [F-03] | P1 | `internal/attendance/review.go:147-251` Correct
(no overlap/roster validation)** | Invariant: corrections maintain correct
relationships; no double-count. | Scenario: Correct accepts arbitrary
bounds and any kind with no overlap check against live siblings and no
break re-classification — overlapping corrections double-count in sums (no
exclusion constraint); kind flips work→break_paid bypass the roster rule
storeSpan enforces. | Missed: correction test changes one end bound
same-day only. | Repro: Correct a work run past the break into its sibling
→ summed minutes exceed wall time. | Fix: reject sibling-overlapping
corrections; re-classify break kinds against roster. | blocks-M1: no.

**SYN-P1-22 [J-01] | P1 | `cmd/workforceops/wfm.go:326-364` rotationCmd**
| Invariant: delivered CLI surface works as reported (rotation pattern
create|list). | Scenario: both subcommands register cobra Name `pattern`;
Find resolves the first, so `rotation pattern list` is UNREACHABLE
(`... pattern list --help` prints CREATE usage). Patterns creatable but
never listable without SQL. | Missed: no CLI help/invocation tests.
| Repro (observed): `rotation pattern list --help` → Usage shows `pattern
create`. | Fix: parent `pattern` command with create|list children.
| blocks-M1: no (rotations inert in M1; promised surface broken).

**SYN-P1-23 [J-02] | P1 | `cmd/workforceops/hr.go:9-24`, `seed.go` direct
SQL, tests' SQL setup** | Invariant: fresh-install workflow without
undocumented SQL. | Scenario: no employer create, pay-period create,
SHG-instruction create, compensation add, or activation/exclusion
management in CLI or REST — a non-seed employer cannot be onboarded and
monthly periods/SHG rows need manual SQL (E2E-001 pattern: readiness stuck
at missing_info without SQL). | Missed: tests/E2E use seed SQL setup, never
CLI-only onboarding. | Repro: migrate empty DB → no command/route creates
an employer or period. | Fix: add the five management ops (or document
interim SQL) before M2. | blocks-M1: no.

**SYN-P1-24 [S7-03] | P1 | all service/parity `testPool`s vs
`tests/integrity/setup_test.go:137-145`** | Invariant: service tests run
with production-like least privilege so permission regressions fail loudly.
| Scenario: every service/parity suite connects with the superuser DSN and
never asserts role — a service change that UPDATEs/DELETEs an insert-only
table PASSES all service tests yet fails in production as app. Kept P1
(not downgraded to a test-coverage P2) because it hides production
failures of the SYN-P1-7 invariant. | Missed: privilege coverage lives
only in raw-SQL freeze tests, never through the called service layer.
| Repro: restrict a service-test role to app grants; observe which paths
break (currently unexercised). | Fix: run service/parity suites as app
(reuse integrity bootstrap) or add per-mutation permission tests.
| blocks-M1: no.

### 4d. Confirmed defects — P2 (medium)

**SYN-P2-1 [S7-02+I-04] | P2 | all List handlers (`httpapi.go`, e.g.
`:145-155`) vs `openapi.yaml` Paginated** | Invariant: served responses
match spec schemas. | Scenario: every list handler returns a bare JSON
array while the spec declares Paginated `{data,page,page_size,total}` —
any spec-conformant client breaks; responses unbounded, no page params
read. Downgraded P1→P2: operations work consistently across CLI/REST for
M1's own consumers; this is a documentation/API inconsistency (§6), but
every new consumer hardcodes the wrong shape, so cost grows. | Missed:
contract tests assert route presence/maturity only; parity_test even works
around the bare array. | Repro (observed live GET): `/programmes` → `[...]`.
| Fix: envelope + `?page&page_size` in the first M2 slice; add shape
assertions. | blocks-M1: no.

**SYN-P2-2 [I-02+F9+L-01] | P2 | `internal/httpapi/httpapi.go:573-582`
fail() × `cmd/workforceops/cli.go:84-92` out() ×
`internal/apperr/apperr.go:63-92` FromDB** | Invariant: single envelope
`{error:message, code:machine_code}` + request IDs; no internals leak.
| Scenario: (a) both adapters emit `{error:<CODE>,message:<text>}` —
inverted vs AGENTS.md and vs openapi Error (which wants
`{error,code,request_id}`, no `message` key); (b) `request_id` provisioned
by middleware but never included in bodies; (c) FromDB default embeds raw
pg text + constraint/column names into CodeInternal, returned verbatim even
in prod (prod HTTPErrorHandler covers echo.HTTPError only, not apperr) —
e.g. 42501 discloses table names. | Missed: zero envelope-shape/status
tests; no 42501-through-FromDB test. | Repro: curl any failing endpoint,
compare keys to openapi Error; force 42501 as app → table name in message.
| Fix: `{error:msg, code:code, request_id:rid}` in both adapters; generic
internal message + server-side logging. | blocks-M1: no.

**SYN-P2-3 [I-05+C7] | P2 | `internal/httpapi/httpapi.go` (no
Idempotency-Key read) × `internal/hr/hr.go:48-79`** | Invariant:
Idempotency-Key on publish/approve/import with replay-original semantics;
safe retry succeeds. | Scenario: retried roster publish after a timeout →
422 invalid_state, indistinguishable from real failure; `idempotency_key`
table exists but NOTHING writes it (dead table); duplicate employee_no
retry → 409 instead of the created row; duplicate membership/skill retries
hit unrecoverable 23P01 (no delete/close endpoints). | Missed: no
retry/replay tests. | Repro: publish a roster twice via REST → 422, not
replay. | Fix: enforce Idempotency-Key on the five endpoints with stored
results; add delete/close endpoints or document retry-by-reconciliation.
| blocks-M1: no.

**SYN-P2-4 [F-05+r5-F4+r5-F5] | P2 | Import autocommit (`:75-134`),
Correct post-commit upserts (`review.go:239-249`)** | Invariant: atomic
mutations; replay repairs; aggregates reflect reviewed truth. | Scenario:
(a) Import stores row-by-row, then pairs only `touched` employees
(populated solely on new stores) — a crash/mid-loop error after storing
but before pairing leaves events permanently unpaired, and replay reports
all-duplicates with empty `touched`, so pairing NEVER runs for them until
brand-new events arrive; (b) Correct commits correction+interval in tx,
then upserts day records OUTSIDE it — crash/interleaving leaves stale
minutes with status reviewed, which Generate reads. | Missed: happy-path
sequential tests only; no failure-injection or replay-after-partial test.
| Repro: insert events via SQL without pairing, Import identical bytes →
Duplicates=N, IntervalsBuilt=0, still unpaired. | Fix: tx-wrap
store+pair+mark (per import or per employee); seed `touched` from unpaired
owners on replay; move upserts inside the correction tx. | blocks-M1: no.

**SYN-P2-5 [r5-F6] | P2 | `internal/timesheet/timesheet.go:100-121`
Generate loop, `:193-300` checkEligible/generateOne** | Invariant:
timesheets generate only over fully reviewed input; no partial runs.
| Scenario: (a) TOCTOU — checkEligible runs OUTSIDE generateOne's tx; a
concurrent Import between check and commit adds unpaired events the
committed REG line never re-checks; (b) per-employee txs, no outer tx —
error at employee N leaves 1..N-1 committed; (c) operator validated but
never persisted. | Missed: single-threaded tests; skip test imports the
bad event BEFORE Generate, never concurrently. | Repro: code-inspected
(no unpaired recheck inside the tx); live race not executed. | Fix:
re-run eligibility inside the tx with locking, or lock the period; document
or remove partial-generation semantics. | blocks-M1: no.

**SYN-P2-6 [r5-F8] | P2 | `internal/wfm/roster.go:260-319` Assign**
| Invariant: assignment requires effective membership + skills. | Scenario:
membership/skill checks are separate SELECTs + autocommit INSERT with no tx
or lock — concurrent membership-end/skill-revoke/team-move between check
and insert persists an assignment for an unqualified employee, with NO DB
constraint backstopping either rule. Cross-employer inputs correctly
rejected (qualification gap, not isolation leak). | Missed: sequential
allow/deny tests only; no concurrency test. | Repro (static): pause Assign
after checks, end membership, resume → assignment persists. | Fix: tx +
SELECT … FOR UPDATE on membership/skill rows (or trigger re-check).
| blocks-M1: no.

**SYN-P2-7 [G5+F13] | P2 | `internal/timesheet/timesheet.go:364-403`
Approve/transition** | Invariant: approval transactional and idempotent.
| Scenario: Get-then-UPDATE in separate statements — two concurrent
Approves on a submitted sheet both read submitted; winner commits, loser
gets 422 invalid_state instead of idempotent success (same Approve-vs-Reject).
Underlying UPDATE atomic: no double-approve, benign spurious error.
Normalized to P2 (F13's P3 understates the contract promise; G5's P2 fits
the operational impact). | Missed: sequential re-approve test only.
| Repro (not executed live): parallel Approves → one 422. | Fix:
single-statement transition (`WHERE status IN (…) RETURNING`) with
compare-and-recover; enforce Idempotency-Key. | blocks-M1: no.

**SYN-P2-8 [H1+F7] | P2 | `internal/payroll/readiness.go:62-72,182-207`
Assess/loadExclusions** | Invariant: exclusions belong to the assessed run
and period; foreign rows never silently applied/dropped. | Scenario:
runID never validated — nonexistent UUID silently yields zero exclusions
(operator typo drops intended exclusions without error); a same-employer
run from ANOTHER period (or another employer) silently applies/drops rows;
e.g. `GET …/payroll-readiness?pay_period=PP-A&run=<run-of-PP-B>` applies
PP-B exclusions to PP-A. Run status unchecked. | Missed: happy-path
same-employer/same-period test only. | Repro: Assess(employer, PP-A,
&runB) → Excluded non-empty though run belongs to PP-B. | Fix: resolve
run by (employer,id), require same period, 404/mismatch otherwise; surface
run status. | blocks-M1: no.

**SYN-P2-9 [H2] | P2 | `internal/payroll/readiness.go:211-222`
checkEmployee** | Invariant: missing/ambiguous inputs produce explicit
blockers. | Scenario: checkEmployee requires only EXISTS(approved sheet) —
a zero-line / zero-hour / all-superseded (via SYN-P0-1) approved sheet
still counts as complete; readiness_test itself inserts bare approved
sheets and treats them as complete, encoding the weak check. | Missed: no
line/hours assertion. | Repro: approved sheet without lines + terms + SHG
→ Included, no MissingInfo. | Fix: require ≥1 line with hours>0 over
current intervals. | blocks-M1: no.

**SYN-P2-10 [D6+F12] | P2 | `internal/wfm/wfm.go:261-325` AssignSkills /
RequireSkills** | Invariant: multi-row setup writes atomic. | Scenario:
per-grant autocommit loop — `[good, bogus, good]` commits grant 1, fails
grant 2; retry of the fixed list → grant 1 now 23P01-conflicts with the
committed row, with NO delete API to recover (SQL-only wedge).
Normalized P1/P3→P2: half-applied setup with operator-visible error and
SQL recovery, not a broken core transition. | Missed: single-grant tests
only. | Repro: AssignSkills with 2 valid + 1 bogus code → 404 after first
committed; retry → 409. | Fix: validate all grants first, insert in one tx.
| blocks-M1: no.

**SYN-P2-11 [D7] | P2 | `internal/wfm/rotation.go:49-82` CreateRotation**
| Invariant: rotation patterns complete, unambiguous day-cycles. | Scenario:
duplicate offsets, sparse cycles (cycle_days 7, 1 row), Template+IsOff
contradictions (enshrined by rotation_test.go:23) accepted; template codes
free JSONB strings with no FK (dangle after rename); assign echoes caller
input as codes (UUID callers get UUIDs). Rotations inert in M1 (no
consumers) — advisory-only impact today. | Missed: validation tests cover
empty-rows/range/missing-target only. | Repro: cycle_days 2 with two
offset-0 rows → 201. | Fix: exactly-once 0..cycle_days-1 offsets, reject
contradictions, snapshot template IDs, echo canonical codes. | blocks-M1: no.

**SYN-P2-12 [D8] | P2 | `internal/wfm/roster.go:376-396` Publish**
| Invariant: double/concurrent publication safe and clearly reported.
| Scenario (limited): crash-safety GOOD (single conditional UPDATE
serializes publishers: one 200, others invalid_state). Gaps: retry → 422
instead of idempotent success; error conflates not-found vs
already-published. | Missed: no double/concurrent publish test. | Repro:
Publish twice → 200 then 422. | Fix: idempotent already_published result
or distinct code + tests. | blocks-M1: no.

**SYN-P2-13 [D9] | P2 | `openapi.yaml:463` vs `internal/wfm/roster.go:382-384`**
| Invariant: operational publish never described as validated/legal; spec
matches behavior. | Scenario: spec summary "Publish a roster (validated,
new immutable revision)" claims BOTH validation (readable as legal
sign-off while SG-MOM-SHIFT is BLOCKED — implementation sets
legal_validation_status=unavailable) AND a new revision (Publish creates
none — see SYN-P1-11). | Missed: contract tests never check
summary/description truth. | Repro: read spec vs Publish body. | Fix:
rewrite summary (operational checks only; legal unavailable; no new
revision in M1) + spec-words test. | blocks-M1: no.

**SYN-P2-14 [D10+C4+E-13a] | P2 | date/time `::$casts` in
`hr.go:208-213`, `wfm.go:220`, `roster.go:76-79,179,266,313`,
`review.go:62-68`** | Invariant: malformed input → 422 with machine codes,
never 500. | Scenario: invalid dates/times (`2026-02-30`, `not-a-date`,
`25:00`), inverted terms ranges, garbage review params surface PG
22007/22008/22P02/daterange errors as CodeInternal 500 via FromDB's
unmapped default — breaking the machine-readable contract. Only
rotation.go:133 validates in Go. | Missed: no malformed date/time tests
anywhere. | Repro: POST roster week `2026-02-30` or template `25:00` → 500.
| Fix: parse/validate every date/time in Go (`time.Parse`), reject
inverted ranges with CodeValidation. | blocks-M1: no.

**SYN-P2-15 [C6] | P2 | `internal/hr/hr.go:261-279` AddCompensation (zero
CLI/REST callers, zero tests)** | Invariant: every write path operable and
validated. | Scenario: compensation lines writable only by direct Go calls;
no CLI/REST add/list/remove; range-within-terms unchecked; float64 amount,
negatives only via DB CHECK. | Missed: literally zero references in tests.
| Repro: grep AddCompensation in cmd/httpapi → nothing. | Fix: expose via
CLI+REST with validation, or explicitly descope with a decision record.
| blocks-M1: no.

**SYN-P2-16 [r5-F15+I-07+E-10] | P2 | edge defaults (`httpapi.go:556-561`,
`cli.go:40-46`) × dropped operators (Import/Generate/Submit/Create/SetStatus)
× zero `audit_event`/`idempotency_key` writers** | Invariant: operator
attribution required and auditable. | Scenario: (a) API/CLI default empty
operator to $USER→'api'/'cli', so service "operator required" errors are
unreachable and unattributed writes silently record defaults; (b) Import
and Generate validate then persist nothing; Submit stamps time only;
Create/SetStatus take no operator; (c) `audit_event` and `idempotency_key`
have NO writers — no audit trail exists at all. Values stored unsanitized.
Normalized P3/P2→P2. | Missed: tests always pass explicit operators, never
assert persistence or omission behavior. | Repro: publish with no
X-Operator → published_by='api'; import then SELECT audit_event → empty.
| Fix: decide reject-vs-default contract; persist actor on every mutation
(or write audit_event rows); use or drop idempotency_key. | blocks-M1: no.

**SYN-P2-17 [E-06] | P2 | `internal/attendance/attendance.go:140-155`
markExceptions + `review.go:60-133`** | Invariant: pairing failures
explicit and auditable. | Scenario: exceptions exist ONLY in the transient
Import Report — bare status=exception flips on days that happen to have
records (record-less days get nothing); ReviewAttendance returns no codes,
so an operator who missed stdout cannot discover WHY a day is flagged;
unknown-employee rows likewise report-only (see SYN-P2-20). Downgraded
P1→P2: discoverability/operability gap; the blocking correction path is
SYN-P1-12. | Missed: tests assert in-memory report only. | Repro: import
dangling-break CSV, discard stdout, call review → no codes anywhere.
| Fix: persist exceptions keyed to employee/date/import; surface in review.
| blocks-M1: no.

**SYN-P2-18 [E-07] | P2 | `attendance.go:231-238` ORDER BY occurred_at,id +
`:118-132` map iteration** | Invariant: deterministic pairing; same bytes
→ same report. | Scenario: uuid v4 ids randomize equal-timestamp event
order across identical replays (in/out ties → different pairing); Go map
iteration randomizes multi-employee exception ORDER in reports. | Missed:
single-employee, distinct-timestamp fixtures only. | Repro: import
same-timestamp in/out twice into fresh DBs → differing outcomes. | Fix:
ORDER BY occurred_at, kind, source, source_event_id; sort touched list.
| blocks-M1: no.

**SYN-P2-19 [E-08] | P2 | `attendance.go:118-132` touched-only pairing vs
contract §7 replay promise** | Invariant: replay determinism. | Scenario:
pairing+exceptions run only for newly-STORED employees — replaying an
all-duplicate file returns Exceptions=[] even when the original reported
missing_punch/dangling/missing_break; operators re-running import to
retrieve exceptions get silence. | Missed: replay test asserts dup counts
only. | Repro: import lone-`in` (1 missing_punch), replay identical CSV,
compare Reports. | Fix: recompute (or persist + re-attach) current
exception state on every import. | blocks-M1: no.

**SYN-P2-20 [E-09] | P2 | `attendance.go:86-100` unknown-employee branch**
| Invariant: preserve raw clock evidence; never silently discard unmatched
events. | Scenario: unknown-employee rows produce a report-only exception
and are NEVER stored — no clock_event, no quarantine. If stdout is lost,
evidence is unrecoverable until the vendor resends (hence P2, not P0).
| Missed: no unknown-employee import test at all. | Repro: import
employee_no=NOSUCH → report exception, zero clock_event rows. | Fix:
quarantine table (raw row + reason) or store unlinked; surface in review.
| blocks-M1: no.

**SYN-P2-21 [E-11] | P2 | `attendance.go:347-360` nil-assignment path
(`ExcNoRosterLink` dead)** | Invariant: unrostered work explicit, not
silently reviewed. | Scenario: spans with no matching current-revision
assignment proceed with assignmentID=nil and NO exception; day reviewed,
flows to timesheets. Only breaks notice (via ambiguous classification);
pure in/out spans pass silently. | Missed: fixture always assigns a
roster; no no-roster test. | Repro: clean in/out with no assignment →
zero exceptions + generated sheet. | Fix: emit no_roster_link + flag day,
or explicitly document unrostered-work acceptance. | blocks-M1: no.

**SYN-P2-22 [E-12] | P2 | `review.go:205-235` cross-day correction**
| Invariant: corrections maintain correct relationships. | Scenario: when
corrected start moves the interval to a new date, the new row keeps the OLD
date's roster_assignment_id (date-specific; FK checks employee only), and
the correction row links only the old day's record — new-day mix + missing
audit link. | Missed: same-day correction test only. | Repro: correct a
start into the next day; compare assignment ids. | Fix: re-resolve
assignment for the new date; link both day records. | blocks-M1: no.

**SYN-P2-23 [G4] | P2 | `00002_attendance_payroll.sql:134-141`
timesheet_line_interval (no employee-match enforcement)** | Invariant:
cross-employee links rejected; critical invariants DB-enforced. | Scenario:
allocation FKs are (employer,line)/(employer,interval) only — direct SQL
can link empA's line to empB's interval (same employer), reading as empA
payable time. Latent: no API/CLI writer besides correct Generate; needs DB
access. | Missed: lineage tests cover cross-employer only, not
cross-employee. | Repro: SQL INSERT joining empA line to empB interval →
succeeds. | Fix: enforce employee match (denormalized id + composite FKs,
or trigger) + negative test. | blocks-M1: no.

**SYN-P2-24 [G6] | P2 | no Edit/Update service, CLI verb, or REST endpoint
for drafts** | Invariant: draft timesheets editable (domain-model "freely
editable in draft"). | Scenario: operators cannot fix a draft except via
SQL; combined with SYN-P1-14, post-submit imperfections force
reject-into-dead-end. | Missed: freeze matrix proves SQL-level draft
mutability only. | Repro: `--help` + openapi show no edit op. | Fix: draft
line-edit endpoint/verb, or explicitly scope out with a documented SQL-free
workaround. | blocks-M1: no.

**SYN-P2-25 [I-03] | P2 | `openapi.yaml:10-12` + contract `:84-87` vs
`internal/employer/employer.go:43-81`** | Invariant: spec promises hold.
| Scenario: spec/contract promise 403 employer_context_rejected — no such
code/status exists; Resolve returns 404/422 only; no configured-context
concept in v0.1. Dead spec consumers may code against. | Missed: no 403
test. | Repro: grep finds only docs/spec refs, zero code. | Fix: drop 403
from spec/contract for v0.1 or implement the check. | blocks-M1: no.

**SYN-P2-26 [I-08+S7-07] | P2 | `internal/httpapi/contract_test.go:55-98`,
`cmd/workforceops/parity_test.go:81-205`** | Invariant: every implemented
path has a contract test; parity covers behavior. | Scenario: contract
tests check method+path presence only (no validation/codes/schemas);
parity covers 4 happy-path legs (employee/status/programmes/roster) of 38
ops — corrections/publish edges/readiness/activations untested at adapter
level; `restDo` fatals on ≥400 so error-parity is inexpressible; harness
mounts bare echo without RequestID/Recover middleware, so prod divergence
untestable. | Missed: structural by design. | Repro: `go test
./internal/httpapi/` passes without touching handler logic. | Fix:
table-driven route tests (happy+422+404+envelope) on disposable DB;
failing-parity legs; build test server via core/api New(false); extend
legs to import/review/generate/readiness. | blocks-M1: no.

**SYN-P2-27 [I-09] | P2 | CLI-only require-skills/rotation-pattern/skill-list
vs openapi (no such paths)** | Invariant: API workflow self-sufficient.
| Scenario: POST /rotations needs a CLI-created pattern; POST /skills needs
CLI-discovered codes; Assign validates required-skills settable ONLY via
CLI (unacknowledged) — API-only consumers cannot complete the roster flow
from scratch. | Missed: parity/E2E mix CLI+REST legs, never API-only.
| Repro: fresh employer via REST alone → no publishable roster reachable.
| Fix: REST equivalents, or normatively document API+CLI hybrid. | blocks-M1: no.

**SYN-P2-28 [L-02] | P2 | `httpapi.go:365-374` FormFile/Bind ×
`core/api/server.go:23-47` (no BodyLimit) × `attendance.go:157-214`
(full-file accumulate)** | Invariant: bounded untrusted inputs. | Scenario:
multipart CSV + JSON bodies unbounded — no MaxBytesReader/BodyLimit/row
cap; huge upload OOMs the loopback server. Impact capped at loopback
self-DoS today. | Missed: no limit tests. | Repro: POST multi-GB file to
/clock-events. | Fix: echo BodyLimit + MaxBytesReader + CSV caps (MUST
before any exposure). | blocks-M1: no.

**SYN-P2-29 [S7-04] | P2 | `wfm_test.go:103-111,132-133`,
`timesheet_test.go:66-76`, `attendance_test.go:72-73`** | Invariant: tests
assert persisted business state, not just nil error. | Scenario:
ListClients/Programmes/Teams results discarded (`if _, err :=`) — an
employer-scoping bug returning [] passes; AddMembership returns ignored;
`c, _ :=` chains panic opaquely or silently change what Assign validates.
| Missed: no negative control (second employer, empty-list rejection).
| Repro: stub ListClients to empty+nil → TestBPOChain still green. | Fix:
assert codes/counts/scoping; check every setup error. | blocks-M1: no.

**SYN-P2-30 [S7-05] | P2 | `internal/payroll/approval.go:108-133` +
`approval_test.go:94-100` vs `readiness.go:143-145`** | Invariant: approval
predicate and readiness verdict agree on empty populations. | Scenario:
AuthorizeApproval with zero Included returns nil (vacuous approve), test
encodes wantErr:false, while Assess verdicts "not ready: no included
employees" — an empty run passes the gating predicate unopposed. | Missed:
table case asserts implementation instead of the all-or-nothing contract.
| Repro (observed): AuthorizeApproval zero-Included → nil. | Fix: decide
contract (reject empty vs allow), align Assess+predicate+test + negative
test. | blocks-M1: no.

**SYN-P2-31 [S7-06] | P2 | Import/Generate/Publish (zero racing coverage)**
| Invariant: idempotent/dedup paths proven under concurrency. | Scenario:
no test runs concurrent same-CSV Import, concurrent Generate (the
CodeConflict→skip branch `:111-116` has zero racing coverage), or double
Publish; suite has no `-race` concurrent case anywhere. | Missed: replay
tests strictly sequential. | Repro: `go test -race` passes but exercises
no goroutines. | Fix: bounded concurrency tests (dual import/generate/
publish) asserting exactly-once effects. | blocks-M1: no.

**SYN-P2-32 [r1-F1] | P2 | `Dockerfile:12-14,24` + `fly.toml:15-18,23-24`**
| Invariant: deleted executables leave no broken references; report's
"straggler search clean" holds. | Scenario: `docker build` necessarily
fails compiling deleted `./cmd/shifthub-api`; CMD + fly `[processes]`
point at binaries that can never exist — any container/deploy attempt
fails. (`docker build` itself not executed — no daemon; inference from
observed absent dirs is deterministic.) | Missed: no test/lint touches
container files; straggler search evidently excluded them. | Repro:
`go build ./cmd/shifthub-api` → directory not found (observed); `ls cmd/`
= workforceops only. | Fix: rebuild Dockerfile/fly for the single binary
(or delete both until cloud deploy authorized). | blocks-M1: no.

**SYN-P2-33 [r5-F3] | P2 | `internal/seed/seed.go:32-131` Demo/resetEmployer**
| Invariant: atomic mutations; reset must work. | Scenario: (a) Demo runs ~8
sequential steps with no tx — mid-seed failure leaves half-seeded MERIDIAN;
(b) resetEmployer issues 39 autocommit DELETEs — failure at N leaves a
half-deleted employer; worse, several deletes CANNOT succeed (published
roster → P0001 on roster_assignment delete even as superuser; approved
artifacts hit freezes; app-role DELETEs hit 42501 revokes). Demo+reset can
wedge: partial seed, unresettable. Static analysis only — reset never
executed live. | Missed: zero seed-path coverage. | Repro: disposable DB,
seed (published by default) → reset → P0001 on roster delete. | Fix: tx-wrap
both; make reset trigger-aware (reopen-first or ordered deletes); add
Demo→reset→Demo idempotency test. | blocks-M1: no.

### 4e. Confirmed defects — P3 (low)

**SYN-P3-1 [r1-F3] | P3 | `internal/httpapi/httpapi.go:513-541` (GET
rule-activations) duplicating `internal/payroll/readiness.go:163-180`**
| Invariant: CLI/API invoke the same `internal/` services; no handler-side
DB (AGENTS.md #1). | Scenario: read-only today, but the handler owns a
private copy of the activation query (own row type `:524-528`) that will
silently diverge from any future readiness filter/mapping; no CLI
counterpart. | Missed: contract tests check route-subset structurally;
no arch test forbids pool.Query in httpapi. | Repro: identical SQL at
httpapi.go:518-519 vs readiness.go:164-165; no CLI `rule-activations`.
| Fix: export `payroll.ListActivations` wrapping loadActivations; call from
the handler (optionally add CLI). | blocks-M1: no.

**SYN-P3-2 [r5-F10] | P3 | `internal/employer/employer.go:16-60` Resolve**
| Invariant: deterministic single-employer default; no ambiguous
resolution. | Scenario: (a) empty input runs two non-atomic queries (SELECT
rows, then count) — concurrent employer creation between them flips the
outcome; first row arbitrary under 2+ employers (count check then rejects,
but only after reading it); (b) UUID-shaped employer codes unresolvable
(id-only branch, never code). | Missed: no Resolve unit test at all.
| Repro: Resolve("") concurrently with creating employer C →
nondeterministic outcome. | Fix: single query (`count(*) OVER ()`) or tx
with ORDER BY + LIMIT 2; fall back to code lookup on UUID-id miss.
| blocks-M1: no.

**SYN-P3-3 [r5-F11] | P3 | freeze triggers (`00006:29,96,112,118,123`,
status lookup by global id)** | Invariant: no cross-employer information
disclosure via error paths. | Scenario: BEFORE triggers fire before FK
validation and look up parent status WITHOUT employer predicate, so
`INSERT timesheet_line(employer A, sheet=<B's approved>)` returns
P0001/frozen (naming the sheet) instead of the FK rejection — a
frozen-vs-validation oracle revealing existence+status of a UUID the caller
must already know. Minimal content, genuine side channel. | Missed:
freeze + cross-employer tests run separately, never combined. | Repro: as
above → observe `frozen` instead of cross-employer validation error.
| Fix: scope trigger lookups by employer (`WHERE employer_id=NEW…`).
| blocks-M1: no.

**SYN-P3-4 [r5-F14] | P3 | `00006:145-149`, `00004:273-316`, `00008:18-23`
Down bodies** | Invariant: rollback restores prior behavior without skew
or privilege widening. | Scenario: (a) 00006 Down is `SELECT 1` no-op —
downgrade leaves v6 bodies live while goose records v3 (status lies);
(b) 00004 Down re-GRANTs UPDATE/DELETE on all seven insert-only tables —
rollback permanently widens privileges; (c) 00008 Down fails when
break rows exist (documented in-file, no data migration). | Missed: no
test runs Down at all. | Repro: up to 8, insert break event, down to 3 →
break failure, then grant-widening if forced. | Fix: real 00006 Down (or
document skew); preserve revokes in 00004 Down; add down/up round-trip
test on scratch DB. | blocks-M1: no.

**SYN-P3-5 [G7] | P3 | `internal/attendance/attendance.go:372-382`
storeSpan refs** | Invariant: every line traces to source intervals
precisely. | Scenario: one event-ID list (in+out+ALL breaks) stamped on
EVERY work run of the span — the 09:00–12:00 interval cites 12:00–13:00
break events it doesn't contain; break events cited by both work and break
intervals. Provenance audit can't attribute hours to events (many-to-many);
unpaired-detection unaffected. | Missed: no source_refs content assertion.
| Repro: import in/break/out, compare the two work runs' refs → identical
supersets. | Fix: cite per-interval boundary events only. | blocks-M1: no.

**SYN-P3-6 [G8] | P3 | `timesheet.go:194-258` (no source filter) ×
`00002:19` lineage CHECK** | Invariant: timesheets only from eligible
reviewed attendance. | Scenario: latent — generation filters by kind, never
by source, so a `source='assumption'` interval would be silently summed.
No M1 path creates assumption intervals (only direct SQL or a future
slice could). | Missed: no negative test. | Repro: SQL-insert
kind=work/source=assumption → included. | Fix: filter `source IN
('clock','correction')` + negative test. | blocks-M1: no.

**SYN-P3-7 [I-10] | P3 | `core/api/server.go:48-63` (no custom 404/405
handler)** | Invariant: single machine-readable envelope. | Scenario:
unknown route/method → Echo default `{message:…}` in dev, or
status-correct but body-misleading `{error:Internal server error}` for
404/405 in prod. Inferred from defaults + handler; serve never
live-probed. | Missed: no 404/405 body tests. | Repro: POST /healthz or
GET /nope, inspect body. | Fix: map echo.HTTPError codes to the envelope
with accurate messages. | blocks-M1: no.

**SYN-P3-8 [J-03] | P3 | `cmd/workforceops/migrate.go:32`, `main.go:42-49`,
`cli.go:84-92`** | Invariant: `--format json` honored; single
machine-readable error channel. | Scenario: `migrate up`/`version` print
plain text regardless of --format; failures print JSON envelope to stderr
AND cobra Error+Usage text (verified double output, exit 1 on the xml
case). | Missed: no CLI output-format tests. | Repro: `--format json
migrate status`; `--format xml employee list`. | Fix: honor format in
migrate/version; SilenceUsage/SilenceErrors with envelope-only stderr.
| blocks-M1: no.

**SYN-P3-9 [J-04] | P3 | `cmd/workforceops/cli.go:114-127` readJSONFile**
| Invariant: bounded, strict file inputs. | Scenario: unbounded
os.ReadFile + Unmarshal without size cap or DisallowUnknownFields — GB
--file OOMs the CLI; typo keys silently ignored (services mostly re-catch
via required-field checks). | Missed: no large/malformed-file CLI tests.
| Repro: `employee create --file <2GB`. | Fix: byte cap +
DisallowUnknownFields for strict shapes. | blocks-M1: no.

**SYN-P3-10 [J-05] | P3 | `cmd/workforceops/wfm.go:98-99,192-193`,
`httpapi.go:266-269`** | Invariant: useful identifiers in success
responses. | Scenario: skill assign / require-skills return `{result:ok}` /
`{employee}` with no assignment IDs (unlike team add-member `{id}` or
corrections `{interval_id}`) — operators can't reference created rows.
| Missed: no success-shape assertions. | Repro: `skill assign --file …` →
result ok only. | Fix: return created assignment IDs. | blocks-M1: no.

**SYN-P3-11 [S7-08] | P3 | `tests/dbutil/dbutil.go:75-95` DSN guard**
| Invariant: destructive-capable setup refuses non-test databases.
| Scenario: guard rejects only empty/`postgres` names — pointing
WORKFORCEOPS_TEST_DATABASE_URL at the real dev DB passes, and suites then
write employers/employees (integrity TestMain would CREATE ROLEs on that
server). | Missed: only integrity asserts who@where. | Repro: inspect
DSN():86-88; any non-postgres name writes freely. | Fix: require a test
marker (e.g. `_test` suffix) or CONFIRM guard + log who@where. | blocks-M1: no.

**SYN-P3-12 [S7-09+F4] | P3 | `Taskfile.yml:91-131` + `.env.example:25-26`
× rootCmd** | Invariant: documented entrypoints exist; no stale command
surface. | Scenario: `task demo` / oracle / scenario / investigate /
evaluate invoke subcommands that don't exist → loud `unknown command`
(observed for `demo` and `oracle --help`); stale A1_SHIFTHUB_PORT/
A1_PAYLOOP_PORT dead while fly.toml still exports them. Normalized P2→P3:
fails loudly, no silent wrong behavior; cosmetic dead surface. | Missed:
no CLI-surface-vs-Taskfile/docs test; acknowledged-but-untested. | Repro:
`./build/workforceops demo` → unknown command (exit 1); `seed demo` works.
| Fix: register `demo` or retarget the task; delete/stub-guard future
tasks; drop the stale port vars. | blocks-M1: no.

**SYN-P3-13 [E-13b/c] | P3 | `attendance.go:157-214` parseCSV; tz use**
| Invariant: robust ingestion; documented tz policy. | Scenario (minor,
grouped): CSV kind case-sensitive (`IN` rejected) while headers are
case-insensitive; no BOM tolerance (BOM breaks the strict header match,
failing the whole file); mixed-tz spans allowed (in-tz wins, undocumented
vs Singapore-business-date language); dangling-break exception date can
mismatch span localDate and no-op. | Missed: no negative CSV tests.
| Repro: import BOM-prefixed CSV → whole-file failure. | Fix: trim BOM,
document tz policy. | blocks-M1: no.

### 4f. Potential risks (plausible, not confirmed)

- **RISK-1 — Race windows unmeasured.** All concurrency findings
  (SYN-P1-20, SYN-P2-5/6/7) are statically reasoned; none executed against
  live PG. Exploitability under realistic single-operator loopback load may
  be low, but M2 multi-user use raises it. Treat as must-test-during-M2,
  not must-fix-now.
- **RISK-2 — Production role/topology unknown.** No manifest beyond fly.toml
  shows which DSN role serve/seed use in prod; SYN-P1-7 analyzes both
  branches (superuser → protections void; app → compose-grant path unproven),
  but the actual prod value and server `timezone` setting (SYN-P1-5
  exploitability) were not inspected.
- **RISK-3 — Seed-reset wedge unexecuted.** SYN-P2-33's trigger-conflict
  analysis is static; `seed demo --reset` was never run live (only `seed
  demo`). Demo-day reset could fail louder than analyzed, or succeed on
  paths assumed broken.
- **RISK-4 — `docker build` unexecuted.** SYN-P2-32's failure inference is
  deterministic (absent dirs) but the exact docker error string was never
  captured; no container exists to smoke-test.
- **RISK-5 — modules/ registry reachability.** `modules/sg-wfm-payroll`
  still references shifthub/payloop (module.yaml:28-29, module.go:41,44);
  whether the module registry is reachable from the binary was only lightly
  traced — likely dead code, unconfirmed.

### 4g. Untested hypotheses (critic gaps — prompt8 items with zero coverage)

Neither confirmed nor refuted; each needs a first inspection pass in M2:

1. **Cross-midnight split** (E): storeSpan stamps one localDate from the
   in-event (`attendance.go:346`); zero midnight/split references in
   `internal/attendance` vs domain-model.md:103-105 — shifts crossing local
   midnight may mis-attribute whole spans to one date.
2. **Late-break imports** (E): breaks arriving in a later import hit the
   unmatched-break path once in/out already paired — attachment outcome
   never examined.
3. **Multiple shifts/day + duplicate break events** (E): prompt8-named;
   never inspected or tested.
4. **Overlapping pay periods** (G): `pay_period` has no date-range exclusion
   (`00002:87-100`, code-unique only) while Generate filters by local_date —
   double-linking across overlapping periods unexamined.
5. **`timesheet_line_interval` lacks UNIQUE(employer_id, interval_id)** (G/K):
   one interval linkable into two lines across periods/employees untested.
6. **Revision publication metadata** (D): Publish writes `roster.published_*`
   only, never `roster_revision.published_*` (`00001:321-322`).
7. **Rotation assignment ownership conflicts** (D7-adjacent): cross-team/
   employer conflicting-ownership validation never explicitly checked.
8. **HTTP status-code matrix** (I): `fail()` maps `ae.Status` but per-error
   statuses (404 vs 422 vs 500) verified on no route — only envelope keys.
9. **Maturity-label alignment** (I): served `/rule-activations` vs draft
   `/rule-packs` contradiction never adjudicated.
10. **Per-route validation matrix** (I/J): malformed-input→500 shown for
    dates/enums only; incorrect-success-on-service-failure never checked.
11. **CLI exit-code table** (J): exactly one observation exists (SYN-P3-8 xml
    case); per-command codes unverified.
12. **Upgrade from baseline `f435158`** (K): all installs tested fresh;
    migration over the legacy shifthub/payloop schema unrun.
13. **Rollback execution** (K): Down bodies analyzed statically only; no
    down/up round-trip run.
14. **Resource hygiene** (K): rows/connection/file-handle discipline never
    systematically checked (spot pattern `defer rows.Close()` seen).
15. **Six-scenario B battery live** (B): cross-employer checks rest on
    existing integrity tests + inspection; the prompt8 B attempt list never
    run as an overlapping-identifier two-employer battery.
16. **E2E/report-claim audit**: the 12-step workflow and numeric claims
    (6/6 MATCH, 39.5h provenance, blocked counts) never re-executed
    end-to-end; SYN-P1-12 notes E2E step 8 itself used SQL.
17. **Rule-source bytes** (Phys): SG_research.pdf + rule-source-register.md
    not re-inspected (bounds the S5 givens below).
18. **Indeterminate-path enumeration** (Phys §5): blocked/unsupported routing
    asserted but Go indeterminate-code paths never listed.
19. **Config/DSN leak** (L): whether config or DSN error paths echo
    connection strings/secrets never examined.
20. **Compensation attribution semantics** (C, minor): beyond SYN-P2-15's
    unreachability, row actor/effective content never reviewed.

## 5. CLEAN areas (evidence examined, no defect found)

- **Single executable.** Sole `func main` (`cmd/workforceops/main.go:15`,
  repo-wide search); `ls cmd/` = workforceops only; Taskfile builds only
  `build/workforceops`; deleted dirs verified absent (observed build
  failures); `GOWORK=off go build ./cmd/workforceops` ok;
  `migrate_test.go:49` shifthub/payloop mentions are legacy-absence
  assertions. (Broken external refs to deleted names are SYN-P2-32 /
  SYN-P3-12, not extra executables.)
- **Shared CLI/REST service layer.** All CLI RunE bodies thin (resolveCLI +
  one `internal/` call); all 38 REST handlers resolve employer then call
  the same functions; no handler/CLI math or second write path (sole
  exception: read-only SYN-P3-1). Cross-adapter proof:
  `TestCLIAPIParity` 4 legs assert business values across CLI↔REST on one PG.
- **Employer scoping (read/write).** Every owned-service query predicates
  `employer_id` (audited across hr/wfm/roster/rotation/attendance/review/
  timesheet/readiness); composite `(employer_id,id)` UNIQUEs + composite
  FKs verified in 00001/00002; per-employer codes, global employer code,
  per-employer clock-event namespace verified; read-path JOINs resolve via
  globally-unique PKs under composite FKs. Global unscoped tables are
  jurisdiction-catalogue by design. No RLS, consistent with AGENTS.md #8.
- **No service bypass of constraints.** Mutations trigger-compatible
  (guarded transitions, draft-only roster writes, supersede-not-UPDATE);
  no UPDATE/DELETE of insert-only tables outside dev-only seed reset; no
  `session_replication_role`/DISABLE TRIGGER/superuser escape in
  internal+cmd. Freeze triggers fire for every role, so SYN-P1-7 voids the
  REVOKE layer only, not trigger immutability.
- **SQL injection.** All runtime SQL parameterized; sole dynamic fragment
  (`timesheet.go:384-395`) chooses between two constant SET clauses with
  bound args; refsJSON passed as bound `$9::jsonb`; no Sprintf-built SQL
  repo-wide; app cannot CREATE in public (shadow-table injection closed).
- **CSV/sink hygiene.** Parser strictly validated (exact header, kind enum,
  RFC3339, IANA tz, required fields, per-row errors); values stored
  verbatim, never evaluated; no CSV export exists (no formula sink).
- **Loopback binding.** Bind 127.0.0.1 hardcoded in serve; no --bind flag
  or bind env (contract §1-2 documents a nonexistent --bind flag — doc-only
  P3: remove or implement loopback-validated); access logs carry
  method/URI/status only; operator attribution-only everywhere. Caveat:
  `core/api.Run` trusts caller Bind — harden with a loopback guard if reused.
- **Employer boundary (adapters).** Every v1 route resolves employer from
  path; CLI resolves flag-or-single-default; all lookups parameterized;
  empty input resolves only when exactly one employer exists (fail-closed
  count check; race direction fails safe per audit, modulo SYN-P3-2).
- **OpenAPI counts.** 36 paths / 46 ops / 38 implemented / 8 draft / 46
  unique operationIds independently verified; served == implemented (36
  routes + /healthz + /readyz); no undocumented handlers; 8 drafts
  correctly unserved; no placeholders; no CORS middleware (same-origin
  default, safe for loopback — not browser-tested).
- **Dedup/replay core.** Dedup key correctly isolates across
  employers/sources; replay never duplicates events (tested); raw
  clock_event never updated/deleted by services; out-of-order rows sorted;
  cross-delivery pairing works (tested); duplicate/missing/dangling/break
  exceptions explicit and block timesheets via the unpaired check.
- **No invented time.** No scheduled-as-observed inference (missing break
  evidence → exception; `assumption` source never created by services); no
  `min(H,72)` payable cap anywhere (grep-verified); monthly excess
  flag-only; corrections preserve rows via supersedes chain with
  reason+operator; sequential repeat-correct rejected; exactly-one-source
  lineage CHECK enforced + tested; regeneration skips existing sheets.
- **Migration mechanics.** 00005 id defaults, 00007 trigger renames (spec
  names verified by test), 00008 break-kind widening with honest Down
  caveat; extension prerequisites fail loudly; MigrateUp idempotent;
  ≥8 versions recorded; live fresh `migrate up` + `seed demo` verified.
- **Integrity suite enforcement.** Dedicated DB, role bootstrap, app
  grant/revoke mirror, who@where assertion; two-employer isolation,
  insert-only revokes, rerun-safety enforced with exact SQLSTATE assertions;
  passed live on disposable DB (serial).
- **No-mock test honesty.** Every DB test uses real PostgreSQL; expected
  values are independent literals (540/480, 9h, counts/reasons), never
  recomputed via implementation helpers; `TestAuthorizeApproval` pure and
  mock-free.
- **No legal inference from job titles.** Classification is recorded data
  (explicit category whitelist + part_iv_eligible + mandatory basis string),
  never derived; readiness checks terms existence only.
- **39.5h adjudication (H3).** The 39.5h-approved-yet-ready:false outcome is
  the correct blocking-matrix consequence (E2E-001 lacks the seed SHG row;
  DEMO-001/003 never submitted/approved; DEMO-002 unpaired-skipped), not
  conflation: Population/Included/Excluded/Blocked/MissingInfo are distinct
  fields. Minor coarseness: the Included==0 verdict string alone doesn't
  separate empty-period vs all-missing vs all-excluded.
- **Phys-Research/Phys-Pseudo boundaries (S5).** SG-MOM-OT-001 BLOCKED with
  load-bearing open ?s; zero overtime/CPF/SDL/SHG/net-pay computation in Go
  (OT15/OT20 map only to pack requirements); zero golden fixtures (dirs
  empty); 72h flag-only; provisionals labeled provisional, never pinned
  legislation; unsupported cases route to blocked/unsupported paths; 1.5x
  documented as the Meridian contractual assumption; removed `min(H,72)`
  cap-branch absent from SPEC v0.3.0 and Go; `rules/` holds no `.go` (spec
  text, not executed).

## 6. Known-limitation assessment (report §6, 7 items)

| # | Limitation | Triage | Judgment |
|---|---|---|---|
| LIM-1 | Bare arrays, no pagination | **Must-fix-before-next** (SYN-P2-1) | Acknowledged accurately re shape; understates cost growth — every new consumer hardcodes the wrong shape. M1 demo unaffected (consistent adapters). |
| LIM-2 | Idempotency-Key not enforced | **Must-fix-before-next** (SYN-P2-3) | Acknowledged; understates publish-retry hazard (422 vs replay indistinguishable). Clock import safe via source IDs; approve safe via status check. |
| LIM-3 | Interval IDs only via SQL | **Must-fix-before-next, blocks M1** (SYN-P1-12) | Acknowledged but severity understated: the advertised correction endpoint is *unusable* via API — any attendance exception forces psql. |
| LIM-4 | Rotation patterns CLI-only | Future-enhancement for REST parity, **but CLI list half broken** (SYN-P1-22) is must-fix | Partially acknowledged (CLI-only yes, broken list no). M1 E2E doesn't use rotations. |
| LIM-5 | Skill catalogue CLI-only | Future-enhancement (SYN-P2-27) | Acknowledged accurately; minor friction, no hard block (codes knowable out-of-band). |
| LIM-6 | Operator is attribution not auth | Acceptable for M1 | Acknowledged honestly; loopback-only synthetic tool, no privilege derived. Keep SYN-P2-2(c) redaction + SYN-P2-28 limits as preconditions for wider exposure. Real auth future work. |
| LIM-7 | Legacy oracle/scenario Taskfile tasks | Acceptable (SYN-P3-12) | Acknowledged; tasks fail loudly, no silent wrong behavior. Cosmetic. |

Unacknowledged gaps found beyond the 7 (new in this review): SYN-P1-22
(shadowed list), SYN-P1-23 (SQL-only onboarding incl. employer/period/SHG),
SYN-P2-27's require-skills CLI-only half, SYN-P2-16 (validated-but-discarded
operator), SYN-P2-3's dead `idempotency_key` table, SYN-P3-7 (§5 loopback
--bind doc), SYN-P2-32/SYN-P3-12 container/Taskfile staleness.

## 7. Final assessment (prompt8 §8)

### A. Milestone 1 functional correctness — NO

The happy path (seed → import → review → generate → submit → approve →
readiness) demonstrably runs, and the service tests + live serial suite
pass. But the workflow does not *genuinely* work: approved sheets silently
detach from corrected evidence (SYN-P0-1), unbounded spans become clean
payable time (SYN-P0-2), duplicate feeds double-count (SYN-P1-2), rejected
sheets dead-end (SYN-P1-14), the second status change always conflicts
(SYN-P1-8), and publish blesses empty/unstale-checked rosters with no
reopen path (SYN-P1-10/11). Too many core transitions are broken or
unguarded for a correctness pass.

### B. Database integrity — PARTIAL (isolation yes; immutability no)

Employer isolation is genuinely enforced: employer_id predicates, composite
FKs/UNIQUEs, and the integrity suite's exact-SQLSTATE assertions all hold
under inspection and live serial execution — no cross-employer leak was
found (the SYN-P3-3 oracle is a status side channel, not a leak).
Immutability is weaker: trigger re-parenting bypasses freeze (SYN-P1-1),
role separation is docs-only with silent fresh-install degradation
(SYN-P1-7), and interval supersede escapes the freeze model entirely
(SYN-P0-1). Lineage is enforced but imprecise (SYN-P3-5, SYN-P2-23 latent).

### C. CLI/API usability — NO

A real operator cannot complete the promised workflow without SQL: the
correction endpoint is unusable via API (SYN-P1-12, acknowledged as LIM-3),
non-seed onboarding and monthly periods/SHG need manual SQL (SYN-P1-23),
published rosters are uncorrectable (SYN-P1-11), rejected sheets unrecoverable
(SYN-P1-14), rotation patterns unlistable (SYN-P1-22), and API-only consumers
cannot reach a publishable roster (SYN-P2-27). The E2E report's own step 8
used SQL for an interval id. Spec-shape divergence (SYN-P2-1/2) further
breaks spec-conformant clients.

### D. Payroll-readiness integrity — YES, with qualifications

The implementation correctly refuses to invent financial results: zero
overtime/CPF/SDL/SHG/net computation in Go, OT-001 BLOCKED with open ?s,
no golden fixtures, 72h flag-only, provisionals honestly labeled, and the
39.5h case is correct blocking-matrix behavior (H3). Qualifications: Assess
population silently drops retroactively-terminated approved staff
(SYN-P1-9), run-id scoping silently mis-applies exclusions (SYN-P2-8), and
empty/stale approved sheets count as complete (SYN-P2-9, SYN-P0-1) — the
gate refuses to compute, but its input accounting has holes M2 payroll
would inherit.

## 8. Overall recommendation: FAIL

Two P0s (approved-sheet staleness, unbounded payable spans) plus thirteen
blocking P1s (freeze bypass, double-count, evidence handling ×3, role
separation, status lifecycle ×2, publish ×2, interval listing, test race)
are material defects that must be corrected before the next milestone.
Per §8 this is FAIL, not pass-with-conditions: the conditions would be the
bulk of the milestone's correctness guarantees. The failure is bounded and
fixable — architecture, isolation, and payroll-boundary discipline are
sound; the defects concentrate in attendance edge semantics, lifecycle
transitions, and install/role plumbing. Clearing §9.A converts this to PASS.

## 9. Minimal remediation

### A. Must fix before M2 (M1 acceptance; maps 1:1 to blocking findings)

1. Reject Correct on intervals linked to non-draft timesheets + reissue flow
   (SYN-P0-1); add the cross-service correct-after-approve test.
2. Cap pairing windows + emit span-breach exceptions; wire NeedsReview into
   checkEligible (SYN-P0-2); revive or remove the dead 72h codes.
3. NEW-parent checks in all freeze triggers + re-parenting freeze cases
   (SYN-P1-1); scope trigger lookups by employer (SYN-P3-3, same edit).
4. Overlap exclusion constraint + Generate/Correct overlap checks + test
   (SYN-P1-2); reject sibling-overlapping corrections (SYN-P1-21, same edit).
5. TZ-aware business-date comparisons in checkEligible + unpaired filter +
   near-midnight tests (SYN-P1-5).
6. Conflicting-redelivery conflict signal (SYN-P1-3); leave unclassifiable
   breaks unpaired + persist the exception (SYN-P1-4); retain exception
   flags across Correct (SYN-P1-6).
7. Honor MIGRATION_DATABASE_URL; fail-loud 00004 preconditions; post-migrate
   grant verification; per-command role docs; one app-role Generate→Approve
   smoke test on compose grants (SYN-P1-7); run service suites as app
   (SYN-P1-24, same effort).
8. Close-prior-row in SetStatus + double-transition test (SYN-P1-8);
   snapshot-or-union Assess population vs retroactive changes (SYN-P1-9).
9. Publish validation gate + negative tests (SYN-P1-10); Revise service on
   CLI+REST (SYN-P1-11).
10. GET intervals endpoint (SYN-P1-12).
11. Serialize fresh-DB setup (`-p 1` in `task test:db` now; 23505-tolerant
    setup later) (SYN-P1-13).

### B. Should fix during M2 (first slices; cost grows with delay)

Unified envelope + request_id + pg-text redaction (SYN-P2-2); Paginated
lists + pagination (SYN-P2-1); Idempotency-Key enforcement (SYN-P2-3);
money-as-decimal + Go validation (SYN-P1-19); Go-side date/time validation
→ 422 (SYN-P2-14); Import/Correct/Generate atomicity + replay repair
(SYN-P2-4/5); Assign/publish locking + idempotent transitions (SYN-P2-6/7/12);
run-id validation (SYN-P2-8); attribution/audit decision + persistence
(SYN-P2-16); persisted exceptions + review surfacing (SYN-P2-17);
reject-resubmit + draft edit (SYN-P1-14, SYN-P2-24); Assign/Template/end-date
validation (SYN-P1-15/16/17); status transition machine (SYN-P1-18);
onboarding commands (SYN-P1-23); rotation-list fix (SYN-P1-22);
route/parity test tables (SYN-P2-26); concurrency tests (SYN-P2-31);
assertion repair (SYN-P2-29); empty-run contract (SYN-P2-30).

### C. Defer (explicitly not M2-gating)

Rotation/skill REST parity (SYN-P2-27, LIM-4/5); dead-403 cleanup
(SYN-P2-25); body/file caps until any non-loopback exposure (SYN-P2-28,
SYN-P3-9 — then mandatory); Down hardening + round-trip test (SYN-P3-4)
unless rollback is exercised; DSN-guard marker (SYN-P3-11); lineage-ref
precision + assumption filter (SYN-P3-5/6); 404/405 envelope (SYN-P3-7);
CLI format/IDs polish (SYN-P3-8/10); Dockerfile/fly + Taskfile demo cleanup
(SYN-P2-32, SYN-P3-12); CSV nits (SYN-P3-13); handler query sharing
(SYN-P3-1); Resolve hardening (SYN-P3-2); false publish summary rewrite
(SYN-P2-13, cheap — do with SYN-P1-10); rotation pattern validation
(SYN-P2-11) until a consumer exists; §4g hypotheses 1–20 triaged into M2
slices (cross-midnight #1 and overlapping periods #4 first, as potential
correctness defects).

## 10. Explicitly not independently verified

Live concurrent-race execution (all races static); live 404/405/prod
bodies; browser CORS; `docker build`; `seed demo --reset`; rollback and
upgrade-from-baseline runs; full E2E 12-step re-run and its numeric claims;
prod DSN role and server timezone; SG_research.pdf / rule-source-register
bytes; per-route status/validation/exit-code matrices; resource hygiene;
the prompt8 B live battery; config/DSN error-path leakage; §4g items 1–20
as listed. Mitigating: fresh `migrate up` + `seed demo` + CLI list, the
full serial suite, the parallel-race reproduction, OpenAPI structural
validation, and lint/build were executed live on disposable infrastructure
(see §2–3); every other finding rests on joint code+schema+test-body
inspection with exact file/line citations and disposable-DB repros.

*Finding counts: 2 P0 · 24 P1 (13 blocking) · 33 P2 · 13 P3 · 5 risks ·
20 untested hypotheses. Provenance: result 1 (A) · result 5 (B/K) ·
result 6 (C/D) · result 3 (E/F) · result 2 (G/H/Phys) · result 4 (I/J/L) ·
result 7 (execution/S7) · result 8 (critic gaps).*





