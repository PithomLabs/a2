# WorkforceOps Lab — Minimum Viable Product Plan

**Version:** 0.1  
**Date:** 8 November 2026  
**Status:** Proposed implementation baseline for coding-agent design and delivery  
**Primary stack:** Go · Neon PostgreSQL · Fly.io  
**Working product name:** WorkforceOps Lab  
**Fictional employer:** Meridian Workforce Pte. Ltd.  
**Simulated systems:** ShiftHub (HRIS/WFM) and PayLoop (payroll)  
**Investigation engine:** CaseTrace (agent-driven Tier 2 investigation)

---

## 1. Executive decision

Build a small but functioning Singapore workforce-operations environment, not a mock dashboard and not a complete HR/payroll product.

The environment will model one fictional employer and two logically independent business systems:

1. **ShiftHub** owns employee/work-contract reference data, shift schedules, clock events, attendance, approved timesheets and overtime approvals.
2. **PayLoop** consumes approved employee/time/pay inputs, runs payroll, calculates the supported statutory items, and produces itemised payslips and payroll exports.
3. **The integration layer** transfers data between ShiftHub and PayLoop through APIs/batches, with mappings, cut-off times, idempotency, retries, partial failures and a dead-letter path.
4. **The reference oracle** independently calculates expected outcomes from source facts, applicable rule versions, employment terms and explicit employer policies. It is for fixture validation and evaluation only; the investigator must never see its answer key.
5. **CaseTrace** investigates tickets using the same restricted evidence surfaces available to a human Tier 2 engineer. It records hypotheses, evidence, claims, findings and escalation decisions.

The MVP's purpose is to create a credible, reproducible environment for testing and training technical support investigations without customer production credentials. It is not a certified payroll/compliance product and does not prove market demand by itself.

**MVP proof:** a clean baseline processes correctly; controlled integration/data defects produce believable symptoms; at least three incidents can be reproduced from a fixed seed; and a human or agent can diagnose them using evidence collected through read-only tools.

## 2. Consolidated review decisions

The attached reviews converge on the following elements, which are adopted:

- Public Singapore rules provide a source-backed basis for deterministic business calculations and expected-behavior tests.
- A synthetic but working enterprise is more valuable than disconnected mock records or fabricated tickets.
- Keep the initial business scope at the integration boundary between workforce/time data and payroll.
- Build an independent reference oracle; do not let PayLoop validate its own calculations.
- Inject **fault mechanisms**, not merely prewritten symptoms. Let the ticket symptoms emerge from the faulty business process.
- Keep the incident ground truth hidden from the investigating agent.
- Store an evidence ledger and score the investigation, rather than judging the quality of a chat transcript.
- Use deterministic seeds and reproducible runs.
- Start with one agent/orchestrator and deterministic checks, not an unconstrained multi-agent swarm.
- Keep the agent read-only. Recommendations and any corrective action require human approval.
- Evaluate root-cause accuracy, evidence quality, hypothesis discipline, unsupported claims, escalation, time and model cost.

The following proposals are deliberately narrowed for an MVP:

- **Two logical systems, not four or six independently operated applications.** ShiftHub and PayLoop have separate API and data boundaries. The sync worker is a separate process. They can be deployed separately from one Go monorepo when needed.
- **One Neon project initially**, with separate schemas and database roles. Cross-system communication goes through the integration API, not direct table access by the other application.
- **No large front-end dashboard in the first build.** Start with APIs, deterministic command-line tools, JSON records and Markdown reports. Add a minimal read-only demonstration console only after the investigation path works.
- **One narrow vertical slice**, not a complete HRIS/WFM/payroll suite. Implement only the workflows needed to demonstrate employee/time data flowing into payroll and a small, verified set of Singapore rules.
- **CPF is a gated implementation.** Its full tables depend on citizenship status, SPR year, age band, wage band, wage ceilings and rounding. Do not copy one summary rate table and call it complete. Implement only after the exact current official tables and test cases are pinned and independently reviewed.
- **Neon branching is optional, not a core dependency.** Fixed seeds, scenario-run isolation and repeatable resets are sufficient for the first release.
- **No random fault injection during acceptance tests.** Faults are explicitly selected, seeded and reproducible. Randomized chaos testing can be a later feature.

