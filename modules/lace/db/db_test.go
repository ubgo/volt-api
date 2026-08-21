package db

import (
	"testing"
	"time"
)

// TestPoolDefaultsApplied — the churn-safe sizing must be ON with no options:
// bounded opens, idle == open (the 2026-08-18 EADDRNOTAVAIL incident was Go's
// MaxIdleConns=2 default converting burst load into TIME_WAIT floods).
func TestPoolDefaultsApplied(t *testing.T) {
	d := New() // pgx driver, never dialed — sql.Open is lazy
	t.Cleanup(func() { _ = d.Close() })
	if got := d.Stats().MaxOpenConnections; got != DefaultMaxOpenConns {
		t.Fatalf("MaxOpenConnections = %d, want %d", got, DefaultMaxOpenConns)
	}
	// MaxIdleConns has no Stats() mirror; the invariant that matters is the
	// constants agreeing — a future edit lowering idle below open reintroduces
	// the churn.
	if DefaultMaxIdleConns != DefaultMaxOpenConns {
		t.Fatalf("DefaultMaxIdleConns (%d) must equal DefaultMaxOpenConns (%d)", DefaultMaxIdleConns, DefaultMaxOpenConns)
	}
	if DefaultConnMaxLifetime < 10*time.Minute {
		t.Fatalf("lifetime %v suspiciously low — recycling faster than needed is churn again", DefaultConnMaxLifetime)
	}
}

// TestPoolOptionsOverride — explicit options beat the defaults.
func TestPoolOptionsOverride(t *testing.T) {
	d := New(WithMaxOpenConns(7), WithMaxIdleConns(7))
	t.Cleanup(func() { _ = d.Close() })
	if got := d.Stats().MaxOpenConnections; got != 7 {
		t.Fatalf("MaxOpenConnections = %d, want 7", got)
	}
}
