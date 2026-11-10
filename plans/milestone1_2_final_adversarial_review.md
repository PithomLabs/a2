# M1.2 Final Adversarial Review (prompt12 synthesis)

Synthesized ONLY from the 8 prior workflow children (7 primary + 1 critic).
No new code inspection, no new live probes, no repo modifications in this
synthesis step. Severity labels are prompt12 P1/P2/P3 as filed by children,
except where this review explicitly adjudicates a conflict.

Decision: **FAIL** per prompt12 §12 (silent DB-level gaps remain under the
app role). Service happy paths are sound and suites are green, but PASS
requires DB-enforced guarantees under app, not green suites alone.

---

## 1. Baseline + scope

- Baseline: HEAD `f43515883668c917edac31be49a7cf8334eca5c2` (confirmed by
  primaries 3 and 7).
- Worktree at review time: ~96 entries — 39 untracked, 18 modified,
  4 deleted, 35 renamed (primary-7). Tracked diff ≈ 57 files +2009/−679.
- Migration state: tracked delta under `migrations/` is ONLY the deletion
  of `00001_schemas.sql`; files `00001_foundation`…`00010_m12_integrity`
  are ALL untracked vs baseline. Prompt12 §1 item 5 ("00001–00009
  untouched") is therefore **unverifiable from git** — there is no
  prompt10 snapshot; M1.1+M1.2 work is one uncommitted layer. Supported
  only by file-novelty + mtime ordering (00001–00009 Nov-8 16:34–20:25,
  00010 Nov-9 00:43) + 00010's self-declaration as the single append-only
  M1.2 migration with reverse-order Down (00010 L1–6, L703–817).
  Filed as S1-BASELINE/P3 (doc precision, not a code defect).
- Toolchain (evidenced): Go 1.27.3, task 3.48.0, psql 16.14 client.
  `GOWORK=off` REQUIRED for every go command (parent go.work captures the
  module) plus sandboxed GOCACHE/GOPATH/GOTMPDIR.
- Scope covered: Area A correction groups (static only), Area B duplicate
  provenance (static only), Area C timezones (live matrix), Area D
  compensation (live as real app), Area E migration/rollback (static
  only), §7 privilege/app-role invariants (live), §8 regression
  reconciliation (live), statutory-BLOCKED check (grep + inspection).
- Environment split: primaries 1 (A), 4 (B), 2 (E), 3 (S1) had **no**
  PostgreSQL server (client-only binaries, no server, no
  `WORKFORCEOPS_TEST_DATABASE_URL`) — their claims are statically
  inspected (bodies read, tests read not run). Primaries 5 (D), 6 (C),
  7 (§7/§8) ran disposable PG16 clusters live. The critic confirms all 8
  handoff parts readable with no read failure.

## 2. Commands evidenced

Static children (no DB): `git rev-parse HEAD`, `git ls-tree -r HEAD`,
`git status`/worktree inventory, `GOWORK=off go build ./...` → BUILD_OK
(B child), `go list` (needs GOWORK=off + GOCACHE=/tmp), file reads of
00010 (817 lines), service bodies, and test bodies; repo-wide greps
(overtime/cpf/netpay/payslip → zero hits; float64 → durations/display
only; Down-execution search → none).

Live children (disposable PG16.15, TCP-only, real app role unless noted):
`task lint` → exit 0, `task build` → exit 0, `git diff --check` → exit 0,
`go test -count=1 ./...` → exit 0 under default TZ, `TZ=UTC`, and
`TZ=Asia/Singapore` (17 ok + 7 no-test-files, 0 skips each),
`go test -race` on attendance/timesheet/payroll/hr/wfm → exit 0, no DATA
RACE; integrity verbose 61 PASS / 0 FAIL / 1 intentional SKIP
(TestReproducibilityDeferred); employer 5/0/0, provisioning 2/0/0, seed
7/0/7 (primary-7; seed run included one port-collision first-run failure
caused by the child's own cluster, not a defect); hr+payroll suites pass
on disposable PG16 (primary-5); 13 TZ/midnight/audit tests pass under
both host TZ (primary-6); app-role SQL probes with SQLSTATE + re-reads
for all §7 invariants (primary-7) and D1–D4 matrices (primary-5);
C3 TZ matrix (TZ=UTC+America/New_York session vs
TZ=Asia/Singapore+Pacific/Kiritimati session) with byte-diff confined to
3 rendering-only strings (primary-6).

NOT executed anywhere: populated 00010 Down / Down+Up / mid-Down failure
/ version-after-failure probes; A2 7-step live replay with
between-segments failure; concurrent direct-SQL A↔B disposition race;
concurrent Go-service AddCompensation-vs-Approve race; fresh-install
10-step combined run past employer creation; Docker Compose path and
serve-loopback smoke.

## 3. Correction-group findings (Area A, static)

Five P1 app-role/service gaps, each blocks-first-payroll=yes. Service
path via `Correct()` is coherent (single tx, advisory lock, liveness
recheck, per-segment roster, both day-records recomputed, audit/source
preserved) but the DB does not independently enforce the invariants:

- **A1-F1 (P1, blocks=yes):** correction segments unbound to their
  group's predecessor. App SQL can insert a segment whose `supersedes_id`
  differs from `group.predecessor_interval_id`, is NULL, or is appended
  to a closed group months later — flipping unrelated intervals
  non-current or creating live rows superseding nothing.
- **A1-F2 = S7-P1-GROUP-OWN (P1, blocks=yes — adjudicated):** group-row
  predecessor unbound. Cross-employee AND cross-employer group inserts
  succeed via app SQL (live-proved P1a/P1d); global
  UNIQUE(predecessor_interval_id) then permanently wedges the victim's
  legitimate correction (Correct aborts); app cannot DELETE the poison
  row (42501). Zero-segment groups persistable. A rated blocks=yes, S7
  rated blocks=no ("loud, no silent payroll change"). **Adjudication:
  YES** — a permanent denial of correction requiring superuser cleanup,
  reachable from any app-role SQL mistake/compromise, sits on the
  first-payroll correction path; "loud" does not make it non-blocking.
- **A1-F3 (P1, blocks=yes):** partial supersede via the SERVICE path.
  Correcting one segment of a two-segment group creates a second group
  for that segment alone and leaves the sibling live — lineage forks
  (G1 half-stale, G2 new) with no warning, and `ListIntervals` never
  surfaces `correction_group_id` so operators cannot see membership.
  This is the one Area-A defect reachable through normal operator use,
  not just direct SQL.
- **A1-F4 (P1, blocks=yes):** `local_date` unvalidated at the DB layer.
  App SQL can insert a misdated segment allocating to the wrong day
  record and pay period; single-chain trigger stops double-allocation
  but not misallocation; no regen runs for direct SQL so affected
  drafts/approvals go silently stale.
- **A1-F5 (P1, blocks=yes):** backdated-evidence dodge. `created_at` on
  `attendance_interval`/`clock_event` is app-writable with no normalize
  trigger (unlike compensation/terms in 00010), so a backdated insert
  (created_at < approved_at) is invisible to `checkAttendanceConflict`
  and payroll proceeds on the old basis with no conflict flag.
- A4-G1 (P2, blocks=no): 2-segment same-sheet lifecycle
  (generate→submit→approve over one sheet) has zero coverage.
- A3-G1 (P3, blocks=no): −1s correction boundary (end 23:59:59) unpinned.

## 4. Duplicate-provenance findings (Area B, static)

- **B-CYC-01 (P1, blocks=yes, conditional):** concurrent direct-SQL A↔B
  cycle race. The disposition trigger is check-then-insert with no lock
  on the canonical row; two sessions inserting A→B and B→A can both
  commit under READ COMMITTED (distinct PKs, each snapshot sees the
  other's canonical undispositioned). Result: mutual exclusion, overlap
  auto-resolves, sibling endpoints orphaned, day permanently blocked
  with no supersede path (B-IMM-01). Service path is immune (employee
  advisory lock + ON CONFLICT). Sequential cycles are rejected in both
  layers and negatively tested. Never executed live (READ COMMITTED
  snapshot reasoning over inspected code only).
- B-BRK-01 (P2, blocks=no): near-duplicate break instants attach as late
  breaks while sibling in/out pairs overlap-block, consuming the break
  events (undispositionable) with no flag linking the split to the open
  overlap. Exact-instant dup breaks are safe (boundary-exclusive cover
  skips) but untested.
- B-VIS-01 (P2, blocks=no): dispositions are write-only. No CLI/REST/
  Review surface lists them; settled dups vanish from unpaired_events
  with no visible decision except raw SQL on the table or audit JSON.
- B-DOWN-01 = E-DOWN-01 (P2, blocks=no): populated Down drops
  `clock_event_disposition` unconditionally — see §7. Deduped under
  E-DOWN-01.
- B-CHAIN-01 (P3, blocks=no): chains A→B→C ARE formable (trigger checks
  only canonical-undispositioned; service checks canonDup only) whenever
  B is unconsumed — benign (same-instant members redundant, terminal
  canonical pairable) but contradicts the 00010 comment "chains and
  cycles are impossible". Apparent conflict with S7's "chain rejected"
  resolved: S7 probed the dup-OF-duplicate direction (rejected);
  B proved the canonical-BECOMES-duplicate direction (allowed). The
  second direction has no live probe but the static proof (trigger
  L66–69 + service L114–123 + missing test) stands.
- B-IMM-01 (P3, blocks=no): wrong dispositions have no supersede/
  correction workflow (app UPDATE/DELETE revoked, tested; admin path
  undocumented). Low standalone harm; makes B-CYC-01 unrecoverable.
- B-DAY-01 (P3, blocks=no): `overlapping_span` excluded from
  blockingCodes → day shows reviewed while Generate skips the employee
  for unpaired events. Diagnostic incoherence only.

## 5. Timezone findings (Area C, live)

Business-date attribution is **clean and live-proved**. Six P3s, all
rendering/docs/validation hygiene, none blocking:

- C-TZ-01: `span_date_mismatch` detail embeds `time.Local`-rendered
  instants → same cause can accumulate one open exception row per host
  TZ (dedup is on detail text).
- C-TZ-02: `interval.correct` audit before-state renders
  `time.Local` instants → cross-host audit diffs for the same instant.
- C-TZ-03: API/CLI contract §7 says vendor `timezone` drives
  attribution — contradicts implemented Amendment B separation
  (employment TZ sole authority; vendor TZ validated provenance only).
- C-TZ-04: `checkEligible` comment claims per-event declared-TZ
  conversion; code correctly uses employment location.
- C-TZ-05: programme TZ stored without IANA validation (today
  display-only, never read by attribution).
- C-TZ-06: Review/List `::text` rendering of TIMESTAMPTZ follows PG
  session TimeZone (display only; filtering uses explicit AT TIME ZONE).
- CLEAN (live C3 matrix): persisted UTC instants, local_dates,
  durations, break kinds, roster links, group segments, records,
  exception codes/dates/status ALL identical across adversarial host/session
  TZ; SGT/UTC/NY-expressed same-wall days attribute identically;
  offset-free rejected loudly; unknown TZ rejected; DST-gap interval
  exact; zero `time.Local` in non-test code; all SQL date conversions
  use explicit AT TIME ZONE; 13 TZ tests pass under both host TZ.

## 6. Compensation findings (Area D, live as real app)

- **D-RACE (P1, blocks=yes):** silent pay-basis change via txn-start
  timestamps. Comp txn starts BEFORE approval, commits AFTER;
  `created_at=now()` (txn-start, P16 proved exact) predates
  `approved_at=now()` so `checkCompConflict` (created_at > approved_at)
  returns zero rows. Live-proved: comp 01:31:14.53 < approved
  01:31:16.53, would_conflict=f. Sequential tests always order
  correctly, masking the race. `clock_timestamp()` only narrows the
  window; robust fix needs commit-ordered comparison
  (track_commit_timestamp / pg_xact_commit_timestamp(xmin) vs approval
  commit, or a lock barrier). Note: A1-F5's normalize trigger does NOT
  fix D-RACE (both use txn-start now()).
- **D-CONTAIN (P1, blocks=yes):** containment is service-only. App SQL
  inserts comp outside parent terms and overlapping same-code rows
  across terms (live P7/P8/P17: cross-terms SHIFT overlap persisted,
  backdated MEAL persisted, open children outlive closed parent, FWD
  starts after parent end). No readiness check models double/ambiguous
  basis. Exclusion is keyed on terms_id, not employee — a trigger is
  needed since exclusion cannot join terms.
- **D-ROUND (P1, blocks=yes — adjudicated):** silent 3dp rounding.
  `NUMERIC(12,2)` rounds on assignment: direct-SQL 199.999 stored as
  200.00 with no error (live P9). Service rejects 3dp but the DB
  silently alters money. Prior rated blocks=no. **Adjudication: YES** —
  §12 FAIL covers silent alteration, and consistency with the A-series
  (service-safe/DB-open gaps all rated yes) demands it. CHECK on a
  NUMERIC(12,2) column is ineffective; needs unconstrained NUMERIC +
  exact-scale CHECK (or domain), same for monthly_basic.
- D-AUDIT / Q9 (P2, blocks=no): direct-SQL comp INSERT is unaudited
  (audit only in service; live: 7 comp rows, 0 audits) and
  future-effective rows raise zero conflict — silent unaudited
  pay-basis change. UPDATE/DELETE impossible (42501). Sequential
  approved-period-effective inserts DO conflict-flag (5 rows); race
  window escapes both (D-RACE).
- D-NOCORRECT (P2, blocks=no): erroneous comp row has NO supported
  correction path — cannot end-date, supersede (overlap rejected), or
  remove; operators must misuse cross-terms insert (D-CONTAIN) or
  privileged unaudited edit. D1.6/D3 explicit-correction state
  therefore untestable/missing.
- D-AUDITFMT (P3, blocks=no): audit stores raw '100' vs persisted/list
  '100.00'; CompConflict carries code+date only, no amount — views
  cannot be mechanically reconciled.

## 7. Migration-00010 + rollback findings (Area E, static)

Forward Up state is sound; rollback is unexecuted and, in two populated
cases, silently destructive or one-way:

- **E-DOWN-01 = B-DOWN-01 (P1, blocks-first-payroll=no):** populated Down
  drops `clock_event_disposition` with no row-count guard (contrast the
  group guard at 00010 L785–788). All operator decisions lost;
  exceptions keep resolved=true citing events whose evidence is gone;
  after Down+Up, settled dups become pairing-eligible again. Raw events/
  exceptions/audits survive; decisions mirrored only in audit JSON
  (unrestorable relationally). Safe direction for payroll (re-blocks,
  never double-allocates) but audit trail destroyed. No Down test.
- **E-DOWN-02 (P1, blocks=no):** any correction history makes Down
  one-way. Single-segment groups pass the L787 guard, headers drop
  silently, and re-Up's `correction_group_present` constraint fails on
  surviving correction rows with NULL group id — schema stranded at
  00009 with no documented restriction. Decisions survive in
  attendance_correction + audit JSON but relational relink is
  impossible.
- E-DOWN-03 (P2, blocks=no): the entire 00010 Down section is
  unexecuted — no CLI subcommand, no core/database Down wrapper, no
  test invokes goose Down. E2/E3 acceptance (verified populated
  rollback policy) cannot pass as claimed.
- E-DOWN-04 (P3, blocks=no): post-Down `maintain_interval_currency`
  keeps Up's REVOKE ALL FROM PUBLIC (never re-granted) — more
  restrictive than pristine-00009; trigger firing unaffected.
- E-DOC-05 (P3, blocks=no): no documented rollback policy anywhere
  (zero rollback/migrate-down hits in docs); §E2 requires unsupported
  populated Down to fail pre-destruction AND be documented — currently
  neither for dispositions/single-segment groups.
- CLEAN (static): multi-segment Down provably fails loudly + atomically
  (shared supersedes_id + L787 index inside goose single-tx incl.
  version delete; no NO TRANSACTION in repo); Down grant-back/restore
  fidelity byte-exact except E-DOWN-04; us-column Down+Up backfill
  deterministic (MicrosOf truncation == numeric ::bigint truncation,
  links unchanged); approved timesheets/rosters/terms/comp/exceptions/
  audits untouched by Down.

## 8. Privilege findings (§7, live as real app)

- S7-P1-GROUP-OWN: merged into A1-F2 (§3). Live proof (P1a/P1b/P1c/P1d/
  P1f) stands as the executed confirmation of the static finding.
- S7-P2-FRESHFLAKE (P2, blocks=no): `go test ./...` on a clean server
  FAILS — employer TestMain connects to a not-yet-created base DB
  (FATAL 3D000) and parallel setups race role creation (23505). Suite
  green only on warm server/retry. Test harness only; zero product
  impact. Fix: CreateDatabase before Connect; tolerate 23505 / single
  global advisory lock for roles.
- S7-P2-MATRIX-OPEN (P2, blocks=no): 19-table privilege matrix is
  allowlist-only — a future table without revokes inherits default app
  full-DML (init-roles.sh L56–57) and the suite stays green. ~30 tables
  unpinned (ket_record/payslip/attendance_correction covered separately
  in freeze_test; rest by convention). Live: app INSERT/UPDATE/DELETE
  on unpinned `employer` all succeed (by design today). Fix: enumerate
  pg_tables, assert an explicit triple per table, fail on unknown.
- S7-P3-LEGACY-EXEC / Q13 (P3, blocks=no): app holds EXECUTE on 7 legacy
  SECURITY-INVOKER trigger fns (PUBLIC default never revoked) — direct
  call errors harmlessly ('only called as triggers'). Only intended
  DEFINER fns executable (close/transition); maintain_interval_currency
  direct call → 42501. Full-catalog probe: 17 migration-owned fns, 0
  event triggers, 0 DEFINER views, app TEMP revoked, pg_temp shadowing
  impossible. Fix: one-line REVOKE + extend catalog pins.
- S7-CLEAN (live): all other §7 invariants hold under direct app SQL
  with SQLSTATE + re-reads — disposition scoping/self-ref/chain-reject/
  42501 on write-back; comp UPDATE/DELETE 42501 + exclusion; currency
  normalization; supersession cross-employee rejection; line/sheet
  identity transplant rejection; roster-parent transition rejections;
  term/status history 42501 + close/transition positives with exact
  audit counts; DEFINER owner+search_path pins (10/10 watched pass).

## 9. Regression-test verification

WATCHED GREEN (primary-7, disposable PG16, this session): task lint 0,
task build 0, diff--check 0; `go test -count=1 ./...` exit 0 under
default/TZ=UTC/TZ=SGT (17 ok + 7 confirmed no-test-files, 0 skips
each); race on 5 pkgs exit 0; integrity 61 PASS/0 FAIL/1 intentional
SKIP; employer 5/0/0; provisioning 2/0/0 (incl. ephemeral clusters +
  scram cycle); seed 7/0/7; onboarding CLI+REST + parity + E5 reset (6)
  + D1 init-roles + §9C fresh-install-chain all pass with real-DB
  assertions (bodies read). hr+payroll suites pass on disposable PG16
  (primary-5). 13 TZ tests pass under both host TZ (primary-6).

CAVEATS (why green ≠ PASS): (a) FRESHFLAKE — green holds only on
warm server/retry, not from a clean server; (b) MATRIX-OPEN —
privilege suite stays green if a future table omits revokes;
(c) count drift (doc-only): report claims 38 Stage-C vs 40 observed,
6 employer tests vs 5 observed, §2 omits E2 textually though
implemented+tested (S1-REPORT-COUNT/S7-P3-COUNTS P3); (d) whole
categories have zero tests by construction: populated Down/Down+Up,
between-segments failure injection, 2-segment same-sheet lifecycle,
cross-employer disposition, break-duplicate dispositions,
concurrent dispositions, concurrent comp-vs-approval, comp
correction path (none exists); (e) Areas A/B/E verdicts rest on
inspected bodies + suite-green, not fresh persisted-state probes.

## 10. Structural Q1–Q15 dispositions

- Q1 multi-employee/employer group? NO for segments (member FK +
  employee FK + C3 same-employee trigger; app-role negative test).
  Caveat: group ROW predecessor unbound — A1-F2/S7-P1 (live-proved).
- Q2 two groups same predecessor? NO — global UNIQUE + direct-SQL and
  concurrent-winner tests (exactly one group).
- Q3 segment double-allocated? NO across employee/period (single-chain
  + employee-match triggers incl. UPDATE-widening; EffectiveApproved
  counts one on same-chain relink; cross-period double-alloc rejected
  in lineage tests). No fresh live probe; rests on bodies + green.
- Q4 partial group approve/reject? NO within a period — Submit/Approve/
  Reject are sheet-level only, one REG line per sheet; cross-period
  segments decide per-period by design (independent units). No live
  2-segment decision probe (A4-G1). Distinct from A1-F3 (partial
  SUPERSEDE via re-correction), which is YES and a P1.
- Q5 chains/cycles of dispositions? CHAINS yes (benign; B-CHAIN-01 P3).
  CYCLES: no sequentially (both layers + tests); YES via concurrent
  direct-SQL A↔B race (B-CYC-01 P1, reasoned not executed). Persisted:
  chain 2 rows + terminal pairable; cycle 2 rows + zero pairable +
  overlap resolved + permanent block.
- Q6 disposition hides canonical? Normal path NO (canonical untouched,
  pairable to approved sheet — asserted in TestDispositionDuplicate-
  ClearsBlock + e2e). Chain PARTIAL (middle hidden, terminal survives).
  Cycle YES fully (B-CYC-01). Consumed-canonical direct-SQL note:
  possible, inspected-harmless (citations persist, no re-pair), never
  live-probed, below P3 bar.
- Q7 same instant, different SG dates under host TZ? NO (live matrix).
- Q8 client TZ alters attribution? NO (live E001-vs-E003 identical).
- Q9 comp direct-SQL / silent basis change? YES for INSERT (unaudited;
  future-effective raises zero conflict). UPDATE/DELETE impossible
  (42501). Sequential retro-effective inserts conflict-flag; race
  window escapes (D-RACE). Q9 answer: yes, partially unguarded.
- Q10 populated rollback behavior? (Wording inferred; no child quotes
  prompt12.) Statically: dispositions silently dropped (E-DOWN-01),
  correction history makes Down one-way (E-DOWN-02), Down never
  executed (E-DOWN-03), no rollback doc (E-DOC-05). **Zero live
  evidence** — violates prompt12 E3 "no rollback claim not directly
  tested". Disposition: STATIC ONLY, unproven.
- Q11 approved-sheet consistency after late evidence? (Wording inferred
  from D-child: comp half + break half.) COMP half: approved rows never
  mutate (freeze trigger + comp touches only compensation_line);
  staleness flagged except D-RACE window and unmodelled D-CONTAIN
  ambiguity. BREAK half: **ZERO evidence anywhere** — explicitly
  out of scope for D, probed by no other child. Disposition: HALF
  ANSWERED / HALF UNRESOLVED.
- Q12 single-employer holds under concurrency? Service path YES
  (advisory xact lock + in-txn check; TestConcurrentCreateSingleWinner
  watched: 1 row, 1 audit). Direct-SQL extra employers possible (no DB
  constraint) but Resolve('') fails loudly and children are
  employer-keyed — no silent misattribution, design-consistent.
  Different-code concurrent pair: inspection only, no direct test.
- Q13 app EXECUTE outside least-privilege? YES on 7 legacy trigger fns
  but harmless (trigger-only; S7-P3-LEGACY-EXEC P3 hygiene).
- Q14 orphans after failed multi-segment? NO via service (single tx;
  TestCorrectRegenFailureAtomic groups=0/segs=0/orig-live). YES via app
  SQL (orphan group persistable — A1-F2, live-proved). Between-segments
  failure injection NEVER tested anywhere (regen/overlap-stage only).
- Q15 readiness counts original+correction? NO — EffectiveApproved only;
  conflict until correction approves (r2 conflict/r3 ready/effective=v2
  in TestCorrectionChainFullFlow). Caveat: backdated-evidence dodge
  (A1-F5) and D-RACE window both defeat the conflict gate.

## 11. Findings / risks / hypotheses (deduped)

Format: ID | severity | location | scenario | evidence | why missed |
fix | blocks-first-payroll.

P1 (blockers — DB-enforced guarantee fails; §12 bars PASS):

1. A1-F1 | P1 | 00010 L98–120 vs review.go L237–247 | App-SQL segment
   with mismatched/NULL supersedes_id or appended to closed group;
   flips unrelated interval or creates live row superseding nothing. |
   Full-00010 inspection (no binding CHECK/trigger); Correct() always
   sets both so service coherent; app INSERT granted (privilege matrix
   L167). Static (no-DB child). | Only positive binding asserted via
   Correct(); no negative direct-SQL segment test. | BEFORE INSERT
   trigger on attendance_interval (source='correction'): supersedes_id
   NOT NULL and = group.predecessor_interval_id (+ optional closed-group
   bar). | yes.
2. A1-F2 ⊕ S7-P1-GROUP-OWN (merged) | P1 | 00010 L98–108 (id-only
   predecessor FK) | Cross-employee/employer group insert succeeds,
   consumes UNIQUE slot, permanently wedges victim correction; app
   cannot DELETE poison (42501); zero-segment groups persist. | Static
   (A: no scope trigger, global UNIQUE, Correct inserts group first so
   any poison wedges service) + LIVE (S7 P1a/P1b/P1c/P1d/P1f as real
   app: cross-employee SUCCEEDED, victim UNIQUE violation, DELETE
   42501, cross-employer SUCCEEDED, incoherent segment SUCCEEDED;
   controls P1g/P1h/P1e pass). | No negative mismatched-group test;
   service-consistent shape only + INSERT-allowed pin. | Composite FK
   (employer_id, predecessor_interval_id)→interval(employer_id,id) +
   BEFORE INSERT trigger (group.employee_id = predecessor.employee_id;
   require predecessor is_current) + app-role negatives. | yes
   (adjudicated; S7's "no" rejected — permanent correction-DoS on the
   payroll path is blocking even when loud).
3. A1-F3 | P1 | review.go L71–150 (Correct loads one interval) +
   L529–562 (group id never surfaced) | Operator re-corrects one
   segment of a 2-segment group via normal service → second group for
   one segment, sibling left live, lineage forks, consumed as coherent.
   | End-to-end Correct inspection (no sibling lookup) + IntervalDetail
   without group column; re-correction tests use single-segment groups
   only. Static. | No test re-corrects one segment of a service-made
   multi-segment group. | Group-atomic re-correction (all live siblings
   in one call/tx) OR surface membership + explicit partial-supersede
   warning + regression test. | yes.
4. A1-F4 | P1 | 00002 L9–38 (DATE unconstrained) + attendance.go
   L816–830 (service-only split) | App-SQL misdated segment allocates
   to wrong day/period; drafts/approvals silently stale (no regen for
   direct SQL). | Full-00010 read (no date rule);
   TestSupersedeCrossPeriodAllowed documents cross-period successors
   allowed at DB with guard "in Correct service". Static. | All
   attribution tests via Correct(); no negative misdate test. |
   Trigger validating local_date vs interval_start in employment TZ
   (min: within [start,end] employment dates). | yes.
5. A1-F5 | P1 | readiness.go L391–438 (created_at > approved_at) + no
   normalize trigger for interval/event | Backdated app-SQL interval or
   clock event (created_at < approved_at) invisible to attendance
   conflict gate → payroll proceeds on stale basis, no flag. | readiness
   L354–472 (all three checks use created_at) vs 00010 normalize
   triggers only for comp/terms. Static. | Conflict tests create
   evidence via services (now()-stamped) only. | Mirror precedent:
   BEFORE INSERT OR UPDATE now() trigger on
   attendance_interval/clock_event.created_at (+ backdate test). Does
   NOT fix D-RACE. | yes.
6. B-CYC-01 | P1 | 00010 L39–72 (check-then-insert, no canonical lock)
   | Concurrent app-SQL A→B + B→A both commit → mutual exclusion,
   orphaned siblings, permanent block, no remedy. | Trigger + advisory
   order inspection (service serializes via lockEmployee,
   disposition.go L98–131; direct SQL does not). Reasoned under READ
   COMMITTED; NOT executed live. | Zero concurrent-disposition tests
   (all TestConcurrent* cover other ops). | Lock both event rows FOR
   UPDATE in id-sorted order (or pair advisory xact lock) before the
   undispositioned check + adversarial concurrent test. | yes
   (conditional: needs concurrent direct app SQL; service immune).
7. D-RACE | P1 | readiness.go L444–472 + timesheet.go L746
   (approved_at=now()) + 00010 L148–160 (created_at=now()) | Comp txn
   spans approval → created_at < approved_at → conflict gate returns
   zero rows → silent pay-basis change. | LIVE as real app (P16
   created_at=now() exact; race: comp 01:31:14.53 < approved
   01:31:16.53, would_conflict=f, 0 rows; scripts /tmp/dreview-*).
   | All comp tests sequential (now() ordering always correct);
   concurrency tests only for terms/status. | Commit-ordered comparison
   (track_commit_timestamp + pg_xact_commit_timestamp(xmin) vs approval
   commit, or lock barrier); clock_timestamp() narrows only. | yes.
8. D-CONTAIN | P1 | hr.go L517–532 (service-only) + 00004 L249–256
   (exclusion on terms_id) | App-SQL comp outside parent terms /
   overlapping same-code across terms → double/ambiguous basis no
   check models; children outlive closed parent. | LIVE as real app
   (P7/P8/P17 persists + re-reads; only trigger is created_at
   normalize). | Containment test is service-only; readiness comp tests
   single-terms; no cross-terms test. | DB range containment + employee
   same-code overlap trigger + bound/validate children in
   close_employment_terms. | yes.
9. D-ROUND | P1 | 00001 L96 NUMERIC(12,2) + hr.go L289 (service-only
   scale check) | App-SQL 199.999 → stored 200.00, no error — silent
   money alteration. | LIVE as real app (P9 re-read 200.00; P10 CHECK
   rejects negatives; service rejects 3dp per TestCompensationExact).
   | Exact-scale test asserts service rejection only, never direct-SQL
   re-read. | Unconstrained NUMERIC + exact-scale CHECK
   (amount = trunc(amount,2)) or domain; same for monthly_basic. | yes
   (adjudicated; prior "no" rejected per §12 silent-alteration FAIL +
   A-series consistency).
10. E-DOWN-01 (= B-DOWN-01, merged) | P1 | 00010 L805–817 | Populated
    Down drops all disposition decisions silently (no guard; outbound
    FKs only); post Down+Up settled dups re-eligible; audit-JSON mirror
    unrestorable. | Full Up/Down inspection + zero inbound FKs (repo
    migration search) + exclusion-filter sites; no Down execution
    anywhere. Static. | No populated-Down test; suites cover Up only.
    | Pre-destruction guard (RAISE when rows exist, mirror L785 group
    comment) + documented restriction, or declare populated Down
    unsupported. | no (forward-only payroll never runs Down; trap on
    deliberate raw-goose rollback only).
11. E-DOWN-02 | P1 | 00010 Up L119–120 vs Down L787–803 | Any correction
    history → Down one-way, re-Up impossible (NULL-group correction
    rows violate re-added constraint; schema stranded at 00009). |
    Up-writer (review.go L220–247) + Down + 00009 L162–163 guard-scope
    proof; zero rollback-policy docs. Static. | No Down/Down+Up test.
    | Guard Down against ANY group rows + document one-way/restriction;
    no silent relink on re-Up. | no (forward-only unaffected).

P2 (non-blocking gaps; fix before scale, several before first payroll
as hygiene):

12. A4-G1 | P2 | timesheet.go L387–437 vs test bodies | 2-segment
    same-sheet lifecycle untested (summed-hours/one-line atomicity by
    reading only). | Every cited test body inspected (no Generate/
    Submit/Approve over 2-seg same-sheet). | Per-scenario coverage
    growth without a combined test. | One test: same-period
    cross-midnight group → Generate (one line, summed hours, both
    links) → Submit → Approve → EffectiveApproved + reject/reissue
    variant. | no.
13. B-BRK-01 | P2 | attendance.go L954–1152 attachLateBreaks | Near-dup
    break instants consumed while siblings overlap-blocked → undisposi-
    tionable, no linking flag. | pairSpans/cover/tiles + refsJSON +
    consumed-check inspection; zero break-disposition tests. Static. |
    spanCSV helper has no breaks; overlap tests never include break
    endpoints. | Exact-instant + near-dup break tests; consider
    evidence_conflict hold when attaching under open overlap. | no.
14. B-VIS-01 | P2 | review.go L409–470, httpapi.go L540–555, CLI
    attendance.go L78–96 | Dispositions write-only; no list surface;
    settled dups vanish silently. | Review/CLI/REST/openapi inspection
    (POST only); e2e reads via raw SQL. | SQL-only read normalized by
    e2e test. | Surface in ReviewAttendance + CLI/REST list. | no.
15. D-AUDIT | P2 | hr.go L541–543 (service-only audit), no audit
    trigger | Direct-SQL comp unaudited; future-effective raises zero
    conflict. | LIVE (P13: 7 rows, 0 audits, 0 future conflicts). |
    Audit asserted on service path only. | INSERT trigger auto-auditing
    comp writes, or route through controlled function. | no.
16. D-NOCORRECT | P2 | hr.go L475 (only mutation) + 00010 L167 + 00004
    L249 | No supported comp correction path (end-date/supersede/
    remove all impossible). | Code search (only AddCompensation; no
    retro check unlike SetTerms) + live P6 exclusion rejection. | No
    test attempts comp correction. | Controlled close/supersede
    function + audit + readiness explicit-correction state. | no.
17. E-DOWN-03 | P2 | migrate.go (up/status only), 00010 L703–817 | Down
    wholly unexecuted; E2/E3 acceptance cannot pass. | Full migrate.go/
    migrate_test.go + repo regex for Down execution (none). | Suite
    never targeted Down. | Disposable-PG16 Down probes (empty/
    populated/guard/Down+Up/version) or declare unsupported. | no.
18. S7-P2-FRESHFLAKE | P2 | employer_test.go L46–55, dbutil.go L205–213
    | Clean-server `go test ./...` fails (FATAL 3D000 + 23505 role
    race); green only warm/retry. | Fresh-cluster reproductions (virgin
    dir: exit=1, only employer FATAL). | Warm shared server reuse
    masked races. | CreateDatabase before Connect; tolerate 23505 /
    global role lock. | no.
19. S7-P2-MATRIX-OPEN | P2 | privilege_test.go L150–190 | 19-table
    matrix stays green if a future table omits revokes (~30 unpinned;
    app full-DML on `employer` live-proved by design today). |
    Migration table enumeration vs 19 rows + default grants
    (init-roles.sh L56–57). | Fixed allowlist at C9 time. | Enumerate
    pg_tables, explicit triple per table, fail on unknown. | no.

P3 (doc/comment/hygiene; no code-behavior risk except C-TZ-01's
exception-row growth):

20. A3-G1, B-CHAIN-01, B-IMM-01, B-DAY-01, C-TZ-01…C-TZ-06, D-AUDITFMT,
    E-DOWN-04, E-DOC-05, S1-BASELINE, S1-REPORT-COUNT, S7-P3-LEGACY-EXEC,
    S7-P3-COUNTS — as filed in §§3–8; fixes are comment/doc/wording,
    one-line REVOKE, display-format normalization (.UTC()/employment-
    zone), IANA validation, audit-format canonicalization, count
    corrections, and E2 line restoration. All blocks-first-payroll=no.
    C-TZ-01 (per-TZ duplicate exception rows) is the only P3 with a
    (minor, visible) persisted-state effect.

Hypotheses explicitly NOT filed (inspected-harmless or contradicted):
consumed-canonical direct-SQL disposition (citations persist, no
re-pair — below bar); cross-employer disposition rejection
(dual-enforced, inspected-sound, untested — unresolved not finding);
genuine-overlap-as-dup (rejected both layers, tested); pg_temp
shadowing (TEMP revoked, no creatable schema); money via float64
(zero money-float hits; all floats are durations/display/config).

## 12. Clean areas (with evidence basis)

- Correction-group service path: atomic single-tx Correct + advisory
  lock + liveness recheck; single-successor UNIQUE incl. direct-SQL +
  concurrent-winner tests; member/employee/period triggers incl.
  app-role negatives; cycle/self/cross-employee supersession; forced
  is_current + live-no-overlap; insert-only revokes; 8/9 attribution
  cases with per-segment roster + NULL/exception + dual records +
  audit; single-line-per-sheet atomic decisions; single effective
  approval (static bodies + green suites; residual: A2 live replay and
  A4-G1/A3-G1 untested).
- Duplicate provenance core: canonical reference EXISTS (NOT NULL + FK
  + self CHECK) with same-employer/employee/kind/exact-instant +
  canonical-undispositioned in BOTH layers; dup excluded at 3 sites
  post-disposition; canonical pairable to APPROVED sheet (asserted
  9h single-interval line, submit/approve, raw-preserved counts in
  TestDispositionDuplicateClearsBlock + e2e); scenarios 1/2/3/4/6/8/10
  tested incl. negatives; raw-event + disposition immutability
  matrix-asserted; genuine-overlap rejection + continued block tested;
  CLI+REST same-service (static).
- Timezone business logic: employment TZ sole authority; vendor/programme
  TZ never applied; offset-aware parse + loud rejections; DST-safe;
  UTC preserved; zero time.Local; explicit AT TIME ZONE everywhere
  (LIVE matrix + 13 tests × 2 host TZ).
- Compensation service path + DB core: UPDATE/DELETE/identity/date/
  re-parent all 42501 + re-reads; same-terms overlap 23P01; negative
  CHECK 23514; sequential created_at normalize; cross-employer FK
  23503; sequential retro conflict rows; future-effective correctly
  silent; string-only money API→CLI→service with exact NUMERIC
  storage; reissue+approve clears conflict (LIVE as real app + suites).
- Forward migration/privilege Up state: multi-segment Down loud-fail
  design; goose single-tx atomicity; E4 privilege model; grant
  fidelity; us recompute determinism (static); all §7 app-role probes
  green except group-ownership (LIVE); full catalog EXECUTE + TEMP +
  DEFINER pins (LIVE); single-employer service concurrency (WATCHED
  test); Q12/Q13 dispositions (LIVE).
- Statutory containment: SG-MOM-OT-001 manifest BLOCKED, zero
  rules/*.go, zero overtime/cpf/netpay/payslip hits repo-wide, no
  money float64, OT-gated approval impossible while BLOCKED
  (grep + inspection). No OT/CPF/SDL/SHG-math/net-pay exists to be
  corrupted — payroll-readiness failures in this review concern the
  readiness GATE, not computed statutory amounts.
- Regression suites: §9 watched-green inventory (LIVE, warm server).

## 13. Four verdicts

1. Functional correctness (service paths): **CONDITIONS**. Happy paths
   (correct incl. midnight/period/roster/break/superseded, disposition
   incl. e2e-to-approval, comp/terms/status flows, concurrency
   single-winners, atomicity rollbacks) are correct and tested. But
   A1-F3 (service-reachable partial supersede with invisible
   membership) and D-NOCORRECT (no comp correction path at all) are
   functional defects requiring restriction + fix. Condition: forbid
   re-correcting one segment of a multi-segment group (re-correct all
   live siblings together, manually verified) until group-atomicity or
   surfacing ships; comp errors need a documented privileged workaround
   with audit until a controlled function exists.
2. Database integrity (app-role enforcement): **FAIL**. Nine silent or
   wedging gaps under legitimate app credentials: A1-F1 (lineage
   corruption), A1-F2 (permanent correction-DoS + incoherent lineage),
   A1-F4 (silent misallocation), A1-F5 (silent stale approval),
   B-CYC-01 (permanent block via race), D-RACE (silent pay-basis
   change), D-CONTAIN (double/ambiguous basis), D-ROUND (silent money
   rounding). Green suites do not compensate — §12 explicitly bars
   PASS on suites alone.
3. CLI/API operability: **PASS**. CLI and API invoke the same services;
   3 new routes wired + OpenAPI + contract tests; onboarding CLI+REST +
   parity suites green (watched); disposition POST-only and review
   visibility gaps (B-VIS-01) are P2 operability debt, not invocation
   failures. Harness flake (FRESHFLAKE) is test-only.
4. Payroll-readiness integrity: **FAIL**. Assess correctly gates
   sequential/visible staleness (conflict-until-reapproved proved),
   but three holes let payroll proceed on a wrong-or-stale basis with
   no flag: A1-F5 backdate dodge, D-RACE race window, D-CONTAIN
   unmodelled ambiguity — plus unaudited direct comp (D-AUDIT) and no
   comp correction state (D-NOCORRECT/D1.6/D3). Mitigating fact: no
   statutory payroll MATH exists yet (OT BLOCKED), so the failure is
   confined to the readiness gate and pay-basis inputs, not computed
   payslips.

## 14. Overall PASS / CONDITIONS / FAIL (§12)

**FAIL.** §12 PASS requires DB-enforced guarantees under app. This
review proves, by live execution (D-RACE, D-CONTAIN, D-ROUND,
S7-P1) and by complete static inspection where live was unavailable
(A1-F1/F4/F5, B-CYC-01), that the app role can silently corrupt or
silently stale payroll inputs (lineage, dates, money, conflict gate)
or permanently wedge correction/disposition with no supported remedy.
CONDITIONS is insufficient: the defects are not loud-only
restrictions — several are silent (no error, no conflict, no audit),
which is the §12 FAIL case. The FAIL is narrow and repairable: service
paths are sound, fixes are small and trigger-local (except D-RACE's
commit-ordering, which needs design), and no statutory amounts exist
yet to have been miscomputed.

## 15. Smallest remediation (deduped, ordered)

1. One forward migration (00011): (a) group↔predecessor scope
   (composite FK + employee-match + is_current-at-insert trigger —
   closes A1-F2/S7-P1); (b) segment↔group binding trigger
   (correction segments: supersedes_id NOT NULL = group predecessor —
   closes A1-F1); (c) local_date-vs-interval_start employment-TZ
   trigger (closes A1-F4); (d) created_at now() triggers on
   attendance_interval + clock_event (closes A1-F5); (e) comp
   containment + employee same-code overlap triggers + child bounds in
   close_employment_terms (closes D-CONTAIN); (f) exact-scale money
   domain (unconstrained NUMERIC + trunc CHECK) for amount_monthly +
   monthly_basic (closes D-ROUND); (g) canonical-pair lock in
   disposition trigger (closes B-CYC-01); (h) populated-Down guards
   for dispositions + any-group-rows (closes E-DOWN-01/02) + one-line
   legacy-EXECUTE REVOKE (S7-P3-LEGACY-EXEC). Each with an app-role
   negative test mirroring the live probes in §§3–8.
2. D-RACE design fix (separate, harder): commit-ordered conflict
   comparison (track_commit_timestamp / pg_xact_commit_timestamp(xmin)
   vs approval commit) or a comp-vs-approval lock barrier; add a
   concurrent comp-vs-approval test. A1-F5's trigger is necessary but
   NOT sufficient for this item.
3. A1-F3 service fix: group-atomic re-correction (or surfaced
   membership + explicit warning) + 2-segment re-correction regression
   test; fold in A4-G1 same-sheet lifecycle test + A3-G1 −1s pin.
4. Comp lifecycle: INSERT auto-audit trigger (D-AUDIT) + controlled
   close/supersede function with audit + readiness explicit-correction
   state (D-NOCORRECT/D1.6/D3) + audit canonical persisted value
   (D-AUDITFMT).
5. Rollback/test-harness/docs: declare populated Down unsupported
   (E-DOC-05) OR add disposable-PG16 Down probes incl. Down+Up and
   version-after-failure (E-DOWN-03); fix FRESHFLAKE (CreateDatabase +
   23505-tolerant roles); fail-closed privilege enumeration
   (MATRIX-OPEN); correct B-CHAIN-01 comment (both chain directions
   pinned), C-TZ-03 contract + C-TZ-04 comment, C-TZ-01/02/06 display
   normalization, C-TZ-05 programme validation, B-DAY-01
   blockingCodes/doc choice, baseline + count restatements
   (S1-BASELINE, S1-REPORT-COUNT/S7-P3-COUNTS incl. E2 line), B-VIS-01
   surfacing, B-IMM-01 finality doc-or-supersede decision.
6. Re-verify: clean-server `go test ./...` (proves FRESHFLAKE fixed),
   full suite × 3 TZ + race (as §9), plus the new adversarial tests
   from items 1–4 — then re-render §13/§14. Item 2 + item 1(a–g) are
   the PASS gate; items 3–5 are required for CONDITIONS-free operation
   but item 3 (A1-F3) independently sustains FAIL/conditions until
   closed.

## 16. Unverified / environment-dependent items

- Q10 populated rollback: zero live runs (E1 10-step, E2 9-artifact
  Down, E3 six-probe matrix); all rollback claims static-only.
- Q11 break half: zero evidence anywhere (approved-sheet consistency
  after late break never probed).
- A2 live 7-step replay + Q14 between-segments injection: never
  executed; rollback proven only for regen/overlap-stage failures.
- A4 2-segment same-sheet lifecycle + A3 −1s + per-case persisted
  matrix: untested (code-reading only).
- B scenario 5 cross-employer, scenario 7 break dups (both
  sub-cases), scenario 9 concurrent service dispositions, Q5 A↔B race
  execution, Q5 second chain direction live probe, consumed-canonical
  direct-SQL harmlessness, Q6/Q3/Q4 fresh live probes: all
  inspection-only.
- C2 offset-free-with-context success path: no fixture (only NEG2
  rejection evidenced).
- D1 seven service scenarios: bodies opened, not re-driven live; D-RACE
  via concurrent Go-service calls not executed (SQL-level only);
  privileged-role comp edits unprobed (noted by design).
- §1 item 5 (00001–00009 untouched): unverifiable from git (no
  snapshot); needs file-novelty/mtime/content restatement.
- §8 clean-server green: fails as-run (FRESHFLAKE); counts/method drift
  doc-only; matrix-open fail-closed missing; different-code employer
  pair inspection-only; Compose path + serve smoke not run.
- Operator runbooks outside the repo not searched (Down may be
  documented out-of-band — U4).
- Static children (A/B/E/S1) never executed the suites they cite; live
  children (C/D/§7/§8) did not re-prove A/B/E persisted states. No
  child rendered the four §11 verdicts or the §12 overall before this
  synthesis.
