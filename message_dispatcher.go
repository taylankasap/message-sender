package main

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/taylankasap/message-sender/model"
	somethirdparty "github.com/taylankasap/message-sender/some_third_party"
)

//go:generate go tool mockgen --package=main --destination=mock_db_interface.go . DBInterface
type DBInterface interface {
	FetchUnsentMessages(limit int) ([]model.Message, error)
	MarkMessageAsSent(id int, sentAt time.Time) error
	MarkMessageAsInvalid(id int) error
}

type MessageDispatcher struct {
	DB        DBInterface
	Client    somethirdparty.ClientWithResponsesInterface
	BatchSize int
	Period    time.Duration

	paused   bool
	pauseMu  sync.Mutex
	pauseCh  chan struct{}
	resumeCh chan struct{}
}

type MessageDispatcherConfig struct {
	BatchSize int           // Number of messages to process in each batch
	Period    time.Duration // Time period to wait before processing the next batch
}

func NewMessageDispatcher(database DBInterface, client somethirdparty.ClientWithResponsesInterface, config *MessageDispatcherConfig) *MessageDispatcher {
	d := &MessageDispatcher{
		DB:        database,
		Client:    client,
		BatchSize: config.BatchSize,
		Period:    config.Period,
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
	messages, err := d.DB.FetchUnsentMessages(d.BatchSize)
	if err != nil {
		log.Printf("failed to fetch unsent messages: %v", err)
		return
	}

	var wg sync.WaitGroup
	for _, msg := range messages {
		wg.Add(1)
		go func(msg model.Message) {
			defer wg.Done()
			if len(msg.Content) > 160 {
				log.Printf("message (id=%d) exceeds 160 character limit, marking as invalid", msg.ID)
				err = d.DB.MarkMessageAsInvalid(msg.ID)
				if err != nil {
					log.Printf("failed to mark message as invalid (id=%d): %v", msg.ID, err)
				}
				return
			}
			ctx := context.Background()
			resp, err := d.Client.SendMessageWithResponse(ctx, somethirdparty.Message{
				Content: msg.Content,
				To:      msg.Recipient,
			})
			if err != nil {
				log.Printf("failed to send message (id=%d): %v", msg.ID, err)
				return
			}
			if resp.JSON202 == nil {
				log.Printf("unexpected response for message (id=%d)", msg.ID)
				return
			}
			now := time.Now()
			err = d.DB.MarkMessageAsSent(msg.ID, now)
			if err != nil {
				log.Printf("failed to update message status (id=%d): %v", msg.ID, err)
			}
			log.Printf("Message sent: id=%d, messageId=%s, sentAt=%s", msg.ID, resp.JSON202.MessageId, now)
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
