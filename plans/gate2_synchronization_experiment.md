# Gate 2 Report — Payroll Snapshot Synchronization Experiment

Prompt: `plans/prompt18.md`. Scope: architecture experiment only. Production
migration `00013` was NOT created, and neither mechanism was wired into
capture. All PostgreSQL claims below were observed live on 16.15 unless
labeled inference.

Evidence key used throughout: **[O]** observed database result,
**[T]** executed test, **[C]** code inspected, **[I]** agent inference,
**[U]** untested assumption.

## 1. Executive decision

**GO for Option B** (READ COMMITTED + transaction-scoped employer gate +
database-enforced writer participation + own-row phase marker) as the
production synchronization architecture. All 10 GO criteria pass with
live evidence (79 green checks). **NO-GO for retaining Option A**: the
post-snapshot release/reacquisition attack is demonstrably open — the
XID ticket proves ordering only, and uninterrupted barrier ownership
fails on both the plain and revoked databases, including a
never-restorable state under the Gate-1 revoke. Prompt17's pre-BEGIN
RR/session-lock protocol should be superseded (see ADR).

## 2. Exact invariant and capture-boundary semantics

**Invariant:** a committed snapshot+validation reflects every
payroll-affecting input committed before the decision commit, and no
input changes between the decision reads and the decision commit — over
ALL writer paths (Go/CLI/REST/trigger/direct SQL).

**Option B boundary:** one gated RC transaction (preferred) in which the
employer gate is acquired before authoritative reads; the database
re-verifies every Go-supplied claim against post-gate reads; snapshot
persist, proof verification, validation audit, and `draft → validated`
commit atomically. Split path: validation re-acquires the same gate in
a new transaction, plants an attempt marker, rebuilds claims from
frozen snapshot rows, and decides atomically — never assuming the
capture gate survived.

## 3. Repository and database baseline

- Root `/home/chaschel/Documents/go/a1`, branch `main`, HEAD `304fec8`;
  59 uncommitted entries preserved, none overwritten. **[O]**
- Migrations `00001`–`00011` sha256 OK; `00012` `c05fc827…41b7d`
  unchanged; no `00013` before or after. **[O]**
- PostgreSQL 16.15 (Ubuntu), scratch loopback; roles `postgres` (super),
  `migration_admin` + `app` (non-super, login, no memberships, app owns
  0 objects). **[O]**
- Gate-1 mechanism suite green (regression still green at close); 12
  attack tests red before and after; `TestOT001RemainsBlocked` green;
  `admission.go` unwired (no non-test importers); no Gate-1-package
  discrepancy found. **[T]**
- Capture/validate are separate RR transactions today
  (`snapshot.go:246`, `validate.go:95`). **[C]**

## 4. Complete payroll-input writer inventory

63 public tables, all `migration_admin`-owned. Snapshot read-set (18
tables): `attendance_interval`, `clock_event`, `clock_event_disposition`,
`compensation_line`, `employee`, `employee_status_history`,
`employment_terms`, `pay_period`, `payroll_run`, `payroll_run_exclusion`,
`payroll_snapshot(+_employee,+_pack)`, `rule_pack_activation(+_history)`,
`shg_instruction`, `timesheet(+_line,+_line_interval)`. **[C]**

Privilege posture **[O]**: app has full DML on employer/period/
timesheets/runs; INSERT-only on attendance/clock/employee/terms/
snapshots/audit; SELECT-only on comp/SHG/pack-activation (DEFINER
writers); NO grants at all on `rule_pack`; no TRUNCATE, TEMP, ownership,
or memberships anywhere. No MERGE/COPY/Batch in non-test Go; no
partitioned tables; no cascading FKs (3 RESTRICT only); upserts are
`DO NOTHING` except one reviewed day-record `DO UPDATE`. All triggers
are ROW-level; DEFINER writers nest the lock helpers. Full matrix with
per-table identity sources, triggers, and paths:
`/tmp/gate2-writers.txt` (working note; key rows above).

