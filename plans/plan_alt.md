## deepseek


Yes — this is a better use of your next 30 days than cold-emailing for production credentials. But be precise about *what it proves*. It does not prove demand, and it does not substitute for customer data. What it does is collapse three risks you can control: you can build the investigation engine, you can define what "correct" looks like, and you can produce a scored demo without a security review. That is exactly the right thing to de-risk while the sales motion is stalled.

The key insight is that **Singapore's statutory rules give you a deterministic oracle**. CPF, overtime, leave, and tax filing are not judgment calls — they are published formulas and tables. That means you can generate data where you *know* the right answer, then inject defects that make the payroll output wrong, and score the agent against ground truth. That is the answer key your earlier plan said you needed.

---

## 1. Separate the deterministic layer from the investigation layer

Tier 2 support difficulty does not come from not knowing the CPF rate. It comes from **the interaction between a correct rule and a broken data path**. So build two things:

**A compliance oracle (deterministic, Go).** This computes what *should* have happened. Encode:

- **Overtime**: 1.5× hourly basic rate; hourly basic = `12 × monthly basic / (52 × 44)` for monthly-rated employees; 44-hour week, 72-hour monthly cap; Part 4 coverage thresholds ($2,600 non-workman / $4,500 workman).
- **CPF**: 2026 rate table by age band — 55 and below 17% employer / 20% employee; above 55–60 16/18; above 60–65 12.5/12.5; above 65–70 9/7.5; above 70 7.5/5, for wages above $750. Ordinary Wage ceiling **$8,000** from 1 Jan 2026. Handle SPR year 1/2 and work-pass-holder exclusions.
- **Leave**: annual leave 7→14 days by service year; 14 days outpatient sick + 60 days hospitalisation; childcare 6/2 days; 11 public holidays with the 2026 observed-date rules.
- **IRAS**: IR8A/AIS by 1 March; IR21 tax clearance and withholding for departing foreign employees.

This oracle is the most valuable artifact you will build. It is also the thing a Singapore payroll practitioner can review in an afternoon and tell you is wrong — which is far cheaper than discovering it in a customer incident.

**A synthetic two-system world.** Do not model a whole HR SaaS. Model the **integration boundary**, because that is where support investigations actually live. Build:

1. **HRIS service** — employee master, attendance punches, leave approvals, salary effective dates.
2. **Payroll service** — pay runs, CPF computation, payslip lines, IR8A extract.
3. A **connector** between them with realistic plumbing: a scheduled job, a CSV/API export, an idempotency key, a cut-off timestamp, and a timezone assumption.

This is the correct scope. The investigation problem is *"the attendance system recorded 480 minutes but payroll read 460"* or *"CPF doubled after a backpay increment"* — not *"what is the formula"*.

---

## 2. Design the defect taxonomy before you write the agent

Your previous plan said "one narrow problem." Here is how to make that concrete. Write a `defects.yaml` where each entry has a root cause, the minimal evidence that proves it, and the misleading surface symptom. Seed at least these classes:

| Defect class | Surface symptom | Minimal evidence |
|---|---|---|
| **Stale CPF rate** (employee crossed age 55 mid-year, rate not updated) | CPF deducted at 37% instead of 34% | DOB, pay period, rate-table version used |
| **OW ceiling not applied** | CPF over-contributed on salary above $8,000 | OW ceiling constant, ordinary wage line, contribution record |
| **Doubled backpay CPF** | Employee's CPF roughly double in the month backpay lands | Backpay line present in both ordinary wage and additional wage bases; recompute delta |
| **Attendance truncation** | Paid hours lower than approved OT | Raw punch records vs connector output row count / minutes |
| **Effective-date cutover** | New salary applied one pay cycle late | Change event timestamp vs payroll cut-off timestamp |
| **Work-pass holder CPF flag** | CPF paid for someone not entitled | Employment pass expiry, citizenship flag copied from local template |
| **OT rounded before aggregation** | Small per-day differences accumulate | Daily OT values vs monthly total; rounding rule |
| **IR8A exclusion after IR21** | Duplicate income reported | IR21 filing date vs IR8A inclusion flag |

Each seeded defect must carry a `ground_truth` block: root cause, affected employee IDs, minimal evidence set, and the correct resolution. That is what makes evaluation possible.

---

## 3. Build the investigation engine as an evidence ledger, not a chat loop

The product differentiator you identified — the investigation record — is also the evaluation mechanism. Use a schema that forces claims to be grounded:

```
scenario(id, seed, defect_class, ground_truth_json)
agent_run(id, scenario_id, status, model, token_cost)
hypothesis(id, run_id, statement, status[open/rejected/confirmed])
evidence(id, run_id, source_type[db/log/api], query_or_call, result_hash, row_count)
claim(id, run_id, hypothesis_id, statement, evidence_ids[], confidence)
verdict(id, run_id, root_cause, resolution, escalated, unsupported_claims)
```

The agent gets **read-only tools** scoped to the synthetic systems: `query_attendance`, `query_payroll`, `get_employee`, `get_payrun_log`, `get_connector_job`, `get_rate_table_version`. It cannot mutate. The verifier re-executes every cited evidence query and checks the `result_hash` — a deterministic check, not a model judging itself.

For the Go implementation, keep it boring: one Fly app with an HTTP API and a separate worker process, both in the same module; `pgx` + `sqlc` against Neon; scenario definitions in Postgres and YAML so you can rerun a fixed seed. Use **deterministic seeded generation** so a scenario is reproducible, and use **Neon branches** to snapshot the database before and after a defect injection.

---

## 4. Score the agent on evidence, not just the answer

The evaluation harness is where this becomes defensible. For every scenario, compute:

- **Root-cause accuracy**: does the final verdict match `ground_truth.root_cause`?
- **Evidence precision**: of the evidence rows cited, how many are in the minimal evidence set?
- **Evidence recall**: did it find the one query or log line that actually proves the cause?
- **Hypothesis discipline**: did it reject the plausible-but-wrong hypotheses (e.g., "connector dropped rows" when the real cause is a rate table)?
- **Unsupported-claim rate**: claims with no evidence ID, or with an evidence ID that doesn't reproduce.
- **Escalation appropriateness**: did it escalate when evidence was insufficient, and not escalate when it had enough?

