# Pre-Milestone-2 Adversarial Review (prompt14, section 11 synthesis)

- HEAD: `f43515883668c917edac31be49a7cf8334eca5c2` (branch `main`, all scopes agree)
- Method: review-only. No repo writes by any scope; pre-existing worktree changes preserved.
  Static inspection of every cited body plus live probes on disposable PostgreSQL 16.15
  clusters (fresh `initdb`, all 11 migrations, real `app` role) and DB-gated Go suites.
- Sources: Scope 1 remap (S1-*), Area A disposition/approval (P14-A-*), Area B SHG
  lifecycle (SHG-*), Area C runs/snapshot (C*), Scope 5 areas D+E+F (S5-*), critic gap list.
- Severity: P0 critical / P1 major / P2 moderate / P3 minor. The M2 gate is the
  `blocks-M2` flag, not severity alone. Original IDs and severities are preserved;
  duplicates are folded under one canonical ID with cross-references.

## 1. Baseline + scope

- Baseline record (S1-BASELINE, P3, blocks-no): `git status --short` = 100 lines
  (18 modified, 4 deleted, 35 renamed `docs/*` -> `docs/parked/*`, 43 untracked);
  `sha256sum -c`: migrations 00001-00010 byte-identical, 00011 (1321 lines) the sole
  new migration. Toolchain: go1.27.3, psql 16.14, task 3.48.0.
- Scope covered: (S1) all 11 M1.3 remediation claims relocated to exact hunks,
  migration objects and regression tests, none unlocatable; (A) direct-disposition vs
  approval interplay, both interleaving orders, 5 timings; (B) SHG 8-state lifecycle;
  (C) run lifecycle, input snapshot, reproducibility; (D) financial/statutory
  boundaries; (E) 10-row source-to-lock matrix; (F) environment closure.
- Reconciliations applied by this synthesis (critic gaps R1-R4):
  - R1: S1-F2 said blocks-yes *pending Area A interleavings*; Area A now proves
    no-barrier-needed, so S1-F2 is SUPERSEDED for sync (blocks-no). Residual
    hardening survives as P14-A-F1/F2 (P2, blocks-no).
  - R2: SHG-B2a (P0, blocks-yes) governs the SHG barrier/detector gap; S5-F3
    (blocks-no) is read consistently as "no widening before a snapshot exists to
    protect" — the snapshot prereq (§13-P2) is the vehicle that makes B2a's fix
    attach. C4-SHG-CHANGE-INVISIBLE folds into SHG-B2a.
  - R3: duplicate findings folded: S1-F1 into SHG-B1a/B1b; S1-F3 into C1/C2/C6;
    S5-F2 into SHG-B1a (Area B owns); S5-F4 into C1 (Area C owns); C4-PACK into
    S5-F1 (single canonical pack finding).
  - R4: matrix row 8 (disposition) is CLOSED by the Area A verdict (§6).

## 2. Disposition / approval findings

**A4 verdict (one-of-three): `no-barrier-needed (proven)`.** Direct-SQL dispositions
do not need to join the employee payroll-input barrier: both interleaving orders were
proven live with zero deadlock and zero silent change to approved state.

- Direct disposition 9-effect summary (Area A, live as `app@p14a2`): (1) raw events
  byte-identical; (2) dup excluded from pairing views immediately, canonical
  untouched; (3) dup exclusion immediate in live views; (4) persisted exceptions stay
  open until a subsequent service pairing run; (5) no interval creation (direct), 0
  new when canonical consumed (service T4); service re-pair interval writes go
  through payroll lock + basis_seq + late-break gate; (6) unpaired list live-accurate,
  day/exception rows stale until re-pair; (7) direct disposition alone does NOT
  unblock generation (day-flag gate stays, proven S1); (8) approval `transition()` is
  a status flip under payroll lock with no evidence re-read, unaffected by
  dispositions in either order; (9) readiness: disposition can only REMOVE
  pending-event conflicts, never add — sound iff canonical consumed (T3), cannot mask
  genuine evidence (T6: pending canonical still conflicts).
- Timings: pre-generation/post-intervals = normal unblock path (needs service re-pair
  to complete); post-generation/post-submit = no draft/submitted change (no regen
  except late-break splits); post-approval = conflict-clear only for
  exact-duplicate-of-consumed; approved hours/lines/links immutable (freeze triggers
  + insert-only history).
- Mechanism: disposition writes only its own insert-only table (no `basis_seq`
  column); approved lines/links frozen by 00009 triggers for ALL writers; exact-match
  + chain-prohibition guarantees every duplicate set keeps an undispositioned
  endpoint carrying any genuine-evidence conflict.
- Lock order single and acyclic everywhere: payroll-input lock strictly before
  pair lock; direct SQL takes pair only, approval takes payroll only; multi-employee
  sorted (`internal/attendance/attendance.go:158-167`). Employee identity derived
  from immutable `clock_event` rows (trigger 00011:943-956, service
  `disposition.go:81-99`); `clock_event` is insert-only for `app`.
