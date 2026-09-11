package dbent

import (
	"config"
	"context"
	"database/sql"
	"dbent/gen/ent"
	"fmt"
	"lace/db"
	"runtime"
	"strconv"
	"strings"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/fatih/color"
)

type EntDB struct {
	client *ent.Client
}

func (e EntDB) Client() *ent.Client {
	return e.client
}

func New(db *sql.DB) EntDB {
	// Create an ent.Driver from `db`.
	drv := entsql.OpenDB(dialect.Postgres, db)

	// client := ent.NewClient(ent.Driver(drv), ent.Debug())
	client := ent.NewClient(ent.Driver(&CustomDriver{drv}), ent.Debug())

	client.Intercept(
		ent.InterceptFunc(func(next ent.Querier) ent.Querier {
			return ent.QuerierFunc(func(ctx context.Context, query ent.Query) (ent.Value, error) {
				count, err := next.Query(ctx, query)
				return count, err
			})
		}),
	)

	// ent.InterceptFunc(func(next ent.Querier) ent.Querier {
	// 	return ent.QuerierFunc(func(ctx context.Context, query ent.Query) (ent.Value, error) {
	// 		// Do something before the query execution.
	// 		value, err := next.Query(ctx, query)
	// 		// Do something after the query execution.
	// 		return value, err
	// 	})
	// })

	return EntDB{
		client: client,
	}
}

type CustomDriver struct {
	*entsql.Driver
}

func (d *CustomDriver) Query(ctx context.Context, query string, args, v any) error {
	err := d.Driver.Query(ctx, query, args, v)
	// fmt.Println("Custom ERROR", err)
	FileWithLineNum()
	return err
}

// FileWithLineNum return the file name and line number of the current file
func FileWithLineNum() string {
	// the second caller usually from gorm internal, so set i start from 2
	for i := 1; i < 15; i++ {
		_, file, line, ok := runtime.Caller(i)
		if ok && (!strings.Contains(file, "/vendor") && !strings.Contains(file, "/ent") && !strings.Contains(file, "generated.go")) {
			// fmt.Println(file, line, ok)
			fmt.Println(time.Now().Format(time.RFC3339), color.GreenString(file), color.GreenString(strconv.FormatInt(int64(line), 10)))
			return file + ":" + strconv.FormatInt(int64(line), 10)
		}
	}

	return ""
}

func WithTx(ctx context.Context, client *ent.Client, fn func(tx *ent.Tx) error) error {
	tx, err := client.Tx(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if v := recover(); v != nil {
			tx.Rollback()
			panic(v)
		}
	}()
	if err := fn(tx); err != nil {
		if rerr := tx.Rollback(); rerr != nil {
			err = fmt.Errorf("%w: rolling back transaction: %v", err, rerr)
		}
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}
	return nil
}

// rollback calls to tx.Rollback and wraps the given error
// with the rollback error if occurred.
func Rollback(tx *ent.Tx, err error) error {
	if rerr := tx.Rollback(); rerr != nil {
		err = fmt.Errorf("%w: %v", err, rerr)
	}
	return err
}

/*
https://github.com/ent/ent/issues/4012
https://discord.com/channels/885059418646003782/885060855283216434/1331379976187936890
chats, err := s.db.Chat.Query().

	WithMessages(func(q *ent.MessageQuery) {
	    q.Order(ent.Desc(message.FieldCreatedAt))
	    q.Limit(1)
	}).
	All(ctx)
*/
func LimitRows(partitionBy string, limit int, orderBy ...string) func(s *entsql.Selector) {
	return func(s *entsql.Selector) {
		d := entsql.Dialect(s.Dialect())
		s.SetDistinct(false)
		if len(orderBy) == 0 {
			orderBy = append(orderBy, "id")
		}
		with := d.With("src_query").
			As(s.Clone()).
			With("limited_query").
			As(
				d.Select("*").
					AppendSelectExprAs(
						entsql.RowNumber().PartitionBy(partitionBy).OrderBy(entsql.Desc(orderBy[0])),
						"row_number",
					).
					From(d.Table("src_query")),
			)
		t := d.Table("limited_query").As(s.TableName())
		*s = *d.Select(s.UnqualifiedColumns()...).
			From(t).
			Where(entsql.LTE(t.C("row_number"), limit)).
			Prefix(with)
	}
}

/*
If the table is empty or column value does not exists then it gives error
failed to aggregate total: sql/scan: failed scanning rows: sql: Scan error on column index 0, name "sum": converting NULL to int is unsupported

// e.g
total, err := client.Payment.Query().Aggregate(AggOrZero(ent.Sum(payment.FieldAmount))).Int(ctx)
*/
func AggOrZero(inner ent.AggregateFunc) ent.AggregateFunc {
	return func(s *entsql.Selector) string {
		// if you put this code inside your generated ent package, then you can
		// uncomment this for better error handling, copied from the generated Sum func:
		/*
			if err := checkColumn(s.TableName(), field); err != nil {
					s.AddError(&ValidationError{Name: field, err: fmt.Errorf("ent: %w", err)})
					return ""
			}
		*/
		var b entsql.Builder
		b.WriteString("COALESCE")
		b.Wrap(func(b *entsql.Builder) {
			b.WriteString(inner(s))
			b.Comma()
			b.WriteString("0")
		})
		return b.String()
	}
}

func InitDB(databaseconfig config.Database) *sql.DB {
	return db.New(db.WithDataSourceName(dsName(databaseconfig)), db.WithDriverName("pgx"))
}

func dsName(databaseconfig config.Database) string {
	username := databaseconfig.Username
	password := databaseconfig.Password
	databaseName := databaseconfig.DbName
	databaseHost := databaseconfig.Host
	databasePort := databaseconfig.Port
	sslmode := databaseconfig.Sslmode

	return fmt.Sprintf("host=%s user=%s dbname=%s sslmode=%s password=%s port=%d", databaseHost, username, databaseName, sslmode, password, databasePort)
}
