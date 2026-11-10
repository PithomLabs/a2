# PROMPT9 — Final Adversarial Review of the Post-Prompt8 POC

Per plans2/prompt9.md. Read-only review: no code modified, no services
touched, no docker available from this session (socket EACCES). All verdicts
are static code evidence plus the DB-free suites and an independent Python
recomputation of the payroll digest executed during recon. Severity scale is
prompt9 §10.

Method: the remediation report, P8 migration proposal, current
final/results/POC reports with supersession notes, and the full article text
were read, then four parallel read-only tracks (PG authority, DBOS execution,
immudb/auditor evidence, cross-layer + article + figures) opened every cited
file body. The six most load-bearing new claims were re-verified against
source by the reviewer (leave trigger gap, invalidate revoke absence,
pin-collapse loop, supersedes self-inclusion, hardcoded period both sides,
digest-golden absence).

## 1. Executive verdict

The remediation is genuine, large, and substantively effective. The three
P0-class root causes from the prior review are actually fixed within the
app-role threat model: direct fabrication now fails closed in the database
(grants + DEFINER-only capture verified table by table), capture identity
is bound to (business run, input digest) with reject-on-reuse and loud
conflict, and the auditor rebuilds expected evidence from PostgreSQL and
compares byte-identical digests. The October figures check out — the
digest `4beee632…b20377` and net 8521857¢ were independently recomputed
with a separate implementation, byte-exact. The article is defensible: 9
of 10 technical claims PASS, none is contradicted or unsupported.

But 14/14 does not mean defect-free. This review finds **zero P0 and five
P1**: a mid-capture leave race through the one untriggered input table,
cross-artifact backlink forgery through an under-bound recorder, a
surviving frozen DEFINER that bricks runs, first-pin-only rule-pack
binding on a two-pin month, and auditor re-execution through the same
mapping it audits. All five are concrete, each is fail-loud or narrowly
scoped rather than silently catastrophic, and none overturns the POC's
credibility — but each narrows a named guarantee and must be recorded
before this POC is preserved as a demonstration.

Bottom line: **preserve the POC as a credible technical demonstration,
with the article qualified as noted and the five P1s plus the understated
limitation (§10, item 5) recorded as residual risks.** No production
authorization — that was never on the table and nothing here changes it.

## 2. Scope and review methodology

Read-only review, same discipline as the prior round: the remediation
report, P8 migration proposal, current final/results/POC reports with
supersession notes, and the full article text were read, then four
parallel read-only tracks (PG authority, DBOS execution, immudb/auditor
evidence, cross-layer + article + figures) opened every cited file body.
The six most load-bearing new claims were re-verified against source by
the reviewer (leave trigger gap, invalidate revoke absence, pin-collapse
loop, supersedes self-inclusion, hardcoded period both sides,
digest-golden absence) and the cross-layer track executed the DB-free
suites plus an independent Python recomputation of the payroll digest. No
files modified, no services touched, no docker available from this session
(socket EACCES), so DB-backed assertions rest on code inspection while
canonical-side figures were executed and recomputed. Severity scale is
prompt9 §10. Only concrete, supportable findings are reported — several
"looks bad, actually fine" candidates (harness reopen, shutdown
workaround, TOFU-vs-pin separation) were investigated and cleared, noted
where instructive.

## 3. Post-Prompt8 architecture summary

