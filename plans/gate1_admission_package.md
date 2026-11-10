# Gate 1 Package — M2.1 D1 Admission Boundary

Unit: prompt17 as amended by plan_prompt17.1 (review6 A–K + review6.1 A–G),
steps 1–6 only. Steps 7–9 (barrier matrix, capture/validation, comp, SHG,
packs, battery, report) are explicitly not authorized until Gate 1 approval.

Evidence base: `core/database/admission.go` (experimental mechanism, NOT
wired into capture), `core/database/admission_test.go` (31 green mechanism
proofs), `internal/payroll/snapshot_admission_test.go` (12 red attack
proofs, still failing as required). All runs: PG 16.15 scratch, loopback,
`GOWORK=off`, mock LLM default.

## Gate 1 decision status (replaces "Open questions: None" per §G)

- P1a expected-population semantics: DECIDED (plan_prompt17.1, pinned).
- ADR-1 snapshot authority (Option 2, Go-canonical + verifying writer):
  DECIDED (plan_prompt17.1, 8 criteria).
- D1 admission mechanism: PROVEN, DECIDED pending privilege approval —
  XID-ticket + session barriers + full 21-overload revoke + DEFINER
  wrappers. See "Mechanism decision" and "Privilege-change request".
- Capture/validation implementation: NOT AUTHORIZED until Gate 1 passes.

## §F live evidence (7 items)

1. **Original stale-RR attack.** RED (vulnerability live):
   `TestStaleSnapshotAttack` FAILS — current capture admits a snapshot
   whose RR transaction predates a committed writer. GREEN (mechanism
   defeats it): `TestTicketAfterSnapshotRejected` — a ticket committed
   after the snapshot began is invisible inside it and rejected;
   `TestTicketBeforeSnapshotAccepted` — the legitimate ordering verifies.
2. **Ticket visibility under the actual RR snapshot.** GREEN:
   `TestTicketAbortedRejected` (aborted-issue invisible),
   `TestTicketNestedIssueRejected` (RR-issued vacuous proof rejected),
   `TestTicketExpiredRejected` (5-minute expiry),
   `TestTicketReleasedRejected`, `TestTicketSingleUseAcrossCommits`,
   `TestTicketReusableAfterRollback` (accepted by design: barriers held).
3. **False rejection under an unrelated long-running transaction.**
   GREEN: `TestTicketOldXminAccepted` — an uncommitted writer pinning
   global xmin does NOT false-reject a valid ticket (per-XID visibility,
   never xmin comparison). The holder-xmin is asserted non-null, so the
   pressure is proven, not assumed.
4. **Lock substitution/release attack.** RED: `TestLockSubstitutionUndetectable`
   FAILS — a session→xact lock swap on the same key is undetectable in
   current code. GREEN (mechanism closes it):
   `TestAdvisoryRevokeInventory` — all 21 overloads denied to app/PUBLIC,
   granted to migration_admin only, with live denial in every overload
   family; `TestSessionXactKeyCompatibility` — wrappers and helpers share
   one namespace (mutual exclusion proven behaviorally);
   `TestWrapperLockIdempotent`; `TestRevokeBootstrapGap` pins the one
   remaining direct app use (employer bootstrap → 00013 wrapper).
5. **Legitimate concurrent capture.** GREEN:
   `TestAdmissionConcurrentLegitimate` — same-employer ×2 plus
   different-employer ×1 all succeed with one attempt each: barriers
   serialize, no false-reject livelock. `TestAdmissionHappy` proves the
   full 8-step ordering end to end with physical lock release asserted.
6. **Complete population and lock-set verification.** GREEN:
   `TestTicketRequiresBarriers` (issue requires full hold set),
   `TestVerifyRequiresHoldsPresent` (use-after-unlock rejected),
   `TestAdmissionMismatchRetriesThenSucceeds` (one divergence → bounded
   retry → success), `TestAdmissionMismatchExhausts` (persistent
   divergence → `MismatchError` after 3 attempts, cleanup still runs).
   All 8 cleanup cases green: success, work error, work SQL error,
   cancellation → discard, deadline → discard, panic → propagate after
   cleanup, unlock failure → discard despite work success (with
   disconnect-proven physical release), clean admission → pool reuse.
   RED (awaiting capture): the 9 writer-first inclusion tests
   (`TestWriterFirst*`, `TestValidationWriterFirst`) FAIL — a writer
   committing before capture is invisible to the uncaptured snapshot.
7. **Snapshot-authority direct-SQL fabrication attack.** RED:
   `TestSnapshotFabricationByApp` FAILS — an invented `valid` snapshot is
   readable today. GREEN (mechanism binds identity): `TestTicketForeignBackendRejected`
   (pid binding), employer binding, consumed/released binding. Full
   closure needs the post-Gate-1 verifying writer (ADR-1); the ticket
   supplies its freshness binding.

## §G report corrections (6 items)

- **Why `pg_locks` alone is insufficient:** granted session and
  transaction locks are byte-identical in `pg_locks` on PG16.15 (live
  proof: session/xact rows identical, `transactionid` NULL). Presence is
  observable; scope and order are not. The mechanism uses `pg_locks` for
  nothing; ordering comes from XID visibility.
- **Why global xmin comparison may false-reject:** an unrelated open
  transaction pins xmin below any ticket XID. The mechanism never
  compares xmins: row visibility plus `pg_visible_in_snapshot(ticket_xid,
  pg_current_snapshot())` decide per-XID.
