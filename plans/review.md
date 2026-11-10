I've read both `architecture.md` and `database-design.md`, along with the Phase 0/Phase 1 completion report.

These are the two most important documents for the initial design-gate review. I need a few more before I can responsibly approve Phase 2. I don't need all 44 documents; I want the ones that expose the actual domain model, security boundaries, module architecture, calculation methodology and execution contracts.

## 1. Please upload these first

Essential for design approval

| Priority | File                      | What I need to verify                                                                    |
| -------- | ------------------------- | ---------------------------------------------------------------------------------------- |
| 1        | `domain-model.md`         | Actual HR/WFM/payroll entities, relationships, effective dates and transaction lifecycle |
| 2        | `authorization-model.md`  | Learner, operator, investigator, evaluator and workspace permissions                     |
| 3        | `oracle-design.md`        | Whether independent calculation and verification are genuinely achievable                |
| 4        | `rule-source-register.md` | Source quality, statutory applicability and verified versus unverified rules             |
| 5        | `scenario-format.md`      | Reproducibility, fault injection and hidden ground-truth isolation                       |
| 6        | `module-contract.md`      | Whether new business modules can actually extend the platform                            |
| 7        | `module-manifest-spec.md` | Module versions, dependencies, capabilities and reproducibility                          |

Please also include `module-lifecycle.md` and `module-versioning.md` if they are separate documents.

## 2. API, CLI and UI design

I need these to establish whether the three interfaces form one coherent application.

| File                                  | What I want to examine                                      |
| ------------------------------------- | ----------------------------------------------------------- |
| `api-contracts.md` and `openapi.yaml` | API consistency and resource boundaries                     |
| `cli-contract.md`                     | Local reproducibility and developer experience              |
| `user-workflows.md`                   | Actual applicant experience                                 |
| `information-architecture.md`         | Navigation and application organization                     |
| `permission-matrix.md`                | Whether users can access only their authorized environments |
| `interaction-specification.md`        | SQL Workbench, investigation and assessment workflows       |

Of these, the SQL Workbench security design and permission matrix are particularly important.

## 3. Environment and data-import design

Because we have explicitly agreed that users should eventually import their own anonymized data, configure business logic and create their own staging environments, I also need:

- Environment Manager design or specification.
- Environment manifest specification.
- CSV/JSON import and validation design.
- Workspace isolation design.
- Configurable business-rule design.

These will tell me whether the architecture can evolve beyond a fixed Singapore training environment.

## 4. Repository and implementation evidence

Finally, I need a small selection of actual implementation files, not just architecture documents:

- `repository-audit.md`
- `refactor-plan.md`
- Root `AGENTS.md`
- Root `go.mod`
- `Taskfile.yml`
- `docker-compose.yml`
- `scripts/db/init-roles.sh`
- Core and platform database migrations.
- Database role-permission tests.
- The eight ADRs under `docs/workforceops/decisions/`.

I want to inspect the real database bootstrap and permission enforcement, because those are difficult to validate from diagrams alone.

## 5. Two issues I have already spotted

These are not automatic blockers, but I want the corresponding design decisions clarified.

Issue A — Cross-system database access

The architecture says:

> Cross-system sync flows through versioned payloads with correlation IDs — never direct table access across the ShiftHub/PayLoop boundary.

However, the database design grants `sync_worker` access to both the ShiftHub and PayLoop schemas.&#x20;

architecture.md

database-design.md



I want to understand whether the worker accesses business data through APIs as intended, or whether the database grants inadvertently undermine that separation.

Issue B — SQL Workbench and workspace isolation

The database design explicitly defers row-level security and relies on application-level workspace scoping. But `investigator_readonly` receives SELECT access across the business schemas.&#x20;

database-design.md

database-design.md



That may be acceptable for a single-user local installation. It needs more thought before hosting multiple applicants or organizations with separate imported datasets.

We should establish whether the MVP promises:

- One isolated local environment per user; or
- Multiple users/workspaces sharing a database.

Those are materially different security requirements.

