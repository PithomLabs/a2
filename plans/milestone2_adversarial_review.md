# Milestone 2 Final Adversarial Review — Payroll Governance, Snapshot Integrity and Reproducibility

**Overall decision: FAIL — NOT READY for the first payroll calculation.**

**Decisive question:** can one payroll run today capture a complete and coherent employee
population, resolve the correct historical inputs, pin verified rule versions, preserve an
immutable snapshot, and prevent unauthorized lifecycle transitions? **No.** Two P0s
(stale-consistent capture/validation; fabricated-snapshot INSERT as `app`) plus blocking P1s
in compensation, SHG, pack trust-root, and direct-validation bypass break every leg of that
guarantee. Minimum corrections only are listed in §16; no redesign is proposed.

- Review baseline: HEAD `304fec88daaadae493309bfbad65f7f472d89c62` (short `304fec8`, release 0.1.3).
- Report reviewed: `plans/milestone2_remediation_report.md` (claims P1–P6 PASS).
- Method: read-only review. No application source, migration, test, contract, config, or doc
  modified. All live probes ran on disposable PostgreSQL 16 clusters/databases as the real
  `app` role; repo tree verified untouched/bit-identical before/after live work.
- Evidence base: 7 prior workflow results (scope-1 baseline map; Areas A/B, C, D, E/F, G/H;
  independent critic). Every finding below carries inspected evidence or is named as a gap.

---

## 1. Git baseline and review scope

- `git rev-parse HEAD` = `304fec88daaadae493309bfbad65f7f472d89c62`; matches reported M2 baseline.
- `git status --short`: 28 tracked-modified (matches report); 20 untracked = 19 M2 files +
  `docs/writeup3.md` (55 KB user prose, excluded from scope); zero deleted/renamed.
- `f435158` explained: superseded local pre-pull commit (same message as remote `eb0cec6`);
  reflog shows commit-initial then pull picked `eb0cec6`+`304fec8`; object unreachable from refs.
- `bchat/` untouched. `git diff --check` clean.
- Migrations `00001`–`00011`: byte-identical verified (`sha256sum -c`, empty `git diff HEAD`).
- `00012_m2_governance.sql` (1521 lines) fully read: SHG 42–433, RULES 435–696,
  SNAPSHOT 698–1138, RUNLIFE 1140–1296, guarded Down 1298–1521. All claimed objects present.
- Scope-1 work was reads + `git`/`grep`/`go list` only (no test/build/lint/DB in that slice);
  live Areas D/E/F/C/G/H executed the §11 battery and targeted races on disposable PG16.
- Environment notes: Go 1.27.3, task 3.48.0, psql 16.14 / PostgreSQL 16 (server bin not on
  PATH in one env; reviewers used explicit initdb/postgres paths, e.g. `/tmp/pg16data:55433`,
  `/tmp/drev_pgdata:55433`, `/tmp/m2rev6_pg16:55429`). `go list` requires
  `GOWORK=off` + `GOCACHE=/tmp/gocache` (parent `go.work` otherwise captures the module).

## 2. Actual commands executed

Scope-1 (read-only map): `git status --short`, `git rev-parse HEAD`, `git reflog`,
`git fsck`, `git ls-tree`, `sha256sum -c /tmp/m2-baseline-sums.txt`,
`git diff HEAD -- migrations`, `GOWORK=off go list`, file-body reads of every cited
implementation/migration/test; `rg` used only to locate. No test, build, lint, or DB command.

Live areas (disposable PG16 only; server+tests co-invoked per shell where the sandbox required it):

- Area C: `GOWORK=off GOCACHE=/tmp go test -run 'TestVerify|TestDigest' ./internal/rules/`
  (ok — proves fabricated-digest/failing-golden gap is untested, not failing); live cluster
  `c3rev_areac_01` with 12 migrations as `migration_admin`, 29/29 app-denial probes,
  repin/supersede/label-forgery/racer probes as real `app`.
- Area D: disposable PG16.15 `127.0.0.1:55433` (`initdb /tmp/drev_pgdata`, DBs
  `drev_pure`/`drev_e2e` dropped+recreated per run, 12 migrations via `migrate up` as
  `migration_admin`), real services as `app` from a `/tmp/drevrepo` copy; pure modes
  (lockfirst/readfirst/settxn/committed-before/sessionlock) + writer-first T1/T3/T4/T5, T2
  validation-level P0, T6 positive controls, T8 40001 proof, T7 privilege matrix; `pg_locks`
  NOT-granted proof per run. Repo HEAD `304fec8` bit-identical before/after.
- Area E/F: disposable PG16.15 clusters `/tmp/p16ef-b..e`, DBs `p16ef_live1..7` +
  `p16ef_gotest2/3`, `current_user=app` verified; fabricated-snapshot, direct-validate,
  vacuous-validate, pin-tamper, UPDATE/DELETE-denial, drift (8 categories), lifecycle batteries;
  committed subset 22 payroll + 3 integrity tests green.
- Area G/H: disposable PG16.15 `/tmp/m2rev6_pg16:55429` (trust, stopped after each run);
  `gofmt`, `go vet`, `git diff --check`, no-`lib/pq`, `task build`, `go test -count=1 ./...`
  default + `TZ=UTC` + `TZ=Asia/Singapore` (fresh base names per run), race subsets
  (attendance/timesheet/hr/payroll/wfm, 0 DATA RACE), provisioning twice, `TestDown*`,
  M2-pattern 125 pass/0 fail/1 intentional skip, contract + CLI/API parity, compbasis carve +
  seed suites; stale-schema proof DB `m2rev6stale2`; G1–G7 `payroll_line` battery as `app`.
- `task lint` / `task build` / full `./...` battery were executed by the live areas (§10), not
  by scope 1.

## 3. Compensation resolver findings

### F-A1-WIDEN — P1 — blocks first calculation: YES

- Location: `internal/compbasis/resolve.go:82-92` (period-overlap load filter),
  `:140-152` (dangling branch); `migrations/core/00011_m13_integrity.sql:672-692`.
- Scenario: legal widening correction permanently bricks resolution. Orig SHIFT 100.00
  2026-01-01..2026-06-30, then `CorrectCompensation` succ 120.00 2026-01-01 open-ended passes
  the overlap trigger (direct predecessor excluded, containment holds). An August monthly query
  loads `{succ}` only (orig filtered by `effective_to < period start`); carve reports
  dangling-supersedes; readiness `CompConflicts` blocks forever. Every further tip correction
  keeps succ loaded with its parent missing — no service action restores resolvability.
- Evidence: exact-query + exact-branch inspection (deterministic); existing tests cover only
  open-parent shapes (B9 future-correction, narrowing) — no widening/disjoint-parent resolve
  test. No live DB in Area A/B (client-only PG, Docker denied); reachability + exact
  conflict strings still need the §16 live probe.
- Minimum fix: separate chain-load from span-carve — load transitive `supersedes` ancestors
  regardless of period overlap for integrity/lineage; carve spans from overlapping rows only.
  Add widening + disjoint-successor resolve tests.

### F-A1-FREETAIL — P1 — blocks: YES

- Location: `internal/compbasis/resolve.go:232-244` + `:289-317`; `00011:677-692`;
  `tests/integrity/comp_lifecycle_test.go:220-232`.
- Scenario: write-allowed state that can never resolve. Orig open-ended → narrowed succ
  01-01..06-30 → fresh add 07-01 open (test asserts all 3 rows persist). August loads
  `{orig, fresh}`: orig keeps a surviving tail span while the fresh chain claims the same
  dates → permanent `multiple-effective` conflict via service. Loud (never a wrong payment)
  but permanently blocking by chain induction.
