I've reviewed `plan_prompt3.md`. This is a strong execution plan, and I agree with its overall approach. It addresses the six findings systematically without reopening the product architecture or accidentally starting Phase 2.&#x20;

plan_prompt3.md



Approve with four refinements

I would let the coding agent execute this correction mandate, subject to the following clarifications.

## 1. Four refinements before execution

R1 — Per-workspace database identities must be fully specified.

The proposed PostgreSQL RLS design using `current_user` mapped to an immutable workspace identity is a reasonable approach.&#x20;

plan_prompt3.md



But specify these additional requirements:

- RLS enabled and `FORCE ROW LEVEL SECURITY` on protected tables.
- Runtime roles must not own protected tables or have `BYPASSRLS`.
- No membership or `SET ROLE` path allowing an applicant to assume another workspace's identity.
- Workspace identity mapping must not be writable or inspectable by applicant roles.
- Connection pools must be associated with the correct workspace database identity.
- Tests must attempt workspace impersonation, not just ordinary cross-workspace SELECT.

Also clarify how workspace identities coexist with ShiftHub, PayLoop, investigator and evaluator service identities. The role-provisioning model needs to handle both dimensions.

The MVP can remain single-workspace, but the multi-workspace design must be credible.

R2 — Prevent fixture-to-production-schema drift.

The plan proposes proving RLS using a fixture schema while the real business tables arrive in Phase 2.&#x20;

plan_prompt3.md



That is acceptable for this correction phase, but the fixture should reuse the actual planned migration definitions or a common SQL source.

Do not maintain two independently written versions of the security policies and assume they will remain equivalent.

R3 — Preserve transactional outbox atomicity after removing sync-worker business access.

This is the most important refinement to R1.

The ShiftHub application must be able to write its business transaction and corresponding outbox event in the same database transaction.

But the sync worker must not access either business schema.

I recommend:

- ShiftHub gets narrowly scoped permission to enqueue an event, possibly through a hardened `SECURITY DEFINER` function.
- The worker reads immutable payloads from the platform outbox.
- The worker updates delivery state, not the original business payload.
- PayLoop receives the event through its authenticated API and writes its own records.

The agent must resolve this explicitly rather than accidentally breaking atomicity while tightening permissions.

R4 — Capability declarations must correspond to real module implementations.

The revised module contract proposes capability descriptors for rules, generators, oracles, faults and other extensions.&#x20;

plan_prompt3.md



Require that a module declaring a capability actually registers the corresponding implementation or provider.

For example, a module declaring `reference-oracle` must supply a usable oracle implementation, not merely a descriptive string in `module.yaml`.

Implement only the capabilities consumed by the reference module now. Do not build generic placeholder frameworks for every future extension.

## 2. Assessment of the execution sequence

I agree with the order proposed:

| Order | Work                           | Assessment                    |
| ----- | ------------------------------ | ----------------------------- |
| 1     | M1 — Workspace isolation       | Highest priority              |
| 2     | M4 — Service identities        | Correct                       |
| 3     | M2 — Oracle secrecy            | Correct                       |
| 4     | M3 — Remediation authorization | Correct                       |
| 5     | R1 — Sync boundary             | Correct, subject to atomicity |
| 6     | R2 — Module contract           | Correct                       |
| 7     | Cross-document audit           | Mandatory                     |

The plan appropriately distinguishes actual database-role testing from features that will only be implemented in later phases.&#x20;

plan_prompt3.md



One other important point: do not let the self-hosted installation claim tamper-proof assessment integrity. The plan correctly acknowledges that an administrator controlling their own installation can inspect its contents. That limitation should remain explicit.

## 3. My decision on the regulatory research question

The plan asks whether `prompt2.md` should run before or alongside Phase 2.

My recommendation is:

Run the Singapore regulatory deep research in parallel with this Phase 1 correction work.

Do not make regulatory research a dependency of the security architecture corrections.

However, its output must be available before implementing the overtime rule pack and constructing its golden tests.

This avoids delaying independent engineering work while ensuring we do not implement statutory calculations from unverified assumptions.

## 4. Exact instruction to append to plan_prompt3

### Reviewer refinements — mandatory execution clarifications

Incorporate these four refinements into the correction mandate before execution:

1. **M1:** Specify per-workspace database identities across application and service roles, connection-pool binding, FORCE ROW LEVEL SECURITY, table ownership, BYPASSRLS restrictions and attempted workspace-impersonation tests.
2. **M1/D2:** Ensure PostgreSQL security fixtures reuse the same canonical migration or policy definitions intended for the Phase 2 business schemas. Avoid independently maintained fixture and production policies.
3. **R1:** Preserve transactional outbox atomicity without granting the sync worker business-schema access. Specify narrowly scoped ShiftHub event-enqueue permission, immutable payload access for the worker, and delivery-state permissions. PayLoop must persist imports through its own API and database identity.
4. **R2:** Capability descriptors must resolve to actual registered implementations. A declared capability without its required implementation must fail validation or registration. Implement only the capabilities consumed by the reference module; do not introduce placeholder frameworks.

Regulatory research (`prompt2.md`) proceeds independently and in parallel with this correction pass. Its verified findings become mandatory inputs before Phase 2 overtime-rule implementation.

All other scope restrictions and acceptance criteria remain unchanged. The correction pass must stop at the Phase 1 gate.

## 5. Final recommendation

Authorize execution of plan_prompt3 with these refinements.

I do not consider this an approval of Phase 2 yet. The coding agent must return actual PostgreSQL permission-test results, document consistency results and a final PASS / CONDITIONAL PASS / BLOCKED decision.

The most important thing to prove is that a learner can investigate and remediate their own sandbox problems without accessing another learner's data or discovering the hidden answer.

That security property, together with the integrity of the independent oracle, is the foundation on which the training platform depends.