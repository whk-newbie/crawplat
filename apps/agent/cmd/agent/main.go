package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"crawler-platform/apps/agent/internal/heartbeat"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	baseURL := strings.TrimSpace(os.Getenv("NODE_SERVICE_URL"))
	if baseURL == "" {
		baseURL = "http://localhost:8084"
	}

	nodeName := strings.TrimSpace(os.Getenv("NODE_NAME"))
	if nodeName == "" {
		nodeName = "node-a"
	}

	if err := heartbeat.Run(ctx, baseURL, nodeName); err != nil {
		log.Fatal(err)
	}
}