- Evidence: code inspection + repo's own tests pinning both halves (3-row write test never
  resolves; `TestResolveNarrowingSegments` asserts tail survival). Live probe still required.
- Minimum fix: align write-time liveness with resolver ownership (overlap check must account
  for predecessor surviving spans outside successor ranges, or forbid fresh adds where a
  predecessor tail survives and require correcting the tip). Add resolve-after-freed-tail test.

### F-A1-MIDPERIOD (= S1-P3-MIDPERIOD, deduped) — P1 — blocks: YES

- Location: `internal/payroll/readiness.go:220-230`; `internal/payroll/snapshot.go:1020-1045`,
  `:860-892`; `internal/payroll/validate.go:412-444`; `internal/compbasis/resolve.go:72`.
- Scenario: mid-period change silently INCLUDED for the monthly profile. Orig 100 01-01 open,
  correction succ 120 effective 08-15 (legal). August carve yields 2 segments
  (08-01..08-14, 08-15..08-31). Nothing checks segment count: readiness blocks only on
  `BasisConflict`, snapshot records `included` with both segments, validation sees no blocker.
  A run can validate on a 2-segment month with no proration semantics — violates A1's required
  explicit deterministic blocker. Report §3/§12 claims a "blocks validation on more than one
  segment / mid-period blocker" guard + test; repo-wide search finds neither (only 2 incidental
  comments). `TestResolveNarrowingSegments` asserts `len==2` SUCCESS.
- Evidence: bodies opened (`resolve.go:1-366`, `readiness.go:1-631`, `snapshot.go:851-895`,
  12-test resolve matrix, 4 carve tests, `readiness_resolve_test.go` real-conflict proof as app).
- Minimum fix: readiness + snapshot decision emit deterministic `COMP_MID_PERIOD_CHANGE`
  blocker when any classification resolves to more than one segment in the period (until
  proration is specified and verified). Add mid-period readiness + snapshot + validation tests.
  Correct report §3/§12.

### F-A2-TERMSLINK — P2 — blocks: NO

- Location: `internal/payroll/snapshot.go:58-65` + `:1029-1037`; `internal/compbasis/resolve.go:23-33`.
- Scenario: snapshot drops per-segment provenance. Resolver outputs RecordID/TermsID/
  Supersedes/Lineage per segment; `SnapSegment` keeps only Code/Amount/From/To/Lineage —
  TermsID dropped, RecordID only implicit as lineage tail. Cross-terms adjacent lines in a
  boundary-spanning period lose terms attribution. Drift still detected via section hashes;
  amounts absolute — provenance/explainability gap, not a wrong-payment risk.
- Evidence: struct-to-mapping comparison; no test asserts terms linkage in snapshots.
- Minimum fix: add `terms_id` + `record_id` (+ explicit `supersedes_id`) to `SnapSegment` and
  the canonical comp form; extend the superseded-comp snapshot test to assert the pin.

### Single-resolver integrity — CLEAN

- One authoritative `ResolveEffectiveCompensation` (`resolve.go:72`); callers readiness +
  snapshot only; no second implementation (4 files reference it incl. tests). Chain integrity,
  successor-range subtraction, terms containment, multiple-effective conflict all in
  `carveSegments:129-341`. No summing, no ignored segment, no in-term violation, no
  predecessor+successor double-payable; no `LIMIT 1` selection issue (LIMIT 1 only in
  ORDER-BY existence-semantics conflict-message queries).

## 4. SHG lifecycle findings

### F-B1-ACTIVE — P1 — blocks: YES (decision-gated)

- Location: `internal/payroll/readiness.go:415-419`; `00012:52-59`; `internal/hr/shg.go:95-99`;
  `internal/payroll/readiness_shg_test.go:115-118`.
- Scenario: every `active` instruction is active-without-verification yet counts as complete.
  No verification path to `active` exists (`VerifySHG` mints `verified_none` only; CHECK forbids
  any evidence triple on `active`); direct-active creation needs only `source_declaration`;
  readiness returns `covered` for any active tip. Contradicts B1 "Active without verification
  blocks readiness" and the 8-state model (only 7 modeled — `active` has no verified/unverified
  split). Stale comment (`hr_test.go:354-357`) suggests the model evolved past B1's wording.
- Evidence: code + matrix test pinning the behavior. Owner intent unresolved.
- Minimum fix (explicit pre-calculation decision): EITHER require verification evidence for
  `active` (pending→active confirmation carrying evidence; disallow direct-active or mark it
  pending) OR amend B1 with pinned rationale documenting declaration-plus-audit as the active
  evidence standard. Either way, record the decision before the first calculation.

### F-B2-PARTIAL — P1 — blocks: YES

- Location: `internal/payroll/readiness.go:378-426`; `internal/payroll/snapshot.go:1049-1081`.
- Scenario: partial-period tips count as full-period coverage (overlap-only filter +
  successorless tips + existential any-active/`verified_none`): (a) active tip starting 08-15
  covers all of August; (b) tip ending 08-15 covers; (c) pending 08-01 superseded by
  `verified_none` 08-15 hides the uncovered first half; (d) mid-period gap between two
  same-fund tips invisible. Violates B2 "partial-period never covers full period" and the
  "fully evidenced coverage or precise blocker" rule.
- Evidence: exact query + branches; matrix tests use 01-01-open ranges only — no
  mid-start/mid-end/gap/successor-date coverage test. Live probe still required.
- Minimum fix: require the union of live active/`verified_none` tips to fully cover
  `[periodStart, periodEnd]` (simplest: ≥1 tip spanning the whole period), else a precise
  blocker naming the uncovered span. Add all 9 B2 coverage cases as tests.

### X-FUND (S1-P4-CROSSFUND, reconciled) — P1 — blocks: YES

- Location: `internal/payroll/readiness.go:378-426`; `snapshot.go:1049-1081`.
- Scenario: report §4 claims "cross-fund multiples are rejected as ambiguous for the first
  profile"; code does the opposite — coverage returns `covered` on the FIRST active/
  `verified_none` tip and ignores other funds, so conflicting multi-fund instructions silently
  pass as complete. `TestSHGFundCoexistence` proves coexistence allowed.
- Evidence: bodies opened (`shg.go:1-325`, readiness, snapshot, 7 lifecycle tests, SHG matrix,
  `shg_governance_test.go:1-163`). Area B noted the schema carries no ethnicity/fund-eligibility
  data so no stricter rule is implementable today, and accepted existential coverage — but the
  multi-fund combinations are untested, so the acceptance is unverified. DECISION: treat as P1
  blocking until either (i) validation/snapshot-covered blocks on cross-fund multiples with a
  coded blocker, or (ii) any-active-covers semantics are pinned with evidence + a live
  multi-fund readiness matrix. Correct report §4.
- Minimum fix: per decision above; either way add live multi-fund readiness tests.

### SHG rejection evidence (F-B4-SHG-EVIDENCE P3 + S1-P4-AUDIT-WORDING P2) — blocks: NO

- Location: `tests/integrity/shg_governance_test.go` (zero audit assertions);
  `internal/hr/shg_lifecycle_test.go`; cf. `comp_lifecycle_test.go`.
