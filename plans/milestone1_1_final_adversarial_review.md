# Milestone 1.1 Final Adversarial Review

**Verdict: FAIL** — bounded, actionable. Seven service-reachable P1 defects threaten
payroll inputs or block the supported workflow, plus one P1 breaks fresh install.
No P0 found. See §16–§18.

- Review basis: `plans/prompt10.md` (§§1–17), `plans/milestone1_1_structural_audit.md`,
  `plans/milestone1_1_remediation_report.md`, migrations `00001`–`00009`, Go
  implementation, regression tests.
- Method: 8 independent evidence scopes (static + live disposable-PG16 probes) plus a
  completeness critic; this report synthesizes only that compact evidence, deduped.
- Review-only: no source, migration, test, or doc modified. This file is the sole new file.

---

## 1. Git and working-tree baseline

| Item | Value |
|---|---|
| HEAD | `f43515883668c917edac31be49a7cf8334eca5c2` |
| prompt10 §1 line 41 hash | matches HEAD exactly (`edac` variant correct; the `ed44` string in one delegated task text is a transcription error, unknown to git) |
| Staged | 3 deleted cmds + 35 docs renames to `parked/` + deleted `00001_schemas.sql` |
| Unstaged | 16 modified |
| Untracked | 32 entries incl `internal/`, `rules/`, `tests/integrity/`, migrations `00001_foundation`–`00009_m11_integrity` |
| `git diff --check` | clean |
| Go / psql / task / docker | 1.27.3 (report said 1.25) / 16.14 / 3.48.0 / 29.5.2 |
| Live probe DB | disposable PG 16.15 from user-space .deb (no daemon/sudo); per-call network namespaces forced pg start/stop inside each DB call; data persisted on disk |

M1-to-M1.1 byte-identical claim for `00001`–`00008` is **unverifiable from git**:
HEAD contains only `00001_schemas.sql` (now deleted); all `00001_foundation`–`00009`
files are untracked additions. No M1 snapshot exists in-repo to diff against.
`00009_m11_integrity.sql` is the sole later migration on disk. Report wording must
carry this caveat (§19).

## 2. Review scope and methodology

- Full read of structural audit + remediation report; line-level reads of `00009`,
  `00001`/`00002`/`00004`/`00006`/`00008`, attendance/timesheet/payroll/HR/WFM services,
  CLI + REST adapters, contracts, OpenAPI.
- Four verification outcomes per prompt10 §2: CONFIRMED FIXED / CONFIRMED DEFECT /
  NOT REPRODUCED / UNVERIFIED. Vacuity auditing: setup, role, SQL, persisted state,
  negative case checked per regression test (systematic pass incomplete — §19).
- Live execution (scope 8 + scopes 2/3): `task lint`/`build`, `go test -count=1 ./...`
  twice against disposable PG (run1 EXIT=1 on F-PROVISION-RACE, run2 EXIT=0),
  `-race` on attendance/timesheet/wfm/payroll/hr, fresh-install provision + 9
  migrations, full CLI E2E on seeded MERIDIAN, adversarial app-role SQL matrix
  (A1–A9), Go-harness real-transaction races (R1–R3), populated `00009` Down/Up,
  wrong-role MigrateUp, CLI preflight refusals, live REST E2E 19/19, seed reset cycle.
  Env note: parent `go.work` breaks bare `go`/`task lint`; all commands ran
  `GOWORK=off` + `GOCACHE=/tmp/gocache` (bare-`task-lint` failure recorded, §19).
- Threat model follows the project's own bar: the `app` role issuing direct SQL is a
  legitimate adversary (SYN-P1-1 was fixed at DB level on exactly this basis).
  Raw-SQL-only findings are real defects, scored P1/P2 with `blocks-first-payroll`
  reflecting service-reachability.

## 3. Actual commands executed

Executed live (disposable PG16.15, `GOWORK=off`, `NO_PROXY` loopback bypass):

- `task lint` (bare: FAIL on go.work; `GOWORK=off`: pass), `task build` (pass),
  `git diff --check` (clean).
- `go test -count=1 ./...` without DB: 59 skip / 0 fail. Full DB suite twice:
  run1 EXIT=1 (integrity 23505 role race = F-PROVISION-RACE), run2 EXIT=0 all pass.
- `-race` on attendance, timesheet, wfm, payroll, hr suites: all pass.
- Fresh install via real `scripts/db/init-roles.sh`: FAILED (F-INITROLES, mechanism
  isolated); then manual provision + `migrate up` (9 applied) + catalog verification
  (owners/grants/DEFINER/extensions/column-grants).
- Full CLI E2E on seeded MERIDIAN: review / generate 40+35h / submit / approve /
  readiness / run / exclude / reissue / correct / revise / status / SHG / comp / shift.
- App-role SQL matrix A1–A9, Go races R1–R3, populated-00009 Down (loud 23505,
  atomic, v9 retained, up no-op), wrong-role MigrateUp (v1–v8 half-apply, stuck),
  CLI preflight 3/3 refusals, REST E2E 19/19, seed reset cycle + refusals.
- Scope-2/3 probe scripts (`/tmp/s2step*.sh`, `/tmp/p10s3_probes_*.sql`, `/tmp/p9.py`)
  live outside the repo and are not attached to this report (§19 gap).
- NOT executed: `docker build` (no daemon), Compose volume inspection, pre-pivot
  upgrade probe, TZ=UTC pairing run, password-auth recycle failure, typo-target drop
  guard, kill-mid-import crash probes. Details in §19.

## 4. Migration 00009 and historical-schema review

`00009_m11_integrity.sql` read line by line against `00001`–`00008` final behavior.

- **Lineage/allocation triggers** (`:61–128`): pay-period non-overlap, single-chain
  guard, employee-match guard — INSERT-path verified live (lineage tests + A-series
  probes). **Gap**: single-chain guard is BEFORE INSERT only; app-role UPDATE
  re-pointing a link succeeds → silent double allocation (F-SINGLECHAIN-UPDATE, §14).
- **Currency/exclusion** (`:130–173`): `is_current` + DEFINER trigger + partial GiST
  exclusion + single-successor index — core paths verified live (§6). **Gaps**:
  explicit-`is_current=false` INSERT bypass (F-IS-CURRENT-BYPASS), cross-employee
  supersede victim-flip (F-XEMP-SUPERSEDE), pg_temp search_path exploit
  (F-PG-TEMP), self-supersede slot-poison (F-INTERVAL-SELFLOOP), all §14.
- **Chain uniqueness** (`:179–188`): partial original + successor indexes — verified
  live incl. true-concurrency single-winner. **Gap**: no employee/period/cycle scope
  guards; self/2-cycle/x-employee links succeed via app SQL (F-CHAIN-SCOPE, §14).
- **Freeze triggers** (`:193–359`): all five functions verified live on OLD+NEW
  parents both directions + child-transition combos (S3 D-series, A9/FK probes).
  **Residuals**: cross-employee line move draft→draft allowed (F-LINE-MOVE);
  draft sheet employee/period keys mutable incl. atomic-with-approval (F-SHEET-KEYS);
  roster unpublish + published-assignment edit allowed, no trigger on roster itself
  (F-ROSTER-REOPEN); scoped oracle persists (frozen vs 23503) as negligible P3
  (F-ORACLE, reconciled with Q18-clean — §13).
- **Grants** (`:42–58` + `00004`): 7 insert-only tables revoked; verified `app` cannot
  UPDATE/DELETE intervals (42501 ×4 variants). **Gap**: HR/WFM parents
  (terms/status-history/roster/membership/skills) hold full UPDATE/DELETE as app with
  no freeze/status trigger (F-HR-DIRECT-WRITE, §14).
- **Down** (`:396–419`): populated Down fails loudly (23505) and atomically — version
  stays 9, rows/constraints intact, re-Up no-op (S2 + S8 probes). **Gaps**: no
  pre-destruction guard (safety relies solely on goose tx-wrapping); limitation
  undocumented; clean-DB Down succeeds but silently destroys `attendance_exception`
  rows (F-ROLLBACK-DOCS, F-EXCEPTION-LOSS, §14).

## 5. Timesheet correction-chain review (prompt10 §3A, 14 checks)

