# Gate 1 package review



D1 mechanism accepted in principle

Boundary A approval on hold



The package is substantially stronger than the earlier plan. I accept the XID-ticket and session-barrier design as the selected D1 mechanism. However, I would not yet authorize the `00013` privilege change exactly as currently specified.

The remaining objections are not about the underlying admission protocol. They concern whether the proposed privilege boundary can be installed, maintained and safely exposed through PostgreSQL functions.

## 1. What I consider demonstrated

The package provides a coherent experimental case for the D1 decision:

- The stale-snapshot ordering proof uses ticket-row visibility and `pg_visible_in_snapshot`, not a global `xmin` comparison. The unrelated-old-transaction test is particularly valuable because it demonstrates the false-rejection problem the earlier design risked.&#x20;

  gate1_admission_package.md

- The package identifies why `pg_locks` is not sufficient to distinguish session-level from transaction-level advisory locks on the tested PostgreSQL 16.15 instance. It therefore moves the enforcement boundary to function privileges rather than inferring lock scope.&#x20;

  gate1_admission_package.md

  gate1_admission_package.md

- The concurrent-admission and cleanup tests address important operational failure modes: contention, bounded retries, cancellation, panic, failed unlock, physical connection discard and pool reuse.&#x20;

  gate1_admission_package.md

- The package correctly distinguishes mechanism proofs from implementation acceptance. The original stale-snapshot attack, writer-first inclusion tests and direct-SQL snapshot-fabrication attack remain unresolved against current code. They should not be declared fixed at this stage.&#x20;

  gate1_admission_package.md

  gate1_admission_package.md



I also accept the scope discipline: no capture/validation implementation yet, zero statutory calculations, and `SG-MOM-OT-001` remains blocked.&#x20;

gate1_admission_package.md



## 2. Four items still prevent privilege-change approval

### A. Who actually has authority to revoke the 21 built-in functions?

This is the most important unresolved issue.

The package proposes revoking privileges on 21 `pg_catalog` advisory-lock functions and granting execution to `migration_admin`. It also says Goose migrations run as `migration_admin`.&#x20;

gate1_admission_package.md



But PostgreSQL does not allow an ordinary role to revoke function privileges merely because it owns the database or runs the migration. The role performing `REVOKE` must have the necessary grant authority, normally through function ownership, membership in the owning role or superuser privileges.&#x20;

