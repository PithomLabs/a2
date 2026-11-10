# Milestone 1.1 — Structural Audit (`milestone1_1_structural_audit.md`)

Read-only audit per prompt9 §2. No code was modified. All file/line
citations below were opened and read in this run; review-report line
numbers were re-verified, not trusted.

## Baseline

- HEAD `f43515883668c917edac31be49a7cf8334eca5c2` ("Phase 1 scaffold baseline (pre-pivot)").
- Working tree: 79 entries (12 modified, 4 deleted, 35 renamed, 28
  untracked); `git diff --stat`: 12 files, +1231/−516. Identical to the
  tree reviewed by Prompt 8.
- Migrations on disk: `00001_foundation` … `00008_break_events` (8 versions).
- User-database probe: no `.env` in repo; `docker` daemon unavailable in
  this environment, so the compose volume could not be inspected; no other
  reachable local database was found. Evidence supports "no deployed user
  database", but the volume's contents are unverified — the pre-release
  reset decision must record this caveat, and **nothing existing will be
  dropped** (review amendment G). All verification uses fresh disposable DBs.

## Confirmed structural blockers (all fixable within the frozen architecture)

Each was re-verified by reading the cited body.

1. **Correction escapes timesheet freeze (SYN-P0-1).** `Correct`
   (`internal/attendance/review.go:147-251`) never consults
   `timesheet_line_interval`; superseding a linked interval succeeds while
   the approved line keeps stale hours. Remediation: reject + append-only
   correction versions per review amendment A.
2. **Unbounded pairing spans (SYN-P0-2).** `pairSpans`
   (`internal/attendance/attendance.go:256-333`) has no span-length guard;
   `ExcLimit72hBreach`/`Exc72hNeedsReview` (`:37-38`) are never emitted.
   Remediation: roster-anchored pairing (amendment C) + 24h defensive
   ceiling + persisted span-breach exceptions; 72h stays compliance-only
   (amendment B).
3. **Freeze re-parenting bypass (SYN-P1-1).** Non-INSERT branches in
   `00006_trigger_return_fix.sql` (`:23-103`) resolve the OLD parent only;
   `forbid_unless_run_draft` (`:122-127`) proves the NEW-parent pattern.
   Remediation: NEW-parent checks + employer-scoped lookups (also closes
   SYN-P3-3) in a new migration redefining the functions.
4. **Duplicate/overlapping intervals unguarded (SYN-P1-2/P1-21).**
   Dedup key is only `(employer, source, source_event_id)`
   (`attendance.go:102-114`); no overlap guard in schema or
   Generate/Correct. Remediation: partial GiST exclusion on live
   intervals + service checks + advisory locks.
5. **Conflicting redelivery silently dropped (SYN-P1-3).** `ON CONFLICT DO
   NOTHING` with no payload comparison. Remediation: fetch-compare-conflict.
6. **Ambiguous breaks mis-cited + flags cleared (SYN-P1-4/P1-6).** All
   break IDs stamped into work refs (`storeSpan :372-375`);
   `upsertRecord` (`:542-554`) unconditionally resets `reviewed`, including
   from `Correct` (`review.go:242-249`). Remediation: cite only consumed
   events; re-validate before clearing flags.
7. **Session-timezone date casts (SYN-P1-5).** `occurred_at::date` in
   `timesheet.go:198` and `review.go:88`; no `SET TIME ZONE` anywhere.
   Remediation: `(occurred_at AT TIME ZONE timezone)::date` in both queries.
8. **Role separation docs-only (SYN-P1-7/P1-24).** `MIGRATION_DATABASE_URL`
   read nowhere (`cmd/workforceops/migrate.go:57-63` uses `DATABASE_URL`
   only); `00004:15-30` revokes skip silently without pre-created roles;
   every service suite uses the superuser DSN. Remediation: migrate
   preflight + env honoring + app-role suites.
9. **Status lifecycle broken (SYN-P1-8/P1-18).** `SetStatus`
   (`hr.go:148-153`) inserts open-ended rows without closing priors
   (second transition always 23P01); any→any transitions incl.
   resurrection. Remediation: close-prior in tx + transition matrix.
10. **Approved-history instability (SYN-P1-9).** Generate (`timesheet.go:82`)
    and Assess (`readiness.go:87`) filter `status='active'`; terms read at
    assessment time, never snapshotted. Remediation: effective-dated
    population + union with approved-sheet holders + retroactive-change
    blocking (amendment E).
11. **Publish without validation or revise path (SYN-P1-10/P1-11).**
    `Publish` (`roster.go:376-396`) is a bare conditional UPDATE;
    `roster_revision.published_*` never written; no service creates v2.
    Remediation: in-tx revalidation gate + `Revise` service on CLI+REST.
