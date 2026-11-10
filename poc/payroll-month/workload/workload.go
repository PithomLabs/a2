// Package workload builds the Phase 2 shared payroll workload: 25 synthetic
// employees x 31 days (October 2026) plus explicitly classified scenarios.
//
// Classification FIRST (review clarification A): every workload element is
// CALCULABLE (fully specified synthetic semantics, runs through
// synth.ComputePay UNCHANGED — this package never reimplements the
// calculator) or MUST-BLOCK (excluded from calculation, listed in the
// exception report with a machine-readable code). Nothing here is statutory:
// every rule carries synthetic:true and SG-MOM-OT-001 stays BLOCKED.
//
// As-given line items (allowances/deductions) are fixture DATA, not rules:
// amounts are declared and reproduced verbatim in outputs, labeled
// source:"fixture-as-given". They exercise capture/rendering of varied
// inputs without inventing calculation semantics.
package workload

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/PithomLabs/a2/poc/payroll-month/synth"
)

// CanonicalVersion pins the digest canonicalization. Any change to the
// canonical bytes layout MUST bump this string.
const CanonicalVersion = "poc-canonical-v1"

// Class is the workload classification of one element.
type Class string

const (
	// Calculable elements run through synth.ComputePay.
	Calculable Class = "calculable"
	// Blocked elements are excluded from calc and reported as exceptions.
	Blocked Class = "blocked"
)

// Block codes. B-BASE-CHANGE and B-LEAVE-WORKED are workload-policy blocks
// (enforced by this package before ComputePay); B-PACK-* and B-BAD-INPUT
// are synth-level errors demonstrated by mini-cases.
const (
	CodeBaseChange  = "B-BASE-CHANGE"  // intra-month base change: proration unresolved
	CodeLeaveWorked = "B-LEAVE-WORKED" // leave day carrying worked minutes
	CodePackGap     = "B-PACK-GAP"     // no pack covers day (synth.ErrNoCoveringPack)
	CodePackOverlap = "B-PACK-OVERLAP" // two packs cover day (synth.ErrAmbiguousPack)
	CodeBadInput    = "B-BAD-INPUT"    // negative minutes (synth.ErrBadInput)
)

// Block is one exclusion with its machine-readable reason.
type Block struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// EmployeeDay is one employee-day with its classification.
type EmployeeDay struct {
	Day           string `json:"day"` // YYYY-MM-DD
	WorkedMin     int    `json:"worked_min"`
	OnUnpaidLeave bool   `json:"on_unpaid_leave"`
	Class         Class  `json:"class"`
	Block         *Block `json:"block,omitempty"`
}

// AsGiven is a fixture-declared line item: reproduced verbatim, never derived.
type AsGiven struct {
	Kind   string `json:"kind"` // "allowance" | "deduction"
	Label  string `json:"label"`
	Cents  int64  `json:"cents"`
	Source string `json:"source"` // always "fixture-as-given"
}

// Employee is one synthetic employee's month. A non-nil Blocked excludes the
// WHOLE employee: monthly base is month-scoped, so an intra-month change
// blocks the month (proration is unresolved — never invented).
type Employee struct {
	ID               string        `json:"id"` // SYN-001 .. SYN-025, no PII anywhere
	MonthlyBaseCents int64         `json:"monthly_base_cents"`
	Days             []EmployeeDay `json:"days"`
	AsGiven          []AsGiven     `json:"as_given"`
	Blocked          *Block        `json:"blocked,omitempty"`
}

// Event is a fixture-narrative process event (late approval / correction).
// Timestamps are story data for the DB-backed capture path, never calc inputs.
type Event struct {
	ID     string `json:"id"`
	At     string `json:"at"` // RFC3339 fixture narrative
	Kind   string `json:"kind"`
	Ref    string `json:"ref"`
	Detail string `json:"detail"`
}

// DemoBlock is a mini-case proving a synth-level error surfaces loudly.
type DemoBlock struct {
	ID      string           `json:"id"`
	Code    string           `json:"code"`
	Inputs  []synth.DayInput `json:"inputs"`
	Packs   []synth.RulePack `json:"packs"`
	WantErr error            `json:"-"`
}

