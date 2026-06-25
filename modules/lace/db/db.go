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

	return db
}
