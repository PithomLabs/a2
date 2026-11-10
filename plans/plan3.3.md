## prompt

just to clarify so we are aligned, the deliverables here are at least a Web-based backend on local Postgres (deploying to Fly.io and Neon should be optional), the goal is to enable Tier 2 support engineer applicants to test their skill, prepare for interview and get meaningful experience while applying for job interviews and actual work

## answer

Yes. That clarifies the product considerably. I understand the objective differently now, and I would make this the primary product requirement, not merely an additional use case.

You are building an open-source platform that helps people become employable as Tier 2 Application Support, Technical Support and Production Support Engineers by giving them realistic, hands-on experience with enterprise business systems.

The Singapore HR/WFM/payroll environment is our first training and assessment environment.

We are building a practical work-experience simulator, not primarily a commercial payroll application or an AI support tool for existing companies.

## 1. What the actual deliverable should be

Mandatory MVP deliverable

A functioning web-based application with:

- Go backend.
- Local PostgreSQL database.
- Web interface accessible through `localhost`.
- Synthetic but realistic enterprise business systems.
- Realistic support tickets and production incidents.
- SQL, API and log investigation capabilities.
- Agent-assisted training and interview preparation.
- Automated assessment and feedback.
- Reproducible scenarios and progress records.

Neon and Fly.io are optional deployment targets. They must not be prerequisites for someone to download, install and use the platform.

| Component           | Requirement                         |
| ------------------- | ----------------------------------- |
| Go backend          | Mandatory                           |
| Web interface       | Mandatory                           |
| Local PostgreSQL    | Mandatory                           |
| Docker Compose      | Mandatory                           |
| CLI                 | Developer and administrator tooling |
| REST API            | Mandatory                           |
| Neon                | Optional                            |
| Fly.io              | Optional                            |
| External LLM        | Optional; mock mode available       |
| RAG/vector database | Not required                        |

A developer or job seeker should be able to clone the project and start using it without creating accounts with Neon, Fly.io or an AI provider.

## 2. Who is the primary user?

The primary user is an aspiring or existing technical support engineer who needs to demonstrate practical skills.

Their journey should look like this:

## The Support Engineer Learning Journey

1. Choose a career path

   Application Support · Production Support · Technical Support · WFM/Payroll Support
2. Enter a simulated enterprise

   Explore employees, workforce transactions, payroll operations, APIs and system integrations.
3. Receive a realistic support ticket

   For example: "Several employees' overtime payments are missing from this month's payroll."
4. Investigate like a real Tier 2 engineer

   Write SQL, inspect application logs, examine API requests, correlate transactions and test hypotheses.
5. Submit a resolution

   Explain the root cause, supporting evidence, business impact, workaround or resolution, and escalation requirements.
6. Receive assessment and feedback

   See which parts of the investigation were correct, what evidence was missed, and what should have been done differently.
7. Prepare for interviews and actual work

   Practise explaining findings, handling follow-up questions and producing professional support reports.

This is what makes the project useful to someone actively applying for jobs.

## 3. The web interface must be designed around the applicant

I would change the previous Operations Console into a more comprehensive Support Engineering Lab.

It should have these main areas:

My Learning Workspace

Incident Labs

Realistic tickets and investigations.

SQL Workbench

Query synthetic business data.

Logs and APIs

Investigate technical evidence.

AI Mentor

Hints, explanations and guidance.

Mock Interviews

Technical questions and follow-ups.

Skill Assessment

Progress and demonstrated competencies.

Simulated business environments

Singapore WFM and Payroll

First module

Order and Inventory Management

Future module

Accounting and Finance Operations

Future module

The simulated enterprise itself remains real from a software-engineering perspective: working Go processes, database transactions, scheduled jobs, APIs, integration failures and application logs.

It is the business and its data that are synthetic.

## 4. An important change to our agent architecture

I would make the human applicant the primary investigator.

