// Package api 的 HTTP 层集成测试，使用 httptest 验证路由行为。
package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"crawler-platform/apps/project-service/internal/model"
	"crawler-platform/apps/project-service/internal/service"
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
	if project.Name != "Core Crawlers" {
		t.Fatalf("expected name Core Crawlers, got %s", project.Name)
	}
	if !strings.Contains(w.Body.String(), `"id":`) || !strings.Contains(w.Body.String(), `"code":`) || !strings.Contains(w.Body.String(), `"name":`) {
		t.Fatalf("expected lower-case JSON keys, got %s", w.Body.String())
	}
}

func TestListProjectsReturnsLowerCaseJSON(t *testing.T) {
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

	var projects []model.Project
	if err := json.Unmarshal(w.Body.Bytes(), &projects); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(projects) != 1 {
		t.Fatalf("expected 1 project, got %d", len(projects))
	}
	if projects[0].ID == "" {
		t.Fatal("expected generated id")
	}
	if projects[0].Code != "core-crawlers" {
		t.Fatalf("expected code core-crawlers, got %s", projects[0].Code)
	}
	if projects[0].Name != "Core Crawlers" {
		t.Fatalf("expected name Core Crawlers, got %s", projects[0].Name)
	}
	if !strings.Contains(w.Body.String(), `"id":`) || !strings.Contains(w.Body.String(), `"code":`) || !strings.Contains(w.Body.String(), `"name":`) {
		t.Fatalf("expected lower-case JSON keys, got %s", w.Body.String())
	}
}

func TestCreateProjectRejectsDuplicateCode(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := NewRouter(service.NewProjectService())
	req := httptest.NewRequest(http.MethodPost, "/api/v1/projects", strings.NewReader(`{"code":"dup","name":"First"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201 for first create, got %d", w.Code)
	}

	req2 := httptest.NewRequest(http.MethodPost, "/api/v1/projects", strings.NewReader(`{"code":"dup","name":"Second"}`))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	if w2.Code != http.StatusConflict {
		t.Fatalf("expected status 409 for duplicate code, got %d", w2.Code)
	}
}

func TestCreateProjectReturnsFriendlyErrorForBadJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := NewRouter(service.NewProjectService())
	req := httptest.NewRequest(http.MethodPost, "/api/v1/projects", strings.NewReader(`not json`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "invalid request body") {
		t.Fatalf("expected friendly error message, got %s", w.Body.String())
	}
}

func TestListProjectsWithPagination(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := service.NewProjectService()
	// 创建 5 个项目
	for i := 0; i < 5; i++ {
		code := "code-" + string(rune('a'+i))
		if _, err := svc.Create(code, "Project"); err != nil {
			t.Fatalf("Create returned error: %v", err)
		}
	}
	router := NewRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects?limit=3&offset=1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var projects []model.Project
	if err := json.Unmarshal(w.Body.Bytes(), &projects); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(projects) != 3 {
		t.Fatalf("expected 3 projects with limit=3 offset=1, got %d", len(projects))
	}
}
