## claude

This matches `POC-REPORT.md`, so my earlier review stands. The environment can't run Docker or PostgreSQL, so the code builds but nothing database-dependent has run. The best use of the agent's no-database time is closing the scaffolding gaps, so the first run on a real host is worth the effort.

Below is a paste-ready follow-up. R1–R3 and R7 are legal, security and business calls, so I've written my recommended answers as drafts. Edit them before sending.

````markdown
# Follow-up: finish the POC scaffolding and prepare the first real execution

## 1. Decisions (human rulings; apply as written)

- R1 (immudb license): APPROVED-WITH-CONDITIONS for POC non-production use only. Record pinned version and Change Date in the report. Keep the narrow replaceable interface. No production or hosted use without a separate ruling.
- R2 (state custody): APPROVED for POC with this interim model: the auditor runs in a separate container with its own volume and credentials; its trusted-state file is never mounted into the writer; the genesis state hash is recorded outside the system (in a git commit message or file). Make no tamper claim beyond this model. Name `migration_admin`, PostgreSQL superuser and immudb admin as outside the model.
- R3 (threat model): Keep Rung 5. Do not use DBOS `Rewind` or `Delete` on payroll workflows. Treat the DBOS system tables as untrusted orchestration state: approvals and state transitions are authorized only by PostgreSQL records enforced in the database, never by DBOS messages or checkpoints. Do not enable immudb truncation.
- R4 (PG 18.6): APPROVED now, on disposable clusters. Do not remove PG16 from the dev stack until E5 and the battery are green on 18.6.
- R5 (Gate 1 interplay): APPROVED. Gate-based capture on disposable databases only, labeled as evidence for Gate 1, not as a ruling.
- R6 (DBOS placement): Same database, separate `dbos` schema. Reopen if the forged-state test (task 4) fails.
- R7 (D1–D8): adopt as labeled working assumptions for the synthetic POC; production sign-off stays open.

## 2. Hygiene first

- Prove the import rename is purely mechanical: show that `git diff -U0 -- '*.go'` has no changed lines other than the `PithomLabs/a1` to `a2` import paths. Keep it as a separate patch from POC work.
- Fix `core/database` DB-gated tests to skip when `WORKFORCEOPS_TEST_DATABASE_URL` is unset. Touch nothing else in that package.
- Replace local-checkout `replace` directives with registry versions for DBOS and immudb if reachable; otherwise record exactly what blocks it. Do not hardcode digests you cannot verify.

## 3. Build now (no database needed to write and compile)

1. External auditor: a small program that keeps verifier state in its own file, runs `VerifiedTxByID` or dual-proof checks against the immudb server, and exits non-zero with a clear message on failure. It must be runnable from a separate container.
2. Failpoints in `poc-run`: named failpoints at the five boundaries (pre-commit in the capture transaction, after PG commit before ledger append, after append before back-link, and the two DBOS step boundaries), enabled by an environment variable.
3. Automated kill matrix: a Go test that spawns `poc-run` with each failpoint, sends SIGKILL, restarts, and asserts persisted state in PG and the ledger (no partial rows, no duplicate or missing ledger records, one outcome row). It must skip loudly without the stack.
4. Forged-state test: as the `app` role, insert or alter DBOS checkpoint, notification and workflow rows (a forged approval `Recv`, a forged step output). Assert that no business invariant is violated: no state transition occurs without a PG-recorded, database-enforced authorization.
5. Tx-threading spike (UNRESOLVED-4): implement the smallest change that lets one `RunAsTransaction` callback run the capture body on the same `pgx.Tx` as the DBOS checkpoint. Write it so it compiles and has a test that runs once a database exists. Estimate the refactor cost for the real services and record it.
6. Missing experiments: idempotent double-enqueue, the performance sanity timings, and a new synthetic rule type added as data without Go changes. If the overtime formula can't be expressed as data, say so and describe the smallest data model that would.
7. A CI workflow (for example GitHub Actions) with PostgreSQL 18.6 and immudb as service containers that runs steps 2-6 of the runbook, and uploads results as artifacts.

## 4. Order of execution on the first real host

1. E5: migrations 00001–00012 and the full suite on 18.6. If it fails, stop and record the failure verbatim. Do not edit frozen migrations.
2. E1–E3 (temporal constraints), then the Tx-threading and gate-plus-capture spike under `kill -9`.
3. Forged-state test.
4. immudb outbox crash matrix, then tamper test with the auditor.
5. Only then run the full month scenario.

After each stage, record a go / no-go for that technology with the evidence tag, and stop if a stop condition fires.

## 5. Rules

Same constraints as before: no edits to migrations 00001–00012 or Gate 1 artifacts, no statutory logic, all work uncommitted, never claim a result that wasn't executed.
````

Two cautions before you send it:

- Check that your Docker host or CI runner can reach the Go module proxy and the image registries, or section 2's version pinning will block again.
- If you'd rather not answer R2 and R3 yet, send the rest and keep those two marked open. Items 1 to 3 don't depend on them.


## deepseek

# Review — Phase 6 Status

## Verdict

**Honest, productive, and correct to stop where the environment ran out.** The work is real: the module rename closed U1, the DBOS interop scratch closed the pgx-MVS risk, and the synthetic calc caught an actual denominator bug via vectors. Those are the right kind of wins. But **"code-complete" must not be read as "POC complete,"** and the gate cannot pass on compile-and-unit evidence alone. The M2 adversarial review warned about exactly this: test-green ≠ invariant-closure. Here, the entire critical path — the translated battery, the Gate-2 mechanism, DBOS crash semantics, immudb recovery, PG18 temporal constraints — is in the "unexecuted" bucket, and the one package where those mechanisms live (`core/database`) is also the one that cannot run its tests.

This is a **compile-and-unit milestone**, not a POC-complete milestone. That is fine — but it changes what the human can sign.

---

## What Phase 6 actually proves

| Claim | Status | Quality of evidence |
|---|---|---|
| Module `a1 → a2` rename consistent | Proven | `go build ./...` exits 0; 258 imports rewritten |
| DBOS SDK + pgx v5.12 interop | Proven | Scratch module compiles + runs under MVS; pgx-skew risk closed |
| POC module builds and tidies | Proven | `go build`/`go mod tidy` clean |
| Synthetic calculation is deterministic and correct on 8 vectors | Proven | 8/8 green; caught a real overtime denominator defect |
| Demo digest is stable across runs | Proven | `259564d6…` reproducible |
| `poc-run` fails loudly without env | Proven | Exits 1 in demo mode |
| F7 (POC-report proposal) | Done | Documented |
| F1 / F5 partial (pure-Go proofs) | Partial | Calc determinism only, not snapshot integrity |

