Choose Option 1: Provisional transcripts, with a strict provenance boundary. Do not defer all of S3, but do not pretend that `web_fetch` transcripts are byte-identical copies of official MOM documents.

This gives us useful progress without weakening the evidence standard we established for Phys-Pseudo.

## 1. The provenance decision

Use a two-stage retrieval process:

Stage 1 — Provisional retrieval

Proceed now

Retrieve official MOM and Singapore Statutes Online content using `web_fetch`. Preserve the exact returned payload or transcript, its SHA-256 digest and retrieval metadata.

Stage 2 — Byte-grade verification

Required before promotion

Obtain the original HTML, PDF or other source artifact and independently calculate its SHA-256 digest. Compare it with the provisional extraction and resolve discrepancies.

The distinction is critical:

- SHA-256 of a transcript proves the transcript has not changed since capture.
- It does not prove that the transcript is byte-identical to the official source.
- A transcript may omit tables, formulas, footnotes, formatting or other legally significant details.
- A web citation or source URL alone is not an immutable snapshot.

Therefore, the provisional artifact must be described as a captured representation of an official source, never as the original source bytes.

## 2. Required metadata for every provisional artifact

I would require the agent to preserve:

```
source_id
issuing_authority
document_title
official_url
retrieval_timestamp_utc
retrieval_method
retrieval_tool
document_publication_date
document_effective_date
source_locator
artifact_type
artifact_sha256
hash_scope
provenance_grade
limitations
```

Use `provenance_grade: provisional_transcript` for the present retrieval method.

Also preserve the original returned content without silently cleaning it up. Any normalized text or extracted claims should be separate artifacts with their own digests.

For PDFs, the agent must inspect rendered pages containing formulas, schedules and tables. The previous research demonstrated that text extraction can lose numerical values.

## 3. What this means for SG-MOM-OT-001

Proceed with retrieving the official sources and investigate:

- The S$2,500 versus S$2,600 threshold.
- The S$20.45 versus S$20.40 cap-mechanism discrepancy.
- The applicable formulas, rounding and eligibility provisions.
- Historical amendments and effective dates.

However, retain `SG-MOM-OT-001` as BLOCKED.

Do not advance it to `SOURCES_PINNED` solely on the strength of provisional transcripts. Do not generate supposedly authoritative golden fixtures from unresolved source material.

The existing Phys-Pseudo `?` findings must remain visible.

## 4. Instructions to the agent

# S3 Provenance Decision — Provisional Retrieval with Byte-Grade Verification Gate

Proceed using Option 1: provisional transcripts retrieved through the available `web_fetch` capability.

The sandbox shell currently cannot retrieve external sources using curl, but `web_fetch` can reach official MOM pages. Do not waste time repeatedly attempting unavailable network paths.

## 1. Provisional retrieval

Retrieve official Singapore MOM and Singapore Statutes Online material directly using `web_fetch`.

Preserve the exact returned payload or transcript as an immutable provisional artifact.

For every capture:

- Record the official URL, authority and document title.
- Record the retrieval timestamp in UTC.
- Record the retrieval method and tool.
- Record publication, amendment and effective dates where available.
- Record relevant statutory sections, subsections, schedules, tables and page locators.
- Calculate SHA-256 over the exact bytes or text actually captured.
- State precisely what was hashed.
- Preserve the original captured content without silently editing it.

Do not call a normalized transcript an original source document.

## 2. Provenance classification

Label these artifacts `provisional_transcript`.

A SHA-256 digest of a web_fetch transcript establishes the integrity of that captured transcript only. It does not establish byte identity with the official MOM HTML or PDF.

Keep this limitation visible in the source register and package manifest.

Do not create a new research lifecycle status. Do not promote the overall package to `SOURCES_PINNED` based solely on provisional transcripts.

## 3. Separate source capture from interpretation

Keep three distinct evidence layers:

1. Provisional official-source captures.
2. Extracted statutory claims and interpretations.
3. Phys-Pseudo specifications and independent derivations.

Every extracted claim must identify its source artifact, document locator and relevant text or table.

Where a transcript omits or ambiguously represents a formula, table or footnote, record the limitation instead of reconstructing the missing content.

For PDFs, inspect rendered pages whenever formulas, schedules or numerical tables are involved.

## 4. Investigate the existing overtime blockers

Use the provisional captures to investigate:

- S$2,500 versus S$2,600 threshold.
- S$20.45 versus S$20.40 cap mechanism.
- Applicable eligibility classifications.
- Statutory formulas and rounding.
- Historical amendment and effective-date treatment.

Record the evidence supporting each conclusion and the remaining uncertainty.

Do not select a convenient interpretation to make the package implementable.

## 5. Byte-grade verification gate

When original source bytes become available through a reliable retrieval method or user-provided official snapshots:

- Preserve the original HTML/PDF or other source artifact.
- Calculate SHA-256 over the original artifact bytes.
- Record the exact retrieval method and timestamp.
- Independently compare the original artifact against the provisional transcript.
- Check formulas, tables, footnotes, amendments and effective dates.
- Record missing, altered or incorrectly extracted content.
- Retain both artifacts and their respective digests.

Only after this verification may the relevant source evidence be considered pinned under the established research methodology.

## 6. Stop conditions

Throughout this pass:

- Keep `SG-MOM-OT-001` BLOCKED.
- Do not claim byte-grade source provenance.
- Do not implement overtime calculations in Go.
- Do not create golden fixtures based on unresolved legal interpretations.
- Do not remove existing Phys-Pseudo `?` findings.
- Do not modify the upstream Phys-Pseudo repository.
- Do not reopen the frozen WorkforceOps architecture.

## 7. Independent review

Have an independent reviewer examine the provisional captures, extracted claims, source locators, arithmetic and unresolved discrepancies.

The reviewer must distinguish what is established by the captured content from what still requires original-source verification.

## 8. Final report

Report:

1. Sources captured.
2. SHA-256 digests and exact hash scope.
3. Provisional provenance limitations.
4. Findings on both overtime discrepancies.
5. Historical and current rule distinctions.
6. Phys-Pseudo package changes.
7. Remaining questions and blockers.
8. Exact evidence needed for byte-grade verification.

The objective is to make useful progress without weakening the project's provenance standard.

## 5. One parallel action

Apply the already approved Phys-Pseudo integration amendments now. They do not depend on external source retrieval.

My final decision: Option 1 for immediate S3 reconnaissance, followed by byte-grade verification before source promotion or implementation. Option 2 remains a fallback if the agent cannot capture important source content accurately. Option 3 is unnecessarily conservative because we can already do useful evidence gathering.

The governing principle remains: we may research provisionally, but we must never implement provisionally verified statutory rules.