package db

import (
	"context"
	"database/sql"
	"log"
	"os"
	"testing"

	"github.com/marco-almeida/mybank/internal/config"
	"github.com/marco-almeida/mybank/internal/postgresql"
)

var testStore Store
var testDB *sql.DB

func TestMain(m *testing.M) {
	config, err := config.LoadConfig("../../..")
	if err != nil {
		log.Fatal("cannot load config:", err)
	}

	connPool, err := postgresql.NewPostgreSQL(context.Background(), &config)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	testStore = NewStore(connPool)

	os.Exit(m.Run())
}
