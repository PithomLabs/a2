-- POC authority overlay (prompt8 Phase B): database-enforced authority
-- boundary for snapshot capture, run transitions, period close, claims,
-- outbox and backlinks on the DISPOSABLE POC database only.
--
-- Never applied to shared/staging/production. Frozen migrations
-- 00001-00012 are untouched; this file is the executable proof for the
-- production-migration proposal (plans2/P8-PROD-MIGRATION-PROPOSAL.md),
-- which describes the equivalent reviewed change.
--
-- Apply AFTER runbook step 3.5 + seed-poc, as migration_admin:
--   psql -U migration_admin -d poc -v ON_ERROR_STOP=1 -f pg_authority.sql
-- Idempotent: re-runnable (IF NOT EXISTS / OR REPLACE / convergent
-- seed + self-healing revokes). Ends with a loud $poc_auth$ assertion.
--
-- Conventions (per maintain_interval_currency precedent): every DEFINER
-- function uses SET search_path = pg_catalog with schema-qualified
-- references; ownership is migration_admin; EXECUTE is granted
-- explicitly per function, never via PUBLIC.

-- ============================================================ Part 1 ==
-- Effective-dated rule-pack schedule (POC-scoped).
--
-- The frozen schema has no effective-dated pack store:
-- rule_pack_activation holds ONE version per (employer, pack), but the
-- October workload carries TWO versions (v1a Oct 01-15, v2a Oct 16-31).
-- Production needs a real table (see the proposal doc); the POC pins
-- the schedule here, admin-written, app-readable, overlap-free by
-- exclusion constraint. Digests are sha256 over the canonical pack
-- JSON (synth.RulePack field order, no whitespace); Stage C and the
-- auditor recompute them from workload.MonthPacks() and compare.
CREATE TABLE IF NOT EXISTS poc_rule_pack_schedule (
    employer_id    uuid NOT NULL REFERENCES employer (id),
    pack           text NOT NULL,
    version        text NOT NULL,
    effective_from date NOT NULL,
    effective_to   date NOT NULL CHECK (effective_to >= effective_from),
    status         text NOT NULL CHECK (status IN ('IMPLEMENTABLE', 'BLOCKED', 'SUPERSEDED')),
    digest         text NOT NULL,
    params         jsonb NOT NULL,
    PRIMARY KEY (employer_id, pack, version),
    EXCLUDE USING gist (
        employer_id WITH =,
        pack WITH =,
        daterange(effective_from, effective_to, '[]') WITH &&
    )
);
REVOKE ALL ON poc_rule_pack_schedule FROM PUBLIC;
REVOKE ALL ON poc_rule_pack_schedule FROM app;
GRANT SELECT ON poc_rule_pack_schedule TO app;

-- Pinned seed (convergent: re-apply refreshes to pinned values).
-- params key order MUST match synth.RulePack field order.
INSERT INTO poc_rule_pack_schedule
    (employer_id, pack, version, effective_from, effective_to, status, digest, params)
SELECT e.id, v.pack, v.version, v.effective_from::date, v.effective_to::date,
       'IMPLEMENTABLE', v.digest, v.params
FROM employer e CROSS JOIN (VALUES
    ('SYNTH-OT-001', 'SYNTH-2026-10-v1a', '2026-10-01', '2026-10-15',
     'ca584074d3bbd3eb1dd874e5a6b0766e98c7f1adaa9210a664535b1556d115f6',
     '{"version":"SYNTH-2026-10-v1a","effective_from":"2026-10-01","effective_to":"2026-10-15","synthetic":true,"overtime_threshold_min":480,"overtime_numer":3,"overtime_denom":2,"leave_deduction_denom":30}'::jsonb),
    ('SYNTH-OT-001', 'SYNTH-2026-10-v2a', '2026-10-16', '2026-10-31',
     '323fd2694a5f015a2f61cbe8cbbad61c802d3ee35efeb02d8abeba0eb1ec273a',
     '{"version":"SYNTH-2026-10-v2a","effective_from":"2026-10-16","effective_to":"2026-10-31","synthetic":true,"overtime_threshold_min":480,"overtime_numer":2,"overtime_denom":1,"leave_deduction_denom":30}'::jsonb)
) AS v(pack, version, effective_from, effective_to, digest, params)
WHERE e.code = 'POC-MERIDIAN'
ON CONFLICT (employer_id, pack, version) DO UPDATE SET
    effective_from = EXCLUDED.effective_from,
    effective_to   = EXCLUDED.effective_to,
    status         = EXCLUDED.status,
    digest         = EXCLUDED.digest,
    params         = EXCLUDED.params;

-- Frozen-shape fix (POC overlay, disposable DB only): the frozen
-- payroll_snapshot_pack PK (snapshot_id, pack) admits ONE pin per pack
-- per snapshot, but a mid-period pack change truthfully pins TWO
-- versions (v1a + v2a). Widen to (snapshot_id, pack, version).
-- Idempotent: only fires while the 2-column PK still stands. No FK
-- dependents exist (verified: readers use plain SELECT). The
-- production proposal carries the same change as a reviewed migration.
DO $do$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_constraint c
                JOIN pg_class t ON t.oid = c.conrelid
               WHERE c.conname = 'payroll_snapshot_pack_pkey'
                 AND t.relname = 'payroll_snapshot_pack'
                 AND array_length(c.conkey, 1) = 2) THEN
        ALTER TABLE public.payroll_snapshot_pack
            DROP CONSTRAINT payroll_snapshot_pack_pkey,
            ADD PRIMARY KEY (snapshot_id, pack, version);
    END IF;
END;
$do$;

-- One draft regular run per employer+period (concurrent-capture
-- serialization): two racing captures must not mint two drafts.
-- The DEFINER retries on conflict (owner-side); racers converge via
-- the already-captured catch. Partial: validated history untouched.
CREATE UNIQUE INDEX IF NOT EXISTS poc_one_draft_run
    ON public.payroll_run (employer_id, pay_period_id)
    WHERE kind = 'regular' AND status = 'draft';

-- ============================================================ Part 2 ==
-- Revoke direct application write on every authoritative table.
-- All writes below flow through the DEFINER operations in Parts 3-9.
-- SELECT stays: the app reads everything it needs for calculation.
-- (Owner migration_admin keeps full rights for harness scaffolding.)

REVOKE ALL ON public.payroll_snapshot FROM app;
GRANT SELECT ON public.payroll_snapshot TO app;
REVOKE ALL ON public.payroll_snapshot_employee FROM app;
GRANT SELECT ON public.payroll_snapshot_employee TO app;
REVOKE ALL ON public.payroll_snapshot_pack FROM app;
GRANT SELECT ON public.payroll_snapshot_pack TO app;

REVOKE ALL ON public.payroll_run FROM app;
GRANT SELECT ON public.payroll_run TO app;

REVOKE ALL ON public.payroll_line FROM app;
GRANT SELECT ON public.payroll_line TO app;
REVOKE ALL ON public.payslip FROM app;
GRANT SELECT ON public.payslip TO app;

REVOKE ALL ON public.poc_claim FROM app;
GRANT SELECT ON public.poc_claim TO app;
REVOKE ALL ON public.poc_period_close FROM app;
GRANT SELECT ON public.poc_period_close TO app;

-- Outbox: worker (app role) keeps SELECT + guarded UPDATE (transition
-- trigger in Part 8 enforces edges + payload immutability). INSERT and
-- DELETE leave app entirely: enqueue is DEFINER-only (Part 8).
REVOKE INSERT, DELETE ON public.poc_outbox FROM app;

-- Backlink: app keeps SELECT only; poc_backlink_record (Part 8) is the
-- sole writer. (UPDATE/DELETE were already revoked in outbox.sql.)
REVOKE INSERT ON public.poc_ledger_backlink FROM app;

-- Direct gate calls leave app (Part 7 makes the writer trigger
-- DEFINER so the gate path keeps working through triggers only).
REVOKE EXECUTE ON FUNCTION public.poc_gate_acquire(uuid) FROM app;

-- The claims verifier runs inside authorized transactions only.
REVOKE EXECUTE ON FUNCTION public.poc_verify_claims(uuid, uuid) FROM app;

-- ============================================================ Part 3 ==
-- Source-row hash projections (single definition used by BOTH the
-- capture writer and the verifier, so they cannot drift apart).
-- Each projection binds the columns the capture decision depends on.
CREATE OR REPLACE FUNCTION poc_src_hash(p_table text, p_row uuid)
RETURNS bytea LANGUAGE plpgsql
    SECURITY DEFINER SET search_path = pg_catalog AS $fn$
DECLARE
    v_text text;
BEGIN
    CASE p_table
    WHEN 'employee' THEN
        SELECT m.id::text || '|' || m.employee_no
          INTO v_text FROM public.employee m WHERE m.id = p_row;
    WHEN 'employment_terms' THEN
        SELECT t.id::text || '|' || t.effective_from::text || '|' ||
               COALESCE(t.effective_to::text, '') || '|' || t.monthly_basic::text
          INTO v_text FROM public.employment_terms t WHERE t.id = p_row;
    WHEN 'timesheet' THEN
        SELECT s.id::text || '|' || s.status || '|' || COALESCE(s.approved_at::text, '')
          INTO v_text FROM public.timesheet s WHERE s.id = p_row;
    WHEN 'attendance_interval' THEN
        SELECT i.id::text || '|' || i.local_date::text || '|' ||
               EXTRACT(EPOCH FROM (i.interval_end - i.interval_start))::bigint::text ||
               '|' || i.kind
          INTO v_text FROM public.attendance_interval i WHERE i.id = p_row;
    WHEN 'leave_request' THEN
        SELECT l.id::text || '|' || l.from_date::text || '|' || l.to_date::text ||
               '|' || l.status || '|' || l.days::text
          INTO v_text FROM public.leave_request l WHERE l.id = p_row;
    WHEN 'compensation_line' THEN
        SELECT c.id::text || '|' || c.kind || '|' || c.code || '|' ||
               c.amount_monthly::text || '|' || c.effective_from::text ||
               '|' || COALESCE(c.effective_to::text, '')
          INTO v_text FROM public.compensation_line c WHERE c.id = p_row;
    WHEN 'poc_rule_pack_schedule' THEN
        -- Schedule rows are keyed by name, not uuid: p_row carries the
        -- digest's first 16 bytes reinterpreted is WRONG; instead the
        -- capture passes the row's (xmin-less) identity via a lookup.
        -- Simpler honest rule: pack pins are claimed by version text
        -- hashed with the row content (see poc_pack_claim_hash).
        RAISE EXCEPTION 'poc_src_hash: use poc_pack_claim_hash for pack rows';
    ELSE
        RAISE EXCEPTION 'poc_src_hash: unknown table %', p_table;
    END CASE;
    IF v_text IS NULL THEN
        RAISE EXCEPTION 'poc_src_hash: row % missing in %', p_row, p_table;
    END IF;
    RETURN public.digest(v_text, 'sha256');
END;
$fn$;
ALTER FUNCTION poc_src_hash(text, uuid) OWNER TO migration_admin;
REVOKE ALL ON FUNCTION poc_src_hash(text, uuid) FROM PUBLIC;

-- Pack-pin claim hash (schedule rows have a composite natural key).
CREATE OR REPLACE FUNCTION poc_pack_claim_hash(p_employer uuid, p_pack text, p_version text)
RETURNS bytea LANGUAGE plpgsql
    SECURITY DEFINER SET search_path = pg_catalog AS $fn$