Track **time-to-first-evidence-backed-diagnosis** and **model cost per investigation** in the same run. This is the number that matters commercially: if a scenario costs $2 in tokens and 90 seconds and beats a human with five SQL tabs, you have a product thesis you can take to a customer.

---

## 5. What this buys you — and what it doesn't

**It buys you:**

- A demo you can show a Head of Support without asking for their production database. You show the investigation record, the rejected hypotheses, and the verified evidence — not a chatbot transcript.
- **CaseTrace Lab earlier than you planned.** Because you generate the scenarios, you have an answer key by construction. The training/evaluation product (Section 6 of your document) is almost free to build on top of this, and it has a buyer who does not need to give you production access. That may be the faster first revenue path while the commercial product waits for trust.
- A way to test whether the investigation loop is actually better than a human with SQL *before* you promise a customer anything.

**It does not buy you:** proof of willingness to pay, and the messy schema/log/API diversity of real customers. Your agent can overfit to your injector — it learns "look for the doubled backpay line" rather than "investigate an unfamiliar discrepancy." Mitigate this with a **held-out defect class** you never tune against, and get a Singapore payroll practitioner to review the scenarios for realism. Keep the manual paid investigation thread alive in parallel; the synthetic world shortens the build, it doesn't replace the first sale.

---

## 6. A 30-day sequence

1. **Days 1–5**: Encode the oracle for CPF, OT, leave, and public holidays. Golden tests against MOM/CPF worked examples. This is your credibility layer.
2. **Days 6–10**: Build the two synthetic services and the connector. Generate 30–50 employees with edge cases: age-55 boundary, SPR year 1, work-pass holder, part-time, mid-month hire, backpay.
3. **Days 11–15**: Write 12–15 scenarios across the defect taxonomy. Each with ground truth and minimal evidence. Manually investigate three of them yourself and write the report by hand — that is your spec for the agent.
4. **Days 16–25**: Build the agent loop and evidence ledger. Start with one hypothesis agent and a deterministic verifier. Add competing hypotheses only when a single-hypothesis run fails.
5. **Days 26–30**: Run the evaluation harness. Publish an internal scorecard: accuracy, evidence recall, unsupported-claim rate, cost per investigation. If the numbers are bad, the problem is the investigation design, not the model — fix that before adding agents.

One rule: **do not build a dashboard in the first 30 days**. The output is a JSON investigation record and a Markdown report. The UI is a distraction until the evidence discipline works.


## Z

It doesn't just make sense — with one reframe, it collapses your entire two-product roadmap into a single build. Let me give you the honest assessment first, then the exact blueprint.

## Why this is smarter than it looks

Don't frame the simulator as "what we build because we can't get customer access." Frame it as what it actually is: **the ground-truth eval harness that nobody in the AI-support space can show, plus CaseTrace Lab, already built.** You can't prove your agent's accuracy on customer data you don't have — but you *can* inject a fault into a synthetic environment, know with certainty what happened, and publish a real score. That converts every demo from "trust me" to "here's the agent resolving 46 of 50 seeded incidents, with evidence." And the same environment is the training/assessment product for BPOs and employers from day one. One artifact, both products. That's the reframe.

The Singapore choice is excellent for a specific reason: the domain is **publicly documented, deterministic, and compact enough to model completely** — which means you can build a reference implementation that acts as a computational oracle. Concretely, the rule packs are all published: MOM's Employment Act rules (44-hour work week, overtime at 1.5× the hourly basic rate, OT eligibility capped at $2,600/month basic salary for non-workmen, payment within 14 days, 72-hour monthly OT cap), CPF Board's age-banded contribution tables with the employee/employer split, the Ordinary Wage ceiling that stepped up to **$8,000 from 1 Jan 2026** (it was $6,800 in 2024, $7,400 in 2025 — and that stepwise schedule is itself incident material), plus SDL and the ethnic funds. The rate changes arriving 1 Jan 2027 for workers above 55 are another gift: **rule-version boundaries are the single most realistic payroll incident class there is.**

## The blueprint: two fake vendors, one integration, one oracle

Build a miniature ecosystem, not a mock. Four fly apps, all Go, all state in one Neon project but separated by **schemas and distinct DB roles** (so "PayLoop" genuinely cannot read "ShiftHub's" tables — the privilege boundary is part of the realism, and your investigation agent later faces the same constraint real L2 engineers face: per-system access, cross-system reasoning).

**1. ShiftHub (WFM/scheduling)** — employees, shift patterns, clock punches (in/out events with device IDs), leave records, OT approvals. REST API. This is the *source of truth for hours worked*.

**2. PayLoop (payroll)** — pay periods, timesheet imports, pay components, CPF/SDL/fund calculations, itemized payslips (the EA requires itemization — encode it). A nightly batch import job with a **mapping config table** (field mappings, transformation rules, filters). This is where symptoms surface.

**3. The sync layer** — the realistic failure surface and where most incidents are born: batch jobs, dedup keys, retry logic, outbox, audit log, config history. Partial failures, duplicates, out-of-order events, clock skew.

**4. The oracle** — an *independent* clean implementation of the SG payroll math that computes expected output from the true underlying punch data. When no fault is injected, oracle ≈ both systems (validates your baseline). When a fault fires, oracle vs. PayLoop's actual output *is* your ground truth diff. This component is non-negotiable and is what makes the whole thing an eval environment instead of a demo.

## The fault catalog: mechanisms, not scripts

This is the design decision that makes or breaks the platform. Never script "ticket #7 has missing OT." Inject a **mechanism** and let symptoms emerge organically — then the same mechanism re-parameterized across tenants, periods, and shift patterns gives you unbounded cases that can't be memorized. First catalog, all traceable to real published rules:

| Mechanism (injector) | Emergent symptom (what the ticket says) | Rule it violates |
|---|---|---|
| Dedup-key collision drops punch events after a clock-firmware update | "8 employees' OT is missing" | OT = 1.5× beyond 44h/week |
| Timezone handling bug attributes post-midnight punches to the wrong day | Friday night shift paid as Saturday rest-day rate (or vice versa) | Rest-day pay rules |
| Mapping filter excludes "pending" OT approvals, silently | Some employees paid, some not, no errors in logs | OT payment within 14 days |
| Partial batch failure: 217 of 250 employees synced, 33 silently skipped | "Half my team got paid wrong" | — (integration integrity) |
| CPF config not updated for the Jan 2026 OW ceiling / rate table | Payslips wrong only for high earners / seniors | CPF tables, $8,000 OW ceiling |
| OT eligibility threshold misapplied to a promoted employee | Overpayment that nobody noticed — the *inverse* incident | $2,600 / $4,500 thresholds |
| Duplicate webhook → double-counted punches | Inflated hours, silent | — |

