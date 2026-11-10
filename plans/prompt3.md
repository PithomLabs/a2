I want the coding agent to perform a targeted Phase 1 design correction, not redesign the platform or start Phase 2.

The following prompt translates all six findings into concrete engineering tasks, specifies the expected changes and defines exactly what must be demonstrated before the design gate can pass.

# PHASE 1 DESIGN-GATE CORRECTION MANDATE

# WorkforceOps Lab — Security, Workspace Isolation, Oracle Integrity and Module Architecture

**Project:** WorkforceOps Lab\
**Repository:** `github.com/PithomLabs/a1`\
**Original reference:** `bchat/` — read-only\
**Current baseline:** plan3.3.1 and completed Phase 0/Phase 1 design package\
**Execution mode:** Targeted architectural correction\
**Current phase:** Phase 1 — design gate\
**Authorization:** Design corrections only. Phase 2 implementation is NOT yet authorized.

---

# 1. OBJECTIVE

The Phase 0 audit and Phase 1 architecture package have been completed.

An independent architectural review identified four MATERIAL findings and two architectural refinements that must be addressed before approval to proceed to Phase 2.

Your assignment is to:

1. Inspect the relevant existing design documents and implementation scaffolding.
2. Resolve all six findings.
3. Update affected architecture, database, authorization, module and workflow specifications.
4. Update the relevant ADRs.
5. Add targeted architectural/security tests where necessary to demonstrate that the proposed design is enforceable.
6. Ensure all affected documents agree with each other.
7. Produce a consolidated correction report and recommendation for the design gate.

**Do not redesign the platform. Do not introduce unrelated features. Do not begin Phase 2 business implementation.**

Preserve the approved product definition and plan3.3.1 scope.

---

# 2. AUTHORITATIVE PRODUCT REQUIREMENTS

The following product requirements remain frozen:

- WorkforceOps Lab is an open-source enterprise simulation, OJT and support-engineering training and assessment platform.
- Primary users are aspiring and practicing Tier 2 application, technical and production support engineers.
- The web application is the primary human interface.
- CLI and REST API remain first-class interfaces.
- Local PostgreSQL is mandatory and self-contained.
- Neon and Fly.io are optional deployment targets.
- ShiftHub and PayLoop are the first reference business systems.
- The system must reproduce realistic business workflows and integration failures.
- A separate reference oracle validates expected business results.
- CaseTrace provides evidence-based investigation.
- Applicants must be able to practise diagnosing and safely resolving problems inside isolated environments.
- Practice, assessment and interview modes have different permissions.
- Users must be able to create workspaces and import their own datasets.
- The architecture must support additional business modules and private customer logic.
- Future SSO and other enterprise capabilities remain roadmap items.

Do not weaken these requirements to make the architecture easier.

---

# 3. FINDING M1 — WORKSPACE ISOLATION AND SQL WORKBENCH SECURITY

## Problem

The current design has workspace entities and application-level workspace checks, but most operational tables do not explicitly show workspace ownership.

The current database design also defers PostgreSQL RLS.

This becomes material because applicants are expected to import their own datasets and use an SQL Workbench to investigate business records.

A shared read-only database role that can read all business records is not an acceptable security boundary for multiple applicants.

Application-level filtering alone is insufficient for unrestricted SQL execution.

## Required work

### 3.1 Establish a complete workspace ownership model

Review the domain model and identify every table containing tenant-owned or workspace-owned data.

At minimum, consider:

- Employee records.
- Employment terms.
- Departments and locations where workspace-specific.
- Shift templates and roster assignments.
- Clock events.
- Attendance records.
- Timesheets and timesheet lines.
- Pay periods.
- Imported employees.
- Payroll mappings.
- Payroll runs.
- Payroll lines.
- Payslips.
- Export batches.
- Integration configurations.
- Import batches.
- Business configuration.
- Application events where they contain workspace-specific data.
- Incident tickets.
- Investigation runs.
- Evidence.
- Claims and verdicts.
- Portfolios.
- Scenario execution records.

Explicitly classify each entity as:

- Global platform data.
- Tenant-owned data.
- Workspace-owned data.
- Evaluator-private data.