DECLARE
    v_text text;
BEGIN
    SELECT s.pack || '|' || s.version || '|' || s.effective_from::text || '|' ||
           s.effective_to::text || '|' || s.status || '|' || s.digest
      INTO v_text FROM public.poc_rule_pack_schedule s
     WHERE s.employer_id = p_employer AND s.pack = p_pack AND s.version = p_version;
    IF v_text IS NULL THEN
        RAISE EXCEPTION 'poc_pack_claim_hash: pin % % missing', p_pack, p_version;
    END IF;
    RETURN public.digest(v_text, 'sha256');
END;
$fn$;
ALTER FUNCTION poc_pack_claim_hash(uuid, text, text) OWNER TO migration_admin;
REVOKE ALL ON FUNCTION poc_pack_claim_hash(uuid, text, text) FROM PUBLIC;

-- ============================================================ Part 4 ==
-- Claims + coverage verifier, redesigned (replaces the BROKEN
-- poc_verify_claims from gate2_poc.sql, which referenced nonexistent
-- columns and failed 42703 on every call).
--
-- Runs INSIDE the capture/validate transaction (caller holds the
-- employer gate). Checks BOTH directions over REAL source-table
-- relationships; raises (P0001) on the first defect class with counts.
-- A zero-row claim set NEVER validates (vacuous-empty rejection).
-- DROP first: the gate2_poc.sql version has an incompatible return
-- type (OR REPLACE cannot change it). Re-apply safe (result is the
-- same function; no stored dependents).
DROP FUNCTION IF EXISTS public.poc_verify_claims(uuid, uuid);
CREATE FUNCTION poc_verify_claims(p_run uuid, p_snapshot uuid)
RETURNS void LANGUAGE plpgsql
    SECURITY DEFINER SET search_path = pg_catalog AS $fn$
DECLARE
    v_employer uuid; v_period uuid; v_start date; v_end date;
    v_members int; v_claims int; v_missing int; v_stale int;
    v_orphan int; v_badrel int; v_baddec int; v_uncovered int;
    r record;
BEGIN
    SELECT s.employer_id, s.pay_period_id INTO v_employer, v_period
      FROM public.payroll_snapshot s WHERE s.id = p_snapshot AND s.payroll_run_id = p_run;
    IF NOT FOUND THEN
        RAISE EXCEPTION 'poc_verify_claims: snapshot % not bound to run %', p_snapshot, p_run
            USING ERRCODE = 'P0001';
    END IF;
    SELECT p.starts_on, p.ends_on INTO v_start, v_end
      FROM public.pay_period p WHERE p.id = v_period;

    SELECT count(*) INTO v_members FROM public.payroll_snapshot_employee WHERE snapshot_id = p_snapshot;
    SELECT count(*) INTO v_claims FROM public.poc_claim WHERE run_id = p_run;
    IF v_members = 0 OR v_claims = 0 THEN
        RAISE EXCEPTION 'poc_verify_claims: vacuous evidence (members=%, claims=%)', v_members, v_claims
            USING ERRCODE = 'P0001';
    END IF;

    -- (a) Every claim names a real row with a matching live hash
    -- (catches forged claims, stale claims, and pack/invented rows).
    SELECT count(*) INTO v_stale FROM public.poc_claim c WHERE c.run_id = p_run
      AND (c.table_name = 'poc_rule_pack_schedule'
           AND c.input_hash IS DISTINCT FROM public.poc_pack_claim_hash(
               c.employer_id, split_part(c.row_id, '|', 1), split_part(c.row_id, '|', 2)));
    SELECT count(*) + v_stale INTO v_stale FROM public.poc_claim c WHERE c.run_id = p_run
      AND c.table_name <> 'poc_rule_pack_schedule'
      AND c.input_hash IS DISTINCT FROM public.poc_src_hash(c.table_name, c.row_id::uuid);
    IF v_stale > 0 THEN
        RAISE EXCEPTION 'poc_verify_claims: % claims stale or forged', v_stale
            USING ERRCODE = 'P0001';
    END IF;

    -- (b) Every expected source row is claimed (omission detection).
    -- Expected sets mirror the capture derivation exactly.
    SELECT count(*) INTO v_missing FROM (
        -- population membership
        SELECT 'employee' AS t, m.id::text AS r FROM public.employee m
         WHERE m.employer_id = v_employer
        UNION ALL
        -- overlapping terms
        SELECT 'employment_terms', t.id::text FROM public.employment_terms t
         WHERE t.employer_id = v_employer
           AND daterange(t.effective_from, COALESCE(t.effective_to, 'infinity'::date), '[]')
               && daterange(v_start, v_end, '[]')
        UNION ALL
        -- approved sheets for the period
        SELECT 'timesheet', s.id::text FROM public.timesheet s
         WHERE s.employer_id = v_employer AND s.pay_period_id = v_period AND s.status = 'approved'
        UNION ALL
        -- current work intervals in the period
        SELECT 'attendance_interval', i.id::text FROM public.attendance_interval i
         WHERE i.employer_id = v_employer AND i.is_current AND i.kind = 'work'
           AND i.local_date BETWEEN v_start AND v_end
        UNION ALL
        -- decided leave overlapping the period
        SELECT 'leave_request', l.id::text FROM public.leave_request l
         WHERE l.employer_id = v_employer AND l.status IN ('approved', 'rejected')
           AND daterange(l.from_date, l.to_date, '[]') && daterange(v_start, v_end, '[]')
        UNION ALL
        -- overlapping comp lines
        SELECT 'compensation_line', c.id::text FROM public.compensation_line c
         WHERE c.employer_id = v_employer
           AND daterange(c.effective_from, COALESCE(c.effective_to, 'infinity'::date), '[]')
               && daterange(v_start, v_end, '[]')
    ) exp
    WHERE NOT EXISTS (SELECT 1 FROM public.poc_claim c
                       WHERE c.run_id = p_run AND c.table_name = exp.t AND c.row_id = exp.r);
    -- pack pins claimed by "pack|version" row identity
    SELECT v_missing + count(*) INTO v_missing
      FROM public.poc_rule_pack_schedule s
     WHERE s.employer_id = v_employer
       AND daterange(s.effective_from, s.effective_to, '[]') && daterange(v_start, v_end, '[]')
       AND NOT EXISTS (SELECT 1 FROM public.poc_claim c
                        WHERE c.run_id = p_run AND c.table_name = 'poc_rule_pack_schedule'
                          AND c.row_id = s.pack || '|' || s.version);
    IF v_missing > 0 THEN
        RAISE EXCEPTION 'poc_verify_claims: % expected source rows unclaimed', v_missing
            USING ERRCODE = 'P0001';
    END IF;

    -- (c) No orphan claims (claimed row not in any expected set).
    -- (Covered by (a)+(b) jointly for known tables; this pins the
    -- table-name allowlist itself.)
    SELECT count(*) INTO v_orphan FROM public.poc_claim c WHERE c.run_id = p_run
      AND c.table_name NOT IN ('employee', 'employment_terms', 'timesheet',
          'attendance_interval', 'leave_request', 'compensation_line',
          'poc_rule_pack_schedule');
    IF v_orphan > 0 THEN
        RAISE EXCEPTION 'poc_verify_claims: % claims outside the source allowlist', v_orphan
            USING ERRCODE = 'P0001';
    END IF;

    -- (d) Member-to-source relationships: every member's pinned
    -- references must belong to that member's employee + period.
    SELECT count(*) INTO v_badrel FROM public.payroll_snapshot_employee e
     WHERE e.snapshot_id = p_snapshot AND (
        EXISTS (SELECT 1 FROM unnest(e.employment_terms_ids) tid
                 WHERE NOT EXISTS (SELECT 1 FROM public.employment_terms t
                                    WHERE t.id = tid AND t.employee_id = e.employee_id))
        OR (e.approved_sheet_id IS NOT NULL AND NOT EXISTS (
                SELECT 1 FROM public.timesheet s
                 WHERE s.id = e.approved_sheet_id AND s.employee_id = e.employee_id
                   AND s.pay_period_id = v_period AND s.status = 'approved'))
     );
    IF v_badrel > 0 THEN
        RAISE EXCEPTION 'poc_verify_claims: % members with invalid source relations', v_badrel
            USING ERRCODE = 'P0001';
    END IF;

    -- (e) Decision re-derivation: recompute every member decision from
    -- LIVE tables and require equality with the recorded one. The
    -- recorded approved_sheet_id is NOT trusted here: sheet existence
    -- is re-derived from the live timesheet table.
    SELECT count(*) INTO v_baddec FROM public.payroll_snapshot_employee e, LATERAL (
        SELECT CASE
            WHEN (SELECT count(DISTINCT t.monthly_basic)
                    FROM public.employment_terms t
                   WHERE t.employee_id = e.employee_id
                     AND daterange(t.effective_from, COALESCE(t.effective_to, 'infinity'::date), '[]')
                         && daterange(v_start, v_end, '[]')) > 1 THEN 'excluded'
            WHEN NOT EXISTS (SELECT 1 FROM public.timesheet s
                              WHERE s.employee_id = e.employee_id
                                AND s.pay_period_id = v_period AND s.status = 'approved')
             AND EXISTS (SELECT 1 FROM public.attendance_interval i
                          WHERE i.employee_id = e.employee_id AND i.is_current AND i.kind = 'work'
                            AND i.local_date BETWEEN v_start AND v_end) THEN 'missing-timesheet'
            ELSE 'included' END AS want
     ) d
     WHERE e.snapshot_id = p_snapshot AND e.decision IS DISTINCT FROM d.want;
    IF v_baddec > 0 THEN
        RAISE EXCEPTION 'poc_verify_claims: % member decisions disagree with live derivation', v_baddec
            USING ERRCODE = 'P0001';
    END IF;

    -- (f) Joint pack coverage: every period day covered by >= 1 pin.
    SELECT count(*) INTO v_uncovered FROM generate_series(v_start, v_end, '1 day') d
     WHERE NOT EXISTS (
        SELECT 1 FROM public.payroll_snapshot_pack p
         WHERE p.snapshot_id = p_snapshot
           AND d::date BETWEEN p.effective_from AND p.effective_to);
    IF v_uncovered > 0 THEN
        RAISE EXCEPTION 'poc_verify_claims: % period days without pack coverage', v_uncovered
            USING ERRCODE = 'P0001';
    END IF;
END;
$fn$;
ALTER FUNCTION poc_verify_claims(uuid, uuid) OWNER TO migration_admin;
REVOKE ALL ON FUNCTION poc_verify_claims(uuid, uuid) FROM PUBLIC;
-- NOTE: no GRANT to app (revoked in Part 2): capture/validate call it
-- as the DEFINER owner inside authorized transactions.

-- Read-only consistency probe (Phase C input-drift detection).
-- Compares LIVE source hashes against the run's recorded claims by
-- reusing the verifier (zero derivation duplication): returns
-- 'consistent', or 'changed: <first defect>' when any source row was
-- added, removed, or altered after capture. Duplicate workflow
-- attempts adopt on 'consistent' and raise an explicit conflict
-- otherwise — never a silent stale replay (DB-3 fix).
CREATE OR REPLACE FUNCTION poc_run_consistency(p_run uuid)
RETURNS text LANGUAGE plpgsql
    SECURITY DEFINER SET search_path = pg_catalog AS $fn$
