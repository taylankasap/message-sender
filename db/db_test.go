package db_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/taylankasap/message-sender/db"
)

func TestNewDatabase(t *testing.T) {
	testFile := "test_db.sqlite3"
	_ = os.Remove(testFile)

	database, err := db.New(&db.Config{Filename: testFile})
	require.NoError(t, err)
	require.NotNil(t, database.Conn)

	row := database.Conn.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='message'")
	var tableName string
	err = row.Scan(&tableName)
	require.NoError(t, err)
	require.Equal(t, "message", tableName)

	database.Conn.Close()
	_ = os.Remove(testFile)
}
