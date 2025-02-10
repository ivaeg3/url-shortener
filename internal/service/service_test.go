package service_test

import (
	"context"
	"testing"

	pb "github.com/ivaeg3/url-shortener/api/proto/gen"
	"github.com/ivaeg3/url-shortener/internal/service"
	"github.com/ivaeg3/url-shortener/internal/storage"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type mockStorage struct {
	storage.Storage
	store map[string]string
}

func (m *mockStorage) Save(url string) (string, error) {
	if _, exists := m.store[url]; exists {
		return "", storage.ErrDuplicateURL
	}
	shortURL := "short_" + url
	m.store[url] = shortURL
	return shortURL, nil
}

func (m *mockStorage) Get(shortURL string) (string, error) {
	for original, short := range m.store {
		if short == shortURL {
			return original, nil
		}
	}
	return "", storage.ErrNotFound
}

func TestCreateShortURL(t *testing.T) {
	store := &mockStorage{
		store: make(map[string]string),
	}
	service := service.NewShortenerService(store)

	grpcServer := grpc.NewServer()
	pb.RegisterShortenerServer(grpcServer, service)

	t.Run("Create valid short URL", func(t *testing.T) {
		req := &pb.CreateRequest{OriginalUrl: "https://www.ozon.ru"}
		resp, err := service.CreateShortURL(context.Background(), req)

		assert.NoError(t, err)
		assert.NotEmpty(t, resp.ShortUrl)
		assert.Equal(t, "short_https://www.ozon.ru", resp.ShortUrl)
	})

	t.Run("Create duplicate short URL", func(t *testing.T) {
		req := &pb.CreateRequest{OriginalUrl: "https://www.ozon.ru"}
		service.CreateShortURL(context.Background(), req)

		resp, err := service.CreateShortURL(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, status.Code(err), codes.AlreadyExists)
	})

	t.Run("Create empty URL", func(t *testing.T) {
		req := &pb.CreateRequest{OriginalUrl: ""}
		resp, err := service.CreateShortURL(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, status.Code(err), codes.InvalidArgument)
	})
}
