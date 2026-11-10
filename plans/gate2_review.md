Gate 2 review — Option B

Synchronization architecture: accepted
 Production migration: not yet approved

The Gate 2 experiment makes a convincing case for replacing Option A with Option B. However, I found several gaps in the supplied Option B SQL that invalidate parts of the report's broader authority and validation claims.

My recommendation is:
- Retain the ADR's decision to select READ COMMITTED with a transaction-scoped employer gate.
- Retire the pre-BEGIN session-lock/XID-ticket approach as the production candidate.
- Do not yet authorize production migration 00013.
- Run a focused Gate 2.1 correction of the snapshot-authority, lifecycle and phase-marker issues below.
This is not a rejection of the experiment. It is a distinction between proving a synchronization mechanism and proving that the complete payroll decision cannot be fabricated or bypassed.
1. What Gate 2 successfully demonstrated
Option A should be retired
The new attack tests are materially stronger than the earlier Gate 1 evidence. They demonstrate both the pre-BEGIN DISCARD ALL failure and the more important post-verification release window.

In particular, the test shows a caller can release a session lock after ticket verification, allow a competing writer to commit, and then commit a stale RR read set without the ticket mechanism detecting the loss of protection. 
gate2_test.go

That is sufficient reason not to carry Option A into the production design.
Option B is the right synchronization direction
The experiment demonstrates the essential mechanism:
- A single employer gate is shared across capture and writer paths.
- Writers that acquire it first complete before capture reads.
- Writers that arrive after capture acquires it wait until the decision commits.
- The same mechanism works through direct DML, trigger paths and privileged writer functions in the fixture.
- Transaction rollback and backend termination release the gate automatically.
The writer-first and capture-first tests inspect the frozen snapshot contents rather than merely asserting that locks were acquired. That is the right test design. 
gate2_optionb_test.go

The experiment also tests the important same-transaction attack instead of assuming a transaction-scoped advisory lock automatically prevents a caller from writing after capture. 
gate2_optionb_test.go

I accept Option B as the selected synchronization architecture.
2. Blocking findings in the supplied implementation
These are not hypothetical concerns about future production code. They are visible in the supplied experimental SQL.
P0 — The canonical hash is not bound to the verified claims
gate2_capture performs these operations:
1. Hashes the caller-supplied p_canonical string.
2. Compares the resulting digest to the caller-supplied p_hash.
3. Independently verifies p_claims against live database rows.
4. Stores the supplied hash and creates snapshot children from p_claims.
It never establishes that p_canonical actually represents p_claims. 
gate2_optionb_test.go

The existing forged-hash test changes an input payload to a value that does not exist in the authoritative source. The database correctly rejects that. 
gate2_optionb_test.go

But consider a different attack:
- Obtain genuine claims that exactly match current authoritative inputs.
- Supply an unrelated canonical string.
- Compute a perfectly correct SHA-256 of that unrelated string.
- Pass both to gate2_capture.
The hash check succeeds. The claims check succeeds. The snapshot is created with a hash that does not identify its actual contents.

gate2_validate reconstructs claims from the frozen snapshot children and checks them against live records, but never checks the stored hash against those children or the canonical representation. 
gate2_optionb_test.go

Required correction: Add an adversarial test using valid, unchanged claims and a deliberately unrelated canonical string with its correctly computed hash. It must not validate.

Then decide how the Go-canonical representation is cryptographically or structurally bound to the verified snapshot without introducing two competing canonical implementations.

This is a blocker because ADR-1 explicitly retains canonical hashing as part of snapshot authority.
P0 — A fully populated planted snapshot can bypass the capture function
The fixture grants app INSERT on:
- gate2_snapshot
- gate2_snapshot_child
- gate2_snapshot_pack
- gate2_validation_attempt
- gate2_validation
It also grants SELECT on these tables. 
gate2_optionb_test.go

The existing planted-snapshot test inserts only a header with no children and confirms that validation rejects it because claims are missing. 
gate2_optionb_test.go

That does not establish that planted snapshots are impossible to validate.

The supplied SQL allows this more complete attack:
1. Insert a snapshot header directly as app.
2. Insert children copied from every actual authoritative input.
3. Insert the applicable pack pins.
4. Set an arbitrary content_hash.
5. Call gate2_validate.
6. The verifier rebuilds claims from those children, finds that the rows match the live authoritative data, and proceeds.
There is no requirement in gate2_validate that the snapshot was actually created by gate2_capture.

The fact that an empty planted snapshot is rejected does not prove that a complete planted snapshot is rejected.

Required correction: Revoke direct INSERT on the snapshot, child, pack, validation-attempt and validation tables from app. The approved SECURITY DEFINER operation should own creation of these artifacts. Then test that direct inserts fail and that the approved capture path still works.

