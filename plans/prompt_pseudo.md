The key is to make Phys-Pseudo the formal specification language for business rules, not merely a documentation convention. Every important rule should make its inputs, assumptions, applicability, claims, derivations, conditional branches and unresolved questions explicit before anyone implements it in Go.

Here is the coding-agent prompt.

# Task: Inspect and Adapt Phys-Pseudo as the Business Rule Specification Language for WorkforceOps Lab

## Objective

Read the actual Phys-Pseudo codebase on the local machine, understand its authoritative specification and conventions, and adapt it as the formal specification layer and language for WorkforceOps Lab.

We are building an integrated Singapore HR, Workforce Management and Payroll backend. Statutory and financial business rules must not be implemented merely because an AI agent or developer believes they understand the requirement.

We need a disciplined way to capture:

- What the authoritative source actually says.
- Which employees, employers, circumstances and dates the rule applies to.
- What facts and inputs are required.
- What assumptions are being made.
- What is explicitly defined.
- What claims need to be established.
- How conclusions follow from the facts, assumptions and rules.
- What happens under different conditions.
- What edge cases and boundary conditions exist.
- What remains unresolved or requires independent verification.

Phys-Pseudo must provide the formal language for expressing this reasoning.

The goal is to ensure that a Go implementation can be traced back to an explicit, independently reviewed business-rule specification rather than relying on undocumented assumptions.

## 1. Inspect the actual local Phys-Pseudo codebase first

Expected local repository:

`/home/chaschel/Documents/go/phys-pseudo`

Do not assume that the repository contents match previously discussed descriptions.

Before making changes:

1. Confirm that the repository exists.
2. Inspect its Git status, current branch, upstream remote and current commit SHA.
3. Read `PSEUDO.md` completely.
4. Read the repository README and all relevant documentation.
5. Inspect examples, fixtures, tests, grammar definitions and any parser or validation implementation that exists.
6. Inspect the license and attribution requirements.
7. Identify which files are authoritative and which are explanatory examples.
8. Identify the exact syntax and semantics of every supported language construct.

Do not inspect only the README and assume you understand the language.

**The actual local `PSEUDO.md` is authoritative for the language specification.** Existing examples and implementation tests must be used to understand its intended usage and constraints.

The previously agreed seven-construct vocabulary is:

- `given`
- `define`
- `assume`
- `claim`
- `derive`
- `if/then/else`
- `?`

Verify this vocabulary and its exact syntax against the actual repository. Do not invent syntax, silently change semantics or add new language constructs.

If the local repository contradicts our previous understanding, document the discrepancy and treat the actual authoritative source as the starting point. Do not silently reconcile the two.

Do not modify the upstream Phys-Pseudo repository.

## 2. Establish how Phys-Pseudo fits into WorkforceOps

Treat Phys-Pseudo as our **business-rule specification language**, adapted from its original purpose without changing its grammar or semantics.

The distinction is:

- Phys-Pseudo defines how we express formal specifications.
- Our business-rule methodology defines how we research, verify and approve those specifications.
- Go implements approved specifications.
- Independent fixtures and tests verify the implementation.

Phys-Pseudo is not itself the source of legal authority.

For WorkforceOps, the authoritative source remains the applicable Singapore legislation, government guidance and other recognized sources for the specific rule being formalized.

A Phys-Pseudo specification must never elevate an assumption, interpretation or generated conclusion into an established legal fact.

### The required reasoning structure

For every important rule, the specification process must explicitly distinguish:

**Given**

Facts and inputs supplied by an authoritative source or by a particular business scenario.

**Define**

Terminology, quantities, variables, classifications and derived expressions needed to state the rule unambiguously.

**Assume**

Conditions temporarily accepted for the scope of a derivation or scenario. Every load-bearing assumption must have an explanation, provenance or rationale, and an indication of what conclusions depend on it.

An assumption is not equivalent to a verified legal fact.

**Claim**

The proposition or business-rule result that needs to be established.

**Derive**

The reasoning that connects the given facts, definitions and assumptions to the claim.

