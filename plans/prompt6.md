## prompt

wait, before doing prompt 7, the agent said Advancing the SG-MOM overtime package toward pinned MOM sources while blocked, does this mean we have non-functional overtime package?

## answer

Yes. We currently have a non-functional overtime package in the sense that it cannot calculate or return payable overtime amounts. But it is not useless or empty. It is a formal research and specification package that is deliberately prevented from producing potentially incorrect payroll results.

There are three distinct levels here:
Level	Current status
Research package	Exists
Formal business-rule specification	Exists, but unresolved
Executable overtime calculation in Go	Does not exist

The evidence is explicit:
1. manifest.json identifies SG-MOM-OT-001 as BLOCKED, with unresolved threshold and rounding issues. 
   manifest(8).json
2. contract.md states that the entire rule is an indeterminate path and that no Go code may compute it. 
   contract.md
3. AGENTS.md requires L1 rules to reach GOLDENS_VERIFIED before Go implementation begins, and IMPLEMENTABLE before activation. 
   AGENTS(20261108-091904).md
4. The MOM material currently available consists of provisional transcripts, not byte-grade verified legislative sources. 
   prompt5_report.md
What does "advancing toward pinned MOM sources" mean?
It means we are improving the evidence supporting the overtime rule, not building the overtime calculator.

The package currently knows the general calculation structure:
\[
OT = \text{Hourly Basic Rate}\times 1.5\times
\text{Overtime Hours}
\]

But it cannot safely determine all the inputs and applicable conditions. In particular, the S$13.60 versus S$13.63636 issue remains unresolved, as does the legislative history of the S$2,600 threshold.

There is also the unresolved treatment of the 72-hour limit, which must not be mistaken for a cap on payable hours.
Does this prevent Prompt 7?
Not entirely. We need to distinguish two deliverables:

Prompt 7's HR/WFM foundation can proceed: employee records, rosters, clock events, attendance, approved timesheets and payroll readiness.

A functioning payroll engine cannot be considered complete without verified and implemented statutory calculations. Prompt 7 explicitly defers those calculations.

My recommendation is to proceed with Prompt 7 only as the first operational backend milestone, while keeping the overtime package on a tightly controlled path to implementation.

But I would change one expectation: we should not call the integrated HR/WFM/payroll MVP complete until at least one real, independently verified payroll calculation works end to end.

That is the distinction we need to maintain: we can build the business system while researching the law, but we cannot claim to have built functioning Singapore payroll merely because a payroll-readiness endpoint works.