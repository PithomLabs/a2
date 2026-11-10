# POC Remediation Report — prompt8 Root-Cause Correction

Date: 2026-11-10. Branch: `stack-poc`. Scope: synthetic one-month payroll
POC only (`poc/payroll-month/`); no statutory computation, no frozen
migration touched (00001–00012 intact, no 00013), no commit/push/merge.

This report supersedes the pre-remediation verdicts in
[POC-FINAL-REPORT.md](POC-FINAL-REPORT.md),
[POC-RESULTS.md](POC-RESULTS.md), and
[poc/payroll-month/POC-REPORT.md](../poc/payroll-month/POC-REPORT.md)
by explicit per-section supersession notes. Historical failed results
are preserved, never erased.

Result taxonomy (binding): IMPLEMENTED / BUILT / EXECUTED / PASSED /
FAILED / BLOCKED / NOT RUN. PASS/FAIL/BLOCKED/NOT RUN for verdicts.

## 0. Verdict summary

| §11 condition | Verdict | Evidence |
|---|---|---|
| 1. App role cannot fabricate+validate | PASSED | PG-1 battery: app direct writes → 42501, 5/5 green ×2 |
| 2. Real capture in PG authoritative gate | PASSED | Capture only via `poc_snapshot_capture` DEFINER + gate trigger; authority battery |
| 3. Inputs captured inside DBOS txn | PASSED | Stage B `RunAsTransaction`; B1 kill leaves no checkpoint, retry converges |
| 4. Calc executes on persisted inputs | PASSED | `LoadSnapshot`→`CalcMonth`; recompute == persisted (auditor C-checks) |
| 5. Outputs persist in PostgreSQL | PASSED | 102 lines + 24 payslips for October MERIDIAN run |
| 6. DBOS recovery, no dups/stale | PASSED | Retry/approval-crash/capture-convergence batteries |
| 7. Changed inputs → explicit handling | PASSED | `TestInputConflict_ChangedInputs`: conflict, no silent adopt |
| 8. Finalization+approval protected | PASSED | Approval-gated validate; forged SUCCESS rejected (DB-5) |
| 9. Finalized evidence → outbox | PASSED | Auto-enqueue at finalization (Stage G); authorized payload derivation |
| 10. immudb anchors snapshot+calc provenance | PASSED | Record v2 (input/output/rule/calc digests); verified append |
| 11. Auditor recomputes+verifies content | PASSED | Content auditor: 16 per-event + 6 global checks, all green on real month |
| 12. Alterations detected (threat model) | PASSED | 9-case tamper battery + forgery/missing + IM-8 host corruption |
| 13. Originals immutable thru corrections | PASSED | §5.5 battery + 2-event audit with lineage |
| 14. Month reproducible from pinned evidence | PASSED | DB-backed digest `4beee632…` == canonical; net 8521857¢ both sides |

Full POC suite: `go test ./...` exit 0 (e2e 638 s final gate
2026-11-10; earlier green runs 447 s / 599 s), `go vet` clean,
`gofmt` clean. Per §11, the remediation is COMPLETE on the evidence
below. Remaining limitations are listed in §10 — none of them fails a
§11 condition.

## 1. Original root causes

### PG-1 — Application-side fabrication path (adversarial review §PG)
The app role held direct INSERT/UPDATE on payroll tables, so a
compromised or buggy application could fabricate a snapshot and mark
it valid without ever passing the capture gate. Reproduced live in
Phase A (Gate B): app-role INSERT into `payroll_snapshot` succeeded.
Root mechanism: authority lived in application code, not in the
database; direct SQL could bypass every check.

### DB-3 — Capture identity accepted stale results (adversarial review §DB)
Month reruns did not re-execute: the same capture ID discarded new
input, so a rerun silently returned the previous outcome instead of
the current truth. Reproduced live in Phase A. Root mechanism:
idempotency keyed on the attempt rather than bound to (business run,
input digest).

