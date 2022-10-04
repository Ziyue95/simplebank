package db

import (
	"database/sql"
	"log"
	"os"
	"testing"

	"db.sqlc.dev/app/util"
	_ "github.com/lib/pq"
)

var testQueries *Queries // global Queries object
var testDB *sql.DB       // global DB object, required for unit test of transactions

func TestMain(m *testing.M) {
	config, err := util.LoadConfig("../..")
	if err != nil {
		log.Fatal("cannot load config:", err)
	}

	testDB, err = sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	testQueries = New(testDB) // use connection testDB to create the Queries object;
	os.Exit(m.Run())
}
