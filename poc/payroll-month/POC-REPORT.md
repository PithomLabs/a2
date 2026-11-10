# POC report: one-month synthetic payroll (Gate A complete, Gate B executed)

Status: GATE B EXECUTED 2026-11-10 (docker restored, no-prompt lifted).
POC suite green (e2e 13/13, synth 10/10, workload 9/9, 0 skipped);
full month exit 0; Exp1/Exp2 PARTIAL + Exp3 PASS with named gaps;
E5 PG18-vs-PG16 comparison identical (zero version deltas). Nothing
below claims a green experiment that was not executed — gaps stay
explicit (Exp2A baseline unbuilt, approval-wait + full writer-race +
host-corruption harnesses missing, poc_verify_claims BROKEN).

Branch: `stack-poc` (uncommitted). Plan: `plans2/plan_prompt2.md` (consolidated
Phase 6.1). Research base: `docs/stack/` + ADRs. Step 0: NO human rulings
arrived with `go` — Gate A ran on POC-scoped provisionals (P1 narrowed, P2/P3
held, P4 closing, P5 superseded; finding #4 = recorded-note default, all
PROPOSED). Gate B/C need rulings + a docker host.

## Evidence taxonomy (Implemented / Built / Executed / Passed / Failed / Blocked / Unexecuted)