The reviews' wider ideas—IRAS/AIS and IR21, full leave administration, public-holiday/rest-day pay, bank disbursement, multiple external statutory endpoints, hundreds of employees, multiple Fly applications, and a complete multi-agent framework—are deferred. They are expansion candidates, not MVP requirements.

## 3. Product objective and non-objectives

### 3.1 Objective

Create a programmable, synthetic Singapore workforce/payroll environment in which a support engineer must trace a customer-visible discrepancy across business records, APIs, integration jobs, configuration and logs.

The first commercial/technical proposition is:

> Given a realistic workforce-to-payroll problem, can a human or AI Tier 2 investigator establish what happened, show the evidence, distinguish a business-rule issue from a data/configuration/integration defect, and recommend a safe next action?

### 3.2 Non-objectives

Do not build these in v0.1:

- A production-ready HRIS or payroll SaaS for real employers.
- A legally certified compliance calculator or actual CPF/IRAS submission service.
- Full Singapore employment-law coverage.
- Full leave, public-holiday, rest-day, shift-premium, tax-clearance or bank-payment workflows.
- Real employee data, real NRIC/FIN numbers, bank details or customer production access.
- Autonomous writes, employee corrections, payroll adjustments or external submissions.
- A generic agent-builder platform, vector database or multi-agent swarm.
- A complex frontend, mobile application or microservice platform.

The simulator must be labelled clearly as synthetic and not suitable for processing real payroll.

## 4. Fictional business and bounded domain

### 4.1 Initial fixture

Use one fictional Singapore employer: **Meridian Workforce Pte. Ltd.** This name is illustrative only.

Initial fixture target:

- 1 tenant / 1 legal entity.
- 2 operating locations and 3 departments.
- 50–100 synthetic employees.
- 3–6 months of historical activity.
- Monthly payroll periods.
- Monthly-rated employees in the first calculation slice.
- A mix of workman, non-workman and manager/executive classifications, with applicability explicit.
- Singapore Citizens and selected Singapore Permanent Resident cases for CPF tests only when the applicable rate table is fully implemented.

These are seed parameters, not claims about a typical Singapore employer. Expand volume only when performance or scenario realism requires it.

All personal and company identities must be synthetic. Do not generate valid-looking NRIC/FIN credentials or real bank account numbers. Use clearly artificial identifiers such as `EMP-SG-000142`.

### 4.2 ShiftHub scope

ShiftHub is the source system for:

- Synthetic employee identifier and employment status.
- Employment classification and effective-dated contract terms needed by the slice.
- Department and work location.
- Shift assignment and roster.
- Raw clock-in/clock-out events and break intervals.
- Attendance interpretation and corrections.
- Timesheet preparation and approval.
- Overtime approval and source pay-code/event classification.

Do not build a complete HR lifecycle. Only represent the employee and contract facts needed by the supported scenarios.

### 4.3 PayLoop scope

PayLoop is the downstream payroll system for:

- Imported employee and approved time/pay inputs.
- Pay periods and payroll cut-off configuration.
- Pay-code and field-mapping configuration.
- Payroll run state and calculation lines.
- Supported basic salary and overtime calculations.
- Itemised payslip output.
- Payroll export/batch state and reconciliation records.
- Supported CPF calculations only after the CPF rule pack passes its own approval gate.

PayLoop must retain the imported source identifiers and correlation IDs so every payroll line can be traced back to its source.

### 4.4 Integration boundary

The integration layer is a first-class part of the product, not a direct database copy. It must model:

- Scheduled/batch and API-based transfer.
- Export/import payloads with schema versions.
- Mapping tables and filters.
- Cut-off timestamps and effective dates.
- Idempotency keys and duplicate detection.
- Retry attempts, timeouts and backoff.
- Partial batch success.
- Outbox records and delivery attempts.
- Dead-letter state and reconciliation.
- Correlation IDs across the entire transaction.

No Kafka, Redis or distributed queue is required in v0.1. A PostgreSQL transactional outbox and a Go worker are sufficient.

## 5. Singapore rulebook: source-backed and versioned

### 5.1 Rule provenance model

Keep four sources of behavior separate:

1. **Statutory rule:** legislation or official government guidance.
2. **Employment term:** the synthetic employee's contract and effective-dated terms.
3. **Employer policy/configuration:** roster policies, approvals, cut-offs, pay-code mappings and permitted benefits.
4. **Simulated defect:** an intentionally faulty record, configuration or processing behavior.

A statutory requirement must never be inferred from an employer policy, and an employer policy must not be described as law.

Every implemented statutory rule must have:

- Rule ID and jurisdiction.
- Responsible authority and official source URL.
- Exact source section/table and retrieval date.
- Source document/page hash where a downloadable source is used.
- Effective-from/effective-until dates.
- Applicable employee population and exclusions.
- Required input fields and calculation/decision logic.
- Rounding and boundary behavior.
- Golden test cases and edge-case tests.
- Review status: `unverified`, `reviewed`, `approved`, `superseded`.

Do not allow an LLM to write or modify executable statutory rules without human review. The Go implementation is deterministic; the agent only investigates and explains its output.

### 5.2 MVP rule scope

**MOM / Employment Act overtime — first rule pack.** The current MOM guidance states that Part IV does not cover managers and executives; overtime entitlement applies to non-workmen with monthly basic salary up to S$2,600 and workmen up to S$4,500. The guidance defines overtime as work beyond normal hours excluding breaks, gives a minimum rate of 1.5 times the hourly basic rate, provides the monthly-rated hourly basic-rate formula `(12 × monthly basic pay) / (52 × 44)`, caps the non-workman overtime rate at the S$2,600 salary level/S$13.60 hourly rate, and states a 72-hour monthly overtime limit (subject to the detailed rules/exemptions). Overtime pay is due within 14 days after the last day of the salary period.

Do not reduce this to “every employee after 44 hours gets 1.5x.” Applicability, classification, contractual hours, break periods, work arrangements, exemptions and exceptions must be represented. The MVP should support a clearly bounded subset and return `needs_review` for unsupported arrangements rather than guessing.

**CPF — second, gated rule pack.** The CPF Board's current guidance explains that the applicable table depends on citizenship status, SPR year, age and wage band; Ordinary Wages and Additional Wages have different ceilings; and contribution rounding rules apply. The Ordinary Wage ceiling is S$8,000 from 1 January 2026. Further rate increases for senior employees are scheduled from 1 January 2027. Implement the complete relevant tables from the official rate-table document for the declared supported population and relevant dates. Do not hardcode only the headline rates for Singapore Citizens/third-year SPRs and assume they cover all SPR cases.

**Itemised payslip structure.** For the covered employee population, generate payslips with the relevant required fields, including employer/employee identity labels (synthetic), payment date, salary period, basic salary, allowances/additional payments, deductions, overtime hours and pay where applicable, and net salary. Store the output and its source payroll-line IDs.

### 5.3 Deferred rule packs

Defer full annual/sick/childcare leave entitlement calculations, public holiday and rest-day payment calculations, IRAS AIS/IR8A/IR21 workflows, self-help-group contributions, SDL edge cases, work-pass administration and tax-clearance logic.

They may be represented as synthetic metadata or future data domains, but must not be presented as legally implemented functionality in v0.1.

### 5.4 Rule-validation gate

Before any statutory calculator is used to generate a scenario answer key:

1. Retrieve the current official source and exact rate/eligibility tables.
2. Record applicability and effective dates.
3. Derive golden test cases, including boundary cases.
4. Compare against official worked examples/calculators where available.
5. Obtain a Singapore payroll practitioner review if available.
6. Have a separate reviewer inspect the implementation and tests.
7. Mark the rule pack approved only after discrepancies are resolved or explicitly carried as limitations.

If this gate is not passed, the platform may still simulate a clearly labelled configuration/integration defect, but must not claim to have independently verified legal payroll correctness for that case.

## 6. Independent reference oracle

The oracle is one of the most important design decisions.

It must be implemented separately from PayLoop's production calculation path. It must not import PayLoop calculation functions or share the same calculation routine. Otherwise a common implementation bug could make both agree incorrectly.

Inputs:

- Canonical ShiftHub source facts (approved punches/timesheets and employee/contract data).
- Versioned statutory rule pack.
- Effective-dated employment terms.
- Explicit employer policy/configuration, where relevant.
- Payroll period and cut-off.

Outputs:

- Expected payable hours and supported pay components.
- Expected statutory contributions for the approved CPF subset.
- Expected payslip lines and totals.
- Applicability status and unresolved/unsupported conditions.
- A reproducible input/output hash and calculation version.

The oracle must distinguish `calculated`, `not_applicable`, and `indeterminate`. It must not force a result where the rule pack or facts do not support one.

For clean baseline data, PayLoop results should reconcile with the oracle within explicitly defined rounding rules. After a defect is injected, the oracle provides the expected result and the discrepancy against PayLoop becomes part of the hidden ground truth.

The oracle is evaluator-only. No investigator tool, API response or agent prompt may expose its expected result or ground-truth record.

## 7. Synthetic data and reproducibility

Build a deterministic business-data generator, not a random row generator.

Generation order:

1. Company, sites and departments.
2. Employees and employment status history.
3. Effective-dated contracts and compensation.
4. Shift templates and rosters.
5. Raw punches and break events.
6. Attendance interpretation and corrections.
7. Timesheet approval and overtime classification.
8. Payroll inputs, runs and payslips.
9. Integration payloads, delivery attempts and logs.

Requirements:

- Fixed seed and explicit generator version.
- Referential integrity and valid temporal relationships.
- Synthetic identities only.
- Normal baseline has no injected faults and reconciles against the oracle.
- Fault injection is a separate operation from data generation.
- Every run has a `scenario_run_id`, seed, fixture version and source rule-pack versions.
- Store event instants as timezone-aware timestamps (UTC) while retaining source offset/timezone metadata; derive workforce business dates and payroll cut-offs using `Asia/Singapore`. Do not compare naive timestamps or silently reinterpret local clock events.
- Same seed + same versions + same scenario must reproduce the same business facts and fault mechanism.

Initial data volume: 50–100 employees and 3–6 months of history. Increase only when an investigation scenario requires larger volumes, pagination, query performance or tenant isolation tests.

## 8. Fault injection and scenario catalogue

### 8.1 Mechanism-based faults

Do not hardcode “ticket 7 has missing overtime.” Define reusable defect mechanisms and parameters. The mechanism should create the inconsistent system state; the ticket should describe only what a customer could observe.

Each mechanism should specify:

- Injection point and permitted mutation.
- Trigger and parameters.
- Prerequisites.
- Observable symptoms.
- Expected causal chain.
- Minimum evidence that proves the root cause.
- Plausible red herrings.
- Correct resolution/recovery recommendation.
- Difficulty tier and scoring rubric.
- Cleanup/reset procedure.

### 8.2 First three required incidents

**INC-001 — Approved overtime missing from payroll**  
A source timesheet contains approved overtime, but the integration mapping/filter or importer excludes the relevant pay code. Symptoms appear in PayLoop/payslip. Investigator must trace approved time → export payload → mapping/filter → import result → payroll line/payslip. The legal overtime calculation is only used if the employee is within the approved rule pack's coverage.

**INC-002 — Attendance hours truncated during synchronization**  
Raw clock events and approved source hours differ from the downstream imported time total because of a reusable defect such as dedup-key collision, event ordering or partial-batch handling. Investigator must distinguish missing source events from lost/duplicated events in the connector.

**INC-003 — Employee missing from payroll after HR update**  
An effective-dated employee or employment update is received by ShiftHub but not correctly applied in PayLoop due to an external-ID mapping issue, cut-off ordering or dead-lettered sync. Investigator traces master record → event/payload → delivery attempts → PayLoop import/status.

A CPF rate/version discrepancy may be added as the next scenario only after the CPF rule pack has passed the validation gate. Do not use an unverified CPF calculation as the answer key.

### 8.3 Hidden ground truth

Store expected root cause and scoring data separately from investigator-accessible schemas and credentials. At minimum it contains:

- Root-cause mechanism and parameters.
- Affected entity IDs and business period.
- Causal chain.
- Required evidence IDs or evidence predicates.
- Red herrings and rejected hypotheses.
- Expected business impact.
- Correct resolution recommendation.
- Acceptable alternative conclusions and escalation conditions.

