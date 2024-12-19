package viewscount

import (
	"context"
	"database/sql"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
)

// Option 1: Convert pgxpool.Pool to sql.DB
func PgxPoolToSqlDB(pool *pgxpool.Pool) *sql.DB {
	config := pool.Config().ConnConfig
	driverName := "pgx_" + config.Database
	stdlib.RegisterConnConfig(config)

	db, err := sql.Open("pgx", driverName)
	if err != nil {
		panic(err)
	}

	db.SetMaxOpenConns(int(pool.Config().MaxConns))
	db.SetMaxIdleConns(int(pool.Config().MinConns))
	db.SetConnMaxLifetime(pool.Config().MaxConnLifetime)

	return db
}

// Option 2: Create new sql.DB from connection string
func NewSqlDBFromConnString(connString string) (*sql.DB, error) {
	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, err
	}

	driverName := "pgx_custom"
	stdlib.RegisterConnConfig(config.ConnConfig)

	return sql.Open("pgx", driverName)
}

// Option 3: Create an adapter that implements sql.DB's methods
type PgxAdapter struct {
	pool *pgxpool.Pool
}

func NewPgxAdapter(pool *pgxpool.Pool) *PgxAdapter {
	return &PgxAdapter{pool: pool}
}

// Implement the exact method needed by ViewTracker
func (a *PgxAdapter) Exec(query string, args ...interface{}) (sql.Result, error) {
	ctx := context.Background()
	result, err := a.pool.Exec(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	// Create a sql.Result compatible type
	return &pgxResult{
		rowsAffected: result.RowsAffected(),
		lastInsertId: 0, // PostgreSQL doesn't support last insert ID
	}, nil
}

type pgxResult struct {
	rowsAffected int64
	lastInsertId int64
}

func (r *pgxResult) LastInsertId() (int64, error) {
	return r.lastInsertId, nil
}

func (r *pgxResult) RowsAffected() (int64, error) {
	return r.rowsAffected, nil
}