The agents should support the applicant, rather than automatically solve every ticket.

| Agent role          | Responsibility                                                                 |
| ------------------- | ------------------------------------------------------------------------------ |
| Customer agent      | Simulates customer complaints and provides additional information              |
| Ticket agent        | Creates realistic incidents and ticket variations                              |
| Mentor agent        | Provides progressive hints in practice mode                                    |
| Interviewer agent   | Conducts technical and behavioral follow-up questions                          |
| Evaluator           | Assesses the submitted investigation against hidden ground truth               |
| CaseTrace benchmark | Independently attempts the incident so we can compare human and AI performance |

There should be three distinct operating modes:

Practice mode: The learner receives hints, explanations, sample queries and guidance.

Assessment mode: The learner receives a ticket with no answer hints. Investigation quality is scored against hidden ground truth.

Interview mode: An AI interviewer asks questions about the incident, challenges the candidate's diagnosis and assesses their ability to explain technical findings.

The evaluator must not expose hidden answers during assessment.

This separation is essential. Otherwise, the system will demonstrate that AI can solve support tickets, rather than demonstrate that a human applicant can solve them.

## 5. What does meaningful work experience mean here?

I would design the labs to produce tangible evidence of demonstrated skills.

A completed investigation should generate a portfolio record containing:

- Scenario and difficulty.
- Business problem encountered.
- SQL queries written.
- APIs and logs inspected.
- Hypotheses tested and rejected.
- Root cause identified.
- Evidence supporting the diagnosis.
- Resolution or escalation recommendation.
- Quality assessment.
- Time taken and hints used.

A candidate could share a report showing what they actually accomplished in the simulator.

We must describe this honestly as simulated enterprise support experience, not claim that someone worked on a real customer's production infrastructure.

However, completing a realistic investigation is much more meaningful than simply watching tutorials or answering multiple-choice questions.

## 6. How this changes plan3.2

I would retain its technical architecture:

- `github.com/PithomLabs/a1`.
- Existing `bchat/` untouched.
- Go backend and modular platform kernel.
- ShiftHub and PayLoop.
- Local PostgreSQL as the default.
- Optional Neon and Fly.io.
- Independent oracle.
- Incident and fault simulation.
- Extensible business modules.
- CLI, REST API and Web UI.

But I would change the product priorities.

| Priority | Deliverable                                           |
| -------- | ----------------------------------------------------- |
| P0       | Local installation and usable web application         |
| P0       | Working synthetic Singapore enterprise                |
| P0       | Realistic support-ticket investigation environment    |
| P0       | SQL, API and log investigation workspace              |
| P0       | Three reproducible incidents                          |
| P0       | Assessment and evidence-based feedback                |
| P1       | AI mentor and mock interviewer                        |
| P1       | Candidate portfolio and progress tracking             |
| P1       | CaseTrace human-versus-agent benchmark                |
| P2       | Second business module                                |
| P2       | Optional Neon/Fly.io deployment                       |
| P2       | Employer assessment and private training environments |

The platform should be extensible from the beginning, but we should not sacrifice the first complete learning experience to build an elaborate plugin ecosystem.

## 7. My revised product definition

I would use this as the authoritative statement for the coding agent:

# WorkforceOps Lab

An open-source, agent-assisted enterprise support engineering training and assessment platform.

WorkforceOps Lab provides aspiring and practicing technical support engineers with realistic, executable business environments in which they can practise investigating enterprise application problems using SQL, APIs, logs, integrations and business rules.

It enables learners to prepare for technical interviews, develop practical troubleshooting skills, demonstrate their capabilities through evidence-backed investigations, and build confidence for actual Tier 2 support responsibilities.

The platform starts with a Singapore WFM/HR/payroll reference environment and is architected to accommodate additional business domains and customer-specific simulation modules.

Primary objective: make practical enterprise support experience accessible to people who do not have access to enterprise production systems.

