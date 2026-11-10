The key architectural decision I would have the agent investigate is not simply how to connect three technologies. It is how to assign each technology a precise responsibility without duplicating functionality, introducing unnecessary infrastructure, or creating new consistency problems.

The architecture from your separate thread provides a good starting hypothesis:

- PostgreSQL 18.6: Operational source of truth, relational integrity, concurrency control, effective-dated business rules and payroll data.
- DBOS: Durable execution of business processes, including retries, crash recovery, scheduling, parallel work and workflow state.
- immudb: Cryptographically verifiable history, tamper-evident payroll records, approvals, rule-version provenance and evidence of finalized payroll calculations.

The supplied thread also raises two particularly important issues: period closure and effective-dated rule coverage. These should be architectural requirements, not optional demonstrations.&#x20;

architecture(3).md



Some current compatibility facts the research agent must account for:

- PostgreSQL 18.6 is the current stable PostgreSQL 18 patch release. PostgreSQL 19 RC1 has been announced, but it should not be substituted for the stable target. PostgreSQL 18 introduces temporal keys (`WITHOUT OVERLAPS` and `PERIOD`), among other improvements.&#x20;

  [image](https://www.google.com/s2/favicons?domain=https://www.postgresql.org\&sz=32)

  PostgreSQL+2

- DBOS Go has reached v1.4. Its semantics matter: ordinary steps are at-least-once, while DBOS datasource transactions can atomically commit application database writes and their durability records. These guarantees are not interchangeable.&#x20;

  [image](https://www.google.com/s2/favicons?domain=https://www.dbos.dev\&sz=32)

  DBOS+2

- The latest immudb release identified is v1.11.2. Its PostgreSQL wire-protocol compatibility is not equivalent to PostgreSQL 18 feature compatibility. Its documentation says it reports itself as PostgreSQL 14 and skips unsupported DDL in imports.&#x20;

  [image](https://www.google.com/s2/favicons?domain=https://github.com\&sz=32)

  GitHub+1



Here is the prompt I recommend giving your implementation agent.

# Architecture Deep Research Assignment: PostgreSQL 18.6 + DBOS + immudb

## Objective: Build a Robust, Extensible Payroll Platform and Complete a One-Month Payroll POC

You are acting as a Principal Software Architect, Senior Go Engineer, Database Architect, and Enterprise Payroll Systems Engineer.

Your assignment is to conduct a deep architectural investigation of the **existing codebase**, determine how PostgreSQL 18.6, DBOS and immudb can materially improve its architecture, and produce a concrete implementation blueprint.

This is a research-first assignment. Do not immediately start adding dependencies, rewriting code, creating migrations, or modifying the Docker environment.

Your responsibility is to understand what already exists, discover architectural weaknesses and opportunities, evaluate the technologies against actual requirements, and design the integration before implementation begins.

The ultimate objective is to deliver an executable POC demonstrating one complete month of payroll processing with sufficiently complex business rules and operational scenarios to establish the credibility of the architecture.

---

# 1. Strategic Objectives

There are four objectives, all equally important.

### Objective A: Improve the existing architecture

Investigate whether PostgreSQL 18.6, DBOS and immudb can materially improve:

- Data integrity and relational consistency.
- Effective-dated and time-dependent business rules.
- Dynamic, configurable and versioned business rules.
- Concurrent payroll processing.
- Workflow durability and recovery.
- Auditability and traceability.
- Reproducibility of historical payroll calculations.
- Extensibility across different companies, policies and regulatory regimes.
- Maintainability and ease of introducing new business capabilities.

Do not introduce technology simply because it offers attractive features.

Every architectural recommendation must solve a demonstrable problem or enable a concrete capability in the existing codebase.

### Objective B: Demonstrate meaningful complexity through a one-month payroll POC

Build toward an end-to-end payroll run covering one complete monthly pay period.

This must not be a trivial CRUD demonstration or a simple salary calculator.

It must demonstrate how the platform handles realistic interactions between employment information, attendance, scheduling, leave, compensation, statutory rules, payroll calculations, approvals, corrections and period closure.

Use the business rules and regulatory jurisdiction already established by the project. Do not invent statutory requirements.

### Objective C: Identify the best architectural use cases for each technology

Determine precisely where PostgreSQL, DBOS and immudb should be used, where they should not be used, and how they should interact.

The objective is to leverage each technology according to its strengths rather than force all business operations through all three.

### Objective D: Make integration seamless

The integration should feel like a natural evolution of the existing Go application.

We want:

- One coherent application architecture.
- Minimal operational complexity.
- Predictable database connectivity.
- Consistent local development and CI environments.
- Explicit transaction boundaries.
- Reliable recovery from failures.
- Easily reproducible development environments.
- No unnecessary infrastructure or duplicated mechanisms.

The existing local PostgreSQL Docker version 16 should be superseded by PostgreSQL 18.6 as the development baseline, subject to the migration assessment.

---

# 2. Phase One: Understand the Existing Codebase Before Proposing Anything

This is the most important part of the assignment.

Do not start from an idealized architecture. Start from the actual repository.

Read the project's instructions, including AGENTS.md, README files, architecture documents, existing implementation plans and relevant technical specifications.

Investigate the following:

**A. Application architecture**

- Go module structure.
- Package boundaries and dependencies.
- Domain models and business entities.
- Services and application use cases.
- Repositories and database access abstractions.
- Business logic placement.
- Dependency injection and initialization.
- Error handling.
- Background jobs and asynchronous processing.
- Existing workflow or orchestration mechanisms.
- Multi-tenancy, if present.
- Configuration and environment management.

**B. Database architecture**

- Existing PostgreSQL schema.
- All database migrations, particularly migrations around 00013 if present.
- Existing constraints, indexes and foreign keys.
- Effective-dated records.
- Range types and exclusion constraints.
- Transactions and isolation levels.
- Locking and concurrency mechanisms.
- Audit and history tables.
- Versioning of business rules.
- Existing period-close and adjustment mechanisms.
- Existing seed data and integration tests.

**C. Existing business rules**

Trace the actual implementation of:

- Employment and employee lifecycle.
- Compensation and salary changes.
- Pay periods and cutoff dates.
- Attendance and timesheets.
- Leave and absence.
- Overtime.
- Allowances and deductions.
- Statutory payroll calculations.
- Effective dates and rule validity.
- Retroactive changes.
- Period closure.
- Payroll adjustments.

Identify which rules are hardcoded, database-enforced, configuration-driven or not implemented.

**D. Infrastructure**

Inspect:

- Docker Compose.
- Existing PostgreSQL 16 container.
- Database volumes and initialization scripts.
- Environment variables.
- Migration tooling.
- GitHub Actions and CI.
- Local startup and shutdown procedures.
- Health checks.
- Existing external services.

**Deliverable:**

Produce a codebase architecture map showing the actual implementation, important dependencies, data flows, database boundaries and unresolved architectural issues.

For every significant observation, provide the actual file path, relevant symbol and line reference.

Do not claim a capability exists simply because a model or migration appears to anticipate it.

---

# 3. Phase Two: PostgreSQL 18.6 Architectural Assessment

Investigate the actual PostgreSQL 18.6 capabilities that could materially improve this codebase.

Do not produce a generic PostgreSQL feature overview.

Map every proposed feature to an existing or anticipated business requirement.

Pay particular attention to:

### A. Temporal constraints and effective-dated data

Investigate:

- Range types.
- `WITHOUT OVERLAPS`.
- Temporal primary and unique keys.
- `PERIOD` foreign keys.
- Exclusion constraints.
- `btree_gist`.
- Temporal coverage across adjacent records.
- Effective-dated rule packs and statutory instructions.

We need to determine whether PostgreSQL 18 can enforce important temporal business invariants directly.

Specifically investigate:

1. Can we guarantee that compensation records do not overlap?
2. Can we guarantee that an employee's applicable employment terms cover the required period?
3. Can we guarantee that the applicable rule pack covers the entire payroll period?
4. Can adjacent effective-dated records jointly satisfy a period-coverage requirement?
5. How do we detect gaps, overlaps and incomplete coverage?
6. What happens when a rule becomes effective halfway through a payroll period?

Do not confuse valid-time constraints with complete bitemporal support. Explain which temporal concerns remain application responsibilities.

Where appropriate, produce executable SQL examples and demonstrate the expected behavior.

### B. Other PostgreSQL 18 improvements

Investigate the applicability of:

- Asynchronous I/O.
- Improved query planning and indexing.
- UUIDv7.
- Virtual and stored generated columns.
- OLD/NEW values in RETURNING.
- Logical replication improvements.
- Connection and concurrency management.

Prioritize practical benefits for the actual payroll workload.

### C. Migration from PostgreSQL 16

Determine:

- Existing dependencies on PostgreSQL 16 behavior.
- Extensions that need verification.
- Compatibility of pgx, ORM and migration tooling.
- Required schema changes.
- Whether existing migrations need revision.
- Whether the upgrade should happen before the next migration.
- Safe migration and rollback strategies.

**Important:** Do not assume PostgreSQL 18.6 is a drop-in replacement for the existing environment. Verify the actual SQL and extension compatibility.

Produce a PostgreSQL 16 versus 18.6 comparison based on this codebase, not a generic feature comparison.

---

# 4. Phase Three: DBOS Architectural Assessment

Investigate the current Go implementation of DBOS and its suitability for this application.

Use the current stable Go SDK, not outdated examples.

Determine how DBOS could manage the long-running business processes that naturally arise in payroll and workforce management.

### A. Identify appropriate workflows

Consider the following candidates:

- Employee onboarding.
- Employment contract changes.
- Attendance ingestion and reconciliation.
- Timesheet validation.
- Leave approval.
- Overtime approval.
- Payroll preparation.
- Payroll calculation.
- Payroll exception resolution.
- Payroll review and approval.
- Payroll finalization.
- Period closure.
- Retroactive adjustments.
- Generation of payroll reports and payment files.

Identify which operations genuinely need durable workflows.

Do not turn every service function into a workflow.

### B. Investigate DBOS capabilities

Examine:

- Durable workflows.
- Durable steps.
- Durable transactions.
- Queues and partitioned queues.
- Scheduling.
- Concurrency controls.
- Retries and backoff.
- Workflow recovery.
- Workflow status and observability.
- Workflow versioning and migrations.
- Human approval and waiting.
- Multi-process execution.
- Multi-tenant workload isolation.

### C. Transactional guarantees

This requires particularly careful investigation.

Differentiate:

1. Ordinary DBOS workflow steps.
2. DBOS datasource transactions.
3. Transactions involving application PostgreSQL.
4. Transactions involving external systems such as immudb.

Explain the precise delivery and retry guarantees.

Determine how to achieve atomic application database writes and DBOS durability records.

Investigate whether DBOS system tables should reside in the same PostgreSQL database as the application tables or in a separate database.

Evaluate the operational, performance and isolation implications of both approaches.

Do not casually claim that ordinary DBOS steps provide exactly-once external side effects.

### D. Workflow architecture

Produce a proposed payroll workflow showing:

- Workflow boundaries.
- Step boundaries.
- Transaction boundaries.
- Durable checkpoints.
- Retry policies.
- Idempotency requirements.
- Parallelization opportunities.
- Failure recovery.
- Human intervention points.

Investigate how the Go implementation should interact with existing repositories and application services without creating a parallel architecture.

---

# 5. Phase Four: immudb Architectural Assessment

Investigate immudb as a specialist cryptographic evidence and immutable-history layer.

The hypothesis is that immudb complements PostgreSQL and DBOS rather than replaces either of them.

Validate that hypothesis against the actual application.

### A. Appropriate use cases

Investigate whether immudb should preserve evidence of:

- Payroll calculation inputs.
- Payroll calculation outputs.
- Rule-pack versions and cryptographic digests.
- Rule-pack activation and effective dates.
- Payroll approvals.
- Payroll finalization.
- Period-close decisions.
- Retroactive adjustments.
- Administrative changes.
- Reconciliation results.
- Workflow execution outcomes.

Determine which events actually deserve immutable evidence.

Avoid writing every operational database change into immudb without a clear purpose.

### B. Technical mechanisms

Investigate:

- Native Go SDK.
- Transaction APIs.
- Immutable record semantics.
- Merkle-tree verification.
- Dual proofs.
- Transaction IDs.
- Accumulative linear hashes.
- Verified transaction retrieval.
- History queries.
- DIFF OF queries.
- Audit logging.
- Backup and recovery.
- Replication.
- Embedded versus standalone service deployment.

Determine whether immudb should run as an independent Docker service or whether embedding is appropriate.

Compare the approaches, including operational isolation, maintainability and failure recovery.

### C. Data integrity and privacy

This is a payroll system. Investigate the implications of storing employee and compensation information in an immutable database.

Determine:

- What should be recorded as actual data.
- What should be represented by hashes or references.
- How sensitive information should be kept out of immutable evidence records.
- How data retention and privacy obligations interact with immutable storage.
- How encryption and key management should work.
- How independent verification would be performed.

Do not assume that cryptographic immutability automatically makes a system secure or compliant.

### D. Important architectural limitation

immudb is not a substitute for PostgreSQL's relational constraints or effective-dated business rules.

Its PostgreSQL wire compatibility must not be mistaken for full PostgreSQL 18 compatibility.

Evaluate its native Go API as the primary integration candidate. Investigate the wire-protocol interface separately and establish which use cases, if any, justify using it.

---

# 6. Phase Five: Design the Three-Technology Architecture

Produce a concrete architecture explaining when each technology participates in the overall business process.

Use this provisional division of responsibilities as a hypothesis to test:

**PostgreSQL — Operational truth**

Responsible for mutable business state, relationships, effective-dated records, relational constraints, concurrency, payroll inputs, calculated results and period state.

**DBOS — Durable execution**

Responsible for orchestrating processes, checkpointing progress, retries, recovery, queue management, scheduling and human approval workflows.

**immudb — Verifiable evidence**

Responsible for immutable provenance, cryptographic evidence, rule-version attestation, approval records and finalized payroll history.

### A. Design the end-to-end interaction

At a minimum, investigate this sequence:

1. A business event or payroll request enters the application.
2. DBOS initiates or resumes the relevant workflow.
3. PostgreSQL validates the current business state and applicable rules.
4. PostgreSQL commits the authoritative business operation.
5. The application produces an evidence/outbox event associated with the committed operation.
6. DBOS processes the evidence event durably.
7. immudb records the corresponding immutable evidence.
8. The application records the resulting immudb transaction reference.
9. Verification and reconciliation establish that both systems agree.

This is a candidate architecture, not an instruction to implement it blindly.

### B. Solve the dual-write problem

This is a critical architectural investigation.

There is no native atomic transaction spanning PostgreSQL and immudb.

Explicitly investigate:

- PostgreSQL transactional outbox.
- DBOS durable workflows.
- Deterministic event identifiers.
- Idempotent immutable writes.
- Duplicate detection.
- Retries after uncertain outcomes.
- Missing ledger records.
- Orphaned ledger records.
- Back-linking.
- Reconciliation.
- Recovery after either service becomes unavailable.

Demonstrate what happens if the application crashes:

- Immediately after PostgreSQL commits.
- Before the immudb write.
- Immediately after immudb commits.
- Before the immudb transaction reference is stored in PostgreSQL.
- During a retry after an ambiguous network failure.

Do not claim that DBOS automatically solves the dual-write problem.

Distinguish PostgreSQL transactional exactly-once guarantees from at-least-once external side effects.

Produce a failure/recovery matrix showing the expected behavior at every boundary.

### C. Dynamic business rules

This is particularly important.

Investigate how we can support increasingly complex and configurable business rules without creating an unmaintainable collection of hardcoded conditions.

Evaluate:

- Rules represented as versioned data.
- Effective-dated rule packs.
- Jurisdiction-specific rules.
- Employer-specific policies.
- Employee-specific compensation arrangements.
- Rule dependencies.
- Rule precedence.
- Rule activation and retirement.
- Full-period coverage.
- Historical reproducibility.
- Approval and release of rule changes.

Determine how PostgreSQL constraints, Go application logic, DBOS workflows and immudb provenance should cooperate.

Evaluate whether a dedicated decision/rules engine is necessary or whether a simpler versioned rule-pack architecture is sufficient for the POC.

Do not introduce another major technology without demonstrating why it is needed.

---

# 7. Phase Six: Design the One-Month Payroll POC

The POC must demonstrate meaningful business complexity, not merely prove that three Docker containers can communicate.

Build the architecture around one complete monthly payroll period.

First determine the jurisdiction and actual statutory rules supported by the project. Use authoritative regulatory sources where rules need verification.

Use realistic but fictitious employee records.

### A. Business scenarios

Identify a suitable set of employees and scenarios that exercise:

- Different compensation structures.
- Mid-period salary changes.
- Employees joining or leaving during the month.
- Different employment arrangements.
- Overtime.
- Public holidays.
- Leave and absence.
- Allowances.
- Deductions.
- Statutory contributions.
- Company-specific policies.
- Effective-dated regulatory changes.
- Late attendance corrections.
- Retroactive adjustments.

Choose the workforce size and complexity deliberately. Explain why the selected scenarios are sufficient to demonstrate the architecture.

### B. Payroll lifecycle

Design an executable workflow covering:

1. Payroll period creation.
2. Employee and employment data snapshot.
3. Attendance and timesheet ingestion.
4. Leave and absence reconciliation.
5. Compensation and rule validation.
6. Payroll calculation.
7. Exception identification.
8. Payroll review.
9. Approval.
10. Finalization.
11. Period closure.
12. Immutable evidence recording.
13. Payroll reporting.
14. Payment-file generation as a non-executing POC output.

### C. Period closure and corrections

Demonstrate that once a payroll period is finalized:

- Unauthorized postings into the closed period are rejected.
- Previously finalized calculations are not silently overwritten.
- Late corrections are represented explicitly.
- Retroactive changes are handled through adjustments in an appropriate subsequent period.
- The original calculation remains reproducible.

The distinction between current operational data and historical payroll truth must be explicit.

### D. Reproducibility

For each payroll run, investigate recording:

- Payroll period.
- Input snapshot or input digest.
- Applicable rule-pack identifiers.
- Rule-pack versions and digests.
- Calculation implementation version.
- Calculation output digest.
- Approval and finalization references.
- PostgreSQL record references.
- immudb transaction references.

A run using identical inputs and rule versions should produce a deterministic result.

Provide an explicit strategy for exact decimal arithmetic, rounding rules, timezone handling and currency.

### E. Failure and concurrency scenarios

The POC must test:

- Duplicate payroll requests.
- Concurrent processing of the same payroll period.
- Worker crashes during payroll calculation.
- Worker crashes during evidence recording.
- Duplicate workflow execution attempts.
- immudb downtime.
- PostgreSQL connection interruption.
- Recovery after application restart.
- Rule changes during a payroll process.
- Corrections submitted after period closure.

Demonstrate that these cases do not silently produce inconsistent payroll results or duplicated financial postings.

### F. POC deliverables

Define the exact data model, workflows, database migrations, fixtures, expected results and test cases.

The POC must produce:

- Payroll register.
- Employee-level payslip calculations.
- Payroll exception report.
- Statutory calculation summaries where applicable.
- Payment-file preview.
- Period-close report.
- Evidence/provenance verification report.

All financial outputs must have independently checked expected values for representative edge cases.

---

# 8. Phase Seven: Local Development and Docker Integration

We want to supersede the existing PostgreSQL 16 Docker environment with PostgreSQL 18.6.

Investigate the cleanest implementation.

### Requirements

1. Pin the PostgreSQL image to 18.6 rather than an unqualified `latest` tag.
2. Update PostgreSQL-related environment configuration.
3. Verify Docker volume and PGDATA conventions for PostgreSQL 18.
4. Review existing initialization scripts.
5. Review database migration execution.
6. Verify health checks and service dependencies.
7. Integrate DBOS into the Go application.
8. Add immudb as an appropriately isolated service.
9. Define application connectivity and credentials.
10. Preserve the ability to run migrations, tests and the application locally with a straightforward startup process.

Investigate the PostgreSQL 18 Docker image's version-specific data directory and volume layout. Do not blindly reuse the old PostgreSQL 16 data volume.

Do not delete or overwrite existing database volumes.

Produce a migration plan that clearly distinguishes:

- Fresh local database initialization.
- Existing development database migration.
- CI initialization.
- Production migration considerations.

The final development experience should be as simple as possible.

A developer should not need to manually start and configure three independent systems every time they want to work on the application.

Evaluate Docker Compose profiles, health checks, initialization dependencies and automated setup scripts.

---

# 9. Phase Eight: Architectural Trade-offs and Adversarial Review

Do not become an advocate for the three technologies.

Challenge the proposed architecture.

Answer these questions:

1. What can PostgreSQL already accomplish without DBOS?
2. What does DBOS genuinely add beyond ordinary Go services, SQL transactions and background jobs?
3. When would DBOS create unnecessary complexity?
4. What does immudb provide that PostgreSQL audit tables and restricted append-only privileges cannot provide?
5. Which records actually need cryptographic verification?
6. Is immudb worth the additional operational cost for this POC?
7. Where can PostgreSQL become a workflow bottleneck?
8. Could DBOS workflow metadata compete with payroll queries for database resources?
9. Are there failure modes introduced by the three-system architecture that do not exist today?
10. Does the design unnecessarily duplicate state or business logic?
11. What is the simplest architecture that still satisfies the requirements?
12. How can we maintain flexibility without prematurely introducing distributed systems complexity?

Explicitly evaluate at least four options:

- PostgreSQL alone.
- PostgreSQL + DBOS.
- PostgreSQL + immudb.
- PostgreSQL + DBOS + immudb.

Recommend the appropriate architecture based on evidence.

Also conduct an adversarial review of the proposed payroll lifecycle and identify invariants that cannot safely be delegated to workflow orchestration or immutable storage.

---

# 10. Required Research Deliverables

Produce one comprehensive architecture research report.

Use this structure:

**Part I — Executive Assessment**

- Existing architecture assessment.
- Principal architectural weaknesses.
- Expected benefits of each technology.
- Recommended architecture.
- Adoption decisions: now, later or not needed.

**Part II — Existing Codebase Architecture**

- Repository and package map.
- Business-domain map.
- Database architecture.
- Existing payroll lifecycle.
- Existing invariants.
- Architectural debt and gaps.

**Part III — PostgreSQL 18.6 Assessment**

- Feature-to-requirement mapping.
- Temporal constraint design.
- Effective-dating examples.
- Migration assessment from version 16.
- Compatibility risks.

**Part IV — DBOS Assessment**

- Go SDK evaluation.
- Workflow candidates.
- Transaction semantics.
- Recovery and concurrency.
- Recommended application integration.

**Part V — immudb Assessment**

- Evidence model.
- Cryptographic verification.
- Native Go integration.
- Privacy implications.
- Retry and reconciliation design.
- Operational deployment.

**Part VI — Integrated Architecture**

- Component responsibilities.
- Architecture diagram.
- Data flow diagram.
- Sequence diagrams.
- Transaction boundaries.
- Failure/recovery matrix.
- Dynamic rule-management design.

**Part VII — One-Month Payroll POC**

- Scope and assumptions.
- Business scenarios.
- Data model.
- Payroll lifecycle.
- Workflow design.
- Rule-pack design.
- Expected outputs.
- Acceptance tests.

**Part VIII — Infrastructure and Migration**

- Docker Compose design.
- PostgreSQL 16 → 18.6 migration.
- DBOS and immudb startup.
- Environment configuration.
- Local development procedure.
- CI integration.

**Part IX — Implementation Roadmap**

Break implementation into ordered, independently verifiable milestones.

For each milestone specify:

- Objective.
- Exact files/packages likely affected.
- Dependencies.
- Implementation tasks.
- Database changes.
- Tests.
- Verification criteria.
- Risks.

**Part X — Adversarial Findings and Decisions**

- Architectural weaknesses discovered.
- Unresolved questions.
- Rejected alternatives and reasons.
- Technical risks.
- Go/no-go recommendation.

---

# 11. Research Quality Requirements

This is not a superficial technology comparison.

Follow these rules:

1. Prioritize the actual codebase over architectural assumptions.
2. Use official PostgreSQL, DBOS and immudb documentation and source repositories.
3. Verify exact SDK versions, Go requirements, SQL capabilities and compatibility.
4. Record the date and version associated with external technical findings.
5. Distinguish verified facts, architectural recommendations and unverified assumptions.
6. Include actual code references and SQL examples where appropriate.
7. Identify existing functionality that should be retained rather than rewritten.
8. Do not introduce speculative features just to make the architecture appear sophisticated.
9. Do not invent payroll regulations or expected statutory calculations.
10. Do not claim a test passes without executing it.
11. Distinguish a proposed design from an implemented and verified capability.
12. Identify any license or long-term maintenance concerns associated with dependencies.

Where a technical claim affects the architecture, verify it with documentation, source code or a reproducible experiment.

---

# 12. Execution Boundaries

This assignment is research and architecture design first.

**Do not modify existing application source code, database migrations, dependencies, Docker configuration or existing database volumes during this assignment.**

Do not run destructive database operations.

Do not upgrade the existing environment yet.

If the repository has a conventional architecture-document location, create one new research report there without overwriting existing files. Otherwise, present the report in your final response.

You may run read-only inspection commands and non-destructive experiments in isolated environments when necessary to verify a technical claim.

Do not ask me to explain things that can be discovered by inspecting the repository or its documentation.

Where information is unavailable, explicitly identify the gap and propose a test or investigation to resolve it.

---

# 13. Definition of Success

The research is successful only when we can answer these questions confidently:

1. What does our existing architecture actually do today?
2. Which specific architectural problems will PostgreSQL 18.6 solve?
3. Which processes genuinely benefit from DBOS?
4. What precisely should immudb preserve and verify?
5. How do these systems cooperate without sacrificing transactional integrity?
6. How do we support dynamic, effective-dated business rules?
7. How do we guarantee reproducibility of historical payroll results?
8. How do we prevent incorrect or duplicate payroll postings?
9. How do we handle period closure and retroactive adjustments?
10. How can a developer start the entire environment without unnecessary manual intervention?
11. What is the smallest credible implementation that demonstrates all these capabilities?
12. What exact sequence of implementation milestones takes us from the current codebase to a successful one-month payroll POC?

**Final architectural principle:**

Do not build three overlapping systems.

Build one coherent business application in which:

- PostgreSQL enforces the integrity of business data.
- DBOS ensures that business processes survive failures.
- immudb establishes independently verifiable evidence of important business events.

The resulting system should be extensible enough to accommodate increasingly dynamic business rules, additional payroll jurisdictions, company-specific policies and complex workforce-management scenarios without requiring a fundamental architectural rewrite.

Your final recommendation must explain not only how to integrate these technologies, but why this particular allocation of responsibilities is the most appropriate architecture for this codebase.

**Stop after completing the research report and implementation blueprint. Await approval before making architectural or source-code changes.**

## Two architectural issues I particularly want the agent to get right

1\. DBOS must not be treated as an exactly-once wrapper for immudb.

DBOS can atomically coordinate its durability record with PostgreSQL application transactions through its datasource mechanism. An ordinary step that calls immudb is different: the external operation may succeed while the workflow checkpoint fails. The next execution must safely recognize the previous ledger write. This is why the outbox, deterministic event identifiers and reconciliation strategy matter.&#x20;

[image](https://www.google.com/s2/favicons?domain=https://docs.dbos.dev\&sz=32)

DBOS Docs+2



2\. PostgreSQL 18.6 should become the foundation, not immudb's PostgreSQL interface.

There is an important distinction between immudb accepting PostgreSQL client connections and implementing PostgreSQL's full feature set. Keep effective-dated payroll relationships, temporal constraints, operational queries and transactional business logic in real PostgreSQL. Use immudb for the evidence it is specifically designed to protect.&#x20;

[image](https://www.google.com/s2/favicons?domain=https://immudb.io\&sz=32)

immudb+1



One additional implementation detail worth preserving: the official PostgreSQL 18 Docker image changes its default data directory to `/var/lib/postgresql/18/docker` and its volume mount convention. The agent should account for this when replacing the existing PG16 container rather than accidentally losing access to the old development database.&#x20;

[image](https://www.google.com/s2/favicons?domain=https://github.com\&sz=32)

GitHub+1



My recommendation: Have the agent complete this research before allowing it to implement the integration. The most valuable outcome isn't merely a working three-container setup. It is proving that a complete payroll period can be calculated, recovered, finalized, independently verified and reproduced without compromising the integrity of historical results.