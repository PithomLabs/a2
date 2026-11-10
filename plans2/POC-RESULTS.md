# POC Results — Verdict Matrix (remediated 2026-11-10)

> SUPERSESSION (prompt8 remediation, 2026-11-10, branch `stack-poc`):
> every verdict below supersedes the pre-remediation matrix of the
> same name. The pre-remediation matrix is preserved verbatim in the
> section below (nothing erased); before/after attack results are in
> [POC-REMEDIATION-REPORT.md](POC-REMEDIATION-REPORT.md) §7. Full
> evidence: that report, §§0–12.

| # | Experiment | Original verdict | Remediated verdict | Evidence |
|---|------------|------------------|--------------------|----------|
| E1 | PostgreSQL integrity (T1–T13 + authority battery) | PARTIAL PASS (T13 BLOCKED) | PASS | App fabrication → 42501 (PG-1 closed); authority battery 5/5 green ×2; superuser-only T13 applied via overlay. T13 remains superuser-only BY DESIGN (documented, not a gap). |
| E2 | DBOS durability (real month workflow + crash battery) | PARTIAL PASS (no production mapping) | PASS | Real Stages A–G workflow; B1 kill converges; error-retry + approval-crash recover exactly once; changed inputs conflict; forged results rejected. E5: PG18≡PG16, zero version deltas. |
| E3 | immudb evidence (content auditor + tamper battery) | PASS with gaps (IM-7/IM-8/IM-10 open) | PASS | Content auditor (16+6 checks) green on real month; 9-case tamper battery; forgery/missing fail loudly; IM-8 host-corruption PASS; genesis pinned + canary-tested; least-privilege identities. |
| E4 | October-2026 integration (Phase E) | NOT RUN | PASS | MERIDIAN month with crash injection: digest `4beee632…` == canonical, net 8521857¢ both sides, 102 lines/24 payslips, close+lockout proven, correction lineage audited. |

Canonical workload pins (deterministic fixture): payroll digest
`4beee632a05c09b9ec0be0dbedc1937d517461603d8660999d66f51d04b20377`,
net 8521857¢, 24 finalized / 1 blocked / 5 exceptions / 102 lines /
24 payslips. (Pre-remediation drafts cited `3cef505e…`/8519957¢ from
an earlier fixture revision — superseded by the EV-002 convergence;
see remediation report §9.)

Gates: `go test ./...` exit 0 (e2e 638 s, this stack, 2026-11-10);
`go vet` clean; `gofmt` clean. Conventional-job baseline: NOT RUN
(never claimed). Statutory SG-MOM-OT-001: BLOCKED (unchanged).
Frozen migrations 00001–00012 untouched; no 00013.

Per-evidence classification (IMPLEMENTED/BUILT/EXECUTED/PASSED/
FAILED/BLOCKED/NOT RUN) is tabulated per finding in
[POC-REMEDIATION-REPORT.md](POC-REMEDIATION-REPORT.md) §§7–9 and
§§0–1.

---

## Pre-remediation record (preserved verbatim, superseded above)

# POC Results (demonstration summary)

Status: EXECUTED 2026-11-10 — POC module e2e 13/13 + synth 10/10 +
workload 9/9 green, 0 skipped; full month exit 0; E5 PG18-vs-PG16
comparison shows ZERO version deltas. Exp1/Exp2 PARTIAL (gaps named),
Exp3 PASS (gaps named). Full record: `POC-FINAL-REPORT.md`.

## What was executed

One shared synthetic month (POC-MERIDIAN, 25 employees × 31 days, Oct
2026, packs v1a→v2a) through three comparative experiments on disposable
PG18.6 + DBOS v1.8.0 + immudb 1.11.2: seeder (25 emp / 24 approved
sheets / 1060 clocks), full `poc-run` month (`month:poc-meridian:2026-10`,
exit 0), canonical digest `3cef505e…` (net 8519957¢), 13 e2e harnesses
(E1/E2/outbox/coverage/concurrency/repro/idem/forged/kill×3/tamper),
standalone auditor run (txs 18/19/20 verified, overall pass), Exp1A/B
contrast probe, and the main-module battery on PG18.6 AND PG16.15.

## What failed in each baseline

- App-only (Exp1A): plain table silently kept 2 overlapping pack rows —
  no error, no decision, nothing to alert on (live probe).
- Conventional-job (Exp2A): NOT BUILT (honest gap) — comparison is
  DBOS-measured vs recovery-inventory-argued, not measured-vs-measured.
- PG-only evidence (Exp3A): privileged DELETE on ordinary (trigger-less)
  outbox rows succeeds — internal history is rewritable by privilege.

## What PostgreSQL prevented

Overlapping pack versions (WITHOUT OVERLAPS exclusion), cross-employer
interference (advisory employer gate: same blocks, different free),
naive snapshot fabrication (payroll_run FK), app DELETE of outbox
history (42501), backlink rewrite by ANY role incl. owner (P0001
trigger). NOT prevented: need⊆coverage (undeclarable — E2 42601;
procedural verifier BROKEN 42703) and fabrication with valid parents
(still RED, EF-E3, both PG versions).

## What DBOS recovered

Pre-commit kill → zero checkpoint → restart exactly-once; post-commit
worker kills → pending → restart converges to backlinked; duplicate
submission → single SUCCESS, byte-identical rerun; forged `dbos.*`
rows move nothing. Shared-pool atomicity verified (business + durability
record commit together). NOT covered: approval-wait crash (no such wait
in the workflow); calc-step kill only approximate (timed window).

## What immudb detected

Independent auditor verifies genuine anchors (`overall: pass`, exit 0)
and fails LOUDLY on a bogus tx; backlink tx == marker origin tx after
backlink-missing recovery; digest-bound idempotency keys make ambiguous
retries safe. Ledger proves records unchanged, not calculations
correct. Host-file corruption harness unbuilt; genesis out-of-band
commit still pending (operator step).

## The final payroll output and its verification status

Canonical month: 24 final + 1 blocked employees, 24 payslips,
5 typed exceptions, net 8519957¢, digest
`3cef505ea5eb…caa46ea` (`poc-run canonical`, DB-free, re-runnable).
DB-backed month: `overall: pass`, exit 0. All outputs labeled
illustrative/non-statutory; SG-MOM-OT-001 BLOCKED; no statutory
calculation anywhere. E5: PG18.6 ≡ PG16.15 (16 ok pkgs both; fails
identical: 11 intentional RED + 1 TZ-env + Gate2 load flakes).
