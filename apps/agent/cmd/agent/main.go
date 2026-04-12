package main

import (
	"context"
	"log"
	"os"
	"strings"

	"crawler-platform/apps/agent/internal/heartbeat"
)

func main() {
	baseURL := strings.TrimSpace(os.Getenv("NODE_SERVICE_URL"))
	if baseURL == "" {
		baseURL = "http://localhost:8084"
	}

	nodeName := strings.TrimSpace(os.Getenv("NODE_NAME"))
	if nodeName == "" {
		nodeName = "node-a"
	}

	if err := heartbeat.Run(context.Background(), baseURL, nodeName); err != nil {
		log.Fatal(err)
	}
}