Add a non-null `workspace_id` to workspace-owned records.

Define ownership inheritance for child records.

For example:

A payroll line must belong to the same workspace as its payroll run. An attendance record must belong to the same workspace as its employee and relevant timesheet.

Do not rely solely on application code to enforce these relationships.

Specify composite constraints or equivalent database enforcement where appropriate.

### 3.2 Choose and document a defensible database isolation model

Evaluate:

A. Shared tables with PostgreSQL RLS and a database identity that securely binds the connection to its authorized workspace.

B. Separate schemas or databases with workspace-specific database roles.

Select the simplest option that provides a defensible security boundary for the actual SQL Workbench.

The decision must account for:

- Shared PostgreSQL.
- Multiple workspaces.
- Multiple applicants.
- Direct SQL execution.
- Connection pooling.
- Database role membership.
- Environment creation and reset.
- Local Docker Compose.
- Future Neon deployment.

**Non-negotiable restrictions:**

- Do not rely solely on application-level workspace filtering.
- Do not rely solely on a mutable session variable such as `app.workspace_id` when executing untrusted SQL. An applicant with unrestricted SQL access may be able to change it.
- Do not give all applicants one shared database credential with access to every workspace.
- Do not allow an applicant to change database roles and access another workspace.
- Do not give the SQL Workbench access to `evaluation_private`.
- Do not give an applicant unrestricted UPDATE, DELETE, DDL or administrative privileges.

If RLS is selected, explain how workspace identity is bound to an authenticated database identity and cannot be changed by the SQL submitted by the applicant.

If workspace-specific roles or schemas are selected, document provisioning, connection-pool and migration implications.

### 3.3 Define the MVP isolation boundary

The first public release must work locally.

Define exactly how a fresh local installation and its workspaces are isolated.

Multiple-workspace support must not be advertised as secure until its database and application controls have been implemented and tested.

If a single workspace per local installation is the safest first release, explicitly define that restriction and the migration path to multiple workspaces.

Do not silently claim support for shared multi-applicant hosting.

### 3.4 SQL Workbench requirements

Specify a secure SQL Workbench that supports genuine SQL investigation.

It should eventually support:

- SELECT queries against approved business data.
- Schema and table discovery within authorized boundaries.
- Query history.
- Execution feedback.
- Statement timeout.
- Row limits.
- Query cancellation.
- Safe result display.
- Export of authorized query results.

Database permissions, not SQL parser blacklists alone, must enforce read-only access and isolation.

The SQL Workbench must not expose:

- `evaluation_private`.
- Ground-truth tables.
- Oracle expected outputs.
- Hidden scoring rubrics.
- Other applicants' data.
- Credentials or secret configuration.

### Required deliverables

Update:

- `domain-model.md`
- `database-design.md`
- `environment-isolation.md`
- `authorization-model.md`
- `permission-matrix.md`
- `architecture.md`

Create or update:

- `workspace-isolation.md`
- `sql-workbench-security.md`
- Relevant ADR.

### Acceptance tests

Specify and, where possible, implement tests demonstrating:

1. Applicant A cannot read Applicant B's business records.
2. Applicant A cannot read Applicant B's imported datasets.
3. Applicant A cannot access another workspace through SQL.
4. Attempted role escalation fails.
5. Attempted access to `evaluation_private` fails.
6. SQL mutations and DDL fail.
7. Cross-workspace foreign-key relationships fail.
8. Workspace reset affects only the authorized workspace.

---

# 4. FINDING M2 — ORACLE OUTPUT AND ANSWER-KEY LEAKAGE

## Problem

The oracle design requires all expected outputs to remain evaluator-only.

However, the domain model places `oracle_version` and reconciliation `diff` data in `payloop.reconciliation`.

The investigator database role can read PayLoop records.

This may expose the expected answer to an applicant.

The design must distinguish genuine business-system reconciliation from evaluator-only oracle comparison.

## Required work

### 4.1 Separate two reconciliation concepts

**A. Operational reconciliation**

This belongs to the simulated business application.

Examples:

- Source export row count versus imported row count.
- Source transaction totals versus destination transaction totals.
- Duplicate delivery detection.
- Missing transaction identification.
- Payroll input completeness.

