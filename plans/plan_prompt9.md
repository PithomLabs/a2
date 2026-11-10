## Goal

Execute `plans/prompt9.md` (Milestone 1.1): remediate the 2 P0 + 13 M1-blocking P1 findings from `plans/milestone1_adversarial_review.md`, adjudicate the 5 contradictory review claims with live reproductions, complete the §5 workflow/integrity corrections and §6 structural verifications, and deliver `plans/milestone1_1_remediation_report.md` with a PASS/FAIL verdict. No statutory calculations, no redesign, no expansion; work stays uncommitted over `f435158`.

## Success Criteria

The 10 conditions of prompt9 §8D, verifiable: both P0s fixed; all 13 blocking P1s fixed or disproved by controlled tests; all §3 discrepancies adjudicated; roles work on fresh install; correction/recovery workflows run without SQL; default `go test` reliable; CLI↔REST result equivalence; OT-001 still BLOCKED; zero new calculations or golden fixtures.

## Approach

Follow prompt9's stage order strictly: (1) baseline + read-only structural audit → `plans/milestone1_1_structural_audit.md`; (2) throwaway `/tmp` repro probes for §3A–E on disposable PG16, adjudicate each; (3) remediate in dependency order (schema guards → attendance → corrections/timesheets → HR → rosters → readiness → roles/plumbing → operability); (4) §6 verification battery; (5) CLI+REST E2E with persisted-result diff; (6) remediation report. Research already done: the full adversarial review (2 P0 · 24 P1 · 33 P2 · 13 P3), the earlier `adv_review_prompt8.md` reproduction claims, Taskfile/roles/Dockerfile/fly state, and file sizes of all target services.

## Key Decisions

1. **Migrations are append-only (`00009+`); `00001–00008` frozen.** Editing applied migrations would break goose tracking and the upgrade-path test. The `00004` fail-loud requirement is met instead by a `migrate` preflight in Go (assert `current_user=migration_admin`, `app` exists, grants as expected) plus a verification migration that raises if protections are missing. Legacy `f435158`-schema upgrade: first probe for any real user database (compose volume, documented dev DBs); if none exists, record an explicit pre-release reset decision per §2D and verify fresh-install only.
2. **P0-1: reject + reissue, not silent repair.** `Correct()` rejects intervals linked via `timesheet_line_interval` to a submitted/approved sheet. New controlled workflows: `reopen` (rejected→draft, also satisfies §5B) and `reissue` (approved→`superseded` plus a freshly generated draft linked via `supersedes_id`; old sheet immutable). Readiness ignores superseded sheets; trigger freeze extended to the new statuses with exemptions only for these transitions.
3. **P0-2: roster-anchored pairing window, orphans stay unpaired.** An `out` pairs only with the next unmatched `in` within 24h (covers night shifts; cross-midnight split per §3B if confirmed); anything beyond stays unpaired with a persisted span-breach exception — events never dropped, hours never truncated. Monthly `NeedsReview` wired into `checkEligible` as a skip. Dead 72h codes: emit the breach exception from the excess computation or delete them.
4. **Concurrency: advisory locks + partial exclusion.** Per-`(employer, employee)` advisory lock around Import/Correct/Generate pairing sections; `UNIQUE(supersedes_id)`; partial GiST exclusion on `tstzrange(start,end)` for live (non-superseded) intervals; service-level overlap checks as backstop. Aggregates (`upsertRecord`) moved inside the correction transaction.
5. **Money without new dependencies.** Decimal-string validation in Go (≥0, ≤2dp → 422), `int64` minor units internally, `NUMERIC(12,2)` in PG. No `shopspring/decimal`, no `math/big` unless an intermediate demands it.
6. **Test race: fix setup, keep parallel.** Tolerate 23505 collisions and advisory-lock `MigrateUp` in `dbutil`; `-p 1` only as a documented fallback, never silent.
7. **Roles end-to-end.** `migrate` honors `MIGRATION_DATABASE_URL`; runtime uses `DATABASE_URL`; fresh-install docs provision roles via `init-roles.sh`; service and parity suites run as `app` (reusing the integrity bootstrap; elevated access only for setup).
8. **Idempotency: domain mechanisms, honest contract.** Keep source-ID replay, status-conditional single-statement transitions, and generate-skip; correct the `Idempotency-Key` contract claim to match; leave the dead `idempotency_key` table documented-but-deferred per §9 (no generic framework).
9. **Rotation materialization (§6.8): document deferred.** No M1 consumer exists and deterministic generation is new behavior, not repair; note it explicitly rather than building it.
10. **No user questions.** Prompt9 plus the two review reports fully determine scope; anything genuinely requiring a frozen-decision change stops that item for approval per the audit instruction.

