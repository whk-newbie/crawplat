package api

import "testing"

func TestNewRouterDoesNotPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("expected router construction to not panic, got %v", r)
		}
	}()

	router := NewRouter()
	if router == nil {
		t.Fatal("expected router to be non-nil")
	}
}
