# prompt5 closeout — verification, drift finding, prompt6 handoff

Date: 2026-11-10. Mode: read-only verification + this note. No retrieval
performed, no package/doc/code edits, no prompt6 work executed.

## 1. Verdict

prompt5 (S3 provisional retrieval + integration amendments) is VERIFIED
COMPLETE as reported in `prompt5_report.md`, with one material drift finding:
the tree is at package v0.3.0, not the reported v0.2.0, and v0.3.0 already
contains most of the `prompt5_imp_review.md` correction pass (§3 below).
`SG-MOM-OT-001` remains BLOCKED, correctly.

## 2. Report-claim verification (report §1–§8 vs tree)

| § | Claim | Verdict | Evidence |
|---|---|---|---|
| 1 | 4 integration amendments applied | PASS with note | AGENTS.md `Statutory rule workflow` present; scope doc lifecycle/activation language present; domain-model §4 activation annotation present; db-design §10 has the substance (`unverified packs seed nothing activatable`) though the exact `citation rule` wording was not located — wording variance, not a gap |
| 1 | 6 artifacts in `sources/provisional/` + CAPTURES.md | PASS | All present: P1–P4 transcripts, P5 PDF, P6 403 record |
| 2 | Digests + hash scope | PASS (6/6) | Independently recomputed below; all match manifest + CAPTURES + report; scope = captured bytes only |
| 3 | Limitations visible | PASS | `provisional_transcript` grades; P6 403 recorded; SSO uncaptured stated |
| 4 | Threshold narrowed; cap reframed | PASS | Corroborated by current SPEC givens + open `?`s (lines 13, 63) |
| 5 | Current vs historical distinguished | PASS | SPEC keeps PDF values as historical givens, MOM values provisional |
| 6 | Package changes, fixtures empty, BLOCKED | PASS | Fixtures = README only; manifest `BLOCKED`, owner empty; contract unchanged |
| 7 | `?`s retained, blockers open | PASS | 14 `?` markers in SPEC, all open (legislation-text needs); no `?` removed |
| 8 | Byte-grade evidence specified | PASS (spec) | SSO Part IV/s38/Fourth Schedule/amendment history listed; still uncaptured |

Recomputed digests (2026-11-10, `sha256sum` over saved bytes — match manifest byte-for-byte):

| Artifact | Bytes | SHA-256 |
|---|---|---|
| P1 | 15017 | 2048e745…23d0bd6 |
| P2 | 8664 | bddfd246…55a6ed |
| P3 | 4162 | 1802efcb…daede |
| P4 | 5841 | adb2bb9a…ed227 |
| P5 | 1041950 | bac2d0ad…fdaf8 |
| P6 | 249 | 967190fb…92d11 |

Manifest S1 `9ffbbc59…55417` also matches the reviewer's independently
computed research-PDF digest. Consistent with the review's 4/4 transcript check.

## 3. Drift finding: v0.2.0 → v0.3.0 (single squashed commit 304fec8)

The report was written at v0.2.0; the committed tree is v0.3.0 and already
contains the imp_review correction pass. Verified present in tree:

- P0 72h-vs-payable: SPEC has `actual_preserved` + `limit_72_check` (no
  `min(H,72)` payable formula anywhere).
- P0 cap branches: SPEC has R1/R2 candidate derivations + the exact
  rounding-rule `?` (no `mech_monthly/mech_hourly` if/else).
- P0 1.5 minimum: `ot_mult` "no less than", `ot_hourly_min`, `mvp_rate`
  restricted-profile assumption, `pot_covered` derivation.
- P0 eligibility: `derive eligibility` chain + conditional `?`.
- P1 emergency split, employer-request `?`, basic-salary exclusions, rest-day
  scope `?`: all present as separate givens/findings.
- S3 register fix ("identified, accessed provisionally"): applied.
- Cross-doc: clock uniqueness aligned (both docs
  `UNIQUE(employer_id, source, source_event_id)`); approval wording aligned
  to all-or-nothing whole-run; P5 regraded `provisional_source_capture`.
- Shift dependency: manifest lists `SG-MOM-SHIFT (package planned)`.

Classification: legitimate follow-up folded before commit, NOT scope leak —
but it means the prompt6 handoff below is smaller than `prompt5_pending.md`
implies. No semantic change was made by THIS closeout pass.

## 4. Stop conditions (prompt5 §6 — all held)

1. BLOCKED: manifest `status: BLOCKED` + `TestOT001RemainsBlocked`
   (`admin_test.go:249`, asserts manifest + zero non-BLOCKED activations).
2. No byte-grade provenance claimed: all grades provisional; manifest
   `blocked_reason` states nothing pinned.
3. No Go overtime calc: `SG-MOM-OT-001` appears in Go only as string IDs in
   3 test files (pack-lock/activation test SQL + the BLOCKED guard test).
4. No golden fixtures: `fixtures/` holds README only.
5. No `?` removed: 14 markers, all open (see §2).
6. No upstream/frozen-architecture changes: no `migrations/` file mentions
   provisional captures; OT migration refs are pre-existing governance
   scaffolding (`00003` params tables, `00012` seed row + grants).

## 5. Prompt6 handoff (remaining only — NOT executed here)

1. P6 block labeling: confirm the 403 record is distinguished from source
   artifacts (grade line still reads `provisional_transcript`).
2. DB-design wording check: `programme_required_skill`/`shift_segment`
   employer-scope description not found as review-quoted; reconcile against
   domain-model (which is explicit) or confirm already consistent.
3. Source-identifier mapping: claims use S1/S2 only (no S4 found — review
   quote not reproducible); confirm S1/S2/P1–P6 mapping unambiguous.
4. Genuine independent derivation (reviewer was review, not derivation).
5. SSO legislation capture (Part IV, s38, Fourth Schedule incl. Item 3,
   threshold-amendment provision + commencement) — still the decisive
   missing evidence; Option 2 (user snapshots) if network stays impaired.
6. Re-review of the already-applied P0/P1 corrections (they landed without
   a recorded second review pass).

## 6. Provenance of this note

Evidence-only pass. Commands used: `sha256sum`, `wc`, `grep`, `git show`,
`git log`, file reads. The working tree is otherwise untouched by this task;
`git status` must show exactly this one new file from it.
