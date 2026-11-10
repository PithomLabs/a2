package workload

import (
	"reflect"
	"testing"

	"github.com/PithomLabs/a2/poc/payroll-month/synth"
)

func TestBuildShape(t *testing.T) {
	w := Build()
	if len(w.Employees) != 25 {
		t.Fatalf("employees = %d, want 25", len(w.Employees))
	}
	for i, e := range w.Employees {
		if len(e.Days) != 31 {
			t.Fatalf("%s: days = %d, want 31", e.ID, len(e.Days))
		}
		if want := int64(240000 + i*8000); e.MonthlyBaseCents != want {
			t.Fatalf("%s: base = %d, want %d", e.ID, e.MonthlyBaseCents, want)
		}
	}
	if len(w.Packs) != 2 || len(w.Demos) != 3 || len(w.Events) != 2 {
		t.Fatalf("packs/demos/events = %d/%d/%d, want 2/3/2",
			len(w.Packs), len(w.Demos), len(w.Events))
	}
}

// TestSundayRest is the regression test for the Gate B seeder failure:
// wfm validatePlacement rejects shift assignments on the contractual rest
// day (Sunday), so the workload must schedule zero minutes on Sundays.
func TestSundayRest(t *testing.T) {
	sundays := map[string]bool{
		"2026-10-04": true, "2026-10-11": true, "2026-10-18": true, "2026-10-25": true,
	}
	for _, e := range Build().Employees {
		for _, d := range e.Days {
			if sundays[d.Day] && d.WorkedMin > 0 {
				t.Fatalf("%s/%s: %d worked minutes on a Sunday", e.ID, d.Day, d.WorkedMin)
			}
		}
	}
}

func TestBuildDeterministic(t *testing.T) {
	// Scripted by index (no PRNG): byte-identical across constructions.
	if !reflect.DeepEqual(Build(), Build()) {
		t.Fatal("Build() not deterministic")
	}
}

func TestClassificationComplete(t *testing.T) {
	w := Build()
	var blockedDays int
	for _, e := range w.Employees {
		for _, d := range e.Days {
			switch d.Class {
			case Calculable:
				if d.Block != nil {
					t.Fatalf("%s/%s: calculable with block set", e.ID, d.Day)
				}
				if d.OnUnpaidLeave && d.WorkedMin > 0 {
					t.Fatalf("%s/%s: calculable leave+minutes", e.ID, d.Day)
				}
			case Blocked:
				if d.Block == nil || d.Block.Code == "" {
					t.Fatalf("%s/%s: blocked without code", e.ID, d.Day)
				}
				blockedDays++
			default:
				t.Fatalf("%s/%s: missing class", e.ID, d.Day)
			}
		}
	}
	if blockedDays != 1 {
		t.Fatalf("blocked days = %d, want 1 (SYN-006/2026-10-10)", blockedDays)
	}
	var blockedEmps int
	for _, e := range w.Employees {
		if e.Blocked != nil {
			blockedEmps++
			if e.ID != "SYN-012" || e.Blocked.Code != CodeBaseChange {
				t.Fatalf("unexpected employee block: %s %+v", e.ID, e.Blocked)
			}
		}
	}
	if blockedEmps != 1 {
		t.Fatalf("blocked employees = %d, want 1", blockedEmps)
	}
}

// TestGoldenHandComputed pins literal amounts derived by hand (not from the
// implementation): base 300000, OT 30 min at 3/2 over 30*8*60 monthly
// minutes = 30*300000*3/28800 = 937.5 -> 937; leave = 300000/30 = 10000.
func TestGoldenHandComputed(t *testing.T) {
	p := MonthPacks()[0]
	res, err := synth.ComputePay([]synth.DayInput{
		{Day: "2026-10-01", WorkedMin: 480, MonthlyBaseCents: 300000},
		{Day: "2026-10-02", WorkedMin: 510, MonthlyBaseCents: 300000},
		{Day: "2026-10-03", OnUnpaidLeave: true, MonthlyBaseCents: 300000},
	}, []synth.RulePack{p})
	if err != nil {
		t.Fatal(err)
	}
	if res.TotalOTCents != 937 || res.TotalDedCents != 10000 {
		t.Fatalf("totals OT/ded = %d/%d, want 937/10000", res.TotalOTCents, res.TotalDedCents)
	}
}

