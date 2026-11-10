## Goal

Implement `prompt.md` (backend-only MVP mandate: ShiftHub + PayLoop + sync worker + independent oracle + INC-001–003 + read-only CaseTrace) **as amended by** `plan3.1.md`: Amendment 1 (API-first, CLI-reproducible, **UI-accessible** via a Go-rendered Operations Console; local-first open source) and Amendment 2 (**platform kernel + business modules**, `sg-wfm-payroll` as first reference module, proven by a tiny second example module, commercial-ready but with no billing system).

## Success Criteria

- Root module `github.com/PithomLabs/a1` builds/tests/runs; `bchat/` untouched as reference; zero React/RAG/chat/embedding code in the new product.
- Clean fixed-seed baseline flows ShiftHub → sync → PayLoop → itemised payslip, reconciles with the independent oracle (overtime rule pack only; CPF gated), identically via CLI, API, and UI.
- INC-001–003 reproduce deterministically with hidden ground truth; CaseTrace investigates all three through read-only tools with every substantive claim citing replayable evidence.
- Operations Console (Go `html/template` + local HTMX + local CSS, no CDN, no npm) supports journeys A–F on real database-backed data; permission boundaries (operator/investigator/evaluator) enforced in Go + Postgres.
- `docker compose up -d && task setup && task demo` works from a fresh clone with no Neon/Fly/LLM credentials; OpenAPI spec, interface coverage matrix, and module contract published.
- `core/` provably independent of `modules/sg-wfm-payroll` (architectural tests), and `examples/modules/order-fulfillment` registers and runs on unmodified core.

## Approach

**Precedence.** plan3.1 explicitly supersedes prompt.md's no-UI prohibition and is additive on extensibility; everywhere else prompt.md's narrower-scope rule governs. plan3.1 §12's preserved decisions match the earlier plan3, which also resolves its Q1 (root module, confirmed) and Q2 (local-first Postgres, confirmed).

**Reconciliations (where the two docs sketch different shapes):**

- D1 — Processes: keep prompt.md's three processes (`shifthub-api`, `payloop-api`, `sync-worker`) + one `workforceops` CLI, implemented as thin hosts that load `core/` and register module capabilities. Rejected Amendment 2's single-`cmd/api` sketch: separate processes + DB roles are load-bearing for INC-001–003's cross-boundary realism.
- D2 — Database: core migrations own `platform`/`investigation`/`evaluation_private`; module migrations own `shifthub`/`payloop`; same six least-privilege roles; no cross-schema business reads.
- D3 — Dependency rule `core/` ↛ `modules/` enforced by package layout plus architectural tests (import check + Singapore-terminology grep + registration contract test).
- D4 — UI designed in Phase 1 (IA, journeys A–F, permission matrix) but built incrementally behind the working vertical slice; never a second business-logic implementation.
- D5 — No Go `plugin` package (static registration), no universal rules engine, no billing/marketplace in the MVP.
- D6 — License: `bchat/` contains **no LICENSE file** (verified), so license selection is an approval-gated Phase 1 deliverable, not an assumption.

## Steps

1. **Phase 0 — Audit (no product code):** bchat tree state, build/test baseline with pre-existing failures separated, toolchain check (installed Go is 1.25.7; bchat declares 1.26), Echo/pgx/Cobra/Fly-pg/Neon/OpenRouter reuse inventory, KEEP/ADAPT/REMOVE/DEFER → `docs/workforceops/repository-audit.md`, `refactor-plan.md`.
2. **Phase 1 — Platform + contracts:** kernel/module boundary, module contract + manifest spec + lifecycle/versioning, tenant activation, domain/ERD, DB/role design, API + OpenAPI, CLI contract, UI IA/workflows/sitemap/wireframes/interaction/design-system/permission-matrix, oracle design, scenario format, rule-source register, extension points, licensing recommendation, commercial/IP boundaries, ADRs. Gate: design review passes; license approved before applying.
3. **Phase 2 — Reference slice + oracle:** `core/` scaffolding (config/db/logging/health/auth/events/jobs) and `modules/sg-wfm-payroll` vertical slice (employee/terms → roster → clock → attendance → timesheet → import/mapping → payroll → payslip) through the module contract; independent oracle + golden tests; deterministic generator. Gate: fixed-seed replay identical; oracle reconciles; unsupported cases indeterminate.
4. **Phase 3 — Sync + incidents:** outbox/worker, idempotency, retries/backoff, partial failure, dead-letter, correlation IDs; reusable fault-mechanism registry; INC-001–003 with red herrings, hidden ground truth, reset/replay; inspectable via CLI/API/UI. Gate: all three reproduce from clean fixtures with stable evidence.
5. **Phase 4 — CaseTrace:** generic investigation/evidence core (case/hypothesis/evidence/claim/verification/verdict), module-registered SG tools, deterministic claim verifier, bounded LLM workflow with mock default, JSON + Markdown reports. Gate: no write path, no `evaluation_private` access, all claims cited.
6. **Phase 5 — Evaluation:** module-provided rubrics, harness (accuracy, evidence precision/recall, hypothesis discipline, unsupported-claim rate, escalation, safety, time/cost), founder human baseline, one held-out fault family. Gate: honest report with failures disclosed.
7. **Phase 6 — Console + OSS readiness:** complete Operations Console journeys A–F, OpenAPI + interface coverage matrix with cross-interface consistency tests, `docker-compose.yml`/`.env.example`/README/`AGENTS.md`/`CONTRIBUTING.md`/`SECURITY.md`/example reports/CI, fresh-clone onboarding proven. Gate: credential-free `task demo` + all journeys green.
8. **Phase 7 — Extensibility proof:** minimal `examples/modules/order-fulfillment` (one entity/workflow/rule/generator/oracle-check/fault/scenario/tool/evaluation) on unmodified core; architectural boundary tests; module SDK/template + contribution guide. Gate: both modules coexist; core untouched by module addition.

## Validation Plan

- Every phase: `go build ./...`, `go test ./...`, `go vet ./...`; scoped suites for overtime/applicability, mapping, idempotency, retry transitions, oracle independence (fails if PayLoop calc shared), API contract tests, CLI integration tests, UI journey/permission/empty-state tests, role-permission denial tests, answer-key isolation tests.
- Highest-risk check: clean-baseline oracle reconciliation + incident replay determinism (`task scenario:run ID=INC-001` twice → identical evidence hashes) + cross-interface consistency (CLI/API/UI equivalent ops → identical persisted results). If any is flaky, later phases are meaningless.
- Release check: fresh-clone `docker compose up -d && task setup && task demo && task web:run` with no credentials; Fly build + migrate + health against an isolated non-production database only — no prod secrets, deploy, or destructive migration without explicit authorisation.

## Risks / Open Questions

- **Q1:** which Fly app/region should the new config target, and who authorises the first non-prod deploy?
- **Q2:** is a Singapore payroll practitioner available to review the overtime rule pack, or do I proceed with source-pinned golden tests + independent code review?
- **Q3:** license preference (e.g. MIT vs Apache-2.0)? I will propose one with implications in Phase 1 regardless; nothing is applied without your approval.
- Risks: MOM misinterpretation (source register + `indeterminate` fallback); overbuilt framework before the slice works (gated by Phase 2; example module stays tiny); hidden-answer leakage through UI/API (permission-matrix tests + template audit); scope creep into full payroll product (phase gates + deferred list enforced).
