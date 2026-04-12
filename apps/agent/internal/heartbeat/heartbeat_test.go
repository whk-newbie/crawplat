package heartbeat

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPostHeartbeatEscapesPathAndSendsBody(t *testing.T) {
	t.Helper()

	var gotPath string
	var gotContentType string
	var gotBody string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.EscapedPath()
		gotContentType = r.Header.Get("Content-Type")
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("failed to read body: %v", err)
		}
		gotBody = string(body)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	if err := postHeartbeat(server.URL, "node/a"); err != nil {
		t.Fatalf("postHeartbeat returned error: %v", err)
	}

	if gotPath != "/api/v1/nodes/node%2Fa/heartbeat" {
		t.Fatalf("expected escaped path, got %s", gotPath)
	}
	if gotContentType != "application/json" {
		t.Fatalf("expected application/json content type, got %s", gotContentType)
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

	if err := postHeartbeat(server.URL, "node-a"); err == nil {
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