- Scenario: all 7 B4 app-role attempts are mechanically blocked (direct INSERT/UPDATE/DELETE
  revoked at `00012:381-382`, 42501 asserted as real app; 5 function-level rejections asserted
  via app-DSN service tests) and audit atomicity holds by single-`SECURITY DEFINER`-transaction
  construction — but unlike comp tests, no SHG rejection test re-reads persisted state or
  asserts zero audit rows. Gaps: no rejected-write-leaves-no-audit per rejection;
  no direct app-function negatives for cross-employee/cross-employer SHG succession (guards
  inspected at `00012:99-103,259-262,339-342` + FK, untested); no SHG post-snapshot drift test
  (`snapshot.go:589-594` inspected only); no backdated-add-affecting-approved-period test
  (mechanism: `created_at` normalize `00012:144-159` + dual conflict legs
  `readiness.go:446-460`). Report §4 "audit atomicity" is therefore unproven for SHG (only
  presence test: `hr_test.go:375`).
- Minimum fix: mirror `comp_lifecycle_test.go` style — row-count + audit-absence assertions
  after each of the 7 rejections, cross-scope function negatives, SHG drift and backdated-add
  tests; or narrow report wording to "audit presence".

### SHG clean core

Missing≠verified-none, pending/opt-out block, superseded stays historical, corrupt blocked
(matrix + `SHG_CORRUPT` tests); `verified_none` triple enforced at 3 layers (CHECK, functions,
Go) and tested; cross-employee/employer comp refused; approval input-lock + dual-leg conflict
detection sound (population-lock gap is §6 P1a, not an SHG-semantics defect); validation
refuses on blockers; correction immutability / successor history / revalidation versioning hold.

## 5. Rule-pack activation findings

### C1-SOURCES — P1 — blocks: YES

- Location: `internal/rules/rules.go:124-132` (`Manifest.Verify`).
- Scenario: an IMPLEMENTABLE pack ships manifest source entries whose sha256 digests match no
  real capture file (fabricated/stale pins); `Verify` accepts (regex-checks digest format +
  `Stat`-checks `SOURCES.md` presence only, never reads/re-hashes anything under `sources/`).
  The manifest has no source-id→file binding, so byte verification is impossible by
  construction; bundle digest (`rules.go:281-300`) covers manifest+REVIEW+spec+gate only.
  Violates AGENTS.md boundary 5 (pinned official sources) and C1 (recompute digests of ACTUAL files).
- Evidence: full `LoadManifest`/`Verify`/`readPackFiles`/`bundleDigest` bodies; grep confirms
  zero reads of `sources/`; 18-case rejection table has no fabricated-digest case; unit tests pass.
- Minimum fix: add source file refs to the manifest schema; recompute each declared digest
  from bytes in `Verify`; reject on mismatch/unresolvable ref before any statutory activation.

### C1-GOLDENS — P1 — blocks: YES

- Location: `internal/rules/rules.go:169-193`.
- Scenario: any pack with non-empty `fixtures_gate` files activates with
  `VerificationStatus=GOLDENS_VERIFIED` although no fixture is executed, parsed, or
  shape-checked (readability only, then hardcoded `GOLDENS_VERIFIED`). A failing/garbage golden
  is recorded in history + audit as goldens-verified. DB compounds it (§5 C2 labels).
- Evidence: `Verify:169-194` read; tests cover missing gate file only, never failing/invalid
  content; happy-path asserts the stored label, never that fixtures ran; no fixture runner on
  the CLI path (`admin.go:86-93` Verify-then-Activate only).
- Minimum fix: execute the golden fixtures (or verify a signed/out-of-band gate artifact)
  inside `Verify`; derive `VerificationStatus` from the outcome; reject failing/unparseable goldens.

### C2-REPIN — P1 — blocks: YES

- Location: `00012` `rule_pack` (450–460) + `activate_rule_pack` (883–986); no trigger/audit
  on `rule_pack`.
- Scenario: `migration_admin` runs `UPDATE rule_pack SET reviewed_version/digest` (as the
  repo's own tests do), changing which version+digest the function accepts, with zero audit
  row, zero history, irrecoverable prior pin. The registry pin is the entire trust root
  (function compares only version+digest), yet pin-setting is raw unaudited SQL with no
  on-disk bundle binding.
- Evidence: LIVE on `c3rev_areac_01`: repin+supersede accepted; trigger grep shows none on
  `rule_pack`/`rule_pack_activation`/history.
- Minimum fix: set pins only through an audited function (or audit trigger on `rule_pack`)
  recording actor/reason/prior pin and requiring digest = freshly recomputed bundle digest;
  document the pin ceremony before the first statutory pin.

### C2-EVIDENCE-LABELS — P2 — blocks: NO (today)

- Location: `00012` `activate_rule_pack:906-941`.
- Scenario: admin activates with `spec_version`/`review_ref`/`verification_status` arbitrarily
  different from the pin; function checks only identity/jurisdiction/scope/version/digest and
  stores forged labels in history+audit. Registry `spec_version`/`verification_status` are
  write-only (never SELECTed). Audit-label forgery, not bundle substitution (bytes protected
  by the manifest_digest pin).
- Evidence: LIVE: forged `spec-FORGED`/`SELF_ATTESTED` activation ACCEPTED, stored verbatim.
- Minimum fix: compare `p_spec_version`/`p_verification_status` to the registry pin (exact
  match); reject on mismatch.

### C2-ADMIN-DIRECT-WRITE — P2 — blocks: NO

- Location: `rule_pack_activation` (`00003:81-88`) + history (`00012:481-500`); no triggers.
- Scenario: `migration_admin` bypasses `activate_rule_pack` with direct INSERT/UPDATE (repo
  tests do this), creating an IMPLEMENTABLE current pointer with no history/evidence/audit —
  fully effective because approval/snapshot read the current pointer. Function governance holds
  for `app` only; admin path unguarded.
- Evidence: live seed via direct admin INSERT; no trigger on either table; `resolvePacks`
  (`snapshot.go:1091-1108`) tolerates missing history.
- Minimum fix: trigger on `rule_pack_activation` rejecting direct writes (allow only
  function/owner-session writes) or requiring a matching history row; route BLOCKED seeding
  through `ensure_pack_blocked`.

### C2-EFFECTIVE-COVERAGE — P2 — blocks: NO today; PRE-PIN prerequisite (disposition)

- Location: `00012:942-944` + `snapshot.go:1085-1115` + `readiness.go`/`approval.go`.
- Scenario: activation `effective_from/to` are caller-supplied, range-checked only against each
  other, never compared to the payroll period: readiness gates on status (`approval.go:124`);
  snapshot pins the dates without coverage comparison. A Jan–Jun pack can silently govern a
  December run. DISPOSITION: no-block today only because nothing is IMPLEMENTABLE (all runs
  block at the approval gate); the semantics + enforcement must exist before the first
  statutory pin, otherwise the first governed calculation is silently mis-governed.
- Evidence: full `resolvePacks` + `AuthorizeApproval` bodies; pack dates carried but never
  period-compared; arbitrary-date activations accepted live.
- Minimum fix: define period-coverage semantics; enforce at readiness (block) or snapshot
  (record decision + hash it) before the first statutory pin.

### C2-HISTORY-STATUS — P2 — blocks: NO

- Location: `00012:481-500` + INSERT `:962-970`.
- Scenario: every history row stays `IMPLEMENTABLE` forever; superseded rows never marked
  `SUPERSEDED` (allowed by CHECK, written nowhere). Audit readers see N IMPLEMENTABLE rows per
  pack; tip must be reconstructed via `ORDER BY activated_at ... LIMIT 1` (tie-fragile).
- Evidence: LIVE: after supersede, history shows 1.0.0/IMPLEMENTABLE + 1.1.0/IMPLEMENTABLE;
  current pointer correctly single.
- Minimum fix: UPDATE the superseded history row to `SUPERSEDED` in the same function (or
  derive status at read time and drop the misleading column).

### C1-SYMLINK / C1-TOCTOU / C4-TEST-WEAKNESS — P3 — blocks: NO

- Symlink (`rules.go:258-277`): manifest file refs followed via `os.ReadFile` with no
  Lstat/EvalSymlinks/regular-file check; a directory named `SOURCES.md` passes the Stat-only
  check. Bounded (contents hashed, never stored/echoed; pack-dir writers can already inline
  bytes). Fix: Lstat, reject non-regular/symlink (or EvalSymlinks + re-contain).
- TOCTOU (`rules.go:89+176`, `admin.go:86-93,118`): manifest read twice + Verify→DB window;
  DB compares digest-to-pin, never digest-to-disk. Narrow same-operator window; snapshots pin
  the recorded DB digest, so calculations unaffected (audit-vs-disk skew only). Fix: single-read
  manifest bytes through; re-hash immediately before submit, abort on difference.
- Test weakness (`admin_test.go:246-271`, `readiness.go:597-631`): `TestOT001RemainsBlocked`
  asserts manifest BLOCKED (real) + zero-row count on a FRESH DB (vacuous); never asserts
  registry membership/unpinned state/activation rejection. OT-001 also outside the
  required-pack closure (OT codes → parent SG-MOM-OT), so activating it would be inert — the L1
  slice is unwired. Neither is a live defect today. Fix: assert registry contains OT-001 with
  NULL pin + activation fails; document OT vs OT-001 wiring in the slice plan.

### Pack clean core

- C3 (app privileges): 29/29 direct writes denied as REAL app with zero partial change
  (registry, activation, history I/U/D, all six param tables I/U/D, EXECUTE
  `activate_rule_pack`); `ensure_pack_blocked` allowed+idempotent, never downgrades live
  IMPLEMENTABLE, rejects unknown packs. (Extend repo tests to UPDATE/DELETE on history+params
  to lock this in.)
- C2 core: unknown/unpinned/invented-version/digest/jurisdiction/scope mismatch/empty evidence
  rejected; repeat-version rejected; downgrade-after-repin rejected; concurrent racers serialize
  (1 winner/1 history/1 atomic audit); supersession links + single current pointer hold;
  capture takes the same pack key in sorted order.
- C4/C5: exactly the ten registry identities, zero pinned, zero non-BLOCKED; nothing
  IMPLEMENTABLE; approval requires IMPLEMENTABLE on every required pack so all runs block;
  CLI/docs never claim statutory readiness; fixture packs isolated.

## 6. Snapshot isolation and population-barrier findings

### D1-P0 — P0 — blocks: YES — PROTOCOL DEFECTIVE (live-proven)

- Location: `core/database/lock.go:67-88` (`BeginSnapshotTx`); `internal/payroll/snapshot.go:282-345`
  (`captureInTx`); `internal/payroll/validate.go:94-212` (`validateOnce`).
- Scenario: writer A holds a barrier lock + uncommitted insert; reader B starts an RR txn
  (snapshot fixed at the FIRST statement — including the lock-acquire SELECT itself, before the
  lock wait completes), blocks on the same lock, A commits, B proceeds blind. Capture commits a
  stale-consistent snapshot; validation's `tipDrifted` (same stale snapshot, zero advisory locks
  in the tip path) reports no drift and the run VALIDATES missing the new employee/input with no
  record of the miss. Worst case is not a mixed-version snapshot (single RR txn yields consistent
  prefixes — no mixed-version vector found) but a run VALIDATED while missing employees/inputs.