My final alignment: the open-source web application, running locally with PostgreSQL, is the actual product. Neon and Fly.io are optional ways to deploy it. The synthetic enterprise is the training environment. The agents are mentors, interviewers and evaluators. The candidate is the person who must learn to investigate and solve the problems.

That should govern the coding agent's design decisions from this point forward.


## prompt

yes, they should be able to simulate a real world production system as a staging environment where they can add anonymized but realistic data to the database, add custom business logic etc, anything that would make the OJT platform a reality


## answer

Yes. Now I understand the bigger ambition: you want to give people and businesses the ability to create their own realistic, production-like software environments—not just complete predefined training exercises.

That changes WorkforceOps Lab from a training simulator into something considerably more powerful.

I would describe it as:

# WorkforceOps Lab

An open-source, agent-driven enterprise staging, simulation and support-engineering training platform.

A user should be able to provision a business environment, populate it with realistic data, configure its business rules, run actual application workflows, introduce failures, investigate problems and practise resolving them.

It should feel like working inside a real company's staging environment.

## 1. What users should be able to do

I envision five fundamental capabilities.

1\. Create a business environment

Start with a reference module such as Singapore WFM/payroll, or create a custom business environment.

2\. Bring their own realistic data

Import anonymized datasets, map database fields, preserve relationships, generate additional synthetic transactions and populate PostgreSQL.

3\. Configure business logic

Define business rules, calculations, workflows, approval conditions, integration mappings and custom policies.

4\. Operate and troubleshoot the environment

Execute transactions, inspect SQL, APIs and logs, introduce defects, reproduce incidents and practise resolutions.

5\. Train, investigate and evaluate

Use AI mentors, interviewers and investigation agents. Build an evidence-backed portfolio of practical work.

This means users aren't limited to the incidents we provide. They can create their own business environment and their own problems to solve.

## 2. Three operating modes

I would organize the product around three modes.

| Mode            | Purpose                              | Example                                       |
| --------------- | ------------------------------------ | --------------------------------------------- |
| Guided Lab      | Learn using prepared environments    | Investigate missing overtime                  |
| Open Sandbox    | Build and experiment freely          | Import data, change rules, introduce failures |
| Private Staging | Reproduce a business's own workflows | Clone an anonymized WFM/payroll environment   |

The third mode is especially important commercially.

Imagine a small software company that cannot easily reproduce a customer's production issue.

It could use WorkforceOps Lab to create an isolated staging environment, import sanitized data, recreate the relevant configuration, and reproduce the problem without touching production.

The same environment could then be used to train its support engineers.

## 3. How the environment should work

I would introduce an Environment Manager as a first-class platform component.

## Environment lifecycle

Create environment

Choose reference module or custom business module

Provision data environment

Local PostgreSQL database · schemas · database roles · application configuration

Import and prepare data

CSV / JSON / approved database exports · anonymization · validation · relationship mapping

Configure business system

Business rules · workflows · API mappings · integrations · permissions

Execute business operations

Transactions · scheduled jobs · payroll · synchronization · application events

Test and investigate

Fault injection · support tickets · SQL · logs · APIs · agent investigations

Snapshot, reset, clone or share

Preserve reproducible environments and learning progress

This lifecycle becomes the foundation of the platform.

## 4. Data import must be a serious capability

I would not restrict users to randomly generated data.

They should be able to bring anonymized datasets from actual business applications.

For the initial version, support:

- CSV imports.
- JSON imports.
- Structured PostgreSQL data exports, subject to validation.
- Synthetic data generation.
- Data mapping and transformation.
- Referential integrity validation.
- Import preview and error reports.

For example, an employer could provide anonymized employee, roster, attendance and payroll records.

The platform should preserve:

- Relationships between employees and transactions.
- Employment timelines.
- Salary and hours distributions where required.
- Transaction ordering.
- Time and date relationships.
- Historical configuration changes.
- Cross-system identifiers through consistent surrogate keys.

