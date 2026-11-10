package seedpoc

// Collateral for the RC-1 repair: the DB-backed calc can only reproduce
// the canonical as-given sums if the seeder persisted them. Requires the
// seeded POC stack; skips loudly without it (never silent green).

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/PithomLabs/a2/poc/payroll-month/workload"
)

func TestSeedCompLinesMatchWorkload(t *testing.T) {
	dsn := os.Getenv("POC_DATABASE_URL")
	if dsn == "" {
		t.Skip("POC_DATABASE_URL unset: needs the seeded POC stack")
	}
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer pool.Close()
	ctx := context.Background()
	w := workload.Build()
	for _, e := range w.Employees {
		var nAllow, nDeduct int
		var sumAllow, sumDeduct int64
		err := pool.QueryRow(ctx,
			`SELECT count(*) FILTER (WHERE kind = 'allowance'),
			        count(*) FILTER (WHERE kind = 'deduction'),
			        COALESCE(sum(amount_monthly) FILTER (WHERE kind = 'allowance'), 0) * 100,
			        COALESCE(sum(amount_monthly) FILTER (WHERE kind = 'deduction'), 0) * 100
			 FROM compensation_line c JOIN employee m ON m.id = c.employee_id
			 WHERE m.employer_id = (SELECT id FROM employer WHERE code = 'POC-MERIDIAN')
			   AND m.employee_no = $1 AND c.effective_from = '2026-10-01'`,
			e.ID).Scan(&nAllow, &nDeduct, &sumAllow, &sumDeduct)
		if err != nil {
			t.Fatalf("%s: %v", e.ID, err)
		}
		var wantAllow, wantDeduct int64
		if e.Blocked == nil {
			for _, ag := range e.AsGiven {
				if ag.Kind == "allowance" {
					wantAllow += ag.Cents
				} else {
					wantDeduct += ag.Cents
				}
			}
		}
		if sumAllow != wantAllow || sumDeduct != wantDeduct {
			t.Fatalf("%s: persisted allow/deduct %d/%d, workload wants %d/%d",
				e.ID, sumAllow, sumDeduct, wantAllow, wantDeduct)
		}
	}
}