// Workload is the full Phase 2 shared input, identical for all experiments.
type Workload struct {
	Period    string           `json:"period"` // 2026-10
	Employer  string           `json:"employer"`
	Packs     []synth.RulePack `json:"packs"`
	Employees []Employee       `json:"employees"`
	Demos     []DemoBlock      `json:"demo_blocks"`
	Events    []Event          `json:"events"`
}

// MonthPacks returns the fully-defined two-pack set: v1 (OT 3/2) covers
// Oct 1-15, v2 (OT 2/1) covers Oct 16-31. The rate change is observable in
// outputs (SYN-025 works OT on both Oct 15 and Oct 16).
func MonthPacks() []synth.RulePack {
	base := synth.RulePack{
		Synthetic: true, OvertimeThresholdMin: 480, LeaveDeductionDenom: 30,
	}
	v1 := base
	v1.Version, v1.EffectiveFrom, v1.EffectiveTo = "SYNTH-2026-10-v1a", "2026-10-01", "2026-10-15"
	v1.OvertimeNumer, v1.OvertimeDenom = 3, 2
	v2 := base
	v2.Version, v2.EffectiveFrom, v2.EffectiveTo = "SYNTH-2026-10-v2a", "2026-10-16", "2026-10-31"
	v2.OvertimeNumer, v2.OvertimeDenom = 2, 1
	return []synth.RulePack{v1, v2}
}

