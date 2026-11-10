# prompt5 final report — S3 provisional retrieval + integration amendments

## 0. Integration amendments (parallel action)

Applied all four `integration-recommendations.md` amendments verbatim, with
one location correction: `rule_pack_activation` is annotated in
domain-model §4 (Rules), not §3, which is where the entity actually lives.
Files: `singapore-rule-scope.md` (§1 spec column, §2 lifecycle paragraph,
§5 applicability sentence), `domain-model.md` (§4 activation annotation),
`database-design.md` (§10 seeding citation rule), `AGENTS.md` (new
Statutory rule workflow section establishing the L1 package standard). No
frozen decisions touched.

## 1. Sources captured

Six artifacts in `rules/SG-MOM-OT-001/sources/provisional/`, full metadata
in `CAPTURES.md`: P1 MOM hours/overtime page, P2 MOM who-is-covered, P3 MOM
FAQ salary limit, P4 MOM calculator page, P5 MOM OT poster PDF (rendered
page inspected — awareness poster, no formulas), P6 SSO Employment Act
attempt (HTTP 403, failure recorded). No curl retries were attempted, per
instruction.

## 2. SHA-256 digests and exact hash scope

| Artifact | SHA-256 | Scope |
|---|---|---|
| P1 (15017 B of 67705 fetched) | 2048e745…23d0bd6 | exact saved transcript bytes |
| P2 (8664 B of 57961 fetched) | bddfd246…39df55a6ed | exact saved transcript bytes |
| P3 (4162 B of 46509 fetched) | 1802efcb…55c115daede | exact saved transcript bytes |
| P4 (5841 B of 52391 fetched) | adb2bb9a…8808ded227 | exact saved transcript bytes |
| P5 (1041950 B) | bac2d0ad…2373b66cfdaf8 | exact saved PDF bytes |
| P6 | 967190fb…6fc7ccc92d11 | exact failure-record bytes |

Digests recomputed independently by the reviewer — all match. Hash scope
is the captured artifact only, never the official source.

## 3. Provisional provenance limitations

Every artifact is `provisional_transcript`. Transcripts are processed
text, not source HTML (markup/scripts absent; P4's calculator JS logic
uncaptured). No MOM page states effective-date history. SSO legislation
text, the Fourth Schedule, and amendment records are entirely uncaptured
(P6 403). Nothing here pins a source; Stage 2 byte-grade verification is
required for every value.

## 4. Findings on both overtime discrepancies

**Threshold ($2,500 vs $2,600):** narrowed, not closed. Four MOM-current
pages (updated 2024–2025) uniformly state non-workman ≤ $2,600, workman ≤
$4,500 — the register's hint is corroborated at provisional grade and the
PDF's $2,500/$2,250 look obsolete. But no effective date and no amendment
text were captured, so historical applicability stays open.

**Cap mechanism ($20.45 vs $20.40):** reframed, not closed. MOM states the
cap as "$2,600 salary level, or $13.60 hourly" with a worked $40.80
example — a single mechanism, not two. The load-bearing question is now
the unstated rounding step: (12×2600)/2288 = 13.6364 while MOM states
13.60 (nearest- and floor-to-$0.10 both fit; stage unstated). Unrounded
2-hour pay is $40.91 vs MOM's $40.80 — $0.11 apart, material.

## 5. Historical and current rule distinctions

Current (provisional): $2,600/$4,500 coverage, $13.60 cap figure, 1.5×,
14-day payment, 72h cap with exemption path, break exclusion.
Historical: PDF values ($2,500/$2,250) presumed pre-amendment — presumed,
not proven. The amendment, its provision, and its effective date are
precisely what SSO text must supply (Option 2 fallback).

## 6. Phys-Pseudo package changes (v0.1.0 → v0.2.0, still BLOCKED)

SOURCES.md +S4; claims.md updated (C2/C3/C10/C11/C13/C14 refined, +C15–C22);
SPEC +6 provisional givens, +1 gap derive with `?`, 3 doc-level `?`s
refined (none removed) + threshold `?` refined; derivation.md +independent
arithmetic section; edge-cases.md +E13/E14; manifest v0.2.0 with P1–P6
digests; REVIEW.md filled by reviewer; contract.md unchanged (no owner);
fixtures still empty by design. SPEC re-swept: no status words, syntax
holds.

## 7. Remaining questions and blockers

All six SPEC `?`s open; contested C9/C13/C14/C22; threshold effective
date; 13.636→13.60 rounding rule + stage; SSO text (Fourth Schedule,
amendment history); genuine independent derivation (reviewer was review,
not derivation). Package correctly stays BLOCKED.

## 8. Exact evidence needed for byte-grade verification

Per prompt5 §5: original HTML bytes of the four MOM pages + original PDF
bytes, each with SHA-256, retrieval method/timestamp, and a
line-by-line comparison against P1–P5 (formulas, tables, footnotes,
Last-Updated stamps). Plus SSO Employment Act Part IV, s38, Fourth
Schedule, and the threshold-amendment provision with commencement date —
via user-provided snapshots (Option 2) or unimpaired network. Only then
may any source count as pinned and the package advance.
