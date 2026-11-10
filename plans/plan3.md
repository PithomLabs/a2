## Goal

Implement `prompt.md` — transform the `bchat` Go/Echo/Postgres foundation into the backend-only **WorkforceOps Lab** MVP (ShiftHub + PayLoop + sync worker + independent oracle + mechanism-based fault injection + read-only CaseTrace investigation), scoped by `WORKFORCEOPS_LAB_MVP_PLAN_v0.1.md` — with the new codebase living under the module `github.com/PithomLabs/a1`.

## Success Criteria

- New root Go module `github.com/PithomLabs/a1` builds, tests, and runs locally with zero frontend/RAG/chat dependencies; `bchat/` untouched as reference.
- Clean fixed-seed baseline flows ShiftHub → integration → PayLoop → itemised payslip and reconciles with the independent oracle (overtime rule pack only; CPF gated).
- INC-001/002/003 each reproduce deterministically from seed + scenario, with hidden ground truth inaccessible to the investigator role.
- CaseTrace investigates all three incidents through typed read-only tools; every substantive claim cites evidence IDs; verifier replays evidence.
- `task demo` runs the full chain end-to-end; Fly.io/Neon deploy config is ready and verified against an isolated (non-production) database; README/`AGENTS.md` describe WorkforceOps Lab.

## Approach

**Placement (authorised deviation from prompt §3.2):** prompt.md says "keep the existing Go module path unless explicitly authorised" — the `github.com/PithomLabs/a1` directive is that authorisation. Create the new module at the workspace root (`/a1`) and **selectively port** working infrastructure from `bchat/` (read-only reference, no edits, no destructive migrations). This satisfies prompt §3.3's removal requirement more safely than deleting RAG/frontend in place.

**Reuse, grounded in inspected `bchat/` evidence:**

- Keep: Echo (`bchat/go.mod`), pgx/v5 as sole Postgres driver (`AGENTS.md` + `store/db/postgres`), viper/cobra bootstrap (`bin/memos/main.go`), `/healthz` checks and Fly pg profile (`fly_pg.toml`, `Dockerfile.pg.fly`), OpenRouter client seam (`revrost/go-openrouter`) isolated behind a provider-neutral interface.
- Drop: RAG/LanceDB/CGO pipeline, embeddings, chat sessions, widget, React frontend, Memos domain tables — none ported; nothing from `server/router/api/v1/agent` reused except the LLM-connectivity pattern.
- Migrations: Phase 0 audit decides between fresh Goose migrations under `a1/migrations/` (leading option — clean boundary, zero legacy-table risk) vs. adapting `bchat`'s Memos-coupled migrator. Either way, new schemas only; never touch existing `bchat` tables.

**Key decisions:**

- D1 — One Go module, three processes (`shifthub-api`, `payloop-api`, `sync-worker`) + one `workforceops` CLI (migrate/seed/oracle/scenario/investigate/evaluate). No microservice fleet, no Kafka/Redis — Postgres transactional outbox + Go worker.
- D2 — One Neon project, five schemas (`shifthub`, `payloop`, `platform`, `investigation`, `evaluation_private`), six least-privilege roles; cross-system traffic via APIs, enforced in Go + Postgres, never prompts alone.
- D3 — Oracle is a separately implemented package that must not import PayLoop calc code; a test fails the build if the routine is ever shared.
- D4 — Overtime (MOM Part IV) is the only v0.1 statutory rule pack, with applicability/exclusions and `indeterminate` for unsupported arrangements; CPF deferred until complete official tables are pinned and reviewed; leave/PH/IRAS/bank explicitly out.
- D5 — No UI. APIs + CLI + JSON/Markdown reports only. Build/test must pass with no live LLM key (mock provider).

## Steps

1. **Phase 0 — Audit (no new product code):** record `bchat` branch/tree state, build/test baseline with pre-existing failures, Neon/Fly reuse assessment, KEEP/ADAPT/REMOVE/DEFER register → `docs/workforceops/repository-audit.md`, `refactor-plan.md`.
2. **Phase 1 — Design + skeleton:** architecture, domain model/ERD, API contracts, DB/role design, rule-source register, oracle design, scenario format, ADRs → `docs/workforceops/`; then root module, config/db/logging/health bootstrap, migrations, API + worker skeletons, Taskfile, Dockerfile/Fly readiness. Gate: builds, starts, connects to isolated DB, health green.
3. **Phase 2 — Clean slice + oracle:** employee/terms → roster → clock events → attendance → approved timesheet → import/mapping → payroll line → payslip; independent oracle + golden tests; deterministic generator (50–100 staff, 3–6 months, `Asia/Singapore` business dates, UTC instants). Gate: fixed-seed replay identical; oracle reconciles; unsupported cases indeterminate. Do not proceed to faults until this passes.
4. **Phase 3 — Integration + incidents:** outbox/worker, idempotency, retries/backoff, partial failure, dead-letter, correlation IDs; INC-001 (pay-code mapping), INC-002 (dedup/ordering/partial-batch), INC-003 (external-ID/cut-off/dead-letter) as reusable mechanisms with red herrings, hidden ground truth, reset/replay. Gate: all three reproduce from clean fixtures with stable evidence.
5. **Phase 4 — CaseTrace:** ticket intake, hypothesis loop with falsifiers, ~13 typed read-only tools, evidence ledger (run/hypothesis/evidence/claim/verdict/tool_execution), deterministic claim verification, bounded LLM workflow, JSON + Markdown reports. Gate: no write path, no `evaluation_private` access, all claims cited.
6. **Phase 5 — Evaluation + demo:** harness scoring root-cause accuracy, evidence precision/recall, hypothesis discipline, unsupported-claim rate, escalation, safety, time/cost; founder human baseline; one held-out fault family; `task demo` end-to-end; final report in prompt §20 format (A–J). Gate: honest results with failures/limitations disclosed.

## Validation Plan

- Per phase: `go build ./...`, `go test ./...`, `go vet ./...`, plus scoped suites — unit (overtime/applicability, mapping, idempotency, retry transitions, oracle), integration (both APIs, migrations, role-permission denials, outbox/dead-letter), scenario (baseline reconciliation, INC-001–003 replay determinism, answer-key isolation, read-only boundary).
- Highest-risk check: clean-baseline oracle reconciliation + INC replay determinism (`task scenario:run ID=INC-001` twice → identical evidence hashes); if this is flaky, Phases 4–5 are meaningless.
- Deploy check: Fly build + migrate + health against isolated Neon branch only; no production secrets touched, no prod deploy without explicit authorisation.

## Risks / Open Questions

- **Q1 (blocking before Phase 1):** confirm new root module `github.com/PithomLabs/a1` with `bchat/` left read-only, vs. in-place `bchat` refactor + module rename. Assumed: root module.
- **Q2:** local Postgres (Docker) for daily dev with a Neon branch only for deploy verification, or Neon branch for all dev? Assumed: local-first.
- **Q3:** which Fly app/region should the new config target, and who authorises the first non-prod deploy? No deploy happens without your explicit go-ahead.
- **Q4:** is a Singapore payroll practitioner available to review the overtime rule pack, or do I proceed with source-pinned golden tests + independent code review only?
- Risks: MOM guidance misinterpretation (mitigate: source register + applicability matrix + `indeterminate` fallback); agent overfitting to injectors (mitigate: mechanism parameters + held-out family); scope creep into full payroll product (mitigate: phase gates, deferred list enforced).