Anonymization is not simply replacing names and email addresses. We need to consider combinations of dates, locations, salary and other attributes that could identify someone indirectly.

Raw imported data must stay inside the user's authorized environment during transformation. It must not be automatically submitted to an external LLM provider.

I would also provide a synthetic-data fallback when importing real data is not possible.

## 5. Custom business logic needs two extension levels

This is particularly important for the architecture.

I would support two complementary approaches.

Level A — Configurable business logic

A user should be able to configure:

- Eligibility conditions.
- Calculation parameters.
- Approval rules.
- Pay-code mappings.
- Validation conditions.
- Business policies.
- Integration transformations.

These should be versioned, testable and inspectable. A constrained rule definition or DSL is preferable to arbitrary code execution.

Level B — Custom Go business modules

Developers and clients should be able to implement more complex business logic in Go using the platform's module SDK.

For example:

- A custom overtime-calculation module.
- A proprietary inventory-allocation policy.
- A special insurance eligibility workflow.
- A custom order-fulfillment process.

These modules should be built and registered as part of a controlled application deployment.

I would not allow arbitrary Go source uploads to execute inside the main application process. Custom code should run in an isolated build/runtime boundary with explicit permissions.

This also becomes a natural paid service: developing and maintaining private customer modules.

## 6. The most important new architectural requirements

I would add these platform components to plan3.2.

| New component                  | Responsibility                                                  |
| ------------------------------ | --------------------------------------------------------------- |
| Environment Manager            | Create, configure, reset, clone and version environments        |
| Data Import Engine             | Import, map, transform and validate realistic datasets          |
| Anonymization Pipeline         | Remove or transform identifying information                     |
| Business Rules Manager         | Configure and version business logic                            |
| Workflow Engine                | Execute configurable business processes                         |
| Module SDK                     | Add custom Go business functionality                            |
| Environment Snapshot Manager   | Reproduce a known application and database state                |
| Workspace Manager              | Isolate individual and organizational environments              |
| Integration Adapter Framework  | Connect external business systems                               |
| Training and Assessment Engine | Turn operational scenarios into structured learning experiences |

One particularly important requirement: every environment needs a manifest containing its module versions, database schema version, rule versions, dataset version, configuration and scenario state.

That allows another developer to reproduce the same environment and incident.

### Database and hosting

Local PostgreSQL remains the default.

Neon and Fly.io remain optional deployment targets.

However, I would distinguish a developer's local environment from customer staging environments. Strong isolation may require a separate database or container per environment, rather than relying only on schemas in a shared database.

For the MVP, local Docker Compose can provision the platform and one isolated training environment. More advanced multi-environment and cloud provisioning can follow.

## 7. How this makes the OJT platform substantially more valuable

Consider the difference.

A conventional training platform gives someone a prepared ticket:

"Investigate why employee overtime is missing."

WorkforceOps Lab should allow that person to:

1. Build or provision the WFM/payroll environment.
2. Import a realistic anonymized employee dataset.
3. Configure the business's overtime policies.
4. Populate six months of attendance and payroll history.
5. Run the payroll workflow.
6. Introduce an integration failure.
7. Receive a support ticket.
8. Investigate using SQL, APIs and logs.
9. Implement an approved correction in the isolated sandbox.
10. Re-run the workflow and demonstrate that the issue is resolved.
11. Submit an evidence-backed technical report.

That is much closer to practising actual Tier 2 work.

They are not just answering questions. They are operating a business application, troubleshooting it and demonstrating a complete resolution.

## 8. How I would extend the implementation roadmap

I would not throw away plan3.2. I would expand its scope carefully.