The agent must not be able to query the hidden answer key, oracle output or scenario authoring metadata.

## 9. CaseTrace investigation workflow

Start with one orchestrator and a bounded loop. Do not build an agent swarm in v0.1.

1. **Triage:** parse the ticket, identify affected entity, period, customer impact and known symptoms.
2. **Hypotheses:** record candidate causes, a falsifier/cheapest test and what evidence would support or reject each.
3. **Evidence collection:** query permitted records, APIs and logs through typed read-only tools.
4. **Analysis:** correlate source and downstream records using IDs, timestamps and correlation IDs.
5. **Verification:** check that each substantive claim is supported by reproducible evidence. Initially use deterministic checks and independent query replay; add a separate verifier model only if necessary.
6. **Verdict:** confirm a root cause, give a conditional diagnosis, or escalate because evidence is insufficient.
7. **Report:** produce a concise Markdown/JSON investigation record with findings, evidence IDs, rejected hypotheses, uncertainty and recommended next action.

### 9.1 Investigator tools

Initial tools should be typed and narrow, for example:

- `get_employee(employee_id)`
- `get_employment_terms(employee_id, at_date)`
- `get_roster(employee_id, date_range)`
- `get_clock_events(employee_id, date_range)`
- `get_attendance(employee_id, date_range)`
- `get_timesheet(employee_id, pay_period)`
- `get_payroll_run(pay_period)`
- `get_payroll_lines(employee_id, pay_period)`
- `get_payslip(employee_id, pay_period)`
- `get_export_batch(batch_id)`
- `get_integration_job(job_id_or_correlation_id)`
- `get_delivery_attempts(correlation_id)`
- `get_application_events(correlation_id)`
- `execute_approved_query(query_id, parameters)`

The first release should use approved query templates and read-only evidence endpoints. Do not give an LLM unrestricted SQL or production credentials. A constrained read-only SQL console over curated views can be added for the human training mode after the safe evidence surface is complete.

### 9.2 Evidence ledger

Every investigation must produce a durable record. Suggested logical entities:

- `scenario` and `scenario_run`
- `incident_ticket`
- `agent_run`
- `hypothesis`
- `evidence`
- `claim`
- `verdict`
- `tool_execution`

Each evidence record stores source type, tool/query and parameters, timestamp, correlation ID, row count, content/result hash and the scenario snapshot/version. Claims link to one or more evidence IDs. A deterministic verifier checks that referenced evidence exists and that its recorded hash matches the immutable run artifact or can be reproduced against the same frozen scenario state.

The accepted investigation output is the structured record and its evidence—not the model's private reasoning or a conversational transcript.

### 9.3 Safety boundaries

- All investigation tools are read-only.
- No update/delete/insert, payroll adjustment, retry-trigger, credential rotation or external submission from the agent.
- Enforce tenant and service boundaries in Go and database roles, not in prompts alone.
- Apply query timeouts, row limits and allow-listed views/query templates.
- Record every tool invocation and result hash.
- Human approval is required for any future corrective operation.
- The agent must state `insufficient_evidence` or `escalate` when the evidence is incomplete.

## 10. Evaluation: score evidence, not eloquence

Every scenario must be run against hidden ground truth. Capture:

- **Root-cause accuracy:** did it identify the fault mechanism, not merely the symptom?
- **Evidence precision:** how much cited evidence is relevant and necessary?
- **Evidence recall:** did it retrieve the evidence required to establish the cause?
- **Hypothesis discipline:** did it test and reject plausible alternatives appropriately?
- **Unsupported-claim rate:** claims with no evidence or evidence that does not support the statement.
- **Escalation appropriateness:** did it stop when evidence was insufficient and avoid unnecessary escalation when sufficient?
- **Safety compliance:** did it remain within read-only permissions and avoid prohibited actions?
- **Efficiency:** time to first evidence-backed diagnosis, tool calls and token/model cost per run.

Run a human baseline on the same tickets where possible. If no external Tier 2 practitioner is available initially, conduct a timed manual investigation and label it as a founder baseline, not an industry benchmark.

Keep at least one fault mechanism or parameter family held out from agent tuning. This helps reveal whether the agent learned to investigate or merely memorized the injector's signature.

