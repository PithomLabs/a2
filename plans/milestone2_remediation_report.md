# Milestone 2 Remediation Report — Payroll Governance and Reproducibility Foundation

## 1. Initial Git baseline and uncommitted state

- Baseline: HEAD `304fec8` (release 0.1.3), clean tree at Phase 0.
- SHA-256 of `migrations/core/00001`–`00011` recorded at baseline
  (`/tmp/m2-baseline-sums.txt`); re-verified byte-identical at the
  acceptance gate (`diff` clean, §13).
- Final state: 28 tracked files modified, 19 new files/directories,
  all uncommitted. `bchat/` untouched (read-only). `docs/writeup3.md`
  is a pre-existing untracked user file, not Milestone 2 work.
- `task lint` was red at baseline (the Taskfile's gofmt scope named
  `examples/`, which never existed in git). Repaired minimally by
  scoping to existing module dirs; the bchat exclusion rationale is
  preserved. Reported as a pre-existing finding, not Milestone 2
  behavior.

## 2. Six prerequisite findings and disposition

Implemented in dependency order P3 → P4 → P5 → P2 → P1 → P6 per
`plans/plan_prompt15.md` as amended (A–F). Review findings
SHG-B1a/B1b/B1c, SHG-B2a/B2b/B2c, C1, C2, C3, C4-STATUS, C6 and S5-F1
are closed; each section below names its evidence.

P6 verification surfaced genuine P1 debt, fixed in-milestone rather
than waived:

- Validation never verified required rule-pack activation (§9C step
  6): a run over BLOCKED packs validated. Fixed with recorded-pin
  assessment in the service and a mirrored trigger check.
- No correction-run successor relationship (§9F) and no
  multiple-validated-run prevention (§9A): added the link, the
  chain/singleton rules and their proofs.
- Direct SQL could flip `draft → validated` with no snapshot, over a
  blocked tip, or behind a standing validated run: the lifecycle
  trigger now enforces tip, unblocked, singleton, chain and
  identity-freeze rules for every role.
- The fixed-name reusable test databases (`a1test_integrity`,
  `a1test_e2e`, …) went stale when `00012` was extended: goose
  records a version once, so appended sections never applied to
  existing databases. The integrity suite passed for a full phase
  without the lifecycle trigger. All `a1test*` databases were
  dropped; affected tests were rewritten against the real trigger;
  the battery procedure recycles first (§13, §14).

Zero statutory calculations were implemented. `SG-MOM-OT-001` remains
BLOCKED (manifest `BLOCKED`, registry unpinned, no IMPLEMENTABLE
row). One migration (`00012`) was added; `00001`–`00011` untouched.

## 3. Compensation resolver design and evidence

`internal/compbasis/resolve.go` (new leaf package; placed outside
`internal/hr` to avoid a payroll↔hr test import cycle) exposes
`ResolveEffectiveCompensation`, returning per-classification ordered,
non-overlapping `[]EffectiveSegment`
`{record ID, classification, effective interval, exact amount,
terms ID, supersession lineage}` plus a machine-readable
`BasisConflict` type.

- Superseded predecessors are excluded via the successor chain;
  multiple live rows covering one date or broken chains return a
  conflict — never summed, never `LIMIT 1`.
- Readiness and snapshot capture call this one implementation; no
  second resolver exists.
- The first restricted profile blocks validation on any
  classification resolving to more than one segment in the period
  (mid-period-change treatment unimplemented), with a coded blocker.

Evidence: 16 resolver tests (`B1`–`B10` chain semantics + §10A
matrix), `internal/payroll/readiness_resolve_test.go`,
`internal/hr/resolve_test.go`. All green, including under `-race`.

## 4. SHG lifecycle and evidence model

Statuses are explicit: `pending`, `active`, `opt_out`,
`verified_none`, with the verified distinction carried by the
`verified_at/by/basis` triple. Services: `SetSHG` (record),
`VerifySHG` (pending-only; mints the verified successor, never
mutates), `CorrectSHG` (governed successor; never regresses to
pending), `ListSHG`.

- `verified_none` requires the full evidence triple; a triple-less
  `verified_none` tip is a `SHG_CORRUPT` blocker.
