## Goal

Execute prompt11 (Milestone 1.2 — Payroll Input Integrity Closure): fix the eight service-reachable P1 defects, close the ten `app`-role database bypasses, repair fresh-install and employer onboarding, apply the five scoped Stage E corrections, and verify the full evidence chain on disposable PostgreSQL 16 — without touching frozen architecture, applied migrations, committed history, or statutory calculations.

## Success Criteria

- All 16 prompt11 §10 gate conditions hold, evidenced by live runs (not green unit tests alone): no silent approved-input mutation, correct cross-midnight attribution, host-TZ independence, no 72h truncation, no publishable shifts for non-active staff, all listed bypasses closed, `pg_temp` exploit fails, fresh-install path works with role separation, SQL-free onboarding on CLI+REST, atomic audit, parity intact, no new calculations, OT-001 still BLOCKED.
- `plans/milestone1_2_remediation_report.md` delivered with all 14 required sections and an honest PASS/FAIL.
- Working tree still uncommitted; `00001`–`00009` byte-untouched; `bchat/`, Phys-Research/Phys-Pseudo, Neon/Fly untouched.

## Approach

Follow the owner's execution priority (prompt11 §12): attendance correctness → historical integrity → database protection → install/onboarding → full regression. Order inside the plan respects real dependencies:

- **B5 (TZ) before B2/B6/B7 tests**, so all new temporal fixtures are written TZ-clean and validated under both timezones from birth.
- **B3 (chain-aware Correct) before B7 (late-break desync)**, because both share one new linkage-policy helper: submitted/approved links block unless a draft successor on the same chain can receive corrected evidence.
- **B6 (Correct split) absorbs E2 (unrostered moves)** and the P3 both-dates audit row, since Correct is being rewritten to return multiple successors anyway.
- **Stage C lands as a single `00010` migration** after services are fixed, so trigger semantics are written against final service behavior; the app-role SQL matrix (§9B) runs against it.
- **Stage D uses the fixed roles**; the committed E2E harness runs last as the closing gate.

Key design decisions (all grounded in code read this turn):

1. **B1 overlap disposition**: new `clock_event_disposition` table (`event_id`, verdict `duplicate_confirmed`, reason, operator, decided_at). `checkEligible` (`timesheet.go:295`) excludes dispositioned events from the unpaired count; pairing skips dispositioned events so they can never allocate later; `overlapping_span` resolves only when all its events are dispositioned; service + CLI + REST with reason/operator, idempotent re-confirm, audit row. Genuine overlaps stay blocking.
2. **B3/B7 linkage policy**: one helper — an approved link blocks Correct/attach unless it is non-effective or has an open draft successor on its chain, in which case the correction proceeds and the draft regenerates in-txn. Readiness gains an attendance-conflict check mirroring `checkTermConflict` (`readiness.go:324`): effective sheet's evidence vs current intervals.
3. **B6 Correct split**: reuse `splitRun` in `Correct`; returns one successor per dated run (signature `Correct(...) (string, ...)` becomes `([]string, ...)`; CLI/REST/tests updated); per-date roster resolution, NULL + `no_roster_link` when absent (E2); both day records upserted with their own assignments; audit rows on both dates.
4. **B5 TZ**: thread the event's declared timezone into `classifyBreak` (`attendance.go:946`), `time.LoadLocation` for anchors, never `start.Location()`; audit all `time.Local`/decode paths; run the attendance/timesheet suites under `TZ=UTC` and `TZ=Asia/Singapore` with identical fixtures plus a mixed-zone input test.
5. **B4 comp-retro**: mirror `SetTerms`' approved-period retro block (`hr.go:362`) in `AddCompensation`; extend the conflict check to `compensation_line` (effective-overlap + `created_at`); validate comp range inside parent terms range (req 7).
6. **B8 status gate**: effective-dated status lookup as-of each roster date inside shared `validatePlacement` (`roster.go:338`), covering Assign and Publish together; historical shifts untouched.
7. **Stage C single `00010`**: unconditional `NEW.is_current=true` + column-level INSERT revoke (Go never writes the column — verified); schema-qualified DEFINER body + restricted `search_path` + `REVOKE TEMPORARY` (no temp-table usage in Go — verified); employee-match + self-CHECK on supersession; single-chain trigger to INSERT OR UPDATE; chain-scope guards; line/sheet parent-key freeze once links exist; roster `published_at` stamping (service fix — `roster.go:520` never stamps) + no-reopen/no-pointer-rollback triggers; C8 narrow-transition matrix (terms/status: only open-row end-date close; compensation/membership/skills/roster-assignments: no UPDATE/DELETE at all — no service writes them, verified for compensation); `EffectiveApproved` errors on multiple terminal approvals instead of `LIMIT 1`.
8. **D1**: rewrite role creation with `\gexec`-based plain SQL (no substitution inside `DO`); the regression test executes the actual script against disposable PG16.
9. **D2**: `.env.example` ships app `DATABASE_URL` + migration_admin `MIGRATION_DATABASE_URL`; `resolveCLI` and `serve` refuse superuser runtime connections loudly (`SHOW is_superuser`); docs updated.
10. **D3**: `employer.Create` service (no employer context) + `employer create` CLI + `POST /api/v1/employers`; single-employer rule enforced at service level (schema stays multi-employer-ready); audit row keyed to the new employer (audit `employer_id` is NOT NULL — verified); repeat create is an explicit conflict.
11. **E3 durations**: integer-minutes quantization at one Go helper (round-half-up, documented), exact decimal strings into unchanged NUMERIC columns; fractional-duration tests in attendance + timesheets.
12. **E5 reset**: `--confirm-db <name>` must equal the DSN target (typo guard) + pre-authenticate the migration DSN before any drop (covers the password-auth drop-then-fail) + existing protected-name/same-server checks.
13. **E2E harness committed** under `tests/e2e/` (prompt11 §9E requires the persisted-state comparison; the review correctly flags `/tmp` scripts as unverifiable).