- CLI (`cmd/workforceops/attendance.go:91`) and REST (`internal/httpapi/httpapi.go:574`)
  call identical `DispositionDuplicate` with identical regen/chain wiring — consistent
  by construction (REST by inspection only; no live HTTP — see §11-R4).

### Findings

- ID: S1-F2 (SUPERSEDED for sync) / Severity: P1 (was) / Location:
  `migrations/core/00011_m13_integrity.sql:937-980` (pair lock only, no barrier, no
  consumed guard) vs `internal/attendance/disposition.go:99,104-114` / Scenario:
  direct-SQL disposition commits off-barrier and could disposition a consumed event /
  Evidence: trigger/service bodies fully read; Area A S1 (disp-wins: A held ONLY pair
  keys 789701469/229825141, B submit+approve in 0s during 15s hold) and S2
  (approval-wins: B held ONLY pay-input key 1294447019, A committed 2 dispositions in
  0s during 12s hold); approved REG 40.00 + 10 links + 15 intervals frozen, seq
  chains intact (S1 seq143/S2 seq146) / blocks-M2: **no** (verdict proven) /
  Minimum fix: none for sync; hardening in P14-A-F1/F2.
- ID: P14-A-F1 / Severity: P2 / Location: 00011:937-980 (no consumed check) vs
  `disposition.go:104-114` (service consumed guard) / Scenario: app-role direct SQL
  dispositions a CONSUMED event as duplicate of a pending event; trigger allows,
  service rejects. Interval citation persists (no payroll change) but the pending
  canonical becomes permanently unsettleable via `app` (forward edge fails
  "canonical is itself dispositioned", third edge fails "already canonical") —
  permanent generation/readiness block repairable only by `migration_admin`
  (insert-only for `app`) / Evidence: LIVE T5 as app: CLI backwards failed
  invalid_state, direct backwards INSERT `0 1`, CLI/direct forward failed,
  third-copy direct failed P0001; cited=1/disp=1 vs cited=0/disp=0 persisted /
  blocks-M2: no / Minimum fix: add consumed guard to
  `check_clock_event_disposition` mirroring the service.
- ID: P14-A-F2 / Severity: P2 / Location: `disposition.go:174-183` (service
  `audit.Write`) vs no DB audit on `clock_event_disposition` / Scenario: direct-SQL
  dispositions change readiness interpretation (T3 cleared a live pending conflict)
  with zero audit rows — a readiness-affecting operator decision invisible in
  history. Bounded: same systemic gap for direct clock inserts; no silent hours
  change / Evidence: LIVE counts as app: 10+ direct rows, 2 audits (both T4 service
  path) / blocks-M2: no / Minimum fix: DB-side audit insert on disposition
  (mirroring `transition_employee_status`) or declare disposition service-only and
  revoke app INSERT; minimum a recorded decision + test.

## 3. SHG findings

**B5 verdict: INSUFFICIENT as-is; READY WITH PREREQUISITES** (bounded 5-item set,
no redesign, no calcs — see §13-P4).

B4 live 8-state matrix (disposable PG16.15, all 11 migrations, real `app` role):
missing -> presence false, fail-closed free text; pending -> unrepresentable
(service 422 + DB CHECK reject); applicable-verified -> `active` satisfies but
nothing verifies; verified-none -> service path always CHECK-violates (dead), direct
SQL works but treated identically to `active`; conflicting historical -> same-fund
overlap rejected (exclusion, live-proven), cross-fund multiples accepted with no
conflict notion; retroactive change -> silent (caller-set `created_at`, no detector);
supersession -> no mechanism (overlap rejected; only destructive UPDATE/DELETE);
post-approval change -> silent (UPDATE/DELETE succeed, 0 audit rows, no conflict).

### Findings

- ID: SHG-B1a (canonical; absorbs S1-F1 service half, S5-F2) / Severity: P1 /
  Location: `internal/hr/shg.go:117-125` + 00002:160-163 CHECK / Scenario:
  `verified_none` unwritable via service: INSERT omits `verified_at/by/basis` which
  the CHECK mandates; zero tests cover `verified_none`; verified-none employees can
  never become readiness-complete except by recording a false `active`/`opt_out` row
  / Evidence: LIVE P1/P6 as app: service-identical INSERT -> CHECK violation;
  direct SQL with evidence cols succeeds / blocks-M2: **yes** / Minimum fix:
  service path recording verified-none with evidence+actor+date+provenance (extend
  `SetSHG` or add verify call) + test.
- ID: SHG-B1b (absorbs S1-F1 readiness half) / Severity: P1 / Location:
  `internal/payroll/readiness.go:339-350` / Scenario: readiness is status-blind —
  any covering row satisfies, so `opt_out` and unverified declarations pass
  identically to applicable instructions; absent-vs-verified-none collapses to
  row-exists vs free text / Evidence: LIVE P2b/P3 verbatim checkEmployee SQL:
  presence=t with ONLY verified_none row, t with ONLY opt_out row, f when missing /
  blocks-M2: **yes** / Minimum fix: status-aware SHG readiness
  (pending/unknown -> blocker; verified_none+evidence -> complete; defined opt_out
  semantics) with machine-readable blocker code.
