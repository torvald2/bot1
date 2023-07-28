package database

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"log"
)

func NewDB(connection string) (*sqlx.DB, error) {

	db, err := sqlx.Open("sqlite3", connection)
	if err != nil {
		log.Fatal(err)
	}
	db.MustExec(schemaSQL)
	return db, nil

}
