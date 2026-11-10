package dbosflow

// Snapshot-to-payroll calculation (prompt8 Stage C).
//
// CalcMember maps ONE persisted snapshot member to the existing
// synth.ComputePay inputs using the SAME rules as workload.Canonical:
// excluded members are skipped (blocked), excluded days are skipped
// (day exceptions), every other day becomes a synth.DayInput. No
// payroll logic is duplicated: ComputePay is the only calculator,
// and MonthPayrollDigest (shared) proves DB-backed == canonical.
//
// Amounts: payroll_line rows carry SIGNED cents (deductions
// negative); net = SUM(amount). CPF treatments are a POC mapping
// (non-statutory): earnings OW, leave/deduction adjustments
// NON_CPF_WAGE.

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/PithomLabs/a2/poc/payroll-month/synth"
	"github.com/PithomLabs/a2/poc/payroll-month/workload"
)

// CalcVersion pins the calculator the DEFINER persist binds to the run.
const CalcVersion = "poc-synth/v1"

// SnapDay is one captured day (see poc_snapshot_capture inputs JSON).
type SnapDay struct {
	Date     string `json:"date"`
	Minutes  int    `json:"minutes"`
	Leave    bool   `json:"leave"`
	Excluded bool   `json:"excluded"`
}

// SnapInputs is the captured per-member inputs JSON.
type SnapInputs struct {
	EmployeeNo       string    `json:"employee_no"`
	MonthlyBaseCents *int64    `json:"monthly_base_cents"`
	TermsIDs         []string  `json:"terms_ids"`
	SheetID          *string   `json:"sheet_id"`
	Days             []SnapDay `json:"days"`
	AllowCents       int64     `json:"allow_cents"`
	DeductCents      int64     `json:"deduct_cents"`
}

// SnapMember is one persisted snapshot member row.
type SnapMember struct {
	EmployeeID string
	EmployeeNo string
	Decision   string
	Blockers   []workload.Block
	Inputs     SnapInputs
	RowHash    string
}

// SnapPin is one persisted pack pin.
type SnapPin struct {
	Pack      string
	Version   string
	Status    string
	Digest    string
	ParamsRaw []byte // schedule params JSON (RulePack shape)
}

// LoadSnapshot reads the valid tip's members + pins for a run.
// Read-only: SELECTs the app role is granted.
func LoadSnapshot(ctx context.Context, pool *pgxpool.Pool, runID string) ([]SnapMember, []SnapPin, error) {
	var snapID, employerID string
	err := pool.QueryRow(ctx,
		`SELECT id, employer_id FROM payroll_snapshot WHERE payroll_run_id = $1 AND status = 'valid'`,
		runID).Scan(&snapID, &employerID)
	if err != nil {
		return nil, nil, fmt.Errorf("calc: valid tip for run %s: %w", runID, err)
	}
	rows, err := pool.Query(ctx,
		`SELECT e.employee_id, e.decision, e.blockers, e.inputs, e.row_hash, m.employee_no
		   FROM payroll_snapshot_employee e JOIN employee m ON m.id = e.employee_id
		  WHERE e.snapshot_id = $1 ORDER BY m.employee_no`, snapID)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	var members []SnapMember
	for rows.Next() {
		var mb SnapMember
		var blockersRaw, inputsRaw []byte
		if err := rows.Scan(&mb.EmployeeID, &mb.Decision, &blockersRaw, &inputsRaw, &mb.RowHash, &mb.EmployeeNo); err != nil {
			return nil, nil, err
		}
		if err := json.Unmarshal(blockersRaw, &mb.Blockers); err != nil {
			return nil, nil, fmt.Errorf("calc: blockers %s: %w", mb.EmployeeNo, err)
		}
		if err := json.Unmarshal(inputsRaw, &mb.Inputs); err != nil {
			return nil, nil, fmt.Errorf("calc: inputs %s: %w", mb.EmployeeNo, err)
		}
		members = append(members, mb)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}
	prows, err := pool.Query(ctx,
		`SELECT p.pack, p.version, p.status, p.manifest_digest, s.params
		   FROM payroll_snapshot_pack p
		   JOIN poc_rule_pack_schedule s
		     ON s.employer_id = $2 AND s.pack = p.pack AND s.version = p.version
		  WHERE p.snapshot_id = $1 ORDER BY p.effective_from, p.pack`, snapID, employerID)
	if err != nil {
		return nil, nil, err
	}
	defer prows.Close()
	var pins []SnapPin
	for prows.Next() {
		var pn SnapPin
		if err := prows.Scan(&pn.Pack, &pn.Version, &pn.Status, &pn.Digest, &pn.ParamsRaw); err != nil {
			return nil, nil, err
		}
		pins = append(pins, pn)
	}
	if err := prows.Err(); err != nil {
		return nil, nil, err
	}
	return members, pins, nil
}

