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

//!Because of that, a variable of type *Store has zero methods — hence the
//!“undefined” error when you call TransferTx BECAUSE ompiler can’t attach methods to it because it has no method set.

var (
	testDB    *sql.DB
	testStore Store
	//* testStore *Store -> X as interface has no set methods they are defined later
)
//test
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
