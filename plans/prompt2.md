I would make this a regulatory research + enterprise SaaS operations + adversarial validation investigation, not simply a search for Singapore payroll rules.

The critical distinction is that we need to discover two things:

1. What a Singapore employer's HR, WFM and payroll systems are actually required to do.
2. What goes wrong when real software implements those requirements, moves data between systems, handles exceptions or processes changes.

The research must ultimately produce a set of evidence-backed operational scenarios that can stress-test WorkforceOps Lab's business rules, calculations, database relationships, integration workflows and Tier 2 investigation capabilities.

Here is the Deep Research prompt I would use.

# DEEP RESEARCH MANDATE

# Singapore HR, Workforce Management and Payroll SaaS: Regulatory Accuracy, Real-World Operational Problems and Tier 2 Support Engineering Scenarios

**Research programme:** WorkforceOps Lab\
**Reference business module:** `sg-wfm-payroll`\
**Research jurisdiction:** Singapore\
**Research cutoff:** Current as of the date of research\
**Primary objective:** Discover, verify and formalize real-world Singapore HR, WFM and payroll business rules, operational workflows, software failure patterns and troubleshooting procedures.

---

# 1. Your role and assignment

Act as a multidisciplinary research team comprising:

1. Singapore employment-law and statutory compliance researcher.
2. Singapore payroll specialist.
3. Workforce-management and timekeeping specialist.
4. HRIS and enterprise SaaS architect.
5. Payroll integration engineer.
6. Tier 2 application and production support engineer.
7. Software quality and test-automation engineer.
8. Independent adversarial reviewer.

Conduct deep, evidence-driven research into the actual operational requirements and problems encountered in Singapore HR, workforce management and payroll SaaS environments.

We are developing WorkforceOps Lab, an open-source, Go-based enterprise simulation, staging, OJT and support-engineering assessment platform.

Our first business module simulates a Singapore employer operating HR, WFM and payroll systems.

The ultimate goal is to help aspiring Tier 2 application support, technical support and production support engineers gain practical experience investigating realistic enterprise software problems.

**Do not produce a generic educational overview of Singapore employment law.**

We need research that can be translated into an executable business system and a credible collection of production-like incidents.

Your research must uncover:

- What actual Singapore employers must do.
- How HR, WFM and payroll software implements those obligations.
- What information moves between the systems.
- Where the business processes and software integrations fail.
- How experienced Tier 2 engineers discover the underlying causes.
- How to reproduce these failures without inventing unrealistic business behavior.
- How to independently determine the mathematically and legally correct result.

The outcome must help us determine whether WorkforceOps Lab's rules, computations and simulated incidents are accurate enough to withstand technical scrutiny.

---

# 2. Research objectives

Investigate the following three interconnected business domains.

## A. Human Resources Information Systems (HRIS)

Research actual Singapore employer requirements and corresponding software workflows.

Areas include:

### Employee lifecycle

- Recruitment and onboarding.
- Employee master-data management.
- Employment commencement dates.
- Employment contracts and key employment terms.
- Employment classification.
- Changes to salary and compensation.
- Changes to job title and department.
- Transfers between business locations.
- Promotions and changes to employee classification.
- Permanent and fixed-term employment.
- Part-time and shift-work arrangements.
- Resignation and termination.
- Final salary processing.
- Employment records and retention.
- Foreign employee status and work-pass-related information where relevant.
- Singapore Permanent Resident status and changes in status.

### HR business rules

Research:

- Employment eligibility.
- Contractual entitlements.
- Statutory minimum entitlements.
- Employment record requirements.
- Itemised payslip requirements.
- Salary-payment obligations.
- Leave eligibility and entitlement.
- Employee status changes and effective dates.
- Employer reporting obligations.
- Required notifications and submissions.

### HRIS operational workflows

Investigate how actual software systems handle:

- Employee creation and modification.
- Duplicate employee records.
- Employee identifier mapping.
- Effective-dated employment changes.
- Employee termination and reactivation.
- Historical employee records.
- Data validation.
- Approval workflows.
- Downstream synchronization to WFM and payroll.

Identify which records must be preserved historically and which changes should affect future transactions only.

## B. Workforce Management (WFM)

This domain is particularly important because it connects employment arrangements to actual hours worked and payroll calculations.

