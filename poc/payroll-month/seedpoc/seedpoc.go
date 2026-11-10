// Package seedpoc loads the Phase 2 shared workload into a disposable
// PostgreSQL database through the same services the CLI and API use
// (AGENTS.md boundary #1; mirrors internal/seed). UNEXECUTED until the
// docker-gated Gate B run: this file compiles in CI but only runs against
// the POC stack (POC-MERIDIAN employer, never MERIDIAN, never shared DBs).
//
// Workload -> service mapping (each row auditable):
//
//	employees/terms ......... hr.CreateEmployee + hr.SetTerms (monthly basic
//	                          = workload base/100 dollars, exact)
//	SYN-012 base change ...... two non-overlapping SetTerms spans (Oct 1-15,
//	                          Oct 16+); the DB records the change faithfully,
//	                          the POC calc blocks it (B-BASE-CHANGE)
//	roster .................. wfm template + 5 weekly rosters + assignments
//	worked minutes .......... attendance.Import CSV (in 09:00, out 09:00+min,
//	                          always paired); zero/leave days emit no clocks
//	leave days .............. leave_request rows (raw SQL: NO leave service
//	                          exists yet — verified; insert mirrors the NOT
//	                          NULL/CHECK columns, status approved; decided
//	                          rows freeze by design)
//	SYN-006 Oct-10 conflict .. clocks AND leave_request both stored: the DB
//	                          faithfully holds conflicting sources; POC policy
//	                          excludes the day (B-LEAVE-WORKED)
//	EV-001 late approval ..... SYN-004 sheet approved LAST (order, not
//	                          backdating: approved_at is DB now(); the Oct-25
//	                          narrative lives only in workload.Events)
//	EV-002 correction ........ attendance.Correct on the Oct-12 interval
//	                          before timesheet Generate (controlled
//	                          succession, never an edit in place)
//	timesheets ............... timesheet.Generate + Submit + Approve;
//	                          SYN-023 (zero clocks) has no sheet by design
//	rule packs ................ NONE activated: SYNTH packs are POC-scoped
//	                          JSON (ActivateStatutory forbids statutory use);
//	                          statutory packs stay BLOCKED via
//	                          ensure_pack_blocked, as in internal/seed.
package seedpoc

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/PithomLabs/a2/internal/apperr"
	"github.com/PithomLabs/a2/internal/attendance"
	"github.com/PithomLabs/a2/internal/hr"
	"github.com/PithomLabs/a2/internal/timesheet"
	"github.com/PithomLabs/a2/internal/wfm"
	"github.com/PithomLabs/a2/poc/payroll-month/workload"
)

const (
	// EmployerCode is POC-scoped: never the demo MERIDIAN.
	EmployerCode = "POC-MERIDIAN"
	// Operator attributes every seeded row.
	Operator = "poc-seed"
	// Period is the canonical pay period code.
	Period = "PP-2026-10"
)

