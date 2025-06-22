package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

var testQueries *Queries

const (
	dbDriver = "postgres"
)

var (
	testDB    *sql.DB
	testStore *Store
)

func TestMain(m *testing.M) {
	const dbSource = "postgres://root:secret@localhost:5432/simple_bank?sslmode=disable"

	fmt.Println("Connecting to:", dbSource)

	var err error

	testDB, err = sql.Open(dbDriver, dbSource)
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
