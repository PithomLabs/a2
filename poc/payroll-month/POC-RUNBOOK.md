# POC runbook: from clone to green

Target stack: PostgreSQL 18.6 + DBOS (Go library) + immudb 1.11, one-month
synthetic payroll. Disposable databases only; never point this at shared,
staging, or production data.

## 0. Prerequisites

- Docker daemon with compose v2, ACCESSIBLE to your user (`docker ps` must
  work — typically membership in the `docker` group; a present-but-unusable
  socket fails Gate B bring-up the same as no daemon).
- Go 1.27+, network access to proxy.golang.org and container registries.
- Checkouts (or registry versions) of dbos-transact-golang and immudb matching
  `poc/payroll-month/go.mod` replace directives.
- Registry digests are pinned in `docker-compose.poc.yml` (VERIFIED 2026-11-10
  via Hub API, linux/amd64). Re-verify with `docker buildx imagetools inspect`
  if the tags ever move.
- Human rulings R1–R7 (`docs/stack/DECISION-REQUESTS.md`) before any claim
  beyond "POC executed"; provisional POC-scoped choices are listed in
  `POC-REPORT.md` and bind nothing in production.

## 1. Environment (sane dev-only defaults)

```bash
set -a; source poc/payroll-month/.env.poc.local; set +a
```

This loads disposable-local credentials, URLs, and ports. Every secret in
that file is insecure by design, labeled dev-only, and for localhost POC use
only — never reuse the values anywhere else. The file is git-ignored
(`.env.*.local`); `.env.poc.example` next to it is the committable template.
Override any variable by exporting it after sourcing.

Role note: `init-roles.sh` creates `migration_admin`/`app` inside the fresh
cluster. `poc-run migrate` uses the privileged URL; the month run uses the
least-privileged URL that still passes preflight.

## 2. Bring-up (one command per step)