## 6. How I propose to review the package

I will evaluate it against six design gates:

1. Product correctness: Does it genuinely serve applicants preparing for Tier 2 support roles?
2. Domain fidelity: Can the Singapore WFM/payroll model produce realistic, internally consistent business transactions?
3. Mathematical and regulatory correctness: Can we independently verify the implemented calculations?
4. Architectural integrity: Are platform, business modules, integrations and interfaces properly separated?
5. Security and reproducibility: Are SQL access, hidden answers, imported data and environments adequately protected?
6. Extensibility: Can someone add a different business module without modifying the platform kernel?

The outcome will be one of three decisions: PASS, CONDITIONAL PASS or BLOCKED, with findings classified as BLOCKER, MATERIAL or MINOR.

My request: upload the seven essential documents first, preferably together with the API contracts and permission matrix. I can then perform the substantive design review before asking for implementation-level details.

I want us to get this right before Phase 2—not spend weeks building the wrong foundation.

I've read all nine documents:

1. `authorization-model.md`
2. `domain-model.md`
3. `module-contract.md`
4. `module-manifest-spec.md`
5. `oracle-design.md`
6. `rule-source-register.md`
7. `scenario-format.md`
8. `module-lifecycle.md`
9. `module-versioning.md`

Initial assessment: the design is promising. I have not yet issued a Phase 1 PASS/BLOCKED decision. I want to cross-check several important design interactions against the remaining documents before doing that.

## 1. My initial findings

Good design decisions

Oracle independence: The design correctly requires a separate implementation, separate derivation, golden fixtures, import-boundary tests and independent human review. It also correctly returns `indeterminate` for unsupported calculations.&#x20;

oracle-design.md



Business-domain integrity: The separation between immutable clock events, derived attendance, approved timesheets and downstream payroll records is appropriate for reproducing real support investigations.&#x20;

domain-model.md



Investigation discipline: Hypotheses, falsifiers, evidence, claims and verdicts are separately represented. This is much better than treating an LLM transcript as the investigation record.&#x20;

domain-model.md



Regulatory honesty: All Singapore overtime sources are explicitly marked unverified. No statutory implementation is authorized merely from remembered formulas. That is exactly the discipline we need.&#x20;

rule-source-register.md



Extensibility: Static registration, module versioning and an independently maintained module migration set are sensible choices for the MVP.&#x20;

module-contract.md



## 2. Issues I want to investigate further

These are preliminary findings, not yet formal blockers.

| Area                    | Concern                                                                                                                                                                | Importance |
| ----------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------- |
| Workspace isolation     | The domain model has workspace entities, but its listed HR/WFM/payroll tables do not explicitly show workspace ownership.                                              | High       |
| SQL Workbench           | The investigator database role can read business schemas, while RLS is deferred. How is access restricted to the applicant's workspace?                                | High       |
| Oracle secrecy          | `payloop.reconciliation` contains oracle differences, but the investigator role can read PayLoop. Could this disclose expected answers?                                | High       |
| Ground-truth injection  | The scenario injector needs to write sealed answers, but the role matrix only defines an evaluator with full access to `evaluation_private`.                           | High       |
| Module extensibility    | The module registrar explicitly registers routes, tools, scenarios and evaluators. How are business services, generators, rules and oracle implementations registered? | Medium     |
| Cross-system references | ShiftHub timesheets reference `pay_period_id`, while pay periods belong to PayLoop. The ownership and contract between these systems need clarification.               | Medium     |

The database design explicitly relies on application-level workspace scoping rather than RLS. That may be adequate for an isolated, single-user local environment, but it needs a defensible solution before shared applicant workspaces are introduced.&#x20;

database-design.md



The potential oracle leakage is particularly worth examining. The oracle design says its outputs must never reach investigators, while the domain model places reconciliation differences in the PayLoop schema.&#x20;

oracle-design.md

domain-model.md



I also want to distinguish genuine assessment security from self-hosted practice. Since this is open source, someone controlling their own installation can inspect its source and database. We should not promise tamper-proof assessment in that environment.

