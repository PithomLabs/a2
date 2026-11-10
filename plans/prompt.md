The key instruction to the coding agent is: reuse `bchat` as an infrastructure and Go-backend foundation, but replace its product. Do not build WorkforceOps as another feature inside the old chat/RAG product.

The public repository describes BChat as a Go/Echo application with Memos-derived functionality, React UI, RAG, vector storage and PostgreSQL deployment options. It also documents the Fly.io deployment profiles and PostgreSQL path that are worth inspecting and preserving.&#x20;

[image](https://www.google.com/s2/favicons?domain=https://github.com\&sz=32)

GitHub



I would keep the existing Go module and working Neon/Fly.io configuration wherever possible, while removing the old product's RAG, chat, widget and frontend capabilities. The prompt below makes the coding agent prove what is reusable before deleting anything.

# CODING AGENT MANDATE

# WorkforceOps Lab — bchat Backend Refactor, Design and MVP Implementation

**Project:** WorkforceOps Lab\
**Existing repository:** `https://github.com/PithomLabs/bchat`\
**Implementation baseline:** `WORKFORCEOPS_LAB_MVP_PLAN_v0.1.md` (attached)\
**Language:** Go\
**Database:** Neon PostgreSQL\
**Deployment:** Fly.io\
**Approach:** Backend-only, API-first, autonomous implementation\
**Status:** Authoritative execution instruction

---

# 1. Your assignment

Act as a senior Go backend architect, database architect, enterprise application engineer, integration engineer and autonomous coding agent.

Your assignment is to transform the existing `bchat` repository into the backend foundation of **WorkforceOps Lab**, a realistic, synthetic Singapore HR/WFM/payroll environment designed to simulate production problems investigated by Tier 2 application and technical support engineers.

We are building the new product from scratch, but we are NOT starting with an empty repository.

We already have a Go codebase with existing infrastructure and integrations involving:

- Go backend.
- Existing HTTP framework and application bootstrap.
- PostgreSQL database support.
- Neon connectivity/configuration.
- Fly.io deployment configuration and workflows.
- Existing environment configuration and operational scripts.
- Potentially reusable authentication, logging, database and LLM integration components.

Your first responsibility is to identify what genuinely works and preserve it.

Your second responsibility is to remove the old product functionality that has no place in WorkforceOps Lab.

Your third responsibility is to design and implement the new backend.

**Do not build WorkforceOps Lab as a feature of the old RAG chatbot. Replace the old application domain while retaining useful infrastructure.**

This is a backend-only project. There must be no React application, embedded chat widget, customer-facing chatbot interface or frontend framework in the MVP.

The deliverable is a functioning Go backend, supported by reproducible data, real business workflows, integration failures, investigation tools and automated tests.

---

# 2. Mandatory project baseline

Read the attached `WORKFORCEOPS_LAB_MVP_PLAN_v0.1.md` completely before making architectural decisions.

Treat it as the scope and product baseline.

Its main architecture is binding:

1. **ShiftHub:** source system for employee/work-contract reference data, schedules, clock events, attendance, approved timesheets and overtime approvals.
2. **PayLoop:** downstream payroll system that imports approved data and produces payroll calculations, payroll records and itemised payslips.
3. **Integration layer:** transfers data between ShiftHub and PayLoop through APIs/batches, mappings, cut-offs, idempotency, retries, partial failures and dead-letter handling.
4. **Reference oracle:** independently calculates expected results for supported cases. It is not accessible to the investigator.
5. **Incident simulator:** injects reusable fault mechanisms that generate realistic operational symptoms.
6. **CaseTrace:** investigates incidents using restricted, read-only tools and produces an evidence-backed report.

The MVP must demonstrate:

- A clean business-data baseline that processes correctly.
- A reproducible integration path.
- Controlled faults that produce realistic symptoms.
- Three reproducible initial incidents.
- An independent expected-results calculation.
- A human- or agent-driven investigation using evidence.
- A repeatable demonstration using one fixed seed.

Do not expand this into a complete commercial HRIS or payroll platform.

If this instruction and the attached plan appear to conflict, preserve the narrower MVP scope and document the conflict. Do not silently expand the scope.

---

# 3. Existing bchat repository: what to preserve and what to remove

## 3.1 First inspect, then modify

Before changing any files:

1. Inspect the Git repository, current branch, commit and working-tree state.
2. Inspect `AGENTS.md`, `go.mod`, build files, Taskfiles, Dockerfiles, Fly configuration, GitHub Actions and database migration structure.
3. Identify the actual working PostgreSQL/Neon connection path.
4. Identify the actual Fly.io deployment path.
5. Run the existing backend tests and build.
6. Record existing failures separately from failures introduced by this refactor.
7. Trace dependencies before removing shared code.

Do not assume that a documented feature works simply because it appears in the README.

Do not assume that a component is disposable simply because it originally supported the chat application.

Create a working branch for the refactor. Do not force-push, overwrite the main branch, or discard pre-existing uncommitted changes.

Do not print secrets, database URLs, API keys or Fly tokens into logs or documentation.

## 3.2 Preserve and reuse

Preserve the following wherever their actual implementations are functional and sufficiently decoupled:

- Go version and existing Go toolchain conventions.
- Existing HTTP framework. Prefer retaining Echo rather than introducing another framework without a demonstrated reason.
- Existing PostgreSQL/pgx connectivity and Neon configuration.
- Existing Fly.io deployment configuration, Docker build and release workflow.
- Existing structured logging and error-handling conventions.
- Existing database connection and transaction utilities where reusable.
- Existing configuration and environment validation.
- Existing health checks and operational scripts where reusable.
- Existing LLM provider integration only if it can be isolated cleanly from chat/RAG functionality.

Do not rewrite the existing deployment setup simply to make it look cleaner.

Before changing Fly.io configuration, explain how the existing application is built, configured, connected to Neon and deployed.

Preserve the working path and make the smallest necessary changes.

Keep the existing Go module path initially. Do not rename the repository or module unless explicitly authorized.

## 3.3 Remove from the new product

After completing the dependency inventory and migration plan, remove or disconnect the old application functionality:

- RAG pipeline.
- Embedding generation and embedding providers.
- Vector search and vector database implementations.
- LanceDB.
- Document chunking and indexing.
- Knowledge-base upload, parsing and reindexing.
- KB/POLICY/SCRIPT chat configuration.
- Chat sessions and chat-specific streaming.
- Embedded chat widget.
- React/TypeScript frontend and build pipeline.
- Chat-specific simulation and transcript analysis.
- Chat-specific observational memory.
- Old chatbot administration and configuration endpoints.
- Memos-specific features that do not directly support the new product.
- Unused chat-specific routes, migrations, dependencies and environment variables.

Remove obsolete npm dependencies and frontend build requirements from the default backend build.

Do not retain dead code, inactive feature flags or misleading configuration simply because they were present in bchat.

However, do not delete generic authentication, PostgreSQL, deployment, logging or integration functionality until you have established that it is either reusable or unnecessary.

Git history preserves the previous application. Do not preserve abandoned functionality in the new runtime.

## 3.4 No frontend

The product is API-first.

Initial interaction should be through:

- HTTP APIs.
- CLI commands.
- JSON responses.
- Structured logs.
- Markdown investigation reports.
- Automated tests.

Do not build a React application or UI dashboard.

A small operational web console can be considered after the backend MVP passes its acceptance gates. It is not part of this assignment.

---

# 4. Target architecture

Build a modular Go monorepo with explicit logical service boundaries.

Do not create a fleet of microservices.

The initial system should have three runnable application processes:

**A. ShiftHub API**

Owns the workforce source data and workflows.

**B. PayLoop API**

Owns imported payroll inputs, mappings, calculations and payslips.

**C. Sync worker**

Transfers data between the two systems and manages scheduled processing, retries, idempotency and reconciliation.

Provide additional CLI commands for:

- Database migration.
- Synthetic data generation.
- Reference-oracle execution.
- Scenario creation and reset.
- Fault injection.
- Investigation execution.
- Evaluation and reporting.

These can be subcommands of one Go CLI rather than independent deployed services.

## 4.1 Suggested repository structure

Use this as the initial architectural direction, adapting it to sound existing bchat conventions:

```
bchat/
├── cmd/
│   ├── shifthub-api/
│   ├── payloop-api/
│   ├── sync-worker/
│   └── workforceops/
│
├── internal/
│   ├── platform/
│   │   ├── config/
│   │   ├── database/
│   │   ├── logging/
│   │   ├── auth/
│   │   ├── health/
│   │   └── clock/
│   │
│   ├── shifthub/
│   │   ├── employee/
│   │   ├── employment/
│   │   ├── scheduling/
│   │   ├── attendance/
│   │   └── timesheet/
│   │
│   ├── payloop/
│   │   ├── employee/
│   │   ├── mapping/
│   │   ├── payroll/
│   │   └── payslip/
│   │
│   ├── singapore/
│   │   ├── rules/
│   │   ├── overtime/
│   │   └── sources/
│   │
│   ├── integration/
│   │   ├── client/
│   │   ├── outbox/
│   │   ├── delivery/
│   │   ├── retry/
│   │   └── reconciliation/
│   │
│   ├── oracle/
│   │   ├── overtime/
│   │   └── validation/
│   │
│   ├── simulation/
│   │   ├── generator/
│   │   ├── scenario/
│   │   └── fault/
│   │
│   ├── investigation/
│   │   ├── ticket/
│   │   ├── hypothesis/
│   │   ├── evidence/
│   │   ├── tools/
│   │   ├── agent/
│   │   └── report/
│   │
│   └── evaluation/
│
├── migrations/
│   ├── platform/
│   ├── shifthub/
│   ├── payloop/
│   ├── investigation/
│   └── evaluation-private/
│
├── scenarios/
├── fixtures/
├── docs/
│   └── workforceops/
├── scripts/
├── Taskfile.yml
├── Dockerfile
├── fly.toml
├── go.mod
└── go.sum
```

This is a proposed organization, not permission to create unnecessary abstraction or empty directories.

Keep packages cohesive and small. Introduce interfaces where they establish a real boundary, not merely to satisfy an architectural pattern.

## 4.2 Service boundaries

ShiftHub and PayLoop must not directly read and modify each other's business tables.

They communicate through defined APIs and integration payloads.

A shared Neon project may host both systems, but the database must preserve their logical separation.

---

# 5. Neon PostgreSQL requirements

Neon PostgreSQL is the primary database.

Inspect and reuse the existing bchat Neon connection configuration.

Do not replace it unnecessarily.

Use one Neon project initially, with separate schemas and restricted database roles:

- `shifthub`
- `payloop`
- `platform`
- `investigation`
- `evaluation_private`

Use separate application roles:

- `shifthub_app`
- `payloop_app`
- `sync_worker`
- `investigator_readonly`
- `evaluator`
- `migration_admin`

Apply least privilege.

The investigator must not be able to access the `evaluation_private` schema.

The PayLoop application must not have unrestricted access to ShiftHub tables.

The sync worker should have only the required cross-system capabilities, preferably through APIs.

Use PostgreSQL transactions, foreign keys, unique constraints, indexes, idempotency keys and a transactional outbox.

Do not implement full event sourcing.

Keep important event and audit records append-only where practical.

Use the existing Neon configuration for the new application only after verifying its connection and database target.

**Critical data-safety requirements:**

- Do not drop or overwrite existing bchat tables as part of an unreviewed migration.
- Do not run destructive migrations against the current Neon database.
- Do not assume the existing Neon database is disposable.
- Create a clean development database, branch or isolated schema for new development.
- Separate schema creation from any eventual legacy-schema cleanup.
- Keep migration credentials separate from runtime application credentials.
- Never expose the evaluator database credentials to the investigator or agent process.

Neon branching is optional, not a required MVP dependency.

---

# 6. Fly.io requirements

Preserve the existing Fly.io integration.

Inspect all relevant existing files before modifying them.

The target runtime consists of:

- ShiftHub API.
- PayLoop API.
- Sync worker.

Prefer using Fly.io process groups or the smallest deployment arrangement that fits the existing configuration.

Do not create separate Fly applications for every Go package or domain.

Keep the services in the same region initially unless a technical constraint requires otherwise.

Requirements:

- Preserve working image builds and deployment configuration.
- Ensure runtime processes can be started independently.
- Use Fly secrets for credentials.
- Keep Neon URLs and credentials out of committed files.
- Provide health and readiness checks.
- Ensure worker restarts do not lose pending work.
- Document deployment and rollback procedures.
- Provide a safe migration release process.

Do not perform an actual production deployment, change production secrets or run destructive database migrations without explicit authorization.

The assignment is to implement and verify deployment readiness, not to modify live production infrastructure.

---

# 7. Business functionality: bounded MVP

Implement only the business operations necessary to demonstrate the integration investigation use case.

## ShiftHub

Implement:

- Fictional employer and organization.
- Synthetic employee records.
- Employment classification and effective-dated terms.
- Departments and locations.
- Shift templates and assignments.
- Raw clock-in/clock-out events.
- Break intervals.
- Attendance interpretation.
- Attendance corrections.
- Timesheet preparation and approval.
- Overtime approval.

ShiftHub is the source of truth for the supported workforce/time data.

## PayLoop

Implement:

- Imported employee/time/pay inputs.
- Pay periods and cut-off configuration.
- Pay-code and field mappings.
- Payroll runs and calculation records.
- Basic salary calculations for the supported population.
- Supported overtime payment calculations.
- Itemised payslip output.
- Payroll batch and export records.
- Reconciliation.

PayLoop is responsible for downstream processing and payroll output.

## Initial population

Generate:

- One fictional Singapore employer.
- One tenant/legal entity.
- Two locations.
- Three departments.
- 50–100 synthetic employees.
- Three to six months of business history.
- Monthly payroll periods.

Use fixed seeds and explicit generator versions.

All people, identities, accounts and transactions must be synthetic.

Do not generate real NRIC/FIN credentials or actual bank account numbers.

Use `Asia/Singapore` for workforce business dates and payroll cut-offs. Persist instants in UTC while retaining source timezone/offset information.

---

# 8. Singapore business rules

The statutory rule implementation must be source-backed, deterministic and versioned.

Do not ask the LLM to calculate payroll or decide legal eligibility.

Use the official MOM and CPF sources referenced in the attached MVP plan.

For every implemented rule, record:

- Rule ID.
- Government authority.
- Official source URL.
- Exact section/table.
- Retrieval date.
- Effective-from and effective-until dates.
- Applicable employee categories.
- Eligibility and exclusion conditions.
- Calculation logic.
- Rounding rules.
- Golden test cases.
- Review status.

Separate:

1. Statutory requirements.
2. Employment contract terms.
3. Employer policies/configuration.
4. Deliberately injected system defects.

Never present an employer setting as a statutory requirement.

## First rule pack: overtime

Implement only the supported MOM Part IV overtime population.

The implementation must account for workman/non-workman classification, salary thresholds, managers/executives exclusions, applicable normal working hours, breaks, overtime calculations and the relevant monthly cap/payment conditions.

Do not simplify the rule into an unconditional "hours above 44 receive 1.5x" formula.

Unsupported arrangements must return an explicit `indeterminate` or `needs_review` result.

## CPF

CPF is a second-stage, gated rule pack.

Do not implement CPF using a few headline percentages.

Before coding CPF calculations, retrieve and record the complete applicable official rate tables, age bands, citizenship/SPR-year conditions, wage ceilings, OW/AW treatment and rounding rules.

Have the implementation and tests independently reviewed.

If the full source-backed rule set is not available or cannot be verified, defer CPF.

## Explicitly deferred

Do not implement these in v0.1:

- Complete annual, sick and childcare leave rules.
- Public holiday and rest-day payment calculations.
- Complete tax filing and clearance.
- IRAS AIS/IR8A/IR21 integration.
- Bank disbursement.
- Full work-pass administration.
- Real CPF or IRAS submissions.
- Every possible employment arrangement.

The simulator is not certified payroll software and must never be marketed as such.

---

# 9. Independent reference oracle

Build a deterministic expected-results calculator separate from PayLoop's calculation implementation.

This is a mandatory architectural component.

It must independently calculate expected outputs using:

- Canonical source facts.
- Applicable employment terms.
- Approved rule-pack versions.
- Explicit employer policy where relevant.
- Payroll period and cut-off.

The oracle must not import or call PayLoop's calculation functions.

It may use the same reviewed source data and statutory tables, but its calculation path must be independently implemented and independently tested.

The oracle returns:

- Expected payable hours.
- Expected supported pay components.
- Expected payroll and payslip totals.
- Applicability status.
- `calculated`, `not_applicable` or `indeterminate`.
- Calculation version and reproducibility metadata.

For clean baseline data, the oracle and PayLoop should reconcile within explicit rounding rules.

After fault injection, the oracle-versus-PayLoop discrepancy becomes part of the evaluation ground truth.

**The oracle and its results must never be accessible to the investigator or AI agent.**

---

# 10. Integration mechanics

The integration boundary is central to the MVP.

Implement:

- API or batch-based synchronization.
- Versioned payload schemas.
- Employee and timesheet mapping.
- Pay-code mapping.
- Filtering/transformation configuration.
- Cut-off timestamps.
- Idempotency keys.
- Duplicate-event detection.
- Retries with bounded backoff.
- Partial-batch failure.
- Transactional outbox.
- Delivery attempts.
- Dead-letter handling.
- Reconciliation.
- End-to-end correlation IDs.

Do not add Kafka, Redis or another distributed queue.

Use PostgreSQL-backed job/outbox state and a Go worker.

Normal processing must be correct and repeatable before introducing faults.

The integration worker must not silently skip failed records.

Every record should have a traceable processing state.

---

# 11. Incident and fault simulation

Build a mechanism-based fault injector.

Do not simply hardcode a ticket with a known answer.

A reusable fault mechanism must cause the actual application process to behave incorrectly and produce observable symptoms.

Each scenario must define:

- Scenario ID.
- Preconditions.
- Fixed seed and parameters.
- Injection point.
- Fault mechanism.
- Observable ticket symptoms.
- Expected causal chain.
- Minimum supporting evidence.
- Plausible red herrings.
- Correct resolution recommendation.
- Hidden ground truth.
- Cleanup/reset procedure.
- Evaluation rubric.

## First three incidents

### INC-001 — Approved overtime missing from payroll

Approved ShiftHub timesheet contains overtime, but an integration mapping/filter or importer excludes the pay code.

Investigator traces:

`approved timesheet → export payload → mapping/filter → import result → payroll line → payslip`

### INC-002 — Attendance hours truncated during synchronization

Raw clock events and approved source hours differ from downstream imported hours because of a reusable mechanism such as deduplication collision, event ordering or partial batch handling.

Investigator distinguishes missing source events from events lost or duplicated by the connector.

### INC-003 — Employee missing from payroll after HR update

An employee or effective-dated employment update exists in ShiftHub but does not appear correctly in PayLoop because of external-ID mapping, cut-off ordering or a dead-lettered synchronization.

Investigator traces:

`employee record → event/payload → delivery attempts → PayLoop import/status`

At least one plausible but incorrect alternative explanation must be available per incident.

## Hidden ground truth

Store the answer key in a separate schema or isolated evaluator data store.

The investigator must not be able to read:

- Expected root cause.
- Fault injector parameters.
- Expected evidence list.
- Oracle outputs.
- Scenario authoring metadata.

The incident ticket must contain only what a real support engineer would reasonably receive.

Same seed, same versions and same scenario must reproduce the same incident.

---

# 12. CaseTrace investigation engine

Do not implement a generic agent platform.

Do not implement a multi-agent swarm in v0.1.

Build one Go-controlled investigation workflow.

The workflow:

1. Parse and classify the ticket.
2. Identify affected employee, period and systems.
3. Generate competing hypotheses.
4. Record a falsifier or cheapest test for each hypothesis.
5. Collect evidence through approved tools.
6. Correlate records, events, requests and logs.
7. Verify substantive claims against evidence.
8. Confirm a diagnosis, produce a conditional result or escalate.
9. Generate JSON and Markdown investigation reports.

## Read-only tool surface

Expose typed tools similar to:

```
get_employee(employee_id)
get_employment_terms(employee_id, date)
get_roster(employee_id, date_range)
get_clock_events(employee_id, date_range)
get_attendance(employee_id, date_range)
get_timesheet(employee_id, pay_period)
get_payroll_run(pay_period)
get_payroll_lines(employee_id, pay_period)
get_payslip(employee_id, pay_period)
get_export_batch(batch_id)
get_integration_job(correlation_id)
get_delivery_attempts(correlation_id)
get_application_events(correlation_id)
```

Do not give an LLM unrestricted SQL, Neon credentials or evaluator access.

Start with typed APIs and approved query templates.

Every tool call and result must be recorded.

## Evidence ledger

Persist:

- Agent run.
- Hypothesis.
- Hypothesis disposition.
- Tool execution.
- Evidence.
- Claim.
- Evidence-to-claim relationship.
- Verdict.
- Escalation decision.

Each evidence item should include source, query/call, parameters, timestamp, correlation ID, row count and a result hash.

Each substantive claim must reference one or more evidence records.

The verifier must check that the referenced evidence exists and supports the claim. Re-execute approved evidence queries against the frozen scenario state where practical.

Do not treat the model's narrative or confidence score as proof.

## LLM integration

Inspect the existing OpenRouter/LLM implementation.

Reuse its provider connectivity only if it can be cleanly separated from the old chat/RAG architecture.

Do not reuse the old chat conversation/session architecture as the investigation engine.

Use a small provider-neutral Go interface for model calls.

All deterministic rules, permissions, tool execution, evidence storage and workflow state remain under Go control.

The application must build and test without requiring a live LLM API key.

Use a mock provider for tests.

---

# 13. Security and operational boundaries

All investigation capabilities are read-only.

The agent cannot:

- Update or delete business records.
- Alter employment terms.
- Correct attendance.
- Modify payroll.
- Re-run payroll.
- Retry integrations.
- Change credentials.
- Submit government records.
- Send external customer communications.

Any future remediation feature requires human approval and a separate permission boundary.

Enforce restrictions in Go and PostgreSQL, not in the agent prompt alone.

Use:

- Least-privilege database roles.
- Query allow-lists and approved views.
- Query timeouts and row limits.
- Tenant scoping.
- Structured audit logs.
- Secret management.
- Input validation.
- Rate limits where appropriate.

The human support engineer must be able to inspect the evidence used by the agent.

---

# 14. Required implementation phases

Do not attempt to implement everything at once.

Design first, then implement the phases sequentially in the same assignment.

Do not stop after producing design documents unless a genuine blocker prevents implementation.

## Phase 0 — Repository and infrastructure audit

Deliver:

- Existing architecture inventory.
- Existing build/test/deployment baseline.
- Fly.io and Neon reuse assessment.
- KEEP / ADAPT / REMOVE / DEFER classification.
- Refactor plan.
- Risks and destructive-change register.

Create:

```
docs/workforceops/repository-audit.md
docs/workforceops/refactor-plan.md
```

Then proceed.

## Phase 1 — Design and backend foundation

Deliver:

```
docs/workforceops/architecture.md
docs/workforceops/domain-model.md
docs/workforceops/api-contracts.md
docs/workforceops/database-design.md
docs/workforceops/rule-source-register.md
docs/workforceops/oracle-design.md
docs/workforceops/scenario-format.md
docs/workforceops/decisions/
```

Implement:

- Clean WorkforceOps application bootstrap.
- Configuration.
- Database connectivity.
- Schema migrations.
- Health/readiness endpoints.
- ShiftHub and PayLoop API skeletons.
- Sync worker skeleton.
- Local build and test commands.
- Fly.io deployment readiness using preserved configuration.

Do not build a frontend.

**Gate:** the new backend builds, starts locally, connects to an isolated Neon/PostgreSQL environment and exposes health endpoints.

## Phase 2 — Clean business workflow and oracle

Implement:

- Synthetic employer and employee model.
- Employment terms needed for the MVP.
- ShiftHub roster and attendance records.
- Timesheet and overtime approval.
- PayLoop import and pay-code mapping.
- Supported overtime calculation.
- Itemised payslip.
- Independent reference oracle.
- Deterministic synthetic-data generator.

**Gate:** a clean fixed-seed baseline flows from ShiftHub through the integration path into PayLoop and reconciles against the independent oracle.

Do not continue to fault injection until this works.

## Phase 3 — Integration and fault mechanisms

Implement:

- Outbox and worker.
- Payload mapping.
- Idempotency.
- Retries.
- Partial failures.
- Dead-letter path.
- Correlation IDs.
- First three incident mechanisms.
- Hidden ground truth.
- Scenario reset/replay.

**Gate:** all three incidents can be reproduced from clean baseline fixtures with consistent evidence and ground truth.

## Phase 4 — Investigation engine

Implement:

- Ticket intake.
- Hypothesis workflow.
- Typed read-only tools.
- Evidence ledger.
- Deterministic claim verification.
- One bounded LLM investigation workflow.
- JSON and Markdown reports.

**Gate:** the agent can investigate the three incidents without access to the answer key or write permissions.

## Phase 5 — Evaluation and demonstration

Implement:

- Scenario evaluation harness.
- Root-cause accuracy.
- Evidence precision and recall.
- Hypothesis discipline.
- Unsupported-claim rate.
- Escalation appropriateness.
- Safety compliance.
- Time/tool-call/model-cost metrics.
- Human founder baseline.
- One held-out mechanism or parameter family.
- Repeatable demo command.

**Gate:** produce an evaluation report showing actual results, failures and limitations. Do not claim production accuracy from synthetic scenarios.

---

# 15. Required commands and developer experience

The backend must be usable without a UI.

Create simple Taskfile commands or equivalent:

```
task setup
task build
task test
task lint
task db:migrate
task db:status
task seed
task seed:reset
task oracle:run
task scenario:list
task scenario:run ID=INC-001
task scenario:reset ID=INC-001
task investigate ID=INC-001
task evaluate
task demo
```

Exact command names may be adapted to existing conventions.

`task demo` should eventually execute a reproducible end-to-end demonstration:

1. Prepare a clean fixture.
2. Seed the fictional employer and business history.
3. Verify baseline calculations.
4. Inject one selected incident.
5. Generate a customer ticket.
6. Run the investigation.
7. Verify evidence.
8. Produce a Markdown report.
9. Produce an evaluation result.

The command must not silently connect to or modify production infrastructure.

---

# 16. Testing requirements

Write tests at each layer.

At minimum:

**Unit tests**

- Business-rule applicability.
- Overtime calculation.
- Timestamp/timezone handling.
- Pay-code mapping.
- Idempotency.
- Retry state transitions.
- Fault injector behavior.
- Oracle calculation.

**Integration tests**

- ShiftHub API.
- PayLoop API.
- Database migrations.
- Database role permissions.
- Outbox processing.
- Synchronization.
- Dead-letter behavior.

**Scenario tests**

- Clean baseline reconciliation.
- INC-001.
- INC-002.
- INC-003.
- Scenario reset and deterministic replay.
- Hidden answer key isolation.
- Investigator read-only boundary.

**Build and static checks**

- `go test ./...`
- `go vet ./...`
- Formatting and linting using the repository's established tooling.

Do not report a test as passing unless you actually execute it.

Do not weaken tests merely to make the build green.

---

# 17. Scope restrictions

Do not implement the following in this assignment unless required by a demonstrated dependency:

- React or another frontend framework.
- Chat widget.
- RAG.
- Embeddings.
- Vector database.
- Document ingestion/indexing.
- Full Memos functionality.
- Full HRIS product.
- Full payroll compliance platform.
- Every Singapore employment regulation.
- CPF until its source-validation gate is passed.
- IRAS/AIS integration.
- Real statutory submissions.
- Automatic payroll correction.
- Unrestricted SQL execution.
- Agent marketplace.
- Multi-agent swarm.
- Kafka, Redis or Kubernetes.
- Multiple independent microservice repositories.
- Unnecessary portability abstractions for databases not being used.

Do not add a feature because it is technically interesting.

Every implementation item must serve the MVP acceptance criteria.

---

# 18. Autonomous execution rules

You are expected to work autonomously.

Do not repeatedly ask for approval on routine implementation details.

Make sensible, reversible decisions within the scope and record them as ADRs.

If an assumption is uncertain:

- Record it.
- Mark it `UNVERIFIED`, `CONDITIONAL` or `BLOCKED`.
- Continue independent work.
- Do not silently promote it into an implemented statutory rule or product requirement.

Ask for director input only when:

1. A decision materially changes the agreed product scope.
2. Existing Fly.io or Neon infrastructure cannot be safely reused.
3. A regulatory source conflict affects a calculation or eligibility rule.
4. Required access or permission is unavailable.
5. A destructive action against existing infrastructure would be necessary.
6. A security boundary cannot be enforced with the available architecture.

Do not perform production deployment, production data deletion or destructive migration without authorization.

Avoid endless review cycles. Perform one focused independent review and one targeted correction pass per major phase unless a new blocker is discovered.

Do not stop after writing an architecture proposal. Continue implementation through the phases unless a genuine blocker prevents progress.

---

# 19. Completion criteria

The assignment is complete only when:

1. The old RAG/chat/widget/frontend product is no longer part of the new backend's runtime or build.
2. Reusable Go, Neon PostgreSQL and Fly.io foundations have been retained and documented.
3. The backend compiles and runs locally without frontend dependencies.
4. ShiftHub creates valid synthetic workforce/time records.
5. PayLoop imports them through the integration boundary.
6. The supported payroll calculation works and produces itemised payslips.
7. The independent oracle validates clean baseline results.
8. The generator is deterministic and produces synthetic data only.
9. Three incident mechanisms are reproducible.
10. Hidden ground truth is inaccessible to the investigator.
11. CaseTrace can investigate with typed, read-only tools.
12. Every substantive conclusion references evidence.
13. Evaluation results are reproducible.
14. Tests, migrations, environment configuration and developer commands are documented.
15. Fly.io deployment configuration is ready and Neon connectivity is verified in a safe environment.
16. The README and `AGENTS.md` describe WorkforceOps Lab rather than the former bchat product.

A dashboard, additional scenarios, expanded statutory coverage and a candidate training portal are not required for completion.

---

# 20. Final reporting format

At the end of the work, provide:

**A. Repository transformation**

What was reused, adapted, removed and deferred.

**B. Architecture**

Actual implemented processes, packages, database schemas, roles and API boundaries.

**C. Infrastructure**

What existing Neon and Fly.io integration was preserved and how it was verified.

**D. Implemented functionality**

What actually works, with concrete examples.

**E. Regulatory rule status**

Which rules are verified and implemented, which are deferred, and which remain uncertain.

**F. Incident results**

Demonstrate all three incidents and their reproducibility.

**G. Investigation evaluation**

Root-cause accuracy, evidence precision/recall, unsupported claims, escalation and cost/time metrics.

**H. Tests**

Commands actually executed and their results.

**I. Known limitations**

Unimplemented capabilities, remaining blockers and technical debt.

**J. Next recommended action**

The single highest-impact next step.

Do not claim success based on generated code or documentation alone. Demonstrate actual execution.

---

# FINAL DIRECTIVE

Build the smallest functioning backend that demonstrates this complete chain:

**Synthetic Singapore workforce data → ShiftHub operations → integration → PayLoop payroll → independently verified expected result → deliberately injected defect → realistic support ticket → evidence-led Tier 2 investigation.**

Reuse the existing bchat infrastructure where it saves time and preserves reliability.

Remove the old product where it creates unnecessary complexity.

Do not build a UI.

Do not build a generic chatbot.

Do not build an elaborate platform before proving the vertical slice.

**The goal is a credible, functioning enterprise simulation and investigation engine—not a large collection of Go packages.**

## How I would hand this to the coding agent

Attach both:

1. `WORKFORCEOPS_LAB_MVP_PLAN_v0.1.md`
2. The prompt above.

Start the coding agent in the local `bchat` repository, not in a newly created empty repository.

One particularly important instruction is that it must not run destructive migrations or deploy to the existing production Neon/Fly environment while stripping the old product. It should use an isolated development database and deployment configuration until the new application passes its acceptance gates.

The first milestone I would expect back is the repository audit and architecture/design artifacts, followed by actual implementation—not a proposal asking you to approve every routine decision.
