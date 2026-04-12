package proxy

import "testing"

func TestResolveServiceURL(t *testing.T) {
	url := ResolveServiceURL("iam-service")
	if url == "" {
		t.Fatal("expected non-empty url")
	}
}
