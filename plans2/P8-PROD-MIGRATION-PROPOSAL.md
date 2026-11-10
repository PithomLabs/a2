# Production migration proposal (prompt8 remediation companion)

Status: PROPOSAL ONLY — no migration file created, frozen 00001–00012
untouched, no 00013. This document describes the reviewed-production
equivalent of `poc/payroll-month/gate/pg_authority.sql` (disposable-DB
overlay, already executed there). Each item needs independent review,
a staged rollout plan, and the existing battery green before it can
become a real migration.

## 1. Revoke direct application write on authoritative tables

- `payroll_snapshot`, `payroll_snapshot_employee`,
  `payroll_snapshot_pack`: app keeps SELECT; INSERT/UPDATE/DELETE
  revoked (UPDATE/DELETE already revoked by 00012; INSERT is the
  change — the 00012 verify block asserting INSERT must be updated
  in the same migration).
- `payroll_run`: app keeps SELECT + INSERT (draft creation moves to
  the DEFINER; see §3 — then INSERT is revoked too).
- `payroll_line`, `payslip`: app keeps SELECT; all writes revoked.
- New control tables (`poc_claim` → `payroll_claim`,
  `poc_period_close` → `payroll_period_close`,
  `poc_rule_pack_schedule` → `rule_pack_schedule`,
  `poc_run_output`, `poc_approval`, outbox, backlink): app SELECT
  (and guarded UPDATE on the outbox state machine) only.

## 2. Structural corrections

- `payroll_snapshot_pack` PK `(snapshot_id, pack)` →
  `(snapshot_id, pack, version)`: a mid-period pack change pins two
  versions truthfully. No FK dependents (verified: readers use
  plain SELECT).
- New `rule_pack_schedule` table (effective-dated, overlap-free per
  pack via exclusion constraint): the frozen `rule_pack_activation`
  holds one version per pack and cannot represent the October
  v1a→v2a change. Migration seeds it from reviewed rule data (never
  from POC fixture JSON without re-verification).
- Outbox states extended to
  pending/appended/backlinked/failed/reconciling/reconciled/poison
  with the transition-guard trigger (evidence identity immutable,
  attempts never rewind, anchored tx immutable).
- Backlink PK `(pg_table, pg_id)` →
  `(pg_table, pg_id, event_uuid)`: one row per evidence event so
  corrections append instead of conflicting.

## 3. Authorized operations (all SECURITY DEFINER, owner
## migration_admin, search_path pg_catalog, explicit EXECUTE)

- `snapshot_capture(employer, period_code, actor)`: derives
  population/terms/sheets/intervals/leave/comp/pins/decisions/hashes
  from authoritative tables; writes header, members, pins, claims,
  run derived fields; calls the verifier in-transaction. The sole
  writer of snapshot/claim/run rows.
- `verify_claims(run, snapshot)`: both-directions check over real
  source relationships (stale/forged claims, omissions, orphan
  tables, member↔source relations, re-derived decisions, joint pack
  coverage); vacuous evidence rejected. EXECUTE withheld from app.
- `result_persist(run, calc_version, input_digest, lines, payslips,
  actor)`: binds results to the tip hash + pinned calc version;
  included-members only; exact-cents arithmetic; recomputes the
  output digest.
- `approve(run, approver, decision, note)`: requires persisted
  results and zero missing-timesheet members.
- `run_validate(run, actor)`: live re-verification + recomputed
  header binding + approval + output checks, then draft→validated
  (the frozen `run_lifecycle` trigger stays as a backstop).
- `close_period(employer, period, run)`: validated-run required,
  DB-clock timestamp, period row locked, duplicate close rejected.
- `outbox_enqueue(employer, run, event_uuid, actor)`: validated-run
  required; payload derived from persisted artifacts; same-op
  replay converges, differing payload rejected loudly.
- `backlink_record(table, id, event, tx, alh)`: requires a matching
  appended/reconciling outbox row; exact replays converge,
  conflicts rejected loudly.

## 4. Writer participation

- `require_gate()` trigger function becomes DEFINER; direct app
  EXECUTE on the gate function revoked (writers reach the gate via
  triggers only).
- Triggers cover INSERT/UPDATE/DELETE on all capture-input tables
  plus `payroll_run`, `pay_period`, `payroll_run_exclusion`.
- Cross-employer UPDATE takes both employer keys in sorted order.
- Closed-period rejection derives the period via real
  `pay_period_id` relationships (timesheet/line/interval, run,
  exclusion, period rows). Tables without a period relationship
  (clock/attendance/terms/leave/comp) stay gate-only until a
  roster/pay-calendar mapping is designed — an explicit, recorded
  limitation, not a silent gap.

## 5. Rollout order + verification

1. Add new tables + DEFINER functions (no behavior change yet).
2. Migrate writers call-site by call-site (capture → persist →
   approve → validate → close → enqueue); each slice keeps the
   battery green with both paths live.
3. Flip revokes + widen keys + install triggers; the `$poc_auth$`
   assertion (ported) fails the migration loudly on any drift.
4. Backfill: no backfill needed (new boundary for new runs);
   historical runs keep their recorded posture, labeled accordingly.
5. Harden the frozen triggers the boundary writes through
   (`payroll_line_freeze` and any sibling without `SET search_path`):
   give each `SET search_path = pg_catalog` plus schema-qualified
   references. Proven live: an unqualified frozen trigger fails
   42P01 when fired inside a pg_catalog-only DEFINER chain; the POC
   overlay works around it with a transaction-local widening, but
   production must fix the trigger (the workaround's safety rests
   on app holding no CREATE on public and no TEMP — true today,
   but the trigger fix removes the dependency).
6. Verify: the prompt8 §3.2 exploit chain (all rejected) + the
   authorized-path composition tests (all green) + the full
   existing battery, on a production-like staging host.

## 6. Explicitly out of this proposal

- Statutory rule content (OT-001 stays BLOCKED) and proration
  semantics (still unresolved — the boundary blocks, never guesses).
- The U4 service-layer Tx threading (application refactor, separate
  review).
- immudb licensing/operations (counsel + R1/R2 decisions).
- Any data migration of existing POC disposable rows (there is no
  production data to migrate; POC rows stay disposable).