12. **Correction loop needs SQL (SYN-P1-12).** `Review` (`review.go:52-56`)
    exposes no interval IDs; no intervals endpoint exists. Remediation:
    GET intervals endpoint + CLI (amendment G).
13. **Parallel setup race (SYN-P1-13).** `dbutil.CreateDatabase` tolerates
    only 42P04/55006; concurrent extension/type creation raises 23505.
    Remediation: tolerate-and-verify 23505 + serialize MigrateUp.
14. **Rejected sheets dead-end (SYN-P1-14).** Reject is terminal; trigger
    freezes rejected; regenerate skips "already exists". Remediation flows
    from amendment-A correction versions (rejected originals stay
    immutable; successor draft links via `supersedes_id`).
15. **`timesheet` UNIQUE blocks append-only corrections (new).**
    `UNIQUE (employer_id, employee_id, pay_period_id)` (`00002`) forbids a
    second (correction) sheet for the same employee+period. Fixable in
    frozen arch: new migration replaces it with a partial unique on
    original sheets + `supersedes_id` + one-successor enforcement.
16. **No explicit current-interval predicate (new).** "Current" is inferred
    via `NOT EXISTS successor`. Per amendment G this must become explicit
    schema: `is_current` maintained by a SECURITY DEFINER trigger owned by
    `migration_admin` (the `app` role cannot UPDATE the table, so a plain
    trigger would fail), with the partial exclusion keyed on it.
17. **Readiness population/exclusion semantics (SYN-P2-8/P2-9/P2-30).**
    `runID` unvalidated (`readiness.go:182-207`); approved-but-empty sheets
    count as complete; `AuthorizeApproval` approves empty Included
    vacuously (`approval.go:75-137`, returns nil on empty). All fixable as
    specified in amendment E.

## Confirmed implementation defects (non-structural)

SYN-P1-15/16/17 (Assign/Template/end-date validation), SYN-P1-19 (float64
money — verified `hr.go:169,182,255`), SYN-P1-20 (Import/Correct races),
SYN-P1-22 (rotation `pattern` shadowing), SYN-P1-23 (SQL-only onboarding:
verified — no employer/period/SHG create in CLI or REST), SYN-P2-1/2/3/4/5/6/7
(envelope, idempotency, atomicity gaps — verified `fail()` emits
`{error:CODE,message}` with no request_id; `audit_event`/`idempotency_key`
have zero writers), SYN-P2-13/14/16/17/18/19/20/21/22/23/24/25/26/27/28,
SYN-P2-31/32/33, SYN-P3-1…13 batch. Dockerfile/fly stale refs re-verified
(`Dockerfile:12-14,24`, `fly.toml:16-18,33`). None requires architectural
change; each maps to a remediation step or an explicit deferral.

## Risks requiring reproduction (Stage 2)

- R1 overlapping pay periods → double allocation (§3A).
- R2 cross-midnight attribution (§3B). Domain contract
  (`domain-model.md:103-105`) requires split at local midnight;
  `storeSpan` stamps one `local_date` from the in-event
  (`attendance.go:346`) — code and contract disagree, behavior untested.
- R3 late break-event attachment (§3C).
- R4 concurrent corrections → dual successors (§3D).
- R5 fresh/upgrade/rollback migration paths (§3E).
- R6 concurrent import/generate/publish races (RISK-1).
- R7 `seed demo --reset` wedge (RISK-3; never executed live).
- R8 `docker build` failure mode (RISK-4; deterministic by inspection).

## Untested hypotheses (not M1.1-gating unless Stage 2 promotes them)

Review §4g items 3, 5(partially — covered by blocker 4 schema work), 7, 8,
9, 10, 11, 14, 15, 16, 17, 18, 19, 20. Items 1, 2, 4, 5, 6, 12, 13 are
absorbed into R1–R5 above.

## Clean areas (verified, preserved)

All 17 review §5 areas stand; additionally verified in this run:

- **Module registry dead (RISK-5 resolved).** `go list -deps
  ./cmd/workforceops` references neither `modules/sg-wfm-payroll` nor
  `core/module`: the ShiftHub/PayLoop descriptors are unreachable dormant
  code. No fix needed; no module framework will be built.
- **Migration mechanics** (`00005`/`00007`/`00008`, idempotent up, ≥8
  recorded) and **no-mock test honesty** hold.
- No evidence of a deployed user database (caveat: compose volume
  uninspectable here); nothing will be dropped regardless.

## Frozen-decision check

No finding requires abandoning a §7 frozen decision. The closest calls —
append-only correction versions (new statuses/columns, same single DB and
service layer) and the SECURITY DEFINER maintenance trigger (same two-role
model) — fit the frozen architecture. No item stops for approval.