## Steps

1. **Stage A reconfirmation.** Baseline record (`git status`, `rev-parse`, DB targets); re-verify each of the 8 + 10 + 3 findings against current code; classify per §3 (confirmed/fixed/regression/hypothesis/limitation); stop-and-resolve any fix that breaks a verified invariant (known interaction: B3 chain-aware Correct vs freeze immutability — successors only, originals never touched).
2. **B5 TZ hardening + B2 midnight breaks.** Thread tz through classification; split spanning breaks at local midnights in pairing and late-break attach; boundary fixtures (exact midnight, ±1s, paid/unpaid, duplicate, missing end, late delivery).
3. **B1 overlap disposition.** Table, service, CLI+REST, gate exemption, pairing skip, audit; E2E test per §4-B1.
4. **B6 Correct split (+E2, both-dates audit).** `splitRun` reuse, multi-successor return, per-date roster/NULL+exception, per-date records; full §4-B6 test matrix incl. pay-period-boundary crossing.
5. **B3 chain-aware Correct.** Linkage helper, draft-successor allowance path, in-txn regen; full B3 matrix (incl. concurrent correction-vs-reissue, rejected-correction recovery, readiness before/after).
6. **B7 late-break linkage.** Reuse B3 helper in `attachLateBreaks`; blocking conflict exception when no correction path; draft regen in-txn; full B7 matrix.
7. **B4 comp-retro + B8 status gate + E1 rest-day allowlist/CHECK + E3 duration policy.** Service changes, conflict-check extension, placement matrix tests (7 cases), fractional tests.
8. **E4 audit gaps.** `audit.Write` in CreateEmployee, membership, skill assignment (atomic); assertion tests that attribution persists.
9. **Stage C migration `00010` + privilege suite.** All C1–C8 guards; C9 per-guard verification (owner/DEFINER/search_path/qualification/privileges, positive+negative SQL, persisted-state readback, atomicity of failures); full §9B 16-case app-role matrix.
10. **Stage D install/onboarding.** D1 script fix + executes-the-script test; D2 env/Taskfile/README + superuser refusal + tests; D3 employer create + single-employer rule + tests; run the §9C 10-step fresh-install sequence twice (CLI leg + REST leg).
11. **E5 reset safety.** Confirm-name flag + pre-auth + negative tests (protected/unknown/wrong-server/bad-DSN/password-auth/success).
12. **Close-out.** Full §9E battery (`task lint`, `task build`, `go test -count=1 ./...`, five `-race` suites, integrity suite, install test, privilege test, CLI E2E, REST E2E, committed persisted-diff, OpenAPI validation, `git diff --check`); TZ=UTC full-suite run; contract/OpenAPI/setup-doc updates; `milestone1_2_remediation_report.md`; deferred-item record (§8 list + anything newly deferred with dependency justification).

## Validation Plan

- Each B-item has its prompt11-specified regression test(s) run against disposable PG16 as `app`; new temporal tests run under both `TZ=UTC` and `TZ=Asia/Singapore`.
- §9B matrix: 16 direct-SQL cases as `app`, each asserting persisted state after pass/fail (error-only assertions rejected per C9).
- §9C fresh-install: real `init-roles.sh` execution, catalog privilege assertions, employer create on both surfaces on separate fresh DBs, superuser-default refusal proof.
- §9D concurrency: nine race scenarios with one-winner verification; import/generate replay recovery without loss/duplication.
- Closing gate: committed `tests/e2e` persisted-diff (CLI vs REST byte-identical on normalized business state), OpenAPI validator output captured, `git diff --check` clean. Highest-risk validation: the B3/B7 chain-aware correction flow (deadlock fix must not reintroduce approved-input mutability — pinned by original-evidence-unchanged + readiness-conflict assertions).

## Risks / Open Questions

- **Risk:** C8 trigger strictness may break existing suites that UPDATE history directly as `app` (the review notes tests depend on this); mitigation: rework those fixtures to governed services, which is itself a hardening proof.
- **Risk:** `REVOKE TEMPORARY` or search-path restriction breaking goose/test helpers; mitigation: full suite re-run immediately after `00010`, before building on it.
- **Risk:** `Correct` signature change ripples through CLI/REST/tests; contained (uncommitted tree, no external consumers).
- **Non-goals (§8):** SHG lifecycle/verify, run state machine, rotation, generic idempotency/reads/pagination, SSO, RLS, cloud, daemon-dependent Docker work, all other P2/P3 except the E-stage five.
- **Open questions:** None — scope, groupings, priority order, and deferrals are fully specified by prompt11 and the owner's disposition; reversible choices above are marked as decisions.
