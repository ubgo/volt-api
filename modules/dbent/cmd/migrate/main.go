package main

import (
	"config"
	"context"
	"dbent"
	"dbent/pkg/dbent_migrate"
)

func main() {
	cfg := config.Get()
	db := dbent.InitDB(cfg.Database)

	dbent_migrate.Migrate(context.Background(), db)
}
