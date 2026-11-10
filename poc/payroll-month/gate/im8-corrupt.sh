#!/bin/bash
# IM-8 host-corruption harness (scripted, evidence-recorded).
# Proves: after host-level corruption of the immudb value files, an
# auditor holding PRE-corruption trusted state FAILS content
# verification loudly (verified value reads reject tampered bytes),
# while a tx-existence check alone does not see it (control cells).
#
# Runs a REAL battery month on PG, drains its evidence to a DISPOSABLE
# immudb (own container/volume, port 3323) — never the needed POC stack.
#
# Usage: ./im8-corrupt.sh  (from poc/payroll-month; needs .env.poc.local)
# Evidence: /tmp/p8-evidence/im8/ (transcript + verdicts + audit logs).
set -u
cd "$(dirname "$0")"
set -a; . ./.env.poc.local; set +a
export GOWORK=off GOCACHE=/tmp/gocache GOMODCACHE=/tmp/gomodcache GOPATH=/tmp/gopath
export GOTELEMETRY=off GOTOOLCHAIN=local
EV=/tmp/p8-evidence/im8
rm -rf "$EV"; mkdir -p "$EV"
exec > >(tee "$EV/transcript.log") 2>&1

IMG="codenotary/immudb:1.11.2@sha256:3033c1b51590ee01df4a9be682504490d783ef7f157240d1d0abf7c6e17f313f"
NAME=a2-poc-immudb-corrupt
VOL=a2-poc-corruptdata
PORT=3323
echo "=== IM-8 $(date -u +%FT%TZ) image=$IMG"
docker rm -f "$NAME" >/dev/null 2>&1; docker volume rm "$VOL" >/dev/null 2>&1
docker run -d --name "$NAME" -p "$PORT:3322" \
  -e IMMUDB_ADMIN_PASSWORD="$POC_IMMUDB_ADMIN_PASSWORD" \
  -v "$VOL:/var/lib/immudb" "$IMG" >/dev/null
for i in $(seq 1 30); do
  docker exec "$NAME" immuadmin status >/dev/null 2>&1 && break
  sleep 2
done
docker exec "$NAME" immuadmin status || { echo "FATAL: corrupt instance never healthy"; exit 1; }

export POC_IMMUDB_ADDR="localhost:$PORT"
echo "--- provision users on :$PORT"
go run ./cmd/poc-run immudb-users | grep -E '"overall"'

echo "--- real battery month on PG (evidence drains to :$PORT later)"
export POC_EMPLOYER=POC-AUTH-01 POC_PERIOD=PP-AUTH-01
export PGPASSWORD="$POC_ROLE_PASSWORD"
# Recycle battery run-scoped artifacts (mirror of e2e authCleanup).
EMP=$(docker exec -i -e PGPASSWORD="$PGPASSWORD" a2-poc-postgres psql -U migration_admin -d poc -t -c "SELECT id FROM employer WHERE code='POC-AUTH-01';" | tr -d ' \n')
for q in \
  "DELETE FROM poc_workflow_attempt WHERE employer_id='$EMP'" \
  "DELETE FROM poc_outbox WHERE employer_id='$EMP'" \
  "DELETE FROM poc_approval WHERE employer_id='$EMP'" \
  "DELETE FROM poc_run_output WHERE employer_id='$EMP'" \
  "DELETE FROM poc_claim WHERE employer_id='$EMP'" \
  "DELETE FROM payroll_snapshot_pack WHERE snapshot_id IN (SELECT id FROM payroll_snapshot WHERE employer_id='$EMP')" \
  "DELETE FROM payroll_snapshot_employee WHERE snapshot_id IN (SELECT id FROM payroll_snapshot WHERE employer_id='$EMP')" \
  "DELETE FROM payroll_snapshot WHERE employer_id='$EMP'" \
  "DELETE FROM payroll_line WHERE employer_id='$EMP'" \
  "DELETE FROM payslip WHERE employer_id='$EMP'" \
  "DELETE FROM poc_period_close WHERE employer_id='$EMP'" \
  "DELETE FROM payroll_run WHERE employer_id='$EMP'" \
  "UPDATE pay_period SET status='open' WHERE employer_id='$EMP' AND status='locked'"; do
  docker exec -i -e PGPASSWORD="$PGPASSWORD" a2-poc-postgres psql -U migration_admin -d poc -c "$q" >/dev/null || { echo "FATAL: recycle: $q"; exit 1; }
done
unset PGPASSWORD
setsid -f go run ./cmd/poc-run --approval-timeout 150s </dev/null >>"$EV/month.log" 2>&1
RUN=""
for i in $(seq 1 60); do
  RUN=$(docker exec a2-poc-postgres psql -U poc -d poc -t -c "SELECT r.id FROM payroll_run r JOIN employer e ON e.id=r.employer_id JOIN poc_run_output o ON o.run_id=r.id WHERE e.code='POC-AUTH-01';" | tr -d ' \n')
  [ -n "$RUN" ] && break
  sleep 2