- ID: SHG-B1c / Severity: P1 / Location: `shg.go:62-66` + 00002:148 + `shg_test.go:74`
  / Scenario: `pending` unrepresentable (service rejects, DB CHECK allows only
  active/opt_out/verified_none), so every recorded row immediately counts as
  complete input / Evidence: LIVE P8: pending INSERT -> CHECK violation / blocks-M2:
  **yes** / Minimum fix: staged/pending status excluded from readiness-complete, or
  explicitly restrict the first scenario to pre-verified instructions (documented).
- ID: SHG-B2a (canonical; absorbs C4-SHG-CHANGE-INVISIBLE; S5-F3 read via R2) /
  Severity: P0 / Location: `shg_instruction` privs/triggers (none in
  00002/00004/00009-00011) + `readiness.go:183-200` (no SHG conflict leg) / Scenario:
  post-approval direct-SQL UPDATE/DELETE as app silently rewrites/destroys the
  approved SHG basis: app holds full DML (pinned `privilege_test.go:301`), 0
  triggers, no `basis_seq`, no `created_at` normalize, no detector, no audit /
  Evidence: LIVE as app: P4 post-approval UPDATE succeeded; P5 DELETE flipped
  presence t->f with 0 audit rows; P11a/b/c: 0 triggers, no basis_seq, upd=t/del=t;
  Area C E1: new-fund SHG post-approval -> Included, 0 conflicts (control E3 comp
  correction DID flag) / blocks-M2: **yes** / Minimum fix: SHG post-approval
  conflict detector mirroring `checkTermConflict` + revoke app UPDATE/DELETE
  (insert-only) + `created_at` normalize trigger + `basis_seq` assignment.
- ID: SHG-B2b / Severity: P1 / Location: `shg.go:114` (service lock) vs 0 DB
  triggers / Scenario: direct-SQL SHG INSERT bypasses the payroll-input barrier;
  `TestSHGVsApprovalBarrier` scenario 1 is vacuous for the direct path (harness
  acquires the lock manually, `barrier_race_test.go:512`) / Evidence: LIVE 2-conn
  interleaving: B acquired the barrier in 0.03s while A held an uncommitted direct
  SHG INSERT / blocks-M2: **yes** / Minimum fix: subsumed by SHG-B2a fix
  (insert-only + detector makes the race detectable); alternatively a
  lock-demanding trigger.
- ID: SHG-B2c / Severity: P1 / Location: `shg.go:44-47` comment vs `:93-105` code
  (clash -> 409, no close); no SHG close/supersede symbol repo-wide / Scenario:
  doc comment claims supersession closes the predecessor, code only rejects
  overlap; unlike terms (`close_employment_terms`) and comp (`correct_compensation`)
  there is no governed correction path, forcing corrections onto destructive
  UPDATE/DELETE (SHG-B2a) / Evidence: LIVE P7 same-fund overlap -> exclusion
  violation (backstop works); inspection: no supersede function / blocks-M2: **yes**
  / Minimum fix: governed supersede/close (new row + audited predecessor close) or
  explicit freeze-plus-reissue rule for approved periods.
- ID: SHG-B2d / Severity: P2 / Location: `shg.go:48-137` (no terms lookup), 0
  triggers / Scenario: SHG dates need not cohere with employment (2020-dated row
  accepted for 2026 hire); caller-controlled `created_at` stored verbatim /
  Evidence: LIVE P9/P6 as app / blocks-M2: no (B2a fix makes backdates conflict
  loudly) / Minimum fix: `created_at` normalize rides with B2a; coherence check or
  documented deferral.
- ID: SHG-B4 / Severity: P2 / Location: `readiness.go:34-38,349` / Scenario: SHG
  blocker is free text (`no SHG instruction for period`), no code field, no
  distinct blockers per state / Evidence: inspection + live presence-shape
  cross-check / blocks-M2: no (covered for the first path by the B1b prereq) /
  Minimum fix: machine-readable blocker codes per SHG state.

## 4. Run-lifecycle assessment

Runs cannot be governed as-is: `CreateRun`/`AddExclusion` are the only service
operations, and status is free-form app-writable text guarded only after `approved`.

- ID: C1-RUN-LIFECYCLE-UNENFORCED (canonical; absorbs S1-F3 lifecycle half, S5-F4)
  / Severity: P0 / Location: 00002:166-189 CHECK + `internal/payroll/runs.go:16-133`
  (no validate/approve/finalize) + `00004_protection.sql:120-148`
  (`forbid_if_run_approved`) / Scenario: app SQL approves a run draft->approved
  skipping validated without Assess; validated->draft reopen allowed; only
  approved->* frozen. No calculated/under-review/finalized/superseded states;
  correction runs approve multiply with no supersede link;
  `payroll_correction.status` unconstrained / Evidence: LIVE as app: draft->approved
  ok; validated->draft reopen ok; approved->draft rejected P0001; 2nd approved
  regular rejected 23505; two approved correction runs both ok. `go test
  ./internal/payroll/` ok 2.4s / blocks-M2: **yes** / Minimum fix: Validate/Approve/
  Finalize service (assess-under-lock -> snapshot -> validated; all-or-nothing from
  snapshot -> approved) + DB transition-guard trigger (allowed edges only;
  validated/approved require a snapshot row).
