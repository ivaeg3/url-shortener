package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net"
	"os"

	"github.com/ivaeg3/url-shortener/internal/service"
	"github.com/ivaeg3/url-shortener/internal/storage"
	"google.golang.org/grpc"

	pb "github.com/ivaeg3/url-shortener/api/proto/gen"
)

func main() {
	port := flag.String("port", getEnv("PORT", "50051"), "port for the gRPC server to listen on")
	storageType := flag.String("storage-type", getEnv("STORAGE_TYPE", "memory"), "storage type to use (memory or postgres)")
	postgresURL := flag.String("postgres-url", os.Getenv("POSTGRES_URL"), "URL for the Postgres database (required if storage=postgres)")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	store, err := initStorage(*storageType, *postgresURL, logger)
	if err != nil {
		logger.Error("failed to initialize storage", "error", err)
		os.Exit(1)
	}

	shortenerService := service.NewShortenerService(store)

	grpcServer := grpc.NewServer()
	pb.RegisterShortenerServer(grpcServer, shortenerService)

	lis, err := net.Listen("tcp", ":"+*port)
	if err != nil {
		logger.Error("failed to start server", "error", err)
		os.Exit(1)
	}

	logger.Info("server started", "port", *port)
	if err := grpcServer.Serve(lis); err != nil {
		logger.Error("server error", "error", err)
		os.Exit(1)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func initStorage(storageType, postgresURL string, logger *slog.Logger) (storage.Storage, error) {
	switch storageType {
	case "memory":
		logger.Info("using memory storage")
		return storage.NewMemoryStorage(), nil
	case "postgres":
		if postgresURL == "" {
			return nil, fmt.Errorf("postgres-url is required if storage=postgres")
		}
		store, err := storage.NewPostgresStorage(postgresURL)
		if err != nil {
			return nil, fmt.Errorf("failed to create postgres storage: %w", err)
		}
		logger.Info("using postgres storage", "url", postgresURL)
		return store, nil
	default:
		return nil, fmt.Errorf("invalid storage type: %s", storageType)
	}
}