| Phase | Deliverable                                                               |
| ----- | ------------------------------------------------------------------------- |
| 0     | Repository and infrastructure audit                                       |
| 1     | Platform kernel, module architecture, API, CLI and UI design              |
| 2     | Singapore WFM/payroll reference environment                               |
| 3     | Integration engine, oracle and incident simulation                        |
| 4     | CaseTrace investigation engine                                            |
| 5     | Training, assessment and interview preparation                            |
| 6     | Operations Console and developer experience                               |
| 7     | Extensibility proof with a second business module                         |
| 8     | Environment provisioning, import, anonymization and customization         |
| 9     | Private staging, customer-specific integrations and commercial deployment |

However, I would pull a small part of Phase 8 forward into the MVP:

Users must be able to create a workspace, configure it, load their own CSV/JSON dataset, and reset it reproducibly.

Without that, we are still delivering primarily a predefined training simulator.

More advanced PostgreSQL imports, arbitrary custom code, private customer deployments and automated environment cloning can be developed progressively.

## 9. The product vision I would now lock

# WorkforceOps Lab

## Your own enterprise. Your own data. Your own business rules. Your own production problems.

An open-source, agent-driven business-systems staging and training platform that allows individuals and organizations to provision realistic enterprise environments, import anonymized data, customize business logic, run operational workflows, simulate failures and investigate incidents.

Its purpose is to make practical enterprise software experience accessible without requiring access to a real employer's production infrastructure.

It provides the foundation for:

- Tier 2 support engineering training.
- Technical interview preparation.
- Hands-on OJT.
- Enterprise staging and integration testing.
- Business-rule verification.
- AI-agent testing and evaluation.
- Custom business-system development.

My strongest recommendation: make the Environment Manager, data-import capability and configurable business logic part of the architectural design now, even though their full implementation will be phased.

Otherwise, we risk building a beautifully engineered fixed Singapore payroll simulator and later discovering that turning it into a user-configurable staging platform requires substantial architectural changes.

The central idea should be that the platform doesn't merely supply training scenarios. It gives every learner or business the tools to create the environment in which real-world work, problems and investigations can happen.


## prompt

in the future they should be able to test SSO too and other enterprise features but that is for future as part of the roadmap, give prompt to incorporate above


## answer

I would add this as a third architecture amendment. The important distinction is that enterprise authentication and features should be part of the platform's future extensibility roadmap, not implementation requirements for the first MVP.

There are actually two future capabilities to support:

1. Organizations should eventually be able to use enterprise SSO to access WorkforceOps Lab.
2. Learners should be able to configure, test and troubleshoot SSO and other enterprise integrations inside their simulated business environments.

Both are valuable, particularly for people preparing for enterprise application support roles.

Here is the prompt I would give the coding agent.

# ARCHITECTURE AMENDMENT 3

# WorkforceOps Lab — Enterprise Identity, Security and Integration Readiness

**Applies to:** WorkforceOps Lab architecture and implementation roadmap\
**Base plan:** plan3.2\
**Previous amendments:** API-first / CLI-reproducible / UI-accessible; extensible business simulation platform; configurable environments and custom business logic.\
**Implementation stack:** Go, PostgreSQL, Docker Compose, optional Neon and Fly.io.

## 1. Purpose

Extend the WorkforceOps Lab architecture and roadmap to accommodate future enterprise identity, authentication, security, integration and operational capabilities.

The long-term ambition is to allow individuals and organizations to create realistic enterprise staging environments in which they can develop, configure, integrate, test and troubleshoot actual enterprise application features.

One important future capability is Single Sign-On (SSO).

However, SSO implementation and the broader enterprise feature suite are NOT part of the initial MVP.

The coding agent must design appropriate architectural boundaries now, document the extension points and include these capabilities in the future roadmap without introducing unnecessary implementation complexity.

**The immediate objective remains delivering a working, locally reproducible OJT and enterprise simulation platform.**

## 2. Two distinct SSO use cases

Do not confuse these two requirements.

### A. WorkforceOps Lab enterprise authentication

In the future, organizations using WorkforceOps Lab should be able to authenticate their employees through their own identity providers.