DECLARE
    v_tip uuid;
BEGIN
    SELECT s.id INTO v_tip FROM public.payroll_snapshot s
     WHERE s.payroll_run_id = p_run AND s.status = 'valid';
    IF NOT FOUND THEN
        RETURN 'changed: no valid snapshot tip';
    END IF;
    BEGIN
        PERFORM public.poc_verify_claims(p_run, v_tip);
    EXCEPTION WHEN SQLSTATE 'P0001' THEN
        RETURN 'changed: ' || SQLERRM;
    END;
    RETURN 'consistent';
END;
$fn$;
ALTER FUNCTION poc_run_consistency(uuid) OWNER TO migration_admin;
REVOKE ALL ON FUNCTION poc_run_consistency(uuid) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION poc_run_consistency(uuid) TO app;

-- Workflow-attempt registry (observability + test assertions): links
-- each DBOS attempt to the business run it captured or adopted.
-- Attempts never decide authority (runs do); squatting a row here
-- moves no business state, so app INSERT is safe. No updates.
CREATE TABLE IF NOT EXISTS poc_workflow_attempt (
    workflow_id  text PRIMARY KEY,
    run_id       uuid REFERENCES payroll_run (id),
    employer_id  uuid NOT NULL,
    period_code  text NOT NULL,
    role         text NOT NULL CHECK (role IN ('captured', 'adopted')),
    input_digest text NOT NULL,
    created_at   timestamptz NOT NULL DEFAULT now()
);
REVOKE ALL ON poc_workflow_attempt FROM PUBLIC;
REVOKE ALL ON poc_workflow_attempt FROM app;
GRANT SELECT, INSERT ON poc_workflow_attempt TO app;

-- ============================================================ Part 5 ==
-- Authorized snapshot capture (the PG-1 fix).
--
-- The ONLY writer of snapshot/claim/run rows on this database. Derives
-- EVERYTHING from authoritative tables inside the caller's transaction
-- (the DBOS capture txn holds the employer gate throughout):
-- population, terms, sheets, intervals, leave, comp lines, pack pins,
-- decisions, blockers, row hashes, content hash. Caller supplies only
-- employer + period code + actor attribution; membership, inputs,
-- pins, digests and status are never caller-supplied.
--
-- POC rules (explicit, derived, reviewable):
-- - One regular run per employer+period; capture runs once per run.
-- - Population = every employee of the employer (ordered by no).
-- - >1 distinct monthly basic overlapping the period -> 'excluded'
--   with blocker B-BASE-CHANGE (proration is BLOCKED, never silent).
-- - Approved sheet missing but worked minutes exist ->
--   'missing-timesheet' (blocks validation: needs HR disposition).
-- - No sheet and no minutes -> 'included' with zero inputs.
-- - Leave day carrying worked minutes -> day excluded in inputs
--   (employee stays included; the calc skips the day).
--
-- Aggregate/schema qualification: aggregates live in pg_catalog
-- (pg_catalog.min/sum/...); pgcrypto (digest/encode/gen_random_uuid)
-- lives in public. No temp tables (app has no TEMP right and the
-- function must stay re-entrant): the 2-row pin set is re-selected.
CREATE OR REPLACE FUNCTION poc_snapshot_capture(
    p_employer uuid, p_period_code text, p_actor text,
    OUT o_run_id uuid, OUT o_snapshot_id uuid, OUT o_content_hash text)
RETURNS record LANGUAGE plpgsql
    SECURITY DEFINER SET search_path = pg_catalog AS $fn$
DECLARE
    v_period uuid; v_start date; v_end date; v_pstatus text;
    v_run uuid; v_snap uuid;
    v_members int := 0;
    v_canon text := '';
    emp record;
    v_terms uuid[]; v_nbases int; v_base_cents bigint;
    v_sheet uuid; v_haswork boolean;
    v_decision text; v_blockers jsonb;
    v_days jsonb; v_rowtext text; v_rowhash text;
    v_status_ids uuid[]; v_pintext text;
    v_allow_cents bigint; v_deduct_cents bigint;
