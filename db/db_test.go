package db_test

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/taylankasap/message-sender/db"
	"github.com/taylankasap/message-sender/model"
)

func TestNew(t *testing.T) {
	t.Run("it should create a database and the 'message' table", func(tt *testing.T) {
		testFile := "test_db.sqlite3"
		_ = os.Remove(testFile)

		database, err := db.New(&db.Config{Filename: testFile})
		require.NoError(tt, err)
		require.NotNil(tt, database.Conn)

		defer func() {
			database.Conn.Close()
			_ = os.Remove(testFile)
		}()

		row := database.Conn.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='message'")
		var tableName string
		err = row.Scan(&tableName)
		require.NoError(tt, err)
		require.Equal(tt, "message", tableName)
	})
}

func TestDatabase_FetchUnsentMessages(t *testing.T) {
	t.Run("it should fetch unsent messages", func(tt *testing.T) {
		testFile := "test_db_fetch_unsent.sqlite3"
		_ = os.Remove(testFile)

		database, err := db.New(&db.Config{Filename: testFile})
		require.NoError(tt, err)
		require.NotNil(tt, database.Conn)

		defer func() {
			database.Conn.Close()
			_ = os.Remove(testFile)
		}()

		require.NoError(tt, database.Seed())

		msgs, err := database.FetchUnsentMessages(2)
		require.NoError(tt, err)
		require.Len(tt, msgs, 2)

		for _, m := range msgs {
			require.Equal(tt, model.StatusUnsent, m.Status)
		}
	})
}

func TestDatabase_MarkMessageAsSent(t *testing.T) {
	t.Run("it should mark a message as sent and set sent_at", func(tt *testing.T) {
		testFile := "test_db_mark_sent.sqlite3"
		_ = os.Remove(testFile)

		database, err := db.New(&db.Config{Filename: testFile})
		require.NoError(tt, err)
		require.NotNil(tt, database.Conn)

		defer func() {
			database.Conn.Close()
			_ = os.Remove(testFile)
		}()

		require.NoError(tt, database.Seed())

		var id int
		row := database.Conn.QueryRow("SELECT id FROM message WHERE status = ?", model.StatusUnsent)
		require.NoError(tt, row.Scan(&id))

		expectedSentAt := time.Now()
		err = database.MarkMessageAsSent(id, expectedSentAt)
		require.NoError(tt, err)

		var actualStatus model.MessageStatus
		var actualSentAt string
		row = database.Conn.QueryRow("SELECT status, sent_at FROM message WHERE id = ?", id)
		require.NoError(tt, row.Scan(&actualStatus, &actualSentAt))
		require.Equal(tt, model.StatusSent, actualStatus)

		parsedSentAt, err := time.Parse(time.RFC3339, actualSentAt)
		require.NoError(tt, err)
		require.WithinDuration(tt, expectedSentAt.UTC(), parsedSentAt.UTC(), time.Second)
	})
}

func TestDatabase_MarkMessageAsInvalid(t *testing.T) {
	t.Run("it should mark a message as invalid", func(tt *testing.T) {
		testFile := "test_db_mark_invalid.sqlite3"
		_ = os.Remove(testFile)

		database, err := db.New(&db.Config{Filename: testFile})
		require.NoError(tt, err)
		require.NotNil(tt, database.Conn)

		defer func() {
			database.Conn.Close()
			_ = os.Remove(testFile)
		}()

		require.NoError(tt, database.Seed())

		var id int
		row := database.Conn.QueryRow("SELECT id FROM message WHERE status = ?", model.StatusUnsent)
		require.NoError(tt, row.Scan(&id))

		err = database.MarkMessageAsInvalid(id)
		require.NoError(tt, err)

		var actualStatus model.MessageStatus
		row = database.Conn.QueryRow("SELECT status FROM message WHERE id = ?", id)
		require.NoError(tt, row.Scan(&actualStatus))
		require.Equal(tt, model.StatusInvalid, actualStatus)
	})
}