- Succession follows the M1.3 compensation precedent: one successor
  (`UNIQUE (supersedes_id)`), original immutable, same-fund
  live-overlap rejection via trigger, successor-range subtraction.
- Readiness is status-aware (`SHG_INSTRUCTION_MISSING/PENDING/
  OPT_OUT/CORRUPT`, coverage by any active or fully-evidenced
  `verified_none` tip); snapshot capture calls the same
  `checkSHGCoverage`, so both paths agree by construction.
- Post-approval SHG changes conflict visibly (`checkSHGConflict`,
  mirroring the term/attendance/comp conflict detectors).
- `opt_out` blocks until its semantics are explicitly defined
  (reversible default); cross-fund multiples are rejected as
  ambiguous for the first profile.

Evidence: `internal/hr/shg_lifecycle_test.go`, §10B matrix,
`tests/integrity/shg_governance_test.go` (direct-SQL
UPDATE/DELETE/INSERT as `app`, audit atomicity), readiness SHG
tests. No SHG contribution calculation exists.

## 5. Rule-pack activation governance

- `rule_pack` registry holds 10 identities, all unpinned
  (`reviewed_version NULL`); nothing is activatable until a later
  migration pins a reviewed version.
- `rule_pack_activation_history` is append-only with
  supersession-preserving links; `activate_rule_pack(...)` is a
  DEFINER function validating identity, exact version, dates,
  jurisdiction, scope, manifest digest, evidence refs, spec version,
  review record, verification status, actor and reason.
- `app` holds no INSERT/UPDATE on activation, history, or the six
  statutory parameter tables; activation is CLI-only
  (`admin activate-rule-pack`, privileged DSN, attribution not
  authentication). No REST activation endpoint exists.
- The CLI reads the on-disk manifest, asserts implementable
  lifecycle state, verifies referenced evidence/spec/review/fixture
  files exist, recomputes SHA-256 digests, and submits
  identity/version/digest/refs; the DB function rejects unknown
  packs, invented versions and illegal transitions. Audited
  atomically.
- Tests use temp-dir fixture packs only; real packs keep actual
  states. OT-001 manifest `BLOCKED`, unpinned, never activated.

Evidence: §10C matrix (`internal/rules/*_test.go`,
`tests/integrity/activation_governance_test.go`,
`cmd/workforceops/admin_test.go` incl. `TestOT001RemainsBlocked`).

## 6. Payroll-run state machine

Only `draft → validated` is reachable. Runs are born draft;
`calculated`, `approved` and `finalized` are rejected structurally
by the `run_lifecycle` trigger for every role — no artifact table
was needed in this milestone because no transition to those states
can satisfy the trigger. The trigger additionally enforces, for
service and direct SQL alike:

- Exactly one valid snapshot tip at validation time.
- Tip terminality: every employee decision `included`/`excluded` and
  every pinned pack IMPLEMENTABLE (mirrors the service assessment,
  so direct SQL cannot bypass the blockers check).
- One validated regular run per employer/period (trigger message +
  partial unique index backstop for races).
- Linear correction chains: `corrects_run_id` links the original in
  the same employer/period (CHECK coherence + FK + scope check); a
  correction validates only after its original; each validated run
  takes at most one validated successor (trigger + partial unique
  index); forks rejected.
- Identity freeze (employer, period, kind, link) once any snapshot
  exists.

`ValidateRun` (validate-by-version): population/employee locks,
expected population, compensation/timesheet/SHG/pack/conflict
assessment, atomic capture, content hash, machine-readable blockers,
draft preserved on refusal. Stale tips invalidate + recapture with
an explicit re-offer verdict — never silent advance. Re-offer is
idempotent; concurrent racers serialize on the run row and the loser
receives a stale verdict naming the winner's tip (never a silent
adoption). Empty populations refuse without mutation (no vacuous
truth). Exclusions stay draft-only and run-scoped.

Evidence: `internal/payroll/validate_test.go`,
`run_lifecycle_test.go`, `runs_test.go`, §10E matrix.

## 7. Snapshot schema and canonical hashing

