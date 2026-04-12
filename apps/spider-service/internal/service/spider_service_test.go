package service

import "testing"

func TestCreateSpiderRejectsUnknownLanguage(t *testing.T) {
	svc := NewSpiderService()
	_, err := svc.Create("p1", "bad", "ruby", "docker")
	if err == nil {
		t.Fatal("expected validation error")
	}
}