- Evidence (LIVE, disposable PG16.15 `127.0.0.1:55433`, app role, 12 migrations): pure modes
  lockfirst/readfirst/settxn ALL MISS-WRITER-COMMIT, committed-before SEES, sessionlock-prefix
  SEES; real-service writer-first races T1-pop (employees=1, RACER missing), T3-comp (RACE
  allowance missing), T4-pack (pinned v0, v1 committed), T5-approval (missing-timesheet,
  approval committed) ALL MISS with post-hoc `drift=true`; T2 VALIDATION returned `valid=true`,
  `stale=false`, `status=validated` while fresh `DetectDrift` shows RACER added. B provably
  blocked via `pg_locks` NOT-granted before A committed in every run.
- Minimum fix: acquire the barrier as SESSION-level advisory locks BEFORE `BEGIN RR`
  (live-proven SEES) in both `CaptureSnapshot` and `ValidateRun` (drift-check path included);
  close P1a–g so every writer joins the pre-txn barrier. Do not rely on in-txn xact locks for
  snapshot ordering.

**D1 verdict: FAIL.** The RR + in-transaction-advisory-lock ordering does not establish a
consistent view against a writer that commits during the lock wait. The design objective is
correct; the acquisition order is defective and must be fixed (session-level pre-BEGIN barrier)
before the first calculation.

### Barrier gaps P1a–g (each fix-gating for D1; 7 items)

- **P1a (approval population-lock) — P1 — blocks: YES (fix-gating).**
  `internal/timesheet/timesheet.go:734-786` (LockPayrollInput only); `00011:139-166` (input lock
  only); `snapshot.go:693-703` (terms-holders UNION approved-sheet holders). Approval can ADD an
  employee to the expected population yet takes no population lock in Go or trigger paths.
  Reconciled against CLEAN-A/B-SHARED (input-lock + seq sound): input side sound,
  population side missing. Fix: approval (service + barrier trigger) takes the population
  lock/session-lock alongside the employee lock, preserving population→employee order.
- **P1b (terms-close input-lock) — P1 — blocks: YES (fix-gating).**
  `00012:835-877` (`close_employment_terms`: population lock only). Close changes snapshotted
  per-employee terms (`canonTerms To`) but the DB-function path never takes the employee input
  lock (Go `SetTerms` takes both — Go path covered). Fix: function takes
  `acquire_payroll_input_lock` after the population lock.
- **P1c (disposition input-lock) — P1 — blocks: YES (fix-gating).**
  `00011:937-981` (pair lock only; comment admits it). Direct-SQL disposition INSERT changes
  snapshotted unpaired-clock evidence without the employee input lock (Go path takes it).
  Fix: trigger takes the employee input lock (resolvable from the loaded event row).
- **P1d (`status_history` direct INSERT) — P1 — blocks: YES (fix-gating).**
  `app` INSERT on `employee_status_history` confirmed live; zero triggers on that table.
  Fix: revoke direct INSERT (function-only via `transition_employee_status`) or add barrier trigger.
- **P1e (`ensure_pack_blocked`) — P1 — blocks: YES (fix-gating).**
  `00012:621-632` (no lock call; EXECUTE granted to app). Concurrent capture can
  miss/include inconsistently under the fixed protocol (drift catches it post-hoc today).
  Fix: function takes `acquire_rule_pack_lock` (session-level under the fix).
- **P1f (draft lines) — P1 — blocks: YES (fix-gating).**
  No input-lock trigger on `timesheet_line`/`timesheet_line_interval`; `00004:70-80` freeze
  covers decided sheets only. Direct-SQL DRAFT-line I/U/D (live INSERT=true) change snapshotted
  lines lock-free. Fix: line-write trigger takes the owning employee's input lock (via timesheet
  join), or confine line writes to locked service paths.
