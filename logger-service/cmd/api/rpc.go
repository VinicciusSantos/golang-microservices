package main

import (
	"context"
	"log"
	"time"

	"github.com/VinicciusSantos/golang-microservices/logger-service/data"
)

type RPCServer struct{}

type RPCPayload struct {
	Name string
	Data string
}

func (s *RPCServer) LogInfo(payload *RPCPayload, response *string) (err error) {
	if _, err = client.Database("logs").Collection("logs").InsertOne(context.TODO(), data.LogEntry{
		Name:      payload.Name,
		Data:      payload.Data,
		CreatedAt: time.Now(),
	}); err != nil {
		log.Println("Error inserting log entry:", err)
		return err
	}

	*response = "Processed payload via RPC"

	return
}