## 3. What I want to read next

You can attach the remaining documents when ready. My next review priorities are:

| Priority | Documents                                                                     |
| -------- | ----------------------------------------------------------------------------- |
| 1        | `api-contracts.md`, `openapi.yaml`                                            |
| 2        | `permission-matrix.md`, `user-workflows.md`                                   |
| 3        | Environment Manager, workspace isolation and import specifications            |
| 4        | SQL Workbench design and interaction specification                            |
| 5        | `oracle-design.md` supporting tests and Phase 1 ADRs                          |
| 6        | `architecture.md` supporting module-registration and migration implementation |

In particular, I want to see how the applicant's SQL Workbench, hidden answer key, imported datasets and workspace isolation fit together.

## 4. What I will evaluate when the package is complete

My review will focus on one central question:

Can this architecture deliver a realistic enterprise support-engineering environment where applicants can safely work with business data, investigate genuinely difficult failures, and demonstrate their skills without being handed the answers?

I will evaluate the design against six gates:

1. Product and applicant experience.
2. Singapore business-domain fidelity.
3. Calculation and oracle correctness.
4. Platform and module extensibility.
5. Security, workspace isolation and assessment integrity.
6. Local reproducibility and open-source maintainability.

I have retained the findings from this first batch. We can continue with the remaining documents without reopening or repeating this review.


I've read all five additional documents:
1. permission-matrix.md
2. user-workflows.md
3. oracle-design.md (revised copy)
4. architecture.md (revised copy)
5. environment-isolation.md
These help considerably. In particular, the applicant journey and environment lifecycle are now much clearer.

My assessment: the design is progressing well, but I have identified four material design issues that should be resolved before Phase 2. I also have two architectural refinements.
1. The good news
The revised architecture correctly establishes:
- Human applicants as primary investigators.
- CLI, REST API and Web Console using common Go application services.
- Guided practice, assessment and interview workflows.
- An evidence-led investigation and portfolio.
- A sealed evaluator environment.
- Reproducible environment manifests.
- Future multi-environment isolation.
The addition of Journey G is particularly important:
Provision/select environment → import own data → operate → encounter incident → investigate → resolve in sandbox → export portfolio.

That is the actual product experience we should optimize for. 
user-workflows.md

The revised oracle design also correctly clarifies that independence is a verifiable engineering requirement, rather than an absolute claim. 
oracle-design(1).md

2. Four material issues
Material findings — correction required
M1. Workspace isolation is not yet represented consistently
The environment document proposes multiple workspaces sharing one database in Phase 8A. 
environment-isolation.md

However, the domain model does not explicitly include workspace_id on most operational entities:
- Employees.
- Employment terms.
- Rosters.
- Clock events.
- Attendance.
- Timesheets.
- Payroll runs.
- Payslips.
The database design also defers RLS.

This creates a significant problem because the SQL Workbench allows learners to query business data directly.

Application-level authorization is not sufficient protection for arbitrary SQL queries.

My recommendation:

Before implementing multi-workspace support, choose one of two models:
- Workspace-scoped records with mandatory workspace_id, PostgreSQL RLS and database-enforced isolation.
- Separate database/schema isolation for each workspace.
For this project, I favor workspace-scoped records plus RLS when multiple workspaces share a database.

For the first local MVP, one active workspace per installation is acceptable. But the design must establish the future isolation model now.
M2. The oracle's expected answers can leak
This is the most concrete contradiction I found.

The oracle design explicitly says:
No investigator tool, API response, template or prompt may expose its outputs. 
oracle-design(1).md


But the domain model puts the oracle comparison in:

payloop.reconciliation

And the investigator's database role has SELECT access to PayLoop tables.

That could expose the exact differences that assessment mode is supposed to hide.

My recommendation:

Move oracle outputs, expected-versus-actual differences and ground-truth comparisons entirely into evaluation_private.

Keep ordinary operational payroll records and legitimate business reconciliation in PayLoop.