| Item | I | B | E | P | F | Bl | U | Note |
|---|---|---|---|---|---|---|---|---|
| synth calc + vectors + pack versions | x | x | x | x | | | | 10/10 green incl. perf + data-only |
| workload 25x31 + canonical + 6 outputs | x | x | x | x | | | | 9/9 green; digest 3cef505e; `poc-run canonical` |
| seedpoc DB seeder (services-mapped) | x | x | x | x | | | | 25 emp / 24 sheets / 1060 clocks, exit 0 |
| DBOS workflow shape (capture/month) | x | x | x | x | | | | v1.8.0 live; 30 SUCCESS / 98 checkpoints |
| ledger appender + outbox + backlink | x | x | x | x | | | | 19 backlinks, max tx 20; P0001 live |
| Gate-2 POC DDL | x | x | x | x | | | | applied split-role; T13 0 warnings; gate blocks live |
| U4 spike + adapters + estimate | x | x | | | | | x | zero-adapter refuted; thin adapters build; refactor still estimated |
| auditor + genesis procedure | x | x | x | x | | | | txs 18/19/20 pass exit 0; out-of-band genesis commit pending |
| worker + failpoints + kill/forged/idem tests | x | x | x | x | | | | kill×3 + forged + idem green (inj #4 missing) |
| poc-run modes + subcommands | x | x | x | x | | | | full month exit 0 `overall: pass`; canonical DB-free |
| E1/E2/coverage/outbox harnesses | x | x | x | x | | | | E1 green; E2 42601 tripwire; Coverage + BROKEN Part 2 pin |
| T1–T13 battery | x(spec) | | | | | | x | spec in ADR-0203; exhaustive matrix out of prompt6 scope |
| E5 suite-on-18.6 + PG16 compare | x | x | x | x | | | | IDENTICAL both versions; 0 version deltas |
| compose (digests pinned) + runbook + CI | x | x | x | x | | | | stack healthy; runbook 1–6 executed; CI YAML still never run |

## DeepSeek 8-point criteria

1. **Executability (E5):** GO — PG18.6 full battery + PG16.15 comparison both
   executed; behavior IDENTICAL (16 ok pkgs each; fails identical: 11 RED +
   1 TZ-env + Gate2 load flakes); zero PG18-specific failures.
2. **Attack battery:** spec RED→COMPLETE (T1–T13 in ADR-0203); exhaustive
   matrix out of prompt6 scope — focused cases executed instead (overlap,
   gate, fabrication, forged-state, tamper).
3. **Forensic readiness:** PROVEN LIVE — 19 backlinks, auditor pass exit 0,
   bogus-tx loud failure, P0001 append-only vs owner, genesis procedure run
   (out-of-band commit still operator-pending).
4. **Crash/tamper:** EXECUTED — kill matrix (pre-commit, 2 outbox
   boundaries, step-window sweep) + tamper battery + forged tripwire all
   green; exactly-once from history (30 SUCCESS / 98 checkpoints, dup
   submit → single row).
5. **Out-of-scope guarantees:** statutory content (OT-001 BLOCKED, activation guard green), Gate 1 ruling, production migration/00013, PG16 removal, paid-hosted licensing — none touched, none implied.
6. **Evidence versioned + timestamped:** DBOS v1.8.0 (tag 2026-10-28), immudb v1.11.2 (tag 2026-09-02, Change ~2030-09-02 computed), pgx v5.12.0 (MVS), both image digests (Hub API 2026-11-10), all access dates recorded.
7. **Single command:** yes — runbook steps 1–6 + CI workflow (unexecuted); `poc-run` exit codes honest.
8. **SUCCESS only on pass:** yes — required-check gating, CI requires DB (fails without), demo-only is an explicit documented mode.

## F1–F7 acceptance (F1 split per Z#6)

| ID | Criterion | Status |
|----|-----------|--------|
| F1a | Blocker-path oracle (SHG/comp fixtures → typed blockers, no silent pass) | PASS (DB-free): 5 typed blockers, tripwire tests green; DB-backed blocking via gate + constraints live |
| F1b | Computed-path oracle + re-derivation incl. calc digest | PASS (synthetic scope): canonical digest `3cef505e…` + rerun-identical green; statutory half red-by-design (expected, not failure) |
| F2 | 13 translated attacks green, mapping committed | PARTIAL: spec committed; focused cases green (overlap, gate, fabrication-naive, forged, tamper); exhaustive matrix out of scope |
| F3 | Kill-tests pass; exactly-once from history | PASS (minus inj #4): pre-commit + 2 outbox boundaries + step-window + dup-submit green; 30 SUCCESS / 98 checkpoints |
| F4 | Tamper fails verification loudly (+ webhook) | PASS (minus webhook/unbuilt host-corruption): P0001 + bogus-tx loud fail + auditor pass green |
| F5 | Pack swap changes outputs; provenance attests; old re-derives | PASS (synthetic): v1a→v2a boundary 5400/7200 + digest stability green; attestation live (19 backlinks, marker-bound) |
| F6 | One-command bring-up on a clean machine | PASS (modulo clean-machine rerun): runbook steps 1–6 executed in order on this host; CI YAML still never run |
| F7 | Deletion ledger kept/deleted with rationale | Proposal stands (RESEARCH.md §6); backlink append-only proven live (P0001 vs owner) |

## Executed evidence (this environment)

Go env throughout: `GOWORK=off GOCACHE/GOMODCACHE/GOPATH=/tmp/* GOTOOLCHAIN=local`
(parent `go.work` + read-only default cache make bare `go` unusable — pre-existing).

1. Rename purity: 95 files, 258+/258−, ZERO non-import lines
   (`git diff -U0 -- '*.go'` filtered both ways); patch saved to
   `/tmp/rename-proof.patch` (1563 lines) for a separate commit.
2. Skip fix: root cause = `t.Skip` inside `sync.Once.Do` (Goexit marks Once
   done → zero TestDB → dial defaults). Hoisted `dbutil.DSN(t)` above Once at
   3 sites (admission, revoke, gate2b); package went 33 FAIL → 53 skip, 0 fail.
3. Versions: DBOS v1.8.0 + immudb v1.11.2 VERIFIED on proxy.golang.org
   (@v/list + @v/*.info git hashes); checkout-vs-v1.8.0 diff confined to
   failure-mirror restructuring + Patch refactor + cancel ctx — atomic path,
   completion dedup, ownership fencing, replay rules re-verified intact;
   immudb tree-identical for API files. Pins build + test green, pgx MVS v5.12.0.
4. Digests: `postgres:18.6-alpine@sha256:300ebd02…` and
   `codenotary/immudb:1.11.2@sha256:3033c1b5…` read from Hub API (amd64);
   composed, never invented.
5. synth: 10/10 green (vectors incl. corrected OT denominator, pack change,
   gap/overlap blocking, determinism, activation guard, bad input, data-only
   v3 variant, perf tripwire 100 months + digest stability).
6. Modes: `--demo-synth-only` exit 0 (`pass`); bare without env exit 1
   (`full-month`, `fail`); demo digest stable `259564d6…` (OT 262500¢).
7. U4 spike: zero-adapter hypothesis REFUTED by compiler (sysdb-internal
   return types); thin adapters build; census 105 pool refs / 434 calls;
   estimate recorded (2–4d mechanical + 1d boundary review + battery re-run).
8. Migrations untouched: `git diff -- migrations/` empty; no 00013 file.
9. `gofmt` clean all touched dirs; `go vet` clean POC module; CI YAML parses.

## Provisionals (current; details in TRACEABILITY.md)

P1 narrowed (registry pins, Change Date computed, R1 still OPEN); P2 held
(auditor built, R2 OPEN); P3 held (same-DB + fallback doc, R6 OPEN); P4
closing (DDL + spike done, bodies + U4 refactor need Gate B, R5 OPEN);
P5 SUPERSEDED by (3)+(4). Finding #4 = recorded-note default (PROPOSED).

## Environment blockers (pre-existing, unchanged)

No docker daemon; no PG server binaries; no sudo; `task lint` fails on the
missing `sdk/` dir + parent `go.work` (gofmt itself clean; zero Go files
outside POC + rename + 3-line skip fix touched).

## Scenario-matrix coverage (delivered vs specified)

Running code (green here): synth + packs + determinism + DBOS/ledger/worker/
auditor shapes + all harnesses compiling + runner modes + CI YAML. Specified
but unexecuted: everything requiring PG/immudb/containers (E1–E6, T1–T13,
crash/tamper/forged/idem/double-enqueue, full month, auditor acceptance).

## Next steps

1. Step 0 rulings — DONE 2026-11-10 (R1–R7 + P1–P5 + #4, recorded in
   DECISION-REQUESTS.md).
2. Docker host: runbook steps 1–6 in order; record E/T/crash/tamper/month
   outcomes back into this report's tables.
3. U4 refactor + Gate-2 gate depth + F1a fixtures against live PG.
4. Gate C per-technology decisions on Gate B evidence.

## Gate B attempt 2026-11-10 (BLOCKED — external prerequisite)

One bounded capability check per prompt4 §3, then stopped. No Gate B step
started; nothing claimed green.

- `docker version`: client 29.5.2 present; daemon socket exists at
  `/var/run/docker.sock` but connect fails: `permission denied while trying
  to connect to the docker API` (socket `srw-rw----`, session identity
  `chaschel`, no `docker` group membership).
- Single standard path attempted: `sg docker` → `Cannot open audit
  interface - aborting` (sandbox restriction). No sudo available
  (no-new-privileges sandbox, broken sudo.conf), per §3 no sudo/cloud
  workarounds attempted or invented.
- Result: daemon PRESENT but NOT USABLE from this session — §3 prerequisite
  unmet. Gate B remains fully blocked; Gate A preserved untouched.
- Fix (operator, outside this session): run Gate B from a login session
  whose user is in the `docker` group (e.g. `sudo usermod -aG docker $USER`
  + re-login), or otherwise provide a reachable daemon; then re-run the
  §3 probes (`docker ps`, registry pull, digest re-verify) and proceed.
- Credentials for that run are ready: dev-only defaults in
  `poc/payroll-month/.env.poc.local` (git-ignored), template in
  `.env.poc.example`, documented in `POC-RUNBOOK.md` §1.

## prompt6 attempt 2026-11-10 (BLOCKED — same external prerequisite)

Preflight green (branch, §1 inputs, builds, unit suites, digest `259564d6…`
stable). Capability re-probe: docker socket EACCES persists, no group/sudo.
Per `plan_prompt6_review.md` (APPROVED + clarifications A/B/C, all accepted
as amendments), wrote `plans2/POC-FINAL-REPORT.md` + `plans2/POC-RESULTS.md`
as NOT RUN/BLOCKED with the exact prerequisite. No baselines built blind, no
results claimed, POC retained as incomplete. Awaiting docker-group session.

## Gate B execution 2026-11-10 (EXECUTED — docker restored, no-prompt lifted)

Full Gate B + prompt6 Phases 2–6 executed on the live disposable stack.
Evidence details in `plans2/POC-FINAL-REPORT.md` (§§2–6) and
`plans2/POC-RESULTS.md`; workload record in `plans2/PHASE2-WORKLOAD.md`.

Executed this session (all live, all quotable):

10. Runbook step 3.5 split-role DDL apply: outbox.sql + gate2_poc.sql
    non-T13 as `migration_admin` (0 warnings), T13 range as superuser
    `poc` (0 warnings); T13 assertion `$t13$` commits silently; grants
    verified claim/close=SI, outbox=SIU, backlink=SI.
11. Seeder `poc-run seed-poc` exit 0 (attempt 4): 25 employees, 5 roster
    weeks, PP-2026-10, 24 approved sheets (SYN-023 nosheet by design),
    29 leave rows, 1060 clocks. Earlier attempts fixed Sunday-rest
    validation (workload zeroes Oct 4/11/18/25 + `TestSundayRest`),
    EV-002 structural match, and `poc_require_gate` search_path
    hardening (SET search_path + schema-qualified refs + assertion).
12. Full month `poc-run` exit 0 (`overall: pass`): workflow
    `month:poc-meridian:2026-10`, 5 phases; DBOS pool-close leak
    visible on stderr (known, `ClosePoolTimeout` workaround).
13. POC suite with stack env: e2e 13/13 (220 s), synth 10/10,
    workload 9/9, 0 skipped; `go vet` + `gofmt` clean. One
    order-dependence found and fixed en route (DBOS sticky IDs →
    unique run IDs in kill/capture-window/idem/forged tests).
14. Exp1A/B contrast probe (live, cleaned up after): plain table kept
    2 overlapping rows silently; WITHOUT OVERLAPS table rejected the
    overlap (exclusion violation, 1 row kept); naive snapshot
    fabrication rejected by the `payroll_run` FK.
15. Standalone auditor: backlinks txs 18/19/20 all verified, overall
    pass, exit 0, genesis `state_hash 4003ecfe…` (out-of-band commit
    still operator-pending).
16. E5 comparison: canonical superuser-base battery on PG18.6 (16 ok /
    15 fail) and PG16.15 (16 ok / 14 fail); failure sets identical
    except a Gate2 load-flake swap (both pass in isolation on both
    versions). 11 payroll fails are intentional TDD RED (D1/T2/EF-E3/
    §A/§B); attendance `TestConflictingRedelivery` is a UTC-session-TZ
    artifact (stored instant correct, `::text` renders 01:00 not
    09:00); codebase carries no `SET TIME ZONE` by design. Zero
    PG18-specific failures. Stale-DB + wrong-base-role + missing-PG16-
    password pitfalls tripped and corrected during the runs are
    recorded in `plans2/POC-FINAL-REPORT.md` §6 (harness lessons).

Remaining gaps (explicit, not rounded up): Exp2A conventional baseline
unbuilt; injection #4 (approval wait) + full writer-vs-capture race +
host-level immudb corruption harnesses missing; `poc_verify_claims`
BROKEN (42703, T10/T11 UNPROVEN); U4 refactor still estimated; CI YAML
never run; auditor webhook unbuilt; BSL-1.1 counsel review pending.

## 11. Remediation supersession (prompt8, 2026-11-10, branch `stack-poc`)

This section supersedes §§2–10 where they disagree. The historical
verdicts above are preserved as the pre-remediation record; the
evidence below replaces their conclusions. Full report:
[plans2/POC-REMEDIATION-REPORT.md](../../plans2/POC-REMEDIATION-REPORT.md).

- E1 (§3): PARTIAL PASS → PASS. PG-1 closed at the trigger layer:
  every app-role fabrication write fails 42501 (battery 5/5 green
  ×2, live). T13 applied superuser-only via the POC overlay
  (`gate/pg_authority.sql`, `$poc_auth$` silent); superuser-only
  is the documented design, not a gap.
- E2 (§4): PARTIAL PASS → PASS. The real Stages A–G month workflow
  runs on DBOS with transactional capture, crash batteries green
  (B1/error-retry/approval-crash/forged/double-enqueue/conflict),
  and E5 PG18≡PG16 zero-delta evidence retained.
- E3 (§5): PASS-with-gaps → PASS. Content auditor rebuilt (§5.2,
  16 per-event + 6 global checks, green on the real month);
  9-case tamper battery + forgery/missing batteries fail loudly;
  verified recovery (§5.4); reconciling state machine + guard
  hardening (§5.6, IM-10 closed even against the owner);
  corrections append with lineage (§5.5); genesis pinned +
  canaried and least-privilege identities (§5.3); IM-8
  host-corruption harness EXECUTED/PASSED (`gate/im8-corrupt.sh`,
  4-cell matrix in `/tmp/p8-evidence/im8/`).
- §7 (IM-7 standby): closed — the content auditor fails swapped,
  mutated, and tampered content loudly (9/9 cases + forgeries).
- §8 (IM-8 standby): closed — executed, PASS (see above).
- New §7-equivalent (Phase E, this report has no §7 slot):
  October-2026 MERIDIAN demonstration EXECUTED/PASSED —
  run `fb6187b6…`, crash injection + recovery, digest
  `4beee632…` == canonical, net 8521857¢ both sides, 102
  lines / 24 payslips / 24 finalized / 1 blocked / 5 exceptions,
  period close + lockout proven (harness-reopened), correction
  lineage audited (2 events PASS).
- Canonical pins (superseding any earlier `3cef505e…` draft
  values): payroll digest `4beee632…b20377`, net 8521857¢
  (EV-002 convergence; remediation report §9).
- Gates (2026-11-10, this stack): `go test ./...` exit 0
  (e2e 638 s), `go vet` clean, `gofmt` clean.
- Unchanged: statutory SG-MOM-OT-001 BLOCKED; conventional-job
  baseline NOT RUN; frozen migrations untouched; no 00013;
  production readiness NOT claimed (promotion checklist in
  `plans2/P8-PROD-MIGRATION-PROPOSAL.md`).
