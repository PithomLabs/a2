# Dynamic-rules classification table (Z#5, Gate A)

Every hardcoded rule → (i) effective-dated PG18 data, (ii) versioned attested
pack, or (iii) code in a versioned durable step. Pack coverage is enforced by
DB constraint (PERIOD FK), not app code. Statuses: DONE (already that way),
POC (this programme), OPEN (needs ruling/semantics), UNTOUCHED (out of scope).

| # | Rule | Current home | Class | Disposition |
|---|---|---|---|---|
| 1 | OT parameters | `mom_ot_params` eff-dated (`00003`, UNIQUE key+from) DOCUMENTED | (i) | DONE (data); formula → row 2 |
| 2 | OT formula (SG-MOM-OT-001, BLOCKED) | `rules/SG-MOM-OT-001/` DOCUMENTED | (ii) | UNTOUCHED (statutory track separate) |
| 3 | Shift params | `mom_shift_params` eff-dated (`00003`) DOCUMENTED | (i) | DONE (data); formula → (ii) UNTOUCHED |
| 4 | CPF/SDL/SHG rates | `cpf_rate_band`, `cpf_ceiling`, `sdl_params`, `shg_instruction` tables DOCUMENTED | (i) params | DONE (data); formulas → (ii) UNTOUCHED |
| 5 | Comp basis carve + resolve | `internal/compbasis/resolve.go:72,129` DOCUMENTED | (iii) + (i) inputs | POC: (iii) in versioned step; `employment_terms` eff-dated (i) DONE (`00001`); liveness stays code+trigger (PG18 cannot express liveness) |
| 6 | Employment eligibility basis | `employment_terms` cols + CHECKs (`00001`) DOCUMENTED | (i) | DONE (data); derivation → (iii) POC |
| 7 | Run lifecycle (draft→…→paid; decided runs reject snapshots) | `internal/payroll/runs.go:112` + 00012 triggers DOCUMENTED | (i) state+triggers + (iii) orchestration | POC: triggers kept; orchestration in month workflow |
| 8 | Timesheet approval barrier | `check_timesheet_approval_barrier()` trigger (`00011`) DOCUMENTED | (i) | DONE, kept (L-list) |
| 9 | Supersession/correction lineage | `payroll_correction`, `attendance_correction_group_predecessor` DOCUMENTED | (i) lineage data | DONE (shape); rule → (iii) POC |
| 10 | Pack activation + coverage | `rule_pack(+_activation,+_history)` DOCUMENTED | (i) partial | POC: ADD PERIOD FK coverage enforcement (step 2) |
| 11 | Timesheet/roster/clock aggregation | `timesheet_line_interval`, `roster_revision`, `clock_event_disposition` DOCUMENTED | (i) data + (iii) formulas | POC: formulas in versioned steps |
| 12 | KET issuance content | statutory output (KET tables) INFERRED | (ii) when statutory | UNTOUCHED |
| 13 | SYNTH-OT/SYNTH-LEAVE params | `poc/.../synth.go` RulePack DOCUMENTED | (ii)-as-data | POC: versions as data; A10 proves data-only variant |
| 14 | SYNTH formulas | `poc/.../synth.go` ComputePay DOCUMENTED | (iii)-POC | POC: new TYPE needs Go change (A10 reports) |
| 15 | Proration / X-FUND / mid-month comp semantics | OPEN (R7 deferred) | UNPLACED | OPEN: F1b red by design until settled |

## Pack coverage enforcement (the DB rule)

`rule_pack_activation` rows carry effective spans; every computation-need row
(run, snapshot build) references its pack via `(pack_key, need PERIOD)`
against the pack's `(pack_key, eff WITHOUT OVERLAPS)` key. Gaps fail at the
FK — `TestCoverage_BlockedByConstraint` pins this. Coverage is therefore a
database property, and row 10's "ADD" is additive DDL on disposable DBs
(POC step 2), never an app-code check.