Provide a restricted evaluation service that can inspect the sealed results. Do not expose those results through ordinary investigator SQL or APIs.
M3. Learners cannot resolve incidents under the current authorization model
The architecture's Journey G says learners should investigate and resolve problems in the sandbox.

But the authorization model currently makes learner/investigator paths read-only, apart from investigation records and portfolio exports. 
authorization-model.md

These two requirements conflict.

An applicant should be able to practise actual remediation, not just write a report recommending what someone else should do.

My recommendation:

Introduce mode-dependent permissions:
Mode	Learner permissions
Practice	Inspect, investigate, receive hints and perform approved sandbox remediation
Assessment	Investigate independently; remediation limited by the assessment specification
Interview	Investigate and explain decisions; interviewer controls the exercise
Production/customer staging	Separate operator authorization and approval

Do not give learners unrestricted SQL UPDATE access.

Provide explicit remediation operations through Go application services, such as correcting a test configuration, retrying an integration or reprocessing a sandbox payroll run.

Record every remediation action.
M4. Scenario injection needs its own database security boundary
The scenario format says the injector writes ground truth into evaluation_private, but the role design does not define a dedicated scenario-injector database role. 
scenario-format.md

I recommend introducing:
- scenario_injector: limited INSERT access to sealed scenario-run ground truth.
- oracle_runner: limited ability to write oracle results.
- evaluator: authorized access to read and evaluate sealed results.
The injector must not have unrestricted read access to historical answer keys.

The evaluator's human-facing authorization remains separate from the credentials used by backend services.
3. Two architectural refinements
R1. Sync worker access contradicts the stated system boundary
The architecture says synchronization must go through APIs rather than direct cross-system table access.

However, the existing database design grants the sync worker access to both ShiftHub and PayLoop schemas.

I recommend a stricter implementation:

Sync worker → ShiftHub event/outbox payload → PayLoop API

The worker should primarily access the platform outbox and delivery records. ShiftHub publishes events, and PayLoop owns the importing and persistence of its own records.

That preserves the system boundary and makes the integration failures more realistic.
R2. The module contract needs to support the extensibility promise
The module contract currently registers routes, tools, scenarios and evaluators. 
module-contract.md

But our future business modules also need to provide their own:
- Domain services.
- Business rules.
- Workflows.
- Synthetic data generators.
- Reference oracles.
- Fault mechanisms.
- Migrations.
- UI contributions.
I would not add every possible registration method prematurely. But the contract needs a documented capability-registration strategy that supports the reference Singapore module and the order-fulfillment example without modifying the core.

4. Current design-gate scorecard
This is my interim assessment, not the final Phase 1 verdict.
Gate	Current assessment
Product and applicant experience	Strong
Business-domain model	Promising
Oracle methodology	Strong design
Regulatory source discipline	Strong discipline
Scenario reproducibility	Good
Module extensibility	Needs refinement
Authorization and isolation	Needs correction
Assessment integrity	Potential leakage
Practical remediation	Permission gap

I am not recommending a major redesign. These are targeted corrections to the existing architecture.
5. What I still need from you
The most valuable remaining files are:
1. api-contracts.md
2. openapi.yaml
3. sql-workbench design or interaction specification.
4. environment-manager.md or equivalent.
5. CSV/JSON import and validation specification.
6. workspace-isolation implementation decisions, if there is a more detailed version.
7. The eight ADRs under docs/workforceops/decisions/.
Also, please include the actual Phase 1 migration and database-role test files when available. I want to compare the intended security model with what the agent has actually implemented.
6. My current recommendation
Do not authorize Phase 2 just yet.

I want to resolve the oracle leakage, workspace isolation, learner remediation and sync-worker access boundaries first.

Once these are corrected, the remaining API and module-contract review can establish whether the design is coherent enough to freeze.

The guiding principle is important:

An applicant must be able to practise the work of a real Tier 2 engineer, including diagnosing and safely resolving problems, without being able to access another learner's environment or see the answer key.

That is the security and functional standard I want us to preserve throughout implementation.


