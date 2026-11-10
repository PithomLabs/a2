# I think this is an excellent starting point. I would build it.

In fact, I would make one important change to the original CaseTrace idea: instead of beginning with an AI support investigation tool and trying to find companies willing to connect their production systems, I would first build a working miniature of a real Singapore workforce-management and payroll platform.

Then I would deliberately introduce the kinds of problems that occur in actual enterprise software.

On top of that, I would build the agent-driven Tier 2 investigation system.

The result is something much more interesting than a collection of mock tickets or coding exercises:

A functioning business application, populated with realistic workforce data, governed by Singapore business rules, connected through APIs, and capable of producing realistic production incidents for AI agents and human support engineers to investigate.

I would call the initial project WorkforceOps Lab, with CaseTrace as its investigation engine.

# WorkforceOps Lab

## A simulated enterprise workforce platform for training and testing Tier 2 support engineers

Go · Neon PostgreSQL · Fly.io · Agent-driven investigations

HR operations

Employees, contracts, eligibility, leave, employment history

Workforce management

Rosters, shifts, time clocks, attendance, overtime

Payroll

Salary, allowances, deductions, CPF, SDL, payslips

Integrations

Payroll exports, APIs, synchronization, retries, failures

Production simulation

Realistic data, application events, defects, configuration problems

AI investigation

Hypotheses, SQL, logs, evidence, root-cause verification

## 1. Why the Singapore government sources are a particularly useful foundation

Singapore has something valuable for this project: government agencies publish fairly detailed operational guidance, not just general statements about employee rights.

We can turn those documents into executable business rules.

For example:

| Government source                                                                                                                                      | What it gives our simulator                                               |
| ------------------------------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------- |
| [MOM — Hours of work, overtime and rest days](https://www.mom.gov.sg/employment-practices/hours-of-work-overtime-and-rest-days?utm_source=chatgpt.com) | Working-hour limits, overtime eligibility and calculation, rest-day rules |
| [MOM — Annual leave](https://www.mom.gov.sg/employment-practices/leave/annual-leave/eligibility-and-entitlement?utm_source=chatgpt.com)                | Eligibility, service-based entitlement, pro-rating                        |
| [MOM — Sick leave](https://www.mom.gov.sg/employment-practices/leave/sick-leave/eligibility-and-entitlement?utm_source=chatgpt.com)                    | Eligibility, medical certification and leave entitlement                  |
| [MOM — Public holidays](https://www.mom.gov.sg/employment-practices/public-holidays-entitlement-and-pay?utm_source=chatgpt.com)                        | Holiday entitlement and payment scenarios                                 |
| [MOM — Key employment terms](https://www.mom.gov.sg/employment-practices/contract-of-service/key-employment-terms?utm_source=chatgpt.com)              | Contract and employee configuration requirements                          |
| [MOM — Itemised payslips](https://www.mom.gov.sg/employment-practices/salary/itemised-payslips?utm_source=chatgpt.com)                                 | Payroll output requirements                                               |
| [MOM — Salary payment](https://www.mom.gov.sg/employment-practices/salary/paying-salary?_sm_au_=iVVkZVrFV0WDZ0Qr\&utm_source=chatgpt.com)              | Salary and overtime payment deadlines                                     |
| [CPF Board — Contribution rates](https://www.cpf.gov.sg/employer/employer-obligations/how-much-cpf-contributions-to-pay?utm_source=chatgpt.com)        | CPF calculations and employee classifications                             |
| [CPF Board — Skills Development Levy](https://www.cpf.gov.sg/employer/employer-obligations/skills-development-levy?utm_source=chatgpt.com)             | SDL calculations                                                          |

These are useful because they give us rules that can be translated into both application functionality and testable expected outcomes.

### A few examples of rules we can encode

Overtime

MOM specifies that Part IV of the Employment Act applies to particular categories of employees. Generally, the relevant basic salary thresholds are $4,500 for workmen and $2,600 for non-workmen, with managers and executives excluded from Part IV. Eligible overtime is generally paid at no less than 1.5 times the hourly basic rate.&#x20;

[image](https://www.google.com/s2/favicons?domain=https://www.mom.gov.sg\&sz=32)

mom.gov.sg+1



Annual leave

An eligible employee who has completed three months of service starts with seven days of annual leave in the first year, increasing by one day per year until reaching fourteen days. The government also specifies how leave should be pro-rated.&#x20;

[image](https://www.google.com/s2/favicons?domain=https://www.mom.gov.sg\&sz=32)

mom.gov.sg



CPF

CPF calculations depend on employee age, citizenship or permanent-resident status, wages and other applicable conditions. The Ordinary Wage ceiling is $8,000 from January 2026. There is also a published increase in CPF rates for workers aged above 55 to 65 effective January 1, 2027.&#x20;

[image](https://www.google.com/s2/favicons?domain=https://www.cpf.gov.sg\&sz=32)

cpf.gov.sg+1



Payroll payslips

We can enforce requirements such as showing basic salary, allowances, deductions, overtime hours and pay, and net salary.&#x20;

[image](https://www.google.com/s2/favicons?domain=https://www.mom.gov.sg\&sz=32)

mom.gov.sg



Notice something important: these aren't merely examples for a training quiz. They are rules the application itself will execute.

There is one crucial qualification: government guidance doesn't specify every employer's working arrangements, payroll configuration, or integration behavior. Employment contracts and company policies supply additional rules. Our system must model those separately, rather than incorrectly presenting every business rule as a statutory requirement.

## 2. The most important architectural decision: separate four kinds of business rules

I would not put all the rules into one large payroll calculation function.

That would create an artificial application that is difficult to troubleshoot.

Instead, we need four distinct rule categories.

Layer 1

Government statutory rules

Employment Act requirements, leave entitlements, overtime, CPF, SDL and public holidays.

Every rule needs an official source, effective date, applicability conditions and test cases.

Layer 2

Employment contract rules

Employee working hours, salary, rest days, overtime arrangements, salary period and contractual entitlements.

Two employees working for the same company may have different applicable terms.

Layer 3

Employer configuration

Rostering policies, approval workflows, pay codes, payroll calendars, department assignments, shift premiums and integration mappings.

Layer 4

Application behavior and defects

Incorrect database records, broken calculations, stale data, missing events, configuration mistakes, integration errors and software bugs.

Why is this separation important?

Imagine an employee receives insufficient overtime pay.

There are at least four possible explanations:

- The statutory eligibility rules were applied incorrectly.
- The employee's employment contract contains different working arrangements.
- The employer's overtime policy or pay-code configuration is wrong.
- The application has a defect that loses overtime records during payroll processing.

A Tier 2 engineer must distinguish these possibilities. Our platform must make that investigation possible.

The government rulebook becomes the expected-behavior reference. The application becomes the system under investigation.

### Make the rulebook versioned

This is especially important in Singapore because rules change.

For example, the 2027 CPF rate changes have already been published. We should not overwrite the 2026 rates with the 2027 rates. Both must be available for historical payroll calculations.&#x20;

[image](https://www.google.com/s2/favicons?domain=https://www.cpf.gov.sg\&sz=32)

cpf.gov.sg



Every rule should have something like:

```
{
  "rule_id": "SG-CPF-SENIOR-2027",
  "jurisdiction": "SG",
  "category": "CPF",
  "effective_from": "2027-01-01",
  "effective_until": null,
  "source": "CPF Board",
  "source_url": "official-source-url",
  "source_version": "published-2026",
  "status": "verified",
  "tests": [
    "cpf_senior_rate_before_effective_date",
    "cpf_senior_rate_after_effective_date"
  ]
}
```

This is an illustrative representation, not a complete CPF rule definition.

I would maintain a source register containing the original URL, retrieved document, hash, relevant section, interpretation, effective date and independently verified test cases.

An LLM may help us interpret the documentation, but it must never silently promote its interpretation into a production rule. A human-approved, deterministic implementation is the authority inside the simulator.

## 3. Build a real HR/WFM/payroll application, not a collection of mock screens

This is probably the most important part of the entire idea.

We need a system that genuinely processes workforce operations.

I would start with one fictional Singapore company:

Meridian Workforce Services Pte. Ltd.

Illustrative company — entirely fictional.

Let's give it:

- 120 employees.
- Three operating departments.
- Two physical work locations.
- Different employee categories.
- Singapore citizens, permanent residents and foreign employees.
- Monthly and hourly-paid employees.
- Managers, supervisors and workmen.
- Different employment contracts.
- Five-day, six-day and rotating shift arrangements.

These are design choices for our simulator, not claims about a typical Singapore business.

The application should allow the company to perform the following actual operations.

## Meridian Workforce — operational flow

1. Hire an employee

   Create employment record, contract, salary, category and statutory eligibility.
2. Create a work schedule

   Assign shift, department, work location, rest day and roster.
3. Record attendance

   Employees clock in and out. The system creates time records and exceptions.
4. Process leave and overtime

   Submit, approve, reject and calculate leave and additional working hours.
5. Run payroll

   Calculate basic salary, overtime, allowances, deductions, CPF and SDL.
6. Generate outputs

   Produce payslips, payroll reports and integration/export records.

The important part:

Every stage creates real records that subsequent stages consume. A defective attendance record must be capable of producing an incorrect payroll result.

This is how we get realistic troubleshooting scenarios: by letting one error propagate through the actual application.

### What I would implement first

| Module           | Initial functionality                               |
| ---------------- | --------------------------------------------------- |
| HR               | Employees, employment contracts, categories, salary |
| WFM              | Shifts, rosters, clock-in/out, attendance           |
| Leave            | Annual leave, sick leave, unpaid leave              |
| Time calculation | Normal hours, overtime, rest days, public holidays  |
| Payroll          | Basic salary, overtime, allowances, deductions      |
| Contributions    | CPF and SDL                                         |
| Outputs          | Payslips, payroll reports, export files             |
| Integrations     | A simulated external payroll/accounting API         |

I would initially exclude complicated bonus calculations, income-tax filing, full work-pass administration, every category of parental leave, and real bank or CPF submissions.

Those can come later. We should not turn the MVP into a complete commercial payroll product.

## 4. How we generate realistic data without production databases

This is where I would be quite strict.

I would not simply generate 10,000 random employees and call that a production simulation.

Random records are not automatically realistic.

We need a synthetic dataset that obeys business relationships and generates believable operational history.

The Singapore PDPC's anonymisation guidance specifically recognizes fictitious data for application development and testing, and also discusses generating synthetic data from scratch. We should use the latter approach rather than try to obtain and modify real employee records.&#x20;

[image](https://www.google.com/s2/favicons?domain=https://www.pdpc.gov.sg\&sz=32)

pdpc.gov.sg



### Build the data generator around business reality

For example, when we create an employee:

1. Assign an employee category.
2. Assign employment terms.
3. Assign an appropriate salary structure.
4. Determine applicable statutory eligibility.
5. Assign a department and supervisor.
6. Generate work schedules.
7. Generate attendance records.
8. Introduce approved leave and absences.
9. Produce payroll transactions.

The generated data must be internally consistent.

An employee should not receive an overtime payment for hours that were never worked, unless we have deliberately introduced a defect.

Similarly, a payroll deduction should have an originating reason, and an employee's CPF calculation should be traceable to the applicable wage, age and citizenship/status attributes.

Use generated employee IDs such as `EMP-SG-000142`. Do not generate real NRICs, FINs, bank accounts or other identity credentials.

### I would generate data at three levels

Level A — Master data

Employees, departments, locations, job roles, employment contracts, salary structures and employer policies.

Level B — Operational history

Six to twelve months of rosters, punches, approved leave, attendance corrections, payroll periods and payroll transactions.

Level C — System activity

API requests, webhooks, integration attempts, error logs, approval events, database audit records and processing history.

The third level is extremely important.

A Tier 2 engineer rarely investigates an isolated database row. They investigate a history of events across multiple systems.

Our simulator should therefore preserve the full journey of a transaction.

## 5. The real differentiator: create realistic incidents

Here is where I would spend a considerable part of the development effort.

We should build an Incident Factory.

It creates realistic failures inside our functioning application, allowing an engineer or agent to investigate the same symptoms that occur in real enterprise systems.

I would begin with these scenarios.

### Initial incident catalogue

| Incident | What happens                                        | What Tier 2 must investigate                                 |
| -------- | --------------------------------------------------- | ------------------------------------------------------------ |
| INC-001  | Overtime missing from payslip                       | Time records, eligibility and payroll mapping                |
| INC-002  | Employee receives incorrect CPF                     | Age/status, applicable rates, wage ceiling and rounding      |
| INC-003  | Leave balance is incorrect                          | Service period, completed months and leave transactions      |
| INC-004  | Public holiday payment is missing                   | Holiday calendar, employee schedule and payment rule         |
| INC-005  | Employee is missing from payroll                    | Employment status, effective dates and payroll inclusion     |
| INC-006  | Duplicate payroll export                            | API idempotency and retry behavior                           |
| INC-007  | Attendance differs from approved roster             | Punch records, shift assignment and approvals                |
| INC-008  | Payroll export contains incorrect salary            | Mapping, stale records or calculation defect                 |
| INC-009  | Employee's overtime is misclassified                | Part IV coverage, contract terms and employee classification |
| INC-010  | Payroll calculation differs across runs             | Rule version, rounding and processing state                  |
| INC-011  | Approved leave did not update payroll               | Workflow events and synchronization                          |
| INC-012  | Payroll API reports success but records are missing | Request/response logs and downstream reconciliation          |

The first version should have at least ten fully reproducible incidents, not fifty shallow ones.

### Let me show you one realistic investigation

INC-001 — Missing overtime pay

High priority

Fictional customer incident

Customer complaint

"Our employee worked overtime during October, but the overtime payment is missing from the November payslip."

What the application actually contains

- Employee: EMP-SG-000142
- Employee category: Non-workman
- Monthly basic salary: $2,500
- Approved overtime: 6 hours
- Attendance: Confirmed
- Timesheet: Approved
- Payroll: Completed
- Payslip: Generated

Deliberately seeded defect

The timesheet produces a pay code called `EXTRA_HOURS`, but the payroll configuration only recognizes `OVERTIME`.

The time calculation is correct. The payroll mapping drops the payment.

What the agent must do

1. Establish that the employee is eligible for the relevant overtime provisions.
2. Verify the approved time records.
3. Inspect the calculated overtime transaction.
4. Trace the pay code into payroll.
5. Identify the incorrect mapping.
6. Establish the expected payment independently.
7. Provide evidence and recommend a correction.

The scenario is constructed from the relevant MOM coverage and overtime rules, but the mapping defect itself is an invented simulation condition.

This one incident tests SQL, data tracing, business rules, payroll, integration between modules and communication.

That is much closer to a real Tier 2 support assignment than asking someone to solve a standalone SQL exercise.

### Make the failures realistic, not theatrical

I would introduce controlled fault types:

- Incorrect configuration.
- Incorrect master data.
- Missing database records.
- Duplicate records.
- Inconsistent records between modules.
- Failed integration messages.
- Duplicate or out-of-order webhooks.
- Incorrect application calculation.
- Stale cached or derived data.
- Permission and tenant-isolation failures.
- Intermittent API timeouts and retry problems.

The Incident Factory should create the failure through the same application interfaces and processing paths that normal operations use.

Do not have the agent simply read a hidden scenario description that says what went wrong.

The scenario's ground truth should be stored separately and hidden from the investigator.

That is how we can genuinely test whether the agent discovers the cause.

## 6. The AI architecture: agents should investigate the same systems as human engineers

I would reuse the research harness architecture we have been developing, but simplify it for operational support.

## Agent investigation pipeline

Customer ticket received

"Overtime missing from payslip."

Agent A — Triage

Identifies affected employee, payroll period, business impact and required evidence.

Agent B — Hypothesis generation

Proposes possible causes and ranks them for investigation.

Agent C — Evidence investigator

Uses approved SQL queries, APIs, logs and application history to test hypotheses.

Agent D — Independent verifier

Receives the findings, evidence, rules and tests. Checks whether the conclusions actually follow.

Investigation report

Root cause (if established), evidence, rejected hypotheses, uncertainty, recommended resolution and escalation details.

The important part is not how many agents we have. It is the division of responsibility and evidence.

The hypothesis generator should not grade itself. The verifier should be given the relevant evidence and definitions, not simply asked to agree with another agent's conclusion.

For the MVP, the verifier should also have deterministic checks. For example, if the agent claims that six approved overtime hours were lost in a pay-code mapping, the verifier can compare:

- Approved overtime source rows.
- Payroll calculation rows.
- Pay-code mapping configuration.
- Generated payslip line items.

This is much more reliable than relying on two language models agreeing with each other.

### Give the agent controlled tools

Initially, I would expose these tools:

```
get_employee(employee_id)
get_employment_terms(employee_id)
get_roster(employee_id, date_range)
get_attendance(employee_id, date_range)
get_timesheet(employee_id, pay_period)
get_leave_transactions(employee_id)
get_payroll_run(pay_period)
get_payslip(employee_id, pay_period)
get_payroll_export(export_id)
get_integration_logs(correlation_id)
execute_approved_query(query_id, parameters)
```

These are examples of application tool contracts.

I would not give an LLM arbitrary database credentials. The Go service should enforce permissions, tenant boundaries, query limits and audit logging.

The human support engineer should see the exact same evidence that the agent sees.

## 7. How I would implement this in Go, Neon and Fly.io

I would keep it as a modular Go monolith initially.

There is no need to make every module a separate microservice.

## Proposed technical architecture

Go web application

Server-rendered HTML · HTMX · JSON APIs

Go domain modules

HR / Contracts

WFM / Attendance

Leave / Overtime

Payroll / CPF / SDL

Integration Gateway

Incident Factory

Agent Orchestrator

Audit / Reporting

Neon PostgreSQL

Business tables · Operational events · Job state · Evidence ledger · Rule versions · Incident ground truth

Fly.io

Web process · Background worker · Scheduled processing

### My preferred Go stack

| Area            | Technology                                              |
| --------------- | ------------------------------------------------------- |
| Language        | Go                                                      |
| Web routing     | Chi or standard `net/http`                              |
| Database        | Neon PostgreSQL                                         |
| SQL             | pgx + sqlc                                              |
| Migrations      | Goose                                                   |
| UI              | Go templates + HTMX                                     |
| Background work | Durable PostgreSQL-backed jobs                          |
| Agent execution | Go interfaces and explicit workflow state machine       |
| Validation      | Standard Go tests, integration tests and scenario tests |
| Observability   | OpenTelemetry + Better Stack                            |
| Hosting         | Fly.io                                                  |

Neon gives us PostgreSQL compatibility without initially having to operate our own database server. Fly.io gives us a practical deployment environment for the application and worker.

The web application and worker should have separate processes, but share the same Go domain packages.

For persistence, I would use ordinary relational tables for business state, append-only records for important events, and an outbox pattern for reliable integration processing.

I would not implement a full event-sourced architecture in version one.

One important rule: every payroll run must be reproducible from its inputs, configuration, applicable rule-pack version and calculation version.

Otherwise, we will have difficulty investigating why one run differs from another.

## 8. The database design

Here is the initial domain model I would use.

| Domain                  | Important tables                                                  |
| ----------------------- | ----------------------------------------------------------------- |
| Organizations           | tenants, companies, sites, departments                            |
| HR                      | employees, employment_contracts, employee_status_history          |
| Rules                   | rule_packs, rules, rule_sources, rule_tests                       |
| Scheduling              | shift_templates, rosters, schedule_assignments                    |
| Time                    | clock_events, attendance_records, timesheets                      |
| Leave                   | leave_types, leave_requests, leave_transactions                   |
| Payroll                 | pay_periods, pay_components, payroll_runs, payroll_items          |
| Statutory contributions | cpf_calculations, sdl_calculations                                |
| Integrations            | integration_configs, api_requests, webhook_events, export_batches |
| Operations              | incidents, hypotheses, evidence, agent_runs, approvals            |
| Audit                   | audit_events, correlation_records, job_attempts                   |

I particularly want to emphasize `clock_events`, `attendance_records`, `timesheets` and `payroll_items`.

These should not be treated as interchangeable data.

For example:

- `clock_events` records what the system received.
- `attendance_records` records the interpreted attendance.
- `timesheets` records the reviewed and approved hours.
- `payroll_items` records what the payroll engine actually calculated.

This separation is extremely useful for Tier 2 investigations.

An engineer can trace exactly where data changed or disappeared.

## 9. What should our first MVP contain?

I would resist building every Singapore HR rule at once.

The first vertical slice should be:

Employee → Roster → Attendance → Overtime → Payroll → CPF/SDL → Payslip.

That gives us a working end-to-end business process.

I would organize the project into these implementation milestones:

1. Milestone 1 — Singapore rulebook

   Collect and classify the initial MOM and CPF/SDL material. Produce a versioned, source-cited rule specification with executable test cases.

   Acceptance: each implemented statutory calculation has a source reference, applicability conditions and passing tests.
2. Milestone 2 — Working HR/WFM application

   Implement employee records, employment terms, rosters, clock events, attendance, leave and approval workflows.

   Acceptance: the company can hire employees, schedule shifts, record attendance and approve timesheets.
3. Milestone 3 — Payroll engine

   Implement basic salary, overtime, applicable holiday/rest-day payments, CPF and SDL for the supported employee categories.

   Acceptance: payroll calculations are repeatable and independently checked against source-derived test cases.
4. Milestone 4 — Synthetic data factory

   Generate a fictional company, employees, contracts, six months of operational history, payroll runs and integration activity.

   Acceptance: every generated record has consistent lineage and valid relationships.
5. Milestone 5 — Production incident factory

   Introduce reproducible incidents into the running application, with hidden ground truth and observable symptoms.

   Acceptance: at least ten incidents can be reproduced and independently diagnosed.
6. Milestone 6 — Agent-driven investigation

   Build the ticket intake, hypothesis generation, evidence collection, verification and final investigation report.

   Acceptance: agents can investigate seeded incidents through the same tools provided to human engineers.

## 10. How do we know whether this is realistic enough?

This is an important question. We should not congratulate ourselves merely because the application has many tables and thousands of records.

I would define four levels of simulation fidelity.

| Level            | What must be realistic                                                              |
| ---------------- | ----------------------------------------------------------------------------------- |
| L1 — Business    | Statutory rules, employment terms, calculations and workflow                        |
| L2 — Data        | Relationships, records, history, volumes and data lineage                           |
| L3 — Integration | API behavior, retries, failures, duplicate messages and configuration               |
| L4 — Operations  | Logs, permissions, asynchronous jobs, performance issues and incident investigation |

Our initial release should pass L1 and L2, and demonstrate selected L3 and L4 incidents.

We do not need to recreate an entire enterprise platform, but we need enough of its behavior to make troubleshooting meaningful.

### Build an objective acceptance test

For every incident, maintain two separate artifacts:

The incident presented to the engineer

The complaint, employee identifier, pay period, and observable symptoms.

The hidden ground-truth record

The deliberately introduced defect, expected evidence, correct explanation, expected business result, and acceptable resolution.

Then evaluate the agent and human candidate against the same ground truth.

For example, the evaluator should be able to say:

- Correctly identified the problem category.
- Found the relevant employee and payroll records.
- Identified the actual defective mapping.
- Did not falsely blame CPF or MOM rules.
- Correctly explained the effect on the payslip.
- Recommended an appropriate correction.
- Did not claim to have applied an unapproved change.

We can measure diagnostic accuracy, evidence quality, time to diagnosis and unsupported conclusions.

This is also how we make the training component credible.

## 11. How I would turn this into a business

I see three possible customers, but I would prioritize them differently.

| Customer                            | What they buy                                                      |
| ----------------------------------- | ------------------------------------------------------------------ |
| SaaS companies                      | A production-like investigation environment for support teams      |
| BPO and technical-support employers | Practical assessment, onboarding and continuing training           |
| Aspiring support engineers          | Access to realistic labs and demonstrable investigation portfolios |

I would pursue B2B customers first, particularly companies that hire or train application support engineers.

I would not depend on unemployed job seekers to finance the business. They are the people we want to help, not necessarily the best initial revenue source.

A possible commercial offering:

WorkforceOps Lab — Tier 2 Support Engineering Assessment

An employer gives a candidate access to a simulated WFM/payroll environment, assigns a ticket, and assesses how the candidate investigates it.

The employer sees the candidate's SQL, diagnostic reasoning, evidence trail, hypothesis elimination and final report.

This is stronger than a multiple-choice technical assessment because it demonstrates whether someone can actually investigate an enterprise application.

# 12. My recommendation: start with one complete, convincing demonstration

I would not begin by writing the whole application.

I would begin with one vertical slice and make it exceptionally realistic.

First demo target

# The Missing Overtime Investigation

A fictional Singapore workforce company has 120 employees. One employee's overtime is missing from the payslip.

The system contains the employee contract, work schedule, clock-in/out history, approved timesheet, payroll calculation, pay-code mapping, export activity and payslip.

The investigator must find the defect and prove its effect.

The final demonstration should show:

1. A real employee and working history.
2. A real payroll calculation.
3. A deliberately introduced integration/configuration defect.
4. A customer ticket.
5. An agent querying the actual simulator.
6. An independent verification of its findings.
7. A final evidence-backed technical report.

That demonstration would prove something substantial: we can build and operate a realistic Go enterprise application, understand Singapore workforce business rules, troubleshoot multi-system failures, and use agents responsibly.

It would also give us a reusable platform for creating more complicated cases without needing to obtain access to a new client's production systems every time.

One final strategic point: do not market this as a Singapore payroll compliance product. Market it as a realistic workforce-systems simulation and technical support training/assessment environment. Its statutory rule packs should be carefully sourced, versioned and tested, but a simulation is not a substitute for legally reviewed payroll software.

My proposed first engineering artifact should therefore be a WorkforceOps Lab v0.1 Product and System Specification, defining the business workflows, rule-source register, database model, synthetic data strategy, incident architecture, Go modules and acceptance tests before implementation. That specification will be the contract for the coding agent, so it builds one coherent working system rather than several disconnected demonstrations.
