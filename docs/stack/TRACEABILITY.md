# Decision traceability matrix (Gate A)

Binding constraint → D1–D8 → R1–R7 → ADR → P1–P5 → implementation → test.
Statuses: PROPOSED (awaiting human), POC-SCOPED (provisionals P1–P5, binding
nothing in production), DONE (executed this programme), OPEN (needs ruling).

## P1–P5 crosswalk (quoted from POC-REPORT.md, current disposition)

| # | Provisional (as written) | Disposition now |
|---|---|---|
| P1 | immudb checkout as replace; `codenotary/immudb:1.11` for run env; non-production inside use grant | PARTLY SUPERSEDED: registry pins now (go.mod v1.11.2, image `1.11.2@sha256:3033…`, Change Date ~2030-09-02 computed from tag time); non-production scope unchanged; R1 ruling still OPEN |
| P2 | verifier custody + genesis undecided; ledger verifies at append; external auditor runbooked | HELD: auditor + genesis procedure land in A8; R2 ruling still OPEN; no beyond-model tamper claims |
| P3 | shared-pool wiring per ADR-0203; general placement open | HELD: same-DB for POC; R6 ruling still OPEN; separate-DB fallback documented in ADR-0203 (finding #3) |
| P4 | capture txn body is SELECT placeholder; Gate-2 + Tx-threading wait on U4/live PG | CLOSING: A6 DDL + A7 spike this gate; R5 ruling still OPEN; Gate 1 not pre-empted |
| P5 | DBOS pinned by checkout path only; images by tag with digest TODOs | SUPERSEDED by A1: DBOS v1.8.0 + immudb v1.11.2 registry pins (re-verified), both image digests pinned from Hub API |

## Full matrix

| Binding constraint | D | R | ADR | P | Implementation | Test | Production disposition |
|---|---|---|---|---|---|---|---|
| Invariants DB-enforced vs direct SQL (AGENTS.md) | D4 rung retained | R3 OPEN | ADR-0203 gate + battery; C10 | — | A6 POC DDL (disposable DBs); L4/L8 keeps | T1–T13 as `app`, both orderings | Needs R3 ruling; no prod DDL |
| Cutoff = business close (decisions.md) | D1 PROPOSED | R7 OPEN (tranche 1) | close-txn pattern (§4.2) | — | A6 `poc_close_period` + DB timestamps | T12 race; F1a blockers | Needs R7 sign-off |
| Corrections supersede, never edit (AGENTS.md §7) | D2 PROPOSED | R7 OPEN (deferred depth) | ADR-0204 (rewritten A3) | — | event-uuid keys + append-only backlink | Crash matrix; byte-equality after adjustment | Needs R7 + legal retention review |
| Lifecycle states are verified facts | D3 PROPOSED | R7 OPEN (tranche 1) | run state machine (existing + POC asserts) | — | 00012 triggers (kept) + POC grant asserts | T10/T11 probes | Needs R7 sign-off |
| Approved-only inputs; loud blockers | D5 PROPOSED | R7 OPEN (tranche 1) | coverage gate (PERIOD FK) | — | E2/coverage harness; capture population | Coverage tests; F1a | B1/B2/X-FUND still open |
| Bitemporal retro representation | D6 PROPOSED | R7 OPEN (deferred) | successor lineage | — | none yet (F1b red by design) | none yet | Blocked on proration semantics |
| Employer = legal boundary | D7 PROPOSED | R7 OPEN (tranche 1) | employer-scoped gates/keys | — | gate key derivation; no x-employer txn | Cross-employer writer test | Needs R7 sign-off |
| Reproducible from pinned evidence | D8 PROPOSED | R7 OPEN (tranche 1) | canon + pins + calc digest (new) | — | synth digests; ledger record schema | Re-derivation equality; F1a/F1b | Needs R7 sign-off |
| Synthetic-only POC; OT-001 BLOCKED | — | — | all ADRs (scope) | — | `synthetic:true` + activation guard | `TestSyntheticCannotActivateAsStatutory` DONE | Statutory track separate (out of scope) |
| No overlapping systems (PG/DBOS/ledger) | — | R6 OPEN | ADR-0201 (alt 2 viable) | P3 held | shared-pool cmd wiring | Crash battery; forged-state tripwire | Needs R6 ruling + AGENTS.md §8 note |
| BSL-1.1 recorded, not assumed | — | R1 OPEN | ADR-0204 | P1 narrowed | registry pin, no vendoring, narrow interface | license-file test (SPDX header present) | Needs R1 ruling + counsel for hosted |
| Tamper-evidence needs external trust | — | R2 OPEN | ADR-0204 | P2 held | A8 auditor + genesis anchor | Tamper battery + webhook | Needs R2 ruling |
| Gate 1 not pre-empted | — | R5 OPEN | ADR-0203 + #4 note | P4 closing | A6/A7 on disposable DBs, labeled evidence | T-battery (evidence, not ruling) | Gate 1 pending regardless |
| PG18 baseline safety | — | R4 OPEN | ADR-0201 | — | digest-pinned compose, 16 volume kept | E5 + parity; rollback = restart 16 | Phase 7 decision (Gate C) |
| Migrations 00001–00012 frozen; no 00013 | — | — | ADR-0202 | — | POC DDL under `poc/` only | `git diff -- migrations/` empty; no 00013 file | 00013 needs Gate C + R7 |

## Consistency notes (checked 2026-11-10, Gate A)

- ADR-0201/0203 select shared-pool while R6 is OPEN: CONSISTENT-BY-SCOPE
  (POC disposable DBs + documented fallback + forged-state tripwire). No
  production implication drawn.
- P4 placeholder vs ADR-0203 atomicity claim: known GAP, closed by A6+A7
  (DDL + spike) and proven in Gate B; until then the claim reads PROPOSED.
- P1–P5 confirmations must never be read as D/R approvals: the matrix keeps
  separate columns; D-column and R-column stay PROPOSED/OPEN throughout Gate A.
- No POC choice in this matrix implies production authorization; every
  production-disposition cell says what is still required.
