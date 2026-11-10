package workload

// Outputs are read models over the finalized Month: projections and sums,
// never recalculation. Register/payslips/payment-preview reproduce the
// finalized amounts; snapshot is the canonical artifact itself.

import "encoding/json"

// RegisterRow is one payroll-register line.
type RegisterRow struct {
	ID               string `json:"id"`
	Status           string `json:"status"` // "final" | "blocked"
	MonthlyBaseCents int64  `json:"monthly_base_cents"`
	OTCents          int64  `json:"ot_cents"`
	LeaveDedCents    int64  `json:"leave_ded_cents"`
	AllowCents       int64  `json:"allow_cents"`
	DeductCents      int64  `json:"deduct_cents"`
	NetCents         int64  `json:"net_cents"`
}

// Register is output 1: per-employee totals plus grand totals. Blocked
// employees appear with status "blocked" and zero amounts: excluded from
// every total, never silently dropped.
type Register struct {
	Period string        `json:"period"`
	Rows   []RegisterRow `json:"rows"`
	Total  RegisterRow   `json:"total"` // ID "TOTAL", sums over final rows only
}

// Payslip is output 2: one employee's day rows plus totals.
type Payslip struct {
	ID               string       `json:"id"`
	MonthlyBaseCents int64        `json:"monthly_base_cents"`
	Days             []PayslipDay `json:"days"`
	TotalOTCents     int64        `json:"total_ot_cents"`
	TotalDedCents    int64        `json:"total_ded_cents"`
	AllowCents       int64        `json:"allow_cents"`
	DeductCents      int64        `json:"deduct_cents"`
	NetCents         int64        `json:"net_cents"`
	Synthetic        bool         `json:"synthetic"`
}

// PayslipDay is one payslip line.
type PayslipDay struct {
	Day              string `json:"day"`
	PackVersion      string `json:"pack_version"`
	RegularMin       int    `json:"regular_min"`
	OvertimeMin      int    `json:"overtime_min"`
	OvertimeCents    int64  `json:"overtime_cents"`
	LeaveDeductCents int64  `json:"leave_deduct_cents"`
}

// PaymentRow is one payment-file preview line. No bank details exist in the
// synthetic workload; the preview carries payee id + net only.
type PaymentRow struct {
	ID       string `json:"id"`
	Status   string `json:"status"` // "payable" | "blocked"
	NetCents int64  `json:"net_cents"`
	Currency string `json:"currency"` // "SGD-synthetic"
}

// PaymentPreview is output 4.
type PaymentPreview struct {
	Period string       `json:"period"`
	Rows   []PaymentRow `json:"rows"`
	Total  int64        `json:"total_cents"` // over payable rows only
}

// Outputs is the six-output bundle. Snapshot is the canonical Month;
// DigestResult repeats the digest with its canonicalization version.
type Outputs struct {
	Register   Register       `json:"register"`
	Payslips   []Payslip      `json:"payslips"`
	Exceptions []Exception    `json:"exceptions"` // output 3
	Payments   PaymentPreview `json:"payment_preview"`
	Snapshot   Month          `json:"snapshot"` // output 5
	Digest     DigestResult   `json:"digest"`   // output 6
}

// DigestResult names the canonicalization alongside the digest.
type DigestResult struct {
	CanonicalVersion string `json:"canonical_version"`
	SHA256Hex        string `json:"sha256_hex"`
}

// BuildOutputs projects the six outputs from a finalized Month.
func BuildOutputs(m Month) Outputs {
	reg := Register{Period: m.Period, Total: RegisterRow{ID: "TOTAL", Status: "final"}}
	pay := PaymentPreview{Period: m.Period}
	var slips []Payslip
	for _, e := range m.Employees {
		reg.Rows = append(reg.Rows, RegisterRow{
			ID: e.ID, Status: "final", MonthlyBaseCents: e.MonthlyBaseCents,
			OTCents: e.TotalOTCents, LeaveDedCents: e.TotalDedCents,
			AllowCents: e.AllowCents, DeductCents: e.DeductCents, NetCents: e.NetCents,
		})
		reg.Total.MonthlyBaseCents += e.MonthlyBaseCents
		reg.Total.OTCents += e.TotalOTCents
		reg.Total.LeaveDedCents += e.TotalDedCents
		reg.Total.AllowCents += e.AllowCents
		reg.Total.DeductCents += e.DeductCents
		reg.Total.NetCents += e.NetCents
		var days []PayslipDay
		for _, d := range e.Days {
			days = append(days, PayslipDay{
				Day: d.Day, PackVersion: d.PackVersion, RegularMin: d.RegularMin,
				OvertimeMin: d.OvertimeMin, OvertimeCents: d.OvertimeCents,
				LeaveDeductCents: d.LeaveDeductCents,
			})
		}
		pay.Rows = append(pay.Rows, PaymentRow{ID: e.ID, Status: "payable", NetCents: e.NetCents, Currency: "SGD-synthetic"})
		pay.Total += e.NetCents
		slips = append(slips, Payslip{ID: e.ID, MonthlyBaseCents: e.MonthlyBaseCents, Days: days,
			TotalOTCents: e.TotalOTCents, TotalDedCents: e.TotalDedCents,
			AllowCents: e.AllowCents, DeductCents: e.DeductCents,
			NetCents: e.NetCents, Synthetic: true})
	}
	for _, b := range m.Blocked {
		reg.Rows = append(reg.Rows, RegisterRow{ID: b.ID, Status: "blocked"})
		pay.Rows = append(pay.Rows, PaymentRow{ID: b.ID, Status: "blocked", NetCents: 0, Currency: "SGD-synthetic"})
	}
	return Outputs{
		Register: reg, Payslips: slips, Exceptions: m.Exceptions,
		Payments: pay, Snapshot: m,
		Digest: DigestResult{CanonicalVersion: m.CanonicalVersion, SHA256Hex: m.Digest},
	}
}

// OutputsJSON renders the bundle deterministically.
func OutputsJSON(o Outputs) ([]byte, error) {
	return json.MarshalIndent(o, "", "  ")
}