- Pack activation (C4-PACK-GATING-BYPASSABLE-AND-UNPINNED / S5-F1, P0, blocks-yes):
  filed canonically in §7; it is also a run-lifecycle input-governance gap (runs
  record no pack pins, `rule_versions` always `'{}'`).

## 5. Snapshot / reproducibility assessment

**C7 verdict (one-of-three): `insufficient (proven)`.** Runs cannot freeze a
coherent reproducible input set: draft-only, placeholder hash, no version capture,
no snapshot table, Assess re-reads current rows on every call.

- ID: C2-NO-INPUT-SNAPSHOT (absorbs S1-F3 snapshot half) / Severity: P0 / Location:
  00002:172-175 placeholders written by `runs.go:50-60`; no snapshot table
  repo-wide; `readiness.go:102-242` reads current rows / Scenario: any source change
  after a run decision silently moves the calculation basis; two regular runs same
  period share identical `inputs_hash` (employer+period+kind only, not content-bound)
  / Evidence: LIVE: D1 placeholders observed; D2 identical hash true; E4 pack flip
  changes Assess verdict with no run binding; H1 Assess during open approval tx sees
  pre-commit, post-commit sees new state / blocks-M2: **yes** / Minimum fix:
  immutable per-run snapshot (run_id -> per-employee pinned ids + basis_seq: sheet,
  terms, live comp lines, SHG rows, pack pins, calc version) written atomically at
  validate; freeze trigger off-draft.
- ID: C3-NO-MULTI-EMPLOYEE-SNAPSHOT-BARRIER / Severity: P1 / Location:
  `readiness.go:102-242` (pool-direct, no tx, read-committed); `lock.go:28`
  (single-employee advisory xact lock); 00011:43-49 (seq comparable WITHIN one
  employee only); only multi-lock caller is clock import (sorted,
  `attendance.go:158-167`) / Scenario: Assess interleaved with approval/correction/
  termination yields a mixed-version view; cross-employee `pay_basis_seq` comparison
  is meaningless, so no run-wide cut exists / Evidence: code inspection + H1 probe /
  blocks-M2: **yes** / Minimum fix: snapshot capture in one tx (lock expected
  employees UUID-sorted, revalidate population, capture basis refs, insert snapshot,
  commit); document seq as per-employee order only.
- ID: C4-STATUS-CHANGE-INVISIBLE / Severity: P1 / Location: Assess never reads
  `employee.status`/history; status rows carry no `basis_seq` by design
  (00011:221-228); `hr.go:185` takes barrier but no population-conflict check /
  Scenario: post-approval termination/suspension invisible — employee stays Included,
  verdict unchanged / Evidence: LIVE: A0 terminated-E004/suspended-E005 both
  Included; E2 post-approval termination -> unchanged / blocks-M2: **yes** /
  Minimum fix: snapshot population compare at approve (status/terms-effective drift
  -> explicit invalidation), or a status post-approval conflict check.
- ID: C6-REPRODUCIBILITY-UNSATISFIABLE / Severity: P0 / Location:
  `EffectiveApproved` resolves CURRENT chain head (`timesheet.go:588-654`);
  terms/comp/SHG/packs all current-row reads; `calc_version` placeholder; NO
  effective-compensation resolver (`ListCompensation` returns original+successor,
  `hr.go:591-611`; readiness only conflict-flags) / Scenario: same nominal inputs
  reproduce nothing — chain head, live comp basis, SHG set, pack versions, calc
  version all mutable and unpinned; a snapshot cannot name "the" comp line without
  a resolver / Evidence: LIVE G1: two live rows same code+range (100.00 original +
  120.00 successor) with no single-effective API; F1 repeat-Assess deterministic
  only because sequential current-reads / blocks-M2: **yes** / Minimum fix:
  live-basis comp resolver (per employee+code+period, superseded-excluded) used by
  snapshot capture; pin all six input groups per C2.
- (C4-SHG-CHANGE-INVISIBLE folds into SHG-B2a, §3. C5-POPULATION-LOGIC is CLEAN —
  expected-population union, all-or-nothing, disclosed run-scoped exclusions,
  double empty-guard — see §12.)

## 6. Lock-coverage matrix (10 rows + SHG row; layers R/I/T/C/P/S)

Row 8 is now CLOSED by the Area A verdict (R4). "direct" = app-role direct SQL.