Anomalies recorded (not Gate-2 blockers): `employee_status_history`
has no trigger (direct app INSERT possible); Go UPDATE callers exist
for `attendance_exception` without an UPDATE grant; `rule_pack` has no
app grants at all. **[C/O]**

## 5. Option A attack sequence, SQL, result, persisted-state evidence

Harness: Gate-1 `admissionExpSQL` reused verbatim on disposable DBs
(`_gate2plain` unrevoked; `_gate2revoked` = revoke + DEFINER
redefinition, i.e. Option A as proposed for production). 7 tests, 13
checks, all green. **[T]**

**Control:** held barriers block a participating writer (visible waiter
row); release lets it through — the harness detects protection. **[T]**

**Test A (pre-BEGIN DISCARD ALL): VULNERABLE on both DBs.**
Barriers → ticket → `DISCARD ALL` drops physical locks (rows survive)
→ wrapper reacquire silently no-ops (row found, no lock taken) →
`pg_locks` proves zero locks with rows present → RR + snapshot →
**verification passes** → writer commits inside the window → snapshot
misses it → stale commits. Persisted: intruder committed strictly
before admission commit, absent from its read set. Under revoke the raw
restoration path is *denied*, so the backend can never restore its
barriers yet still verifies — strictly worse. Raw re-lock (unrevoked
only) restores protection (control branch). **[T]**

