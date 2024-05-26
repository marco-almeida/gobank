package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/marco-almeida/mybank/internal/config"
)

var testStore Store
var testDB *sql.DB

func TestMain(m *testing.M) {
	config, err := config.LoadConfig("../../..")
	if err != nil {
		log.Fatal("cannot load config:", err)
	}

	connPool, err := newPostgreSQL(context.Background(), &config)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	testStore = NewStore(connPool)

	os.Exit(m.Run())
}

func newPostgreSQL(ctx context.Context, cfg *config.Config) (*pgxpool.Pool, error) {
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.PostgresUser,
		cfg.PostgresPassword,
		cfg.PostgresHost,
		cfg.PostgresPort,
		cfg.PostgresDatabase)

	return pgxpool.New(ctx, connStr)
}