The attempt marker should not need caller-writable permissions to function.
P0 — Validation does not establish that the snapshot belongs to the run being validated
In gate2_validate, the snapshot is looked up using:
- Snapshot ID.
- Employer ID.
- Snapshot status.
It is not checked against p_run. The function then independently updates the supplied run ID. 
gate2_optionb_test.go

That permits a caller to attempt:
SELECT gate2_validate(
    employer_id,
    run_b,
    snapshot_belonging_to_run_a
);


Provided the snapshot contents match the current authoritative inputs, the function can validate run_b using a snapshot belonging to run_a.

This is a broken business relationship, even if the content itself is genuine.

Required correction: Require the snapshot's run_id and employer_id to match the exact run being validated. Enforce the relationship at the database boundary and add a negative test with two draft runs belonging to the same employer.
P0 — The run lifecycle can be bypassed during INSERT
The test fixture grants app INSERT on gate2_run. Its lifecycle trigger is registered for UPDATE OR DELETE, not INSERT. The table permits both draft and validated as status values. 
gate2_optionb_test.go
 gate2_optionb_test.go

Consequently, a direct insert supplying status = 'validated' is not checked by that trigger.

The current fixture does not test this.

There is a related production concern: the report identifies the existing application role as having full DML on payroll runs. A trigger that permits draft → validated without checking the protected validation boundary would not be enough if direct application updates are allowed.

Required correction:
- New runs must be created only in draft.
- Application code must not directly set a run to validated.
- The approved validation function must be the authoritative transition path.
- The transition must prove that a corresponding snapshot and successful validation decision exist.
- Add negative tests for forged initial status, unauthorized update and validation without a matching snapshot.
These four findings mean the current authority and lifecycle evidence is not sufficient for production acceptance.
3. Additional correctness and coverage findings
P1 — The own-transaction marker loses the XID epoch
The phase-marker trigger uses:
txid_status(s.xmin::text::bigint) = 'in progress'


The report describes this as having no XID casts or epoch games, but that is not what the SQL implements. 
gate2_optionb_test.go

PostgreSQL documents that:
- xmin uses the 32-bit xid type.
- Transaction IDs wrap around.
- xid8 carries the epoch and does not wrap around during the installation's lifetime.
- txid_status(bigint) is the older interface corresponding to pg_xact_status(xid8). 
  
  postgresql.org+1
Converting the textual representation of xmin to bigint does not restore the missing epoch.

The marker must not depend on that conversion being correct indefinitely. This is particularly important because failure to recognize the marker could allow a same-transaction write after capture has entered its protected phase.

Required correction: Store the top-level transaction identity explicitly as xid8 when creating the protected decision marker. PostgreSQL documents that pg_current_xact_id() returns the top-level transaction ID even when called in a subtransaction. Use that full identity in the marker test and verify its behavior under savepoints. 

postgresql.org+1

Do not graduate the current xmin cast into production unchanged.
P1 — Production writer coverage is not complete
The report identifies 18 production snapshot read-set tables but the Option B experiment models only three input tables and one pack table. It explicitly acknowledges that per-table employer derivation and trigger coverage remain unproven. 
gate2_synchronization_experiment.md

More importantly, the report already identifies a concrete gap:

employee_status_history is part of the snapshot read set, has no trigger, and permits direct application inserts.

That is not merely a future maintenance concern. It is a current write path to an authoritative snapshot input that does not participate in the proposed employer gate.

A direct insert could therefore commit during the production capture decision without acquiring the gate.

Required correction: Treat every table in the actual production read set as part of the acceptance boundary. The production migration must either install and prove the appropriate gate participation or establish that a table cannot change through any permitted runtime path. The known employee_status_history gap must be closed.

Do not describe the nine fixture writer-operation classes as proof that all nine or all eighteen production input tables have been covered. They are different dimensions of the test matrix.
P1 — The decision and audit tables have excessive application write privileges
The fixture gives app direct INSERT privileges on both the validation-attempt and validation tables, as well as snapshot-related tables. 
gate2_optionb_test.go

This can produce records that look like legitimate operational evidence but were not created by the authorized workflow.

For example, a direct INSERT into gate2_validation can create a valid verdict without the controlled validation function. That may not change the payroll-run status, but it contaminates the validation record and can mislead downstream readers.

The attempt marker is similarly an operational record, not ordinary application input.

Required correction: Make creation of snapshots, their children, validation attempts, validation verdicts and related decision artifacts exclusive to the controlled database functions. The normal application interface should have read access where appropriate and narrowly scoped EXECUTE permissions—not arbitrary INSERT rights.

