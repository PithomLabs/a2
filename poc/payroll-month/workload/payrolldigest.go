package workload

// Shared payroll-equality digest (prompt8 RC-1 repair).
//
// MonthPayrollDigest projects a Month to every financially meaningful
// value — per-employee results, blocked population, day exceptions,
// packs, period, employer — and hashes the canonical JSON. The
// DB-backed workflow (dbosflow/calc.go) builds an equivalent Month
// from persisted snapshot rows and MUST produce the identical digest;
// the pinned Month.Digest (a35cea69… since the EV-002 v2 fixture
// correction; covering messages/demos/events too) stays untouched
// as the DB-free tripwire.
//
// Excluded from the projection, deliberately: human messages, demo
// mini-cases, narrative events (story data, never calc inputs).

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

	"github.com/PithomLabs/a2/poc/payroll-month/synth"
)

type payrollDigestEmp struct {
	ID        string            `json:"id"`
	Base      int64             `json:"base"`
	Days      []synth.DayResult `json:"days"`
	OT        int64             `json:"ot"`
	LeaveDed  int64             `json:"leave_ded"`
	Allow     int64             `json:"allow"`
	Deduct    int64             `json:"deduct"`
	Net       int64             `json:"net"`
	EmpDigest string            `json:"emp_digest"`
}

type payrollDigestBlocked struct {
	ID   string `json:"id"`
	Code string `json:"code"`
}

type payrollDigestDayExc struct {
	Ref  string `json:"ref"`
	Code string `json:"code"`
}

type payrollDigestDoc struct {
	Version   string                 `json:"v"`
	Period    string                 `json:"period"`
	Employer  string                 `json:"employer"`
	Packs     []synth.RulePack       `json:"packs"`
	Employees []payrollDigestEmp     `json:"employees"`
	Blocked   []payrollDigestBlocked `json:"blocked"`
	DayExc    []payrollDigestDayExc  `json:"day_exc"`
}

// MonthPayrollDigest returns the hex sha256 of the payroll projection.
func MonthPayrollDigest(m Month) (string, error) {
	doc := payrollDigestDoc{
		Version: "poc-payroll-digest-v1", Period: m.Period,
		Employer: m.Employer, Packs: m.Packs,
	}
	for _, e := range m.Employees {
		doc.Employees = append(doc.Employees, payrollDigestEmp{
			ID: e.ID, Base: e.MonthlyBaseCents, Days: e.Days,
			OT: e.TotalOTCents, LeaveDed: e.TotalDedCents,
			Allow: e.AllowCents, Deduct: e.DeductCents,
			Net: e.NetCents, EmpDigest: e.Digest,
		})
	}
	for _, b := range m.Blocked {
		doc.Blocked = append(doc.Blocked, payrollDigestBlocked{ID: b.ID, Code: b.Block.Code})
	}
	for _, x := range m.Exceptions {
		if x.Scope != "day" {
			continue
		}
		doc.DayExc = append(doc.DayExc, payrollDigestDayExc{Ref: x.Ref, Code: x.Code})
	}
	raw, err := json.Marshal(doc)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:]), nil
}
