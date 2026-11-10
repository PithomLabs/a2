-- POC-scoped transactional outbox (disposable databases only).
-- Never applied to shared/staging/production databases.
-- One row per PG->immudb append intent; op_uuid is the idempotency key shared
-- with the immudb idem/... precondition key.
CREATE TABLE IF NOT EXISTS poc_outbox (
    op_uuid      uuid        PRIMARY KEY,
    employer_id  uuid        NOT NULL,
    action       text        NOT NULL,
    payload_hash bytea       NOT NULL,
    state        text        NOT NULL DEFAULT 'pending'
                 CHECK (state IN ('pending', 'appended', 'backlinked', 'failed',
                                  'reconciling', 'reconciled', 'poison')),
    immu_tx      bigint      NULL,
    attempts     integer     NOT NULL DEFAULT 0,
    created_at   timestamptz NOT NULL DEFAULT now(),
    updated_at   timestamptz NOT NULL DEFAULT now()
);
-- State meanings (prompt8 §5.6; transitions enforced by the
-- poc_outbox_guard trigger in gate/pg_authority.sql, which also
-- re-asserts this CHECK idempotently for older stacks):
-- pending (no proven append) -> appended (tx anchored) -> backlinked;
-- backlink failures route appended -> reconciling -> backlinked or
-- reconciled (stays actionable; never discarded by retry accounting);
-- pending -> failed (append never proven, operator may reset to
-- pending) or poison (terminal, append never proven).
-- Payload carries everything the worker needs to build the ledger Record.
-- Idempotent for stacks created before this column existed.
ALTER TABLE poc_outbox ADD COLUMN IF NOT EXISTS payload JSONB NOT NULL DEFAULT '{}';
CREATE INDEX IF NOT EXISTS poc_outbox_pending_idx
    ON poc_outbox (state, created_at) WHERE state = 'pending';
-- Grants (worker runs as app; state machine needs UPDATE, history needs no
-- DELETE). Self-healing vs default privileges (app=arwd on new tables).
REVOKE ALL ON poc_outbox FROM PUBLIC;
REVOKE ALL ON poc_outbox FROM app;
GRANT SELECT, INSERT, UPDATE ON poc_outbox TO app;

-- Append-only ledger back-links (ADR-0204, Gate A): one row per
-- EVIDENCE EVENT (prompt8 §5.5: corrections append new events for the
-- same artifact; the key is (pg_table, pg_id, event_uuid)), inserted
-- only after VerifiedTxByID via poc_backlink_record. No UPDATE/DELETE
-- path by trigger + revoke: provenance-bearing rows are never edited
-- in place.
CREATE TABLE IF NOT EXISTS poc_ledger_backlink (
    pg_table    text        NOT NULL,
    pg_id       text        NOT NULL,
    event_uuid  text        NOT NULL,
    immu_tx     bigint      NOT NULL,
    alh         bytea       NOT NULL,
    verified_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (pg_table, pg_id, event_uuid)
);
CREATE OR REPLACE FUNCTION poc_backlink_no_rewrite()
RETURNS trigger LANGUAGE plpgsql AS $fn$
BEGIN
    RAISE EXCEPTION 'poc_ledger_backlink is append-only';
END;
$fn$;
DROP TRIGGER IF EXISTS poc_backlink_no_rewrite ON poc_ledger_backlink;
CREATE TRIGGER poc_backlink_no_rewrite
    BEFORE UPDATE OR DELETE ON poc_ledger_backlink
    FOR EACH ROW EXECUTE FUNCTION poc_backlink_no_rewrite();
REVOKE UPDATE, DELETE ON poc_ledger_backlink FROM app;
GRANT SELECT, INSERT ON poc_ledger_backlink TO app;