- **P1g (period/employer UPDATE) — P1 — blocks: TRANSITIVELY YES (disposition).**
  `app` UPDATE `pay_period` + `employer` both true live; no revokes. Capture inputs (period
  bounds, jurisdiction) directly updatable with no lock/trigger — missed exactly like the P0
  scenario. Reconciled with EF-E1N-P3 (jurisdiction drift-blind, P3): single disposition —
  REVOKE app UPDATE on `pay_period`/`employer` (route through locked functions) and/or compare
  header fields in `tipDrifted`. Marked "admin-hygiene, first payroll needs no period mutation"
  by Area D, but since the D1 fix is pre-calculation, P1g must close with the fix — hence
  transitively pre-calculation.

### D5-P2a (40001 misclassification) — P2 — blocks: NO

- `lock.go:75-81` claims no-40001; `snapshot.go:287-297` run `FOR UPDATE` has no 40001 mapping;
  `apperr.go:79-94` defaults to `CodeInternal`. Concurrent Validate (updates run row) + Capture
  (`FOR UPDATE` same row): loser's `FOR UPDATE` raises 40001 under RR, surfaced as
  `CodeInternal` "lock payroll run failed" instead of retryable conflict. LIVE T8 confirms
  (raw-SQL twin: SQLSTATE 40001). Fix: map 40001 at capture run-lock to `CodeConflict`.

### D5-P2b (vacuous revalidation) — P2 — blocks: NO

- `snapshot.go:326-345`: same-transaction re-derivation under RR is provably identical, so the
  "changed during lock acquisition" conflict can never fire; "any lock-convention violator fails
  loudly" is false (violators committing after snapshot-fix are silently invisible). Fix: remove
  or replace with fresh-snapshot post-lock verification.

### D6-P2c (writer-first test gap) — P2 — blocks: NO (gates the fix)

- `snapshot_race_test.go:33-113` + `dbutil.go:279-349`: all 8 race tests are capture-first
  (stall capture → writer fails → release → retry → drift). The D1 order (writer-first, capture
  waits, writer commits) is never exercised — the suite passes despite the P0. Fix: add
  writer-first interleavings asserting inclusion (expected to fail until the protocol is fixed).

### Isolation clean core

Single-txn capture verified (all queries on `tx`, no connection escape, uniform RR, full
rollback); multi-employee locks UUID-sorted; all keys via the three helpers (Go/capture/
activation uniform — first-body inline formula superseded by OR REPLACE); pack pins cannot mix
manifest/activation within one RR snapshot; blocked pins recorded and refused by validation.
Population barrier covered for: employee create (Go + trigger), terms create (Go + triggers),
status/termination (Go + function; direct UPDATE/DELETE revoked), comp add/correct
(function-only + locks; direct DML revoked), SHG (function-only + locks; direct revoked),
clock/interval INSERT triggers, approval input lock, exclusions run-row serialization.
Exceptions are exactly the P0/P1/P2 items above.

## 7. Snapshot completeness, hashing and immutability findings

### EF-E3-P0 — P0 — blocks: YES

- Location: `00012:1094-1099` (REVOKE/GRANT; no freeze trigger on `payroll_snapshot*`).
- Scenario: as `app`: (a) INSERT fabricated valid header+children with arbitrary
  `content_hash` on a fresh run; (b) INSERT extra employee/pack rows into a standing VALID
  snapshot — hash unchanged, content silently altered. Combined with the F1 trigger gap this
  yields a validated run over attacker-chosen inputs.
- Evidence: LIVE (`p16ef_live4/6`, `current_user=app`): fabricated header+child+4 packs
  exit=0; child INSERT count 1→2 exit=0; pack INSERT exit=0; `content_hash` byte-identical
  after injections; direct invalidated-INSERT exit=0 (fake history/version-squat). No trigger
  on snapshot tables (grep).
- Minimum fix: REVOKE INSERT ON `payroll_snapshot*` FROM `app`; perform capture INSERTs inside
  a `SECURITY DEFINER` capture function (role alone cannot distinguish capture from attacker).

### Snapshot clean core (E1/E2/E3m/E4)

- **E1 completeness + E2 canonical hash — CLEAN.** No omitted mutable input found: every
  unpinned attribute lives on an INSERT-only row (terms attrs, comp fields, SHG evidence,
  interval/clock content, employee identity) with change-by-new-ID; the two in-place writers
  (close→`effective_to`, transition→status+history) touch only pinned fields. Canon excludes
  run/version/UUIDs/`captured_at`; all collections sorted; money `NUMERIC::text`; dates
  `::text`; `approved_at` UTC RFC3339Nano. LIVE: cross-run same content H1==H2; comp+sheet /
  pack-flip / exclusion each change the hash (+ named `SG-SDL` blocker on pack flip).
  Committed determinism/difference tests concur (green).
- **E3 UPDATE/DELETE denials + state preservation — CLEAN.** LIVE as app: all 7 UPDATE/DELETE
  attempts (header hash, child decision, pack version, child DELETE, re-parent, direct status
  flip, header DELETE) fail 42501; hash byte-identical before/after; second-valid INSERT fails
  23505; invalidate→recapture lifecycle works. (INSERT gap is EF-E3-P0 above.)
- **E4 invalidation/revalidation 8/8 — CLEAN.** LIVE via CLI+SQL: comp+sheet→stale v1→v2 with
  byte-identical original + single tip; pack→stale+blocker; status/SHG/population/attendance/
  terms each→stale + new version + new hash + single tip; stale offer after v2 returns
  version-mismatch (no adoption); audit shows capture + invalidate rows. Service offer protocol
  (stale-never-adopts, idempotent replay), full-doc rehash, pin stamping + audit all hold.

## 8. Payroll-run lifecycle and direct-SQL findings

### EF-F1A-P1 — P1 — blocks: YES

- Location: `00012:1174-1268` (`check_run_lifecycle()`) + missing UPDATE revoke on `payroll_run`.
- Scenario: as `app`: `UPDATE payroll_run SET status='validated'` over a fabricated/stale tip
  succeeds with `inputs_hash='placeholder'`, `rule_versions='{}'`, zero `run.validate` audit
  rows, no offer check, no drift check. Post-validation, pins are freely rewritable
  (status-unchanged updates RETURN NEW). A calculator reading `inputs_hash`/pins cannot trust
  validated runs.
- Evidence: LIVE as app: direct-validate exit=0 (`status=validated`, placeholder pins, 0 audit
  rows); pin-tamper UPDATE exit=0; `has_table_privilege` app UPDATE/DELETE on `payroll_run` = true.
- Minimum fix: revoke app UPDATE(`status`,`inputs_hash`,`rule_versions`) on `payroll_run`;
  expose validation only via a `SECURITY DEFINER` function that stamps pins + audit; harden the
  trigger to require `inputs_hash` = tip `content_hash`, pins match tip, ≥1 employee row.

### EF-F1B-P1 — P1 — blocks: YES

- Location: `00012:1225-1242` (no population count in draft→validated branch).
- Scenario: as `app`: fabricate a valid tip with ZERO employee rows + IMPLEMENTABLE packs, then
  UPDATE run→validated: trigger counts (blocked=0, packs ok) pass vacuously — a validated run
  the service path (`validate.go:133-138` empty-population refusal) would reject.
- Evidence: LIVE as app on fresh employer: vacuous-validate exit=0. Committed
  `TestValidateEmptyPopulation` (green) proves the service refuses.
- Minimum fix: same trigger hardening (require ≥1 snapshot employee row).

