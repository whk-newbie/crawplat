package service

import "testing"

func TestCreateDatasourceRejectsUnknownType(t *testing.T) {
	svc := NewDatasourceService()
	_, err := svc.Create("project-1", "main", "mysql")
	if err == nil {
		t.Fatal("expected validation error")
	}
}
