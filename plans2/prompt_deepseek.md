Based on your architecture and the industry patterns we discussed, I'll write a detailed research prompt for an implementation agent. The prompt is designed to force a rigorous, evidence-based evaluation of how PostgreSQL 18.6, DBOS, and immudb can be integrated into your existing WorkforceOps codebase, with the explicit goal of building a 1-month payroll POC and identifying reusable architectural patterns.

---

## Prompt for Implementation Agent: Deep Research on PostgreSQL 18.6 + DBOS + immudb Integration

**Context:** The WorkforceOps codebase currently uses PostgreSQL 16 for payroll snapshot integrity. The architectural reviews identified that the current approach (REPEATABLE READ + session-scoped advisory locks + XID tickets) is complex and may not fully close the authority boundary. The industry standard splits the problem across layers: **Postgres** for operational truth and constraints, **DBOS** for durable workflow execution and exactly-once coordination, and **immudb** for tamper-evident audit and provenance. Your task is to deeply research the existing codebase and produce a concrete integration architecture that demonstrates a 1-month payroll run as a POC.

**Primary Goals:**

1.  **POC: 1-Month Payroll Run.** Design and plan the implementation of a complete, non-trivial payroll cycle (input capture → resolution → snapshot → validation → calculation stub) using the three technologies. The POC must demonstrate the integration patterns, not just pass tests.
2.  **Architectural Pattern Identification.** For each of the three technologies, identify the specific payroll use cases where it is the *best fit* (e.g., DBOS for snapshot capture orchestration, immudb for audit of validation decisions, Postgres 18 for effective-dated compensation constraints). Map these patterns to the existing defects (D1, EF-E3, P1a–g, F-A1, F-B1/B2, C1/C2).
3.  **Smooth Integration & Upgrade.** Plan the upgrade from local Postgres 16 to Postgres 18.6, the integration of DBOS as a Go library (not a server), and immudb as a sidecar ledger. The integration must minimize disruption to existing migrations (`00001`–`00012`) and be compatible with the frozen architecture (one Go executable, one database, no microservices).

**Required Research Areas:**

**A. Codebase Architecture Deep-Dive**
- Map the current Go services, database functions, triggers, and migration structure. Identify the exact boundaries between `internal/payroll`, `internal/compbasis`, `internal/hr`, `internal/rules`, and `core/database`.
- Document the current snapshot capture and validation flow, including all lock acquisition points and transaction boundaries.
- Identify all writer paths (service calls, triggers, direct SQL) that mutate payroll inputs.

**B. PostgreSQL 18.6 Evaluation**
- Research `WITHOUT OVERLAPS` temporal primary keys and `PERIOD` foreign keys. Determine how they can enforce compensation effective-date non-overlap, SHG full-period coverage, and rule-pack period validation natively at the database level.
- Assess the migration path: can `00013` use `WITHOUT OVERLAPS` while preserving the existing schema for `00001`–`00012`? What are the `btree_gist` or exclusion-constraint fallbacks if an upgrade is not immediately feasible?
- Evaluate whether the current advisory-lock protocol can be retired in favor of transaction-scoped gates or `SERIALIZABLE` + retry, now that Postgres 18 provides stronger temporal constraints.

**C. DBOS Integration for Durable Capture**
- Research the DBOS Go SDK (`github.com/dbos-inc/dbos-transact-golang`). How do you register workflows and steps? How does it provide exactly-once semantics for database writes?
- Design a **Snapshot Capture Workflow**: a DBOS workflow that (1) acquires a transaction-scoped employer gate, (2) reads inputs at a consistent snapshot, (3) computes the canonical hash in Go, (4) writes the snapshot via a controlled function, and (5) writes an audit record. Use DBOS's recovery mechanism to handle crashes between the Postgres commit and the immudb append.
- Determine how DBOS's system database can coexist with the payroll schema. Should it use a separate schema (`dbos`) or the same database? How does this affect the frozen "one database" constraint?

**D. immudb Integration for Tamper-Evident Audit**
- Research immudb 1.11+ (which supports the PostgreSQL wire protocol). How can it act as a sidecar ledger for payroll? What are the client libraries for Go (`github.com/codenotary/immudb/pkg/client`)?
- Design an **Audit Trail Pattern**: for every snapshot validation, write a compact provenance record to immudb (inputs hash, rule version, output hash, Postgres transaction ID, actor, timestamp). Use immudb's `VerifiedTxByID` for later audit.
- Evaluate the **dual-write problem**: if Postgres commits but the immudb append fails, how does the system reconcile? Propose a reconciliation pattern (e.g., DBOS durable workflow, outbox table, or periodic audit scan).
- Assess whether immudb can replace or complement the current snapshot hash verification. If immudb stores the canonical hash, does the Postgres snapshot table still need to store it?

**E. Architectural Patterns & Defect Mapping**
- For each of the major findings from the adversarial review (D1-P0, EF-E3-P0, P1a–g, F-A1-WIDEN/FREETAIL/MIDPERIOD, F-B1/B2, X-FUND, C1-SOURCES/GOLDENS, C2-REPIN), propose a specific integration pattern using one or more of the three technologies.
- Example: D1-P0 → DBOS durable workflow with a transaction-scoped gate; EF-E3-P0 → immudb audit trail with derived-in-DB capture; F-A1-WIDEN → Postgres 18 temporal constraints + resolver split.

**F. Integration & Upgrade Plan**
- Propose a step-by-step migration from Postgres 16 to 18.6. What are the compatibility risks? How do you test the upgrade on a disposable database?
- Propose the DBOS integration plan: how to add the library to the Go executable, how to register workflows, and how to handle schema changes for the DBOS system tables.
- Propose the immudb integration plan: how to run it locally (Docker or embedded), how to connect from Go, and how to test the tamper-evidence proofs.
- Identify the minimal changes required to existing migrations (`00001`–`00012`) and confirm they remain untouched.

**Deliverable:**

A detailed **Architecture Decision Record (ADR)** with:
1.  A recommended integration architecture diagram (three-layer: DBOS workflow → Postgres truth → immudb ledger).
2.  A POC implementation plan for the 1-month payroll run, including specific code changes, migration files, and test scenarios.
3.  A mapping table: each adversarial finding → proposed pattern → technology → enforcement point → test.
4.  A risk register: what could go wrong during integration, and mitigation strategies.
5.  A clear recommendation: which of the three technologies should be adopted first, and what is the minimal viable integration path to demonstrate non-trivial complexity.

**Constraints:**
- Work uncommitted; do not modify `00001`–`00012`.
- Use disposable PostgreSQL 18.6 databases only.
- Preserve OT-001 BLOCKED and zero statutory calculations.
- The POC must be demonstrable via CLI/REST, not just unit tests.