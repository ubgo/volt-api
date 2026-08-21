package db

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type DB *sql.DB

func New(opts ...Option) *sql.DB {
	cfg := config{
		DataSourceName: "host=localhost user=postgres dbname=lacedb sslmode=disable password=postgres port=5432",
		DriverName:     "pgx",
	}

	cfg.options(opts...)

	db, err := sql.Open(cfg.DriverName, cfg.DataSourceName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}

	// Apply churn-safe pool sizing (see config.go constants for the incident
	// this prevents). Zero-value knobs take the defaults; explicit options win.
	maxOpen := cfg.MaxOpenConns
	if maxOpen == 0 {
		maxOpen = DefaultMaxOpenConns
	}
	maxIdle := cfg.MaxIdleConns
	if maxIdle == 0 {
		maxIdle = DefaultMaxIdleConns
	}
	lifetime := cfg.ConnMaxLifetime
	if lifetime == 0 {
		lifetime = DefaultConnMaxLifetime
	}
	idleTime := cfg.ConnMaxIdleTime
	if idleTime == 0 {
		idleTime = DefaultConnMaxIdleTime
	}
	db.SetMaxOpenConns(maxOpen)
	db.SetMaxIdleConns(maxIdle)
	db.SetConnMaxLifetime(lifetime)
	db.SetConnMaxIdleTime(idleTime)

	return db
}
