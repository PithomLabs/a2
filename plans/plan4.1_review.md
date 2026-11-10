Q1 is resolved: yes, `SG_research.pdf` is the completed output of the `prompt2.md` deep-research assignment. `plan4.1(4).md` is the plan for incorporating that research into WorkforceOps Lab. I have read both, including the PDF's 27 rendered pages and the formula images and reference lists.

My overall decision is:

Approve plan4.1 with targeted corrections

The plan has the right philosophy: use the research as evidence and a source of candidate requirements, never as unquestionable truth. Its source verification, golden-test quarantine and scope controls are particularly good.&#x20;

plan4.1(4).md



However, I found several issues worth correcting before the coding agent executes it.

## 1. The most important finding: the overtime conflict is now resolved

The PDF contains two historical figures:

| PDF location | PDF claim                                    | Current position |
| ------------ | -------------------------------------------- | ---------------- |
| Page 1       | Non-workman Part IV coverage up to S$2,500   | S$2,600          |
| Page 2       | Non-workman overtime calculation cap S$2,250 | S$2,600          |

These are not two competing current interpretations.

The figures S$2,500 and S$2,250 describe the historical position before 1 April 2019. The Employment Act changes effective on that date raised the non-workman coverage threshold from S$2,500 to S$2,600 and the overtime calculation cap from S$2,250 to S$2,600. Current MOM guidance confirms the latter position.&#x20;

