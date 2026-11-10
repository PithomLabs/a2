# Milestone 1.3 Payroll-Input Integrity Closure — Remediation Report

Execution of `plans/prompt13.md` as amended by `plans/plan_prompt13_review.md`
(verdict APPROVE WITH AMENDMENTS, Amendments A–D incorporated), per
`plans/plan_prompt13.md`. Single forward migration `00011`; everything
uncommitted; `00001`–`00010` byte-identical to baseline; `bchat/` untouched;
OT-001 stays BLOCKED; no statutory calculations added.

## 1. Initial Git baseline and uncommitted state

- `rev-parse HEAD`: `f435158` ("Phase 1 scaffold baseline (pre-pivot)"),
  unchanged at completion; no commits, amends, rebases, or tag moves.
- Phase 0 record (`/tmp/m13-baseline.txt`): 75 tracked / 140 untracked
  entries, plus sha256 for `00001`–`00010`.
- Post-battery re-verification (`sha256sum -c` against the baseline):
  all ten files OK.
- One new migration only: `migrations/core/00011_m13_integrity.sql`
  (1321 lines, Up plus guarded Down). No second migration was needed.
- `git status` at completion: all M1.3 work uncommitted as required;
  `git diff --check` clean.

## 2. Findings reconfirmed

Every finding below reproduced red before its fix (new tests failed on
the pre-fix tree; details per area in §§4–8):

- A1-F1 (correction lineage not enforced): multi-predecessor groups,
  stray segments, and orphan predecessors all insertable.
- A1-F2 (partial re-correction): one segment of a two-segment group
  corrected alone, forking business meaning.
- A1-F3 (group atomicity): no whole-unit rule in `Correct()`.
- A1-F4 (business-date attribution): vendor timezone accepted for
  `local_date`.
- A1-F5 (backdated evidence): caller `created_at` on clock events and
  intervals persisted verbatim (2020 stamps observed).
- B-CYC-01 (disposition cycles): canonical→duplicate chains allowed
  sequentially, and concurrent reversed A→B/B→A pairs both committed.
- D-RACE (compensation/approval race): no serialization; wall-clock
  comparison missed commit-late winners.
- D-CONTAIN (containment): compensation outside parent-terms span
  accepted; children outlived closed terms.
- D-ROUND (precision): `NUMERIC(12,2)` silently rounded 199.999→200.00.
- D-AUDIT (audit gaps): unaudited direct writes possible.
- D-NOCORRECT (no controlled correction): no successor workflow.
- E-DOWN (rollback safety): no guarded Down; no documented policy.
- S7-P2-FRESHFLAKE: Phase 0 fixed the warm/cold setup race, and the
  §14F battery caught a residual variant (advisory role lock taken on
  the per-test database, where it cannot exclude across databases):
  one clean-server pass-1 failed with `pg_authid_rolname_index`
  23505. Fixed and regression-tested (§11).
- S7-P2-MATRIX-OPEN: allowlist matrix stayed green on unclassified
  tables inheriting default full DML.
- H1–H7: disposition comment accuracy (fixed in the 00011 trigger
  header), audit UTC rendering, contract/checkEligible TZ wording,
  programme timezone validation, overlap day-status consistency
  (status=reviewed while blocked, reproduced), legacy EXECUTE, and
  battery counts.

## 3. New migration and schema changes

`00011` Up lands, in order: (a) the Amendment A payroll-input barrier
(`acquire_payroll_input_lock`, `pay_basis_seq`, five `basis_seq`
columns, per-row assign triggers, the approval barrier trigger);
(b) Stage A group lineage (link table, same-xid open rule, deferred
completeness/uniform-kind/binding/scope/sibling checks, employment-TZ
`local_date` trigger); (c) Stage E2–E5 compensation lifecycle
(`employee_id` + `supersedes_id` + single-successor UNIQUE,
exact-scale CHECKs on `amount_monthly`/`monthly_basic`, containment +
live-basis-overlap trigger, `add_compensation`/`correct_compensation`
DEFINER functions, INSERT revoke, `close_employment_terms`
child-bounds); (d) Stage C disposition pair-lock + bidirectional
recheck (OR REPLACE of the 00010 trigger, pinned search_path kept);
(e) Stage D `created_at` normalization on clock events and intervals
(conditional on UPDATE so the supersession flip preserves history);
(f) H6 EXECUTE revokes on seven legacy trigger functions;
(g) H4 programme-timezone probe. Every block ends with a verify DO
that fails the migration loudly on drift. Down is guarded
forward-only (§9).

