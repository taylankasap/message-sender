package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/taylankasap/message-sender/api"

	"github.com/taylankasap/message-sender/db"
	somethirdparty "github.com/taylankasap/message-sender/some_third_party"
)

func main() {
	// database
	database, databaseErr := db.New(&db.Config{Filename: "data/db.sqlite3"})
	if databaseErr != nil {
		panic(databaseErr)
	}
	defer database.Conn.Close()

	if err := database.Seed(); err != nil {
		panic(fmt.Errorf("failed to seed database: %w", err))
	}

	// third party client
	const someThirdPartyBaseUrl = "https://webhook.site/e8318d16-f749-428e-9103-f1ca43e8c0dd"
	client, err := somethirdparty.NewClientWithResponses(someThirdPartyBaseUrl)
	if err != nil {
		panic(err)
	}

	// Redis
	redisAddr := "redis:6379"
	redisClient := NewRedisClient(redisAddr)

	// message dispatcher
	dispatcherConfig := &MessageDispatcherConfig{
		Period:    2 * time.Minute,
		BatchSize: 2,
	}

	dispatcher := NewMessageDispatcher(database, client, redisClient, dispatcherConfig)
	go dispatcher.Start()

	// API server
	server := api.NewServer(database, dispatcher)

	r := http.NewServeMux()

	h := api.HandlerFromMux(server, r)

	s := &http.Server{
		Handler: h,
		Addr:    "0.0.0.0:8080",
	}

	log.Fatal(s.ListenAndServe())
}