What changed is architectural, not cosmetic. PostgreSQL now authorizes
through `gate/pg_authority.sql` (1612 lines): app is SELECT-only on
authoritative tables with narrow EXECUTE on seven DEFINERs (capture,
persist, approve, validate, close, enqueue, backlink-record), a 7-state
outbox machine under a transition guard, widened backlink PK, and a
`$poc_auth$` assertion battery. DBOS runs a real seven-stage workflow
(`dbosflow/payroll.go` A–G: readiness, transactional capture, reload +
`ComputePay`, DEFINER persist, Send/Recv approval, validate + enqueue)
with unique attempt IDs, Reject-on-reuse, and adopt-or-conflict
convergence on (run, input digest). Go calculates from reloaded snapshot
rows via `LoadSnapshot → CalcMonth → BuildPersistLines`. immudb anchors
Record v2 (input/output/rule/calc digests, approval, supersedes) through
verified-reads-only recovery plus a reconcile path, and a rewritten
content auditor reconstructs expected records from PostgreSQL, re-executes
payroll, and checks 16 per-event + 6 global conditions against verified
ledger bytes, pinned to a committed genesis anchor. The IM-8
host-corruption matrix executed with logs preserved (ephemerally, in
`/tmp/p8-evidence/im8`).

## 4. PostgreSQL findings

The authority boundary holds against direct attack — and three P1s
survive around its edges.

**P1-1. `leave_request` has no gate trigger; mid-capture race validates
stale day-inputs.** Capture claims leave rows
(`pg_authority.sql:714-719`) and the member loop reads them (`:612-617`),
but `leave_request` is absent from the 15-table trigger array
(`:1204-1210`, verified by reading the list) while holding full default
app DML with zero revokes. An approve/insert landing between the member
read and the claim `INSERT..SELECT` — same READ COMMITTED txn, no
serialization — yields members missing the row while claims include it
with matching hash; the in-txn verifier sees final state and the decision
check never re-derives day-level inputs. Post-capture drift is caught
(fail-closed); only the mid-capture window is silent. PROVEN
(composition; all three code sites confirmed). Fix is one line: add the
table to the trigger array and assertion list.

**P1-2. `poc_backlink_record` binds neither artifact identity nor
employer — cross-artifact forgery + near-permanent audit poisoning.** The
function (`:1421-1459`) checks outbox-row existence, state, and tx match,
but never compares `(p_pg_table, p_pg_id)` to the outbox payload's refs.
Pure app-role SQL (enqueue own run → set `appended` with invented tx →
record backlink naming a victim run) returns `'linked'`. The auditor then
fails loudly for the victim — detective control works — but the row is
un-deletable by every role (unconditional trigger, recycle skips
backlinks), so the victim's audit fails until owner DDL surgery.
Same-artifact squats are properly rejected; the "squat-proof" claim is
true only within one artifact. PROVEN.

**P1-3. Frozen `invalidate_payroll_snapshot` left executable by app — any
run brickable.** The 00012 grant of EXECUTE to app was never revoked
(zero mentions in `pg_authority.sql` — confirmed by grep) and no
assertion covers it. One call flips a tip `valid → invalidated`;
re-capture then dies 23505 on `(run, version 1)` while validation fails
for want of a tip — run bricked both directions. Post-validate it
de-tips a validated run and strands evidence (`poc_outbox_enqueue`
fails). No POC caller uses it, so the fix is a one-line REVOKE in the
overlay. PROVEN.

**P2.** Re-applying the *old* files regresses the boundary (`outbox.sql`
re-grants; `gate2_poc.sql` downgrades gate/close then halts mid-file on a
return-type error, leaving a partially committed downgrade under
autocommit) with no runbook order guard — fresh-stack order converges,
re-apply does not (P2). `poc_result_persist` validates membership but
never amounts, and payslips skip even membership (excluded/non-member
employees can receive payslips) — the amounts half is a by-design trust
boundary that should be stated alongside limitation 1, the payslip
asymmetry is an oversight worth one `decision='included'` check (P2).
App-INSERTABLE `poc_workflow_attempt` lets app SQL hijack `approveCmd`'s
latest-waiter routing via `ORDER BY created_at DESC LIMIT 1` —
availability only, the DEFINER approval record stays correct (P2).
`$poc_auth$` never asserts `poc_outbox_guard` trigger existence, so a
dropped guard leaves app UPDATE unconstrained with the assertion still
green (P2). Verifier tolerates extra claims for known tables and hashes
omit owner columns — mostly latent behind frozen revokes,
"both directions" is true for freshness+omission but not
claim-membership (P2, partial).