These are worth having. The calc vectors catching a defect is genuinely the strongest signal in this report that the methodology is working — the vectors were not written to pass, they were written to catch.

---

## What it does not prove — against the 8-point criteria

The POC-success definition from the prior review had eight points. Phase 6 delivers on roughly two of them and parts of two more:

| # | Criterion | Phase 6 |
|---|---|---|
| 1 | Full month run end-to-end via CLI+REST | **Not executed** — no DB |
| 2 | T1–T13 translated attacks both orderings | **Not executed** — no DB, and Gate-2 gate DDL not written |
| 3 | Kill-9 at 5 dual-write boundaries, recovery replays once | **Not executed** |
| 4 | PG tamper detected; ledger tamper → `ErrCorruptedData`; auditor webhook fires | **Not executed** |
| 5 | PATTERNS C1–C12 demonstrated live | **Partial** — C9 (money purity) and C12 (rule dynamics) partially via vectors; C1–C8, C10–C11 untested |
| 6 | Byte-identical re-derivation from pinned inputs + pack | **Partial** — demo digest is stable, but the full capture → persist → re-derive loop does not exist yet |
| 7 | POC-REPORT.md naming what's proven and what remains | **Done in draft** — should be checked against the final POC state |
| 8 | `00001–00012` untouched; no `00013` exists | **Should be verified** — is it, in the report? |

So of eight criteria: roughly **2 green, 2 partial, 4 not run**. That is not a gate failure — it is what the environment allows. But the gate result must say so.

---

## Three gate-blockers that are *work*, not environment

These are worth calling out because the status frames them as remaining code work, but they are actually gate-decision inputs:

### 1. `core/database` unit tests cannot run (pre-existing skip-logic gap)

The status says: "all packages pass except `core/database`, whose admission/gate2 tests dial PG unconditionally — a pre-existing skip-logic gap."

That is the package where the D1 mechanism, the XID ticket, and the Gate-2 gate direction live. **"Unit suite green except the one package that owns the P0 fix" is not a benign exclusion — it is a gap that must be closed before the gate can say "the unit suite is green."** The fix is small (skip if no `WORKFORCEOPS_TEST_DATABASE_URL`, matching the pattern used elsewhere), but it should be done and it should be tracked as a defect, not a footnote.

Otherwise, on the docker host, the first thing the runtime battery will do is *discover* whether those tests actually pass — which is not the same as "unit suite green here."

### 2. Gate-2 gate DDL not written

ADR-0203 chooses the Gate-2 transaction-scoped gate as the POC capture mechanism and specifies the translated battery (T1–T13) against it. The status says the "Gate-2 gate DDL on disposable DBs" is remaining code work.

That means **T1–T13 cannot run today even with a database**, because the mechanism they test does not yet exist as DDL. The battery is a specification, not a suite. This is the most consequential outstanding item: the entire point of Phase 6 was to build the mechanism so the battery proves it. Phase 6 has built the *harness* and the *harness's dependencies*, but not the mechanism.

ADR-0203's acceptance criterion — "no POC capture code path bypasses the gate txn (writer-matrix review)" — cannot be evaluated until the DDL exists.

### 3. UNRESOLVED-4 (Tx threading) is a correctness blocker, not a refactor

PATTERNS C5 says atomic capture requires `RunAsTransaction` on the shared pool, meaning the capture body must run inside an explicit `*pgx.Tx`. The status lists this as "remaining code work."

It is more than refactor. Without it:
- PATTERNS C5 ("Capture is atomic incl. durability record") is not implemented.
- The kill-9 pre-commit experiment (Phase 6 criterion 3) is not meaningful — there is no atomic boundary to crash inside.
- ADR-0203's consequence "`a2` services take `*pgxpool.Pool`; threading `Tx` into `RunAsTransaction` callbacks needs refactor" was flagged as **cost unestimated** in the ADR. It remains unestimated.

The prior review asked for a spike estimate before sign-off. That is still pending, and it is the only item in Phase 6 that can plausibly move the POC from 3–5 days of runtime work to 2–3 weeks.

---

## Answering the R-items is not gated on runtime

R1–R7 are business and architecture decisions. They do not require a database to answer. The correct sequence is:

1. **Answer R1–R7 now** (with the amendments from the prior review, especially R6's three-way re-framing and R2's "POC departure" language).
2. **Confirm or reject P1–P5** — these are POC-scoped and explicitly binding nothing in production. That framing is correct.
3. **Close the three work-blockers above** on a docker-capable host:
   - `core/database` skip logic
   - Gate-2 gate DDL
   - UNRESOLVED-4 Tx-threading spike (estimate first, implement second)
4. **Then run the 8-point POC** on the docker host per the runbook.

Doing 1 and 2 first means the runtime work happens against signed decisions, not against provisional ones. Doing 3 before 4 means the battery runs against a complete mechanism, not a partial one.

---

## What to record in `POC-REPORT.md`

The status already mentions the report; it should be verified to contain:

- The **8-point POC criteria** and the current pass/partial/not-run status for each.
- The **three work-blockers** as open items with owners.
- The **P1–P5 provisional choices** with their non-binding scope.
- The **U4 estimate** once spiked.
- The **F-table** (currently F7 done, F1/F5 partial, F2/F3/F4/F6 not run).
- The **demo digest** (`259564d6…`) as evidence that determinism holds in the pure-Go layer, explicitly distinguished from snapshot integrity.
- The **calc defect caught by vectors** (overtime denominator) as evidence the vector method works — this is a genuine sign of health, not a footnote.

---

## Bottom line

Phase 6 is exactly what a research stage looks like when the environment stops cooperating: honest, evidenced, and correctly scoped. The right response is not to push harder on compile-only claims or to declare the POC "done pending runtime." The right response is:

- **Answer R1–R7 and confirm P1–P5 now** — they are decision items, not runtime items.
- **Close the three work-blockers** (skip logic, Gate-2 DDL, U4 spike) so the runtime battery runs against a complete mechanism.
- **Run the 8-point POC on a docker host** and let the evidence decide the gate.

The strongest sentence in the status is "everything needing a database honestly reported unexecuted." Keep that discipline. The gate's integrity depends on it — the M2 adversarial review was written because someone, at some point, did not keep it.


## Z

Reviewed the full research-gate package (9 documents) against the original brief, the repo's own binding constraints as cited, and internal consistency across docs. 

