# Migration & adoption plan (Phase 7)

Superseding the PG16-only local stack with PostgreSQL 18.6 + DBOS + immudb,
sequenced so every step proves itself and rolls back. Nothing here is
executed: the POC run that gates adoption is blocked on a docker-capable
host (see `poc/payroll-month/POC-REPORT.md`). All production moves additionally
need R1–R7 (`DECISION-REQUESTS.md`) and the pending Gate 1 ruling.

## 1. Supersede PG16 Docker (dev stack)

Recommended path (RESEARCH.md §2.4): fresh 18.6 cluster + `migrate up` +
reload. No production data exists (recycle-by-drop is the norm), so
`pg_upgrade --link` is reserved for the future real-data path (dry-run spec E6).

1. Pin `postgres:18.6-alpine` by registry digest in a new compose service;
   mount the data volume at `/var/lib/postgresql` (PG18 versioned layout —
   `.../18/docker` underneath; the old `.../data` mount does NOT persist).
2. Run `scripts/db/init-roles.sh`, then `migrate up` for 00001–00012.
3. Parity check: seed MERIDIAN demo on 16 and 18, diff canonical snapshot
   bytes + suite results (E5). Investigate any `pg_locks`/advisory-count drift.
4. Keep the 16 compose file + volume untouched during evaluation; rollback is
   `compose down` on 18 and restart of the 16 stack (never `-v`).
5. "Superseded" = the PG16 service is deleted from the dev stack only after
   E5 green + battery green on 18. Do not delete the old volume until then.

## 2. Sequenced adoption (each step: proof + rollback)

Steps 2–5 below are the POST-GATE adoption phase (Gate C inputs); they are
not POC steps and none executes before Gate B evidence exists.

| Step | Change | Proves it worked | Rollback |
|------|--------|------------------|----------|
| 0 | Criteria pinned | ADRs 0201–0204 accepted by human | n/a |
| 1 | POC stack bring-up (`poc/`) | `poc-run` exit 0 full-month on fresh host | `compose down -v`; main tree untouched (nested module) |
| 2 | PG18 temporal constraints (additive DDL only) | E1–E3 green; plain-exclusion tables migrated; comp/SHG triggers kept | Drop added constraints/columns (new migration down) |
| 3 | DBOS capture + month orchestration | Crash battery green; exactly-once history evidence | Feature-flag to legacy capture path; workflows drained |
| 4 | immudb provenance + attestation | Tamper battery green; auditor verifies from external state | Ledger is additive sidecar; PG stays authoritative; stop appending |
| 5 | Cutover + deletion | F1–F7 green; deletion ledger L1–L11 proven | Revert per-step; L-deletions are code removals, restorable from git |

Rules: no step edits migrations 00001–00012; step 2+ additive DDL ships as new
migrations only after ADR-0202's "no 00013 yet" is explicitly superseded; the
`a1→a2` import-path rename on `stack-poc` merges separately first (it is a
baseline fix, not POC work).

## 3. CI

- Matrix job on PG18 (fresh service + `migrate up` + full suite with
  `WORKFORCEOPS_TEST_DATABASE_URL`).
- Required: translated 12-attack battery (ADR-0203), kill-test job, tamper-test
  job. Pre-existing gap to fix first: `core/database` admission/gate2 tests
  must skip without a DB URL (currently fail at dial).
- Fix `task lint`: drop the nonexistent `sdk/` dir from gofmt scope; make
  `go vet` immune to the parent `go.work` (`GOWORK=off` or workspace-aware).
- POC job (nested module): `go build/test` + `poc-run` demo-mode exit-code
  assertion + `gofmt -l poc`.

## 4. Risk register (owner + trigger)

| Risk | Owner | Trigger / mitigation |
|------|-------|----------------------|
| DBOS Go SDK drift (no version pin) | POC bring-up | Pin registry version before step 3; record in go.mod |
| immudb BSL-1.1 vs open-source posture | HUMAN (R1) | No production use before ruling; narrow replaceable interface kept |
| Dual-write crash windows | POC bring-up | Outbox + idem-key + back-link proven by kill matrix before step 4 |
| PG18 behavior drift (locks/isolation) | POC bring-up | E5 diff; any drift reopens ADR-0201 |
| DBOS privilege boundary (system tables vs app) | HUMAN (R6) | `SkipMigrations` + minimal grants verified in bring-up |
| Tx-threading refactor (UNRESOLVED-4) | POC build | Costed against live PG; blocks step 3 capture depth |
| Verifier-state custody + genesis | HUMAN (R2) | Auditor unwired until decided; no tamper claims before |
| Team learning curve (workflows, proofs) | Eng lead | Runbook + who-owns-what (below) before step 3 |

## 5. Who owns what + onboarding

Canonical technology-allocation table: PATTERNS.md §3. This section is the
adoption-facing summary of it; on any conflict the PATTERNS table wins.

| Concern | Owner | Non-owner contract |
|---|---|---|
| Operational truth, constraints, effective-dating | PostgreSQL (+ a2 services) | DBOS/immudb never write business state directly |
| Durable execution, retries, scheduling, approvals | DBOS workflows | Steps are deterministic shells; business logic in services/txns |
| Tamper evidence, attestations, close records | immudb sidecar | Hashes/metadata only; PG stays queryable truth |
| Money math, rule semantics | `rules/` + `internal/` (pure, exact) | No float; no handler-side calculations (AGENTS.md) |

Onboarding path: read AGENTS.md → RESEARCH.md → PATTERNS.md → run
`POC-RUNBOOK.md` steps 1–5 → read POC-REPORT.md gaps → pick a red cell.
