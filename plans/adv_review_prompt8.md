# Milestone 1 Adversarial Review

**Repository:** `github.com/PithomLabs/a1` (WorkforceOps Lab)
**Baseline commit:** `f43515883668c917edac31be49a7cf8334eca5c2` (branch `main`, `HEAD`)
**Review date:** 2026-11-08
**PostgreSQL:** 16.15, session timezone UTC
**Disposable test database:** `postgresql://postgres@127.0.0.1:55432/a1_test` (container `wfops-review-pg` on 127.0.0.1:55432)
**API server during review:** `http://127.0.0.1:18081` (PID `/tmp/opencode/serve.pid`)

---

## 1. Repository State at Review

```text
HEAD = f43515883668c917edac31be49a7cf8334eca5c2 (main)
Uncommitted entries: 139 (git status --short)
New file to be created: plans/milestone1_adversarial_review.md
```

**Notable tracked changes vs baseline:**
- `cmd/payloop-api`, `cmd/shifthub-api`, `cmd/sync-worker` deleted (consolidation)
- `docs/workforceops/*.md` → `docs/workforceops/parked/*` (archived)
- `migrations/core/00001_schemas.sql` deleted, replaced by `00001_foundation.sql` + `00002_attendance_payroll.sql` + `00003_rules_audit.sql` + `00004_protection.sql` + `00005_id_defaults.sql` + `00006_trigger_return_fix.sql` + `00007_trigger_names.sql` + `00008_break_events.sql`
- `openapi.yaml`, `Taskfile.yml`, `AGENTS.md`, `README.md` modified

**Untracked files present:** `plans/aonymisation.pdf`, `plan2.md`, `plan3*.md`, `plan4*.md`, `plan_alt.md` — ignored.

---

## 2. Review Methodology and Commands Executed

| Step | Command | Purpose |
|------|---------|---------|
| 1 | `git status --short && git rev-parse HEAD` | Confirm baseline and working tree |
| 2 | `GOWORK=off task lint && GOWORK=off task build` | Lint and build (requires `GOWORK=off` due to parent `go.work`) |
| 3 | `GOWORK=off WORKFORCEOPS_TEST_DATABASE_URL=postgresql://postgres@127.0.0.1:55432/a1_test go test -p 1 -count=1 ./...` | Full test suite (serial, fresh DB) |
| 4 | `GOWORK=off WORKFORCEOPS_TEST_DATABASE_URL=postgresql://postgres@127.0.0.1:55432/a1_test go test -count=1 ./...` | Parallel test run (reproduces race) |
| 5 | Manual API calls via `curl` against live server | Verify error envelope, readiness, employer context |
| 6 | Direct `psql` queries against `a1_test` | Inspect schema, constraints, data lineage, triggers |
| 7 | Source code inspection (`grep`, `read`) | Verify service boundaries, validation, invariants |
| 8 | OpenAPI structural analysis (Python YAML parse) | Count paths, operations, maturity flags |

**Test results (serial `-p 1`):**
- All 13 test-bearing packages PASS (cmd/workforceops, core/api, core/config, core/database, core/module, internal/attendance, internal/hr, internal/httpapi, internal/payroll, internal/timesheet, internal/wfm, tests/arch, tests/integrity)
- `GOWORK=off task lint` PASS (gofmt, go vet, lib/pq guard)
- `GOWORK=off task build` PASS (single `build/workforceops` binary)
- Parallel test run (`go test -count=1 ./...`): **FAIL** — concurrent `CREATE DATABASE` / extension / migration duplicate-key errors (SQLSTATE 23505) across packages; the `dbutil.DSN()` helper creates a per-package database but races on extension creation and migration type registration.

**OpenAPI verification:**
- 36 paths, 46 operations, 46 unique `operationId`s
- 38 marked `x-maturity: implemented`, 8 marked `draft`
- Contract tests (`TestServedSubsetOfSpec`, `TestSpecImplementedAllServed`) PASS — served/implemented route sets match nominally.

---

## 3. Findings — Ranked by Severity

### P0 — Critical (blocks Milestone 1)

#### F-001: Approved timesheet diverges from corrected attendance (lineage break)
- **File/line:** `internal/attendance/review.go:242-249` (post-commit `upsertRecord`), `internal/timesheet/timesheet.go:290-295` (link at generation time only)
- **Invariant:** Approved artifacts must reflect the latest corrected attendance; corrections must not leave approved timesheets referencing superseded intervals without trace.
- **Scenario:** 1) Generate + submit + approve timesheet for PP-2026-08 (DEMO-001 = 40h). 2) Correct a linked interval (Fri work 18:00 → 17:30). 3) Attendance record becomes 39.5h; approved timesheet remains 40h with old interval still linked; new corrected interval is unlinked.
- **Why tests missed it:** `TestGenerateSubmitApprove` (internal/timesheet/timesheet_test.go:101) stops at approval; no test covers correction after approval. `TestMissingPunchAndCorrection` (internal/attendance/attendance_test.go:154) corrects before generation.
- **Reproduction:** Verified on `a1_test` — approved sheet `572981ad-...` (40h) vs attendance 39.5h; `timesheet_line_interval` links old interval `b198b282-...` (10h); corrected interval `9dad64d6-...` (9.5h) unlinked.
- **Correction:** On correction, if any linked timesheet is `approved`/`submitted`, either (a) reject correction with `CodeInvalidState` requiring explicit reopen, or (b) create a new `correction` kind timesheet version (per AGENTS.md §7). Must record the lineage break in `audit_event`.
- **Blocks:** Milestone 1 — approved data silently incorrect.