## Steps

1. **Baseline + audit (read-only).** `git status --short`, `rev-parse`, `diff --stat`; confirm 79-entry tree over `f435158`; record migration versions. Audit prompt9 §2A–G against code; write `plans/milestone1_1_structural_audit.md` with the 5-way classification. Confirm `modules/sg-wfm-payroll` dead via import graph (`go list`).
2. **§3 repro probes (`/tmp`, disposable PG16).** Overlapping periods → double allocation; overnight shift → business-date attribution; late break import → attachment; concurrent corrections → dual successors; fresh/upgrade/rollback migration paths. Each probe ends in a fixed / disproved / deferred verdict.
3. **Schema guards (new migration).** Period-range exclusion, `UNIQUE(interval_id)` allocation, employee-match enforcement on `timesheet_line_interval`, `UNIQUE(supersedes_id)`, partial live-interval overlap exclusion, freeze-trigger NEW-parent checks + employer-scoped lookups, re-parenting freeze tests as `app`.
4. **Attendance remediation.** Pairing window + span-breach exceptions; TZ-aware business-date comparisons everywhere (`AT TIME ZONE`); conflicting-redelivery conflict signal; unclassifiable breaks left unpaired with persisted exceptions; exception-clear only after validation pass; deterministic ordering; cross-midnight split if §3B confirms.
5. **Correction × timesheet lifecycle.** P0-1 reject + `reopen`/`reissue` services with CLI+REST; rejected-recovery flow `Generate→Submit→Reject→Correct→Resubmit→Approve`; cross-service regression tests; line/hours completeness in readiness (`SYN-P2-9`).
6. **HR + roster + readiness.** `SetStatus` close-prior-row + transition matrix (terminals final); terms close/successor atomicity, no overlaps; Assess population = union with approved-sheet holders, retroactive-change blocking; publish revalidation gate in one tx; `Revise` workflow (copy-forward, pointer advance, metadata); rotation `pattern list` shadowing fix.
7. **Roles, money, errors, operability.** Preflight + `MIGRATION_DATABASE_URL`; app-role suites + one app-role Generate→Approve smoke test; decimal money inputs; Go date/time validation → 422; envelope `{error, code, request_id}` + pg-text redaction; GET intervals endpoint + CLI; minimum onboarding ops (employer, pay-period, SHG instruction); Dockerfile/fly single-binary rebuild; stale spec/contract wording fixes.
8. **Verification + E2E + report.** §6 15-item battery with 4-way marking; §8A regression list; fresh-DB CLI+REST 12-step E2E with persisted-result diff (zero SQL); full gate set (`task lint`, `task build`, `go test -count=1 ./...`, integrity, migration fresh/upgrade, OpenAPI validation, `git diff --check`); write `plans/milestone1_1_remediation_report.md` with PASS/FAIL.

## Validation Plan

- Each fix ships with a fresh-PG regression test named for its finding ID; concurrency fixes get `-race` dual-invocation tests asserting exactly-once effects.
- Highest-risk validation: the app-role E2E on a fresh install (fresh roles → migrate preflight → seed → full workflow → readiness), since it exercises §2D, P1-7, and P1-24 jointly.
- Gates run exactly as prompt9 §8C lists; a skipped DB suite is a failed gate, never a pass.

## Risks / Open Questions

- Risk: if a real user database on the legacy schema exists, the pre-release reset decision is invalid and the upgrade path must be engineered instead — step 1 probes this before any migration work.
- Risk: `reissue` adds two statuses and a link column; if the audit finds a smaller controlled workflow (e.g. void-and-regenerate without linkage), it will be taken and the report will say so.
- Open questions: None — all material choices are resolved above from workspace evidence.