**Test B (post-snapshot release): re-verify fails closed; the work
window does not.** Enumerated as app: raw bigint unlock (succeeds
unrevoked; denied revoked), `unlock_all` (same), `(int,int)` spelling
(**distinct namespace** — releases nothing; fewer paths than feared),
`DISCARD ALL` in-txn (rejected by PostgreSQL itself), wrapper unlock
(granted both; deletes rows). After release + writer + no-op reacquire,
re-verification fails `already consumed` on both DBs. **Extension:**
release AFTER verification, during work → writer commits → stale read
set commits **undetected on both DBs**; cleanup notices too late
(unlock returns false — `Do` would discard after committing stale).
Rollback undoes the consume (by design); termination orphans hold rows
(time-GC'd) without blocking fresh backends. **[T]**

**§10 verdicts:** Q1 YES (unrevoked raw; revoked wrapper + always-available
pre-BEGIN DISCARD). Q2 ORDERING ONLY — barrier ownership is never
continuously proven (rows⟺locks correspondence breaks; re-lock no-ops;
no verify-to-commit re-check). Retention would require a
physical-presence probe + pre-commit re-check + non-idempotent repair;
none exists. **[I from T]**

## 6. Option B architecture and exact transaction protocol

One-txn path (proven): `BEGIN (RC)` → `gate2_acquire_employer_gate`
(RC check + `pg_advisory_xact_lock(hashtextextended('gate2-payroll:'||
employer, 0))`) → Go reads + resolves (advisory) → `gate2_capture`
(re-checks RC, re-acquires gate re-entrantly, locks run `FOR UPDATE`,
recomputes sha256 over *supplied* bytes, proof-checks every claim
against post-gate reads, persists snapshot + full-row children + pack
pins) → `gate2_validate` (gate, attempt marker, claims rebuilt from
frozen snapshot rows, verdict, run transition) → `COMMIT`. Split path:
same calls across two transactions. **[T]**

Why pre-gate reads are harmless: the verifier re-establishes every
claim from its own post-gate reads — stale submitted claims are
rejected for omission (proven). **[T]**

## 7. Canonical key derivation and trigger coverage

One function owns the key; capture, writers, and DEFINER nests share
one namespace (locks are not role-scoped, so INVOKER suffices
everywhere). Same-key proven behaviorally: waiters arriving via direct
write, DEFINER nest, and second capture wait on the exact
`(classid,objid)` the holder owns; two employers held at once carry
different keys with zero cross-blocking. 64-bit hash: one employer
never splits; collisions only over-serialize (accepted, negligible).
**[T]** One gate per txn + gate-before-row + sorted dual-lock on
employer-changing UPDATE ⇒ single-employer discipline is deadlock-free
by construction; opposite swaps serialize (proven). **[T]**

Coverage is fail-closed: BEFORE ROW I/U/D trigger on every input
table, asserted from the catalog; no CASCADE FKs; app holds no
TRUNCATE (denial tested). **[T]**

## 8. Same-transaction post-capture protection

Mechanism: writer triggers reject when the current txn holds an
uncommitted decision row (snapshot header or validation attempt) for
that employer. Membership test: row **visible in this snapshot AND
`txid_status(xmin) = 'in progress'`** — exactly own uncommitted rows,
top-level and subtransaction alike, no xid casts, no epoch games.
INSERT-only decision tables make the marker unplantable-by-others,
unforgable (planting = self-restriction), and undeletable by app
(denials tested); UPDATE-by-invalidate keeps the row, so the marker
persists fail-closed. **[T]**

Proven (8 subtests): prompt-exact sequence rejected with nothing
persisted; pre-marker writes allowed and self-consistently included;
UPDATE/DELETE/DEFINER-add/pack-activate rejected; attempt-marker
branch rejected; cross-employer writes allowed; savepoint
plant→block→rollback-to→allow→replant→block (instant-wise correct);
committed snapshots block nothing new. The tests caught two real holes
during development (DELETE skipped the marker; subtxn-planted markers
dodged top-xid comparison) — both fixed and green. **[T]**

Rejected alternatives: xid-keyed table (retention/recycling), pg_locks
probe (key math), temp table (privilege uncertainty), GUC/Go flags
(forgable). **[I]**

## 9. Direct-SQL privilege tests

As app: INSERT/UPDATE/DELETE/multi-row/upsert/COPY/MERGE all participate
(wait-then-proceed per class, §10); TRUNCATE, DISABLE/DROP TRIGGER,
OWNER change, `SET ROLE`, replica-mode GUC, CREATE TABLE, and direct
restricted-table writes all denied with expected errors; triggers
proven still enforced after the failed replica SET. **[T]**

## 10. Writer-first and capture-first results

Full matrices for all 9 writer classes, both orderings, with frozen
fingerprints: capture-first commits exclude the not-yet-committed
write (`seed|P:v1` frozen); writer-first commits include it. Plus
detailed inclusion/exclusion counts for the representative path
(children 3 incl. early row; 2 + late row committed-after and
excluded). Concurrent same-employer captures serialize and both settle
with distinct sequence versions; different employers never block.
**[T]**

## 11. Snapshot authority and negative-test results

ADR-1 preserved: Go resolver + canonical bytes + sha256 stay the single
implementation; SQL recomputes digests over supplied bytes and
proof-checks claims (set-equality both directions, per-field equality,
lineage scope, pack pins) — never re-solves, never generates canonical
form. 17 negatives green: omitted/fabricated input, wrong
employee/payload/source/version, broken + cross-employer lineage,
wrong/omitted/fabricated pack, wrong employer, wrong hash, empty
claims, **forged hash over consistent-but-fabricated bytes (rejected)**,
planted snapshot cannot validate, validated content immutable (8
UPDATE/DELETE denials + bytes re-read). A read of the design caught and
fixed a tautological validate (rebuilding claims from live tables);
frozen children + the split-stale tests now guard it. **[T]**

## 12. Rollback, timeout, cancellation, connection-failure results

Capture failure → zero snapshot/child/attempt/verdict rows, gate
released, marker gone (fresh capture + writer proceed); writer failure
→ row absent; cancel → waiter released, holder commits, canceled row
absent (discard-on-cancel: pgx breaks the canceled conn — pools must
not reuse); statement/lock timeouts typed (`57014`/`55P03`) with usable
conns; termination releases the gate with nothing orphaned; crafted
multi-employer deadlock yields exactly one typed `40P01` with the
loser's writes fully rolled back and the winner intact. **[T]**

## 13. Known limitations and residual risks

1. Multi-employer write txns can deadlock (typed `40P01`, clean);
   single-employer discipline is deadlock-free. Production rule needed:
   one employer per write txn, or sorted pre-acquisition via a Go
   helper. **[T observed; rule I]**
2. `txid_status` runs per written row; production performance
   unmeasured. **[U]**
3. Fixture⊂production: 3 input tables + 1 pack model 18 read-set
   tables; per-table employer derivation, DEFINER-nest coverage, and
   the run-lifecycle mapping must be re-proven per table in the
   production battery. **[U]**
4. New input tables need the trigger + coverage-list entry (fail-closed
   pattern graduates). **[I]**
5. Pooler-safety (txn-scoped state ⇒ transaction-pooling compatible,
   unlike Option A) follows from mechanism properties but no pooler was
   tested. **[I]**

## 14. Option A versus Option B comparison

| | A (XID-ticket + session barriers) | B (RC + tx gate) |
|---|---|---|
| Freshness proof | Ticket ordering only; ownership gaps demonstrated (§5) | Gate serialization + re-verification; both orderings × 9 classes |
| Same-txn protection | None between verify and commit (stale commits) | Marker rejects; 8/8 subtests |
| Bypass surface | DISCARD ALL always; raw unlock unrevoked | All 9 paths gated; DDL/privilege denials |
| Privilege cost | 21-revoke + 3 DEFINER redefinitions + wrapper | No revoke; 1 INVOKER fn + triggers; no new role |
| Connection discipline | Dedicated session-pinned conns + discard state machine | Plain txns; pooler-safe by construction [I] |
| Mechanism size | ~350-line Go bracket + 2 tables + 6 fns + revoke | 1 gate fn + 1 trigger fn + N triggers + capture/validate; no Go bracket |
| Failure residuals | Orphan rows (time-GC); unrestorable-after-DISCARD under revoke | 40P01 only for out-of-MVP multi-employer txns |

Parent-row gate (narrow comparison, analysis only — building both was
out of scope): equivalent serialization but still needs the same
writer triggers to force the lock, couples to the employer row's
existence (bootstrap gap advisory keys avoid), and offers no
monitoring or privilege advantage. No material benefit; advisory
retained. **[I]**

## 15. Final GO/NO-GO decision

**GO for Option B for production implementation** (via a future
production migration — this experiment does not authorize writing it).
All 10 GO criteria pass: RC rejection, gate-before-reads, DB-enforced
writer participation, direct-SQL non-bypass, both orderings × 9
classes, one protected decision (both paths), same-txn rejection,
authority negatives, clean rollback/failure behavior, intact
historical/statutory boundaries (migrations byte-identical, OT-001
BLOCKED, succession immutability). **NO-GO for retaining Option A**
(post-snapshot release demonstrably open; retention bar unmet).

## 16. Exact proposed production migration scope if GO

Single forward migration (number TBD at implementation; NOT authorized
now): canonical gate function (INVOKER) + writer-participation trigger
function + BEFORE ROW I/U/D triggers on every §4 input table +
validation-attempt table + capture/validate DEFINER boundary +
run-lifecycle fit + per-table key-derivation + catalog coverage guards
+ privilege pins (decision tables append-only; no app DDL). Down:
remove triggers/functions/attempt table in reverse order; base schema
untouched. Production battery must re-prove: per-table derivation and
nest coverage, full ordering matrices on real tables, activation
live-fire, 40P01 retry posture at callers, and performance of the
per-row checks. **[I proposal]**

## 17. Explicitly deferred work

Production migration authoring; per-table production trigger rollout;
Go writer gate-first adoption beyond the proven representative;
multi-employer txn rule + sorted pre-acquire helper; timeout/deadlock
retry policy; `txid_status` performance measurement; pooler
compatibility test; parent-row gate build-out (not recommended);
event sourcing (longer-term data-model note, never `MAX(event_id)` for
order); SERIALIZABLE evaluation (only on explicit business acceptance
of weaker snapshots); the 12 red attack tests turning green against
the production capture implementation.
