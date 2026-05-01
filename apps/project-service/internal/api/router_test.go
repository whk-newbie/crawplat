package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"crawler-platform/apps/project-service/internal/model"
	"crawler-platform/apps/project-service/internal/service"
	"crawler-platform/packages/go-common/httpx"
	"github.com/gin-gonic/gin"
)

func TestCreateProjectReturnsLowerCaseJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := NewRouter(service.NewProjectService())
	req := httptest.NewRequest(http.MethodPost, "/api/v1/projects", strings.NewReader(`{"code":"core-crawlers","name":"Core Crawlers"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", w.Code)
	}

	var project model.Project
	if err := json.Unmarshal(w.Body.Bytes(), &project); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if project.ID == "" {
		t.Fatal("expected generated id")
	}
	if project.Code != "core-crawlers" {
		t.Fatalf("expected code core-crawlers, got %s", project.Code)
	}
}

func TestListProjectsReturnsPaginatedResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := service.NewProjectService()
	if _, err := svc.Create("core-crawlers", "Core Crawlers"); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	router := NewRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp struct {
		Items  []model.Project `json:"items"`
		Total  int64           `json:"total"`
		Limit  int             `json:"limit"`
		Offset int             `json:"offset"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Total != 1 {
		t.Fatalf("expected total 1, got %d", resp.Total)
	}
	if len(resp.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(resp.Items))
	}
	if resp.Items[0].Code != "core-crawlers" {
		t.Fatalf("expected code core-crawlers, got %s", resp.Items[0].Code)
	}
	if resp.Limit != 20 || resp.Offset != 0 {
		t.Fatalf("expected default pagination, got limit=%d offset=%d", resp.Limit, resp.Offset)
	}
}

func TestListProjectsRespectsPaginationParams(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := service.NewProjectService()
	for i := 0; i < 3; i++ {
		if _, err := svc.Create("code", "name"); err != nil {
			t.Fatal(err)
		}
	}
	router := NewRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects?limit=1&offset=1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var resp httpx.PaginatedResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Total != 3 {
		t.Fatalf("expected total 3, got %d", resp.Total)
	}
	if resp.Limit != 1 || resp.Offset != 1 {
		t.Fatalf("expected limit=1 offset=1, got limit=%d offset=%d", resp.Limit, resp.Offset)
	}
}