This is legitimate operational evidence and may be visible to an authorized investigator.

**B. Oracle reconciliation**

This compares the application result against independently calculated expected results.

Examples:

- Expected payroll lines.
- Expected overtime amounts.
- Actual-versus-expected differences.
- Oracle output hashes.
- Hidden expected business results.
- Scoring evidence.

This is evaluation material and must remain evaluator-private.

### 4.2 Relocate oracle comparison data

Remove evaluator-sensitive oracle results and differences from ordinary PayLoop tables.

Create appropriate evaluator-private records for:

- Oracle outputs.
- Oracle-versus-application comparisons.
- Expected results.
- Hidden discrepancies.
- Scoring evidence.

Preserve genuine operational reconciliation records in PayLoop without exposing the expected oracle answer.

### 4.3 Review every exposure path

Audit all possible leakage paths:

- REST APIs.
- Investigation tools.
- SQL Workbench.
- UI templates.
- UI page metadata.
- HTMX responses.
- Logs.
- Error messages.
- Traces.
- CLI output.
- LLM prompts.
- Agent memory or conversation state.
- Exported reports.

The investigator must not receive oracle outputs through indirect means.

The evaluator interface must use an explicit authorization boundary.

### Required deliverables

Update:

- `oracle-design.md`
- `domain-model.md`
- `database-design.md`
- `authorization-model.md`
- `permission-matrix.md`
- `architecture.md`

Create or update:

- `oracle-result-isolation.md`
- Relevant ADR.

### Acceptance tests

1. Investigator cannot query expected oracle outputs.
2. Investigator cannot retrieve oracle differences through PayLoop APIs.
3. Investigator cannot discover hidden answers through SQL Workbench.
4. Rendered HTML and HTMX responses contain no hidden evaluation values.
5. Ordinary operational reconciliation remains available.
6. Authorized evaluator access works.
7. Agent investigation prompts receive only permitted business evidence.
8. Exported applicant reports contain no hidden answer-key data.

---

# 5. FINDING M3 — LEARNER REMEDIATION AND MODE-DEPENDENT AUTHORIZATION

## Problem

The product requires applicants to practise resolving problems inside a sandbox.

However, the authorization model currently describes learner and investigator paths as read-only.

These requirements conflict.

We must distinguish read-only investigation from controlled human remediation.

The learner should be able to perform realistic corrective work without obtaining unrestricted database access.

## Required work

### 5.1 Separate role permissions from training mode

Roles define what a person is authorized to do.

Modes define what the current learning exercise permits.

Do not encode every mode as a separate user role.

Retain roles such as:

- platform-admin
- org-admin
- env-admin
- operator
- investigator
- learner
- evaluator
- auditor
- service

Add explicit permissions for sandbox remediation.

### 5.2 Define mode-dependent capabilities

Specify the following behavior:

**Practice mode**

The learner may:

- Inspect business data.
- Run approved SQL SELECT queries.
- Inspect logs and APIs.
- Form and test hypotheses.
- Receive progressive mentor hints.
- Execute scenario-permitted remediation actions.
- Re-run relevant business workflows.
- Verify whether the problem was resolved.

**Assessment mode**

The learner may:

- Investigate independently.
- Use the permitted SQL/API/log tools.
- Submit diagnosis and evidence.
- Execute only those remediation actions explicitly allowed by the scenario.
- Receive no answer-revealing hints.
- Receive evaluation and feedback only after the assessment is finalized.

A scenario may be diagnosis-only or diagnosis-and-remediation. Record this explicitly in its definition.

**Interview mode**

The learner investigates and explains findings while the interviewer asks follow-up questions.

Any remediation execution must be explicitly enabled by the interview scenario.

**Operator mode**

An authorized lab operator may prepare, inject, reset and administer incidents.

### 5.3 Keep AI agents read-only

This is a hard boundary.

Mentor, interviewer, ticket and investigation agents must not be able to independently modify business records or perform remediation.

A human learner may invoke explicitly authorized remediation operations in their own sandbox.

The agent may recommend an operation, but the human must initiate it.

### 5.4 Provide controlled remediation services

Examples:

