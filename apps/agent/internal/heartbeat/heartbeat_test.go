package heartbeat

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPostHeartbeatSendsExpectedRequest(t *testing.T) {
	t.Helper()

	var gotPath string
	var gotContentType string
	var gotBody string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotContentType = r.Header.Get("Content-Type")
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("failed to read body: %v", err)
		}
		gotBody = string(body)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	if err := postHeartbeat(context.Background(), server.URL, "node-a"); err != nil {
		t.Fatalf("postHeartbeat returned error: %v", err)
	}

	if gotPath != "/api/v1/nodes/node-a/heartbeat" {
		t.Fatalf("expected request path /api/v1/nodes/node-a/heartbeat, got %s", gotPath)
	}
	if gotContentType != "application/json" {
		t.Fatalf("expected content type application/json, got %s", gotContentType)
	}
	if gotBody != `{"capabilities":["docker","python","go"]}` {
		t.Fatalf("unexpected body: %s", gotBody)
	}
}

func TestPostHeartbeatReturnsErrorOnNon2xx(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	if err := postHeartbeat(context.Background(), server.URL, "node-a"); err == nil {
		t.Fatal("expected error for non-2xx response")
	}
}

func TestRunReturnsErrorOnInitialHeartbeatFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer server.Close()

	if err := Run(context.Background(), server.URL, "node-a"); err == nil {
		t.Fatal("expected startup heartbeat failure")
	}
}

func TestPostHeartbeatRejectsInvalidNodeName(t *testing.T) {
	called := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	err := postHeartbeat(context.Background(), server.URL, "node/a")
	if err == nil {
		t.Fatal("expected invalid node name error")
	}
	if called {
		t.Fatal("expected no request to be sent for invalid node name")
	}
}

func TestPostHeartbeatRespectsContextCancellation(t *testing.T) {
	started := make(chan struct{}, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started <- struct{}{}
		<-r.Context().Done()
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := postHeartbeat(ctx, server.URL, "node-a")
	if err == nil || !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context canceled error, got %v", err)
	}
	select {
	case <-started:
		t.Fatal("expected request not to reach handler when context is canceled")
	default:
	}
}