BEGIN
    IF p_actor IS NULL OR length(trim(p_actor)) = 0 THEN
        RAISE EXCEPTION 'poc_snapshot_capture: actor attribution required';
    END IF;
    SELECT p.id, p.starts_on, p.ends_on, p.status
      INTO v_period, v_start, v_end, v_pstatus
      FROM public.pay_period p
     WHERE p.employer_id = p_employer AND p.code = p_period_code;
    IF NOT FOUND THEN
        RAISE EXCEPTION 'poc_snapshot_capture: period % unknown', p_period_code;
    END IF;
    IF v_pstatus NOT IN ('open', 'processing') THEN
        RAISE EXCEPTION 'poc_snapshot_capture: period % is %', p_period_code, v_pstatus;
    END IF;

    -- Authoritative synchronization: employer gate + population lock,
    -- same keys the writer triggers take (Part 7).
    PERFORM public.poc_gate_acquire(p_employer);
    PERFORM public.acquire_payroll_population_lock(p_employer);

    -- One regular run per employer+period; capture once per run.
    IF EXISTS (SELECT 1 FROM public.payroll_run r
                WHERE r.employer_id = p_employer AND r.pay_period_id = v_period
                  AND r.kind = 'regular' AND r.status = 'validated') THEN
        RAISE EXCEPTION 'poc_snapshot_capture: a validated run already stands; correction path required'
            USING ERRCODE = 'P0001';
    END IF;
    -- Find-or-create the draft (owner-side serialization: the
    -- partial unique index makes double-mint impossible; a lost race
    -- retries and finds the winner's row).
    LOOP
        SELECT r.id INTO v_run FROM public.payroll_run r
         WHERE r.employer_id = p_employer AND r.pay_period_id = v_period
           AND r.kind = 'regular' AND r.status = 'draft'
         ORDER BY r.created_at LIMIT 1 FOR UPDATE;
        IF FOUND THEN
            EXIT;
        END IF;
        BEGIN
            INSERT INTO public.payroll_run
                (id, employer_id, pay_period_id, kind, status, inputs_hash,
                 rule_versions, calc_version, rounding_conventions, started_at)
            VALUES (public.gen_random_uuid(), p_employer, v_period, 'regular', 'draft',
                    'pending', '{}', 'poc-synth/v1', '{}', pg_catalog.now())
            RETURNING id INTO v_run;
            EXIT;
        EXCEPTION WHEN unique_violation THEN
            -- Lost the mint race; loop back and find the winner.
        END;
    END LOOP;
    IF EXISTS (SELECT 1 FROM public.payroll_snapshot s
                WHERE s.payroll_run_id = v_run AND s.status = 'valid') THEN
        RAISE EXCEPTION 'poc_snapshot_capture: run already captured; correction path required'
            USING ERRCODE = 'P0001';
    END IF;
    IF NOT EXISTS (SELECT 1 FROM public.poc_rule_pack_schedule s
                    WHERE s.employer_id = p_employer
                      AND pg_catalog.daterange(s.effective_from, s.effective_to, '[]')
                          && pg_catalog.daterange(v_start, v_end, '[]')) THEN
        RAISE EXCEPTION 'poc_snapshot_capture: no rule-pack pins overlap the period'
            USING ERRCODE = 'P0001';
    END IF;

    -- Header first (members FK to it); provisional hash replaced below.
    INSERT INTO public.payroll_snapshot
        (employer_id, payroll_run_id, pay_period_id, version, status,
         content_hash, calc_engine, jurisdiction, currency, captured_by)
    VALUES (p_employer, v_run, v_period, 1, 'valid',
            'pending', 'poc-synth/v1', 'SG', 'SGD', p_actor)
    RETURNING id INTO v_snap;

    -- Population loop (deterministic order: employee_no).
    FOR emp IN SELECT m.id, m.employee_no FROM public.employee m
                WHERE m.employer_id = p_employer ORDER BY m.employee_no
    LOOP
        SELECT COALESCE(pg_catalog.array_agg(t.id ORDER BY t.effective_from), '{}'::uuid[]),
               pg_catalog.count(DISTINCT t.monthly_basic)
          INTO v_terms, v_nbases
          FROM public.employment_terms t
         WHERE t.employee_id = emp.id
           AND pg_catalog.daterange(t.effective_from,
                 COALESCE(t.effective_to, 'infinity'::date), '[]')
               && pg_catalog.daterange(v_start, v_end, '[]');
        SELECT COALESCE(pg_catalog.array_agg(h.id ORDER BY h.effective_from), '{}'::uuid[])
          INTO v_status_ids
          FROM public.employee_status_history h
         WHERE h.employee_id = emp.id
           AND pg_catalog.daterange(h.effective_from,
                 COALESCE(h.effective_to, 'infinity'::date), '[]')
               && pg_catalog.daterange(v_start, v_end, '[]');
        SELECT s.id INTO v_sheet FROM public.timesheet s
         WHERE s.employee_id = emp.id AND s.pay_period_id = v_period
           AND s.status = 'approved' ORDER BY s.approved_at DESC NULLS LAST LIMIT 1;
        SELECT EXISTS (SELECT 1 FROM public.attendance_interval i
                        WHERE i.employee_id = emp.id AND i.is_current AND i.kind = 'work'
                          AND i.local_date BETWEEN v_start AND v_end)
          INTO v_haswork;

        IF v_nbases > 1 THEN
            v_decision := 'excluded';
            v_blockers := pg_catalog.jsonb_build_array(pg_catalog.jsonb_build_object(
                'code', 'B-BASE-CHANGE',
                'message', 'more than one monthly basic overlaps the period: proration unresolved'));
            v_base_cents := NULL;
        ELSIF v_sheet IS NULL AND v_haswork THEN
            v_decision := 'missing-timesheet';
            v_blockers := pg_catalog.jsonb_build_array(pg_catalog.jsonb_build_object(
                'code', 'B-NO-SHEET',
                'message', 'worked minutes exist without an approved sheet: needs HR disposition'));
            SELECT (pg_catalog.min(t.monthly_basic) * 100)::bigint INTO v_base_cents
              FROM public.employment_terms t WHERE t.id = ANY (v_terms);
        ELSE
            v_decision := 'included';
            v_blockers := '[]'::jsonb;
            SELECT (pg_catalog.min(t.monthly_basic) * 100)::bigint INTO v_base_cents
              FROM public.employment_terms t WHERE t.id = ANY (v_terms);
        END IF;

        -- Per-day inputs (all period days, deterministic).
        SELECT COALESCE(pg_catalog.jsonb_agg(pg_catalog.jsonb_build_object(
                   'date', (d::date)::text,
                   'minutes', COALESCE(m.mins, 0),
                   'leave', COALESCE(l.leave, false),
                   'excluded', (COALESCE(m.mins, 0) > 0 AND COALESCE(l.leave, false))
               ) ORDER BY d), '[]'::jsonb)
          INTO v_days
          FROM pg_catalog.generate_series(v_start, v_end, '1 day') d
          LEFT JOIN LATERAL (
              SELECT pg_catalog.sum((EXTRACT(EPOCH FROM (i.interval_end - i.interval_start)) / 60)::bigint) AS mins
                FROM public.attendance_interval i
               WHERE i.employee_id = emp.id AND i.is_current AND i.kind = 'work'
                 AND i.local_date = d::date
          ) m ON true
          LEFT JOIN LATERAL (
              SELECT true AS leave
               WHERE EXISTS (SELECT 1 FROM public.leave_request l
                              WHERE l.employee_id = emp.id AND l.status = 'approved'
                                AND d::date BETWEEN l.from_date AND l.to_date)
          ) l ON true;

        -- As-given comp sums overlapping the period.
        SELECT COALESCE(pg_catalog.sum(c.amount_monthly) FILTER (WHERE c.kind = 'allowance'), 0) * 100,
               COALESCE(pg_catalog.sum(c.amount_monthly) FILTER (WHERE c.kind = 'deduction'), 0) * 100
          INTO v_allow_cents, v_deduct_cents
          FROM public.compensation_line c
         WHERE c.employee_id = emp.id
           AND pg_catalog.daterange(c.effective_from,
                 COALESCE(c.effective_to, 'infinity'::date), '[]')
               && pg_catalog.daterange(v_start, v_end, '[]');

        -- Canonical member text (explicit field order, ordered arrays).
        SELECT INTO v_rowtext
            emp.employee_no || '|' || COALESCE(v_base_cents::text, '') || '|' ||
            COALESCE((SELECT pg_catalog.string_agg(x::text, ',' ORDER BY x::text)
                        FROM pg_catalog.unnest(v_terms) x), '') || '|' ||
            COALESCE(v_sheet::text, '') || '|' ||
            COALESCE((SELECT pg_catalog.string_agg(
                        d->>'date' || ':' || (d->>'minutes') || ':' ||
                        (d->>'leave') || ':' || (d->>'excluded'), ';' ORDER BY d->>'date')
                        FROM pg_catalog.jsonb_array_elements(v_days) d), '') || '|' ||
            v_allow_cents::text || '|' || v_deduct_cents::text || '|' ||
            v_decision;
        v_rowhash := pg_catalog.encode(public.digest(v_rowtext, 'sha256'), 'hex');

        INSERT INTO public.payroll_snapshot_employee
            (snapshot_id, employee_id, employment_terms_ids, status_history_ids,
             employee_basis_seq, approved_sheet_id, decision, blockers, inputs, row_hash)
        VALUES (v_snap, emp.id, v_terms, v_status_ids,
                0, v_sheet, v_decision, v_blockers,
                pg_catalog.jsonb_build_object(
                    'employee_no', emp.employee_no,
                    'monthly_base_cents', v_base_cents,
                    'terms_ids', (SELECT COALESCE(pg_catalog.jsonb_agg(x ORDER BY x::text), '[]'::jsonb)
                                   FROM pg_catalog.unnest(v_terms) x),
                    'sheet_id', v_sheet,
                    'days', v_days,
                    'allow_cents', v_allow_cents,
                    'deduct_cents', v_deduct_cents,
                    'blockers', v_blockers),
                v_rowhash);
        v_members := v_members + 1;
        v_canon := v_canon || emp.employee_no || ':' || v_rowhash || ';';
    END LOOP;
    IF v_members = 0 THEN
        RAISE EXCEPTION 'poc_snapshot_capture: empty population cannot capture'
            USING ERRCODE = 'P0001';
    END IF;

    -- Pack canonical text (ordered by effective_from, pack).
    SELECT COALESCE(pg_catalog.string_agg(
               pack || '|' || version || '|' || digest, ';'
               ORDER BY effective_from, pack), '')
      INTO v_pintext
      FROM public.poc_rule_pack_schedule s
     WHERE s.employer_id = p_employer
       AND pg_catalog.daterange(s.effective_from, s.effective_to, '[]')
           && pg_catalog.daterange(v_start, v_end, '[]');
    v_canon := v_canon || v_pintext;

    -- Finalize header hash + pack pins.
    UPDATE public.payroll_snapshot s
       SET content_hash = pg_catalog.encode(public.digest(
               p_employer::text || '|' || p_period_code || '|' ||
               v_run::text || '|' || v_canon, 'sha256'), 'hex')
     WHERE s.id = v_snap
    RETURNING s.content_hash INTO o_content_hash;
    INSERT INTO public.payroll_snapshot_pack
        (snapshot_id, pack, version, status, manifest_digest,
         effective_from, effective_to)
    SELECT v_snap, pack, version, status, digest, effective_from, effective_to
      FROM public.poc_rule_pack_schedule s
     WHERE s.employer_id = p_employer
       AND pg_catalog.daterange(s.effective_from, s.effective_to, '[]')
           && pg_catalog.daterange(v_start, v_end, '[]');

    -- Claims (authorized writer: this function only).
    INSERT INTO public.poc_claim (run_id, employer_id, table_name, row_id, input_hash)
    SELECT v_run, p_employer, 'employee', m.id::text, public.poc_src_hash('employee', m.id)
      FROM public.employee m WHERE m.employer_id = p_employer;
    INSERT INTO public.poc_claim (run_id, employer_id, table_name, row_id, input_hash)
    SELECT v_run, p_employer, 'employment_terms', t.id::text, public.poc_src_hash('employment_terms', t.id)
      FROM public.employment_terms t
     WHERE t.employer_id = p_employer
       AND pg_catalog.daterange(t.effective_from, COALESCE(t.effective_to, 'infinity'::date), '[]')
           && pg_catalog.daterange(v_start, v_end, '[]');
    INSERT INTO public.poc_claim (run_id, employer_id, table_name, row_id, input_hash)
    SELECT v_run, p_employer, 'timesheet', s.id::text, public.poc_src_hash('timesheet', s.id)
      FROM public.timesheet s
     WHERE s.employer_id = p_employer AND s.pay_period_id = v_period AND s.status = 'approved';
    INSERT INTO public.poc_claim (run_id, employer_id, table_name, row_id, input_hash)
    SELECT v_run, p_employer, 'attendance_interval', i.id::text,
           public.poc_src_hash('attendance_interval', i.id)
      FROM public.attendance_interval i
     WHERE i.employer_id = p_employer AND i.is_current AND i.kind = 'work'
       AND i.local_date BETWEEN v_start AND v_end;
    INSERT INTO public.poc_claim (run_id, employer_id, table_name, row_id, input_hash)
    SELECT v_run, p_employer, 'leave_request', l.id::text, public.poc_src_hash('leave_request', l.id)
      FROM public.leave_request l
     WHERE l.employer_id = p_employer AND l.status IN ('approved', 'rejected')
       AND pg_catalog.daterange(l.from_date, l.to_date, '[]')
           && pg_catalog.daterange(v_start, v_end, '[]');
    INSERT INTO public.poc_claim (run_id, employer_id, table_name, row_id, input_hash)
    SELECT v_run, p_employer, 'compensation_line', c.id::text,
           public.poc_src_hash('compensation_line', c.id)
      FROM public.compensation_line c
     WHERE c.employer_id = p_employer
       AND pg_catalog.daterange(c.effective_from, COALESCE(c.effective_to, 'infinity'::date), '[]')
           && pg_catalog.daterange(v_start, v_end, '[]');
    INSERT INTO public.poc_claim (run_id, employer_id, table_name, row_id, input_hash)
    SELECT v_run, p_employer, 'poc_rule_pack_schedule', s.pack || '|' || s.version,
           public.poc_pack_claim_hash(p_employer, s.pack, s.version)
      FROM public.poc_rule_pack_schedule s
     WHERE s.employer_id = p_employer
       AND pg_catalog.daterange(s.effective_from, s.effective_to, '[]')
           && pg_catalog.daterange(v_start, v_end, '[]');

    -- Run derived fields (never caller-supplied).
    UPDATE public.payroll_run r
       SET inputs_hash = o_content_hash,
           rule_versions = (SELECT COALESCE(pg_catalog.jsonb_agg(
                                  pg_catalog.jsonb_build_object('pack', pack, 'version', version,
                                      'digest', digest) ORDER BY effective_from, pack),
                                  '[]'::jsonb)
                              FROM public.poc_rule_pack_schedule s
                             WHERE s.employer_id = p_employer
                               AND pg_catalog.daterange(s.effective_from, s.effective_to, '[]')
                                   && pg_catalog.daterange(v_start, v_end, '[]'))
     WHERE r.id = v_run;

    -- In-transaction verification (both directions; failure rolls
    -- back the entire capture).
    PERFORM public.poc_verify_claims(v_run, v_snap);

    INSERT INTO public.audit_event (employer_id, actor, service, action, entity, entity_id)
    VALUES (p_employer, p_actor, 'payroll', 'snapshot.capture', 'payroll_snapshot', v_snap);

    o_run_id := v_run;
    o_snapshot_id := v_snap;
END;
$fn$;
ALTER FUNCTION poc_snapshot_capture(uuid, text, text) OWNER TO migration_admin;
REVOKE ALL ON FUNCTION poc_snapshot_capture(uuid, text, text) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION poc_snapshot_capture(uuid, text, text) TO app;

-- ============================================================ Part 6 ==
-- Calculation-result persistence, approval, and run validation.
--
-- poc_run_output records the recomputed output digest (the auditor
-- and the finalization check read it; writers are DEFINER-only).
-- poc_approval records the POC approval decision (one per run).
CREATE TABLE IF NOT EXISTS poc_run_output (
    run_id       uuid PRIMARY KEY REFERENCES payroll_run (id),
    employer_id  uuid NOT NULL,
    output_digest text NOT NULL,
    lines        int NOT NULL,
    payslips     int NOT NULL,
    persisted_at timestamptz NOT NULL DEFAULT now(),
    persisted_by text NOT NULL
);
REVOKE ALL ON poc_run_output FROM PUBLIC;
REVOKE ALL ON poc_run_output FROM app;
GRANT SELECT ON poc_run_output TO app;
-- Binds results to the inputs they were computed from (adopt-path
-- verification). Backfills from the run's recorded inputs hash.
ALTER TABLE public.poc_run_output ADD COLUMN IF NOT EXISTS input_digest text;
UPDATE public.poc_run_output o
   SET input_digest = r.inputs_hash
  FROM public.payroll_run r
 WHERE o.run_id = r.id AND o.input_digest IS NULL;
ALTER TABLE public.poc_run_output ALTER COLUMN input_digest SET NOT NULL;

CREATE TABLE IF NOT EXISTS poc_approval (
    run_id     uuid PRIMARY KEY REFERENCES payroll_run (id),
    employer_id uuid NOT NULL,
    approver   text NOT NULL,
    decision   text NOT NULL CHECK (decision IN ('approve', 'reject')),
    decided_at timestamptz NOT NULL DEFAULT now(),
    note       text NOT NULL DEFAULT ''
);
REVOKE ALL ON poc_approval FROM PUBLIC;
REVOKE ALL ON poc_approval FROM app;
GRANT SELECT ON poc_approval TO app;

-- Authorized result persistence (Stage D writer).
--
-- Binds every persisted row to the captured snapshot: the run must be
-- a draft, exactly one valid tip must stand, the caller-supplied input
-- digest must EQUAL the tip's content hash, and the calc version must
-- match the run's pinned version. Only 'included' members may receive
-- rows (excluded members carry no financial effect). Amounts arrive as
-- exact integer cents and convert without float. The output digest is
-- recomputed over the inserted rows (never caller-supplied).
--
-- p_lines: [{employee_no, component, pay_code, hours|null,
--   amount_cents, cpf_treatment, work_period, payable_date,
--   contribution_period, source_refs}]
-- p_payslips: [{employee_no, payload}]
CREATE OR REPLACE FUNCTION poc_result_persist(
    p_run uuid, p_calc_version text, p_input_digest text,
    p_lines jsonb, p_payslips jsonb, p_actor text)
RETURNS text LANGUAGE plpgsql
    SECURITY DEFINER SET search_path = pg_catalog AS $fn$
DECLARE
    v_employer uuid; v_status text; v_tip uuid; v_tiphash text; v_calc text;
    v_canon text := '';
    ln jsonb; sl jsonb; v_emp uuid; v_dec text;
    v_nlines int := 0; v_nslips int := 0;
    v_digest text;
BEGIN
    IF p_actor IS NULL OR length(trim(p_actor)) = 0 THEN
        RAISE EXCEPTION 'poc_result_persist: actor attribution required';
    END IF;
    SELECT r.employer_id, r.status, r.calc_version
      INTO v_employer, v_status, v_calc
      FROM public.payroll_run r WHERE r.id = p_run FOR UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION 'poc_result_persist: run % unknown', p_run;
    END IF;
    IF v_status <> 'draft' THEN
        RAISE EXCEPTION 'poc_result_persist: run is %, want draft', v_status;
    END IF;
    SELECT s.id, s.content_hash INTO v_tip, v_tiphash
      FROM public.payroll_snapshot s WHERE s.payroll_run_id = p_run AND s.status = 'valid';
    IF NOT FOUND THEN
        RAISE EXCEPTION 'poc_result_persist: no valid snapshot tip' USING ERRCODE = 'P0001';
    END IF;
    IF v_tiphash IS DISTINCT FROM p_input_digest THEN
        RAISE EXCEPTION 'poc_result_persist: input digest mismatch (stale capture)'
            USING ERRCODE = 'P0001';
    END IF;
    IF v_calc IS DISTINCT FROM p_calc_version THEN
        RAISE EXCEPTION 'poc_result_persist: calc version % does not match pinned %',
            p_calc_version, v_calc USING ERRCODE = 'P0001';
    END IF;
    IF EXISTS (SELECT 1 FROM public.payroll_line l WHERE l.run_id = p_run) THEN
        RAISE EXCEPTION 'poc_result_persist: results already persisted for run'
            USING ERRCODE = 'P0001';
    END IF;
    IF p_lines IS NULL OR pg_catalog.jsonb_typeof(p_lines) <> 'array'
       OR pg_catalog.jsonb_array_length(p_lines) = 0 THEN
        RAISE EXCEPTION 'poc_result_persist: empty result set cannot persist'
            USING ERRCODE = 'P0001';
    END IF;
    -- Trigger interop: the frozen payroll_line_freeze trigger resolves
    -- `payroll_run` unqualified with no SET search_path of its own, so
    -- it inherits this chain's pg_catalog-only path and fails 42P01
    -- (proven live). Widen resolution for THIS TRANSACTION ONLY; the
    -- function's own references stay fully qualified (asserted). Safe:
    -- app holds no CREATE on schema public and no TEMP right, so no
    -- shadow surface exists. Production must SET search_path on the
    -- frozen trigger instead (see the proposal doc §5).
    PERFORM pg_catalog.set_config('search_path', 'public, pg_catalog', true);
    PERFORM public.poc_gate_acquire(v_employer);

    FOR ln IN SELECT * FROM pg_catalog.jsonb_array_elements(p_lines) LOOP
        SELECT m.id INTO v_emp FROM public.employee m
         WHERE m.employer_id = v_employer AND m.employee_no = ln->>'employee_no';
        IF NOT FOUND THEN
            RAISE EXCEPTION 'poc_result_persist: unknown employee %', ln->>'employee_no';
        END IF;
        SELECT e.decision INTO v_dec FROM public.payroll_snapshot_employee e
         WHERE e.snapshot_id = v_tip AND e.employee_id = v_emp;
        IF NOT FOUND OR v_dec IS DISTINCT FROM 'included' THEN
            RAISE EXCEPTION 'poc_result_persist: employee % not an included member',
                ln->>'employee_no' USING ERRCODE = 'P0001';
        END IF;
        INSERT INTO public.payroll_line
            (id, employer_id, run_id, employee_id, component, pay_code,
             hours, amount, cpf_treatment, work_period, payable_date,
             contribution_period, source_refs)
        VALUES (public.gen_random_uuid(), v_employer, p_run, v_emp,
                ln->>'component', ln->>'pay_code',
                CASE WHEN ln->>'hours' IS NULL THEN NULL ELSE (ln->>'hours')::numeric END,
                ((ln->>'amount_cents')::bigint)::numeric / 100,
                ln->>'cpf_treatment', ln->>'work_period', (ln->>'payable_date')::date,
                ln->>'contribution_period', ln->'source_refs');
        v_nlines := v_nlines + 1;
        v_canon := v_canon || (ln->>'employee_no') || '|' || (ln->>'component') || '|' ||
                   (ln->>'pay_code') || '|' || (ln->>'amount_cents') || ';';
    END LOOP;

    IF p_payslips IS NOT NULL THEN
        FOR sl IN SELECT * FROM pg_catalog.jsonb_array_elements(p_payslips) LOOP
            SELECT m.id INTO v_emp FROM public.employee m
             WHERE m.employer_id = v_employer AND m.employee_no = sl->>'employee_no';
            IF NOT FOUND THEN
                RAISE EXCEPTION 'poc_result_persist: unknown payslip employee %', sl->>'employee_no';
            END IF;
            INSERT INTO public.payslip (id, employer_id, run_id, employee_id, payload)
            VALUES (public.gen_random_uuid(), v_employer, p_run, v_emp, sl->'payload');
            v_nslips := v_nslips + 1;
        END LOOP;
    END IF;

    v_digest := pg_catalog.encode(
        public.digest('output|' || p_run::text || '|' || v_canon, 'sha256'), 'hex');
    INSERT INTO public.poc_run_output (run_id, employer_id, output_digest, input_digest, lines, payslips, persisted_by)
    VALUES (p_run, v_employer, v_digest, p_input_digest, v_nlines, v_nslips, p_actor);
    INSERT INTO public.audit_event (employer_id, actor, service, action, entity, entity_id)
    VALUES (v_employer, p_actor, 'payroll', 'result.persist', 'payroll_run', p_run);
    RETURN v_digest;
END;
$fn$;
ALTER FUNCTION poc_result_persist(uuid, text, text, jsonb, jsonb, text) OWNER TO migration_admin;
REVOKE ALL ON FUNCTION poc_result_persist(uuid, text, text, jsonb, jsonb, text) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION poc_result_persist(uuid, text, text, jsonb, jsonb, text) TO app;

-- Authorized approval (POC: operator label, no auth subsystem).
-- The decision is verified against business authority: results must
-- exist and no member may sit in 'missing-timesheet' ('excluded'
-- members are explicit and lawful).
CREATE OR REPLACE FUNCTION poc_approve(p_run uuid, p_approver text, p_decision text, p_note text)
RETURNS text LANGUAGE plpgsql
    SECURITY DEFINER SET search_path = pg_catalog AS $fn$
DECLARE
    v_employer uuid; v_status text;
BEGIN
    IF p_approver IS NULL OR length(trim(p_approver)) = 0 THEN
        RAISE EXCEPTION 'poc_approve: approver attribution required';
    END IF;
    IF p_decision NOT IN ('approve', 'reject') THEN
        RAISE EXCEPTION 'poc_approve: decision must be approve|reject, not %', p_decision;
    END IF;
    SELECT r.employer_id, r.status INTO v_employer, v_status
      FROM public.payroll_run r WHERE r.id = p_run FOR UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION 'poc_approve: run % unknown', p_run;
    END IF;
    IF v_status <> 'draft' THEN
        RAISE EXCEPTION 'poc_approve: run is %, want draft', v_status;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM public.payroll_line l WHERE l.run_id = p_run) THEN
        RAISE EXCEPTION 'poc_approve: no persisted results to approve' USING ERRCODE = 'P0001';
    END IF;
    IF EXISTS (SELECT 1 FROM public.payroll_snapshot_employee e
                JOIN public.payroll_snapshot s ON s.id = e.snapshot_id
               WHERE s.payroll_run_id = p_run AND s.status = 'valid'
                 AND e.decision = 'missing-timesheet') THEN
        RAISE EXCEPTION 'poc_approve: unresolved missing-timesheet members' USING ERRCODE = 'P0001';
    END IF;
    IF EXISTS (SELECT 1 FROM public.poc_approval a WHERE a.run_id = p_run) THEN
        RAISE EXCEPTION 'poc_approve: run already decided' USING ERRCODE = 'P0001';
    END IF;
    INSERT INTO public.poc_approval (run_id, employer_id, approver, decision, note)
    VALUES (p_run, v_employer, p_approver, p_decision, COALESCE(p_note, ''));
    INSERT INTO public.audit_event (employer_id, actor, service, action, entity, entity_id)
    VALUES (v_employer, p_approver, 'payroll', 'run.' || p_decision, 'payroll_run', p_run);
    RETURN p_decision;