| # | Check | Result |
|---|---|---|
| 1 | Approved original unchanged after reissue | CONFIRMED FIXED (freeze + live) |
| 2 | Rejected original unchanged after recovery | CONFIRMED FIXED (`TestRejectedRecoveryFlow` pins byte-identical 9h) |
| 3 | Correction = new draft, original untouched | CONFIRMED FIXED |
| 4 | Correction cannot belong to another employer/employee/period | **DEFECT**: cross-employer blocked by composite FK (live G1); cross-employee/period succeed via app SQL (F-CHAIN-SCOPE) |
| 5 | No concurrent multiple successors | CONFIRMED FIXED (SQL race count=1, CLI reissue race rc 0/1, approve race one winner) |
| 6 | No cycles | **DEFECT (service-unreachable)**: self-supersede + 2-cycles succeed via app SQL; reachable multi-node cycles proven impossible (single-successor + frozen decided rows); Reissue walk has no cycle guard but hang hypothesis DISPROVED |
| 7 | No branching | CONFIRMED FIXED via service (single-successor); raw-SQL second successor with `source='clock'` allowed as documented (S3-C5, no fix) |
| 8 | No invalid period-boundary crossing | **DEFECT**: same gap as #4 (F-CHAIN-SCOPE) |
| 9 | Approved correction becomes effective | CONFIRMED FIXED via service ordering |
| 10 | Later draft/rejected never displaces effective | CONFIRMED FIXED (C2/J probes) |
| 11 | Readiness cannot double-count original + correction | **GAP**: behavior live-verified (effective-only counting); zero Assess tests with corrections, no EffectiveApproved double-count service test (F-TEST-GAPS) |
| 12 | Reissue of approved correction sound | CONFIRMED FIXED (chain mechanics) — but see F-APPROVED-DEADLOCK: correction after reissue-of-approved is permanently rejected |
| 13 | History queryable + immutable | CONFIRMED FIXED |
| 14 | Audit per transition | **GAP**: live H3/I1 covers submit/approve/reissue only; zero `audit_event` assertions in suites; correct/reject/generate/exclude unpinned (F-TEST-GAPS) |

`EffectiveApproved` (`timesheet.go:566–582`): correct via service ordering
(latest-wins); fragile predicate — direct-approved-successor exclusion + `ORDER BY
approved_at DESC LIMIT 1` yields TWO qualifiers on approved→rejected→approved chains
(live J probe); tie/backdate mispick never forced live (F-EFFECTIVE-TIE, P3).

## 6. SECURITY DEFINER and is_current analysis (prompt10 §3C)

Catalog-verified on disposable PG16.15: function owner `migration_admin`, secdef,
`search_path=public`, single BEFORE INSERT trigger, no recursion; app UPDATE/DELETE
on `attendance_interval` → 42501; second correction successor → 23505; 2-cycle needs
UPDATE → 42501; overlapping live INSERT → 23P01; adjacent `[)` spans succeed; direct
DEFINER call → 0A000; GiST predicate `WHERE (is_current)` on
(employer,employee,range) correct; RACE1/RACE2 exactly one winner; Q4 UPDATE-escape
impossible; Q17 failed write rolls back co-transactional audit row.

Adversarial matrix (§3C items 1–10): items 1–4, 7, 9 hold; **item 5 FAILS**
(explicit `is_current=false` INSERT bypasses maintenance → F-IS-CURRENT-BYPASS);
**item 6 FAILS** (cross-employee supersede flips victim → F-XEMP-SUPERSEDE);
**item 8 PARTIAL** (self-supersede accepted → F-INTERVAL-SELFLOOP; 2-cycle blocked);
**item 10 FAILS** (pg_temp decoy → double-current → F-PG-TEMP). Writable-schema
vector clean (`app_can_create=f`, schema owned by `migration_admin`); PG<15
PUBLIC-CREATE divergence untested (§19).

## 7. Attendance and temporal correctness (prompt10 §4)

- **§4A pairing**: roster-anchored + 24h ceiling verified live (span_too_long,
  span_date_mismatch, compliance-only 72h). **Defects**: TZ-dependent
  misclassification under non-local process TZ (F-TZ-CLASSIFY, P1); shift
  start/end never consulted, rest-day rows valid anchors (F-OFF-WINDOW, P2);
  cross-source vendor-A-in + vendor-B-out pair silently (F-XSOURCE-PAIR, P3);
  multiple-shifts-per-day unpinned anywhere (§19).
- **§4B 72h**: compliance-only CONFIRMED (MonthlyExcess flag-only; buildRegLine sums
  all current work+break_paid uncapped; 210h test asserts full line + NeedsReview).
- **§4C cross-midnight**: import split verified (both dates cited, tz-safe casts,
  session-TZ safe SQL). **Defects**: Correct inserts ONE interval with
  local_date=start even across midnight, misallocating all hours to day 1 / one
  period (F-CORRECT-SPLIT, P1); midnight ±1s / exactly-24h / out-of-order CSV
  unpinned (F-BOUNDARY-GAPS, P3).
- **§4D late break**: 7-step same-date repro verified live (3 successors, kinds
  work/break_unpaid/work, 0 open). **Defects**: late break after approval/generation
  supersedes approved-linked intervals with no flag — approved sheet keeps stale
  hours (F-LATEBREAK-DESYNC, P1); midnight-spanning break fits neither half, stays
  unpaired forever, blocks generation (F-MIDNIGHT-BREAK, P1); paid-break attach and
  duplicate/ambiguous attribution unpinned (§19).
- **§4E replay**: idempotent replay + conflict-preserves-original + stranded-event
  recovery in single-txn import verified. **Defects**: overlapping_span
  non-blocking + unresolvable + unpaired-gate blocks whole-period generation forever
  (F-OVERLAP-BLOCK, P1); exception dedup key collapses distinct causes
  (F-EXC-DEDUP, P2); crash-recovery by construction only, no live probe (§19).
- **§4F correction**: auditable successor, exactly-once concurrency, decided-sheet
  gate, draft regen, both-day recalc in one txn verified. **Defects**: approved-link
  deadlock (F-APPROVED-DEADLOCK, P1); unrostered new-date keeps old assignment
  (F-MOVE-UNROSTERED, P2); cross-date audit row only on old day (F-CORRECT-AUDIT, P3).

## 8. HR and WFM lifecycle correctness (prompt10 §6)

- **§6A terms**: close-prior-in-tx, overlap exclusion, retroactive block over approved
  periods, audit — service level CONFIRMED FIXED. **Defects**: direct app-SQL UPDATE
  of closed terms succeeds silently and evades `created_at`-keyed conflict detection
  (F-HR-DIRECT-WRITE, P1); post-approval compensation accepted + invisible to
  readiness (F-COMP-RETRO, P1); conflict detection keyed solely on `created_at`
  (F-CREATEDAT, P2); compensation range vs parent terms unchecked (P3); rest_day
  free-text voids rest-day rule (F-REST-DAY, P2).
- **§6B status**: transition matrix, terminal-finality, coherent initial history,
  actor/reason/effective-date CONFIRMED FIXED. **Defects**: direct UPDATE resurrects
  terminated staff bypassing matrix (F-HR-DIRECT-WRITE); concurrent
  SetStatus/SetTerms/SetSHG exactly-once never executed (F-HR-RACE, P2 unverified).
- **§6C roster**: Publish gate (non-empty/week/membership/skills+validity/rest/
  current-revision, one txn, audit metadata), concurrent-publish exactly-once, Revise
  round-trip, v1-immutable CONFIRMED FIXED. **Defects**: no employee-status check in
  Assign/Publish — terminated staff assignable + publishable, hours flow to
  timesheets (F-EMP-STATUS, P1); direct UPDATE publishes invalid rosters / reopens
  published ones (F-HR-DIRECT-WRITE, F-ROSTER-REOPEN); no API ends memberships/skill
  grants (F-NOCLOSE-LIFECYCLE, P2); template validation gaps unchanged
  (F-TEMPLATE-VALID, P2); published_at/by dead columns (P3).
- `SG-MOM-SHIFT` stays unavailable with operational-only note; rotation is
  catalogue-only in code/CLI/OpenAPI/contract — no materialization or legal promise
  (clean). SHG: overlap guard verified; lifecycle write-once per fund, doc comment
  false (F-SHG-NOCLOSE, P1); `verified_none` unusable — CHECK requires fields the
  input lacks (F-SHG-VERIFY, P2).

## 9. Payroll readiness integrity (prompt10 §5)

- **§5A allocation**: R1 INSERT-path, no x-employee/x-employer links, superseded
  excluded, no double-count, period non-overlap, concurrent-generate exactly-once
  (23505→conflict) CONFIRMED FIXED. **Defects**: UPDATE-path single-chain bypass
  (F-SINGLECHAIN-UPDATE); midnight allocation sound by construction for imports but
  broken for corrections (F-CORRECT-SPLIT); crash-after-selection unprobed (§19).
- **§5B lifecycle**: Generate→Submit→Reject→Reissue→Submit→Approve,
  multi-reissue, concurrent-generate, originals immutable CONFIRMED FIXED.
  **Defects**: approved-deadlock (F-APPROVED-DEADLOCK); chain-scope conversion
  sequence static-only (F-CHAIN-SCOPE); Approve TOCTOU spurious 422 under
  concurrency (P3); Generate batch audit after commit (P3).