Research:

### Scheduling

- Shift patterns.
- Rotating shifts.
- Five-day and six-day workweeks.
- Normal working hours.
- Rest days.
- Break intervals.
- Shift changes.
- Shift swaps.
- Cross-midnight shifts.
- Part-time schedules.
- Irregular working hours.
- Public holiday schedules.
- Overtime scheduling and authorization.

### Attendance and timekeeping

- Clock-in and clock-out.
- Missing punches.
- Duplicate punches.
- Late arrivals.
- Early departures.
- Break deductions.
- Attendance corrections.
- Manager approvals.
- Device synchronization.
- Timezone handling.
- Overnight attendance.
- Manual timesheet adjustments.
- Attendance reconciliation.

### Overtime, rest days and public holidays

Research the actual Singapore statutory treatment, including:

- Employment Act Part IV coverage.
- Workmen and non-workmen distinctions.
- Salary thresholds and applicability.
- Managers and executives exclusions.
- Normal working-hour arrangements.
- Overtime eligibility.
- Overtime hourly basic rate calculations.
- Overtime payment rate.
- Overtime monthly limits.
- Overtime payment deadlines.
- Rest-day work.
- Public holiday work.
- Work performed on days that are both public holidays and rest days where applicable.

Do not generalize the rules into an unconditional calculation based on hours exceeding 44 per week.

Identify the applicable employee categories, statutory conditions, alternative arrangements and exceptions.

### Leave management

Investigate the applicable rules and software workflows for:

- Annual leave.
- Sick leave.
- Hospitalisation leave.
- Childcare leave.
- Parental and other relevant statutory leave.
- Unpaid leave.
- Pro-rated entitlements.
- Service-period calculations.
- Leave accrual.
- Leave carry-forward.
- Leave encashment where applicable.
- Leave balance adjustments.
- Leave approvals and cancellations.

Identify what is statutory, what is contractual and what is employer-configurable.

## C. Payroll and statutory contributions

This must be one of the deepest parts of the research.

We need more than general descriptions of payroll calculations.

We need the precise conditions, data inputs, formulas, effective dates, rounding rules and exception handling required to reproduce correct outcomes.

### Salary computation

Investigate:

- Monthly-rated employees.
- Daily-rated employees.
- Hourly-rated employees.
- Part-time employees.
- Salary periods and cut-off dates.
- Partial-month salary calculations.
- Mid-month hiring.
- Mid-month termination.
- Unpaid leave deductions.
- Salary revisions.
- Backpay and retrospective salary adjustments.
- Allowances.
- Reimbursements.
- Bonuses.
- Commission.
- Deductions.
- Overtime payments.
- Final salary.

### CPF

Research the complete applicable CPF rules and tables.

Include:

- Singapore citizens.
- Singapore Permanent Residents.
- Relevant PR contribution year distinctions.
- Age-band contribution rates.
- Employer contributions.
- Employee contributions.
- Ordinary Wages.
- Additional Wages.
- Ordinary Wage ceilings.
- Additional Wage ceilings.
- Annual limits.
- Wage classification.
- Backpay.
- Retroactive adjustments.
- Contribution rounding.
- Contribution submission timelines.
- Changes in contribution rates and effective dates.
- Relevant exclusions and special circumstances.

Do not simply list headline contribution percentages.

Construct the complete applicability and calculation matrix.

For each supported calculation, identify the exact source table, effective date, wage definition, employee category, rounding convention and boundary conditions.

### Skills Development Levy and other payroll obligations

Investigate:

- SDL applicability.
- SDL calculation.
- Minimum and maximum levy considerations where applicable.
- Relevant employer obligations.
- Other applicable statutory payroll deductions or contributions.
- Community fund contributions and their applicable conditions where relevant.

Clearly separate statutory employer obligations, employee deductions and amounts merely configured by employers.

### IRAS and employment-income reporting

Research:

- Employment income reporting.
- AIS requirements.
- IR8A and relevant appendices.
- IR21 tax clearance.
- Employee cessation and foreign-employee workflows.
- Reporting deadlines.
- Payroll-to-tax reporting reconciliation.
- Submission validation and error handling.

Distinguish payroll calculation obligations from actual tax filing and reporting obligations.

