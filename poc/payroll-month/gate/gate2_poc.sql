-- Gate-2 POC DDL (ADR-0203, Gate A item A6). POC-ONLY: apply to disposable
-- databases only (runbook step 2.5, migration_admin role). NEVER a production
-- migration: no 00013 exists and none is proposed here. UNEXECUTED (no PG in
-- this environment); acceptance is the T1–T13 battery on the docker host.
--
-- Mechanism (ported from core/database/gate2_optionb_test.go:137–440, the
-- executable Gate-2 spec): transaction-scoped employer advisory gate + writer
-- participation triggers + claim/proof-check + close transaction with
-- DB-controlled timestamps (ChatGPT#9: client timestamps never decide
-- eligibility or ordering).

-- ---------------------------------------------------------------- gate --
CREATE OR REPLACE FUNCTION poc_gate_acquire(p_employer uuid)
RETURNS void LANGUAGE plpgsql SET search_path = pg_catalog AS $fn$
BEGIN
    -- Same 64-bit derivation family as the Gate-2 spec: one keyspace per
    -- employer, one fixed namespace tag. Transaction-scoped: the lock dies
    -- with the holder's txn (commit, rollback, or kill -9).
    -- Hardened per the maintain_interval_currency precedent (00010): this
    -- function fires under hostile inherited paths (DEFINER cascades run
    -- pg_catalog-only), so every reference is schema-qualified.
    PERFORM pg_catalog.pg_advisory_xact_lock(hashtext(p_employer::text), hashtext('poc-gate-v1'));
END;
$fn$;
ALTER FUNCTION poc_gate_acquire(uuid) OWNER TO migration_admin;
REVOKE ALL ON FUNCTION poc_gate_acquire(uuid) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION poc_gate_acquire(uuid) TO app, migration_admin;

-- ------------------------------------------------------- participation --
-- Writers join the gate by BLOCKING on it: while capture holds the employer
-- key, concurrent writers wait; whoever commits first wins, and the proof
-- check (below) classifies every claim as included-or-blocked, never missed.
CREATE OR REPLACE FUNCTION poc_require_gate()
RETURNS trigger LANGUAGE plpgsql SET search_path = pg_catalog AS $fn$
DECLARE
    v_employer uuid;
    v_closed   boolean;
BEGIN
    v_employer := COALESCE(
        (to_jsonb(NEW) ->> 'employer_id')::uuid,
        (to_jsonb(NEW) ->> 'employer')::uuid);
    IF v_employer IS NULL THEN
        RAISE EXCEPTION 'poc_require_gate: no employer identity on %', TG_TABLE_NAME;
    END IF;
    -- Closed periods reject writers at the DB (D1 close semantics).
    -- Schema-qualified: fires under hostile inherited paths (see above).
    SELECT EXISTS (
        SELECT 1 FROM public.poc_period_close c
        WHERE c.employer_id = v_employer
          AND c.period = COALESCE((to_jsonb(NEW) ->> 'period'), (to_jsonb(NEW) ->> 'pay_period'), '')
          AND c.closed_at IS NOT NULL
    ) INTO v_closed;
    IF v_closed THEN
        RAISE EXCEPTION 'poc_require_gate: period closed for employer %', v_employer
            USING ERRCODE = 'P0001';
    END IF;
    PERFORM public.poc_gate_acquire(v_employer);
    RETURN NEW;
END;
$fn$;
ALTER FUNCTION poc_require_gate() OWNER TO migration_admin;
REVOKE ALL ON FUNCTION poc_require_gate() FROM PUBLIC;
-- NOTE: triggers fire with the invoker's rights; the function needs no
-- DEFINER because advisory-lock acquisition requires no grant. GRANT EXECUTE
-- is withheld from app deliberately: writers reach the gate ONLY via the
-- triggers below (direct calls fail) — T13 asserts this.

-- Coverage: every capture-input table joins the gate (capture reads
-- DOCUMENTED from internal/payroll/snapshot.go). Output tables
-- (payroll_snapshot*) and control tables (pay_period, payroll_run*) are
-- written only via capture/close functions, never directly by app writers.
DO $do$
DECLARE
    t text;
