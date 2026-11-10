I would position this as a technical feature article rather than a project report or an architecture tutorial. The story should begin with the problem, take readers through the discoveries, explain why the architectural choices matter to both engineers and business owners, and finish with lessons that readers can apply to their own systems.

I have grounded the writeup in the completed Prompt8 remediation report and final POC results. I have deliberately left out the detailed adversarial findings: readers can discover those themselves by exploring the repository.

# When Payroll Fails, Who Remembers What Happened?

## What building a small payroll experiment taught us about PostgreSQL, durable execution, and cryptographic evidence

A payroll system has a peculiar responsibility: it must remember not only what it calculated, but why it calculated it, which rules it applied, who authorized the result, and whether that result can still be trusted after something goes wrong.

That sounds like a database problem. It isn't entirely.

Consider a seemingly ordinary monthly payroll run. An employee changes compensation halfway through the month. A timesheet arrives late. A manager approves an adjustment after the payroll process has already started. Two operators accidentally initiate the same run. Somewhere in the middle, a worker crashes. The application restarts, and someone asks whether it can resume without duplicating payments or silently changing the result.

Months later, an auditor questions a finalized payslip. The company must reconstruct the original calculation, identify the rules that applied at the time, and establish whether the historical record has been altered.

Each problem belongs to a different architectural concern. Some involve the integrity of business data. Others concern the execution of long-running processes. Still others concern the trustworthiness of historical evidence.

The mistake is to assume that one database, however sophisticated, automatically solves all three.

We set out to explore that distinction through a deliberately limited experiment: one synthetic monthly payroll, built in Go, backed by PostgreSQL, orchestrated with DBOS, and anchored to immudb.

We were not trying to build another payroll product. We wanted to understand what happens when a software team assigns different responsibilities to technologies designed to handle them, rather than implementing every mechanism itself.

The result was more revealing than we expected.

---

## The experiment was deliberately small

The workload consisted of one legal employer, 25 synthetic employees, and a 31-day payroll period covering October 2026.

The data included employment terms, compensation, attendance, timesheets, leave, synthetic overtime, allowances, deductions, and two effective-dated rule packs.

One pack applied during the first half of the month; another applied during the second half. This gave us a way to exercise rule changes without building an entire statutory payroll engine.

Some cases were deliberately allowed to calculate. Others were deliberately blocked because their business semantics had not been established. In particular, we did not invent a proration rule for an unresolved intra-month compensation change.

That distinction matters. A system that silently calculates an ambiguous entitlement may look productive, but it is not trustworthy.

The final October demonstration produced 102 payroll lines, 24 payslips, 24 finalized employees, one blocked employee, and five exceptions. The net payroll was 8,521,857 cents.

The database-backed calculation and its canonical counterpart produced the same digest:

`4beee632a05c09b9ec0be0dbedc1937d517461603d8660999d66f51d04b20377`

More important than the numbers was the path taken to produce them.

The calculation came from captured database records. Results were persisted in PostgreSQL. The execution survived a forced process termination during approval. Finalized evidence was sent to immudb, and an independent auditor reconstructed and verified the resulting content.

Fourteen remediation acceptance conditions passed, and the final Go test suite exited successfully. These are POC results, not a certification of production payroll compliance. Singapore's statutory overtime package remained explicitly blocked.

The limited scope was intentional: enough complexity to expose architectural problems, but not so much functionality that the experiment became a product-development programme.

---

## PostgreSQL taught us that storing data and governing data are different jobs

Most application developers initially think of PostgreSQL as the place where business data lives.

That description is incomplete.

In a serious enterprise application, PostgreSQL must also help determine which business states are permitted to exist.

Consider compensation records. An employee should not accidentally have two overlapping compensation arrangements governing the same effective period. There should not be an unexplained gap where the application cannot determine which compensation arrangement applies.

An application can check for these conditions before inserting records. But application-side checks are vulnerable to concurrency.

Two requests can arrive almost simultaneously. Each reads the database, sees no conflict, and concludes that its proposed change is valid. Both then attempt to write.

The problem is not that either request performed an incorrect calculation. The problem is that the business invariant was never protected at the authoritative data boundary.

PostgreSQL provides mechanisms for expressing such invariants through constraints, transactions, range types, exclusion constraints, triggers, and privileges.

