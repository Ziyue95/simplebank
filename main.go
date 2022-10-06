package main

import (
	"database/sql"
	"log"

	"db.sqlc.dev/app/api"
	db "db.sqlc.dev/app/db/sqlc"
	"db.sqlc.dev/app/util"
	_ "github.com/lib/pq"
)

func main() {
	// load config values using Viper
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal("cannot load config:", err)
	}

	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	// create store object to support db operation using connect conn
	store := db.NewStore(conn)
	// create server object
	server, err := api.NewServer(config, store)
	if err != nil {
		log.Fatal("cannot create server:", err)
	}

	err = server.Start(config.ServerAddress)
	if err != nil {
		log.Fatal("cannot start server:", err)
	}

}