**P3.** `closed_by` always records `migration_admin` (DEFINER
`current_user` semantics) and the outbox-guard DELETE role check is
vacuous (enforcement rests solely on the grant revoke, which holds);
limitation 10's gate-only list omits five tables and mislabels
`leave_request` as gate-only when it is triggerless; the
transaction-local `search_path` widening (`:862-870`) is safe under
current grants with minor residue (persists for the caller's txn, sibling
paths rely on test-green rather than asserted qualification);
authority-test gaps (no app-role probes for line/slip/output/approval/
close-table writes, close probes confounded by frozen triggers except
INSERT, no negative P1-2 probe); bundled nits (app DML on unused tables,
CHECK re-apply window, mutable `created_at` feeding supersedes order,
TEXT event_uuid case variants, no wrong-tx correction path, doc nits).

What held: direct fabrication blocked on all 10 tables (grant→trigger
matrix verified); verifier actually wired in-transaction by capture
(`:750`) and validate (`:1007`), EXECUTE withheld and asserted, vacuous
sets rejected; close is DEFINER with DB clock, duplicates rejected, real
`pay_period_id` joins for six writers.

## 5. DBOS findings

No P1. The four priority properties hold: tx-threading is real, the
identity/conflict matrix is correct, approval cannot be forged into
validation, and tests assert business rows. Two P2 convergence/integrity
gaps follow.

**P2-1. Finalize replay after evidence drain fails with a squat error;
post-drain retries do not converge.** `poc_outbox_enqueue` derives
`supersedes` with no self-exclusion (`pg_authority.sql:1392-1396`; no
`c.op_uuid <> v_op`), while the auditor excludes self. First enqueue
stores `supersedes=''`; after the event backlinks, any replayed finalize
(unconditional, `payroll.go:386-390`) re-derives `supersedes=<self>` ≠
stored `''` → `squats a different payload` P0001 → workflow ERROR. Month
completes, drains, operator reruns → loud failure. Fail-closed
(rollback, nothing corrupts) but contradicts retry-converges for the
entire post-drain window; pre-backlink retries work. Fix mirrors the
auditor: exclude `v_op`. PROVEN.

**P2-2. Calc params read live, never pinned or claim-hashed; owner-level
params mutation is fully silent.** `LoadSnapshot` joins live
`poc_rule_pack_schedule.params` (`calc.go:105-110`); the pack claim hash
covers pack/version/dates/status/digest but not params; the auditor
compares the stored digest column and re-executes with the same live
params. A params-only mutation (same version, digest column untouched)
passes consistency, changes Stage C output, passes validate, passes
audit. Requires owner write (app is SELECT-only) — outside the
app-attacker model but contradicts "pinned evidence" reproducibility.
The Part 1 comment claiming Stage C and the auditor "recompute them from
workload.MonthPacks() and compare" is false — no such code exists
(grep-verified). Fix: pin params bytes at capture or hash live params in
the claim. PROVEN.

**P2 (partial).** Forged DBOS rows/notifications are a fail-closed DoS on
approval delivery, never a bypass: app can write `dbos.*` (same-DB grants
unavoidable) and squat `poc_workflow_attempt`, but every variant fails
closed at the DEFINER approval gate — `Recv` is a wake-up, the PG row is
the gate. Availability impact reasoned, not executed — hence partial, P2.