Our experiment demonstrated the value of `WITHOUT OVERLAPS`: a deliberately unconstrained table accepted two overlapping records, while the constrained version rejected the conflicting insertion.

But an interesting lesson emerged.

A constraint is not a complete business policy. Preventing overlap does not automatically guarantee complete temporal coverage. Preventing an invalid foreign key does not prove that a payroll snapshot contains every employee who should have been included.

The database must enforce the rules that actually define valid business state, not merely the easiest relationships to express in SQL.

The more important architectural change was moving authority away from unrestricted application writes.

The final POC uses narrowly scoped database functions for operations such as capturing, validating, approving, closing, and publishing payroll artifacts. The application runtime can read data and invoke authorized operations, but it cannot freely manufacture authoritative payroll records.

This is a meaningful change in how we think about database design.

Instead of asking, "Can the application write this record?" we begin asking, "Under what conditions is the system authorized to recognize this record as a business fact?"

For business owners, the distinction is consequential. A database that merely stores payroll results is an archive. A database that enforces the conditions under which payroll results become authoritative is part of the organization's control system.

---

## DBOS revealed that business processes have a memory problem

The next discovery concerned something less visible: the lifetime of business processes.

A typical Go application can execute a payroll job using a background worker. The worker reads data, performs calculations, writes results, and reports completion.

That approach works until the process is interrupted.

Suppose payroll has completed its calculations but is waiting for managerial approval. The application crashes. When it restarts, what does it know?

It must determine which operations completed, which are still pending, whether the approval was received, whether a previous attempt partially wrote data, and whether restarting the process would duplicate an existing result.

One approach is to build all of this ourselves.

We could create workflow-state tables, checkpoint records, worker ownership leases, retry policies, fencing mechanisms, deduplication tables, and recovery coordinators.

None of these mechanisms is particularly mysterious. The difficulty is making them work together under every failure sequence.

DBOS offers a different approach: represent long-running business processes as durable workflows whose execution history is persisted.

In our implementation, DBOS coordinates the payroll lifecycle through several stages, including readiness, capture, calculation, persistence, approval, validation, and evidence publication.

The critical detail is its transaction boundary.

DBOS can coordinate application database writes and its own durability record in an atomic PostgreSQL transaction when using the appropriate shared datasource and transactional execution mechanism.

That means the application does not have to guess whether a business operation committed while its corresponding workflow checkpoint did not.

We tested this by forcibly terminating execution before the capture transaction committed. The transaction rolled back without leaving a successful capture checkpoint. After restart, the workflow converged on one authoritative result.

We also tested an interruption during approval. The workflow recovered, completed approval and validation, and produced the expected result without duplicating the payroll artifacts.

This changed our understanding of DBOS.

It is not simply a job queue or a convenient way to execute background functions. Its more important role is to provide durable memory for business processes.

For an application involving employee onboarding, leave approvals, shift changes, timesheet disputes, payroll, and corrections, this capability can be reused across many business workflows.

It allows engineers to concentrate more on the actual business process and less on reinventing the machinery required to keep that process alive.

But durability is not the same as correctness.

DBOS cannot decide whether an overtime rule is valid, whether a salary adjustment is authorized, or whether an employee belongs in a particular payroll period. Those remain responsibilities of the domain model and PostgreSQL.

Nor does a durable workflow automatically guarantee exactly-once execution of an external service.

A transaction can atomically commit database changes and a checkpoint. An external network operation still needs its own idempotency and recovery strategy.

That distinction became central to our implementation.

---

## immudb introduced a different question: can we trust the history?

Even a carefully designed PostgreSQL database has a trust boundary.

A runtime application role may have restricted permissions. Triggers may prevent unauthorized changes. Audit tables may record important events.

But what happens when someone has administrative control over the database itself?

Ordinary audit records are stored in the same system whose history they are supposed to describe.

An administrator with sufficient privileges may be able to change both a business record and its corresponding audit history.

This is not an argument that PostgreSQL is insecure. It is a recognition that operational data and independently verifiable evidence serve different purposes.

We introduced immudb to explore that distinction.

immudb maintains an append-oriented transaction history with cryptographic verification mechanisms. Rather than use it as a second operational database, we treated it as an evidence layer.

When a payroll result is finalized, PostgreSQL creates an evidence event containing references and cryptographic digests associated with the actual business artifact.

