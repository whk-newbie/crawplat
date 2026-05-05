package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"crawler-platform/apps/agent/internal/heartbeat"
	"crawler-platform/apps/agent/internal/poller"
	agentruntime "crawler-platform/apps/agent/internal/runtime"
)

type config struct {
	nodeServiceURL      string
	executionServiceURL string
	nodeName            string
	internalToken       string
	pollInterval        time.Duration
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg := loadConfig()

	execPoller := poller.New(
		poller.NewExecutionClient(cfg.executionServiceURL, cfg.internalToken),
		agentruntime.NewDockerRunner(nil),
		cfg.nodeName,
		cfg.pollInterval,
	)

	if err := run(
		ctx,
		func(ctx context.Context) error { return heartbeat.Run(ctx, cfg.nodeServiceURL, cfg.nodeName) },
		execPoller.Run,
	); err != nil {
		log.Fatal(err)
	}
}

func loadConfig() config {
	pollInterval := 5 * time.Second
	if raw := strings.TrimSpace(os.Getenv("POLL_INTERVAL")); raw != "" {
		if parsed, err := time.ParseDuration(raw); err == nil && parsed > 0 {
			pollInterval = parsed
		}
	}

	return config{
		nodeServiceURL:      envOrDefault("NODE_SERVICE_URL", "http://localhost:8084"),
		executionServiceURL: envOrDefault("EXECUTION_SERVICE_URL", "http://localhost:8085"),
		nodeName:            envOrDefault("NODE_NAME", "node-a"),
		internalToken:       strings.TrimSpace(os.Getenv("INTERNAL_API_TOKEN")),
		pollInterval:        pollInterval,
	}
}

func envOrDefault(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func run(ctx context.Context, heartbeatRunner, pollerRunner func(context.Context) error) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	errCh := make(chan error, 2)
	var wg sync.WaitGroup

	start := func(runner func(context.Context) error) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errCh <- runner(ctx)
		}()
	}

	start(heartbeatRunner)
	start(pollerRunner)

	var runErr error
	for i := 0; i < 2; i++ {
		err := <-errCh
		if err != nil && !errors.Is(err, context.Canceled) && runErr == nil {
			runErr = err
			cancel()
		}
	}

	wg.Wait()
	return runErr
}