| # | Mutation | Service path | Direct-SQL path | basis_seq | Audit | Readiness detector | Layers |
|---|----------|--------------|-----------------|-----------|-------|--------------------|--------|
| 1 | comp-add | locking audited function (`hr.go:516-552` -> 00011:734) | same function only (INSERT revoked) | trigger (00011:120-122) | canonical in-function (00011:740-745) | post-approval add -> CompConflict (`readiness.go:451-484`) | C |
| 2 | comp-correct | locking audited function (`hr.go:557-588` -> 00011:776) | same function only | trigger | before/after in-function (00011:801-808) | flags "(correction; reissue required)" until reissue | C |
| 3 | term-change | lock + retro-block vs approved periods (`hr.go:366-410`) | INSERT allowed, trigger locks+seqs (00011:123-125), NO retro-block | yes | svc yes; direct no | dual-leg detector catches direct path (`readiness.go:360-387`; sc3 proves basis_seq leg) | C+P |
| 4 | status-change | lock (`hr.go:185`) + function (00011:248) | UPDATE/DELETE revoked, function only | none by design (population recomputed per Assess) | in-function (00011:277-280) | expected-population re-derived; no stale pointer (but see C4-STATUS for post-approval drift) | P |
| 5 | clock-import | sorted employee locks (`attendance.go:163-167`) | trigger locks+seqs (00011:129-131) | yes | svc yes; direct no | unpaired post-approval event -> AttendanceConflict (`readiness.go:425-443`) | R (+I) |
| 6 | break-attach | lock (`attach.go:60`); pairing op only | can only insert break events == row 5 | via event/interval triggers | svc yes | post-approval intervals -> conflict (`readiness.go:407-420`); decided-chain hold in linkage gate | I |
| 7 | attendance-correction | lock (`review.go:223`) | group+segment inserts; segment trigger locks+seqs; group open/complete triggers (00011:329-466) | yes (segments) | svc yes; direct no | post-approval segments -> conflict (3 scenarios green) | I |
| 8 | disposition | employee lock (`disposition.go:99`) + consumed/canonical guards | PAIR-LOCK ONLY, no employee lock (probe P8), no basis_seq | none | svc yes; direct no | no disposition check needed — CLOSED: Area A proves removal-only, sound-iff-consumed, genuine evidence always still blocks (T3/T6, S1/S2) | R/I |
| 9 | timesheet-approval | lock + state machine (`timesheet.go:744-786`) | trigger locks + assigns seq on ->approved (00011:139-166); decided sheets frozen to ALL writers (00004/`trg_timesheet_frozen`) | ASSIGNED at approval = freeze point | svc yes; direct no | freeze point for all conflict legs | T |
| 10 | exclusion | FOR UPDATE + draft-only (`runs.go:93-105`) | allowed, draft-only trigger (00004:156-185) | none | svc yes; direct no | re-read per Assess (`readiness.go:271-309`); empty/exclusion validation (`approval.go:85-139`); pre-snapshot nothing frozen | P (run) |
| 11 | SHG-set | lock (`shg.go:114`) | NO lock, full DML, no seq | none | svc yes; direct no | presence-only, no post-approval detector — see SHG-B2a (P0, blocks-yes) | S |

## 7. Financial / statutory-boundary findings

- ID: S5-F1 (canonical; absorbs C4-PACK-GATING-BYPASSABLE-AND-UNPINNED) /
  Severity: P0 / Location: `rule_pack_activation` (00003:81-88) + no activation
  service (only seed BLOCKED `seed.go:341-347`; GET-only route `httpapi.go:761-789`;
  reads `readiness.go:252-269`, `approval.go:85-139`) / Scenario: app-role direct SQL
  flips any pack BLOCKED->IMPLEMENTABLE with an invented version, no audit, no
  manifest check; readiness then reports ready and M2's first calc would run on
  unverified rules; no Go code reads `rules/*/manifest.json` / Evidence: LIVE P5/P2
  as app: UPDATE to `(SG-CPF-2026,invented.IMPLEMENTABLE)` succeeded; app holds full
  DML (pinned `privilege_test.go:298`); E4 SUPERSEDED+version bump flips verdicts /
  blocks-M2: **yes** / Minimum fix: controlled audited activation function
  (actor+reason+version pin, BLOCKED->IMPLEMENTABLE only with review evidence) +
  revoke app INSERT/UPDATE on `rule_pack_activation` and param tables
  (`mom_ot_params`, `cpf_*`, `sdl_params`, `shg_contribution`); pin (pack,version,
  status) per run at snapshot; migrate tests to the function.