A transactional outbox bridges the two systems.

The PostgreSQL transaction records the business operation and its evidence event together. A durable worker subsequently publishes that event to immudb. Once the ledger append is verified, PostgreSQL records the corresponding transaction reference.

There is no magical distributed transaction covering both databases. The application must handle the possibility that the ledger append succeeds while the PostgreSQL back-link fails, or that a network response is lost even though the append committed.

Deterministic event identities, idempotency markers, verified reads, and reconciliation are therefore essential.

The other important component is the auditor.

We deliberately separated the auditor from the writing application. It uses a read-only PostgreSQL identity, a read-only immudb identity, and a pinned genesis checkpoint.

It reconstructs the expected evidence from the database, re-executes the relevant payroll calculation, and compares the resulting content against cryptographically verified ledger records.

This is the distinction between verifying that a transaction exists and verifying that the transaction actually represents the payroll result being audited.

A cryptographic ledger does not prove that the calculation was correct. It helps establish that the evidence being examined matches what was previously anchored.

That is a narrower claim, but a much more useful one.

For businesses, the potential value lies in historical accountability: being able to establish what was finalized, which rules were associated with the result, and whether the evidence has changed.

---

## The most important insight was not technological. It was architectural.

At the beginning, it was tempting to think of the project as a combination of three technologies.

We eventually found a more useful way to describe it: three different responsibilities that happen to require different technical mechanisms.

PostgreSQL answers:

**What is the authoritative business state?**

DBOS answers:

**How does a business process continue correctly when execution is interrupted?**

immudb answers:

**Can we independently verify the historical evidence associated with that business state?**

Go remains responsible for the domain logic and calculations.

These responsibilities must cooperate, but they should not be confused.

A workflow should not become the authority for a payroll record simply because DBOS says it completed. A ledger should not become the authority for whether a salary calculation is correct simply because the transaction is immutable. A database constraint should not be mistaken for a complete interpretation of employment legislation.

This separation gives us an architecture in which every component has a clear job and an explicit boundary.

It also reveals an important business consideration: adding technologies does not automatically reduce complexity.

DBOS introduces workflow lifecycle and resource-management considerations. immudb introduces another persistence system, trust management, licensing considerations, and cross-system recovery. PostgreSQL still requires deliberate schema design, privilege management, and transaction discipline.

The value comes from eliminating responsibilities that would otherwise need to be built and maintained in the wrong layer.

The objective is not to minimize the number of technologies. It is to minimize the amount of custom machinery required to achieve the guarantees the business actually needs.

---

## The failures were more educational than the successful demonstrations

One of the most valuable aspects of this journey was the repeated adversarial examination of the implementation.

A feature could compile. A test could pass. A workflow could return success. Yet the actual business operation might not have been connected to the mechanism that supposedly protected it.

That distinction forced us to become more precise about what constituted evidence.

We learned to separate several states that developers frequently collapse into one:

- Implemented does not mean built.
- Built does not mean executed.
- Executed does not mean passed.
- A passed infrastructure test does not necessarily prove a business guarantee.
- A verified transaction does not necessarily prove the correctness of its contents.

The early implementation demonstrated several isolated mechanisms. Subsequent independent review showed that the real payroll computation, workflow execution and ledger publication were not yet connected.

We had to return to the underlying authority and transaction boundaries, not merely add more tests around the existing interfaces.

The final remediation connected the actual business path and repeated the critical tests.

This is an important lesson for engineering teams: **the greatest risk in a complex system is sometimes not an obvious failure, but a successful operation that never performed the work everyone assumed it had performed.**

Adversarial review is especially valuable because it challenges the relationship between what a system claims to guarantee and what its code actually enforces.

The repository preserves that journey. Readers can examine the earlier findings and decide for themselves whether the corrections fully resolve the issues.

---

## Reproducibility changes how we think about business history

Another insight emerged from the payroll calculation itself.

A historical financial result is not reproducible merely because we saved the final amount.

To explain an old payroll calculation, we need enough information to reconstruct the inputs and the interpretation applied to them.

That includes:

- The relevant employment and compensation records.
- The attendance and approved timesheets.
- The effective-dated rule versions.
- The calculator version.
- The original calculation output.
- The approval and finalization evidence.

This becomes especially important when business rules change.

A company might revise its overtime policy, introduce a new allowance, or change the way a particular employee category is treated.