- **§5C population**: effective-dated population (terms-overlap UNION approved
  holders), terminated-with-approved kept, term conflicts loud, exclusions honored,
  empty/zero-hour blocked CONFIRMED FIXED. **Gaps**: readiness-with-correction
  untested (F-TEST-GAPS); unrelated-period exclusions, terminated-during-period,
  corrected-original inclusion unpinned (§19).
- **§5D approval**: empty-population + zero-hour blocked, run/period binding,
  exclusion ownership/draft-only, same-rule readiness/approval CONFIRMED FIXED
  (readiness-vs-approval live agree/disagree matrix missing, §19). **Defects**: no
  service/CLI/REST transitions a run out of draft — validated/approved states and
  exclusion freeze reachable only via SQL (F-RUN-LIFECYCLE, P2); statutory/
  activation-gate bypass rests on grep + BLOCKED labels, no live attempt (§19).

## 10. Database-role and migration review (prompt10 §7)

- **§7A two-role**: migration_admin owns objects + DDL; all 8 suites assert
  `current_user='app'`; full workflow run as app; loopback-only serve; $n
  parameterization; FromDB sanitization CONFIRMED. **Defects**: documented default
  workflow runs everything as superuser (`.env.example` DATABASE_URL=a1, tasks
  consume it) — privilege layer void outside tests (F-SUPERUSER-DEFAULT, P1);
  post-migrate verification covers only table owners + 7-table revokes, not
  function/search_path/EXECUTE/schema/extension/column-grants (F-POSTMIGRATE, P3).
- **§7B preflight/fresh-install**: Preflight role gate on `migrate up`,
  MIGRATION_DATABASE_URL honored, fresh-install-only decision documented
  CONFIRMED. **Defects**: real `init-roles.sh` fails with syntax error
  (`:'devpw'` inside DO block never substituted) — compose fresh-install dead
  (F-INITROLES, P1); Preflight enforces 3 of 6 §7B conditions, `migrate status`
  never calls it (F-PREFLIGHT-SCHEMA, P2); programmatic superuser MigrateUp applies
  v1–v8 then wedges at v9 ownership guard (F-HALFMIGRATE, P2); migration DSN
  undocumented in README/.env (P2); Q14: yes — bounded divergences across setup
  paths (passwords, PUBLIC-create revoke, wedge owners, superuser-vs-app default);
  pre-pivot upgrade + compose volumes unverified (§19).
- **§7C seed reset**: recycle (explicit admin DSN, protected names, same-server,
  sanitized identifiers, migrate+seed, dedicated-DB test) CONFIRMED FIXED.
  **Defects**: Reset reuses admin password for migration_admin → drop-then-auth-fail
  on password-protected servers (F-RESET-PASSWORD, P2); no disposable-name guard —
  one typo from destroying a dev DB (F-RECYCLE-GUARD, P2); password-auth + typo
  negatives unexecuted (§19).

## 11. API/CLI consistency (prompt10 §9)

Core shared-service rule CLEAN: every handler/cobra resolves employer at the
boundary and delegates to the same `internal/` service; no handler-side business
calculations; transaction boundaries and business defaults live in services;
serve binds 127.0.0.1:8081; X-Operator/--operator documented as attribution.
CLI/REST persisted parity verified live (hours 40/10 + identical readiness;
REST 19/19 incl parity-hours/readiness/drafts-404/envelope/malformed-500/
missing-operator).

- OpenAPI: 43 paths ✓ (grep), 49 served = 49 implemented ✓, 7 drafts unserved + live
  404 ✓. Formal validator tool output missing (§19).
- **Defects**: three envelope shapes, not one — CLI `{error,message}` without
  request_id, prod framework handler `{error}` only (F-ENVELOPE, P3); REST-only
  reads (timesheet list, compensation, SHG, rule-activations) and CLI-only reads
  (skill catalogue, rotation patterns) (F-PARITY-READS, P2); CLI/REST error codes
  diverge on invalid input + CLI prints raw errors (F-CLI-CODES, P3); two handlers
  bypass services with direct SQL (rule-activations, checkTermsOwner) (F-HANDLER-SQL,
  P3); correction handler silently overwrites body interval_id with path param
  (P3); 22P02 malformed UUID → 500 instead of 404/422 (F-UUID-500, P2); contract
  doc drift (envelope shape, aspirational 403, CLI map gaps) + demo-data reset
  wording (F-DOC-API, P3). Rotation materialization + Idempotency-Key correctly
  absent everywhere (clean).

## 12. Statutory-rule isolation (prompt10 §10)

CLEAN with two bounded notes. No hidden OT/CPF/SDL/SHG/net-pay path (grep-clean
across services/seeds/handlers/CLI); readiness sums hours only for >0 presence;
all seed activations BLOCKED v0; `SG-MOM-OT-001` BLOCKED with open ?s, empty
golden/boundary dirs (no fabricated goldens), provisionals honestly labeled;
72h compliance-only; 1.5× documented as Meridian contractual/demo assumption; seed
OT15/OT20 catalogue rows with no rate math; readiness fails closed honestly.

- Money: decimal strings at boundaries → validated strings → NUMERIC(12,2); no
  float64 money anywhere (floats are hours/minutes only). Report §5D claims "Go
  int64 minor units" — zero int64 hits: overclaim, code clean (ST10-01, P3 doc).
  Ruling: string→NUMERIC satisfies §10 intent (exact, no binary float); the int64
  bullet fails as written and the report must be corrected.
- Time rounding: float64 epoch-hours → NUMERIC(10,2)/(8,2) lets PG silently round
  to 2dp with no documented rule; tests use whole-minute spans so never triggers
  (F-TIME-ROUND, P2 — document or quantize in Go).
- Activation-gate bypass: no live attempt past a BLOCKED activation (§19).

## 13. Verification of all Milestone 1.1 claims

### 13a. Prompt 8 cross-reference (audit → remediation → independent result)

| ID | Subject | Regression test | Verdict |
|---|---|---|---|
| P0-1 | Approved-sheet staleness | `TestCorrectRejectsDecidedSheet`, `TestRejectedRecoveryFlow` | CONFIRMED FIXED (rejected path; approved path deadlocks — F-APPROVED-DEADLOCK) |
| P0-2 | Unbounded spans | `TestSpanTooLong`, `TestSpanDateMismatch`, `TestMonthlyExcessIsComplianceOnly` | CONFIRMED FIXED |
| P1-1 | Freeze re-parenting bypass | `TestFreezeReparenting` (P0001 ×4) | CONFIRMED FIXED (core shapes; adjacent paths F-LINE-MOVE/F-SHEET-KEYS/F-ROSTER-REOPEN remain) |
| P1-2 | Duplicate/overlapping intervals | `TestDuplicateSpanFreshIDs`, `TestConcurrentImportsExactlyOnce` | CONFIRMED FIXED (dedup/exclusion; recovery gate F-OVERLAP-BLOCK remains) |
| P1-3 | Conflicting redelivery dropped | `TestConflictingRedelivery` | CONFIRMED FIXED |
| P1-4 | Ambiguous breaks cited-as-paired | `TestMissingBreakEvidence`, `TestLateBreakAttach` | CONFIRMED FIXED (same-date; midnight F-MIDNIGHT-BREAK + desync F-LATEBREAK-DESYNC remain) |
| P1-5 | Session-TZ date casts | `TestUnpairedEventsTimezone` | CONFIRMED FIXED (SQL casts; process-TZ F-TZ-CLASSIFY remains) |
| P1-6 | Flags cleared by upsert/Correct | `TestCorrectOnExceptionDay` | CONFIRMED FIXED |
| P1-7 | Docs-only role separation | `TestRolesAreNotSuperuser`, per-suite app assert | CONFIRMED FIXED (mechanism; F-INITROLES/F-SUPERUSER-DEFAULT/F-HALFMIGRATE/F-PREFLIGHT-SCHEMA remain) |
| P1-8 | Second-transition 23P01 | `TestSetStatus` | CONFIRMED FIXED |
| P1-9 | Approved-history instability | `TestAssessTerminatedWithApprovedSheet`, `TestAssessTermConflict` | CONFIRMED FIXED |
| P1-10 | Publish without validation | `TestPublishGate` | CONFIRMED FIXED (4 negatives; status F-EMP-STATUS + rest-day F-REST-DAY remain) |
| P1-11 | No revise path | `TestReviseRoundTrip`, `TestConcurrentPublishExactlyOnce` | CONFIRMED FIXED |
| P1-12 | Interval IDs SQL-only | `TestListIntervals` | CONFIRMED FIXED |
| P1-13 | Parallel setup race | none | **CONFIRMED DEFECT** (was UNVERIFIED; S8 run1 23505 `pg_authid_rolname_index` → F-PROVISION-RACE, P2 test-infra) |
| P1-14 | Rejected sheets dead-end | `TestRejectedRecoveryFlow` | CONFIRMED FIXED |
| P1-15 | Assign validation gaps | none for status | CONFIRMED DEFECT residual-partial → F-EMP-STATUS (P1) |
| P1-16 | Template validation gaps | none | CONFIRMED DEFECT unchanged → F-TEMPLATE-VALID (P2, bounded: unconsumed) |
| P1-17 | No end-date params | none | CONFIRMED DEFECT unchanged → F-NOCLOSE-LIFECYCLE (P2) |
| P1-18 | Any-to-any transitions | `TestSetStatus` | CONFIRMED FIXED |
| P1-19 | float64 money | `TestCompensationExact` | CONFIRMED FIXED |
| P1-20 | Import/Correct races | `TestConcurrentImportsExactlyOnce`, `TestConcurrentCorrectionsSingleSuccessor` | CONFIRMED FIXED |
| P1-21 | Correction overlap/kind-flip | none dedicated | UNVERIFIED (mechanism present: service check + exclusion; live import-path overlap verified; Correct-path test gap → remediation test item) |
| P1-22 | Rotation pattern list shadowed | none | CONFIRMED DEFECT unchanged (P1, blocks=no) |
| P1-23 | SQL-only onboarding | partial (period/SHG/comp/run added) | CONFIRMED DEFECT residual-partial: employer create still missing CLI+REST (P1, blocks fresh onboarding) |
| P1-24 | Superuser test pools | per-suite `current_user=app` hard-fail | CONFIRMED FIXED |

