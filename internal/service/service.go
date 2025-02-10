package service

import (
	"context"
	"log/slog"

	"github.com/ivaeg3/url-shortener/internal/storage"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/ivaeg3/url-shortener/api/proto/gen"
)

type ShortenerService struct {
	pb.UnimplementedShortenerServer
	store  storage.Storage
	logger *slog.Logger
}

func NewShortenerService(store storage.Storage) *ShortenerService {
	return &ShortenerService{
		store:  store,
		logger: slog.Default(),
	}
}

func (s *ShortenerService) CreateShortURL(ctx context.Context, req *pb.CreateRequest) (*pb.CreateResponse, error) {
	const op = "internal.service.CreateShortURL"
	originalURL := req.GetOriginalUrl()

	if originalURL == "" {
		s.logger.WarnContext(ctx, "empty original URL received",
			slog.String("op", op),
		)
		return nil, status.Error(codes.InvalidArgument, "URL is required")
	}

	shortURL, err := s.store.Save(originalURL)
	if err != nil {
		if err == storage.ErrDuplicateURL {
			s.logger.InfoContext(ctx, "URL already exists",
				slog.String("op", op),
				slog.String("original_url", originalURL),
			)
			return nil, status.Error(codes.AlreadyExists, "URL already exists")
		}

		s.logger.ErrorContext(ctx, "failed to save URL",
			slog.String("op", op),
			slog.String("original_url", originalURL),
			slog.Any("error", err),
		)
		return nil, status.Errorf(codes.Internal, "failed to save URL: %v", err)
	}

	s.logger.InfoContext(ctx, "successfully created short URL",
		slog.String("op", op),
		slog.String("original_url", originalURL),
		slog.String("short_url", shortURL),
	)
	return &pb.CreateResponse{ShortUrl: shortURL}, nil
}

func (s *ShortenerService) GetOriginalURL(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
	const op = "internal.service.GetOriginalURL"
	shortURL := req.GetShortUrl()

	if shortURL == "" {
		s.logger.WarnContext(ctx, "empty short URL received",
			slog.String("op", op),
		)
		return nil, status.Error(codes.InvalidArgument, "short URL is required")
	}

	originalURL, err := s.store.Get(shortURL)
	if err != nil {
		if err == storage.ErrNotFound {
			s.logger.InfoContext(ctx, "URL not found",
				slog.String("op", op),
				slog.String("short_url", shortURL),
			)
			return nil, status.Error(codes.NotFound, "URL not found")
		}

		s.logger.ErrorContext(ctx, "failed to get original URL",
			slog.String("op", op),
			slog.String("short_url", shortURL),
			slog.Any("error", err),
		)
		return nil, status.Errorf(codes.Internal, "failed to get URL: %v", err)
	}

	s.logger.InfoContext(ctx, "successfully retrieved original URL",
		slog.String("op", op),
		slog.String("short_url", shortURL),
		slog.String("original_url", originalURL),
	)
	return &pb.GetResponse{OriginalUrl: originalURL}, nil
}