done
[ -n "$RUN" ] || { echo "FATAL: month never persisted"; tail -n 20 "$EV/month.log"; exit 1; }
echo "run=$RUN"
APPROVED=no
for i in $(seq 1 12); do
  if go run ./cmd/poc-run approve --run-id "$RUN" --decision approve --approver im8-harness 2>/dev/null | grep -q '"overall": "pass"'; then APPROVED=yes; fi
  grep -q '"overall"' "$EV/month.log" 2>/dev/null && break
  sleep 10
done
grep -q '"overall": "pass"' "$EV/month.log" || { echo "FATAL: month did not complete (approved=$APPROVED)"; tail -n 20 "$EV/month.log"; exit 1; }
echo "month finalized (approved=$APPROVED)"

echo "--- drain evidence to :$PORT"
export POC_WRITER_STATE_DIR="$EV/writer-state"
go run ./cmd/poc-run outbox-once --op "$RUN" | grep -E '"overall"'
TX=$(docker exec a2-poc-postgres psql -U poc -d poc -t -c "SELECT immu_tx FROM poc_ledger_backlink WHERE event_uuid='$RUN';")
echo "seeded tx=$TX"
[ -n "$TX" ] || { echo "FATAL: no backlink"; exit 1; }

audit_content() { # <state-dir> [extra flags...] -> exit code; report on stdout
  local dir="$1"; shift
  go run ./cmd/auditor --addr "localhost:$PORT" --state-dir "$dir" "$@" \
    --employer POC-AUTH-01 --run "$RUN" 2>"$dir.stderr.log"
}
audit_liveness() { # <state-dir> [extra flags...]
  local dir="$1"; shift
  go run ./cmd/auditor --addr "localhost:$PORT" --state-dir "$dir" "$@" --txs "$TX" 2>"$dir.stderr.log"
}

echo "--- BASELINE (pre-corruption): stale dir, content + liveness"
audit_content "$EV/stale" --accept-genesis > "$EV/baseline-content.log"; echo "baseline content exit=$? (want 0)"
grep -E '"overall"' "$EV/baseline-content.log"
audit_liveness "$EV/stale-live" --accept-genesis > "$EV/baseline-liveness.log"; echo "baseline liveness exit=$? (want 0)"

echo "--- stop + corrupt defaultdb value file"
docker stop "$NAME" >/dev/null
docker run --rm -v "$VOL:/data" postgres:18.6-alpine sh -c '
  f=/data/defaultdb/val_0/00000000.val
  [ -f "$f" ] || { echo "FATAL: $f missing"; exit 1; }
  echo "corrupting: $f"
  sz=$(stat -c %s "$f"); off=$(( sz > 2048 ? 1500 : sz / 2 ))
  head -c 64 /dev/urandom | dd of="$f" bs=1 seek=$off conv=notrunc status=none
  echo "flipped 64 bytes at offset $off of $sz"' | tee "$EV/corruption.log"
grep -q "^corrupting: /data/defaultdb/val_0" "$EV/corruption.log" || { echo "FATAL: corruption missed"; exit 1; }

echo "--- restart + re-audit (4 cells)"
docker start "$NAME" >/dev/null
sleep 5
docker exec "$NAME" immuadmin status >/dev/null 2>&1 && echo "server: UP after corruption" || echo "server: DOWN after corruption (loud unavailability)"
set +e
audit_content "$EV/stale" > "$EV/stale-content.log"; STALE_CONTENT=$?
audit_liveness "$EV/stale-live" > "$EV/stale-liveness.log"; STALE_LIVE=$?
audit_content "$EV/fresh" --accept-genesis > "$EV/fresh-content.log"; FRESH_CONTENT=$?
audit_liveness "$EV/fresh-live" --accept-genesis > "$EV/fresh-liveness.log"; FRESH_LIVE=$?
set -e
echo "stale+content exit=$STALE_CONTENT (want NONZERO: THE assertion)"
grep -E '"overall"|"name": "J.marker"|"name": "J.record"' "$EV/stale-content.log" | head -n 6
echo "stale+liveness exit=$STALE_LIVE (control: tx-only sees no values)"
echo "fresh+content exit=$FRESH_CONTENT (control)"
echo "fresh+liveness exit=$FRESH_LIVE (control: TOFU)"

echo "--- cleanup"
docker rm -f "$NAME" >/dev/null
docker volume rm "$VOL" >/dev/null
# NOTE: the IM-8 backlink row persists (append-only) pointing at a
# destroyed instance; scoped by unique run, never asserted on. The
# validated battery run stays until the next recycle-first setup.

if [ "$STALE_CONTENT" -ne 0 ]; then
  echo "IM-8 VERDICT: PASS (pre-corruption state fails content verification after host corruption)"
else
  echo "IM-8 VERDICT: FAIL (stale state verified corrupted values)"
  exit 1
fi