Derivations must be independently reviewable. Do not simply restate the expected answer.

**If/then/else**

Express conditional applicability and alternative outcomes explicitly. Include both the conditions under which a rule applies and the consequences when it does not apply.

**?**

Use the exact unresolved-question or indeterminate mechanism supported by the actual Phys-Pseudo specification. Do not invent alternative syntax to represent uncertainty.

An unresolved load-bearing question must remain visible. It must not be converted into a convenient default value.

## 3. Adapt the specification methodology to business rules

We previously agreed on a research-to-implementation sequence. Preserve it:

`Acquire authoritative source → Pin source → Extract claims → Formalize → Independently derive → Generate golden/anti-golden/boundary fixtures → Independent review → Implement in Go → Verify against fixtures`

The statuses remain:

`DISCOVERED → SOURCES_PINNED → CLAIMS_EXTRACTED → FORMALIZED → INDEPENDENTLY_CHECKED → GOLDENS_VERIFIED → IMPLEMENTABLE`

Only a rule package that has reached `IMPLEMENTABLE` may be activated for its corresponding production calculation.

An AI-generated specification is not independently verified merely because the AI says it is correct.

### Source provenance

Every formalized rule must connect to its source evidence.

Record, where applicable:

- Source identifier.
- Official issuing authority.
- Document title and source URL.
- Source artifact or snapshot.
- SHA-256 digest.
- Retrieval timestamp.
- Publication and effective dates.
- Relevant section, paragraph, table, formula or page.
- Applicable jurisdiction.
- Extraction or transcription method.
- Any redistribution or licensing restrictions.

A search-result snippet is not sufficient evidence for a legal claim.

Keep source evidence separate from extracted claims, and extracted claims separate from formalized interpretations.

### Applicability and scope

Every rule must explicitly describe its applicability boundaries.

Consider:

- Jurisdiction.
- Employer and employee categories.
- Employment arrangement.
- Wage basis and salary period.
- Effective date.
- Applicable work arrangement.
- Eligibility conditions.
- Exemptions and exclusions.
- Relevant calendar period.
- Required input availability.
- Unsupported or indeterminate situations.

Do not infer legal classifications from job titles or convenient database defaults.

A rule that applies to one employee category must not silently be applied to another.

## 4. Capture edge cases systematically

An important purpose of adopting Phys-Pseudo is to prevent developers from formalizing only the happy path.

For every L1 rule, require a deliberate edge-case review.

Select relevant categories from the following list rather than blindly applying every category to every rule:

- Lower and upper boundaries.
- Values immediately below and above a threshold.
- Exact threshold values.
- Minimum and maximum caps.
- Eligibility and ineligibility.
- Exemptions and exceptions.
- Missing, ambiguous or contradictory inputs.
- Zero values and negative values where relevant.
- Effective-date transitions.
- Retroactive employment-term changes.
- Calendar-month boundaries.
- Pay-period boundaries.
- Cross-midnight shifts.
- Timezone and daylight-saving transitions where applicable.
- Duplicate or repeated source events.
- Overlapping employment or leave records.
- Rounding and precision boundaries.
- Partial periods.
- Multiple concurrent conditions.
- Conflicting source provisions.
- Unsupported employee or employment categories.

For each relevant edge case, explicitly identify:

1. Its triggering condition.
2. The expected outcome if established.
3. The source or derivation supporting that outcome.
4. Whether it is covered by the current specification.
5. Whether it requires a test fixture.
6. Whether it remains unresolved.

Distinguish among:

- A valid zero result.
- A valid absence of an obligation.
- An inapplicable rule.
- Missing information.
- An unresolved interpretation.
- An unsupported scenario.

Never collapse these into a single zero, false or null result merely for implementation convenience.

## 5. Define three specification rigor levels

Do not force every piece of software into a formal rule package.

Use the previously agreed classification:

### L1 — Statutory, financial and high-consequence business rules

Require a complete Business Rule Research Package, including:

- Source evidence and provenance.
- Extracted claims and applicability.
- Phys-Pseudo specification.
- Explicit assumptions and unresolved questions.
- Independent derivation.
- Golden, anti-golden and boundary fixtures.
- Independent review.
- Implementation contract.

Examples include:

- Ordinary overtime eligibility and calculation.
- CPF Ordinary Wage and Additional Wage classification.
- CPF contribution calculations.
- Skills Development Levy.
- Self-Help Group deductions.
- Leave entitlements and deductions.
- Public-holiday pay.
- Statutory record requirements.

### L2 — Complex operational workflows and integrations

Require explicit state transitions, conditions, invariants and exception paths.

Use Phys-Pseudo selectively where a calculation, eligibility decision or critical invariant benefits from formal specification.

Examples include:

- Roster publication and revision.
- Attendance correction.
- Timesheet approval.
- Payroll approval and exclusion.
- Employee-status synchronization.
- WFM-to-payroll attribution.

### L3 — Ordinary CRUD and technical infrastructure

Use conventional Go validation and tests.

Do not introduce Phys-Pseudo specifications for ordinary CRUD, basic configuration, migration metadata or infrastructure without a genuine business-rule reason.

The purpose is rigor where it matters, not paperwork everywhere.

## 6. Establish the WorkforceOps rule package format

After understanding the actual language, propose a minimal reusable directory structure for a business-rule package.

It should provide clear separation between:

- Manifest and identity.
- Immutable source evidence.
- Extracted claims.
- Phys-Pseudo specification.
- Independent derivation.
- Golden, anti-golden and boundary fixtures.
- Verification and review records.
- Implementation contract.

The manifest should capture at least:

- Stable rule ID.
- Name and description.
- Jurisdiction and responsible authority.
- Rigor level.
- Version.
- Status.
- Effective dates.
- Applicability profile.
- Source references.
- Dependencies.
- Specification files.
- Reviewer and verification status.
- Implementation ownership and version once implemented.

Prefer existing Phys-Pseudo conventions wherever possible. Do not invent a competing specification structure merely to make the business-rule packages look different.

For the specification file itself, use the exact syntax and file conventions established by the inspected Phys-Pseudo codebase.

The result must be easy for both a human reviewer and an AI coding agent to understand.

## 7. Apply the methodology to one initial rule package

Use `SG-MOM-OT-001` — ordinary overtime for the restricted monthly-rated employee profile — as the proposed first L1 exemplar.

Do not resolve the existing S$2,600 cap versus S$13.60/hour discrepancy by choosing one value.

The exemplar must show how the adapted language and package structure handle:

- Source-backed facts.
- Employee eligibility.
- Applicable employment arrangements.
- Required salary and working-hour inputs.
- Formula definitions.
- Conditions and exclusions.
- Boundary cases.
- Rounding requirements.
- Unresolved source conflicts.
- Expected outputs and verification evidence.

Use only evidence already present in the project or verified directly from the pinned authoritative source material.

If a fact is not verified, represent it as unresolved using the actual Phys-Pseudo conventions.

The exemplar must remain `BLOCKED` or otherwise non-implementable while any load-bearing legal or mathematical discrepancy remains unresolved.

Do not produce a specification that looks complete by hiding unresolved questions in prose.

## 8. Vendor the authoritative specification responsibly

If the license permits redistribution:

- Vendor the minimum authoritative Phys-Pseudo specification files and necessary examples into WorkforceOps.
- Preserve their original contents and attribution.
- Record upstream repository URL, pinned commit SHA, license, copied file paths, retrieval date and file hashes.
- Keep adapted WorkforceOps guidance separate from the original specification.
- Make it possible to identify exactly which upstream version our local specification is based on.

Do not copy an interpreter, compiler, code generator or unrelated implementation merely because it exists in the upstream repository.

Do not modify the upstream source to make it fit WorkforceOps.

If licensing or provenance is unclear, do not copy the files until that issue is resolved. Document the blocker.

## 9. Explicit implementation boundaries

This task is about inspecting, adapting and establishing the specification layer.

Do NOT:

- Implement a Phys-Pseudo interpreter.
- Implement a compiler or code generator.
- Generate Go code from specifications.
- Build a generic rules engine.
- Introduce a new programming language or additional language constructs.
- Implement payroll calculations.
- Implement statutory calculation services.
- Introduce a physics-library dependency into WorkforceOps.
- Reopen the frozen single-application architecture.
- Introduce microservices, event buses or a web UI.
- Modify the upstream Phys-Pseudo repository.
- Claim legal compliance because a rule has been written in Phys-Pseudo.

Phys-Pseudo specifications are reviewed engineering artifacts. Go code remains the executable implementation.

Keep the current WorkforceOps architecture and plan4.3 scope intact. This task must not silently expand the database-correction work into business-service implementation.

## 10. Deliverables

Produce the following:

1. **Phys-Pseudo local codebase assessment**
   - Repository and commit examined.
   - Authoritative files.
   - Actual language constructs and semantics.
   - Grammar and formatting conventions.
   - Existing examples and validation facilities.
   - License and attribution findings.
   - Any discrepancies with previous project assumptions.
2. **WorkforceOps Phys-Pseudo adaptation specification**
   - The role of Phys-Pseudo in the project.
   - How each language construct is used in business-rule formalization.
   - Treatment of facts, definitions, assumptions, claims, derivations, conditional branches and unresolved questions.
   - Applicability and edge-case requirements.
   - Research and verification lifecycle.
   - L1/L2/L3 rigor levels.
   - Relationship between specifications, Go implementation and tests.
3. **Business-rule package template**
   - Minimal reusable directory structure.
   - Manifest schema.
   - Source-evidence format.
   - Claim ledger format.
   - Specification file conventions.
   - Derivation and fixture organization.
   - Independent review record.
4. **`SG-MOM-OT-001` exemplar or scaffold**
   - Use the actual Phys-Pseudo syntax.
   - Identify verified facts and unresolved issues separately.
   - Do not claim `IMPLEMENTABLE` while the known discrepancy remains open.
5. **Vendor provenance record**
   - Only if redistribution is permitted.
   - Include pinned commit and hashes.
6. **Integration recommendations**
   - Identify the exact additions needed to the existing WorkforceOps design documents.
   - Prefer targeted amendments over rewriting existing documents.
   - Identify any impact on `singapore-rule-scope.md`, `domain-model.md`, `database-design.md` and the implementation workflow.
   - Do not modify frozen architecture decisions.

## 11. Acceptance criteria

The task is complete only when:

- The agent has read the actual local Phys-Pseudo specification, not reconstructed it from memory.
- Every adopted language construct conforms to the authoritative specification.
- No new syntax or semantics have been invented.
- The distinction between source facts, assumptions, derivations and unresolved questions is explicit.
- A reviewer can trace a formal claim back to its source evidence.
- Relevant edge cases are systematically identified.
- A rule cannot become implementable merely because its specification is syntactically complete.
- Independent derivation and independent review remain separate from the original formalization.
- The exemplar demonstrates the language without concealing the existing overtime discrepancy.
- The resulting package is sufficiently simple for developers and AI coding agents to use repeatedly.
- No payroll implementation, interpreter, compiler or generic rules engine has been added.

## 12. Final report

Conclude with:

1. What the actual Phys-Pseudo codebase contains.
2. What we can safely adopt unchanged.
3. What WorkforceOps-specific conventions are necessary.
4. The final proposed business-rule package structure.
5. The status of `SG-MOM-OT-001`.
6. Any unresolved design or licensing issues.
7. The precise files created or amended.
8. A recommendation on whether the specification layer is ready to become the required standard for future L1 business-rule implementation.

Be adversarial about ambiguity. Do not invent legal facts, language semantics or evidence. The objective is a small, rigorous specification layer that prevents incorrect business rules from reaching production code.

One architectural principle to preserve: Phys-Pseudo should make uncertainty visible, not make uncertainty disappear behind formal-looking syntax. A beautifully written specification with an unverified assumption must still be blocked from implementation.&#x20;