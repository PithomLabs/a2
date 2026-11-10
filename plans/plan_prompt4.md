I reviewed all five documents as one design package: `mvp-scope.md`, `domain-model.md`, `database-design.md`, `api-cli-contract.md` and `singapore-rule-scope.md`.

The direction is right. I would retain the single Go application, single PostgreSQL database, CLI/API-only interfaces, and one employee-to-payroll workflow. The documents are substantially more disciplined than the earlier architecture: they remove microservices, separate databases, event buses, UI, AI and premature multi-tenancy. The scope also correctly makes unresolved legal rules a blocker rather than allowing the implementation to guess.&#x20;

mvp-scope.md

database-design(1).md

singapore-rule-scope.md



However, I would not freeze these documents for implementation yet. There are several cross-document gaps that could cause us to build a backend that appears to work but cannot reliably produce a legally defensible payroll result.

The most important findings are:

1. The API is configured to listen on `:8081`, while the domain model says there is no authentication and this is a localhost tool. That is an unsafe default for a mutating API.
2. The WFM model cannot yet enforce its promised skill eligibility or reliably represent roster revisions and rotating assignments.
3. Attendance, approved timesheets and payroll do not yet have a sufficiently explicit calculation and evidence chain.
4. The payroll scope includes CPF, SDL and SHG, but the employee and payroll data models do not yet contain everything needed to calculate all three reliably.
5. The acceptance criterion is broader than the verified rule packs currently support.

I’ll separate actual design defects from deliberate MVP exclusions.## 1. The largest issue: the system's claimed legal coverage exceeds its data model

The scope says the system will maintain employment records and generate KET information and itemised payslips. But the current employee entity does not capture several fields required by MOM's employment-record requirements.

For employees covered by the Employment Act, MOM lists employee-record fields including address, NRIC number, date of birth, gender, employment start and end dates, working hours including meal and tea breaks, and leave/public-holiday details. Salary records must also capture the prescribed salary components and overtime details.&#x20;