- `payroll_snapshot` header (run, schema version `m2/v1`, hash) +
  `payroll_snapshot_employee` rows (terms/status/sheet + line and
  interval IDs, comp-segment IDs + intervals, SHG record + evidence
  refs, per-employee `basis_seq`, inclusion decision + blockers) +
  run-level `payroll_snapshot_pack` pins (pack, version,
  jurisdiction, interval, activation record, digest, evidence refs).
- Canonical hash: stable-ordered JSON in Go (`crypto/sha256`, exact
  amount strings, sorted collections), excluding run/snapshot UUIDs
  and all timestamps — identical substantive inputs hash identically
  across runs. Proven: two draft runs over the same inputs capture
  byte-identical `content_hash` (`TestM2Workflow` step 14,
  `TestCaptureHashDeterminism`).
- Lifecycle: immutable content vs append-only
  `snapshot_lifecycle_event`; at most one valid tip per run
  (partial unique + trigger); freeze trigger + revoked `app` writes;
  drift invalidates and revalidation inserts v(n+1) — history is
  never rewritten. Re-reads confirm the original row keeps its hash
  with `invalidated` status.

## 8. Snapshot concurrency and population consistency

- Capture and validation run at REPEATABLE READ (`BeginSnapshotTx`)
  so all reads see one committed view; advisory locks order capture
  against writers (locks alone are not a view guarantee).
- One global order: employer population lock →
  employee pay-basis locks in UUID-sorted order → per-required-pack
  activation keys → population revalidation. `CreateEmployee`,
  `SetTerms`, `SetStatus`, compensation writers, approval and
  activation all join the convention; BEFORE INSERT triggers on
  `employee`/`employment_terms` take the population lock so direct
  SQL cannot slip population changes past capture.
- Reader transactions cannot hit 40001; no retry loop exists or is
  needed. Concurrent commits during capture stay invisible to the
  capture and surface afterward as drift, never as mixed versions.
- 8-way concurrency races (employee creation, term change,
  termination, comp insertion, timesheet approval, attendance
  correction, SHG change, pack supersession) plus isolation proofs,
  all with separate connections and real sleeps/holds.

Evidence: `snapshot_race_test.go`, `snapshot_governance_test.go`,
barrier-race harnesses, `-race` clean.

## 9. Database constraints, triggers and privileges

Constraint/trigger inventory added by `00012` (all owned by
`migration_admin`, `search_path=pg_catalog`, schema-qualified
bodies, pinned in `tests/integrity/privilege_test.go`'s fail-closed
function catalog, now including `check_run_lifecycle`):

- SHG: `shg_status_values`, `shg_verified_consistency`,
  `shg_effective_order`, single-successor UNIQUE, supersedes FK,
  lifecycle/overlap trigger, `basis_seq` trigger, `add/verify/
  correct_shg_instruction` (DEFINER, app-executable), created-at
  normalization; `app` UPDATE/DELETE revoked.
- Activation: history append-only design, `activate_rule_pack`
  (DEFINER, deliberately not app-executable),
  `ensure_pack_blocked`; `app` INSERT/UPDATE revoked on activation,
  history and all six parameter tables.
- Snapshot: one-valid partial unique, freeze trigger, lifecycle
  trigger, `invalidate_payroll_snapshot` (narrow DEFINER),
  `acquire_payroll_population_lock` /
  `acquire_rule_pack_lock` / `take_payroll_population_lock`.
- Runs: `payroll_run_kind_link_ck`, `payroll_run_corrects_fk`,
  `payroll_run_one_validated_regular`,
  `payroll_run_one_validated_successor`, `run_lifecycle` trigger.
- Privilege matrix: every public-schema table and owned function
  classified; PUBLIC holds EXECUTE on none of the governed
  functions; `app` cannot mutate snapshots, activations, parameters
  or SHG rows outside the narrow functions.

## 10. CLI/API changes

Minimum ops per §11, with CLI and REST sharing the same services
(`InspectRun`, `InspectSnapshot`, `ValidateRun`, `CreateRun`,
`ListSHG`/`VerifySHG`/`CorrectSHG`, `rules.ListActivations`):

- CLI: `payroll run-create` (+`--corrects`), `payroll exclude`,
  `payroll readiness`, `payroll run-validate --offered-version`,
  `payroll run-show`, `payroll snapshot-show`, `payroll rule-packs`,
  `employee shg`, `employee shg-verify`, `employee shg-correct`,
  `admin activate-rule-pack` (CLI-only activation).