**P3.** Approval binds run identity only (no digest in record or message;
drift fails closed at validate — verified path by path; post-hoc audit
cannot show what the approver authorized beyond run identity;
`approveCmd` wakes only the latest PENDING attempt). Validated-stands
capture race returns a hard error instead of adopting (tiny window; retry
converges). Hardcoded `"2026-10"` in calc path (consistent both sides;
wrong for any other period). Dead `CalcVersionPinned` const (real
enforcement is the DEFINER check). Mod-10 `businessCounts` fingerprint
(compensated by explicit zero-checks). `verifyPayrollResult` `[:16]`
slices panic on short digests (unreachable via genuine outputs; corrupt
digest → CLI panic, still fail-closed). Stale "commit landed, checkpoint
lost" comment (window impossible on shared-pool path; replay logic still
needed for cross-attempt adoption).

What held, specifically: all seven stages do real business work
(readiness SELECTs, DEFINER capture, `LoadSnapshot → CalcMonth →
synth.ComputePay → BuildPersistLines`, DEFINER persist of
lines/slips/output, real `Send`/`Recv`, validate + enqueue in one txn);
all four `RunAsTransaction` bodies use only the tx handle with zero pool
references in txn paths (U4 threading genuinely done — the `dbexec`
adapters are an unreferenced spike); shared-pool pointer identity
confirmed through SDK `runAsTxn`; Reject-on-reuse at both submissions
with adopt-or-conflict exactly as specified (same digest adopts,
different digest errors, never adopts); `verifyPayrollResult` re-reads
status + tip + output digests via INNER JOINs requiring `validated`;
kill/error-retry/approval-crash/forged-SUCCESS/double-enqueue/
changed-inputs batteries assert business counts with status rows
supporting only.

## 6. immudb findings

Two P1s, both evidence-completeness rather than funds-exploit class —
plus confirmation that IM-8 and genesis need careful scoping.

**P1-4. October evidence binds only the FIRST of 2 rule-pack pins, and
same-pack versions collapse in the map.** The pin loop
(`worker/outbox.go:281-285`) overwrites `ruleVersions[pin.Pack]` per
iteration and keeps the first digest: ledger binds `{SYNTH-OT-001: v2a}`
+ v1a's digest only. The real month has 2 pins (asserted in
`phasee_test.go:145-146`, stated in remediation §8, two packs seeded).
An owner edit of v2a's manifest digest — or v1a's version label — passes
`E.rule-pack`, `E.rule-versions`, and `C` (params come from the live
schedule, not digests): audit green, provenance silently altered.
Amount-impact attacks are still caught compositionally via C/D recompute
+ canon, so this is evidence-incompleteness, not a funds exploit. But
limitation 2's parenthetical "(single-pin fixture)" is **false for the
real month** — contradicted by the report's own §8. The battery's
pin-digest case tampers all pins at once, so it cannot catch a
second-pin-only edit. PROVEN, including a live log line showing the
collapsed map.

**P1-5. Auditor re-execution shares the exact mapping under audit;
mapping bugs verify clean.** The auditor calls
`dbosflow.CalcMonth + BuildPersistLines` (`content.go:164,169`); Stage C
calls the identical functions (`payroll.go:278,282`), both hardcoding
`"2026-10"`. A sign/pay-code/CPF-treatment/`source_refs`/payslip-shape
bug persists wrong outputs and recomputes the identical wrong rows →
`C.lines`/`C.slips` and `D` all pass. Independent *data* through shared
*mapping*: a tamper oracle, not a correctness oracle. No independent
line-level oracle exists (`workload.Canonical` covers totals, never the
line mapping). The "shared pure mapping" disclosure is honest but the
consequence is unstated, and remediation §4 frames sharing as pure
virtue. PROVEN.