// Build constructs the deterministic shared workload. Fully scripted by
// index (no PRNG): same code always yields byte-identical output.
func Build() Workload {
	packs := MonthPacks()
	w := Workload{Period: "2026-10", Employer: "POC-MERIDIAN", Packs: packs}
	for e := 0; e < 25; e++ {
		id := fmt.Sprintf("SYN-%03d", e+1)
		emp := Employee{ID: id, MonthlyBaseCents: int64(240000 + e*8000)}
		leaveDay := 3 + (e%9)*3 // spread 3..27 step 3
		for d := 1; d <= 31; d++ {
			day := EmployeeDay{Day: fmt.Sprintf("2026-10-%02d", d), Class: Calculable}
			worked := 480
			switch {
			case e >= 15 && e <= 19:
				worked = 240 // short-day cohort
			case e == 20 || e == 21:
				worked = 600 // heavy-OT cohort (120 OT min)
			case e == 22:
				worked = 0 // zero-month edge
			}
			switch {
			case e == 24 && (d == 15 || d == 16):
				worked = 600 // pack-boundary observer
			case e != 22 && (d == leaveDay || e == 23 && (d == 3 || d == 10 || d == 17 || d == 24)):
				// Normal leave day: no worked minutes (leave+minutes is a
				// B-LEAVE-WORKED conflict, never silently accepted).
				day.OnUnpaidLeave, worked = true, 0
			case d%7 == 0:
				worked = 0 // rest pattern
			case d%5 == 0 && worked == 480:
				worked = 540 // 60 OT min
			}
			// Contractual rest day is Sunday (Oct 4/11/18/25): no shift
			// assignment may land on it (wfm validatePlacement). Rest on
			// other weekdays is always fine, so the Wednesday rest
			// pattern stands. Leave flags are untouched: a leave day
			// that falls on a Sunday stays leave with zero minutes.
			if d == 4 || d == 11 || d == 18 || d == 25 {
				worked = 0
			}
			day.WorkedMin = worked
			emp.Days = append(emp.Days, day)
		}
		// As-given items: declared data, same for every month construction.
		emp.AsGiven = []AsGiven{
			{Kind: "allowance", Label: "SYNTH-TRANSPORT", Cents: int64(10000 + e*500), Source: "fixture-as-given"},
		}
		if e%2 == 0 {
			emp.AsGiven = append(emp.AsGiven, AsGiven{Kind: "deduction", Label: "SYNTH-FUND", Cents: 5000, Source: "fixture-as-given"})
		}
		w.Employees = append(w.Employees, emp)
	}
	// Scripted MUST-BLOCK cases (review clarification A: explicit blockers).
	// SYN-006 Oct-10: leave day carrying worked minutes.
	oct10 := &w.Employees[5].Days[9]
	oct10.OnUnpaidLeave, oct10.WorkedMin = true, 480
	oct10.Class = Blocked
	oct10.Block = &Block{Code: CodeLeaveWorked, Message: "leave day carries 480 worked minutes: excluded, needs HR disposition"}
	// EV-002: SYN-009 Oct-12 corrected v1 (480) -> v2 (540) via
	// controlled succession (out punch was 17:00, corrected to 18:00).
	// The fixture MUST encode v2: the seeder persists both versions
	// (v1 superseded, v2 current) and capture takes current. A fixture
	// still scripting v1 diverges from the seeded truth it claims to
	// share (proven live: DB-backed OT exceeded canonical by exactly
	// the 60 corrected minutes).
	w.Employees[8].Days[11].WorkedMin = 540
	// SYN-012: mid-month base change (328000 -> 336000 from Oct-16).
	// Monthly base is month-scoped: the whole employee month blocks.
	w.Employees[11].Blocked = &Block{Code: CodeBaseChange,
		Message: "base changes 328000 -> 336000 on 2026-10-16: proration unresolved, month blocked"}
	w.Demos = []DemoBlock{
		{ID: "demo-gap", Code: CodePackGap,
			Inputs:  []synth.DayInput{{Day: "2026-10-16", WorkedMin: 480, MonthlyBaseCents: 300000}},
			Packs:   []synth.RulePack{packs[0]},
			WantErr: synth.ErrNoCoveringPack},
		{ID: "demo-overlap", Code: CodePackOverlap,
			Inputs:  []synth.DayInput{{Day: "2026-10-12", WorkedMin: 480, MonthlyBaseCents: 300000}},
			Packs:   packsOverlap(),
			WantErr: synth.ErrAmbiguousPack},
		{ID: "demo-neg", Code: CodeBadInput,
			Inputs:  []synth.DayInput{{Day: "2026-10-05", WorkedMin: -1, MonthlyBaseCents: 300000}},
			Packs:   packs,
			WantErr: synth.ErrBadInput},
	}
	w.Events = []Event{
		{ID: "EV-001", At: "2026-10-25T15:04:00+08:00", Kind: "late-approval",
			Ref: "SYN-004/2026-10-05", Detail: "Oct-05 sheet approved Oct-25, after first capture window"},
		{ID: "EV-002", At: "2026-10-20T10:30:00+08:00", Kind: "correction",
			Ref: "SYN-009/2026-10-12", Detail: "Oct-12 interval corrected v1 -> v2 via controlled succession"},
	}
	return w
}

func packsOverlap() []synth.RulePack {
	p := MonthPacks()[0]
	q := p
	q.Version, q.EffectiveFrom, q.EffectiveTo = "SYNTH-2026-10-v1b", "2026-10-10", "2026-10-20"
	return []synth.RulePack{p, q}
}

// EmpResult is one employee's finalized calculation.
type EmpResult struct {
	ID               string            `json:"id"`
	MonthlyBaseCents int64             `json:"monthly_base_cents"`
	Days             []synth.DayResult `json:"days"`
	TotalOTCents     int64             `json:"total_ot_cents"`
	TotalDedCents    int64             `json:"total_ded_cents"`
	AllowCents       int64             `json:"allow_cents"` // as-given sums, reproduced
	DeductCents      int64             `json:"deduct_cents"`
	NetCents         int64             `json:"net_cents"` // base + OT - leave - fund + transport
	Digest           string            `json:"digest"`    // synth digest of this employee's days
}

// BlockedEmp records a whole-employee exclusion.
type BlockedEmp struct {
	ID    string `json:"id"`
	Block Block  `json:"block"`
}