### IM-7 — Auditor verified existence, never content (review §IM)
The auditor called `VerifiedTxByID` per tx and reported pass — it
never compared ledger bytes against PostgreSQL content. A swapped or
mutated record verified clean. Reproduced live in Phase A (swapped
and mutated records audited PASS). Root mechanism: no content
binding anywhere in the verify path; the backlinks file was trusted
input rather than a claim to check.

### Contributing causes
- IM-9: recovery read the idempotency marker with unverified `Get`
  (single unverified read on the link path).
- IM-10: retry accounting could mark an appended-but-unlinked row
  failed, orphaning attested evidence by threshold.
- IM-11: backlink PK `(pg_table, pg_id)` blocked legitimate
  correction events for the same artifact.
- IM-12/IM-13: outbox/backlink writes were direct app SQL; no
  production path enqueued real evidence.
- DB-4/DB-5: no proven crash-retry story; returned results were
  trusted without persisted-state corroboration.
- IM-1/IM-2/IM-3: genesis never pinned or compared; writer and
  auditor shared the superuser; no auditor read-only identity.
- IM-8: no host-corruption evidence (server-side file tamper).

## 2. Corrected architecture

PostgreSQL authorizes; DBOS durably executes; Go calculates from
captured records; PostgreSQL persists results; immudb anchors
evidence; the auditor independently reconstructs and verifies.

- Capture/validate/approve/close/enqueue are `SECURITY DEFINER`
  functions owned by `migration_admin`, with `search_path =
  pg_catalog`, explicit `EXECUTE` grants, and trigger guards. The app
  role has SELECT plus narrow EXECUTE — every direct fabrication
  write fails 42501 (proven, not asserted).
- The DBOS month workflow (Stages A–G) runs capture inside the
  transactional boundary, calculates from reloaded snapshot rows,
  persists lines/payslips, waits for approval, validates, and
  enqueues. Attempt identity is unique + Reject-on-reuse; business
  identity is (employer, period, kind) + input digest, with
  adopt-or-conflict convergence (`poc_run_consistency`).
- The outbox worker appends canonical evidence records under
  MustNotExist preconditions, recovers through VERIFIED reads only,
  and links via the squat-proof recorder. Proven appends can never
  be discarded (trigger-enforced, even against the owner).
- The auditor is a separate identity (read-only immudb user +
  SELECT-only PG role, own state dir, pinned genesis) that rebuilds
  expected evidence from PG sources, re-executes the payroll, and
  compares everything against verified ledger bytes.

## 3. PostgreSQL authority changes

Overlay: `poc/payroll-month/gate/pg_authority.sql` (Parts 1–9,
re-applied idempotently; live DB converged; `$poc_auth$`
assertions silent). Base coherence: `ledger/outbox.sql` CHECK + PK
updated to the converged state for fresh stacks.

- Phase B: revoked app direct writes on all payroll/POC tables;
  authorized DEFINERs for capture (`poc_snapshot_capture`), approve
  (`poc_approve`), validate (`poc_run_validate`), close
  (`poc_close_period`), persist (`poc_result_persist`), enqueue
  (`poc_outbox_enqueue`), backlink record (`poc_backlink_record`);
  gate trigger participation for the real capture path
  (`poc_gate_participate`); transition guard `poc_outbox_guard`;
  backlink PK widened to `(pg_table, pg_id, event_uuid)`; outbox
  CHECK widened to the 7-state machine.
- Phase D1: guard hardening (IM-10 closure) — `failed`/`poison`
  require `immu_tx IS NULL`; `appended`/`reconciling`/
  `backlinked`/`reconciled` require it set; `appended → failed/
  poison` and `reconciling → failed/poison` removed; INSERT branch
  added (construction invariant). Probed live: 6 owner-level
  discard attempts → P0001; legitimate paths unaffected
  (`TestOutbox_Authority` probes 6–7).
- Fresh-stack roles: `gate/init-auditor-role.sh` (SELECT-only
  `auditor` role + default privileges), mounted as `20-auditor.sh`.

## 4. DBOS workflow and transaction changes