// Seed loads the workload. It refuses when POC-MERIDIAN already exists;
// reload means recycling the disposable database (seed.Reset pattern).
func Seed(ctx context.Context, pool *pgxpool.Pool) (string, error) {
	w := workload.Build()
	var exists bool
	err := pool.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM employer WHERE code = $1)`, EmployerCode).Scan(&exists)
	if err != nil {
		return "", apperr.FromDB(err, "check poc seed state")
	}
	if exists {
		return "", apperr.New(apperr.CodeConflict, "employer POC-MERIDIAN already seeded; recycle the disposable database")
	}
	if err := globals(ctx, pool); err != nil {
		return "", err
	}
	var employerID string
	err = pool.QueryRow(ctx,
		`INSERT INTO employer (code, name, uen, address) VALUES ('POC-MERIDIAN', 'POC Meridian (synthetic)', 'SYNTH-POC-001', '1 Synthetic Row, Singapore') RETURNING id`,
	).Scan(&employerID)
	if err != nil {
		return "", apperr.FromDB(err, "seed poc employer")
	}
	empIDs, err := people(ctx, pool, w, employerID)
	if err != nil {
		return "", err
	}
	if err := structure(ctx, pool, w, employerID); err != nil {
		return "", err
	}
	if err := compLines(ctx, pool, w, employerID, empIDs); err != nil {
		return "", err
	}
	if err := rosters(ctx, pool, w, employerID); err != nil {
		return "", err
	}
	csv := clocksCSV(w)
	if _, err := attendance.Import(ctx, pool, employerID, strings.NewReader(csv), Operator, timesheet.RegenerateDraft, timesheet.ChainTerminal); err != nil {
		return "", err
	}
	if err := leave(ctx, pool, w, employerID, empIDs); err != nil {
		return "", err
	}
	if _, err := pool.Exec(ctx,
		`INSERT INTO pay_period (employer_id, code, starts_on, ends_on, pay_date, salary_deadline, ot_deadline)
		 VALUES ($1, $2, '2026-10-01', '2026-10-31', '2026-11-07', '2026-11-07', '2026-11-14')`,
		employerID, Period); err != nil {
		return "", apperr.FromDB(err, "seed poc pay period")
	}
	if err := correctEV002(ctx, pool, employerID); err != nil {
		return "", err
	}
	gen, err := timesheet.Generate(ctx, pool, employerID, Period, Operator)
	if err != nil {
		return "", err
	}
	// EV-001: approve SYN-004's sheet last (late-approval order).
	var late []timesheet.SheetSummary
	for _, no := range employeeNos(w) {
		sheets, err := timesheet.ListTimesheets(ctx, pool, employerID, Period, no)
		if err != nil {
			return "", err
		}
		for _, s := range sheets {
			if no == "SYN-004" {
				late = append(late, s)
				continue
			}
			if _, err := timesheet.Submit(ctx, pool, employerID, s.ID, Operator); err != nil {
				return "", err
			}
			if _, err := timesheet.Approve(ctx, pool, employerID, s.ID, Operator); err != nil {
				return "", err
			}
		}
	}
	for _, s := range late {
		if _, err := timesheet.Submit(ctx, pool, employerID, s.ID, Operator); err != nil {
			return "", err
		}
		if _, err := timesheet.Approve(ctx, pool, employerID, s.ID, Operator); err != nil {
			return "", err
		}
	}
	// Idempotency proof: replaying the clock file stores nothing new.
	rep, err := attendance.Import(ctx, pool, employerID, strings.NewReader(csv), Operator, timesheet.RegenerateDraft, timesheet.ChainTerminal)
	if err != nil {
		return "", err
	}
	if rep.EventsStored != 0 || rep.Duplicates == 0 {
		return "", apperr.New(apperr.CodeInternal, "poc seed replay was not idempotent: %+v", rep)
	}
	for _, p := range []string{"SG-MOM-OT", "SG-MOM-SHIFT", "SG-CPF-2026", "SG-SDL", "SG-SHG", "SG-MOM-LEAVE", "SG-MOM-KET", "SG-MOM-PH", "SG-MOM-PAYDEAD"} {
		if _, err := pool.Exec(ctx, `SELECT ensure_pack_blocked($1::uuid, $2)`, employerID, p); err != nil {
			return "", apperr.FromDB(err, "seed poc activations")
		}
	}
	return fmt.Sprintf("seeded POC-MERIDIAN: 25 employees, 5 roster weeks, PP-2026-10, %d sheets (%d skipped)",
		len(gen.Generated), len(gen.Skipped)), nil
}

func employeeNos(w workload.Workload) []string {
	var out []string
	for _, e := range w.Employees {
		out = append(out, e.ID)
	}
	return out
}

// globals inserts the catalogue rows the POC uses (idempotent).
func globals(ctx context.Context, pool *pgxpool.Pool) error {
	if _, err := pool.Exec(ctx,
		`INSERT INTO skill (code, name, kind) VALUES ('EN-SUPPORT', 'English support', 'language')
		 ON CONFLICT (code) DO NOTHING`); err != nil {
		return apperr.FromDB(err, "seed poc skills")
	}
	for _, p := range [][4]string{
		{"REG", "Regular", "earning", "OW"},
		{"OT15", "Overtime 1.5x", "earning", "OW"},
		{"LEAVE", "Leave", "leave", "OW"},
		{"ALLOW", "Allowance", "earning", "OW"},
		{"DEDUCT", "Deduction", "info", "NON_CPF_WAGE"},
	} {
		if _, err := pool.Exec(ctx,
			`INSERT INTO pay_code (code, name, kind, default_treatment) VALUES ($1, $2, $3, $4)
			 ON CONFLICT (code) DO NOTHING`, p[0], p[1], p[2], p[3]); err != nil {
			return apperr.FromDB(err, "seed poc pay codes")
		}
	}
	return nil
}

// people creates the 25 employees with terms. SYN-012 gets two
// non-overlapping terms spans recording the mid-month base change.
func people(ctx context.Context, pool *pgxpool.Pool, w workload.Workload, employerID string) (map[string]string, error) {
	ids := map[string]string{}
	for _, e := range w.Employees {
		emp, err := hr.CreateEmployee(ctx, pool, employerID, hr.EmployeeInput{
			EmployeeNo: e.ID, LegalName: "Synthetic " + e.ID, Address: "1 Synthetic Row",
			Gender: "X", DateOfBirth: "1990-01-01", Citizenship: "SG",
			Status: hr.StatusActive, StatusEffectiveDate: "2026-01-01",
		}, Operator)
		if err != nil {
			return nil, err
		}
		ids[e.ID] = emp.ID
		if err := wfm.AssignSkills(ctx, pool, employerID, e.ID, "2026-01-01", []wfm.SkillGrant{
			{Skill: "EN-SUPPORT", Level: 3},
		}, Operator); err != nil {
			return nil, err
		}
		terms := func(from string, to *string, dollars string) hr.TermsInput {
			return hr.TermsInput{
				EffectiveFrom: from, EffectiveTo: to,
				JobTitle: "Support Agent", JobDuties: "Tier 1 support",
				EmploymentCategory: "non_workman", PartIVEligible: true, EligibilityBasis: "synthetic poc data",
				MonthlyBasic: dollars, WorkingArrangement: "shift", NormalHoursWeekly: 44,
				WorkingDays: "Mon-Fri", DailyHours: 8, RestDay: "Sunday",
				StartDate: "2026-01-01", ContractType: "permanent",
				PlaceOfWork: "Singapore", SalaryPeriod: "monthly",
			}
		}
		if e.ID == "SYN-012" {
			oct15 := "2026-10-15"
			if _, err := hr.SetTerms(ctx, pool, employerID, emp.ID, terms("2026-10-01", &oct15, "3280"), Operator); err != nil {
				return nil, err
			}
			if _, err := hr.SetTerms(ctx, pool, employerID, emp.ID, terms("2026-10-16", nil, "3360"), Operator); err != nil {
				return nil, err
			}
			continue
		}
		if _, err := hr.SetTerms(ctx, pool, employerID, emp.ID,
			terms("2026-01-01", nil, fmt.Sprintf("%d", e.MonthlyBaseCents/100)), Operator); err != nil {
			return nil, err
		}
	}
	return ids, nil
}

// compLines persists the workload as-given allowances/deductions as
// compensation_line rows via the authorized add_compensation writer
// (raw INSERT is revoked by 00011). One line per nonzero amount,
// effective for the whole of Oct 2026. SYN-012 is SKIPPED: its two
// terms spans each contain only half the month, and a full-month
// as-given line fits neither span (containment trigger, verified
// live); splitting the amount would invent proration, which is
// BLOCKED. SYN-012 captures as excluded (B-BASE-CHANGE) and its
// as-given sums never enter any calculation, so no calc input is
// lost. Without this step the DB-backed calc cannot reproduce the
// canonical as-given sums from persisted records (RC-1 repair).
func compLines(ctx context.Context, pool *pgxpool.Pool, w workload.Workload, employerID string, empIDs map[string]string) error {
	for _, e := range w.Employees {
		if e.Blocked != nil {
			continue
		}
		var termsID string
		err := pool.QueryRow(ctx,
			`SELECT id FROM employment_terms WHERE employer_id = $1 AND employee_id = $2
			 ORDER BY effective_from DESC LIMIT 1`,
			employerID, empIDs[e.ID]).Scan(&termsID)
		if err != nil {
			return apperr.FromDB(err, "seed comp terms lookup "+e.ID)
		}
		for _, ag := range e.AsGiven {
			if ag.Cents == 0 {
				continue
			}
			amount := fmt.Sprintf("%d.%02d", ag.Cents/100, ag.Cents%100)
			var id string
			err := pool.QueryRow(ctx,
				`SELECT add_compensation($1, $2, $3, $4, $5, $6, '2026-10-01', '2026-10-31', $7, 'poc-seed as-given')`,
				employerID, empIDs[e.ID], termsID, ag.Kind, ag.Label, amount, Operator).Scan(&id)
			if err != nil {
				return apperr.FromDB(err, "seed comp "+ag.Kind+" "+e.ID)
			}
		}
	}
	return nil
}

// structure creates one client, programme, team and all memberships.
func structure(ctx context.Context, pool *pgxpool.Pool, w workload.Workload, employerID string) error {
	c, err := wfm.CreateClient(ctx, pool, employerID, "POCME", "POC Demo Client")
	if err != nil {
		return err
	}
	p, err := wfm.CreateProgramme(ctx, pool, employerID, wfm.ProgrammeInput{
		Client: c.Code, Code: "POCME-CARE", Name: "POC Care", Timezone: "Asia/Singapore",
	})
	if err != nil {
		return err
	}
	if err := wfm.RequireSkills(ctx, pool, employerID, p.Code, []wfm.SkillGrant{{Skill: "EN-SUPPORT", Level: 2}}, Operator); err != nil {
		return err
	}
	t, err := wfm.CreateTeam(ctx, pool, employerID, wfm.TeamInput{Programme: p.Code, Code: "POC-A", Name: "POC Team A"})
	if err != nil {
		return err
	}
	for _, no := range employeeNos(w) {
		if _, err := wfm.AddMembership(ctx, pool, employerID, t.Code, no, "agent", "2026-01-01", Operator); err != nil {
			return err
		}
	}
	return nil
}

// rosters publishes 5 weekly rosters covering October 2026 (weeks starting
// Sep 28, Oct 5, 12, 19, 26): DAY template on worked days, rest otherwise.
func rosters(ctx context.Context, pool *pgxpool.Pool, w workload.Workload, employerID string) error {
	day, err := wfm.CreateTemplate(ctx, pool, employerID, wfm.TemplateInput{
		Code: "POCDAY", Name: "POC day shift", StartLocal: "09:00", EndLocal: "18:00",
		Segments: []wfm.SegmentInput{
			{Kind: "work", StartOffsetMin: 0, DurationMin: 480, Paid: true},
		},
	})
	if err != nil {
		return err
	}
	worked := map[string]map[string]bool{}
	for _, e := range w.Employees {
		worked[e.ID] = map[string]bool{}
		for _, d := range e.Days {
			worked[e.ID][d.Day] = d.WorkedMin > 0
		}
	}
	weeks := []struct {
		start string
		days  []string
	}{
		{"2026-09-28", []string{"2026-10-01", "2026-10-02", "2026-10-03", "2026-10-04"}},
		{"2026-10-05", []string{"2026-10-05", "2026-10-06", "2026-10-07", "2026-10-08", "2026-10-09", "2026-10-10", "2026-10-11"}},
		{"2026-10-12", []string{"2026-10-12", "2026-10-13", "2026-10-14", "2026-10-15", "2026-10-16", "2026-10-17", "2026-10-18"}},
		{"2026-10-19", []string{"2026-10-19", "2026-10-20", "2026-10-21", "2026-10-22", "2026-10-23", "2026-10-24", "2026-10-25"}},
		{"2026-10-26", []string{"2026-10-26", "2026-10-27", "2026-10-28", "2026-10-29", "2026-10-30", "2026-10-31"}},
	}
	for _, wk := range weeks {
		r, err := wfm.CreateRoster(ctx, pool, employerID, wfm.RosterInput{Programme: "POCME-CARE", WeekStarting: wk.start})
		if err != nil {
			return err
		}
		for _, no := range employeeNos(w) {
			for _, d := range wk.days {
				in := wfm.AssignmentInput{Employee: no, Date: d}
				if worked[no][d] {
					in.Template = day.Code
				} else {
					in.RestDay = true
				}
				if _, err := wfm.Assign(ctx, pool, employerID, r.ID, in); err != nil {
					return err
				}
			}
		}
		if _, err := wfm.Publish(ctx, pool, employerID, r.ID, Operator); err != nil {
			return err
		}
	}
	return nil
}

// clocksCSV renders one paired in/out per worked day (09:00 in, 09:00+min
// out). Deterministic event IDs make the replay-idempotency proof exact.
func clocksCSV(w workload.Workload) string {
	var b strings.Builder
	b.WriteString("source,source_event_id,employee_no,occurred_at,timezone,kind\n")
	n := 0
	for _, e := range w.Employees {
		for _, d := range e.Days {
			if d.WorkedMin <= 0 || (d.OnUnpaidLeave && d.Class == workload.Calculable) {
				continue
			}
			n++
			fmt.Fprintf(&b, "poc-clock,poc-%05d,%s,%sT09:00:00+08:00,Asia/Singapore,in\n", n, e.ID, d.Day)
			n++
			out := 9*60 + d.WorkedMin
			fmt.Fprintf(&b, "poc-clock,poc-%05d,%s,%sT%02d:%02d:00+08:00,Asia/Singapore,out\n",
				n, e.ID, d.Day, out/60, out%60)
		}
	}
	return b.String()
}

// leave inserts approved single-day leave rows. Raw SQL is used ONLY because
// no leave service exists yet (verified: no Go writer references
// leave_request); columns mirror the table's NOT NULL/CHECK surface.
func leave(ctx context.Context, pool *pgxpool.Pool, w workload.Workload, employerID string, empIDs map[string]string) error {
	for _, e := range w.Employees {
		for _, d := range e.Days {
			if !d.OnUnpaidLeave {
				continue
			}
			if _, err := pool.Exec(ctx,
				`INSERT INTO leave_request (id, employer_id, employee_id, type, from_date, to_date, days, status, reason, approved_by, approved_at)
				 VALUES (gen_random_uuid(), $1::uuid, $2::uuid, 'UNPAID', $3::date, $3::date, 1, 'approved', 'synthetic poc fixture', 'poc-seed', now())`,
				employerID, empIDs[e.ID], d.Day); err != nil {
				return apperr.FromDB(err, "seed poc leave")
			}
		}
	}
	return nil
}

// correctEV002 applies the scripted Oct-12 interval correction for SYN-009
// (out 17:00 -> 18:00) before timesheet Generate, via controlled succession.
func correctEV002(ctx context.Context, pool *pgxpool.Pool, employerID string) error {
	ivs, err := attendance.ListIntervals(ctx, pool, employerID, "SYN-009", "2026-10-12", "2026-10-12")
	if err != nil {
		return err
	}
	// Match structurally: timestamptz::text renders in the session
	// timezone, so timestamp substrings are not portable. Oct-12 holds
	// exactly one current work interval (single in/out pair).
	var target *attendance.IntervalDetail
	for i := range ivs {
		iv := &ivs[i]
		if iv.BusinessDate == "2026-10-12" && iv.Kind == "work" && iv.IsCurrent {
			if target != nil {
				return apperr.New(apperr.CodeInvalidState, "EV-002: multiple current work intervals on Oct-12")
			}
			target = iv
		}
	}
	if target == nil {
		return apperr.New(apperr.CodeInvalidState, "EV-002: SYN-009 Oct-12 current work interval not found")
	}
	_, err = attendance.Correct(ctx, pool, employerID, attendance.CorrectionInput{
		IntervalIDs: []string{target.ID},
		Start:       "2026-10-12T09:00:00+08:00",
		End:         "2026-10-12T18:00:00+08:00",
		Kind:        "work",
		Reason:      "EV-002: fixture correction, out punch was 17:00",
	}, Operator, timesheet.RegenerateDraft, timesheet.CorrectionDrafts)
	return err
}