**F1 disposition: GOVERNANCE DEFECT, P1 blocking — not acceptable.** The acknowledged
"outcome-equivalent direct draft→validated" bypass is outcome-equivalent only in the status
string: it drops validation actor, audit, offer identity, pin-stamping, drift/offer checks, and
permits placeholder/vacuous/tampered pins. Blocked-tip/snapshotless/second-regular direct
rejections hold (verified), but the clean-tip direct-flip gap is real and must be closed with the
revoke-behind-DEFINER + trigger-hardening correction before the first calculation. Do not treat
the trigger's terminality check as making service audit dispensable.

### EF-F4-P3 (empty actor) / EF-F3-P3 (self-link) — P3 — blocks: NO

- Empty actor (`00002:191-206`): `actor=''` INSERT succeeds on a draft run (reason has length
  check, actor only NOT NULL). No financial impact; service mandates actor; audit row still
  records actor on the service path. Fix: `CHECK (char_length(actor)>0)`.
- Self-link (`00012:1202-1221`): pre-snapshot `corrects_run_id=<self>` UPDATE succeeds
  (employer/period-only check); 2-cycles constructible. Harmless: validation requires the
  original validated first, so cyclic runs can never validate (proven live). Fix (optional):
  `CHECK (corrects_run_id IS DISTINCT FROM id)`.

### Lifecycle clean core (F2/F3/F4/F5)

State machine + correction linearity hold (born-draft, exactly-one-tip, included/excluded-only,
IMPLEMENTABLE-only, original-first, successor-singleton; validated→calculated/approved/finalized/
draft rejected; pre-validated/second-regular/fork/identity-change rejected; `calculated`/
`approved`/`finalized` structurally unreachable without genuine artifacts; no fabricated
calculation artifact). Exclusions run-scoped, draft-only, visible in readiness; bare-draft DELETE
succeeds leaving an orphaned-but-present `run_create` audit row (no service delete API; accepted,
no replay-guarantee compromise); history-backed runs FK-protected.

## 9. Payroll-line future-calculation boundary

**Verdict: harmless today AND a mandatory pre-calculation prerequisite (P2, G-PREREQ — blocks
the first calculation the day a calculator writes).**

- Zero influence today (G-CLEAN): no Go production code references `payroll_line` (hits only in
  freeze/lineage/privilege tests + migrations + docs); `canonSnapshot`/`canonEmployee` contain
  no line fields; capture reads run/period/employer/employee/terms/status/comp/timesheet/
  intervals/clock/SHG/packs/exclusions only; validation assesses snapshot tables only;
  readiness never queries lines; zero line endpoints in `openapi.yaml`. LIVE as app on
  `m2rev6probe2` (validated run H1 via real CLI): post-validation INSERT ok, UPDATE amount to
  99999.00 persisted, `content_hash` identical after both (H1==H2==H3), CLI readiness verdict
  byte-identical. CHECK `payroll_line_source_refs_check` still enforced; `run_lifecycle` alive.
- Missing protections the day a calculator writes (G-PREREQ): (a) no FK/reference from
  `payroll_line` to `payroll_snapshot` (catalog count 0); (b) only trigger fires on
  `status='approved'` which `run_lifecycle` makes unreachable (dead freeze);
  (c) `app` holds full SELECT/INSERT/UPDATE/DELETE (live true×4); (d) source_refs CHECK verifies
  key presence only (zero-UUID `terms_id` accepted). LIVE as app post-validation: INSERT,
  UPDATE, cross-run re-parent, DELETE all ok. Corroborated by `freeze_test.go:242` (UPDATE after
  validated succeeds) and `:88-101` (re-parent allowed).
- Minimum before the first calculation (all 6): (1) exact line→frozen-snapshot relation (FK to
  snapshot tip + version); (2) line-level provenance (verified IDs, not free-form JSON keys);
  (3) keep `NUMERIC(12,2)` NOT NULL + add the rounding-boundary rule; (4) post-calc immutability
  trigger keyed on validated/calculated status; (5) recalc/supersede as new auditable versions,
  never in-place; (6) revoke app direct writes (controlled calculator function only).

## 10. Migration and test-harness verification

### H-BATTERY-CLEAN — CLEAN

19-package claim independently reconciled by explicit method: `go list ./...` = 26 packages;
7 without test files (observability, apperr, audit, duration, `migrations/core`,
`sg-wfm-payroll`, dbutil) ⇒ 19 test-bearing. Results on reviewer-owned disposable PG16.15
(fresh initdb, fresh base names per run): `gofmt`/`go vet`/no-`lib/pq`/`git diff --check` clean;
`task build` ok; `go test -count=1 ./...` 19ok/0fail default TZ, 19ok `TZ=UTC`, 19ok
`TZ=Asia/Singapore`; race (attendance/timesheet/hr/payroll/wfm) all ok, 0 DATA RACE;
provisioning twice PASS; `TestDown*` PASS; M2-pattern 125 pass/0 fail/1 intentional skip
(`TestReproducibilityDeferred`, deferred-by-design); contract
(`TestServedSubsetOfSpec`+`TestSpecImplementedAllServed`) PASS; parity
(`TestCLIAPIParity`+`TestCLIAPIParityM2`) PASS; compbasis carve + seed suites PASS with 0 skips
(env set, no DB-gated skip taken). Fixed-name reusable DBs confirmed (`Setup` uses DSN database
name) so the recycle procedure is load-bearing; lint repair is a one-word diff (`examples/`
never existed in HEAD); `payroll_line` has zero REVOKE/GRANT (still app-writable — matches §9);
no Docker claim (no daemon). Preserve `GOWORK=off` + fresh-DB discipline for acceptance.

### H-STALE-SILENT — P2 — blocks: NO (procedure mandatory for acceptance)

- Location: `core/database/migrate.go:67-85` (no checksum); `tests/dbutil/dbutil.go:166-271`
  (create-if-missing + pending-only `MigrateUp`, no drop/guard); goose v3.27.3 `Status`
  (version-row presence only, zero checksum); fixed-name suite DBs (except employer drop-first
  and nano-unique down/fresh/role_race DBs).
- Scenario (P6 recurrence): a developer appends a section to applied `00012` (or edits any
  applied file), then runs `go test` on a warm server reusing `a1test*` names: `MigrateUp`
  applies 0, tests run against stale schema, governance gaps pass green — exactly as integrity
  once passed a phase without `run_lifecycle`. The standing drop-all-`a1test*` procedure exists
  ONLY in prose; repo-wide grep finds no script/Taskfile target/test performing it.
  `TestMigrateUpAndStatus` counts FS sources, not applied content — cannot catch this.
- Evidence: LIVE on `m2rev6stale2`: migrate up (trigger=1, goose max=12) → DROP TRIGGER →
  migrate up exit 0 "applied 0" → trigger still 0, version still 12, status "12 applied".
- Minimum: before calculation acceptance, either (a) automate drop-first for every suite-owned
  DB (extend `employer_test.go:58` pattern) or use unique DBs per run, or (b) add a
  migration-content guard (hash of embedded FS vs stored hash table, checked in
  `SetupDB`/`MigrateUp`; goose cannot do this natively). Until then, acceptance must run on a
  provably clean server with fresh names (as this review did).

### H-GOOSE-VERWRITE — P3 — blocks: NO

`privilege_test.go:363` pins app full DML on `goose_db_version` as-is. In-scope app-role SQL can
forge/delete a version row; a later admin `MigrateUp` then silently skips/re-applies that
version. No production path needs app writes here (single repo-wide hit). Fix: revoke app
INSERT/UPDATE/DELETE (keep SELECT at most), update the fail-closed pin, re-run the suite.

### H-EPIHEMERAL-PORT — P3 — blocks: NO