Tally: 2 P0 both FIXED. 13 blocking P1: 12 FIXED + P1-13 now CONFIRMED DEFECT
(P2 test-infra, live-proven). Other P1: P1-14/18/19/20/24 FIXED; P1-21 UNVERIFIED
test gap; P1-15/23 residual-partial; P1-16/17/22 unchanged.

### 13b. Eighteen structural questions (prompt10 §13)

1. **Multiple current approved versions?** NO via service (single-successor +
   latest-wins); YES via app-SQL x-employee link (two qualifiers live, LIMIT 1
   silent pick) → F-CHAIN-SCOPE. Tie/backdate mispick unforced (§19).
2. **Cyclic / cross-boundary chains?** Cross-employer NO (composite FK live);
   self/2-cycle/x-employee/x-period YES via app SQL (service-unreachable);
   reachable multi-node cycles impossible (freeze + single-parent) → F-CHAIN-SCOPE.
3. **EffectiveApproved wrong under concurrent approval?** NO observed
   (concurrent-approve race: exactly one winner, loser clean InvalidState, single
   audit row). Residual: timestamp-ordering fragility → F-EFFECTIVE-TIE (P3).
4. **Interval removed from exclusion by metadata change?** YES: explicit
   `is_current=false` INSERT bypasses currency maintenance (predecessor+successor
   non-current, hours vanish; overlapping non-current rows exempt) → F-IS-CURRENT-BYPASS.
5. **DEFINER manipulated via writable schema/search_path?** YES via pg_temp
   (implicit first; decoy flip → double-current, 3h+5h counted) despite
   `SET search_path=public`; writable-schema vector clean → F-PG-TEMP.
6. **Stored event never pairable/discoverable?** Discoverable YES (unpaired listing
   + lineage — reconciled); recoverable NO (no void/discard; overlapping-span
   events block generation forever) → F-OVERLAP-BLOCK.
7. **Late break incomplete supersession?** Same-date chain complete; midnight-span
   break never attaches (F-MIDNIGHT-BREAK); paid-break/duplicate attribution
   unpinned (§19).
8. **Corrected interval wrong roster?** YES when new date unrostered (keeps old
   assignment) → F-MOVE-UNROSTERED. Rostered path correct.
9. **Approved sheet loses inputs via later HR/attendance update?** Per-vector:
   terms-via-service NO (retro block + loud conflict); compensation-via-service
   YES silent (F-COMP-RETRO); late-break-via-service YES silent
   (F-LATEBREAK-DESYNC); terms-via-SQL YES silent (F-HR-DIRECT-WRITE,
   F-CREATEDAT). S5-CLEAN "NOT REPRODUCED" holds only for the service-terms vector.
10. **Two runs same period → ambiguous approval?** NO (explicit runID scoping +
    one-approved-regular unique); residual: multi-draft runs coexist, no current-run
    pointer, no validate/approve service → F-RUN-LIFECYCLE.
11. **Exclusion wrong employee/version?** NO (employer-scoped resolve + FK +
    reason/actor + draft-only; version N/A by design).
12. **Historical employee omitted?** NO (terms-overlap UNION approved-holders;
    terminated kept; drift → MissingInfo, never silent drop).
13. **Zero-eligible approval pass?** NO (empty Included → BlockedError; zero-hour
    rejected; tests pin).
14. **Fresh install privilege divergence?** YES, bounded: passwords (compose WITH
    vs recycle/test WITHOUT + Reset password-reuse bug), PUBLIC-create revoke only
    in init-roles.sh, superuser-wedge owners, superuser-vs-app default → §10
    findings. Core grants otherwise identical.
15. **Rollback destroys/invalidates correction history?** Populated: NO destroy —
    loud 23505, atomic, v9 retained, up no-op — but unsupported + undocumented +
    no pre-guard (F-ROLLBACK-DOCS); clean-DB: YES destroys exceptions
    (F-EXCEPTION-LOSS). Published-roster Down unprobed (§19).
16. **CLI/REST diverge on invalid/missing/concurrent input?** Non-persisted YES:
    missing-param codes (CLI internal vs REST validation), CLI raw errors,
    body/path id strictness; persisted state NO divergence (shared services) →
    F-CLI-CODES. No 500-persisted divergence found.
17. **Audit atomic with mutations?** Attendance/timesheet transitions YES (in-txn;
    failed write rolls back audit row live); Generate batch audit NO (after commit,
    P3); CreateEmployee/Assign/catalogue writes have NO audit rows (F-AUDIT-GAPS);
    run/exclusion atomicity unprobed (§19).
18. **Error paths leak protected/cross-employer info?** RECONCILED: unscoped probes
    → 23503 (scoping works); scoped oracle persists (known employer + guessed UUID:
    frozen vs 23503 distinguishes existence) → F-ORACLE P3 negligible (raw-SQL-only;
    app holds SELECT anyway). REST sanitization clean; CLI prints raw local errors.

### 13c. Eight new-defect verdicts (prompt10 §11)

| # | Defect | Verdict | Limit |
|---|---|---|---|
| 1 | UNIQUE blocked append-only | CONFIRMED FIXED | chain-scope guards service-side only (F-CHAIN-SCOPE) |
| 2 | No current predicate | CONFIRMED FIXED | trigger-hardening gaps (F-IS-CURRENT-BYPASS/F-PG-TEMP/F-XEMP-SUPERSEDE) |
| 3 | Dup-span batch failure | CONFIRMED FIXED (import succeeds, advisory exception) | **residual P1**: recovery gate blocks generation forever (F-OVERLAP-BLOCK); report claim "reviewed originals still generate" FALSE |
| 4 | Seed reset vs immutability | CONFIRMED FIXED (recycle) | any non-protected same-server name recyclable by design (F-RECYCLE-GUARD) |
| 5 | Envelope leak, no request IDs | CONFIRMED FIXED for REST | CLI envelope lacks request_id; framework-error shape differs (F-ENVELOPE) |
| 6 | Missing ops | CONFIRMED FIXED (all mutations both surfaces) | read-only asymmetries (F-PARITY-READS) |
| 7 | Dockerfile executables | CONDITIONAL FIXED | shape inspected only; `docker build` unrun — no daemon (wording fix, §19) |
| 8 | Cross-date correction days/roster | CONFIRMED FIXED | audit row on old day only (F-CORRECT-AUDIT); unrostered-date wrong assignment (F-MOVE-UNROSTERED) |

Old-version repros not re-run (no pre-pivot DB; "only if safely possible" not met —
preserved probe evidence + migration diff + test bodies used instead).

### 13d. 98-test and 43-path checks

- **"98 tests"**: UNVERIFIED — no scope reconciles a 98 count; S8 reports 59 skips
  without DB and green DB runs without a 98 tally. Claim must be re-stated with a
  counting method or dropped.
