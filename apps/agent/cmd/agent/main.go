package main

import (
	"context"
	"errors"
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

	if err := run(ctx, baseURL, nodeName, heartbeat.Run); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context, baseURL, nodeName string, runner func(context.Context, string, string) error) error {
	if err := runner(ctx, baseURL, nodeName); err != nil && !errors.Is(err, context.Canceled) {
		return err
	}
	return nil
}