**P2.** IM-8 proves server-side envelope refusal, not client
proof-vs-forgery: both content cells fail with the identical server error
(`data is corrupted: value length or digest mismatch` — the server
refuses to serve), so trusted state played no role and the stale/fresh
distinction is vacuous for the assertion; the Merkle path against
tampered-but-well-formed values was never exercised. Verdict PASS is fair
for accidental-host-corruption, not for a Byzantine server; limitation
11's "covered compositionally" is generous (the state-tamper cell
exercises tx-proof, not value-proof rejection). The evidence was
independently verified present and exact at `/tmp/p8-evidence/im8` —
ephemeral location is itself a preservation risk. Genesis: the
pin-enforcement code is correct (fresh-tempdir check, exit 2, ceremony +
canary paths), but the "committed out-of-band" step is unexecuted
(anchor file untracked; its own "lives in git" note is false — and
`poc/` and `plans2/` are entirely untracked worktree-wide, so every
"committed" in the report means worktree-present), "byte-identical across
two ceremonies" has no transcript, the canary never runs in CI
(UNEXECUTED workflow, no auditor env), and no live substitute-server test
exists (P2). Untriggered negative paths: tx-origin checks,
`H.supersedes` failure, and second-pin digest tamper have zero battery
coverage; tx-swap is prevention-only (P2 — code reads correctly, but
prompt9 §6C demands per-attack detection evidence).

**P3.** Auditor PG-role EXECUTE denial untested (mechanism sound via
PUBLIC revokes; report claims "probed live" but the committed test
probes INSERT only). Poison-accounting residuals are loud and convergent
(crash-window + 10 failures → visible `failed` needing documented
operator reset; owner DELETE of `reconciling` orphans the driver with
loud audit failure) but untested. Hygiene: limitation 6's pointer to the
P8 proposal as "promotion checklist" dangles (the 121-line doc contains
no TLS/signing/volume/retention checklist — "NOT implemented" itself is
accurate); tiebreak flake risk is fail-loud not silent; `VerifiedTxHash`
cross-append stability rests on reported green canaries, not re-proven
here.

What held: zero unverified reads on any link path (plain `Get` exists but
is dial-tone/test-only); reconcile never re-appends or poisons
(construction + trigger + tests); `J.bytes` byte-closure genuine (same
marshaled bytes stored and hashed); line/payslip field matching exact;
canon-digest forms identical both sides; least-privilege identities with
committed `ExecAll`-denied and writer-dir-refusal tests; `calc_digest`
limitation stated in all four places with no oversell found; 16+6 check
counts exact per the baseline log.

## 7. Cross-layer transaction and recovery analysis

Traced one operation end to end: PG records → DEFINER capture (gate +
verifier in-txn) → DBOS checkpoint (same txn) → snapshot → Go calc from
reloaded rows → DEFINER persist → Send/Recv approval → DEFINER validate +
enqueue (one txn) → worker claim → immudb append → backlink txn → auditor
rebuild + re-execute + compare. Every *executed* boundary re-verifies
rather than trusting labels: capture/persist/finalize replays re-read
state, worker recovery is verified-reads-only, the auditor rebuilds from
PG sources. Three trust gaps, all minor-to-medium: **(a)** no autonomous
durable delivery exists — `RunOnceOp` callers are only the CLI single-row
drain and test harnesses, so crash-after-enqueue leaves evidence
undelivered with no alarm until something invokes a drain (P2, fails
silent on omission; the article's "durable worker subsequently publishes"
is true per-invocation but the worker is not self-driving); **(b)**
`approve-send` success reports `Send` enqueue, not waiter receipt —
pre-`Recv` Sends can miss, with retry living only in harnesses (P3);
**(c)** battery content audits run TOFU — pin enforcement proven only in
liveness mode, never combined content+pin in one test (P3). The
"harness-reopened" period flag investigated and **cleared**: reopen is
owner-only test privilege, app is SELECT-only on the close table, and the
writer trigger rejects locked-period flips while a close stands — not an
app-reachable backdoor. Crash semantics per boundary: pre-commit capture
kill → zero checkpoint, retry converges to one run (business rows
asserted); kill at approval wait → PENDING survives, `recover` completes
exactly once (1 run/output/approval/outbox row); B2/B3/B4 worker kills →
precondition recovery / reconcile, never re-append, never poison;
duplicate delivery → `replayed` convergence pre-drain, squat-loud
post-drain (see §5 P2-1 — the one convergence hole).