- CLEAN (reconfirmed §D): no float64 money end-to-end (string->NUMERIC:
  `validateMoney` `hr.go:292-309`, `::numeric` params, exact-scale CHECKs
  00011:580-589; zero ParseFloat/FormatFloat/Float64 repo-wide; float64 confined to
  hours display and a `>0` hours check — time, not money); excess precision
  rejected-not-rounded at both layers (live 23514s, `TestCompMoneyPrecision`/
  `TestTermsMoneyPrecision` PASS, exact 2600.50/199.90 persistence); single
  effective comp basis per code (`UNIQUE(supersedes_id)`, live-overlap trigger,
  successor contract; original immutable, P4); no hidden OT/CPF/SDL/SHG/net calc
  (`MonthlyExcess` is an hours review signal that never gates generation);
  SG-MOM-OT-001 BLOCKED with empty golden dirs; readiness unbypassable
  (empty-population + exclusion validation, no skip flags in config/CLI/REST);
  restricted profile representable (free-TEXT citizenship/salary_period, terms
  range, monthly_basic, REG-only sheet; no inference needed).
- Observations, not findings: hours columns NUMERIC fed by float64 Go inputs with
  no precision validation (silent 2dp rounding possible on hours, never money);
  citizenship/salary_period unenforced TEXT; readiness sums NUMERIC hours into
  float64 for a `>0` check.

## 8. Environment findings

- ID: S5-F5 / Severity: P3 / Location: `.env.example` + `Taskfile.yml` vs
  `cmd/workforceops/main.go:21-41`, `serve.go:15`, `core/config/config.go:42-52` /
  Scenario: stale env names (real var is `A1_API_PORT`) and Taskfile tasks for
  nonexistent commands mislead M2 setup; unused LLM vars imply unbuilt deps /
  Evidence: grep — OPENROUTER/LLM_MODEL/A1_SHIFTHUB/A1_PAYLOOP unreferenced except
  a Port-unit test / blocks-M2: no / Minimum fix: sync `.env.example`, drop unused
  vars, remove/park dead tasks.
- ID: S5-F6 / Severity: P3 / Location: `docker-compose.yml` + `Dockerfile` /
  Scenario: compose stack + serve-loopback smoke not executed (docker socket
  permission-denied in sandbox); no compose verification claimed / Evidence:
  inspected only — compose pins `postgres:16-alpine` + `init-roles.sh` mount;
  Dockerfile builds embedded-migration binary with `CMD serve --port 8081` /
  blocks-M2: no / Minimum fix: run compose + serve smoke where docker exists; never
  gates local M2.
- CLEAN (reconfirmed §F): app vs `migration_admin` separation enforced
  (`RefusePrivilegedRuntime` wired into CLI `cli.go:63` and serve `host.go:50-53`;
  migrate preflight requires `migration_admin`); CLI/REST env parity (both
  `DATABASE_URL` + refusal; migrate `MIGRATION_DATABASE_URL`; serve loopback-only
  `host.go:61`, `A1_API_PORT` override); no readiness-bypass config; migrations
  embedded in binary; M2 local workflow reproduced on disposable PG16.15.
- Test-env notes (recorded, not findings): no `go.work` in repo; parent
  `/home/chaschel/Documents/go/go.work` captures the module => `GOWORK=off`
  required; `GOCACHE=/tmp/p14d5/gocache` (home cache read-only in sandbox).

## 9. Tests executed

Exact record; all green unless named otherwise. Default TZ only (3-TZ battery not
re-run — see §11-R7).

- (S5-a) `GOWORK=off GOCACHE=/tmp/p14d5/gocache WORKFORCEOPS_TEST_DATABASE_URL=
  postgresql://a1@127.0.0.1:55436/a1_test go test -count=1 -v -run
  'TestCompMoneyPrecision|TestTermsMoneyPrecision' ./tests/integrity/` -> PASS
  (proves DB-gated, not SKIP).
- (S5-b) Same env, `go test -count=1 ./tests/arch/ ./tests/integrity/ ./tests/e2e/
  ./tests/provisioning/ ./internal/hr/ ./internal/payroll/ ./internal/timesheet/
  ./internal/attendance/ ./core/database/ ./core/config/ ./cmd/workforceops/` ->
  all 11 packages ok (includes barrier-race, group-matrix, disposition-cycle,
  privilege-matrix and correction-barrier suites).
- (S5-c) Fresh probe DB + roles, `MIGRATION_DATABASE_URL=... go run
  ./cmd/workforceops migrate up` -> applied 11 migration(s).
- (S5-d) Probes P1-P9 as `app`: P1 exact numerics; P2a/P2b 23514 CHECK rejections;
  P4 succession seq4<seq5<seq6 with original preserved; P5 activation flip; P6
  verified_none CHECK behavior; P7 19-row trigger matrix + disp lock flags (f,t) +
  5 basis_seq tables + 0 SHG triggers; P9 run flip. Logs under `/tmp/p14d5/`.
- (Area A) as `app@p14a2`, disposable PG16.15, 11 migrations: S1/S2 two-connection
  interleavings (15s/12s holds, zero deadlock); T3 conflict-clear soundness; T4
  service post-approval disposition (0 intervals, 2 audits); T5 consumed-event
  trap; T6 dup-of-pending still blocks.
- (Area B) 19 tests green on disposable PG16.15 as designed roles (5 HR SHG incl
  `TestSHGVsApprovalBarrier` + 14 `TestAssess*`); live probes P1-P9/P11
  (`/tmp/shgrev/run.log`).