[image](https://www.google.com/s2/favicons?domain=https://www.mom.gov.sg\&sz=32)

mom.gov.sg+2



The current applicable rules include:

- Non-workmen: basic monthly salary not exceeding S$2,600.
- Workmen: basic monthly salary not exceeding S$4,500.
- Minimum overtime rate: 1.5 times the hourly basic rate.
- Monthly-rated hourly basic rate: \\(12 \times\\) monthly basic rate / \\(52 \times 44\\).
- Non-workman overtime calculation cap: S$2,600, equivalent to the published S$13.60 hourly rate.

MOM also specifies the 72-hour monthly overtime limit and its treatment of rest-day and public-holiday overtime.&#x20;

[image](https://www.google.com/s2/favicons?domain=https://www.mom.gov.sg\&sz=32)

mom.gov.sg+2



### An additional defect in the PDF's golden test

Page 24 contains `TC-OT-PART4-002`:

- Basic monthly salary: S$2,400.
- Statutory basic rate used: S$2,250.
- Hourly basic rate: 11.7998.
- Ten overtime hours.
- Expected payment: S$177.

This fixture is invalid for the current rules.

Since S$2,400 is below the current S$2,600 cap, the cap does not bind. The uncapped hourly calculation is:

\\[ \frac{12 \times 2400}{52 \times 44} =12.587412587\ldots \\]

The corresponding unrounded overtime amount for ten hours at 1.5 times that rate is approximately S$188.8112.

There is another problem: even using the PDF's obsolete S$2,250 input, its displayed hourly figure does not exactly follow its own formula.

Instruction: quarantine the old fixture, preserve its original contents for transcription auditing, and create a new fixture from the currently verified rule and rounding procedure. Do not simply replace 2250 with 2600.

## 2. Other substantive findings in the PDF

The research is useful, but not every section is equally reliable.

| PDF section                                    | Assessment                              | Treatment                                   |
| ---------------------------------------------- | --------------------------------------- | ------------------------------------------- |
| Pages 1–2: Employment Act Part IV              | Contains historical values              | Correct against MOM                         |
| Pages 3–5: CPF                                 | Useful detailed candidate specification | Keep gated                                  |
| Pages 6–8: SDL, SHG, IR21                      | Valuable operational coverage           | Official verification before implementation |
| Pages 9–10: Data lineage                       | Strong architectural material           | Adapt to ShiftHub/PayLoop                   |
| Pages 11–13: Failure taxonomy and competencies | Valuable training framework             | Retain with evidence qualifications         |
| Pages 14–22: Twelve incidents                  | Useful scenario candidates              | Review and selectively adapt                |
| Pages 23–24: Database and golden tests         | Useful draft material                   | Do not copy uncritically                    |
| Pages 25–27: Edge cases and sources            | Important research leads                | Verify every material claim                 |

Two particularly important corrections:

Employment classifications: The PDF groups managers/executives, domestic workers and seafarers together as excluded employees. That is misleading. Domestic workers and seafarers are excluded from the Employment Act itself, whereas managers and executives are excluded from Part IV but can still be covered by the Act's core provisions.&#x20;

[image](https://www.google.com/s2/favicons?domain=https://www.mom.gov.sg\&sz=32)

mom.gov.sg+1



Working arrangements: The PDF mainly discusses the familiar five-day and six-day work arrangements. MOM also documents alternative working arrangements, including certain shift-worker averaging arrangements. Our MVP should implement a clearly specified subset and return `indeterminate` for unsupported arrangements rather than claiming complete Part IV coverage.&#x20;

[image](https://www.google.com/s2/favicons?domain=https://www.mom.gov.sg\&sz=32)

mom.gov.sg



## 3. Corrections required in plan4.1

I agree with its seven-step process, but I would add these refinements.

### A. Change INC-001 to avoid rest-day calculations

The PDF's `SG-PAY-001` uses `OT_REST_DAY` and refers to Sunday overtime.

That conflicts with our frozen decision to defer rest-day pay.

Also, Sunday is not necessarily an employee's statutory rest day; the employer determines the rest day, subject to the applicable rules.&#x20;

[image](https://www.google.com/s2/favicons?domain=https://www.mom.gov.sg\&sz=32)

mom.gov.sg



Keep the pay-code-mapping failure mechanism, but change the scenario to ordinary overtime on a non-rest day, using a supported code such as `OT_STANDARD`.

That preserves the valuable integration investigation without requiring deferred rest-day calculations.

### B. Correct INC-002's duplicate-punch mechanism

The PDF describes duplicate events received 50 milliseconds apart, while proposing a unique constraint on `(employee_id, punch_timestamp)`.

That constraint will not prevent two different timestamps 50 milliseconds apart.

The scenario should instead model a repeated delivery of the same logical event using the same source event ID or idempotency key, with the fault disabling the corresponding uniqueness/deduplication control.

That creates a technically credible concurrency and retry problem.

### C. Narrow INC-003

`SG-INT-008` is a useful dead-letter-queue scenario.

I would use it for downstream employee-status synchronization failure, but not initially include the full final-salary overpayment and proration consequences.

Those require additional payroll and statutory logic outside the supported MVP subset.

### D. Separate the two evidence classifications

The PDF's E1–E5 taxonomy is useful, but some scenario classifications do not match their cited sources.

For example, `SG-PAY-001` is labelled E3, vendor technical documentation, but its cited reference is an old legislative document, not vendor documentation establishing the specific silent mapping failure.

I recommend two independent fields:

- `rule_evidence_class`: evidence supporting the relevant statutory or operational rule.
- `failure_mechanism_evidence_class`: evidence supporting the actual software failure mechanism.

An engineered fault can be E5 even when the statutory rule it violates is supported by E1 or E2.

Also, describe invented scenarios as reproducible synthetic incidents, not documented incidents from real customer support tickets.

### E. Strengthen the bitemporal specification

The PDF's proposed compensation table uses effective dates and `created_at`.

That is a useful starting point, but `created_at` alone does not provide a complete bitemporal history.

The architecture should distinguish:

- When a business fact is effective.
- When the system learned or recorded that fact.
- When a correction superseded a previous record.

This is particularly important for retroactive salary changes and reproducible historical payroll.

### F. Clarify exact arithmetic

The plan's decision to use exact intermediate arithmetic rather than floating point is sound.

I would specify `math/big.Rat` or an equivalently exact fixed-point approach, rather than leave the implementation agent to choose an imprecise floating-point representation.

Apply rounding only at the verified statutory boundaries.

## 4. What the plan gets particularly right

I would retain these decisions without modification:

1. Classify all 27 works-cited entries and identify irrelevant or outdated sources.
2. Verify every value against current official sources.
3. Inspect formula images and double-check their transcription.
4. Quarantine unsupported golden fixtures.
5. Incorporate the strongest incident mechanisms into the initial catalogue.
6. Preserve the existing platform architecture instead of following the PDF's microservice proposal.
7. Keep CPF, SDL, SHG, IR21 and rest-day payment outside the first rule-implementation scope.
8. Run research and design corrections in parallel, but do not implement unverified statutory calculations.

These are exactly the right safeguards.&#x20;

plan4.1(4).md



One useful verification already available from official sources: CPF Board publishes the 2026 contribution rates, the S$8,000 OW ceiling, the contribution-rounding rules and the S$102,000 annual salary ceiling. It also specifies the annual AW-ceiling calculation using estimated ordinary wages at payment time, followed by reconciliation. The official SDL rules confirm the 0.25% rate, S$2 minimum and S$11.25 maximum. These are good candidates for the later CPF and SDL research gates, not a reason to bring them into the current MVP.&#x20;

[image](https://www.google.com/s2/favicons?domain=https://www.cpf.gov.sg\&sz=32)

How much CPF contributions to pay+4



## 5. Exact amendment I would give the research agent

# PLAN4.1 REVIEW AMENDMENT — REQUIRED CORRECTIONS

Q1 is resolved: SG_research.pdf is the completed prompt2 research output. No additional research batch is currently expected.

Retain the plan4.1 verification-first methodology and implementation boundaries, subject to these corrections:

1. **Overtime historical values:** Record S$2,500 coverage threshold and S$2,250 calculation cap as historical values applicable before 1 April 2019. Current MOM guidance establishes S$2,600 for both non-workman Part IV coverage and the calculation cap. Record source, effective date and retrieval date. Do not treat these as unresolved current alternatives.
2. **TC-OT-PART4-002:** Quarantine the original PDF fixture permanently from statutory test runners. Record that its salary S$2,400 is below the current S$2,600 cap, so the cap does not bind. Its reported hourly calculation also does not exactly follow its own formula. Create a fresh golden fixture only after confirming the prescribed calculation and rounding treatment against current official MOM sources.
3. **INC-001:** Adapt SG-PAY-001 from `OT_REST_DAY`/Sunday work to ordinary overtime on a documented non-rest day using a supported pay code. Preserve the missing-mapping integration failure. Rest-day pay remains deferred.
4. **INC-002:** Correct the duplicate-punch mechanism. A unique constraint on employee ID and timestamp cannot prevent two events with distinct timestamps 50 milliseconds apart. Model duplicate delivery of the same logical source event using a source event ID or idempotency key, and inject failure of that protection.
5. **INC-003:** Use SG-INT-008 for a reproducible downstream employee-status synchronization failure caused by dead-letter handling. Do not implement final-pay proration or overpayment calculations unless their rule requirements separately pass verification.
6. **Evidence classifications:** Introduce separate rule-source and failure-mechanism evidence classifications. Do not inherit E1–E5 labels from the PDF without checking the cited sources. Classify engineered scenarios as E5 unless actual vendor or public incident evidence supports a stronger classification. Never describe synthetic scenarios as verified customer incidents.
7. **Employment Act classification:** Distinguish employees excluded from the Employment Act entirely from employees excluded only from Part IV. In particular, managers/executives must not be grouped legally with domestic workers and seafarers.
8. **Overtime scope:** Replace any claim of a complete Part IV implementation with a precise, source-verified supported subset. Define eligible employee classifications, monthly-basic-salary conditions, supported pay frequency, basic-salary definition, normal-hours arrangement, overtime formula, cap, payment deadline and unsupported cases. Unsupported arrangements must return `indeterminate`. Rest-day and public-holiday pay remain deferred.
9. **Bitemporal integrity:** Specify effective/business time and system/recorded time independently. `created_at` alone is not a complete bitemporal model. Define correction and supersession behavior and historical recalculation requirements.
10. **Arithmetic:** Use exact rational or fixed-point intermediate calculation. Do not use binary floating point for statutory monetary calculations. Define the exact legal rounding boundary for every implemented rule.
11. **Scenario 3:** Explicitly defer SG-WFM-003 because its present specification depends on rest-day treatment. Keep it in the backlog rather than implying that cross-midnight/rest-day calculations are covered by the first overtime rule pack.
12. Preserve all other plan4.1 scope restrictions. Do not introduce CPF, SDL, SHG, IR21, leave, rest-day or public-holiday calculation code into the MVP.

Deliver the updated triage, source register, conflict log, scenario mapping, golden-test disposition and roadmap reconciliation. Keep every unverified item explicitly marked as such.

## 6. Final recommendation

I would proceed with plan4.1 after incorporating that amendment.

The research has real value. Its strongest contributions are the cross-system data-lineage model, the operational failure taxonomy, the Tier 2 competency matrix and the candidate incident library.

But the PDF is a research and design input, not a validated statutory specification. Its own historical overtime mistakes demonstrate precisely why we need the independent oracle and source-pinned golden tests.

I would also confirm the existing decision: rest-day payment rules stay deferred. We can still train cross-system investigation using ordinary overtime, attendance, payroll mapping, synchronization and dead-letter failures.

The next gate should be the completion of the source-triage and rule-verification dossier—not immediate payroll coding.


## prompt

I like to apply phys-pseudo to ensure we capture the business rules accurately with all the assumptions, claims and edge cases just like what we did on the other physics thread, does this make sense from a test-driven development perspective or overkill? what do you suggest instead if you are against it


## answer

I strongly recommend applying Phys-Pseudo to WorkforceOps Lab. I don't think it is overkill. However, I would use it selectively as a formal business-rule specification and verification layer, not as a replacement for conventional test-driven development.

In fact, I think the approach could become one of WorkforceOps Lab's strongest differentiators.

You are trying to build an environment where an applicant can encounter a realistic payroll discrepancy and determine whether the cause is a statutory rule, employment contract, employer configuration, data problem, calculation defect or integration failure.

If we cannot precisely describe what the system is supposed to do, we cannot reliably determine whether it has done something wrong.

That is precisely the problem Phys-Pseudo can help solve.

# 1. Why Phys-Pseudo makes sense here

Consider a seemingly straightforward requirement:

"Calculate overtime correctly under Singapore employment rules."

That sentence is dangerously underspecified.

A correct implementation requires answers to several questions:

- Is the employee covered by Part IV of the Employment Act?
- Is the employee a workman or non-workman?
- What is the employee's monthly basic salary?
- Is the employee a manager or executive?
- Which employment arrangement applies?
- What constitutes normal working hours?
- Was the additional work requested by the employer?
- Were breaks excluded correctly?
- Which statutory rate and calculation cap apply?
- Which version of the rule was effective during the pay period?
- Are all the necessary facts actually available?

A developer might write a perfectly functioning calculation function that produces the wrong legal result because one of these conditions was misunderstood.

Phys-Pseudo forces us to establish the meaning of the rule before implementing its calculation.

I would particularly preserve the distinction between:

- `given`: facts supplied by the source system or authoritative documentation.
- `define`: the precise meaning of variables and concepts.
- `assume`: conditions under which a derivation is valid.
- `claim`: what the rule or calculation asserts.
- `derive`: how the result follows.
- `if/then/else`: conditional applicability and alternative outcomes.
- `?`: missing information, unresolved ambiguity or an unsupported case.

That is a very good foundation for formalizing business requirements.

# 2. But don't apply Phys-Pseudo indiscriminately

This is where I would exercise restraint.

Not every software requirement needs a mathematical-style derivation.

For example:

| Requirement                                 | Recommended treatment                                                        |
| ------------------------------------------- | ---------------------------------------------------------------------------- |
| Employee must have an identifier            | Database constraint + unit test                                              |
| Employee cannot belong to another workspace | Authorization invariant + integration test                                   |
| Overtime eligibility                        | Phys-Pseudo + decision table + golden tests                                  |
| CPF contribution calculation                | Phys-Pseudo + source-pinned calculation specification + golden tests         |
| Cross-midnight attendance                   | Phys-Pseudo where statutory attribution is involved + state/transition tests |
| Failed webhook retry                        | State machine + integration tests                                            |
| Pay-code mapping                            | Mapping specification + contract tests                                       |
| Oracle isolation                            | Security invariant + database permission tests                               |

My recommendation is three levels of rigor.

Level 1 — Full formalization

High-risk business rules

Use Phys-Pseudo for statutory applicability, payroll calculations, effective-dated rules, employee classification, monetary rounding and other rules where an incorrect conclusion could cause financial or compliance consequences.

Level 2 — Structured specification

Complex workflows and integrations

Use state machines, event contracts, invariants and selective Phys-Pseudo where business-rule decisions are involved.

Level 3 — Conventional testing

For simple CRUD operations, formatting, ordinary validation and straightforward API behavior, use normal Go tests, database constraints and API contract tests.

This prevents a formal specification exercise from becoming a bureaucratic requirement for every line of code.

# 3. How I would apply Phys-Pseudo to a real Singapore business rule

Let's use the overtime eligibility and calculation problem.

This is an illustrative specification, not an approved statutory implementation. Actual parameters and applicability conditions must come from the verified MOM source register.

### Rule: SG-MOM-OT-001 — Part IV overtime calculation

Given

- Employee classification.
- Monthly basic salary.
- Employment arrangement.
- Normal contractual working hours.
- Additional hours worked.
- Whether the work was requested by the employer.
- Applicable pay period.
- Effective statutory rule version.

Define

- Employee coverage.
- Statutory basic-rate calculation base.
- Applicable overtime rate.
- Qualifying overtime hours.
- Applicable calculation limits.

Assume

- The employee belongs to a supported employee category.
- The employment arrangement is covered by the implemented rule pack.
- Required attendance and approval information is available.

Claim

A statutory overtime calculation is valid only when employee coverage and all required conditions have been established.

Derive

The calculation base must be determined from the applicable employee classification and statutory conditions before computing the hourly rate.

If/then/else

```

given employee_classification
given monthly_basic_salary
given overtime_hours
given employment_arrangement
given employer_requested_work
given effective_rule_version

define supported_part_iv_population
define applicable_basic_rate_cap
define hourly_basic_rate
define statutory_overtime_amount

assume monthly_rated_employee
assume supported_employment_arrangement

claim overtime_calculation_requires_verified_eligibility

if employee_is_not_covered then
    claim status = not_applicable
else
    if required_facts_are_missing then
        ?
    else
        derive statutory_basic_rate_base
        derive hourly_basic_rate
        derive qualifying_overtime_hours
        derive statutory_overtime_amount
        claim status = calculated
    end
end

```

The important part is not the pseudocode itself.

It is that an unknown employee classification cannot silently become an assumed eligible employee, and an unsupported employment arrangement cannot silently receive a guessed calculation.

The `?` construct becomes especially valuable here.

We should prefer an explicit indeterminate result over a plausible but potentially unlawful payroll calculation.

# 4. Connect Phys-Pseudo directly to TDD

This is where I see the biggest opportunity.

I would establish a formal sequence:

## Business Rule → Verified Software

1. Authoritative source

   Locate the relevant MOM legislation and official guidance. Record the exact provision and effective date.
2. Phys-Pseudo specification

   Capture facts, definitions, assumptions, claims, derivation and unresolved cases.
3. Decision table and test cases

   Translate the formal specification into executable examples, including boundaries and exceptions.
4. Independent expected-results calculation

   Derive expected outputs separately from the PayLoop implementation.
5. Red-green-refactor TDD

   Write failing tests, implement the Go calculation, make tests pass, and refactor without changing the rule's meaning.
6. Adversarial verification

   Challenge the assumptions, construct counterexamples and verify that unsupported cases are rejected.
7. Version and preserve

   Pin the source, rule specification, test fixtures and implementation versions so historical results can be reproduced.

There is one particularly important safeguard:

Do not generate all golden tests mechanically from the same Phys-Pseudo specification and then claim independent verification.

That would risk reproducing the same misunderstanding three times.

Instead, use distinct routes:

- The specification author formalizes the source.
- The test designer constructs boundary cases directly from the authoritative source and worked examples.
- The oracle developer independently derives the expected results.
- The PayLoop developer implements the production calculation.
- An independent reviewer challenges the result.

Phys-Pseudo becomes the shared contract against which these different efforts can be compared.

# 5. I would introduce a Business Rule Contract

Rather than putting everything into one Phys-Pseudo file, I would create a small, consistent rule package.

For example:

```
modules/sg-wfm-payroll/rules/
└── SG-MOM-OT-001/
    ├── source.md
    ├── rule.phys
    ├── assumptions.md
    ├── decision-table.yaml
    ├── edge-cases.md
    ├── golden-tests.json
    ├── implementation-contract.md
    └── review.md
```

The files have different responsibilities.

| Artifact                     | Purpose                                           |
| ---------------------------- | ------------------------------------------------- |
| `source.md`                  | Pin official legal sources and relevant sections  |
| `rule.phys`                  | Formal business-rule statement                    |
| `assumptions.md`             | State what is accepted, conditional or unresolved |
| `decision-table.yaml`        | Explicit applicability and branching conditions   |
| `edge-cases.md`              | Boundary, exceptional and unsupported cases       |
| `golden-tests.json`          | Executable expected results and provenance        |
| `implementation-contract.md` | Input/output contract for Go                      |
| `review.md`                  | Independent verification and outstanding issues   |

I would not require all eight files for every simple rule. They are appropriate for high-risk statutory rules.

For a simple integration mapping, a smaller contract may suffice.

## 6. Apply your Three Mathematical Reality Checks as well

I would carry the spirit of your existing MRC into this work, but adapt it to business systems.

RC1 — Business-rule landscape

What are the relevant authorities and business structures?

For overtime, that means distinguishing legislation, employment terms, employer configuration and application behavior.

We must identify applicability, exclusions, assumptions, validity periods and limitations.

RC2 — Business-rule intersection

Which conditions must interact before a result can be computed?

For example:

\\[ \text{Employee classification} \cap \text{Employment terms} \cap \text{Attendance} \cap \text{Applicable rule version} \\]

A payroll result is only defensible when those inputs are compatible and correctly interpreted.

RC3 — Business-system interoperability

How does the information move between systems?

For example:

\\[ \text{ShiftHub timesheet} \xrightarrow{\text{mapping}} \text{PayLoop payroll input} \\]

We must record what fields are preserved, transformed, lost or defaulted, what assumptions the mapping makes, and how to verify the transformation.

This is particularly useful because many of our most valuable training incidents occur at the intersections between otherwise-correct systems.

# 7. What this could become for WorkforceOps Lab

Here is where I think the approach becomes commercially interesting.

Eventually, users should be able to inspect the business rules behind a simulated environment.

For example, a candidate investigating missing overtime could see:

- The documented statutory rule.
- The applicable employment category.
- The effective rule version.
- The required business inputs.
- The expected processing path.
- The relevant assumptions and limitations.

In practice mode, the AI mentor could explain the rule and guide the candidate through its conditions.

In assessment mode, those explanations remain hidden. The candidate must investigate independently.

The evaluator can assess not only whether the person found the software defect, but whether they understood the business rule that made the outcome incorrect.

This is substantially better than SQL exercises detached from business context.

It also gives future customer modules a systematic way to capture their own proprietary business logic. A client could provide an internal policy, and we could formalize its conditions, identify ambiguities and create testable examples.

That could become a paid business-rule formalization service.

# 8. My recommended decision

I would formally adopt this approach, with one caveat: Phys-Pseudo is a specification layer, not an executable programming language or a substitute for statutory interpretation.

Do not build a Phys-Pseudo interpreter or code generator in the MVP.

Use it to establish the intended semantics, then translate the specification into Go implementations and executable tests.

I would give the coding agent this instruction:

## WorkforceOps Lab — Phys-Pseudo Business Rule Formalization Policy

Apply the existing seven-construct Phys-Pseudo methodology to high-risk business rules before implementing their Go calculations.

Prioritize statutory applicability, payroll calculations, employee classification, effective-dated rules, monetary rounding and other financially or legally consequential logic.

For each such rule:

1. Pin the authoritative source, applicable version and effective dates.
2. Specify all given facts, definitions, assumptions, claims and derivations using Phys-Pseudo.
3. Identify exclusions, edge cases, validity conditions and unresolved questions.
4. Use `?` for unsupported or insufficiently specified conditions. Never silently assume an answer.
5. Translate the specification into decision tables, golden tests, boundary tests and negative tests.
6. Develop the independent reference oracle separately from the application calculation.
7. Implement the Go calculation using test-driven development.
8. Independently review the derivation and implementation.
9. Preserve source, specification, test and implementation provenance.

For complex operational workflows and integrations, use state machines, event contracts and invariants, supplemented by Phys-Pseudo where the business decisions require formalization.

For routine CRUD, schema validation and straightforward API behavior, use conventional tests without mandatory Phys-Pseudo.

Do not build an interpreter, code generator or general-purpose business-rule engine in the MVP.

The specification must distinguish statutory requirements, employment contract terms, employer configuration and software implementation behavior.

An unresolved load-bearing assumption blocks promotion of the affected calculation. Independent development and testing of unrelated rules may continue.

Acceptance requires traceability from authoritative source to formal specification, tests, independent expected result, Go implementation and reproducible verification.

My conclusion: this is not overengineering. Applied selectively, it is a disciplined way to prevent business-rule hallucination, incorrect assumptions and silent calculation errors.

I would make it part of the WorkforceOps Lab engineering methodology from the beginning. The key is not to formalize everything; it is to formalize the rules whose incorrect interpretation would cause the greatest damage.

It also gives you something valuable that many training platforms lack: a visible, inspectable connection between the rules governing a business, the code implementing those rules and the evidence proving that the implementation behaves correctly.