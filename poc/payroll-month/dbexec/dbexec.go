// Package dbexec is the U4 Tx-threading spike: the smallest interface that
// lets capture-path code run unchanged on a plain pool AND inside a DBOS
// shared-pool transaction.
//
// Spike result (VERIFIED by compilation): zero-adapter unification FAILS —
// dbos.Tx returns sysdb-internal interfaces (Result/Rows/Row), not pgx
// types. The thin adapters below are REQUIRED (each: forward + capture the
// values we need without naming internal types). Runtime proof
// (same-txn visibility via RunAsTransaction) is Gate B acceptance (B2).
// No a2 service is refactored here.
package dbexec

import (
	"context"

	"github.com/dbos-inc/dbos-transact-golang/dbos"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Result is our command outcome. RowsAffected carries an error to match the
// DBOS side (sysdb.Result); the pgx side never errors here (returns nil).
type Result interface {
	RowsAffected() (int64, error)
}

// Row is our single-row outcome.
type Row interface {
	Scan(dest ...any) error
}

// Executor is the query subset the capture path needs. Deliberately small:
// widening it is a reviewed decision, not a convenience.
type Executor interface {
	Exec(ctx context.Context, sql string, args ...any) (Result, error)
	QueryRow(ctx context.Context, sql string, args ...any) Row
}

// --- pgx side (pool and plain txn share pgx types) ---

type pgResult struct{ n int64 }

func (r pgResult) RowsAffected() (int64, error) { return r.n, nil }

// PoolExec adapts *pgxpool.Pool.
type PoolExec struct{ P *pgxpool.Pool }

func (e PoolExec) Exec(ctx context.Context, sql string, args ...any) (Result, error) {
	tag, err := e.P.Exec(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	return pgResult{n: tag.RowsAffected()}, nil
}

func (e PoolExec) QueryRow(ctx context.Context, sql string, args ...any) Row {
	return e.P.QueryRow(ctx, sql, args...)
}

// PgTxExec adapts pgx.Tx (tests, scripts, plain-pgx callers).
type PgTxExec struct{ T pgx.Tx }

func (e PgTxExec) Exec(ctx context.Context, sql string, args ...any) (Result, error) {
	tag, err := e.T.Exec(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	return pgResult{n: tag.RowsAffected()}, nil
}

func (e PgTxExec) QueryRow(ctx context.Context, sql string, args ...any) Row {
	return e.T.QueryRow(ctx, sql, args...)
}

// --- DBOS side (internal types never named: values are captured eagerly) ---

type dbosRow struct{ r any } // holds sysdb.Row; Scan forwarded below

// DbosExec adapts dbos.Tx. NOTE: sysdb.Row/Rows/Result are internal to the
// DBOS module and cannot be named here; the adapters below only work if the
// forwarded method sets match. Any mismatch fails THIS build, loudly.
type DbosExec struct{ T dbos.Tx }

func (e DbosExec) Exec(ctx context.Context, sql string, args ...any) (Result, error) {
	r, err := e.T.Exec(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	n, err := r.RowsAffected()
	if err != nil {
		return nil, err
	}
	return pgResult{n: n}, nil
}

func (e DbosExec) QueryRow(ctx context.Context, sql string, args ...any) Row {
	return dbosRow{r: e.T.QueryRow(ctx, sql, args...)}
}

// Scan forwards to the captured row. The type assertion pins the exact
// method set we depend on; a DBOS-side change breaks here, not silently.
func (r dbosRow) Scan(dest ...any) error {
	row, ok := r.r.(interface{ Scan(...any) error })
	if !ok {
		panic("dbexec: sysdb.Row lost Scan(...any) error")
	}
	return row.Scan(dest...)
}

var (
	_ Executor = PoolExec{}
	_ Executor = PgTxExec{}
	_ Executor = DbosExec{}
)

// SameTxProbe is the runtime proof body, shared by the skipped test and the
// B2 acceptance: write a marker, read it back in the SAME executor. It
// performs no commit/rollback itself; the caller owns the transaction
// boundary (pool txn, pgx.Tx, or DBOS txn).
func SameTxProbe(ctx context.Context, ex Executor, table, id string) (readBack bool, err error) {
	res, err := ex.Exec(ctx, "INSERT INTO "+table+"(id) VALUES ($1)", id)
	if err != nil {
		return false, err
	}
	if n, err := res.RowsAffected(); err != nil || n != 1 {
		return false, err
	}
	var got string
	if err := ex.QueryRow(ctx, "SELECT id FROM "+table+" WHERE id = $1", id).Scan(&got); err != nil {
		return false, err
	}
	return got == id, nil
}
