package main

import (
	"github.com/taylankasap/message-sender/db"
)

func main() {
	database, databaseErr := db.New(&db.Config{Filename: "db.sqlite3"})
	if databaseErr != nil {
		panic(databaseErr)
	}
	defer database.Conn.Close()
}