`provision_test.go:49-71` fixed port 5435 (trust, `t.TempDir`): two concurrent full-suite runs
on one host collide; the loser fails spuriously. Single-run safe. Fix: port 0 / free-port pick
(or env override).

## 11. Confirmed defects

Blocking (must close before the first calculation):

| ID | Sev | Defect | Evidence |
|---|---|---|---|
| D1-P0 | P0 | Stale-consistent capture/validation (writer commits during lock wait invisible; run VALIDATES missing inputs) | Live writer-first T1/T3/T4/T5 MISS + T2 `valid=true`/`validated` missing RACER |
| EF-E3-P0 | P0 | Fabricated/post-validation snapshot INSERT as `app`, hash unchanged | Live INSERT×3 exit=0, hash byte-identical |
| F-A1-WIDEN | P1 | Widening correction permanently bricks resolution | Deterministic query+branch; live probe still required |
| F-A1-FREETAIL | P1 | Narrowed-then-fresh-add permanently unresolvable | Deterministic + repo tests pin both halves; live probe required |
| F-A1-MIDPERIOD | P1 | 2-segment month silently included, no blocker | Code + report-claim gap (no guard/test exists) |
| F-B1-ACTIVE | P1 | Active-without-verification counts as complete (vs B1) | Code + matrix; needs enforce-or-amend decision |
| F-B2-PARTIAL | P1 | Partial-period SHG covers full period | Exact query+branches; 9 live cases missing |
| X-FUND | P1 | Cross-fund multiples silently pass as covered (report claims rejection) | Code + coexistence test; needs decision + live matrix |
| C1-SOURCES | P1 | Manifest source digests never re-hashed from bytes | Full bodies + grep; unit pass proves untested |
| C1-GOLDENS | P1 | `GOLDENS_VERIFIED` hardcoded without executing fixtures | Body + CLI path; no runner exists |
| C2-REPIN | P1 | Registry pin silently mutable via raw SQL, zero audit | Live repin+supersede accepted |
| P1a–g | P1 | 7 barrier gaps gating any D1 fix (approval-pop, terms-close, disposition, status_history, ensure_blocked, draft lines, period/employer) | Inspection + live privilege matrix |
| EF-F1A | P1 | Direct draft→validated with placeholder pins, zero audit, rewritable | Live exit=0 + tamper exit=0 |
| EF-F1B | P1 | Vacuous empty-tip direct validation | Live exit=0 (service refuses) |
| G-PREREQ | P2* | `payroll_line` ungoverned the day a calculator writes | Live I/U/D + re-parent ok (*blocks at calculation) |

Non-blocking confirmed (bounded gaps, test/hardening):

C2-EVIDENCE-LABELS, C2-ADMIN-DIRECT-WRITE, C2-EFFECTIVE-COVERAGE (pre-pin prerequisite),
C2-HISTORY-STATUS (P2); D5-P2a (40001), D5-P2b (vacuous check), D6-P2c (writer-first tests),
F-A2-TERMSLINK, H-STALE-SILENT (acceptance procedure), S1-P4-AUDIT-WORDING (P2);
P3s: EF-F4, EF-E1N (folded into P1g disposition), EF-F3, C1-SYMLINK, C1-TOCTOU,
C4-TEST-WEAKNESS, F-B4-SHG-EVIDENCE, H-GOOSE-VERWRITE, H-EPIHEMERAL-PORT.

Report-claim corrections required: mid-period blocker (§3/§12), cross-fund rejection (§4),
SHG "audit atomicity" → "audit presence" unless tests added (§4), §10D "26 cases" tally (22
test functions counted — minor, no missing coverage found).

## 12. Risks and untested hypotheses

1. Area A/B have zero live-DB execution (client-only PG, Docker denied): WIDEN/FREETAIL/
   MIDPERIOD/PARTIAL/ACTIVE reachability + exact conflict/blocker strings need the four live
   probes on disposable PG16 (§16). Evidence for those findings is deterministic
   exact-query+exact-branch inspection + existing test bodies.
2. D1: 4 writer classes (term-change, termination, attendance-correction, SHG-correction) lack
   individual live writer-first rows; MISS inferred from shared RR-fix timing + advisory-xact
   primitive + 4 live class proofs. Run one live writer-first race per class.
3. B2: 9 coverage cases (mid-start/mid-end/gap/successor-date) have no live test. B1: owner
   intent undecided. Cross-fund: live multi-fund matrix missing (decision recorded in §4, tests
   still required). SHG B4: per-rejection audit-absence, cross-scope function negatives, SHG
   drift + backdated-add live probes missing.
4. Pack: C1-SOURCES/GOLDENS Go-inspection + unit only (no live fabricated-digest/failing-golden
   rejection probe); C2-EFFECTIVE-COVERAGE period semantics undefined (pre-pin prerequisite).
5. E/F concurrency: interval-correction-vs-capture race static-only; dual-capture/
   dual-invalidate beyond committed `TestValidateConcurrent` not live. Run both live races
   (correction-during-capture; dual-capture) asserting single-tip + drift + no adoption.
6. CLI/API: only committed parity/contract PASS cited; no independent live CLI-vs-API
   differential on a disposable DB (same services + error envelope + request IDs under
   direct-app-role + concurrency). Required to close V4 (§16).
7. Harness: 19-name import-path list + per-run skip table not published (only 26−7 count +
   M2-pattern 125/0/1); prior-review IDs (SHG-B1a/B1b/B1c, B2a/B2b/B2c, C1–C4, C6, S5-F1) not
   cross-checked; §10D tally open. Jurisdiction/TZ hash behavior under employer change not
   probed live (fails closed at capture; tracked via P1g/E1N).
8. `SnapSHG` pins only a `Tripled` boolean, not verified_at/by/basis values (Area E observation):
   app-role immutability + new-row-on-change keeps drift detection sound; value-level evidence
   pinning belongs to a future E1-completeness decision, not this gate.
9. PostgreSQL C-level snapshot-fixation call site not source-inspected; timing established
   behaviorally (lockfirst MISS / sessionlock-prefix SEES) — dispositive for the protocol
   question. Cross-employer key separation inspected only, not live-tested.

## 13. Clean areas

- Baseline/migrations: HEAD, tree inventory, `00001`–`00011` byte-identical, `00012` complete
  with all claimed objects; `bchat/` untouched.
- Comp: single authoritative resolver shared by readiness + snapshots; no fork/summing/
  double-payable/ignored-segment/in-term violation; chain/subtraction/containment/guard logic
  + 16-test matrix sound (except the P1 findings above).
- SHG: 7-state mechanics (missing/pending/active/verified_none/opt-out/superseded/corrupt),
  verified_none triple at 3 layers, missing≠verified-none, pending/opt-out/corrupt blocking,
  superseded historical, cross-scope comp refusal, approval input-lock + dual-leg conflicts,
  validation-refuses-on-blockers, correction immutability/history/versioning.
- Pack: 29/29 app denials with zero partial change; unknown/unpinned/mismatch/empty-evidence/
  repeat/downgrade rejections; racer serialization 1-winner/1-history/1-audit; exactly ten
  registry identities, zero pinned, zero non-BLOCKED; all runs block at the approval gate;
  no false statutory-readiness claim; fixture isolation.
- Snapshot/run: single-txn RR capture with sorted multi-employee locks via uniform helpers;
  canonical hash excludes run/version/UUIDs (cross-run identity proven); 8/8 drift categories
  force stale revalidation with byte-identical originals + single tips + no-adoption;
  UPDATE/DELETE denials with hash-identical state; state machine + correction linearity +
  exclusion scoping + deletion semantics (22 payroll + 3 integrity green).