- **43 paths / parity**: 43 OpenAPI paths ✓ (grep), 49 served = 49 implemented ✓,
  7 drafts unserved + live 404 ✓, contract tests enforce both directions.
  Committed `parity_test.go` compares response payloads only; the claimed 137-line
  persisted dump / READINESS-DIFF-EMPTY came from uncommitted `/tmp/e2e_m11.sh`
  (not re-executed); S8 ran its own parity (40/10h + identical readiness) but not
  the claimed dump → §9/§12 claim not re-verifiable until the harness is committed.
  Formal OpenAPI validator output missing (counts + live 404s only).
- **Bare `task lint` go.work failure**: recorded; unruled repo-vs-environment (§19).

## 14. Confirmed defects, risks and untested hypotheses

Severity per prompt10 §14. `blocks` = blocks first payroll calculation milestone.
IDs deduped across all scopes (equivalents merged; source scopes cited).

### P1 — High

**F-OVERLAP-BLOCK** (S8+S4-05; merges D11-03 residual) | `attendance.go:653–662,
346–401`, `timesheet.go:295–309` | Invariant: a handled duplicate delivery must not
permanently remove an employee from generation. Scenario: vendor resend under fresh
IDs raises `overlapping_span` (non-blocking, unresolvable — absent from
resolveExceptions), yet `checkEligible` counts the events unpaired and skips the
employee's generation for the whole period forever; day shows "reviewed".
Evidence: live CLI E2E ("E-PROBE skipped: 6 unpaired clock events", day reviewed).
Why missed: `TestDuplicateSpanFreshIDs` asserts import success + persistence, never
attempts generation after. Fix: operator disposition for duplicate-vendor events
(confirm-duplicate resolution citing event_ids) or exempt overlapping-cited events
from the unpaired gate. **blocks: yes.**

**F-MIDNIGHT-BREAK** (S8) | `attendance.go:878–884,843–931` | Invariant: late break
evidence must attach or flag, never strand. Scenario: break spanning local midnight
fits in neither midnight-split half (containment-in-one-live-interval required),
stays unpaired forever with unmatched/ambiguous exceptions, permanently blocks the
employee's generation; pre-approval recovery needs two order-dependent corrections,
impossible after approval (F-APPROVED-DEADLOCK). Evidence: live E-PROBE 08-26
(built 0, mb-3/mb-4 unpaired). Why missed: `TestLateBreakAttach` covers same-date
only. Fix: split spanning breaks at midnights and attach each half, or match
against predecessor-spanning runs. **blocks: yes.**

**F-APPROVED-DEADLOCK** (S8+S5-01) | `review.go:112–140`, `timesheet.go:474–560` |
Invariant: Correct's own contract — reissue unblocks correction. Scenario: Correct
rejects ANY interval linked to a submitted/approved sheet with no chain awareness;
Reissue links the new draft to the same current intervals but cannot unlink the
immutable approved original → after approve→(find error)→reissue, Correct still
rejects ("reissue the sheet" directs to an action that cannot unblock); approved
hours immutable forever (rejected sheets work; approved never). Evidence: live
DEMO-001 S1→R1 approved→R2 draft→Correct invalid_state. Why missed: recovery test
uses rejected-only originals; reissue-of-approved test never Corrects behind it.
Fix: chain-aware Correct (ignore non-effective approved links with an open/decided
successor covering the chain) + readiness attendance-conflict detection mirroring
checkTermConflict. **blocks: yes.**

**F-COMP-RETRO** (S8) | `hr.go:453–509`, `readiness.go:281–319` | Invariant: payroll
basis covered by an approval must not change silently. Scenario: allowance effective
inside an approved period accepted after approval; readiness checks terms+SHG only,
never `compensation_line` → no conflict, silently changed basis. Evidence: live CLI
compensate MEAL 150.00 from 08-01 after approval → accepted, readiness unchanged.
Why missed: no test adds post-approval compensation or reads compensation in
readiness. Fix: mirror SetTerms retro block in AddCompensation; add compensation to
term-conflict check. **blocks: yes.**

**F-TZ-CLASSIFY** (S4-01) | `attendance.go:968 classifyBreak`, pgx
`timestamptz.go`, `db.go:14–23` | Invariant: pairing independent of process/host TZ.
Scenario: UTC host serving SGT roster — break decodes with `time.Local` (pgx
default; app never sets ScanLocation), shift anchor parsed in UTC → windows shift
+8h → legitimate break misclassified ambiguous, stays unpaired, unpaid minutes
remain in work run (paid overstated), employee skipped from generation. Evidence:
driver-source + code inspection (static; TZ=UTC run never executed — §19).
Why missed: all pairing tests run where Local==event zone (SGT-only). Fix: pass
event declared tz into classifyBreak, `time.LoadLocation(tz)` for anchor, never
`start.Location()`; add TZ=UTC + mixed-zone tests. **blocks: yes.**

**F-CORRECT-SPLIT** (S4-02) | `review.go:155–179` vs `attendance.go:719–737` |
Invariant: corrected bounds attribute to business dates like imports. Scenario:
Correct inserts ONE interval with local_date=start even across midnight (import
splits via splitRun) → night-shift 22:00–06:00 lands all 8h on day 1, day 2 misses
6h; across period boundary all hours misallocate to one period. Evidence: static
(no live SELECT probe — §19; conflicts with S5-CLEAN "by construction" claim —
reconciled: construction holds for imports only). Why missed:
`TestCorrectCrossDate` moves within single dates only. Fix: reuse splitRun in
Correct (one row per dated run, per-date roster, same txn). **blocks: yes.**

**F-LATEBREAK-DESYNC** (S4-03) | `attendance.go:846–931` vs `review.go:110–140` |
Invariant: approved sheets must not silently desync from current intervals.
Scenario: import→generate→approve→late break arrives: attachLateBreaks supersedes
the approved-linked work interval with 3 successors; approved sheet keeps stale
hours, no flag; draft sheets go stale too (Import never calls RegenerateDraft).
Evidence: static (no late-break-after-approval test exists). Why missed:
`TestLateBreakAttach` has no timesheet in play. Fix: mirror Correct's linkage gate
in attachLateBreaks (reject-or-flag on submitted/approved links; regen drafts
in-txn) or emit blocking exception on desync. **blocks: yes.**

**F-EMP-STATUS** (S6-P1-02 + SYN-P1-15 + F-PUBLISH-GAPS-a) | `roster.go:234–391,
450–545` | Invariant: non-active staff must not receive publishable shifts.
Scenario: terminate employee (membership+skills rows remain), Assign every day of a
draft week, Publish — all succeed; clock events roster-anchor and flow into payable
intervals/timesheets for resigned/terminated staff. Remediation claim "Publish
revalidates status" means roster status, not employee status. Evidence: static
(validatePlacement has no employee.status reference) + live S8 repro (terminated
08-10 assigned, 08-17 roster published). Why missed: publish tests use active staff
only; report wording masks gap. Fix: employee-status check in validatePlacement
(reject non-active per decided policy incl. suspended) + negative tests at assign
and publish. **blocks: yes.**

**F-IS-CURRENT-BYPASS** (S3-C1) | `00009:141–156`, `00002:9–38` | Invariant: app
must not control currency metadata. Scenario: app INSERT with explicit
`is_current=false`: (a) successor leaves predecessor+successor non-current
(current_visible=0, verified work vanishes from timesheets/readiness); (b)
overlapping non-current rows insert cleanly, exempt from exclusion. Evidence: live
as app (P5b/P5c/P10b; f/f rows + count 0). Why missed: no test inserts explicit
is_current; Go omits the column (DEFAULT true); "app cannot UPDATE" ignores
INSERT-time control. Fix: `NEW.is_current:=true` unconditionally in trigger AND
`REVOKE INSERT (is_current)` from app. **blocks: yes** (silent input loss;
raw-SQL-only reachability).

**F-XEMP-SUPERSEDE** (S3-C2 + F-XEMP) | `00009:144–146`, `00002:29` | Invariant:
corrections must not touch another employee's lineage. Scenario: E002 "correction"
superseding E001's live interval succeeds; victim flips non-current (hours vanish
from their sheet), attacker current. Evidence: live as app (P6a: E001 f / E002 t).
Why missed: single-employee chains only; Correct() reuses orig employee. Fix:
BEFORE INSERT trigger requiring NEW.employee_id = predecessor's (or composite
FK). **blocks: yes** (raw-SQL-only).

**F-PG-TEMP** (S3-C3) | `00009:142–150` | Invariant: DEFINER maintenance must hit the
real table. Scenario: pg_temp implicitly searched first despite `SET
search_path=public`: app creates TEMP `attendance_interval` + grants to
migration_admin → successor INSERT flips TEMP decoy (f), PUBLIC predecessor stays
current (t) alongside successor (t) → double-current, both feed REG line (3h+5h not
5h). Evidence: live exploit as app (temp f / public t+t); without GRANT fails
closed (42501/42703). Why missed: no temp-table tests; search_path assumed
sufficient. Fix: schema-qualify `UPDATE public.attendance_interval`; defense in
depth: `REVOKE TEMPORARY ON DATABASE FROM app` + explicit `REVOKE CREATE ON SCHEMA
public`. **blocks: yes** (raw-SQL-only).

