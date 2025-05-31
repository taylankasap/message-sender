package main

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/taylankasap/message-sender/api"

	"github.com/redis/go-redis/v9"
	somethirdparty "github.com/taylankasap/message-sender/some_third_party"
)

//go:generate go tool mockgen --package=main --destination=mock_db_interface.go . DBInterface
type DBInterface interface {
	GetUnsentMessages(limit int) ([]api.Message, error)
	GetSentMessages() ([]api.Message, error)
	MarkMessageAsSent(id int, sentAt time.Time) error
	MarkMessageAsInvalid(id int) error
}

//go:generate go tool mockgen --package=main --destination=mock_redis_cache.go . RedisCache
type RedisCache interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
}

type MessageDispatcher struct {
	DB        DBInterface
	Client    somethirdparty.ClientWithResponsesInterface
	BatchSize int
	Period    time.Duration

	Redis RedisCache // Optional, can be nil

	paused   bool
	pauseMu  sync.Mutex
	pauseCh  chan struct{}
	resumeCh chan struct{}
}

type MessageDispatcherConfig struct {
	BatchSize int           // Number of messages to process in each batch
	Period    time.Duration // Time period to wait before processing the next batch
}

func NewMessageDispatcher(database DBInterface, client somethirdparty.ClientWithResponsesInterface, redisClient RedisCache, config *MessageDispatcherConfig) *MessageDispatcher {
	d := &MessageDispatcher{
		DB:        database,
		Client:    client,
		BatchSize: config.BatchSize,
		Period:    config.Period,
		Redis:     redisClient,
		pauseCh:   make(chan struct{}),
		resumeCh:  make(chan struct{}),
	}
	return d
}

func (d *MessageDispatcher) Start() {
	ticker := time.NewTicker(d.Period)
	defer ticker.Stop()

	for {
		d.pauseMu.Lock()
		paused := d.paused
		pauseCh := d.pauseCh
		d.pauseMu.Unlock()

		if paused {
			<-pauseCh // Block until resumed
			continue
		}

		d.processUnsentMessages()
		<-ticker.C
	}
}

func (d *MessageDispatcher) processUnsentMessages() {
	messages, err := d.DB.GetUnsentMessages(d.BatchSize)
	if err != nil {
		log.Printf("failed to fetch unsent messages: %v", err)
		return
	}

	var wg sync.WaitGroup
	for _, msg := range messages {
		wg.Add(1)
		go func(msg api.Message) {
			defer wg.Done()
			if len(msg.Content) > 160 {
				log.Printf("message (id=%d) exceeds 160 character limit, marking as invalid", msg.Id)
				err = d.DB.MarkMessageAsInvalid(msg.Id)
				if err != nil {
					log.Printf("failed to mark message as invalid (id=%d): %v", msg.Id, err)
				}
				return
			}
			ctx := context.Background()
			resp, err := d.Client.SendMessageWithResponse(ctx, somethirdparty.Message{
				Content: msg.Content,
				To:      msg.Recipient,
			})
			if err != nil {
				log.Printf("failed to send message (id=%d): %v", msg.Id, err)
				return
			}
			if resp.JSON202 == nil {
				log.Printf("unexpected response for message (id=%d)", msg.Id)
				return
			}
			now := time.Now()
			err = d.DB.MarkMessageAsSent(msg.Id, now)
			if err != nil {
				log.Printf("failed to update message status (id=%d): %v", msg.Id, err)
			}
			log.Printf("Message sent: id=%d, messageId=%s, sentAt=%s", msg.Id, resp.JSON202.MessageId, now)

			if d.Redis != nil {
				redisKey := "sent_message:" + strconv.Itoa(msg.Id)
				redisValue := fmt.Sprintf(`{"messageId":"%s","sentAt":"%s"}`, resp.JSON202.MessageId, now.Format(time.RFC3339))

				err := d.Redis.Set(ctx, redisKey, redisValue, 0).Err()
				if err != nil {
					log.Printf("failed to cache sent message in Redis (id=%d): %v", msg.Id, err)
				}
			}
		}(msg)
	}
	wg.Wait()
}

func (d *MessageDispatcher) Pause() {
	d.pauseMu.Lock()
	defer d.pauseMu.Unlock()
	if !d.paused {
		d.paused = true
		close(d.pauseCh)
	}
	log.Print("dispatcher is paused")
}

func (d *MessageDispatcher) Resume() {
	d.pauseMu.Lock()
	defer d.pauseMu.Unlock()
	if d.paused {
		d.paused = false
		d.pauseCh = make(chan struct{})
		close(d.resumeCh)
		d.resumeCh = make(chan struct{})
	}
	log.Print("dispatcher is resumed")
}