- (Area C) `go test ./internal/payroll/` ok 2.4s; live probes A0/B1/B3/D1/D2/
  E1-E4/G1/H1/P2 as app.
- (S1 census, static): 335 top-level Test funcs + ~94 subtests reconcile the
  report's "429 PASS lines" (S1-N1: relabel, blocks-no); M1.3 new-test counts
  verified per file (80 as claimed); S1-N2 trivial doc arithmetic 22->23
  (blocks-no); `go test -list` shows 211 without DB (integrity+e2e exit 0 when
  skipped).
- NOT executed: 3-TZ battery (S1-N1 open item), docker/serve smoke (S5-F6),
  loaded true-concurrency Assess race (only sequential H1 probed).

## 10. Confirmed defects (blocks-M2 = yes)

| ID | Sev | Defect | Canonical location |
|----|-----|--------|--------------------|
| SHG-B1a | P1 | verified_none unwritable via service | `shg.go:117-125`, 00002:160-163 |
| SHG-B1b | P1 | readiness status-blind to SHG state | `readiness.go:339-350` |
| SHG-B1c | P1 | pending/unverified state unrepresentable | `shg.go:62-66`, 00002:148 |
| SHG-B2a | P0 | post-approval SHG change silent; full DML, no detector/seq/audit | `shg_instruction` (no triggers), `readiness.go:183-200` |
| SHG-B2b | P1 | direct SHG INSERT bypasses barrier (subsumed by B2a fix) | `shg.go:114` vs 0 triggers |
| SHG-B2c | P1 | no governed SHG supersede/close | `shg.go:93-105` |
| C1 | P0 | run lifecycle unenforced (direct approve/reopen, no states) | `runs.go:16-133`, 00002:166-189 |
| C2 | P0 | no per-run input snapshot; placeholder hash | `runs.go:50-60`, no snapshot table |
| C3 | P1 | no multi-employee snapshot barrier; Assess non-atomic | `readiness.go:102-242`, `lock.go:28` |
| C4-STATUS | P1 | post-approval status change invisible | Assess vs `hr.go:185`, 00011:221-228 |
| C6 | P0 | reproducibility unsatisfiable; no comp resolver | `timesheet.go:588-654`, `hr.go:591-611` |
| S5-F1 | P0 | pack activation bypassable + unpinned | 00003:81-88, no activation service |

Folded duplicates (not double-counted): S1-F1 -> SHG-B1a/B1b; S1-F3 -> C1/C2/C6;
S5-F2 -> SHG-B1a; S5-F3/C4-SHG -> SHG-B2a; S5-F4 -> C1; C4-PACK -> S5-F1.
S1-F2 superseded for sync (Area A proof); residuals are P14-A-F1/F2 (P2, non-blocking).

## 11. Risks / hypotheses (open, not defects)

- R1 cross-fund SHG multiples (e.g. CDAC active + SINDA active, one employee/period):
  model permits, readiness accepts; no in-repo spec. Needs M2 rule research. Blocks
  a "sufficient" determination for multi-fund scenarios only.
- R2 Assess-vs-approval race: only sequential H1 interleaving probed; no loaded
  measurement. Code-level non-atomicity is already conclusive (C3); measurement
  would quantify, not decide.
- R3 SHG Go-path per-state proof missing: B4 probes used verbatim-copied SQL, not
  live Go Assess per state (new harness forbidden in review-only scope); 14
  `TestAssess*` cover the Go path on the same schema.
- R4 REST disposition parity by construction only (no live HTTP); server smoke is
  Area F, unexecuted (S5-F6).
- R5 generation-commit-to-approval-commit invisibility: evidence imported between
  the two commits is invisible to readiness (predates approval; `transition()` does
  no eligibility recheck). Boundary note, owned by no area verdict; disposition
  neither causes nor widens it. Snapshot capture (§13-P2) supersedes this window.
- R6 residual lock-free Assess race (any input committing between Assess and run
  approval) is architecture-wide, not SHG-specific; closed by the snapshot prereq,
  not by per-table fixes.
- R7 3-TZ full-suite battery not re-run (S1-N1); S5 ran 11 packages default-TZ
  only. No TZ-sensitive failure is suspected; the claim is unreplicated.
- R8 `payroll_line.source_refs` trustworthiness under a future calculator: no
  calculator exists; CHECK constrains key presence only. M2 design concern.

## 12. Clean areas

- Disposition/approval sync (Area A CLEAN): no barrier needed, proven live both
  orders; approved state immutable to dispositions.
- Population logic (C5): terms-union expected set, all-or-nothing approval,
  disclosed run-scoped exclusions, double empty-guard.
- Money boundaries (S1-MAP-COMP, S5-CLEAN-D): exact NUMERIC end-to-end,
  reject-not-round, single-effective-basis constraints, original-immutable
  succession, concurrency single-winner tests green.
- Barrier core rows 1-7, 9-10 (§6): locking functions / lock+seq triggers /
  decided-freeze / draft-only gates all verified with green detector tests.
