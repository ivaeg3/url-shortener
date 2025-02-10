package main

import (
	"flag"
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
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	store := storage.NewMemoryStorage()
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