END;
$fn$;
ALTER FUNCTION poc_approve(uuid, text, text, text) OWNER TO migration_admin;
REVOKE ALL ON FUNCTION poc_approve(uuid, text, text, text) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION poc_approve(uuid, text, text, text) TO app;

-- Authorized run validation (finalization gate).
--
-- Re-verifies against LIVE tables (not recorded labels): exactly one
-- valid tip, claims verifier green, header binding recomputed from
-- current members+pins, approval present, output recorded. The frozen
-- run_lifecycle trigger re-checks as a backstop on the UPDATE.
CREATE OR REPLACE FUNCTION poc_run_validate(p_run uuid, p_actor text)
RETURNS text LANGUAGE plpgsql
    SECURITY DEFINER SET search_path = pg_catalog AS $fn$
DECLARE
    v_employer uuid; v_period uuid; v_pcode text; v_status text;
    v_tip uuid; v_tiphash text; v_recomputed text;
BEGIN
    IF p_actor IS NULL OR length(trim(p_actor)) = 0 THEN
        RAISE EXCEPTION 'poc_run_validate: actor attribution required';
    END IF;
    SELECT r.employer_id, r.pay_period_id, r.status
      INTO v_employer, v_period, v_status
      FROM public.payroll_run r WHERE r.id = p_run FOR UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION 'poc_run_validate: run % unknown', p_run;
    END IF;
    IF v_status <> 'draft' THEN
        RAISE EXCEPTION 'poc_run_validate: run is %, want draft', v_status;
    END IF;
    SELECT p.code INTO v_pcode FROM public.pay_period p WHERE p.id = v_period;
    PERFORM public.poc_gate_acquire(v_employer);

    SELECT s.id, s.content_hash INTO v_tip, v_tiphash
      FROM public.payroll_snapshot s WHERE s.payroll_run_id = p_run AND s.status = 'valid';
    IF NOT FOUND THEN
        RAISE EXCEPTION 'poc_run_validate: no valid snapshot tip' USING ERRCODE = 'P0001';
    END IF;
    -- Live re-verification (claims, relations, decisions, coverage).
    PERFORM public.poc_verify_claims(p_run, v_tip);
    -- Header binding recomputed from CURRENT members + pins.
    SELECT pg_catalog.encode(public.digest(
        v_employer::text || '|' || v_pcode || '|' || p_run::text || '|' ||
        (SELECT pg_catalog.string_agg(m.employee_no || ':' || e.row_hash, ';' ORDER BY m.employee_no)
           FROM public.payroll_snapshot_employee e
           JOIN public.employee m ON m.id = e.employee_id
          WHERE e.snapshot_id = v_tip) || ';' ||
        (SELECT COALESCE(pg_catalog.string_agg(
                   p.pack || '|' || p.version || '|' || p.manifest_digest, ';'
                   ORDER BY p.effective_from, p.pack), '')
           FROM public.payroll_snapshot_pack p WHERE p.snapshot_id = v_tip),
        'sha256'), 'hex')
      INTO v_recomputed;
    IF v_recomputed IS DISTINCT FROM v_tiphash THEN
        RAISE EXCEPTION 'poc_run_validate: header binding mismatch (members or pins changed)'
            USING ERRCODE = 'P0001';
    END IF;
    IF NOT EXISTS (SELECT 1 FROM public.poc_approval a
                    WHERE a.run_id = p_run AND a.decision = 'approve') THEN
        RAISE EXCEPTION 'poc_run_validate: no recorded approval' USING ERRCODE = 'P0001';
    END IF;
    IF NOT EXISTS (SELECT 1 FROM public.poc_run_output o WHERE o.run_id = p_run) THEN
        RAISE EXCEPTION 'poc_run_validate: no persisted output digest' USING ERRCODE = 'P0001';
    END IF;
    UPDATE public.payroll_run r SET status = 'validated' WHERE r.id = p_run;
    INSERT INTO public.audit_event (employer_id, actor, service, action, entity, entity_id)
    VALUES (v_employer, p_actor, 'payroll', 'run.validate', 'payroll_run', p_run);
    RETURN 'validated';