// TestPackBoundaryObserved proves the mid-month pack change is effective:
// SYN-025 base 432000, 120 OT min on Oct-15 (v1, 3/2) and Oct-16 (v2, 2/1).
// v1: 120*432000*3/28800 = 5400. v2: 120*432000*2/14400 = 7200.
func TestPackBoundaryObserved(t *testing.T) {
	m, err := Canonical(Build())
	if err != nil {
		t.Fatal(err)
	}
	var e25 *EmpResult
	for i := range m.Employees {
		if m.Employees[i].ID == "SYN-025" {
			e25 = &m.Employees[i]
		}
	}
	if e25 == nil {
		t.Fatal("SYN-025 missing from canonical month")
	}
	byDay := map[string]int64{}
	ver := map[string]string{}
	for _, d := range e25.Days {
		byDay[d.Day] = d.OvertimeCents
		ver[d.Day] = d.PackVersion
	}
	if byDay["2026-10-15"] != 5400 || ver["2026-10-15"] != "SYNTH-2026-10-v1a" {
		t.Fatalf("Oct-15: %d %q, want 5400 v1a", byDay["2026-10-15"], ver["2026-10-15"])
	}
	if byDay["2026-10-16"] != 7200 || ver["2026-10-16"] != "SYNTH-2026-10-v2a" {
		t.Fatalf("Oct-16: %d %q, want 7200 v2a", byDay["2026-10-16"], ver["2026-10-16"])
	}
}

func TestCanonicalExceptions(t *testing.T) {
	m, err := Canonical(Build())
	if err != nil {
		t.Fatal(err)
	}
	if len(m.Employees) != 24 || len(m.Blocked) != 1 {
		t.Fatalf("employees/blocked = %d/%d, want 24/1", len(m.Employees), len(m.Blocked))
	}
	// 1 employee block + 1 day block + 3 demo blocks.
	if len(m.Exceptions) != 5 {
		t.Fatalf("exceptions = %d, want 5", len(m.Exceptions))
	}
	seen := map[string]bool{}
	for _, x := range m.Exceptions {
		seen[x.Code] = true
	}
	for _, c := range []string{CodeBaseChange, CodeLeaveWorked, CodePackGap, CodePackOverlap, CodeBadInput} {
		if !seen[c] {
			t.Fatalf("code %s missing from exceptions", c)
		}
	}
}

func TestTotalsCrossCheck(t *testing.T) {
	m, err := Canonical(Build())
	if err != nil {
		t.Fatal(err)
	}
	o := BuildOutputs(m)
	var ot, ded, net int64
	for _, s := range o.Payslips {
		ot += s.TotalOTCents
		ded += s.TotalDedCents
		net += s.NetCents
		// payslip day rows sum to payslip totals
		var dot, dded int64
		for _, d := range s.Days {
			dot += d.OvertimeCents
			dded += d.LeaveDeductCents
		}
		if dot != s.TotalOTCents || dded != s.TotalDedCents {
			t.Fatalf("%s: day rows do not sum to totals", s.ID)
		}
	}
	if o.Register.Total.OTCents != ot || o.Register.Total.LeaveDedCents != ded ||
		o.Register.Total.NetCents != net || o.Payments.Total != net {
		t.Fatalf("register/payment totals disagree with payslip sums")
	}
	if len(o.Register.Rows) != 25 || len(o.Payments.Rows) != 25 || len(o.Payslips) != 24 {
		t.Fatalf("rows register/payment/payslips = %d/%d/%d, want 25/25/24",
			len(o.Register.Rows), len(o.Payments.Rows), len(o.Payslips))
	}
	// Blocked employee present with zero amounts, excluded from totals.
	for _, r := range o.Register.Rows {
		if r.ID == "SYN-012" && (r.Status != "blocked" || r.NetCents != 0) {
			t.Fatalf("SYN-012 register row: %+v", r)
		}
	}
}

func TestDigestStability(t *testing.T) {
	// Stability tripwire only: identical inputs MUST yield identical
	// digests across runs. This asserts determinism, NOT independent
	// verification of the amounts (see TestGoldenHandComputed for that).
	a, err := Canonical(Build())
	if err != nil {
		t.Fatal(err)
	}
	b, err := Canonical(Build())
	if err != nil {
		t.Fatal(err)
	}
	if a.Digest != b.Digest || len(a.Digest) != 64 {
		t.Fatalf("digest unstable: %q vs %q", a.Digest, b.Digest)
	}
}
