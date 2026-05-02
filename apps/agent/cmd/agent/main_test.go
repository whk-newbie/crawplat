package main

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestRunTreatsContextCanceledAsCleanShutdown(t *testing.T) {
	err := run(
		context.Background(),
		func(context.Context) error { return context.Canceled },
		func(context.Context) error { return context.Canceled },
	)
	if err != nil {
		t.Fatalf("expected clean shutdown, got %v", err)
	}
}

func TestRunReturnsUnexpectedErrors(t *testing.T) {
	want := errors.New("boom")
	err := run(
		context.Background(),
		func(context.Context) error { return want },
		func(context.Context) error { return context.Canceled },
	)
	if !errors.Is(err, want) {
		t.Fatalf("expected unexpected error to be returned, got %v", err)
	}
}

func TestLoadConfigUsesDefaults(t *testing.T) {
	cfg := loadConfig()

	if cfg.nodeServiceURL != "http://localhost:8084" {
		t.Fatalf("unexpected node service url: %q", cfg.nodeServiceURL)
	}
	if cfg.executionServiceURL != "http://localhost:8085" {
		t.Fatalf("unexpected execution service url: %q", cfg.executionServiceURL)
	}
	if cfg.nodeName != "node-a" {
		t.Fatalf("unexpected node name: %q", cfg.nodeName)
	}
	if cfg.pollInterval != 5*time.Second {
		t.Fatalf("unexpected poll interval: %v", cfg.pollInterval)
	}
	if len(cfg.registryCreds) != 0 {
		t.Fatalf("expected no registry credentials by default, got %+v", cfg.registryCreds)
	}
}

func TestLoadConfigReadsEnvironment(t *testing.T) {
	t.Setenv("NODE_SERVICE_URL", "http://node-service:8084")
	t.Setenv("EXECUTION_SERVICE_URL", "http://execution-service:8085")
	t.Setenv("NODE_NAME", "mvp-node")
	t.Setenv("INTERNAL_API_TOKEN", "secret")
	t.Setenv("POLL_INTERVAL", "2s")
	t.Setenv("IMAGE_REGISTRY_AUTH_MAP", `{"ghcr.io":{"username":"user","password":"pass"}}`)

	cfg := loadConfig()

	if cfg.nodeServiceURL != "http://node-service:8084" {
		t.Fatalf("unexpected node service url: %q", cfg.nodeServiceURL)
	}
	if cfg.executionServiceURL != "http://execution-service:8085" {
		t.Fatalf("unexpected execution service url: %q", cfg.executionServiceURL)
	}
	if cfg.nodeName != "mvp-node" {
		t.Fatalf("unexpected node name: %q", cfg.nodeName)
	}
	if cfg.internalToken != "secret" {
		t.Fatalf("unexpected internal token: %q", cfg.internalToken)
	}
	if cfg.pollInterval != 2*time.Second {
		t.Fatalf("unexpected poll interval: %v", cfg.pollInterval)
	}
	cred, ok := cfg.registryCreds["ghcr.io"]
	if !ok || cred.Username != "user" || cred.Password != "pass" {
		t.Fatalf("unexpected registry credentials: %+v", cfg.registryCreds)
	}
}
