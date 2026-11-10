I reviewed plan3.2 against the original coding-agent mandate and both architecture amendments.

Approved with minor corrections

This is now a coherent implementation plan. I would use it as the authoritative execution baseline rather than continue producing more architectural prompts.

More importantly, it resolves the earlier competing directions: the project now has a modular platform kernel, separate business modules, three application processes, a CLI, REST APIs and a Go-rendered Operations Console.&#x20;

plan3.2.md



## 1. What plan3.2 gets right

| Area                  | Assessment                                            |
| --------------------- | ----------------------------------------------------- |
| Repository strategy   | New `a1` module; original `bchat` untouched           |
| Infrastructure reuse  | Echo, pgx, Neon, Fly.io and reusable LLM connectivity |
| Platform architecture | Generic kernel separated from business modules        |
| Service boundaries    | ShiftHub, PayLoop and sync worker                     |
| Database security     | Separate schemas and least-privilege roles            |
| Business rules        | Singapore overtime first; CPF gated                   |
| Oracle                | Independently implemented expected-results calculator |
| Simulation            | Reproducible, mechanism-based incidents               |
| CaseTrace             | Generic investigation and evidence infrastructure     |
| User experience       | CLI, REST API and Operations Console                  |
| Open source           | Credential-free local setup                           |
| Extensibility         | Second order-fulfillment example module               |

I particularly agree with the decision to keep three separate processes instead of collapsing everything into one API process. This preserves the ability to simulate failures across actual application boundaries.&#x20;

plan3.2.md



The phased implementation is also sensible: no fault injection before baseline reconciliation, and no agent evaluation before the investigation environment works.&#x20;

plan3.2.md



## 2. Five corrections I recommend

These are refinements, not reasons to redesign the architecture.

1\. Make oracle independence a verifiable requirement, not an absolute claim.

The oracle should have a separate implementation, independent derivation, golden test fixtures and no dependency on PayLoop's calculation package.

An architectural test can prevent imports and shared calculation calls. It cannot mathematically prove that two independently written implementations are correct or genuinely independent in their reasoning.

Require separate review of the calculation logic.

2\. Refine the core/module architectural tests.

Plan3.2 proposes an import check, Singapore terminology grep and registration contract test.

Keep the import and registration tests mandatory. Treat terminology scanning as a heuristic, not proof of architectural independence.

A Go dependency-graph test is stronger than simply searching for words.&#x20;

plan3.2.md



3\. Define cross-interface consistency precisely.

CLI, API and UI must produce equivalent outcomes, but test them against separate, identical fixture states.

For example:

- Scenario A executed through CLI.
- Scenario A executed through REST API.
- Scenario A executed through UI.

Compare normalized resulting business state, integration events and evidence hashes.

Do not execute the same mutating operation three times against one shared state and call that a consistency test.

4\. Clarify Fly.io deployment verification.

Keep local build, Docker validation and isolated Neon migration tests mandatory.

A remote Fly.io health check should be conditional on explicit authorization for the non-production deployment. The plan already identifies this outstanding deployment decision.&#x20;

plan3.2.md



5\. Resolve the Go toolchain discrepancy deliberately.

The audit identifies installed Go 1.25.7 versus the existing repository's Go 1.26 declaration.

The agent must inspect the actual toolchain and decide the new module's Go version deliberately. It must not blindly inherit an incompatible directive or silently downgrade dependencies.

## 3. One important commercial consideration

I would retain the tiny second module as a mandatory extensibility proof.

But the commercial architecture should also make one distinction explicit:

- Public modules: available in the open-source repository.
- Private customer modules: maintained in separate repositories, compiled into customer-specific builds.
- Customer-specific data, rules, credentials and incident scenarios: isolated from public examples.

This supports paid customization without requiring proprietary customer logic to become public.

Static module registration, as selected in plan3.2, is entirely adequate for this initial approach. There is no need for a plugin marketplace or dynamic module-loading framework.&#x20;

plan3.2.md



## 4. Decisions to freeze

I would resolve the remaining questions as follows:

| Decision              | Recommendation                                  |
| --------------------- | ----------------------------------------------- |
| New repository/module | `github.com/PithomLabs/a1`                      |
| Existing bchat        | Preserve untouched                              |
| Development database  | Local PostgreSQL                                |
| Cloud database        | Isolated Neon branch                            |
| Deployment            | Fly.io, authorization required                  |
| First business module | Singapore WFM/payroll                           |
| First statutory rule  | MOM overtime                                    |
| CPF                   | Deferred until independently verified           |
| UI                    | Go templates + HTMX                             |
| Extensibility         | Static module registration                      |
| Example extension     | Minimal order fulfillment                       |
| License               | Approve Apache-2.0 or MIT before public release |

The absence of an existing `bchat` LICENSE makes license selection an actual approval decision, not something the coding agent should assume.&#x20;

plan3.2.md



## 5. What I would tell the coding agent now

I would give it plan3.2 and the earlier coding-agent mandate, with plan3.2 explicitly designated as the consolidated authority.

The immediate instruction should be:

> Execute Phase 0 and Phase 1. Inspect the original repository, document reusable infrastructure, establish the new module, and produce the complete architecture and interface-design artifacts. Incorporate the five review corrections above. Do not begin expanding business functionality until the architecture, domain boundaries, module contract, database permissions, API contracts, CLI and UI workflows are documented. Once those artifacts pass the design gate, proceed autonomously through the implementation phases. Preserve all existing bchat files and do not modify production Neon or Fly.io infrastructure.

My overall recommendation: freeze plan3.2 after these refinements.

We are now at the point where more architectural discussion risks becoming a substitute for building the product.

The first real proof will be a clean synthetic workforce transaction flowing through ShiftHub, the integration worker and PayLoop, with an independently verified payroll result.

Once that works, we have the foundation for the incident simulator, CaseTrace, the Operations Console and eventually customer-specific business modules.