Potential capabilities include:

- OpenID Connect (OIDC).
- SAML 2.0.
- Enterprise identity-provider configuration.
- Single Sign-On.
- Single Logout where supported.
- Just-in-Time user provisioning.
- SCIM 2.0 user and group provisioning.
- Identity-provider group-to-platform-role mapping.
- Session management.
- Centralized identity and access policies.

Examples of enterprise identity providers that may eventually be supported include Microsoft Entra ID, Okta, Google Workspace and Keycloak.

Do not implement these integrations in v0.1.

### B. SSO as a simulated enterprise business capability

WorkforceOps Lab should eventually allow a learner or customer to create an enterprise environment with identity and access-management infrastructure.

For example, a simulated organization could have:

- An identity provider.
- Multiple business applications.
- Employees and organizational groups.
- Application-specific roles.
- Authentication policies.
- User provisioning and deprovisioning.
- SSO integrations.
- Application sessions.
- Authentication and authorization logs.

Learners should be able to configure and troubleshoot scenarios such as:

- OIDC redirect URI mismatch.
- Invalid client credentials.
- Expired tokens.
- Incorrect audience or issuer.
- SAML assertion validation failures.
- Invalid certificate or signing-key rotation.
- Incorrect group or role mapping.
- SCIM provisioning failures.
- User successfully authenticated but denied application access.
- Session expiration and synchronization problems.

These should eventually become realistic Tier 2 support and enterprise integration exercises.

Do not implement a complete identity provider or SSO laboratory in the initial MVP.

Design the platform so an identity-management business module or external identity-provider simulator can be added later.

## 3. Future enterprise capability catalogue

Add the following capability groups to the architectural roadmap.

### 3.1 Identity and access management

- OIDC.
- SAML 2.0.
- SSO.
- SCIM 2.0.
- MFA integration.
- LDAP/Active Directory integration where applicable.
- User lifecycle management.
- Group and role mapping.
- RBAC.
- Attribute-based authorization (ABAC).
- Session and token management.
- Service accounts.
- API keys and machine identities.

### 3.2 Enterprise security

- Tenant isolation.
- Fine-grained authorization.
- Audit trails.
- Encryption at rest and in transit.
- Secret and credential management.
- Key and certificate rotation.
- Security-event logging.
- Data-retention policies.
- Access-review workflows.
- Approval controls.
- Configurable security policies.

### 3.3 Enterprise integration

- REST APIs.
- Webhooks.
- SFTP and managed file exchanges.
- OAuth 2.0.
- Service-to-service authentication.
- Message queues.
- Event-driven integration.
- Rate limiting.
- API versioning.
- Retry and reconciliation mechanisms.
- External-system simulators.

### 3.4 Enterprise operations

- Environment isolation.
- Multiple staging environments.
- Environment promotion.
- Snapshot and restore.
- Configuration versioning.
- Deployment and rollback simulation.
- Background-job monitoring.
- Centralized audit and observability.
- Backup and recovery exercises.
- High-availability and disaster-recovery testing.
- Monitoring and SIEM integrations.

These are future capabilities, not a mandate to build them all in the MVP.

## 4. Architectural provisions to make now

The MVP must remain simple, but it must not make future enterprise development unnecessarily difficult.

### 4.1 Identity abstraction

Keep user identity separate from authentication mechanisms.

The internal application should operate on a normalized user/principal representation.

For example:

```
type Principal struct {
    ID           string
    TenantID     string
    Subject      string
    DisplayName  string
    Email        string
    Groups       []string
    Roles        []string
    AuthProvider string
}
```

This is illustrative. Design the actual identity model according to the existing authentication infrastructure and tenant requirements.

The business modules should not know whether a user authenticated through a local account, OIDC, SAML or another identity provider.

Authentication establishes identity.

Authorization determines what that identity may do.

Keep those concerns separate.