#### F-002: Overlapping pay periods double-count intervals
- **File/line:** `internal/timesheet/timesheet.go:252-257` (eligibility uses `local_date BETWEEN`), `migrations/core/00002_attendance_payroll.sql:87-100` (no overlap exclusion on `pay_period`)
- **Invariant:** One attendance interval must contribute to at most one pay-period timesheet.
- **Scenario:** Create `PP-2026-08-B` overlapping `PP-2026-08` (both 2026-08-01..31). Generate timesheets for both. Each interval links into both periods.
- **Why tests missed it:** No test creates overlapping pay periods. `ListPeriods` orders by `starts_on` but no uniqueness on date ranges.
- **Reproduction:** On `a1_test` — DEMO-001 has 20 links across 2 periods (11 distinct intervals × 2); DEMO-003 has 20 links across 2 periods (10 distinct × 2). Readiness for PP-2026-08 shows population 2 (DEMO-001, DEMO-003 approved) but each interval counted twice.
- **Correction:** Add `EXCLUDE USING gist (employer_id WITH =, daterange(starts_on, ends_on, '[]') WITH &&)` on `pay_period`. In `Generate`, filter intervals by `pay_period_id` via roster assignment period or explicit period membership, not by `local_date` alone.
- **Blocks:** Milestone 1 — fundamental payroll population integrity.

#### F-003: Readiness population omits non-active employees with approved period timesheets
- **File/line:** `internal/payroll/readiness.go:86-88` (queries `WHERE status = 'active'`)
- **Invariant:** Readiness must assess every employee who has an approved timesheet for the period, regardless of current status.
- **Scenario:** DEMO-003 terminated effective 2026-09-01. August readiness population = 2 (only active DEMO-001, DEMO-002). DEMO-003's approved August timesheet (35h) excluded from readiness.
- **Why tests missed it:** `TestAssessBlocked` (internal/payroll/readiness_test.go:109) creates 2 active employees; no test covers terminated employee with approved historical timesheet.
- **Reproduction:** On `a1_test` — readiness `population=2`, `included=["DEMO-001"]`, `missing_info=[{"employee_no":"DEMO-002","reason":"no approved timesheet"}]`. DEMO-003 absent despite approved sheet `2f704805-...`.
- **Correction:** Population query must be `SELECT id, employee_no FROM employee WHERE employer_id = $1 AND EXISTS (SELECT 1 FROM timesheet t WHERE t.employee_id = employee.id AND t.pay_period_id = $2 AND t.status = 'approved') UNION SELECT ...` — include any employee with an approved sheet for the period.
- **Blocks:** Milestone 1 — readiness gate incorrectly passes/fails.

#### F-004: Draft payroll run exclusion removes employee from unrelated period readiness
- **File/line:** `internal/payroll/readiness.go:182-207` (`loadExclusions` joins `payroll_run_exclusion` by `payroll_run_id` only, no period check)
- **Invariant:** Exclusions apply only to the period of the referenced run.
- **Scenario:** Draft run for `PP-2026-08-B` excludes DEMO-001. Readiness for `PP-2026-08` (different period) omits DEMO-001.
- **Why tests missed it:** `TestAssessWithExclusion` (readiness_test.go:136) uses same period for run and readiness; no cross-period test.
- **Reproduction:** On `a1_test` — run `PP-2026-08-B` draft excludes DEMO-001; readiness `PP-2026-08` population drops from 2 to 1.
- **Correction:** In `loadExclusions`, verify `payroll_run.pay_period_id = target_period_id` before applying exclusions.
- **Blocks:** Milestone 1 — readiness corruptible by unrelated draft runs.

#### F-005: Effective-dated employment terms and status are write-once (no revision path)
- **File/line:** `internal/hr/hr.go:197-228` (`SetTerms` only INSERT), `internal/hr/hr.go:118-158` (`SetStatus` only INSERT), `migrations/core/00004_protection.sql:205-217` (exclusion constraints)
- **Invariant:** Effective-dated data must support closing a prior row (setting `effective_to`) and opening a new row without overlap error.
- **Scenario:** 1) Employee created with status `active` — no `employee_status_history` row written. 2) `SetStatus` to `suspended` 2026-09-01 succeeds (first history row). 3) Second `SetStatus` to `resigned` 2026-10-01 fails `employee_status_history_no_overlap` (daterange `[2026-09-01,∞)` overlaps `[2026-10-01,∞)`). Same for terms — cannot end a terms row to start a new one.
- **Why tests missed it:** `TestSetStatus` (hr_test.go:107) only does one status change. `TestTermsOverlap` (hr_test.go:135) tests overlapping INSERTs, not UPDATE-to-close + INSERT.
- **Reproduction:** On `a1_test` — `employee_status_history` has only terminated/suspended rows; initial `active` never recorded. Second `SetStatus` for same employee fails. Terms second INSERT with overlapping range fails correctly, but no API to `UPDATE` existing row's `effective_to` then insert new.
- **Correction:** Add `CloseTerms` / `CloseStatus` service functions that `UPDATE effective_to` on current row, then allow new row insert. Or use `effective_to` = day before new `effective_from` in a single transaction.
- **Blocks:** Milestone 1 — HR data model cannot represent real employment changes.

