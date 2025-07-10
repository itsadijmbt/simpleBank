package main

import (
	"database/sql"
	"log"

	"github.com/itsadijmbt/simple_bank/api"
	db "github.com/itsadijmbt/simple_bank/db/sqlc"
	"github.com/itsadijmbt/simple_bank/db/util"
	_ "github.com/lib/pq"
)

func main() {
	config, err := util.LoadConfig(".")

	if err != nil {
		log.Fatal("Cannot Load Config Files")
	}
	//& create a server we need to connect a db and connect to store first

	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatalf("cannot open db testDBection: %v", err)
	}

	store := db.NewStore(conn)

	server, err := api.NewServer(config, store)

	if err != nil {
		log.Fatal("cannot create server", err)
	}
	err = server.Start(config.ServerAddress)

	if err != nil {
		log.Fatal("cannot start server")
	}
}