- OT-001 BLOCKED; no statutory calc code anywhere (grep-clean); money never
  float64 (S1-MAP-OTNOSTAT).
- Privilege matrix (S1-MAP-PRIV): 55-table classification, 30 function specs,
  catalog-completeness + no-PUBLIC + default-ACL pins, suite green.
- Provisioning (S1-MAP-PROV): clean-server path, role verification, FRESHFLAKE
  fix, parallel-race suite green.
- Correction groups (S1-MAP-GROUP): atomicity + DB lineage + 8-way completion
  check; 17 + 10 + 6 tests green (coverage-rule drop as reported, no union check).
- Disposition trigger/service contract as mapped (S1-MAP-DISP): concurrency/cycle
  handling green (6/6 cycle bodies + e2e parity).
- SHG clean subset (SHG-CLEAN): employer/employee binding, same-fund overlap
  exclusion live-proven with cross-fund coexistence + preserved history,
  service-path audit + barrier participation, IMPLEMENTABLE-only approval gate,
  data minimization.
- Environment (S5-CLEAN-F): role separation, CLI/REST parity, loopback-only
  serve, embedded migrations, reproduced local M2 workflow.

## 13. M2 prerequisites — first-payroll path only (bounded, no calcs, no redesign)

Complete all six before the first payroll calculation runs. Each maps to confirmed
defects; nothing here is open-ended.

- P1 Governed run lifecycle (C1): Validate/Approve/Finalize service
  (assess-under-lock -> snapshot -> validated; all-or-nothing from snapshot ->
  approved) + DB transition-guard trigger (allowed edges only; validated/approved
  require a snapshot row). Closes direct approve/reopen and multi-approval.
- P2 Immutable per-run input snapshot (C2, C3, C4-STATUS): one-tx capture
  (UUID-sorted employee locks, revalidate population, pin per-employee ids +
  basis_seq for sheet/terms/comp/SHG + pack pins + calc version); freeze trigger
  off-draft; approve compares snapshot-vs-current population (status/terms drift ->
  explicit invalidation). Document seq as per-employee order only. Closes R5/R6
  windows by construction.
- P3 Live-basis comp resolver (C6): per employee+code+period single-effective API
  (superseded-excluded), used by snapshot capture; deterministic under existing
  constraints. Pin all six input groups per C2.
- P4 SHG first-path lifecycle (SHG-B1a/B1b/B1c/B2a/B2b/B2c): (a) recordable
  verified-none with evidence+actor+date+provenance via service; (b) status-aware
  readiness + machine-readable SHG blocker codes; (c) post-approval conflict
  detector + insert-only app privs + `created_at` normalize + `basis_seq`;
  (d) governed supersede/close or explicit freeze-plus-reissue for approved
  periods; (e) pending state or documented pre-verified-only scenario restriction.
- P5 Controlled rule-pack activation + pins (S5-F1): audited SECURITY DEFINER
  activation (actor+reason+version pin, gated transition) + revoke app direct
  writes on activation/param tables + per-run (pack,version,status) pins at
  snapshot; migrate tests to the function.
- P6 Regression tests per prerequisite: verified-none round-trip, SHG detector
  legs, run-transition guards, snapshot-vs-current drift, resolver
  single-effective, activation control — each failing pre-fix, green post-fix.

## 14. Deferred concerns (explicitly NOT first-payroll prerequisites)

- P14-A-F1 (consumed guard in trigger), P14-A-F2 (disposition audit contract):
  harden after P1-P6; record the audit decision.
- SHG-B2d employment-coherence check (`created_at` normalize rides with P4c);
  SHG-B4 extended blocker taxonomy beyond the P4b codes.
- S5-F5 (`.env.example`/Taskfile hygiene), S5-F6 (compose + serve smoke where
  docker exists).
- S1-N1 "429 PASS lines" relabel, S1-N2 22->23 correction.
- Hours-float NUMERIC observation (hours only, never money); free-TEXT
  citizenship/salary_period validation; R1 cross-fund semantics (M2 rule
  research); R7 3-TZ replication; R8 source_refs trust (M2 design).

## 15. Overall verdict

**READY WITH PREREQUISITES — NOT READY as-is.**

Green suites alone authorize nothing: the tree is clean where Milestone 1 absolutized
(barrier core, money exactness, privilege/provisioning closure, disposition sync)
but the first payroll calculation has no governable run, no freezable input set, no
usable SHG lifecycle, and no trustworthy rule-pack gate. Those are the only
blockers, and all are bounded: no redesign, no calculations, no new architecture.

**Decisive transition:** complete the six prerequisites in §13 (lifecycle, snapshot,
resolver, SHG, activation, regression tests), in that dependency order as far as
practical (P3 before P2 capture; P4c/P5 before P2 pins). M2 work may begin with
P1-P6 as its first gated slices. No payroll calculation — draft or otherwise — is
authorized until all six are merged and green. Everything in §14 is explicitly out
of the gate.