- **Ticket visibility mechanism:** the ticket row commits in a
  pre-snapshot RC operation; `SELECT` inside the RR snapshot finds it
  iff its XID predates the snapshot (aborted/after → absent → rejected).
  Belt-and-braces `pg_visible_in_snapshot` on the stored `ticket_xid`.
  Top-level XID, RC-issued only, pid+employer bound, 5-minute expiry,
  single-use across commits. Subtransactions, prepared xacts, and
  wraparound are out of scope for short-lived tickets (stated limit).
- **Raw advisory-function privilege boundary:** all 21 overloads revoked
  from PUBLIC, granted to migration_admin only; app reaches advisory
  solely through DEFINER wrappers. Every legitimate path uses identical
  key derivation (proven behaviorally; 00013 introduces one canonical
  key function to end the duplication).
- **Canonical resolver and hash authority:** ADR-1 Option 2 — Go resolver
  stays canonical (366 lines, carve logic unported); DB proof-checks
  claims and derives decisions; Go keeps the single canonical byte form
  and SHA-256; SQL never generates canonical bytes.
- **Privilege-change request:** see below. Approval required before 00013.

## Mechanism decision

**Selected: XID-ticket + session barriers + full revoke (boundary A).**
One dedicated connection per admission: population session barrier →
post-barrier enumerate (RC) → employee/pack barriers in canonical order →
ticket (committed RC) → BEGIN RR → verify (visibility, pid, employer,
expiry, single-use, holds) → in-snapshot revalidation (mismatch →
bounded retry) → work → commit/rollback → reverse-order release, discard
on any uncertainty. Readers (session) and writers (transaction) share one
advisory namespace and one key derivation, so they mutually exclude.

Rejected alternative (boundary B, session-only revoke): integrity-sound
but leaves timeout-bounded xact-DoS and unenforced discipline — a future
direct advisory call would silently work. Boundary A is self-enforcing
(any direct use fails loudly) and deadlock-free by construction (both
sides ascend the same canonical key order).

Empirical corrections found during proof (all fixed in-experiment):
- The revoke inventory pattern `pg\_advisory\_%` misses all 8
  `pg_try_advisory_*` overloads (fail-open). Fixed to `%advisory%` with
  the count pinned at 21; the inventory test caught this.
- The three frozen helpers are SECURITY INVOKER: the revoke breaks all
  direct and trigger-transitive writer paths as app. 00013 redefines
  them as DEFINER (byte-identical bodies) — proven by
  `TestRevokeDefinerRedefinition`.
- Pool-slot leak in the discard path (close without release): fixed to
  destroy-then-release with rationale comment.
- `TestTicketOldXminAccepted` used a temp table; SetupDB revokes TEMP
  from app. Rewritten as an uncommitted INSERT (truer pressure).

## Privilege-change request (approval required)

1. **Revoke all 21 advisory overloads from PUBLIC; grant EXECUTE to
   migration_admin only.** Per-database: provisioning plus a
   migration-time guard (00013). Fail-closed inventory test pins the
   count at 21 — a future PG overload breaks loudly.
2. **Redefine the 3 frozen helpers as SECURITY DEFINER** (owner and
   grants preserved by CREATE OR REPLACE; bodies byte-identical):
   `acquire_payroll_input_lock`, `acquire_payroll_population_lock`,
   `acquire_rule_pack_lock`. Transparent to all callers (Go frozen
   helpers, timesheet/population triggers, DEFINER writers that nest
   them). Proven: direct, writer, and trigger-transitive paths all work
   as app post-redefinition. Already DEFINER (no change):
   `activate_rule_pack`, `add/verify/correct_shg_instruction`,
   `transition_employee_status`, `close_employment_terms`,
   `add/correct_compensation`. No change: disposition trigger (final
   version is lock-free), goose (migrates as migration_admin; Down/Up
   cycle proven post-revoke), provision/dbutil (superuser).
3. **Bootstrap wrapper:** `internal/employer/employer.go` takes
   `pg_advisory_xact_lock(hashtext('employer bootstrap'))` as app.
   00013 adds a narrow DEFINER wrapper; Go routes through it.
4. **Contract evolution:** `tests/integrity/privilege_test.go` pins
   `definer=false` for the 3 helpers → flip to `true` with comment.
   The 00011 migration-time INVOKER guard is unaffected (runs once at
   00011 Up; 00013 supersedes by design, files 00001–00012 untouched).
5. **No third role.** Two roles preserved throughout; migration_admin
   remains human/migration-held; functions are the capability boundary.

## 00013 scope (proposed, post-approval)

Forwards: revoke block (programmatic, count-guarded) + 3 helper
redefinitions + canonical key function (old-vs-new key equality test) +
session-barrier wrappers + hold table + ticket table + lifecycle
functions + bootstrap wrapper + migration-time privilege guards. Down:
restore INVOKER helper definitions verbatim, re-grant PUBLIC, drop
experiment-graduate objects in reverse order; 00012 Down stays valid
below it. Single forward migration; 00001–00012 byte-identical.

## Post-Gate-1 acceptance items (not authorized yet)

Barrier matrix + capture/validate once, comp/SHG/packs hardening,
statement_timeout-bounded barrier waits, deadlock-retry posture at the
caller, full battery on the revoked+redefined DB (incl. activation
live-fire), 20-section report. The 12 red attack tests turn green
against the capture implementation; the pinned gap tests
(`TestRevokeBootstrapGap`, INVOKER-denial assertions) are rewritten to
success once 00013 lands.

## Invariants verified this unit

- 00001–00011 sha256 OK against `/tmp/m2-baseline-sums.txt`; 00012
  `c05fc827…41b7d` matches Phase-0 record (file untracked, unchanged).
- OT-001 BLOCKED: `TestOT001RemainsBlocked` PASS.
- Zero statutory calculations; no capture/validation implementation.
- `gofmt` clean (module dirs), `go vet ./...` clean.
- Work uncommitted.