## 8. October payroll reconstruction verification

Independently recomputed from the canonical CLI JSON with a separate
Python implementation of `MonthPayrollDigest`: net **8521857¢**
(identical in register and payments totals), digest **byte-exact**
`4beee632a05c09b9ec0be0dbedc1937d517461603d8660999d66f51d04b20377`,
102 lines derived via the zero-omission rule, 25/24/5 shape confirmed.
The canonical-side numbers are arithmetically correct.

The gaps are assertion/pinning-side: **(P2)** no committed golden for
the digest, net, or 102-line count — the three most-quoted literals
appear only in Markdown (`4beee632` has zero `.go` hits), `TestDigestStability`
asserts self-equality + length only, and Phase E asserts `lines>0` plus
DB↔canonical equality. A correlated drift shifting both sides
identically passes silently, and `POC-RESULTS.md` calling them "pins"
overstates — there is no pin test. **(P2)** "five exceptions" blends
scopes: canonical = 1 employee + 1 day + 3 synthetic demos, while
DB-backed `CalcMonth` can only produce day-scope exceptions and Phase E
never asserts a DB-side count (it logs the canonical count in its run
line). Figures that *are* nailed down both sides: 25 members / 24
included / 1 excluded, 673 claims, 2 pins, 24 payslips, blocked-count and
digest/net equality (`phasee_test.go:131-187`). Run `fb6187b6` is
narrative-only (IDs are per-run UUIDs by design; figures deterministic,
IDs not — documented). 14-condition spot-check: evidence pointers check
out for all 14 with three weakenings — condition 5's "102 lines" words
exceed its `lines>0` assertion, condition 6's step-window sweep is
probabilistic by design (acknowledged in-test) with status-only counts in
the idem/step-window tests, and condition 8's "(DB-5)" attribution is
slightly off (e2e forge rejection comes from Reject-on-reuse; the
payroll-variant corroboration itself has no unit test).

## 9. Article technical-claim verification

Article: `plans2/writeup2.md`, "When Payroll Fails, Who Remembers What
Happened?" All 10 claims plus 6 flags checked against code with exact
quotes:

- Claims 1–6 (PG governs; DBOS durable memory; shared atomic boundary;
  calc from captured records; immudb anchors actual evidence; auditor
  reconstructs + verifies): **PASS** — each maps to verified
  implementation (DEFINER-only ops + 5×42501 exploit chain; Stages A–G +
  Recv + recovery batteries; shared-pool `RunAsTransaction` + B1;
  `LoadSnapshot → CalcMonth`; DEFINER-derived payload + Record v2;
  checks A–L with all-green genuine-pass test).
- Claim 7 (October figures + digest): **QUALIFY** — numbers verified
  correct by independent recompute and DB↔canonical equality is in-test,
  but no literal golden exists and "five exceptions" blends canonical
  scope into DB-run figures.
- Claim 8 (crash-during-approval): **PASS** —
  `TestApprovalCrash_RecoverAndComplete` plus Phase E SIGKILL at the
  approval wait with single-effect counts.
- Claim 9 (original unchanged after corrections): **PASS** — event-1
  tx-unchanged assertion plus green 2-event audits.
- Claim 10 (immutability ≠ correctness): **PASS** — the L183 disclaimer
  is exact and consistent with the disclosed calc-label limitation.
- Flags: exactly-once (explicit disclaimer present, reports' phrasing
  business-count-backed), immudb-proves-correctness (disclaimed),
  privileged-host protection (never claimed; detection-framed),
  production readiness (disclaimed twice), measured savings (qualitative
  only, no baseline comparison anywhere) — **no violation**.
  Omitted limitations: **QUALIFY** — the article discloses calc-label,
  multi-pack, and serialization limits plus open questions, but omits
  TLS/BSL/retention/Shutdown/gate-only-writers/hardcoded-period/
  tiebreak/baseline; mitigated by explicit repo pointers. Most material
  omissions for a business reader: **BSL-1.1 license and no-TLS**.

