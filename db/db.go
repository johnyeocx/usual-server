package db

import (
	"database/sql"
	"os"

	_ "github.com/lib/pq"
)

func Connect() (*sql.DB) {
	connStr := os.Getenv("psqlConnString")
	psqlDb, err := sql.Open("postgres", connStr)

	
	if err != nil {
		panic(err)
	}

	if err = psqlDb.Ping(); err != nil {
		panic(err)
	}

	return psqlDb
}