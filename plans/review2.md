I reviewed the five revised documents as revision 2 following plan_prompt4, with particular attention to the new multi-employer requirement and whether the previous review findings have actually been incorporated.

## Overall verdict



Substantially incorporated

&#x20;The multi-employer-ready architecture is now present across the five documents.



The revised design correctly establishes:

- `employer` as the legal employing entity and tenant root.
- `employer_id` as the business-data ownership boundary.
- Composite keys and foreign keys to prevent cross-employer relationships.
- Explicit employer context passed through application services.
- Employer-scoped API paths and CLI selection.
- One employer operationally in v0.1, with production multi-employer authorization and isolation deferred.

The distinction between a BPO client and a legal employing entity is also explicit. That is important: Meridian's clients are not tenants of the WorkforceOps database.&#x20;

domain-model(2).md

mvp-scope(1).md



The previous findings concerning loopback-only API access, attendance intervals, CPF OW/AW treatment, SHG instructions, roster revisions, skill requirements, KET/payslip schemas and legal-rule gating have also been addressed.

I would not freeze the design just yet. I found five remaining issues that should be corrected first.

## 1. P0 — The employer-scoping rule is not completely consistent

The database document says every employer-owned table carries `employer_id`. However, the domain model omits it on several child entities, including:

- `programme_required_skill`
- `shift_segment`
- `attendance_correction`
- `timesheet_line`
- `timesheet_line_interval`
- `payroll_line`

The database document itself describes some of these as employer-scoped "via programme", "via template" or "via line".&#x20;

database-design(2).md

domain-model(2).md

domain-model(2).md

domain-model(2).md



There are two legitimate designs:

1. Every employer-owned table carries `employer_id`, with composite foreign keys.
2. Some child tables inherit employer ownership exclusively through a parent relationship, and all access and relationships are constrained through that parent.

The current documents claim the first but sometimes describe the second.

My recommendation: choose the first.

For this project, consistently carrying `employer_id` is easier to inspect, query, test and eventually secure with PostgreSQL RLS.

Make it explicit in every employer-owned entity, including child and join tables. Use composite foreign keys to the parent and, where necessary, to the employee or other related entity.

Do not leave this as an implementation detail.

## 2. P0 — The immutability rules prevent normal draft editing

This is the most concrete database defect I found.

The database design revokes UPDATE and DELETE for:

- `roster_assignment`
- `timesheet_line_interval`

But these records must be editable while their parent roster revision or timesheet is still a draft.

The document subsequently proposes triggers that freeze roster assignments only after publication and timesheet lines after approval. That logic cannot provide the intended behavior if the application role has already lost permission to update or delete those records entirely.&#x20;

database-design(2).md



Correction:

| Record                        | Draft       | Published/approved |
| ----------------------------- | ----------- | ------------------ |
| Roster assignment             | Mutable     | Immutable          |
| Timesheet line                | Mutable     | Immutable          |
| Timesheet interval allocation | Mutable     | Immutable          |
| Roster revision               | Insert-only | Insert-only        |
| Clock event                   | Insert-only | Insert-only        |
| Attendance interval           | Insert-only | Insert-only        |
| Payslip                       | Insert-only | Insert-only        |

Remove the unconditional UPDATE/DELETE revocation for `roster_assignment` and `timesheet_line_interval`.

Use parent-state-aware triggers to enforce immutability once published or approved.

This preserves the intended rule without making normal operations impossible.

## 3. P0 — Clock-event uniqueness must include the employer

The current constraint is:

```
UNIQUE(source, source_event_id)
```

That assumes a source event ID is globally unique across every employer in the entire database. The design does not establish that assumption.&#x20;

database-design(2).md



Use:

```
UNIQUE (employer_id, source, source_event_id)
```

This gives each employer an independent event namespace while preventing duplicate delivery within that employer.

The same principle must be applied to any future integration identifiers that are scoped to a particular employer.

## 4. P0 — Payroll blocking behavior needs one unambiguous rule

The Singapore rule-scope document says an unverified rule pack blocks the affected employee's payroll result. But the domain model represents a payroll run with one status:

`draft → validated → approved`

There is no per-employee payroll approval status or partial-run mechanism.&#x20;

singapore-rule-scope(1).md

domain-model(2).md



We should not introduce partial payroll approval in this MVP.

My recommendation:

Use all-or-nothing approval for an employer's regular payroll run.

- If every included employee's required calculations pass, the run can be validated and approved.
- If any included employee has an unresolved statutory calculation, the run cannot be approved.
- The operator may explicitly exclude that employee before validation, with an audit reason, and handle the employee through a separately controlled payroll process.
- Never silently omit a blocked employee or present an incomplete run as complete.