- Correct a test pay-code mapping.
- Reprocess a failed synchronization.
- Retry a failed batch.
- Create a superseding attendance correction.
- Re-run a sandbox payroll calculation.
- Correct an approved test configuration.

These operations must be implemented as typed Go application services.

No arbitrary SQL write access.

No direct application-table mutations from AI tools.

Every remediation operation must record:

- Actor.
- Workspace.
- Scenario.
- Operation.
- Original state.
- Resulting state.
- Approval or authorization.
- Timestamp.
- Correlation ID.

Remediation must be restricted to the designated sandbox workspace.

### Required deliverables

Update:

- `authorization-model.md`
- `permission-matrix.md`
- `scenario-format.md`
- `user-workflows.md`
- `architecture.md`

Create or update:

- `remediation-model.md`
- Relevant ADR.

### Acceptance tests

1. Learner can remediate an allowed practice-mode scenario.
2. Learner cannot execute arbitrary SQL mutations.
3. Assessment hints are unavailable.
4. Assessment remediation follows the scenario's `allowed_actions`.
5. Interviewer agent cannot execute remediation.
6. CaseTrace agent remains read-only.
7. Every remediation action is audited.
8. Cross-workspace remediation fails.
9. Reset returns the sandbox to its defined reproducible baseline.

---

# 6. FINDING M4 — SCENARIO INJECTOR AND ORACLE DATABASE PRIVILEGES

## Problem

The scenario injector must create hidden ground truth, and the oracle runner must store expected results.

However, the existing role matrix does not explicitly define narrowly scoped database identities for these operations.

The existing evaluator role has access to sealed information, but the generation and evaluation processes must not all receive the same privileges.

## Required work

Define separate identities and permissions for:

### A. Scenario injector

Responsibilities:

- Prepare and inject an authorized fault.
- Generate scenario-specific ground truth.
- Write ground-truth records into sealed storage.
- Create a ticket exposing symptoms only.

Permissions:

- Required fault-execution capabilities.
- INSERT-only access to the relevant sealed ground-truth records.
- No general SELECT access to historical ground truth.
- No access to unrelated evaluator material.
- No unrestricted DDL.

### B. Oracle runner

Responsibilities:

- Obtain permitted canonical source facts.
- Run the independent module oracle.
- Compare calculated expected results with actual application results.
- Store oracle results and comparison data.

Permissions:

- Read only the required source/application facts.
- Write only the designated oracle-result records.
- No general access to historical answer keys.
- No access to unrelated evaluator data.
- No unrestricted DDL.

### C. Evaluator service

Responsibilities:

- Retrieve hidden ground truth.
- Retrieve authorized oracle results.
- Compare investigation claims and evidence against the rubric.
- Produce assessment results.

Permissions:

- Read sealed evaluation material.
- Write evaluation scores and feedback where authorized.
- No modification of business application records.
- No unrestricted DDL.

### D. Human evaluator

Human evaluator permissions must be enforced through the application authorization system.

Do not expose database credentials to browser sessions.

The evaluator service performs privileged reads after checking the human evaluator's permissions.

### E. Investigator and learner

No access to any of the above service identities or sealed material.

### Required deliverables

Update:

- `database-design.md`
- `authorization-model.md`
- `permission-matrix.md`
- `scenario-format.md`
- `oracle-design.md`

Create:

- `service-identity-permission-matrix.md`
- Relevant ADR.

### Acceptance tests

Demonstrate using actual PostgreSQL role checks:

1. Scenario injector can create a sealed record.
2. Scenario injector cannot read historical sealed records.
3. Oracle runner can store its result.
4. Oracle runner cannot retrieve evaluator-only history.
5. Evaluator service can retrieve authorized sealed material.
6. Investigator and learner roles cannot access sealed material.
7. Runtime service identities cannot execute DDL.
8. Database credentials are never included in CLI output, logs or API responses.

Do not rely on documentation alone for privilege claims.

---

# 7. REFINEMENT R1 — ENFORCE THE SHIFT HUB / PAYLOOP SERVICE BOUNDARY

## Problem

The architecture says cross-system synchronization must occur through APIs and versioned event payloads.

The existing database role design grants `sync_worker` access to both the ShiftHub and PayLoop schemas.

