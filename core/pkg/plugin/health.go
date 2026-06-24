package plugin

import (
	"context"
	"database/sql"
	"time"

	"github.com/ubgo/health"
)

// sqlDBChecker is a health.Checker for a database/sql connection pool.
//
// The boilerplate connects to Postgres through the pgx stdlib driver as a
// *sql.DB, so the contrib health-postgres checker (which expects a pgx-style
// Ping(ctx) error signature on *pgxpool.Pool) does not fit. This local
// checker pings via *sql.DB.PingContext instead, honouring the registry's
// per-check timeout.
type sqlDBChecker struct {
	name string
	db   *sql.DB
}

// newSQLDBChecker returns a critical-severity checker named name that pings db.
func newSQLDBChecker(name string, db *sql.DB) *sqlDBChecker {
	return &sqlDBChecker{name: name, db: db}
}

// Name implements health.Checker.
func (c *sqlDBChecker) Name() string { return c.name }

// Check implements health.Checker: it pings the pool honouring ctx and reports
// Up on success, or Down (critical) with the error string on failure.
func (c *sqlDBChecker) Check(ctx context.Context) health.Result {
	start := time.Now()
	err := c.db.PingContext(ctx)
	r := health.Result{
		Severity:  health.SeverityCritical,
		Latency:   time.Since(start),
		Timestamp: time.Now(),
	}
	if err != nil {
		r.Status = health.StatusDown
		r.Error = err.Error()
		return r
	}
	r.Status = health.StatusUp
	return r
}
