# Milestone 1.2 Payroll Input Integrity Closure — Remediation Report

Prompt: `plans/prompt11.md` as amended by `plans/plan_prompt11_review.md`
(Amendments A–G). Owner priority (§12): attendance correctness →
historical integrity → DB protection → install/onboarding → full regression.
Work executed via `go`; all changes remain uncommitted for review.

## 1. Initial Git state and baseline

- HEAD `f435158` ("Phase 1 scaffold baseline (pre-pivot)"), working tree
  uncommitted-but-unaltered per the frozen constraints: migrations
  00001–00009 untouched (single new append-only `00010`), committed
  history intact, OT-001 BLOCKED, no payroll calculations.
- Local-first constraints held throughout: scratch PostgreSQL 16
  (`/tmp/a1pg`, port 5433, TCP loopback, no Unix sockets), `pgx` only,
  single binary, no cloud/LLM credentials, `bchat/` read-only.
- One operational note: unit suites share one long-lived test database
  (`a1test`) whose goose version predates the 00010 appends, so stale
  `a1test*` databases were dropped mid-execution to force a fresh
  migrate; this is a local artifact of iterative runs, not a product
  issue (§9C covers the fresh-install path).

## 2. Confirmed findings and their disposition

Stage A reconfirmed all 21 findings against the baseline; all are now
closed by implementation (not by dismissal):

**P1 functional (8)** — B1 overlap disposition via
`clock_event_disposition` + operator-confirmed exclusion; B2
midnight-break containment (single-assignment tiling, NULL +
`no_roster_link` off-roster); B3 chain-aware Correct + readiness
`AttendanceConflict`; B4 compensation retro containment (terms
containment, insert-only, readiness `CompConflict`); B5 employment-TZ
hardening (`employer.EmploymentLocation`, SG→Asia/Singapore, unknown
jurisdiction fails loudly); B6 Correct split into per-date segments via
`attendance_correction_group`; B7 late-break linkage gate +
`AttachPending`; B8 effective-dated status gate in `validatePlacement`.

**App-role DB bypasses (10 findings → C1–C9)** — all closed in migration
00010 (this session): C1 unconditional currency trigger; C2 DEFINER
hardening + TEMP revocation from app and PUBLIC; C3 supersession
identity trigger; C4 single-chain UPDATE guard + `EffectiveApproved`
chain-walk rewrite; C5/C6 line/sheet identity triggers; C7 roster
parent transition trigger; C8 history insert-only + controlled
close/transition functions; C9 catalog/privilege regression suite.

**P2/P3 + E-stage (5)** — E1 canonical rest days (service + DB CHECK);
E3 exact integer microseconds (`internal/duration`, Go-computed,
quantized legacy decimals); E4 audit gaps (4 real gaps closed, 12 pins
on existing coverage — see §6); E5 reset safety (disposable-name
policy, `--target` confirmation, password-server cycle).

## 3. Structural/architectural observations

- The application role is now a true least-privilege runtime: history
  tables are insert-only, currency/lineage/identity are trigger-owned,
  and the only UPDATE/DELETE paths into history are two narrowly
  scoped SECURITY DEFINER functions (`close_employment_terms`,
  `transition_employee_status`). Services keep their Go-side
  validation (error codes preserved) and call the functions for the
  locked mutation; the functions re-validate as backstops.
- `pg_temp` shadowing is closed in depth: TEMPORARY revoked from app
  and PUBLIC at provisioning (superuser), DEFINER functions run under
  `search_path = pg_catalog` with schema-qualified references, and the
  migration verifies all of it loudly.
- Corrections are lineage, not edits: predecessors are never mutated
  (original-evidence-unchanged assertions hold), successors cite
  correction groups, and readiness detects post-approval change by
  `created_at > approved_at` — which is why `created_at` is
  trigger-normalized on both compensation and terms.
- Two prompt-mandated contract changes to existing tests (documented,
  not convenience rewrites): `TestRepairByNewRevision` now uses the
  legitimate revise shape (C7 kills its bare reopen), and
  `TestFreezeReparenting` re-parents within one chain so it still
  exercises the freeze guard rather than the new single-chain guard.