```bash
# 1. Start the stack (PostgreSQL 18.6 + immudb), wait for healthchecks.
docker compose -f poc/payroll-month/docker-compose.poc.yml up -d
docker compose -f poc/payroll-month/docker-compose.poc.yml ps  # both healthy

# 2. Apply a2 migrations 00001-00012 to the fresh 18.6 cluster (E5 setup).
MIGRATION_DATABASE_URL="$POC_DBOS_MIGRATE_URL" go run ./cmd/workforceops migrate up

# 3. Create DBOS system tables (privileged step, once per cluster).
(cd poc/payroll-month && POC_DBOS_MIGRATE_URL="$POC_DBOS_MIGRATE_URL" go run ./cmd/poc-run migrate)

# 3.5. Apply POC DDL (outbox + gate; disposable DB only, NEVER a migration).
# Tables/triggers/table-grants run as migration_admin; the T13 advisory-ACL
# section MUST run as the superuser (role poc): pg_catalog ACLs cannot be
# altered by migration_admin (its REVOKE/GRANT warn-and-no-op instead).
set -a; . poc/payroll-month/.env.poc.local; set +a
export PGPASSWORD="$POC_ROLE_PASSWORD"
docker exec -i -e PGPASSWORD="$PGPASSWORD" a2-poc-postgres psql -U migration_admin -d poc -v ON_ERROR_STOP=1 -f - < poc/payroll-month/ledger/outbox.sql
sed -n '1,/T13 hardening/p' poc/payroll-month/gate/gate2_poc.sql | docker exec -i -e PGPASSWORD="$PGPASSWORD" a2-poc-postgres psql -U migration_admin -d poc -v ON_ERROR_STOP=1 -f -
export PGPASSWORD="$POC_POSTGRES_PASSWORD"
sed -n '/T13 hardening/,$p' poc/payroll-month/gate/gate2_poc.sql | docker exec -i -e PGPASSWORD="$PGPASSWORD" a2-poc-postgres psql -U "$POC_POSTGRES_USER" -d poc -v ON_ERROR_STOP=1 -f -
# Expect: clean REVOKE/GRANT lines, then the $t13$ DO block commits silently.
# ANY WARNING on the T13 range = wrong role: stop, do not proceed to step 4.
unset PGPASSWORD

# 3.6. Authority overlay (prompt8 remediation; disposable DB only).
# Run AFTER seed-poc (the pack schedule seeds for POC-MERIDIAN), as
# migration_admin. Installs the DEFINER authority boundary (capture /
# persist / approve / validate / close / enqueue / backlink), the
# writer-gate triggers, and the outbox transition guard. The file ends
# with a loud $poc_auth$ assertion: ANY error = stop, do not proceed.
# (Fresh clusters also need, once: GRANT CREATE ON DATABASE poc TO
# migration_admin — as the superuser, before `poc-run migrate`.)
set -a; . poc/payroll-month/.env.poc.local; set +a
export PGPASSWORD="$POC_ROLE_PASSWORD"
docker exec -i -e PGPASSWORD="$PGPASSWORD" a2-poc-postgres psql -U migration_admin -d poc -v ON_ERROR_STOP=1 -f - < poc/payroll-month/gate/pg_authority.sql
# Expect: only CREATE/GRANT/REVOKE/ALTER/INSERT lines + a few first-apply
# NOTICEs; the $poc_auth$ block commits silently.
unset PGPASSWORD

# 4. Run the month (preflight + synth demo + DBOS workflow + JSON report).
(cd poc/payroll-month && go run ./cmd/poc-run)  # exit 0 == all required checks pass

Subcommands (see `e2e/*_test.go` callers): `migrate` (DBOS system DDL,
privileged URL), `capture-once --run-id X`, `outbox-once [--op UUID]`
(single-row drain, oldest-first unless pinned), `immudb-users`
(provision poc_writer/poc_auditor, once per immudb volume),
`canonical` (Phase 2 six outputs as JSON, DB-free),
`--demo-synth-only` (documented exit 0 without services). Worker state dir:
POC_WRITER_STATE_DIR (default `.poc-writer-state`, never the auditor's).

# 5. Run the experiment harnesses (E1/E2/coverage/outbox execute; chaos specs skip).
(cd poc/payroll-month && go test ./...)

# 6. Run the a2 battery against 18.6 (E5 evidence; needs TEST url on :5544).
WORKFORCEOPS_TEST_DATABASE_URL="..." go test -count=1 ./...
```

## 3. Auditor bring-up (once per cluster, before the month)

```bash
# 3.7. Identities (prompt8 §5.3): provision the ledger users once per
# immudb volume (admin bootstrap credential), then verify. The PG
# auditor role is created by the 20-auditor.sh init script on fresh
# clusters (existing volumes: run that SQL as the superuser once).
(cd poc/payroll-month && go run ./cmd/poc-run immudb-users)  # expect overall pass

# 3.8. Genesis ceremony: accept genesis on an empty auditor volume and
# commit the printed anchor (server UUID + tx-1 content hash) to
# gate/immudb-genesis.json out-of-band (git, signed) per ADR-0204.
# Ordinary audits pin with --expect-uuid/--expect-tx/--expect-tx-hash
# from that file; any mismatch exits 2 before any trust decision.
(cd poc/payroll-month && go run ./cmd/auditor --accept-genesis --state-dir /var/lib/auditor \
  --employer POC-MERIDIAN --run <validated-run-uuid>)
```

Content mode audits a finalized run end to end (PG sources →
recomputed evidence vs verified ledger bytes); `--txs` is a
liveness-only probe (labeled as such, never content evidence).
The auditor connects as poc_auditor (read-only) + the PG auditor
role (SELECT-only) and refuses the writer's state dir.

## 4. Chaos appendix (kill -9 / tamper; manual)

1. Start the month run; `kill -9` the `poc-run` worker (a) inside the capture
   txn pre-commit, (b) between PG commit and ledger append, (c) between append
   and back-link. Restart; recovery must complete exactly once (no partial
   rows, no duplicate/missing ledger records, single outcome row).
2. Tamper (automated: `TestContentAudit_TamperBattery`,
   `TestContentAudit_ForgedAndMissing`): edit amounts, digests, snapshot
   rows, pins, calc version, approval, or payslips in PG as the owner;
   forge backlinks; claim missing events. The content auditor must fail
   loudly naming the violated check; exact restoration must pass again.
   Ledger-side host corruption: see the IM-8 harness note in
   plans2/POC-REMEDIATION-REPORT.md (scripted, evidence-recorded).
3. Concurrency: race a writer against capture in both orderings under the
   Gate-2 gate DDL; assert included-or-blocked, never silently missed.

## 4. Teardown

```bash
docker compose -f poc/payroll-month/docker-compose.poc.yml down -v  # deletes POC data
```

## 5. Troubleshooting

- `POC_DATABASE_URL unset` / exit 1 with `demo-synth-only`: expected without
  the stack — the synth demo still proves determinism, nothing else.
- Digest pins: `docker buildx imagetools inspect <image>:<tag>` prints the
  digest to substitute for each TODO tag.
- If `migrate up` fails on 18.6, that IS the E5 finding: record the error
  verbatim in POC-REPORT.md, do not edit migrations 00001–00012.