No article claim is CONTRADICTED or UNSUPPORTED. The strongest
article-level caveat is claim 7's figure-scope blending.

## 10. Known limitations and additional risks

All 12 §10 limitations re-verified against code: 1 (calc-label digest),
2 (first-pin digest — mechanism accurate; see below), 3 (hardcoded period
both sides — all three sites grepped), 4 (jsonb text canon), 6 (no
TLS/keys), 7 (BSL pending), 8 (baseline NOT RUN), 9 (Shutdown workaround
retained), 10 (gate-only writers — both sides tested, but the named list
is incomplete per §4), 11 (IM-8 bounded — accurate against the verified
logs), 12 (retention open — accurate, though outbox/DBOS/attempt-row
growth goes unmentioned) are **accurate**. Item 2's "(single-pin
fixture)" parenthetical is **false for the real month** (2 pins — P1-4).
**Item 5 is UNDERSTATED (P2):** beyond tie order, enqueueing a
correction *before* its predecessor backlinks bakes `supersedes=''`
into the immutable payload; once both backlink, the auditor permanently
fails `H.supersedes` with no repair path (payload immutable, same-op
re-enqueue squat-rejects, backlinks owner-append-only). All committed
tests enqueue corrections only after draining the predecessor — the
hazardous order is never tested and the constraint is documented nowhere.
Fails closed (loud), hence P2.

Additional risks not in the 12: the five P1s (§§4, 6); post-drain
finalize-replay failure (§5 P2-1); live-params-not-pinned (§5 P2-2); no
durable-delivery driver (P2, §7); payload-aliasing refusal claimed from a
test bug with zero covering tests (P2); IM-8 evidence in ephemeral `/tmp`
only; genesis anchor untracked with unexecuted out-of-band step;
PG16-"untouched" unverifiable read-only (the E5 battery necessarily
exercised a PG16 instance with test writes — "untouched" can only mean
non-test data, with no evidence in-repo either way).

Preservation checks: migrations 00001–00012 untouched (`git status` under
`migrations/` empty), no 00013 file, OT-001 still BLOCKED (no
`SG-MOM-OT-001` schedule row; `ActivateStatutory` rejects synthetic packs
with green guard test), no production-readiness claim (§11 POC-scoped
with promotion checklist required), supersession notes present in all
three historical reports with old PARTIAL/failed verdicts preserved
verbatim.

## 11. Findings register by severity

P0: none.

P1 (5): P1-1 leave_request triggerless mid-capture race
(`gate/pg_authority.sql:612-617,714-719,1204-1210`); P1-2 cross-artifact
backlink forgery (`:1421-1459`); P1-3 `invalidate_payroll_snapshot` still
app-executable (`00012:1089`, absent from overlay); P1-4 first-pin-only
binding + version-map collapse on the 2-pin month
(`worker/outbox.go:281-285`, `content.go:358-365`); P1-5 shared-mapping
auditor circularity (`content.go:164,169` vs `payroll.go:278,282`).

P2 (17): post-drain finalize replay squat-error; live calc params
unpinned (+ false Part 1 comment); forged-DBOS fail-closed DoS
(partial); old-file re-apply regression; persist-amounts trust boundary
+ payslip membership asymmetry; attempt-table waiter hijack; unasserted
outbox-guard trigger; verifier extra-claim tolerance (partial); IM-8
envelope-vs-proof scope; genesis out-of-band unexecuted + canary absent
from CI; untriggered tx-origin/lineage/second-pin negatives; no
digest/net/102 goldens ("pins" overstated); five-exceptions scope blend;
no durable-delivery driver; aliasing refusal untested; supersedes
ordering poisoning (limitation 5 understated); payload-aliasing refusal
without regression test.

