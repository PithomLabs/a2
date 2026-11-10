# Consolidated Phase 6.1 plan (Claude + DeepSeek + Z + ChatGPT reviews merged)

Target: `plans2/prompt2.md` (four Phase 6 reviews). Consolidation rule: merge
where coherent, keep valid and cohesive; `plans/prompt2.md` (Singapore
regulatory deep-research mandate) is explicitly excluded — incoherent with the
synthetic-only POC boundary (OT-001 BLOCKED); it becomes its own future track.

## Goal

Close the Phase 6 POC to a gate-decidable state: implement every review finding once (deduplicated across the four reviewers in `plans2/prompt2.md`), execute the adversarial battery on a docker-capable host, and produce per-technology go/no-go evidence — without touching frozen migrations, Gate 1 artifacts, or production decisions.

## Success Criteria

- Gate A: all code compiles; unit suites green with honest skips; ADRs revised; traceability matrix, `EXPERIMENTS.md`, dynamic-rules table, and CI workflow written; U4 cost estimated; POC-REPORT uses the Implemented/Built/Executed/Passed/Failed/Blocked/Unexecuted taxonomy.
- Gate B: E5 green on PG18.6; T1–T13 green in both orderings as `app`; kill-9, tamper, and auditor experiments green; full month produces all 6 artifacts with F1a/F1b stated.
- Gate C: per-technology adopt/defer/reject recorded with evidence tags; nothing merged, no 00013, Gate 1 still pending.

## Approach

Backbone is ChatGPT's 3 gates (A: no-DB closure → B: docker execution → C: adoption). Claude's hygiene + 7 build items, DeepSeek's 3 work-blockers, and Z's 9 findings slot into Gate A as single merged tasks (table below). DeepSeek's sequencing binds: **R/P decisions first, then blockers, then runtime**. All R1–R7 draft rulings stay PROPOSAL until human sign-off — the plan executes Gate A on POC-scoped provisionals only.

