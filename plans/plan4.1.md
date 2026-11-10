## Goal

Incorporate `SG_research.pdf` (27-page Singapore HR/WFM/payroll research: Part IV overtime, CPF/SDL/SHG/IR21 rules, failure taxonomy E1–E5, Tier 2 competency matrix, 12 incident scenarios, schema specs, P0–P3 roadmap, golden-test JSON, adversarial edges) into the WorkforceOps Lab implementation — as rigorously verified input, never as ground truth.

## Success Criteria

- Every PDF value slated for implementation is cross-checked against a current official source (section, hash, retrieval date recorded) and the `rule-source-register.md` reflects verified vs still-unverified status honestly.
- The S$2,500-vs-S$2,600 threshold conflict (plus the PDF's own 2250 anomaly in TC-OT-PART4-002) is resolved from current official sources before any overtime code or golden fixture treats one as truth.
- The 12 scenarios are merged into our catalogue: INC-001–003 enriched with PDF mechanisms, the rest tagged to rule-pack gates in a backlog; E1–E5 evidence classes adopted into the scenario format.
- Domain deltas (bitemporal rules, money/rounding conventions), roadmap reconciliation, and a triaged citation list are documented; nothing deferred (CPF/SDL/SHG/IR21/rest-day) leaks into MVP scope.

## Approach

**Verification before encoding.** I read the full PDF (text extraction plus rendered-page check). Three findings force this stance: (1) key numbers live in formula images invisible to text extraction — values must be hand-transcribed and double-checked; (2) the PDF cites a 2013 bill and a 2015 historical Employment Act snapshot, and its rendered page 1 states a S$2,500 non-workman threshold against our baseline's S$2,600 — likely stale, provable only from current sources; (3) works-cited mixes official pages with vendor blogs and plainly irrelevant PDFs, and its own golden test uses a third figure (2250), so internal consistency cannot be assumed.

**Precedence and conflicts:** frozen product scope wins — the PDF's P0 (full CPF engine in MVP) is rejected in favor of our CPF gate; its microservice framing is ignored (we adopt calculations, scenarios, and taxonomy, not its topology); its single-ledger schema is adapted to our five-schema/workspace design. Money stays `NUMERIC(12,2)` in storage with exact intermediate arithmetic in Go via stdlib `math/big` (no new dependency, no float drift) — the PDF's 4dp rates are computed, not stored.

**Sequencing:** verification and catalogue work run parallel to the pending prompt3 corrections (no dependency — consistent with plan4's parallel-research decision); doc deltas merge before or with the prompt3 consistency audit; no rule *code* until the design gate passes AND per-pack verification completes.

## Steps

1. **Triage.** Classify all 27 works-cited entries (official / third-party / irrelevant); build the content map (rule → pack → phase); open the conflict log (S$2,500/S$2,600/2250 first entry); list values needing visual transcription. Deliverable: `docs/workforceops/sg-research-triage.md`.
2. **Official-source verification.** Per rule: retrieve the current official page, record URL/section/hash/date, confirm or refute the PDF value, derive golden tests including boundaries, and fold the adversarial edges (allowance separation, birth-month bands, mid-month SPR proration, IR21 day-30 directive) into the test backlog. Deliverable: register promotions plus verified rule specs (overtime complete; CPF draft-gated; SDL/SHG/IR21/proration as outline backlog).
3. **Golden fixtures.** Encode only verified cases as JSON fixtures with provenance headers; reproduce the PDF's claimed arithmetic in transcription-accuracy checks (catches copy errors without asserting statutory truth); quarantine TC-OT-002 until its anomaly resolves.
4. **Scenario merge.** Map SG-PAY-001→INC-001, SG-WFM-010→INC-002 (missing-constraint mechanism), SG-INT-008→INC-003; SG-PAY-007 becomes the fourth-incident candidate once the cap rule verifies; remaining seven to `scenarios/backlog.md` tagged by pack, difficulty, and evidence class; add `evidence_class` (E1–E5) to `scenario-format.md`; file the competency matrix as Phase 5 content backlog.
5. **Domain deltas.** Specify bitemporal effective-dating + backpay recalculation rules; validate the PDF's six lineage stages against the R1 sync design; record the money/rounding/`math/big` decision in an ADR; update `domain-model.md` accordingly.
6. **Roadmap reconciliation.** Map PDF P0–P3 onto our phases with gate annotations (what's in, what's gated, what's post-MVP) in the triage doc; confirm rest-day pay stays deferred.
7. **Gate coordination.** Merge doc deltas into the prompt3 consistency audit; report verification coverage and blockers in the design-gate package.

## Validation Plan

- Per-rule checklist: current official source opened, section + hash + date recorded, PDF value confirmed/refuted, golden tests derived with boundaries, independent review noted — any gap keeps the rule `unverified`.
- Fixture checks: JSON schema validity, provenance headers present, transcription arithmetic reproduced, quarantined cases excluded from every runner; `task lint` / `task build` / `go test ./...` stay green throughout (docs + fixtures only, no behavior change).
- Highest-risk check: the threshold-conflict resolution. If current sources cannot be retrieved or conflict further, overtime stays `unverified`, Phase 2 rule work stays blocked, and the verdict says so explicitly.

## Risks / Open Questions

- **Q1:** Is this PDF the completed `prompt2.md` deep-research output? If more research drops are coming, I'll batch verification rather than close it out.
- **Q2 (non-blocking):** Confirm rest-day pay rules stay deferred even though scenario SG-WFM-003 needs them — assumed yes per frozen scope.
- Risks: stale-value contamination (fence: nothing enters fixtures-as-truth without an official cross-check); scope creep via the PDF's P0 (fence: gate annotations enforced at merge); transcription error on formula images (fence: double transcription + arithmetic reproduction checks).