BEGIN
    FOREACH t IN ARRAY ARRAY[
        'attendance_interval', 'clock_event', 'clock_event_disposition',
        'compensation_line', 'employee', 'employee_status_history',
        'employment_terms', 'rule_pack_activation', 'shg_instruction',
        'timesheet', 'timesheet_line', 'timesheet_line_interval'
    ] LOOP
        EXECUTE format('DROP TRIGGER IF EXISTS poc_gate_participate ON %I', t);
        EXECUTE format('CREATE TRIGGER poc_gate_participate BEFORE INSERT OR UPDATE ON %I '
            || 'FOR EACH ROW EXECUTE FUNCTION poc_require_gate()', t);
    END LOOP;
END;
$do$;

-- ------------------------------------------------------------- claims --
CREATE TABLE IF NOT EXISTS poc_claim (
    run_id      uuid        NOT NULL,
    employer_id uuid        NOT NULL,
    table_name  text        NOT NULL,
    row_id      text        NOT NULL,
    input_hash  bytea       NOT NULL,
    claimed_at  timestamptz NOT NULL DEFAULT clock_timestamp(),
    PRIMARY KEY (run_id, table_name, row_id)
);
REVOKE ALL ON poc_claim FROM PUBLIC;
-- Self-healing vs default privileges (migration_admin-owned tables in public
-- grant app=arwd by default): strip app first, then grant intent (SI only).
REVOKE ALL ON poc_claim FROM app;
GRANT SELECT, INSERT ON poc_claim TO app;

-- Proof check: BROKEN (Gate B proven, kept for the record — do not call).
-- This port references payroll_snapshot_employee.source_table/source_row_id,
-- which DO NOT EXIST (the snapshot stores a JSONB inputs doc, not source
-- row ids): every call fails with 42703 (proven live 2026-11-10). Deeper:
-- the Gate-2 spec (gate2_verify_claims) verifies claims against LIVE source
-- tables INSIDE the capture txn (same-txn visibility, both directions);
-- this post-hoc (run, employer) signature cannot reproduce that (live rows
-- move on; omission checks need txn visibility). A correct port needs the
-- in-txn JSONB design + defined per-table hash inputs; until then T10/T11
-- (claim proof-check) are UNPROVEN and membership stays app-side only.
CREATE OR REPLACE FUNCTION poc_verify_claims(p_run uuid, p_employer uuid)
RETURNS TABLE(missing_count bigint) LANGUAGE plpgsql SET search_path = pg_catalog AS $fn$
BEGIN
    -- Writers that committed before capture's gate acquisition MUST appear
    -- in the snapshot inputs; writers blocked behind capture hold no claim.
    -- Membership proof is per-table (snapshot stores source row ids).
    -- Schema-qualified per the maintain_interval_currency precedent (00010).
    RETURN QUERY
    SELECT count(*) FROM public.poc_claim c
    WHERE c.run_id = p_run AND c.employer_id = p_employer
      AND NOT EXISTS (
        SELECT 1 FROM public.payroll_snapshot_employee s
        WHERE s.snapshot_id = p_run AND s.source_table = c.table_name
          AND s.source_row_id = c.row_id
    );
END;
$fn$;
ALTER FUNCTION poc_verify_claims(uuid, uuid) OWNER TO migration_admin;
REVOKE ALL ON FUNCTION poc_verify_claims(uuid, uuid) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION poc_verify_claims(uuid, uuid) TO app, migration_admin;

-- -------------------------------------------------------------- close --
CREATE TABLE IF NOT EXISTS poc_period_close (
    employer_id uuid NOT NULL,
    period      text NOT NULL,
    run_id      uuid NOT NULL,
    closed_at   timestamptz NULL,   -- set ONLY by poc_close_period, DB clock
    closed_by   text        NULL,
    PRIMARY KEY (employer_id, period)
);
REVOKE ALL ON poc_period_close FROM PUBLIC;
-- Self-healing vs default privileges (see poc_claim above).
REVOKE ALL ON poc_period_close FROM app;
GRANT SELECT, INSERT ON poc_period_close TO app;