Each mechanism carries: injection point, difficulty tier, the ground-truth explanation, red herrings to plant (resolved retry storms, unrelated warnings), and a scoring rubric. Seven mechanisms × tenants × pay periods × shift patterns is already hundreds of defensible cases.

## Grading: what makes it sellable

The case harness renders each fault as a realistic complaint ("My employees' overtime is missing from this month's export — worked fine last cycle, we changed nothing"). Investigation happens through a read-only evidence surface: SQL console limited to views, log viewer, API explorer — every action logged. Scoring covers what your earlier document promised: hypothesis discipline, evidence cited vs. claims asserted, whether the candidate found the **mechanism** (not just the symptom), unsupported-claims count, and the escalation decision. The same rubric scores your CaseTrace agent and a human candidate — which means you can publish the comparison table that no competitor can fake: *agent vs. human baseline on N seeded SG-payroll integration incidents*.

## Build order (six weeks, brutal honesty)

1. **Weeks 1–2** — ShiftHub + PayLoop baseline with *correct* payroll math, validated against the oracle on zero-fault data. This is the real engineering; do not rush it.
2. **Week 3** — sync layer + mapping config + audit trail; run the nightly batch for simulated months of history (backfilled so investigations have a "worked fine last week" baseline — essential for realism).
3. **Week 4** — injector + first four mechanisms + ticket renderer.
4. **Week 5** — investigation UI + scoring rubric; run a human baseline (2–3 friendly support engineers if you can find them; otherwise yourself, timed).
5. **Week 6** — point CaseTrace's agent loop at the evidence surface. First benchmark report exists.

## Three guardrails

- **Rules are data, not code.** Encode the SG rulebook as versioned rule packs with effective dates, so a PH or US pack later is a config, not a rewrite. (SG first is right: deterministic, public, English-language; PH next because it plugs straight into your BPO/OJT research.)
- **The dual-purpose test.** Every fault mechanism must be simultaneously a valid eval case *and* a valid training scenario. If something only serves demos, cut it — this is your defense against the classic trap of polishing a beautiful simulator nobody pays for.
- **Timebox at six weeks.** The moment the harness is good enough to score investigations, switch to building the agent and hunting the first pilot. The simulator's job is to make the pilot conversation unnecessary to fake.

When a real customer finally does appear, this environment becomes their sandbox, their team's training ground, and the eval harness for their anonymized cases — so nothing you build gets thrown away.

Want me to start concretely — the Neon schema for ShiftHub/PayLoop with the role separation, the oracle's OT/CPF calculation spec, or the injector + fault-catalog data model? I can scaffold the Go project so week one starts today.



## gemini

This is a masterstroke. Using official regulatory frameworks as your business logic backbone completely solves the cold-start problem of enterprise sales and NDA compliance, while giving you something even better: **deterministic, mathematically rigorous ground truth.**

Relying on Singapore’s Ministry of Manpower (MOM) Employment Act, Central Provident Fund (CPF) Board contribution rules, and public holiday calculation standards provides a bulletproof foundation. Singapore’s business rules are famously strict, well-documented, and full of subtle edge cases that break naive software integrations.

Building a **Synthetic Enterprise Sandbox (SES)** based on these rules creates a high-fidelity environment where you can simulate brutal Tier 2 support tickets.

---

## 1. Why This Architecture Works

Instead of waiting for a real client to hand over production access, you build two decoupled mock SaaS applications in a single Go workspace, backed by Neon PostgreSQL:

* **System A (Workforce Management / Rostering):** Handles employee shift schedules, clock-in/clock-out telemetry, leave applications, and overtime approvals.
* **System B (Payroll & HRIS):** Handles master employee profiles, salary structures, statutory deductions (CPF), and bank disbursement exports (GIRO files).

By forcing these two systems to synchronize via simulated webhooks, scheduled cron jobs, and API polling, you naturally generate the exact class of integration nightmares that plague L2 support teams.

---

## 2. Codifying Singapore Labor Rules into Your Go Engine

To make the platform feel genuinely authentic, embed specific Singapore statutory mechanics into your core Go domain logic:

1. **The MOM Part IV Overtime Thresholds:** The Employment Act dictates strict overtime eligibility (e.g., workmen earning up to $4,500, non-workmen manual workers earning up to $2,600). Overtime is capped at 72 hours a month, paid at 1.5x hourly rate, with 2.0x for rest days and public holidays.
2. **CPF Ordinary Wage (OW) & Additional Wage (AW) Ceilings:** Calculations must respect statutory monthly salary caps and annual ceilings.
3. **Shift Differential & Public Holiday Mapping:** Rules governing when a shift crosses midnight or falls on a gazetted public holiday.

---

## 3. Injecting Realistic L2 Support Scenarios

With this synthetic sandbox, you can programmatically generate realistic multi-system support tickets with known root causes to test CaseTrace:

* **Scenario A: The Mapped Pay-Code Discrepancy**
* *The Ticket:* "Employee John Doe worked 8 hours on Deepavali (a gazetted public holiday), but his payroll export shows normal hours instead of 2.0x holiday pay."
* *The Root Cause:* The WFM system sent event code `PH_WORKED`, but the Payroll system's mapping table was updated last Tuesday, expecting `HOLIDAY_OT_V2`. The webhook payload dropped silently in the middleware retry queue due to a strict schema validation error.


* **Scenario B: The CPF Ceiling Truncation Bug**
* *The Ticket:* "Monthly CPF contribution for a senior engineer is $10 short compared to internal manual calculations."
* *The Root Cause:* An unhandled edge case where an employee received a retroactive backpay adjustment that crossed the monthly Ordinary Wage (OW) ceiling mid-pay-period, causing the calculation engine to misattribute the sum to Additional Wages (AW).



---

## 4. How to Build the MVP on Fly.io and Neon

You can spin this up on your exact stack with zero friction:

* **Neon Branching as a Test Fixture:** Utilize Neon's database branching feature. When your test harness runs a specific L2 scenario, it spins off an isolated database branch pre-loaded with corrupted or desynchronized state, runs the CaseTrace agent against it, and then discards the branch.
* **Fly.io Deployment:** Deploy two separate Go worker services:
1. `ses-simulator`: Continuously runs cron jobs that simulate employee clock-ins, trigger payroll batch runs, and inject random integration failures or network latency.
2. `casetrace-core`: The investigation engine that listens for generated tickets, queries the Neon branch, inspects application audit logs, and produces the evidence-backed investigation report.



---

## chatgpt

# I would turn this into a Singapore Workforce Operations Simulation Platform.

I think you've identified a much stronger approach than waiting for a company to give you access to its production environment.

In fact, I would make this a deliberate product strategy:

Build a functioning, fictional Singapore-based enterprise that operates like a real employer, using real Singapore employment rules, realistic business transactions, simulated integrations, and deliberately engineered operational failures.

Then build CasePilot on top of that environment.

There are two products inside this concept:

1. The simulated enterprise platform: A realistic HR, workforce management, timekeeping, payroll and integration system.
2. The agent-driven Tier-2 support platform: An intelligent investigation system that diagnoses problems occurring inside the simulated enterprise.

The first generates the operational reality. The second investigates it.

This is substantially better than populating a database with random employees, salaries and fabricated support tickets.

## 1. Why Singapore is a particularly good starting point

Singapore has an unusually useful collection of publicly documented employer obligations, calculation rules, operational requirements and government reporting processes.

We can turn these into executable business rules.

Here are some examples I've verified against official government sources.

| Domain               | Officially documented requirements           | What we can simulate                                      |
| -------------------- | -------------------------------------------- | --------------------------------------------------------- |
| Employment Act       | Salary payment, working hours, rest days     | Salary disputes, incorrect working-hour calculations      |
| Overtime             | Eligibility, hourly rates, overtime limits   | Missing or incorrect overtime                             |
| Leave                | Annual leave and sick leave entitlements     | Incorrect accrual and leave balances                      |
| Public holidays      | Holiday entitlement and payment              | Incorrect holiday pay                                     |
| CPF                  | Contributions, wage classification, ceilings | Incorrect CPF deductions                                  |
| Employment records   | Required employee and salary records         | Missing or inconsistent records                           |
| IRAS                 | Employment-income reporting and AIS          | Failed payroll reporting                                  |
| Employment practices | Fair employment principles                   | Incorrect employment classification and workflow controls |

Some important details illustrate how realistic this can become:

- MOM specifies that Part 4 of the Employment Act does not cover managers and executives, and overtime eligibility has additional salary and worker-category conditions.&#x20;

  [image](https://www.google.com/s2/favicons?domain=https://www.mom.gov.sg\&sz=32)

  Ministry of Manpower Singapore

- Annual leave entitlement under the Employment Act begins after three months of service, with the statutory minimum increasing according to years of service.&#x20;

  [image](https://www.google.com/s2/favicons?domain=https://www.mom.gov.sg\&sz=32)

  Ministry of Manpower Singapore

- For 2026, the CPF Ordinary Wage ceiling is S$8,000 per month, and applicable contribution rates depend on age, citizenship status and other eligibility factors.&#x20;

  [image](https://www.google.com/s2/favicons?domain=https://www.cpf.gov.sg\&sz=32)

  How much CPF contributions to pay+1

- Salary generally must be paid within seven days after the salary period, while overtime pay must be paid within fourteen days.&#x20;

  [image](https://www.google.com/s2/favicons?domain=https://www.mom.gov.sg\&sz=32)

  Ministry of Manpower Singapore

- Employers covered by the Employment Act must issue itemised payslips and maintain prescribed employment records.&#x20;

  [image](https://www.google.com/s2/favicons?domain=https://www.mom.gov.sg\&sz=32)

  Ministry of Manpower Singapore+1

- Employers under the Auto-Inclusion Scheme submit employment-income information to IRAS electronically by 1 March each year.&#x20;

  [image](https://www.google.com/s2/favicons?domain=https://www.iras.gov.sg\&sz=32)

  Submit employment income records



These aren't merely documentation topics. They produce calculations, records, deadlines, transactions, dependencies and possible failures.

That is exactly what we need.

One caveat: government guidance needs to be translated into versioned, applicability-aware business rules. We must not treat every rule as applying to every employee or every employer.

# 2. The product architecture: build a fictional but functioning company

I would create a fictional Singapore company called Meridian Workforce Pte. Ltd.

[NCS expands Sunshine.AI with new industry partnerships and regional hiring drive to accelerate sovereign AI deployment | Noah Intelligence](https://images.openai.com/static-rsc-4/IWAq5XOAkmAV-TDVA-21DAkGOCORbHUchv_LOhN387hCBPa0hbRrv3yJDctSLU3eeoZXRQR-yxr2ISsUT3sVDudgeQzmXAVtYo7klAeB10Mkj7blLWiS3UPrCaFLrlVa2QohNfqSWjVzFKP-PiKc88OyKPFrgTEGq7RPr3AiC38?purpose=inline)

[noah-news.com](https://noah-news.com/ncs-expands-sunshineai-with-new-industry-partnerships-and-regional-hiring-drive/?utm_source=chatgpt.com)

This is not merely a demo company. It is our simulated production customer.

It has:

- Several business locations in Singapore.
- Approximately 500 employees.
- Full-time, part-time and shift workers.
- Managers, executives, workmen and non-workmen.
- Singapore citizens, permanent residents and foreign employees.
- Multiple departments.
- Different employment contracts.
- Shift rosters and attendance.
- Leave applications and approvals.
- Payroll periods and payment records.
- CPF contribution calculations.
- Payroll integration with an external HRIS.
- Government reporting workflows.

The platform should behave as though a real company is running its business through it.

The distinction is important.

A synthetic employee should not just be a row containing a name, age and salary.

That employee should have an employment history, organizational relationships, contracts, roster assignments, attendance events, leave applications, payroll transactions, CPF calculations and historical changes.

All these records must agree unless an incident deliberately makes them disagree.

## 3. The functional scope

I would divide Meridian into six business domains.

[School LMS & ERP Software UAE | ENIT Solutions](https://images.openai.com/static-rsc-4/8ZkqSEiSANBgJWnTWbvpLieZ-45ZT7KE9T0Ob8eseAr_0hgr6FKr2pM6qRsWM-4ZqcfxoAR9-6Dw-1X9pTaYt29RVaIqB4JiX-_8XCpNtNxkZa603Mr3jH7IBNhv3b2nl_TNrBrL-38XGjvv6_o7P3OOWiPESk-WDu2wIGGEV8w?purpose=inline)

A. Human Resources Information System (HRIS)

Employee master records, employment contracts, departments, job positions, employment status, compensation, employee lifecycle and organizational hierarchy.

[Humanity Schedule | Workday Marketplace](https://images.openai.com/static-rsc-4/Zduspj_BzhgaRPlokfDw37bbqKfhoaY28ovjpF4_0qXGdGODrCKnFMMd4Pjc3bywpz_Xcfm9_RsfqSxN93TlSN4weY7nQ7YxItW6LyLl_q1VkoeUz9uppylgDemuceDrVUL5roG9iggIRBrIOEccZ4-0MKRWSTgucXAK0qyz5SY?purpose=inline)

B. Workforce Management (WFM)

Shift scheduling, attendance, clock-in/clock-out, break intervals, overtime, rest days, shift swaps and location assignments.

[Payroll Solutions | Self-Service & Automated Payroll | Centrally HR](https://images.openai.com/static-rsc-4/0iJTmoBQMBGZ3Y8HMXD0c_gxPUk1p0cwtL4PAM02FwMuHgTGKCOpxDopJgLkJf1JxRW_YjJFS68hj8PVKzOmfFPC4i5mMzEMkd8KN__2BZMkXyVw-8rntKt97vGNxNZ7oypgOTLYMxJfvoVe2cfJUl2hz7TwLlbYHolnSiBQmcA?purpose=inline)

C. Payroll

Salary components, allowances, deductions, overtime, bonuses, leave-related adjustments, payroll computation, payslips and payment batches.

[Singapore Payroll Compliance: How to Generate CPF E-Submission TXT File in Deskera People?](https://images.openai.com/static-rsc-4/5Zday3YwuZpf9KdOwmTLgvNkg1nU2V5-yniyjuWIDim5c9UKks0YcvUFAUJp9z9U9zMfKGHN28So96wODNkM04twq1pTfwMP2IGQ2wU-w5EWnK_cHNiZKy_r-pvaoTMknGWhO0jktHaJVj8joaJgt5wc87sWy8IgV0XHaHdtHGs?purpose=inline)

D. Statutory Compliance

CPF calculations, wage classification, contribution ceilings, statutory eligibility and IRAS employment-income reporting preparation.

[API Integration Services | TecSAP | TecSAP | TecSAP](https://images.openai.com/static-rsc-4/q8JAeOI4kCOOdlqEX1MEbEkRyP3RcX-f6HoYIOhckFoPp9ouVDMaOrTeAZo07QMCM68hmeWBBnQhSIh0BrurBh9e4u-9hlBwAw8-t85s0FFwFewMLndD5W09aamYEExoGu9CS9KMjwktMoNDufHORg0EXsRGlQVqWthVPJd_cd0?purpose=inline)

E. Integration Platform

HRIS synchronization, payroll exports, external APIs, webhooks, authentication, message queues, retry policies and reconciliation.

[BLDR by Inova AI Solutions | No-Code Enterprise AI OS](https://images.openai.com/static-rsc-4/kojA63n6Wpmp6h4ApZHHhINI8i3-og2YPyPGgEdgDlWnDDw5nG-m0MJVEW3dgChzLLQt0VLQrshKOXcJgOmhmmyGxwjQpH6ImpjmR81sMd4jyLt8sjCtCh1xpjoTERYMesxkAsCENK0gPuACZqWFfUedils6BtYnT2hAhyFYxcg?purpose=inline)

F. Technical Support Operations

Tickets, customer communications, investigation history, logs, SQL investigation, API inspection, incident management and escalation to engineering.

I would not build six independent applications. These are six business domains inside one functioning platform.

## 4. The most important design decision: simulate the entire business lifecycle

I would design the platform around one continuous business process.

## An employee's journey through Meridian

1. Employee hired

   HR creates an employee record, contract and compensation package.
2. Employee assigned to work

   Manager assigns a department, work location and schedule.
3. Employee works

   Clock-in, clock-out, breaks, shift changes and overtime generate attendance events.
4. Employee requests leave

   Leave balance and eligibility are checked, then the manager approves or rejects the request.
5. Timesheet approved

   Attendance and leave records are consolidated into payable working hours.
6. Payroll calculated

   Basic salary, allowances, deductions, overtime and applicable statutory contributions are computed.
7. Payroll finalized

   Payslips are produced, payment batches are created and CPF contributions are prepared.
8. External reporting

   Employment income records are prepared for the appropriate IRAS reporting workflow.

Every step must generate actual database changes, audit events, API calls and integration events.

This produces an important capability:

We can trace a payroll discrepancy all the way back to the employee's original attendance record, roster assignment, contract and configuration.

That is what makes the environment useful for Tier-2 support training.

### Example: one employee, one realistic problem

Suppose an employee receives S$2,800 instead of the expected S$3,200.

The apparent problem is payroll.

But the underlying cause could be:

- A shift was incorrectly assigned.
- An attendance event was missing.
- An approved overtime claim never reached the timesheet.
- The timesheet was approved before the correction.
- The payroll engine used the old approved timesheet.
- The integration worker failed to deliver the updated record.

A realistic investigation requires following the transaction through multiple systems.

I want every ticket to be connected to an actual sequence of business operations, not just a fabricated error message.

## 5. Building realistic data without accessing actual customer databases

This is arguably the most important engineering challenge.

I would build a deterministic business-data generator, not rely on Faker or ask an LLM to invent thousands of employee records.

### Initial simulation population

My starting configuration would be:

| Dataset                   | Initial target |
| ------------------------- | -------------- |
| Legal entities            | 1              |
| Business locations        | 4              |
| Departments               | 8              |
| Employees                 | 500            |
| Employment contracts      | 500+           |
| Historical activity       | 12 months      |
| Shift assignments         | 100,000+       |
| Attendance events         | 200,000+       |
| Payroll periods           | 12             |
| Payroll transactions      | 6,000+         |
| Integration connections   | 5–10           |
| Seeded incident scenarios | 20             |

These are proposed simulation parameters, not claims about an actual Singapore employer.

I would intentionally mix employee characteristics:

- Singapore citizens.
- Permanent residents with different PR tenure.
- Foreign employees.
- Different age groups.
- Managers and executives.
- Workmen and non-workmen.
- Salaried and hourly employees.
- Full-time and part-time arrangements.
- Fixed-term and permanent contracts.
- Employees with different working schedules.

This is important because the same payroll calculation may produce different results depending on the employee's circumstances.

For example, CPF contribution rates vary by age and citizenship status, including the year of Singapore Permanent Resident status.&#x20;

[image](https://www.google.com/s2/favicons?domain=https://www.cpf.gov.sg\&sz=32)

How much CPF contributions to pay+1



### Do not generate disconnected records

I would implement this as a sequence of business generators.

Business data generation pipeline

Employee generator

Employment dates, citizenship, age, job classification

Contract generator

Salary, allowances, working conditions, eligibility

Workforce activity generator

Shifts, time punches, leave, overtime

Payroll generator

Pay calculations, deductions, CPF, payslips

Reconciliation and validation

Verify relationships, calculations, balances and business invariants

The generator needs a fixed random seed so that a particular dataset can be reproduced.

For example:

```
meridian seed \
  --profile singapore-wfm \
  --employees 500 \
  --history 12m \
  --seed meridian-2026-v1
```

Reproducibility is essential when investigating why an agent passed a scenario yesterday but failed it today.

The generator should also distinguish normal business activity from injected failures. Otherwise, we won't know whether an anomaly originated naturally or from our testing framework.

### Important data-safety rule

Every person, company, identifier, bank account and transaction must be synthetic.

Do not generate real NRICs, work-pass numbers, bank details or genuine personal information. Use clearly marked synthetic identifiers.

We can reproduce the data structures without reproducing actual people's identities.

## 6. Government rules must become executable, versioned rules

This is where I would invest considerable engineering effort.

I don't want the LLM deciding how much CPF an employee should pay.

I want Go to calculate it deterministically.

The LLM's job is to investigate and explain the outcome.

### My proposed regulatory architecture

I would introduce a dedicated `Singapore Employment Rules Engine`.

It would have five major rule groups:

| Rule group       | Responsibility                                      |
| ---------------- | --------------------------------------------------- |
| Employment rules | Eligibility, employment categories, working hours   |
| Leave rules      | Annual leave, sick leave, childcare leave           |
| WFM rules        | Overtime, rest days, public holidays                |
| Payroll rules    | Salary components, deductions and payment deadlines |
| Statutory rules  | CPF and IRAS reporting requirements                 |

Each rule would have:

```
Rule ID
Jurisdiction
Government authority
Official source URL
Source verification date
Effective from
Effective until
Applicability conditions
Calculation or evaluation logic
Rule version
Test cases
```

For example:

```
Rule: SG-CPF-OW-2026

Authority: CPF Board

Jurisdiction: Singapore

Effective From: 2026-01-01

Ordinary Wage Ceiling: SGD 8,000

Applicable Population:
  Singapore Citizens and eligible SPRs

Calculation:
  Use applicable CPF contribution tables.

Tests:
  Age boundary
  Citizenship change
  PR status year
  OW ceiling
  AW ceiling
  Employment termination
```

The actual calculation would live in Go, backed by authoritative, versioned rate tables.

### Why effective dates matter

Imagine a payroll transaction for December 2026 that is processed in January 2027.

The platform must distinguish:

- The payroll period.
- The date wages become payable.
- The statutory rules applicable to the transaction.
- Any relevant rule changes effective on 1 January 2027.

CPF Board has already announced changes to some contribution rates effective 1 January 2027. We should therefore build effective-dated rules instead of hardcoding the 2026 figures permanently.&#x20;

[image](https://www.google.com/s2/favicons?domain=https://www.cpf.gov.sg\&sz=32)

How much CPF contributions to pay



This gives us excellent support scenarios involving outdated rules, incorrect deployment dates and incorrect historical recalculations.

### One more critical distinction

Singapore legislation is only one layer of employment policy.

I would model four distinct sources of business decisions:

1. Statutory requirements.
2. Employer policies.
3. Individual employment contracts.
4. Actual system configuration and transactions.

An employer may offer benefits exceeding statutory minimums. Some arrangements also have specific legal conditions.

The platform must evaluate applicability and lawful exceptions rather than simply applying the most generous or most restrictive number.

We should also store the original source URL and a verification history. When government guidance changes, a rule update must go through review and regression testing.

## 7. Make the integrations behave like actual enterprise systems

This is where most Tier-2 support problems become interesting.

A functioning WFM platform is rarely isolated.

It exchanges data with other systems.

I would create five external system simulators.

| Simulated system         | Function                                    |
| ------------------------ | ------------------------------------------- |
| HRIS Partner             | Employee master data and employment changes |
| Attendance Devices       | Physical clock-in and clock-out events      |
| Payroll/Billing Partner  | Payroll exports and reconciliation          |
| CPF Submission Simulator | Contribution submission files and responses |
| IRAS AIS Integration     | Employment income reporting                 |

One particularly interesting opportunity is IRAS.

The published IRAS API specification describes a sandbox environment intended to mimic the production API so developers can test integrations before submitting actual employment-income records. It requires a testing entity, client credentials and sandbox access. The published specification I located is version 1.2 from November 2020, so I would verify the current portal documentation before implementing against it.&#x20;

[image](https://www.google.com/s2/favicons?domain=https://www.iras.gov.sg\&sz=32)

IRAS



IRAS also documents technical integration requirements involving server-to-server requests, HTTP/2, TLS 1.2/1.3 and distinct sandbox and production callback URLs.&#x20;

[image](https://www.google.com/s2/favicons?domain=https://ask.gov.sg\&sz=32)

AskGov



I would use this as a possible real external integration test, but would initially simulate the other statutory submission channels.

### The integrations must fail realistically

For each integration, implement:

- Authentication and token expiration.
- Request and response payloads.
- Schema validation.
- Timeouts.
- HTTP status codes.
- Rate limiting.
- Retry policies.
- Idempotency.
- Duplicate event detection.
- Partial failures.
- Message queues.
- Dead-letter queues.
- Reconciliation.

Consider this workflow:

Employee information synchronization

HRIS sends employee update

Meridian receives webhook

Injected failure

External employee ID mapping is invalid

Synchronization worker retries and eventually moves the event to the dead-letter queue

Employee is missing from downstream payroll

Support ticket generated

"New employee is not appearing in the payroll system."

This is an actual end-to-end investigation exercise.

The agent needs to identify that the HRIS record exists, the webhook was received, the mapping failed and the retry exhausted.

That is considerably more valuable than asking an agent to answer a hypothetical question about HTTP 400 errors.

## 8. Build a fault-injection and incident-generation engine

I would consider this the most important differentiator of the entire simulation.

A realistic production system needs realistic production problems.

I would create an independent `Incident Simulator`.

It should be capable of injecting failures at controlled points in the business lifecycle.

### The first 20 scenarios

These are the scenarios I would prioritize.

HRIS and employee management

1. Newly hired employee does not appear in payroll.
2. Duplicate employee records caused by external ID mismatch.
3. Employee termination is not synchronized.
4. Employee's department or location is incorrect downstream.

Workforce management

5. Clock-in event missing from attendance records.
6. Duplicate clock-in events produce incorrect working hours.
7. Employee works across midnight and receives incorrect time calculations.
8. Overtime hours are missing following a shift change.

Leave and scheduling

9. Annual leave balance is incorrect after a service anniversary.
10. Sick leave eligibility is incorrectly calculated.
11. Public holiday substitution is incorrectly configured.
12. Rest-day work is classified as ordinary working time.

Payroll and CPF

13. Payslip salary differs from approved timesheet.
14. CPF contributions are incorrect after a PR status change.
15. Additional wages are classified incorrectly for CPF.
16. A retrospective salary adjustment is missing from payroll.

Integration and reporting

17. API access token expires and synchronization stops.
18. Webhook is delivered twice, causing duplicate processing.
19. Payroll export fails due to a schema change.
20. IRAS employment-income submission fails due to a validation or integration error.

These are deliberately selected to cross system boundaries.

A good Tier-2 investigation should rarely be solved by looking at only one database table.

### How the scenario engine should work

Every incident must have a complete scenario definition.

For example:

```
scenario_id: SG-WFM-005

title: Missing Employee in Payroll

difficulty: intermediate

environment:
  profile: singapore-wfm

preconditions:
  employee_exists: true
  contract_active: true
  payroll_period_open: true

trigger:
  operation: employee_sync

fault:
  component: integration_worker
  type: invalid_external_mapping

expected_symptoms:
  - employee_missing_from_payroll
  - webhook_received
  - sync_failed

expected_diagnosis:
  category: integration
  root_cause: external_id_mapping

evidence_required:
  - employee_record
  - integration_payload
  - sync_attempt
  - error_response

permitted_actions:
  - inspect
  - query
  - compare
  - recommend

prohibited_actions:
  - modify_production_record
  - bypass_validation
```

The actual implementation can use JSON or Go structs instead of YAML. The example illustrates the scenario contract.

Crucially, the agent must not be allowed to see `expected_diagnosis`.

That information belongs to the evaluation harness, not the investigation environment.

I would keep the answer key separate from the database and tools accessible to the agent.

Otherwise, the agent could simply discover the intended answer instead of investigating.

# 9. The Go implementation

I would build this as a modular monolith, not microservices.

The initial repository structure could look like this:

```
meridian/
├── cmd/
│   ├── api/
│   ├── worker/
│   ├── simulator/
│   └── seed/
│
├── internal/
│   ├── company/
│   ├── employee/
│   ├── employment/
│   ├── workforce/
│   ├── attendance/
│   ├── leave/
│   ├── payroll/
│   ├── cpf/
│   ├── compliance/
│   │   └── singapore/
│   │
│   ├── integration/
│   │   ├── hris/
│   │   ├── attendance/
│   │   ├── payroll/
│   │   ├── cpf/
│   │   └── iras/
│   │
│   ├── event/
│   ├── audit/
│   ├── scenario/
│   ├── fault/
│   ├── support/
│   └── agent/
│
├── migrations/
├── scenarios/
├── seeds/
├── tests/
├── docs/
└── go.mod
```

### Database design

I would use Neon PostgreSQL.

The core database would include:

| Domain        | Important entities                                  |
| ------------- | --------------------------------------------------- |
| Company       | tenants, legal_entities, locations, departments     |
| HRIS          | employees, contracts, employee_status_history       |
| WFM           | shifts, rosters, time_punches, timesheets           |
| Leave         | leave_policies, balances, requests                  |
| Payroll       | pay_periods, payroll_runs, payroll_items, payslips  |
| Compliance    | rule_versions, cpf_assessments, reporting_batches   |
| Integration   | connections, sync_jobs, webhooks, delivery_attempts |
| Observability | application_events, audit_events, integration_logs  |
| Support       | tickets, investigations, evidence, agent_runs       |
| Simulation    | scenario_executions, injected_faults, test_results  |

For historical accuracy, I would make business events append-only where practical, retain audit history, and use explicit effective dates for employment contracts, employee changes and regulatory policies.

I would not implement full event sourcing everywhere. Conventional relational tables plus an append-only audit/event history and transactional outbox are sufficient to start.

### Infrastructure

Deployment architecture

Fly.io

API process · Worker process · Simulator process

Neon PostgreSQL

Operational data · Job state · Events · Audit · Scenarios

Agent and integration runtime

LLM provider · Tool execution · External service simulators

The Go application should use:

- `net/http` or Chi for HTTP.
- pgx and sqlc for data access.
- Goose for migrations.
- OpenTelemetry for traces and instrumentation.
- A PostgreSQL-backed durable job mechanism.
- Structured JSON logging.
- Explicit interfaces for LLM and integration providers.

I would make the API-first approach the priority, with a small server-rendered operational console for inspecting company data, executing scenarios and reviewing tickets. No elaborate frontend framework is necessary.

One non-negotiable requirement: the actual business transactions, integration workers and investigation tools must be usable without the UI.

That is how we prove the platform is a real functioning backend system rather than a visual prototype.

# 10. Build the agent into the platform from the beginning

I would not finish Meridian and only then attach an AI chatbot.

The system should be designed to be investigated by an agent.

Each operational domain should expose typed, permission-controlled investigation tools.

For example:

```
get_employee(employee_id)
get_employment_contract(employee_id)
get_employee_schedule(employee_id)
get_attendance_records(employee_id, period)
get_leave_balance(employee_id)
get_payroll_details(employee_id, period)
get_cpf_assessment(employee_id, period)

get_integration_history(entity_id)
get_webhook_delivery(event_id)
get_application_logs(correlation_id)
get_audit_history(entity_id)
```

These are examples of internal tool contracts, not unrestricted database access.

The agent should be able to:

1. Understand the ticket.
2. Plan an investigation.
3. Retrieve relevant records.
4. Correlate records across systems.
5. Generate competing hypotheses.
6. Collect supporting evidence.
7. Challenge its initial conclusion.
8. Produce a resolution recommendation.

I would start with one Go-controlled orchestrator and specialized stages rather than a free-for-all multi-agent swarm.

### Agent evaluation

Every scenario should have a known expected result, but the agent must discover it independently.

I would measure:

| Metric                   | What it tells us                                                 |
| ------------------------ | ---------------------------------------------------------------- |
| Root cause accuracy      | Did the agent identify the actual problem?                       |
| Evidence completeness    | Did it retrieve the records necessary to support its conclusion? |
| False diagnosis rate     | Did it blame the wrong system?                                   |
| Investigation efficiency | How many tool calls did it need?                                 |
| Resolution quality       | Would a senior support engineer accept the recommendation?       |
| Safety compliance        | Did it respect access boundaries and approval requirements?      |

For the first benchmark, I would create 20 scenarios with different levels of difficulty, each with known ground truth, expected evidence and acceptable alternative explanations.

Then I would run every new agent version against the same scenarios.

This becomes your own regression-testing suite for technical support intelligence.

# 11. The actual implementation roadmap

I would build this in the following order, with strict boundaries on what counts as complete at each stage.

Phase 1 — Singapore employment foundation

First priority

Build:

- Company and employee management.
- Contracts and employment classifications.
- Salary components.
- Singapore statutory rule registry.
- Versioned CPF rates and employment rules.

Acceptance: Create and retrieve an employee with a valid employment history and applicable rule classification.

Phase 2 — Operational WFM

Build:

- Shift scheduling.
- Attendance events.
- Breaks and overtime.
- Leave requests.
- Timesheet approval.

Acceptance: Reconstruct an employee's working history from raw events and derive an approved timesheet.

Phase 3 — Payroll and statutory calculations

Build:

- Payroll runs.
- Payslips.
- Salary deductions.
- CPF calculation.
- Payroll reconciliation.
- IRAS reporting preparation.

Acceptance: Reproduce each payroll result from its underlying business transactions and the exact rule versions used.

Phase 4 — Production-like integrations

Build:

- HRIS partner simulator.
- Attendance device simulator.
- External payroll simulator.
- Webhooks and retry mechanisms.
- Integration logs and dead-letter queue.

Acceptance: Synchronization succeeds normally and produces traceable results when failures occur.

Phase 5 — Incident simulator

Build:

- Scenario definitions.
- Controlled fault injection.
- Automatic ticket creation.
- Deterministic resets.
- Hidden answer keys.

Acceptance: Reproduce the same incident repeatedly and produce the same underlying evidence.

Phase 6 — Agent-driven Tier-2 support

Build:

- Investigation orchestrator.
- Typed investigation tools.
- Evidence correlation.
- Hypothesis verification.
- Resolution recommendations.
- Automated evaluation suite.

Acceptance: Agent independently investigates the seeded tickets and achieves an acceptable score against the hidden ground truth.

# 12. There is an additional commercial opportunity here

I would keep two related product possibilities in mind.

Primary business

# CasePilot

AI-powered technical support investigation

The customer connects CasePilot to their own systems. Meridian is our internal proving ground, demo environment and regression-test platform.

Revenue comes from SaaS subscriptions and enterprise integrations.

Future product

# Meridian Support Lab

A realistic production-support engineering training environment

Students and aspiring technical support engineers receive realistic tickets, investigate SQL records, inspect logs, debug API failures and resolve business problems.

Revenue could come from individual subscriptions, corporate training and partnerships with technical training providers.

This second product has another interesting advantage: the training scenarios are the same scenarios used to evaluate CasePilot.

You are effectively creating a common testing ground for humans and AI agents.

A human support engineer and an AI agent can receive the same ticket, access the same permitted evidence and be evaluated against the same expected diagnosis.

That is a potentially valuable differentiator.

# 13. What I would build first, concretely

I would resist the temptation to immediately implement all six domains.

I would start with one complete vertical slice:

Employee onboarding → Shift assignment → Attendance → Timesheet → Payroll → CPF → Payslip → Integration failure → Tier-2 investigation.

The first milestone should support three deliberately difficult incidents:

Scenario A: Missing employee in payroll

Investigate HRIS synchronization, employee mapping, webhook delivery and downstream records.

Scenario B: Incorrect overtime payment

Investigate employee classification, contractual hours, break deductions, timesheets and applicable Singapore employment rules.

Scenario C: Incorrect CPF contribution

Investigate citizenship status, PR tenure, age, wage classification, salary components and applicable contribution tables.

If the system can realistically reproduce these three problems and an agent can independently investigate them using SQL, APIs and logs, we have established the architecture.

Then we expand the scenario library.

## My final recommendation

I would name the overall environment Meridian SG and treat it as a programmable, synthetic enterprise.

Its defining characteristics should be:

1. Real Singapore government rules, translated into executable and versioned logic.
2. Synthetic but internally consistent employee and business histories.
3. Functional HRIS, WFM and payroll workflows.
4. Realistic third-party integration protocols and failures.
5. Reproducible incidents with hidden ground truth.
6. An agent capable of investigating the same evidence available to a human Tier-2 engineer.

One important boundary: Meridian should be described as a realistic Singapore enterprise simulation, not as a certified payroll or statutory compliance product.

The strategic insight is that you don't need access to a real company's production database to build a credible production-support investigation platform. You need a sufficiently faithful model of the business, the systems that execute its rules, and the failures that occur when those systems interact.

And unlike a conventional software demo, this environment becomes more valuable every time you add a verified failure scenario.

That is what I would build.

The first engineering deliverable I would lock down is a `MERIDIAN_SG_V1_SPEC.md` specification. It should define the domain boundaries, Singapore regulatory source registry, canonical database model, business lifecycle, Go package structure, integration contracts, synthetic-data generation, incident scenario format and acceptance tests. That specification should be the implementation contract for the coding agent—not an invitation to start generating code before the business model is properly defined.