## 4. Files and migrations modified

- `migrations/core/00010_m12_integrity.sql` (817 lines): the single
  new migration — B1/B6/E1/B4/E3 sections plus C1–C9 (currency
  rewrite, TEMP verification, supersession/identity/roster triggers,
  single-chain UPDATE, terms normalization, two controlled functions,
  revokes, EXECUTE hygiene), each with a reverse-order Down block.
- Services: `internal/attendance` (TZ threading, disposition, attach,
  Correct split/chain), `internal/timesheet` (chains, regen,
  EffectiveApproved walk), `internal/payroll/readiness.go`
  (attendance/comp conflicts), `internal/wfm` (status gate,
  membership/skills audit), `internal/hr` (rest days, comp
  containment, function-based SetTerms/SetStatus),
  `internal/employer` (location + D3 Create),
  `internal/duration` (new), `core/database` (role refusal, recycle
  policy, password plumbing), `core/api/host.go` (serve refusal),
  `cmd/workforceops` (disposition/attach/employer/reset wiring),
  `internal/httpapi` + `openapi.yaml` (3 new routes).
- Provisioning/docs: `scripts/db/init-roles.sh` (rewritten),
  `tests/dbutil` (TEMP revoke), `.env.example`, `Taskfile.yml`,
  `README.md`, `docs/workforceops/database-design.md`.
- Untracked additions under review: `internal/`, `tests/e2e/`,
  `tests/integrity/`, `tests/provisioning/`, `migrations/core/00010*`.

## 5. Database constraints and triggers changed

New in 00010 (all owned by migration_admin, PUBLIC EXECUTE revoked):
triggers `attendance_interval_currency` (unconditional),
`attendance_interval_supersession` (employee/self/cycle),
`timesheet_line_interval_single_chain` (now INSERT OR UPDATE),
`timesheet_line_identity`, `timesheet_identity`,
`roster_transition` (reopen/pointer/stamp rules),
`employment_terms_created_at_normalize`; functions
`close_employment_terms` and `transition_employee_status`
(DEFINER, pg_catalog-only, EXECUTE to app); revokes of UPDATE/DELETE
on employee, terms, status history, memberships, skills; TEMPORARY
revoked from app+PUBLIC (provisioning) and verified in-migration.

## 6. Transaction and audit changes

- E4 survey found most lifecycle ops already audited; the four real
  gaps — `employee.create`, `membership.add`, `skills.assign`,
  `skills.require` — gained operator parameters, atomic transactions
  (wfm ops were auto-commit), and audit rows. Twelve pins lock the
  pre-existing coverage (terms/status/SHG/publish/revise/
  generate/submit/reissue/run-create/exclusion/disposition/correct).
- SHG has no verified flag in the model (§8 non-goal); the pin asserts
  the recorded source declaration instead.
- C8 moves two mutations into audited functions: `terms.close` rows
  and function-written `employee.status` rows (payload shapes match
  the prior Go audits). Failed closes/transitions leave no audit rows
  (asserted). Concurrency: advisory + row locks serialize correction,
  attach, import, reissue, approve, terms, status, publish and revise
  (9 race tests, exactly-once assertions).

## 7. Tests actually executed

- Full suite `go test -count=1 ./...`: green, 24 packages (also green
  under `TZ=UTC`).
- `go test -race` on attendance, timesheet, payroll, hr, wfm: green.
- `task lint` (gofmt/vet/nodeps), `task build`, `git diff --check`:
  pass. `openapi.yaml` parity via contract tests: pass.
- New suites: 16 audit tests, 38 Stage-C integrity tests (§9B matrix
  + C9 catalog/privilege), 5 new + 4 pre-existing concurrency tests,
  import replay, 6 employer service tests, 2 onboarding (CLI+REST),
  D1 script + §9C chain on ephemeral clusters, 6 E5 reset tests
  (incl. scram password-server cycle), D2 role tests at helper and
  CLI level. Real PostgreSQL throughout; no test uses a superuser to
  claim application-role success (superuser sessions appear only
  where the assertion is about trigger behavior on owner paths or
  about provisioning itself, labeled as such).

