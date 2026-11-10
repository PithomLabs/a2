# ADR — Payroll Snapshot Synchronization: Option B (RC + Transaction-Scoped Employer Gate)

Status: DECIDED (Gate 2 experiment). Production migration NOT authorized by this ADR.

## Decision

Adopt READ COMMITTED + one transaction-scoped employer advisory gate +
database-enforced writer participation + own-row phase marker as the
payroll snapshot synchronization architecture, and explicitly
**supersede Prompt17's pre-BEGIN REPEATABLE READ / session-lock /
XID-ticket protocol** for snapshot admission.

## Context

Gate 1 proved an XID-ticket + session-barrier mechanism. Gate 2 tested
whether barrier ownership survives post-snapshot release/reacquisition,
and built an isolated Option B alternative. Full evidence:
`plans/gate2_synchronization_experiment.md`.

## Why Option A was not retained

Live attacks (`core/database/gate2_test.go`, 13 checks) show the ticket
proves ordering only: pre-BEGIN `DISCARD ALL` yields verify-passing
snapshots with zero physical locks (unrestorable under the Gate-1
revoke, which denies the raw re-lock); post-verify release commits
stale read sets undetected on both plain and revoked databases.
Retention would need a physical-presence probe, a pre-commit re-check,
and a repair path — none exists, and the result would exceed Option B
in complexity.

## Why Option B

`core/database/gate2_optionb_test.go` (18 tests, 66 checks) proves on
PG 16.15: non-RC rejected at gate/capture/validate; one canonical
gate shared by capture, all 9 writer classes (direct I/U/D, DEFINER
nests, upsert, multi-row, COPY, MERGE), and triggers; both orderings
correct per class; same-transaction post-decision mutation rejected
(visible + in-progress ⟺ own row; savepoint-exact); Go claims
proof-checked against post-gate reads with forged-hash and planted
artifacts rejected; failures atomic with typed errors; no revoke, no
new role, no session state (plain transactions).

## Consequences

- Production design target for the future snapshot-hardening migration
  (single forward migration; scope in report §16).
- ADR-1 stands: Go resolver + canonical bytes + hash remain the single
  implementation; SQL proof-checks, never re-solves.
- The 12 red attack tests remain the acceptance battery for the
  production capture implementation.
- Residuals: single-employer-per-write-txn discipline required
  (multi-employer can deadlock with typed `40P01`); per-row check cost
  unmeasured; new input tables must join the trigger coverage list;
  pooler compatibility inferred, not run.
