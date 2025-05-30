package db

import (
	"database/sql"
	"fmt"
	"log"

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

	_, err = d.Conn.Exec(`INSERT INTO message (content, recipient, status) VALUES
		('Insider - Project', '+905551111111', $1),
		('Hello universe!', '+14181234567', $1)
	`, model.StatusUnsent)
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