**F-LINE-MOVE** (S3-D1) | `00009:206–231` | Invariant: frozen evidence must not move
via relationship changes. Scenario: cross-employee line move draft→draft ALLOWED
as app (sheet_emp=E002 vs iv_emp=E001); linked intervals ride along without
re-firing employee guard; victim can never relink ("already feeds sheet of another
employee"); line/sheet period coherence silently broken. Evidence: live D4a + victim
lockout probe. Why missed: freeze tests move lines within same employee/period
into approved parents only. Fix: on `timesheet_line` parent-key UPDATE with links,
require NEW sheet (employee,period) match all linked intervals, or forbid parent
change with links (service-safe: no Go UPDATEs this table). **blocks: yes**
(raw-SQL-only).

**F-SHEET-KEYS** (S3-D2) | `00009:201–205` | Invariant: sheet identity immutable once
linked. Scenario: draft sheet employee/period keys mutable as app, incl. atomically
WITH approval: `UPDATE timesheet SET status='approved',employee_id=E001` on E002's
sheet → APPROVED E001 sheet holding E002's interval; period transplant leaves
line/sheet mismatch; victim locked out. Evidence: live D5-ter + D6-redo (UPDATE 1).
Why missed: transition() only sets status+stamps; no test moves keys with links.
Fix: forbid employer/employee/period change on UPDATE when links exist (or always;
service-safe). **blocks: yes** (raw-SQL-only).

**F-ROSTER-REOPEN** (S3-D3 + S6-P1-01c) | no roster trigger; `roster.go:520` never
stamps `published_at` | Invariant: published assignments immutable. Scenario:
unpublish as app (`UPDATE roster SET status='draft'` → UPDATE 1) then EDIT
formerly-published assignment; current_version rollback re-opens superseded
revisions; no audit; interval roster context silently reinterpretable. Evidence:
live D7b/D7c. Why missed: `TestRepairByNewRevision` blesses reopen but never edits
v1 inside the window. Fix: stamp `published_at` on Publish; extend assignment
freeze to stamped revisions (Revise unaffected: new revision unstamped). **blocks:
yes** (raw-SQL-only).

**F-HR-DIRECT-WRITE** (S6-P1-01) | `00004:14–30` revoke list; HR/WFM parents |
Invariant: employment history immutable except via governed transitions.
Scenario: (a) UPDATE closed `employment_terms.monthly_basic` as app succeeds
silently; readiness conflict keys on `created_at` (UPDATE doesn't touch it) → no
flag, approved inputs change silently. (b) UPDATE employee status resurrects
terminated staff bypassing matrix. (c) UPDATE roster status publishes invalid
rosters bypassing gate. Evidence: static + proof-by-tests (readiness/publish/
revision tests themselves depend on direct-UPDATE-as-app succeeding). Why missed:
suites exercise service API only; revoke checks cover 7 insert-only tables.
Fix: freeze triggers for closed HR/WFM history (or revoke UPDATE/DELETE + DEFINER
maintainers) + negative app-role tests per table. **blocks: yes** (raw-SQL-only;
reviewer judgment per §16 text).

**F-INITROLES** (S8) | `scripts/db/init-roles.sh:10–25` | Invariant: documented
fresh install must provision roles. Scenario: `:'devpw'` inside `DO $do$` block
never substituted by psql (no interpolation in dollar quotes) → syntax error;
roles/schema/grants never created; compose fresh-install dead. Evidence: live run
against disposable PG16 (ERROR syntax error at :'devpw'); mechanism isolated
(heredoc outside DO substitutes, inside fails). Why missed: no test executes the
script; dbutil provisions differently. Fix: CREATE ROLE outside dollar quotes
(format+%L via \gexec) or :'var' in plain SQL. **blocks: yes** (fresh install).

**F-SHG-NOCLOSE** (S8+S6-P1-03) | `shg.go:43–130` | Invariant: fund instructions
supersedable over time. Scenario: SetSHG rejects any same-fund overlap and no API
closes an open row → open-ended active instruction blocks all future same-fund
rows (active→opt_out always conflict); doc comment promises predecessor-closing
that doesn't exist. Evidence: live CLI (opt_out from 09-01 → conflict). Why
missed: tests pin overlap-rejection only. Fix: close-prior-in-tx in SetSHG (mirror
SetTerms) + fix comment + round-trip test. **blocks: no** (SHG unconsumed).

**F-SUPERUSER-DEFAULT** (S6-P1-04) | `.env.example`, Taskfile serve/seed/db tasks,
`cli.go:28–32` | Invariant: supported workflow must run least-privilege.
Scenario: default local workflow runs everything as superuser (DATABASE_URL=a1,
MIGRATION empty, comment blesses it); "Runtime uses app" holds only for tests +
manual overrides → privilege layer void for README followers (only triggers bind).
Evidence: static (config read). Why missed: suites override DSN; default env
untested. Fix: ship app-role DATABASE_URL + migration_admin MIGRATION_DATABASE_URL
examples; document both DSNs; startup loud warning/refusal on superuser runtime.
**blocks: no** (config; voids protections).

**P1-22 rotation CLI** (SYN, unchanged) | `cmd/workforceops/wfm.go:343–381` | Both
subcommands register cobra Name `pattern`; `rotation pattern list` unreachable
(prints create usage). Fix: parent pattern command with create|list children.
**blocks: no.**

**P1-23 employer create** (SYN, residual-partial) | CLI `employerCmd` show-only, GET
employer only | Fresh DB without seed cannot onboard any employer without manual
SQL. Fix: employer create on CLI+REST or normatively document interim SQL.
**blocks: yes** (fresh onboarding).

### P2 — Medium

F-SINGLECHAIN-UPDATE (S8+S5-02; INSERT-only guard; app UPDATE re-points link →
one interval feeds two periods; fix: BEFORE INSERT OR UPDATE); F-CHAIN-SCOPE
(S2+S5-03+F-CHAIN; supersedes lacks employee/period/cycle guards; self/2-cycle/
x-employee live; LIMIT-1 silent pick; fix: scope-match trigger + CHECK + multiple-
effective error); F-INTERVAL-SELFLOOP (S3-C4; self-supersede poisons successor
slot, legit correction then 23505 unrepairable; fix: CHECK); F-MOVE-UNROSTERED
(S8+S4-04; cross-date move to unrostered day keeps old assignment; fix:
resolve-or-NULL + no_roster_link); F-OVERLAP-DISPOSITION (S4-05 part of overlap
root: overlapping_span non-blocking + unresolvable + dedup — fix folded into
F-OVERLAP-BLOCK; listed here for the blocking-code/resolve-list half);
F-EXC-DEDUP (S4-06; dedup key collapses distinct causes; fix: include event_ids);
F-OFF-WINDOW (S4-07; shift window/rest-day never consulted; fix: flag exception or
spec waiver); F-HALFMIGRATE (S8+S6-P2-04; superuser MigrateUp applies v1–v8 then
wedges; fix: Preflight inside MigrateUp + recovery doc); F-PREFLIGHT-SCHEMA
(S8+S6-P2-03; 3 of 6 §7B conditions; status never preflights; fix: extend + call
from status); F-PROVISION-RACE (S8; ex-SYN-P1-13 now live-proven 23505 on fresh
clusters; fix: cluster advisory lock or duplicate_object handler);
F-PUBLISH-REST (F-REST-DAY: S6-P2-01 + F-PUBLISH-GAPS-b; free-text rest_day voids
rule; fix: allowlist/canonicalize + DB CHECK); F-UUID-500 (S8; 22P02 → 500 not
404/422 on all UUID paths; fix: map input-error class); F-CREATEDAT (S8; conflict
keys solely on created_at, evadable via SQL; fix: updated_at trigger + GREATEST +
revoke writes); F-NOCLOSE-LIFECYCLE (S8+S6-P2-07 + SYN-P1-17; no API ends
memberships/skills; fix: end-date operations); F-RUN-LIFECYCLE (S5-04; no
validate/approve service — states SQL-only, exclusions mutable forever via
workflow; fix: payroll slice owns run state machine); F-ROLLBACK-DOCS
(S2-Q15-POPULATED; loud+atomic but no pre-guard + undocumented; fix: Down
preflight + §10 wording); F-EXCEPTION-LOSS (S2-Q15-EXC; clean-DB Down drops
exception log; fix: document destructive / preserve / refuse); F-RESET-PASSWORD
(S6-P1-05; Reset reuses admin password → drop-then-auth-fail; fix: explicit
migration DSN or pre-auth before drop); F-RECYCLE-GUARD (S6-P2-05; no
disposable-name guard; fix: allowlist/--force + move same-server check in);
F-SHG-VERIFY (S6-P2-02; verified_none always 23514; fix: add verification fields
or drop status); F-TEMPLATE-VALID (SYN-P1-16, unchanged; overlap/containment/
kind-paid/cross-midnight unchecked; bounded — unconsumed; fix: validation +
tests); F-PARITY-READS (S7-02; REST-only + CLI-only reads; fix: add list verbs or
document); F-HANDLER-SQL minor→P3 (see below); F-TEST-GAPS (S2; no concurrent-
Reissue / readiness-with-correction / audit-assertion / negative-SQL tests; fix:
add pinned tests); F-TIME-ROUND (ST10-02; float-hours → NUMERIC silent 2dp; fix:
document or quantize in Go); F-HR-RACE (S6-P2-08; HR concurrency UNVERIFIED —
mechanism sound, unexecuted; fix: race tests); F-Q16 (S7-Q16; CLI/REST invalid-
input code divergence; fix: carry validation code + negative parity tests);
F-CLI-CODES (S7-03; same root as Q16 CLI half); F-MIGRATE-DOCS (S6-P2-09;
migration DSN undocumented → README fresh-install fails loudly; fix: document);
F-CONTRACT-403 (S6-P2-10; aspirational 403, any resolvable employer accepted;
fix: pin or correct contract); F-AUDIT-GAPS (S6-P2-06; require-and-discard
operator on create/assign/catalogue writes; fix: audit.Write or downgrade
requirement); F-EFFECTIVE-TIE→P3 (see below).

### P3 — Low

F-EFFECTIVE-TIE (S2+S5-05; two-qualifier chains rely on wall-clock order; fix:
chain-terminal resolution or deterministic tiebreaker); F-ORACLE (S3-D4 reconciled
with S7-Q18; scoped frozen-vs-23503 oracle; negligible; optional generic message);
F-ENVELOPE (S7-01; 3 shapes; fix: CLI request_id + align framework handler);
F-DOC-API (S7-05 + F-DOC-SKEW; contract/doc drift: envelope/"subtree"/TLS/403;
fix: correct wording); F-HANDLER-SQL (S7-04; rule-activations + checkTermsOwner
direct SQL, no divergence found; fix: move into services); F-CORRECT-ID (S7-06;
body interval_id silently overwritten; fix: reject mismatch); F-CORRECT-AUDIT
(S4-09 + D11-08 limit; new-day record lacks audit row; fix: write after upsert or
document); F-BOUNDARY-GAPS (S4-10; midnight±1s/24h/out-of-order/orphan/stale-
revision unpinned; fix: pinned CSV cases); F-XSOURCE-PAIR (S4-08; cross-source
pairing silent; fix: segregate or flag); F-APPROVE-TOCTOU (S5-06; concurrent
re-approve spurious 422, retry succeeds; fix: idempotent-aware transition);
F-GEN-AUDIT (S5-07; batch audit after commit; fix: best-effort or per-sheet rows);
F-ROBUST-NITS (S8: 40P01→internal incl concurrent CreatePeriod, generateOne
"already exists" misreport, reset-target-must-exist, positional args ignored,
grammar, overtime/rest kinds 23514, corrections lack span ceiling, inert
is_current=false plants, no_roster_link re-link, unordered classifyBreak scan);
F-MODEL-NITS (S8+S6-P3-01/02/03: segment paid unread, Generate summary audit
non-atomic by design, activation-list/rotation-pattern surface deltas, dead
published_at/by columns, compensation-range coherence); F-POSTMIGRATE (S6-P3-04;
function/search_path/EXECUTE/schema/extension/column-grants unasserted; fix: post-
migrate assertions + optional EXECUTE revoke); F-REPORT-INT64 (ST10-01; "int64
minor units" overclaim, code clean; fix: correct §5D); F-D11-07-WORDING (build
unrun; conditional FIXED).

### Risks (need evidence) / untested hypotheses

- Reissue-walk hang on self-cycle: predicted (no visited set) but deliberately
  unexecuted; DISPROVED as reachable (single-parent invariant + freeze) — residual
  risk negligible; add depth cap as defense in depth.
- EffectiveApproved tie/backdate mispick: mechanism live (two qualifiers), mispick
  never forced (freeze blocks forgery).
- HR concurrency (F-HR-RACE), session-TZ live run, boundary matrix, crash/replay
  probes, chain-conversion sequence live, activation-gate attempt: §19.
- S6-P1-01 blocking judgment: silent employment-input mutation meets §16 text but
  requires direct SQL as app — owner to confirm threat-model weight.

## 15. Clean areas

Correction-chain core (single-successor incl true concurrency, originals immutable,
draft/rejected never displace effective, NEW-parent freeze both directions);
DEFINER core matrix (owner/search_path/revokes/exclusion/races/Q4/Q17-attendance);
72h compliance-only incl 210h full line; replay idempotent + conflict preserves
original; concurrent corrections/imports/generate exactly-once; P0-2 span ceilings;
session-TZ-safe SQL casts; rejected-recovery chain + reissue guards; concurrent-
approve one-winner; terminated-with-approved + term conflicts (service vectors);
empty/zero-hour approval blocks; run/period/exclusion binding; Publish/Revise
atomicity + races + metadata; status matrix + history coherence; SHG overlap guard;
8 app-role suites; recycle protected/same-server + dedicated-DB reset test;
loopback serve; parameterization; FromDB/REST sanitization; shared-service rule
(no handler business logic); OpenAPI counts + live draft-404s; CLI/REST persisted
parity; money string→NUMERIC, no float money, no hidden calcs; OT-001 BLOCKED with
empty goldens; rotation/SG-MOM-SHIFT correctly deferred; single-binary packaging
shape; exception stickiness; ListIntervals evidence fields.

## 16. Four independent verdicts

1. **Functional correctness: FAIL.** Service-reachable P1s corrupt or block the
   evidence chain through the supported workflow: F-OVERLAP-BLOCK (permanent
   generation skip), F-MIDNIGHT-BREAK (permanent strand), F-APPROVED-DEADLOCK
   (post-approval errors uncorrectable), F-COMP-RETRO (silent basis change),
   F-TZ-CLASSIFY (misclassification + overpay), F-CORRECT-SPLIT (midnight/period
   misallocation), F-LATEBREAK-DESYNC (approved desync), F-EMP-STATUS (terminated
   staff payable). Each independently violates §16.
2. **Database integrity: FAIL (bounded).** Core invariants verified live
   (single-successor, freeze, exclusion, atomicity, loud rollback), but nine
   hardening gaps let app-role SQL silently change payroll inputs or history:
   F-IS-CURRENT-BYPASS, F-XEMP-SUPERSEDE, F-PG-TEMP, F-LINE-MOVE, F-SHEET-KEYS,
   F-ROSTER-REOPEN, F-HR-DIRECT-WRITE, F-SINGLECHAIN-UPDATE, F-CHAIN-SCOPE (+
   F-INTERVAL-SELFLOOP). Under the project's own app-role threat model these are
   genuine integrity defects; all are service-unreachable, hence bounded.
3. **CLI/API operability: PASS WITH CONDITIONS.** Full CLI E2E + REST 19/19 +
   persisted parity verified; OpenAPI counts match. Conditions: F-ENVELOPE,
   F-PARITY-READS, F-UUID-500, F-CLI-CODES, P1-23 employer create, F-DOC-API,
   F-INITROLES (fresh install dead), F-SUPERUSER-DEFAULT (documented default voids
   privilege layer).
4. **Payroll-readiness integrity: FAIL.** F-COMP-RETRO + F-LATEBREAK-DESYNC
   silently move approved inputs via service; F-CREATEDAT + F-HR-DIRECT-WRITE
   evade conflict detection via SQL; readiness-with-correction untested
   (F-TEST-GAPS); run lifecycle draft-only (F-RUN-LIFECYCLE). Population/approval
   gates themselves verified.

## 17. Overall recommendation: FAIL

Per prompt10 §16, FAIL requires a P0 or material P1 threatening inputs or
rule-gating. This review finds **no P0** but **eight service-reachable P1s that
independently corrupt, misattribute, duplicate, lose, or silently change
attendance/timesheet/employment/readiness data** (F-OVERLAP-BLOCK,
F-MIDNIGHT-BREAK, F-APPROVED-DEADLOCK, F-COMP-RETRO, F-TZ-CLASSIFY,
F-CORRECT-SPLIT, F-LATEBREAK-DESYNC, F-EMP-STATUS), plus F-INITROLES breaking
fresh install and P1-23 blocking SQL-free onboarding. The first independently
verified payroll calculation **must not begin** until the §18 list is resolved and
re-verified live. The core architecture is sound — correction chains, freeze
triggers, exclusion enforcement, and readiness gates verified live — so the FAIL
is a bounded defect list, not an open-ended exercise.

## 18. Smallest justified remediation list

Order matters within groups. Each item needs its stated regression test + live
re-verification; re-run the §3 command set on a disposable PG16 at the end.

**A. Service-reachable payroll-input P1s (all block first payroll):**

1. F-OVERLAP-BLOCK: confirm-duplicate disposition (or gate exemption) + test:
   resend → disposition → Generate includes employee.
2. F-APPROVED-DEADLOCK: chain-aware Correct + readiness attendance-conflict
   detection + test: Approve→Reissue→Correct succeeds; stale effective flagged.
3. F-LATEBREAK-DESYNC: linkage gate in attachLateBreaks (reject/flag on
   submitted/approved; regen drafts in-txn) + test with timesheet in play.
4. F-COMP-RETRO: retro block in AddCompensation + compensation in conflict check
   + post-approval compensation test.
5. F-TZ-CLASSIFY: tz-aware classifyBreak + TZ=UTC and mixed-zone pairing tests.
6. F-CORRECT-SPLIT: splitRun reuse in Correct + cross-midnight/period-boundary
   correction tests asserting per-date rows.
7. F-MIDNIGHT-BREAK: split spanning breaks at midnights + midnight-spanning
   late-break test.
8. F-EMP-STATUS: status check in validatePlacement (decide suspended policy) +
   assign/publish negative tests for terminated/resigned.

**B. Fresh-install / onboarding P1s:**

9. F-INITROLES: fix `:'devpw'` substitution; execute real script on disposable PG
   in CI-equivalent step.
10. P1-23: employer create on CLI+REST (or normatively document interim SQL).
11. F-SUPERUSER-DEFAULT: app-role `.env.example` + migration DSN docs + startup
    superuser warning.

**C. DB-hardening set (one migration + tests; raw-SQL-only but same bar as P1-1):**

12. F-IS-CURRENT-BYPASS + F-PG-TEMP + F-XEMP-SUPERSEDE + F-INTERVAL-SELFLOOP:
    unconditional `NEW.is_current`, schema-qualified UPDATE, employee-match guard,
    self-CHECK, TEMPORARY revoke; adversarial app-role SQL tests per vector.
13. F-LINE-MOVE + F-SHEET-KEYS + F-ROSTER-REOPEN + F-HR-DIRECT-WRITE: parent-key/
    roster/HR-history freeze rules; negative app-role UPDATE tests per table.
14. F-SINGLECHAIN-UPDATE: BEFORE INSERT OR UPDATE; UPDATE re-point rejection test.
15. F-CHAIN-SCOPE + F-EFFECTIVE-TIE: scope-match trigger + self-CHECK +
    multiple-effective error + chain-terminal resolution; negative SQL + tie tests.

**D. P2 correctness/robustness before broader use:**

16. F-SHG-NOCLOSE (close-prior supersession), F-REST-DAY (allowlist + CHECK),
    F-MOVE-UNROSTERED (resolve-or-NULL), F-HALFMIGRATE (Preflight in MigrateUp),
    F-PREFLIGHT-SCHEMA (ownership/grants + status), F-PROVISION-RACE (cluster
    lock), F-UUID-500 (22P02 mapping), F-CREATEDAT (updated_at trigger),
    F-NOCLOSE-LIFECYCLE (end-date ops), F-RUN-LIFECYCLE (document draft-only or
    build state machine), F-ROLLBACK-DOCS + F-EXCEPTION-LOSS (Down guard + docs),
    F-RESET-PASSWORD (explicit migration DSN), F-RECYCLE-GUARD, F-SHG-VERIFY,
    F-PARITY-READS, F-TEST-GAPS (concurrent-Reissue, readiness-with-correction,
    audit assertions), F-TIME-ROUND (document), F-HR-RACE (race tests),
    F-AUDIT-GAPS, F-TEMPLATE-VALID, P1-22 rotation CLI, P1-21 Correct-overlap test.
17. P3/doc sweep: F-ENVELOPE, F-DOC-API, F-HANDLER-SQL, F-CORRECT-ID,
    F-CORRECT-AUDIT, F-BOUNDARY-GAPS, F-XSOURCE-PAIR, F-APPROVE-TOCTOU,
    F-GEN-AUDIT, F-ROBUST-NITS, F-MODEL-NITS, F-POSTMIGRATE, F-REPORT-INT64
    (correct §5D to string→NUMERIC), F-D11-07-WORDING, 98-test count wording,
    F-ORACLE (optional), commit E2E persisted-diff harness, run formal OpenAPI
    validator.

## 19. Explicit unverified areas

1. M1 snapshot diff (§1 item 6): unverifiable — no M1 snapshot in repo.
2. Full-diff read attestation (§1 item 5): inventory done; line-complete read of
   every untracked file unattested beyond scope coverage.
3. Systematic vacuity audit (§2): isolated pins only (e.g. rows==1 matches
   code-not-comment per F-CORRECT-AUDIT).
4. Populated Down with published rosters (§3B/§12): S2/S8 record no published
   rosters in the rollback DB.
5. Reissue-walk execution on self-cycle: deliberately unexecuted (DISPROVED
   reachable; cap still recommended).
6. EffectiveApproved tie/backdate mispick: never forced live.
7. Chain-conversion sequence live (S5-03 variant b): static only.
8. Multiple-shifts-per-day pairing: no probe/test anywhere.
9. Boundary matrix live (midnight±1s, exactly-24h, out-of-order CSV, month-long
   orphan, stale-revision roster): code-inspected only.
10. Session-TZ live run (TZ=UTC pairing tests): never executed; S4-01 rests on
    driver-source inspection. Suite green may reflect TZ=Asia/Singapore masking.
11. F-CORRECT-SPLIT live SELECT probe: static only.
12. Late-break variants live (paid attach, duplicate/ambiguous attribution).
13. Import crash/replay + generation crash/replay live probes: by construction only.
14. Readiness sub-cases (unrelated-period exclusions, terminated-during-period,
    corrected-original inclusion): unpinned.
15. Statutory/activation-gate live bypass attempts: grep + labels only.
16. Readiness-vs-approval live agree/disagree matrix: missing.
17. HR live + concurrency (S6 had no DB): SetStatus/SetTerms/SetSHG races unexecuted.
18. Assignment-uniqueness publish violation: no live/test cite.
19. Post-migrate catalog assertions: S8 "catalog verification" noted without pasted
    output; function/search_path/EXECUTE/schema/extension/column-grants unasserted.
20. Incompatible-migration-state + `migrate status` negatives: code reading only.
21. Pre-pivot upgrade + compose volumes: not re-run/inspected (no daemon).
22. Seed-reset negatives live (password-auth failure, typo-target guard): unexecuted.
23. Security live probes: injection (grep only), CLI raw-error capture (static).
24. §10 int64 bullet: fails as written — ruled satisfied-by-string→NUMERIC (owner
    to confirm intent).
25. Old-version repros (§11): not re-run — no pre-pivot DB ("only if safely
    possible" not met).
26. Docker build: unrun (no daemon) — D11-07 conditional FIXED.
27. Parallel-vs-serial run2: command line proving run2 stayed parallel unrecorded.
28. Probe scripts (`/tmp/s2step*`, `/tmp/p10s3*`, `/tmp/p9*.py`): outside repo,
    unattached — commands+outputs not reader-reviewable.
29. Bare `task lint` go.work failure: unruled repo-vs-environment.
30. PG<15 PUBLIC-CREATE divergence: untested (all probes PG16.15).
31. Exception-regen after Down loss: never probed.
32. Q17 run/exclusion audit atomicity: unprobed; S5-07 defect stands.
33. 98-test count: unreconciled; E2E dump harness uncommitted; OpenAPI validator
    tool not run.
34. F-TZ-CLASSIFY, F-CORRECT-SPLIT, F-LATEBREAK-DESYNC, F-HR-DIRECT-WRITE rest on
    static inspection (reviewers had no live DB in those scopes) — re-verify live
    with the §18 regression tests before accepting fixes.
35. Load/performance (large-import tx, ListIntervals N+1, unbounded lists): not
    measured — noted, not findings.

---

*Synthesized from 8 evidence scopes + completeness critic. No repo file modified
except this report. Working tree left untouched (uncommitted M1/M1.1 work preserved).*
