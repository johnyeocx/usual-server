package db

import (
	"database/sql"
	"fmt"
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
	// this will be printed in the terminal, confirming the connection to the database
	fmt.Println("The database is connected")

	return psqlDb
}