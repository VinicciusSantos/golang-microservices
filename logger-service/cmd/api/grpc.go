package main

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/VinicciusSantos/golang-microservices/logger-service/data"
	"github.com/VinicciusSantos/golang-microservices/logger-service/logs"
	"google.golang.org/grpc"
)

type LogServer struct {
	logs.UnimplementedLoggerServiceServer
	Models data.Models
}

func (s *LogServer) WriteLog(ctx context.Context, req *logs.LogRequest) (*logs.LogResponse, error) {
	input := req.GetLogEntry()

	if err := s.Models.LogEntry.Insert(data.LogEntry{
		Name: input.Name,
		Data: input.Data,
	}); err != nil {
		return &logs.LogResponse{Result: "Error processing payload via gRPC"}, err
	}

	return &logs.LogResponse{Result: "Processed payload via gRPC"}, nil
}

func (app *Config) gRPCListen() {
	var (
		server = grpc.NewServer()
		lis    net.Listener
		err    error
	)

	if lis, err = net.Listen("tcp", fmt.Sprintf(":%s", grpcPort)); err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	logs.RegisterLoggerServiceServer(server, &LogServer{Models: app.Models})
	log.Printf("gRPC server listening on port %s", grpcPort)

	if err = server.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