Do not publish a success rate until the scenario set, scoring rubric, baseline and held-out design are documented. Synthetic test performance is not proof of performance on customer production systems.

## 11. Technical architecture and repository shape

### 11.1 Architecture decision

Use one Go monorepo with explicit service boundaries and three runnable processes:

- `shifthub-api`
- `payloop-api`
- `sync-worker`

Also provide non-production commands for seeding, oracle/evaluation and scenario execution. The two APIs should have separate database roles and should not read each other's schemas. The integration worker communicates through their APIs. This gives credible access and integration boundaries without the operational cost of a fleet of microservices.

Deploy the APIs and worker as separate Fly.io processes/apps from the same repository when required. Keep them in one region initially. Do not create four or six independently maintained applications for v0.1.

### 11.2 Neon layout

Use one Neon PostgreSQL project initially, with logically separated schemas and database roles, for example:

- `shifthub` schema / `shifthub_app` role.
- `payloop` schema / `payloop_app` role.
- `platform` schema for job/outbox/operational coordination with narrow role grants.
- `investigation` schema for tickets, evidence and agent runs.
- `evaluation_private` schema for hidden ground truth and oracle results, accessible only to the evaluator role.

The worker receives only the permissions needed to read/export from ShiftHub and submit to PayLoop. Prefer the service APIs for cross-system operations; do not grant broad cross-schema access just for convenience. The investigator receives only curated read-only views or service APIs. The agent has no credentials for `evaluation_private`.

Use PostgreSQL transactions, unique constraints, idempotency keys and a transactional outbox. Keep append-only audit/event history where it is useful; do not implement full event sourcing across the application.

### 11.3 Go stack

- Go current stable version pinned in `go.mod` and CI.
- `net/http` or Chi for HTTP routing.
- `pgx` and `sqlc` for PostgreSQL access.
- Goose or an equivalent migration tool.
- Structured JSON logs with request/correlation IDs.
- OpenTelemetry-compatible tracing and metrics where practical.
- PostgreSQL-backed durable job/outbox processing.
- Provider-neutral interface for the LLM API.
- No frontend framework requirement. Use API, CLI and Markdown/JSON artifacts first.

Illustrative repository layout:

```text
workforceops-lab/
  cmd/
    shifthub-api/
    payloop-api/
    sync-worker/
    seed/
    oracle/
    evaluate/
    scenario/
  internal/
    shifthub/
    payloop/
    integration/
    rulebook/
      singapore/
    oracle/
    simulation/
    fault/
    investigation/
    evidence/
    evaluation/
    audit/
    platform/
  db/
    migrations/
    queries/
  scenarios/
    definitions/
  docs/
    architecture/
    rules/
    decisions/
  tests/
```

This is a proposed logical layout, not a mandate to create empty packages. Keep packages cohesive and avoid dependency cycles. The oracle must not import PayLoop calculation code.

## 12. Delivery sequence and phase gates

Use the following order. A six-week target is a planning aid, not a guarantee. Do not sacrifice rule correctness or reproducibility to meet a calendar date.

### Phase 0 — Design and regulatory source lock

Deliver:

- System context and component diagram.
- Domain glossary and supported workflows.
- ERD and database role/schema plan.
- API contracts for ShiftHub, PayLoop and sync worker.
- Rule source register and applicability matrix.
- Oracle design and independent test plan.
- Three scenario definitions and hidden-ground-truth format.
- ADRs for service boundaries, database isolation, outbox and scenario isolation.

**Gate:** no unresolved architectural choice that changes the MVP boundary; no unverified rule is marked implemented. Start coding after this design package is internally reviewed.

### Phase 1 — Clean business path and reference oracle

Implement the smallest clean path: synthetic employee/contract → roster → clock events → attendance → approved timesheet → import → payroll line → payslip.

Build the independent reference calculation for the supported subset and golden tests. Run the generator with a fixed seed and prove the clean baseline reconciles.

**Gate:** same seed and versions reproduce the same baseline; oracle and PayLoop reconcile for supported cases; unsupported rule cases are marked indeterminate rather than guessed.

### Phase 2 — Integration mechanics

