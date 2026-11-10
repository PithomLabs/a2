#!/bin/sh
# POC auditor role bootstrap (first container init only). POC-scoped companion
# to scripts/db/init-roles.sh: the independent auditor (§5.2/§5.3) reads
# PostgreSQL through this SELECT-only role — never app, never an owner.
# Disposable POC databases only.
#
# Required environment: POSTGRES_DB, POSTGRES_USER, POC_AUDIT_PG_PASSWORD.
# Re-running is safe: an existing role keeps its password (reported, not reset).
set -eu

: "${POSTGRES_DB:?POSTGRES_DB is required}"
: "${POSTGRES_USER:?POSTGRES_USER is required}"
: "${POC_AUDIT_PG_PASSWORD:?POC_AUDIT_PG_PASSWORD is required}"

echo "WARNING: creating the POC auditor role with a development password (localhost only)."

psql -v ON_ERROR_STOP=1 -v auditpw="$POC_AUDIT_PG_PASSWORD" -v db="$POSTGRES_DB" \
  --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<'EOSQL'
SELECT NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'auditor') AS need_auditor \gset
\if :need_auditor
CREATE ROLE auditor LOGIN PASSWORD :'auditpw';
\else
\echo 'role auditor exists, keeping its password'
\endif

GRANT CONNECT ON DATABASE :"db" TO auditor;
REVOKE TEMPORARY ON DATABASE :"db" FROM auditor;
GRANT USAGE ON SCHEMA public TO auditor;
-- Existing tables (init order: roles first, so usually none yet) plus all
-- future migration_admin tables: SELECT only, never writes, never DDL.
GRANT SELECT ON ALL TABLES IN SCHEMA public TO auditor;
ALTER DEFAULT PRIVILEGES FOR ROLE migration_admin IN SCHEMA public
  GRANT SELECT ON TABLES TO auditor;

DO $verify$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'auditor' AND rolcanlogin) THEN
    RAISE EXCEPTION 'provisioning incomplete: role auditor missing or NOLOGIN';
  END IF;
  IF NOT has_database_privilege('auditor', current_database(), 'CONNECT') THEN
    RAISE EXCEPTION 'provisioning incomplete: auditor lacks CONNECT on %', current_database();
  END IF;
  IF has_database_privilege('auditor', current_database(), 'TEMP') THEN
    RAISE EXCEPTION 'provisioning incomplete: auditor keeps TEMPORARY on %', current_database();
  END IF;
  IF NOT has_schema_privilege('auditor', 'public', 'USAGE') THEN
    RAISE EXCEPTION 'provisioning incomplete: auditor lacks USAGE on schema public';
  END IF;
END
$verify$;
EOSQL