### 4.2 Authentication provider boundary

Introduce an authentication-provider interface only where justified by the existing application architecture.

Future authentication adapters should be capable of integrating with different providers without requiring changes to business-domain packages.

For v0.1:

- Local development authentication is sufficient.
- No external identity provider is required.
- The platform must work without external authentication credentials.
- Existing CLI and automated-test workflows must remain usable.
- Preserve the operator, investigator and evaluator permission boundaries.

Do not implement a homemade OIDC or SAML protocol.

When SSO is introduced, use established, maintained Go libraries and standards-compliant implementations.

### 4.3 Authorization architecture

The authorization layer should be independent of authentication.

Maintain explicit distinctions between:

- Platform administrators.
- Organization administrators.
- Environment administrators.
- Operators.
- Investigators.
- Evaluators.
- Learners.
- Read-only auditors.
- Service identities.

Do not implement an elaborate enterprise authorization engine in the MVP.

Establish a clean permission model and avoid hardcoding permission decisions inside individual business modules.

Business modules may contribute domain-specific permissions, but the platform remains responsible for enforcing authorization.

### 4.4 Environment isolation

The Environment Manager introduced by the previous amendment must be compatible with future enterprise testing.

An environment should eventually have its own:

- Module versions.
- Business rules.
- User and group configuration.
- Integration credentials.
- Identity configuration.
- Database state.
- Test scenarios.
- Audit history.

A learner must not accidentally access another user's environment.

For the MVP, provide one reproducible local environment with clear boundaries.

Do not build a complex multi-tenant enterprise hosting platform prematurely.

### 4.5 Integration framework

The integration framework should eventually support identity-provider connectors alongside conventional business integrations.

Do not hardcode HRIS, payroll or SSO assumptions into the generic integration infrastructure.

Use explicit connector capabilities, versioned contracts, configuration validation and auditable execution.

### 4.6 Enterprise capability registration

The module contract should eventually allow a business module to declare which enterprise capabilities it requires or provides.

Examples:

- Requires OIDC authentication.
- Provides SAML integration scenarios.
- Requires a SCIM provisioning connector.
- Provides RBAC test fixtures.
- Provides service-account authentication tests.

Do not build a universal enterprise capability marketplace.

Use explicit module metadata and versioned capability definitions.

## 5. Future SSO laboratory architecture

Record the following target design as a future architectural direction:

```
       Simulated Enterprise
                |
       Identity Provider
                |
        OIDC / SAML / SCIM
                |
       +--------+---------+
       |                  |
    HR System        WFM System
       |                  |
       +--------+---------+
                |
       Business Applications
                |
       Logs / Tokens / Events
                |
       Support Incident
                |
       CaseTrace Investigation
                |
       Evidence and Evaluation
```

A learner should eventually be able to configure an identity provider and multiple applications, execute login and provisioning workflows, inject authentication or authorization failures, and investigate the resulting incidents.

The platform should distinguish authentication failures from authorization failures, provisioning failures and application-session problems.

The learner should be able to inspect relevant logs, protocol exchanges and configuration through safe, simulated interfaces.

Never expose actual credentials, signing keys or private tokens in training logs.

Use synthetic credentials and test certificates.

## 6. How this extends the OJT experience

The future platform should support dedicated enterprise-feature training paths.

Examples:

**SSO and Identity Support Engineer**

Learners troubleshoot login failures, OIDC, SAML, token validation, group mapping and provisioning.

**Enterprise Integration Support Engineer**

Learners troubleshoot API authentication, expired credentials, payload mapping, synchronization, retries and reconciliation.

**Production Support Engineer**

Learners investigate application failures, job processing, database inconsistencies, resource problems and operational incidents.

**Enterprise Application Support Engineer**

Learners investigate end-to-end business workflows across HR, WFM, payroll, inventory and other enterprise modules.

These paths should reuse the same environment, incident, investigation and evaluation infrastructure.