#### F-006: Rejected timesheets are a dead end (no reopen/edit path)
- **File/line:** `internal/timesheet/timesheet.go:375-398` (`Reject` → `transition` from `submitted` to `rejected`), `migrations/core/00004_protection.sql:34-62` (trigger `forbid_if_timesheet_decided` freezes `approved`/`rejected`)
- **Invariant:** Rejected timesheet must be correctable and resubmittable per contract (§7: "workflow transitions only through their endpoints").
- **Scenario:** Timesheet rejected. `Submit` fails `timesheet not found or not draft`. `Generate` skips `timesheet already exists`. No CLI/API to reopen or create replacement.
- **Why tests missed it:** No test covers reject → correct → resubmit flow.
- **Reproduction:** CLI `timesheet reject <id>` → `timesheet submit <id>` → `Error: timesheet not found or not draft`. `timesheet generate` → `Skip: timesheet already exists`.
- **Correction:** Add `Reopen` transition (`rejected` → `draft`) or `Supersede` (new draft timesheet linking to same period with `kind='correction'`). Must preserve rejected sheet per immutability.
- **Blocks:** Milestone 1 — workflow incomplete.

#### F-007: Late break events cannot attach to already-paired work interval
- **File/line:** `internal/attendance/attendance.go:256-333` (`pairSpans` processes events in single pass; breaks only attach if seen before `out`), `internal/attendance/attendance.go:520-540` (`Import` calls `pairSpans` once per import)
- **Invariant:** Break evidence arriving after the initial import must update the attendance intervals (correction or re-pair).
- **Scenario:** Import in/out first → creates 09:00–18:00 work interval + `missing_break_evidence` exception. Later import break_start/break_end → produces `unmatched_event: break_start without preceding in`; 9-hour work interval unchanged.
- **Why tests missed it:** `TestMissingBreakEvidence` (attendance_test.go:206) imports all events at once; no test for staged imports.
- **Reproduction:** Manual CSV import in two stages — first without breaks, then with breaks. Exception remains, intervals not updated.
- **Correction:** On re-import, re-pair all unpaired events for touched employees/dates (already loads all unpaired events per `unpairedEvents`). But `pairSpans` emits `unmatched_event` for break_start without pending in — need to allow break attachment to existing work interval via correction or extended pairing logic.
- **Blocks:** Milestone 1 — attendance correction workflow broken for common case.

#### F-008: Cross-midnight intervals not split at local midnight (domain model violation)
- **File/line:** `internal/attendance/attendance.go:346` (`localDate := in.at.In(loc).Format("2006-01-02")` — single date from in-event), `docs/workforceops/domain-model.md:103-105` ("cross-midnight spans are split at local midnight into two attendance records")
- **Invariant:** Each `attendance_record` represents one business date; cross-midnight work must produce two records.
- **Scenario:** DEMO-003 night shift 22:00–06:00+1 day. All intervals (work + breaks) have `local_date = 2026-08-03` (from in-event 22:00). Attendance record for 2026-08-04 has 0 minutes despite 4 hours worked post-midnight.
- **Why tests missed it:** `TestImportPairReview` checks totals, not per-date split.
- **Reproduction:** On `a1_test` — intervals for DEMO-003 2026-08-03 show 22:00–02:00, 02:00–03:00, 03:00–06:00 all `local_date=2026-08-03`. Record 2026-08-04 has `actual_min=0` (except separate import next night).
- **Correction:** In `storeSpan`, split work/break runs at local midnight boundaries; create separate intervals and records per business date.
- **Blocks:** Milestone 1 — contradicts documented domain model.

#### F-009: UTC-based date filtering misclassifies local-midnight boundary events
- **File/line:** `internal/attendance/attendance.go:353` (roster assignment lookup uses `business_date = localDate`), `internal/timesheet/timesheet.go:252-256` (eligibility uses `local_date BETWEEN`), `internal/payroll/readiness.go:225-227` (terms check uses `p.starts_on` date)
- **Invariant:** All date boundaries must use employee/local timezone, not UTC.
- **Scenario:** Clock event at 2026-08-07T17:00:00Z (2026-08-08 01:00+08). `occurred_at::date` in UTC = 2026-08-07; local date = 2026-08-08. Attendance review/timesheet eligibility may include/exclude wrong day.
- **Why tests missed it:** All test fixtures use daytime UTC+8 events; no boundary cases.
- **Correction:** All date filtering must convert to employee timezone first (`occurred_at AT TIME ZONE tz::date`).
- **Blocks:** Milestone 1 — silent misclassification at DST/boundary.

#### F-010: Concurrent corrections create duplicate successors (race)
- **File/line:** `internal/attendance/review.go:170-241` (`Correct` checks `old.superseded` but no `FOR UPDATE`; window between check and insert)
- **Invariant:** At most one successor per interval.
- **Scenario:** Two simultaneous `Correct` calls for same interval. Both pass `superseded` check, both insert new interval with same `supersedes_id`.
- **Why tests missed it:** No concurrency test.
- **Reproduction:** On `a1_test` — two corrections of DEMO-003 2026-08-05 work interval both succeeded; created `d1b72310-...` (18:15) and `a81bf63a-...` (18:30) both superseding `df68b4c3-...`. Attendance record `actual_min=765`, `paid_min=705` (double-counted 255 min).
- **Correction:** `SELECT ... FOR UPDATE` on the interval row in `Correct`; or unique partial index `ON attendance_interval (employer_id, supersedes_id) WHERE supersedes_id IS NOT NULL`.
- **Blocks:** Milestone 1 — data corruption under concurrency.