The phase marker must be produced by the same controlled decision boundary that owns its meaning.
P1 — Multi-employer transaction behavior needs an explicit contract
The report correctly records that transactions accumulating multiple employer gates in opposite orders can deadlock and that PostgreSQL returns 40P01. 
gate2_synchronization_experiment.md
 gate2_optionb_test.go

For the current one-employer MVP, this is manageable. But the proposed workaround must be database-enforceable if direct SQL remains inside the threat model.

A Go helper that sorts employer IDs does not constrain a direct SQL caller.

The production decision should state one of the following:
- Cross-employer writes in a single transaction are rejected by an enforceable database invariant.
- Or multiple employer gates are acquired in a canonical global order, with the full multi-employer transaction tested.
At minimum, callers must handle the deadlock as a complete-transaction retry, never a partial retry of the failed statement.
4. What I would retain from the report and ADR
I would preserve these decisions:
Area	Review decision
Option A	Retire from production candidacy
Option B synchronization	Adopt as selected architecture
Isolation level	READ COMMITTED, explicitly checked
Employer gate	One canonical transaction-scoped key
Writer participation	Database-enforced
Same-transaction protection	Retain the own-row phase-marker concept, but correct its XID implementation
ADR-1	Keep Go as the single canonical resolver and serialization implementation, subject to hash-binding resolution
Production writer matrix	Expand to the actual snapshot read set
Historical evidence	Preserve
OT-001	Remain BLOCKED
Production migration 00013	Not authorized yet

The ADR correctly distinguishes its architecture decision from production-migration authorization. 
gate2_synchronization_adr.md

I would retain that distinction.
5. Why the report's GO statement needs narrowing
The report says all ten GO criteria pass and recommends Option B for production implementation. 
gate2_synchronization_experiment.md

I would distinguish three different conclusions:
Decision	My verdict
Option A is unsuitable for production	Accepted
Option B is the correct synchronization architecture to pursue	Accepted
Option B's complete snapshot authority and lifecycle implementation is production-ready	Not demonstrated

The 79 reported green checks are meaningful evidence for the experiment. I have reviewed the supplied report, SQL and Go test source, but have not independently rerun those tests against the repository or PostgreSQL. More importantly, the uncovered issues follow directly from the supplied SQL; additional green tests cannot resolve them without implementation changes.

The report's own production limitations support this narrower interpretation. It defers the full production trigger rollout, per-table derivation, deadlock handling, marker performance and the final acceptance battery. 
gate2_synchronization_experiment.md

6. The focused Gate 2.1 work I recommend
Do not reopen the Option A versus Option B architecture debate. That decision has enough evidence.

The next work should be limited to closing these findings:
Priority	Required test or correction	Acceptance condition
P0	Valid claims with unrelated canonical bytes and matching hash	Rejected, or a proven authority model binds hash to claims
P0	Fully populated, directly planted snapshot	Cannot validate
P0	Snapshot from Run A used to validate Run B	Rejected
P0	Direct INSERT of a run with status='validated'	Rejected
P0	Direct UPDATE of a draft run to validated	Rejected outside the authorized validation boundary
P1	Replace the truncated xmin-to-bigint marker check	Correct top-level xid8 semantics, including savepoints
P1	Direct update of employee_status_history	Participates in the employer gate or is structurally blocked
P1	Direct INSERT into validation and decision records	Rejected
P1	Multi-employer transaction	Defined, enforced and tested lock-order behavior

After these changes, repeat the relevant adversarial tests against the actual production schema and role privileges, not only the isolated fixture.

The snapshot-authority tests should also explicitly demonstrate that the hash identifies the actual frozen child content—not merely that the submitted hash matches some caller-supplied byte string.
7. Final architectural disposition
Option B remains the selected architecture
Preserve:
- READ COMMITTED.
- One transaction-scoped employer gate.
- Database-enforced writer participation.
- Own-transaction phase protection.
- Go-canonical compensation resolution.
- Database verification of authoritative claims.
Hold:
- Production 00013.
- Production capture/validation wiring.
- Any claim that the snapshot authority boundary is already complete.

The important distinction is that Gate 2 solved the synchronization architecture decision, but it did not finish the entire payroll snapshot integrity problem.

I would therefore amend the ADR's status to:
Option B selected as the production synchronization architecture. Production implementation remains blocked pending snapshot-authority, run-lifecycle, phase-marker and full writer-coverage acceptance.

That preserves the good architectural progress without promoting an experimental fixture into a production guarantee.

My recommendation is a narrowly scoped Gate 2.1 — Snapshot Authority and Lifecycle Closure, followed by a production-specific implementation prompt. No reason to redo the 79-check experiment or reopen the synchronization choice.