The platform should eventually record demonstrated skills and evidence-backed practical accomplishments across these areas.

## 7. Roadmap additions

Add a future roadmap section without changing the original MVP implementation gates.

Suggested progression:

**MVP — WorkforceOps Lab foundation**

- Local PostgreSQL.
- Go backend.
- Web Operations Console.
- CLI and REST API.
- Singapore WFM/payroll reference module.
- Synthetic and anonymized data support.
- Business-rule configuration.
- Incident simulation.
- CaseTrace.
- Practice, assessment and interview preparation.

**Future Phase A — Environment authoring**

- Multiple isolated environments.
- Advanced data imports.
- Anonymization and transformation.
- Custom business rules.
- Custom workflow definitions.
- Snapshot, restore and cloning.
- Customer-specific Go modules.

**Future Phase B — Enterprise identity laboratory**

- OIDC.
- SAML 2.0.
- SCIM 2.0.
- Identity-provider simulators.
- Enterprise user and group provisioning.
- Authentication and authorization scenarios.
- SSO troubleshooting exercises.

**Future Phase C — Enterprise platform capabilities**

- Advanced RBAC and ABAC.
- LDAP/Active Directory integrations.
- Service identities.
- API gateway and security testing.
- Secrets and certificate lifecycle testing.
- Audit and SIEM integration.
- Enterprise environment promotion.
- Backup, recovery and operational resilience.

**Future Phase D — Customer-specific enterprise staging**

- Private business modules.
- Actual customer system connectors.
- Customer-specific data and configurations.
- Private deployment.
- Custom operational scenarios.
- Production-like integration and support testing.

Do not assign arbitrary delivery dates to future phases.

## 8. Required documentation

Before finalizing Phase 1 design, add:

```
docs/workforceops/
├── enterprise-capability-roadmap.md
├── identity-and-access-architecture.md
├── authentication-provider-boundary.md
├── authorization-model.md
├── environment-isolation.md
└── future-identity-lab.md
```

Document what the MVP implements, what it merely accommodates architecturally and what is explicitly deferred.

Do not imply that future SSO capabilities are already implemented.

## 9. Additional architecture acceptance criteria

For the MVP:

1. Local authentication works without an external identity provider.
2. Authentication and authorization are distinct concepts.
3. The Principal/User representation is not coupled to a specific identity provider.
4. Business modules do not implement their own independent authentication systems.
5. Existing operator, investigator and evaluator access boundaries remain enforced.
6. Environment isolation is represented in the domain and authorization design.
7. The integration framework does not assume only payroll or WFM connectors.
8. No SSO implementation is required for the first release.
9. Local onboarding and demonstration require no external identity credentials.
10. Future enterprise capability boundaries and roadmap are documented.

Do not add stub OIDC, SAML or SCIM endpoints merely to claim support.

An unimplemented capability must be explicitly marked as planned.

## 10. Final directive

WorkforceOps Lab should ultimately allow a learner, developer or company to reproduce more than business transactions.

It should allow them to reproduce the enterprise environment in which those transactions occur.

That includes business rules, data, integrations, identity, access control, security policies, operational workflows and their associated failure modes.

For now, focus on the first working enterprise business environment and the Tier 2 training experience.

Establish sound architectural boundaries for future enterprise identity and security features, but do not allow those future requirements to inflate the MVP.

**The long-term goal is a configurable, extensible enterprise staging and learning environment in which people can practise supporting the kinds of complex systems they will encounter in actual employment.**

## One important distinction to preserve

I would make the future SSO capability serve both sides of the product:

- For the platform itself: companies can eventually use their own SSO to access WorkforceOps Lab.
- For the learner: they can practise configuring and troubleshooting SSO in a simulated customer environment.

These are separate capabilities and should not be conflated in the architecture.

The coding agent should document both now, implement neither in the MVP, and ensure that the first release does not make either difficult to add later.



