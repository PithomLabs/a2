Yes. I think this is the missing piece that will make the business-rule formalization genuinely rigorous. I would adopt the Phys-Research methodology, not just Phys-Pseudo.

There is an important distinction:

- Phys-Research establishes where the knowledge comes from, how it is acquired, verified, derived and independently checked.
- Phys-Pseudo expresses what the business rule actually means.
- Go implements the rule.
- The independent oracle and test suite establish whether the implementation produces the expected results.

Together, these form a reproducible business-rule research and verification pipeline.

I would make this a formal part of WorkforceOps Lab.

I couldn't retrieve the actual `phys-research` repository contents directly during this session, so I won't claim to have verified its exact internal commands or conventions. However, the published PhysLib reference workflow describes several relevant practices: pinned sources, explicit derivation steps, independent goldens and anti-goldens, propagated assumptions, verification entry points, manifests and documented limitations.&#x20;

[image](https://www.google.com/s2/favicons?domain=https://github.com\&sz=32)

GitHub



# 1. The proposed architecture: Business Rule Research Pipeline

Think of how we investigate a physics paper.

We do not ask an agent to read a PDF and immediately produce Go code.

We acquire the paper, preserve its identity, identify its equations, formalize assumptions, reconstruct derivations, test examples and independently verify the results.

We should do precisely the same with Singapore business rules.

## WorkforceOps — Business Rule Research Pipeline

1. Authoritative source acquisition

   Download current Singapore legislation, subsidiary legislation, MOM guidance, CPF tables, IRAS documentation and applicable technical specifications.
2. Source preservation and provenance

   Preserve original documents, exact URLs, versions, effective dates, retrieval timestamps, hashes and precise section/page references.
3. Rule extraction

   Identify statutory claims, definitions, formulas, applicability conditions, exceptions and operational requirements.
4. Phys-Pseudo formalization

   Convert the extracted rules into explicit `given`, `define`, `assume`, `claim`, `derive`, `if/then/else` and `?` statements.
5. Independent derivation

   Reconstruct the mathematical calculation and decision logic without relying on the eventual Go implementation.
6. Golden tests and adversarial cases

   Establish expected results, boundaries, counterexamples, anti-goldens and unsupported conditions.
7. Independent verification

   A separate reviewer checks sources, assumptions, derivations and expected results.
8. Go implementation and TDD

   Only after the rule passes its research gate does it become an approved implementation contract.
9. Verification and reproducibility

   Execute tests, compare the application with the independent oracle, preserve versions and publish the evidence.

This is much stronger than simply putting an AI-generated `.phys` file beside some Go tests.

# 2. The equivalent of downloading physics papers

For WorkforceOps, the definitive source material is different.

Instead of arXiv papers, we acquire primary Singapore government material.

| Physics research         | WorkforceOps research                                  |
| ------------------------ | ------------------------------------------------------ |
| Published research paper | Singapore legislation and official regulatory document |
| arXiv version            | Versioned legislation or published official document   |
| Equation                 | Statutory formula or business calculation              |
| Definition               | Legal or business definition                           |
| Assumption               | Applicability condition                                |
| Derivation               | Calculation or decision-rule derivation                |
| Theorem                  | Conditional business-rule claim                        |
| Mathematical fixture     | Golden business-rule test                              |
| Independent verification | Separate rule and calculation review                   |
| Research manifest        | Source and rule manifest                               |

Sources should be acquired from:

- Singapore Statutes Online — governing legislation and amendments.
- Ministry of Manpower — employment rules and operational guidance.
- CPF Board — contribution tables, wage classifications and submission requirements.
- IRAS — employer income reporting and tax clearance.
- Other applicable Singapore government authorities.

Third-party vendor documentation is a separate source category. It helps establish how software operates, not what the law requires.

## Source acquisition must be reproducible

I would require the agent to download and preserve the actual reference material.

For each source, record:

```
{
  "source_id": "SG-MOM-OT-001-SRC-01",
  "authority": "MOM",
  "document_title": "Hours of work, overtime and rest day",
  "source_url": "https://www.mom.gov.sg/employment-practices/hours-of-work-overtime-and-rest-days",
  "document_type": "official_guidance",
  "published_at": null,
  "effective_from": null,
  "retrieved_at": "2026-11-08T00:00:00Z",
  "sha256": "...",
  "sections": [
    "Overtime pay"
  ],
  "verification_status": "retrieved_not_yet_adjudicated"
}
```

This is illustrative.

A source being successfully downloaded does not mean its interpretation is verified.

The acquisition system should preserve:

1. Original downloaded document.
2. Extracted text or structured tables.
3. Exact source locator.
4. Retrieval metadata.
5. Content hash.
6. Extraction tool/version.
7. Any image-based formula or table requiring manual verification.

For dynamically updated HTML, preserve a dated snapshot rather than assuming the URL always represents the same content.

For redistribution, inspect government document licensing and retain original notices; do not automatically republish every downloaded document.

# 3. Phys-Pseudo becomes the formal derivation layer

I would retain the seven existing constructs exactly.

Here is how I would use them in a business-rule research package.

Consider `SG-MOM-OT-001`.

Given

Facts obtained from the employee's employment record, contractual terms, approved attendance and applicable source version.

Define

What constitutes a workman, non-workman, basic salary, normal working hours and overtime for the specific provision being implemented.

Assume

The supported employment arrangement, employee classification and other necessary conditions.

Claim

The employee qualifies for the particular statutory overtime treatment.

Derive

The applicable hourly rate and overtime payment.

If/then/else

Branch between covered, excluded, unsupported and insufficient-information cases.

?

Capture unresolved interpretations, missing inputs and conditions that cannot safely produce an answer.

The important discipline is that the agent must not quietly fill a gap using a remembered payroll convention.

I would make one requirement especially strict:

Every load-bearing premise in a derivation must trace to a pinned source or an explicitly identified contract/employer configuration.

This prevents a legal requirement, a company's own policy and a convenient implementation assumption from becoming indistinguishable.

# 4. Independent derivation and verification

This is where I would borrow heavily from the PhysMath/PhysLib discipline.

We should have separate work products.

| Role                   | Responsibility                              |
| ---------------------- | ------------------------------------------- |
| Source researcher      | Retrieves and pins official references      |
| Rule analyst           | Extracts conditions, definitions and claims |
| Formalizer             | Authors Phys-Pseudo                         |
| Calculation researcher | Independently derives numerical results     |
| Test designer          | Creates boundary cases and anti-goldens     |
| Go implementer         | Implements the approved contract            |
| Independent verifier   | Challenges the entire chain                 |

The same agent can perform different tasks sequentially, but it must not certify its own work as independent verification.

A reviewer should receive the source references, definitions, assumptions, derivation artifacts and test evidence. It should not merely be shown the author's preferred conclusion.

This is particularly important for CPF and payroll rounding.

For example, we must distinguish:

- Mathematical arithmetic.
- Legally prescribed rounding.
- Software storage precision.
- Final monetary presentation.

Passing a Go test does not establish that the underlying legal interpretation was correct.

The PhysMath design makes an analogous distinction: replay and structural validation establish representational facts, not scientific truth. The same limitation applies here.&#x20;

[image](https://www.google.com/s2/favicons?domain=https://pkg.go.dev\&sz=32)

Go Packages



# 5. The business-rule package should become a research fixture

I would introduce a reusable package structure.

```
modules/sg-wfm-payroll/rules/
└── SG-MOM-OT-001/
    ├── sources/
    │   ├── manifest.json
    │   └── extracted-references.md
    │
    ├── formalization/
    │   ├── rule.phys
    │   ├── assumptions.md
    │   ├── decision-table.yaml
    │   └── derivation.md
    │
    ├── verification/
    │   ├── golden-tests.json
    │   ├── anti-goldens.json
    │   ├── boundary-cases.json
    │   └── independent-review.md
    │
    ├── implementation/
    │   └── contract.md
    │
    ├── manifest.json
    └── verify_test.go
```

I would establish a strict lifecycle:

## Rule promotion lifecycle

`DISCOVERED`

Candidate

`SOURCES_PINNED`

`CLAIMS_EXTRACTED`

`FORMALIZED`

`INDEPENDENTLY_CHECKED`

`GOLDENS_VERIFIED`

`IMPLEMENTABLE`

Approved for TDD

Failure at any stage retains the debt and blocks promotion of dependent work.

The manifest should distinguish implementation readiness from the status of the source itself.

A source may be authoritative but insufficiently specific to justify a particular implementation.

# 6. Our current S$13.60 discrepancy is the perfect test of this approach

Consider the problem we identified in the Singapore research.

The current MOM guidance gives S$13.60 for a non-workman's maximum hourly overtime rate, while the published monthly-rated formula evaluated at S$2,600 produces approximately S$13.63636.&#x20;

[image](https://www.google.com/s2/favicons?domain=https://www.mom.gov.sg\&sz=32)

mom.gov.sg+1



A conventional coding agent might simply choose one value, write a test and make the implementation pass.

Our proposed research pipeline must instead:

1. Download the actual current MOM guidance.
2. Download the applicable legislation and amendment history.
3. Locate the precise statutory provision.
4. Record the difference between the displayed rate and formula.
5. Formalize both statements without conflating them.
6. Identify whether a prescribed rounding or calculation convention resolves the difference.
7. Independently calculate relevant boundary examples.
8. Ask for authoritative clarification if necessary.
9. Keep dependent fixtures unverified until the issue is settled.

This is exactly the behavior I want from the platform.

An agent that says `?` when it cannot establish a correct result is more useful than an agent that confidently produces an unsupported payroll calculation.

# 7. How this fits into the existing plan4.2

I would not replace plan4.2. I would expand it into a formal Phys-Research-inspired Business Rule Research Protocol.

The division of responsibility would be:

| Existing component                        | New responsibility                                              |
| ----------------------------------------- | --------------------------------------------------------------- |
| Phys-Research methodology                 | Source acquisition, evidence management and derivation workflow |
| Phys-Pseudo                               | Formal representation of business-rule semantics                |
| PhysArtifact-inspired provenance          | Stable source, derivation, rule and test identities             |
| PhysLib-inspired domain structure         | Singapore-specific business-rule interpretation                 |
| PhysMath-inspired verification discipline | Explicit, replayable calculations and derivation checks         |
| Go                                        | Actual business-rule implementation                             |
| TDD                                       | Implementation verification                                     |
| Independent oracle                        | Expected-results calculation                                    |
| CaseTrace                                 | Investigates failures in the running system                     |

I would not import PhysMath wholesale into the payroll engine. Business rules have legal applicability, policy configuration, temporal conditions and human interpretation that are not necessarily represented by pure mathematical structures.

Instead, we should adapt the useful architectural principles: stable references, explicit assumptions, derivation provenance, validation, replay and evidence-bound results.

## 8. Exact instruction to add to plan4.2

This is the amendment I would give the research agent.

# ARCHITECTURE AMENDMENT

# Phys-Research-Inspired Business Rule Research and Verification Pipeline

## Objective

Extend plan4.2 to adopt the Phys-Research methodology for acquiring authoritative Singapore business documentation, extracting business-rule claims, formalizing them, independently deriving expected results, and preserving reproducible evidence.

The objective is to establish a rigorous research pipeline for Go business rules, adapted from the scientific research and verification workflow used across the Phys ecosystem.

Phys-Research provides the research-process model.

Phys-Pseudo provides the seven-construct formalization language.

PhysArtifact/PhysMath/PhysLib provide architectural inspiration for provenance, structured representations, derivation records and independent verification. Do not introduce unnecessary physics-specific dependencies into WorkforceOps.

## 1. Authoritative source acquisition

The research agent must retrieve the actual primary documents from authoritative Singapore sources.

Prioritize:

- Singapore Statutes Online.
- Ministry of Manpower.
- CPF Board.
- IRAS.
- SkillsFuture Singapore.
- Other relevant official Singapore authorities.

The agent must download the actual documents or preserve a reproducible snapshot of the authoritative web content.

Search snippets and secondary summaries are discovery material only.

For every source, record:

- Stable source identifier.
- Authority.
- Document title.
- Original URL.
- Document type.
- Publication/version date where available.
- Legal effective date where applicable.
- Retrieval timestamp.
- Exact section/table/page.
- Original-file SHA-256.
- Extracted-content hash.
- Extraction method/tool version.
- Access or redistribution limitations.
- Verification status.

Preserve original source artifacts and extracted representations separately.

For image-based formulas, tables and scanned documents, retain the original page and record the transcription independently.

A successful retrieval does not imply that a source claim is verified.

## 2. Source and claim ledger

Create a durable source and claim register.

Every material business-rule claim must reference one or more source artifacts.

Record:

- Claim identifier.
- Exact claim.
- Source reference.
- Locator.
- Applicable dates.
- Assumptions.
- Supporting evidence.
- Contradictory evidence.
- Verification status.
- Dependent rules and tests.

Do not silently replace an older source.

A new source version creates a new evidence record and may reopen previously verified claims when the applicable meaning changes.

Keep legislation, official guidance, employer policy, contract terms, vendor behavior and engineering inference as distinct evidence categories.

## 3. Phys-Pseudo formalization

Apply the existing authoritative seven-construct Phys-Pseudo specification.

Every L1 high-risk business rule must have a conforming `.phys` specification.

The specification must explicitly capture:

- Given facts.
- Definitions.
- Assumptions.
- Claims.
- Derivations.
- Conditional branches.
- Unsupported and unresolved cases using `?`.

All premises must be traceable to pinned sources or explicitly identified business configuration.

Do not modify the upstream Phys-Pseudo syntax.

No interpreter, compiler or code generator is authorized.

## 4. Independent derivation

For every high-risk calculation:

1. Reconstruct the rule from the pinned source.
2. Identify all applicability conditions.
3. Define the inputs and outputs.
4. Specify units, precision and legal rounding boundaries.
5. Derive worked examples independently.
6. Record boundary conditions.
7. Identify unsupported arrangements.
8. Produce independently calculated expected results.

Do not derive all expected results mechanically from the same `.phys` file or the eventual Go implementation.

Preserve discrepancies and unresolved interpretation as explicit research debt.

## 5. Golden tests and adversarial verification

Create versioned golden and anti-golden fixtures.

Include:

- Normal cases.
- Lower and upper boundaries.
- Just below and above thresholds.
- Effective-date transitions.
- Exceptional conditions.
- Retroactive corrections.
- Missing or conflicting information.
- Unsupported cases.

Each fixture must include its source and derivation provenance.

Quarantined or unverified fixtures must be excluded from statutory test runners.

An independent verifier must examine the source, formalization, derivation and expected results.

No researcher or implementer may independently certify their own work.

## 6. TDD promotion gate

A business rule becomes eligible for Go implementation only after:

- Authoritative sources are pinned.
- Material claims are extracted.
- Applicability and assumptions are formalized.
- Phys-Pseudo passes conformance checks.
- Derivation has been independently checked.
- Golden cases have independent provenance.
- Adversarial review has no unresolved load-bearing issue.
- Unsupported cases have defined `indeterminate` behavior.

Only then may the approved implementation contract be used to write failing Go tests and implement the business rule.

## 7. Rule package and manifest

Adopt a standard Business Rule Research Package containing:

- Source manifest.
- Extracted claims.
- Phys-Pseudo specification.
- Assumption register.
- Derivation record.
- Decision table.
- Golden tests.
- Anti-goldens.
- Boundary cases.
- Independent review.
- Implementation contract.
- Rule status manifest.

Keep the package versioned and reproducible.

## 8. Rigor tiers

Apply three levels:

- L1: full research/formalization/independent verification package for statutory, financial and high-consequence rules.
- L2: structured state-machine or workflow specification with selective Phys-Pseudo for complex integrations.
- L3: conventional software specifications and tests for routine CRUD and infrastructure behavior.

Maintain a rigor-tier register and justify every L2/L3 classification.

Do not force every software feature into Phys-Pseudo.

## 9. Adapted Mathematical Reality Checks

Apply the three Reality Checks to every L1 rule:

- RC1: identify the full legal, contractual, policy and operational landscape, including applicability, regimes, exceptions and limitations.
- RC2: identify which conditions and data structures must intersect to produce the claimed business result, test compatibility and remove unsupported assumptions.
- RC3: specify the transformation from official source to formal rule, from rule to calculation, and from source-system data to downstream application result. Record preserved and lost meaning, assumptions, approximations, validity conditions and failure conditions.

## 10. Current phase restriction

For the current plan4.2 research and dossier phase, this amendment authorizes source acquisition, extraction, formalization, derivation, verification, fixtures and documentation only.

It does not authorize payroll calculation code or a production business-rule engine.

The current source-verification and dossier gate remains authoritative.

## Final principle

**No business rule enters the Go implementation merely because an AI agent found a plausible formula. It enters only after its sources, assumptions, formalization, derivation and independent expected results have been established and recorded.**

# 9. My final recommendation

I would adopt this as the long-term engineering methodology, with one very important qualification.

Do not build a massive research framework before building the first verified rule.

Use `SG-MOM-OT-001` as the complete reference experiment. Run the entire process manually with reproducible source artifacts, manifests, Phys-Pseudo, independently calculated fixtures and review records.

Once that works, identify which steps genuinely deserve automation.

Eventually, the same workflow can be reused for CPF, leave, inventory calculations, insurance eligibility, accounting rules, SSO policies and customer-specific business logic.

That is the bigger opportunity: WorkforceOps Lab won't merely contain business rules. It will have a disciplined way to acquire, understand, formalize, test and verify them before they become executable software.

We should aim to make every important business rule as traceable and reproducible as a well-constructed scientific derivation.


## prompt

just to clarify so we are aligned, we are just building a real backend for payroll, WFM and HR with populated realistic data (synthetic but based on real life-like scenarios), what aspiring Tier 2 support engineers or business owners can do next is up to them but that is the foundation

## answer

Yes. That is an important clarification, and I agree. I was allowing the OJT and training vision to expand the initial product scope too much.

The product we are building first is a real, functioning Singapore HR, WFM and payroll backend, populated with realistic synthetic business data.

Everything else is a potential use of that foundation.

## 1. The actual product

# WorkforceOps Lab

## A realistic, executable enterprise workforce-management system.

HR backend

Employees, employment contracts, compensation, employee lifecycle, effective-dated records.

WFM backend

Rosters, shifts, attendance, clock events, leave, timesheets and overtime.

Payroll backend

Salary calculations, supported statutory deductions, payroll runs, payslips and reconciliation.

Integration backend

APIs, synchronization, mappings, processing jobs, failures and audit trails.

Foundation: Go · PostgreSQL · realistic synthetic data · executable Singapore business rules · web interface · REST API.

Local PostgreSQL is mandatory. Neon and Fly.io are optional deployment targets.

These must be functioning business systems, not mock APIs, static dashboards or collections of simulated tickets.

For example, creating an employee in HR must have consequences for employment records, WFM eligibility, scheduling and eventually payroll. Approved attendance must flow through timesheets and payroll processing.

## 2. What makes the backend realistic?

Three things.

Real business logic: Rules must reflect actual Singapore employment and payroll requirements, within explicitly defined and verified applicability boundaries.

Real data relationships: Employees, contracts, shifts, attendance, leave, payroll transactions and payslips must be internally consistent and traceable.

Real operational behavior: The system must execute transactions, process asynchronous jobs, integrate with other systems, preserve historical records and exhibit realistic failure modes.

Synthetic data is perfectly appropriate. The important thing is that it represents believable business activity rather than arbitrary random records.

## 3. Where Phys-Research and Phys-Pseudo fit

I would retain both, but strictly as our engineering methodology.

They help us establish that the backend implements the intended business rules correctly.

The workflow is:

1. Acquire authoritative Singapore legislation and government documentation.
2. Extract business rules, conditions, exceptions and effective dates.
3. Formalize high-risk rules using Phys-Pseudo.
4. Independently derive expected calculations and edge cases.
5. Create golden tests and an independent reference oracle.
6. Implement and test the Go business logic.

This is how we prevent the backend from quietly implementing incorrect payroll rules.

Phys-Pseudo, the research harness and the oracle are engineering tools. They do not need to become product features exposed to ordinary users.

## 4. What we should NOT make prerequisites

I would explicitly avoid making these mandatory for the foundation:

- A complete OJT curriculum.
- AI mentors and interviewers.
- Candidate assessment and scoring.
- A public skills marketplace.
- A complete incident-training catalogue.
- Automated anonymization.
- Arbitrary custom business-rule programming.
- SSO and enterprise identity laboratory.
- Customer billing and subscription management.

We should preserve architectural extensibility where sensible, but not build these capabilities before the underlying business systems work.

## 5. What the foundation enables later

Once the actual backend exists, different users can decide what to do with it.

| Potential user           | Possible use                                                    |
| ------------------------ | --------------------------------------------------------------- |
| Aspiring Tier 2 engineer | Practise SQL, investigate incidents and prepare for interviews  |
| Support team             | Train engineers using realistic business transactions           |
| Business owner           | Explore workflows and test business configurations              |
| SaaS developer           | Test APIs, integrations and business logic                      |
| QA engineer              | Test calculations, boundary conditions and regression scenarios |
| Enterprise customer      | Build a customized staging environment                          |

These are downstream applications of the same foundation, not separate products we must build immediately.

## 6. The product definition I would now freeze

> WorkforceOps Lab is an open-source Go-based enterprise HR, workforce-management and payroll application implementing realistic Singapore business rules and operational workflows, backed by PostgreSQL and populated with internally consistent synthetic business data.
>
> Its purpose is to provide a functioning, extensible business-system environment that individuals and organizations can operate, inspect, test, customize and extend.
>
> The platform is designed to support future applications including Tier 2 support engineering training, OJT, technical assessment, interview preparation, business staging, integration testing and customer-specific business environments.

The ultimate acceptance test for our first foundation is simple:

Can we create a fictional Singapore employer, hire employees, define their employment terms, schedule their work, record attendance, approve timesheets, process payroll, calculate supported statutory obligations and generate correct payslips—all through a functioning Go application backed by PostgreSQL?

And can we independently demonstrate that the business rules and calculations are correct?

That is what I want us to build first. Everything else can grow from it.


## prompt

and to be more relatively complex add in WFM component for our model employer as typical BPO company with multiple shifts and all the nuances of WFM SaaS but as minimal viable backend, I want that baked in to the design stage


## answer

I agree. In fact, I would make a Singapore-based BPO/contact-center company the reference employer, rather than an ordinary company with a simple employee attendance system.

That gives us a much richer and more realistic WFM environment without needing to build an enormous enterprise product.

A typical BPO has to manage two interconnected problems:

1. Workforce and employment: Who works, when they work, what they are entitled to, and how their hours flow into payroll.
2. Client-service delivery: How many agents are needed for each client, campaign, channel and time interval, whether the right skills are available, and whether actual staffing meets the service requirements.

These are related, but they must not be confused. Statutory employment rules govern the employee. Client service-level agreements govern the BPO's delivery obligations.

The distinction is essential to our data model.

MOM's Singapore guidance specifically accommodates different working arrangements, including shifts of up to 12 hours averaged over a continuous three-week period, and a special 30-hour continuous rest period for shift workers.&#x20;

[image](https://www.google.com/s2/favicons?domain=https://www.mom.gov.sg\&sz=32)

mom.gov.sg

&#x20;Meanwhile, modern contact-center WFM systems connect demand forecasting, capacity planning, scheduling and intraday management, often using channel/queue-level demand and average handling time.&#x20;

[image](https://www.google.com/s2/favicons?domain=https://learn.microsoft.com\&sz=32)

Microsoft Learn+2



# 1. The reference employer we should model

# Meridian CX Services Pte. Ltd.

Fictional Singapore BPO

Business: Outsourced customer service and business-process operations.

Operating model: Multiple clients, campaigns, channels, queues, skills and shifts.

Illustrative client portfolio

| Client programme    | Operating coverage | Channels           |
| ------------------- | ------------------ | ------------------ |
| Helios Commerce     | Extended hours     | Voice, chat, email |
| Northstar Financial | 24/7               | Voice, chat        |
| Atlas Travel        | 24/7               | Voice, email       |

Illustrative workforce

- 80–100 employees.
- Multiple supervisors and team leaders.
- Different language and product skills.
- Day, evening and overnight shifts.
- Rotating schedules and rest days.
- Full-time and part-time employees.
- Employees assigned to different client programmes.
- Employees who can support multiple channels or queues.

These are synthetic design assumptions. We will make client service hours, shift patterns and staffing requirements configurable rather than hardcoding them as universal BPO practices.

# 2. The WFM component must be a genuine BPO workforce system

I would organize the domain into six interconnected capabilities.

A. Client and service operations

Clients, outsourcing contracts, programmes, campaigns, contact queues, voice/chat/email channels, operating hours and service-level targets.

B. Demand and capacity planning

Historical contact volume, average handling time, demand forecasts, interval staffing requirements, planned shrinkage and staffing coverage.

C. Scheduling and rostering

Multiple shifts, rotating patterns, shift segments, skill requirements, break scheduling, rest days, shift swaps, overtime and roster publication.

D. Timekeeping and attendance

Clock events, actual agent states, late arrivals, absences, break deviations, cross-midnight shifts, attendance corrections and timesheets.

E. Real-time operations and adherence

Compare published schedules with actual activity, monitor staffing gaps, identify skill shortages, record schedule exceptions and track queue performance.

F. Payroll integration

Approved timesheets, normal hours, eligible overtime, activity/pay-code mappings and payroll-period exports to PayLoop.

These six capabilities give us a recognizable BPO WFM architecture.

## 3. The most important BPO-specific complexity

I want the data model to support the following situations from the beginning.

| WFM concept                     | Realistic business problem                                                                 |
| ------------------------------- | ------------------------------------------------------------------------------------------ |
| Multi-client operations         | Employees support different outsourcing clients                                            |
| Campaigns and queues            | Different clients have different work volumes and service requirements                     |
| Multi-channel operations        | An agent handles voice, chat or email depending on qualification                           |
| Skills-based routing            | Having 20 available employees does not mean 20 can handle a particular language or product |
| Rotating shifts                 | Employees change between day, evening and night patterns                                   |
| Split shifts and shift segments | One rostered shift contains different activities                                           |
| Cross-midnight shifts           | A shift crosses a Singapore calendar or payroll-period boundary                            |
| Break and meal planning         | Break schedules may create temporary staffing shortages                                    |
| Leave and absence               | A published roster becomes understaffed after employees become unavailable                 |
| Schedule adherence              | Planned activity differs from actual agent activity                                        |
| Intraday changes                | Supervisors move breaks, reassign agents or request additional staffing                    |
| Overtime                        | Actual working hours differ from the published roster                                      |
| Integration failures            | HR, WFM, contact-routing and payroll contain inconsistent records                          |
| Timezone differences            | Singapore employees serve customers in other time zones                                    |

A particularly important distinction:

Scheduled hours, actual attendance, agent activity and paid working hours are four different things.

For example, an employee might be rostered for eight hours, clock in for eight hours and spend only six and a half hours handling customer contacts because the remaining time consists of breaks, training, meetings and administrative work.

That is not necessarily an attendance or payroll discrepancy.

The system must distinguish these concepts rather than store one generic `hours_worked` value.

## 4. How the business process should operate

The WFM engine must connect customer demand to actual workforce activity.

## BPO operational lifecycle

Client contracts and service requirements

Client programmes, channels, service windows, target service levels and required skills.

Forecast and capacity planning

Forecast contact volume and handling time; establish staffing requirements per queue, skill and interval.

Shift planning and roster publication

Assign qualified agents, shifts, breaks, activities and rest days while validating schedule constraints.

Actual operations

Simulated contact volumes, agent login states, availability, handling, after-call work and breaks.

Intraday management

Detect understaffing, adherence deviations, unplanned absence and service-performance problems.

Attendance and timesheet approval

Reconcile scheduled time against clock events, approved exceptions and actual working time.

Payroll

Transfer supported pay components, approved hours and overtime transactions into PayLoop.

This is consistent with the planning cycle described in Microsoft Dynamics 365's WFM documentation: forecast volume and handling time by queue and channel, calculate capacity requirements, then translate those requirements into schedules. Intraday management provides the feedback loop.&#x20;

[image](https://www.google.com/s2/favicons?domain=https://learn.microsoft.com\&sz=32)

Microsoft Learn+2



# 5. What I would actually build in the MVP

I want to avoid two extremes:

- A basic attendance application that we call WFM.
- An attempt to recreate NICE, Verint or a complete contact-center workforce optimization platform.

We should build a minimal but genuinely functional BPO WFM backend.

| Capability              | MVP scope                                                      |
| ----------------------- | -------------------------------------------------------------- |
| Multi-client operations | Multiple client accounts and programmes                        |
| Campaigns and queues    | Client-specific queues and service windows                     |
| Channels                | Voice, chat and email                                          |
| Skills                  | Agent qualifications, proficiency and queue eligibility        |
| Demand planning         | Versioned interval forecasts and actual workload               |
| Capacity planning       | Required staffing per interval and skill                       |
| Scheduling              | Shift templates, rotations and assigned rosters                |
| Shift activities        | On-queue, breaks, meals, training, coaching and administration |
| Roster management       | Draft, publish, revise and approve shift swaps                 |
| Timekeeping             | Clock events, attendance, corrections and exceptions           |
| Adherence               | Compare planned activities with simulated actual agent states  |
| Leave and overtime      | Availability, approvals and approved hours                     |
| Timesheets              | Reconciled, approved working time                              |
| Payroll integration     | Pay-code mapping and export to PayLoop                         |

### What I would deliberately defer

Advanced Erlang C forecasting, machine-learning demand prediction, automatic roster optimization, shift bidding, sophisticated multi-skill optimization and real Genesys/NICE/Verint integrations.

We should design the interfaces for these capabilities but not implement them prematurely.

For the first release, provide deterministic, importable or generated 15-minute demand intervals. This is consistent with the interval-level forecasting and planning approach documented by Microsoft Dynamics 365, although 15-minute intervals should be a configurable modelling choice rather than a universal requirement.&#x20;

[image](https://www.google.com/s2/favicons?domain=https://learn.microsoft.com\&sz=32)

Microsoft Learn



The capacity engine should calculate transparent workload measures and compare the roster against explicit staffing requirements. It must not pretend to predict service-level performance accurately using a simplistic workload formula.

# 6. The domain model needs to change now

I would ask the coding agent to add the following entities to the design.

| Domain                    | Core entities                                                                                    |
| ------------------------- | ------------------------------------------------------------------------------------------------ |
| BPO operations            | Client, client contract, service programme, campaign, channel, queue, service window, SLA target |
| Agent capabilities        | Employee skill, skill level, language, proficiency, qualification validity                       |
| Demand                    | Forecast, forecast interval, actual workload, handling-time metrics                              |
| Capacity                  | Staffing requirement, skill requirement, interval coverage                                       |
| Scheduling                | Shift template, rotation pattern, roster, roster revision, shift assignment                      |
| Activity                  | Shift segment, activity type, break, meal, training, coaching, meeting                           |
| Availability              | Leave, availability constraint, shift swap, absence                                              |
| Timekeeping               | Clock event, attendance record, attendance exception                                             |
| Contact-center operations | Agent state event, queue activity, intraday metric                                               |
| Compliance                | Applicable working-hours rule, rest-day record, break compliance assessment                      |
| Payroll handoff           | Approved timesheet, payroll export, pay-code mapping                                             |

There should be a clear separation of ownership:

- HR owns employees and employment terms.
- ShiftHub owns WFM operations, scheduling, attendance and timesheets.
- PayLoop owns payroll periods, calculations, payroll records and payslips.
- The integration layer transfers approved facts between systems.

One employee may belong to multiple qualified skills and serve multiple client programmes, but every assignment must respect their eligibility and the employer's configuration.

### Two timezones must be supported

This is especially important for a Singapore BPO serving overseas clients.

An employee's employment and statutory time calculations use the Singapore employment context. A client programme may operate according to another timezone.

Store UTC timestamps alongside explicit IANA timezone identifiers and retain the original local business date/time context.

Do not let a client-service timezone silently change Singapore payroll or statutory working-time calculations.

# 7. The Singapore employment rules must interact with WFM scheduling

This is where Phys-Research and Phys-Pseudo become especially valuable.

MOM guidance covers different arrangements for shift workers, breaks and rest periods. For example, shifts up to 12 hours may be subject to averaging over three continuous weeks, while the rest-day provision for shift workers has its own 30-hour continuous-period treatment.&#x20;

[image](https://www.google.com/s2/favicons?domain=https://www.mom.gov.sg\&sz=32)

mom.gov.sg



Our design must therefore capture:

- Employment classification.
- Contractual working arrangement.
- Scheduled working hours.
- Actual working hours.
- Break and meal intervals.
- Rest-day assignment.
- Weekly and multi-week working-hour aggregates.
- Overtime approvals.
- Effective dates of schedule and contract changes.

We should not implement every statutory calculation now. However, the data model must not make those calculations impossible later.

In particular, defer rest-day payment calculations as previously agreed, but do not omit rest-day scheduling and its underlying attendance facts.

# 8. New architectural instruction for the coding agent

I would add the following amendment to the current WorkforceOps design baseline. This is a design-stage assignment: the agent should update the architecture and implementation plan before proceeding with the affected business-system implementation.

# ARCHITECTURE AMENDMENT 4

# WorkforceOps Lab — Singapore BPO Multi-Shift Workforce Management Reference Environment

**Applies to:** WorkforceOps Lab platform architecture, domain model and MVP specification\
**Reference module:** `sg-wfm-payroll`\
**Model employer:** Meridian CX Services Pte. Ltd. (fictional)\
**Technology:** Go, PostgreSQL, local-first deployment, optional Neon/Fly.io\
**Scope:** Architecture and design-stage incorporation

## 1. Objective

Extend the Singapore HR/WFM/payroll reference environment to model a realistic Business Process Outsourcing (BPO) employer operating multiple client programmes, campaigns, channels, queues, skills and shifts.

The WFM component must be a genuine workforce-management business system, not merely an employee attendance or shift-calendar module.

The model employer operates contact-center and customer-service outsourcing operations for multiple clients.

The system must connect:

- Employee and employment management.
- Client service requirements.
- Workforce demand and capacity planning.
- Skills-based workforce allocation.
- Shift scheduling and roster management.
- Breaks, meals and shift activities.
- Timekeeping and attendance.
- Agent activity and schedule adherence.
- Leave, absence and overtime.
- Approved timesheets.
- Payroll integration.

This capability must be included in the domain model, database design, API contracts, data-generation strategy, scenario catalogue and MVP acceptance criteria.

**Do not postpone the fundamental BPO WFM data model until after the HR/payroll backend is built.**

## 2. Model employer

Use the fictional Singapore company:

`Meridian CX Services Pte. Ltd.`

Design an initial environment containing:

- Approximately 80–100 employees.
- Multiple departments and operating teams.
- Multiple team leaders and supervisors.
- Three fictional client programmes.
- Multiple service queues.
- Voice, chat and email channels.
- Different language and product skills.
- Day, evening and overnight shifts.
- Rotating schedules.
- Full-time and part-time employment arrangements.
- Multiple work activity types.
- Multiple payroll periods and historical workforce activity.

At least one client programme should require round-the-clock coverage, while others should have configurable operating windows.

These are simulation requirements, not claims about universal BPO arrangements.

All employees, clients, workloads, service agreements and operational records must be synthetic.

## 3. Separate employee employment from BPO service operations

The architecture must distinguish two related but different business domains.

### Employment and workforce obligations

Determine:

- Who the employee is.
- Their employment classification and contract.
- Their applicable working arrangement.
- Their salary and approved compensation terms.
- Their working hours.
- Their breaks and rest days.
- Their attendance.
- Their overtime and leave.
- Their approved paid hours.

### Client service delivery

Determine:

- Which client the employee supports.
- Which programme, campaign or queue they are assigned to.
- Which skills are required.
- What demand exists during each interval.
- How many qualified agents are required.
- Whether the published roster provides adequate coverage.
- Whether actual agent activity follows the schedule.
- Whether client service targets are being met.

Client SLA targets and staffing objectives must never be treated as Singapore statutory employment rules.

## 4. Required BPO WFM capabilities

### 4.1 Client operations

Model:

- Clients.
- Client contracts.
- Service programmes.
- Campaigns.
- Contact channels.
- Work queues.
- Service windows.
- Client-specific timezones.
- Service-level targets.
- Queue and campaign assignments.

Allow one BPO employer to operate several client programmes concurrently.

A client programme may contain multiple campaigns, queues and channels.

### 4.2 Agent skills and eligibility

Model:

- Employee skills.
- Language proficiency.
- Product and process qualifications.
- Skill levels.
- Certification validity periods where applicable.
- Queue eligibility.
- Campaign eligibility.
- Team membership.
- Supervisor relationships.

An employee being present and available does not automatically mean they can be assigned to every queue.

The coverage calculation must account for the required skills.

### 4.3 Forecast and demand planning

Model workload and staffing requirements at configurable time intervals.

Use 15-minute intervals as the initial contact-center reference granularity.

Capture:

- Forecast contact volume.
- Actual contact volume.
- Average handling time.
- Offered and handled contacts.
- Abandoned contacts.
- Queue or channel.
- Required staffing.
- Skill requirements.
- Forecast version and source.
- Actual-versus-forecast variance.

Support generated or imported forecasts.

For the MVP, staffing requirements may be supplied through seeded data or operator configuration.

Implement transparent workload calculations and basic coverage comparisons.

Do not claim accurate voice service-level forecasting from a simplistic workload calculation.

Defer advanced Erlang C calculations, machine-learning forecasting and automated workforce optimization.

### 4.4 Shift scheduling and rostering

Model:

- Shift templates.
- Shift assignments.
- Rotating patterns.
- Published rosters.
- Roster revisions.
- Shift swaps.
- Shift change approvals.
- Rest-day assignments.
- Break and meal placement.
- Client and campaign allocation.
- Skill requirements.
- Scheduled activity segments.

A rostered shift must be capable of containing multiple segments:

- On-queue work.
- Offline administrative work.
- Paid breaks.
- Unpaid meal breaks where applicable.
- Training.
- Coaching.
- Team meetings.
- Other employer-configured activities.

Support:

- Day shifts.
- Evening shifts.
- Overnight shifts.
- Cross-midnight shifts.
- Repeating patterns.
- Configurable part-time schedules.
- Multiple working-day arrangements.

Store a shift as a logical shift instance with its own segments.

Do not group all working time merely by the calendar date on which clock-in occurred.

### 4.5 Attendance and timekeeping

Keep the following distinct:

1. Published schedule.
2. Raw clock events.
3. Interpreted attendance.
4. Actual agent activity.
5. Approved timesheet.
6. Payroll input.

Implement:

- Clock-in and clock-out.
- Duplicate-event protection.
- Missing punches.
- Late arrival.
- Early departure.
- Break deviations.
- Absence.
- Attendance corrections.
- Supervisor approval.
- Cross-midnight attendance.
- Schedule-to-attendance reconciliation.

Preserve raw source events and record corrections as auditable events rather than silently overwriting historical facts.

### 4.6 Agent states and schedule adherence

Provide a simulated contact-center agent-state feed.

Examples include:

- Available.
- On queue.
- Busy/in interaction.
- After-call work.
- Break.
- Meal.
- Training.
- Coaching.
- Meeting.
- Administrative work.
- Offline.

Compare scheduled activities with actual agent-state events.

Calculate basic adherence intervals and record deviations.

The system should support explanations such as:

- Agent was scheduled for a break but remained on queue.
- Agent was expected to be on queue but was not logged in.
- Agent was assigned to the wrong client programme.
- Agent was available but did not possess the required skill.
- A schedule change was approved but the actual activity system did not receive it.

Use synthetic agent-state events in the initial environment. Real telephony or ACD integrations are future extensions.

### 4.7 Intraday operations

Provide basic operational visibility into:

- Required versus scheduled staffing.
- Qualified versus unqualified coverage.
- Planned versus actual staffing.
- Absence and adherence deviations.
- Demand versus actual workload.
- Queue staffing gaps.
- Unexpected workload increases.

Allow an authorized operator to record an intraday change, such as a break reschedule or agent reassignment.

Every schedule change must be versioned and auditable.

Do not build a full real-time workforce optimization engine.

### 4.8 Payroll handoff

Approved timesheets must supply PayLoop with the supported payroll inputs.

Maintain traceability between:

- Scheduled hours.
- Actual attendance.
- Approved work hours.
- Normal hours.
- Eligible overtime.
- Paid and unpaid activities.
- Pay-code mappings.
- Export records.
- Payroll import records.
- Final payslip lines.

Do not assume all on-queue time and all paid working time are the same.

Payroll must consume approved source facts, not raw contact-center metrics.

## 5. Singapore working-time considerations

The WFM architecture must accommodate Singapore-specific employment requirements.

Research and formally specify applicable provisions before implementing statutory validations.

Relevant areas include:

- Part IV applicability.
- Normal contractual hours.
- Shift-worker arrangements.
- Break and meal intervals.
- Daily working-hour limits.
- Overtime.
- Weekly rest days.
- Shift-worker rest-period arrangements.
- Applicable monthly averaging periods.
- Effective-dated employment changes.

Capture sufficient scheduling and attendance data to validate supported working arrangements.

Apply Phys-Pseudo to high-risk statutory rules and preserve unresolved conditions explicitly.

Important restrictions:

- Do not assume all BPO employees qualify for Part IV.
- Do not assume all employees work the same shift arrangement.
- Do not confuse client coverage hours with an employee's statutory working hours.
- Do not assume Sunday is every employee's rest day.
- Do not automatically treat every schedule extension as statutory overtime.
- Rest-day and public-holiday payment calculations remain deferred under the existing MVP scope.
- Unsupported statutory arrangements must return `indeterminate`.

## 6. Timezone and temporal requirements

The employer is based in Singapore but may serve overseas client operations.

Support separate:

- Employment timezone.
- Site timezone.
- Client service timezone.
- Queue operating timezone.

Use UTC instants with explicit IANA timezone identifiers and preserve original local business-date context.

Pay periods and Singapore employment calculations must not silently inherit the client's service timezone.

Support cross-midnight shifts and daylight-saving transitions where client-service geography requires them.

Continue to distinguish business effective time from database recorded time for employment, roster and schedule revisions.

## 7. Minimum viable WFM implementation

The MVP must implement:

1. Multiple clients and programmes.
2. Campaigns, queues and service channels.
3. Agent skill and eligibility records.
4. Interval-based workload and staffing requirements.
5. Shift templates and rotating assignments.
6. Roster publication and revisions.
7. Shift activity segments.
8. Breaks, meals and rest-day scheduling records.
9. Leave and availability impact.
10. Clock events and attendance.
11. Simulated agent-state events.
12. Basic schedule adherence.
13. Attendance and timesheet approval.
14. WFM-to-payroll data transfer.

Advanced demand forecasting, Erlang C staffing models, automatic scheduling optimization, shift bidding, advanced adherence analytics and real ACD integrations remain deferred.

The implementation must be sufficient to run a real end-to-end business workflow rather than returning static mock responses.

## 8. Synthetic data and acceptance scenarios

Update the synthetic data generator to create:

- Multiple client programmes.
- Different queue service windows.
- A realistic interval demand pattern.
- Agents with different skills.
- Several shift templates.
- Rotating schedules.
- Overnight and cross-midnight shifts.
- Planned breaks and training segments.
- Actual agent-state events.
- Attendance exceptions.
- Approved timesheets.
- Payroll history.

Include WFM-specific candidate incidents in addition to the existing integration incidents.

At minimum, design scenarios for:

- Queue understaffing caused by incorrect capacity assumptions.
- Wrong skill assignment despite apparent headcount availability.
- Break overlap causing a staffing coverage gap.
- Approved shift changes not reaching the activity system.
- Actual agent activity differing from published schedules.
- Cross-midnight attendance being incorrectly attributed.
- Leave approval not updating workforce capacity.
- Approved overtime missing from payroll.

Each scenario must identify its business consequences, evidence sources, causal chain and expected result.

Use the established evidence classifications and distinguish documented behavior from engineered synthetic fault mechanisms.

## 9. Architecture documents to update

Before business implementation, update:

- `architecture.md`
- `domain-model.md`
- `database-design.md`
- `user-workflows.md`
- `api-contracts.md`
- `module-contract.md`
- `module-manifest-spec.md`
- `environment-manifest.md`
- `scenario-format.md`
- `oracle-design.md`
- `rule-source-register.md`
- `implementation-roadmap.md`

Add:

- `bpo-reference-employer.md`
- `wfm-domain-specification.md`
- `wfm-scheduling-model.md`
- `wfm-demand-and-capacity-model.md`
- `wfm-agent-state-and-adherence.md`
- `wfm-to-payroll-data-lineage.md`
- `wfm-acceptance-scenarios.md`

Include an entity-relationship model and complete business-data lifecycle.

## 10. Required design demonstrations

Before approving the design, demonstrate these end-to-end conceptual workflows:

**A. Demand to roster**

Client demand → interval staffing requirement → qualified-agent coverage → published schedule.

**B. Roster to actual activity**

Published schedule → agent clock events → simulated agent states → adherence exceptions.

**C. WFM to payroll**

Attendance → approved timesheet → supported overtime and pay-code processing → PayLoop payroll inputs.

**D. Multi-client staffing**

One employee with multiple skills can be assigned to different programmes, subject to qualifications and availability.

**E. Shift exception**

An absent employee creates a staffing gap; an operator records a replacement assignment and preserves the roster revision history.

**F. Cross-midnight shift**

A shift crossing a Singapore calendar or payroll boundary preserves its shift identity, timestamps, activity segments and relevant business-date attribution.

**G. Service versus employment rules**

Demonstrate that a queue failing a client SLA is an operational service problem, not automatically an employment-law or payroll violation.

## 11. Scope discipline

Do not allow the BPO WFM expansion to turn the MVP into a complete commercial contact-center product.

Do not implement:

- Full call-routing infrastructure.
- Actual PSTN/telephony systems.
- Production Genesys/NICE/Verint integrations.
- Advanced AI demand forecasts.
- Full automated roster optimization.
- Every possible shift and employee arrangement.
- Deferred statutory payment calculations.

The platform must provide clear extension boundaries for these capabilities.

## Final directive

Design the Singapore reference employer as a realistic multi-client BPO company with genuine workforce-management complexity.

The WFM system must connect customer demand, skill requirements, shift schedules, actual activity, attendance, adherence, approved timesheets and payroll.

Build the minimum viable operational foundation that can execute and persist these workflows.

Do not reduce WFM to a shift calendar.

Do not confuse operational service targets with statutory employment obligations.

The result must be an extensible Go business system with realistic data, traceable transactions and enough operational complexity to reproduce meaningful enterprise support problems.

# 9. The most important design decision

I would make one thing non-negotiable:

WFM must have its own coherent business lifecycle, rather than merely being a collection of tables attached to HR or payroll.

The key relationship is:

\\[ \text{Client demand} \rightarrow \text{Required capacity} \rightarrow \text{Qualified workforce} \rightarrow \text{Published roster} \rightarrow \text{Actual activity} \rightarrow \text{Attendance} \rightarrow \text{Approved timesheet} \rightarrow \text{Payroll} \\]

That gives us realistic operational and data-integrity problems at every boundary.

It also creates a much stronger reference backend than a conventional employee-management and payroll demonstration.

My recommendation is to incorporate this amendment before freezing the Phase 2 domain model. The BPO WFM lifecycle, entities and time semantics must be designed alongside HR and PayLoop—not added afterwards.

# PLAN TO DO THIS

## prompt

now let us get disciplined, trim down the features to bare essentials for our integrated HR, payroll, WFM system per Singapore business rules law, to make it minimum viable backend, do not add any UI, let us make it CLI and API based only


## answer

I recommend we make a hard cut. WorkforceOps Lab v0.1 should be one integrated Go backend that takes an employee from hiring, through scheduling and attendance, all the way to a completed monthly payroll run.

Not a training platform yet. Not a WFM SaaS clone. Not a collection of microservices. First, build one functioning business system.

# 1. Freeze the MVP definition

Product: WorkforceOps Lab\
Purpose: A Singapore-based BPO company's essential HR, workforce management and payroll backend.

The single end-to-end workflow is:

HR

Employee records · Employment terms · Leave

WFM

Client programme · Shift roster · Attendance

Timesheets

Actual hours · Overtime · Corrections · Approval

Payroll

Gross pay · Statutory deductions · Payslips

Payroll outputs

Payroll register · Employee payslips · Statutory summaries

Everything is connected. We do not build three disconnected demonstrations.

# 2. The actual MVP features

## A. HR — essential employee administration

Build only:

1. Employer and employee records.
2. Employment contracts and effective-dated employment terms.
3. Employee status: active, suspended, resigned or terminated.
4. Job title, employment category, basic salary, fixed allowances and deductions.
5. Employment start/end dates and contractual working arrangements.
6. Annual leave and sick leave records, balances and approvals.
7. Employment record and KET information.

The system must retain the information needed to produce employment records and itemised payslips. MOM specifies employee and salary record requirements, including working hours, leave, salary components and overtime. It also specifies record-retention periods.&#x20;

[image](https://www.google.com/s2/favicons?domain=https://www.mom.gov.sg\&sz=32)

mom.gov.sg+1



Do not build: recruitment, applicant tracking, performance management, benefits administration, employee engagement or learning management.

## B. WFM — enough to operate a multi-shift BPO

Keep the BPO structure very small:

`Client → Programme → Team → Employee`

Add a simple skill assignment where it is needed to prevent an unqualified employee from being assigned to a programme.

The MVP needs:

- Multiple clients and programmes.
- Teams and employee programme assignments.
- Day, evening and night shift templates.
- Weekly and rotating rosters.
- Cross-midnight shifts.
- Paid and unpaid break segments.
- Designated rest days.
- Leave and availability affecting roster eligibility.
- Published rosters and controlled revisions.
- Clock-in and clock-out records.
- Attendance corrections with reasons and audit history.
- Supervisor-approved timesheets.

A basic minimum-headcount requirement per programme/shift is enough to identify obvious staffing gaps.

Do not build:

- AI demand forecasting.
- Erlang C or advanced capacity prediction.
- Automatic roster optimization.
- Intraday agent-state monitoring.
- Voice routing or telephony integration.
- Real-time adherence dashboards.
- Advanced multi-skill optimization.
- Shift bidding or workforce marketplace features.

We need to distinguish four things in the data model: scheduled time, recorded attendance, actual working time and paid time. They must not be represented by one `hours_worked` field.

## C. Payroll — the minimum usable payroll engine

The monthly payroll engine should support:

1. Monthly basic salary.
2. Fixed allowances.
3. Approved overtime.
4. Paid and unpaid leave treatment within the supported rule scope.
5. Employee statutory deductions.
6. Employer statutory contributions and levies.
7. Gross salary, deductions and net salary.
8. Payroll approval and locking.
9. Itemised payslip records.
10. Payroll register and statutory calculation exports.

The system must preserve the calculation inputs and results for every payroll run. Once a run is approved, changes should require an auditable correction or adjustment rather than silently overwriting it.

### One important correction to our previous plan

We previously deferred the CPF calculation engine. That decision cannot stand if we intend to call this a genuine Singapore payroll MVP.

CPF is mandatory for eligible Singapore Citizen and Permanent Resident employees earning more than S$50 monthly. The applicable rates vary with age, citizenship/residency status and wages. The Ordinary Wage ceiling is S$8,000 from January 2026, and further rate changes are scheduled for January 2027.&#x20;

[image](https://www.google.com/s2/favicons?domain=https://www.cpf.gov.sg\&sz=32)

How much CPF contributions to pay+2



SDL is also compulsory for employees working in Singapore, including foreign employees. Its current calculation is 0.25% of monthly total wages, subject to a S$2 minimum and S$11.25 maximum per employee, with the aggregate rounded down to the nearest dollar.&#x20;

[image](https://www.google.com/s2/favicons?domain=https://www.cpf.gov.sg\&sz=32)

Skills Development Levy



Therefore:

- CPF must be in the payroll MVP.
- SDL must be in the payroll MVP.
- Applicable self-help-group deductions must be accounted for before declaring net pay complete.
- Rates and thresholds must be effective-dated rule data, not hardcoded constants.
- CPF submission integration, actual bank transfers and IRAS filing can wait.

We also need to preserve the distinction between MOM employment rules and CPF Board contribution rules. They are separate rule sources.

# 3. Narrow the legal scope deliberately

Trying to implement every Singapore employment and payroll situation in the first release would defeat the purpose of an MVP.

I recommend an explicit initial support profile:

| Area            | Initial support boundary                                                              |
| --------------- | ------------------------------------------------------------------------------------- |
| Employer        | One Singapore legal entity                                                            |
| Employees       | Direct employees; Singapore Citizen reference population                              |
| Pay             | Monthly-rated employees                                                               |
| Salary          | Basic salary and fixed allowances                                                     |
| Work            | Regular day/evening/night shifts                                                      |
| Scheduling      | Cross-midnight shifts and rotating rosters                                            |
| Hours           | Ordinary working-hour arrangements; Part IV eligibility recorded explicitly           |
| Overtime        | Verified ordinary overtime calculation for eligible employees                         |
| Leave           | Annual and sick leave                                                                 |
| Rest days       | Recorded and protected; no rest-day work in the initial release                       |
| Public holidays | Calendar and standard paid holiday treatment; no special work-on-holiday arrangements |
| Payroll         | Monthly payroll                                                                       |
| Jurisdiction    | Singapore                                                                             |

This is a restricted, source-verified legal coverage profile, not a claim that the software handles every Singapore employment scenario.

For example, MOM's Part IV overtime provisions do not apply to every employee. Eligibility depends on the employee category and basic salary; the current published thresholds distinguish non-workmen and workmen. We must record and verify that classification rather than infer it from a job title such as Contact Centre Agent.&#x20;

[image](https://www.google.com/s2/favicons?domain=https://www.mom.gov.sg\&sz=32)

mom.gov.sg



There is also one existing blocker we must resolve before implementing overtime: the discrepancy in the S$2,600 non-workman overtime-rate cap and the corresponding hourly-rate calculation identified in our earlier research. The system must not choose whichever number is convenient.

Foreign workers, PR-specific CPF cases, IR21 tax clearance, rest-day work, public-holiday work arrangements and other special employment situations should be explicitly unsupported initially. The API must reject or flag those cases rather than silently calculate them incorrectly.

MOM also specifies that ordinary salary is generally payable within seven days after the salary period and overtime pay within fourteen days. Those deadlines should be represented in the payroll workflow.&#x20;

[image](https://www.google.com/s2/favicons?domain=https://www.mom.gov.sg\&sz=32)

mom.gov.sg



# 4. Architecture: one Go application, not microservices

I would simplify the previous architecture substantially.

One Go executable

workforceops

REST API

JSON · OpenAPI

CLI

Commands · Scripts

Shared application services

HR

WFM

Attendance

Payroll

Singapore rules

Audit

PostgreSQL

One database · One application-owned data model

This means:

- One Go module: `github.com/PithomLabs/a1`.
- One Go application.
- One PostgreSQL database.
- One migration system.
- One REST API.
- One CLI.
- One shared set of application services.

The CLI and API must invoke the same business logic. They must not implement their own payroll calculations or WFM rules.

We do not need ShiftHub API, PayLoop API or a separate synchronization worker for this MVP. Those boundaries can be reconsidered only when actual scaling or integration requirements justify them.

Technology constraints:

| Layer             | Choice                       |
| ----------------- | ---------------------------- |
| Language          | Go                           |
| Database          | PostgreSQL                   |
| Database access   | pgx and explicit SQL         |
| Migrations        | Versioned SQL migrations     |
| API               | REST/JSON                    |
| API contract      | OpenAPI                      |
| CLI               | Single executable            |
| Local environment | Docker Compose               |
| Deployment        | Optional Fly.io / Neon later |
| UI                | None                         |

No React. No HTMX. No Go templates. No frontend application.

# 5. Minimum CLI and API surface

The CLI should expose the business workflows directly.

```
# Environment
workforceops migrate
workforceops seed demo

# HR
workforceops employee create
workforceops employee list
workforceops employee show EMP-001
workforceops employee update EMP-001

# WFM
workforceops client list
workforceops programme list
workforceops roster create
workforceops roster publish
workforceops attendance import
workforceops timesheet review
workforceops timesheet approve

# Payroll
workforceops payroll preview 2026-11
workforceops payroll validate 2026-11
workforceops payroll run 2026-11
workforceops payroll approve 2026-11
workforceops payslip export EMP-001 2026-11
workforceops payroll export 2026-11
```

The API exposes corresponding versioned endpoints for employee, contract, client, programme, roster, attendance, timesheet and payroll operations.

The exact commands and endpoint names can be finalized during the API contract design. The important decision is that both interfaces use the same services.

# 6. What we explicitly remove from v0.1

| Feature                                   | Decision |
| ----------------------------------------- | -------- |
| Any web UI                                | Removed  |
| Microservices                             | Removed  |
| Separate ShiftHub and PayLoop databases   | Removed  |
| Event bus and distributed synchronization | Removed  |
| AI/ML components                          | Removed  |
| RAG, embeddings and vector databases      | Removed  |
| Generic simulation engine                 | Deferred |
| Scenario marketplace                      | Deferred |
| Advanced WFM optimization                 | Deferred |
| Multi-tenant SaaS infrastructure          | Deferred |
| SSO and enterprise identity               | Deferred |
| External HR/payroll integrations          | Deferred |
| Banking and CPF submission automation     | Deferred |
| Applicant/OJT training platform           | Deferred |

We retain a small number of deterministic seeded business cases and test fixtures. We do not build a generic simulation framework yet.

# 7. The MVP acceptance test

The first version is not complete until this single scenario works from beginning to end:

1. Create a Singapore BPO employer.
2. Create two clients and several programmes.
3. Add employees and their employment terms.
4. Assign employees to teams and programmes.
5. Create day, evening and overnight shift templates.
6. Publish a rotating roster.
7. Record clock-in and clock-out events.
8. Correct an attendance exception with an audit trail.
9. Approve employee timesheets.
10. Run monthly payroll.
11. Verify salary, overtime and statutory calculations against independently checked fixtures.
12. Generate an itemised payslip and payroll register.
13. Re-run validation and demonstrate that the approved payroll result is reproducible.

MVP completion criterion

A CLI/API user can hire an employee, schedule that employee, record their actual work, approve their hours and produce a traceable monthly payroll result.

That is the product.

## My recommendation for the next design decision

Freeze this as WorkforceOps Lab MVP v0.1 — Integrated HR, WFM and Payroll Backend.

Before coding, produce only five design documents:

1. `mvp-scope.md` — the supported and unsupported business cases.
2. `domain-model.md` — the minimum HR, WFM and payroll entities and relationships.
3. `singapore-rule-scope.md` — applicable legislation, coverage boundaries and verified rule packs.
4. `database-design.md` — PostgreSQL tables, constraints and effective-dated records.
5. `api-cli-contract.md` — the complete initial API and CLI workflows.

The first implementation milestone should be one complete employee-to-payroll vertical slice, not all the modules independently.

The key discipline is this: we are building the smallest real business system, not the largest possible simulation of one.

