package main

import (
	"context"
	"errors"
	"testing"
)

func TestRunTreatsContextCanceledAsCleanShutdown(t *testing.T) {
	err := run(context.Background(), "http://example.com", "node-a", func(context.Context, string, string) error {
		return context.Canceled
	})
	if err != nil {
		t.Fatalf("expected clean shutdown, got %v", err)
	}
}

func TestRunReturnsUnexpectedErrors(t *testing.T) {
	want := errors.New("boom")
	err := run(context.Background(), "http://example.com", "node-a", func(context.Context, string, string) error {
		return want
	})
	if !errors.Is(err, want) {
		t.Fatalf("expected unexpected error to be returned, got %v", err)
	}
}