## 4. Correction-group enforcement

DB is the authority, Go the ergonomic path. Groups bind predecessors
through the link table with `UNIQUE(predecessor_interval_id)` (never
per-interval unique); a deferred trigger enforces set completeness,
uniform kind, bidirectional binding, scope, and sibling completeness;
the same-xid open rule keeps later appends out; `local_date` pins
from the employer jurisdiction (SG→Asia/Singapore, unknown fails
loudly), mirroring `splitRun`. `Correct()` was rewritten group-atomic
(Stage B): it resolves the full current sibling set, rejects partial
sets with `partial re-correction of group … requires … missing …`
(CodeValidation), binds every predecessor, assigns each segment to
its first overlapping predecessor with an orphan-repair pass, and
regenerates every touched period atomically. Deliberate deviation
from plan text: the plan's "new bounds must cover the union" rule
was dropped — explicit bounds are operator intent, set-completeness
is the complete A1-F3 fix, and coverage would forbid legitimate
moves (demonstrated by the cross-period test). Prompt 13 §5 never
asks for coverage; all ten §5 behaviors hold.

Tests: 17 group-matrix + 6 date-attribution cases (direct SQL as
app, §14A), 9 re-correction lifecycle tests + ListIntervals group-id
test, and the A4 two-segment sheet lifecycle (summed 8h, conflict
until approval). 22/22 + 10/10 + 1/1 green.

## 5. Duplicate disposition concurrency protection

Canonical policy, enforced in the trigger with a pair advisory lock
(`disposition:<lo>:<hi>`) plus post-lock recheck in both directions:
many duplicates per canonical; a duplicate cannot become canonical;
a canonical cannot become a duplicate; no chains, no cycles. Lock
order (employee barrier above pair lock; direct SQL takes only the
pair lock) admits no cycle. Service parity check added. Out of
scope, observed: direct-SQL dispositions bypass the employee
barrier against approvals (Prompt 13 §6 asks only for
disposition/disposition serialization).

Tests: §14B matrix (multi-duplicate positive, both chain
directions, sequential cycle, concurrent reversed pair in both
start orders with evidence/usability/residue assertions,
byte-identical evidence) at direct-SQL level, plus the service
parity test with audit-residue assertions. 6/6 + 1/1 green.

## 6. Compensation and approval transaction barrier

Amendment A + D-RACE: shared advisory lock
`pay-input:<employer>:<employee>` (single helper,
`acquire_payroll_input_lock`, called by Go and every trigger) plus
`pay_basis_seq`, where sequence order equals commit order per
employee; readiness treats a row as post-approval when EITHER
`created_at` or `basis_seq` exceeds the approval's (legacy seq 0
keeps the wall-clock leg meaningful). Approval transitions assign
the sequence in-trigger so raw SQL cannot bypass. Rejected, as
planned: `track_commit_timestamp` (server-config dependency) and
`clock_timestamp()` (tie risk).

Tests: 6 live two-connection barrier races in both winner orders
(Import/Correct vs Approval, Terms/Comp/Status/SHG vs Approval),
all green, plus the comp-scenario-3 normalization-divergence proof
and the attendance D-RACE regression that fails if the seq leg
ever regresses while timestamps stay green.

## 7. Compensation containment and precision enforcement

Containment trigger (effective span inside parent terms, closed-end
aware, employee pinned to the terms row); closing terms refuses
live lines that would outlive the close (superseded history never
blocks; the controlled path caps first via correction). Overlap is
an employee-level live-basis invariant validated in-trigger under
the function-held lock — a deliberate mechanism change from the
plan's exclusion sketch, because an exclusion cannot express
liveness: a superseded original keeps `supersedes_id` NULL, so any
exclusion covering it would reject legitimate rows in a
narrowed-away range (proven by the narrowing test). The 00004
per-terms exclusion is dropped (Down restores it). Precision:
unconstrained NUMERIC with exact-scale + non-negative + <10^10
CHECKs on `amount_monthly` and `monthly_basic`, matching Go
`validateMoney` digit-for-digit (Go==SQL parity tested both
directions, including the 9999999999.99 boundary). Persisted values
match the exact input scale; the audit canonicalizes to cents.