#### F-011: No `audit_event` or `idempotency_key` writes outside seed deletes
- **File/line:** `internal/seed/seed.go:85-120` (only place with `INSERT INTO audit_event`/`idempotency_key`), `internal/httpapi/httpapi.go:513-541` (`rule-activations` uses direct SQL, bypasses service layer)
- **Invariant:** Every audited state change must write `audit_event`; every idempotent operation must consume `idempotency_key`.
- **Scenario:** Corrections, timesheet transitions, roster publish, clock import — none write `audit_event`. `Idempotency-Key` header accepted but ignored.
- **Why tests missed it:** `TestSpecTriggerNamesExist` (freeze_test.go:259) checks trigger names exist; no test verifies `audit_event` population.
- **Correction:** Implement `audit_event` writes in all correction/transition/publish paths. Enforce `Idempotency-Key` on mutating endpoints per contract.
- **Blocks:** Milestone 1 — audit trail incomplete; idempotency not enforced.

#### F-012: Money fields use Go `float64` (violates AGENTS.md §3)
- **File/line:** `internal/hr/hr.go:169,182,184,186,255` (`MonthlyBasic`, `NormalHoursWeekly`, `DailyHours`, `AmountMonthly` all `float64`)
- **Invariant:** No binary floating point for money; `NUMERIC(12,2)` storage, exact intermediates.
- **Scenario:** Any calculation using these fields (e.g., future payroll slice) will accumulate FP error.
- **Why tests missed it:** No calculation tests exist yet (all packs BLOCKED).
- **Correction:** Use `decimal`/`shopspring/decimal` or `big.Rat` in Go; keep `NUMERIC(12,2)` in DB; round only at statutory boundaries.
- **Blocks:** Milestone 1 — architectural violation before first calculation.

---

### P1 — High (blocks Milestone 2, should fix before next milestone)

#### F-101: `403 employer_context_rejected` documented but never implemented
- **File/line:** `openapi.yaml:12`, `docs/workforceops/api-cli-contract.md:85`, `internal/employer/employer.go:43-80` (`Resolve` accepts any code/UUID existing in DB)
- **Invariant:** Requests outside configured employer context must return 403.
- **Scenario:** Multi-employer DB — employer A requests `/api/v1/employers/B/employees`. Currently resolves B and returns data.
- **Why tests missed it:** `TestCrossEmployerAssignmentRejected` (isolation_test.go:14) tests assignment FK, not API employer scoping.
- **Correction:** In `resolve()`, verify `h.employer.ID` matches the resolved employer from path (already done via `Resolve`), but add configurable allow-list or context binding. For v0.1 single-employer, document as accepted limitation; for Milestone 2, implement.
- **Blocks:** Milestone 2.

#### F-102: Default runtime DSN uses superuser `a1` (bypasses least-privilege revokes)
- **File/line:** `.env.example:17` (`DATABASE_URL=postgresql://a1:change-me@...`), `scripts/db/init-roles.sh:30-36` (`app` role granted DML; `migration_admin` owns schema)
- **Invariant:** Runtime must use `app` role; migrations use `migration_admin`.
- **Scenario:** Default `.env` uses superuser `a1` which owns schema and bypasses all revokes/triggers. `app` role only used by integrity tests.
- **Why tests missed it:** Integration tests use `WORKFORCEOPS_TEST_DATABASE_URL` with `a1` user; no test verifies runtime role.
- **Correction:** Update `.env.example` to use `app` role (password via `A1_DEV_ROLE_PASSWORD`). Document migration vs runtime DSN separation.
- **Blocks:** Milestone 2.

#### F-103: `MIGRATION_DATABASE_URL` is dead config
- **File/line:** `cmd/workforceops/migrate.go:57-62` (`mustDSN` uses only `--dsn` or `DATABASE_URL`), `.env.example:19` (`MIGRATION_DATABASE_URL=` empty), `docs/workforceops/database-design.md:18`
- **Invariant:** Migrations must run as `migration_admin` role.
- **Scenario:** `task db:migrate` uses `DATABASE_URL` (superuser `a1`), not `MIGRATION_DATABASE_URL`.
- **Correction:** Implement `MIGRATION_DATABASE_URL` fallback in `mustDSN`; update docs and Taskfile.
- **Blocks:** Milestone 2.

#### F-104: Seed `demo --reset` fails under `app` role (permission denied) and as superuser (freeze trigger)
- **File/line:** `internal/seed/seed.go:78-120` (`resetEmployer` deletes in dependency order), `migrations/core/00004_protection.sql:70-80` (`forbid_if_timesheet_decided` triggers on `timesheet`/`timesheet_line`/`timesheet_line_interval`), `migrations/core/00004_protection.sql:183-185` (`attendance_correction` delete blocked?)
- **Invariant:** `seed demo --reset` must work for development iteration.
- **Scenario:** As `app` role: `permission denied for table attendance_correction (SQLSTATE 42501)`. As superuser: `frozen: timesheet ... already decided` (trigger fires on `timesheet_line_interval` delete).
- **Why tests missed it:** No test runs seed reset.
- **Correction:** Add `TRUNCATE ... CASCADE` with `SET session_replication_role = replica` for reset, or order deletes to respect triggers (delete timesheet_line_interval before timesheet_line before timesheet). Grant `app` delete on all tables.
- **Blocks:** Milestone 2.

#### F-105: WFM rest-day and week-boundary validation missing
- **File/line:** `internal/wfm/roster.go:220-322` (`Assign` validates team membership + skills only), `migrations/core/00001_foundation.sql:336-350` (roster_assignment has `rest_day` CHECK but no week/date validation)
- **Invariant:** Assignments must fall within roster week and respect rest days.
- **Scenario:** Assigned DEMO-001 to Sunday 2026-08-16 (outside 2026-08-10 roster week) with work template — accepted and published. Assigned 2026-08-20 (next week) — accepted.
- **Why tests missed it:** `TestRosterPublishFreeze` (wfm_test.go:153) tests freeze, not assignment validation.
- **Correction:** In `Assign`, verify `business_date BETWEEN roster.week_starting AND roster.week_starting + 6`. If `rest_day=true`, reject work template.
- **Blocks:** Milestone 2.