This undermines the intended separation and makes the two systems less realistic.

## Required design

Implement the following intended interaction:

```
ShiftHub transaction
       |
       | same transaction
       v
ShiftHub business record
       |
       v
Platform transactional outbox
       |
       v
Sync worker
       |
       | versioned HTTP request
       | correlation ID
       | idempotency key
       v
PayLoop API
       |
       v
PayLoop database
       |
       v
Import result / acknowledgement
```

### ShiftHub responsibility

ShiftHub owns:

- Workforce business records.
- Source event creation.
- Immutable outbound event payload.
- Transactional outbox publication.

The business transaction and outbox event must commit atomically.

### Sync worker responsibility

The sync worker:

- Claims pending outbox events.
- Reads immutable event payloads.
- Delivers versioned requests to PayLoop.
- Records attempts, responses and delivery states.
- Handles bounded retries and dead-letter transitions.
- Maintains correlation and idempotency.

It must not query ShiftHub or PayLoop business tables directly.

### PayLoop responsibility

PayLoop owns:

- Import validation.
- Employee/pay-code mapping.
- Idempotent processing.
- Payroll data persistence.
- Import acknowledgement.

PayLoop must not query ShiftHub tables.

### Required role changes

Revisit `sync_worker` grants.

Its normal database access should be limited to:

- Platform outbox.
- Delivery attempts.
- Idempotency state.
- Dead-letter records.
- Necessary operational records.

It must not hold direct SELECT/UPDATE access to either business schema.

ShiftHub and PayLoop APIs own their respective domain writes.

Authentication between the worker and PayLoop must use a service identity.

### Failure handling

The design must address:

- Duplicate delivery.
- Timeout after successful processing.
- Retry after partial response.
- Invalid payload.
- Unsupported event version.
- PayLoop unavailable.
- Poison message.
- Dead-letter recovery.
- Out-of-order delivery.

### Required deliverables

Update:

- `architecture.md`
- `database-design.md`
- `domain-model.md`
- API contracts.
- Relevant ADRs.

### Acceptance tests

1. Sync worker cannot SELECT from `shifthub.*`.
2. Sync worker cannot SELECT from `payloop.*`.
3. Normal integration succeeds via HTTP.
4. Duplicate messages do not create duplicate payroll inputs.
5. PayLoop can acknowledge and reject messages.
6. Failed deliveries are reproducible and observable.
7. Correlation IDs trace the entire journey.
8. ShiftHub transaction and outbox event are atomic.

---

# 8. REFINEMENT R2 — COMPLETE THE BUSINESS MODULE EXTENSION CONTRACT

## Problem

The current module registrar explicitly supports routes, tools, scenarios and evaluators.

However, our product must eventually support independent business modules contributing their own domain logic, rules, workflows, data generation, oracles, fault mechanisms, migrations and UI functionality.

Do not create a universal plugin framework.

Do not introduce Go's dynamic plugin package.

Extend the existing statically registered module contract only as far as required to prove genuine extensibility.

## Required capabilities

Review and define optional module capabilities for:

1. Application services.
2. Domain workflows.
3. Business rules and rule packs.
4. Synthetic data generators.
5. Independent reference oracles.
6. Fault mechanisms.
7. Scenario definitions.
8. Investigation tools.
9. Evaluation rubrics.
10. Database migrations.
11. HTTP/API routes.
12. Web UI navigation and module-specific views.

A module does not need to implement every capability.

Use explicit optional capabilities or capability-specific interfaces.

Do not create empty placeholder methods or generic interfaces without a concrete consumer.

### Core dependency rule

Preserve:

`core/` must never import `modules/sg-wfm-payroll` or `examples/`.

The core owns the contracts.

Business modules implement the contracts.

### Registration and lifecycle

Document:

- Module registration.
- Descriptor validation.
- Capability discovery.
- Version compatibility.
- Migration ownership.
- Module activation.
- Workspace-specific module version pinning.
- Module deactivation.
- Upgrade behavior.

Module registration must fail on duplicate IDs or incompatible versions.

### Extensibility proof

Keep the existing plan to implement a tiny order-fulfillment example module later.

For the design correction, ensure that the contract can support these capabilities without requiring changes to the platform kernel.