Tests: 15-case direct-SQL lifecycle matrix (§14C: containment ×3
with the cap-then-close flow, cross-terms refusal with adjacency
positive, concurrent overlap single-winner, historical+future
positive, precision ×2, revoke, audit canonicality, draft/future
correction, concurrent successors, cross-employee). Finding
documented in-test: terms are strictly disjoint by exclusion, so a
cross-terms reach-back necessarily escapes its own terms and
containment refuses it; the employee-level rule stands behind as
the backstop. 15/15 green.

## 8. Compensation correction and audit lifecycle

All writes route through `add_compensation`/`correct_compensation`
(DEFINER, pg_catalog, employee lock + validation + row + one
canonical audit event atomically; failed writes leave no audit
shadow). Successors preserve classification, overlap only live
rows through their direct predecessor, and chain (correcting the
live successor stays open); the original is never mutated; double
succession reports `already has a successor` consistently
(sequential and concurrent). Raw INSERT/UPDATE/DELETE all revoked.
Approved-period corrections are allowed-and-flagged per the review
decision: readiness carries `created after approval (correction;
reissue required)` until reissue+approve. Go service calls the
functions (no duplicate audit), with `reason` mandatory end to end
(CLI/REST/OpenAPI).

Tests: service correction + validation/parity tests, the
approved-period allow-and-flag flow (flag → reissue → clear), and
the full §14C matrix above. Green.

## 9. Migration rollback policy

M1.2+ integrity migrations are forward-only once dependent business
data exists (documented in `docs/workforceops/database-design.md`
§10): no Down in the application or Taskfile (verified: zero `Down`
references in product code; the only `down` is `docker compose
down -v`), populated Down+Up safety not claimed, destructive
recovery via verified backup + controlled restore only. `00011`
Down inspects first — any compensation/terms/group/link rows or any
nonzero `basis_seq` fails with `00011 Down refused` before anything
destructive — and unwinds fully only on empty databases (Down
restores the 00010 trigger bodies verbatim, the 00004 exclusion,
`NUMERIC(12,2)` types, the single-predecessor column, INSERT grant,
and legacy EXECUTE).

Tests: `TestDownRefusesPopulatedDatabase` (populated → refusal,
schema intact at v11) and `TestDownEmptyDatabaseRoundTrip`
(15-assertion shape check + Up re-applies). 2/2 green.

## 10. Fail-closed privilege matrix

The matrix enumerates every table (`relkind r/p`) and every owned
(non-extension) function from the catalog: 55 tables with explicit
SELECT/INSERT/UPDATE/DELETE classifications (16 insert-only
history, compensation sealed, 37 mutable operational, migration
bookkeeping pinned as-is), 30 functions with
owner/definer/search_path/body/EXECUTE pins. Any unclassified table
or function fails; stale entries fail. Plus: zero PUBLIC table
privileges, the single default-ACL rule pinned
(`migration_admin public r {app=arwd/...}` — the hazard the matrix
guards), trigger-owned `basis_seq` caller-control rejection, and
the H6 revokes. §14E's 11 app-role cases are covered by the stage
matrices, each with persisted-state re-reads after rejection.

Tests: 7 privilege tests green; the suite fails closed by
construction (an unmapped table aborts the run).

## 11. Clean-server test evidence

`TestFreshInstallChain` (ephemeral cluster, 11 migrations, role
grants, app bootstrap) green. Full `go test ./...` on genuinely
fresh `initdb` clusters: pass 1 EXIT=0, immediate pass 2 EXIT=0
(idempotent setup, no recycle); `-race` on the five packages green
on a clean server with zero data races. During the battery, one
pre-fix clean pass-1 failed with the residual FRESHFLAKE
(`pg_authid_rolname_index` 23505): root cause — the global role
advisory lock was taken on the per-test database, where advisory
locks cannot exclude across databases. Fixed in `dbutil` and
`RecycleDatabase` (lock on the shared `postgres` maintenance DB)
with `TestParallelCleanServerRoleProvisioning` (8 parallel setups
on an ephemeral clean server): proven red on the buggy lock
(same 23505) and 3/3 green fixed. No warm-server-only PASS is
claimed anywhere; no serial mode was substituted for parallel runs.

## 12. CLI/REST persisted-state parity