- REST: `POST payroll-runs` (+`corrects`), `POST .../exclusions`,
  `GET payroll-runs/{id}`, `POST payroll-runs/{id}/validate`,
  `GET snapshots/{id}`, `GET payroll-readiness`,
  `POST/GET employees/{id}/shg`,
  `POST .../shg/{instruction_id}/verify|correct` (with ownership
  checks), `GET rule-activations`. No activation endpoint.
- `openapi.yaml` marks every served route `implemented`; the
  contract test enforces route↔spec parity in both directions and
  forbids serving draft ops. Mutations require `X-Operator`
  attribution, documented as attribution, not authentication.
- Parity: `TestCLIAPIParityM2` writes through one adapter and reads
  through the other for the full M2 flow (roster→clocks→timesheet→
  run→validate→inspect→packs→SHG chain); persisted state agrees.

## 11. Migrations

One new forward migration: `migrations/core/00012_m2_governance.sql`
(SHG block, RULES block, SNAPSHOT block, RUNLIFE block), each with a
verify block asserting its objects stand. Guarded forward-only Down:
refuses populated databases before anything destructive (SHG rows,
activation history, non-default activations, snapshot rows,
correction links), unwinds cleanly only on freshly migrated empty
databases, and reverses each section in dependency order.

- `TestDownRefusesPopulatedDatabase`, `TestDownEmptyDatabaseRoundTrip`
  green; clean-server provisioning (`TestFreshInstallChain`,
  `TestInitRolesScript`,
  `TestParallelCleanServerRoleProvisioning`) green twice.
- No Down path is exposed to product (`migrate`, Taskfile): the Down
  exists for development/test databases only.
- `00001`–`00011` byte-identical to baseline (verified, §13).

## 12. Independent regression tests

§10 matrices, all on real PostgreSQL 16 as `app` (direct-SQL
adversarial tests) with disposable databases:

- §10A resolver: 10 chain semantics + matrix + mid-period blocker.
- §10B SHG: 15 cases incl. direct-SQL rejections and audit
  atomicity.
- §10C activation: 12 cases on isolated fixture packs; OT-001
  BLOCKED asserted at manifest, registry and activation level.
- §10D snapshots: 26 cases incl. 8 concurrency races, hash
  determinism/difference, direct-mutation rejection, drift,
  invalidation, revalidation.
- §10E lifecycle: creation, exclusion, blockers, validation,
  direct-SQL rejections (approved flip, reopen, snapshotless,
  blocked-tip, second-regular), artifact-absence proofs, replay,
  concurrency, post-validation drift, correction chain, fork
  rejection.
- §10F workflow: `TestM2Workflow` runs all 15 steps on the seeded
  MERIDIAN profile and proves the blocked-run outcome, cross-run
  hash identity and preservation under drift.
- §10G battery: §13. No database-gated test skips silently; no
  superuser execution is used as app-role proof (admin DSNs appear
  only for privileged fixture setup).

## 13. Exact commands and results

Scratch PostgreSQL 16 on loopback (`/tmp/pg16data`, port 55433);
`GOWORK=off GOCACHE=/tmp/gocache GOMODCACHE=/tmp/gomodcache`;
`WORKFORCEOPS_TEST_DATABASE_URL=postgres://postgres@127.0.0.1:55433/
a1test?sslmode=disable`. Databases recycled before every full run
(terminate `a1test%` backends + drop all `a1test*`):

- `task lint` → PASS (gofmt clean, `go vet ./...` clean, no
  `lib/pq`; includes the pre-existing `examples/` scope repair).
- `task build` → PASS (`build/workforceops`, ignored by git).
- `go test -count=1 ./...` (default TZ) → PASS, 19 packages ok,
  0 FAIL.
- `TZ=UTC go test -count=1 ./...` → PASS, 19 ok, 0 FAIL.
- `TZ=Asia/Singapore go test -count=1 ./...` → PASS, 19 ok, 0 FAIL.
- `go test -race -count=1 ./internal/attendance/
  ./internal/timesheet/ ./internal/hr/ ./internal/payroll/
  ./internal/wfm/` → PASS, 0 DATA RACE.