Do not implement the full example module during this correction pass.

### Required deliverables

Update:

- `module-contract.md`
- `module-manifest-spec.md`
- `module-lifecycle.md`
- `module-versioning.md`
- `architecture.md`

Create or update:

- `extension-points.md`
- Relevant ADR.

---

# 9. CROSS-DOCUMENT CONSISTENCY AUDIT

After implementing the six corrections, inspect the entire relevant Phase 1 document package.

At minimum, cross-check:

- Architecture.
- Domain model.
- Database design.
- Authorization model.
- Permission matrix.
- Environment isolation.
- SQL Workbench security.
- User workflows.
- Oracle design.
- Scenario format.
- Module contract.
- Manifest specification.
- Module lifecycle.
- Module versioning.
- API and CLI contracts.
- Relevant ADRs.
- Phase 1 gate report.

The documents must describe the same system.

Specifically verify:

- Every workspace-owned table has explicit ownership.
- Every role has a documented permission boundary.
- Every database identity has narrowly defined grants.
- Oracle data is not in applicant-readable tables.
- Human remediation is separate from AI investigation permissions.
- Scenario injection and oracle execution have separate identities.
- Cross-system communication uses APIs and versioned events.
- Business modules can register the required capabilities.
- UI, API and CLI respect the same authorization rules.
- Assessment mode does not leak ground truth.
- The SQL Workbench cannot bypass workspace isolation.
- Environment manifests pin all relevant module, data, configuration and rule versions.

Do not simply update individual documents and leave contradictory references elsewhere.

---

# 10. PHASE 1 CORRECTION TESTS

Inspect the existing test suite before introducing duplicates.

Create or update focused architectural and database permission tests.

Use a scratch local PostgreSQL 16 database.

Required groups:

**Workspace security**

- Cross-workspace SELECT denied.
- Workspace ID forgery denied.
- SQL Workbench access restricted.
- Cross-workspace mutation denied.

**Oracle isolation**

- Investigator cannot read oracle results.
- Oracle outputs absent from ordinary API/UI results.
- Evaluator can retrieve permitted results.

**Remediation**

- Practice-mode authorized action succeeds.
- Prohibited action fails.
- Assessment action follows scenario permissions.
- All remediation actions are audited.

**Service identities**

- Injector cannot read historical ground truth.
- Oracle runner cannot read sealed evaluator history.
- Evaluator can access authorized sealed information.
- Sync worker cannot access business schemas.

**Module architecture**

- Core-to-module dependency violation fails the architectural test.
- Unsupported module capability is rejected cleanly.
- Duplicate module registration fails.
- Incompatible version fails registration.
- Existing reference module remains compatible.

Do not use mocks as the sole proof of database privilege enforcement.

Where feasible, run tests as the actual PostgreSQL roles.

---

# 11. DOCUMENTS TO UPDATE AND DELIVER

Produce a consolidated correction package:

```
docs/workforceops/
├── design-gate-corrections.md
├── design-gate-correction-test-results.md
├── architecture.md
├── domain-model.md
├── database-design.md
├── authorization-model.md
├── permission-matrix.md
├── environment-isolation.md
├── sql-workbench-security.md
├── oracle-design.md
├── oracle-result-isolation.md
├── scenario-format.md
├── remediation-model.md
├── service-identity-permission-matrix.md
├── module-contract.md
├── module-manifest-spec.md
├── module-lifecycle.md
├── module-versioning.md
└── decisions/
```

Keep the existing naming conventions and avoid duplicate documentation where an existing document can be properly updated.

Every material architectural decision must have a corresponding ADR.

Update the existing Phase 1 design-gate status.

Do not delete previous design decisions without documenting why they changed.

---

# 12. SCOPE RESTRICTIONS

This is a Phase 1 design correction.

Do not start:

- Singapore overtime implementation.
- Payroll calculation implementation.
- Reference oracle business logic.
- Synthetic workforce generation.
- Integration worker business processing.
- Fault injection.
- CaseTrace implementation.
- Applicant-facing training functionality.
- Full SQL Workbench implementation.
- Full Environment Manager implementation.
- SSO/OIDC/SAML/SCIM.
- Second business module implementation.
- Cloud deployment.
- Production migration.

