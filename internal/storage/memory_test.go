package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMemoryStorage_SaveAndGet(t *testing.T) {
	store := NewMemoryStorage()
	originalURL := "https://www.ozon.ru"

	t.Run("Save and Get the same URL", func(t *testing.T) {
		shortURL, err := store.Save(originalURL)
		assert.NoError(t, err, "unexpected error while saving URL")

		shortURL2, err := store.Save(originalURL)
		assert.NoError(t, err, "unexpected error while saving URL")
		assert.Equal(t, shortURL, shortURL2, "expected same short URL for duplicate original URL")

		retrievedURL, err := store.Get(shortURL)
		assert.NoError(t, err, "unexpected error while retrieving URL")
		assert.Equal(t, originalURL, retrievedURL, "expected %s, got %s", originalURL, retrievedURL)
	})
}

func TestMemoryStorage_GetNotFound(t *testing.T) {
	store := NewMemoryStorage()

	t.Run("Get non-existent URL", func(t *testing.T) {
		_, err := store.Get("nonexistent")
		assert.Equal(t, ErrNotFound, err, "expected ErrNotFound, got %v", err)
	})
}
