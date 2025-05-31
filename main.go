package main

import (
	"fmt"
	"time"

	"github.com/taylankasap/message-sender/db"
	somethirdparty "github.com/taylankasap/message-sender/some_third_party"
)

func main() {
	database, databaseErr := db.New(&db.Config{Filename: "db.sqlite3"})
	if databaseErr != nil {
		panic(databaseErr)
	}
	defer database.Conn.Close()

	if err := database.Seed(); err != nil {
		panic(fmt.Errorf("failed to seed database: %w", err))
	}

	const someThirdPartyBaseUrl = "https://webhook.site/7f22bdd7-ae91-48ca-be94-90bc6688bac1"
	client, err := somethirdparty.NewClientWithResponses(someThirdPartyBaseUrl)
	if err != nil {
		panic(err)
	}

	dispatcherConfig := &MessageDispatcherConfig{
		Period:    2 * time.Minute,
		BatchSize: 2,
	}
	dispatcher := NewMessageDispatcher(database, client, dispatcherConfig)
	go dispatcher.Start()

	select {} // keep main alive
}