# Verdict: APPROVE AS GATE PACKAGE — with revisions required before R1–R7 sign-off

This is strong work. The agent honored the method rules I pinned (evidence tags VERIFIED/DOCUMENTED/INFERRED/ASSUMPTION, criteria-before-build, deletion ledger, honest no-go conditionality), and it caught four landmines that would each have cost real time: the **immudb BSL-1.1 license flip** (R1 is exactly the right top decision request), the **PG18 Docker volume-layout change** (old `/data` mount silently not persisting), **DBOS at-least-once reality for externals** (verbatim documented, treated as binding until crash-proven), and the **immudb PG-wire silent-DDL-skip hazard**. The self-approval records correctly state the ADRs do not pass the gate. Stopping before POC build is legitimate — the repo's binding plan mandates research-gate-first, and the environment lacked Docker (my brief's stop condition §2.10 covers both). Know what that means, though: **F1–F7, the kill matrix, and the tamper battery are specified, not demonstrated.** Nothing here is evidence of the stack working; it is evidence the plan is sound.

## Findings (prioritized)

| # | Sev | Finding | Fix |
|---|---|---|---|
| 1 | **HIGH** | **ADR-0204's back-link design violates the repo's own immutability enforcement.** Back-links are `UPDATE ... SET ledger_tx` on "provenance-bearing tables" — but snapshot tables are insert-only enforced (M1.1), run rows are state-machine-triggered (K2/00004-pattern; a `ledger_tx` update isn't a status transition), and K2's own test byte-compares original rows. As written, step 4 of the dual-write protocol is rejected by the triggers the POC keeps. | Replace with an **append-only `ledger_backlink` table** (`pg_table, pg_id, immu_tx, alh, verified_at`, insert-once, unique on artifact). Preserves "supersede = new row, never edit," satisfies freeze triggers, simplifies the auditor join. Update ADR-0204 acceptance + PATTERNS C8/K7 accordingly. |
| 2 | **HIGH** | **The translated battery drops the privilege/lock-substitution class.** Legacy 12 included lock substitution and revoke-boundary tests (`TestLockSubstitutionUndetectable` et al.); T1–T11 covers 11 legacy classes, T12 is *new* — so the substitution/privilege class has no row. It still matters: L4 *keeps* advisory helpers, and an `app` self-acquiring session-level locks on gate keys can still conflict/DoS the txn-scoped gate. | Add **T13: advisory-key substitution/self-acquisition as `app`** + grant-assertion tests (advisory overloads revoked from PUBLIC, trigger ownership). Fold the revoke assertions into POC DDL and R3's scope. |
| 3 | **MED-HIGH** | **R6 is load-bearing but framed as configuration.** ADR-0201/0203 and PATTERNS C5 already *decide* same-DB shared-pool; if R6 goes separate-DB, atomic capture collapses to replay semantics and the whole battery changes shape. | Make the dependency explicit: sequence R6 *before* ADR-0203 is treated as settled; write the separate-DB fallback battery variant in ADR-0203 so the branch is planned, not improvised. |
| 4 | **MED** | **Binding Key Decision #4's letter is unmet.** It required the ADR to decide ticket-vs-gate *with a translated battery*; ADR-0203 instead rejects the ticket by citing Gate-2 ADR authority. The battery tests the chosen mechanism only — there's no head-to-head. | Either run the cheap bake-off on disposable DBs (both mechanisms are small; ~1 day) **or** record explicitly that prior Gate-2 authority + refusal-to-resurrect satisfies the requirement, and get that confirmed via R5. Don't leave it implicit. |
| 5 | **MED** | **The dynamic-rules classification table (brief §3.5) is missing** — the deliverable I called the heart of the flexibility objective. The package has the pattern (C12) and pack governance, but no per-rule inventory (eligibility / proration / rounding / supersession / thresholds → effective-dated data vs versioned pack vs code-in-durable-step). | Add it to Phase 6 entry criteria or as an R7-adjacent artifact. Without it, "extensible across dynamic business rules" has no falsifiable plan. |
| 6 | **MED** | **Proration-dependent scenarios will demonstrate blockers, not amounts.** Per DECISIONS D5 + the open proration/B1/B2/X-FUND sub-questions, mid-period changes produce typed blockers. That's the correct discipline, but it silently narrows my scenario matrix items 2–4. | Split F1 into **F1a (blocker-path green)** and **F1b (computed-path green)** so gate approval states exactly what "green" means. Note it in POC-RUNBOOK too. |
| 7 | **MED** | **Experiment specs live outside the reviewed artifact set.** E1–E6, DBOS crash experiments 1–8, and harness requirements are cited but not in the gate package the human is approving. | Commit an `EXPERIMENTS.md` (specs + expected outcomes) so Phase 6 executes pinned specs, not memory. |
| 8 | **LOW-MED** | **Genesis pinning has an acceptance criterion but no procedure** (ADR-0204). First `GetState` bootstraps from the server — the one gap the agent itself flagged. | Sketch now: immediately post-create, export state (tx/hash/signature), anchor out-of-band (signed file on auditor-only storage; record in ADR). One paragraph; unblocks R2. |
| 9 | **LOW** | Hygiene: forward references to not-yet-existing `POC-RUNBOOK.md`/`POC-REPORT.md`; ownership and risk tables duplicated across PATTERNS/MIGRATION (declare one canonical); MIGRATION steps 2–5 should be labeled "post-gate adoption phase" to avoid reading as POC steps; vocabulary note — the repo's "Rung 5 = DB-enforced" (D4) is *not* the reliability ladder's rung 5 (formal verification); add one line to DECISIONS to prevent conflation. R1 addition: record the pinned version's **Change Date and Change License** (per-version Apache-2.0 conversion date) in the ADR. | Cosmetic edits at next revision. |

## Recommended responses to R1–R7 (you own these; my recommendation as reviewer)

- **R1 — APPROVED-WITH-CONDITIONS**: internal POC use is within BSL-1.1's grant; pin exact version, record Change Date + conversion date, keep the narrow replaceable interface, and treat any paid-hosted future as a new decision requiring counsel. Fallback = PG+DBOS (ADR-0201 alt 2) is genuinely viable.
- **R2 — APPROVED-WITH-CONDITIONS**: verifier state on separate host/volume from the writer; genesis pinned immediately post-create into an out-of-band anchor; no tamper-evidence claims until both are done (auditor stays unwired, as MIGRATION already says).
- **R3 — APPROVED-WITH-CONDITIONS**: rung retained; no truncate rights; DBOS `Rewind`/`DeleteWorkflow` disabled or guarded in POC config (they delete history — unusable on approved artifacts); `migration_admin` powers unchanged. Add finding #2's revoke assertions here.
- **R4 — APPROVED** as proposed: fresh 18.6 cluster, 16 volume preserved untouched, full supersede deferred to Phase 7 post-E5.
- **R5 — APPROVED-WITH-CONDITIONS**: Gate-2 direction on disposable DBs, no production migration — **and** resolve finding #4 (bake-off or recorded prior-authority note) as part of the approval.
- **R6 — decide SAME-DB for POC** (atomic path, shared pool), explicitly recorded as a POC-scoped exception requiring an AGENTS.md §8 amendment decision before any production adoption; keep the separate-DB replay variant as the designed fallback per finding #3.
- **R7 — approve in two tranches**: D1/D3/D4/D5/D7/D8 are internally complete and can be signed now; D2/D6 (and calc depth generally) stay PROPOSED until proration, B1, B2, X-FUND, and the cutoff sub-questions are answered — they gate F1b, not the gate itself.

## Bottom line

The package is coherent, honestly gated, and materially better than what my brief asked for at this stage — the license catch alone justifies the read-only phase. Fix findings #1 and #2 (both are concrete, bounded edits to ADR-0204 and ADR-0203) before human sign-off, since they change what the battery and the dual-write protocol *are*. Everything else can ride as conditions on the approvals.

Want me to produce the corrected ADR-0204 back-link section and the T13 battery rows as ready-to-paste patches, or write this review to a file (`GATE-REVIEW.md`) to sit alongside the package?


## chatgpt

# Adversarial architectural review — updated for Phase 6

My assessment has changed with your latest execution report.

Phase 6 represents substantial engineering progress. The project is no longer at the architecture-proposal stage. It has a compilable implementation, working synthetic payroll calculations, DBOS integration scaffolding, an immudb append path and database-backed test harnesses.

However, there is an important distinction:

Code implemented

Build and pure-Go evidence

Integration pending

Database-backed execution and adversarial verification

Production not authorized

Existing Gate 1 and business decisions remain binding

I would describe the current state as POC implementation substantially complete, with critical database integration and acceptance testing outstanding.

That wording matters. The unimplemented transaction-threading work and Gate-2 gate DDL are not merely environmental testing issues. They are remaining implementation work.

I also want to recognize an important result: the synthetic test vectors exposed and corrected an overtime calculation error. That is exactly the kind of defect discovery we expect from a credible engineering POC. It demonstrates why calculation specifications, executable golden vectors and deterministic results must evolve together.

The module rename and DBOS/pgx compatibility experiment also reduce two previously identified risks. I would preserve all of this work on `stack-poc`, without prematurely merging or committing it.



RESEARCH.md

MIGRATION.md



## 1. Findings ranked by architectural importance



Critical

&#x20;Finding 1 — The DBOS transactional integration is not finished



This is the most important unresolved engineering issue.

The architecture depends on an important DBOS guarantee: `RunAsTransaction` can atomically commit application writes and the corresponding DBOS durability record when the datasource shares the system database engine.

The current Go documentation explicitly describes this guarantee and specifies that the datasource engine is a `*pgxpool.Pool`. It also specifies a transaction callback receiving a DBOS transaction abstraction.&#x20;

[image](https://www.google.com/s2/favicons?domain=https://docs.dbos.dev\&sz=32)

DBOS Docs+2



That creates a specific integration requirement:

Your existing application services must execute their capture queries through the transaction supplied by DBOS, not independently acquire another connection from the pool.

Otherwise, the code may compile, the DBOS workflow may run, and the application's database writes may still fall outside the transaction that protects the durability checkpoint.

The unresolved Tx-threading refactor is therefore a fundamental acceptance blocker.

Required resolution:

Introduce or adapt a common database execution interface, where appropriate, so existing business services can operate with the DBOS transaction supplied by the workflow.

Prove that the entire critical capture operation uses that transaction:

- Employer gate acquisition.
- Claim and concurrency proof checks.
- Input reads.
- Population resolution.
- Snapshot header and child inserts.
- Rule-version pins.
- Audit/outbox writes.
- DBOS durability checkpoint.

Do not assume that passing the same PostgreSQL connection URL establishes transaction sharing. The current DBOS documentation describes sharing the same database engine handle with the system database. Verify this with the actual configuration and runtime behavior.

Also, ordinary `RunAsStep` and transactional steps must remain distinct. External immudb operations are still subject to at-least-once execution and require idempotency.&#x20;

[image](https://www.google.com/s2/favicons?domain=https://pkg.go.dev\&sz=32)

Go Packages+1





Critical

&#x20;Finding 2 — ADR-0204 has a potentially serious provenance-model defect



The proposed immudb record contains:

`immu_tx`

inside the canonical JSON value being written to immudb.

But the transaction ID is returned by immudb after committing the transaction. The application cannot ordinarily know that transaction ID when constructing the value inside that same transaction.

This is a circular reference.

Remove `immu_tx` from the payload. Store the returned transaction ID and ALH as the PostgreSQL back-link after successful verification.

There is another issue.

The proposed key:

`prov/{employer_id}/{artifact}/{artifact_uuid}`

is a stable key that can receive multiple versions. immudb retains the history of updates, but an ordinary `Set` can create a newer value under the same key. Its own documentation demonstrates repeated writes to a key and retrieval of earlier versions with `GetAt`.&#x20;

[image](https://www.google.com/s2/favicons?domain=https://immudb.io\&sz=32)

immudb+1



For an evidence ledger, I recommend event-oriented keys instead:

`prov/{employer_id}/{artifact}/{artifact_uuid}/{event_uuid}`

Every evidence event receives a unique key and a `MustNotExist` precondition.

The record should refer to the previous event through `supersedes`, rather than overwrite its predecessor.

There is one more subtlety: immudb's `MustNotExist` precondition treats logically deleted or expired entries as nonexistent. The idempotency namespace must therefore have explicit no-delete/no-expiration governance. Otherwise, the same idempotency key can become reusable.&#x20;

[image](https://www.google.com/s2/favicons?domain=https://immudb.io\&sz=32)

immudb+1



I consider these changes necessary before accepting ADR-0204 as an implementation-ready provenance design.



Critical

&#x20;Finding 3 — The threat model is not sufficiently bounded



D4 proposes retaining Rung 5 database-enforced protection.

I support retaining it, but we must define exactly what Rung 5 protects against.

There are different adversaries:

- An ordinary application user.
- An application runtime role with direct SQL access.
- A compromised application process.
- A database owner or `migration_admin`.
- A PostgreSQL superuser.
- An administrator with control over the host and persistent volumes.
- An attacker who compromises both the ledger writer and trusted verifier state.

PostgreSQL constraints and triggers can enforce important invariants against the application role. They cannot guarantee that a superuser who controls the database and its schema cannot disable those mechanisms.

Similarly, cryptographic ledger proofs cannot provide independent assurance if the same administrator can rewrite the ledger and every trusted copy of the auditor's state.

Rung 5 should therefore mean:

The invariant is enforced at the database boundary against all ordinary application writers, including direct SQL through the application role.

The stronger threat of compromised database administrators must be addressed through external evidence, privileged-role separation and independent verification.

Do not advertise protection against an adversary that the implementation has not actually excluded or detected.

This is exactly why R2 and R3 matter.



High

&#x20;Finding 4 — R1 through R7 and P1 through P5 need a formal crosswalk



The ADRs sometimes describe technical choices as selected while the decision-request register still identifies them as human approval requirements.

For instance:

- ADR-0201 selects the shared-pool DBOS topology.
- ADR-0203 selects Gate-2's transaction-scoped mechanism for the POC.
- R5 and R6 still request authorization for those respective decisions.

These are not necessarily contradictory architectural choices, because the ADRs explicitly restrict themselves to disposable POC databases. But their status must be made consistent.

The agent should create a decision traceability matrix showing:

`Existing binding constraint → ADR proposal → POC provisional choice → human authorization → verification evidence → production disposition`

P1–P5 must be included in that matrix. I have not received the contents of `POC-REPORT.md`, so I cannot independently determine whether those five provisional choices conflict with R1–R7.

Do not silently treat confirmation of P1–P5 as approval of D1–D8 or production decisions.



High

&#x20;Finding 5 — Your POC status reporting needs more precision



The execution report has four different kinds of evidence, which should never be conflated.

| Evidence category                             | Actual status |
| --------------------------------------------- | ------------- |
| Root Go module builds                         | Green         |
| Pure Go and synthetic calculations            | Green         |
| DBOS/pgx scratch interoperability             | Green         |
| DBOS application workflow runtime             | Not verified  |
| PostgreSQL integration and constraints        | Not executed  |
| immudb crash recovery and tamper verification | Not executed  |
| Complete one-month payroll run                | Not executed  |

There is also a subtle issue with reporting the unit suite as green. `core/database` tests currently fail when they attempt to connect to an unavailable database. Those tests should skip explicitly when no test database is configured, but must execute and fail normally in database-enabled CI.

I would also separate the two `poc-run` modes:

- `--demo-synth-only`: successful synthetic demonstration with a stable digest.
- Full POC mode: must fail loudly when PostgreSQL or other required dependencies are unavailable.

A synthetic demonstration must never be mistaken for a successful database-backed payroll run.



High

&#x20;Finding 6 — The research dependency versions need updating



I checked the official release information as of November 10, 2026.

PostgreSQL

18.6 current

# 18.6

Current PostgreSQL 18 minor release.

DBOS Transact Go

Research needs refresh

# 1.8.0

Latest release, October 28, 2026. The earlier research baseline is outdated.

immudb

1.11.2 latest stable

# 1.11.2

Use the stable release rather than a floating image tag or release candidate.

[image](https://www.google.com/s2/favicons?domain=https://www.postgresql.org\&sz=32)

PostgreSQL+2

The DBOS update is especially important. The agent should revisit the changes between the previously investigated version and v1.8.0, pin the version or exact commit, and run the scratch compatibility and crash tests against that exact build.

The existing DBOS interoperability experiment is useful, but its evidence must identify which source revision was actually tested.



High

&#x20;Finding 7 — PostgreSQL 18 temporal features are being positioned correctly, but their adoption should remain selective



The research has correctly identified that PostgreSQL 18 temporal constraints can help with:

- No-overlap requirements.
- Compensation-line containment.
- Effective-period coverage.
- Rule-pack validity across a payroll period.

However, the existing schema uses pairs of DATE columns rather than actual PostgreSQL range columns.

This means the migration requires an explicit range representation and constraint design. Do not assume a generated range column will work with every proposed temporal key without testing.

The official documentation confirms that `WITHOUT OVERLAPS` requires a range or multirange column and that a `PERIOD` foreign key requires an appropriate temporal unique or primary key.&#x20;

[image](https://www.google.com/s2/favicons?domain=https://www.postgresql.org\&sz=32)

PostgreSQL



My recommendation:

Keep the existing GiST exclusions during the initial POC. Prove E1–E3 independently, then adopt temporal constraints only where they give a concrete advantage.

Do not rewrite the compensation and SHG successor logic merely to take advantage of PostgreSQL 18.

The research itself acknowledges that their live-basis rules require separate enforcement.&#x20;

RESEARCH.md





High

&#x20;Finding 8 — The one-month payroll POC must prove business complexity, not infrastructure integration alone



The eight synthetic calculation vectors are a good start. They do not yet establish a complete payroll system.

I would require the actual database-backed month to demonstrate three different dimensions of complexity:

Business complexity: multiple employee profiles, approved attendance, varying compensation, leave, overtime, allowances, deductions and effective-dated rules.

Temporal complexity: a cutoff, mid-period changes, late arrivals, correction lineage, and period closure.

Operational complexity: concurrent writers, retries, worker crashes, recovery, approval and independent verification.

The project already correctly identifies Singapore payroll context, but its synthetic scenarios must not be presented as statutory calculations.

The existing `SG-MOM-OT-001` package is marked blocked and lacks its Go implementation. The POC should use explicit synthetic rules until actual regulatory calculations are independently researched and approved.&#x20;

RESEARCH.md



One particularly important improvement: reproducibility must pin not just inputs and rule-pack versions, but also the calculation implementation version or source/build digest.

Otherwise, an unchanged input snapshot and unchanged rate table could produce different results after a calculator upgrade.

The resulting provenance should identify:

- Input snapshot digest.
- Rule-pack content digest.
- Rule-pack identity and effective version.
- Calculator implementation/build digest.
- Canonicalized output digest.
- Approval and finalization references.

This is essential to the long-term extensibility objective.



High

&#x20;Finding 9 — The cutoff definition needs stronger protection against backdating



D1 proposes an employer-scoped cutoff timestamp and eligibility based on approval before cutoff.

The agent needs to verify the timestamp semantics.

For example, what happens when an approval transaction starts before the cutoff but commits afterward? What happens when an API client supplies an `approved_at` timestamp that claims the approval occurred before the cutoff?

A client-provided timestamp must not determine eligibility.

Use database-controlled event timestamps and the transaction-scoped employer gate to establish the authoritative ordering of approvals and close operations.

Also explicitly distinguish:

- Effective date: when the business fact applies.
- Recorded date: when the system recorded it.
- Approval time: when the authorized approval occurred.
- Cutoff time: the employer's payroll eligibility boundary.

This is especially important when historical changes are entered after the original payroll period.

## 2. Review of the immudb and DBOS integration pattern

The overall architecture is sound, but two details need to be elevated to explicit acceptance criteria.

### DBOS: distinguish transactional exactly-once from external at-least-once

The appropriate pattern remains:

PostgreSQL + DBOS transaction

Business capture and DBOS durability record commit atomically.

PostgreSQL transactional outbox

Evidence event committed with the business operation.

DBOS durable worker

At-least-once delivery, retry and recovery.

immudb immutable evidence

Idempotent append, independent verification and PostgreSQL back-link.

The distinction is essential: the transactional capture guarantee does not extend across the immudb network call.&#x20;

[image](https://www.google.com/s2/favicons?domain=https://docs.dbos.dev\&sz=32)

DBOS Docs+2



### immudb: independent verification must actually be independent

The proposed external auditor is the right direction.

However, placing the auditor state on another Docker volume on the same machine is not full independence against a host administrator who controls both volumes.

For the POC, explicitly state which attacker the tamper test simulates.

The genesis checkpoint and subsequent trusted roots must be retained outside the writer's control.

Also, the BSL-1.1 license deserves particular attention because this project has potential commercial SaaS implications. The current license permits certain production uses but contains a restriction on paid hosted or embedded competitive offerings. The exact application matters; this should receive legal review rather than an engineering assumption.&#x20;

[image](https://www.google.com/s2/favicons?domain=https://github.com\&sz=32)

GitHub+1



## 3. Human decisions: my recommended disposition of R1–R7

I would not ask you to blindly approve all seven requests. Some are legitimate business and legal decisions; others are engineering choices that should be supported by evidence.

| Request                    | My recommendation                                                                                                                            |
| -------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------- |
| R1 — immudb license        | Approve with conditions for internal synthetic POC only; no commercial product commitment without legal clearance.                           |
| R2 — trusted-state custody | Approve with conditions. Establish a separately controlled auditor and out-of-band genesis checkpoint.                                       |
| R3 — threat model          | Approve with conditions. Retain Rung 5 against ordinary application writers and explicitly define privileged administrator boundaries.       |
| R4 — PostgreSQL 18.6       | Approve with conditions for the disposable POC cluster; preserve PG16 and its volume until E5 and parity tests pass.                         |
| R5 — Gate-2 mechanism      | Approve with conditions for disposable POC experiments only; production Gate 1 remains pending.                                              |
| R6 — DBOS placement        | Prefer the same PostgreSQL database, `dbos` schema and shared pool handle for the atomic capture path. Require least-privilege verification. |
| R7 — business semantics    | Do not blanket-approve yet. Resolve the substantive business questions and explicitly document any POC-only synthetic assumptions.           |

These recommendations do not constitute your actual approvals. The decision register should retain HUMAN ownership until you authorize them.&#x20;

DECISION-REQUESTS.md



On R7, I would specifically preserve these decisions as principles:

D1: Explicit employer-scoped cutoff.

D2: Finalized payroll is corrected through linked adjustments, not overwritten.

D3: Clear separation of draft, captured, validated, approved, finalized and paid.

D4: Rung 5 protection for authoritative payroll artifacts.

D5: Only approved, sufficiently evidenced input records enter authoritative calculation.

D6: Preserve effective time, recorded time, affected period and adjustment period.

D7: One legal employer for the POC.

D8: Every approved/finalized run reproducible from pinned evidence.

What I would not approve without further definition is the missing active-evidence standard, SHG coverage semantics, X-FUND handling, proration, cutoff grace-period rules and payroll-line governance.

This prevents the agent from making business decisions simply to unblock a technical implementation.

## 4. Problems in the migration and adoption plan

The sequencing is generally good, particularly the separation of disposable POC schema from migration 00013. ADR-0202 is one of the strongest architectural decisions in the package.&#x20;

ADR-0202.md



I would nevertheless correct three things.

First: protect the PG16 environment.

Do not let the POC Compose project share volume names or project identity with the existing PG16 environment.

`docker compose down -v` must only target explicitly named disposable POC volumes. The existing PG16 volume must never be a candidate for deletion.

Second: distinguish POC schema from eventual production migrations.

The POC can use isolated SQL and a separate migration chain. Successful POC DDL must not automatically become production migration 00013.

Third: distinguish implementation rollback from data rollback.

Disabling a DBOS workflow or stopping immudb does not undo already-finalized business records. The rollback plan needs to preserve authoritative PostgreSQL state and prevent duplicate evidence or payment outputs when a feature is disabled.

MIGRATION.md

## 5. What I would authorize next

I recommend dividing the remaining work into three gates.

1. Gate A — Complete and reconcile Phase 6

   No database required.

   Fix transaction-threading, resolve provenance schema concerns, update version pins, fix database test skip logic, and reconcile P1–P5 with R1–R7.

   Preserve the existing uncommitted branch.
2. Gate B — Run the integration POC

   Requires Docker-capable infrastructure and relevant human decisions.

   Run PostgreSQL 18.6 and immudb with pinned image digests. Execute E1–E6, the translated attack battery, DBOS crash experiments, ledger tampering experiments and the complete payroll-month workflow.
3. Gate C — Decide adoption

   Only after Gate B.

   Decide whether PostgreSQL 18.6 supersedes PG16 for development, whether DBOS is retained, whether immudb's licensing and operational costs are acceptable, and what eventually belongs in production migration 00013.

   Production migration and the pending Gate 1 decision remain separate approvals.

This is preferable to asking the agent to implement more features before establishing that the existing integration works.

## 6. Next implementation agent prompt

This is the prompt I would issue now. It assumes the agent has access to the existing `stack-poc` working tree and Phase 6 implementation.

# Phase 6.1 — Adversarial Closure, Integration Completion and POC Execution

## Role

Act as a Principal Go Engineer, PostgreSQL Database Architect, DBOS Workflow Engineer and Adversarial Software Reviewer.

You are continuing an existing implementation. You are NOT starting a new project and must not repeat the research or implementation work already completed.

Your objective is to close the remaining architectural and engineering gaps, execute the one-month synthetic payroll POC on a Docker-capable environment, and establish whether the proposed PostgreSQL 18.6 + DBOS + immudb architecture actually works under adversarial conditions.

## 1. Current state: preserve this work

All Phase 6 changes are uncommitted on branch `stack-poc`.

Completed and reported:

- Root module renamed from `github.com/PithomLabs/a1` to `github.com/PithomLabs/a2`; 258 imports mechanically updated; `go build ./...` passes.
- DBOS/pgx scratch interoperability tested successfully with pgx v5.12.0.
- Nested POC module builds successfully using local-checkout replacements.
- Synthetic payroll calculator has eight passing golden vectors.
- Overtime denominator defect identified, corrected and covered by updated specification and vectors.
- DBOS capture and month workflow skeletons compile against real SDK APIs.
- immudb outbox appender and SQL schema exist.
- Database-dependent E1/E2/coverage/outbox tests are implemented but unexecuted.
- Synthetic demo produces a stable 25-employee × 31-day digest: `259564d6…`.
- No Docker daemon, PostgreSQL server or registry access was available during the preceding execution.
- Remaining known implementation work includes UNRESOLVED-4 (Tx-threading) and Gate-2 gate DDL.

The complete execution status is in:

- `poc/payroll-month/POC-REPORT.md`
- `poc/payroll-month/POC-RUNBOOK.md`
- `docs/stack/MIGRATION.md`
- `docs/stack/RESEARCH.md`
- `docs/stack/PATTERNS.md`
- `docs/stack/DECISIONS.md`
- `docs/stack/DECISION-REQUESTS.md`
- ADR-0201 through ADR-0204.

Read the actual repository instructions and these files before changing anything.

## 2. Non-negotiable boundaries

1. Stay on `stack-poc`.
2. Preserve all existing uncommitted work.
3. Do not commit, push, rebase or merge.
4. Do not modify migrations 00001–00012.
5. Do not create production migration 00013.
6. All POC schema additions remain under `poc/payroll-month/`.
7. Use synthetic data only.
8. Do not claim statutory, regulatory or production compliance.
9. Do not modify the existing PG16 database volume.
10. Do not use destructive commands against existing development databases.
11. Do not bypass existing database security boundaries to make tests pass.
12. Do not convert an unexecuted test into a reported pass.
13. Do not silently approve R1–R7 or change production architectural decisions.

If a human-only decision blocks a particular operation, document it and continue with independently permissible work.

## 3. First task: reconcile all decisions and implementation status

Inspect `POC-REPORT.md`, particularly provisional decisions P1–P5.

Build a decision traceability matrix:

`Binding constraint → D1–D8 → R1–R7 → ADR → P1–P5 → implementation → acceptance test`

Identify contradictions, duplicate decisions and assumptions.

In particular, clarify:

- Which P1–P5 decisions are strictly POC-scoped.
- Which require explicit human confirmation.
- Which can be established through experiments.
- Which must remain unresolved until business semantics are approved.
- Whether any POC choice accidentally implies production authorization.

Do not infer the contents of P1–P5. Inspect and quote their actual definitions.

Update documentation only when the proposed correction is supported by evidence.

## 4. Refresh exact technology versions

As of November 10, 2026, the research baseline should be reconciled against:

- PostgreSQL 18.6.
- DBOS Transact Go v1.8.0.
- immudb v1.11.2.

Verify the exact releases and source revisions, including release-specific changes that could affect transaction or replay semantics.

Do not use floating `latest` tags.

For Docker images, record the exact image tag, architecture and OCI registry digest. Never invent a digest or copy one from an unrelated platform.

DBOS is a Go dependency rather than a separate required application container.

Inspect the actual dependency versions in the POC's go.mod and any local `replace` directives.

Replace machine-specific dependency assumptions with a reproducible version or commit pin where appropriate. Ensure a fresh checkout can build without relying on undocumented sibling directories.

## 5. Critical implementation task: complete DBOS transaction threading

Resolve UNRESOLVED-4.

The existing application services use PostgreSQL pools, while DBOS `RunAsTransaction` supplies its own transaction abstraction.

Investigate the current v1.8.0 SDK implementation and its official datasource semantics.

Implement the smallest clean refactor that allows the existing business operations to execute through the transaction supplied by DBOS.

Do not obtain an independent pool connection inside the atomic capture callback.

Verify that the DBOS system database and application datasource use the same actual pool handle for the capture path, not merely the same connection URL.

Preserve the existing service and repository boundaries.

The atomic capture transaction must contain:

1. Employer-scoped gate acquisition.
2. Transaction-scoped claim and proof checks.
3. Population validation.
4. Approved-input resolution.
5. Compensation and rule-pack coverage validation.
6. Canonical snapshot construction.
7. Snapshot header and child insertion.
8. Rule-version pins.
9. Transactional audit/outbox insertion.
10. DBOS durability checkpoint.

Make transaction ownership explicit.

Demonstrate through execution that a process crash before commit leaves no partial capture and a crash after commit allows recovery without duplicate capture.

Do not claim atomicity merely because the application compiled.

## 6. Gate-2 disposable database implementation

Complete the Gate-2 transaction-scoped employer gate and claim proof-check on POC-only databases, subject to the documented POC authorization.

Do not pre-empt the production Gate 1 ruling.

Do not reintroduce the retired XID-ticket path.

Use the approved transaction isolation level and lock-order protocol.

Translate the existing writer matrix into executable database enforcement.

All critical attack tests must run under the real `app` role, never as superuser.

Verify privileges, triggers and DEFINER function ownership.

Do not weaken Rung 5 to make the POC easier.

## 7. Adversarial correction of immudb provenance

Review ADR-0204 against the actual v1.11.2 API.

Correct the following design issues before accepting the ledger schema:

**A. Transaction identity**

Remove `immu_tx` from a value that is written in the same immudb transaction. The actual transaction ID is learned after commit and must be recorded as a PostgreSQL back-link after successful verification.

**B. Immutable artifact identities**

Replace reusable provenance keys with event-specific unique keys.

Recommended shape:

`prov/{employer_id}/{artifact}/{artifact_uuid}/{event_uuid}`

Every correction or successor must create a new evidence event, with an explicit `supersedes` reference.

**C. Idempotency**

Use deterministic event and idempotency identifiers.

Protect the provenance event key and idempotency marker with appropriate existence preconditions.

The idempotency value must bind the event identity to the expected content digest. A retry with an identical request must recover the original result; an identical key with a different payload must fail loudly.

Account for the fact that logical deletion and expiry affect `MustNotExist` semantics.

No ledger idempotency key may be deleted, expired or silently reused.

**D. Dual-write recovery**

Prove all important failure boundaries:

- PostgreSQL commits but ledger append has not started.
- immudb commits but the application loses the response.
- The immudb transaction exists but PostgreSQL back-link is missing.
- Reconciliation encounters an already-appended event.
- immudb is unavailable.
- PostgreSQL is unavailable.
- An outbox event is delivered more than once.

There must be no lost evidence, duplicate financial postings or silent inconsistencies.

**E. Sensitive information**

Do not assume that hashes-only means no personally identifiable or linkable information.

Review actor IDs, employer IDs, artifact identifiers and all metadata for privacy implications.

Keep PII, NRICs, bank details, secrets and unnecessary employee identifiers out of immudb.

**F. Independent verification**

Specify and execute the genesis-pinning and trusted-state procedure.

The independent auditor's trust anchor must be outside the writer's mutable control.

A second Docker volume on the same host is not sufficient evidence of independence against a hostile host administrator.

Do not claim tamper resistance beyond the threat model actually tested.

## 8. Fix test and command semantics

Resolve the pre-existing `core/database` test skip-logic gap.

Tests without an explicitly configured database must skip cleanly where appropriate.

Database-enabled CI must require the database environment and fail if it is missing or unavailable.

Do not make CI green by silently skipping database tests.

Separate the POC commands:

- Synthetic demonstration mode must have a successful, documented exit code.
- Full database-backed POC mode must fail loudly when required dependencies are missing.
- Full database-backed mode may return success only after all mandatory acceptance checks have passed.

Ensure `go test`, `go vet`, `gofmt`, `task build`, `task test` and `task lint` operate correctly with the repository's Go workspace.

Do not weaken assertions or suppress meaningful failures.

## 9. Docker-capable host: execute the integration plan

Use the runbook and the verified image digests.

Use a fresh PostgreSQL 18.6 cluster and a standalone immudb 1.11.2 instance.

Keep PG16 available for parity testing and rollback.

Ensure the POC Compose project has isolated names and volumes. A teardown must never delete the existing PG16 database.

Run the database-backed E1–E6 experiments specified in RESEARCH.md.

Run the full 00001–00012 migration chain against PostgreSQL 18.6.

Run the PG16/18 parity exercise using canonical snapshot bytes and the existing regression suite.

Execute:

- All translated T1–T12 attacks.
- Both writer-first and capture-first interleavings wherever specified.
- The eight DBOS crash/recovery experiments.
- The five PG-to-immudb dual-write crash boundaries.
- The immudb tamper battery.
- Independent auditor verification.
- Outbox reconciliation.
- Fresh-process restart and workflow recovery.

All attack tests must assert persisted state, not merely inspect error messages.

Use deterministic test interleavings.

Record complete command output and test results.

## 10. Execute the complete one-month payroll POC

Do not stop at successful container startup or a passing unit suite.

Run the real end-to-end monthly payroll fixture against PostgreSQL 18.6.

The demonstration must prove:

- Employee and employment records.
- Approved attendance and timesheets.
- Compensation resolution.
- Effective-dated rule selection.
- Overtime.
- Leave and absence.
- Allowances and deductions.
- Payroll calculation.
- Exceptions and blockers.
- Approval.
- Finalization.
- Period closure.
- Immutable provenance recording.
- Independent re-verification.
- Reproducible calculation output.

Respect the outstanding business decisions.

Do not invent missing proration or statutory semantics to produce a successful result. Ambiguous mid-period changes must block where the business policy is not settled.

Require exact financial arithmetic, explicitly defined rounding and timezone semantics, deterministic canonical input/output representations and pinned calculator versions.

Generate:

- Payroll register.
- Employee payslip calculations.
- Exception report.
- Payment-file preview.
- Period-close report.
- Evidence verification report.

Verify that the original finalized output remains unchanged after a linked post-finalization adjustment.

## 11. Required final report

Update `poc/payroll-month/POC-REPORT.md` and the relevant implementation documentation.

Report separately:

**Implemented:** actual code present.

**Built:** compilation verified.

**Executed:** command actually ran.

**Passed:** assertion successfully satisfied.

**Failed:** reproducible defect.

**Blocked:** external dependency or human decision prevents execution.

**Unexecuted:** test or experiment not run.

Include:

- Exact dependency and image versions.
- Exact image digests.
- Commit and working-tree status.
- Database migration results.
- E1–E6 outcomes.
- T1–T12 outcomes.
- DBOS crash matrix.
- immudb tamper and recovery matrix.
- Full payroll POC results.
- Remaining unresolved decisions.
- Any changes to architectural assumptions.
- Exact next action and its responsible owner.

Do not call the POC successful unless the actual full-month database-backed run and mandatory adversarial tests pass.

## 12. Definition of completion

Phase 6.1 is complete only when:

1. The module builds reproducibly.
2. The DBOS integration uses the intended atomic transaction boundary.
3. Gate-2 enforcement works under the actual application role.
4. PostgreSQL 18.6 migration and parity tests pass.
5. The translated adversarial battery passes.
6. DBOS recovery tests pass.
7. immudb provenance is idempotent and independently verifiable.
8. The full one-month payroll fixture runs successfully.
9. Historical payroll outputs are reproducible.
10. The project remains within its approved POC boundaries.

**Final instruction:**

Do not optimize for the appearance of completion.

Prioritize correctness, reproducibility and adversarial verification over feature count.

Preserve the existing working branch. Do not make production changes, do not create migration 00013, and do not silently resolve human-owned decisions.

The deliverable is an executable, independently testable demonstration that the proposed architecture materially improves the existing platform.

## 7. My final recommendation

There is one thing I would change about the overall direction.

Do not make the immediate next milestone "integrate three technologies."

Make it:

Prove that one payroll month can survive adversarial execution, produce immutable historical results, and be independently reproduced.

That is a much stronger architectural achievement.

And I would not let the addition of PostgreSQL 18.6, DBOS or immudb obscure the original objective: a highly configurable workforce-management and payroll platform that can accommodate new employers, policies, effective-dated instructions and complex business rules without repeated architectural rewrites.

The technology allocation is promising. The next challenge is proving the invariants.

My verdict: proceed to Phase 6.1, but do not yet approve production adoption or merge the branch.