## 8. Application-role SQL evidence

Every §9B matrix item is a committed negative/positive SQL test run
as `app` with persisted-state assertions (rejected write + row
re-read): currency override/overlap/flip, temp-table denial (42501)
and shadow redirection (0 shadow rows, public flip lands),
cross-employee/employer/self/cycle supersession, cross-period
positive, chain UPDATE re-pointing (3 negatives + same-chain
positive), line moves, sheet transplants (draft + decided + bundled
approve), unpublish/rollback/missing-revision/metadata/week
transplant, revise-shape positive, closed-term/comp/status/
membership/skill mutation (42501 + values intact), close/transition
function positives with audit payload assertions, created_at
normalization (2020→current year), and the 19-table privilege matrix
plus per-function catalog checks (owner, DEFINER, search_path,
qualification, EXECUTE).

## 9. Fresh-install and onboarding evidence

- D1 `tests/provisioning`: the actual `init-roles.sh` runs twice
  against an ephemeral PG16 (creation branch first, idempotent re-run
  with a changed password); roles/grants/ownership/extensions/TEMP
  revokes verified, default privileges proven functionally
  (migration_admin table → app full DML, app DDL denied), passwords
  byte-identical across runs.
- §9C chain: provision → preflight → migrate (exactly 00001–00010) →
  ownership/privilege verification → `employer.Create` as `app` →
  superuser runtime and app-migrate refusals on the fresh install.
- D3: `employer create --file` / `show` and `POST /api/v1/employers`
  (201/200+existed/409 + operator-required 4xx) against isolated
  fresh databases; concurrent creates admit one creator and one
  observer with a single row and audit.

## 10. CLI/REST parity evidence

Pre-existing `parity_test.go` (CLI-vs-REST comparisons) passes
unchanged; onboarding adds the employer flow to both adapters with
matching semantics (create/repeat-flags/conflict/validation).
`openapi.yaml` carries the three new routes with contract-test
enforcement (`TestServedSubsetOfSpec`, `TestSpecImplementedAllServed`).

## 11. Remaining deferred issues

- OT-001 (SG-MOM overtime) stays BLOCKED; no payroll calculations
  were added — readiness/conflict detection only.
- §8 non-goals untouched: SHG lifecycle/verify, run state machine,
  rotation, generic idempotency/reads/pagination, SSO, RLS, cloud,
  daemon Docker work, other P2/P3 beyond the E-stage five.
- Residual accepted risks (documented in design): a direct-SQL
  revise-shaped reopen skips only the audit row (publish revalidates,
  history stays frozen); membership/skill INSERT gap-fill is
  structurally impossible (no gaps, exclusion, no DELETE) but INSERT
  stays granted by design.

## 12. Unverified environmental limitations

- Docker Compose path not exercised (no daemon in sandbox):
  `init-roles.sh` is proven on ephemeral PG16, which is exactly what
  the compose init hook invokes.
- `serve` loopback smoke (bind + `/healthz`) not run as a process
  here; the refusal it would hit is covered at helper and CLI level,
  and `/healthz`/`/readyz` handlers have unit coverage.
- Ephemeral-cluster tests skip honestly if `initdb`/`pg_ctl` are
  absent (present in this environment; both ran).

## 13. Final PASS/FAIL decision

**PASS.** All 8 P1 defects, all 10 app-role bypass findings, all 5
Stage-E corrections, install/onboarding (D1–D3 per Amendments D–G),
and reset safety (E5) are implemented with real-PostgreSQL proof and
committed regression tests. Full suite green (also `TZ=UTC`), race
detector green on the five concurrency packages, lint/build/diff
clean, OT-001 BLOCKED, no payroll calculations, history uncommitted
for review. No updated structural audit reference is required: the
design-doc updates (`database-design.md`, README, env/tasks) are the
record.