You may implement small schema/role fixtures and architectural tests needed to demonstrate that the corrected design is enforceable.

Do not turn this correction pass into the implementation of Phase 2.

---

# 13. AUTONOMOUS EXECUTION REQUIREMENTS

Work autonomously.

Do not ask for approval on routine design or implementation decisions within this mandate.

For each finding:

1. Identify the existing conflicting statements.
2. Determine the minimal correct architectural change.
3. Update all affected documents.
4. Update relevant tests and ADRs.
5. Run the applicable checks.
6. Record what changed and why.

If a genuine contradiction cannot be resolved within the frozen product scope, document it as a BLOCKER rather than silently inventing a solution.

Do not use generic statements such as "enforce workspace isolation", "keep the oracle private" or "apply least privilege" as substitutes for concrete permission, data-flow and test specifications.

Every security claim must identify the enforcing mechanism.

Every architectural capability must identify its owner and dependency direction.

---

# 14. COMPLETION REPORT

At completion, provide:

### A. Findings register

For M1, M2, M3, M4, R1 and R2:

- Original problem.
- Root cause.
- Architectural decision.
- Documents changed.
- Tests added or updated.
- Verification result.
- Remaining limitations.

### B. Final architecture

Show:

- Process boundaries.
- Database schemas.
- Database roles.
- Workspace isolation.
- Oracle/evaluator separation.
- Sync data flow.
- Remediation authorization.
- Module registration and capability boundaries.

### C. Database privilege evidence

Show actual PostgreSQL role checks and their results.

Do not claim permissions are tested if only documented.

### D. Cross-document consistency

List the documents reviewed and identify remaining conflicts.

### E. Regression results

Report actual outputs of:

- `task lint`
- `task build`
- `go test ./...`
- `go vet ./...`
- Relevant PostgreSQL security tests.

Separate pre-existing failures from newly introduced failures.

### F. Final gate recommendation

Choose exactly one:

- **PASS:** All material findings resolved and required architectural tests pass.
- **CONDITIONAL PASS:** No unresolved security blocker; only explicitly listed non-blocking refinements remain.
- **BLOCKED:** One or more material findings remain unresolved.

Do not declare PASS merely because all documents were updated.

---

# FINAL DIRECTIVE

The purpose of this correction pass is to ensure that WorkforceOps Lab can safely support the intended applicant experience.

An applicant must be able to:

1. Own or create a workspace.
2. Import their own realistic business data.
3. Explore business records through SQL.
4. Investigate real integration and application failures.
5. Practise authorized remediation within the sandbox.
6. Receive meaningful feedback and assessment.
7. Export an honest portfolio of simulated enterprise experience.

At the same time:

- Applicants must not see one another's data.
- Investigators must not see hidden oracle answers.
- AI agents must not perform unauthorized remediation.
- Scenario injectors must not gain unrestricted evaluator access.
- Sync workers must respect application boundaries.
- New business modules must extend the platform without contaminating its kernel.

**Correct the architecture to support these requirements, demonstrate the security properties where possible, and stop at the Phase 1 design gate. Do not proceed to Phase 2 until the design has been reviewed and explicitly approved.**

## What I expect back from the agent

I would judge its work on six concrete outcomes:

| Finding                  | Required proof                                                                  |
| ------------------------ | ------------------------------------------------------------------------------- |
| M1 — Workspace isolation | Actual cross-workspace SQL access denial                                        |
| M2 — Oracle secrecy      | Investigator cannot retrieve expected answers through any interface             |
| M3 — Remediation         | Learner can safely resolve an authorized sandbox problem                        |
| M4 — Service identities  | Injector, oracle runner and evaluator have separate database privileges         |
| R1 — Sync boundary       | Worker functions without direct business-schema access                          |
| R2 — Extensibility       | Module contract supports domain-specific capabilities without core dependencies |

My most important concern is M1. Do not accept a design that merely adds `workspace_id` columns and claims application-level scoping is sufficient while applicants have direct SQL access.

And I would not authorize Phase 2 simply because the agent says it has fixed the documents. The database privilege tests and the actual permission boundaries are what matter.