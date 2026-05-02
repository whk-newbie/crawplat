package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"crawler-platform/apps/spider-service/internal/model"
	"crawler-platform/apps/spider-service/internal/service"
	"crawler-platform/packages/go-common/httpx"
	"github.com/gin-gonic/gin"
)

func TestCreateSpiderReturnsCreatedSpider(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := NewRouter(service.NewSpiderService())
	req := httptest.NewRequest(http.MethodPost, "/api/v1/projects/p1/spiders", strings.NewReader(`{"name":"crawler","language":"go","runtime":"docker","image":"crawler/go:latest","command":["./crawler"]}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", w.Code)
	}

	var spider model.Spider
	if err := json.Unmarshal(w.Body.Bytes(), &spider); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if spider.ID == "" {
		t.Fatal("expected generated id")
	}
	if spider.ProjectID != "p1" || spider.Name != "crawler" || spider.Language != "go" || spider.Runtime != "docker" {
		t.Fatalf("unexpected spider contents: %+v", spider)
	}
	if spider.Image != "crawler/go:latest" {
		t.Fatalf("expected image crawler/go:latest, got %s", spider.Image)
	}
	if len(spider.Command) != 1 || spider.Command[0] != "./crawler" {
		t.Fatalf("expected command to round-trip, got %#v", spider.Command)
	}
	if !strings.Contains(w.Body.String(), `"id":`) || !strings.Contains(w.Body.String(), `"projectId":`) || !strings.Contains(w.Body.String(), `"name":`) || !strings.Contains(w.Body.String(), `"language":`) || !strings.Contains(w.Body.String(), `"runtime":`) || !strings.Contains(w.Body.String(), `"image":`) {
		t.Fatalf("expected lower-case JSON keys, got %s", w.Body.String())
	}
}

func TestCreateSpiderRejectsInvalidInput(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := NewRouter(service.NewSpiderService())
	cases := []struct {
		name string
		body string
	}{
		{name: "invalid language", body: `{"name":"crawler","language":"ruby","runtime":"docker"}`},
		{name: "invalid runtime", body: `{"name":"crawler","language":"go","runtime":"vm"}`},
		{name: "missing image", body: `{"name":"crawler","language":"go","runtime":"docker"}`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/v1/projects/p1/spiders", strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != http.StatusBadRequest {
				t.Fatalf("expected status 400, got %d", w.Code)
			}
			if !strings.Contains(w.Body.String(), `"error":`) {
				t.Fatalf("expected error response, got %s", w.Body.String())
			}
		})
	}
}

func TestListSpidersReturnsOnlyRequestedProject(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := service.NewSpiderService()
	spiderP1, err := svc.Create("p1", "crawler-a", "go", "docker", "crawler/go-a:latest", []string{"./crawler-a"})
	if err != nil {
		t.Fatalf("expected create success, got error: %v", err)
	}
	spiderP2, err := svc.Create("p2", "crawler-b", "python", "host", "", []string{"python", "main.py"})
	if err != nil {
		t.Fatalf("expected create success, got error: %v", err)
	}
	router := NewRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/p1/spiders", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp struct {
		Items  []model.Spider `json:"items"`
		Total  int64          `json:"total"`
		Limit  int            `json:"limit"`
		Offset int            `json:"offset"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Total != 1 {
		t.Fatalf("expected total 1, got %d", resp.Total)
	}
	if len(resp.Items) != 1 {
		t.Fatalf("expected 1 spider, got %d", len(resp.Items))
	}
	if resp.Items[0].ID != spiderP1.ID {
		t.Fatalf("expected p1 spider %+v, got %+v", spiderP1, resp.Items[0])
	}
	if resp.Items[0].ID == spiderP2.ID {
		t.Fatalf("expected response to exclude p2 spider %+v", spiderP2)
	}
	if resp.Limit != 20 || resp.Offset != 0 {
		t.Fatalf("expected default pagination, got limit=%d offset=%d", resp.Limit, resp.Offset)
	}
	if !strings.Contains(w.Body.String(), `"projectId":"p1"`) || strings.Contains(w.Body.String(), `"projectId":"p2"`) {
		t.Fatalf("expected response to contain only p1 spider, got %s", w.Body.String())
	}
}

func TestListSpidersRespectsPaginationParams(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := service.NewSpiderService()
	for i := 0; i < 3; i++ {
		_, err := svc.Create("p1", "crawler", "go", "docker", "crawler/go:latest", []string{"./crawler"})
		if err != nil {
			t.Fatalf("expected create success, got error: %v", err)
		}
	}
	router := NewRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/p1/spiders?limit=1&offset=1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp httpx.PaginatedResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Total != 3 {
		t.Fatalf("expected total 3, got %d", resp.Total)
	}
	if resp.Limit != 1 || resp.Offset != 1 {
		t.Fatalf("expected limit=1 offset=1, got limit=%d offset=%d", resp.Limit, resp.Offset)
	}
}

func TestSpiderVersionRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := service.NewSpiderService()
	spider, err := svc.Create("p1", "crawler-a", "go", "docker", "crawler/go:latest", []string{"./crawler-a"})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	router := NewRouter(svc)

	createVersionReq := httptest.NewRequest(http.MethodPost, "/api/v1/spiders/"+spider.ID+"/versions", strings.NewReader(`{"registryAuthRef":"ghcr-prod","image":"crawler/go:v2","command":["./crawler-a","--fast"]}`))
	createVersionReq.Header.Set("Content-Type", "application/json")
	createVersionResp := httptest.NewRecorder()
	router.ServeHTTP(createVersionResp, createVersionReq)
	if createVersionResp.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d body=%s", createVersionResp.Code, createVersionResp.Body.String())
	}
	if !strings.Contains(createVersionResp.Body.String(), `"version":2`) {
		t.Fatalf("expected created version 2, got %s", createVersionResp.Body.String())
	}
	if !strings.Contains(createVersionResp.Body.String(), `"registryAuthRef":"ghcr-prod"`) {
		t.Fatalf("expected created version to include registryAuthRef, got %s", createVersionResp.Body.String())
	}

	listVersionsReq := httptest.NewRequest(http.MethodGet, "/api/v1/spiders/"+spider.ID+"/versions", nil)
	listVersionsResp := httptest.NewRecorder()
	router.ServeHTTP(listVersionsResp, listVersionsReq)
	if listVersionsResp.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", listVersionsResp.Code, listVersionsResp.Body.String())
	}
	if !strings.Contains(listVersionsResp.Body.String(), `"version":2`) || !strings.Contains(listVersionsResp.Body.String(), `"version":1`) {
		t.Fatalf("expected version list with v2 and v1, got %s", listVersionsResp.Body.String())
	}
	if !strings.Contains(listVersionsResp.Body.String(), `"registryAuthRef":"ghcr-prod"`) {
		t.Fatalf("expected version list to include registryAuthRef, got %s", listVersionsResp.Body.String())
	}
}

func TestListRegistryAuthRefsRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := service.NewSpiderService()
	spiderA, err := svc.Create("p1", "crawler-a", "go", "docker", "crawler/go-a:latest", []string{"./crawler-a"})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	spiderB, err := svc.Create("p1", "crawler-b", "go", "docker", "crawler/go-b:latest", []string{"./crawler-b"})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if _, err := svc.CreateVersion(spiderA.ID, "ghcr-prod", "crawler/go-a:v2", []string{"./crawler-a"}); err != nil {
		t.Fatalf("CreateVersion returned error: %v", err)
	}
	if _, err := svc.CreateVersion(spiderB.ID, "harbor-ci", "crawler/go-b:v2", []string{"./crawler-b"}); err != nil {
		t.Fatalf("CreateVersion returned error: %v", err)
	}
	router := NewRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/p1/registry-auth-refs", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", resp.Code, resp.Body.String())
	}
	if !strings.Contains(resp.Body.String(), "ghcr-prod") || !strings.Contains(resp.Body.String(), "harbor-ci") {
		t.Fatalf("expected refs payload, got %s", resp.Body.String())
	}
}