This is simpler and safer than designing partial-run accounting and reconciliation in v0.1.

The rule-pack blocking matrix should explicitly say that an affected employee blocks the entire run unless deliberately excluded before validation.

## 5. P0 — The CLI employer identifier has no defined database field

The CLI documents:

```
workforceops --employer MERIDIAN employee list
```

But the employer entity currently has:

```
employer(id, name, uen, address, jurisdiction)
```

There is no `code` or `slug` corresponding to `MERIDIAN`.&#x20;

api-cli-contract(1).md

domain-model(2).md



Add:

```
employer(
    id,
    code,
    name,
    uen,
    address,
    jurisdiction
)
```

With:

```
UNIQUE (code)
UNIQUE (jurisdiction, uen)
```

The CLI can then accept either an employer code or UUID, while still defaulting to the only configured employer in v0.1.

## 6. Additional improvements before design freeze

These are important, but I would not turn them into another architecture project.

| Priority | Finding                                                                                                            | Recommended correction                                                                                                  |
| -------- | ------------------------------------------------------------------------------------------------------------------ | ----------------------------------------------------------------------------------------------------------------------- |
| P1       | API has no employer rule-pack activation resource.                                                                 | Add a read-only employer activation/status endpoint.                                                                    |
| P1       | `roster.legal_validation` records only `verified` or `unavailable`.                                                | Record which rule-pack version was used, when it was checked and the validation result.                                 |
| P1       | `attendance_interval.source_refs` and `payroll_line.source_refs` are insufficiently specified.                     | Define their schemas and mandatory references; use relational links where referential integrity matters.                |
| P1       | `--operator` appears only on `serve`, not clearly on CLI mutations.                                                | Define how CLI writes and API writes receive their operator attribution.                                                |
| P1       | The domain model does not specify employee-to-roster-assignment consistency at the composite-FK level.             | Ensure an attendance interval cannot reference a different employee's roster assignment, even within the same employer. |
| P2       | The SHG instruction has `active` and `opt_out`, but no explicit representation of a verified no-deduction outcome. | Define an explicit, evidenced no-deduction instruction instead of assuming missing data means no deduction.             |

The source-to-payroll evidence chain deserves particular care. The intended chain is:

`Clock events → Attendance intervals → Corrections → Approved timesheet lines → Payroll lines → Payslip`

The revised design makes significant progress, but `source_refs` must have a defined and testable meaning. Otherwise, a payroll calculation may have an inputs hash without enough structured information to reconstruct where each figure came from.&#x20;

domain-model(2).md

domain-model(2).md



## 7. What I would now freeze

The following decisions are sound and should not be reopened during the next revision:

| Architectural decision                                      | Verdict                               |
| ----------------------------------------------------------- | ------------------------------------- |
| One Go application                                          | Freeze                                |
| One PostgreSQL database                                     | Freeze                                |
| CLI and REST API only                                       | Freeze                                |
| Shared application services                                 | Freeze                                |
| Employer as legal-entity tenant root                        | Freeze                                |
| One employer operationally in v0.1                          | Freeze                                |
| Multi-employer-ready schema                                 | Freeze after corrections              |
| Composite foreign keys                                      | Freeze                                |
| No RLS in local v0.1                                        | Freeze, with production security gate |
| No microservices or event bus                               | Freeze                                |
| Singapore-specific rule packs with independent verification | Freeze                                |
| No automatic acceptance of unverified legal calculations    | Freeze                                |

The Singapore rule-scope document also correctly distinguishes global jurisdictional rules from employer-specific activation. That is the right foundation when multiple legal employers eventually use the same Singapore rule catalogue.&#x20;

singapore-rule-scope(1).md



## Final recommendation

I would call the current set Design Candidate v0.1, revision 2 — not yet frozen.

The next revision should be a narrowly scoped correction, not another redesign. Address the five P0 issues, then the six smaller improvements, and run these three groups of tests:

1. Two-employer integrity: same employee/programme codes allowed across employers, cross-employer relationships rejected.
2. Workflow integrity: draft rosters and timesheets editable; published/approved records immutable.
3. Payroll integrity: unverified rules block approval, approved results remain reproducible, and every payroll amount has traceable inputs.

Once those tests pass against actual PostgreSQL, I would be comfortable approving the design gate and moving to implementation.

The important achievement is that we now have a genuinely small application with a future-proof employer boundary. We should protect that simplicity rather than keep expanding the architecture.