#### F-106: Published roster revisions never get `published_at`/`published_by`; no reopen/new-revision service
- **File/line:** `internal/wfm/roster.go:382-384` (`UPDATE roster SET published_at, published_by`), `migrations/core/00001_foundation.sql:307-308, 321-322` (`roster_revision` has `published_at`, `published_by` but never set), `internal/wfm/roster.go` (no `Reopen` or `NewRevision` function)
- **Invariant:** Published revision must carry publication metadata; published roster must be reopenable to draft for corrections.
- **Scenario:** Publish roster → `roster.published_at`/`published_by` set, but `roster_revision` (v1) remains NULL. No API to reopen or create v2.
- **Why tests missed it:** `TestRosterPublishFreeze` checks assignments frozen, not revision metadata.
- **Correction:** In `Publish`, also `UPDATE roster_revision SET published_at=now(), published_by=$3 WHERE id=current_revision_id`. Add `Reopen` (set roster status=`draft`, increment `current_version`, insert new revision copying assignments).
- **Blocks:** Milestone 2.

#### F-107: Rotation assignments stored but never materialized into `roster_assignment`
- **File/line:** `internal/wfm/rotation.go:161-166` (`CreateRotationAssign` inserts `rotation_assignment` only), `internal/wfm/roster.go` (no rotation→roster materialization), `migrations/core/00001_foundation.sql:183-190` (`programme_staffing_requirement` table exists, unused)
- **Invariant:** Rotation patterns must drive roster assignments for their effective dates.
- **Scenario:** Rotation assigned to employee/team; roster creation requires manual `Assign` per day. No automation.
- **Why tests missed it:** `TestRotationPatternAndAssignment` (rotation_test.go:7) only tests storage.
- **Correction:** Add `MaterializeRotation` service to expand pattern assignments into `roster_assignment` for a roster week. Wire into roster create/publish.
- **Blocks:** Milestone 2.

#### F-108: `timesheet_line_interval` lacks unique `(employer_id, interval_id)` and employee-consistency enforcement
- **File/line:** `migrations/core/00002_attendance_payroll.sql:134-141` (PK = `(employer_id, line_id, interval_id)`; no unique on `interval_id` alone), `internal/timesheet/timesheet.go:290-295` (links without checking interval's employee matches timesheet's employee)
- **Invariant:** One interval cannot belong to two timesheet lines (even across employees/periods).
- **Scenario:** Overlapping periods (F-002) demonstrate double-linking. Cross-employee link possible if intervals share UUID space (unlikely but not prevented).
- **Correction:** Add `UNIQUE (employer_id, interval_id)` on `timesheet_line_interval`. In `generateOne`, assert `i.employee_id = employeeID`.
- **Blocks:** Milestone 2.

#### F-109: `attendance_record.status` lacks CHECK constraint
- **File/line:** `migrations/core/00002_attendance_payroll.sql:50` (`status TEXT NOT NULL DEFAULT 'reviewed'` — no CHECK)
- **Invariant:** Status must be constrained to known values.
- **Correction:** Add `CHECK (status IN ('reviewed', 'exception', ...))`.
- **Blocks:** Milestone 2.

#### F-110: `attendance_interval.supersedes_id` lacks uniqueness (allows multiple successors)
- **File/line:** `migrations/core/00002_attendance_payroll.sql:22,29` (`supersedes_id uuid`, FK only), F-010 shows duplicate successors possible.
- **Correction:** Add `UNIQUE (employer_id, supersedes_id) WHERE supersedes_id IS NOT NULL`.
- **Blocks:** Milestone 2.

#### F-111: OpenAPI error envelope mismatch (runtime vs contract)
- **File/line:** `internal/httpapi/httpapi.go:581` (`{"error": code, "message": msg}`), `openapi.yaml:818-830` (requires `{error, code, request_id}`)
- **Invariant:** Runtime error body must match OpenAPI `Error` schema.
- **Scenario:** `GET /employers/MERIDIAN/employees/NOPE` → `{"error":"not_found","message":"employee NOPE not found"}` + `X-Request-Id` header. Contract expects `{"error":"...","code":"not_found","request_id":"..."}`.
- **Why tests missed it:** Contract test validates routes, not response body shape.
- **Correction:** In `fail()`, return `map[string]string{"error": msg, "code": code, "request_id": requestID}`. Extract `X-Request-Id` from Echo context.
- **Blocks:** Milestone 2.

#### F-112: Parity test covers only 4 legs (not 6 as claimed)
- **File/line:** `cmd/workforceops/parity_test.go:127-203` (legs: 1=employee create/read, 2=status, 3=catalogue, 4=roster publish), `plans/plan_prompt7_imp.md:84` ("6/6 MATCH (employee, publish, import, review, sheet, readiness)")
- **Invariant:** Claimed parity must be evidenced.
- **Scenario:** Import/review/sheet/readiness not in parity test.
- **Correction:** Update report claim to "4/4 legs" or add missing legs to test.
- **Blocks:** Milestone 2 (credibility).

#### F-113: `Idempotency-Key` not enforced (contract §8 requires it on publish, approve, import, etc.)
- **File/line:** `openapi.yaml` (no `Idempotency-Key` parameter on mutating endpoints), `internal/httpapi/httpapi.go` (no middleware), `internal/seed/seed.go` (only seed uses `idempotency_key` table)
- **Correction:** Add `Idempotency-Key` header validation middleware; consume key on mutating endpoints; return 409 on replay.
- **Blocks:** Milestone 2.

