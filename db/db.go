package db

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/taylankasap/message-sender/model"
)

type Database struct {
	Conn *sql.DB
}

type Config struct {
	Filename string
}

func New(cfg *Config) (*Database, error) {
	db, err := sql.Open("sqlite3", cfg.Filename)
	if err != nil {
		log.Fatal("failed to open database:", err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS message (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		content TEXT NOT NULL,
		recipient TEXT NOT NULL,
		status TEXT NOT NULL DEFAULT 'unsent',
		sent_at DATETIME
	)`)
	if err != nil {
		return nil, fmt.Errorf("failed to create table: %w", err)
	}

	return &Database{Conn: db}, nil
}

// Seed inserts initial messages if the table is empty
func (d *Database) Seed() error {
	row := d.Conn.QueryRow("SELECT COUNT(*) FROM message")

	var count int

	err := row.Scan(&count)
	if err != nil {
		return err
	}

	if count > 0 {
		return nil // already seeded
	}

	_, err = d.Conn.Exec(`INSERT INTO message (content, recipient, status, sent_at) VALUES
		('Huge sale :)', '+905551234567', $1, '2024-02-12T03:00:06+03:00'),
		('Insider - Project', '+905551111111', $2, NULL),
		('Tiny sale :(', '+905551234567', $1, '2025-05-30T21:17:09+07:00'),
		('Hello universe!', '+14181234567', $2, NULL),
		('You can use this one time password to log in to somewhere: 526184', '+821260542022', $2, NULL),
		('Check out our products!', '+821251876804', $2, NULL)
	`, model.StatusSent, model.StatusUnsent)
	return err
}

// FetchUnsentMessages fetches up to n unsent messages from the database
func (d *Database) FetchUnsentMessages(limit int) ([]model.Message, error) {
	rows, err := d.Conn.Query("SELECT id, content, recipient, status, sent_at FROM message WHERE status = $1 ORDER BY id ASC LIMIT $2", model.StatusUnsent, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []model.Message
	for rows.Next() {
		var m model.Message
		err := rows.Scan(&m.ID, &m.Content, &m.Recipient, &m.Status, &m.SentAt)
		if err != nil {
			return nil, err
		}
		messages = append(messages, m)
	}

	return messages, nil
}

// MarkMessageAsSent updates the status and sent_at fields for a message
func (d *Database) MarkMessageAsSent(id int, sentAt time.Time) error {
	_, err := d.Conn.Exec(
		"UPDATE message SET status = ?, sent_at = ? WHERE id = ?",
		model.StatusSent, sentAt.Format(time.RFC3339), id,
	)
	return err
}
