# prompt_pseudo final report

Task §12 — the eight items.

## 1. What the actual Phys-Pseudo codebase contains

52 files at `/home/chaschel/Documents/go/phys-pseudo`, `main @ b9bb9a4`
(clean, untouched): the canonical language spec (`PSEUDO.md`, 397 lines,
seven constructs, representation-≠-endorsement invariant), the MRC review
protocol, a 10-fixture conformance corpus (documents preserved verbatim
including deliberate defects), PhysTrace implementation-handoff guidelines,
agent/human entry docs, and `plan/` design/review history. **No parser,
checker, test suite, or executable of any kind.** Full findings in
`docs/workforceops/rules/pseudo-assessment.md`.

## 2. What we can safely adopt unchanged

The seven constructs with exact syntax and semantics; the `?` attachment
rules; the no-status invariant; prose-only attribution; the MRC protocol's
eight responsibilities and evidential-boundary close. Nothing in the
language needed modification — only a statement of what each construct
carries for payroll rules (`adaptation-spec.md`).

## 3. What WorkforceOps-specific conventions were necessary

Per-construct business-rule usage (notably: uncertain legal readings are
`claim` + `?`, never `assume`); the applicability block; the edge-case
review with six distinct outcomes; the research lifecycle with manifest
(not spec) statuses wired to `rule_pack_activation`; L1/L2/L3 rigor
levels; the spec↔Go↔fixture contract. Plus one evidence rule the PDF
forced: rendered-page reads for formulas, because text extraction drops
every value.

## 4. Final package structure

```text
rules/<RULE-ID>/manifest.json sources/SOURCES.md claims.md SPEC.pseudo.md
  derivation.md edge-cases.md fixtures/{golden,anti-golden,boundary}/
  REVIEW.md contract.md
```

Template at `rules/_template/`; upstream verbatims + provenance at
`docs/workforceops/rules/upstream/`; method docs at
`docs/workforceops/rules/`.

## 5. Status of SG-MOM-OT-001

`BLOCKED` (manifest), honestly. The SPEC formalizes the PDF's overtime
shape with real arithmetic, two contested thresholds (S$2,500 vs S$2,600),
two unendorsed cap mechanisms (S$20.45 vs S$20.40), and three
document-level `?`s. No MOM source is pinned; the derivation is marked
same-author/non-independent; fixtures are empty by design; review is
pending. It demonstrates the language without concealing anything.

## 6. Unresolved design or licensing issues

Licensing: none (MIT, same holder, notice vendored). Design: the MOM
retrieval pass (S3) is the critical path for every L1 package; no
automation beyond a `?`-sweep should be built without a reviewed tooling
milestone, per the upstream no-tooling stance.

## 7. Files created or amended

Created: `docs/workforceops/rules/{pseudo-assessment,adaptation-spec,integration-recommendations}.md`,
`docs/workforceops/rules/upstream/{PROVENANCE.md,PSEUDO.md,MRC_REVIEW_PROTOCOL.md,LICENSE,conformance/{README.md,01_operator_composition/fixture.md,04_conditional/fixture.md}}`,
`rules/_template/{README.md,manifest.json,SPEC.pseudo.md,claims.md,edge-cases.md,derivation.md,REVIEW.md,contract.md,sources/,fixtures/{golden,anti-golden,boundary}/}`,
`rules/SG-MOM-OT-001/{manifest.json,SPEC.pseudo.md,claims.md,derivation.md,edge-cases.md,REVIEW.md,contract.md,fixtures/README.md,sources/SOURCES.md}`,
this report. Amended: none. Upstream: untouched (verified clean after
copying). No Go, no migrations, no frozen-arch changes.

## 8. Recommendation

**Yes — adopt as the required standard for future L1 rules**, with two
conditions: (a) apply the `integration-recommendations.md` paragraph
amendments so the design docs point at the packages; (b) run the MOM
retrieval pass before any package advances past `SOURCES_PINNED`. The
layer is small, mechanically checkable where it counts (no status words
in specs, `?` attachment, byte-identical vendors — all swept in
execution), and it has already proven its value: the exemplar surfaced a
real threshold conflict and a S$0.05/hour mechanism ambiguity that a
prose spec would have smoothed over.