#### F-114: No attendance interval listing endpoint (corrections need interval ID)
- **File/line:** `openapi.yaml` (no `GET /attendance/{date}/intervals` or similar), `docs/workforceops/api-cli-contract.md` (acknowledged limitation)
- **Impact:** API consumer cannot discover interval ID to submit correction without SQL access.
- **Correction:** Add `GET /employers/{eid}/attendance/{date}/intervals` returning intervals for a date.
- **Blocks:** Milestone 2.

---

### P2 — Medium (operational limitations, fix during Milestone 2)

#### F-201: Bare JSON arrays instead of collection envelope (contract §8 variance)
- **File/line:** `internal/httpapi/httpapi.go` (e.g., `listEmployees` returns `[]Employee`), `openapi.yaml` (defines `Paginated` but not used)
- **Impact:** No pagination, no metadata; clients must handle raw arrays.
- **Status:** Documented as Milestone 1 variance (rev. 3). Acceptable for Milestone 1.

#### F-202: Rotation patterns CLI-only (no REST endpoints)
- **File/line:** `cmd/workforceops/wfm.go` (rotation commands), `internal/httpapi/httpapi.go` (no rotation handlers), `openapi.yaml:344,374` (POST/GET rotations marked `implemented` but not served — contract test would fail if served)
- **Correction:** Add REST handlers or mark spec `draft`.

#### F-203: Skill catalogue listing CLI-only
- **File/line:** `cmd/workforceops/hr.go:115` (`skill list`), `internal/httpapi/httpapi.go` (POST `/skills` only), `openapi.yaml:303` (POST only)
- **Correction:** Add `GET /skills`.

#### F-204: Leave requests draft endpoints (GET/POST/approve) in spec but not implemented
- **File/line:** `openapi.yaml:544-573` (3 operations `draft`), `internal/httpapi/httpapi.go` (no leave handlers)
- **Correction:** Mark spec `draft` (already) — do not serve until implemented.

#### F-205: Payroll runs/payslips/register draft in spec, not implemented
- **File/line:** `openapi.yaml:706-747` (`draft`), `internal/httpapi/httpapi.go` (no handlers)
- **Status:** Correctly `draft`.

#### F-206: Rule packs endpoint `draft` in spec but served as `implemented`
- **File/line:** `openapi.yaml:771` (`x-maturity: draft` for `GET /rule-packs`), `internal/httpapi/httpapi.go:513-541` (serves `/rule-activations` as `implemented`), contract test PASS because `/rule-packs` not in served routes.
- **Correction:** Align maturity or implement `/rule-packs`.

#### F-207: No interval-listing endpoint for corrections workflow
- **See F-114.**

#### F-208: UTC timezone handling in attendance review/timesheet eligibility
- **See F-009.**

#### F-209: Demo workflow produces 39.5h only via manual correction (not in seed)
- **File/line:** `internal/seed/seed.go:273-310` (clock events for DEMO-001 09:00–18:00 × 5 days = 40h; correction not in seed), `plans/plan_prompt7_imp.md:78` ("Timesheet approved with one REG line of 39.5 hours" — achieved only after manual `attendance correct`)
- **Correction:** Either update seed to include the correction, or update claim to 40h.

---

### P3 — Low (maintainability, documentation)

#### F-301: `go.work` parent breaks `task lint`/`task build` without `GOWORK=off`
- **File/line:** `/home/chaschel/Documents/go/go.work` (covers `phys*` modules), `Taskfile.yml` (no `GOWORK=off` prefix)
- **Impact:** Contributor friction; documented in `docs/workforceops/repository-audit.md:24`.
- **Correction:** Add `GOWORK=off` to Taskfile commands or remove parent work file from workspace.

#### F-302: Legacy Taskfile tasks (`oracle:*`, `scenario:*`, `investigate`, `demo`) remain
- **Status:** Documented as known limitation (out of milestone scope).