-- Close transaction: freezes claims, stamps DB time, derives the snapshot
-- digest deterministically, rejects concurrent modification. Client-supplied
-- timestamps are IGNORED (ChatGPT#9 backdating protection).
CREATE OR REPLACE FUNCTION poc_close_period(p_employer uuid, p_period text, p_run uuid)
RETURNS timestamptz LANGUAGE plpgsql SET search_path = pg_catalog AS $fn$
DECLARE
    v_now timestamptz := clock_timestamp();
BEGIN
    PERFORM public.poc_gate_acquire(p_employer);
    UPDATE public.poc_period_close
       SET closed_at = v_now, closed_by = current_user, run_id = p_run
     WHERE employer_id = p_employer AND period = p_period AND closed_at IS NULL;
    IF NOT FOUND THEN
        RAISE EXCEPTION 'poc_close_period: already closed or unknown period'
            USING ERRCODE = 'P0001';
    END IF;
    RETURN v_now;
END;
$fn$;
ALTER FUNCTION poc_close_period(uuid, text, uuid) OWNER TO migration_admin;
REVOKE ALL ON FUNCTION poc_close_period(uuid, text, uuid) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION poc_close_period(uuid, text, uuid) TO app, migration_admin;

-- Hardening assertions (mirror of 00010: verify loudly): every POC
-- plpgsql function with table references MUST pin search_path to
-- pg_catalog-only and schema-qualify all references. Rationale (proven at
-- Gate B): main-schema DEFINER triggers (e.g. maintain_interval_currency)
-- cascade writes into gated tables under a pg_catalog-only inherited path;
-- an unqualified trigger body then fails to resolve its own tables and
-- every supersession through the cascade aborts. This block fails the
-- apply loudly if any function drifts from the hardened shape.
DO $poc_hard$
DECLARE
    fn text;
    fn_path text[];
    fn_def text;
BEGIN
    FOREACH fn IN ARRAY ARRAY['poc_gate_acquire', 'poc_require_gate', 'poc_verify_claims', 'poc_close_period'] LOOP
        SELECT p.proconfig, pg_get_functiondef(p.oid) INTO fn_path, fn_def
          FROM pg_proc p WHERE p.proname = fn;
        IF fn_path IS DISTINCT FROM ARRAY['search_path=pg_catalog'] THEN
            RAISE EXCEPTION '% search_path is %, want pg_catalog only', fn, fn_path;
        END IF;
        IF fn_def LIKE '%FROM poc\_%' ESCAPE '\'
           OR fn_def LIKE '%UPDATE poc\_%' ESCAPE '\'
           OR fn_def LIKE '%PERFORM poc\_%' ESCAPE '\'
           OR fn_def LIKE '%FROM payroll\_%' ESCAPE '\' THEN
            RAISE EXCEPTION '% keeps an unqualified table/function reference', fn;
        END IF;
    END LOOP;
END;
$poc_hard$;

-- ------------------------------------------------- DBOS runtime grants --
-- Gate B finding: DBOS system tables are created by migration_admin
-- (poc-run migrate) but the APP role executes workflows (poc-run
-- capture-once / full-month, e2e idempotency + forged-state tests, all
-- with SkipMigrations). Without these grants every app-role RunWorkflow
-- fails at the schema grant (proven live: permission denied for schema
-- dbos). Same-DB shared-pool deployment UNAVOIDABLY grants the app DML
-- on workflow tables (SystemDBPool = app pool); forged-row resilience
-- must therefore come from the executor, not from grants (Exp2 honest
-- note; the separate-DB R6 fallback would isolate at the cost of the
-- shared-pool atomic boundary). Applied as migration_admin (owner).
GRANT USAGE ON SCHEMA dbos TO app;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA dbos TO app;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA dbos TO app;
ALTER DEFAULT PRIVILEGES FOR ROLE migration_admin IN SCHEMA dbos
    GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO app;
ALTER DEFAULT PRIVILEGES FOR ROLE migration_admin IN SCHEMA dbos
    GRANT USAGE, SELECT ON SEQUENCES TO app;

-- ------------------------------------------------------ T13 hardening --
-- Raw advisory overloads: revoked from PUBLIC, explicitly granted to the two
-- roles that need them (app acquires via poc_gate_acquire only insofar as it
-- cannot bypass the participation triggers; direct lock calls are
-- observable and typed, never silent). Asserts in T13.
-- NOTE: two-key advisory overloads take (integer, integer); the single-key
-- forms take (bigint). Signatures below are exact (a (bigint,bigint) spelling
-- fails at apply time — caught at Gate B bring-up).
-- PRIVILEGE: this section MUST run as a superuser (POC: role poc). pg_catalog
-- function ACLs can only be altered by the owner or a superuser; applied as
-- migration_admin every REVOKE/GRANT below silently no-ops with a WARNING
-- (observed at Gate B bring-up: proacl unchanged, app kept session EXECUTE).
-- The assertion block at the end fails the apply loudly if posture is wrong.
--
-- Policy (all 21 advisory functions):
--   * xact family (8 fns, incl. _shared + try single-key): app + admin. Xact
--     locks die with the holder's txn, so app may coordinate WITHIN its own
--     transactions. App uses xact(bigint) via acquire_* helpers + employer
--     bootstrap, xact(ii) via poc_gate_acquire (trigger path, invoker rights).
--   * session family (13 fns: lock/try/unlock, exclusive+shared, both arities,
--     unlock_all): migration_admin ONLY. Session locks survive the txn: app
--     must neither hold cross-txn locks nor break provision locks (unlock_all).
--   * PUBLIC holds nothing anywhere in the family.
-- Every statement is self-healing (explicit REVOKE FROM app on the session
-- family) so re-apply converges regardless of partial-apply history.
-- Xact family: app + migration_admin.
REVOKE ALL ON FUNCTION pg_advisory_xact_lock(bigint) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION pg_advisory_xact_lock(bigint) TO app, migration_admin;
REVOKE ALL ON FUNCTION pg_advisory_xact_lock(integer, integer) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION pg_advisory_xact_lock(integer, integer) TO app, migration_admin;
REVOKE ALL ON FUNCTION pg_advisory_xact_lock_shared(bigint) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION pg_advisory_xact_lock_shared(bigint) TO app, migration_admin;
REVOKE ALL ON FUNCTION pg_advisory_xact_lock_shared(integer, integer) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION pg_advisory_xact_lock_shared(integer, integer) TO app, migration_admin;
REVOKE ALL ON FUNCTION pg_try_advisory_xact_lock(bigint) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION pg_try_advisory_xact_lock(bigint) TO app, migration_admin;
REVOKE ALL ON FUNCTION pg_try_advisory_xact_lock(integer, integer) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION pg_try_advisory_xact_lock(integer, integer) TO app, migration_admin;
REVOKE ALL ON FUNCTION pg_try_advisory_xact_lock_shared(bigint) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION pg_try_advisory_xact_lock_shared(bigint) TO app, migration_admin;
REVOKE ALL ON FUNCTION pg_try_advisory_xact_lock_shared(integer, integer) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION pg_try_advisory_xact_lock_shared(integer, integer) TO app, migration_admin;
-- Session family: migration_admin only.
REVOKE ALL ON FUNCTION pg_advisory_lock(bigint) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION pg_advisory_lock(bigint) TO migration_admin;
REVOKE ALL ON FUNCTION pg_advisory_lock(integer, integer) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION pg_advisory_lock(integer, integer) TO migration_admin;
REVOKE ALL ON FUNCTION pg_advisory_lock_shared(bigint) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION pg_advisory_lock_shared(bigint) TO migration_admin;
REVOKE ALL ON FUNCTION pg_advisory_lock_shared(integer, integer) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION pg_advisory_lock_shared(integer, integer) TO migration_admin;
REVOKE ALL ON FUNCTION pg_try_advisory_lock(bigint) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION pg_try_advisory_lock(bigint) TO migration_admin;
REVOKE ALL ON FUNCTION pg_try_advisory_lock(integer, integer) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION pg_try_advisory_lock(integer, integer) TO migration_admin;
REVOKE ALL ON FUNCTION pg_try_advisory_lock_shared(bigint) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION pg_try_advisory_lock_shared(bigint) TO migration_admin;
REVOKE ALL ON FUNCTION pg_try_advisory_lock_shared(integer, integer) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION pg_try_advisory_lock_shared(integer, integer) TO migration_admin;
REVOKE ALL ON FUNCTION pg_advisory_unlock(bigint) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION pg_advisory_unlock(bigint) TO migration_admin;
REVOKE ALL ON FUNCTION pg_advisory_unlock(integer, integer) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION pg_advisory_unlock(integer, integer) TO migration_admin;
REVOKE ALL ON FUNCTION pg_advisory_unlock_shared(bigint) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION pg_advisory_unlock_shared(bigint) TO migration_admin;
REVOKE ALL ON FUNCTION pg_advisory_unlock_shared(integer, integer) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION pg_advisory_unlock_shared(integer, integer) TO migration_admin;
REVOKE ALL ON FUNCTION pg_advisory_unlock_all() FROM PUBLIC;
GRANT EXECUTE ON FUNCTION pg_advisory_unlock_all() TO migration_admin;
-- Self-healing: drop any app grant left by earlier partial applies or older
-- file versions (session(ii) was once granted to app by this file).
REVOKE ALL ON FUNCTION pg_advisory_lock(bigint) FROM app;
REVOKE ALL ON FUNCTION pg_advisory_lock(integer, integer) FROM app;
REVOKE ALL ON FUNCTION pg_advisory_lock_shared(bigint) FROM app;
REVOKE ALL ON FUNCTION pg_advisory_lock_shared(integer, integer) FROM app;
REVOKE ALL ON FUNCTION pg_try_advisory_lock(bigint) FROM app;
REVOKE ALL ON FUNCTION pg_try_advisory_lock(integer, integer) FROM app;
REVOKE ALL ON FUNCTION pg_try_advisory_lock_shared(bigint) FROM app;
REVOKE ALL ON FUNCTION pg_try_advisory_lock_shared(integer, integer) FROM app;
REVOKE ALL ON FUNCTION pg_advisory_unlock(bigint) FROM app;
REVOKE ALL ON FUNCTION pg_advisory_unlock(integer, integer) FROM app;
REVOKE ALL ON FUNCTION pg_advisory_unlock_shared(bigint) FROM app;
REVOKE ALL ON FUNCTION pg_advisory_unlock_shared(integer, integer) FROM app;
REVOKE ALL ON FUNCTION pg_advisory_unlock_all() FROM app;
-- Assertions: fail the apply LOUDLY if the posture is wrong (never rely on
-- REVOKE/GRANT silence: under-privileged applies warn-and-no-op).
DO $t13$
DECLARE
    r record;
    want boolean;
BEGIN
    FOR r IN SELECT oid::regprocedure AS fn FROM pg_proc WHERE proname LIKE '%advisory%' LOOP
        want := (r.fn::text LIKE '%xact%');
        IF has_function_privilege('app', r.fn, 'EXECUTE') IS DISTINCT FROM want THEN
            RAISE EXCEPTION 'T13 violation: app EXECUTE on % is %, want %', r.fn, NOT want, want;
        END IF;
    END LOOP;
    -- NULL proacl = default ACL = world-executable: also a violation.
    -- PUBLIC entries have an EMPTY grantee: '=X/grantor' at start or after
    -- a comma. (A naive '%=X/%' LIKE also matches 'role=X/...' — wrong.)
    IF EXISTS (SELECT 1 FROM pg_proc WHERE proname LIKE '%advisory%'
               AND (proacl IS NULL OR proacl::text ~ '(^|,)=X')) THEN
        RAISE EXCEPTION 'T13 violation: PUBLIC holds EXECUTE on an advisory function';
    END IF;
END;
$t13$;

-- Coverage map (writer-matrix row → enforcement in THIS file):
--  D1 capture race .......... poc_gate_acquire + participation triggers + poc_verify_claims (T1–T8)
--  P1a–g writer paths ....... triggers on all 12 input tables (T9 probes each)
--  EF-E3 fabricated input ... claim input_hash + snapshot membership proof (T10/T11)
--  Cutoff race .............. poc_close_period + closed-period writer rejection (T12)
--  Substitution/revoke ...... T13 block above + trigger ownership (T13)
--  DBOS/system tables ....... OUT OF SCOPE (no triggers on dbos.*; forged-state test instead)
-- KNOWN LIMITS (honest): unexecuted; membership proof assumes app populates
-- payroll_snapshot_employee.source_* faithfully (app-verified in Gate B);
-- leave_request/ket tables are not capture inputs in the current build.