The new rule must apply according to its effective dates. It must not silently rewrite historical payroll.

Our synthetic demonstration used two rule packs during the same month. This gave us a concrete way to show that rule changes can influence calculations without making every historical record mutable.

It also exposed a limitation worth acknowledging: a calculator version label is not the same as a cryptographic digest of the actual implementation binary. Likewise, our multi-pack digesting and cross-version canonical serialization remain narrower than a fully mature historical-reproduction system would require.

The point is not that every historical question has been solved.

It is that reproducibility must be an architectural property considered when data is first captured, not a reporting feature added after a dispute occurs.

---

## Why we deliberately stopped short of building a complete payroll system

There is a temptation, once an interesting technical foundation exists, to keep adding functionality.

More employee classifications. More statutory packages. More approval types. More integrations. Perhaps even payment execution and regulatory filing.

That would have defeated the purpose of this exercise.

We wanted to demonstrate architectural responsibilities, not compete with established payroll vendors.

The synthetic workload was sufficient to expose temporal changes, calculation boundaries, missing evidence, approvals, interrupted execution, finalization, and historical corrections.

It also allowed us to keep statutory interpretation separate from the infrastructure experiment. The Singapore overtime package remained blocked because the underlying legislative evidence was incomplete. We did not replace missing legal evidence with assumptions merely to make the demonstration more impressive.

This restraint is important.

An incomplete but honestly bounded engineering specimen is more useful to a reader than a purportedly complete system whose assumptions are hidden.

The repository therefore should not be read as a turnkey payroll platform or a universal blueprint for every business.

It is a working exploration of a particular architectural idea. It contains executable paths, supporting schemas, tests, runbooks, and recorded results. It also retains unresolved limitations and intentionally leaves room for readers to inspect, question, and extend the implementation themselves.

---

## What this means for engineers and business owners

For software engineers, the central lesson is to treat data integrity, durable execution and historical evidence as different engineering problems.

Do not expect application validation to replace database authority. Do not assume a background worker can reliably recover from every interruption. Do not confuse ordinary audit history with independent cryptographic verification.

For business owners, the lesson is equally practical.

The cost of a business application is not simply the cost of building its visible features. It includes the cost of ensuring those features behave correctly when people, data, rules and processes change.

Payroll is a useful example because the consequences of incorrect results are immediate and tangible. But the same concerns appear in insurance claims, financial operations, healthcare administration, supply chains, compliance workflows and enterprise resource planning.

A system that can explain what it did, recover when interrupted, and establish the integrity of its historical evidence has a different quality from one that merely produces a successful response.

That quality is difficult to capture in a feature checklist. It becomes visible under failure.

## The experiment is finished. The questions are not.

We began with a relatively simple proposition: PostgreSQL, DBOS and immudb might provide more value together than a conventional application that places every responsibility in its own Go code and database tables.

The completed experiment gives us a working demonstration of that proposition.

PostgreSQL governs business authority. DBOS preserves the continuity of execution. Go performs the domain calculations. immudb anchors evidence that can be independently examined.

The technologies do not eliminate the need for careful engineering. They make it possible to assign difficult responsibilities to components built to handle them, while leaving the application responsible for the business meaning.

There are still questions worth exploring.

How should calculator implementation provenance be secured across years of software changes? How should an organization retain independent trust anchors over long periods? What is the right balance between database-enforced invariants and domain-level validation? When does an external evidence ledger justify its operational and licensing costs? Could a carefully engineered PostgreSQL-only design provide sufficient assurance for a particular business?

Those questions are deliberately not answered here.

The code, experiments and recorded limitations provide a place to start.

The more useful outcome of this exercise is not a declaration that three technologies are always better than one.

It is a better way to ask an architectural question:

**What must be true about our business data, what must remain true about our processes, and what evidence would convince someone else that both remained true?**

That is a more demanding standard than simply making software work.

It is also a more valuable one.

Editorial recommendation: I would publish this with the title When Payroll Fails, Who Remembers What Happened? It speaks to engineers through the technical architecture and to business owners through accountability, continuity and financial trust.

The article intentionally avoids turning the repository into a step-by-step tutorial or reproducing the adversarial review. The implementation history, original findings and remaining technical limitations can speak for themselves in the repository.

The figures and final architecture claims are grounded in the remediation report and updated final report.