[image](https://www.google.com/s2/favicons?domain=https://www.mom.gov.sg\&sz=32)

mom.gov.sg+1



The current `employee` entity has name, date of birth, citizenship and status, but deliberately omits NRIC and does not include address or gender. The current timesheet and payslip structures also need more explicit fields to demonstrate that the required information is retained.&#x20;

domain-model(1).md



My recommendation: keep real NRIC handling out of the initial synthetic-data environment, but change the product claim. It should say MOM-informed reference implementation with a restricted employee-record profile, not complete statutory employment-record compliance. Add explicit fields to the design for the records we intend to support. Mark identity fields requiring real-P​​II protection as deliberately unsupported.

That preserves the no-real-PII constraint without pretending that an incomplete record is legally complete.

## 2. Payroll: three material design corrections

### CPF needs a wage-classification model, not just a Boolean

The domain model currently proposes `pay_code.cpf_payable`. That is insufficient.

CPF distinguishes Ordinary Wages (OW) and Additional Wages (AW). Overtime is CPF-payable, but its classification depends on when it is payable and which calendar month the overtime was performed. CPF requires overtime pay to be apportioned by the calendar month in which it was performed before classification. The AW ceiling is annual.&#x20;

[image](https://www.google.com/s2/favicons?domain=https://www.cpf.gov.sg\&sz=32)

cpf.gov.sg+3



This matters because a November overtime payment might not all belong to November's OW.

Minimum correction: replace the Boolean-only model with a wage classification:

- `OW`
- `AW`
- `NON_CPF_WAGE`

A payroll component must carry its CPF treatment and the relevant work period, payable date and contribution period. Do not implement AW as an afterthought; it is already part of the CPF rule pack listed in your document.&#x20;

singapore-rule-scope.md



The 2026 CPF rates also need effective-dated bands, age classification, contribution ceilings and the documented rounding sequence. The CPF Board's current page confirms the 2026 rates and a further senior-worker rate increase from 1 January 2027.&#x20;

[image](https://www.google.com/s2/favicons?domain=https://www.cpf.gov.sg\&sz=32)

How much CPF contributions to pay+3



### SHG is in scope, but its required employee inputs are missing

The model has an `shg_contribution` rule table, but no employee-level instruction identifying which applicable fund contribution should be deducted.

The CPF Board says the applicable SHG is determined by race or religion, and employees may contact the relevant group to request a different contribution amount or to stop contributing.&#x20;

[image](https://www.google.com/s2/favicons?domain=https://www.cpf.gov.sg\&sz=32)

Contributions to self-help groups+2



You also deliberately omitted NRIC and have not included race or religion in the employee model. That is understandable for a synthetic reference system, but it means the current model cannot determine SHG liability.

Minimum correction: do not store race or religion simply to make payroll work. Store an effective-dated, auditable employee SHG instruction: applicable fund, contribution status, amount or rule basis, source/declaration and effective date. Never infer `NONE` from missing information. If the instruction is absent or ambiguous, payroll validation should block the affected employee's final net-pay calculation.

### SDL must remain an employer liability

The existing scope correctly puts employer levies alongside employer contributions. Preserve that distinction in the calculation and payslip: SDL must not reduce employee net salary.

The current CPF Board description specifies SDL at 0.25% of monthly total wages, with a S$2 minimum and S$11.25 maximum per employee, followed by aggregation and downward rounding.&#x20;

[image](https://www.google.com/s2/favicons?domain=https://www.cpf.gov.sg\&sz=32)

Skills Development Levy



## 3. The WFM model has three missing relationships

The scope promises skill-based eligibility, rotating rosters and controlled roster revisions. The domain model does not fully implement those promises yet.&#x20;

mvp-scope.md



Skill eligibility: `employee_skill` records what an employee can do, but there is no `programme_required_skill` or equivalent table stating what a programme requires. Consequently, the system has no rule against assigning an employee to a programme for which they are unqualified.

Add one small relationship:

`programme_required_skill(programme_id, skill_id, minimum_level)`

Roster validation can then compare programme requirements against the employee's effective-dated skills. No advanced multiskill optimization is required.

Rotation patterns: `rotation_pattern` is currently an orphan. Nothing associates a rotation pattern with an employee or team, and nothing defines how it generates roster assignments. Either add an effective-dated employee/team rotation assignment, or remove rotation patterns from v0.1 and explicitly make rosters manually assigned. I recommend keeping simple, deterministic rotation.

Roster revisions: `roster_revision` stores a version and diff, but `roster_assignment` does not identify which revision it belongs to. That makes it difficult to reconstruct exactly what was published. Make assignments members of an immutable roster revision. Publishing a change should create a new revision rather than silently changing the previous published schedule.

These are small design corrections, not a reason to introduce another service or a scheduling engine.

## 4. Attendance-to-payroll needs a more precise data chain

This is the most important WFM-to-payroll integration gap.

The model distinguishes scheduled, recorded, actual and paid minutes, which is correct. But the current attendance entity is still one row per employee and business date, while clock events only have `in` and `out` kinds.&#x20;

domain-model(1).md



That does not fully define:

- How unpaid meal breaks are treated when actual break duration differs from the planned duration.
- How multiple work intervals within one shift are represented.
- How an overnight shift is divided when it crosses a calendar day or payroll-period boundary.
- How actual overtime is attributed to the calendar month in which it was performed.
- How approved timesheet lines can be independently traced to the original punches and attendance corrections.

This last issue matters for CPF as well as MOM overtime. CPF classifies overtime by the month the work was performed and the date it becomes payable, not simply the month in which the payroll batch is run.&#x20;

[image](https://www.google.com/s2/favicons?domain=https://www.cpf.gov.sg\&sz=32)

cpf.gov.sg+1



My minimum correction: preserve the raw clock events, but introduce an explicit attendance-interval or time-entry layer. Each interval should have start and end instants, employee, roster assignment, local business context, and source references. Break intervals must be distinguishable from work intervals.

Then use this lineage:

`Clock event → Attendance interval → Attendance review/correction → Approved timesheet line → Payroll component`

Timesheet lines should retain the underlying time interval references and the applicable pay period. Cross-midnight intervals must be split for calendar-based calculations where required, rather than simply assigned to the shift's starting date.

Also, do not treat scheduled unpaid breaks as proof that the employee actually took those breaks. Either record actual break intervals or state and enforce a narrow assumption for the demo environment.

## 5. Rotating BPO shifts conflict with the current legal-rule boundary

This is a particularly important scope inconsistency.

The MVP includes day/evening/night shifts and rotating rosters, but the Singapore rule-scope document defers shift-worker averaging.&#x20;

mvp-scope.md

singapore-rule-scope.md



MOM's current guidance distinguishes ordinary five-/six-day arrangements from shift arrangements. For shifts of up to 12 hours per day, it describes an average of 44 hours over a continuous three-week period. It also describes a 30-hour continuous rest period for shift workers.&#x20;

[image](https://www.google.com/s2/favicons?domain=https://www.mom.gov.sg\&sz=32)

mom.gov.sg



So we cannot simply call a rotating BPO roster Singapore-law compliant while excluding the applicable shift-worker rules.

I recommend adding a narrowly scoped `SG-MOM-SHIFT` rule pack covering only the shift arrangements we actually support. It need not implement rest-day work pay, advanced roster optimization or every exceptional arrangement. It should establish whether a proposed roster satisfies the supported hours and rest-period conditions.

If that rule pack is not independently verified, label the roster as operationally scheduled, legal validation unavailable. Do not silently treat the schedule as compliant.

This is the one area where trimming too aggressively would undermine the product's central BPO use case.

## 6. The API and database need a couple of integrity and safety changes

### API must actually be local-only

The API contract says:

```
workforceops serve    # run the REST API (default :8081)
```

But the domain model says there is no authentication in v0.1 because it is a localhost tool.&#x20;

api-cli-contract.md

domain-model(1).md



Those statements do not agree. In Go, listening on `:8081` normally binds the wildcard interface, not just loopback.

Change the default to:

```
127.0.0.1:8081
```

For v0.1, reject non-loopback binding. Do not mistake an operator string in an audit record for authentication: anyone able to call an unauthenticated API can claim to be any operator. That is acceptable only for a local synthetic-data development environment, not for a remotely accessible business application.

### Append-only records are not protected by the database design

The database design grants the `app` role CRUD on every table, while the domain model says clock events, approved timesheets, payslips and other records are immutable. The database design then relies on application discipline and tests.&#x20;

database-design(1).md

database-design(1).md



For a payroll system, that is too weak for the most important records.

Keep the simple two-role arrangement, but make the important invariants explicit:

- Clock events, audit events and payslips: insert-only.
- KET snapshots: insert-only; a changed employment term creates another snapshot.
- Approved timesheet lines: no editing or deletion after approval.
- Approved payroll runs and their lines: locked; corrections create another run.
- Roster publication: a new revision, not an overwrite.

This can be implemented with narrow SQL privileges and a small number of database triggers. It does not require RLS, another database role for every module, or a distributed architecture.

Also change clock-event deduplication to a unique key on `(source, source_event_id)`, unless the design can guarantee that all source systems issue globally unique event IDs. A vendor-specific event ID is not necessarily globally unique.&#x20;

database-design(1).md



## 7. The CLI contract does not yet operate the whole business workflow

The API lists many resources, but the CLI currently has mostly list and workflow verbs. It does not explicitly provide commands for creating or changing employment terms, programme setup, team memberships, skills, shift templates, rotation assignments or roster assignments.&#x20;

api-cli-contract.md



That leaves two possibilities: users must use raw SQL to configure the system, or the seeded demo is the only usable environment.

Neither is desirable.

Keep the CLI small, but add the minimum commands required to configure and operate the system. For example:

```
workforceops employee terms set EMP-001 --file terms.json

workforceops client create --file client.json
workforceops programme create --file programme.json
workforceops team create --file team.json
workforceops skill assign EMP-001 --file skills.json

workforceops shift-template create --file shift.json
workforceops rotation assign --file rotation.json
workforceops roster create --file roster.json
workforceops roster publish ROSTER-001
```

File-based inputs are preferable to dozens of fragile command-line flags. The same operations must be available through the REST API, and both must call the same application services.

I would also remove the `exit 2 = unimplemented slice` behavior from the public CLI contract. An unsupported business case should have a meaningful error; an unimplemented command should not be advertised as a usable command in the first place. For a valid but unsupported legal case, use the documented machine-readable `422` error.&#x20;

api-cli-contract.md



## 8. KETs and payslips need explicit schemas

The current model stores KETs and payslips as generic `payload` records. That is flexible, but it is not a sufficient specification for a statutory record.&#x20;

domain-model(1).md

domain-model(1).md



MOM's KET requirements cover more than the fields presently listed in `employment_terms`: job duties, fixed-term duration where applicable, daily working hours, working days, rest day, overtime period and rate, leave types, medical benefits and probation, among other items.&#x20;

[image](https://www.google.com/s2/favicons?domain=https://www.mom.gov.sg\&sz=32)

mom.gov.sg



I recommend:

- Define a versioned KET snapshot schema.
- Generate the snapshot from effective-dated employment terms and compensation.
- Include only applicable fields, but explicitly represent why a field is not applicable.
- Keep every issued snapshot immutable.

Similarly, define the payslip output schema explicitly. It should include the applicable salary period, payment date, basic salary, allowances, additional payments, deductions, overtime hours and pay, and net salary.&#x20;

[image](https://www.google.com/s2/favicons?domain=https://www.mom.gov.sg\&sz=32)

mom.gov.sg



MOM also specifies retention periods for employment and salary records: the latest two years for current employees, and the latest two years retained for one year after an employee leaves. Add a retention requirement to the design and acceptance tests.&#x20;

[image](https://www.google.com/s2/favicons?domain=https://www.mom.gov.sg\&sz=32)

mom.gov.sg+1



Do not build a delivery notification system yet. Just record whether a payslip was generated and issued, and when.

## 9. The research gate and the demo seed need to agree

The rule-scope document says all packs begin `unverified` and cannot enter implementation until they reach `IMPLEMENTABLE`.&#x20;

singapore-rule-scope.md



The database design, however, says `seed demo` loads rule data from verified research.&#x20;

database-design(1).md



Make this an explicit runtime rule:

- Unverified rule data can be stored for research, but cannot be activated for payroll.
- Only an `IMPLEMENTABLE` rule-pack version can be used in a payroll calculation.
- Each payroll run records the exact rule-pack versions used.
- An unsupported or blocked calculation prevents approval of that employee's payroll result.
- A rule update must never change the result of an already approved payroll run.

There is also an important existing blocker in the overtime pack. MOM's current page displays the non-workman hourly cap as S$13.60 while giving a monthly-rated formula that produces a different value at the S$2,600 threshold. The research document correctly flags this discrepancy; retain it as a blocker until independently resolved. Do not hardcode either value just to complete a demo.&#x20;

[image](https://www.google.com/s2/favicons?domain=https://www.mom.gov.sg\&sz=32)

mom.gov.sg

singapore-rule-scope.md



The same rule gate should apply to unpaid-leave deductions and public-holiday calculations, whose treatment is not yet completely verified in your rule-scope document.

# 10. My consolidated amendment list

I would make the following changes before declaring the design frozen.

| Priority | Document                  | Required amendment                                                                                                         |
| -------- | ------------------------- | -------------------------------------------------------------------------------------------------------------------------- |
| P0       | `mvp-scope.md`            | Clarify the restricted legal coverage profile; distinguish operational scheduling from verified legal compliance.          |
| P0       | `domain-model.md`         | Add CPF OW/AW classification, SHG employee instructions, structured KET/payslip records, and attendance intervals.         |
| P0       | `domain-model.md`         | Add programme skill requirements, rotation assignment and roster revision linkage.                                         |
| P0       | `database-design.md`      | Protect immutable and approved records; correct source-event uniqueness; define the attendance/time-entry relationships.   |
| P0       | `api-cli-contract.md`     | Bind API to loopback by default; define operator attribution and complete the minimum configuration commands.              |
| P0       | `singapore-rule-scope.md` | Resolve the overtime cap conflict; introduce shift-worker validation or explicitly mark that legal validation unavailable. |
| P1       | `domain-model.md`         | Define actual break intervals, cross-midnight payroll attribution and timesheet source lineage.                            |
| P1       | `database-design.md`      | Add explicit employee and payroll record-retention requirements.                                                           |
| P1       | `api-cli-contract.md`     | Define the structured request/response contracts, idempotency behavior and workflow transitions.                           |
| P1       | `singapore-rule-scope.md` | Specify exactly which rule packs block which payroll components and scenarios.                                             |

I would not add a new microservice, event bus, multi-tenant framework, web UI, simulation engine or generic plugin system as part of these amendments.

# 11. Final assessment

Architecture direction

Retain

Single Go application, one PostgreSQL database, shared services, CLI and REST API.

MVP scope

Retain with corrections

One Singapore employer, one BPO reference operation, restricted employee population, one monthly payroll workflow.

Domain and database

Revise

The relationships and auditability are not yet sufficient for the promised workflow.

Singapore rule coverage

Not implementation-ready

Several load-bearing rules remain unverified. The implementation gate must remain closed for those calculations.

One final distinction is important: this is a single-employer local reference backend, not yet a production-ready HR/payroll SaaS product. That is a perfectly reasonable v0.1 boundary. Synthetic data, no remote unauthenticated access, no real NRIC handling and no claim of universal statutory compliance are all appropriate constraints.

My recommendation is to preserve the five-document structure and make these surgical amendments, not start another architecture redesign. The next deliverable should be the revised five files, with the existing naming, organization and level of detail retained, followed by a final design-gate review before implementation.


## prompt

one more requirement, single employer for this MVP but it should be ready for multi-employer database design


## answer

Agreed. This is an important architectural requirement, and it should be incorporated now, at the database design stage, rather than added when we eventually support multiple employers.

The decision should be:

> WorkforceOps Lab v0.1 supports one employer operationally, but its database schema and application-service boundaries must be multi-employer capable from the beginning.

This does not mean building multi-tenant SaaS in v0.1. It means avoiding a single-employer schema that we would later have to restructure.

## 1. Make `employer_id` the tenant boundary

The current database design explicitly excludes tenant-scoping columns, and several entities are identified only by their own IDs. That needs to change.&#x20;

database-design(1).md



The `employer` table becomes the root of every employer-owned business record.

Employer

Tenant root · Legal entity · UEN · Jurisdiction

HR

Employees · Terms · Leave

BPO operations

Clients · Programmes · Teams

WFM

Rosters · Attendance · Timesheets

Payroll

Pay periods · Runs · Payslips

Employer-scoped database records

Every business record belongs to exactly one employer

## 2. Add `employer_id` to all employer-owned tables

This is not just an `employer_id` column on employees. It must be consistently applied throughout the business model.

| Domain     | Tables that must be employer-scoped                                                          |
| ---------- | -------------------------------------------------------------------------------------------- |
| HR         | employee, status history, employment terms, compensation, KET, leave                         |
| BPO        | client, programme, team, team membership                                                     |
| WFM        | skills, shift templates, rotations, rosters, assignments, clock events, attendance           |
| Timesheets | timesheets, timesheet lines                                                                  |
| Payroll    | pay codes, pay periods, payroll runs, payroll lines, payslips, corrections                   |
| Operations | audit events, idempotency records, employer-specific staffing requirements and configuration |

The `employer` table itself is the root, so it does not need an `employer_id` referencing itself.

Some reference data should remain global rather than being duplicated unnecessarily:

- Immutable Singapore statutory rule-pack definitions and versions.
- Jurisdiction-level public reference information.
- System-wide migration metadata.

An employer-specific configuration or rule-pack activation record should still identify the employer.

## 3. Enforce employer boundaries with composite foreign keys

This is the most important database-level change.

Suppose we have two employers:

- Meridian CX Services Pte. Ltd.
- Acme Operations Pte. Ltd.

Both might have an employee numbered `EMP-001`. That must be valid because employee numbers are unique within an employer, not necessarily across the entire system.

Use:

```
UNIQUE (employer_id, employee_no)
```

The same principle applies to programme codes, team codes, pay-period codes and other employer-owned business identifiers.

More importantly, relationships must prevent cross-employer references.

For example:

```
ALTER TABLE employee
    ADD CONSTRAINT employee_employer_id_id_unique
    UNIQUE (employer_id, id);

ALTER TABLE team
    ADD CONSTRAINT team_employer_id_id_unique
    UNIQUE (employer_id, id);

ALTER TABLE team_membership
    ADD CONSTRAINT membership_employee_fk
    FOREIGN KEY (employer_id, employee_id)
    REFERENCES employee (employer_id, id);

ALTER TABLE team_membership
    ADD CONSTRAINT membership_team_fk
    FOREIGN KEY (employer_id, team_id)
    REFERENCES team (employer_id, id);
```

The same principle must cover the entire relationship graph.

An employee belonging to Employer A must never be assignable to Employer B's team, roster, timesheet or payroll run.

This is stronger than relying on application code to remember to compare employer IDs.

## 4. Carry employer context through application services

The database is only one side of the design.

Every application service and repository operation must operate within an explicit employer context.

Conceptually:

```
type EmployerID uuid.UUID

type EmployeeService interface {
    Create(
        ctx context.Context,
        employerID EmployerID,
        input CreateEmployeeInput,
    ) (*Employee, error)

    Get(
        ctx context.Context,
        employerID EmployerID,
        employeeID uuid.UUID,
    ) (*Employee, error)
}
```

Apply the same rule to HR, WFM, attendance, timesheets and payroll.

The CLI and REST API must resolve the employer context at their boundaries and pass it down to the shared services. A service must never assume that an unqualified employee ID identifies a globally accessible employee.

This also preserves the existing architectural decision that the CLI and API use the same application services.&#x20;

api-cli-contract.md



## 5. Make the API employer-scoped

Use employer-scoped resource paths from the beginning:

```
/api/v1/employers/{employer_id}

/api/v1/employers/{employer_id}/employees
/api/v1/employers/{employer_id}/clients
/api/v1/employers/{employer_id}/programmes
/api/v1/employers/{employer_id}/teams
/api/v1/employers/{employer_id}/rosters
/api/v1/employers/{employer_id}/attendance
/api/v1/employers/{employer_id}/timesheets
/api/v1/employers/{employer_id}/payroll-runs
/api/v1/employers/{employer_id}/payslips
```

For the MVP, the application is configured for one employer and rejects requests outside that configured employer context.

The CLI should support:

```
# Single-employer default
workforceops employee list

# Explicit employer context
workforceops --employer MERIDIAN employee list
```

The default may be omitted only while exactly one employer is configured. When the system eventually supports multiple employers, the user must select an employer context explicitly.

An employer ID in a URL is not authentication. Actual multi-employer deployment must introduce identity, authorization and verification that the authenticated principal is permitted to access the requested employer.

## 6. What about PostgreSQL RLS?

I would make a deliberate distinction between schema readiness and production multi-tenant security.

For v0.1:

- Add `employer_id` consistently.
- Add composite foreign keys and tenant-scoped unique constraints.
- Require employer context in every service and repository.
- Test that queries and business operations cannot cross employer boundaries.
- Keep the local-only API and the existing simple database role arrangement.
- Do not implement full multi-tenant authentication or PostgreSQL RLS yet.

Before enabling multiple real employers in a shared production database:

- Introduce a verified tenant-access model.
- Implement PostgreSQL Row-Level Security or an equivalently robust database-isolation boundary.
- Ensure the database's tenant identity cannot be freely changed by an untrusted API caller.
- Add adversarial cross-employer access tests.
- Review backup, restore, export, deletion and audit isolation by employer.

I would not pretend that adding `employer_id` alone provides production-grade tenant security. It provides the correct data architecture; the authorization and isolation layer is a separate release gate.

## 7. Required acceptance tests

Even though the MVP only exposes Meridian CX, the database design must be tested with two employer fixtures.

| Test                                                   | Expected result                                         |
| ------------------------------------------------------ | ------------------------------------------------------- |
| Create two employers                                   | Both can exist in the database                          |
| Same employee number under both employers              | Allowed                                                 |
| Same programme code under both employers               | Allowed                                                 |
| Employer A employee assigned to Employer B team        | Rejected                                                |
| Employer A timesheet references Employer B employee    | Rejected                                                |
| Employer A payroll run references Employer B timesheet | Rejected                                                |
| Employer A API request accesses Employer B employee    | Rejected                                                |
| Employer A payroll export includes Employer B records  | Impossible                                              |
| Delete or deactivate an employer                       | Must not silently orphan or expose its business records |

These should be actual PostgreSQL integration tests, not merely mocked repository tests.

## 8. Amend the five design documents

| Document                  | Required change                                                                                          |
| ------------------------- | -------------------------------------------------------------------------------------------------------- |
| `mvp-scope.md`            | State: one active employer operationally; multi-employer-capable database and service architecture.      |
| `domain-model.md`         | Add employer ownership to every employer-scoped entity and employer context to all application services. |
| `database-design.md`      | Add `employer_id`, composite unique keys, composite foreign keys and tenant-isolation tests.             |
| `api-cli-contract.md`     | Add employer-scoped resource paths and CLI employer selection.                                           |
| `singapore-rule-scope.md` | Keep legal rules jurisdiction-scoped, with employer-specific activation and applicability records.       |

One further design decision: `employer` should represent the legal employing entity, not an arbitrary customer account or BPO client. Meridian's client programmes remain children of Meridian's employer record. This distinction matters later when one business group operates multiple legal employing entities.

## Final architecture decision



Adopt

Single-employer MVP, multi-employer-ready architecture.



The scope remains deliberately small: one fictional Singapore employer, its employees, BPO clients, programmes, shifts, attendance and payroll.

But the database will be designed as though a second employer could be added tomorrow without restructuring the tables or rewriting every relationship.

That is the right compromise: prepare the foundation for multiple employers without building multi-tenant SaaS before we need it.