Implement the scheduled/batch or API transfer, mappings, cut-off, idempotency, retries, partial failure, correlation IDs, outbox and dead-letter handling.

**Gate:** normal transfer works; a deliberately failed transfer is observable end-to-end; retries do not create duplicate business records.

### Phase 3 — Fault mechanisms and first three incidents

Implement INC-001 through INC-003 as mechanism-based injectors. Store hidden ground truth separately. Add red herrings and at least one plausible but incorrect hypothesis per incident.

**Gate:** each incident can be recreated from a clean fixture with a fixed seed; its symptoms and evidence are stable; reset returns the scenario to baseline.

### Phase 4 — Evidence surface and investigation agent

Implement read-only tools, evidence ledger, one bounded agent loop, deterministic claim/evidence verification and report generation. Start with one agent and add model-based verification only if deterministic checks are insufficient.

**Gate:** no tool can mutate business data; all substantive claims cite evidence IDs; the agent cannot access hidden ground truth or oracle outputs.

### Phase 5 — Evaluation and demo

Run all scenarios; record root cause/evidence/unsupported-claim/safety/time/cost metrics; conduct a human baseline; add a held-out fault test. Prepare a reproducible demo run and Markdown investigation report.

**Gate:** disclose limitations; do not present synthetic pass rates as customer-production performance; document failures and next work before broadening scope.

## 13. Definition of MVP done

The MVP is done only when all of the following are true:

1. ShiftHub creates valid synthetic employees, contracts, schedules, punches, approved timesheets and overtime records for the supported slice.
2. PayLoop imports source records through the integration boundary and produces payroll lines and itemised payslips.
3. The independent oracle validates the clean baseline for the supported employee/rule population.
4. Statutory rules have source citations, effective dates, applicability and golden tests; unsupported cases are explicitly excluded.
5. The fixed-seed generator is reproducible and contains no real personal data.
6. Three fault mechanisms produce realistic, reproducible incidents with hidden ground truth.
7. An investigator can inspect relevant records, APIs and logs through read-only tools.
8. The investigation record links hypotheses and claims to evidence IDs and hashes.
9. The deterministic verifier can validate the evidence references and scenario state.
10. At least one human baseline and one held-out test are run and reported.
11. The API and worker deploy to Fly.io and use Neon with restricted roles.
12. Tests, migrations, seed/reset instructions and a one-command local demo are documented.

A polished dashboard, additional legal domains, more tenants, real external government integration and automatic remediation are not part of this definition. A human candidate assessment/training mode is a natural next product because it can reuse the same scenarios and evidence surfaces, but it is not required to complete the first MVP.

## 14. Principal risks and controls

| Risk | Control |
|---|---|
| Incorrect interpretation of Singapore rules | Pin official sources, applicability, effective dates, golden tests and practitioner review. Unsupported cases return indeterminate. |
| Oracle shares a bug with PayLoop | Separate implementation and tests; no imported calculation functions. |
| Agent memorizes fault signatures | Mechanism-based injectors, red herrings, parameter variation and held-out fault family. |
| Agent accesses answer key | Separate schema/role and tool surface; no ground-truth credentials in agent process. |
| Agent damages data | Read-only tools, database permissions, query limits and human approval for any future write. |
| Synthetic data looks realistic but is inconsistent | Deterministic generation, referential invariants, clean baseline reconciliation and audit trails. |
| Scope expands into full payroll product | Enforce explicit in-scope/deferred list and phase gates. |
| Demo is mistaken for proof of customer demand | Separate technical validation from willingness-to-pay validation; run paid pilots when a buyer is available. |

## 15. Coding-agent execution contract

The coding agent should execute the following instructions:

