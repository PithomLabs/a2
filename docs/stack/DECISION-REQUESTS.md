# DECISION-REQUESTS — human decision-request list (research gate)

Status: OPEN — every item blocks the Phase 6 gate until answered. Owner of all
answers: HUMAN. Context: [DECISIONS.md](DECISIONS.md) (D1–D8),
[RESEARCH.md](RESEARCH.md) (evidence), `adr/` (technical proposals).

| # | Request | Why human-only | Blocks | Evidence |
|---|---|---|---|---|
| R1 | immudb BSL-1.1 license: accept source-available dependency for POC? Record Change Date of pinned build; rule on paid-hosted future (counsel + Codenotary capability list?) | License-class mismatch with MIT open-source backend; competitive-offering clause is a legal judgment | ADR-0204; any POC ledger code; fallback = PG+DBOS | Phase 4 §1 + UNRESOLVED 1–2 |
| R2 | Trusted-state custody: where do verifier state + auditor root hashes live (host/volume/zone), who holds keys, genesis pinning procedure? | Custody inside the writer defeats tamper-evidence; first `GetState` bootstraps from server | ADR-0204 acceptance; auditor design | Phase 4 §3 + UNRESOLVED 3, 5 |
| R3 | Threat-model changes: confirm Rung 5 retained for POC (D4); rule on `migration_admin`/superuser powers, immudb truncate rights vs "never edit approved artifacts", DBOS Rewind/Delete exposure | Lowering a rung or granting destructive powers silently is forbidden (Phase 0 D4; AGENTS.md) | Enforcement tier; deletion ledger L7/L9 | Phase 0 D4; Phase 3 §5; Phase 4 UNRESOLVED 7 |
| R4 | Upgrade timing: authorize PG 18.6 as POC baseline now (fresh cluster, old 16 volume preserved)? Defer full dev-baseline supersede to Phase 7? | Volume layout change; old data access; E5 unexecuted | ADR-0201; POC bring-up | Phase 2 §2.4; binding plan §Phase 7 |
| R5 | Gate 1 interplay: confirm POC may use Gate-2 direction (RC + txn-scoped gate) on disposable DBs WITHOUT pre-empting the pending production Gate 1 ruling | Pre-empting Gate 1 forbidden (binding Key Decision #4); production migration not authorized | ADR-0203; capture build | Gate 2 ADR :3; Phase 0 §3 |
| R6 | DBOS system-DB placement: same-DB `dbos` schema (atomic path, shares app DB) vs separate DB (replay path)? Rule vs "one database" constraint | Placement decides atomicity-vs-replay for every workflow; "one database" is an AGENTS.md architectural identity | ADR-0201/0203; workflow catalogue | Phase 3 §2; AGENTS.md §1–8 |
| R7 | D1–D8 business sign-off + Ten Killer Questions (esp. B1/B2/X-FUND, proration, cutoff sub-questions, G-PREREQ) | Business semantics precede mechanism (checklist "Business Closure Before Technical Synchronization") | 00013 scoping; capture population; coverage triggers | [DECISIONS.md](DECISIONS.md) §§1–4 |

Response format requested: per item, APPROVED / APPROVED-WITH-CONDITIONS /
REJECTED + rationale + effective date. Conditions become acceptance criteria.

## Staged facts (Gate A, NOT rulings — owner still HUMAN)

- R1: immudb v1.11.2 pinned (proxy VERIFIED, tag 2026-09-02 git bfdce036);
  computed Change Date ~2030-09-02 (counsel to confirm); image digest
  `sha256:3033c1b5…` pinned in compose; no vendoring; narrow interface kept.
- R4/R6: DBOS v1.8.0 pinned (proxy VERIFIED, tag 2026-10-28 git f543516f);
  checkout-vs-release delta re-verified non-load-bearing; PG image digest
  `sha256:300ebd02…` pinned; PG16 volume untouched.
- R5: ticket-vs-gate disposition recorded in ADR-0203 (prior-authority note,
  PROPOSED) — bake-off remains available on request.
- R2/R3/R7: no new facts; auditor + genesis (A8) and forged-state tripwire
  (A9) implement the proposed models without claiming rulings.

## Rulings (human, recorded verbatim as decided)

- R1 (2026-11-10): APPROVED as drafted — BSL-1.1 accepted for POC
  non-production use; Change Date ~2030-09-02 recorded (computed from
  v1.11.2 tag time); counsel required before any hosted/paid use; no
  vendoring; narrow replaceable interface kept.
- R2 (2026-11-10): APPROVED as drafted — external auditor (separate
  container/volume/creds, genesis hash to git signed); threat boundary at
  ledger-operator/disk tampering; hostile host admin out of scope.
- R3 (2026-11-10): APPROVED as drafted — rung-5 scope kept for
  capture/provenance claims; no Rewind/Delete/truncate; DBOS system tables
  untrusted for authorization; forged-state failure reopens R6.
- R4 (2026-11-10): APPROVED as drafted — PG18.6 on disposable clusters only
  (fresh + migrate up + reload); PG16 stack untouched until E5 + battery green.
- R5 (2026-11-10): APPROVED as drafted — Gate-2 direction as POC evidence
  only; production Gate 1 ruling unaffected and still pending.
- R6 (2026-11-10): APPROVED as drafted — shared-pool placement for the POC
  (dbos schema, least-privilege app role); separate-DB replay fallback
  documented, not built; forged-state failure reopens.
- R7 (2026-11-10): TRANCHES APPROVED — D1/D3/D4/D5/D7/D8 principles signed;
  D2/D6 depth + proration/B1/B2/X-FUND/cutoff-subquestions/payroll-line-
  governance stay PROPOSED; F1b red by design until settled.
- P1–P5 + finding #4 (2026-11-10): CONFIRMED — P1 narrowed to registry pins,
  P2/P3 held, P4 closing via A6/A7, P5 superseded; recorded prior-authority
  note stands, no ticket-vs-gate bake-off.