// PacksToRulePacks parses pinned schedule params into synth packs.
// Params JSON uses the exact synth.RulePack field tags.
func PacksToRulePacks(pins []SnapPin) ([]synth.RulePack, error) {
	var packs []synth.RulePack
	for _, pn := range pins {
		var rp synth.RulePack
		if err := json.Unmarshal(pn.ParamsRaw, &rp); err != nil {
			return nil, fmt.Errorf("calc: pack %s params: %w", pn.Version, err)
		}
		if rp.Version != pn.Version {
			return nil, fmt.Errorf("calc: pin %s binds params %s", pn.Version, rp.Version)
		}
		packs = append(packs, rp)
	}
	return packs, nil
}

// CalcMember runs one included member through synth.ComputePay.
// Excluded members return a BlockedEmp (no financial effect).
// Day exclusion mirrors Canonical: excluded days are skipped and
// recorded as day exceptions (never silently processed, never guessed).
func CalcMember(mb SnapMember, packs []synth.RulePack, allowCents, deductCents int64) (workload.EmpResult, *workload.BlockedEmp, []workload.Exception, error) {
	if mb.Decision != "included" {
		code := "B-UNKNOWN"
		if len(mb.Blockers) > 0 {
			code = mb.Blockers[0].Code
		}
		return workload.EmpResult{}, &workload.BlockedEmp{
			ID:    mb.EmployeeNo,
			Block: workload.Block{Code: code},
		}, nil, nil
	}
	if mb.Inputs.MonthlyBaseCents == nil {
		return workload.EmpResult{}, nil, nil,
			fmt.Errorf("calc: included member %s has no monthly base", mb.EmployeeNo)
	}
	var inputs []synth.DayInput
	var exc []workload.Exception
	for _, d := range mb.Inputs.Days {
		if d.Excluded {
			exc = append(exc, workload.Exception{
				Scope: "day", Ref: mb.EmployeeNo + "/" + d.Date,
				Code: workload.CodeLeaveWorked,
			})
			continue
		}
		inputs = append(inputs, synth.DayInput{
			Day: d.Date, WorkedMin: d.Minutes, OnUnpaidLeave: d.Leave,
			MonthlyBaseCents: *mb.Inputs.MonthlyBaseCents,
		})
	}
	res, err := synth.ComputePay(inputs, packs)
	if err != nil {
		return workload.EmpResult{}, nil, nil,
			fmt.Errorf("calc: ComputePay %s: %w", mb.EmployeeNo, err)
	}
	return workload.EmpResult{
		ID: mb.EmployeeNo, MonthlyBaseCents: *mb.Inputs.MonthlyBaseCents,
		Days: res.Days, TotalOTCents: res.TotalOTCents,
		TotalDedCents: res.TotalDedCents, AllowCents: allowCents,
		DeductCents: deductCents,
		NetCents: *mb.Inputs.MonthlyBaseCents + res.TotalOTCents -
			res.TotalDedCents + allowCents - deductCents,
		Digest: res.Digest,
	}, nil, exc, nil
}

