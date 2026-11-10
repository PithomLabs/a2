# Gate B bring-up notes (2026-11-10, live session)

Honest taxonomy: VERIFIED = observed on the live disposable stack this
session; WRITTEN = file changed, apply pending; DEFERRED = needs docker
(user asked for no more permission prompts, so no docker-gated step was
run after 2026-11-10 bring-up audit).

## VERIFIED on the live stack (a2-poc-postgres / a2-poc-immudb, both healthy)

- PG 18.6 disposable DB `poc`: 12/12 migrations applied; `dbos.*` schema
  present (13 tables, migrate passes).
- POC tables present: poc_outbox, poc_ledger_backlink, poc_claim,
  poc_period_close. Triggers: poc_backlink_no_rewrite (1),
  poc_gate_participate (12/12 tables). poc_require_gate proacl =
  `{migration_admin=X/migration_admin}` (app cannot call directly).
- T13 core posture ENFORCED as superuser `poc` (only superuser/owner can
  alter pg_catalog ACLs; migration_admin REVOKE/GRANT warn-and-no-op):
  app xact(b)=t, session(b)=f, admin session(b)=t, xact(ii)=t, try(ii)=t.
- Advisory usage census (Go + SQL): app path needs xact(bigint) only
  (acquire_* helpers, employer bootstrap, all via hashtext->bigint) and
  xact(ii) via poc_gate_acquire trigger path. Provision uses session
  lock/unlock(bigint) as admin. `_shared` family unused anywhere.
- Default privileges: `migration_admin|public|r|{app=arwd}` — every new
  table grants app CRUD unless the DDL strips it.
- Live grant drift found: poc_claim / poc_period_close / poc_outbox held
  app=CRUD while file intent was SI (claim/close) / unstated (outbox).
  poc_ledger_backlink correct (INSERT,SELECT).
- Ports open: PG :5544, immudb :3322. Auditor correctly not running
  (compose profile `audit`, on-demand only).

## WRITTEN this session (apply pending, all in poc/payroll-month)

- `gate/gate2_poc.sql` T13 section rewritten: all 21 advisory functions
  (8 xact -> app+admin; 13 session incl. unlock_all -> admin only),
  self-healing REVOKEs FROM app, loud `$t13$` assertion DO block
  (app-per-function check + PUBLIC/=X/NULL check). Fixed the old
  session(ii)-to-app grant. Statement counts verified: 8/13/13, and all
  21 spellings match the live `pg_proc` inventory.
- `gate/gate2_poc.sql` poc_claim / poc_period_close: added
  `REVOKE ALL ... FROM app` before the SI grants (self-healing vs
  default arwd).
- `ledger/outbox.sql`: added poc_outbox grants (SIU to app; worker runs
  as app, needs UPDATE for the state machine; no DELETE path in
  worker/outbox.go — comment-verified, not assumed).
- `POC-RUNBOOK.md`: inserted missing step 3.5 (POC DDL apply split by
  role: migration_admin for tables/triggers/grants, superuser `poc` for
  the T13 range; any WARNING on the T13 range = stop).
- `dbexec/ESTIMATE.md`: appended evidenced U4 design (exact census:
  qtx takers, in/out-of-path pgx.Tx params, zero tag reads, ~20
  ErrNoRows `==` sites as a live-PG gate, audit.WriteExec additive,
  layering core/database + POC-only DbosExec).

## DEFERRED (exact commands; run from /home/chaschel/Documents/go/a2)

Runbook step 3.5 (idempotent re-apply of both files — all statements are
`IF NOT EXISTS` / `OR REPLACE` / re-runnable REVOKE+GRANT), then:

```bash
set -a; . poc/payroll-month/.env.poc.local; set +a
export PGPASSWORD="$POC_ROLE_PASSWORD"
# Verify grants (want: claim/close/backlink=SI, outbox=SIU, no DELETE):
echo "SELECT table_name, grantee, string_agg(privilege_type,',' ORDER BY 1) \
FROM information_schema.role_table_grants WHERE table_name LIKE 'poc%' \
GROUP BY 1,2 ORDER BY 1,2;" | docker exec -i -e PGPASSWORD="$PGPASSWORD" \
a2-poc-postgres psql -U migration_admin -d poc -t -A
# Verify T13 (want: 21 explicit proacls, none NULL, none with a PUBLIC entry):
echo "SELECT count(*) FILTER (WHERE proacl IS NULL OR proacl::text ~ '(^|,)=X') \
FROM pg_proc WHERE proname LIKE '%advisory%';" | docker exec -i \
-e PGPASSWORD="$PGPASSWORD" a2-poc-postgres psql -U migration_admin -d poc -t -A
# Want: 0
# Then: E1/E2/coverage/outbox harnesses, auditor build
# (docker compose -f poc/payroll-month/docker-compose.poc.yml \
#   --profile audit build auditor), Exp1-3.
unset PGPASSWORD
```

## Todo state

- DONE: B1 bring-up audit (stack healthy, DDL verified, T13 core
  enforced, drift found + fixed in files, U4 evidenced, runbook step
  3.5 written).
- NEXT (docker-gated, in order): apply §DEFERRED 1+2, verify, shared
  fixture + canonical month, Exp1/Exp2/Exp3, comparison + final reports.
