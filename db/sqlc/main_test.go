package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/itsadijmbt/simple_bank/db/util"
	_ "github.com/lib/pq"
)

var testQueries *Queries

var (
	testDB    *sql.DB
	testStore *Store
)

func TestMain(m *testing.M) {
	config, err := util.LoadConfig("../..")

	if err != nil {
		log.Fatal("Comfig Load failed~!")
	}

	fmt.Println("Connecting to:", config.DBSource)

	testDB, err = sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatalf("cannot open db testDBection: %v", err)
	}

	if err = testDB.Ping(); err != nil {
		log.Fatalf("cannot ping db: %v", err)
	}
	testStore = NewStore(testDB)
	testQueries = New(testDB)

	code := m.Run()

	os.Exit(code)
}
