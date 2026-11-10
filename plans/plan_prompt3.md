## Goal

Execute `prompt3.md` — the Phase 1 design-gate correction mandate — resolving findings M1–M4 and refinements R1–R2 from `review.md`, with doc updates, ADRs, real role-based proof tests, a consistency audit, and a completion report ending in a PASS / CONDITIONAL / BLOCKED gate verdict. No Phase 2 work; stop at the gate.

## Success Criteria

- M1: workspace ownership model complete (every entity classified, `workspace_id` + composite enforcement specified); a defensible SQL-Workbench isolation model proven by actual cross-workspace denial tests run as real PG roles; MVP explicitly fenced to a single local workspace with a documented multi-workspace path.
- M2: oracle expected outputs and diffs live only in `evaluation_private`; operational reconciliation stays investigator-visible in PayLoop; all 14 exposure paths audited with per-path verdicts; no hidden values reachable via SQL, API, templates, logs, prompts, or exports.
- M3: mode × role × scenario-`allowed_actions` authorization decided by explicit, tested logic; remediation specified as typed, audited Go services (built Phase 4/5); agents provably without write paths.
- M4: `scenario_injector` (INSERT-only sealed), `oracle_runner` (SELECT business, INSERT-only results), and evaluator identities exist with least-privilege grants proven by real-role tests; human evaluators go through app authz, never DB credentials.
- R1: sync worker holds zero business-schema access; ShiftHub→outbox→worker→PayLoop-API flow specified with the full failure matrix; `pay_period` ownership clarified (PayLoop owns; ShiftHub references by contract, resolving review's cross-reference concern).
- R2: module contract extended with optional capability descriptors (metadata registration in the existing `ToolSet` pattern — real, consumed by the reference module, no placeholders); core↛modules rule intact; order-fulfillment deferred as instructed.
- Consistency audit (§9 checklist) passes with no contradictory references; correction report (§14 A–F) delivered with evidence, not claims.

## Approach

**Authority.** prompt3 governs end to end; its §2 product requirements stay frozen and its §12 fence holds (fixtures + architectural tests only — no overtime, payroll, oracle logic, generators, worker processing, CaseTrace, or training code). `prompt2.md` is a separate regulatory-research mandate, assumed out of scope here (see questions).

**Key decisions (mine to make under the mandate):**

- D1 — M1 model: per-workspace login roles + RLS keyed off the authenticated DB identity (`current_user` → workspace map in a platform-owned table the role cannot write). Satisfies all six non-negotiables with no session vars, no shared credential, no role switching. Rejected: app-filtering-only and session-var RLS (explicitly banned); per-workspace schemas/DBs now (heavier provisioning/migration fan-out — documented as the Future A escalation).
- D2 — Prove the mechanism now on a fixture schema mirroring the design (RLS policies + roles + composite FKs), while entity-level `workspace_id` changes land as doc updates (real tables arrive Phase 2) and provisioning automation stays in 8A.
- D3 — M3 authorization ships as implemented, tested decision logic now; remediation service bodies wait for Phase 4/5. Agent read-only asserted by a test over the tool registry (no write-capable tool can register).
- D4 — Self-hosted trust caveat documented (from review.md): operators of their own installs can inspect anything; assessment integrity is guaranteed for hosted/multi-user deployments, honest practice for self-hosted. Consistent with prompt3 §3.3's no-silent-claims rule.

## Steps

1. **Setup.** Verify baseline green (`task lint`, `task build`, `go test ./...`); remove the stray empty `{decisions,ui}` entry under `docs/workforceops` after verifying it is empty; extend `tests/dbutil` with role + fixture-schema provisioning (superuser bootstrap, per-test roles).
2. **M1 — Workspace isolation.** Classify every domain entity (global/tenant/workspace/evaluator-private); specify `workspace_id` + ownership inheritance + composite constraints; select and document D1 with pooling/provisioning/Neon analysis; define the single-workspace MVP fence + migration path; specify the Workbench (SELECT-only, discovery, timeouts, limits, cancellation, export); deliverables: update `domain-model`, `database-design`, `environment-isolation`, `authorization-model`, `ui/permission-matrix`, `architecture`; create `workspace-isolation.md`, `sql-workbench-security.md`, ADR; implement §3 acceptance tests 1–8 as real-role tests on fixtures.
3. **M4 — Service identities.** Define injector/runner/evaluator-service/human-evaluator grants; deliverables: update `database-design`, `authorization-model`, `permission-matrix`, `scenario-format`, `oracle-design`; create `service-identity-permission-matrix.md`, ADR; implement §6 tests 1–8 as real-role checks (INSERT-only proofs, DDL denial, credential-absence scan of CLI/log/API output paths).
4. **M2 — Oracle secrecy.** Split operational vs oracle reconciliation; relocate sealed records; audit all 14 exposure paths with verdicts; deliverables: update `oracle-design`, `domain-model`, `database-design`, `authorization-model`, `permission-matrix`, `architecture`; create `oracle-result-isolation.md`, ADR; implement §4 tests (role denials now; API/UI/prompt leak assertions scoped to what the Phase 1 skeleton serves, remainder specified as Phase 4–6 gates without claiming them tested).
5. **M3 — Remediation.** Specify role-vs-mode matrix, scenario `allowed_actions` + diagnosis-only flag, remediation audit shape; implement the authorization decision function + tests; deliverables: update `authorization-model`, `permission-matrix`, `scenario-format`, `user-workflows`, `architecture`; create `remediation-model.md`, ADR; implement §5 tests (decision-logic + audit-shape + agent-registry assertions; service-execution tests specified for Phase 4/5).
6. **R1 — Sync boundary.** Specify outbox→HTTP delivery with service identity, idempotency, and the nine failure cases; cut worker grants to platform-only; clarify `pay_period` ownership; deliverables: update `architecture`, `database-design`, `domain-model`, API contracts; update ADR-005 + new delivery ADR; implement role-denial tests now; HTTP/duplicate/atomicity behaviors specified as Phase 3 contract tests.
7. **R2 — Module contract.** Add optional capability descriptors (services, rules, generators, oracles, faults, migrations, UI contributions) following the existing metadata pattern; reference module registers real metadata; deliverables: update `module-contract`, `module-manifest-spec`, `module-lifecycle`, `module-versioning`, `architecture`, `extension-points`, ADR; extend registry tests (unknown capability rejected, dup/incompatible still fail, reference module compatible).
8. **Consistency audit + report.** Run the §9 cross-document checklist item by item, fix contradictions at the source; write `design-gate-corrections.md` and `design-gate-correction-test-results.md`; deliver the §14 A–F completion report with actual command outputs and exactly one gate verdict.

## Validation Plan

- Per finding: its prompt3 acceptance list, each item marked proven-by-test (with test name + run output) or specified-for-later-phase (with gate) — never blurred.
- Final: `task lint`, `task build`, `go test ./...`, `go vet ./...`, and the new PG security suites executed against scratch PostgreSQL 16 here (contributors: Compose), run as the actual least-privilege roles — mocks never the sole proof for privilege claims.
- Highest-risk check: M1's cross-workspace SQL denial + workspace-forgery denial as the investigator/workbench roles. The reviewer named M1 the top concern; if applicants can read across workspaces, the verdict is BLOCKED regardless of document quality.

## Risks / Open Questions

- **Q1 (non-blocking):** `prompt2.md` (Singapore regulatory deep-research) is assumed out of this correction pass. Should it run before or alongside Phase 2 to feed rule golden tests, or stay fully independent?
- Risks: fixture-to-real-schema drift (fence: fixtures mirror the documented design verbatim, regenerated from it); overbuilding toward Phase 2 (fence: §12 scope list enforced per work unit; anything smelling like business logic gets specified, not built); contradictory doc residue (fence: §9 audit is a separate step with its own checklist, not a skim).