`dbosflow/payroll.go` (Stages A–G), `capture.go`, `calc.go`,
`close.go` (pool-hang workaround); `cmd/poc-run` wiring for
`month` / `capture-once` / `approve` / `recover` (+ `verify`).

- Real payroll workflow: readiness → capture (transactional) →
  calc (reload + `ComputePay`) → persist → approval wait (Send/
  Recv) → validate → enqueue. Unique attempt IDs + Reject;
  adopt-or-conflict on (business run, input digest); DB-5
  submitter-side corroboration (`verifyPayrollResult`: status +
  tip + output digests re-read from PG).
- Crash battery (all EXECUTED/PASSED): B1 capture pre-commit kill
  (no checkpoint; retry converges; killed+retry reference one
  run); error-retry under new ID converges to the single verified
  outcome; approval-crash `recover` completes exactly once
  (drain-early `recover`, 1 run/output/approval/outbox row);
  forged SUCCESS moves nothing; double enqueue Rejects;
  changed inputs conflict loudly.
- Shared pure mapping extracted for the auditor:
  `BuildPersistLines` (+ `PersistLine`/`PersistSlip`) used by both
  Stage C and the independent re-execution (no duplicated logic).

## 5. immudb evidence and auditor changes

- Record v2 (`ledger/ledger.go`): `input_digest` (tip hash),
  `output_digest` (persist canon digest), `sha256` (payload-bytes
  hash), rule versions + first-pin pack digest, `calc_digest =
  sha256(calc_version label)` — honestly a version binding, not
  implementation provenance (stated, not oversold) — approval,
  supersedes. No PII/amounts/secrets in the record.
- Verified-read API: `VerifiedReadMarker`, `VerifiedTx`,
  `VerifiedReadRecord`, `VerifiedReadRecordAt`, `VerifiedTxHash`
  (deterministic-protobuf content hash for the genesis pin).
- Worker (`worker/outbox.go`): verified recovery (§5.4: marker →
  origin tx → tx re-verify → record-at-tx digest + artifact
  binding); reconcile path for proven anchors (never re-appends,
  never poisons); backlink failures route `appended → reconciling`
  (stays actionable); B4 failpoint (`after-appended-before-
  backlink`); `--op` pinned drains for hermetic batteries.
- Content auditor (`cmd/auditor`, rewritten): content mode
  (default) + labeled liveness mode; no backlinks-file input;
  read-only immudb user + SELECT-only PG role; writer-dir refusal;
  genesis pin (`--expect-uuid/--expect-tx/--expect-tx-hash`,
  fresh-tempdir check before any trust, exit 2); `--accept-genesis`
  ceremony printing the committable anchor. Verification equation
  A–L (§2 of `content.go`): output/input bindings, persist-bind,
  field-level re-execution match (lines + payslips), PG-side canon
  recompute, pins, calc digest, approval, supersedes (+chain
  resolution), payload SHA, identity + marker/record origins +
  byte-level digest closure, finalized-only, all-events + coverage.
- Identities (§5.3, all EXECUTED): `poc_writer` (RW) +
  `poc_auditor` (R) via `poc-run immudb-users` (idempotent);
  auditor `ExecAll` → permission denied (probed live + committed
  test); PG `auditor` role SELECT-only (INSERT/EXECUTE denied,
  probed live); compose + env + runbook updated.
- Genesis (EXECUTED 2026-11-10): anchor committed at
  `gate/immudb-genesis.json` (`dbpeo9bns5bc73erm8f0`, tx 1,
  `cf000ba4…f63cc`), byte-identical across two independent
  ceremonies. Design correction found live: verified reads advance
  trusted state to the TIP, so a state-dir hash stales on every
  append — the pin is UUID + tx-1 content hash instead (proven
  deterministic; canary test pins the live stack continuously).