P3 (17): approval run-only binding + latest-waiter wake;
validated-stands race error; hardcoded period; dead const; mod-10
fingerprints (2 files); `[:16]` panic-on-corrupt; stale comment; PG
EXECUTE-denial untested; poison residuals untested; `closed_by`
attribution + vacuous DELETE check; limitation-10 list incomplete;
search_path residue; authority-test gaps; Send-receipt gap;
`PayrollDigest` unpersisted + DB-5 attribution; `poison` unreachable +
`failed` reset undocumented; `im8-corrupt.sh` env bootstrap wart; TOFU
battery audits; untested `replayed` branches; bundled nits
(unused-table DML, CHECK re-apply window, mutable `created_at`, TEXT
uuid, no wrong-tx path, doc nits).

Reproduction: every finding reproduces from the cited file bodies without
a live stack. Highest-value live confirmations (disposable stack only):
mid-capture leave approve + validate; cross-artifact
`poc_backlink_record` returning `'linked'`; `invalidate_payroll_snapshot`
bricking a run; v2a-only digest edit auditing green; post-drain month
rerun hitting the squat error.

## 12. Final architecture assessment

Component verdicts (verified → residual → verdict):

| Component | Independently verified | Residual risk | Verdict |
|---|---|---|---|
| PostgreSQL | DEFINER-only writes; in-txn verify | Leave race; backlink bind; de-tip | Partially proven |
| DBOS | Real stages; tx atomicity; identity | Post-drain replay; live params | Partially proven |
| Go calc | Persisted-input calc; shared map | Mapping unaudited by auditor | Partially proven |
| immudb | Verified appends; reconcile | First-pin bind; envelope-only IM-8 | Partially proven |
| Auditor | Byte-closure; 16+6 checks | Circular mapping; TOFU batteries | Partially proven |
| Combined | Connected path, crash-tested | Delivery undriven; no goldens | Partially proven |

Guarantee classifications: fabrication blocked PROVEN (direct);
validated-run integrity PROVEN (via authorized path; de-tip is
availability); capture atomicity PROVEN; exactly-once business effects
PROVEN (per-operation mechanisms + business-count batteries);
dedup-under-resubmission PROVEN (Reject + conflict, new ground since the
prior review); durable approval PROVEN (real wait + crash recovery + PG
record); forged-state resilience PARTIALLY PROVEN (bypass blocked; DoS
reasoned); outbox recovery PROVEN (verified-only + reconcile); content
audit PROVEN as tamper oracle, NOT PROVEN as correctness oracle (shared
mapping); rule-pack binding PARTIALLY PROVEN (first pin only — the
month's second pin unbound); genesis pinning PARTIALLY PROVEN (code
correct, out-of-band step unexecuted); host-corruption detection
PARTIALLY PROVEN (envelope refusal, not proof-vs-forgery); PG↔canonical
equivalence PROVEN (equality in-test + independent recompute), goldens
NOT PROVEN (absent); close authority PROVEN for mapped writers, NOT
PROVEN for unmapped/triggerless inputs; article defensibility PROVEN
with two QUALIFYs.

The remediation report's core sentence — "the architecture now works" —
**survives scrutiny as a POC claim**: the connected path (authorize →
capture → calculate → persist → approve → validate → anchor → verify)
executes over real payroll data, crashes converge without duplication,
and tampering fails loudly across a genuine 9-case battery plus
forgery/missing/corruption coverage. It does not survive as anything
stronger: five P1s narrow the guarantees, limitation 5 understates a
permanent-poisoning hazard, and the POC still has no goldens, no delivery
driver, and no out-of-band trust root. Preserve it as a credible
demonstration with the qualifications in §§9–10 attached — and fix P1-1
through P1-3 (each a small, surgical change) before inviting anyone to
extend the authority boundary.