END;
$fn$;
ALTER FUNCTION poc_run_validate(uuid, text) OWNER TO migration_admin;
REVOKE ALL ON FUNCTION poc_run_validate(uuid, text) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION poc_run_validate(uuid, text) TO app;

-- ============================================================ Part 7 ==
-- Period close + writer gate participation (PG-2/PG-3/PG-6/PG-10 fix).
--
-- Close is a DEFINER-only operation (direct INSERT revoked in Part 2,
-- so squatting and pre-dating are impossible): it resolves the REAL
-- pay_period by code, requires a validated run for that employer and
-- period, stamps the DB clock, locks the period row, and records the
-- close. Re-close raises loudly (duplicate handling).
CREATE OR REPLACE FUNCTION poc_close_period(p_employer uuid, p_period text, p_run uuid)
RETURNS timestamptz LANGUAGE plpgsql
    SECURITY DEFINER SET search_path = pg_catalog AS $fn$
DECLARE
    v_period uuid; v_pstatus text; v_now timestamptz;
BEGIN
    SELECT p.id, p.status INTO v_period, v_pstatus
      FROM public.pay_period p
     WHERE p.employer_id = p_employer AND p.code = p_period;
    IF NOT FOUND THEN
        RAISE EXCEPTION 'poc_close_period: period % unknown', p_period;
    END IF;
    IF v_pstatus = 'locked' THEN
        RAISE EXCEPTION 'poc_close_period: period % already locked', p_period
            USING ERRCODE = 'P0001';
    END IF;
    IF NOT EXISTS (SELECT 1 FROM public.payroll_run r
                    WHERE r.id = p_run AND r.employer_id = p_employer
                      AND r.pay_period_id = v_period AND r.status = 'validated') THEN
        RAISE EXCEPTION 'poc_close_period: run is not a validated run of this period'
            USING ERRCODE = 'P0001';
    END IF;
    PERFORM public.poc_gate_acquire(p_employer);
    v_now := pg_catalog.now();
    -- Lock the period row FIRST: the writer trigger (below) rejects
    -- writes to closed periods, so inserting the close row first would
    -- self-block this UPDATE. Single transaction: no crash window.
    UPDATE public.pay_period p SET status = 'locked' WHERE p.id = v_period;
    INSERT INTO public.poc_period_close (employer_id, period, run_id, closed_at, closed_by)
    VALUES (p_employer, p_period, p_run, v_now, current_user)
    ON CONFLICT (employer_id, period) DO NOTHING;
    IF NOT FOUND THEN
        RAISE EXCEPTION 'poc_close_period: close row for % already exists', p_period
            USING ERRCODE = 'P0001';
    END IF;
    INSERT INTO public.audit_event (employer_id, actor, service, action, entity, entity_id)
    VALUES (p_employer, current_user, 'payroll', 'period.close', 'pay_period', v_period);
    RETURN v_now;
END;
$fn$;
ALTER FUNCTION poc_close_period(uuid, text, uuid) OWNER TO migration_admin;
REVOKE ALL ON FUNCTION poc_close_period(uuid, text, uuid) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION poc_close_period(uuid, text, uuid) TO app;

-- Period-code derivation for the writer trigger: real schema
-- relationships only (pay_period_id joins). Tables WITHOUT a period
-- relationship return NULL and are gate-only (documented limitation:
-- clock/attendance/terms rows carry dates, not period membership;
-- production would map them via roster/pay-calendar joins).
CREATE OR REPLACE FUNCTION poc_writer_period_code(p_table text, p_row jsonb)
RETURNS text LANGUAGE plpgsql
    SECURITY DEFINER SET search_path = pg_catalog AS $fn$
DECLARE
    v_code text;
BEGIN
    CASE p_table
    WHEN 'timesheet', 'timesheet_line', 'payroll_run' THEN
        SELECT p.code INTO v_code FROM public.pay_period p
         WHERE p.id = (p_row->>'pay_period_id')::uuid;
    WHEN 'timesheet_line_interval' THEN
        SELECT p.code INTO v_code
          FROM public.timesheet_line l JOIN public.pay_period p ON p.id = l.pay_period_id
         WHERE l.id = (p_row->>'line_id')::uuid;
    WHEN 'pay_period' THEN
        SELECT p.code INTO v_code FROM public.pay_period p
         WHERE p.id = (p_row->>'id')::uuid;
    WHEN 'payroll_run_exclusion' THEN
        SELECT p.code INTO v_code
          FROM public.payroll_run r JOIN public.pay_period p ON p.id = r.pay_period_id
         WHERE r.id = (p_row->>'payroll_run_id')::uuid;
    ELSE
        RETURN NULL;
    END CASE;
    RETURN v_code;
END;
$fn$;
ALTER FUNCTION poc_writer_period_code(text, jsonb) OWNER TO migration_admin;
REVOKE ALL ON FUNCTION poc_writer_period_code(text, jsonb) FROM PUBLIC;

-- Writer participation trigger, rewritten (replaces the
-- gate2_poc.sql version):
-- - DEFINER (owner): writers reach the gate ONLY through triggers;
--   direct app EXECUTE on poc_gate_acquire is revoked (Part 2).
-- - Covers INSERT OR UPDATE OR DELETE (DELETE bypass closed, PG-10a).
-- - Cross-employer UPDATE takes both employer keys in sorted order
--   (PG-10b); single-employer writes take one key.
-- - Closed-period rejection uses poc_writer_period_code (PG-3 fix);
--   unmapped tables are gate-only (see above).
-- DROP + CREATE: OR REPLACE cannot change the security attribute.
DROP FUNCTION IF EXISTS public.poc_require_gate() CASCADE;
CREATE FUNCTION poc_require_gate()
RETURNS trigger LANGUAGE plpgsql
    SECURITY DEFINER SET search_path = pg_catalog AS $fn$
DECLARE
    v_row jsonb; v_employer uuid; v_old uuid;
    v_first uuid; v_second uuid;
    v_code text; v_closed boolean;
BEGIN
    IF TG_OP = 'DELETE' THEN
        v_row := pg_catalog.to_jsonb(OLD);
    ELSE
        v_row := pg_catalog.to_jsonb(NEW);
    END IF;
    v_employer := (v_row->>'employer_id')::uuid;
    IF v_employer IS NULL THEN
        RAISE EXCEPTION 'poc_require_gate: no employer identity on %', TG_TABLE_NAME;
    END IF;
    -- Sorted dual-key acquisition for cross-employer UPDATE.
    v_first := v_employer; v_second := NULL;
    IF TG_OP = 'UPDATE' THEN
        v_old := (pg_catalog.to_jsonb(OLD)->>'employer_id')::uuid;
        IF v_old IS DISTINCT FROM v_employer AND v_old IS NOT NULL THEN
            IF v_old < v_employer THEN
                v_first := v_old; v_second := v_employer;
            ELSE
                v_first := v_employer; v_second := v_old;
            END IF;
        END IF;
    END IF;
    -- Closed-period rejection via real period relationships.
    v_code := public.poc_writer_period_code(TG_TABLE_NAME, v_row);
    IF v_code IS NOT NULL THEN
        SELECT EXISTS (
            SELECT 1 FROM public.poc_period_close c
             WHERE c.employer_id = v_employer AND c.period = v_code
               AND c.closed_at IS NOT NULL
        ) INTO v_closed;
        IF v_closed THEN
            RAISE EXCEPTION 'poc_require_gate: period % closed for employer %', v_code, v_employer
                USING ERRCODE = 'P0001';
        END IF;
    END IF;
    PERFORM public.poc_gate_acquire(v_first);
    IF v_second IS NOT NULL THEN
        PERFORM public.poc_gate_acquire(v_second);
    END IF;
    IF TG_OP = 'DELETE' THEN
        RETURN OLD;
    END IF;
    RETURN NEW;
END;
$fn$;
ALTER FUNCTION poc_require_gate() OWNER TO migration_admin;
REVOKE ALL ON FUNCTION poc_require_gate() FROM PUBLIC;
-- No GRANT: reachable via triggers only (which run as the DEFINER
-- owner regardless of the invoking role).

-- Trigger coverage: the 12 capture-input tables (now with DELETE)
-- plus the run/period/exclusion writers, so the validate/transition
-- path cannot race capture outside the gate (PG-10d).
DO $do$
DECLARE
    t text;
BEGIN
    FOREACH t IN ARRAY ARRAY[
        'attendance_interval', 'clock_event', 'clock_event_disposition',
        'compensation_line', 'employee', 'employee_status_history',
        'employment_terms', 'rule_pack_activation', 'shg_instruction',
        'timesheet', 'timesheet_line', 'timesheet_line_interval',
        'payroll_run', 'pay_period', 'payroll_run_exclusion'
    ] LOOP
        EXECUTE format('DROP TRIGGER IF EXISTS poc_gate_participate ON %I', t);
        EXECUTE format('CREATE TRIGGER poc_gate_participate '
            || 'BEFORE INSERT OR UPDATE OR DELETE ON %I '
            || 'FOR EACH ROW EXECUTE FUNCTION poc_require_gate()', t);
    END LOOP;
END;
$do$;