- `go test -count=1 ./tests/provisioning/` twice → PASS, PASS.
- `go test ./core/database/ -run TestDown` → both PASS.
- Contract test (`./internal/httpapi/`), CLI/REST E2E
  (`./cmd/workforceops/` incl. onboarding + both parity tests),
  integrity, e2e → PASS inside the full runs above.
- `git diff --check` → clean. `00001`–`00011` sums identical to
  baseline. `bchat/` untouched. OT-001 manifest `BLOCKED`.

## 14. Remaining limitations and risks

- Fixed-name reusable test databases + goose version-once
  semantics can hide appended migration sections (found in P6:
  integrity ran a phase without the lifecycle trigger). Standing
  procedure: drop all `a1test*` before every full run. A harness
  change (hash-pinned migrations or unique DBs per run) is future
  work, not Milestone 2 scope.
- Direct SQL on a genuinely clean tip can flip `draft → validated`
  (outcome-equivalent: tip, terminality and singleton checks still
  apply) but skips pin-stamping and audit. The governed path is the
  service; the gap is unaudited metadata, never a bypass of the
  blockers assessment.
- `opt_out` SHG blocks by default pending defined semantics
  (reversible: the blocker code stays, the rule narrows later).
- `calculated`/`approved`/`finalized` and their pre-M2 freeze
  helpers are structurally unreachable; run deletion is governed
  only by foreign keys (history-backed runs undeletable, bare
  drafts deletable); `payroll_line` stays mutable post-validation
  until a calculator pins it. All three are first-calculation
  milestone concerns.
- Lock-order surface is wide (population → employee → pack across
  services); the single documented convention + deadlock-sensitive
  race tests hold it. New writers must join the convention.

## 15. Explicitly deferred capabilities

Per §13, all deferred with dependencies: actual CPF calculation,
actual SDL calculation, SHG contribution calculation, overtime,
gross/net payroll, itemised payslip financials, salary payment and
banking integrations, statutory submission, payroll taxation
(depend on this milestone passing + pinned IMPLEMENTABLE packs +
golden/anti-golden/boundary fixtures + independent review);
cross-fund SHG semantics (research); full multi-employer auth, RLS,
SSO (identity milestone); additional jurisdictions; generic
workflow/rules engines; Phys-Pseudo interpreter/compiler;
microservices; event bus; UI; advanced WFM forecasting; rotation
materialization; production deployment. No placeholder financial
values or fabricated successful payrolls were introduced: every
blocked run in this milestone stays observably blocked.

## 16. Final PASS/FAIL decision

**PASS.** All six prerequisites satisfy §14:

- P1 Run governance: draft→validated enforced in service and
  trigger; direct SQL cannot bypass validation (tip, terminality,
  singleton, chain, freeze); calculated/approved/finalized
  structurally unreachable; identity and exclusions protected.
- P2 Snapshot: population correct and lock-ordered; capture atomic
  at REPEATABLE READ; compensation, timesheet, employment, SHG and
  pack versions pinned; hash genuinely content-bound (cross-run
  identical); immutable with append-only invalidation; drift
  detectable; revalidation versions, never rewrites.
- P3 Resolver: one effective basis per employee/classification/
  period; superseded excluded; ambiguity blocks; historical and
  future periods resolve.
- P4 SHG: missing/pending/applicable/verified-none distinct and
  status-aware; verified-none evidenced; post-approval changes
  detected; history protected; controlled succession exists; no
  unsupported statutory inference.
- P5 Activation: direct app activation impossible; administrative
  activation controlled and audited; identity/version/evidence
  pinned; unsupported versions and illegal transitions rejected;
  OT-001 BLOCKED at every level.
- P6 Verification: real PostgreSQL throughout; direct-SQL tests as
  `app`; concurrency on separate connections; fresh-server
  provisioning green; CLI/REST parity proved on persisted state;
  no critical regression remains (19/19 packages × 3 timezones,
  race clean).

Stop at the acceptance gate. No statutory financial calculation was
implemented. The first actual Singapore payroll calculation may
proceed only on this foundation: pinned IMPLEMENTABLE packs,
immutable input snapshots, and independent golden fixtures.
