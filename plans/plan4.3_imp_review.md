I reviewed both reports. This is a successful milestone. I recommend closing the design-correction pass and moving to the MOM source-retrieval pass (S3).

The two tracks have passed different kinds of gates, and it is important not to confuse them.

## 1. Track A — WorkforceOps design gate

PASS

The evidence documented in `plan4.3_gate.md` is strong:

- Actual PostgreSQL 16.15, using the real seven-version baseline migration covering 50 tables.
- Integrity tests executed as the unprivileged `app` role in a separate database.
- All Amendments A–D passed.
- Draft mutations work; published and approved records reject UPDATE and DELETE.
- Employer isolation, including same-employer employee-to-roster consistency, is tested.
- Payroll exclusion and all-or-nothing approval behavior are covered.
- Twenty integrity tests passed, with one explicitly documented deferred test.
- Seven Go packages passed across three consecutive runs and a fresh-install run.

Particularly important: the tests caught five real bugs, including triggers silently discarding permitted updates and the test suite connecting to the wrong database under the wrong role. These were corrected and retested.&#x20;

plan4.3_gate.md

plan4.3_gate.md



I accept the report's recommendation:

Close the design-correction pass. Do not reopen the frozen architecture.

Two boundaries remain correctly deferred:

- End-to-end payroll calculation reproducibility awaits the payroll implementation.
- Statutory calculations remain blocked until their respective rule packages reach `IMPLEMENTABLE`.

The report makes no unsupported claim that these have passed.&#x20;

plan4.3_gate.md



I have reviewed the submitted evidence and reports; I have not independently rerun the agent's PostgreSQL tests in this session.

## 2. Track B — Phys-Pseudo specification layer

ESTABLISHED

This is equally important, perhaps more significant for the long-term quality of WorkforceOps.

The agent inspected the actual local repository rather than reconstructing the language from memory. The assessment establishes:

- Upstream `main` at `b9bb9a4`.
- Seven authoritative constructs.
- `PSEUDO.md` as the sole language authority.
- The representation-versus-endorsement invariant.
- No parser, checker or executable implementation.
- MIT licensing permits the reported vendoring.
- Upstream remains untouched.&#x20;

  prompt_pseudo_report.md



The adaptation is appropriately restrained. It does not invent a new language or build a rules engine. It establishes a business-rule package structure, applicability conventions, edge-case reviews, research statuses and the relationship between the specification, Go implementation and fixtures.&#x20;

prompt_pseudo_report.md



The proposed package layout is also sensible:

```
rules/
├── _template/
└── SG-MOM-OT-001/
    ├── manifest.json
    ├── sources/
    ├── claims.md
    ├── SPEC.pseudo.md
    ├── derivation.md
    ├── edge-cases.md
    ├── fixtures/
    │   ├── golden/
    │   ├── anti-golden/
    │   └── boundary/
    ├── REVIEW.md
    └── contract.md
```

Most importantly, the initial exemplar remains BLOCKED. The agent found a S$2,500 versus S$2,600 threshold conflict, a S$20.45 versus S$20.40 cap-mechanism ambiguity, and three unresolved questions. It did not invent a resolution, populate unverified fixtures or claim independent verification.&#x20;

prompt_pseudo_report.md



That is precisely the behavior we wanted from the specification layer.

## 3. One remaining integration task

The Phys-Pseudo report recommends applying the paragraph-level amendments in:

`docs/workforceops/rules/integration-recommendations.md`

These have been proposed but not yet applied to the existing design documents.&#x20;

prompt_pseudo_report.md



I approve those targeted amendments, provided they do not reopen any frozen decisions.

After that, I would establish this as a permanent project rule:

All future L1 statutory and financial business rules must use the WorkforceOps Phys-Pseudo package standard.

L2 uses it selectively; L3 does not require it.

## 4. Next critical path — S3: authoritative MOM retrieval

NEXT

This is now the correct next investment.

The initial research exposed discrepancies, but no direct MOM source has yet been pinned for `SG-MOM-OT-001`. The priority is not to write more specifications. It is to establish what the authoritative sources actually say.

I recommend giving the agent the following instruction.

# S3 — Authoritative Singapore MOM Retrieval for SG-MOM-OT-001

## Objective

Resolve the outstanding source and mathematical uncertainties in the existing `SG-MOM-OT-001` Business Rule Research Package.

Use the adopted Phys-Pseudo specification methodology and the existing WorkforceOps research conventions.

This is a source-retrieval, evidence-verification and derivation task.

It is not a payroll implementation task.

## 1. Inspect existing evidence first

Read:

- `rules/SG-MOM-OT-001/`
- `rules/_template/`
- `docs/workforceops/rules/adaptation-spec.md`
- `docs/workforceops/rules/integration-recommendations.md`
- The existing `SG_research.pdf` source artifact and its relevant rendered pages.
- Existing source, claim, specification, derivation and review records.

Do not discard existing work or overwrite unresolved findings.

Preserve the three unresolved questions and the existing discrepancy records until each is explicitly resolved or remains blocked with an evidence-based explanation.

## 2. Retrieve authoritative sources directly

Retrieve the relevant official Singapore sources from:

- Ministry of Manpower (MOM).
- Singapore Statutes Online.
- Applicable legislation, subsidiary legislation, schedules and amendments.

Do not rely on secondary summaries, search-result snippets, AI-generated explanations or previously extracted PDF text as legal authority.

For every source, record:

- Issuing authority.
- Exact document title.
- Official source URL.
- Publication or amendment date, where available.
- Effective and commencement dates.
- Retrieval timestamp.
- Source artifact and SHA-256 digest.
- Relevant section, subsection, schedule, item, formula or table.
- Historical applicability where rules have changed.

Preserve immutable source snapshots and distinguish current rules from historical rules.

For documents containing formulas or tables, inspect rendered pages. Do not assume text extraction preserved numerical values.

## 3. Resolve the two known discrepancies

### A. S$2,500 versus S$2,600

Determine exactly which historical and current thresholds apply to the relevant employee classification and overtime provisions.

Establish:

- What the official source says.
- When the threshold changed.
- Which provision changed.
- The effective date.
- Whether the historical research document used an obsolete value.
- Which employee population and calculation are affected.

Do not resolve this by selecting the latest number without establishing its legal applicability.

### B. S$20.45 versus S$20.40

Investigate the original cap-mechanism discrepancy.

Establish:

- The exact statutory or official source defining the cap.
- Whether the difference concerns a published hourly limit, a derived monthly-rated hourly rate, rounding, an amendment or another mechanism.
- The applicable formula and inputs.
- The exact rounding stage, where prescribed.
- Whether the two values refer to different legal periods or different calculation contexts.

Show the independent arithmetic explicitly.

Do not force the numbers to agree.

If the official sources do not resolve the discrepancy, preserve it as an unresolved `?` and leave the rule blocked.

## 4. Apply Phys-Pseudo correctly

Use the actual vendored `PSEUDO.md` syntax and semantics.

Update the package only when source evidence justifies the change.

Keep the following distinct:

- Authoritative facts.
- Extracted claims.
- Definitions.
- Assumptions.
- Mathematical derivations.
- Interpretive claims.
- Unresolved questions.

An uncertain legal interpretation must not be disguised as an assumption merely to make the specification complete.

The formal specification must not imply endorsement of a contested interpretation.

## 5. Independent derivation and adversarial review

Have an independent reviewer examine:

- The pinned official sources.
- Extracted statutory claims.
- Applicability conditions.
- Arithmetic and formula derivations.
- Historical effective-date treatment.
- Boundary cases.
- All remaining unresolved questions.

The reviewer must not simply inherit the preferred conclusion of the original researcher.

Produce golden, anti-golden and boundary fixtures only for outcomes sufficiently established by authoritative evidence and independently checked derivations.

Do not fabricate expected values for unresolved cases.

## 6. Update the research package

Update only relevant files within `rules/SG-MOM-OT-001/`.

Maintain stable rule IDs and preserve provenance.

Update:

- `manifest.json`
- `sources/SOURCES.md`
- Claims ledger.
- `SPEC.pseudo.md`
- `derivation.md`
- `edge-cases.md`
- Fixtures, where justified.
- `REVIEW.md`
- `contract.md`

Record the disposition of each original discrepancy and each unresolved question.

## 7. Explicit stop conditions

Do not:

- Implement overtime calculations in Go.
- Change the PostgreSQL schema.
- Add a payroll rules engine.
- Introduce a Phys-Pseudo interpreter or compiler.
- Promote the package based on incomplete evidence.
- Invent statutory interpretations.
- Alter the frozen WorkforceOps architecture.

The package may advance through its research lifecycle only as supported by actual evidence and independent review.

If the threshold or cap issue remains unresolved, retain `BLOCKED` and document precisely what additional authoritative evidence is necessary.

## 8. Final report

Provide:

1. Official sources retrieved and their provenance.
2. Historical/current rule comparison.
3. Resolution or remaining status of both discrepancies.
4. Independent mathematical derivation.
5. Changes to the existing rule package.
6. Golden/anti-golden/boundary fixture disposition.
7. Independent reviewer findings.
8. Final package status.
9. Remaining blockers.

The objective is not to force `SG-MOM-OT-001` to become implementable. The objective is to establish the correct rule, or demonstrate exactly why it cannot yet be established.

## 5. What happens after S3

I recommend the following sequence:

1. Close plan4.3 design correction.

   Preserve the PostgreSQL baseline and passing integrity suite.
2. Apply Phys-Pseudo integration amendments.

   Make the new specification standard discoverable from the WorkforceOps design documents.
3. Complete S3 MOM retrieval.

   Resolve or precisely document the overtime uncertainties.
4. Independently verify SG-MOM-OT-001.

   Promote only if the source, derivation and fixture requirements pass.
5. Begin the first implementation slice.

   Implement only the rule packages that have actually reached `IMPLEMENTABLE`.

My overall assessment: We now have two valuable foundations: a database whose key integrity guarantees have been tested against PostgreSQL, and a formal specification methodology that refuses to hide unresolved business rules.

The next major risk is no longer architectural complexity. It is whether we correctly translate Singapore's authoritative rules into verified, executable business logic. That is exactly what S3 should address.