-- ============================================================ Part 8 ==
-- Outbox authority (PG-5/IM-12 fix) + backlink identity (IM-11 fix).
--
-- States distinguish the recovery substates prompt8 §5.6 requires:
-- pending -> appended -> backlinked (happy path);
-- pending -> failed (retryable, append never proven) or poison
-- (terminal, append never proven); failed -> pending (retry reset)
-- or poison; appended -> reconciling -> backlinked|reconciled
-- (backlink-missing recovery, stays actionable forever).
-- IM-10 closure: a PROVEN append (immu_tx set) can never move to
-- failed/poison — retry accounting cannot orphan attested evidence.
-- 'reconciled' = recovered via reconciliation (duplicate convergence
-- is an ordinary success, never a worker error).
ALTER TABLE public.poc_outbox DROP CONSTRAINT IF EXISTS poc_outbox_state_check;
ALTER TABLE public.poc_outbox ADD CONSTRAINT poc_outbox_state_check CHECK (
    state IN ('pending', 'appended', 'backlinked', 'failed',
              'reconciling', 'reconciled', 'poison'));

-- Transition guard: the worker (app role) may move rows along the DAG
-- but can never forge completion (skip states), mutate evidence
-- payloads, substitute digests, or rewind terminal states. Payload,
-- hash, identity and action columns are immutable after enqueue.
-- State/tx invariant (IM-10): immu_tx is set exactly for proven
-- appends (appended/reconciling/backlinked/reconciled) and NULL
-- exactly otherwise (pending/failed/poison); a proven append can
-- never be discarded to failed/poison by retry accounting.
CREATE OR REPLACE FUNCTION poc_outbox_guard()
RETURNS trigger LANGUAGE plpgsql
    SECURITY DEFINER SET search_path = pg_catalog AS $fn$
BEGIN
    IF TG_OP = 'DELETE' THEN
        -- Owner-only hygiene (test scaffolding); app DELETE is revoked.
        IF current_user <> 'migration_admin' THEN
            RAISE EXCEPTION 'poc_outbox: DELETE reserved for the owner'
                USING ERRCODE = 'P0001';
        END IF;
        RETURN OLD;
    END IF;
    IF TG_OP = 'INSERT' THEN
        -- Construction invariant only (no OLD row to compare).
        IF NEW.state IN ('appended', 'reconciling', 'backlinked', 'reconciled')
           AND NEW.immu_tx IS NULL THEN
            RAISE EXCEPTION 'poc_outbox: state % requires a proven append', NEW.state
                USING ERRCODE = 'P0001';
        END IF;
        IF NEW.state IN ('pending', 'failed', 'poison')
           AND NEW.immu_tx IS NOT NULL THEN
            RAISE EXCEPTION 'poc_outbox: state % forbids an anchored tx', NEW.state
                USING ERRCODE = 'P0001';
        END IF;
        RETURN NEW;
    END IF;
    IF NEW.op_uuid IS DISTINCT FROM OLD.op_uuid
       OR NEW.employer_id IS DISTINCT FROM OLD.employer_id
       OR NEW.action IS DISTINCT FROM OLD.action
       OR NEW.payload_hash IS DISTINCT FROM OLD.payload_hash
       OR NEW.payload IS DISTINCT FROM OLD.payload THEN
        RAISE EXCEPTION 'poc_outbox: evidence identity is immutable'
            USING ERRCODE = 'P0001';
    END IF;
    IF NEW.attempts < OLD.attempts THEN
        RAISE EXCEPTION 'poc_outbox: attempt accounting never rewinds'
            USING ERRCODE = 'P0001';
    END IF;
    IF OLD.immu_tx IS NOT NULL AND NEW.immu_tx IS DISTINCT FROM OLD.immu_tx THEN
        RAISE EXCEPTION 'poc_outbox: anchored tx reference is immutable'
            USING ERRCODE = 'P0001';
    END IF;
    IF NEW.state IN ('appended', 'reconciling', 'backlinked', 'reconciled')
       AND NEW.immu_tx IS NULL THEN
        RAISE EXCEPTION 'poc_outbox: state % requires a proven append', NEW.state
            USING ERRCODE = 'P0001';
    END IF;
    IF NEW.state IN ('pending', 'failed', 'poison')
       AND NEW.immu_tx IS NOT NULL THEN
        RAISE EXCEPTION 'poc_outbox: proven append cannot move to %', NEW.state
            USING ERRCODE = 'P0001';
    END IF;
    IF NEW.state = OLD.state THEN
        RETURN NEW; -- touch-ups (attempts, updated_at) allowed
    END IF;
    IF (OLD.state = 'pending' AND NEW.state IN ('appended', 'failed', 'poison'))
       OR (OLD.state = 'appended' AND NEW.state IN ('backlinked', 'reconciling'))
       OR (OLD.state = 'failed' AND NEW.state IN ('pending', 'poison'))
       OR (OLD.state = 'reconciling' AND NEW.state IN ('backlinked', 'reconciled')) THEN
        RETURN NEW;
    END IF;
    RAISE EXCEPTION 'poc_outbox: illegal transition % -> %', OLD.state, NEW.state
        USING ERRCODE = 'P0001';
END;
$fn$;
ALTER FUNCTION poc_outbox_guard() OWNER TO migration_admin;
REVOKE ALL ON FUNCTION poc_outbox_guard() FROM PUBLIC;
DROP TRIGGER IF EXISTS poc_outbox_guard ON public.poc_outbox;
CREATE TRIGGER poc_outbox_guard
    BEFORE INSERT OR UPDATE OR DELETE ON public.poc_outbox
    FOR EACH ROW EXECUTE FUNCTION poc_outbox_guard();

-- Backlink identity: one row per EVENT, not per artifact
-- (corrections append; nothing is silently dropped by IM-11).
-- Idempotent: only fires while the 2-column PK still stands.
DO $do$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_constraint c
                JOIN pg_class t ON t.oid = c.conrelid
               WHERE c.conname = 'poc_ledger_backlink_pkey'
                 AND t.relname = 'poc_ledger_backlink'
                 AND array_length(c.conkey, 1) = 2) THEN
        ALTER TABLE public.poc_ledger_backlink
            DROP CONSTRAINT poc_ledger_backlink_pkey,
            ADD PRIMARY KEY (pg_table, pg_id, event_uuid);
    END IF;
END;
$do$;

-- Authorized enqueue (the ONLY producer; closes IM-13 for the
-- finalized-payroll path). Derives the evidence payload from
-- PERSISTED artifacts: the run must be validated, the tip valid,
-- results persisted. Same op_uuid twice with identical payload =
-- idempotent replay ('replayed'); same op with DIFFERENT payload =
-- loud squat rejection (never silent DO NOTHING).
CREATE OR REPLACE FUNCTION poc_outbox_enqueue(
    p_employer uuid, p_run uuid, p_event_uuid text, p_actor text)
RETURNS text LANGUAGE plpgsql
    SECURITY DEFINER SET search_path = pg_catalog AS $fn$
DECLARE
    v_op uuid; v_status text; v_tip uuid; v_tiphash text;
    v_out text; v_payload jsonb; v_hash bytea; v_old_hash bytea;
BEGIN
    IF p_actor IS NULL OR length(trim(p_actor)) = 0 THEN
        RAISE EXCEPTION 'poc_outbox_enqueue: actor attribution required';
    END IF;
    BEGIN
        v_op := p_event_uuid::uuid;
    EXCEPTION WHEN invalid_text_representation THEN
        RAISE EXCEPTION 'poc_outbox_enqueue: event identity must be a uuid';
    END;
    SELECT r.status INTO v_status FROM public.payroll_run r
     WHERE r.id = p_run AND r.employer_id = p_employer;
    IF NOT FOUND THEN
        RAISE EXCEPTION 'poc_outbox_enqueue: run unknown for employer';
    END IF;
    IF v_status <> 'validated' THEN
        RAISE EXCEPTION 'poc_outbox_enqueue: run is %, want validated', v_status
            USING ERRCODE = 'P0001';
    END IF;
    SELECT s.id, s.content_hash INTO v_tip, v_tiphash
      FROM public.payroll_snapshot s WHERE s.payroll_run_id = p_run AND s.status = 'valid';
    IF NOT FOUND THEN
        RAISE EXCEPTION 'poc_outbox_enqueue: no valid snapshot tip' USING ERRCODE = 'P0001';
    END IF;
    SELECT o.output_digest INTO v_out FROM public.poc_run_output o WHERE o.run_id = p_run;
    IF NOT FOUND THEN
        RAISE EXCEPTION 'poc_outbox_enqueue: no persisted output digest' USING ERRCODE = 'P0001';
    END IF;
    -- Canonical evidence payload, derived (never caller-supplied).
    SELECT pg_catalog.jsonb_build_object(
               'pg_table', 'payroll_run',
               'pg_id', p_run::text,
               'event_uuid', p_event_uuid,
               'employer_id', p_employer::text,
               'input_snapshot_digest', v_tiphash,
               'output_digest', v_out,
               'rule_pins', (SELECT COALESCE(pg_catalog.jsonb_agg(
                                  pg_catalog.jsonb_build_object('pack', p.pack, 'version', p.version,
                                      'digest', p.manifest_digest)
                                  ORDER BY p.effective_from, p.pack), '[]'::jsonb)
                              FROM public.payroll_snapshot_pack p WHERE p.snapshot_id = v_tip),
               'calc_version', r.calc_version,
               'approval', (SELECT pg_catalog.jsonb_build_object('approver', a.approver,
                                   'decided_at', a.decided_at::text)
                              FROM public.poc_approval a
                             WHERE a.run_id = p_run AND a.decision = 'approve'),
               'supersedes', COALESCE((
                   SELECT c.op_uuid::text FROM public.poc_outbox c
                    WHERE c.employer_id = p_employer AND c.action = 'payroll.finalized'
                      AND c.payload->>'pg_id' = p_run::text AND c.state IN ('backlinked', 'reconciled')
                    ORDER BY c.created_at DESC LIMIT 1), '')
           ) INTO v_payload
      FROM public.payroll_run r WHERE r.id = p_run;
    v_hash := public.digest(v_payload::text, 'sha256');
    SELECT c.payload_hash INTO v_old_hash FROM public.poc_outbox c WHERE c.op_uuid = v_op;
    IF FOUND THEN
        IF v_old_hash IS DISTINCT FROM v_hash THEN
            RAISE EXCEPTION 'poc_outbox_enqueue: op % squats a different payload', v_op
                USING ERRCODE = 'P0001';
        END IF;
        RETURN 'replayed';
    END IF;
    INSERT INTO public.poc_outbox (op_uuid, employer_id, action, payload_hash, payload)
    VALUES (v_op, p_employer, 'payroll.finalized', v_hash, v_payload);
    RETURN 'enqueued';
END;
$fn$;
ALTER FUNCTION poc_outbox_enqueue(uuid, uuid, text, text) OWNER TO migration_admin;
REVOKE ALL ON FUNCTION poc_outbox_enqueue(uuid, uuid, text, text) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION poc_outbox_enqueue(uuid, uuid, text, text) TO app;

-- Authorized backlink record (the ONLY backlink writer; closes the
-- PG-5 squat). The outbox row must pre-exist in an active state with
-- a MATCHING anchored tx; exact replays converge ('replayed'),
-- conflicting events fail loudly (never ON CONFLICT DO NOTHING).
CREATE OR REPLACE FUNCTION poc_backlink_record(
    p_pg_table text, p_pg_id text, p_event_uuid text, p_tx bigint, p_alh bytea)
RETURNS text LANGUAGE plpgsql
    SECURITY DEFINER SET search_path = pg_catalog AS $fn$
DECLARE
    v_state text; v_tx bigint;
    v_old_tx bigint; v_old_alh bytea;