[image](https://www.google.com/s2/favicons?domain=https://www.postgresql.org\&sz=32)

postgresql.org+1



The package does not show the actual owners of the 21 `pg_catalog` functions or demonstrate that `migration_admin` has the necessary authority over them.

Required resolution:

Choose and document one of these execution models:

1. A privileged provisioning step performs the `REVOKE` and `GRANT` operations, while migration `00013`, running as `migration_admin`, verifies that the exact ACL state is already installed and fails closed otherwise.
2. `migration_admin` performs the changes itself, but only after a live test proves it has the required ownership or grant authority for every function.

Do not silently assume that being the migration role grants authority over PostgreSQL's built-in functions. The package already mentions superuser provisioning, so the first model may fit the existing two-role design better.

### B. The `SECURITY DEFINER` functions need an explicit security contract

The package says the three frozen helpers will become `SECURITY DEFINER`, with their owners and grants preserved.&#x20;

gate1_admission_package.md



That is not sufficient by itself.

A `SECURITY DEFINER` function executes with the privileges of its owner. PostgreSQL specifically recommends restricting its `search_path` to trusted schemas and placing `pg_temp` last, to prevent an untrusted caller from substituting objects used by the function.&#x20;

[image](https://www.google.com/s2/favicons?domain=https://www.postgresql.org\&sz=32)

postgresql.org



For the three existing helpers and every new privileged wrapper, `00013` must establish and test:

- The exact intended owner, which must be capable of executing the raw advisory functions. Under the proposed model, that should be `migration_admin`.
- An explicit, safe `search_path`; no schema writable by `app` or another untrusted role should be searched as trusted.
- `pg_temp` last, with schema-qualified references wherever practical.
- The exact `EXECUTE` grants for `app`, `PUBLIC` and `migration_admin`.
- That caller-controlled arguments cannot turn a narrowly scoped helper into a generic privileged-operation interface.

There is an additional ACL issue here: PostgreSQL grants `EXECUTE` on newly created functions to `PUBLIC` by default. Existing functions retain their ACLs when replaced using `CREATE OR REPLACE`.&#x20;

[image](https://www.google.com/s2/favicons?domain=https://www.postgresql.org\&sz=32)

postgresql.org+1



Consequently, the new bootstrap wrapper, barrier wrappers and ticket lifecycle functions must explicitly revoke unintended `PUBLIC` execution and grant only the intended roles. The existing three helpers must have their preserved ACLs inspected rather than assumed safe.

### C. Prove the effective role graph, not just the direct grants

The package reports that all 21 functions are denied to `app` and `PUBLIC`, and granted to `migration_admin`.&#x20;

gate1_admission_package.md



That is an important test, but a direct ACL inspection is not the complete permission test.

PostgreSQL combines privileges granted directly to a role, privileges inherited through role membership, and privileges granted to `PUBLIC`. Membership can also allow `SET ROLE` into another role.&#x20;

[image](https://www.google.com/s2/favicons?domain=https://www.postgresql.org\&sz=32)

postgresql.org+1



Add explicit post-migration assertions that:

- `app` is not a direct or indirect member of `migration_admin`.
- `app` is not a superuser and cannot assume a role with superuser or equivalent advisory-function privileges.
- `has_function_privilege('app', ...)` returns false for each raw advisory overload.
- The effective execution privileges on every `SECURITY DEFINER` wrapper are exactly those intended.
- The app can still call the approved narrow helpers needed by existing application and trigger paths.

The goal is to prove that revoking `PUBLIC` really removes the capability from the runtime role, rather than just removing one route through which it could obtain that capability.

### D. The Down migration does not yet describe an exact privilege restoration

The package states that Down restores the old `INVOKER` definitions, re-grants `PUBLIC`, and drops the new objects.&#x20;

gate1_admission_package.md



But Up also explicitly grants execution on the 21 raw functions to `migration_admin`.

Unless that grant existed before Up, Down must revoke it as well. Otherwise the previous privilege state is not restored exactly.

Require a before/after ACL comparison covering:

- Original ACLs on all 21 raw functions.
- The `migration_admin` grants added by Up.
- The original helper owners, ACLs and function security attributes.
- All new wrapper grants and ownership.
- The state after `00013 Down`, followed by a successful `00013 Up` again.

The test should compare effective privileges and recorded ACLs, not just whether the application can still start.

## 3. Two additional implementation checks

These are not reasons to reject the D1 mechanism, but they should be recorded before capture implementation proceeds.

Expiry semantics. The package specifies a five-minute ticket lifetime. Confirm that the expiry check uses actual wall-clock or statement time, rather than PostgreSQL's `now()`/`transaction_timestamp()`, which represent transaction-start time. Include a test in which a transaction begins before expiry but attempts verification after expiry. This matters because timestamps must not accidentally become the authority for a temporal decision.

Connection identity. Ticket binding is described as PID plus employer.&#x20;

gate1_admission_package.md

&#x20;Since a PID is not a permanent connection identity, consider including the backend start time in the ticket binding, or explicitly documenting why the existing lock-hold verification and ticket lifecycle make PID reuse harmless within the threat model.



## 4. Final decision

| Decision                                 | My assessment                                                            |
| ---------------------------------------- | ------------------------------------------------------------------------ |
| P1a expected-population semantics        | Accept as recorded                                                       |
| ADR-1, Go-canonical resolver             | Accept as recorded                                                       |
| XID-ticket and session-barrier mechanism | Accept as the selected D1 design                                         |
| 21-function revocation boundary          | Conditionally acceptable; installation authority must be resolved        |
| Three `SECURITY DEFINER` helper changes  | Not yet approved; owner, `search_path` and ACL contract must be explicit |
| Snapshot capture and validation          | Not authorized                                                           |
| Statutory calculations and OT-001        | Remain blocked                                                           |

I recommend approving the D1 design decision, but withholding authorization to implement `00013` until the four privilege-boundary items above are incorporated into the package.

This is a narrow amendment, not a request to reopen the architecture. The two-role model remains intact; there is no need to introduce a third role simply to address these questions.

The governing acceptance test remains:

> What invariant was previously unenforced, which operations could violate it, and what now makes every such violation impossible or explicitly blocked within the documented threat model?

The next package should demonstrate not only that the advisory calls are denied in a scratch database, but also that the intended provisioning and migration process can reliably install and preserve that boundary.