// PersistLine is one result line in poc_result_persist's JSON shape.
// Order is canonical: members in LoadSnapshot order (employee_no),
// components in fixed order (basic, overtime, leave, allowance,
// deduction), zero-amount lines omitted. The persist DEFINER and
// the auditor's canon digest both depend on this exact order.
type PersistLine struct {
	EmployeeNo         string         `json:"employee_no"`
	Component          string         `json:"component"`
	PayCode            string         `json:"pay_code"`
	Hours              *float64       `json:"hours"`
	AmountCents        int64          `json:"amount_cents"`
	CPFTreatment       string         `json:"cpf_treatment"`
	WorkPeriod         string         `json:"work_period"`
	PayableDate        string         `json:"payable_date"`
	ContributionPeriod string         `json:"contribution_period"`
	SourceRefs         map[string]any `json:"source_refs"`
}

// PersistSlip is one payslip in poc_result_persist's JSON shape.
type PersistSlip struct {
	EmployeeNo string         `json:"employee_no"`
	Payload    map[string]any `json:"payload"`
}

// BuildPersistLines joins calc results back to snapshot members
// (CalcMonth preserves member order). Shared by the workflow's
// Stage C and the independent auditor (prompt8 §5.2): the auditor
// re-derives these rows from persisted inputs and compares them
// field-by-field against persisted outputs — shared pure mapping,
// independent data.
func BuildPersistLines(members []SnapMember, pins []SnapPin, m workload.Month, payDate string) ([]PersistLine, []PersistSlip) {
	packVers := map[string]any{}
	for _, pn := range pins {
		packVers[pn.Pack] = pn.Version
	}
	var lines []PersistLine
	var slips []PersistSlip
	er := 0
	for _, mb := range members {
		if mb.Decision != "included" {
			continue
		}
		res := m.Employees[er]
		er++
		termsID := ""
		if len(mb.Inputs.TermsIDs) > 0 {
			termsID = mb.Inputs.TermsIDs[0]
		}
		refs := map[string]any{
			"timesheet_line_ids": []string{},
			"terms_id":           termsID,
			"rule_pack_versions": packVers,
			"calculation":        map[string]any{"emp_digest": res.Digest, "calc": CalcVersion},
		}
		mk := func(comp, code string, cents int64, cpf string) {
			if cents == 0 {
				return
			}
			lines = append(lines, PersistLine{
				EmployeeNo: res.ID, Component: comp, PayCode: code,
				AmountCents: cents, CPFTreatment: cpf,
				WorkPeriod: "2026-10", PayableDate: payDate,
				ContributionPeriod: "2026-10", SourceRefs: refs,
			})
		}
		mk("basic", "REG", res.MonthlyBaseCents, "OW")
		mk("overtime", "OT15", res.TotalOTCents, "OW")
		mk("leave", "LEAVE", -res.TotalDedCents, "NON_CPF_WAGE")
		mk("allowance", "ALLOW", res.AllowCents, "OW")
		mk("deduction", "DEDUCT", -res.DeductCents, "NON_CPF_WAGE")
		slips = append(slips, PersistSlip{EmployeeNo: res.ID, Payload: map[string]any{
			"net_cents": res.NetCents, "base_cents": res.MonthlyBaseCents,
			"ot_cents": res.TotalOTCents, "leave_ded_cents": res.TotalDedCents,
			"allow_cents": res.AllowCents, "deduct_cents": res.DeductCents,
			"emp_digest": res.Digest, "calc_version": CalcVersion,
		}})
	}
	return lines, slips
}

// CalcMonth runs the whole snapshot: every member through CalcMember,
// then the shared payroll digest. allowCents/deductCents come from
// the member's captured as-given sums (never re-derived).
func CalcMonth(members []SnapMember, packs []synth.RulePack, period, employer string) (workload.Month, string, error) {
	m := workload.Month{
		CanonicalVersion: "poc-canonical-v1", Period: period, Employer: employer,
		Packs: packs,
	}
	for _, mb := range members {
		er, blocked, exc, err := CalcMember(mb, packs, mb.Inputs.AllowCents, mb.Inputs.DeductCents)
		if err != nil {
			return workload.Month{}, "", err
		}
		m.Exceptions = append(m.Exceptions, exc...)
		if blocked != nil {
			m.Blocked = append(m.Blocked, *blocked)
			continue
		}
		m.Employees = append(m.Employees, er)
	}
	digest, err := workload.MonthPayrollDigest(m)
	if err != nil {
		return workload.Month{}, "", err
	}
	return m, digest, nil
}