1. Inspect the existing repository and report its current state before changing files. Preserve useful existing code and conventions.
2. Treat this plan as the scope baseline. Do not silently add domains, services, frontend frameworks, databases or infrastructure.
3. Produce the Phase 0 design artifacts first: architecture, domain model/ERD, API contracts, rule-source register, oracle plan, scenario contract and ADRs.
4. Verify current official Singapore sources before implementing any statutory rule. Record exact source locators and effective dates. Do not guess missing values or copy summary rates without the complete applicable table.
5. Mark uncertain or unsupported regulatory behavior `BLOCKED` or `INDETERMINATE`; do not silently fill gaps with model knowledge.
6. Implement one vertical slice at a time, with tests and acceptance evidence for each phase. Do not generate a large codebase before proving the clean end-to-end path.
7. Keep PayLoop's calculation path and the reference oracle independent. Add tests that would fail if the same calculation routine were accidentally shared.
8. Keep hidden ground truth and oracle results inaccessible to the investigator.
9. Maintain deterministic seeds, reset procedures, correlation IDs, immutable evidence artifacts and reproducible scenario runs.
10. Keep the agent read-only. Enforce restrictions in Go and PostgreSQL, not prompt instructions alone.
11. Do not claim that passing synthetic scenarios proves production readiness, legal compliance certification or customer demand.
12. After each phase, report: files changed, tests run and results, supported behavior, known limitations, unresolved blockers and whether the phase gate passes. Do not claim a test passed unless it was actually executed.
13. Do not enter endless review loops. One focused independent review and one targeted correction pass per major phase are sufficient unless a new blocker appears.
14. Ask for director input only when a decision would materially change the product scope, a source conflict cannot be resolved, a required external permission is unavailable, or a safety/legal interpretation requires a qualified human. Otherwise record uncertainty and continue unaffected work.

### Required initial deliverables from the coding agent

Before implementation:

- `docs/architecture.md`
- `docs/domain-model.md`
- `docs/api-contracts.md`
- `docs/rule-source-register.md`
- `docs/oracle-design.md`
- `docs/scenario-format.md`
- `docs/decisions/ADR-*.md`
- A phase-by-phase implementation checklist with test gates.

Then implement Phase 1 only. Do not start the agent loop until the clean ShiftHub → integration → PayLoop → oracle path is working.

---

## 16. Initial official source register

The following official pages were checked on 8 November 2026. The coding agent must re-open them and record the exact applicable sections/tables before translating them into executable rules.

1. **MOM — Hours of work, overtime and rest days**  
   https://www.mom.gov.sg/employment-practices/hours-of-work-overtime-and-rest-days  
   Use for Part IV applicability, workman/non-workman salary thresholds, definition of overtime, hourly basic-rate calculation, overtime rate/cap, monthly overtime limit and payment timing. Do not implement rest-day/public-holiday pay in v0.1.

2. **CPF Board — How much CPF contributions to pay**  
   https://www.cpf.gov.sg/employer/employer-obligations/how-much-cpf-contributions-to-pay  
   Use for citizenship/SPR-year, age and wage-band tables, contribution computations and rounding. Retrieve the complete applicable rate-table PDF; do not rely only on the summary table on the webpage.

3. **CPF Board — Ordinary Wage ceiling**  
   https://www.cpf.gov.sg/service/article/what-is-the-ordinary-wage-ow-ceiling-mbr  
   Use for effective-dated OW ceiling history, including S$8,000 from 1 January 2026.

4. **MOM — Itemised payslips**  
   https://www.mom.gov.sg/employment-practices/salary/itemised-payslips  
   Use to define the fields included in the simulated payslip for employees covered by the Employment Act.

5. **MOM — Employment records**  
   https://www.mom.gov.sg/employment-practices/employment-records  
   Future use for retention and employment/salary record requirements.

6. **IRAS — Submit employment income records (AIS)**  
   https://www.iras.gov.sg/taxes/individual-income-tax/employers/auto-inclusion-scheme-(ais)-for-employment-income/submit-employment-income-records  
   Deferred from v0.1; retain as a future integration/source candidate.

7. **MOM — Annual leave eligibility and entitlement**  
   https://www.mom.gov.sg/employment-practices/leave/annual-leave/eligibility-and-entitlement  
   Deferred from v0.1; use only when leave becomes an approved scope item.

## 17. Final implementation principle

Build the smallest functioning business system that can produce a real discrepancy across a system boundary, preserve the history of how that discrepancy arose, and let a human or agent prove its cause using evidence.

The simulator generates operational reality. The fault mechanisms make it fail in controlled ways. The oracle establishes expected outcomes for supported cases. CaseTrace investigates. The evaluation harness measures whether the investigation was correct.

That is the MVP. Everything else is an expansion.
