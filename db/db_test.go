package db_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/taylankasap/message-sender/db"
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
			require.Equal(tt, "unsent", m.Status)
		}
	})
}
