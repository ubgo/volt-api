package dbent_migrate

import (
	"context"
	"database/sql"
	"dbent/gen/ent"
	"dbent/gen/ent/migrate"
	"fmt"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
)

func Migrate(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return fmt.Errorf("db is nil")
	}
	drv := entsql.OpenDB(dialect.Postgres, db)
	client := ent.NewClient(ent.Driver(drv), ent.Debug())
	defer client.Close()

	// Run the auto migration tool.
	err := client.Schema.Create(
		context.Background(),
		migrate.WithForeignKeys(false), // Disable foreign keys.
	)

	if err != nil {
		// gozap.FromCtx(ctx).Error("failed creating schema resources", zap.Error(err))
		return err
	}

	return nil
}
