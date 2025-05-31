package model

import "time"

type MessageStatus string

const (
	StatusUnsent  MessageStatus = "unsent"
	StatusSent    MessageStatus = "sent"
	StatusInvalid MessageStatus = "invalid"
)

type Message struct {
	ID        int
	Content   string
	Recipient string
	Status    MessageStatus
	SentAt    *time.Time
}
