package storage

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getTestDSN() string {
	dsn := os.Getenv("TEST_DB_DSN")
	if dsn == "" {
		dsn = "postgres://postgres:password@localhost:5432/testdb?sslmode=disable"
	}
	return dsn
}

func setupTestStorage(t *testing.T) *PostgresStorage {
	storage, err := NewPostgresStorage(getTestDSN())
	require.NoError(t, err, "Failed to create test storage")

	_, err = storage.pool.Exec(context.Background(), "TRUNCATE TABLE urls")
	require.NoError(t, err, "Failed to truncate table")

	return storage
}

func TestNewPostgresStorage(t *testing.T) {
	t.Run("Valid DSN", func(t *testing.T) {
		storage := setupTestStorage(t)
		defer storage.Close()

		var exists bool
		err := storage.pool.QueryRow(context.Background(), `
			SELECT EXISTS (
				SELECT FROM information_schema.tables 
				WHERE table_name = 'urls'
			)`).Scan(&exists)

		assert.NoError(t, err)
		assert.True(t, exists, "Table 'urls' should exist")
	})

	t.Run("Invalid DSN", func(t *testing.T) {
		_, err := NewPostgresStorage("invalid-dsn")
		assert.Error(t, err, "Should return error for invalid DSN")
	})
}

func TestPostgresStorage_Save(t *testing.T) {
	storage := setupTestStorage(t)
	defer storage.Close()

	originalURL := "https://www.ozon.ru"

	t.Run("Save new URL", func(t *testing.T) {
		shortURL, err := storage.Save(originalURL)
		assert.NoError(t, err)
		assert.NotEmpty(t, shortURL, "Short URL should not be empty")
	})

	t.Run("Save duplicate URL", func(t *testing.T) {
		shortURL1, err := storage.Save(originalURL)
		assert.NoError(t, err)

		shortURL2, err := storage.Save(originalURL)
		assert.NoError(t, err)
		assert.Equal(t, shortURL1, shortURL2, "Duplicate URL should return same short URL")
	})

	t.Run("Counter increments", func(t *testing.T) {
		initialCounter := storage.counter
		_, err := storage.Save("https://finance.ozon.ru")
		assert.NoError(t, err)
		assert.Equal(t, initialCounter+1, storage.counter, "Counter should increment by 1")
	})
}

func TestPostgresStorage_Get(t *testing.T) {
	storage := setupTestStorage(t)
	defer storage.Close()

	originalURL := "https://www.ozon.ru"
	shortURL, err := storage.Save(originalURL)
	require.NoError(t, err)

	t.Run("Get existing URL", func(t *testing.T) {
		resultURL, err := storage.Get(shortURL)
		assert.NoError(t, err)
		assert.Equal(t, originalURL, resultURL, "Should return correct original URL")
	})

	t.Run("Get non-existent URL", func(t *testing.T) {
		_, err := storage.Get("invalid-short")
		assert.Equal(t, ErrNotFound, err, "Should return NotFound error")
	})
}

func TestPostgresStorage_Close(t *testing.T) {
	storage := setupTestStorage(t)
	assert.NotPanics(t, func() {
		storage.Close()
	}, "Close should not panic")
}