BEGIN
    SELECT c.state, c.immu_tx INTO v_state, v_tx
      FROM public.poc_outbox c WHERE c.op_uuid = p_event_uuid::uuid;
    IF NOT FOUND THEN
        RAISE EXCEPTION 'poc_backlink_record: no outbox row for event' USING ERRCODE = 'P0001';
    END IF;
    IF v_state NOT IN ('appended', 'reconciling') THEN
        RAISE EXCEPTION 'poc_backlink_record: outbox state % cannot link', v_state
            USING ERRCODE = 'P0001';
    END IF;
    IF v_tx IS DISTINCT FROM p_tx THEN
        RAISE EXCEPTION 'poc_backlink_record: tx % does not match anchored %', p_tx, v_tx
            USING ERRCODE = 'P0001';
    END IF;
    SELECT b.immu_tx, b.alh INTO v_old_tx, v_old_alh
      FROM public.poc_ledger_backlink b
     WHERE b.pg_table = p_pg_table AND b.pg_id = p_pg_id AND b.event_uuid = p_event_uuid;
    IF FOUND THEN
        IF v_old_tx IS DISTINCT FROM p_tx OR v_old_alh IS DISTINCT FROM p_alh THEN
            RAISE EXCEPTION 'poc_backlink_record: conflicting event for artifact'
                USING ERRCODE = 'P0001';
        END IF;
        RETURN 'replayed';
    END IF;
    INSERT INTO public.poc_ledger_backlink (pg_table, pg_id, event_uuid, immu_tx, alh)
    VALUES (p_pg_table, p_pg_id, p_event_uuid, p_tx, p_alh);
    RETURN 'linked';
END;
$fn$;
ALTER FUNCTION poc_backlink_record(text, text, text, bigint, bytea) OWNER TO migration_admin;
REVOKE ALL ON FUNCTION poc_backlink_record(text, text, text, bigint, bytea) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION poc_backlink_record(text, text, text, bigint, bytea) TO app;

-- ============================================================ Part 9 ==
-- Loud posture assertion: the apply FAILS unless every authority
-- boundary below holds exactly. Re-run safe (pure checks).
DO $poc_auth$
DECLARE
    t text; fn text;
    v_grants text; v_owner text; v_secdef bool; v_path text[];
    v_def text; v_states text; v_keylen int;
BEGIN
    -- App reads but never writes authoritative tables.
    FOREACH t IN ARRAY ARRAY[
        'payroll_snapshot', 'payroll_snapshot_employee', 'payroll_snapshot_pack',
        'payroll_run', 'payroll_line', 'payslip',
        'poc_claim', 'poc_period_close', 'poc_ledger_backlink',
        'poc_rule_pack_schedule', 'poc_run_output', 'poc_approval'
    ] LOOP
        SELECT string_agg(privilege_type, ',' ORDER BY 1) INTO v_grants
          FROM information_schema.role_table_grants
         WHERE table_schema = 'public' AND table_name = t AND grantee = 'app';
        IF v_grants IS DISTINCT FROM 'SELECT' THEN
            RAISE EXCEPTION 'poc_auth: app holds % on %, want SELECT only', v_grants, t;
        END IF;
    END LOOP;
    -- Outbox: worker keeps SELECT + guarded UPDATE only.
    SELECT string_agg(privilege_type, ',' ORDER BY 1) INTO v_grants
      FROM information_schema.role_table_grants
     WHERE table_schema = 'public' AND table_name = 'poc_outbox' AND grantee = 'app';
    IF v_grants IS DISTINCT FROM 'SELECT,UPDATE' THEN
        RAISE EXCEPTION 'poc_auth: app holds % on poc_outbox, want SELECT,UPDATE', v_grants;
    END IF;

    -- DEFINER shape: owner + security + search_path + qualified refs.
    FOREACH fn IN ARRAY ARRAY[
        'poc_src_hash', 'poc_pack_claim_hash', 'poc_verify_claims',
        'poc_run_consistency',
        'poc_snapshot_capture', 'poc_result_persist', 'poc_approve',
        'poc_run_validate', 'poc_close_period', 'poc_writer_period_code',
        'poc_require_gate', 'poc_outbox_guard',
        'poc_outbox_enqueue', 'poc_backlink_record'
    ] LOOP
        SELECT u.rolname, p.prosecdef, p.proconfig, pg_get_functiondef(p.oid)
          INTO v_owner, v_secdef, v_path, v_def
          FROM pg_proc p JOIN pg_roles u ON u.oid = p.proowner
         WHERE p.proname = fn;
        IF v_owner IS DISTINCT FROM 'migration_admin' THEN
            RAISE EXCEPTION 'poc_auth: % owned by %, want migration_admin', fn, v_owner;
        END IF;
        IF NOT v_secdef THEN
            RAISE EXCEPTION 'poc_auth: % is not SECURITY DEFINER', fn;
        END IF;
        IF v_path IS DISTINCT FROM ARRAY['search_path=pg_catalog'] THEN
            RAISE EXCEPTION 'poc_auth: % search_path is %, want pg_catalog only', fn, v_path;
        END IF;
        IF v_def LIKE '%FROM poc\_%' ESCAPE '\'
           OR v_def LIKE '%UPDATE poc\_%' ESCAPE '\'
           OR v_def LIKE '%PERFORM poc\_%' ESCAPE '\'
           OR v_def LIKE '%FROM payroll\_%' ESCAPE '\'
           OR v_def LIKE '%INTO payroll\_%' ESCAPE '\' THEN
            RAISE EXCEPTION 'poc_auth: % keeps an unqualified reference', fn;
        END IF;
    END LOOP;

    -- EXECUTE posture: app reaches the authorized operations, and
    -- nothing else in this file's surface.
    FOREACH fn IN ARRAY ARRAY[
        'poc_snapshot_capture', 'poc_result_persist', 'poc_approve',
        'poc_run_validate', 'poc_close_period',
        'poc_outbox_enqueue', 'poc_backlink_record',
        'poc_run_consistency'
    ] LOOP
        IF NOT pg_catalog.has_function_privilege('app', 'public.' || fn ||
            CASE fn WHEN 'poc_snapshot_capture' THEN '(uuid, text, text)'
                    WHEN 'poc_result_persist' THEN '(uuid, text, text, jsonb, jsonb, text)'
                    WHEN 'poc_approve' THEN '(uuid, text, text, text)'
                    WHEN 'poc_run_validate' THEN '(uuid, text)'
                    WHEN 'poc_close_period' THEN '(uuid, text, uuid)'
                    WHEN 'poc_outbox_enqueue' THEN '(uuid, uuid, text, text)'
                    WHEN 'poc_run_consistency' THEN '(uuid)'
                    ELSE '(text, text, text, bigint, bytea)' END, 'EXECUTE') THEN
            RAISE EXCEPTION 'poc_auth: app cannot EXECUTE %', fn;
        END IF;
    END LOOP;
    -- Attempt registry: app INSERT (observability) but never UPDATE/DELETE.
    SELECT string_agg(privilege_type, ',' ORDER BY 1) INTO v_grants
      FROM information_schema.role_table_grants
     WHERE table_schema = 'public' AND table_name = 'poc_workflow_attempt' AND grantee = 'app';
    IF v_grants IS DISTINCT FROM 'INSERT,SELECT' THEN
        RAISE EXCEPTION 'poc_auth: app holds % on poc_workflow_attempt, want INSERT,SELECT', v_grants;
    END IF;
    IF pg_catalog.has_function_privilege('app', 'public.poc_verify_claims(uuid, uuid)', 'EXECUTE') THEN
        RAISE EXCEPTION 'poc_auth: app keeps EXECUTE on poc_verify_claims';
    END IF;
    IF pg_catalog.has_function_privilege('app', 'public.poc_gate_acquire(uuid)', 'EXECUTE') THEN
        RAISE EXCEPTION 'poc_auth: app keeps direct EXECUTE on poc_gate_acquire';
    END IF;

    -- Gate trigger coverage (15 tables, INSERT+UPDATE+DELETE).
    FOREACH t IN ARRAY ARRAY[
        'attendance_interval', 'clock_event', 'clock_event_disposition',
        'compensation_line', 'employee', 'employee_status_history',
        'employment_terms', 'rule_pack_activation', 'shg_instruction',
        'timesheet', 'timesheet_line', 'timesheet_line_interval',
        'payroll_run', 'pay_period', 'payroll_run_exclusion'
    ] LOOP
        IF NOT EXISTS (SELECT 1 FROM pg_trigger g JOIN pg_class c ON c.oid = g.tgrelid
                        WHERE g.tgname = 'poc_gate_participate' AND c.relname = t
                          AND g.tgenabled = 'O') THEN
            RAISE EXCEPTION 'poc_auth: gate trigger missing on %', t;
        END IF;
    END LOOP;

    -- Outbox: exactly one state check, all seven states.
    SELECT string_agg(pg_get_constraintdef(o.oid), ';') INTO v_states
      FROM pg_constraint o JOIN pg_class c ON c.oid = o.conrelid
     WHERE c.relname = 'poc_outbox' AND o.contype = 'c'
       AND pg_get_constraintdef(o.oid) LIKE '%state%';
    IF v_states IS NULL
       OR v_states NOT LIKE '%pending%' OR v_states NOT LIKE '%appended%'
       OR v_states NOT LIKE '%backlinked%' OR v_states NOT LIKE '%failed%'
       OR v_states NOT LIKE '%reconciling%' OR v_states NOT LIKE '%reconciled%'
       OR v_states NOT LIKE '%poison%' OR v_states LIKE '%;%' THEN
        RAISE EXCEPTION 'poc_auth: outbox state check wrong: %', v_states;
    END IF;

    -- Widened keys: pack pins and backlinks keyed per version/event.
    SELECT array_length(c.conkey, 1) INTO v_keylen FROM pg_constraint c
      JOIN pg_class t ON t.oid = c.conrelid
     WHERE c.conname = 'payroll_snapshot_pack_pkey';
    IF v_keylen IS DISTINCT FROM 3 THEN
        RAISE EXCEPTION 'poc_auth: pack PK has % columns, want 3', v_keylen;
    END IF;
    SELECT array_length(c.conkey, 1) INTO v_keylen FROM pg_constraint c
      JOIN pg_class t ON t.oid = c.conrelid
     WHERE c.conname = 'poc_ledger_backlink_pkey';
    IF v_keylen IS DISTINCT FROM 3 THEN
        RAISE EXCEPTION 'poc_auth: backlink PK has % columns, want 3', v_keylen;
    END IF;

    -- Schedule seeded for the POC employer (overlay applies post-seed).
    IF NOT EXISTS (SELECT 1 FROM public.poc_rule_pack_schedule s
                    JOIN public.employer e ON e.id = s.employer_id
                   WHERE e.code = 'POC-MERIDIAN') THEN
        RAISE EXCEPTION 'poc_auth: pack schedule has no POC-MERIDIAN rows (apply after seed-poc)';
    END IF;
    -- One-draft serialization index present.
    IF NOT EXISTS (SELECT 1 FROM pg_class c JOIN pg_namespace n ON n.oid = c.relnamespace
                    WHERE c.relname = 'poc_one_draft_run' AND n.nspname = 'public') THEN
        RAISE EXCEPTION 'poc_auth: poc_one_draft_run index missing';
    END IF;
END;
$poc_auth$;