- IM-8 host corruption (EXECUTED, `gate/im8-corrupt.sh`,
  evidence `/tmp/p8-evidence/im8/`): disposable instance, real
  battery month, 64-byte corruption of `defaultdb/val_0`: stale
  state + content → exit 1 (`J.record: data is corrupted: value
  length or digest mismatch`, server UP); stale + liveness →
  pass (control: tx-only checks never touch values — the IM-7
  critique confirmed empirically); fresh + content → fail;
  fresh + liveness → pass (TOFU control). VERDICT: PASS.

## 6. Exact file and SQL changes

Go (all under `poc/payroll-month/`):
`dbosflow/{payroll,capture,calc,close}.go`,
`cmd/poc-run/main.go` (+ `verify_test.go`),
`cmd/auditor/{main,content}.go` (client.go removed),
`ledger/ledger.go`, `ledger/outbox.sql`, `worker/outbox.go`,
`workload/*` (refactor-neutral), `seedpoc/*` (EV-002 convergence),
`e2e/{authority,chaos,dbos,forged,idempotency,poc,auditor,phasee}_test.go`.

SQL/ops: `gate/pg_authority.sql` (guard hardening + state
machine), `gate/init-auditor-role.sh` (new),
`gate/immudb-genesis.json` (new, ceremony output),
`gate/im8-corrupt.sh` (new, IM-8 harness),
`docker-compose.poc.yml` (identities + init mount),
`.env.poc.local` / `.env.poc.example` (5 credential sets),
`POC-RUNBOOK.md` (§3.7/3.8, tamper §4), auditor Dockerfile
(unchanged — env-agnostic).

Frozen migrations 00001–00012: untouched (verified by git status
showing no modification under the migrations tree). No 00013.
PG16 volume/data: untouched (POC ran on disposable PG18.6 :5544).

## 7. Before-and-after attack results

| Attack | Before (Phase A repro) | After (remediation battery) |
|---|---|---|
| App fabricates+validates snapshot | SUCCEEDED (live) | 42501 on every direct write; DEFINER-only paths |
| Rerun with new inputs, same ID | DISCARDED new input, stale result | Reject-on-reuse; changed inputs → loud conflict |
| Swapped/mutated ledger record | AUDITED PASS (IM-7) | Content audit FAILS naming the check (9/9 tamper cases) |
| Unverified marker recovery | Single unverified read (IM-9) | All recovery reads verified (marker+tx+record binding) |
| Threshold orphans append | Possible (IM-10) | Trigger-forbidden, even for owner; reconciler repairs |
| Second event for artifact | PK rejection (IM-11) | Appends with lineage; both verify |
| Threshold→failed on proven row | Worker did it | Never: reconciling stays actionable (B4 + reconcile tests) |
| Genesis/substitution games | Unpinned, shared superuser | Pinned anchor + canary; distinct least-privilege users |
| Host file corruption | No evidence (IM-8) | Stale state fails loudly; 4-cell matrix recorded |
| Approval crash / error retry | No proven story | Recovered exactly once; retried under new ID, converged |
| Forged SUCCESS / forged backlink | — (new battery) | Rejected (DB-5 corroboration / J-checks) |
| Cross-run idem-key reuse | — (new battery) | Payload-aliasing refusal (proven live by a test bug, kept honest) |

## 8. Actual monthly payroll output (Phase E)

Run `fb6187b6-3f9b-4cee-946d-0f3aa2b61e52` (POC-MERIDIAN, PP-2026-10,
kill -9 mid-approval + DBOS recovery + approval + validation):

- Members 25 (included 24, excluded 1: B-BASE-CHANGE); claims 673; pins 2.
- Lines 102, payslips 24; finalized 24; blocked 1; exceptions 5.
- Net pay 8,521,857¢ (canonical and DB-backed identical).
- Payroll digest `4beee632a05c09b9ec0be0dbedc1937d517461603d8660999d66f51d04b20377` (both sides).
- Period closed by app role (writers locked, duplicate rejected),
  harness-reopened afterwards (documented).
- Evidence drained (`backlinked`), content audit PASS (16+6
  checks), correction event appended → 2-event audit PASS.