`Correct()` contract migrated: `interval_id` → `interval_ids[]`
(CLI file format, REST `POST .../attendance/corrections` with
body ids, OpenAPI schema + corrections requestBody).
Compensation: `reason` mandatory in add input; new
`compensate-correct` CLI, `POST
.../compensation/{line_id}/corrections`, OpenAPI
`correctCompensation`. `ListIntervals`/`ListCompensation` expose
`correction_group_id`/`supersedes_id`. OpenAPI validation is the
bidirectional served↔spec contract test (green every leg); CLI/REST
E2E and persisted-state parity are the e2e suite (green every
leg). All behavior changes are covered by persisted-state re-reads
in the stage matrices.

## 13. Test commands and exact outcomes

Environment: `GOWORK=off` (parent `go.work` excludes a1), `/tmp` Go
caches, scratch PostgreSQL 16 on 127.0.0.1:5433, test DBs recycled
(`a1test*` dropped) before migration-sensitive legs.

- `task lint`: PASS (gofmt/vet/nodeps; `lstat examples` stderr is
  pre-existing Taskfile noise — `examples/` never existed).
- `task build`: PASS (`build/workforceops`).
- `GOWORK=off go test -count=1 ./...`: EXIT=0 — 17 packages ok,
  7 no-test-files, 429 top-level PASS / 0 FAIL.
- `TZ=UTC go test -count=1 ./... -v`: 429 PASS / 0 FAIL.
- `TZ=Asia/Singapore go test -count=1 ./... -v`: 429 PASS / 0 FAIL.
- `go test -race` on attendance/timesheet/payroll/hr/wfm:
  all ok, `DATA RACE` count 0 (both recycled and clean servers).
- Clean server: pass1 EXIT=0, pass2 EXIT=0; fresh-install green.
- `git diff --check`: clean.
- New M1.3 tests: 80 top-level across 18 new files (17 group,
  6 dates, 15 comp lifecycle, 6 disposition, 6 evidence/H4,
  10 re-correction, 6 barrier-race, 2 comp service, 2 timestamp
  service, 1 comp readiness flow, 1 disposition parity, 1 overlap
  status, 2 UTC rendering, 1 programme TZ, 2 Down, 1 role race,
  1 A4 lifecycle), plus the privilege-test extension to 7.

## 14. Deferred findings and reasons

- OT-001 (SG-MOM overtime) remains BLOCKED: no practitioner, no rule
  package activation, no calculation code; untouched this milestone.
- Per Prompt 13 §12/H: disposition listing, break-duplicate nuances,
  supersede workflow details, −1s pin, display rendering, Q11
  break-half, and deferred systems — explicitly out of scope, not
  started.
- Observed, not in scope: direct-SQL dispositions bypass the
  employee barrier against approvals (§6 asked only for
  disposition/disposition serialization); M1.2 `close_employment_terms`
  resets closed-terms `created_at` via the UPDATE normalization leg
  (established behavior, coherent with readiness, left alone).
- No P0/P1 found that can silently corrupt or stale approved
  payroll inputs.

## 15. Environment-dependent limitations

- Docker/Compose: daemon unavailable in this sandbox — no Docker
  build or Compose verification is claimed.
- No Neon/Fly/LLM credentials used or needed; mock LLM default;
  local PostgreSQL 16 + `GOWORK=off` as mandated.
- Scratch clusters ran on TCP loopback 5433 with trust auth;
  ephemeral test clusters on 5435; nothing reachable off-host.
- Parent `go.work` excludes a1: every Go command ran with
  `GOWORK=off` and `/tmp` caches (recorded in §13 commands).

## 16. Final PASS/FAIL verdict

**PASS.** §15 mapped: (1) A1-F1–F5 closed with DB enforcement +
matrices; (2) B-CYC-01 with concurrent direct-SQL proof both
orders; (3) D-RACE with 6 live races both winner orders;
(4) D-CONTAIN at DB level incl. close-bounds; (5) D-ROUND for
amounts and monthly basic with parity tests; (6) insertion and
correction controlled, single-audit, revoked raw writes;
(7) populated Down refuses pre-destruction, no supported Down
tooling, policy documented; (8) matrix fails closed (tables and
functions enumerated); (9) clean-server provisioning green without
retry, twice; (10) M1.2 invariants intact (429/429 incl. all legacy
tests, old migrations byte-identical); (11) full workflow green
under default/UTC/Singapore hosts; (12) no silent-corruption P0/P1
remains; (13) OT-001 BLOCKED; (14) no statutory calculations added.
Stop condition met: integrity remediation ends here; next is the
first independently verified payroll calculation.