// Exception is one row of the exception report.
type Exception struct {
	Scope   string `json:"scope"` // "day" | "employee" | "demo"
	Ref     string `json:"ref"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Month is the finalized canonical artifact. Canonical bytes are
// json.Marshal of this struct with Digest unset; Digest is the sha256 hex
// of those bytes. encoding/json struct marshaling is field-ordered, hence
// deterministic; employees/exceptions/events are pre-sorted by construction.
type Month struct {
	CanonicalVersion string           `json:"canonical_version"`
	Period           string           `json:"period"`
	Employer         string           `json:"employer"`
	Packs            []synth.RulePack `json:"packs"`
	Employees        []EmpResult      `json:"employees"`
	Blocked          []BlockedEmp     `json:"blocked"`
	Exceptions       []Exception      `json:"exceptions"`
	Events           []Event          `json:"events"`
	Digest           string           `json:"digest"`
}

// Canonical runs the canonical month: every calculable input through
// synth.ComputePay, every demo block proving its error, then digest.
// A ComputePay error on workload-built inputs is a WORKLOAD bug (the
// builder must only emit calculable days) and fails loudly.
func Canonical(w Workload) (Month, error) {
	m := Month{
		CanonicalVersion: CanonicalVersion,
		Period:           w.Period,
		Employer:         w.Employer,
		Packs:            w.Packs,
		Events:           w.Events,
	}
	for _, e := range w.Employees {
		if e.Blocked != nil {
			m.Blocked = append(m.Blocked, BlockedEmp{ID: e.ID, Block: *e.Blocked})
			m.Exceptions = append(m.Exceptions, Exception{
				Scope: "employee", Ref: e.ID,
				Code: e.Blocked.Code, Message: e.Blocked.Message,
			})
			continue
		}
		var inputs []synth.DayInput
		for _, d := range e.Days {
			if d.Class == Blocked {
				m.Exceptions = append(m.Exceptions, Exception{
					Scope: "day", Ref: e.ID + "/" + d.Day,
					Code: d.Block.Code, Message: d.Block.Message,
				})
				continue
			}
			// Trust-but-verify: the builder owns classification, Canonical
			// refuses to compute a day that violates workload policy.
			if d.OnUnpaidLeave && d.WorkedMin > 0 {
				return Month{}, fmt.Errorf("workload bug: employee %s day %s marked calculable with leave+minutes", e.ID, d.Day)
			}
			inputs = append(inputs, synth.DayInput{
				Day: d.Day, WorkedMin: d.WorkedMin,
				OnUnpaidLeave: d.OnUnpaidLeave, MonthlyBaseCents: e.MonthlyBaseCents,
			})
		}
		res, err := synth.ComputePay(inputs, w.Packs)
		if err != nil {
			return Month{}, fmt.Errorf("workload bug: employee %s: %w", e.ID, err)
		}
		var allow, deduct int64
		for _, a := range e.AsGiven {
			switch a.Kind {
			case "allowance":
				allow += a.Cents
			case "deduction":
				deduct += a.Cents
			}
		}
		m.Employees = append(m.Employees, EmpResult{
			ID: e.ID, MonthlyBaseCents: e.MonthlyBaseCents,
			Days: res.Days, TotalOTCents: res.TotalOTCents,
			TotalDedCents: res.TotalDedCents,
			AllowCents:    allow, DeductCents: deduct,
			NetCents: e.MonthlyBaseCents + res.TotalOTCents - res.TotalDedCents + allow - deduct,
			Digest:   res.Digest,
		})
	}
	for _, d := range w.Demos {
		_, err := synth.ComputePay(d.Inputs, d.Packs)
		if !errors.Is(err, d.WantErr) {
			return Month{}, fmt.Errorf("demo %s: want %v, got %v", d.ID, d.WantErr, err)
		}
		m.Exceptions = append(m.Exceptions, Exception{
			Scope: "demo", Ref: d.ID, Code: d.Code, Message: err.Error(),
		})
	}
	raw, err := json.Marshal(m)
	if err != nil {
		return Month{}, err
	}
	sum := sha256.Sum256(raw)
	m.Digest = hex.EncodeToString(sum[:])
	return m, nil
}