Consolidation resolutions (verified against the tree: ADR-0204's `immu_tx`-in-payload + UPDATE back-link, ADR-0203's 11+1 battery, PATTERNS C5/C8/K7, `poc-run`'s implicit modes):

| # | Merged finding | Resolution |
|---|---|---|
| 1 | Z#1 + ChatGPT#2 (provenance defects) | One ADR-0204 rewrite: drop `immu_tx` from payload, event-uuid keys `prov/.../{event_uuid}`, **append-only** `ledger_backlink` table (no UPDATE), MustNotExist + no-delete/no-expire governance, idempotency value binds content digest, pin calculator build digest; update `ledger.go`, `outbox.sql`, C8/K7 |
| 2 | Z#2 + DeepSeek T1–T13 | Add T13 advisory-substitution/self-acquisition + grant assertions (revokes, trigger ownership); extends R3 scope |
| 3 | Z#3 + R6 drafts (all: same-DB) | Same-DB shared pool for POC; document separate-DB replay fallback in ADR-0203; forged-state test is the tripwire to reopen |
| 4 | Z#4 (ticket-vs-gate bake-off?) | Human picks in Step 0: ~1-day disposable-DB bake-off, or recorded prior-authority note (recommended: Gate-2 already retired Option A; cheaper, no resurrection) |
| 5 | Z#5 + Z#7 + Z#8 + Z#9 | New artifacts: dynamic-rules classification table, `EXPERIMENTS.md` (E1–E6 + DBOS 1–8 pinned), genesis-pinning procedure (with auditor task), hygiene pass (canonical tables, rung-vocabulary note, Change Date in R1) |
| 6 | Z#6 + DeepSeek 8-point | Split F1→F1a (blocker-path) / F1b (computed-path, red-by-design until proration/X-FUND settle); merge F-table with the 8-point criteria into one acceptance table |
| 7 | ChatGPT#5 (reporting precision) | `poc-run --demo-synth-only` (documented exit 0) vs full mode (fail-loud); fix `core/database` skip gap; CI *requires* DB (never green-by-skip) |
| 8 | ChatGPT#6 + Claude hygiene (versions) | Verify-then-pin: DBOS v1.8.0 and immudb 1.11.2 claims are **unverified** (checkouts show DBOS unversioned, immudb v1.11.1) — confirm, re-verify datasource/replay semantics on any delta, replace `replace` directives if reachable, never invent digests |
| 9 | ChatGPT#9 (cutoff backdating) | Gate-2 DDL acceptance: DB-controlled timestamps only (client `approved_at` never decides eligibility), gate-ordered approve-vs-close |
| 10 | ChatGPT#4/§3 (traceability) | Decision matrix: binding constraint → D1–D8 → R1–R7 → ADR → P1–P5 → implementation → test; flags any POC choice implying production authorization |

## Steps

**Step 0 — Human rulings (blocks Gate B; Gate A proceeds on provisionals where marked independent).** Confirm/amend: R1–R6 (reviewers converge on approve-with-conditions — license Appx records Change Date; auditor separate container/volume/creds, genesis hash in git; rung kept, no Rewind/Delete/truncate, DBOS tables untrusted for authorization; PG18 on disposable clusters, PG16 kept; Gate-2 evidence-only; same-DB), R7 two tranches (sign D1/D3/D4/D5/D7/D8 principles now; D2/D6 + proration/B1/B2/X-FUND/cutoff-subquestions/payroll-line-governance stay PROPOSED), P1–P5 confirm/amend, finding #4 disposition.

**Gate A — closure without a database (ordered by dependency).**
A1. Hygiene: rename-purity proof (`git diff -U0 -- '*.go'` shows import paths only, kept as separate patch); `core/database` skip fix (nothing else in package); version verify-then-pin (#8).
A2. Traceability matrix + P1–P5 crosswalk (#10).
A3. ADR-0204 rewrite + `ledger.go`/`outbox.sql`/C8/K7 to match (#1).
A4. ADR-0203: T13 + grant assertions + fallback variant + #4 disposition (#2–#4).
A5. Rules table + `EXPERIMENTS.md` + F1a/b + hygiene (#5, #6).
A6. Gate-2 gate DDL (POC-only SQL under `poc/`, disposable DBs) with #9 criteria + writer-matrix enforcement.
A7. U4 Tx-threading spike: minimal compilable change + recorded cost estimate (full refactor needs live PG).
A8. External auditor program (own state file, separate-container runnable) + genesis procedure + no-beyond-model tamper claims.
A9. `poc-run` failpoints (5 boundaries, env-enabled) + kill-matrix test + forged-state test (as `app`: forged Recv/step-output must move nothing without PG-enforced authorization), all skip-loud.
A10. Mode split (#7) + missing experiments (double-enqueue idempotency, perf timings, data-only new rule type — or report why the formula can't be data).
A11. CI workflow (PG18.6 + immudb services, runbook steps 2–6, artifact upload).
A12. POC-REPORT restructure (taxonomy + 8-point + F-table + demo digest + P-status + U4 estimate).

**Gate B — execution on a docker-capable host (Claude's order; go/no-go per technology after each stage; stop conditions armed).** B1: E5 (migrations + full suite + PG16 parity; on failure stop, record verbatim, never edit frozen files). B2: E1–E3, Tx-spike execution, capture under `kill -9`. B3: forged-state test. B4: outbox crash matrix, tamper battery with auditor. B5: full month (register, payslips, exceptions, payment preview, close + evidence reports; post-finalization adjustment leaves original bytes equal; reproducibility pins inputs + pack + calculator digest).

**Gate C — adoption (separate human approvals).** PG18-supersedes-16?, DBOS retained?, immudb license/ops acceptable?, 00013 scope? — each decided on Gate B evidence. Gate 1 remains pending; no merge until C passes.

## Validation Plan

- `go build/test/vet/gofmt` green in both modules (Go env flags from POC-REPORT; `task lint` after its two pre-existing fixes land).
- Rename purity: `git diff -U0 -- '*.go'` contains only import-path lines; `git diff -- migrations/` empty; no `00013` file exists.
- Skip semantics: suite without DB URL skips (exit 0); CI without DB URL fails.
- Pins: exact DBOS source revision + immudb version + Change Date + image digests recorded; datasource/replay semantics re-verified on any version delta.
- Battery: T1–T13 both orderings, real `app` role, persisted-state asserts, three time zones.
- Crash/tamper: 5 boundaries exactly-once; PG-row + ledger-disk tamper both fail loudly with webhook; auditor verifies from external state.
- Month: 6 artifacts; F1a/F1b stated; adjustment byte-equality; rerun identical digest.
- No invented evidence: every new claim tagged; skipped tests never counted.

## Risks / Open Questions

- R/P sign-off still pending — until Step 0, only provisional-scoped work proceeds; any item needing a ruling is documented and skipped, not improvised.
- U4 estimate may re-size Gate B from days to weeks (DeepSeek) — A7 exists to surface this before runtime.
- DBOS v1.8.0 delta is the highest semantic risk: if datasource/replay behavior changed, Phase 3 conclusions need revision before trusting the atomic path.
- F1b stays red by design until proration/X-FUND/B1/B2 settle — expected outcome, not gate failure.
- Docker-host + registry/proxy reachability unconfirmed (Claude caution) — B1 doubles as the environment proof.
- Non-goals: production 00013, Gate 1 ruling, merge, statutory content, paid-hosted licensing, PG16 removal, `plans/prompt2.md` regulatory track.