---

# 3. Official Singapore source research

Primary government and statutory sources take precedence over secondary summaries.

Research and pin relevant material from:

### Ministry of Manpower

[https://www.mom.gov.sg/](https://www.mom.gov.sg/)

Prioritize:

- Employment Act.
- Employment practices.
- Working hours and overtime.
- Rest days.
- Public holidays.
- Salary payment.
- Itemised payslips.
- Employment records.
- Annual leave.
- Sick leave.
- Childcare and parental leave.
- Key employment terms.
- Part-time employment.
- Employer reporting and recordkeeping obligations.

### Singapore Statutes Online

[https://sso.agc.gov.sg/](https://sso.agc.gov.sg/)

Locate the actual legislation and subsidiary legislation supporting the relevant MOM guidance.

Where a government webpage summarizes a statutory rule, locate the corresponding legislation where possible.

Do not treat a summary webpage as an adequate replacement for the governing legal text.

### Central Provident Fund Board

[https://www.cpf.gov.sg/](https://www.cpf.gov.sg/)

Research:

- Employer contribution requirements.
- Complete contribution tables.
- Ordinary and Additional Wage treatment.
- Age-band calculations.
- PR contribution treatment.
- Contribution ceilings.
- Submission and payment requirements.
- Effective-dated changes.

### Inland Revenue Authority of Singapore

[https://www.iras.gov.sg/](https://www.iras.gov.sg/)

Research:

- Employer reporting requirements.
- AIS.
- IR8A.
- IR21.
- Tax clearance.
- Employment income reporting.
- Relevant electronic submission specifications and error responses.

### Other relevant government sources

Where relevant, investigate:

- SkillsFuture Singapore for SDL.
- PDPC for personal-data protection and handling.
- MOM and other official agencies for work-pass-related rules.
- Official Singapore public holiday and calendar publications.

Do not assume that every item appearing in a payroll vendor's feature list represents a statutory requirement.

---

# 4. Research actual HR, WFM and payroll SaaS implementations

This is a separate and essential research track.

We need to understand how real enterprise and Singapore-serving SaaS products implement the rules and workflows.

Identify a representative sample of established Singapore-serving products.

Examples to investigate include:

- Employment Hero.
- Talenox.
- QuickHR.
- JustLogin.
- Payboy.
- HReasily.
- Swingvy.
- Info-Tech HRMS.
- Other relevant Singapore WFM, attendance and payroll platforms.

Do not assume that every product is comparable or that every product serves the same business segment.

For each selected vendor, investigate publicly available materials:

- Product documentation.
- Official help centres.
- Knowledge bases.
- Implementation guides.
- API documentation.
- Integration documentation.
- Troubleshooting articles.
- Release notes.
- Known limitations.
- Data import/export specifications.
- Payroll processing documentation.
- Attendance and timesheet workflows.
- Customer support guidance.

The purpose is to learn how real products behave and where their operational complexity lies.

## Important evidence restriction

We do not have access to these vendors' internal production databases, private support tickets or confidential incident histories.

Do not pretend otherwise.

A vendor help article describing a problem is evidence that the vendor documents a troubleshooting path. It is not proof of how frequently that problem occurs.

A customer forum complaint is anecdotal evidence, not a verified general defect.

A job advertisement describing Tier 2 support responsibilities is evidence of expected operational skills, not proof that a specific incident is prevalent.

Label these evidence categories separately.

We want to discover recurring problem patterns, but we must not invent prevalence claims.

---

# 5. Discover real operational problems

This is the most important part of the research.

Investigate actual, documented operational problems encountered when HR, WFM and payroll applications execute business processes and exchange data.

Organize findings into the following categories.

## A. HRIS problems

Examples to investigate, not assume:

- Duplicate employee identities.
- Incorrect employee status.
- Employment effective-date errors.
- Incorrect salary history.
- Incorrect employee classification.
- Missing employment changes.
- Inconsistent organizational records.
- Incorrect leave eligibility.
- Failed employee synchronization.
- Terminated employee remaining active downstream.

## B. WFM problems

Investigate:

- Missing attendance events.
- Duplicate clock punches.
- Incorrect shift assignment.
- Cross-midnight time calculation errors.
- Incorrect break deductions.
- Incorrect rest-day classification.
- Public holiday calendar mismatch.
- Incorrect overtime approval status.
- Incorrect overtime aggregation.
- Approved attendance corrections not propagated.
- Clock-device synchronization problems.
- Timezone and date-boundary errors.

## C. Payroll problems

Investigate:

- Incorrect basic salary.
- Incorrect partial-month salary.
- Incorrect unpaid leave deductions.
- Incorrect overtime calculations.
- Incorrect employee eligibility.
- Incorrect CPF contribution.
- Incorrect OW/AW classification.
- Outdated contribution rate tables.
- Incorrect backpay treatment.
- Duplicate payroll transactions.
- Payroll run discrepancies.
- Missing payslip components.
- Incorrect deductions.
- Incorrect payroll cut-off treatment.
- Failed statutory submissions.
- Incorrect tax-reporting data.

## D. Integration problems

Investigate:

- Missing employees downstream.
- Missing attendance records.
- Missing approved timesheets.
- Duplicate events.
- Duplicate payroll exports.
- Partial synchronization.
- Invalid field mappings.
- Schema changes.
- Token expiry.
- Authentication failure.
- Webhook delivery problems.
- Incorrect retry behavior.
- Dead-letter queue accumulation.
- Out-of-order events.
- Stale data.
- Failed reconciliation.
- Incorrect effective-date propagation.

## E. Production and database problems

Investigate:

- Data integrity violations.
- Referential integrity problems.
- Incorrect transactions.
- Race conditions.
- Duplicate processing.
- Long-running jobs.
- Database locks.
- Connection pool exhaustion.
- Slow queries.
- Failed migrations.
- Incorrect indexes.
- Background job failures.
- Permissions problems.
- Incorrect tenant isolation.
- Inconsistent reads.
- Cache invalidation.

Do not assume every issue is a statutory compliance defect.

A technically valid application can still process a legally incorrect calculation. Conversely, a technically failed integration can leave perfectly valid source data but incomplete downstream records.

Keep those distinctions clear.

---

# 6. Identify the most valuable Tier 2 investigation scenarios

We are building an OJT environment for aspiring support engineers.

The goal is to expose them to the problems that require actual technical investigation rather than memorization of definitions.

For each operational problem, identify:

1. What the customer or end user experiences.
2. What business process should have happened.
3. What actually happened.
4. What data should exist at each stage.
5. What the engineer needs to inspect.
6. Which SQL queries or database relationships are relevant.
7. Which API calls or payloads should be inspected.
8. Which logs, correlation IDs or job records are relevant.
9. Which competing root causes are plausible.
10. How the engineer establishes the actual root cause.
11. Whether the issue requires a configuration correction, data correction, integration retry, code fix or escalation.
12. How to verify the result after resolution.

Do not write trivial scenarios such as "the API returns HTTP 500."

Explain the actual business consequences.

For example:

An attendance system records an employee's approved working hours. The synchronization worker exports those hours to payroll. A mapping filter excludes one approved overtime pay code. Payroll completes successfully but calculates a lower amount.

A meaningful Tier 2 investigation must establish:

- The original attendance records were correct.
- The approved timesheet was correct.
- The outbound payload contained the relevant records.
- The integration configuration excluded the pay code.
- Payroll never received the omitted component.
- The payslip therefore differs from the independently calculated expected result.

The scenario should contain sufficient misleading symptoms and alternative hypotheses to require real investigation.

Produce an evidence-backed incident catalogue, not merely a list of possible bugs.

---

# 7. Reconstruct the complete business data lifecycle

The research must connect statutory rules and SaaS operational behavior through actual data flows.

For example:

```
Employee onboarding
       |
Employment contract
       |
Salary and employee classification
       |
Shift and roster assignment
       |
Clock-in / clock-out
       |
Attendance interpretation
       |
Leave and overtime approval
       |
Timesheet
       |
Integration export
       |
Payroll import and mapping
       |
Payroll calculation
       |
CPF / SDL assessment
       |
Payslip generation
       |
Reporting and reconciliation
```

For each transition, identify:

- Source system.
- Target system.
- Business event.
- Data fields.
- Validation conditions.
- Effective dates.
- Processing timing.
- Approval requirements.
- Expected database changes.
- Possible failure modes.
- Required audit evidence.

The goal is to understand where incorrect business results are introduced, not just identify the final incorrect payslip.

Create an end-to-end data lineage model.

---

# 8. Verify statutory rules and computations mathematically

Do not merely extract formulas from webpages.

Reconstruct the relevant calculations and independently verify them.

For every implemented calculation, supply:

### A. Source evidence

- Authoritative source.
- Exact URL.
- Legislation or document version.
- Section, table or paragraph.
- Publication and effective dates.

### B. Applicability

- Employee categories.
- Conditions.
- Exclusions.
- Salary thresholds.
- Citizenship or residency distinctions.
- Employment arrangements.
- Exceptions.

### C. Mathematical definition

Define all input variables, units, rounding conventions and formulas.

Distinguish basic salary, gross salary, ordinary wages, additional wages, allowances, reimbursements and other relevant payment classifications.

### D. Worked examples

Provide independently calculated examples:

- Normal case.
- Lower boundary.
- Upper boundary.
- Just below threshold.
- Exactly at threshold.
- Just above threshold.
- Mid-period employment change.
- Retroactive adjustment.
- Multiple interacting conditions.
- Unsupported or ambiguous case.

### E. Test oracle

Provide exact expected results and derivation.

Where possible, use official worked examples or independently verified computations.

Do not manufacture a supposedly official example.

### F. Failure consequences

Explain what goes wrong if the calculation or applicability condition is implemented incorrectly.

Identify the downstream employee, employer, payroll or reporting impact.

---

# 9. Distinguish statutory correctness from software correctness

This distinction must govern the entire research.

We have four different sources of expected behavior:

1. Singapore statutory requirements.
2. Individual employment contracts.
3. Employer policies and configuration.
4. Actual application and integration behavior.

For every finding, explicitly identify which category it belongs to.

For example:

An employee receives insufficient overtime payment.

Possible explanations:

- Incorrect statutory eligibility determination.
- Incorrect employment contract data.
- Incorrect employer overtime configuration.
- Incorrect shift calculation.
- Incorrect attendance record.
- Incorrect pay-code mapping.
- Incorrect payroll calculation.
- Incorrect downstream synchronization.

Do not collapse these into one generic category called "payroll bug."

Our platform must be able to distinguish the legal/business rule problem from the software/system implementation problem.

---

# 10. Produce an evidence-ranked operational problem catalogue

For each identified problem, provide:

- Problem ID.
- Business domain.
- Process stage.
- Customer-visible symptom.
- Actual documented source.
- Evidence classification.
- Likely root causes.
- Regulatory or contractual implications.
- Affected data.
- Diagnostic approach.
- Expected resolution.
- Verification procedure.
- Training value.
- Implementation priority.

Use these evidence levels:

**E1 — Primary statutory evidence**

Actual legislation, official regulations or authoritative government documentation.

**E2 — Official operational documentation**

Government procedures, technical specifications, employer guidance and official submission documentation.

**E3 — Vendor documentation**

Actual SaaS help articles, APIs, integration instructions, product limitations and troubleshooting guides.

**E4 — Public operational evidence**

Credible support-engineering job descriptions, published incident reports, customer forum discussions, public GitHub issues or documented user problems.

**E5 — Engineering inference**

A technically plausible failure inferred from real architectures and business processes but not directly documented as an observed incident.

Never present E5 as an established real-world incident.

For repeated problem patterns, explain how many independent sources support the finding.

Do not claim frequency or prevalence without representative evidence.

---

# 11. Design the WorkforceOps Lab reference module from the findings

The research must translate into specific implementation recommendations.

Our existing target architecture includes:

- Go backend.
- Module kernel.
- `sg-wfm-payroll` reference module.
- ShiftHub.
- PayLoop.
- PostgreSQL.
- Independent calculation oracle.
- Integration worker.
- Incident and fault simulation.
- CaseTrace investigation.
- Support Engineering Lab web interface.
- CLI and REST API.
- Local-first deployment.
- Optional Neon and Fly.io.

Do not replace this architecture without compelling evidence.

Your role is to establish what this system must realistically model.

## Required module outputs

Produce:

### A. Domain model

Identify actual entities, relationships, attributes and effective dates.

### B. Rule register

Provide all verified and candidate rules, with exact source references and applicable versions.

### C. Data lifecycle

Define how information moves from HR through WFM, payroll and reporting.

### D. Integration contracts

Identify relevant APIs, payload fields, mapping requirements, events, cut-offs and failure paths.

### E. Calculation specification

Provide formulas and golden test cases.

### F. Incident catalogue

Create realistic, reproducible operational failures.

### G. Investigation specification

Define diagnostic steps, evidence requirements and escalation criteria.

### H. Acceptance tests

Translate statutory and operational findings into testable software requirements.

Do not assume everything researched must be implemented in the first release.

Separate required MVP functionality from later roadmap items.

---

# 12. Construct the initial Tier 2 incident library

Build an evidence-backed catalogue of at least 30 candidate scenarios across HRIS, WFM, payroll and integration.

From that larger catalogue, select the 12 highest-value scenarios for initial development planning.

For each selected scenario, provide a complete scenario specification.

Use this structure:

```
scenario_id: SG-PAY-001

title: Approved overtime missing from payroll

domain: payroll-integration

evidence_class: E3

business_context:
  description: >
    A realistic description of the employer's
    workflow and the affected employee.

customer_ticket:
  description: >
    The actual complaint that a support engineer
    would receive.

preconditions:
  - employee_exists
  - contract_valid
  - attendance_approved

business_rules:
  - rule_id: SG-MOM-OT-001

expected_business_process:
  - attendance_recorded
  - timesheet_approved
  - export_created
  - payroll_imported

injected_fault:
  mechanism: paycode_mapping_filter

observable_symptoms:
  - overtime_missing_from_payroll
  - source_timesheet_correct

candidate_hypotheses:
  - mapping_error
  - payroll_eligibility_error
  - missing_attendance
  - failed_integration

minimum_evidence:
  - approved_timesheet
  - outbound_payload
  - mapping_configuration
  - payroll_import_record

ground_truth:
  visibility: evaluator_only

expected_diagnosis: evaluator_only

expected_resolution:
  visibility: evaluator_only

verification:
  expected_payroll_result: evaluator_only
```

The actual scenario format may be adapted.

The important condition is that the investigator cannot see the ground truth or expected diagnosis.

## Scenario difficulty

Classify scenarios as:

- Beginner: isolated error and relatively direct evidence.
- Intermediate: multiple related records or systems.
- Advanced: competing explanations and effective-date interactions.
- Expert: multiple interacting failures, misleading symptoms and complex reconciliation.

Prioritize scenarios that genuinely require SQL, logs, API inspection and business-rule knowledge.

---

# 13. Identify the technical competencies the job market actually expects

Research job advertisements for Tier 2 Application Support, Technical Support, Production Support and Enterprise SaaS Support positions.

Prioritize Singapore-based roles, but also examine relevant regional B2B SaaS and enterprise support roles.

Extract actual technical requirements.

Investigate:

- SQL.
- PostgreSQL and relational databases.
- REST APIs.
- JSON and XML.
- Authentication and authorization.
- Logs and observability.
- Linux.
- HTTP.
- Network troubleshooting.
- Integration debugging.
- Data reconciliation.
- Batch processing.
- Background jobs.
- Incident management.
- Jira Service Management or equivalent.
- ITIL-related incident and problem management.
- Customer communication.
- Technical documentation.
- Escalation to engineering.

Distinguish:

- Frequently repeated responsibilities.
- Explicitly required skills.
- Preferred skills.
- Skills inferred from the job description.

Do not assume that a skill is universal merely because it appears in a few advertisements.

Produce a competency matrix mapping each skill to concrete WorkforceOps Lab exercises.

For example:

| Competency                   | Training environment         | Evidence of competence               |
| ---------------------------- | ---------------------------- | ------------------------------------ |
| SQL investigation            | Payroll inconsistency        | Correct query and evidence           |
| API debugging                | Failed HRIS sync             | Identify payload or response problem |
| Log analysis                 | Dead-letter queue            | Trace processing failure             |
| Business-rule interpretation | Overtime discrepancy         | Correct eligibility and calculation  |
| Reconciliation               | Attendance vs payroll        | Identify divergence                  |
| Communication                | Customer incident            | Clear technical explanation          |
| Escalation                   | Confirmed application defect | Reproducible engineering report      |

---

# 14. Identify which problems are best for training and assessment

Not every real-world problem makes a good learning exercise.

Rank each scenario according to:

- Business realism.
- Evidence strength.
- Regulatory significance.
- SQL investigation value.
- API and integration investigation value.
- Log investigation value.
- Business-rule complexity.
- Root-cause uniqueness or ambiguity.
- Ability to reproduce deterministically.
- Educational value.
- Relevance to actual Tier 2 job requirements.

Prefer scenarios that teach transferable investigation skills.

Avoid scenarios whose answer depends on obscure legal trivia without meaningful system investigation.

A strong scenario should require the learner to understand both the business process and the software behavior.

---

# 15. Adversarial validation

Before recommending any business rule or incident for implementation, subject it to adversarial review.

Look for:

- Incorrect interpretations of Singapore statutory rules.
- Outdated rates or effective dates.
- Conflicting official guidance.
- Misclassified employee categories.
- Incorrect wage definitions.
- Wrong payroll cut-off assumptions.
- Rounding mistakes.
- Cross-midnight errors.
- Incorrect overtime eligibility.
- Double-counted or omitted wages.
- Invalid assumptions about public holidays or rest days.
- Software behavior presented as law.
- Employer policy presented as law.
- Plausible incidents presented as documented actual incidents.
- Scenarios that disclose their answer through visible symptoms.
- Tests that simply reproduce the same implementation's mistakes.

For mathematical calculations, independently reconstruct the expected result.

For critical statutory rules, seek more than one authoritative source where available, and identify the governing source when the wording differs.

Use the stop-or-carry discipline:

- Verified: supported and independently checked.
- Conditional: valid only under stated assumptions.
- Unverified: insufficient evidence.
- Blocked: contradictory or missing load-bearing evidence.
- Rejected: disproven by source or derivation.

Do not turn uncertainty into a guessed implementation requirement.

---

# 16. Research priorities and implementation boundaries

The research must be comprehensive, but implementation recommendations must remain practical.

Organize findings into:

**P0 — MVP rule and workflow foundation**

The minimum verified rules and workflows required to make the initial environment function credibly.

**P1 — High-value Tier 2 scenarios**

Additional payroll, HR, WFM and integration cases with strong training value.

**P2 — Advanced enterprise cases**

Complex integrations, advanced payroll exceptions, reporting, access control and operational incidents.

**P3 — Future business modules**

Other domains and enterprise features such as SSO, identity management and advanced business integrations.

Do not let completeness of statutory research turn the MVP into a complete Singapore payroll software implementation.

However, do not oversimplify or misrepresent any rule that the MVP actually implements.

Every implemented calculation must be correct within its declared applicability and validity boundary.

---

# 17. Required deliverables

Produce a complete research dossier with the following documents.

```
research/singapore-hr-wfm-payroll/
│
├── 00_EXECUTIVE_SUMMARY.md
│
├── 01_RESEARCH_METHODOLOGY.md
│
├── 02_OFFICIAL_SOURCE_REGISTER.md
│
├── 03_SINGAPORE_EMPLOYMENT_RULES.md
│
├── 04_HR_SYSTEM_REQUIREMENTS.md
│
├── 05_WFM_SYSTEM_REQUIREMENTS.md
│
├── 06_PAYROLL_COMPUTATION_SPECIFICATION.md
│
├── 07_CPF_SDL_IRAS_REQUIREMENTS.md
│
├── 08_SAAS_VENDOR_IMPLEMENTATION_RESEARCH.md
│
├── 09_OPERATIONAL_PROBLEM_CATALOGUE.md
│
├── 10_END_TO_END_DATA_LINEAGE.md
│
├── 11_TIER2_COMPETENCY_MATRIX.md
│
├── 12_INCIDENT_SCENARIO_CATALOGUE.md
│
├── 13_MVP_REFERENCE_MODULE_REQUIREMENTS.md
│
├── 14_REGULATORY_GOLDEN_TESTS.md
│
├── 15_ADVERSARIAL_REVIEW.md
│
├── 16_UNKNOWNS_AND_UNVERIFIED_AREAS.md
│
└── 17_IMPLEMENTATION_PRIORITY_AND_ROADMAP.md
```

Also provide structured machine-readable artifacts:

- `rules.json` or an equivalent versioned rule format.
- `rule_test_cases.json`.
- `incident_catalogue.yaml`.
- `competency_matrix.json`.
- `source_manifest.json`.

The structured files should be usable by future coding agents and test generators.

---

# 18. Citation and source-integrity requirements

Every material statutory, calculation or vendor-behavior claim must have an exact source reference.

For every primary source, record:

- Full document title.
- Issuing agency or vendor.
- Official URL.
- Publication or update date, where available.
- Effective date, where relevant.
- Exact section, paragraph, table or API endpoint.
- Retrieval date.
- Version or document identifier.
- Whether the document was actually opened and reviewed.

Do not cite search-result snippets as though they were complete documents.

Do not cite a vendor's product marketing page as evidence of actual implementation behavior.

Do not fabricate vendor defects or support cases.

If a relevant source is inaccessible, record the gap rather than reconstructing its contents from other sources.

Use original government publications and statutes wherever possible.

Secondary articles may help discover topics but should not establish statutory requirements.

---

# 19. Required final synthesis

Conclude with direct answers to these questions:

1. What are the most important real Singapore HR, WFM and payroll business rules that WorkforceOps Lab must implement correctly?
2. Which business workflows connect those rules?
3. Which operational problems are documented by government, vendors or credible public evidence?
4. Which common failure patterns are well-supported, and which are engineering inferences?
5. What are the most valuable Tier 2 troubleshooting skills demonstrated by those problems?
6. Which 12 incidents should be implemented first, and why?
7. Which calculations require independent golden tests?
8. Which areas of the current WorkforceOps Lab architecture need changes?
9. What must be implemented for the MVP?
10. What should be deferred?
11. Which assumptions remain unverified?
12. What would a senior Singapore payroll practitioner or enterprise support engineer most likely challenge in our implementation?

Finish with a prioritized implementation specification that can be handed directly to the WorkforceOps Lab coding agent.

---

# 20. Non-negotiable research principles

1. Research actual Singapore requirements, not generic international HR practices.
2. Prefer primary legislation and official government documentation.
3. Examine actual SaaS product documentation and integration behavior.
4. Distinguish legal requirements from employer configuration and software design.
5. Do not invent real-world incidents.
6. Label inferences and hypothetical failures honestly.
7. Do not simplify statutory calculations for convenience.
8. Identify exact applicability conditions and effective dates.
9. Independently derive calculation examples.
10. Treat unknowns and contradictory evidence explicitly.
11. Produce implementation-ready test cases rather than a generic narrative.
12. Keep the original WorkforceOps Lab product scope and architecture in view.
13. Do not recommend building every researched capability in the MVP.
14. Prioritize realistic Tier 2 investigation experience over legal trivia.
15. Evaluate findings from the perspective of an engineer who must diagnose a real customer's business problem.

**The ultimate objective is to build a Singapore-grounded enterprise simulation whose business rules and computations withstand scrutiny, whose failures resemble genuine operational problems, and whose investigation exercises prepare people for actual Tier 2 support work.**

Do not give us a simplified mock enterprise.

Give us an evidence-backed specification for a technically credible enterprise system and its operational failure modes.

## What I particularly want the research agent to discover

I would pay special attention to three areas because they will determine whether this becomes a credible OJT platform:

1\. Effective dates and historical data. Employment changes, salary revisions, CPF changes, payroll cut-offs and attendance corrections interact across time. A system can have individually correct calculations and still produce incorrect payroll because it used the wrong version of a record or rule.

2\. Cross-system data integrity. The most valuable support cases will often involve a correct HR record, a correct attendance record and an incorrect payroll result caused by something in between. This is where SQL, APIs, logs, reconciliation and business knowledge meet.

3\. Realistic troubleshooting evidence. Every scenario must expose enough evidence to investigate, but not make the answer obvious. The learner should have to determine whether the failure originated in HR, WFM, payroll, configuration, integration or the underlying statutory interpretation.

The desired outcome is not just a collection of Singapore regulations. It is a source-verified business-rule catalogue, a realistic operational failure taxonomy, a Tier 2 competency framework and an implementation-ready incident library that the coding agent can turn into WorkforceOps Lab.