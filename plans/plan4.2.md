## Goal

Implement `plan4.1_review.md`: apply its 12-item amendment to the SG-research incorporation program and adopt its phys-pseudo business-rule formalization policy (grounded in the actual spec at `/home/chaschel/Documents/go/phys-pseudo`), terminating at the triage-plus-verification dossier gate — no payroll code.

## Success Criteria

- Overtime conflict closed per the amendment: S$2,500/S$2,250 recorded as pre-1-Apr-2019 history, S$2,600 current for coverage and cap, each with source, effective date, and retrieval record from a directly opened current MOM page.
- TC-OT-PART4-002 quarantined permanently with its anomaly documented; a fresh fixture exists only if the calculation *and* rounding treatment verify against official sources.
- INC-001/002/003 specs corrected (ordinary `OT_STANDARD` overtime; same-event-ID/idempotency-key dedup; narrowed DLQ without final-pay proration); SG-WFM-003 explicitly deferred; dual evidence classes applied with engineered faults at E5 and "reproducible synthetic incident" language throughout.
- Employment classification split (Act-excluded vs Part-IV-only-excluded), precise supported overtime subset with `indeterminate` for the rest, bitemporal model, and exact-arithmetic standard all specified.
- Formalization adopted: business-rule profile of the seven constructs, three rigor tiers with a tier register, adapted MRC review, Business Rule Contract packages with a first fully conforming SG-MOM-OT-001 exemplar, and an enforced no-interpreter/no-engine fence.
- Dossier delivered: updated triage, source register, conflict log, scenario mapping, golden-test disposition, roadmap reconciliation, and review records — every unverified item still marked as such.

## Approach

**Authority.** The review's amendment (§5, items 1–12) plus its Formalization Policy (§8) govern; plan4.1's seven steps remain but amended; this supersedes the unamended plan4.1 program. No `plan4.1(4).md` file exists — that's the reviewer's local name for our plan4.1.md.

**Key decisions:**

- D1 — The review's illustrative overtime block is *not* conforming phys-pseudo (uses `end`, bare `?`, `claim status = …`, missing `from` lists and name colons). Execution rewrites it as a strictly conforming exemplar; every `.phys` file is checked against PSEUDO.md mechanically (syntax, declaration-before-use, non-empty premise lists, `?` placement, no smuggled premises via the substitution test, no status labels per §7).
- D2 — Phys-pseudo (MIT, same copyright holder) is vendored as pinned copies with attribution into `docs/workforceops/formalization/`, alongside our business-rule profile (seven constructs verbatim; RC1–RC3 business adaptations; `?` maps to `indeterminate`; status lives in `review.md`/manifest, never in `.phys`).
- D3 — The review's MOM resolution settles the historical-vs-current dilemma, but our register still requires one direct current-MOM-page retrieval (URL/section/hash/date) before encoding — the review's favicon links are not citable retrievals.
- D4 — Rigor tiers become file-requirement tiers (L1 full 8-file package; L2 state machine plus selective `.phys`; L3 conventional tests), tracked in a tier register so rigor can't silently drift.
- D5 — Engineered faults default to failure-mechanism class E5; inheriting a PDF E-label requires checking its cited source; the TC-OT-002 replacement waits for verified rounding, not just the verified cap.

## Steps

1. **Setup.** Verify baseline green; vendor pinned phys-pseudo specs with attribution; write the business-rule profile, rigor tiers plus register, adapted MRC review procedure, and the rule-package template with its contract checklist.
2. **Overtime resolution (items 1–2, 7–8).** Confirm current MOM values directly and record provenance; quarantine the PDF fixture with anomaly notes; derive the fresh fixture only on verified calculation plus rounding; fix the classification split; write the precise supported-subset spec.
3. **First formal specs.** Author SG-MOM-OT-001 (eligibility plus calculation) as a full L1 package in conforming syntax; adversarial-review it under the adapted protocol with an evidential-boundary record; adopt it as the exemplar.
4. **Incident corrections (items 3–5, 11).** Rewrite INC-001/002/003 specs per the amendment; defer SG-WFM-003 explicitly; update the catalogue and backlog.
5. **Evidence discipline (item 6).** Add `rule_evidence_class` and `failure_mechanism_evidence_class` to the scenario format; reclassify merged scenarios (E5 default for engineered faults); sweep language to "reproducible synthetic incident."
6. **Temporal and arithmetic standards (items 9–10).** Specify effective/recorded time plus supersession and recalculation rules; mandate `math/big.Rat` (or exact fixed-point) intermediates with rounding only at verified statutory boundaries; record both in an ADR with a float-ban test over rules packages.
7. **Scope and dossier (item 12, §4 retains).** Enforce the MVP fence (no CPF/SDL/SHG/IR21/leave/rest-day/public-holiday code); file the CPF/SDL official confirmations as later-gate candidates; finalize triage, register, conflict log, scenario mapping, golden-test disposition, and roadmap reconciliation; deliver the dossier report and stop.

## Validation Plan

- Per amendment item: acceptance evidence (retrieval records, quarantine enforcement test proving quarantined fixtures can't load in statutory runners, spec-review records with evidential boundaries, full suite plus lint green).
- Every `.phys` file passes the mechanical conformance checklist and an adapted-MRC review before its package counts complete; any unresolved load-bearing assumption blocks that rule's promotion only.
- Highest-risk check: the direct-MOM confirmation of S$2,600 coverage/cap, 1.5× rate, hourly formula, S$13.60 equivalence, 72-hour limit, and rounding treatment. If any value can't be confirmed on a current official page, it stays `unverified` and dependent specs stay blocked.

## Risks / Open Questions

- **Q1:** Confirm vendoring pinned phys-pseudo copies into our docs (assumed yes — MIT, same copyright holder) versus referencing the sibling repo path, which would break fresh clones.
- **Q2:** Confirm this execution terminates at the dossier gate with no rule code, per review §6 — assumed yes.
- Risks: review-text-as-truth creep (fence: D3 direct retrieval, no exceptions); formalization sprawl into L3 territory (fence: tier register enforced at package creation); stale-PDF value leakage (fence: quarantine test plus conflict log).
