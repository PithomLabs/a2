## Goal

Implement `plan3.3.md` **on top of** `plan3.2.md` (as corrected by `plan3.2_review.md`): re-aim the platform at Tier 2 applicant training/interview-prep/OJT, evolve the console into an applicant-centric Support Engineering Lab with human-primary investigation, add the environment/data-import/configurable-logic slice, and future-proof for enterprise identity (Amendment 3) without building SSO in the MVP.

## Success Criteria

- plan3.2 baseline holds: root `a1` module, `bchat/` untouched, three processes, local-first Postgres, overtime-only oracle, INC-001–003, read-only CaseTrace core — with all five review corrections applied.
- Web app on `localhost` delivers the applicant journey: guided labs, SQL/API/log investigation workspace, practice/assessment/interview modes, evidence-based feedback, and shareable portfolio records honestly labelled as simulated experience.
- Agents serve the applicant (customer, ticket, mentor, interviewer, evaluator) plus a CaseTrace benchmark for human-vs-AI comparison; assessment mode provably leaks no hidden answers.
- Users can create/configure a workspace, load their own CSV/JSON dataset with mapping preview and validation, and reset reproducibly; every environment carries a version manifest (modules, schemas, rules, dataset, config, scenario state).
- Identity is abstracted (`Principal` decoupled from provider), local auth works with zero external credentials, the extended role model exists in domain + authorization design, and the Amendment 3 roadmap/docs are published with SSO explicitly marked planned, not stubbed.

## Approach

**Authority stack.** plan3.2 stays the execution baseline; plan3.2_review's five corrections are folded in (below); plan3.3 governs product direction, phase extensions (8–9), and Amendment 3 future-proofing. Where plan3.3's phase table renames plan3.2 phases, the plan3.3 names win (Phase 5 = training/assessment/interview; Phase 6 = console + devX).

**Review corrections applied:** (1) oracle independence = separate implementation + independent derivation + golden fixtures + no PayLoop calc-package dependency + separate human review (import test is a guardrail, not proof); (2) core/module boundary proven by a Go dependency-graph test + registration test, terminology grep demoted to heuristic; (3) cross-interface consistency = same scenario executed against separate identical fixtures, comparing normalized business state, integration events, and evidence hashes — never triple-mutation of shared state; (4) Fly remote health check conditional on explicit non-prod deploy authorisation; local Docker + isolated-Neon migration tests mandatory; (5) Go version chosen deliberately in Phase 0 (installed 1.25.7 vs bchat's 1.26 directive).

**Reconciliations:** the Operations Console evolves into the Support Engineering Lab (same stack: Go templates + local HTMX/CSS, no CDN/npm) with applicant areas — workspace, incident labs, SQL workbench, logs/APIs, AI mentor, mock interviews, skill assessment; P0/P1/P2 ordering from plan3.3 §6 sequences work but the tiny order-fulfillment module stays mandatory (review-frozen extensibility proof); Phase 8's full scope (PG-export import, anonymization pipeline, Level A rules engine, Level B isolated Go modules, snapshots/cloning) stays post-MVP except the pulled-forward slice (workspace + CSV/JSON + reset + manifest); per-environment DB/container isolation is designed for, not built — MVP ships one reproducible local environment.

## Steps

1. **Phase 0 — Audit:** per plan3.2 + deliberate Go-version decision, bchat auth/logging reuse check for the identity boundary, record missing LICENSE. Gate: audit docs + frozen toolchain/module decisions.
2. **Phase 1 — Platform + contracts:** plan3.2 artifacts plus applicant-journey designs (lab modes, mentor/interviewer/evaluator contracts, portfolio schema), Environment Manager + manifest + workspace-isolation design, import/validation/preview design, Level A/B extension design, Amendment 3 docs (capability roadmap, identity/access architecture, provider boundary, authorization model incl. learner/auditor/service roles, environment isolation, future identity lab). Gate: design review; license approved before applying.
3. **Phase 2 — Reference slice + oracle:** unchanged scope, correction (1) enforced. Gate: fixed-seed replay + reconciliation + indeterminate handling.
4. **Phase 3 — Sync + incidents:** unchanged scope + mode-neutral ticket model (tickets servable in practice/assessment/interview). Gate: INC-001–003 reproducible with stable evidence.
5. **Phase 4 — CaseTrace + applicant agents:** generic investigation core; SG tools; customer/ticket/mentor/interviewer/evaluator roles; mock-LLM default; CaseTrace benchmark runner. Gate: read-only boundaries hold; mentor hints and interviewer follow-ups never reveal ground truth (sealed-evaluator tests).
6. **Phase 5 — Training + assessment:** three modes, scoring against hidden ground truth, feedback reports (missed evidence, wrong hypotheses), portfolio + progress records, human-vs-agent comparison. Gate: assessment-mode leak tests green; portfolio export reproducible without an LLM session.
7. **Phase 6 — Lab console + OSS readiness:** applicant UI areas on real data, SQL workbench over allow-listed views with limits, OpenAPI + interface coverage matrix + correction-(3) consistency tests, fresh-clone onboarding. Gate: credential-free `task demo`, journeys green.
8. **Phase 7 — Extensibility proof:** minimal order-fulfillment module + dep-graph/registration tests + SDK guide. Gate: unmodified core runs both modules.
9. **Phase 8 — Environments (MVP slice first):** workspace CRUD/configure, CSV/JSON import with mapping preview/error reports/integrity validation, reproducible reset, manifest v0; then PG-export import, anonymization pipeline (design must consult `plans/Guide to Basic Anonymisation…pdf`; raw imports never leave the authorized env or auto-submit to external LLMs), Level A configurable rules, Level B isolated module boundary, snapshots. Gate: imported dataset reproduces identical env from manifest; isolation tests pass.
10. **Phase 9 + Future A–D (roadmap):** private staging/customer connectors/commercial deployment and the Amendment 3 enterprise track (identity lab, RBAC/ABAC, promotion/backup-DR) documented with boundaries; nothing beyond design spikes without a customer driver.

## Validation Plan

- Every phase: `go build ./...`, `go test ./...`, `go vet ./...`, plus boundary tests (dep-graph, oracle-import, role denials incl. new learner/auditor roles), mode-separation tests (assessment/interview transcripts scanned for ground-truth leakage), import round-trip + manifest-replay tests, and correction-(3) normalized cross-interface comparisons.
- Highest-risk check: end-to-end applicant loop on a seeded incident — ticket → SQL/log/API investigation → submitted diagnosis → score + feedback + portfolio — with evidence hashes stable across replay and modes; if scoring/feedback is nondeterministic or leaky, Phases 6–9 rest on sand.
- Release check: fresh-clone credential-free onboarding; isolated-Neon migration test; Fly remote check only with explicit authorisation; no stub SSO/OIDC/SAML/SCIM endpoints anywhere.

## Risks / Open Questions

- **Q1:** which Fly app/region is the non-prod target, and who authorises the first deploy? (Remote check stays conditional until answered.)
- **Q2:** is a Singapore payroll practitioner available for the overtime rule-pack review, or do I proceed with source-pinned golden tests + independent code review?
- **Q3:** license decision — MIT or Apache-2.0? Nothing is applied before your explicit approval (required before any public release).
- Risks: applicant-mode answer leakage (sealed-evaluator + leak-scan tests); anonymization overclaim (PDF-grounded design, honest capability labels); Phase 8 scope swallowing the MVP (pulled-forward slice is the fence; remainder gated behind Phase 7); future-enterprise gold-plating (Amendment 3 acceptance criteria forbid stubs and unneeded abstractions).
