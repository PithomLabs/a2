// Package synth implements clearly-labeled SYNTHETIC, non-statutory payroll
// rules for the one-month POC. Nothing in this package is a statutory
// calculation: every rule carries synthetic:true, amounts are exact integer
// cents (no binary floating point), and time is whole minutes. The POC proves
// architecture (capture, orchestration, evidence), not payroll correctness.
package synth

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
)

// Rule identifiers. The SYNTH- prefix is mandatory and asserted in tests.
type RuleID string

const (
	// SynthOT001 pays overtime at 3/2 (1.5x) for minutes worked above the
	// daily threshold. A regular minute is worth base/(30*8*60) cents; an
	// overtime minute is numer/denom times that. Overtime for a day is
	// otMin*base*numer/(denom*30*8*60) cents with ONE truncation at the end.
	SynthOT001 RuleID = "SYNTH-OT-001"
	// SynthLeave001 deducts base/30 cents (truncated) for each unpaid-leave day.
	SynthLeave001 RuleID = "SYNTH-LEAVE-001"
)

// RulePack is a versioned set of synthetic rules with an effective span.
// Days are YYYY-MM-DD; lexicographic order equals chronological order.
type RulePack struct {
	Version              string `json:"version"`
	EffectiveFrom        string `json:"effective_from"`         // inclusive
	EffectiveTo          string `json:"effective_to"`           // inclusive
	Synthetic            bool   `json:"synthetic"`              // must be true for POC packs
	OvertimeThresholdMin int    `json:"overtime_threshold_min"` // 480
	OvertimeNumer        int    `json:"overtime_numer"`         // 3
	OvertimeDenom        int    `json:"overtime_denom"`         // 2
	LeaveDeductionDenom  int    `json:"leave_deduction_denom"`  // 30
}

// Covers reports whether day falls inside the pack's effective span.
func (p RulePack) Covers(day string) bool {
	return p.EffectiveFrom <= day && day <= p.EffectiveTo
}

// DayInput is one employee-day of synthetic payroll input.
type DayInput struct {
	Day              string `json:"day"` // YYYY-MM-DD
	WorkedMin        int    `json:"worked_min"`
	OnUnpaidLeave    bool   `json:"on_unpaid_leave"`
	MonthlyBaseCents int64  `json:"monthly_base_cents"`
}

// DayResult is the deterministic synthetic outcome for one day.
type DayResult struct {
	Day              string `json:"day"`
	PackVersion      string `json:"pack_version"`
	RegularMin       int    `json:"regular_min"`
	OvertimeMin      int    `json:"overtime_min"`
	OvertimeCents    int64  `json:"overtime_cents"`
	LeaveDeductCents int64  `json:"leave_deduct_cents"`
}

// PayResult is the whole-run synthetic outcome with its content digest.
type PayResult struct {
	Days          []DayResult `json:"days"`
	TotalOTCents  int64       `json:"total_ot_cents"`
	TotalDedCents int64       `json:"total_ded_cents"`
	Digest        string      `json:"digest"` // sha256 of canonical days JSON
}

var (
	// ErrNoCoveringPack is returned when no pack covers a day: missing
	// coverage blocks loudly, it never defaults or infers.
	ErrNoCoveringPack = errors.New("synth: no rule pack covers day")
	// ErrAmbiguousPack is returned when two packs cover one day.
	ErrAmbiguousPack = errors.New("synth: multiple rule packs cover day")
	// ErrSyntheticActivation is returned when a synthetic pack is offered
	// for statutory activation.
	ErrSyntheticActivation = errors.New("synth: synthetic pack cannot activate as statutory")
	// ErrBadInput is returned for negative or inconsistent inputs.
	ErrBadInput = errors.New("synth: bad input")
)

// ActivateStatutory is the guard proving synthetic rules can never become
// statutory: it always rejects packs with Synthetic=true.
func ActivateStatutory(p RulePack) error {
	if p.Synthetic {
		return fmt.Errorf("%w: %s", ErrSyntheticActivation, p.Version)
	}
	return errors.New("synth: statutory activation not implemented in POC")
}

// ComputePay runs the pure synthetic calculation. It is deterministic:
// same inputs + same packs always yield byte-identical JSON and digest.
func ComputePay(inputs []DayInput, packs []RulePack) (PayResult, error) {
	days := append([]DayInput(nil), inputs...)
	sort.Slice(days, func(i, j int) bool { return days[i].Day < days[j].Day })

	var out PayResult
	for _, d := range days {
		if d.WorkedMin < 0 || d.MonthlyBaseCents < 0 {
			return PayResult{}, fmt.Errorf("%w: %+v", ErrBadInput, d)
		}
		pack, err := coveringPack(d.Day, packs)
		if err != nil {
			return PayResult{}, err
		}
		r := DayResult{Day: d.Day, PackVersion: pack.Version}
		if d.OnUnpaidLeave {
			// Leave day: no regular/overtime minutes, flat deduction.
			r.LeaveDeductCents = d.MonthlyBaseCents / int64(pack.LeaveDeductionDenom)
		} else {
			if worked := d.WorkedMin; worked > pack.OvertimeThresholdMin {
				r.OvertimeMin = worked - pack.OvertimeThresholdMin
				r.RegularMin = pack.OvertimeThresholdMin
			} else {
				r.RegularMin = worked
			}
			// otMin * base * numer / (denom * 30 * 8 * 60), truncated once.
			r.OvertimeCents = int64(r.OvertimeMin) * d.MonthlyBaseCents *
				int64(pack.OvertimeNumer) / (int64(pack.OvertimeDenom) * 30 * 8 * 60)
		}
		out.Days = append(out.Days, r)
		out.TotalOTCents += r.OvertimeCents
		out.TotalDedCents += r.LeaveDeductCents
	}
	raw, err := json.Marshal(out.Days)
	if err != nil {
		return PayResult{}, err
	}
	sum := sha256.Sum256(raw)
	out.Digest = hex.EncodeToString(sum[:])
	return out, nil
}

func coveringPack(day string, packs []RulePack) (RulePack, error) {
	var found *RulePack
	for i := range packs {
		if !packs[i].Covers(day) {
			continue
		}
		if found != nil {
			return RulePack{}, fmt.Errorf("%w: %s", ErrAmbiguousPack, day)
		}
		found = &packs[i]
	}
	if found == nil {
		return RulePack{}, fmt.Errorf("%w: %s", ErrNoCoveringPack, day)
	}
	return *found, nil
}