## 9. Digest and reproducibility evidence

- Canonical (DB-free `workload.Canonical` + `MonthPayrollDigest`):
  `4beee632…b20377`, net 8521857¢.
- DB-backed (snapshot reload → `CalcMonth` → same digest fn):
  identical, asserted in-test (`TestPhaseE`).
- Persist canon digest (`poc_result_persist` form) recomputed by
  the auditor from persisted rows: equals ledger `OutputDigest`
  (check D, every content audit).
- Byte-level closure: ledger record digest == reconstructed
  record digest (check J.bytes, every event).
- Full suite `go test ./...`: exit 0 (e2e 638 s final gate);
  `go vet` clean; `gofmt` clean. Rerunnable: battery employers
  recycle, MERIDIAN recycles, unique IDs per run (aliasing guard
  proven).
- E5 PG18≡PG16: zero version deltas (prior evidence, retained).

## 10. Remaining limitations (honest, none failing §11)

1. Calculator-version binding, not implementation provenance:
   `calc_digest = sha256(version label)`. A compromised calculator
   binary is outside the stated threat model (noted in code).
2. `PackDigest` = first pin's digest; multi-pin joint digest is
   future work (single-pin fixture; documented in code).
3. Period `"2026-10"` hardcoded in Stage C and auditor alike
   (same literal both sides; fixture-scoped, not data-driven).
4. Auditor `SHA256` check hashes `payload::text` as the instance
   serializes it (same-instance determinism; cross-version jsonb
   text form not staked).
5. `supersedes` predecessor lookup has no sub-microsecond tiebreak
   (matches enqueue exactly; ties unobserved).
6. Production ledger hardening (TLS, signing keys, stable volumes,
   retention, monitoring) is a promotion checklist
   (`P8-PROD-MIGRATION-PROPOSAL.md`), NOT implemented in the POC.
7. immudb BSL-1.1 counsel review still pending (~2030-09-02 note).
8. Conventional-job baseline still unmeasured (classified
   NOT RUN throughout; never claimed).
9. DBOS `Shutdown` pool leak persists upstream (ClosePoolTimeout
   workaround retained + documented).
10. Clock/attendance/terms/leave/comp writers remain gate-only
    (no period mapping) — recorded proposal limitation.
11. IM-8 mechanism observed is server-side envelope detection;
    proof-vs-forged-bytes is covered compositionally (state-tamper
    + pin + binding checks), not by a forged-serving server cell.
12. Stale `poc_ledger_backlink` rows accumulate (append-only by
    design); audit_event retention hygiene still open.

## 11. Revised technology recommendations

- PostgreSQL: RETAIN as the authority. The remediation proves the
  value is in constraints/triggers/DEFINERs + RLS-free grant
  discipline: every fabrication path now fails closed in the
  database, independent of application code. Gate C verdict: PASS.
- DBOS (Transact-Go): RETAIN for durable execution with the
  documented workaround. Proven: transactional capture,
  at-least-once recovery converging without duplicates, approval
  orchestration, Reject-on-reuse identity. Cost: Shutdown pool
  leak (workaround), stream-polling recovery latency. No
  alternative engine evaluated (out of scope, §12). Gate C: PASS
  with workaround noted.
- immudb: RETAIN as tamper-evident anchor, NARROWLY scoped.
  Proven: append-only records, verified reads, independent
  auditor verification, host-corruption detection, genesis
  pinning. NOT proven: production hardening (TLS/keys/ops).
  The POC uses immudb only where PG cannot serve: history
  attestation external to the database under audit. Gate C:
  PASS for the POC scope; production requires the promotion
  checklist + license counsel.
- Conventional-job baseline: NOT RUN — no comparative claim made.

## 12. §11 completion mapping

See the verdict table in §0 (14/14 PASSED with named evidence).
Per prompt8 §11, the remediation is COMPLETE. No condition was
failed, bypassed, or weakened; every check that could be executed
was executed, and everything not executed is classified NOT RUN
above — never implied green.
