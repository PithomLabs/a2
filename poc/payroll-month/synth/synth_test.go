package synth

import (
	"errors"
	"strings"
	"testing"
)

func testPack(version, from, to string) RulePack {
	return RulePack{
		Version:              version,
		EffectiveFrom:        from,
		EffectiveTo:          to,
		Synthetic:            true,
		OvertimeThresholdMin: 480,
		OvertimeNumer:        3,
		OvertimeDenom:        2,
		LeaveDeductionDenom:  30,
	}
}

func TestRuleIDsAreSynthetic(t *testing.T) {
	for _, id := range []RuleID{SynthOT001, SynthLeave001} {
		if !strings.HasPrefix(string(id), "SYNTH-") {
			t.Fatalf("rule %q missing SYNTH- prefix", id)
		}
	}
}

func TestSyntheticCannotActivateAsStatutory(t *testing.T) {
	if err := ActivateStatutory(testPack("v1", "2026-10-01", "2026-10-31")); !errors.Is(err, ErrSyntheticActivation) {
		t.Fatalf("expected ErrSyntheticActivation, got %v", err)
	}
}

func TestVectors(t *testing.T) {
	pack := testPack("v1", "2026-10-01", "2026-10-31")
	const base = int64(300000) // $3000.00
	cases := []struct {
		name string
		in   DayInput
		reg  int
		ot   int
		otc  int64
		ded  int64
	}{
		// Exactly 8h: no overtime (boundary).
		{"exact-8h", DayInput{"2026-10-05", 480, false, base}, 480, 0, 0, 0},
		// 8h30m: 30 OT min; 30*300000*3/28800 = 937.5 -> 937.
		{"ot-30m", DayInput{"2026-10-06", 510, false, base}, 480, 30, 937, 0},
		// Short day: all regular.
		{"short-day", DayInput{"2026-10-07", 240, false, base}, 240, 0, 0, 0},
		// Unpaid leave: 300000/30 = 10000 deduction, no minutes.
		{"leave", DayInput{"2026-10-08", 0, true, base}, 0, 0, 0, 10000},
		// Zero day.
		{"zero", DayInput{"2026-10-09", 0, false, base}, 0, 0, 0, 0},
	}
	for _, c := range cases {
		res, err := ComputePay([]DayInput{c.in}, []RulePack{pack})
		if err != nil {
			t.Fatalf("%s: %v", c.name, err)
		}
		got := res.Days[0]
		if got.RegularMin != c.reg || got.OvertimeMin != c.ot ||
			got.OvertimeCents != c.otc || got.LeaveDeductCents != c.ded {
			t.Fatalf("%s: got %+v", c.name, got)
		}
		if got.PackVersion != "v1" {
			t.Fatalf("%s: pack = %q", c.name, got.PackVersion)
		}
	}
}

func TestMidMonthPackChange(t *testing.T) {
	v1 := testPack("v1", "2026-10-01", "2026-10-15")
	v2 := testPack("v2", "2026-10-16", "2026-10-31")
	v2.OvertimeNumer, v2.OvertimeDenom = 2, 1 // v2 doubles OT
	const base = int64(300000)
	res, err := ComputePay([]DayInput{
		{"2026-10-15", 510, false, base}, // v1: 28125
		{"2026-10-16", 510, false, base}, // v2: 30*300000*2/960 = 18750... see below
	}, []RulePack{v1, v2})
	if err != nil {
		t.Fatal(err)
	}
	// v2: 30*300000*2/(1*30*8*60) = 18000000/14400 = 1250.
	if res.Days[0].PackVersion != "v1" || res.Days[0].OvertimeCents != 937 {
		t.Fatalf("day15: %+v", res.Days[0])
	}
	if res.Days[1].PackVersion != "v2" || res.Days[1].OvertimeCents != 1250 {
		t.Fatalf("day16: %+v", res.Days[1])
	}
}

func TestCoverageGapBlocks(t *testing.T) {
	v1 := testPack("v1", "2026-10-01", "2026-10-15")
	_, err := ComputePay([]DayInput{{"2026-10-16", 480, false, 300000}}, []RulePack{v1})
	if !errors.Is(err, ErrNoCoveringPack) {
		t.Fatalf("expected ErrNoCoveringPack, got %v", err)
	}
	overlap := testPack("v1b", "2026-10-10", "2026-10-20")
	_, err = ComputePay([]DayInput{{"2026-10-12", 480, false, 300000}}, []RulePack{v1, overlap})
	if !errors.Is(err, ErrAmbiguousPack) {
		t.Fatalf("expected ErrAmbiguousPack, got %v", err)
	}
}

func TestDeterminismAndDigest(t *testing.T) {
	pack := testPack("v1", "2026-10-01", "2026-10-31")
	inputs := []DayInput{
		{"2026-10-06", 510, false, 300000},
		{"2026-10-05", 480, false, 300000}, // out of order on purpose
		{"2026-10-08", 0, true, 300000},
	}
	a, err := ComputePay(inputs, []RulePack{pack})
	if err != nil {
		t.Fatal(err)
	}
	b, err := ComputePay(inputs, []RulePack{pack})
	if err != nil {
		t.Fatal(err)
	}
	if a.Digest != b.Digest || len(a.Digest) != 64 {
		t.Fatalf("digest unstable: %q vs %q", a.Digest, b.Digest)
	}
	// Sorted output regardless of input order.
	if a.Days[0].Day != "2026-10-05" || a.Days[2].Day != "2026-10-08" {
		t.Fatalf("days not sorted: %+v", a.Days)
	}
	if want := int64(937); a.TotalOTCents != want {
		t.Fatalf("total OT = %d, want %d", a.TotalOTCents, want)
	}
	if want := int64(10000); a.TotalDedCents != want {
		t.Fatalf("total ded = %d, want %d", a.TotalDedCents, want)
	}
}

func TestBadInputRejected(t *testing.T) {
	pack := testPack("v1", "2026-10-01", "2026-10-31")
	_, err := ComputePay([]DayInput{{"2026-10-05", -1, false, 300000}}, []RulePack{pack})
	if !errors.Is(err, ErrBadInput) {
		t.Fatalf("expected ErrBadInput, got %v", err)
	}
}