#### F-303: `bchat/` read-only reference not enforced by arch test
- **File/line:** `tests/arch/boundary_test.go` (checks `core` doesn't import `modules`, `lib/pq` forbidden) — no check for `bchat` imports.
- **Correction:** Add `bchat` to forbidden imports.

#### F-304: `attendance_interval` CHECK constraint on `source_refs` uses JSONB logic (complex, may have edge cases)
- **File/line:** `migrations/core/00002_attendance_payroll.sql:30-37`
- **Impact:** Hard to debug; consider application-level validation + simpler DB constraint.

---

## 4. Independent Verification of Implementation Report Claims

| Claim in `plan_prompt7_imp.md` | Verified? | Evidence |
|---|---|---|
| 36 paths / 46 ops / 38 implemented | **YES** | Python parse of `openapi.yaml`: 36 paths, 46 ops, 46 unique IDs, 38 implemented, 8 draft |
| `task lint` green | **YES** | `GOWORK=off task lint` PASS |
| `task build` green | **YES** | `GOWORK=off task build` PASS |
| `go test -count=1 ./...` green twice on scratch PG16 | **NO** | Parallel run FAILs (CREATE DATABASE race); serial `-p 1` PASS |
| 13/13 packages green | **YES (serial)** | All 13 test-bearing packages PASS with `-p 1` |
| CLI↔API parity test (6/6 normalized matches) | **NO** | Parity test has 4 legs only (employee, status, catalogue, roster); import/review/sheet/readiness not tested |
| Timesheet approved with 39.5h REG line | **CONDITIONAL** | 39.5h only after **manual** correction (shorten Fri 18:00→17:30); seed produces 40h |
| Readiness false, 8 blocked, 1 missing input | **PARTIAL** | On fresh seed: readiness `ready=false`, `verdict="not ready: no included employees with complete input"`, 9 packs BLOCKED, 1 missing (E2E-001 no SHG). Report's "8 blocked" not reproduced. |
| No statutory calculations, all 9 packs BLOCKED | **YES** | Verified: all activations `BLOCKED`, OT-001 `BLOCKED` |
| 72-hour logic only excess-review flag | **YES** | No `min(H,72)` payable-hours code found |
| Single binary, CLI/REST parity, no duplicated logic | **YES** | Single `workforceops` binary; handlers call `internal/*` services |

---

## 5. Verdicts

### A. Milestone 1 Functional Correctness — **FAIL**
The happy-path demo (employee → roster → clock import → review → timesheet → approve) works **only if** no corrections, no overlapping periods, no status/terms changes, and no late-breaking events occur. Four P0 findings (F-001, F-002, F-003, F-006) demonstrate that approved artifacts diverge from source data, intervals double-count across periods, readiness omits eligible employees, and rejected timesheets cannot be corrected. These are not edge cases — they are core workflow failures.

### B. Database Integrity — **FAIL**
Employer isolation holds for single-employer operation (composite FKs, triggers). However:
- Cross-midnight split missing (F-008) violates domain model.
- `supersedes_id` uniqueness missing (F-010/F-110) allows duplicate successors.
- `timesheet_line_interval` allows one interval in multiple lines (F-108).
- `attendance_record.status` unconstrained (F-109).
- Effective-dated revisions cannot close prior rows (F-005).
- No `audit_event` writes (F-011).
- Float64 money fields (F-012) violate storage invariant.
The schema has strong foundations but critical gaps for correction lineage and effective-dating.

### C. CLI/API Usability — **PASS WITH CONDITIONS**
- CLI covers all documented commands; JSON output; operator attribution; file-based input.
- REST serves 38 implemented operations; contract tests PASS; loopback-only binding.
- **Conditions:**
  1. Error envelope mismatch (F-111) — client cannot rely on `code` field.
  2. No interval listing (F-114) — corrections need SQL access.
  3. Rotation/skill CLI-only (F-202/F-203) — REST consumers missing operations.
  4. `Idempotency-Key` not enforced (F-113) — replay safety absent.
  5. Seed reset broken (F-104) — dev iteration requires superuser + manual cleanup.
  6. Default runtime DSN uses superuser (F-102) — least-privilege not default.

### D. Payroll Readiness Integrity — **FAIL**
- Readiness correctly refuses to compute money while all packs BLOCKED.
- **But** readiness population logic is broken (F-003, F-004): omits terminated employees with approved sheets; corrupted by unrelated draft run exclusions. The gate cannot be trusted to reflect actual payroll population.

---

## 6. Overall Recommendation: **FAIL**

**Rationale:** Four P0 findings (F-001, F-002, F-003, F-006) represent material defects in the core employee→timesheet→readiness workflow that prevent a trustworthy Milestone 1 gate. The happy path works only in the absence of corrections, overlapping periods, status changes, and rejections — which are normal operations. Database integrity has critical gaps in correction lineage and effective-dating. Payroll readiness, the milestone's exit gate, produces incorrect population counts.

**Accepted v0.1 limitations (do not block):**
- No authentication, loopback-only, operator attribution not identity.
- Bare arrays, no pagination.
- `Idempotency-Key` not enforced.
- CLI-only rotations/skills.
- No interval listing endpoint.
- No statutory calculations (all packs BLOCKED).
- Legal validation unavailable.

**Unacknowledged gaps that are defects (must fix):**
- Effective-dated revisions write-once (F-005).
- Rejected timesheet dead end (F-006).
- Late break attachment (F-007).
- Cross-midnight split missing (F-008).
- UTC date filtering (F-009).
- Concurrent correction race (F-010).
- No audit_event/idempotency_key (F-011).
- Float64 money (F-012).
- 403 employer_context_rejected missing (F-101).
- Default runtime superuser (F-102).
- MIGRATION_DATABASE_URL dead (F-103).
- Seed reset broken (F-104).
- WFM validation gaps (F-105).
- Roster revision metadata + reopen (F-106).
- Rotation not materialized (F-107).
- Interval link constraints (F-108/F-109/F-110).
- Error envelope mismatch (F-111).
- Parity test overstated (F-112).

---

## 7. Minimal Remediation Sequence

### Must fix before Milestone 2 (gate for next milestone)

1. **F-001** — Correction after approval: reject correction if linked timesheet `approved`/`submitted`, or create correction-timesheet version.
2. **F-002** — Pay period overlap exclusion constraint + generation period filter.
3. **F-003** — Readiness population includes any employee with approved period timesheet.
4. **F-004** — Exclusions scoped to run's pay period.
5. **F-005** — Add `CloseTerms`/`CloseStatus` (UPDATE `effective_to` + INSERT new) in single transaction.
6. **F-006** — Add `Reopen` transition (`rejected`→`draft`) or `Supersede` (new `correction` kind timesheet).
7. **F-007** — Re-pair on re-import: attach late breaks to existing work intervals via correction.
8. **F-008** — Split cross-midnight intervals at local midnight in `storeSpan`.
9. **F-010** — `SELECT FOR UPDATE` in `Correct`; add unique partial index on `supersedes_id`.
10. **F-011** — Implement `audit_event` writes in all mutation paths; add `Idempotency-Key` middleware.
11. **F-012** — Replace `float64` money with `decimal.Decimal` (or `big.Rat`) in `internal/hr`.
12. **F-101** — Implement employer context allow-list or document as single-employer only.
13. **F-102** — Update `.env.example` runtime DSN to `app` role.
14. **F-103** — Implement `MIGRATION_DATABASE_URL` fallback in `migrate.go`.
15. **F-104** — Fix seed reset: order deletes to respect triggers; grant `app` DELETE.
16. **F-105** — Add week-boundary and rest-day validation in `Assign`.
17. **F-106** — Set `roster_revision.published_at`/`published_by` on publish; add `Reopen` service.
18. **F-108/F-109/F-110** — Add unique constraints: `timesheet_line_interval(employer_id, interval_id)`, `attendance_interval(employer_id, supersedes_id) WHERE supersedes_id IS NOT NULL`, `attendance_record.status` CHECK.
19. **F-111** — Fix error envelope to match OpenAPI `{error, code, request_id}`.

### Should fix during Milestone 2

20. **F-009** — All date filtering in employee timezone.
21. **F-107** — Rotation materialization service.
22. **F-113** — Enforce `Idempotency-Key` on mutating endpoints.
23. **F-114** — Add attendance interval listing endpoint.
24. **F-201** — Implement `Paginated` envelope for list endpoints.
25. **F-202/F-203** — Add REST endpoints for rotation patterns and skill catalogue.
26. **F-206** — Align `/rule-packs` maturity or implement.
27. **F-209** — Seed includes correction for 39.5h or update claim.

### Can safely defer

28. **F-301** — `GOWORK=off` in Taskfile or remove parent work file.
29. **F-302** — Legacy Taskfile tasks cleanup.
30. **F-303** — Arch test forbid `bchat` imports.
31. **F-304** — Simplify `attendance_interval.source_refs` CHECK.

---

## 8. Items Not Independently Verified

The following were **not** fully verified in this review due to scope/time:

1. **Full REST coverage of all 38 implemented operations** — only contract tests (route existence) and spot checks (readiness, error envelope, employer resolve) executed.
2. **Multi-employer API 403 behavior** — tested only single-employer; F-101 confirmed missing.
3. **`app` role execution of all service/parity tests** — tests run as superuser (`a1`); only integrity tests create `app` role.
4. **Crash/partial-transaction behavior** — no kill-testing of in-flight transactions.
5. **Exhaustive concurrency scenarios** — only F-010 reproduced; other races (simultaneous publish, generate, import) not tested.
6. **OpenAPI external validator** — used structural parse only; not validated against `openapi-spec-validator` or similar.
7. **Statutory rules implementation** — all 9 packs `BLOCKED`; no calculation code exists (correct per design).
8. **Migration upgrade from baseline** — tested fresh install only; upgrade path not tested.
9. **Cross-midnight roster assignment lookup** — `storeSpan` uses `local_date` from in-event; roster assignment for next business date not found (silent `assignmentID=nil`).
10. **CSV import idempotency across employers** — same `source_event_id` under different employers accepted (different `employer_id` partition); not tested.

---

## 9. Appendix: Key Evidence Commands

```bash
# Baseline & repo state
git rev-parse HEAD
git status --short

# Test suite (serial required)
GOWORK=off WORKFORCEOPS_TEST_DATABASE_URL=postgresql://postgres@127.0.0.1:55432/a1_test go test -p 1 -count=1 ./...

# Parallel race reproduction
GOWORK=off WORKFORCEOPS_TEST_DATABASE_URL=postgresql://postgres@127.0.0.1:55432/a1_test go test -count=1 ./...

# Lint & build
GOWORK=off task lint && GOWORK=off task build

# OpenAPI counts
python3 -c "import re,yaml; t=yaml.safe_load(open('openapi.yaml')); p=len(t['paths']); ops=sum(len(v) for v in t['paths'].values()); m=[op.get('x-maturity') for v in t['paths'].values() for op in v.values()]; print(p, ops, m.count('implemented'), m.count('draft'))"

# API error envelope
curl -sS -D - http://127.0.0.1:18081/api/v1/employers/MERIDIAN/employees/NOPE

# Readiness with period
curl -sS "http://127.0.0.1:18081/api/v1/employers/MERIDIAN/payroll-readiness?pay_period=PP-2026-08" | python3 -m json.tool

# DB inspection
psql "postgresql://postgres@127.0.0.1:55432/a1_test" -Atc "select conname, pg_get_constraintdef(oid) from pg_constraint where conrelid='attendance_interval'::regclass and contype in ('u','x');"
psql "postgresql://postgres@127.0.0.1:55432/a1_test" -Atc "select e.employee_no, r.actual_min, r.paid_min from attendance_record r join employee e on e.id=r.employee_id where r.actual_min > 600;"
psql "postgresql://postgres@127.0.0.1:55432/a1_test" -Atc "select i.id, i.local_date, i.interval_start at time zone i.timezone, i.interval_end at time zone i.timezone from attendance_interval i join employee e on e.id=i.employee_id where e.employee_no='DEMO-003' and i.local_date='2026-08-03' order by i.interval_start;"
```

---

## 10. Cleanup

- Stop API server: `kill $(cat /tmp/opencode/serve.pid) 2>/dev/null || true`
- Remove scratch container: `docker rm -f wfops-review-pg`
- Leave pre-existing `build/workforceops` alone (git-ignored).
- `git status` confirms only `plans/milestone1_adversarial_review.md` added.

---

**End of Report**