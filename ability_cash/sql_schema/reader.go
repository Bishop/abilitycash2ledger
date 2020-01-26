package sql_schema

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

func ReadDb(fileMame string) {
	db, err := sql.Open("sqlite3", "productdb.db")

	if err != nil {
		panic(err)
	}

	defer db.Close()
}
