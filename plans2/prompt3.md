My recommendation is to select Option 1 — Approve as drafted (Recommended).

Recommended: APPROVED-WITH-CONDITIONS

I verified the actual immudb v1.11.2 license, not just the release notes.

The BSL 1.1 license explicitly permits non-production use, including copying and modification. It also provides a limited production-use grant, but restricts certain paid hosted or embedded competitive offerings.&#x20;

[image](https://www.google.com/s2/favicons?domain=https://raw.githubusercontent.com\&sz=32)

GitHub+1



### Enter this ruling

R1 — immudb BSL-1.1

Decision: APPROVED-WITH-CONDITIONS

Effective date: 2026-11-10.

Conditions:

1. Authorize immudb v1.11.2 exclusively for the synthetic, non-production POC on disposable environments.
2. Pin the exact version and image digest. Retain the applicable LICENSE and version provenance.
3. No vendoring or embedding immudb implementation code into the main application codebase. Maintain the narrow, replaceable evidence-layer interface.
4. No production payroll data, commercial deployment, paid hosted offering or production dependency authorization under this ruling.
5. The estimated Change Date of 2030-09-02 is provisional. Verify the first-public-distribution date and timezone; obtain legal confirmation before treating it as definitive.
6. Any future production, hosted or paid use requires a separate licensing review and explicit approval.

Rationale:

The cryptographic provenance and independent verification capabilities justify evaluating immudb within the synthetic POC. Its BSL-1.1 licensing introduces a legitimate future commercial dependency risk, but that risk does not justify abandoning the technical experiment at this stage.

This approval does not authorize production adoption or distribution of immudb as an unrestricted open-source dependency.

One nuance: avoiding vendoring is a sensible containment measure, but it does not automatically eliminate licensing questions about a Go SDK linked into an application executable. That is another reason to require legal review before commercial deployment.

The v1.11.2 release was published around September 2–3, 2026, according to its release artifacts and GitHub release record. Therefore, the proposed 2030 date is a reasonable working estimate, not a legally established date.&#x20;

[image](https://www.google.com/s2/favicons?domain=https://github.com\&sz=32)

GitHub+1



### What this accomplishes

It allows the agent to proceed with:

- immudb append and verification experiments.
- Provenance and audit implementation.
- Tamper-detection testing.
- Dual-write recovery testing.
- Independent auditor experiments.

It does not resolve R2–R7 or authorize production changes.

My view is that we should not abandon a potentially valuable architectural capability because of a licensing question that can be isolated during the POC. Equally, we should not allow a promising technical experiment to become an irreversible commercial dependency.

Select Option 1. Keep the commercial decision separate until the POC demonstrates measurable value.