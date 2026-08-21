package db

import "time"

// Pool defaults. Go's database/sql ships MaxIdleConns=2 with UNLIMITED opens —
// a churn machine under burst load: the pool opens dozens of conns, closes all
// but 2 as queries finish, reopens on the next query… Each open/close leaves a
// TIME_WAIT socket, and ~28k of those to one (ip,port) within 60s exhausts the
// ephemeral-port tuple space → `connect: cannot assign requested address`
// (live incident 2026-08-18 on a Volt-derived app: nine concurrent burst jobs
// killed dials to pgbouncer while total ESTABLISHED sockets never left double
// digits — see volt/docs/incidents/2026-08-18-db-pool-connection-exhaustion.md).
//
// The cure is idle == open (closed conns are the problem, not open ones) plus
// a bounded open count. Sizing: 30 per process × 5 processes per box = 150,
// comfortably under pgbouncer's max_client_conn=1000; pgbouncer's
// transaction-mode pooling multiplexes these onto default_pool_size real
// Postgres conns, so idle client conns are nearly free.
const (
	// DefaultMaxOpenConns bounds concurrent DB work per process.
	DefaultMaxOpenConns = 30
	// DefaultMaxIdleConns == DefaultMaxOpenConns — NEVER lower: every point of
	// gap converts finished queries into close+reopen churn.
	DefaultMaxIdleConns = 30
	// DefaultConnMaxLifetime recycles conns so pgbouncer/Postgres restarts and
	// DNS moves are picked up without a process restart.
	DefaultConnMaxLifetime = 30 * time.Minute
	// DefaultConnMaxIdleTime trims a burst's idle tail gradually (way past
	// TIME_WAIT concerns at this cadence).
	DefaultConnMaxIdleTime = 5 * time.Minute
)

type config struct {
	DataSourceName string
	DriverName     string
	// Pool knobs — zero values are replaced by the Default* constants in New,
	// so a caller passing no options still gets a churn-safe pool.
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

func (c *config) options(opts ...Option) {
	for _, opt := range opts {
		opt(c)
	}
}

type Option func(*config)

func WithDataSourceName(name string) Option {
	return func(c *config) {
		c.DataSourceName = name
	}
}

func WithDriverName(name string) Option {
	return func(c *config) {
		c.DriverName = name
	}
}

// WithMaxOpenConns / WithMaxIdleConns / WithConnMaxLifetime /
// WithConnMaxIdleTime override the pool defaults (see the constants above for
// why the defaults are what they are — in particular keep idle == open).
func WithMaxOpenConns(n int) Option { return func(c *config) { c.MaxOpenConns = n } }

func WithMaxIdleConns(n int) Option { return func(c *config) { c.MaxIdleConns = n } }

func WithConnMaxLifetime(d time.Duration) Option {
	return func(c *config) { c.ConnMaxLifetime = d }
}

func WithConnMaxIdleTime(d time.Duration) Option {
	return func(c *config) { c.ConnMaxIdleTime = d }
}
