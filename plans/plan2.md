I consolidated the DeepSeek, Z, Gemini and ChatGPT reviews into a single implementation plan.

Created

Version 0.1 · 8 November 2026 · 629 lines

Download WorkforceOps Lab — MVP Plan v0.1

## The important decisions I consolidated

I resolved the differences between the reviews in favor of a smaller, more coherent implementation.

| Decision             | Consolidated direction                                  |
| -------------------- | ------------------------------------------------------- |
| Product              | WorkforceOps Lab                                        |
| Fictional employer   | Meridian Workforce Pte. Ltd.                            |
| System A             | ShiftHub — employee, scheduling and attendance          |
| System B             | PayLoop — payroll and payslips                          |
| Integration          | Separate Go worker, realistic API/batch synchronization |
| Database             | Neon PostgreSQL, isolated schemas and roles             |
| Deployment           | Fly.io                                                  |
| Reference oracle     | Independently implemented expected-results calculator   |
| Investigation        | CaseTrace, one bounded agent workflow initially         |
| Synthetic population | 50–100 employees, 3–6 months of history                 |
| Initial incidents    | Three deeply reproducible scenarios                     |

The strongest ideas from the reviews are retained: deterministic data generation, a separate calculation oracle, mechanism-based fault injection, hidden ground truth, evidence-led investigations and human-versus-agent evaluation.&#x20;

plan_alt.md

plan_alt.md



## What I deliberately changed

First, I avoided building an entire HRIS/payroll platform. The two-system integration boundary is the focus. Four or six independently deployed applications would introduce unnecessary operational complexity.

Second, I made the independent oracle a core component. PayLoop must not validate its own calculations. This is essential if we want credible ground truth.

Third, I narrowed the initial regulatory implementation. MOM overtime is the first rule pack. CPF is a gated second rule pack requiring complete, applicable contribution tables—not merely a few headline rates. The current MOM guidance includes important eligibility distinctions and calculation limits, while CPF depends on citizenship, SPR year, age and wage bands.&#x20;

[image](https://www.google.com/s2/favicons?domain=https://www.mom.gov.sg\&sz=32)

mom.gov.sg+2



Fourth, I postponed the dashboard and multi-agent swarm. First prove that the underlying business transactions, integration failures and evidence-based diagnosis work.

## What the coding agent receives

The document includes a coding-agent execution contract and acceptance gates. It directs the agent to produce the architecture, domain model, API contracts, regulatory source register, oracle design and scenario definitions before implementation.

The implementation sequence is:

1. Clean ShiftHub → integration → PayLoop business path.
2. Independent reference oracle.
3. Realistic integration failures.
4. Three reproducible incidents.
5. Read-only investigation tools and evidence ledger.
6. Agent evaluation and demonstrable results.

My recommendation: hand the agent this plan as its scope baseline. Have it complete Phase 0 and present the design artifacts before it starts generating the application. That gives you an opportunity to catch architectural mistakes before they become expensive code.

The objective is not to produce a large synthetic database. It is to create a small enterprise environment where an engineer can investigate a realistic problem and prove exactly what went wrong.