- Harness: full battery green across 3 TZs + race + provisioning + down/fresh/role_race +
  parity + contract as detailed in §10 (clean-server, fresh names).
- `payroll_line`: zero influence today (§9); CHECK + `run_lifecycle` alive.

## 14. Four verdicts

### V1 — Compensation and employee-input integrity: FAIL

WIDEN/FREETAIL (permanent bricks), MIDPERIOD (silent 2-segment inclusion vs required explicit
blocker), and the terms-provenance gap break "one deterministic, explainable compensation
basis; ambiguity blocks." Single-resolver architecture is sound, but the first monthly profile
cannot trust resolved inputs until the three P1s are fixed and live-probed.

### V2 — SHG and rule-pack governance: FAIL

SHG: B1-active decision missing, B2 partial-period coverage, cross-fund ambiguity, and
under-evidenced rejections break "fully evidenced coverage or precise blocker."
Packs: C1-SOURCES/GOLDENS (unverified bytes labeled verified) + C2-REPIN (silently mutable
trust root) break the evidence-backed activation gate; the DB core (denials, serialization,
all-BLOCKED status) is sound but the trust root and evidence verification are not. Nothing
statutory is falsely claimed ready — the mechanism must be fixed before the first pin.

### V3 — Snapshot and payroll-run integrity: FAIL

D1-P0 (stale-consistent capture + validation missing committed inputs, live-proven at both
capture and validation levels), EF-E3-P0 (fabricated snapshot INSERT as app), and EF-F1A/B
(direct draft→validated with placeholder/vacuous pins, zero audit, rewritable) mean there is
today no immutable, complete, governed snapshot and no trustworthy validated state. Hash
canon/drift/state-machine/correction-linearity are sound on the service path — the isolation
ordering and the direct-SQL governance are not.

### V4 — CLI/API operability: PASS WITH CONDITIONS

Committed `TestCLIAPIParityM2`, `TestCLIAPIParity`, `TestServedSubsetOfSpec`, and
`TestSpecImplementedAllServed` pass in the §10 battery (same services, both-directions
contract, implemented-only serving; 12 new `00012` functions in the fail-closed privilege
catalog). No operability defect found. CONDITION: run one independent live CLI-vs-API
differential on a disposable DB (same scenario via both surfaces incl. error envelope +
request IDs, under app role + one concurrent pair) before calculation acceptance, and publish
the 19-name package list + per-run skip table proving 0 DB-gated skips.

## 15. Overall Milestone 2 acceptance decision

**FAIL — NOT READY (per §14: a remaining defect can produce an incomplete, inconsistent,
untraceable, or ungoverned payroll-input set).**

This is not a "green packages" judgment: the 19-package battery is genuinely green (§10), and
large clean cores exist (§13). It fails because live transaction behavior and direct app-role
operations do not support the same invariants as the Go services:

- A run can VALIDATE while missing employees/inputs (D1-P0, proven at validation level).
- A snapshot can be fabricated or extended as `app` with an unchanged hash (EF-E3-P0).
- A run can be directly validated with placeholder/vacuous/rewritable pins and zero audit (F1).
- Compensation/SHG inputs can brick, silently include, or silently pass ambiguous coverage.
- A pack can be pinned to unverified bytes via an unaudited trust root.

Do not close the infrastructure/governance phase. Complete §16, re-verify each item live, then
re-run §10 acceptance on a provably clean server before any calculation work.

## 16. Minimum prerequisites before the first calculation

Code (blocking):

1. D1 + P1a–g: session-level pre-BEGIN barrier locks in `CaptureSnapshot` + `ValidateRun`
   (drift path included); approval takes the population lock; terms-close/disposition take the
   input lock; `status_history` INSERT revoked-or-triggered; `ensure_pack_blocked` takes the
   pack lock; draft-line writes locked; `pay_period`/`employer` UPDATE revoked/routed (+ header
   compare in `tipDrifted`). Map capture run-lock 40001 → conflict. Re-verify: all §11 live
   races writer-first asserting inclusion (incl. the 4 missing classes) + T2-validation proof.
2. E3/F1: REVOKE app INSERT on `payroll_snapshot*` behind a DEFINER capture function; revoke app
   UPDATE(`status`,`inputs_hash`,`rule_versions`) behind a DEFINER validate function; harden
   `check_run_lifecycle` (pins = tip, ≥1 employee). Re-verify: E3/F1A/F1B live batteries as app.
3. Comp: chain-load/span-carve split (WIDEN); liveness alignment (FREETAIL);
   `COMP_MID_PERIOD_CHANGE` blocker in readiness + snapshot + validation (MIDPERIOD) + tests.
   Re-verify: 3 live service-as-app probes + new resolve/readiness/snapshot tests.
4. Pack trust: source-ref schema + byte re-hash in `Verify` (C1-SOURCES); execute goldens /
   verify gate artifact (C1-GOLDENS); audited pin ceremony function/trigger (C2-REPIN);
   evidence-label match + effective-coverage semantics enforced before the first pin
   (C2-LABELS/EFFECTIVE). Re-verify: fabricated-digest + failing-golden rejections live.

Decisions (blocking, recorded with pinned rationale):

5. B1 active-evidence standard (enforce verification vs amend B1); X-FUND semantics (coded
   blocker vs pinned any-active-covers + live multi-fund matrix); B2 full-cover rule (union
   covers period or uncovered-span blocker + 9 tests).

Procedures/tests (blocking for acceptance):

6. Writer-first race tests added (D6-P2c set); SHG B4 completion (audit-absence per rejection,
   cross-scope negatives, drift + backdated-add probes); E/F concurrency pair
   (correction-during-capture, dual-capture); independent CLI-vs-API differential + 19-name
   list + skip table (V4 condition); clean-server/fresh-DB-or-content-guard for acceptance
   (H-STALE-SILENT); report corrections (§11 paragraph).

Pre-calculation (at calculator introduction, not today):

7. G-PREREQ 6-item `payroll_line` governance (snapshot FK, provenance, rounding rule,
   immutability trigger, versioned supersede, revoke direct writes).

## 17. Explicitly deferred issues

Non-blocking for the restricted first calculation (do not gate §15 once §16 is complete):

- P2 hardening: C2-ADMIN-DIRECT-WRITE trigger, C2-HISTORY-STATUS superseded marking,
  D5-P2a/P2b (fold D5-P2a into the D1 fix opportunistically), F-A2-TERMSLINK provenance pins.
- P3s: EF-F4 empty-actor CHECK, EF-F3 self-link CHECK, EF-E1N remainder after the P1g
  disposition, C1-SYMLINK/TOCTOU file handling, C4-TEST-WEAKNESS assertions + OT vs OT-001
  wiring doc, F-B4 remainder after §16 probes, H-GOOSE-VERWRITE revoke, H-EPIHEMERAL-PORT.
- Prior-review ID closure table (SHG-B1a/B1b/B1c, B2a/B2b/B2c, C1–C4, C6, S5-F1) + §10D tally
  reconciliation; `SnapSHG` value-level evidence pinning decision; C-level snapshot-fixation
  source read; cross-employer key-separation live test.
- All prompt16 §15 scope exclusions (CPF/SDL/SHG calc, OT, gross/net, payslips, rules engine,
  Phys-Pseudo, more countries, RLS/SSO, UI, microservices, event bus, advanced WFM).
  `SG-MOM-OT-001` remains BLOCKED; real statutory packs stay in actual evidence-backed states.
