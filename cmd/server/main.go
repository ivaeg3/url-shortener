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
	port := flag.String("port", "50051", "port for the gRPC server to listen on")
	storageType := flag.String("storage", "memory", "storage type to use (memory or postgres)")
	postgresURL := flag.String("postgres-url", "", "URL for the Postgres database (required if storage=postgres)")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	var store storage.Storage

	switch *storageType {
	case "memory":
		store = storage.NewMemoryStorage()
		logger.Info("using memory storage")
	case "postgres":
		if *postgresURL == "" {
			fmt.Println("-postgres-url is required if -storage=postgres")
			flag.Usage()
			os.Exit(1)
		}
		var err error
		store, err = storage.NewPostgresStorage(*postgresURL)
		if err != nil {
			logger.Error("failed to create postgres storage", "error", err)
			os.Exit(1)
		}
		logger.Info("using postgres storage", "url", *postgresURL)
	default:
		fmt.Println("invalid storage type:", *storageType)
		flag.Usage()
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

	logger.Info("server started", "port", ":"+*port)
	if err := grpcServer.Serve(lis); err != nil {
		logger.Error("server error", "error", err)
		os.Exit(1)
	}
}
