package model

import "time"

type Message struct {
	ID        int
	Content   string
	Recipient string
	Status    string
	SentAt    *time.Time
}
