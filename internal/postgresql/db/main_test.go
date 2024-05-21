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

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		config.PostgresUser,
		config.PostgresPassword,
		config.PostgresHost,
		config.PostgresPort,
		config.PostgresDatabase)
		
	connPool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	testStore = NewStore(connPool)

	os.Exit(m.Run())
}
