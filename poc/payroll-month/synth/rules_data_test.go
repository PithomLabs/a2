package synth

import (
	"testing"
	"time"
)

// TestDataOnlyPackVariant proves pack VERSIONS are data: v3 changes the
// threshold (8h -> 6h) and rate (3/2 -> 2/1) with zero synth.go changes —
// only a new RulePack value. Honest boundary (documented, not code): a new
// FORMULA (e.g. weekend multiplier, tiered rates) still needs a Go change;
// packs carry parameters, not expressions.
func TestDataOnlyPackVariant(t *testing.T) {
	v3 := RulePack{
		Version: "v3-threshold-6h-double", EffectiveFrom: "2026-10-01", EffectiveTo: "2026-10-31",
		Synthetic: true, OvertimeThresholdMin: 360,
		OvertimeNumer: 2, OvertimeDenom: 1, LeaveDeductionDenom: 30,
	}
	res, err := ComputePay([]DayInput{{"2026-10-05", 480, false, 300000}}, []RulePack{v3})
	if err != nil {
		t.Fatal(err)
	}
	got := res.Days[0]
	// 480-360 = 120 OT min; 120*300000*2/14400 = 5000.
	if got.OvertimeMin != 120 || got.OvertimeCents != 5000 {
		t.Fatalf("v3 day: %+v", got)
	}
	if got.PackVersion != "v3-threshold-6h-double" {
		t.Fatalf("pack = %q", got.PackVersion)
	}
}

// TestSynthPerf is the no-DB half of the perf-sanity experiment: 100 full
// demo-months must complete quickly with a stable digest. Ceiling is generous
// (order-of-magnitude tripwire, not a benchmark); per-month cost is logged.
func TestSynthPerf(t *testing.T) {
	if testing.Short() {
		t.Skip("perf tripwire skipped in -short")
	}
	pack := testPack("v1", "2026-10-01", "2026-10-31")
	var inputs []DayInput
	for d := 1; d <= 31; d++ {
		inputs = append(inputs, DayInput{Day: "2026-10-" + two(d), WorkedMin: 510, MonthlyBaseCents: 300000})
	}
	start := time.Now()
	var digest string
	for i := 0; i < 100; i++ {
		res, err := ComputePay(inputs, []RulePack{pack})
		if err != nil {
			t.Fatal(err)
		}
		if i == 0 {
			digest = res.Digest
		} else if res.Digest != digest {
			t.Fatal("digest unstable across iterations")
		}
	}
	elapsed := time.Since(start)
	t.Logf("100 months x 31 days: %s (%.2f ms/month)", elapsed, float64(elapsed.Microseconds())/100/1000)
	if elapsed > 60*time.Second {
		t.Fatalf("perf tripwire: %s > 60s", elapsed)
	}
}

func two(d int) string {
	if d < 10 {
		return "0" + string(rune('0'+d))
	}
	return string(rune('0'+d/10)) + string(rune('0'+d%10))
}
